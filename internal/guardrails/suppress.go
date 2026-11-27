// Package guardrails detects guardrail erosion introduced by a change (REQ-62,
// REQ-72–REQ-76). It has three mechanisms, all producing
// Finding values and combined by ChangesAll:
//
//   - SuppressAdded (suppress.go): lint/type-check suppressions on lines added
//     in head relative to base.
//   - FocusedTestsAdded (focused_tests.go): committed focused JS/TS tests
//     (`.only`, `fit`, `fdescribe`).
//   - ConfigChanges (config.go, config_*.go): effective erosion of tool
//     configuration between two Trees: ignore sets that grew, strictness that
//     dropped.
//
// Suppression check IDs:
//
//   - guardrails.SUPPRESS.added.unscoped (blocking-eligible): rule-less
//     suppressions: Python bare `# noqa`, bare `# type: ignore`,
//     `# pylint: disable=all`, `# mypy: ignore-errors`, `@no_type_check`; Go bare
//     `//nolint` and `//nolint:all`; TS/JS bare `eslint-disable[-next-line|-line]`,
//     `@ts-ignore`, `@ts-nocheck`.
//   - guardrails.SUPPRESS.added.scoped (report-only): the same mechanisms naming
//     a rule/linter/code.
//
// SuppressAdded examines only non-test, non-generated files and
// FocusedTestsAdded only non-generated JS/TS test files; the language, test
// path and generated-file classification lives in classify.go.
//
// Limits (documented, deliberate): a small lexer per language decides whether a
// marker sits in a comment. It handles line/block comments, single-line
// strings, Python triple-quoted strings, Go raw strings and TS template
// literals. It does not track `${...}` nesting inside template literals, TS/JS
// regex literals containing quotes or `//`, or Python line-continued strings;
// in those rare cases a marker may be missed or, for regex literals, a
// following comment misread. Line pairing for "modified, not added" uses
// positional pairing inside a diff hunk.
package guardrails

import (
	"regexp"
	"slices"
	"strings"
)

const (
	// UnscopedCheckID is the blocking-eligible check id.
	UnscopedCheckID = "guardrails.SUPPRESS.added.unscoped"
	// ScopedCheckID is the report-only check id for scoped suppressions.
	ScopedCheckID = "guardrails.SUPPRESS.added.scoped"
	// UnscopedVersion and ScopedVersion are the rule versions (id@N).
	UnscopedVersion = UnscopedCheckID + "@1"
	ScopedVersion   = ScopedCheckID + "@1"
)

// Evidence carries the line content that justifies a finding.
type Evidence struct {
	// Head is the added line in head.
	Head string `json:"head"`
	// Base is the base line it replaced in the same hunk ("" when the line is
	// purely new).
	Base string `json:"base"`
}

// Finding is one added suppression.
type Finding struct {
	Path      string   `json:"path"`
	Line      int      `json:"line"` // 1-based, in head
	CheckID   string   `json:"check_id"`
	Version   string   `json:"version"`
	Mechanism string   `json:"mechanism"`
	Evidence  Evidence `json:"evidence"`
}

// SuppressAdded returns suppressions on lines added in head vs base. base and
// head map repo-relative path to content; a path missing from base is a new
// file. changed restricts the examined paths; empty means every head path.
// Results are sorted by path, line, mechanism.
func SuppressAdded(base, head map[string][]byte, changed []string) []Finding {
	return scanHeadFiles(head, changed, func(p string, hc []byte) []Finding {
		l := langOf(p)
		if l == langNone || isTestPath(p, l) || isGenerated(hc) {
			return nil
		}
		return fileFindings(p, l, base[p], hc)
	})
}

// hit is one suppression marker on a line.
type hit struct {
	mechanism string
	scoped    bool
}

func fileFindings(p string, l lang, baseSrc, headSrc []byte) []Finding {
	headLines := splitLines(headSrc)
	baseLines := splitLines(baseSrc)
	headHits := lineHits(l, headLines)
	if !slices.ContainsFunc(headHits, func(h []hit) bool { return len(h) > 0 }) {
		return nil
	}
	baseHits := lineHits(l, baseLines)
	added := addedLines(baseLines, headLines)
	var out []Finding
	for _, a := range added {
		hs := headHits[a.idx]
		if len(hs) == 0 {
			continue
		}
		for _, h := range hs {
			if a.baseIdx >= 0 && containsHit(baseHits[a.baseIdx], h) {
				continue // existing suppression on an edited line
			}
			f := Finding{
				Path: p, Line: a.idx + 1, Mechanism: h.mechanism,
				Evidence: Evidence{Head: headLines[a.idx]},
			}
			if a.baseIdx >= 0 {
				f.Evidence.Base = baseLines[a.baseIdx]
			}
			if h.scoped {
				f.CheckID, f.Version = ScopedCheckID, ScopedVersion
			} else {
				f.CheckID, f.Version = UnscopedCheckID, UnscopedVersion
			}
			out = append(out, f)
		}
	}
	return out
}

func containsHit(hs []hit, h hit) bool {
	for _, x := range hs {
		if x == h {
			return true
		}
	}
	return false
}

// ---- line diff ----

type addedLine struct {
	idx     int // head line index
	baseIdx int // paired base line index in the same hunk, or -1
}

// addedLines returns head lines not matched by an LCS against base, excluding
// lines whose trimmed content was merely moved (present among unmatched base
// lines). Added lines are paired positionally with removed lines of the same
// hunk to provide base evidence.
func addedLines(base, head []string) []addedLine {
	// common prefix / suffix
	pre, suf := commonAffixes(base, head)
	bm, hm := base[pre:len(base)-suf], head[pre:len(head)-suf]
	matchedB := make([]bool, len(bm))
	matchedH := make([]bool, len(hm))
	if len(bm)*len(hm) <= 4_000_000 {
		lcsMark(bm, hm, matchedB, matchedH)
	}
	// hunks: maximal runs of unmatched lines between matches
	hunks := unmatchedHunks(matchedB, matchedH, pre)
	return unmovedAdditions(base, head, hunks)
}

func commonAffixes(base, head []string) (int, int) {
	pre := 0
	for pre < len(base) && pre < len(head) && base[pre] == head[pre] {
		pre++
	}
	suf := 0
	for suf < len(base)-pre && suf < len(head)-pre && base[len(base)-1-suf] == head[len(head)-1-suf] {
		suf++
	}
	return pre, suf
}

type hunk struct{ bs, hs []int }

func unmatchedRun(matched []bool, i, offset int) ([]int, int) {
	var run []int
	for i < len(matched) && !matched[i] {
		run = append(run, i+offset)
		i++
	}
	return run, i
}

func unmatchedHunks(matchedB, matchedH []bool, pre int) []hunk {
	var hunks []hunk
	bi, hi := 0, 0
	for bi < len(matchedB) || hi < len(matchedH) {
		var h hunk
		h.bs, bi = unmatchedRun(matchedB, bi, pre)
		h.hs, hi = unmatchedRun(matchedH, hi, pre)
		if len(h.bs) > 0 || len(h.hs) > 0 {
			hunks = append(hunks, h)
		}
		// skip matched pair
		if bi < len(matchedB) && hi < len(matchedH) && matchedB[bi] && matchedH[hi] {
			bi++
			hi++
		}
	}
	return hunks
}

func unmovedAdditions(base, head []string, hunks []hunk) []addedLine {
	removed := map[string]int{}
	for _, h := range hunks {
		for _, b := range h.bs {
			removed[strings.TrimSpace(base[b])]++
		}
	}
	var out []addedLine
	for _, h := range hunks {
		for k, hidx := range h.hs {
			t := strings.TrimSpace(head[hidx])
			if removed[t] > 0 {
				removed[t]-- // moved or re-indented, not added
				continue
			}
			bidx := -1
			if k < len(h.bs) {
				bidx = h.bs[k]
			}
			out = append(out, addedLine{idx: hidx, baseIdx: bidx})
		}
	}
	return out
}

func lcsMark(a, b []string, ma, mb []bool) {
	n, m := len(a), len(b)
	dp := make([][]int32, n+1)
	for i := range dp {
		dp[i] = make([]int32, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[i] == b[j]:
			ma[i], mb[j] = true, true
			i++
			j++
		case dp[i+1][j] >= dp[i][j+1]:
			i++
		default:
			j++
		}
	}
}

// ---- comment-aware lexing ----

type comment struct {
	text  string
	block bool
}

type lineInfo struct {
	code     string // code outside comments with string contents blanked
	comments []comment
}

type lexer struct {
	l       lang
	inBlock bool   // /* */ (Go, TS)
	triple  string // python triple quote delimiter in effect
	raw     bool   // Go raw string / TS template literal
}

type lineScan struct {
	line string
	i    int
	code strings.Builder
	cms  []comment
}

func lex(l lang, lines []string) []lineInfo {
	out := make([]lineInfo, len(lines))
	lx := &lexer{l: l}
	for li, line := range lines {
		s := &lineScan{line: line}
		for s.i < len(s.line) {
			lx.step(s)
		}
		out[li] = lineInfo{code: s.code.String(), comments: s.cms}
	}
	return out
}

func (lx *lexer) step(s *lineScan) {
	switch {
	case lx.inBlock:
		lx.blockComment(s)
	case lx.triple != "":
		lx.tripleString(s)
	case lx.raw:
		lx.rawString(s)
	case lx.l == langPython:
		lx.pythonCode(s)
	default:
		lx.cStyleCode(s)
	}
}

func (lx *lexer) blockComment(s *lineScan) {
	end := strings.Index(s.line[s.i:], "*/")
	if end < 0 {
		s.cms = append(s.cms, comment{s.line[s.i:], true})
		s.i = len(s.line)
	} else {
		s.cms = append(s.cms, comment{s.line[s.i : s.i+end], true})
		s.i += end + 2
		lx.inBlock = false
	}
}

func (lx *lexer) tripleString(s *lineScan) {
	end := strings.Index(s.line[s.i:], lx.triple)
	if end < 0 {
		s.i = len(s.line)
	} else {
		s.i += end + 3
		lx.triple = ""
		s.code.WriteString("   ")
	}
}

func (lx *lexer) rawString(s *lineScan) {
	end := strings.IndexAny(s.line[s.i:], "`")
	if end < 0 {
		s.i = len(s.line)
	} else {
		s.i += end + 1
		lx.raw = false
	}
}

func (lx *lexer) pythonCode(s *lineScan) {
	c := s.line[s.i]
	switch {
	case c == '#':
		s.cms = append(s.cms, comment{s.line[s.i+1:], false})
		s.i = len(s.line)
	case (c == '"' || c == '\'') && strings.HasPrefix(s.line[s.i:], strings.Repeat(string(c), 3)):
		lx.triple = strings.Repeat(string(c), 3)
		s.i += 3
	default:
		s.stringOrByte(c)
	}
}

func (lx *lexer) cStyleCode(s *lineScan) {
	c := s.line[s.i]
	switch {
	case strings.HasPrefix(s.line[s.i:], "//"):
		s.cms = append(s.cms, comment{s.line[s.i+2:], false})
		s.i = len(s.line)
	case strings.HasPrefix(s.line[s.i:], "/*"):
		lx.inBlock = true
		s.i += 2
	case c == '`' && (lx.l == langGo || lx.l == langTS):
		lx.raw = true
		s.i++
	default:
		s.stringOrByte(c)
	}
}

func (s *lineScan) stringOrByte(c byte) {
	if c == '"' || c == '\'' {
		s.i = skipString(s.line, s.i)
		s.code.WriteString(`""`)
		return
	}
	s.code.WriteByte(c)
	s.i++
}

// skipString returns the index after the single-line string starting at i
// (unterminated strings end at end of line).
func skipString(line string, i int) int {
	q := line[i]
	for j := i + 1; j < len(line); j++ {
		switch line[j] {
		case '\\':
			j++
		case q:
			return j + 1
		}
	}
	return len(line)
}

// ---- marker matching ----

var (
	reNoqa        = regexp.MustCompile(`(?i)(?:^|[^\w])noqa\b`)
	reNoqaCodes   = regexp.MustCompile(`^\s*:\s*[A-Za-z]+\d`)
	reTypeIgnore  = regexp.MustCompile(`\btype:\s*ignore\b(\s*\[)?`)
	rePylint      = regexp.MustCompile(`\bpylint:\s*disable\s*=\s*([^#]*)`)
	reMypyIgnore  = regexp.MustCompile(`\bmypy:\s*ignore-errors\b`)
	reNoTypeCheck = regexp.MustCompile(`^\s*@(?:\w+\.)*no_type_check\b`)
	reNolint      = regexp.MustCompile(`^\s*nolint\b(:(\S*))?`)
	reEslint      = regexp.MustCompile(`^[\s/*]*eslint-disable(-next-line|-line)?(?:\s|$)(.*)$`)
	reTSDirective = regexp.MustCompile(`^[\s/*]*@ts-(ignore|nocheck)\b`)
)

func lineHits(l lang, lines []string) [][]hit {
	infos := lex(l, lines)
	out := make([][]hit, len(infos))
	for i, in := range infos {
		out[i] = infoHits(l, in)
	}
	return out
}

func infoHits(l lang, in lineInfo) []hit {
	var hs []hit
	switch l {
	case langPython:
		if reNoTypeCheck.MatchString(in.code) {
			hs = append(hs, hit{"no_type_check", false})
		}
		for _, c := range in.comments {
			hs = append(hs, pythonComment(c.text)...)
		}
	case langGo:
		hs = goNolintHits(in.comments)
	case langTS:
		for _, c := range in.comments {
			hs = append(hs, tsComment(c.text)...)
		}
	}
	return hs
}

func goNolintHits(comments []comment) []hit {
	var hs []hit
	for _, c := range comments {
		if c.block {
			continue
		}
		if m := reNolint.FindStringSubmatch(c.text); m != nil {
			scoped := m[1] != "" && m[2] != "" && !nolintHasAll(m[2])
			hs = append(hs, hit{"nolint", scoped})
		}
	}
	return hs
}

func nolintHasAll(list string) bool {
	for _, p := range strings.Split(list, ",") {
		if p == "all" {
			return true
		}
	}
	return false
}

func pythonComment(text string) []hit {
	var hs []hit
	if loc := reNoqa.FindStringIndex(text); loc != nil {
		hs = append(hs, hit{"noqa", reNoqaCodes.MatchString(text[loc[1]:])})
	}
	if m := reTypeIgnore.FindStringSubmatch(text); m != nil {
		hs = append(hs, hit{"type-ignore", m[1] != ""})
	}
	if m := rePylint.FindStringSubmatch(text); m != nil {
		all := false
		for _, p := range strings.FieldsFunc(m[1], func(r rune) bool { return r == ',' || r == ' ' || r == '\t' }) {
			if p == "all" {
				all = true
			}
		}
		hs = append(hs, hit{"pylint-disable", !all})
	}
	if reMypyIgnore.MatchString(text) {
		hs = append(hs, hit{"mypy-ignore-errors", false})
	}
	return hs
}

func tsComment(text string) []hit {
	var hs []hit
	if m := reEslint.FindStringSubmatch(text); m != nil {
		rest := m[2]
		if i := strings.Index(rest, "--"); i >= 0 { // description separator
			rest = rest[:i]
		}
		mech := "eslint-disable" + m[1]
		hs = append(hs, hit{mech, strings.TrimSpace(rest) != ""})
	}
	if m := reTSDirective.FindStringSubmatch(text); m != nil {
		hs = append(hs, hit{"ts-" + m[1], false})
	}
	return hs
}

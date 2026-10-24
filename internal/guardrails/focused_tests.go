package guardrails

import "regexp"

const (
	// FocusedCheckID is the blocking-eligible check for committed focused tests.
	FocusedCheckID = "guardrails.TESTQ.focused-test-added"
	// FocusedVersion is its rule version (id@N).
	FocusedVersion = FocusedCheckID + "@1"
)

var (
	reOnly    = regexp.MustCompile(`\b(?:test|it|describe|context|specify|suite|bench)(?:\.\w+)*\.only\b`)
	reFocused = regexp.MustCompile(`(?:^|[^.\w$])(fit|fdescribe)\s*\(`)
)

// focusHits returns the focus markers on a line's code (comments and string
// contents excluded by the lexer).
func focusHits(code string) []string {
	var out []string
	if reOnly.MatchString(code) {
		out = append(out, "only")
	}
	for _, m := range reFocused.FindAllStringSubmatch(code, -1) {
		out = append(out, m[1])
	}
	return out
}

// FocusedTestsAdded reports focused tests (`.only`, `fit`, `fdescribe`) on
// lines added in head inside JS/TS test files. Non-test files, generated files
// and focus markers that already existed on the replaced base line are
// ignored. There is no Go or Python construct (allow-list table 1).
func FocusedTestsAdded(base, head map[string][]byte, changed []string) []Finding {
	return scanHeadFiles(head, changed, func(p string, hc []byte) []Finding {
		if langOf(p) != langTS || !isTestPath(p, langTS) || isGenerated(hc) {
			return nil
		}
		return focusedFindings(p, base[p], hc)
	})
}

func focusedFindings(p string, baseSrc, headSrc []byte) []Finding {
	var out []Finding
	headLines, baseLines := splitLines(headSrc), splitLines(baseSrc)
	headInfo := lex(langTS, headLines)
	baseInfo := lex(langTS, baseLines)
	for _, a := range addedLines(baseLines, headLines) {
		for _, h := range focusHits(headInfo[a.idx].code) {
			if a.baseIdx >= 0 && contains(focusHits(baseInfo[a.baseIdx].code), h) {
				continue
			}
			f := Finding{
				Path: p, Line: a.idx + 1, CheckID: FocusedCheckID, Version: FocusedVersion,
				Mechanism: "focused-" + h, Evidence: Evidence{Head: headLines[a.idx]},
			}
			if a.baseIdx >= 0 {
				f.Evidence.Base = baseLines[a.baseIdx]
			}
			out = append(out, f)
		}
	}
	return out
}

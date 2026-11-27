package tooling

import (
	"fmt"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"

	"scout/tools/gate/check"
)

const (
	ratchetRule    = "TEST.coverage-ratchet"
	coverageConfig = ".testcoverage.yml"
	coveragePragma = "coverage-ignore"
)

type coverageFloor struct {
	Threshold struct {
		File    *float64 `yaml:"file"`
		Package *float64 `yaml:"package"`
		Total   *float64 `yaml:"total"`
	} `yaml:"threshold"`
	Override []struct {
		Threshold float64 `yaml:"threshold"`
		Path      string  `yaml:"path"`
	} `yaml:"override"`
	Exclude struct {
		Paths []string `yaml:"paths"`
	} `yaml:"exclude"`
}

func parseFloor(text string) (coverageFloor, error) {
	var f coverageFloor
	err := yaml.Unmarshal([]byte(text), &f)
	return f, err
}

func floorChanges(base, head coverageFloor) []string {
	var out []string
	pairs := []struct {
		name string
		b, h *float64
	}{{"threshold.file", base.Threshold.File, head.Threshold.File}, {"threshold.package", base.Threshold.Package, head.Threshold.Package}, {"threshold.total", base.Threshold.Total, head.Threshold.Total}}
	for _, p := range pairs {
		if p.b != nil && (p.h == nil || *p.h < *p.b) {
			out = append(out, fmt.Sprintf("lowers %s from %v to %v", p.name, *p.b, deref(p.h)))
		}
	}
	headOverride := map[string]float64{}
	for _, o := range head.Override {
		headOverride[o.Path] = o.Threshold
	}
	for _, o := range base.Override {
		h, ok := headOverride[o.Path]
		if !ok || h < o.Threshold {
			out = append(out, fmt.Sprintf("lowers the override for %s from %v to %v", o.Path, o.Threshold, h))
		}
	}
	b, h := sortedCopy(base.Exclude.Paths), sortedCopy(head.Exclude.Paths)
	if strings.Join(b, "\x00") != strings.Join(h, "\x00") {
		out = append(out, fmt.Sprintf("changes exclude.paths from %v to %v", b, h))
	}
	return out
}

func deref(v *float64) string {
	if v == nil {
		return "unset"
	}
	return strconv.FormatFloat(*v, 'f', -1, 64)
}

func runRatchet(r *check.Run) ([]check.Finding, error) {
	if r.Explicit {
		return explicitPragmas(r)
	}
	return ratchet(r)
}

func explicitPragmas(r *check.Run) ([]check.Finding, error) {
	files, err := selectFiles(r, byExt(".go"))
	if err != nil {
		return nil, err
	}
	var out []check.Finding
	for _, f := range files {
		data, err := os.ReadFile(filepath.Join(r.Root, f))
		if err != nil {
			return nil, err
		}
		for _, line := range pragmaLines(data) {
			out = append(out, pragmaFinding(f, line))
		}
	}
	return out, nil
}

func pragmaLines(src []byte) []int {
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	var s scanner.Scanner
	s.Init(file, src, nil, scanner.ScanComments)
	var lines []int
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			return lines
		}
		if tok == token.COMMENT && pragmaComment.MatchString(lit) {
			lines = append(lines, fset.Position(pos).Line)
		}
	}
}

func pragmaFinding(file string, line int) check.Finding {
	return check.Finding{File: file, Line: line, Rule: ratchetRule, Problem: "adds a " + coveragePragma + " coverage pragma", Fix: "test the code instead of excluding it from coverage"}
}

func ratchet(r *check.Run) ([]check.Finding, error) {
	revs, ok, err := revisions(r)
	if err != nil {
		return nil, err
	}
	head, missing, err := headFloor(r, revs)
	if err != nil || missing != nil {
		return missing, err
	}
	if !ok {
		return nil, nil
	}
	out, err := floorFindings(r, revs, head)
	if err != nil {
		return nil, err
	}
	pragmas, err := addedPragmas(r, revs)
	if err != nil {
		return nil, err
	}
	return append(out, pragmas...), nil
}

func headFloor(r *check.Run, revs revisionPair) (coverageFloor, []check.Finding, error) {
	text, found, err := showAt(r, revs.head, coverageConfig)
	if err != nil {
		return coverageFloor{}, nil, err
	}
	if !found {
		return coverageFloor{}, []check.Finding{{File: coverageConfig, Rule: ratchetRule, Problem: "no coverage floor config", Fix: "restore " + coverageConfig + " with the owner-approved floor"}}, nil
	}
	head, err := parseFloor(text)
	if err != nil {
		return coverageFloor{}, []check.Finding{{File: coverageConfig, Rule: ratchetRule, Problem: "does not parse: " + oneLine(err.Error()), Fix: "fix the YAML"}}, nil
	}
	return head, nil, nil
}

func floorFindings(r *check.Run, revs revisionPair, head coverageFloor) ([]check.Finding, error) {
	baseText, found, err := showAt(r, revs.base, coverageConfig)
	if err != nil || !found {
		return nil, err
	}
	base, err := parseFloor(baseText)
	if err != nil {
		return nil, fmt.Errorf("%s at %s: %w", coverageConfig, revs.base, err)
	}
	changes := floorChanges(base, head)
	if len(changes) == 0 {
		return nil, nil
	}
	approved, err := approvalAdded(r, revs)
	if err != nil || approved {
		return nil, err
	}
	var out []check.Finding
	for _, c := range changes {
		out = append(out, check.Finding{File: coverageConfig, Rule: ratchetRule, Problem: c + " without an approval reference", Fix: "restore the floor, or add an owner-approved proofs.toml [[exception]] with rule = \"" + ratchetRule + "\" and path = \"" + coverageConfig + "\" in the same change"})
	}
	return out, nil
}

type proofsExceptions struct {
	Exception []struct {
		Rule string `toml:"rule"`
		Path string `toml:"path"`
	} `toml:"exception"`
}

func approvalAdded(r *check.Run, revs revisionPair) (bool, error) {
	count := func(rev string) (int, error) {
		text, found, err := showAt(r, rev, "proofs.toml")
		if err != nil || !found {
			return 0, err
		}
		var p proofsExceptions
		if _, err := toml.Decode(text, &p); err != nil {
			return 0, fmt.Errorf("proofs.toml at %q: %w", rev, err)
		}
		n := 0
		for _, e := range p.Exception {
			if e.Rule == ratchetRule && e.Path == coverageConfig {
				n++
			}
		}
		return n, nil
	}
	before, err := count(revs.base)
	if err != nil {
		return false, err
	}
	after, err := count(revs.head)
	return after > before, err
}

var pragmaComment = regexp.MustCompile(`//\s*` + coveragePragma + `\b`)

var hunkHeader = regexp.MustCompile(`^@@ -\d+(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

func addedPragmas(r *check.Run, revs revisionPair) ([]check.Finding, error) {
	args := []string{"diff", "--no-ext-diff", "--unified=0", "--no-color"}
	if r.Mode == check.Commit {
		args = append(args, "--cached")
	} else {
		args = append(args, revs.base, revs.head)
	}
	raw, err := r.Git(append(args, "--", "*.go")...)
	if err != nil {
		return nil, err
	}
	l, err := readLayout(r)
	if err != nil {
		return nil, err
	}
	added := addedLines(raw, l)
	var out []check.Finding
	for _, f := range sortedCopy(keys(added)) {
		src, found, err := showAt(r, revs.head, f)
		if err != nil {
			return nil, err
		}
		for _, n := range pragmaLines([]byte(src)) {
			if found && added[f][n] {
				out = append(out, pragmaFinding(f, n))
			}
		}
	}
	return out, nil
}

func addedLines(diff string, l layout) map[string]map[int]bool {
	added := map[string]map[int]bool{}
	file, line := "", 0
	for _, text := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(text, "+++ "):
			file = strings.TrimPrefix(strings.TrimPrefix(text, "+++ "), "b/")
		case hunkHeader.MatchString(text):
			line, _ = strconv.Atoi(hunkHeader.FindStringSubmatch(text)[1])
		case strings.HasPrefix(text, "+"):
			if !l.routineExcluded(file) {
				if added[file] == nil {
					added[file] = map[int]bool{}
				}
				added[file][line] = true
			}
			line++
		}
	}
	return added
}

func keys(m map[string]map[int]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

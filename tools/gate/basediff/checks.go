package basediff

import (
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"

	"scout/tools/gate/check"
	"scout/tools/gate/pathscope"
)

const (
	ruleGolden   = "TEST.golden-change-record"
	ruleIdentity = "TEST.identity-artefacts"
)

type layout struct {
	GoldenFiles      []string `toml:"golden_files"`
	ChangeRecord     string   `toml:"change_record"`
	TestFilePatterns []string `toml:"test_file_patterns"`
}

func Checks() []check.Check {
	return []check.Check{
		{Name: "golden", Rules: []string{ruleGolden}, Run: runGolden},
		{Name: "identity", Rules: []string{ruleIdentity}, Run: runIdentity},
	}
}

func changes(r *check.Run) ([]check.Change, string, error) {
	if r.Explicit {
		return nil, "", errors.New("needs base and head: run it with --base, or through the scenario canary")
	}
	base, err := pathscope.BaseRev(r)
	if err != nil || (r.Mode == check.CI && base == "") {
		return nil, "", err
	}
	cs, err := r.Diff()
	return cs, base, err
}

func runGolden(r *check.Run) ([]check.Finding, error) {
	var l layout
	if _, err := r.Proofs.Section("layout", &l); err != nil {
		return nil, err
	}
	if len(l.GoldenFiles) == 0 || l.ChangeRecord == "" {
		return nil, errors.New("proofs.toml [layout] needs golden_files and change_record")
	}
	cs, base, err := changes(r)
	if err != nil || len(cs) == 0 {
		return nil, err
	}
	args := []string{"diff", "-U0", "--no-ext-diff"}
	if r.Mode == check.Commit {
		args = append(args, "--cached")
	} else {
		args = append(args, base, "HEAD")
	}
	raw, err := r.Git(append(args, "--", l.ChangeRecord)...)
	if err != nil {
		return nil, err
	}
	return Golden(l.GoldenFiles, l.ChangeRecord, cs, added(raw)), nil
}

func added(diff string) []string {
	var out []string
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			out = append(out, line[1:])
		}
	}
	return out
}

func matchAny(pats []string, p string) bool {
	for _, pat := range pats {
		if pathscope.Match(pat, p) {
			return true
		}
	}
	return false
}

func Golden(golden []string, record string, cs []check.Change, recordLines []string) []check.Finding {
	seen := map[string]bool{}
	var out []check.Finding
	for _, c := range cs {
		for _, p := range []string{c.Path, c.From} {
			if p == "" || seen[p] || !matchAny(golden, p) {
				continue
			}
			seen[p] = true
			named := false
			for _, l := range recordLines {
				if strings.Contains(l, p) {
					named = true
				}
			}
			if !named {
				out = append(out, check.Finding{File: p, Rule: ruleGolden,
					Problem: "golden file changed with no change-record entry naming it",
					Fix:     fmt.Sprintf("add an entry to %s in this change that names %s and why its expected output changed", record, p)})
			}
		}
	}
	return out
}

func runIdentity(r *check.Run) ([]check.Finding, error) {
	var l layout
	if _, err := r.Proofs.Section("layout", &l); err != nil {
		return nil, err
	}
	artefacts := map[string]string{}
	ok, err := r.Proofs.Section("identity_artefacts", &artefacts)
	if err != nil {
		return nil, err
	}
	if !ok || len(artefacts) == 0 {
		return nil, errors.New("proofs.toml [identity_artefacts] declares no artefacts")
	}
	cs, _, err := changes(r)
	if err != nil {
		return nil, err
	}
	return Identity(artefacts, l.TestFilePatterns, cs), nil
}

func IsTest(p string, patterns []string) bool {
	name := path.Base(p)
	for _, pat := range []string{"*_test.go", "test_*.py", "*_test.py"} {
		if ok, _ := path.Match(pat, name); ok {
			return true
		}
	}
	for _, s := range strings.Split(path.Dir(p), "/") {
		if s == "testdata" {
			return true
		}
	}
	return matchAny(patterns, p)
}

func Identity(artefacts map[string]string, patterns []string, cs []check.Change) []check.Finding {
	var touched []string
	for _, c := range cs {
		for _, p := range []string{c.Path, c.From} {
			if p == "" {
				continue
			}
			if _, ok := artefacts[p]; ok {
				touched = append(touched, p)
				continue
			}
			if !IsTest(p, patterns) {
				return nil
			}
		}
	}
	sort.Strings(touched)
	var out []check.Finding
	for i, p := range touched {
		if i > 0 && touched[i-1] == p {
			continue
		}
		out = append(out, check.Finding{File: p, Rule: ruleIdentity,
			Problem: fmt.Sprintf("test-only change modifies an identity-bearing artefact (%s)", artefacts[p]),
			Fix:     "revert the artefact, or land it with the production change that requires it"})
	}
	return out
}

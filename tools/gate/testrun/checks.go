package testrun

import (
	"fmt"
	"sort"
	"strings"

	"scout/tools/gate/check"
	"scout/tools/gate/testsource"
)

func Checks() []check.Check {
	return []check.Check{
		{Name: "test.tiers", Rules: []string{"TEST.tier-selectors", "TEST.failure-propagation"}, Run: runTiers},
		{Name: "test.run", Rules: []string{"TEST.failure-propagation", "TEST.no-evasion", "TEST.t1-isolation", "TEST.user-data-guard"}, Run: runTests},
		{Name: "test.budget", Rules: []string{"TEST.budget.t1-time", "TEST.budget.per-test-time"}, Modes: []check.Mode{check.Commit}, Run: runBudget},
	}
}

func runTiers(r *check.Run) ([]check.Finding, error) {
	s := sessionFor(r)
	if err := s.prepare(r); err != nil {
		return nil, err
	}
	out := append([]check.Finding(nil), s.list.failures...)
	for _, f := range s.col.Files {
		if !testsource.ValidTier(f.Constraint, s.col.Tiers) {
			out = append(out, check.Finding{File: f.Display, Rule: "TEST.tier-selectors", Problem: fmt.Sprintf("build constraint %q selects no test tier", f.Constraint), Fix: "use no constraint (T1), //go:build boundary (T2), //go:build deferred or //go:build boundary && deferred"})
		}
	}
	if len(s.list.failures) > 0 {
		return out, nil
	}
	for _, t := range s.col.Tests {
		if exp := expectedTier(t.Constraint, s.col.Tiers); exp != "" {
			out = append(out, selectionFindings(s, t, exp)...)
		}
	}
	return append(out, emptyTiers(s)...), nil
}

func selectionFindings(s *session, t testsource.StaticTest, exp string) []check.Finding {
	got := s.list.tierOf(pkgKey(t.ImportPath, t.Name))
	switch {
	case len(got) == 0:
		return []check.Finding{{File: t.Display, Rule: "TEST.tier-selectors", Problem: fmt.Sprintf("%s is selected by no tier listing", t.Name), Fix: "give the file one tier build constraint and remove other constraints"}}
	case len(got) > 1:
		return []check.Finding{{File: t.Display, Rule: "TEST.tier-selectors", Problem: fmt.Sprintf("%s is selected by tiers %s", t.Name, strings.Join(got, ", ")), Fix: "keep one tier constraint per file"}}
	case got[0] != exp:
		return []check.Finding{{File: t.Display, Rule: "TEST.tier-selectors", Problem: fmt.Sprintf("%s lists in %s but its file declares %s", t.Name, got[0], exp), Fix: "rename the duplicate test name in this package or fix the file constraint"}}
	}
	return nil
}

func emptyTiers(s *session) []check.Finding {
	declared := map[string]bool{}
	for _, t := range s.col.Tests {
		declared[expectedTier(t.Constraint, s.col.Tiers)] = true
	}
	var out []check.Finding
	for _, f := range s.col.Files {
		tier := expectedTier(f.Constraint, s.col.Tiers)
		if tier != "" && !declared[tier] && len(s.list.byTier[tier]) == 0 {
			out = append(out, check.Finding{File: f.Display, Rule: "TEST.failure-propagation", Problem: fmt.Sprintf("tier %s has test files but collects no tests", tier), Fix: "add the test functions the file is meant to hold or delete the empty test file"})
			declared[tier] = true
		}
	}
	return out
}

func runTests(r *check.Run) ([]check.Finding, error) {
	s := sessionFor(r)
	if err := s.prepare(r); err != nil {
		return nil, err
	}
	tiers, err := s.tiersFor(r)
	if err != nil {
		return nil, err
	}
	var out []check.Finding
	for _, tier := range tiers {
		path, err := s.run(r, tier)
		if err != nil {
			return nil, err
		}
		want, err := s.selection(r, tier)
		if err != nil {
			return nil, err
		}
		rep, err := ReadReport(path)
		if err != nil {
			out = append(out, check.Finding{File: path, Rule: "TEST.failure-propagation", Problem: err.Error(), Fix: "rerun the gate; a gate run must produce a complete report for every tier it runs"})
			continue
		}
		out = append(out, judgeReport(r, s, tier, rep, want)...)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].String() < out[j].String() })
	return out, nil
}

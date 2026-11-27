package testrun

import (
	"fmt"
	"regexp"
	"strings"

	"scout/tools/gate/check"
)

const isolationFix = "remove the process start, network use or real-home write from this T1 test, or tag the file //go:build boundary if it drives a process or transport by design"

func judgeReport(r *check.Run, s *session, tier string, rep Report, want map[string]bool) []check.Finding {
	var out []check.Finding
	seen := map[string]bool{}
	failed := false
	for _, suite := range rep.Suites {
		found, suiteFailed := s.judgeSuite(r, tier, rep, suite, tracked{want: want, seen: seen})
		out = append(out, found...)
		failed = failed || suiteFailed
	}
	out = append(out, s.notRun(tier, want, seen)...)
	if rep.Exit != 0 && !failed && len(out) == 0 {
		out = append(out, check.Finding{File: "go test", Rule: "TEST.failure-propagation", Problem: fmt.Sprintf("%s run exited %d without a recorded failure", tier, rep.Exit), Fix: "fix the infrastructure error go test reports"})
	}
	return out
}

type tracked struct {
	want map[string]bool
	seen map[string]bool
}

func (s *session) judgeSuite(r *check.Run, tier string, rep Report, suite Suite, sel tracked) ([]check.Finding, bool) {
	var out []check.Finding
	failed := false
	for _, c := range suite.Cases {
		top := pkgKey(c.Class, strings.SplitN(c.Name, "/", 2)[0])
		if !sel.want[top] {
			continue
		}
		sel.seen[top] = true
		if f := s.caseFinding(r, tier, rep, c); f != nil {
			out = append(out, *f)
			failed = failed || c.Outcome() == "fail"
		}
	}
	if suite.Outcome == "fail" && len(suite.Cases) == 0 {
		out = append(out, check.Finding{File: suite.Name, Rule: "TEST.failure-propagation", Problem: fmt.Sprintf("%s %s: %s", tier, suite.Name, suite.Reason), Fix: "fix the package so its tests build and run"})
		failed = true
	}
	return out, failed
}

func (s *session) notRun(tier string, want, seen map[string]bool) []check.Finding {
	var out []check.Finding
	for key := range want {
		if !seen[key] {
			st := s.list.byKey[key]
			out = append(out, check.Finding{File: st.Display, Rule: "TEST.failure-propagation", Problem: fmt.Sprintf("%s was selected for %s but did not run", st.Name, tier), Fix: "fix the package build or the earlier crash that stopped the run"})
		}
	}
	return out
}

func (s *session) caseFinding(r *check.Run, tier string, rep Report, c Case) *check.Finding {
	top := strings.SplitN(c.Name, "/", 2)[0]
	file := s.list.byKey[pkgKey(c.Class, top)].Display
	switch c.Outcome() {
	case "fail":
		if rep.Guard != "" && killed(c) {
			if s.homeWrite(tier, c.Class, top) {
				return &check.Finding{File: file, Rule: "TEST.user-data-guard", Problem: fmt.Sprintf("%s tried to write under the real home directory %s; the guard blocked it", c.Name, s.realHome), Fix: "write only under t.TempDir() or the redirected HOME and XDG directories"}
			}
			return &check.Finding{File: file, Rule: "TEST.t1-isolation", Problem: fmt.Sprintf("%s was stopped by the T1 guard (%s)", c.Name, rep.Guard), Fix: isolationFix}
		}
		return &check.Finding{File: file, Rule: "TEST.failure-propagation", Problem: fmt.Sprintf("%s failed in %s: %s", c.Name, tier, oneLine(c)), Fix: "fix the code or the test until it passes"}
	case "skip":
		if r.Proofs.Excepted("TEST.no-evasion", s.list.byKey[pkgKey(c.Class, top)].File, "test.run") {
			return nil
		}
		return &check.Finding{File: file, Rule: "TEST.no-evasion", Problem: fmt.Sprintf("%s skipped in %s: %s", c.Name, tier, c.Reason()), Fix: "provide the prerequisite and make its absence a failure, or record an owner-approved [[exception]] for this path"}
	}
	return nil
}

func oneLine(c Case) string {
	body := ""
	if c.Failure != nil {
		body = c.Failure.Body
	}
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "=== ") && !strings.HasPrefix(line, "--- ") {
			return line
		}
	}
	return c.Reason()
}

func (s *session) homeWrite(tier, pkg, name string) bool {
	if s.guard.homeOnly == "" {
		return false
	}
	args := []string{"test", "-count=1", "-timeout", s.timeout.String(), "-exec", s.guard.homeOnly, "-run", "^" + regexp.QuoteMeta(name) + "$"}
	if tier == TierT1Deferred {
		args = append(args, "-tags", s.col.Tiers.Deferred)
	}
	res, err := goCommand(s.col.Root, s.list.env, 2*s.timeout, append(args, pkg)...)
	return err == nil && res.exit != 0 && strings.Contains(string(res.out), "signal: killed")
}

func killed(c Case) bool {
	return c.Failure != nil && (strings.HasPrefix(c.Failure.Message, "did not finish") || strings.Contains(c.Failure.Body, "signal: killed"))
}

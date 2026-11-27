package testrun

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"scout/tools/gate/check"
	"scout/tools/gate/proof"
	"scout/tools/gate/testsource"
)

func writeReport(t *testing.T, rep Report) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "report.xml")
	if err := WriteReport(path, rep); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestReadReportRejectsMissingInvalidAndNegativeReports(t *testing.T) {
	if _, err := ReadReport(filepath.Join(t.TempDir(), "absent.xml")); err == nil {
		t.Fatal("missing report must fail")
	}
	bad := filepath.Join(t.TempDir(), "bad.xml")
	if err := os.WriteFile(bad, []byte("<testsuites"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadReport(bad); err == nil {
		t.Fatal("unparsable report must fail")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	neg := writeReport(t, Report{Timestamp: now, Suites: []Suite{{Name: "p", Cases: []Case{{Name: "TestX", Class: "p", Time: -1}}}}})
	if _, err := ReadReport(neg); err == nil {
		t.Fatal("negative duration must fail")
	}
	if _, err := ReadReport(writeReport(t, Report{Timestamp: "yesterday"})); err == nil {
		t.Fatal("invalid timestamp must fail")
	}
}

func TestNormaliseFailsTheTestAKilledBinaryWasRunning(t *testing.T) {
	events := []event{
		{Action: "run", Package: "p", Test: "TestDone"},
		{Action: "pass", Package: "p", Test: "TestDone", Elapsed: 0.1},
		{Action: "run", Package: "p", Test: "TestStarts"},
		{Action: "output", Package: "p", Output: "signal: killed\n"},
		{Action: "fail", Package: "p", Elapsed: 0.2},
	}
	rep := Normalise(TierT1, "guard", result{started: time.Now()}, events)
	cases := rep.Suites[0].Cases
	if len(cases) != 2 || cases[0].Outcome() != "pass" || cases[1].Outcome() != "fail" || !killed(cases[1]) {
		t.Fatalf("cases = %+v", cases)
	}
}

func syntheticSession() *session {
	st := testsource.StaticTest{ImportPath: "p", Name: "TestSlowPath", Display: "p/slow_test.go", File: "p/slow_test.go"}
	other := testsource.StaticTest{ImportPath: "p", Name: "TestSkipped", Display: "p/skip_test.go", File: "p/skip_test.go"}
	return &session{list: &listing{byKey: map[string]testsource.StaticTest{"p.TestSlowPath": st, "p.TestSkipped": other}}, realHome: "/home/u"}
}

func TestOverBudgetNamesTheSlowTestAndTheWallLimit(t *testing.T) {
	rep := Report{Wall: 9, Suites: []Suite{{Name: "p", Cases: []Case{{Name: "TestSlowPath", Class: "p", Time: 3}, {Name: "TestSlowPath/row", Class: "p", Time: 3}}}}}
	findings := overBudget(syntheticSession(), rep, machine{ID: "m", Wall: 5, PerTest: 1})
	if len(findings) != 2 || findings[0].Rule != "TEST.budget.t1-time" || findings[1].Rule != "TEST.budget.per-test-time" || findings[1].File != "p/slow_test.go" {
		t.Fatalf("findings = %v", findings)
	}
	if !strings.Contains(findings[1].Fix, "never move it to another tier") {
		t.Fatalf("fix = %q", findings[1].Fix)
	}
	if got := overBudget(syntheticSession(), Report{Wall: 1}, machine{ID: "m", Wall: 5, PerTest: 1}); len(got) != 0 {
		t.Fatalf("within budget: %v", got)
	}
}

func TestReportOlderThanTheRunIsStale(t *testing.T) {
	started := time.Now()
	old := Report{Timestamp: started.Add(-time.Minute).UTC().Format(time.RFC3339Nano)}
	if f := staleReport("r.xml", old, started); f == nil || f.Rule != "TEST.budget.t1-time" {
		t.Fatalf("stale report must fail: %v", f)
	}
	fresh := Report{Timestamp: started.Add(time.Second).UTC().Format(time.RFC3339Nano)}
	if f := staleReport("r.xml", fresh, started); f != nil {
		t.Fatalf("fresh report: %v", f)
	}
}

func TestJudgeReportFailsSkippedMissingAndFailedTests(t *testing.T) {
	path := filepath.Join(t.TempDir(), "proofs.toml")
	if err := os.WriteFile(path, []byte("schema = 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	pf, err := proof.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	r := &check.Run{Proofs: pf}
	rep := Report{Exit: 1, Suites: []Suite{{Name: "p", Cases: []Case{{Name: "TestSkipped", Class: "p", Skipped: &Message{Message: "tool missing"}}}}}}
	want := map[string]bool{"p.TestSkipped": true, "p.TestSlowPath": true}
	got := rules(judgeReport(r, syntheticSession(), TierT1, rep, want))
	if !got["TEST.no-evasion"] || !got["TEST.failure-propagation"] {
		t.Fatalf("rules = %v", got)
	}
	exitOnly := judgeReport(r, syntheticSession(), TierT1, Report{Exit: 2}, map[string]bool{})
	if len(exitOnly) != 1 || exitOnly[0].Rule != "TEST.failure-propagation" {
		t.Fatalf("an exit without a recorded failure must fail: %v", exitOnly)
	}
}

func rules(fs []check.Finding) map[string]bool {
	out := map[string]bool{}
	for _, f := range fs {
		out[f.Rule] = true
	}
	return out
}

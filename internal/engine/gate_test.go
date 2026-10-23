package engine

import (
	"errors"
	"strings"
	"testing"

	"scout/internal/domain"
	"scout/internal/policy"
)

func TestExitCodePrecedence(t *testing.T) {
	for _, c := range []struct {
		unavailable, incomplete, gated bool
		want                           int
	}{
		{false, false, false, 0},
		{false, false, true, 1},
		{false, true, false, 2},
		{false, true, true, 2},
		{true, false, false, 3},
		{true, true, true, 3},
		{true, false, true, 3},
	} {
		if got := ExitCode(c.unavailable, c.incomplete, c.gated); got != c.want {
			t.Errorf("ExitCode(%v,%v,%v) = %d, want %d", c.unavailable, c.incomplete, c.gated, got, c.want)
		}
	}
}

func issue(id, rule, path, fingerprint string, state domain.IssueState) domain.Issue {
	return domain.Issue{IssueID: id, State: state, Title: id, EffectiveSeverity: domain.SeverityMedium,
		Anchor: domain.IssueAnchor{Path: path, RuleKey: rule, Fingerprint: fingerprint}}
}

func TestAttributeClassifiesAgainstBase(t *testing.T) {
	base := []domain.Issue{
		issue("b1", "ruff:C901", "a.py", "f1", domain.IssueOpen),
		issue("b2", "ruff:E1", "b.py", "f2", domain.IssueResolved),
		issue("b3", "ruff:C901", "c.py", "old", domain.IssueOpen),
	}
	head := []domain.Issue{
		issue("h1", "ruff:C901", "a.py", "f1", domain.IssueOpen),      // exact -> pre-existing
		issue("h2", "ruff:E1", "b.py", "f2", domain.IssueOpen),        // resolved twin -> regressed
		issue("h3", "ruff:C901", "c.py", "drifted", domain.IssueOpen), // same rule+path, drift -> pre-existing
		issue("h4", "ruff:C901", "d.py", "new", domain.IssueOpen),     // absent -> introduced
	}
	got := Attribute(head, base, domain.ComparisonRecomputed)
	want := map[string]domain.Attribution{"h1": domain.AttributionPreExisting, "h2": domain.AttributionRegressed, "h3": domain.AttributionPreExisting, "h4": domain.AttributionIntroduced}
	for id, attribution := range want {
		if got[id] != attribution {
			t.Errorf("%s = %s, want %s", id, got[id], attribution)
		}
	}
	for id, attribution := range Attribute(head, nil, domain.ComparisonNone) {
		if attribution != domain.AttributionUnknown {
			t.Errorf("no baseline: %s = %s, want unknown (nothing is introduced)", id, attribution)
		}
	}
}

const guardrailsID = "guardrails.SUPPRESS.added.unscoped"
const guardrailsVersion = guardrailsID + "@1"

func guardianPolicy(t *testing.T, version string) policy.Policy {
	t.Helper()
	p, err := policy.Parse([]byte("profile: guardian\nmaintainers: [alice]\nblocking:\n  - id: " + guardrailsID + "\n    version: \"" + version + "\"\n    activated_at: 2026-01-01\n    evidence: {would_block: 1, warranted: 1, window_days: 15, pr_runs: 6}\n"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func gateInput(p policy.Policy) GateInput {
	head := []domain.Issue{
		issue("sup", "guardrails:"+guardrailsID, "a.py", "x", domain.IssueOpen),
		issue("pat", "gitleaks:github-pat", "a.py", "y", domain.IssueOpen),
		issue("old", "ruff:C901", "z.py", "o", domain.IssueOpen),
	}
	return GateInput{
		Policy: p, Head: head, Comparison: domain.ComparisonRecomputed, BaseRevision: "b", HeadRevision: "h",
		Base: []domain.Issue{issue("bo", "ruff:C901", "z.py", "o", domain.IssueOpen)},
		Facts: map[string]CheckFact{
			"sup": {CheckID: guardrailsID, RuleVersion: guardrailsVersion},
			"pat": {CheckID: "security.secrets.github-pat", RuleVersion: "v"},
			"old": {CheckID: "ruff:C901"},
		},
		Changed: map[string]bool{"a.py": true},
		Lanes:   LaneHealthOf(domain.TriggerPR, domain.Run{}, nil),
	}
}

func TestEvaluateActiveVersusShadow(t *testing.T) {
	out := Evaluate(gateInput(guardianPolicy(t, guardrailsVersion)))
	if out.Gate.ExitCode != ExitGated || len(out.Gate.Blocking) != 1 || out.Gate.Blocking[0] != "sup" {
		t.Fatalf("active check: %+v", out.Gate)
	}
	if len(out.WouldBlock) != 1 || out.WouldBlock[0].IssueID != "pat" {
		t.Fatalf("PAT must stay would_block, got %+v", out.WouldBlock)
	}
	if out.Pre != 1 {
		t.Fatalf("pre-existing = %d, want 1", out.Pre)
	}
	bumped := Evaluate(gateInput(guardianPolicy(t, guardrailsID+"@0")))
	if bumped.Gate.ExitCode != ExitClean || len(bumped.WouldBlock) != 2 {
		t.Fatalf("version bump must return to shadow: exit %d, would_block %d", bumped.Gate.ExitCode, len(bumped.WouldBlock))
	}
	report := Evaluate(gateInput(policy.Default()))
	if report.Gate.ExitCode != ExitClean || len(report.WouldBlock) != 2 {
		t.Fatalf("report profile: exit %d, would_block %d", report.Gate.ExitCode, len(report.WouldBlock))
	}
}

func TestEvaluateOverrides(t *testing.T) {
	request := ParseOverrides("hello\nscout:allow " + guardrailsID + " reason=legacy shim ref=#42\n")
	if len(request) != 1 || request[0].CheckID != guardrailsID || request[0].Reason != "legacy shim" || request[0].Ref != "#42" {
		t.Fatalf("parsed %+v", request)
	}
	in := gateInput(guardianPolicy(t, guardrailsVersion))
	in.Overrides, in.OverrideActor = request, "mallory"
	rejected := Evaluate(in)
	if rejected.Gate.ExitCode != ExitGated || len(rejected.Accepted) != 0 || len(rejected.Gate.Overrides) != 1 || rejected.Gate.Overrides[0].Accepted {
		t.Fatalf("non-maintainer override: %+v", rejected.Gate)
	}
	in.OverrideActor = "alice"
	accepted := Evaluate(in)
	if accepted.Gate.ExitCode != ExitClean || len(accepted.Accepted) != 1 || accepted.Accepted[0].IssueID != "sup" || accepted.Accepted[0].Ref != "#42" {
		t.Fatalf("maintainer override: %+v %+v", accepted.Gate, accepted.Accepted)
	}
	if got := ParseOverrides("scout-override"); len(got) != 1 || !got[0].All {
		t.Fatalf("label not parsed: %+v", got)
	}
}

func TestOptionalLanesNeverMakeRequiredLaneRed(t *testing.T) {
	run := domain.Run{Coverage: []domain.Coverage{
		{Key: domain.CoverageKey{FilePath: "a.py", Family: domain.FamilySecurity, Category: "credential-leakage"}, Status: domain.CoverageInspected},
		{Key: domain.CoverageKey{FeatureID: "f", Family: domain.FamilySecurity, Category: "sensitive_data_flows"}, Status: domain.CoverageBlocked, Reason: "contextual provider is not configured"},
		{Key: domain.CoverageKey{FeatureID: "f", Family: domain.FamilyClawpatch, Category: "review"}, Status: domain.CoverageBlocked},
	}}
	health := LaneHealthOf(domain.TriggerPush, run, nil)
	if bad, reasons := Incomplete(health); bad {
		t.Fatalf("optional lanes made the run incomplete: %v", reasons)
	}
	byLane := map[string]domain.LaneHealth{}
	for _, h := range health {
		byLane[h.Lane] = h
	}
	if byLane[LaneSecurity].Status != domain.LaneRan || byLane[LaneContextual].Status != domain.LaneBlocked || byLane[LaneClawpatchReview].Status != domain.LaneBlocked {
		t.Fatalf("health = %+v", byLane)
	}
	run.Coverage[0].Status = domain.CoverageBlocked
	run.Coverage[0].Reason = "gitleaks is not installed"
	bad, reasons := Incomplete(LaneHealthOf(domain.TriggerPush, run, nil))
	if !bad || !strings.Contains(reasons[0], "security blocked") {
		t.Fatalf("blocked required lane must be incomplete, got %v %v", bad, reasons)
	}
	in := gateInput(policy.Default())
	in.Lanes = LaneHealthOf(domain.TriggerPush, run, nil)
	if out := Evaluate(in); out.Gate.ExitCode != ExitIncomplete {
		t.Fatalf("required lane blocked on head: exit %d, want 2", out.Gate.ExitCode)
	}
	in.Unavailable = errors.New("boom")
	if out := Evaluate(in); out.Gate.ExitCode != ExitUnavailable {
		t.Fatalf("unavailable must outrank incomplete: exit %d", out.Gate.ExitCode)
	}
}

func TestSummaryIsBounded(t *testing.T) {
	var issues []domain.AttributedIssue
	var rows []domain.WouldBlock
	for range 200 {
		issues = append(issues, domain.AttributedIssue{IssueID: "i", CheckID: "c", Attribution: domain.AttributionIntroduced, Title: "t"})
		rows = append(rows, domain.WouldBlock{CheckID: "c", Path: "p", Triage: "undecided"})
	}
	out := Summary(domain.Gate{Status: "gated", ExitCode: 1, Reasons: strings.Split(strings.Repeat("r\n", 100), "\n")}, DeterministicRanking(issues, "decisions not enabled"), issues, rows, 5, LaneHealthOf(domain.TriggerPR, domain.Run{}, nil))
	if n := strings.Count(out, "\n"); n > 60 {
		t.Fatalf("summary has %d lines", n)
	}
}

package engine

import (
	"testing"

	"scout/internal/domain"
	"scout/internal/modelcli"
	"scout/internal/policy"
)

func TestNoModelCLILanesAreSkippedAndExitCodeIsUnchanged(t *testing.T) {
	skipped := func(key domain.CoverageKey) domain.Coverage {
		return domain.Coverage{Key: key, Status: domain.CoverageSkipped, Reason: modelcli.NoCLIReason}
	}
	run := domain.Run{Coverage: []domain.Coverage{
		{Key: domain.CoverageKey{FilePath: "a.py", Family: domain.FamilySecurity, Category: "credential-leakage"}, Status: domain.CoverageInspected},
		skipped(domain.CoverageKey{FeatureID: "f", Family: domain.FamilySecurity, Category: "sensitive_data_flows"}),
		skipped(domain.CoverageKey{FeatureID: "f", Family: domain.FamilyClawpatch, Category: "review"}),
	}}
	health := LaneHealthOf(domain.TriggerNightly, run, nil)
	byLane := map[string]domain.LaneHealth{}
	for _, h := range health {
		byLane[h.Lane] = h
	}
	for _, id := range []string{LaneContextual, LaneClawpatchReview} {
		got := byLane[id]
		if got.Status != domain.LaneSkipped || got.Reason != "skipped (no model CLI)" || got.Requires != domain.RequiresModelCLI || got.Required {
			t.Fatalf("%s health = %+v", id, got)
		}
	}
	if bad, reasons := Incomplete(health); bad {
		t.Fatalf("skipped model-CLI lanes made the run incomplete: %v", reasons)
	}
	with := gateInput(policy.Default())
	with.Lanes = health
	without := gateInput(policy.Default())
	if a, b := Evaluate(with).Gate.ExitCode, Evaluate(without).Gate.ExitCode; a != b || a != ExitClean {
		t.Fatalf("exit with skipped model-CLI lanes = %d, without = %d, want both %d", a, b, ExitClean)
	}
}

func TestSkippedCoverageForOtherReasonsStillFailsTheLane(t *testing.T) {
	run := domain.Run{Coverage: []domain.Coverage{
		{Key: domain.CoverageKey{FeatureID: "f", Family: domain.FamilySecurity, Category: "sensitive_data_flows"}, Status: domain.CoverageSkipped, Reason: "operator skipped"},
	}}
	for _, h := range LaneHealthOf(domain.TriggerNightly, run, nil) {
		if h.Lane == LaneContextual && h.Status != domain.LaneFailed {
			t.Fatalf("only the no-model-CLI reason may read as skipped: %+v", h)
		}
	}
}

func TestNotOptedInLanesAreSkippedNotFailedAndExitCodeIsUnchanged(t *testing.T) {
	run := domain.Run{Coverage: []domain.Coverage{
		{Key: domain.CoverageKey{FeatureID: "f", Family: domain.FamilySecurity, Category: "sensitive_data_flows"}, Status: domain.CoverageSkipped, Reason: ReasonNotOptedIn},
		{Key: domain.CoverageKey{FeatureID: "f", Family: domain.FamilyClawpatch, Category: "review"}, Status: domain.CoverageSkipped, Reason: ReasonNotOptedIn},
	}}
	health := LaneHealthOf(domain.TriggerNightly, run, nil)
	for _, h := range health {
		if (h.Lane == LaneContextual || h.Lane == LaneClawpatchReview) && (h.Status != domain.LaneSkipped || h.Reason != "skipped (not opted in)" || h.Required) {
			t.Fatalf("%s = %+v", h.Lane, h)
		}
	}
	with := gateInput(policy.Default())
	with.Lanes = health
	if a, b := Evaluate(with).Gate.ExitCode, Evaluate(gateInput(policy.Default())).Gate.ExitCode; a != b {
		t.Fatalf("exit with = %d, without = %d", a, b)
	}
}

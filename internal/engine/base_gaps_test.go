package engine

import (
	"context"
	"strings"
	"testing"

	"scout/internal/domain"
	"scout/internal/policy"
)

const secretsID = "security.secrets.github-pat"

func secretsPolicy(t *testing.T) policy.Policy {
	t.Helper()
	p, err := policy.Parse([]byte("profile: guardian\nmaintainers: [alice]\nblocking:\n  - id: " + secretsID + "\n    version: \"2\"\n    activated_at: 2026-01-01\n    evidence: {would_block: 1, warranted: 1, window_days: 15, pr_runs: 6}\n"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// secretsInput is a head with one new secret in the security family and a base
// that does not contain it; gaps describe what the base side did not inspect.
func secretsInput(t *testing.T, gaps BaseGaps) GateInput {
	t.Helper()
	pat := issue("pat", "gitleaks:github-pat", "a.py", "y", domain.IssueOpen)
	pat.Family = domain.FamilySecurity
	return GateInput{
		Policy: secretsPolicy(t), Head: []domain.Issue{pat}, Comparison: domain.ComparisonRecomputed, BaseRevision: "b", HeadRevision: "h",
		Facts:    map[string]CheckFact{"pat": {CheckID: secretsID, RuleVersion: "2"}},
		Changed:  map[string]bool{"a.py": true},
		Lanes:    LaneHealthOf(domain.TriggerPR, domain.Run{}, nil),
		BaseGaps: gaps,
	}
}

func baseRunWithSecurity(status domain.CoverageStatus, reason string) domain.Run {
	return domain.Run{Coverage: []domain.Coverage{{
		Key:    domain.CoverageKey{FilePath: "a.py", Family: domain.FamilySecurity, Category: "secrets"},
		Status: status, Reason: reason,
	}}}
}

func requireUnknown(t *testing.T, in GateInput, lane string) {
	t.Helper()
	out := Evaluate(in)
	if len(out.Issues) != 1 || out.Issues[0].Attribution != domain.AttributionUnknown {
		t.Fatalf("attribution = %+v, want unknown", out.Issues)
	}
	if out.Gate.ExitCode != ExitClean || len(out.Gate.Blocking) != 0 || len(out.WouldBlock) != 0 || out.Issues[0].Blocking || out.Issues[0].WouldBlock {
		t.Fatalf("unknown must neither gate nor would_block: gate %+v would_block %+v", out.Gate, out.WouldBlock)
	}
	if !strings.Contains(strings.Join(out.Gate.Reasons, "\n"), lane) {
		t.Fatalf("banner naming %q missing: %v", lane, out.Gate.Reasons)
	}
}

func TestRecomputedBaseWithFailedLaneAttributesUnknown(t *testing.T) {
	for _, status := range []domain.CoverageStatus{domain.CoverageBlocked, domain.CoverageIncomplete, domain.CoverageSkipped} {
		health := LaneHealthOf(domain.TriggerPR, baseRunWithSecurity(status, "gitleaks exited 2"), nil)
		gaps := gapsOfLanes(health)
		if _, ok := gaps[domain.FamilySecurity]; !ok {
			t.Fatalf("%s: security gap missing from %v", status, gaps)
		}
		requireUnknown(t, secretsInput(t, gaps), "security")
	}
}

func TestInspectedRecomputedBaseStillIntroducesAndBlocks(t *testing.T) {
	health := LaneHealthOf(domain.TriggerPR, baseRunWithSecurity(domain.CoverageInspected, ""), nil)
	gaps := gapsOfLanes(health)
	if len(gaps) != 0 {
		t.Fatalf("inspected base has gaps: %v", gaps)
	}
	out := Evaluate(secretsInput(t, gaps))
	if out.Issues[0].Attribution != domain.AttributionIntroduced || out.Gate.ExitCode != ExitGated || len(out.Gate.Blocking) != 1 {
		t.Fatalf("inspected base: %+v %+v", out.Issues, out.Gate)
	}
	// A gap in another family never affects security findings.
	out = Evaluate(secretsInput(t, BaseGaps{domain.FamilyQuality: "base lane quality failed"}))
	if out.Issues[0].Attribution != domain.AttributionIntroduced || out.Gate.ExitCode != ExitGated {
		t.Fatalf("unrelated gap changed attribution: %+v", out.Issues)
	}
}

func TestGuardrailsFindingsIgnoreBaseLaneGaps(t *testing.T) {
	in := gateInput(guardianPolicy(t, guardrailsVersion))
	sup := in.Head[0]
	sup.Family = domain.FamilyQuality
	in.Head[0] = sup
	in.BaseGaps = BaseGaps{domain.FamilyQuality: "base lane quality failed"}
	out := Evaluate(in)
	if out.Gate.ExitCode != ExitGated || len(out.Gate.Blocking) != 1 || out.Gate.Blocking[0] != "sup" {
		t.Fatalf("guardrails finding must still block: %+v", out.Gate)
	}
}

func TestImportedBaselineGapsAttributeUnknown(t *testing.T) {
	plan := domain.Plan{ToolDigest: "tool", PolicyDigest: "policy"}
	imported := &BaseState{
		Revision: "mb", ToolDigest: "tool", PolicyDigest: "policy",
		Gaps: BaseGaps{domain.FamilySecurity: "baseline family security not fully inspected: blocked: gitleaks missing"},
	}
	cfg := DiffConfig{Baseline: func(context.Context) (*BaseState, error) { return imported, nil }}
	_, gaps, comparison, _, err := resolveBase(context.Background(), cfg, t.TempDir(), "mb", plan, Result{})
	if err != nil || comparison != domain.ComparisonBaseline || len(gaps) != 1 {
		t.Fatalf("imported partial baseline: gaps %v comparison %v err %v", gaps, comparison, err)
	}
	in := secretsInput(t, gaps)
	in.Comparison = comparison
	requireUnknown(t, in, "security")

	// A fully inspected baseline (empty, non-nil gaps) keeps `introduced`.
	imported.Gaps = BaseGaps{}
	_, gaps, comparison, _, err = resolveBase(context.Background(), cfg, t.TempDir(), "mb", plan, Result{})
	if err != nil || comparison != domain.ComparisonBaseline {
		t.Fatalf("inspected imported baseline: %v %v", comparison, err)
	}
	in = secretsInput(t, gaps)
	in.Comparison = comparison
	if out := Evaluate(in); out.Issues[0].Attribution != domain.AttributionIntroduced || out.Gate.ExitCode != ExitGated {
		t.Fatalf("inspected imported baseline must still block: %+v", out.Gate)
	}
}

func TestImportedBaselineWithoutLaneHealthIsRecomputed(t *testing.T) {
	plan := domain.Plan{ToolDigest: "tool", PolicyDigest: "policy"}
	imported := &BaseState{Revision: "mb", ToolDigest: "tool", PolicyDigest: "policy"}
	cfg := DiffConfig{Baseline: func(context.Context) (*BaseState, error) { return imported, nil }}
	_, _, comparison, banners, err := resolveBase(context.Background(), cfg, t.TempDir(), "mb", plan, Result{})
	// The scratch root is not a git repository, so the recompute has no base.
	if err != nil || comparison != domain.ComparisonNone || !strings.Contains(strings.Join(banners, " "), "no lane health") {
		t.Fatalf("legacy baseline: %v %v %v", comparison, banners, err)
	}
}

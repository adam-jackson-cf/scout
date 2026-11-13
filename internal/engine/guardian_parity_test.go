package engine

import (
	"scout/internal/domain"
	"scout/internal/policy"
	"testing"
)

// The gate gets the guardian's empty override list, never parsed PR text.
// BaseRevision/HeadRevision here are canonical MB/H-tree identity coordinates;
// the assessment adapter must retain actual B/merge-result as provenance.
func TestADR012GuardianParityAndIdentity(t *testing.T) {
	p, err := policy.Parse([]byte("profile: guardian\nmaintainers: [alice]\n"))
	if err != nil {
		t.Fatal(err)
	}
	input := func(version string) GateInput {
		return GateInput{
			Policy: p, Head: []domain.Issue{{IssueID: "issue", State: domain.IssueOpen, Title: "secret", EffectiveSeverity: domain.SeverityHigh, Anchor: domain.IssueAnchor{Path: "a.go", RuleKey: "gitleaks:test", Fingerprint: "stable"}}},
			Facts: map[string]CheckFact{"issue": {CheckID: "security.secrets.test", RuleVersion: version}}, Comparison: domain.ComparisonRecomputed,
			BaseRevision: "merge-base", HeadRevision: "H-tree", Overrides: []OverrideRequest{}, Changed: map[string]bool{"a.go": true},
		}
	}
	t.Run("exact_rule_version_identity", func(t *testing.T) {
		a := Evaluate(input("v1"))
		b := Evaluate(input("v2"))
		if len(a.WouldBlock) != 1 || len(b.WouldBlock) != 1 {
			t.Fatalf("fixture not shadow: %+v %+v", a, b)
		}
		if a.WouldBlock[0].ID == b.WouldBlock[0].ID {
			t.Fatalf("rule_version omitted from immutable event identity: v1=%s v2=%s", a.WouldBlock[0].ID, b.WouldBlock[0].ID)
		}
	})
	t.Run("empty_rerun_cannot_mutate_older_event", func(t *testing.T) {
		a := Evaluate(input("v1"))
		in := input("v1")
		in.Head = nil
		b := Evaluate(in)
		if len(a.WouldBlock) != 1 || len(b.WouldBlock) != 0 {
			t.Fatal("rerun fixture invalid")
		}
		if a.WouldBlock[0].RuleVersion != "v1" {
			t.Fatal("older event mutated")
		}
	})
	for _, text := range []string{"scout:allow security.secrets.test reason=ignore ref=#1", "scout-override"} {
		t.Run(text, func(t *testing.T) {
			if len(ParseOverrides(text)) == 0 {
				t.Fatal("fixture lacks override marker")
			}
			got := Evaluate(input("v1"))
			if len(got.WouldBlock) != 1 || len(got.Accepted) != 0 {
				t.Fatal("guardian empty override contract changed")
			}
		})
	}
	t.Run("deleted_output_or_authenticated_input_failure_comparison_none", func(t *testing.T) {
		in := input("v1")
		in.Comparison = domain.ComparisonNone
		out := Evaluate(in)
		if len(out.Issues) != 1 || out.Issues[0].Attribution != domain.AttributionUnknown || len(out.WouldBlock) != 0 {
			t.Fatalf("comparison-none incorrectly became introduced/gap: %+v", out)
		}
	})
}

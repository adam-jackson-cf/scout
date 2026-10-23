package engine

import (
	"context"
	"strings"
	"testing"

	"scout/internal/domain"
)

func TestResolveBaseUsesImportedBaselineOnlyWhenDigestsAndRevisionMatch(t *testing.T) {
	plan := domain.Plan{ToolDigest: "tool", PolicyDigest: "policy"}
	imported := &BaseState{Issues: []domain.Issue{issue("b", "ruff:E1", "a.py", "f", domain.IssueOpen)}, Revision: "mb", ToolDigest: "tool", PolicyDigest: "policy"}
	cfg := DiffConfig{Baseline: func(context.Context) (*BaseState, error) { return imported, nil }}
	got, comparison, banners, err := resolveBase(context.Background(), cfg, t.TempDir(), "mb", plan, Result{})
	if err != nil || comparison != domain.ComparisonBaseline || len(got) != 1 || len(banners) != 0 {
		t.Fatalf("matching baseline: %v %v %v %v", got, comparison, banners, err)
	}
	// Mismatches fall back to a paired scan; the scratch root is not a git
	// repository, so no base exists and nothing may be classified introduced.
	for name, mutate := range map[string]func(*BaseState){
		"tool digest": func(b *BaseState) { b.ToolDigest = "other" },
		"policy":      func(b *BaseState) { b.PolicyDigest = "other" },
		"revision":    func(b *BaseState) { b.Revision = "elsewhere" },
	} {
		changed := *imported
		mutate(&changed)
		cfg.Baseline = func(context.Context) (*BaseState, error) { return &changed, nil }
		_, comparison, banners, err := resolveBase(context.Background(), cfg, t.TempDir(), "mb", plan, Result{})
		if err != nil || comparison != domain.ComparisonNone || !strings.Contains(strings.Join(banners, " "), "recomputed") {
			t.Fatalf("%s mismatch: %v %v %v", name, comparison, banners, err)
		}
	}
}

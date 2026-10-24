package engine

import (
	"context"
	"strings"
	"testing"

	"scout/internal/domain"
)

func TestResolveBaseUsesImportedBaselineOnlyWhenDigestsAndRevisionMatch(t *testing.T) {
	plan := domain.Plan{ToolDigest: "tool", PolicyDigest: "policy"}
	imported := &BaseState{Issues: []domain.Issue{issue("b", "ruff:E1", "a.py", "f", domain.IssueOpen)}, Revision: "mb", ToolDigest: "tool", PolicyDigest: "policy", Gaps: BaseGaps{}}
	cfg := DiffConfig{Baseline: func(context.Context) (*BaseState, error) { return imported, nil }}
	got, _, comparison, banners, err := resolveBase(context.Background(), cfg, t.TempDir(), "mb", plan, Result{})
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
		_, _, comparison, banners, err := resolveBase(context.Background(), cfg, t.TempDir(), "mb", plan, Result{})
		if err != nil || comparison != domain.ComparisonNone || !strings.Contains(strings.Join(banners, " "), "recomputed") {
			t.Fatalf("%s mismatch: %v %v %v", name, comparison, banners, err)
		}
	}
}

func TestDiffWithoutMergeBaseFailsInsteadOfComparingAgainstBaseTip(t *testing.T) {
	root, _ := deltaRepo(t, map[string]string{"a.py": "x = 1\n"})
	deltaGit(t, root, "checkout", "-q", "--orphan", "unrelated")
	deltaWrite(t, root, "b.py", "y = 2\n")
	deltaGit(t, root, "add", "-A")
	deltaGit(t, root, "commit", "-qm", "unrelated")
	_, err := Diff(context.Background(), DiffConfig{Config: Config{Root: root}, BaseRef: "main"})
	if err == nil || !strings.Contains(err.Error(), "merge base") {
		t.Fatalf("unrelated histories have no merge base and must fail, got %v", err)
	}
}

func TestGuardrailsLaneReportsConfigDeletedInAPRThatOnlyDeletes(t *testing.T) {
	root, base := deltaRepo(t, map[string]string{
		".golangci.yml": "linters:\n  enable:\n    - errcheck\n    - govet\n",
		"a.go":          "package a\n",
	})
	deltaGit(t, root, "rm", "-q", ".golangci.yml")
	deltaGit(t, root, "commit", "-qm", "drop lint config")
	ctx := context.Background()
	changed, err := changedFiles(ctx, root, base)
	if err != nil || len(changed) != 0 {
		t.Fatalf("a deletion-only PR has no added or modified files, got %v %v", changed, err)
	}
	observations, err := guardrailsLane(ctx, root, base, changed).Run(ctx, "run")
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range observations {
		if o.RuleID == "guardrails.CONFIG.strictness-lowered.golangci" && o.Mechanism == "config-removed:golangci" {
			return
		}
	}
	t.Fatalf("deleting the lint config must lower strictness, got %+v", observations)
}

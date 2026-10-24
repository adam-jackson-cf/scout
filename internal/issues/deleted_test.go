package issues

import (
	"fmt"
	"testing"
	"time"

	"scout/internal/domain"
	"scout/internal/state/statetest"
)

// inventoryRun saves a run with the given state, scope and inventory (no
// findings) and correlates it with opts.
func (h *harness) inventoryRun(state domain.RunState, scope domain.Scope, inventory []string, opts ...Option) error {
	h.t.Helper()
	h.runs++
	suffix := fmt.Sprint(h.runs)
	snapshot := domain.Snapshot{RepositoryID: "repository", WorktreeDigest: "worktree-" + suffix, IgnoreRulesDigest: "ignore", ExclusionsDigest: "exclude"}
	scope.SelectedFamilies = []domain.Family{domain.FamilySecurity}
	plan := domain.Plan{PlanID: "plan-" + suffix, RunID: "run-" + suffix, Snapshot: snapshot, Scope: scope, CreatedAt: time.Now().UTC()}
	finished := time.Now().UTC()
	run := domain.Run{RunID: plan.RunID, PlanID: plan.PlanID, Snapshot: snapshot, Scope: scope, State: state, StartedAt: finished, FinishedAt: &finished}
	files := make([]domain.File, 0, len(inventory))
	for _, path := range inventory {
		files = append(files, domain.File{Path: path, Eligible: true})
	}
	statetest.SeedRun(h.t, h.ctx, h.store, plan, run, nil, files)
	_, err := Correlate(h.ctx, h.store, run, nil, opts...)
	return err
}

func TestDeletedFileResolvesOnlyInCompleteInScopeRuns(t *testing.T) {
	repo := domain.Scope{Kind: domain.ScopeRepository}
	elsewhere := domain.Scope{Kind: domain.ScopeSubpath, Paths: []string{"other"}}
	setup := func() (*harness, string) {
		h := newHarness(t)
		return h, h.run([]finding{{path: "a.go", quote: "eval(x)", line: 4}}, covered("a.go"))[0].IssueID
	}

	t.Run("partial run does not resolve", func(t *testing.T) {
		h, id := setup()
		if err := h.inventoryRun(domain.RunPartial, repo, []string{"keep.go"}, WithDeleted([]string{"a.go"})); err != nil {
			t.Fatal(err)
		}
		if h.load(id).State != domain.IssueOpen {
			t.Fatal("partial run resolved a deleted file's issue")
		}
	})
	t.Run("scope excluding the path does not resolve", func(t *testing.T) {
		h, id := setup()
		if err := h.inventoryRun(domain.RunComplete, elsewhere, []string{"other/x.go"}, WithDeleted([]string{"a.go"})); err != nil {
			t.Fatal(err)
		}
		if h.load(id).State != domain.IssueOpen {
			t.Fatal("out-of-scope deletion resolved the issue")
		}
	})
	t.Run("rename follows the issue", func(t *testing.T) {
		h, id := setup()
		// One diff reports a.go as renamed to b.go; a stale deletion claim for the
		// old path must not win over the rename.
		if err := h.inventoryRun(domain.RunComplete, repo, []string{"b.go"}, WithRenames(map[string]string{"a.go": "b.go"}), WithDeleted([]string{"a.go"})); err != nil {
			t.Fatal(err)
		}
		if h.load(id).State != domain.IssueOpen {
			t.Fatal("rename was treated as deletion")
		}
	})
	t.Run("inventory still holding the path is rejected", func(t *testing.T) {
		h, id := setup()
		if err := h.inventoryRun(domain.RunComplete, repo, []string{"a.go"}, WithDeleted([]string{"a.go"})); err == nil {
			t.Fatal("deletion claim contradicted by the inventory was accepted")
		}
		if h.load(id).State != domain.IssueOpen {
			t.Fatal("contradicted deletion resolved the issue")
		}
	})
	t.Run("complete in-scope deletion resolves with file_deleted", func(t *testing.T) {
		h, id := setup()
		if err := h.inventoryRun(domain.RunComplete, repo, []string{"keep.go"}, WithDeleted([]string{"a.go"})); err != nil {
			t.Fatal(err)
		}
		issue := h.load(id)
		if issue.State != domain.IssueResolved || issue.Verification == nil || issue.Verification.Reason != domain.ResolutionFileDeleted || !hasEvent(issue, domain.EventResolved) {
			t.Fatalf("deleted file did not resolve: %#v", issue)
		}
		// Re-introducing the file's finding regresses the same issue.
		if back := h.run([]finding{{path: "a.go", quote: "eval(x)", line: 4}}, covered("a.go"))[0]; back.IssueID != id || h.load(id).State != domain.IssueRegressed {
			t.Fatal("reappearing finding did not regress the resolved issue")
		}
	})
}

func TestChangedMessageNumberRecordsEvidenceDrift(t *testing.T) {
	h := newHarness(t)
	id := h.run([]finding{{path: "a.py", symbol: "complex_a", mechanism: "complex_a is too complex (12 > 10)", quote: "", line: 3}}, covered("a.py"))[0].IssueID
	if _, err := h.store.TriageIssue(h.ctx, id, domain.IssueAccepted, "o", "known", "human", "reviewed"); err != nil {
		t.Fatal(err)
	}
	h.run([]finding{{path: "a.py", symbol: "complex_a", mechanism: "complex_a is too complex (13 > 10)", quote: "", line: 3}}, covered("a.py"))
	issue := h.load(id)
	if issue.State != domain.IssueAccepted || !hasEvent(issue, domain.EventEvidenceDrift) || !hasEvent(issue, domain.EventTriaged) || len(h.all()) != 1 {
		t.Fatalf("numeric message change: %#v", issue)
	}
}

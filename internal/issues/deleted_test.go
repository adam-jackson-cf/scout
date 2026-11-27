package issues

import (
	"fmt"
	"testing"
	"time"

	"scout/internal/domain"
	"scout/internal/state"
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
	statetest.SeedRun(h.t, h.ctx, h.store, statetest.Seed{Plan: plan, Run: run, Files: files})
	_, err := Correlate(h.ctx, h.store, run, nil, opts...)
	return err
}

func TestDeletedFileResolvesOnlyInCompleteInScopeRuns(t *testing.T) {
	repo := domain.Scope{Kind: domain.ScopeRepository}
	elsewhere := domain.Scope{Kind: domain.ScopeSubpath, Paths: []string{"other"}}
	stayOpen := []struct {
		name      string
		state     domain.RunState
		scope     domain.Scope
		inventory []string
		opts      []Option
		rejected  bool
		failure   string
	}{
		{"partial run does not resolve", domain.RunPartial, repo, []string{"keep.go"}, []Option{WithDeleted([]string{"a.go"})}, false, "partial run resolved a deleted file's issue"},
		{"scope excluding the path does not resolve", domain.RunComplete, elsewhere, []string{"other/x.go"}, []Option{WithDeleted([]string{"a.go"})}, false, "out-of-scope deletion resolved the issue"},
		// One diff reports a.go as renamed to b.go; a stale deletion claim for the
		// old path must not win over the rename.
		{"rename follows the issue", domain.RunComplete, repo, []string{"b.go"}, []Option{WithRenames(map[string]string{"a.go": "b.go"}), WithDeleted([]string{"a.go"})}, false, "rename was treated as deletion"},
		{"inventory still holding the path is rejected", domain.RunComplete, repo, []string{"a.go"}, []Option{WithDeleted([]string{"a.go"})}, true, "contradicted deletion resolved the issue"},
	}
	for _, tc := range stayOpen {
		t.Run(tc.name, func(t *testing.T) {
			h, id := newDeletedFileHarness(t)
			requireInventoryOutcome(t, h.inventoryRun(tc.state, tc.scope, tc.inventory, tc.opts...), tc.rejected)
			if h.load(id).State != domain.IssueOpen {
				t.Fatal(tc.failure)
			}
		})
	}
	t.Run("complete in-scope deletion resolves with file_deleted", func(t *testing.T) {
		h, id := newDeletedFileHarness(t)
		if err := h.inventoryRun(domain.RunComplete, repo, []string{"keep.go"}, WithDeleted([]string{"a.go"})); err != nil {
			t.Fatal(err)
		}
		requireResolvedAsDeleted(t, h.load(id))
		// Re-introducing the file's finding regresses the same issue.
		if back := h.run([]finding{{path: "a.go", quote: "eval(x)", line: 4}}, covered("a.go"))[0]; back.IssueID != id || h.load(id).State != domain.IssueRegressed {
			t.Fatal("reappearing finding did not regress the resolved issue")
		}
	})
}

func newDeletedFileHarness(t *testing.T) (*harness, string) {
	t.Helper()
	h := newHarness(t)
	return h, h.run([]finding{{path: "a.go", quote: "eval(x)", line: 4}}, covered("a.go"))[0].IssueID
}

func requireInventoryOutcome(t *testing.T, err error, rejected bool) {
	t.Helper()
	if !rejected {
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	if err == nil {
		t.Fatal("deletion claim contradicted by the inventory was accepted")
	}
}

func requireResolvedAsDeleted(t *testing.T, issue domain.Issue) {
	t.Helper()
	if issue.State != domain.IssueResolved || issue.Verification == nil || issue.Verification.Reason != domain.ResolutionFileDeleted || !hasEvent(issue, domain.EventResolved) {
		t.Fatalf("deleted file did not resolve: %#v", issue)
	}
}

func TestChangedMessageNumberRecordsEvidenceDrift(t *testing.T) {
	h := newHarness(t)
	id := h.run([]finding{{path: "a.py", symbol: "complex_a", mechanism: "complex_a is too complex (12 > 10)", quote: "", line: 3}}, covered("a.py"))[0].IssueID
	if _, err := h.store.TriageIssue(h.ctx, state.Triage{IssueID: id, State: domain.IssueAccepted, Owner: "o", DispositionReason: "known", Actor: "human", Reason: "reviewed"}); err != nil {
		t.Fatal(err)
	}
	h.run([]finding{{path: "a.py", symbol: "complex_a", mechanism: "complex_a is too complex (13 > 10)", quote: "", line: 3}}, covered("a.py"))
	issue := h.load(id)
	if issue.State != domain.IssueAccepted || !hasEvent(issue, domain.EventEvidenceDrift) || !hasEvent(issue, domain.EventTriaged) || len(h.all()) != 1 {
		t.Fatalf("numeric message change: %#v", issue)
	}
}

package state

import (
	"context"
	"errors"
	"testing"
	"time"

	"scout/internal/domain"
)

// A partial run's inventory is a full scan: it is available for rename/delete
// diffs but is never a coverage-reuse baseline.
func TestPartialRunInventoryServesLifecycleButNotReuseBaseline(t *testing.T) {
	ctx := context.Background()
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	plan, run, _, _, _ := progressFixture("partial")
	if err := store.SavePlan(ctx, plan); err != nil {
		t.Fatal(err)
	}
	if err := store.BeginRun(ctx, run); err != nil {
		t.Fatal(err)
	}
	pending, err := store.LoadInProgressRun(ctx, plan.PlanID)
	if err != nil {
		t.Fatal(err)
	}
	finished := time.Now().UTC()
	pending.FinishedAt = &finished
	pending.State = domain.RunPartial
	if err := store.FinishRun(ctx, pending, []domain.File{{Path: "pkg/b.py", Eligible: true}}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.LoadLatestInventory(ctx); !errors.Is(err, ErrNotFound) {
		t.Fatalf("partial inventory became a reuse baseline: %v", err)
	}
	got, files, err := store.LoadLatestRecordedInventory(ctx)
	if err != nil || got.RunID != run.RunID || len(files) != 1 || files[0].Path != "pkg/b.py" {
		t.Fatalf("recorded inventory = %v %v %v", got.RunID, files, err)
	}
}

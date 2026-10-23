package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"scout/internal/clawpatch"
	"scout/internal/domain"
	"scout/internal/policy"
	"scout/internal/state"
)

type memRefreshStore struct{ record state.MappingRefresh }

func (m *memRefreshStore) LoadMappingRefresh(context.Context) (state.MappingRefresh, error) {
	return m.record, nil
}
func (m *memRefreshStore) SaveMappingRefresh(_ context.Context, r state.MappingRefresh) error {
	m.record = r
	return nil
}

type fakeMapper struct {
	calls  int
	result clawpatch.MapResult
	err    error
}

func (f *fakeMapper) Map(_ context.Context, changed bool, source clawpatch.MapSource) (clawpatch.MapResult, error) {
	f.calls++
	if !changed || source != clawpatch.MapSourceAgent {
		panic("refresh must request a changed agent map")
	}
	return f.result, f.err
}

type refreshHarness struct {
	store      *memRefreshStore
	mapper     *fakeMapper
	now        time.Time
	reconciled int
	refresher  *Refresher
}

func newHarness() *refreshHarness {
	h := &refreshHarness{
		store:  &memRefreshStore{},
		mapper: &fakeMapper{result: clawpatch.MapResult{Status: clawpatch.MapMapped, Source: clawpatch.MapSourceAgent, Features: []domain.Feature{{FeatureID: "agent-1", Source: "agent-mapper"}}}},
		now:    time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC),
	}
	h.refresher = &Refresher{
		Store: h.store, Mapper: h.mapper, Now: func() time.Time { return h.now },
		Mapping:   policy.DefaultMapping(),
		Reconcile: func(context.Context, clawpatch.MapResult) error { h.reconciled++; return nil },
	}
	return h
}

func (h *refreshHarness) run(t *testing.T, run RefreshRun) RefreshOutcome {
	t.Helper()
	out, err := h.refresher.Run(context.Background(), run)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

var (
	ownedFeature = domain.Feature{FeatureID: "h1", Source: "python-source-group", OwnedFiles: []domain.FeatureFileRef{{Path: "src/a.go"}}}
	newFile      = []FileChange{{Path: "src/new.go", Status: ChangeAdded}}
	contentOnly  = []FileChange{{Path: "src/a.go", Status: ChangeModified}}
)

func pushRun(changes []FileChange) RefreshRun {
	return RefreshRun{Event: RefreshPush, DefaultBranch: true, Revision: "rev", Changes: changes, Features: []domain.Feature{ownedFeature}}
}

func onboarded(h *refreshHarness) {
	at := h.now.Add(-24 * time.Hour)
	h.store.record.LastAgentRefreshAt = &at
}

func TestOnboardingRefreshesOnce(t *testing.T) {
	h := newHarness()
	if out := h.run(t, pushRun(contentOnly)); out.Status != RefreshDone || h.mapper.calls != 1 || h.reconciled != 1 {
		t.Fatalf("onboarding: %+v calls=%d", out, h.mapper.calls)
	}
	if out := h.run(t, pushRun(contentOnly)); out.Status != RefreshNone || h.mapper.calls != 1 {
		t.Fatalf("second run must not refresh: %+v", out)
	}
}

func TestBurstOfMergesYieldsOneRefreshAfterDebounce(t *testing.T) {
	h := newHarness()
	onboarded(h)
	for i := range 3 {
		h.now = h.now.Add(10 * time.Minute)
		out := h.run(t, pushRun(newFile))
		if out.Status != RefreshPending || h.mapper.calls != 0 {
			t.Fatalf("merge %d inside debounce: %+v", i, out)
		}
	}
	p := h.store.record.Pending
	if p == nil || len(p.Revisions) != 3 || !p.FirstTriggerAt.Before(p.LastTriggerAt) {
		t.Fatalf("pending = %+v", p)
	}
	h.now = h.now.Add(59 * time.Minute)
	if out := h.run(t, RefreshRun{Event: RefreshNightly}); out.Status != RefreshPending {
		t.Fatalf("still inside debounce: %+v", out)
	}
	h.now = h.now.Add(time.Minute)
	if out := h.run(t, RefreshRun{Event: RefreshNightly}); out.Status != RefreshDone || h.mapper.calls != 1 || h.reconciled != 1 {
		t.Fatalf("later run: %+v calls=%d", out, h.mapper.calls)
	}
	h.now = h.now.Add(48 * time.Hour)
	h.run(t, RefreshRun{Event: RefreshNightly})
	if h.mapper.calls != 1 || h.store.record.Pending != nil {
		t.Fatalf("exactly one refresh expected, calls=%d pending=%+v", h.mapper.calls, h.store.record.Pending)
	}
}

func TestContentOnlyMergeRecordsNothing(t *testing.T) {
	h := newHarness()
	onboarded(h)
	h.now = h.now.Add(time.Hour)
	out := h.run(t, pushRun(contentOnly))
	if out.Status != RefreshNone || h.store.record.Pending != nil || h.mapper.calls != 0 {
		t.Fatalf("content-only: %+v %+v", out, h.store.record.Pending)
	}
}

func TestPullRequestAndNonDefaultPushNeverRefresh(t *testing.T) {
	h := newHarness() // not onboarded: even onboarding must wait for a non-PR run
	pr := pushRun(newFile)
	pr.Event = RefreshPR
	other := pushRun(newFile)
	other.DefaultBranch = false
	for _, run := range []RefreshRun{pr, other} {
		if out := h.run(t, run); out.Status != RefreshNone || h.mapper.calls != 0 || h.store.record.Pending != nil {
			t.Fatalf("%s run: %+v", run.Event, out)
		}
	}
}

func TestPredicate(t *testing.T) {
	features := []domain.Feature{{OwnedFiles: []domain.FeatureFileRef{{Path: "src/a.go"}}, Entrypoints: []domain.FeatureEntrypoint{{Path: "src/serve.go"}}}}
	cases := []struct {
		name    string
		changes []FileChange
		want    int
	}{
		{"content only", contentOnly, 0},
		{"new unowned", newFile, 1},
		{"manifest added", []FileChange{{Path: "go.mod", Status: ChangeAdded}}, 1},
		{"manifest modified", []FileChange{{Path: "go.mod", Status: ChangeModified}}, 0},
		{"entrypoint removed", []FileChange{{Path: "cmd/x/main.go", Status: ChangeDeleted}}, 1},
		{"feature entrypoint removed", []FileChange{{Path: "src/serve.go", Status: ChangeDeleted}}, 1},
		{"owned removed", []FileChange{{Path: "src/a.go", Status: ChangeDeleted}}, 1},
		{"unowned removed", []FileChange{{Path: "docs/x.md", Status: ChangeDeleted}}, 0},
		{"committed Scout state", []FileChange{{Path: ".scout/clawpatch/features/feat_x.json", Status: ChangeAdded}, {Path: ".scout/policy.yaml", Status: ChangeAdded}, {Path: ".scout/artifacts/00/ab", Status: ChangeDeleted}}, 0},
	}
	for _, c := range cases {
		if got := RefreshTriggers(c.changes, features); len(got) != c.want {
			t.Errorf("%s: %v", c.name, got)
		}
	}
}

func TestNoModelCLISkipsKeepingPending(t *testing.T) {
	h := newHarness()
	onboarded(h)
	h.run(t, pushRun(newFile))
	h.now = h.now.Add(2 * time.Hour)
	h.mapper.result = clawpatch.MapResult{Status: clawpatch.MapSkipped, Source: clawpatch.MapSourceAgent, Reason: clawpatch.SkippedNoModelCLI}
	out := h.run(t, RefreshRun{Event: RefreshNightly})
	if out.Status != RefreshSkipped || out.Reason != "skipped (no model CLI)" || h.store.record.Pending == nil || h.reconciled != 0 {
		t.Fatalf("skip: %+v pending=%+v", out, h.store.record.Pending)
	}
}

func TestAgentFailureKeepsHeuristicAndPending(t *testing.T) {
	for name, result := range map[string]clawpatch.MapResult{
		"fallback":      {Status: clawpatch.MapMapped, Source: clawpatch.MapSourceHeuristic, ExplorationIncomplete: true, Features: []domain.Feature{{FeatureID: "heur", Source: "python-source-group"}}},
		"zero features": {Status: clawpatch.MapMapped, Source: clawpatch.MapSourceAgent, ExplorationIncomplete: true},
	} {
		h := newHarness()
		onboarded(h)
		h.run(t, pushRun(newFile))
		h.now = h.now.Add(2 * time.Hour)
		h.mapper.result = result
		out := h.run(t, RefreshRun{Event: RefreshNightly})
		if out.Status != RefreshIncomplete || !out.ExplorationIncomplete || h.store.record.Pending == nil || h.reconciled != 0 {
			t.Fatalf("%s: %+v", name, out)
		}
	}
}

func TestBlockedAgentKeepsPending(t *testing.T) {
	h := newHarness()
	onboarded(h)
	h.run(t, pushRun(newFile))
	h.now = h.now.Add(2 * time.Hour)
	h.mapper.result = clawpatch.MapResult{Status: clawpatch.MapBlocked}
	h.mapper.err = errors.New("egress is not authorized")
	out := h.run(t, RefreshRun{Event: RefreshNightly})
	if out.Status != RefreshBlocked || h.store.record.Pending == nil {
		t.Fatalf("blocked: %+v", out)
	}
}

func TestForceRefreshesOnceRegardlessOfTriggersOrDebounce(t *testing.T) {
	h := newHarness()
	onboarded(h)
	out := h.run(t, RefreshRun{Force: true})
	if out.Status != RefreshDone || h.mapper.calls != 1 || h.reconciled != 1 {
		t.Fatalf("force: %+v", out)
	}
	if h.run(t, RefreshRun{Force: true}); h.mapper.calls != 2 {
		t.Fatal("each force is one refresh")
	}
}

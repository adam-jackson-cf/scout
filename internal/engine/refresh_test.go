package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"scout/internal/clawpatch"
	"scout/internal/domain"
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

func pushRun() RefreshRun {
	return RefreshRun{Event: domain.TriggerPush, DefaultBranch: true}
}

func onboarded(h *refreshHarness) {
	at := h.now.Add(-24 * time.Hour)
	h.store.record.LastAgentRefreshAt = &at
}

func TestOnboardingRefreshesOnce(t *testing.T) {
	h := newHarness()
	if out := h.run(t, pushRun()); out.Status != RefreshDone || h.mapper.calls != 1 || h.reconciled != 1 {
		t.Fatalf("onboarding: %+v calls=%d", out, h.mapper.calls)
	}
	for _, run := range []RefreshRun{pushRun(), {Event: domain.TriggerNightly}} {
		if out := h.run(t, run); out.Status != RefreshNone || h.mapper.calls != 1 {
			t.Fatalf("%s after onboarding must not map: %+v", run.Event, out)
		}
	}
}

func TestBurstOfMergesNeverRemaps(t *testing.T) {
	h := newHarness()
	h.run(t, pushRun())
	for range 3 {
		h.now = h.now.Add(2 * time.Hour)
		if out := h.run(t, pushRun()); out.Status != RefreshNone {
			t.Fatalf("merge after onboarding: %+v", out)
		}
	}
	if h.mapper.calls != 1 {
		t.Fatalf("only onboarding maps, calls=%d", h.mapper.calls)
	}
}

func TestPullRequestAndNonDefaultPushNeverRefresh(t *testing.T) {
	h := newHarness() // not onboarded: even onboarding must wait for a non-PR run
	pr := pushRun()
	pr.Event = domain.TriggerPR
	other := pushRun()
	other.DefaultBranch = false
	for _, run := range []RefreshRun{pr, other} {
		if out := h.run(t, run); out.Status != RefreshNone || h.mapper.calls != 0 || h.store.record.LastAgentRefreshAt != nil {
			t.Fatalf("%s run: %+v", run.Event, out)
		}
	}
}

func TestUnknownEventIsAnError(t *testing.T) {
	h := newHarness()
	if _, err := h.refresher.Run(context.Background(), RefreshRun{Event: "weekly"}); err == nil || h.mapper.calls != 0 {
		t.Fatalf("unknown event: err=%v calls=%d", err, h.mapper.calls)
	}
}

func TestNoModelCLIKeepsOnboardingOpen(t *testing.T) {
	h := newHarness()
	h.mapper.result = clawpatch.MapResult{Status: clawpatch.MapSkipped, Source: clawpatch.MapSourceAgent, Reason: clawpatch.SkippedNoModelCLI}
	out := h.run(t, RefreshRun{Event: domain.TriggerNightly})
	if out.Status != RefreshSkipped || out.Reason != "skipped (no model CLI)" || h.store.record.LastAgentRefreshAt != nil || h.reconciled != 0 {
		t.Fatalf("skip: %+v record=%+v", out, h.store.record)
	}
}

func TestAgentFailureKeepsHeuristicAndOnboardingOpen(t *testing.T) {
	for name, result := range map[string]clawpatch.MapResult{
		"fallback":      {Status: clawpatch.MapMapped, Source: clawpatch.MapSourceHeuristic, ExplorationIncomplete: true, Features: []domain.Feature{{FeatureID: "heur", Source: "python-source-group"}}},
		"zero features": {Status: clawpatch.MapMapped, Source: clawpatch.MapSourceAgent, ExplorationIncomplete: true},
	} {
		h := newHarness()
		h.mapper.result = result
		out := h.run(t, RefreshRun{Event: domain.TriggerNightly})
		if out.Status != RefreshIncomplete || !out.ExplorationIncomplete || h.store.record.LastAgentRefreshAt != nil || h.reconciled != 0 {
			t.Fatalf("%s: %+v", name, out)
		}
	}
}

func TestBlockedAgentKeepsOnboardingOpen(t *testing.T) {
	h := newHarness()
	good := h.mapper.result
	h.mapper.result = clawpatch.MapResult{Status: clawpatch.MapBlocked}
	h.mapper.err = errors.New("egress is not authorized")
	out := h.run(t, RefreshRun{Event: domain.TriggerNightly})
	if out.Status != RefreshBlocked || h.store.record.LastAgentRefreshAt != nil {
		t.Fatalf("blocked: %+v", out)
	}
	h.mapper.result, h.mapper.err = good, nil
	if out := h.run(t, pushRun()); out.Status != RefreshDone || h.mapper.calls != 2 {
		t.Fatalf("next default-branch push must retry onboarding: %+v calls=%d", out, h.mapper.calls)
	}
}

func TestForceRefreshesOnceEvenWhenOnboarded(t *testing.T) {
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

package engine

import (
	"context"
	"fmt"
	"time"

	"scout/internal/clawpatch"
	"scout/internal/domain"
	"scout/internal/state"
)

// RefreshStatus is the outcome of one agent-map evaluation.
type RefreshStatus string

const (
	RefreshNone       RefreshStatus = "none"       // nothing due: onboarded, or not an onboarding run
	RefreshDone       RefreshStatus = "refreshed"  // one agent map ran and reconciled
	RefreshSkipped    RefreshStatus = "skipped"    // no model CLI; onboarding stays open, run not failed
	RefreshIncomplete RefreshStatus = "incomplete" // agent failed or found nothing; onboarding stays open
	RefreshBlocked    RefreshStatus = "blocked"    // agent mapping not authorized; onboarding stays open
)

// RefreshStore is the durable refresh state. *state.Store satisfies it.
type RefreshStore interface {
	LoadMappingRefresh(ctx context.Context) (state.MappingRefresh, error)
	SaveMappingRefresh(ctx context.Context, record state.MappingRefresh) error
}

// AgentMapper runs the Clawpatch mapper. *clawpatch.Client satisfies it.
type AgentMapper interface {
	Map(ctx context.Context, inputsChanged bool, source clawpatch.MapSource) (clawpatch.MapResult, error)
}

// Refresher decides and performs agent maps. The clock is injected.
type Refresher struct {
	Store     RefreshStore
	Mapper    AgentMapper
	Reconcile func(ctx context.Context, mapped clawpatch.MapResult) error
	Now       func() time.Time
}

// RefreshRun describes the run being evaluated.
type RefreshRun struct {
	Event         domain.LaneTrigger
	DefaultBranch bool // push runs: true only for pushes to the default branch
	// Force is `scout map --refresh`: run the agent map now, ignoring the event.
	Force bool
}

// RefreshOutcome reports what happened.
type RefreshOutcome struct {
	Status                RefreshStatus        `json:"status"`
	Reason                string               `json:"reason,omitempty"`
	ExplorationIncomplete bool                 `json:"explorationIncomplete,omitempty"`
	Map                   *clawpatch.MapResult `json:"map,omitempty"`
}

// Run performs at most one agent map: at onboarding (default-branch pushes and
// nightly runs attempt it until one has completed) or when forced. PR runs and
// non-default pushes never map. Design §4.
func (r *Refresher) Run(ctx context.Context, run RefreshRun) (RefreshOutcome, error) {
	record, err := r.Store.LoadMappingRefresh(ctx)
	if err != nil {
		return RefreshOutcome{}, err
	}
	if run.Force {
		return r.refresh(ctx, record)
	}
	switch {
	case run.Event == domain.TriggerPR:
		return RefreshOutcome{Status: RefreshNone, Reason: "pull-request runs reuse the active snapshot"}, nil
	case run.Event == domain.TriggerPush && !run.DefaultBranch:
		return RefreshOutcome{Status: RefreshNone, Reason: "only default-branch pushes onboard the map"}, nil
	case run.Event != domain.TriggerPush && run.Event != domain.TriggerNightly:
		return RefreshOutcome{}, fmt.Errorf("unknown refresh event %q", run.Event)
	case record.LastAgentRefreshAt != nil:
		return RefreshOutcome{Status: RefreshNone, Reason: "onboarded; run `scout map --refresh` to remap"}, nil
	}
	return r.refresh(ctx, record)
}

func (r *Refresher) refresh(ctx context.Context, record state.MappingRefresh) (RefreshOutcome, error) {
	mapped, err := r.Mapper.Map(ctx, true, clawpatch.MapSourceAgent)
	out := RefreshOutcome{Status: RefreshIncomplete, Reason: mapped.Reason, Map: &mapped, ExplorationIncomplete: mapped.ExplorationIncomplete}
	switch {
	case mapped.Status == clawpatch.MapSkipped:
		out.Status = RefreshSkipped
		return out, nil
	case err != nil || mapped.Status == clawpatch.MapBlocked:
		out.Status = RefreshBlocked
		if err != nil {
			out.Reason = err.Error()
		}
		return out, nil
	case mapped.ExplorationIncomplete || mapped.Source != clawpatch.MapSourceAgent || len(mapped.Features) == 0:
		out.ExplorationIncomplete = true
		return out, nil
	}
	if r.Reconcile != nil {
		if err := r.Reconcile(ctx, mapped); err != nil {
			return out, fmt.Errorf("reconcile areas: %w", err)
		}
	}
	now := r.Now().UTC()
	record.LastAgentRefreshAt = &now
	if err := r.Store.SaveMappingRefresh(ctx, record); err != nil {
		return out, err
	}
	out.Status, out.Reason = RefreshDone, ""
	return out, nil
}

package engine

import (
	"context"

	"scout/internal/association"
	"scout/internal/domain"
	"scout/internal/policy"
	"scout/internal/state"
)

// RunAssociationLane runs the optional push-only association signal lane.
func RunAssociationLane(ctx context.Context, store *state.Store, root string, trigger domain.LaneTrigger, options association.LaneOptions) ([]domain.LaneHealth, error) {
	return association.Execute(ctx, store, root, trigger, options)
}

func (c Config) associationGuard() policy.AssociationGuard {
	if c.AssociationGuard == (policy.AssociationGuard{}) {
		return policy.DefaultMapping().Association.Guard
	}
	return c.AssociationGuard
}

// associationOutput replays persisted decisions without changing assignments or
// running a mapper. The returned rows belong only to this output.
func associationOutput(ctx context.Context, store *state.Store, cfg policy.AssociationGuard) ([]domain.Association, association.GuardState, error) {
	rows, err := store.ListAssociations(ctx)
	if err != nil {
		return nil, association.GuardState{}, err
	}
	refresh, err := store.LoadMappingRefresh(ctx)
	if err != nil {
		return nil, association.GuardState{}, err
	}
	guard := association.Guard(rows, refresh.LastAgentRefreshAt, cfg, association.GuardState{})
	if rows == nil {
		rows = []domain.Association{}
	}
	for i := range rows {
		rows[i].CatalogueStale = guard.CatalogueStale
	}
	return rows, guard, nil
}

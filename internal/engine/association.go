package engine

import (
	"context"

	"scout/internal/association"
	"scout/internal/domain"
	"scout/internal/state"
)

// RunAssociationLane runs the optional push-only association signal lane.
func RunAssociationLane(ctx context.Context, store *state.Store, root string, trigger domain.LaneTrigger, options association.LaneOptions) (domain.LaneHealth, error) {
	return association.Execute(ctx, store, root, trigger, options)
}

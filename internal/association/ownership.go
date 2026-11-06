package association

import (
	"context"
	"slices"
	"sort"

	"scout/internal/areas"
	"scout/internal/domain"
	"scout/internal/state"
)

// storeOwnership snapshots labels without changing the agent-map FileSets.
type storeOwnership struct {
	owners    map[string][]string
	pending   map[string][]string
	heuristic map[string]bool
	siblings  map[SiblingKey][]string
}

// NewStoreOwnership combines active FileSets with nonsuperseded association
// labels targeting active areas. Stale decisions remain labels, not retries.
func NewStoreOwnership(ctx context.Context, store *state.Store, features []domain.Feature) (Ownership, error) {
	catalogue, err := store.ListAreas(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := store.ListAssociations(ctx)
	if err != nil {
		return nil, err
	}
	own := &storeOwnership{
		owners:    make(map[string][]string),
		pending:   make(map[string][]string),
		heuristic: make(map[string]bool),
		siblings:  make(map[SiblingKey][]string),
	}
	active := make(map[string]bool, len(catalogue))
	for _, area := range catalogue {
		switch {
		case area.Active():
			active[area.ScoutAreaID] = true
			for _, path := range area.FileSet {
				own.owners[path] = append(own.owners[path], area.ScoutAreaID)
			}
		case area.Status == areas.StatusPending && area.RetiredSnapshot == "":
			for _, path := range area.FileSet {
				own.pending[path] = append(own.pending[path], area.ScoutAreaID)
			}
		}
	}
	for _, row := range rows {
		if row.Superseded {
			continue
		}
		var ids []string
		switch row.Outcome {
		case domain.AssociationAssigned:
			ids = row.AreaIDs
		case domain.AssociationInherited:
			ids = row.InheritedAreaIDs
		default:
			continue
		}
		for _, id := range ids {
			if !active[id] {
				continue
			}
			for _, path := range row.Files {
				own.owners[path] = append(own.owners[path], id)
			}
		}
	}
	for _, feature := range features {
		if areas.IsAgent(feature) {
			continue
		}
		for _, ref := range feature.OwnedFiles {
			own.heuristic[ref.Path] = true
		}
	}
	for path, ids := range own.owners {
		sort.Strings(ids)
		own.owners[path] = slices.Compact(ids)
		key := SiblingKeyOf(path)
		own.siblings[key] = append(own.siblings[key], path)
	}
	for path, ids := range own.pending {
		sort.Strings(ids)
		own.pending[path] = slices.Compact(ids)
	}
	for _, paths := range own.siblings {
		sort.Strings(paths)
	}
	return own, nil
}

func (o *storeOwnership) AreaOwners(path string) []string { return o.owners[path] }

func (o *storeOwnership) PendingAreas(path string) []string { return o.pending[path] }

func (o *storeOwnership) HeuristicOnly(path string) bool { return o.heuristic[path] }

func (o *storeOwnership) SiblingFiles(key SiblingKey) []string { return o.siblings[key] }

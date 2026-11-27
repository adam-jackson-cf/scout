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
	active := own.addCatalogue(catalogue)
	for _, row := range rows {
		own.addLabel(row, active)
	}
	own.addHeuristic(features)
	own.normalise()
	return own, nil
}

func (o *storeOwnership) addCatalogue(catalogue []areas.Area) map[string]bool {
	active := make(map[string]bool, len(catalogue))
	for _, area := range catalogue {
		switch {
		case area.Active():
			active[area.ScoutAreaID] = true
			for _, path := range area.FileSet {
				o.owners[path] = append(o.owners[path], area.ScoutAreaID)
			}
		case area.Status == areas.StatusPending && area.RetiredSnapshot == "":
			for _, path := range area.FileSet {
				o.pending[path] = append(o.pending[path], area.ScoutAreaID)
			}
		}
	}
	return active
}

func (o *storeOwnership) addLabel(row domain.Association, active map[string]bool) {
	if row.Superseded {
		return
	}
	var ids []string
	switch row.Outcome {
	case domain.AssociationAssigned:
		ids = row.AreaIDs
	case domain.AssociationInherited:
		ids = row.InheritedAreaIDs
	default:
		return
	}
	for _, id := range ids {
		if !active[id] {
			continue
		}
		for _, path := range row.Files {
			o.owners[path] = append(o.owners[path], id)
		}
	}
}

func (o *storeOwnership) addHeuristic(features []domain.Feature) {
	for _, feature := range features {
		if areas.IsAgent(feature) {
			continue
		}
		for _, ref := range feature.OwnedFiles {
			o.heuristic[ref.Path] = true
		}
	}
}

func (o *storeOwnership) normalise() {
	for path, ids := range o.owners {
		sort.Strings(ids)
		o.owners[path] = slices.Compact(ids)
		key := SiblingKeyOf(path)
		o.siblings[key] = append(o.siblings[key], path)
	}
	for path, ids := range o.pending {
		sort.Strings(ids)
		o.pending[path] = slices.Compact(ids)
	}
	for _, paths := range o.siblings {
		sort.Strings(paths)
	}
}

func (o *storeOwnership) AreaOwners(path string) []string { return o.owners[path] }

func (o *storeOwnership) PendingAreas(path string) []string { return o.pending[path] }

func (o *storeOwnership) HeuristicOnly(path string) bool { return o.heuristic[path] }

func (o *storeOwnership) SiblingFiles(key SiblingKey) []string { return o.siblings[key] }

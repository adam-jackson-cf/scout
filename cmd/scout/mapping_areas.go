package main

import (
	"context"
	"path/filepath"

	"scout/internal/areas"
	"scout/internal/clawpatch"
	"scout/internal/inventory"
)

// mappingAreas lists the area labels (design §4); retired areas are included
// with their retiredSnapshot so triage can see why a label disappeared.
func mappingAreas(ctx context.Context, flags *commonFlags) int {
	store, err := openStore(flags)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	defer store.Close()
	list, err := store.ListAreas(ctx)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	if list == nil {
		list = []areas.Area{}
	}
	return emit(flags, "complete", exitComplete, list, nil)
}

// mappingUnowned lists the directory scopes holding files no heuristic feature
// owns: the mapping-coverage gap, keyed by the stable directory scope.
func mappingUnowned(ctx context.Context, flags *commonFlags) int {
	root, err := filepath.Abs(flags.root)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	features, err := clawpatch.LoadFeatures(filepath.Join(root, ".scout", "clawpatch"))
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	scope, err := scopeOptions("repository", nil, "", "", nil)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	scan, err := inventory.Scan(ctx, root, inventory.ScanOptions{Scope: scope, Features: features})
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	paths := make([]string, len(scan.Files))
	for i, file := range scan.Files {
		paths[i] = file.Path
	}
	scopes := areas.UnownedScopes(paths, areas.HeuristicOwner(features))
	if scopes == nil {
		scopes = []areas.DirScope{}
	}
	return emit(flags, "complete", exitComplete, scopes, nil)
}

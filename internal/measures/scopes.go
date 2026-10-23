package measures

import (
	"sort"
	"strings"

	"scout/internal/domain"
)

// Scope is one measurement scope (design §4): a heuristic feature, or a
// directory for files no feature owns. Paths are sorted repository-relative
// file paths.
type Scope struct {
	Kind   string // domain.SignalScopeFeature | domain.SignalScopeDirectory
	Anchor string // "feature:<id>" or the directory
	Paths  []string
}

// BuildScopes assigns files to scopes. A file belongs to every feature that
// owns it (features may overlap) and, when no feature owns it, to its
// directory scope: the directory truncated to two components below the source
// root. The source root is the repository root, or `src/` when the path starts
// with it; a series key therefore never changes while the map version does not.
// Every feature in features yields a scope, even one with no files, so absence
// of population is reported rather than hidden. Output is sorted by anchor.
func BuildScopes(files []domain.File, features []domain.Feature) []Scope {
	byAnchor := map[string]*Scope{}
	add := func(kind, anchor, path string) {
		scope := byAnchor[anchor]
		if scope == nil {
			scope = &Scope{Kind: kind, Anchor: anchor}
			byAnchor[anchor] = scope
		}
		if path != "" {
			scope.Paths = append(scope.Paths, path)
		}
	}
	for _, feature := range features {
		add(domain.SignalScopeFeature, FeatureAnchor(feature.FeatureID), "")
	}
	for _, file := range files {
		if len(file.OwnedFeatureIDs) == 0 {
			add(domain.SignalScopeDirectory, DirectoryScope(file.Path), file.Path)
			continue
		}
		for _, id := range file.OwnedFeatureIDs {
			add(domain.SignalScopeFeature, FeatureAnchor(id), file.Path)
		}
	}
	scopes := make([]Scope, 0, len(byAnchor))
	for _, scope := range byAnchor {
		sort.Strings(scope.Paths)
		scopes = append(scopes, *scope)
	}
	sort.Slice(scopes, func(i, j int) bool { return scopes[i].Anchor < scopes[j].Anchor })
	return scopes
}

// FeatureAnchor is the scope anchor of a feature.
func FeatureAnchor(featureID string) string { return "feature:" + featureID }

// DirectoryScope truncates a file's directory to two components below the
// source root. Top-level files map to ".".
func DirectoryScope(path string) string {
	index := strings.LastIndex(path, "/")
	if index < 0 {
		return "."
	}
	parts := strings.Split(path[:index], "/")
	keep := 2
	if parts[0] == "src" {
		keep = 3
	}
	if keep > len(parts) {
		keep = len(parts)
	}
	return strings.Join(parts[:keep], "/")
}

// PopulationDigest is the sha256 of the sorted (path, content digest) pairs of
// the eligible population. It changes exactly when population membership or an
// eligible file's content changes.
func PopulationDigest(files []domain.File) string {
	lines := make([]string, 0, len(files))
	for _, file := range files {
		lines = append(lines, file.Path+"\x00"+file.ContentDigest+"\n")
	}
	sort.Strings(lines)
	return domain.Digest(lines...)
}

// SnapshotKey identifies the measured tree in a signal: the commit when the
// tree is clean, else the worktree digest.
func SnapshotKey(snapshot domain.Snapshot) string {
	if snapshot.Revision != "" && !snapshot.Dirty {
		return snapshot.Revision
	}
	return "worktree:" + snapshot.WorktreeDigest
}

package inventory

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"scout/internal/domain"
)

// ChangeKind describes a source-level delta. A rename preserves both paths and
// is detected from unique content fingerprints rather than Git metadata alone.
type ChangeKind string

const (
	ChangeAdded    ChangeKind = "added"
	ChangeModified ChangeKind = "modified"
	ChangeRemoved  ChangeKind = "removed"
	ChangeRenamed  ChangeKind = "renamed"
)

// FileChange preserves the current record when it exists and the prior record
// for removals, allowing callers to avoid scanning deleted content.
type FileChange struct {
	Kind     ChangeKind   `json:"kind"`
	Current  *domain.File `json:"current,omitempty"`
	Previous *domain.File `json:"previous,omitempty"`
}

// DependencyEdge is an externally recorded, justified impact relationship.
// Inventory never infers a semantic dependency graph. InputPaths names the
// manifest/config/shared input that makes the relationship applicable.
type DependencyEdge struct {
	FromFeatureID string   `json:"fromFeatureId"`
	ToFeatureID   string   `json:"toFeatureId"`
	Reason        string   `json:"reason"`
	InputPaths    []string `json:"inputPaths"`
}

// FeatureInvalidation explains why a canonical feature was scheduled despite
// its own Clawpatch status possibly remaining unchanged.
type FeatureInvalidation struct {
	FeatureID string `json:"featureId"`
	Reason    string `json:"reason"`
	Path      string `json:"path,omitempty"`
}

// DiffInput compares prior and current immutable indexes and canonical maps.
type DiffInput struct {
	PreviousFiles    []domain.File
	CurrentFiles     []domain.File
	PreviousFeatures []domain.Feature
	CurrentFeatures  []domain.Feature
	Dependencies     []DependencyEdge
}

// DiffResult is deterministic: all collections are path/ID sorted and every
// mapping gap or dependent invalidation has an explicit reason.
type DiffResult struct {
	Changes            []FileChange          `json:"changes"`
	Added              []domain.File         `json:"added,omitempty"`
	Modified           []domain.File         `json:"modified,omitempty"`
	Removed            []domain.File         `json:"removed,omitempty"`
	Renamed            []domain.File         `json:"renamed,omitempty"`
	AffectedFeatureIDs []string              `json:"affectedFeatureIds,omitempty"`
	Invalidations      []FeatureInvalidation `json:"invalidations,omitempty"`
	MappingGaps        []MappingGap          `json:"mappingGaps,omitempty"`
	Diagnostics        []Diagnostic          `json:"diagnostics,omitempty"`
}

// Diff compares old and current files, detects content-stable renames, and
// marks only canonical feature inputs or explicit dependency edges as affected.
func Diff(input DiffInput) (DiffResult, error) {
	result := DiffResult{}
	previous, err := indexedFiles(input.PreviousFiles)
	if err != nil {
		return result, fmt.Errorf("previous index: %w", err)
	}
	current, err := indexedFiles(input.CurrentFiles)
	if err != nil {
		return result, fmt.Errorf("current index: %w", err)
	}
	priorFeatures, featureErr := indexedFeatures(input.PreviousFeatures)
	currentFeatures, currentFeatureErr := indexedFeatures(input.CurrentFeatures)
	if featureErr != nil || currentFeatureErr != nil {
		return result, errors.Join(featureErr, currentFeatureErr)
	}

	var added, removed []domain.File
	for path, now := range current {
		old, exists := previous[path]
		if !exists {
			added = append(added, now)
			continue
		}
		if changedFile(old, now) {
			result.Modified = append(result.Modified, now)
			result.Changes = append(result.Changes, FileChange{Kind: ChangeModified, Current: filePointer(now), Previous: filePointer(old)})
			addLinkedInvalidations(&result, now, "canonical file input changed", now.Path)
			addLinkedInvalidations(&result, old, "canonical file input changed", now.Path)
		}
	}
	for path, old := range previous {
		if _, exists := current[path]; !exists {
			removed = append(removed, old)
		}
	}

	renamedOld, renamedNew := detectRenames(removed, added)
	for _, now := range added {
		if old, exists := renamedNew[now.Path]; exists {
			now.PreviousPath = old.Path
			result.Renamed = append(result.Renamed, now)
			result.Changes = append(result.Changes, FileChange{Kind: ChangeRenamed, Current: filePointer(now), Previous: filePointer(old)})
			addLinkedInvalidations(&result, now, "canonical file input renamed", now.Path)
			addLinkedInvalidations(&result, old, "canonical file input renamed", old.Path)
			continue
		}
		result.Added = append(result.Added, now)
		result.Changes = append(result.Changes, FileChange{Kind: ChangeAdded, Current: filePointer(now)})
		addLinkedInvalidations(&result, now, "canonical file input added", now.Path)
	}
	for _, old := range removed {
		if _, exists := renamedOld[old.Path]; exists {
			continue
		}
		result.Removed = append(result.Removed, old)
		result.Changes = append(result.Changes, FileChange{Kind: ChangeRemoved, Previous: filePointer(old)})
		addLinkedInvalidations(&result, old, "canonical file input removed", old.Path)
	}

	for id, feature := range currentFeatures {
		previousFeature, existed := priorFeatures[id]
		if !existed {
			addInvalidation(&result, FeatureInvalidation{FeatureID: id, Reason: "new canonical Clawpatch feature"})
			continue
		}
		if FeatureFingerprint(feature) != FeatureFingerprint(previousFeature) {
			addInvalidation(&result, FeatureInvalidation{FeatureID: id, Reason: "canonical feature inputs changed"})
		}
	}
	for id := range priorFeatures {
		if _, exists := currentFeatures[id]; !exists {
			result.MappingGaps = append(result.MappingGaps, MappingGap{FeatureID: id, Reason: "previous canonical feature is absent from current map"})
		}
	}

	changedPaths := changedPaths(result)
	applyDependencyClosure(&result, input.Dependencies, changedPaths)
	result.MappingGaps = append(result.MappingGaps, mappingGaps(input.CurrentFiles, input.CurrentFeatures)...)
	result.AffectedFeatureIDs = invalidatedIDs(result.Invalidations)
	sortDiff(&result)
	return result, nil
}

// FeatureFingerprint hashes only substantive canonical fields. It intentionally
// excludes status, locks, timestamps, findings, patch IDs, and history.
func FeatureFingerprint(feature domain.Feature) string {
	parts := []string{fingerprintTuple("feature", feature.FeatureID, feature.Title, feature.Summary, feature.Kind, feature.Source, string(feature.Confidence))}
	for _, entry := range feature.Entrypoints {
		parts = append(parts, fingerprintTuple("entry", entry.Path, stringValue(entry.Symbol), stringValue(entry.Route), stringValue(entry.Command)))
	}
	for _, ref := range feature.OwnedFiles {
		parts = append(parts, fingerprintTuple("owned", ref.Path, ref.Reason))
	}
	for _, ref := range feature.ContextFiles {
		parts = append(parts, fingerprintTuple("context", ref.Path, ref.Reason))
	}
	for _, ref := range feature.Tests {
		parts = append(parts, fingerprintTuple("test", ref.Path, stringValue(ref.Command)))
	}
	for _, tag := range feature.Tags {
		parts = append(parts, fingerprintTuple("tag", tag))
	}
	for _, boundary := range feature.TrustBoundaries {
		parts = append(parts, fingerprintTuple("boundary", boundary))
	}
	return digestStrings(parts)
}

func indexedFiles(files []domain.File) (map[string]domain.File, error) {
	index := make(map[string]domain.File, len(files))
	for _, file := range files {
		if err := file.Validate(); err != nil {
			return nil, err
		}
		if _, exists := index[file.Path]; exists {
			return nil, fmt.Errorf("duplicate file %q", file.Path)
		}
		index[file.Path] = file
	}
	return index, nil
}

func indexedFeatures(features []domain.Feature) (map[string]domain.Feature, error) {
	index := make(map[string]domain.Feature, len(features))
	for _, feature := range features {
		if err := feature.Validate(); err != nil {
			return nil, err
		}
		if _, exists := index[feature.FeatureID]; exists {
			return nil, fmt.Errorf("duplicate feature %q", feature.FeatureID)
		}
		index[feature.FeatureID] = feature
	}
	return index, nil
}

func changedFile(old, now domain.File) bool {
	return old.ContentDigest != now.ContentDigest || old.Eligible != now.Eligible || old.Excluded != now.Excluded || old.ExclusionReason != now.ExclusionReason || old.Language != now.Language
}

func detectRenames(removed, added []domain.File) (map[string]domain.File, map[string]domain.File) {
	oldByDigest, newByDigest := map[string][]domain.File{}, map[string][]domain.File{}
	for _, file := range removed {
		if file.ContentDigest != "" {
			oldByDigest[file.ContentDigest] = append(oldByDigest[file.ContentDigest], file)
		}
	}
	for _, file := range added {
		if file.ContentDigest != "" {
			newByDigest[file.ContentDigest] = append(newByDigest[file.ContentDigest], file)
		}
	}
	oldMatches, newMatches := map[string]domain.File{}, map[string]domain.File{}
	for digest, oldFiles := range oldByDigest {
		newFiles := newByDigest[digest]
		if len(oldFiles) == 1 && len(newFiles) == 1 {
			oldMatches[oldFiles[0].Path] = newFiles[0]
			newMatches[newFiles[0].Path] = oldFiles[0]
		}
	}
	return oldMatches, newMatches
}

func addLinkedInvalidations(result *DiffResult, file domain.File, reason, path string) {
	for _, id := range append(append([]string{}, file.OwnedFeatureIDs...), file.ContextFeatureIDs...) {
		addInvalidation(result, FeatureInvalidation{FeatureID: id, Reason: reason, Path: path})
	}
}

func addInvalidation(result *DiffResult, invalidation FeatureInvalidation) {
	if invalidation.FeatureID == "" {
		return
	}
	for _, existing := range result.Invalidations {
		if existing == invalidation {
			return
		}
	}
	result.Invalidations = append(result.Invalidations, invalidation)
}

func changedPaths(result DiffResult) map[string]struct{} {
	paths := map[string]struct{}{}
	for _, change := range result.Changes {
		if change.Current != nil {
			paths[change.Current.Path] = struct{}{}
		}
		if change.Previous != nil {
			paths[change.Previous.Path] = struct{}{}
		}
	}
	return paths
}

func applyDependencyClosure(result *DiffResult, edges []DependencyEdge, changedPaths map[string]struct{}) {
	if len(changedPaths) == 0 {
		return
	}
	adjacent := make(map[string][]DependencyEdge)
	for _, edge := range edges {
		if edge.FromFeatureID == "" || edge.ToFeatureID == "" || strings.TrimSpace(edge.Reason) == "" || len(edge.InputPaths) == 0 {
			continue
		}
		for _, path := range edge.InputPaths {
			if _, changed := changedPaths[path]; changed {
				adjacent[edge.FromFeatureID] = append(adjacent[edge.FromFeatureID], edge)
				addInvalidation(result, FeatureInvalidation{FeatureID: edge.FromFeatureID, Path: path, Reason: "shared input changed: " + edge.Reason})
				break
			}
		}
	}
	queue := invalidatedIDs(result.Invalidations)
	seen := make(map[string]bool, len(queue))
	for _, id := range queue {
		seen[id] = true
	}
	for len(queue) != 0 {
		from := queue[0]
		queue = queue[1:]
		for _, edge := range adjacent[from] {
			if seen[edge.ToFeatureID] {
				continue
			}
			seen[edge.ToFeatureID] = true
			addInvalidation(result, FeatureInvalidation{FeatureID: edge.ToFeatureID, Reason: "dependent of " + edge.FromFeatureID + ": " + edge.Reason})
			queue = append(queue, edge.ToFeatureID)
		}
	}
}

func mappingGaps(files []domain.File, features []domain.Feature) []MappingGap {
	paths := make(map[string]struct{}, len(files))
	var gaps []MappingGap
	for _, file := range files {
		paths[file.Path] = struct{}{}
		if file.Eligible && !file.Excluded && len(file.OwnedFeatureIDs) == 0 {
			gaps = append(gaps, MappingGap{Path: file.Path, Reason: "eligible file has no canonical Clawpatch owner"})
		}
		if len(file.OwnedFeatureIDs) > 1 {
			gaps = append(gaps, MappingGap{Path: file.Path, Reason: "canonical ownership is ambiguous", Related: file.OwnedFeatureIDs})
		}
	}
	for _, feature := range features {
		for _, ref := range append(append([]domain.FeatureFileRef{}, feature.OwnedFiles...), feature.ContextFiles...) {
			if _, exists := paths[ref.Path]; !exists {
				gaps = append(gaps, MappingGap{FeatureID: feature.FeatureID, Path: ref.Path, Reason: "canonical feature path is absent from current inventory"})
			}
		}
	}
	return gaps
}


func invalidatedIDs(invalidations []FeatureInvalidation) []string {
	ids := make([]string, 0, len(invalidations))
	for _, invalidation := range invalidations {
		ids = append(ids, invalidation.FeatureID)
	}
	return sortedUnique(ids)
}

func sortDiff(result *DiffResult) {
	sort.Slice(result.Changes, func(i, j int) bool { return changePath(result.Changes[i]) < changePath(result.Changes[j]) })
	sort.Slice(result.Added, func(i, j int) bool { return result.Added[i].Path < result.Added[j].Path })
	sort.Slice(result.Modified, func(i, j int) bool { return result.Modified[i].Path < result.Modified[j].Path })
	sort.Slice(result.Removed, func(i, j int) bool { return result.Removed[i].Path < result.Removed[j].Path })
	sort.Slice(result.Renamed, func(i, j int) bool { return result.Renamed[i].Path < result.Renamed[j].Path })
	sort.Slice(result.Invalidations, func(i, j int) bool {
		if result.Invalidations[i].FeatureID != result.Invalidations[j].FeatureID {
			return result.Invalidations[i].FeatureID < result.Invalidations[j].FeatureID
		}
		if result.Invalidations[i].Path != result.Invalidations[j].Path {
			return result.Invalidations[i].Path < result.Invalidations[j].Path
		}
		return result.Invalidations[i].Reason < result.Invalidations[j].Reason
	})
	sortMappingGaps(result.MappingGaps)
}

func changePath(change FileChange) string {
	if change.Current != nil {
		return change.Current.Path
	}
	if change.Previous != nil {
		return change.Previous.Path
	}
	return ""
}
func filePointer(file domain.File) *domain.File { copy := file; return &copy }
func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
func filepathBase(path string) string {
	pieces := strings.Split(path, "/")
	return pieces[len(pieces)-1]
}

func fingerprintTuple(values ...string) string {
	var builder strings.Builder
	for _, value := range values {
		_, _ = fmt.Fprintf(&builder, "%d:", len(value))
		builder.WriteString(value)
	}
	return builder.String()
}

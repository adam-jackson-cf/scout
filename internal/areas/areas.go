// Package areas keeps Scout's area labels over Clawpatch output (design §4).
//
// An area is a stable label for a set of files. Agent mapping yields features
// whose Clawpatch IDs churn between runs (CP07 §5: agent-ID Jaccard 0.06–0.16),
// so Reconcile matches each new agent feature to a prior area by file-set
// Jaccard and keeps the area's ScoutAreaID. Areas are labels only: they never
// enter issue identity and never re-key a signal series.
package areas

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"scout/internal/domain"
	"scout/internal/measures"
)

const (
	// DefaultMatchJaccard is the file-set Jaccard at or above which an agent
	// feature matches a prior area. [ASSUMPTION, X2]: not validated.
	DefaultMatchJaccard = 0.5

	// OriginAgent marks an area labelled from an agent-mapped feature.
	OriginAgent = "agent"

	// AgentSource is the Clawpatch feature source of agent-mapped features
	// (third_party/clawpatch/src/agent-mapper.ts). Every other source is a
	// heuristic mapper (python-source-group, package-json-bin, ...).
	AgentSource = "agent-mapper"
)

// IsAgent reports whether a Clawpatch feature came from the agent mapper.
func IsAgent(feature domain.Feature) bool { return feature.Source == AgentSource }

// Area is one stable label. FileSet and ClawpatchIDs are sorted and unique.
// RetiredSnapshot is set once the area is fully heuristic-owned and no agent
// feature matches it; a retired area is kept for audit and never matched again.
type Area struct {
	ScoutAreaID       string   `json:"scoutAreaId"`
	Title             string   `json:"title"`
	FileSet           []string `json:"fileSet"`
	Origin            string   `json:"origin"`
	FirstSeenSnapshot string   `json:"firstSeenSnapshot"`
	ClawpatchIDs      []string `json:"clawpatchIds"`
	RetiredSnapshot   string   `json:"retiredSnapshot,omitempty"`
}

// Active reports whether the area still labels files.
func (a Area) Active() bool { return a.RetiredSnapshot == "" }

// Config parameterizes Reconcile. MatchJaccard is policy
// `mapping.areas.match_jaccard`; zero selects DefaultMatchJaccard.
type Config struct {
	MatchJaccard float64
	// Snapshot identifies the refresh being reconciled; it becomes
	// FirstSeenSnapshot of new areas and RetiredSnapshot of retired ones.
	Snapshot string
}

func (c Config) threshold() float64 {
	if c.MatchJaccard <= 0 {
		return DefaultMatchJaccard
	}
	return c.MatchJaccard
}

// Match records one feature matched to an existing area.
type Match struct {
	ScoutAreaID string  `json:"scoutAreaId"`
	FeatureID   string  `json:"featureId"`
	Jaccard     float64 `json:"jaccard"`
}

// Report describes one reconciliation. All lists are sorted.
type Report struct {
	Matched []Match  `json:"matched"`
	Created []string `json:"created"`
	Retired []string `json:"retired"`
	// Kept lists active areas no feature matched that still own files the
	// heuristic map does not; they stay as they were.
	Kept []string `json:"kept"`
}

// Reconcile folds agent features into prev. Only agent-mapper features that
// own at least one file participate. Pairs with file-set
// Jaccard ≥ the threshold match one-to-one, best first (ties by area ID, then
// feature ID). A matched area takes the feature's file set, title and Clawpatch
// ID; an unmatched feature becomes a new area; an unmatched active area whose
// every file is heuristic-owned retires. prev is not modified.
func Reconcile(prev []Area, agentFeatures []domain.Feature, heuristicOwned func(path string) bool, cfg Config) ([]Area, Report) {
	threshold := cfg.threshold()
	next := make([]Area, len(prev))
	for i, area := range prev {
		next[i] = cloneArea(area)
	}

	type candidate struct {
		area, feature int
		jaccard       float64
	}
	type featureSet struct {
		feature domain.Feature
		files   []string
	}
	var features []featureSet
	for _, feature := range agentFeatures {
		if !IsAgent(feature) {
			continue
		}
		if files := FeatureFiles(feature); len(files) != 0 {
			features = append(features, featureSet{feature, files})
		}
	}
	sort.Slice(features, func(i, j int) bool { return features[i].feature.FeatureID < features[j].feature.FeatureID })

	var candidates []candidate
	for ai, area := range next {
		if !area.Active() {
			continue
		}
		for fi, feature := range features {
			if j := Jaccard(area.FileSet, feature.files); j >= threshold {
				candidates = append(candidates, candidate{ai, fi, j})
			}
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if a.jaccard != b.jaccard {
			return a.jaccard > b.jaccard
		}
		if next[a.area].ScoutAreaID != next[b.area].ScoutAreaID {
			return next[a.area].ScoutAreaID < next[b.area].ScoutAreaID
		}
		return features[a.feature].feature.FeatureID < features[b.feature].feature.FeatureID
	})

	var report Report
	areaTaken := make([]bool, len(next))
	featureTaken := make([]bool, len(features))
	for _, c := range candidates {
		if areaTaken[c.area] || featureTaken[c.feature] {
			continue
		}
		areaTaken[c.area], featureTaken[c.feature] = true, true
		feature := features[c.feature]
		area := &next[c.area]
		area.FileSet = feature.files
		area.ClawpatchIDs = []string{feature.feature.FeatureID}
		area.Origin = OriginAgent
		if feature.feature.Title != "" {
			area.Title = feature.feature.Title
		}
		report.Matched = append(report.Matched, Match{ScoutAreaID: area.ScoutAreaID, FeatureID: feature.feature.FeatureID, Jaccard: c.jaccard})
	}

	for ai := range next {
		area := &next[ai]
		if !area.Active() || areaTaken[ai] {
			continue
		}
		if fullyOwned(area.FileSet, heuristicOwned) {
			area.RetiredSnapshot = cfg.Snapshot
			report.Retired = append(report.Retired, area.ScoutAreaID)
		} else {
			report.Kept = append(report.Kept, area.ScoutAreaID)
		}
	}

	used := make(map[string]bool, len(next))
	for _, area := range next {
		used[area.ScoutAreaID] = true
	}
	for fi, feature := range features {
		if featureTaken[fi] {
			continue
		}
		area := Area{
			ScoutAreaID:       mintID(cfg.Snapshot, feature.feature.FeatureID, feature.files, used),
			Title:             feature.feature.Title,
			FileSet:           feature.files,
			Origin:            OriginAgent,
			FirstSeenSnapshot: cfg.Snapshot,
			ClawpatchIDs:      []string{feature.feature.FeatureID},
		}
		used[area.ScoutAreaID] = true
		next = append(next, area)
		report.Created = append(report.Created, area.ScoutAreaID)
	}

	sort.Slice(next, func(i, j int) bool { return next[i].ScoutAreaID < next[j].ScoutAreaID })
	sort.Slice(report.Matched, func(i, j int) bool { return report.Matched[i].ScoutAreaID < report.Matched[j].ScoutAreaID })
	sort.Strings(report.Created)
	sort.Strings(report.Retired)
	sort.Strings(report.Kept)
	return next, report
}

// FeatureFiles is the sorted, de-duplicated owned-file set of a feature.
func FeatureFiles(feature domain.Feature) []string {
	set := make(map[string]struct{}, len(feature.OwnedFiles))
	for _, ref := range feature.OwnedFiles {
		set[ref.Path] = struct{}{}
	}
	files := make([]string, 0, len(set))
	for path := range set {
		files = append(files, path)
	}
	sort.Strings(files)
	return files
}

// Jaccard is |a∩b| / |a∪b| over two sorted unique slices; two empty sets
// have Jaccard 0 (an empty file set matches nothing).
func Jaccard(a, b []string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 0
	}
	inter, i, j := 0, 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			inter++
			i++
			j++
		case a[i] < b[j]:
			i++
		default:
			j++
		}
	}
	return float64(inter) / float64(len(a)+len(b)-inter)
}

func fullyOwned(files []string, owned func(string) bool) bool {
	if owned == nil {
		return false
	}
	for _, path := range files {
		if !owned(path) {
			return false
		}
	}
	return true
}

func cloneArea(a Area) Area {
	a.FileSet = append([]string(nil), a.FileSet...)
	a.ClawpatchIDs = append([]string(nil), a.ClawpatchIDs...)
	return a
}

// mintID derives an opaque area ID from the refresh snapshot, the feature ID
// and the file set, extending the digest on the (unlikely) collision.
func mintID(snapshot, featureID string, files []string, used map[string]bool) string {
	hash := sha256.New()
	hash.Write([]byte(snapshot))
	hash.Write([]byte{0})
	hash.Write([]byte(featureID))
	for _, path := range files {
		hash.Write([]byte{0})
		hash.Write([]byte(path))
	}
	digest := hex.EncodeToString(hash.Sum(nil))
	for length := 16; ; length += 8 {
		if length > len(digest) {
			length = len(digest)
		}
		if id := "area:" + digest[:length]; !used[id] || length == len(digest) {
			return id
		}
	}
}

// HeuristicOwner returns a predicate over the owned files of the heuristic
// features in active.
func HeuristicOwner(active []domain.Feature) func(path string) bool {
	owned := map[string]struct{}{}
	for _, feature := range active {
		if IsAgent(feature) {
			continue
		}
		for _, ref := range feature.OwnedFiles {
			owned[ref.Path] = struct{}{}
		}
	}
	return func(path string) bool {
		_, ok := owned[path]
		return ok
	}
}

// DirScope is a directory measurement scope for files no heuristic feature owns.
type DirScope struct {
	Directory string   `json:"directory"`
	Files     []string `json:"files"`
}

// UnownedScopes groups the files no heuristic feature owns by their directory
// scope (measures.DirectoryScope: two components below the source root, fixed
// for a map version so a series never re-keys). Output is sorted by directory.
func UnownedScopes(paths []string, heuristicOwned func(path string) bool) []DirScope {
	byDir := map[string][]string{}
	for _, path := range paths {
		if heuristicOwned != nil && heuristicOwned(path) {
			continue
		}
		dir := measures.DirectoryScope(path)
		byDir[dir] = append(byDir[dir], path)
	}
	scopes := make([]DirScope, 0, len(byDir))
	for dir, files := range byDir {
		sort.Strings(files)
		scopes = append(scopes, DirScope{Directory: dir, Files: files})
	}
	sort.Slice(scopes, func(i, j int) bool { return scopes[i].Directory < scopes[j].Directory })
	return scopes
}

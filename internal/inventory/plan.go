package inventory

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"scout/internal/domain"
)

// CachedCoverage is the exact successful per-unit cache record. A coverage row
// without its input fingerprint is never reused by the planner.
type CachedCoverage struct {
	Coverage         domain.Coverage `json:"coverage"`
	SourceRunID      string          `json:"sourceRunId"`
	InputFingerprint string          `json:"inputFingerprint"`
	CacheKey         string          `json:"cacheKey"`
}

// PlanInput contains the externally established map and recorded dependency
// edges. It deliberately has no alternative feature-map fields.
type PlanInput struct {
	RunID             string
	Snapshot          domain.Snapshot
	Scope             domain.Scope
	ScopeDigest       string
	Files             []domain.File
	Features          []domain.Feature
	Diff              DiffResult
	Dependencies      []DependencyEdge
	FileCategories    map[domain.Family][]string
	FeatureCategories map[domain.Family][]string
	Baseline          []CachedCoverage
	ForceFull         bool
	MappingRequired   bool
	ClawpatchReview   bool
	PolicyDigest      string
	ToolDigest        string
	ModelDigest       string
	MapperDigest      string
}

// PlanResult separates scheduled and reused work and preserves any unit that
// cannot safely be scheduled because mapping or inputs are incomplete.
type PlanResult struct {
	Plan        domain.Plan       `json:"plan"`
	Reused      []domain.Coverage `json:"reused,omitempty"`
	Omissions   []domain.Omission `json:"omissions,omitempty"`
	Diagnostics []Diagnostic      `json:"diagnostics,omitempty"`
}

// Plan derives stable per-unit jobs. It schedules only changed file inputs,
// affected canonical features, forced-full units, and incomplete prior units;
// it never turns a whole prior run into a single cache entry.
func Plan(input PlanInput) (PlanResult, error) {
	result := PlanResult{}
	if strings.TrimSpace(input.RunID) == "" {
		return result, errors.New("plan requires run ID")
	}
	if err := input.Snapshot.Validate(); err != nil {
		return result, fmt.Errorf("plan snapshot: %w", err)
	}
	if err := input.Scope.Validate(); err != nil {
		return result, fmt.Errorf("plan scope: %w", err)
	}
	if input.ScopeDigest == "" {
		input.ScopeDigest = scopeDigest(input.Scope, nil)
	}
	if err := validateCategories(input.FileCategories, input.Scope.SelectedFamilies); err != nil {
		return result, err
	}
	if err := validateCategories(input.FeatureCategories, input.Scope.SelectedFamilies); err != nil {
		return result, err
	}

	files, err := indexedFiles(input.Files)
	if err != nil {
		return result, fmt.Errorf("plan files: %w", err)
	}
	features, err := indexedFeatures(input.Features)
	if err != nil {
		return result, fmt.Errorf("plan features: %w", err)
	}
	if input.ForceFull {
		input.Baseline = nil
	}
	baseline, err := baselineIndex(input.Baseline)
	if err != nil {
		return result, fmt.Errorf("plan baseline: %w", err)
	}

	jobs := make(map[string]domain.Job)
	reusedCoverage := make(map[string]domain.Coverage)
	if input.MappingRequired {
		fingerprint := digestStrings([]string{"mapping", input.Snapshot.WorktreeDigest, input.Snapshot.IgnoreRulesDigest, input.Snapshot.ExclusionsDigest, input.ScopeDigest, input.MapperDigest})
		addJob(jobs, job(domain.StageMapping, nil, fingerprint, "canonical map inputs changed", domain.JobPending, ""))
	}

	for _, file := range selectedFiles(input, files, baseline) {
		if !file.Eligible || file.Excluded || file.ContentDigest == "" {
			continue
		}
		for family, categories := range input.FileCategories {
			for _, category := range categories {
				key := domain.CoverageKey{FilePath: file.Path, Family: family, Category: category}
				fingerprint := fileFingerprint(file, key, input)
				planned, reused := reuseOrSchedule(key, domain.StageFileAnalysis, fingerprint, "file changed or lacks reusable coverage", baseline)
				addJob(jobs, planned)
				if reused != nil {
					recordReused(reusedCoverage, *reused)
				}
			}
		}
	}

	affected := selectedFeatureIDs(input, features, baseline)
	for _, id := range affected {
		feature, exists := features[id]
		if !exists {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{Stage: domain.StagePlanning, Code: "affected_feature_missing", Message: "affected canonical feature is absent from current map", Path: id})
			continue
		}
		for family, categories := range input.FeatureCategories {
			for _, category := range categories {
				key := domain.CoverageKey{FeatureID: id, Family: family, Category: category}
				fingerprint := featureFingerprint(feature, files, key, input)
				planned, reused := reuseOrSchedule(key, domain.StageFeatureAnalysis, fingerprint, "feature affected or lacks reusable coverage", baseline)
				addJob(jobs, planned)
				if reused != nil {
					recordReused(reusedCoverage, *reused)
				}
			}
		}
		if input.ClawpatchReview {
			key := domain.CoverageKey{FeatureID: id, Family: domain.FamilyClawpatch, Category: "review"}
			fingerprint := featureFingerprint(feature, files, key, input)
			planned, reused := reuseOrSchedule(key, domain.StageClawpatchReview, fingerprint, "feature affected or lacks reusable review", baseline)
			addJob(jobs, planned)
			if reused != nil {
				recordReused(reusedCoverage, *reused)
			}
		}
	}

	for _, cached := range input.Baseline {
		key := cached.Coverage.Key
		if !scopeAllowsKey(key, input.Scope) {
			continue
		}
		fingerprint, found := fingerprintForKey(key, files, features, input)
		if !found {
			if !cached.Coverage.Status.Accounted() {
				result.Omissions = append(result.Omissions, domain.Omission{CoverageKey: key, Reason: "unfinished prior unit has no current eligible input"})
			}
			continue
		}
		stage := stageForKey(key)
		if cached.Coverage.Status.Accounted() && cached.SourceRunID != "" && cached.InputFingerprint == fingerprint {
			coverage := cached.Coverage
			coverage.Status = domain.CoverageReused
			coverage.ReusedFromRunID = cached.SourceRunID
			recordReused(reusedCoverage, coverage)
			addJob(jobs, job(stage, &key, fingerprint, "exact per-unit cache match", domain.JobReused, cached.CacheKey))
			continue
		}
		reason := "prior coverage was incomplete"
		if cached.Coverage.Status.Accounted() {
			reason = "cache inputs changed"
			if cached.SourceRunID == "" {
				reason = "cache baseline has no source run ID"
			}
		}
		addJob(jobs, job(stage, &key, fingerprint, reason, domain.JobPending, ""))
	}

	result.Plan = domain.Plan{
		PlanID:       "plan-" + digestStrings([]string{input.RunID, input.Snapshot.WorktreeDigest, input.ScopeDigest, input.PolicyDigest, input.ToolDigest, input.ModelDigest, input.MapperDigest}),
		RunID:        input.RunID,
		PolicyDigest: input.PolicyDigest,
		ToolDigest:   input.ToolDigest,
		ModelDigest:  input.ModelDigest,
		MapperDigest: input.MapperDigest,
		Snapshot:     input.Snapshot,
		Scope:        input.Scope,
		Jobs:         sortedJobs(jobs),
		CreatedAt:    time.Now().UTC(),
	}
	for _, coverage := range reusedCoverage {
		result.Reused = append(result.Reused, coverage)
	}
	sort.Slice(result.Reused, func(i, j int) bool { return result.Reused[i].Key.String() < result.Reused[j].Key.String() })
	sort.Slice(result.Omissions, func(i, j int) bool {
		return result.Omissions[i].CoverageKey.String() < result.Omissions[j].CoverageKey.String()
	})
	if err := result.Plan.Validate(); err != nil {
		return result, fmt.Errorf("build plan: %w", err)
	}
	return result, nil
}

func validateCategories(categories map[domain.Family][]string, selected []domain.Family) error {
	selectedSet := map[domain.Family]bool{}
	for _, family := range selected {
		selectedSet[family] = true
	}
	for family, values := range categories {
		if !family.Valid() || family == domain.FamilyClawpatch {
			return fmt.Errorf("invalid contextual family %q", family)
		}
		if len(selectedSet) != 0 && !selectedSet[family] {
			return fmt.Errorf("category family %q is not selected", family)
		}
		for _, category := range values {
			if strings.TrimSpace(category) == "" {
				return fmt.Errorf("family %q has blank category", family)
			}
		}
	}
	return nil
}

func baselineIndex(cached []CachedCoverage) (map[string]CachedCoverage, error) {
	index := make(map[string]CachedCoverage, len(cached))
	for _, record := range cached {
		if err := record.Coverage.Validate(); err != nil {
			return nil, err
		}
		if record.InputFingerprint == "" {
			continue
		}
		key := record.Coverage.Key.String()
		if _, exists := index[key]; exists {
			return nil, fmt.Errorf("duplicate baseline coverage %q", key)
		}
		index[key] = record
	}
	return index, nil
}

func selectedFiles(input PlanInput, files map[string]domain.File, baseline map[string]CachedCoverage) []domain.File {
	selected := map[string]domain.File{}
	if input.ForceFull {
		for path, file := range files {
			selected[path] = file
		}
	}
	for _, file := range append(append([]domain.File{}, input.Diff.Added...), input.Diff.Modified...) {
		selected[file.Path] = file
	}
	// A newly registered category has no baseline row even when source is
	// unchanged. Schedule that file rather than silently omitting its coverage.
	for _, file := range files {
		if !file.Eligible || file.Excluded || file.ContentDigest == "" {
			continue
		}
		for family, categories := range input.FileCategories {
			for _, category := range categories {
				key := domain.CoverageKey{FilePath: file.Path, Family: family, Category: category}
				if _, exists := baseline[key.String()]; !exists {
					selected[file.Path] = file
					break
				}
			}
			if _, selectedNow := selected[file.Path]; selectedNow {
				break
			}
		}
	}
	for _, cached := range baseline {
		if cached.Coverage.Status.Accounted() || cached.Coverage.Key.FilePath == "" {
			continue
		}
		if file, exists := files[cached.Coverage.Key.FilePath]; exists {
			selected[file.Path] = file
		}
	}
	values := make([]domain.File, 0, len(selected))
	for _, file := range selected {
		values = append(values, file)
	}
	sort.Slice(values, func(i, j int) bool { return values[i].Path < values[j].Path })
	return values
}

func selectedFeatureIDs(input PlanInput, features map[string]domain.Feature, baseline map[string]CachedCoverage) []string {
	ids := append([]string(nil), input.Diff.AffectedFeatureIDs...)
	if input.ForceFull {
		for id := range features {
			ids = append(ids, id)
		}
	}
	for id := range features {
		missing := false
		for family, categories := range input.FeatureCategories {
			for _, category := range categories {
				key := domain.CoverageKey{FeatureID: id, Family: family, Category: category}
				if _, exists := baseline[key.String()]; !exists {
					missing = true
					break
				}
			}
			if missing {
				break
			}
		}
		if input.ClawpatchReview {
			key := domain.CoverageKey{FeatureID: id, Family: domain.FamilyClawpatch, Category: "review"}
			_, exists := baseline[key.String()]
			missing = missing || !exists
		}
		if missing {
			ids = append(ids, id)
		}
	}
	for _, cached := range baseline {
		if !cached.Coverage.Status.Accounted() && cached.Coverage.Key.FeatureID != "" {
			ids = append(ids, cached.Coverage.Key.FeatureID)
		}
	}
	return sortedUnique(ids)
}

func reuseOrSchedule(key domain.CoverageKey, stage domain.Stage, fingerprint, reason string, baseline map[string]CachedCoverage) (domain.Job, *domain.Coverage) {
	if cached, exists := baseline[key.String()]; exists && cached.Coverage.Status.Accounted() && cached.SourceRunID != "" && cached.InputFingerprint == fingerprint {
		coverage := cached.Coverage
		coverage.Status = domain.CoverageReused
		coverage.ReusedFromRunID = cached.SourceRunID
		return job(stage, &key, fingerprint, "exact per-unit cache match", domain.JobReused, cached.CacheKey), &coverage
	}
	return job(stage, &key, fingerprint, reason, domain.JobPending, ""), nil
}

func fingerprintForKey(key domain.CoverageKey, files map[string]domain.File, features map[string]domain.Feature, input PlanInput) (string, bool) {
	if key.FilePath != "" {
		file, exists := files[key.FilePath]
		if !exists || !file.Eligible || file.Excluded || file.ContentDigest == "" {
			return "", false
		}
		return fileFingerprint(file, key, input), true
	}
	feature, exists := features[key.FeatureID]
	if !exists {
		return "", false
	}
	return featureFingerprint(feature, files, key, input), true
}

func fileFingerprint(file domain.File, key domain.CoverageKey, input PlanInput) string {
	return digestStrings([]string{fingerprintTuple("file", file.Path, file.ContentDigest, file.Language, string(key.Family), key.Category, input.ScopeDigest, input.Snapshot.IgnoreRulesDigest, input.Snapshot.ExclusionsDigest, input.PolicyDigest, input.ToolDigest)})
}

func featureFingerprint(feature domain.Feature, files map[string]domain.File, key domain.CoverageKey, input PlanInput) string {
	parts := []string{fingerprintTuple("feature", FeatureFingerprint(feature), string(key.Family), key.Category, input.ScopeDigest, input.Snapshot.IgnoreRulesDigest, input.Snapshot.ExclusionsDigest, input.PolicyDigest, input.ToolDigest, input.ModelDigest, input.MapperDigest)}
	for _, ref := range append(append([]domain.FeatureFileRef{}, feature.OwnedFiles...), feature.ContextFiles...) {
		if file, exists := files[ref.Path]; exists {
			parts = append(parts, fingerprintTuple("input", ref.Path, file.ContentDigest))
		} else {
			parts = append(parts, fingerprintTuple("input", ref.Path, "missing"))
		}
	}
	for _, ref := range feature.Tests {
		if file, exists := files[ref.Path]; exists {
			parts = append(parts, fingerprintTuple("test", ref.Path, file.ContentDigest))
		} else {
			parts = append(parts, fingerprintTuple("test", ref.Path, "missing"))
		}
	}
	parts = append(parts, dependencyFootprint(feature.FeatureID, files, input.Dependencies)...)
	return digestStrings(parts)
}

func dependencyFootprint(featureID string, files map[string]domain.File, edges []DependencyEdge) []string {
	parts := make([]string, 0)
	for _, edge := range edges {
		if edge.FromFeatureID == "" || edge.ToFeatureID == "" || strings.TrimSpace(edge.Reason) == "" || len(edge.InputPaths) == 0 {
			continue
		}
		if edge.FromFeatureID != featureID && edge.ToFeatureID != featureID {
			continue
		}
		paths := sortedUnique(edge.InputPaths)
		for _, path := range paths {
			digest := "missing"
			if file, exists := files[path]; exists {
				digest = file.ContentDigest
			}
			parts = append(parts, fingerprintTuple("dependency", edge.FromFeatureID, edge.ToFeatureID, edge.Reason, path, digest))
		}
	}
	sort.Strings(parts)
	return parts
}

func recordReused(reused map[string]domain.Coverage, coverage domain.Coverage) {
	reused[coverage.Key.String()] = coverage
}

func stageForKey(key domain.CoverageKey) domain.Stage {
	if key.Family == domain.FamilyClawpatch {
		return domain.StageClawpatchReview
	}
	if key.FeatureID != "" {
		return domain.StageFeatureAnalysis
	}
	return domain.StageFileAnalysis
}

func scopeAllowsKey(key domain.CoverageKey, scope domain.Scope) bool {
	if key.Family != domain.FamilyClawpatch && len(scope.SelectedFamilies) != 0 {
		found := false
		for _, family := range scope.SelectedFamilies {
			if family == key.Family {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if scope.Kind == domain.ScopeRepository || scope.Kind == domain.ScopeStaged || scope.Kind == domain.ScopeCommitRange {
		return true
	}
	path := key.FilePath
	if path == "" {
		return true
	}
	for _, scopePath := range scope.Paths {
		if path == scopePath || strings.HasPrefix(path, scopePath+"/") {
			return true
		}
	}
	return false
}

func job(stage domain.Stage, key *domain.CoverageKey, fingerprint, reason string, state domain.JobState, cacheKey string) domain.Job {
	identity := string(stage) + "\x00" + fingerprint
	if key != nil {
		identity += "\x00" + key.String()
	}
	return domain.Job{JobID: "job-" + digestStrings([]string{identity}), Stage: stage, CoverageKey: key, InputFingerprint: fingerprint, State: state, InvalidationReason: reason, CacheKey: cacheKey}
}

func addJob(jobs map[string]domain.Job, planned domain.Job) {
	if existing, exists := jobs[planned.JobID]; !exists || (existing.State == domain.JobReused && planned.State == domain.JobPending) {
		jobs[planned.JobID] = planned
	}
}

func sortedJobs(jobs map[string]domain.Job) []domain.Job {
	values := make([]domain.Job, 0, len(jobs))
	for _, planned := range jobs {
		values = append(values, planned)
	}
	sort.Slice(values, func(i, j int) bool {
		left, right := values[i], values[j]
		if left.Stage != right.Stage {
			return left.Stage < right.Stage
		}
		leftKey, rightKey := "", ""
		if left.CoverageKey != nil {
			leftKey = left.CoverageKey.String()
		}
		if right.CoverageKey != nil {
			rightKey = right.CoverageKey.String()
		}
		if leftKey != rightKey {
			return leftKey < rightKey
		}
		return left.JobID < right.JobID
	})
	return values
}

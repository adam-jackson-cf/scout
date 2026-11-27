package inventory

import (
	"errors"
	"fmt"
	"slices"
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
	if err := validatePlanInput(input); err != nil {
		return result, err
	}
	planner, err := newPlanner(input)
	if err != nil {
		return result, err
	}
	planner.planMapping()
	planner.planFiles()
	result.Diagnostics = planner.planFeatures()
	result.Omissions = planner.planBaseline()
	result.Plan = planner.plan()
	result.Reused = planner.sortedReused()
	sort.Slice(result.Omissions, func(i, j int) bool {
		return result.Omissions[i].CoverageKey.String() < result.Omissions[j].CoverageKey.String()
	})
	if err := result.Plan.Validate(); err != nil {
		return result, fmt.Errorf("build plan: %w", err)
	}
	return result, nil
}

func validatePlanInput(input PlanInput) error {
	if strings.TrimSpace(input.RunID) == "" {
		return errors.New("plan requires run ID")
	}
	if err := input.Snapshot.Validate(); err != nil {
		return fmt.Errorf("plan snapshot: %w", err)
	}
	if err := input.Scope.Validate(); err != nil {
		return fmt.Errorf("plan scope: %w", err)
	}
	if input.ScopeDigest == "" {
		return errors.New("plan requires scope digest")
	}
	if err := validateDependencies(input.Dependencies); err != nil {
		return fmt.Errorf("plan dependencies: %w", err)
	}
	if err := validateCategories(input.FileCategories, input.Scope.SelectedFamilies); err != nil {
		return err
	}
	return validateCategories(input.FeatureCategories, input.Scope.SelectedFamilies)
}

type planner struct {
	input    PlanInput
	files    map[string]domain.File
	features map[string]domain.Feature
	baseline map[string]CachedCoverage
	jobs     map[string]domain.Job
	reused   map[string]domain.Coverage
}

func newPlanner(input PlanInput) (*planner, error) {
	files, err := indexedFiles(input.Files)
	if err != nil {
		return nil, fmt.Errorf("plan files: %w", err)
	}
	features, err := indexedFeatures(input.Features)
	if err != nil {
		return nil, fmt.Errorf("plan features: %w", err)
	}
	if input.ForceFull {
		input.Baseline = nil
	}
	baseline, err := baselineIndex(input.Baseline)
	if err != nil {
		return nil, fmt.Errorf("plan baseline: %w", err)
	}
	return &planner{
		input:    input,
		files:    files,
		features: features,
		baseline: baseline,
		jobs:     make(map[string]domain.Job),
		reused:   make(map[string]domain.Coverage),
	}, nil
}

func (p *planner) schedule(key domain.CoverageKey, stage domain.Stage, fingerprint, reason string) {
	planned, reused := reuseOrSchedule(key, stage, fingerprint, reason, p.baseline)
	addJob(p.jobs, planned)
	if reused != nil {
		recordReused(p.reused, *reused)
	}
}

func (p *planner) planMapping() {
	if !p.input.MappingRequired {
		return
	}
	input := p.input
	fingerprint := orderedDigest("mapping", input.Snapshot.WorktreeDigest, input.Snapshot.IgnoreRulesDigest, input.Snapshot.ExclusionsDigest, input.ScopeDigest, input.MapperDigest)
	addJob(p.jobs, pendingJob(domain.StageMapping, nil, fingerprint, "canonical map inputs changed"))
}

func (p *planner) planFiles() {
	for _, file := range selectedFiles(p.input, p.files, p.baseline) {
		if !analyzableFile(file) {
			continue
		}
		for family, categories := range p.input.FileCategories {
			for _, category := range categories {
				key := domain.CoverageKey{FilePath: file.Path, Family: family, Category: category}
				fingerprint := fileFingerprint(file, key, p.input)
				p.schedule(key, domain.StageFileAnalysis, fingerprint, "file changed or lacks reusable coverage")
			}
		}
	}
}

func (p *planner) planFeatures() []Diagnostic {
	var diagnostics []Diagnostic
	for _, id := range selectedFeatureIDs(p.input, p.features, p.baseline) {
		feature, exists := p.features[id]
		if !exists {
			diagnostics = append(diagnostics, Diagnostic{Stage: domain.StagePlanning, Code: "affected_feature_missing", Message: "affected canonical feature is absent from current map", Path: id})
			continue
		}
		p.planFeature(id, feature)
	}
	return diagnostics
}

func (p *planner) planFeature(id string, feature domain.Feature) {
	for family, categories := range p.input.FeatureCategories {
		for _, category := range categories {
			key := domain.CoverageKey{FeatureID: id, Family: family, Category: category}
			fingerprint := featureFingerprint(feature, p.files, key, p.input)
			p.schedule(key, domain.StageFeatureAnalysis, fingerprint, "feature affected or lacks reusable coverage")
		}
	}
	if p.input.ClawpatchReview {
		key := domain.CoverageKey{FeatureID: id, Family: domain.FamilyClawpatch, Category: "review"}
		fingerprint := featureFingerprint(feature, p.files, key, p.input)
		p.schedule(key, domain.StageClawpatchReview, fingerprint, "feature affected or lacks reusable review")
	}
}

func (p *planner) planBaseline() []domain.Omission {
	var omissions []domain.Omission
	for _, cached := range p.input.Baseline {
		key := cached.Coverage.Key
		if !scopeAllowsKey(key, p.input.Scope) {
			continue
		}
		fingerprint, found := fingerprintForKey(key, p.files, p.features, p.input)
		if !found {
			if !cached.Coverage.Status.Accounted() {
				omissions = append(omissions, domain.Omission{CoverageKey: key, Reason: "unfinished prior unit has no current eligible input"})
			}
			continue
		}
		p.schedule(key, stageForKey(key), fingerprint, baselineReason(cached))
	}
	return omissions
}

func baselineReason(cached CachedCoverage) string {
	reason := "prior coverage was incomplete"
	if cached.Coverage.Status.Accounted() {
		reason = "cache inputs changed"
		if cached.SourceRunID == "" {
			reason = "cache baseline has no source run ID"
		}
	}
	return reason
}

func (p *planner) plan() domain.Plan {
	input := p.input
	return domain.Plan{
		PlanID:       "plan-" + orderedDigest(input.RunID, input.Snapshot.WorktreeDigest, input.ScopeDigest, input.PolicyDigest, input.ToolDigest, input.ModelDigest, input.MapperDigest),
		RunID:        input.RunID,
		PolicyDigest: input.PolicyDigest,
		ToolDigest:   input.ToolDigest,
		ModelDigest:  input.ModelDigest,
		MapperDigest: input.MapperDigest,
		Snapshot:     input.Snapshot,
		Scope:        input.Scope,
		Jobs:         sortedJobs(p.jobs),
		CreatedAt:    time.Now().UTC(),
	}
}

func (p *planner) sortedReused() []domain.Coverage {
	var reused []domain.Coverage
	for _, coverage := range p.reused {
		reused = append(reused, coverage)
	}
	sort.Slice(reused, func(i, j int) bool { return reused[i].Key.String() < reused[j].Key.String() })
	return reused
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
		if !analyzableFile(file) {
			continue
		}
		selectUncoveredFile(selected, file, input.FileCategories, baseline)
	}
	selectIncompleteFiles(selected, files, baseline)
	return sortedFiles(selected)
}

func analyzableFile(file domain.File) bool {
	return file.Eligible && !file.Excluded && file.ContentDigest != ""
}

func selectUncoveredFile(selected map[string]domain.File, file domain.File, fileCategories map[domain.Family][]string, baseline map[string]CachedCoverage) {
	for family, categories := range fileCategories {
		if fileLacksFamilyCoverage(file.Path, family, categories, baseline) {
			selected[file.Path] = file
			break
		}
		if _, selectedNow := selected[file.Path]; selectedNow {
			break
		}
	}
}

func fileLacksFamilyCoverage(path string, family domain.Family, categories []string, baseline map[string]CachedCoverage) bool {
	for _, category := range categories {
		key := domain.CoverageKey{FilePath: path, Family: family, Category: category}
		if _, exists := baseline[key.String()]; !exists {
			return true
		}
	}
	return false
}

func selectIncompleteFiles(selected, files map[string]domain.File, baseline map[string]CachedCoverage) {
	for _, cached := range baseline {
		if cached.Coverage.Status.Accounted() || cached.Coverage.Key.FilePath == "" {
			continue
		}
		if file, exists := files[cached.Coverage.Key.FilePath]; exists {
			selected[file.Path] = file
		}
	}
}

func sortedFiles(selected map[string]domain.File) []domain.File {
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
		if featureLacksCoverage(id, input, baseline) {
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

func featureLacksCoverage(id string, input PlanInput, baseline map[string]CachedCoverage) bool {
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
	return missing
}

func reuseOrSchedule(key domain.CoverageKey, stage domain.Stage, fingerprint, reason string, baseline map[string]CachedCoverage) (domain.Job, *domain.Coverage) {
	if cached, exists := baseline[key.String()]; exists && cached.Coverage.Status.Accounted() && cached.SourceRunID != "" && cached.InputFingerprint == fingerprint {
		coverage := cached.Coverage
		coverage.Status = domain.CoverageReused
		coverage.ReusedFromRunID = cached.SourceRunID
		reusedJob := pendingJob(stage, &key, fingerprint, "exact per-unit cache match")
		reusedJob.State = domain.JobReused
		reusedJob.CacheKey = cached.CacheKey
		return reusedJob, &coverage
	}
	return pendingJob(stage, &key, fingerprint, reason), nil
}

func fingerprintForKey(key domain.CoverageKey, files map[string]domain.File, features map[string]domain.Feature, input PlanInput) (string, bool) {
	if key.FilePath != "" {
		file, exists := files[key.FilePath]
		if !exists || !analyzableFile(file) {
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
	if !scopeAllowsFamily(key.Family, scope) {
		return false
	}
	if scope.Kind == domain.ScopeRepository || scope.Kind == domain.ScopeStaged || scope.Kind == domain.ScopeCommitRange {
		return true
	}
	path := key.FilePath
	if path == "" {
		return true
	}
	return pathInScope(path, scope.Paths)
}

func scopeAllowsFamily(family domain.Family, scope domain.Scope) bool {
	if family == domain.FamilyClawpatch || len(scope.SelectedFamilies) == 0 {
		return true
	}
	return slices.Contains(scope.SelectedFamilies, family)
}

func pathInScope(path string, scopePaths []string) bool {
	for _, scopePath := range scopePaths {
		if path == scopePath || strings.HasPrefix(path, scopePath+"/") {
			return true
		}
	}
	return false
}

func pendingJob(stage domain.Stage, key *domain.CoverageKey, fingerprint, reason string) domain.Job {
	identity := string(stage) + "\x00" + fingerprint
	if key != nil {
		identity += "\x00" + key.String()
	}
	return domain.Job{JobID: "job-" + digestStrings([]string{identity}), Stage: stage, CoverageKey: key, InputFingerprint: fingerprint, State: domain.JobPending, InvalidationReason: reason}
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

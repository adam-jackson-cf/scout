package engine

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"scout/internal/analyzers"
	"scout/internal/clawpatch"
	"scout/internal/contextual"
	"scout/internal/domain"
	"scout/internal/inventory"
	"scout/internal/state"
)

// AssessReadiness checks only prerequisites needed by the selected pending work.
// It does not initialize state, map features, scan with analyzers, or call a model.
func AssessReadiness(ctx context.Context, config Config, options Options) (Readiness, error) {
	result := Readiness{Ready: true, Capabilities: []Capability{}}
	add := func(name string, required bool, err error) {
		entry := Capability{Name: name, Required: required, Ready: err == nil}
		if err != nil {
			entry.Reason = err.Error()
			if required {
				result.Ready = false
			}
		}
		result.Capabilities = append(result.Capabilities, entry)
	}
	if options.Scope.Kind == "" {
		options.Scope.Kind = domain.ScopeRepository
	}
	if len(options.Scope.SelectedFamilies) == 0 {
		options.Scope.SelectedFamilies = []domain.Family{domain.FamilyArchitecture, domain.FamilySecurity, domain.FamilyQuality, domain.FamilyPerformance}
	}
	if err := options.Scope.Validate(); err != nil {
		return result, err
	}
	root, err := filepath.Abs(config.Root)
	if err != nil {
		return result, err
	}
	info, err := os.Stat(root)
	if err == nil && !info.IsDir() {
		err = errors.New("assessment root is not a directory")
	}
	add("repository", true, err)
	if err != nil {
		return result, nil
	}

	initial, err := inventory.Scan(ctx, root, inventory.ScanOptions{Scope: options.Scope, Exclusions: options.Exclusions})
	if err == nil && len(initial.Omissions) != 0 {
		err = errors.New("repository inventory has unreadable inputs")
	}
	add("inventory", true, err)
	if err != nil {
		return result, nil
	}
	registry := config.Registry
	if registry == nil {
		registry, err = analyzers.BuiltinRegistry()
	}
	add("scanner_catalog", true, err)
	if err != nil {
		return result, nil
	}
	specs := registry.Specs()
	toolBytes, err := json.Marshal(specs)
	if err != nil {
		return result, err
	}

	previousFiles := []domain.File(nil)
	previousPlan := domain.Plan{}
	latest := domain.Run{}
	first := false
	store, err := state.OpenExistingReadOnly(root)
	switch {
	case errors.Is(err, state.ErrNotFound):
		first = true
	case err != nil:
		add("baseline_state", true, errors.New("existing Scout state is unavailable"))
		return result, nil
	default:
		defer store.Close()
		latest, err = store.LatestFinalizedRun(ctx)
		if errors.Is(err, state.ErrNotFound) {
			first = true
		} else if err == nil {
			previousFiles, err = store.LoadInventory(ctx, latest.RunID)
			if err == nil {
				previousPlan, err = store.LoadPlan(ctx, latest.PlanID)
			}
		}
		if err != nil && !errors.Is(err, state.ErrNotFound) {
			add("baseline_state", true, errors.New("existing Scout baseline is unavailable"))
			return result, nil
		}
		if errors.Is(err, state.ErrNotFound) && !first {
			add("baseline_state", true, errors.New("latest Scout run lacks a finalized inventory"))
			return result, nil
		}
	}
	add("baseline_state", true, nil)
	features, err := clawpatch.LoadFeatures(filepath.Join(root, ".scout", "clawpatch"))
	if err != nil {
		add("canonical_map_records", true, errors.New("canonical feature records are unavailable"))
		return result, nil
	}
	mapperDigest := digest(clawpatch.SafetyFingerprint())
	modelDigest := digest(config.Provider.Endpoint + "\x00" + config.Provider.Model + "\x00" + config.Provider.PromptVersion)
	var resumePlan *domain.Plan
	if !options.ForceFull {
		resumePlan, err = matchingResumablePlan(ctx, store, initial.Snapshot, options.Scope, digest(string(toolBytes)), modelDigest, mapperDigest)
		if err != nil {
			return result, err
		}
	}
	mappingRequired := options.ForceFull || first || len(features) == 0 || previousPlan.MapperDigest != mapperDigest || latest.Snapshot.WorktreeDigest != initial.Snapshot.WorktreeDigest || latest.Snapshot.IgnoreRulesDigest != initial.Snapshot.IgnoreRulesDigest || latest.Snapshot.ExclusionsDigest != initial.Snapshot.ExclusionsDigest
	current, err := inventory.Scan(ctx, root, inventory.ScanOptions{Scope: options.Scope, Exclusions: options.Exclusions, Features: features})
	if err != nil {
		add("mapped_inventory", true, errors.New("feature-linked inventory is unavailable"))
		return result, nil
	}
	diff, err := inventory.Diff(inventory.DiffInput{PreviousFiles: previousFiles, CurrentFiles: current.Files, PreviousFeatures: features, CurrentFeatures: features, Dependencies: config.Dependencies})
	if err != nil {
		add("change_plan", true, err)
		return result, nil
	}
	selected := make(map[domain.Family]bool, len(options.Scope.SelectedFamilies))
	featureCategories := make(map[domain.Family][]string, len(options.Scope.SelectedFamilies))
	fileCategories := make(map[domain.Family][]string, len(options.Scope.SelectedFamilies))
	for _, family := range options.Scope.SelectedFamilies {
		selected[family] = true
		featureCategories[family] = contextualCategories[family]
	}
	byCategory := make(map[string]analyzers.Spec, len(specs))
	for _, spec := range specs {
		if !selected[spec.Family] {
			continue
		}
		fileCategories[spec.Family] = append(fileCategories[spec.Family], spec.Category)
		byCategory[categoryKey(spec.Family, spec.Category)] = spec
	}
	baseline := make([]inventory.CachedCoverage, 0)
	if !first {
		fingerprints := make(map[string]string)
		for _, job := range previousPlan.Jobs {
			if job.CoverageKey != nil {
				fingerprints[job.CoverageKey.String()] = job.InputFingerprint
			}
		}
		for _, row := range latest.Coverage {
			sourceRun := latest.RunID
			if row.Status == domain.CoverageReused {
				sourceRun = row.ReusedFromRunID
			} else if !row.Status.Accounted() {
				sourceRun = ""
			}
			baseline = append(baseline, inventory.CachedCoverage{Coverage: row, SourceRunID: sourceRun, InputFingerprint: fingerprints[row.Key.String()]})
		}
	}
	runID := "readiness"
	if resumePlan != nil {
		runID = resumePlan.RunID
	}
	plan, err := inventory.Plan(inventory.PlanInput{
		RunID: runID, Snapshot: current.Snapshot, Scope: current.Scope, ScopeDigest: current.ScopeDigest,
		Files: current.Files, Features: features, Diff: diff, Dependencies: config.Dependencies,
		FileCategories: fileCategories, FeatureCategories: featureCategories, Baseline: baseline,
		ForceFull: options.ForceFull || first, MappingRequired: mappingRequired, ClawpatchReview: true,
		PolicyDigest: digest("scout-policy-v2"), ToolDigest: digest(string(toolBytes)),
		ModelDigest:  modelDigest,
		MapperDigest: mapperDigest,
	})
	if err != nil {
		add("change_plan", true, err)
		return result, nil
	}
	committed := make(map[string]bool)
	if resumePlan != nil {
		jobs := append([]domain.Job(nil), plan.Plan.Jobs...)
		jobs = append(jobs, domain.Job{
			JobID: "job-" + digest(runID+"\x00final-inventory"),
			Stage: domain.StageInventory, InputFingerprint: current.Snapshot.WorktreeDigest,
			State: domain.JobPending, InvalidationReason: "verify source snapshot after analysis",
		})
		if plan.Plan.PlanID != resumePlan.PlanID || !reflect.DeepEqual(jobs, resumePlan.Jobs) {
			add("change_plan", true, errors.New("interrupted plan no longer matches current selected work"))
			return result, nil
		}
		commits, err := store.ListRunJobCommits(ctx, resumePlan.RunID)
		if err != nil {
			return result, err
		}
		for _, commit := range commits {
			committed[commit.JobID] = true
		}
	}
	add("change_plan", true, nil)
	files := make(map[string]domain.File, len(current.Files))
	for _, file := range current.Files {
		files[file.Path] = file
	}
	scannerIDs := make(map[string]bool)
	featurePending, reviewPending, mappingPending := false, false, false
	for _, job := range plan.Plan.Jobs {
		if job.State == domain.JobReused || committed[job.JobID] {
			continue
		}
		if job.Stage == domain.StageMapping {
			mappingPending = true
		}
		if job.CoverageKey == nil {
			continue
		}
		switch job.Stage {
		case domain.StageFileAnalysis:
			if spec, ok := byCategory[categoryKey(job.CoverageKey.Family, job.CoverageKey.Category)]; ok && spec.Applicable(files[job.CoverageKey.FilePath]) {
				scannerIDs[spec.ID] = true
			}
		case domain.StageFeatureAnalysis:
			featurePending = true
		case domain.StageClawpatchReview:
			reviewPending = true
		}
	}
	clawpatchNeeded := mappingPending || reviewPending
	client, clientErr := clawpatch.New(root, clawpatch.Config{BinaryPath: config.ClawpatchBinary, EgressAuthorized: config.ClawpatchEgressAuthorized})
	if clientErr == nil && clawpatchNeeded {
		clientErr = client.CheckPrerequisites(ctx)
	}
	add("pinned_clawpatch", clawpatchNeeded, clientErr)
	var clawpatchEgressErr error
	if !config.ClawpatchEgressAuthorized {
		clawpatchEgressErr = errors.New("Clawpatch provider source egress is not authorized")
	}
	add("clawpatch_provider_egress", clawpatchNeeded, clawpatchEgressErr)
	add("contextual_provider", mappingPending || featurePending, contextual.CheckPrerequisites(config.Provider))
	ids := make([]string, 0, len(scannerIDs))
	for id := range scannerIDs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		runner, runnerErr := registry.Runner(id, "readiness")
		if runnerErr == nil {
			runnerErr = runner.CheckPrerequisites(ctx, root)
		}
		add(id, true, runnerErr)
	}
	return result, nil
}

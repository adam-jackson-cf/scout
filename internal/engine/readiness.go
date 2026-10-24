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
	"scout/internal/modelcli"
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
		options.Scope.SelectedFamilies = DefaultFamilies()
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

	var previousFiles []domain.File
	base := baselineState{}
	store, err := state.OpenExistingReadOnly(root)
	switch {
	case errors.Is(err, state.ErrNotFound):
		base.first = true
	case err != nil:
		add("baseline_state", true, errors.New("existing Scout state is unavailable"))
		return result, nil
	default:
		defer store.Close()
		base.run, err = store.LatestFinalizedRun(ctx)
		if errors.Is(err, state.ErrNotFound) {
			base.first = true
		} else if err == nil {
			previousFiles, err = store.LoadInventory(ctx, base.run.RunID)
			if err == nil {
				base.plan, err = store.LoadPlan(ctx, base.run.PlanID)
			}
		}
		if err != nil && !errors.Is(err, state.ErrNotFound) {
			add("baseline_state", true, errors.New("existing Scout baseline is unavailable"))
			return result, nil
		}
		if errors.Is(err, state.ErrNotFound) && !base.first {
			add("baseline_state", true, errors.New("latest Scout run lacks a finalized inventory"))
			return result, nil
		}
	}
	add("baseline_state", true, nil)
	features, err := clawpatch.LoadFeatures(clawpatch.StateDir(root))
	if err != nil {
		add("canonical_map_records", true, errors.New("canonical feature records are unavailable"))
		return result, nil
	}
	mapperDigest := digest(clawpatch.SafetyFingerprint())
	modelDigest := digest(config.Provider.Fingerprint())
	var resumePlan *domain.Plan
	if !options.ForceFull {
		resumePlan, err = matchingResumablePlan(ctx, store, initial.Snapshot, options.Scope, digest(string(toolBytes)), modelDigest, mapperDigest)
		if err != nil {
			return result, err
		}
	}
	mappingRequired := needsMapping(options, base, len(features), initial.Snapshot, mapperDigest)
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
	categories, err := selectCategories(specs, options.Scope.SelectedFamilies)
	if err != nil {
		add("change_plan", true, err)
		return result, nil
	}
	runID, err := newRunID()
	if err != nil {
		return result, err
	}
	if resumePlan != nil {
		runID = resumePlan.RunID
	}
	plan, err := planWork(planInput{
		runID: runID, current: current, features: features, diff: diff, dependencies: config.Dependencies,
		categories: categories, base: base, options: options, mappingRequired: mappingRequired,
		toolDigest: digest(string(toolBytes)), modelDigest: modelDigest, mapperDigest: mapperDigest,
	})
	if err != nil {
		add("change_plan", true, err)
		return result, nil
	}
	committed := make(map[string]bool)
	if resumePlan != nil {
		jobs := append([]domain.Job(nil), plan.Plan.Jobs...)
		jobs = append(jobs, finalInventoryJob(runID, current.Snapshot))
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
	reviewPending, mappingPending := false, false
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
			if spec, ok := categories.spec[categoryKey(job.CoverageKey.Family, job.CoverageKey.Category)]; ok && spec.Applies(files[job.CoverageKey.FilePath]) {
				scannerIDs[spec.ID] = true
			}
		case domain.StageClawpatchReview:
			reviewPending = true
		}
	}
	clawpatchNeeded := mappingPending || reviewPending
	client, clientErr := clawpatch.New(root, clawpatch.Config{BinaryPath: config.ClawpatchBinary, EgressAuthorized: config.ClawpatchEgressAuthorized, Provider: config.ClawpatchProvider})
	if clientErr == nil && clawpatchNeeded {
		clientErr = client.CheckPrerequisites(ctx)
	}
	add("pinned_clawpatch", clawpatchNeeded, clientErr)
	// Reasoning-LLM lanes are optional (design §3.4, §7.3): a missing model CLI
	// or withheld egress is reported, never makes the assessment not ready.
	_, _, cliErr := modelcli.Detect(config.Provider.CLI)
	add("model_cli", false, cliErr)
	var clawpatchEgressErr error
	if !config.ClawpatchEgressAuthorized {
		clawpatchEgressErr = errors.New("Clawpatch provider source egress is not authorized")
	}
	add("clawpatch_provider_egress", false, clawpatchEgressErr)
	add("contextual_provider", false, contextual.CheckPrerequisites(config.Provider))
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

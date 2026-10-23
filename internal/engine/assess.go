package engine

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"time"

	"scout/internal/analyzers"
	"scout/internal/clawpatch"
	"scout/internal/contextual"
	"scout/internal/domain"
	"scout/internal/inventory"
	"scout/internal/issues"
	"scout/internal/state"
)

// Assess persists the plan before file and feature jobs, then commits the run,
// its source index, observations and issue links. Mapping is canonical Clawpatch
// work; no Scout code synthesizes a feature ID or edits target source.
func Assess(ctx context.Context, config Config, options Options) (Result, error) {
	result := Result{}
	if options.Scope.Kind == "" {
		options.Scope.Kind = domain.ScopeRepository
	}
	if len(options.Scope.SelectedFamilies) == 0 {
		options.Scope.SelectedFamilies = DefaultFamilies()
	}
	if err := options.Scope.Validate(); err != nil {
		return result, err
	}
	registry := config.Registry
	if registry == nil {
		var err error
		registry, err = analyzers.BuiltinRegistry()
		if err != nil {
			return result, err
		}
	}
	store, err := state.Open(config.Root)
	if err != nil {
		return result, err
	}
	defer store.Close()
	root := store.Root()
	snapshot, err := inventory.Scan(ctx, root, inventory.ScanOptions{Scope: options.Scope, Exclusions: options.Exclusions})
	if err != nil {
		return result, err
	}
	_, recordedFiles, recordedErr := store.LoadLatestRecordedInventory(ctx)
	if recordedErr != nil && !errors.Is(recordedErr, state.ErrNotFound) {
		return result, recordedErr
	}
	latest, latestErr := store.LatestRun(ctx)
	first := errors.Is(latestErr, state.ErrNotFound)
	if latestErr != nil && !first {
		return result, latestErr
	}
	var previousFiles []domain.File
	var previousPlan domain.Plan
	if !first {
		previousFiles, err = store.LoadLatestInventory(ctx)
		if errors.Is(err, state.ErrNotFound) {
			first = true
			previousFiles = nil
		} else if err != nil {
			return result, err
		}
		previousPlan, err = store.LoadPlan(ctx, latest.PlanID)
		if err != nil {
			return result, err
		}
	}
	client, err := clawpatch.New(root, clawpatch.Config{BinaryPath: config.ClawpatchBinary, EgressAuthorized: config.ClawpatchEgressAuthorized})
	if err != nil {
		return result, err
	}
	previousFeatures, err := clawpatch.LoadFeatures(client.StateDir())
	if err != nil {
		return result, err
	}
	specs := registry.Specs()
	toolDigest, err := json.Marshal(specs)
	if err != nil {
		return result, err
	}
	modelDigest := digest(config.Provider.Endpoint + "\x00" + config.Provider.Model + "\x00" + config.Provider.PromptVersion)
	mapperDigest := digest(clawpatch.SafetyFingerprint())
	var resumePlan *domain.Plan
	if !options.ForceFull {
		resumePlan, err = matchingResumablePlan(ctx, store, snapshot.Snapshot, options.Scope, digest(string(toolDigest)), modelDigest, mapperDigest)
		if err != nil {
			return result, err
		}
	}
	mappingRequired := options.ForceFull || first || len(previousFeatures) == 0 || previousPlan.MapperDigest != mapperDigest || latest.Snapshot.WorktreeDigest != snapshot.Snapshot.WorktreeDigest || latest.Snapshot.IgnoreRulesDigest != snapshot.Snapshot.IgnoreRulesDigest || latest.Snapshot.ExclusionsDigest != snapshot.Snapshot.ExclusionsDigest
	mapped, mapErr, persistErr := mapBeforeAnalysis(ctx, store, client, snapshot, mappingRequired && resumePlan == nil, digest("scout-policy-v2"), digest(string(toolDigest)), modelDigest, mapperDigest)
	if persistErr != nil {
		return result, persistErr
	}
	result.Map = mapped
	features := mapped.Features
	if mapErr != nil && len(features) == 0 {
		features = previousFeatures
	}
	current, err := inventory.Scan(ctx, root, inventory.ScanOptions{Scope: options.Scope, Exclusions: options.Exclusions, Features: features})
	if err != nil {
		return result, err
	}
	diff, err := inventory.Diff(inventory.DiffInput{PreviousFiles: previousFiles, CurrentFiles: current.Files, PreviousFeatures: previousFeatures, CurrentFeatures: features, Dependencies: config.Dependencies})
	if err != nil {
		return result, err
	}
	result.Diff = diff
	// Identity follows renames/deletions against the inventory recorded before
	// this assessment began (loaded before the mapping preplan writes one).
	lifecycle := diff
	if recordedErr == nil {
		lifecycle, err = inventory.Diff(inventory.DiffInput{PreviousFiles: recordedFiles, CurrentFiles: current.Files})
		if err != nil {
			return result, err
		}
	}
	result.MappingGaps = mergeMappingGaps(diff.MappingGaps, current.MappingGaps)
	specByKey := make(map[string]analyzers.Spec, len(specs))
	fileCategories := map[domain.Family][]string{}
	selected := map[domain.Family]bool{}
	for _, family := range options.Scope.SelectedFamilies {
		selected[family] = true
	}
	for _, spec := range specs {
		if !selected[spec.Family] {
			continue
		}
		key := categoryKey(spec.Family, spec.Category)
		if _, exists := specByKey[key]; exists {
			return result, fmt.Errorf("multiple scanners claim %s", key)
		}
		specByKey[key] = spec
		fileCategories[spec.Family] = append(fileCategories[spec.Family], spec.Category)
	}
	featureCategories := map[domain.Family][]string{}
	for _, family := range options.Scope.SelectedFamilies {
		featureCategories[family] = contextualCategories[family]
	}
	runID, err := newRunID()
	if err != nil {
		return result, err
	}
	if resumePlan != nil {
		runID = resumePlan.RunID
	}
	baseline := make([]inventory.CachedCoverage, 0)
	if !first {
		fingerprints := map[string]string{}
		for _, job := range previousPlan.Jobs {
			if job.CoverageKey != nil {
				fingerprints[job.CoverageKey.String()] = job.InputFingerprint
			}
		}
		for _, row := range latest.Coverage {
			sourceRun := ""
			if row.Status == domain.CoverageReused {
				sourceRun = row.ReusedFromRunID
			} else if row.Status.Accounted() {
				sourceRun = latest.RunID
			}
			baseline = append(baseline, inventory.CachedCoverage{Coverage: row, SourceRunID: sourceRun, InputFingerprint: fingerprints[row.Key.String()]})
		}
	}
	planResult, err := inventory.Plan(inventory.PlanInput{RunID: runID, Snapshot: current.Snapshot, Scope: current.Scope, ScopeDigest: current.ScopeDigest, Files: current.Files, Features: features, Diff: diff, Dependencies: config.Dependencies, FileCategories: fileCategories, FeatureCategories: featureCategories, Baseline: baseline, ForceFull: options.ForceFull || first, MappingRequired: mappingRequired, ClawpatchReview: true, PolicyDigest: digest("scout-policy-v2"), ToolDigest: digest(string(toolDigest)), ModelDigest: modelDigest, MapperDigest: mapperDigest})
	if err != nil {
		return result, err
	}
	// Source drift is an explicit planned job so its failure is durable too.
	finalJob := domain.Job{
		JobID: "job-" + digest(runID+"\x00final-inventory"),
		Stage: domain.StageInventory, InputFingerprint: current.Snapshot.WorktreeDigest,
		State: domain.JobPending, InvalidationReason: "verify source snapshot after analysis",
	}
	planResult.Plan.Jobs = append(planResult.Plan.Jobs, finalJob)
	if mapErr != nil && !mappingRequired {
		planResult.Plan.Jobs = append(planResult.Plan.Jobs, domain.Job{
			JobID: "job-" + digest(runID+"\x00mapping-failure"),
			Stage: domain.StageMapping, InputFingerprint: mapperDigest,
			State: domain.JobPending, InvalidationReason: "canonical map could not be read",
		})
	}
	if resumePlan != nil {
		if planResult.Plan.PlanID != resumePlan.PlanID || !reflect.DeepEqual(planResult.Plan.Jobs, resumePlan.Jobs) {
			return result, errors.New("interrupted plan no longer matches canonical mapped inputs; a new full assessment is required")
		}
		planResult.Plan = *resumePlan
	} else if err := store.SavePlan(ctx, planResult.Plan); err != nil {
		return result, err
	}
	result.Plan = planResult.Plan
	started := time.Now().UTC()
	run := domain.Run{RunID: runID, PlanID: planResult.Plan.PlanID, Snapshot: current.Snapshot, Scope: current.Scope, State: domain.RunPartial, StartedAt: started, Coverage: append([]domain.Coverage(nil), planResult.Reused...), Omissions: append(append([]domain.Omission(nil), current.Omissions...), planResult.Omissions...)}
	for _, gap := range result.MappingGaps {
		key := domain.CoverageKey{FilePath: gap.Path, FeatureID: gap.FeatureID, Family: domain.FamilyClawpatch, Category: "mapping"}
		if key.Validate() == nil {
			run.Omissions = append(run.Omissions, domain.Omission{CoverageKey: key, Reason: gap.Reason})
		}
	}
	files := map[string]domain.File{}
	for _, file := range current.Files {
		files[file.Path] = file
	}
	featureByID := map[string]domain.Feature{}
	for _, feature := range features {
		featureByID[feature.FeatureID] = feature
	}
	priorByFeature, err := priorIssueObservations(ctx, store, features)
	if err != nil {
		return result, err
	}
	var provider *contextual.Client
	if config.Provider.Endpoint != "" {
		provider, err = contextual.NewClient(config.Provider)
		if err != nil {
			return result, err
		}
	}
	observations := []domain.Observation{}
	committed := make(map[string]bool)
	if resumePlan != nil {
		stored, loadErr := store.LoadInProgressRun(ctx, planResult.Plan.PlanID)
		if loadErr != nil && !errors.Is(loadErr, state.ErrNotFound) {
			return result, loadErr
		}
		if loadErr == nil {
			run = stored
			run.ExpectedCoverage = nil
			observed, observationErr := store.LoadObservations(ctx, runID)
			if observationErr != nil {
				return result, observationErr
			}
			observations = observed
			for _, ref := range run.ArtifactRefs {
				if ref.ContentType != "application/vnd.scout.adapter+json" {
					continue
				}
				raw, artifactErr := store.LoadArtifact(ctx, ref)
				if artifactErr != nil {
					return result, artifactErr
				}
				var adapter domain.AdapterResult
				if err := json.Unmarshal(raw, &adapter); err != nil {
					return result, err
				}
				result.Adapters = append(result.Adapters, adapter)
			}
			commits, commitErr := store.ListRunJobCommits(ctx, runID)
			if commitErr != nil {
				return result, commitErr
			}
			for _, commit := range commits {
				committed[commit.JobID] = true
			}
		} else if err := store.BeginRun(ctx, run); err != nil {
			return result, err
		}
	} else if err := store.BeginRun(ctx, run); err != nil {
		return result, err
	}
	for _, job := range planResult.Plan.Jobs {
		if job.Stage != domain.StageMapping || job.State == domain.JobReused || committed[job.JobID] {
			continue
		}
		commit := state.RunJobCommit{JobID: job.JobID, CommittedAt: time.Now().UTC()}
		if mapErr != nil {
			failure := domain.Failure{Stage: domain.StageMapping, JobID: job.JobID, Code: "mapping_blocked", Message: "canonical Clawpatch mapping did not complete"}
			commit.Failures = []domain.Failure{failure}
			run.Failures = append(run.Failures, failure)
		}
		if err := store.CommitRunJob(ctx, runID, commit); err != nil {
			return result, fmt.Errorf("persist mapping job %s: %w", job.JobID, err)
		}
	}
	// Pending scanner jobs are grouped per adapter and scanned in bounded
	// chunks; each job still commits only its own file's results.
	type scanChunk struct {
		files   []domain.File
		ran     bool
		adapter domain.AdapterResult
		found   []domain.Observation
	}
	chunkOf := map[string]*scanChunk{}
	runners := map[string]*analyzers.Runner{}
	pendingBySpec := map[string][]domain.File{}
	specOrder := []string{}
	for _, job := range planResult.Plan.Jobs {
		if job.CoverageKey == nil || job.State == domain.JobReused || committed[job.JobID] || job.Stage != domain.StageFileAnalysis {
			continue
		}
		spec, exists := specByKey[categoryKey(job.CoverageKey.Family, job.CoverageKey.Category)]
		if !exists {
			continue
		}
		if _, seen := pendingBySpec[spec.ID]; !seen {
			specOrder = append(specOrder, spec.ID)
		}
		pendingBySpec[spec.ID] = append(pendingBySpec[spec.ID], files[job.CoverageKey.FilePath])
	}
	for _, specID := range specOrder {
		var spec analyzers.Spec
		for _, candidate := range specs {
			if candidate.ID == specID {
				spec = candidate
			}
		}
		for _, batch := range spec.Batches(root, pendingBySpec[specID]) {
			chunk := &scanChunk{files: batch}
			for _, file := range batch {
				chunkOf[domain.CoverageKey{FilePath: file.Path, Family: spec.Family, Category: spec.Category}.String()] = chunk
			}
		}
	}
	for _, job := range planResult.Plan.Jobs {
		if job.CoverageKey == nil || job.State == domain.JobReused {
			continue
		}
		key := *job.CoverageKey
		run.ExpectedCoverage = append(run.ExpectedCoverage, key)
		if committed[job.JobID] {
			continue
		}
		coverageStart, observationStart, failureStart := len(run.Coverage), len(observations), len(run.Failures)
		artifactStart := len(run.ArtifactRefs)
		switch job.Stage {
		case domain.StageFileAnalysis:
			spec, exists := specByKey[categoryKey(key.Family, key.Category)]
			if !exists {
				run.Coverage = append(run.Coverage, blocked(key, "no registered scanner for category"))
				break
			}
			chunk := chunkOf[domain.CoverageKey{FilePath: key.FilePath, Family: key.Family, Category: key.Category}.String()]
			if chunk == nil {
				run.Coverage = append(run.Coverage, blocked(key, "no scanner batch was planned for this input"))
				break
			}
			if !chunk.ran {
				runner := runners[spec.ID]
				if runner == nil {
					var runnerErr error
					if runner, runnerErr = registry.Runner(spec.ID, runID); runnerErr != nil {
						return result, runnerErr
					}
					runners[spec.ID] = runner
				}
				chunk.adapter, chunk.found, _ = runner.Run(ctx, root, chunk.files)
				chunk.ran = true
			}
			adapter, found := splitAdapter(chunk.adapter, chunk.found, key)
			result.Adapters = append(result.Adapters, adapter)
			run.Coverage = append(run.Coverage, adapter.Coverage...)
			run.Failures = append(run.Failures, adapter.Failures...)
			observations = append(observations, found...)
			if err := storeAdapter(ctx, store, adapter, &run); err != nil {
				return result, err
			}
		case domain.StageFeatureAnalysis:
			if provider == nil {
				run.Coverage = append(run.Coverage, blocked(key, "contextual provider is not configured"))
				break
			}
			reviewed, reviewErr := provider.Review(ctx, contextual.ReviewRequest{
				Root: root, Snapshot: current.Snapshot.WorktreeDigest, RunID: runID,
				Feature: featureByID[key.FeatureID], Family: key.Family, Category: key.Category,
				ToolResults:       relevantAdapters(result.Adapters, featureByID[key.FeatureID], key.Family),
				PriorObservations: priorByFeature[key.FeatureID],
			})
			if reviewErr != nil {
				run.Failures = append(run.Failures, domain.Failure{Stage: domain.StageFeatureAnalysis, Code: "provider_incomplete", Message: "contextual inspection did not complete"})
			}
			if reviewed.Coverage.Status == "" {
				reviewed.Coverage = blocked(key, "contextual inspection returned no coverage")
			}
			if err := storeContextualReview(ctx, store, reviewed, &run); err != nil {
				return result, fmt.Errorf("persist contextual review manifest: %w", err)
			}
			run.Coverage = append(run.Coverage, reviewed.Coverage)
			observations = append(observations, reviewed.Observations...)
		case domain.StageClawpatchReview:
			reviewed, reviewErr := client.Review(ctx, []string{key.FeatureID})
			if reviewErr != nil {
				run.Failures = append(run.Failures, domain.Failure{Stage: domain.StageClawpatchReview, Code: "review_incomplete", Message: "Clawpatch review did not complete"})
			}
			if len(reviewed.Coverage) == 0 {
				run.Coverage = append(run.Coverage, blocked(key, "Clawpatch review returned no coverage"))
				break
			}
			row := reviewed.Coverage[0]
			for index := range reviewed.Observations {
				priorID := reviewed.Observations[index].ObservationID
				reviewed.Observations[index].RunID = runID
				reviewed.Observations[index].ObservationID = "clawpatch:" + runID + ":" + priorID
				for observationIndex, id := range row.ObservationIDs {
					if id == priorID {
						row.ObservationIDs[observationIndex] = reviewed.Observations[index].ObservationID
					}
				}
			}
			run.Coverage = append(run.Coverage, row)
			if err := storeClawpatchManifests(ctx, store, reviewed.Manifests, &run); err != nil {
				return result, fmt.Errorf("persist Clawpatch review manifest: %w", err)
			}
			observations = append(observations, reviewed.Observations...)
		default:
			run.Coverage = append(run.Coverage, blocked(key, "unsupported planned analysis stage"))
		}
		if err := store.CommitRunJob(ctx, runID, state.RunJobCommit{
			JobID: job.JobID, Coverage: run.Coverage[coverageStart:],
			Observations: observations[observationStart:], Failures: run.Failures[failureStart:],
			ArtifactRefs: run.ArtifactRefs[artifactStart:],
			CommittedAt:  time.Now().UTC(),
		}); err != nil {
			return result, fmt.Errorf("persist assessment job %s: %w", job.JobID, err)
		}
	}
	for _, row := range planResult.Reused {
		run.ExpectedCoverage = append(run.ExpectedCoverage, row.Key)
	}
	final, scanErr := inventory.Scan(ctx, root, inventory.ScanOptions{Scope: options.Scope, Exclusions: options.Exclusions, Features: features})
	var inventoryFailures []domain.Failure
	if scanErr != nil || final.Snapshot.WorktreeDigest != current.Snapshot.WorktreeDigest {
		if committed[finalJob.JobID] {
			return result, errors.New("source changed after the interrupted final inventory check")
		}
		failure := domain.Failure{Stage: domain.StageInventory, JobID: finalJob.JobID, Code: "source_drift", Message: "source changed during assessment or could not be rechecked"}
		inventoryFailures = append(inventoryFailures, failure)
		run.Failures = append(run.Failures, failure)
	}
	var inlineHealth []domain.LaneHealth
	if !committed[finalJob.JobID] {
		var inlineObservations []domain.Observation
		for _, lane := range options.Inline {
			found, laneErr := lane.Run(ctx, runID)
			inlineHealth = append(inlineHealth, inlineLaneHealth(lane, len(found), laneErr))
			if laneErr == nil {
				inlineObservations = append(inlineObservations, found...)
			}
		}
		observationStart := len(observations)
		observations = append(observations, inlineObservations...)
		if err := store.CommitRunJob(ctx, runID, state.RunJobCommit{JobID: finalJob.JobID, Observations: observations[observationStart:], Failures: inventoryFailures, CommittedAt: time.Now().UTC()}); err != nil {
			return result, fmt.Errorf("persist final source check: %w", err)
		}
	}
	for _, row := range run.Coverage {
		if !row.Status.Accounted() {
			run.State = domain.RunPartial
			break
		}
	}
	if len(run.Failures) == 0 && len(run.Omissions) == 0 {
		complete := true
		for _, row := range run.Coverage {
			if !row.Status.Accounted() {
				complete = false
				break
			}
		}
		if complete {
			run.State = domain.RunComplete
		}
	}
	finished := time.Now().UTC()
	run.FinishedAt = &finished
	if err := store.FinishRun(ctx, run, current.Files); err != nil {
		return result, err
	}
	result.Run = run
	renames := make(map[string]string, len(lifecycle.Renamed))
	for _, renamed := range lifecycle.Renamed {
		renames[renamed.PreviousPath] = renamed.Path
	}
	removed := make([]string, 0, len(lifecycle.Removed))
	for _, file := range lifecycle.Removed {
		removed = append(removed, file.Path)
	}
	if _, err := issues.Correlate(ctx, store, run, observations, issues.WithRenames(renames), issues.WithDeleted(removed)); err != nil {
		return result, fmt.Errorf("correlate persisted observations: %w", err)
	}
	allIssues, err := store.ListIssues(ctx)
	if err != nil {
		return result, err
	}
	observed := make(map[string]struct{})
	for _, observation := range observations {
		observed[observation.ObservationID] = struct{}{}
	}
	for _, row := range run.Coverage {
		for _, id := range row.ObservationIDs {
			observed[id] = struct{}{}
		}
	}
	result.Issues = make([]domain.Issue, 0)
	for _, issue := range allIssues {
		for _, id := range issue.ObservationIDs {
			if _, present := observed[id]; present {
				result.Issues = append(result.Issues, issue)
				break
			}
		}
	}
	result.Lanes = LaneHealthOf(options.Trigger, result.Run, inlineHealth)
	return result, nil
}

func matchingResumablePlan(ctx context.Context, store *state.Store, snapshot domain.Snapshot, scope domain.Scope, toolDigest, modelDigest, mapperDigest string) (*domain.Plan, error) {
	if store == nil {
		return nil, nil
	}
	plans, err := store.ListResumablePlans(ctx)
	if err != nil {
		return nil, err
	}
	for index := len(plans) - 1; index >= 0; index-- {
		plan := plans[index]
		if len(plan.Jobs) == 1 && plan.Jobs[0].Stage == domain.StageMapping {
			continue // A mapping preplan has no file or feature jobs to resume.
		}
		if plan.Snapshot == snapshot && reflect.DeepEqual(plan.Scope, scope) &&
			plan.PolicyDigest == digest("scout-policy-v2") && plan.ToolDigest == toolDigest &&
			plan.ModelDigest == modelDigest && plan.MapperDigest == mapperDigest {
			return &plan, nil
		}
	}
	return nil, nil
}

// mapBeforeAnalysis records a mapping-only partial run before an authorized
// mapper can inspect source. It never becomes a successful scan baseline.
func mapBeforeAnalysis(ctx context.Context, store *state.Store, client *clawpatch.Client, snapshot inventory.ScanResult, requested bool, policyDigest, toolDigest, modelDigest, mapperDigest string) (clawpatch.MapResult, error, error) {
	if !requested {
		mapped, err := client.Map(ctx, false, clawpatch.MapSourceHeuristic)
		return mapped, err, nil
	}
	runID, err := newRunID()
	if err != nil {
		return clawpatch.MapResult{}, nil, err
	}
	job := domain.Job{
		JobID: "job-" + digest(runID+"\x00canonical-mapping"), Stage: domain.StageMapping,
		InputFingerprint: mapperDigest, State: domain.JobPending,
		InvalidationReason: "canonical Clawpatch mapping inputs changed",
	}
	started := time.Now().UTC()
	plan := domain.Plan{
		PlanID: "plan-" + digest(runID+"\x00canonical-mapping"), RunID: runID,
		Snapshot: snapshot.Snapshot, Scope: snapshot.Scope, Jobs: []domain.Job{job},
		PolicyDigest: policyDigest, ToolDigest: toolDigest, ModelDigest: modelDigest,
		MapperDigest: mapperDigest, CreatedAt: started,
	}
	if err := store.SavePlan(ctx, plan); err != nil {
		return clawpatch.MapResult{}, nil, err
	}
	run := domain.Run{RunID: runID, PlanID: plan.PlanID, Snapshot: snapshot.Snapshot, Scope: snapshot.Scope, State: domain.RunPartial, StartedAt: started}
	if err := store.BeginRun(ctx, run); err != nil {
		return clawpatch.MapResult{}, nil, err
	}
	mapped, mapErr := client.Map(ctx, true, clawpatch.MapSourceHeuristic)
	commit := state.RunJobCommit{JobID: job.JobID, CommittedAt: time.Now().UTC()}
	if mapErr != nil {
		failure := domain.Failure{Stage: domain.StageMapping, JobID: job.JobID, Code: "mapping_blocked", Message: "canonical Clawpatch mapping did not complete"}
		run.Failures = append(run.Failures, failure)
		commit.Failures = []domain.Failure{failure}
	}
	if err := store.CommitRunJob(ctx, runID, commit); err != nil {
		return mapped, mapErr, fmt.Errorf("persist pre-mapping job: %w", err)
	}
	finished := time.Now().UTC()
	run.FinishedAt = &finished
	if err := store.FinishRun(ctx, run, snapshot.Files); err != nil {
		return mapped, mapErr, fmt.Errorf("finish pre-mapping run: %w", err)
	}
	return mapped, mapErr, nil
}

func newRunID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return "run-" + hex.EncodeToString(bytes[:]), nil
}
func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
func categoryKey(family domain.Family, category string) string {
	return string(family) + "\x00" + category
}
func blocked(key domain.CoverageKey, reason string) domain.Coverage {
	return domain.Coverage{Key: key, Status: domain.CoverageBlocked, Reason: reason}
}
func storeAdapter(ctx context.Context, store *state.Store, adapter domain.AdapterResult, run *domain.Run) error {
	encoded, err := json.Marshal(adapter)
	if err != nil {
		return err
	}
	ref, err := store.StoreArtifact(ctx, encoded, "application/vnd.scout.adapter+json", func(raw []byte) ([]byte, error) {
		var safe domain.AdapterResult
		if err := json.Unmarshal(raw, &safe); err != nil {
			return nil, err
		}
		if err := safe.Validate(); err != nil {
			return nil, err
		}
		return json.Marshal(safe)
	})
	if err != nil {
		return err
	}
	run.ArtifactRefs = append(run.ArtifactRefs, ref)
	return nil
}

// splitAdapter projects a batch scan onto one input: that file's coverage rows
// and observations, plus the batch-level failures that apply to every file in
// it. A file the batch never accounted for is incomplete, never inspected.
func splitAdapter(batch domain.AdapterResult, found []domain.Observation, key domain.CoverageKey) (domain.AdapterResult, []domain.Observation) {
	adapter := batch
	adapter.Coverage = nil
	for _, row := range batch.Coverage {
		if row.Key.FilePath == key.FilePath {
			adapter.Coverage = append(adapter.Coverage, row)
		}
	}
	if len(adapter.Coverage) == 0 {
		reason := "scanner batch did not account for this input"
		if len(batch.Failures) > 0 {
			reason = batch.Failures[0].Message
		}
		adapter.Coverage = []domain.Coverage{{
			Key: domain.CoverageKey{FilePath: key.FilePath, Family: key.Family, Category: key.Category}, Status: domain.CoverageIncomplete,
			Reason: reason, AdapterResultIDs: []string{batch.ResultID},
		}}
	}
	var mine []domain.Observation
	adapter.Observations = nil
	for _, observation := range found {
		if len(observation.Evidence) > 0 && observation.Evidence[0].Location.Path == key.FilePath {
			mine = append(mine, observation)
			adapter.Observations = append(adapter.Observations, observation.ObservationID)
		}
	}
	return adapter, mine
}

// SortedCategories exposes the versioned contextual work queue for readiness.
func SortedCategories() []string {
	var result []string
	for family, categories := range contextualCategories {
		for _, category := range categories {
			result = append(result, string(family)+":"+category)
		}
	}
	sort.Strings(result)
	return result
}

func mergeMappingGaps(groups ...[]inventory.MappingGap) []inventory.MappingGap {
	result := []inventory.MappingGap{}
	seen := make(map[string]bool)
	for _, group := range groups {
		for _, gap := range group {
			key := gap.Path + "\x00" + gap.FeatureID + "\x00" + gap.Reason
			if !seen[key] {
				seen[key] = true
				result = append(result, gap)
			}
		}
	}
	return result
}

func relevantAdapters(results []domain.AdapterResult, feature domain.Feature, family domain.Family) []domain.AdapterResult {
	paths := make(map[string]bool, len(feature.OwnedFiles)+len(feature.ContextFiles)+len(feature.Tests))
	for _, ref := range feature.OwnedFiles {
		paths[ref.Path] = true
	}
	for _, ref := range feature.ContextFiles {
		paths[ref.Path] = true
	}
	for _, ref := range feature.Tests {
		paths[ref.Path] = true
	}
	var relevant []domain.AdapterResult
	for _, adapter := range results {
		for _, coverage := range adapter.Coverage {
			if coverage.Key.Family == family && paths[coverage.Key.FilePath] {
				relevant = append(relevant, adapter)
				break
			}
		}
	}
	return relevant
}

func priorIssueObservations(ctx context.Context, store *state.Store, features []domain.Feature) (map[string][]domain.Observation, error) {
	issues, err := store.ListIssues(ctx)
	if err != nil {
		return nil, err
	}
	featurePaths := make(map[string]map[string]bool, len(features))
	for _, feature := range features {
		paths := make(map[string]bool)
		for _, ref := range feature.OwnedFiles {
			paths[ref.Path] = true
		}
		for _, ref := range feature.ContextFiles {
			paths[ref.Path] = true
		}
		for _, ref := range feature.Tests {
			paths[ref.Path] = true
		}
		featurePaths[feature.FeatureID] = paths
	}
	prior := make(map[string][]domain.Observation)
	for _, issue := range issues {
		if issue.State == domain.IssueResolved || issue.State == domain.IssueFalsePositive || issue.State == domain.IssueWontFix {
			continue
		}
		rows, err := store.ListIssueObservations(ctx, issue.IssueID)
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			continue
		}
		last := rows[len(rows)-1]
		for id, paths := range featurePaths {
			relevant := false
			for _, featureID := range issue.FeatureIDs {
				relevant = relevant || featureID == id
			}
			for _, path := range issue.Files {
				relevant = relevant || paths[path]
			}
			if relevant {
				prior[id] = append(prior[id], last)
			}
		}
	}
	return prior, nil
}

func storeContextualReview(ctx context.Context, store *state.Store, reviewed contextual.ReviewResult, run *domain.Run) error {
	if err := reviewed.Coverage.Validate(); err != nil {
		return err
	}
	for _, observation := range reviewed.Observations {
		if err := observation.Validate(); err != nil {
			return err
		}
	}
	encoded, err := json.Marshal(reviewed)
	if err != nil {
		return err
	}
	ref, err := store.StoreArtifact(ctx, encoded, "application/vnd.scout.contextual-review+json", func(raw []byte) ([]byte, error) {
		var safe contextual.ReviewResult
		if err := json.Unmarshal(raw, &safe); err != nil {
			return nil, err
		}
		if err := safe.Coverage.Validate(); err != nil {
			return nil, err
		}
		return json.Marshal(safe)
	})
	if err != nil {
		return err
	}
	run.ArtifactRefs = append(run.ArtifactRefs, ref)
	return nil
}

func storeClawpatchManifests(ctx context.Context, store *state.Store, manifests []clawpatch.ReviewManifest, run *domain.Run) error {
	if len(manifests) == 0 {
		return nil
	}
	encoded, err := json.Marshal(manifests)
	if err != nil {
		return err
	}
	ref, err := store.StoreArtifact(ctx, encoded, "application/vnd.scout.clawpatch-review+json", func(raw []byte) ([]byte, error) {
		var safe []clawpatch.ReviewManifest
		if err := json.Unmarshal(raw, &safe); err != nil {
			return nil, err
		}
		if len(safe) == 0 {
			return nil, fmt.Errorf("Clawpatch review manifest is empty")
		}
		return json.Marshal(safe)
	})
	if err != nil {
		return err
	}
	run.ArtifactRefs = append(run.ArtifactRefs, ref)
	return nil
}

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
	"time"

	"scout/internal/analyzers"
	"scout/internal/association"
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
func Assess(ctx context.Context, config Config, options Options) (result Result, err error) {
	if scopeErr := normalizeAssessScope(&options); scopeErr != nil {
		return result, scopeErr
	}
	registry, err := analyzers.BuiltinRegistry()
	if err != nil {
		return result, err
	}
	store, err := state.Open(config.Root)
	if err != nil {
		return result, err
	}
	defer func() {
		if closeErr := store.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()
	run := &assessment{ctx: ctx, config: config, options: options, registry: registry, store: store, root: store.Root()}
	runErr := run.execute()
	return run.result, runErr
}

func normalizeAssessScope(options *Options) error {
	if options.Scope.Kind == "" {
		options.Scope.Kind = domain.ScopeRepository
	}
	if len(options.Scope.SelectedFamilies) == 0 {
		options.Scope.SelectedFamilies = DefaultFamilies()
	}
	return options.Scope.Validate()
}

type assessDigests struct {
	tool   string
	model  string
	mapper string
}

type assessScanChunk struct {
	files   []domain.File
	ran     bool
	adapter domain.AdapterResult
	found   []domain.Observation
}

type assessment struct {
	ctx      context.Context
	config   Config
	options  Options
	registry *analyzers.Registry
	store    *state.Store
	root     string
	result   Result

	snapshot         inventory.ScanResult
	recordedFiles    []domain.File
	recordedErr      error
	previousFiles    []domain.File
	base             baselineState
	client           *clawpatch.Client
	previousFeatures []domain.Feature
	specs            []analyzers.Spec
	digests          assessDigests
	resumePlan       *domain.Plan
	mappingRequired  bool
	mapErr           error
	features         []domain.Feature

	current   inventory.ScanResult
	diff      inventory.DiffResult
	lifecycle inventory.DiffResult
	specByKey map[string]analyzers.Spec
	runID     string
	plan      inventory.PlanResult
	finalJob  domain.Job

	run               domain.Run
	files             map[string]domain.File
	featureByID       map[string]domain.Feature
	priorByFeature    map[string][]domain.Observation
	provider          *contextual.Client
	observations      []domain.Observation
	committed         map[string]bool
	chunkOf           map[string]*assessScanChunk
	runners           map[string]*analyzers.Runner
	inventoryFailures []domain.Failure
	inlineHealth      []domain.LaneHealth
}

func (a *assessment) execute() error {
	stages := []func() error{
		a.scanSnapshot, a.loadBaseline, a.loadMapperInputs, a.mapFeatures, a.diffInventory,
		a.planJobs, a.beginRun, a.commitMappingJobs, a.analyzeJobs, a.checkFinalSource,
		a.finishRun, a.correlateIssues, a.runAssociation,
	}
	for _, stage := range stages {
		if err := stage(); err != nil {
			return err
		}
	}
	return nil
}

func (a *assessment) scanSnapshot() error {
	snapshot, err := inventory.Scan(a.ctx, a.root, inventory.ScanOptions{Scope: a.options.Scope, Exclusions: a.options.Exclusions})
	if err != nil {
		return err
	}
	a.snapshot = snapshot
	return nil
}

func (a *assessment) loadBaseline() error {
	_, recordedFiles, recordedErr := a.store.LoadLatestRecordedInventory(a.ctx)
	if recordedErr != nil && !errors.Is(recordedErr, state.ErrNotFound) {
		return recordedErr
	}
	a.recordedFiles, a.recordedErr = recordedFiles, recordedErr
	latest, latestErr := a.store.LatestFinalizedRun(a.ctx)
	first := errors.Is(latestErr, state.ErrNotFound)
	if latestErr != nil && !first {
		return latestErr
	}
	var previousPlan domain.Plan
	if !first {
		previousFiles, err := a.store.LoadLatestInventory(a.ctx)
		if errors.Is(err, state.ErrNotFound) {
			first = true
			previousFiles = nil
		} else if err != nil {
			return err
		}
		a.previousFiles = previousFiles
		previousPlan, err = a.store.LoadPlan(a.ctx, latest.PlanID)
		if err != nil {
			return err
		}
	}
	a.base = baselineState{first: first, run: latest, plan: previousPlan}
	return nil
}

func (a *assessment) loadMapperInputs() error {
	client, err := clawpatch.New(a.root, clawpatch.Config{BinaryPath: a.config.ClawpatchBinary, EgressAuthorized: a.config.ClawpatchEgressAuthorized, Provider: a.config.ClawpatchProvider})
	if err != nil {
		return err
	}
	a.client = client
	a.previousFeatures, err = clawpatch.LoadFeatures(client.StateDir())
	if err != nil {
		return err
	}
	a.specs = a.registry.Specs()
	toolDigest, err := json.Marshal(a.specs)
	if err != nil {
		return err
	}
	a.digests = assessDigests{tool: digest(string(toolDigest)), model: digest(a.config.Provider.Fingerprint()), mapper: digest(clawpatch.SafetyFingerprint())}
	if !a.options.ForceFull {
		a.resumePlan, err = findResumablePlan(a.ctx, a.store, resumablePlanKey{snapshot: a.snapshot.Snapshot, scope: a.options.Scope, digests: a.digests})
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *assessment) mapFeatures() error {
	a.mappingRequired = needsMapping(a.options, a.base, len(a.previousFeatures), a.snapshot.Snapshot, a.digests.mapper)
	outcome, persistErr := mapBeforeAnalysis(a.ctx, a.store, a.client, preMappingRequest{
		snapshot: a.snapshot, requested: a.mappingRequired && a.resumePlan == nil, digests: a.digests,
	})
	if persistErr != nil {
		return persistErr
	}
	mapped := outcome.result
	a.mapErr = outcome.err
	a.result.Map = mapped
	a.features = mapped.Features
	if a.mapErr != nil && len(a.features) == 0 {
		a.features = a.previousFeatures
	}
	return nil
}

func (a *assessment) diffInventory() error {
	current, err := inventory.Scan(a.ctx, a.root, inventory.ScanOptions{Scope: a.options.Scope, Exclusions: a.options.Exclusions, Features: a.features})
	if err != nil {
		return err
	}
	a.current = current
	diff, err := inventory.Diff(inventory.DiffInput{PreviousFiles: a.previousFiles, CurrentFiles: current.Files, PreviousFeatures: a.previousFeatures, CurrentFeatures: a.features, Dependencies: a.config.Dependencies})
	if err != nil {
		return err
	}
	a.diff = diff
	a.result.Diff = diff
	// Identity follows renames/deletions against the inventory recorded before
	// this assessment began (loaded before the mapping preplan writes one).
	a.lifecycle = diff
	if a.recordedErr == nil {
		a.lifecycle, err = inventory.Diff(inventory.DiffInput{PreviousFiles: a.recordedFiles, CurrentFiles: current.Files})
		if err != nil {
			return err
		}
	}
	a.result.MappingGaps = mergeMappingGaps(diff.MappingGaps, current.MappingGaps)
	return nil
}

func (a *assessment) planJobs() error {
	categories, err := selectCategories(a.specs, a.options.Scope.SelectedFamilies)
	if err != nil {
		return err
	}
	a.specByKey = categories.spec
	runID, err := newRunID()
	if err != nil {
		return err
	}
	if a.resumePlan != nil {
		runID = a.resumePlan.RunID
	}
	a.runID = runID
	a.plan, err = planWork(planInput{
		runID: runID, current: a.current, features: a.features, diff: a.diff, dependencies: a.config.Dependencies,
		categories: categories, base: a.base, options: a.options, mappingRequired: a.mappingRequired,
		toolDigest: a.digests.tool, modelDigest: a.digests.model, mapperDigest: a.digests.mapper,
	})
	if err != nil {
		return err
	}
	// Source drift is an explicit planned job so its failure is durable too.
	a.finalJob = finalInventoryJob(runID, a.current.Snapshot)
	a.plan.Plan.Jobs = append(a.plan.Plan.Jobs, a.finalJob)
	if a.mapErr != nil && !a.mappingRequired {
		a.plan.Plan.Jobs = append(a.plan.Plan.Jobs, domain.Job{
			JobID: "job-" + digest(runID+"\x00mapping-failure"),
			Stage: domain.StageMapping, InputFingerprint: a.digests.mapper,
			State: domain.JobPending, InvalidationReason: "canonical map could not be read",
		})
	}
	if err := a.adoptPlan(); err != nil {
		return err
	}
	a.result.Plan = a.plan.Plan
	return nil
}

func (a *assessment) adoptPlan() error {
	if a.resumePlan == nil {
		return a.store.SavePlan(a.ctx, a.plan.Plan)
	}
	if a.plan.Plan.PlanID != a.resumePlan.PlanID || !reflect.DeepEqual(a.plan.Plan.Jobs, a.resumePlan.Jobs) {
		return errors.New("interrupted plan no longer matches canonical mapped inputs; a new full assessment is required")
	}
	a.plan.Plan = *a.resumePlan
	return nil
}

func (a *assessment) beginRun() error {
	a.initRunRecord()
	prior, err := priorIssueObservations(a.ctx, a.store, a.features)
	if err != nil {
		return err
	}
	a.priorByFeature = prior
	a.provider, err = contextual.NewClient(a.config.Provider)
	if err != nil {
		return err
	}
	a.observations = []domain.Observation{}
	a.committed = make(map[string]bool)
	if a.resumePlan == nil {
		return a.store.BeginRun(a.ctx, a.run)
	}
	stored, loadErr := a.store.LoadInProgressRun(a.ctx, a.plan.Plan.PlanID)
	if loadErr != nil && !errors.Is(loadErr, state.ErrNotFound) {
		return loadErr
	}
	if loadErr != nil {
		return a.store.BeginRun(a.ctx, a.run)
	}
	return a.resumeRun(stored)
}

func (a *assessment) initRunRecord() {
	started := time.Now().UTC()
	current, planResult := a.current, a.plan
	a.run = domain.Run{RunID: a.runID, PlanID: planResult.Plan.PlanID, Snapshot: current.Snapshot, Scope: current.Scope, State: domain.RunPartial, StartedAt: started, Coverage: append([]domain.Coverage(nil), planResult.Reused...), Omissions: append(append([]domain.Omission(nil), current.Omissions...), planResult.Omissions...)}
	for _, gap := range a.result.MappingGaps {
		key := domain.CoverageKey{FilePath: gap.Path, FeatureID: gap.FeatureID, Family: domain.FamilyClawpatch, Category: "mapping"}
		if key.Validate() == nil {
			a.run.Omissions = append(a.run.Omissions, domain.Omission{CoverageKey: key, Reason: gap.Reason})
		}
	}
	a.files = map[string]domain.File{}
	for _, file := range current.Files {
		a.files[file.Path] = file
	}
	a.featureByID = map[string]domain.Feature{}
	for _, feature := range a.features {
		a.featureByID[feature.FeatureID] = feature
	}
}

func (a *assessment) resumeRun(stored domain.Run) error {
	a.run = stored
	a.run.ExpectedCoverage = nil
	observed, observationErr := a.store.LoadObservations(a.ctx, a.runID)
	if observationErr != nil {
		return observationErr
	}
	a.observations = observed
	if err := a.reloadAdapters(); err != nil {
		return err
	}
	commits, commitErr := a.store.ListRunJobCommits(a.ctx, a.runID)
	if commitErr != nil {
		return commitErr
	}
	for _, commit := range commits {
		a.committed[commit.JobID] = true
	}
	return nil
}

func (a *assessment) reloadAdapters() error {
	for _, ref := range a.run.ArtifactRefs {
		if ref.ContentType != "application/vnd.scout.adapter+json" {
			continue
		}
		raw, artifactErr := a.store.LoadArtifact(a.ctx, ref)
		if artifactErr != nil {
			return artifactErr
		}
		var adapter domain.AdapterResult
		if err := json.Unmarshal(raw, &adapter); err != nil {
			return err
		}
		a.result.Adapters = append(a.result.Adapters, adapter)
	}
	return nil
}

func (a *assessment) commitMappingJobs() error {
	for _, job := range a.plan.Plan.Jobs {
		if job.Stage != domain.StageMapping || job.State == domain.JobReused || a.committed[job.JobID] {
			continue
		}
		commit := state.RunJobCommit{JobID: job.JobID, CommittedAt: time.Now().UTC()}
		if a.mapErr != nil {
			failure := mappingBlockedFailure(job.JobID)
			commit.Failures = []domain.Failure{failure}
			a.run.Failures = append(a.run.Failures, failure)
		}
		if err := a.store.CommitRunJob(a.ctx, a.runID, commit); err != nil {
			return fmt.Errorf("persist mapping job %s: %w", job.JobID, err)
		}
	}
	return nil
}

func (a *assessment) planScanChunks() {
	// Pending scanner jobs are grouped per adapter and scanned in bounded
	// chunks; each job still commits only its own file's results.
	pendingBySpec, specOrder := a.pendingScanInputs()
	a.chunkOf = map[string]*assessScanChunk{}
	a.runners = map[string]*analyzers.Runner{}
	for _, specID := range specOrder {
		spec := a.specWithID(specID)
		for _, batch := range spec.Batches(a.root, pendingBySpec[specID]) {
			chunk := &assessScanChunk{files: batch}
			for _, file := range batch {
				a.chunkOf[domain.CoverageKey{FilePath: file.Path, Family: spec.Family, Category: spec.Category}.String()] = chunk
			}
		}
	}
}

func (a *assessment) pendingScanInputs() (map[string][]domain.File, []string) {
	pendingBySpec := map[string][]domain.File{}
	specOrder := []string{}
	for _, job := range a.plan.Plan.Jobs {
		if job.CoverageKey == nil || job.State == domain.JobReused || a.committed[job.JobID] || job.Stage != domain.StageFileAnalysis {
			continue
		}
		spec, exists := a.specByKey[categoryKey(job.CoverageKey.Family, job.CoverageKey.Category)]
		if !exists {
			continue
		}
		if _, seen := pendingBySpec[spec.ID]; !seen {
			specOrder = append(specOrder, spec.ID)
		}
		pendingBySpec[spec.ID] = append(pendingBySpec[spec.ID], a.files[job.CoverageKey.FilePath])
	}
	return pendingBySpec, specOrder
}

func (a *assessment) specWithID(specID string) analyzers.Spec {
	var spec analyzers.Spec
	for _, candidate := range a.specs {
		if candidate.ID == specID {
			spec = candidate
		}
	}
	return spec
}

func (a *assessment) analyzeJobs() error {
	a.planScanChunks()
	for _, job := range a.plan.Plan.Jobs {
		if job.CoverageKey == nil || job.State == domain.JobReused {
			continue
		}
		key := *job.CoverageKey
		a.run.ExpectedCoverage = append(a.run.ExpectedCoverage, key)
		if a.committed[job.JobID] {
			continue
		}
		if err := a.analyzeJob(job, key); err != nil {
			return err
		}
	}
	for _, row := range a.plan.Reused {
		a.run.ExpectedCoverage = append(a.run.ExpectedCoverage, row.Key)
	}
	return nil
}

func (a *assessment) analyzeJob(job domain.Job, key domain.CoverageKey) error {
	coverageStart, observationStart, failureStart := len(a.run.Coverage), len(a.observations), len(a.run.Failures)
	artifactStart := len(a.run.ArtifactRefs)
	if err := a.runJobStage(job.Stage, key); err != nil {
		return err
	}
	if err := a.store.CommitRunJob(a.ctx, a.runID, state.RunJobCommit{
		JobID: job.JobID, Coverage: a.run.Coverage[coverageStart:],
		Observations: a.observations[observationStart:], Failures: a.run.Failures[failureStart:],
		ArtifactRefs: a.run.ArtifactRefs[artifactStart:],
		CommittedAt:  time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("persist assessment job %s: %w", job.JobID, err)
	}
	return nil
}

func (a *assessment) runJobStage(stage domain.Stage, key domain.CoverageKey) error {
	switch stage {
	case domain.StageFileAnalysis:
		return a.analyzeFile(key)
	case domain.StageFeatureAnalysis:
		return a.reviewFeature(key)
	case domain.StageClawpatchReview:
		return a.reviewWithClawpatch(key)
	default:
		a.run.Coverage = append(a.run.Coverage, blocked(key, "unsupported planned analysis stage"))
		return nil
	}
}

func (a *assessment) analyzeFile(key domain.CoverageKey) error {
	spec, exists := a.specByKey[categoryKey(key.Family, key.Category)]
	if !exists {
		a.run.Coverage = append(a.run.Coverage, blocked(key, "no registered scanner for category"))
		return nil
	}
	chunk := a.chunkOf[domain.CoverageKey{FilePath: key.FilePath, Family: key.Family, Category: key.Category}.String()]
	if chunk == nil {
		a.run.Coverage = append(a.run.Coverage, blocked(key, "no scanner batch was planned for this input"))
		return nil
	}
	if !chunk.ran {
		if err := a.runScanChunk(spec, chunk); err != nil {
			return err
		}
	}
	adapter, found := splitAdapter(chunk.adapter, chunk.found, key)
	a.result.Adapters = append(a.result.Adapters, adapter)
	a.run.Coverage = append(a.run.Coverage, adapter.Coverage...)
	a.run.Failures = append(a.run.Failures, adapter.Failures...)
	a.observations = append(a.observations, found...)
	return storeAdapter(a.ctx, a.store, adapter, &a.run)
}

func (a *assessment) runScanChunk(spec analyzers.Spec, chunk *assessScanChunk) error {
	runner := a.runners[spec.ID]
	if runner == nil {
		var runnerErr error
		if runner, runnerErr = a.registry.Runner(spec.ID, a.runID); runnerErr != nil {
			return runnerErr
		}
		a.runners[spec.ID] = runner
	}
	var runErr error
	chunk.adapter, chunk.found, runErr = runner.Run(a.ctx, a.root, chunk.files)
	recordScanError(&chunk.adapter, runErr)
	chunk.ran = true
	return nil
}

func (a *assessment) reviewFeature(key domain.CoverageKey) error {
	if !a.config.laneOptedIn(LaneContextual) {
		a.run.Coverage = append(a.run.Coverage, notOptedIn(key))
		return nil
	}
	reviewed, reviewErr := a.provider.Review(a.ctx, contextual.ReviewRequest{
		Root: a.root, Snapshot: a.current.Snapshot.WorktreeDigest, RunID: a.runID,
		Feature: a.featureByID[key.FeatureID], Family: key.Family, Category: key.Category,
		ToolResults:       relevantAdapters(a.result.Adapters, a.featureByID[key.FeatureID], key.Family),
		PriorObservations: a.priorByFeature[key.FeatureID],
	})
	if reviewErr != nil && reviewed.Coverage.Status != domain.CoverageSkipped && reviewed.Coverage.Status != domain.CoverageBlocked {
		a.run.Failures = append(a.run.Failures, domain.Failure{Stage: domain.StageFeatureAnalysis, Code: "provider_incomplete", Message: "contextual inspection did not complete"})
	}
	if reviewed.Coverage.Status == "" {
		reviewed.Coverage = blocked(key, "contextual inspection returned no coverage")
	}
	if err := storeContextualReview(a.ctx, a.store, reviewed, &a.run); err != nil {
		return fmt.Errorf("persist contextual review manifest: %w", err)
	}
	a.run.Coverage = append(a.run.Coverage, reviewed.Coverage)
	a.observations = append(a.observations, reviewed.Observations...)
	return nil
}

func (a *assessment) reviewWithClawpatch(key domain.CoverageKey) error {
	if !a.config.laneOptedIn(LaneClawpatchReview) {
		a.run.Coverage = append(a.run.Coverage, notOptedIn(key))
		return nil
	}
	reviewed, reviewErr := a.client.Review(a.ctx, []string{key.FeatureID})
	if reviewErr != nil {
		a.run.Failures = append(a.run.Failures, domain.Failure{Stage: domain.StageClawpatchReview, Code: "review_incomplete", Message: "Clawpatch review did not complete"})
	}
	if len(reviewed.Coverage) == 0 {
		a.run.Coverage = append(a.run.Coverage, blocked(key, "Clawpatch review returned no coverage"))
		return nil
	}
	row := reviewed.Coverage[0]
	rebindClawpatchObservations(reviewed.Observations, row.ObservationIDs, a.runID)
	a.run.Coverage = append(a.run.Coverage, row)
	if err := storeClawpatchManifests(a.ctx, a.store, reviewed.Manifests, &a.run); err != nil {
		return fmt.Errorf("persist Clawpatch review manifest: %w", err)
	}
	a.observations = append(a.observations, reviewed.Observations...)
	return nil
}

func rebindClawpatchObservations(observations []domain.Observation, rowIDs []string, runID string) {
	for index := range observations {
		priorID := observations[index].ObservationID
		observations[index].RunID = runID
		observations[index].ObservationID = "clawpatch:" + runID + ":" + priorID
		for observationIndex, id := range rowIDs {
			if id == priorID {
				rowIDs[observationIndex] = observations[index].ObservationID
			}
		}
	}
}

func (a *assessment) checkFinalSource() error {
	final, scanErr := inventory.Scan(a.ctx, a.root, inventory.ScanOptions{Scope: a.options.Scope, Exclusions: a.options.Exclusions, Features: a.features})
	if scanErr != nil || final.Snapshot.WorktreeDigest != a.current.Snapshot.WorktreeDigest {
		if a.committed[a.finalJob.JobID] {
			return errors.New("source changed after the interrupted final inventory check")
		}
		failure := domain.Failure{Stage: domain.StageInventory, JobID: a.finalJob.JobID, Code: failureSourceDrift, Message: "source changed during assessment or could not be rechecked"}
		a.inventoryFailures = append(a.inventoryFailures, failure)
		a.run.Failures = append(a.run.Failures, failure)
	}
	if a.committed[a.finalJob.JobID] {
		return nil
	}
	return a.runInlineLanes()
}

func (a *assessment) runInlineLanes() error {
	var inlineObservations []domain.Observation
	for _, lane := range a.options.Inline {
		found, laneErr := lane.Run(a.ctx, a.runID)
		a.inlineHealth = append(a.inlineHealth, inlineLaneHealth(lane, found, laneErr))
		if laneErr == nil {
			inlineObservations = append(inlineObservations, found.Observations...)
		}
	}
	if a.options.Measures != nil {
		if len(a.inventoryFailures) > 0 {
			a.inlineHealth = append(a.inlineHealth, domain.LaneHealth{Lane: LaneMeasures, Status: domain.LaneSkipped, Reason: "source changed during assessment"})
		} else {
			a.inlineHealth = append(a.inlineHealth, runMeasuresLanes(a.ctx, measuresLaneInput{store: a.store, root: a.root, current: a.current, features: a.features, options: a.options})...)
		}
	}
	observationStart := len(a.observations)
	a.observations = append(a.observations, inlineObservations...)
	if err := a.store.CommitRunJob(a.ctx, a.runID, state.RunJobCommit{JobID: a.finalJob.JobID, Observations: a.observations[observationStart:], Failures: a.inventoryFailures, CommittedAt: time.Now().UTC()}); err != nil {
		return fmt.Errorf("persist final source check: %w", err)
	}
	return nil
}

func (a *assessment) finishRun() error {
	settleAssessedRunState(&a.run)
	finished := time.Now().UTC()
	a.run.FinishedAt = &finished
	if err := a.store.FinishRun(a.ctx, a.run, a.current.Files); err != nil {
		return err
	}
	a.result.Run = a.run
	return nil
}

func settleAssessedRunState(run *domain.Run) {
	for _, row := range run.Coverage {
		if !row.Status.Accounted() {
			run.State = domain.RunPartial
			return
		}
	}
	if len(run.Failures) == 0 && len(run.Omissions) == 0 {
		run.State = domain.RunComplete
	}
}

func (a *assessment) correlateIssues() error {
	renames := make(map[string]string, len(a.lifecycle.Renamed))
	for _, renamed := range a.lifecycle.Renamed {
		renames[renamed.PreviousPath] = renamed.Path
	}
	removed := make([]string, 0, len(a.lifecycle.Removed))
	for _, file := range a.lifecycle.Removed {
		removed = append(removed, file.Path)
	}
	if _, err := issues.Correlate(a.ctx, a.store, a.run, a.observations, issues.WithRenames(renames), issues.WithDeleted(removed)); err != nil {
		return fmt.Errorf("correlate persisted observations: %w", err)
	}
	allIssues, err := a.store.ListIssues(a.ctx)
	if err != nil {
		return err
	}
	a.result.Issues = issuesObservedInRun(allIssues, assessedObservationIDs(a.observations, a.run.Coverage))
	return nil
}

func assessedObservationIDs(observations []domain.Observation, coverage []domain.Coverage) map[string]struct{} {
	observed := make(map[string]struct{})
	for _, observation := range observations {
		observed[observation.ObservationID] = struct{}{}
	}
	for _, row := range coverage {
		for _, id := range row.ObservationIDs {
			observed[id] = struct{}{}
		}
	}
	return observed
}

func issuesObservedInRun(allIssues []domain.Issue, observed map[string]struct{}) []domain.Issue {
	matched := make([]domain.Issue, 0)
	for _, issue := range allIssues {
		for _, id := range issue.ObservationIDs {
			if _, present := observed[id]; present {
				matched = append(matched, issue)
				break
			}
		}
	}
	return matched
}

func (a *assessment) runAssociation() error {
	associationOptions := association.LaneOptions{}
	if a.options.Association != nil {
		associationOptions = *a.options.Association
	}
	associationOptions.RunID = a.run.RunID
	trigger := a.options.Trigger
	if trigger == "" {
		trigger = domain.TriggerPush
	}
	associationHealth, associationErr := RunAssociationLane(a.ctx, a.store, a.root, trigger, associationOptions)
	if associationErr != nil {
		for i := range associationHealth {
			associationHealth[i].Status = domain.LaneFailed
			associationHealth[i].Reason = associationErr.Error()
		}
	}
	a.inlineHealth = append(a.inlineHealth, associationHealth...)
	a.result.Lanes = LaneHealthOf(a.options.Trigger, a.result.Run, a.inlineHealth)
	var err error
	a.result.Associations, a.result.AssociationGuard, err = associationOutput(a.ctx, a.store, a.config.associationGuard())
	return err
}

type resumablePlanKey struct {
	snapshot domain.Snapshot
	scope    domain.Scope
	digests  assessDigests
}

func findResumablePlan(ctx context.Context, store *state.Store, key resumablePlanKey) (*domain.Plan, error) {
	if store == nil {
		return nil, nil
	}
	plans, err := store.ListResumablePlans(ctx)
	if err != nil {
		return nil, err
	}
	want := domain.Plan{Snapshot: key.snapshot, Scope: key.scope, PolicyDigest: digest(policyDigestSeed), ToolDigest: key.digests.tool, ModelDigest: key.digests.model, MapperDigest: key.digests.mapper}
	for index := len(plans) - 1; index >= 0; index-- {
		plan := plans[index]
		if len(plan.Jobs) == 1 && plan.Jobs[0].Stage == domain.StageMapping {
			continue // A mapping preplan has no file or feature jobs to resume.
		}
		if resumablePlanMatches(plan, want) {
			return &plan, nil
		}
	}
	return nil, nil
}

func resumablePlanMatches(plan, want domain.Plan) bool {
	return plan.Snapshot == want.Snapshot && reflect.DeepEqual(plan.Scope, want.Scope) &&
		plan.PolicyDigest == want.PolicyDigest && plan.ToolDigest == want.ToolDigest &&
		plan.ModelDigest == want.ModelDigest && plan.MapperDigest == want.MapperDigest
}

// mapOutcome is a canonical map attempt: err is the mapping failure, which the
// run records and survives; it is distinct from a persistence failure.
type mapOutcome struct {
	result clawpatch.MapResult
	err    error
}

// mapBeforeAnalysis records a mapping-only partial run before an authorized
// mapper can inspect source. It never becomes a successful scan baseline. The
// returned error is a persistence failure; the mapping failure is in the outcome.
type preMappingRequest struct {
	snapshot  inventory.ScanResult
	requested bool
	digests   assessDigests
}

func mapBeforeAnalysis(ctx context.Context, store *state.Store, client *clawpatch.Client, request preMappingRequest) (mapOutcome, error) {
	snapshot, digests := request.snapshot, request.digests
	if !request.requested {
		mapped, err := client.Map(ctx, false, clawpatch.MapSourceHeuristic)
		return mapOutcome{result: mapped, err: err}, nil
	}
	runID, err := newRunID()
	if err != nil {
		return mapOutcome{}, err
	}
	job := domain.Job{
		JobID: "job-" + digest(runID+"\x00canonical-mapping"), Stage: domain.StageMapping,
		InputFingerprint: digests.mapper, State: domain.JobPending,
		InvalidationReason: "canonical Clawpatch mapping inputs changed",
	}
	started := time.Now().UTC()
	plan := domain.Plan{
		PlanID: "plan-" + digest(runID+"\x00canonical-mapping"), RunID: runID,
		Snapshot: snapshot.Snapshot, Scope: snapshot.Scope, Jobs: []domain.Job{job},
		PolicyDigest: digest(policyDigestSeed), ToolDigest: digests.tool, ModelDigest: digests.model,
		MapperDigest: digests.mapper, CreatedAt: started,
	}
	if err := store.SavePlan(ctx, plan); err != nil {
		return mapOutcome{}, err
	}
	run := domain.Run{RunID: runID, PlanID: plan.PlanID, Snapshot: snapshot.Snapshot, Scope: snapshot.Scope, State: domain.RunPartial, StartedAt: started}
	if err := store.BeginRun(ctx, run); err != nil {
		return mapOutcome{}, err
	}
	mapped, mapErr := client.Map(ctx, true, clawpatch.MapSourceHeuristic)
	outcome := mapOutcome{result: mapped, err: mapErr}
	commit := state.RunJobCommit{JobID: job.JobID, CommittedAt: time.Now().UTC()}
	if mapErr != nil {
		failure := mappingBlockedFailure(job.JobID)
		run.Failures = append(run.Failures, failure)
		commit.Failures = []domain.Failure{failure}
	}
	if err := store.CommitRunJob(ctx, runID, commit); err != nil {
		return outcome, fmt.Errorf("persist pre-mapping job: %w", err)
	}
	finished := time.Now().UTC()
	run.FinishedAt = &finished
	if err := store.FinishRun(ctx, run, snapshot.Files); err != nil {
		return outcome, fmt.Errorf("finish pre-mapping run: %w", err)
	}
	return outcome, nil
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
func notOptedIn(key domain.CoverageKey) domain.Coverage {
	return domain.Coverage{Key: key, Status: domain.CoverageSkipped, Reason: ReasonNotOptedIn}
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
		featurePaths[feature.FeatureID] = featurePathSet(feature)
	}
	prior := make(map[string][]domain.Observation)
	for _, issue := range issues {
		if issueClosedToReview(issue.State) {
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
			if issueTouchesFeature(issue, id, paths) {
				prior[id] = append(prior[id], last)
			}
		}
	}
	return prior, nil
}

func featurePathSet(feature domain.Feature) map[string]bool {
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
	return paths
}

func issueClosedToReview(issueState domain.IssueState) bool {
	return issueState == domain.IssueResolved || issueState == domain.IssueFalsePositive || issueState == domain.IssueWontFix
}

func issueTouchesFeature(issue domain.Issue, featureID string, paths map[string]bool) bool {
	relevant := false
	for _, id := range issue.FeatureIDs {
		relevant = relevant || id == featureID
	}
	for _, path := range issue.Files {
		relevant = relevant || paths[path]
	}
	return relevant
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
			return nil, fmt.Errorf("clawpatch review manifest is empty")
		}
		return json.Marshal(safe)
	})
	if err != nil {
		return err
	}
	run.ArtifactRefs = append(run.ArtifactRefs, ref)
	return nil
}

// recordScanError keeps a scanner batch error visible in the run: a batch that
// returned an error without recording a failure of its own gets one, so the
// error is never reduced to an unexplained result.
func recordScanError(adapter *domain.AdapterResult, err error) {
	if err == nil || len(adapter.Failures) > 0 {
		return
	}
	adapter.Failures = []domain.Failure{{Stage: domain.StageFileAnalysis, Code: failureScannerFailed, Message: "scanner batch did not complete"}}
}

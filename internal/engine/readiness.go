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
	"scout/internal/association"
	"scout/internal/clawpatch"
	"scout/internal/contextual"
	"scout/internal/domain"
	"scout/internal/inventory"
	"scout/internal/modelcli"
	"scout/internal/policy"
	"scout/internal/state"
)

type readinessCheck struct {
	ctx           context.Context
	config        Config
	options       Options
	result        Readiness
	root          string
	initial       inventory.ScanResult
	registry      *analyzers.Registry
	specs         []analyzers.Spec
	toolBytes     []byte
	store         *state.Store
	base          baselineState
	previousFiles []domain.File
	features      []domain.Feature
	current       inventory.ScanResult
	diffResult    inventory.DiffResult
	categories    plannedCategories
	plan          inventory.PlanResult
	committed     map[string]bool
}

// AssessReadiness checks only prerequisites needed by the selected pending work.
// It does not initialize state, map features, scan with analyzers, or call a model.
func AssessReadiness(ctx context.Context, config Config, options Options) (result Readiness, err error) {
	check := &readinessCheck{ctx: ctx, config: config, options: options, result: Readiness{Ready: true, Capabilities: []Capability{}}}
	check.result.AssociationGuard = association.Guard(nil, nil, config.associationGuard(), association.GuardState{})
	defer func() {
		if check.store != nil {
			err = errors.Join(err, check.store.Close())
		}
	}()
	err = check.run()
	return check.result, err
}

func (r *readinessCheck) add(name string, required bool, err error) {
	entry := Capability{Name: name, Required: required, Ready: err == nil}
	if err != nil {
		entry.Reason = err.Error()
		if required {
			r.result.Ready = false
		}
	}
	r.result.Capabilities = append(r.result.Capabilities, entry)
}

func (r *readinessCheck) run() error {
	stages := []func() (bool, error){r.checkScope, r.checkRepository, r.checkInventory, r.checkScannerCatalog, r.checkActivation, r.checkBaseline, r.checkChangePlan}
	for _, stage := range stages {
		next, err := stage()
		if err != nil || !next {
			return err
		}
	}
	r.checkScanners(r.pendingScanners())
	return nil
}

func (r *readinessCheck) checkScope() (bool, error) {
	if r.options.Scope.Kind == "" {
		r.options.Scope.Kind = domain.ScopeRepository
	}
	if len(r.options.Scope.SelectedFamilies) == 0 {
		r.options.Scope.SelectedFamilies = DefaultFamilies()
	}
	return true, r.options.Scope.Validate()
}

func (r *readinessCheck) checkRepository() (bool, error) {
	root, err := filepath.Abs(r.config.Root)
	if err != nil {
		return false, err
	}
	r.root = root
	info, err := os.Stat(root)
	if err == nil && !info.IsDir() {
		err = errors.New("assessment root is not a directory")
	}
	r.add("repository", true, err)
	return err == nil, nil
}

func (r *readinessCheck) checkInventory() (bool, error) {
	initial, err := inventory.Scan(r.ctx, r.root, inventory.ScanOptions{Scope: r.options.Scope, Exclusions: r.options.Exclusions})
	if err == nil && len(initial.Omissions) != 0 {
		err = errors.New("repository inventory has unreadable inputs")
	}
	r.initial = initial
	r.add("inventory", true, err)
	return err == nil, nil
}

func (r *readinessCheck) checkScannerCatalog() (bool, error) {
	registry, err := analyzers.BuiltinRegistry()
	r.add("scanner_catalog", true, err)
	if err != nil {
		return false, nil
	}
	r.registry = registry
	r.specs = registry.Specs()
	return true, nil
}

func (r *readinessCheck) checkActivation() (bool, error) {
	pol, err := policy.Load(r.root, "HEAD")
	if err != nil {
		return false, err
	}
	for _, spec := range r.specs {
		for _, ruleID := range spec.Rules.RuleIDs {
			checkID := CheckIDOf(spec.ID, ruleID)
			if reason := pol.ActivationDiagnostic(checkID, spec.Rules.Version); reason != "" {
				r.result.Capabilities = append(r.result.Capabilities, Capability{Name: checkID, Ready: true, Reason: reason, PublicRuleVersion: spec.Rules.Version})
			}
		}
	}
	r.toolBytes, err = json.Marshal(r.specs)
	return err == nil, err
}

func (r *readinessCheck) checkBaseline() (bool, error) {
	store, err := state.OpenExistingReadOnly(r.root)
	switch {
	case errors.Is(err, state.ErrNotFound):
		r.base.first = true
	case err != nil:
		r.add("baseline_state", true, errors.New("existing Scout state is unavailable"))
		return false, nil
	default:
		r.store = store
		if next, err := r.loadBaseline(); err != nil || !next {
			return next, err
		}
	}
	r.add("baseline_state", true, nil)
	return true, nil
}

func (r *readinessCheck) loadBaseline() (bool, error) {
	var err error
	_, r.result.AssociationGuard, err = associationOutput(r.ctx, r.store, r.config.associationGuard())
	if err != nil {
		return false, err
	}
	r.base.run, err = r.store.LatestFinalizedRun(r.ctx)
	if errors.Is(err, state.ErrNotFound) {
		r.base.first = true
	} else if err == nil {
		r.previousFiles, err = r.store.LoadInventory(r.ctx, r.base.run.RunID)
		if err == nil {
			r.base.plan, err = r.store.LoadPlan(r.ctx, r.base.run.PlanID)
		}
	}
	if err != nil && !errors.Is(err, state.ErrNotFound) {
		r.add("baseline_state", true, errors.New("existing Scout baseline is unavailable"))
		return false, nil
	}
	if errors.Is(err, state.ErrNotFound) && !r.base.first {
		r.add("baseline_state", true, errors.New("latest Scout run lacks a finalized inventory"))
		return false, nil
	}
	return true, nil
}

func (r *readinessCheck) checkChangePlan() (bool, error) {
	features, err := clawpatch.LoadFeatures(clawpatch.StateDir(r.root))
	if err != nil {
		r.add("canonical_map_records", true, errors.New("canonical feature records are unavailable"))
		return false, nil
	}
	r.features = features
	mapperDigest := digest(clawpatch.SafetyFingerprint())
	modelDigest := digest(r.config.Provider.Fingerprint())
	var resumePlan *domain.Plan
	if !r.options.ForceFull {
		resumePlan, err = findResumablePlan(r.ctx, r.store, resumablePlanKey{snapshot: r.initial.Snapshot, scope: r.options.Scope, digests: assessDigests{tool: digest(string(r.toolBytes)), model: modelDigest, mapper: mapperDigest}})
		if err != nil {
			return false, err
		}
	}
	mappingRequired := needsMapping(r.options, r.base, len(features), r.initial.Snapshot, mapperDigest)
	if next := r.loadPlanInputs(); !next {
		return false, nil
	}
	runID, err := newRunID()
	if err != nil {
		return false, err
	}
	if resumePlan != nil {
		runID = resumePlan.RunID
	}
	plan, err := planWork(planInput{
		runID: runID, current: r.current, features: features, diff: r.diffResult, dependencies: r.config.Dependencies,
		categories: r.categories, base: r.base, options: r.options, mappingRequired: mappingRequired,
		toolDigest: digest(string(r.toolBytes)), modelDigest: modelDigest, mapperDigest: mapperDigest,
	})
	if err != nil {
		r.add("change_plan", true, err)
		return false, nil
	}
	r.plan = plan
	if next, err := r.resumeCommitted(runID, resumePlan); err != nil || !next {
		return next, err
	}
	r.add("change_plan", true, nil)
	return true, nil
}

func (r *readinessCheck) loadPlanInputs() bool {
	current, err := inventory.Scan(r.ctx, r.root, inventory.ScanOptions{Scope: r.options.Scope, Exclusions: r.options.Exclusions, Features: r.features})
	if err != nil {
		r.add("mapped_inventory", true, errors.New("feature-linked inventory is unavailable"))
		return false
	}
	r.current = current
	diff, err := inventory.Diff(inventory.DiffInput{PreviousFiles: r.previousFiles, CurrentFiles: current.Files, PreviousFeatures: r.features, CurrentFeatures: r.features, Dependencies: r.config.Dependencies})
	if err != nil {
		r.add("change_plan", true, err)
		return false
	}
	r.diffResult = diff
	categories, err := selectCategories(r.specs, r.options.Scope.SelectedFamilies)
	if err != nil {
		r.add("change_plan", true, err)
		return false
	}
	r.categories = categories
	return true
}

func (r *readinessCheck) resumeCommitted(runID string, resumePlan *domain.Plan) (bool, error) {
	r.committed = make(map[string]bool)
	if resumePlan == nil {
		return true, nil
	}
	jobs := append([]domain.Job(nil), r.plan.Plan.Jobs...)
	jobs = append(jobs, finalInventoryJob(runID, r.current.Snapshot))
	if r.plan.Plan.PlanID != resumePlan.PlanID || !reflect.DeepEqual(jobs, resumePlan.Jobs) {
		r.add("change_plan", true, errors.New("interrupted plan no longer matches current selected work"))
		return false, nil
	}
	commits, err := r.store.ListRunJobCommits(r.ctx, resumePlan.RunID)
	if err != nil {
		return false, err
	}
	for _, commit := range commits {
		r.committed[commit.JobID] = true
	}
	return true, nil
}

type pendingReadinessWork struct {
	scannerIDs      map[string]bool
	clawpatchNeeded bool
}

func (r *readinessCheck) pendingScanners() pendingReadinessWork {
	files := make(map[string]domain.File, len(r.current.Files))
	for _, file := range r.current.Files {
		files[file.Path] = file
	}
	pending := pendingReadinessWork{scannerIDs: make(map[string]bool)}
	for _, job := range r.plan.Plan.Jobs {
		if job.State == domain.JobReused || r.committed[job.JobID] {
			continue
		}
		if job.Stage == domain.StageMapping {
			pending.clawpatchNeeded = true
		}
		if job.CoverageKey != nil {
			r.notePendingCoverage(&pending, job, files)
		}
	}
	return pending
}

func (r *readinessCheck) notePendingCoverage(pending *pendingReadinessWork, job domain.Job, files map[string]domain.File) {
	switch job.Stage {
	case domain.StageFileAnalysis:
		if spec, ok := r.categories.spec[categoryKey(job.CoverageKey.Family, job.CoverageKey.Category)]; ok && spec.Applies(files[job.CoverageKey.FilePath]) {
			pending.scannerIDs[spec.ID] = true
		}
	case domain.StageClawpatchReview:
		pending.clawpatchNeeded = true
	}
}

func (r *readinessCheck) checkScanners(pending pendingReadinessWork) {
	client, clientErr := clawpatch.New(r.root, clawpatch.Config{BinaryPath: r.config.ClawpatchBinary, EgressAuthorized: r.config.ClawpatchEgressAuthorized, Provider: r.config.ClawpatchProvider})
	if clientErr == nil && pending.clawpatchNeeded {
		clientErr = client.CheckPrerequisites(r.ctx)
	}
	r.add("pinned_clawpatch", pending.clawpatchNeeded, clientErr)
	// Reasoning-LLM lanes are optional (REQ-18): a missing model CLI
	// or withheld egress is reported, never makes the assessment not ready.
	_, _, cliErr := modelcli.Detect(r.config.Provider.CLI)
	r.add("model_cli", false, cliErr)
	var clawpatchEgressErr error
	if !r.config.ClawpatchEgressAuthorized {
		clawpatchEgressErr = errors.New("clawpatch provider source egress is not authorized")
	}
	r.add("clawpatch_provider_egress", false, clawpatchEgressErr)
	r.add("contextual_provider", false, contextual.CheckPrerequisites(r.config.Provider))
	ids := make([]string, 0, len(pending.scannerIDs))
	for id := range pending.scannerIDs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		runner, runnerErr := r.registry.Runner(id, "readiness")
		if runnerErr == nil {
			runnerErr = runner.CheckPrerequisites(r.ctx, r.root)
		}
		r.add(id, true, runnerErr)
	}
}

package engine

import (
	"fmt"

	"scout/internal/analyzers"
	"scout/internal/domain"
	"scout/internal/inventory"
)

// Failure codes recorded on runs; lane health reads them back.
const (
	failureSourceDrift    = "source_drift"
	failureMappingBlocked = "mapping_blocked"
	failureScannerFailed  = "scanner_failed"
)

// policyDigestSeed names the decision policy that planned work depends on.
const policyDigestSeed = "scout-policy-v2"

// baselineState is the finalized run an assessment compares against; first is
// true when there is none.
type baselineState struct {
	first bool
	run   domain.Run
	plan  domain.Plan
}

// needsMapping reports whether canonical mapping must run before analysis.
func needsMapping(options Options, base baselineState, featureCount int, snapshot domain.Snapshot, mapperDigest string) bool {
	return options.ForceFull || base.first || featureCount == 0 ||
		base.plan.MapperDigest != mapperDigest ||
		base.run.Snapshot.WorktreeDigest != snapshot.WorktreeDigest ||
		base.run.Snapshot.IgnoreRulesDigest != snapshot.IgnoreRulesDigest ||
		base.run.Snapshot.ExclusionsDigest != snapshot.ExclusionsDigest
}

// cachedCoverage turns the baseline run's coverage into reusable plan input.
func cachedCoverage(base baselineState) []inventory.CachedCoverage {
	cached := make([]inventory.CachedCoverage, 0)
	if base.first {
		return cached
	}
	fingerprints := map[string]string{}
	for _, job := range base.plan.Jobs {
		if job.CoverageKey != nil {
			fingerprints[job.CoverageKey.String()] = job.InputFingerprint
		}
	}
	for _, row := range base.run.Coverage {
		sourceRun := ""
		if row.Status == domain.CoverageReused {
			sourceRun = row.ReusedFromRunID
		} else if row.Status.Accounted() {
			sourceRun = base.run.RunID
		}
		cached = append(cached, inventory.CachedCoverage{Coverage: row, SourceRunID: sourceRun, InputFingerprint: fingerprints[row.Key.String()]})
	}
	return cached
}

// plannedCategories is the scanner and contextual work the selected families
// ask for, with the scanner spec that owns each file category.
type plannedCategories struct {
	file    map[domain.Family][]string
	feature map[domain.Family][]string
	spec    map[string]analyzers.Spec
}

func selectCategories(specs []analyzers.Spec, families []domain.Family) (plannedCategories, error) {
	planned := plannedCategories{
		file:    map[domain.Family][]string{},
		feature: map[domain.Family][]string{},
		spec:    make(map[string]analyzers.Spec, len(specs)),
	}
	selected := map[domain.Family]bool{}
	for _, family := range families {
		selected[family] = true
		planned.feature[family] = contextualCategories[family]
	}
	for _, spec := range specs {
		if !selected[spec.Family] {
			continue
		}
		key := categoryKey(spec.Family, spec.Category)
		if _, exists := planned.spec[key]; exists {
			return plannedCategories{}, fmt.Errorf("multiple scanners claim %s", key)
		}
		planned.spec[key] = spec
		planned.file[spec.Family] = append(planned.file[spec.Family], spec.Category)
	}
	return planned, nil
}

// planInput assembles the inventory plan input shared by assessment and readiness.
type planInput struct {
	runID           string
	current         inventory.ScanResult
	features        []domain.Feature
	diff            inventory.DiffResult
	dependencies    []inventory.DependencyEdge
	categories      plannedCategories
	base            baselineState
	options         Options
	mappingRequired bool
	toolDigest      string
	modelDigest     string
	mapperDigest    string
}

func planWork(in planInput) (inventory.PlanResult, error) {
	return inventory.Plan(inventory.PlanInput{
		RunID: in.runID, Snapshot: in.current.Snapshot, Scope: in.current.Scope, ScopeDigest: in.current.ScopeDigest,
		Files: in.current.Files, Features: in.features, Diff: in.diff, Dependencies: in.dependencies,
		FileCategories: in.categories.file, FeatureCategories: in.categories.feature, Baseline: cachedCoverage(in.base),
		ForceFull: in.options.ForceFull || in.base.first, MappingRequired: in.mappingRequired, ClawpatchReview: true,
		PolicyDigest: digest(policyDigestSeed), ToolDigest: in.toolDigest, ModelDigest: in.modelDigest, MapperDigest: in.mapperDigest,
	})
}

// finalInventoryJob is the planned source-drift check, so its failure is durable.
func finalInventoryJob(runID string, snapshot domain.Snapshot) domain.Job {
	return domain.Job{
		JobID: "job-" + digest(runID+"\x00final-inventory"),
		Stage: domain.StageInventory, InputFingerprint: snapshot.WorktreeDigest,
		State: domain.JobPending, InvalidationReason: "verify source snapshot after analysis",
	}
}

func mappingBlockedFailure(jobID string) domain.Failure {
	return domain.Failure{Stage: domain.StageMapping, JobID: jobID, Code: failureMappingBlocked, Message: "canonical Clawpatch mapping did not complete"}
}

package engine

import (
	"context"
	"fmt"

	"scout/internal/artifacts"
	"scout/internal/domain"
	"scout/internal/history"
	"scout/internal/inventory"
	"scout/internal/measures"
	"scout/internal/prmeta"
	"scout/internal/state"
)

// MeasuresOptions enables the measures and history signal lanes.
type MeasuresOptions struct {
	// Only, when non-nil, restricts measurement to these repository-relative
	// paths (pull requests measure the changed files); nil measures the whole
	// inventory.
	Only  []string
	Tools measures.Tools
	// History also computes the §8.3 exposure and interest measures over the
	// repository's commit history (push and nightly runs).
	History bool
	// PRMetadata, when set, supplies merged-PR metadata for pr_effort,
	// batch_size and the PR part of dispersion (§8.3). Nil: those signals are
	// `unmeasured`.
	PRMetadata *PRMetadataOptions
	// Artifacts are the coverage and test-result files to ingest for this
	// run (design §8.2); only revision-matched coverage yields scout.crap.
	Artifacts []artifacts.Spec
}

// PRMetadataOptions selects the PR-metadata source: a file, or the opt-in
// GitHub fetch (Fetch.Enabled must be set explicitly; Fetch.Token comes from
// GITHUB_TOKEN in the caller). File wins when both are given.
type PRMetadataOptions struct {
	File  string
	Fetch prmeta.FetchOptions
}

// runMeasuresLanes computes the §8.2 principal signals over the run's
// inventory, and optionally the §8.3 history signals, and records them. Both
// lanes are signal-mode: a failure is reported in lane health and never fails
// the run.
func runMeasuresLanes(ctx context.Context, store *state.Store, root string, current inventory.ScanResult, features []domain.Feature, options Options) []domain.LaneHealth {
	files := current.Files
	if options.Measures.Only != nil {
		keep := make(map[string]bool, len(options.Measures.Only))
		for _, path := range options.Measures.Only {
			keep[path] = true
		}
		files = nil
		for _, file := range current.Files {
			if keep[file.Path] {
				files = append(files, file)
			}
		}
	}
	coverage, artifactNote := ingestArtifacts(ctx, store, root, current.Snapshot, options.Measures.Artifacts)
	result := measures.Run(ctx, measures.Config{
		Root: root, Snapshot: current.Snapshot, RunScope: string(options.Scope.Kind),
		Files: files, Features: features, Tools: options.Measures.Tools, Coverage: coverage,
	})
	health := domain.LaneHealth{Lane: LaneMeasures, Units: len(result.Eligible), Status: domain.LaneRan}
	switch {
	case result.Candidates == 0:
		health.Status, health.Reason = domain.LaneSkipped, "no production source files in scope"
	case result.LizardErr != "":
		health.Status, health.Reason = domain.LaneFailed, "lizard: "+result.LizardErr
	case result.JscpdErr != "":
		health.Status, health.Reason = domain.LaneFailed, "jscpd: "+result.JscpdErr
	}
	if err := store.RecordSignals(ctx, result.Signals); err != nil {
		health.Status, health.Reason = domain.LaneFailed, fmt.Sprintf("record signals: %v", err)
	}
	out := []domain.LaneHealth{health}
	if options.Measures.History {
		out = append(out, runHistoryLane(ctx, store, root, current, files, result, options))
	}
	if artifactNote != "" && out[0].Reason == "" {
		out[0].Reason = artifactNote
	}
	return out
}

// runHistoryLane computes exposure and interest measures per scope from git
// history. Hotspot uses the per-file CC of the measures run; it is
// `unmeasured` when lizard did not run.
func runHistoryLane(ctx context.Context, store *state.Store, root string, current inventory.ScanResult, files []domain.File, result measures.Result, options Options) domain.LaneHealth {
	health := domain.LaneHealth{Lane: LaneHistory, Status: domain.LaneRan}
	commits, err := history.Load(ctx, root)
	if err != nil {
		health.Status, health.Reason = domain.LaneFailed, "git history: "+err.Error()
		return health
	}
	if len(commits) == 0 {
		health.Status, health.Reason = domain.LaneSkipped, "repository has no commits"
		return health
	}
	health.Units = len(commits)
	var scopes []history.Scope
	for _, scope := range measures.BuildScopes(measures.Candidates(files), nil) {
		scopes = append(scopes, history.FileSetScope(scope.Kind, scope.Anchor, scope.Paths))
	}
	var metrics map[string]history.FileMetric
	if result.Lizard != nil {
		metrics = make(map[string]history.FileMetric, len(result.Lizard))
		for path, measure := range result.Lizard {
			metrics[path] = history.FileMetric{CC: float64(measure.FileCC()), NLOC: measure.NLOC}
		}
	}
	opts := history.Options{
		Scopes: scopes, Eligible: measures.IsProductionSource,
		Snapshot: measures.SnapshotKey(current.Snapshot), Salt: current.Snapshot.RepositoryID,
		Files: metrics, ToolInfo: history.GitVersion(ctx),
	}
	var prNote string
	opts.PRs, opts.PRSource, prNote = loadPRMetadata(ctx, store, options.Measures.PRMetadata, opts.Snapshot)
	signals := history.Compute(commits, opts)
	if prNote != "" {
		health.Status, health.Reason = domain.LaneFailed, prNote
	}
	rework, err := history.Rework(ctx, root, commits, opts)
	if err != nil {
		health.Status, health.Reason = domain.LaneFailed, "rework: "+err.Error()
	}
	signals = append(signals, rework...)
	if err := store.RecordSignals(ctx, signals); err != nil {
		health.Status, health.Reason = domain.LaneFailed, fmt.Sprintf("record signals: %v", err)
	}
	return health
}

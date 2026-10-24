package inventory

import (
	"testing"

	"scout/internal/domain"
)

func TestForceFullReschedulesCachedFile(t *testing.T) {
	file := domain.File{Path: "main.go", Language: "go", ContentDigest: "abc", Eligible: true}
	snapshot := domain.Snapshot{RepositoryID: "repository", WorktreeDigest: "worktree", IgnoreRulesDigest: "ignore", ExclusionsDigest: "exclude"}
	input := PlanInput{ScopeDigest: "scope", RunID: "first", Snapshot: snapshot, Scope: domain.Scope{Kind: domain.ScopeRepository}, Files: []domain.File{file}, Diff: DiffResult{Added: []domain.File{file}}, FileCategories: map[domain.Family][]string{domain.FamilyQuality: {"complexity"}}, ForceFull: true}
	first, err := Plan(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Plan.Jobs) != 1 || first.Plan.Jobs[0].State != domain.JobPending {
		t.Fatalf("initial analysis not scheduled: %+v", first.Plan.Jobs)
	}
	job := first.Plan.Jobs[0]
	input.RunID = "cached"
	input.ForceFull = false
	input.Diff = DiffResult{}
	input.Baseline = []CachedCoverage{{Coverage: domain.Coverage{Key: *job.CoverageKey, Status: domain.CoverageInspected, InspectedRanges: []domain.SourceRange{{Path: file.Path, StartLine: 1, EndLine: 1}}}, SourceRunID: "first", InputFingerprint: job.InputFingerprint}}
	cached, err := Plan(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(cached.Reused) != 1 || cached.Plan.Jobs[0].State != domain.JobReused {
		t.Fatal("unchanged accounted unit was not reused")
	}
	input.RunID = "forced"
	input.ForceFull = true
	forced, err := Plan(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(forced.Reused) != 0 || len(forced.Plan.Jobs) != 1 || forced.Plan.Jobs[0].State != domain.JobPending {
		t.Fatal("forced-full analysis reused cached evidence")
	}
}

func TestNewCategoryIsPlannedWithoutSourceChanges(t *testing.T) {
	file := domain.File{Path: ".github/workflows/release.yml", Language: "yaml", ContentDigest: "abc", Eligible: true}
	input := PlanInput{
		ScopeDigest:    "scope",
		RunID:          "first",
		Snapshot:       domain.Snapshot{RepositoryID: "repository", WorktreeDigest: "worktree", IgnoreRulesDigest: "ignore", ExclusionsDigest: "exclude"},
		Scope:          domain.Scope{Kind: domain.ScopeRepository},
		Files:          []domain.File{file},
		Diff:           DiffResult{Added: []domain.File{file}},
		FileCategories: map[domain.Family][]string{domain.FamilySecurity: {"credential-leakage"}},
		ForceFull:      true,
	}
	first, err := Plan(input)
	if err != nil {
		t.Fatal(err)
	}
	prior := first.Plan.Jobs[0]
	input.RunID = "expanded"
	input.ForceFull = false
	input.Diff = DiffResult{}
	input.FileCategories = map[domain.Family][]string{domain.FamilySecurity: {"credential-leakage", "workflow-supply-chain"}}
	input.Baseline = []CachedCoverage{{
		Coverage: domain.Coverage{
			Key: *prior.CoverageKey, Status: domain.CoverageInspected,
			InspectedRanges: []domain.SourceRange{{Path: file.Path, StartLine: 1, EndLine: 1}},
		},
		SourceRunID: "first", InputFingerprint: prior.InputFingerprint,
	}}
	expanded, err := Plan(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(expanded.Plan.Jobs) != 2 {
		t.Fatalf("new category omitted on unchanged source: %+v", expanded.Plan.Jobs)
	}
	for _, job := range expanded.Plan.Jobs {
		if job.CoverageKey.Category == "workflow-supply-chain" && job.State != domain.JobPending {
			t.Fatalf("new category must be analyzed: %+v", job)
		}
		if job.CoverageKey.Category == "credential-leakage" && job.State != domain.JobReused {
			t.Fatalf("unchanged category should reuse prior inspection: %+v", job)
		}
	}
}

func TestNewFeatureCategoryIsPlannedWithoutMapChanges(t *testing.T) {
	feature := domain.Feature{
		SchemaVersion: 1, FeatureID: "feature-1", Kind: "config", Source: "heuristic",
		Confidence:  domain.ConfidenceHigh,
		RawOriginal: []byte(`{"schemaVersion":1,"featureId":"feature-1","kind":"config","source":"heuristic"}`),
	}
	input := PlanInput{
		ScopeDigest:       "scope",
		RunID:             "first",
		Snapshot:          domain.Snapshot{RepositoryID: "repository", WorktreeDigest: "worktree", IgnoreRulesDigest: "ignore", ExclusionsDigest: "exclude"},
		Scope:             domain.Scope{Kind: domain.ScopeRepository},
		Features:          []domain.Feature{feature},
		FeatureCategories: map[domain.Family][]string{domain.FamilySecurity: {"trust-boundary"}},
		ForceFull:         true,
	}
	first, err := Plan(input)
	if err != nil {
		t.Fatal(err)
	}
	prior := first.Plan.Jobs[0]
	input.RunID = "expanded"
	input.ForceFull = false
	input.FeatureCategories = map[domain.Family][]string{domain.FamilySecurity: {"trust-boundary", "data-flow"}}
	input.Baseline = []CachedCoverage{{
		Coverage: domain.Coverage{
			Key: *prior.CoverageKey, Status: domain.CoverageInspected,
			InspectedRanges: []domain.SourceRange{{Path: "main.go", StartLine: 1, EndLine: 1}},
		},
		SourceRunID: "first", InputFingerprint: prior.InputFingerprint,
	}}
	expanded, err := Plan(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(expanded.Plan.Jobs) != 2 {
		t.Fatalf("new feature category omitted on unchanged map: %+v", expanded.Plan.Jobs)
	}
	for _, job := range expanded.Plan.Jobs {
		if job.CoverageKey.Category == "data-flow" && job.State != domain.JobPending {
			t.Fatalf("new category must be analyzed: %+v", job)
		}
		if job.CoverageKey.Category == "trust-boundary" && job.State != domain.JobReused {
			t.Fatalf("unchanged category should reuse prior inspection: %+v", job)
		}
	}
}

func TestPlanDependencyFootprintInvalidatesAffectedDependentFeature(t *testing.T) {
	snapshot := domain.Snapshot{RepositoryID: "repository", WorktreeDigest: "worktree", IgnoreRulesDigest: "ignore", ExclusionsDigest: "exclude"}
	consumer := testFeature("consumer", "consumer/handler.go")
	input := PlanInput{
		ScopeDigest: "scope",
		RunID:       "first",
		Snapshot:    snapshot,
		Scope:       domain.Scope{Kind: domain.ScopeRepository},
		Files: []domain.File{
			{Path: "shared/contract.go", ContentDigest: "contract-v1", Eligible: true, OwnedFeatureIDs: []string{"producer"}},
			{Path: "consumer/handler.go", ContentDigest: "consumer-v1", Eligible: true, OwnedFeatureIDs: []string{"consumer"}},
		},
		Features:          []domain.Feature{consumer},
		Dependencies:      []DependencyEdge{{FromFeatureID: "producer", ToFeatureID: "consumer", Reason: "consumer implements the shared contract", InputPaths: []string{"shared/contract.go"}}},
		FeatureCategories: map[domain.Family][]string{domain.FamilySecurity: {"trust-boundary"}},
	}
	first, err := Plan(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Plan.Jobs) != 1 || first.Plan.Jobs[0].State != domain.JobPending {
		t.Fatalf("dependent feature was not scheduled: %+v", first.Plan.Jobs)
	}
	job := first.Plan.Jobs[0]
	input.RunID = "unchanged"
	input.Baseline = []CachedCoverage{{
		Coverage: domain.Coverage{
			Key:             *job.CoverageKey,
			Status:          domain.CoverageInspected,
			InspectedRanges: []domain.SourceRange{{Path: "consumer/handler.go", StartLine: 1, EndLine: 1}},
		},
		SourceRunID:      "first",
		InputFingerprint: job.InputFingerprint,
	}}
	unchanged, err := Plan(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(unchanged.Plan.Jobs) != 1 || unchanged.Plan.Jobs[0].State != domain.JobReused {
		t.Fatalf("unchanged dependency footprint did not reuse coverage: %+v", unchanged.Plan.Jobs)
	}

	input.RunID = "dependency-changed"
	input.Files[0].ContentDigest = "contract-v2"
	changed, err := Plan(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(changed.Plan.Jobs) != 1 || changed.Plan.Jobs[0].State != domain.JobPending {
		t.Fatalf("changed dependency footprint reused dependent coverage: %+v", changed.Plan.Jobs)
	}
}

package state

import (
	"context"
	"errors"
	"testing"
	"time"

	"scout/internal/domain"
)

func TestCommittedJobSurvivesFreshStore(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	plan, run, job, coverage, observation := progressFixture("one")
	if err := store.SavePlan(ctx, plan); err != nil {
		t.Fatal(err)
	}
	if err := store.BeginRun(ctx, run); err != nil {
		t.Fatal(err)
	}
	artifact, err := store.StoreArtifact(ctx, []byte("sanitized scanner evidence"), "text/plain", func(raw []byte) ([]byte, error) { return raw, nil })
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CommitRunJob(ctx, run.RunID, RunJobCommit{JobID: job.JobID, Coverage: []domain.Coverage{coverage}, Observations: []domain.Observation{observation}, ArtifactRefs: []domain.SafeArtifactRef{artifact}, CommittedAt: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	fresh, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer fresh.Close()
	commits, err := fresh.ListRunJobCommits(ctx, run.RunID)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 1 || commits[0].JobID != job.JobID {
		t.Fatalf("commits = %#v", commits)
	}
	observations, err := fresh.LoadObservations(ctx, run.RunID)
	if err != nil {
		t.Fatal(err)
	}
	if len(observations) != 1 || observations[0].ObservationID != observation.ObservationID {
		t.Fatalf("observations = %#v", observations)
	}
	pending, err := fresh.LoadInProgressRun(ctx, plan.PlanID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pending.ArtifactRefs) != 1 || pending.ArtifactRefs[0] != artifact {
		t.Fatalf("committed artifact was lost after reopening: %#v", pending.ArtifactRefs)
	}
	content, err := fresh.LoadArtifact(ctx, pending.ArtifactRefs[0])
	if err != nil || string(content) != "sanitized scanner evidence" {
		t.Fatalf("committed artifact content = %q, err = %v", content, err)
	}
	if _, err := fresh.LatestRun(ctx); !errors.Is(err, ErrNotFound) {
		t.Fatalf("interrupted run became a baseline: %v", err)
	}
	finished := time.Now().UTC()
	pending.FinishedAt = &finished
	if err := fresh.FinishRun(ctx, pending, nil); err != nil {
		t.Fatal(err)
	}
	recorded, err := fresh.LatestRecordedRun(ctx)
	if err != nil || recorded.RunID != run.RunID || recorded.State != domain.RunPartial {
		t.Fatalf("latest report run = %#v, err = %v", recorded, err)
	}
	if _, err := fresh.LatestFinalizedRun(ctx); !errors.Is(err, ErrNotFound) {
		t.Fatalf("partial report run became an incremental baseline: %v", err)
	}
}

func TestOpenExistingReadOnlyDoesNotInitializeMissingState(t *testing.T) {
	root := t.TempDir()
	if _, err := OpenExistingReadOnly(root); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing state error = %v", err)
	}
	writable, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := writable.Close(); err != nil {
		t.Fatal(err)
	}
	readOnly, err := OpenExistingReadOnly(root)
	if err != nil {
		t.Fatal(err)
	}
	defer readOnly.Close()
	if _, err := readOnly.StoreArtifact(context.Background(), []byte("safe"), "text/plain", func(raw []byte) ([]byte, error) { return raw, nil }); err == nil {
		t.Fatal("read-only store wrote an artifact")
	}
}
func TestStoreRejectsUnsanitizedCredentialArtifact(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	_, err = store.StoreArtifact(context.Background(), []byte("password=raw-secret-value"), "text/plain", func(raw []byte) ([]byte, error) { return raw, nil })
	if err == nil {
		t.Fatal("raw credential artifact was persisted")
	}
}

func progressFixture(suffix string) (domain.Plan, domain.Run, domain.Job, domain.Coverage, domain.Observation) {
	key := domain.CoverageKey{FilePath: "main.go", Family: domain.FamilySecurity, Category: "security:test"}
	job := domain.Job{JobID: "job-" + suffix, Stage: domain.StageFileAnalysis, CoverageKey: &key, InputFingerprint: "input-" + suffix, State: domain.JobPending}
	snapshot := domain.Snapshot{RepositoryID: "repository-" + suffix, WorktreeDigest: "worktree-" + suffix, IgnoreRulesDigest: "ignore-" + suffix, ExclusionsDigest: "exclude-" + suffix}
	plan := domain.Plan{PlanID: "plan-" + suffix, RunID: "run-" + suffix, Snapshot: snapshot, Scope: domain.Scope{Kind: domain.ScopeRepository, SelectedFamilies: []domain.Family{domain.FamilySecurity}}, Jobs: []domain.Job{job}, CreatedAt: time.Now().UTC()}
	run := domain.Run{RunID: plan.RunID, PlanID: plan.PlanID, Snapshot: snapshot, Scope: plan.Scope, State: domain.RunPartial, StartedAt: time.Now().UTC(), ExpectedCoverage: []domain.CoverageKey{key}}
	coverage := domain.Coverage{Key: key, Status: domain.CoverageFlagged, InspectedRanges: []domain.SourceRange{{Path: "main.go", StartLine: 1, EndLine: 1}}, ObservationIDs: []string{"observation-" + suffix}}
	observation := domain.Observation{ObservationID: "observation-" + suffix, RunID: run.RunID, Source: "scanner", SourceFindingID: "finding-" + suffix, Family: domain.FamilySecurity, OriginalCategory: "security:test", RuleID: "rule", RuleVersion: "1", SourceSeverity: domain.SeverityHigh, EffectiveSeverity: domain.SeverityHigh, Confidence: domain.ConfidenceHigh, Mechanism: "unsafe path", Evidence: []domain.Evidence{{Location: domain.SourceRange{Path: "main.go", StartLine: 1, EndLine: 1}, RedactedQuote: "safe evidence"}}, Validated: true, ObservedAt: time.Now().UTC()}
	return plan, run, job, coverage, observation
}

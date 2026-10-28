package state

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"scout/internal/domain"
)

var costEpoch = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

// costPlan builds a plan of jobCount file jobs whose snapshot-bearing JSON is
// padded to roughly paddingBytes, so any per-commit plan decode is expensive.
func costPlan(jobCount, paddingBytes int) (domain.Plan, domain.Run) {
	snapshot := domain.Snapshot{RepositoryID: "repository-cost", WorktreeDigest: "worktree", IgnoreRulesDigest: "ignore", ExclusionsDigest: "exclude"}
	scope := domain.Scope{Kind: domain.ScopeRepository, SelectedFamilies: []domain.Family{domain.FamilySecurity}}
	jobs := make([]domain.Job, 0, jobCount)
	keys := make([]domain.CoverageKey, 0, jobCount)
	for index := range jobCount {
		key := domain.CoverageKey{FilePath: fmt.Sprintf("pkg/file-%05d.go", index), Family: domain.FamilySecurity, Category: "security:test"}
		keys = append(keys, key)
		jobs = append(jobs, domain.Job{JobID: fmt.Sprintf("job-%05d", index), Stage: domain.StageFileAnalysis, CoverageKey: &key, InputFingerprint: "input", State: domain.JobPending})
	}
	plan := domain.Plan{PlanID: "plan-cost", RunID: "run-cost", Snapshot: snapshot, Scope: scope, Jobs: jobs, PolicyDigest: strings.Repeat("p", paddingBytes), CreatedAt: costEpoch}
	run := domain.Run{RunID: plan.RunID, PlanID: plan.PlanID, Snapshot: snapshot, Scope: scope, State: domain.RunPartial, StartedAt: costEpoch, ExpectedCoverage: keys}
	return plan, run
}

func costCommit(job domain.Job, index int) RunJobCommit {
	coverage := domain.Coverage{Key: *job.CoverageKey, Status: domain.CoverageInspected, InspectedRanges: []domain.SourceRange{{Path: job.CoverageKey.FilePath, StartLine: 1, EndLine: 1}}}
	commit := RunJobCommit{JobID: job.JobID, Coverage: []domain.Coverage{coverage}, CommittedAt: costEpoch.Add(time.Duration(index) * time.Millisecond)}
	if index%7 == 0 {
		commit.Failures = []domain.Failure{{Stage: domain.StageFileAnalysis, Message: "flaky", JobID: job.JobID, Retryable: true}}
	}
	return commit
}

func median(values []time.Duration) time.Duration {
	sorted := append([]time.Duration(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	return sorted[len(sorted)/2]
}

// Commit cost must not depend on how many jobs the plan has or how many were
// already committed: no plan snapshot decode, no run payload rewrite.
func TestCommitRunJobCostIsFlatInJobCount(t *testing.T) {
	jobCount, padding := 5000, 5<<20
	if testing.Short() {
		jobCount, padding = 1000, 1<<20
	}
	ctx := context.Background()
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	plan, run := costPlan(jobCount, padding)
	if err := store.SavePlan(ctx, plan); err != nil {
		t.Fatal(err)
	}
	if err := store.BeginRun(ctx, run); err != nil {
		t.Fatal(err)
	}
	// Make any plan snapshot read or runs payload write in the commit path fail:
	// plans.payload becomes undecodable and runs.payload is write-protected.
	for _, statement := range []string{
		`UPDATE plans SET payload = x'00'`,
		`CREATE TRIGGER forbid_run_payload_write BEFORE UPDATE OF payload ON runs BEGIN SELECT RAISE(ABORT, 'runs payload written during commit'); END`,
	} {
		if _, err := store.db.ExecContext(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}

	bucket := jobCount / 10
	durations := make([]time.Duration, jobCount)
	var mallocs [2]uint64 // first and last bucket
	var before, after runtime.MemStats
	for index, job := range plan.Jobs {
		commit := costCommit(job, index)
		inFirst, inLast := index < bucket, index >= jobCount-bucket
		if inFirst && index == 0 || inLast && index == jobCount-bucket {
			runtime.ReadMemStats(&before)
		}
		started := time.Now()
		if err := store.CommitRunJob(ctx, run.RunID, commit); err != nil {
			t.Fatalf("commit %d: %v", index, err)
		}
		durations[index] = time.Since(started)
		if inFirst && index == bucket-1 {
			runtime.ReadMemStats(&after)
			mallocs[0] = after.Mallocs - before.Mallocs
		}
		if inLast && index == jobCount-1 {
			runtime.ReadMemStats(&after)
			mallocs[1] = after.Mallocs - before.Mallocs
		}
	}

	first, last := median(durations[:bucket]), median(durations[jobCount-bucket:])
	t.Logf("jobs=%d plan padding=%d bytes: median commit first %d = %s, last %d = %s; mallocs first=%d last=%d", jobCount, padding, bucket, first, bucket, last, mallocs[0], mallocs[1])
	// Tolerance: 3x on median time (scheduler/fsync noise), 1.25x on allocations.
	if last > 3*first {
		t.Errorf("per-commit time grew with job count: first bucket %s, last bucket %s", first, last)
	}
	if float64(mallocs[1]) > 1.25*float64(mallocs[0]) {
		t.Errorf("per-commit allocations grew with job count: first bucket %d, last bucket %d", mallocs[0], mallocs[1])
	}
}

func BenchmarkCommitRunJob(b *testing.B) {
	ctx := context.Background()
	store, err := Open(b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer store.Close()
	plan, run := costPlan(b.N, 5<<20)
	if err := store.SavePlan(ctx, plan); err != nil {
		b.Fatal(err)
	}
	if err := store.BeginRun(ctx, run); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for index, job := range plan.Jobs {
		if err := store.CommitRunJob(ctx, run.RunID, costCommit(job, index)); err != nil {
			b.Fatal(err)
		}
	}
}

func finishedRunAfter(t *testing.T, interruptAfter int) domain.Run {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	plan, run := costPlan(6, 16)
	reused := domain.Coverage{Key: domain.CoverageKey{FilePath: "reused.go", Family: domain.FamilySecurity, Category: "security:test"}, Status: domain.CoverageInspected, InspectedRanges: []domain.SourceRange{{Path: "reused.go", StartLine: 1, EndLine: 1}}}
	run.ExpectedCoverage = append(run.ExpectedCoverage, reused.Key)
	run.Coverage = []domain.Coverage{reused}

	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SavePlan(ctx, plan); err != nil {
		t.Fatal(err)
	}
	if err := store.BeginRun(ctx, run); err != nil {
		t.Fatal(err)
	}
	for index, job := range plan.Jobs[:interruptAfter] {
		if err := store.CommitRunJob(ctx, run.RunID, costCommit(job, index)); err != nil {
			t.Fatal(err)
		}
	}
	if interruptAfter < len(plan.Jobs) {
		// Interruption: the process dies; a fresh store resumes from disk.
		if err := store.Close(); err != nil {
			t.Fatal(err)
		}
		if store, err = Open(root); err != nil {
			t.Fatal(err)
		}
	}
	defer store.Close()
	resumed, err := store.LoadInProgressRun(ctx, plan.PlanID)
	if err != nil {
		t.Fatal(err)
	}
	done := map[string]bool{}
	commits, err := store.ListRunJobCommits(ctx, resumed.RunID)
	if err != nil {
		t.Fatal(err)
	}
	for _, commit := range commits {
		done[commit.JobID] = true
	}
	for index, job := range plan.Jobs {
		if done[job.JobID] {
			continue
		}
		if err := store.CommitRunJob(ctx, run.RunID, costCommit(job, index)); err != nil {
			t.Fatal(err)
		}
	}
	pending, err := store.LoadInProgressRun(ctx, plan.PlanID)
	if err != nil {
		t.Fatal(err)
	}
	finished := costEpoch.Add(time.Hour)
	pending.State = domain.RunPartial
	pending.FinishedAt = &finished
	if err := store.FinishRun(ctx, pending, nil); err != nil {
		t.Fatal(err)
	}
	final, err := store.LoadRun(ctx, run.RunID)
	if err != nil {
		t.Fatal(err)
	}
	return final
}

func TestInterruptedRunResumesToSameFinalRun(t *testing.T) {
	uninterrupted := finishedRunAfter(t, 6)
	resumed := finishedRunAfter(t, 2)
	if len(uninterrupted.Coverage) != 7 {
		t.Fatalf("coverage rows = %d, want reused + 6 jobs", len(uninterrupted.Coverage))
	}
	if !reflect.DeepEqual(uninterrupted, resumed) {
		t.Fatalf("resumed run differs from uninterrupted run:\n%#v\n%#v", resumed, uninterrupted)
	}
}

func TestDuplicateCommitRejectedAndLeavesProgressIntact(t *testing.T) {
	ctx := context.Background()
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	plan, run := costPlan(2, 16)
	if err := store.SavePlan(ctx, plan); err != nil {
		t.Fatal(err)
	}
	if err := store.BeginRun(ctx, run); err != nil {
		t.Fatal(err)
	}
	if err := store.CommitRunJob(ctx, run.RunID, costCommit(plan.Jobs[0], 0)); err != nil {
		t.Fatal(err)
	}
	if err := store.CommitRunJob(ctx, run.RunID, costCommit(plan.Jobs[0], 0)); err == nil {
		t.Fatal("duplicate commit accepted")
	}
	if err := store.CommitRunJob(ctx, run.RunID, RunJobCommit{JobID: "not-in-plan", CommittedAt: costEpoch}); err == nil {
		t.Fatal("unplanned job accepted")
	}
	pending, err := store.LoadInProgressRun(ctx, plan.PlanID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pending.Coverage) != 1 || len(pending.Failures) != 1 {
		t.Fatalf("progress after rejected commits = %d coverage, %d failures", len(pending.Coverage), len(pending.Failures))
	}
}

// A v3 store (plan snapshot only, runs.payload rewritten per commit) is
// brought to the current schema on open and resumes exactly where it stopped.
// (The schema equality of a migrated store is TestReanchoredV1StoreHasTheSchemaOfANewStore.)
func TestV3StoreMigratesAndResumes(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	plan, run := costPlan(3, 16)
	if err := store.SavePlan(ctx, plan); err != nil {
		t.Fatal(err)
	}
	if err := store.BeginRun(ctx, run); err != nil {
		t.Fatal(err)
	}
	for index, job := range plan.Jobs[:2] {
		if err := store.CommitRunJob(ctx, run.RunID, costCommit(job, index)); err != nil {
			t.Fatal(err)
		}
	}
	want, err := store.LoadInProgressRun(ctx, plan.PlanID)
	if err != nil {
		t.Fatal(err)
	}
	// Downgrade to the v3 layout: no plan_jobs, full progress in runs.payload.
	v3Payload, err := encode(want)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range []struct {
		query string
		args  []any
	}{
		{`DROP TABLE plan_jobs`, nil},
		{`DROP TABLE issue_migration_ambiguities`, nil},
		{`UPDATE runs SET payload = ? WHERE run_id = ?`, []any{v3Payload, run.RunID}},
		{`PRAGMA user_version=3`, nil},
	} {
		if _, err := store.db.ExecContext(ctx, step.query, step.args...); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	migrated, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer migrated.Close()
	got, err := migrated.LoadInProgressRun(ctx, plan.PlanID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("migrated progress differs:\n%#v\n%#v", got, want)
	}
	if err := migrated.CommitRunJob(ctx, run.RunID, costCommit(plan.Jobs[2], 2)); err != nil {
		t.Fatalf("commit after migration: %v", err)
	}
}

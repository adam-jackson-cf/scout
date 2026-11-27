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
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer closeOrFail(t, store)
	plan, run := seedGuardedCostRun(t, store, jobCount, padding)

	bucket := jobCount / 10
	durations, mallocs := measureCommits(t, store, run.RunID, plan.Jobs, bucket)

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

func seedGuardedCostRun(t *testing.T, store *Store, jobCount, padding int) (domain.Plan, domain.Run) {
	t.Helper()
	ctx := context.Background()
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
	return plan, run
}

func measureCommits(t *testing.T, store *Store, runID string, jobs []domain.Job, bucket int) ([]time.Duration, [2]uint64) {
	t.Helper()
	ctx := context.Background()
	jobCount := len(jobs)
	durations := make([]time.Duration, jobCount)
	var mallocs [2]uint64 // first and last bucket
	var before, after runtime.MemStats
	windowStarts := map[int]bool{0: true, jobCount - bucket: true}
	windowEnds := map[int]int{bucket - 1: 0, jobCount - 1: 1}
	for index, job := range jobs {
		commit := costCommit(job, index)
		if windowStarts[index] {
			runtime.ReadMemStats(&before)
		}
		started := time.Now()
		if err := store.CommitRunJob(ctx, runID, commit); err != nil {
			t.Fatalf("commit %d: %v", index, err)
		}
		durations[index] = time.Since(started)
		if slot, ok := windowEnds[index]; ok {
			runtime.ReadMemStats(&after)
			mallocs[slot] = after.Mallocs - before.Mallocs
		}
	}
	return durations, mallocs
}

func BenchmarkCommitRunJob(b *testing.B) {
	ctx := context.Background()
	store, err := Open(b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer closeOrFail(b, store)
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
	store, plan, run := interruptedCostRun(t, t.TempDir(), interruptAfter)
	defer closeOrFail(t, store)
	return completeCostRun(t, store, plan, run)
}

func interruptedCostRun(t *testing.T, root string, interruptAfter int) (*Store, domain.Plan, domain.Run) {
	t.Helper()
	ctx := context.Background()
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
	return store, plan, run
}

func completeCostRun(t *testing.T, store *Store, plan domain.Plan, run domain.Run) domain.Run {
	t.Helper()
	ctx := context.Background()
	resumed, err := store.LoadInProgressRun(ctx, plan.PlanID)
	if err != nil {
		t.Fatal(err)
	}
	commitRemainingCostJobs(t, store, plan, resumed.RunID)
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

func commitRemainingCostJobs(t *testing.T, store *Store, plan domain.Plan, runID string) {
	t.Helper()
	ctx := context.Background()
	done := map[string]bool{}
	commits, err := store.ListRunJobCommits(ctx, runID)
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
		if err := store.CommitRunJob(ctx, plan.RunID, costCommit(job, index)); err != nil {
			t.Fatal(err)
		}
	}
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
	defer closeOrFail(t, store)
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
	store, plan, run, want := partialCostStore(t, t.TempDir())
	// Build a genuine historical v3 schema, not a current store with its
	// version rewound (which would accidentally retain newly added tables).
	v3 := newV1Store(t)
	writeV3Progress(t, v3, plan, want)
	v3.close(t)
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	migrated, err := Open(v3.root)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOrFail(t, migrated)
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

func partialCostStore(t *testing.T, root string) (*Store, domain.Plan, domain.Run, domain.Run) {
	t.Helper()
	ctx := context.Background()
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
	return store, plan, run, want
}

func writeV3Progress(t *testing.T, v3 *v1Store, plan domain.Plan, want domain.Run) {
	t.Helper()
	v3.insertRun(t, plan, want, nil)
	if _, err := v3.db.Exec(`UPDATE runs SET finalized=0 WHERE run_id=?`, want.RunID); err != nil {
		t.Fatal(err)
	}
	for _, coverage := range want.Coverage {
		payload, err := encode(coverage)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := v3.db.Exec(`INSERT INTO coverage VALUES (?, ?, ?, ?)`, want.RunID, coverage.Key.String(), coverage.Status, payload); err != nil {
			t.Fatal(err)
		}
	}
	for index, job := range plan.Jobs[:2] {
		commit := costCommit(job, index)
		payload, err := encode(commit)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := v3.db.Exec(`INSERT INTO run_job_commits VALUES (?, ?, ?, ?)`, want.RunID, job.JobID, unixMillis(commit.CommittedAt), payload); err != nil {
			t.Fatal(err)
		}
	}
}

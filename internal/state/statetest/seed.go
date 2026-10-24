// Package statetest seeds a durable store for tests through the production
// BeginRun, CommitRunJob and FinishRun path.
package statetest

import (
	"context"
	"time"

	"scout/internal/domain"
	"scout/internal/state"
)

// TB is the subset of testing.TB the seeder needs.
type TB interface {
	Helper()
	Fatalf(format string, args ...any)
}

// SeedRun persists plan and run the way an assessment does: one planned job
// carries every observation and failure, the run's coverage is written when it
// begins, and the run is then finished with files as its inventory. A run
// without a finish time is finished now.
func SeedRun(tb TB, ctx context.Context, store *state.Store, plan domain.Plan, run domain.Run, observations []domain.Observation, files []domain.File) {
	tb.Helper()
	job := domain.Job{JobID: "seed", Stage: domain.StageFileAnalysis, InputFingerprint: "seed", State: domain.JobCompleted}
	plan.Jobs = append(append([]domain.Job(nil), plan.Jobs...), job)
	if err := store.SavePlan(ctx, plan); err != nil {
		tb.Fatalf("save plan: %v", err)
	}
	shell := run
	shell.State, shell.FinishedAt, shell.Failures = domain.RunPartial, nil, nil
	if err := store.BeginRun(ctx, shell); err != nil {
		tb.Fatalf("begin run: %v", err)
	}
	committedAt := run.StartedAt
	if err := store.CommitRunJob(ctx, run.RunID, state.RunJobCommit{JobID: job.JobID, Observations: observations, Failures: run.Failures, CommittedAt: committedAt}); err != nil {
		tb.Fatalf("commit run job: %v", err)
	}
	if run.FinishedAt == nil {
		finished := time.Now().UTC()
		run.FinishedAt = &finished
	}
	if files == nil {
		files = []domain.File{}
	}
	if err := store.FinishRun(ctx, run, files); err != nil {
		tb.Fatalf("finish run: %v", err)
	}
}

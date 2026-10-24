package state

import (
	"context"
	"testing"
	"time"

	"scout/internal/domain"
)

// seedRun persists plan and run through the production BeginRun, CommitRunJob
// and FinishRun path (internal/state/statetest does the same for other
// packages, which this package's tests cannot import).
func seedRun(t *testing.T, ctx context.Context, store *Store, plan domain.Plan, run domain.Run, observations []domain.Observation, files []domain.File) {
	t.Helper()
	job := domain.Job{JobID: "seed", Stage: domain.StageFileAnalysis, InputFingerprint: "seed", State: domain.JobCompleted}
	plan.Jobs = append(append([]domain.Job(nil), plan.Jobs...), job)
	if err := store.SavePlan(ctx, plan); err != nil {
		t.Fatal(err)
	}
	shell := run
	shell.State, shell.FinishedAt, shell.Failures = domain.RunPartial, nil, nil
	if err := store.BeginRun(ctx, shell); err != nil {
		t.Fatal(err)
	}
	if err := store.CommitRunJob(ctx, run.RunID, RunJobCommit{JobID: job.JobID, Observations: observations, Failures: run.Failures, CommittedAt: run.StartedAt}); err != nil {
		t.Fatal(err)
	}
	if run.FinishedAt == nil {
		finished := time.Now().UTC()
		run.FinishedAt = &finished
	}
	if files == nil {
		files = []domain.File{}
	}
	if err := store.FinishRun(ctx, run, files); err != nil {
		t.Fatal(err)
	}
}

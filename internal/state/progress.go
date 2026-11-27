package state

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"sort"
	"time"

	"scout/internal/domain"
)

// RunJobCommit is one atomically persisted job result. Coverage, failures, and
// sanitized observations are committed together, so interrupted assessments can
// resume without discarding completed work.
type RunJobCommit struct {
	JobID        string                   `json:"jobId"`
	Coverage     []domain.Coverage        `json:"coverage"`
	Observations []domain.Observation     `json:"observations"`
	Failures     []domain.Failure         `json:"failures,omitempty"`
	ArtifactRefs []domain.SafeArtifactRef `json:"artifactRefs,omitempty"`
	CommittedAt  time.Time                `json:"committedAt"`
}

func (c RunJobCommit) validate(runID string, job domain.Job) error {
	if c.JobID == "" || c.JobID != job.JobID || c.CommittedAt.IsZero() {
		return errors.New("job commit requires the planned job ID and commit time")
	}
	if err := c.validateCoverage(job); err != nil {
		return err
	}
	if err := c.validateObservations(runID); err != nil {
		return err
	}
	if err := c.validateFailures(); err != nil {
		return err
	}
	for _, ref := range c.ArtifactRefs {
		if err := ref.Validate(); err != nil {
			return fmt.Errorf("job artifact reference: %w", err)
		}
	}
	return nil
}

func (c RunJobCommit) validateCoverage(job domain.Job) error {
	if job.CoverageKey == nil {
		if len(c.Coverage) != 0 {
			return errors.New("job without a coverage key cannot commit coverage")
		}
	} else if len(c.Coverage) != 1 || c.Coverage[0].Key.String() != job.CoverageKey.String() {
		return errors.New("job commit coverage must exactly match its planned coverage key")
	}
	for _, coverage := range c.Coverage {
		if err := coverage.Validate(); err != nil {
			return fmt.Errorf("job coverage: %w", err)
		}
	}
	return nil
}

func (c RunJobCommit) validateObservations(runID string) error {
	for _, observation := range c.Observations {
		if err := observation.Validate(); err != nil {
			return fmt.Errorf("job observation %q: %w", observation.ObservationID, err)
		}
		if observation.RunID != runID {
			return fmt.Errorf("job observation %q belongs to run %q, not %q", observation.ObservationID, observation.RunID, runID)
		}
	}
	return nil
}

func (c RunJobCommit) validateFailures() error {
	for _, failure := range c.Failures {
		if err := failure.Validate(); err != nil {
			return fmt.Errorf("job failure: %w", err)
		}
		if failure.JobID != "" && failure.JobID != c.JobID {
			return errors.New("job failure references a different job")
		}
	}
	return nil
}

// BeginRun persists a partial run shell and any already-reused coverage after
// its immutable plan exists. It never finalizes the run or publishes inventory.
func (s *Store) BeginRun(ctx context.Context, run domain.Run) error {
	if err := run.Validate(); err != nil {
		return fmt.Errorf("run: %w", err)
	}
	if run.State != domain.RunPartial || run.FinishedAt != nil {
		return errors.New("begin run requires an unfinished partial run")
	}
	payload, err := shellPayload(run)
	if err != nil {
		return err
	}
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		if err := checkRunPlanTx(ctx, tx, run); err != nil {
			return err
		}
		if err := s.ensureRepositoryIdentity(ctx, tx, run.Snapshot.RepositoryID); err != nil {
			return err
		}
		if err := ensureRunArtifactsTx(ctx, tx, run.ArtifactRefs); err != nil {
			return err
		}
		if err := insertRunShellTx(ctx, tx, run, payload); err != nil {
			return err
		}
		return insertCoverageTx(ctx, tx, run.RunID, run.Coverage, "initial")
	})
}

func checkRunPlanTx(ctx context.Context, tx *sql.Tx, run domain.Run) error {
	plan, err := loadPlanTx(ctx, tx, run.PlanID)
	if err != nil {
		return err
	}
	if plan.RunID != run.RunID || plan.Snapshot.RepositoryID != run.Snapshot.RepositoryID {
		return errors.New("run does not match its immutable plan identity")
	}
	return nil
}

func ensureRunArtifactsTx(ctx context.Context, tx *sql.Tx, refs []domain.SafeArtifactRef) error {
	for _, artifact := range refs {
		if err := ensureArtifactTx(ctx, tx, artifact); err != nil {
			return fmt.Errorf("run artifact: %w", err)
		}
	}
	return nil
}

func insertRunShellTx(ctx context.Context, tx *sql.Tx, run domain.Run, payload []byte) error {
	if _, err := tx.ExecContext(ctx, `INSERT INTO runs(run_id, plan_id, repository_id, state, started_at, finished_at, payload, finalized) VALUES (?, ?, ?, ?, ?, NULL, ?, 0)`, run.RunID, run.PlanID, run.Snapshot.RepositoryID, run.State, unixMillis(run.StartedAt), payload); err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("run %q: %w", run.RunID, ErrImmutable)
		}
		return fmt.Errorf("begin run: %w", err)
	}
	return nil
}

func insertCoverageTx(ctx context.Context, tx *sql.Tx, runID string, coverage []domain.Coverage, what string) error {
	for _, row := range coverage {
		payload, err := encode(row)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO coverage(run_id, coverage_key, status, payload) VALUES (?, ?, ?, ?)`, runID, row.Key.String(), row.Status, payload); err != nil {
			return fmt.Errorf("save %s coverage %q: %w", what, row.Key.String(), err)
		}
	}
	return nil
}

// CommitRunJob atomically writes one completed job's coverage, sanitized
// observations, and failures. Recommitting a job is rejected to preserve the
// first durable result after an interruption.
func (s *Store) CommitRunJob(ctx context.Context, runID string, commit RunJobCommit) error {
	if runID == "" {
		return errors.New("run ID is required")
	}
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		planID, err := openRunPlanIDTx(ctx, tx, runID)
		if err != nil {
			return err
		}
		job, err := plannedJobTx(ctx, tx, planID, runID, commit.JobID)
		if err != nil {
			return err
		}
		if err := commit.validate(runID, job); err != nil {
			return err
		}
		if err := ensureJobNotCommittedTx(ctx, tx, runID, commit.JobID); err != nil {
			return err
		}
		if err := ensureCommitArtifactsTx(ctx, tx, commit); err != nil {
			return err
		}
		if err := insertCoverageTx(ctx, tx, runID, commit.Coverage, "job"); err != nil {
			return err
		}
		if err := insertObservationsTx(ctx, tx, runID, commit.Observations); err != nil {
			return err
		}
		return insertJobCommitTx(ctx, tx, runID, commit)
	})
}

func insertJobCommitTx(ctx context.Context, tx *sql.Tx, runID string, commit RunJobCommit) error {
	commitPayload, err := encode(commit)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO run_job_commits(run_id, job_id, committed_at, payload) VALUES (?, ?, ?, ?)`, runID, commit.JobID, unixMillis(commit.CommittedAt), commitPayload); err != nil {
		return fmt.Errorf("save job commit: %w", err)
	}
	return nil
}

func openRunPlanIDTx(ctx context.Context, tx *sql.Tx, runID string) (string, error) {
	var planID string
	var finalized bool
	if err := tx.QueryRowContext(ctx, `SELECT plan_id, finalized FROM runs WHERE run_id = ?`, runID).Scan(&planID, &finalized); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("run %q: %w", runID, ErrNotFound)
		}
		return "", fmt.Errorf("load run: %w", err)
	}
	if finalized {
		return "", errors.New("cannot commit a job to a finalized run")
	}
	return planID, nil
}

func plannedJobTx(ctx context.Context, tx *sql.Tx, planID, runID, jobID string) (domain.Job, error) {
	var planRunID string
	var jobPayload []byte
	err := tx.QueryRowContext(ctx, `SELECT run_id, payload FROM plan_jobs WHERE plan_id = ? AND job_id = ?`, planID, jobID).Scan(&planRunID, &jobPayload)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Job{}, fmt.Errorf("run job %q is not in plan %q", jobID, planID)
	}
	if err != nil {
		return domain.Job{}, fmt.Errorf("load plan job: %w", err)
	}
	if planRunID != runID {
		return domain.Job{}, fmt.Errorf("plan %q does not belong to run %q", planID, runID)
	}
	var job domain.Job
	if err := decode(jobPayload, &job); err != nil {
		return domain.Job{}, err
	}
	return job, nil
}

func ensureJobNotCommittedTx(ctx context.Context, tx *sql.Tx, runID, jobID string) error {
	var existing string
	err := tx.QueryRowContext(ctx, `SELECT job_id FROM run_job_commits WHERE run_id = ? AND job_id = ?`, runID, jobID).Scan(&existing)
	if err == nil {
		return fmt.Errorf("run job %q: %w", jobID, ErrImmutable)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("load committed job: %w", err)
	}
	return nil
}

func ensureCommitArtifactsTx(ctx context.Context, tx *sql.Tx, commit RunJobCommit) error {
	for _, observation := range commit.Observations {
		if err := ensureObservationArtifactsTx(ctx, tx, observation); err != nil {
			return err
		}
	}
	for _, ref := range commit.ArtifactRefs {
		if err := ensureArtifactTx(ctx, tx, ref); err != nil {
			return fmt.Errorf("job artifact: %w", err)
		}
	}
	return nil
}

func ensureObservationArtifactsTx(ctx context.Context, tx *sql.Tx, observation domain.Observation) error {
	for _, artifact := range observation.ArtifactRefs {
		if err := ensureArtifactTx(ctx, tx, artifact); err != nil {
			return fmt.Errorf("job observation %q artifact: %w", observation.ObservationID, err)
		}
	}
	for _, evidence := range observation.Evidence {
		if evidence.Artifact != nil {
			if err := ensureArtifactTx(ctx, tx, *evidence.Artifact); err != nil {
				return fmt.Errorf("job observation %q evidence artifact: %w", observation.ObservationID, err)
			}
		}
	}
	return nil
}

func insertObservationsTx(ctx context.Context, tx *sql.Tx, runID string, observations []domain.Observation) error {
	for _, observation := range observations {
		payload, err := encode(observation)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO observations(observation_id, run_id, observed_at, payload) VALUES (?, ?, ?, ?)`, observation.ObservationID, runID, unixMillis(observation.ObservedAt), payload); err != nil {
			if isUniqueViolation(err) {
				return fmt.Errorf("job observation %q: %w", observation.ObservationID, ErrImmutable)
			}
			return fmt.Errorf("save job observation %q: %w", observation.ObservationID, err)
		}
	}
	return nil
}

// FinishRun atomically publishes the caller's final run summary and inventory.
// All coverage must have been persisted by BeginRun or CommitRunJob first.
func (s *Store) FinishRun(ctx context.Context, run domain.Run, files []domain.File) error {
	if err := validateFinishedRun(run, files); err != nil {
		return err
	}
	payload, err := encode(run)
	if err != nil {
		return err
	}
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		if err := checkFinalRunTx(ctx, tx, run); err != nil {
			return err
		}
		if err := s.ensureRepositoryIdentity(ctx, tx, run.Snapshot.RepositoryID); err != nil {
			return err
		}
		if err := ensureRunArtifactsTx(ctx, tx, run.ArtifactRefs); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE runs SET state = ?, finished_at = ?, payload = ?, finalized = 1 WHERE run_id = ? AND finalized = 0`, run.State, unixMillis(*run.FinishedAt), payload, run.RunID); err != nil {
			return fmt.Errorf("finish run: %w", err)
		}
		indexPayload, err := encode(files)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO inventory_snapshots(run_id, payload) VALUES (?, ?)`, run.RunID, indexPayload); err != nil {
			return fmt.Errorf("save inventory snapshot: %w", err)
		}
		return nil
	})
}

func validateFinishedRun(run domain.Run, files []domain.File) error {
	if err := run.Validate(); err != nil {
		return fmt.Errorf("run: %w", err)
	}
	if run.FinishedAt == nil {
		return errors.New("finished run requires a finish time")
	}
	for _, file := range files {
		if err := file.Validate(); err != nil {
			return fmt.Errorf("inventory file: %w", err)
		}
	}
	return nil
}

func checkFinalRunTx(ctx context.Context, tx *sql.Tx, run domain.Run) error {
	stored, finalized, err := loadRunRecord(ctx, tx, run.RunID)
	if err != nil {
		return err
	}
	if finalized {
		return fmt.Errorf("run %q: %w", run.RunID, ErrImmutable)
	}
	if stored.PlanID != run.PlanID || stored.Snapshot.RepositoryID != run.Snapshot.RepositoryID {
		return errors.New("final run does not match persisted run identity")
	}
	coverage := stored.Coverage
	if same, err := sameRows(coverage, run.Coverage); err != nil {
		return err
	} else if !same {
		return errors.New("final run coverage does not match atomically committed jobs")
	}
	if same, err := sameRows(stored.Failures, run.Failures); err != nil {
		return err
	} else if !same {
		return errors.New("final run failures do not match atomically committed jobs")
	}
	return nil
}

// LoadInProgressRun returns the exact unfinished run created for a plan. A
// plan without such a row has not started; callers may then BeginRun it.
func (s *Store) LoadInProgressRun(ctx context.Context, planID string) (domain.Run, error) {
	if planID == "" {
		return domain.Run{}, errors.New("plan ID is required")
	}
	var run domain.Run
	err := s.withRead(ctx, func(db *sql.DB) error {
		var runID string
		err := db.QueryRowContext(ctx, `SELECT run_id FROM runs WHERE plan_id = ? AND finalized = 0 ORDER BY started_at DESC, run_id DESC LIMIT 1`, planID).Scan(&runID)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("load in-progress run: %w", err)
		}
		run, _, err = loadRunRecord(ctx, db, runID)
		if err != nil {
			return err
		}
		if run.PlanID != planID {
			return errors.New("stored in-progress run plan identity does not match")
		}
		if run.State != domain.RunPartial {
			return errors.New("in-progress run is not partial")
		}
		return nil
	})
	return run, err
}

// ListRunJobCommits exposes completed work for a resumed assessment in stable order.
func (s *Store) ListRunJobCommits(ctx context.Context, runID string) ([]RunJobCommit, error) {
	if runID == "" {
		return nil, errors.New("run ID is required")
	}
	commits := []RunJobCommit{}
	err := s.withRead(ctx, func(db *sql.DB) error {
		rows, err := db.QueryContext(ctx, `SELECT payload FROM run_job_commits WHERE run_id = ? ORDER BY committed_at, job_id`, runID)
		if err != nil {
			return fmt.Errorf("query run job commits: %w", err)
		}
		return scanRows(rows, func(rows *sql.Rows) error {
			var payload []byte
			if err := rows.Scan(&payload); err != nil {
				return fmt.Errorf("scan run job commit: %w", err)
			}
			var commit RunJobCommit
			if err := decode(payload, &commit); err != nil {
				return err
			}
			commits = append(commits, commit)
			return nil
		})
	})
	return commits, err
}

func loadPlanTx(ctx context.Context, tx *sql.Tx, planID string) (domain.Plan, error) {
	var payload []byte
	if err := tx.QueryRowContext(ctx, `SELECT payload FROM plans WHERE plan_id = ?`, planID).Scan(&payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Plan{}, fmt.Errorf("plan %q: %w", planID, ErrNotFound)
		}
		return domain.Plan{}, fmt.Errorf("load plan: %w", err)
	}
	var plan domain.Plan
	if err := decode(payload, &plan); err != nil {
		return domain.Plan{}, err
	}
	if err := plan.Validate(); err != nil {
		return domain.Plan{}, fmt.Errorf("stored plan is invalid: %w", err)
	}
	return plan, nil
}

type rowQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// commitDelta is the part of a job commit that contributes to run-level
// progress; decoding only these fields skips the (large) observations.
type commitDelta struct {
	Failures     []domain.Failure         `json:"failures,omitempty"`
	ArtifactRefs []domain.SafeArtifactRef `json:"artifactRefs,omitempty"`
}

// savePlanJobsTx writes the per-job rows that let CommitRunJob validate a job
// without decoding the plan snapshot.
func savePlanJobsTx(ctx context.Context, tx *sql.Tx, plan domain.Plan) error {
	for _, job := range plan.Jobs {
		payload, err := encode(job)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO plan_jobs(plan_id, run_id, job_id, payload) VALUES (?, ?, ?, ?)`, plan.PlanID, plan.RunID, job.JobID, payload); err != nil {
			return fmt.Errorf("save plan job %q: %w", job.JobID, err)
		}
	}
	return nil
}

// loadRunRecord restores a run. A finalized run is its stored payload; an
// unfinished run is its shell payload plus coverage, failures, and artifact
// references reassembled from the coverage and run_job_commits tables, which
// are the only per-commit writes.
func loadRunRecord(ctx context.Context, q rowQuerier, runID string) (domain.Run, bool, error) {
	var payload []byte
	var finalized bool
	if err := q.QueryRowContext(ctx, `SELECT payload, finalized FROM runs WHERE run_id = ?`, runID).Scan(&payload, &finalized); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Run{}, false, fmt.Errorf("run %q: %w", runID, ErrNotFound)
		}
		return domain.Run{}, false, fmt.Errorf("load run: %w", err)
	}
	var run domain.Run
	if err := decode(payload, &run); err != nil {
		return domain.Run{}, false, err
	}
	if !finalized {
		coverage, err := loadRunCoverage(ctx, q, runID)
		if err != nil {
			return domain.Run{}, false, err
		}
		run.Coverage = coverage
		deltas, err := loadCommitDeltas(ctx, q, runID)
		if err != nil {
			return domain.Run{}, false, err
		}
		for _, delta := range deltas {
			run.Failures = append(run.Failures, delta.Failures...)
			run.ArtifactRefs = append(run.ArtifactRefs, delta.ArtifactRefs...)
		}
	}
	if err := run.Validate(); err != nil {
		return domain.Run{}, false, fmt.Errorf("stored run is invalid: %w", err)
	}
	return run, finalized, nil
}

func loadCommitDeltas(ctx context.Context, q rowQuerier, runID string) ([]commitDelta, error) {
	rows, err := q.QueryContext(ctx, `SELECT payload FROM run_job_commits WHERE run_id = ? ORDER BY rowid`, runID)
	if err != nil {
		return nil, fmt.Errorf("query run job commits: %w", err)
	}
	var deltas []commitDelta
	err = scanRows(rows, func(rows *sql.Rows) error {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return fmt.Errorf("scan run job commit: %w", err)
		}
		var delta commitDelta
		if err := decode(payload, &delta); err != nil {
			return err
		}
		deltas = append(deltas, delta)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return deltas, nil
}

func loadRunCoverage(ctx context.Context, q rowQuerier, runID string) ([]domain.Coverage, error) {
	rows, err := q.QueryContext(ctx, `SELECT payload FROM coverage WHERE run_id = ? ORDER BY rowid`, runID)
	if err != nil {
		return nil, fmt.Errorf("query run coverage: %w", err)
	}
	coverage := []domain.Coverage{}
	err = scanRows(rows, func(rows *sql.Rows) error {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return fmt.Errorf("scan run coverage: %w", err)
		}
		var row domain.Coverage
		if err := decode(payload, &row); err != nil {
			return err
		}
		coverage = append(coverage, row)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return coverage, nil
}

// shellPayload is the unfinished run's stored base: progress lives in the
// coverage and run_job_commits tables, not here.
func shellPayload(run domain.Run) ([]byte, error) {
	run.Coverage = nil
	return encode(run)
}

// schemaPlanJobs is the per-plan job table. A new store creates it directly;
// a re-anchored v1 store creates it in backfillPlanJobs.
const schemaPlanJobs = `CREATE TABLE plan_jobs (plan_id TEXT NOT NULL, run_id TEXT NOT NULL, job_id TEXT NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(plan_id, job_id), FOREIGN KEY(plan_id) REFERENCES plans(plan_id))`

// backfillPlanJobs adds plan_jobs, backfills it for existing plans, and rewrites
// unfinished run payloads into the shell form whose progress lives in the
// coverage and run_job_commits tables. It serves the v1 re-anchor path only.
func backfillPlanJobs(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, schemaPlanJobs); err != nil {
		return err
	}
	plans, err := storedPlansTx(ctx, tx)
	if err != nil {
		return err
	}
	for _, plan := range plans {
		if err := savePlanJobsTx(ctx, tx, plan); err != nil {
			return err
		}
	}

	runs, err := unfinishedRunsTx(ctx, tx)
	if err != nil {
		return err
	}
	for _, item := range runs {
		if err := rewriteRunShellTx(ctx, tx, item); err != nil {
			return err
		}
	}
	return nil
}

func storedPlansTx(ctx context.Context, tx *sql.Tx) ([]domain.Plan, error) {
	rows, err := tx.QueryContext(ctx, `SELECT payload FROM plans`)
	if err != nil {
		return nil, err
	}
	var plans []domain.Plan
	err = scanRows(rows, func(rows *sql.Rows) error {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return err
		}
		var plan domain.Plan
		if err := decode(payload, &plan); err != nil {
			return err
		}
		plans = append(plans, plan)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return plans, nil
}

type pendingRun struct {
	runID   string
	payload []byte
}

func unfinishedRunsTx(ctx context.Context, tx *sql.Tx) ([]pendingRun, error) {
	runRows, err := tx.QueryContext(ctx, `SELECT run_id, payload FROM runs WHERE finalized = 0`)
	if err != nil {
		return nil, err
	}
	var runs []pendingRun
	err = scanRows(runRows, func(runRows *sql.Rows) error {
		var item pendingRun
		if err := runRows.Scan(&item.runID, &item.payload); err != nil {
			return err
		}
		runs = append(runs, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return runs, nil
}

func rewriteRunShellTx(ctx context.Context, tx *sql.Tx, item pendingRun) error {
	var run domain.Run
	if err := decode(item.payload, &run); err != nil {
		return err
	}
	deltas, err := loadCommitDeltas(ctx, tx, item.runID)
	if err != nil {
		return err
	}
	// v3 appended every commit's failures and artifact refs after the
	// shell's own; peel them off to recover the shell.
	failures, refs := 0, 0
	for _, delta := range deltas {
		failures += len(delta.Failures)
		refs += len(delta.ArtifactRefs)
	}
	if failures > len(run.Failures) || refs > len(run.ArtifactRefs) {
		return fmt.Errorf("run %q progress does not match its commits", item.runID)
	}
	run.Failures = run.Failures[:len(run.Failures)-failures]
	run.ArtifactRefs = run.ArtifactRefs[:len(run.ArtifactRefs)-refs]
	payload, err := shellPayload(run)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE runs SET payload = ? WHERE run_id = ?`, payload, item.runID); err != nil {
		return err
	}
	return nil
}

// sameRows reports whether left and right hold the same encoded rows,
// regardless of order. An unencodable row is an error, not a mismatch.
func sameRows[T any](left, right []T) (bool, error) {
	if len(left) != len(right) {
		return false, nil
	}
	encoded := func(rows []T) ([]string, error) {
		values := make([]string, 0, len(rows))
		for _, row := range rows {
			payload, err := encode(row)
			if err != nil {
				return nil, err
			}
			values = append(values, string(payload))
		}
		sort.Strings(values)
		return values, nil
	}
	first, err := encoded(left)
	if err != nil {
		return false, err
	}
	second, err := encoded(right)
	if err != nil {
		return false, err
	}
	return slices.Equal(first, second), nil
}

package state

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	for _, observation := range c.Observations {
		if err := observation.Validate(); err != nil {
			return fmt.Errorf("job observation %q: %w", observation.ObservationID, err)
		}
		if observation.RunID != runID {
			return fmt.Errorf("job observation %q belongs to run %q, not %q", observation.ObservationID, observation.RunID, runID)
		}
	}
	for _, failure := range c.Failures {
		if err := failure.Validate(); err != nil {
			return fmt.Errorf("job failure: %w", err)
		}
		if failure.JobID != "" && failure.JobID != c.JobID {
			return errors.New("job failure references a different job")
		}
	}
	for _, ref := range c.ArtifactRefs {
		if err := ref.Validate(); err != nil {
			return fmt.Errorf("job artifact reference: %w", err)
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
	payload, err := encode(run)
	if err != nil {
		return err
	}
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		plan, err := loadPlanTx(ctx, tx, run.PlanID)
		if err != nil {
			return err
		}
		if plan.RunID != run.RunID || plan.Snapshot.RepositoryID != run.Snapshot.RepositoryID {
			return errors.New("run does not match its immutable plan identity")
		}
		if err := s.ensureRepositoryIdentity(ctx, tx, run.Snapshot.RepositoryID); err != nil {
			return err
		}
		for _, artifact := range run.ArtifactRefs {
			if err := ensureArtifactTx(ctx, tx, artifact); err != nil {
				return fmt.Errorf("run artifact: %w", err)
			}
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO runs(run_id, plan_id, repository_id, state, started_at, finished_at, payload, finalized) VALUES (?, ?, ?, ?, ?, NULL, ?, 0)`, run.RunID, run.PlanID, run.Snapshot.RepositoryID, run.State, unixMillis(run.StartedAt), payload); err != nil {
			if isUniqueViolation(err) {
				return fmt.Errorf("run %q: %w", run.RunID, ErrImmutable)
			}
			return fmt.Errorf("begin run: %w", err)
		}
		for _, coverage := range run.Coverage {
			coveragePayload, err := encode(coverage)
			if err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO coverage(run_id, coverage_key, status, payload) VALUES (?, ?, ?, ?)`, run.RunID, coverage.Key.String(), coverage.Status, coveragePayload); err != nil {
				return fmt.Errorf("save initial coverage %q: %w", coverage.Key.String(), err)
			}
		}
		return nil
	})
}

// CommitRunJob atomically writes one completed job's coverage, sanitized
// observations, and failures. Recommitting a job is rejected to preserve the
// first durable result after an interruption.
func (s *Store) CommitRunJob(ctx context.Context, runID string, commit RunJobCommit) error {
	if runID == "" {
		return errors.New("run ID is required")
	}
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		run, finalized, err := loadRunProgressTx(ctx, tx, runID)
		if err != nil {
			return err
		}
		if finalized {
			return errors.New("cannot commit a job to a finalized run")
		}
		plan, err := loadPlanTx(ctx, tx, run.PlanID)
		if err != nil {
			return err
		}
		job, found := plannedJob(plan, commit.JobID)
		if !found {
			return fmt.Errorf("run job %q is not in plan %q", commit.JobID, plan.PlanID)
		}
		if err := commit.validate(runID, job); err != nil {
			return err
		}
		var existing string
		err = tx.QueryRowContext(ctx, `SELECT job_id FROM run_job_commits WHERE run_id = ? AND job_id = ?`, runID, commit.JobID).Scan(&existing)
		if err == nil {
			return fmt.Errorf("run job %q: %w", commit.JobID, ErrImmutable)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("load committed job: %w", err)
		}
		for _, observation := range commit.Observations {
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
		}
		for _, ref := range commit.ArtifactRefs {
			if err := ensureArtifactTx(ctx, tx, ref); err != nil {
				return fmt.Errorf("job artifact: %w", err)
			}
		}
		for _, coverage := range commit.Coverage {
			payload, err := encode(coverage)
			if err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO coverage(run_id, coverage_key, status, payload) VALUES (?, ?, ?, ?)`, runID, coverage.Key.String(), coverage.Status, payload); err != nil {
				return fmt.Errorf("save job coverage %q: %w", coverage.Key.String(), err)
			}
		}
		for _, observation := range commit.Observations {
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
		commitPayload, err := encode(commit)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO run_job_commits(run_id, job_id, committed_at, payload) VALUES (?, ?, ?, ?)`, runID, commit.JobID, unixMillis(commit.CommittedAt), commitPayload); err != nil {
			return fmt.Errorf("save job commit: %w", err)
		}
		run.Coverage = append(run.Coverage, commit.Coverage...)
		run.Failures = append(run.Failures, commit.Failures...)
		run.ArtifactRefs = append(run.ArtifactRefs, commit.ArtifactRefs...)
		runPayload, err := encode(run)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE runs SET payload = ? WHERE run_id = ? AND finalized = 0`, runPayload, runID); err != nil {
			return fmt.Errorf("refresh run progress: %w", err)
		}
		return nil
	})
}

// FinishRun atomically publishes the caller's final run summary and inventory.
// All coverage must have been persisted by BeginRun or CommitRunJob first.
func (s *Store) FinishRun(ctx context.Context, run domain.Run, files []domain.File) error {
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
	payload, err := encode(run)
	if err != nil {
		return err
	}
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		stored, finalized, err := loadRunProgressTx(ctx, tx, run.RunID)
		if err != nil {
			return err
		}
		if finalized {
			return fmt.Errorf("run %q: %w", run.RunID, ErrImmutable)
		}
		if stored.PlanID != run.PlanID || stored.Snapshot.RepositoryID != run.Snapshot.RepositoryID {
			return errors.New("final run does not match persisted run identity")
		}
		coverage, err := loadRunCoverageTx(ctx, tx, run.RunID)
		if err != nil {
			return err
		}
		if !sameCoverage(coverage, run.Coverage) {
			return errors.New("final run coverage does not match atomically committed jobs")
		}
		if !sameFailures(stored.Failures, run.Failures) {
			return errors.New("final run failures do not match atomically committed jobs")
		}
		if err := s.ensureRepositoryIdentity(ctx, tx, run.Snapshot.RepositoryID); err != nil {
			return err
		}
		for _, artifact := range run.ArtifactRefs {
			if err := ensureArtifactTx(ctx, tx, artifact); err != nil {
				return fmt.Errorf("run artifact: %w", err)
			}
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

// LoadInProgressRun returns the exact unfinished run created for a plan. A
// plan without such a row has not started; callers may then BeginRun it.
func (s *Store) LoadInProgressRun(ctx context.Context, planID string) (domain.Run, error) {
	if planID == "" {
		return domain.Run{}, errors.New("plan ID is required")
	}
	var run domain.Run
	err := s.withRead(ctx, func(db *sql.DB) error {
		var payload []byte
		err := db.QueryRowContext(ctx, `SELECT payload FROM runs WHERE plan_id = ? AND finalized = 0 ORDER BY started_at DESC, run_id DESC LIMIT 1`, planID).Scan(&payload)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("load in-progress run: %w", err)
		}
		if err := decode(payload, &run); err != nil {
			return err
		}
		if err := run.Validate(); err != nil {
			return fmt.Errorf("stored in-progress run is invalid: %w", err)
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
		defer rows.Close()
		for rows.Next() {
			var payload []byte
			if err := rows.Scan(&payload); err != nil {
				return fmt.Errorf("scan run job commit: %w", err)
			}
			var commit RunJobCommit
			if err := decode(payload, &commit); err != nil {
				return err
			}
			commits = append(commits, commit)
		}
		return rows.Err()
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

func loadRunProgressTx(ctx context.Context, tx *sql.Tx, runID string) (domain.Run, bool, error) {
	var payload []byte
	var finalized bool
	if err := tx.QueryRowContext(ctx, `SELECT payload, finalized FROM runs WHERE run_id = ?`, runID).Scan(&payload, &finalized); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Run{}, false, fmt.Errorf("run %q: %w", runID, ErrNotFound)
		}
		return domain.Run{}, false, fmt.Errorf("load run: %w", err)
	}
	var run domain.Run
	if err := decode(payload, &run); err != nil {
		return domain.Run{}, false, err
	}
	if err := run.Validate(); err != nil {
		return domain.Run{}, false, fmt.Errorf("stored run is invalid: %w", err)
	}
	return run, finalized, nil
}

func loadRunCoverageTx(ctx context.Context, tx *sql.Tx, runID string) ([]domain.Coverage, error) {
	rows, err := tx.QueryContext(ctx, `SELECT payload FROM coverage WHERE run_id = ? ORDER BY coverage_key`, runID)
	if err != nil {
		return nil, fmt.Errorf("query run coverage: %w", err)
	}
	defer rows.Close()
	coverage := []domain.Coverage{}
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, fmt.Errorf("scan run coverage: %w", err)
		}
		var row domain.Coverage
		if err := decode(payload, &row); err != nil {
			return nil, err
		}
		coverage = append(coverage, row)
	}
	return coverage, rows.Err()
}

func plannedJob(plan domain.Plan, jobID string) (domain.Job, bool) {
	for _, job := range plan.Jobs {
		if job.JobID == jobID {
			return job, true
		}
	}
	return domain.Job{}, false
}

func sameCoverage(left, right []domain.Coverage) bool {
	if len(left) != len(right) {
		return false
	}
	encoded := func(rows []domain.Coverage) ([]string, bool) {
		values := make([]string, 0, len(rows))
		for _, row := range rows {
			payload, err := encode(row)
			if err != nil {
				return nil, false
			}
			values = append(values, string(payload))
		}
		sort.Strings(values)
		return values, true
	}
	first, ok := encoded(left)
	if !ok {
		return false
	}
	second, ok := encoded(right)
	if !ok {
		return false
	}
	for index := range first {
		if first[index] != second[index] {
			return false
		}
	}
	return true
}

func sameFailures(left, right []domain.Failure) bool {
	if len(left) != len(right) {
		return false
	}
	encoded := func(rows []domain.Failure) ([]string, bool) {
		values := make([]string, 0, len(rows))
		for _, row := range rows {
			payload, err := encode(row)
			if err != nil {
				return nil, false
			}
			values = append(values, string(payload))
		}
		sort.Strings(values)
		return values, true
	}
	first, ok := encoded(left)
	if !ok {
		return false
	}
	second, ok := encoded(right)
	if !ok {
		return false
	}
	for index := range first {
		if first[index] != second[index] {
			return false
		}
	}
	return true
}

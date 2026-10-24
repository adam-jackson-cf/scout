package state

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"scout/internal/domain"
)

// SavePlan persists an immutable execution plan before its jobs execute.
func (s *Store) SavePlan(ctx context.Context, plan domain.Plan) error {
	if err := plan.Validate(); err != nil {
		return fmt.Errorf("plan: %w", err)
	}
	payload, err := encode(plan)
	if err != nil {
		return err
	}
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		if err := s.ensureRepositoryIdentity(ctx, tx, plan.Snapshot.RepositoryID); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO plans(plan_id, run_id, repository_id, created_at, payload) VALUES (?, ?, ?, ?, ?)`, plan.PlanID, plan.RunID, plan.Snapshot.RepositoryID, unixMillis(plan.CreatedAt), payload)
		if err != nil {
			if isUniqueViolation(err) {
				return fmt.Errorf("plan %q: %w", plan.PlanID, ErrImmutable)
			}
			return fmt.Errorf("save plan: %w", err)
		}
		return savePlanJobsTx(ctx, tx, plan)
	})
}

// LoadPlan restores the immutable execution plan for a run.
func (s *Store) LoadPlan(ctx context.Context, planID string) (domain.Plan, error) {
	if planID == "" {
		return domain.Plan{}, errors.New("plan ID is required")
	}
	var plan domain.Plan
	err := s.withRead(ctx, func(db *sql.DB) error {
		var payload []byte
		if err := db.QueryRowContext(ctx, `SELECT payload FROM plans WHERE plan_id = ?`, planID).Scan(&payload); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("plan %q: %w", planID, ErrNotFound)
			}
			return fmt.Errorf("load plan: %w", err)
		}
		if err := decode(payload, &plan); err != nil {
			return err
		}
		if err := plan.Validate(); err != nil {
			return fmt.Errorf("stored plan is invalid: %w", err)
		}
		return nil
	})
	return plan, err
}

// ListResumablePlans returns plans without a completed run so orchestration can
// resume unfinished units after a process interruption.
func (s *Store) ListResumablePlans(ctx context.Context) ([]domain.Plan, error) {
	plans := []domain.Plan{}
	err := s.withRead(ctx, func(db *sql.DB) error {
		rows, err := db.QueryContext(ctx, `SELECT p.payload
			FROM plans p LEFT JOIN runs r ON r.run_id = p.run_id
			WHERE r.run_id IS NULL OR r.finalized = 0
			ORDER BY p.created_at, p.plan_id`)
		if err != nil {
			return fmt.Errorf("query resumable plans: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var payload []byte
			if err := rows.Scan(&payload); err != nil {
				return fmt.Errorf("scan resumable plan: %w", err)
			}
			var plan domain.Plan
			if err := decode(payload, &plan); err != nil {
				return err
			}
			if err := plan.Validate(); err != nil {
				return fmt.Errorf("stored resumable plan is invalid: %w", err)
			}
			plans = append(plans, plan)
		}
		return rows.Err()
	})
	return plans, err
}

// LoadRun restores a persisted run without mutating its recorded coverage state.
func (s *Store) LoadRun(ctx context.Context, runID string) (domain.Run, error) {
	if runID == "" {
		return domain.Run{}, errors.New("run ID is required")
	}
	var run domain.Run
	err := s.withRead(ctx, func(db *sql.DB) error {
		var err error
		run, _, err = loadRunRecord(ctx, db, runID)
		return err
	})
	return run, err
}

// LoadObservations restores immutable observations from one run.
func (s *Store) LoadObservations(ctx context.Context, runID string) ([]domain.Observation, error) {
	if runID == "" {
		return nil, errors.New("run ID is required")
	}
	observations := []domain.Observation{}
	err := s.withRead(ctx, func(db *sql.DB) error {
		rows, err := db.QueryContext(ctx, `SELECT payload FROM observations WHERE run_id = ? ORDER BY observed_at, observation_id`, runID)
		if err != nil {
			return fmt.Errorf("query observations: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var payload []byte
			if err := rows.Scan(&payload); err != nil {
				return fmt.Errorf("scan observation: %w", err)
			}
			var observation domain.Observation
			if err := decode(payload, &observation); err != nil {
				return err
			}
			if err := observation.Validate(); err != nil {
				return fmt.Errorf("stored observation is invalid: %w", err)
			}
			if observation.RunID != runID {
				return errors.New("stored observation run ID does not match query")
			}
			observations = append(observations, observation)
		}
		return rows.Err()
	})
	return observations, err
}

// CreateIssue persists a new durable issue and its initial audit history.
func (s *Store) CreateIssue(ctx context.Context, issue domain.Issue) error {
	if err := issue.Validate(); err != nil {
		return fmt.Errorf("issue: %w", err)
	}
	if len(issue.History) == 0 {
		return errors.New("new issue requires initial audit history")
	}
	return s.saveIssue(ctx, issue, true)
}

// LoadIssue restores a durable issue and its audit history.
func (s *Store) LoadIssue(ctx context.Context, issueID string) (domain.Issue, error) {
	if issueID == "" {
		return domain.Issue{}, errors.New("issue ID is required")
	}
	var issue domain.Issue
	err := s.withRead(ctx, func(db *sql.DB) error {
		var payload []byte
		if err := db.QueryRowContext(ctx, `SELECT payload FROM issues WHERE issue_id = ?`, issueID).Scan(&payload); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("issue %q: %w", issueID, ErrNotFound)
			}
			return fmt.Errorf("load issue: %w", err)
		}
		if err := decode(payload, &issue); err != nil {
			return err
		}
		if err := issue.Validate(); err != nil {
			return fmt.Errorf("stored issue is invalid: %w", err)
		}
		return nil
	})
	return issue, err
}

// ListIssues returns durable issues, optionally restricted to the requested states.
func (s *Store) ListIssues(ctx context.Context, states ...domain.IssueState) ([]domain.Issue, error) {
	for _, state := range states {
		if !state.Valid() {
			return nil, fmt.Errorf("invalid issue state %q", state)
		}
	}
	issues := []domain.Issue{}
	err := s.withRead(ctx, func(db *sql.DB) error {
		query := `SELECT payload FROM issues`
		arguments := make([]any, 0, len(states))
		if len(states) != 0 {
			placeholders := make([]string, len(states))
			for index, state := range states {
				placeholders[index] = "?"
				arguments = append(arguments, state)
			}
			query += ` WHERE state IN (` + strings.Join(placeholders, ",") + `)`
		}
		query += ` ORDER BY updated_at DESC, issue_id`
		rows, err := db.QueryContext(ctx, query, arguments...)
		if err != nil {
			return fmt.Errorf("query issues: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var payload []byte
			if err := rows.Scan(&payload); err != nil {
				return fmt.Errorf("scan issue: %w", err)
			}
			var issue domain.Issue
			if err := decode(payload, &issue); err != nil {
				return err
			}
			if err := issue.Validate(); err != nil {
				return fmt.Errorf("stored issue is invalid: %w", err)
			}
			issues = append(issues, issue)
		}
		return rows.Err()
	})
	return issues, err
}

// TriageIssue records an explicit human disposition change and audit entry.
// False-positive dispositions require contradictory evidence through
// TriageIssueWithEvidence.
func (s *Store) TriageIssue(ctx context.Context, issueID string, state domain.IssueState, owner, dispositionReason, actor, reason string) (domain.Issue, error) {
	if state == domain.IssueFalsePositive {
		return domain.Issue{}, errors.New("false-positive disposition requires contradictory evidence")
	}
	return s.transitionIssue(ctx, issueID, state, actor, reason, nil, func(issue *domain.Issue) error {
		issue.Owner = owner
		issue.DispositionReason = dispositionReason
		return nil
	})
}

// TriageIssueWithEvidence records a human disposition with immutable,
// sanitized contradictory or scope evidence. False-positive disposition cannot
// be inferred from silence or a skipped scan.
func (s *Store) TriageIssueWithEvidence(ctx context.Context, issueID string, state domain.IssueState, owner, dispositionReason, actor, reason string, evidence []domain.SafeArtifactRef) (domain.Issue, error) {
	if state != domain.IssueFalsePositive && state != domain.IssueWontFix {
		return domain.Issue{}, errors.New("evidence-backed triage is only for false-positive or won't-fix dispositions")
	}
	if len(evidence) == 0 {
		return domain.Issue{}, errors.New("disposition requires contradictory or scope evidence")
	}
	return s.transitionIssue(ctx, issueID, state, actor, reason, evidence, func(issue *domain.Issue) error {
		issue.Owner = owner
		issue.DispositionReason = dispositionReason
		return nil
	})
}

func (s *Store) transitionIssue(ctx context.Context, issueID string, state domain.IssueState, actor, reason string, artifacts []domain.SafeArtifactRef, mutate func(*domain.Issue) error) (domain.Issue, error) {
	if issueID == "" || !state.Valid() || reason == "" {
		return domain.Issue{}, errors.New("issue ID, valid destination state, and audit reason are required")
	}
	var updated domain.Issue
	err := s.withTransaction(ctx, func(tx *sql.Tx) error {
		issue, err := loadIssueTx(ctx, tx, issueID)
		if err != nil {
			return err
		}
		if !allowedIssueTransition(issue.State, state) {
			return fmt.Errorf("issue transition from %q to %q is not allowed", issue.State, state)
		}
		previousState := issue.State
		if mutate != nil {
			if err := mutate(&issue); err != nil {
				return err
			}
		}
		issue.State = state
		issue.UpdatedAt = time.Now().UTC()
		issue.History = append(issue.History, domain.IssueHistoryEntry{At: issue.UpdatedAt, From: previousState, To: state, Event: domain.EventTriaged, Actor: actor, Reason: reason, ArtifactRefs: append([]domain.SafeArtifactRef(nil), artifacts...)})
		if err := issue.Validate(); err != nil {
			return fmt.Errorf("updated issue: %w", err)
		}
		if err := saveIssueTx(ctx, tx, issue, false); err != nil {
			return err
		}
		if err := recordDispositionContextTx(ctx, tx, issue.IssueID, issue.UpdatedAt); err != nil {
			return err
		}
		updated = issue
		return nil
	})
	return updated, err
}

func (s *Store) saveIssue(ctx context.Context, issue domain.Issue, create bool) error {
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		return saveIssueTx(ctx, tx, issue, create)
	})
}

func saveIssueTx(ctx context.Context, tx *sql.Tx, issue domain.Issue, create bool) error {
	if err := ensureIssueVerificationRunTx(ctx, tx, issue); err != nil {
		return err
	}
	if err := ensureIssueArtifactsTx(ctx, tx, issue); err != nil {
		return err
	}
	payload, err := encode(issue)
	if err != nil {
		return err
	}
	imported := map[string]bool{}
	if create {
		if _, err := tx.ExecContext(ctx, `INSERT INTO issues(issue_id, state, updated_at, payload) VALUES (?, ?, ?, ?)`, issue.IssueID, issue.State, unixMillis(issue.UpdatedAt), payload); err != nil {
			if isUniqueViolation(err) {
				return fmt.Errorf("issue %q already exists", issue.IssueID)
			}
			return fmt.Errorf("create issue: %w", err)
		}
	} else {
		var previous []byte
		if err := tx.QueryRowContext(ctx, `SELECT payload FROM issues WHERE issue_id = ?`, issue.IssueID).Scan(&previous); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("issue %q: %w", issue.IssueID, ErrNotFound)
			}
			return fmt.Errorf("load stored issue: %w", err)
		}
		var stored domain.Issue
		if err := decode(previous, &stored); err != nil {
			return fmt.Errorf("decode stored issue: %w", err)
		}
		for _, id := range stored.ObservationIDs {
			imported[id] = true
		}
		result, err := tx.ExecContext(ctx, `UPDATE issues SET state = ?, updated_at = ?, payload = ? WHERE issue_id = ?`, issue.State, unixMillis(issue.UpdatedAt), payload, issue.IssueID)
		if err != nil {
			return fmt.Errorf("update issue: %w", err)
		}
		count, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("count updated issues: %w", err)
		}
		if count != 1 {
			return fmt.Errorf("issue %q: %w", issue.IssueID, ErrNotFound)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM issue_observations WHERE issue_id = ?`, issue.IssueID); err != nil {
			return fmt.Errorf("replace issue observations: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM issue_history WHERE issue_id = ?`, issue.IssueID); err != nil {
			return fmt.Errorf("replace issue history: %w", err)
		}
	}
	for _, observationID := range uniqueSorted(issue.ObservationIDs) {
		if _, err := tx.ExecContext(ctx, `INSERT INTO issue_observations(issue_id, observation_id) VALUES (?, ?)`, issue.IssueID, observationID); err != nil {
			if isForeignKeyViolation(err) {
				// A reference carried by the stored issue whose observation is
				// not in this ledger came from an imported baseline (exports
				// omit observations). It stays in the issue payload as an
				// opaque reference; it just has no evidence row to link.
				if imported[observationID] {
					continue
				}
				return fmt.Errorf("issue %q references unknown observation %q", issue.IssueID, observationID)
			}
			return fmt.Errorf("link issue observation: %w", err)
		}
	}
	for ordinal, entry := range issue.History {
		entryPayload, err := encode(entry)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO issue_history(issue_id, ordinal, payload) VALUES (?, ?, ?)`, issue.IssueID, ordinal, entryPayload); err != nil {
			return fmt.Errorf("save issue history: %w", err)
		}
	}
	return nil
}

func ensureIssueArtifactsTx(ctx context.Context, tx *sql.Tx, issue domain.Issue) error {
	for _, entry := range issue.History {
		for _, artifact := range entry.ArtifactRefs {
			if err := ensureArtifactTx(ctx, tx, artifact); err != nil {
				return fmt.Errorf("issue history artifact: %w", err)
			}
		}
	}
	if issue.Verification != nil {
		for _, artifact := range issue.Verification.ArtifactRefs {
			if err := ensureArtifactTx(ctx, tx, artifact); err != nil {
				return fmt.Errorf("verification artifact: %w", err)
			}
		}
	}
	return nil
}

func ensureIssueVerificationRunTx(ctx context.Context, tx *sql.Tx, issue domain.Issue) error {
	if issue.Verification == nil {
		return nil
	}
	var runState domain.RunState
	if err := tx.QueryRowContext(ctx, `SELECT state FROM runs WHERE run_id = ?`, issue.Verification.RunID).Scan(&runState); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("verification run %q: %w", issue.Verification.RunID, ErrNotFound)
		}
		return fmt.Errorf("load verification run: %w", err)
	}
	return nil
}

func loadIssueTx(ctx context.Context, tx *sql.Tx, issueID string) (domain.Issue, error) {
	var payload []byte
	if err := tx.QueryRowContext(ctx, `SELECT payload FROM issues WHERE issue_id = ?`, issueID).Scan(&payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Issue{}, fmt.Errorf("issue %q: %w", issueID, ErrNotFound)
		}
		return domain.Issue{}, fmt.Errorf("load issue: %w", err)
	}
	var issue domain.Issue
	if err := decode(payload, &issue); err != nil {
		return domain.Issue{}, err
	}
	if err := issue.Validate(); err != nil {
		return domain.Issue{}, fmt.Errorf("stored issue is invalid: %w", err)
	}
	return issue, nil
}

func allowedIssueTransition(from, to domain.IssueState) bool {
	if from == to {
		return true
	}
	switch from {
	case domain.IssueOpen, domain.IssueInvestigating, domain.IssueAccepted, domain.IssueFalsePositive, domain.IssueWontFix, domain.IssueRegressed:
		return to != domain.IssueResolved && to != domain.IssueRegressed
	default:
		// A resolved issue only changes through correlation (regression).
		return false
	}
}

func uniqueSorted(values []string) []string {
	copyValues := append([]string(nil), values...)
	sort.Strings(copyValues)
	return slicesCompact(copyValues)
}

func slicesCompact(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	result := values[:1]
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func isForeignKeyViolation(err error) bool {
	return strings.Contains(err.Error(), "FOREIGN KEY constraint failed")
}

// verificationCoverageRelevantTx proves the verification run contains a
// successful re-check of the issue's own file, family and category.
func verificationCoverageRelevantTx(ctx context.Context, tx *sql.Tx, issue domain.Issue, verification domain.Verification) error {
	rows, err := tx.QueryContext(ctx, `SELECT payload FROM coverage WHERE run_id = ? AND status = ?`, verification.RunID, verification.InspectionStatus)
	if err != nil {
		return fmt.Errorf("query verification coverage: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return fmt.Errorf("scan verification coverage: %w", err)
		}
		var coverage domain.Coverage
		if err := decode(payload, &coverage); err != nil {
			return err
		}
		if issue.CoverageRelevant(coverage) {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return errors.New("resolution verification did not successfully inspect the issue's file")
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

package state

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"scout/internal/domain"
)

// ListIssueObservations returns immutable, sanitized provenance linked to an
// issue, ordered by observation time and ID for stable API and web output.
func (s *Store) ListIssueObservations(ctx context.Context, issueID string) ([]domain.Observation, error) {
	if issueID == "" {
		return nil, errors.New("issue ID is required")
	}
	observations := []domain.Observation{}
	err := s.withRead(ctx, func(db *sql.DB) error {
		var exists string
		if err := db.QueryRowContext(ctx, `SELECT issue_id FROM issues WHERE issue_id = ?`, issueID).Scan(&exists); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("issue %q: %w", issueID, ErrNotFound)
			}
			return fmt.Errorf("load issue: %w", err)
		}
		rows, err := db.QueryContext(ctx, `SELECT o.payload FROM issue_observations io JOIN observations o ON o.observation_id = io.observation_id WHERE io.issue_id = ? ORDER BY o.observed_at, o.observation_id`, issueID)
		if err != nil {
			return fmt.Errorf("query issue observations: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var payload []byte
			if err := rows.Scan(&payload); err != nil {
				return fmt.Errorf("scan issue observation: %w", err)
			}
			var observation domain.Observation
			if err := decode(payload, &observation); err != nil {
				return err
			}
			if err := observation.Validate(); err != nil {
				return fmt.Errorf("stored issue observation is invalid: %w", err)
			}
			observations = append(observations, observation)
		}
		return rows.Err()
	})
	return observations, err
}

func loadObservationsTx(ctx context.Context, tx *sql.Tx, observationIDs []string) ([]domain.Observation, error) {
	observations := make([]domain.Observation, 0, len(observationIDs))
	for _, observationID := range uniqueSorted(observationIDs) {
		var payload []byte
		if err := tx.QueryRowContext(ctx, `SELECT payload FROM observations WHERE observation_id = ?`, observationID).Scan(&payload); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("observation %q: %w", observationID, ErrNotFound)
			}
			return nil, fmt.Errorf("load observation: %w", err)
		}
		var observation domain.Observation
		if err := decode(payload, &observation); err != nil {
			return nil, err
		}
		if err := observation.Validate(); err != nil {
			return nil, fmt.Errorf("stored observation is invalid: %w", err)
		}
		observations = append(observations, observation)
	}
	return observations, nil
}

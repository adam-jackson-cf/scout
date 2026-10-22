package state

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

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

// supportingEvidenceChanged distinguishes source-history line movement from
// changed source/rule/version/location/quote support. A stable source finding
// ID may move, but different quoted evidence still requires re-triage.
func supportingEvidenceChanged(previous, current []domain.Observation) bool {
	if len(previous) == 0 || len(current) == 0 {
		return true
	}
	known := make(map[string]struct{}, len(previous))
	for _, observation := range previous {
		known[observationSupportKey(observation)] = struct{}{}
	}
	for _, observation := range current {
		if _, found := known[observationSupportKey(observation)]; !found {
			return true
		}
	}
	return false
}

func observationSupportKey(observation domain.Observation) string {
	locations := make([]string, 0, len(observation.Evidence))
	for _, evidence := range observation.Evidence {
		location := evidence.Location
		locations = append(locations, strings.Join([]string{location.Path, location.Symbol, fmt.Sprint(location.StartLine), fmt.Sprint(location.EndLine), fmt.Sprint(location.StartByte), fmt.Sprint(location.EndByte), evidence.RedactedQuote}, "\x00"))
	}
	sort.Strings(locations)
	return strings.Join(append([]string{observation.Source, observation.RuleID, observation.RuleVersion, observation.ModelVersion, observation.SourceFindingID}, locations...), "\x00")
}

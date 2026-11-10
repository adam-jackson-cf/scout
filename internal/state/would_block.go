package state

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"scout/internal/domain"
)

// schemaWouldBlock is the shadow-mode block ledger (REQ-22).
var schemaWouldBlock = []string{
	`CREATE TABLE IF NOT EXISTS would_block (id TEXT PRIMARY KEY, head_revision TEXT NOT NULL, base_revision TEXT NOT NULL, issue_id TEXT NOT NULL, check_id TEXT NOT NULL, rule_version TEXT NOT NULL, path TEXT NOT NULL, attribution TEXT NOT NULL, profile TEXT NOT NULL, triage TEXT NOT NULL DEFAULT 'undecided', triage_actor TEXT NOT NULL DEFAULT '', triage_reason TEXT NOT NULL DEFAULT '', recorded_at INTEGER NOT NULL, triaged_at INTEGER)`,
	`CREATE INDEX IF NOT EXISTS would_block_by_check ON would_block(check_id, recorded_at)`,
}

// RecordWouldBlock persists a shadow-mode block. It is idempotent per row ID
// and never overwrites a maintainer's triage.
func (s *Store) RecordWouldBlock(ctx context.Context, row domain.WouldBlock) error {
	if row.Triage == "" {
		row.Triage = domain.TriageUndecided
	}
	if err := row.Validate(); err != nil {
		return err
	}
	if s.readOnly {
		return errors.New("state store is read-only")
	}
	for _, value := range []string{row.CheckID, row.Path, row.RuleVersion, row.TriageReason} {
		if err := rejectUnsafePersistence(value); err != nil {
			return err
		}
	}
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO would_block(id, head_revision, base_revision, issue_id, check_id, rule_version, path, attribution, profile, triage, triage_actor, triage_reason, recorded_at, triaged_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '', '', ?, NULL)
			ON CONFLICT(id) DO UPDATE SET attribution = excluded.attribution, profile = excluded.profile, rule_version = excluded.rule_version`,
			row.ID, row.HeadRevision, row.BaseRevision, row.IssueID, row.CheckID, row.RuleVersion, row.Path, string(row.Attribution), row.Profile, domain.TriageUndecided, unixMillis(nonZero(row.RecordedAt)))
		if err != nil {
			return fmt.Errorf("record would_block: %w", err)
		}
		return nil
	})
}

// ListLegacyWouldBlock exposes only the pre-ADR-012 compatibility representation
// used by the v1 export codec. Its mutable triage fields are never CI authority;
// authorized shadow consumers must use ReadShadowLedger instead.
func (s *Store) ListLegacyWouldBlock(ctx context.Context) ([]domain.WouldBlock, error) {
	rows := []domain.WouldBlock{}
	err := s.withRead(ctx, func(db *sql.DB) error {
		result, err := db.QueryContext(ctx, `SELECT id, head_revision, base_revision, issue_id, check_id, rule_version, path, attribution, profile, triage, triage_actor, triage_reason, recorded_at, triaged_at FROM would_block ORDER BY check_id, recorded_at, id`)
		if err != nil {
			return fmt.Errorf("query would_block: %w", err)
		}
		defer result.Close()
		for result.Next() {
			var row domain.WouldBlock
			var attribution string
			var recorded int64
			var triaged sql.NullInt64
			if err := result.Scan(&row.ID, &row.HeadRevision, &row.BaseRevision, &row.IssueID, &row.CheckID, &row.RuleVersion, &row.Path, &attribution, &row.Profile, &row.Triage, &row.TriageActor, &row.TriageReason, &recorded, &triaged); err != nil {
				return fmt.Errorf("scan would_block: %w", err)
			}
			row.Attribution = domain.Attribution(attribution)
			row.RecordedAt = time.UnixMilli(recorded).UTC()
			if triaged.Valid {
				at := time.UnixMilli(triaged.Int64).UTC()
				row.TriagedAt = &at
			}
			rows = append(rows, row)
		}
		return result.Err()
	})
	return rows, err
}

func nonZero(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now()
	}
	return t
}

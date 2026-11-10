package state

import (
	"context"
	"database/sql"
	"time"

	"scout/internal/domain"
)

// All three supported creation paths use these exact statements. Legacy tables
// remain untouched: the v1 export codec still owns their original representation.
var schemaShadow = []string{
	`CREATE TABLE pr_run_facts (scope TEXT NOT NULL CHECK(scope IN ('ci','local','legacy')), repository_id TEXT NOT NULL, attempt_id TEXT NOT NULL, check_id TEXT NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(scope, repository_id, attempt_id, check_id))`,
	`CREATE TABLE shadow_events (scope TEXT NOT NULL CHECK(scope IN ('ci','local','legacy')), repository_id TEXT NOT NULL, event_id TEXT NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(scope, repository_id, event_id))`,
	`CREATE TABLE pr_run_events (scope TEXT NOT NULL, repository_id TEXT NOT NULL, attempt_id TEXT NOT NULL, check_id TEXT NOT NULL, event_id TEXT NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(scope, repository_id, attempt_id, event_id), FOREIGN KEY(scope, repository_id, attempt_id, check_id) REFERENCES pr_run_facts(scope, repository_id, attempt_id, check_id), FOREIGN KEY(scope, repository_id, event_id) REFERENCES shadow_events(scope, repository_id, event_id))`,
	`CREATE TABLE would_block_triage (scope TEXT NOT NULL CHECK(scope IN ('ci','local','legacy')), repository_id TEXT NOT NULL, action_id TEXT NOT NULL, event_id TEXT NOT NULL, publication_id TEXT, payload BLOB NOT NULL, PRIMARY KEY(scope, repository_id, action_id), UNIQUE(scope, repository_id, publication_id, event_id), FOREIGN KEY(scope, repository_id, event_id) REFERENCES shadow_events(scope, repository_id, event_id), FOREIGN KEY(repository_id, publication_id) REFERENCES shadow_publications(repository_id, publication_id))`,
	`CREATE TABLE shadow_publications (repository_id TEXT NOT NULL, publication_id TEXT NOT NULL, parent_id TEXT, recorded_at INTEGER NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(repository_id, publication_id), UNIQUE(repository_id, parent_id), FOREIGN KEY(repository_id, parent_id) REFERENCES shadow_publications(repository_id, publication_id))`,
	`CREATE UNIQUE INDEX shadow_publication_genesis ON shadow_publications(repository_id) WHERE parent_id IS NULL`,
	`CREATE TABLE shadow_frontiers (repository_id TEXT NOT NULL, check_id TEXT NOT NULL, rule_version TEXT NOT NULL, complete_through INTEGER NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(repository_id, check_id, rule_version, complete_through))`,
	`CREATE TABLE shadow_gap_points (repository_id TEXT NOT NULL, point_id TEXT NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(repository_id, point_id))`,
	`CREATE TABLE shadow_intervals (repository_id TEXT NOT NULL, interval_id TEXT NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(repository_id, interval_id))`,
	`CREATE TABLE shadow_event_obligations (scope TEXT NOT NULL CHECK(scope IN ('ci','local','legacy')), repository_id TEXT NOT NULL, event_id TEXT NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(scope, repository_id, event_id))`,
	`CREATE TABLE pr_run_obligations (scope TEXT NOT NULL, repository_id TEXT NOT NULL, attempt_id TEXT NOT NULL, check_id TEXT NOT NULL, event_id TEXT NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(scope, repository_id, attempt_id, check_id, event_id), FOREIGN KEY(scope, repository_id, attempt_id, check_id) REFERENCES pr_run_facts(scope, repository_id, attempt_id, check_id), FOREIGN KEY(scope, repository_id, event_id) REFERENCES shadow_event_obligations(scope, repository_id, event_id))`,
	`CREATE TABLE shadow_import_refusals (scope TEXT NOT NULL CHECK(scope='ci'), repository_id TEXT NOT NULL, refusal_id TEXT NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(scope, repository_id, refusal_id))`,
	`CREATE TABLE shadow_pending_imports (scope TEXT NOT NULL CHECK(scope='ci'), repository_id TEXT NOT NULL, import_id TEXT NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(scope, repository_id, import_id))`,
}

func migrateLegacyShadowTriage(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `SELECT id,triage,triage_actor,triage_reason,triaged_at FROM would_block WHERE triaged_at IS NOT NULL ORDER BY id`)
	if err != nil {
		return err
	}
	var legacy []domain.ShadowTriage
	for rows.Next() {
		var row domain.ShadowTriage
		var at int64
		if err := rows.Scan(&row.EventID, &row.Verdict, &row.Actor, &row.Reason, &at); err != nil {
			rows.Close()
			return err
		}
		row.ActionID = "legacy:" + row.EventID
		row.Scope = domain.ShadowScopeLegacy
		row.TriagedAt = time.UnixMilli(at).UTC()
		legacy = append(legacy, row)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, row := range legacy {
		// Historical rows lack provider repository/PR identity. Empty legacy-only
		// identity preserves that absence rather than forging authenticated facts.
		event := domain.ProviderPREvent{EventID: row.EventID, Scope: domain.ShadowScopeLegacy}
		if err := insertShadowJSON(ctx, tx, "shadow_events", []string{"scope", "repository_id", "event_id"}, []any{row.Scope, "", row.EventID}, event); err != nil {
			return err
		}
		if err := insertShadowJSON(ctx, tx, "would_block_triage", []string{"scope", "repository_id", "action_id", "event_id"}, []any{row.Scope, "", row.ActionID, row.EventID}, row); err != nil {
			return err
		}
	}
	return nil
}

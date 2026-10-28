package state

import (
	"context"
	"database/sql"
	"errors"
)

// schemaPRMetadata records which PR-metadata document backed the PR-level signals of
// a snapshot (design §8.3 provenance). Never holds tokens or PR contents.
var schemaPRMetadata = []string{
	`CREATE TABLE IF NOT EXISTS pr_metadata_sources (seq INTEGER PRIMARY KEY AUTOINCREMENT, snapshot TEXT NOT NULL, kind TEXT NOT NULL, origin TEXT NOT NULL, digest TEXT NOT NULL, fetched_at TEXT NOT NULL, prs INTEGER NOT NULL, UNIQUE(snapshot, digest))`,
}

// PRMetadataSource is the provenance of an ingested PR-metadata document.
type PRMetadataSource struct {
	Snapshot  string `json:"snapshot"`
	Kind      string `json:"kind"`
	Origin    string `json:"origin"`
	Digest    string `json:"digest"`
	FetchedAt string `json:"fetched_at,omitempty"`
	PRs       int    `json:"prs"`
}

// RecordPRMetadataSource appends the provenance row; an identical
// (snapshot, digest) is left untouched.
func (s *Store) RecordPRMetadataSource(ctx context.Context, src PRMetadataSource) error {
	if s.readOnly {
		return errors.New("state store is read-only")
	}
	for _, v := range []string{src.Snapshot, src.Origin} {
		if err := rejectUnsafePersistence(v); err != nil {
			return err
		}
	}
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO pr_metadata_sources(snapshot, kind, origin, digest, fetched_at, prs)
			VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT DO NOTHING`, src.Snapshot, src.Kind, src.Origin, src.Digest, src.FetchedAt, src.PRs)
		return err
	})
}

// PRMetadataSources lists recorded sources for a snapshot, oldest first.
func (s *Store) PRMetadataSources(ctx context.Context, snapshot string) ([]PRMetadataSource, error) {
	var out []PRMetadataSource
	err := s.withRead(ctx, func(db *sql.DB) error {
		rows, err := db.QueryContext(ctx, `SELECT snapshot, kind, origin, digest, fetched_at, prs FROM pr_metadata_sources WHERE snapshot = ? ORDER BY seq`, snapshot)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var r PRMetadataSource
			if err := rows.Scan(&r.Snapshot, &r.Kind, &r.Origin, &r.Digest, &r.FetchedAt, &r.PRs); err != nil {
				return err
			}
			out = append(out, r)
		}
		return rows.Err()
	})
	return out, err
}

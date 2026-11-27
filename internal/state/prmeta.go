package state

import (
	"context"
	"database/sql"
	"errors"
)

// schemaPRMetadata records which PR-metadata document backed the PR-level signals of
// a snapshot (REQ-132 provenance). Never holds tokens or PR contents.
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

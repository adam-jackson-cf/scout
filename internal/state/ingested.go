package state

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// schemaIngested records the provenance ledger of ingested coverage and test-result
// artifacts (design §8.2, §8.7): which file was read for which snapshot, who
// produced it, the revision it claims, and whether it matched the snapshot.
var schemaIngested = []string{
	`CREATE TABLE IF NOT EXISTS ingested_artifacts (snapshot TEXT NOT NULL, digest TEXT NOT NULL, kind TEXT NOT NULL, revision TEXT NOT NULL, producer TEXT NOT NULL, source TEXT NOT NULL, revision_match TEXT NOT NULL, reason TEXT NOT NULL, ingested_at INTEGER NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(snapshot, digest))`,
	`CREATE INDEX IF NOT EXISTS ingested_artifacts_by_snapshot ON ingested_artifacts(snapshot, kind)`,
}

// IngestedArtifact is the stored provenance of one artifact ingested for one
// snapshot. Payload is the artifact's JSON summary (for example JUnit totals).
type IngestedArtifact struct {
	Snapshot string `json:"snapshot"`
	// Digest is the SHA-256 of the artifact bytes as read.
	Digest   string `json:"digest"`
	Kind     string `json:"kind"` // lcov, cobertura, junit
	Revision string `json:"revision,omitempty"`
	Producer string `json:"producer,omitempty"`
	Source   string `json:"source"` // path as given on the command line or discovered
	// RevisionMatch is "match", "mismatch" or "unknown" (the artifact states no revision).
	RevisionMatch string `json:"revisionMatch"`
	Reason        string `json:"reason,omitempty"`
	Payload       []byte `json:"payload,omitempty"`
}

// RecordIngestedArtifact stores artifact provenance; re-ingesting the same
// bytes for the same snapshot is idempotent.
func (s *Store) RecordIngestedArtifact(ctx context.Context, a IngestedArtifact) error {
	if s.readOnly {
		return errors.New("state store is read-only")
	}
	if a.Snapshot == "" || a.Digest == "" || a.Kind == "" || a.RevisionMatch == "" {
		return errors.New("ingested artifact requires snapshot, digest, kind and revision match")
	}
	if err := rejectUnsafePersistence(string(a.Payload)); err != nil {
		return fmt.Errorf("ingested artifact payload: %w", err)
	}
	if a.Payload == nil {
		a.Payload = []byte{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.New("state store is closed")
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO ingested_artifacts(snapshot, digest, kind, revision, producer, source, revision_match, reason, ingested_at, payload) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.Snapshot, a.Digest, a.Kind, a.Revision, a.Producer, a.Source, a.RevisionMatch, a.Reason, unixMillis(time.Now()), a.Payload)
	if err == nil {
		return nil
	}
	if !isUniqueViolation(err) {
		return fmt.Errorf("record ingested artifact: %w", err)
	}
	var stored IngestedArtifact
	if err := s.db.QueryRowContext(ctx, `SELECT kind, revision, source, revision_match FROM ingested_artifacts WHERE snapshot = ? AND digest = ?`, a.Snapshot, a.Digest).
		Scan(&stored.Kind, &stored.Revision, &stored.Source, &stored.RevisionMatch); err != nil {
		return fmt.Errorf("record ingested artifact: %w", err)
	}
	if stored.Kind != a.Kind || stored.Revision != a.Revision || stored.Source != a.Source || stored.RevisionMatch != a.RevisionMatch {
		return fmt.Errorf("ingested artifact %s for snapshot %s is already recorded with different provenance", a.Digest, a.Snapshot)
	}
	return nil
}

// ListIngestedArtifacts returns a snapshot's artifacts ordered by kind, digest.
func (s *Store) ListIngestedArtifacts(ctx context.Context, snapshot string) ([]IngestedArtifact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, errors.New("state store is closed")
	}
	rows, err := s.db.QueryContext(ctx, `SELECT snapshot, digest, kind, revision, producer, source, revision_match, reason, payload FROM ingested_artifacts WHERE snapshot = ? ORDER BY kind, digest`, snapshot)
	if err != nil {
		return nil, fmt.Errorf("list ingested artifacts: %w", err)
	}
	defer rows.Close()
	var out []IngestedArtifact
	for rows.Next() {
		var a IngestedArtifact
		if err := rows.Scan(&a.Snapshot, &a.Digest, &a.Kind, &a.Revision, &a.Producer, &a.Source, &a.RevisionMatch, &a.Reason, &a.Payload); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

package statetest

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"
	"path/filepath"
	"sort"

	"scout/internal/domain"
	"scout/internal/state"
)

func query[T any](ctx context.Context, store *state.Store, scan func(*sql.Rows) (T, error), statement string, args ...any) (out []T, err error) {
	path := filepath.Join(store.Root(), ".scout", "state.sqlite")
	db, err := sql.Open("sqlite", (&url.URL{Scheme: "file", Path: path, RawQuery: "mode=ro"}).String())
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	rows, err := db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		value, scanErr := scan(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, value)
	}
	return out, rows.Err()
}

func payload[T any](rows *sql.Rows) (T, error) {
	var raw []byte
	var value T
	if err := rows.Scan(&raw); err != nil {
		return value, err
	}
	return value, json.Unmarshal(raw, &value)
}

func Rankings(ctx context.Context, store *state.Store, runID string) ([]domain.RankingRecord, error) {
	out, err := query(ctx, store, payload[domain.RankingRecord], `SELECT payload FROM ranking_records WHERE run_id = ?`, runID)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].RankPosition != out[j].RankPosition {
			return out[i].RankPosition < out[j].RankPosition
		}
		return out[i].IssueID < out[j].IssueID
	})
	return out, err
}

func Annotations(ctx context.Context, store *state.Store, issueID string) ([]domain.Annotation, error) {
	return query(ctx, store, func(rows *sql.Rows) (domain.Annotation, error) {
		var raw []byte
		var stale int
		var annotation domain.Annotation
		if err := rows.Scan(&raw, &stale); err != nil {
			return annotation, err
		}
		err := json.Unmarshal(raw, &annotation)
		annotation.Stale = stale != 0
		return annotation, err
	}, `SELECT payload, stale FROM annotations WHERE issue_id = ? ORDER BY created_at, annotation_id`, issueID)
}

func AuditDraws(ctx context.Context, store *state.Store) ([]domain.AuditDraw, error) {
	return query(ctx, store, payload[domain.AuditDraw], `SELECT payload FROM audit_draws ORDER BY issue_id, formula_version, sampler_version`)
}

func IngestedArtifacts(ctx context.Context, store *state.Store, snapshot string) ([]state.IngestedArtifact, error) {
	return query(ctx, store, func(rows *sql.Rows) (state.IngestedArtifact, error) {
		var a state.IngestedArtifact
		err := rows.Scan(&a.Snapshot, &a.Digest, &a.Kind, &a.Revision, &a.Producer, &a.Source, &a.RevisionMatch, &a.Reason, &a.Payload)
		return a, err
	}, `SELECT snapshot, digest, kind, revision, producer, source, revision_match, reason, payload FROM ingested_artifacts WHERE snapshot = ? ORDER BY kind, digest`, snapshot)
}

func PRMetadataSources(ctx context.Context, store *state.Store, snapshot string) ([]state.PRMetadataSource, error) {
	return query(ctx, store, func(rows *sql.Rows) (state.PRMetadataSource, error) {
		var r state.PRMetadataSource
		err := rows.Scan(&r.Snapshot, &r.Kind, &r.Origin, &r.Digest, &r.FetchedAt, &r.PRs)
		return r, err
	}, `SELECT snapshot, kind, origin, digest, fetched_at, prs FROM pr_metadata_sources WHERE snapshot = ? ORDER BY seq`, snapshot)
}

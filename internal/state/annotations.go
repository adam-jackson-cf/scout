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

// schemaV8 adds the decision-model tables (design §6): annotations (advisory
// answers), ranking records (disclosure plus baseline and rank-formula
// versions for every finding), audit draws (stored inclusion probability and
// sampler version) and the view context of every human disposition.
var schemaV8 = []string{
	`CREATE TABLE IF NOT EXISTS annotations (annotation_id TEXT PRIMARY KEY, run_id TEXT NOT NULL, issue_id TEXT NOT NULL, unit_key TEXT NOT NULL, question_version TEXT NOT NULL, stale INTEGER NOT NULL DEFAULT 0, created_at INTEGER NOT NULL, payload BLOB NOT NULL)`,
	`CREATE INDEX IF NOT EXISTS annotations_by_issue ON annotations(issue_id, created_at, annotation_id)`,
	`CREATE TABLE IF NOT EXISTS ranking_records (run_id TEXT NOT NULL, issue_id TEXT NOT NULL, mode TEXT NOT NULL, created_at INTEGER NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(run_id, issue_id))`,
	`CREATE INDEX IF NOT EXISTS ranking_records_by_issue ON ranking_records(issue_id, created_at)`,
	`CREATE TABLE IF NOT EXISTS audit_draws (issue_id TEXT NOT NULL, formula_version TEXT NOT NULL, sampler_version TEXT NOT NULL, run_id TEXT NOT NULL, drawn INTEGER NOT NULL, inclusion_probability REAL NOT NULL, created_at INTEGER NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(issue_id, formula_version, sampler_version))`,
	`CREATE TABLE IF NOT EXISTS disposition_contexts (id INTEGER PRIMARY KEY AUTOINCREMENT, issue_id TEXT NOT NULL, at INTEGER NOT NULL, view TEXT NOT NULL, annotation_visible INTEGER NOT NULL, rank_formula_version TEXT NOT NULL, baseline_version TEXT NOT NULL)`,
	`CREATE INDEX IF NOT EXISTS disposition_contexts_by_issue ON disposition_contexts(issue_id, id)`,
}

type dispositionViewKey struct{}

type dispositionView struct {
	view    string
	visible bool
}

// WithDispositionView tags dispositions recorded under ctx with the view they
// are made in and whether the annotation was visible to the maintainer. A
// disposition recorded without it is stored as view "cli", annotation hidden.
func WithDispositionView(ctx context.Context, view string, annotationVisible bool) context.Context {
	return context.WithValue(ctx, dispositionViewKey{}, dispositionView{view: view, visible: annotationVisible})
}

// RecordAnnotations persists a unit's annotations in one transaction: all rows
// are stored or none. Request bodies and answers are refused when they carry
// credential-shaped content.
func (s *Store) RecordAnnotations(ctx context.Context, annotations []domain.Annotation) error {
	if s.readOnly {
		return errors.New("state store is read-only")
	}
	type row struct {
		annotation domain.Annotation
		payload    []byte
	}
	rows := make([]row, 0, len(annotations))
	for _, annotation := range annotations {
		if err := annotation.Validate(); err != nil {
			return err
		}
		if annotation.CreatedAt.IsZero() {
			annotation.CreatedAt = time.Now().UTC()
		}
		for _, value := range []string{string(annotation.RequestBody), string(annotation.Answer), string(annotation.Usage)} {
			if err := rejectUnsafePersistence(value); err != nil {
				return err
			}
		}
		payload, err := encode(annotation)
		if err != nil {
			return err
		}
		rows = append(rows, row{annotation, payload})
	}
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		for _, r := range rows {
			a := r.annotation
			if _, err := tx.ExecContext(ctx, `INSERT INTO annotations(annotation_id, run_id, issue_id, unit_key, question_version, stale, created_at, payload) VALUES (?, ?, ?, ?, ?, 0, ?, ?)`,
				a.AnnotationID, a.RunID, a.IssueID, a.UnitKey, a.QuestionVersion, unixMillis(a.CreatedAt), r.payload); err != nil {
				return fmt.Errorf("record annotation: %w", err)
			}
		}
		return nil
	})
}

// ListAnnotations returns the annotations of one issue, oldest first, with
// Stale reflecting the current stale flag.
func (s *Store) ListAnnotations(ctx context.Context, issueID string) ([]domain.Annotation, error) {
	out := []domain.Annotation{}
	err := s.withRead(ctx, func(db *sql.DB) error {
		rows, err := db.QueryContext(ctx, `SELECT payload, stale FROM annotations WHERE issue_id = ? ORDER BY created_at, annotation_id`, issueID)
		if err != nil {
			return fmt.Errorf("query annotations: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var payload []byte
			var stale int
			if err := rows.Scan(&payload, &stale); err != nil {
				return fmt.Errorf("scan annotation: %w", err)
			}
			var annotation domain.Annotation
			if err := decode(payload, &annotation); err != nil {
				return err
			}
			annotation.Stale = stale != 0
			out = append(out, annotation)
		}
		return rows.Err()
	})
	return out, err
}

// MarkAnnotationsStale flags every annotation whose question version is not
// in current (a model bump or question edit creates a new version key) and
// returns how many rows changed.
func (s *Store) MarkAnnotationsStale(ctx context.Context, current []string) (int, error) {
	if s.readOnly {
		return 0, errors.New("state store is read-only")
	}
	query := `UPDATE annotations SET stale = 1 WHERE stale = 0`
	args := make([]any, 0, len(current))
	if len(current) > 0 {
		marks := strings.TrimSuffix(strings.Repeat("?,", len(current)), ",")
		query += ` AND question_version NOT IN (` + marks + `)`
		for _, version := range current {
			args = append(args, version)
		}
	}
	changed := 0
	err := s.withTransaction(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("mark annotations stale: %w", err)
		}
		count, err := result.RowsAffected()
		changed = int(count)
		return err
	})
	return changed, err
}

// RecordRanking stores one finding's ranking disclosure for a run (replacing
// a prior row of the same run and issue).
func (s *Store) RecordRanking(ctx context.Context, record domain.RankingRecord) error {
	if s.readOnly {
		return errors.New("state store is read-only")
	}
	if record.RunID == "" || record.IssueID == "" || record.Mode == "" || record.BaselineVersion == "" || record.RankFormulaVersion == "" {
		return errors.New("ranking record requires run, issue, mode, baseline version and rank-formula version")
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	}
	payload, err := encode(record)
	if err != nil {
		return err
	}
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO ranking_records(run_id, issue_id, mode, created_at, payload) VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(run_id, issue_id) DO UPDATE SET mode = excluded.mode, created_at = excluded.created_at, payload = excluded.payload`,
			record.RunID, record.IssueID, record.Mode, unixMillis(record.CreatedAt), payload)
		if err != nil {
			return fmt.Errorf("record ranking: %w", err)
		}
		return nil
	})
}

// ListRankings returns the ranking records of one run in rank order.
func (s *Store) ListRankings(ctx context.Context, runID string) ([]domain.RankingRecord, error) {
	out := []domain.RankingRecord{}
	err := s.withRead(ctx, func(db *sql.DB) error {
		rows, err := db.QueryContext(ctx, `SELECT payload FROM ranking_records WHERE run_id = ?`, runID)
		if err != nil {
			return fmt.Errorf("query rankings: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var payload []byte
			if err := rows.Scan(&payload); err != nil {
				return fmt.Errorf("scan ranking: %w", err)
			}
			var record domain.RankingRecord
			if err := decode(payload, &record); err != nil {
				return err
			}
			out = append(out, record)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	sortRankings(out)
	return out, nil
}

func sortRankings(rows []domain.RankingRecord) {
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].RankPosition != rows[j].RankPosition {
			return rows[i].RankPosition < rows[j].RankPosition
		}
		return rows[i].IssueID < rows[j].IssueID
	})
}

// AuditDraw returns the stored draw for (issue, formula, sampler), if any.
func (s *Store) AuditDraw(ctx context.Context, issueID, formulaVersion, samplerVersion string) (domain.AuditDraw, bool, error) {
	var draw domain.AuditDraw
	found := false
	err := s.withRead(ctx, func(db *sql.DB) error {
		var payload []byte
		err := db.QueryRowContext(ctx, `SELECT payload FROM audit_draws WHERE issue_id = ? AND formula_version = ? AND sampler_version = ?`, issueID, formulaVersion, samplerVersion).Scan(&payload)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("query audit draw: %w", err)
		}
		found = true
		return decode(payload, &draw)
	})
	return draw, found, err
}

// RecordAuditDraw stores a draw once per (issue, formula version, sampler
// version): the first eligible run's draw and its inclusion probability are
// never replaced. The stored row is returned.
func (s *Store) RecordAuditDraw(ctx context.Context, draw domain.AuditDraw) (domain.AuditDraw, error) {
	if s.readOnly {
		return draw, errors.New("state store is read-only")
	}
	if draw.IssueID == "" || draw.FormulaVersion == "" || draw.SamplerVersion == "" || draw.RunID == "" {
		return draw, errors.New("audit draw requires issue, run, formula version and sampler version")
	}
	if draw.InclusionProbability <= 0 || draw.InclusionProbability > 1 {
		return draw, errors.New("audit draw inclusion probability must be in (0, 1]")
	}
	if draw.CreatedAt.IsZero() {
		draw.CreatedAt = time.Now().UTC()
	}
	payload, err := encode(draw)
	if err != nil {
		return draw, err
	}
	err = s.withTransaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO audit_draws(issue_id, formula_version, sampler_version, run_id, drawn, inclusion_probability, created_at, payload) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			draw.IssueID, draw.FormulaVersion, draw.SamplerVersion, draw.RunID, boolInt(draw.Drawn), draw.InclusionProbability, unixMillis(draw.CreatedAt), payload)
		if err != nil {
			return fmt.Errorf("record audit draw: %w", err)
		}
		return nil
	})
	if err != nil {
		return draw, err
	}
	stored, _, err := s.AuditDraw(ctx, draw.IssueID, draw.FormulaVersion, draw.SamplerVersion)
	return stored, err
}

// ListAuditDraws returns every stored draw in issue order.
func (s *Store) ListAuditDraws(ctx context.Context) ([]domain.AuditDraw, error) {
	out := []domain.AuditDraw{}
	err := s.withRead(ctx, func(db *sql.DB) error {
		rows, err := db.QueryContext(ctx, `SELECT payload FROM audit_draws ORDER BY issue_id, formula_version, sampler_version`)
		if err != nil {
			return fmt.Errorf("query audit draws: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var payload []byte
			if err := rows.Scan(&payload); err != nil {
				return fmt.Errorf("scan audit draw: %w", err)
			}
			var draw domain.AuditDraw
			if err := decode(payload, &draw); err != nil {
				return err
			}
			out = append(out, draw)
		}
		return rows.Err()
	})
	return out, err
}

// ListDispositionContexts returns the view context of every disposition of an
// issue, oldest first.
func (s *Store) ListDispositionContexts(ctx context.Context, issueID string) ([]domain.DispositionContext, error) {
	out := []domain.DispositionContext{}
	err := s.withRead(ctx, func(db *sql.DB) error {
		rows, err := db.QueryContext(ctx, `SELECT issue_id, at, view, annotation_visible, rank_formula_version, baseline_version FROM disposition_contexts WHERE issue_id = ? ORDER BY id`, issueID)
		if err != nil {
			return fmt.Errorf("query disposition contexts: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var row domain.DispositionContext
			var at int64
			var visible int
			if err := rows.Scan(&row.IssueID, &at, &row.View, &visible, &row.RankFormulaVersion, &row.BaselineVersion); err != nil {
				return fmt.Errorf("scan disposition context: %w", err)
			}
			row.At = time.UnixMilli(at).UTC()
			row.AnnotationVisible = visible != 0
			out = append(out, row)
		}
		return rows.Err()
	})
	return out, err
}

// recordDispositionContextTx stores the view context of a disposition inside
// the transition's transaction, with the versions of the issue's latest
// ranking record.
func recordDispositionContextTx(ctx context.Context, tx *sql.Tx, issueID string, at time.Time) error {
	view := dispositionView{view: domain.ViewCLI}
	if tagged, ok := ctx.Value(dispositionViewKey{}).(dispositionView); ok {
		view = tagged
	}
	var formula, baseline string
	var payload []byte
	err := tx.QueryRowContext(ctx, `SELECT payload FROM ranking_records WHERE issue_id = ? ORDER BY created_at DESC, run_id DESC LIMIT 1`, issueID).Scan(&payload)
	switch {
	case errors.Is(err, sql.ErrNoRows):
	case err != nil:
		return fmt.Errorf("read latest ranking: %w", err)
	default:
		var record domain.RankingRecord
		if err := decode(payload, &record); err != nil {
			return err
		}
		formula, baseline = record.RankFormulaVersion, record.BaselineVersion
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO disposition_contexts(issue_id, at, view, annotation_visible, rank_formula_version, baseline_version) VALUES (?, ?, ?, ?, ?, ?)`,
		issueID, unixMillis(at), view.view, boolInt(view.visible), formula, baseline); err != nil {
		return fmt.Errorf("record disposition context: %w", err)
	}
	return nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

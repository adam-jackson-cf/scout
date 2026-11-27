package state

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"scout/internal/domain"
)

const viewAudit = "audit"

// DispositionContext records, for one human disposition, the view it was made
// in, whether the annotation was visible, and the versions then current.
type DispositionContext struct {
	IssueID            string    `json:"issueId"`
	At                 time.Time `json:"at"`
	View               string    `json:"view"`
	AnnotationVisible  bool      `json:"annotationVisible"`
	RankFormulaVersion string    `json:"rankFormulaVersion,omitempty"`
	BaselineVersion    string    `json:"baselineVersion,omitempty"`
}

func annotationFor(issue, version string) domain.Annotation {
	return domain.Annotation{
		AnnotationID: "annotation:" + issue + version, RunID: "run-1", IssueID: issue, UnitKey: issue, QuestionID: "triage.real_defect",
		QuestionVersion: version, Kind: "noul", ModelRequested: "typesafe/jev-1.14-20261014", ModelResolved: "typesafe/jev-1.14-20261014",
		Probability: 0.83, Answer: json.RawMessage(`{"type":"noul","noul":0.83}`), RequestBody: json.RawMessage(`{"model":"m","state":{"excerpt":["x"]}}`),
		OptionOrder: []string{"true", "false"},
	}
}

// WithDispositionView tags dispositions recorded under ctx with the view they
// are made in and whether the annotation was visible to the maintainer. A
// disposition recorded without it is stored as view "cli", annotation hidden.
func WithDispositionView(ctx context.Context, view string, annotationVisible bool) context.Context {
	return context.WithValue(ctx, dispositionViewKey{}, dispositionView{view: view, visible: annotationVisible})
}

// ListDispositionContexts returns the view context of every disposition of an
// issue, oldest first.
func (s *Store) ListDispositionContexts(ctx context.Context, issueID string) ([]DispositionContext, error) {
	out := []DispositionContext{}
	err := s.withRead(ctx, func(db *sql.DB) error {
		rows, err := db.QueryContext(ctx, `SELECT issue_id, at, view, annotation_visible, rank_formula_version, baseline_version FROM disposition_contexts WHERE issue_id = ? ORDER BY id`, issueID)
		if err != nil {
			return fmt.Errorf("query disposition contexts: %w", err)
		}
		return scanRows(rows, func(rows *sql.Rows) error {
			var row DispositionContext
			var at int64
			var visible int
			if err := rows.Scan(&row.IssueID, &at, &row.View, &visible, &row.RankFormulaVersion, &row.BaselineVersion); err != nil {
				return fmt.Errorf("scan disposition context: %w", err)
			}
			row.At = time.UnixMilli(at).UTC()
			row.AnnotationVisible = visible != 0
			out = append(out, row)
			return nil
		})
	})
	return out, err
}

func TestSchemaV8StoresAnnotationsRankingsAndAuditDraws(t *testing.T) {
	ctx := context.Background()
	store := seedLedger(t, ctx, t.TempDir())
	defer closeOrFail(t, store)
	var version int
	if err := store.db.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&version); err != nil || version != schemaVersion {
		t.Fatalf("schema version %d err %v", version, err)
	}
	issue := "issue:security:a"
	checkAnnotationsStaleness(t, ctx, store, issue)
	checkRankingReplacement(t, ctx, store, issue)
	checkAuditDrawKeepsFirst(t, ctx, store, issue)
}

func checkAnnotationsStaleness(t *testing.T, ctx context.Context, store *Store, issue string) {
	t.Helper()
	if err := store.RecordAnnotations(ctx, []domain.Annotation{annotationFor(issue, "triage.real_defect@old")}); err != nil {
		t.Fatal(err)
	}
	if err := store.RecordAnnotations(ctx, []domain.Annotation{annotationFor(issue, "triage.real_defect@new")}); err != nil {
		t.Fatal(err)
	}
	// A model bump creates a new version key: annotations under another key are stale.
	changed, err := store.MarkAnnotationsStale(ctx, []string{"triage.real_defect@new"})
	if err != nil || changed != 1 {
		t.Fatalf("stale %d err %v", changed, err)
	}
	rows, err := store.ListAnnotations(ctx, issue)
	if err != nil || len(rows) != 2 {
		t.Fatalf("annotations %d err %v", len(rows), err)
	}
	checkAnnotationRows(t, rows)
	// An annotation carrying a credential is refused, never persisted.
	unsafe := annotationFor(issue, "triage.real_defect@new")
	unsafe.AnnotationID = "annotation:unsafe"
	unsafe.RequestBody = json.RawMessage(`{"state":{"excerpt":["api_key = abcdef123456789"]}}`)
	if err := store.RecordAnnotations(ctx, []domain.Annotation{unsafe}); err == nil {
		t.Fatal("credential-bearing request body must be refused")
	}
}

func checkAnnotationRows(t *testing.T, rows []domain.Annotation) {
	t.Helper()
	for _, row := range rows {
		wantStale := strings.HasSuffix(row.QuestionVersion, "@old")
		if row.Stale != wantStale || row.Probability != 0.83 || len(row.RequestBody) == 0 || row.ModelResolved == "" {
			t.Fatalf("row %+v", row)
		}
	}
}

func checkRankingReplacement(t *testing.T, ctx context.Context, store *Store, issue string) {
	t.Helper()
	record := domain.RankingRecord{RunID: "run-1", IssueID: issue, Mode: "deterministic", Disclosure: "deterministic (fork)", BaselineVersion: "base@1", RankFormulaVersion: "formula@1", RankPosition: 1}
	if err := store.RecordRanking(ctx, record); err != nil {
		t.Fatal(err)
	}
	record.Mode, record.RankFormulaVersion = "jev", "formula@2"
	if err := store.RecordRanking(ctx, record); err != nil {
		t.Fatal(err)
	}
	rankings, err := store.ListRankings(ctx, "run-1")
	if err != nil || len(rankings) != 1 || rankings[0].RankFormulaVersion != "formula@2" || rankings[0].BaselineVersion != "base@1" {
		t.Fatalf("rankings %+v err %v", rankings, err)
	}
	if err := store.RecordRanking(ctx, domain.RankingRecord{RunID: "run-1", IssueID: issue, Mode: "jev"}); err == nil {
		t.Fatal("a ranking record without versions must be refused")
	}
}

func checkAuditDrawKeepsFirst(t *testing.T, ctx context.Context, store *Store, issue string) {
	t.Helper()
	draw := domain.AuditDraw{RunID: "run-1", IssueID: issue, FormulaVersion: "formula@2", SamplerVersion: "interim-audit/v1", Stratum: "top", InclusionProbability: 0.125, Drawn: true, RankPosition: 1}
	stored, err := store.RecordAuditDraw(ctx, draw)
	if err != nil || stored.InclusionProbability != 0.125 || !stored.Drawn {
		t.Fatalf("draw %+v err %v", stored, err)
	}
	second := draw
	second.RunID, second.InclusionProbability, second.Drawn = "run-2", 0.5, false
	if kept, err := store.RecordAuditDraw(ctx, second); err != nil || kept.RunID != "run-1" || kept.InclusionProbability != 0.125 || !kept.Drawn {
		t.Fatalf("first eligible draw must be kept: %+v err %v", kept, err)
	}
	for _, p := range []float64{0, 1.5} {
		if _, err := store.RecordAuditDraw(ctx, domain.AuditDraw{RunID: "r", IssueID: "i", FormulaVersion: "f", SamplerVersion: "s", InclusionProbability: p}); err == nil {
			t.Fatalf("inclusion probability %v must be refused", p)
		}
	}
}

func TestDispositionsRecordViewAnnotationVisibilityAndRankingVersions(t *testing.T) {
	ctx := context.Background()
	store := seedLedger(t, ctx, t.TempDir())
	defer closeOrFail(t, store)
	issue := "issue:security:a"
	if err := store.RecordRanking(ctx, domain.RankingRecord{RunID: "run-1", IssueID: issue, Mode: "jev", BaselineVersion: "base@1", RankFormulaVersion: "formula@1", CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.TriageIssue(WithDispositionView(ctx, viewAudit, false), Triage{IssueID: issue, State: domain.IssueAccepted, Owner: "owner", DispositionReason: "known", Actor: "maintainer", Reason: "reviewed blind"}); err != nil {
		t.Fatal(err)
	}
	contexts, err := store.ListDispositionContexts(ctx, issue)
	if err != nil || len(contexts) != 1 {
		t.Fatalf("contexts %+v err %v", contexts, err)
	}
	got := contexts[0]
	if got.View != viewAudit || got.AnnotationVisible || got.RankFormulaVersion != "formula@1" || got.BaselineVersion != "base@1" {
		t.Fatalf("context %+v", got)
	}
	checkUntaggedAndVisibleDispositions(t, ctx, store)
}

func checkUntaggedAndVisibleDispositions(t *testing.T, ctx context.Context, store *Store) {
	t.Helper()
	// An untagged disposition is stored as the CLI view with the annotation hidden.
	if _, err := store.TriageIssue(ctx, Triage{IssueID: "issue:security:c", State: domain.IssueAccepted, Owner: "o", DispositionReason: "known", Actor: "maintainer", Reason: "reviewed"}); err != nil {
		t.Fatal(err)
	}
	untagged, err := store.ListDispositionContexts(ctx, "issue:security:c")
	if err != nil || len(untagged) != 1 || untagged[0].View != domain.ViewCLI || untagged[0].AnnotationVisible {
		t.Fatalf("untagged %+v err %v", untagged, err)
	}
	if _, err := store.TriageIssue(WithDispositionView(ctx, "list", true), Triage{IssueID: "issue:security:e", State: domain.IssueAccepted, Owner: "o", DispositionReason: "known", Actor: "maintainer", Reason: "reviewed"}); err != nil {
		t.Fatal(err)
	}
	visible, _ := store.ListDispositionContexts(ctx, "issue:security:e")
	if len(visible) != 1 || !visible[0].AnnotationVisible || visible[0].View != "list" {
		t.Fatalf("visible %+v", visible)
	}
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
		return scanRows(rows, func(rows *sql.Rows) error {
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
			return nil
		})
	})
	return out, err
}

// ListRankings returns the ranking records of one run in rank order.
func (s *Store) ListRankings(ctx context.Context, runID string) ([]domain.RankingRecord, error) {
	out := []domain.RankingRecord{}
	err := s.withRead(ctx, func(db *sql.DB) error {
		rows, err := db.QueryContext(ctx, `SELECT payload FROM ranking_records WHERE run_id = ?`, runID)
		if err != nil {
			return fmt.Errorf("query rankings: %w", err)
		}
		return scanRows(rows, func(rows *sql.Rows) error {
			var payload []byte
			if err := rows.Scan(&payload); err != nil {
				return fmt.Errorf("scan ranking: %w", err)
			}
			var record domain.RankingRecord
			if err := decode(payload, &record); err != nil {
				return err
			}
			out = append(out, record)
			return nil
		})
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

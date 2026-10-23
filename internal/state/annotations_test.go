package state

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"scout/internal/domain"
)

func annotationFor(issue, version string) domain.Annotation {
	return domain.Annotation{
		AnnotationID: "annotation:" + issue + version, RunID: "run-1", IssueID: issue, UnitKey: issue, QuestionID: "triage.real_defect",
		QuestionVersion: version, Kind: "noul", ModelRequested: "typesafe/jev-1.14-20261014", ModelResolved: "typesafe/jev-1.14-20261014",
		Probability: 0.83, Answer: json.RawMessage(`{"type":"noul","noul":0.83}`), RequestBody: json.RawMessage(`{"model":"m","state":{"excerpt":["x"]}}`),
		OptionOrder: []string{"true", "false"},
	}
}

func TestSchemaV8StoresAnnotationsRankingsAndAuditDraws(t *testing.T) {
	ctx := context.Background()
	store := seedLedger(t, ctx, t.TempDir())
	defer store.Close()
	var version int
	if err := store.db.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&version); err != nil || version < 8 {
		t.Fatalf("schema version %d err %v", version, err)
	}
	issue := "issue:security:a"
	if err := store.RecordAnnotation(ctx, annotationFor(issue, "triage.real_defect@old")); err != nil {
		t.Fatal(err)
	}
	if err := store.RecordAnnotation(ctx, annotationFor(issue, "triage.real_defect@new")); err != nil {
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
	for _, row := range rows {
		wantStale := strings.HasSuffix(row.QuestionVersion, "@old")
		if row.Stale != wantStale || row.Probability != 0.83 || len(row.RequestBody) == 0 || row.ModelResolved == "" {
			t.Fatalf("row %+v", row)
		}
	}
	// An annotation carrying a credential is refused, never persisted.
	unsafe := annotationFor(issue, "triage.real_defect@new")
	unsafe.AnnotationID = "annotation:unsafe"
	unsafe.RequestBody = json.RawMessage(`{"state":{"excerpt":["api_key = abcdef123456789"]}}`)
	if err := store.RecordAnnotation(ctx, unsafe); err == nil {
		t.Fatal("credential-bearing request body must be refused")
	}

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
	defer store.Close()
	issue := "issue:security:a"
	if err := store.RecordRanking(ctx, domain.RankingRecord{RunID: "run-1", IssueID: issue, Mode: "jev", BaselineVersion: "base@1", RankFormulaVersion: "formula@1", CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.TriageIssue(WithDispositionView(ctx, domain.ViewAudit, false), issue, domain.IssueAccepted, "owner", "known", "maintainer", "reviewed blind"); err != nil {
		t.Fatal(err)
	}
	contexts, err := store.ListDispositionContexts(ctx, issue)
	if err != nil || len(contexts) != 1 {
		t.Fatalf("contexts %+v err %v", contexts, err)
	}
	got := contexts[0]
	if got.View != domain.ViewAudit || got.AnnotationVisible || got.RankFormulaVersion != "formula@1" || got.BaselineVersion != "base@1" {
		t.Fatalf("context %+v", got)
	}
	// An untagged disposition is stored as the CLI view with the annotation hidden.
	if _, err := store.TriageIssue(ctx, "issue:security:c", domain.IssueAccepted, "o", "known", "maintainer", "reviewed"); err != nil {
		t.Fatal(err)
	}
	untagged, err := store.ListDispositionContexts(ctx, "issue:security:c")
	if err != nil || len(untagged) != 1 || untagged[0].View != domain.ViewCLI || untagged[0].AnnotationVisible {
		t.Fatalf("untagged %+v err %v", untagged, err)
	}
	if _, err := store.TriageIssue(WithDispositionView(ctx, "list", true), "issue:security:e", domain.IssueAccepted, "o", "known", "maintainer", "reviewed"); err != nil {
		t.Fatal(err)
	}
	visible, _ := store.ListDispositionContexts(ctx, "issue:security:e")
	if len(visible) != 1 || !visible[0].AnnotationVisible || visible[0].View != "list" {
		t.Fatalf("visible %+v", visible)
	}
}

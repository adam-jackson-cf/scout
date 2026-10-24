package state

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"scout/internal/domain"
)

// seedLedger builds a store holding issues in every interesting condition:
// plain open, false-positive disposition, regressed, ambiguous, evidence drift
// and resolved, plus would_block rows (one maintainer-triaged).
func seedLedger(t *testing.T, ctx context.Context, root string) *Store {
	t.Helper()
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	snapshot := domain.Snapshot{RepositoryID: "repository", Revision: "abc123", WorktreeDigest: "worktree", IgnoreRulesDigest: "ignore", ExclusionsDigest: "exclude"}
	scope := domain.Scope{Kind: domain.ScopeRepository, SelectedFamilies: []domain.Family{domain.FamilySecurity}}
	plan := domain.Plan{PlanID: "plan-1", RunID: "run-1", Snapshot: snapshot, Scope: scope, ToolDigest: "tool-digest", PolicyDigest: "policy-digest", CreatedAt: now}
	finished := now.Add(time.Minute)
	run := domain.Run{RunID: "run-1", PlanID: "plan-1", Snapshot: snapshot, Scope: scope, State: domain.RunComplete, StartedAt: now, FinishedAt: &finished}
	var observations []domain.Observation
	for _, id := range []string{"obs-1", "obs-2", "obs-3", "obs-4", "obs-5", "obs-6"} {
		observations = append(observations, legacyObservation(id, "run-1", id+".go", "", "eval("+id+")", now))
	}
	seedRun(t, ctx, store, plan, run, observations, nil)
	issue := func(id, observation string, state domain.IssueState, events ...domain.IssueHistoryEntry) domain.Issue {
		history := []domain.IssueHistoryEntry{{At: now, To: domain.IssueOpen, Event: domain.EventCreated, Reason: "first observed"}}
		return domain.Issue{
			IssueID: "issue:security:" + id, State: state, Family: domain.FamilySecurity, Category: "security:test",
			Title: "unsafe <path> & more", Mechanism: "unsafe path " + id,
			Anchor:           domain.IssueAnchor{Path: observation + ".go", RuleKey: "scanner\x00rule", Mechanism: "unsafe path", Fingerprint: "fp-" + id},
			OriginalSeverity: domain.SeverityHigh, EffectiveSeverity: domain.SeverityMedium,
			ObservationIDs: []string{observation}, Files: []string{observation + ".go"},
			History: append(history, events...), CreatedAt: now, UpdatedAt: now.Add(time.Duration(len(events)) * time.Hour),
		}
	}
	open := issue("a", "obs-1", domain.IssueOpen)
	falsePositive := issue("b", "obs-2", domain.IssueFalsePositive, domain.IssueHistoryEntry{At: now.Add(time.Hour), From: domain.IssueOpen, To: domain.IssueFalsePositive, Event: domain.EventTriaged, Actor: "maintainer", Reason: "test fixture"})
	falsePositive.DispositionReason, falsePositive.Owner = "generated code", "team-a"
	regressed := issue("c", "obs-3", domain.IssueRegressed,
		domain.IssueHistoryEntry{At: now.Add(time.Hour), From: domain.IssueOpen, To: domain.IssueResolved, Event: domain.EventResolved, Reason: "fixed"},
		domain.IssueHistoryEntry{At: now.Add(2 * time.Hour), From: domain.IssueResolved, To: domain.IssueRegressed, Event: domain.EventRegressed, Reason: "reappeared"})
	ambiguous := issue("d", "obs-4", domain.IssueOpen, domain.IssueHistoryEntry{At: now.Add(time.Hour), To: domain.IssueOpen, Event: domain.EventAmbiguous, Reason: "two candidates"})
	ambiguous.AmbiguousWith, ambiguous.Uncertainty = []string{open.IssueID}, "could be a"
	drift := issue("e", "obs-5", domain.IssueOpen, domain.IssueHistoryEntry{At: now.Add(time.Hour), To: domain.IssueOpen, Event: domain.EventEvidenceDrift, Reason: "quote changed"})
	drift.Claim, drift.Rule = domain.ClaimDescriptive, "delta.cc-threshold@1"
	resolved := issue("f", "obs-6", domain.IssueResolved)
	resolved.Verification = &domain.Verification{RunID: "run-1", Relevant: true, InspectionStatus: domain.CoverageInspected, FailureModeAbsent: true, VerifiedAt: finished}
	for _, i := range []domain.Issue{open, falsePositive, regressed, ambiguous, drift, resolved} {
		if err := store.CreateIssue(ctx, i); err != nil {
			t.Fatalf("create %s: %v", i.IssueID, err)
		}
	}
	for _, row := range []domain.WouldBlock{
		{ID: "wb-1", HeadRevision: "h1", BaseRevision: "b1", IssueID: open.IssueID, CheckID: "guardrails.SUPPRESS.added.unscoped", RuleVersion: "1", Path: "a.go", Attribution: domain.Attribution("introduced"), Profile: "report", RecordedAt: now},
		{ID: "wb-2", HeadRevision: "h2", BaseRevision: "b2", IssueID: drift.IssueID, CheckID: "secrets.PAT", RuleVersion: "1", Path: "e.go", Attribution: domain.Attribution("introduced"), Profile: "report", RecordedAt: now},
	} {
		if err := store.RecordWouldBlock(ctx, row); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.TriageWouldBlock(ctx, "wb-1", domain.TriageWarranted, "maintainer", "real"); err != nil {
		t.Fatal(err)
	}
	if err := store.RecordSignals(ctx, testSignals()); err != nil {
		t.Fatal(err)
	}
	return store
}

func testSignals() []domain.Signal {
	base := domain.Signal{Version: 1, ScopeKind: domain.SignalScopeDirectory, ScopeAnchor: "pkg/sub", Snapshot: "abc123", PopulationDigest: "pop", ToolDigest: "lizard-digest", ConfigDigest: "config", Claim: domain.ClaimDescriptive}
	cc := base
	cc.MetricID, cc.Numerator, cc.Denominator, cc.Value = "scout.cc", domain.Float(1), domain.Float(2), domain.Float(1)
	cc.Detail = map[string]any{"cc_p90": 12, "cc_max": 12, "note": "a < b & c", "nested": map[string]any{"z": 1.5, "a": []any{1, "x"}}}
	crap := base
	crap.MetricID, crap.Missing, crap.MissingReason = "scout.crap", domain.SignalUnmeasured, "no coverage artifact"
	history := base
	history.MetricID, history.ScopeKind, history.ScopeAnchor, history.Window = "scout.changes", domain.SignalScopeFeature, "feature:f1", "2026-07-01/2026-09-29"
	history.Numerator, history.Denominator, history.Value = domain.Float(4), domain.Float(1), domain.Float(4)
	return []domain.Signal{crap, cc, history}
}

func TestExportImportExportIsByteIdentical(t *testing.T) {
	ctx := context.Background()
	source := seedLedger(t, ctx, t.TempDir())
	defer source.Close()
	first, err := source.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	again, err := source.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for path, data := range first {
		if !bytes.Equal(data, again[path]) {
			t.Fatalf("export of %s is not deterministic", path)
		}
		if bytes.Contains(data, []byte("\r")) || (len(data) > 0 && data[len(data)-1] != '\n') {
			t.Fatalf("%s is not LF-terminated JSON lines", path)
		}
	}
	if got := len(splitLines(first[ExportIssues])); got != 6 {
		t.Fatalf("issues.jsonl has %d rows, want 6", got)
	}
	if got := len(splitLines(first[ExportWouldBlock])); got != 2 {
		t.Fatalf("would_block.jsonl has %d rows, want 2", got)
	}
	if bytes.Contains(first[ExportIssues], []byte(`\u003c`)) {
		t.Fatal("export must not HTML-escape")
	}

	target, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	if err := target.Import(ctx, first); err != nil {
		t.Fatal(err)
	}
	second, err := target.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != len(first) {
		t.Fatalf("re-export has %d files, want %d", len(second), len(first))
	}
	for path, data := range first {
		if !bytes.Equal(data, second[path]) {
			t.Fatalf("re-export of %s differs:\n%s\n---\n%s", path, data, second[path])
		}
	}

	// Dispositions and audit history survive the round trip.
	want, err := source.ListIssues(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, issue := range want {
		got, err := target.LoadIssue(ctx, issue.IssueID)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got.History, issue.History) || got.State != issue.State || got.DispositionReason != issue.DispositionReason ||
			got.Owner != issue.Owner || !reflect.DeepEqual(got.AmbiguousWith, issue.AmbiguousWith) || !reflect.DeepEqual(got.Verification, issue.Verification) {
			t.Fatalf("issue %s did not survive import:\nwant %+v\ngot  %+v", issue.IssueID, issue, got)
		}
	}
	rows, err := target.ListWouldBlock(ctx)
	if err != nil {
		t.Fatal(err)
	}
	triaged := map[string]string{}
	for _, row := range rows {
		triaged[row.ID] = row.Triage
	}
	if triaged["wb-1"] != domain.TriageWarranted || triaged["wb-2"] != domain.TriageUndecided {
		t.Fatalf("would_block triage not preserved: %v", triaged)
	}
}

// Regression: a restored baseline carries issues whose observations are not in
// the fresh ledger; updating such an issue (as correlation does on every
// assessment) must keep the opaque references instead of failing, while a
// genuinely unknown new observation is still refused.
func TestImportedIssueUpdatesKeepOpaqueObservationReferences(t *testing.T) {
	ctx := context.Background()
	source := seedLedger(t, ctx, t.TempDir())
	defer source.Close()
	files, err := source.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	target, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	if err := target.Import(ctx, files); err != nil {
		t.Fatal(err)
	}
	id := "issue:security:a"
	if _, err := target.TriageIssue(ctx, id, domain.IssueWontFix, "team-a", "accepted", "maintainer", "triage after restore"); err != nil {
		t.Fatalf("update of imported issue: %v", err)
	}
	got, err := target.LoadIssue(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.ObservationIDs, []string{"obs-1"}) || got.State != domain.IssueWontFix {
		t.Fatalf("imported reference lost: %+v", got)
	}
	got.ObservationIDs = append(got.ObservationIDs, "never-seen")
	if err := target.saveIssue(ctx, got, false); err == nil {
		t.Fatal("a new unknown observation must still be refused")
	}
}

func TestExportManifestCarriesBaseIdentity(t *testing.T) {
	ctx := context.Background()
	store := seedLedger(t, ctx, t.TempDir())
	defer store.Close()
	files, err := store.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	meta, err := ReadManifestMeta(files[ExportManifest])
	if err != nil {
		t.Fatal(err)
	}
	if meta.Revision != "abc123" || meta.ToolDigest != "tool-digest" || meta.PolicyDigest != "policy-digest" || meta.RepositoryID != "repository" || meta.WouldBlock != WouldBlockPresent {
		t.Fatalf("unexpected manifest meta %+v", meta)
	}
}

func TestImportRefusesNonFreshStoreAndTamperedExport(t *testing.T) {
	ctx := context.Background()
	source := seedLedger(t, ctx, t.TempDir())
	defer source.Close()
	files, err := source.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := source.Import(ctx, files); !errors.Is(err, ErrImportNotEmpty) {
		t.Fatalf("import into populated store = %v, want ErrImportNotEmpty", err)
	}
	tampered := map[string][]byte{}
	for path, data := range files {
		tampered[path] = data
	}
	tampered[ExportIssues] = append(append([]byte{}, files[ExportIssues]...), []byte(`{"issueId":"issue:security:x"}`+"\n")...)
	fresh, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer fresh.Close()
	if err := fresh.Import(ctx, tampered); err == nil {
		t.Fatal("import accepted an export that does not match its manifest")
	}
}

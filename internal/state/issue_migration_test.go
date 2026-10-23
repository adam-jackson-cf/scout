package state

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"scout/internal/domain"
)

func legacyID(fill string) string { return "issue:security:" + strings.Repeat(fill, 64) }

func legacyObservation(id, runID, path, symbol, quote string, observed time.Time) domain.Observation {
	return domain.Observation{
		ObservationID: id, RunID: runID, Source: "scanner", SourceFindingID: id, Family: domain.FamilySecurity,
		OriginalCategory: "security:test", RuleID: "rule", RuleVersion: "1", SourceSeverity: domain.SeverityHigh,
		EffectiveSeverity: domain.SeverityHigh, Confidence: domain.ConfidenceHigh, Mechanism: "unsafe path",
		Evidence:  []domain.Evidence{{Location: domain.SourceRange{Path: path, StartLine: 3, EndLine: 3, Symbol: symbol}, RedactedQuote: quote}},
		Validated: true, ObservedAt: observed,
	}
}

// A v1-era issue store (hash IDs, approval fields and states, feature-merged
// issues) is re-anchored by the v4→v5 migration: every row gets a durable ID and
// anchor, keeps its legacy ID, loses approval data, and ambiguous rows are
// listed instead of merged.
func TestV4IssueStoreIsReanchoredAndAmbiguousRowsListed(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Millisecond)
	snapshot := domain.Snapshot{RepositoryID: "repository", WorktreeDigest: "worktree", IgnoreRulesDigest: "ignore", ExclusionsDigest: "exclude"}
	scope := domain.Scope{Kind: domain.ScopeRepository, SelectedFamilies: []domain.Family{domain.FamilySecurity}}
	plan := domain.Plan{PlanID: "plan-1", RunID: "run-1", Snapshot: snapshot, Scope: scope, CreatedAt: now}
	run := domain.Run{RunID: "run-1", PlanID: "plan-1", Snapshot: snapshot, Scope: scope, State: domain.RunPartial, StartedAt: now}
	observations := []domain.Observation{
		legacyObservation("obs-a", "run-1", "a.go", "", "eval(a)", now),
		legacyObservation("obs-b", "run-1", "b.go", "pkg.F", "eval(b)", now),
		legacyObservation("obs-c", "run-1", "b.go", "pkg.F", "eval(b)", now),
		legacyObservation("obs-d1", "run-1", "d.go", "", "eval(d1)", now),
		legacyObservation("obs-d2", "run-1", "z.go", "", "eval(d2)", now),
		legacyObservation("obs-e", "run-1", "e.go", "", "eval(e)", now),
	}
	if err := store.SavePlan(ctx, plan); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveRun(ctx, run, observations); err != nil {
		t.Fatal(err)
	}
	legacyRows := []struct {
		id, state string
		observed  []string
		approved  bool
	}{
		{legacyID("a"), "open", []string{"obs-a"}, false},
		{legacyID("b"), "open", []string{"obs-b"}, false},
		{legacyID("c"), "open", []string{"obs-c"}, false},
		{legacyID("d"), "open", []string{"obs-d1", "obs-d2"}, false},
		{legacyID("e"), "pending_verification", []string{"obs-e"}, true},
	}
	for _, row := range legacyRows {
		payload := map[string]any{
			"issueId": row.id, "state": row.state, "title": "security:test: unsafe path", "mechanism": "unsafe path",
			"originalSeverity": "high", "effectiveSeverity": "high", "observationIds": row.observed, "files": []string{"a.go"},
			"history":   []any{map[string]any{"at": now, "to": row.state, "actor": "legacy", "reason": "legacy history"}},
			"createdAt": now, "updatedAt": now,
		}
		if row.approved {
			payload["approvalRef"], payload["proposedAction"] = "change:123", "replace unsafe path"
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := store.db.ExecContext(ctx, `INSERT INTO issues(issue_id, state, updated_at, payload) VALUES (?, ?, ?, ?)`, row.id, row.state, unixMillis(now), encoded); err != nil {
			t.Fatal(err)
		}
		for _, observationID := range row.observed {
			if _, err := store.db.ExecContext(ctx, `INSERT INTO issue_observations(issue_id, observation_id) VALUES (?, ?)`, row.id, observationID); err != nil {
				t.Fatal(err)
			}
		}
		history, _ := json.Marshal(payload["history"].([]any)[0])
		if _, err := store.db.ExecContext(ctx, `INSERT INTO issue_history(issue_id, ordinal, payload) VALUES (?, 0, ?)`, row.id, history); err != nil {
			t.Fatal(err)
		}
	}
	for _, statement := range []string{`DROP TABLE issue_migration_ambiguities`, `PRAGMA user_version=4`} {
		if _, err := store.db.ExecContext(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	migrated, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer migrated.Close()
	var version int
	if err := migrated.db.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&version); err != nil || version != schemaVersion {
		t.Fatalf("schema version = %d, %v", version, err)
	}
	issues, err := migrated.ListIssues(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != len(legacyRows) {
		t.Fatalf("migrated %d issues, want %d", len(issues), len(legacyRows))
	}
	byLegacy := map[string]domain.Issue{}
	for _, issue := range issues {
		if !strings.HasPrefix(issue.IssueID, "issue:security:") || len(issue.IssueID) != len("issue:security:")+26 {
			t.Fatalf("issue was not given a durable ULID ID: %q", issue.IssueID)
		}
		byLegacy[issue.LegacyID] = issue
		if _, err := migrated.LoadIssue(ctx, issue.LegacyID); !errors.Is(err, ErrNotFound) {
			t.Fatalf("legacy ID still addresses an issue: %v", err)
		}
		linked, err := migrated.ListIssueObservations(ctx, issue.IssueID)
		if err != nil || len(linked) != len(issue.ObservationIDs) {
			t.Fatalf("observation links lost for %s: %v", issue.IssueID, err)
		}
		if issue.History[len(issue.History)-1].Event == "" && len(issue.History) < 2 {
			t.Fatalf("migration left no audit entry: %#v", issue.History)
		}
	}
	clean := byLegacy[legacyID("a")]
	if clean.Anchor.Path != "a.go" || clean.Anchor.RuleKey != "scanner:rule" || clean.Anchor.Fingerprint == "" || clean.Family != domain.FamilySecurity || clean.Category != "security:test" {
		t.Fatalf("issue was not re-anchored from its observation: %#v", clean.Anchor)
	}
	if approved := byLegacy[legacyID("e")]; approved.State != domain.IssueOpen {
		t.Fatalf("approval state survived migration: %q", approved.State)
	}
	var stored string
	if err := migrated.db.QueryRowContext(ctx, `SELECT payload FROM issues WHERE issue_id = ?`, byLegacy[legacyID("e")].IssueID).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stored, "approvalRef") || strings.Contains(stored, "proposedAction") || strings.Contains(stored, "pending_verification") {
		t.Fatalf("approval data survived migration: %s", stored)
	}

	ambiguities, err := migrated.ListMigrationAmbiguities(ctx)
	if err != nil {
		t.Fatal(err)
	}
	reasons := map[string]string{}
	for _, row := range ambiguities {
		reasons[row.LegacyID] = row.Reason
		if row.IssueID != byLegacy[row.LegacyID].IssueID {
			t.Fatalf("ambiguity row points at the wrong issue: %#v", row)
		}
	}
	if len(reasons) != 3 || !strings.Contains(reasons[legacyID("b")], "indistinguishable") || !strings.Contains(reasons[legacyID("c")], "indistinguishable") || !strings.Contains(reasons[legacyID("d")], "merged") {
		t.Fatalf("ambiguous rows = %#v", reasons)
	}
	b, c := byLegacy[legacyID("b")], byLegacy[legacyID("c")]
	if len(b.AmbiguousWith) != 1 || b.AmbiguousWith[0] != c.IssueID || len(c.AmbiguousWith) != 1 || c.AmbiguousWith[0] != b.IssueID || b.State != domain.IssueOpen {
		t.Fatalf("indistinguishable issues were not linked and kept open: %#v %#v", b.AmbiguousWith, c.AmbiguousWith)
	}
}

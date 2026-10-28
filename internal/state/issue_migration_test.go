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

// A v1 issue store (user_version 3, built from the v1 DDL: hash IDs, approval
// fields and states, feature-merged issues) is re-anchored on open: every row
// gets a durable ID and anchor, keeps its legacy ID, loses approval data, and
// ambiguous rows are listed instead of merged.
func TestV1IssueStoreIsReanchoredAndAmbiguousRowsListed(t *testing.T) {
	ctx := context.Background()
	v1 := newV1Store(t)
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
	v1.insertRun(t, plan, run, observations)
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
		if _, err := v1.db.ExecContext(ctx, `INSERT INTO issues(issue_id, state, updated_at, payload) VALUES (?, ?, ?, ?)`, row.id, row.state, unixMillis(now), encoded); err != nil {
			t.Fatal(err)
		}
		for _, observationID := range row.observed {
			if _, err := v1.db.ExecContext(ctx, `INSERT INTO issue_observations(issue_id, observation_id) VALUES (?, ?)`, row.id, observationID); err != nil {
				t.Fatal(err)
			}
		}
		history, _ := json.Marshal(payload["history"].([]any)[0])
		if _, err := v1.db.ExecContext(ctx, `INSERT INTO issue_history(issue_id, ordinal, payload) VALUES (?, 0, ?)`, row.id, history); err != nil {
			t.Fatal(err)
		}
	}
	v1.close(t)
	root := v1.root

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

	ambiguityRows, err := migrated.db.QueryContext(ctx, `SELECT issue_id, legacy_id, reason FROM issue_migration_ambiguities`)
	if err != nil {
		t.Fatal(err)
	}
	defer ambiguityRows.Close()
	reasons := map[string]string{}
	for ambiguityRows.Next() {
		var issueID, legacy, reason string
		if err := ambiguityRows.Scan(&issueID, &legacy, &reason); err != nil {
			t.Fatal(err)
		}
		reasons[legacy] = reason
		if issueID != byLegacy[legacy].IssueID {
			t.Fatalf("ambiguity row points at the wrong issue: %s %s", issueID, legacy)
		}
	}
	if err := ambiguityRows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(reasons) != 3 || !strings.Contains(reasons[legacyID("b")], "indistinguishable") || !strings.Contains(reasons[legacyID("c")], "indistinguishable") || !strings.Contains(reasons[legacyID("d")], "merged") {
		t.Fatalf("ambiguous rows = %#v", reasons)
	}
	b, c := byLegacy[legacyID("b")], byLegacy[legacyID("c")]
	if len(b.AmbiguousWith) != 1 || b.AmbiguousWith[0] != c.IssueID || len(c.AmbiguousWith) != 1 || c.AmbiguousWith[0] != b.IssueID || b.State != domain.IssueOpen {
		t.Fatalf("indistinguishable issues were not linked and kept open: %#v %#v", b.AmbiguousWith, c.AmbiguousWith)
	}
}

// legacyStoreWithRow returns the root of a v1 store (built from the v1 DDL)
// holding one legacy issue row, optionally linked to a real observation.
func legacyStoreWithRow(t *testing.T, id string, history any, link bool) string {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Millisecond)
	v1 := newV1Store(t)
	snapshot := domain.Snapshot{RepositoryID: "repository", WorktreeDigest: "worktree", IgnoreRulesDigest: "ignore", ExclusionsDigest: "exclude"}
	scope := domain.Scope{Kind: domain.ScopeRepository, SelectedFamilies: []domain.Family{domain.FamilySecurity}}
	observation := legacyObservation("obs-a", "run-1", "a.go", "", "eval(a)", now)
	v1.insertRun(t, domain.Plan{PlanID: "plan-1", RunID: "run-1", Snapshot: snapshot, Scope: scope, CreatedAt: now}, domain.Run{RunID: "run-1", PlanID: "plan-1", Snapshot: snapshot, Scope: scope, State: domain.RunPartial, StartedAt: now}, []domain.Observation{observation})
	payload, err := json.Marshal(map[string]any{
		"issueId": id, "state": "open", "title": "security:test: unsafe path", "mechanism": "unsafe path",
		"originalSeverity": "high", "effectiveSeverity": "high", "observationIds": []string{"obs-a"},
		"history": history, "createdAt": now, "updatedAt": now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v1.db.Exec(`INSERT INTO issues(issue_id, state, updated_at, payload) VALUES (?, 'open', ?, ?)`, id, unixMillis(now), payload); err != nil {
		t.Fatal(err)
	}
	if link {
		if _, err := v1.db.Exec(`INSERT INTO issue_observations(issue_id, observation_id) VALUES (?, 'obs-a')`, id); err != nil {
			t.Fatal(err)
		}
	}
	v1.close(t)
	return v1.root
}

// A legacy row whose history is corrupt cannot be re-anchored faithfully: the
// migration fails and names the legacy ID.
func TestCorruptLegacyHistoryFailsMigrationNamingItsID(t *testing.T) {
	id := legacyID("a")
	root := legacyStoreWithRow(t, id, []any{"corrupt"}, true)
	if migrated, err := Open(root); err == nil {
		migrated.Close()
		t.Fatal("migration accepted a corrupt legacy history entry")
	} else if !strings.Contains(err.Error(), id) {
		t.Fatalf("migration error does not name the legacy ID: %v", err)
	}
}

// A legacy row that cannot be anchored, or whose family cannot be derived,
// does not fail the store open: it is listed as ambiguous and never given a
// guessed family.
func TestUnanchorableLegacyRowIsListedAsAmbiguousAndStoreOpens(t *testing.T) {
	history := []any{map[string]any{"at": time.Now().UTC(), "to": "open", "actor": "legacy", "reason": "legacy history"}}
	for name, id := range map[string]string{
		"valid family, no observations": legacyID("a"),
		"unknown family":                "issue:bogus:" + strings.Repeat("b", 64),
	} {
		t.Run(name, func(t *testing.T) {
			root := legacyStoreWithRow(t, id, history, false)
			migrated, err := Open(root)
			if err != nil {
				t.Fatalf("unanchorable row failed the store open: %v", err)
			}
			defer migrated.Close()
			var issueID, reason string
			if err := migrated.db.QueryRowContext(context.Background(), `SELECT issue_id, reason FROM issue_migration_ambiguities WHERE legacy_id = ?`, id).Scan(&issueID, &reason); err != nil {
				t.Fatalf("row was not listed as ambiguous: %v", err)
			}
			issue, err := migrated.LoadIssue(context.Background(), issueID)
			if err != nil {
				t.Fatal(err)
			}
			if issue.LegacyID != id || issue.Anchor.RuleKey != "" {
				t.Fatalf("unanchorable row was given an anchor or lost its legacy ID: %#v", issue)
			}
			if id == legacyID("a") && issue.IssueID == id {
				t.Fatal("row with a valid family kept its legacy ID instead of a durable one")
			}
			if strings.Contains(id, "bogus") && issue.Family.Valid() {
				t.Fatalf("row with an unknown family was given family %q", issue.Family)
			}
		})
	}
}

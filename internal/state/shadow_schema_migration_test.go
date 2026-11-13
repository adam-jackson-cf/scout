package state

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"scout/internal/domain"
)

var adr012Time = time.Date(2026, 11, 8, 12, 0, 0, 123000000, time.UTC)

func adr012V13(t *testing.T) (string, *sql.DB) {
	t.Helper()
	root := t.TempDir()
	db := openFixtureDB(t, root, "delete")
	for _, statement := range adr012V13DDL {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`PRAGMA user_version=13`); err != nil {
		t.Fatal(err)
	}
	return root, db
}

func adr012Open(t *testing.T, root string) *Store {
	t.Helper()
	s, err := Open(root)
	if err != nil {
		t.Fatalf("missing schema14 writable migration: %v", err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	return s
}

func adr012Ledger(t *testing.T, s *Store) domain.ShadowLedger {
	t.Helper()
	ledger, err := s.ReadShadowLedger(context.Background())
	if err != nil {
		t.Fatalf("missing immutable scoped ledger read: %v", err)
	}
	return ledger
}

func adr012Record(t *testing.T, s *Store, f domain.PRRunFact, events []domain.ProviderPREvent, links []domain.PRRunEventLink) {
	t.Helper()
	if err := s.RecordPRRun(context.Background(), f, events, links); err != nil {
		t.Fatalf("missing immutable attempt/event persistence: %v", err)
	}
}

func adr012Fact(scope domain.ShadowScope, run string) domain.PRRunFact {
	return domain.PRRunFact{
		ID:    domain.AttemptID{RepositoryID: "repository", RunID: run, JobID: "999", Attempt: "1000"},
		Scope: scope, PRNumber: "17", HeadCommit: strings.Repeat("a", 40), HeadTree: strings.Repeat("b", 40),
		BaseCommit: strings.Repeat("c", 40), MergeBase: strings.Repeat("d", 40), AssessedRevision: strings.Repeat("e", 40),
		CheckID: "security:test", RuleVersion: "1", CreatedAt: adr012Time.Add(-time.Minute),
		StartedAt: adr012Time, FinishedAt: adr012Time.Add(time.Minute), Status: "assessed", Comparison: "comparable", Inspection: "inspected",
		Pins: domain.ShadowPins{Scout: "scout-1", Tools: "tools-1", Catalogue: "catalogue-1", Policy: "policy-1", Recipe: "guardian-1"},
	}
}

func adr012Event(f domain.PRRunFact, id string) domain.ProviderPREvent {
	return domain.ProviderPREvent{EventID: id, RepositoryID: f.ID.RepositoryID, Scope: f.Scope, PRNumber: f.PRNumber,
		HeadCommit: f.HeadCommit, HeadTree: f.HeadTree, MergeBase: f.MergeBase, CheckID: f.CheckID,
		IssueID: "issue:security:01ARZ3NDEKTSV4RRFFQ69G5FAV", RuleVersion: f.RuleVersion, EventAt: f.StartedAt}
}

func adr012Link(f domain.PRRunFact, event domain.ProviderPREvent) domain.PRRunEventLink {
	return domain.PRRunEventLink{AttemptID: f.ID, Scope: f.Scope, EventID: event.EventID}
}

func adr012Unchanged(t *testing.T, s *Store, before domain.ShadowLedger) {
	t.Helper()
	if after := adr012Ledger(t, s); !reflect.DeepEqual(before, after) {
		t.Fatalf("missing atomic rejection: ledger changed\nbefore=%#v\nafter=%#v", before, after)
	}
}

func adr012SeedV13(t *testing.T, db *sql.DB) []byte {
	t.Helper()
	// Deliberately retain formatting: migration must not decode/re-encode old rows.
	issue := []byte(`{ "issueId": "issue:security:01ARZ3NDEKTSV4RRFFQ69G5FAV", "state":"open", "title":"unpublished" }`)
	if _, err := db.Exec(`INSERT INTO issues VALUES (?, 'open', ?, ?)`, "issue:security:01ARZ3NDEKTSV4RRFFQ69G5FAV", adr012Time.UnixMilli(), issue); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO would_block(id,head_revision,base_revision,issue_id,check_id,rule_version,path,attribution,profile,triage,triage_actor,triage_reason,recorded_at,triaged_at) VALUES ('legacy-event','head','base','old-issue','security:test','1','a.go','introduced','guardian','warranted','original-actor','original reason',?,?)`, adr012Time.Add(-time.Hour).UnixMilli(), adr012Time.UnixMilli()); err != nil {
		t.Fatal(err)
	}
	return issue
}

func TestADR012Schema14MigrationParity(t *testing.T) {
	t.Run("fresh_v3_v13_full_schema_parity", func(t *testing.T) {
		fresh := adr012Open(t, t.TempDir())
		v3 := newV1Store(t)
		v3.close(t)
		v13Root, db := adr012V13(t)
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		want := describeSchema(t, fresh.db)
		stores := []*Store{fresh, adr012Open(t, v3.root), adr012Open(t, v13Root)}
		for i, name := range []string{"fresh", "v3", "v13"} {
			s := stores[i]
			var version int
			if err := s.db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
				t.Fatal(err)
			}
			if version != 14 {
				t.Errorf("missing schema14 migration for %s: user_version=%d", name, version)
			}
			if got := describeSchema(t, s.db); !slices.Equal(want, got) {
				t.Errorf("missing fresh/v3/v13 ordered-column/index uniqueness/origin/partial/collation/FK/object parity for %s:\nwant %v\ngot %v", name, want, got)
			}
			for _, table := range []string{"pr_run_facts", "pr_run_events", "would_block_triage"} {
				var count int
				if err := s.db.QueryRow(`SELECT count(*) FROM sqlite_schema WHERE type='table' AND name=?`, table).Scan(&count); err != nil {
					t.Fatal(err)
				}
				if count != 1 {
					t.Errorf("missing schema14 %s table in %s", table, name)
				}
			}
		}
	})

	t.Run("v3_legacy_ids_ambiguities_approval_removal", func(t *testing.T) {
		v3 := newV1Store(t)
		snapshot := domain.Snapshot{RepositoryID: "repository", WorktreeDigest: "tree", IgnoreRulesDigest: "ignore", ExclusionsDigest: "exclude"}
		scope := domain.Scope{Kind: domain.ScopeRepository, SelectedFamilies: []domain.Family{domain.FamilySecurity}}
		observations := []domain.Observation{legacyObservation("obs-a", "run", "a.go", "pkg.F", "eval(a)", adr012Time), legacyObservation("obs-b", "run", "a.go", "pkg.F", "eval(a)", adr012Time)}
		v3.insertRun(t, domain.Plan{PlanID: "plan", RunID: "run", Snapshot: snapshot, Scope: scope, CreatedAt: adr012Time}, domain.Run{RunID: "run", PlanID: "plan", Snapshot: snapshot, Scope: scope, State: domain.RunPartial, StartedAt: adr012Time}, observations)
		for i, observation := range observations {
			id := legacyID([]string{"a", "b"}[i])
			payload, err := json.Marshal(map[string]any{"issueId": id, "state": "pending_verification", "title": "unsafe", "mechanism": "unsafe", "originalSeverity": "high", "effectiveSeverity": "high", "observationIds": []string{observation.ObservationID}, "createdAt": adr012Time, "updatedAt": adr012Time, "approvalRef": "change:123", "proposedAction": "replace"})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := v3.db.Exec(`INSERT INTO issues VALUES (?, 'pending_verification', ?, ?)`, id, adr012Time.UnixMilli(), payload); err != nil {
				t.Fatal(err)
			}
			if _, err := v3.db.Exec(`INSERT INTO issue_observations VALUES (?, ?)`, id, observation.ObservationID); err != nil {
				t.Fatal(err)
			}
		}
		v3.close(t)
		s := adr012Open(t, v3.root)
		issues, err := s.ListIssues(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(issues) != 2 {
			t.Fatalf("missing v3 re-anchor: %#v", issues)
		}
		for _, issue := range issues {
			if issue.IssueID == issue.LegacyID || (issue.LegacyID != legacyID("a") && issue.LegacyID != legacyID("b")) || issue.Anchor.Path != "a.go" || issue.Anchor.Fingerprint == "" || issue.State != domain.IssueOpen || len(issue.AmbiguousWith) != 1 {
				t.Errorf("missing legacy ID/anchor/ambiguity/approval re-anchor: %#v", issue)
			}
			var raw string
			if err := s.db.QueryRow(`SELECT payload FROM issues WHERE issue_id=?`, issue.IssueID).Scan(&raw); err != nil {
				t.Fatal(err)
			}
			for _, obsolete := range []string{"approvalRef", "proposedAction", "pending_verification"} {
				if strings.Contains(raw, obsolete) {
					t.Errorf("missing approval removal: %s", raw)
				}
			}
		}
		var ambiguities int
		if err := s.db.QueryRow(`SELECT count(*) FROM issue_migration_ambiguities`).Scan(&ambiguities); err != nil {
			t.Fatal(err)
		}
		if ambiguities != 2 {
			t.Errorf("missing persistent v3 ambiguity rows: %d", ambiguities)
		}
		ledger := adr012Ledger(t, s)
		if len(ledger.Facts) != 0 || len(ledger.Frontiers) != 0 {
			t.Fatal("missing v3 unmeasured/legacy separation")
		}
	})

	t.Run("v13_preserves_rows_and_copies_triage_once", func(t *testing.T) {
		root, db := adr012V13(t)
		rawIssue := adr012SeedV13(t, db)
		before := adr012LegacyRows(t, db)
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		s := adr012Open(t, root)
		for open := range 2 {
			if got := adr012LegacyRows(t, s.db); !reflect.DeepEqual(before, got) {
				t.Fatalf("missing byte-preserving v13 issue/event migration: %#v != %#v", got, before)
			}
			var raw []byte
			if err := s.db.QueryRow(`SELECT payload FROM issues`).Scan(&raw); err != nil || !bytes.Equal(raw, rawIssue) {
				t.Fatalf("missing unpublished issue preservation: %q, %v", raw, err)
			}
			ledger := adr012Ledger(t, s)
			if len(ledger.Triage) != 1 {
				t.Fatalf("missing exactly-once v13 triage copy after open %d: %#v", open, ledger.Triage)
			}
			triage := ledger.Triage[0]
			if triage.EventID != "legacy-event" || triage.Scope != domain.ShadowScopeLegacy || triage.Verdict != "warranted" || triage.Actor != "original-actor" || triage.Reason != "original reason" || !triage.TriagedAt.Equal(adr012Time) || triage.RecordedAt != nil || triage.PolicyRevision != "" {
				t.Fatalf("missing original-time legacy-only triage provenance: %#v", triage)
			}
			if open == 0 {
				if err := s.Close(); err != nil {
					t.Fatal(err)
				}
				s = adr012Open(t, root)
			}
		}
	})

	t.Run("legacy_history_does_not_read_mutable_triage_columns", func(t *testing.T) {
		root, db := adr012V13(t)
		adr012SeedV13(t, db)
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		s := adr012Open(t, root)
		before := adr012Ledger(t, s)
		if _, err := s.db.Exec(`UPDATE would_block SET triage='not_warranted',triage_actor='untrusted',triage_reason='rewrite',triaged_at=?`, adr012Time.Add(time.Hour).UnixMilli()); err != nil {
			t.Fatal(err)
		}
		adr012Unchanged(t, s, before)
	})

	t.Run("migration_failure_rolls_back_schema_data_and_version", func(t *testing.T) {
		root, db := adr012V13(t)
		adr012SeedV13(t, db)
		// A hostile pre-existing view forces DDL failure without changing fixture DDL.
		if _, err := db.Exec(`CREATE VIEW would_block_triage AS SELECT 'collision' AS action_id`); err != nil {
			t.Fatal(err)
		}
		beforeSchema, beforeRows := describeSchema(t, db), adr012LegacyRows(t, db)
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		if s, err := Open(root); err == nil {
			s.Close()
			t.Fatal("missing transactional schema14 migration failure on conflicting object")
		}
		db = openFixtureDB(t, root, "wal")
		defer db.Close()
		var version int
		if err := db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
			t.Fatal(err)
		}
		if version != 13 || !slices.Equal(beforeSchema, describeSchema(t, db)) || !reflect.DeepEqual(beforeRows, adr012LegacyRows(t, db)) {
			t.Fatal("missing migration rollback of schema, rows, and user_version")
		}
	})
}

// Captures every old event column and issue payload, not a derived authorized view.
func adr012LegacyRows(t *testing.T, db *sql.DB) []string {
	t.Helper()
	var out []string
	for _, query := range []string{`SELECT issue_id,state,updated_at,hex(payload) FROM issues ORDER BY issue_id`, `SELECT * FROM would_block ORDER BY id`} {
		rows, err := db.Query(query)
		if err != nil {
			t.Fatal(err)
		}
		columns, err := rows.Columns()
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			values := make([]sql.NullString, len(columns))
			args := make([]any, len(columns))
			for i := range values {
				args[i] = &values[i]
			}
			if err := rows.Scan(args...); err != nil {
				t.Fatal(err)
			}
			out = append(out, fmt.Sprintf("%#v", values))
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		rows.Close()
	}
	return out
}

func TestADR012UpgradeRefusalPreflight(t *testing.T) {
	t.Run("unsupported_versions_untouched_before_lock_or_journal", func(t *testing.T) {
		for _, mode := range []string{"delete", "wal"} {
			for _, version := range []int{1, 2, 4, 5, 6, 7, 8, 9, 10, 11, 12, 15, 99} {
				t.Run(fmt.Sprintf("%s_v%d", mode, version), func(t *testing.T) {
					root := t.TempDir()
					path := writeVersionedDB(t, root, version, mode)
					before, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					for _, readonly := range []bool{false, true} {
						var s *Store
						if readonly {
							s, err = OpenExistingReadOnly(root)
						} else {
							s, err = Open(root)
						}
						if err == nil {
							s.Close()
							t.Fatalf("missing refusal of schema %d (readonly=%t)", version, readonly)
						}
						message := strings.ToLower(err.Error())
						if !strings.Contains(message, fmt.Sprintf("version %d", version)) {
							t.Errorf("missing refused-version diagnostic: %v", err)
						}
						if version > 14 && (!strings.Contains(message, "upgrade") || strings.Contains(message, "remove") || strings.Contains(message, "re-baseline")) {
							t.Errorf("missing newer-schema upgrade-writer remedy without deletion: %v", err)
						}
						if version < 14 && !strings.Contains(message, "re-baseline") {
							t.Errorf("missing REQ-46 unsupported-old remedy: %v", err)
						}
					}
					after, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(before, after) {
						t.Fatal("missing read-only preflight: refused open mutated database/journal mode")
					}
					// SQLite may create read-only WAL/SHM sidecars; no writer lock,
					// artifacts directory, or rollback journal is allowed.
					for _, name := range []string{"writer.lock", "artifacts", "state.sqlite-journal"} {
						if _, err := os.Stat(filepath.Join(root, ".scout", name)); !errors.Is(err, os.ErrNotExist) {
							t.Errorf("missing mutation-free refusal: %s exists (%v)", name, err)
						}
					}
				})
			}
		}
	})
	t.Run("query_only_v13_requires_nondestructive_writable_upgrade", func(t *testing.T) {
		root, db := adr012V13(t)
		adr012SeedV13(t, db)
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(root, ".scout", "state.sqlite")
		before, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if s, err := OpenExistingReadOnly(root); err == nil {
			s.Close()
			t.Error("missing query-only v13 writable-upgrade refusal")
		} else {
			message := strings.ToLower(err.Error())
			if !strings.Contains(message, "writable") || !strings.Contains(message, "upgrade") || strings.Contains(message, "remove") || strings.Contains(message, "re-baseline") {
				t.Errorf("missing nondestructive writable-upgrade remedy: %v", err)
			}
		}
		after, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before, after) {
			t.Error("missing query-only v13 byte preservation")
		}
		for _, name := range []string{"writer.lock", "artifacts"} {
			if _, err := os.Stat(filepath.Join(root, ".scout", name)); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("missing v13 read-only preflight: %s", name)
			}
		}
	})
	t.Run("writable_v13_keeps_unpublished_history_and_empty_is_unmeasured", func(t *testing.T) {
		root, db := adr012V13(t)
		adr012SeedV13(t, db)
		before := adr012LegacyRows(t, db)
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		s := adr012Open(t, root)
		if !reflect.DeepEqual(before, adr012LegacyRows(t, s.db)) {
			t.Error("missing unpublished v13 row preservation")
		}
		for _, store := range []*Store{s, adr012Open(t, t.TempDir())} {
			ledger := adr012Ledger(t, store)
			if len(ledger.Facts) != 0 || len(ledger.Frontiers) != 0 || len(ledger.Intervals) != 0 {
				t.Fatalf("missing no-facts unmeasured history: %#v", ledger)
			}
		}
	})
}

func TestADR012ImmutableScopeAndLinks(t *testing.T) {
	t.Run("zero_event_and_comparison_none_survive_reopen", func(t *testing.T) {
		root := t.TempDir()
		s := adr012Open(t, root)
		good := adr012Fact(domain.ShadowScopeCI, "999")
		failed := adr012Fact(domain.ShadowScopeCI, "1000")
		failed.Status = "failed"
		failed.Comparison = "none"
		failed.Inspection = "not_inspected"
		adr012Record(t, s, good, nil, nil)
		adr012Record(t, s, failed, nil, nil)
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		s = adr012Open(t, root)
		ledger := adr012Ledger(t, s)
		if len(ledger.Facts) != 2 || len(ledger.Events) != 0 || len(ledger.Links) != 0 {
			t.Fatalf("missing durable zero-event/failed attempts: %#v", ledger)
		}
		for _, want := range []domain.PRRunFact{good, failed} {
			if !slices.ContainsFunc(ledger.Facts, func(got domain.PRRunFact) bool { return reflect.DeepEqual(got, want) }) {
				t.Errorf("missing exact fact/pin/provider numeric identity persistence: %#v", want)
			}
		}
		if len(ledger.Frontiers) != 0 {
			t.Error("missing staged-only semantics: RecordPRRun fabricated completeness")
		}
	})

	t.Run("immutable_duplicates_and_many_to_many_links", func(t *testing.T) {
		s := adr012Open(t, t.TempDir())
		f := adr012Fact(domain.ShadowScopeCI, "999")
		e := adr012Event(f, "event-1")
		link := adr012Link(f, e)
		adr012Record(t, s, f, []domain.ProviderPREvent{e}, []domain.PRRunEventLink{link})
		before := adr012Ledger(t, s)
		adr012Record(t, s, f, []domain.ProviderPREvent{e}, []domain.PRRunEventLink{link})
		adr012Unchanged(t, s, before)
		for _, change := range []string{"final", "rule_version", "event_rule_version"} {
			mutated, event := f, e
			switch change {
			case "final":
				mutated.Inspection = "not_inspected"
			case "rule_version":
				mutated.RuleVersion = "2"
			case "event_rule_version":
				event.RuleVersion = "2"
			}
			if err := s.RecordPRRun(context.Background(), mutated, []domain.ProviderPREvent{event}, []domain.PRRunEventLink{link}); !errors.Is(err, ErrImmutable) {
				t.Errorf("missing immutable %s rejection: %v", change, err)
			}
			adr012Unchanged(t, s, before)
		}
		second := f
		second.ID.RunID = "1000"
		second.BaseCommit = strings.Repeat("f", 40)
		adr012Record(t, s, second, nil, []domain.PRRunEventLink{adr012Link(second, e)})
		ledger := adr012Ledger(t, s)
		if len(ledger.Facts) != 2 || len(ledger.Events) != 1 || len(ledger.Links) != 2 {
			t.Fatalf("missing many-to-many immutable old-event reuse: %#v", ledger)
		}
		// A later empty rerun cannot erase events from earlier attempts.
		third := second
		third.ID.RunID = "1001"
		adr012Record(t, s, third, nil, nil)
		ledger = adr012Ledger(t, s)
		if len(ledger.Facts) != 3 || len(ledger.Events) != 1 || len(ledger.Links) != 2 {
			t.Fatal("missing append-only rerun/event union")
		}
	})

	t.Run("dangling_and_repository_mismatch_roll_back_whole_batch", func(t *testing.T) {
		for _, kind := range []string{"missing_event", "missing_attempt", "foreign_event", "foreign_attempt"} {
			t.Run(kind, func(t *testing.T) {
				s := adr012Open(t, t.TempDir())
				seed := adr012Fact(domain.ShadowScopeCI, "999")
				adr012Record(t, s, seed, nil, nil)
				before := adr012Ledger(t, s)
				f := adr012Fact(domain.ShadowScopeCI, "1000")
				e := adr012Event(f, "event-1")
				link := adr012Link(f, e)
				switch kind {
				case "missing_event":
					link.EventID = "absent"
				case "missing_attempt":
					link.AttemptID.RunID = "absent"
				case "foreign_event":
					e.RepositoryID = "foreign"
				case "foreign_attempt":
					link.AttemptID.RepositoryID = "foreign"
				}
				if err := s.RecordPRRun(context.Background(), f, []domain.ProviderPREvent{e}, []domain.PRRunEventLink{link}); err == nil {
					t.Error("missing dangling/cross-repository link rejection")
				}
				adr012Unchanged(t, s, before)
			})
		}
	})

	t.Run("explicit_local_and_legacy_never_supply_ci", func(t *testing.T) {
		s := adr012Open(t, t.TempDir())
		for _, scope := range []domain.ShadowScope{domain.ShadowScopeLocal, domain.ShadowScopeLegacy} {
			f := adr012Fact(scope, "999")
			e := adr012Event(f, "event-"+string(scope))
			adr012Record(t, s, f, []domain.ProviderPREvent{e}, []domain.PRRunEventLink{adr012Link(f, e)})
		}
		ledger := adr012Ledger(t, s)
		if len(ledger.Facts) != 2 || len(ledger.Events) != 2 || len(ledger.Links) != 2 {
			t.Fatalf("missing explicit independent local/legacy scope persistence: %#v", ledger)
		}
		for _, f := range ledger.Facts {
			if f.Scope == domain.ShadowScopeCI || f.RecordedAt != nil {
				t.Error("missing non-CI provenance separation")
			}
		}
		if len(ledger.Frontiers) != 0 || len(ledger.Publications) != 0 {
			t.Error("missing local/legacy unmeasured CI separation")
		}
		ci := adr012Fact(domain.ShadowScopeCI, "999")
		before := ledger
		if err := s.RecordPRRun(context.Background(), ci, nil, []domain.PRRunEventLink{{AttemptID: ci.ID, Scope: domain.ShadowScopeCI, EventID: "event-local"}}); err == nil {
			t.Error("missing local-to-CI link refusal")
		}
		adr012Unchanged(t, s, before)
		adr012Record(t, s, ci, nil, nil)
		if len(adr012Ledger(t, s).Facts) != 3 {
			t.Error("missing independent CI namespace for same provider identity")
		}
	})

	t.Run("pending_material_cannot_mint_trusted_stamps", func(t *testing.T) {
		for _, stampTarget := range []string{"fact", "event"} {
			t.Run(stampTarget, func(t *testing.T) {
				s := adr012Open(t, t.TempDir())
				before := adr012Ledger(t, s)
				f := adr012Fact(domain.ShadowScopeCI, "999")
				e := adr012Event(f, "event-1")
				stamp := adr012Time.Add(time.Hour)
				if stampTarget == "fact" {
					f.RecordedAt = &stamp
				} else {
					e.RecordedAt = &stamp
				}
				if err := s.RecordPRRun(context.Background(), f, []domain.ProviderPREvent{e}, []domain.PRRunEventLink{adr012Link(f, e)}); err == nil {
					t.Error("missing RecordPRRun trusted-stamp refusal")
				}
				adr012Unchanged(t, s, before)
			})
		}
	})

	t.Run("publication_preserves_immutable_triage_policy_intervals_and_gaps", func(t *testing.T) {
		root := t.TempDir()
		s := adr012Open(t, root)
		f := adr012Fact(domain.ShadowScopeCI, "999")
		e := adr012Event(f, "event-1")
		link := adr012Link(f, e)
		adr012Record(t, s, f, []domain.ProviderPREvent{e}, []domain.PRRunEventLink{link})
		stamp := adr012Time.Add(2 * time.Hour)
		f.RecordedAt = &stamp
		e.RecordedAt = &stamp
		publication := domain.ShadowPublication{ID: "publication-1", RepositoryID: "repository", RecordedAt: stamp, Facts: []domain.PRRunFact{f}, Events: []domain.ProviderPREvent{e}, Links: []domain.PRRunEventLink{link},
			Triage:    []domain.ShadowTriage{{ActionID: "action-1", EventID: e.EventID, RepositoryID: "repository", Scope: domain.ShadowScopeCI, Verdict: "warranted", Actor: "maintainer", Reason: "reviewed", PolicyRevision: "protected-tip-1", TriagedAt: stamp.Add(-time.Minute), RecordedAt: &stamp}},
			Frontiers: []domain.ShadowFrontier{{RepositoryID: "repository", CheckID: f.CheckID, RuleVersion: "1", CompleteThrough: stamp.Add(-time.Minute), PolicyRevision: "protected-tip-1", MinDays: 14, MinObservations: 5}},
			GapPoints: []domain.ShadowGapPoint{{ID: "gap-1", RepositoryID: "repository", CheckID: f.CheckID, At: adr012Time.Add(-time.Hour), Reason: "unstarted eligible event", RecordedAt: &stamp}},
			Intervals: []domain.ShadowInterval{{ID: "interval-1", RepositoryID: "repository", CheckID: f.CheckID, RuleVersion: "1", After: adr012Time.Add(-time.Hour), Through: stamp.Add(-time.Minute)}},
		}
		if err := s.ApplyShadowPublication(context.Background(), publication); err != nil {
			t.Fatalf("missing atomic shadow publication: %v", err)
		}
		before := adr012Ledger(t, s)
		if len(before.Publications) != 1 || !reflect.DeepEqual(before.Publications[0], publication) || !reflect.DeepEqual(before.Triage, publication.Triage) || !reflect.DeepEqual(before.Frontiers, publication.Frontiers) || !reflect.DeepEqual(before.GapPoints, publication.GapPoints) || !reflect.DeepEqual(before.Intervals, publication.Intervals) || !reflect.DeepEqual(before.Facts, publication.Facts) || !reflect.DeepEqual(before.Events, publication.Events) {
			t.Fatalf("missing exact publication/stamp/triage/frontier policy/gap/interval persistence: %#v", before)
		}
		if err := s.ApplyShadowPublication(context.Background(), publication); err != nil {
			t.Errorf("missing exact publication replay no-op: %v", err)
		}
		adr012Unchanged(t, s, before)
		for _, target := range []string{"triage", "frontier", "gap", "interval", "publication"} {
			// JSON clone avoids aliasing the already persisted expectation.
			encoded, err := json.Marshal(publication)
			if err != nil {
				t.Fatal(err)
			}
			var changed domain.ShadowPublication
			if err := json.Unmarshal(encoded, &changed); err != nil {
				t.Fatal(err)
			}
			if target != "publication" {
				// Use a new publication identity: rejecting only a duplicate
				// publication must not masquerade as immutable row enforcement.
				changed.ID = "collision-" + target
				changed.ParentID = publication.ID
				changed.RecordedAt = stamp.Add(time.Millisecond)
				changed.Facts, changed.Events, changed.Links = nil, nil, nil
				if target != "triage" {
					changed.Triage = nil
				} else {
					changed.Triage[0].RecordedAt = &changed.RecordedAt
				}
				if target != "frontier" {
					changed.Frontiers = nil
				}
				if target != "gap" {
					changed.GapPoints = nil
				} else {
					changed.GapPoints[0].RecordedAt = &changed.RecordedAt
				}
				if target != "interval" {
					changed.Intervals = nil
				}
			}
			switch target {
			case "triage":
				changed.Triage[0].Verdict = "not_warranted"
			case "frontier":
				changed.Frontiers[0].MinDays = 1
			case "gap":
				changed.GapPoints[0].Reason = "rewritten"
			case "interval":
				changed.Intervals[0].After = adr012Time
			case "publication":
				changed.RecordedAt = stamp.Add(time.Millisecond)
			}
			if err := s.ApplyShadowPublication(context.Background(), changed); !errors.Is(err, ErrImmutable) {
				t.Errorf("missing append-only %s collision refusal: %v", target, err)
			}
			adr012Unchanged(t, s, before)
		}
		// A valid new fact plus invalid triage must not partially append either.
		bad := domain.ShadowPublication{ID: "publication-2", ParentID: publication.ID, RepositoryID: "repository", RecordedAt: stamp.Add(time.Millisecond), Facts: []domain.PRRunFact{adr012Fact(domain.ShadowScopeCI, "1000")}, Triage: []domain.ShadowTriage{{ActionID: "action-2", EventID: "absent", RepositoryID: "repository", Scope: domain.ShadowScopeCI, Verdict: "warranted", Actor: "maintainer", PolicyRevision: "protected-tip-1", TriagedAt: stamp}}}
		if err := s.ApplyShadowPublication(context.Background(), bad); err == nil {
			t.Error("missing dangling publication triage refusal")
		}
		adr012Unchanged(t, s, before)
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		s = adr012Open(t, root)
		adr012Unchanged(t, s, before)
	})

	t.Run("read_only_store_refuses_both_write_routes", func(t *testing.T) {
		root := t.TempDir()
		s := adr012Open(t, root)
		before := adr012Ledger(t, s)
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		readonly, err := OpenExistingReadOnly(root)
		if err != nil {
			t.Fatal(err)
		}
		defer readonly.Close()
		if err := readonly.RecordPRRun(context.Background(), adr012Fact(domain.ShadowScopeCI, "999"), nil, nil); err == nil {
			t.Error("missing query-only attempt-write refusal")
		}
		if err := readonly.ApplyShadowPublication(context.Background(), domain.ShadowPublication{ID: "forbidden"}); err == nil {
			t.Error("missing query-only publication refusal")
		}
		adr012Unchanged(t, readonly, before)
	})
}

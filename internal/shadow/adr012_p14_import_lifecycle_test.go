package shadow

import (
	"encoding/json"
	"reflect"
	"scout/internal/domain"
	"scout/internal/state"
	"strings"
	"testing"
	"time"
)

func p14Decode(t *testing.T, raw string, dst any) {
	t.Helper()
	if err := json.Unmarshal([]byte(raw), dst); err != nil {
		t.Fatal(err)
	}
}
func p14Ledger() domain.ShadowLedger {
	at := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	stamp := at.Add(time.Hour)
	f := domain.PRRunFact{ID: domain.AttemptID{RepositoryID: "repo", RunID: "999", JobID: "1000", Attempt: "1"}, Scope: domain.ShadowScopeCI, PRNumber: "17", HeadCommit: "H", HeadTree: "tree", BaseCommit: "B", MergeBase: "MB", AssessedRevision: "merge", CheckID: "security:test", RuleVersion: "v1", CreatedAt: at.Add(-time.Minute), StartedAt: at, FinishedAt: at.Add(time.Minute), Status: "assessed", Comparison: "comparable", Inspection: "inspected", Pins: domain.ShadowPins{Scout: "scout", Tools: "tools", Catalogue: "catalogue", Policy: "policy", Recipe: "guardian"}, RecordedAt: &stamp}
	e := domain.ProviderPREvent{EventID: "event", RepositoryID: "repo", Scope: domain.ShadowScopeCI, PRNumber: "17", HeadCommit: "H", HeadTree: "tree", MergeBase: "MB", CheckID: "security:test", IssueID: "issue", RuleVersion: "v1", EventAt: at, RecordedAt: &stamp}
	link := domain.PRRunEventLink{AttemptID: f.ID, Scope: domain.ShadowScopeCI, EventID: e.EventID}
	c := domain.ShadowFrontier{RepositoryID: "repo", CheckID: "security:test", RuleVersion: "v1", CompleteThrough: at.Add(-time.Hour), PolicyRevision: "policy", MinDays: 14, MinObservations: 5}
	pub := domain.ShadowPublication{ID: "genesis", RepositoryID: "repo", RecordedAt: stamp, Facts: []domain.PRRunFact{f}, Events: []domain.ProviderPREvent{e}, Links: []domain.PRRunEventLink{link}, Frontiers: []domain.ShadowFrontier{c}}
	return domain.ShadowLedger{Facts: pub.Facts, Events: pub.Events, Links: pub.Links, Frontiers: pub.Frontiers, Publications: []domain.ShadowPublication{pub}}
}

// S3 write path: successively restamp original rows and persist refusals without
// replay/reopen. S4 owns stamp visibility, per-event Undecided and dossiers.
func TestADR012CrossLineageImportLifecycle(t *testing.T) {
	for _, verdicts := range [][]string{{domain.TriageWarranted, domain.TriageNotWarranted}, {domain.TriageNotWarranted, domain.TriageWarranted}} {
		t.Run(strings.Join(verdicts, "_then_"), func(t *testing.T) {
			source := p14Ledger()
			original := source.Publications[0].RecordedAt
			rows := []domain.ShadowTriage{}
			for i, id := range []string{"repo:999", "repo:1000"} {
				recorded := original.Add(time.Duration(i) * time.Millisecond)
				rows = append(rows, domain.ShadowTriage{ActionID: id, EventID: "event", RepositoryID: "repo", Scope: domain.ShadowScopeCI, Verdict: verdicts[i], Actor: "alice", Reason: "source review", PolicyRevision: "source-policy", TriagedAt: original.Add(time.Duration(i-2) * time.Hour), RecordedAt: &recorded})
			}
			// Reverse input order, but retain valid one-row/event source publications.
			source.Triage = []domain.ShadowTriage{rows[1], rows[0]}
			source.Publications[0].Triage = []domain.ShadowTriage{rows[0]}
			source.Publications = append(source.Publications, domain.ShadowPublication{ID: "source-second", ParentID: "genesis", RepositoryID: "repo", RecordedAt: *rows[1].RecordedAt, Triage: []domain.ShadowTriage{rows[1]}})
			pending, err := state.PendingShadowImport(source, domain.ShadowInputs{})
			if err != nil {
				t.Fatal(err)
			}
			batch := domain.ShadowBatch{Inputs: domain.ShadowInputs{PendingImports: []domain.ShadowPendingImport{pending}}}
			var auth PublicationAuthorization
			p14Decode(t, `{"RepositoryID":"repo","ParentCommit":"destination-parent","PublishingJobStartedAt":"2026-11-10T12:00:00Z","CurrentProtectedTip":"destination-policy","Authenticated":true,"SourceReachable":true,"Policy":{"RepositoryID":"repo","Revision":"destination-policy","Maintainers":["alice"],"Authenticated":true},"OriginalAuthorizations":{"repo:999":true,"repo:1000":true}}`, &auth)
			parent := p14Ledger()
			parent.Triage = nil
			var seen []string
			for i := range 2 {
				pub, err := PreparePublication(parent, batch, auth)
				if err != nil {
					t.Fatal(err)
				}
				if len(pub.Triage) != 1 {
					t.Fatalf("one row/event/publication, iteration %d: %+v", i, pub)
				}
				row := pub.Triage[0]
				seen = append(seen, row.ActionID)
				if row.Verdict != rows[i].Verdict || row.Reason != rows[i].Reason || row.Actor != "alice" || !row.TriagedAt.Equal(rows[i].TriagedAt) || row.RecordedAt == nil || !row.RecordedAt.After(original) {
					t.Fatalf("source provenance/destination stamp lost: %+v", row)
				}
				if row.EventID != "event" {
					t.Fatal("E-prime publication reopened E")
				}
				parent.Triage = append(parent.Triage, row)
				parent.Publications = append(parent.Publications, pub)
				// A destination verdict between import publications must not suppress the
				// next source row. Its per-event Undecided effect is checked by S4.
				if i == 0 {
					local := row
					local.ActionID = "repo:2000"
					local.Verdict = domain.TriageWarranted
					stamp := row.RecordedAt.Add(time.Millisecond)
					local.RecordedAt = &stamp
					parent.Triage = append(parent.Triage, local)
				}
			}
			if !reflect.DeepEqual(seen, []string{"repo:999", "repo:1000"}) {
				t.Fatalf("numeric original order=%v", seen)
			}
			replay, err := PreparePublication(parent, batch, auth)
			if err != nil {
				t.Fatal(err)
			}
			if len(replay.Triage) != 0 || len(replay.Inputs.ImportRefusals) != 0 {
				t.Fatalf("replay reopened event: %+v", replay)
			}
		})
	}
	for _, reason := range []string{"revoked_current_author", "unverifiable_original_authorization"} {
		t.Run(reason, func(t *testing.T) {
			source := p14Ledger()
			stamp := source.Publications[0].RecordedAt
			original := domain.ShadowTriage{ActionID: "repo:1000", EventID: "event", RepositoryID: "repo", Scope: domain.ShadowScopeCI, Verdict: domain.TriageWarranted, Actor: "revoked", Reason: "source review", PolicyRevision: "source-policy", TriagedAt: stamp, RecordedAt: &stamp}
			source.Triage = []domain.ShadowTriage{original}
			source.Publications[0].Triage = source.Triage
			pending, err := state.PendingShadowImport(source, domain.ShadowInputs{})
			if err != nil {
				t.Fatal(err)
			}
			batch := domain.ShadowBatch{Inputs: domain.ShadowInputs{PendingImports: []domain.ShadowPendingImport{pending}}}
			var auth PublicationAuthorization
			maintainers := []string{"alice"}
			verified := true
			if reason == "unverifiable_original_authorization" {
				maintainers = []string{"revoked"}
				verified = false
			}
			raw, _ := json.Marshal(map[string]any{"RepositoryID": "repo", "ParentCommit": "destination-parent", "PublishingJobStartedAt": "2026-11-10T12:00:00Z", "CurrentProtectedTip": "destination-policy", "Authenticated": true, "SourceReachable": true, "Policy": map[string]any{"RepositoryID": "repo", "Revision": "destination-policy", "Maintainers": maintainers, "Authenticated": true}, "OriginalAuthorizations": map[string]bool{"repo:1000": verified}})
			p14Decode(t, string(raw), &auth)
			parent := p14Ledger()
			pub, err := PreparePublication(parent, batch, auth)
			if err != nil {
				t.Fatal(err)
			}
			if len(pub.Triage) != 0 || len(pub.Inputs.ImportRefusals) != 1 {
				t.Fatalf("failed import authorization not persisted as refusal: %+v", pub)
			}
			refusal := pub.Inputs.ImportRefusals[0]
			if !reflect.DeepEqual(refusal.Original, original) || refusal.PolicyRevision != "destination-policy" || refusal.Reason == "" || !refusal.RecordedAt.Equal(pub.RecordedAt) {
				t.Fatalf("refusal lost provenance: %+v", refusal)
			}
			parent.Publications = append(parent.Publications, pub)
			batch.Inputs.ImportRefusals = pub.Inputs.ImportRefusals
			replay, err := PreparePublication(parent, batch, auth)
			if err != nil {
				t.Fatal(err)
			}
			if len(replay.Inputs.ImportRefusals) != 0 || len(replay.Triage) != 0 {
				t.Fatal("refused action replay not idempotent")
			}
			files, err := state.EncodeShadow(parent, pub.Inputs)
			if err != nil {
				t.Fatal(err)
			}
			_, inputs, err := state.DecodeShadow(files)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(inputs.ImportRefusals, pub.Inputs.ImportRefusals) {
				t.Fatal("immutable refusal export/import lost")
			}
		})
	}
}

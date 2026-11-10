package shadow

import (
	"encoding/json"
	"reflect"
	"scout/internal/domain"
	"testing"
	"time"
)

func p11sDecode(t *testing.T, raw string, dst any) {
	t.Helper()
	if err := json.Unmarshal([]byte(raw), dst); err != nil {
		t.Fatal(err)
	}
}
func p11sLedger() domain.ShadowLedger {
	at := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	stamp := at.Add(time.Hour)
	f := domain.PRRunFact{ID: domain.AttemptID{RepositoryID: "repo", RunID: "999", JobID: "1000", Attempt: "1"}, Scope: domain.ShadowScopeCI, PRNumber: "17", HeadCommit: "H", HeadTree: "tree", BaseCommit: "B", MergeBase: "MB", AssessedRevision: "merge", CheckID: "security:test", RuleVersion: "v1", CreatedAt: at.Add(-time.Minute), StartedAt: at, FinishedAt: at.Add(time.Minute), Status: "assessed", Comparison: "comparable", Inspection: "inspected", Pins: domain.ShadowPins{Scout: "scout", Tools: "tools", Catalogue: "catalogue", Policy: "policy", Recipe: "guardian"}, RecordedAt: &stamp}
	e := domain.ProviderPREvent{EventID: "event", RepositoryID: "repo", Scope: domain.ShadowScopeCI, PRNumber: "17", HeadCommit: "H", HeadTree: "tree", MergeBase: "MB", CheckID: "security:test", IssueID: "issue", RuleVersion: "v1", EventAt: at, RecordedAt: &stamp}
	link := domain.PRRunEventLink{AttemptID: f.ID, Scope: domain.ShadowScopeCI, EventID: e.EventID}
	c := domain.ShadowFrontier{RepositoryID: "repo", CheckID: "security:test", RuleVersion: "v1", CompleteThrough: at.Add(-time.Hour), PolicyRevision: "policy", MinDays: 14, MinObservations: 5}
	pub := domain.ShadowPublication{ID: "genesis", RepositoryID: "repo", RecordedAt: stamp, Facts: []domain.PRRunFact{f}, Events: []domain.ProviderPREvent{e}, Links: []domain.PRRunEventLink{link}, Frontiers: []domain.ShadowFrontier{c}}
	return domain.ShadowLedger{Facts: pub.Facts, Events: pub.Events, Links: pub.Links, Frontiers: pub.Frontiers, Publications: []domain.ShadowPublication{pub}}
}

// Shadow half of P11. Actual retry/fetched-parent identity is exercised in
// baseline; query visibility and not-ready/ready dossiers remain S4-owned.
func TestADR012CASParentStampAndFrontier(t *testing.T) {
	for _, tc := range []struct {
		name                        string
		provider, frontier, maximum time.Time
	}{
		{"provider_start_wins", time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC), time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC), time.Date(2026, 11, 1, 11, 0, 0, 0, time.UTC)},
		{"parent_C_wins", time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC), time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC), time.Date(2026, 11, 1, 11, 0, 0, 0, time.UTC)},
		{"parent_max_stamp_wins", time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC), time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC), time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parent := p11sLedger()
			parent.Frontiers[0].CompleteThrough = tc.frontier
			parent.Publications[0].RecordedAt = tc.maximum
			// Keep envelope and row stamps consistent when the parent has a fast clock.
			parent.Facts[0].RecordedAt = &tc.maximum
			parent.Events[0].RecordedAt = &tc.maximum
			parent.Publications[0].Facts = parent.Facts
			parent.Publications[0].Events = parent.Events
			parent.Publications[0].Frontiers = parent.Frontiers
			raw, _ := json.Marshal(parent)
			var before domain.ShadowLedger
			p11sDecode(t, string(raw), &before)
			var auth PublicationAuthorization
			raw, _ = json.Marshal(map[string]any{"RepositoryID": "repo", "ParentCommit": "actual-parent", "PublishingJobStartedAt": tc.provider, "CurrentProtectedTip": "current-policy", "Authenticated": true, "SourceReachable": true, "Policy": map[string]any{"RepositoryID": "repo", "Revision": "current-policy", "MinDays": 21, "MinObservations": 8, "Authenticated": true}})
			p11sDecode(t, string(raw), &auth)
			newC := tc.frontier.Add(24 * time.Hour)
			batch := domain.ShadowBatch{Complete: true, CompleteThrough: newC}
			pub, err := PreparePublication(parent, batch, auth)
			if err != nil {
				t.Fatal(err)
			}
			want := tc.provider
			for _, candidate := range []time.Time{tc.frontier.Add(time.Millisecond), tc.maximum.Add(time.Millisecond)} {
				if candidate.After(want) {
					want = candidate
				}
			}
			if !pub.RecordedAt.Equal(want) {
				t.Fatalf("stamp %s want max formula %s", pub.RecordedAt, want)
			}
			if len(pub.Frontiers) != 1 || !pub.Frontiers[0].CompleteThrough.Equal(newC) || pub.Frontiers[0].PolicyRevision != "current-policy" || pub.Frontiers[0].MinDays != 21 || pub.Frontiers[0].MinObservations != 8 {
				t.Fatalf("frontier policy not snapshotted: %+v", pub.Frontiers)
			}
			if !reflect.DeepEqual(parent, before) {
				t.Fatal("sealed parent mutated")
			}
		})
	}
	t.Run("missing_publishing_provider_start", func(t *testing.T) {
		var auth PublicationAuthorization
		p11sDecode(t, `{"RepositoryID":"repo","ParentCommit":"parent","CurrentProtectedTip":"tip","Authenticated":true,"SourceReachable":true}`, &auth)
		if _, err := PreparePublication(p11sLedger(), domain.ShadowBatch{}, auth); err == nil {
			t.Fatal("writer clock replaced missing publishing-job provider time")
		}
	})
	t.Run("normal_points_and_equal_parent_late_point", func(t *testing.T) {
		parent := p11sLedger()
		c := parent.Frontiers[0].CompleteThrough
		newC := c.Add(time.Hour)
		f := parent.Facts[0]
		f.ID.RunID = "1001"
		f.StartedAt = c.Add(time.Minute)
		f.CreatedAt = f.StartedAt.Add(-time.Minute)
		f.FinishedAt = f.StartedAt.Add(time.Minute)
		f.RecordedAt = nil
		points := []domain.ShadowGapPoint{{ID: "normal", RepositoryID: "repo", CheckID: "security:test", At: c.Add(30 * time.Minute), Reason: "deadline"}, {ID: "at-newC", RepositoryID: "repo", CheckID: "security:test", At: newC, Reason: "deadline"}, {ID: "late-at-parentC", RepositoryID: "repo", CheckID: "security:test", At: c, Reason: "deadline"}}
		batch := domain.ShadowBatch{Complete: true, CompleteThrough: newC, Ledger: domain.ShadowLedger{Facts: []domain.PRRunFact{f}, GapPoints: points}}
		var auth PublicationAuthorization
		p11sDecode(t, `{"RepositoryID":"repo","ParentCommit":"parent","PublishingJobStartedAt":"2026-11-03T12:00:00Z","CurrentProtectedTip":"tip","Authenticated":true,"SourceReachable":true}`, &auth)
		pub, err := PreparePublication(parent, batch, auth)
		if err != nil {
			t.Fatal(err)
		}
		if len(pub.Facts) != 1 || !pub.Facts[0].StartedAt.Equal(f.StartedAt) || len(pub.GapPoints) != 3 {
			t.Fatalf("provider coordinates lost: %+v", pub)
		}
		for _, p := range pub.GapPoints {
			if p.RecordedAt == nil || !p.RecordedAt.Equal(pub.RecordedAt) {
				t.Fatal("point not stamped")
			}
		}
		if len(pub.Frontiers) != 1 || !pub.Frontiers[0].CompleteThrough.Equal(newC) {
			t.Fatalf("late point held C: %+v", pub.Frontiers)
		}
	})
	t.Run("fast_clock_correction_then_quiet_reconcile", func(t *testing.T) {
		parent := p11sLedger()
		fast := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
		parent.Publications = append(parent.Publications, domain.ShadowPublication{ID: "fast-clock", ParentID: "genesis", RepositoryID: "repo", RecordedAt: fast})
		var action TriageAction
		var writer WriterIdentity
		var policy ProtectedPolicy
		var auth PublicationAuthorization
		p11sDecode(t, `{"RepositoryID":"repo","DispatchRunID":"1000","EventID":"event","Verdict":"not_warranted","Actor":"alice","Reason":"correction","OccurredAt":"2026-11-01T11:00:00Z"}`, &action)
		p11sDecode(t, `{"RepositoryID":"repo","OriginalActor":"alice","ExecutingActor":"alice","SourceRevision":"tip","DefaultBranch":true,"SourceReachable":true,"Authenticated":true}`, &writer)
		p11sDecode(t, `{"RepositoryID":"repo","Revision":"tip","Maintainers":["alice"],"Authenticated":true}`, &policy)
		correction, err := AuthorizeTriage(action, writer, policy)
		if err != nil {
			t.Fatal(err)
		}
		p11sDecode(t, `{"RepositoryID":"repo","ParentCommit":"parent","PublishingJobStartedAt":"2026-11-01T12:00:00Z","CurrentProtectedTip":"tip","Authenticated":true,"SourceReachable":true}`, &auth)
		pub, err := PreparePublication(parent, domain.ShadowBatch{AuthorizedTriage: []domain.AuthorizedTriage{correction}}, auth)
		if err != nil {
			t.Fatal(err)
		}
		if len(pub.Triage) != 1 || pub.Triage[0].Verdict != domain.TriageNotWarranted || !pub.RecordedAt.Equal(fast.Add(time.Millisecond)) {
			t.Fatalf("correction lost under fast clock: %+v", pub)
		}
		parent.Publications = append(parent.Publications, pub)
		parent.Triage = append(parent.Triage, pub.Triage...)
		quiet, err := PreparePublication(parent, domain.ShadowBatch{Complete: true, CompleteThrough: fast.Add(time.Hour)}, auth)
		if err != nil {
			t.Fatal(err)
		}
		if len(quiet.Triage) != 0 || len(quiet.Inputs.ImportRefusals) != 0 || len(quiet.Frontiers) != 1 || quiet.Frontiers[0].CompleteThrough.Before(pub.RecordedAt) {
			t.Fatalf("quiet reconcile did not pass correction stamp: %+v", quiet)
		}
	})
}

package shadow

import (
	"encoding/json"
	"scout/internal/domain"
	"testing"
	"time"
)

func p13Decode(t *testing.T, raw string, dst any) {
	t.Helper()
	if err := json.Unmarshal([]byte(raw), dst); err != nil {
		t.Fatal(err)
	}
}
func TestADR012TriageAuthorizationReplay(t *testing.T) {
	for _, tc := range []struct {
		name, actor, executing, source                      string
		authenticated, defaultBranch, reachable, maintainer bool
		ok                                                  bool
	}{
		{"maintainer", "alice", "alice", "protected", true, true, true, true, true},
		{"nonmaintainer", "mallory", "mallory", "protected", true, true, true, true, false},
		{"revoked", "alice", "alice", "protected", true, true, true, false, false},
		{"different_rerunner", "alice", "bob", "protected", true, true, true, true, false},
		{"untrusted_actor_claim", "alice", "alice", "protected", false, true, true, true, false},
		{"nondefault_execution", "alice", "alice", "feature", true, false, true, true, false},
		{"unreachable_source", "alice", "alice", "feature", true, true, false, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var action TriageAction
			var writer WriterIdentity
			var policy ProtectedPolicy
			p13Decode(t, `{"RepositoryID":"repo","DispatchRunID":"999","EventID":"event","Verdict":"warranted","Actor":"alice","Reason":"reviewed","OccurredAt":"2026-11-01T10:00:00Z"}`, &action)
			raw, _ := json.Marshal(map[string]any{"RepositoryID": "repo", "OriginalActor": tc.actor, "ExecutingActor": tc.executing, "SourceRevision": tc.source, "Authenticated": tc.authenticated, "DefaultBranch": tc.defaultBranch, "SourceReachable": tc.reachable})
			p13Decode(t, string(raw), &writer)
			maintainers := []string{}
			if tc.maintainer {
				maintainers = []string{"alice"}
			}
			raw, _ = json.Marshal(map[string]any{"RepositoryID": "repo", "Revision": "current-protected-tip", "Maintainers": maintainers, "Authenticated": true})
			p13Decode(t, string(raw), &policy)
			got, err := AuthorizeTriage(action, writer, policy)
			if !tc.ok {
				if err == nil {
					t.Fatalf("unauthorized action accepted: %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.Row.PolicyRevision != "current-protected-tip" || got.Row.Actor != "alice" || got.Row.ActionID != "repo:999" {
				t.Fatalf("missing current-tip/stable action provenance: %+v", got)
			}
		})
	}
	t.Run("CODEOWNERS_exact_names_no_team_expansion", func(t *testing.T) {
		var action TriageAction
		var writer WriterIdentity
		var policy ProtectedPolicy
		p13Decode(t, `{"RepositoryID":"repo","DispatchRunID":"1000","EventID":"event","Verdict":"not_warranted","Actor":"alice","Reason":"reviewed","OccurredAt":"2026-11-01T10:00:00Z"}`, &action)
		p13Decode(t, `{"RepositoryID":"repo","OriginalActor":"alice","ExecutingActor":"alice","SourceRevision":"protected","DefaultBranch":true,"SourceReachable":true,"Authenticated":true}`, &writer)
		p13Decode(t, `{"RepositoryID":"repo","Revision":"protected","CODEOWNERS":"* @alice\n","Authenticated":true}`, &policy)
		if _, err := AuthorizeTriage(action, writer, policy); err != nil {
			t.Fatal(err)
		}
		p13Decode(t, `{"RepositoryID":"repo","Revision":"protected","CODEOWNERS":"* @org/team\n","Authenticated":true}`, &policy)
		if _, err := AuthorizeTriage(action, writer, policy); err == nil {
			t.Fatal("team expansion authorized actor")
		}
	})
	t.Run("stable_rerun_replay_cannot_overtake_newer", func(t *testing.T) {
		var action TriageAction
		var writer WriterIdentity
		var policy ProtectedPolicy
		var auth PublicationAuthorization
		p13Decode(t, `{"RepositoryID":"repo","DispatchRunID":"999","EventID":"event","Verdict":"warranted","Actor":"alice","Reason":"reviewed","OccurredAt":"2026-11-01T10:00:00Z"}`, &action)
		p13Decode(t, `{"RepositoryID":"repo","OriginalActor":"alice","ExecutingActor":"alice","SourceRevision":"protected","DefaultBranch":true,"SourceReachable":true,"Authenticated":true}`, &writer)
		p13Decode(t, `{"RepositoryID":"repo","Revision":"protected","Maintainers":["alice"],"Authenticated":true}`, &policy)
		authorized, err := AuthorizeTriage(action, writer, policy)
		if err != nil {
			t.Fatal(err)
		}
		p13Decode(t, `{"RepositoryID":"repo","ParentCommit":"parent","PublishingJobStartedAt":"2026-11-01T12:00:00Z","CurrentProtectedTip":"protected","Authenticated":true,"SourceReachable":true}`, &auth)
		at := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
		stamp := at.Add(time.Hour)
		old := authorized.Row
		old.RecordedAt = &stamp
		newer := old
		newer.ActionID = "repo:1000"
		newer.Verdict = domain.TriageNotWarranted
		next := stamp.Add(time.Millisecond)
		newer.RecordedAt = &next
		event := domain.ProviderPREvent{EventID: "event", RepositoryID: "repo", Scope: domain.ShadowScopeCI, PRNumber: "17", HeadCommit: "H", HeadTree: "tree", MergeBase: "MB", CheckID: "security:test", IssueID: "issue", RuleVersion: "v1", EventAt: at, RecordedAt: &stamp}
		parent := domain.ShadowLedger{Events: []domain.ProviderPREvent{event}, Triage: []domain.ShadowTriage{old, newer}}
		batch := domain.ShadowBatch{AuthorizedTriage: []domain.AuthorizedTriage{authorized}}
		pub, err := PreparePublication(parent, batch, auth)
		if err != nil {
			t.Fatal(err)
		}
		if len(pub.Triage) != 0 {
			t.Fatalf("rerun overtook newer verdict: %+v", pub.Triage)
		}
		batch.AuthorizedTriage = append(batch.AuthorizedTriage, authorized)
		pub, err = PreparePublication(domain.ShadowLedger{Events: parent.Events}, batch, auth)
		if err != nil {
			t.Fatal(err)
		}
		if len(pub.Triage) != 1 {
			t.Fatalf("at most one event row/publication: %+v", pub.Triage)
		}
	})
}

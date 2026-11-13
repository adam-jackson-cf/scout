package shadow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"scout/internal/domain"
)

// This transcript is conditional provider evidence, not a claim about GitHub's
// live capabilities. In particular an HTTP Date or an empty page proves nothing.
type p08Provider struct {
	client *http.Client
	url    string
}

func (p p08Provider) get(ctx context.Context, path string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url+path, nil)
	if err != nil {
		return err
	}
	res, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("provider HTTP %d", res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(dst)
}
func (p p08Provider) Enumerate(ctx context.Context, _ CollectionRequest) (CollectionSnapshot, error) {
	var v CollectionSnapshot
	err := p.get(ctx, "/enumerate", &v)
	return v, err
}
func (p p08Provider) AuthenticateWriter(ctx context.Context) (WriterIdentity, error) {
	var v WriterIdentity
	err := p.get(ctx, "/writer", &v)
	return v, err
}
func (p p08Provider) ProtectedTip(ctx context.Context, _ string) (ProtectedPolicy, error) {
	var v ProtectedPolicy
	err := p.get(ctx, "/policy", &v)
	return v, err
}
func (p p08Provider) ProveChain(ctx context.Context, _ ChainRequest) (domain.ChainProof, error) {
	var v domain.ChainProof
	err := p.get(ctx, "/chain", &v)
	return v, err
}

var _ Provider = p08Provider{}

func p08Decode(t *testing.T, raw string, dst any) {
	t.Helper()
	if err := json.Unmarshal([]byte(raw), dst); err != nil {
		t.Fatal(err)
	}
}
func p08Collect(t *testing.T, transcript map[string]any) (domain.ShadowBatch, error) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/enumerate":
			_ = json.NewEncoder(w).Encode(transcript)
		case "/writer":
			_, _ = w.Write([]byte(`{"RepositoryID":"repo","OriginalActor":"alice","ExecutingActor":"alice","SourceRevision":"protected","DefaultBranch":true,"SourceReachable":true,"Authenticated":true}`))
		case "/policy":
			_, _ = w.Write([]byte(`{"RepositoryID":"repo","Revision":"protected","Maintainers":["alice"],"Authenticated":true}`))
		case "/chain":
			_, _ = w.Write([]byte(`{"RepositoryID":"repo","StateRef":"refs/heads/scout/state","GenesisCommit":"genesis","SourceTip":"protected","DestinationTip":"parent","Authenticated":true,"DestinationAncestor":true}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	var req CollectionRequest
	p08Decode(t, `{"RepositoryID":"repo","CheckID":"security:test","RuleVersion":"v1","Selector":"guardian-trigger-v1/collector-v1","Now":"2026-11-01T12:00:00Z","Deadline":"2026-11-01T11:00:00Z","ConsistencyLag":60000000000}`, &req)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return Collect(ctx, p08Provider{server.Client(), server.URL}, req)
}
func p08Snapshot(t *testing.T) map[string]any {
	t.Helper()
	var v map[string]any
	p08Decode(t, `{"RepositoryID":"repo","ProviderTime":"2026-11-01T12:00:00Z","Complete":true,"EventCompleteThrough":"2026-11-01T12:00:00Z","AttemptCompleteThrough":"2026-11-01T12:00:00Z","PagesComplete":true,"Selector":"guardian-trigger-v1/collector-v1","Trigger":{"Events":["pull_request"],"BaseBranch":"main","Paths":["**/*.go"]},"Events":[{"ID":"E1","RepositoryID":"repo","Scope":"ci","PRNumber":"17","HeadCommit":"H1","EventAt":"2026-11-01T10:00:00Z","Selector":"guardian-trigger-v1/collector-v1","ProviderProof":"event-ledger-page-1","BaseBranch":"main","BaseReachable":true,"Author":"dependabot[bot]","Fork":true,"State":"open","DeadlineAt":"2026-11-01T11:00:00Z"}],"Attempts":[{"Fact":{"ID":{"RepositoryID":"repo","RunID":"1000","JobID":"2000","Attempt":"1"},"Scope":"ci","PRNumber":"17","HeadCommit":"H1","HeadTree":"tree-H1","BaseCommit":"B","MergeBase":"MB","AssessedRevision":"merge-H1","CheckID":"security:test","RuleVersion":"v1","CreatedAt":"2026-11-01T10:01:00Z","StartedAt":"2026-11-01T10:02:00Z","FinishedAt":"2026-11-01T10:03:00Z","Status":"assessed","Comparison":"comparable","Inspection":"inspected","Pins":{"Scout":"scout","Tools":"tools","Catalogue":"catalogue","Policy":"policy","Recipe":"guardian"}},"AccountedEventIDs":["E1"],"Terminal":true,"DeadlineAt":"2026-11-01T11:00:00Z","Lane":"guardian","ConcurrencyGroup":"guardian-pr-17","PinnedConfigDigest":"config","ExecutedConfigDigest":"config","CancelInProgress":true,"ConfigurationVerified":true}],"EnumerationProof":"complete-event-and-attempt-ledger-not-run-list","TimeProof":"provider-job-clock-ms","LagJustification":"fixture-consistency-contract-v1"}`, &v)
	return v
}
func p08Points(t *testing.T, b domain.ShadowBatch, want ...string) {
	t.Helper()
	for _, raw := range want {
		at, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, point := range b.Ledger.GapPoints {
			if point.At.Equal(at) {
				found = true
			}
		}
		if !found {
			t.Errorf("missing independent obligation point %s: %+v", raw, b.Ledger.GapPoints)
		}
	}
}

func TestADR012CancellationExclusions(t *testing.T) {
	// Every negative varies independently from a fully proven positive control.
	for _, tc := range []struct {
		name     string
		mutate   func(map[string]any)
		excluded bool
	}{
		{"proven_supersession", func(a map[string]any) {}, true},
		{"manual_cancel", func(a map[string]any) { a["CancellationCause"] = "manual" }, false},
		{"manual_cancel_then_later_push", func(a map[string]any) {
			a["CancellationCause"] = "manual"
			a["SuccessorCreatedAt"] = "2026-11-01T10:06:00Z"
		}, false},
		{"skip_ci_text", func(a map[string]any) { a["CancellationCause"] = "unknown"; a["CommitMessage"] = "[skip ci]" }, false},
		{"unknown_actor", func(a map[string]any) { a["CancellationCause"] = "unknown"; a["CancellationActor"] = "" }, false},
		{"all_cancel_false", func(a map[string]any) { a["CancelInProgress"] = false; a["SuccessorCancelInProgress"] = false }, false},
		{"pinned_false_H2_before_manual", func(a map[string]any) { a["CancelInProgress"] = false; a["CancellationCause"] = "manual" }, false},
		{"R1_true_R2_false_PR_edit", func(a map[string]any) {
			a["SuccessorCancelInProgress"] = false
			a["SuccessorExecutedConfigDigest"] = "PR-edited"
		}, false},
		{"R1_true_R2_false_base_change", func(a map[string]any) {
			a["SuccessorCancelInProgress"] = false
			a["SuccessorExecutedConfigDigest"] = "new-base"
		}, false},
		{"unverifiable_R2", func(a map[string]any) { a["SuccessorConfigurationVerified"] = false }, false},
		{"foreign_PR_added_job_no_H2", func(a map[string]any) {
			a["SuccessorPRNumber"] = "18"
			a["SuccessorHeadCommit"] = ""
			a["CancellationCause"] = "added-job"
		}, false},
		{"pinned_false_PR_edited_R2_true_inferred", func(a map[string]any) {
			a["CancelInProgress"] = false
			a["SuccessorExecutedConfigDigest"] = "PR-edited"
			a["CancellationCause"] = "inferred-concurrency"
		}, false},
		{"pinned_false_PR_edited_R2_true_provider", func(a map[string]any) {
			a["CancelInProgress"] = false
			a["SuccessorExecutedConfigDigest"] = "PR-edited"
		}, false},
		{"R2_created_after_cancel", func(a map[string]any) { a["SuccessorCreatedAt"] = "2026-11-01T10:06:00Z" }, false},
		{"R2_created_at_start", func(a map[string]any) { a["SuccessorCreatedAt"] = "2026-11-01T10:02:00Z" }, false},
		{"R2_created_at_cancel", func(a map[string]any) { a["SuccessorCreatedAt"] = "2026-11-01T10:05:00Z" }, false},
		{"older_R1_added_job", func(a map[string]any) {
			a["SuccessorCreatedAt"] = "2026-11-01T09:00:00Z"
			a["CancellationCause"] = "added-job"
		}, false},
		{"older_R1_manual_no_actor", func(a map[string]any) {
			a["SuccessorCreatedAt"] = "2026-11-01T09:00:00Z"
			a["CancellationCause"] = "manual"
			a["CancellationActor"] = ""
		}, false},
		{"terminal_R1_rerun_provider_cause", func(a map[string]any) {
			a["SuccessorCreatedAt"] = "2026-11-01T09:00:00Z"
			a["SuccessorAttemptStartedAt"] = "2026-11-01T10:04:00Z"
			a["SuccessorAttempt"] = "2"
		}, false},
		{"terminal_R1_rerun_inferred_cause", func(a map[string]any) {
			a["SuccessorCreatedAt"] = "2026-11-01T09:00:00Z"
			a["SuccessorAttemptStartedAt"] = "2026-11-01T10:04:00Z"
			a["CancellationCause"] = "inferred-concurrency"
		}, false},
		{"unknown_original_creation", func(a map[string]any) { a["SuccessorCreatedAt"] = "" }, false},
		{"different_lane", func(a map[string]any) { a["SuccessorLane"] = "PR-added-job" }, false},
		{"different_group", func(a map[string]any) { a["SuccessorConcurrencyGroup"] = "other" }, false},
		{"proven_conflict_no_merge", func(a map[string]any) {
			a["CancellationCause"] = "merge-conflict"
			a["MergeConflictProven"] = true
			a["MergeResultAvailable"] = false
		}, true},
		{"system_cancellation", func(a map[string]any) { a["CancellationCause"] = "system"; a["SystemCancellationProven"] = true }, true},
		{"configured_skip", func(a map[string]any) { a["CancellationCause"] = "configured-skip"; a["ConfiguredSkipProven"] = true }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := p08Snapshot(t)
			a := s["Attempts"].([]any)[0].(map[string]any)
			f := a["Fact"].(map[string]any)
			f["Status"] = "cancelled"
			f["FinishedAt"] = "2026-11-01T10:05:00Z"
			for k, v := range map[string]any{"CancelledAt": "2026-11-01T10:05:00Z", "CancellationCause": "provider-concurrency", "CancellationActor": "provider", "SuccessorPRNumber": "17", "SuccessorHeadCommit": "H2", "SuccessorLane": "guardian", "SuccessorConcurrencyGroup": "guardian-pr-17", "SuccessorCreatedAt": "2026-11-01T10:04:00Z", "SuccessorAttemptStartedAt": "2026-11-01T10:04:30Z", "SuccessorAttempt": "1", "SuccessorExecutedConfigDigest": "config", "SuccessorPinnedConfigDigest": "config", "SuccessorConfigurationVerified": true, "SuccessorCancelInProgress": true} {
				a[k] = v
			}
			tc.mutate(a)
			b, err := p08Collect(t, s)
			if err != nil {
				t.Fatal(err)
			}
			if tc.excluded {
				if len(b.Ledger.GapPoints) != 0 {
					t.Fatalf("proven exclusion produced gap: %+v", b.Ledger)
				}
			} else {
				p08Points(t, b, "2026-11-01T10:00:00Z", "2026-11-01T10:02:00Z")
			}
		})
	}
}

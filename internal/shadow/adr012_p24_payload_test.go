package shadow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"scout/internal/domain"
	"scout/internal/state"
)

// This transcript is conditional provider evidence, not a claim about GitHub's
// live capabilities. In particular an HTTP Date or an empty page proves nothing.
type p24Provider struct {
	client *http.Client
	url    string
}

func (p p24Provider) get(ctx context.Context, path string, dst any) error {
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
func (p p24Provider) Enumerate(ctx context.Context, _ CollectionRequest) (CollectionSnapshot, error) {
	var v CollectionSnapshot
	err := p.get(ctx, "/enumerate", &v)
	return v, err
}
func (p p24Provider) AuthenticateWriter(ctx context.Context) (WriterIdentity, error) {
	var v WriterIdentity
	err := p.get(ctx, "/writer", &v)
	return v, err
}
func (p p24Provider) ProtectedTip(ctx context.Context, _ string) (ProtectedPolicy, error) {
	var v ProtectedPolicy
	err := p.get(ctx, "/policy", &v)
	return v, err
}
func (p p24Provider) ProveChain(ctx context.Context, _ ChainRequest) (domain.ChainProof, error) {
	var v domain.ChainProof
	err := p.get(ctx, "/chain", &v)
	return v, err
}

var _ Provider = p24Provider{}

func p24Decode(t *testing.T, raw string, dst any) {
	t.Helper()
	if err := json.Unmarshal([]byte(raw), dst); err != nil {
		t.Fatal(err)
	}
}
func p24Collect(t *testing.T, transcript map[string]any) (domain.ShadowBatch, error) {
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
	p24Decode(t, `{"RepositoryID":"repo","CheckID":"security:test","RuleVersion":"v1","Selector":"guardian-trigger-v1/collector-v1","Now":"2026-11-01T12:00:00Z","Deadline":"2026-11-01T11:00:00Z","ConsistencyLag":60000000000}`, &req)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return Collect(ctx, p24Provider{server.Client(), server.URL}, req)
}
func p24Snapshot(t *testing.T) map[string]any {
	t.Helper()
	var v map[string]any
	p24Decode(t, `{"RepositoryID":"repo","ProviderTime":"2026-11-01T12:00:00Z","Complete":true,"EventCompleteThrough":"2026-11-01T12:00:00Z","AttemptCompleteThrough":"2026-11-01T12:00:00Z","PagesComplete":true,"Selector":"guardian-trigger-v1/collector-v1","Trigger":{"Events":["pull_request"],"BaseBranch":"main","Paths":["**/*.go"]},"Events":[{"ID":"E1","RepositoryID":"repo","Scope":"ci","PRNumber":"17","HeadCommit":"H1","EventAt":"2026-11-01T10:00:00Z","Selector":"guardian-trigger-v1/collector-v1","ProviderProof":"event-ledger-page-1","BaseBranch":"main","BaseReachable":true,"Author":"dependabot[bot]","Fork":true,"State":"open","DeadlineAt":"2026-11-01T11:00:00Z"}],"Attempts":[{"Fact":{"ID":{"RepositoryID":"repo","RunID":"1000","JobID":"2000","Attempt":"1"},"Scope":"ci","PRNumber":"17","HeadCommit":"H1","HeadTree":"tree-H1","BaseCommit":"B","MergeBase":"MB","AssessedRevision":"merge-H1","CheckID":"security:test","RuleVersion":"v1","CreatedAt":"2026-11-01T10:01:00Z","StartedAt":"2026-11-01T10:02:00Z","FinishedAt":"2026-11-01T10:03:00Z","Status":"assessed","Comparison":"comparable","Inspection":"inspected","Pins":{"Scout":"scout","Tools":"tools","Catalogue":"catalogue","Policy":"policy","Recipe":"guardian"}},"AccountedEventIDs":["E1"],"Terminal":true,"DeadlineAt":"2026-11-01T11:00:00Z","Lane":"guardian","ConcurrencyGroup":"guardian-pr-17","PinnedConfigDigest":"config","ExecutedConfigDigest":"config","CancelInProgress":true,"ConfigurationVerified":true}],"EnumerationProof":"complete-event-and-attempt-ledger-not-run-list","TimeProof":"provider-job-clock-ms","LagJustification":"fixture-consistency-contract-v1"}`, &v)
	return v
}
func p24Points(t *testing.T, b domain.ShadowBatch, want ...string) {
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

// S3 control-plane composition only. Real executable/configuration/symlink and
// credential isolation belongs to P23/S6, not to this HTTP fixture.
func TestADR012UntrustedPayloadCannotAuthorizePublication(t *testing.T) {
	for _, tc := range []struct {
		name   string
		forged map[string]any
	}{
		{"authority", map[string]any{"Actor": "mallory", "RepositoryID": "foreign", "Authenticated": true, "PolicyRevision": "PR-tip"}},
		{"attempt", map[string]any{"ID": map[string]any{"RepositoryID": "foreign", "RunID": "9999", "JobID": "9999", "Attempt": "9"}}},
		{"recipe", map[string]any{"Pins": map[string]any{"Scout": "PR-scout", "Recipe": "PR-recipe", "Tools": "PR-tool", "Policy": "PR-policy"}}},
		{"stamps_frontier", map[string]any{"RecordedAt": "2099-01-01T00:00:00Z", "Frontiers": []any{map[string]any{"CompleteThrough": "2099-01-01T00:00:00Z"}}}},
		{"paths_executables", map[string]any{"../../scout": "#!/bin/sh evil", "Files": map[string]any{"policy.yaml": "forged", "scout": "evil"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := p24Snapshot(t)
			a := s["Attempts"].([]any)[0].(map[string]any)
			a["ScannerOutput"] = tc.forged
			b, err := p24Collect(t, s)
			if err != nil {
				t.Fatal(err)
			}
			if len(b.Ledger.Facts) != 1 {
				t.Fatalf("lost bounded own attempt: %+v", b)
			}
			f := b.Ledger.Facts[0]
			if f.ID.RepositoryID != "repo" || f.ID.RunID != "1000" || f.ID.JobID != "2000" || f.ID.Attempt != "1" || f.Pins.Recipe != "guardian" || f.Pins.Scout != "scout" || f.RecordedAt != nil {
				t.Fatalf("scanner acquired provenance authority: %+v", f)
			}
			if len(b.Ledger.Frontiers) != 0 || len(b.Ledger.Triage) != 0 || len(b.AuthorizedTriage) != 0 {
				t.Fatalf("scanner acquired publication authority: %+v", b)
			}
		})
	}
	t.Run("raw_triage_in_batch_not_authorized", func(t *testing.T) {
		var auth PublicationAuthorization
		p24Decode(t, `{"RepositoryID":"repo","ParentCommit":"parent","PublishingJobStartedAt":"2026-11-01T12:00:00Z","CurrentProtectedTip":"protected","Authenticated":true,"SourceReachable":true}`, &auth)
		b := domain.ShadowBatch{Ledger: domain.ShadowLedger{Triage: []domain.ShadowTriage{{ActionID: "repo:999", EventID: "event", RepositoryID: "repo", Scope: domain.ShadowScopeCI, Actor: "alice", PolicyRevision: "protected", Verdict: domain.TriageWarranted}}}}
		p, err := PreparePublication(domain.ShadowLedger{}, b, auth)
		if err == nil && len(p.Triage) != 0 {
			t.Fatalf("scanner supplied raw authorization accepted: %+v", p)
		}
	})
	t.Run("own_status_and_events_residual_risk", func(t *testing.T) {
		s := p24Snapshot(t)
		a := s["Attempts"].([]any)[0].(map[string]any)
		a["ScannerOutput"] = map[string]any{"Status": "assessed", "Events": []any{map[string]any{"IssueID": "scanner-claim", "CheckID": "security:test", "RuleVersion": "v1"}}}
		b, err := p24Collect(t, s)
		if err != nil {
			t.Fatal(err)
		}
		if len(b.Ledger.Events) != 1 || len(b.Ledger.Links) != 1 || b.Ledger.Links[0].AttemptID != b.Ledger.Facts[0].ID {
			t.Fatalf("own bounded measurement claims not linked: %+v", b.Ledger)
		}
	})
	t.Run("readiness_append_preserves_baseline_content", func(t *testing.T) {
		s := p24Snapshot(t)
		batch, err := p24Collect(t, s)
		if err != nil {
			t.Fatal(err)
		}
		var auth PublicationAuthorization
		p24Decode(t, `{"RepositoryID":"repo","ParentCommit":"parent","PublishingJobStartedAt":"2026-11-01T12:00:00Z","CurrentProtectedTip":"protected","Authenticated":true,"SourceReachable":true}`, &auth)
		pub, err := PreparePublication(domain.ShadowLedger{}, batch, auth)
		if err != nil {
			t.Fatal(err)
		}
		baseline := map[string][]byte{
			state.ExportIssues:          []byte("{\"issueId\":\"baseline\",\"title\":\"keep\"}\n"),
			state.ExportDispositions:    []byte("{\"issueId\":\"baseline\",\"state\":\"open\"}\n"),
			state.ExportWouldBlock:      []byte("{\"legacy\":\"byte-preserved\"}\n"),
			state.ExportCoverageSummary: []byte("{\"baseline\":\"coverage\"}\n"),
		}
		before := map[string][]byte{}
		for k, v := range baseline {
			before[k] = append([]byte(nil), v...)
		}
		ledger := domain.ShadowLedger{Publications: []domain.ShadowPublication{pub}, Facts: pub.Facts, Events: pub.Events, Links: pub.Links, Triage: pub.Triage, Frontiers: pub.Frontiers, GapPoints: pub.GapPoints, Intervals: pub.Intervals}
		addition, err := state.EncodeShadow(ledger, pub.Inputs)
		if err != nil {
			t.Fatal(err)
		}
		for k, v := range addition {
			baseline[k] = v
		}
		for k, v := range before {
			if !bytes.Equal(baseline[k], v) {
				t.Fatalf("readiness-only append mutated %s", k)
			}
		}
	})
}

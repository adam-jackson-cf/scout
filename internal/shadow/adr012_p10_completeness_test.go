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
type p10Provider struct {
	client *http.Client
	url    string
}

func (p p10Provider) get(ctx context.Context, path string, dst any) error {
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
func (p p10Provider) Enumerate(ctx context.Context, _ CollectionRequest) (CollectionSnapshot, error) {
	var v CollectionSnapshot
	err := p.get(ctx, "/enumerate", &v)
	return v, err
}
func (p p10Provider) AuthenticateWriter(ctx context.Context) (WriterIdentity, error) {
	var v WriterIdentity
	err := p.get(ctx, "/writer", &v)
	return v, err
}
func (p p10Provider) ProtectedTip(ctx context.Context, _ string) (ProtectedPolicy, error) {
	var v ProtectedPolicy
	err := p.get(ctx, "/policy", &v)
	return v, err
}
func (p p10Provider) ProveChain(ctx context.Context, _ ChainRequest) (domain.ChainProof, error) {
	var v domain.ChainProof
	err := p.get(ctx, "/chain", &v)
	return v, err
}

var _ Provider = p10Provider{}

func p10Decode(t *testing.T, raw string, dst any) {
	t.Helper()
	if err := json.Unmarshal([]byte(raw), dst); err != nil {
		t.Fatal(err)
	}
}
func p10Collect(t *testing.T, transcript map[string]any) (domain.ShadowBatch, error) {
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
	p10Decode(t, `{"RepositoryID":"repo","CheckID":"security:test","RuleVersion":"v1","Selector":"guardian-trigger-v1/collector-v1","Now":"2026-11-01T12:00:00Z","Deadline":"2026-11-01T11:00:00Z","ConsistencyLag":60000000000}`, &req)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return Collect(ctx, p10Provider{server.Client(), server.URL}, req)
}
func p10Snapshot(t *testing.T) map[string]any {
	t.Helper()
	var v map[string]any
	p10Decode(t, `{"RepositoryID":"repo","ProviderTime":"2026-11-01T12:00:00Z","Complete":true,"EventCompleteThrough":"2026-11-01T12:00:00Z","AttemptCompleteThrough":"2026-11-01T12:00:00Z","PagesComplete":true,"Selector":"guardian-trigger-v1/collector-v1","Trigger":{"Events":["pull_request"],"BaseBranch":"main","Paths":["**/*.go"]},"Events":[{"ID":"E1","RepositoryID":"repo","Scope":"ci","PRNumber":"17","HeadCommit":"H1","EventAt":"2026-11-01T10:00:00Z","Selector":"guardian-trigger-v1/collector-v1","ProviderProof":"event-ledger-page-1","BaseBranch":"main","BaseReachable":true,"Author":"dependabot[bot]","Fork":true,"State":"open","DeadlineAt":"2026-11-01T11:00:00Z"}],"Attempts":[{"Fact":{"ID":{"RepositoryID":"repo","RunID":"1000","JobID":"2000","Attempt":"1"},"Scope":"ci","PRNumber":"17","HeadCommit":"H1","HeadTree":"tree-H1","BaseCommit":"B","MergeBase":"MB","AssessedRevision":"merge-H1","CheckID":"security:test","RuleVersion":"v1","CreatedAt":"2026-11-01T10:01:00Z","StartedAt":"2026-11-01T10:02:00Z","FinishedAt":"2026-11-01T10:03:00Z","Status":"assessed","Comparison":"comparable","Inspection":"inspected","Pins":{"Scout":"scout","Tools":"tools","Catalogue":"catalogue","Policy":"policy","Recipe":"guardian"}},"AccountedEventIDs":["E1"],"Terminal":true,"DeadlineAt":"2026-11-01T11:00:00Z","Lane":"guardian","ConcurrencyGroup":"guardian-pr-17","PinnedConfigDigest":"config","ExecutedConfigDigest":"config","CancelInProgress":true,"ConfigurationVerified":true}],"EnumerationProof":"complete-event-and-attempt-ledger-not-run-list","TimeProof":"provider-job-clock-ms","LagJustification":"fixture-consistency-contract-v1"}`, &v)
	return v
}
func p10Points(t *testing.T, b domain.ShadowBatch, want ...string) {
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

func TestADR012CompletenessBoundsAndLag(t *testing.T) {
	for _, field := range []string{"ProviderTime", "EventCompleteThrough", "AttemptCompleteThrough", "EnumerationProof", "TimeProof", "LagJustification"} {
		t.Run("missing_"+field, func(t *testing.T) {
			s := p10Snapshot(t)
			delete(s, field)
			b, err := p10Collect(t, s)
			if err == nil && b.Complete {
				t.Fatalf("missing %s certified complete: %+v", field, b)
			}
		})
	}
	for _, field := range []string{"Complete", "PagesComplete"} {
		t.Run(field, func(t *testing.T) {
			s := p10Snapshot(t)
			s[field] = false
			s["ListingCapped"] = true
			b, err := p10Collect(t, s)
			if err == nil && b.Complete {
				t.Fatal("capped/omitted page certified complete")
			}
		})
	}
	t.Run("lagged_bound_recovers", func(t *testing.T) {
		s := p10Snapshot(t)
		s["EventCompleteThrough"] = "2026-11-01T10:00:00Z"
		s["AttemptCompleteThrough"] = "2026-11-01T10:00:00Z"
		b, err := p10Collect(t, s)
		if err != nil {
			t.Fatal(err)
		}
		if b.CompleteThrough.After(time.Date(2026, 11, 1, 9, 59, 0, 0, time.UTC)) {
			t.Fatalf("ignored bound minus lag: %s", b.CompleteThrough)
		}
		b, err = p10Collect(t, p10Snapshot(t))
		if err != nil {
			t.Fatal(err)
		}
		if !b.Complete || !b.CompleteThrough.Equal(time.Date(2026, 11, 1, 11, 59, 0, 0, time.UTC)) {
			t.Fatalf("quiet complete bounds did not recover: %+v", b)
		}
	})
	t.Run("all_started_attempts_discharged", func(t *testing.T) {
		s := p10Snapshot(t)
		a := s["Attempts"].([]any)[0].(map[string]any)
		a["Terminal"] = false
		a["Fact"].(map[string]any)["FinishedAt"] = "0001-01-01T00:00:00Z"
		b, err := p10Collect(t, s)
		if err != nil {
			t.Fatal(err)
		}
		p10Points(t, b, "2026-11-01T10:00:00Z", "2026-11-01T10:02:00Z")
		if len(b.Ledger.Facts) != 0 {
			t.Fatal("nonterminal attempt became final")
		}
	})
	t.Run("no_silent_obligation_removal", func(t *testing.T) {
		s := p10Snapshot(t)
		s["Attempts"] = []any{}
		b, err := p10Collect(t, s)
		if err != nil {
			t.Fatal(err)
		}
		if len(b.Inputs.Obligations) != 1 {
			t.Fatal("omitted run erased eligible population")
		}
		p10Points(t, b, "2026-11-01T10:00:00Z")
	})
	t.Run("enumeration_attestation_failure", func(t *testing.T) {
		s := p10Snapshot(t)
		s["EnumerationError"] = "withheld page 2"
		s["Complete"] = false
		b, err := p10Collect(t, s)
		if err == nil && b.Complete {
			t.Fatal("enumeration failure became complete")
		}
	})
}

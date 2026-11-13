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
type p07Provider struct {
	client *http.Client
	url    string
}

func (p p07Provider) get(ctx context.Context, path string, dst any) error {
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
func (p p07Provider) Enumerate(ctx context.Context, _ CollectionRequest) (CollectionSnapshot, error) {
	var v CollectionSnapshot
	err := p.get(ctx, "/enumerate", &v)
	return v, err
}
func (p p07Provider) AuthenticateWriter(ctx context.Context) (WriterIdentity, error) {
	var v WriterIdentity
	err := p.get(ctx, "/writer", &v)
	return v, err
}
func (p p07Provider) ProtectedTip(ctx context.Context, _ string) (ProtectedPolicy, error) {
	var v ProtectedPolicy
	err := p.get(ctx, "/policy", &v)
	return v, err
}
func (p p07Provider) ProveChain(ctx context.Context, _ ChainRequest) (domain.ChainProof, error) {
	var v domain.ChainProof
	err := p.get(ctx, "/chain", &v)
	return v, err
}

var _ Provider = p07Provider{}

func p07Decode(t *testing.T, raw string, dst any) {
	t.Helper()
	if err := json.Unmarshal([]byte(raw), dst); err != nil {
		t.Fatal(err)
	}
}
func p07Collect(t *testing.T, transcript map[string]any) (domain.ShadowBatch, error) {
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
	p07Decode(t, `{"RepositoryID":"repo","CheckID":"security:test","RuleVersion":"v1","Selector":"guardian-trigger-v1/collector-v1","Now":"2026-11-01T12:00:00Z","Deadline":"2026-11-01T11:00:00Z","ConsistencyLag":60000000000}`, &req)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return Collect(ctx, p07Provider{server.Client(), server.URL}, req)
}
func p07Snapshot(t *testing.T) map[string]any {
	t.Helper()
	var v map[string]any
	p07Decode(t, `{"RepositoryID":"repo","ProviderTime":"2026-11-01T12:00:00Z","Complete":true,"EventCompleteThrough":"2026-11-01T12:00:00Z","AttemptCompleteThrough":"2026-11-01T12:00:00Z","PagesComplete":true,"Selector":"guardian-trigger-v1/collector-v1","Trigger":{"Events":["pull_request"],"BaseBranch":"main","Paths":["**/*.go"]},"Events":[{"ID":"E1","RepositoryID":"repo","Scope":"ci","PRNumber":"17","HeadCommit":"H1","EventAt":"2026-11-01T10:00:00Z","Selector":"guardian-trigger-v1/collector-v1","ProviderProof":"event-ledger-page-1","BaseBranch":"main","BaseReachable":true,"Author":"dependabot[bot]","Fork":true,"State":"open","DeadlineAt":"2026-11-01T11:00:00Z"}],"Attempts":[{"Fact":{"ID":{"RepositoryID":"repo","RunID":"1000","JobID":"2000","Attempt":"1"},"Scope":"ci","PRNumber":"17","HeadCommit":"H1","HeadTree":"tree-H1","BaseCommit":"B","MergeBase":"MB","AssessedRevision":"merge-H1","CheckID":"security:test","RuleVersion":"v1","CreatedAt":"2026-11-01T10:01:00Z","StartedAt":"2026-11-01T10:02:00Z","FinishedAt":"2026-11-01T10:03:00Z","Status":"assessed","Comparison":"comparable","Inspection":"inspected","Pins":{"Scout":"scout","Tools":"tools","Catalogue":"catalogue","Policy":"policy","Recipe":"guardian"}},"AccountedEventIDs":["E1"],"Terminal":true,"DeadlineAt":"2026-11-01T11:00:00Z","Lane":"guardian","ConcurrencyGroup":"guardian-pr-17","PinnedConfigDigest":"config","ExecutedConfigDigest":"config","CancelInProgress":true,"ConfigurationVerified":true}],"EnumerationProof":"complete-event-and-attempt-ledger-not-run-list","TimeProof":"provider-job-clock-ms","LagJustification":"fixture-consistency-contract-v1"}`, &v)
	return v
}
func p07Points(t *testing.T, b domain.ShadowBatch, want ...string) {
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

func TestADR012EligibilityAndEventAccounting(t *testing.T) {
	for _, tc := range []struct {
		name, change   string
		wantGap        bool
		wantAccounting int
	}{
		{"fork_bot_zero_findings", "", false, 1},
		{"omitted_failing_eligible_PR", "missing", true, 0},
		{"collector_skipped_unstarted", "skipped", true, 0},
		{"H2_does_not_account_H1", "head", true, 0},
		{"wrong_PR", "pr", true, 0},
		{"start_before_event", "early", true, 0},
		{"closed_unstarted", "closed", true, 0},
		{"merged_unstarted", "merged", true, 0},
		{"deferred_deadline", "deferred", true, 0},
		{"approval_waiting_deadline", "approval", true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := p07Snapshot(t)
			e := s["Events"].([]any)[0].(map[string]any)
			a := s["Attempts"].([]any)[0].(map[string]any)
			f := a["Fact"].(map[string]any)
			switch tc.change {
			case "missing", "skipped", "closed", "merged", "deferred", "approval":
				s["Attempts"] = []any{}
				e["State"] = tc.change
			case "head":
				f["HeadCommit"] = "H2"
			case "pr":
				f["PRNumber"] = "18"
			case "early":
				f["StartedAt"] = "2026-11-01T09:59:59Z"
			}
			b, err := p07Collect(t, s)
			if err != nil {
				t.Fatal(err)
			}
			if len(b.Inputs.Obligations) != 1 {
				t.Fatalf("eligible event disappeared: %+v", b.Inputs)
			}
			if len(b.Inputs.Accounting) != tc.wantAccounting {
				t.Errorf("accounting=%+v want %d", b.Inputs.Accounting, tc.wantAccounting)
			}
			if tc.wantGap {
				p07Points(t, b, "2026-11-01T10:00:00Z")
			} else if len(b.Ledger.Facts) != 1 || len(b.Ledger.Events) != 0 {
				t.Errorf("zero-event fact lost: %+v", b.Ledger)
			}
		})
	}
	t.Run("waiting_alone_not_gap", func(t *testing.T) {
		s := p07Snapshot(t)
		s["Attempts"] = []any{}
		e := s["Events"].([]any)[0].(map[string]any)
		e["State"] = "approval_waiting"
		e["DeadlineAt"] = "2026-11-02T12:00:00Z"
		b, err := p07Collect(t, s)
		if err != nil {
			t.Fatal(err)
		}
		if len(b.Ledger.GapPoints) != 0 || !b.CompleteThrough.Before(time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)) {
			t.Fatalf("waiting event must hold C without inventing a gap: %+v", b)
		}
	})
	for _, reason := range []string{"non_default_base", "unreachable_base"} {
		t.Run(reason, func(t *testing.T) {
			s := p07Snapshot(t)
			e := s["Events"].([]any)[0].(map[string]any)
			if reason == "non_default_base" {
				e["BaseBranch"] = "release"
			} else {
				e["BaseReachable"] = false
			}
			b, err := p07Collect(t, s)
			if err != nil {
				t.Fatal(err)
			}
			if len(b.Ledger.Facts) != 1 || b.Ledger.Facts[0].Status != "not_assessed" || b.Ledger.Facts[0].Comparison != "non_default_base" {
				t.Fatalf("must disclose excluded base: %+v", b.Ledger)
			}
		})
	}
	t.Run("selector_not_pin_establishes_population", func(t *testing.T) {
		a, err := p07Collect(t, p07Snapshot(t))
		if err != nil {
			t.Fatal(err)
		}
		s := p07Snapshot(t)
		s["Attempts"].([]any)[0].(map[string]any)["Fact"].(map[string]any)["Pins"].(map[string]any)["Tools"] = "tools-next"
		b, err := p07Collect(t, s)
		if err != nil {
			t.Fatal(err)
		}
		if a.Inputs.Obligations[0].Selector != b.Inputs.Obligations[0].Selector {
			t.Fatal("pin change reset selector")
		}
		if a.Inputs.Obligations[0].Selector != "guardian-trigger-v1/collector-v1" {
			t.Fatal("population selector omitted")
		}
	})
}

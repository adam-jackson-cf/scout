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
type p12Provider struct {
	client *http.Client
	url    string
}

func (p p12Provider) get(ctx context.Context, path string, dst any) error {
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
func (p p12Provider) Enumerate(ctx context.Context, _ CollectionRequest) (CollectionSnapshot, error) {
	var v CollectionSnapshot
	err := p.get(ctx, "/enumerate", &v)
	return v, err
}
func (p p12Provider) AuthenticateWriter(ctx context.Context) (WriterIdentity, error) {
	var v WriterIdentity
	err := p.get(ctx, "/writer", &v)
	return v, err
}
func (p p12Provider) ProtectedTip(ctx context.Context, _ string) (ProtectedPolicy, error) {
	var v ProtectedPolicy
	err := p.get(ctx, "/policy", &v)
	return v, err
}
func (p p12Provider) ProveChain(ctx context.Context, _ ChainRequest) (domain.ChainProof, error) {
	var v domain.ChainProof
	err := p.get(ctx, "/chain", &v)
	return v, err
}

var _ Provider = p12Provider{}

func p12Decode(t *testing.T, raw string, dst any) {
	t.Helper()
	if err := json.Unmarshal([]byte(raw), dst); err != nil {
		t.Fatal(err)
	}
}
func p12Collect(t *testing.T, transcript map[string]any) (domain.ShadowBatch, error) {
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
	p12Decode(t, `{"RepositoryID":"repo","CheckID":"security:test","RuleVersion":"v1","Selector":"guardian-trigger-v1/collector-v1","Now":"2026-11-01T12:00:00Z","Deadline":"2026-11-01T11:00:00Z","ConsistencyLag":60000000000}`, &req)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return Collect(ctx, p12Provider{server.Client(), server.URL}, req)
}
func p12Snapshot(t *testing.T) map[string]any {
	t.Helper()
	var v map[string]any
	p12Decode(t, `{"RepositoryID":"repo","ProviderTime":"2026-11-01T12:00:00Z","Complete":true,"EventCompleteThrough":"2026-11-01T12:00:00Z","AttemptCompleteThrough":"2026-11-01T12:00:00Z","PagesComplete":true,"Selector":"guardian-trigger-v1/collector-v1","Trigger":{"Events":["pull_request"],"BaseBranch":"main","Paths":["**/*.go"]},"Events":[{"ID":"E1","RepositoryID":"repo","Scope":"ci","PRNumber":"17","HeadCommit":"H1","EventAt":"2026-11-01T10:00:00Z","Selector":"guardian-trigger-v1/collector-v1","ProviderProof":"event-ledger-page-1","BaseBranch":"main","BaseReachable":true,"Author":"dependabot[bot]","Fork":true,"State":"open","DeadlineAt":"2026-11-01T11:00:00Z"}],"Attempts":[{"Fact":{"ID":{"RepositoryID":"repo","RunID":"1000","JobID":"2000","Attempt":"1"},"Scope":"ci","PRNumber":"17","HeadCommit":"H1","HeadTree":"tree-H1","BaseCommit":"B","MergeBase":"MB","AssessedRevision":"merge-H1","CheckID":"security:test","RuleVersion":"v1","CreatedAt":"2026-11-01T10:01:00Z","StartedAt":"2026-11-01T10:02:00Z","FinishedAt":"2026-11-01T10:03:00Z","Status":"assessed","Comparison":"comparable","Inspection":"inspected","Pins":{"Scout":"scout","Tools":"tools","Catalogue":"catalogue","Policy":"policy","Recipe":"guardian"}},"AccountedEventIDs":["E1"],"Terminal":true,"DeadlineAt":"2026-11-01T11:00:00Z","Lane":"guardian","ConcurrencyGroup":"guardian-pr-17","PinnedConfigDigest":"config","ExecutedConfigDigest":"config","CancelInProgress":true,"ConfigurationVerified":true}],"EnumerationProof":"complete-event-and-attempt-ledger-not-run-list","TimeProof":"provider-job-clock-ms","LagJustification":"fixture-consistency-contract-v1"}`, &v)
	return v
}
func p12Points(t *testing.T, b domain.ShadowBatch, want ...string) {
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

// S3 write path only: visibility, dossier digest and ready/not-ready recovery
// are owned by S4, not approximated by a second evaluator here.
func TestADR012S3DeadlineLateLifecycle(t *testing.T) {
	s := p12Snapshot(t)
	e := s["Events"].([]any)[0].(map[string]any)
	e["EventAt"] = "2026-11-01T09:59:00Z"
	e2 := map[string]any{}
	for k, v := range e {
		e2[k] = v
	}
	e2["ID"] = "E2"
	e2["EventAt"] = "2026-11-01T10:01:00Z"
	s["Events"] = []any{e, e2}
	a := s["Attempts"].([]any)[0].(map[string]any)
	a["Terminal"] = false
	a["AccountedEventIDs"] = []string{"E1", "E2"}
	a["Fact"].(map[string]any)["FinishedAt"] = "0001-01-01T00:00:00Z"
	batch, err := p12Collect(t, s)
	if err != nil {
		t.Fatal(err)
	}
	p12Points(t, batch, "2026-11-01T09:59:00Z", "2026-11-01T10:01:00Z", "2026-11-01T10:02:00Z")
	if len(batch.Ledger.Facts) != 0 {
		t.Fatal("deadline manufactured a terminal final")
	}
	var auth PublicationAuthorization
	p12Decode(t, `{"RepositoryID":"repo","ParentCommit":"parent","PublishingJobStartedAt":"2026-11-01T12:01:00Z","CurrentProtectedTip":"protected","Authenticated":true,"SourceReachable":true}`, &auth)
	first, err := PreparePublication(domain.ShadowLedger{}, batch, auth)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Frontiers) != 1 || !first.Frontiers[0].CompleteThrough.After(time.Date(2026, 11, 1, 10, 2, 0, 0, time.UTC)) {
		t.Fatalf("deadline failed to release frontier: %+v", first)
	}
	parent := domain.ShadowLedger{Publications: []domain.ShadowPublication{first}, Frontiers: first.Frontiers, GapPoints: first.GapPoints}
	a["Terminal"] = true
	a["Fact"].(map[string]any)["FinishedAt"] = "2026-11-01T12:02:00Z"
	a["ScannerOutput"] = map[string]any{"Status": "assessed", "Events": []any{map[string]any{"IssueID": "late-finding", "CheckID": "security:test", "RuleVersion": "v1"}}}
	s["ProviderTime"] = "2026-11-01T12:04:00Z"
	s["EventCompleteThrough"] = "2026-11-01T12:04:00Z"
	s["AttemptCompleteThrough"] = "2026-11-01T12:04:00Z"
	later, err := p12Collect(t, s)
	if err != nil {
		t.Fatal(err)
	}
	auth.PublishingJobStartedAt = time.Date(2026, 11, 1, 12, 3, 0, 0, time.UTC)
	final, err := PreparePublication(parent, later, auth)
	if err != nil {
		t.Fatal(err)
	}
	if len(final.Facts) != 1 || final.Facts[0].RecordedAt == nil || !final.Facts[0].RecordedAt.After(first.Frontiers[0].CompleteThrough) {
		t.Fatalf("late terminal not independently stamped: %+v", final)
	}
	if !final.Facts[0].StartedAt.Equal(time.Date(2026, 11, 1, 10, 2, 0, 0, time.UTC)) {
		t.Fatal("late lifecycle rewrote provider coordinate")
	}
	if len(final.Events) != 1 || len(final.Links) != 1 || final.Events[0].RecordedAt == nil || !final.Events[0].RecordedAt.Equal(final.RecordedAt) || final.Links[0].AttemptID != final.Facts[0].ID {
		t.Fatalf("late event/link not published with terminal fact: %+v", final)
	}
}

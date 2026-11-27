//go:build boundary

package prmeta

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"scout/internal/history"
)

const secret = "ghp_SECRETTOKEN123"

func graphqlPage(first, n int, next bool) string {
	var nodes []string
	for i := first; i < first+n; i++ {
		nodes = append(nodes, fmt.Sprintf(`{"number":%d,"createdAt":"2024-02-01T00:00:00Z","mergedAt":"2024-02-01T10:00:00Z","author":{"__typename":"Bot","login":"dependabot"},
"files":{"totalCount":1,"nodes":[{"path":"a/f.go","additions":2,"deletions":1}]},
"commits":{"totalCount":1,"nodes":[{"commit":{"authoredDate":"2024-02-01T09:00:00Z"}}]},"reviews":{"totalCount":2},"comments":{"totalCount":3}}`, i))
	}
	return fmt.Sprintf(`{"data":{"repository":{"pullRequests":{"pageInfo":{"hasNextPage":%t,"endCursor":"c%d"},"nodes":[%s]}}}}`, next, first, strings.Join(nodes, ","))
}

func TestFetchIsOptInAndNeverLeaksTheToken(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Error(w, "denied "+r.Header.Get("Authorization"), http.StatusForbidden)
	}))
	defer srv.Close()
	base := FetchOptions{Repo: "o/r", Token: secret, Endpoint: srv.URL}

	if _, _, err := Fetch(context.Background(), base); err == nil || calls != 0 {
		t.Fatalf("disabled fetch must fail without any request (calls=%d, err=%v)", calls, err)
	}
	base.Enabled = true
	noToken := base
	noToken.Token = ""
	if _, _, err := Fetch(context.Background(), noToken); err == nil || calls != 0 {
		t.Fatalf("missing token must fail without a request (calls=%d)", calls)
	}
	_, _, err := Fetch(context.Background(), base)
	if err == nil || strings.Contains(err.Error(), secret) || !strings.Contains(err.Error(), "403") {
		t.Fatalf("HTTP failure error = %v", err)
	}
	srv.Close()
	_, _, connectionRefusedErr := Fetch(context.Background(), base)
	if connectionRefusedErr == nil || strings.Contains(connectionRefusedErr.Error(), secret) {
		t.Fatalf("transport failure error = %v", connectionRefusedErr)
	}
}

func TestFetchBuildsDocumentWithProvenance(t *testing.T) {
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(string(body), `"after":"c1"`) {
			_, _ = io.WriteString(w, graphqlPage(3, 1, false))
			return
		}
		_, _ = io.WriteString(w, graphqlPage(1, 2, true))
	}))
	defer srv.Close()
	res, raw, err := Fetch(context.Background(), FetchOptions{
		Enabled: true, Repo: "o/r", Token: secret, Endpoint: srv.URL, Limit: 3,
		Now: func() time.Time { return time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatal(err)
	}
	if auth != "Bearer "+secret {
		t.Error("token not sent as bearer credential")
	}
	if strings.Contains(string(raw), secret) {
		t.Error("document contains the token")
	}
	checkFetchedSource(t, res)
	checkFetchedPR(t, res.PRs[0])
}

func checkFetchedSource(t *testing.T, res Result) {
	t.Helper()
	if len(res.PRs) != 3 || res.Source.Kind != "github_api" || res.Source.Origin != "o/r" ||
		res.Source.FetchedAt != "2024-06-01T00:00:00Z" || res.Source.Digest == "" {
		t.Fatalf("result = %+v", res)
	}
}

func checkFetchedPR(t *testing.T, p history.PR) {
	t.Helper()
	if p.AuthorType != "Bot" || p.Reviews != 2 || p.Comments != 3 || len(p.CommitTimes) != 1 || p.Files[0].Added != 2 {
		t.Fatalf("PR = %+v", p)
	}
}

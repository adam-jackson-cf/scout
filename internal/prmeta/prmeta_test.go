package prmeta

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"scout/internal/domain"
	"scout/internal/history"
)

var q1 = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

// fixtureDoc has n merged PRs in 2024-Q1, each with three commits whose
// author-timestamp gaps are 1 h and 1.5 h: active hours 0.5+1+1.5 = 3.0.
func fixtureDoc(n int) Document {
	doc := Document{Schema: Schema, Repository: "o/r", FetchedAt: "2024-06-01T00:00:00Z"}
	for i := 1; i <= n; i++ {
		m := q1.Add(time.Duration(10+i) * 24 * time.Hour)
		at := func(d time.Duration) string { return m.Add(d).Format(time.RFC3339) }
		doc.PullRequests = append(doc.PullRequests, PR{
			Number: i, AuthorLogin: "dev", CreatedAt: at(-10 * time.Hour), MergedAt: at(0),
			Files:      []File{{Path: "a/f.go", Additions: 3, Deletions: 1}, {Path: "b/g.go", Additions: 2}},
			TotalFiles: 2, TotalCommits: 3, Reviews: 1, Comments: 1,
			Commits: []Commit{{at(-3 * time.Hour)}, {at(-2 * time.Hour)}, {at(-30 * time.Minute)}},
		})
	}
	return doc
}

func writeDoc(t *testing.T, doc Document) (string, []byte) {
	t.Helper()
	raw, err := json.MarshalIndent(doc, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "pr-metadata.json")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	return path, raw
}

// 25 non-merge commits in scope "a" so churn distributions are measured.
func commits() []history.Commit {
	var out []history.Commit
	for i := range 25 {
		out = append(out, history.Commit{
			Hash: fmt.Sprintf("%040d", i), AuthorName: "dev", AuthorEmail: "d@example.com",
			Time: q1.Add(time.Duration(i+1) * 24 * time.Hour), Parents: 1,
			Files: []history.FileChange{{Path: "a/f.go", Added: i%5 + 1, Removed: 1}},
		})
	}
	return out
}

func compute(prs []history.PR, src *history.PRSource) []domain.Signal {
	sc := history.PrefixScope(domain.SignalScopeDirectory, "a", "a")
	return history.Compute(commits(), history.Options{
		Scopes: []history.Scope{sc}, End: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		Snapshot: "s", PRs: prs, PRSource: src,
	})
}

func pick(t *testing.T, sigs []domain.Signal, id string) domain.Signal {
	t.Helper()
	for _, s := range sigs {
		if s.ID() == id && s.ScopeAnchor == "a" && s.Window == "2024-Q1" {
			return s
		}
	}
	t.Fatalf("no %s for scope a in 2024-Q1", id)
	return domain.Signal{}
}

func TestMetadataFilePresentSignalsCarryProvenance(t *testing.T) {
	path, raw := writeDoc(t, fixtureDoc(6))
	res, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	wantDigest := hex.EncodeToString(sum[:])
	if res.Source.Digest != wantDigest || res.Source.Kind != "file" || res.Source.Count != 6 {
		t.Fatalf("source = %+v", res.Source)
	}
	sigs := compute(res.PRs, &res.Source)

	effort := pick(t, sigs, "scout.pr_effort@1")
	if effort.Missing != "" || effort.Value == nil || *effort.Value != 3.0 {
		t.Fatalf("pr_effort = %+v", effort)
	}
	batch := pick(t, sigs, "scout.batch_size@1")
	if batch.Missing != "" || batch.Value == nil || *batch.Value != 6 {
		t.Fatalf("batch_size = %+v", batch)
	}
	for _, s := range []domain.Signal{effort, batch} {
		p, _ := s.Detail["provenance"].(map[string]any)
		if p["digest"] != wantDigest || p["kind"] != "file" || p["origin"] != "pr-metadata.json" {
			t.Errorf("%s provenance = %v", s.ID(), s.Detail["provenance"])
		}
	}
	disp := pick(t, sigs, "scout.dispersion@1")
	if disp.Detail["pr_p90_over_p50"] != 1.0 {
		t.Errorf("dispersion PR part = %v", disp.Detail)
	}
	if p, _ := disp.Detail["pr_provenance"].(map[string]any); p["digest"] != wantDigest {
		t.Errorf("dispersion provenance = %v", disp.Detail["pr_provenance"])
	}
}

func TestWithoutMetadataPRSignalsAreUnmeasuredWithReason(t *testing.T) {
	sigs := compute(nil, nil)
	for _, id := range []string{"scout.pr_effort@1", "scout.batch_size@1"} {
		s := pick(t, sigs, id)
		if s.Missing != domain.SignalUnmeasured || s.MissingReason == "" || s.Value != nil {
			t.Errorf("%s = %+v", id, s)
		}
	}
	d := pick(t, sigs, "scout.dispersion@1")
	if d.Value == nil || d.Detail["pr_part"] != "unmeasured" || d.Detail["pr_part_reason"] == "" {
		t.Errorf("dispersion = %+v", d)
	}
	if _, ok := d.Detail["pr_p90_over_p50"]; ok {
		t.Error("PR part must be absent, not zero")
	}
}

func TestMetadataChangesConfigDigest(t *testing.T) {
	path, _ := writeDoc(t, fixtureDoc(6))
	res, _ := LoadFile(path)
	with := pick(t, compute(res.PRs, &res.Source), "scout.dispersion@1")
	without := pick(t, compute(nil, nil), "scout.dispersion@1")
	if with.ConfigDigest == without.ConfigDigest {
		t.Error("a later ingestion must not be dropped as a duplicate of the git-only signal")
	}
}

func TestParseRejectsInvalidDocuments(t *testing.T) {
	good := fixtureDoc(1)
	bad := map[string]func(*Document){
		"schema":    func(d *Document) { d.Schema = "x" },
		"duplicate": func(d *Document) { d.PullRequests = append(d.PullRequests, d.PullRequests[0]) },
		"time":      func(d *Document) { d.PullRequests[0].CreatedAt = "yesterday" },
		"totals":    func(d *Document) { d.PullRequests[0].TotalFiles = 1 },
	}
	for name, mod := range bad {
		doc := good
		doc.PullRequests = append([]PR(nil), good.PullRequests...)
		mod(&doc)
		raw, _ := json.Marshal(doc)
		if _, err := Parse(raw, "file", "x"); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if _, err := Parse([]byte(`{"schema":"`+Schema+`","pull_requests":[],"extra":1}`), "file", "x"); err == nil {
		t.Error("unknown field accepted")
	}
	doc := fixtureDoc(2)
	doc.PullRequests[1].MergedAt = ""
	raw, _ := json.Marshal(doc)
	res, err := Parse(raw, "file", "x")
	if err != nil || len(res.PRs) != 1 || res.Unmerged != 1 {
		t.Errorf("unmerged PR handling: %+v %v", res, err)
	}
}

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
	_, _, err = Fetch(context.Background(), base) // connection refused
	if err == nil || strings.Contains(err.Error(), secret) {
		t.Fatalf("transport failure error = %v", err)
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
	if len(res.PRs) != 3 || res.Source.Kind != "github_api" || res.Source.Origin != "o/r" ||
		res.Source.FetchedAt != "2024-06-01T00:00:00Z" || res.Source.Digest == "" {
		t.Fatalf("result = %+v", res)
	}
	p := res.PRs[0]
	if p.AuthorType != "Bot" || p.Reviews != 2 || p.Comments != 3 || len(p.CommitTimes) != 1 || p.Files[0].Added != 2 {
		t.Fatalf("PR = %+v", p)
	}
}

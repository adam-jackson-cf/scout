//go:build boundary

package shadow

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var githubControlPlaneResponses = map[string]string{
	"/repos/o/r":                           `{"id":9007199254740993,"default_branch":"main"}`,
	"/repos/o/r/actions/runs/999":          `{"id":999,"head_sha":"source","head_branch":"main","actor":{"login":"alice"},"triggering_actor":{"login":"alice"},"repository":{"id":9007199254740993},"head_repository":{"id":9007199254740993}}`,
	"/repos/o/r/branches/main":             `{"protected":true,"commit":{"sha":"tip"}}`,
	"/repos/o/r/git/ref/heads/scout/state": `{"object":{"sha":"parent"}}`,
	"/repos/o/r/compare/source...tip":      `{"status":"ahead","merge_base_commit":{"sha":"source"}}`,
	"/repos/o/r/compare/genesis...parent":  `{"status":"ahead","merge_base_commit":{"sha":"genesis"}}`,
	"/repos/o/r/compare/parent...import":   `{"status":"ahead","merge_base_commit":{"sha":"parent"}}`,
}

func githubControlPlaneHandler(t *testing.T, requests *[]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		*requests = append(*requests, r.URL.RequestURI())
		if r.Header.Get("Authorization") != "Bearer fixture-only" {
			t.Error("missing authentication header")
		}
		switch r.URL.Path {
		case "/repos/o/r/actions/runs":
			w.Header().Set("Date", "Sun, 01 Nov 2026 12:00:00 GMT")
			w.Header().Set("Link", `<https://example.invalid/page2>; rel="next"`)
			_, _ = w.Write([]byte(`{"total_count":1000,"workflow_runs":[]}`))
		case "/repos/o/r/contents/.scout/policy.yaml":
			if r.URL.Query().Get("ref") != "tip" {
				t.Error("policy not pinned to protected tip")
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"encoding": "base64", "content": base64.StdEncoding.EncodeToString([]byte("maintainers: [alice]\n"))})
		case "/repos/o/r/contents/.github/CODEOWNERS":
			_ = json.NewEncoder(w).Encode(map[string]string{"encoding": "base64", "content": base64.StdEncoding.EncodeToString([]byte("* @bob\n"))})
		default:
			body, known := githubControlPlaneResponses[r.URL.Path]
			if !known {
				t.Errorf("unexpected path %s", r.URL.Path)
				http.NotFound(w, r)
				return
			}
			_, _ = w.Write([]byte(body))
		}
	}
}

func TestGitHubAuthenticatedControlPlane(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5e9)
	defer cancel()
	var requests []string
	server := httptest.NewServer(githubControlPlaneHandler(t, &requests))
	defer server.Close()
	g := GitHub{Client: server.Client(), Endpoint: server.URL, Repository: "o/r", RepositoryID: "9007199254740993", RunID: "999", Token: "fixture-only"}
	assertGitHubEnumerationIsNotComplete(ctx, t, g)
	assertGitHubWriterProvenance(ctx, t, g)
	assertGitHubProtectedPolicy(ctx, t, g)
	assertGitHubChainProof(ctx, t, g)
	if len(requests) == 0 {
		t.Fatal("adapter made no HTTP requests")
	}
}

func assertGitHubEnumerationIsNotComplete(ctx context.Context, t *testing.T, g GitHub) {
	t.Helper()
	snap, err := g.Enumerate(ctx, CollectionRequest{RepositoryID: g.RepositoryID, Selector: "selector"})
	if err != nil {
		t.Fatal(err)
	}
	if snap.Complete || !snap.ProviderTime.IsZero() || snap.EnumerationError == "" {
		t.Fatalf("run list or HTTP Date fabricated completeness: %+v", snap)
	}
}

func assertGitHubWriterProvenance(ctx context.Context, t *testing.T, g GitHub) {
	t.Helper()
	writer, err := g.AuthenticateWriter(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !writer.Authenticated || !writer.SourceReachable || !writer.DefaultBranch || writer.OriginalActor != "alice" || writer.ExecutingActor != "alice" {
		t.Fatalf("writer provenance=%+v", writer)
	}
}

func assertGitHubProtectedPolicy(ctx context.Context, t *testing.T, g GitHub) {
	t.Helper()
	p, err := g.ProtectedTip(ctx, g.RepositoryID)
	if err != nil {
		t.Fatal(err)
	}
	if !p.Authenticated || p.Revision != "tip" || !authorizedActor("bob", p) || !authorizedActor("alice", p) || p.MinDays != 14 || p.MinObservations != 5 {
		t.Fatalf("policy=%+v", p)
	}
}

func assertGitHubChainProof(ctx context.Context, t *testing.T, g GitHub) {
	t.Helper()
	proof, err := g.ProveChain(ctx, ChainRequest{RepositoryID: g.RepositoryID, StateRef: "refs/heads/scout/state", GenesisCommit: "genesis", DestinationTip: "parent", SourceTip: "import"})
	if err != nil {
		t.Fatal(err)
	}
	if !proof.Authenticated || !proof.DestinationAncestor {
		t.Fatalf("chain=%+v", proof)
	}
}

func TestGitHubRefusesUnattestedResponses(t *testing.T) {
	for _, scenario := range []string{"identity", "permission", "malformed", "redirect"} {
		t.Run(scenario, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch scenario {
				case "identity":
					_, _ = w.Write([]byte(`{"id":2,"default_branch":"main"}`))
				case "permission":
					http.Error(w, "private provider text", http.StatusForbidden)
				case "malformed":
					_, _ = w.Write([]byte(`{`))
				case "redirect":
					w.Header().Set("Location", "https://example.invalid")
					w.WriteHeader(http.StatusFound)
				}
			}))
			defer server.Close()
			g := GitHub{Client: server.Client(), Endpoint: server.URL, Repository: "o/r", RepositoryID: "1", Token: "fixture-only"}
			ctx, cancel := context.WithTimeout(context.Background(), 5e9)
			defer cancel()
			if _, err := g.Enumerate(ctx, CollectionRequest{RepositoryID: "1"}); err == nil || strings.Contains(err.Error(), "fixture-only") || strings.Contains(err.Error(), "private provider text") {
				t.Fatalf("expected token-safe refusal: %v", err)
			}
		})
	}
}

func TestGitHubRefusesForkWriterOnNamedDefaultBranch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/o/r" {
			_, _ = w.Write([]byte(`{"id":1,"default_branch":"main"}`))
			return
		}
		_, _ = w.Write([]byte(`{"id":999,"head_sha":"source","head_branch":"main","actor":{"login":"alice"},"triggering_actor":{"login":"alice"},"repository":{"id":1},"head_repository":{"id":2}}`))
	}))
	defer server.Close()
	g := GitHub{Client: server.Client(), Endpoint: server.URL, Repository: "o/r", RepositoryID: "1", RunID: "999", Token: "fixture-only"}
	ctx, cancel := context.WithTimeout(context.Background(), 5e9)
	defer cancel()
	if _, err := g.AuthenticateWriter(ctx); err == nil {
		t.Fatal("fork branch named main acquired default-branch writer identity")
	}
}

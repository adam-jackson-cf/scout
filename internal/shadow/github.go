package shadow

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"scout/internal/domain"
	"scout/internal/policy"
)

// GitHub uses only authenticated REST responses. Run listing is deliberately
// not an eligible-event ledger: GitHub cannot attest skipped/unstarted pushes
// from it, so this adapter cannot yet establish a completeness frontier.
// Production event capture and action transport require the wiring ADR.
type GitHub struct {
	Client                                           *http.Client
	Endpoint, Repository, RepositoryID, RunID, Token string
}

type githubHTTPError int

func (e githubHTTPError) Error() string { return fmt.Sprintf("GitHub API HTTP %d", e) }

func (g GitHub) get(ctx context.Context, path string, out any) error {
	endpoint := g.Endpoint
	if endpoint == "" {
		endpoint = "https://api.github.com"
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" {
		return fmt.Errorf("invalid GitHub API endpoint")
	}
	if g.Token == "" {
		return fmt.Errorf("GitHub authentication required")
	}
	client := g.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	// Never forward authentication through redirects, including a caller client.
	bounded := *client
	bounded.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(endpoint, "/")+path, nil)
	if err != nil {
		return fmt.Errorf("invalid GitHub API request")
	}
	req.Header.Set("Authorization", "Bearer "+g.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	res, err := bounded.Do(req)
	if err != nil {
		return fmt.Errorf("GitHub API transport unavailable")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return githubHTTPError(res.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, 8<<20+1))
	if err != nil || len(data) > 8<<20 {
		return fmt.Errorf("GitHub API response unavailable or oversized")
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("invalid GitHub API response")
	}
	return nil
}
func (g GitHub) repositoryPath() (string, error) {
	parts := strings.Split(g.Repository, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("GitHub repository must be owner/name")
	}
	return "/repos/" + url.PathEscape(parts[0]) + "/" + url.PathEscape(parts[1]), nil
}

type githubRepository struct {
	ID            json.Number `json:"id"`
	DefaultBranch string      `json:"default_branch"`
}

func (g GitHub) repository(ctx context.Context) (string, githubRepository, error) {
	path, err := g.repositoryPath()
	if err != nil {
		return "", githubRepository{}, err
	}
	var r githubRepository
	if err := g.get(ctx, path, &r); err != nil {
		return "", r, err
	}
	if r.ID.String() == "" || r.ID.String() != g.RepositoryID || r.DefaultBranch == "" {
		return "", r, fmt.Errorf("GitHub repository identity mismatch")
	}
	return path, r, nil
}
func (g GitHub) Enumerate(ctx context.Context, req CollectionRequest) (CollectionSnapshot, error) {
	path, r, err := g.repository(ctx)
	if err != nil {
		return CollectionSnapshot{}, err
	}
	if req.RepositoryID != r.ID.String() {
		return CollectionSnapshot{}, fmt.Errorf("collection repository mismatch")
	}
	var runs struct {
		TotalCount json.Number       `json:"total_count"`
		Runs       []json.RawMessage `json:"workflow_runs"`
	}
	if err := g.get(ctx, path+"/actions/runs?per_page=100", &runs); err != nil {
		return CollectionSnapshot{}, err
	}
	return CollectionSnapshot{RepositoryID: r.ID.String(), Selector: req.Selector, ListingCapped: true, EnumerationError: "GitHub Actions run listing does not attest complete PR-event or attempt enumeration"}, nil
}
func (g GitHub) AuthenticateWriter(ctx context.Context) (WriterIdentity, error) {
	path, r, err := g.repository(ctx)
	if err != nil {
		return WriterIdentity{}, err
	}
	if g.RunID == "" {
		return WriterIdentity{}, fmt.Errorf("writer run identity required")
	}
	var run struct {
		ID         json.Number `json:"id"`
		HeadSHA    string      `json:"head_sha"`
		HeadBranch string      `json:"head_branch"`
		Actor      struct {
			Login string `json:"login"`
		} `json:"actor"`
		TriggeringActor struct {
			Login string `json:"login"`
		} `json:"triggering_actor"`
		Repository struct {
			ID json.Number `json:"id"`
		} `json:"repository"`
		HeadRepository struct {
			ID json.Number `json:"id"`
		} `json:"head_repository"`
	}
	if err := g.get(ctx, path+"/actions/runs/"+url.PathEscape(g.RunID), &run); err != nil {
		return WriterIdentity{}, err
	}
	if run.ID.String() != g.RunID || run.Repository.ID.String() != g.RepositoryID || run.HeadRepository.ID.String() != g.RepositoryID || run.Actor.Login == "" || run.TriggeringActor.Login == "" || run.HeadSHA == "" {
		return WriterIdentity{}, fmt.Errorf("writer run provenance incomplete")
	}
	var branch struct {
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}
	if err := g.get(ctx, path+"/branches/"+url.PathEscape(r.DefaultBranch), &branch); err != nil {
		return WriterIdentity{}, err
	}
	reachable, err := g.ancestor(ctx, path, run.HeadSHA, branch.Commit.SHA)
	if err != nil {
		return WriterIdentity{}, err
	}
	return WriterIdentity{RepositoryID: g.RepositoryID, OriginalActor: run.Actor.Login, ExecutingActor: run.TriggeringActor.Login, SourceRevision: run.HeadSHA, DefaultBranch: run.HeadBranch == r.DefaultBranch, SourceReachable: reachable, Authenticated: true}, nil
}
func (g GitHub) content(ctx context.Context, path, file, ref string) ([]byte, error) {
	var c struct{ Content, Encoding string }
	if err := g.get(ctx, path+"/contents/"+file+"?ref="+url.QueryEscape(ref), &c); err != nil {
		return nil, err
	}
	if c.Encoding != "base64" {
		return nil, fmt.Errorf("unsupported GitHub content encoding")
	}
	raw, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(c.Content, "\n", ""))
	if err != nil {
		return nil, fmt.Errorf("invalid GitHub content encoding")
	}
	return raw, nil
}
func (g GitHub) ProtectedTip(ctx context.Context, id string) (ProtectedPolicy, error) {
	path, r, err := g.repository(ctx)
	if err != nil {
		return ProtectedPolicy{}, err
	}
	if id != g.RepositoryID {
		return ProtectedPolicy{}, fmt.Errorf("policy repository mismatch")
	}
	var branch struct {
		Protected bool `json:"protected"`
		Commit    struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}
	if err := g.get(ctx, path+"/branches/"+url.PathEscape(r.DefaultBranch), &branch); err != nil {
		return ProtectedPolicy{}, err
	}
	if !branch.Protected || branch.Commit.SHA == "" {
		return ProtectedPolicy{}, fmt.Errorf("protected default tip not attested")
	}
	raw, err := g.content(ctx, path, ".scout/policy.yaml", branch.Commit.SHA)
	if err != nil {
		return ProtectedPolicy{}, err
	}
	parsed, err := policy.Parse(raw)
	if err != nil {
		return ProtectedPolicy{}, err
	}
	var declared struct {
		Maintainers []string `yaml:"maintainers"`
	}
	if err := yaml.Unmarshal(raw, &declared); err != nil {
		return ProtectedPolicy{}, err
	}
	var owners []byte
	for _, file := range []string{".github/CODEOWNERS", "CODEOWNERS", "docs/CODEOWNERS"} {
		owners, err = g.content(ctx, path, file, branch.Commit.SHA)
		if err == nil {
			break
		}
		var status githubHTTPError
		if !errors.As(err, &status) || status != http.StatusNotFound {
			return ProtectedPolicy{}, err
		}
	}
	return ProtectedPolicy{RepositoryID: id, Revision: branch.Commit.SHA, Maintainers: declared.Maintainers, CODEOWNERS: string(owners), MinDays: parsed.Shadow().MinDays, MinObservations: parsed.Shadow().MinPRRuns, Authenticated: true}, nil
}
func (g GitHub) ancestor(ctx context.Context, path, base, head string) (bool, error) {
	if base == "" || head == "" {
		return false, fmt.Errorf("ancestry endpoints required")
	}
	var comparison struct {
		Status    string `json:"status"`
		MergeBase struct {
			SHA string `json:"sha"`
		} `json:"merge_base_commit"`
	}
	if err := g.get(ctx, path+"/compare/"+url.PathEscape(base)+"..."+url.PathEscape(head), &comparison); err != nil {
		return false, err
	}
	return (comparison.Status == "ahead" || comparison.Status == "identical") && comparison.MergeBase.SHA == base, nil
}
func (g GitHub) ProveChain(ctx context.Context, req ChainRequest) (domain.ChainProof, error) {
	path, _, err := g.repository(ctx)
	if err != nil {
		return domain.ChainProof{}, err
	}
	if req.RepositoryID != g.RepositoryID || req.StateRef != "refs/heads/scout/state" || req.GenesisCommit == "" {
		return domain.ChainProof{}, fmt.Errorf("state chain identity required")
	}
	var ref struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	if err := g.get(ctx, path+"/git/ref/heads/scout/state", &ref); err != nil {
		return domain.ChainProof{}, err
	}
	if ref.Object.SHA != req.DestinationTip {
		return domain.ChainProof{}, fmt.Errorf("state ref moved from requested parent")
	}
	genesis, err := g.ancestor(ctx, path, req.GenesisCommit, req.DestinationTip)
	if err != nil {
		return domain.ChainProof{}, err
	}
	forward, err := g.ancestor(ctx, path, req.DestinationTip, req.SourceTip)
	if err != nil {
		return domain.ChainProof{}, err
	}
	return domain.ChainProof{RepositoryID: req.RepositoryID, StateRef: req.StateRef, GenesisCommit: req.GenesisCommit, SourceTip: req.SourceTip, DestinationTip: req.DestinationTip, Authenticated: genesis, DestinationAncestor: genesis && forward}, nil
}

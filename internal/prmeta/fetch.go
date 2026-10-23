package prmeta

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultEndpoint is GitHub's GraphQL endpoint.
const DefaultEndpoint = "https://api.github.com/graphql"

// FetchOptions configures the opt-in GitHub fetch. Fetching is never implicit:
// callers must set Enabled (CLI: --fetch-pr-metadata). The token is sent only
// in the Authorization header and appears in no error or log text.
type FetchOptions struct {
	Enabled  bool
	Repo     string // owner/name
	Token    string // GITHUB_TOKEN; needs pull-requests: read
	Limit    int    // newest merged PRs to fetch; 0 = 100
	Endpoint string // empty = DefaultEndpoint
	Client   *http.Client
	Now      func() time.Time
}

const query = `query($owner:String!,$name:String!,$n:Int!,$after:String){repository(owner:$owner,name:$name){
pullRequests(states:MERGED,first:$n,after:$after,orderBy:{field:UPDATED_AT,direction:DESC}){
pageInfo{hasNextPage endCursor}
nodes{number createdAt mergedAt author{__typename login}
files(first:50){totalCount nodes{path additions deletions}}
commits(first:100){totalCount nodes{commit{authoredDate}}}
reviews{totalCount} comments{totalCount}}}}}`

type gqlResponse struct {
	Data struct {
		Repository *struct {
			PullRequests struct {
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
				Nodes []struct {
					Number    int    `json:"number"`
					CreatedAt string `json:"createdAt"`
					MergedAt  string `json:"mergedAt"`
					Author    *struct {
						Typename string `json:"__typename"`
						Login    string `json:"login"`
					} `json:"author"`
					Files struct {
						TotalCount int    `json:"totalCount"`
						Nodes      []File `json:"nodes"`
					} `json:"files"`
					Commits struct {
						TotalCount int `json:"totalCount"`
						Nodes      []struct {
							Commit struct {
								AuthoredDate string `json:"authoredDate"`
							} `json:"commit"`
						} `json:"nodes"`
					} `json:"commits"`
					Reviews  struct{ TotalCount int } `json:"reviews"`
					Comments struct{ TotalCount int } `json:"comments"`
				} `json:"nodes"`
			} `json:"pullRequests"`
		} `json:"repository"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// Fetch downloads merged-PR metadata and returns it as a validated Result
// plus the canonical document bytes (callers may archive them).
func Fetch(ctx context.Context, o FetchOptions) (Result, []byte, error) {
	if !o.Enabled {
		return Result{}, nil, errors.New("PR metadata fetch is not enabled")
	}
	owner, name, ok := strings.Cut(o.Repo, "/")
	if !ok || owner == "" || name == "" || strings.Contains(name, "/") {
		return Result{}, nil, fmt.Errorf("repository %q is not owner/name", o.Repo)
	}
	if o.Token == "" {
		return Result{}, nil, errors.New("GITHUB_TOKEN is not set")
	}
	limit := o.Limit
	if limit <= 0 {
		limit = 100
	}
	endpoint, client, now := o.Endpoint, o.Client, o.Now
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	if now == nil {
		now = time.Now
	}
	doc := Document{Schema: Schema, Repository: o.Repo, FetchedAt: now().UTC().Format(time.RFC3339)}
	after := ""
	for len(doc.PullRequests) < limit {
		page := min(25, limit-len(doc.PullRequests))
		resp, err := post(ctx, client, endpoint, o.Token, map[string]any{
			"query": query, "variables": map[string]any{"owner": owner, "name": name, "n": page, "after": nilIfEmpty(after)},
		})
		if err != nil {
			return Result{}, nil, err
		}
		if len(resp.Errors) > 0 {
			return Result{}, nil, fmt.Errorf("GitHub GraphQL error: %s", resp.Errors[0].Message)
		}
		if resp.Data.Repository == nil {
			return Result{}, nil, fmt.Errorf("repository %s not found or not readable", o.Repo)
		}
		prs := resp.Data.Repository.PullRequests
		for _, n := range prs.Nodes {
			p := PR{
				Number: n.Number, CreatedAt: n.CreatedAt, MergedAt: n.MergedAt,
				Files: n.Files.Nodes, TotalFiles: n.Files.TotalCount, TotalCommits: n.Commits.TotalCount,
				Reviews: n.Reviews.TotalCount, Comments: n.Comments.TotalCount,
			}
			if n.Author != nil {
				p.AuthorLogin = n.Author.Login
				if n.Author.Typename == "Bot" {
					p.AuthorType = "Bot"
				}
			}
			for _, c := range n.Commits.Nodes {
				p.Commits = append(p.Commits, Commit{AuthoredAt: c.Commit.AuthoredDate})
			}
			doc.PullRequests = append(doc.PullRequests, p)
		}
		if !prs.PageInfo.HasNextPage || len(prs.Nodes) == 0 {
			break
		}
		after = prs.PageInfo.EndCursor
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		return Result{}, nil, err
	}
	res, err := Parse(raw, "github_api", o.Repo)
	return res, raw, err
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func post(ctx context.Context, client *http.Client, endpoint, token string, body any) (*gqlResponse, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		return nil, errors.New("invalid GitHub endpoint")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		// Do not wrap err: net/http errors embed the URL; keep the text token-free.
		return nil, errors.New("GitHub request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub returned HTTP %d", resp.StatusCode)
	}
	var out gqlResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 32<<20)).Decode(&out); err != nil {
		return nil, errors.New("GitHub response is not valid JSON")
	}
	return &out, nil
}

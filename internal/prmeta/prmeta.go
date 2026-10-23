// Package prmeta ingests merged-PR metadata for the PR-level measures of
// design §8.3 (pr_effort, batch_size, the active-hours part of dispersion).
//
// Two sources produce the same document: a JSON file supplied with
// --pr-metadata, and an opt-in GitHub GraphQL fetch (see Fetch). Both yield
// history.PR values and a history.PRSource (kind, origin, SHA-256 of the
// document). Without a source the measures stay `unmeasured`.
package prmeta

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"scout/internal/history"
)

// Schema is the document schema id.
const Schema = "scout.pr-metadata@1"

// Document is the on-disk / on-wire PR metadata format.
type Document struct {
	Schema       string `json:"schema"`
	Repository   string `json:"repository,omitempty"` // owner/name
	FetchedAt    string `json:"fetched_at,omitempty"` // RFC 3339
	PullRequests []PR   `json:"pull_requests"`
}

// PR is one pull request. Unmerged PRs are ignored on load.
type PR struct {
	Number       int      `json:"number"`
	AuthorLogin  string   `json:"author_login"`
	AuthorType   string   `json:"author_type,omitempty"` // "Bot" for GitHub App identities
	CreatedAt    string   `json:"created_at"`
	MergedAt     string   `json:"merged_at,omitempty"`
	Files        []File   `json:"files"`
	TotalFiles   int      `json:"total_files"`
	Commits      []Commit `json:"commits"`
	TotalCommits int      `json:"total_commits"`
	Reviews      int      `json:"reviews"`
	Comments     int      `json:"comments"`
}

// File is one changed file of a PR.
type File struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

// Commit carries the author timestamp of one PR commit.
type Commit struct {
	AuthoredAt string `json:"authored_at"`
}

// Result is a validated document converted for history.Options.
type Result struct {
	PRs    []history.PR
	Source history.PRSource
	// Unmerged counts PRs without merged_at that were ignored.
	Unmerged int
}

// LoadFile reads and validates a PR metadata JSON file.
func LoadFile(path string) (Result, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Result{}, fmt.Errorf("read PR metadata: %w", err)
	}
	return Parse(raw, "file", filepath.Base(path))
}

// Parse validates a document and converts it. Unknown fields, a wrong schema
// id, duplicate PR numbers and unparseable timestamps are errors: silently
// dropping rows would change the measure.
func Parse(raw []byte, kind, origin string) (Result, error) {
	var doc Document
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&doc); err != nil {
		return Result{}, fmt.Errorf("PR metadata is not a valid %s document: %w", Schema, err)
	}
	if dec.More() {
		return Result{}, errors.New("PR metadata has trailing content")
	}
	if doc.Schema != Schema {
		return Result{}, fmt.Errorf("PR metadata schema %q, want %q", doc.Schema, Schema)
	}
	if doc.FetchedAt != "" {
		if _, err := time.Parse(time.RFC3339, doc.FetchedAt); err != nil {
			return Result{}, fmt.Errorf("fetched_at: %w", err)
		}
	}
	sum := sha256.Sum256(raw)
	res := Result{Source: history.PRSource{
		Kind: kind, Origin: origin, Digest: hex.EncodeToString(sum[:]), FetchedAt: doc.FetchedAt,
	}}
	seen := map[int]bool{}
	for _, p := range doc.PullRequests {
		if p.Number <= 0 || seen[p.Number] {
			return Result{}, fmt.Errorf("PR number %d is missing or duplicated", p.Number)
		}
		seen[p.Number] = true
		if p.MergedAt == "" {
			res.Unmerged++
			continue
		}
		hp, err := convert(p)
		if err != nil {
			return Result{}, fmt.Errorf("PR #%d: %w", p.Number, err)
		}
		res.PRs = append(res.PRs, hp)
	}
	res.Source.Count = len(res.PRs)
	return res, nil
}

func convert(p PR) (history.PR, error) {
	created, err := time.Parse(time.RFC3339, p.CreatedAt)
	if err != nil {
		return history.PR{}, fmt.Errorf("created_at: %w", err)
	}
	merged, err := time.Parse(time.RFC3339, p.MergedAt)
	if err != nil {
		return history.PR{}, fmt.Errorf("merged_at: %w", err)
	}
	if p.TotalFiles < len(p.Files) || p.TotalCommits < len(p.Commits) {
		return history.PR{}, errors.New("total_files/total_commits smaller than the lists")
	}
	out := history.PR{
		Number: p.Number, AuthorLogin: p.AuthorLogin, AuthorType: p.AuthorType,
		Created: created, Merged: merged, TotalFiles: p.TotalFiles, TotalCommits: p.TotalCommits,
		Reviews: p.Reviews, Comments: p.Comments,
	}
	for _, f := range p.Files {
		out.Files = append(out.Files, history.FileChange{Path: f.Path, Added: f.Additions, Removed: f.Deletions})
	}
	for _, c := range p.Commits {
		t, err := time.Parse(time.RFC3339, c.AuthoredAt)
		if err != nil {
			return history.PR{}, fmt.Errorf("commit authored_at: %w", err)
		}
		out.CommitTimes = append(out.CommitTimes, t)
	}
	return out, nil
}

package prmeta

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

func directoryScope(dir string) history.Scope {
	return history.Scope{Kind: domain.SignalScopeDirectory, Anchor: dir, Contains: func(path string) bool {
		return path == dir || strings.HasPrefix(path, dir+"/")
	}}
}

func compute(prs []history.PR, src *history.PRSource) []domain.Signal {
	sc := directoryScope("a")
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
	res, wantDigest := loadFixtureMetadata(t)
	sigs := compute(res.PRs, &res.Source)

	effort := checkMeasuredPRSignal(t, sigs, "scout.pr_effort@1", "pr_effort", 3.0)
	batch := checkMeasuredPRSignal(t, sigs, "scout.batch_size@1", "batch_size", 6)
	checkFileProvenance(t, []domain.Signal{effort, batch}, wantDigest)
	disp := pick(t, sigs, "scout.dispersion@1")
	if disp.Detail["pr_p90_over_p50"] != 1.0 {
		t.Errorf("dispersion PR part = %v", disp.Detail)
	}
	if p, _ := disp.Detail["pr_provenance"].(map[string]any); p["digest"] != wantDigest {
		t.Errorf("dispersion provenance = %v", disp.Detail["pr_provenance"])
	}
}

func loadFixtureMetadata(t *testing.T) (Result, string) {
	t.Helper()
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
	return res, wantDigest
}

func checkMeasuredPRSignal(t *testing.T, sigs []domain.Signal, id, label string, want float64) domain.Signal {
	t.Helper()
	signal := pick(t, sigs, id)
	if signal.Missing != "" || signal.Value == nil || *signal.Value != want {
		t.Fatalf("%s = %+v", label, signal)
	}
	return signal
}

func checkFileProvenance(t *testing.T, signals []domain.Signal, wantDigest string) {
	t.Helper()
	for _, s := range signals {
		p, _ := s.Detail["provenance"].(map[string]any)
		if p["digest"] != wantDigest || p["kind"] != "file" || p["origin"] != "pr-metadata.json" {
			t.Errorf("%s provenance = %v", s.ID(), s.Detail["provenance"])
		}
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

// A fetch that succeeds on a repository with no merged PRs is metadata
// supplied, not missing: the reason must say so (dogfood CI, 2026-10-23).
func TestSuppliedButEmptyMetadataSaysSo(t *testing.T) {
	src := &history.PRSource{Kind: "github_api", Origin: "o/r", Digest: "d"}
	for _, id := range []string{"scout.pr_effort@1", "scout.batch_size@1"} {
		s := pick(t, compute(nil, src), id)
		if s.Missing != domain.SignalUnmeasured || s.Value != nil || s.MissingReason != "PR metadata supplied (github_api) holds no merged PRs" {
			t.Errorf("%s = %+v", id, s)
		}
	}
	if r := pick(t, compute(nil, nil), "scout.pr_effort@1").MissingReason; r != "no PR metadata supplied (local history lacks PR commits)" {
		t.Errorf("no source reason = %q", r)
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

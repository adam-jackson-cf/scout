package engine

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"scout/internal/domain"
	"scout/internal/inventory"
	"scout/internal/measures"
	"scout/internal/prmeta"
	"scout/internal/state"
)

// runHistory runs the history lane over a one-commit fixture repository with
// the given PR-metadata options and returns lane health, the store and the
// recorded pr_effort signal for the pkg/core scope.
func runHistory(t *testing.T, pr *PRMetadataOptions) (domain.LaneHealth, *state.Store, domain.Signal) {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "pkg", "core"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "pkg", "core", "a.py"), []byte("def f():\n    return 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, root, "init", "--quiet")
	gitRun(t, root, "add", ".")
	gitRun(t, root, "commit", "--quiet", "-m", "init")
	store, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	scan, err := inventory.Scan(ctx, root, inventory.ScanOptions{Scope: domain.Scope{Kind: domain.ScopeRepository, SelectedFamilies: []domain.Family{domain.FamilyQuality}}})
	if err != nil {
		t.Fatal(err)
	}
	options := Options{Scope: domain.Scope{Kind: domain.ScopeRepository}, Measures: &MeasuresOptions{
		History: true, PRMetadata: pr, Tools: measures.Tools{Lizard: "scout-no-lizard", Jscpd: "scout-no-jscpd"},
	}}
	var health domain.LaneHealth
	for _, h := range runMeasuresLanes(ctx, store, root, scan, nil, options) {
		if h.Lane == LaneHistory {
			health = h
		}
	}
	sigs, err := store.ListSignals(ctx, state.SignalFilter{MetricID: "scout.pr_effort@1", ScopeAnchor: "pkg/core"})
	if err != nil || len(sigs) == 0 {
		t.Fatalf("pr_effort signals = %v, err %v", sigs, err)
	}
	return health, store, sigs[0]
}

func writePRFile(t *testing.T, n int) string {
	t.Helper()
	now := time.Now().UTC().Add(-time.Hour)
	doc := prmeta.Document{Schema: prmeta.Schema, Repository: "o/r"}
	for i := 1; i <= n; i++ {
		at := func(d time.Duration) string { return now.Add(d).Format(time.RFC3339) }
		doc.PullRequests = append(doc.PullRequests, prmeta.PR{
			Number: i, AuthorLogin: "dev", CreatedAt: at(-5 * time.Hour), MergedAt: at(0),
			Files:      []prmeta.File{{Path: "pkg/core/a.py", Additions: 4, Deletions: 1}},
			TotalFiles: 1, TotalCommits: 2, Commits: []prmeta.Commit{{AuthoredAt: at(-2 * time.Hour)}, {AuthoredAt: at(-time.Hour)}},
		})
	}
	raw, _ := json.Marshal(doc)
	path := filepath.Join(t.TempDir(), "prs.json")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestHistoryLaneIngestsPRMetadataFileWithProvenance(t *testing.T) {
	path := writePRFile(t, 5)
	health, store, sig := runHistory(t, &PRMetadataOptions{File: path})
	if health.Status != domain.LaneRan {
		t.Fatalf("history lane = %+v", health)
	}
	if sig.Missing != "" || sig.Value == nil || *sig.Value != 1.5 {
		t.Fatalf("pr_effort = %+v", sig)
	}
	prov, _ := sig.Detail["provenance"].(map[string]any)
	if prov["kind"] != "file" || prov["digest"] == "" {
		t.Fatalf("provenance = %v", sig.Detail["provenance"])
	}
	srcs, err := store.PRMetadataSources(context.Background(), sig.Snapshot)
	if err != nil || len(srcs) != 1 || srcs[0].Digest != prov["digest"] || srcs[0].PRs != 5 {
		t.Fatalf("recorded sources = %+v, err %v", srcs, err)
	}
}

func TestHistoryLaneWithoutPRMetadataIsUnmeasuredAndUnreadableFileFailsTheLane(t *testing.T) {
	health, _, sig := runHistory(t, nil)
	if health.Status != domain.LaneRan || sig.Missing != domain.SignalUnmeasured || sig.MissingReason == "" {
		t.Fatalf("no metadata: lane %+v signal %+v", health, sig)
	}
	health, _, sig = runHistory(t, &PRMetadataOptions{File: filepath.Join(t.TempDir(), "missing.json")})
	if health.Status != domain.LaneFailed || health.Reason == "" || sig.Missing != domain.SignalUnmeasured {
		t.Fatalf("unreadable file: lane %+v signal %+v", health, sig)
	}
}

func TestHistoryLaneDoesNotFetchWhenNotEnabled(t *testing.T) {
	// Token present but Enabled false: no request is made and the signal is unmeasured.
	_, _, sig := runHistory(t, &PRMetadataOptions{Fetch: prmeta.FetchOptions{Repo: "o/r", Token: "x", Endpoint: "http://127.0.0.1:1"}})
	if sig.Missing != domain.SignalUnmeasured {
		t.Fatalf("pr_effort = %+v", sig)
	}
}

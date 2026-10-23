package baseline

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"scout/internal/state"
)

func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = dir
	out, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

// remoteWithClones returns a bare remote and n independent clones of it.
func remoteWithClones(t *testing.T, n int) (string, []string) {
	t.Helper()
	bare := t.TempDir()
	gitIn(t, bare, "init", "--bare", "--quiet")
	clones := make([]string, n)
	for i := range clones {
		clones[i] = t.TempDir()
		gitIn(t, clones[i], "init", "--quiet")
		gitIn(t, clones[i], "remote", "add", "origin", bare)
	}
	return bare, clones
}

// ledger builds the export files of a ledger holding one issue per id.
func ledger(t *testing.T, updatedAt string, ids ...string) Tree {
	t.Helper()
	files := Tree{state.ExportWouldBlock: nil, state.ExportCoverageSummary: []byte("{}\n")}
	var issues, dispositions []byte
	for _, id := range ids {
		issue, _ := state.CanonicalLine(map[string]any{"issueId": id, "title": "t"})
		disposition, _ := state.CanonicalLine(map[string]any{"issueId": id, "state": "open", "updatedAt": updatedAt})
		issues, dispositions = append(issues, issue...), append(dispositions, disposition...)
	}
	files[state.ExportIssues], files[state.ExportDispositions] = issues, dispositions
	files[state.ExportManifest] = state.RenderManifest(state.ManifestMeta{RepositoryID: "repo", WouldBlock: state.WouldBlockPresent}, files)
	return files
}

func issueIDs(t *testing.T, tree Tree) []string {
	t.Helper()
	var ids []string
	for _, line := range state.SplitLines(tree[state.ExportIssues]) {
		var row struct {
			IssueID string `json:"issueId"`
		}
		if err := json.Unmarshal(line, &row); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, row.IssueID)
	}
	return ids
}

func TestConcurrentPublishersBothLandWithLinearHistory(t *testing.T) {
	ctx := context.Background()
	bare, clones := remoteWithClones(t, 2)
	first := Options{Repo: clones[0], Remote: bare, Trusted: true}
	second := Options{Repo: clones[1], Remote: bare, Trusted: true}

	aStarted, bLanded := make(chan struct{}), make(chan struct{})
	var once sync.Once
	var aResult Result
	var aErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		local := ledger(t, "2026-10-01T00:00:00Z", "issue:a")
		aResult, aErr = Publish(ctx, first, func(current Tree) (Tree, error) {
			once.Do(func() {
				close(aStarted)
				<-bLanded // B wins the race after A read the (empty) tip
			})
			return MergeState(current, local)
		})
	}()
	<-aStarted
	bResult, err := Publish(ctx, second, func(current Tree) (Tree, error) {
		return MergeState(current, ledger(t, "2026-10-01T00:00:00Z", "issue:b"))
	})
	if err != nil {
		t.Fatal(err)
	}
	close(bLanded)
	wg.Wait()
	if aErr != nil {
		t.Fatal(aErr)
	}
	if bResult.Attempts != 1 || aResult.Attempts != 2 {
		t.Fatalf("attempts a=%d b=%d, want a=2 (CAS retry) b=1", aResult.Attempts, bResult.Attempts)
	}
	if aResult.Parent != bResult.Commit {
		t.Fatalf("retry built on %s, want B's commit %s", aResult.Parent, bResult.Commit)
	}

	tree, tip, err := Fetch(ctx, first)
	if err != nil {
		t.Fatal(err)
	}
	if tip != aResult.Commit {
		t.Fatalf("tip %s, want %s", tip, aResult.Commit)
	}
	if got := strings.Join(issueIDs(t, tree), ","); got != "issue:a,issue:b" {
		t.Fatalf("final tip issues = %q, want both publishers", got)
	}
	if got := gitIn(t, bare, "rev-list", "--count", DefaultBranch); got != "2" {
		t.Fatalf("history has %s commits, want 2", got)
	}
	if merges := gitIn(t, bare, "rev-list", "--merges", DefaultBranch); merges != "" {
		t.Fatalf("history has merge commits: %s", merges)
	}
	root := gitIn(t, bare, "rev-list", "--max-parents=0", DefaultBranch)
	if root != bResult.Commit {
		t.Fatalf("root commit %s, want B's %s", root, bResult.Commit)
	}
}

func TestManyConcurrentPublishersAllLand(t *testing.T) {
	ctx := context.Background()
	const n = 4
	bare, clones := remoteWithClones(t, n)
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := "issue:" + string(rune('a'+i))
			_, errs[i] = Publish(ctx, Options{Repo: clones[i], Remote: bare, Trusted: true, MaxAttempts: 32}, func(current Tree) (Tree, error) {
				return MergeState(current, ledger(t, "2026-10-01T00:00:00Z", id))
			})
		}()
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("publisher %d: %v", i, err)
		}
	}
	tree, _, err := Fetch(ctx, Options{Repo: clones[0], Remote: bare})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(issueIDs(t, tree), ","); got != "issue:a,issue:b,issue:c,issue:d" {
		t.Fatalf("final issues = %q", got)
	}
	if got := gitIn(t, bare, "rev-list", "--count", DefaultBranch); got != "4" {
		t.Fatalf("history has %s commits, want 4", got)
	}
	if merges := gitIn(t, bare, "rev-list", "--merges", DefaultBranch); merges != "" {
		t.Fatalf("history has merge commits: %s", merges)
	}
}

func TestMergeKeepsNewerDispositionPerIssue(t *testing.T) {
	older := ledger(t, "2026-10-01T00:00:00Z", "issue:a", "issue:b")
	newer := ledger(t, "2026-10-02T00:00:00Z", "issue:b")
	merged, err := MergeState(newer, older) // tip is newer for b; local has a and an older b
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(issueIDs(t, merged), ","); got != "issue:a,issue:b" {
		t.Fatalf("merged issues = %q", got)
	}
	if !strings.Contains(string(merged[state.ExportDispositions]), "2026-10-02T00:00:00Z") {
		t.Fatalf("newer tip disposition lost: %s", merged[state.ExportDispositions])
	}
}

func TestPublishIsReadOnlyWithoutTrust(t *testing.T) {
	bare, clones := remoteWithClones(t, 1)
	_, err := Publish(context.Background(), Options{Repo: clones[0], Remote: bare}, func(Tree) (Tree, error) {
		t.Fatal("apply must not run for an untrusted publisher")
		return nil, nil
	})
	if !errors.Is(err, ErrUntrusted) {
		t.Fatalf("err = %v, want ErrUntrusted", err)
	}
	if refs := gitIn(t, bare, "for-each-ref"); refs != "" {
		t.Fatalf("untrusted publish wrote refs: %s", refs)
	}
}

func TestPublishNeverCommitsDatabaseFiles(t *testing.T) {
	bare, clones := remoteWithClones(t, 1)
	for _, path := range []string{"state/v1/state.sqlite", "state/v1/state.sqlite-wal", ".scout/state.sqlite", "state/v1/../x"} {
		_, err := Publish(context.Background(), Options{Repo: clones[0], Remote: bare, Trusted: true}, func(Tree) (Tree, error) {
			return Tree{path: []byte("x")}, nil
		})
		if err == nil {
			t.Fatalf("publish accepted %s", path)
		}
	}
	if refs := gitIn(t, bare, "for-each-ref"); refs != "" {
		t.Fatalf("refused publish still wrote refs: %s", refs)
	}
}

func TestPublishUnchangedTreeAddsNoCommit(t *testing.T) {
	ctx := context.Background()
	bare, clones := remoteWithClones(t, 1)
	opts := Options{Repo: clones[0], Remote: bare, Trusted: true, Now: func() time.Time { return time.Unix(1790000000, 0) }}
	local := ledger(t, "2026-10-01T00:00:00Z", "issue:a")
	apply := func(current Tree) (Tree, error) { return MergeState(current, local) }
	first, err := Publish(ctx, opts, apply)
	if err != nil || !first.Changed {
		t.Fatalf("first publish = %+v, %v", first, err)
	}
	second, err := Publish(ctx, opts, apply)
	if err != nil || second.Changed || second.Commit != first.Commit {
		t.Fatalf("second publish = %+v, %v; want unchanged at %s", second, err, first.Commit)
	}
}

func TestFetchReportsMissingBranch(t *testing.T) {
	bare, clones := remoteWithClones(t, 1)
	if _, _, err := Fetch(context.Background(), Options{Repo: clones[0], Remote: bare}); !errors.Is(err, ErrNoState) {
		t.Fatalf("err = %v, want ErrNoState", err)
	}
	if _, ok, err := LoadBase(context.Background(), Options{Repo: clones[0], Remote: bare}); ok || err != nil {
		t.Fatalf("LoadBase on missing branch = %v, %v", ok, err)
	}
}

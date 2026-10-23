package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"scout/internal/state"
)

func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = dir
	out, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func repoWithRemote(t *testing.T) (repo, bare string) {
	t.Helper()
	bare = t.TempDir()
	gitRun(t, bare, "init", "--bare", "--quiet")
	repo = t.TempDir()
	gitRun(t, repo, "init", "--quiet")
	gitRun(t, repo, "remote", "add", "origin", bare)
	store, err := state.Open(repo)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	return repo, bare
}

func TestStatePublishWithoutTrustedRefusesWithExit3(t *testing.T) {
	repo, bare := repoWithRemote(t)
	code := run(context.Background(), []string{"state", "publish", "--root", repo, "--remote", bare})
	if code != 3 {
		t.Fatalf("publish without --trusted exit = %d, want 3", code)
	}
	if refs := gitRun(t, bare, "for-each-ref"); refs != "" {
		t.Fatalf("untrusted publish wrote refs: %s", refs)
	}
}

func TestStatePublishRestoreRoundTripNeverCommitsSQLite(t *testing.T) {
	repo, bare := repoWithRemote(t)
	ctx := context.Background()
	if code := run(ctx, []string{"state", "publish", "--root", repo, "--remote", bare, "--trusted"}); code != 0 {
		t.Fatalf("trusted publish exit = %d, want 0", code)
	}
	files := gitRun(t, bare, "ls-tree", "-r", "--name-only", "scout/state")
	for _, want := range []string{"state/v1/issues.jsonl", "state/v1/dispositions.jsonl", "state/v1/manifest.json"} {
		if !strings.Contains(files, want) {
			t.Fatalf("scout/state lacks %s:\n%s", want, files)
		}
	}
	if strings.Contains(files, "sqlite") || strings.Contains(files, "-wal") {
		t.Fatalf("scout/state contains database files:\n%s", files)
	}

	other := t.TempDir()
	gitRun(t, other, "init", "--quiet")
	if code := run(ctx, []string{"state", "restore", "--root", other, "--remote", bare}); code != 0 {
		t.Fatalf("restore exit = %d, want 0", code)
	}
	// A remote without scout/state is unavailable, not an empty baseline.
	emptyBare := t.TempDir()
	gitRun(t, emptyBare, "init", "--bare", "--quiet")
	if code := run(ctx, []string{"state", "restore", "--root", t.TempDir(), "--remote", emptyBare}); code != 3 {
		t.Fatalf("restore without a branch exit = %d, want 3", code)
	}
}

func TestExportThenImportCommandsReproduceFiles(t *testing.T) {
	repo, _ := repoWithRemote(t)
	ctx := context.Background()
	out := t.TempDir()
	if code := run(ctx, []string{"export", "--root", repo, "--out", out}); code != 0 {
		t.Fatalf("export exit = %d", code)
	}
	target := t.TempDir()
	if code := run(ctx, []string{"import", "--root", target, "--from", out}); code != 0 {
		t.Fatalf("import exit = %d", code)
	}
	again := t.TempDir()
	if code := run(ctx, []string{"export", "--root", target, "--out", again}); code != 0 {
		t.Fatalf("re-export exit = %d", code)
	}
	for _, name := range []string{"issues.jsonl", "dispositions.jsonl", "would_block.jsonl", "coverage-summary.json", "manifest.json"} {
		first, err := os.ReadFile(filepath.Join(out, "state", "v1", name))
		if err != nil {
			t.Fatal(err)
		}
		second, err := os.ReadFile(filepath.Join(again, "state", "v1", name))
		if err != nil {
			t.Fatal(err)
		}
		if string(first) != string(second) {
			t.Fatalf("%s differs after import", name)
		}
	}
}

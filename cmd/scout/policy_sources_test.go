package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"scout/internal/policy"
)

func gitRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "t@example.test"}, {"config", "user.name", "t"}} {
		if out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	return root
}

func commitAll(t *testing.T, root string) {
	t.Helper()
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-q", "--allow-empty", "-m", "c"}} {
		if out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
}

// The policy comes from HEAD alone: a worktree edit or deletion must not
// change it, and a HEAD that cannot be read is an error, not the default.
func TestHeadPolicyReadsCommittedPolicyNotWorktree(t *testing.T) {
	root := gitRepo(t)
	path := filepath.Join(root, filepath.FromSlash(policy.Path))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("profile: guardian\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	commitAll(t, root)
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	pol, err := headPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	if pol.Profile() != policy.ProfileGuardian {
		t.Fatalf("profile = %q; a worktree deletion changed the committed policy", pol.Profile())
	}
}

func TestHeadPolicyWithoutCommittedPolicyIsDefaultAndUncommittedOneIsIgnored(t *testing.T) {
	root := gitRepo(t)
	commitAll(t, root)
	path := filepath.Join(root, filepath.FromSlash(policy.Path))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("profile: guardian\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	pol, err := headPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	if pol.Profile() != policy.ProfileReport {
		t.Fatalf("profile = %q, want the default for a repository with no committed policy", pol.Profile())
	}
}

func TestHeadPolicyUnreadableHeadIsAnError(t *testing.T) {
	if _, err := headPolicy(t.TempDir()); err == nil {
		t.Fatal("a directory that is not a git repository yielded a policy")
	}
}

func TestSanitizedNeverReturnsEmptyForRejectedText(t *testing.T) {
	if got := sanitized("bad\x00text"); got == "" {
		t.Fatal("rejected text produced an empty message")
	}
	if got := sanitized("fine"); got != "fine" {
		t.Fatalf("sanitized(fine) = %q", got)
	}
}

func TestAssessRejectsBothPRMetadataSources(t *testing.T) {
	code := assess(t.Context(), []string{"--root", t.TempDir(), "--pr-metadata", "x.json", "--fetch-pr-metadata", "--pr-repo", "o/r"})
	if code != exitUnavailable {
		t.Fatalf("exit = %d, want %d", code, exitUnavailable)
	}
}

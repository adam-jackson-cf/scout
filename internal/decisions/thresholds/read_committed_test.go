//go:build boundary

package thresholds

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.CommandContext(boundedTestContext(t), "git", append([]string{"-c", "user.name=t", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func writeRepoFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// The override is read from the committed ref, so a working-tree edit (a pull
// request's own change) cannot move the thresholds it is ranked under.
func TestReadCommittedReadsTheRefNotTheWorkingTree(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init", "-q")
	writeRepoFile(t, filepath.Join(root, "a.txt"), "x")
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-q", "-m", "one")
	if got, err := ReadCommitted(root, "HEAD"); err != nil || got != nil {
		t.Fatalf("no committed file must read as nil, got %q %v", got, err)
	}
	writeRepoFile(t, filepath.Join(root, filepath.FromSlash(Path)), `{"schema":1,"decisions":[]}`)
	if got, err := ReadCommitted(root, "HEAD"); err != nil || got != nil {
		t.Fatalf("an uncommitted working-tree file must not be read, got %q %v", got, err)
	}
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-q", "-m", "two")
	if got, err := ReadCommitted(root, "HEAD"); err != nil || string(got) != `{"schema":1,"decisions":[]}` {
		t.Fatalf("committed file %q %v", got, err)
	}
	if _, err := ReadCommitted(root, "no-such-ref"); err == nil {
		t.Fatal("an unresolvable ref must be an error")
	}
}

//go:build boundary

package policy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func run(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.CommandContext(boundedTestContext(t), "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func writePolicyFile(t *testing.T, dir, content string) {
	t.Helper()
	p := filepath.Join(dir, Path)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func commitBaseWithoutPolicy(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "add", ".")
	run(t, dir, "commit", "-qm", "base")
	run(t, dir, "tag", "base-nopolicy")
}

func commitBaseWithReportPolicy(t *testing.T, dir string) {
	t.Helper()
	writePolicyFile(t, dir, "profile: report\n")
	run(t, dir, "add", ".")
	run(t, dir, "commit", "-qm", "policy")
	run(t, dir, "branch", "base")
}

func commitPRHeadActivatingCheckInWorkingTree(t *testing.T, dir string) {
	t.Helper()
	run(t, dir, "checkout", "-q", "-b", "pr")
	writePolicyFile(t, dir, guardianPolicy)
	run(t, dir, "add", ".")
	run(t, dir, "commit", "-qm", "pr activates")
}

func TestLoadReadsBaseRefOnly(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "init", "-q", "-b", "main")
	commitBaseWithoutPolicy(t, dir)
	commitBaseWithReportPolicy(t, dir)
	commitPRHeadActivatingCheckInWorkingTree(t, dir)

	assertBaseLoadIgnoresPRHead(t, dir)
	assertAbsentAndUnresolvableRefs(t, dir)
	assertInvalidPolicyAtRefSurfacesError(t, dir)
}

func assertBaseLoadIgnoresPRHead(t *testing.T, dir string) {
	t.Helper()
	p, err := Load(dir, "base")
	if err != nil {
		t.Fatal(err)
	}
	if p.Profile() != "report" || p.CheckStatus("guardrails.SUPPRESS.added.unscoped", "1") != Shadow {
		t.Fatal("PR-head policy leaked into base load")
	}
	head, err := Load(dir, "pr")
	if err != nil || head.CheckStatus("guardrails.SUPPRESS.added.unscoped", "1") != Active {
		t.Fatalf("control: head policy should activate: %v", err)
	}
}

func assertAbsentAndUnresolvableRefs(t *testing.T, dir string) {
	t.Helper()
	abs, err := Load(dir, "base-nopolicy")
	if err != nil || abs.Profile() != "report" || len(abs.activations) != 0 {
		t.Fatalf("absent file: %v %+v", err, abs)
	}
	if _, err := Load(dir, "no-such-ref"); err == nil {
		t.Fatal("unresolvable ref must error")
	}
}

func assertInvalidPolicyAtRefSurfacesError(t *testing.T, dir string) {
	t.Helper()
	run(t, dir, "checkout", "-q", "-b", "bad", "main")
	writePolicyFile(t, dir, "profile: guardian\nblocking:\n  - id: guardrails.CONFIG.*\n")
	run(t, dir, "add", ".")
	run(t, dir, "commit", "-qm", "bad")
	if _, err := Load(dir, "bad"); err == nil || !strings.Contains(err.Error(), "wildcard") {
		t.Fatalf("expected wildcard error, got %v", err)
	}
}

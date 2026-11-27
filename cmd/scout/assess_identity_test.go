//go:build boundary

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"scout/internal/domain"
	"scout/internal/engine"
	"scout/internal/state"
)

func pinnedClawpatch(t *testing.T) string {
	t.Helper()
	if path := os.Getenv("SCOUT_CLAWPATCH_BINARY"); path != "" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("no home directory: %v", err)
	}
	path := filepath.Join(home, ".cache", "scout", "clawpatch", "734f81bc1514b52a0627edb0c24ebe1a18b36105", "dist", "cli.js")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("pinned Clawpatch bundle is not installed (scripts/bootstrap-tools.sh): %v", err)
	}
	return path
}

func gitIn(t *testing.T, root string, args ...string) {
	t.Helper()
	command := exec.CommandContext(boundedTestContext(t), "git", append([]string{"-C", root, "-c", "user.email=t@t", "-c", "user.name=t"}, args...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}

func complexFn(name string, branches int) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "def %s(x):\n", name)
	for i := 1; i <= branches; i++ {
		fmt.Fprintf(&builder, "    if x == %d:\n        return %d\n", i, i)
	}
	return builder.String() + "    return 0\n"
}

const perfFn = "def build(items):\n    out = []\n    for i in items:\n        if i:\n            out.append(i * 2)\n    return out\n"

func writeIdentityFixture(t *testing.T, root string) {
	t.Helper()
	for name, content := range map[string]string{"pkg/a.py": complexFn("complex_a", 11), "pkg/b.py": perfFn} {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitIn(t, root, "init", "-q")
	gitIn(t, root, "add", "-A")
	gitIn(t, root, "commit", "-qm", "r1")
}

func assessIdentityOnce(t *testing.T, root, binary string) domain.Run {
	t.Helper()
	scope, err := scopeOptions("repository", nil, "", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := engine.Assess(context.Background(), engine.Config{Root: root, ClawpatchBinary: binary, Provider: providerConfig(false)}, engine.Options{Scope: scope})
	if err != nil {
		t.Fatal(err)
	}
	return result.Run
}

func issuesByAnchor(t *testing.T, root string) map[string]domain.Issue {
	t.Helper()
	store, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOrFail(t, store)
	all, err := store.ListIssues(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	byKey := map[string]domain.Issue{}
	for _, issue := range all {
		byKey[issue.Anchor.Path+"|"+issue.Anchor.RuleKey] = issue
	}
	return byKey
}

func findAnchoredIssue(t *testing.T, issues map[string]domain.Issue, path, rulePart string) domain.Issue {
	t.Helper()
	for key, issue := range issues {
		if strings.HasPrefix(key, path+"|") && strings.Contains(key, rulePart) {
			return issue
		}
	}
	t.Fatalf("fixture produced no %s finding at %s (ruff rules differ)", rulePart, path)
	return domain.Issue{}
}

func acceptIdentityIssue(t *testing.T, root, issueID string) {
	t.Helper()
	store, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOrFail(t, store)
	if _, err := store.TriageIssue(context.Background(), state.Triage{IssueID: issueID, State: domain.IssueAccepted, Owner: "owner", DispositionReason: "known", Actor: "human", Reason: "reviewed"}); err != nil {
		t.Fatal(err)
	}
}

func renameAndRaiseComplexity(t *testing.T, root string) {
	t.Helper()
	pad := strings.Repeat("# pad\n", 20)
	if err := os.WriteFile(filepath.Join(root, "pkg", "a.py"), []byte(pad+complexFn("complex_a", 12)), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, root, "mv", "pkg/b.py", "pkg/b_renamed.py")
	gitIn(t, root, "add", "-A")
	gitIn(t, root, "commit", "-qm", "r2")
}

func assertOldPathResolved(t *testing.T, issues map[string]domain.Issue, path string) {
	t.Helper()
	for key, issue := range issues {
		if strings.HasPrefix(key, path+"|") && issue.State != domain.IssueResolved {
			t.Fatalf("old path left %s issue %s", issue.State, issue.IssueID)
		}
	}
}

func hasEvidenceDrift(issue domain.Issue) bool {
	for _, entry := range issue.History {
		if entry.Event == domain.EventEvidenceDrift {
			return true
		}
	}
	return false
}

func requirePinnedRuff(t *testing.T) {
	t.Helper()
	out, err := exec.CommandContext(boundedTestContext(t), "ruff", "--version").Output()
	if err != nil || !strings.Contains(string(out), "0.15.8") {
		t.Fatalf("ruff 0.15.8 is not installed (source .tools/activate): %v %s", err, out)
	}
}

// Real assess path on a git fixture: the first run is partial (opt-in lanes are
// blocked), yet a later rename keeps the issue ID and a numeric C901 change
// records evidence_drift with the accepted disposition intact.
func TestAssessKeepsIdentityAcrossRenameAfterPartialRun(t *testing.T) {
	requirePinnedRuff(t)
	binary := pinnedClawpatch(t)
	root := t.TempDir()
	writeIdentityFixture(t, root)

	if run := assessIdentityOnce(t, root, binary); run.State != domain.RunPartial {
		t.Fatalf("run 1 state = %q, want partial", run.State)
	}
	first := issuesByAnchor(t, root)
	a, b := findAnchoredIssue(t, first, "pkg/a.py", "C901"), findAnchoredIssue(t, first, "pkg/b.py", "PERF")
	acceptIdentityIssue(t, root, a.IssueID)

	renameAndRaiseComplexity(t, root)
	assessIdentityOnce(t, root, binary)

	second := issuesByAnchor(t, root)
	renamed := findAnchoredIssue(t, second, "pkg/b_renamed.py", "PERF")
	if renamed.IssueID != b.IssueID {
		t.Fatalf("rename minted a new ID: %s -> %s", b.IssueID, renamed.IssueID)
	}
	assertOldPathResolved(t, second, "pkg/b.py")
	drifted := findAnchoredIssue(t, second, "pkg/a.py", "C901")
	hasDrift := hasEvidenceDrift(drifted)
	if drifted.IssueID != a.IssueID || drifted.State != domain.IssueAccepted || !hasDrift {
		t.Fatalf("complexity change: id %s->%s state %s drift=%v", a.IssueID, drifted.IssueID, drifted.State, hasDrift)
	}
}

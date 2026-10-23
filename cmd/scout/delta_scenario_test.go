package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func pyBranchy(name string, branches int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "def %s(x):\n    total = 0\n", name)
	for i := range branches {
		fmt.Fprintf(&b, "    if x == %d:\n        total += %d\n", i, i+1)
	}
	b.WriteString("    return total\n")
	return b.String()
}

// Design §11 Phase 3: a PR raising a function CC 9 → 14 yields one descriptive,
// report-only delta issue and leaves the guardian-profile exit code unchanged.
func TestDiffDeltaIsReportOnlyUnderGuardianProfile(t *testing.T) {
	for _, tool := range []string{"ruff", "gitleaks", "semgrep", "lizard", "jscpd"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not installed (source .tools/activate)", tool)
		}
	}
	t.Setenv("SCOUT_CLAWPATCH_BINARY", pinnedClawpatch(t))
	root := t.TempDir()
	gitIn(t, root, "init", "-q", "-b", "main")
	commitFiles(t, root, "base", map[string]string{
		"pkg/wide.py":        pyBranchy("wide", 8),
		".scout/policy.yaml": "profile: guardian\nmaintainers: [alice]\n",
	})
	base := gitRev(t, root, "HEAD")
	gitIn(t, root, "checkout", "-q", "-b", "pr")

	// Unchanged control: touching only a docs file yields no delta.
	commitFiles(t, root, "docs only", map[string]string{"README.md": "hello\n"})
	code, report := runDiffJSON(t, root, "--base", base)
	if code != 0 || len(deltaIssues(t, report)) != 0 {
		t.Fatalf("docs-only PR: exit %d issues %v", code, report["issues"])
	}

	commitFiles(t, root, "raise CC", map[string]string{"pkg/wide.py": pyBranchy("wide", 13)})
	code, report = runDiffJSON(t, root, "--base", base)
	if code != 0 {
		t.Fatalf("delta changed the exit code: %d", code)
	}
	issues := deltaIssues(t, report)
	if len(issues) != 1 {
		t.Fatalf("want one delta issue, got %v", report["issues"])
	}
	issue := issues[0].(map[string]any)
	if issue["claim"] != "descriptive" || issue["checkId"] != "delta.cc-threshold@1" || issue["attribution"] != "introduced" || issue["path"] != "pkg/wide.py" ||
		issue["blocking"] != false || issue["wouldBlock"] != false || issue["checkStatus"] != "not_eligible" {
		t.Fatalf("delta issue = %v", issue)
	}
	if blocking, _ := report["would_block"].([]any); len(blocking) != 0 {
		t.Fatalf("delta reached would_block: %v", blocking)
	}

	// The ledger holds one claim-bearing issue, and a re-run on the same PR matches it.
	ledger := ledgerDeltaIssues(t, root)
	if len(ledger) != 1 || ledger[0]["claim"] != "descriptive" || ledger[0]["rule"] != "delta.cc-threshold@1" {
		t.Fatalf("ledger delta issues = %v", ledger)
	}
	code, report = runDiffJSON(t, root, "--base", base)
	again := deltaIssues(t, report)
	if code != 0 || len(again) != 1 || again[0].(map[string]any)["issueId"] != issue["issueId"] {
		t.Fatalf("re-run: exit %d issues %v, want the same issue %v", code, again, issue["issueId"])
	}
	if after := ledgerDeltaIssues(t, root); len(after) != 1 || after[0]["issueId"] != ledger[0]["issueId"] {
		t.Fatalf("re-run changed the ledger: %v", after)
	}
}

// deltaIssues returns the claim-bearing issues of a diff report.
func deltaIssues(t *testing.T, report map[string]any) []any {
	t.Helper()
	all, ok := report["issues"].([]any)
	if !ok {
		t.Fatalf("report has no issues array: %v", report)
	}
	var out []any
	for _, item := range all {
		if item.(map[string]any)["claim"] != nil {
			out = append(out, item)
		}
	}
	return out
}

// ledgerDeltaIssues reads `scout issues --json` and keeps the claim-bearing issues.
func ledgerDeltaIssues(t *testing.T, root string) []map[string]any {
	t.Helper()
	out, code := captureStdout(t, func() int {
		return run(t.Context(), []string{"issues", "--root", root, "--json"})
	})
	if code != 0 {
		t.Fatalf("scout issues exit %d: %s", code, out)
	}
	var envelope struct {
		Result []map[string]any `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("decode issues JSON: %v\n%s", err, out)
	}
	var claimed []map[string]any
	for _, issue := range envelope.Result {
		if issue["claim"] != nil {
			claimed = append(claimed, issue)
		}
	}
	return claimed
}

func TestReadinessMatrixPrintsSection53Rows(t *testing.T) {
	root := t.TempDir()
	gitIn(t, root, "init", "-q", "-b", "main")
	commitFiles(t, root, "base", map[string]string{"main.go": "package main\n"})
	t.Setenv("PATH", t.TempDir()+":"+"/usr/bin:/bin")
	t.Setenv("SCOUT_MODEL_CLI", "")
	out, code := captureStdout(t, func() int {
		return run(t.Context(), []string{"readiness", "--matrix", "--root", root})
	})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	for _, want := range []string{"| Capability | Go * |", "CC/NLOC/duplication; history (git); secrets/CI security", "Import graph / cycles; guardrails; rubric", "CRAP (needs coverage artifact)", "Semantic review (contextual) | skipped (no model CLI)", "present in this repository: Go"} {
		if !strings.Contains(out, want) {
			t.Fatalf("matrix output lacks %q:\n%s", want, out)
		}
	}
}

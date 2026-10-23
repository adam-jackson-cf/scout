package main

import (
	"os/exec"
	"testing"

	"scout/internal/policy"
)

// A PR that erodes guardrails in all three languages is reported by
// `scout diff` with allow-list-eligible check ids, attributed as introduced;
// the scoped/justified counterparts in the same PR are not reported as
// blocking-eligible findings.
func TestDiffReportsGuardrailErosion(t *testing.T) {
	for _, tool := range []string{"ruff", "gitleaks", "semgrep"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not installed (source .tools/activate)", tool)
		}
	}
	t.Setenv("SCOUT_CLAWPATCH_BINARY", pinnedClawpatch(t))

	root := t.TempDir()
	gitIn(t, root, "init", "-q", "-b", "main")
	commitFiles(t, root, "base", map[string]string{
		".golangci.yml":      "linters:\n  disable:\n    - lll\n",
		"tsconfig.json":      "{\"extends\": \"./base.tsconfig.json\"}\n",
		"base.tsconfig.json": "{\"compilerOptions\": {\"strict\": true}}\n",
		"web/a.test.ts":      "it('works', () => {});\n",
		"pkg/a.py":           "x = 1\n",
		"main.go":            "package main\n\nfunc main() {}\n",
	})
	base := gitRev(t, root, "HEAD")
	gitIn(t, root, "checkout", "-q", "-b", "pr")
	commitFiles(t, root, "erode", map[string]string{
		".golangci.yml":      "linters:\n  disable:\n    - lll\n    - errcheck\n",
		"base.tsconfig.json": "{\"compilerOptions\": {\"strict\": false}}\n",
		"web/a.test.ts":      "it.only('works', () => {});\n",
		"pkg/a.py":           "x = 1\ny = 2  # noqa\nz = 3  # noqa: E501 long url\n",
		"main.go":            "package main\n\nfunc main() {} //nolint:errcheck // best effort\n",
	})

	code, report := runDiff(t, root, "--base", base)
	if code != 0 {
		t.Fatalf("report profile must not gate: exit %d", code)
	}
	got := map[string]int{}
	for _, issue := range report.Issues {
		if issue.Attribution != "introduced" {
			continue
		}
		got[issue.CheckID]++
	}
	for _, id := range []string{
		"guardrails.CONFIG.ignores-added.golangci",
		"guardrails.CONFIG.strictness-lowered.tsconfig",
		"guardrails.SUPPRESS.added.unscoped",
		"guardrails.TESTQ.focused-test-added",
	} {
		if got[id] == 0 {
			t.Errorf("%s not reported as introduced; got %v", id, got)
		}
		if err := policy.ValidateEligibleID(id); err != nil {
			t.Errorf("%s not allow-list eligible: %v", id, err)
		}
	}
	if got["guardrails.SUPPRESS.added.unscoped"] != 1 {
		t.Errorf("only the bare # noqa is unscoped; got %d", got["guardrails.SUPPRESS.added.unscoped"])
	}
}

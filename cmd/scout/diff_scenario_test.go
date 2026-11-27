//go:build boundary

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"scout/internal/domain"
	"scout/internal/guardrails"
	"scout/internal/state"
)

type diffReport struct {
	Gate struct {
		Status    string   `json:"status"`
		Profile   string   `json:"profile"`
		Blocking  []string `json:"blocking"`
		Overrides []struct {
			Accepted bool   `json:"accepted"`
			Reason   string `json:"reason"`
		} `json:"overrides"`
	} `json:"gate"`
	Comparison string `json:"comparison"`
	Health     []struct {
		Lane     string `json:"lane"`
		Required bool   `json:"required"`
		Status   string `json:"status"`
	} `json:"health"`
	WouldBlock []struct {
		CheckID string `json:"checkId"`
		IssueID string `json:"issueId"`
	} `json:"would_block"`
	Ranking struct {
		Method string `json:"method"`
	} `json:"ranking"`
	Issues []struct {
		IssueID     string `json:"issueId"`
		CheckID     string `json:"checkId"`
		Attribution string `json:"attribution"`
		Blocking    bool   `json:"blocking"`
	} `json:"issues"`
	PreExisting int `json:"preExisting"`
}

type diffOutcome struct {
	ExitCode int        `json:"exitCode"`
	Result   diffReport `json:"result"`
}

// runDiff runs `scout diff --json` in-process and decodes the outcome.
func runDiff(t *testing.T, root string, extra ...string) (int, diffReport) {
	t.Helper()
	text, code := captureStdout(t, func() int {
		return run(context.Background(), append([]string{"diff", "--root", root, "--json"}, extra...))
	})
	var outcome diffOutcome
	data := []byte(text)
	if err := json.Unmarshal(data, &outcome); err != nil {
		t.Fatalf("decode diff JSON: %v\n%s", err, data)
	}
	if outcome.ExitCode != code {
		t.Fatalf("JSON exitCode %d != process exit %d", outcome.ExitCode, code)
	}
	return code, outcome.Result
}

// runDiffJSON is runDiff with an untyped report for fields the typed one omits.
func runDiffJSON(t *testing.T, root string, extra ...string) (int, map[string]any) {
	t.Helper()
	text, code := captureStdout(t, func() int {
		return run(context.Background(), append([]string{"diff", "--root", root, "--json"}, extra...))
	})
	var outcome struct {
		Result map[string]any `json:"result"`
	}
	if err := json.Unmarshal([]byte(text), &outcome); err != nil {
		t.Fatal(err)
	}
	return code, outcome.Result
}

func writeRepoFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func commitFiles(t *testing.T, root, message string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		writeRepoFile(t, root, name, content)
		gitIn(t, root, "add", "--", name)
	}
	gitIn(t, root, "commit", "-qm", message)
}

func fakeGitHubPAT(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	return "ghp" + "_" + strings.ToUpper(hex.EncodeToString(sum[:])[:18]) + hex.EncodeToString(sum[:])[20:38]
}

func policyYAML(version string) string {
	return "profile: guardian\nmaintainers: [alice]\nblocking:\n  - id: " + guardrails.UnscopedCheckID + "\n    version: \"" + version + "\"\n    activated_at: 2026-01-01\n    evidence: {would_block: 3, warranted: 3, window_days: 15, pr_runs: 6}\n"
}

func wouldBlockChecks(t *testing.T, root string) map[string]int {
	t.Helper()
	store, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOrFail(t, store)
	rows, err := store.ListLegacyWouldBlock(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, row := range rows {
		counts[row.CheckID]++
	}
	return counts
}

type diffGateFixture struct {
	root, c1, c2, c3, legacyToken string
	prA                           map[string]string
}

func newDiffGateFixture(t *testing.T) diffGateFixture {
	t.Helper()
	root := t.TempDir()
	legacyToken := fakeGitHubPAT("legacy-debt")
	prToken := fakeGitHubPAT("pr-a-token")
	gitIn(t, root, "init", "-q", "-b", "main")
	commitFiles(t, root, "c1", map[string]string{"pkg/legacy.py": "OLD = \"" + legacyToken + "\"\n\n\ndef old(x):\n    return x\n", "README.md": "demo\n"})
	c1 := gitRev(t, root, "HEAD")
	commitFiles(t, root, "c2: activate guardrails check", map[string]string{".scout/policy.yaml": policyYAML(guardrails.UnscopedVersion)})
	c2 := gitRev(t, root, "HEAD")
	gitIn(t, root, "checkout", "-q", "-b", "bumped")
	commitFiles(t, root, "c3: rule version bump", map[string]string{".scout/policy.yaml": policyYAML(guardrails.UnscopedCheckID + "@0")})
	c3 := gitRev(t, root, "HEAD")

	prA := map[string]string{"pkg/new.py": "GITHUB_TOKEN = \"" + prToken + "\"\n\n\ndef f(x):\n    return x  # noqa\n"}
	return diffGateFixture{root: root, c1: c1, c2: c2, c3: c3, legacyToken: legacyToken, prA: prA}
}

func (f diffGateFixture) branchWith(t *testing.T, name, from string, files map[string]string) {
	t.Helper()
	gitIn(t, f.root, "checkout", "-q", "-B", name, from)
	commitFiles(t, f.root, name, files)
}

// (a) report profile at the PR's base: exit 0, two would_block rows persisted.
func assertReportProfilePR(t *testing.T, f diffGateFixture) {
	t.Helper()
	f.branchWith(t, "pr-a", f.c1, f.prA)
	code, report := runDiff(t, f.root, "--base", f.c1)
	if code != 0 || report.Gate.Status != "clean" || report.Gate.Profile != "report" {
		t.Fatalf("(a) report profile: exit %d gate %+v", code, report.Gate)
	}
	if len(report.WouldBlock) != 2 || report.Ranking.Method == "" || len(report.Health) == 0 {
		t.Fatalf("(a) would_block=%d ranking=%q health=%d", len(report.WouldBlock), report.Ranking.Method, len(report.Health))
	}
	counts := wouldBlockChecks(t, f.root)
	if counts["security.secrets.github-pat"] != 1 || counts[guardrails.UnscopedCheckID] != 1 || len(counts) != 2 {
		t.Fatalf("(a) persisted would_block rows = %v", counts)
	}
	assertContextualOptional(t, report)
}

func assertContextualOptional(t *testing.T, report diffReport) {
	t.Helper()
	for _, health := range report.Health {
		if health.Lane == "contextual" && health.Required {
			t.Fatalf("optional lane must never be required: %+v", health)
		}
	}
}

func blockingIssueID(t *testing.T, report diffReport) string {
	t.Helper()
	var blocking string
	for _, item := range report.Issues {
		if item.Blocking {
			if item.CheckID != guardrails.UnscopedCheckID || item.Attribution != "introduced" {
				t.Fatalf("(b) blocked on %+v", item)
			}
			blocking = item.IssueID
		}
	}
	return blocking
}

// (b) base policy activates the check at the current version: exit 1, exactly that issue.
func assertActivatedCheckGates(t *testing.T, f diffGateFixture) string {
	t.Helper()
	f.branchWith(t, "pr-a-active", f.c2, f.prA)
	code, report := runDiff(t, f.root, "--base", f.c2)
	if code != 1 || report.Gate.Status != "gated" || len(report.Gate.Blocking) != 1 {
		t.Fatalf("(b) exit %d gate %+v", code, report.Gate)
	}
	blocking := blockingIssueID(t, report)
	if blocking == "" || blocking != report.Gate.Blocking[0] {
		t.Fatalf("(b) blocking issue not listed: %q %v", blocking, report.Gate.Blocking)
	}
	if len(report.WouldBlock) != 1 || report.WouldBlock[0].CheckID != "security.secrets.github-pat" {
		t.Fatalf("(b) PAT must stay would_block: %+v", report.WouldBlock)
	}
	return blocking
}

func loadIssueFrom(t *testing.T, root, id string) (domain.Issue, error) {
	t.Helper()
	store, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOrFail(t, store)
	return store.LoadIssue(context.Background(), id)
}

// (e) override: a non-maintainer is rejected and logged, a maintainer is accepted as a disposition.
func assertMaintainerOverride(t *testing.T, f diffGateFixture, blocking string) {
	t.Helper()
	allow := "scout:allow " + guardrails.UnscopedCheckID + " reason=legacy shim ref=#7"
	code, report := runDiff(t, f.root, "--base", f.c2, "--override-actor", "mallory", "--override-text", allow)
	if code != 1 || len(report.Gate.Overrides) != 1 || report.Gate.Overrides[0].Accepted {
		t.Fatalf("(e) non-maintainer override: exit %d %+v", code, report.Gate.Overrides)
	}
	code, report = runDiff(t, f.root, "--base", f.c2, "--override-actor", "alice", "--override-text", allow)
	if code != 0 || len(report.Gate.Overrides) != 1 || !report.Gate.Overrides[0].Accepted {
		t.Fatalf("(e) maintainer override: exit %d %+v", code, report.Gate)
	}
	assertOverrideDisposition(t, f.root, blocking)
}

func assertOverrideDisposition(t *testing.T, root, blocking string) {
	t.Helper()
	accepted, err := loadIssueFrom(t, root, blocking)
	if err != nil || accepted.State != domain.IssueAccepted || !strings.Contains(accepted.DispositionReason, "pr=#7") {
		t.Fatalf("(e) disposition = %+v, %v", accepted, err)
	}
	last := accepted.History[len(accepted.History)-1]
	if last.Actor != "alice" || !strings.Contains(last.Reason, "legacy shim") {
		t.Fatalf("(e) override audit entry = %+v", last)
	}
}

// (c) PR B touches only pre-existing debt: exit 0, nothing introduced.
func assertPreExistingDebtPasses(t *testing.T, f diffGateFixture) {
	t.Helper()
	f.branchWith(t, "pr-b", f.c2, map[string]string{"pkg/legacy.py": "OLD = \"" + f.legacyToken + "\"\n\n\ndef old(x):\n    # touched\n    return x\n"})
	code, report := runDiff(t, f.root, "--base", f.c2)
	if code != 0 || len(report.Gate.Blocking) != 0 || len(report.WouldBlock) != 0 {
		t.Fatalf("(c) exit %d gate %+v would_block %+v", code, report.Gate, report.WouldBlock)
	}
	if report.PreExisting < 1 || report.Comparison != "recomputed" {
		t.Fatalf("(c) pre-existing = %d comparison = %q", report.PreExisting, report.Comparison)
	}
}

// (d) a rule-version bump in policy returns the check to shadow: exit 0.
// A different suppression: the earlier one carries a maintainer disposition.
func assertBumpedCheckReturnsToShadow(t *testing.T, f diffGateFixture) {
	t.Helper()
	bumpedPR := map[string]string{"pkg/other.py": "def g(y):\n    z = y  # noqa\n    return z\n"}
	f.branchWith(t, "pr-a-bumped", f.c3, bumpedPR)
	code, report := runDiff(t, f.root, "--base", f.c3)
	if code != 0 || len(report.Gate.Blocking) != 0 {
		t.Fatalf("(d) exit %d gate %+v", code, report.Gate)
	}
	shadowed := false
	for _, row := range report.WouldBlock {
		shadowed = shadowed || row.CheckID == guardrails.UnscopedCheckID
	}
	if !shadowed {
		t.Fatalf("(d) bumped check must be would_block again: %+v", report.WouldBlock)
	}
}

// REQ-179, PR A / PR B and version bump on a real git fixture.
func TestDiffGateScenario(t *testing.T) {
	requireTools(t, "ruff", "gitleaks", "semgrep")
	binary := pinnedClawpatch(t)
	t.Setenv("SCOUT_CLAWPATCH_BINARY", binary)

	f := newDiffGateFixture(t)
	assertReportProfilePR(t, f)
	blocking := assertActivatedCheckGates(t, f)
	assertMaintainerOverride(t, f, blocking)
	assertPreExistingDebtPasses(t, f)
	assertBumpedCheckReturnsToShadow(t, f)
}

func gitRev(t *testing.T, root, ref string) string {
	t.Helper()
	out, err := exec.CommandContext(boundedTestContext(t), "git", "-C", root, "rev-parse", ref).Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

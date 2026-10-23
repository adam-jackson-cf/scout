package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
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
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = writer
	captured := make(chan []byte, 1)
	go func() {
		data, _ := io.ReadAll(reader)
		captured <- data
	}()
	code := run(context.Background(), append([]string{"diff", "--root", root, "--json"}, extra...))
	os.Stdout = original
	_ = writer.Close()
	var outcome diffOutcome
	data := <-captured
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
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = writer
	captured := make(chan []byte, 1)
	go func() {
		data, _ := io.ReadAll(reader)
		captured <- data
	}()
	code := run(context.Background(), append([]string{"diff", "--root", root, "--json"}, extra...))
	os.Stdout = original
	_ = writer.Close()
	var outcome struct {
		Result map[string]any `json:"result"`
	}
	if err := json.Unmarshal(<-captured, &outcome); err != nil {
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
	defer store.Close()
	rows, err := store.ListWouldBlock(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, row := range rows {
		counts[row.CheckID]++
	}
	return counts
}

// Design §11 Phase 2, PR A / PR B and version bump on a real git fixture.
func TestDiffGateScenario(t *testing.T) {
	for _, tool := range []string{"ruff", "gitleaks", "semgrep"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not installed (source .tools/activate)", tool)
		}
	}
	binary := pinnedClawpatch(t)
	t.Setenv("SCOUT_CLAWPATCH_BINARY", binary)

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
	branchWith := func(name, from string, files map[string]string) {
		gitIn(t, root, "checkout", "-q", "-B", name, from)
		commitFiles(t, root, name, files)
	}

	// (a) report profile at the PR's base: exit 0, two would_block rows persisted.
	branchWith("pr-a", c1, prA)
	code, report := runDiff(t, root, "--base", c1)
	if code != 0 || report.Gate.Status != "clean" || report.Gate.Profile != "report" {
		t.Fatalf("(a) report profile: exit %d gate %+v", code, report.Gate)
	}
	if len(report.WouldBlock) != 2 || report.Ranking.Method == "" || len(report.Health) == 0 {
		t.Fatalf("(a) would_block=%d ranking=%q health=%d", len(report.WouldBlock), report.Ranking.Method, len(report.Health))
	}
	counts := wouldBlockChecks(t, root)
	if counts["security.secrets.github-pat"] != 1 || counts[guardrails.UnscopedCheckID] != 1 || len(counts) != 2 {
		t.Fatalf("(a) persisted would_block rows = %v", counts)
	}
	for _, health := range report.Health {
		if health.Lane == "contextual" && health.Required {
			t.Fatalf("optional lane must never be required: %+v", health)
		}
	}

	// (b) base policy activates the check at the current version: exit 1, exactly that issue.
	branchWith("pr-a-active", c2, prA)
	code, report = runDiff(t, root, "--base", c2)
	if code != 1 || report.Gate.Status != "gated" || len(report.Gate.Blocking) != 1 {
		t.Fatalf("(b) exit %d gate %+v", code, report.Gate)
	}
	var blocking string
	for _, item := range report.Issues {
		if item.Blocking {
			if item.CheckID != guardrails.UnscopedCheckID || item.Attribution != "introduced" {
				t.Fatalf("(b) blocked on %+v", item)
			}
			blocking = item.IssueID
		}
	}
	if blocking == "" || blocking != report.Gate.Blocking[0] {
		t.Fatalf("(b) blocking issue not listed: %q %v", blocking, report.Gate.Blocking)
	}
	if len(report.WouldBlock) != 1 || report.WouldBlock[0].CheckID != "security.secrets.github-pat" {
		t.Fatalf("(b) PAT must stay would_block: %+v", report.WouldBlock)
	}

	// (e) override: a non-maintainer is rejected and logged, a maintainer is accepted as a disposition.
	allow := "scout:allow " + guardrails.UnscopedCheckID + " reason=legacy shim ref=#7"
	code, report = runDiff(t, root, "--base", c2, "--override-actor", "mallory", "--override-text", allow)
	if code != 1 || len(report.Gate.Overrides) != 1 || report.Gate.Overrides[0].Accepted {
		t.Fatalf("(e) non-maintainer override: exit %d %+v", code, report.Gate.Overrides)
	}
	code, report = runDiff(t, root, "--base", c2, "--override-actor", "alice", "--override-text", allow)
	if code != 0 || len(report.Gate.Overrides) != 1 || !report.Gate.Overrides[0].Accepted {
		t.Fatalf("(e) maintainer override: exit %d %+v", code, report.Gate)
	}
	store, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	accepted, err := store.LoadIssue(context.Background(), blocking)
	store.Close()
	if err != nil || accepted.State != domain.IssueAccepted || !strings.Contains(accepted.DispositionReason, "pr=#7") {
		t.Fatalf("(e) disposition = %+v, %v", accepted, err)
	}
	last := accepted.History[len(accepted.History)-1]
	if last.Actor != "alice" || !strings.Contains(last.Reason, "legacy shim") {
		t.Fatalf("(e) override audit entry = %+v", last)
	}

	// (c) PR B touches only pre-existing debt: exit 0, nothing introduced.
	branchWith("pr-b", c2, map[string]string{"pkg/legacy.py": "OLD = \"" + legacyToken + "\"\n\n\ndef old(x):\n    # touched\n    return x\n"})
	code, report = runDiff(t, root, "--base", c2)
	if code != 0 || len(report.Gate.Blocking) != 0 || len(report.WouldBlock) != 0 {
		t.Fatalf("(c) exit %d gate %+v would_block %+v", code, report.Gate, report.WouldBlock)
	}
	if report.PreExisting < 1 || report.Comparison != "recomputed" {
		t.Fatalf("(c) pre-existing = %d comparison = %q", report.PreExisting, report.Comparison)
	}

	// (d) a rule-version bump in policy returns the check to shadow: exit 0.
	// A different suppression: the earlier one carries a maintainer disposition.
	bumpedPR := map[string]string{"pkg/other.py": "def g(y):\n    z = y  # noqa\n    return z\n"}
	branchWith("pr-a-bumped", c3, bumpedPR)
	code, report = runDiff(t, root, "--base", c3)
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

func gitRev(t *testing.T, root, ref string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", root, "rev-parse", ref).Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

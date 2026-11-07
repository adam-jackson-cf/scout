package association

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"scout/internal/areas"
	"scout/internal/domain"
	"scout/internal/policy"
)

// fakeTerra is an actual executable boundary, never a real model CLI. Every
// output file and invocation trace is isolated to this test's temporary tree.
func fakeTerra(t *testing.T, answer, diagnostic, before string) string {
	t.Helper()
	dir := t.TempDir()
	script := "#!/bin/sh\necho call >> '" + dir + "/calls'\n" + "printf '%s\\n' \"$*\" >> '" + dir + "/args'\n" + "/usr/bin/env > '" + dir + "/env'\n/bin/pwd > '" + dir + "/cwd'\n/bin/cat > '" + dir + "/prompt'\n" + before + "\nout=''\nwhile [ $# -gt 0 ]; do if [ \"$1\" = '-o' ]; then shift; out=$1; fi; shift; done\nprintf '%s' '" + strings.ReplaceAll(answer, "'", "'\"'\"'") + "' > \"$out\"\nprintf '%s\\n' '" + diagnostic + "' >&2\n"
	if err := os.WriteFile(filepath.Join(dir, "codex"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "claude"), []byte("#!/bin/sh\necho invoked >> '"+dir+"/claude-calls'\nexit 99\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	return dir
}
func terraConfig() policy.AssociationFallback {
	return policy.AssociationFallback{Model: "gpt-6.1-terra", Effort: "high", Timeout: 10 * time.Second, MaxPerRun: 40}
}
func terraCatalogue() []areas.Area {
	return []areas.Area{{ScoutAreaID: "a", Title: "Requests", Origin: areas.OriginAgent}, {ScoutAreaID: "b", Title: "Build", Origin: areas.OriginAgent}}
}
func terraExisting() string {
	return `{"decision":"existing","area_id":"a","new_title":null,"new_summary":null}`
}
func terraNew() string {
	return `{"decision":"new","area_id":null,"new_title":"Artifacts","new_summary":"Builds application artifacts."}`
}
func terraDiagnostic() string { return "model: gpt-6.1-terra\nreasoning effort: high" }
func terraCallCount(t *testing.T, dir string) int {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "calls"))
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatal(err)
	}
	return strings.Count(string(b), "call\n")
}
func assertNoClaude(t *testing.T, dir string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(dir, "claude-calls")); !os.IsNotExist(err) {
		t.Fatalf("claude invoked or trace unavailable: %v", err)
	}
}

func TestTerraExistingBoundary(t *testing.T) {
	dir := fakeTerra(t, terraExisting(), terraDiagnostic(), "")
	t.Setenv("OPENAI_API_KEY", "fixture-not-a-secret")
	t.Setenv("GITHUB_TOKEN", "fixture-not-a-secret")
	state := json.RawMessage(`{"unit":{"members":[]},"pr":{"title":"untrusted title","body":"untrusted body"},"catalogue":[]}`)
	answer, err := NewTerraDecider(terraConfig()).Decide(context.Background(), state, terraCatalogue())
	if err != nil || answer.Decision != "existing" || answer.AreaID != "a" {
		t.Fatalf("answer=%+v err=%v", answer, err)
	}
	if terraCallCount(t, dir) != 1 {
		t.Fatal("successful call retried")
	}
	b, err := os.ReadFile(filepath.Join(dir, "args"))
	if err != nil {
		t.Fatal(err)
	}
	args := string(b)
	if !strings.HasPrefix(args, "exec -m gpt-6.1-terra -c model_reasoning_effort=high ") {
		t.Fatalf("wrong invocation: %s", args)
	}
	for _, flag := range strings.Fields("--ephemeral --ignore-user-config --ignore-rules --skip-git-repo-check") {
		if !strings.Contains(args, flag) {
			t.Fatalf("missing %s: %s", flag, args)
		}
	}
	if !strings.Contains(args, "-s read-only") || !strings.Contains(args, "-C ") || !strings.Contains(args, "-o ") || !strings.HasSuffix(strings.TrimSpace(args), " -") {
		t.Fatalf("unsafe invocation %s", args)
	}
	b, err = os.ReadFile(filepath.Join(dir, "env"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "OPENAI_API_KEY=") || strings.Contains(string(b), "GITHUB_TOKEN=") {
		t.Fatal("credentials reached codex")
	}
	b, err = os.ReadFile(filepath.Join(dir, "prompt"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(b), TerraPrompt()+"\n=== STATE (JSON) ===\n") || !strings.Contains(string(b), string(state)) {
		t.Fatalf("prompt does not contain exact template and state")
	}
	b, err = os.ReadFile(filepath.Join(dir, "cwd"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(b)) == dir || strings.TrimSpace(string(b)) == "" {
		t.Fatal("codex not isolated in empty directory")
	}
	assertNoClaude(t, dir)
}
func TestTerraNewBoundary(t *testing.T) {
	dir := fakeTerra(t, terraNew(), terraDiagnostic(), "")
	a, err := NewTerraDecider(terraConfig()).Decide(context.Background(), json.RawMessage(`{}`), terraCatalogue())
	if err != nil || a.Decision != "new" || a.NewTitle != "Artifacts" || a.NewSummary != "Builds application artifacts." {
		t.Fatalf("%+v %v", a, err)
	}
	if terraCallCount(t, dir) != 1 {
		t.Fatal("new answer retried")
	}
}
func terraRejected(t *testing.T, answer, diagnostic, before string, timeout time.Duration) {
	t.Helper()
	dir := fakeTerra(t, answer, diagnostic, before)
	cfg := terraConfig()
	cfg.Timeout = timeout
	_, err := NewTerraDecider(cfg).Decide(context.Background(), json.RawMessage(`{}`), terraCatalogue())
	if err == nil || terraCallCount(t, dir) != 2 {
		t.Fatalf("need error after exactly one retry: err=%v calls=%d", err, terraCallCount(t, dir))
	}
	assertNoClaude(t, dir)
}
func TestTerraRejectsModelMismatch(t *testing.T) {
	terraRejected(t, terraExisting(), "model: wrong\nreasoning effort: high", "", 10*time.Second)
}
func TestTerraRejectsEffortMismatch(t *testing.T) {
	terraRejected(t, terraExisting(), "model: gpt-6.1-terra\nreasoning effort: low", "", 10*time.Second)
}
func TestTerraRejectsExecutedCommand(t *testing.T) {
	terraRejected(t, terraExisting(), terraDiagnostic()+"\nexec\n/bin/pwd", "", 10*time.Second)
}
func TestTerraRejectsTimeout(t *testing.T) {
	root, store, opts, _ := terraLaneFixture(t, 1, "none")
	opts.Mapping.Fallback.Timeout = 3 * time.Second
	dir := fakeTerra(t, terraExisting(), terraDiagnostic(), "exec /bin/sleep 30")
	enableTerraGit(t, dir)
	health, rows := executeTerraLane(t, root, store, opts)
	fallbackHealth(t, health, domain.LaneFailed, "", 1)
	if terraCallCount(t, dir) != 2 || len(rows) != 1 || rows[0].Outcome != domain.AssociationUnresolved || rows[0].Reason != domain.AssociationReasonTerraFailure {
		t.Fatalf("timeout must exhaust exactly one retry: calls=%d rows=%+v", terraCallCount(t, dir), rows)
	}
	assertNoClaude(t, dir)
}
func TestTerraRejectsUnparsable(t *testing.T) {
	terraRejected(t, "not JSON", terraDiagnostic(), "", 10*time.Second)
}
func TestTerraRejectsUnknownArea(t *testing.T) {
	terraRejected(t, `{"decision":"existing","area_id":"missing"}`, terraDiagnostic(), "", 10*time.Second)
}
func TestTerraRejectsProcessError(t *testing.T) {
	terraRejected(t, terraExisting(), terraDiagnostic(), "exit 17", 10*time.Second)
}
func TestTerraRejectsMissingNewDescription(t *testing.T) {
	terraRejected(t, `{"decision":"new","new_title":"","new_summary":""}`, terraDiagnostic(), "", 10*time.Second)
}
func TestTerraPromptPin(t *testing.T) {
	sum := sha256.Sum256([]byte(TerraPrompt()))
	got := hex.EncodeToString(sum[:])
	const want = "ca24fd764b745dbcb261a1aa6a0d740710809e82a13a8a103e55bbc4d8555e76"
	if got != want || TerraPromptDigest != want {
		t.Fatalf("prompt digest %s constant %s", got, TerraPromptDigest)
	}
}
func TestTerraStalePrompt(t *testing.T) {
	staleProof(t, func(f *Fingerprint) { f.PromptDigest = "changed-terra-prompt" }, true)
}
func TestTerraStaleModel(t *testing.T) {
	staleProof(t, func(f *Fingerprint) { f.FallbackModel = "gpt-6.1-terra-other-pin" }, true)
}

func TestTerraRecordedReplay(t *testing.T) {
	rows := readReplay[struct {
		CaseID string `json:"case_id"`
		Jev    struct {
			Output string
			AreaID string `json:"area_id"`
		} `json:"jev"`
		Gold struct {
			Covered    bool
			Acceptable []string
		} `json:"gold"`
		Call *struct {
			Raw    string
			Parsed TerraAnswer
		} `json:"call"`
	}](t, "testdata/terra-replay.jsonl")
	assigned, correct, pending, pendingCorrect, unresolved := 0, 0, 0, 0, 0
	for _, row := range rows {
		output, id := row.Jev.Output, row.Jev.AreaID
		if output != "assigned" {
			if row.Call == nil {
				t.Fatalf("no recorded escalation for %s", row.CaseID)
			}
			dir := fakeTerra(t, row.Call.Raw, terraDiagnostic(), "")
			cat := make([]areas.Area, 0, len(row.Gold.Acceptable)+1)
			for _, id := range row.Gold.Acceptable {
				cat = append(cat, areas.Area{ScoutAreaID: id, Origin: areas.OriginAgent})
			}
			cat = append(cat, areas.Area{ScoutAreaID: row.Call.Parsed.AreaID, Origin: areas.OriginAgent})
			a, err := NewTerraDecider(terraConfig()).Decide(context.Background(), json.RawMessage(fmt.Sprintf(`{"case_id":%q}`, row.CaseID)), cat)
			if err != nil {
				unresolved++
				t.Errorf("%s: %v", row.CaseID, err)
				continue
			}
			if terraCallCount(t, dir) != 1 {
				t.Fatal("recorded answer retried")
			}
			if a.Decision == "new" {
				pending++
				if !row.Gold.Covered {
					pendingCorrect++
				}
				continue
			}
			output, id = "assigned", a.AreaID
		}
		if output == "assigned" {
			assigned++
			for _, accepted := range row.Gold.Acceptable {
				if id == accepted && row.Gold.Covered {
					correct++
					break
				}
			}
		}
	}
	if len(rows) != 92 || assigned != 77 || correct != 71 || unresolved != 0 || pending != 15 || pendingCorrect != 13 {
		t.Fatalf("n=%d assigned=%d correct=%d unresolved=%d pending=%d gold-none-pending=%d", len(rows), assigned, correct, unresolved, pending, pendingCorrect)
	}
}

func TestTerraRetryRecovers(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "first")
	before := "if [ ! -f '" + marker + "' ]; then : > '" + marker + "'; exit 19; fi"
	dir := fakeTerra(t, terraExisting(), terraDiagnostic(), before)
	a, err := NewTerraDecider(terraConfig()).Decide(context.Background(), json.RawMessage(`{}`), terraCatalogue())
	if err != nil || a.AreaID != "a" || terraCallCount(t, dir) != 2 {
		t.Fatalf("answer=%+v err=%v calls=%d", a, err, terraCallCount(t, dir))
	}
}

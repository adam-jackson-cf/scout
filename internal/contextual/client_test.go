package contextual

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"scout/internal/domain"
	"scout/internal/modelcli"
)

const testSecret = "ghp_abcdefghijklmnopqrstuvwxyz1234567890"

func flaggedAnswer(t *testing.T) string {
	t.Helper()
	return answerFor(t, modelResponse{
		Status:                 "flagged",
		Reason:                 "token=" + testSecret,
		DeterministicBlindSpot: "deterministic review did not inspect token=" + testSecret,
		Inspected:              []modelCitation{{Path: "main.go", StartLine: 1, EndLine: 1, Quote: testSecret}},
		Findings: []modelFinding{{
			Mechanism:    "credential is embedded in source",
			Evidence:     modelCitation{Path: "main.go", StartLine: 1, EndLine: 1, Quote: testSecret},
			Confidence:   "high",
			Uncertainty:  "the credential's current validity is unknown",
			Reproduction: "inspect token=" + testSecret,
			Severity:     json.RawMessage(`{"impact":3,"likelihood":2}`),
		}},
	})
}

func writeSecretSource(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	content := "const credential = \"" + testSecret + "\"\n"
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return root, content
}

func TestClaudeAdapterUsesToollessArgvCredentialFreeEnvAndRecordsResolvedModel(t *testing.T) {
	setCredentialEnv(t)
	root, content := writeSecretSource(t)
	response := writeResponse(t, claudeEnvelope(t, flaggedAnswer(t), "claude-sonnet-4-5-20250929"))
	fake := installFakeCLI(t, "claude", "cat '"+response+"'")
	tool := domain.AdapterResult{ResultID: "adapter-1", RunID: "run-1", Adapter: "semgrep", AdapterVersion: "1.0", Status: domain.AdapterCompleted}

	request := testReviewRequest(root, tool)
	request.Feature.OwnedFiles[0].Reason = "token=" + testSecret
	result, err := newTestClient(t, "").Review(context.Background(), request)
	if err != nil {
		t.Fatalf("Review() error = %v", err)
	}

	wantArgv := strings.Join([]string{
		"-p", "--output-format", "json", "--tools", "", "--permission-mode", "dontAsk",
		"--no-session-persistence", "--disable-slash-commands", "--strict-mcp-config", "--mcp-config", `{"mcpServers":{}}`,
		"--no-chrome", "--system-prompt", reviewInstructions,
	}, "\n") + "\n"
	if got := fake.read(t, fake.argv); got != wantArgv {
		t.Fatalf("claude argv =\n%q\nwant\n%q", got, wantArgv)
	}
	assertNoCredentialEnv(t, fake.read(t, fake.env))
	if cwd := strings.TrimSpace(fake.read(t, fake.cwd)); cwd == root || strings.HasPrefix(cwd, root) {
		t.Fatalf("CLI ran inside the repository: %s", cwd)
	}
	stdin := fake.read(t, fake.stdin)
	if strings.Contains(stdin, testSecret) || !strings.Contains(stdin, "[R]") {
		t.Fatalf("source sent to the CLI was not redacted: %q", stdin)
	}
	var prompt promptRequest
	if err := json.Unmarshal([]byte(stdin), &prompt); err != nil {
		t.Fatalf("stdin is not the contextual prompt: %v", err)
	}
	if prompt.Snapshot != "snapshot-1" || len(prompt.Manifest.Inspected) != 1 || prompt.Manifest.Inspected[0].SourceDigest != sourceDigest([]byte(content)) {
		t.Fatalf("prompt lost identity or manifest: %#v", prompt)
	}

	if result.Model == nil || result.Model.CLI != "claude" || result.Model.ResolvedModel != "claude-sonnet-4-5-20250929" {
		t.Fatalf("resolved model not recorded: %#v", result.Model)
	}
	if result.Coverage.Status != domain.CoverageFlagged || strings.Contains(result.Coverage.Reason, testSecret) {
		t.Fatalf("coverage = %#v", result.Coverage)
	}
	if len(result.Observations) != 1 || result.Observations[0].ModelVersion != "claude-sonnet-4-5-20250929" || strings.Contains(result.Observations[0].Evidence[0].RedactedQuote, testSecret) {
		t.Fatalf("observation did not carry the resolved model and safe evidence: %#v", result.Observations)
	}
}

func TestClaudeAdapterPassesConfiguredModel(t *testing.T) {
	root, _ := writeSecretSource(t)
	response := writeResponse(t, claudeEnvelope(t, flaggedAnswer(t), "claude-haiku-4-5"))
	fake := installFakeCLI(t, "claude", "cat '"+response+"'")
	client, err := NewClient(Config{CLI: "claude", Model: "claude-haiku-4-5", EgressAuthorized: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Review(context.Background(), testReviewRequest(root)); err != nil {
		t.Fatal(err)
	}
	if argv := fake.read(t, fake.argv); !strings.HasSuffix(argv, "--model\nclaude-haiku-4-5\n") {
		t.Fatalf("model flag missing from argv: %q", argv)
	}
}

func TestCodexAdapterUsesReadOnlySandboxEphemeralStdinAndRecordsModel(t *testing.T) {
	setCredentialEnv(t)
	root, _ := writeSecretSource(t)
	answer := writeResponse(t, flaggedAnswer(t))
	// The fake copies the answer to the --output-last-message path it was given.
	fake := installFakeCLI(t, "codex", `out=""; prev=""; for a in "$@"; do if [ "$prev" = --output-last-message ]; then out="$a"; fi; prev="$a"; done
cp '`+answer+`' "$out"
echo "model: gpt-5.1-codex" >&2`)
	result, err := newTestClient(t, "codex").Review(context.Background(), testReviewRequest(root))
	if err != nil {
		t.Fatalf("Review() error = %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(fake.read(t, fake.argv), "\n"), "\n")
	if len(lines) != 8 || !strings.HasSuffix(lines[6], "last-message.txt") {
		t.Fatalf("codex argv = %q", lines)
	}
	lines[6] = "<last-message>"
	if got, want := strings.Join(lines, " "), "exec --sandbox read-only --skip-git-repo-check --ephemeral --output-last-message <last-message> -"; got != want {
		t.Fatalf("codex argv = %q, want %q", got, want)
	}
	assertNoCredentialEnv(t, fake.read(t, fake.env))
	if stdin := fake.read(t, fake.stdin); !strings.HasPrefix(stdin, reviewInstructions) || strings.Contains(stdin, testSecret) {
		t.Fatalf("codex stdin must carry the fixed instructions and only redacted source")
	}
	if result.Model == nil || result.Model.CLI != "codex" || result.Model.ResolvedModel != "gpt-5.1-codex" {
		t.Fatalf("resolved model not recorded: %#v", result.Model)
	}
	if result.Coverage.Status != domain.CoverageFlagged || len(result.Observations) != 1 || result.Observations[0].ModelVersion != "gpt-5.1-codex" {
		t.Fatalf("codex review = %#v", result)
	}
}

func TestMalformedModelOutputIsIncompleteNotClean(t *testing.T) {
	root, _ := writeSecretSource(t)
	for name, output := range map[string]string{
		"not JSON at all":          "I think this file is fine.",
		"envelope with prose":      claudeEnvelope(t, "Here is the JSON: {\"status\":\"covered\"}", "claude-x"),
		"fenced JSON is not JSON":  claudeEnvelope(t, "```json\n"+flaggedAnswer(t)+"\n```", "claude-x"),
		"unknown answer field":     claudeEnvelope(t, `{"status":"covered","deterministicBlindSpot":"none_observed","extra":1}`, "claude-x"),
		"error envelope":           `{"type":"result","is_error":true,"result":"boom"}`,
		"covered without evidence": claudeEnvelope(t, `{"status":"covered","deterministicBlindSpot":"none_observed"}`, "claude-x"),
	} {
		t.Run(name, func(t *testing.T) {
			fake := installFakeCLI(t, "claude", "cat '"+writeResponse(t, output)+"'")
			result, err := newTestClient(t, "claude").Review(context.Background(), testReviewRequest(root))
			if !fake.invoked() {
				t.Fatal("CLI was not invoked")
			}
			if err == nil || result.Coverage.Status != domain.CoverageIncomplete || len(result.Observations) != 0 {
				t.Fatalf("malformed output must be incomplete: err=%v result=%#v", err, result)
			}
		})
	}
}

func TestFailingCLIIsIncomplete(t *testing.T) {
	root, _ := writeSecretSource(t)
	installFakeCLI(t, "codex", "exit 7")
	result, err := newTestClient(t, "codex").Review(context.Background(), testReviewRequest(root))
	if err == nil || result.Coverage.Status != domain.CoverageIncomplete {
		t.Fatalf("failing CLI must be incomplete: err=%v coverage=%#v", err, result.Coverage)
	}
}

func TestNoModelCLIIsSkippedWithoutFailure(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	root, _ := writeSecretSource(t)
	result, err := newTestClient(t, "").Review(context.Background(), testReviewRequest(root))
	if !errors.Is(err, modelcli.ErrNoCLI) {
		t.Fatalf("Review() error = %v, want ErrNoCLI", err)
	}
	if result.Coverage.Status != domain.CoverageSkipped || result.Coverage.Reason != "skipped (no model CLI)" {
		t.Fatalf("coverage = %#v", result.Coverage)
	}
	if err := CheckPrerequisites(Config{EgressAuthorized: true}); !errors.Is(err, modelcli.ErrNoCLI) {
		t.Fatalf("CheckPrerequisites() = %v, want ErrNoCLI", err)
	}
	// Skipped, not blocked, even when egress was not authorized either.
	if err := CheckPrerequisites(Config{}); !errors.Is(err, modelcli.ErrNoCLI) {
		t.Fatalf("CheckPrerequisites() without egress = %v, want ErrNoCLI", err)
	}
}

func TestReviewUsesFirstAvailableCLIAndHonoursPreference(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "codex"), []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	if name, _, err := modelcli.Detect(""); err != nil || name != modelcli.Codex {
		t.Fatalf("Detect() = %v, %v; want codex", name, err)
	}
	if _, _, err := modelcli.Detect("claude"); !errors.Is(err, modelcli.ErrNoCLI) {
		t.Fatalf("Detect(claude) = %v, want ErrNoCLI", err)
	}
	if _, err := NewClient(Config{CLI: "gemini"}); err == nil {
		t.Fatal("a non-Anthropic/OpenAI CLI was accepted")
	}
}

func TestReviewMarksSourceDriftIncomplete(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "main.go")
	if err := os.WriteFile(path, []byte("const enabled = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	answer := answerFor(t, modelResponse{
		Status:                 "covered",
		DeterministicBlindSpot: "none_observed",
		Inspected:              []modelCitation{{Path: "main.go", StartLine: 1, EndLine: 1, Quote: "const enabled = true"}},
	})
	response := writeResponse(t, claudeEnvelope(t, answer, "claude-x"))
	installFakeCLI(t, "claude", "echo 'const enabled = false' > '"+path+"'\ncat '"+response+"'")

	result, err := newTestClient(t, "claude").Review(context.Background(), testReviewRequest(root))
	if err == nil {
		t.Fatal("Review() error = nil, want source-drift error")
	}
	if result.Coverage.Status != domain.CoverageIncomplete || !contains(result.Coverage.OmittedPaths, "main.go") {
		t.Fatalf("source drift coverage = %#v", result.Coverage)
	}
}

func TestReviewMarksOversizedRequestIncompleteWithoutInvokingCLI(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("const bounded = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fake := installFakeCLI(t, "claude", "exit 1")
	client, err := NewClient(Config{CLI: "claude", EgressAuthorized: true, MaxPromptBytes: 1})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	result, err := client.Review(context.Background(), testReviewRequest(root))
	if err == nil {
		t.Fatal("Review() error = nil, want prompt-limit error")
	}
	if fake.invoked() || result.Coverage.Status != domain.CoverageIncomplete || !result.Manifest.Truncated {
		t.Fatalf("oversized request result = %#v, CLI invoked = %t", result, fake.invoked())
	}
}

func TestReviewBlocksUnauthorizedSourceEgressWithoutInvokingCLI(t *testing.T) {
	fake := installFakeCLI(t, "claude", "exit 1")
	client, err := NewClient(Config{})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	result, err := client.Review(context.Background(), testReviewRequest(t.TempDir()))
	if err == nil {
		t.Fatal("Review() error = nil, want egress-authorization error")
	}
	if fake.invoked() || result.Coverage.Status != domain.CoverageBlocked {
		t.Fatalf("unauthorized egress coverage = %#v, CLI invoked = %t", result.Coverage, fake.invoked())
	}
	if err := CheckPrerequisites(Config{CLI: "claude"}); err == nil || errors.Is(err, modelcli.ErrNoCLI) {
		t.Fatalf("CheckPrerequisites() = %v, want egress authorization error", err)
	}
}

func assertNoCredentialEnv(t *testing.T, env string) {
	t.Helper()
	for _, line := range strings.Split(env, "\n") {
		name, _, _ := strings.Cut(line, "=")
		if strings.Contains(strings.ToUpper(name), "API_KEY") || name == "ANTHROPIC_AUTH_TOKEN" || strings.HasPrefix(name, "OPENROUTER_") {
			t.Fatalf("credential %s reached the model CLI", name)
		}
	}
	if strings.Contains(env, "must-not-reach-the-cli") {
		t.Fatal("credential value reached the model CLI")
	}
}

func TestValidateModelResponseRejectsUnverifiedQuoteAndMissingFindingEvidence(t *testing.T) {
	request := testReviewRequest(t.TempDir())
	manifest := Manifest{Inspected: []ManifestEntry{{Range: domain.SourceRange{Path: "main.go", StartLine: 1, EndLine: 1, StartByte: 1, EndByte: 12}, Reason: "owned: service", Bytes: 12, SourceDigest: sourceDigest([]byte("const ok = 1"))}}}
	sources := []promptSource{{promptRange: promptRange{Path: "main.go", StartLine: 1, EndLine: 1, StartByte: 1, EndByte: 12}, Reason: "owned: service", Text: "const ok = 1"}}
	raw := []byte(`{"status":"flagged","deterministicBlindSpot":"none_observed","inspected":[{"path":"main.go","startLine":1,"endLine":1,"quote":"const ok = 1"}],"findings":[{"mechanism":"unsafe","evidence":{"path":"main.go","startLine":1,"endLine":1,"quote":"invented"},"confidence":"high","uncertainty":"external state unknown","reproduction":"inspect source","severity":{"impact":2,"likelihood":2}}]}`)
	result, err := validateModelResponse(request, raw, manifest, sources, "model-1", "prompt-1")
	if err == nil {
		t.Fatal("validateModelResponse() error = nil, want invalid quote")
	}
	if len(result.Observations) != 0 {
		t.Fatalf("invalid evidence produced observations: %#v", result.Observations)
	}
}

func TestValidateModelResponseRejectsMissingBlindSpotOrReproduction(t *testing.T) {
	request := testReviewRequest(t.TempDir())
	manifest := Manifest{Inspected: []ManifestEntry{{Range: domain.SourceRange{Path: "main.go", StartLine: 1, EndLine: 1, StartByte: 1, EndByte: 12}, Reason: "owned: service", Bytes: 12, SourceDigest: sourceDigest([]byte("const ok = 1"))}}}
	sources := []promptSource{{promptRange: promptRange{Path: "main.go", StartLine: 1, EndLine: 1, StartByte: 1, EndByte: 12}, Reason: "owned: service", Text: "const ok = 1"}}
	missingBlindSpot := []byte(`{"status":"covered","inspected":[{"path":"main.go","startLine":1,"endLine":1,"quote":"const ok = 1"}]}`)
	if _, err := validateModelResponse(request, missingBlindSpot, manifest, sources, "model-1", "prompt-1"); err == nil {
		t.Fatal("response without blind-spot judgment was accepted")
	}
	missingReproduction := []byte(`{"status":"flagged","deterministicBlindSpot":"none_observed","inspected":[{"path":"main.go","startLine":1,"endLine":1,"quote":"const ok = 1"}],"findings":[{"mechanism":"unsafe","evidence":{"path":"main.go","startLine":1,"endLine":1,"quote":"const ok = 1"},"confidence":"high","uncertainty":"external state unknown","severity":{"impact":2,"likelihood":2}}]}`)
	if _, err := validateModelResponse(request, missingReproduction, manifest, sources, "model-1", "prompt-1"); err == nil {
		t.Fatal("finding without reproduction route was accepted")
	}
}

func TestSeverityForRejectsOtherFamilyDimensions(t *testing.T) {
	if _, err := severityFor(domain.FamilyArchitecture, json.RawMessage(`{"changeRisk":"major","boundary":"systemic"}`)); err == nil {
		t.Fatal("architecture accepted a non-architecture boundary dimension")
	}
	if _, err := severityFor(domain.FamilySecurity, json.RawMessage(`{"impact":2,"likelihood":2,"model":"assertion"}`)); err == nil {
		t.Fatal("security accepted undeclared severity dimensions")
	}
}

// fakeCLI installs an executable named name that logs argv (one per line),
// the environment, the working directory and stdin, then runs body.
// PATH is narrowed to that directory plus the system shell utilities.
type fakeCLI struct{ dir, argv, env, stdin, cwd string }

func installFakeCLI(t *testing.T, name, body string) fakeCLI {
	t.Helper()
	dir := t.TempDir()
	logs := t.TempDir()
	f := fakeCLI{dir: dir, argv: filepath.Join(logs, "argv"), env: filepath.Join(logs, "env"), stdin: filepath.Join(logs, "stdin"), cwd: filepath.Join(logs, "cwd")}
	script := "#!/bin/sh\n" +
		"for a in \"$@\"; do printf '%s\\n' \"$a\" >> '" + f.argv + "'; done\n" +
		"env > '" + f.env + "'\npwd > '" + f.cwd + "'\ncat > '" + f.stdin + "'\n" + body + "\n"
	if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":/usr/bin:/bin")
	return f
}

func (f fakeCLI) read(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("fake CLI did not write %s: %v", filepath.Base(path), err)
	}
	return string(raw)
}

func (f fakeCLI) invoked() bool {
	_, err := os.Stat(f.argv)
	return err == nil
}

func setCredentialEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "OPENROUTER_API_KEY", "CLAWPATCH_OPENAI_API_KEY", "ANTHROPIC_AUTH_TOKEN"} {
		t.Setenv(name, "must-not-reach-the-cli")
	}
}

func answerFor(t *testing.T, response modelResponse) string {
	t.Helper()
	raw, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func claudeEnvelope(t *testing.T, answer, model string) string {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"type": "result", "is_error": false, "result": answer, "modelUsage": map[string]any{model: map[string]int{"outputTokens": 1}}})
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func writeResponse(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "response")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func newTestClient(t *testing.T, cli string) *Client {
	t.Helper()
	client, err := NewClient(Config{CLI: cli, EgressAuthorized: true})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return client
}

func testReviewRequest(root string, tools ...domain.AdapterResult) ReviewRequest {
	raw := json.RawMessage(`{"schemaVersion":1,"featureId":"feature-1","kind":"go","source":"heuristic"}`)
	return ReviewRequest{
		Root: root, Snapshot: "snapshot-1", RunID: "run-1", Family: domain.FamilySecurity, Category: "sensitive_data_flows", ToolResults: tools,
		Feature: domain.Feature{SchemaVersion: 1, FeatureID: "feature-1", Kind: "go", Source: "heuristic", Confidence: domain.ConfidenceHigh, OwnedFiles: []domain.FeatureFileRef{{Path: "main.go", Reason: "service"}}, RawOriginal: raw},
	}
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

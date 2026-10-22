package contextual

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"scout/internal/domain"
)

const testSecret = "ghp_abcdefghijklmnopqrstuvwxyz1234567890"

func TestReviewSendsRedactedSourceManifestAndPersistsValidatedEvidence(t *testing.T) {
	root := t.TempDir()
	content := "const credential = \"" + testSecret + "\"\n"
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	tool := domain.AdapterResult{ResultID: "adapter-1", RunID: "run-1", Adapter: "semgrep", AdapterVersion: "1.0", Status: domain.AdapterCompleted}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var providerRequest struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(request.Body).Decode(&providerRequest); err != nil {
			t.Errorf("decode provider request: %v", err)
			return
		}
		if len(providerRequest.Messages) != 2 {
			t.Errorf("messages = %d, want 2", len(providerRequest.Messages))
			return
		}
		var prompt promptRequest
		if err := json.Unmarshal([]byte(providerRequest.Messages[1].Content), &prompt); err != nil {
			t.Errorf("decode contextual prompt: %v", err)
			return
		}
		if prompt.Snapshot != "snapshot-1" || prompt.Family != domain.FamilySecurity || prompt.Category != "sensitive_data_flows" {
			t.Errorf("prompt did not retain assessment identity: %#v", prompt)
		}
		if len(prompt.Manifest.Inspected) != 1 || prompt.Manifest.Inspected[0].SourceDigest != sourceDigest([]byte(content)) || strings.Contains(prompt.Manifest.Inspected[0].Reason, testSecret) {
			t.Errorf("prompt did not retain safe actual source manifest: %#v", prompt.Manifest)
		}
		if strings.Contains(prompt.Sources[0].Text, testSecret) || !strings.Contains(prompt.Sources[0].Text, "[R]") {
			t.Errorf("provider source was not redacted: %q", prompt.Sources[0].Text)
		}
		if len(prompt.ToolResults) != 1 || prompt.ToolResults[0].ResultID != tool.ResultID || prompt.ToolResults[0].Status != domain.AdapterCompleted {
			t.Errorf("prompt did not retain completed diagnostic: %#v", prompt.ToolResults)
		}
		response := providerResponse{
			Status:                 "flagged",
			Reason:                 "token=" + testSecret,
			DeterministicBlindSpot: "deterministic review did not inspect token=" + testSecret,
			Inspected:              []providerCitation{{Path: "main.go", StartLine: 1, EndLine: 1, Quote: testSecret}},
			Findings: []providerFinding{{
				Mechanism:    "credential is embedded in source",
				Evidence:     providerCitation{Path: "main.go", StartLine: 1, EndLine: 1, Quote: testSecret},
				Confidence:   "high",
				Uncertainty:  "the credential's current validity is unknown",
				Reproduction: "inspect token=" + testSecret,
				Severity:     json.RawMessage(`{"impact":3,"likelihood":2}`),
			}},
		}
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(response); err != nil {
			t.Errorf("encode provider response: %v", err)
		}
	}))
	defer server.Close()
	client := newTestClient(t, server.URL)

	request := testReviewRequest(root, tool)
	request.Feature.OwnedFiles[0].Reason = "token=" + testSecret
	result, err := client.Review(context.Background(), request)
	if err != nil {
		t.Fatalf("Review() error = %v", err)
	}
	if result.Coverage.Status != domain.CoverageFlagged || len(result.Coverage.AdapterResultIDs) != 1 || result.Coverage.AdapterResultIDs[0] != tool.ResultID || strings.Contains(result.Coverage.Reason, testSecret) {
		t.Fatalf("coverage = %#v", result.Coverage)
	}
	if len(result.Observations) != 1 || strings.Contains(result.Observations[0].Evidence[0].RedactedQuote, testSecret) || !strings.Contains(result.Observations[0].Evidence[0].RedactedQuote, "[R]") {
		t.Fatalf("observation did not retain only safe provenance: %#v", result.Observations)
	}
	reproduction := ""
	if result.Metadata != nil {
		reproduction = result.Metadata.ReproductionByObservationID[result.Observations[0].ObservationID]
	}
	if result.Metadata == nil || strings.Contains(result.Metadata.DeterministicBlindSpot, testSecret) || strings.Contains(reproduction, testSecret) || !strings.Contains(reproduction, "[R]") {
		t.Fatalf("metadata did not retain safe review details: %#v", result.Metadata)
	}
}

func TestReviewMarksSourceDriftIncomplete(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "main.go")
	if err := os.WriteFile(path, []byte("const enabled = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := os.WriteFile(path, []byte("const enabled = false\n"), 0o600); err != nil {
			t.Errorf("modify source during provider request: %v", err)
		}
		response := providerResponse{
			Status:                 "covered",
			DeterministicBlindSpot: "none_observed",
			Inspected:              []providerCitation{{Path: "main.go", StartLine: 1, EndLine: 1, Quote: "const enabled = true"}},
		}
		if err := json.NewEncoder(writer).Encode(response); err != nil {
			t.Errorf("encode provider response: %v", err)
		}
	}))
	defer server.Close()

	result, err := newTestClient(t, server.URL).Review(context.Background(), testReviewRequest(root))
	if err == nil {
		t.Fatal("Review() error = nil, want source-drift error")
	}
	if result.Coverage.Status != domain.CoverageIncomplete || !contains(result.Coverage.OmittedPaths, "main.go") {
		t.Fatalf("source drift coverage = %#v", result.Coverage)
	}
}

func TestReviewMarksOversizedProviderRequestIncompleteWithoutEgress(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("const bounded = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		called = true
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("SCOUT_CONTEXTUAL_TEST_TOKEN", "test-token")
	client, err := NewClient(Config{Endpoint: server.URL, Model: "model-1", PromptVersion: "prompt-1", CredentialEnv: "SCOUT_CONTEXTUAL_TEST_TOKEN", EgressAuthorized: true, MaxPromptBytes: 1})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	result, err := client.Review(context.Background(), testReviewRequest(root))
	if err == nil {
		t.Fatal("Review() error = nil, want prompt-limit error")
	}
	if called || result.Coverage.Status != domain.CoverageIncomplete || !result.Manifest.Truncated {
		t.Fatalf("oversized request result = %#v, provider called = %t", result, called)
	}
}

func TestReviewBlocksProhibitedSourceEgress(t *testing.T) {
	client, err := NewClient(Config{})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	result, err := client.Review(context.Background(), testReviewRequest(t.TempDir()))
	if err == nil {
		t.Fatal("Review() error = nil, want egress-authorization error")
	}
	if result.Coverage.Status != domain.CoverageBlocked {
		t.Fatalf("prohibited egress coverage = %#v", result.Coverage)
	}
}

func TestCheckPrerequisitesRequiresAuthorizationAndCredentialWithoutEgress(t *testing.T) {
	config := Config{Endpoint: "https://provider.example/v1/review", Model: "model-1", PromptVersion: "prompt-1", CredentialEnv: "SCOUT_CONTEXTUAL_CHECK_TOKEN", EgressAuthorized: true}
	t.Setenv(config.CredentialEnv, "")
	if err := CheckPrerequisites(config); err == nil {
		t.Fatal("CheckPrerequisites() error = nil, want missing credential")
	}
	t.Setenv(config.CredentialEnv, "test-token")
	if err := CheckPrerequisites(config); err != nil {
		t.Fatalf("CheckPrerequisites() error = %v", err)
	}
	config.EgressAuthorized = false
	if err := CheckPrerequisites(config); err == nil {
		t.Fatal("CheckPrerequisites() error = nil, want egress authorization error")
	}
}

func TestValidateProviderResponseRejectsUnverifiedQuoteAndMissingFindingEvidence(t *testing.T) {
	request := testReviewRequest(t.TempDir())
	manifest := Manifest{Inspected: []ManifestEntry{{Range: domain.SourceRange{Path: "main.go", StartLine: 1, EndLine: 1, StartByte: 1, EndByte: 12}, Reason: "owned: service", Bytes: 12, SourceDigest: sourceDigest([]byte("const ok = 1"))}}}
	sources := []promptSource{{promptRange: promptRange{Path: "main.go", StartLine: 1, EndLine: 1, StartByte: 1, EndByte: 12}, Reason: "owned: service", Text: "const ok = 1"}}
	raw := []byte(`{"status":"flagged","deterministicBlindSpot":"none_observed","inspected":[{"path":"main.go","startLine":1,"endLine":1,"quote":"const ok = 1"}],"findings":[{"mechanism":"unsafe","evidence":{"path":"main.go","startLine":1,"endLine":1,"quote":"invented"},"confidence":"high","uncertainty":"external state unknown","reproduction":"inspect source","severity":{"impact":2,"likelihood":2}}]}`)
	result, err := validateProviderResponse(request, raw, manifest, sources, "model-1", "prompt-1")
	if err == nil {
		t.Fatal("validateProviderResponse() error = nil, want invalid quote")
	}
	if len(result.Observations) != 0 {
		t.Fatalf("invalid evidence produced observations: %#v", result.Observations)
	}
}

func TestValidateProviderResponseRejectsMissingBlindSpotOrReproduction(t *testing.T) {
	request := testReviewRequest(t.TempDir())
	manifest := Manifest{Inspected: []ManifestEntry{{Range: domain.SourceRange{Path: "main.go", StartLine: 1, EndLine: 1, StartByte: 1, EndByte: 12}, Reason: "owned: service", Bytes: 12, SourceDigest: sourceDigest([]byte("const ok = 1"))}}}
	sources := []promptSource{{promptRange: promptRange{Path: "main.go", StartLine: 1, EndLine: 1, StartByte: 1, EndByte: 12}, Reason: "owned: service", Text: "const ok = 1"}}
	missingBlindSpot := []byte(`{"status":"covered","inspected":[{"path":"main.go","startLine":1,"endLine":1,"quote":"const ok = 1"}]}`)
	if _, err := validateProviderResponse(request, missingBlindSpot, manifest, sources, "model-1", "prompt-1"); err == nil {
		t.Fatal("response without blind-spot judgment was accepted")
	}
	missingReproduction := []byte(`{"status":"flagged","deterministicBlindSpot":"none_observed","inspected":[{"path":"main.go","startLine":1,"endLine":1,"quote":"const ok = 1"}],"findings":[{"mechanism":"unsafe","evidence":{"path":"main.go","startLine":1,"endLine":1,"quote":"const ok = 1"},"confidence":"high","uncertainty":"external state unknown","severity":{"impact":2,"likelihood":2}}]}`)
	if _, err := validateProviderResponse(request, missingReproduction, manifest, sources, "model-1", "prompt-1"); err == nil {
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

func newTestClient(t *testing.T, endpoint string) *Client {
	t.Helper()
	t.Setenv("SCOUT_CONTEXTUAL_TEST_TOKEN", "test-token")
	client, err := NewClient(Config{Endpoint: endpoint, Model: "model-1", PromptVersion: "prompt-1", CredentialEnv: "SCOUT_CONTEXTUAL_TEST_TOKEN", EgressAuthorized: true})
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

package contextual

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"scout/internal/domain"
	"scout/internal/modelcli"
)

const testSecret = "ghp_abcdefghijklmnopqrstuvwxyz1234567890"

func writeSecretSource(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	content := "const credential = \"" + testSecret + "\"\n"
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return root, content
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

func TestValidateModelResponseRejectsUnverifiedQuoteAndMissingFindingEvidence(t *testing.T) {
	request := testReviewRequest(t.TempDir())
	manifest := Manifest{Inspected: []ManifestEntry{{Range: domain.SourceRange{Path: "main.go", StartLine: 1, EndLine: 1, StartByte: 1, EndByte: 12}, Reason: "owned: service", Bytes: 12, SourceDigest: sourceDigest([]byte("const ok = 1"))}}}
	sources := []promptSource{{promptRange: promptRange{Path: "main.go", StartLine: 1, EndLine: 1, StartByte: 1, EndByte: 12}, Reason: "owned: service", Text: "const ok = 1"}}
	raw := []byte(`{"status":"flagged","deterministicBlindSpot":"none_observed","inspected":[{"path":"main.go","startLine":1,"endLine":1,"quote":"const ok = 1"}],"findings":[{"mechanism":"unsafe","evidence":{"path":"main.go","startLine":1,"endLine":1,"quote":"invented"},"confidence":"high","uncertainty":"external state unknown","reproduction":"inspect source","severity":{"impact":2,"likelihood":2}}]}`)
	result, err := validateModelResponse(request, raw, manifest, sources, answerProvenance{model: "model-1", promptVersion: "prompt-1"})
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
	if _, err := validateModelResponse(request, missingBlindSpot, manifest, sources, answerProvenance{model: "model-1", promptVersion: "prompt-1"}); err == nil {
		t.Fatal("response without blind-spot judgment was accepted")
	}
	missingReproduction := []byte(`{"status":"flagged","deterministicBlindSpot":"none_observed","inspected":[{"path":"main.go","startLine":1,"endLine":1,"quote":"const ok = 1"}],"findings":[{"mechanism":"unsafe","evidence":{"path":"main.go","startLine":1,"endLine":1,"quote":"const ok = 1"},"confidence":"high","uncertainty":"external state unknown","severity":{"impact":2,"likelihood":2}}]}`)
	if _, err := validateModelResponse(request, missingReproduction, manifest, sources, answerProvenance{model: "model-1", promptVersion: "prompt-1"}); err == nil {
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

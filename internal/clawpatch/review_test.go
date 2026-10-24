package clawpatch

import (
	"os"
	"path/filepath"
	"testing"

	"scout/internal/domain"
)

// testFindingCap is an arbitrary per-feature finding cap for coverage tests.
const testFindingCap = 10

func TestCoverageFromReviewPreservesCompletePromptManifest(t *testing.T) {
	feature := reviewTestFeature("feature", "src/service.go")
	response := reviewResponse{ReviewManifests: []featureReviewManifest{{
		FeatureID: feature.FeatureID,
		Manifest: reviewPromptManifest{
			MaxOwnedFiles:     reviewInt(12),
			MaxContextFiles:   reviewInt(24),
			PromptBytes:       reviewInt(128),
			ApproximateTokens: reviewInt(32),
			IncludedFiles: []reviewPromptFile{{
				Path:               "src/service.go",
				Role:               "owned",
				Bytes:              reviewInt(24),
				IncludedBytes:      reviewInt(24),
				Readable:           true,
				IncludedLineRanges: []reviewLineRange{{StartLine: 1, EndLine: 1}},
			}},
		},
	}}}

	coverage, manifest, incomplete := coverageFromReview(feature, response, testFindingCap)
	if incomplete || coverage.Status != domain.CoverageInspected {
		t.Fatalf("complete prompt manifest became incomplete: %+v", coverage)
	}
	if manifest.PromptBytes != 128 || len(manifest.IncludedFiles) != 1 || manifest.IncludedFiles[0].IncludedBytes != 24 {
		t.Fatalf("review manifest lost prompt attribution: %+v", manifest)
	}
}

func TestCoverageFromReviewMarksDefaultFindingCapIncomplete(t *testing.T) {
	feature := reviewTestFeature("feature", "src/service.go")
	response := reviewResponse{
		Findings: testFindingCap,
		ReviewManifests: []featureReviewManifest{{
			FeatureID: feature.FeatureID,
			Manifest: reviewPromptManifest{
				MaxOwnedFiles:     reviewInt(12),
				MaxContextFiles:   reviewInt(24),
				PromptBytes:       reviewInt(128),
				ApproximateTokens: reviewInt(32),
				IncludedFiles: []reviewPromptFile{{
					Path:               "src/service.go",
					Role:               "owned",
					Bytes:              reviewInt(24),
					IncludedBytes:      reviewInt(24),
					Readable:           true,
					IncludedLineRanges: []reviewLineRange{{StartLine: 1, EndLine: 1}},
				}},
			},
		}},
	}

	coverage, manifest, incomplete := coverageFromReview(feature, response, testFindingCap)
	if !incomplete || coverage.Status != domain.CoverageIncomplete || !manifest.FindingCapReached {
		t.Fatalf("finding-cap review was treated as complete: coverage=%+v manifest=%+v", coverage, manifest)
	}
}

func TestLoadReviewObservationsFiltersPriorRuns(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "service.go"), []byte("package src\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	stateDir := filepath.Join(root, ".scout", "clawpatch")
	if err := os.MkdirAll(filepath.Join(stateDir, "features"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(stateDir, "findings"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "features", "feature.json"), []byte(`{"schemaVersion":1,"featureId":"feature","kind":"service","source":"heuristic","confidence":"high","findingIds":["old","current"]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	for id, runID := range map[string]string{"old": "old-run", "current": "current-run"} {
		record := `{"schemaVersion":1,"findingId":"` + id + `","featureId":"feature","createdByRunId":"` + runID + `","title":"finding","category":"bug","severity":"low","confidence":"high","evidence":[{"path":"src/service.go","startLine":1,"endLine":1}],"reasoning":"supported by source"}`
		if err := os.WriteFile(filepath.Join(stateDir, "findings", id+".json"), []byte(record), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	observations, err := loadReviewObservations(root, stateDir, "feature", "current-run")
	if err != nil {
		t.Fatal(err)
	}
	if len(observations) != 1 || observations[0].SourceFindingID != "current" {
		t.Fatalf("prior-run findings were attributed to the current review: %+v", observations)
	}
}

func TestReviewFindingCapReadsCanonicalConfiguration(t *testing.T) {
	stateDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(stateDir, "config.json"), []byte(`{"schemaVersion":1,"review":{"maxFindingsPerFeature":3}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cap, err := reviewFindingCap(stateDir)
	if err != nil {
		t.Fatal(err)
	}
	if cap != 3 {
		t.Fatalf("configured finding cap was not retained: %d", cap)
	}
}

func reviewTestFeature(id, path string) domain.Feature {
	return domain.Feature{
		SchemaVersion: 1,
		FeatureID:     id,
		Kind:          "service",
		Source:        "heuristic",
		Confidence:    domain.ConfidenceHigh,
		OwnedFiles:    []domain.FeatureFileRef{{Path: path, Reason: "test ownership"}},
		RawOriginal:   []byte(`{"schemaVersion":1,"featureId":"` + id + `","kind":"service","source":"heuristic"}`),
	}
}

func reviewInt(value int) *int { return &value }

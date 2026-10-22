package analyzers

import (
	"path/filepath"
	"testing"
)

func TestGitleaksCompletionRequiresExactSingleFileBytes(t *testing.T) {
	inputs := map[string]sourceInput{"src/file.py": {manifest: ManifestInput{Path: "src/file.py", Bytes: 112}}}
	for _, scenario := range []struct {
		name     string
		stderr   string
		complete bool
	}{
		{"matched", "INF scanned ~112 bytes (112 bytes) in 1ms", true},
		{"skipped bytes", "INF scanned ~0 bytes (0 bytes) in 1ms", false},
		{"missing progress", "WRN leaks found: 1", false},
		{"ambiguous progress", "INF scanned ~112 bytes (112 bytes)\nINF scanned ~112 bytes (112 bytes)", false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			parsed := parsedOutput{success: true, processed: make(map[string]struct{})}
			attestGitleaksInput(&parsed, inputs, []byte(scenario.stderr))
			if parsed.success != scenario.complete {
				t.Fatalf("complete=%v, want %v", parsed.success, scenario.complete)
			}
			if scenario.complete && len(parsed.processed) != 1 {
				t.Fatal("accepted input was not accounted")
			}
		})
	}
}

func TestGitleaksAbsoluteSARIFPathMustRemainInsideRoot(t *testing.T) {
	root := t.TempDir()
	inside := filepath.Join(root, "src", "file.py")
	path, err := sarifPath(sarifArtifactLocation{URI: inside}, root)
	if err != nil || path != "src/file.py" {
		t.Fatalf("expected contained scanner path, got %q: %v", path, err)
	}
	if _, err := sarifPath(sarifArtifactLocation{URI: filepath.Join(root, "..", "outside.py")}, root); err == nil {
		t.Fatal("SARIF evidence outside the selected repository was accepted")
	}
}

package state

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func adr012GateFiles(t *testing.T, schema string, shadow bool) map[string][]byte {
	t.Helper()
	files := map[string][]byte{ExportIssues: nil, ExportDispositions: nil, ExportWouldBlock: nil, ExportCoverageSummary: []byte("[]\n")}
	if shadow {
		files["state/v1/pr-runs.jsonl"] = nil
		files["state/v1/would-block-triage.jsonl"] = nil
	}
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	entries := []map[string]any{}
	for _, path := range paths {
		sum := sha256.Sum256(files[path])
		count := 0
		if strings.HasSuffix(path, ".json") {
			count = 1
		}
		entries = append(entries, map[string]any{"path": path, "records": count, "sha256": hex.EncodeToString(sum[:])})
	}
	var err error
	files[ExportManifest], err = CanonicalLine(map[string]any{"schema": schema, "files": entries, "wouldBlock": "present", "repositoryId": "repository", "uninspected": []any{}})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestADR012S2ValidateStateGate(t *testing.T) {
	t.Run("genuine_v1_and_v2_validate_without_mutation", func(t *testing.T) {
		for _, schema := range []string{"scout.state/v1", "scout.state/v2"} {
			files := adr012GateFiles(t, schema, schema == "scout.state/v2")
			before, _ := json.Marshal(files)
			meta, err := ValidateState(files)
			if err != nil || meta.RepositoryID != "repository" {
				t.Errorf("valid %s rejected: %#v %v", schema, meta, err)
			}
			after, _ := json.Marshal(files)
			if !reflect.DeepEqual(before, after) {
				t.Error("validation mutated wire bytes")
			}
		}
	})
	for _, tc := range []struct {
		name, schema string
		shadow       bool
		remedies     []string
	}{
		{"downgraded_hybrid", "scout.state/v1", true, []string{"republish the last intact v2 commit", "reapply subsequent state"}},
		{"newer_schema", "scout.state/v3", true, []string{"upgrade the writer", "never overwrite it with an older snapshot"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := adr012GateFiles(t, tc.schema, tc.shadow)
			_, err := ValidateState(files)
			if err == nil {
				t.Fatal("incompatible state accepted")
			}
			for _, remedy := range tc.remedies {
				if !strings.Contains(err.Error(), remedy) {
					t.Errorf("missing exact remedy %q in %q", remedy, err)
				}
			}
			before, _ := json.Marshal(files)
			_, _ = ValidateState(files)
			after, _ := json.Marshal(files)
			if !reflect.DeepEqual(before, after) {
				t.Error("compatibility refusal mutated state")
			}
		})
	}
	t.Run("missing_manifest_is_invalid", func(t *testing.T) {
		invalid := map[string][]byte{ExportIssues: []byte("{}\n")}
		if _, err := ValidateState(invalid); err == nil {
			t.Error("manifest-free state accepted")
		}
	})
}

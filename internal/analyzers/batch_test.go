package analyzers

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"scout/internal/domain"
)

func batchFixture(t *testing.T) (string, []domain.File) {
	t.Helper()
	root := t.TempDir()
	sources := map[string]string{
		"a_unsafe.py": "import yaml\n\ndef load(raw):\n    return yaml.unsafe_load(raw)\n",
		"b_clean.py":  "def add(a, b):\n    return a + b\n",
		"c_unsafe.py": "import yaml\n\ndef other(raw):\n    return yaml.unsafe_load(raw)\n",
	}
	var files []domain.File
	for _, name := range []string{"a_unsafe.py", "b_clean.py", "c_unsafe.py"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(sources[name]), 0o600); err != nil {
			t.Fatal(err)
		}
		files = append(files, domain.File{Path: name, Language: "python", Eligible: true})
	}
	return root, files
}

func TestBatchRunMatchesSingleFileRuns(t *testing.T) {
	if _, err := exec.LookPath("semgrep"); err != nil {
		t.Skip("semgrep is not installed")
	}
	root, files := batchFixture(t)
	spec := specForTest(t, "security.semgrep.python-unsafe-yaml-load")
	runner, err := NewRunner(spec, "run-batch")
	if err != nil {
		t.Fatal(err)
	}
	batch, batchFound, _ := runner.Run(context.Background(), root, files)
	single := map[string]domain.Coverage{}
	singleCount := map[string]int{}
	for _, file := range files {
		adapter, found, _ := runner.Run(context.Background(), root, []domain.File{file})
		if len(adapter.Coverage) != 1 {
			t.Fatalf("single run coverage = %d rows", len(adapter.Coverage))
		}
		single[file.Path] = adapter.Coverage[0]
		singleCount[file.Path] = len(found)
	}
	if single["b_clean.py"].Status != domain.CoverageInspected {
		t.Skipf("pinned semgrep is unavailable: %s", single["b_clean.py"].Reason)
	}
	if len(batch.Coverage) != len(files) {
		t.Fatalf("batch coverage = %d rows, want %d", len(batch.Coverage), len(files))
	}
	batchCount := map[string]int{}
	for _, observation := range batchFound {
		batchCount[observation.Evidence[0].Location.Path]++
	}
	for _, row := range batch.Coverage {
		want := single[row.Key.FilePath]
		if row.Status != want.Status || row.Reason != want.Reason {
			t.Errorf("%s batch = %s/%q, single = %s/%q", row.Key.FilePath, row.Status, row.Reason, want.Status, want.Reason)
		}
		if batchCount[row.Key.FilePath] != singleCount[row.Key.FilePath] {
			t.Errorf("%s batch observations = %d, single = %d", row.Key.FilePath, batchCount[row.Key.FilePath], singleCount[row.Key.FilePath])
		}
	}
	if singleCount["a_unsafe.py"] == 0 || singleCount["b_clean.py"] != 0 {
		t.Errorf("fixture did not exercise findings and clean files: %v", singleCount)
	}
}

func TestFailingBatchMarksEveryFileNotInspected(t *testing.T) {
	root, files := batchFixture(t)
	spec := securitySpecs()[0]
	spec.Binary.Path = "scout-nonexistent-scanner"
	runner, err := NewRunner(spec, "run-batch-fail")
	if err != nil {
		t.Fatal(err)
	}
	adapter, found, runErr := runner.Run(context.Background(), root, files)
	if runErr == nil || len(found) != 0 || len(adapter.Coverage) != len(files) {
		t.Fatalf("failing batch = err %v, %d observations, %d rows", runErr, len(found), len(adapter.Coverage))
	}
	for _, row := range adapter.Coverage {
		if row.Status == domain.CoverageInspected || row.Reason == "" {
			t.Errorf("%s = %s/%q, want non-inspected with reason", row.Key.FilePath, row.Status, row.Reason)
		}
	}
}

func TestBatchesRespectDeclaredLimits(t *testing.T) {
	root, files := batchFixture(t)
	spec := securitySpecs()[0]
	spec.Inputs.MaxFiles = 2
	batches := spec.Batches(root, files)
	if len(batches) != 2 || len(batches[0]) != 2 || len(batches[1]) != 1 || batches[0][0].Path != "a_unsafe.py" || batches[1][0].Path != "c_unsafe.py" {
		t.Fatalf("batches = %v", batches)
	}
}

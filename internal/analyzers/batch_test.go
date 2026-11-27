//go:build boundary

package analyzers

import (
	"context"
	"os/exec"
	"testing"

	"scout/internal/domain"
)

func TestBatchRunMatchesSingleFileRuns(t *testing.T) {
	if _, err := exec.LookPath("semgrep"); err != nil {
		t.Fatalf("semgrep is not installed: %v", err)
	}
	root, files := batchFixture(t)
	spec := specForTest(t, "security.semgrep.python-unsafe-yaml-load")
	runner, err := NewRunner(spec, "run-batch")
	if err != nil {
		t.Fatal(err)
	}
	batch, batchFound, _ := runner.Run(context.Background(), root, files)
	single, singleCount := singleFileRuns(t, runner, root, files)
	if single["b_clean.py"].Status != domain.CoverageInspected {
		t.Fatalf("pinned semgrep is unavailable: %s", single["b_clean.py"].Reason)
	}
	if len(batch.Coverage) != len(files) {
		t.Fatalf("batch coverage = %d rows, want %d", len(batch.Coverage), len(files))
	}
	assertBatchRowsMatchSingleRuns(t, batch.Coverage, observationsByPath(batchFound), single, singleCount)
	if singleCount["a_unsafe.py"] == 0 || singleCount["b_clean.py"] != 0 {
		t.Errorf("fixture did not exercise findings and clean files: %v", singleCount)
	}
}

func assertBatchRowsMatchSingleRuns(t *testing.T, rows []domain.Coverage, batchCount map[string]int, single map[string]domain.Coverage, singleCount map[string]int) {
	t.Helper()
	for _, row := range rows {
		want := single[row.Key.FilePath]
		if row.Status != want.Status || row.Reason != want.Reason {
			t.Errorf("%s batch = %s/%q, single = %s/%q", row.Key.FilePath, row.Status, row.Reason, want.Status, want.Reason)
		}
		if batchCount[row.Key.FilePath] != singleCount[row.Key.FilePath] {
			t.Errorf("%s batch observations = %d, single = %d", row.Key.FilePath, batchCount[row.Key.FilePath], singleCount[row.Key.FilePath])
		}
	}
}

func singleFileRuns(t *testing.T, runner *Runner, root string, files []domain.File) (map[string]domain.Coverage, map[string]int) {
	t.Helper()
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
	return single, singleCount
}

func observationsByPath(observations []domain.Observation) map[string]int {
	counts := map[string]int{}
	for _, observation := range observations {
		counts[observation.Evidence[0].Location.Path]++
	}
	return counts
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

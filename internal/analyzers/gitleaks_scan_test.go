//go:build boundary

package analyzers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"scout/internal/domain"
)

type gitleaksOutcome struct {
	status      domain.CoverageStatus
	observation []string
}

func gitleaksOutcomes(t *testing.T, result domain.AdapterResult, observations []domain.Observation) map[string]gitleaksOutcome {
	t.Helper()
	outcomes := make(map[string]gitleaksOutcome)
	for _, row := range result.Coverage {
		outcomes[row.Key.FilePath] = gitleaksOutcome{status: row.Status}
	}
	for _, observation := range observations {
		path := observation.Evidence[0].Location.Path
		outcome := outcomes[path]
		outcome.observation = append(outcome.observation, fmt.Sprintf("%s|%s|%d|%s", observation.RuleID, observation.Evidence[0].Location.Path, observation.Evidence[0].Location.StartLine, observation.SourceFindingID))
		sort.Strings(outcome.observation)
		outcomes[path] = outcome
	}
	return outcomes
}

func gitleaksFixture(t *testing.T) (string, []domain.File) {
	t.Helper()
	secret := "AKIA" + "IOSFODNN7EXAMPLQ"
	root := t.TempDir()
	for path, content := range map[string]string{
		"a.py":                 "x = 1\nAWS_KEY = \"" + secret + "\"\n",
		"clean.py":             "print('hello')\n",
		"sub/dir/c.py":         "# moved\n\n\nTOKEN_ID = \"" + secret + "\"\n",
		"sub/other/clean.json": "{\"a\": 1}\n",
		"empty.py":             "",
	} {
		full := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root, []domain.File{
		{Path: "a.py", Language: "python", Eligible: true},
		{Path: "clean.py", Language: "python", Eligible: true},
		{Path: "empty.py", Language: "python", Eligible: true},
		{Path: "sub/dir/c.py", Language: "python", Eligible: true},
		{Path: "sub/other/clean.json", Language: "json", Eligible: true},
	}
}

func TestGitleaksBatchProvesEveryFileAndMatchesSingleFileRuns(t *testing.T) {
	runner := gitleaksRunner(t)
	root, files := gitleaksFixture(t)
	ctx := context.Background()
	rawBatchWithoutRetry, observations, err := runner.runBatch(ctx, root, files)
	if err != nil || rawBatchWithoutRetry.Status != domain.AdapterCompleted {
		t.Fatalf("batch status %s error %v failures %v", rawBatchWithoutRetry.Status, err, rawBatchWithoutRetry.Failures)
	}
	batch := gitleaksOutcomes(t, rawBatchWithoutRetry, observations)
	if len(observations) != 2 {
		t.Fatalf("batch observations = %d, want 2", len(observations))
	}
	assertBatchMatchesSingleFileRuns(t, runner, root, files, batch)
	assertSharedSecretKeepsOneDigest(t, batch)
}

func assertBatchMatchesSingleFileRuns(t *testing.T, runner *Runner, root string, files []domain.File, batch map[string]gitleaksOutcome) {
	t.Helper()
	for _, file := range files {
		single, singleObservations, singleErr := runner.runBatch(context.Background(), root, []domain.File{file})
		if singleErr != nil || single.Status != domain.AdapterCompleted {
			t.Fatalf("%s single status %s error %v", file.Path, single.Status, singleErr)
		}
		want := gitleaksOutcomes(t, single, singleObservations)[file.Path]
		got := batch[file.Path]
		if got.status != domain.CoverageInspected || got.status != want.status || fmt.Sprint(got.observation) != fmt.Sprint(want.observation) {
			t.Fatalf("%s: batch %+v differs from single %+v", file.Path, got, want)
		}
	}
}

func assertSharedSecretKeepsOneDigest(t *testing.T, batch map[string]gitleaksOutcome) {
	t.Helper()
	if a, c := batch["a.py"].observation, batch["sub/dir/c.py"].observation; len(a) != 1 || len(c) != 1 || strings.Split(a[0], "|")[3] != strings.Split(c[0], "|")[3] || strings.Split(a[0], "|")[3] == "" {
		t.Fatalf("same secret in two files must keep one digest: %v %v", a, c)
	}
}

func TestGitleaksLockfileExclusionKeepsAdapterCompleteAndProvesSource(t *testing.T) {
	runner := gitleaksRunner(t)
	root, files := gitleaksFixture(t)
	if err := os.WriteFile(filepath.Join(root, "package-lock.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	files = append(files, domain.File{Path: "package-lock.json", Language: "json", Eligible: true})
	result, observations, err := runner.Run(context.Background(), root, files)
	if err != nil || result.Status != domain.AdapterCompleted {
		t.Fatalf("status %s error %v failures %v", result.Status, err, result.Failures)
	}
	outcomes := gitleaksOutcomes(t, result, observations)
	if outcomes["package-lock.json"].status != domain.CoverageNotApplicable || outcomes["a.py"].status != domain.CoverageInspected || len(observations) != 2 {
		t.Fatalf("unexpected outcomes %+v (%d observations)", outcomes, len(observations))
	}
}

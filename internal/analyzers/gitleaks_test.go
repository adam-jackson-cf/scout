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

func TestGitleaksBatchProofRequiresEveryFileAndExactBytes(t *testing.T) {
	stage := "/state/stage-1"
	inputs := map[string]sourceInput{
		"a.py":     {manifest: ManifestInput{Path: "a.py", Bytes: 100}},
		"sub/b.py": {manifest: ManifestInput{Path: "sub/b.py", Bytes: 12}},
		"empty.py": {manifest: ManifestInput{Path: "empty.py", Bytes: 0}},
	}
	scan := func(path string) string { return "10:36AM TRC scanning path path=" + stage + "/" + path + "\n" }
	emptySkip := "10:36AM DBG skipping empty file path=" + stage + "/empty.py\n"
	for _, scenario := range []struct {
		name   string
		stderr string
		proven []string
	}{
		{"every file matched", scan("a.py") + scan("sub/b.py") + emptySkip + "10:36AM INF scanned ~112 bytes (112 bytes) in 1ms\n", []string{"a.py", "empty.py", "sub/b.py"}},
		{"quoted path", strings.Replace(scan("a.py"), stage+"/a.py", `"`+stage+`/a.py"`, 1) + scan("sub/b.py") + emptySkip + "INF scanned ~112 bytes (112 bytes)\n", []string{"a.py", "empty.py", "sub/b.py"}},
		{"skipped file leaves total short", scan("a.py") + emptySkip + "INF scanned ~100 bytes (100 bytes)\n", []string{"a.py", "empty.py"}},
		{"total disagrees with traced files", scan("a.py") + scan("sub/b.py") + emptySkip + "INF scanned ~100 bytes (100 bytes)\n", nil},
		{"repeated file event", scan("a.py") + scan("a.py") + scan("sub/b.py") + emptySkip + "INF scanned ~212 bytes (212 bytes)\n", nil},
		{"undeclared path", scan("a.py") + scan("other.py") + scan("sub/b.py") + emptySkip + "INF scanned ~112 bytes (112 bytes)\n", nil},
		{"path outside stage", scan("a.py") + "10:36AM TRC scanning path path=/etc/passwd\n" + "INF scanned ~100 bytes (100 bytes)\n", nil},
		{"missing total", scan("a.py") + scan("sub/b.py") + emptySkip + "WRN leaks found: 1\n", nil},
		{"ambiguous total", scan("a.py") + scan("sub/b.py") + emptySkip + "INF scanned ~112 bytes (112 bytes)\nINF scanned ~112 bytes (112 bytes)\n", nil},
		{"events without total", scan("a.py") + scan("sub/b.py"), nil},
		{"forged line inside other output", "DBG segment found: x TRC scanning path path=" + stage + "/a.py\n" + scan("sub/b.py") + emptySkip + "INF scanned ~12 bytes (12 bytes)\n", []string{"empty.py", "sub/b.py"}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			parsed := parsedOutput{success: true, processed: make(map[string]struct{})}
			attestGitleaksInput(&parsed, inputs, []byte(scenario.stderr), stage)
			var got []string
			for path := range parsed.processed {
				got = append(got, path)
			}
			sort.Strings(got)
			if fmt.Sprint(got) != fmt.Sprint(scenario.proven) {
				t.Fatalf("proven = %v, want %v", got, scenario.proven)
			}
		})
	}
}

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

func gitleaksRunner(t *testing.T) *Runner {
	t.Helper()
	for _, spec := range securitySpecs() {
		if spec.ID == "security.gitleaks.default-credentials" {
			runner, err := NewRunner(spec, "run-batch")
			if err != nil {
				t.Fatal(err)
			}
			return runner
		}
	}
	t.Fatal("gitleaks spec is missing")
	return nil
}

func TestGitleaksBatchProvesEveryFileAndMatchesSingleFileRuns(t *testing.T) {
	runner := gitleaksRunner(t)
	root, files := gitleaksFixture(t)
	ctx := context.Background()
	// runBatch is the raw batch: no fallback may be needed to prove these files.
	result, observations, err := runner.runBatch(ctx, root, files)
	if result.Status == domain.AdapterMissingTool {
		t.Skip("gitleaks or sandbox is not installed")
	}
	if err != nil || result.Status != domain.AdapterCompleted {
		t.Fatalf("batch status %s error %v failures %v", result.Status, err, result.Failures)
	}
	batch := gitleaksOutcomes(t, result, observations)
	if len(observations) != 2 {
		t.Fatalf("batch observations = %d, want 2", len(observations))
	}
	for _, file := range files {
		single, singleObservations, singleErr := runner.runBatch(ctx, root, []domain.File{file})
		if singleErr != nil || single.Status != domain.AdapterCompleted {
			t.Fatalf("%s single status %s error %v", file.Path, single.Status, singleErr)
		}
		want := gitleaksOutcomes(t, single, singleObservations)[file.Path]
		got := batch[file.Path]
		if got.status != domain.CoverageInspected || got.status != want.status || fmt.Sprint(got.observation) != fmt.Sprint(want.observation) {
			t.Fatalf("%s: batch %+v differs from single %+v", file.Path, got, want)
		}
	}
	if a, c := batch["a.py"].observation, batch["sub/dir/c.py"].observation; len(a) != 1 || len(c) != 1 || strings.Split(a[0], "|")[3] != strings.Split(c[0], "|")[3] || strings.Split(a[0], "|")[3] == "" {
		t.Fatalf("same secret in two files must keep one digest: %v %v", a, c)
	}
}

func TestGitleaksGlobalAllowlistPathsAreNotApplicableAndRuleAllowlistsStillScan(t *testing.T) {
	runner := gitleaksRunner(t)
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	paths := []string{"toolchain/node/package-lock.json", "web/yarn.lock", "assets/logo.PNG", "gitleaks.toml", "src/app.py", "node_modules/pkg/data.json"}
	var files []domain.File
	for _, path := range paths {
		full := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("x = 1\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		files = append(files, domain.File{Path: path, Language: "json", Eligible: true})
	}
	accepted, coverage, _, eligible, inputErr := runner.selectInputs(root, files, "result")
	if inputErr {
		t.Fatal("unexpected input error")
	}
	want := map[string]bool{"toolchain/node/package-lock.json": true, "web/yarn.lock": true, "assets/logo.PNG": true, "gitleaks.toml": true, "node_modules/pkg/data.json": true}
	for _, row := range coverage {
		excluded := row.Status == domain.CoverageNotApplicable && row.Reason == gitleaksAllowlistReason
		if excluded != want[row.Key.FilePath] {
			t.Fatalf("%s: status %s reason %q, want excluded=%v", row.Key.FilePath, row.Status, row.Reason, want[row.Key.FilePath])
		}
	}
	if eligible != 1 || len(accepted) != 1 || accepted[0].manifest.Path != "src/app.py" {
		t.Fatalf("eligible %d accepted %d, want only src/app.py", eligible, len(accepted))
	}
	// Rule-level [[rules.allowlists]] paths only suppress one rule's findings:
	// they are not part of the global allowlist.
	pack, err := bundledRule(gitleaksDefaultRule)
	if err != nil {
		t.Fatal(err)
	}
	patterns, err := gitleaksGlobalPathAllowlist(pack)
	if err != nil {
		t.Fatal(err)
	}
	if pathAllowlisted(patterns, "docs/@octokit/auth-token/README.md") || pathAllowlisted(patterns, "recipes/app.bb") {
		t.Fatal("rule-level allowlist path must not be in the global allowlist")
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
	if result.Status == domain.AdapterMissingTool {
		t.Skip("gitleaks or sandbox is not installed")
	}
	if err != nil || result.Status != domain.AdapterCompleted {
		t.Fatalf("status %s error %v failures %v", result.Status, err, result.Failures)
	}
	outcomes := gitleaksOutcomes(t, result, observations)
	if outcomes["package-lock.json"].status != domain.CoverageNotApplicable || outcomes["a.py"].status != domain.CoverageInspected || len(observations) != 2 {
		t.Fatalf("unexpected outcomes %+v (%d observations)", outcomes, len(observations))
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

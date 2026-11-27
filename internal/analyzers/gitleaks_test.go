package analyzers

import (
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

func TestGitleaksGlobalAllowlistPathsAreNotApplicableAndRuleAllowlistsStillScan(t *testing.T) {
	runner := gitleaksRunner(t)
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	files := writeAllowlistFixture(t, root, []string{"toolchain/node/package-lock.json", "web/yarn.lock", "assets/logo.PNG", "gitleaks.toml", "src/app.py", "node_modules/pkg/data.json"})
	selection := runner.selectInputs(root, files, "result")
	if selection.inputErr {
		t.Fatal("unexpected input error")
	}
	assertAllowlistExclusions(t, selection.coverage, map[string]bool{"toolchain/node/package-lock.json": true, "web/yarn.lock": true, "assets/logo.PNG": true, "gitleaks.toml": true, "node_modules/pkg/data.json": true})
	if selection.eligibleCount != 1 || len(selection.accepted) != 1 || selection.accepted[0].manifest.Path != "src/app.py" {
		t.Fatalf("eligible %d accepted %d, want only src/app.py", selection.eligibleCount, len(selection.accepted))
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

func writeAllowlistFixture(t *testing.T, root string, paths []string) []domain.File {
	t.Helper()
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
	return files
}

func assertAllowlistExclusions(t *testing.T, coverage []domain.Coverage, want map[string]bool) {
	t.Helper()
	for _, row := range coverage {
		excluded := row.Status == domain.CoverageNotApplicable && row.Reason == gitleaksAllowlistReason
		if excluded != want[row.Key.FilePath] {
			t.Fatalf("%s: status %s reason %q, want excluded=%v", row.Key.FilePath, row.Status, row.Reason, want[row.Key.FilePath])
		}
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

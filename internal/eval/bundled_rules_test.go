//go:build boundary

package eval

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type semgrepFixtureResult struct {
	Results []struct {
		CheckID string `json:"check_id"`
		Path    string `json:"path"`
	} `json:"results"`
	Errors []json.RawMessage `json:"errors"`
}

func writeFixture(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestBundledRulesStillDetectCorpusMechanisms runs the current bundled rules
// behind the corpus's true issues (mutable Actions refs, python-multipart
// CVE-2024-24762, currentScript DOM clobbering, CSS height transition) on
// fixtures reproducing each mechanism and its negative boundary.
func TestBundledRulesStillDetectCorpusMechanisms(t *testing.T) {
	semgrep, err := exec.LookPath("semgrep")
	if err != nil {
		t.Fatalf("semgrep is required to check the bundled rules (run `source .tools/activate`): %v", err)
	}
	rules := filepath.Join(repoRoot(t), "internal", "analyzers", "rules")
	dir := t.TempDir()
	writeFixture(t, dir, ".github/workflows/mutable.yml", "jobs:\n  a:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n")
	writeFixture(t, dir, ".github/workflows/pinned.yml", "jobs:\n  a:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@0123456789abcdef0123456789abcdef01234567\n")
	writeFixture(t, dir, "vuln/pyproject.toml", "[project]\ndependencies = [\"python-multipart<0.0.7\"]\n")
	writeFixture(t, dir, "fixed/pyproject.toml", "[project]\ndependencies = [\"python-multipart>=0.0.7\"]\n")
	writeFixture(t, dir, "css/identity.css", ".menu {\n  transition: height 0.3s ease;\n}\n")
	writeFixture(t, dir, "css/safe.css", ".menu {\n  transition: max-height 0.3s ease;\n}\n")
	writeFixture(t, dir, "runtime/AutoPublicPathRuntimeModule.js", "const x = Template.indent([\n\t`if (document.currentScript)`,\n\tTemplate.indent(`scriptUrl = document.currentScript.src;`)\n]);\n")
	writeFixture(t, dir, "runtime/guarded.js", "const x = Template.indent([\n\t`if (document.currentScript && document.currentScript.tagName.toUpperCase() === 'SCRIPT')`,\n\tTemplate.indent(`scriptUrl = document.currentScript.src;`)\n]);\n")

	got := runBundledRules(t, semgrep, rules, dir)
	want := map[string]string{
		".github/workflows/mutable.yml":          "scout.security.actions.mutable-ref",
		"vuln/pyproject.toml":                    "scout.security.python-multipart-cve-2024-24762",
		"css/identity.css":                       "scout.performance.css.layout-height-transition",
		"runtime/AutoPublicPathRuntimeModule.js": "scout.security.javascript.current-script-dom-clobbering",
	}
	for path, rule := range want {
		if got[path] != rule {
			t.Errorf("%s: want %s, got %q", path, rule, got[path])
		}
	}
	for path, rule := range got {
		if want[path] == "" {
			t.Errorf("negative-boundary fixture %s wrongly flagged by %s", path, rule)
		}
	}
}

func runBundledRules(t *testing.T, semgrep, rules, dir string) map[string]string {
	t.Helper()
	args := []string{"--metrics=off", "--disable-version-check", "--json", "--quiet"}
	for _, rule := range []string{
		"security/semgrep-actions-mutable-ref.yaml",
		"security/semgrep-python-multipart-cve-2024-24762.yaml",
		"security/semgrep-javascript-current-script-dom-clobbering.yaml",
		"performance/semgrep-css-height-transition.yaml",
	} {
		args = append(args, "--config", filepath.Join(rules, rule))
	}
	args = append(args, ".")
	cmd := exec.CommandContext(boundedTestContext(t), semgrep, args...)
	cmd.Dir = dir
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		t.Fatalf("semgrep: %v", err)
	}
	var out semgrepFixtureResult
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Errors) != 0 {
		t.Fatalf("semgrep errors: %s", out.Errors)
	}
	got := map[string]string{}
	for _, r := range out.Results {
		got[filepath.ToSlash(r.Path)] = r.CheckID[strings.LastIndex(r.CheckID, "scout."):]
	}
	return got
}

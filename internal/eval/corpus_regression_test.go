package eval

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// The v1 corpus regression (design §7.3 Track A, §9.8; REQ-19): the 9
// adjudicated true issues stay detected and the 13 adjudicated false positives
// stay unflagged.
//
// Two checks, because the corpus targets are pinned remote repositories that
// tests cannot clone (no network):
//  1. Replay: the recorded Scout observations of the latest corpus run
//     (evaluation/scout-requests-v2.json) are compared with reference-v1.json
//     through CompareOutcomes.
//  2. Live rules: the bundled Semgrep rules behind the 9 true issues are run on
//     minimal fixtures reproducing each mechanism, plus the negative boundary.

const (
	wantValidatedTrue  = 9
	wantFalsePositives = 13
)

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestCorpusReplayKeepsTrueIssuesAndDoesNotFlagFalsePositives(t *testing.T) {
	root := repoRoot(t)
	loaded, err := LoadCorpus(filepath.Join(root, "evaluation", "reference-v1.json"))
	if err != nil || loaded.Status != CorpusAvailable {
		t.Fatalf("load reference corpus: %s %v", loaded.Status, err)
	}
	data, err := os.ReadFile(filepath.Join(root, "evaluation", "scout-requests-v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	var requests []EvaluationRequest
	if err := json.Unmarshal(data, &requests); err != nil {
		t.Fatal(err)
	}

	wantTrue := map[string]bool{}
	falsePositives := map[string][]ReferenceIssue{} // corpus entry id -> FP issues
	for _, entry := range loaded.Corpus.Entries {
		for _, issue := range entry.Issues {
			switch issue.Adjudication {
			case AdjudicationValidatedTrue:
				wantTrue[issue.ID] = true
			case AdjudicationFalsePositive:
				falsePositives[entry.ID] = append(falsePositives[entry.ID], issue)
			}
		}
	}
	nFP := 0
	for _, issues := range falsePositives {
		nFP += len(issues)
	}
	if len(wantTrue) != wantValidatedTrue || nFP != wantFalsePositives {
		t.Fatalf("reference corpus has %d validated true / %d false positives, want %d / %d", len(wantTrue), nFP, wantValidatedTrue, wantFalsePositives)
	}

	matched := map[string]bool{}
	flagged := []string{}
	for _, request := range requests {
		report, err := CompareOutcomes(loaded, request)
		if err != nil {
			t.Fatalf("compare %s/%s/%s: %v", request.Target.Repository, request.Language, request.Family, err)
		}
		for _, match := range report.Matches {
			matched[match.ReferenceIssueID] = true
		}
		for _, miss := range report.Misses {
			t.Errorf("true issue %s lost: %s %s", miss.ID, miss.Mechanism, miss.Path)
		}
		for _, issue := range falsePositives[report.CorpusEntryID] {
			for _, observation := range request.Observations {
				if pathsRelevant(issue.Path, observation.Path) {
					flagged = append(flagged, issue.ID+" <- "+observation.Source+" "+observation.Path)
				}
			}
		}
	}
	for id := range wantTrue {
		if !matched[id] {
			t.Errorf("true issue %s not detected", id)
		}
	}
	sort.Strings(flagged)
	if len(flagged) != 0 {
		t.Errorf("false positives flagged:\n%s", strings.Join(flagged, "\n"))
	}
	t.Logf("corpus replay: %d/%d true detected, %d/%d false positives unflagged", len(matched), len(wantTrue), nFP-len(flagged), nFP)
}

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
	cmd := exec.Command(semgrep, args...)
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

// Package toolchain holds the parity checks between toolchain/versions.env (what
// scripts/bootstrap-tools.sh installs) and the versions Scout's Go code demands at
// run time. A drift makes every affected lane `incomplete` on every machine.
package toolchain

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"scout/internal/analyzers"
	"scout/internal/clawpatch"
)

const repoRoot = "../.."

func manifest(t *testing.T) map[string]string {
	t.Helper()
	file, err := os.Open(filepath.Join(repoRoot, "toolchain", "versions.env"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	values := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || value == "" {
			t.Fatalf("malformed versions.env line %q", line)
		}
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return values
}

// Every scanner binary a lane can run must be pinned in the manifest at the exact
// version the lane verifies; a tool without a manifest entry is never provisioned.
func TestAnalyzerBinariesMatchManifest(t *testing.T) {
	pins := manifest(t)
	keyFor := map[string]string{"semgrep": "SEMGREP_VERSION", "ruff": "RUFF_VERSION", "gitleaks": "GITLEAKS_VERSION", "lizard": "LIZARD_VERSION", "jscpd": "JSCPD_VERSION"}
	registry, err := analyzers.BuiltinRegistry()
	if err != nil {
		t.Fatal(err)
	}
	for _, spec := range registry.Specs() {
		key, known := keyFor[spec.Binary.Path]
		if !known {
			t.Errorf("adapter %s runs %q, which toolchain/versions.env does not provision", spec.ID, spec.Binary.Path)
			continue
		}
		if spec.Binary.Version != pins[key] {
			t.Errorf("adapter %s requires %s %s but toolchain/versions.env pins %s=%s", spec.ID, spec.Binary.Path, spec.Binary.Version, key, pins[key])
		}
	}
}

func TestClawpatchPinMatchesManifestAndSubmodule(t *testing.T) {
	if clawpatch.Version != manifest(t)["CLAWPATCH_VERSION"] {
		t.Errorf("clawpatch.Version %s != CLAWPATCH_VERSION", clawpatch.Version)
	}
	output, err := exec.Command("git", "-C", repoRoot, "ls-files", "-s", "third_party/clawpatch").Output()
	if err != nil || len(output) == 0 {
		t.Fatalf("the third_party/clawpatch submodule gitlink cannot be read from a git checkout (the pin check cannot be skipped): %v", err)
	}
	fields := strings.Fields(string(output))
	if len(fields) < 2 || fields[0] != "160000" {
		t.Fatalf("third_party/clawpatch is not a submodule gitlink: %q", output)
	}
	if fields[1] != clawpatch.Revision {
		t.Errorf("submodule gitlink %s != clawpatch.Revision %s (an upgrade needs a compatibility review)", fields[1], clawpatch.Revision)
	}
}

func TestGoToolchainMatchesManifest(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	match := regexp.MustCompile(`(?m)^toolchain go(\S+)$`).FindSubmatch(data)
	if match == nil || string(match[1]) != manifest(t)["GO_VERSION"] {
		t.Errorf("go.mod toolchain %q != GO_VERSION %s", match, manifest(t)["GO_VERSION"])
	}
}

// The lock files are what bootstrap installs; their pins must equal the manifest.
func TestLockFilesMatchManifest(t *testing.T) {
	pins := manifest(t)
	lock, err := os.ReadFile(filepath.Join(repoRoot, "toolchain", "python-requirements.lock"))
	if err != nil {
		t.Fatal(err)
	}
	for name, key := range map[string]string{"semgrep": "SEMGREP_VERSION", "ruff": "RUFF_VERSION", "lizard": "LIZARD_VERSION"} {
		if !regexp.MustCompile(`(?m)^` + name + `==` + regexp.QuoteMeta(pins[key]) + ` `).Match(lock) {
			t.Errorf("python-requirements.lock does not pin %s==%s", name, pins[key])
		}
	}
	data, err := os.ReadFile(filepath.Join(repoRoot, "toolchain", "node", "package-lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	var nodeLock struct {
		Packages map[string]struct {
			Version string `json:"version"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(data, &nodeLock); err != nil {
		t.Fatal(err)
	}
	for name, key := range map[string]string{"jscpd": "JSCPD_VERSION", "pnpm": "PNPM_VERSION"} {
		if got := nodeLock.Packages["node_modules/"+name].Version; got != pins[key] {
			t.Errorf("package-lock.json pins %s %q, versions.env %s=%s", name, got, key, pins[key])
		}
	}
}

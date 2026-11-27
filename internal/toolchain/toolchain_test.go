// Package toolchain holds the parity checks between toolchain/versions.env (what
// scripts/bootstrap-tools.sh installs) and the versions Scout's Go code demands at
// run time. A drift makes every affected lane `incomplete` on every machine.
package toolchain

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"scout/internal/analyzers"
)

const repoRoot = "../.."

func manifest(t *testing.T) map[string]string {
	t.Helper()
	values, err := Manifest(repoRoot)
	if err != nil {
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

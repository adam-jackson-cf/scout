package toolchain

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

var cleanTree = map[string]string{
	ManifestPath:                         "GO_VERSION=1.26.11\nNODE_VERSION=22.18.0\nRUFF_VERSION=0.15.8\nPRETTIER_VERSION=3.6.2\n",
	BootstrapPath:                        "#!/usr/bin/env bash\nfor key in GO_VERSION NODE_VERSION RUFF_VERSION PRETTIER_VERSION; do\n  :\ndone\ncase \"${1:-}\" in\n  --relock) ;;\nesac\nGOTOOLCHAIN=\"go$GO_VERSION\" node \"$NODE_VERSION\"\n",
	"toolchain/python-requirements.in":   "ruff==0.15.8\n",
	"toolchain/python-requirements.lock": "ruff==0.15.8 \\\n    --hash=sha256:aa\n",
	"toolchain/node/package.json":        `{"packageManager": "npm@10.9.3", "dependencies": {"prettier": "3.6.2"}}`,
	"toolchain/node/package-lock.json":   `{"packages": {"": {"dependencies": {"prettier": "3.6.2"}}, "node_modules/prettier": {"version": "3.6.2"}}}`,
	"README.md": "Install: `mise use go@1.26.11`, then run `scripts/bootstrap-tools.sh --relock`.\n" +
		"`scripts/bootstrap-tools.sh` installs Node, Ruff and Prettier and selects the pinned Go toolchain.\n",
}

func writeTree(t *testing.T, files map[string]string, edits map[string]string) Tree {
	t.Helper()
	root := t.TempDir()
	merged := map[string]string{}
	for k, v := range files {
		merged[k] = v
	}
	for k, v := range edits {
		merged[k] = v
	}
	var names []string
	for rel, text := range merged {
		if text == "" {
			continue
		}
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		names = append(names, rel)
	}
	sort.Strings(names)
	return Tree{Root: root, Files: names}
}

func problems(t *testing.T, edits map[string]string) []string {
	t.Helper()
	got, err := Check(writeTree(t, cleanTree, edits))
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, p := range got {
		out = append(out, p.Rule+" "+p.File+" "+p.Problem)
	}
	return out
}

func TestCleanToolchainTreeHasNoParityProblems(t *testing.T) {
	if got := problems(t, nil); len(got) != 0 {
		t.Fatalf("clean tree reported %v", got)
	}
}

func TestParityRejectsEachDrift(t *testing.T) {
	cases := []struct {
		name  string
		edits map[string]string
		want  string
	}{
		{"python manifest differs from versions.env", map[string]string{"toolchain/python-requirements.in": "ruff==0.15.9\n", "toolchain/python-requirements.lock": "ruff==0.15.9 \\\n"}, "PARITY.lock-version toolchain/python-requirements.in ruff 0.15.9 differs"},
		{"python lock misses the manifest pin", map[string]string{"toolchain/python-requirements.lock": "ruff==0.15.7 \\\n"}, "PARITY.lock-version toolchain/python-requirements.lock does not lock ruff==0.15.8"},
		{"node lock differs from package.json", map[string]string{"toolchain/node/package-lock.json": `{"packages": {"": {"dependencies": {"prettier": "3.6.2"}}, "node_modules/prettier": {"version": "3.6.1"}}}`}, "PARITY.lock-version toolchain/node/package-lock.json locks prettier"},
		{"installed package without a pin", map[string]string{"toolchain/python-requirements.in": "ruff==0.15.8\nsetuptools==84.1.0\n", "toolchain/python-requirements.lock": "ruff==0.15.8 \\\nsetuptools==84.1.0 \\\n"}, "PARITY.dependency-inventory toolchain/python-requirements.in setuptools is installed"},
		{"pin missing from the required-key list", map[string]string{ManifestPath: cleanTree[ManifestPath] + "UV_VERSION=0.13.5\n"}, "PARITY.dependency-inventory toolchain/versions.env UV_VERSION is not in the required-key list"},
		{"pinned tool nothing installs", map[string]string{ManifestPath: cleanTree[ManifestPath] + "UV_VERSION=0.13.5\n", BootstrapPath: strings.Replace(cleanTree[BootstrapPath], "PRETTIER_VERSION; do", "PRETTIER_VERSION UV_VERSION; do", 1)}, "PARITY.dependency-inventory toolchain/versions.env UV_VERSION is pinned but nothing installs it"},
		{"documented Go version drifts", map[string]string{"README.md": strings.Replace(cleanTree["README.md"], "go@1.26.11", "go@1.25.0", 1)}, "PARITY.install-guidance README.md documents Go 1.25.0"},
		{"documented flag the installer lacks", map[string]string{"README.md": strings.Replace(cleanTree["README.md"], "--relock", "--force", 1)}, "PARITY.install-guidance README.md documents scripts/bootstrap-tools.sh --force"},
		{"install list omits a pinned tool", map[string]string{"README.md": strings.Replace(cleanTree["README.md"], " and Prettier", "", 1)}, "PARITY.install-guidance README.md the install list omits prettier"},
		{"package manager without an exact pin", map[string]string{"toolchain/node/package.json": `{"dependencies": {"prettier": "3.6.2"}}`}, "PARITY.package-manager toolchain/node/package.json packageManager"},
		{"foreign lock file", map[string]string{"toolchain/node/yarn.lock": "# yarn\n"}, "PARITY.package-manager toolchain/node/yarn.lock lock file of a package manager"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := problems(t, tc.edits)
			for _, g := range got {
				if strings.HasPrefix(g, tc.want) {
					return
				}
			}
			t.Fatalf("want a problem starting %q, got %v", tc.want, got)
		})
	}
}

func TestBundledNpmVersionMustMatchPackageManager(t *testing.T) {
	tree := writeTree(t, cleanTree, nil)
	tools := t.TempDir()
	npm := filepath.Join(tools, "node", "lib", "node_modules", "npm")
	if err := os.MkdirAll(npm, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(npm, "package.json"), []byte(`{"version": "10.9.2"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	tree.ToolsDir = tools
	got, err := Check(tree)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || !strings.Contains(got[0].Problem, "bundles npm 10.9.2") {
		t.Fatalf("want one bundled-npm mismatch, got %+v", got)
	}
}

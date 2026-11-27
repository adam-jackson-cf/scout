//go:build boundary

package pyquality

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"scout/tools/gate/check"
	"scout/tools/gate/proof"
)

const policy = `schema = 1

[scope]
mode = "whole"

[python_quality]
maintained_roots = ["src", "bin", "tests"]
script_roots = ["bin"]
script_names = ["run-job.py", "shared.py"]
test_roots = ["tests"]
test_groups = ["unit"]
generic_name_allowlist = ["src/legacy"]
class_max_lines = 700
class_max_init_args = 3
class_max_instance_attributes = 3
subprocess_wrappers = { "src/seam.py" = ["run_git"] }
dynamic_import_adapters = ["src/adapter.py"]
contract_files = ["src/contract.py"]
max_loop_depth = 2
fanout_max = 2
large_module_nloc = 5
large_module_fanout_max = 1
`

func runFamily(t *testing.T, name string, files map[string]string) []check.Finding {
	t.Helper()
	root := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("proofs.toml", policy)
	var paths []string
	for rel, body := range files {
		write(rel, body)
		paths = append(paths, rel)
	}
	pf, err := proof.Load(filepath.Join(root, "proofs.toml"))
	if err != nil {
		t.Fatal(err)
	}
	r := &check.Run{Root: root, Mode: check.CI, Paths: paths, Explicit: true, Proofs: pf, TempDir: t.TempDir()}
	for _, c := range Checks() {
		if c.Name == name {
			out, err := c.Run(r)
			if err != nil {
				t.Fatal(err)
			}
			return out
		}
	}
	t.Fatalf("no check %s", name)
	return nil
}

func rulesOf(found []check.Finding) string {
	var parts []string
	for _, f := range found {
		parts = append(parts, f.File+":"+f.Rule)
	}
	return strings.Join(parts, " ")
}

func TestQualityAcceptsApprovedSeamsAndShapes(t *testing.T) {
	found := runFamily(t, "python-quality", map[string]string{
		"src/seam.py":          "import subprocess\n\ndef run_git(a: list[str]) -> str:\n    def inner() -> str:\n        return 'x'\n    return subprocess.run(a, check=True, timeout=5).stdout + inner()\n",
		"src/adapter.py":       "import importlib.util\n\ndef load(p: str) -> object:\n    return importlib.util.spec_from_file_location('m', p)\n",
		"src/contract.py":      "from collections.abc import Callable\nHandler = Callable[[int], str]\n\ndef use(h: Handler) -> str:\n    return h(1)\n",
		"src/legacy/thing.py":  "X = 1\n",
		"bin/run-job.py":       "def main() -> None:\n    return None\n\nif __name__ == '__main__':\n    main()\n",
		"bin/shared.py":        "X = 1\n",
		"tests/unit/test_a.py": "def test_value_is_one() -> None:\n    assert 1\n",
		"src/small.py":         "from dataclasses import dataclass\n\n@dataclass\nclass P:\n    a: int\n    b: int\n    c: int\n",
		"src/cookie.py":        "# -*- coding: latin-1 -*-\nNAME = 'caf\xe9'\n",
	})
	if len(found) != 0 {
		t.Fatalf("unexpected findings: %v", found)
	}
}

func TestQualityRejectsEveryViolation(t *testing.T) {
	found := runFamily(t, "python-quality", map[string]string{
		"src/run.py":            "from subprocess import run as go\n\ndef f() -> None:\n    go(['ls'], check=True)\n",
		"src/seam.py":           "import os\n\ndef other() -> None:\n    os.system('ls')\n",
		"src/dyn.py":            "X = __import__('json')\n",
		"src/contract.py":       "from typing import Any\nfrom collections.abc import Callable\n\ndef use(h: Callable[[int], str], v: Any) -> None:\n    return None\n",
		"src/wide.py":           "from dataclasses import dataclass\n\n@dataclass\nclass W:\n    a: int\n    b: int\n    c: int\n    d: int\n",
		"src/attrs.py":          "class A:\n    x: int\n    def __init__(self) -> None:\n        self.a = 1\n        self.b = 2\n        self.c = 3\n",
		"src/tool.py":           "if __name__ == '__main__':\n    pass\n",
		"bin/run_job.py":        "#!/usr/bin/env python3\nX = 1\n",
		"tests/test_flat.py":    "def test_value_is_one() -> None:\n    assert 1\n",
		"tests/e2e/test_e2e.py": "def test_value_is_one() -> None:\n    assert 1\n",
		"src/common/x.py":       "X = 1\n",
		"src/misc.py":           "X = 1\n",
		"src/Bad-Name.py":       "X = 1\n",
		"lib/stray.py":          "X = 1\n",
		"src/broken.py":         "def (:\n",
	})
	want := []string{
		"src/run.py:PYQ220", "src/seam.py:PYQ220", "src/dyn.py:PYQ221", "src/contract.py:PYQ230",
		"src/contract.py:PYQ231", "src/wide.py:PYQ202", "src/attrs.py:PYQ203", "src/tool.py:PYQ101",
		"bin/run_job.py:PYQ101", "tests/test_flat.py:PYQ102", "tests/e2e/test_e2e.py:PYQ102",
		"src/common/x.py:PYQ103", "src/misc.py:PYQ104", "src/Bad-Name.py:PYQ107", "lib/stray.py:PYQ100",
		"src/broken.py:PYQ299",
	}
	got := rulesOf(found)
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("missing %s in %s", w, got)
		}
	}
	for _, f := range found {
		if f.Fix == "" || f.Problem == "" || f.Line < 1 {
			t.Errorf("finding lacks a problem, fix or line: %+v", f)
		}
	}
}

func TestScalabilityCountsLoopsComprehensionsAndResets(t *testing.T) {
	found := runFamily(t, "python-scalability", map[string]string{
		"src/ok.py":    "def f(n: int) -> list[int]:\n    out = []\n    for a in range(n):\n        out += [b for b in range(a)]\n        g = lambda: [c for c in range(3) for d in range(3)]\n    return out\n",
		"src/while.py": "def f(n: int) -> None:\n    while n:\n        for a in range(n):\n            for b in range(a):\n                n -= 1\n",
		"src/comp.py":  "X = [a for a in range(3) for b in range(3) for c in range(3)]\n",
		"src/bad.py":   "def (:\n",
	})
	got := rulesOf(found)
	if got != "src/bad.py:PYS000 src/comp.py:PYS250 src/while.py:PYS250" {
		t.Fatalf("got %s", got)
	}
}

func TestFanoutNormalisesFirstPartyAndRelativeImports(t *testing.T) {
	found := runFamily(t, "python-fanout", map[string]string{
		"src/pkg/__init__.py": "",
		"src/pkg/a.py":        "from . import b\nfrom .b import thing\nimport pkg.b\nimport json\nimport os.path\nimport requests\n",
		"src/pkg/b.py":        "thing = 1\n",
		"src/wide.py":         "import requests\nimport yaml\nimport attr\n",
		"src/big.py":          "import requests\nimport yaml\n" + strings.Repeat("X = 1\n", 6),
	})
	got := rulesOf(found)
	if got != "src/big.py:FANOUT.large-module src/pkg/a.py:FANOUT.large-module src/wide.py:FANOUT.general" {
		t.Fatalf("got %s", got)
	}
	if !strings.Contains(found[1].Problem, "has 2 architectural imports > 1 (pkg.b, requests)") {
		t.Fatalf("first-party imports not normalised: %s", found[1].Problem)
	}
}

func TestMissingPolicyFailsClosed(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.py"), []byte("X = 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	r := &check.Run{Root: root, Mode: check.CI, Paths: []string{"a.py"}, Explicit: true, Proofs: &proof.File{}, TempDir: t.TempDir()}
	for _, c := range Checks() {
		if _, err := c.Run(r); err == nil || !strings.Contains(err.Error(), "[python_quality]") {
			t.Errorf("%s: err = %v", c.Name, err)
		}
	}
}

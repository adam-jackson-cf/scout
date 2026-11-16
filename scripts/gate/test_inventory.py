"""Boundary canaries for the contributor repository gate (standard-library unittest)."""

from __future__ import annotations

import os
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
GOOD_GO = "package domain\n\nfunc Value() int { return 1 }\n"
BAD_GO = "package domain\nfunc Value()int{return 1}\n"
GOOD_TEST = """package domain

import "testing"

func TestValue(t *testing.T) {
	if Value() < 1 {
		t.Fatal("expected a positive value")
	}
}
"""


class GateFixture:
    """Run unchanged production checkers in an isolated, dependency-free Git repo."""

    def __init__(self, root: Path) -> None:
        self.root = root
        self.environment = {
            key: value
            for key, value in os.environ.items()
            if not key.startswith("GIT_")
        }
        # Resolve the installed tool before isolating HOME from tool-manager config.
        go_environment = subprocess.run(
            ["go", "env", "GOROOT", "GOCACHE"],
            cwd=ROOT,
            env={**self.environment, "GOPROXY": "off"},
            text=True,
            capture_output=True,
            check=True,
            timeout=15,
        ).stdout.splitlines()
        go_root, go_cache = go_environment
        self.environment.update({
            "GIT_CONFIG_NOSYSTEM": "1",
            "GIT_CONFIG_GLOBAL": os.devnull,
            "GIT_TERMINAL_PROMPT": "0",
            "GOPROXY": "off",
            "GOSUMDB": "off",
            "GOTOOLCHAIN": "local",
            "PATH": str(Path(go_root) / "bin") + os.pathsep + self.environment["PATH"],
            "GOCACHE": go_cache,
            "PYTHONDONTWRITEBYTECODE": "1",
            "HOME": str(root / "home"),
            "XDG_CONFIG_HOME": str(root / "home/config"),
            "XDG_CACHE_HOME": str(root / "home/cache"),
        })
        (root / "home").mkdir()
        # Copy the real checker/configuration, never replace owned subprocesses.
        for name in (
            "scripts/gate.sh",
            "scripts/bootstrap-tools.sh",
            "scripts/gate/imports.go",
            "scripts/gate/structure.json",
            ".github/workflows/gate.yml",
            "toolchain/versions.env",
            "toolchain/node/package.json",
            "toolchain/node/package-lock.json",
            "toolchain/python-requirements.in",
            "toolchain/python-requirements.lock",
            "CONTRIBUTING.md",
            "mypy.ini",
            "ruff.toml",
            ".prettierrc.json",
        ):
            target = root / name
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(ROOT / name, target)
        for source in sorted((ROOT / "scripts/gate").glob("*.py")):
            if not source.name.startswith("test_"):
                shutil.copyfile(source, root / "scripts/gate" / source.name)
        (root / ".tools").symlink_to(
            (ROOT / ".tools").resolve(), target_is_directory=True
        )
        version = next(
            line
            for line in (ROOT / "go.mod").read_text().splitlines()
            if line.startswith("toolchain ")
        )
        self.write("go.mod", f"module scout\n\ngo 1.26.0\n\n{version}\n")
        self.write(
            ".gitignore", ".tools\nhome/\n__pycache__/\n.mypy_cache/\n.ruff_cache/\n"
        )
        self.write("internal/domain/value.go", GOOD_GO)
        self.write("internal/domain/value_test.go", GOOD_TEST)
        self.write(
            "internal/areas/value.go",
            "package areas\n\nfunc Value() int { return 1 }\n",
        )
        self.write(
            "internal/areas/value_test.go",
            GOOD_TEST.replace("package domain", "package areas"),
        )
        self.write(
            "scripts/gate/sample.py",
            '"""A maintained gate input."""\n\nvalue: int = 1\n',
        )
        self.write(
            "internal/association/testdata/gen/sample.py",
            '"""A maintained generator input."""\n\nvalue: int = 1\n',
        )
        self.write("docs/sample.md", "# Sample\n\nA conforming contributor document.\n")
        self.write(
            "scripts/gate/test-tiers.json",
            """{
  "packages": {
    "internal/domain": {"tier": "fast"},
    "internal/areas": {"tier": "fast"}
  },
  "coverage": {"packages": ["internal/domain", "internal/areas"], "minimumPercent": 100}
}
""",
        )
        self.git("init", "--quiet")
        self.git("config", "user.name", "Gate Canary")
        self.git("config", "user.email", "gate-canary@example.invalid")
        self.git("add", ".")
        self.git("commit", "--quiet", "-m", "Conforming fixture")

    def write(self, name: str, content: str) -> None:
        target = self.root / name
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(content)

    def run(self, *argv: str) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            argv,
            cwd=self.root,
            env=self.environment,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            timeout=90,
            check=False,
        )

    def git(self, *args: str) -> None:
        result = self.run("git", *args)
        if result.returncode:
            raise AssertionError(f"fixture git {args!r} failed:\n{result.stdout}")

    def stage_valid_change(self) -> None:
        self.write("internal/domain/value.go", GOOD_GO.replace("return 1", "return 2"))
        self.git("add", "internal/domain/value.go")

    def gate(self) -> subprocess.CompletedProcess[str]:
        return self.run("bash", "scripts/gate.sh", "commit")


class InventoryCanaries(unittest.TestCase):
    """Whole-repository refusals and preserved staged-byte acceptance cases."""

    def setUp(self) -> None:
        self.temporary = tempfile.TemporaryDirectory(prefix="scout-gate-canary-")
        self.addCleanup(self.temporary.cleanup)
        root = Path(self.temporary.name) / "repo"
        root.mkdir()
        self.fixture = GateFixture(root)

    def assert_accepted(self, result: subprocess.CompletedProcess[str]) -> None:
        assert result.returncode == 0, result.stdout
        assert "fast:" in result.stdout, "all T1 tests must run:\n" + result.stdout
        assert "passing events=2" in result.stdout, result.stdout

    def assert_rejected(
        self, result: subprocess.CompletedProcess[str], marker: str
    ) -> None:
        assert result.returncode != 0, "gate accepted the violation:\n" + result.stdout
        assert marker in result.stdout, result.stdout

    def test_unstaged_tracked_go_drift_is_rejected(self) -> None:
        self.fixture.stage_valid_change()
        self.fixture.write("internal/areas/value.go", BAD_GO.replace("domain", "areas"))
        self.assert_rejected(self.fixture.gate(), "gofmt")

    def test_untracked_go_drift_is_rejected(self) -> None:
        self.fixture.stage_valid_change()
        self.fixture.write(
            "internal/domain/extra.go", "package domain\nfunc Extra()int{return 3}\n"
        )
        self.assert_rejected(self.fixture.gate(), "gofmt")

    def test_unstaged_python_drift_is_rejected(self) -> None:
        self.fixture.stage_valid_change()
        self.fixture.write(
            "scripts/gate/sample.py",
            '"""A maintained gate input."""\n\nvalue = missing_name\n',
        )
        self.assert_rejected(self.fixture.gate(), "F821")

    def test_untracked_markdown_drift_is_rejected(self) -> None:
        self.fixture.stage_valid_change()
        self.fixture.write("docs/new.md", "# New\n\n-   badly spaced list\n")
        self.assert_rejected(self.fixture.gate(), "prettier-format")

    def test_empty_index_still_checks_static_sources(self) -> None:
        self.fixture.write("internal/domain/value.go", BAD_GO)
        self.assert_rejected(self.fixture.gate(), "gofmt")

    def test_empty_index_runs_every_fast_package(self) -> None:
        self.fixture.write(
            "internal/areas/value.go",
            "package areas\n\nfunc Value() int { return 0 }\n",
        )
        self.assert_rejected(self.fixture.gate(), "TestValue")

    def test_conforming_empty_index_runs_all_fast_tests(self) -> None:
        self.assert_accepted(self.fixture.gate())

    def test_deleted_source_is_not_a_formatter_input(self) -> None:
        self.fixture.write(
            "internal/domain/obsolete.go", "package domain\n\nconst Old = 1\n"
        )
        self.fixture.git("add", "internal/domain/obsolete.go")
        self.fixture.git("commit", "--quiet", "-m", "Add removable source")
        self.fixture.git("rm", "internal/domain/obsolete.go")
        self.assert_accepted(self.fixture.gate())

    def test_partially_staged_source_uses_index_bytes(self) -> None:
        self.fixture.stage_valid_change()
        self.fixture.write(
            "internal/domain/value.go", BAD_GO.replace("return 1", "return 0")
        )
        before = self.fixture.run("git", "diff", "--binary").stdout
        self.assert_accepted(self.fixture.gate())
        after = self.fixture.run("git", "diff", "--binary").stdout
        assert after == before, f"working-tree diff changed:\n{before!r}\n!=\n{after!r}"

    def test_alternate_index_preserves_staged_bytes(self) -> None:
        alternate = self.fixture.root / "home/alternate-index"
        shutil.copyfile(self.fixture.root / ".git/index", alternate)
        self.fixture.environment["GIT_INDEX_FILE"] = str(alternate)
        self.fixture.stage_valid_change()
        self.fixture.write(
            "internal/domain/value.go", BAD_GO.replace("return 1", "return 0")
        )
        self.assert_accepted(self.fixture.gate())

    def test_tracked_source_symlink_cannot_escape_repository(self) -> None:
        # The target exists and is valid Go: failure must be containment, not ENOENT.
        outside = self.fixture.root.parent / "outside.go"
        outside.write_text("package domain\n\nconst Outside = 1\n")
        (self.fixture.root / "internal/domain/escape.go").symlink_to(outside)
        self.fixture.git("add", "internal/domain/escape.go")
        self.assert_rejected(self.fixture.gate(), "symlink")

    def test_commit_types_every_maintained_root(self) -> None:
        for name in (
            "scripts/gate/sample.py",
            "internal/association/testdata/gen/sample.py",
        ):
            with self.subTest(root=name):
                self.fixture.write(
                    name, '"""A maintained Python input."""\n\nvalue: int = "wrong"\n'
                )
                self.fixture.git("add", name)
                self.assert_rejected(self.fixture.gate(), "[assignment]")
                self.fixture.write(
                    name, '"""A maintained Python input."""\n\nvalue: int = 1\n'
                )
                self.fixture.git("add", name)

    def test_normal_import_following_checks_unselected_dependency(self) -> None:
        # A real explicit mypy target must still type-check its imported dependency.
        self.fixture.write(
            "scripts/gate/dependency.py",
            '"""An imported maintained module."""\n\nvalue: int = "wrong"\n',
        )
        self.fixture.write(
            "scripts/gate/sample.py",
            '"""A maintained gate input."""\n\n'
            "from dependency import value\n\nanswer: int = value\n",
        )
        self.fixture.git("add", "scripts/gate/dependency.py", "scripts/gate/sample.py")
        result = self.fixture.run(
            str(self.fixture.root / ".tools/bin/mypy"),
            "--config-file",
            "mypy.ini",
            "--explicit-package-bases",
            "--no-incremental",
            "scripts/gate/sample.py",
        )
        self.assert_rejected(result, "[assignment]")

    def test_conforming_staged_repository_passes(self) -> None:
        self.fixture.stage_valid_change()
        self.assert_accepted(self.fixture.gate())


if __name__ == "__main__":
    unittest.main(verbosity=2)

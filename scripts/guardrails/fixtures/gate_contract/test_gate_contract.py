"""Public command proofs for the installed repository gate contract."""

from __future__ import annotations

import ast
import hashlib
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[4]
SKILL = Path.home() / ".omp/agent/skills/setup-repository-guardrails"


def command(argv: list[str], root: Path) -> subprocess.CompletedProcess[str]:
    environment = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}
    environment.update({
        "GIT_CONFIG_NOSYSTEM": "1",
        "GIT_CONFIG_GLOBAL": os.devnull,
        "GIT_TERMINAL_PROMPT": "0",
        "GOPROXY": "off",
        "GOSUMDB": "off",
    })
    return subprocess.run(
        argv,
        cwd=root,
        env=environment,
        text=True,
        capture_output=True,
        check=False,
        timeout=180,
    )


def git(root: Path, *args: str) -> subprocess.CompletedProcess[str]:
    result = command(["git", *args], root)
    if result.returncode:
        raise AssertionError(result.stdout + result.stderr)
    return result


def write_json(path: Path, value: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(value, indent=2) + "\n")


def initialise(root: Path) -> None:
    git(root, "init", "-q", "-b", "main")
    git(root, "config", "user.name", "Gate proof")
    git(root, "config", "user.email", "proof@example.invalid")
    git(root, "config", "commit.gpgsign", "false")


def copy_contract(root: Path) -> None:
    for name in ("scripts", ".omp", ".github", ".githooks", "toolchain"):
        shutil.copytree(ROOT / name, root / name)
    for name in ("CONTRIBUTING.md", "ruff.toml", "mypy.ini", ".gitleaksignore"):
        shutil.copy2(ROOT / name, root / name)
    (root / ".tools").symlink_to(ROOT / ".tools", target_is_directory=True)


class GateConfigurationProof(unittest.TestCase):
    def test_p1_schema(self) -> None:
        config = json.loads((ROOT / ".omp/gate.json").read_text())
        assert config.get("schemaVersion") == 2, "P1: schemaVersion must be 2"
        for name in ("commit", "ci"):
            argv = config.get("checks", {}).get(name)
            assert isinstance(argv, list)
            assert argv
            assert all(isinstance(item, str) and item for item in argv)
        assert isinstance(config.get("pathCheck"), list)
        assert config.get("pathCheck")
        assert all(isinstance(item, str) and item for item in config["pathCheck"])
        assert config.get("ci") == {
            "requiredContext": "gate",
            "runner": "github-actions",
        }
        assert config.get("agentMerge") == "allowed"
        assert {"gate", "commit" + "Gate", "fast" + "Gate"}.isdisjoint(config)

    def test_p4_obsolete_surfaces_removed(self) -> None:
        assert not (ROOT / ".githooks/pre-push").exists(), "P4: remove pre-push hook"
        files = git(ROOT, "ls-files", "-z").stdout.split("\0")
        obsolete = re.compile(r"(?:fast|commit)Gate")
        violations = [
            name
            for name in files
            if name
            and not name.startswith("docs/adrs/")
            and (ROOT / name).is_file()
            and obsolete.search((ROOT / name).read_text(errors="replace"))
        ]
        assert violations == [], "P4: tracked obsolete gate fields"
        with tempfile.TemporaryDirectory(prefix="scout-no-push-proof-") as directory:
            root = Path(directory)
            initialise(root)
            copy_contract(root)
            result = command(["bash", "scripts/gate.sh", "fast"], root)
        assert result.returncode != 0, "P4: fast must not be an accepted mode"
        assert re.search(
            r"(?i)(invalid|unknown|use |unsupported|unrecognized)",
            result.stdout + result.stderr,
        )

    def test_p5_ci_argv(self) -> None:
        workflow = (ROOT / ".github/workflows/gate.yml").read_text()
        # Parse this repository's block-style YAML steps without a YAML dependency.
        triggers = re.search(r"(?m)^on:\n((?:[ \t]+.*\n)+)", workflow)
        assert triggers is not None
        assert re.search(r"(?m)^  push:", triggers[1])
        assert re.search(r"(?m)^  pull_request:", triggers[1])
        job = re.search(r"(?ms)^  gate:\n(.*?)(?=^  \w|\Z)", workflow)
        assert job is not None, "P5: missing gate job"
        steps = re.split(r"(?m)^      - ", job[1])[1:]
        runs = []
        for step in steps:
            run = re.search(r"(?m)^(?:        )?run: (.*)$", step)
            if run is not None:
                if run[1] == "|":
                    body = re.search(
                        r"(?m)^(?:        )?run: \|\n((?:          .*\n?)+)", step
                    )
                    assert body is not None
                    runs.append("\n".join(line[10:] for line in body[1].splitlines()))
                else:
                    runs.append(run[1])
        selector = re.compile(r"\[([\"\x27])checks\1\]\s*\[([\"\x27])ci\2\]")
        declared = [run for run in runs if selector.search(run)]
        assert len(declared) == 1, "P5: YAML step must read checks.ci argv"
        source = re.search(r"python\s+-c\s+'(.*)'", declared[0])
        assert source is not None, "P5: expose auditable Python argv execution"
        tree = ast.parse(source[1])
        calls = [node for node in ast.walk(tree) if isinstance(node, ast.Call)]
        runners = [node for node in calls if ast.unparse(node.func) == "subprocess.run"]
        assert len(runners) == 1
        assert len(runners[0].args) == 1
        assert ast.dump(runners[0].args[0]) == ast.dump(
            ast.parse(
                'json.load(open(".omp/gate.json"))["checks"]["ci"]', mode="eval"
            ).body
        )
        exits = [node for node in calls if ast.unparse(node.func) == "sys.exit"]
        assert len(exits) == 1
        assert ast.unparse(exits[0].args[0]) == ast.unparse(runners[0]) + ".returncode"
        text = "\n".join(runs)
        assert "source toolchain/versions.env" in text
        assert "go-version: ${{ env.GO_VERSION }}" in job[1]
        assert "bash scripts/ci-sandbox.sh --strict" in text
        assert text.index("ci-sandbox.sh --strict") < text.index(
            "scripts/bootstrap-tools.sh"
        )

    def test_p6_vendored_coverage(self) -> None:
        vendor = ROOT / "scripts/guardrails/vendor"
        for name, source in (
            ("guardrail-rules.json", "assets/guardrail-rules.json"),
            ("validate_rule_coverage.py", "scripts/validate_rule_coverage.py"),
        ):
            assert (vendor / name).is_file(), "P6: missing " + name
            assert (
                hashlib.sha256((vendor / name).read_bytes()).hexdigest()
                == hashlib.sha256((SKILL / source).read_bytes()).hexdigest()
            )
        assert not (vendor / "enaible-rules.json").exists()
        catalog = json.loads((vendor / "guardrail-rules.json").read_text())["rules"]
        ledger = json.loads((ROOT / ".omp/coverage-ledger.json").read_text())["rules"]
        ids = [row["id"] for row in catalog]
        assert len(ids) == 104
        assert len(set(ids)) == 104
        assert sorted(row["id"] for row in ledger) == sorted(ids)
        result = command(
            [
                sys.executable,
                str(vendor / "validate_rule_coverage.py"),
                "--root",
                str(ROOT),
                "--catalog",
                str(vendor / "guardrail-rules.json"),
                "--ledger",
                ".omp/coverage-ledger.json",
            ],
            ROOT,
        )
        assert result.returncode == 0, result.stdout + result.stderr

    def test_p8_path_preservation(self) -> None:
        argv = json.loads((ROOT / ".omp/gate.json").read_text())["pathCheck"]
        clean = command(
            [*argv, "scripts/guardrails/fixtures/gate_contract/test_gate_contract.py"],
            ROOT,
        )
        assert clean.returncode == 0, clean.stdout + clean.stderr
        bad = command([*argv, "unowned/proof.txt"], ROOT)
        assert bad.returncode != 0
        assert bad.stdout.splitlines() == [
            "unowned/proof.txt: STRUCT.folder-placement: place file in an approved owner path"
        ]
        assert bad.stderr == ""


class CommitHookProof(unittest.TestCase):
    def test_p2_exact_argv_and_exit(self) -> None:
        with tempfile.TemporaryDirectory(prefix="scout-hook-proof-") as directory:
            root = Path(directory)
            git(
                root,
                "clone",
                "--quiet",
                "--local",
                "--no-hardlinks",
                str(ROOT),
                str(root / "repo"),
            )
            root = root / "repo"
            git(root, "config", "user.name", "Gate proof")
            git(root, "config", "user.email", "proof@example.invalid")
            git(root, "config", "commit.gpgsign", "false")
            git(root, "config", "core.hooksPath", ".githooks")
            (root / ".tools/venv/bin").mkdir(parents=True)
            (root / ".tools/venv/bin/python").symlink_to(sys.executable)
            shutil.copy2(ROOT / ".githooks/pre-commit", root / ".githooks/pre-commit")
            shutil.copy2(
                ROOT / "scripts/guardrails/hook.py", root / "scripts/guardrails/hook.py"
            )
            recorder = root / "record.py"
            recorder.write_text(
                "import json,pathlib,sys\npathlib.Path('record.json').write_text(json.dumps(sys.argv[1:]))\nsys.exit(int(pathlib.Path('status').read_text()))\n"
            )
            argv = [
                sys.executable,
                "record.py",
                "space argument",
                ";not-shell",
                "--literal",
            ]
            config = {"schemaVersion": 2, "checks": {"commit": argv}}
            write_json(root / ".omp/gate.json", config)
            (root / "status").write_text("23")
            git(root, "add", ".omp/gate.json", "record.py", "status")
            denied = command(["git", "commit", "-qm", "must block"], root)
            assert denied.returncode != 0, "P2: failing checks.commit must block"
            assert (root / "record.json").exists(), (
                "P2: hook did not execute checks.commit"
            )
            assert json.loads((root / "record.json").read_text()) == argv[2:]
            direct = command(["bash", ".githooks/pre-commit"], root)
            assert direct.returncode == 23, direct.stdout + direct.stderr
            (root / "status").write_text("0")
            git(root, "add", "status")
            allowed = command(["git", "commit", "-qm", "must pass"], root)
            assert allowed.returncode == 0, allowed.stdout + allowed.stderr
            assert json.loads((root / "record.json").read_text()) == argv[2:]


if __name__ == "__main__":
    unittest.main()

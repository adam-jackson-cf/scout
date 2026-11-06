#!/usr/bin/env python3
"""Prove pre-commit blocking and isolated staged snapshot preservation."""

from __future__ import annotations

import importlib.util
import json
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path
from types import ModuleType
from typing import Any

ROOT = Path(__file__).resolve().parents[2]


def run(args: list[str], root: Path, **kwargs: Any) -> subprocess.CompletedProcess[str]:
    kwargs.setdefault("check", True)
    environment = {
        key: value for key, value in os.environ.items() if not key.startswith("GIT_")
    }
    return subprocess.run(
        args, cwd=root, text=True, capture_output=True, env=environment, **kwargs
    )


def git(root: Path, *args: str) -> str:
    return run(["git", "-c", "core.hooksPath=/dev/null", *args], root).stdout.strip()


def state(root: Path) -> dict[str, str]:
    commands = [
        ["for-each-ref"],
        ["config", "--list", "--local"],
        ["status", "--porcelain"],
        ["stash", "list"],
        ["diff", "--binary"],
        ["diff", "--cached", "--binary"],
    ]
    return {" ".join(command): git(root, *command) for command in commands}


def load_gate() -> ModuleType:
    spec = importlib.util.spec_from_file_location(
        "gate", ROOT / "scripts/guardrails/gate.py"
    )
    assert spec
    assert spec.loader
    gate = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(gate)
    return gate


def snapshots() -> None:
    gate = load_gate()
    with tempfile.TemporaryDirectory(prefix="scout-hook-snapshots-") as tmp:
        root = Path(tmp) / "repo"
        root.mkdir()
        git(root, "init", "-q")
        git(root, "config", "user.name", "Canary")
        git(root, "config", "user.email", "canary@example.invalid")
        git(root, "config", "commit.gpgsign", "false")
        path = root / "value.txt"
        path.write_text("staged-invalid")
        git(root, "add", "value.txt")
        git(root, "commit", "-qm", "base")
        path.write_text("index-content")
        git(root, "add", "value.txt")
        git(root, "update-index", "--split-index")
        alternate = root / ".git/alternate-index"
        shutil.copy2(root / ".git/index", alternate)
        path.write_text("different-working-tree-content")
        git(root, "add", "value.txt")
        before = state(root)
        old = Path.cwd()
        previous_index = os.environ.get("GIT_INDEX_FILE")
        os.environ["GIT_INDEX_FILE"] = str(alternate)
        try:
            os.chdir(root)
            gate.__dict__["ROOT"] = root
            destination = Path(tmp) / "index"
            gate.materialize(destination)
            assert (destination / "value.txt").read_text() == "index-content"
            assert git(destination, "rev-parse", "--absolute-git-dir") != git(
                root, "rev-parse", "--absolute-git-dir"
            )
            assert state(root) == before
        finally:
            os.chdir(old)
            if previous_index is None:
                os.environ.pop("GIT_INDEX_FILE", None)
            else:
                os.environ["GIT_INDEX_FILE"] = previous_index


def main() -> int:
    if sys.argv[1:] != ["pre-commit"]:
        raise ValueError("use pre-commit")
    hook = "pre-commit"
    entry = ROOT / ".githooks" / hook
    if (
        not entry.is_file()
        or not entry.stat().st_mode & 0o111
        or "scripts/guardrails/hook.py" not in entry.read_text()
    ):
        return 1
    config = json.loads((ROOT / ".omp/gate.json").read_text())
    if config.get("schemaVersion") != 2 or not config.get("checks", {}).get("commit"):
        return 1
    for kind, expected in [("positive", 0), ("negative", 1)]:
        fixture = ROOT / "scripts/guardrails/fixtures" / f"GATE.{hook}.{kind}.json"
        result = run(
            [sys.executable, str(ROOT / "scripts/guardrails/canary.py"), str(fixture)],
            ROOT,
            check=False,
        )
        if (result.returncode != 0) != bool(expected):
            return 1
    snapshots()
    print(
        f"GATE.{hook}: installed hook blocking and independent staged checkout proved"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())

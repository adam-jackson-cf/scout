#!/usr/bin/env python3
"""Prove native hook blocking and index/non-HEAD snapshot semantics."""

from __future__ import annotations

import importlib.util
import json
import os
import subprocess
import sys
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]


def run(args: list[str], root: Path) -> str:
    return subprocess.check_output(args, cwd=root, text=True).strip()


def snapshots() -> None:
    spec = importlib.util.spec_from_file_location(
        "gate", ROOT / "scripts/guardrails/gate.py"
    )
    assert spec
    assert spec.loader
    gate = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(gate)
    with tempfile.TemporaryDirectory(prefix="scout-hook-snapshots-") as tmp:
        root = Path(tmp) / "repo"
        root.mkdir()
        run(["git", "init", "-q"], root)
        run(["git", "config", "user.name", "Canary"], root)
        run(["git", "config", "user.email", "canary@example.invalid"], root)
        path = root / "value.txt"
        path.write_text("staged-invalid")
        run(["git", "add", "value.txt"], root)
        run(["git", "commit", "-qm", "base"], root)
        base = run(["git", "rev-parse", "HEAD"], root)
        path.write_text("index-content")
        run(["git", "add", "value.txt"], root)
        path.write_text("different-working-tree-content")
        old = Path.cwd()
        os.chdir(root)
        gate.__dict__["ROOT"] = root
        destination = Path(tmp) / "index"
        destination.mkdir()
        gate.materialize(destination, None)
        assert (destination / "value.txt").read_text() == "index-content"
        run(["git", "commit", "-qm", "pushed"], root)
        pushed = run(["git", "rev-parse", "HEAD"], root)
        # Leave HEAD on base; the pushed alternate ref must still be checked.
        run(["git", "branch", "alternate", pushed], root)
        run(["git", "symbolic-ref", "HEAD", "refs/heads/other"], root)
        run(["git", "update-ref", "refs/heads/other", base], root)
        os.environ["SCOUT_PUSH_RANGES"] = (
            f"refs/heads/alternate {pushed} refs/heads/main {'0' * 40}\n"
        )
        assert gate.pushed_commits(base) == [pushed]
        del os.environ["SCOUT_PUSH_RANGES"]
        os.chdir(old)


def main() -> int:
    hook = sys.argv[1]
    field = {"pre-commit": "commitGate", "pre-push": "fastGate"}[hook]
    entry = ROOT / ".githooks" / hook
    if (
        not entry.is_file()
        or not entry.stat().st_mode & 0o111
        or "scripts/guardrails/hook.py" not in entry.read_text()
    ):
        return 1
    config = json.loads((ROOT / ".omp/gate.json").read_text())
    if not config.get(field):
        return 1
    for kind, expected in [("positive", 0), ("negative", 1)]:
        fixture = ROOT / "scripts/guardrails/fixtures" / f"GATE.{hook}.{kind}.json"
        result = subprocess.run(
            [sys.executable, str(ROOT / "scripts/guardrails/canary.py"), str(fixture)],
            check=False,
        )
        if (result.returncode != 0) != bool(expected):
            return 1
    snapshots()
    print(
        f"GATE.{hook}: isolated installed hook propagates declared argv failure; staged/non-HEAD snapshots proved"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())

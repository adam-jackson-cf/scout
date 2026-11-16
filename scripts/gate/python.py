#!/usr/bin/env python3
"""Lint, format-check and type-check maintained Python with the pinned tools."""

from __future__ import annotations

import subprocess
import sys
from pathlib import Path

from inventory import repository_files

ROOT = Path(__file__).resolve().parents[2]
EXCLUDED_ROOTS = ("evaluation/", "third_party/", ".tools/")


def maintained(paths: list[str]) -> list[str]:
    return [
        p
        for p in paths
        if p.endswith(".py")
        and not p.startswith(EXCLUDED_ROOTS)
        and (ROOT / p).is_file()
    ]


def run(argv: list[str], *, capture: bool = False) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        argv, cwd=ROOT, text=True, capture_output=capture, check=False, timeout=300
    )


def tool(name: str, *args: str) -> int:
    return run([str(ROOT / ".tools/bin" / name), *args]).returncode


def main() -> int:
    mode = sys.argv[1]
    if mode not in {"full", "commit"}:
        raise ValueError("unknown mode: " + mode)
    files = maintained(repository_files())
    if not files:
        return 0
    failed = tool("ruff", "check", "--config", "ruff.toml", *files)
    failed |= tool("ruff", "format", "--config", "ruff.toml", "--check", *files)
    for root in ("scripts/gate/", "internal/association/testdata/gen/"):
        targets = [name for name in files if name.startswith(root)]
        if targets:
            failed |= tool(
                "mypy",
                "--config-file",
                "mypy.ini",
                "--explicit-package-bases",
                *targets,
            )
    return int(failed != 0)


if __name__ == "__main__":
    sys.exit(main())

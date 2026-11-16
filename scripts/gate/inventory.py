"""Enumerate maintained repository files without walking tool or vendor trees."""

from __future__ import annotations

import os
import subprocess
from pathlib import Path

EXCLUDED = {
    ".git",
    ".tools",
    ".codegraph",
    "evaluation",
    "third_party",
    "vendor",
    "node_modules",
    "__pycache__",
    ".mypy_cache",
    ".ruff_cache",
}


def excluded(name: str) -> bool:
    return bool(EXCLUDED.intersection(Path(name).parts))


def repository_files() -> list[str]:
    """Respect the caller's index; reject escaping links before opening sources."""
    root = Path.cwd().resolve()
    result = subprocess.run(
        ["git", "ls-files", "--cached", "--others", "--exclude-standard", "-z"],
        text=True,
        capture_output=True,
        check=True,
        timeout=30,
        env=dict(os.environ),
    )
    files = []
    for name in sorted(set(result.stdout.split("\0")) - {""}):
        if excluded(name):
            continue
        path = root / name
        if not path.resolve().is_relative_to(root):
            raise ValueError(
                f"{name}: source-symlink: replace escaping symlink with an in-repository source"
            )
        files.append(name)
    return files

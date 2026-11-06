#!/usr/bin/env python3
"""Dispatch the exact schema-2 checks.commit argv."""

from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path


def main() -> int:
    root = Path(
        subprocess.check_output(
            ["git", "rev-parse", "--show-toplevel"], text=True
        ).strip()
    )
    if sys.argv[1:] != ["pre-commit"]:
        raise ValueError("use pre-commit")
    config = json.loads((root / ".omp/gate.json").read_text())
    if config.get("schemaVersion") != 2:
        raise ValueError("configure schemaVersion 2")
    argv = config["checks"]["commit"]
    if (
        not isinstance(argv, list)
        or not argv
        or not all(isinstance(x, str) and x for x in argv)
    ):
        raise ValueError("configure a nonempty argv")
    return subprocess.run(argv, cwd=root, check=False).returncode


if __name__ == "__main__":
    sys.exit(main())

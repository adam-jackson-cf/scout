#!/usr/bin/env python3
"""Dispatch exact configured argv, retaining Git's pre-push ref updates."""

from __future__ import annotations

import json
import os
import subprocess
import sys
from pathlib import Path


def main() -> int:
    root = Path(
        subprocess.check_output(
            ["git", "rev-parse", "--show-toplevel"], text=True
        ).strip()
    )
    field = {"pre-commit": "commitGate", "pre-push": "fastGate"}[sys.argv[1]]
    argv = json.loads((root / ".omp/gate.json").read_text())[field]
    if (
        not isinstance(argv, list)
        or not argv
        or not all(isinstance(x, str) and x for x in argv)
    ):
        raise ValueError("configure a nonempty argv")
    environment = dict(os.environ)
    if field == "fastGate":
        environment["SCOUT_PUSH_RANGES"] = sys.stdin.read()
        environment["SCOUT_PUSH_REMOTE"] = (
            sys.argv[2] if len(sys.argv) > 2 else "origin"
        )
        if len(sys.argv) > 3:
            environment["SCOUT_PUSH_REMOTE_URL"] = sys.argv[3]
    return subprocess.run(argv, cwd=root, env=environment, check=False).returncode


if __name__ == "__main__":
    sys.exit(main())

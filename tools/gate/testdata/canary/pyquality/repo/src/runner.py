from __future__ import annotations

import subprocess


def run_tool(argv: list[str]) -> str:
    return subprocess.run(argv, capture_output=True, text=True, check=True, timeout=30).stdout

#!/usr/bin/env python3
"""Merge independently owned fragments, rejecting duplicate rule ownership."""

from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any


def merged_ledger(root: Path) -> tuple[str, int]:
    """Render fragments deterministically without changing tracked files."""
    rows: dict[str, Any] = {}
    for path in sorted((root / ".omp/ledger.d").glob("*.json")):
        for row in json.loads(path.read_text())["rows"]:
            if row["id"] in rows:
                raise ValueError(f"duplicate rule ownership: {row['id']}")
            rows[row["id"]] = row
    output = {
        "schema_version": 1,
        "scope": {
            "mode": "whole-repo",
            "paths": [],
            "protected": [
                "evaluation/**",
                "**/.codegraph/**",
                "governed design/ledger documents",
            ],
        },
        "rules": [rows[key] for key in sorted(rows)],
    }
    return json.dumps(output, indent=2) + "\n", len(rows)


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[2]
    content, count = merged_ledger(root)
    destination = root / ".omp/coverage-ledger.json"
    if args.check:
        if not destination.is_file() or destination.read_bytes() != content.encode():
            print(
                ".omp/coverage-ledger.json: GATE.local-ci-parity: run scripts/guardrails/merge-ledger.py to synchronize fragments"
            )
            raise SystemExit(1)
        print(f"Coverage ledger matches {count} fragment rows")
        return
    destination.write_text(content)
    print(f"Merged {count} rules from .omp/ledger.d/*.json")


if __name__ == "__main__":
    main()

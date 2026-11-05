"""Exercise real Python checkers against isolated rule inputs."""

from __future__ import annotations

import configparser
import json
import tempfile
from pathlib import Path
from typing import Any

import policy
import scalability
from check import ROOT, ast_errors, decode_source, diagnostic, lint_source, run_tool


def fixture_findings(fixture: dict[str, Any]) -> list[str]:
    kind = fixture["kind"]
    if kind == "ruff-config":
        return policy.ruff_errors(fixture["config"])
    if kind == "mypy-config":
        config = configparser.ConfigParser()
        config.read_string(fixture["config"])
        return policy.mypy_errors(config)
    if kind == "coverage":
        return policy.coverage_errors(fixture["maintained"], fixture["checked"])
    return boundary_findings(fixture)


def boundary_findings(fixture: dict[str, Any]) -> list[str]:
    if fixture["kind"] == "encoding":
        return [rule for rule, _fix in decode_source(bytes.fromhex(fixture["hex"]))[1]]
    if fixture["kind"] == "ownership":
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / scalability.ANALYZER_PREFIX).mkdir(parents=True)
            if fixture["owned"]:
                (root / "CODEOWNERS").write_text(
                    "/scripts/guardrails/analyzers/ @scout-maintainers\n"
                )
            return [d.code for d in scalability.ownership_diagnostics(root)]
    return source_findings(fixture)


def source_findings(fixture: dict[str, Any]) -> list[str]:
    source, relpath = fixture["source"], fixture["path"]
    if fixture["kind"] == "ruff":
        return [rule for rule, _fix in lint_source(source, relpath)]
    if fixture["kind"] == "mypy":
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "canary.py"
            path.write_text(source)
            result = run_tool([
                str(ROOT / ".tools/bin/mypy"),
                "--config-file",
                str(ROOT / "mypy.ini"),
                str(path),
            ])
            code = (
                "[unused-ignore]"
                if fixture["rule"] == "MYPY.warn_unused_ignores"
                else "[no-untyped-def]"
            )
            return (
                [fixture["rule"]]
                if result.returncode == 1 and code in result.stdout
                else []
            )
    return [rule for rule, _fix in ast_errors(source, relpath)]


def check_fixture(path: Path) -> int:
    fixture: dict[str, Any] = json.loads(path.read_text())
    findings = fixture_findings(fixture)
    for rule in findings:
        diagnostic(
            fixture.get("path", str(path)),
            rule,
            "repair the violating Python contract input",
        )
    return int(bool(findings))

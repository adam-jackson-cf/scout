"""Exercise real Python checkers, preserving their actionable diagnostics."""

from __future__ import annotations

import configparser
import json
import tempfile
from pathlib import Path
from typing import Any

from check import ROOT, ast_errors, decode_source, diagnostic, lint_source, run_tool

import policy
import scalability


def fixture_findings(fixture: dict[str, Any]) -> list[tuple[str, str]]:
    kind = fixture["kind"]
    if kind == "ruff-config":
        return [
            (rule, "restore the approved Ruff setting " + rule.rsplit(".", 1)[-1])
            for rule in policy.ruff_errors(fixture["config"], fixture.get("target"))
        ]
    if kind == "mypy-config":
        config = configparser.ConfigParser()
        config.read_string(fixture["config"])
        return [
            (rule, "restore the approved mypy setting " + rule.rsplit(".", 1)[-1])
            for rule in policy.mypy_errors(config, fixture.get("target"))
        ]
    if kind == "coverage":
        missing = sorted(set(fixture["maintained"]) - set(fixture["checked"]))
        return [
            (
                rule,
                "add maintained files to the explicit mypy invocation: "
                + ", ".join(missing),
            )
            for rule in policy.coverage_errors(
                fixture["maintained"],
                fixture["checked"],
                fixture.get("target"),
                fixture.get("packages"),
            )
        ]
    return boundary_findings(fixture)


def boundary_findings(fixture: dict[str, Any]) -> list[tuple[str, str]]:
    if fixture["kind"] == "encoding":
        return decode_source(bytes.fromhex(fixture["hex"]))[1]
    if fixture["kind"] == "ownership":
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / scalability.ANALYZER_PREFIX).mkdir(parents=True)
            if fixture["owned"]:
                (root / "CODEOWNERS").write_text(
                    "/scripts/guardrails/analyzers/ @scout-maintainers\n"
                )
            return [
                (d.code, d.message) for d in scalability.ownership_diagnostics(root)
            ]
    return source_findings(fixture)


def source_findings(fixture: dict[str, Any]) -> list[tuple[str, str]]:
    source, relpath = fixture["source"], fixture["path"]
    if fixture["kind"] == "ruff":
        return lint_source(source, relpath)
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
            return [
                (fixture["rule"], "repair typing: " + line.split("error: ", 1)[-1])
                for line in result.stdout.splitlines()
                if result.returncode == 1 and code in line and "error: " in line
            ]
    target = fixture.get("target", policy.target_policy())
    return ast_errors(source, relpath, target["canonical_test_groups"])


def check_fixture(path: Path) -> int:
    fixture: dict[str, Any] = json.loads(path.read_text())
    rule_id = fixture.get("rule", path.parent.name)
    findings = [
        finding for finding in fixture_findings(fixture) if finding[0] == rule_id
    ]
    for rule, fix in findings:
        diagnostic(fixture.get("path", str(path)), rule, fix)
    return int(bool(findings))

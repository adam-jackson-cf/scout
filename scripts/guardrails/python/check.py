#!/usr/bin/env python3
"""Check every maintained Python rule without mutating source files."""

from __future__ import annotations

import argparse
import configparser
import io
import json
import subprocess
import tokenize
from pathlib import Path
from typing import Any

import fanout
import policy
import quality
import scalability
from python_quality_scope import (
    discover_maintained_python_files,
    discover_unexcluded_python_files,
    is_maintained_python_relpath,
)

ROOT = Path(__file__).resolve().parents[3]
LIMIT_RULES = {
    "PLR0904": "max-public-methods",
    "PLR0911": "max-returns",
    "PLR0912": "max-branches",
    "PLR0913": "max-args",
    "PLR0914": "max-locals",
    "PLR0915": "max-statements",
    "PLR0917": "max-positional-args",
}


def run_tool(
    argv: list[str], *, source: str | None = None
) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        argv,
        cwd=ROOT,
        input=source,
        text=True,
        capture_output=True,
        check=False,
        timeout=120,
    )


def diagnostic(path: str, rule: str, fix: str) -> None:
    print(f"{path}: {rule}: {' '.join(fix.splitlines())}")


def ruff_rules(code: str) -> list[str]:
    selectors = [
        "E",
        "W",
        "F",
        "I",
        "B",
        "C4",
        "C901",
        "UP",
        "SIM",
        "PT",
        "PGH",
        "A",
        "D",
        "N",
        *LIMIT_RULES,
    ]
    rules = [f"RUFF.SELECT.{prefix}" for prefix in selectors if code.startswith(prefix)]
    if code in LIMIT_RULES:
        rules.append(f"RUFF.LIMIT.{LIMIT_RULES[code]}")
    if code == "C901":
        rules.append("RUFF.SETTING.mccabe.max-complexity")
    return rules


def lint_source(source: str, relpath: str) -> list[tuple[str, str]]:
    result = run_tool(
        [
            str(ROOT / ".tools/bin/ruff"),
            "check",
            "--config",
            str(ROOT / "ruff.toml"),
            "--no-fix",
            "--output-format",
            "json",
            "--stdin-filename",
            relpath,
            "-",
        ],
        source=source,
    )
    if result.returncode not in (0, 1):
        raise RuntimeError(result.stderr)
    findings: list[Any] = json.loads(result.stdout)
    return suppression_errors(source) + [
        (rule, f"line {item['location']['row']}: {item['message']}")
        for item in findings
        for rule in ruff_rules(item["code"])
    ]


def suppression_errors(source: str) -> list[tuple[str, str]]:
    findings: list[tuple[str, str]] = []
    try:
        comments = [
            token.string.lower()
            for token in tokenize.generate_tokens(io.StringIO(source).readline)
            if token.type == tokenize.COMMENT
        ]
    except (tokenize.TokenError, IndentationError):
        return findings  # AST parsing reports malformed source separately.
    for comment in comments:
        if "noqa" in comment or "ruff:" in comment:
            findings.append((
                "RUFF.IGNORES",
                "remove the unapproved source-level Ruff suppression",
            ))
        if "mypy:" in comment:
            findings.append((
                "MYPY.strict",
                "remove the unapproved source-level mypy override",
            ))
    return findings


def ast_errors(source: str, relpath: str) -> list[tuple[str, str]]:
    import ast

    try:
        tree = ast.parse(source, filename=relpath)
    except SyntaxError as exc:
        return [
            ("PYS000", f"repair syntax: {exc}"),
            ("PYQ299", f"repair syntax: {exc}"),
        ]
    findings = [
        (d.code, d.message) for d in quality.structure_diagnostics_for_relpath(relpath)
    ]
    findings.extend(
        (d.code, d.message)
        for d in quality.class_interface_diagnostics_for_source(source, relpath)
    )
    visitor = scalability.ScalabilityVisitor(relpath)
    visitor.visit(tree)
    findings.extend((d.code, d.message) for d in visitor.diagnostics)
    return findings


def config_errors() -> list[tuple[str, str]]:
    config = configparser.ConfigParser()
    config.read(ROOT / "mypy.ini")
    return [
        (rule, "restore the exact approved Python configuration")
        for rule in policy.ruff_errors(policy.read_ruff(ROOT / "ruff.toml"))
        + policy.mypy_errors(config)
    ]


def decode_source(data: bytes) -> tuple[str | None, list[tuple[str, str]]]:
    try:
        return data.decode("utf-8"), []
    except UnicodeDecodeError:
        return None, [("PYQ299", "encode maintained Python as UTF-8")]


def check_files(paths: list[Path], *, typecheck: bool) -> int:
    failures = 0
    for rule, fix in config_errors():
        diagnostic("ruff.toml / mypy.ini", rule, fix)
        failures += 1
    for path in paths:
        relpath = path.relative_to(ROOT).as_posix()
        source, findings = decode_source(path.read_bytes())
        if source is not None:
            findings += ast_errors(source, relpath) + lint_source(source, relpath)
        for rule, fix in findings:
            diagnostic(relpath, rule, fix)
            failures += 1
    failures += check_tools(paths, typecheck=typecheck)
    return int(failures != 0)


def formatting_errors(paths: list[Path]) -> int:
    failures = 0
    for path in paths:
        result = run_tool([
            str(ROOT / ".tools/bin/ruff"),
            "format",
            "--config",
            str(ROOT / "ruff.toml"),
            "--check",
            str(path),
        ])
        if result.returncode:
            diagnostic(
                path.relative_to(ROOT).as_posix(),
                "RUFF.FORMAT.quote-style",
                "run ruff format explicitly to apply the approved formatter settings",
            )
            failures += 1
    return failures


def report_typing(result: subprocess.CompletedProcess[str]) -> None:
    errors = [
        line.split(": error: ", 1)
        for line in result.stdout.splitlines()
        if ": error: " in line
    ]
    for location, message in errors:
        diagnostic(
            location.removeprefix(str(ROOT) + "/"),
            "MYPY.strict",
            "repair typing: " + message,
        )
    if result.returncode and not errors:
        diagnostic(
            "mypy.ini",
            "MYPY.strict",
            "repair mypy invocation: " + result.stdout + result.stderr,
        )


def check_tools(paths: list[Path], *, typecheck: bool) -> int:
    if not paths:
        return 0
    names = [str(path) for path in paths]
    formatted = formatting_errors(paths)
    ownership = scalability.ownership_diagnostics(ROOT)
    for finding in ownership:
        diagnostic(finding.relpath, finding.code, finding.message)
    if not typecheck:
        return int(bool(formatted or ownership))
    typing_argv = [
        str(ROOT / ".tools/bin/mypy"),
        "--config-file",
        str(ROOT / "mypy.ini"),
        "--explicit-package-bases",
        *names,
    ]
    coverage = policy.coverage_errors(names, typing_argv[4:])
    for rule in coverage:
        diagnostic(
            "maintained Python",
            rule,
            "pass every maintained source file explicitly to mypy",
        )
    typed = run_tool(typing_argv)
    report_typing(typed)
    return int(bool(formatted or typed.returncode or coverage or ownership))


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("mode", choices=("full", "commit", "fixture"))
    parser.add_argument("paths", nargs="*")
    args = parser.parse_args()
    if args.mode == "fixture":
        from canaries import check_fixture

        return check_fixture(Path(args.paths[0]))
    discovered = discover_maintained_python_files(ROOT)
    paths = (
        discovered
        if args.mode == "full"
        else [
            path
            for path in discover_unexcluded_python_files(
                ROOT, [ROOT / p for p in args.paths]
            )
            if is_maintained_python_relpath(path.relative_to(ROOT).as_posix())
        ]
    )
    failures = check_files(paths, typecheck=args.mode == "full")
    roots = fanout._normalise_source_roots(
        ROOT, ["scripts/guardrails", "internal/association/testdata/gen"]
    )
    results = fanout.analyse_fanout(
        ROOT,
        source_roots=roots,
        root_packages={"scripts", "internal"},
        limits=fanout.FanoutLimits(10, 800, 3),
    )
    selected = {p.relative_to(ROOT).as_posix() for p in paths}
    for result in results:
        if result.relpath in selected and result.reasons:
            diagnostic(
                result.relpath,
                "FANOUT.general" if result.count > 10 else "FANOUT.large-module",
                "split architectural dependencies: " + "; ".join(result.reasons),
            )
            failures = 1
    return failures


if __name__ == "__main__":
    raise SystemExit(main())

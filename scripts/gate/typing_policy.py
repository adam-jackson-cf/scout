"""Enforce mypy ownership and evidence-backed typing and Ruff exceptions.

Policy is the ``python`` object from structure.json. ``maintained_packages`` must
match populated ``maintained_roots``. ``mypy_overrides`` maps module patterns to
evidence strings. ``ruff_approvals`` maps exact suppression keys to evidence:
``ruff.toml:lint.ignore:CODE``, ``ruff.toml:lint.extend-ignore:CODE``,
``ruff.toml:lint.per-file-ignores:PATTERN:CODE`` (or extend-per-file-ignores),
``PATH:noqa:CODE`` and ``PATH:ruff-noqa:CODE``. Blanket suppression uses ``ALL``.
Approvals apply only to growth relative to the merge base, not existing debt.
"""

from __future__ import annotations

import configparser
import io
import re
import subprocess
import tokenize
import tomllib
from collections import Counter
from pathlib import Path
from typing import Any

POLICY_PATH = "scripts/gate/structure.json"
NOQA = re.compile(r"#\s*(?:(ruff|flake8)\s*:\s*)?noqa\b(?:\s*:\s*([^#]*))?", re.I)
RULE = re.compile(r"[A-Z]+[0-9]+\Z")


def _diagnostic(path: str, rule: str, fix: str) -> str:
    return f"{path}: {rule}: {fix}".replace("\n", " ").replace("\r", " ")


def _evidence(approvals: object, key: str) -> bool:
    if not isinstance(approvals, dict):
        return False
    evidence = approvals.get(key)
    return isinstance(evidence, str) and bool(evidence.strip())


def _roots(policy: dict[str, Any], key: str) -> list[str]:
    return [str(name).rstrip("/") + "/" for name in policy.get(key, [])]


def _coverage(root: Path, files: list[str], policy: dict[str, Any]) -> list[str]:
    declared = _roots(policy, "maintained_packages")
    populated = {
        prefix
        for prefix in _roots(policy, "maintained_roots")
        if any(name.startswith(prefix) for name in files)
    }
    diagnostics = []
    if set(declared) != populated or len(declared) != len(set(declared)):
        diagnostics.append(
            _diagnostic(
                POLICY_PATH,
                "MYPY.package-coverage",
                "make maintained_packages exactly match populated maintained root packages "
                f"{sorted(populated)!r}",
            )
        )
    for name in files:
        if not any(name.startswith(prefix) for prefix in declared):
            diagnostics.append(
                _diagnostic(
                    name,
                    "MYPY.package-coverage",
                    "declare this maintained package in maintained_packages and mypy_path",
                )
            )
    config = configparser.ConfigParser(interpolation=None)
    try:
        config.read_string((root / "mypy.ini").read_text())
    except (OSError, configparser.Error) as error:
        return diagnostics + [
            _diagnostic(
                "mypy.ini",
                "MYPY.package-coverage",
                f"repair package configuration: {error}",
            )
        ]
    paths = {
        name.strip().removeprefix("$MYPY_CONFIG_FILE_DIR/").rstrip("/") + "/"
        for name in re.split(r"[:,\n]", config.get("mypy", "mypy_path", fallback=""))
        if name.strip()
    }
    if not set(declared).issubset(paths):
        diagnostics.append(
            _diagnostic(
                "mypy.ini",
                "MYPY.package-coverage",
                "include every maintained package root in mypy_path: "
                + ", ".join(sorted(set(declared) - paths)),
            )
        )
    diagnostics.extend(_overrides(config, policy))
    return diagnostics


def _overrides(config: configparser.ConfigParser, policy: dict[str, Any]) -> list[str]:
    diagnostics = []
    approvals = policy.get("mypy_overrides", {})
    for section in config.sections():
        if section != "mypy" and not section.startswith("mypy-"):
            continue
        ignore = config.get(section, "ignore_missing_imports", fallback="false").lower()
        follow = config.get(section, "follow_imports", fallback="normal").lower()
        problem = ""
        modules = [name.strip() for name in section.removeprefix("mypy-").split(",")]
        if section != "mypy" and any(
            not _evidence(approvals, name) for name in modules
        ):
            problem = (
                "record evidence for every named module override in mypy_overrides"
            )
        if follow.strip() == "skip":
            problem = "remove the follow_imports=skip override; prefer follow_untyped_imports=True"
        elif ignore.strip() in {"1", "yes", "true", "on"}:
            modules = [
                name.strip() for name in section.removeprefix("mypy-").split(",")
            ]
            if section == "mypy" or "*" in modules:
                problem = "remove global ignore_missing_imports; use an evidence-backed module override"
            elif any(not _evidence(approvals, name) for name in modules):
                problem = (
                    "record nonempty mypy_overrides evidence for each module override; "
                    "prefer follow_untyped_imports=True"
                )
        if problem:
            diagnostics.append(
                _diagnostic(
                    "mypy.ini", "MYPY.module-overrides", f"{section}: {problem}"
                )
            )
    return diagnostics


def _git(root: Path, *arguments: str) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["git", *arguments],
        cwd=root,
        text=True,
        capture_output=True,
        check=False,
        timeout=15,
    )


def _config_suppressions(source: str) -> Counter[str]:
    config = tomllib.loads(source)
    lint = config.get("lint", {})
    suppressions: Counter[str] = Counter()
    # Older Ruff configurations may still put these options at the top level.
    for namespace, settings in (("", config), ("lint.", lint)):
        suppressions.update(_table_suppressions(namespace, settings))
    return suppressions


def _table_suppressions(namespace: str, settings: dict[str, Any]) -> Counter[str]:
    suppressions: Counter[str] = Counter()
    for option in ("ignore", "extend-ignore"):
        for code in settings.get(option, []):
            suppressions[f"ruff.toml:{namespace}{option}:{code}"] += 1
    for option in ("per-file-ignores", "extend-per-file-ignores"):
        suppressions.update(
            _pattern_suppressions(namespace + option, settings.get(option, {}))
        )
    return suppressions


def _pattern_suppressions(option: str, patterns: dict[str, list[str]]) -> Counter[str]:
    return Counter(
        f"ruff.toml:{option}:{pattern}:{code}"
        for pattern, codes in patterns.items()
        for code in codes
    )


def _inline_suppressions(path: str, source: str) -> Counter[str]:
    suppressions: Counter[str] = Counter()
    tokens = tokenize.generate_tokens(io.StringIO(source).readline)
    try:
        for token in tokens:
            if token.type != tokenize.COMMENT:
                continue
            match = NOQA.search(token.string)
            if match is None:
                continue
            kind = "ruff-noqa" if match[1] else "noqa"
            codes = match[2]
            # An unqualified noqa suppresses every rule, not an empty set.
            rules = (
                {"ALL"}
                if codes is None
                else {
                    code
                    for code in re.split(r"[\s,]+", codes.strip().upper())
                    if RULE.fullmatch(code)
                }
            )
            for code in rules:
                suppressions[f"{path}:{kind}:{code}"] += 1
    except (tokenize.TokenError, IndentationError):
        # Ruff diagnoses invalid Python; retain any suppressions already tokenized.
        pass
    return suppressions


def _growth(
    path: str,
    current: Counter[str],
    previous: Counter[str],
    approvals: object,
) -> list[str]:
    return [
        _diagnostic(
            path,
            "RUFF.IGNORES",
            f"remove added suppression or record approval evidence for {key}",
        )
        for key in sorted(current - previous)
        if not _evidence(approvals, key)
    ]


def _ruff(root: Path, files: list[str], policy: dict[str, Any]) -> list[str]:
    base = _git(root, "merge-base", "origin/main", "HEAD")
    if base.returncode:
        base = _git(root, "rev-parse", "--verify", "HEAD")
    if base.returncode:
        return [
            _diagnostic(
                "ruff.toml",
                "RUFF.IGNORES",
                "restore Git baseline before evaluating approval evidence",
            )
        ]
    revision = base.stdout.strip()
    previous = _git(root, "show", f"{revision}:ruff.toml")
    approvals = policy.get("ruff_approvals", {})
    try:
        diagnostics = _growth(
            "ruff.toml",
            _config_suppressions((root / "ruff.toml").read_text()),
            _config_suppressions(previous.stdout if previous.returncode == 0 else ""),
            approvals,
        )
    except (OSError, ValueError, TypeError, AttributeError) as error:
        return [
            _diagnostic(
                "ruff.toml",
                "RUFF.IGNORES",
                f"repair suppression approval configuration: {error}",
            )
        ]
    changed = _git(root, "diff", "--name-only", "-z", revision, "--")
    untracked = _git(root, "ls-files", "--others", "--exclude-standard", "-z")
    if changed.returncode or untracked.returncode:
        return diagnostics + [
            _diagnostic(
                "ruff.toml",
                "RUFF.IGNORES",
                "restore Git diff to evaluate suppression approval evidence",
            )
        ]
    candidates = set(changed.stdout.split("\0")) | set(untracked.stdout.split("\0"))
    for path in sorted(set(files) & candidates):
        old = _git(root, "show", f"{revision}:{path}")
        diagnostics.extend(
            _growth(
                path,
                _inline_suppressions(path, (root / path).read_text()),
                _inline_suppressions(path, old.stdout if old.returncode == 0 else ""),
                approvals,
            )
        )
    return diagnostics


def check(root: Path, files: list[str], policy: dict[str, Any]) -> list[str]:
    """Return actionable diagnostics for maintained Python tool policy violations."""
    maintained = [
        name for name in files if name.endswith(".py") and (root / name).is_file()
    ]
    return _coverage(root, maintained, policy) + _ruff(root, maintained, policy)

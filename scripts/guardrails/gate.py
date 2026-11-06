#!/usr/bin/env python3
"""Check-only repository gate shared by native hooks and CI."""

from __future__ import annotations

import json
import math
import os
import re
import subprocess
import sys
import tempfile
import time
from pathlib import Path
from typing import Any

ROOT = Path.cwd()
POLICY = ROOT / ".omp/guardrails.json"
ERRORS: list[str] = []


def error(path: str, rule: str, fix: str) -> None:
    ERRORS.append(f"{path}: {rule}: {fix}")


def git_environment() -> dict[str, str]:
    """Do not let hook repository selectors escape into isolated subprocesses."""
    return {
        key: value for key, value in os.environ.items() if not key.startswith("GIT_")
    }


def run(args: list[str], **kwargs: Any) -> subprocess.CompletedProcess[str]:
    kwargs.setdefault("check", True)
    kwargs.setdefault("env", git_environment())
    return subprocess.run(args, text=True, **kwargs)


def git(*args: str) -> str:
    return run(["git", *args], capture_output=True, env=dict(os.environ)).stdout


def paths() -> list[str]:
    output = git("ls-files", "-z")
    return [p for p in output.split("\0") if p]


def inspect(files: list[str]) -> list[dict[str, Any]]:
    if not files:
        return []
    with tempfile.TemporaryDirectory(prefix="scout-inspect-") as tmp:
        exe = str(Path(tmp) / "inspect")
        run(["go", "build", "-o", exe, "scripts/guardrails/inspect.go"])
        rows: list[dict[str, Any]] = json.loads(
            run([exe, *files], capture_output=True).stdout
        )
        return rows


def fanout(row: dict[str, Any]) -> None:
    # Lizard's lexer counts nonblank, noncomment physical source lines, not LOC.
    from lizard import analyze_file

    nloc = analyze_file(row["Path"]).nloc
    imports = {
        p
        for p in row["Imports"] or []
        if p.startswith("scout/") or "." in p.split("/")[0]
    }
    limit = 3 if nloc > 800 else 10
    if len(imports) > limit:
        rule = "FANOUT.large-module" if nloc > 800 else "FANOUT.general"
        error(
            row["Path"],
            rule,
            f"reduce {len(imports)} architectural imports to {limit} ({nloc} NLOC)",
        )


def check_row(row: dict[str, Any], policy: dict[str, Any]) -> None:
    owner = str(Path(row["Path"]).parent)
    if not re.fullmatch(r"[a-z][a-z0-9_]*", row["Package"]):
        error(row["Path"], "NAME.conventions", "use a lowercase Go package identifier")
    for symbol in row["Symbols"] or []:
        if not re.fullmatch(r"[A-Za-z][A-Za-z0-9_]*", symbol):
            error(
                row["Path"], "NAME.conventions", f"rename {symbol} to a Go identifier"
            )
    for dependency in row["Imports"] or []:
        if dependency.startswith("scout/"):
            target = dependency.removeprefix("scout/")
            if target not in policy["dependencies"].get(owner, []):
                error(
                    row["Path"],
                    "STRUCT.dependency-direction",
                    f"forbidden edge {owner} -> {target}; use an approved lower-layer contract",
                )
    fanout(row)


def relative_path(name: str) -> str | None:
    path = Path(name)
    if ".." not in path.parts and ".codegraph" not in path.parts:
        if not path.is_absolute():
            return name
        try:
            relative = path.resolve().relative_to(ROOT.resolve())
            if ".codegraph" not in relative.parts:
                return relative.as_posix()
        except ValueError:
            pass
    error(name, "STRUCT.folder-placement", "use an owned path within this repository")
    return None


def excluded_source(name: str) -> bool:
    return name.startswith("evaluation/") or bool(
        {".tools", ".git", "third_party"} & set(Path(name).parts)
    )


def path_check(files: list[str], policy: dict[str, Any]) -> None:
    gofiles = []
    for name in files:
        relative = relative_path(name)
        if relative is None:
            continue
        name = relative
        p = Path(name)
        if name.startswith("evaluation/"):
            continue
        if not any(re.fullmatch(pattern, name) for pattern in policy["owned_paths"]):
            error(
                name, "STRUCT.folder-placement", "place file in an approved owner path"
            )
        if excluded_source(name):
            continue
        if p.suffix == ".go" and p.exists():
            gofiles.append(name)
            if not re.fullmatch(r"[a-z][a-z0-9_]*\.go", p.name):
                error(
                    name,
                    "NAME.conventions",
                    "rename Go source to lowercase snake_case.go",
                )
            if name.endswith("_test.go") and not name.startswith(("internal/", "cmd/")):
                error(
                    name,
                    "TEST.placement",
                    "place fast Go tests beside their internal/ or cmd/ owner",
                )
    for row in inspect(gofiles):
        check_row(row, policy)


def valid_measurement(value: Any) -> bool:
    return type(value) in (int, float) and math.isfinite(value) and value >= 0


def budget(report_path: str | Path, limits: dict[str, Any]) -> None:
    try:
        report = json.loads(Path(report_path).read_text())
        started, wall = report["startedUnix"], report["fastTierWallClockMs"]
        if (
            not all(valid_measurement(v) for v in (started, wall))
            or not 0 <= time.time() - started <= 3600
        ):
            raise ValueError("invalid or stale measurement")
        tests = report["tests"]
        if not tests or not any(t["tier"] == "fast" for t in tests):
            raise ValueError("missing fast cases")
        for test in tests:
            duration = test["durationMs"]
            if not test["id"] or not valid_measurement(duration):
                raise ValueError("invalid case")
            if test["tier"] == "fast" and duration > limits["perTestMs"]:
                error(
                    test["id"],
                    "TEST.budget.per-test-time",
                    f"{duration}ms exceeds {limits['perTestMs']}ms; inject clock/runner or cut needless steps",
                )
        if wall > limits["wallMs"]:
            error(
                "go test",
                "TEST.budget.fast-tier-time",
                f"{wall}ms exceeds {limits['wallMs']}ms; cut needless steps",
            )
    except (OSError, ValueError, KeyError, TypeError) as exc:
        error(
            "go-test-report",
            "TEST.budget.fast-tier-time",
            f"record a fresh valid JSON measurement ({exc})",
        )


def materialize(destination: Path) -> None:
    """Borrow source objects, but give snapshots independent refs/config/index."""
    environment = git_environment()
    run(
        [
            "git",
            "-c",
            "core.hooksPath=/dev/null",
            "clone",
            "--quiet",
            "--shared",
            "--no-checkout",
            str(ROOT),
            str(destination),
        ],
        env=environment,
    )
    # One index transfer handles partial staging and split indexes without
    # one cat-file process per tracked file or any source repository writes.
    entries = git("ls-files", "--stage", "-z")
    run(["git", "read-tree", "--empty"], cwd=destination, env=environment)
    run(
        ["git", "update-index", "-z", "--index-info"],
        cwd=destination,
        env=environment,
        input=entries,
    )
    run(
        ["git", "checkout-index", "--all", "--force"],
        cwd=destination,
        env=environment,
    )
    (destination / ".tools").symlink_to(
        (ROOT / ".tools").resolve(), target_is_directory=True
    )


def snapshot_gate(files: list[str]) -> None:
    with tempfile.TemporaryDirectory(prefix="scout-gate-snapshot-") as tmp:
        destination = Path(tmp) / "checkout"
        materialize(destination)
        environment = git_environment()
        environment["SCOUT_GATE_SNAPSHOT"] = "1"
        environment["SCOUT_GATE_PATHS_STDIN"] = "1"
        # Use the installed runner while every checker reads the staged checkout.
        run(
            [sys.executable, str(ROOT / "scripts/guardrails/gate.py"), "commit"],
            cwd=destination,
            env=environment,
            input="\0".join(files),
        )


def resolve_base(base: str) -> str:
    result = run(
        ["git", "rev-parse", "--verify", f"{base}^{{commit}}"],
        check=False,
        capture_output=True,
    )
    if result.returncode:
        raise ValueError(f"{base}: fetch the configured test-growth base and retry")
    return result.stdout.strip()


def source_checks(files: list[str], mode: str, policy: dict[str, Any]) -> None:
    path_check(files, policy)
    gofiles = [
        p
        for p in files
        if p.endswith(".go") and not excluded_source(p) and Path(p).is_file()
    ]
    if gofiles:
        for name in run(
            ["gofmt", "-l", *gofiles], capture_output=True
        ).stdout.splitlines():
            error(name, "GO.format", "run gofmt")
    run([
        sys.executable,
        "scripts/guardrails/python/check.py",
        "commit" if mode == "commit" else "full",
        *(
            [p for p in files if p.endswith(".py") and not excluded_source(p)]
            if mode == "commit"
            else []
        ),
    ])
    run([
        sys.executable,
        "scripts/guardrails/infrastructure.py",
        mode,
        *(
            [
                p
                for p in files
                if Path(p).suffix in {".md", ".yml", ".yaml"} and not excluded_source(p)
            ]
            if mode == "commit"
            else []
        ),
    ])


def commit_gate() -> None:
    files = [
        p
        for p in git(
            "diff", "--cached", "--name-only", "--diff-filter=ACMRD", "-z"
        ).split("\0")
        if p
    ]
    if files:
        snapshot_gate(files)
    run(
        ["gitleaks", "git", "--pre-commit", "--staged", "--no-banner", "--redact", "."],
        env=dict(os.environ),
    )


def selected_paths() -> list[str]:
    if os.environ.get("SCOUT_GATE_PATHS_STDIN"):
        return [p for p in sys.stdin.read().split("\0") if p]
    return sys.argv[2:]


def main() -> bool:
    policy = json.loads(POLICY.read_text())
    policy["base"] = os.environ.get("SCOUT_GATE_BASE", policy["base"])
    mode = sys.argv[1] if len(sys.argv) > 1 else "ci"
    if mode not in {"ci", "commit", "path", "budget"}:
        raise ValueError("unknown mode: " + mode)
    if mode == "path":
        path_check(sys.argv[2:], policy)
    elif mode == "budget":
        budget(sys.argv[2], policy["budget"])
    elif mode == "commit" and not os.environ.get("SCOUT_GATE_SNAPSHOT"):
        commit_gate()
    else:
        policy["base"] = resolve_base(policy["base"])
        files = selected_paths() if mode == "commit" else paths()
        source_checks(files, mode, policy)
        run([
            sys.executable,
            str(Path(__file__).resolve().parent / "tests/check.py"),
            "full" if mode == "ci" else mode,
            "--base",
            policy["base"],
            "--",
            *(files if mode == "commit" else []),
        ])
        if mode == "ci":
            run(["go", "build", "./..."])
            run(["go", "vet", "./..."])
            run([
                sys.executable,
                str(Path(__file__).resolve().parent / "tests/output-canary.py"),
            ])
            run([sys.executable, "scripts/guardrails/merge-ledger.py", "--check"])
            run([
                sys.executable,
                "scripts/guardrails/vendor/validate_rule_coverage.py",
                "--root",
                ".",
                "--catalog",
                "scripts/guardrails/vendor/guardrail-rules.json",
                "--ledger",
                ".omp/coverage-ledger.json",
            ])
    for message in ERRORS:
        print(message)
    return bool(ERRORS)


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (subprocess.CalledProcessError, OSError, ValueError) as exc:
        print(
            f"gate: TEST.failure-propagation: repair failed prerequisite: {exc}",
            file=sys.stderr,
        )
        sys.exit(1)

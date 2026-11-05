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


def run(args: list[str], **kwargs: Any) -> subprocess.CompletedProcess[str]:
    return subprocess.run(args, check=True, text=True, **kwargs)


def git(*args: str) -> str:
    return run(["git", *args], capture_output=True).stdout


def paths() -> list[str]:
    revision = os.environ.get("SCOUT_GATE_REVISION")
    output = (
        git("ls-tree", "-r", "--name-only", "-z", revision)
        if revision
        else git("ls-files", "-z")
    )
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


def path_check(files: list[str], policy: dict[str, Any]) -> None:
    gofiles = []
    for name in files:
        p = Path(name)
        if p.is_absolute() or ".." in p.parts or ".codegraph" in p.parts:
            error(
                name, "STRUCT.folder-placement", "use an owned repository-relative path"
            )
            continue
        if name.startswith("evaluation/"):
            continue
        if not any(re.fullmatch(pattern, name) for pattern in policy["owned_paths"]):
            error(
                name, "STRUCT.folder-placement", "place file in an approved owner path"
            )
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


def materialize(destination: Path, revision: str | None) -> None:
    """Read blobs, never reset or otherwise mutate the contributor's index."""
    entries = (
        git("ls-files", "--stage", "-z")
        if revision is None
        else git("ls-tree", "-rz", revision)
    )
    for entry in entries.split("\0"):
        if not entry:
            continue
        metadata, name = entry.split("\t", 1)
        mode, *rest = metadata.split()
        if mode == "160000":
            continue
        oid = rest[0] if revision is None else rest[1]
        target = destination / name
        target.parent.mkdir(parents=True, exist_ok=True)
        content = subprocess.check_output(["git", "cat-file", "blob", oid])
        if mode == "120000":
            target.symlink_to(os.fsdecode(content))
        else:
            target.write_bytes(content)
            target.chmod(int(mode, 8) & 0o777)
    (destination / ".tools").symlink_to(ROOT / ".tools", target_is_directory=True)


def snapshot_gate(mode: str, revision: str | None, files: list[str]) -> None:
    with tempfile.TemporaryDirectory(prefix="scout-gate-snapshot-") as tmp:
        destination = Path(tmp)
        materialize(destination, revision)
        env = dict(
            os.environ,
            GIT_DIR=git("rev-parse", "--absolute-git-dir").strip(),
            GIT_WORK_TREE=tmp,
            SCOUT_GATE_SNAPSHOT="1",
            SCOUT_GATE_PATHS_STDIN="1",
        )
        if revision:
            env["SCOUT_GATE_REVISION"] = revision
        run(
            [sys.executable, "scripts/guardrails/gate.py", mode],
            cwd=destination,
            env=env,
            input="\0".join(files),
        )


def pushed_commits(base: str) -> list[str]:
    commits = set()
    for line in os.environ["SCOUT_PUSH_RANGES"].splitlines():
        fields = line.split()
        if len(fields) != 4:
            raise ValueError("pre-push requires four fields per ref update")
        local_sha = fields[1]
        if set(local_sha) == {"0"}:
            continue
        commits.update(
            git("rev-list", "--reverse", f"{base}..{local_sha}").splitlines()
        )
    return sorted(commits)


def source_checks(files: list[str], mode: str, policy: dict[str, Any]) -> None:
    path_check(files, policy)
    gofiles = [
        p
        for p in files
        if p.endswith(".go") and Path(p).is_file() and not p.startswith("evaluation/")
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
            [
                p
                for p in files
                if p.endswith(".py")
                and not p.startswith(("evaluation/", "third_party/"))
            ]
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
                if Path(p).suffix in {".md", ".yml", ".yaml"}
                and not p.startswith(("evaluation/", "third_party/"))
            ]
            if mode == "commit"
            else []
        ),
    ])


def commit_gate() -> None:
    files = [
        p
        for p in git(
            "diff", "--cached", "--name-only", "--diff-filter=ACMR", "-z"
        ).split("\0")
        if p
    ]
    if files:
        snapshot_gate("commit", None, files)
    run(["gitleaks", "git", "--pre-commit", "--staged", "--no-banner", "--redact", "."])


def selected_paths() -> list[str]:
    if os.environ.get("SCOUT_GATE_PATHS_STDIN"):
        return [p for p in sys.stdin.read().split("\0") if p]
    return sys.argv[2:]


def main() -> bool:
    policy = json.loads(POLICY.read_text())
    mode = sys.argv[1] if len(sys.argv) > 1 else "full"
    if mode not in {"full", "fast", "commit", "path", "budget"}:
        raise ValueError("use full, fast, commit, path or budget")
    if mode == "path":
        path_check(sys.argv[2:], policy)
    elif mode == "budget":
        budget(sys.argv[2], policy["budget"])
    elif mode == "commit" and not os.environ.get("SCOUT_GATE_SNAPSHOT"):
        commit_gate()
    elif (
        mode == "fast"
        and "SCOUT_PUSH_RANGES" in os.environ
        and not os.environ.get("SCOUT_GATE_SNAPSHOT")
    ):
        for revision in pushed_commits(policy["base"]):
            snapshot_gate(mode, revision, [])
    else:
        files = selected_paths() if mode == "commit" else paths()
        source_checks(files, mode, policy)
        if mode in ("full", "fast"):
            run([
                sys.executable,
                "scripts/guardrails/tests/check.py",
                mode,
                "--base",
                policy["base"],
            ])
        if mode == "full":
            run(["go", "build", "./..."])
            run(["go", "vet", "./..."])
            run([sys.executable, "scripts/guardrails/merge-ledger.py", "--check"])
            run([
                sys.executable,
                "scripts/guardrails/vendor/validate_rule_coverage.py",
                "--root",
                ".",
                "--catalog",
                "scripts/guardrails/vendor/enaible-rules.json",
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

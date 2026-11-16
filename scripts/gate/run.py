#!/usr/bin/env python3
"""Check-only repository gate shared by native hooks and CI.

Scout's own rules: owned-path, test-placement, layering and gofmt here;
toolchain-pin, tool-inventory, install-guidance, ci-parity, prettier-pin,
prettier-format and dependency-audit in infrastructure.py.
"""

from __future__ import annotations

import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path
from typing import Any

from inventory import excluded, repository_files

ROOT = Path.cwd()
POLICY = ROOT / "scripts/gate/structure.json"
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


def imports(files: list[str]) -> list[dict[str, Any]]:
    if not files:
        return []
    with tempfile.TemporaryDirectory(prefix="scout-imports-") as tmp:
        exe = str(Path(tmp) / "imports")
        run(["go", "build", "-o", exe, "scripts/gate/imports.go"])
        rows: list[dict[str, Any]] = json.loads(
            run([exe, *files], capture_output=True).stdout
        )
        return rows


def check_row(row: dict[str, Any], policy: dict[str, Any]) -> None:
    owner = str(Path(row["Path"]).parent)
    for dependency in row["Imports"] or []:
        if dependency.startswith("scout/"):
            target = dependency.removeprefix("scout/")
            if target not in policy["dependencies"].get(owner, []):
                error(
                    row["Path"],
                    "layering",
                    f"forbidden edge {owner} -> {target}; use an approved lower-layer contract",
                )


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
    error(name, "owned-path", "use an owned path within this repository")
    return None


def excluded_source(name: str) -> bool:
    return excluded(name)


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
            error(name, "owned-path", "place file in an approved owner path")
        if excluded_source(name):
            continue
        if p.suffix == ".go" and p.exists():
            gofiles.append(name)
            if name.endswith("_test.go") and not name.startswith(("internal/", "cmd/")):
                error(
                    name,
                    "test-placement",
                    "place fast Go tests beside their internal/ or cmd/ owner",
                )
    for row in imports(gofiles):
        check_row(row, policy)


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


def snapshot_gate(files: list[str], inventory: list[str]) -> None:
    with tempfile.TemporaryDirectory(prefix="scout-gate-snapshot-") as tmp:
        destination = Path(tmp) / "checkout"
        materialize(destination)
        # Preserve staged bytes, overlay all other maintained working-tree inputs.
        staged = set(files)
        for name in inventory:
            if name in staged:
                continue
            target = destination / name
            if not (ROOT / name).is_file():
                target.unlink(missing_ok=True)
                continue
            target.parent.mkdir(parents=True, exist_ok=True)
            if target.is_symlink():
                target.unlink()
            shutil.copy2(ROOT / name, target)
        environment = git_environment()
        environment["SCOUT_GATE_SNAPSHOT"] = "1"
        environment["SCOUT_GATE_PATHS_STDIN"] = "1"
        # Use the installed runner while every checker reads the staged checkout.
        run(
            [sys.executable, str(ROOT / "scripts/gate/run.py"), "commit"],
            cwd=destination,
            env=environment,
            input="\0".join(files),
        )


def source_checks(files: list[str], mode: str, policy: dict[str, Any]) -> None:
    # Deleted paths are inventory entries, but never checker inputs.
    path_check([p for p in files if Path(p).exists()], policy)
    gofiles = [
        p
        for p in files
        if p.endswith(".go") and not excluded_source(p) and Path(p).is_file()
    ]
    if gofiles:
        for name in run(
            ["gofmt", "-l", *gofiles], capture_output=True
        ).stdout.splitlines():
            error(name, "gofmt", "run gofmt")
    run([
        sys.executable,
        "scripts/gate/python.py",
        "commit" if mode == "commit" else "full",
    ])
    run([
        sys.executable,
        "scripts/gate/infrastructure.py",
        mode,
    ])


def commit_gate() -> None:
    files = [
        p
        for p in git(
            "diff", "--cached", "--name-only", "--diff-filter=ACMRD", "-z"
        ).split("\0")
        if p
    ]
    snapshot_gate(files, repository_files())
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
    mode = sys.argv[1] if len(sys.argv) > 1 else "ci"
    if mode not in {"ci", "commit", "path"}:
        raise ValueError("unknown mode: " + mode)
    if mode == "path":
        path_check(sys.argv[2:], policy)
    elif mode == "commit" and not os.environ.get("SCOUT_GATE_SNAPSHOT"):
        commit_gate()
    else:
        files = repository_files()
        source_checks(files, mode, policy)
        if ERRORS:
            for message in ERRORS:
                print(message)
            return True
        run([
            sys.executable,
            str(Path(__file__).resolve().parent / "tests.py"),
            "full" if mode == "ci" else mode,
            "--",
            *(selected_paths() if mode == "commit" else []),
        ])
        checker_tests = [
            name
            for name in files
            if name.startswith("scripts/gate/test_") and name.endswith(".py")
        ]
        checker_environment = {
            key: value
            for key, value in git_environment().items()
            if not key.startswith("SCOUT_GATE_")
        }
        for name in checker_tests:
            run([sys.executable, "-B", name], env=checker_environment)
        if mode == "ci":
            run(["go", "build", "./..."])
            run(["go", "vet", "./..."])
    for message in ERRORS:
        print(message)
    return bool(ERRORS)


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (subprocess.CalledProcessError, OSError, ValueError) as exc:
        print(
            f"gate: repair failed prerequisite: {exc}",
            file=sys.stderr,
        )
        sys.exit(1)

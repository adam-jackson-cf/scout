#!/usr/bin/env python3
"""Prove hook blocking, isolated snapshots, and pusher repository preservation."""

from __future__ import annotations

import importlib.util
import json
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path
from types import ModuleType
from typing import Any

ROOT = Path(__file__).resolve().parents[2]


def run(args: list[str], root: Path, **kwargs: Any) -> subprocess.CompletedProcess[str]:
    kwargs.setdefault("check", True)
    environment = {
        key: value for key, value in os.environ.items() if not key.startswith("GIT_")
    }
    return subprocess.run(
        args, cwd=root, text=True, capture_output=True, env=environment, **kwargs
    )


def git(root: Path, *args: str) -> str:
    return run(["git", "-c", "core.hooksPath=/dev/null", *args], root).stdout.strip()


def state(root: Path) -> dict[str, str]:
    commands = [
        ["for-each-ref"],
        ["config", "--list", "--local"],
        ["status", "--porcelain"],
        ["stash", "list"],
        ["diff", "--binary"],
        ["diff", "--cached", "--binary"],
    ]
    return {" ".join(command): git(root, *command) for command in commands}


def load_gate() -> ModuleType:
    spec = importlib.util.spec_from_file_location(
        "gate", ROOT / "scripts/guardrails/gate.py"
    )
    assert spec
    assert spec.loader
    gate = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(gate)
    return gate


def default_branch(gate: ModuleType, root: Path, base: str) -> None:
    remote = root.parent / "advertised.git"
    git(root, "init", "--bare", "-q", str(remote))
    git(root, "--git-dir", str(remote), "symbolic-ref", "HEAD", "refs/heads/main")
    git(root, "push", "--quiet", str(remote), base + ":refs/heads/main")
    previous = os.environ.get("SCOUT_PUSH_REMOTE_URL")
    os.environ["SCOUT_PUSH_REMOTE_URL"] = str(remote)
    try:
        assert gate.resolve_base("origin/missing-default") == base
        empty = root.parent / "empty.git"
        git(root, "init", "--bare", "-q", str(empty))
        os.environ["SCOUT_PUSH_REMOTE_URL"] = str(empty)
        try:
            gate.resolve_base("origin/missing-default")
        except SystemExit as exc:
            message = str(exc)
        else:
            raise AssertionError(
                "missing advertised default branch must fail with a fetch fix"
            )
        assert "GATE.pre-push" in message
        assert "git fetch" in message
    finally:
        if previous is None:
            os.environ.pop("SCOUT_PUSH_REMOTE_URL", None)
        else:
            os.environ["SCOUT_PUSH_REMOTE_URL"] = previous


def pre_gate_commit(gate: ModuleType, root: Path, base: str) -> None:
    git(root, "commit", "-qm", "commit predating repository gate")
    revision = git(root, "rev-parse", "HEAD")
    previous = os.environ.get("SCOUT_PUSH_RANGES")
    os.environ["SCOUT_PUSH_RANGES"] = (
        f"refs/heads/pre-gate {revision} refs/heads/pre-gate {'0' * 40}"
    )
    try:
        gate.push_gate(base)
    finally:
        if previous is None:
            os.environ.pop("SCOUT_PUSH_RANGES", None)
        else:
            os.environ["SCOUT_PUSH_RANGES"] = previous


def snapshots() -> None:
    gate = load_gate()
    with tempfile.TemporaryDirectory(prefix="scout-hook-snapshots-") as tmp:
        root = Path(tmp) / "repo"
        root.mkdir()
        git(root, "init", "-q")
        git(root, "config", "user.name", "Canary")
        git(root, "config", "user.email", "canary@example.invalid")
        path = root / "value.txt"
        path.write_text("staged-invalid")
        git(root, "add", "value.txt")
        git(root, "commit", "-qm", "base")
        base = git(root, "rev-parse", "HEAD")
        path.write_text("index-content")
        git(root, "add", "value.txt")
        git(root, "update-index", "--split-index")
        alternate = root / ".git/alternate-index"
        shutil.copy2(root / ".git/index", alternate)
        path.write_text("different-working-tree-content")
        git(root, "add", "value.txt")
        before = state(root)
        old = Path.cwd()
        previous_index = os.environ.get("GIT_INDEX_FILE")
        os.environ["GIT_INDEX_FILE"] = str(alternate)
        try:
            os.chdir(root)
            gate.__dict__["ROOT"] = root
            destination = Path(tmp) / "index"
            gate.materialize(destination, None)
            assert (destination / "value.txt").read_text() == "index-content"
            assert git(destination, "rev-parse", "--absolute-git-dir") != git(
                root, "rev-parse", "--absolute-git-dir"
            )
            assert state(root) == before
            default_branch(gate, root, base)
            pre_gate_commit(gate, root, base)
        finally:
            os.chdir(old)
            if previous_index is None:
                os.environ.pop("GIT_INDEX_FILE", None)
            else:
                os.environ["GIT_INDEX_FILE"] = previous_index


def fixture_repo(parent: Path) -> Path:
    repo = parent / "pusher"
    git(parent, "clone", "--quiet", "--shared", str(ROOT), str(repo))
    git(repo, "config", "user.name", "Canary")
    git(repo, "config", "user.email", "canary@example.invalid")
    shutil.copytree(
        ROOT / "scripts/guardrails",
        repo / "scripts/guardrails",
        dirs_exist_ok=True,
        ignore=shutil.ignore_patterns("__pycache__", "*.pyc"),
    )
    for name in [
        "ruff.toml",
        "mypy.ini",
        ".prettierrc.json",
        ".prettierignore",
        ".omp/gate.json",
        ".omp/guardrails.json",
        ".omp/test-map.json",
        ".omp/test-tiers.json",
        "cmd/scout/guardian_workflow_test.go",
    ]:
        shutil.copy2(ROOT / name, repo / name)
    (repo / ".tools").symlink_to((ROOT / ".tools").resolve(), target_is_directory=True)
    git(
        repo,
        "add",
        "scripts/guardrails",
        "ruff.toml",
        "mypy.ini",
        ".prettierrc.json",
        ".prettierignore",
        ".omp",
        "cmd/scout/guardian_workflow_test.go",
    )
    git(repo, "commit", "--allow-empty", "-qm", "isolated hook fixture runner")
    return repo


def commit_file(repo: Path, name: str, text: str, message: str) -> str:
    (repo / name).write_text(text)
    git(repo, "add", name)
    git(repo, "commit", "-qm", message)
    return git(repo, "rev-parse", "HEAD")


def failed_test(output: str, name: str) -> bool:
    for line in output.splitlines():
        if line.startswith("go test JSON: "):
            output += "\n" + Path(line.removeprefix("go test JSON: ")).read_text()
    for line in output.splitlines():
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            continue
        if (
            isinstance(event, dict)
            and event.get("Action") == "fail"
            and event.get("Package") == "scout/internal/policy"
            and event.get("Test") == name
        ):
            return True
    return False


def check_push(
    repo: Path, remote: Path, branch: str, *, failure: str | None = None
) -> None:
    before = state(repo)
    result = run(
        ["git", "push", str(remote), f"refs/heads/{branch}:refs/heads/{branch}"],
        repo,
        check=False,
    )
    after = state(repo)
    if before != after:
        changed = [key for key in before if before[key] != after[key]]
        raise AssertionError(f"push changed the pusher repository: {changed}")
    output = result.stdout + result.stderr
    if failure is None:
        if result.returncode:
            raise AssertionError("clean push was blocked:\n" + output)
    elif result.returncode == 0 or not failed_test(output, failure):
        raise AssertionError(
            f"non-HEAD push must fail specifically in {failure}:\n{output}"
        )
    print(
        f"push canary {branch}: {'passes' if failure is None else failure + ' fails'}; refs/config/status/stash/diffs unchanged"
    )


def push_commits(repo: Path, base: str) -> str:
    name = "scripts/ci-sandbox.sh"
    clean = commit_file(
        repo,
        name,
        (repo / name).read_text() + "\n# Clean script-only hook-push canary.\n",
        "change only a non-test script",
    )
    git(repo, "branch", "clean-push", clean)
    git(repo, "checkout", "--quiet", "--detach", base)
    name = "internal/policy/policy_test.go"
    text = (repo / name).read_text()
    assert 'if p.Profile() != "report" {' in text
    broken = commit_file(
        repo,
        name,
        text.replace('if p.Profile() != "report" {', 'if p.Profile() == "report" {', 1),
        "break TestAbsentIsReportNothingActive",
    )
    git(repo, "branch", "broken-push", broken)
    git(repo, "checkout", "--quiet", "--detach", clean)
    return clean


def preserve_pusher_inputs(repo: Path) -> None:
    name = "README.md"
    (repo / name).write_text(
        (repo / name).read_text() + "\nStashed input must survive push checks.\n"
    )
    git(repo, "stash", "push", "-qm", "hook preservation canary", "--", name)
    (repo / name).write_text(
        (repo / name).read_text() + "\nUnstaged input must survive push checks.\n"
    )
    (repo / "pusher-index.txt").write_text("staged input must survive")
    git(repo, "add", "pusher-index.txt")
    (repo / "pusher-untracked.txt").write_text("untracked input must survive")


def real_pushes() -> None:
    with tempfile.TemporaryDirectory(prefix="scout-real-push-") as tmp:
        parent = Path(tmp)
        repo = fixture_repo(parent)
        base = git(repo, "rev-parse", "HEAD")
        remote = parent / "remote.git"
        git(parent, "init", "--bare", "-q", str(remote))
        git(parent, "--git-dir", str(remote), "symbolic-ref", "HEAD", "refs/heads/main")
        git(repo, "push", "--quiet", str(remote), base + ":refs/heads/main")
        git(repo, "update-ref", "refs/remotes/origin/main", base)
        push_commits(repo, base)
        git(repo, "config", "core.hooksPath", ".githooks")
        preserve_pusher_inputs(repo)
        check_push(repo, remote, "clean-push")
        check_push(
            repo, remote, "broken-push", failure="TestAbsentIsReportNothingActive"
        )


def main() -> int:
    if sys.argv[1:] == ["pre-push", "--real-pushes"]:
        # Behavioural push proof: runs the fast test tier twice, so it is a
        # full-gate step, not part of the validator's 60 s install check.
        real_pushes()
        print("GATE.pre-push: real clean push passes, broken non-HEAD push blocked")
        return 0
    hook = sys.argv[1]
    field = {"pre-commit": "commitGate", "pre-push": "fastGate"}[hook]
    entry = ROOT / ".githooks" / hook
    if (
        not entry.is_file()
        or not entry.stat().st_mode & 0o111
        or "scripts/guardrails/hook.py" not in entry.read_text()
    ):
        return 1
    config = json.loads((ROOT / ".omp/gate.json").read_text())
    if not config.get(field):
        return 1
    for kind, expected in [("positive", 0), ("negative", 1)]:
        fixture = ROOT / "scripts/guardrails/fixtures" / f"GATE.{hook}.{kind}.json"
        result = run(
            [sys.executable, str(ROOT / "scripts/guardrails/canary.py"), str(fixture)],
            ROOT,
            check=False,
        )
        if (result.returncode != 0) != bool(expected):
            return 1
    snapshots()
    print(
        f"GATE.{hook}: installed hook blocking, independent staged checkout, and advertised default-branch selection proved"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())

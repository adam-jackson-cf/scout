#!/usr/bin/env python3
"""Check-only repository gate shared by native hooks and CI.

Scout's own rules: owned-path, test-placement, layering and gofmt here;
toolchain-pin, tool-inventory, install-guidance, ci-parity, prettier-pin,
prettier-format and dependency-audit in infrastructure.py.
"""

from __future__ import annotations

import ast
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


def dependency_edge(name: str, owner: str, target: str, policy: dict[str, Any]) -> None:
    directions = policy["dependencies"]
    if owner not in directions or target not in directions:
        error(
            name,
            "layering",
            f"unmapped edge {owner} -> {target}; declare both owners and their approved lower-layer direction in scripts/gate/structure.json",
        )
    elif target not in directions[owner]:
        error(
            name,
            "layering",
            f"forbidden edge {owner} -> {target}; use an approved lower-layer contract",
        )


def check_row(row: dict[str, Any], policy: dict[str, Any]) -> None:
    name = row["Path"]
    owner = str(Path(name).parent)
    if name == "scripts/gate/imports.go":
        owner = name
    for dependency in row["Imports"] or []:
        if dependency.startswith("scout/"):
            dependency_edge(name, owner, dependency.removeprefix("scout/"), policy)
    for symbol in row["Symbols"] or []:
        if symbol != "_" and not re.fullmatch(policy["naming"]["go_symbol"], symbol):
            error(name, "symbol-naming", f"rename {symbol} to {snake_case(symbol)}")
    package = row["Package"].removesuffix("_test")
    if name.endswith("_test.go"):
        expected = Path(owner).name
        if package != expected and not (owner == "cmd/scout" and package == "main"):
            destination = f"internal/{package}/{Path(name).name}"
            error(name, "test-placement", f"move test to its owner at {destination}")
    support = policy["test_layout"]["support_packages"].get(package)
    if support and str(Path(name).parent) != support:
        error(name, "test-placement", f"move support to {support}/{Path(name).name}")


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


def snake_case(value: str) -> str:
    value = re.sub(r"([A-Z]+)([A-Z][a-z])", r"\1_\2", value)
    value = re.sub(r"([a-z0-9])([A-Z])", r"\1_\2", value)
    return re.sub(r"[^a-zA-Z0-9_]+", "_", value).lower()


def owner_destination(name: str, policy: dict[str, Any]) -> str:
    path = Path(name)
    for parent in path.parents:
        if str(parent) in policy["dependencies"]:
            return str(parent / path.name)
    home = policy["destinations"].get(path.parts[0], "scripts/gate")
    return f"{home}/{path.name}"


def layout_check(name: str, policy: dict[str, Any]) -> bool:
    path = Path(name)
    layout = policy["test_layout"]
    parts = path.parts
    for folder in layout["fixture_folders"]:
        if folder in parts:
            index = parts.index(folder)
            if folder == layout["fixture_home"] and parts[0] == "internal":
                break
            remaining = parts[index + 1 :]
            if len(remaining) > 1:
                owner = f"internal/{remaining[0]}"
                destination = str(Path(owner, layout["fixture_home"], *remaining[1:]))
            else:
                destination = str(Path(path.parent, layout["fixture_home"], path.name))
            error(name, "test-placement", f"move fixture to {destination}")
            return False
    for folder in layout["support_folders"]:
        if folder in parts and not name.startswith(layout["support_home"] + "/"):
            error(
                name,
                "test-placement",
                f"move support to {layout['support_home']}/{path.name}",
            )
            return False
    if set(parts[:-1]).intersection(policy["banned_folders"]):
        error(name, "owned-path", f"move file to {owner_destination(name, policy)}")
        return False
    if not any(re.fullmatch(pattern, name) for pattern in policy["owned_paths"]):
        error(name, "owned-path", f"move file to {owner_destination(name, policy)}")
        return False
    return True


def test_name(name: str, policy: dict[str, Any]) -> bool:
    path = Path(name)
    naming = policy["naming"]
    layout = policy["test_layout"]
    is_test = (
        path.stem.startswith("test_")
        or path.stem.endswith("_test")
        or (path.suffix == ".py" and path.name == "test.py")
    )
    if not is_test or path.suffix not in {".go", ".py"}:
        return False
    stem = path.stem.removeprefix("test_").removesuffix("_test")
    stem = "_".join(
        part
        for part in stem.split("_")
        if not re.fullmatch(naming["test_tokens"], part)
    )
    stem = snake_case(stem) or "behaviour"
    correct = f"{stem}_test.go" if path.suffix == ".go" else f"test_{stem}.py"
    pattern = layout["go_pattern" if path.suffix == ".go" else "python_pattern"]
    if path.name != correct or not re.fullmatch(pattern, path.name):
        error(name, "test-naming", f"rename test to {path.with_name(correct)}")
    return True


def naming_check(name: str, policy: dict[str, Any]) -> bool:
    path = Path(name)
    is_test = test_name(name, policy)
    if is_test:
        return True
    naming = policy["naming"]
    pattern = naming["paths"].get(path.suffix)
    if pattern:
        if not re.fullmatch(pattern, path.stem):
            stem = snake_case(path.stem)
            if path.suffix == ".sh":
                stem = stem.replace("_", "-")
            error(name, "path-naming", f"rename file to {path.with_stem(stem)}")
    elif (
        path.name not in naming["standard_names"]
        and not path.name.startswith(".")
        and not re.fullmatch(naming["asset"], path.name)
    ):
        error(
            name, "path-naming", f"rename file to {path.with_name(path.name.lower())}"
        )
    return False


def test_owner(name: str, policy: dict[str, Any], tiers: dict[str, Any]) -> None:
    path = Path(name)
    owner = str(path.parent)
    if path.suffix == ".go":
        mapping = tiers.get("packages", {}).get(owner, {})
        fix = "scripts/gate/test-tiers.json"
    else:
        mapping = policy["test_layout"]["python_owners"].get(owner, {})
        fix = "scripts/gate/structure.json"
    if mapping.get("tier") not in {"fast", "seam"}:
        error(
            name,
            "test-placement",
            f"declare owner {owner} with exactly one fast or seam tier in {fix}; keep test at {owner}/{path.name}",
        )


def python_symbol(name: str, node: ast.AST, naming: dict[str, Any]) -> None:
    symbol = None
    kind = "python_variable"
    if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
        symbol, kind = node.name, "python_function"
        if symbol in naming["python_protocol_methods"]:
            return
    elif isinstance(node, ast.ClassDef):
        symbol, kind = node.name, "python_class"
    elif isinstance(node, ast.Name) and isinstance(node.ctx, ast.Store):
        symbol = node.id
    elif isinstance(node, ast.arg):
        symbol = node.arg
    if symbol and symbol != "_" and not re.fullmatch(naming[kind], symbol):
        correct = snake_case(symbol)
        if kind == "python_class":
            correct = "".join(part.title() for part in correct.split("_"))
        error(name, "symbol-naming", f"rename {symbol} to {correct}")


def python_imports(name: str, node: ast.AST) -> list[str]:
    if isinstance(node, ast.Import):
        return [alias.name for alias in node.names]
    if isinstance(node, ast.ImportFrom):
        prefix = node.module or ""
        if node.level:
            parent = Path(name).parent
            for _ in range(node.level - 1):
                parent = parent.parent
            prefix = ".".join((*parent.parts, prefix)).strip(".")
        return [prefix, *(f"{prefix}.{alias.name}" for alias in node.names)]
    return []


def python_edges(
    name: str, node: ast.AST, policy: dict[str, Any], modules: dict[str, str]
) -> None:
    dependencies = python_imports(name, node)
    for dependency in dependencies:
        local = str(Path(name).parent / (dependency.replace(".", "/") + ".py"))
        target = local if local in modules.values() else modules.get(dependency)
        if target:
            dependency_edge(name, name, target, policy)
        elif dependency.split(".")[0] not in sys.stdlib_module_names:
            # ImportFrom members are not independent imports.
            if isinstance(node, ast.ImportFrom) and dependency != dependencies[0]:
                continue
            dependency_edge(name, name, dependency, policy)


def python_check(name: str, policy: dict[str, Any], modules: dict[str, str]) -> None:
    try:
        tree = ast.parse(Path(name).read_text(), filename=name)
    except SyntaxError as exc:
        error(name, "python-syntax", f"repair Python syntax at line {exc.lineno}")
        return
    for node in ast.walk(tree):
        python_symbol(name, node, policy["naming"])
        python_edges(name, node, policy, modules)


def shell_check(name: str, policy: dict[str, Any]) -> None:
    for line in Path(name).read_text().splitlines():
        match = re.match(r"^\s*(?:source|\.)\s+(.+)", line)
        if not match:
            continue
        source = match[1].strip().strip("\"'")
        if source in policy["shell_sources"]["external"]:
            continue
        for prefix, destination in policy["shell_sources"]["prefixes"].items():
            if source.startswith(prefix):
                source = (
                    destination.format(parent=Path(name).parent) + source[len(prefix) :]
                )
                break
        dependency_edge(name, name, source, policy)


def content_check(name: str, policy: dict[str, Any], modules: dict[str, str]) -> bool:
    path = Path(name)
    if not path.is_file():
        return False
    if path.suffix == ".py":
        python_check(name, policy, modules)
    elif path.suffix == ".sh":
        shell_check(name, policy)
    return path.suffix == ".go"


def path_check(files: list[str], policy: dict[str, Any]) -> None:
    inventory = repository_files()
    modules = {
        str(Path(name).with_suffix("")).replace("/", "."): name
        for name in inventory
        if name.endswith(".py")
    }
    tiers = json.loads(Path("scripts/gate/test-tiers.json").read_text())
    gofiles = []
    for name in files:
        relative = relative_path(name)
        if relative is None:
            continue
        name = relative
        if excluded_source(name):
            continue
        if not layout_check(name, policy):
            continue
        if naming_check(name, policy):
            test_owner(name, policy, tiers)
        if content_check(name, policy, modules):
            gofiles.append(name)
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

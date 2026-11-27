from __future__ import annotations

import ast
import sys
from dataclasses import dataclass
from pathlib import PurePosixPath

from python_source import (
    Finding,
    Parsed,
    emit,
    integer,
    parse,
    read_request,
    strings,
    under,
)


@dataclass(frozen=True)
class Module:
    name: str | None
    is_package: bool


@dataclass(frozen=True)
class Limits:
    general: int
    large_nloc: int
    large: int


@dataclass(frozen=True)
class Graph:
    roots: frozenset[str]
    known: frozenset[str]


def module_for(path: str, roots: tuple[str, ...]) -> Module:
    for root in sorted(roots, key=len, reverse=True):
        if not under(path, (root,)):
            continue
        relative = PurePosixPath(path).relative_to(root)
        if relative.name == "__init__.py":
            parts = relative.parts[:-1]
            return Module(".".join(parts) or None, is_package=True)
        stem = relative.name.removesuffix(".pyi").removesuffix(".py")
        return Module(".".join((*relative.parts[:-1], stem)), is_package=False)
    return Module(None, is_package=False)


def build_graph(paths: list[str], roots: tuple[str, ...]) -> Graph:
    known: set[str] = set()
    for path in paths:
        name = module_for(path, roots).name
        if name is None:
            continue
        parts = name.split(".")
        known.update(".".join(parts[:end]) for end in range(1, len(parts) + 1))
    return Graph(
        roots=frozenset(name.split(".", 1)[0] for name in known), known=frozenset(known)
    )


def normalise(module: str, graph: Graph) -> str:
    root = module.split(".", 1)[0]
    if root not in graph.roots:
        return module
    parts = module.split(".")
    for end in range(len(parts), 0, -1):
        candidate = ".".join(parts[:end])
        if candidate in graph.known:
            return candidate
    return root


def resolve_relative(current: Module, node: ast.ImportFrom) -> str | None:
    if node.level == 0:
        return node.module
    if current.name is None:
        return None
    parts = (
        current.name.split(".") if current.is_package else current.name.split(".")[:-1]
    )
    if node.level > 1:
        parts = parts[: -(node.level - 1)]
    if node.module:
        parts.extend(node.module.split("."))
    return ".".join(parts)


def from_imports(node: ast.ImportFrom, current: Module, graph: Graph) -> set[str]:
    module = resolve_relative(current, node)
    if module is None:
        return set()
    if module.split(".", 1)[0] not in graph.roots:
        return {module}
    found: set[str] = set()
    for alias in node.names:
        candidate = module if alias.name == "*" else f"{module}.{alias.name}"
        found.add(normalise(candidate if candidate in graph.known else module, graph))
    return found


def literal_argument(node: ast.Call, position: int, keyword: str) -> str | None:
    argument = (
        node.args[position]
        if len(node.args) > position
        else next((item.value for item in node.keywords if item.arg == keyword), None)
    )
    if isinstance(argument, ast.Constant) and isinstance(argument.value, str):
        return argument.value
    return None


def relative_dynamic(name: str, package: str) -> str | None:
    level = len(name) - len(name.lstrip("."))
    parts = package.split(".")
    if not package or level > len(parts):
        return None
    base = parts[: len(parts) - level + 1]
    suffix = name[level:]
    return ".".join((*base, suffix)) if suffix else ".".join(base)


def dynamic_import(node: ast.AST) -> str | None:
    if not (
        isinstance(node, ast.Call)
        and isinstance(node.func, ast.Attribute)
        and node.func.attr == "import_module"
        and isinstance(node.func.value, ast.Name)
        and node.func.value.id == "importlib"
    ):
        return None
    name = literal_argument(node, 0, "name")
    if name is None or not name.startswith("."):
        return name
    package = literal_argument(node, 1, "package")
    return None if package is None else relative_dynamic(name, package)


def imported(parsed: Parsed, current: Module, graph: Graph) -> set[str]:
    found: set[str] = set()
    for node in ast.walk(parsed.tree):
        dynamic = dynamic_import(node)
        if dynamic is not None:
            found.add(normalise(dynamic, graph))
        elif isinstance(node, ast.Import):
            found.update(normalise(alias.name, graph) for alias in node.names)
        elif isinstance(node, ast.ImportFrom):
            found.update(from_imports(node, current, graph))
    if current.name is not None:
        found.discard(current.name)
    return found


def is_stdlib(module: str) -> bool:
    root = module.split(".", 1)[0]
    return root == "__future__" or root in sys.stdlib_module_names


def nloc(text: str) -> int:
    return sum(
        1
        for line in text.splitlines()
        if line.strip() and not line.lstrip().startswith("#")
    )


def fanout_findings(
    parsed: Parsed, dependencies: list[str], limits: Limits
) -> list[Finding]:
    listed = ", ".join(dependencies)
    count = len(dependencies)
    lines = nloc(parsed.text)
    findings: list[Finding] = []
    if count > limits.general:
        findings.append(
            Finding(
                parsed.path,
                1,
                "FANOUT.general",
                f"{count} architectural imports > {limits.general} ({listed})",
                "split the module by responsibility, or reach collaborators through fewer owned interfaces",
            )
        )
    if lines > limits.large_nloc and count > limits.large:
        findings.append(
            Finding(
                parsed.path,
                1,
                "FANOUT.large-module",
                f"large module of {lines} NLOC has {count} architectural imports > {limits.large} ({listed})",
                "split the module, or cut its architectural imports",
            )
        )
    return findings


def main() -> int:
    request = read_request()
    roots = strings(request.config, "maintained_roots")
    limits = Limits(
        general=integer(request.config, "fanout_max"),
        large_nloc=integer(request.config, "large_module_nloc"),
        large=integer(request.config, "large_module_fanout_max"),
    )
    graph = build_graph(request.modules, roots)
    findings: list[Finding] = []
    for item in request.files:
        parsed = parse(item, "FANOUT.general")
        if isinstance(parsed, Finding):
            findings.append(parsed)
            continue
        current = module_for(parsed.path, roots)
        dependencies = sorted(
            module
            for module in imported(parsed, current, graph)
            if not is_stdlib(module)
        )
        findings.extend(fanout_findings(parsed, dependencies, limits))
    return emit(findings)


if __name__ == "__main__":
    raise SystemExit(main())

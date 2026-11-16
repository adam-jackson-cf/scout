"""Native maintained-Python shape, ownership and effect-boundary checks."""

from __future__ import annotations

import ast
import fnmatch
import re
import sys
from pathlib import Path
from typing import Any

FUNCTIONS = (ast.FunctionDef, ast.AsyncFunctionDef)
PROCESSES = {
    "subprocess.run",
    "subprocess.Popen",
    "subprocess.call",
    "subprocess.check_call",
    "subprocess.check_output",
    "os.system",
    "os.popen",
    "asyncio.create_subprocess_exec",
    "asyncio.create_subprocess_shell",
}
DYNAMIC = {
    "__import__",
    "importlib.import_module",
    "importlib.util.spec_from_file_location",
    "importlib.util.module_from_spec",
}


def within(name: str, roots: list[str]) -> bool:
    return any(
        name == root.rstrip("/")
        or name.startswith(root.rstrip("/") + "/")
        or fnmatch.fnmatchcase(name, root)
        for root in roots
    )


def qualified(node: ast.AST, aliases: dict[str, str]) -> str:
    if isinstance(node, ast.Name):
        return aliases.get(node.id, node.id)
    if isinstance(node, ast.Attribute):
        return qualified(node.value, aliases) + "." + node.attr
    return ""


def imports(tree: ast.AST) -> dict[str, str]:
    names: dict[str, str] = {}
    for node in ast.walk(tree):
        if isinstance(node, ast.Import):
            for item in node.names:
                names[item.asname or item.name.split(".")[0]] = (
                    item.name if item.asname else item.name.split(".")[0]
                )
        elif isinstance(node, ast.ImportFrom):
            for item in node.names:
                names[item.asname or item.name] = (node.module or "") + "." + item.name
    return names


def path_findings(name: str, policy: dict[str, Any], banned: list[str]) -> list[str]:
    found: list[str] = []

    def reject(rule: str, fix: str) -> None:
        found.append(f"{name}: {rule}: {fix}")

    roots = policy["maintained_roots"]
    if not within(name, roots):
        reject("PYQ100", "declare the maintained root and package before adding Python")
        return found
    path = Path(name)
    script = within(name, policy.get("script_roots", []))
    if script and name not in policy.get("script_name_exceptions", []):
        if path.name not in policy.get("script_names", []):
            reject("PYQ101", "approve the command script name in python.script_names")
        if not re.fullmatch(r"[a-z][a-z0-9]*(?:-[a-z0-9]+)*\.py", path.name):
            reject("PYQ101", "rename the command script to kebab-case")
    groups = policy.get("test_groups", {})
    if (
        path.name.startswith("test_")
        and within(name, policy.get("test_roots", []))
        and not any(
            fnmatch.fnmatchcase(name, pattern)
            for patterns in groups.values()
            for pattern in patterns
        )
    ):
        reject("PYQ102", "map the test to a canonical test group")
    found.extend(directory_findings(name, policy, banned))
    if path.stem in banned and name not in policy.get("generic_name_allowlist", []):
        reject("PYQ104", "rename the catch-all module for its responsibility")
    if not script and not re.fullmatch(r"[a-z_][a-z0-9_]*", path.stem):
        reject("PYQ107", "rename the module to snake_case")
    return found


def directory_findings(
    name: str, policy: dict[str, Any], banned: list[str]
) -> list[str]:
    found: list[str] = []

    def reject(rule: str, fix: str) -> None:
        found.append(f"{name}: {rule}: {fix}")

    root = max(
        (root for root in policy["maintained_roots"] if within(name, [root])), key=len
    )
    parts = Path(name[len(root.rstrip("/")) :].lstrip("/")).parts
    allowed = policy.get("generic_name_allowlist", [])
    skill = within(name, policy.get("skill_roots", []))
    for index, part in enumerate(parts[:-1]):
        prefix = root.rstrip("/") + "/" + "/".join(parts[: index + 1])
        if part in banned and prefix not in allowed:
            reject("PYQ103", "rename the catch-all directory for its owner")
        if (
            not skill
            and part not in policy.get("non_package_dirs", [])
            and not re.fullmatch(r"[a-z_][a-z0-9_]*", part)
        ):
            reject("PYQ105", "rename the package directory to snake_case")
    for skill_root in policy.get("skill_roots", []):
        if within(name, [skill_root]):
            directory = name[len(skill_root.rstrip("/")) + 1 :].split("/")[0]
            if not re.fullmatch(r"[a-z][a-z0-9]*(?:-[a-z0-9]+)*", directory):
                reject("PYQ106", "rename the skill directory to kebab-case")
    return found


def class_findings(node: ast.ClassDef, policy: dict[str, Any]) -> list[tuple[str, str]]:
    limits = policy.get("class_limits", {})
    methods = [child for child in node.body if isinstance(child, FUNCTIONS)]
    attributes = {
        child.attr
        for child in ast.walk(node)
        if isinstance(child, ast.Attribute)
        and isinstance(child.ctx, ast.Store)
        and isinstance(child.value, ast.Name)
        and child.value.id == "self"
    }
    attributes.update(
        child.target.id
        for child in node.body
        if isinstance(child, ast.AnnAssign) and isinstance(child.target, ast.Name)
    )
    measures = [
        (
            "PYQ200",
            (node.end_lineno or node.lineno) - node.lineno + 1,
            limits.get("max_lines", 700),
            "split the class by responsibility",
        ),
        (
            "PYQ201",
            sum(not method.name.startswith("_") for method in methods),
            limits.get("max_public_methods"),
            "split the public interface",
        ),
        (
            "PYQ203",
            len(attributes),
            limits.get("max_instance_attributes", 12),
            "group state into a value object",
        ),
    ]
    for method in methods:
        if method.name == "__init__":
            args = [
                *method.args.posonlyargs,
                *method.args.args,
                *method.args.kwonlyargs,
            ]
            measures.append((
                "PYQ202",
                sum(arg.arg != "self" for arg in args),
                limits.get("max_init_args", 8),
                "introduce a parameter object",
            ))
    return [
        (rule, fix)
        for rule, count, limit, fix in measures
        if limit is not None and count > limit
    ]


class SyntaxPolicy(ast.NodeVisitor):
    """Inspect one file while retaining callable ownership and loop depth."""

    def __init__(self, name: str, tree: ast.Module, policy: dict[str, Any]) -> None:
        self.name = name
        self.policy = policy
        self.aliases = imports(tree)
        self.owner = ""
        self.depth = 0
        self.found: list[str] = []

    def reject(self, node: ast.AST, rule: str, fix: str) -> None:
        self.found.append(f"{self.name}:{getattr(node, 'lineno', 1)}: {rule}: {fix}")

    def visit_ClassDef(self, node: ast.ClassDef) -> None:
        for rule, fix in class_findings(node, self.policy):
            self.reject(node, rule, fix)
        bases = {ast.unparse(base).split(".")[-1] for base in node.bases}
        if (
            within(self.name, self.policy.get("analyzer_roots", []))
            and node.name.endswith("Analyzer")
            and not bases.intersection(self.policy.get("analyzer_bases", []))
        ):
            self.reject(node, "PYQ210", "inherit an approved analyzer base")
        methods = {child.name for child in node.body if isinstance(child, FUNCTIONS)}
        for base, entry in self.policy.get("analyzer_entrypoints", {}).items():
            if base in bases and entry not in methods:
                self.reject(node, "PYQ211", f"implement {entry} on the analyzer")
        previous = self.owner
        self.owner = f"{previous}.{node.name}".lstrip(".")
        self.generic_visit(node)
        self.owner = previous

    def visit_FunctionDef(self, node: ast.FunctionDef | ast.AsyncFunctionDef) -> None:
        previous, depth = self.owner, self.depth
        self.owner = f"{previous}.{node.name}".lstrip(".")
        self.depth = 0
        if self.name in self.policy.get("contract_files", []):
            self.contracts(node)
        self.generic_visit(node)
        self.owner, self.depth = previous, depth

    def visit_AsyncFunctionDef(self, node: ast.AsyncFunctionDef) -> None:
        self.visit_FunctionDef(node)

    def contracts(self, node: ast.FunctionDef | ast.AsyncFunctionDef) -> None:
        args = [*node.args.posonlyargs, *node.args.args, *node.args.kwonlyargs]
        for arg in args:
            if arg.annotation is None:
                continue
            text = ast.unparse(arg.annotation)
            if re.search(
                r"(?:\bCallable|collections\.abc\.Callable|typing\.Callable)\[", text
            ):
                self.reject(arg, "PYQ230", "use a named callable alias")
            if any(
                qualified(part, self.aliases) in {"typing.Any", "Any"}
                for part in ast.walk(arg.annotation)
            ):
                self.reject(
                    arg, "PYQ231", "use a concrete collaborator type or protocol"
                )

    def visit_Call(self, node: ast.Call) -> None:
        target = qualified(node.func, self.aliases)
        if (
            within(
                self.name,
                self.policy.get("process_roots", self.policy["maintained_roots"]),
            )
            and target in PROCESSES
            and self.owner
            not in self.policy.get("subprocess_wrappers", {}).get(self.name, [])
        ):
            self.reject(
                node, "PYQ220", "move process creation into an approved wrapper"
            )
        if (
            within(
                self.name,
                self.policy.get(
                    "dynamic_import_roots", self.policy["maintained_roots"]
                ),
            )
            and target in DYNAMIC
            and self.name not in self.policy.get("dynamic_import_adapters", [])
        ):
            self.reject(
                node, "PYQ221", "use a static import or approved dynamic import adapter"
            )
        if within(self.name, self.policy.get("analyzer_like_roots", [])):
            if target in {"open", "io.open"} or target.endswith((
                ".read_text",
                ".read_bytes",
                ".write_text",
                ".write_bytes",
            )):
                self.reject(
                    node, "PYS251", "inject filesystem access through an owned seam"
                )
            if target in PROCESSES or target in {
                "os.getenv",
                "os.environ.get",
                "os.environ.copy",
            }:
                self.reject(
                    node,
                    "PYS252",
                    "inject process or environment access through an owned seam",
                )
        self.generic_visit(node)

    def visit_Subscript(self, node: ast.Subscript) -> None:
        if (
            within(self.name, self.policy.get("analyzer_like_roots", []))
            and qualified(node.value, self.aliases) == "os.environ"
        ):
            self.reject(
                node, "PYS252", "inject environment access through an owned seam"
            )
        self.generic_visit(node)

    def visit_For(self, node: ast.For | ast.AsyncFor | ast.While) -> None:
        self.depth += 1
        if self.depth > self.policy.get("max_loop_depth", 2):
            self.reject(node, "PYS250", "extract a loop into a focused callable")
        self.generic_visit(node)
        self.depth -= 1

    def visit_AsyncFor(self, node: ast.AsyncFor) -> None:
        self.visit_For(node)

    def visit_While(self, node: ast.While) -> None:
        self.visit_For(node)

    def visit_ListComp(
        self, node: ast.ListComp | ast.SetComp | ast.DictComp | ast.GeneratorExp
    ) -> None:
        self.depth += len(node.generators)
        if self.depth > self.policy.get("max_loop_depth", 2):
            self.reject(
                node, "PYS250", "extract nested comprehension work into a callable"
            )
        self.generic_visit(node)
        self.depth -= len(node.generators)

    def visit_SetComp(self, node: ast.SetComp) -> None:
        self.visit_ListComp(node)

    def visit_DictComp(self, node: ast.DictComp) -> None:
        self.visit_ListComp(node)

    def visit_GeneratorExp(self, node: ast.GeneratorExp) -> None:
        self.visit_ListComp(node)


def architectural_imports(tree: ast.Module) -> set[str]:
    edges: set[str] = set()
    for node in ast.walk(tree):
        if isinstance(node, ast.Import):
            edges.update(item.name.split(".")[0] for item in node.names)
        elif isinstance(node, ast.ImportFrom):
            edges.add((node.module or ".").split(".")[0] or ".")
    return edges.difference(sys.stdlib_module_names)


def check(
    root: Path, files: list[str], policy: dict[str, Any], banned: list[str]
) -> list[str]:
    found: list[str] = []
    if policy.get("analyzer_like_roots") and not any(
        (root / marker).is_file() for marker in policy.get("ownership_markers", [])
    ):
        found.append(
            "scripts/gate/structure.json: PYS253: declare an existing analyzer owner marker"
        )
    for name in files:
        found.extend(path_findings(name, policy, banned))
        try:
            source = (root / name).read_text()
            tree = ast.parse(source, filename=name)
        except (SyntaxError, UnicodeError) as error:
            found.extend(
                f"{name}: {rule}: restore valid Python syntax ({error})"
                for rule in ("PYQ299", "PYS000")
            )
            continue
        visitor = SyntaxPolicy(name, tree, policy)
        visitor.visit(tree)
        found.extend(visitor.found)
        edges = architectural_imports(tree)
        nloc = sum(
            bool(line.strip()) and not line.lstrip().startswith("#")
            for line in source.splitlines()
        )
        if len(edges) > 10:
            found.append(
                f"{name}: FANOUT.general: split the module to reduce {len(edges)} architectural imports to 10"
            )
        if nloc > 800 and len(edges) > 3:
            found.append(
                f"{name}: FANOUT.large-module: split the large module to reduce {len(edges)} architectural imports to 3"
            )
    return found

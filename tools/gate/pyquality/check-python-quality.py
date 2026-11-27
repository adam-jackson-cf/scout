from __future__ import annotations

import ast
import re
from dataclasses import dataclass
from pathlib import PurePosixPath

from python_source import (
    Finding,
    Parsed,
    dotted_name,
    emit,
    integer,
    mapping,
    parse,
    read_request,
    strings,
    under,
)

BANNED_WORDS = frozenset({
    "common",
    "compat",
    "fallback",
    "helpers",
    "legacy",
    "manager",
    "misc",
    "new",
    "old",
    "stuff",
    "temp",
    "utils",
})
SNAKE_CASE = re.compile(r"^_*[a-z][a-z0-9_]*$")
KEBAB_CASE = re.compile(r"^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$")
TEST_MODULE = re.compile(r"^(?:test_.+|.+_test|conftest)$")
SUBPROCESS_CALLS = frozenset({
    "subprocess.run",
    "subprocess.Popen",
    "subprocess.call",
    "subprocess.check_call",
    "subprocess.check_output",
    "subprocess.getoutput",
    "subprocess.getstatusoutput",
    "os.system",
    "os.popen",
    "asyncio.create_subprocess_exec",
    "asyncio.create_subprocess_shell",
})
SUBPROCESS_PREFIXES = ("os.exec", "os.spawn", "os.posix_spawn")
DYNAMIC_IMPORT_CALLS = frozenset({
    "__import__",
    "importlib.__import__",
    "importlib.import_module",
    "importlib.util.spec_from_file_location",
    "importlib.util.spec_from_loader",
    "importlib.util.module_from_spec",
})
FunctionNode = ast.FunctionDef | ast.AsyncFunctionDef


@dataclass(frozen=True)
class PathPolicy:
    maintained_roots: tuple[str, ...]
    script_roots: tuple[str, ...]
    script_names: frozenset[str]
    test_roots: tuple[str, ...]
    test_groups: frozenset[str]
    generic_name_allowlist: frozenset[str]


@dataclass(frozen=True)
class SourcePolicy:
    class_max_lines: int
    class_max_init_args: int
    class_max_instance_attributes: int
    subprocess_wrappers: dict[str, frozenset[str]]
    dynamic_import_adapters: frozenset[str]
    contract_files: frozenset[str]


def path_policy(config: dict[str, object]) -> PathPolicy:
    maintained = strings(config, "maintained_roots")
    if not maintained:
        raise SystemExit(
            "proofs.toml [python_quality].maintained_roots must name every maintained root"
        )
    return PathPolicy(
        maintained_roots=maintained,
        script_roots=strings(config, "script_roots"),
        script_names=frozenset(strings(config, "script_names")),
        test_roots=strings(config, "test_roots"),
        test_groups=frozenset(strings(config, "test_groups")),
        generic_name_allowlist=frozenset(strings(config, "generic_name_allowlist")),
    )


def source_policy(config: dict[str, object]) -> SourcePolicy:
    return SourcePolicy(
        class_max_lines=integer(config, "class_max_lines"),
        class_max_init_args=integer(config, "class_max_init_args"),
        class_max_instance_attributes=integer(config, "class_max_instance_attributes"),
        subprocess_wrappers=mapping(config, "subprocess_wrappers"),
        dynamic_import_adapters=frozenset(strings(config, "dynamic_import_adapters")),
        contract_files=frozenset(strings(config, "contract_files")),
    )


def module_stem(path: str) -> str:
    return PurePosixPath(path).name.removesuffix(".pyi").removesuffix(".py")


def snake_name(stem: str) -> str:
    return re.sub(r"(?<=[a-z0-9])(?=[A-Z])", "_", stem).replace("-", "_").lower()


def is_main_guard(statement: ast.stmt) -> bool:
    if not isinstance(statement, ast.If) or not isinstance(statement.test, ast.Compare):
        return False
    test = statement.test
    return (
        isinstance(test.left, ast.Name)
        and test.left.id == "__name__"
        and len(test.comparators) == 1
        and isinstance(test.comparators[0], ast.Constant)
        and test.comparators[0].value == "__main__"
    )


def is_command(parsed: Parsed) -> bool:
    return parsed.text.startswith("#!") or any(
        is_main_guard(statement) for statement in parsed.tree.body
    )


def generic_name_findings(path: str, policy: PathPolicy) -> list[Finding]:
    parts = PurePosixPath(path).parts
    findings: list[Finding] = []
    for index, part in enumerate(parts[:-1]):
        prefix = "/".join(parts[: index + 1])
        if part in BANNED_WORDS and prefix not in policy.generic_name_allowlist:
            findings.append(
                Finding(
                    path,
                    1,
                    "PYQ103",
                    f"generic directory name '{part}'",
                    f"rename {prefix} for what it holds, or move its files under their owner",
                )
            )
    stem = module_stem(path)
    if stem in BANNED_WORDS and path not in policy.generic_name_allowlist:
        findings.append(
            Finding(
                path,
                1,
                "PYQ104",
                f"generic module name '{stem}'",
                "rename the module for what it holds, or move its code to its owner",
            )
        )
    return findings


def script_findings(path: str, command: bool, policy: PathPolicy) -> list[Finding]:
    name = PurePosixPath(path).name
    stem = module_stem(path)
    findings: list[Finding] = []
    if name not in policy.script_names:
        findings.append(
            Finding(
                path,
                1,
                "PYQ101",
                f"script '{name}' is not an approved script name",
                "add it to proofs.toml [python_quality].script_names with owner approval, or remove it",
            )
        )
    if command and not KEBAB_CASE.match(stem):
        findings.append(
            Finding(
                path,
                1,
                "PYQ101",
                f"command script '{name}' is not kebab-case",
                f"rename it to {stem.replace('_', '-').lower()}.py and update its callers",
            )
        )
    if not command and not SNAKE_CASE.match(stem):
        findings.append(snake_finding(path))
    return findings


def snake_finding(path: str) -> Finding:
    name = PurePosixPath(path).name
    return Finding(
        path,
        1,
        "PYQ107",
        f"Python module '{name}' is not snake_case",
        f"rename it to {snake_name(module_stem(path))}.py and update its imports",
    )


def test_group_findings(path: str, policy: PathPolicy) -> list[Finding]:
    root = next((root for root in policy.test_roots if under(path, (root,))), None)
    groups = ", ".join(sorted(policy.test_groups)) or "none declared"
    if root is None:
        return [
            Finding(
                path,
                1,
                "PYQ102",
                "test module is outside the declared test roots",
                "move it under a proofs.toml [python_quality].test_roots entry",
            )
        ]
    parts = PurePosixPath(path).relative_to(root).parts
    if len(parts) > 1 and parts[0] in policy.test_groups:
        return []
    return [
        Finding(
            path,
            1,
            "PYQ102",
            "test module is not inside a canonical test group",
            f"move it into a group folder under {root} ({groups})",
        )
    ]


def module_findings(path: str, command: bool, policy: PathPolicy) -> list[Finding]:
    findings: list[Finding] = []
    if command:
        findings.append(
            Finding(
                path,
                1,
                "PYQ101",
                f"command script '{PurePosixPath(path).name}' is outside the script roots",
                "move it under a proofs.toml [python_quality].script_roots entry, or drop its entry point",
            )
        )
    if not SNAKE_CASE.match(module_stem(path)):
        findings.append(snake_finding(path))
    if TEST_MODULE.match(module_stem(path)):
        findings.extend(test_group_findings(path, policy))
    return findings


def path_findings(path: str, command: bool, policy: PathPolicy) -> list[Finding]:
    if not under(path, policy.maintained_roots):
        return [
            Finding(
                path,
                1,
                "PYQ100",
                "Python file is outside the maintained roots",
                "move it under a maintained root, or add its root to proofs.toml [python_quality].maintained_roots",
            )
        ]
    findings = generic_name_findings(path, policy)
    if under(path, policy.script_roots):
        return findings + script_findings(path, command, policy)
    return findings + module_findings(path, command, policy)


def last_name(node: ast.expr) -> str:
    if isinstance(node, ast.Call):
        return last_name(node.func)
    name = dotted_name(node)
    return name.rsplit(".", 1)[-1] if name else ""


def annotation_names(annotation: ast.expr | None) -> set[str]:
    names: set[str] = set()
    pending: list[ast.expr] = [annotation] if annotation is not None else []
    while pending:
        node = pending.pop()
        if isinstance(node, ast.Name):
            names.add(node.id)
        elif isinstance(node, ast.Attribute):
            names.add(node.attr)
        elif isinstance(node, ast.Subscript):
            pending.extend(subscript_children(node))
        elif isinstance(node, ast.BinOp):
            pending.extend([node.left, node.right])
        elif isinstance(node, ast.Tuple | ast.List):
            pending.extend(node.elts)
        elif isinstance(node, ast.Constant) and isinstance(node.value, str):
            pending.extend(string_annotation(node.value, names))
    return names


def string_annotation(text: str, names: set[str]) -> list[ast.expr]:
    try:
        return [ast.parse(text, mode="eval").body]
    except SyntaxError:
        names.add(text)
        return []


def subscript_children(node: ast.Subscript) -> list[ast.expr]:
    outer = last_name(node.value)
    if outer == "Literal":
        return [node.value]
    if outer == "Annotated" and isinstance(node.slice, ast.Tuple):
        return [node.value, node.slice.elts[0]]
    return [node.value, node.slice]


def dataclass_fields(node: ast.ClassDef) -> list[str]:
    return [
        statement.target.id
        for statement in node.body
        if isinstance(statement, ast.AnnAssign)
        and isinstance(statement.target, ast.Name)
        and "ClassVar" not in annotation_names(statement.annotation)
    ]


def is_dataclass(node: ast.ClassDef) -> bool:
    return any(last_name(decorator) == "dataclass" for decorator in node.decorator_list)


def constructor_parameters(node: ast.ClassDef) -> int:
    for statement in node.body:
        if isinstance(statement, FunctionNode) and statement.name == "__init__":
            arguments = [*statement.args.posonlyargs, *statement.args.args]
            if arguments and arguments[0].arg == "self":
                arguments = arguments[1:]
            return len(arguments) + len(statement.args.kwonlyargs)
    return len(dataclass_fields(node)) if is_dataclass(node) else 0


def instance_attributes(node: ast.ClassDef) -> set[str]:
    attributes = set(dataclass_fields(node))
    pending: list[ast.AST] = list(node.body)
    while pending:
        current = pending.pop()
        if isinstance(current, ast.ClassDef):
            continue
        if (
            isinstance(current, ast.Attribute)
            and isinstance(current.ctx, ast.Store)
            and isinstance(current.value, ast.Name)
            and current.value.id == "self"
        ):
            attributes.add(current.attr)
        pending.extend(ast.iter_child_nodes(current))
    return attributes


def class_findings(
    node: ast.ClassDef, path: str, policy: SourcePolicy
) -> list[Finding]:
    findings: list[Finding] = []
    lines = (node.end_lineno or node.lineno) - node.lineno + 1
    if lines > policy.class_max_lines:
        findings.append(
            Finding(
                path,
                node.lineno,
                "PYQ200",
                f"class '{node.name}' has {lines} lines > {policy.class_max_lines}",
                "split the class by responsibility into smaller collaborators",
            )
        )
    parameters = constructor_parameters(node)
    if parameters > policy.class_max_init_args:
        findings.append(
            Finding(
                path,
                node.lineno,
                "PYQ202",
                f"constructor of '{node.name}' takes {parameters} parameters > {policy.class_max_init_args}",
                "group related parameters into a value object, or split the class",
            )
        )
    attributes = len(instance_attributes(node))
    if attributes > policy.class_max_instance_attributes:
        findings.append(
            Finding(
                path,
                node.lineno,
                "PYQ203",
                f"class '{node.name}' holds {attributes} instance attributes > {policy.class_max_instance_attributes}",
                "move related attributes into a component object",
            )
        )
    return findings


def import_aliases(tree: ast.Module) -> dict[str, str]:
    aliases: dict[str, str] = {"__import__": "__import__"}
    for node in ast.walk(tree):
        if isinstance(node, ast.Import):
            for alias in node.names:
                head = alias.name.split(".", 1)[0]
                aliases[alias.asname or head] = alias.name if alias.asname else head
        elif isinstance(node, ast.ImportFrom) and node.module and node.level == 0:
            for alias in node.names:
                aliases[alias.asname or alias.name] = f"{node.module}.{alias.name}"
    return aliases


def resolved_call(node: ast.Call, aliases: dict[str, str]) -> str | None:
    name = dotted_name(node.func)
    if name is None:
        return None
    head, _, rest = name.partition(".")
    target = aliases.get(head)
    if target is None:
        return None
    return f"{target}.{rest}" if rest else target


def calls_with_owner(tree: ast.Module) -> list[tuple[ast.Call, str | None]]:
    found: list[tuple[ast.Call, str | None]] = []
    pending: list[tuple[ast.AST, str | None]] = [(tree, None)]
    while pending:
        node, owner = pending.pop()
        if isinstance(node, ast.Call):
            found.append((node, owner))
        inner = node.name if isinstance(node, FunctionNode) else owner
        pending.extend((child, inner) for child in ast.iter_child_nodes(node))
    return found


def is_subprocess(name: str) -> bool:
    return name in SUBPROCESS_CALLS or name.startswith(SUBPROCESS_PREFIXES)


def boundary_findings(parsed: Parsed, policy: SourcePolicy) -> list[Finding]:
    aliases = import_aliases(parsed.tree)
    wrappers = policy.subprocess_wrappers.get(parsed.path, frozenset())
    adapter = parsed.path in policy.dynamic_import_adapters
    findings: list[Finding] = []
    for node, owner in calls_with_owner(parsed.tree):
        name = resolved_call(node, aliases)
        if name is None:
            continue
        if is_subprocess(name) and owner not in wrappers:
            findings.append(
                Finding(
                    parsed.path,
                    node.lineno,
                    "PYQ220",
                    f"subprocess call {name} outside an approved wrapper",
                    "call it from a wrapper function listed in proofs.toml [python_quality].subprocess_wrappers",
                )
            )
        if name in DYNAMIC_IMPORT_CALLS and not adapter:
            findings.append(
                Finding(
                    parsed.path,
                    node.lineno,
                    "PYQ221",
                    f"dynamic import {name} outside an adapter module",
                    "use a static import, or move the loading into a module listed in [python_quality].dynamic_import_adapters",
                )
            )
    return findings


def contract_annotations(tree: ast.Module) -> list[tuple[int, ast.expr]]:
    found: list[tuple[int, ast.expr]] = []
    for node in ast.walk(tree):
        if isinstance(node, ast.AnnAssign):
            found.append((node.lineno, node.annotation))
        elif isinstance(node, FunctionNode):
            arguments = node.args
            every = [*arguments.posonlyargs, *arguments.args, *arguments.kwonlyargs]
            every.extend(
                arg for arg in (arguments.vararg, arguments.kwarg) if arg is not None
            )
            found.extend(
                (arg.lineno, arg.annotation)
                for arg in every
                if arg.annotation is not None
            )
    return found


def contract_findings(parsed: Parsed, policy: SourcePolicy) -> list[Finding]:
    if parsed.path not in policy.contract_files:
        return []
    findings: list[Finding] = []
    for line, annotation in contract_annotations(parsed.tree):
        names = annotation_names(annotation)
        if "Callable" in names:
            findings.append(
                Finding(
                    parsed.path,
                    line,
                    "PYQ230",
                    "contract annotation uses an inline Callable",
                    "annotate with a named alias (Handler = Callable[...]) or a Protocol",
                )
            )
        if "Any" in names:
            findings.append(
                Finding(
                    parsed.path,
                    line,
                    "PYQ231",
                    "contract annotation uses Any as a collaborator boundary",
                    "replace Any with an owned Protocol, TypedDict or concrete type",
                )
            )
    return findings


def source_findings(parsed: Parsed, policy: SourcePolicy) -> list[Finding]:
    findings = boundary_findings(parsed, policy) + contract_findings(parsed, policy)
    for node in ast.walk(parsed.tree):
        if isinstance(node, ast.ClassDef):
            findings.extend(class_findings(node, parsed.path, policy))
    return findings


def main() -> int:
    request = read_request()
    paths = path_policy(request.config)
    sources = source_policy(request.config)
    findings: list[Finding] = []
    for item in request.files:
        parsed = parse(item, "PYQ299")
        if isinstance(parsed, Finding):
            findings.append(parsed)
            findings.extend(path_findings(item["path"], command=False, policy=paths))
            continue
        findings.extend(path_findings(parsed.path, is_command(parsed), paths))
        if under(parsed.path, paths.maintained_roots):
            findings.extend(source_findings(parsed, sources))
    return emit(findings)


if __name__ == "__main__":
    raise SystemExit(main())

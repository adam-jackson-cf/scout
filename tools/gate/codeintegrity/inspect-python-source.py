from __future__ import annotations

import ast
import base64
import io
import json
import sys
import tokenize
from typing import TypedDict

DECLARATION_DECORATORS = frozenset({
    "abstractmethod",
    "abstractproperty",
    "abstractclassmethod",
    "abstractstaticmethod",
    "overload",
})
FunctionNode = ast.FunctionDef | ast.AsyncFunctionDef
DocumentedNode = ast.Module | ast.ClassDef | ast.FunctionDef | ast.AsyncFunctionDef


class Span(TypedDict, total=False):
    line: int
    end_line: int
    text: str
    name: str
    module: str


class Report(TypedDict):
    error: str
    error_line: int
    comments: list[Span]
    docstrings: list[Span]
    placeholders: list[Span]
    imports: list[Span]


def comment_spans(data: bytes) -> list[Span]:
    return [
        Span(line=token.start[0], end_line=token.end[0], text=token.string)
        for token in tokenize.tokenize(io.BytesIO(data).readline)
        if token.type == tokenize.COMMENT
    ]


def docstring(node: DocumentedNode) -> ast.Expr | None:
    first = node.body[0] if node.body else None
    if (
        isinstance(first, ast.Expr)
        and isinstance(first.value, ast.Constant)
        and isinstance(first.value.value, str)
    ):
        return first
    return None


def docstring_spans(tree: ast.Module) -> list[Span]:
    spans: list[Span] = []
    for node in ast.walk(tree):
        if not isinstance(node, DocumentedNode):
            continue
        found = docstring(node)
        if found is not None:
            spans.append(
                Span(line=found.lineno, end_line=found.end_lineno or found.lineno)
            )
    return spans


def last_name(node: ast.expr) -> str:
    if isinstance(node, ast.Call):
        return last_name(node.func)
    if isinstance(node, ast.Subscript):
        return last_name(node.value)
    if isinstance(node, ast.Attribute):
        return node.attr
    if isinstance(node, ast.Name):
        return node.id
    return ""


def is_protocol(node: ast.ClassDef) -> bool:
    return any(last_name(base) == "Protocol" for base in node.bases)


def is_placeholder_statement(statement: ast.stmt) -> bool:
    if isinstance(statement, ast.Pass):
        return True
    if isinstance(statement, ast.Expr):
        return (
            isinstance(statement.value, ast.Constant)
            and statement.value.value is Ellipsis
        )
    if isinstance(statement, ast.Raise) and statement.exc is not None:
        return last_name(statement.exc) == "NotImplementedError"
    return False


def placeholder(node: FunctionNode, in_protocol: bool) -> Span | None:
    if in_protocol or any(
        last_name(decorator) in DECLARATION_DECORATORS
        for decorator in node.decorator_list
    ):
        return None
    body = node.body[1:] if docstring(node) is not None else node.body
    if len(body) != 1 or not is_placeholder_statement(body[0]):
        return None
    statement = body[0]
    return Span(
        line=statement.lineno,
        end_line=statement.end_lineno or statement.lineno,
        name=node.name,
    )


def placeholder_spans(tree: ast.Module) -> list[Span]:
    spans: list[Span] = []
    pending: list[tuple[ast.AST, bool]] = [(tree, False)]
    while pending:
        node, in_protocol = pending.pop()
        for child in ast.iter_child_nodes(node):
            if isinstance(child, ast.ClassDef):
                pending.append((child, is_protocol(child)))
                continue
            if isinstance(child, FunctionNode):
                found = placeholder(child, in_protocol)
                if found is not None:
                    spans.append(found)
                in_protocol_child = False
            else:
                in_protocol_child = in_protocol
            pending.append((child, in_protocol_child))
    return spans


def import_spans(tree: ast.Module) -> list[Span]:
    spans: list[Span] = []
    for node in ast.walk(tree):
        end = node.end_lineno if isinstance(node, ast.stmt) else None
        if isinstance(node, ast.Import):
            spans.extend(
                Span(line=node.lineno, end_line=end or node.lineno, module=alias.name)
                for alias in node.names
            )
        elif isinstance(node, ast.ImportFrom) and node.level == 0 and node.module:
            names = [
                node.module,
                *(f"{node.module}.{alias.name}" for alias in node.names),
            ]
            spans.extend(
                Span(line=node.lineno, end_line=end or node.lineno, module=name)
                for name in names
            )
    return spans


def empty_report(error: str, line: int) -> Report:
    return Report(
        error=error,
        error_line=line,
        comments=[],
        docstrings=[],
        placeholders=[],
        imports=[],
    )


def inspect(data: bytes) -> Report:
    try:
        encoding, _ = tokenize.detect_encoding(io.BytesIO(data).readline)
        data.decode(encoding)
        comments = comment_spans(data)
        tree = ast.parse(data)
    except (SyntaxError, UnicodeDecodeError, ValueError, tokenize.TokenError) as exc:
        line = exc.lineno if isinstance(exc, SyntaxError) and exc.lineno else 1
        return empty_report(f"{type(exc).__name__}: {exc}", line)
    report = empty_report("", 0)
    report["comments"] = comments
    report["docstrings"] = docstring_spans(tree)
    report["placeholders"] = placeholder_spans(tree)
    report["imports"] = import_spans(tree)
    return report


def main() -> int:
    request = json.load(sys.stdin)
    reports = {
        str(item["path"]): inspect(base64.b64decode(str(item["source"])))
        for item in request["files"]
    }
    json.dump({"files": reports}, sys.stdout)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

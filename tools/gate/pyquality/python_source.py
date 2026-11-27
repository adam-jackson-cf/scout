from __future__ import annotations

import ast
import base64
import io
import json
import sys
import tokenize
from dataclasses import asdict, dataclass
from typing import TypedDict


class SourceFile(TypedDict):
    path: str
    source: str


@dataclass(frozen=True)
class Finding:
    path: str
    line: int
    rule: str
    problem: str
    fix: str


@dataclass(frozen=True)
class Request:
    config: dict[str, object]
    files: list[SourceFile]
    modules: list[str]


@dataclass(frozen=True)
class Parsed:
    path: str
    text: str
    tree: ast.Module


def read_request() -> Request:
    payload = json.load(sys.stdin)
    return Request(
        config=dict(payload["config"]),
        files=list(payload["files"]),
        modules=[str(module) for module in payload.get("modules") or []],
    )


def emit(findings: list[Finding]) -> int:
    ordered = sorted(findings, key=lambda item: (item.path, item.line, item.rule))
    json.dump([asdict(finding) for finding in ordered], sys.stdout)
    return 0


def parse(item: SourceFile, rule: str) -> Parsed | Finding:
    path = item["path"]
    data = base64.b64decode(item["source"])
    try:
        encoding, _ = tokenize.detect_encoding(io.BytesIO(data).readline)
        text = data.decode(encoding)
    except (SyntaxError, UnicodeDecodeError) as exc:
        reason = exc.reason if isinstance(exc, UnicodeDecodeError) else exc.msg
        return Finding(
            path,
            1,
            rule,
            f"could not decode Python source: {reason}",
            "save the file as UTF-8 or declare its PEP 263 encoding",
        )
    try:
        tree = ast.parse(text, filename=path)
    except (SyntaxError, ValueError) as exc:
        line = exc.lineno if isinstance(exc, SyntaxError) and exc.lineno else 1
        message = exc.msg if isinstance(exc, SyntaxError) else str(exc)
        return Finding(
            path,
            line,
            rule,
            f"could not parse Python source: {message}",
            "fix the syntax error so the file parses",
        )
    return Parsed(path, text, tree)


def strings(config: dict[str, object], key: str) -> tuple[str, ...]:
    value = config.get(key) or []
    if not isinstance(value, list) or not all(isinstance(item, str) for item in value):
        raise SystemExit(
            f"proofs.toml [python_quality].{key} must be a list of strings"
        )
    return tuple(str(item) for item in value)


def integer(config: dict[str, object], key: str) -> int:
    value = config.get(key)
    if not isinstance(value, int) or isinstance(value, bool) or value < 1:
        raise SystemExit(
            f"proofs.toml [python_quality].{key} must be a positive integer"
        )
    return value


def mapping(config: dict[str, object], key: str) -> dict[str, frozenset[str]]:
    value = config.get(key) or {}
    if not isinstance(value, dict):
        raise SystemExit(
            f"proofs.toml [python_quality].{key} must be a table of string lists"
        )
    result: dict[str, frozenset[str]] = {}
    for name, items in value.items():
        if not isinstance(items, list) or not all(
            isinstance(item, str) for item in items
        ):
            raise SystemExit(
                f"proofs.toml [python_quality].{key}.{name} must be a list of strings"
            )
        result[str(name)] = frozenset(str(item) for item in items)
    return result


def under(path: str, roots: tuple[str, ...]) -> bool:
    return any(
        path == root or path.startswith(f"{root.rstrip('/')}/") for root in roots
    )


def dotted_name(node: ast.expr) -> str | None:
    parts: list[str] = []
    while isinstance(node, ast.Attribute):
        parts.append(node.attr)
        node = node.value
    if not isinstance(node, ast.Name):
        return None
    parts.append(node.id)
    return ".".join(reversed(parts))

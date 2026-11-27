from __future__ import annotations

import ast
from dataclasses import dataclass, field

from python_source import Finding, emit, integer, parse, read_request

FunctionNode = ast.FunctionDef | ast.AsyncFunctionDef
LoopNode = ast.For | ast.AsyncFor | ast.While
ComprehensionNode = ast.ListComp | ast.SetComp | ast.GeneratorExp | ast.DictComp


@dataclass
class LoopDepth:
    path: str
    limit: int
    findings: list[Finding] = field(default_factory=list)
    functions: list[str] = field(default_factory=list)

    def scan(self, node: ast.AST, depth: int) -> None:
        if isinstance(node, FunctionNode):
            self.scan_function(node)
        elif isinstance(node, ast.Lambda):
            self.scan_children(node.args, depth)
            self.scan(node.body, 0)
        elif isinstance(node, LoopNode):
            self.scan_loop(node, depth)
        elif isinstance(node, ComprehensionNode):
            self.scan_comprehension(node, node.generators, depth)
        else:
            self.scan_children(node, depth)

    def scan_children(self, node: ast.AST, depth: int) -> None:
        for child in ast.iter_child_nodes(node):
            self.scan(child, depth)

    def scan_function(self, node: ast.FunctionDef | ast.AsyncFunctionDef) -> None:
        for decorator in node.decorator_list:
            self.scan(decorator, 0)
        self.scan_children(node.args, 0)
        self.functions.append(node.name)
        for statement in node.body:
            self.scan(statement, 0)
        self.functions.pop()

    def scan_loop(self, node: ast.For | ast.AsyncFor | ast.While, depth: int) -> None:
        if not isinstance(node, ast.While):
            self.scan(node.iter, depth)
        inner = self.enter(node.lineno, depth)
        if isinstance(node, ast.While):
            self.scan(node.test, inner)
        else:
            self.scan(node.target, inner)
        for statement in node.body:
            self.scan(statement, inner)
        for statement in node.orelse:
            self.scan(statement, depth)

    def scan_comprehension(
        self,
        node: ast.ListComp | ast.SetComp | ast.GeneratorExp | ast.DictComp,
        generators: list[ast.comprehension],
        depth: int,
    ) -> None:
        if not generators:
            results = (
                [node.key, node.value] if isinstance(node, ast.DictComp) else [node.elt]
            )
            for result in results:
                self.scan(result, depth)
            return
        generator = generators[0]
        self.scan(generator.iter, depth)
        inner = self.enter(generator.target.lineno, depth)
        self.scan(generator.target, inner)
        for condition in generator.ifs:
            self.scan(condition, inner)
        self.scan_comprehension(node, generators[1:], inner)

    def enter(self, line: int, depth: int) -> int:
        inner = depth + 1
        if inner > self.limit:
            owner = self.functions[-1] if self.functions else "<module>"
            self.findings.append(
                Finding(
                    self.path,
                    line,
                    "PYS250",
                    f"nested loop depth {inner} exceeds {self.limit} in {owner}",
                    "extract the inner loop into a function, or index the data to remove a nesting level",
                )
            )
        return inner


def main() -> int:
    request = read_request()
    limit = integer(request.config, "max_loop_depth")
    findings: list[Finding] = []
    for item in request.files:
        parsed = parse(item, "PYS000")
        if isinstance(parsed, Finding):
            findings.append(parsed)
            continue
        visitor = LoopDepth(parsed.path, limit)
        visitor.scan(parsed.tree, 0)
        findings.extend(visitor.findings)
    return emit(findings)


if __name__ == "__main__":
    raise SystemExit(main())

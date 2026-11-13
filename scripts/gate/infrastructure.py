#!/usr/bin/env python3
"""Check toolchain pin parity, the CI gate workflow, Markdown/YAML formatting and dependency audit."""

from __future__ import annotations

import json
import re
import subprocess
import sys
from pathlib import Path

ROOT = Path.cwd()
ERRORS: list[str] = []
GOVERNED = {
    "docs/scout-design.html",
    "docs/research-ledger.md",
    "docs/hypothesis-ledger.md",
    "CLAUDE.md",
    "AGENTS.md",
}


def run(argv: list[str], *, capture: bool = False) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        argv, text=True, capture_output=capture, check=False, timeout=600
    )


def fail(path: str, rule: str, fix: str) -> None:
    ERRORS.append(f"{path}: {rule}: {fix}")


def pins(root: Path) -> dict[str, str]:
    return dict(
        re.findall(
            r"^([A-Z_]+)=(\S+)$", (root / "toolchain/versions.env").read_text(), re.M
        )
    )


def parity(root: Path) -> None:
    versions = pins(root)
    packages = json.loads((root / "toolchain/node/package.json").read_text())[
        "dependencies"
    ]
    lock = json.loads((root / "toolchain/node/package-lock.json").read_text())[
        "packages"
    ]
    for name, version in packages.items():
        if (
            version != versions[name.upper() + "_VERSION"]
            or lock["node_modules/" + name]["version"] != version
            or lock[""]["dependencies"].get(name) != version
        ):
            fail(
                "toolchain/node/package-lock.json",
                "PARITY.lock-version",
                f"relock {name} from versions.env",
            )
    declared = dict(
        re.findall(
            r"^([\w-]+)==([^\s]+)$",
            (root / "toolchain/python-requirements.in").read_text(),
            re.M,
        )
    )
    installed = dict(
        re.findall(
            r"^([\w-]+)==([^\s\\]+)",
            (root / "toolchain/python-requirements.lock").read_text(),
            re.M,
        )
    )
    for name, version in declared.items():
        if (
            installed.get(name) != version
            or versions.get(name.upper() + "_VERSION", version) != version
        ):
            fail(
                "toolchain/python-requirements.lock",
                "PARITY.lock-version",
                f"relock {name} from versions.env",
            )
    if f"toolchain go{versions['GO_VERSION']}" not in (root / "go.mod").read_text():
        fail("go.mod", "PARITY.lock-version", "match the pinned Go toolchain")
    bootstrap = (root / "scripts/bootstrap-tools.sh").read_text()
    for tool in ("semgrep", "ruff", "lizard", "gitleaks", "jscpd", "mypy", "prettier"):
        if tool not in bootstrap or tool not in {*declared, *packages, "gitleaks"}:
            fail(
                "scripts/bootstrap-tools.sh",
                "PARITY.dependency-inventory",
                f"provision declared analyzer/tool {tool}",
            )
    guidance = (root / "CONTRIBUTING.md").read_text()
    for command in (
        "scripts/bootstrap-tools.sh",
        "source .tools/activate",
        "bash scripts/gate.sh ci",
        "git config core.hooksPath .githooks",
    ):
        if command not in guidance:
            fail(
                "CONTRIBUTING.md",
                "PARITY.install-guidance",
                f"document the executable installer command: {command}",
            )


def ci_prerequisites(workflow: str) -> None:
    pin = "source toolchain/versions.env"
    setup = "go-version: ${{ env.GO_VERSION }}"
    sandbox = "bash scripts/ci-sandbox.sh --strict"
    bootstrap = "scripts/bootstrap-tools.sh"
    if not all(token in workflow for token in (pin, setup, sandbox, bootstrap)):
        fail(
            ".github/workflows/gate.yml",
            "GATE.local-ci-parity",
            "read the Go pin from versions.env and require ci-sandbox.sh --strict before bootstrap",
        )
    elif workflow.index(pin) > workflow.index(setup) or workflow.index(
        sandbox
    ) > workflow.index(bootstrap):
        fail(
            ".github/workflows/gate.yml",
            "GATE.local-ci-parity",
            "load the Go pin before setup-go and require the scanner sandbox before bootstrap",
        )


def runner(root: Path) -> None:
    workflow = (root / ".github/workflows/gate.yml").read_text()
    ci_prerequisites(workflow)
    if "name: gate" not in workflow or "bash scripts/gate.sh ci" not in workflow:
        fail(
            ".github/workflows/gate.yml",
            "GATE.local-ci-parity",
            "run bash scripts/gate.sh ci in the gate job",
        )


def prettier(files: list[str], root: Path) -> None:
    version = pins(root)["PRETTIER_VERSION"]
    actual = run(["prettier", "--version"], capture=True).stdout.strip()
    if version != "3.6.2" or actual != version:
        fail(
            "toolchain/versions.env",
            "PRETTIER.version",
            "install pinned Prettier 3.6.2",
        )
        return
    maintained = [
        p
        for p in files
        if Path(p).suffix in {".md", ".yaml", ".yml"}
        and p not in GOVERNED
        and not p.startswith((
            "evaluation/",
            "docs/adrs/",
            "internal/analyzers/rules/",
            "third_party/",
            ".codegraph/",
        ))
        and ".codegraph" not in Path(p).parts
        and not (p.startswith("internal/") and "/testdata/" in p)
        and (root / p).is_file()
    ]
    if maintained and run(["prettier", "--check", *maintained]).returncode:
        fail(
            "maintained Markdown/YAML",
            "PRETTIER.version",
            "format explicitly with pinned prettier; never format frozen/governed paths",
        )


def audit(args: list[str]) -> None:
    if run(args).returncode:
        fail(
            "go.mod",
            "SECURITY.dependency-audit",
            "resolve dependency vulnerabilities or audit failure; do not suppress findings",
        )


def main() -> bool:
    mode = sys.argv[1]
    parity(ROOT)
    runner(ROOT)
    files = (
        sys.argv[2:]
        if mode == "commit"
        else run(["git", "ls-files", "-z"], capture=True).stdout.split("\0")
    )
    prettier(files, ROOT)
    if mode == "ci":
        audit(["govulncheck", "-C", "cmd/scout", "-scan=module"])
    for message in ERRORS:
        print(message)
    return bool(ERRORS)


if __name__ == "__main__":
    sys.exit(main())

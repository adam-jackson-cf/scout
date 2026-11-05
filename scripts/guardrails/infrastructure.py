#!/usr/bin/env python3
"""Enforce tool, contributor, workflow, skill and runtime-seam contracts."""

from __future__ import annotations

import hashlib
import json
import re
import subprocess
import sys
import time
from pathlib import Path
from typing import Any

ROOT = Path.cwd()
ERRORS: list[str] = []
GOVERNED = {
    "docs/scout-design.md",
    "SCOUT_REBUILD_REQUIREMENTS.md",
    "docs/research-ledger.md",
    "docs/hypothesis-ledger.md",
    "CLAUDE.md",
    "AGENTS.md",
    "HANDOFF.md",
}


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
        "bash scripts/gate.sh full",
        "git config core.hooksPath .githooks",
    ):
        if command not in guidance:
            fail(
                "CONTRIBUTING.md",
                "PARITY.install-guidance",
                f"document the executable installer command: {command}",
            )


def gate_schema(config: dict[str, Any]) -> None:
    rule = "RUNNER.local-ci-parity"
    if type(config.get("schemaVersion")) is not int or config["schemaVersion"] != 1:
        fail(".omp/gate.json", rule, "set schemaVersion to integer 1")
    timeout = config.get("pathCheckTimeoutMs")
    if type(timeout) is not int or timeout <= 0:
        fail(
            ".omp/gate.json",
            rule,
            "set a positive integer pathCheckTimeoutMs with measured headroom",
        )
    for field in ("testGlobs", "sensitive"):
        values = config.get(field)
        if (
            not isinstance(values, list)
            or not values
            or not all(isinstance(value, str) and value for value in values)
        ):
            fail(
                ".omp/gate.json",
                rule,
                f"populate {field} with nonempty string patterns",
            )
    if config.get("agentMerge") not in {"allowed", "disabled"}:
        fail(".omp/gate.json", rule, "set agentMerge to allowed or disabled")


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
    config = json.loads((root / ".omp/gate.json").read_text())
    gate_schema(config)
    for field, mode in {
        "gate": "full",
        "commitGate": "commit",
        "fastGate": "fast",
        "pathCheck": "path",
    }.items():
        if config.get(field) != ["bash", "scripts/gate.sh", mode]:
            fail(
                ".omp/gate.json",
                "RUNNER.local-ci-parity",
                f"use authoritative gate.sh {mode} argv for {field}",
            )
    workflow = (root / ".github/workflows/gate.yml").read_text()
    ci_prerequisites(workflow)
    if (
        'json.load(open(".omp/gate.json"))["gate"]' not in workflow
        or "name: gate" not in workflow
        or "subprocess.run" not in workflow
        or "sys.exit" not in workflow
    ):
        fail(
            ".github/workflows/gate.yml",
            "GATE.local-ci-parity",
            "run exact gate argv and propagate its exit in the gate job",
        )


def skills(root: Path) -> None:
    for path in sorted((root / ".agents/skills").glob("*/SKILL.md")):
        text = path.read_text()
        front = text.split("---", 2)
        metadata = dict(
            re.findall(
                r"^(\w+):\s*[\"']?(.*?)[\"']?$",
                front[1] if len(front) == 3 else "",
                re.M,
            )
        )
        if metadata.get("name") != path.parent.name or not metadata.get(
            "description", ""
        ).lower().startswith("use when"):
            fail(
                str(path.relative_to(root)),
                "SKILL.metadata",
                "match name to directory and begin description with USE WHEN",
            )
        if "global" in metadata and metadata["global"].lower() not in {"true", "false"}:
            fail(
                str(path.relative_to(root)), "SKILL.metadata", "set global to a boolean"
            )
        for link in re.findall(r"\]\(([^)#]+)(?:#[^)]*)?\)", text):
            if "://" not in link and not (path.parent / link).exists():
                fail(
                    str(path.relative_to(root)),
                    "SKILL.metadata",
                    f"repair missing support file {link}",
                )


def host(record: dict[str, Any]) -> None:
    protection = record.get("protection", {})
    contexts = protection.get("required_status_checks", {}).get("contexts", [])
    if (
        "gate" not in contexts
        or not protection.get("enforce_admins", {}).get("enabled")
        or not record.get("allow_auto_merge")
        or not record.get("allow_rebase_merge")
    ):
        fail(
            ".omp/host-settings.json",
            "GIT.required-checks",
            "require gate on default branch, enforce admins and enable auto/rebase merge",
        )


def prettier(files: list[str], root: Path) -> None:
    version = pins(root)["PRETTIER_VERSION"]
    actual = subprocess.check_output(["prettier", "--version"], text=True).strip()
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
            "docs/design-reviews/",
            "docs/archive/",
            "internal/analyzers/rules/",
            "scripts/guardrails/vendor/",
            "third_party/",
            ".codegraph/",
        ))
        and ".codegraph" not in Path(p).parts
        and not (p.startswith("internal/") and "/testdata/" in p)
        and (root / p).is_file()
    ]
    if (
        maintained
        and subprocess.run(["prettier", "--check", *maintained], check=False).returncode
    ):
        fail(
            "maintained Markdown/YAML",
            "PRETTIER.version",
            "format explicitly with pinned prettier; never format frozen/governed paths",
        )


def seam_record(record: dict[str, Any], root: Path) -> None:
    revision = subprocess.check_output(
        ["git", "rev-parse", "HEAD"], cwd=root, text=True
    ).strip()
    if (
        record.get("revision") != revision
        or not 0 <= time.time() - record.get("startedUnix", 0) <= 3600
    ):
        fail(
            ".tools/guardrails-seam-runtime.json",
            "SEAM.evidence",
            "regenerate fresh runtime seam evidence at this revision",
        )
    for artifact in record.get("artifacts", []):
        path = root / artifact["path"]
        if (
            not path.is_file()
            or hashlib.sha256(path.read_bytes()).hexdigest() != artifact["sha256"]
        ):
            fail(
                str(path),
                "SEAM.evidence",
                "repair missing or stale manifest/report artifact",
            )
    if not record.get("artifacts") or record.get("exit_code") != 0:
        fail(
            ".tools/guardrails-seam-runtime.json",
            "SEAM.evidence",
            "record successful real registry/artifact freshness tests",
        )


def runtime_seams(root: Path) -> None:
    result = subprocess.run(
        ["go", "test", "-count=1", "./internal/toolchain", "./internal/artifacts"],
        capture_output=True,
        text=True,
        check=False,
    )
    if result.returncode:
        fail(
            "internal/toolchain,internal/artifacts",
            "SEAM.evidence",
            "repair registry pin/provenance/revision freshness contract tests",
        )
        print(result.stdout + result.stderr)
    report = root / ".tools/guardrails-seam-report.txt"
    report.write_text(result.stdout + result.stderr)
    artifacts = [
        "toolchain/versions.env",
        "toolchain/python-requirements.lock",
        "toolchain/node/package-lock.json",
        ".tools/guardrails-seam-report.txt",
    ]
    record = {
        "startedUnix": time.time(),
        "revision": subprocess.check_output(
            ["git", "rev-parse", "HEAD"], text=True
        ).strip(),
        "exit_code": result.returncode,
        "artifacts": [
            {"path": p, "sha256": hashlib.sha256((root / p).read_bytes()).hexdigest()}
            for p in artifacts
        ],
    }
    (root / ".tools/guardrails-seam-runtime.json").write_text(
        json.dumps(record, indent=2) + "\n"
    )
    seam_record(record, root)


def audit(args: list[str]) -> None:
    if subprocess.run(args, check=False).returncode:
        fail(
            "go.mod",
            "SECURITY.dependency-audit",
            "resolve dependency vulnerabilities or audit failure; do not suppress findings",
        )


def main() -> bool:
    mode = sys.argv[1]
    parity(ROOT)
    runner(ROOT)
    skills(ROOT)
    files = (
        sys.argv[2:]
        if mode == "commit"
        else subprocess.check_output(["git", "ls-files", "-z"], text=True).split("\0")
    )
    prettier(files, ROOT)
    if mode == "full":
        runtime_seams(ROOT)
        audit(["govulncheck", "-C", "cmd/scout", "-scan=module"])
    for message in ERRORS:
        print(message)
    return bool(ERRORS)


if __name__ == "__main__":
    sys.exit(main())

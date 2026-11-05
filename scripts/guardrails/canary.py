#!/usr/bin/env python3
"""Exercise real infrastructure checkers against disposable violating roots."""

from __future__ import annotations

import hashlib
import importlib.util
import json
import os
import shutil
import subprocess
import sys
import tempfile
import time
from pathlib import Path
from types import ModuleType
from typing import Any

SOURCE = Path(__file__).resolve().parents[2]


def load(name: str) -> ModuleType:
    spec = importlib.util.spec_from_file_location(
        name, SOURCE / f"scripts/guardrails/{name}.py"
    )
    assert spec
    assert spec.loader
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def command(args: list[str], cwd: Path) -> subprocess.CompletedProcess[str]:
    return subprocess.run(args, cwd=cwd, text=True, capture_output=True, check=False)


def git(root: Path, *args: str) -> None:
    subprocess.run(["git", *args], cwd=root, check=True, capture_output=True)


def hooks(root: Path, rule: str, bad: bool) -> bool:
    hook = rule.removeprefix("GATE.")
    repo = root / "repo"
    repo.mkdir()
    git(repo, "init", "-q")
    for key, value in [
        ("user.name", "Canary"),
        ("user.email", "canary@example.invalid"),
        ("core.hooksPath", ".githooks"),
    ]:
        git(repo, "config", key, value)
    for directory in (".githooks", ".omp", "scripts/guardrails", ".tools/venv/bin"):
        (repo / directory).mkdir(parents=True, exist_ok=True)
    (repo / ".tools/venv/bin/python").symlink_to(sys.executable)
    shutil.copy2(
        SOURCE / "scripts/guardrails/hook.py", repo / "scripts/guardrails/hook.py"
    )
    shutil.copy2(SOURCE / ".githooks" / hook, repo / ".githooks" / hook)
    field = "commitGate" if hook == "pre-commit" else "fastGate"
    script = "import sys;sys.exit(" + str(int(bad)) + ")"
    if hook == "pre-push":
        script = (
            "import os;assert len(os.environ['SCOUT_PUSH_RANGES'].split())==4;" + script
        )
    (repo / ".omp/gate.json").write_text(
        json.dumps({"schemaVersion": 1, field: [sys.executable, "-c", script]})
    )
    (repo / "a").write_text("test")
    git(repo, "add", ".")
    result = command(["git", "commit", "-qm", "canary"], repo)
    if hook == "pre-push":
        if result.returncode:
            raise RuntimeError(result.stderr)
        git(root, "init", "--bare", "-q", str(root / "remote"))
        result = command(["git", "push", str(root / "remote"), "HEAD:main"], repo)
    if result.returncode:
        print(f"fixture: {rule}: repair failing declared gate before operation")
    return result.returncode != 0


def path_canary(root: Path, fixture: dict[str, Any], gate: ModuleType) -> None:
    rule, bad = fixture["rule"], fixture["bad"]
    (root / "scripts/guardrails").mkdir(parents=True)
    shutil.copy2(
        SOURCE / "scripts/guardrails/inspect.go", root / "scripts/guardrails/inspect.go"
    )
    policy: dict[str, Any] = {
        "owned_paths": [r"internal/[^/]+/[^/]+\.go"],
        "dependencies": {"internal/unit": ["internal/domain"]},
    }
    name, text = "internal/unit/value.go", "package unit\nfunc Value() {}\n"
    if bad:
        names = {
            "STRUCT.folder-placement": "unowned/value.go",
            "NAME.conventions": "internal/unit/BadName.go",
            "TEST.placement": "scripts/unit/value_test.go",
        }
        name = names.get(rule, name)
        if rule == "TEST.placement":
            policy["owned_paths"].append(r"scripts/[^/]+/[^/]+\.go")
        if rule in ("STRUCT.dependency-direction", "GATE.path-check"):
            text = 'package unit\nimport _ "scout/internal/forbidden"\n'
        if rule.startswith("FANOUT."):
            count = 4 if rule.endswith("large-module") else 11
            text = (
                "package unit\nimport (\n"
                + "\n".join(f'_ "example.invalid/pkg{i}"' for i in range(count))
                + "\n)\n"
            )
            if rule.endswith("large-module"):
                text += "\n".join(f"var Value{i} int" for i in range(805))
    elif rule == "FANOUT.large-module":
        text += "// comment\n" * 1000  # Physical LOC must not trigger the NLOC limit.
    path = root / name
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text)
    os.chdir(root)
    gate.__dict__["ROOT"] = root
    if rule == "GATE.path-check" and not bad:
        name = str(path)
    gate.path_check([name], policy)
    if rule == "GATE.path-check":
        gate.__dict__["ERRORS"] = [
            line.replace("STRUCT.dependency-direction", rule) for line in gate.ERRORS
        ]


def copy_contract(root: Path) -> None:
    for name in (
        "toolchain/versions.env",
        "toolchain/node/package.json",
        "toolchain/node/package-lock.json",
        "toolchain/python-requirements.in",
        "toolchain/python-requirements.lock",
        "scripts/bootstrap-tools.sh",
        "CONTRIBUTING.md",
        "go.mod",
        ".omp/gate.json",
        ".github/workflows/gate.yml",
    ):
        target = root / name
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(SOURCE / name, target)


def activation_canary(rule: str, infra: ModuleType) -> None:
    script = """set -e
source .tools/activate
export EXPECTED_PATH="$PATH"
bash -c 'set -e
if declare -F scout_deactivate >/dev/null; then exit 2; fi
source .tools/activate
[[ "$PATH" == "$EXPECTED_PATH" ]]
source .tools/activate
[[ "$PATH" == "$EXPECTED_PATH" ]]
bash scripts/gate.sh path'
"""
    result = command(["bash", "-c", script], SOURCE)
    if result.returncode:
        infra.fail(
            ".tools/activate",
            rule,
            "support activated parents without missing functions or duplicate PATH entries",
        )
        print(result.stdout + result.stderr)


def ledger_canary(root: Path, fixture: dict[str, Any], infra: ModuleType) -> None:
    script = root / "scripts/guardrails/merge-ledger.py"
    script.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(SOURCE / "scripts/guardrails/merge-ledger.py", script)
    fragments = root / ".omp/ledger.d"
    fragments.mkdir()
    (fragments / "canary.json").write_text('{"rows":[{"id":"canary"}]}\n')
    result = command([sys.executable, str(script)], root)
    if result.returncode:
        raise RuntimeError(result.stderr)
    if fixture["bad"]:
        ledger = root / ".omp/coverage-ledger.json"
        ledger.write_text(ledger.read_text() + "\n")
    result = command([sys.executable, str(script), "--check"], root)
    if result.returncode:
        infra.fail(
            ".omp/coverage-ledger.json",
            fixture["rule"],
            "run scripts/guardrails/merge-ledger.py to synchronize fragments",
        )


def configuration_canary(
    root: Path, fixture: dict[str, Any], infra: ModuleType
) -> None:
    rule, bad = fixture["rule"], fixture["bad"]
    copy_contract(root)
    if rule.startswith("PARITY."):
        targets = {
            "PARITY.lock-version": "toolchain/node/package-lock.json",
            "PARITY.install-guidance": "CONTRIBUTING.md",
            "PARITY.dependency-inventory": "scripts/bootstrap-tools.sh",
        }
        if bad:
            path = root / targets[rule]
            if rule == "PARITY.lock-version":
                data = json.loads(path.read_text())
                data["packages"]["node_modules/prettier"]["version"] = "0.0.0"
                path.write_text(json.dumps(data))
            else:
                path.write_text(
                    path.read_text().replace(
                        "scripts/bootstrap-tools.sh"
                        if rule.endswith("install-guidance")
                        else "semgrep",
                        "missing",
                    )
                )
        infra.parity(root)
    elif rule in {"GATE.local-ci-parity", "RUNNER.local-ci-parity"}:
        path = root / (
            ".github/workflows/gate.yml"
            if rule.startswith("GATE.")
            else ".omp/gate.json"
        )
        if bad:
            path.write_text(
                path.read_text().replace('"gate"', '"notGate"')
                if rule.startswith("GATE.")
                else "{}"
            )
        infra.runner(root)
        activation_canary(rule, infra)
        ledger_canary(root, fixture, infra)
    elif rule == "PRETTIER.version":
        path = root / "sample.md"
        path.write_text("# Title\n\ntext\n" if not bad else "# Title\n\n-   text\n")
        os.chdir(root)
        infra.prettier(["sample.md"], root)


def metadata_canary(root: Path, fixture: dict[str, Any], infra: ModuleType) -> None:
    path = root / ".agents/skills/sample/SKILL.md"
    path.parent.mkdir(parents=True)
    path.write_text(
        '---\nname: "'
        + ("wrong" if fixture["bad"] else "sample")
        + '"\ndescription: "USE WHEN testing metadata"\n---\n'
    )
    infra.skills(root)


def evidence_canary(root: Path, fixture: dict[str, Any], infra: ModuleType) -> None:
    path = SOURCE / "toolchain/versions.env"
    record = {
        "startedUnix": time.time() - (4000 if fixture["bad"] else 1),
        "revision": subprocess.check_output(
            ["git", "rev-parse", "HEAD"], cwd=SOURCE, text=True
        ).strip(),
        "exit_code": 0,
        "artifacts": [
            {
                "path": "toolchain/versions.env",
                "sha256": hashlib.sha256(path.read_bytes()).hexdigest(),
            }
        ],
    }
    infra.seam_record(record, SOURCE)


def security_canary(root: Path, bad: bool, infra: ModuleType) -> None:
    module = "module auditcanary\n\ngo 1.26.0\n"
    source = "package auditcanary\n"
    if bad:
        module += "\nrequire golang.org/x/text v0.3.0\n"
        source += 'import _ "golang.org/x/text/language"\n'
    (root / "go.mod").write_text(module)
    (root / "value.go").write_text(source)
    result = command(["go", "mod", "tidy"], root)
    if result.returncode:
        raise RuntimeError(result.stderr)
    os.chdir(root)
    infra.audit(["govulncheck", "-scan=module"])


def main() -> bool:
    fixture = json.loads(Path(sys.argv[1]).read_text())
    rule = fixture["rule"]
    gate, infra = load("gate"), load("infrastructure")
    with tempfile.TemporaryDirectory(prefix="scout-canary-") as temp:
        root = Path(temp)
        if rule.startswith("GATE.pre-"):
            return hooks(root, rule, fixture["bad"])
        if rule.startswith(("PARITY.", "RUNNER.")) or rule in {
            "GATE.local-ci-parity",
            "PRETTIER.version",
        }:
            configuration_canary(root, fixture, infra)
        elif rule == "SKILL.metadata":
            metadata_canary(root, fixture, infra)
        elif rule == "SEAM.evidence":
            evidence_canary(root, fixture, infra)
        elif rule == "GIT.required-checks":
            infra.host({
                "protection": {
                    "required_status_checks": {
                        "contexts": [] if fixture["bad"] else ["gate"]
                    },
                    "enforce_admins": {"enabled": True},
                },
                "allow_auto_merge": True,
                "allow_rebase_merge": True,
            })
        elif rule == "SECURITY.dependency-audit":
            security_canary(root, fixture["bad"], infra)
        elif rule == "TEST.failure-propagation":
            try:
                gate.run([
                    sys.executable,
                    "-c",
                    f"import sys; sys.exit({int(fixture['bad'])})",
                ])
            except subprocess.CalledProcessError:
                gate.error("fixture", rule, "repair failed prerequisite")
        else:
            path_canary(root, fixture, gate)
        messages = gate.ERRORS + infra.ERRORS
        for line in messages:
            print(line)
        return bool(messages)


if __name__ == "__main__":
    sys.exit(main())

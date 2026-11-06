"""One public-entrypoint proof for each changed guardrail clause."""

from __future__ import annotations

import subprocess
import sys
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[4]
FIXTURES = ROOT / "scripts/guardrails/fixtures"


def check_case(case: Path, *, infrastructure: bool) -> subprocess.CompletedProcess[str]:
    argv = (
        ["scripts/guardrails/canary.py", str(case)]
        if infrastructure
        else ["scripts/guardrails/python/check.py", "fixture", str(case)]
    )
    return subprocess.run(
        [sys.executable, *argv],
        cwd=ROOT,
        text=True,
        capture_output=True,
        check=False,
        timeout=90,
    )


class ChangedRuleProof(unittest.TestCase):
    def assert_contract(self, rule: str, *, infrastructure: bool = False) -> None:
        prefix = FIXTURES / rule
        cases = (
            sorted(FIXTURES.glob(rule + ".contract-*.json"))
            if infrastructure
            else sorted(prefix.glob("contract-*.json"))
        )
        assert len(cases) >= 2, rule + ": missing proof inputs"
        for case in cases:
            with self.subTest(rule=rule, case=case.name):
                result = check_case(case, infrastructure=infrastructure)
                output = result.stdout + result.stderr
                if case.name.endswith("contract-passing.json"):
                    assert result.returncode == 0, output
                else:
                    assert result.returncode != 0, (
                        rule + ": NEW contract violation accepted: " + case.name
                    )
                    assert ": " + rule + ": " in output
                    assert "Traceback" not in output

    def test_mypy_exclude(self) -> None:
        self.assert_contract("MYPY.SETTING.exclude")

    def test_mypy_import_roots(self) -> None:
        self.assert_contract("MYPY.SETTING.mypy_path")

    def test_mypy_supported_version(self) -> None:
        self.assert_contract("MYPY.SETTING.python_version")

    def test_mypy_override_evidence(self) -> None:
        self.assert_contract("MYPY.module-overrides")

    def test_mypy_package_inventory(self) -> None:
        self.assert_contract("MYPY.package-coverage")

    def test_canonical_test_groups(self) -> None:
        self.assert_contract("PYQ102")

    def test_first_party_inventory(self) -> None:
        self.assert_contract("RUFF.ISORT.known-first-party")

    def test_ruff_exclusion_rationales(self) -> None:
        self.assert_contract("RUFF.ROOT.exclude")

    def test_ruff_supported_version(self) -> None:
        self.assert_contract("RUFF.SETTING.target-version")

    def test_skill_metadata_schema(self) -> None:
        self.assert_contract("SKILL.metadata", infrastructure=True)

    def test_ci_schema_and_argv(self) -> None:
        self.assert_contract("GATE.local-ci-parity", infrastructure=True)

    def test_commit_field_and_installation(self) -> None:
        self.assert_contract("GATE.pre-commit", infrastructure=True)


if __name__ == "__main__":
    unittest.main()

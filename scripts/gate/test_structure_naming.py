"""U2 boundary proofs for structure, naming and test-layout diagnostics."""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path

from test_inventory import GOOD_GO, GOOD_TEST, GateFixture

BANNED_FOLDERS = (
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
)


class StructureNamingCanaries(unittest.TestCase):
    """Exercise the public changed-path command with isolated real policy files."""

    def setUp(self) -> None:
        self.temporary = tempfile.TemporaryDirectory(prefix="scout-structure-canary-")
        self.addCleanup(self.temporary.cleanup)
        root = Path(self.temporary.name) / "repo"
        root.mkdir()
        self.fixture = GateFixture(root)

    def path_check(self, *paths: str) -> subprocess.CompletedProcess[str]:
        return self.fixture.run("bash", "scripts/gate.sh", "path", *paths)

    def require_rejection(
        self, result: subprocess.CompletedProcess[str], path: str, *details: str
    ) -> None:
        assert result.returncode != 0, f"gate accepted {path}:\n{result.stdout}"
        lines = [
            line for line in result.stdout.splitlines() if line.startswith(path + ":")
        ]
        assert lines, f"missing file-specific diagnostic for {path}:\n{result.stdout}"
        assert any(all(detail in line for detail in details) for line in lines), (
            f"expected one diagnostic naming {path} and {details!r}:\n{result.stdout}"
        )

    def test_unmapped_go_dependency_target_is_rejected_with_edge_and_fix(self) -> None:
        # An allowlisted edge cannot silently introduce an owner with no direction map.
        policy_path = "scripts/gate/structure.json"
        policy = json.loads((self.fixture.root / policy_path).read_text())
        policy["dependencies"]["internal/domain"] = ["internal/unmapped"]
        self.fixture.write(policy_path, json.dumps(policy))
        self.fixture.write(
            "internal/unmapped/value.go", GOOD_GO.replace("domain", "unmapped")
        )
        path = "internal/domain/edge.go"
        self.fixture.write(
            path, 'package domain\n\nimport _ "scout/internal/unmapped"\n'
        )
        self.require_rejection(
            self.path_check(path),
            path,
            "internal/domain -> internal/unmapped",
            "declare",
        )

    def test_python_reverse_dependency_is_rejected_with_edge_and_fix(self) -> None:
        # Inventory is a leaf; importing the dispatcher reverses its declared direction.
        policy_path = "scripts/gate/structure.json"
        policy = json.loads((self.fixture.root / policy_path).read_text())
        policy["dependencies"].update({
            "scripts/gate/run.py": ["scripts/gate/inventory.py"],
            "scripts/gate/inventory.py": [],
        })
        self.fixture.write(policy_path, json.dumps(policy))
        path = "scripts/gate/inventory.py"
        original = (self.fixture.root / path).read_text()
        self.fixture.write(path, original + "\nimport run\n")
        self.require_rejection(
            self.path_check(path), path, "scripts/gate/run.py", "approved"
        )

    def test_unmapped_folder_reports_concrete_owner_destination(self) -> None:
        # A generic owned-path refusal is insufficient without the repair destination.
        path = "internal/domain/nested/extra.go"
        self.fixture.write(path, GOOD_GO)
        self.require_rejection(self.path_check(path), path, "internal/domain/extra.go")

    def test_grab_bag_folders_are_rejected_with_destination(self) -> None:
        # Each exact banned directory segment must be refused, not just one example.
        for folder in BANNED_FOLDERS:
            with self.subTest(folder=folder):
                path = f"scripts/{folder}/scan.py"
                self.fixture.write(path, '"""Scan repository inputs."""\n')
                self.require_rejection(
                    self.path_check(path), path, "scripts/gate/scan.py"
                )

    def test_python_path_casing_reports_correct_name(self) -> None:
        # Python module names follow snake_case rather than accepting arbitrary stems.
        path = "scripts/gate/BadName.py"
        self.fixture.write(path, '"""Scan repository inputs."""\n')
        self.require_rejection(self.path_check(path), path, "scripts/gate/bad_name.py")

    def test_python_symbol_casing_reports_correct_name(self) -> None:
        # The changed-path entry point checks symbols as well as file names.
        path = "scripts/gate/sample.py"
        self.fixture.write(
            path,
            '"""Scan repository inputs."""\n\ndef badFunction() -> None:\n    pass\n',
        )
        self.require_rejection(
            self.path_check(path), path, "badFunction", "bad_function"
        )

    def test_test_file_ticket_version_and_date_tokens_report_correct_name(self) -> None:
        # Test identities describe behaviour, not issue numbers, revisions or dates.
        for token in ("issue42", "v2", "20261116"):
            with self.subTest(token=token):
                path = f"internal/domain/value_{token}_test.go"
                self.fixture.write(path, GOOD_TEST)
                self.require_rejection(
                    self.path_check(path), path, "internal/domain/value_test.go"
                )

    def test_nonmatching_test_file_reports_profile_pattern(self) -> None:
        # Test-like files using a prefix must use the declared Go suffix instead.
        path = "internal/domain/test_value.go"
        self.fixture.write(path, GOOD_TEST)
        self.require_rejection(
            self.path_check(path), path, "internal/domain/value_test.go"
        )

    def test_missing_test_tier_mapping_reports_owner_and_policy_fix(self) -> None:
        # Package placement alone must not hide a test absent from the tier map.
        policy_path = "scripts/gate/test-tiers.json"
        policy = json.loads((self.fixture.root / policy_path).read_text())
        del policy["packages"]["internal/domain"]
        self.fixture.write(policy_path, json.dumps(policy))
        path = "internal/domain/value_test.go"
        self.require_rejection(
            self.path_check(path),
            path,
            "internal/domain",
            "scripts/gate/test-tiers.json",
        )

    def test_invalid_test_tier_mapping_reports_file_and_policy_fix(self) -> None:
        # A mapped owner still needs one of the declared fast or seam tiers.
        policy_path = "scripts/gate/test-tiers.json"
        policy = json.loads((self.fixture.root / policy_path).read_text())
        policy["packages"]["internal/domain"]["tier"] = "unassigned"
        self.fixture.write(policy_path, json.dumps(policy))
        path = "internal/domain/value_test.go"
        self.require_rejection(
            self.path_check(path), path, "scripts/gate/test-tiers.json", "fast", "seam"
        )

    def test_wrong_test_owner_reports_destination(self) -> None:
        # A file declaring the areas package does not belong beside domain sources.
        path = "internal/domain/areas_test.go"
        self.fixture.write(path, GOOD_TEST.replace("package domain", "package areas"))
        self.require_rejection(
            self.path_check(path), path, "internal/areas/areas_test.go"
        )

    def test_fixture_outside_profile_layout_reports_destination(self) -> None:
        # Contributor docs cannot become an unowned fixture storage area.
        path = "docs/fixtures/domain/value.json"
        self.fixture.write(path, '{"value": 1}\n')
        self.require_rejection(
            self.path_check(path), path, "internal/domain/testdata/value.json"
        )

    def test_support_outside_profile_layout_reports_destination(self) -> None:
        # Shared state-test support has one declared home rather than arbitrary tooling.
        path = "scripts/test_support/store.go"
        self.fixture.write(path, "package statetest\n")
        self.require_rejection(
            self.path_check(path), path, "internal/state/statetest/store.go"
        )

    def test_path_command_reports_each_naming_and_placement_violation_on_one_line(
        self,
    ) -> None:
        # Multiple changed paths retain independent actionable, single-line findings.
        naming = "scripts/gate/BadName.py"
        placement = "docs/fixtures/domain/value.json"
        self.fixture.write(naming, '"""Scan repository inputs."""\n')
        self.fixture.write(placement, '{"value": 1}\n')
        result = self.path_check(naming, placement)
        self.require_rejection(result, naming, "scripts/gate/bad_name.py")
        self.require_rejection(result, placement, "internal/domain/testdata/value.json")
        assert len(result.stdout.splitlines()) == 2, result.stdout

    def test_conforming_owned_paths_and_allowed_dependency_are_accepted(self) -> None:
        # Existing owner conventions, mapped tests and approved dependency direction survive.
        self.fixture.write(
            "internal/areas/edge.go",
            'package areas\n\nimport _ "scout/internal/domain"\n',
        )
        paths = (
            "internal/domain/value.go",
            "internal/domain/value_test.go",
            "internal/areas/edge.go",
            "scripts/gate/inventory.py",
            "scripts/gate/run.py",
            "scripts/gate/test_structure_naming.py",
            "docs/features/repository-gate.md",
            "internal/association/testdata/value.json",
            "internal/state/statetest/store.go",
        )
        self.fixture.write(
            "scripts/gate/test_structure_naming.py", '"""Gate boundary tests."""\n'
        )
        self.fixture.write("docs/features/repository-gate.md", "# Repository gate\n")
        self.fixture.write("internal/association/testdata/value.json", "{}\n")
        self.fixture.write("internal/state/statetest/store.go", "package statetest\n")
        result = self.path_check(*paths)
        assert result.returncode == 0, result.stdout
        assert not result.stdout.strip(), result.stdout

    def test_frozen_history_is_excluded_without_renaming(self) -> None:
        # Historical names are evidence, not maintained naming or placement inputs.
        paths = (
            "evaluation/frozen-v2/helpers/test_issue42_20261116.py",
            "docs/adrs/2026-11-16-topic/round-1.md",
            "third_party/dependency/helpers/BadName.py",
        )
        for path in paths:
            self.fixture.write(path, "historical bytes\n")
        result = self.path_check(*paths)
        assert result.returncode == 0, result.stdout
        assert not result.stdout.strip(), result.stdout


if __name__ == "__main__":
    unittest.main(verbosity=2)

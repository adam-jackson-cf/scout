"""U3 process-boundary proofs for maintained Python policy and tool ratchets."""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from collections.abc import Iterator
from contextlib import contextmanager
from pathlib import Path

from test_inventory import ROOT, GateFixture

POLICY_PATH = "scripts/gate/structure.json"
CASE_PATH = "scripts/gate/policy_case.py"
GOOD_MODULE = '"""A maintained policy input."""\n\nvalue: int = 1\n'


@contextmanager
def policy_fixture() -> Iterator[GateFixture]:
    """Keep the real tools and checker, isolating every policy example in Git."""
    with tempfile.TemporaryDirectory(prefix="scout-python-canary-") as temporary:
        root = Path(temporary) / "repo"
        root.mkdir()
        fixture = GateFixture(root)
        policy = json.loads((root / POLICY_PATH).read_text())
        policy.setdefault("python", {}).update({
            "maintained_roots": [
                "scripts/gate/",
                "internal/association/testdata/gen/",
            ],
            "maintained_packages": [
                "scripts/gate/",
                "internal/association/testdata/gen/",
            ],
            "script_roots": ["scripts/gate/commands/"],
            "script_names": ["refresh-cache.py", "refresh_cache.py"],
            "script_name_exceptions": [],
            "test_roots": ["scripts/gate/tests/"],
            "test_groups": {"seam": ["scripts/gate/tests/seam/test_*.py"]},
            "generic_name_allowlist": [],
            "skill_roots": ["scripts/gate/skills/"],
            "non_package_dirs": ["commands", "skills"],
            "class_limits": {
                "max_lines": 700,
                "max_init_args": 8,
                "max_instance_attributes": 12,
            },
            "analyzer_roots": ["scripts/gate/analyzers/"],
            "analyzer_bases": ["AnalyzerBase"],
            "analyzer_entrypoints": {"AnalyzerBase": "analyze"},
            "process_roots": [CASE_PATH],
            "subprocess_wrappers": {CASE_PATH: ["run_process"]},
            "dynamic_import_roots": [CASE_PATH],
            "dynamic_import_adapters": [],
            "contract_files": [CASE_PATH],
            "analyzer_like_roots": ["scripts/gate/analyzers/"],
            "ownership_markers": ["scripts/gate/analyzers/OWNER"],
        })
        fixture.write(POLICY_PATH, json.dumps(policy, indent=2) + "\n")
        fixture.write("scripts/gate/analyzers/OWNER", "Repository gate maintainers\n")
        # Ratchets compare against an actual conforming parent, never an invented diff.
        fixture.git("add", ".")
        fixture.git("commit", "--quiet", "-m", "Declare Python fixture policy")
        yield fixture


def update_policy(fixture: GateFixture, **settings: object) -> None:
    policy = json.loads((fixture.root / POLICY_PATH).read_text())
    policy["python"].update(settings)
    fixture.write(POLICY_PATH, json.dumps(policy, indent=2) + "\n")


def python_gate(
    fixture: GateFixture, mode: str = "commit", *, format_inputs: bool = True
) -> subprocess.CompletedProcess[str]:
    if format_inputs:
        # These proofs isolate policy, not formatting (covered by the inventory suite).
        # Format only new fixture inputs, never the copied production checker modules.
        inventory = fixture.run(
            "git", "ls-files", "--others", "--exclude-standard", "-z"
        )
        assert inventory.returncode == 0, inventory.stdout
        inputs = [name for name in inventory.stdout.split("\0") if name.endswith(".py")]
        if inputs:
            formatted = fixture.run("ruff", "format", "--config", "ruff.toml", *inputs)
            assert formatted.returncode == 0, formatted.stdout
    return fixture.run("python", "-B", "scripts/gate/python.py", mode)


def require_accepted(result: subprocess.CompletedProcess[str]) -> None:
    assert result.returncode == 0, result.stdout


def require_rejected(
    result: subprocess.CompletedProcess[str], path: str, rule: str, fix: str
) -> None:
    assert result.returncode != 0, (
        f"missing {rule}: gate accepted {path}; expected fix containing {fix!r}:\n"
        f"{result.stdout}"
    )
    lines = [
        line for line in result.stdout.splitlines() if path in line and rule in line
    ]
    assert any(fix.casefold() in line.casefold() for line in lines), (
        f"missing {rule} diagnostic naming {path} and fix {fix!r}:\n{result.stdout}"
    )


def import_source(count: int, assignments: int = 0) -> str:
    imports = "\n".join(f"import dependency_{number:02d}" for number in range(count))
    references = ", ".join(f"dependency_{number:02d}" for number in range(count))
    body = "\n".join(f"value_{number}: int = {number}" for number in range(assignments))
    return f"{imports}\n\nimports = ({references})\n{body}\n"


class PythonStructureCanaries(unittest.TestCase):
    def test_root_command_group_and_name_policies(self) -> None:
        # Each path violates a different native ownership or naming contract.
        cases = (
            ("PYQ100", "scripts/unowned.py", "maintained"),
            ("PYQ101", "scripts/gate/commands/unapproved-command.py", "approve"),
            ("PYQ101", "scripts/gate/commands/refresh_cache.py", "kebab"),
            ("PYQ102", "scripts/gate/tests/unmapped/test_value.py", "group"),
            ("PYQ103", "scripts/gate/utils/value.py", "rename"),
            ("PYQ104", "scripts/gate/utils.py", "rename"),
            ("PYQ105", "scripts/gate/BadPackage/value.py", "snake_case"),
            ("PYQ106", "scripts/gate/skills/bad_skill/value.py", "kebab"),
            ("PYQ107", "scripts/gate/BadModule.py", "snake_case"),
        )
        for rule, path, fix in cases:
            with self.subTest(rule=rule, path=path), policy_fixture() as fixture:
                fixture.write(path, GOOD_MODULE)
                require_rejected(python_gate(fixture), path, rule, fix)

    def test_class_size_constructor_and_state_limits(self) -> None:
        # Cross each limit by one; constructor self is not a collaborator parameter.
        cases = (
            (
                "PYQ200",
                "class Payload:\n"
                + "    # Reserved field documentation.\n" * 699
                + "    pass\n",
                "split",
            ),
            (
                "PYQ201",
                "class Payload:\n"
                + "\n".join(
                    f"    def method_{n}(self) -> None:\n        pass\n"
                    for n in range(3)
                ),
                "split",
            ),
            (
                "PYQ202",
                "class Payload:\n    def __init__(self, *, "
                + ", ".join(f"value_{n}: int" for n in range(9))
                + ") -> None:\n        pass\n",
                "parameter object",
            ),
            (
                "PYQ203",
                "class Payload:\n    def __init__(self) -> None:\n"
                + "".join(f"        self.value_{n} = {n}\n" for n in range(13)),
                "value object",
            ),
        )
        for rule, source, fix in cases:
            with self.subTest(rule=rule), policy_fixture() as fixture:
                if rule == "PYQ201":
                    update_policy(fixture, class_limits={"max_public_methods": 2})
                fixture.write(CASE_PATH, source)
                require_rejected(python_gate(fixture), CASE_PATH, rule, fix)

    def test_analyzer_base_and_entrypoint_contracts(self) -> None:
        # An approved mixin does not excuse a missing project entrypoint.
        path = "scripts/gate/analyzers/policy.py"
        cases = (
            ("PYQ210", "class PolicyAnalyzer:\n    pass\n", "inherit"),
            (
                "PYQ211",
                "class AnalyzerBase:\n    pass\n\n\nclass PolicyAnalyzer(AnalyzerBase):\n    pass\n",
                "implement analyze",
            ),
        )
        for rule, source, fix in cases:
            with self.subTest(rule=rule), policy_fixture() as fixture:
                fixture.write(path, source)
                require_rejected(python_gate(fixture), path, rule, fix)

    def test_process_dynamic_import_and_collaborator_boundaries(self) -> None:
        # Aliased effects and annotations must not bypass the declared boundary.
        cases = (
            (
                "PYQ220",
                'import subprocess as sp\n\n\ndef launch() -> None:\n    sp.run(["true"], check=True, timeout=1)\n',
                "wrapper",
            ),
            (
                "PYQ221",
                'from importlib import import_module as load\n\n\ndef plugin() -> object:\n    return load("json")\n',
                "adapter",
            ),
            (
                "PYQ230",
                "from collections.abc import Callable\n\n\ndef apply(callback: Callable[[int], int]) -> int:\n    return callback(1)\n",
                "named callable",
            ),
            (
                "PYQ231",
                "from typing import Any\n\n\ndef apply(collaborator: Any) -> object:\n    return collaborator\n",
                "concrete",
            ),
        )
        for rule, source, fix in cases:
            with self.subTest(rule=rule), policy_fixture() as fixture:
                fixture.write(CASE_PATH, source)
                require_rejected(python_gate(fixture), CASE_PATH, rule, fix)


class PythonScalabilityCanaries(unittest.TestCase):
    def test_loop_and_comprehension_depth(self) -> None:
        # Both statement loops and comprehension generators count toward depth two.
        sources = (
            "def visit() -> None:\n    for first in range(2):\n        for second in range(first):\n            for third in range(second):\n                print(third)\n",
            "values = [(a, b, c) for a in range(2) for b in range(a) for c in range(b)]\n",
        )
        for source in sources:
            with self.subTest(source=source), policy_fixture() as fixture:
                fixture.write(CASE_PATH, source)
                require_rejected(python_gate(fixture), CASE_PATH, "PYS250", "extract")

    def test_analyzer_effects_require_owned_seams(self) -> None:
        # Filesystem, processes and environment reads are distinct effect classes.
        path = "scripts/gate/analyzers/policy.py"
        cases = (
            (
                "PYS251",
                'from pathlib import Path\n\n\ndef read() -> str:\n    return Path("input").read_text()\n',
            ),
            (
                "PYS252",
                'import subprocess as sp\n\n\ndef launch() -> None:\n    sp.run(["true"], check=True, timeout=1)\n',
            ),
            (
                "PYS252",
                'import os\n\n\ndef setting() -> str | None:\n    return os.getenv("SCOUT_INPUT")\n',
            ),
        )
        for rule, source in cases:
            with self.subTest(rule=rule, source=source), policy_fixture() as fixture:
                fixture.write(path, source)
                require_rejected(python_gate(fixture), path, rule, "seam")

    def test_analyzer_scope_requires_existing_ownership_marker(self) -> None:
        # Listing a nonexistent owner file is not evidence of ownership.
        with policy_fixture() as fixture:
            update_policy(
                fixture, ownership_markers=["scripts/gate/analyzers/MISSING_OWNER"]
            )
            require_rejected(python_gate(fixture), POLICY_PATH, "PYS253", "owner")

    def test_distinct_architectural_import_limits(self) -> None:
        # Standard-library imports and repeated edges do not consume architectural fanout.
        for rule, count, assignments in (
            ("FANOUT.general", 11, 0),
            ("FANOUT.large-module", 4, 800),
        ):
            with self.subTest(rule=rule), policy_fixture() as fixture:
                for number in range(count):
                    fixture.write(
                        f"scripts/gate/dependency_{number:02d}.py", GOOD_MODULE
                    )
                fixture.write(CASE_PATH, import_source(count, assignments))
                require_rejected(python_gate(fixture), CASE_PATH, rule, "split")


class PythonToolPolicyCanaries(unittest.TestCase):
    def test_new_package_cannot_escape_mypy(self) -> None:
        # Ruff accepts this source, but the current hardcoded mypy roots miss its type error.
        path = "scripts/added_package/value.py"
        with policy_fixture() as fixture:
            update_policy(
                fixture,
                maintained_roots=[
                    "scripts/gate/",
                    "internal/association/testdata/gen/",
                    "scripts/added_package/",
                ],
            )
            fixture.write(path, 'value: int = "unchecked"\n')
            require_rejected(
                python_gate(fixture), path, "MYPY.package-coverage", "package"
            )

    def test_package_inventory_and_import_root_drift(self) -> None:
        # A stale declared package and an omitted import root are independently rejected.
        for drift in ("package", "import-root"):
            with self.subTest(drift=drift), policy_fixture() as fixture:
                if drift == "package":
                    update_policy(
                        fixture,
                        maintained_packages=[
                            "scripts/gate/",
                            "internal/association/testdata/gen/",
                            "scripts/deleted_package/",
                        ],
                    )
                    path = POLICY_PATH
                else:
                    path = "mypy.ini"
                    config = (fixture.root / path).read_text()
                    fixture.write(
                        path, config.replace(":internal/association/testdata/gen", "")
                    )
                require_rejected(
                    python_gate(fixture), path, "MYPY.package-coverage", "package"
                )

    def test_mypy_overrides_are_scoped_and_evidence_backed(self) -> None:
        # A global exception and an unexplained named exception both weaken checking.
        for override in ("global", "unjustified"):
            with self.subTest(override=override), policy_fixture() as fixture:
                path = "mypy.ini"
                config = (fixture.root / path).read_text()
                if override == "global":
                    config = config.replace(
                        "[mypy]\n", "[mypy]\nignore_missing_imports = True\n"
                    )
                else:
                    config += "\n[mypy-calendar]\nignore_missing_imports = True\n"
                fixture.write(path, config)
                require_rejected(
                    python_gate(fixture), path, "MYPY.module-overrides", "override"
                )

    def test_ruff_suppressions_cannot_grow_without_approval(self) -> None:
        # Neither inline suppression nor a config-wide file exception is self-approving.
        for suppression in ("noqa", "per-file-ignores"):
            with self.subTest(suppression=suppression), policy_fixture() as fixture:
                if suppression == "noqa":
                    path = CASE_PATH
                    fixture.write(path, "import calendar  # noqa: F401\n")
                else:
                    path = "ruff.toml"
                    config = (fixture.root / path).read_text()
                    fixture.write(
                        path,
                        config.replace(
                            "[lint.per-file-ignores]\n",
                            '[lint.per-file-ignores]\n"scripts/gate/policy_case.py" = ["F401"]\n',
                        ),
                    )
                    fixture.write(CASE_PATH, "import calendar\n")
                require_rejected(python_gate(fixture), path, "RUFF.IGNORES", "approval")


class PythonPolicyControls(unittest.TestCase):
    def test_current_maintained_repository_files_pass_both_modes(self) -> None:
        # The proof must not require suppressions or edits to today's maintained sources.
        with policy_fixture() as fixture:
            for source in sorted((ROOT / "scripts/gate").glob("test_*.py")):
                if source.name != Path(__file__).name:
                    fixture.write("scripts/gate/" + source.name, source.read_text())
            generator = ROOT / "internal/association/testdata/gen"
            for source in sorted(generator.glob("*.py")):
                fixture.write(
                    "internal/association/testdata/gen/" + source.name,
                    source.read_text(),
                )
            for mode in ("commit", "full"):
                with self.subTest(mode=mode):
                    require_accepted(python_gate(fixture, mode, format_inputs=False))

    def test_conforming_boundaries_names_and_seams_pass(self) -> None:
        # At-limit state and two-deep loops remain valid; named callbacks and wrappers work.
        with policy_fixture() as fixture:
            fixture.write(
                CASE_PATH,
                'import subprocess\nfrom collections.abc import Callable\n\nCallback = Callable[[int], int]\n\n\ndef apply(callback: Callback) -> int:\n    return callback(1)\n\n\ndef run_process() -> None:\n    subprocess.run(["true"], check=True, timeout=1)\n',
            )
            fixture.write(
                "scripts/gate/analyzers/policy.py",
                "from collections.abc import Callable\n\nReadInput = Callable[[], str]\n\n\nclass AnalyzerBase:\n    pass\n\n\nclass PolicyAnalyzer(AnalyzerBase):\n    def analyze(self, read_input: ReadInput) -> str:\n        return read_input()\n",
            )
            fixture.write("scripts/gate/tests/seam/test_value.py", GOOD_MODULE)
            fixture.write("scripts/gate/commands/refresh-cache.py", GOOD_MODULE)
            fixture.write("scripts/gate/skills/valid-skill/value.py", GOOD_MODULE)
            fixture.write(
                "scripts/gate/limits.py",
                "class Payload:\n    def __init__(self) -> None:\n"
                + "".join(f"        self.value_{n} = {n}\n" for n in range(12))
                + "\n\ndef pairs() -> list[tuple[int, int]]:\n    return [(a, b) for a in range(2) for b in range(a)]\n",
            )
            require_accepted(python_gate(fixture))

    def test_fanout_thresholds_stdlib_and_duplicate_imports_pass(self) -> None:
        # Ten edges are valid normally, three in a large module; repeated imports count once.
        for count, assignments in ((10, 0), (3, 800)):
            with self.subTest(count=count), policy_fixture() as fixture:
                for number in range(count):
                    fixture.write(
                        f"scripts/gate/dependency_{number:02d}.py", GOOD_MODULE
                    )
                source = import_source(count, assignments)
                source += "\n\ndef stdlib() -> str:\n    import os\n\n    return os.name\n\n\ndef repeat() -> object:\n    import dependency_00\n\n    return dependency_00\n"
                fixture.write(CASE_PATH, source)
                require_accepted(python_gate(fixture))


if __name__ == "__main__":
    unittest.main(verbosity=2)

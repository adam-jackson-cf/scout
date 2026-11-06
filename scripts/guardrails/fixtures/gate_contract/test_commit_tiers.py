"""Exercise staged tier selection through the real declared commit command."""

from __future__ import annotations

import json
import tempfile
import unittest
from pathlib import Path

from scripts.guardrails.fixtures.gate_contract.test_gate_contract import (
    ROOT,
    command,
    copy_contract,
    git,
    initialise,
    write_json,
)


def write_package(root: Path, package: str, *, seam: bool) -> None:
    directory = root / "internal" / package
    directory.mkdir(parents=True)
    (directory / "value.go").write_text(
        f"package {package}\n\nfunc Value() int {{ return 1 }}\n"
    )
    extra_import = '\n\t"os"' if seam else ""
    assertion = (
        (
            '\n\tif err := os.WriteFile("'
            + str(root / "seam-ran")
            + '", []byte("ran"), 0600); err != nil { t.Fatal(err) }'
        )
        if seam
        else ""
    )
    (directory / "value_test.go").write_text(
        f'package {package}\n\nimport (\n\t"testing"{extra_import}\n)\n\n'
        f"func Test{package.title()}Contract(t *testing.T) {{\n"
        f'\tif Value() != 1 {{ t.Fatal("value contract") }}{assertion}\n}}\n'
    )
    result = command(["gofmt", "-w", str(directory)], root)
    if result.returncode:
        raise AssertionError(result.stderr)


def fixture_repository(root: Path) -> list[str]:
    initialise(root)
    copy_contract(root)
    (root / ".gitignore").write_text(".tools\nseam-ran\n")
    versions = (ROOT / "toolchain/versions.env").read_text()
    version = next(
        line.split("=", 1)[1]
        for line in versions.splitlines()
        if line.startswith("GO_VERSION=")
    )
    (root / "go.mod").write_text(
        f"module scout\n\ngo 1.26.0\n\ntoolchain go{version}\n"
    )
    write_package(root, "domain", seam=False)
    write_package(root, "artifacts", seam=True)
    write_json(
        root / ".omp/test-tiers.json",
        {
            "schemaVersion": 1,
            "owner": "Gate proof",
            "packages": {
                "internal/domain": {"tier": "fast", "owner": "Gate proof"},
                "internal/artifacts": {"tier": "seam", "owner": "Gate proof"},
            },
            "coverage": {"packages": ["internal/domain"], "minimumPercent": 100},
            "budget": {"wallMs": 159000, "perTestMs": 46000},
        },
    )
    write_json(root / ".omp/test-map.json", {"cases": {}})
    git(root, "add", ".")
    git(root, "commit", "-qm", "fixture baseline")
    git(root, "update-ref", "refs/remotes/origin/main", "HEAD")
    config = json.loads((ROOT / ".omp/gate.json").read_text())
    # The command under test is never replaced with a test-only tier runner.
    argv = config.get("checks", {}).get("commit")
    if not argv:
        raise AssertionError("P3: missing real checks.commit argv")
    return list(argv)


def stage_value(root: Path, package: str, value: int) -> None:
    path = root / "internal" / package / "value.go"
    text = path.read_text().replace("return 1", f"return {value}")
    path.write_text(text + "\n// staged proof change\n")
    git(root, "add", str(path.relative_to(root)))


class CommitTierProof(unittest.TestCase):
    def test_p3_t1_and_affected_seam(self) -> None:
        with tempfile.TemporaryDirectory(prefix="scout-tier-proof-") as directory:
            root = Path(directory)
            argv = fixture_repository(root)
            stage_value(root, "domain", 1)
            unrelated = command(argv, root)
            assert unrelated.returncode == 0, unrelated.stdout + unrelated.stderr
            assert not (root / "seam-ran").exists(), "P3: unrelated T2 seam ran"
            stage_value(root, "artifacts", 1)
            affected = command(argv, root)
            assert affected.returncode == 0, affected.stdout + affected.stderr
            assert (root / "seam-ran").exists(), "P3: affected T2 seam not executed"
            (root / "seam-ran").unlink()
            stage_value(root, "domain", 2)
            broken = command(argv, root)
            assert broken.returncode != 0, "P3: staged broken T1 accepted"
            assert "TestDomainContract" in broken.stdout + broken.stderr


if __name__ == "__main__":
    unittest.main()

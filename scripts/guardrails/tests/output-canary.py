"""Focused regressions for test classification and concise failure artifacts."""

import contextlib
import io
import json
import subprocess
import unittest
from pathlib import Path
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from scripts.guardrails.tests import check
else:
    import check


class OutputTests(unittest.TestCase):
    def test_script_only_change_does_not_introduce_tests(self) -> None:
        rows = [
            {"Path": "scripts/ci-toolchain.sh", "Tests": []},
            {"Path": "scripts/ci-sandbox.sh", "Tests": []},
        ]
        errors: list[str] = check.growth_errors(
            rows, [], {}, {row["Path"] for row in rows}
        )
        assert errors == []

    def test_unmapped_test_helper_still_fails(self) -> None:
        errors: list[str] = check.growth_errors(
            [{"Path": "internal/example/process_test.go", "Tests": []}], [], {}, set()
        )
        assert any("TEST.growth.justified" in error for error in errors)

    def test_failure_output_is_concise_and_full_json_is_retained(self) -> None:
        events = [
            {
                "Action": "output",
                "Package": "scout/internal/policy",
                "Test": "TestPass",
                "Output": "passing test noise\n",
            },
            {
                "Action": "output",
                "Package": "scout/internal/policy",
                "Test": "TestAbsentIsReportNothingActive",
                "Output": "=== RUN TestAbsentIsReportNothingActive\n    policy_test.go:7: expected report\n--- FAIL: TestAbsentIsReportNothingActive (0.00s)\n",
            },
            {
                "Action": "fail",
                "Package": "scout/internal/policy",
                "Test": "TestAbsentIsReportNothingActive",
            },
            {
                "Action": "output",
                "Package": "scout/internal/policy",
                "Output": "FAIL\tscout/internal/policy\t0.001s\n",
            },
            {"Action": "fail", "Package": "scout/internal/policy"},
        ]
        raw = "\n".join(json.dumps(event) for event in events) + "\n"
        result = subprocess.CompletedProcess(
            ["go", "test", "-json"], 1, raw, "compiler diagnostic\n"
        )
        output = io.StringIO()
        with contextlib.redirect_stdout(output):
            check.report_failure(result)
        text = output.getvalue()
        assert "passing test noise" not in text
        assert '"Action"' not in text
        assert "FAILED scout/internal/policy/TestAbsentIsReportNothingActive" in text
        assert "policy_test.go:7: expected report" in text
        assert "--- FAIL: TestAbsentIsReportNothingActive" in text
        assert "FAIL\tscout/internal/policy" in text
        marker = next(
            line for line in text.splitlines() if line.startswith("go test JSON: ")
        )
        artifact = Path(marker.removeprefix("go test JSON: "))
        stderr = Path(str(artifact) + ".stderr")
        try:
            assert artifact.is_absolute()
            assert artifact.read_text() == raw
            assert stderr.read_text() == "compiler diagnostic\n"
        finally:
            artifact.unlink()
            stderr.unlink()


if __name__ == "__main__":
    unittest.main()

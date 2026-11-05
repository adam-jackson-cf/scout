#!/usr/bin/env python3
"""Exercise production test checkers with passing and violating fixtures."""

import json
import sys
import tempfile
from pathlib import Path
from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from scripts.guardrails.tests import check
else:
    import check


def check_fixture(fixture: dict[str, Any]) -> list[str]:
    rule = fixture["rule"]
    errors: list[str]
    if rule.startswith("TEST.budget."):
        errors = check.budget_errors(fixture["report"], fixture["limits"])
        return errors
    if rule == "TEST.tiers":
        errors = check.tier_errors(fixture["packages"], fixture["policy"])
        return errors
    if rule == "TEST.coverage-minimum":
        errors = check.coverage_errors(fixture["profile"], fixture["policy"])
        return errors
    with tempfile.TemporaryDirectory(prefix="scout-test-canary-") as tmp:
        executable = Path(tmp) / "inspect"
        check.run([
            "go",
            "build",
            "-o",
            str(executable),
            "scripts/guardrails/inspect.go",
        ])
        rows = []
        for name in ("previous", "current"):
            path = Path(tmp) / (name + "_test.go")
            path.write_text(fixture[name])
            parsed = check.inspect(executable, [str(path)])
            parsed[0]["Path"] = "internal/example/example_test.go"
            rows.append(parsed)
        if rule == "TEST.subprocess-timeouts":
            errors = check.timeout_errors(rows[1])
            return errors
        errors = check.growth_errors(
            rows[1],
            rows[0],
            fixture.get("mapping", {}),
            set(fixture.get("changed", [])),
        )
        return errors


def main() -> bool:
    fixture = json.loads(Path(sys.argv[1]).read_text())
    errors = check_fixture(fixture)
    print("\n".join(errors))
    return bool(errors)


if __name__ == "__main__":
    sys.exit(main())

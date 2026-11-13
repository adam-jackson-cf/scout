#!/usr/bin/env python3
"""Run Scout's Go test tiers and enforce the declared coverage floor."""

import argparse
import json
import subprocess
import sys
import tempfile
import time
from pathlib import Path
from typing import Any


def run(
    argv: list[str], *, check: bool = True, timeout: int | None = None
) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        argv, check=check, text=True, capture_output=True, timeout=timeout
    )


def error(path: str, rule: str, fix: str) -> str:
    return f"{path}: {rule}: {fix}"


def tier_errors(packages: list[str], policy: dict[str, Any]) -> list[str]:
    return [
        error(
            p,
            "test-tiers",
            "declare tier fast or seam in scripts/gate/test-tiers.json",
        )
        for p in packages
        if policy["packages"].get(p, {}).get("tier") not in ("fast", "seam")
    ]


def coverage_errors(profile: str, policy: dict[str, Any]) -> list[str]:
    blocks: dict[str, tuple[int, bool]] = {}
    owners = {"scout/" + owner for owner in policy["packages"]}
    for line in profile.splitlines()[1:]:
        location, statements, count = line.split()
        filename = location.split(":", 1)[0]
        if filename.rsplit("/", 1)[0] not in owners:
            continue
        size, executions = int(statements), int(count)
        if size < 0 or executions < 0:
            raise ValueError("negative coverage measurement")
        previous_size, previous_covered = blocks.get(location, (size, False))
        if previous_size != size:
            raise ValueError("inconsistent statement count for coverage block")
        blocks[location] = (size, previous_covered or executions > 0)
    total = sum(size for size, _ in blocks.values())
    covered = sum(size for size, hit in blocks.values() if hit)
    percent = 100 * covered / total if total else 0
    print(f"coverage: {covered}/{total} statements ({percent:.6f}%)")
    return (
        []
        if total and percent >= policy["minimumPercent"]
        else [
            error(
                ",".join(policy["packages"]),
                "coverage-minimum",
                f"cover missing consumer-visible behavior: {percent:.6f}% < {policy['minimumPercent']}%",
            )
        ]
    )


def tested_packages() -> list[str]:
    return sorted({
        str(path.parent)
        for owner in ("cmd", "internal")
        for path in Path(owner).rglob("*_test.go")
    })


def failure_lines(events: list[dict[str, Any]]) -> list[str]:
    failures = {
        (event.get("Package"), event.get("Test"))
        for event in events
        if event.get("Action") == "fail" and event.get("Test")
    }
    lines = [f"FAILED {package}/{test}" for package, test in sorted(failures)]
    for event in events:
        if (event.get("Package"), event.get("Test")) in failures:
            lines.extend(
                line
                for line in event.get("Output", "").splitlines()
                if not line.startswith(("=== RUN", "=== PAUSE", "=== CONT"))
            )
        elif not event.get("Test"):
            lines.extend(
                line
                for line in event.get("Output", "").splitlines()
                if line.startswith("FAIL")
            )
    lines.extend(
        f"FAIL {event['Package']}"
        for event in events
        if event.get("Action") == "fail" and not event.get("Test")
    )
    return list(dict.fromkeys(lines))


def report_failure(result: subprocess.CompletedProcess[str]) -> None:
    with tempfile.NamedTemporaryFile(
        mode="w", prefix="scout-go-test-", suffix=".jsonl", delete=False
    ) as artifact:
        artifact.write(result.stdout)
        path = Path(artifact.name).resolve()
    if result.stderr:
        Path(str(path) + ".stderr").write_text(result.stderr)
    events = [json.loads(line) for line in result.stdout.splitlines()]
    print("\n".join(failure_lines(events)))
    print(f"go test JSON: {path}")


def execute(packages: list[str], tier: str, policy: dict[str, Any]) -> list[str]:
    errors = []
    with tempfile.TemporaryDirectory(prefix="scout-coverage-") as tmp:
        profile = Path(tmp) / "coverage.out"
        argv = ["go", "test", "-json", "-count=1", "-timeout=150s"]
        coverage = tier == "fast"
        if coverage:
            argv += [
                "-coverprofile=" + str(profile),
                "-coverpkg="
                + ",".join("./" + p for p in policy["coverage"]["packages"]),
            ]
        argv += ["./" + p for p in packages]
        start = time.monotonic()
        result = run(argv, check=False, timeout=600)
        elapsed = (time.monotonic() - start) * 1000
        tests = []
        failed = result.returncode != 0
        for line in result.stdout.splitlines():
            event = json.loads(line)
            failed |= event.get("Action") == "fail"
            if event.get("Action") == "pass" and event.get("Test"):
                tests.append({
                    "id": event["Package"] + "/" + event["Test"],
                    "durationMs": event.get("Elapsed", float("nan")) * 1000,
                })
        if failed:
            report_failure(result)
            return [
                error(
                    tier,
                    "go-test",
                    "repair failed Go test prerequisite",
                )
            ]
        print(f"{tier}: wallMs={elapsed:.3f}; passing events={len(tests)}")
        if coverage:
            errors += coverage_errors(profile.read_text(), policy["coverage"])
    return errors


def affected_package(package: str, paths: list[str]) -> bool:
    """Select a package when a path is in its subtree or is a module-wide input."""
    return any(
        path
        in {
            "go.mod",
            "go.sum",
            "scripts/gate/test-tiers.json",
            "toolchain/versions.env",
        }
        or path == package
        or path.startswith(package + "/")
        for path in paths
    )


def main() -> bool:
    parser = argparse.ArgumentParser()
    parser.add_argument("mode", choices=["full", "commit"])
    parser.add_argument("paths", nargs="*")
    args = parser.parse_intermixed_args()
    policy = json.loads(
        (Path(__file__).resolve().parent / "test-tiers.json").read_text()
    )
    errors = tier_errors(tested_packages(), policy)
    if not errors:
        for tier in ["fast", "seam"]:
            packages = [
                name
                for name, entry in policy["packages"].items()
                if entry["tier"] == tier
                and (
                    args.mode == "full"
                    or tier == "fast"
                    or affected_package(name, args.paths)
                )
            ]
            if packages:
                errors += execute(packages, tier, policy)
    print("\n".join(errors))
    return bool(errors)


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (OSError, ValueError, KeyError, subprocess.SubprocessError) as exc:
        print(error("test checker", "go-test", str(exc)), file=sys.stderr)
        sys.exit(1)

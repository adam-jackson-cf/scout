#!/usr/bin/env python3
"""Owned Go test tiers, AST growth/duplicates/timeouts and measured coverage/budgets."""

import argparse
import json
import math
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
            "TEST.tiers",
            "declare fast/seam and owning commit/CI path in .omp/test-tiers.json",
        )
        for p in packages
        if p not in policy["packages"]
        or policy["packages"][p].get("tier") not in ("fast", "seam")
        or not policy["packages"][p].get("owner")
    ]


def budget_errors(report: dict[str, Any], limits: dict[str, Any]) -> list[str]:
    errors = []
    wall = report.get("wallMs")
    if (
        not isinstance(wall, (int, float))
        or isinstance(wall, bool)
        or not math.isfinite(wall)
        or wall < 0
        or wall > limits["wallMs"]
    ):
        errors.append(
            error(
                "fast tier",
                "TEST.budget.fast-tier-time",
                f"record valid wall <= {limits['wallMs']}ms; remove needless work",
            )
        )
    tests = report.get("tests", [])
    if not tests:
        errors.append(
            error(
                "go test",
                "TEST.budget.per-test-time",
                "record nonempty successful case measurements",
            )
        )
    for test in tests:
        value = test.get("durationMs")
        if (
            not test.get("id")
            or not isinstance(value, (int, float))
            or isinstance(value, bool)
            or not math.isfinite(value)
            or value < 0
            or value > limits["perTestMs"]
        ):
            errors.append(
                error(
                    test.get("id", "unknown"),
                    "TEST.budget.per-test-time",
                    f"record valid case <= {limits['perTestMs']}ms; inject clock/runner or remove needless work, never move tier",
                )
            )
    return errors


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
                "TEST.coverage-minimum",
                f"cover missing consumer-visible behavior: {percent:.6f}% < {policy['minimumPercent']}%",
            )
        ]
    )


def row_cases(row: dict[str, Any]) -> dict[str, Any]:
    result = {}
    seen: dict[str, int] = {}
    for test in row.get("Tests") or []:
        for case in test["Cases"]:
            name = row["Path"] + "/" + case["Name"]
            seen[name] = seen.get(name, 0) + 1
            identity = (
                name if seen[name] == 1 else name + "/duplicate-" + str(seen[name])
            )
            result[identity] = case
    return result


def cases(rows: list[dict[str, Any]]) -> dict[str, Any]:
    result = {}
    for row in rows:
        result.update(row_cases(row))
    return result


def is_test_path(path: str) -> bool:
    return path.endswith("_test.go") and Path(path).parts[0] in ("cmd", "internal")


def growth_errors(
    current: list[dict[str, Any]],
    previous: list[dict[str, Any]],
    mapping: dict[str, Any],
    changed: set[str],
) -> list[str]:
    # Inspector output is shared with non-test policy checks. Only colocated Go
    # test files can introduce test identities, including helper-only test files.
    current = [row for row in current if is_test_path(row["Path"])]
    previous = [row for row in previous if is_test_path(row["Path"])]
    errors = []
    old, new = cases(previous), cases(current)
    fingerprints: dict[str, list[str]] = {}
    for identity, case in {**old, **new}.items():
        fingerprints.setdefault(case["Fingerprint"], []).append(identity)
    additions = [(name, case) for name, case in new.items() if name not in old]
    old_paths = {row["Path"] for row in previous}
    additions += [
        (row["Path"], None)
        for row in current
        if row["Path"] not in old_paths and not row.get("Tests")
    ]
    for identity, case in additions:
        entry = mapping.get(identity, {})
        sources = entry.get("source", [])
        if (
            not entry.get("behavior", "").strip()
            or not sources
            or any(
                source not in changed or source.endswith("_test.go")
                for source in sources
            )
        ):
            errors.append(
                error(
                    identity,
                    "TEST.growth.justified",
                    "map changed source and behavior in .omp/test-map.json or remove case",
                )
            )
        if case:
            duplicates = [
                other
                for other in fingerprints[case["Fingerprint"]]
                if other != identity
            ]
            if duplicates:
                errors.append(
                    error(
                        identity,
                        "TEST.duplicate-path",
                        f"duplicates {duplicates[0]}; merge or remove one of both tests",
                    )
                )
    return errors


def timeout_errors(rows: list[dict[str, Any]]) -> list[str]:
    errors = []
    helpers = {
        str(Path(row["Path"]).parent) for row in rows if row.get("BoundedHelper")
    }
    for row in rows:
        if row.get("UsesHelper") and str(Path(row["Path"]).parent) not in helpers:
            errors.append(
                error(
                    row["Path"],
                    "TEST.subprocess-timeouts",
                    "provide parser-verified boundedTestContext with timeout, cancellation cleanup and returned context",
                )
            )
        for call in row.get("Unbounded") or []:
            errors.append(
                error(
                    row["Path"],
                    "TEST.subprocess-timeouts",
                    f"{call}: use CommandContext with timeout/deadline or the owned boundedTestContext; retain failure output",
                )
            )
    return errors


def inspect(
    executable: Path, paths: list[str] | dict[str, str]
) -> list[dict[str, Any]]:
    result: list[dict[str, Any]] = (
        json.loads(run([str(executable), *map(str, paths)]).stdout) if paths else []
    )
    return result


def static_checks(base: str, executable: Path, policy: dict[str, Any]) -> list[str]:
    files = sorted(
        str(path)
        for owner in ("cmd", "internal")
        for path in Path(owner).rglob("*_test.go")
    )
    current = inspect(executable, files)
    changed = set(run(["git", "diff", "--name-only", base, "--"]).stdout.splitlines())
    changed.update(
        run(["git", "ls-files", "--others", "--exclude-standard"]).stdout.splitlines()
    )
    previous = []
    with tempfile.TemporaryDirectory(prefix="scout-test-base-") as tmp:
        old_files = {}
        for name in files:
            result = run(["git", "show", base + ":" + name], check=False)
            if result.returncode == 0:
                target = Path(tmp) / name
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_text(result.stdout)
                old_files[str(target)] = name
        previous = inspect(executable, old_files)
        for row in previous:
            row["Path"] = old_files[row["Path"]]
    mapping = json.loads(Path(".omp/test-map.json").read_text())["cases"]
    packages = sorted({str(Path(name).parent) for name in files})
    return (
        tier_errors(packages, policy)
        + growth_errors(current, previous, mapping, changed)
        + timeout_errors(current)
    )


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
                    "TEST.failure-propagation",
                    "repair failed Go test prerequisite",
                )
            ]
        print(f"{tier}: wallMs={elapsed:.3f}; passing events={len(tests)}")
        if tier == "fast":
            errors += budget_errors(
                {"wallMs": elapsed, "tests": tests}, policy["budget"]
            )
        if coverage:
            errors += coverage_errors(profile.read_text(), policy["coverage"])
    return errors


def affected_package(package: str, paths: list[str]) -> bool:
    """A package owns its source subtree; module-wide inputs affect every seam."""
    return any(
        path in {"go.mod", "go.sum", ".omp/test-tiers.json", "toolchain/versions.env"}
        or path == package
        or path.startswith(package + "/")
        for path in paths
    )


def main() -> bool:
    parser = argparse.ArgumentParser()
    parser.add_argument("mode", choices=["full", "commit"])
    parser.add_argument("--base", default="origin/main")
    parser.add_argument("paths", nargs="*")
    args = parser.parse_intermixed_args()
    policy = json.loads(Path(".omp/test-tiers.json").read_text())
    with tempfile.TemporaryDirectory(prefix="scout-test-inspect-") as tmp:
        executable = Path(tmp) / "inspect"
        inspector = Path(__file__).resolve().parents[1] / "inspect.go"
        run(["go", "build", "-o", str(executable), str(inspector)])
        errors = static_checks(args.base, executable, policy)
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
        print(
            error("test checker", "TEST.failure-propagation", str(exc)), file=sys.stderr
        )
        sys.exit(1)

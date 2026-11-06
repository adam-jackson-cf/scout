"""Target-owned configuration and maintained-package coverage contracts."""

from __future__ import annotations

import configparser
import fnmatch
import json
import re
import tomllib
from pathlib import Path
from typing import Any

SETTINGS = Path(__file__).with_name("settings.json")
MYPY = {
    "MYPY.SETTING.show_error_codes": ("show_error_codes", "True"),
    "MYPY.follow_imports": ("follow_imports", "skip"),
    "MYPY.strict": ("strict", "True"),
    "MYPY.warn_unused_ignores": ("warn_unused_ignores", "True"),
}


def target_policy() -> dict[str, Any]:
    """Load Scout's maintained inventory and exception evidence."""
    from python_quality_scope import discover_maintained_python_files

    target: dict[str, Any] = json.loads(SETTINGS.read_text())["target"]
    root = SETTINGS.parents[3]
    target["maintained_sources"] = sorted(
        set(target["maintained_sources"])
        | {
            path.relative_to(root).as_posix()
            for path in discover_maintained_python_files(root)
        }
    )
    return target


def read_ruff(path: Path) -> dict[str, Any]:
    with path.open("rb") as stream:
        return tomllib.load(stream)


def excluded_path(path: str, pattern: str) -> bool:
    """Match Ruff's root-relative paths and basename exclusion globs."""
    pattern = pattern.rstrip("/")
    return any(
        fnmatch.fnmatchcase(candidate, pattern)
        for candidate in [path, *Path(path).parts]
    ) or path.startswith(pattern + "/")


def ruff_scope_valid(patterns: list[str], target: dict[str, Any]) -> bool:
    defaults = json.loads(SETTINGS.read_text())["ruff_default_exclusions"]
    rationales = target["exclusion_rationales"]
    return (
        set(defaults).issubset(patterns)
        and all(
            pattern in defaults or rationales.get(pattern.rstrip("/"), "").strip()
            for pattern in patterns
        )
        and not any(
            excluded_path(source, pattern)
            for source in target["maintained_sources"]
            for pattern in patterns
        )
    )


def ruff_setting_valid(
    rule: str, value: Any, setting: dict[str, Any], target: dict[str, Any]
) -> bool:
    if rule == "RUFF.ISORT.known-first-party":
        return isinstance(value, list) and set(value) == set(
            target["maintained_first_party_names"]
        )
    if rule == "RUFF.ROOT.exclude":
        return isinstance(value, list) and ruff_scope_valid(value, target)
    if rule == "RUFF.SETTING.target-version":
        return value in [
            "py" + version.replace(".", "")
            for version in target["supported_python_versions"]
        ]
    if rule.startswith("RUFF.SELECT."):
        return setting["value"] in (value or [])
    return bool(value == setting["value"])


def ruff_errors(
    config: dict[str, Any], target: dict[str, Any] | None = None
) -> list[str]:
    target = target_policy() if target is None else target
    errors: list[str] = []
    settings: dict[str, Any] = json.loads(SETTINGS.read_text())
    for rule, setting in settings.items():
        if not rule.startswith("RUFF."):
            continue
        value: Any = config
        for key in setting["key"].split("."):
            value = value.get(key) if isinstance(value, dict) else None
        valid = ruff_setting_valid(rule, value, setting, target)
        if rule == "RUFF.IGNORES":
            valid = (
                valid
                and config.get("lint", {}).get("per-file-ignores")
                == setting["per_file"]
                and not (
                    {"extend-ignore", "extend-per-file-ignores"}
                    & config.get("lint", {}).keys()
                )
            )
        if not valid:
            errors.append(rule)
    return errors


def mypy_exclusions_valid(pattern: str, target: dict[str, Any]) -> bool:
    try:
        expression = re.compile(pattern)
    except re.error:
        return False
    rationales = target["exclusion_rationales"]
    # Every regex branch must be an explicit, anchored non-maintained subtree.
    branches = pattern.removeprefix("(").removesuffix(")").split("|")
    paths = [
        branch.removeprefix("^").rstrip("/").replace("\\.", ".") for branch in branches
    ]
    return (
        bool(pattern)
        and all(
            branch.startswith("^")
            and branch.endswith("/")
            and rationales.get(path, "").strip()
            and branch == "^" + re.escape(path).replace("\\/", "/") + "/"
            for branch, path in zip(branches, paths, strict=True)
        )
        and not any(
            expression.search(source) for source in target["maintained_sources"]
        )
    )


def overrides_valid(config: configparser.ConfigParser, target: dict[str, Any]) -> bool:
    evidence = target["module_override_evidence"]
    for section in config.sections():
        if section == "mypy":
            continue
        module = section.removeprefix("mypy-")
        record = evidence.get(module, {})
        if not section.startswith("mypy-") or not all(
            part.isidentifier() for part in module.split(".")
        ):
            return False
        if (
            record.get("kind") not in {"missing-stub", "decorator"}
            or not record.get("evidence", "").strip()
        ):
            return False
        expected = (
            {"ignore_missing_imports": "true"}
            if record["kind"] == "missing-stub"
            else {"disable_error_code": "untyped-decorator"}
        )
        actual = {key: value.strip().lower() for key, value in config[section].items()}
        if actual != expected:
            return False
    return True


def mypy_errors(
    config: configparser.ConfigParser, target: dict[str, Any] | None = None
) -> list[str]:
    target = target_policy() if target is None else target
    errors = [
        rule
        for rule, (key, value) in MYPY.items()
        if config.get("mypy", key, fallback="") != value
    ]
    if not mypy_exclusions_valid(config.get("mypy", "exclude", fallback=""), target):
        errors.append("MYPY.SETTING.exclude")
    roots = config.get("mypy", "mypy_path", fallback="").replace(",", ":").split(":")
    if {root.strip().rstrip("/") for root in roots} != set(
        target["maintained_import_roots"]
    ):
        errors.append("MYPY.SETTING.mypy_path")
    if (
        config.get("mypy", "python_version", fallback="")
        not in target["supported_python_versions"]
    ):
        errors.append("MYPY.SETTING.python_version")
    if not overrides_valid(config, target):
        errors.append("MYPY.module-overrides")
    if config.has_option("mypy", "files") or config.has_option("mypy", "packages"):
        errors.append("MYPY.package-coverage")
    allowed = {key for key, _value in MYPY.values()} | {
        "exclude",
        "mypy_path",
        "python_version",
    }
    if config.has_section("mypy") and set(config["mypy"]) != allowed:
        errors.append("MYPY.strict")
    return errors


def coverage_errors(
    maintained: list[str],
    checked: list[str],
    target: dict[str, Any] | None = None,
    packages: list[str] | None = None,
) -> list[str]:
    target = target_policy() if target is None else target
    checked_packages = (
        {Path(path).parent.as_posix().replace("/", ".") for path in checked}
        if packages is None
        else set(packages)
    )
    valid = (
        set(target["maintained_sources"]) == set(maintained) == set(checked)
        and set(target["maintained_packages"]) == checked_packages
    )
    return [] if valid else ["MYPY.package-coverage"]

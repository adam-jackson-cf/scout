"""Exact configuration and maintained-file coverage assertions."""

from __future__ import annotations

import configparser
import json
import tomllib
from pathlib import Path
from typing import Any

SETTINGS = Path(__file__).with_name("settings.json")
MYPY = {
    "MYPY.SETTING.exclude": (
        "exclude",
        r"(^evaluation/|^\.tools/|^scripts/guardrails/fixtures/|^scripts/guardrails/vendor/)",
    ),
    "MYPY.SETTING.mypy_path": (
        "mypy_path",
        "scripts/guardrails/python:scripts/guardrails",
    ),
    "MYPY.SETTING.python_version": ("python_version", "3.11"),
    "MYPY.SETTING.show_error_codes": ("show_error_codes", "True"),
    "MYPY.follow_imports": ("follow_imports", "skip"),
    "MYPY.strict": ("strict", "True"),
    "MYPY.warn_unused_ignores": ("warn_unused_ignores", "True"),
}


def read_ruff(path: Path) -> dict[str, Any]:
    with path.open("rb") as stream:
        return tomllib.load(stream)


def ruff_errors(config: dict[str, Any]) -> list[str]:
    errors: list[str] = []
    settings: dict[str, Any] = json.loads(SETTINGS.read_text())
    for rule, setting in settings.items():
        value: Any = config
        for key in setting["key"].split("."):
            value = value.get(key) if isinstance(value, dict) else None
        valid = (
            setting["value"] in (value or [])
            if rule.startswith("RUFF.SELECT.")
            else value == setting["value"]
        )
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


def mypy_errors(config: configparser.ConfigParser) -> list[str]:
    errors = [
        rule
        for rule, (key, value) in MYPY.items()
        if config.get("mypy", key, fallback="") != value
    ]
    if config.sections() != ["mypy", "mypy-lizard"] or dict(config["mypy-lizard"]) != {
        "ignore_missing_imports": "True"
    }:
        errors.append("MYPY.module-overrides")
    if config.has_option("mypy", "files") or config.has_option("mypy", "packages"):
        errors.append("MYPY.package-coverage")
    if config.has_section("mypy") and set(config["mypy"]) != {
        key for key, _value in MYPY.values()
    }:
        errors.append("MYPY.strict")
    return errors


def coverage_errors(maintained: list[str], checked: list[str]) -> list[str]:
    return [] if set(maintained) == set(checked) else ["MYPY.package-coverage"]

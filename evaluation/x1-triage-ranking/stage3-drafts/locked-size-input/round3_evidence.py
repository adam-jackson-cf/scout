#!/usr/bin/env python3
"""Synthetic extraction derivation, focused mutants and timing; run explicitly."""
import argparse
import hashlib
import inspect
import json
from pathlib import Path
import time
from unittest.mock import patch

import locked_commands as commands
from test_size_input import LockedModeTests, measurement_fixture


def independent_answers():
    # Literal fixture quantities, not production extraction or resource arithmetic.
    units = 3 * 20 * 50
    return {
        "construction_cpu_per_unit": 6 / units,
        "requested_base_cpu_per_unit": 9 / units,
        "calibration_cpu_per_unit": 12 / units,
        "enumeration_cpu_per_population": 4.,
        "gate_cpu_per_check": 25000 / 100000,
        "complete_trial_cpu": .03,
        "full512_trial_cpu": .07,
        "construction_rss": 3000,
        "calibration_rss": 4000,
        "trial_plus_helper_rss": 5000,
    }


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", required=True, type=Path)
    args = parser.parse_args()
    if args.output.exists():
        raise FileExistsError("synthetic receipts are create-once")
    source, data, calibration, gate = measurement_fixture()
    function = commands.measurement_run
    original = inspect.getsource(function)
    mutants = {
        "construction-minimum": (
            'prep = max(r["construction"]["cpu_seconds"] for r in data["population_measurements"])',
            'prep = min(r["construction"]["cpu_seconds"] for r in data["population_measurements"])'),
        "construction-first-row": (
            'prep = max(r["construction"]["cpu_seconds"] for r in data["population_measurements"])',
            'prep = data["population_measurements"][0]["construction"]["cpu_seconds"]'),
        "full512-uses-complete": ('"full512_trial_cpu": worst[True]', '"full512_trial_cpu": worst[False]'),
        "gate-divisor-dropped": ('r["size_only_sampling_gate"]["cpu_seconds"]/100000',
                                 'r["size_only_sampling_gate"]["cpu_seconds"]'),
        "calibration-frame-dropped": (
            '+ [max(r["frame_evaluation_cpu_seconds"]) for r in calibration["measurements"]]', '+ []'),
        "fixed-key-dropped": (
            '[e["cpu_seconds"] for r in data["population_measurements"] for e in r["fixed_key_evaluations"]]',
            '[]'),
    }
    actual = function(source, data, calibration, gate, {"synthetic": "1" * 64})["binding_inputs"]
    expected = independent_answers()
    if actual != expected:
        raise AssertionError((actual, expected))
    reports = []
    for name, (old, new) in mutants.items():
        if original.count(old) != 1:
            raise RuntimeError(f"mutant site drift: {name}")
        mutated = original.replace(old, new)
        namespace = dict(vars(commands))
        exec(compile(mutated, f"<synthetic-mutant:{name}>", "exec"), namespace)
        killed = False
        with patch.object(commands, "measurement_run", namespace["measurement_run"]):
            try:
                LockedModeTests("test_extraction_uses_locked_maxima_not_slopes").test_extraction_uses_locked_maxima_not_slopes()
            except AssertionError:
                killed = True
        reports.append({"mutant": name, "killed": killed,
                        "mutated_function_sha256": hashlib.sha256(mutated.encode()).hexdigest()})
    timings = []
    for _ in range(5):
        start = time.perf_counter()
        function(source, data, calibration, gate, {"synthetic": "1" * 64})
        timings.append(time.perf_counter() - start)
    receipt = {
        "schema": "x1-locked-size-round3-checks/synthetic-v1", "synthetic_only": True,
        "independent_inputs": expected, "extracted_inputs": actual,
        "mutants": reports, "extraction_wall_seconds": timings,
        "timing_scope": "Five fabricated-record extractions, not scientific benchmark evidence or binding measurements",
        "source_sha256": {name: hashlib.sha256((commands.HERE / name).read_bytes()).hexdigest()
                          for name in ("locked_commands.py", "test_size_input.py", "round3_evidence.py")},
        **commands.interpreter(),
    }
    commands.publish(args.output, receipt)
    print(json.dumps(receipt, sort_keys=True, allow_nan=False))
    if not all(report["killed"] for report in reports):
        raise RuntimeError("surviving extraction mutant; receipt preserved")


if __name__ == "__main__":
    main()

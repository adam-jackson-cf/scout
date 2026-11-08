#!/usr/bin/env python3
"""Runnable synthetic-only locked-size benchmark, calibration, gate and fit modes.

Signed files remain byte-for-byte unchanged. Guarded size-site substitutions
compile the signed function bodies, retaining all non-size scientific code.
"""
import argparse
import ast
from concurrent.futures import ProcessPoolExecutor
import hashlib
import importlib.util
import inspect
import itertools
import json
import math
import os
from pathlib import Path
import platform
import subprocess
import sys
import tempfile
import time

from size_input import SizeInput, REPOSITORIES, canonical, synthetic_record, CONTROLLER_PATH
from population import prepare_population

HERE = Path(__file__).resolve().parent
VALIDATION = HERE.parents[1] / "validation"
ROOT = HERE.parents[3]


def load_file(name, path):
    spec = importlib.util.spec_from_file_location(name, path)
    module = importlib.util.module_from_spec(spec)
    sys.modules[name] = module
    spec.loader.exec_module(module)
    return module


# population imported the signed numerical module directory before these imports.
benchmark = load_file("x1_signed_size_benchmark", VALIDATION / "compute_benchmark.py")
reporter = load_file("x1_signed_size_reporter", VALIDATION / "reproduce_reports.py")


# Every replacement is an exact AST subtree, required exactly once. No fallback
# to running the planning size source if signed source shape changes.
SIZE_SITES = {
    "benchmark_population": (
        ("(2000,)*(L//4)+(500,)*(L//2)+(1000,)*(L//4)", "size_input"),
        ("prepare_population(seed=seed_for('independent-population', L), prevalence=.55, heterogeneity=.10, overlap=.5, run_count=L, sizes=sizes, cap=400, failure=.02, rank_structure='linear-gradient', size_dependence=.05)",
         "locked_prepare(seed=seed_for('independent-population', L), prevalence=.55, heterogeneity=.10, overlap=.5, size_input=size_input, cap=400, failure=.02, rank_structure='linear-gradient', size_dependence=.05)"),
        ("3*sum(sizes)", "size_input.units"),
        ("[sum(rng.sample(sizes, 20)) for _ in range(3)]",
         "[sum(size_input.vectors[r][i] for i in rng.sample(range(len(size_input.vectors[r])), 20)) for r in repositories]"),
        ("sum(sizes)/L", "size_input.units/sum(size_input.lengths.values())"),
        ("max(sizes)", "size_input.maximum"),
    ),
    "calibration_measurement": (
        ("(2000,)*(L//4)+(500,)*(L//2)+(1000,)*(L//4)", "size_input"),
        ("kernel.prepare_population(seed=202611080000+L, prevalence=.55, heterogeneity=.10, overlap=.5, run_count=L, sizes=sizes, cap=400, failure=.02, rank_structure='linear-gradient', size_dependence=.05)",
         "locked_prepare(seed=202611080000+L, prevalence=.55, heterogeneity=.10, overlap=.5, size_input=size_input, cap=400, failure=.02, rank_structure='linear-gradient', size_dependence=.05)"),
        ("3*sum(sizes)", "size_input.units"),
    ),
    "trial_frame": (("range(frame_runs)", "range(size_input.lengths[repo])"),),
    "execute_cell": (("range(frame_runs)", "range(size_input.lengths[repo])"),),
}


class SizeSite(ast.NodeTransformer):
    def __init__(self, old, new):
        self.old = ast.dump(ast.parse(old, mode="eval").body)
        self.new = ast.parse(new, mode="eval").body
        self.count = 0

    def visit(self, node):
        if ast.dump(node) == self.old:
            self.count += 1
            return ast.copy_location(self.new, node)
        return super().visit(node)


def locked_functions(inputs):
    namespace = dict(vars(benchmark))
    namespace.update(size_input=inputs, locked_prepare=prepare_population, repositories=REPOSITORIES)
    for name, module in (("benchmark_population", benchmark), ("trial_frame", benchmark),
                         ("execute_cell", benchmark), ("calibration_measurement", reporter)):
        tree = ast.parse(inspect.getsource(getattr(module, name)))
        for old, new in SIZE_SITES[name]:
            change = SizeSite(old, new)
            tree = change.visit(tree)
            if change.count != 1:
                raise RuntimeError(f"signed size site drift: {name}: {old}: {change.count}")
        ast.fix_missing_locations(tree)
        exec(compile(tree, str(HERE / (name + "[size-source]")), "exec"), namespace)
    return namespace


def provenance(inputs):
    paths = [HERE/name for name in ("size_input.py", "population.py", "locked_commands.py")]
    paths += [VALIDATION/name for name in ("compute_benchmark.py", "reproduce_reports.py", "ranking.py",
                                         "fixed_population.py", "interval_diagnostic.go")]
    paths += [CONTROLLER_PATH, VALIDATION.parent / "oc-manifest.json"]
    return {"size_input_sha256": inputs.sha256, "size_input_record": json.loads(inputs.record_bytes),
            "L_r": inputs.lengths, "units": inputs.units,
            "source_sha256": {str(p.relative_to(ROOT)): hashlib.sha256(p.read_bytes()).hexdigest() for p in paths},
            "size_site_substitutions": SIZE_SITES}


def population_task(task):
    slot, record_bytes, kind = task
    inputs = SizeInput(json.loads(record_bytes))
    result = locked_functions(inputs)[kind](slot)
    result.update(L_r=inputs.lengths, size_input_sha256=inputs.sha256)
    if kind == "benchmark_population":
        result["size_only_sampling_gate"].update(
            scope="CPU-only non-binding gate timing over complete locked vectors; not the HMAC p_gate estimate",
            binding_p_gate=False)
    return result


def trial_task(task):
    record_bytes, signed_task = task
    inputs = SizeInput(json.loads(record_bytes))
    result = locked_functions(inputs)["execute_cell"](signed_task)
    result["cell"]["L_r"] = inputs.lengths
    result["size_input_sha256"] = inputs.sha256
    return result


def execute_trials(inputs, tasks, workers):
    started = time.perf_counter()
    with ProcessPoolExecutor(max_workers=workers) as pool:
        cells = list(pool.map(trial_task, ((inputs.record_bytes, task) for task in tasks)))
    return {"workers": workers, "elapsed_seconds": time.perf_counter()-started, "cells": cells}


def measurements(inputs, kind):
    # Both signed slots are identity/seed labels only; fresh process for each.
    with ProcessPoolExecutor(max_workers=1, max_tasks_per_child=1) as pool:
        return list(pool.map(population_task, ((slot, inputs.record_bytes, kind) for slot in (40, 400))))


def host():
    def query(*args):
        return subprocess.check_output(args, text=True).strip()
    return {"hw.model": query("sysctl", "-n", "hw.model"),
            "hw.memsize": int(query("sysctl", "-n", "hw.memsize")),
            "logical_cpu_count": os.cpu_count(), "os_build": query("sw_vers", "-buildVersion"),
            "platform": platform.platform(), "python_version": platform.python_version(),
            "python_executable_sha256": hashlib.sha256(Path(sys.executable).resolve().read_bytes()).hexdigest()}


def benchmark_mode(inputs, method, trials):
    populations = measurements(inputs, "benchmark_population")
    with tempfile.TemporaryDirectory(prefix="x1-locked-size-") as temporary:
        binary = str(Path(temporary)/"interval-kernel")
        start = time.perf_counter()
        subprocess.run(["go", "build", "-o", binary, str(VALIDATION/"interval_diagnostic.go")], check=True)
        build = {"elapsed_seconds": time.perf_counter()-start,
                 "binary_sha256": hashlib.sha256(Path(binary).read_bytes()).hexdigest(),
                 "go_version": subprocess.check_output(["go", "version"], text=True).strip()}
        tasks = [(slot, missing, method, trials, binary, fixture)
                 for slot, missing, fixture in itertools.product((40, 400), (False, True),
                                                                 ("uniform-srs", "mixed-explicit"))]
        executions = [execute_trials(inputs, tasks, workers) for workers in (1, 2)]
    science = [canonical(benchmark.science(run)) for run in executions]
    if science[0] != science[1]:
        raise AssertionError("scientific output differs by worker count")
    return {"method": method, "trials": trials, "replicates": benchmark.REPLICATES,
            "population_measurements": populations, "executions": executions, "compiled_kernel": build,
            "worker_count_equality": {"workers": [1, 2], "byte_identical_scientific_json": True,
                                      "scientific_sha256": hashlib.sha256(science[0]).hexdigest()},
            "frame_design": {"identity_slots": [40, 400], "L_r": inputs.lengths,
                             "selected_units_every_trial": 60000, "selected_runs_per_repository": 20,
                             "uniform_run_sizes": benchmark.SIZES,
                             "mixed_explicit_selected_run_sizes": benchmark.MIXED_SIZES}}


def calibration_mode(inputs):
    rows = measurements(inputs, "calibration_measurement")
    with ProcessPoolExecutor(max_workers=1, max_tasks_per_child=1) as pool:
        enumeration = pool.submit(reporter.worst_case_enumeration).result()
    return {"measurements": rows, "worst_case_candidate_enumeration": enumeration}


def locked_fit(inputs, data, calibration, gate):
    """P181 fit: full locked cost maxima; no planning slope/intercept."""
    for result in (data, calibration, gate):
        if canonical(result["provenance"]) != canonical(provenance(inputs)):
            raise ValueError("source/input closure mismatch")
    if data["host"] != calibration["host"] or data["host"] != gate["host"]:
        raise ValueError("locked measurements require the same host/interpreter")
    if data["trials"] < 2000 or data["replicates"] != 1999 or not data["worker_count_equality"]["byte_identical_scientific_json"]:
        raise ValueError("incomplete benchmark replicate structure")
    if gate["draws"] != 100000:
        raise ValueError("100000 keyed checks required")
    successes = gate["successes"]
    if type(successes) is not int or not 0 <= successes <= 100000:
        raise ValueError("invalid gate success count")
    cap = (4000*100000+successes-1)//successes if successes else None
    if gate["attempt_cap_per_population"] != cap:
        raise ValueError("attempt cap mismatch")
    prep = max(r["construction"]["cpu_seconds"] for r in data["population_measurements"])
    reevaluate = max([e["cpu_seconds"] for r in data["population_measurements"] for e in r["fixed_key_evaluations"]]
                     + [max(r["frame_evaluation_cpu_seconds"]) for r in calibration["measurements"]])
    cal = max(r["conservative_21_truth_one_frame_no_enumeration_cpu_seconds"] for r in calibration["measurements"])
    enumeration = calibration["worst_case_candidate_enumeration"]["cpu_seconds"]
    cheap_gate = max(r["size_only_sampling_gate"]["cpu_seconds"]/100000 for r in data["population_measurements"])
    worst = {missing: max(t["cpu_seconds"] for run in data["executions"] for cell in run["cells"]
                         if cell["cell"]["missing_labels"] == missing for t in cell["trial_timing"])
             for missing in (False, True)}
    rss = {"construction": max(r["construction"]["peak_rss_bytes"] for r in data["population_measurements"]),
           "calibration": max(r["peak_rss_bytes"] for r in calibration["measurements"]),
           "trial_worker_plus_helper": max(c["peak_rss_bytes"] for r in data["executions"] for c in r["cells"])}
    if any(not math.isfinite(v) or v <= 0 for v in (prep, reevaluate, cal, enumeration, cheap_gate, *worst.values(), *rss.values())):
        raise ValueError("unmeasured CPU/RSS")
    binding = {"construction": (17920+252)*5*prep, "requested_base_evaluation": 16128*5*reevaluate,
               "calibration_no_enumeration": (1792+252)*5*cal,
               "enumeration": (1792+252)*5*enumeration,
               "trials": 5*4000*(17920*worst[False]+252*worst[True]),
               "gate_checks": (17920+252)*5*cap*cheap_gate if cap else None}
    safety = {"construction": 336*5*prep, "canonical_state_evaluation": 8064*5*reevaluate,
              "calibration_no_enumeration": 336*5*cal, "enumeration": 336*5*enumeration,
              "trials": 5*2000*(336*worst[False]+1680*worst[True]),
              "gate_checks": (336+1680)*5*cap*cheap_gate if cap else None}
    binding_hours = 2*sum(binding.values())/3600 if cap else None
    total_hours = binding_hours+2*sum(safety.values())/3600 if cap else None
    memory_fit = 8*max(rss.values()) <= data["host"]["hw.memsize"]-8*1024**3
    fits = binding_hours is not None and binding_hours <= 2880 and binding_hours/192 <= 15 and memory_fit
    return {"status": "synthetic fit" if fits else "not testable", "fits_caps": fits,
            "binding_oc_gate": False, "attempt_cap_per_population": cap,
            "hard_trial_execution_cap_per_population": 4000, "overhead_multiplier": 2,
            "binding_cpu_seconds_before_overhead": binding, "descriptive_safety_cpu_seconds_before_overhead": safety,
            "binding_core_hours": binding_hours, "binding_wall_days_at_8_workers": binding_hours/192 if cap else None,
            "total_planning_core_hours": total_hours, "memory": {"locked_phase_maxima": rss,
            "eight_worker_bytes": 8*max(rss.values()), "fixed_reserve_bytes": 8*1024**3, "fits": memory_fit},
            "resource_method": "maxima at complete locked sizes; no L40/L400 planning slope; safety never causes binding stop"}


def publish(path, value):
    with path.open("x") as stream:
        json.dump(value, stream, indent=2, sort_keys=True, allow_nan=False)
        stream.write("\n")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("mode", choices=("synthetic-input", "gate", "benchmark", "calibration", "fit"))
    parser.add_argument("--input", type=Path)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--method", choices=("studentized", "percentile"), default="studentized")
    parser.add_argument("--trials", type=int, default=2000)
    parser.add_argument("--benchmark", type=Path)
    parser.add_argument("--calibration", type=Path)
    parser.add_argument("--gate", type=Path)
    args = parser.parse_args()
    if args.output.exists():
        raise FileExistsError("locked evidence is create-once")
    if args.mode == "synthetic-input":
        publish(args.output, synthetic_record({r: [50, 100, 500, 1000, 2000]*(5+i) for i, r in enumerate(REPOSITORIES)}))
        return
    if not args.input:
        parser.error("--input required")
    if platform.python_version() != "3.11.11" or sys.flags.optimize:
        raise RuntimeError("pinned unoptimized CPython 3.11.11 required")
    if args.trials < 2000:
        parser.error("trials>=2000 required; no reduced evidence mode")
    inputs = SizeInput(json.loads(args.input.read_text()))
    closure, machine = provenance(inputs), host()
    started = time.perf_counter()
    if args.mode == "gate":
        result = inputs.estimate_p_gate()
    elif args.mode == "benchmark":
        result = benchmark_mode(inputs, args.method, args.trials)
    elif args.mode == "calibration":
        result = calibration_mode(inputs)
    else:
        if not all((args.benchmark, args.calibration, args.gate)):
            parser.error("fit requires --benchmark, --calibration and --gate")
        result = locked_fit(inputs, *(json.loads(p.read_text()) for p in (args.benchmark, args.calibration, args.gate)))
        result["evidence_sha256"] = {str(p): hashlib.sha256(p.read_bytes()).hexdigest()
                                      for p in (args.benchmark, args.calibration, args.gate)}
    if closure != provenance(inputs) or inputs.record_bytes != canonical(json.loads(args.input.read_text())):
        raise RuntimeError("source or size input changed; output not published")
    result.update(schema="x1-locked-size-"+args.mode+"/synthetic-v1", synthetic_only=True,
                  authority="Draft synthetic evidence only; no preparation, inventory, remeasurement or OC authority",
                  provenance=closure, host=machine, elapsed_seconds=time.perf_counter()-started)
    publish(args.output, result)


if __name__ == "__main__":
    main()

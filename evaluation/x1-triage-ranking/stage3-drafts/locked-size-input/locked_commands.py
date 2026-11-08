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

HERE = Path(__file__).resolve().parent
VALIDATION = HERE.parents[1] / "validation"
ROOT = HERE.parents[3]
SIGNED = {
    "compute_benchmark.py": "d39f40a28a6fd70698ea9e53e30448dec193c395d264002436dad9278ea6c90d",
    "reproduce_reports.py": "023e2e5bc6ca1ee282e4ebc75ef6b3bccdb96111d0f218c0341189d6a400168e",
    "fixed_population.py": "ed670a6bf8076d791c06cfc2426aff03a0a2f9b1730982cf3fdfb2ec13bc49a2",
    "ranking.py": "8bc2e7cbbc5da6ba47aeff3effb5dd04d5e5ecc9c1e01051647900095770a780",
    "interval_diagnostic.go": "437f9c2559156f8befb613d3e17f19e2aad5437217081dbd4a1d0fb95c069417",
    "../oc-manifest.json": "1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d",
}


def pin_sources():
    for name, expected in SIGNED.items():
        if hashlib.sha256((VALIDATION/name).read_bytes()).hexdigest() != expected:
            raise ValueError(f"substituted signed source: {name}")


pin_sources()
from population import prepare_population


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
    pin_sources()
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
            "logical_cpus": os.cpu_count(), "os_build": query("sw_vers", "-buildVersion")}


def host_load():
    return {"load_averages": list(os.getloadavg()),
            "concurrent_process_count": len(subprocess.check_output(
                ["ps", "-A", "-o", "pid="], text=True).splitlines())}


def interpreter():
    return {"python_executable": sys.executable, "python_version": platform.python_version(),
            "python_optimize": sys.flags.optimize,
            "python_executable_sha256": hashlib.sha256(Path(sys.executable).resolve().read_bytes()).hexdigest()}


def benchmark_mode(inputs, trials):
    method = "studentized"
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


def measurement_run(inputs, data, calibration, gate, evidence_sha256):
    """Extract binding operation maxima; only the final evaluator decides fit."""
    for result in (data, calibration, gate):
        if result["provenance"]["size_input_sha256"] != inputs.sha256:
            raise ValueError("size input mismatch")
        if result["provenance"] != json.loads(canonical(provenance(inputs))) and result["integrity_failure"] is None:
            raise ValueError("source/input closure mismatch")
    if data["host"] != calibration["host"] or data["host"] != gate["host"]:
        raise ValueError("locked measurements require the same host")
    if any(data[k] != r[k] for r in (calibration, gate)
           for k in interpreter()):
        raise ValueError("locked measurements require the same interpreter")
    validate_benchmark(data)
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
    values = {"construction_cpu_per_unit": prep/inputs.units,
              "requested_base_cpu_per_unit": reevaluate/inputs.units,
              "calibration_cpu_per_unit": cal/inputs.units,
              "enumeration_cpu_per_population": enumeration, "gate_cpu_per_check": cheap_gate,
              "complete_trial_cpu": worst[False], "full512_trial_cpu": worst[True],
              "construction_rss": rss["construction"], "calibration_rss": rss["calibration"],
              "trial_plus_helper_rss": rss["trial_worker_plus_helper"]}
    failed = any(r["integrity_failure"] for r in (data, calibration, gate))
    record = {"id": hashlib.sha256(canonical(evidence_sha256)).hexdigest(),
              "host": data["host"], **data["host_load"]["start"],
              "host_load": {name: r["host_load"] for name, r in
                            (("benchmark", data), ("calibration", calibration), ("gate", gate))},
              **{k: data[k] for k in interpreter()},
              "size_input_sha256": inputs.sha256, "binding_inputs": values,
              "completed": True, "integrity_failure": "closure mismatch" if failed else None}
    if failed:
        record["preserved_evidence_sha256"] = hashlib.sha256(canonical(evidence_sha256)).hexdigest()
    return record


def validate_benchmark(data):
    expected = {(slot, missing, "studentized", fixture) for slot, missing, fixture in
                itertools.product((40, 400), (False, True), ("uniform-srs", "mixed-explicit"))}
    if (data["method"] != "studentized" or type(data["trials"]) is not int or data["trials"] < 2000
        or data["replicates"] != 1999
        or not data["worker_count_equality"]["byte_identical_scientific_json"]
        or [r["workers"] for r in data["executions"]] != [1, 2]):
        raise ValueError("incomplete studentized benchmark")
    for run in data["executions"]:
        actual = [(c["cell"]["L"], c["cell"]["missing_labels"], c["cell"]["method"],
                   c["cell"]["fixture"]) for c in run["cells"]]
        if len(actual) != 8 or set(actual) != expected or any(
                c["measured_trials"] != data["trials"] for c in run["cells"]):
            raise ValueError("frozen benchmark cell set required")


def locked_inventory(inputs, gate):
    return {"size_input_sha256": inputs.sha256, "L_r": inputs.lengths,
            "eligible_sizes": {r: list(v) for r, v in inputs.vectors.items()},
            "gate_redraws": gate["draws"], "gate_successes": gate["successes"],
            "p_gate": gate["p_gate"]}


def read_evidence(paths):
    records, digests = [], {}
    for path in paths:
        data = path.read_bytes()
        records.append(json.loads(data))
        digests[str(path)] = hashlib.sha256(data).hexdigest()
    return records, digests


def publish(path, value):
    with path.open("x") as stream:
        json.dump(value, stream, indent=2, sort_keys=True, allow_nan=False)
        stream.write("\n")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("mode", choices=("synthetic-input", "gate", "benchmark", "calibration", "fit"))
    parser.add_argument("--input", type=Path)
    parser.add_argument("--output", type=Path, required=True)
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
    pin_sources()
    closure, machine, load_start = provenance(inputs), host(), host_load()
    runtime = interpreter()
    started = time.perf_counter()
    if args.mode == "gate":
        result = inputs.estimate_p_gate()
        result["locked_inventory"] = locked_inventory(inputs, result)
    elif args.mode == "benchmark":
        result = benchmark_mode(inputs, args.trials)
    elif args.mode == "calibration":
        result = calibration_mode(inputs)
    else:
        if not all((args.benchmark, args.calibration, args.gate)):
            parser.error("fit requires --benchmark, --calibration and --gate")
        records, digests = read_evidence((args.benchmark, args.calibration, args.gate))
        inventory = locked_inventory(inputs, records[2])
        result = {"measurement_runs": {
                      "locked_inventory_sha256": hashlib.sha256(canonical(inventory)).hexdigest(),
                      "runs": [measurement_run(inputs, *records, digests)]},
                  "locked_inventory": inventory, "evidence_sha256": digests}
    finish(inputs, args.input, args.output, args.mode, result, closure, machine,
           load_start, runtime, started)


def finish(inputs, input_path, output, mode, result, closure, machine, load_start, runtime, started):
    end_closure = provenance(inputs)
    input_end = input_path.read_bytes()
    try:
        unchanged_input = inputs.record_bytes == canonical(json.loads(input_end))
    except (ValueError, UnicodeError):
        unchanged_input = False
    failed = closure != end_closure or not unchanged_input
    if failed and "measurement_runs" in result:
        for record in result["measurement_runs"]["runs"]:
            record["integrity_failure"] = "closure mismatch"
            record["preserved_evidence_sha256"] = hashlib.sha256(canonical(result["evidence_sha256"])).hexdigest()
    result.update(schema="x1-locked-size-"+mode+"/synthetic-v1", synthetic_only=True,
                  authority="Draft synthetic evidence only; no preparation, inventory, remeasurement or OC authority",
                  provenance=closure, end_provenance=end_closure, input_end_sha256=hashlib.sha256(input_end).hexdigest(),
                  host=machine, host_load={"start": load_start, "end": host_load()}, **runtime,
                  completed=True, integrity_failure="closure mismatch" if failed else None,
                  elapsed_seconds=time.perf_counter()-started)
    publish(output, result)
    if failed:
        raise RuntimeError("closure mismatch; completed evidence published")


if __name__ == "__main__":
    main()

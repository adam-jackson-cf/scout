"""Synthetic-only deterministic journal I/O benchmark; no scans/models/randomness.

Counts here are instrumentation samples, never substitutes for scientific trials.
Print full JSON to stdout; preserve it as the round's performance evidence.
"""
import argparse
from fractions import Fraction
import json
from pathlib import Path
import tempfile
import time
from types import SimpleNamespace
from unittest.mock import patch

import controller as c


def measure(mode, count, probability):
    rate = Fraction(str(probability))
    with tempfile.TemporaryDirectory(prefix="x1-controller-synthetic-io-") as directory:
        # This temporary synthetic sample does not reserve full-OC storage.
        # Real admission is unchanged and its reservation is reported below.
        with patch.object(c.shutil, "disk_usage", return_value=SimpleNamespace(free=10**16)):
            s = c.SafetySchedule(p_gate=probability, inventory_sha256="c"*64,
                meter=lambda: 0., oc_started_at=100000., clock=lambda: 100000., run_root=directory)
        journal = Path(directory) / "x1-oc-controller/journal.jsonl"
        lane = s.binding_population(c.canonical({"synthetic_benchmark": mode}).decode(), 0, slot=0)
        initialization_bytes = journal.stat().st_size
        synchronization = c.os.fsync
        began = time.perf_counter_ns()
        with patch.object(c.os, "fsync", wraps=synchronization) as sync:
            for i in range(count):
                passes = mode == "passing" or (mode == "mixed" and (i * rate.numerator) % rate.denominator < rate.numerator)
                lane.step(lambda index, result=passes: result, lambda index: True)
            elapsed_us = (time.perf_counter_ns() - began) / 1000
            attempt_syncs = sync.call_count
            # Force pending gate deltas durable, separately costed from attempts.
            began_flush = time.perf_counter_ns()
            s.pause({"kind": "execution integrity", "evidence_sha256": "b"*64})
            final_flush_us = (time.perf_counter_ns() - began_flush) / 1000
            final_flush_syncs = sync.call_count - attempt_syncs
        all_rows = [json.loads(line) for line in journal.read_bytes().splitlines()]
        rows = all_rows[1:-1]  # population creation / final pause are separate
        attempt_bytes = sum(len(c.canonical(row))+1 for row in rows)
        counters = lane.report()
        result = {"mode": mode, "synthetic_sample_attempts": count,
            "executions": counters["executions"], "gate_rejections": counters["gate_failures"],
            "records": len(rows), "fsyncs": attempt_syncs, "bytes": attempt_bytes,
            "microseconds": elapsed_us, "records_per_attempt": len(rows)/count,
            "bytes_per_attempt": attempt_bytes/count, "microseconds_per_attempt": elapsed_us/count,
            "initialization_bytes": initialization_bytes,
            "final_gate_flush": {"records": 1, "fsyncs": final_flush_syncs,
                "bytes": len(c.canonical(all_rows[-1]))+1, "microseconds": final_flush_us},
            "maximum_observed_record_bytes": max(len(c.canonical(row))+1 for row in all_rows),
            "freeze": json.loads(c.canonical(s._freeze))}
        s.close()
        return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--attempts", type=int, default=1000)
    parser.add_argument("--p-gate", type=float, default=.5)
    args = parser.parse_args()
    if not 0 < args.attempts <= 1999 or not 0 < args.p_gate <= 1:
        parser.error("synthetic sample requires 1..1999 attempts and 0<p_gate<=1")
    samples = [measure(mode, args.attempts, args.p_gate) for mode in ("passing", "rejected", "mixed")]
    mixed = samples[-1]
    populations = 90860 + 10080
    attempts = populations * 2000 / args.p_gate
    print(json.dumps({"synthetic_only": True, "authority": "none", "controller_randomness": False,
        "sample_note": "Instrumentation only: no reduced scientific trial count or resource-fit claim; real work callbacks are not timed here. CPU/total-OC fit remains separate.",
        "p_gate": args.p_gate, "samples": samples,
        "full_population_projection": {"binding_populations": 90860, "safety_populations": 10080,
            "minimum_executions": populations*2000, "expected_attempts_at_frozen_p_gate": attempts,
            "sample_scaled_bytes": attempts*mixed["bytes_per_attempt"] + populations*(mixed["initialization_bytes"]+mixed["final_gate_flush"]["bytes"]),
            "sample_scaled_serial_microseconds": attempts*mixed["microseconds_per_attempt"],
            "qualification": "Linear synthetic extrapolation, not a scientific run or an upper resource-fit bound; binding inadequacy may require up to 4000 executions per population. Control-phase payload and platform I/O costs differ.",
            "conservative_freeze_reservation": c.journal_projection(args.p_gate)},
        "source_sha256": {name: c.digest((c.HERE/name).read_bytes()) for name in ("controller.py", "policy.json", "benchmark_journal.py")}}, indent=2, sort_keys=True))


if __name__ == "__main__":
    main()

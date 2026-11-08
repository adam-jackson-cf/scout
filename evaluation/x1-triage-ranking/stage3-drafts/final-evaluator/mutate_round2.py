"""Reproducible synthetic round-2 guard probes; mutate only temporary copies."""
from pathlib import Path
import hashlib
import os
import subprocess
import sys
import tempfile

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]

# Each replacement must match exactly once; changed source requires explicit review.
MUTANTS = (
    ("R2-01-resource-row-gate", "evaluator.py", 'passed = {m: resources["passed"]',
     'passed = {m: True', "test_resource_fit_propagates_to_every_row"),
    ("R2-02-signed-manifest-pin", "provenance.py",
     '    if digest(payloads["signed_manifest"]) != MANIFEST_SHA256:\n        raise ValueError("substituted signed manifest bytes")\n',
     '', "test_signed_manifest_pin_with_consistent_bundle_resealing"),
    ("R2-03-missing-finite-gain", "evaluator.py", 'type(realised) in (int, float) and math.isfinite(realised)\n',
     'type(realised) in (int, float)\n', "test_nonfinite_exponent_gain_bytes_fail_cells"),
    ("R1-05-reference-finite-gain", "evaluator.py",
     'if not number(realised) and not (type(realised) in (int, float) and math.isfinite(realised) and realised < 0):',
     'if type(realised) not in (int, float):', "test_nonfinite_exponent_gain_bytes_fail_cells"),
    ("R2-04-host-identity", "provenance.py",
     '    if (set(host) != {"hw.model", "hw.memsize", "logical_cpus", "os_build"}\n'
     '        or any(type(host[k]) is not str or not host[k] for k in ("hw.model", "os_build"))\n'
     '        or any(type(host[k]) is not int or host[k] <= 0 for k in ("hw.memsize", "logical_cpus"))):\n'
     '        raise ValueError("missing host identity")',
     '    pass', "test_complete_host_and_load_provenance"),
    ("R2-04-load-provenance", "provenance.py",
     '        if (run["host"] != binding_host or type(run.get("load_averages")) is not list\n'
     '            or len(run["load_averages"]) != 3 or not all(number(v) for v in run["load_averages"])\n'
     '            or type(run.get("concurrent_process_count")) is not int or run["concurrent_process_count"] < 1):',
     '        if run["host"] != binding_host:', "test_complete_host_and_load_provenance"),
    ("R2-05-point-cutoff", "evaluator.py", 'point >= .10 and lower > 0',
     'lower > 0', "test_ordered_typed_verdict_and_interval_boundaries"),
    ("R2-05-reversed-interval", "evaluator.py", ' or lower > upper',
     '', "test_ordered_typed_verdict_and_interval_boundaries"),
    ("R2-06-bootstrap-truthiness", "evaluator.py", 'not complete_bootstrap_evidence(evidence.get("bootstrap_evidence"))',
     'not evidence.get("bootstrap_evidence")', "test_complete_typed_bootstrap_evidence"),
    ("R2-06-bootstrap-inverse", "evaluator.py", 'not complete_bootstrap_evidence(evidence.get("bootstrap_evidence"))',
     'complete_bootstrap_evidence(evidence.get("bootstrap_evidence"))', "test_complete_typed_bootstrap_evidence"),
)


def main():
    for filename in ("evaluator.py", "provenance.py", "test_evaluator.py", "mutate_round2.py"):
        print(f"SHA256 {filename} {hashlib.sha256((HERE/filename).read_bytes()).hexdigest()}", flush=True)
    failures = []
    for name, filename, before, after, test in MUTANTS:
        with tempfile.TemporaryDirectory(prefix="x1g-evaluator-r2-") as temporary:
            root = Path(temporary)/"x1"
            target = root/"stage3-drafts"/"final-evaluator"
            target.mkdir(parents=True)
            (root/"validation").mkdir()
            for source in ("oc-manifest.json", "validation/reproduce_reports.py"):
                (root/source).write_bytes((ROOT/source).read_bytes())
            for source in ("evaluator.py", "provenance.py", "test_evaluator.py"):
                (target/source).write_bytes((HERE/source).read_bytes())
            path = target/filename
            original = path.read_text()
            if original.count(before) != 1:
                raise ValueError(f"{name}: mutation anchor must occur exactly once")
            path.write_text(original.replace(before, after, 1))
            command = [sys.executable, str(target/"test_evaluator.py"), "-v", f"EvaluatorTests.{test}"]
            print(f"\nMUTANT {name}\nCOMMAND {command!r}", flush=True)
            result = subprocess.run(command, cwd=target, env={**os.environ, "PYTHONDONTWRITEBYTECODE": "1"},
                                    text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
            print(result.stdout, end="", flush=True)
            killed = result.returncode != 0 and "FAIL:" in result.stdout
            print(f"RESULT {name} {'KILLED' if killed else 'SURVIVED/INVALID'}", flush=True)
            if not killed:
                failures.append(name)
    print(f"\nSUMMARY {len(MUTANTS)-len(failures)}/{len(MUTANTS)} mutants killed", flush=True)
    return bool(failures)


if __name__ == "__main__":
    sys.exit(main())

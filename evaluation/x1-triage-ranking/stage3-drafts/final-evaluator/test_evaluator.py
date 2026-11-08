"""Deterministic planted effect/null and adversarial evaluator fixtures."""
import hashlib
import json
from pathlib import Path
import unittest
from evaluator import expected_cells, evaluate, resource_fit, verdict, wilson
from provenance import maximum_completed_inputs


class EvaluatorTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.manifest = json.loads((Path(__file__).resolve().parents[2]/"oc-manifest.json").read_text())
        cls.records = []
        for cell in expected_cells(cls.manifest):
            for p in range(5):
                cls.records.append(dict(cell=cell, population=p, B=1999, attempts=2000,
                    executions=2000, eligible=2000, inadequate=0, gate_failures=0,
                    lower_miss=20, upper_miss=20, support=0 if cell.get("target") == 0 else 1900,
                    realised_gain=cell.get("target", cell.get("realised_gain", .20)), reachable=True, mask_reachable=True))
        cls.resources = dict(cpu_seconds={k: 100 for k in ("construction", "calibration", "enumeration", "requested_base", "trials", "gate_checks")},
            peak_rss_bytes={k: 1000000 for k in ("construction", "calibration", "trial_plus_helper")},
            wall_days=1, memory_bytes=68719476736, workers=8)
        payloads = {k: ("synthetic:"+k).encode() for k in ("locked_inventory", "signed_manifest", "source_runtime_closure", "controller_freeze", "scoped_council_record")}
        host = {"hw.model": "synthetic-host", "hw.memsize": 68719476736, "logical_cpus": 18, "os_build": "synthetic-os"}
        flat = {**{"cpu:"+k: v for k, v in cls.resources["cpu_seconds"].items()},
                **{"rss:"+k: v for k, v in cls.resources["peak_rss_bytes"].items()}, "wall_days": 1}
        cls.provenance = dict(provenance_payloads=payloads,
            provenance_digests={k: hashlib.sha256(v).hexdigest() for k, v in payloads.items()},
            binding_host=host, measurement_runs=[dict(id="synthetic-run-1", host=host,
                load_averages=[0, 0, 0], concurrent_process_count=1, integrity_failure=None,
                completed=True, binding_inputs=flat)])

    def run_gate(self, records=None, **kwargs):
        return evaluate(self.manifest, self.records if records is None else records,
            p_gate=.5, resources=self.resources, mhat=848, synthetic_only=True, **self.provenance, **kwargs)

    def test_exact_full_grid_and_planted_reference_effect_null(self):
        self.assertEqual(len(self.records), 90860)
        result = self.run_gate()
        self.assertEqual(result["selected_M"], 700)
        self.assertEqual(result["status"], "draft_synthetic_gate_passed")
        self.assertTrue(all(result["rows"].values()))
        self.assertGreater(wilson(9500, 10000)[0], .80)
        self.assertLess(wilson(0, 10000)[1], .05)

    def test_missing_duplicate_wrong_B_and_counter_fail_closed(self):
        for rows in (self.records[:-1], self.records+[self.records[0]]):
            with self.assertRaises(ValueError):
                self.run_gate(rows)
        for changes in ({"B": 1998}, {"executions": 4001}, {"attempts": 8001}, {"eligible": 1999}):
            rows = list(self.records)
            index = next(i for i, r in enumerate(rows) if r["cell"]["cap_M"] == 700)
            rows[index] = {**rows[index], **changes}
            result = self.run_gate(rows)
            self.assertEqual(result["status"], "not testable")
            self.assertEqual(result["selected_M"], 700)  # No lower-row bypass.

    def test_null_false_support_and_unreachable_power_fail(self):
        for target, changes in ((0, {"support": 2000}), (.2, {"reachable": False}), (.2, {"realised_gain": .19})):
            rows = list(self.records)
            for i, r in enumerate(rows):
                if r["cell"].get("target") == target and r["cell"]["cap_M"] == 700:
                    rows[i] = {**r, **changes}
                    break
            self.assertEqual(self.run_gate(rows)["status"], "not testable")

    def test_resource_limits_and_no_smaller_row(self):
        for changes in ({"workers": 7}, {"wall_days": 16}, {"memory_bytes": 8*1024**3}):
            self.assertFalse(resource_fit({**self.resources, **changes}))
        result = evaluate(self.manifest, self.records, p_gate=.5, resources=self.resources, mhat=399, synthetic_only=True, **self.provenance)
        self.assertEqual(result["status"], "not testable")
        with self.assertRaises(ValueError):
            evaluate(self.manifest, self.records, p_gate=.5, resources=self.resources, mhat=700, synthetic_only=False, **self.provenance)

    def test_substituted_provenance_and_completed_closure_failure_maximum(self):
        bad = dict(self.provenance["provenance_payloads"])
        bad["locked_inventory"] = b"substituted"
        with self.assertRaises(ValueError):
            evaluate(self.manifest, self.records, p_gate=.5, resources=self.resources,
                mhat=700, synthetic_only=True, **{**self.provenance, "provenance_payloads": bad})
        run = self.provenance["measurement_runs"][0]
        failed = {**run, "integrity_failure": "closure mismatch", "preserved_evidence_sha256": "a"*64,
                  "binding_inputs": {k: v*2 for k, v in run["binding_inputs"].items()}}
        rerun = {**run, "id": "synthetic-run-2"}
        maxima = maximum_completed_inputs([failed, rerun], self.provenance["binding_host"])
        self.assertEqual(maxima, failed["binding_inputs"])
        for cause in ("operator termination", "host load", "resource exhaustion"):
            with self.assertRaises(ValueError):
                maximum_completed_inputs([{**failed, "integrity_failure": cause}, rerun], self.provenance["binding_host"])

    def test_ordered_verdict_planted_effect_null_and_boundaries(self):
        effect = dict(integrity=True, violating=False, adequate=True, upper=.4, unresolved=False, point=.3, lower=.2)
        self.assertEqual(verdict(**effect), "support")
        for changes, want in ((dict(point=0, lower=-.05, upper=.05), "not supported"),
            (dict(integrity=False, violating=True), "inconclusive"),
            (dict(violating=True, adequate=False), "not supported"),
            (dict(adequate=False, upper=.05, lower=-.05), "inconclusive"),
            (dict(unresolved=True, upper=.05, lower=-.05), "not supported"),
            (dict(unresolved=True), "inconclusive"),
            (dict(point=.10, lower=0), "inconclusive"),
            (dict(point=.10, lower=.001), "support")):
            self.assertEqual(verdict(**{**effect, **changes}), want)


if __name__ == "__main__":
    unittest.main()

"""Synthetic-only behavioral boundaries. Run only after integration, not during edits."""
from collections import Counter
from pathlib import Path
import math
import unittest
from evaluator import expected_cells, identity, evaluate, resource_fit, select_postdev, verdict, wilson
from provenance import INPUTS, MANIFEST_SHA256, encode, decode, digest, maximum_completed_inputs

ROOT = Path(__file__).resolve().parents[2]
IDENTITIES_SHA256 = "3499e423b90df60972b1003c4fc8b65f77a03eada73d67afafa1288e8a43dda2"
REV8_SOURCE_SHA256 = "023e2e5bc6ca1ee282e4ebc75ef6b3bccdb96111d0f218c0341189d6a400168e"


class EvaluatorTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.manifest_bytes = (ROOT / "oc-manifest.json").read_bytes()
        cls.manifest = decode(cls.manifest_bytes)
        cls.cells = list(expected_cells(cls.manifest))
        cls.records = [dict(cell=c, population=p, B=1999, attempts=2000, executions=2000,
            eligible=2000, inadequate=0, gate_failures=0, lower_miss=20, upper_miss=20,
            support=0 if c.get("target") == 0 else 1900,
            realised_gain=c.get("target", c.get("realised_gain", .2)), reachable=True,
            mask_reachable=True) for c in cls.cells for p in range(5)]
        cls.host = {"hw.model": "synthetic-host", "hw.memsize": 68719476736,
                    "logical_cpus": 18, "os_build": "synthetic-os"}
        cls.inputs = {k: 1000000 if k.endswith("rss") else .000001 for k in INPUTS}
        cls.inventory = dict(gate_redraws=100000, gate_successes=50000, p_gate=.5,
                             eligible_sizes={k: [20]*20 for k in ("synthetic/a", "synthetic/b", "synthetic/c")})
        cls.measurement_run = dict(id="synthetic-run-1", host=cls.host, load_averages=[0, 0, 0],
                       concurrent_process_count=1, integrity_failure=None, completed=True,
                       binding_inputs=cls.inputs)

    def bundle(self, *, records=None, inventory=None, runs=None, host=None):
        inv = encode(self.inventory if inventory is None else inventory)
        rows = self.records if records is None else records
        oc = dict(records=rows, records_sha256=digest(encode(rows)), host=self.host if host is None else host,
                  locked_inventory_sha256=digest(inv), signed_manifest_sha256=MANIFEST_SHA256)
        payloads = dict(locked_inventory=inv, signed_manifest=self.manifest_bytes,
            source_runtime_closure=b"synthetic source closure", controller_freeze=b"synthetic controller freeze",
            scoped_council_record=b"synthetic council record; no authority",
            measurement_runs=encode(dict(locked_inventory_sha256=digest(inv), runs=[self.measurement_run] if runs is None else runs)),
            sealed_oc_output=encode(oc))
        return payloads, {k: digest(v) for k, v in payloads.items()}

    def gate(self, **kwargs):
        payloads, pins = self.bundle(**kwargs)
        return evaluate(self.manifest_bytes, provenance_payloads=payloads,
                        provenance_digests=pins, synthetic_only=True)

    def select(self, predev, mhat=848, *, payloads=None, pins=None, evidence_changes=None):
        if payloads is None:
            payloads, pins = self.bundle()
        payloads, pins = dict(payloads), dict(pins)
        evidence = dict(Mhat=mhat, c_U=.50/(mhat+.5), bootstrap_evidence={"synthetic_bootstrap": [1, 1, 1]},
                        predev_sha256=predev["result_sha256"], locked_inventory_sha256=pins["locked_inventory"])
        evidence.update(evidence_changes or {})
        payloads["mhat_evidence"] = encode(evidence)
        pins["mhat_evidence"] = digest(payloads["mhat_evidence"])
        return select_postdev(predev["sealed_result"], predev["result_sha256"],
                             provenance_payloads=payloads, provenance_digests=pins, synthetic_only=True)

    def changed_cell(self, kind, changes, *, target=None, pooled=None):
        rows = list(self.records)
        cell = next(c for c in self.cells if c["kind"] == kind and c["cap_M"] == 700
                    and (target is None or c.get("target", c.get("realised_gain")) == target))
        indexes = [i for i, r in enumerate(rows) if r["cell"] == cell]
        for j, i in enumerate(indexes):
            patch = changes if pooled else (changes if j == 0 else {})
            rows[i] = {**rows[i], **patch}
        return rows, identity(cell), indexes

    def cell_passes(self, kind, changes, **kwargs):
        rows, key, _ = self.changed_cell(kind, changes, **kwargs)
        result = self.gate(records=rows)["result"]
        return result["cell_reports"][key]["passed"]

    def test_identity_counts_and_known_source_pins(self):
        self.assertEqual(Counter(c["kind"] for c in self.cells),
                         {"requested-coverage": 16128, "calibrated-reference": 1792, "missing": 252})
        self.assertEqual(len(self.records), 90860)
        self.assertEqual(digest(encode(sorted(identity(c) for c in self.cells))), IDENTITIES_SHA256)
        self.assertEqual(digest(self.manifest_bytes), MANIFEST_SHA256)
        self.assertEqual(digest((ROOT/"validation/reproduce_reports.py").read_bytes()), REV8_SOURCE_SHA256)

    def test_full_grid_predev_postdev_and_selection_boundaries(self):
        pre = self.gate()
        self.assertEqual(pre["result"]["status"], "draft_synthetic_rows_passed")
        self.assertTrue(all(pre["result"]["rows"].values()))
        for mhat, selected in ((0, None), (399, None), (400, 400), (699, 550), (700, 700), (848, 700), (2000, 2000)):
            result = self.select(pre, mhat)
            self.assertEqual(result["selected_M"], selected)
            self.assertEqual(result["status"], "not testable" if selected is None else "draft_synthetic_gate_passed")
        self.assertEqual(pre["result"]["measurement_runs"][0]["run"], self.measurement_run)
        self.assertEqual(pre["result"]["measurement_runs"][0]["classification"], "completed-valid-included")
        # Exact floor thresholds and the hard 2000 cap are independently bound.
        self.assertEqual(self.select(pre, 400, evidence_changes={"c_U": .00125})["selected_M"], 400)
        self.assertEqual(self.select(pre, 2000, evidence_changes={"c_U": 1e-300})["selected_M"], 2000)
        for patch in ({"Mhat": 2000}, {"Mhat": 848.0}, {"Mhat": -1}, {"Mhat": 2001},
                      {"c_U": 0}, {"c_U": .5}, {"c_U": math.nextafter(.50/848, math.inf)}):
            with self.assertRaises(ValueError):
                self.select(pre, evidence_changes=patch)

    def test_byte_bundle_substitutions_and_no_parallel_inputs(self):
        payloads, pins = self.bundle()
        for member in payloads:
            with self.subTest(member=member), self.assertRaises(ValueError):
                evaluate(self.manifest_bytes, provenance_payloads={**payloads, member: payloads[member]+b" "},
                         provenance_digests=pins, synthetic_only=True)
        for replacement in (encode(self.manifest), self.manifest_bytes+b" "):
            with self.assertRaises(ValueError):
                evaluate(replacement, provenance_payloads=payloads, provenance_digests=pins, synthetic_only=True)
        # There are deliberately no unbound parallel records, p_gate, run, resource or Mhat arguments.
        for name, value in (("records", []), ("p_gate", 1), ("measurement_runs", []),
                            ("resources", {}), ("mhat", 2000)):
            with self.subTest(name=name), self.assertRaises(TypeError):
                evaluate(self.manifest_bytes, provenance_payloads=payloads, provenance_digests=pins,
                         synthetic_only=True, **{name: value})
        with self.assertRaises(ValueError):
            evaluate(self.manifest_bytes, provenance_payloads=payloads, provenance_digests=pins, synthetic_only=False)
        for member in ("locked_inventory", "sealed_oc_output", "measurement_runs"):
            bad = dict(payloads)
            data = decode(bad[member])
            if member == "locked_inventory": data["p_gate"] = .8
            elif member == "sealed_oc_output": data["records"][0]["eligible"] = 1999
            else: data["runs"][0]["binding_inputs"]["complete_trial_cpu"] = 1
            bad[member] = encode(data)
            with self.assertRaises(ValueError):
                evaluate(self.manifest_bytes, provenance_payloads=bad, provenance_digests=pins, synthetic_only=True)

    def test_seal_internal_references_and_inventory_types(self):
        payloads, pins = self.bundle()
        for member, field in (("sealed_oc_output", "records_sha256"),
                              ("sealed_oc_output", "locked_inventory_sha256"),
                              ("sealed_oc_output", "signed_manifest_sha256"),
                              ("measurement_runs", "locked_inventory_sha256")):
            bad, hashes = dict(payloads), dict(pins)
            data = decode(bad[member]); data[field] = "0"*64
            bad[member] = encode(data); hashes[member] = digest(bad[member])
            with self.assertRaises(ValueError):
                evaluate(self.manifest_bytes, provenance_payloads=bad, provenance_digests=hashes, synthetic_only=True)
        for patch in ({"p_gate": True}, {"p_gate": .6}, {"gate_redraws": 99999},
                      {"gate_successes": True}, {"p_gate": 0}, {"eligible_sizes": {"one": [20]*20}}):
            with self.assertRaises(ValueError): self.gate(inventory={**self.inventory, **patch})
        for length in (19, 2001):
            sizes = {**self.inventory["eligible_sizes"], "synthetic/a": [20]*length}
            with self.assertRaises(ValueError):
                self.gate(inventory={**self.inventory, "eligible_sizes": sizes})
        upper_sizes = {**self.inventory["eligible_sizes"], "synthetic/a": [20]*2000}
        self.assertEqual(self.gate(inventory={**self.inventory, "eligible_sizes": upper_sizes})
                         ["result"]["resources"]["units_per_population"], 40800)
        for value in (float("nan"), float("inf")):
            with self.assertRaises(ValueError): self.bundle(inventory={**self.inventory, "p_gate": value})

    def test_exact_membership_B_and_integer_mutants(self):
        for rows in (self.records[:-1], self.records+[self.records[0]],
                     [{**self.records[0], "population": 5}]+self.records[1:],
                     [{**self.records[0], "cell": {**self.records[0]["cell"], "extra": 1}}]+self.records[1:]):
            with self.assertRaises(ValueError): self.gate(records=rows)
        self.assertFalse(self.cell_passes("requested-coverage", {"B": 1998}))
        with self.assertRaises(ValueError): self.cell_passes("requested-coverage", {"attempts": True})

    def test_reconciled_counter_caps_and_eligible_boundary(self):
        for patch, want in ((dict(executions=4000, inadequate=2000, attempts=4000), True),
                            (dict(attempts=8000, gate_failures=6000), True),
                            (dict(executions=4001, inadequate=2001, attempts=4001), False),
                            (dict(attempts=8001, gate_failures=6001), False),
                            (dict(eligible=1999, inadequate=1), False)):
            self.assertEqual(self.cell_passes("requested-coverage", patch), want)

    def test_each_tail_boundary_for_each_cell_kind(self):
        self.assertLessEqual(wilson(313, 10000)[1], .035)
        self.assertGreater(wilson(314, 10000)[1], .035)
        for kind in ("requested-coverage", "calibrated-reference", "missing"):
            for field in ("lower_miss", "upper_miss"):
                for misses, want in ((313, True), (314, False)):
                    rows, key, indexes = self.changed_cell(kind, {})
                    for j, i in enumerate(indexes):
                        rows[i] = {**rows[i], field: misses//5+(j < misses%5)}
                    self.assertEqual(self.gate(records=rows)["result"]["cell_reports"][key]["passed"], want)

    def test_reference_power_null_and_inadequacy_boundaries(self):
        self.assertGreaterEqual(wilson(8079, 10000)[0], .8)
        self.assertLess(wilson(8078, 10000)[0], .8)
        self.assertLessEqual(wilson(457, 10000)[1], .05)
        self.assertGreater(wilson(458, 10000)[1], .05)
        for target, boundaries in ((.2, ((8079, True), (8078, False))), (0, ((457, True), (458, False)))):
            for count, want in boundaries:
                rows, key, indexes = self.changed_cell("calibrated-reference", {}, target=target)
                for j, i in enumerate(indexes): rows[i] = {**rows[i], "support": count//5+(j < count%5)}
                self.assertEqual(self.gate(records=rows)["result"]["cell_reports"][key]["passed"], want)
        for inadequate, want in ((84, True), (85, False)):
            self.assertEqual(wilson(inadequate, 2000+inadequate)[1] <= .05, want)
            self.assertEqual(self.cell_passes("calibrated-reference", dict(inadequate=inadequate,
                executions=2000+inadequate, attempts=2000+inadequate), target=.2), want)
        for target, gain, want in ((0, 0, True), (0, .001, False), (.2, .2, True), (.2, .199, False)):
            self.assertEqual(self.cell_passes("calibrated-reference", {"realised_gain": gain}, target=target), want)

    def test_reachability_gain_and_missing_mask_types(self):
        for kind in ("calibrated-reference", "missing"):
            for field, value in (("reachable", False), ("reachable", "false"), ("realised_gain", True)):
                self.assertFalse(self.cell_passes(kind, {field: value}, target=.2))
        for value in (False, "false"):
            self.assertFalse(self.cell_passes("missing", {"mask_reachable": value}))
        for target in (0, .1, .2):
            self.assertTrue(self.cell_passes("missing", {"realised_gain": target}, target=target))
            self.assertFalse(self.cell_passes("missing", {"realised_gain": .001 if target == 0 else target-.001}, target=target))
        for gain in (float("nan"), float("inf")):
            with self.assertRaises(ValueError): self.cell_passes("calibrated-reference", {"realised_gain": gain}, target=.2)

    def test_rev8_maxima_arithmetic_and_p_gate_cpu_boundary(self):
        # Only gate CPU is nonzero here to put the analytic boundary exactly at 2880.
        inputs = {k: 0 for k in INPUTS}
        inputs["gate_cpu_per_check"] = 2880*3600/(2*18172*5*8000)
        fit = resource_fit(inputs, self.inventory, self.manifest, self.host)
        self.assertEqual(fit["core_hours"], 2880)
        self.assertEqual(fit["wall_days"], 15)
        self.assertTrue(fit["passed"])
        above = {**inputs, "gate_cpu_per_check": inputs["gate_cpu_per_check"]*(1+1e-10)}
        self.assertFalse(resource_fit(above, self.inventory, self.manifest, self.host)["passed"])
        self.assertFalse(resource_fit(inputs, {**self.inventory, "p_gate": .49999}, self.manifest, self.host)["passed"])
        first = {**self.measurement_run, "integrity_failure": "closure mismatch", "preserved_evidence_sha256": "a"*64,
                 "binding_inputs": {**self.inputs, "construction_cpu_per_unit": .01}}
        second = {**self.measurement_run, "id": "second", "binding_inputs": {**self.inputs, "gate_cpu_per_check": .02}}
        maxima, reports = maximum_completed_inputs([first, second], self.host)
        self.assertEqual(maxima, {**self.inputs, "construction_cpu_per_unit": .01, "gate_cpu_per_check": .02})
        self.assertEqual(reports[0]["classification"], "completed-invalid-included")
        resource = resource_fit(maxima, self.inventory, self.manifest, self.host)
        expected = 2*(18172*5*1200*.01 + 16128*5*1200*.000001 +
            2380*5*(1200*.000001+.000001)*(2044/2380) +
            5*4000*(17920*.000001+252*.000001) + 18172*5*8000*.02)/3600
        self.assertAlmostEqual(resource["core_hours"], expected)
        # Full report retains all per-run inputs, flags, causes and identities.
        full = self.gate(runs=[first, second])["result"]
        self.assertEqual([r["run"] for r in full["measurement_runs"]], [first, second])
        self.assertEqual(full["resources"]["maximum_inputs"], maxima)

    def test_memory_boundary_and_bound_host(self):
        budget = self.host["hw.memsize"]-8*1024**3
        self.assertEqual(budget, 60129542144)
        exact = {**self.inputs, "construction_rss": budget/8}
        self.assertTrue(resource_fit(exact, self.inventory, self.manifest, self.host)["passed"])
        self.assertFalse(resource_fit({**exact, "construction_rss": budget/8+1}, self.inventory, self.manifest, self.host)["passed"])
        for field, value in (("hw.model", "other"), ("hw.memsize", 68719476735),
                             ("logical_cpus", 17), ("os_build", "other")):
            with self.assertRaises(ValueError): self.gate(host={**self.host, field: value})
        # No caller-supplied memory_bytes or resources can override measured maxima/host.
        payloads, pins = self.bundle()
        with self.assertRaises(TypeError):
            evaluate(self.manifest_bytes, provenance_payloads=payloads, provenance_digests=pins,
                     synthetic_only=True, resources={"memory_bytes": 1})

    def test_rerun_qualifications_evidence_and_measurement_types(self):
        valid = self.measurement_run
        crash = {**valid, "completed": False, "integrity_failure": "non-resource crash", "preserved_evidence_sha256": "a"*64}
        rerun = {**valid, "id": "second"}
        self.assertEqual(maximum_completed_inputs([crash, rerun], self.host)[0], self.inputs)
        os_failure = {**crash, "integrity_failure": "independent OS malfunction", "contemporaneous_os_evidence_sha256": "b"*64}
        maximum_completed_inputs([os_failure, rerun], self.host)
        for runs in ([valid, rerun], [crash, valid], [crash], [{**valid, "completed": False}],
                     [{**valid, "completed": "false"}],
                     [{**crash, "preserved_evidence_sha256": None}, rerun],
                     [{**os_failure, "contemporaneous_os_evidence_sha256": None}, rerun]):
            with self.assertRaises(ValueError): maximum_completed_inputs(runs, self.host)
        for cause in ("host load", "operator termination", "resource exhaustion"):
            with self.assertRaises(ValueError): maximum_completed_inputs([{**crash, "integrity_failure": cause}, rerun], self.host)
        for value in (True, float("inf"), float("nan"), -1):
            with self.assertRaises(ValueError):
                maximum_completed_inputs([{**valid, "binding_inputs": {**self.inputs, "complete_trial_cpu": value}}], self.host)
        with self.assertRaises(ValueError): maximum_completed_inputs([{**valid, "binding_inputs": {"supplied_total": 1}}], self.host)

    def test_predev_none_no_bypass_and_postdev_tampering(self):
        rows = [{**r, "B": 1998} for r in self.records]
        self.assertEqual(self.gate(records=rows)["result"]["status"], "not testable before dev")
        rows, _, _ = self.changed_cell("requested-coverage", {"B": 1998})
        pre = self.gate(records=rows)
        payloads, pins = self.bundle(records=rows)
        self.assertEqual(self.select(pre, payloads=payloads, pins=pins)["status"], "not testable")
        self.assertEqual(self.select(pre, payloads=payloads, pins=pins)["selected_M"], 700)
        bad = decode(pre["sealed_result"]); bad["rows"]["700"] = True
        with self.assertRaises(ValueError):
            select_postdev(encode(bad), pre["result_sha256"], provenance_payloads=payloads,
                           provenance_digests=pins, synthetic_only=True)
        for patch in ({"Mhat": True}, {"c_U": True}, {"predev_sha256": "0"*64},
                      {"locked_inventory_sha256": "0"*64}, {"bootstrap_evidence": None}):
            with self.assertRaises(ValueError): self.select(pre, payloads=payloads, pins=pins, evidence_changes=patch)
        with self.assertRaises(ValueError): self.select(pre)  # Original records cannot replace sealed failed rows.
        postpayloads = dict(payloads, mhat_evidence=encode(dict(Mhat=700)))
        postpins = {k: digest(v) for k, v in postpayloads.items()}
        postpayloads["mhat_evidence"] = encode(dict(Mhat=2000))
        with self.assertRaises(ValueError):
            select_postdev(pre["sealed_result"], pre["result_sha256"], provenance_payloads=postpayloads,
                           provenance_digests=postpins, synthetic_only=True)

    def test_ordered_typed_verdict_and_interval_boundaries(self):
        effect = dict(integrity=True, violating=False, adequate=True, upper=.4, unresolved=False, point=.3, lower=.2)
        self.assertEqual(verdict(**effect), "support")
        for patch, want in ((dict(integrity=False, upper=float("nan")), "inconclusive"),
                            (dict(violating=True, upper=float("nan")), "not supported"),
                            (dict(upper=float("nan")), "inconclusive"),
                            (dict(upper=.10, lower=0, point=.05), "inconclusive"),
                            (dict(upper=math.nextafter(.10, 0), lower=0, point=.05), "not supported"),
                            (dict(integrity=False, violating=True), "inconclusive"),
                            (dict(adequate=False, upper=.05, lower=0), "inconclusive"),
                            (dict(unresolved=True, upper=.05, lower=0), "not supported"),
                            (dict(unresolved=True), "inconclusive"),
                            (dict(point=.10, lower=0), "inconclusive"),
                            (dict(point=.10, lower=.001), "support")):
            self.assertEqual(verdict(**{**effect, **patch}), want)
        for flag in ("integrity", "violating", "adequate", "unresolved"):
            with self.assertRaises(ValueError): verdict(**{**effect, flag: "false"})


if __name__ == "__main__":
    unittest.main()

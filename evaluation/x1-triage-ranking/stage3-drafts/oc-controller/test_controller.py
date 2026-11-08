"""Synthetic counter/schedule regressions for all scoped round-1 findings."""
import hashlib
import hmac
import json
from pathlib import Path
import subprocess
import tempfile
import unittest

from controller import (Population, SafetySchedule, ORDER_SHA256, POLICY_SHA256,
                        canonical, criticality_key, population_identity, seed, wilson)


class Meter:
    def __init__(self):
        self.core = 0.0
        self.now = 100.0

    def cpu(self):
        return self.core

    def clock(self):
        return self.now


class ControllerTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.counter = 0
        self.known = json.loads(Path(__file__).with_name("known_answers.json").read_text())

    def schedule(self, p_gate=.5):
        self.counter += 1
        meter = Meter()
        path = Path(self.temporary.name)/f"freeze-{self.counter}.json"
        schedule = SafetySchedule(freeze_path=path, p_gate=p_gate, meter=meter.cpu,
                                  clock=meter.clock, oc_started_at=100.0)
        return schedule, meter, path

    @staticmethod
    def drive(population, adequate=True, count=2000):
        for _ in range(count):
            population.step(lambda _: True, lambda _: adequate)

    def finish_group(self, schedule, adequate=True):
        populations = schedule.next_group(binding_rows_complete=True)
        for p in populations:
            self.drive(p, adequate=adequate)
        for j in (0.0, .1, 1.0, 3.0):
            schedule.run_criticality(j, lambda key: {"key": key, "completed": True})
        return populations

    def test_signed_identity_exact_dimensions_and_independent_order_pin(self):
        schedule, _, path = self.schedule()
        self.assertEqual(len(schedule.ordered), 2016)
        required = {"cap_M", "prevalence", "realised_gain", "scanner_control_true_at_prevalence",
                    "rank_structure", "size_dependence", "unresolved_kappa"}
        zero = 0
        for identity in schedule.ordered:
            kind, dims = json.loads(identity)
            self.assertEqual(kind, "safety-interval-group")
            self.assertEqual(set(dims), required)
            self.assertEqual(identity, population_identity(kind, dims))
            for name, axis in schedule.manifest["final_safety_OC"]["dimensions"].items():
                self.assertIn(dims[name], axis)
            zero += dims["unresolved_kappa"] == 0
        self.assertEqual(zero, 336)
        self.assertEqual(schedule.ordered[0], self.known["first_identity"])
        self.assertEqual(schedule.ordered[-1], self.known["last_identity"])
        self.assertEqual(hashlib.sha256(json.dumps(list(schedule.ordered), separators=(",", ":")).encode()).hexdigest(), self.known["ordered_groups_sha256"])
        self.assertEqual(ORDER_SHA256, self.known["ordered_groups_sha256"])
        self.assertEqual(POLICY_SHA256, self.known["policy_sha256"])
        record = json.loads(path.read_text())
        self.assertEqual(record["ordered_groups_sha256"], ORDER_SHA256)
        self.assertEqual(record["policy_sha256"], POLICY_SHA256)
        self.assertEqual(len(record["source_sha256"]), 3)
        self.assertIn("created_at_utc", record)
        actual = subprocess.check_output(["openssl", "dgst", "-sha256", "-hmac", "169-x1-safety-order"], input=schedule.ordered[0].encode()).decode().split()[-1]
        self.assertEqual(actual, self.known["hmac_order_first"])

    def test_seed_known_answers_openssl_and_input_refusals(self):
        for row in self.known["seeds"]:
            cell = row["cell"]
            raw = cell.encode() if isinstance(cell, str) else json.dumps(cell, sort_keys=True, separators=(",", ":")).encode()
            message = b"\0".join((raw, str(row["population"]).encode(), str(row["attempt"]).encode(), row["phase"].encode()))
            actual = subprocess.check_output(["openssl", "dgst", "-sha256", "-hmac", "169-x1-oc-v4"], input=message).decode().split()[-1]
            self.assertEqual(actual, row["hmac_sha256"])
            self.assertEqual(seed(cell, row["population"], row["attempt"], row["phase"]), int(actual[:16], 16))
        for p, a, phase in ((True, 0, "audit"), (-1, 0, "audit"), (0, -1, "audit"),
                             (0, .5, "audit"), (0, 0, "unknown"), (0, 1, "population")):
            with self.assertRaises(ValueError):
                seed(self.known["first_identity"], p, a, phase)

    def test_manifest_substitution_and_create_once_freeze(self):
        schedule, meter, path = self.schedule()
        with self.assertRaises(FileExistsError):
            SafetySchedule(freeze_path=path, p_gate=.5, meter=meter.cpu, clock=meter.clock, oc_started_at=100)
        with self.assertRaises(ValueError):
            SafetySchedule(freeze_path=Path(self.temporary.name)/"bad", p_gate=.5,
                           meter=meter.cpu, clock=meter.clock, oc_started_at=100, manifest_bytes=b"{}")
        with self.assertRaises(AttributeError):
            schedule.ordered = tuple(reversed(schedule.ordered))
        with self.assertRaises(TypeError):
            schedule.policy["reserved_worker_slots"] = 7
        with self.assertRaises(TypeError):
            schedule.policy["limits"]["core_hours"] = 1

    def test_serial_lease_refuses_reentry_and_gold_inside_trial_discards_result(self):
        schedule, _, _ = self.schedule()
        p = schedule.next_group(binding_rows_complete=True)[0]
        def trial(_):
            with self.assertRaises(RuntimeError):
                p.step(lambda _: self.fail("reentrant gate executed"), lambda _: True)
            return True
        p.step(lambda _: True, trial)
        self.assertEqual(p.executions, 1)
        def stop_inside_trial(_):
            schedule.stop_gold()
            return True
        with self.assertRaises(RuntimeError):
            p.step(lambda _: True, stop_inside_trial)
        report = schedule.gold_report()
        counts = report["groups"][0]["populations"][0]
        self.assertEqual((counts["executions"], counts["eligible"], counts["crashed"]), (2, 1, 1))
        self.assertEqual(report["groups"][0]["status"], "not completed by gold")

    def test_gate_rejections_never_execute_and_attempt_cap(self):
        schedule, _, _ = self.schedule(1)
        p = schedule.binding_population("binding-cell", 0)
        for _ in range(4000):
            p.step(lambda _: False, lambda _: self.fail("rejected gate executed"))
        self.assertTrue(p.exhausted)
        self.assertEqual((p.attempts, p.executions), (4000, 0))
        with self.assertRaises(RuntimeError):
            p.step(lambda _: True, lambda _: True)
        report = p.report()
        self.assertEqual(report["gate_failures"], 4000)
        self.assertIsNone(report["conditional_inadequacy_Wilson_MC95"])

    def test_execution_cap_is_independent_inadequacy_is_charged(self):
        schedule, _, _ = self.schedule(.5)
        p = schedule.binding_population("binding-cell", 0)
        self.drive(p, adequate=False, count=4000)
        self.assertTrue(p.exhausted)
        self.assertEqual((p.attempts, p.executions, p.inadequate), (4000, 4000, 4000))
        self.assertLess(p.attempts, p.attempt_cap)
        with self.assertRaises(RuntimeError):
            p.step(lambda _: True, lambda _: True)

    def test_completion_on_last_attempt_and_execution_wins(self):
        schedule, _, _ = self.schedule(1)
        p = schedule.binding_population("attempt-last", 0)
        for _ in range(2000):
            p.step(lambda _: False, lambda _: self.fail("unexpected execution"))
        self.drive(p)
        self.assertEqual(p.attempts, p.attempt_cap)
        self.assertTrue(p.complete)
        self.assertFalse(p.exhausted)
        with self.assertRaises(RuntimeError):
            p.step(lambda _: True, lambda _: True)
        other, _, _ = self.schedule(.5)
        q = other.binding_population("execution-last", 0)
        self.drive(q, adequate=False)
        self.drive(q)
        self.assertEqual(q.executions, 4000)
        self.assertTrue(q.complete)
        self.assertFalse(q.exhausted)

    def test_safety_stops_at_executions_including_inadequacy(self):
        schedule, _, _ = self.schedule()
        p = schedule.next_group(binding_rows_complete=True)[0]
        self.drive(p, adequate=False)
        self.assertTrue(p.complete)
        self.assertEqual((p.executions, p.eligible, p.inadequate), (2000, 0, 2000))
        with self.assertRaises(RuntimeError):
            p.step(lambda _: True, lambda _: True)

    def test_population_lane_mapping_and_finish_membership(self):
        schedule, _, _ = self.schedule()
        binding = schedule.binding_population("binding", 0)
        with self.assertRaises(RuntimeError):
            schedule.next_group(binding_rows_complete=False)
        pops = schedule.next_group(binding_rows_complete=True)
        with self.assertRaises(RuntimeError):
            pops[1].step(lambda _: True, lambda _: True)
        for invalid in ([pops[0]]*5, pops[:4], [binding]+list(pops[1:]),
                        [Population("foreign", 0, .5, True)]+list(pops[1:])):
            with self.assertRaises(RuntimeError):
                schedule.finish(invalid, {})
        self.assertEqual([p.population_index for p in pops], list(range(5)))
        with self.assertRaises(RuntimeError):
            schedule.finish(pops, {})
        with self.assertRaises(RuntimeError):
            schedule.run_criticality(0.0, lambda key: {"key": key, "completed": True})

    def test_criticality_exact_keys_order_and_auto_completion(self):
        schedule, _, _ = self.schedule()
        pops = schedule.next_group(binding_rows_complete=True)
        identity = pops[0].group_identity
        for p in pops:
            self.drive(p)
        with self.assertRaises(ValueError):
            schedule.run_criticality(.1, lambda key: {"key": key, "completed": True})
        with self.assertRaises(ValueError):
            schedule.run_criticality(0.0, lambda key: {"key": key, "completed": False})
        for j in (0, .1, 1, 3):
            actual = schedule.run_criticality(j, lambda key: {"key": key, "completed": True})
            self.assertEqual(actual, criticality_key(identity, float(j)))
        following = schedule.next_group(binding_rows_complete=True)
        self.assertEqual(following[0].group_identity, schedule.ordered[1])
        schedule.stop_gold()
        report = schedule.gold_report()
        self.assertEqual(report["groups"][0]["status"], "measured")
        self.assertEqual(len(report["groups"][0]["populations"]), 5)
        self.assertEqual(len(report["groups"][0]["criticality"]), 4)
        self.assertEqual(report["pooled_group_indices"], [0])

    def test_gold_before_during_and_after_last_trial_is_irreversible(self):
        unstarted, _, _ = self.schedule()
        with self.assertRaises(RuntimeError):
            unstarted.gold_report()
        unstarted.stop_gold()
        self.assertTrue(all(g["status"] == "not completed by gold" for g in unstarted.gold_report()["groups"]))
        for count in (1, 1999, 2000):
            schedule, _, _ = self.schedule()
            pops = schedule.next_group(binding_rows_complete=True)
            self.drive(pops[0], count=count)
            schedule.stop_gold()
            before = schedule.gold_report()
            self.assertEqual(before["groups"][0]["populations"][0]["executions"], count)
            with self.assertRaises(RuntimeError):
                pops[0].step(lambda _: True, lambda _: True)
            with self.assertRaises(RuntimeError):
                schedule.finish(pops, {})
            self.assertEqual(before, schedule.gold_report())
            self.assertEqual(before["events"][-1]["cause"], "gold")

    def test_caps_at_execution_boundary_and_wall_clock_unpaused(self):
        schedule, meter, _ = self.schedule()
        p = schedule.next_group(binding_rows_complete=True)[0]
        meter.core = 1440
        self.drive(p, count=1)  # Actual meter is not multiplied by two.
        meter.core = 2880
        with self.assertRaises(RuntimeError):
            p.step(lambda _: self.fail("gate ran after cap"), lambda _: True)
        self.assertEqual(p.attempts, 1)
        self.assertEqual(schedule.gold_report()["events"][-1]["cause"], "OC core-hour cap")
        other, clock, _ = self.schedule()
        q = other.next_group(binding_rows_complete=True)[0]
        other.pause({"kind": "closure mismatch", "evidence_sha256": "a"*64})
        clock.now += 15*86400
        with self.assertRaises(RuntimeError):
            other.resume({"kind": "reviewed integrity resolution", "evidence_sha256": "b"*64})
        self.assertEqual(other.gold_report()["events"][-1]["cause"], "OC wall-day cap")
        self.assertEqual(q.attempts, 0)

    def test_cap_during_last_trial_and_criticality_freezes_partial_counts(self):
        schedule, meter, _ = self.schedule()
        p = schedule.next_group(binding_rows_complete=True)[0]
        self.drive(p, count=1999)
        def last_trial(_):
            meter.core = 2880
            return True
        with self.assertRaises(RuntimeError):
            p.step(lambda _: True, last_trial)
        self.assertEqual(schedule.gold_report()["groups"][0]["populations"][0]["executions"], 2000)
        other, _, _ = self.schedule()
        pops = other.next_group(binding_rows_complete=True)
        for population in pops:
            self.drive(population)
        for j in (0.0, .1, 1.0):
            other.run_criticality(j, lambda key: {"key": key, "completed": True})
        def gold_at_last_phase(key):
            other.stop_gold()
            return {"key": key, "completed": True}
        with self.assertRaises(RuntimeError):
            other.run_criticality(3.0, gold_at_last_phase)
        frozen = other.gold_report()
        self.assertEqual(frozen["groups"][0]["status"], "not completed by gold")
        self.assertEqual(len(frozen["groups"][0]["criticality"]), 3)
        self.assertEqual(frozen["pooled_group_indices"], [])

    def test_integrity_pause_resume_preserves_attempt_and_group(self):
        schedule, meter, _ = self.schedule()
        p = schedule.next_group(binding_rows_complete=True)[0]
        self.drive(p, count=3)
        schedule.pause({"kind": "closure mismatch", "evidence_sha256": "a"*64})
        with self.assertRaises(RuntimeError):
            p.step(lambda _: True, lambda _: True)
        with self.assertRaises(ValueError):
            schedule.pause({"kind": "bad safety outcome", "evidence_sha256": "a"*64})
        meter.now += 900
        schedule.resume({"kind": "reviewed integrity resolution", "evidence_sha256": "b"*64})
        seen = []
        p.step(lambda i: seen.append(i) is None, lambda _: True)
        self.assertEqual(seen, [3])
        self.assertEqual(p.group_identity, schedule.ordered[0])
        self.assertEqual([e["event"] for e in schedule.events], ["start", "pause", "resume"])
        self.assertEqual(schedule.events[-1]["wall_seconds"], 900)

    def test_crash_and_nonbool_latch_integrity_with_reconciled_counts(self):
        for trial in (lambda _: (_ for _ in ()).throw(RuntimeError("synthetic crash")), lambda _: "false"):
            schedule, _, _ = self.schedule()
            p = schedule.next_group(binding_rows_complete=True)[0]
            with self.assertRaises((RuntimeError, ValueError)):
                p.step(lambda _: True, trial)
            self.assertTrue(schedule.paused)
            self.assertEqual((p.executions, p.eligible, p.inadequate, p.crashed), (1, 0, 0, 1))
            self.assertEqual(p.report()["crashed"], 1)
            with self.assertRaises(RuntimeError):
                schedule.resume({"kind": "reviewed integrity resolution", "evidence_sha256": "b"*64})
            schedule.resolve_crash(p, 0, False, {"kind": "reviewed integrity resolution", "evidence_sha256": "c"*64})
            schedule.resume({"kind": "reviewed integrity resolution", "evidence_sha256": "b"*64})
            self.assertEqual((p.executions, p.inadequate, p.crashed), (1, 1, 0))
        with self.assertRaises(TypeError):
            Population("identity", 0, .5, attempts=2000)

    def test_exhaustion_is_bound_only_to_active_group(self):
        schedule, _, _ = self.schedule(1)
        p = schedule.next_group(binding_rows_complete=True)[0]
        for _ in range(4000):
            p.step(lambda _: False, lambda _: self.fail("rejected gate executed"))
        schedule.stop_gold()
        report = schedule.gold_report()
        self.assertEqual(report["groups"][0]["status"], "unmeasured; cell fails")
        self.assertEqual(report["groups"][0]["populations"][0]["gate_failures"], 4000)
        self.assertEqual(report["groups"][1]["status"], "not completed by gold")
        self.assertEqual(report["groups"][1]["populations"], [])
        self.assertEqual(report["pooled_group_indices"], [])

    def test_unreachable_provenance_advances_in_order_and_grid_completion_logged(self):
        schedule, _, _ = self.schedule()
        for index in range(2016):
            pops = schedule.next_group(binding_rows_complete=True)
            self.assertEqual(pops[0].group_identity, schedule.ordered[index])
            record = {"group_identity": pops[0].group_identity, "reachable": False,
                      "target": json.loads(pops[0].group_identity)[1]["realised_gain"], "calibration_source_sha256": "ed670a6bf8076d791c06cfc2426aff03a0a2f9b1730982cf3fdfb2ec13bc49a2"}
            if index == 0:
                with self.assertRaises(ValueError):
                    schedule.mark_unreachable(record, "0"*64)
                wrong_source = {**record, "calibration_source_sha256": "a"*64}
                with self.assertRaises(ValueError):
                    schedule.mark_unreachable(wrong_source, hashlib.sha256(canonical(wrong_source)).hexdigest())
            schedule.mark_unreachable(record, hashlib.sha256(canonical(record)).hexdigest())
        self.assertTrue(schedule.terminated)
        report = schedule.gold_report()
        self.assertEqual(report["events"][-1]["cause"], "grid complete")
        self.assertEqual(report["pooled_group_indices"], [])
        self.assertTrue(all(g["status"] == "unmeasured (unreachable target)" for g in report["groups"]))
        self.assertTrue(all(p["attempts"] == 0 for g in report["groups"] for p in g["populations"]))

    def test_outcome_independent_dispatch_events_statuses_and_integrity_stop(self):
        outcomes = []
        for adequate in (False, True):
            schedule, _, _ = self.schedule()
            self.finish_group(schedule, adequate=adequate)
            next_identity = schedule.next_group(binding_rows_complete=True)[0].group_identity
            schedule.stop_integrity({"kind": "independent OS malfunction", "evidence_sha256": "a"*64})
            report = schedule.gold_report()
            outcomes.append((next_identity, report["events"], [g["status"] for g in report["groups"]]))
        self.assertEqual(outcomes[0], outcomes[1])

    def test_probability_meter_and_wilson_boundaries(self):
        for value in (float("nan"), -1, 0, 1.01, True):
            with self.assertRaises(ValueError):
                self.schedule(value)
        schedule, meter, _ = self.schedule(.3)
        p = schedule.binding_population("binding", 0)
        self.assertEqual(p.attempt_cap, 13334)
        for value in (-1, float("nan"), True):
            meter.core = value
            with self.assertRaises(ValueError):
                p.step(lambda _: True, lambda _: True)
        # Independent 60-digit Decimal endpoint: z²/(2000+z²), z=1.959963984540054.
        self.assertAlmostEqual(wilson(0, 2000)[1], 0.001917047281252934, places=14)
        self.assertAlmostEqual(wilson(2000, 2000)[0], 0.9980829527187471, places=14)
        self.assertEqual(wilson(0, 0), None)


if __name__ == "__main__":
    unittest.main()

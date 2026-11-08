"""Counter and schedule invariants; synthetic callbacks, no actual OC."""
import hashlib
import hmac
import json
import unittest
from pathlib import Path
from controller import Population, SafetySchedule, canonical, seed, safety_groups


class ControllerTests(unittest.TestCase):
    def test_rejections_never_execute_and_attempt_exhaustion(self):
        p = Population(1)
        for _ in range(4000):
            p.step(lambda _: False, lambda _: self.fail("rejected gate executed"))
        self.assertEqual((p.attempts, p.executions), (4000, 0))
        self.assertTrue(p.exhausted)
        with self.assertRaises(RuntimeError):
            p.step(lambda _: True, lambda _: True)

    def test_inadequacy_consumes_independent_execution_cap(self):
        p = Population(.5)
        for _ in range(4000):
            p.step(lambda _: True, lambda _: False)
        self.assertEqual((p.attempts, p.executions, p.inadequate), (4000, 4000, 4000))
        self.assertTrue(p.exhausted)

    def test_binding_eligible_stop_safety_execution_stop(self):
        p, s = Population(.5), Population(.5, safety=True)
        for _ in range(2000):
            p.step(lambda _: True, lambda _: True)
            s.step(lambda _: True, lambda _: False)
        self.assertTrue(p.complete)
        self.assertTrue(s.complete)
        self.assertEqual(s.eligible, 0)
        self.assertFalse(s.exhausted)

    def test_freeze_and_criticality_order(self):
        manifest = json.loads((Path(__file__).resolve().parents[2]/"oc-manifest.json").read_text())
        full = list(safety_groups(manifest))
        self.assertEqual(len(full), 2016)
        self.assertEqual(sum(g["unresolved_kappa"] == 0 for g in full), 336)
        self.assertEqual(len({canonical(g) for g in full}), 2016)
        groups = [{"kind": "safety-interval-group", "index": i} for i in range(10)]
        schedule = SafetySchedule(groups)
        order = sorted(groups, key=lambda g: hmac.digest(b"169-x1-safety-order", canonical(g), "sha256"))
        self.assertEqual(schedule.freeze["ordered_groups_sha256"], hashlib.sha256(canonical(order)).hexdigest())
        self.assertEqual(schedule.freeze, SafetySchedule(list(reversed(groups))).freeze)
        with self.assertRaises(RuntimeError):
            schedule.next_group(binding_rows_complete=False, core_hours=0, wall_days=0)
        self.assertEqual(schedule.next_group(binding_rows_complete=True, core_hours=0, wall_days=0), order[0])
        with self.assertRaises(RuntimeError):
            schedule.next_group(binding_rows_complete=True, core_hours=0, wall_days=0)
        pops = [Population(1, safety=True, attempts=2000, executions=2000) for _ in range(5)]
        schedule.finish(pops, criticality_complete=True)
        self.assertEqual(schedule.next_group(binding_rows_complete=True, core_hours=1, wall_days=1), order[1])
        self.assertIsNone(schedule.next_group(binding_rows_complete=True, core_hours=2880, wall_days=1))
        report = schedule.gold_report()
        self.assertEqual(report["pooled_group_indices"], [0])
        self.assertEqual(report["events"][0]["cause"], "OC resource cap")
        self.assertEqual(report["statuses"][1:], ["not completed by gold"]*9)
        schedule.policy["workers"] = 7
        with self.assertRaises(RuntimeError):
            schedule.gold_report()

    def test_exhaustion_precedence_and_seed_phase_separation(self):
        schedule = SafetySchedule([{"kind": "safety-interval-group", "index": 0}])
        p = Population(1, safety=True, attempts=4000)
        self.assertEqual(schedule.gold_report([p])["statuses"], ["unmeasured; cell fails"])
        self.assertNotEqual(seed({"kind": "missing"}, 0, 0, "audit"), seed({"kind": "missing"}, 0, 0, "bootstrap"))
        self.assertNotEqual(seed({}, 0, 0, "audit"), seed({}, 0, 1, "audit"))
        with self.assertRaises(ValueError):
            Population(0)


if __name__ == "__main__":
    unittest.main()

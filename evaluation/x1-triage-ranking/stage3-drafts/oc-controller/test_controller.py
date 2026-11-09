"""Synthetic state-machine tests, including real fsync/replay and threaded stops.

Some transition tests seed a feasible prior counter state through the private
journal reducer; these are unit fixtures, never claimed as actual OC trials.
"""
import copy
from dataclasses import replace
import importlib.util
import inspect
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import threading
from types import SimpleNamespace
import unittest
from unittest.mock import patch

HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location("x1_controller_test", HERE / "controller.py")
c = importlib.util.module_from_spec(spec)
sys.modules[spec.name] = c
spec.loader.exec_module(c)
REVIEW = {"reviewed": True, "evidence_sha256": "a" * 64}
PAUSE = {"kind": "execution integrity", "evidence_sha256": "b" * 64}


class Meter:
    def __init__(self): self.core, self.now = 0., 100000.
    def __call__(self): return self.core
    def clock(self): return self.now


class ControllerTests(unittest.TestCase):
    def make(self, *, root=None, meter=None, **kw):
        if root is None:
            temp = tempfile.TemporaryDirectory()
            self.addCleanup(temp.cleanup)
            root = Path(temp.name)
        meter = Meter() if meter is None else meter
        args = dict(p_gate=.5, inventory_sha256="c" * 64, meter=meter,
                    oc_started_at=100000., clock=meter.clock, run_root=root)
        args.update(kw)
        # Synthetic fixtures reserve no real production-scale disk.
        with patch.object(c.shutil, "disk_usage", return_value=SimpleNamespace(free=10**16)):
            schedule = c.SafetySchedule(**args)
        self.addCleanup(schedule.close)
        return schedule, meter, root

    def lane(self, s, index=0, slot=0):
        return s.binding_population(c.canonical({"cell": slot}).decode(), index, slot=slot)

    def prepared(self, profile, index, seed):
        dimensions = json.loads(profile)[1]
        return c._kernel().prepare_population(seed=seed, run_count=1, sizes=(50,),
            prevalence=dimensions["prevalence"], cap=dimensions["cap_M"])

    def seed_active(self, s, *, eligible=1999, inadequate=0, complete=False):
        # Feasible journal fixture immediately before the final trial per lane.
        with s._state_lock:
            s._emit("dispatch", {"identity": s.ordered[0]})
            for index in range(5):
                s._new_population(s.ordered[0], index, True, index)
        s._calibrate(self.prepared)
        with s._state_lock:
            for key in s._active["populations"]:
                p = copy.deepcopy(s._populations[key])
                count = 2000 if complete else eligible + inadequate
                p.update(attempts=count, executions=count, gate_passes=count,
                         eligible=2000 if complete else eligible, inadequate=0 if complete else inadequate)
                s._save_population(key, p)

    def start_run(self, s, gate=None, trial=None, calibration=None, criticality=None):
        errors = []
        def execute():
            try:
                s.run(gate or (lambda *a: True), trial or (lambda *a: True),
                      calibration or self.prepared,
                      criticality or (lambda identity, j, key: {"key": key, "completed": True}))
            except BaseException as error:
                errors.append(error)
        thread = threading.Thread(target=execute, daemon=True)
        thread.start()
        return thread, errors

    def stop_join(self, s, thread, errors):
        s.stop_gold()
        thread.join(15)
        self.assertFalse(thread.is_alive())
        self.assertEqual(errors, [])

    def assert_report(self, s):
        report = s.gold_report()
        self.assertEqual(tuple(g["identity"] for g in report["groups"]), s.ordered)
        self.assertEqual(len(set(g["identity"] for g in report["groups"])), 2016)
        self.assertEqual(sum(e["kind"] == "terminate" for e in report["events"]), 1)
        for group in report["groups"]:
            for p in group["populations"]:
                self.assertEqual(p["executions"], p["eligible"] + p["inadequate"] + p["crashed"])
        return report

    def test_independent_known_answers_and_canonical_seed_refusals(self):
        known = json.loads((HERE / "known_answers.json").read_bytes())
        s, _, _ = self.make()
        self.assertEqual(s.ordered[0], known["first_identity"])
        self.assertEqual(s.ordered[-1], known["last_identity"])
        self.assertEqual(c.digest(c.canonical(s.ordered)), known["ordered_groups_sha256"])
        self.assertEqual(c.POLICY_SHA256, known["policy_sha256"])
        for row in known["seeds"]:
            self.assertEqual(c.seed(row["cell"], row["population"], row["attempt"], row["phase"]), row["seed64"])
        for cell in ('{"b":1,"a":2}', '{"a": 2}', 'garbage', 'NaN'):
            with self.subTest(cell=cell), self.assertRaises(ValueError): c.seed(cell, 0, 0, "audit")
        for args in (({}, True, 0, "audit"), ({}, -1, 0, "audit"), ({}, 0, -1, "audit"), ({}, 0, 1, "population"), ({}, 0, 0, "bad")):
            with self.assertRaises(ValueError): c.seed(*args)
        expected = [float(v) for v in known["wilson_50_2000_decimal60"]]
        for actual, want in zip(c.wilson(50, 2000), expected): self.assertAlmostEqual(actual, want, places=15)
        self.assertIsNone(c.wilson(0, 0))
        for args in ((True, 2), (3, 2), (-1, 2)):
            with self.assertRaises(ValueError): c.wilson(*args)

    def test_source_order_policy_refusals_before_freeze(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            bad = root / "bad.py"
            bad.write_text("# substituted signed source\n")
            with self.assertRaises(RuntimeError): self.make(root=root, kernel_path=bad)
            for target, replacement in (("ORDER_SHA256", "0"*64), ("POLICY", {"changed": True})):
                with patch.object(c, target, replacement), self.assertRaises(RuntimeError): self.make(root=root)
                self.assertFalse((root / "x1-oc-controller/freeze.json").exists())
            with self.assertRaises(ValueError): self.make(root=root, manifest_bytes=b"{}")
        for probability in (0, -1, 1.01, True, float("nan")):
            with self.assertRaises(ValueError): self.make(p_gate=probability)

    def test_eight_concurrent_binding_slots_equal_serial_and_refuse_reentry(self):
        s, _, _ = self.make()
        lanes = [self.lane(s, slot=i) for i in range(8)]
        barrier = threading.Barrier(9)
        entered = [threading.Event() for _ in lanes]
        failures = []
        def worker(i):
            try:
                def trial(index):
                    entered[i].set()
                    barrier.wait(10)
                    return True
                lanes[i].step(lambda index: True, trial)
            except BaseException as error: failures.append(error)
        threads = [threading.Thread(target=worker, args=(i,)) for i in range(8)]
        for thread in threads: thread.start()
        for event in entered: self.assertTrue(event.wait(10))
        with self.assertRaises(ValueError): self.lane(s, slot=8)
        with self.assertRaises(RuntimeError): lanes[0].step(lambda i: True, lambda i: True)
        barrier.wait(10)
        for thread in threads: thread.join(10)
        self.assertEqual(failures, [])
        serial, _, _ = self.make()
        for i, lane in enumerate(lanes):
            other = self.lane(serial, slot=i)
            other.step(lambda n: True, lambda n: True)
            self.assertEqual(lane.report(), other.report())

    def test_binding_completion_caps_and_gate_failures(self):
        for adequate, complete in ((True, True), (False, False)):
            s, _, _ = self.make()
            lane = self.lane(s)
            with s._state_lock:
                p = copy.deepcopy(s._populations[lane._key])
                p.update(attempts=3999, executions=3999, gate_passes=3999, eligible=1999, inadequate=2000)
                s._save_population(lane._key, p)
            lane.step(lambda i: True, lambda i: adequate)
            self.assertEqual(lane.complete, complete)
            self.assertEqual(lane.exhausted, not complete)
            self.assertEqual(lane.report()["executions"], 4000)
            with self.assertRaises(RuntimeError): lane.step(lambda i: True, lambda i: True)
        s, _, _ = self.make(p_gate=.3)
        lane = self.lane(s)
        cap = 13334
        with s._state_lock:
            p = copy.deepcopy(s._populations[lane._key])
            p.update(attempts=cap-1, gate_failures=cap-1)
            s._save_population(lane._key, p)
        lane.step(lambda i: False, lambda i: self.fail("rejected gate executed trial"))
        self.assertTrue(lane.exhausted)
        self.assertEqual(lane.report()["gate_failures"], cap)

    def test_gate_window_gold_pause_and_cap_never_start_trial(self):
        for action in ("gold", "pause", "cap"):
            s, meter, _ = self.make()
            lane = self.lane(s)
            inside, release = threading.Event(), threading.Event()
            errors, called = [], []
            def gate(index): inside.set(); release.wait(10); return True
            def work():
                try: lane.step(gate, lambda i: called.append(i) or True)
                except c.Stopped: pass
                except BaseException as error: errors.append(error)
            thread = threading.Thread(target=work)
            thread.start()
            self.assertTrue(inside.wait(10))
            if action == "gold": s.stop_gold()
            elif action == "pause": s.pause(PAUSE)
            else: meter.core = 2880
            release.set(); thread.join(10)
            self.assertEqual(errors, [])
            self.assertEqual(called, [])
            self.assertEqual(lane.report()["executions"], 0)
            if action == "pause":
                s.resume(REVIEW)
                lane.step(lambda i: True, lambda i: called.append(i) or True)
                self.assertEqual(called, [0])
            s.stop_gold(); self.assert_report(s)

    def test_trial_window_gold_pause_and_cap_classification(self):
        for action in ("gold", "pause", "cap", "integrity"):
            s, meter, _ = self.make()
            lane = self.lane(s)
            inside, release = threading.Event(), threading.Event()
            errors = []
            def trial(index): inside.set(); release.wait(10); return True
            def work():
                try: lane.step(lambda i: True, trial)
                except c.Stopped: pass
                except BaseException as error: errors.append(error)
            thread = threading.Thread(target=work); thread.start()
            self.assertTrue(inside.wait(10))
            if action == "gold": s.stop_gold()
            elif action == "pause": s.pause(PAUSE)
            elif action == "integrity": s.stop_integrity(PAUSE)
            else: meter.core = 2880
            release.set(); thread.join(10)
            self.assertEqual(errors, [])
            report = lane.report()
            self.assertEqual(report["executions"], 1)
            self.assertEqual(report["eligible"], int(action == "cap"))
            self.assertEqual(report["crashed"], int(action != "cap"))
            if action == "pause":
                s.resume(REVIEW)
                seen = []
                lane.step(lambda i: self.fail("trial recovery reran gate"), lambda i: seen.append(i) or False)
                self.assertEqual(seen, [0])
                self.assertEqual(lane.report()["attempts"], 1)
                self.assertEqual(lane.report()["executions"], 1)
            s.stop_gold(); self.assert_report(s)

    def fault_case(self, phase, value):
        s, _, _ = self.make()
        lane = self.lane(s)
        calls = []
        def broken(index):
            calls.append(index)
            if isinstance(value, BaseException): raise value
            return value
        caught = []
        try:
            lane.step(broken if phase == "gate" else lambda i: True, broken if phase == "trial" else lambda i: True)
        except BaseException as error:
            caught.append(error)
        self.assertEqual(len(caught), 1)
        self.assertIsInstance(caught[0], c.Stopped)
        self.assertTrue(s.paused)
        before = lane.report()
        self.assertEqual(before["gate_failures"], 0)
        self.assertEqual(before["gate_faults"], int(phase == "gate"))
        with self.assertRaises(ValueError): s.resume({"reviewed": False, "evidence_sha256": "a"*64})
        s.resume(REVIEW)
        recovered = []
        lane.step(lambda i: recovered.append(("gate", i)) or True, lambda i: recovered.append(("trial", i)) or False)
        self.assertEqual(recovered, [("gate", 0), ("trial", 0)] if phase == "gate" else [("trial", 0)])
        self.assertEqual(lane.report()["attempts"], 1)
        self.assertEqual(lane.report()["inadequate"], 1)
        self.assertTrue(lane.report()["incidents"][0]["resolved"])
        self.assertFalse(hasattr(s, "resolve_crash"))
        self.assertNotIn("adequate", inspect.signature(s.resume).parameters)

    def test_gate_exception_is_latched(self): self.fault_case("gate", RuntimeError("gate crash"))
    def test_nonbool_gate_is_latched_not_gate_failure(self): self.fault_case("gate", 1)
    def test_trial_exception_is_latched(self): self.fault_case("trial", RuntimeError("trial crash"))
    def test_nonbool_trial_is_latched(self): self.fault_case("trial", "true")
    def test_baseexception_is_latched(self): self.fault_case("trial", KeyboardInterrupt())

    def test_owned_loop_calibration_order_disclosure_and_outcome_independence(self):
        projections = []
        for outcome in (True, False):
            s, _, _ = self.make()
            self.seed_active(s, eligible=1999 if outcome else 0, inadequate=0 if outcome else 1999)
            seen = []
            def calibration(profile, index, seed):
                s.stop_gold()  # Next controller-issued group, not an external next_group call.
                return self.prepared(profile, index, seed)
            report = s.run(lambda *a: True, lambda *a: seen.append(a[:3]) or outcome,
                           calibration, lambda identity, j, key: {"key": key, "completed": True})
            self.assertEqual(len(seen), 5)
            self.assertEqual(report["groups"][0]["status"], "measured")
            self.assertEqual(len(report["groups"][0]["criticality"]), 4)
            self.assertEqual(len(report["groups"][0]["calibration"]), 5)
            for record in report["groups"][0]["calibration"]:
                self.assertTrue(set(("target", "realised", "excess", "latent_gain", "reachable", "kernel_sha256", "record_sha256")) <= set(record))
                payload = {k: v for k, v in record.items() if k != "record_sha256"}
                self.assertEqual(c.digest(c.canonical(payload)), record["record_sha256"])
            kinds = [e["kind"] for e in report["events"]]
            for kind in ("dispatch", "phase_start", "calibration", "lane_start", "lane_stop", "criticality", "complete", "terminate"):
                self.assertIn(kind, kinds)
            self.assertTrue(all(type(e["timestamp"]) in (int, float) for e in report["events"]))
            projections.append([(e["kind"], e["data"].get("identity"), e["data"].get("population"), e["data"].get("j")) for e in report["events"]])
            for removed in ("next_group", "mark_unreachable", "finish", "run_criticality"):
                self.assertFalse(hasattr(s, removed))
            self.assert_report(s)
        self.assertEqual(*projections)

    def test_unreachable_population_three_before_attempt_and_profile_reuse(self):
        s, _, _ = self.make()
        called = []
        def calibration(profile, index, seed):
            called.append(index)
            prepared = self.prepared(profile, index, seed)
            if index == 3:
                prepared = replace(prepared, runs=tuple(replace(r, annotated=frozenset()) for r in prepared.runs))
            return prepared
        with s._state_lock:
            s._emit("dispatch", {"identity": s.ordered[0]})
            for i in range(5): s._new_population(s.ordered[0], i, True, i)
        s._calibrate(calibration)
        self.assertEqual(called, list(range(5)))
        self.assertFalse(s._active["calibration"][3]["reachable"])
        with s._state_lock:
            s._complete("unmeasured (unreachable target)")
            profile = s._profile(s.ordered[0])
            other = next(identity for identity in s.ordered[1:] if s._profile(identity) == profile)
            s._emit("dispatch", {"identity": other})
            for i in range(5): s._new_population(other, i, True, i)
        s._calibrate(lambda *a: self.fail("profile was recalibrated for a different kappa"))
        self.assertFalse(s._active["calibration"][3]["reachable"])
        self.assertEqual(called, list(range(5)))
        self.assertTrue(all(p["attempts"] == 0 for p in s._populations.values()))
        # Isolate the cross-kappa consistency guard with a corrupt private fixture.
        s._active["calibration"][3]["reachable"] = True
        with self.assertRaises(RuntimeError):
            s._calibrate(lambda *a: self.fail("cached profile invoked callback"))
        # A supplied assertion is not a PreparedPopulation and cannot skip work.
        other_s, _, _ = self.make()
        entered = threading.Event()
        def asserted(*args): entered.set(); return {"reachable": False}
        thread, errors = self.start_run(other_s, calibration=asserted)
        self.assertTrue(entered.wait(10))
        with other_s._condition:
            self.assertTrue(other_s._condition.wait_for(lambda: other_s.paused, timeout=10))
        self.assertEqual(len(other_s._completed), 0)
        self.stop_join(other_s, thread, errors)

    def test_unreachable_owned_loop_advances_in_frozen_order(self):
        s, _, _ = self.make()
        def calibration(profile, index, seed):
            if s._active["identity"] != s.ordered[0]: s.stop_gold()
            prepared = self.prepared(profile, index, seed)
            if index == 3: prepared = replace(prepared, runs=tuple(replace(r, annotated=frozenset()) for r in prepared.runs))
            return prepared
        s.run(lambda *a: self.fail("unreachable group executed gate"), lambda *a: True, calibration, lambda *a: {})
        report = self.assert_report(s)
        self.assertEqual(report["groups"][0]["status"], "unmeasured (unreachable target)")
        self.assertEqual([p["attempts"] for p in report["groups"][0]["populations"]], [0]*5)
        dispatch = [e["data"]["identity"] for e in report["events"] if e["kind"] == "dispatch"]
        self.assertEqual(dispatch, list(s.ordered[:2]))

    def criticality_fault(self, bad):
        s, _, _ = self.make()
        self.seed_active(s, complete=True)
        entered = threading.Event()
        calls = []
        def criticality(identity, j, key):
            calls.append(key)
            if len(calls) == 1:
                entered.set()
                if isinstance(bad, BaseException): raise bad
                return bad
            return {"key": key, "completed": True}
        def calibration(profile, index, seed): s.stop_gold(); return self.prepared(profile, index, seed)
        thread, errors = self.start_run(s, calibration=calibration, criticality=criticality)
        self.assertTrue(entered.wait(10))
        with s._condition: self.assertTrue(s._condition.wait_for(lambda: s.paused, timeout=10))
        self.assertEqual(len(calls), 1)
        s.resume(REVIEW)
        thread.join(15)
        self.assertFalse(thread.is_alive()); self.assertEqual(errors, [])
        self.assertEqual(calls[0], calls[1])
        self.assertEqual(len(calls), 5)
        self.assertTrue(any(e["kind"] == "pause" and e["data"].get("incident", {}).get("phase") == "criticality" for e in s.events))

    def test_criticality_exception_is_latched(self): self.criticality_fault(RuntimeError("criticality crash"))
    def test_criticality_malformed_is_latched(self): self.criticality_fault({"completed": "true"})
    def test_criticality_baseexception_is_latched(self): self.criticality_fault(KeyboardInterrupt())

    def test_replay_mid_attempt_and_after_gold_and_fixed_freeze(self):
        s, meter, root = self.make()
        lane = self.lane(s)
        with self.assertRaises(c.Stopped): lane.step(lambda i: True, lambda i: (_ for _ in ()).throw(RuntimeError("loss")))
        before = lane.report()
        s.close()
        recovered, _, _ = self.make(root=root, meter=meter)
        self.assertTrue(recovered.paused)
        self.assertEqual(recovered.freeze["oc_started_at"], 100000.)
        self.assertEqual(recovered.freeze["p_gate"], .5)
        recovered.resume(REVIEW)
        restored = self.lane(recovered)
        seen = []
        restored.step(lambda i: self.fail("recovery changed pending phase"), lambda i: seen.append(i) or True)
        self.assertEqual(seen, [0])
        self.assertEqual(restored.report()["attempts"], before["attempts"])
        recovered.stop_gold(); frozen = recovered.gold_report(); recovered.close()
        last, _, _ = self.make(root=root, meter=meter)
        self.assertEqual(last.gold_report(), frozen)
        self.assertEqual(last.run(None, None, None, None), frozen)
        last.close()
        with self.assertRaises(RuntimeError): self.make(root=root, meter=meter, oc_started_at=99999.)
        with self.assertRaises(RuntimeError): self.make(root=root, meter=meter, p_gate=.4)
        with self.assertRaises(RuntimeError): self.make(root=root, meter=meter, inventory_sha256="d"*64)

    def test_process_lock_and_torn_tail_preserves_evidence(self):
        s, meter, root = self.make()
        lane = self.lane(s)
        lane.step(lambda i: False, lambda i: True)
        with self.assertRaises(RuntimeError): self.make(root=root, meter=meter)
        s.close()
        journal = root / "x1-oc-controller/journal.jsonl"
        with journal.open("ab") as stream: stream.write(b'{"torn":')
        original = journal.read_bytes()
        recovered, _, _ = self.make(root=root, meter=meter)
        self.assertTrue(recovered.paused)
        self.assertTrue(journal.read_bytes().startswith(original))
        self.assertTrue(recovered.events[-1]["data"]["tail_sha256"])
        recovered.resume(REVIEW)
        recovered.stop_gold(); expected = recovered.gold_report(); recovered.close()
        again, _, _ = self.make(root=root, meter=meter)
        self.assertEqual(again.gold_report(), expected)

    def test_meter_origin_decrease_and_fault_latch_without_meter_read(self):
        s, meter, _ = self.make(oc_started_at=100000.-14*86400)
        lane = self.lane(s)
        meter.now += 86400
        with self.assertRaises(c.Stopped): lane.step(lambda i: True, lambda i: True)
        self.assertEqual(s.gold_report()["termination"]["cause"], "OC wall-day cap")
        for fault in (float("nan"), -1., .5):
            other, m, _ = self.make()
            lease = self.lane(other)
            m.core = 1.
            lease.step(lambda i: False, lambda i: True)
            m.core = fault
            with self.assertRaises(c.Stopped): lease.step(lambda i: True, lambda i: True)
            self.assertTrue(other.paused)
            self.assertEqual(other.events[-1]["data"]["kind"], "meter integrity")
        other, m, _ = self.make()
        lease = self.lane(other)
        m.now -= 1
        with self.assertRaises(c.Stopped): lease.step(lambda i: True, lambda i: True)
        self.assertTrue(other.paused)

    def test_private_readonly_state_and_completion_guard(self):
        s, _, _ = self.make()
        for field, value in (("terminated", False), ("paused", False), ("oc_started_at", 0), ("manifest", {})):
            with self.assertRaises(AttributeError): setattr(s, field, value)
        with self.assertRaises(TypeError): s.manifest["final_safety_OC"]["critical_j"][0] = 9
        self.seed_active(s, complete=True)
        with s._state_lock, self.assertRaises(RuntimeError): s._complete("measured")
        self.assertFalse(hasattr(s, "finish"))
        s.stop_gold()
        with self.assertRaises(AttributeError): s.terminated = False
        with self.assertRaises(c.Stopped): self.lane(s)
        report = s.gold_report(); report["groups"].clear()
        self.assertEqual(len(s.gold_report()["groups"]), 2016)

    def test_actual_process_loss_mid_safety_replays_same_attempt_and_seeds(self):
        with tempfile.TemporaryDirectory() as directory:
            code = """
import json, os, sys
from pathlib import Path
sys.path.insert(0, sys.argv[1])
import controller as c
root = Path(sys.argv[2])
c.shutil.disk_usage = lambda path: type("Disk", (), {"free": 10**16})()
s = c.SafetySchedule(p_gate=.5, inventory_sha256='c'*64, meter=lambda: 0.,
                     oc_started_at=100000., clock=lambda: 100000., run_root=root)
def prepare(profile, index, seed):
    d = json.loads(profile)[1]
    return c._kernel().prepare_population(seed=seed, run_count=1, sizes=(50,),
        prevalence=d['prevalence'], cap=d['cap_M'])
def crash(identity, index, attempt, seeds):
    (root/'child-seeds.json').write_text(json.dumps(seeds, sort_keys=True))
    os._exit(17)
s.run(lambda *a: True, crash, prepare, lambda *a: {})
"""
            result = subprocess.run([sys.executable, "-c", code, str(HERE), directory], timeout=30)
            self.assertEqual(result.returncode, 17)
            s, meter, _ = self.make(root=Path(directory))
            self.assertTrue(s.paused)
            s.resume(REVIEW)
            seen = []
            def trial(identity, index, attempt, seeds):
                seen.append((index, attempt, seeds))
                meter.core = 2880
                return True
            s.run(lambda *a: self.fail("pending trial reran gate"), trial,
                  lambda *a: self.fail("completed calibration repeated"), lambda *a: {})
            self.assertEqual(len(seen), 1)
            self.assertEqual(seen[0][:2], (0, 0))
            self.assertEqual(seen[0][2], json.loads((Path(directory)/"child-seeds.json").read_text()))
            first = self.assert_report(s)["groups"][0]["populations"][0]
            self.assertEqual((first["attempts"], first["executions"], first["eligible"]), (1, 1, 1))


    def test_unfinished_binding_refuses_safety_and_grid_complete(self):
        s, _, _ = self.make()
        lane = self.lane(s)
        def stop_calibration(*args):
            s.stop_gold()
            return self.prepared(*args)
        with self.assertRaises(RuntimeError): s.run(None, None, stop_calibration, None)
        with self.assertRaises(c.Stopped): lane.step(lambda i: True, lambda i: (_ for _ in ()).throw(RuntimeError()))
        s.resume(REVIEW)
        with self.assertRaises(RuntimeError): s.run(None, None, stop_calibration, None)
        other, _, _ = self.make()
        # Seed a completed-grid transition fixture; no claim to have run full OC.
        other._completed = [{"identity": identity, "status": "measured", "populations": [], "calibration": [], "criticality": {}} for identity in other.ordered]
        report = other.run(None, None, None, None)
        self.assertEqual(report["termination"]["cause"], "grid complete")
        self.assert_report(other)

    def test_gold_cap_and_completion_races_are_single_atomic_termination(self):
        for race in ("cap", "completion"):
            s, meter, _ = self.make()
            if race == "completion":
                self.seed_active(s, complete=True)
                with s._state_lock:
                    for j in s._critical_j:
                        key = c.criticality_key(s.ordered[0], j)
                        s._emit("criticality", {"key": key, "record": {"key": key, "completed": True}})
            barrier = threading.Barrier(3)
            errors = []
            def first(): barrier.wait(); s.stop_gold()
            def second():
                barrier.wait()
                try:
                    with s._state_lock:
                        if race == "cap": meter.core = 2880; s._boundary()
                        else: s._complete("measured")
                except c.Stopped: pass
                except BaseException as error: errors.append(error)
            threads = [threading.Thread(target=first), threading.Thread(target=second)]
            for thread in threads: thread.start()
            barrier.wait()
            for thread in threads: thread.join(10)
            self.assertEqual(errors, [])
            self.assert_report(s)


    def test_fault_arriving_after_reviewed_resume_latches_new_incident(self):
        for phase in ("gate", "trial"):
            s, _, _ = self.make()
            lane = self.lane(s)
            inside, release = threading.Event(), threading.Event()
            errors = []
            def late_fault(index):
                inside.set()
                release.wait(10)
                raise RuntimeError("fault observed after earlier resolution")
            def work():
                try:
                    lane.step(late_fault if phase == "gate" else lambda i: True,
                              late_fault if phase == "trial" else lambda i: True)
                except c.Stopped:
                    pass
                except BaseException as error:
                    errors.append(error)
            thread = threading.Thread(target=work)
            thread.start()
            self.assertTrue(inside.wait(10))
            s.pause(PAUSE)
            s.resume(REVIEW)
            release.set()
            thread.join(10)
            self.assertEqual(errors, [])
            self.assertTrue(s.paused)
            self.assertEqual(len(lane.report()["incidents"]), 1)
            s.resume(REVIEW)
            lane.step(lambda i: True, lambda i: True)
            self.assertEqual(lane.report()["attempts"], 1)


    def test_process_loss_between_passing_gate_and_trial_start(self):
        s, meter, root = self.make()
        lane = self.lane(s)
        original = s._boundary
        calls = []
        def boundary():
            calls.append(None)
            if len(calls) == 3:
                raise RuntimeError("synthetic process loss before durable trial start")
            original()
        with patch.object(s, "_boundary", side_effect=boundary):
            with self.assertRaises(RuntimeError):
                lane.step(lambda i: True, lambda i: self.fail("lost before trial"))
        self.assertEqual(lane.report()["pending"]["phase"], "ready")
        s.close()
        restored, _, _ = self.make(root=root, meter=meter)
        self.assertTrue(restored.paused)
        restored.resume(REVIEW)
        recovered = self.lane(restored)
        seen = []
        recovered.step(lambda i: seen.append(("gate", i)) or True,
                       lambda i: seen.append(("trial", i)) or True)
        self.assertEqual(seen, [("gate", 0), ("trial", 0)])
        self.assertEqual(recovered.report()["attempts"], 1)
        self.assertEqual(recovered.report()["executions"], 1)

    def test_gate_window_classifications_are_isolated(self):
        for action in ("gold", "cap", "pause", "pause-resume"):
            with self.subTest(action=action):
                s, meter, root = self.make()
                lane = self.lane(s)
                seen = []
                def gate(index):
                    seen.append(index)
                    if len(seen) == 1:
                        if action == "gold": s.stop_gold()
                        elif action == "cap": meter.core = 2880
                        else:
                            s.pause(PAUSE)
                            if action == "pause-resume": s.resume(REVIEW)
                    return True
                caught = []
                try: lane.step(gate, lambda i: self.fail("stopped gate started trial"))
                except BaseException as error: caught.append(error)
                self.assertEqual(len(caught), 1)
                self.assertIsInstance(caught[0], c.Stopped)
                self.assertEqual(lane.report()["gate_passes"], 0)
                self.assertEqual(lane.report()["pending"], {"phase": "gate", "index": 0})
                rows = [json.loads(line) for line in (root / "x1-oc-controller/journal.jsonl").read_bytes().splitlines()]
                if s.terminated:
                    position = next(i for i, row in enumerate(rows) if row["kind"] == "terminate")
                    self.assertFalse(any(row["kind"] == "population_state" for row in rows[position+1:]))
                else:
                    if s.paused: s.resume(REVIEW)
                    lane.step(gate, lambda i: True)
                    self.assertEqual(seen, [0, 0])
                s.stop_gold()

    def test_gate_baseexception_is_latched(self):
        self.fault_case("gate", KeyboardInterrupt())

    def test_calibration_baseexception_is_latched(self):
        s, _, _ = self.make()
        with s._state_lock: s._emit("dispatch", {"identity": s.ordered[0]})
        caught = []
        try: s._calibrate(lambda *args: (_ for _ in ()).throw(KeyboardInterrupt()))
        except BaseException as error: caught.append(error)
        self.assertEqual(len(caught), 1)
        self.assertIsInstance(caught[0], c.Stopped)
        self.assertTrue(s.paused)
        self.assertEqual(s.events[-1]["data"]["incident"]["phase"], "calibration")

    def test_safety_handle_guard(self):
        s, _, _ = self.make()
        with s._state_lock:
            s._emit("dispatch", {"identity": s.ordered[0]})
            key = s._new_population(s.ordered[0], 0, True, 0)
        caught = []
        try: c.BindingLane(s, key).step(lambda i: False, lambda i: True)
        except BaseException as error: caught.append(error)
        self.assertEqual(len(caught), 1)
        self.assertIsInstance(caught[0], RuntimeError)
        self.assertEqual(s._populations[key]["attempts"], 0)

    def test_paused_run_atomically_closes_binding_admission(self):
        s, _, _ = self.make()
        s.pause(PAUSE)
        waiting = threading.Event()
        original = s._wait_ready
        def wait():
            waiting.set()
            return original()
        s._wait_ready = wait
        def stop(*args):
            s.stop_gold()
            return self.prepared(*args)
        thread, errors = self.start_run(s, calibration=stop)
        self.assertTrue(waiting.wait(10))
        caught = []
        # Holding the condition makes the resumed lease win the dispatch race,
        # so this fails deterministically if admission only checks dispatch.
        with s._condition:
            s.resume(REVIEW)
            try: self.lane(s)
            except BaseException as error: caught.append(error)
        thread.join(15)
        self.assertFalse(thread.is_alive())
        self.assertEqual(errors, [])
        self.assertEqual(len(caught), 1)
        self.assertIsInstance(caught[0], ValueError)
        self.assertTrue(s._started_safety)
        self.assertFalse(any(not p["safety"] for p in s._populations.values()))

    def test_construction_base_vector_and_336_by_six(self):
        known = json.loads((HERE / "known_answers.json").read_bytes())
        s, _, _ = self.make()
        counts = {}
        for identity in s.ordered:
            base = s._profile(identity)
            counts[base] = counts.get(base, 0) + 1
        self.assertEqual(len(counts), 336)
        self.assertEqual(set(counts.values()), {6})
        base = known["construction_base"]
        vector = next(row for row in known["seeds"] if row["cell"] == base and row["phase"] == "population")
        seen = []
        with s._state_lock: s._emit("dispatch", {"identity": s.ordered[0]})
        def prepare(profile, index, seed):
            seen.append((profile, index, seed))
            return self.prepared(profile, index, seed)
        s._calibrate(prepare)
        self.assertEqual(seen[0], (base, 0, vector["seed64"]))

    def test_compact_record_counts_and_lost_gate_replay(self):
        s, meter, root = self.make()
        lane = self.lane(s)
        journal = root / "x1-oc-controller/journal.jsonl"
        before = journal.stat().st_size
        with patch.object(c.os, "fsync", wraps=c.os.fsync) as sync:
            lane.step(lambda i: False, lambda i: self.fail("gate rejected"))
            self.assertEqual(sync.call_count, 0)
            self.assertEqual(journal.stat().st_size, before)
            lane.step(lambda i: True, lambda i: True)
            self.assertEqual(sync.call_count, 2)
        durable = lane.report()
        lane.step(lambda i: False, lambda i: True)
        s.close()
        recovered, _, _ = self.make(root=root, meter=meter)
        self.assertTrue(recovered.paused)
        restored = recovered._population_report(lane._key)
        self.assertEqual(restored, durable)
        recovered.resume(REVIEW)
        restored_lane = self.lane(recovered)
        seen = []
        restored_lane.step(lambda i: seen.append(i) or False, lambda i: True)
        self.assertEqual(seen, [2])
        rows = [json.loads(line) for line in journal.read_bytes().splitlines()]
        self.assertFalse(any(row["kind"] == "meter" for row in rows))
        counter_rows = [row for row in rows if row["kind"] == "population_state"]
        self.assertTrue(all(set(row["data"]) == {"i", "d"} for row in counter_rows))
        self.assertTrue(all("meter" in row for row in rows))

    def test_disk_refusal_and_interpreter_provenance(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            with patch.object(c.shutil, "disk_usage", return_value=SimpleNamespace(free=0)):
                with self.assertRaises(RuntimeError):
                    c.SafetySchedule(p_gate=.5, inventory_sha256="c"*64, meter=lambda: 0,
                                     oc_started_at=100000., run_root=root)
            self.assertFalse((root / "x1-oc-controller/freeze.json").exists())
            refusal = json.loads((root / "x1-oc-controller/freeze-refused.json").read_bytes())
            self.assertEqual(refusal["journal_projection"]["binding_populations"], 90860)
            self.assertEqual(refusal["journal_projection"]["safety_populations"], 10080)
            self.assertGreater(refusal["disk_refusal"]["required_free_bytes"], 0)
        s, _, root = self.make()
        self.assertEqual(s.freeze["interpreter"]["python_optimize"], 0)
        self.assertEqual(s.freeze["interpreter"]["version"], "3.11.11")
        self.assertEqual(s.freeze["interpreter"]["executable_sha256"], c.digest(Path(sys.executable).resolve().read_bytes()))
        s.close()
        with patch.object(c.platform, "python_version", return_value="different"):
            with self.assertRaises(RuntimeError): self.make(root=root)
        for arguments, mode in ((["-O"], "optimization"), ([], "version"), ([], "implementation")):
            with tempfile.TemporaryDirectory() as directory:
                code = """
import sys
from pathlib import Path
from types import SimpleNamespace
sys.path.insert(0, sys.argv[1])
import controller as c
c.shutil.disk_usage = lambda path: SimpleNamespace(free=10**16)
if sys.argv[3] == "version": c.sys.version_info = (3, 11, 12)
if sys.argv[3] == "implementation": c.platform.python_implementation = lambda: "PyPy"
try:
    c.SafetySchedule(p_gate=.5, inventory_sha256='c'*64, meter=lambda: 0, oc_started_at=0, run_root=sys.argv[2])
except RuntimeError as error:
    print(str(error))
    sys.exit(0 if str(error) == "requires CPython 3.11.11 with python_optimize=0" and not (Path(sys.argv[2])/'x1-oc-controller/freeze.json').exists() else 2)
sys.exit(3)
"""
                result = subprocess.run([sys.executable, *arguments, "-c", code, str(HERE), directory, mode], capture_output=True, timeout=30)
                self.assertEqual(result.returncode, 0, result.stderr.decode())
                self.assertEqual(result.stdout.decode().strip(), "requires CPython 3.11.11 with python_optimize=0")

    def test_journal_failure_poison_blocks_other_lane_and_replays(self):
        for stage in ("write", "short", "newline", "flush", "fsync", "apply"):
            with self.subTest(stage=stage):
                s, meter, root = self.make()
                first, second = self.lane(s), self.lane(s, slot=1)
                entered, release = threading.Event(), threading.Event()
                errors = []
                def gate(index):
                    entered.set()
                    release.wait(10)
                    return True
                def worker():
                    try: second.step(gate, lambda i: True)
                    except BaseException as error: errors.append(error)
                thread = threading.Thread(target=worker)
                thread.start()
                self.assertTrue(entered.wait(10))
                original_file, original_apply = s._journal_file, s._apply
                expected = copy.deepcopy(s._populations[first._key])
                class FailingFile:
                    def write(self, raw):
                        if stage == "write":
                            original_file.write(raw[:len(raw)//2])
                            raise OSError("injected partial append")
                        if stage == "short": return original_file.write(raw[:len(raw)//2])
                        if stage == "newline": return original_file.write(raw[:-1])
                        return original_file.write(raw)
                    def flush(self):
                        if stage == "flush": raise OSError("injected flush")
                        return original_file.flush()
                    def fileno(self): return original_file.fileno()
                    def close(self): return original_file.close()
                s._journal_file = FailingFile()
                def failed_apply(event):
                    if stage == "apply" and "seq" in event: raise RuntimeError("injected reducer")
                    return original_apply(event)
                caught = []
                with patch.object(s, "_apply", side_effect=failed_apply), patch.object(c.os, "fsync", side_effect=OSError("injected fsync") if stage == "fsync" else c.os.fsync):
                    try: first.step(lambda i: True, lambda i: True)
                    except BaseException as error: caught.append(error)
                try:
                    journal = root / "x1-oc-controller/journal.jsonl"
                    frozen_bytes = journal.read_bytes()
                    self.assertEqual(len(caught), 1)
                    self.assertTrue(s._poisoned)
                    release.set()
                    thread.join(10)
                    self.assertFalse(thread.is_alive())
                    self.assertEqual(len(errors), 1)
                    self.assertIsInstance(errors[0], c.JournalPoisoned)
                    for operation in (lambda: s.pause(PAUSE), lambda: s.resume(REVIEW),
                                      lambda: first.step(lambda i: True, lambda i: True),
                                      lambda: s.run(None, None, None, None),
                                      lambda: s._emit("resume", REVIEW), s._boundary):
                        blocked = []
                        try:
                            with s._state_lock: operation()
                        except BaseException as error: blocked.append(error)
                        self.assertEqual(len(blocked), 1)
                        self.assertIsInstance(blocked[0], c.JournalPoisoned)
                    self.assertEqual(journal.read_bytes(), frozen_bytes)
                    s.close()
                    rows = []
                    for line in frozen_bytes.splitlines():
                        try: rows.append(json.loads(line))
                        except ValueError: pass
                    for row in rows:
                        expected.update(row.get("updates", {}).get("0", {}))
                        if row["kind"] == "population_state" and row["data"]["i"] == 0:
                            expected.update(row["data"]["d"])
                    for restart in range(2):
                        failures = []
                        try: recovered, _, _ = self.make(root=root, meter=meter)
                        except BaseException as error: failures.append(repr(error))
                        self.assertEqual(failures, [], f"{stage} restart {restart}")
                        self.assertTrue(recovered.paused)
                        self.assertEqual(recovered.events[-1]["data"]["kind"], "process loss")
                        if stage == "newline":
                            self.assertIsNone(recovered.events[-1]["data"]["tail_sha256"])
                        self.assertEqual(recovered._populations[first._key], expected)
                        recovered.close()
                finally:
                    # Assertion-killed mutants must not strand their live peer.
                    release.set()
                    thread.join(10)
                    s._journal_file = original_file
                    s.close()


if __name__ == "__main__":
    unittest.main()

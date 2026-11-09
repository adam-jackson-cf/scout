"""Synthetic-only, durable X1 OC controller. This draft grants no run authority."""
from dataclasses import dataclass
import copy
import fcntl
import hashlib
import hmac
import importlib.util
import itertools
import json
import math
import os
from pathlib import Path
import platform
import shutil
import sys
import threading
import time
from types import MappingProxyType

HERE = Path(__file__).resolve().parent
EXPERIMENT = HERE.parents[1]
MANIFEST_SHA256 = "1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d"
KERNEL_SHA256 = "ed670a6bf8076d791c06cfc2426aff03a0a2f9b1730982cf3fdfb2ec13bc49a2"
RANKING_SHA256 = "8bc2e7cbbc5da6ba47aeff3effb5dd04d5e5ecc9c1e01051647900095770a780"
ORDER_SHA256 = "04aa2184f79cb3795f60305a6b8fa31687d5acd7069846e78b6acf1b1f4dd4ff"
POLICY_SHA256 = "9b6de56992b00129edcb09d8c0b1e95e7f3bfb538dc3fea02c0529e94886bfd5"
POLICY = json.loads((HERE / "policy.json").read_bytes())
PHASES = frozenset(("population", "audit", "bootstrap", "missing-mask", "canonical-safety", "canonical-criticality", "sampling-gate-estimate"))


def canonical(value):
    return json.dumps(value, sort_keys=True, separators=(",", ":"), allow_nan=False).encode()


def digest(data):
    return hashlib.sha256(data).hexdigest()


def _kernel(path=None):
    path = EXPERIMENT / "validation/fixed_population.py" if path is None else Path(path)
    if digest(path.read_bytes()) != KERNEL_SHA256:
        raise RuntimeError("signed population identity source changed")
    ranking = EXPERIMENT / "validation/ranking.py"
    if digest(ranking.read_bytes()) != RANKING_SHA256:
        raise RuntimeError("signed calibration dependency changed")
    if "ranking" not in sys.modules:
        spec = importlib.util.spec_from_file_location("ranking", ranking)
        module = importlib.util.module_from_spec(spec)
        sys.modules["ranking"] = module
        spec.loader.exec_module(module)
    name = "x1_signed_controller_population"
    if name not in sys.modules:
        spec = importlib.util.spec_from_file_location(name, path)
        module = importlib.util.module_from_spec(spec)
        sys.modules[name] = module
        spec.loader.exec_module(module)
    return sys.modules[name]


population_identity = _kernel().population_identity
criticality_key = _kernel().criticality_key


def seed(cell, population, attempt, phase):
    if phase not in PHASES or any(type(i) is not int or i < 0 for i in (population, attempt)):
        raise ValueError("invalid phase or index")
    if phase == "population" and attempt != 0:
        raise ValueError("population phase uses attempt zero")
    if isinstance(cell, str):
        cell_bytes = cell.encode("utf-8")
        if canonical(json.loads(cell)) != cell_bytes:
            raise ValueError("cell string is not canonical JSON")
    else:
        cell_bytes = canonical(cell)
    message = b"\0".join((cell_bytes, str(population).encode(), str(attempt).encode(), phase.encode()))
    return int.from_bytes(hmac.digest(b"169-x1-oc-v4", message, "sha256")[:8], "big")


def safety_groups(manifest):
    specification = manifest["final_safety_OC"]
    dimensions = specification["dimensions"]
    for values in itertools.product(*dimensions.values()):
        for kappa in specification["unresolved_kappa"]:
            yield population_identity("safety-interval-group", {**dict(zip(dimensions, values)), "unresolved_kappa": kappa})


def wilson(k, n):
    if type(k) is not int or type(n) is not int or not 0 <= k <= n:
        raise ValueError("invalid Wilson counts")
    if not n:
        return None
    z = 1.959963984540054
    p = k/n
    centre = (p+z*z/(2*n))/(1+z*z/n)
    half = z*math.sqrt(p*(1-p)/n+z*z/(4*n*n))/(1+z*z/n)
    return [centre-half, centre+half]


def immutable(value):
    if isinstance(value, dict):
        return MappingProxyType({k: immutable(v) for k, v in value.items()})
    if isinstance(value, list):
        return tuple(immutable(v) for v in value)
    return value


def _sha(value):
    return isinstance(value, str) and len(value) == 64 and all(c in "0123456789abcdef" for c in value)


class Stopped(RuntimeError):
    """No new work may start until reviewed resume, or ever after termination."""


class JournalPoisoned(RuntimeError):
    """Uncertain append; only a new process/replay may recover this writer."""


def journal_projection(p_gate):
    # Conservative storage reservation, not a scientific trial-count reduction.
    populations = 90860 + 10080
    attempts = math.ceil(4000 / p_gate)
    records = populations * (2 * 4000 + 16)
    return {"binding_populations": 90860, "safety_populations": 10080,
            "p_gate": p_gate, "attempt_cap_per_population": attempts,
            "maximum_executions_per_population": 4000,
            "reserved_records": records, "reserved_bytes_per_record": 2048,
            "required_free_bytes": records * 2048,
            "basis": "upper execution cap; two durable records per execution plus 16 control records per population; gate deltas coalesced; conservative 2048-byte record reservation, synthetic benchmark reports observed sizes separately"}


@dataclass(frozen=True)
class BindingLane:
    _owner: object
    _key: str

    def step(self, gate, trial):
        return self._owner._attempt(self._key, gate, trial)

    def report(self):
        with self._owner._state_lock:
            return self._owner._population_report(self._key)

    @property
    def complete(self):
        with self._owner._state_lock:
            return self._owner._done(self._key)

    @property
    def exhausted(self):
        with self._owner._state_lock:
            return self._owner._exhausted(self._key)


class SafetySchedule:
    """Eight binding slots, serial owned safety loop, append-before-apply state."""

    def __init__(self, *, p_gate, inventory_sha256, meter, oc_started_at,
                 run_root=None, clock=time.time, manifest_bytes=None, kernel_path=None):
        if platform.python_implementation() != "CPython" or sys.version_info[:3] != (3, 11, 11) or sys.flags.optimize:
            raise RuntimeError("requires CPython 3.11.11 with python_optimize=0")
        self._state_lock = threading.RLock()
        self._condition = threading.Condition(self._state_lock)
        self._run_lock = threading.Lock()
        self._kernel = _kernel(kernel_path)
        raw = (EXPERIMENT / "oc-manifest.json").read_bytes() if manifest_bytes is None else manifest_bytes
        if digest(raw) != MANIFEST_SHA256:
            raise ValueError("signed OC manifest bytes changed")
        self._manifest = immutable(json.loads(raw))
        identities = tuple(safety_groups(json.loads(raw)))
        self._ordered = tuple(sorted(identities, key=lambda s: (hmac.digest(b"169-x1-safety-order", s.encode(), "sha256"), s.encode())))
        if len(set(identities)) != 2016 or digest(canonical(self._ordered)) != ORDER_SHA256 or digest(canonical(POLICY)) != POLICY_SHA256:
            raise RuntimeError("frozen safety order or policy mismatch")
        self._policy = immutable(copy.deepcopy(POLICY))
        self._critical_j = tuple(self._manifest["final_safety_OC"]["critical_j"])
        if type(p_gate) not in (int, float) or not math.isfinite(p_gate) or not 0 < p_gate <= 1 or not _sha(inventory_sha256):
            raise ValueError("unmeasured gate probability or inventory digest")
        if type(oc_started_at) not in (int, float) or not math.isfinite(oc_started_at):
            raise ValueError("invalid full-OC origin")
        self._p_gate, self._oc_started_at = p_gate, oc_started_at
        self._meter, self._clock = meter, clock
        self._last_core = self._last_wall = 0
        self._epoch = 0
        self._paused = self._terminated = self._started_safety = False
        self._final = self._active = None
        self._completed, self._events = [], []
        self._populations, self._profiles, self._lane_locks = {}, {}, {}
        self._slots, self._busy_slots = {}, set()
        self._sequence, self._previous = 0, "0" * 64
        self._poisoned = False
        self._pending_deltas = {}
        self._population_ids, self._population_keys = {}, {}
        self._closed = False
        directory = (HERE if run_root is None else Path(run_root)) / "x1-oc-controller"
        directory.mkdir(parents=True, exist_ok=True)
        self._directory = directory
        self._lock_file = (directory / "owner.lock").open("a+b")
        try:
            fcntl.flock(self._lock_file, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BaseException:
            self._lock_file.close()
            raise RuntimeError("another controller owns this run directory") from None
        self._journal_file = None
        try:
            sources = (Path(__file__).resolve(), HERE / "policy.json", EXPERIMENT / "validation/fixed_population.py",
                       EXPERIMENT / "validation/ranking.py", EXPERIMENT / "oc-manifest.json")
            expected = {"ordered_groups_sha256": ORDER_SHA256, "policy_sha256": POLICY_SHA256,
                        "oc_started_at": oc_started_at, "p_gate": p_gate, "inventory_sha256": inventory_sha256,
                        "synthetic_only": True, "authority": "none; scoped review required",
                        "source_sha256": {str(p.relative_to(EXPERIMENT)): digest(p.read_bytes()) for p in sources}}
            executable = Path(sys.executable).resolve()
            expected["interpreter"] = {"executable": str(executable),
                                       "executable_sha256": digest(executable.read_bytes()),
                                       "version": platform.python_version(),
                                       "implementation": platform.python_implementation(),
                                       "python_optimize": sys.flags.optimize}
            expected["journal_projection"] = journal_projection(p_gate)
            freeze_path = directory / "freeze.json"
            recovered = freeze_path.exists()
            self._tail_sha = None
            self._tail_needs_newline = False
            if recovered:
                self._freeze = json.loads(freeze_path.read_bytes())
                if {k: self._freeze.get(k) for k in expected} != expected:
                    raise RuntimeError("existing freeze differs; second freeze refused")
                self._replay(directory / "journal.jsonl")
            else:
                if (directory / "journal.jsonl").exists():
                    raise RuntimeError("journal exists without freeze")
                free = shutil.disk_usage(directory).free
                refusal = {"required_free_bytes": expected["journal_projection"]["required_free_bytes"],
                           "observed_free_bytes": free, "timestamp": self._clock()}
                if free < refusal["required_free_bytes"]:
                    with (directory / "freeze-refused.json").open("xb") as stream:
                        stream.write(canonical({**expected, "disk_refusal": refusal}) + b"\n")
                        stream.flush()
                        os.fsync(stream.fileno())
                    raise RuntimeError("insufficient journal free space; refusal recorded without freeze")
                self._freeze = {**expected, "created_at": self._clock(), "disk_admission": refusal}
                with freeze_path.open("xb") as stream:
                    stream.write(canonical(self._freeze) + b"\n")
                    stream.flush()
                    os.fsync(stream.fileno())
            self._journal_file = (directory / "journal.jsonl").open("ab", buffering=0)
            fd = os.open(directory, os.O_RDONLY)
            try:
                os.fsync(fd)
            finally:
                os.close(fd)
            if recovered and not self._terminated:
                if self._tail_needs_newline:
                    self._journal_file.write(b"\n")
                    self._journal_file.flush()
                    os.fsync(self._journal_file.fileno())
                with self._state_lock:
                    self._emit("pause", {"kind": "process loss", "evidence_sha256": self._tail_sha or self._previous,
                                         "tail_sha256": self._tail_sha})
        except BaseException:
            self.close()
            raise

    @property
    def manifest(self): return self._manifest
    @property
    def ordered(self): return self._ordered
    @property
    def policy(self): return self._policy
    @property
    def freeze(self): return immutable(copy.deepcopy(self._freeze))
    @property
    def oc_started_at(self): return self._oc_started_at
    @property
    def p_gate(self): return self._p_gate
    @property
    def paused(self): return self._paused
    @property
    def terminated(self): return self._terminated
    @property
    def events(self):
        with self._state_lock:
            return copy.deepcopy(self._events)

    def close(self):
        if not self._closed:
            if (self._run_lock.locked() or self._busy_slots) and not self._terminated:
                raise RuntimeError("cannot close a live controller")
            if self._journal_file is not None:
                self._journal_file.close()
            self._lock_file.close()
            self._closed = True

    def _replay(self, path):
        if not path.exists():
            return
        pending_tail = None
        with path.open("rb") as stream:
            for raw in stream:
                try:
                    event = json.loads(raw)
                except (ValueError, UnicodeDecodeError):
                    event = None
                if event is None or (not raw.endswith(b"\n") and
                                     (event["seq"] != self._sequence or event["previous_sha256"] != self._previous)):
                    if pending_tail is not None:
                        raise RuntimeError("multiple corrupt journal records")
                    pending_tail = digest(raw.removesuffix(b"\n"))
                    self._tail_needs_newline = not raw.endswith(b"\n")
                    continue
                if pending_tail is not None:
                    if event.get("kind") != "pause" or event.get("data", {}).get("tail_sha256") != pending_tail:
                        raise RuntimeError("unreviewed corrupt journal interior")
                    pending_tail = None
                    self._tail_needs_newline = False
                if event["seq"] != self._sequence or event["previous_sha256"] != self._previous:
                    raise RuntimeError("journal chain mismatch")
                self._apply(event)
                self._sequence += 1
                self._tail_needs_newline = not raw.endswith(b"\n")
                self._previous = digest(raw if raw.endswith(b"\n") else raw + b"\n")
        self._tail_sha = pending_tail

    def _check_writer(self):
        if self._poisoned:
            raise JournalPoisoned("journal writer poisoned; restart and replay required")
        if self._closed:
            raise RuntimeError("controller closed")

    def _emit(self, kind, data, *, durable=True):
        # All callers hold the state lock. Gate deltas are memory-only until
        # the next linearization point; no uncertain append is ever retried.
        self._check_writer()
        if not durable:
            if kind != "population_state":
                raise RuntimeError("only counter deltas may be deferred")
            try:
                self._pending_deltas.setdefault(str(data["i"]), {}).update(copy.deepcopy(data["d"]))
                self._apply({"kind": kind, "data": data})
            except BaseException:
                self._poisoned = True
                self._condition.notify_all()
                raise
            return
        if kind in ("pause", "terminate") and self._active is not None:
            if self._active.get("current_lane") is not None:
                self._emit("lane_stop", {"identity": self._active["identity"],
                                         "population": self._active["current_lane"], "reason": kind})
            if self._active["pending_phase"] is not None:
                self._emit("phase_stop", {"identity": self._active["identity"],
                                          "phase": self._active["pending_phase"]["phase"], "reason": kind})
        try:
            timestamp = self._clock()
            if type(timestamp) not in (int, float) or not math.isfinite(timestamp):
                timestamp = None
        except BaseException:
            timestamp = None
        event = {"seq": self._sequence, "previous_sha256": self._previous,
                 "timestamp": timestamp, "kind": kind, "data": data,
                 "meter": [self._last_core, self._last_wall]}
        if self._pending_deltas:
            event["updates"] = copy.deepcopy(self._pending_deltas)
        raw = canonical(event) + b"\n"
        self._poisoned = True
        try:
            if self._journal_file.write(raw) != len(raw):
                raise OSError("short journal write")
            self._journal_file.flush()
            os.fsync(self._journal_file.fileno())
            self._apply(event)
            self._sequence += 1
            self._previous = digest(raw)
            self._pending_deltas.clear()
        except BaseException:
            self._condition.notify_all()
            raise
        self._poisoned = False

    def _apply(self, event):
        kind, data = event["kind"], event["data"]
        for index, delta in event.get("updates", {}).items():
            self._populations[self._population_keys[int(index)]].update(copy.deepcopy(delta))
        if "meter" in event:
            self._last_core, self._last_wall = event["meter"]
        # Counter commits stay on disk, not millions of duplicated RAM records.
        if kind not in ("population", "population_state", "meter"):
            self._events.append(copy.deepcopy(event))
        if kind == "meter":
            self._last_core, self._last_wall = data["core"], data["wall"]
        elif kind == "population":
            self._populations[data["key"]] = copy.deepcopy(data["population"])
            index = data.get("i", len(self._population_ids))
            self._population_ids[data["key"]] = index
            self._population_keys[index] = data["key"]
            self._lane_locks[data["key"]] = threading.Lock()
            if not data["population"]["safety"]:
                self._slots[data["population"]["slot"]] = data["key"]
            elif self._active is not None:
                self._active["populations"].append(data["key"])
        elif kind == "population_state":
            self._populations[self._population_keys[data["i"]]].update(copy.deepcopy(data["d"]))
        elif kind == "pause":
            self._paused = True
            self._epoch += 1
        elif kind == "resume":
            self._paused = False
            self._epoch += 1
        elif kind == "dispatch":
            self._started_safety = True
            self._active = {"identity": data["identity"], "populations": [], "calibration": [],
                            "criticality": {}, "pending_phase": None, "current_lane": None}
        elif kind == "calibration":
            self._active["calibration"].append(copy.deepcopy(data["record"]))
            self._active["pending_phase"] = None
        elif kind == "profile":
            self._profiles[data["profile"]] = copy.deepcopy(data["records"])
        elif kind == "phase_start":
            self._active["pending_phase"] = copy.deepcopy(data)
        elif kind == "lane_start":
            self._active["current_lane"] = data["population"]
        elif kind == "lane_stop":
            self._active["current_lane"] = None
        elif kind == "criticality":
            self._active["criticality"][data["key"]] = copy.deepcopy(data["record"])
            self._active["pending_phase"] = None
        elif kind == "complete":
            self._completed.append(copy.deepcopy(data["snapshot"]))
            for key in self._active["populations"]:
                del self._populations[key]
                del self._lane_locks[key]
            self._active = None
        elif kind == "terminate":
            self._terminated = True
            self._epoch += 1
            self._final = self._build_report(data["cause"], event["timestamp"])

    def _boundary(self):
        self._check_writer()
        if self._terminated:
            raise Stopped("OC terminated")
        if self._paused:
            raise Stopped("OC paused for integrity review")
        try:
            core, wall = self._meter(), (self._clock() - self._oc_started_at) / 86400
            if any(type(x) not in (int, float) or not math.isfinite(x) or x < 0 for x in (core, wall)) or core < self._last_core or wall < self._last_wall:
                raise ValueError("invalid/nonmonotone OC meter")
        except BaseException as error:
            self._emit("pause", {"kind": "meter integrity", "error": type(error).__name__, "evidence_sha256": digest(str(error).encode())})
            raise Stopped("meter integrity pause") from error
        self._last_core, self._last_wall = core, wall
        if core >= 2880 or wall >= 15:
            self._terminate("OC core-hour cap" if core >= 2880 else "OC wall-day cap")
            raise Stopped("OC cap reached")

    def _new_population(self, identity, index, safety, slot):
        key = identity + "\0" + str(index)
        if key in self._populations:
            return key
        population = {"identity": identity, "index": index, "safety": safety, "slot": slot,
                      "attempts": 0, "executions": 0, "eligible": 0, "inadequate": 0,
                      "gate_failures": 0, "gate_passes": 0, "gate_faults": 0,
                      "pending": None, "incidents": []}
        self._emit("population", {"key": key, "population": population})
        return key

    def binding_population(self, identity, index, *, slot):
        with self._state_lock:
            self._boundary()
            if self._started_safety or self._run_lock.locked() or type(index) is not int or index not in range(5) or type(slot) is not int or slot not in range(8):
                raise ValueError("invalid binding lease")
            seed(identity, index, 0, "population")
            old = self._slots.get(slot)
            key = identity + "\0" + str(index)
            if old is not None and old != key and not (self._done(old) or self._exhausted(old)):
                raise RuntimeError("binding slot already leased")
            key = self._new_population(identity, index, False, slot)
            if self._populations[key]["slot"] != slot:
                raise RuntimeError("population already bound to another slot")
            return BindingLane(self, key)

    def _done(self, key):
        p = self._populations[key]
        return p["pending"] is None and (p["executions"] if p["safety"] else p["eligible"]) == 2000

    def _exhausted(self, key):
        p = self._populations[key]
        return p["pending"] is None and not self._done(key) and (p["attempts"] >= math.ceil(4000/self._p_gate) or p["executions"] >= 4000)

    def _save_population(self, key, p, *, durable=True):
        delta = {k: copy.deepcopy(v) for k, v in p.items() if self._populations[key][k] != v}
        self._emit("population_state", {"i": self._population_ids[key], "d": delta}, durable=durable)

    def _fault(self, phase, error, *, key=None, index=None):
        incident = {"phase": phase, "index": index, "error": type(error).__name__, "evidence_sha256": digest(str(error).encode()), "resolved": False}
        if key is not None:
            p = copy.deepcopy(self._populations[key])
            p["incidents"].append(incident)
            if phase == "gate":
                p["gate_faults"] += 1
            self._save_population(key, p)
        self._emit("pause", {"kind": "execution integrity", "key": key, "incident": incident})
        self._condition.notify_all()

    def _attempt(self, key, gate, trial):
        lane_lock = self._lane_locks[key]
        if not lane_lock.acquire(blocking=False):
            raise RuntimeError("parallel/reentrant population lane refused")
        slot = self._populations[key]["slot"]
        try:
            with self._state_lock:
                self._boundary()
                p = self._populations[key]
                if p["safety"] and not self._run_lock.locked():
                    raise RuntimeError("safety work belongs to run loop")
                if slot in self._busy_slots:
                    raise RuntimeError("worker slot busy")
                if self._done(key) or self._exhausted(key):
                    raise RuntimeError("population stopped")
                self._busy_slots.add(slot)
                epoch = self._epoch
                if p["pending"] is None:
                    p = copy.deepcopy(p)
                    p["pending"] = {"phase": "gate", "index": p["attempts"]}
                    p["attempts"] += 1
                    self._save_population(key, p, durable=False)
                index, phase = p["pending"]["index"], p["pending"]["phase"]
            if phase == "gate":
                try:
                    passes = gate(index)
                    if type(passes) is not bool:
                        raise ValueError("gate must report a bool")
                except BaseException as error:
                    with self._state_lock:
                        if not self._terminated:
                            self._fault("gate", error, key=key, index=index)
                    raise Stopped("gate integrity pause") from error
                with self._state_lock:
                    self._check_writer()
                    if self._terminated or self._paused or epoch != self._epoch:
                        raise Stopped("gate result remains unclassified after stop/pause")
                    self._boundary()
                    p = copy.deepcopy(self._populations[key])
                    p["pending"] = {"phase": "ready", "index": index} if passes else None
                    p["gate_passes" if passes else "gate_failures"] += 1
                    for incident in p["incidents"]:
                        if incident["phase"] == "gate":
                            incident["resolved"] = True
                    self._save_population(key, p, durable=False)
                    self._boundary()
                    if not passes:
                        return
            with self._state_lock:
                self._boundary()
                p = copy.deepcopy(self._populations[key])
                if p["pending"]["phase"] == "ready":
                    # This durable transition is the trial-start linearization point.
                    p["executions"] += 1
                    p["pending"] = {"phase": "trial", "index": index}
                    self._save_population(key, p)
            try:
                adequate = trial(index)
                if type(adequate) is not bool:
                    raise ValueError("trial must report a bool")
            except BaseException as error:
                with self._state_lock:
                    if not self._terminated:
                        self._fault("trial", error, key=key, index=index)
                raise Stopped("trial integrity pause") from error
            with self._state_lock:
                self._check_writer()
                if self._terminated or self._paused or epoch != self._epoch:
                    raise Stopped("trial result remains unclassified after stop/pause")
                p = copy.deepcopy(self._populations[key])
                p["pending"] = None
                p["eligible" if adequate else "inadequate"] += 1
                for incident in p["incidents"]:
                    incident["resolved"] = True
                self._save_population(key, p)
                # A newly observed cap retains the completed trial result.
                self._boundary()
        finally:
            with self._state_lock:
                self._busy_slots.discard(slot)
            lane_lock.release()

    def pause(self, evidence):
        with self._state_lock:
            self._check_writer()
            if self._terminated:
                raise Stopped("OC terminated")
            if not isinstance(evidence, dict) or evidence.get("kind") not in ("execution integrity", "closure mismatch", "host integrity", "process loss") or not _sha(evidence.get("evidence_sha256")):
                raise ValueError("section-11 evidence required")
            self._emit("pause", copy.deepcopy(evidence))
            self._condition.notify_all()

    def resume(self, evidence):
        with self._state_lock:
            self._check_writer()
            if self._terminated or not self._paused:
                raise Stopped("not a resumable integrity pause")
            if not isinstance(evidence, dict) or evidence.get("reviewed") is not True or not _sha(evidence.get("evidence_sha256")):
                raise ValueError("reviewed resolution evidence required")
            self._emit("resume", copy.deepcopy(evidence))
            self._condition.notify_all()

    def _wait_ready(self):
        with self._condition:
            while self._paused and not self._terminated and not self._poisoned:
                self._condition.wait()
            self._check_writer()
            return not self._terminated

    def _profile(self, identity):
        dimensions = dict(json.loads(identity)[1])
        dimensions.pop("unresolved_kappa")
        return population_identity("safety-profile", dimensions)

    def _calibrate(self, calibration):
        identity = self._active["identity"]
        profile = self._profile(identity)
        cached = self._profiles.get(profile)
        for index in range(len(self._active["calibration"]), 5):
            with self._state_lock:
                self._boundary()
                self._emit("phase_start", {"phase": "calibration", "population": index, "identity": identity})
                epoch = self._epoch
            try:
                if cached is None:
                    prepared = calibration(profile, index, seed(profile, index, 0, "population"))
                    result = self._kernel.calibrate_population(prepared, json.loads(profile)[1]["realised_gain"])
                    record = {"profile": profile, "population_index": index, "target": result["target"], "realised": result["realised"], "excess": result["excess"], "latent_gain": result["latent_gain"], "reachable": result["reachable"], "kernel_sha256": KERNEL_SHA256}
                else:
                    record = copy.deepcopy(cached[index])
                    record.pop("identity", None)
                    record.pop("record_sha256", None)
                record["identity"] = identity
                record["record_sha256"] = digest(canonical(record))
            except BaseException as error:
                with self._state_lock:
                    if not self._terminated:
                        self._fault("calibration", error, index=index)
                raise Stopped("calibration integrity pause") from error
            with self._state_lock:
                if self._terminated or self._paused or epoch != self._epoch:
                    raise Stopped("calibration interrupted")
                self._emit("calibration", {"record": record})
                self._boundary()
        with self._state_lock:
            if cached is None:
                self._emit("profile", {"profile": profile, "records": copy.deepcopy(self._active["calibration"])})
            # Every kappa reuses the first profile's five signed-kernel records.
            expected = self._profiles[profile]
            for a, b in zip(expected, self._active["calibration"]):
                fields = ("target", "realised", "excess", "latent_gain", "reachable", "kernel_sha256", "population_index", "profile")
                if any(a[f] != b[f] for f in fields):
                    raise RuntimeError("profile calibration conflicts across kappa")

    def _run_criticality(self, work):
        identity = self._active["identity"]
        for j in self._critical_j:
            key = criticality_key(identity, j)
            if key in self._active["criticality"]:
                continue
            with self._state_lock:
                self._boundary()
                self._emit("phase_start", {"phase": "criticality", "identity": identity, "j": j, "key": key})
                epoch = self._epoch
            try:
                result = work(identity, j, key)
                if not isinstance(result, dict) or result.get("key") != key or result.get("completed") is not True:
                    raise ValueError("criticality must return completed exact key")
                canonical(result)
            except BaseException as error:
                with self._state_lock:
                    if not self._terminated:
                        self._fault("criticality", error, index=j)
                raise Stopped("criticality integrity pause") from error
            with self._state_lock:
                if self._terminated or self._paused or epoch != self._epoch:
                    raise Stopped("criticality interrupted")
                self._emit("criticality", {"key": key, "record": copy.deepcopy(result)})
                self._boundary()

    def run(self, gate, trial, calibration, criticality):
        """Own safety pacing. calibration returns prepared data, never a verdict."""
        if not self._run_lock.acquire(blocking=False):
            raise RuntimeError("controller run already active")
        try:
            with self._state_lock:
                self._check_writer()
                if self._terminated:
                    return self.gold_report()
                if any(not (self._done(k) or self._exhausted(k)) for k, p in self._populations.items() if not p["safety"]):
                    raise RuntimeError("binding populations unfinished or faulted")
            while self._wait_ready():
                try:
                    with self._state_lock:
                        self._boundary()
                        if self._active is None:
                            if len(self._completed) == len(self._ordered):
                                self._terminate("grid complete")
                                break
                            if not self._started_safety and any(not (self._done(k) or self._exhausted(k)) for k, p in self._populations.items() if not p["safety"]):
                                raise RuntimeError("binding populations unfinished or faulted")
                            self._emit("dispatch", {"identity": self._ordered[len(self._completed)]})
                            for index in range(5):
                                self._new_population(self._active["identity"], index, True, index)
                    self._calibrate(calibration)
                    with self._state_lock:
                        if any(not r["reachable"] for r in self._active["calibration"]):
                            self._complete("unmeasured (unreachable target)")
                            continue
                        identity = self._active["identity"]
                    for index in range(5):
                        with self._state_lock:
                            key = self._new_population(identity, index, True, index)
                            self._emit("lane_start", {"identity": identity, "population": index, "slot": index})
                        try:
                            while not self._done(key) and not self._exhausted(key):
                                self._attempt(key,
                                    lambda attempt: gate(identity, index, attempt, seed(identity, index, attempt, "audit")),
                                    lambda attempt: trial(identity, index, attempt, {phase: seed(identity, index, attempt, phase) for phase in ("audit", "bootstrap", "missing-mask", "canonical-safety")}))
                        finally:
                            with self._state_lock:
                                if not self._terminated:
                                    self._emit("lane_stop", {"identity": identity, "population": index, "slot": index,
                                                             "paused": self._paused})
                    self._run_criticality(criticality)
                    with self._state_lock:
                        self._complete("measured" if all(self._done(k) for k in self._active["populations"]) else "unmeasured; cell fails")
                except Stopped:
                    continue
            return self.gold_report()
        finally:
            self._run_lock.release()

    def _population_report(self, key):
        p = copy.deepcopy(self._populations[key])
        pending = p["pending"]
        crashed = int(pending is not None and pending["phase"] == "trial")
        if p["executions"] != p["eligible"] + p["inadequate"] + crashed:
            raise RuntimeError("execution classification invariant violated")
        classified_gates = p["gate_failures"] + p["gate_passes"]
        if p["attempts"] != classified_gates + int(pending is not None and pending["phase"] == "gate"):
            raise RuntimeError("gate classification invariant violated")
        return {**p, "crashed": crashed, "attempt_cap": math.ceil(4000/self._p_gate),
                "gate_failure_Wilson_MC95": wilson(p["gate_failures"], classified_gates),
                "conditional_inadequacy_Wilson_MC95": None if crashed else wilson(p["inadequate"], p["executions"])}

    def _snapshot(self, status):
        active = copy.deepcopy(self._active)
        active["status"] = status
        active["populations"] = [self._population_report(k) for k in active["populations"]]
        return active

    def _complete(self, status):
        self._boundary()
        if status == "measured" and len(self._active["criticality"]) != 4:
            raise RuntimeError("all criticality phases required")
        self._emit("complete", {"snapshot": self._snapshot(status)})

    def _build_report(self, cause, timestamp):
        groups = copy.deepcopy(self._completed)
        if self._active is not None:
            keys = [k for k, p in self._populations.items() if p["safety"] and p["identity"] == self._active["identity"]]
            exhausted = any(self._exhausted(k) for k in keys)
            groups.append(self._snapshot("unmeasured; cell fails" if exhausted else "not completed by gold"))
        for identity in self._ordered[len(groups):]:
            groups.append({"identity": identity, "status": "not completed by gold", "populations": [], "calibration": [], "criticality": {}})
        return {"freeze": copy.deepcopy(self._freeze), "termination": {"cause": cause, "timestamp": timestamp},
                "groups": groups, "binding": [self._population_report(k) for k, p in self._populations.items() if not p["safety"]],
                "events": copy.deepcopy(self._events)}

    def _terminate(self, cause):
        if not self._terminated:
            self._emit("terminate", {"cause": cause})
            self._condition.notify_all()

    def stop_gold(self):
        with self._state_lock:
            self._terminate("gold")

    def stop_integrity(self, evidence):
        with self._state_lock:
            if not isinstance(evidence, dict) or not _sha(evidence.get("evidence_sha256")):
                raise ValueError("integrity evidence required")
            if not self._terminated:
                self._emit("integrity_stop", copy.deepcopy(evidence))
                self._terminate("integrity stop")

    def gold_report(self):
        with self._state_lock:
            if not self._terminated:
                raise RuntimeError("no terminated report")
            return copy.deepcopy(self._final)

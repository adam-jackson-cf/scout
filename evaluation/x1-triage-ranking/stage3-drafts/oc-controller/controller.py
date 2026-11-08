"""Synthetic-only stage-3 counter/safety state machine; no run authority."""
from dataclasses import dataclass, field
from datetime import datetime, timezone
import copy
import hashlib
import hmac
import importlib.util
import itertools
import json
import math
from pathlib import Path
import sys
import time
import threading
from types import MappingProxyType

HERE = Path(__file__).resolve().parent
EXPERIMENT = HERE.parents[1]
MANIFEST_SHA256 = "1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d"
KERNEL_SHA256 = "ed670a6bf8076d791c06cfc2426aff03a0a2f9b1730982cf3fdfb2ec13bc49a2"
ORDER_SHA256 = "04aa2184f79cb3795f60305a6b8fa31687d5acd7069846e78b6acf1b1f4dd4ff"
POLICY_SHA256 = "e2c8c9ec1da43b4eeface3cf3c52c711ecad8cef11be97c00defe51fc12353ca"
POLICY = {'checks': 'before every sampling attempt, criticality phase, dispatch and completion; caps use >=', 'criticality_order': [0.0, 0.1, 1.0, 3.0], 'criticality_scheduling': 'after all five population lanes stop, four synchronous j phases in listed order before next group', 'group_order': 'ascending HMAC-SHA256 of signed population_identity bytes; byte-order tie break', 'limits': {'core_hours': 2880, 'wall_days': 15}, 'meter': 'raw measured cumulative binding-plus-safety core-hours, no overhead multiplier', 'parallel_attempts': False, 'pause': 'only recorded section-11 integrity evidence; resume same group and next fixed attempt index; clock continues', 'reserved_worker_slots': 8, 'start': 'immediately after every binding row completes', 'termination': 'irreversible gold, shared caps, grid completion or recorded section-11 integrity stop; snapshot counts immediately', 'wall_clock': 'seconds since first full-OC process start, never paused', 'worker_mapping': 'population i uses slot i, i=0..4; slots 5..7 reserved idle; one synchronous lane at a time, increasing population index; helper is a child of the active lane'}
PHASES = frozenset(("population", "audit", "bootstrap", "missing-mask", "canonical-safety", "canonical-criticality", "sampling-gate-estimate"))


def canonical(value):
    return json.dumps(value, sort_keys=True, separators=(",", ":"), allow_nan=False).encode()


def digest(data):
    return hashlib.sha256(data).hexdigest()


def _kernel():
    path = EXPERIMENT / "validation/fixed_population.py"
    if digest(path.read_bytes()) != KERNEL_SHA256:
        raise RuntimeError("signed population identity source changed")
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
    cell_bytes = cell.encode("utf-8") if isinstance(cell, str) else canonical(cell)
    message = b"\0".join((cell_bytes, str(population).encode(), str(attempt).encode(), phase.encode()))
    return int.from_bytes(hmac.digest(b"169-x1-oc-v4", message, "sha256")[:8], "big")


def safety_groups(manifest):
    specification = manifest["final_safety_OC"]
    dimensions = specification["dimensions"]
    for values in itertools.product(*dimensions.values()):
        for kappa in specification["unresolved_kappa"]:
            yield population_identity("safety-interval-group", {
                **dict(zip(dimensions, values)), "unresolved_kappa": kappa})


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


@dataclass(eq=False)
class Population:
    group_identity: str
    population_index: int
    p_gate: float
    safety: bool = False
    attempts: int = field(init=False, default=0)
    executions: int = field(init=False, default=0)
    eligible: int = field(init=False, default=0)
    inadequate: int = field(init=False, default=0)
    crashed: int = field(init=False, default=0)
    incidents: list = field(init=False, default_factory=list)
    _owner: object = field(init=False, default=None, repr=False)

    def __post_init__(self):
        if type(self.p_gate) not in (int, float) or not math.isfinite(self.p_gate) or not 0 < self.p_gate <= 1:
            raise ValueError("zero/unmeasured gate probability")
        if type(self.population_index) is not int or self.population_index not in range(5) or type(self.safety) is not bool:
            raise ValueError("invalid population identity")
        self.attempt_cap = math.ceil(4000/self.p_gate)

    @property
    def complete(self):
        return self.crashed == 0 and (self.executions if self.safety else self.eligible) == 2000

    @property
    def exhausted(self):
        return not self.complete and (self.attempts >= self.attempt_cap or self.executions >= 4000)

    def step(self, gate, trial):
        if self._owner is None:
            raise RuntimeError("population has no execution lease")
        self._owner.acquire(self)
        try:
            self._step(gate, trial)
        finally:
            self._owner.release()

    def _step(self, gate, trial):
        if self.complete or self.exhausted:
            raise RuntimeError("population stopped")
        index = self.attempts
        self.attempts += 1
        passes = gate(index)
        if type(passes) is not bool:
            self._owner.pause({"kind": "execution integrity", "evidence_sha256": digest(b"non-bool gate")})
            raise ValueError("gate must report a bool")
        if passes:
            self.executions += 1
            self.crashed += 1  # Unclassified in-flight work remains charged at an irreversible stop.
            try:
                adequate = trial(index)
                if self._owner.terminated:
                    raise RuntimeError("result discarded after termination")
                if type(adequate) is not bool:
                    raise ValueError("trial must report a bool")
            except Exception as error:
                self.incidents.append({"attempt": index, "type": type(error).__name__, "resolved": False})
                if not self._owner.terminated:
                    self._owner.pause({"kind": "execution integrity", "evidence_sha256": digest(type(error).__name__.encode())})
                raise
            self.crashed -= 1
            self.eligible += int(adequate)
            self.inadequate += int(not adequate)
        self._owner.after_attempt()

    def report(self):
        if self.executions != self.eligible+self.inadequate+self.crashed:
            raise RuntimeError("execution classification invariant violated")
        return {"group_identity": self.group_identity, "population_index": self.population_index,
                "attempts": self.attempts, "executions": self.executions, "eligible": self.eligible,
                "inadequate": self.inadequate, "crashed": self.crashed,
                "gate_failures": self.attempts-self.executions,
                "gate_failure_Wilson_MC95": wilson(self.attempts-self.executions, self.attempts),
                "conditional_inadequacy_Wilson_MC95": wilson(self.inadequate, self.executions),
                "incidents": copy.deepcopy(self.incidents)}


class SafetySchedule:
    """Eight reserved slots; serial population lanes and criticality phases.

    Meter is raw measured cumulative OC CPU, including binding. Clock is the
    unpaused full-OC wall clock; 2x is a projection rule, never an actual meter.
    Every execution lease consults this state machine before each attempt.
    """
    def __init__(self, *, freeze_path, p_gate, meter, oc_started_at, clock=time.time,
                 manifest_bytes=None):
        raw = (EXPERIMENT/"oc-manifest.json").read_bytes() if manifest_bytes is None else manifest_bytes
        if digest(raw) != MANIFEST_SHA256:
            raise ValueError("signed OC manifest bytes changed")
        self.manifest = json.loads(raw)
        identities = tuple(safety_groups(self.manifest))
        if len(identities) != 2016 or len(set(identities)) != 2016 or sum(json.loads(s)[1]["unresolved_kappa"] == 0 for s in identities) != 336:
            raise RuntimeError("safety grid changed")
        self._ordered = tuple(sorted(identities, key=lambda s: (hmac.digest(b"169-x1-safety-order", s.encode(), "sha256"), s.encode())))
        self._ordered_sha256 = digest(canonical(self._ordered))
        self._policy_bytes = canonical(POLICY)
        self._policy = immutable(copy.deepcopy(POLICY))
        self.check_freeze()
        Population("validation", 0, p_gate)
        if type(oc_started_at) not in (int, float) or not math.isfinite(oc_started_at):
            raise ValueError("invalid full-OC start")
        self.p_gate, self.meter, self.clock = p_gate, meter, clock
        self.oc_started_at = oc_started_at
        self._last_core = self._last_wall = 0
        self._leases = {}
        self._execution_lock = threading.Lock()
        self._binding = {}
        self._active = None
        self._completed = []
        self.events = []
        self.paused = False
        self.terminated = False
        self._final = None
        self._started_safety = False
        now = self.clock()
        self.freeze = {"ordered_groups_sha256": ORDER_SHA256, "policy_sha256": POLICY_SHA256,
                       "created_at_utc": datetime.fromtimestamp(now, timezone.utc).isoformat(),
                       "synthetic_only": True, "authority": "none; scoped review required",
                       "source_sha256": {str(p.relative_to(EXPERIMENT)): digest(p.read_bytes()) for p in
                           (Path(__file__).resolve(), EXPERIMENT/"validation/fixed_population.py", EXPERIMENT/"oc-manifest.json")}}
        with Path(freeze_path).open("x") as stream:
            stream.write(canonical(self.freeze).decode()+"\n")

    @property
    def ordered(self):
        return self._ordered

    @property
    def policy(self):
        return self._policy

    def check_freeze(self):
        if self._ordered_sha256 != ORDER_SHA256 or digest(self._policy_bytes) != POLICY_SHA256:
            raise RuntimeError("frozen safety order or policy mismatch")

    def _boundary(self, *, allow_paused=False):
        self.check_freeze()
        if self.terminated:
            raise RuntimeError("OC terminated")
        core, wall = self.meter(), (self.clock()-self.oc_started_at)/86400
        if any(type(x) not in (int, float) or not math.isfinite(x) or x < 0 for x in (core, wall)) or core < self._last_core or wall < self._last_wall:
            raise ValueError("invalid/nonmonotone OC meter")
        self._last_core, self._last_wall = core, wall
        if core >= 2880 or wall >= 15:
            self._terminate("OC core-hour cap" if core >= 2880 else "OC wall-day cap")
            raise RuntimeError("OC cap reached")
        if self.paused and not allow_paused:
            raise RuntimeError("OC paused for integrity review")

    def binding_population(self, identity, index):
        self._boundary()
        if self._started_safety:
            raise RuntimeError("binding cannot restart after safety")
        key = (identity, index)
        if key in self._binding:
            raise ValueError("duplicate binding lease")
        p = Population(identity, index, self.p_gate)
        p._owner = self
        self._binding[key] = p
        self._leases[id(p)] = key
        return p

    def next_group(self, *, binding_rows_complete):
        if type(binding_rows_complete) is not bool or not binding_rows_complete:
            raise RuntimeError("all binding rows must complete before safety")
        self._boundary()
        if self._active is not None:
            raise RuntimeError("active group unfinished")
        if len(self._completed) == len(self.ordered):
            self._terminate("grid complete")
            return None
        if not self._started_safety:
            self.events.append({"event": "start", "cause": "all binding rows complete", "wall_seconds": self.clock()-self.oc_started_at})
            self._started_safety = True
        identity = self.ordered[len(self._completed)]
        populations = tuple(Population(identity, i, self.p_gate, safety=True) for i in range(5))
        for p in populations:
            p._owner = self
            self._leases[id(p)] = (identity, p.population_index)
        self._active = {"identity": identity, "populations": populations, "criticality": {}}
        return populations

    def acquire(self, population):
        if not self._execution_lock.acquire(blocking=False):
            raise RuntimeError("parallel or reentrant execution refused")
        try:
            self.authorize(population)
        except BaseException:
            self._execution_lock.release()
            raise

    def release(self):
        self._execution_lock.release()

    def authorize(self, population):
        self._boundary()
        if self._leases.get(id(population)) != (population.group_identity, population.population_index):
            raise RuntimeError("stale or foreign execution lease")
        if population.safety:
            active = self._active
            if active is None or population not in active["populations"]:
                raise RuntimeError("foreign safety population")
            pending = next((p for p in active["populations"] if not p.complete and not p.exhausted), None)
            if pending is not population:
                raise RuntimeError("fixed population lane order violated")
        elif self._started_safety:
            raise RuntimeError("binding execution after safety start")

    def after_attempt(self):
        self._boundary()

    @staticmethod
    def _evidence(evidence):
        if (not isinstance(evidence, dict) or evidence.get("kind") not in
            ("execution integrity", "closure mismatch", "independent OS malfunction", "reviewed integrity resolution")
            or not isinstance(evidence.get("evidence_sha256"), str) or len(evidence["evidence_sha256"]) != 64
            or any(c not in "0123456789abcdef" for c in evidence["evidence_sha256"])):
            raise ValueError("section-11 integrity evidence required")

    def pause(self, evidence):
        self._boundary(allow_paused=True)
        self._evidence(evidence)
        if self.paused:
            raise RuntimeError("already paused")
        self.paused = True
        self.events.append({"event": "pause", "cause": copy.deepcopy(evidence), "wall_seconds": self.clock()-self.oc_started_at})

    def resolve_crash(self, population, attempt, adequate, evidence):
        self._boundary(allow_paused=True)
        self._evidence(evidence)
        if not self.paused or evidence["kind"] != "reviewed integrity resolution" or type(adequate) is not bool or id(population) not in self._leases:
            raise RuntimeError("reviewed paused crash resolution required")
        incident = next((i for i in population.incidents if i["attempt"] == attempt and not i["resolved"]), None)
        if incident is None:
            raise ValueError("no unresolved execution")
        incident.update(resolved=True, evidence_sha256=evidence["evidence_sha256"])
        population.crashed -= 1
        population.eligible += int(adequate)
        population.inadequate += int(not adequate)
        self.events.append({"event": "resolve", "population": population.population_index, "attempt": attempt, "cause": copy.deepcopy(evidence)})

    def resume(self, evidence):
        self._boundary(allow_paused=True)
        self._evidence(evidence)
        if not self.paused or evidence["kind"] != "reviewed integrity resolution" or any(p.crashed for p in self._all_populations()):
            raise RuntimeError("unresolved integrity pause")
        self.paused = False
        self.events.append({"event": "resume", "cause": copy.deepcopy(evidence), "wall_seconds": self.clock()-self.oc_started_at})

    def _all_populations(self):
        return list(self._binding.values()) + (list(self._active["populations"]) if self._active else [])

    def run_criticality(self, j, work):
        if not self._execution_lock.acquire(blocking=False):
            raise RuntimeError("parallel or reentrant criticality refused")
        try:
            return self._run_criticality(j, work)
        finally:
            self._execution_lock.release()

    def _run_criticality(self, j, work):
        self._boundary()
        active = self._active
        if active is None or not all(p.complete or p.exhausted for p in active["populations"]) or any(p.crashed for p in active["populations"]):
            raise RuntimeError("populations unfinished")
        values = self.manifest["final_safety_OC"]["critical_j"]
        if type(j) not in (int, float) or isinstance(j, bool) or j != values[len(active["criticality"])]:
            raise ValueError("criticality order changed")
        key = criticality_key(active["identity"], values[len(active["criticality"])])
        result = work(key)
        self._boundary()
        if not isinstance(result, dict) or result.get("completed") is not True or result.get("key") != key:
            raise ValueError("incomplete or foreign criticality record")
        active["criticality"][key] = copy.deepcopy(result)
        if len(active["criticality"]) == len(values):
            self.finish(active["populations"], active["criticality"])
        return key

    def finish(self, populations, criticality_records):
        self._boundary()
        active = self._active
        if active is None or len(populations) != 5 or {id(p) for p in populations} != {id(p) for p in active["populations"]}:
            raise RuntimeError("duplicate, missing or foreign population")
        if {p.population_index for p in populations} != set(range(5)) or any(not p.safety or p.group_identity != active["identity"] for p in populations):
            raise RuntimeError("population identity mismatch")
        keys = {criticality_key(active["identity"], j) for j in self.manifest["final_safety_OC"]["critical_j"]}
        if set(criticality_records) != keys or criticality_records != active["criticality"]:
            raise RuntimeError("incomplete or substituted criticality")
        if not all(p.complete or p.exhausted for p in populations) or any(p.crashed for p in populations):
            raise RuntimeError("populations unfinished")
        self._complete_active("measured" if all(p.complete for p in populations) else "unmeasured; cell fails")

    def mark_unreachable(self, calibration, expected_sha256):
        self._boundary()
        active = self._active
        if (active is None or any(p.attempts for p in active["populations"])
            or digest(canonical(calibration)) != expected_sha256
            or calibration.get("group_identity") != active["identity"] or calibration.get("reachable") is not False
            or calibration.get("calibration_source_sha256") != KERNEL_SHA256
            or calibration.get("target") != json.loads(active["identity"])[1]["realised_gain"]):
            raise ValueError("unbound unreachable calibration")
        active["calibration"] = copy.deepcopy(calibration)
        self._complete_active("unmeasured (unreachable target)")

    def _snapshot(self, active, status):
        return {"identity": active["identity"], "status": status,
                "populations": [p.report() for p in active["populations"]],
                "criticality": copy.deepcopy(active["criticality"]),
                "calibration": copy.deepcopy(active.get("calibration"))}

    def _complete_active(self, status):
        self._completed.append(self._snapshot(self._active, status))
        self._active = None
        if len(self._completed) == len(self.ordered):
            self._terminate("grid complete")

    def _terminate(self, cause):
        if self.terminated:
            return
        self.terminated = True
        self.events.append({"event": "termination", "cause": cause, "group_index": len(self._completed), "wall_seconds": self.clock()-self.oc_started_at})
        groups = copy.deepcopy(self._completed)
        if self._active:
            status = "unmeasured; cell fails" if any(p.exhausted for p in self._active["populations"]) else "not completed by gold"
            groups.append(self._snapshot(self._active, status))
        for identity in self.ordered[len(groups):]:
            groups.append({"identity": identity, "status": "not completed by gold", "populations": [], "criticality": {}, "calibration": None})
        self._final = {"groups": groups, "pooled_group_indices": [i for i, g in enumerate(groups) if g["status"] == "measured"],
                       "events": copy.deepcopy(self.events), "freeze": copy.deepcopy(self.freeze)}

    def stop_gold(self):
        self.check_freeze()
        self._terminate("gold")

    def stop_integrity(self, evidence):
        self._evidence(evidence)
        self._terminate({"section_11_integrity_stop": copy.deepcopy(evidence)})

    def gold_report(self):
        self.check_freeze()
        if not self.terminated:
            raise RuntimeError("termination event required before final report")
        return copy.deepcopy(self._final)

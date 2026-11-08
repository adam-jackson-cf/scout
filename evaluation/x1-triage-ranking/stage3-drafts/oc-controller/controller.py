"""Synthetic draft controller: deterministic counters and outcome-blind scheduling."""
import hashlib
import hmac
import json
import math
import itertools
from dataclasses import dataclass


def canonical(value):
    return json.dumps(value, sort_keys=True, separators=(",", ":"), allow_nan=False).encode()


def seed(cell, population, attempt, phase):
    message = b"\0".join((canonical(cell), str(population).encode(), str(attempt).encode(), phase.encode()))
    return int.from_bytes(hmac.digest(b"169-x1-oc-v4", message, "sha256")[:8], "big")


def safety_groups(manifest):
    specification = manifest["final_safety_OC"]
    dimensions = specification["dimensions"]
    for values in itertools.product(*dimensions.values()):
        for kappa in specification["unresolved_kappa"]:
            yield {"kind": "safety-interval-group", **dict(zip(dimensions, values)),
                   **specification["fixed_factors"], "unresolved_kappa": kappa}


def wilson(k, n):
    if not n:
        return None
    z = 1.959963984540054
    p = k/n
    centre = (p+z*z/(2*n))/(1+z*z/n)
    half = z*math.sqrt(p*(1-p)/n+z*z/(4*n*n))/(1+z*z/n)
    return [centre-half, centre+half]


@dataclass
class Population:
    p_gate: float
    safety: bool = False
    attempts: int = 0
    executions: int = 0
    eligible: int = 0
    inadequate: int = 0

    def __post_init__(self):
        if not math.isfinite(self.p_gate) or not 0 < self.p_gate <= 1:
            raise ValueError("zero/unmeasured gate probability")
        self.attempt_cap = math.ceil(4000/self.p_gate)

    @property
    def complete(self):
        return (self.executions if self.safety else self.eligible) == 2000

    @property
    def exhausted(self):
        return not self.complete and (self.attempts >= self.attempt_cap or self.executions >= 4000)

    def step(self, gate, trial):
        if self.complete or self.exhausted:
            raise RuntimeError("population stopped")
        index = self.attempts
        self.attempts += 1
        if not gate(index):
            return  # Never invoke expensive trial after rejection.
        self.executions += 1  # Charge even a crashing/inadequate trial.
        adequate = trial(index)
        if type(adequate) is not bool:
            raise ValueError("trial must report adequacy")
        self.eligible += int(adequate)
        self.inadequate += int(not adequate)

    def report(self):
        return dict(attempts=self.attempts, executions=self.executions, eligible=self.eligible,
                    gate_failures=self.attempts-self.executions, inadequate=self.inadequate,
                    gate_failure_Wilson_MC95=wilson(self.attempts-self.executions, self.attempts),
                    conditional_inadequacy_Wilson_MC95=wilson(self.inadequate, self.executions),
                    status="measured" if self.complete else "unmeasured; cell fails" if self.exhausted else "not completed by gold")


class SafetySchedule:
    """One serial group, eight reserved workers, no elective pauses; stop at gold/caps.

    No outcome argument is accepted by any scheduling operation. Resource values
    are cumulative OC usage (binding included), already charged with 2x overhead.
    """
    def __init__(self, groups):
        encoded = [canonical(g) for g in groups]
        if len(set(encoded)) != len(encoded):
            raise ValueError("duplicate group")
        if any(g.get("kind") != "safety-interval-group" or "critical_share" in g for g in groups):
            raise ValueError("not an interval-group identity")
        self.ordered = tuple(sorted(encoded, key=lambda g: (hmac.digest(b"169-x1-safety-order", g, "sha256"), g)))
        self.policy = {"start": "after all binding rows", "workers": 8, "pauses": "none", "termination": "gold or cumulative OC caps", "groups": "serial; canonical-criticality completes before next group"}
        self.freeze = {"ordered_groups_sha256": hashlib.sha256(canonical([json.loads(g) for g in self.ordered])).hexdigest(),
                       "policy_sha256": hashlib.sha256(canonical(self.policy)).hexdigest()}
        self.index = 0
        self.active = False
        self.terminated = False
        self.events = []
        self.statuses = []

    def check_freeze(self):
        actual = {"ordered_groups_sha256": hashlib.sha256(canonical([json.loads(g) for g in self.ordered])).hexdigest(),
                  "policy_sha256": hashlib.sha256(canonical(self.policy)).hexdigest()}
        if actual != self.freeze:
            raise RuntimeError("frozen order or scheduling policy changed")

    def next_group(self, *, binding_rows_complete, core_hours, wall_days, gold=False):
        self.check_freeze()
        if not binding_rows_complete:
            raise RuntimeError("binding rows must precede safety")
        if not all(math.isfinite(x) and x >= 0 for x in (core_hours, wall_days)):
            raise ValueError("unmeasured resource usage")
        if gold or core_hours >= 2880 or wall_days >= 15:
            if not self.terminated:
                self.events.append({"cause": "gold" if gold else "OC resource cap", "group_index": self.index})
            self.terminated = True
            return None
        if self.terminated:
            return None
        if self.active:
            raise RuntimeError("finish canonical-criticality before next group")
        if self.index == len(self.ordered):
            return None
        self.active = True
        return json.loads(self.ordered[self.index])

    def finish(self, populations, *, criticality_complete):
        self.check_freeze()
        if not self.active or self.terminated or not criticality_complete or len(populations) != 5 or any(not p.safety for p in populations):
            raise RuntimeError("invalid group completion")
        if not all(p.complete or p.exhausted for p in populations):
            raise RuntimeError("unfinished populations")
        status = "measured" if all(p.complete for p in populations) else "unmeasured; cell fails"
        self.statuses.append(status)
        self.index += 1
        self.active = False

    def gold_report(self, partial_populations=()):
        self.check_freeze()
        # Exhaustion has precedence over gold completion status.
        pending = "unmeasured; cell fails" if any(p.exhausted for p in partial_populations) else "not completed by gold"
        remaining = len(self.ordered)-self.index
        statuses = self.statuses + ([pending]+["not completed by gold"]*(remaining-1) if remaining else [])
        return {"statuses": statuses, "partial_counts": [p.report() for p in partial_populations],
                "pooled_group_indices": [i for i, s in enumerate(statuses) if s == "measured"],
                "events": list(self.events), "freeze": dict(self.freeze)}

"""Synthetic-only, immutable locked-size boundary; no repository access."""
from dataclasses import dataclass
import hashlib
import importlib.util
import json
from pathlib import Path
import random
import re
import sys
from types import MappingProxyType

REPOSITORIES = ("synthetic-python", "synthetic-typescript", "synthetic-go")
MAPPING = ("django/django", "microsoft/TypeScript", "prometheus/prometheus")
TIPS = ("a05e16a7230fe33d3f43f4a3aaed103cebb42efc", "6161772c1370bdd4a82f808e6823bd755eb1870d", "ef78ff64267724c8ecd1a1283cad49b365a8b67f")
SCHEMA = "x1-locked-size-input/synthetic-v1"
GATE_CELL = '{"kind":"sampling-gate-estimate"}'
DRAWS = 100000
CONTROLLER_PATH = Path(__file__).resolve().parent.parent / "oc-controller/controller.py"


def canonical(value):
    return json.dumps(value, sort_keys=True, separators=(",", ":"), allow_nan=False).encode()


def controller():
    # Explicit file import: the controller owns seed and Wilson conventions.
    name = "x1_locked_size_controller"
    if name not in sys.modules:
        spec = importlib.util.spec_from_file_location(name, CONTROLLER_PATH)
        module = importlib.util.module_from_spec(spec)
        sys.modules[name] = module
        spec.loader.exec_module(module)
    return sys.modules[name]


def controller_seed(cell, population, attempt, phase):
    return controller().seed(cell, population, attempt, phase)


def _hex(value, width):
    return isinstance(value, str) and re.fullmatch(f"[0-9a-f]{{{width}}}", value) is not None


@dataclass(frozen=True, init=False)
class SizeInput:
    """Full ordered synthetic inventory record, bound by canonical SHA-256.

    Positions attest first-parent oldest-to-newest order; this synthetic-only
    draft does not inspect repositories or establish inventory authenticity.
    """
    vectors: object
    sha256: str
    minimum: int
    maximum: int
    record_bytes: bytes

    def __init__(self, record):
        if set(record) != {"schema", "synthetic_only", "inventory_seal_sha256", "repositories"}:
            raise ValueError("exact size-record fields required")
        if record["schema"] != SCHEMA or record["synthetic_only"] is not True:
            raise ValueError("only synthetic draft inputs accepted")
        if not _hex(record["inventory_seal_sha256"], 64):
            raise ValueError("inventory seal SHA-256 required")
        repositories = record["repositories"]
        if not isinstance(repositories, list) or len(repositories) != 3:
            raise ValueError("exactly three repository records required")
        vectors = {}
        for alias, repository, tip, item in zip(REPOSITORIES, MAPPING, TIPS, repositories):
            if set(item) != {"alias", "repository", "pinned_tip", "runs"}:
                raise ValueError("exact repository fields required")
            if (item["alias"], item["repository"], item["pinned_tip"]) != (alias, repository, tip):
                raise ValueError("fixed repository mapping, pinned tip and traversal order required")
            runs = item["runs"]
            if not isinstance(runs, list) or not 20 <= len(runs) <= 2000:
                raise ValueError("20..2000 complete eligible runs required")
            previous, shas, sizes = -1, set(), []
            for run in runs:
                if set(run) != {"commit_sha", "first_parent_position", "size"}:
                    raise ValueError("exact run fields required")
                position, n, sha = run["first_parent_position"], run["size"], run["commit_sha"]
                if type(position) is not int or not previous < position < 2000:
                    raise ValueError("runs must follow first-parent positions 0..1999, oldest first")
                if not _hex(sha, 40) or sha in shas:
                    raise ValueError("unique full lowercase commit SHA required")
                if type(n) is not int or not 20 <= n <= 2000:
                    raise ValueError("run sizes must be integer 20..2000")
                previous = position
                shas.add(sha)
                sizes.append(n)
            vectors[alias] = tuple(sizes)
        data = canonical(record)
        object.__setattr__(self, "vectors", MappingProxyType(vectors))
        object.__setattr__(self, "record_bytes", data)
        object.__setattr__(self, "sha256", hashlib.sha256(data).hexdigest())
        object.__setattr__(self, "minimum", min(map(min, vectors.values())))
        object.__setattr__(self, "maximum", max(map(max, vectors.values())))

    def size(self, repository, run_index):
        return self.vectors[repository][run_index]

    def offset(self, repository, run_index, factor):
        if factor not in (0, .05, .20):
            raise ValueError("unregistered factor")
        return factor * (self.size(repository, run_index)-self.minimum) / max(1, self.maximum-self.minimum)

    @property
    def lengths(self):
        return {r: len(self.vectors[r]) for r in REPOSITORIES}

    @property
    def units(self):
        return sum(map(sum, self.vectors.values()))

    def gate_selection(self, attempt):
        rng = random.Random(controller_seed(GATE_CELL, 0, attempt, "sampling-gate-estimate"))
        selected = {r: tuple(rng.sample(range(len(self.vectors[r])), 20)) for r in REPOSITORIES}
        totals = [sum(self.vectors[r][i] for i in selected[r]) for r in REPOSITORIES]
        return selected, gate_passes(totals)

    def estimate_p_gate(self):
        successes = sum(self.gate_selection(i)[1] for i in range(DRAWS))
        p = successes/DRAWS
        return {"draws": DRAWS, "successes": successes, "p_gate": p,
                "Wilson_MC95": controller().wilson(successes, DRAWS),
                "attempt_cap_per_population": (4000*DRAWS+successes-1)//successes if successes else None,
                "hard_trial_execution_cap_per_population": 4000,
                "status": "estimated" if successes else "not testable",
                "cell_identity": GATE_CELL, "size_input_sha256": self.sha256, "L_r": self.lengths}


def gate_passes(totals):
    return min(totals) >= 1000 and sum(totals) <= 60000


def synthetic_record(vectors):
    """Fabricate identities for synthetic fixtures only, never an inventory seal."""
    if set(vectors) != set(REPOSITORIES):
        raise ValueError("exactly three synthetic repository vectors required")
    return {"schema": SCHEMA, "synthetic_only": True,
            "inventory_seal_sha256": "0"*64,
            "repositories": [{"alias": alias, "repository": repository, "pinned_tip": tip,
                              "runs": [{"commit_sha": f"{(repo_index+1)*10000+i:040x}",
                                        "first_parent_position": i, "size": n}
                                       for i, n in enumerate(vectors[alias])]}
                             for repo_index, (alias, repository, tip) in enumerate(zip(REPOSITORIES, MAPPING, TIPS))]}

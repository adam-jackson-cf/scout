"""Draft size-source boundary. No repository access, scans or authority."""
import hashlib
import json
import random

REPOSITORIES = ("synthetic-python", "synthetic-typescript", "synthetic-go")


def canonical(value):
    return json.dumps(value, sort_keys=True, separators=(",", ":"), allow_nan=False).encode()


class SizeInput:
    """Complete, ordered per-repository vectors; never templates or padding."""

    def __init__(self, vectors):
        if set(vectors) != set(REPOSITORIES):
            raise ValueError("exactly three synthetic repository vectors required")
        self.vectors = {r: tuple(vectors[r]) for r in REPOSITORIES}
        for values in self.vectors.values():
            if not 20 <= len(values) <= 2000:
                raise ValueError("20..2000 complete eligible runs required")
            if any(type(n) is not int or not 20 <= n <= 2000 for n in values):
                raise ValueError("run sizes must be integer 20..2000")
        self.sha256 = hashlib.sha256(canonical(self.vectors)).hexdigest()
        self.minimum = min(map(min, self.vectors.values()))
        self.maximum = max(map(max, self.vectors.values()))

    def size(self, repository, run_index):
        return self.vectors[repository][run_index]

    def offset(self, repository, run_index, factor):
        if factor not in (0, .05, .20):
            raise ValueError("unregistered factor")
        return factor * (self.size(repository, run_index)-self.minimum) / max(1, self.maximum-self.minimum)

    def draw(self, seed):
        rng = random.Random(seed)
        selected = {r: tuple(sorted(rng.sample(range(len(v)), 20))) for r, v in self.vectors.items()}
        totals = [sum(self.vectors[r][i] for i in selected[r]) for r in REPOSITORIES]
        return selected, min(totals) >= 1000 and sum(totals) <= 60000

    def measurement_sizes(self):
        # Preserve every actual run. No cyclic replication or L40/L400 truncation.
        return {r: tuple(enumerate(v)) for r, v in self.vectors.items()}

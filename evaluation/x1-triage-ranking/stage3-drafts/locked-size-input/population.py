"""Draft locked-size variant: only the size source differs from the signed kernel.

No benchmark, repository IO, attestation or real-data execution occurs on import.
"""
import math
import random
from pathlib import Path
import sys

VALIDATION = Path(__file__).resolve().parents[2] / "validation"
sys.path.insert(0, str(VALIDATION))
from fixed_population import PreparedPopulation, PreparedRun, gradient_slots


def prepare_population(*, seed=9, prevalence=.20, heterogeneity=0., overlap=0.,
                       size_input, cap=2000, failure=0.,
                       scanner_true=False, rank_structure='uniform', size_dependence=0.):
    if overlap not in (0., .5) or cap not in (400, 550, 700, 850, 1000, 1500, 2000):
        raise ValueError('unregistered overlap/cap')
    if rank_structure not in ('uniform', 'linear-gradient') or size_dependence not in (0., .05, .20):
        raise ValueError('unregistered rank structure/size dependence')
    rng, runs = random.Random(seed), []
    for repo in ('synthetic-python', 'synthetic-typescript', 'synthetic-go'):
        for run in range(len(size_input.vectors[repo])):
            n = size_input.size(repo, run)
            positives, k = round(n*prevalence), math.ceil(.30*n)
            ceiling = min(1., k/positives)
            offset = rng.gauss(0., heterogeneity)
            while not 0 <= .30-offset <= ceiling:
                offset = rng.gauss(0., heterogeneity)
            base_count = round(positives*(.30-offset))
            positive, negative = list(range(positives)), list(range(positives, n))
            base_keys = {i: rng.random() for i in positive}
            jev_keys = {i: base_keys[i] if rng.random() < overlap else rng.random() for i in positive}
            positive_order = sorted(positive, key=jev_keys.get)
            rng.shuffle(negative)
            top = sorted(positive, key=base_keys.get)[:base_count]+negative[:k-base_count]
            top_set = set(top)
            rest = [i for i in range(n) if i not in top_set]
            rng.shuffle(top)
            if rank_structure == 'uniform':
                rng.shuffle(rest)
            else:
                p = [i for i in rest if i < positives]
                q = [i for i in rest if i >= positives]
                rng.shuffle(p)
                rng.shuffle(q)
                rest = [p.pop() if slot else q.pop() for slot in gradient_slots(len(rest), len(p), rng.getrandbits(64))]
            base_order = tuple(top+rest)
            candidate = base_order[:cap]
            failed = frozenset(i for i in candidate if rng.random() < failure)
            annotated = frozenset(candidate)-failed
            scores = tuple(rng.random() for _ in range(n))
            pool = list(range(n)) if scanner_true else list(range(positives, n))
            severity = frozenset(rng.sample(pool, max(1, round(.02*n))))
            size_offset = size_input.offset(repo, run, size_dependence)
            runs.append(PreparedRun(repo, f'{run:04}', n, positives, k, base_order,
                                    tuple(positive_order), tuple(negative), annotated,
                                    failed, severity, scores, size_offset))
    return PreparedPopulation(tuple(runs))


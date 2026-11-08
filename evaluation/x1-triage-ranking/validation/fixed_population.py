"""Synthetic fixed-key X1 populations: allocation changes never redraw keys."""
from dataclasses import dataclass
import hashlib
import json
import math
import random
import struct


@dataclass(frozen=True)
class PreparedRun:
    repo: str
    run: str
    n: int
    positives: int
    k: int
    base_order: tuple
    positive_order: tuple
    negative_order: tuple
    annotated: frozenset
    failed: frozenset
    severity: frozenset
    scores: tuple
    size_offset: float


@dataclass(frozen=True)
class PreparedPopulation:
    runs: tuple


def gradient_slots(length, positives, seed):
    """Fixed-count systematic sampling of the registered linear density."""
    density = positives/length
    slope = min(density, 1-density)
    phase = random.Random(seed).random()
    cumulative = phase
    slots = []
    for index in range(length):
        probability = density+slope*(1-(2*index+1)/length)
        following = phase+(index+1)*density+slope*(index+1)*(1-(index+1)/length)
        slots.append(math.floor(following) > math.floor(cumulative))
        cumulative = following
    return slots


def prepare_population(*, seed=9, prevalence=.20, heterogeneity=0., overlap=0.,
                       run_count=20, sizes=(50,), cap=2000, failure=0.,
                       scanner_true=False, rank_structure='uniform', size_dependence=0.):
    if overlap not in (0., .5) or cap not in (400, 550, 700, 850, 1000, 1500, 2000):
        raise ValueError('unregistered overlap/cap')
    if rank_structure not in ('uniform', 'linear-gradient') or size_dependence not in (0., .05, .20):
        raise ValueError('unregistered rank structure/size dependence')
    rng, runs = random.Random(seed), []
    for repo in ('synthetic-python', 'synthetic-typescript', 'synthetic-go'):
        for run in range(run_count):
            n = sizes[run % len(sizes)]
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
            size_offset = size_dependence*(n-min(sizes))/max(1, max(sizes)-min(sizes))
            runs.append(PreparedRun(repo, f'{run:04}', n, positives, k, base_order,
                                    tuple(positive_order), tuple(negative), annotated,
                                    failed, severity, scores, size_offset))
    return PreparedPopulation(tuple(runs))


def _run_order(run, gain):
    count = round(run.positives*(.30+gain-run.size_offset))
    if not 0 <= count <= min(run.k, run.positives):
        raise ValueError('infeasible integer retrieval allocation')
    desired = set(run.positive_order[:count]+run.negative_order[:run.k-count])
    scored = sorted(run.annotated, key=lambda i: (i not in desired, run.scores[i]))
    return scored+[i for i in run.base_order if i not in run.annotated]


def _population_gain(prepared, gain):
    """Exact full-population truth without allocating per-unit analysis rows."""
    totals = {}
    for run in prepared.runs:
        order = _run_order(run, gain)
        captured = sum(issue < run.positives for issue in order[:run.k])
        baseline = sum(issue < run.positives for issue in run.base_order[:run.k])
        numerator, denominator = totals.get(run.repo, (0, 0))
        totals[run.repo] = (numerator+captured-baseline, denominator+run.positives)
    return sum(totals[repo][0]/totals[repo][1] for repo in sorted(totals))/len(totals)


def evaluate_population(prepared, gain):
    from ranking import Unit, assign
    frame = []
    for run in prepared.runs:
        order = _run_order(run, gain)
        jev = {issue: rank for rank, issue in enumerate(order, 1)}
        base = {issue: rank for rank, issue in enumerate(run.base_order, 1)}
        for issue in range(run.n):
            reason = 'network' if issue in run.failed else 'annotated' if issue in run.annotated else 'cost cap reached'
            frame.append(Unit(run.repo, run.run, f'{run.run}/{issue:05}', run.n,
                              jev[issue], base[issue], issue in run.annotated,
                              issue in run.severity, int(issue < run.positives), False, reason=reason))
    return assign(frame)


def realised_gain(frame):
    repos = sorted({r.repo for r in frame})
    return sum(sum(r.truth*r.delta for r in frame if r.repo == repo)/
               sum(r.truth for r in frame if r.repo == repo) for repo in repos)/len(repos)


def _float_index(value):
    """Monotone integer index for finite IEEE-754 doubles."""
    bits = struct.unpack('>Q', struct.pack('>d', value))[0]
    return (~bits & ((1 << 64)-1)) if bits >> 63 else bits ^ (1 << 63)


def _indexed_float(index):
    bits = index ^ (1 << 63) if index >> 63 else ~index & ((1 << 64)-1)
    return struct.unpack('>d', struct.pack('>Q', bits))[0]


def _allocation_candidates(prepared):
    """Represent every attainable allocation, including floating-point ties."""
    low = max(-.30+r.size_offset for r in prepared.runs)
    high = min(min(r.k, r.positives)/r.positives-.30+r.size_offset for r in prepared.runs)
    candidates = {low, high}
    # Repeated run sizes share transitions; search each structural type once.
    kinds = {(r.positives, min(r.k, r.positives), r.size_offset) for r in prepared.runs}
    for positives, maximum, offset in sorted(kinds):
        for count in range(maximum):
            if round(positives*(.30+low-offset)) > count:
                continue
            if round(positives*(.30+high-offset)) <= count:
                continue
            left, right = _float_index(low), _float_index(high)
            while left < right:
                middle = (left+right)//2
                gain = _indexed_float(middle)
                if round(positives*(.30+gain-offset)) > count:
                    right = middle
                else:
                    left = middle+1
            candidates.add(_indexed_float(left))
    return sorted(candidates)


def calibrate_population(prepared, target, steps=20):
    """Search integer allocation breakpoints, choosing the registered side exactly."""
    if target not in (0., .10, .20):
        raise ValueError('unregistered calibration target')
    # A nextafter on an algebraic half-integer threshold can still round back
    # across the transition after addition/multiplication. Search the exact
    # floating-point expression used by _run_order instead.
    candidates = _allocation_candidates(prepared)
    left, right = 0, len(candidates)-1
    positive_side = target > 0
    while left < right:
        middle = (left+right)//2 if positive_side else (left+right+1)//2
        value = _population_gain(prepared, candidates[middle])
        if (value >= target) if positive_side else (value > target):
            right = middle if positive_side else middle-1
        else:
            left = middle+1 if positive_side else middle
    frame = evaluate_population(prepared, candidates[left])
    realised = realised_gain(frame)
    if _population_gain(prepared, candidates[left]) != realised:
        raise RuntimeError('calibration truth/full-frame mismatch')
    reachable = realised >= target if positive_side else realised <= target
    return dict(frame=frame, target=target, realised=realised, excess=realised-target,
                latent_gain=candidates[left], reachable=reachable)


def population_identity(kind, dimensions):
    if kind not in ('safety-profile', 'safety-interval-group'):
        raise ValueError('unknown safety identity kind')
    values = dict(dimensions)
    if kind == 'safety-interval-group':
        values.pop('critical_share', None)
    return json.dumps([kind, values], sort_keys=True, separators=(',', ':'))


def criticality_key(interval_identity, critical_share):
    return hashlib.sha256(json.dumps(['canonical-criticality', interval_identity, critical_share],
                                     separators=(',', ':')).encode()).hexdigest()

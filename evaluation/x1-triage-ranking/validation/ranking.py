"""X1 rev4 synthetic statistical kernel; no production/model/network interface."""
from dataclasses import dataclass, replace
from itertools import product
import hashlib
import hmac
import math
import random


@dataclass(frozen=True)
class Unit:
    repo: str
    run: str
    issue: str
    n: int
    jev: int
    baseline: int
    annotated: bool
    severity_critical: bool
    truth: int | None
    critical: bool | None = None
    pi: float = 0.0
    selected: bool = False
    reason: str = 'annotated'

    @property
    def k(self):
        return max(1, math.ceil(self.n * .30))

    @property
    def delta(self):
        return int(self.jev <= self.k) - int(self.baseline <= self.k)

    @property
    def below(self):
        return self.jev > (self.n + 1) // 2


def audit_uniform(repo, run, issue):
    key = '\0'.join((repo, run, issue)).encode()
    digest = hmac.new(b'169-x1-audit', key, hashlib.sha256).digest()
    return (int.from_bytes(digest[:8], 'big') >> 11) / (1 << 53)


def label_payload(state):
    return {key: value for key, value in state.items() if key != 'severity'}


def assign(frame, fraction=1.):
    """Synthetic mirror of product Assign; product parity remains a later gate."""
    if not 0 < fraction <= 1:
        raise ValueError('invalid audit fraction')
    groups = {}
    for row in frame:
        groups.setdefault((row.repo, row.run), []).append(row)
    output = []
    for rows in groups.values():
        def stratum(row):
            if row.severity_critical and row.below:
                return 'certainty'
            if not row.annotated:
                return 'unannotated'
            return 'bottom' if row.below else 'top'
        sizes = {}
        for row in rows:
            s = stratum(row)
            sizes[s] = sizes.get(s, 0) + 1
        count = len(set(sizes) - {'certainty'})
        for row in rows:
            s = stratum(row)
            pi = 1. if s == 'certainty' else min(1., fraction * len(rows) / count / sizes[s])
            output.append(replace(row, pi=pi))
    return output


def draw(frame, rng=None):
    return [replace(row, selected=(rng.random() if rng else audit_uniform(row.repo, row.run, row.issue)) < row.pi) for row in frame]


def validate(frame):
    keys, groups = set(), {}
    for row in frame:
        key = (row.repo, row.issue)
        if key in keys:
            raise ValueError('duplicate issue across runs')
        keys.add(key)
        if type(row.truth) not in (int, type(None)) or row.truth not in (0, 1, None):
            raise ValueError('invalid truth')
        if type(row.critical) not in (bool, type(None)):
            raise ValueError('invalid criticality')
        if type(row.annotated) is not bool or type(row.severity_critical) is not bool or not 0 < row.pi <= 1:
            raise ValueError('invalid probability or frame flags')
        if row.reason not in STUDY_REASONS | PRODUCT_REASONS:
            raise ValueError('unknown treatment-delivery reason')
        groups.setdefault((row.repo, row.run), []).append(row)
    for rows in groups.values():
        n = len(rows)
        if any(row.n != n for row in rows):
            raise ValueError('incomplete frame')
        positions = list(range(1, n+1))
        if sorted(row.jev for row in rows) != positions or sorted(row.baseline for row in rows) != positions:
            raise ValueError('invalid rank permutation')


def safety(frame):
    violating = unresolved = 0
    for row in frame:
        if not row.below:
            continue
        if row.truth == 1 and (row.severity_critical or row.critical is True):
            violating += 1
        elif row.truth == 0 or (row.critical is False and not row.severity_critical):
            pass
        else:
            unresolved += 1
    return {'violating': violating, 'unresolved': unresolved}


def totals(rows, unknown):
    by_run = {}
    for row in rows:
        y = row.truth if row.truth is not None else unknown[row.delta]
        t, j, b = by_run.get(row.run, (0., 0., 0.))
        by_run[row.run] = (t+y/row.pi, j+y*int(row.jev <= row.k)/row.pi,
                           b+y*int(row.baseline <= row.k)/row.pi)
    return by_run


def fit(rows, unknown):
    values = totals(rows, unknown)
    total = sum(v[0] for v in values.values())
    if total <= 0:
        return None
    j, b = (sum(v[index] for v in values.values()) for index in (1, 2))
    gain = (j-b)/total
    influence = [math.sqrt(1-row.pi) * (row.truth if row.truth is not None else unknown[row.delta])
                 * (row.delta-gain)/row.pi/total for row in rows]
    return {'gain': gain, 'jev': j/total, 'baseline': b/total,
            'positive_total': total, 'jev_total': j, 'baseline_total': b,
            'by_run': values, 'influence': influence}


def percentile(values):
    values = sorted(values)
    n = len(values)
    return [max(-1., values[math.floor(.025*n)]), min(1., values[min(n-1, math.ceil(.975*n)-1)])]


def multiplier(point, influence, replicates, seed):
    rng = random.Random(seed)
    terms = [v for v in influence if v]
    return percentile([point+sum(v if rng.getrandbits(1) else -v for v in terms) for _ in range(replicates)])


def cluster_distribution(fitted, runs, draws):
    values = [fitted['by_run'].get(run, (0., 0., 0.)) for run in runs]
    result = []
    for indices in draws:
        denominator = sum(values[i][0] for i in indices)
        if denominator <= 0:
            result.append(None)
        else:
            result.append(sum(values[i][1]-values[i][2] for i in indices)/denominator)
    return result


def ratio_variance(values):
    """No-FPC run-linearised variance, s²(N-D*T)/(G*mean(T)²)."""
    count = len(values)
    total = sum(v[0] for v in values)
    if count < 2 or total <= 0:
        return None
    gain = sum(v[1]-v[2] for v in values)/total
    # CPython 3.11 left-to-right sums; Go fit uses these same two passes.
    residuals = [(v[1]-v[2])-gain*v[0] for v in values]
    center = sum(residuals)/count
    squares = [(v-center)*(v-center) for v in residuals]
    return count*sum(squares)/((count-1)*(total*total))


def interval_from_distribution(point, variance, distribution, variances, method):
    if method == 'percentile':
        return percentile(distribution), len(distribution)
    if method != 'studentized':
        raise ValueError('unknown interval method')
    pivots = []
    for value, var in zip(distribution, variances):
        if var is None:
            continue
        if var > 0:
            pivots.append((value-point)/math.sqrt(var))
        elif abs(value-point) <= 1e-14:
            pivots.append(0.)
        else:
            pivots.append(math.copysign(math.inf, value-point))
    if not pivots or variance is None:
        return [-1., 1.], 0
    # Percentile indices use retained Bprime; pivots MUST NOT be clipped.
    pivots.sort()
    n = len(pivots)
    low, high = pivots[math.floor(.025*n)], pivots[min(n-1, math.ceil(.975*n)-1)]
    se = math.sqrt(variance)
    def endpoint(pivot):
        return -math.copysign(1., pivot) if math.isinf(pivot) else max(-1., min(1., point-pivot*se))
    return [endpoint(high), endpoint(low)], n


def verdict(point_low, upper, lower, adequate, safety_counts, integrity=True):
    if not integrity:
        return 'inconclusive'
    if safety_counts['violating']:
        return 'not supported'
    if not adequate:
        return 'inconclusive'
    if upper < .10:
        return 'not supported'
    if safety_counts['unresolved']:
        return 'inconclusive'
    return 'support' if point_low >= .10 and lower > 0 else 'inconclusive'


def analyse(frame, replicates=1999, seed=169, integrity=True, include_conditional=True, method='studentized'):
    if replicates < 99:
        raise ValueError('at least 99 replicates required')
    if method not in ('percentile', 'studentized'):
        raise ValueError('unknown interval method')
    frame = sorted(frame, key=lambda row: (row.repo, row.run, row.issue))
    validate(frame)
    integrity = integrity and delivery_integrity(frame)
    repos = sorted({row.repo for row in frame})
    rng = random.Random(seed)
    choices, detail = [], {}
    adequate = len(repos) == 3
    for repo in repos:
        complete = [row for row in frame if row.repo == repo]
        rows = [row for row in complete if row.selected]
        runs = sorted({row.run for row in complete})
        draws = [rng.choices(range(len(runs)), k=max(1, len(runs)-1)) for _ in range(replicates)]
        vertices, seen = [], set()
        for bits in product((0, 1), repeat=3):
            fitted = fit(rows, dict(zip((-1, 0, 1), bits)))
            if fitted is None:
                continue
            identity = tuple((run, *values) for run, values in sorted(fitted['by_run'].items()))
            if identity in seen:
                continue
            seen.add(identity)
            values = [fitted['by_run'].get(run, (0., 0., 0.)) for run in runs]
            distribution = cluster_distribution(fitted, runs, draws)
            variances = [ratio_variance([values[i] for i in indices]) for indices in draws] if method == 'studentized' else [None]*replicates
            vertices.append((fitted, distribution, ratio_variance(values), variances))
        known_positive = sum(row.truth == 1 for row in rows)
        missing = sum(1/row.pi for row in rows if row.truth is None)
        selected_total = sum(1/row.pi for row in rows)
        ok = bool(vertices) and known_positive >= 60 and len(runs) >= 20 and selected_total > 0 and missing/selected_total <= .05
        adequate = adequate and ok
        detail[repo] = {'audit_selected': len(rows), 'known_positive': known_positive,
                        'unmeasured': sum(row.truth is None for row in rows),
                        'weighted_missing_share': missing/selected_total if selected_total else None,
                        'adequate': ok, 'vertices': len(vertices),
                        'gain_range': [min(v[0]['gain'] for v in vertices), max(v[0]['gain'] for v in vertices)] if vertices else None}
        choices.append(vertices)
    counts = safety(frame)
    if not repos or any(not values for values in choices):
        return {'verdict': verdict(0, 1, -1, False, counts, integrity), 'adequate': False,
                'reason': 'no positive denominator', 'safety': counts, 'repositories': detail}
    point_low = sum(min(v[0]['gain'] for v in values) for values in choices)/len(repos)
    point_high = sum(max(v[0]['gain'] for v in values) for values in choices)/len(repos)
    intervals, omitted = [], 0
    conditional = []
    for combination in product(*choices):
        distribution, variances = [], []
        for index in range(replicates):
            values = [pair[1][index] for pair in combination]
            if any(v is None for v in values):
                continue
            distribution.append(sum(values)/len(repos))
            parts = [pair[3][index] for pair in combination]
            variances.append(sum(parts)/len(repos)**2 if all(v is not None for v in parts) else None)
        point = sum(pair[0]['gain'] for pair in combination)/len(repos)
        parts = [pair[2] for pair in combination]
        variance = sum(parts)/len(repos)**2 if all(v is not None for v in parts) else None
        ci, retained = interval_from_distribution(point, variance, distribution, variances, method)
        omitted = max(omitted, replicates-retained)
        if retained:
            intervals.append(ci)
        if include_conditional:
            conditional.append(multiplier(point, [x/len(repos) for pair in combination for x in pair[0]['influence']], replicates, seed))
    adequate = adequate and omitted/replicates <= .05 and bool(intervals)
    ci = [min(v[0] for v in intervals), max(v[1] for v in intervals)] if intervals else [-1., 1.]
    return {'verdict': verdict(point_low, ci[1], ci[0], adequate, counts, integrity),
            'gain_range': [point_low, point_high], 'interval': ci,
            'conditional_interval': [min(v[0] for v in conditional), max(v[1] for v in conditional)] if conditional else None,
            'vertex_combinations': len(intervals), 'zero_denominator_replicates': omitted,
            'safety': counts, 'adequate': adequate, 'repositories': detail}




def wilson(success, attempts):
    z = 1.959963984540054
    p = success/attempts
    denominator = 1+z*z/attempts
    midpoint = (p+z*z/(2*attempts))/denominator
    radius = z*math.sqrt(p*(1-p)/attempts+z*z/(4*attempts*attempts))/denominator
    return [max(0., midpoint-radius), min(1., midpoint+radius)]


def select_runs(repository, commits):
    """Preregistered deterministic SRS ordering; no redraw after size gates."""
    if len(commits) < 20 or len(set(commits)) != len(commits):
        raise ValueError('insufficient or duplicate eligible runs')
    def key(commit):
        message = (repository+'\0'+commit).encode()
        return hmac.new(b'169-x1-runs', message, hashlib.sha256).digest(), commit
    return sorted(commits, key=key)[:20]


def pending_uniforms(frame, stored_keys):
    """DrawAudit consumes no RNG value for already persisted rows."""
    return [audit_uniform(row.repo, row.run, row.issue) for row in frame
            if (row.repo, row.issue) not in stored_keys]


STUDY_REASONS = {'study_latch', 'jev_clock', 'disabled', 'key_missing',
                 'config_invalid', 'threshold_invalid', 'annotations not persisted'}
FAILURE_REASONS = {'network', 'timeout', 'http', 'non-retryable', 'rate-limited',
                   'schema-invalid', 'model-mismatch', 'unsafe-to-persist'}
PRODUCT_REASONS = {'annotated', 'call cap reached', 'cost cap reached',
                   'billed cost unknown'} | FAILURE_REASONS


def delivery_integrity(frame):
    if any(row.reason in STUDY_REASONS for row in frame):
        return False
    for repo in {row.repo for row in frame}:
        attempted = [row for row in frame if row.repo == repo and row.reason in
                     ({'annotated'} | FAILURE_REASONS)]
        failures = sum(row.reason in FAILURE_REASONS for row in attempted)
        if attempted and failures/len(attempted) > .10:
            return False
    return True


def zero_tail(items, count):
    if count < 0:
        raise ValueError('negative tail count')
    return items[-count:] if count else []


def synthetic(gain=.20, seed=9, prevalence=.20, heterogeneity=0., overlap=0.,
              run_count=20, sizes=(50,), cap=2000, failure=0.,
              scanner_true=False, rank_structure='uniform', size_dependence=0.):
    """Fresh deterministic construction using the fixed-key population helpers."""
    from fixed_population import prepare_population, evaluate_population
    prepared = prepare_population(seed=seed, prevalence=prevalence,
                                  heterogeneity=heterogeneity, overlap=overlap,
                                  run_count=run_count, sizes=sizes, cap=cap,
                                  failure=failure, scanner_true=scanner_true,
                                  rank_structure=rank_structure, size_dependence=size_dependence)
    return evaluate_population(prepared, gain)


def bootstrap_totals(repositories, draws, method='studentized'):
    """Paired ratio bootstrap; studentization uses the same run draws in both arms."""
    estimates, variances, samples, sample_variances = [], [], [], []
    for values, repo_draws in zip(repositories, draws):
        total = sum(v[0] for v in values)
        if total <= 0:
            return None, [-1., 1.], len(repo_draws)
        estimates.append(sum(v[1]-v[2] for v in values)/total)
        variances.append(ratio_variance(values))
        fitted = {'by_run': {str(i): value for i, value in enumerate(values)}}
        samples.append(cluster_distribution(fitted, [str(i) for i in range(len(values))], repo_draws))
        sample_variances.append([ratio_variance([values[i] for i in indices]) for indices in repo_draws] if method == 'studentized' else [None]*len(repo_draws))
    distribution, replicate_variances = [], []
    count = len(repositories)
    for index, combination in enumerate(zip(*samples)):
        if all(value is not None for value in combination):
            distribution.append(sum(combination)/count)
            parts = [values[index] for values in sample_variances]
            replicate_variances.append(sum(parts)/count**2 if all(v is not None for v in parts) else None)
    point = sum(estimates)/count
    variance = sum(variances)/count**2 if all(v is not None for v in variances) else None
    interval, retained = interval_from_distribution(point, variance, distribution, replicate_variances, method)
    return point, interval, len(draws[0])-retained


def provenance(paths):
    import platform
    from pathlib import Path
    import sys
    if sys.flags.optimize:
        raise RuntimeError('X1 numerical validation refuses optimized Python')
    return {'python_version': platform.python_version(),
            'python_optimize': sys.flags.optimize,
            'script_sha256': {str(path): hashlib.sha256(Path(path).read_bytes()).hexdigest()
                              for path in paths}}

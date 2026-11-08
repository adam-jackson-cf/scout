#!/usr/bin/env python3
"""Synthetic outer-SRS planning OC. No real inventory or model input accepted."""
from collections import Counter, defaultdict
from dataclasses import replace
from itertools import product
import argparse
import json
from pathlib import Path
import random
import time

from ranking import bootstrap_totals, fit, provenance, synthetic, verdict, wilson


def aggregate(frame):
    grouped = defaultdict(lambda: defaultdict(Counter))
    for row in frame:
        key = (row.pi, row.truth, int(row.jev <= row.k), int(row.baseline <= row.k))
        grouped[row.repo][row.run][key] += 1
    return {repo: dict(runs) for repo, runs in sorted(grouped.items())}


def audit_totals(groups, rng):
    t = j = b = known = 0.
    for (pi, truth, top_j, top_b), count in groups.items():
        selected = count if pi == 1 else rng.binomialvariate(count, pi)
        known += selected*truth
        weight = selected*truth/pi
        t += weight
        j += weight*top_j
        b += weight*top_b
    return (t, j, b), known


def run(trials=100, replicates=199):
    started = time.monotonic()
    scenarios = []
    # Bootstrap index streams are independent of outer SRS/audit and common
    # across scenarios. The full protocol uses B=1999 and >=10000 trials.
    streams = []
    for trial in range(trials):
        rng = random.Random(269000+trial)
        streams.append([[rng.choices(range(20), k=19) for _ in range(replicates)] for _ in range(3)])
    # Distinct sizes exercise cap-order rows, unlike rev2's all-50-unit runs.
    sizes = (50, 100, 500, 1000, 2000)
    for prevalence, sd, overlap, gain, cap, failure, scanner_true in product(
            (.10, .20, .40, .55), (0., .10), (0., .5),
            (0., .05, .08, .10, .12, .20), (2000, 850, 400), (0., .02), (False, True)):
        seed = 900
        population = synthetic(gain, seed, prevalence, sd, overlap, 40, sizes, cap, failure, scanner_true)
        groups = aggregate(population)
        by_repo = {}
        for row in population:
            by_repo.setdefault(row.repo, []).append(replace(row, pi=1.))
        truth = sum(fit(rows, {-1: 0, 0: 0, 1: 0})['gain'] for rows in by_repo.values())/3
        run_rows = defaultdict(list)
        for row in population:
            run_rows[(row.repo, row.run)].append(row)
        summary = {key: {'bottom': sum(r.below for r in rows),
                         'bottom_positive': sum(r.below and r.truth == 1 for r in rows),
                         'scanner_violation': sum(r.below and r.truth == 1 and r.severity_critical for r in rows)}
                   for key, rows in run_rows.items()}
        outcomes = {(u, c): Counter() for u, c in product((0., .0001, .0005, .001, .005, .02), (0., .001, .01, .05))}
        miss_lower = miss_upper = inadequate = 0
        cap_excess = sum(row.baseline > cap for row in population)/len(population)
        for trial in range(trials):
            rng = random.Random(169000+trial)
            repository_totals, selected_summary = [], []
            adequate = True
            for repo, runs in groups.items():
                selected = sorted(rng.sample(sorted(runs), 20))
                values, known = [], 0
                for run_id in selected:
                    value, count = audit_totals(runs[run_id], rng)
                    values.append(value)
                    known += count
                    selected_summary.append(summary[(repo, run_id)])
                adequate = adequate and known >= 60
                repository_totals.append(values)
            point, interval, omitted = bootstrap_totals(repository_totals, streams[trial])
            adequate = adequate and point is not None and omitted/replicates <= .05
            inadequate += not adequate
            if point is None:
                continue
            miss_lower += interval[0] > truth+1e-12
            miss_upper += interval[1] < truth-1e-12
            bottom = sum(item['bottom'] for item in selected_summary)
            bottom_positive = sum(item['bottom_positive'] for item in selected_summary)
            scanner_violations = sum(item['scanner_violation'] for item in selected_summary)
            for (unresolved_rate, critical_share), counts in outcomes.items():
                # Per-case Bernoulli events, sampled via exact binomial counts.
                # This safety diagnostic conditions on known retrieval labels;
                # final missing-label OC must jointly apply unknown-Y vertices.
                unresolved = rng.binomialvariate(bottom, unresolved_rate)
                critical = rng.binomialvariate(bottom_positive, critical_share)
                state = {'violating': scanner_violations+critical, 'unresolved': unresolved}
                decision = verdict(point, interval[1], interval[0], adequate, state)
                counts[decision] += 1
                counts['veto'] += bool(state['violating'])
                counts['safety_unmeasured'] += bool(unresolved)
        for (u, c), counts in outcomes.items():
            scenarios.append({'prevalence': prevalence, 'gain_sd': sd, 'overlap': overlap,
                              'requested_gain': gain, 'population_realised_gain': truth,
                              'cap_M': cap, 'random_failure_rate': failure,
                              'scanner_critical_true_at_prevalence': scanner_true,
                              'unresolved_rate': u, 'critical_positive_share': c,
                              'trials': trials, 'counts': dict(counts), 'inadequate_trials': inadequate,
                              'unit_share_beyond_cap': cap_excess,
                              'support_MC95': wilson(counts['support'], trials),
                              'veto_MC95': wilson(counts['veto'], trials),
                              'safety_unmeasured_MC95': wilson(counts['safety_unmeasured'], trials),
                              'lower_miss_count': miss_lower, 'upper_miss_count': miss_upper,
                              'lower_miss_MC95': wilson(miss_lower, trials),
                              'upper_miss_MC95': wilson(miss_upper, trials),
                              'tail_criterion_met_at_pilot_precision':
                                  wilson(miss_lower, trials)[1] <= .035 and wilson(miss_upper, trials)[1] <= .035,
                              'fails_at_pilot_precision':
                                  wilson(miss_lower, trials)[0] > .035 or wilson(miss_upper, trials)[0] > .035})
    elapsed = time.monotonic()-started
    return {'schema': 'x1-outer-srs-planning/rev4-legacy-grid', 'synthetic_only': True,
            'full_inventory_OC': False, 'safety_diagnostic_conditional_on_known_retrieval_labels': True,
            'population_runs_per_repository': 40, 'sampled_runs_per_repository': 20,
            'population_size_cycle': list(sizes), 'bootstrap_m': 19,
            'trials_per_cell': trials, 'bootstrap_replicates': replicates,
            'conditional_multiplier_excluded': True, 'elapsed_seconds': elapsed,
            'provenance': provenance([Path(__file__).with_name('ranking.py'), Path(__file__)]),
            'scenario_count': len(scenarios), 'scenarios': scenarios}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--trials', type=int, default=100)
    parser.add_argument('--replicates', type=int, default=199)
    parser.add_argument('--output', required=True, type=Path)
    args = parser.parse_args()
    result = run(args.trials, args.replicates)
    with args.output.open('x', encoding='utf-8') as stream:
        json.dump(result, stream, sort_keys=True, indent=2)
        stream.write('\n')
    print(json.dumps({key: result[key] for key in ('scenario_count', 'trials_per_cell', 'elapsed_seconds')}))


if __name__ == '__main__':
    main()

#!/usr/bin/env python3
"""Reproduce synthetic historical-pilot summaries and measured resource projections.

Run from the repository root; output is create-once. No models or real data.
"""
import argparse
import hashlib
import json
import math
from pathlib import Path
import time

ROOT = Path(__file__).resolve().parents[3]
HERE = Path(__file__).resolve().parent


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def relative(path):
    return str(path.resolve().relative_to(ROOT))


def wilson(k, n):
    z = 1.959963984540054
    p = k/n
    centre = (p+z*z/(2*n))/(1+z*z/n)
    half = z*math.sqrt(p*(1-p)/n+z*z/(4*n*n))/(1+z*z/n)
    return [centre-half, centre+half]


def pilot_summary(source):
    data = json.loads(source.read_text())
    selected = [row for row in data['scenarios']
                if not row['scanner_critical_true_at_prevalence']
                and row['unresolved_rate'] == 0 and row['critical_positive_share'] == 0]
    blocks = []
    for sd in (0., .10):
        for M in (2000, 850, 400):
            for failure in (0., .02):
                rows = [row for row in selected if row['cap_M'] == M
                        and row['gain_sd'] == sd and row['random_failure_rate'] == failure]
                n = sum(row['trials'] for row in rows)
                lower = sum(row['lower_miss_count'] for row in rows)
                upper = sum(row['upper_miss_count'] for row in rows)
                effect = [row for row in rows if row['requested_gain'] == .20]
                gains = [row['population_realised_gain'] for row in effect]
                powers = [row['counts'].get('support', 0)/row['trials'] for row in effect]
                blocks.append({'M': M, 'sd': sd, 'failure': failure, 'denominator': n,
                               'lower': lower, 'lower_rate': lower/n,
                               'lower_MC95_correlated_seeds': wilson(lower, n),
                               'upper': upper, 'upper_rate': upper/n,
                               'upper_MC95_correlated_seeds': wilson(upper, n),
                               'requested_point2_realised_gain_range': [min(gains), max(gains)],
                               'requested_point2_power_range': [min(powers), max(powers)]})
    tail = sum(math.comb(100, k)*.025**k*.975**(100-k) for k in range(8, 101))
    return {'schema': 'x1-historical-pilot-summary/rev5', 'source': relative(source),
            'source_sha256': digest(source),
            'population_selection': 'scanner false, unresolved zero, critical share zero; safety dimensions deduplicated',
            'shared_seeds_across_cells': True,
            'Wilson_intervals': 'descriptive nominal binomial intervals only; shared cross-cell seeds invalidate pooled independence precision',
            'nominal_expected_tails_at_least_8_among_1152': 1152*tail, 'blocks': blocks}


def calibration_overhead():
    import fixed_population as kernel
    sources = {relative(HERE/name): digest(HERE/name)
               for name in ('fixed_population.py', 'ranking.py', 'reproduce_reports.py')}
    measurements = []
    for L in (40, 400):
        sizes = (2000,)*(L//4)+(500,)*(L//2)+(1000,)*(L//4)
        prepared = kernel.prepare_population(seed=202611080000+L, prevalence=.55,
                                             heterogeneity=.10, overlap=.5, run_count=L,
                                             sizes=sizes, cap=400, failure=.02,
                                             rank_structure='linear-gradient', size_dependence=.05)
        original = kernel.evaluate_population
        calls = []

        def measured_evaluate(population, gain):
            started = time.process_time()
            frame = original(population, gain)
            calls.append(time.process_time()-started)
            return frame

        kernel.evaluate_population = measured_evaluate
        cpu, wall = time.process_time(), time.perf_counter()
        try:
            result = kernel.calibrate_population(prepared, .20)
        finally:
            kernel.evaluate_population = original
        total_cpu, elapsed = time.process_time()-cpu, time.perf_counter()-wall
        units = 3*sum(sizes)
        # Residual includes candidate construction, realised-truth scans, search
        # and wrapper bookkeeping. Scale the entire residual to 22 evaluations,
        # conservatively scaling the one-time candidate cost too.
        residual = max(0., total_cpu-sum(calls))
        measurements.append({'L': L, 'units': units, 'full_calibration_cpu_seconds': total_cpu,
                             'elapsed_seconds': elapsed, 'frame_evaluations': len(calls),
                             'frame_evaluation_cpu_seconds': calls,
                             'maximum_frame_cpu_seconds_per_unit': max(calls)/units,
                             'non_frame_cpu_seconds': residual,
                             'conservative_22_evaluation_overhead_cpu_seconds_per_unit':
                                 residual*max(1., 22/len(calls))/units,
                             'target': result['target'], 'realised': result['realised'],
                             'reachable': result['reachable']})
        del result, prepared
    if sources != {path: digest(ROOT/path) for path in sources}:
        raise RuntimeError('calibration source changed; output not published')
    return {'schema': 'x1-calibration-overhead/rev5', 'synthetic_only': True,
            'source_sha256': sources, 'measurements': measurements,
            'scope': 'Full exact fixed-key calibration API on genuine independent populations. Nested frame evaluation CPU is separated from breakpoint construction, truth aggregation and search; no population reconstruction per evaluation.'}


def projection(source, p_gate, inventory_runs, mean_size):
    data = json.loads(source.read_text())
    calibration_source = HERE/'calibration-overhead-rev5.json'
    calibration = json.loads(calibration_source.read_text())
    worst = {kind: max(t['cpu_seconds'] for run in data['executions'] for cell in run['cells']
                      if cell['cell']['missing_labels'] == missing for t in cell['trial_timing'])
             for kind, missing in (('complete', False), ('full512', True))}
    prep = max(row['construction']['cpu_seconds_per_unit'] for row in data['population_measurements'])
    reevaluate = max(
        max(row['max_reevaluation_cpu_seconds_per_unit'] for row in data['population_measurements']),
        max(row['maximum_frame_cpu_seconds_per_unit'] for row in calibration['measurements']))
    calibration_extra = max(row['conservative_22_evaluation_overhead_cpu_seconds_per_unit']
                            for row in calibration['measurements'])
    units = 3*inventory_runs*mean_size
    attempts = math.ceil(4000/p_gate) if p_gate else None
    attempt_factor = attempts/2000 if attempts else None
    counts = {'requested_coverage': 16128, 'calibrated_references': 1792,
              'complete_total': 17920, 'missing': 252,
              'safety_profiles': 336, 'safety_outcomes': 8064,
              'safety_complete_interval_groups': 336, 'safety_full512_interval_groups': 1680}
    constructions = {'base': (17920+252)*5, 'safety_profiles_shared_across_24_states': 336*5}
    evaluations = {'requested_base': 16128*5,
                   'calibration_endpoints_and_bisections': (1792+252+336)*5*22,
                   'canonical_state': 8064*5}
    overhead = 2
    preparation_hours = {key: count*units*prep*overhead/3600 for key, count in constructions.items()}
    reevaluation_hours = {key: count*units*reevaluate*overhead/3600 for key, count in evaluations.items()}
    reevaluation_hours['calibration_non_frame_overhead'] = (
        (1792+252+336)*5*units*calibration_extra*overhead/3600)
    trials = {'binding_complete': (17920, 'complete'), 'binding_missing': (252, 'full512'),
              'descriptive_safety_complete': (336, 'complete'),
              'descriptive_safety_full512': (1680, 'full512')}
    trial_hours = {key: count*10000*worst[kind]*attempt_factor*overhead/3600
                   if attempt_factor else None for key, (count, kind) in trials.items()}
    total = (sum(preparation_hours.values())+sum(reevaluation_hours.values())+sum(trial_hours.values())
             if attempt_factor else None)
    binding_hours = (
        preparation_hours['base']+reevaluation_hours['requested_base']
        +(reevaluation_hours['calibration_endpoints_and_bisections']
          +reevaluation_hours['calibration_non_frame_overhead'])*(1792+252)/(1792+252+336)
        +trial_hours['binding_complete']+trial_hours['binding_missing']
        if total is not None else None)
    return {'schema': 'x1-authoritative-OC-resource-projection/rev5',
            'synthetic_only': True, 'binding_oc_gate': False,
            'status': 'projection_not_observed_full_grid_runtime',
            'benchmark': relative(source), 'benchmark_sha256': digest(source), 'counts': counts,
            'calibration_overhead': relative(calibration_source),
            'calibration_overhead_sha256': digest(calibration_source),
            'measured_inputs': {'maximum_trial_cpu_seconds': worst,
                                'construction_cpu_seconds_per_unit': prep,
                                'reevaluation_cpu_seconds_per_unit': reevaluate,
                                'calibration_non_frame_overhead_cpu_seconds_per_unit': calibration_extra},
            'population_extrapolation': {'repositories': 3, 'eligible_runs_per_repository': inventory_runs,
                                         'mean_run_size': mean_size, 'units_per_population': units,
                                         'maximum_registered_L': 2000,
                                         'method': 'maximum measured per-unit CPU across genuine L40/400 populations; linear extrapolation, not observed inventory runtime'},
            'gate_attempt_contract': {'assumed_projection_p_gate': p_gate,
                                      'assumption_not_empirical': True,
                                      'observed_synthetic_gate_estimates': [
                                          {'L': row['L'], **row['size_only_sampling_gate']}
                                          for row in data['population_measurements']],
                                      'locked_inventory_keyed_SRS20_gate_draws': 100000,
                                      'attempt_cap_per_population': attempts,
                                      'formula': 'ceil(4000/p_gate)',
                                      'eligible_trials_per_population': 2000,
                                      'projected_trial_attempt_multiplier': attempt_factor,
                                      'zero_gate_probability': 'not testable',
                                      'locked_inventory_requirement': 'Remeasure p_gate using 100000 keyed size-only SRS20 draws; remeasure independent construction, fixed-key calibration and complete/full512 trial CPU at locked sizes up to L2000; recompute this projection before full authority.'},
            'execution_overhead_multiplier': overhead,
            'construction_counts': constructions, 'reevaluation_counts': evaluations,
            'fixed_key_calibration': 'one construction per population, at most 22 reevaluations; never 22 constructions',
            'canonical_bound': 'one independent construction per safety profile/population, shared across six unresolved and four critical axes; conservatively charge 24 full fixed-key reevaluations per profile/population',
            'projected_core_hours': {'trials': trial_hours, 'construction': preparation_hours,
                                      'reevaluation': reevaluation_hours, 'total': total},
            'projected_wall_days_at_8_workers': total/(8*24) if total is not None else None,
            'binding_critical_path_core_hours': binding_hours,
            'binding_critical_path_wall_days_at_8_workers': binding_hours/(8*24) if binding_hours is not None else None,
            'descriptive_safety_core_hours': total-binding_hours if total is not None else None,
            'caps': {'workers': 8, 'core_hours': 2880, 'wall_days': 15},
            'fits_caps': total is not None and total <= 2880,
            'safety_schedule': 'Start after binding rows, concurrently with labelling; descriptive, no efficacy gate. Finish before canonical gold or report unmeasured. All safety CPU remains inside the full OC resource cap.',
            'authority': 'No preparation or full-study authority. Failed caps require reviewed revision before preparation; no automatic grid reduction.'}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('report', choices=('pilot-summary', 'projection', 'calibration-overhead'))
    parser.add_argument('--input', type=Path)
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--p-gate', type=float, default=.5)
    parser.add_argument('--inventory-runs', type=int, default=400)
    parser.add_argument('--mean-run-size', type=float, default=1000.)
    args = parser.parse_args()
    if args.output.exists():
        raise FileExistsError('report evidence is create-once')
    if not 0 <= args.p_gate <= 1 or not 20 <= args.inventory_runs <= 2000 or not 1 <= args.mean_run_size <= 2000:
        parser.error('p-gate in [0,1], inventory-runs in [20,2000], mean-run-size in [1,2000] required')
    source = args.input or HERE/('outer-results.json' if args.report == 'pilot-summary' else 'compute-benchmark-rev5.json')
    if args.report == 'pilot-summary':
        result = pilot_summary(source)
    elif args.report == 'calibration-overhead':
        result = calibration_overhead()
    else:
        result = projection(source, args.p_gate, args.inventory_runs, args.mean_run_size)
    result['reporter_sha256'] = digest(Path(__file__))
    result['reproduce_cli'] = ('source .tools/activate && python evaluation/x1-triage-ranking/validation/reproduce_reports.py '
                               f'{args.report} --output {relative(args.output)}'
                               + (f' --input {relative(source)}' if args.report != 'calibration-overhead' else '')
                               + (f' --p-gate {args.p_gate} --inventory-runs {args.inventory_runs} --mean-run-size {args.mean_run_size}'
                                  if args.report == 'projection' else ''))
    with args.output.open('x') as stream:
        json.dump(result, stream, indent=2, sort_keys=True, allow_nan=False)
        stream.write('\n')
    print(json.dumps({'output': relative(args.output), 'fits_caps': result.get('fits_caps')}))


if __name__ == '__main__':
    main()

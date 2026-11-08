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
import platform
import resource
import sys
from concurrent.futures import ProcessPoolExecutor
from types import SimpleNamespace

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


def peak_rss_bytes():
    return int(resource.getrusage(resource.RUSAGE_SELF).ru_maxrss*
               (1 if sys.platform == 'darwin' else 1024))


def calibration_measurement(L):
    import fixed_population as kernel
    sizes = (2000,)*(L//4)+(500,)*(L//2)+(1000,)*(L//4)
    prepared = kernel.prepare_population(
        seed=202611080000+L, prevalence=.55, heterogeneity=.10, overlap=.5,
        run_count=L, sizes=sizes, cap=400, failure=.02,
        rank_structure='linear-gradient', size_dependence=.05)
    original_frame, original_truth = kernel.evaluate_population, kernel._population_gain
    original_candidates = kernel._allocation_candidates
    frame_calls, truth_calls, candidate_calls = [], [], []

    def measured(function, calls):
        def wrapper(*args):
            started = time.process_time()
            value = function(*args)
            calls.append(time.process_time()-started)
            return value
        return wrapper

    kernel.evaluate_population = measured(original_frame, frame_calls)
    kernel._population_gain = measured(original_truth, truth_calls)
    kernel._allocation_candidates = measured(original_candidates, candidate_calls)
    cpu, wall = time.process_time(), time.perf_counter()
    try:
        result = kernel.calibrate_population(prepared, .20)
    finally:
        kernel.evaluate_population, kernel._population_gain = original_frame, original_truth
        kernel._allocation_candidates = original_candidates
    total_cpu, elapsed = time.process_time()-cpu, time.perf_counter()-wall
    # <=20 binary-search calls plus the mandatory returned-frame truth assertion.
    if len(frame_calls) != 1 or not 0 < len(truth_calls) <= 21 or len(candidate_calls) != 1:
        raise AssertionError('calibration must use <=21 truth calls, one frame, one enumeration')
    units = 3*sum(sizes)
    # Remove measured enumeration before applying any per-unit extrapolation.
    nonenumeration = total_cpu-sum(candidate_calls)+(21-len(truth_calls))*max(truth_calls)
    return {'L': L, 'units': units, 'full_calibration_cpu_seconds': total_cpu,
            'elapsed_seconds': elapsed, 'frame_evaluations': len(frame_calls),
            'frame_evaluation_cpu_seconds': frame_calls,
            'maximum_frame_cpu_seconds_per_unit': max(frame_calls)/units,
            'truth_only_evaluations': len(truth_calls), 'truth_only_cpu_seconds': truth_calls,
            'candidate_enumeration_cpu_seconds': sum(candidate_calls),
            'nonenumeration_bookkeeping_cpu_seconds':
                total_cpu-sum(candidate_calls)-sum(frame_calls)-sum(truth_calls),
            'conservative_21_truth_one_frame_no_enumeration_cpu_seconds': nonenumeration,
            'conservative_nonenumeration_cpu_seconds_per_unit': nonenumeration/units,
            'peak_rss_bytes': peak_rss_bytes(),
            'rss_scope': 'fresh spawned process lifetime high-water; includes construction and calibration',
            'target': result['target'], 'realised': result['realised'], 'reachable': result['reachable']}


def worst_case_enumeration():
    import fixed_population as kernel
    # These are exactly the fields read by _allocation_candidates. Population
    # allocation and random keys cannot affect this deterministic enumeration.
    runs = tuple(SimpleNamespace(n=n, positives=round(n*.55), k=math.ceil(.30*n),
                                 size_offset=.05*(n-20)/(2000-20))
                 for n in range(20, 2001))
    prepared = SimpleNamespace(runs=runs)
    cpu, wall = time.process_time(), time.perf_counter()
    candidates = kernel._allocation_candidates(prepared)
    return {'sizes': {'minimum': 20, 'maximum': 2000, 'step': 1, 'count': len(runs)},
            'prevalence': .55, 'size_dependence': .05,
            'candidate_count': len(candidates),
            'cpu_seconds': time.process_time()-cpu,
            'elapsed_seconds': time.perf_counter()-wall,
            'peak_rss_bytes': peak_rss_bytes(),
            'scope': 'exact production enumeration on every size; structural fields only, no per-unit frame work',
            'charge_per_calibrated_population': True, 'calibrated_populations': 11900,
            'cache_assumed': False}


def calibration_overhead():
    sources = {relative(HERE/name): digest(HERE/name)
               for name in ('fixed_population.py', 'ranking.py', 'reproduce_reports.py')}
    with ProcessPoolExecutor(max_workers=1, max_tasks_per_child=1) as pool:
        measurements = list(pool.map(calibration_measurement, (40, 400)))
        enumeration = pool.submit(worst_case_enumeration).result()
    if sources != {path: digest(ROOT/path) for path in sources}:
        raise RuntimeError('calibration source changed; output not published')
    return {'schema': 'x1-calibration-overhead/rev7', 'synthetic_only': True,
            'source_sha256': sources, 'measurements': measurements,
            'worst_case_candidate_enumeration': enumeration,
            'python_version': platform.python_version(),
            'python_optimize': sys.flags.optimize,
            'python_executable': '.tools/venv/bin/python',
            'python_executable_sha256': digest(Path(sys.executable).resolve()),
            'scope': 'Separate worst-case enumeration per calibrated population; per-unit API CPU excludes its measured enumeration and pads to 21 truth calls including returned-frame assertion. No double-counted enumeration.'}


def source_closure(benchmark, calibration):
    sources = {**benchmark['source_sha256'], **calibration['source_sha256']}
    if any(benchmark['source_sha256'].get(path, expected) != expected
           for path, expected in calibration['source_sha256'].items()):
        raise RuntimeError('benchmark and calibration source bytes differ')
    # Revision 8 changes only reporter arithmetic/provenance; timing inputs
    # retain their executed revision-7 reporter bytes in the sealed archive.
    reporter = relative(Path(__file__))
    if any(digest(ROOT/path) != expected for path, expected in sources.items() if path != reporter):
        raise RuntimeError('executed numerical sources changed after measurement')
    snapshot_path = HERE/'rev7-executed-sources.json'
    snapshot = json.loads(snapshot_path.read_text())
    if any(snapshot['files'][path]['sha256'] != expected or
           hashlib.sha256(snapshot['files'][path]['content'].encode()).hexdigest() != expected
           for path, expected in sources.items()):
        raise RuntimeError('pre-execution source snapshot differs from measured bytes')
    return {'snapshot': relative(snapshot_path), 'snapshot_sha256': digest(snapshot_path),
            'executed_source_sha256': sources, 'exact_bytes_identical': True,
            'scope': 'measured source bytes match revision-7 archive; arithmetic-only reporter is separately hashed, no timing rerun',
            'python_version': benchmark['python_version'],
            'python_executable_sha256': benchmark['python_executable_sha256']}


def projection(source, p_gate, inventory_runs, mean_size):
    data = json.loads(source.read_text())
    calibration_source = HERE/'calibration-overhead-rev7.json'
    calibration = json.loads(calibration_source.read_text())
    closure = source_closure(data, calibration)
    worst = {kind: max(t['cpu_seconds'] for run in data['executions'] for cell in run['cells']
                      if cell['cell']['missing_labels'] == missing for t in cell['trial_timing'])
             for kind, missing in (('complete', False), ('full512', True))}
    prep = max(row['construction']['cpu_seconds_per_unit'] for row in data['population_measurements'])
    reevaluate = max(
        max(row['max_reevaluation_cpu_seconds_per_unit'] for row in data['population_measurements']),
        max(row['maximum_frame_cpu_seconds_per_unit'] for row in calibration['measurements']))
    calibration_cpu = max(row['conservative_nonenumeration_cpu_seconds_per_unit']
                          for row in calibration['measurements'])
    enumeration = calibration['worst_case_candidate_enumeration']['cpu_seconds']
    gate_cpu = max(row['size_only_sampling_gate']['cpu_seconds']/row['size_only_sampling_gate']['draws']
                   for row in data['population_measurements'])
    rss = {'construction': max(row['construction']['peak_rss_bytes']
                               for row in data['population_measurements'] if row['L'] == 400),
           'calibration': max(row['peak_rss_bytes'] for row in calibration['measurements'] if row['L'] == 400),
           'trials': max(cell['peak_rss_bytes'] for run in data['executions'] for cell in run['cells']
                         if cell['cell']['L'] == 400)}
    if any(not math.isfinite(value) or value <= 0
           for value in (prep, reevaluate, calibration_cpu, enumeration, gate_cpu, *worst.values(), *rss.values())):
        raise RuntimeError('unmeasured CPU/RSS: fail closed')
    units = 3*inventory_runs*mean_size
    attempts = math.ceil(4000/p_gate) if p_gate else None
    counts = {'requested_coverage': 16128, 'calibrated_references': 1792, 'complete_total': 17920,
              'missing': 252, 'safety_profiles': 336, 'safety_outcomes': 8064,
              'safety_complete_interval_groups': 336, 'safety_full512_interval_groups': 1680}
    constructions = {'base': (17920+252)*5, 'safety_profiles_shared_across_24_states': 336*5}
    evaluations = {'requested_base': 16128*5, 'canonical_state': 8064*5}
    calibration_populations = (1792+252+336)*5
    overhead = 2
    preparation_hours = {key: count*units*prep*overhead/3600 for key, count in constructions.items()}
    reevaluation_hours = {key: count*units*reevaluate*overhead/3600 for key, count in evaluations.items()}
    reevaluation_hours['truth_only_calibration_and_final_frame_excluding_enumeration'] = (
        calibration_populations*units*calibration_cpu*overhead/3600)
    enumeration_hours = calibration_populations*enumeration*overhead/3600
    trials = {'binding_complete': (17920, 'complete'), 'binding_missing': (252, 'full512'),
              'descriptive_safety_complete': (336, 'complete'), 'descriptive_safety_full512': (1680, 'full512')}
    # Safety stops at 2000 gate-passing executions; the 4000 guard is inert.
    trial_hours = {key: count*5*(2000 if key.startswith('descriptive') else 4000)*worst[kind]*overhead/3600
                   for key, (count, kind) in trials.items()}
    gate_hours = {key: count*5*attempts*gate_cpu*overhead/3600 if attempts else None
                  for key, (count, _) in trials.items()}
    fixed_hours = sum(trial_hours.values())+enumeration_hours
    per_unit_hours = (sum(constructions.values())*prep+sum(evaluations.values())*reevaluate
                      +calibration_populations*calibration_cpu)*overhead/3600
    total = fixed_hours+units*per_unit_hours+sum(gate_hours.values()) if attempts else None
    binding_hours = (preparation_hours['base']+reevaluation_hours['requested_base']
                     +(reevaluation_hours['truth_only_calibration_and_final_frame_excluding_enumeration']
                       +enumeration_hours)*(1792+252)/(1792+252+336)
                     +trial_hours['binding_complete']+trial_hours['binding_missing']
                     +gate_hours['binding_complete']+gate_hours['binding_missing']) if attempts else None
    # Integer attempt-cap envelope is exact: ceil(4000/p)<=floor(remaining/cost).
    gate_hours_per_attempt = sum(count*5 for count, _ in trials.values())*gate_cpu*overhead/3600
    max_attempts = math.floor((2880-fixed_hours-units*per_unit_hours)/gate_hours_per_attempt)
    minimum_p = 4000/max_attempts if max_attempts >= 4000 else None
    break_even_units = ((2880-fixed_hours-sum(gate_hours.values()))/per_unit_hours) if attempts else None
    binding_fraction = (1792+252)/(1792+252+336)
    binding_fixed = trial_hours['binding_complete']+trial_hours['binding_missing']+enumeration_hours*binding_fraction
    binding_per_unit = (constructions['base']*prep+evaluations['requested_base']*reevaluate+
                        calibration_populations*binding_fraction*calibration_cpu)*overhead/3600
    binding_gate_per_attempt = (17920+252)*5*gate_cpu*overhead/3600
    binding_max_attempts = math.floor((2880-binding_fixed-units*binding_per_unit)/binding_gate_per_attempt)
    binding_break_even = ((2880-binding_fixed-attempts*binding_gate_per_attempt)/binding_per_unit) if attempts else None
    small, large = sorted(calibration['measurements'], key=lambda row: row['units'])
    rss_slope = (large['peak_rss_bytes']-small['peak_rss_bytes'])/(large['units']-small['units'])
    rss_intercept = small['peak_rss_bytes']-rss_slope*small['units']
    host_memory = data['host_physical_memory_bytes']
    memory_reserve = 8*1024**3
    memory_budget = host_memory-memory_reserve
    memory_break_even = (memory_budget/8-rss_intercept)/rss_slope
    return {'schema': 'x1-authoritative-OC-resource-projection/rev8', 'synthetic_only': True,
            'binding_oc_gate': False, 'status': 'projection_not_observed_full_grid_runtime',
            'benchmark': relative(source), 'benchmark_sha256': digest(source),
            'calibration_overhead': relative(calibration_source),
            'calibration_overhead_sha256': digest(calibration_source), 'counts': counts,
            'source_closure': closure,
            'measured_inputs': {'maximum_trial_cpu_seconds': worst,
                                'construction_cpu_seconds_per_unit': prep,
                                'reevaluation_cpu_seconds_per_unit': reevaluate,
                                'calibration_excluding_enumeration_cpu_seconds_per_unit': calibration_cpu,
                                'worst_case_enumeration_cpu_seconds_per_calibrated_population': enumeration,
                                'sampling_gate_cpu_seconds_per_check': gate_cpu},
            'population_extrapolation': {'repositories': 3, 'eligible_runs_per_repository': inventory_runs,
                                         'mean_run_size': mean_size, 'units_per_population': units,
                                         'maximum_registered_L': 2000,
                                         'method': 'maximum measured per-unit CPU at independent L40/400; enumeration separately fixed per calibrated population'},
            'gate_attempt_contract': {'assumed_projection_p_gate': p_gate, 'assumption_not_empirical': True,
                                      'attempt_cap_per_population': attempts, 'formula': 'ceil(4000/p_gate)',
                                      'hard_trial_execution_cap_per_population': 4000,
                                      'trial_execution_definition': 'gate-passing attempts including inadequate audits',
                                      'eligible_trials_per_population': 2000,
                                      'projected_trial_execution_multiplier': 2,
                                      'exhaustion': 'unmeasured; cell fails',
                                      'zero_gate_probability': 'not testable',
                                      'locked_inventory_requirement': 'Remeasure p_gate with 100000 keyed checks, CPU and RSS at locked sizes; missing evidence or failed fit is not testable'},
            'execution_overhead_multiplier': overhead, 'construction_counts': constructions,
            'reevaluation_counts': evaluations, 'calibrated_populations': calibration_populations,
            'candidate_enumeration_accounting': 'Worst-case every size 20..2000, p=.55, a=.05 charged 11900 times; measured enumeration removed from per-unit calibration CPU before extrapolation',
            'projected_core_hours': {'trials': trial_hours, 'sampling_gate_checks': gate_hours,
                                      'construction': preparation_hours, 'reevaluation': reevaluation_hours,
                                      'candidate_enumeration': enumeration_hours, 'total': total},
            'break_even_envelope': {'budget_core_hours': 2880,
                                    'scope': 'total planning envelope; never a testability stop',
                                    'formula': 'fixed_hours + units_per_population*per_unit_hours + ceil(4000/p_gate)*gate_hours_per_attempt <=2880',
                                    'fixed_hours': fixed_hours, 'per_unit_hours': per_unit_hours,
                                    'gate_hours_per_attempt': gate_hours_per_attempt,
                                    'maximum_units_per_population_at_assumed_p_gate': break_even_units,
                                    'maximum_L_at_mean1000': break_even_units/3000 if break_even_units else None,
                                    'minimum_p_gate_at_assumed_units': minimum_p,
                                    'maximum_attempt_cap_at_assumed_units': max_attempts},
            'binding_break_even_envelope': {
                'budget_core_hours': 2880,
                'formula': 'binding_fixed_hours + U*binding_per_unit_hours + ceil(4000/p_gate)*binding_gate_hours_per_attempt <= 2880',
                'fixed_hours': binding_fixed, 'per_unit_hours': binding_per_unit,
                'gate_hours_per_attempt': binding_gate_per_attempt,
                'maximum_units_per_population_at_assumed_p_gate': binding_break_even,
                'maximum_L_at_mean1000': binding_break_even/3000 if binding_break_even else None,
                'maximum_attempt_cap_at_assumed_units': binding_max_attempts,
                'minimum_p_gate_at_assumed_units': 4000/binding_max_attempts if binding_max_attempts >= 4000 else None},
            'memory': {'L400_peak_rss_bytes': rss, 'eight_worker_peak_rss_bytes': 8*max(rss.values()),
                       'host_physical_memory_bytes': host_memory,
                       'fixed_reserve_bytes': memory_reserve,
                       'worker_memory_budget_bytes': memory_budget,
                       'calibration_rss_bytes_per_unit': rss_slope,
                       'calibration_rss_intercept_bytes': rss_intercept,
                       'eight_worker_memory_break_even_units': memory_break_even,
                       'eight_worker_projected_bytes_at_assumed_units': 8*(rss_intercept+units*rss_slope),
                       'envelope_caveat': 'linear calibration RSS inference with fixed 8 GiB reserve; locked max-phase lifetime high-water RSS remeasurement required',
                       'scope': 'conservative lifetime high-water; trials sum worker+helper; linear 8-worker projection, not measured simultaneous 8 workers',
                       'locked_requirement': 'Fixed 8 workers: 8*max(locked construction RSS, locked calibration RSS, locked trial worker+helper RSS) <= hw.memsize-8 GiB; lifetime high-water method; failed or unmeasured means not testable; no worker reduction'},
            'projected_wall_days_at_8_workers': total/192 if total is not None else None,
            'binding_critical_path_core_hours': binding_hours,
            'binding_critical_path_wall_days_at_8_workers': binding_hours/192 if binding_hours is not None else None,
            'descriptive_safety_core_hours': total-binding_hours if total is not None else None,
            'caps': {'workers': 8, 'core_hours': 2880, 'wall_days': 15},
            'fits_caps': binding_hours is not None and binding_hours <= 2880 and units <= memory_break_even,
            'total_planning_fits_cpu_caps': total is not None and total <= 2880,
            'authority': 'Synthetic resource evidence only; no preparation or full-study authority. No automatic grid reduction.'}


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
    if platform.python_version() != '3.11.11':
        raise RuntimeError('explicit pinned CPython 3.11.11 is required')
    if sys.flags.optimize:
        raise RuntimeError('X1 numerical validation refuses optimized Python')
    source = args.input or HERE/('outer-results.json' if args.report == 'pilot-summary' else 'compute-benchmark-rev7.json')
    if args.report == 'pilot-summary':
        result = pilot_summary(source)
    elif args.report == 'calibration-overhead':
        result = calibration_overhead()
    else:
        result = projection(source, args.p_gate, args.inventory_runs, args.mean_run_size)
    result['reporter_sha256'] = digest(Path(__file__))
    result['python_optimize'] = sys.flags.optimize
    result['reproduce_cli'] = ('source .tools/activate && .tools/venv/bin/python evaluation/x1-triage-ranking/validation/reproduce_reports.py '
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

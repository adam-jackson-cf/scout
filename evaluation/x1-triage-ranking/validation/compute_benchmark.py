#!/usr/bin/env python3
"""Synthetic-only process benchmark; no binding OC result or model interface.

Each cell has a SHA-256-derived independent stream. Scientific JSON is canonical
and independent of process scheduling. The same four cells execute with one and
two workers; only measured timing and process metadata are excluded from equality.
"""
import argparse
from collections import Counter
from concurrent.futures import ProcessPoolExecutor
from dataclasses import replace
import hashlib
import itertools
import json
import math
import os
from pathlib import Path
import platform
import random
import resource
import sys
import subprocess
import tempfile
import time

from ranking import synthetic
from fixed_population import prepare_population, evaluate_population

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
SIZES = (1000,)*5
MIXED_SIZES = (2000,)*5+(500,)*10+(1000,)*5
REPLICATES = 1999


def canonical(value):
    return json.dumps(value, sort_keys=True, separators=(',', ':'), allow_nan=False).encode()


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()



def peak_rss_bytes(children=False):
    usage = resource.getrusage(resource.RUSAGE_CHILDREN if children else resource.RUSAGE_SELF)
    return int(usage.ru_maxrss*(1 if sys.platform == 'darwin' else 1024))

def seed_for(*parts):
    return int.from_bytes(hashlib.sha256(canonical(['x1-compute-rev5', *parts])).digest()[:8], 'big')


def trial_frame(frame_runs, missing, trial, fixture='uniform-srs'):
    """Selected-frame trial fixture only, NEVER population-preparation evidence.

    Uniform grouped templates exercise SRS20/L. The mixed fixture supplies an
    explicit valid 60,000-unit selected frame (max run 2,000). Genuine independent
    L-run construction and fixed-key reevaluation are measured separately.
    """
    cell = [frame_runs, missing, fixture]
    rng = random.Random(seed_for('frame', cell, trial))
    sizes = SIZES if fixture == 'uniform-srs' else MIXED_SIZES
    template = synthetic(gain=.20, seed=seed_for('population', cell),
                         prevalence=.55, heterogeneity=.10, overlap=.5,
                         run_count=len(sizes), sizes=sizes, cap=400, failure=.02)
    by_repo = {}
    for row in template:
        by_repo.setdefault(row.repo, {}).setdefault(int(row.run), []).append(row)
    frame, selections = [], {}
    for repo, groups in sorted(by_repo.items()):
        selected = sorted(rng.sample(range(frame_runs), 20))
        selections[repo] = selected
        for index, run in enumerate(selected):
            for row in groups[index % len(sizes)]:
                frame.append(replace(row, run=f'{run:04}', issue=f'{run:04}/{row.issue.split("/")[-1]}',
                                     selected=True))
    if missing:
        # Every delta class has unknown Y in every repository, giving exactly
        # eight distinct vertices per repository and all 512 combinations.
        unknown = set()
        for repo in sorted(by_repo):
            for delta in (-1, 0, 1):
                rows = [row for row in frame if row.repo == repo and row.delta == delta and row.pi == 1.]
                if not rows:
                    raise ValueError('fixture does not span all missing-label vertices')
                rng.shuffle(rows)
                unknown.update((row.repo, row.issue) for row in rows[:max(1, len(rows)//100)])
                for run in sorted({row.run for row in rows}):
                    candidate = next(row for row in rows if row.run == run)
                    unknown.add((candidate.repo, candidate.issue))
        frame = [replace(row, truth=None) if (row.repo, row.issue) in unknown else row for row in frame]
    return frame, selections


def child_cpu():
    usage = resource.getrusage(resource.RUSAGE_CHILDREN)
    return usage.ru_utime+usage.ru_stime


def execute_cell(task):
    frame_runs, missing, method, trials, binary, fixture = task
    elapsed_start, cpu_start = time.perf_counter(), time.process_time()+child_cpu()
    setup = time.perf_counter()
    frame, selected_ids = trial_frame(frame_runs, missing, 0, fixture)
    if len(frame) != 60000 or any(row.n > 2000 for row in frame):
        raise ValueError('invalid worst-case envelope')
    population = []
    for repo in sorted({row.repo for row in frame}):
        rows = [row for row in frame if row.repo == repo]
        run_ids = sorted({row.run for row in rows})
        groups = []
        for run in run_ids:
            counts = Counter((row.pi, int(row.jev <= row.k), int(row.baseline <= row.k), row.truth is None)
                             for row in rows if row.run == run and row.truth != 0)
            groups.append({'Groups': [{'Count': count, 'Pi': pi, 'J': j, 'B': b, 'Unknown': unknown}
                                      for (pi, j, b, unknown), count in sorted(counts.items())],
                           'Capped': True})
        population.append([groups[index % len(groups)] for index in range(frame_runs)]
                          if fixture == 'uniform-srs' else groups)
    setup_seconds = time.perf_counter()-setup
    request = {'Method': method, 'Population': population, 'B': REPLICATES,
               'Trials': trials, 'Seed': seed_for('helper-cell', frame_runs, missing, fixture) & ((1 << 63)-1)}
    wall, cpu = time.perf_counter(), time.process_time()+child_cpu()
    process = subprocess.run([binary], input=canonical(request), capture_output=True, check=True,
                             env={**os.environ, 'GOMAXPROCS': '1'})
    outputs = [json.loads(line) for line in process.stdout.splitlines()]
    trial_elapsed = time.perf_counter()-wall
    trial_cpu = time.process_time()+child_cpu()-cpu
    if len(outputs) != trials:
        raise ValueError('helper emitted incorrect trial count')
    if missing and any(output.get('Combinations') != 512 for output in outputs):
        raise ValueError('missing-label benchmark must evaluate all 512 combinations in every trial')
    return {'cell': {'L': frame_runs, 'missing_labels': missing, 'method': method, 'fixture': fixture},
            'scientific': {'seed': request['Seed'], 'population_sha256': hashlib.sha256(canonical(population)).hexdigest(),
                           'explicit_selected_run_ids': selected_ids if fixture == 'mixed-explicit' else None,
                           'selected_units_every_trial': 60000, 'trial_outputs': outputs},
            'trial_timing': [{'elapsed_seconds': trial_elapsed/trials, 'cpu_seconds': trial_cpu/trials,
                              'measurement': 'amortized measured cell execution over actual trials'}],
            'measured_trials': trials, 'trial_execution_elapsed_seconds': trial_elapsed,
            'trial_execution_cpu_seconds': trial_cpu, 'setup_elapsed_seconds': setup_seconds,
            'peak_rss_bytes': peak_rss_bytes()+peak_rss_bytes(children=True),
            'rss_scope': 'conservative sum of worker and helper lifetime high-water RSS; includes setup',
            'elapsed_seconds': time.perf_counter()-elapsed_start,
            'cpu_seconds': time.process_time()+child_cpu()-cpu_start, 'pid': os.getpid()}


def execute(tasks, workers):
    started = time.perf_counter()
    with ProcessPoolExecutor(max_workers=workers) as pool:
        cells = list(pool.map(execute_cell, tasks))
    return {'workers': workers, 'elapsed_seconds': time.perf_counter()-started, 'cells': cells}


def science(run):
    return [{'cell': cell['cell'], 'scientific': cell['scientific']} for cell in run['cells']]




def benchmark_population(L):
    measurements = []
    for L in (L,):
        sizes = (2000,)*(L//4)+(500,)*(L//2)+(1000,)*(L//4)
        cpu, wall = time.process_time(), time.perf_counter()
        prepared = prepare_population(seed=seed_for('independent-population', L),
                                      prevalence=.55, heterogeneity=.10, overlap=.5,
                                      run_count=L, sizes=sizes, cap=400, failure=.02,
                                      rank_structure='linear-gradient', size_dependence=.05)
        construction = {'cpu_seconds': time.process_time()-cpu,
                        'elapsed_seconds': time.perf_counter()-wall,
                        'peak_rss_bytes': peak_rss_bytes(),
                        'rss_scope': 'fresh spawned process lifetime high-water after construction'}
        evaluations = []
        # Two endpoint evaluations and twenty bisection-like fixed-key evaluations.
        # These are real reevaluations, not 22 reconstructions or timing estimates.
        low = max(-.30+run.size_offset for run in prepared.runs)
        high = min(min(run.k, run.positives)/run.positives-.30+run.size_offset
                   for run in prepared.runs)
        for gain in (low, high, *(low+(high-low)*index/21 for index in range(1, 21))):
            cpu, wall = time.process_time(), time.perf_counter()
            frame = evaluate_population(prepared, gain)
            evaluations.append({'gain': gain, 'units': len(frame),
                                'cpu_seconds': time.process_time()-cpu,
                                'elapsed_seconds': time.perf_counter()-wall})
            del frame
        units = 3*sum(sizes)
        gate_start, gate_cpu = time.perf_counter(), time.process_time()
        passes = 0
        for draw in range(100000):
            rng = random.Random(seed_for('size-only-SRS20-gate', L, draw))
            totals = [sum(rng.sample(sizes, 20)) for _ in range(3)]
            passes += int(min(totals) >= 1000 and sum(totals) <= 60000)
        gate = {'draws': 100000, 'passes': passes, 'p_gate': passes/100000,
                'attempt_cap_per_population': math.ceil(4000/(passes/100000)) if passes else None,
                'cpu_seconds': time.process_time()-gate_cpu,
                'elapsed_seconds': time.perf_counter()-gate_start,
                'key': 'SHA256 canonical x1-compute-rev5,size-only-SRS20-gate,L,draw',
                'gate': 'each repository selected N>=1000; total selected N<=60000',
                'scope': 'synthetic benchmark sizes, not a locked-inventory estimate'}
        measurements.append({'L': L, 'repositories': 3, 'units': units,
                             'mean_run_size': sum(sizes)/L, 'max_run_size': max(sizes),
                             'independent_runs': True, 'run_template_replication': False,
                             'size_only_sampling_gate': gate,
                             'construction': {**construction,
                                              'cpu_seconds_per_unit': construction['cpu_seconds']/units},
                             'fixed_key_evaluations': evaluations,
                             'max_reevaluation_cpu_seconds_per_unit':
                                 max(item['cpu_seconds'] for item in evaluations)/units})
        del prepared
    return measurements[0]


def benchmark_populations():
    # Fresh interpreters keep L40/L400 high-water marks independent.
    with ProcessPoolExecutor(max_workers=1, max_tasks_per_child=1) as pool:
        return list(pool.map(benchmark_population, (40, 400)))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--method', choices=('percentile', 'studentized'), required=True)
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--trials', type=int, default=2000)
    parser.add_argument('--cores', type=int, default=8)
    parser.add_argument('--oc-days', type=float, default=15.)
    args = parser.parse_args()
    if args.output.exists():
        raise FileExistsError('benchmark evidence is create-once')
    if args.trials < 2000 or args.cores < 1 or not 0 < args.oc_days < 45:
        parser.error('trials>=2000, positive cores and 0 < oc-days < 45 required')
    if platform.python_version() != '3.11.11':
        raise RuntimeError('explicit pinned CPython 3.11.11 is required')
    if sys.flags.optimize:
        raise RuntimeError('X1 numerical validation refuses optimized Python')
    paths = [HERE/name for name in ('compute_benchmark.py', 'reproduce_reports.py', 'ranking.py',
                                   'fixed_population.py', 'interval_diagnostic.go')]
    sources = {str(path.relative_to(ROOT)): digest(path) for path in paths}
    started = time.perf_counter()
    population_measurements = benchmark_populations()
    with tempfile.TemporaryDirectory(prefix='compute-rev6-', dir=HERE) as temporary:
        binary = str(Path(temporary)/'interval-kernel')
        build_start = time.perf_counter()
        subprocess.run(['go', 'build', '-o', binary, str(HERE/'interval_diagnostic.go')], check=True)
        build = {'elapsed_seconds': time.perf_counter()-build_start,
                 'binary_sha256': digest(Path(binary)),
                 'go_version': subprocess.check_output(['go', 'version'], text=True).strip()}
        tasks = [(L, missing, args.method, args.trials, binary, fixture)
                 for L, missing, fixture in itertools.product((40, 400), (False, True),
                                                             ('uniform-srs', 'mixed-explicit'))]
        runs = [execute(tasks, workers) for workers in (1, 2)]
    identical = canonical(science(runs[0])) == canonical(science(runs[1]))
    if not identical:
        raise AssertionError('scientific output differs by worker count')
    if sources != {str(path.relative_to(ROOT)): digest(path) for path in paths}:
        raise RuntimeError('source changed during benchmark; output not published')
    result = {'schema': 'x1-compute-benchmark/rev7', 'synthetic_only': True,
              'binding_oc_gate': False, 'method': args.method, 'replicates': REPLICATES,
              'python_version': platform.python_version(), 'platform': platform.platform(),
              'python_optimize': sys.flags.optimize,
              'host_physical_memory_bytes': int(subprocess.check_output(['sysctl', '-n', 'hw.memsize'], text=True)),
              'python_executable': '.tools/venv/bin/python',
              'python_executable_sha256': digest(Path(sys.executable).resolve()),
              'available_logical_cpus': os.cpu_count(), 'source_sha256': sources,
              'kernel': 'compiled', 'compiled_kernel': build,
              'gomaxprocs_per_worker': 1,
              'elapsed_seconds': time.perf_counter()-started,
              'timing_scope': 'Actual trial costs include grouped audit sampling, common m19 draws, studentized interval/full512 envelope, helper CPU and transport. Selected-frame setup is NOT population preparation. Independent full-population construction and 22 fixed-key reevaluations are separately measured.',
              'frame_design': {'eligible_runs_per_repository': [40, 400], 'selected_runs_per_repository': 20,
                               'selected_units_total': 60000, 'uniform_run_sizes': SIZES,
                               'mixed_explicit_selected_run_sizes': MIXED_SIZES,
                               'population_representation': 'Uniform: L grouped1000unit runs, SRS20; mixed: explicit20 selected from L40/400, maxN2000, heterogeneous sizes and gains',
                               'missing_vertices_per_repository': 8, 'missing_vertex_combinations': 512},
              'worker_count_equality': {'workers': [1, 2], 'byte_identical_scientific_json': identical,
                                        'scientific_sha256': hashlib.sha256(canonical(science(runs[0]))).hexdigest(),
                                        'excluded': ['timing', 'pid', 'worker count']},
              'executions': runs, 'population_measurements': population_measurements,
              'reproduce_cli': 'source .tools/activate && .tools/venv/bin/python evaluation/x1-triage-ranking/validation/compute_benchmark.py --method studentized --trials 2000 --output evaluation/x1-triage-ranking/validation/compute-benchmark-rev7.json'}
    with args.output.open('x') as stream:
        json.dump(result, stream, indent=2, sort_keys=True, allow_nan=False)
        stream.write('\n')
    print(json.dumps({'output': args.output.name, 'elapsed_seconds': result['elapsed_seconds']},
                     sort_keys=True))


if __name__ == '__main__':
    main()

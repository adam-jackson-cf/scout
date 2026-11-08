#!/usr/bin/env python3
"""Create-once, synthetic-only R3-01 diagnostic; specification precedes all outcomes."""
import argparse
from collections import Counter
import hashlib
import json
from pathlib import Path
import statistics
import subprocess
import tempfile
import time
from ranking import synthetic, provenance, wilson

HERE = Path(__file__).resolve().parent
SOURCES = [HERE/'ranking.py', Path(__file__).resolve(), HERE/'interval_diagnostic.go']


def seed_for(phase, cell, role):
    message = f'x1-rev4-interval-v1/{phase}/{cell}/{role}'.encode()
    return int.from_bytes(hashlib.sha256(message).digest()[:8], 'big') & ((1 << 63)-1)


def specification():
    cells = []
    for block, cap, failure, dependence in [('M850', 850, 0., 0.), ('M2000-failure02', 2000, .02, 0.), ('no-cap-size-dependent', 2000, 0., .20)]:
        for length in (40, 100, 400):
            cell = f'{block}-L{length}'
            cells.append({'id': cell, 'block': block, 'L': length, 'cap': cap,
                          'failure': failure, 'size_dependence': dependence,
                          'sizes': [50, 250, 1000, 2000], 'gain': .20,
                          'prevalence': .20, 'heterogeneity': 0., 'overlap': 0.,
                          'rank_structure': 'uniform', 'scanner_true': False,
                          'seeds': {phase: {role: seed_for(phase, cell, role) for role in ('population', 'trials')} for phase in ('initial', 'fresh-studentized')}})
    return {'schema': 'x1-interval-diagnostic-spec/rev4', 'synthetic_only': True,
            'status': 'draft-for-council-not-binding-authority', 'trials_per_cell': 2000,
            'B': 1999, 'sample_runs_per_repository': 20, 'bootstrap_runs': 19,
            'population_count_per_cell': 1, 'cells': cells,
            'rng': 'population Python random.Random; trial Go math/rand/v2 PCG(seed, seed xor 0x9e3779b97f4a7c15)',
            'selection': 'all nine immutable named representative cells; no outcome-dependent cell selection',
            'denominators': 'all 2000 trials reported; >5% undefined replicates inadequate and excluded from tail denominator, with explicit counts',
            'rule': 'At L400, any block either one-sided pooled Wilson MC95 lower >.025 selects studentized; otherwise percentile. Each block has one representative L400 cell; no pooling across L or blocks. If selected, rerun all nine cells using fresh-studentized independent population and trial seeds. No automatic fallback.',
            'studentization': 'Vr=G sum((Ni-Dr Ti)-meanResidual)^2/((G-1)(sum Ti)^2), Ni=Ji-Bi; V=sum Vr/R^2. Replicates use m=G-1 and their own fitted ratio/variance. t=(Dstar-D)/sqrt(Vstar). Zero variance retains t=0 only if abs(Dstar-D)<=1e-14, otherwise undefined. Omit undefined ratios/t, require retained>=.95B. Retained Bprime order indices floor(.025Bprime), ceil(.975Bprime)-1; interval D-t975*sqrt(V), D-t025*sqrt(V), clipped [-1,1]. Same paired draws, unknown-d vertices and envelope. No FPC.',
            'provenance': provenance(SOURCES)}


def save_once(path, value):
    with path.open('x', encoding='utf-8') as stream:
        json.dump(value, stream, indent=2, sort_keys=True)
        stream.write('\n')


def population(cell, phase):
    frame = synthetic(gain=cell['gain'], seed=cell['seeds'][phase]['population'], prevalence=cell['prevalence'],
                      heterogeneity=cell['heterogeneity'], overlap=cell['overlap'], run_count=cell['L'],
                      sizes=tuple(cell['sizes']), cap=cell['cap'], failure=cell['failure'],
                      rank_structure=cell['rank_structure'], size_dependence=cell['size_dependence'])
    repositories, truths = [], []
    for repo in sorted({r.repo for r in frame}):
        runs = {}
        for row in frame:
            if row.repo == repo:
                runs.setdefault(row.run, []).append(row)
        result, t, x = [], 0, 0
        for rows in runs.values():
            groups = Counter((r.pi, int(r.jev <= r.k), int(r.baseline <= r.k)) for r in rows if r.truth == 1)
            result.append({'Groups': [{'Count': count, 'Pi': key[0], 'J': key[1], 'B': key[2]} for key, count in sorted(groups.items())],
                           'Capped': rows[0].n > cell['cap']})
            t += sum(r.truth for r in rows)
            x += sum(r.truth*r.delta for r in rows)
        repositories.append(result)
        truths.append(x/t)
    return repositories, statistics.mean(truths)


def run_cell(cell, phase, method, binary):
    start = time.monotonic()
    pop, truth = population(cell, phase)
    request = {'Seed': cell['seeds'][phase]['trials'], 'Trials': 2000, 'B': 1999,
               'Method': method, 'Population': pop}
    completed = subprocess.run([str(binary)], input=json.dumps(request), text=True, capture_output=True, check=True)
    results = [json.loads(line) for line in completed.stdout.splitlines()]
    if len(results) != 2000:
        raise ValueError('wrong diagnostic trial count')
    eligible = [r for r in results if r['Omitted']/1999 <= .05]
    low = sum(r['Interval'][0] > truth for r in eligible)
    high = sum(r['Interval'][1] < truth for r in eligible)
    def tail(count):
        return {'misses': count, 'trials': len(eligible), 'rate': count/len(eligible) if eligible else None,
                'wilson_mc95': wilson(count, len(eligible)) if eligible else None}
    errors = [r['Point']-truth for r in results]
    return {'cell': cell, 'phase': phase, 'method': method, 'truth': truth, 'trials': 2000,
            'eligible_trials': len(eligible), 'inadequate_trials': 2000-len(eligible),
            'bias_mean': statistics.mean(errors), 'bias_median': statistics.median(errors),
            'estimator_sd': statistics.stdev(r['Point'] for r in results),
            'mean_bootstrap_sd': statistics.mean(r['BootstrapSD'] for r in results),
            'mean_bootstrap_skewness': statistics.mean(r['Skew'] for r in results),
            'capped_runs_sampled_histogram': dict(sorted(Counter(r['Capped'] for r in results).items())),
            'lower_endpoint_above_truth': tail(low), 'upper_endpoint_below_truth': tail(high),
            'elapsed_seconds': time.monotonic()-start}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--spec', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    started = time.monotonic()
    with tempfile.TemporaryDirectory(prefix='.diagnostic-rev4-', dir=HERE) as temporary:
        binary = Path(temporary)/'interval-diagnostic'
        subprocess.run(['go', 'build', '-o', str(binary), str(HERE/'interval_diagnostic.go')], check=True)
        spec = specification()
        executed_provenance = spec['provenance']
        binary_digest = hashlib.sha256(binary.read_bytes()).hexdigest()
        spec['binary_sha256'] = binary_digest
        spec['go_version'] = subprocess.check_output(['go', 'version'], text=True).strip()
        save_once(args.spec, spec)
        initial = [run_cell(cell, 'initial', 'percentile', binary) for cell in spec['cells']]
        triggers = [r['cell']['id'] for r in initial if r['cell']['L'] == 400 and any(r[key]['wilson_mc95'] is not None and r[key]['wilson_mc95'][0] > .025 for key in ('lower_endpoint_above_truth', 'upper_endpoint_below_truth'))]
        method = 'studentized' if triggers else 'percentile'
        print(json.dumps({'chosen_method': method, 'trigger_cells': triggers}), flush=True)
        fresh = [run_cell(cell, 'fresh-studentized', method, binary) for cell in spec['cells']] if triggers else []
        if provenance(SOURCES) != executed_provenance or hashlib.sha256(binary.read_bytes()).hexdigest() != binary_digest:
            raise ValueError('executed source or binary changed during diagnostic')
        report = {'schema': 'x1-interval-diagnostic/rev4', 'synthetic_only': True,
                  'chosen_method': method, 'trigger_cells': triggers,
                  'initial': initial, 'fresh_studentized': fresh, 'spec_sha256': hashlib.sha256(args.spec.read_bytes()).hexdigest(),
                  'elapsed_seconds': time.monotonic()-started, 'provenance': executed_provenance,
                  'binary_sha256': binary_digest, 'go_version': spec['go_version']}
        save_once(args.output, report)
        print(json.dumps({'chosen_method': method, 'trigger_cells': triggers, 'elapsed_seconds': report['elapsed_seconds']}))


if __name__ == '__main__':
    main()

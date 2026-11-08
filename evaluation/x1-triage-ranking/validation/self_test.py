#!/usr/bin/env python3
"""X1 rev4 synthetic statistical validation; no real/model inputs."""
import argparse
from dataclasses import replace
from itertools import product
import json
from pathlib import Path
import random
import time
import math
import subprocess
import hashlib
import tempfile
from collections import Counter
import sys
import hmac
import fixed_population
import ranking
from fixed_population import (prepare_population, evaluate_population, calibrate_population,
                              realised_gain, gradient_slots, population_identity, criticality_key)

from ranking import (Unit, analyse, audit_uniform, bootstrap_totals, cluster_distribution,
                     delivery_integrity, draw, fit, label_payload, pending_uniforms,
                     percentile, provenance, safety, select_runs, synthetic, verdict,
                     zero_tail, assign, ratio_variance, interval_from_distribution, FAILURE_REASONS, wilson)


def hand_frame():
    rows = []
    for repo in ('a', 'b', 'c'):
        for run in range(20):
            for i, (j, b, y, pi, selected) in enumerate([
                    (1, 7, 1, .25, True), (7, 1, 1, 1., True),
                    (2, 2, 1, .5, True), (6, 3, 0, .5, True),
                    (3, 4, 0, .8, False), (4, 5, 0, .8, True),
                    (5, 6, 0, .8, False)]):
                rows.append(Unit(repo, f'{run:02}', f'{run:02}/{i}', 7, j, b, i != 3,
                                 False, y, False, pi, selected))
    return rows


def strong_effect_check(binary, analysis, check, null_frame):
    started = time.monotonic()
    spec = dict(seed=169503, prevalence=.20, heterogeneity=.10, run_count=40,
                sizes=(250, 500, 750, 1000), cap=2000, failure=0.)
    frame = synthetic(.45, **spec)
    truth = realised_gain(frame)
    check('strong effect registered acceptance frame', truth >= .30 and all(r.annotated for r in frame))
    population = []
    for repo in sorted({r.repo for r in frame}):
        groups = {}
        for row in frame:
            if row.repo == repo and row.truth == 1:
                groups.setdefault(row.run, Counter())[(row.pi, int(row.jev <= row.k), int(row.baseline <= row.k))] += 1
        population.append([{'Capped': False, 'Groups': [
            {'Count': count, 'Pi': pi, 'J': j, 'B': b}
            for (pi, j, b), count in sorted(counts.items())]} for _, counts in sorted(groups.items())])
    request = {'Seed': 169504, 'Trials': 2000, 'B': 1999, 'Method': 'studentized', 'Population': population}
    completed = subprocess.run([str(binary)], input=json.dumps(request), text=True, capture_output=True, check=True)
    results = [json.loads(line) for line in completed.stdout.splitlines()]
    supported = sum(verdict(r['Point'], r['Interval'][1], r['Interval'][0], r['Omitted']/1999 <= .05,
                            {'violating': 0, 'unresolved': 0}) == 'support' for r in results)
    interval = wilson(supported, len(results))
    check('strong effect OC2000 support Wilson lower at least .99', len(results) >= 2000 and interval[0] >= .99)
    fixed = {}
    for seed in (169, 170, 171):
        rng = random.Random(seed)
        chosen = {}
        for repo in sorted({r.repo for r in frame}):
            runs = sorted({r.run for r in frame if r.repo == repo})
            chosen[repo] = set(rng.sample(runs, 20))
        sampled = [r for r in frame if r.run in chosen[r.repo]]
        check(f'strong effect SRS20 size gates seed{seed}',
              len(sampled) <= 60000 and all(sum(r.repo == repo for r in sampled) >= 1000 and
              len({r.run for r in sampled if r.repo == repo}) == 20 for repo in chosen))
        check(f'strong sampled realised gain seed{seed}', realised_gain(sampled) >= .30)
        result = analysis(draw(sampled, rng), seed=seed, include_conditional=False)
        check(f'strong effect support seed{seed}', result['verdict'] == 'support')
        check(f'null unequal-pi G20 size gates seed{seed}',
              len(null_frame) <= 60000 and min(r.pi for r in null_frame) < 1 and
              len({r.pi for r in null_frame}) > 1 and
              all(len({r.run for r in null_frame if r.repo == repo}) == 20 and
                  sum(r.repo == repo for r in null_frame) >= 1000 for repo in chosen))
        null = analysis(draw(null_frame, random.Random(seed)), seed=seed, include_conditional=False)
        check(f'null non-support nonzero-width seed{seed}', null['verdict'] != 'support' and null['interval'][0] < null['interval'][1])
        fixed[str(seed)] = {'effect': result, 'null': null, 'effect_sampled_units': len(sampled),
                            'effect_sampled_runs': {repo: sorted(runs) for repo, runs in chosen.items()},
                            'null_sampled_units': len(null_frame), 'bootstrap_runs': 19}
    return {'schema': 'x1-strong-effect-OC/rev6', 'synthetic_only': True,
            'frame': dict(spec, gain=.45), 'realised_gain': truth, 'request': request,
            'trials': len(results), 'support_count': supported, 'support_wilson_mc95': interval,
            'inadequate_trials': sum(r['Omitted']/1999 > .05 for r in results),
            'fixed_seeds': fixed, 'elapsed_seconds': time.monotonic()-started}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--strong-output', type=Path, required=True)
    parser.add_argument('--method', choices=('percentile', 'studentized'), default='studentized')
    args = parser.parse_args()
    if sys.version_info[:3] != (3, 11, 11):
        raise RuntimeError('X1 numerical validation requires CPython 3.11.11')
    started = time.monotonic()
    source_paths = [Path(__file__).with_name(name) for name in ('ranking.py', 'self_test.py', 'interval_diagnostic.go', 'fixed_population.py')]
    executed_provenance = provenance(source_paths)
    executed_provenance['python_executable'] = sys.executable
    executed_provenance['python_executable_sha256'] = hashlib.sha256(Path(sys.executable).read_bytes()).hexdigest()
    temporary = tempfile.TemporaryDirectory(prefix='x1-rev6-')
    binary = Path(temporary.name)/'interval-kernel'
    subprocess.run(['go', 'build', '-o', str(binary), str(Path(__file__).with_name('interval_diagnostic.go'))], check=True)
    executed_provenance['binary_sha256'] = hashlib.sha256(binary.read_bytes()).hexdigest()
    executed_provenance['go_version'] = subprocess.check_output(['go', 'version'], text=True).strip()
    kernel_analyse = analyse
    def analysis(frame, **kwargs):
        return kernel_analyse(frame, method=args.method, **kwargs)
    tests, mutations = [], {}
    def check(name, condition):
        if not condition:
            raise AssertionError(name)
        tests.append(name)
    # Expected digests computed independently with openssl dgst -sha256 -hmac.
    for key, message, expected in (
            ('169-x1-runs', 'microsoft/TypeScript\0'+'a'*40, '182e905b75b8362f7cedbf343fae291a234aadd3319981adeef77ee3155f4c0c'),
            ('169-x1-audit', 'django/django\0'+'b'*40+'\0Scout:issue-1', 'fef1f60e670959bd7c4726a76a5e65cea7027cb5c8597249bb588002327cf8ab'),
            ('169-x1-dev', 'tokio-rs/axum\0'+'c'*40+'\0Scout:issue-2', '1ef09394ddc496a41e7c53608ab0e53bfdeb11b8f66f77bd9bdc1640518a6ffa')):
        check('canonical keyed digest '+key, hmac.new(key.encode(), message.encode(), hashlib.sha256).hexdigest() == expected)

    effect_frame = synthetic(.20, seed=9)
    null_frame = synthetic(0., seed=9, heterogeneity=.10)
    effect = analysis(draw(effect_frame))
    null = analysis(draw(null_frame))
    check('planted independent-set effect recovered', effect['verdict'] == 'support')
    check('noisy null nonzero width', null['interval'][0] < null['interval'][1])
    check('null not supported', null['verdict'] == 'not supported')
    check('null includes positive discordance both ways', all(any(row.truth == 1 and row.delta == sign for row in null_frame) for sign in (-1, 1)))
    # Cap-preserving mirrored runs give exact population null with unequal pi.
    source = synthetic(0., seed=169400, heterogeneity=.10, run_count=10,
                       sizes=(250, 1000), cap=400)
    noisy_frame = []
    for row in source:
        noisy_frame.append(replace(row, run=row.run+'a', issue=row.issue+'a'))
        noisy_frame.append(replace(row, run=row.run+'b', issue=row.issue+'b',
                                   jev=row.baseline, baseline=row.jev))
    noisy_frame = assign(noisy_frame)
    noisy_draw = draw(noisy_frame)
    noisy_null = analysis(noisy_draw, include_conditional=False)
    check('binding-cap noisy null population truth zero', sum(r.truth*r.delta for r in noisy_frame) == 0)
    check('binding-cap noisy null unequal stored pi', len({r.pi for r in noisy_frame}) > 1 and min(r.pi for r in noisy_frame) < 1)
    check('binding-cap noisy null annotation order', all(not r.annotated or r.baseline <= 400 for r in noisy_frame))
    check('binding-cap noisy null discordance both ways', all(any(r.truth == 1 and r.delta == sign for r in noisy_frame) for sign in (-1, 1)))
    check('binding-cap noisy null nonzero width', noisy_null['interval'][0] < noisy_null['interval'][1])
    check('binding-cap noisy null not supported', noisy_null['verdict'] == 'not supported')
    noisy_mutants = {}
    for name, pi in [('ignored stored pi', 1.), ('nominal pi', .5)]:
        mutant = analysis([replace(r, pi=pi) for r in noisy_draw], include_conditional=False)
        noisy_mutants[name] = mutant
        mutations['noisy null '+name] = (abs(mutant['gain_range'][0]-noisy_null['gain_range'][0]) > 1e-12 or
                                        any(abs(a-b) > 1e-12 for a, b in zip(mutant['interval'], noisy_null['interval'])))
        check('actual noisy-null mutation detected: '+name, mutations['noisy null '+name])
    check('permutation invariance', analysis(list(reversed(draw(effect_frame)))) == effect)
    selected_runs = select_runs('repo', [f'{i:040x}' for i in range(40)])
    check('HMAC SRS order independent', selected_runs == select_runs('repo', [f'{i:040x}' for i in reversed(range(40))]))
    check('SRS sample20 without replacement', len(selected_runs) == len(set(selected_runs)) == 20)
    check('zero-tail slice empty', zero_tail([1, 2, 3], 0) == [])
    frame = hand_frame()
    run = [row for row in frame if row.repo == 'a' and row.run == '00' and row.selected]
    fitted = fit(run, {-1: 0, 0: 0, 1: 0})
    check('hand HT totals7/6/3', (fitted['positive_total'], fitted['jev_total'], fitted['baseline_total']) == (7., 6., 3.))
    check('hand gain3/7', abs(fitted['gain']-3/7) < 1e-12)
    for name, pi in [('ignored stored pi', 1.), ('nominal pi', .5), ('selected fraction', 5/7)]:
        result = fit([replace(row, pi=pi) for row in run], {-1: 0, 0: 0, 1: 0})
        mutations[name] = abs(result['gain']-fitted['gain']) > .1
        check('mutation detected: '+name, mutations[name])
    swapped = analysis([replace(row, jev=row.baseline, baseline=row.jev) for row in draw(effect_frame)])
    mutations['arm swap'] = swapped['gain_range'][0] < 0 < effect['gain_range'][0]
    check('mutation detected: arm swap', mutations['arm swap'])
    # Differing run totals permit an independent arithmetic bootstrap oracle.
    values = {'by_run': {'a': (10., 8., 2.), 'b': (20., 2., 8.)}}
    distribution = cluster_distribution(values, ['a', 'b'], [[0], [1], [0, 1]])
    check('hand differing-run paired bootstrap', distribution == [.6, -.3, 0.])
    # Separate J/B resampling destroys pairing: run a J with run b B gives zero,
    # whereas the same-draw paired run a gives .6.
    unpaired = (values['by_run']['a'][1]-values['by_run']['b'][2])/10
    mutations['unpaired arm resampling'] = unpaired != distribution[0]
    check('mutation detected: unpaired arms', mutations['unpaired arm resampling'])
    # Mutate the actual resampling unit: thirty individual unit totals replace
    # the two known run totals, destroying the run-level variance.
    unit_values = [(1., 1., 0.)]*8+[(1., 0., 1.)]*2+[(1., 1., 0.)]*2+[(1., 0., 1.)]*8+[(1., 0., 0.)]*10
    unit_rng = random.Random(169)
    unit_draws = [[unit_rng.choices(range(30), k=29) for _ in range(199)] for _ in range(3)]
    _, unit_ci, _ = bootstrap_totals([unit_values]*3, unit_draws)
    mutations['unit rather than run resampling'] = unit_ci[1] < .6
    check('mutation detected: unit resampling', mutations['unit rather than run resampling'])
    rng = random.Random(169)
    draws = [[rng.choices(range(20), k=19) for _ in range(199)] for _ in range(3)]
    check('McCarthy-Snowden19 draws', all(len(indices) == 19 for repo in draws for indices in repo))
    point, interval, omitted = bootstrap_totals([[(10., 8., 2.)]*20]*3, draws)
    check('bootstrap totals arithmetic oracle', abs(point-.6) < 1e-12 and all(abs(v-.6) < 1e-12 for v in interval) and omitted == 0)
    # Independent hand oracle: D=0, residuals +/-6, G2, T30 => V=.16.
    check('hand run-linearized ratio variance .16', abs(ratio_variance([(10., 8., 2.), (20., 2., 8.)])-.16) < 1e-12)
    # Pivots [-2,0,2] and original SE .2 imply [-.4,.4], no pivot clipping.
    hand_ci, retained = interval_from_distribution(0., .04, [-.2, 0., .2], [.01]*3, 'studentized')
    check('hand percentile-t reverse quantiles and no pivot clipping', all(abs(a-b) < 1e-12 for a,b in zip(hand_ci, [-.4, .4])) and retained == 3)
    check('studentized zero variance signed infinities retained', interval_from_distribution(0., .04, [-.1, .1], [0., 0.], 'studentized') == ([-1., 1.], 2))
    check('studentized undefined variance omitted', interval_from_distribution(0., .04, [.1], [None], 'studentized')[1] == 0)
    check('studentized zero variance equal point retained', interval_from_distribution(0., 0., [0.], [0.], 'studentized') == ([0., 0.], 1))
    helper = Path(__file__).with_name('interval_diagnostic.go')
    fixture_rng = random.Random(169401)
    fixture_repos = [[(float(10+i), float((i*7)%11), float((i*3)%9)) for i in range(20)] for _ in range(3)]
    fixture_draws = [[fixture_rng.choices(range(20), k=19) for _ in range(1999)] for _ in range(3)]
    for method in ('percentile', 'studentized'):
        expected = bootstrap_totals(fixture_repos, fixture_draws, method=method)
        actual = json.loads(subprocess.check_output([str(binary)], input=json.dumps({'Method': method, 'Repositories': fixture_repos, 'Draws': fixture_draws}), text=True))
        check('Go explicit-draw differential '+method, abs(expected[0]-actual['Point']) <= 1e-12 and expected[2] == actual['Omitted'] and all(abs(a-b) <= 1e-12 for a,b in zip(expected[1], actual['Interval'])))
    infinite_repos = [[(10., 8., 2.), (10., 2., 8.)]]*3
    infinite_draws = [[[0]*19, [1]*19, [0, 1]*9+[0]]]*3
    expected = bootstrap_totals(infinite_repos, infinite_draws)
    actual = json.loads(subprocess.check_output([str(binary)], input=json.dumps({'Method': 'studentized', 'Repositories': infinite_repos, 'Draws': infinite_draws}), text=True))
    check('Go signed-infinite pivot differential', expected[1] == actual['Interval'] == [-1., 1.] and expected[2] == actual['Omitted'] == 0)
    # Independent boundary oracle: original D=shift, resample D*=0 and V*=0.
    # A singleton vertex per repository exercises the separate Go envelope path.
    tolerance_fixtures = []
    for magnitude in (0., 5e-15, 1e-14, 2e-14):
        for sign in (-1, 1):
            shift = sign*magnitude
            values = [(1., max(0., 2*shift), max(0., -2*shift)), (1., 0., 0.)]
            band_draws = [[[1, 1]]]
            expected_ci = [shift, shift] if magnitude <= 1e-14 else [float(sign), float(sign)]
            python_result = bootstrap_totals([values], band_draws)
            check(f'zero-variance independent complete tolerance {sign}/{magnitude}',
                  python_result == (shift, expected_ci, 0))
            # Inject the arithmetic oracle to isolate analyse's vertex combiner.
            original_fit, original_distribution, original_variance = ranking.fit, ranking.cluster_distribution, ranking.ratio_variance
            try:
                ranking.fit = lambda rows, unknown: {
                    'gain': shift, 'by_run': {f'{run:02}': (1., 0., 0.) for run in range(20)},
                    'influence': []}
                ranking.cluster_distribution = lambda fitted, runs, draws: [0.]*len(draws)
                ranking.ratio_variance = lambda values: shift*shift if len(values) == 20 else 0.
                python_vertex = ranking.analyse([row for row in frame if row.repo == 'a'],
                                                replicates=99, include_conditional=False)
            finally:
                ranking.fit, ranking.cluster_distribution, ranking.ratio_variance = original_fit, original_distribution, original_variance
            check(f'zero-variance independent Python vertex tolerance {sign}/{magnitude}',
                  python_vertex['interval'] == expected_ci and python_vertex['zero_denominator_replicates'] == 0)
            complete = json.loads(subprocess.check_output([str(binary)], input=json.dumps(
                {'Method': 'studentized', 'Repositories': [values], 'Draws': band_draws}), text=True))
            vertex = json.loads(subprocess.check_output([str(binary)], input=json.dumps(
                {'Method': 'studentized', 'Vertices': [[values]], 'Draws': band_draws}), text=True))
            check(f'Go complete tolerance oracle differential {sign}/{magnitude}',
                  complete['Point'] == shift and complete['Interval'] == expected_ci and complete['Omitted'] == 0)
            check(f'Go vertex tolerance oracle differential {sign}/{magnitude}',
                  vertex['GainRange'] == [shift, shift] and vertex['Interval'] == expected_ci and
                  vertex['Omitted'] == 0 and vertex['Combinations'] == 1)
            tolerance_fixtures.append(dict(shift=shift, expected_interval=expected_ci,
                                           python_complete=python_result, python_vertex=python_vertex,
                                           go_complete=complete, go_vertex=vertex))
    # Heterogeneous denominators expose fused multiply/subtract rounding.
    ratio_values = [(10., 2., 0.), (30., 6., 0.)]
    ratio_draws = [[[0, 0], [1, 1], [0, 1]]]
    ratio_expected = bootstrap_totals([ratio_values], ratio_draws)
    check('heterogeneous constant-ratio Python zero classification',
          ratio_variance(ratio_values) == 0. and all(
              ratio_variance([ratio_values[i] for i in indices]) == 0.
              for indices in ratio_draws[0]))
    for field, values in (('Repositories', [ratio_values]), ('Vertices', [[ratio_values]])):
        actual = json.loads(subprocess.check_output([str(binary)], input=json.dumps(
            {'Method': 'studentized', field: values, 'Draws': ratio_draws}), text=True))
        check('heterogeneous constant-ratio '+field+' byte-identical interval',
              [float(x).hex() for x in actual['Interval']] ==
              [float(x).hex() for x in ratio_expected[1]] and actual['Omitted'] == 0)
    disassembly = subprocess.check_output(['go', 'tool', 'objdump', '-s', 'main.fit', str(binary)], text=True)
    check('Go fit disassembly forbids fused multiply add/subtract',
          bool(disassembly.strip()) and not any(op in disassembly for op in ('FMADDD', 'FMSUBD', 'FNMSUBD')))
    zero_values = {'by_run': {'a': (0., 0., 0.), 'b': (1., 1., 0.)}}
    check('zero denominator explicitly omitted', cluster_distribution(zero_values, ['a', 'b'], [[0], [1]]) == [None, 1.])
    sparse = [replace(row, truth=1 if row.run == '00' and row.issue.endswith('/0') else 0, selected=True) for row in frame]
    sparse_result = analysis(sparse, replicates=99)
    check('greater than5percent zero denominator inadequate',
          sparse_result['zero_denominator_replicates']/99 > .05 and sparse_result['adequate'] is False)
    check('percentile uses retained Bprime', percentile([-.4, .2, .9]) == [-.4, .9])
    # Six missing bits across all d groups, unequal pi, differing runs.
    missing_keys = {('a', '00/0'), ('a', '01/3'), ('b', '00/1'),
                    ('b', '01/0'), ('c', '00/3'), ('c', '01/1')}
    missing = [replace(row, truth=None) if (row.repo, row.issue) in missing_keys else row for row in frame]
    # Heterogeneous measured runs are retained in every completion.
    missing = [replace(row, truth=0) if row.run == '02' and row.issue.endswith('/0') else row for row in missing]
    envelope = analysis(missing, replicates=99, include_conditional=False)
    keys = sorted(missing_keys)
    excess = 0.
    for bits in product((0, 1), repeat=6):
        mapping = dict(zip(keys, bits))
        completed = [replace(row, truth=mapping[(row.repo, row.issue)]) if row.truth is None else row for row in missing]
        ci = analysis(completed, replicates=99, include_conditional=False)['interval']
        excess = max(excess, envelope['interval'][0]-ci[0], ci[1]-envelope['interval'][1])
    check('64 heterogeneous all-d unequal-pi completions covered', excess <= 1e-12)
    # Full 8^3 envelope parity, including the Python vertex de-duplication.
    full_missing_keys = {(repo, f'0{i}/{i}') for repo in ('a', 'b', 'c') for i in range(3)}
    full_missing = [replace(r, truth=None) if (r.repo, r.issue) in full_missing_keys else r for r in frame]
    for method in ('percentile', 'studentized'):
        rng = random.Random(169)
        vertices, common_draws = [], []
        for repo in ('a', 'b', 'c'):
            rows = [r for r in full_missing if r.repo == repo and r.selected]
            runs = sorted({r.run for r in full_missing if r.repo == repo})
            common_draws.append([rng.choices(range(20), k=19) for _ in range(99)])
            seen, repo_vertices = set(), []
            for bits in product((0, 1), repeat=3):
                fitted_vertex = fit(rows, dict(zip((-1, 0, 1), bits)))
                values = [fitted_vertex['by_run'].get(run, (0., 0., 0.)) for run in runs]
                identity = tuple(values)
                if identity not in seen:
                    seen.add(identity)
                    repo_vertices.append(values)
            vertices.append(repo_vertices)
        expected = kernel_analyse(full_missing, replicates=99, seed=169, method=method, include_conditional=False)
        check('full512vertex combinations '+method, expected['vertex_combinations'] == 512)
        actual = json.loads(subprocess.check_output([str(binary)], input=json.dumps({'Method': method, 'Vertices': vertices, 'Draws': common_draws}), text=True))
        check('Go full vertex-envelope differential '+method, actual['Omitted'] == expected['zero_denominator_replicates'] and actual['Combinations'] == expected['vertex_combinations'] and all(abs(a-b) <= 1e-12 for a,b in zip(actual['Interval'], expected['interval'])) and all(abs(a-b) <= 1e-12 for a,b in zip(actual['GainRange'], expected['gain_range'])))
    row = next(row for row in effect_frame if row.below and row.truth == 1)
    check('missing criticality unresolved', safety([replace(row, critical=None)])['unresolved'] == 1)
    check('unknown truth noncritical cleared', safety([replace(row, truth=None, critical=False)])['unresolved'] == 0)
    check('census-only violation veto', analysis([replace(r, selected=False, critical=True) if r.repo == row.repo and r.issue == row.issue else r for r in draw(effect_frame)])['verdict'] == 'not supported')
    check('integrity failure outranks safety veto', verdict(.2, .3, .1, True, {'violating': 1, 'unresolved': 0}, False) == 'inconclusive')
    check('retrieval negative outranks safety unknown', verdict(0., .05, -.1, True, {'violating': 0, 'unresolved': 1}) == 'not supported')
    check('supporting retrieval unresolved safety inconclusive', verdict(.2, .3, .1, True, {'violating': 0, 'unresolved': 1}) == 'inconclusive')
    check('point exactly .10 supports', verdict(.10, .3, .01, True, {'violating': 0, 'unresolved': 0}) == 'support')
    check('lower exactly zero not support', verdict(.2, .3, 0., True, {'violating': 0, 'unresolved': 0}) != 'support')
    check('upper exactly .10 not negative', verdict(.05, .10, -.1, True, {'violating': 0, 'unresolved': 0}) != 'not supported')
    check('violation despite inadequate retrieval negative', verdict(.2, .3, .1, False, {'violating': 1, 'unresolved': 0}) == 'not supported')
    check('study fallback fails delivery integrity', not delivery_integrity([replace(row, reason='study_latch')]))
    check('product unknown-cost stop excluded from denominator', delivery_integrity([replace(row, reason='billed cost unknown')]))
    for reason in sorted(FAILURE_REASONS):
        check('product failure numerator '+reason, not delivery_integrity([replace(row, reason=reason)]))
    check('missing key run-level integrity precedence', not delivery_integrity([replace(row, reason='key_missing')]))
    check('sink failure integrity', not delivery_integrity([replace(row, reason='annotations not persisted')]))
    boundary = [replace(row, reason='annotated')]*9+[replace(row, reason='network')]
    check('failure threshold exactly ten percent permitted', delivery_integrity(boundary))
    check('unknown-cost units do not dilute failures', not delivery_integrity(boundary+[replace(row, reason='timeout')]+[replace(row, reason='billed cost unknown')]*100))
    check('severity hidden in payload', 'severity' not in label_payload({'severity': 'critical', 'excerpt': []}))
    pending = pending_uniforms(effect_frame[:3], {(effect_frame[0].repo, effect_frame[0].issue)})
    check('resume HMAC excludes already-stored draw', pending == [audit_uniform(r.repo, r.run, r.issue) for r in effect_frame[1:3]])
    capped = synthetic(.20, sizes=(1000,), cap=400, failure=.02)
    check('cap order only baseline topM eligible for annotation', all(not r.annotated or r.baseline <= 400 for r in capped))
    check('network failures have product reason', any(r.reason == 'network' for r in capped))
    for cap in (400, 550, 700, 850, 1000, 1500, 2000):
        registered = synthetic(.20, run_count=1, sizes=(1000,), cap=cap, rank_structure='linear-gradient')
        check('refined cap grid and gradient '+str(cap), all(not r.annotated or r.baseline <= cap for r in registered))
    dependent = synthetic(.20, sizes=(250, 1000), size_dependence=.05)
    gains = {n: sum(r.truth*r.delta for r in dependent if r.n == n)/sum(r.truth for r in dependent if r.n == n) for n in (250, 1000)}
    check('registered .05 run-size gain dependence', abs(gains[250]-.20) < 1e-12 and abs(gains[1000]-.15) < 1e-12)
    for length, positives in ((40, 8), (70, 21), (100, 55)):
        for seed in (169, 170, 171):
            slots = gradient_slots(length, positives, seed)
            density = positives/length
            slope = min(density, 1-density)
            discrepancies = [abs(sum(slots[:j])-(j*density+slope*j*(1-j/length))) for j in range(1, length+1)]
            check(f'gradient count/prefix/decline H{length} P{positives} seed{seed}', sum(slots) == positives and max(discrepancies) < 1 and sum(slots[:length//2]) > sum(slots[length//2:]))
    for seed in (169, 170, 171):
        spec = dict(seed=seed, run_count=2, sizes=(50, 100), cap=400, heterogeneity=.10, rank_structure='linear-gradient')
        prepared = prepare_population(**spec)
        for gain in (0., .10, .30):
            check(f'fixed-key fresh-construction determinism {seed}/{gain}', evaluate_population(prepared, gain) == synthetic(gain, **spec))
            check(f'truth-only gain equals full-frame truth {seed}/{gain}',
                  fixed_population._population_gain(prepared, gain) ==
                  realised_gain(evaluate_population(prepared, gain)))
        calibrated = calibrate_population(prepared, .10)
        candidates = fixed_population._allocation_candidates(prepared)
        attainable = [realised_gain(evaluate_population(prepared, gain)) for gain in candidates if 0 <= gain <= .5]
        check(f'calibration .10 smallest attainable {seed}', calibrated['reachable'] and calibrated['realised'] == min(value for value in attainable if value >= .10))
        for target in (0., .10, .20):
            optimized = calibrate_population(prepared, target)
            truth_only = fixed_population._population_gain
            try:
                fixed_population._population_gain = lambda population, gain: realised_gain(evaluate_population(population, gain))
                reference = calibrate_population(prepared, target)
            finally:
                fixed_population._population_gain = truth_only
            check(f'truth-only calibration exact full-frame equivalence {seed}/{target}',
                  optimized == reference and hashlib.sha256(repr(optimized['frame']).encode()).hexdigest() ==
                  hashlib.sha256(repr(reference['frame']).encode()).hexdigest())
    calibration_profiles = [
        dict(seed=169600+index, run_count=2, sizes=(500, 1000), cap=cap,
             prevalence=prevalence, size_dependence=size_dependence, failure=.02,
             overlap=.5, heterogeneity=.10, scanner_true=True, rank_structure='linear-gradient')
        for index, (cap, prevalence, size_dependence) in
        enumerate(product((400, 850), (.10, .55), (.05, .20)))]
    calibration_profiles.append(dict(seed=169503, run_count=3, sizes=(2000, 500, 1000),
                                     cap=400, prevalence=.55, size_dependence=.05,
                                     failure=.02, overlap=.5, heterogeneity=.10,
                                     scanner_true=True, rank_structure='linear-gradient'))
    calibration_evidence = []
    for profile_index, spec in enumerate(calibration_profiles):
        population = prepare_population(**spec)
        check(f'extended binding-cap/failure/scanner profile {profile_index}',
              any(run.n > spec['cap'] for run in population.runs) and
              any(run.failed for run in population.runs) and
              any(any(issue < run.positives for issue in run.severity) for run in population.runs))
        results = []
        for target in (0., .10, .20):
            check(f'extended truth-only/full-frame {profile_index}/{target}',
                  fixed_population._population_gain(population, target) ==
                  realised_gain(evaluate_population(population, target)))
            optimized = calibrate_population(population, target)
            truth_only = fixed_population._population_gain
            try:
                fixed_population._population_gain = lambda prepared, gain: realised_gain(evaluate_population(prepared, gain))
                reference = calibrate_population(population, target)
            finally:
                fixed_population._population_gain = truth_only
            check(f'extended optimized/full-frame oracle {profile_index}/{target}', optimized == reference)
            results.append({key: value for key, value in optimized.items() if key != 'frame'})
        calibration_evidence.append(dict(profile=spec, results=results))
    truth_only = fixed_population._population_gain
    try:
        fixed_population._population_gain = lambda population, gain: -999.
        try:
            calibrate_population(population, .10)
        except RuntimeError as error:
            check('mutant calibration truth mismatch refused', str(error) == 'calibration truth/full-frame mismatch')
        else:
            raise AssertionError('mutant calibration truth mismatch accepted')
    finally:
        fixed_population._population_gain = truth_only
    five = prepare_population(seed=169, prevalence=.10, run_count=2, sizes=(50,), cap=400)
    five_result = calibrate_population(five, .20)
    check('P5 calibration retains attainable .20 rather than skipping to .40',
          five_result['reachable'] and abs(five_result['realised']-.20) < 1e-15 and
          all(5*sum(r.truth*r.delta for r in five_result['frame'] if r.repo == repo) ==
              sum(r.truth for r in five_result['frame'] if r.repo == repo)
              for repo in ('synthetic-python', 'synthetic-typescript', 'synthetic-go')))
    for name, transition_population in (
            ('P5', five),
            ('mixed-sizes', prepared),
            ('size-offset', prepare_population(seed=170, run_count=2, sizes=(50, 100), size_dependence=.05))):
        candidates = fixed_population._allocation_candidates(transition_population)
        adjacent_valid = True
        for gain in candidates[1:-1]:
            previous = math.nextafter(gain, -math.inf)
            before = [round(r.positives*(.30+previous-r.size_offset)) for r in transition_population.runs]
            after = [round(r.positives*(.30+gain-r.size_offset)) for r in transition_population.runs]
            adjacent_valid &= all(a-b in (0, 1) for b, a in zip(before, after)) and any(a-b == 1 for b, a in zip(before, after))
            adjacent_valid &= fixed_population._indexed_float(fixed_population._float_index(gain)) == gain
        check('exact adjacent IEEE754 allocation transitions '+name, adjacent_valid)
    first = {'cap': 400, 'critical_share': .01}
    second = dict(first, critical_share=.05)
    group = population_identity('safety-interval-group', first)
    check('safety profile identity includes critical share', population_identity('safety-profile', first) != population_identity('safety-profile', second))
    check('safety interval identity shares masks and draws', group == population_identity('safety-interval-group', second))
    check('canonical criticality phase separates critical shares', criticality_key(group, .01) != criticality_key(group, .05))
    strong = strong_effect_check(binary, analysis, check, noisy_frame)
    strong['provenance'] = executed_provenance
    with args.strong_output.open('x', encoding='utf-8') as stream:
        json.dump(strong, stream, indent=2, sort_keys=True)
        stream.write('\n')
    for name, broken in [('malformed criticality', [replace(effect_frame[0], critical='no')]+effect_frame[1:]),
                         ('invalid pi', [replace(effect_frame[0], pi=0)]+effect_frame[1:]),
                         ('duplicate key', effect_frame+effect_frame[:1])]:
        try:
            analysis(broken, replicates=99)
        except ValueError:
            tests.append(name+' refused')
        else:
            raise AssertionError(name)
    check('source provenance unchanged during selftest', provenance(source_paths)['script_sha256'] == executed_provenance['script_sha256'])
    report = {'schema': 'x1-synthetic-validation/rev7', 'synthetic_only': True, 'method': args.method,
              'production_rule4_complete': False, 'checks_passed': len(tests), 'checks': tests,
              'mutations_detected': mutations, 'planted_effect': effect, 'planted_census_null': null,
              'planted_unequal_pi_null': noisy_null, 'unequal_pi_null_mutants': noisy_mutants,
              'exhaustive_completion_count': 64, 'exhaustive_interval_excess': excess,
              'zero_variance_tolerance_fixtures': tolerance_fixtures,
              'heterogeneous_constant_ratio': {'expected': ratio_expected, 'draws': ratio_draws},
              'fit_disassembly': disassembly,
              'extended_calibration_equivalence': calibration_evidence,
              'elapsed_seconds': time.monotonic()-started,
              'provenance': executed_provenance}
    report['historical_selection_diagnostic_rerun'] = False
    report['historical_selection_diagnostic_reason'] = 'Studentized branch already fixed; signed-infinite pivots checked by hand and Go differential, not method reselection.'
    with args.output.open('x', encoding='utf-8') as stream:
        json.dump(report, stream, sort_keys=True, indent=2)
        stream.write('\n')
    print(json.dumps({'checks_passed': len(tests), 'elapsed_seconds': report['elapsed_seconds'],
                      'planted_effect': effect['verdict'], 'planted_null': null['verdict']}))
    temporary.cleanup()


if __name__ == '__main__':
    main()

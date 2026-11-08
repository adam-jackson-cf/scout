"""Authorised synthetic-only rev7 OC timing; no inventory or model access."""
import hashlib
import importlib.util
import json
from pathlib import Path
import re
import subprocess
import sys
import tempfile

ROOT = Path(__file__).resolve().parent


def sha(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def main():
    if not sys.flags.isolated or not sys.dont_write_bytecode:
        raise ValueError('requires pinned Python -I -B')
    recipe_path = ROOT / 'tool-pins-v6/oc-build-recipe-v6.json'
    recipe = json.loads(recipe_path.read_text())
    helper = ROOT / 'pin_tool_closure.py'
    if sha(helper) != recipe['sources'][str(helper)]:
        raise ValueError('runtime helper drift')
    if any(ROOT.rglob('__pycache__')) or any(ROOT.glob('*.pyc')):
        raise ValueError('stale study bytecode')
    spec = importlib.util.spec_from_file_location('pin_tool_closure', helper)
    pin = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(pin)
    runtime = pin.load_pin(recipe['python_runtime'])
    pin.verify_python_runtime({'python_runtime': recipe['python_runtime'], 'tools': {
        'python': {'path': runtime['path'], 'sha256': runtime['sha256']}}})
    closure = pin.load_pin(recipe['closure'])
    # The only reviewed rev7 compiler-input changes are this source digest and
    # its three extra direct imports, all already present in the pinned stdlib closure.
    closure['files'][recipe['source']] = sha(recipe['source'])
    main_package = next(package for package in closure['packages']
        if package['ImportPath'] == 'command-line-arguments')
    main_package['Imports'] = sorted(set(main_package['Imports']) | {'bytes', 'strconv', 'strings'})
    actual_closure = pin.verify_go_closure(recipe, closure)
    closure_path = ROOT / 'oc-v7-timing-go-closure.json'
    output = ROOT / 'oc-v7-timing-metadata.json'
    names = [f'oc-v7-timing-{profile}{suffix}.json' for suffix in ('', '-maximum')
             for profile in ('balanced', 'concentrated', 'singleton')]
    if output.exists() or closure_path.exists() or any((ROOT / name).exists() for name in names):
        raise ValueError('timing evidence is create-once')
    with closure_path.open('x') as stream:
        json.dump(actual_closure, stream, indent=2, sort_keys=True)
        stream.write('\n')
    metadata = {'schema': 'x11a-oc-v7-timing-metadata-v1',
        'scope': 'Authorised synthetic timing only; not inventory-conditioned OC or statistical evidence',
        'source': str(ROOT / 'oc_v5.go'), 'source_sha256': sha(ROOT / 'oc_v5.go'),
        'timing_source': str(Path(__file__).resolve()), 'timing_source_sha256': sha(__file__),
        'recipe': {'path': str(recipe_path), 'sha256': sha(recipe_path)},
        'recipe_source_pins_note': 'Recipe6 build argv/environment reused read-only; closure change permits only revised Go source digest and bytes/strconv/strings direct imports already in pinned stdlib closure. Exact current closure and revised source digests recorded; Main refreshes final packet pins.',
        'current_sources': {name: sha(name) for name in recipe['sources']},
        'python_runtime': recipe['python_runtime'], 'go': recipe['go'], 'closure': {
            'path': str(closure_path), 'sha256': sha(closure_path)}, 'prior_closure': recipe['closure'],
        'profiles': []}
    with tempfile.TemporaryDirectory(prefix='x11a-oc-v7-timing-') as temporary:
        base = Path(temporary)
        environment = dict(recipe['environment'])
        for key in ('HOME', 'GOCACHE', 'GOTMPDIR'):
            environment[key] = str(base / key.lower())
            Path(environment[key]).mkdir()
        binary = base / 'oc-v7'
        argv = list(recipe['argv'])
        argv[argv.index('-o') + 1] = str(binary)
        subprocess.run(argv, cwd=recipe['cwd'], env=environment, check=True)
        version_argv = [recipe['go']['path'], 'version', '-m', str(binary)]
        metadata.update(build_argv=argv, build_environment=environment, build_cwd=recipe['cwd'],
            binary_sha256=sha(binary), buildinfo=subprocess.check_output(version_argv,
                cwd=recipe['cwd'], env=environment, text=True), version_m_argv=version_argv)
        for suffix, lengths in (('', '1500,1666,2000'), ('-maximum', '2000,2000,2000')):
            for profile in ('balanced', 'concentrated', 'singleton'):
                artifact = ROOT / f'oc-v7-timing-{profile}{suffix}.json'
                command = [str(binary), '--mode', 'timing', '--source', recipe['source'],
                    '--out', str(artifact), '--bootstrap', '399', '--timing-replicates', '20',
                    '--directory-profile', profile, '--segment-lengths', lengths]
                measured = subprocess.run(['/usr/bin/time', '-l', *command], cwd=recipe['cwd'],
                    env=environment, text=True, capture_output=True, check=True, timeout=120)
                report = json.loads(artifact.read_text())
                rss = re.search(r'(\d+)\s+maximum resident set size', measured.stderr)
                row = {'profile': profile, 'segment_lengths': report['segment_lengths'],
                    'argv': command, 'measurement_argv': ['/usr/bin/time', '-l', *command],
                    'artifact': artifact.name, 'artifact_sha256': sha(artifact),
                    'elapsed_seconds': report['elapsed_seconds'], 'studies': report['studies'],
                    'peak_rss_bytes': int(rss.group(1)) if rss else None,
                    'time_stderr': measured.stderr, 'directory_summary': report['directory_summary'],
                    'commit_block_draws': report['commit_block_draws'],
                    'final_block_positions': report['final_block_positions'],
                    'projected_minimum_core_hours': report['projected_minimum_seconds'] / 3600,
                    'projected_maximum_core_hours': report['projected_maximum_seconds'] / 3600,
                    'minimum_fits_384h': report['projected_minimum_seconds'] <= 384 * 3600,
                    'maximum_fits_384h': report['projected_maximum_seconds'] <= 384 * 3600}
                metadata['profiles'].append(row)
                print(json.dumps({key: row[key] for key in ('artifact', 'elapsed_seconds',
                    'projected_minimum_core_hours', 'projected_maximum_core_hours', 'peak_rss_bytes')}), flush=True)
    metadata['projection'] = {'binding_scenarios': 56160, 'minimum_replicates': 256,
        'minimum_studies': 14376960, 'maximum_studies': 57507840, 'single_core_budget_hours': 384,
        'formula': 'measured elapsed_seconds / measured studies * 56160 * replicates / 3600',
        'capacity_risk': 'Measured profiles are not a runtime guarantee or universal worst-case bound. Actual OC uses minimum256-replicate all-scenario projection and elapsed cap; exceeding384h stops pre-map without filtering or reselection. Full1024-replicate projections can exceed cap even where minimum fits. Sensitivity retains separate8h cap; incomplete remainder unmeasured.'}
    metadata['timing_definition'] = 'Six synthetic scenario indices x20 replicates=120 studies/profile; bootstrap399,2600cases/repository-age; mixed lengths1500/1666/2000 and all2000 upper envelope; balanced100shared directory keys, concentrated100shared keys with largest1300/2600, singleton5200age-disjoint union keys/repository; same generator, quality/missingness, paired commit and directory bounds as OC. Go elapsed excludes build and report assembly; /usr/bin/time -l supplies process peak RSS on Darwin.'
    with output.open('x') as stream:
        json.dump(metadata, stream, indent=2, sort_keys=True)
        stream.write('\n')


if __name__ == '__main__':
    main()

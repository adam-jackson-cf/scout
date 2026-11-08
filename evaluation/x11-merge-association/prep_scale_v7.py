#!/usr/bin/env python3
"""Offline synthetic production PREP/verification/frontend benchmark, never candidates."""
import argparse
import hashlib
import importlib
import json
import os
from pathlib import Path
import platform
import resource
import shutil
import sys
import time
from types import SimpleNamespace

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))


def dump(path, value):
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open('x') as stream:
        json.dump(value, stream, sort_keys=True, indent=2)
        stream.write('\n')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--inputs', type=Path, required=True)
    parser.add_argument('--work', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--scenario', choices=['files', 'bytes', 'construct'], required=True)
    args = parser.parse_args()
    for value in (args.inputs, args.work, args.output):
        if not value.is_absolute():
            parser.error('absolute paths required')
    args.work.mkdir(parents=True, exist_ok=False)
    # Freeze the reviewed Python/protocol source into a retained synthetic-only packet.
    # All actual operations use these identical bytes; no functions are mocked.
    snapshot = args.work/'source'
    snapshot.mkdir()
    import stages as original_stages
    names = set(original_stages.REVIEWED) | {p.name for p in HERE.glob('*.py')}
    original_sources = []
    for name in sorted(names):
        source, target = HERE/name, snapshot/name
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(source, target)
        original_sources.append({'original': {'path':str(source), 'sha256':hashlib.sha256(source.read_bytes()).hexdigest()},
                                 'snapshot': {'path':str(target), 'sha256':hashlib.sha256(target.read_bytes()).hexdigest()}})
    for name in list(sys.modules):
        module_path = getattr(sys.modules[name], '__file__', None)
        if name != '__main__' and module_path and Path(module_path).resolve().parent == HERE:
            del sys.modules[name]
    sys.path.insert(0, str(snapshot))
    prep = importlib.import_module('prepare_exports')
    stages = importlib.import_module('stages')
    closure = importlib.import_module('pin_tool_closure')
    importlib.import_module('inventory')
    config = prep.load(args.inputs)
    config['recipe_sources'] = [prep.file_pin(snapshot/Path(p['path']).relative_to(HERE))
        if Path(p['path']).is_relative_to(HERE) and (snapshot/Path(p['path']).relative_to(HERE)).is_file()
        else prep.file_pin(Path(p['path'])) for p in config['recipe_sources']]
    tools = {key: prep.pin(value) for key, value in config['export_preparation']['tools'].items()}
    prep.require(Path(sys.executable).resolve() == tools['python'], 'pinned Python required')
    original_runtime = config['export_preparation']['python_runtime']
    runtime = prep.load(prep.pin(original_runtime))
    prep.require(runtime['local_module_paths'] == [str(HERE)], 'unexpected original local module root')
    runtime['local_module_paths'] = [str(snapshot)]
    runtime_path = args.work/'python-runtime-location-only.json'
    dump(runtime_path, runtime)
    config['export_preparation']['python_runtime'] = prep.file_pin(runtime_path)
    evidence = {'schema': 'x11a-preparation-scale-v7', 'synthetic_only': True,
        'scenario': args.scenario, 'command': list(sys.orig_argv), 'started_utc': prep.utc(),
        'runner': prep.file_pin(Path(__file__)), 'preparation_script': prep.file_pin(snapshot/'prepare_exports.py'),
        'source_snapshot': [prep.file_pin(p) for p in sorted(snapshot.rglob('*')) if p.is_file()],
        'original_snapshot_pairs': original_sources,
        'runtime_location_derivation': {'original':original_runtime,
            'derived':config['export_preparation']['python_runtime'],
            'only_changed_field':'local_module_paths', 'from':[str(HERE)], 'to':[str(snapshot)]},
        'tool_pins': config['export_preparation']['tools'], 'python_runtime': config['export_preparation']['python_runtime'],
        'scanner_tree': config['export_preparation']['clawpatch_tree'],
        'platform': {'system': platform.system(), 'release': platform.release(), 'version': platform.version(),
                     'machine': platform.machine(), 'python': platform.python_version()}, 'operations': []}
    started = time.monotonic()
    home, binaries, repo = args.work/'home', args.work/'bin', args.work/'repo'
    home.mkdir(); binaries.mkdir(); repo.mkdir()
    (binaries/'git').symlink_to(tools['git'])
    env = prep.environment(home, binaries)
    evidence['fixture_environment'] = env
    def git(*argv, data=None):
        return prep.git(tools, env, repo, *argv, input_bytes=data)
    git('init', '-q')
    env.update(GIT_AUTHOR_NAME='Synthetic capacity', GIT_AUTHOR_EMAIL='synthetic@example.invalid',
               GIT_COMMITTER_NAME='Synthetic capacity', GIT_COMMITTER_EMAIL='synthetic@example.invalid',
               GIT_AUTHOR_DATE='2026-01-01T00:00:00Z', GIT_COMMITTER_DATE='2026-01-01T00:00:00Z')
    count, revisions, padding = (80000, 20, 0) if args.scenario == 'files' else (2000, 3, 1000000)
    if args.scenario == 'construct':
        count, revisions, padding = 40, 3, 0
    module = b'module example.invalid/scalesynthetic\n\ngo 1.24\n'
    (repo/'go.mod').write_bytes(module)
    paths = []
    def source(position, revision):
        return ('package component%03d\nfunc Item%05d() int { return %d }\n' %
                (position//1000, position, revision)).encode() + (b'/*'+b'x'*padding+b'*/\n' if padding else b'')
    for position in range(count):
        path = repo/'internal'/('component%03d' % (position//1000))/('file%05d.go' % position)
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(source(position, 0)); paths.append(path)
    commits = []
    side = None
    changed = max(1, count//20)
    for revision in range(revisions):
        if revision:
            for position in range((revision-1)*changed, revision*changed):
                paths[position % count].write_bytes(source(position % count, revision))
        git('add', '.')
        tree = git('write-tree').decode().strip()
        if side is None:
            side = git('commit-tree', tree, data=b'synthetic independent side parent\n').decode().strip()
        parents = [] if not commits else ['-p', commits[-1], '-p', side]
        commits.append(git('commit-tree', tree, *parents, data=('synthetic revision %d\n' % revision).encode()).decode().strip())
    git('update-ref', 'refs/heads/master', commits[-1])
    byte_count = sum(path.stat().st_size for path in paths) + len(module)
    evidence['fixture'] = {'files_per_revision': count+1, 'source_files_per_revision': count,
        'bytes_per_revision_min': byte_count, 'revision_count': revisions,
        'changed_source_files_each_revision_after_base': changed, 'change_fraction': changed/count,
        'language': 'Go', 'padding_bytes_per_source_file': padding,
        'shape': 'package/function sources; byte-heavy scenario adds one block comment per source; no candidate-informed complexity tuning',
        'first_parent_revisions': commits, 'merge_revisions': commits[1:]}
    config['repositories'] = [{'repository': 'synthetic/production-benchmark', 'path': str(repo),
        'default_branch': 'master', 'resolved_remote_tip': commits[-1], 'tip': commits[-1],
        'Cold': commits[0], 'C0': commits[0], 'canonical_clone_url': 'https://example.invalid/synthetic.git'}]
    recipe = prep.load(prep.pin(config['export_preparation']['frontend_build_recipe']))
    binary = args.work/'built/frontend'
    recipe['argv'][recipe['argv'].index('-o')+1] = str(binary)
    recipe['version_m_argv'][-1] = str(binary)
    recipe['required_buildinfo_path'] = str(binary.parent/'go-version-m.txt')
    build_env = dict(recipe['environment'])
    for key in ('HOME', 'GOCACHE', 'GOTMPDIR'):
        path = args.work/('build-'+key.lower()); path.mkdir(); build_env[key] = str(path)
    closure_value = closure.go_metadata(recipe, build_env, lambda argv,cwd,env: prep.run(tools,env,argv,cwd))
    closure_path = args.work/'frontend-closure.json'; dump(closure_path, closure_value)
    recipe['closure'] = prep.file_pin(closure_path)
    recipe_path = args.work/'frontend-recipe.json'; dump(recipe_path, recipe)
    config['export_preparation']['frontend_build_recipe'] = prep.file_pin(recipe_path)
    inputs, authority = args.work/'inputs.json', args.work/'authority.json'
    dump(inputs, config)
    review = args.work/'SYNTHETIC-AUTHORITY.txt'
    review.write_text('SYNTHETIC TEST FIXTURE ONLY. ZERO ISSUES is a protocol test value, not an innovation-council claim. No real repository or experiment authority.\n')
    pins = [prep.file_pin(snapshot/name) for name in stages.REVIEWED]
    pins += [prep.file_pin(inputs), *config['recipe_sources'], *config['export_preparation']['tools'].values(),
             config['export_preparation']['clawpatch_tree'], config['export_preparation']['python_runtime'],
             config['export_preparation']['frontend_build_recipe'], recipe['closure'],
             config['export_preparation']['exposure_extraction']['seal'],
             *config['export_preparation']['exposure_extraction']['files'], *prep.script_pins()]
    dump(authority, {'scope':'export-preparation', 'verdict':'ZERO ISSUES', 'review':prep.file_pin(review), 'files':pins})
    budget = inputs.parent/'artifacts/preparation-budget-v7.json'
    prep.start_budget(budget, inputs, authority)
    evidence['fixture_setup_seconds'] = time.monotonic()-started
    evidence['authority'] = prep.file_pin(authority)
    evidence['inputs'] = prep.file_pin(inputs)
    evidence['budget'] = prep.file_pin(budget)
    try:
        for operation in ('build-frontend', 'ownership'):
            before = time.monotonic()
            call = SimpleNamespace(inputs=str(inputs), authorization=str(authority), budget=str(budget),
                operation=operation, repository='synthetic/production-benchmark' if operation == 'ownership' else None,
                scratch=str(args.work/'scratch'), output=str(args.work/(operation+'-output.json')),
                seal=str(args.work/(operation+'-seal.json')))
            record = prep.execute(call)
            verified = prep.verify_preparation_seal(Path(call.seal), config, inputs, authority)
            evidence['operations'].append({'operation':operation, 'status':verified['status'],
                'elapsed_seconds':time.monotonic()-before, 'seal':prep.file_pin(Path(call.seal)), 'refusal':record['refusal']})
            prep.require(record['status'] == 'completed', 'production operation refused')
        index = prep.load(Path(call.output))
        stream_path = args.work/'ownership.ndjson'
        owned_counts = []
        expected = {path.relative_to(repo).as_posix() for path in paths}
        with stream_path.open('xb') as stream:
            stream.write((json.dumps({'repo':str(repo), 'base':commits[0], 'head':commits[-1]})+'\n').encode())
            for commit in commits:
                owned = prep.load(prep.pin(index['revisions'][commit]))
                prep.require(expected <= set(owned['paths']), 'scanner construct lost planted owned source')
                owned_counts.append(len(owned['paths']))
                stream.write((json.dumps({'revision':commit, 'paths':owned['paths']})+'\n').encode())
        frontend_before = time.monotonic()
        frontend_output = prep.run(tools, env, [binary], args.work, stream_path.read_bytes())
        decoded = json.loads(frontend_output)
        prep.require(len(decoded['merges']) == revisions-1, 'frontend did not consume every merge record')
        output_path = args.work/'frontend-output.json'; output_path.write_bytes(frontend_output)
        evidence.update(frontend_seconds=time.monotonic()-frontend_before,
            frontend_output=prep.file_pin(output_path), frontend_merge_count=len(decoded['merges']),
            stream=prep.file_pin(stream_path), stream_bytes=stream_path.stat().st_size,
            owned_paths_per_revision=owned_counts, streamed_owned_paths=sum(owned_counts),
            ownership_record_bytes=sum(Path(p['path']).stat().st_size for p in index['revisions'].values()), status='completed')
    except BaseException as error:
        evidence.update(status='failed', failure={'class':type(error).__name__, 'message':str(error)})
    evidence.update(elapsed_seconds=time.monotonic()-started, ended_utc=prep.utc(),
        peak_rss_bytes=max(resource.getrusage(resource.RUSAGE_CHILDREN).ru_maxrss, resource.getrusage(resource.RUSAGE_SELF).ru_maxrss),
        rss_method='Darwin getrusage max child/self bytes; not simultaneous process-group aggregate',
        retained_work=str(args.work), cleanup='Synthetic source/repository/seals/stream retained for audit; execute cleaned fresh per-revision roots/states')
    multiplier = max(100000/(count+1), 2000000000/byte_count)*(2002/revisions)
    projection = evidence['elapsed_seconds']*multiplier
    evidence['projection'] = {'not_candidate_measurement':True, 'assumed_files_per_revision':100000,
        'assumed_bytes_per_revision':2000000000, 'assumed_revisions':2002,
        'rule':'total measured seconds * max(file ratio, byte ratio) * revision ratio; includes setup/build/verification/frontend',
        'multiplier':multiplier, 'per_repository_attempt_seconds':projection,
        'three_attempt_seconds':3*projection, 'attempt_ceiling_seconds':1209600,
        'whole_preparation_ceiling_seconds':3628800, 'max_concurrent_repositories':5,
        'attempt_fits':projection <= 1209600, 'three_attempts_fit':3*projection <= 3628800,
        'limitation':'Linear envelope projection, not a runtime guarantee; heterogeneous synthetic shapes reported separately without tuning or dropping failures.'}
    prep.require(prep.file_pin(Path(__file__)) == evidence['runner'], 'runner changed during benchmark')
    prep.require(prep.file_pin(snapshot/'prepare_exports.py') == evidence['preparation_script'], 'PREP changed during benchmark')
    dump(args.output, evidence)
    print(json.dumps({'output':str(args.output),'status':evidence['status'],'elapsed_seconds':evidence['elapsed_seconds'],
                     'projection':evidence['projection']}))
    if evidence['status'] != 'completed':
        raise SystemExit(1)


if __name__ == '__main__':
    main()

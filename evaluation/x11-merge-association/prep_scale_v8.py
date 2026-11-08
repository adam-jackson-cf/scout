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
from fractions import Fraction
from types import SimpleNamespace

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))


def dump(path, value):
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open('x') as stream:
        json.dump(value, stream, sort_keys=True, indent=2)
        stream.write('\n')


TS_SPEC = {'source_files': 81400, 'revision_count': 3, 'churn_denominator': 20,
           'minimum_normal_code_bytes': 4800, 'large_code_files': 4,
           'minimum_large_code_bytes': 1050000, 'test_fraction': '1/2',
           'tests_tree_fraction_of_tests': '1/2', 'minimum_directory_depth': 3,
           'generator': 'nested-typescript-code-v8'}


def ts_path(position):
    half = TS_SPEC['source_files']//2
    index = position if position < half else position-half
    root = 'src' if position < half or index >= half//2 else 'tests'
    suffix = '.ts' if position < half else '.test.ts'
    return Path(root)/('domain%03d' % (index//1000))/('unit%03d' % (index//100))/('file%05d%s' % (index, suffix))


def ts_source(position, revision):
    target = ts_path(1 if position == 0 else 0)
    relative = os.path.relpath(target, ts_path(position).parent).replace(os.sep, '/')
    if not relative.startswith('.'):
        relative = './'+relative
    relative = relative[:-3]+'.js'
    head = ("import { item as dependency } from '%s';\nexport function item(): number {\n"
            "  let value = %03d;\n" % (relative, revision)).encode()
    tail = b'  return value + dependency();\n}\n'
    line = b'  value = (value + 1) % 104729;\n'
    minimum = 1050000 if position < 4 else 4800
    return head + line*((minimum-len(head)-len(tail)+len(line)-1)//len(line)) + tail


def revision_trees(context, prep):
    tools = {k:Path(v) for k,v in context['tools'].items()}
    result = []
    for commit in context['commits']:
        raw = prep.git(tools, context['env'], Path(context['repo']), 'ls-tree', '-r', '-l', '-z', commit)
        entries = [entry.split(b'\t', 1)[0].split() for entry in raw.split(b'\0') if entry]
        prep.require(all(e[1] == b'blob' for e in entries), 'unexpected synthetic tree entry')
        result.append({'revision':commit, 'files':len(entries), 'bytes':sum(int(e[3]) for e in entries),
                       'ls_tree_sha256':hashlib.sha256(raw).hexdigest()})
    return result


def freeze(args, context, prep):
    from preparation_capacity_v8 import observations, fit
    before = time.monotonic()
    rows = observations(args.go_receipts)
    coefficients = fit(rows)
    trees = revision_trees(context, prep)
    fixture = context['evidence']['fixture']
    prep.require(all(t['files'] == fixture['files_per_revision'] and
                     t['bytes'] == fixture['bytes_per_revision_min'] for t in trees), 'actual revision size drift')
    repo = Path(context['repo'])
    manifest = [dict(prep.file_pin(p), relative=p.relative_to(repo).as_posix(), bytes=p.stat().st_size)
                for p in sorted(repo.rglob('*')) if p.is_file() and '.git' not in p.relative_to(repo).parts]
    manifest_path = args.work/'fixture-manifest.json'
    dump(manifest_path, manifest)
    context['fixture_freeze_seconds'] = time.monotonic()-before
    state_path = args.work/'frozen-context.json'
    dump(state_path, context)
    prediction = coefficients[0]*fixture['files_per_revision']+coefficients[1]*fixture['bytes_per_revision_min']+coefficients[2]
    frozen = {'schema':'x11a-heldout-prediction-v8', 'created_utc':prep.utc(), 'pre_run':True,
        'fixture':fixture, 'fixture_spec':TS_SPEC, 'revision_trees':trees,
        'manifest':prep.file_pin(manifest_path), 'context':prep.file_pin(state_path),
        'runner':prep.file_pin(Path(__file__)), 'original_inputs':prep.file_pin(args.inputs),
        'go_observations':rows, 'coefficients_exact':list(map(str, coefficients)),
        'prediction_seconds_per_revision_exact':str(prediction),
        'prediction_seconds_per_revision':float(prediction),
        'work':str(args.work), 'output':str(args.output),
        'rule':'Create once before scanner work; no refit, rerun, fixture or ceiling changes.'}
    dump(args.prediction, frozen)
    print(json.dumps({'prediction':prep.file_pin(args.prediction), 'fixture':fixture,
                      'prediction_seconds_per_revision':float(prediction), 'status':'frozen-not-run'}))


def run_frozen(args):
    before = time.monotonic()
    import prepare_exports as prep
    prep.require(prep.file_pin(args.prediction)['sha256'] == args.prediction_sha256, 'frozen prediction digest mismatch')
    frozen = prep.load(args.prediction)
    prep.require(frozen['schema'] == 'x11a-heldout-prediction-v8' and frozen['pre_run'] is True and
                 frozen['fixture_spec'] == TS_SPEC and frozen['work'] == str(args.work) and
                 frozen['output'] == str(args.output), 'frozen fixture specification/path mismatch')
    dump(args.work/'heldout-invocation.json', {'command':list(sys.orig_argv), 'prediction':prep.file_pin(args.prediction),
                                             'started_utc':prep.utc(), 'rule':'one invocation, retained on any failure'})
    prep.require(not args.output.exists(), 'heldout receipt already exists')
    prep.require(prep.file_pin(Path(__file__)) == frozen['runner'], 'runner changed after prediction freeze')
    prep.require(prep.file_pin(args.inputs) == frozen['original_inputs'], 'original inputs changed after freeze')
    context = prep.load(prep.pin(frozen['context']))
    prep.require(Path(sys.executable).resolve() == Path(context['tools']['python']), 'pinned Python required')
    evidence = context['evidence']
    for pair in evidence['original_snapshot_pairs']:
        prep.pin(pair['original']); prep.pin(pair['snapshot'])
    for item in evidence['source_snapshot']:
        prep.pin(item)
    for key in ('inputs', 'authority'):
        prep.pin(evidence[key])
    manifest = prep.load(prep.pin(frozen['manifest']))
    repo = Path(context['repo'])
    actual = {p.relative_to(repo).as_posix() for p in repo.rglob('*')
              if p.is_file() and '.git' not in p.relative_to(repo).parts}
    prep.require(actual == {item['relative'] for item in manifest}, 'fixture working file set drift')
    for item in manifest:
        prep.require(Path(item['path']) == repo/item['relative'], 'fixture manifest path mismatch')
        prep.pin({'path':item['path'], 'sha256':item['sha256']})
        prep.require(Path(item['path']).stat().st_size == item['bytes'], 'fixture byte size drift')
    prep.require(revision_trees(context, prep) == frozen['revision_trees'], 'frozen revision digests/sizes drift')
    prep.require(evidence['fixture'] == frozen['fixture'], 'fixture metadata drift')
    from preparation_capacity_v8 import observations, fit
    rows = observations([Path(row['receipt']['path']) for row in frozen['go_observations']])
    coefficients = fit(rows)
    fixture = frozen['fixture']
    prediction = coefficients[0]*fixture['files_per_revision']+coefficients[1]*fixture['bytes_per_revision_min']+coefficients[2]
    prep.require(rows == frozen['go_observations'] and list(map(str, coefficients)) == frozen['coefficients_exact'] and
                 str(prediction) == frozen['prediction_seconds_per_revision_exact'], 'frozen prediction does not reproduce')
    snapshot = Path(context['snapshot'])
    for name in list(sys.modules):
        module_path = getattr(sys.modules[name], '__file__', None)
        if name != '__main__' and module_path and Path(module_path).resolve().parent == HERE:
            del sys.modules[name]
    sys.path.insert(0, str(snapshot))
    prep = importlib.import_module('prepare_exports')
    prep.start_budget(Path(context['budget']), Path(context['inputs']), Path(context['authority']))
    evidence['budget'] = prep.file_pin(Path(context['budget']))
    evidence['prediction'] = prep.file_pin(args.prediction)
    evidence['command'] = list(sys.orig_argv)
    evidence['started_utc'] = prep.utc()
    return execute_stage(args, context, prep, evidence['fixture_setup_seconds']+
                         context['fixture_freeze_seconds']+time.monotonic()-before)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--inputs', type=Path, required=True)
    parser.add_argument('--work', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--scenario', choices=['files', 'bytes', 'construct', 'heldout-ts'], required=True)
    parser.add_argument('--phase', choices=['freeze', 'run'], default='run')
    parser.add_argument('--prediction', type=Path)
    parser.add_argument('--prediction-sha256')
    parser.add_argument('--go-receipts', nargs=3, type=Path)
    args = parser.parse_args()
    for value in (args.inputs, args.work, args.output):
        if not value.is_absolute():
            parser.error('absolute paths required')
    if args.scenario == 'heldout-ts':
        if not args.prediction or not args.prediction.is_absolute():
            parser.error('heldout-ts requires absolute --prediction')
        if args.phase == 'run':
            if not args.prediction_sha256:
                parser.error('heldout run requires pre-recorded --prediction-sha256')
            return run_frozen(args)
        if not args.go_receipts:
            parser.error('freeze requires three final-PREP --go-receipts')
    elif args.phase != 'run':
        parser.error('freeze is only for heldout-ts')
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
    evidence = {'schema': 'x11a-preparation-scale-v8', 'synthetic_only': True,
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
    if args.scenario == 'heldout-ts':
        count, revisions, padding = TS_SPEC['source_files'], 3, 0
        extras = {'package.json': b'{"name":"synthetic-capacity","private":true,"scripts":{"test":"node --test"}}\n',
                  'tsconfig.json': b'{"compilerOptions":{"target":"ES2022","module":"NodeNext","moduleResolution":"NodeNext"},"include":["**/*.ts"]}\n'}
        for name, data in extras.items():
            (repo/name).write_bytes(data)
        paths = [repo/ts_path(position) for position in range(count)]
        source = ts_source
    else:
        count, revisions, padding = (80000, 20, 0) if args.scenario == 'files' else (2000, 3, 1000000)
        if args.scenario == 'construct':
            count, revisions, padding = 40, 3, 0
        extras = {'go.mod': b'module example.invalid/scalesynthetic\n\ngo 1.24\n'}
        (repo/'go.mod').write_bytes(extras['go.mod'])
        paths = [repo/'internal'/('component%03d' % (position//1000))/('file%05d.go' % position) for position in range(count)]
        def source(position, revision):
            return ('package component%03d\nfunc Item%05d() int { return %d }\n' %
                    (position//1000, position, revision)).encode() + (b'/*'+b'x'*padding+b'*/\n' if padding else b'')
    for position, path in enumerate(paths):
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(source(position, 0))
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
    byte_count = sum(path.stat().st_size for path in paths) + sum(map(len, extras.values()))
    evidence['fixture'] = {'files_per_revision': count+len(extras), 'source_files_per_revision': count,
        'bytes_per_revision_min': byte_count, 'revision_count': revisions,
        'changed_source_files_each_revision_after_base': changed, 'change_fraction': changed/count,
        'language': 'Go', 'padding_bytes_per_source_file': padding,
        'shape': 'package/function sources; byte-heavy scenario adds one block comment per source; no candidate-informed complexity tuning',
        'first_parent_revisions': commits, 'merge_revisions': commits[1:]}
    if args.scenario == 'heldout-ts':
        prep.require(count+len(extras) >= 81369 and byte_count >= 380592869, 'heldout size floor not met')
        evidence['fixture'].update(language='TypeScript', padding_bytes_per_source_file=0,
            shape='Nested >=3-level TS imports/exports/functions; exactly half source files tests, equally split tests/ tree and sibling *.test.ts; four >=1MB executable-code files, no padding comments',
            source_test_files=count//2, tests_tree_files=count//4, sibling_test_files=count//4,
            large_code_files=4, minimum_large_file_bytes=1050000,
            fixture_spec=TS_SPEC, pre_run_frozen=True,
            ownership_construct='Node mapper excludes tests from reviewable owned source groups; require genuine nonempty source ownership, not ownership of every test path. Report actual source/test ownership counts.',
            association_limit='Pinned mapper source packageTestFiles caps association candidates at 200; 40700 test files exercise discovery but do not imply groups times all40700 association filters.')
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
    budget = prep.canonical_budget(inputs)
    if args.scenario != 'heldout-ts':
        prep.start_budget(budget, inputs, authority)
    evidence['fixture_setup_seconds'] = time.monotonic()-started
    evidence['authority'] = prep.file_pin(authority)
    evidence['inputs'] = prep.file_pin(inputs)
    if args.scenario != 'heldout-ts':
        evidence['budget'] = prep.file_pin(budget)
    context = dict(evidence=evidence, repo=str(repo), paths=[str(p) for p in paths], commits=commits,
                   tools={k:str(v) for k,v in tools.items()}, env=env, inputs=str(inputs), authority=str(authority),
                   budget=str(budget), binary=str(binary), config=config, count=count, revisions=revisions,
                   byte_count=byte_count, extras={k:len(v) for k,v in extras.items()}, snapshot=str(snapshot))
    if args.scenario == 'heldout-ts':
        return freeze(args, context, prep)
    return execute_stage(args, context, prep, time.monotonic()-started)


def execute_stage(args, context, prep, setup_seconds):
    evidence = context['evidence']
    repo, inputs, authority, budget, binary, snapshot = (Path(context[k]) for k in ('repo','inputs','authority','budget','binary','snapshot'))
    paths = [Path(p) for p in context['paths']]
    commits, env, config = (context[k] for k in ('commits','env','config'))
    tools = {k:Path(v) for k,v in context['tools'].items()}
    count, revisions, byte_count, extras = (context[k] for k in ('count','revisions','byte_count','extras'))
    started = time.monotonic()-setup_seconds
    try:
        for operation in ('build-frontend', 'ownership'):
            before = time.monotonic()
            output, seal = prep.canonical_paths(inputs, operation, 'synthetic/production-benchmark' if operation == 'ownership' else None)
            call = SimpleNamespace(inputs=str(inputs), authorization=str(authority), budget=str(budget),
                operation=operation, repository='synthetic/production-benchmark' if operation == 'ownership' else None,
                scratch=str(args.work/'scratch'), output=str(output), seal=str(seal))
            record = prep.execute(call)
            verified = prep.verify_preparation_seal(Path(call.seal), config, inputs, authority)
            evidence['operations'].append({'operation':operation, 'status':verified['status'],
                'elapsed_seconds':time.monotonic()-before, 'seal':prep.file_pin(Path(call.seal)), 'refusal':record['refusal']})
            prep.require(record['status'] == 'completed', 'production operation refused')
        index = prep.load(Path(call.output))
        stream_path = args.work/'ownership.ndjson'
        owned_counts = []
        source_owned_counts, test_owned_counts = [], []
        expected = {path.relative_to(repo).as_posix() for path in paths}
        with stream_path.open('xb') as stream:
            stream.write((json.dumps({'repo':str(repo), 'base':commits[0], 'head':commits[-1]})+'\n').encode())
            for commit in commits:
                owned = prep.load(prep.pin(index['revisions'][commit]))
                if args.scenario == 'heldout-ts':
                    sources = {path.relative_to(repo).as_posix() for path in paths[:count//2]}
                    tests = expected-sources
                    source_owned_counts.append(len(sources & set(owned['paths'])))
                    test_owned_counts.append(len(tests & set(owned['paths'])))
                    prep.require(source_owned_counts[-1] > 0, 'heldout Node construct has no genuine source ownership')
                else:
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
        if args.scenario == 'heldout-ts':
            evidence.update(owned_source_paths_per_revision=source_owned_counts,
                            owned_test_paths_per_revision=test_owned_counts)
    except BaseException as error:
        evidence.update(status='failed', failure={'class':type(error).__name__, 'message':str(error)})
    evidence.update(elapsed_seconds=time.monotonic()-started, ended_utc=prep.utc(),
        peak_rss_bytes=max(resource.getrusage(resource.RUSAGE_CHILDREN).ru_maxrss, resource.getrusage(resource.RUSAGE_SELF).ru_maxrss),
        rss_method='Darwin getrusage max child/self bytes; not simultaneous process-group aggregate',
        retained_work=str(args.work), cleanup='Synthetic source/repository/seals/stream retained for audit; execute cleaned fresh per-revision roots/states')
    if args.scenario != 'heldout-ts':
        multiplier = max(100000/(count+len(extras)), 2000000000/byte_count)*(2002/revisions)
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
                     'projection':evidence.get('projection')}))
    if evidence['status'] != 'completed':
        raise SystemExit(1)


if __name__ == '__main__':
    main()

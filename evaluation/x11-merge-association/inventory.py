#!/usr/bin/env python3
"""X11a model-free inventory; product Go frontend, no labels/model/decisions.

Source inputs and generated exports have separate reviewed digest bindings.
This module never scans/exports or fetches automatically.
"""
import argparse
from collections import Counter
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parent
POPULATIONS = ('all', 'non_test', 'added_only')
SHA = re.compile(r'[0-9a-f]{40}')

GIT_BINARY = None
EMPTY_BLOB = 'e69de29bb2d1d6434b8b29ae775ad8c2e48c5391'


def exact_keys(value, required, optional=()):
    if not isinstance(value, dict) or not set(required) <= set(value) or set(value) - set(required) - set(optional):
        raise ValueError('unexpected or missing schema keys')


def git_env():
    environment = dict(os.environ, GIT_NO_LAZY_FETCH='1', GIT_TERMINAL_PROMPT='0',
        GIT_CONFIG_NOSYSTEM='1', GIT_CONFIG_GLOBAL='/dev/null', GIT_CONFIG_COUNT='0',
        GIT_NO_REPLACE_OBJECTS='1', LC_ALL='C')
    environment.pop('GIT_CONFIG_PARAMETERS', None)
    if GIT_BINARY:
        root = Path(GIT_BINARY).resolve().parent.parent
        environment['GIT_EXEC_PATH'] = str(root / 'libexec/git-core')
        environment['GIT_TEMPLATE_DIR'] = str(root / 'share/git-core/templates')
    return environment


def verify_python(config):
    from pin_tool_closure import verify_python_runtime
    verify_python_runtime(config)


def configure_git(config):
    global GIT_BINARY
    tool = config['export_preparation']['tools']['git']
    if digest(tool['path']) != tool['sha256']:
        raise ValueError('git executable mismatch')
    GIT_BINARY = tool['path']
    if git('.', '--version').decode().strip() != config['export_preparation']['git_version']:
        raise ValueError('git version mismatch')

def digest(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def reject_pending(value):
    if isinstance(value, str) and 'PENDING' in value.upper():
        raise ValueError('pending token in pinned input')
    if isinstance(value, dict):
        for k, v in value.items():
            reject_pending(k)
            reject_pending(v)
    if isinstance(value, list):
        for v in value:
            reject_pending(v)


def git(repo, *args):
    if GIT_BINARY is None:
        raise ValueError('pinned git not configured')
    return subprocess.check_output([GIT_BINARY, '-C', str(repo), *args], stderr=subprocess.PIPE, env=git_env())


def rename_detection_skipped(stderr):
    text = stderr.decode('utf-8', errors='replace').lower()
    return ('rename detection was skipped' in text or 'inexact rename detection was skipped' in text
        or 'too many files' in text and 'rename' in text)


def check_rename_detection(repo, parent, sha):
    result = subprocess.run([GIT_BINARY, '-C', str(repo), 'diff-tree', '-r', '--raw',
        '--no-abbrev', '-M50%', parent, sha], stdout=subprocess.DEVNULL, stderr=subprocess.PIPE,
        env=git_env(), check=True)
    return rename_detection_skipped(result.stderr)


def role(path):
    p = path.lower()
    parts = p.split('/')
    name = parts[-1]
    if (any(s in ('test', 'tests', 'testdata', 'baselines') for s in parts[:-1])
        or name.endswith(('_test.go', '_test.py')) or (name.startswith('test_') and name.endswith('.py'))
        or re.search(r'\.(?:test|spec)\.(?:tsx?|jsx?)$', name)):
        return 'test'
    if any(s in ('example', 'examples', 'docs_src') for s in parts):
        return 'example'
    if parts[0] in ('docs', 'doc'):
        return 'documentation-non-markdown'
    if parts[0] in ('.github', 'scripts', 'toolchain'):
        return 'infrastructure'
    return 'other-source-config'


def wilson(x, n):
    if not n:
        return None
    z = 1.959963984540054
    den = 1 + z*z/n
    center = (x/n + z*z/(2*n))/den
    half = z*((x/n*(1-x/n)/n + z*z/(4*n*n))**0.5)/den
    return [center-half, center+half]


def load_pinned(entry, base):
    if set(entry) != {'path', 'sha256'} or not re.fullmatch('[0-9a-f]{64}', entry['sha256']):
        raise ValueError('invalid pinned artifact entry')
    p = Path(entry['path'])
    p = p if p.is_absolute() else base / p
    if digest(p) != entry['sha256']:
        raise ValueError('artifact digest mismatch: ' + str(p))
    return json.loads(p.read_text())


def validate_heuristic(value, repository, scanner_sha, scanner_tree_sha):
    exact_keys(value, {'schema', 'source', 'repository', 'scanner_sha256', 'scanner_tree_sha256',
        'revision_order', 'revisions'})
    if value['schema'] != 'x11a-ownership-index-v6':
        raise ValueError('ownership digest index required')
    if value['scanner_tree_sha256'] != scanner_tree_sha:
        raise ValueError('scanner tree provenance mismatch')
    if value['source'] != 'heuristic' or value['repository'] != repository or value['scanner_sha256'] != scanner_sha:
        raise ValueError('heuristic provenance mismatch')
    if not isinstance(value['revisions'], dict) or not value['revisions']:
        raise ValueError('empty heuristic revisions')
    if (not isinstance(value['revision_order'], list) or len(value['revision_order']) > 2002
            or len(set(value['revision_order'])) != len(value['revision_order'])
            or set(value['revision_order']) != set(value['revisions'])):
        raise ValueError('invalid revision order')
    for sha, entry in value['revisions'].items():
        if not SHA.fullmatch(sha):
            raise ValueError('invalid heuristic revision')
        validate_pin(entry)


def ownership_record(index, revision):
    record = load_pinned(index['revisions'][revision], ROOT)
    exact_keys(record, {'revision', 'paths', 'excluded_codegraph_entries', 'gitlinks',
                       'entry_count', 'blob_count', 'oid_list_sha256', 'tree_list_sha256',
                       'batch_stdout_sha256'})
    paths = record['paths']
    if (record['revision'] != revision or not isinstance(paths, list)
            or not all(isinstance(path, str) for path in paths) or paths != sorted(set(paths))):
        raise ValueError('invalid ownership record')
    from prepare_exports import relative_path
    for path in paths:
        relative_path(path)
    if (any(type(record[key]) is not int or record[key] < 0
            for key in ('excluded_codegraph_entries', 'gitlinks', 'entry_count', 'blob_count'))
            or any(not isinstance(record[key], str) or not re.fullmatch('[0-9a-f]{64}', record[key])
                   for key in ('oid_list_sha256', 'tree_list_sha256', 'batch_stdout_sha256'))):
        raise ValueError('invalid materialization record')
    return record


def validate_exposure(value, studies):
    if set(value) != {'studies', 'records'} or not isinstance(value['studies'], list) or set(value['studies']) != set(studies):
        raise ValueError('incomplete exposure enumeration')
    if not isinstance(value['records'], list):
        raise ValueError('invalid exposure record collection')
    for r in value['records']:
        allowed = {'repository', 'study', 'task', 'commit', 'path', 'blob'}
        if not isinstance(r, dict) or not set(r) <= allowed or not {'repository', 'study', 'task'} <= set(r):
            raise ValueError('non-identifier or missing exposure keys')
        if not all(isinstance(v, str) and v for v in r.values()) or r['study'] not in studies:
            raise ValueError('exposure value/study mismatch')
        if r['task'] not in ('association', 'other') or not {'commit', 'path', 'blob'} & set(r):
            raise ValueError('invalid task or no identifier')
        for k in ('commit', 'blob'):
            if k in r and not SHA.fullmatch(r[k]):
                raise ValueError('non-full exposure SHA')
        if 'path' in r and (r['path'].startswith('/') or '\\' in r['path'] or '\0' in r['path']
                or any(p in ('', '.', '..', '.codegraph') for p in r['path'].split('/'))):
            raise ValueError('unsafe exposure identifier path')


def validate_unscoped(value, studies):
    exact_keys(value, {'records'})
    for record in value['records']:
        exact_keys(record, {'source', 'study', 'kind', 'identifier', 'task', 'matchability'})
        if not all(isinstance(v, str) and v for v in record.values()) or record['study'] not in studies:
            raise ValueError('invalid unscoped identifier')
        kind = record['kind']
        if kind not in ('commit', 'blob', 'path', 'pr') or record['task'] not in ('association', 'other'):
            raise ValueError('invalid unscoped kind/task')
        expected = 'repository_independent_sha' if kind in ('commit', 'blob') else 'unmatchable_without_repository'
        if record['matchability'] != expected or kind in ('commit', 'blob') and not SHA.fullmatch(record['identifier']):
            raise ValueError('invalid unscoped matchability')


def unscoped_matches(record, sha, blobs):
    kind, identifier = record['kind'], record['identifier']
    return (kind == 'commit' and identifier == sha or
        kind == 'blob' and identifier != EMPTY_BLOB and identifier in blobs)


def pooled_accuracy_support(age_points, pooled_point, pooled_lower_bound, adequacy_gates):
    """Age points are descriptive; only pooled accuracy has a threshold."""
    exact_keys(adequacy_gates, {'quality', 'denominators', 'missingness', 'cap', 'interval_validity'})
    return (len(age_points) == 2 and all(value is True for value in adequacy_gates.values())
        and pooled_point >= .80 and pooled_lower_bound >= .80)


def projection_vectors(rows, repository):
    """Fixed model-free OC input, not observed replay/label supply."""
    vectors = {age: [] for age in ('Cold', 'C0')}
    label_candidates = {age: [] for age in vectors}
    population_vectors = {population: [] for population in POPULATIONS}
    resource_selected = sum(len(row['selected'] or []) for row in rows)
    for position, row in enumerate(rows):
        sizes = Counter()
        kept = Counter()
        if not row['association_exposed']:
            for cluster in row['selected'] or []:
                kinds = {member['kind'] for member in cluster['members']}
                sizes['all'] += 1
                sizes['non_test'] += not any(role(member['path']) == 'test' for member in cluster['members'])
                sizes['added_only'] += kinds == {'added'}
                for age in vectors:
                    probability = (0 if 'modified' in kinds else
                        (90 if age == 'Cold' else 80) if kinds == {'added'} else
                        (75 if age == 'Cold' else 50) if 'renamed' in kinds else 0)
                    key = '\0'.join(('x11a-project-v4', repository, age, row['commit'], cluster['id']))
                    draw = int.from_bytes(hashlib.sha256(key.encode('utf-8')).digest(), 'big')
                    if draw * 100 < probability * (1 << 256):
                        kept[age] += 1
                        case_key = '\0'.join((repository, age, row['commit'], cluster['id']))
                        label_hash = hashlib.sha256(('x11a-label-v5\0' + case_key).encode()).digest()
                        directory_hash = hashlib.sha256(
                            ('x11a-directory-v6\0' + repository + '\0' + cluster['directory']).encode('utf-8')).hexdigest()
                        label_candidates[age].append((label_hash, case_key, position, directory_hash))
        for population in population_vectors:
            population_vectors[population].append(sizes[population])
        for age in vectors:
            vectors[age].append(kept[age])
    label_vectors = {age: [0] * len(rows) for age in vectors}
    directory_cases = {}
    largest_share = {}
    directory_counts, effective_counts, retained_counts = {}, {}, {}
    for age, candidates in label_candidates.items():
        chosen = sorted(candidates)[:2600]
        directory_cases[age] = [{'position': position, 'directory_hash': directory_hash}
            for _, case_key, position, directory_hash in sorted(chosen, key=lambda row: (row[2], row[1]))]
        for case in directory_cases[age]:
            label_vectors[age][case['position']] += 1
        sizes = Counter(case['directory_hash'] for case in directory_cases[age])
        largest_share[age] = max(sizes.values(), default=0) / len(chosen) if chosen else None
        directory_counts[age] = len(sizes)
        effective_counts[age] = len(chosen)**2 / sum(n*n for n in sizes.values()) if chosen else 0
        retained_counts[age] = len(chosen)
    return {'commit_order': [row['commit'] for row in rows],
        'population_selected_counts': population_vectors, 'projected_selected_counts': vectors,
        'projected_label_selected_counts': label_vectors,
        'projected_label_directory_cases': directory_cases,
        'largest_retained_directory_share': largest_share,
        'retained_directory_count': directory_counts,
        'effective_directory_count': effective_counts,
        'retained_directory_case_count': retained_counts,
        'resource_proxy_selected_including_exposed': resource_selected,
        'label_proxy_upper_bound_per_cell': min(2600, sum(population_vectors['all'])),
        'projected_label_capacity_per_cell': {age: min(2600, sum(vector)) for age, vector in vectors.items()}}


def validate_pin(pin):
    exact_keys(pin, {'path', 'sha256'})
    if (not isinstance(pin['path'], str) or not pin['path']
            or not isinstance(pin['sha256'], str) or not re.fullmatch('[0-9a-f]{64}', pin['sha256'])):
        raise ValueError('invalid pin value')


def validate_inputs(config):
    exact_keys(config, {'schema', 'status', 'pin_resolution_rule', 'start_epoch', 'end_epoch',
        'repositories', 'exposure_studies', 'observation_rule', 'clone_procedure',
        'export_preparation', 'recipe_sources'})
    if config['schema'] != 'x11a-inventory-source-inputs-v4' or config['status'] != 'PINNED':
        raise ValueError('source inputs not v4 PINNED')
    for field in ('pin_resolution_rule', 'observation_rule', 'clone_procedure'):
        if not isinstance(config[field], str):
            raise ValueError('invalid source text')
    if any(type(config[field]) is not int for field in ('start_epoch', 'end_epoch')):
        raise ValueError('invalid source epoch')
    if (not isinstance(config['repositories'], list) or not config['repositories']
            or not isinstance(config['recipe_sources'], list)):
        raise ValueError('invalid source collections')
    if (not isinstance(config['exposure_studies'], list)
            or not all(isinstance(study, str) and study for study in config['exposure_studies'])
            or len(config['exposure_studies']) != len(set(config['exposure_studies']))):
        raise ValueError('invalid study enumeration')
    for item in config['repositories']:
        exact_keys(item, {'repository', 'path', 'default_branch', 'resolved_remote_tip',
            'tip', 'Cold', 'C0', 'canonical_clone_url'})
        if not all(isinstance(v, str) and v for v in item.values()):
            raise ValueError('invalid repository fields')
        if not all(SHA.fullmatch(item[field]) for field in ('resolved_remote_tip', 'tip', 'Cold', 'C0')):
            raise ValueError('invalid repository SHA')
    prep = config['export_preparation']
    exact_keys(prep, {'status', 'clawpatch_revision', 'clawpatch_version', 'tools', 'git_version',
        'exposure_availability', 'exposure_extraction', 'clawpatch_tree', 'frontend_build_recipe',
        'python_runtime'})
    exact_keys(prep['tools'], {'git', 'node', 'clawpatch', 'clawpatch_safety_marker',
        'sandbox_exec', 'go', 'python'})
    if not all(isinstance(prep[field], str) for field in ('status', 'clawpatch_revision',
            'clawpatch_version', 'git_version', 'exposure_availability')) or prep['status'] != 'PINNED':
        raise ValueError('invalid preparation fields')
    for pin in [*prep['tools'].values(), *config['recipe_sources'],
                prep['clawpatch_tree'], prep['frontend_build_recipe'], prep['python_runtime']]:
        validate_pin(pin)
    extraction = prep['exposure_extraction']
    exact_keys(extraction, {'seal', 'scout_commit', 'inertia_commit', 'files'})
    if (not all(isinstance(extraction[field], str) and SHA.fullmatch(extraction[field])
                for field in ('scout_commit', 'inertia_commit')) or not isinstance(extraction['files'], list)):
        raise ValueError('invalid extraction fields')
    for pin in [extraction['seal'], *extraction['files']]:
        validate_pin(pin)
    reject_pending(config)


def validate_artifacts(value, repositories):
    exact_keys(value, {'status', 'source_inputs_sha256', 'preparation_authorization',
        'exposure', 'unscoped_exposure', 'repository_exposure', 'exposure_studies', 'frontend',
        'ownership', 'frontend_build', 'preparation_seals'})
    if value['status'] != 'SEALED' or not isinstance(value['source_inputs_sha256'], str) or not re.fullmatch('[0-9a-f]{64}', value['source_inputs_sha256']):
        raise ValueError('invalid artifact status/source digest')
    if not isinstance(value['exposure_studies'], list) or not all(isinstance(study, str) for study in value['exposure_studies']):
        raise ValueError('invalid artifact studies')
    exact_keys(value['ownership'], set(repositories))
    exact_keys(value['frontend_build'], {'recipe', 'closure', 'buildinfo'})
    if not isinstance(value['preparation_seals'], list) or not value['preparation_seals']:
        raise ValueError('missing preparation seals')
    for union in value['ownership'].values():
        exact_keys(union, {'status', 'seal'})
        if union['status'] not in ('completed', 'refused'):
            raise ValueError('invalid ownership union status')
        validate_pin(union['seal'])
    for pin in [value['preparation_authorization'], value['exposure'], value['unscoped_exposure'],
                value['repository_exposure'], value['frontend'], *value['preparation_seals'],
                *value['frontend_build'].values()]:
        validate_pin(pin)
    reject_pending(value)


def repository_exposure_by_study(value, repository, studies):
    """Repository identity only: descriptive, never commit overlap or exclusion."""
    exact_keys(value, {'records'})
    counts = {study: 0 for study in studies}
    for record in value['records']:
        exact_keys(record, {'repository', 'study', 'task'}, {'pr', 'repository_identity'})
        if record['study'] not in counts or record['task'] not in ('association', 'other'):
            raise ValueError('invalid additional exposure record')
        if record['repository'] == repository and 'repository_identity' in record:
            counts[record['study']] += 1
    return counts


def verify_artifact_bindings(value, inputs_path, artifact_base):
    """Adopt only outputs derived from reviewed, verified preparation operations."""
    from prepare_exports import verify_preparation_seal
    config = json.loads(Path(inputs_path).read_text())
    def located(pin):
        verify_pin = Path(pin['path'])
        return verify_pin if verify_pin.is_absolute() else artifact_base / verify_pin
    authority = located(value['preparation_authorization'])
    if digest(authority) != value['preparation_authorization']['sha256']:
        raise ValueError('preparation authority digest mismatch')
    operations = {}
    for pin in value['preparation_seals']:
        path = located(pin)
        if digest(path) != pin['sha256']:
            raise ValueError('preparation seal digest mismatch')
        seal = verify_preparation_seal(path, config, inputs_path, authority)
        if seal['operation'] in operations or seal['status'] != 'completed':
            raise ValueError('duplicate or incomplete preparation dependency')
        operations[seal['operation']] = seal
    if set(operations) != {'build-frontend', 'exposures'}:
        raise ValueError('missing exact preparation dependencies')
    build = load_pinned(operations['build-frontend']['output'], Path(inputs_path).parent)
    exposures = load_pinned(operations['exposures']['output'], Path(inputs_path).parent)
    exact_keys(build, {'frontend', 'frontend_build'})
    exact_keys(exposures, {'exposure', 'unscoped_exposure', 'repository_exposure', 'exposure_studies'})
    if (build['frontend'] != value['frontend'] or build['frontend_build'] != value['frontend_build']
            or any(exposures[key] != value[key] for key in
                ('exposure', 'unscoped_exposure', 'repository_exposure', 'exposure_studies'))):
        raise ValueError('artifact manifest differs from sealed preparation outputs')
    ownership = {}
    for repository, union in value['ownership'].items():
        path = located(union['seal'])
        if digest(path) != union['seal']['sha256']:
            raise ValueError('ownership seal digest mismatch')
        seal = verify_preparation_seal(path, config, inputs_path, authority)
        if (seal['operation'] != 'ownership' or seal['repository'] != repository
                or seal['status'] != union['status']):
            raise ValueError('ownership union differs from seal')
        if (seal['status'] == 'refused' and seal['refusal']['class'] == 'operational'
                and seal['refusal']['code'] not in ('PreparationInfeasibleOperational',
                    'PreparationInterruptedOperational', 'PreparationBudgetExhausted')):
            raise ValueError('nonterminal operational preparation refusal')
        ownership[repository] = seal
    return ownership


def annotate_exposure(rows, scoped, unscoped, observe):
    """Resolve identifier records once per candidate, retaining observed-commit matches."""
    records = [('scoped', r) for r in scoped] + [('unscoped', r) for r in unscoped]
    lookup = {kind: {} for kind in ('commit', 'blob', 'path')}
    for index, (scope, record) in enumerate(records):
        identifiers = ({record['kind']: record['identifier']} if scope == 'unscoped'
            else {kind: record[kind] for kind in lookup if kind in record})
        for kind, identifier in identifiers.items():
            if kind not in lookup or scope == 'unscoped' and kind == 'path' or kind == 'blob' and identifier == EMPTY_BLOB:
                continue
            lookup[kind].setdefault(identifier, set()).add(index)
    resolved = set()
    other_matches, excluded = [], []
    for row in rows:
        sha = row['commit']
        paths, blobs = observe(row)
        row['association_exposed'] = False
        hits = set(lookup['commit'].get(sha, ()))
        for kind, identifiers in (('path', paths), ('blob', blobs)):
            for identifier in identifiers:
                hits.update(lookup[kind].get(identifier, ()))
        for index in sorted(hits):
            scope, record = records[index]
            resolved.add(index)
            association = record['task'] == 'association' or record['study'].startswith('x2d-')
            if association:
                row['association_exposed'] = True
            else:
                other_matches.append({'observed_commit': sha, 'scope': scope, 'identifier': record})
        if row['association_exposed']:
            excluded.append(sha)
    summary = {}
    for index, (scope, record) in enumerate(records):
        key = record['study']
        counts = summary.setdefault(key, {'resolved': 0, 'unresolved': 0, 'by_kind': {}})
        kind = record['kind'] if scope == 'unscoped' else '+'.join(k for k in ('commit', 'blob', 'path') if k in record)
        bucket = counts['by_kind'].setdefault(kind, {'resolved': 0, 'unresolved': 0})
        status = 'resolved' if index in resolved else 'unresolved'
        counts[status] += 1
        bucket[status] += 1
    return excluded, other_matches, summary


def images(repo, parent, sha):
    fields = git(repo, 'diff-tree', '-r', '-z', '--raw', '--no-abbrev', '-M50%', parent, sha).split(b'\0')
    paths, blobs, i = set(), set(), 0
    while i < len(fields):
        token = fields[i]
        i += 1
        if not token.startswith(b':'):
            continue
        meta = token[1:].split()
        if len(meta) != 5:
            raise ValueError('malformed raw git record')
        oldblob, newblob, status = meta[2], meta[3], meta[4]
        for blob in (oldblob, newblob):
            if set(blob) != {48}:
                blobs.add(blob.decode('ascii'))
        for _ in range(2 if status.startswith((b'R', b'C')) else 1):
            path = fields[i].decode('utf-8', errors='strict')
            paths.add(path)
            i += 1
    return paths, blobs


def matches(record, sha, paths, blobs):
    return (record.get('commit') == sha or record.get('path') in paths
        or record.get('blob') != EMPTY_BLOB and record.get('blob') in blobs)


def metrics(rows, owned_dirs):
    values = {p: {'N': 0, 'U': 0, 'composition': {'all_added': 0, 'any_renamed': 0,
        'any_heuristic_only_modified': 0}, 'merge_size_histogram': {}} for p in POPULATIONS}
    roles, extensions = Counter(), Counter()
    byrole = {}
    live = cap = merges = 0
    for row in rows:
        if row['association_exposed']:
            continue
        selected = row['selected'] or []
        merge_sizes = Counter()
        capped = row['capped'] or []
        merges += bool(selected)
        live += len(selected) + len(capped)
        cap += len(capped)
        for c in selected:
            members = c['members']
            test = any(role(m['path']) == 'test' for m in members)
            populations = ['all']
            if not test:
                populations.append('non_test')
            if all(m['kind'] == 'added' for m in members):
                populations.append('added_only')
            unowned = c['directory'] not in owned_dirs
            for population in populations:
                values[population]['N'] += 1
                values[population]['U'] += unowned
                merge_sizes[population] += 1
                composition = values[population]['composition']
                composition['all_added'] += all(m['kind'] == 'added' for m in members)
                composition['any_renamed'] += any(m['kind'] == 'renamed' for m in members)
                composition['any_heuristic_only_modified'] += any(m['kind'] == 'modified' for m in members)
            cluster_role = 'test' if test else role(members[0]['path'])
            rv = byrole.setdefault(cluster_role, {'N': 0, 'U': 0})
            rv['N'] += 1
            rv['U'] += unowned
            for member in members:
                roles[role(member['path'])] += 1
                filename = member['path'].rsplit('/', 1)[-1]
                extensions[filename.rsplit('.', 1)[-1].lower() if '.' in filename else '<none>'] += 1
        for population in POPULATIONS:
            histogram = values[population]['merge_size_histogram']
            size = str(merge_sizes[population])
            histogram[size] = histogram.get(size, 0) + 1
    for v in list(values.values()) + list(byrole.values()):
        v['ratio'] = v['U']/v['N'] if v['N'] else None
        v['wilson95_independence'] = wilson(v['U'], v['N'])
    return {'populations': values, 'role_proxy': byrole, 'role_files': dict(roles), 'extension_files': dict(extensions),
        'residual_merges': merges, 'pre_cap': live, 'pr_cap': cap, 'pr_cap_share': cap/live if live else None}


def select_repositories(repositories):
    # Inputs here are count summaries, never outcomes; every branch is fixed.
    for r in repositories:
        if r.get('preparation_refusal'):
            refusal = r['preparation_refusal']
            if refusal['class'] == 'operational' and refusal['code'] not in (
                    'PreparationInfeasibleOperational', 'PreparationInterruptedOperational', 'PreparationBudgetExhausted'):
                raise ValueError('nonterminal operational preparation refusal')
            r['gates'] = {'preparation': False}
            r['eligible'] = False
            continue
        m = r['segment_metrics']
        length = r['segment_commits']
        r['gates'] = {'segment_length': type(length) is int and 1500 <= length <= 2000,
            'all_supply': m['populations']['all']['N'] >= 1800,
            'residual_merges': m['residual_merges'] >= 200,
            'cap_share': m['pr_cap_share'] is not None and m['pr_cap_share'] <= .10,
            'non_test_supply': m['populations']['non_test']['N'] >= 200,
            'added_supply': m['populations']['added_only']['N'] >= 400}
        r['eligible'] = all(r['gates'].values())
    mandatory = repositories[:2]
    eligible = [r for r in repositories[2:] if r['eligible']]
    rankings = {p: [r['repository'] for r in sorted(eligible,
        key=lambda r: (-r['segment_metrics']['populations'][p]['ratio'], repositories.index(r)))] for p in POPULATIONS}
    leaders = {rank[0] for rank in rankings.values() if rank}
    winner = next(iter(leaders)) if len(leaders) == 1 and all(rankings.values()) else None
    margins = {}
    if winner and all(r['eligible'] for r in mandatory):
        wr = next(r for r in eligible if r['repository'] == winner)
        margins = {p: {r['repository']: wr['segment_metrics']['populations'][p]['ratio'] - r['segment_metrics']['populations'][p]['ratio'] for r in mandatory} for p in POPULATIONS}
    margin_pass = bool(margins) and all(v >= .05-1e-12 for comparisons in margins.values() for v in comparisons.values())
    # Freeze the concordant winner first. Projection failures never promote a runner-up.
    chosen = mandatory + ([next(r for r in eligible if r['repository'] == winner)] if winner else [])
    projection_gates = {}
    for index, r in enumerate(chosen):
        floor = 1404 if index < 2 else 2106
        projection = r.get('segment_projection_inputs', {})
        projected = projection.get('projected_selected_counts', {})
        capacities = projection.get('projected_label_capacity_per_cell', {})
        projection_gates[r['repository']] = {
            'required_per_age': floor,
            'projected_supply': all(sum(projected.get(age, [])) >= floor for age in ('Cold', 'C0')),
            'label_capacity': all(capacities.get(age, 0) >= floor for age in ('Cold', 'C0'))}
    projection_pass = len(chosen) == 3 and all(
        gate['projected_supply'] and gate['label_capacity'] for gate in projection_gates.values())
    proxy_by_repository = {r['repository']: r.get('segment_projection_inputs', {}).get(
        'resource_proxy_selected_including_exposed', 0) for r in chosen}
    calls = 4 * sum(proxy_by_repository.values())
    reserved_sessions = 24960 * 6 + 180
    resource_gates = {'fallback_calls': calls <= 40000,
        'jev_total_cost': .001674 * calls + 20 <= 200,
        'jev_repository_cost': all(.001674 * 4 * count <= 60 for count in proxy_by_repository.values())}
    global_operational_stop = any(r.get('preparation_refusal', {}).get('code') in (
        'PreparationInterruptedOperational', 'PreparationBudgetExhausted') for r in repositories)
    passes = (not global_operational_stop and all(r['eligible'] for r in mandatory) and winner is not None and margin_pass
        and projection_pass and all(resource_gates.values()))
    return {'rankings': rankings, 'concordant_winner': winner, 'margins': margins, 'margin_pass': margin_pass,
        'projection_gates': projection_gates, 'projection_pass': projection_pass,
        'resource_proxy_including_exposed': proxy_by_repository, 'resource_gates': resource_gates,
        'projected_fallback_calls': calls, 'projected_jev_usd': .001674 * calls + 20,
        'reserved_sessions': reserved_sessions, 'pre_map_calibration_required': True,
        'selected_repository': winner if passes else None,
        'verdict': 'eligible' if passes else ('not_testable_operational' if global_operational_stop or any(
            r.get('preparation_refusal', {}).get('class') == 'operational' for r in mandatory)
            else 'not_testable_no_substitution')}


def inventory(inputs_path, artifact_path):
    inputs_path, artifact_path = Path(inputs_path).resolve(), Path(artifact_path).resolve()
    config = json.loads(inputs_path.read_text())
    validate_inputs(config)
    configure_git(config)
    verify_python(config)
    artifacts = json.loads(artifact_path.read_text())
    validate_artifacts(artifacts, [r['repository'] for r in config['repositories']])
    if artifacts['source_inputs_sha256'] != digest(inputs_path) or artifacts['status'] != 'SEALED':
        raise ValueError('wrong generated export binding')
    ownership = verify_artifact_bindings(artifacts, inputs_path, artifact_path.parent)
    additional = load_pinned(artifacts['repository_exposure'], artifact_path.parent)
    exposure = load_pinned(artifacts['exposure'], artifact_path.parent)
    if set(artifacts['exposure_studies']) != set(config['exposure_studies']):
        raise ValueError('artifact study enumeration differs from source pins')
    validate_exposure(exposure, config['exposure_studies'])
    unscoped = load_pinned(artifacts['unscoped_exposure'], artifact_path.parent)
    validate_unscoped(unscoped, config['exposure_studies'])
    frontend = artifacts['frontend']
    executable = Path(frontend['path'])
    executable = executable if executable.is_absolute() else artifact_path.parent / executable
    if digest(executable) != frontend['sha256']:
        raise ValueError('frontend executable pin mismatch')
    scanner = config['export_preparation']['tools']['clawpatch']['sha256']
    result = {'schema': 'x11a-inventory-v5', 'inputs_sha256': digest(inputs_path), 'artifacts_sha256': digest(artifact_path), 'repositories': []}
    for item in config['repositories']:
        seal = ownership[item['repository']]
        if seal['status'] == 'refused':
            result['repositories'].append({'repository': item['repository'],
                'preparation_refusal': seal['refusal'],
                'repository_only_exposure_by_study': repository_exposure_by_study(
                    additional, item['repository'], config['exposure_studies'])})
            continue
        repo = Path(item['path'])
        # Recheck structural Git prerequisites without scanning candidate content.
        common = Path(git(repo, 'rev-parse', '--path-format=absolute', '--git-common-dir').decode().strip())
        if ((common / 'shallow').exists() or (common / 'info/grafts').exists()
                or git(repo, 'for-each-ref', '--format=%(refname)', 'refs/replace').strip()
                or list((common / 'objects/pack').glob('*.promisor'))):
            raise ValueError('operational: shallow/graft/replacement/promisor repository')
        partial = subprocess.run([GIT_BINARY, '-C', str(repo), 'config', '--get-regexp',
            r'^(extensions\.partialClone|remote\..*\.promisor)$'], env=git_env(),
            stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        if partial.returncode != 1:
            raise ValueError('operational: partial/promisor repository')
        for field in ('tip', 'Cold', 'C0'):
            if not SHA.fullmatch(item[field]) or git(repo, 'rev-parse', item[field]).decode().strip() != item[field]:
                raise ValueError('repository SHA mismatch')
        chain = git(repo, 'rev-list', '--first-parent', '--reverse', item['C0'] + '..' + item['tip']).decode().splitlines()
        ancestry = git(repo, 'rev-list', '--first-parent', item['tip']).decode().splitlines()
        if item['C0'] not in ancestry or chain != list(reversed(ancestry[:ancestry.index(item['C0'])])):
            raise ValueError('non-contiguous first-parent envelope')
        heur = load_pinned(seal['output'], inputs_path.parent)
        validate_heuristic(heur, item['repository'], scanner, config['export_preparation']['clawpatch_tree']['sha256'])
        segment_chain = chain[:2000]
        expected_order = list(dict.fromkeys([item['Cold'], item['C0'], *segment_chain]))
        if heur['revision_order'] != expected_order:
            raise ValueError('ownership revision scope mismatch')
        request = {'repo': str(repo), 'base': item['C0'],
                   'head': segment_chain[-1] if segment_chain else item['C0']}
        # Disk-spooled NDJSON bounds both processes to one ownership record.
        with tempfile.TemporaryFile() as stream:
            stream.write((json.dumps(request) + '\n').encode())
            for revision in [item['C0'], *segment_chain]:
                record = ownership_record(heur, revision)
                stream.write((json.dumps({'revision': revision, 'paths': record['paths']}) + '\n').encode())
            stream.seek(0)
            response = subprocess.check_output([str(executable)], stdin=stream,
                env=dict(git_env(), PATH=str(Path(GIT_BINARY).parent)))
        payload = json.loads(response)
        exact_keys(payload, {'merges', 'owned_directories'})
        merges = payload['merges']
        if (not isinstance(payload['owned_directories'], list)
                or not all(isinstance(p, str) for p in payload['owned_directories'])
                or payload['owned_directories'] != sorted(set(payload['owned_directories']))):
            raise ValueError('invalid frontend owned directory set')
        if [r['commit'] for r in merges] != segment_chain:
            raise ValueError('frontend sequence mismatch')
        records = [r for r in exposure['records'] if r['repository'] == item['repository']]
        excluded, other_matches, resolution = annotate_exposure(merges, records, unscoped['records'],
            lambda row: images(repo, row['parent'], row['commit']))
        for study in config['exposure_studies']:
            resolution.setdefault(study, {'resolved': 0, 'unresolved': 0, 'by_kind': {}})
        violations, rename_skipped = [], []
        previous_time = int(git(repo, 'show', '-s', '--format=%ct', item['C0']))
        for row in merges:
            if check_rename_detection(repo, row['parent'], row['commit']):
                rename_skipped.append(row['commit'])
            timestamp = int(git(repo, 'show', '-s', '--format=%ct', row['commit']))
            if timestamp < previous_time:
                violations.append(row['commit'])
            previous_time = timestamp
        segment = merges[:2000]
        owned_dirs = set(payload['owned_directories'])
        result['repositories'].append({'repository': item['repository'], 'tip': item['tip'], 'Cold': item['Cold'], 'C0': item['C0'],
            'envelope_commits': len(chain), 'segment_commits': len(segment),
            'segment_start': segment[0]['commit'] if segment else None, 'segment_end': segment[-1]['commit'] if segment else None,
            'segment_metrics': metrics(segment, owned_dirs), 'envelope_metrics': {'status': 'unscanned'},
            'segment_projection_inputs': projection_vectors(segment, item['repository']),
            'envelope_projection_inputs': {'status': 'unscanned'},
            'timestamp_inversions': violations, 'association_excluded_score_only': excluded, 'other_task_matches': other_matches,
            'exposure_resolution_by_study': resolution,
            'repository_only_exposure_by_study': repository_exposure_by_study(
                additional, item['repository'], config['exposure_studies']),
            'rename_detection_skipped_count': len(rename_skipped), 'rename_detection_skipped_merges': rename_skipped,
            'merges': merges, 'projection_scope': 'one age-independent no-active-owner supply proxy'})
    result['selection'] = select_repositories(result['repositories'])
    winner = result['selection']['selected_repository']
    result['oc_vectors'] = None
    if winner is not None:
        selected = result['repositories'][:2] + [
            r for r in result['repositories'] if r['repository'] == winner]
        result['oc_vectors'] = {'segment_lengths': [r['segment_commits'] for r in selected], 'counts': [
            [r['segment_projection_inputs']['projected_label_selected_counts'][age]
             for age in ('Cold', 'C0')] for r in selected],
            'directory_cases': [[r['segment_projection_inputs']['projected_label_directory_cases'][age]
                for age in ('Cold', 'C0')] for r in selected]}
    return result


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--inputs', type=Path, required=True)
    p.add_argument('--artifacts', type=Path, required=True)
    p.add_argument('--authorization', type=Path, required=True)
    p.add_argument('--output', type=Path, required=True)
    args = p.parse_args()
    from stages import verify_authorization
    verify_authorization(args.authorization, args.inputs, scope='inventory', artifacts=args.artifacts)
    data = inventory(args.inputs, args.artifacts)
    with args.output.open('x') as f:
        json.dump(data, f, sort_keys=True, indent=2)
        f.write('\n')


if __name__ == '__main__':
    main()

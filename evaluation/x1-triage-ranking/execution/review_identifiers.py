"""Project frozen identifier-only v4 outputs for X1 test-only council review."""
import argparse
import json
import os
from pathlib import Path
import sys

sys.path.insert(0, str(Path(__file__).resolve().parent))
from preparation import Journal, load_config, operation_lock, paths
from verify_inputs import exclusive_json, pin
from identifiers.projection import TESTS, project, reconcile

SCOUT = '04d5e632ae34d8056c3df5116b9ba87cfbf99282'
INERTIA = '85e7329bef230358740c2b7ae90b6364a15b675c'
SIGNED = {'frozen-exposure-inputs.json': 'ec92fba6d3f91c671301244539a8caefc5025c0d6b03efca77c5bdfc4bd96425',
          'exposure-sources.json': '67dc5a4bd7322317cb3785e1ddf7ad1ccbe75a4c73a9cdeab5ccb4d6afbadf60'}
OUTPUTS = ('exposure-identifiers.json', 'additional-identifiers.json', 'unattributed-identifiers.json',
           'coverage.json', 'inertia-enumeration.json')
INPUTS = {'seal.json', 'extract_exposures.py', 'exposure-contexts-v4.json', *OUTPUTS}
COVER_TOP = {'schema', 'scout_commit', 'inertia_commit', 'studies', 'counts', 'files',
             'limitations', 'candidate_repository_data_read', 'label_or_outcome_values_emitted'}


def read_json(path):
    def unique(pairs):
        result = {}
        for key, value in pairs:
            if key in result:
                raise ValueError('duplicate JSON key')
            result[key] = value
        return result
    return json.loads(Path(path).read_bytes(), object_pairs_hook=unique)


def frozen_inputs(frozen_path):
    frozen = read_json(frozen_path)
    if frozen['schema'] != 'scout.x1.frozen-exposure-inputs.v1' or frozen['inertia_commit'] != INERTIA or frozen['upstream_scout_commit'] != SCOUT:
        raise ValueError('frozen snapshot identity mismatch')
    files = {}
    for expected in frozen['files']:
        name = Path(expected['path']).name
        if name in files:
            raise ValueError('duplicate frozen input name')
        actual = pin(expected['path'])
        if any(actual[key] != expected[key] for key in ('sha256', 'bytes')):
            raise ValueError('frozen input pin mismatch: '+name)
        files[name] = actual
    if set(files) != INPUTS:
        raise ValueError('exact eight frozen inputs required')
    seal = read_json(files['seal.json']['path'])
    if seal['schema'] != 'x11a-exposure-extraction-seal-v2' or seal['scout_commit'] != SCOUT or seal['inertia_commit'] != INERTIA:
        raise ValueError('seal snapshot identity mismatch')
    if seal != frozen['upstream_seal_metadata']:
        raise ValueError('frozen seal metadata differs from signed snapshot')
    for key, name in (('recipe', 'extract_exposures.py'), ('contexts', 'exposure-contexts-v4.json')):
        if seal[key] != {k: files[name][k] for k in ('path', 'sha256')}:
            raise ValueError('seal source pin mismatch')
    if seal['files'] != [{k: files[name][k] for k in ('path', 'sha256')} for name in OUTPUTS]:
        raise ValueError('seal output pins mismatch')
    contexts = read_json(files['exposure-contexts-v4.json']['path'])
    if contexts['schema'] != 'x11a-exposure-contexts-v1' or contexts['scout_commit'] != SCOUT:
        raise ValueError('context identity mismatch')
    return files


def index_ranges(indices):
    ranges = []
    for index in sorted(indices):
        if ranges and index == ranges[-1][1] + 1:
            ranges[-1][1] = index
        else:
            ranges.append([index, index])
    return ranges


def validate_emission(groups, unresolved, projected, conservation):
    """Bind conservation to the encoded artefact, not just intermediate rows."""
    def refuse():
        raise ValueError('emitted identifier conservation failure')
    for name in OUTPUTS[:3]:
        expected = sorted((row['record_index'], row['value'])
                          for row in projected if row['output'] == name)
        actual = []
        for group in groups:
            if group['output'] != name:
                continue
            if not (len(group['record_indices']) == len(group['values']) == len(group['source_ids'])):
                refuse()
            actual.extend(zip(group['record_indices'], group['values']))
        rejected = [issue['record_index'] for issue in unresolved
                    if issue.get('output') == name and 'record_index' in issue]
        if (sorted(actual) != expected or
                sorted([index for index, _ in actual] + rejected) !=
                list(range(conservation[name]['input']))):
            refuse()
        scopes = [row['record_index'] for row in projected if row['output'] == name and
                  row['repository'] is None and row['kind'] in {'path', 'pr'}]
        for reason, indices in (
                ('unattributed_scope', scopes),
                ('source_field_mechanism_provenance_not_emitted_by_frozen_v4',
                 [index for index, _ in expected])):
            entries = [issue for issue in unresolved if issue.get('output') == name
                       and issue['reason'] == reason]
            if len(entries) != (1 if indices else 0):
                refuse()
            if entries:
                entry = entries[0]
                expanded = []
                for start, end in entry['record_index_ranges']:
                    if start > end:
                        refuse()
                    expanded.extend(range(start, end + 1))
                if entry['count'] != len(indices) or expanded != sorted(indices):
                    refuse()


def produce(files, source):
    from collections import Counter
    coverage = read_json(files['coverage.json']['path'])
    if set(coverage) != COVER_TOP or coverage['schema'] != 'x11a-identifier-extraction-coverage-v2' or coverage['scout_commit'] != SCOUT or coverage['inertia_commit'] != INERTIA or coverage['candidate_repository_data_read'] is not False or coverage['label_or_outcome_values_emitted'] is not False:
        raise ValueError('unsupported frozen coverage schema/identity')
    enumeration = read_json(files['inertia-enumeration.json']['path'])
    if set(enumeration) != {'repository', 'commit', 'recursive_listing', 'enumeration_result', 'scope', 'data_files'} or enumeration['commit'] != INERTIA:
        raise ValueError('unsupported inertia enumeration schema/identity')
    groups, inertia = reconcile(source, coverage)
    listing = enumeration['recursive_listing']
    paths_seen = [row['path'] for row in inertia]
    if len(listing) != len(set(listing)) or len(paths_seen) != len(set(paths_seen)) or set(listing) != set(paths_seen):
        raise ValueError('inertia recursive listing mismatch or duplicate')
    if any(row not in inertia for row in enumeration['data_files']):
        raise ValueError('inertia enumeration absent from coverage')
    if len({row['path'] for row in enumeration['data_files']}) != len(enumeration['data_files']):
        raise ValueError('duplicate inertia data file')
    if enumeration['enumeration_result']['data_files'] != len(enumeration['data_files']):
        raise ValueError('inertia enumeration count mismatch')
    summary = read_json(files['seal.json']['path'])['summary']
    rows, unresolved, conservation = [], [], {}
    for name in OUTPUTS[:3]:
        document = read_json(files[name]['path'])
        projected, issues = project(document, name, files[name])
        record_issues = [issue for issue in issues if issue['reason'] != 'unattributed_scope']
        if any('record_index' not in issue for issue in record_issues):
            raise ValueError('unsupported identifier output schema')
        indices = [row['record_index'] for row in projected] + [issue['record_index'] for issue in record_issues]
        if sorted(indices) != list(range(len(document['records']))):
            raise ValueError('identifier record conservation failure: '+name)
        conservation[name] = dict(input=len(document['records']), projected=len(projected),
                                  record_issues=len(record_issues))
        rows.extend(projected)
        unresolved.extend(record_issues)
        scope = [issue['record_index'] for issue in issues if issue['reason'] == 'unattributed_scope']
        if scope:
            unresolved.append(dict(reason='unattributed_scope', output=name, count=len(scope),
                                   record_index_ranges=index_ranges(scope)))
        if projected:
            unresolved.append(dict(reason='source_field_mechanism_provenance_not_emitted_by_frozen_v4',
                output=name, count=len(projected),
                record_index_ranges=index_ranges(row['record_index'] for row in projected)))
    coverage_counts = dict(Counter(row['status'] for row in coverage['files']))
    unscoped = dict(Counter(record['kind'] for record in
        read_json(files['unattributed-identifiers.json']['path'])['records']))
    unscoped = {kind: unscoped.get(kind, 0) for kind in ('commit', 'blob', 'path', 'pr')}
    if (conservation[OUTPUTS[0]]['input'] != summary['compatible_records'] or
            conservation[OUTPUTS[2]]['input'] != summary['unattributed_records'] or
            unscoped != summary['unscoped_by_kind'] or
            coverage_counts != summary['coverage_status_counts'] or coverage_counts != coverage['counts'] or
            len(enumeration['data_files']) != summary['inertia_data_files']):
        raise ValueError('frozen seal summary count mismatch')
    coverage_by_source = {}
    for record in coverage['files']:
        if isinstance(record, dict) and 'path' in record and 'study' in record:
            coverage_by_source.setdefault((record['study'], record['path']), []).append(record)
    source_table, source_ids, grouped = [], {}, {}
    count_table, familiarity = Counter(), {repo: set() for repo in TESTS}
    for row in rows:
        source_id = None
        if row['source_file']:
            key = (row['study'], row['source_file'])
            if key not in source_ids:
                candidates = coverage_by_source.get(key, [])
                match = candidates[0] if len(candidates) == 1 else {}
                source_ids[key] = len(source_table)
                source_table.append(dict(source_id=source_ids[key], study=key[0],
                    source_relative_path=key[1], source_bytes_sha256=match.get('sha256'),
                    source_blob=match.get('blob'), resolution='unique' if len(candidates) == 1 else
                    'ambiguous' if candidates else 'missing'))
            source_id = source_ids[key]
        keys = ('output', 'study', 'task', 'repository', 'kind', 'matching_rule')
        key = tuple(row[k] for k in keys) + (tuple(row['applies_to_test_repositories']),)
        if key not in grouped:
            grouped[key] = dict(zip(keys, key[:6]), applies_to_test_repositories=list(key[6]),
                                record_indices=[], values=[], source_ids=[])
        group = grouped[key]
        group['record_indices'].append(row['record_index'])
        group['values'].append(row['value'])
        group['source_ids'].append(source_id)
        for repo in row['applies_to_test_repositories']:
            count_table[(repo, row['study'], row['kind'], row['matching_rule'])] += 1
            familiarity[repo].add((row['output'], row['study'], row['task']))
    validate_emission(list(grouped.values()), unresolved, rows, conservation)
    for index, row in enumerate(coverage['files']):
        if isinstance(row, dict) and row.get('status') != 'parsed_identifier_projection':
            unresolved.append(dict(reason='source_not_identifier_parsed', coverage_index=index,
                source_repository=row.get('source_repository'), source_relative_path=row.get('path'),
                coverage_status=row.get('status')))
    def metadata(row):
        return {('source_relative_path' if key == 'path' else
                 'source_bytes_sha256' if key == 'sha256' else key): value
                for key, value in row.items()}
    for group in groups.values():
        for item in group:
            for key in ('source', 'frozen'):
                if isinstance(item.get(key), dict):
                    item[key] = metadata(item[key])
    enumeration = dict(enumeration, data_files=[metadata(row) for row in enumeration['data_files']])
    counts = dict(conservation=conservation, unscoped_by_kind=unscoped,
        coverage_status_counts=coverage_counts, inertia_data_files=len(enumeration['data_files']),
        by_test_repository_study_kind_rule=[dict(repository=k[0], study=k[1], kind=k[2],
            matching_rule=k[3], count=v) for k, v in sorted(count_table.items())],
        unresolved_by_reason=dict(Counter({reason: sum(item.get('count', 1) for item in unresolved
            if item['reason'] == reason) for reason in {item['reason'] for item in unresolved}})),
        familiarity_sources={repo: [dict(output=k[0], study=k[1], task=k[2]) for k in sorted(values)]
                             for repo, values in familiarity.items()})
    return dict(identifier_groups=list(grouped.values()), unattributed_sources=source_table,
                unresolved=unresolved, source_groups=groups, inertia_enumeration=enumeration, counts=counts)


def payloads_for(result, common):
    groups = result['source_groups']
    payloads = {
        'manifest.json': dict(schema='scout.x1.identifier-manifest.v1',
            groups=result['identifier_groups'], unattributed_sources=result['unattributed_sources'],
            counts=result['counts'], limitations=[
                'Scoped source-file/field/mechanism provenance is absent from frozen v4; unresolved.',
                'No issue/evidence/span/rename/excerpt fields exist in the frozen record schema.']),
        'source-coverage.json': dict(schema='scout.x1.source-coverage.v1', scout=groups,
            inertia_enumeration=result['inertia_enumeration'], counts=result['counts']),
        'delta-review.json': dict(schema='scout.x1.delta-review.v1',
            counts={key: len(value) for key, value in groups.items()},
            delta={key: value for key, value in groups.items() if key != 'matched'},
            council_completeness_decision_required=True, new_raw_schema_handling_authorised=False),
        'unresolved.json': dict(schema='scout.x1.unresolved.v1', records=result['unresolved'],
            counts=result['counts']['unresolved_by_reason'], coverage_unknown_schema=groups['unknown_schema'],
            council_sufficiency_required=True)}
    return {name: dict(common, **payload) for name, payload in payloads.items()}


def run(args):
    if pin(sys.executable)['sha256'] != '570b41051d135bcfc011b41272224f7cc66b0d0e7f735761dcb8611029f51d6d':
        raise ValueError('run with the pinned CPython 3.11.11 executable')
    config = load_config(args.config)
    p = paths(config)
    for argument, expected in ((args.frozen, p['STUDY']/'preparation/frozen-exposure-inputs.json'),
                               (args.sources, p['STUDY']/'preparation/exposure-sources.json'),
                               (args.output_root, p['OUT']/'identifiers')):
        if Path(argument).absolute() != expected:
            raise ValueError('noncanonical identifier input/output path')
    with operation_lock(config):
        journal = Journal(config, args.config, args.clock, 'd',
                          source_pins=[pin(__file__), pin(Path(__file__).parent/'identifiers/projection.py')])
        written = []
        try:
            for path in (args.frozen, args.sources):
                if pin(path)['sha256'] != SIGNED[Path(path).name]:
                    raise ValueError('signed recipe pin mismatch')
            files = frozen_inputs(args.frozen)
            source = read_json(args.sources)
            if source['schema'] != 'scout.x1.exposure-source-manifest.v1' or len(source['tracked_prior_evaluation_files']) != 24254:
                raise ValueError('signed source universe mismatch')
            result = produce(files, source)
            root = Path(args.output_root)
            root.mkdir()  # Exclusive reservation; partial outputs require reviewed recovery.
            common = dict(status='passed', clock=pin(args.clock),
                          inputs=[pin(args.frozen), pin(args.sources), *files.values()],
                          test_repositories=list(TESTS), development_freshness_filter=False,
                          historical_metadata_key_mapping=['path -> source_relative_path',
                                                           'sha256 -> source_bytes_sha256'],
                          unknown_freshness='ineligible', candidate_exclusions_computed=False)
            payloads = payloads_for(result, common)
            for name, payload in payloads.items():
                exclusive_json(root/name, payload)
                written.append(pin(root/name))
            with (root/'familiarity.jsonl').open('x') as stream:
                stream.write(json.dumps(dict(common, schema='scout.x1.familiarity.v1',
                    stage='preparation', entry_kind='source_table',
                    unattributed_sources=result['unattributed_sources']), sort_keys=True)+'\n')
                for group in result['identifier_groups']:
                    stream.write(json.dumps(dict(schema='scout.x1.familiarity.v1',
                        stage='preparation', entry_kind='known_evidence_not_actual_case_exclusion',
                        unknown_freshness='ineligible', development_freshness_filter=False,
                        identifier_group=group), sort_keys=True)+'\n')
                stream.flush()
                os.fsync(stream.fileno())
            directory_fd = os.open(root, os.O_RDONLY)
            try:
                os.fsync(directory_fd)
            finally:
                os.close(directory_fd)
            written.append(pin(root/'familiarity.jsonl'))
            journal.finish('passed', written)
        except Exception as error:
            journal.finish('refused', written, reason=str(error), exit_code=1)
            raise


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for argument in ('config', 'clock', 'frozen', 'sources', 'output-root'):
        parser.add_argument('--'+argument, required=True)
    run(parser.parse_args())


if __name__ == '__main__':
    try:
        main()
    except (ValueError, OSError, KeyError, TypeError) as error:
        print('Identifier review refused: '+str(error), file=sys.stderr)
        sys.exit(1)

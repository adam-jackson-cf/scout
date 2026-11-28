"""Append-only CC2 identifier delta. Real execution requires scoped council review."""
import argparse
from collections import Counter
import gzip
import hashlib
import json
import os
from pathlib import Path
import re
import sys

EXEC = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(EXEC))
from preparation import Journal, load_config, operation_lock, paths, review_receipt
from verify_inputs import pin
from identifiers.cc2_readers import csv_records, json_records
from identifiers.cc2_schemas import ALIASES, BASE, PANEL_COLUMNS, PRODUCERS, SCHEMAS, TABLES, TARGETS
from identifiers.projection import TESTS

SOURCES_SHA = '67dc5a4bd7322317cb3785e1ddf7ad1ccbe75a4c73a9cdeab5ccb4d6afbadf60'
PYTHON_SHA = '570b41051d135bcfc011b41272224f7cc66b0d0e7f735761dcb8611029f51d6d'
REVIEW_SCOPE = 'identifier-delta-cc2-implementation'
BASELINE_JOURNAL = '906389d7212a401b8c7f781d37960d84.json'
BASELINE_NAMES = ('manifest.json', 'source-coverage.json', 'delta-review.json', 'unresolved.json', 'familiarity.jsonl')


def metadata(path):
    # This function reads metadata/control documents only, never CC2 raw tables.
    def unique(pairs):
        result = {}
        for key, value in pairs:
            if key in result:
                raise ValueError('duplicate control-document key')
            result[key] = value
        return result
    return json.loads(Path(path).read_bytes(), object_pairs_hook=unique)


def table_records(raw, schema, framing):
    if framing == 'csv':
        columns = PANEL_COLUMNS + (['y_incomplete', 'y_incomplete_explicit'] if schema == 'panel-final' else [])
        return csv_records(raw, columns, {'repo', 'number'})
    return json_records(raw, SCHEMAS[schema], framing == 'jsonl')


def projected_identifiers(row, schema, fixed_repo):
    alias = row.get('repo', fixed_repo)
    blinded = schema.endswith('-blind')
    if blinded:
        repository = None
    else:
        if alias not in ALIASES or (fixed_repo is not None and alias != fixed_repo):
            raise ValueError('missing, unknown or inconsistent repository scope')
        repository = TARGETS.get(alias)
        if repository is None:
            return [], False
    output = []
    def emit(kind, value, field, nullable=False):
        if value is None and nullable:
            return
        if kind == 'commit':
            if not isinstance(value, str) or not re.fullmatch('[0-9a-fA-F]{40}', value):
                raise ValueError('invalid commit identifier')
            value = value.lower()
            rule, applies = 'commit_or_parent_either_image_any_scope', list(TESTS)
        elif kind == 'pr':
            if isinstance(value, bool) or not isinstance(value, (str, int)) or not re.fullmatch('[1-9][0-9]*', str(value)):
                raise ValueError('invalid PR identifier')
            value = str(value)
            rule, applies = 'repository_scoped_pr', [repository]
        else:
            if (not isinstance(value, str) or not value or value.startswith('/') or
                    any(part in {'.', '..', '.codegraph'} for part in value.split('/')) or
                    any(char in value for char in ('\x00', '\r', '\n'))):
                raise ValueError('invalid repository-relative path')
            rule = ('repository_scoped_path_unknown_mechanism' if repository else
                    'unresolved_repository_scope_unknown_freshness_ineligible')
            applies = [repository] if repository else list(TESTS)
        output.append(dict(repository=repository, kind=kind, value=value, field_path=field,
                           matching_rule=rule, applies_to_test_repositories=applies))
    def required(obj, keys):
        if not isinstance(obj, dict) or not set(keys) <= set(obj):
            raise ValueError('missing identifier structure')
    if schema in {'data', 'raw', 'panel', 'panel-final', 'type-call', 'type-sample'}:
        required(row, ['number'])
        emit('pr', row['number'], 'number')
    if schema in {'candidate', 'pair-call', 'pair-sample'}:
        required(row, ['parent', 'follow'])
        for key in ('parent', 'follow'):
            emit('pr', row[key], key)
    if schema == 'raw':
        required(row, ['base_oid', 'head_oid', 'merge_commit', 'merge_parents', 'commits', 'files'])
        for key in ('base_oid', 'head_oid', 'merge_commit'):
            emit('commit', row[key], key, nullable=True)
        for index, value in enumerate(row['merge_parents']):
            emit('commit', value, 'merge_parents[%d]' % index)
        for index, commit in enumerate(row['commits']):
            required(commit, ['oid', 'parent'])
            for key in ('oid', 'parent'):
                emit('commit', commit[key], 'commits[%d].%s' % (index, key), nullable=key == 'parent')
        for index, file in enumerate(row['files']):
            required(file, ['path'])
            emit('path', file['path'], 'files[%d].path' % index)
    if schema == 'data':
        for key in ('S', 'trail'):
            if key in row:
                emit('commit', row[key], key, nullable=True)
        for index, file in enumerate(row.get('files', [])):
            required(file, ['p'])
            emit('path', file['p'], 'files[%d].p' % index)
        for index, partner in enumerate(row.get('partners', [])):
            required(partner, ['g', 'via', 'vias', 'via_ref'])
            for key in ('g', 'via', 'via_ref'):
                emit('path', partner[key], 'partners[%d].%s' % (index, key), nullable=key == 'via_ref')
            for subindex, value in enumerate(partner['vias']):
                emit('path', value, 'partners[%d].vias[%d]' % (index, subindex))
    if schema in {'type-sample', 'type-blind'}:
        required(row, ['files'])
        for index, value in enumerate(row['files']):
            emit('path', value, 'files[%d]' % index)
    if schema in {'pair-sample', 'pair-blind'}:
        required(row, ['earlier', 'later'])
        for key in ('earlier', 'later'):
            required(row[key], ['files'])
            for index, value in enumerate(row[key]['files']):
                emit('path', value, '%s.files[%d]' % (key, index))
    if schema in {'dep-sample', 'dep-blind'}:
        required(row, ['file_a', 'file_b'])
        for key in ('file_a', 'file_b'):
            emit('path', row[key], key)
    return output, True


def project_table(raw, schema, framing, fixed_repo, source_id):
    records, total, selected = [], 0, 0
    for index, row in table_records(raw, schema, framing):
        total += 1
        identifiers, in_scope = projected_identifiers(row, schema, fixed_repo)
        selected += int(in_scope)
        records.extend(dict(identifier, source_id=source_id, record_index=index,
                            source_schema=schema) for identifier in identifiers)
    return records, dict(total_records=total, selected_records=selected,
                          other_repository_records=total-selected, emitted_identifiers=len(records))


def verified_bytes(root, row):
    path = root/row['path']
    if '.codegraph' in path.parts or path.is_symlink() or not path.is_file():
        raise ValueError('missing or unsafe pinned source')
    raw = path.read_bytes()
    actual = dict(path=str(path.absolute()), sha256=hashlib.sha256(raw).hexdigest(), bytes=len(raw))
    if actual['sha256'] != row['sha256'] or actual['bytes'] != row['bytes']:
        raise ValueError('signed CC2 source pin mismatch')
    return raw, actual


def baseline(p):
    journal = p['OUT']/'journal'/BASELINE_JOURNAL
    record = metadata(journal)
    if record['action'] != 'd' or record['status'] != 'passed' or record['exit_code'] != 0:
        raise ValueError('incorrect baseline action-d journal')
    outputs = [pin(p['OUT']/'identifiers'/name) for name in BASELINE_NAMES]
    if record['output_pins'] != outputs:
        raise ValueError('baseline identifier outputs changed')
    return [pin(journal), *outputs]


def review_gate(path, required):
    receipt = review_receipt(path, REVIEW_SCOPE)
    reviewed = {(row['path'], row['sha256']) for row in receipt['reviewed']}
    if not {(row['path'], row['sha256']) for row in required} <= reviewed:
        raise ValueError('delta review does not bind complete implementation/request/inputs')
    return receipt


def create_json(path, payload, created):
    with path.open('x', encoding='utf-8') as stream:
        created.append(path)
        json.dump(payload, stream, sort_keys=True, indent=2)
        stream.write('\n')
        stream.flush()
        os.fsync(stream.fileno())


def publish(root, common, records, sources, coverage, created=None):
    created = [] if created is None else created
    root.mkdir()  # Reservation is exclusive and retained after any failure.
    manifest = dict(common, schema='scout.x1.identifier-delta-cc2.v1', records=records,
        sources=sources, merge_semantics='union_with_frozen_v4_and_existing_X1_identifiers_never_replace',
        unknown_freshness='ineligible', candidate_exclusions_computed=False,
        counts_by_kind=dict(Counter(record['kind'] for record in records)))
    for name, payload in (
            ('manifest.json', manifest),
            ('coverage.json', dict(common, schema='scout.x1.identifier-delta-cc2-coverage.v1',
                source_coverage=coverage, source_count=len(sources),
                emitted_identifiers=len(records)))):
        create_json(root/name, payload, created)
    path = root/'familiarity.jsonl'
    with path.open('x') as stream:
        created.append(path)
        stream.write(json.dumps(dict(common, schema='scout.x1.familiarity-delta-cc2.v1',
            entry_kind='source_table', sources=sources), sort_keys=True)+'\n')
        for record in records:
            stream.write(json.dumps(dict(schema='scout.x1.familiarity-delta-cc2.v1',
                entry_kind='known_evidence_not_actual_case_exclusion',
                merge_operation='append_union', identifier=record), sort_keys=True)+'\n')
        stream.flush()
        os.fsync(stream.fileno())
    fd = os.open(root, os.O_RDONLY)
    try:
        os.fsync(fd)
    finally:
        os.close(fd)
    return [pin(path) for path in created]


def run(args):
    config = load_config(args.config)
    p = paths(config)
    if pin(sys.executable)['sha256'] != PYTHON_SHA:
        raise ValueError('pinned CPython 3.11.11 required')
    expected = {'sources': p['STUDY']/'preparation/exposure-sources.json',
                'review': p['STUDY']/'preparation-reviews/identifiers-delta-cc2-implementation.json',
                'output_root': p['OUT']/'identifiers-delta-cc2'}
    for key, path in expected.items():
        if Path(getattr(args, key)).absolute() != path.absolute():
            raise ValueError('noncanonical CC2 delta path')
    source_pin = pin(args.sources)
    if source_pin['sha256'] != SOURCES_SHA:
        raise ValueError('signed source manifest mismatch')
    implementation = [pin(Path(__file__).parent/name) for name in
                      ('cc2_delta.py', 'cc2_readers.py', 'cc2_schemas.py')]
    implementation += [pin(EXEC/'tests/test_cc2_delta.py'),
        pin(p['STUDY']/'preparation-reviews/identifiers-delta-cc2-REQUEST.md')]
    implementation += [pin(EXEC/'identifiers/projection.py'),
                       pin(EXEC/'preparation.py'), pin(EXEC/'verify_inputs.py')]
    with operation_lock(config):
        journal = Journal(config, args.config, args.clock, 'd', source_pins=implementation+[source_pin])
        root = expected['output_root']
        created = []
        try:
            prior = baseline(p)
            control_pins = implementation + [source_pin, pin(args.config)] + prior
            review_gate(args.review, control_pins)
            source_manifest = metadata(args.sources)
            if source_manifest['schema'] != 'scout.x1.exposure-source-manifest.v1':
                raise ValueError('wrong source manifest schema')
            rows = source_manifest['tracked_prior_evaluation_files']
            index = {row['path']: row for row in rows}
            if len(index) != len(rows):
                raise ValueError('duplicate source metadata path')
            records, source_table, coverage = [], [], []
            producer_pins = []
            for name in PRODUCERS:
                _, actual = verified_bytes(p['ROOT'], index[BASE+name])
                producer_pins.append(actual)
            # Verify every table before decoding any of them. These exact bytes,
            # not a second read, supply the reviewed projection.
            tables = []
            for relative, schema, framing, fixed in TABLES:
                raw, actual = verified_bytes(p['ROOT'], index[BASE+relative])
                tables.append((relative, schema, framing, fixed, raw, actual))
            for relative, schema, framing, fixed, raw, actual in tables:
                source_id = len(source_table)
                source_table.append(dict(source_id=source_id, pin=actual, source_schema=schema,
                                         source_relative_path=BASE+relative))
                content = gzip.decompress(raw) if relative.endswith('.gz') else raw
                projected, counts = project_table(content, schema, framing, fixed, source_id)
                records.extend(projected)
                coverage.append(dict(source_id=source_id, source_schema=schema, **counts))
            common = dict(status='passed', clock=pin(args.clock),
                implementation_review=pin(args.review), sources_manifest=source_pin,
                producer_code=producer_pins, predecessors=prior,
                development_freshness_filter=False)
            outputs = publish(root, common, records, source_table, coverage, created)
            journal.finish('passed', outputs)
        except Exception as error:
            outputs = [pin(path) for path in created]
            journal.finish('refused', outputs, reason=str(error), exit_code=1)
            raise


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for key in ('config', 'clock', 'sources', 'review', 'output-root'):
        parser.add_argument('--'+key, required=True)
    run(parser.parse_args())


if __name__ == '__main__':
    try:
        main()
    except (ValueError, OSError, KeyError, TypeError, EOFError) as error:
        print('CC2 delta refused: '+str(error), file=sys.stderr)
        sys.exit(1)

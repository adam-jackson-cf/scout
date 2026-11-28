"""Frozen X11 v4 identifier projection only; never opens prior-study raw files."""
import re

TESTS = ('django/django', 'microsoft/TypeScript', 'prometheus/prometheus')
ALIASES = {'django': TESTS[0], 'typescript': TESTS[1], 'prometheus': TESTS[2],
           'fastapi': 'fastapi/fastapi', 'axum': 'tokio-rs/axum',
           'gin': 'gin-gonic/gin', 'cobra': 'spf13/cobra'}
KINDS = {'commit', 'blob', 'path', 'pr', 'repository_identity'}
EMPTY_BLOB = 'e69de29bb2d1d6434b8b29ae775ad8c2e48c5391'
COVER_KEYS = {'source_repository', 'source_commit', 'path', 'blob', 'study', 'status',
              'inherited_repository', 'context_mapping_prefix', 'sha256', 'bytes',
              'unique_identifiers_in_file', 'new_unique_identifiers',
              'new_unattributed_identifiers', 'error_class'}
COVER_STATUSES = {'parsed_identifier_projection', 'non_data_source_or_document',
                  'unclassified_non_data', 'unsupported_git_object', 'could_not_parse'}


def canonical(value):
    if not isinstance(value, str):
        return None
    if value.lower() in ALIASES:
        return ALIASES[value.lower()]
    for repo in TESTS:
        if value.lower() == repo.lower():
            return repo
    return value if re.fullmatch(r'[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+', value) else None


def project(document, name, output_pin):
    """Return rows and unresolved metadata, without recursive field extraction."""
    issues, result = [], []
    allowed_top = {'records', 'studies'} if name == 'exposure-identifiers.json' else {'records'}
    if not isinstance(document, dict) or set(document) - allowed_top or not isinstance(document.get('records'), list):
        return [], [dict(reason='unknown_output_schema', output=name)]
    for index, record in enumerate(document['records']):
        provenance = dict(output=name, record_index=index,
                          schema='frozen-x11-extraction-v4', output_record='records[%d]' % index)
        unattributed = name == 'unattributed-identifiers.json'
        permitted_kinds = ({'commit', 'blob', 'path'} if name == 'exposure-identifiers.json' else
                           {'pr', 'repository_identity'} if name == 'additional-identifiers.json' else
                           {'commit', 'blob', 'path', 'pr'})
        allowed = {'source', 'study', 'task', 'kind', 'identifier', 'matchability'} if unattributed else {'repository', 'study', 'task'} | KINDS
        if not isinstance(record, dict) or set(record) - allowed:
            issues.append(dict(provenance, reason='unknown_record_schema'))
            continue
        kind = record.get('kind') if unattributed else next(iter(set(record) & KINDS), None)
        required = allowed if unattributed else {'repository', 'study', 'task', kind}
        if set(record) != required or kind not in permitted_kinds or not all(isinstance(v, str) for v in record.values()):
            issues.append(dict(provenance, reason='invalid_record_structure'))
            continue
        value = record['identifier'] if unattributed else record[kind]
        repo = None if unattributed else canonical(record['repository'])
        if not unattributed and repo is None:
            issues.append(dict(provenance, reason='unknown_repository'))
            continue
        if kind in {'commit', 'blob'} and not re.fullmatch('[0-9a-f]{40}', value):
            issues.append(dict(provenance, reason='invalid_object_identifier'))
            continue
        if kind == 'path' and (not value or value.startswith('/') or any(p in {'.', '..', '.codegraph'} for p in value.split('/')) or '\n' in value or '\x00' in value):
            issues.append(dict(provenance, reason='invalid_path_identifier'))
            continue
        if kind == 'repository_identity' and canonical(value) != repo:
            issues.append(dict(provenance, reason='repository_identity_mismatch'))
            continue
        rule = {'commit': 'commit_or_parent_either_image_any_scope',
                'blob': 'blob_either_image_any_scope',
                'path': 'repository_scoped_path_unknown_mechanism',
                'pr': 'repository_scoped_pr',
                'repository_identity': 'repository_familiarity_not_case_clearance'}[kind]
        if repo is None and kind in {'path', 'pr'}:
            rule = 'unresolved_repository_scope_unknown_freshness_ineligible'
        if kind == 'blob' and value == EMPTY_BLOB:
            rule = 'empty_blob_not_a_match'
        row = dict(provenance, repository=repo, kind=kind, value=value, matching_rule=rule,
                   study=record['study'], task=record['task'],
                   applies_to_test_repositories=([] if kind == 'blob' and value == EMPTY_BLOB else
                       list(TESTS) if kind in {'commit', 'blob'} or repo is None else
                       [repo] if repo in TESTS else []),
                   source_file=record.get('source'), source_file_sha256=None,
                   source_file_blob=None, source_field=None, mechanism=None)
        if unattributed:
            expected = 'repository_independent_sha' if kind in {'commit', 'blob'} else 'unmatchable_without_repository'
            if record['matchability'] != expected:
                issues.append(dict(provenance, reason='unknown_matchability'))
                continue
        result.append(row)
        if repo is None and kind in {'path', 'pr'}:
            issues.append(dict(provenance, reason='unattributed_scope'))
    return result, issues


def reconcile(source, coverage):
    """Compare pinned metadata, not the filesystem or prior source contents."""
    groups = {key: [] for key in ('matched', 'changed', 'missing', 'new', 'unknown_schema')}
    scout, inertia, seen = {}, [], set()
    for index, row in enumerate(coverage['files']):
        if (not isinstance(row, dict) or set(row) - COVER_KEYS or
                not {'source_repository', 'source_commit', 'path', 'blob', 'study', 'status'} <= set(row) or
                row.get('status') not in COVER_STATUSES):
            groups['unknown_schema'].append(dict(coverage_index=index, reason='coverage_row_schema'))
            continue
        if row['source_repository'] == 'inertia-bench':
            inertia.append(row)
        elif row['source_repository'] == 'scout':
            if row['path'] in scout:
                raise ValueError('duplicate frozen Scout source path')
            scout[row['path']] = row
        else:
            groups['unknown_schema'].append(dict(coverage_index=index, reason='unknown_source_repository'))
    for index, row in enumerate(source['tracked_prior_evaluation_files']):
        if not isinstance(row, dict) or set(row) != {'path', 'git_mode', 'git_blob', 'bytes', 'sha256'}:
            groups['unknown_schema'].append(dict(source_index=index, reason='source_row_schema'))
            continue
        path = row['path']
        if path in seen:
            raise ValueError('duplicate signed source path')
        seen.add(path)
        prior = scout.get(path)
        item = dict(source=row, frozen=prior)
        if prior is None:
            groups['new'].append(item)
        elif row['git_blob'] != prior['blob'] or ('sha256' in prior and row['sha256'] != prior['sha256']) or ('bytes' in prior and row['bytes'] != prior['bytes']):
            groups['changed'].append(item)
        else:
            item['comparison'] = 'blob_and_available_sha256_bytes' if 'sha256' in prior else 'git_blob_only_no_source_bytes_parsed'
            groups['matched'].append(item)
    groups['missing'] = [dict(frozen=row) for path, row in sorted(scout.items()) if path not in seen]
    return groups, inertia

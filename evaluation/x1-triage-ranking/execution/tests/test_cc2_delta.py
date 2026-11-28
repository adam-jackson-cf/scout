"""CC2 fabricated-schema tests; no real CC2 payload is opened."""
from contextlib import ExitStack
import gzip
import hashlib
import json
from pathlib import Path
import sys
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import patch

EXEC = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(EXEC))
import preparation
from identifiers import cc2_delta as delta
from identifiers import cc2_readers as readers
from identifiers.cc2_schemas import BASE, PANEL_COLUMNS, PRODUCERS, SCHEMAS, TABLES
from verify_inputs import pin


def encoded(row):
    return json.dumps(row).encode()


def raw_row():
    return dict(number=1, base_oid='a'*40, head_oid='b'*40, merge_commit='c'*40,
        merge_parents=['d'*40, 'e'*40], commits=[dict(oid='f'*40, parent='a'*40)],
        files=[dict(path='src/a.py')])


def data_row():
    return dict(repo='django', number=1, S='a'*40, trail='b'*40,
        files=[dict(p='src/a.py', hi='FORBIDDEN_SENTINEL')],
        partners=[dict(g='partner.py', via='src/a.py', vias=['src/a.py', 'src/b.py'], via_ref=None,
                       conf='FORBIDDEN_SENTINEL', explained='FORBIDDEN_SENTINEL')])


def sample(schema):
    if schema == 'raw':
        return raw_row()
    if schema == 'data':
        return data_row()
    if schema in {'candidate', 'pair-call'}:
        return dict(repo='django', parent=1, follow=2)
    if schema == 'type-call':
        return dict(repo='django', number=1)
    if schema == 'type-sample':
        return dict(repo='django', number=1, files=['a.py'])
    if schema == 'type-blind':
        return dict(files=['a.py'])
    if schema == 'pair-sample':
        return dict(repo='django', parent=1, follow=2, earlier=dict(files=['a.py']), later=dict(files=['b.py']))
    if schema == 'pair-blind':
        return dict(earlier=dict(files=['a.py']), later=dict(files=['b.py']))
    if schema == 'dep-sample':
        return dict(repo='django', file_a='a.py', file_b='b.py')
    if schema == 'dep-blind':
        return dict(file_a='a.py', file_b='b.py')
    raise AssertionError(schema)


def csv_fixture(final=False):
    columns = PANEL_COLUMNS + (['y_incomplete', 'y_incomplete_explicit'] if final else [])
    return (','.join(columns)+'\n'+','.join('django' if key == 'repo' else '1' if key == 'number' else
            'FORBIDDEN_SENTINEL' for key in columns)+'\n').encode()


class ReaderTests(unittest.TestCase):
    def test_excluded_fields_never_decoded_or_emitted(self):
        row = data_row()
        row['labels'] = {'FORBIDDEN_SENTINEL': [{'S': 'FORBIDDEN_SENTINEL', 'score': 1.25}]}
        actual_loads = readers.json.loads
        decoded = []
        def guarded(token, *args, **kwargs):
            self.assertNotIn(b'FORBIDDEN_SENTINEL', token if isinstance(token, bytes) else token.encode())
            decoded.append(token)
            return actual_loads(token, *args, **kwargs)
        with patch.object(readers.json, 'loads', side_effect=guarded):
            records, counts = delta.project_table(encoded(row), 'data', 'jsonl', 'django', 0)
        self.assertTrue(decoded)
        self.assertEqual(counts['selected_records'], 1)
        self.assertNotIn('FORBIDDEN_SENTINEL', json.dumps(records))
        self.assertNotIn('score', json.dumps(records))
        self.assertEqual([r['value'] for r in records if r['kind'] == 'commit'], ['a'*40, 'b'*40])

    def test_unknown_forbidden_key_refuses_before_its_value(self):
        row = data_row()
        row['score'] = 'FORBIDDEN_SENTINEL'
        with self.assertRaisesRegex(ValueError, 'unknown or duplicate'):
            delta.project_table(encoded(row), 'data', 'jsonl', 'django', 0)

    def test_oid_and_parent_are_commits_never_blobs(self):
        records, _ = delta.project_table(encoded([raw_row()]), 'raw', 'json', 'prometheus', 9)
        commits = {r['field_path']: r for r in records if r['kind'] == 'commit'}
        self.assertEqual(set(commits), {'base_oid', 'head_oid', 'merge_commit', 'merge_parents[0]',
                                      'merge_parents[1]', 'commits[0].oid', 'commits[0].parent'})
        self.assertEqual(commits['commits[0].oid']['value'], 'f'*40)
        self.assertEqual(commits['commits[0].oid']['source_id'], 9)
        self.assertTrue(all(r['kind'] != 'blob' for r in records))
        self.assertTrue(all(r['applies_to_test_repositories'] == list(delta.TESTS) for r in commits.values()))

    def test_snapshot_trail_measured_and_all_partner_paths(self):
        records, _ = delta.project_table(encoded(data_row()), 'data', 'jsonl', 'django', 0)
        fields = {r['field_path'] for r in records}
        self.assertEqual(fields, {'number', 'S', 'trail', 'files[0].p', 'partners[0].g',
            'partners[0].via', 'partners[0].vias[0]', 'partners[0].vias[1]'})
        self.assertTrue(all(r['applies_to_test_repositories'] == ['django/django'] for r in records if r['kind'] == 'path'))

    def test_null_parent_and_snapshot_do_not_invent_identifiers(self):
        row = raw_row()
        row['commits'][0]['parent'] = None
        row['base_oid'] = None
        records, _ = delta.project_table(encoded([row]), 'raw', 'json', 'django', 0)
        self.assertNotIn('commits[0].parent', {r['field_path'] for r in records})
        self.assertNotIn('base_oid', {r['field_path'] for r in records})
        rows, _ = delta.project_table(encoded(dict(repo='django', number=1, S=None, trail=None)), 'data', 'jsonl', 'django', 0)
        self.assertEqual([r['kind'] for r in rows], ['pr'])

    def test_all_schema_families(self):
        for schema in SCHEMAS:
            with self.subTest(schema=schema):
                fixed = 'django' if schema in {'raw', 'data'} else None
                framing = 'json' if schema == 'raw' else 'jsonl'
                row = sample(schema)
                raw = encoded([row] if framing == 'json' else row)
                records, counts = delta.project_table(raw, schema, framing, fixed, 0)
                self.assertTrue(records)
                self.assertEqual(counts['total_records'], 1)
                if schema.endswith('-blind'):
                    self.assertTrue(all(r['repository'] is None for r in records))
                    self.assertTrue(all(r['matching_rule'].startswith('unresolved_') for r in records))

    def test_parent_follow_prs_not_commit_sha(self):
        for schema in ('candidate', 'pair-call', 'pair-sample'):
            with self.subTest(schema=schema):
                records, _ = delta.project_table(encoded(sample(schema)), schema, 'jsonl', None, 0)
                self.assertEqual([(r['field_path'], r['kind']) for r in records[:2]], [('parent', 'pr'), ('follow', 'pr')])

    def test_other_repository_rows_not_used(self):
        row = dict(data_row(), repo='fastapi')
        records, counts = delta.project_table(encoded(row), 'data', 'jsonl', None, 0)
        self.assertFalse(records)
        self.assertEqual(counts['other_repository_records'], 1)
        with self.assertRaisesRegex(ValueError, 'inconsistent'):
            delta.project_table(encoded(row), 'data', 'jsonl', 'django', 0)

    def test_missing_scope_unknown_scope_bad_pr_sha_path_refuse(self):
        for row in (dict(number=1), dict(repo='unknown', number=1), dict(repo='django', number=True),
                    dict(repo='django', number=1, S='bad'), dict(repo='django', number=1, files=[dict(p='../escape')]),
                    dict(repo='django', number=1, partners=[dict(g='x')])):
            with self.subTest(row=row):
                with self.assertRaises(ValueError):
                    delta.project_table(encoded(row), 'data', 'jsonl', None, 0)

    def test_unknown_nested_structures_and_duplicate_keys_refuse(self):
        for raw in (b'{"repo":"django","number":1,"files":{"p":"a.py"}}',
                    b'{"repo":"django","number":1,"files":[{"unexpected":"a.py"}]}',
                    b'{"repo":"django","number":1,"number":2}',
                    b'{"repo":"django","number":1,"S":{"oid":"a"}}'):
            with self.subTest(raw=raw):
                with self.assertRaises(ValueError):
                    delta.project_table(raw, 'data', 'jsonl', None, 0)

    def test_json_bad_framing_refuses(self):
        for raw in (b'{"repo":"django","number":1} trailing', b'{"repo":"django","number":1,}',
                    b'{"repo":"django","number":1,"title":"bad\\q"}', b'\n', b'{"repo":'):
            with self.subTest(raw=raw):
                with self.assertRaises(ValueError):
                    delta.project_table(raw, 'data', 'jsonl', None, 0)

    def test_csv_never_decodes_excluded_cells(self):
        actual_decode = readers.csv_decode
        def guarded(raw, span):
            self.assertNotIn(b'FORBIDDEN_SENTINEL', raw[span[0]:span[1]])
            return actual_decode(raw, span)
        for final in (False, True):
            with self.subTest(final=final), patch.object(readers, 'csv_decode', side_effect=guarded):
                records, counts = delta.project_table(csv_fixture(final), 'panel-final' if final else 'panel', 'csv', None, 0)
                self.assertEqual([r['value'] for r in records], ['1'])
                self.assertEqual(counts['total_records'], 1)

    def test_csv_quoted_newline_skipped_without_decode(self):
        raw = csv_fixture().replace(b'FORBIDDEN_SENTINEL', b'"FORBIDDEN,\nSENTINEL"', 1)
        records, _ = delta.project_table(raw, 'panel', 'csv', None, 0)
        self.assertEqual(len(records), 1)

    def test_csv_unknown_header_duplicate_and_column_counts_refuse(self):
        for raw in (b'repo,number,score\ndjango,1,2\n', b'repo,number,number\ndjango,1,2\n',
                    csv_fixture()+b'django,2\n'):
            with self.subTest(raw=raw[:30]):
                with self.assertRaises(ValueError):
                    delta.project_table(raw, 'panel', 'csv', None, 0)


class DeltaPublicationTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)

    def test_source_pin_mismatch_and_symlink_refuse(self):
        path = self.root/'source'
        path.write_bytes(b'fabricated')
        row = dict(path='source', sha256=hashlib.sha256(b'fabricated').hexdigest(), bytes=10)
        raw, actual = delta.verified_bytes(self.root, row)
        self.assertEqual(raw, b'fabricated')
        self.assertEqual(actual, pin(path))
        with self.assertRaisesRegex(ValueError, 'pin mismatch'):
            delta.verified_bytes(self.root, dict(row, sha256='0'*64))
        link = self.root/'linked'
        link.symlink_to(path)
        with self.assertRaisesRegex(ValueError, 'unsafe'):
            delta.verified_bytes(self.root, dict(row, path='linked'))

    def test_publish_is_create_once_union_and_preserves_original(self):
        records, _ = delta.project_table(encoded(data_row()), 'data', 'jsonl', 'django', 0)
        output = self.root/'delta'
        outputs = delta.publish(output, dict(status='passed'), records, [], [])
        original = {p.name: p.read_bytes() for p in output.iterdir()}
        with self.assertRaises(FileExistsError):
            delta.publish(output, dict(status='passed'), records, [], [])
        self.assertEqual({p.name: p.read_bytes() for p in output.iterdir()}, original)
        self.assertEqual(len(outputs), 3)
        manifest = json.loads(original['manifest.json'])
        self.assertIn('union', manifest['merge_semantics'])
        self.assertFalse(manifest['candidate_exclusions_computed'])
        lines = [json.loads(line) for line in original['familiarity.jsonl'].splitlines()]
        self.assertEqual([line['identifier'] for line in lines[1:]], records)
        self.assertTrue(all(line['merge_operation'] == 'append_union' for line in lines[1:]))

    def test_partial_publication_retained(self):
        real_dump = delta.json.dump
        def failed(payload, stream, *args, **kwargs):
            if Path(stream.name).name == 'coverage.json':
                stream.write('{"partial":')
                raise OSError('planted publication failure')
            return real_dump(payload, stream, *args, **kwargs)
        output, created = self.root/'delta', []
        with patch.object(delta.json, 'dump', side_effect=failed):
            with self.assertRaisesRegex(OSError, 'planted'):
                delta.publish(output, dict(status='passed'), [], [], [], created)
        self.assertEqual(created, [output/'manifest.json', output/'coverage.json'])
        self.assertTrue((output/'manifest.json').is_file())
        self.assertEqual((output/'coverage.json').read_bytes(), b'{"partial":')

    def test_review_gate_requires_each_hash(self):
        required = [dict(path='/fabricated/code.py', sha256='a'*64)]
        for reviewed, accepted in ((required, True), ([], False), ([dict(required[0], sha256='b'*64)], False)):
            with self.subTest(reviewed=reviewed), patch.object(delta, 'review_receipt', return_value={'reviewed': reviewed}) as review:
                if accepted:
                    delta.review_gate('fabricated-review', required)
                else:
                    with self.assertRaisesRegex(ValueError, 'complete implementation'):
                        delta.review_gate('fabricated-review', required)
                review.assert_called_once_with('fabricated-review', delta.REVIEW_SCOPE)

    def test_changed_imported_dependency_refuses_review_gate(self):
        for name in ('projection.py', 'preparation.py', 'verify_inputs.py'):
            with self.subTest(dependency=name):
                path = self.root/name
                path.write_bytes(b'reviewed dependency')
                reviewed = [pin(path)]
                path.write_bytes(b'changed after review')
                with patch.object(delta, 'review_receipt', return_value={'reviewed': reviewed}):
                    with self.assertRaisesRegex(ValueError, 'complete implementation'):
                        delta.review_gate('fabricated-review', [pin(path)])

    def test_request_names_exact_baseline_journal_pin(self):
        request = (EXEC.parent/'preparation-reviews/identifiers-delta-cc2-REQUEST.md').read_text()
        self.assertIn('906389d7212a401b8c7f781d37960d84.json', request)
        self.assertIn('**10,301 bytes**', request)
        self.assertIn('6aa8802a23ac0de75eea9a78de5e6422e8cb16c4762f875dc7ebf0787d77dc6c', request)
        self.assertIn('c400e46e3f2ba85767165257b425a9719a223303b71cf78700d3111ae06f8738', request)
        self.assertNotIn('Main supplies the correct digest/start-record binding to the next output review', request)

    def test_request_requires_final_shared_code_freeze_order(self):
        request = (EXEC.parent/'preparation-reviews/identifiers-delta-cc2-REQUEST.md').read_text()
        command_section = request.split('## Main command, after zero-issue review only', 1)[1]
        self.assertIn('Core implementation review and the CC2 delta seal-integration review have both passed with zero issues', command_section)
        self.assertIn('Both sign-offs must name the same final `preparation.py` and `verify_inputs.py` digests', command_section)
        self.assertIn('then hold a confirming CC2 round at those exact final bytes naming all 16 bound digests', command_section)
        self.assertIn('After the producer run, none of the 16 bound files may change before the seal.', command_section)
        self.assertIn('exact standalone verdict line `**Zero issues. Signed off.**`', request)
        self.assertIn('exactly one standalone line `Signed-off scope: identifier-delta-cc2-implementation`', request)
        self.assertIn('the scope block must name every one of the 16 bound SHA-256 digests in full', request)
        for helper in ('preparation.py', 'verify_inputs.py'):
            self.assertIn(pin(EXEC/helper)['sha256'], request)
        self.assertIn('preparation-run/reviews/identifier-delta-cc2-output.json', command_section)

    def synthetic_fixture(self):
        study = self.root/'study'
        out = self.root/'out'
        (study/'preparation').mkdir(parents=True)
        (study/'preparation-reviews').mkdir()
        (out/'identifiers').mkdir(parents=True)
        (out/'journal').mkdir()
        clock = out/'clock.json'
        clock.write_text('{}')
        config = {'paths': {'ROOT': str(self.root), 'STUDY': str(study), 'OUT': str(out)}}
        config_path = self.root/'config.json'
        config_path.write_text(json.dumps(config))
        for name in delta.BASELINE_NAMES:
            (out/'identifiers'/name).write_text('{}')
        baseline_outputs = [pin(out/'identifiers'/name) for name in delta.BASELINE_NAMES]
        (out/'journal'/delta.BASELINE_JOURNAL).write_text(json.dumps(dict(action='d', status='passed', exit_code=0,
            clock=pin(clock), output_pins=baseline_outputs)))
        request = study/'preparation-reviews/identifiers-delta-cc2-REQUEST.md'
        request.write_text('fabricated request')
        review_path = study/'preparation-reviews/identifiers-delta-cc2-implementation.json'
        review_path.write_text('{}')
        rows = []
        for name in PRODUCERS:
            path = self.root/(BASE+name)
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(b'fabricated producer code')
            rows.append(dict(pin(path), path=BASE+name))
        for relative, schema, framing, fixed in TABLES:
            if framing == 'csv':
                raw = csv_fixture(schema == 'panel-final')
            else:
                row = sample(schema)
                if schema == 'data':
                    row['repo'] = fixed
                raw = encoded([row] if framing == 'json' else row)
            path = self.root/(BASE+relative)
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(gzip.compress(raw) if relative.endswith('.gz') else raw)
            rows.append(dict(pin(path), path=BASE+relative))
        sources = study/'preparation/exposure-sources.json'
        sources.write_text(json.dumps(dict(schema='scout.x1.exposure-source-manifest.v1', tracked_prior_evaluation_files=rows)))
        args = SimpleNamespace(config=str(config_path), clock=str(clock), sources=str(sources), review=str(review_path),
                               output_root=str(out/'identifiers-delta-cc2'))
        return args, config, sources, out, baseline_outputs

    def run_context(self, config, sources):
        actual_pin = delta.pin
        def runtime_pin(path, *rest):
            return dict(path=sys.executable, sha256=delta.PYTHON_SHA, bytes=1) if str(path) == sys.executable else actual_pin(path, *rest)
        stack = ExitStack()
        stack.enter_context(patch.object(delta, 'load_config', return_value=config))
        stack.enter_context(patch.object(delta, 'pin', side_effect=runtime_pin))
        stack.enter_context(patch.object(delta, 'SOURCES_SHA', pin(sources)['sha256']))
        self.review_mock = stack.enter_context(patch.object(delta, 'review_gate', return_value={}))
        stack.enter_context(patch.object(preparation, 'read_clock', return_value={'preparation_started_at': '2026-11-09T00:00:00Z'}))
        return stack

    def test_synthetic_journalled_run_and_refusal_keep_baseline(self):
        args, config, sources, out, baseline_outputs = self.synthetic_fixture()
        with self.run_context(config, sources):
            delta.run(args)
            with self.assertRaises(FileExistsError):
                delta.run(args)
        terminal = [json.loads(path.read_bytes()) for path in (out/'journal').glob('*.json')
                    if not path.name.endswith('.start.json') and path.name != delta.BASELINE_JOURNAL]
        self.assertEqual(sorted(row['status'] for row in terminal), ['passed', 'refused'])
        self.assertTrue(all(row['action'] == 'd' for row in terminal))
        self.assertEqual([pin(out/'identifiers'/name) for name in delta.BASELINE_NAMES], baseline_outputs)
        by_status = {row['status']: row for row in terminal}
        self.assertEqual(len(by_status['passed']['output_pins']), 3)
        self.assertEqual(by_status['refused']['output_pins'], [])
        helpers = [pin(EXEC/name) for name in
                   ('identifiers/projection.py', 'preparation.py', 'verify_inputs.py')]
        for record in terminal:
            for helper in helpers:
                self.assertIn(helper, record['source_pins'])
        journal_pin = pin(out/'journal'/delta.BASELINE_JOURNAL)
        self.assertEqual(self.review_mock.call_count, 2)
        for call in self.review_mock.call_args_list:
            required = call.args[1]
            self.assertIn(journal_pin, required)
            for helper in helpers:
                self.assertIn(helper, required)

    def test_journal_pins_partial_familiarity_from_current_run(self):
        args, config, sources, out, baseline_outputs = self.synthetic_fixture()
        real_dumps = delta.json.dumps
        def failed(payload, *args, **kwargs):
            if payload.get('entry_kind') == 'known_evidence_not_actual_case_exclusion':
                raise OSError('planted familiarity write failure')
            return real_dumps(payload, *args, **kwargs)
        with self.run_context(config, sources), patch.object(delta.json, 'dumps', side_effect=failed):
            with self.assertRaisesRegex(OSError, 'planted familiarity'):
                delta.run(args)
        terminal = [json.loads(path.read_bytes()) for path in (out/'journal').glob('*.json')
                    if not path.name.endswith('.start.json') and path.name != delta.BASELINE_JOURNAL]
        self.assertEqual(len(terminal), 1)
        self.assertEqual(terminal[0]['status'], 'refused')
        expected = [pin(out/'identifiers-delta-cc2'/name) for name in
                    ('manifest.json', 'coverage.json', 'familiarity.jsonl')]
        self.assertEqual(terminal[0]['output_pins'], expected)
        self.assertGreater(expected[-1]['bytes'], 0)
        lines = (out/'identifiers-delta-cc2/familiarity.jsonl').read_text().splitlines()
        self.assertEqual(len(lines), 1)
        self.assertEqual(json.loads(lines[0])['entry_kind'], 'source_table')
        self.assertEqual([pin(out/'identifiers'/name) for name in delta.BASELINE_NAMES], baseline_outputs)


if __name__ == '__main__':
    unittest.main()

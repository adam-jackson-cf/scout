"""Fabricated-only tests for review R1-01 through R1-05."""
from collections import Counter
from contextlib import ExitStack
from copy import deepcopy
import json
import inspect
from pathlib import Path
from types import SimpleNamespace
import sys
import tempfile
import unittest
from unittest.mock import patch

EXEC = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(EXEC))
import preparation
import review_identifiers as runner
from identifiers.projection import EMPTY_BLOB, TESTS, canonical, project, reconcile
from verify_inputs import pin


def scoped(**identifier):
    return dict(repository='django', study='old-study', task='other', **identifier)


def cover(path, **extra):
    return dict(source_repository='scout', source_commit=runner.SCOUT, path=path,
                blob='a'*40, study='old-study', status='parsed_identifier_projection', **extra)


def source(path):
    return dict(path=path, git_mode='100644', git_blob='a'*40, bytes=5, sha256='c'*64)


class Fixture:
    def __init__(self, root):
        self.root = root
        self.study = root/'study'
        self.prep = self.study/'preparation'
        self.prep.mkdir(parents=True)
        self.out = root/'out'
        self.out.mkdir()
        self.frozen = self.prep/'frozen-exposure-inputs.json'
        self.sources = self.prep/'exposure-sources.json'
        self.config_path = root/'config.json'
        self.clock = self.out/'clock.json'
        self.clock.write_text('{}')
        self.config = {'paths': {'OUT': str(self.out), 'STUDY': str(self.study)}}
        self.config_path.write_text(json.dumps(self.config))
        scout = cover('old.json', sha256='c'*64, bytes=5)
        inertia = dict(cover('i.json'), source_repository='inertia-bench', source_commit=runner.INERTIA)
        self.docs = {
            'extract_exposures.py': {},
            'exposure-contexts-v4.json': dict(schema='x11a-exposure-contexts-v1', scout_commit=runner.SCOUT),
            'exposure-identifiers.json': {'studies': [], 'records': [scoped(commit='a'*40)]},
            'additional-identifiers.json': {'records': [scoped(repository_identity='django')]},
            'unattributed-identifiers.json': {'records': [dict(source='old.json', study='old-study',
                task='other', kind='blob', identifier='b'*40, matchability='repository_independent_sha')]},
            'coverage.json': dict(schema='x11a-identifier-extraction-coverage-v2', scout_commit=runner.SCOUT,
                inertia_commit=runner.INERTIA, studies=[], counts={'parsed_identifier_projection': 2},
                files=[scout, inertia], limitations=[], candidate_repository_data_read=False,
                label_or_outcome_values_emitted=False),
            'inertia-enumeration.json': dict(repository='/fabricated/inertia', commit=runner.INERTIA,
                recursive_listing=['i.json'], enumeration_result={'data_files': 1}, scope='fabricated',
                data_files=[inertia])}
        self.summary = dict(compatible_records=1, unattributed_records=1,
            unscoped_by_kind={'commit': 0, 'blob': 1, 'path': 0, 'pr': 0},
            coverage_status_counts={'parsed_identifier_projection': 2}, inertia_data_files=1)
        self.source_doc = dict(schema='scout.x1.exposure-source-manifest.v1',
                               tracked_prior_evaluation_files=[source('old.json')])
        self.refresh()

    def refresh(self, seal_change=None, snapshot_change=None):
        for name, doc in self.docs.items():
            (self.root/name).write_text(json.dumps(doc))
        def short(name):
            return {k: pin(self.root/name)[k] for k in ('path', 'sha256')}
        seal = dict(schema='x11a-exposure-extraction-seal-v2', scout_commit=runner.SCOUT,
                    inertia_commit=runner.INERTIA, recipe=short('extract_exposures.py'),
                    contexts=short('exposure-contexts-v4.json'),
                    files=[short(name) for name in runner.OUTPUTS], summary=self.summary)
        if seal_change:
            seal_change(seal)
        (self.root/'seal.json').write_text(json.dumps(seal))
        self.files = {name: pin(self.root/name) for name in runner.INPUTS}
        frozen = dict(schema='scout.x1.frozen-exposure-inputs.v1', inertia_commit=runner.INERTIA,
            upstream_scout_commit=runner.SCOUT, files=list(self.files.values()), upstream_seal_metadata=seal)
        if snapshot_change:
            snapshot_change(frozen)
        self.frozen.write_text(json.dumps(frozen))
        self.sources.write_text(json.dumps(self.source_doc))

    def produce(self):
        return runner.produce(self.files, self.source_doc)

    def run_context(self):
        stack = ExitStack()
        stack.enter_context(patch.object(runner, 'SIGNED', {p.name: pin(p)['sha256'] for p in (self.frozen, self.sources)}))
        stack.enter_context(patch.object(runner, 'load_config', return_value=self.config))
        stack.enter_context(patch.object(preparation, 'read_clock', return_value={'preparation_started_at': '2026-11-09T00:00:00Z'}))
        real_pin = runner.pin
        def runtime_pin(path, *args):
            if str(path) == sys.executable:
                return dict(path=sys.executable, sha256='570b41051d135bcfc011b41272224f7cc66b0d0e7f735761dcb8611029f51d6d', bytes=1)
            return real_pin(path, *args)
        stack.enter_context(patch.object(runner, 'pin', side_effect=runtime_pin))
        return stack

    def args(self):
        return SimpleNamespace(config=str(self.config_path), clock=str(self.clock), frozen=str(self.frozen),
                               sources=str(self.sources), output_root=str(self.out/'identifiers'))

    def full_source_count(self):
        self.source_doc['tracked_prior_evaluation_files'] = [source('old.json')] + [source('new-%d' % i) for i in range(24253)]
        self.refresh()

    def multirecords(self):
        self.docs[runner.OUTPUTS[0]]['records'] = [
            scoped(commit='a'*40), scoped(commit='invalid'), scoped(commit='c'*40),
            dict(scoped(commit='a'*40), task='association'),
            dict(scoped(commit='a'*40), study='second-study'), scoped(path='django.py')]
        self.docs[runner.OUTPUTS[1]]['records'] = [
            scoped(repository_identity='django'), scoped(repository_identity='prometheus'),
            scoped(repository_identity='django'),
            dict(scoped(repository_identity='prometheus'), repository='prometheus', study='prometheus-study')]
        def unattributed(kind, value, source_name, matchability):
            return dict(source=source_name, study='old-study', task='other', kind=kind,
                        identifier=value, matchability=matchability)
        self.docs[runner.OUTPUTS[2]]['records'] = [
            unattributed('blob', 'b'*40, 'old.json', 'repository_independent_sha'),
            unattributed('blob', 'c'*40, 'old.json', 'invalid'),
            unattributed('path', 'a.py', 'old.json', 'unmatchable_without_repository'),
            unattributed('blob', 'd'*40, 'other.json', 'repository_independent_sha'),
            unattributed('path', 'b.py', 'other.json', 'unmatchable_without_repository')]
        self.summary.update(compatible_records=6, unattributed_records=5,
            unscoped_by_kind={'commit': 0, 'blob': 3, 'path': 2, 'pr': 0})
        self.refresh()


class ProjectionTests(unittest.TestCase):
    def project_one(self, record, name='exposure-identifiers.json'):
        return project({'records': [record]}, name, {})

    def test_any_scope_sha_and_repository_scoped_path_pr(self):
        for kind, name, value, applies in (
            ('commit', runner.OUTPUTS[0], 'a'*40, list(TESTS)),
            ('blob', runner.OUTPUTS[0], 'b'*40, list(TESTS)),
            ('path', runner.OUTPUTS[0], 'a.py', []),
            ('pr', runner.OUTPUTS[1], '1', []),
            ('blob', runner.OUTPUTS[0], EMPTY_BLOB, [])):
            with self.subTest(kind=kind, value=value):
                rows, _ = self.project_one(dict(scoped(**{kind: value}), repository='fastapi/fastapi'), name)
                self.assertEqual(rows[0]['applies_to_test_repositories'], applies)
                if kind in {'commit', 'blob'} and value != EMPTY_BLOB:
                    self.assertTrue(rows[0]['matching_rule'].endswith('_any_scope'))
                elif value == EMPTY_BLOB:
                    self.assertEqual(rows[0]['matching_rule'], 'empty_blob_not_a_match')

    def test_scoped_paths_prs_and_familiarity(self):
        for kind, name, value, rule in (
            ('path', runner.OUTPUTS[0], 'src/a.py', 'repository_scoped_path_unknown_mechanism'),
            ('pr', runner.OUTPUTS[1], '12', 'repository_scoped_pr'),
            ('repository_identity', runner.OUTPUTS[1], 'django', 'repository_familiarity_not_case_clearance')):
            with self.subTest(kind=kind):
                rows, issues = self.project_one(scoped(**{kind: value}), name)
                self.assertFalse(issues)
                self.assertEqual(rows[0]['applies_to_test_repositories'], [TESTS[0]])
                self.assertEqual(rows[0]['matching_rule'], rule)
                self.assertEqual(rows[0]['output'], name)

    def test_rejection_boundaries(self):
        cases = [(scoped(commit='bad'), runner.OUTPUTS[0], 'invalid_object_identifier'),
            (scoped(blob='G'*40), runner.OUTPUTS[0], 'invalid_object_identifier'),
            (scoped(path='../escape'), runner.OUTPUTS[0], 'invalid_path_identifier'),
            (scoped(repository_identity='prometheus'), runner.OUTPUTS[1], 'repository_identity_mismatch'),
            (dict(scoped(commit='a'*40), repository='not-a-repository'), runner.OUTPUTS[0], 'unknown_repository'),
            (dict(scoped(commit='a'*40), unexpected={'opaque': 'not-mined'}), runner.OUTPUTS[0], 'unknown_record_schema'),
            (scoped(commit='a'*40, blob='b'*40), runner.OUTPUTS[0], 'invalid_record_structure')]
        for record, name, reason in cases:
            with self.subTest(reason=reason, record=record):
                rows, issues = self.project_one(record, name)
                self.assertFalse(rows)
                self.assertEqual(issues[0]['reason'], reason)

    def test_unattributed_boundaries(self):
        for kind, value, match, reason in (
            ('blob', 'b'*40, 'repository_independent_sha', None),
            ('commit', 'c'*40, 'repository_independent_sha', None),
            ('path', 'src/a.py', 'unmatchable_without_repository', 'unattributed_scope'),
            ('blob', 'b'*40, 'unknown', 'unknown_matchability')):
            with self.subTest(kind=kind, match=match):
                record = dict(source='old.json', study='old', task='other', kind=kind, identifier=value, matchability=match)
                rows, issues = self.project_one(record, runner.OUTPUTS[2])
                self.assertEqual([i['reason'] for i in issues], [reason] if reason else [])
                if reason != 'unknown_matchability':
                    self.assertEqual(rows[0]['applies_to_test_repositories'], list(TESTS))
                else:
                    self.assertFalse(rows)

    def test_aliases_and_case(self):
        for value, expected in [('typescript', TESTS[1]), ('MICROSOFT/typescript', TESTS[1]),
                ('DJANGO/DJANGO', TESTS[0]), ('PROMETHEUS/PROMETHEUS', TESTS[2])]:
            with self.subTest(value=value):
                self.assertEqual(canonical(value), expected)

    def test_unknown_top_level_and_duplicate_json(self):
        rows, issues = project({'records': [], 'unknown': []}, runner.OUTPUTS[0], {})
        self.assertFalse(rows)
        self.assertEqual(issues[0]['reason'], 'unknown_output_schema')
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory)/'x.json'
            path.write_text('{"records": [], "records": []}')
            with self.assertRaisesRegex(ValueError, 'duplicate'):
                runner.read_json(path)


class PipelineTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.fx = Fixture(Path(self.temp.name))

    def test_group_conservation_pin_once_and_source_table(self):
        self.fx.multirecords()
        result = self.fx.produce()
        common = {'inputs': list(self.fx.files.values())}
        for payload in runner.payloads_for(result, common).values():
            found = preparation.verify_tree_pins(payload)
            self.assertEqual(Counter(p['path'] for p in found), Counter(p['path'] for p in self.fx.files.values()))
        self.assert_multirecord_encoding(result)
        self.assertEqual(result['unattributed_sources'], [
            dict(source_id=0, study='old-study', source_relative_path='old.json',
                 source_bytes_sha256='c'*64, source_blob='a'*40, resolution='unique'),
            dict(source_id=1, study='old-study', source_relative_path='other.json',
                 source_bytes_sha256=None, source_blob=None, resolution='missing')])
        self.assertEqual(result['counts']['conservation'], {
            runner.OUTPUTS[0]: dict(input=6, projected=5, record_issues=1),
            runner.OUTPUTS[1]: dict(input=4, projected=3, record_issues=1),
            runner.OUTPUTS[2]: dict(input=5, projected=4, record_issues=1)})

    def assert_multirecord_encoding(self, result):
        self.assertEqual(result['identifier_groups'], self.expected_groups())
        expected_familiarity = {
            TESTS[0]: [(runner.OUTPUTS[1], 'old-study', 'other'),
                (runner.OUTPUTS[0], 'old-study', 'association'),
                (runner.OUTPUTS[0], 'old-study', 'other'),
                (runner.OUTPUTS[0], 'second-study', 'other'),
                (runner.OUTPUTS[2], 'old-study', 'other')],
            TESTS[1]: [(runner.OUTPUTS[0], 'old-study', 'association'),
                (runner.OUTPUTS[0], 'old-study', 'other'),
                (runner.OUTPUTS[0], 'second-study', 'other'),
                (runner.OUTPUTS[2], 'old-study', 'other')],
            TESTS[2]: [(runner.OUTPUTS[1], 'prometheus-study', 'other'),
                (runner.OUTPUTS[0], 'old-study', 'association'),
                (runner.OUTPUTS[0], 'old-study', 'other'),
                (runner.OUTPUTS[0], 'second-study', 'other'),
                (runner.OUTPUTS[2], 'old-study', 'other')]}
        self.assertEqual(result['counts']['familiarity_sources'], {
            repo: [dict(output=name, study=study, task=task) for name, study, task in entries]
            for repo, entries in expected_familiarity.items()})
        def count(repo, study, kind, rule, total):
            return dict(repository=repo, study=study, kind=kind, matching_rule=rule, count=total)
        commit = 'commit_or_parent_either_image_any_scope'
        blob = 'blob_either_image_any_scope'
        unknown_path = 'unresolved_repository_scope_unknown_freshness_ineligible'
        familiarity = 'repository_familiarity_not_case_clearance'
        self.assertEqual(result['counts']['by_test_repository_study_kind_rule'], [
            count(TESTS[0], 'old-study', 'blob', blob, 2),
            count(TESTS[0], 'old-study', 'commit', commit, 3),
            count(TESTS[0], 'old-study', 'path', 'repository_scoped_path_unknown_mechanism', 1),
            count(TESTS[0], 'old-study', 'path', unknown_path, 2),
            count(TESTS[0], 'old-study', 'repository_identity', familiarity, 2),
            count(TESTS[0], 'second-study', 'commit', commit, 1),
            count(TESTS[1], 'old-study', 'blob', blob, 2),
            count(TESTS[1], 'old-study', 'commit', commit, 3),
            count(TESTS[1], 'old-study', 'path', unknown_path, 2),
            count(TESTS[1], 'second-study', 'commit', commit, 1),
            count(TESTS[2], 'old-study', 'blob', blob, 2),
            count(TESTS[2], 'old-study', 'commit', commit, 3),
            count(TESTS[2], 'old-study', 'path', unknown_path, 2),
            count(TESTS[2], 'prometheus-study', 'repository_identity', familiarity, 1),
            count(TESTS[2], 'second-study', 'commit', commit, 1)])
        expected = []
        gap = 'source_field_mechanism_provenance_not_emitted_by_frozen_v4'
        for name, reason, indices in (
                (runner.OUTPUTS[0], 'invalid_object_identifier', [[0, 0], [2, 5]]),
                (runner.OUTPUTS[1], 'repository_identity_mismatch', [[0, 0], [2, 3]]),
                (runner.OUTPUTS[2], 'unknown_matchability', [[0, 0], [2, 4]])):
            expected.append(dict(output=name, record_index=1, schema='frozen-x11-extraction-v4',
                output_record='records[1]', reason=reason))
            if name == runner.OUTPUTS[2]:
                expected.append(dict(reason='unattributed_scope', output=name, count=2,
                                     record_index_ranges=[[2, 2], [4, 4]]))
            gap_count = {runner.OUTPUTS[0]: 5, runner.OUTPUTS[1]: 3, runner.OUTPUTS[2]: 4}[name]
            expected.append(dict(reason=gap, output=name, count=gap_count,
                                 record_index_ranges=indices))
        self.assertEqual(result['unresolved'], expected)
        self.assertEqual(result['counts']['unresolved_by_reason'], {
            'invalid_object_identifier': 1, 'repository_identity_mismatch': 1,
            'unknown_matchability': 1, 'unattributed_scope': 2, gap: 12})

    def expected_groups(self):
        def group(output, kind, repository, rule, applies, indices, values, sources,
                  study='old-study', task='other'):
            return dict(output=output, study=study, task=task, repository=repository, kind=kind,
                matching_rule=rule, applies_to_test_repositories=applies, record_indices=indices,
                values=values, source_ids=sources)
        commit = 'commit_or_parent_either_image_any_scope'
        familiarity = 'repository_familiarity_not_case_clearance'
        return [
            group(runner.OUTPUTS[0], 'commit', TESTS[0], commit, list(TESTS), [0, 2], ['a'*40, 'c'*40], [None, None]),
            group(runner.OUTPUTS[0], 'commit', TESTS[0], commit, list(TESTS), [3], ['a'*40], [None], task='association'),
            group(runner.OUTPUTS[0], 'commit', TESTS[0], commit, list(TESTS), [4], ['a'*40], [None], study='second-study'),
            group(runner.OUTPUTS[0], 'path', TESTS[0], 'repository_scoped_path_unknown_mechanism', [TESTS[0]],
                  [5], ['django.py'], [None]),
            group(runner.OUTPUTS[1], 'repository_identity', TESTS[0], familiarity, [TESTS[0]],
                  [0, 2], ['django', 'django'], [None, None]),
            group(runner.OUTPUTS[1], 'repository_identity', TESTS[2], familiarity, [TESTS[2]],
                  [3], ['prometheus'], [None], study='prometheus-study'),
            group(runner.OUTPUTS[2], 'blob', None, 'blob_either_image_any_scope', list(TESTS),
                  [0, 3], ['b'*40, 'd'*40], [0, 1]),
            group(runner.OUTPUTS[2], 'path', None, 'unresolved_repository_scope_unknown_freshness_ineligible',
                  list(TESTS), [2, 4], ['a.py', 'b.py'], [0, 1])]

    def test_required_header_and_table_mutants_fail_oracle(self):
        self.fx.multirecords()
        original = inspect.getsource(runner.produce)
        mutants = [
            ('all-test headers', 'applies_to_test_repositories=list(key[6])',
             'applies_to_test_repositories=list(TESTS)'),
            ('first-test accumulation', "for repo in row['applies_to_test_repositories']:",
             "for repo in row['applies_to_test_repositories'][:1]:"),
            ('count overwrite', "count_table[(repo, row['study'], row['kind'], row['matching_rule'])] += 1",
             "count_table[(repo, row['study'], row['kind'], row['matching_rule'])] = 1"),
            ('missing familiarity', "familiarity[repo].add((row['output'], row['study'], row['task']))", 'pass'),
            ('task merge', 'tuple(row[k] for k in keys)',
             "tuple(row[k] if k != 'task' else 'other' for k in keys)"),
            ('study merge', 'tuple(row[k] for k in keys)',
             "tuple(row[k] if k != 'study' else 'old-study' for k in keys)")]
        for name, before, after in mutants:
            with self.subTest(mutant=name):
                self.assertEqual(original.count(before), 1)
                namespace = dict(vars(runner))
                exec(compile(original.replace(before, after), runner.__file__, 'exec'), namespace)
                result = namespace['produce'](self.fx.files, self.fx.source_doc)
                with self.assertRaises(AssertionError):
                    self.assert_multirecord_encoding(result)

    def test_planted_emitted_range_loss_refuses(self):
        self.fx.multirecords()
        actual_ranges = runner.index_ranges
        with patch.object(runner, 'index_ranges', side_effect=lambda indices: actual_ranges(indices)[:-1]):
            with self.assertRaisesRegex(ValueError, 'emitted identifier conservation failure'):
                self.fx.produce()

    def test_nonparsed_statuses_all_unresolved(self):
        statuses = ['non_data_source_or_document', 'unclassified_non_data', 'could_not_parse', 'unsupported_git_object']
        for status in statuses:
            row = cover(status)
            row['status'] = status
            self.fx.docs['coverage.json']['files'].append(row)
        counts = dict(Counter(r['status'] for r in self.fx.docs['coverage.json']['files']))
        self.fx.docs['coverage.json']['counts'] = counts
        self.fx.summary['coverage_status_counts'] = counts
        self.fx.refresh()
        result = self.fx.produce()
        self.assertEqual({r['coverage_status'] for r in result['unresolved'] if r['reason'] == 'source_not_identifier_parsed'}, set(statuses))

    def test_missing_and_ambiguous_source_lookup(self):
        for mode in ('missing', 'ambiguous'):
            with self.subTest(mode=mode):
                fx = self.fx
                fx.docs[runner.OUTPUTS[2]]['records'][0]['source'] = 'missing.json' if mode == 'missing' else 'old.json'
                if mode == 'ambiguous':
                    fx.docs['coverage.json']['files'].append(dict(cover('old.json'), source_repository='inertia-bench', source_commit=runner.INERTIA))
                    fx.docs['inertia-enumeration.json']['recursive_listing'].append('old.json')
                    fx.docs['coverage.json']['counts']['parsed_identifier_projection'] += 1
                    fx.summary['coverage_status_counts']['parsed_identifier_projection'] += 1
                fx.refresh()
                table = fx.produce()['unattributed_sources'][0]
                self.assertEqual(table['resolution'], mode)
                self.assertIsNone(table['source_bytes_sha256'])
                self.assertIsNone(table['source_blob'])

    def test_dropped_record_refuses(self):
        real_project = runner.project
        def dropped(*args):
            rows, issues = real_project(*args)
            return rows[1:], issues
        with patch.object(runner, 'project', side_effect=dropped):
            with self.assertRaisesRegex(ValueError, 'conservation'):
                self.fx.produce()

    def test_summary_count_refusals(self):
        for key in ('compatible_records', 'unattributed_records', 'unscoped_by_kind', 'coverage_status_counts', 'inertia_data_files'):
            with self.subTest(key=key):
                saved = deepcopy(self.fx.summary)
                self.fx.summary[key] = {} if isinstance(saved[key], dict) else 999
                self.fx.refresh()
                with self.assertRaisesRegex(ValueError, 'summary count'):
                    self.fx.produce()
                self.fx.summary = saved

    def test_listing_mismatch_and_duplicates_refuse(self):
        for listing in ([], ['i.json', 'i.json'], ['i.json', 'extra']):
            with self.subTest(listing=listing):
                self.fx.docs['inertia-enumeration.json']['recursive_listing'] = listing
                self.fx.refresh()
                with self.assertRaisesRegex(ValueError, 'listing'):
                    self.fx.produce()

    def test_coverage_and_enumeration_identity_refuse(self):
        for name, key in (('coverage.json', 'scout_commit'), ('coverage.json', 'schema'), ('inertia-enumeration.json', 'commit')):
            with self.subTest(name=name, key=key):
                saved = self.fx.docs[name][key]
                self.fx.docs[name][key] = 'wrong'
                self.fx.refresh()
                with self.assertRaisesRegex(ValueError, 'schema/identity'):
                    self.fx.produce()
                self.fx.docs[name][key] = saved

    def test_frozen_valid_eight_file_fixture(self):
        self.assertEqual(runner.frozen_inputs(self.fx.frozen), self.fx.files)

    def test_frozen_pin_duplicate_and_count_refusals(self):
        for mutation, reason in (
            (lambda f: f['files'][0].update(sha256='0'*64), 'pin mismatch'),
            (lambda f: f['files'].append(f['files'][0]), 'duplicate'),
            (lambda f: f['files'].pop(), 'eight'),
            (lambda f: f.update(upstream_scout_commit='wrong'), 'snapshot identity'),
            (lambda f: f['upstream_seal_metadata'].update(extra='wrong'), 'differs')):
            with self.subTest(reason=reason):
                self.fx.refresh(snapshot_change=mutation)
                with self.assertRaisesRegex(ValueError, reason):
                    runner.frozen_inputs(self.fx.frozen)
    def test_successful_publication_groups_and_pin_once(self):
        self.fx.multirecords()
        self.fx.full_source_count()
        with self.fx.run_context():
            runner.run(self.fx.args())
        root = self.fx.out/'identifiers'
        expected = Counter(p['path'] for p in self.fx.files.values())
        for name in ('manifest.json', 'source-coverage.json', 'delta-review.json', 'unresolved.json'):
            payload = runner.read_json(root/name)
            found = preparation.verify_tree_pins(payload)
            frozen = Counter(p['path'] for p in found if Path(p['path']).name in runner.INPUTS)
            self.assertEqual(frozen, expected)
        lines = [json.loads(line) for line in (root/'familiarity.jsonl').read_text().splitlines()]
        frozen = Counter(p['path'] for p in preparation.verify_tree_pins(lines)
                         if Path(p['path']).name in runner.INPUTS)
        self.assertEqual(frozen, expected)
        groups = [line['identifier_group'] for line in lines if 'identifier_group' in line]
        manifest = runner.read_json(root/'manifest.json')
        self.assertEqual(groups, manifest['groups'])
        self.assert_multirecord_encoding(dict(identifier_groups=manifest['groups'],
            unresolved=runner.read_json(root/'unresolved.json')['records'], counts=manifest['counts']))
        unresolved = runner.read_json(root/'unresolved.json')
        self.assert_multirecord_encoding(dict(identifier_groups=groups,
            unresolved=unresolved['records'], counts=manifest['counts']))


    def test_seal_identity_and_crosslinks_refuse(self):
        for mutation, reason in (
            (lambda s: s.update(scout_commit='wrong'), 'snapshot identity'),
            (lambda s: s['recipe'].update(sha256='0'*64), 'source pin'),
            (lambda s: s['contexts'].update(sha256='0'*64), 'source pin'),
            (lambda s: s['files'][0].update(sha256='0'*64), 'output pins')):
            with self.subTest(reason=reason):
                self.fx.refresh(seal_change=mutation)
                with self.assertRaisesRegex(ValueError, reason):
                    runner.frozen_inputs(self.fx.frozen)

    def test_context_identity_refuses(self):
        self.fx.docs['exposure-contexts-v4.json']['scout_commit'] = 'wrong'
        self.fx.refresh()
        with self.assertRaisesRegex(ValueError, 'context identity'):
            runner.frozen_inputs(self.fx.frozen)

    def test_runtime_noncanonical_signed_and_source_count_refuse(self):
        with patch.object(runner, 'pin', return_value={'sha256': 'wrong'}):
            with self.assertRaisesRegex(ValueError, 'CPython'):
                runner.run(self.fx.args())
        with self.fx.run_context():
            args = self.fx.args()
            args.output_root += '-other'
            with self.assertRaisesRegex(ValueError, 'noncanonical'):
                runner.run(args)
            with patch.object(runner, 'SIGNED', {self.fx.frozen.name: 'wrong'}):
                with self.assertRaisesRegex(ValueError, 'signed recipe'):
                    runner.run(self.fx.args())
            with self.assertRaisesRegex(ValueError, 'source universe'):
                runner.run(self.fx.args())
        terminal = [json.loads(p.read_text()) for p in (self.fx.out/'journal').glob('*.json') if not p.name.endswith('.start.json')]
        self.assertEqual([r['status'] for r in terminal], ['refused', 'refused'])

    def test_independent_sources_digest_and_canonical_input_refusals(self):
        self.fx.full_source_count()
        relocated = self.fx.root/'relocated'
        relocated.mkdir()
        with self.fx.run_context():
            for argument, original in (('frozen', self.fx.frozen), ('sources', self.fx.sources)):
                with self.subTest(argument=argument):
                    copy = relocated/original.name
                    copy.write_bytes(original.read_bytes())
                    self.assertEqual(pin(copy)['sha256'], pin(original)['sha256'])
                    args = self.fx.args()
                    setattr(args, argument, str(copy))
                    with self.assertRaisesRegex(ValueError, 'noncanonical'):
                        runner.run(args)
                    self.assertFalse((self.fx.out/'journal').exists())
            expected = dict(runner.SIGNED)
            self.assertEqual(expected[self.fx.frozen.name], pin(self.fx.frozen)['sha256'])
            expected[self.fx.sources.name] = '0'*64
            with patch.object(runner, 'SIGNED', expected):
                with self.assertRaisesRegex(ValueError, 'signed recipe'):
                    runner.run(self.fx.args())
        terminal = [json.loads(path.read_text()) for path in
                    (self.fx.out/'journal').glob('*.json') if not path.name.endswith('.start.json')]
        self.assertEqual([record['status'] for record in terminal], ['refused'])
        self.assertEqual(terminal[0]['failure_reason'], 'signed recipe pin mismatch')
        self.assertFalse((self.fx.out/'identifiers').exists())

    def test_existing_root_refuses_and_preserves_bytes(self):
        self.fx.full_source_count()
        root = self.fx.out/'identifiers'
        root.mkdir()
        original = root/'original'
        original.write_bytes(b'immutable')
        with self.fx.run_context():
            with self.assertRaises(FileExistsError):
                runner.run(self.fx.args())
        self.assertEqual(original.read_bytes(), b'immutable')
        self.assertEqual(list(root.iterdir()), [original])
        terminal = [json.loads(p.read_text()) for p in (self.fx.out/'journal').glob('*.json') if not p.name.endswith('.start.json')]
        self.assertEqual(terminal[0]['status'], 'refused')

    def test_partial_publication_retained(self):
        self.fx.full_source_count()
        real_write = runner.exclusive_json
        def failing(path, payload):
            if Path(path).name == 'delta-review.json':
                raise OSError('planted publishing failure')
            return real_write(path, payload)
        with self.fx.run_context(), patch.object(runner, 'exclusive_json', side_effect=failing):
            with self.assertRaisesRegex(OSError, 'planted'):
                runner.run(self.fx.args())
        root = self.fx.out/'identifiers'
        self.assertTrue((root/'manifest.json').is_file())
        self.assertTrue((root/'source-coverage.json').is_file())
        self.assertFalse((root/'delta-review.json').exists())
        terminal = [json.loads(p.read_text()) for p in (self.fx.out/'journal').glob('*.json') if not p.name.endswith('.start.json')]
        self.assertEqual(terminal[0]['status'], 'refused')
        self.assertEqual(len(terminal[0]['output_pins']), 2)

    def test_metadata_delta_and_duplicate(self):
        groups, inertia = reconcile({'tracked_prior_evaluation_files': [source('same'), source('changed'), source('new')]},
            {'files': [cover('same'), dict(cover('changed'), blob='d'*40), cover('gone'), dict(cover('unknown'), unknown=True)]})
        self.assertEqual({k: len(v) for k, v in groups.items()}, dict(matched=1, changed=1, missing=1, new=1, unknown_schema=1))
        self.assertFalse(inertia)
        with self.assertRaisesRegex(ValueError, 'duplicate'):
            reconcile({'tracked_prior_evaluation_files': []}, {'files': [cover('x'), cover('x')]})


if __name__ == '__main__':
    unittest.main()

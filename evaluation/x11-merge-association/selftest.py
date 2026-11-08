#!/usr/bin/env python3
"""Synthetic-only selection/schema/seal fixtures. Go frontend owns git parity."""
import copy
import json
from pathlib import Path
import tempfile
import subprocess
import unittest
from unittest.mock import patch

import inventory
import stages
from bootstrap_selftest_v6 import BootstrapTrustTests
from oc_selftest_v6 import OCContractTests


def summary(name, ratio=.1):
    return {'repository': name, 'segment_commits': 2000,
        'segment_metrics': {'populations': {p: {'N': 1800, 'U': round(ratio*1800), 'ratio': ratio} for p in inventory.POPULATIONS},
        'residual_merges': 200, 'pr_cap_share': .1},
        'segment_projection_inputs': {
            'projected_selected_counts': {'Cold': [2200], 'C0': [2200]},
            'projected_label_capacity_per_cell': {'Cold': 2200, 'C0': 2200},
            'resource_proxy_selected_including_exposed': 2800}}


class InventoryFixtures(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        # No ambient executable fallback: use the same reviewed git as inventory.
        config = json.loads((inventory.ROOT/'inventory-inputs.json').read_text())
        inventory.verify_python(config)
        inventory.configure_git(config)

    def candidates(self):
        return [summary('django'), summary('typescript'), summary('gin', .3), summary('cobra', .2), summary('prometheus', .4)]

    def test_selection_branches(self):
        good = self.candidates()
        r = inventory.select_repositories(good)
        self.assertEqual(r['selected_repository'], 'prometheus')
        self.assertTrue(r['margin_pass'])
        for length in (1500, 1666, 1999, 2000):
            data = self.candidates(); data[0]['segment_commits'] = length
            self.assertEqual(inventory.select_repositories(data)['selected_repository'], 'prometheus')
        for key, value in [('segment_commits', value) for value in (1499, 2001, None, True)]:
            data = self.candidates(); data[0][key] = value
            self.assertEqual(inventory.select_repositories(data)['verdict'], 'not_testable_no_substitution')
        for gate in ('all', 'non_test', 'added_only'):
            data = self.candidates(); data[0]['segment_metrics']['populations'][gate]['N'] = 0
            self.assertIsNone(inventory.select_repositories(data)['selected_repository'])
        for field, value in [('residual_merges', 199), ('pr_cap_share', .101)]:
            data = self.candidates(); data[0]['segment_metrics'][field] = value
            self.assertIsNone(inventory.select_repositories(data)['selected_repository'])
        data = self.candidates()
        for r in data[2:]: r['segment_commits'] = 0
        self.assertIsNone(inventory.select_repositories(data)['selected_repository'])
        data = self.candidates(); data[2]['segment_metrics']['populations']['non_test']['ratio'] = .9
        self.assertIsNone(inventory.select_repositories(data)['selected_repository'])
        data = [summary('django', .39), summary('typescript', .1), summary('gin', .4)]
        self.assertFalse(inventory.select_repositories(data)['margin_pass'])
        data = [summary('django'), summary('typescript'), summary('gin', .4), summary('cobra', .4)]
        self.assertEqual(inventory.select_repositories(data)['selected_repository'], 'gin')

    def test_roles_and_exact_proxy(self):
        for path in ['src/x_test.go', 'src/x.test.ts', 'src/x.spec.tsx', 'src/x.test.js',
                     'src/x.spec.jsx', 'app/test_one.py', 'app/one_test.py',
                     'tests/baselines/x.txt', 'django/test/a.py', 'a/testdata/x',
                     'a/baselines/x', 'a/tests/x', 'a/test/x']:
            self.assertEqual(inventory.role(path), 'test')
        for path in ['src/test_helpers.go', 'test_helpers.go', 'test_dir/source.py',
                     'src/testRunner/x.ts', 'src/harness/x.ts', 'util/testutil/x.go',
                     'src/test_x.ts', 'src/foo_test.txt']:
            self.assertNotEqual(inventory.role(path), 'test')
        self.assertEqual(inventory.role('docs/topic.txt'), 'documentation-non-markdown')
        rows = [{'association_exposed': False, 'capped': [], 'selected': [
            {'directory': 'src', 'members': [{'path': 'src/x.go', 'kind': 'added'}]},
            {'directory': 'tests', 'members': [{'path': 'tests/test_x.py', 'kind': 'modified'}]},
            {'directory': 'other', 'members': [{'path': 'other/x.go', 'kind': 'renamed'}]}]}]
        result = inventory.metrics(rows, {'src'})
        self.assertEqual(result['populations']['all']['U'], 2)
        self.assertEqual(result['populations']['all']['N'], 3)
        self.assertEqual(result['populations']['non_test']['U'], 1)
        self.assertEqual(result['populations']['added_only']['U'], 0)
        self.assertEqual(result['populations']['all']['composition'],
            {'all_added': 1, 'any_renamed': 1, 'any_heuristic_only_modified': 1})
        self.assertEqual(result['populations']['all']['merge_size_histogram'], {'3': 1})
        self.assertEqual(result['populations']['added_only']['merge_size_histogram'], {'1': 1})
        rows[0]['association_exposed'] = True
        self.assertEqual(inventory.metrics(rows, {'src'})['populations']['all']['N'], 0)

    def test_modified_pre_and_post_images(self):
        with tempfile.TemporaryDirectory(prefix='x11a-images-') as temporary:
            repo = Path(temporary)
            def git(*args):
                return inventory.git(repo, *args).decode().strip()
            git('init', '-q')
            git('config', 'user.name', 'Synthetic')
            git('config', 'user.email', 'synthetic@example.invalid')
            git('config', 'commit.gpgsign', 'false')
            (repo/'source.go').write_text('before\n')
            git('add', '.'); git('commit', '-q', '-m', 'before')
            parent = git('rev-parse', 'HEAD')
            old = git('rev-parse', 'HEAD:source.go')
            (repo/'source.go').write_text('after\n')
            git('add', '.'); git('commit', '-q', '-m', 'after')
            head = git('rev-parse', 'HEAD')
            new = git('rev-parse', 'HEAD:source.go')
            paths, blobs = inventory.images(repo, parent, head)
            self.assertEqual(paths, {'source.go'})
            self.assertEqual(blobs, {old, new})
            self.assertTrue(inventory.matches({'blob': old}, head, paths, blobs))
            self.assertTrue(inventory.matches({'blob': new}, head, paths, blobs))

    def test_global_schema_refusals(self):
        valid = {'studies': ['s'], 'records': [{'study': 's', 'repository': 'other/repo', 'task': 'other', 'blob': '1'*40}]}
        inventory.validate_exposure(valid, ['s'])
        for key, value in [('label', 'forbidden'), ('commit', 'not-sha'), ('study', 'unknown')]:
            data = copy.deepcopy(valid); data['records'][0][key] = value
            with self.assertRaises(ValueError): inventory.validate_exposure(data, ['s'])
        self.assertTrue(inventory.matches(valid['records'][0], '2'*40, set(), {'1'*40}))
        self.assertTrue(inventory.matches({'path': 'old.go'}, '2'*40, {'old.go'}, set()))
        heuristic = {'schema': 'x11a-ownership-index-v6', 'source': 'heuristic',
            'repository': 'r', 'scanner_sha256': 'digest', 'scanner_tree_sha256': 'tree',
            'revision_order': ['1'*40], 'revisions': {'1'*40: {'path': '/unit-only-record', 'sha256': '2'*64}}}
        inventory.validate_heuristic(heuristic, 'r', 'digest', 'tree')
        for mutation in ('agent', 'wrong-scanner', 'wrong-tree', 'bad-paths', 'extra'):
            h = copy.deepcopy(heuristic)
            if mutation == 'agent': h['source'] = 'agent'
            if mutation == 'wrong-scanner': h['scanner_sha256'] = 'other'
            if mutation == 'wrong-tree': h['scanner_tree_sha256'] = 'other'
            if mutation == 'bad-paths': h['revisions']['1'*40] = ['z', 'a']
            if mutation == 'extra': h['label'] = 'forbidden'
            with self.assertRaises(ValueError): inventory.validate_heuristic(h, 'r', 'digest', 'tree')
        with self.assertRaises(ValueError): inventory.reject_pending({'nested': ['PENDING']})

    def test_protocol_seal_refusals(self):
        with tempfile.TemporaryDirectory(prefix='x11a-seals-') as temporary:
            root = Path(temporary)
            payload = root/'payload'; payload.write_text('synthetic only')
            def entry(path): return {'path': str(path), 'sha256': stages.sha(path)}
            authority = root/'authority'; authority.write_text('synthetic original authority')
            seal = root/'seal.json'
            def write_seal(stage='inventory', digest=None):
                seal.write_text(json.dumps({'stage': stage, 'authorization_sha256': digest or stages.sha(authority),
                    'inputs': [entry(authority)], 'argv': ['synthetic-driver', 'inventory'],
                    'predecessors': [], 'artifacts': [entry(payload)]}))
            with self.assertRaises(FileNotFoundError): stages.verify_seal(seal, 'inventory', stages.sha(authority))
            write_seal(stage='wrong')
            with self.assertRaises(ValueError): stages.verify_seal(seal, 'inventory', stages.sha(authority))
            write_seal(digest='stale')
            with self.assertRaises(ValueError): stages.verify_seal(seal, 'inventory', stages.sha(authority))
            write_seal()
            stages.verify_seal(seal, 'inventory', stages.sha(authority))
            original = json.loads(seal.read_text())
            for field in ('inputs', 'argv', 'predecessors'):
                broken = copy.deepcopy(original); del broken[field]
                seal.write_text(json.dumps(broken))
                with self.subTest(field=field), self.assertRaises(ValueError):
                    stages.verify_seal(seal, 'inventory', stages.sha(authority))
            for field, value in (('inputs', []), ('argv', []), ('argv', [1]),
                    ('predecessors', [entry(authority)]), ('label', 'unexpected')):
                broken = copy.deepcopy(original); broken[field] = value
                seal.write_text(json.dumps(broken))
                with self.subTest(field=field, value=value), self.assertRaises(ValueError):
                    stages.verify_seal(seal, 'inventory', stages.sha(authority))
            write_seal()
            with self.assertRaisesRegex(ValueError, 'omits bound source inputs'):
                stages.verify_seal(seal, 'inventory', stages.sha(authority), payload)
            write_seal()
            payload.write_text('changed')
            with self.assertRaises(ValueError): stages.verify_seal(seal, 'inventory', stages.sha(authority))

    def test_unscoped_positive_and_unmatchable_identifiers(self):
        sha, blob = '1'*40, '2'*40
        def record(kind, identifier, study='x2d-merge-association-pilot', task='association'):
            return {'source': 'synthetic.json', 'study': study, 'kind': kind,
                'identifier': identifier, 'task': task,
                'matchability': 'repository_independent_sha' if kind in ('commit', 'blob') else 'unmatchable_without_repository'}
        records = [record('commit', sha), record('blob', blob, 'other', 'other'),
            record('blob', inventory.EMPTY_BLOB), record('path', 'same/path.go'), record('pr', '1')]
        inventory.validate_unscoped({'records': records}, ['x2d-merge-association-pilot', 'other'])
        rows = [{'commit': sha}, {'commit': '3'*40}]
        excluded, reported, counts = inventory.annotate_exposure(rows, [], records,
            lambda row: ({'same/path.go'}, {blob, inventory.EMPTY_BLOB}))
        self.assertEqual(excluded, [sha])
        self.assertTrue(rows[0]['association_exposed'])
        self.assertFalse(rows[1]['association_exposed'])
        self.assertEqual(len(reported), 2)
        self.assertEqual(counts['x2d-merge-association-pilot']['resolved'], 1)
        self.assertEqual(counts['x2d-merge-association-pilot']['unresolved'], 3)
        self.assertFalse(inventory.matches({'blob': inventory.EMPTY_BLOB}, sha, set(), {inventory.EMPTY_BLOB}))

    def test_rename_warning_and_isolation(self):
        self.assertTrue(inventory.rename_detection_skipped(b'warning: exhaustive rename detection was skipped due to too many files.'))
        self.assertTrue(inventory.rename_detection_skipped(b'warning: inexact rename detection was skipped due to too many files.'))
        self.assertFalse(inventory.rename_detection_skipped(b''))
        with patch.dict('os.environ', {'GIT_CONFIG_COUNT': '1', 'GIT_CONFIG_GLOBAL': '/untrusted'}):
            environment = inventory.git_env()
            self.assertEqual(environment['GIT_CONFIG_COUNT'], '0')
            self.assertEqual(environment['GIT_CONFIG_GLOBAL'], '/dev/null')
            self.assertEqual(environment['GIT_CONFIG_NOSYSTEM'], '1')
        with patch.object(subprocess, 'run', return_value=subprocess.CompletedProcess([], 0,
                stderr=b'warning: rename detection was skipped')) as run:
            self.assertTrue(inventory.check_rename_detection('/synthetic', '1'*40, '2'*40))
            self.assertEqual(run.call_args.kwargs['env']['GIT_CONFIG_NOSYSTEM'], '1')

    def test_pooled_threshold_not_age_veto(self):
        gates = {'quality': True, 'denominators': True, 'missingness': True, 'cap': True, 'interval_validity': True}
        self.assertTrue(inventory.pooled_accuracy_support([.78, .92], .85, .80, gates))
        for gate in gates:
            failed = dict(gates, **{gate: False})
            self.assertFalse(inventory.pooled_accuracy_support([.78, .92], .85, .80, failed))
        self.assertFalse(inventory.pooled_accuracy_support([.78, .92], .85, .799, gates))
        with self.assertRaises(ValueError):
            inventory.pooled_accuracy_support([.78, .92], .85, .80, {'quality': True})

    def test_fixed_projection_vectors(self):
        def cluster(identifier, kind):
            return {'id': identifier, 'directory': 'src', 'members': [{'path': 'src/a.go', 'kind': kind}]}
        rows = [{'commit': '1'*40, 'association_exposed': False,
            'selected': [cluster('added', 'added'), cluster('renamed', 'renamed'), cluster('modified', 'modified')]},
            {'commit': '2'*40, 'association_exposed': True, 'selected': [cluster('exposed', 'added')]},
            {'commit': '3'*40, 'association_exposed': False, 'selected': []}]
        result = inventory.projection_vectors(rows, 'synthetic/repo')
        self.assertEqual(result['population_selected_counts']['all'], [3, 0, 0])
        self.assertEqual(result['population_selected_counts']['added_only'], [1, 0, 0])
        self.assertEqual(result['resource_proxy_selected_including_exposed'], 4)
        self.assertEqual(result['label_proxy_upper_bound_per_cell'], 3)
        for age, probabilities in [('Cold', [90, 75, 0]), ('C0', [80, 50, 0])]:
            expected = 0
            for identifier, probability in zip(('added', 'renamed', 'modified'), probabilities):
                key = f'x11a-project-v4\0synthetic/repo\0{age}\0{"1"*40}\0{identifier}'
                expected += int(inventory.hashlib.sha256(key.encode()).hexdigest(), 16) < probability * (2**256) / 100
            self.assertEqual(result['projected_selected_counts'][age], [expected, 0, 0])
            self.assertEqual(result['projected_label_capacity_per_cell'][age], expected)
            self.assertEqual(result['projected_label_selected_counts'][age], [expected, 0, 0])
        many = [{'commit': '4'*40, 'association_exposed': False,
            'selected': [cluster(str(i), 'modified') for i in range(2601)]}]
        many_result = inventory.projection_vectors(many, 'synthetic/repo')
        self.assertEqual(many_result['label_proxy_upper_bound_per_cell'], 2600)
        self.assertEqual(many_result['projected_label_capacity_per_cell'], {'Cold': 0, 'C0': 0})
        label_rows = [{'commit': str(position) * 40, 'association_exposed': False,
            'selected': [cluster(str(index), 'added') for index in range(2000)]}
            for position in (4, 5)]
        capped = inventory.projection_vectors(label_rows, 'synthetic/repo')
        for age in ('Cold', 'C0'):
            self.assertEqual(sum(capped['projected_label_selected_counts'][age]), 2600)
            self.assertTrue(all(sampled <= available for sampled, available in zip(
                capped['projected_label_selected_counts'][age], capped['projected_selected_counts'][age])))
            cases = capped['projected_label_directory_cases'][age]
            self.assertEqual(len(cases), 2600)
            self.assertEqual(capped['largest_retained_directory_share'][age], 1)
            self.assertEqual(capped['retained_directory_count'][age], 1)
            self.assertEqual(capped['effective_directory_count'][age], 1)
            self.assertEqual(capped['retained_directory_case_count'][age], 2600)
            expected_hash = inventory.hashlib.sha256(b'x11a-directory-v6\0synthetic/repo\0src').hexdigest()
            self.assertTrue(all(case['directory_hash'] == expected_hash for case in cases))
            counts = [sum(case['position'] == position for case in cases) for position in range(2)]
            self.assertEqual(counts, capped['projected_label_selected_counts'][age])

    def test_projection_segment_lengths_and_directory_effective_count(self):
        for length in (1500, 1666, 2000):
            rows = [{'commit': str(position), 'association_exposed': False,
                'selected': [{'id': str(position), 'directory': 'a' if position % 4 else 'b',
                    'members': [{'path': 'a/file.go', 'kind': 'added'}]}]}
                for position in range(length)]
            result = inventory.projection_vectors(rows, 'synthetic/length')
            for age in ('Cold', 'C0'):
                self.assertEqual(len(result['projected_selected_counts'][age]), length)
                self.assertEqual(len(result['projected_label_selected_counts'][age]), length)
                cases = result['projected_label_directory_cases'][age]
                sizes = inventory.Counter(case['directory_hash'] for case in cases)
                self.assertEqual(result['retained_directory_count'][age], 2)
                self.assertEqual(result['effective_directory_count'][age],
                    len(cases)**2 / sum(n*n for n in sizes.values()))
        empty = inventory.projection_vectors([], 'synthetic/empty')
        self.assertEqual(empty['retained_directory_count'], {'Cold': 0, 'C0': 0})
        self.assertEqual(empty['effective_directory_count'], {'Cold': 0, 'C0': 0})
        self.assertEqual(empty['retained_directory_case_count'], {'Cold': 0, 'C0': 0})

    def test_content_refusal_selection_no_substitution(self):
        refusal = {'class': 'content', 'code': 'unsafe-symlink', 'message': 'synthetic only'}
        data = self.candidates()
        data[4] = {'repository': 'prometheus', 'preparation_refusal': refusal}
        selected = inventory.select_repositories(data)
        self.assertEqual(selected['selected_repository'], 'gin')
        self.assertTrue(all('prometheus' not in rank for rank in selected['rankings'].values()))
        self.assertFalse(data[4]['eligible'])
        data = self.candidates()
        data[0] = {'repository': 'django', 'preparation_refusal': refusal}
        self.assertEqual(inventory.select_repositories(data)['verdict'], 'not_testable_no_substitution')
        operational = self.candidates()
        operational[4] = {'repository': 'prometheus',
            'preparation_refusal': {'class': 'operational', 'code': 'disk-full', 'message': 'synthetic'}}
        with self.assertRaisesRegex(ValueError, 'nonterminal operational'):
            inventory.select_repositories(operational)
        operational[4]['preparation_refusal']['code'] = 'PreparationInfeasibleOperational'
        self.assertEqual(inventory.select_repositories(operational)['selected_repository'], 'gin')
        operational = self.candidates()
        operational[0] = {'repository': 'django', 'preparation_refusal':
            {'class': 'operational', 'code': 'PreparationInfeasibleOperational', 'message': 'synthetic'}}
        self.assertEqual(inventory.select_repositories(operational)['verdict'], 'not_testable_operational')
        for code in ('PreparationInterruptedOperational', 'PreparationBudgetExhausted'):
            for index in (0, 4):
                operational = self.candidates()
                operational[index] = {'repository': operational[index]['repository'],
                    'preparation_refusal': {'class': 'operational', 'code': code, 'message': 'synthetic'}}
                result = inventory.select_repositories(operational)
                self.assertEqual(result['verdict'], 'not_testable_operational')
                self.assertIsNone(result['selected_repository'])

    def test_repository_identity_exposure_is_descriptive_only(self):
        additional = {'records': [
            {'repository': 'r', 'study': 's', 'task': 'other', 'repository_identity': 'r'},
            {'repository': 'elsewhere', 'study': 's', 'task': 'other', 'repository_identity': 'elsewhere'}]}
        self.assertEqual(inventory.repository_exposure_by_study(additional, 'r', ['s', 'empty']),
            {'s': 1, 'empty': 0})
        row = {'commit': '1'*40, 'association_exposed': False, 'selected': []}
        before = copy.deepcopy(row)
        inventory.repository_exposure_by_study(additional, 'r', ['s'])
        self.assertEqual(row, before)

    def test_false_support_only_defective_own_arm(self):
        import oc_v5
        truth = {'primary': [.9, .9], 'jev_only': [.75, .9]}
        self.assertFalse(oc_v5.false_support(truth, {'primary': True, 'jev_only': False}))
        self.assertTrue(oc_v5.false_support(truth, {'primary': False, 'jev_only': True}))
        self.assertFalse(oc_v5.false_support({'primary': [.8, .8], 'jev_only': [.8, .8]},
            {'primary': True, 'jev_only': True}))
        both = {'primary': [.9, .78], 'jev_only': [.75, .9]}
        for arm in both:
            self.assertTrue(oc_v5.false_support(both, {key: key == arm for key in both}))

    def test_exact_count_plants_and_near_boundary(self):
        import oc_v5
        for correct in (95, 70, 50, 60, 79, 81):
            for repository in ('mandatory-a', 'mandatory-b', 'selected'):
                for age in ('Cold', 'C0'):
                    for arm in ('primary', 'jev_only', 'clawpatch_only'):
                        with self.subTest(repository=repository, age=age, arm=arm, correct=correct):
                            self.assertEqual(oc_v5.score_counts(correct, 100, 100, 0, 0),
                                {'worst_precision': [correct, 100], 'worst_recall': [correct, 100],
                                 'best_precision': [correct, 100], 'best_recall': [correct, 100]})
                            self.assertEqual(oc_v5.score_counts(correct, 100, 100, 3, 2),
                                {'worst_precision': [correct, 103], 'worst_recall': [correct, 105],
                                 'best_precision': [correct+3, 103], 'best_recall': [correct+3, 103]})

    def test_fixed_postwinner_projection_floor(self):
        for position, floor in ((0, 1404), (1, 1404), (4, 2106)):
            for age in ('Cold', 'C0'):
                data = self.candidates()
                projection = data[position]['segment_projection_inputs']
                projection['projected_selected_counts'][age] = [floor]
                projection['projected_label_capacity_per_cell'][age] = floor
                self.assertEqual(inventory.select_repositories(data)['selected_repository'], 'prometheus')
                projection['projected_selected_counts'][age] = [floor-1]
                verdict = inventory.select_repositories(data)
                self.assertEqual(verdict['concordant_winner'], 'prometheus')
                self.assertIsNone(verdict['selected_repository'])
                self.assertEqual(verdict['verdict'], 'not_testable_no_substitution')
                projection['projected_selected_counts'][age] = [floor]
                projection['projected_label_capacity_per_cell'][age] = floor-1
                self.assertIsNone(inventory.select_repositories(data)['selected_repository'])

    def test_resource_budget_counts_exposed_continuity(self):
        data = self.candidates()
        for r in data:
            r['segment_projection_inputs']['resource_proxy_selected_including_exposed'] = 3400
        verdict = inventory.select_repositories(data)
        self.assertEqual(verdict['projected_fallback_calls'], 40800)
        self.assertFalse(verdict['resource_gates']['fallback_calls'])
        self.assertIsNone(verdict['selected_repository'])
        self.assertEqual(verdict['concordant_winner'], 'prometheus')
        data = self.candidates()
        data[4]['segment_projection_inputs']['resource_proxy_selected_including_exposed'] = 9000
        self.assertFalse(inventory.select_repositories(data)['resource_gates']['jev_repository_cost'])
        valid = inventory.select_repositories(self.candidates())
        self.assertEqual(valid['reserved_sessions'], 149940)
        self.assertTrue(valid['pre_map_calibration_required'])
        self.assertEqual(set(valid['resource_gates']), {'fallback_calls', 'jev_total_cost', 'jev_repository_cost'})
        self.assertNotIn('projected_subscription_model_hours', valid)

    def test_immutable_tool_platform_closure(self):
        import copy
        import pin_tool_closure as pin
        config = json.loads((inventory.ROOT/'inventory-inputs.json').read_text())
        pin.verify_preparation_closure(config)
        wrong = copy.deepcopy(config)
        wrong['export_preparation']['tools']['git']['sha256'] = '0'*64
        with self.assertRaisesRegex(ValueError, 'executable/immutable'):
            pin.verify_preparation_closure(wrong)
        with patch.object(pin.platform, 'machine', return_value='wrong-architecture'):
            with self.assertRaisesRegex(ValueError, 'OS build drift'):
                pin.verify_preparation_closure(config)
        for key in ('tools', 'python_install'):
            metadata = json.loads((inventory.ROOT/'tool-pins-v6'/'persistent-locations-v6b.json').read_text())
            files = metadata[key]['files']
            files[next(iter(files))]['sha256'] = '0'*64
            with patch.object(pin, 'load_pin', return_value=metadata):
                with self.assertRaisesRegex(ValueError, 'installation closure drift'):
                    pin.verify_preparation_closure(config)

    def test_produced_authorization_chain(self):
        from prep_selftest import PreparationTests, synthetic_artifact_chain, write_json
        import prepare_exports as preparation
        fixture = PreparationTests()
        fixture.setUp()
        self.addCleanup(fixture.doCleanups)
        repo, commit = fixture.repository()
        config, inputs, authority = fixture.authority(repo, commit)
        artifacts, extras = synthetic_artifact_chain(fixture.root, config, inputs, authority)
        artifact_path = fixture.root/'artifacts.json'
        write_json(artifact_path, artifacts)
        inventory_authority = fixture.root/'inventory-authority.json'
        original = preparation.load(authority)
        original['scope'] = 'inventory'
        original['files'].extend([*extras, preparation.file_pin(artifact_path)])
        for pin in [artifacts['frontend'], *artifacts['frontend_build'].values(),
                    artifacts['exposure'], artifacts['unscoped_exposure'], artifacts['repository_exposure']]:
            original['files'].append(pin)
        write_json(inventory_authority, original)
        stages.verify_authorization(inventory_authority, inputs, 'inventory', artifact_path)
        for failure in ('missing_dependency', 'unsealed_frontend', 'ownership_status',
                        'stale_source_binding', 'extra_outcome'):
            broken = copy.deepcopy(artifacts)
            if failure == 'missing_dependency':
                broken['preparation_seals'].pop()
            elif failure == 'unsealed_frontend':
                broken['frontend'] = config['export_preparation']['tools']['git']
            elif failure == 'ownership_status':
                broken['ownership']['synthetic/preparation']['status'] = 'refused'
            elif failure == 'stale_source_binding':
                broken['source_inputs_sha256'] = '0'*64
            else:
                broken['label'] = 'forbidden'
            write_json(artifact_path, broken)
            rebound = copy.deepcopy(original)
            rebound['files'] = [preparation.file_pin(artifact_path) if pin['path'] == str(artifact_path)
                                else pin for pin in rebound['files']]
            write_json(inventory_authority, rebound)
            with self.subTest(failure=failure), self.assertRaises(ValueError):
                stages.verify_authorization(inventory_authority, inputs, 'inventory', artifact_path)

    def test_scenario_index_and_seed_literals(self):
        import oc_v5
        for indices, expected_index, replicate, stream, expected_seed in (
            ((0, 0, 0, 0, 0, 0, 0), 0, 0, 'study', 3040335876284042375),
            ((11, 38, 1, 2, 1, 4, 1), 56159, 1023, 'study', 9710704855294934111),
            ((0, 11, 1, 2, 1, 4, 0), 1438, 7, 'bootstrap/2', 6257622249336655891)):
            self.assertEqual(oc_v5.scenario_index(*indices), expected_index)
            self.assertEqual(oc_v5.scenario_seed(expected_index, replicate, stream), expected_seed)

    def test_additive_capacity_fit_and_fail_closed_gate(self):
        from fractions import Fraction
        from preparation_capacity_v8 import fit, projected
        observations = [{'files': files, 'bytes': size, 'revisions': 3,
                         'elapsed_seconds': 3*(2*files + size/10 + 3)}
                        for files, size in ((10, 100), (20, 100), (10, 200))]
        coefficients = fit(observations)
        self.assertEqual(coefficients, (Fraction(2), Fraction(1, 10), Fraction(3)))
        rows = [{'files': 10, 'bytes': 100}]
        self.assertEqual(projected(rows, coefficients, Fraction(1))['per_attempt_seconds'], 33)
        self.assertEqual(projected(rows, coefficients)['per_attempt_seconds'], 49.5)
        self.assertFalse(projected([{'files': 1000000, 'bytes': 100}], coefficients)['attempt_fits'])
        with self.assertRaisesRegex(ValueError, 'not identifiable'):
            fit([observations[0]]*3)
        negative = [dict(row, elapsed_seconds=3*(100-row['files'])) for row in observations]
        with self.assertRaisesRegex(ValueError, 'negative'):
            fit(negative)

    def test_revision8_capacity_contract(self):
        from capacity_selftest_v8 import run
        run()

    def test_offline_synthetic_preparation(self):
        from prep_selftest import run
        run(inventory.ROOT/'inventory-inputs.json')


if __name__ == '__main__':
    unittest.main()

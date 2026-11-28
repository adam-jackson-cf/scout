"""Synthetic controller boundary evidence; no scanners, models, or real acceptance."""
import copy
import json
import os
from pathlib import Path
import sys
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import Mock, patch

# Reuse the reviewed evidence builders rather than maintaining a second schema.
import test_binding as helpers
import acceptance as controller
import attest_closure
import preparation as prep
from interlock import InterlockFailure
from verify_inputs import pin


class ControllerKeyBoundaryTests(unittest.TestCase):
    def test_reviewed_inputs_checks_presence_not_truthiness(self):
        for value in (None, '', 'synthetic-not-a-secret'):
            with self.subTest(value=value), tempfile.TemporaryDirectory() as directory:
                config = {'paths': {'STUDY': directory, 'OUT': directory}}
                environment = {} if value is None else {'OPENROUTER_API_KEY': value}
                with patch.dict(os.environ, environment, clear=True), \
                     patch.object(controller.subprocess, 'run') as launch:
                    # Without the key, the next real guard is reached. No adoption
                    # or review validation is mocked to manufacture a passing input.
                    message = 'explicit v2 binding recipe' if value is None else '^key present$'
                    with self.assertRaisesRegex(InterlockFailure, message):
                        controller.reviewed_inputs(config, [], {})
                launch.assert_not_called()

    def test_rehearsal_entry_checks_presence_before_output_reservation(self):
        for value in (None, '', 'synthetic-not-a-secret'):
            with self.subTest(value=value), tempfile.TemporaryDirectory() as directory:
                root = Path(directory).resolve()
                out = root/'out'
                out.mkdir()
                config = {'paths': {'STUDY': str(controller.EXEC.parent), 'OUT': str(out),
                                    'SCRATCH': str(root/'scratch')},
                          'execution_commit': 'synthetic',
                          'deferred_to_b_contract': {'gitlink': 'synthetic'}}
                config_path = root/'config.json'
                config_path.write_text(json.dumps(config))
                (out/'input-verification.json').write_text('{}')
                clock = out/'clock.json'
                clock.write_text(json.dumps(prep.clock_value(config, config_path, 'a',
                                                             '2026-01-01T00:00:00Z')))
                output = root/'scratch/report.json'
                args = SimpleNamespace(fixture='synthetic', objects=None, rehearsal_report=None,
                                       config=config_path, clock=clock, output=output)
                environment = {} if value is None else {'OPENROUTER_API_KEY': value}
                with patch.dict(os.environ, environment, clear=True), \
                     patch.object(controller.subprocess, 'run') as launch:
                    if value is None:
                        self.assertEqual(controller.rehearse(config, args, []), 1)
                        report = json.loads(output.read_bytes())
                        self.assertIn('explicit v2 binding recipe', report['failures'][0])
                    else:
                        with self.assertRaisesRegex(InterlockFailure, '^key present$'):
                            controller.rehearse(config, args, [])
                        self.assertFalse(output.exists())
                        self.assertFalse(output.parent.exists())
                launch.assert_not_called()

    def test_every_launch_checks_key_presence_before_direct_execution(self):
        for route in controller.ROUTE_DIRS:
            for resolve_only in (False, True):
                for value in (None, '', 'synthetic-not-a-secret'):
                    with self.subTest(route=route, resolve_only=resolve_only, value=value), \
                         tempfile.TemporaryDirectory() as directory:
                        root = Path(directory)
                        runtime = launch_runtime()
                        environment = {} if value is None else {'OPENROUTER_API_KEY': value}
                        def frontend(argv, **kwargs):
                            write_frontend_outputs(argv, route)
                            if route == 'synthetic-unbound' and not resolve_only:
                                write_attempt(root/'run', 'claude')
                            return Mock(returncode=0)
                        with patch.dict(os.environ, environment, clear=True), \
                             patch.object(attest_closure, 'verify_installed_closure') as closure, \
                             patch.object(controller.subprocess, 'run', side_effect=frontend) as launch:
                            if value is None:
                                record = controller.run_revision(root, 'sha', route, root/'run', runtime,
                                                                 root/'policy', root/'closure',
                                                                 resolve_only=resolve_only)
                                self.assertFalse(record['key_present'])
                                launch.assert_called_once()
                                self.assertEqual(closure.call_count,2)
                            else:
                                with self.assertRaisesRegex(InterlockFailure, 'key present at launch'):
                                    controller.run_revision(root, 'sha', route, root/'run', runtime,
                                                            root/'policy', root/'closure',
                                                            resolve_only=resolve_only)
                                launch.assert_not_called()
                                closure.assert_not_called()
                                self.assertFalse((root/'run').exists())


def launch_runtime():
    return {'python': {'path': sys.executable}, 'executables': [],
            'frontend': {'path': '/synthetic/frontend'}, 'clawpatch': {'path': '/synthetic/clawpatch'}}


def write_frontend_outputs(argv, route):
    policy = helpers.model_proof()
    if route == 'synthetic-unbound':
        policy['original_policy'] = {'model_cli': {'provider': 'claude'}}
    Path(argv[argv.index('--effective-output')+1]).write_text(json.dumps(policy))
    if '--output' in argv:
        Path(argv[argv.index('--output')+1]).write_text(json.dumps(helpers.lane_result('bound')))


def write_attempt(run, provider):
    (run/'stubs/attempts.jsonl').write_text(json.dumps(helpers.ModelAttemptTests().attempt(provider))+'\n')


class PostLaunchClosureTests(unittest.TestCase):
    def test_real_closure_null_and_planted_launch_writes(self):
        cases=[('bound',None,0),('synthetic-unbound',None,1),('reference',None,0),
               ('bound','append',0),('synthetic-unbound','append',1),('reference','added',0)]
        for route,mutation,exit_code in cases:
            with self.subTest(route=route,mutation=mutation), tempfile.TemporaryDirectory() as directory:
                root=Path(directory).resolve();runtime_root=root/'runtime';runtime_root.mkdir()
                source=runtime_root/'file';source.write_bytes(b'frozen')
                link=runtime_root/'link';link.symlink_to('file')
                inventory=attest_closure.Inventory([runtime_root]);enumeration=inventory.collect()
                retained=attest_closure.retain(list(inventory.files.values()),root/'retained')
                value={'schema':'scout.x1.installed-closure.v1','status':'passed','complete':True,
                       'enumeration':enumeration,'inputs':[],'source_pins':[],
                       'original_runtime_files':list(inventory.files.values()),
                       'submodule_working_files':[],'deferred_working_files_verified':[],
                       'retained_files':retained,'executable_resolutions':[attest_closure.checked(link,pin(source))]}
                receipt=root/'closure.json';receipt.write_text(json.dumps(value))
                events=[];real_verify=attest_closure.verify_installed_closure
                def verify(path):
                    events.append('verify')
                    return real_verify(path)
                def frontend(argv,**kwargs):
                    events.append('launch')
                    write_frontend_outputs(argv,route)
                    if route=='synthetic-unbound':
                        write_attempt(root/'run','claude')
                    if mutation=='append':
                        with source.open('ab') as stream:stream.write(b'!')
                    elif mutation=='added':
                        (runtime_root/'added').write_bytes(b'added')
                    return Mock(returncode=exit_code)
                with patch.dict(os.environ,{},clear=True), \
                     patch.object(attest_closure,'verify_installed_closure',side_effect=verify) as checked, \
                     patch.object(controller.subprocess,'run',side_effect=frontend) as launched:
                    if mutation is None:
                        controller.run_revision(root,'sha',route,root/'run',launch_runtime(),
                                                root/'policy',receipt)
                    else:
                        with self.assertRaisesRegex(ValueError,'file/link/mode set drift'):
                            controller.run_revision(root,'sha',route,root/'run',launch_runtime(),
                                                    root/'policy',receipt)
                    self.assertEqual(events,['verify','launch','verify'])
                    self.assertEqual(checked.call_count,2)
                    launched.assert_called_once()


class PredicateBoundaryTests(unittest.TestCase):
    def test_positive_control_false_boundaries(self):
        cases = [('one-control', None, None), ('swapped-providers', None, None),
                 ('duplicate-providers', None, None), ('no-attempt', 'model_attempts', 0),
                 ('wrong-exit', 'stub_exit_code', 96)]
        self.assertTrue(all(controller.acceptance_flags(helpers.DeterministicProofTests().evidence()).values()))
        for name, field, value in cases:
            for index in range(2) if field else (0,):
                with self.subTest(name=name, control=index):
                    evidence = helpers.DeterministicProofTests().evidence()
                    controls = evidence['positive_control']
                    if name == 'one-control':
                        controls.pop()
                    elif name == 'swapped-providers':
                        controls.reverse()
                    elif name == 'duplicate-providers':
                        controls[1]['interlock']['provider'] = 'claude'
                    else:
                        controls[index]['interlock'][field] = value
                    self.assertFalse(controller.acceptance_flags(evidence)['positive_control_expected_failure'])

    def test_each_bound_reference_attempt_count_must_be_zero(self):
        for route in ('bound', 'reference'):
            for revision in ('parent', 'head'):
                with self.subTest(route=route, revision=revision):
                    evidence = helpers.DeterministicProofTests().evidence()
                    evidence['runs'][route][revision]['interlock']['model_attempts'] = 1
                    self.assertFalse(controller.acceptance_flags(evidence)['zero_model_attempts'])

    def test_resolve_attempt_and_exit_status_boundaries(self):
        cases = [('bound', False, 0, False, None),
                 ('bound', False, 7, False, 'frontend/scanner exited nonzero'),
                 ('reference', False, 7, False, 'frontend/scanner exited nonzero'),
                 ('synthetic-unbound', True, 0, False, None),
                 ('synthetic-unbound', True, 0, True, 'model'),
                 ('synthetic-unbound', True, 7, False, 'frontend/scanner exited nonzero'),
                 ('synthetic-unbound', False, 97, True, None)]
        for route, resolve_only, exit_code, attempt, refusal in cases:
            with self.subTest(route=route, resolve_only=resolve_only, exit_code=exit_code, attempt=attempt), \
                 tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                def frontend(argv, **kwargs):
                    write_frontend_outputs(argv, route)
                    if attempt:
                        write_attempt(root/'run', 'claude')
                    return Mock(returncode=exit_code)
                with patch.dict(os.environ, {}, clear=True), \
                     patch.object(attest_closure, 'verify_installed_closure'), \
                     patch.object(controller.subprocess, 'run', side_effect=frontend) as launch:
                    if refusal:
                        with self.assertRaisesRegex(InterlockFailure, refusal):
                            controller.run_revision(root, 'sha', route, root/'run', launch_runtime(),
                                                    root/'policy', root/'closure', resolve_only=resolve_only)
                    else:
                        record = controller.run_revision(root, 'sha', route, root/'run', launch_runtime(),
                                                         root/'policy', root/'closure', resolve_only=resolve_only)
                        self.assertEqual(record['interlock']['model_attempts'], int(attempt))
                        if attempt:
                            self.assertTrue(record['interlock']['expected_interlock_failure'])
                    launch.assert_called_once()


class DjangoFixtureBoundaryTests(unittest.TestCase):
    def fixture(self, root):
        out = root/'out'
        (out/'binding').mkdir(parents=True)
        (out/'objects').mkdir()
        config = {'paths': {'STUDY': str(root), 'OUT': str(out)}}
        synthetic = helpers.DeterministicProofTests().evidence()
        synthetic.update(schema='scout.x1.binding-acceptance.v2', status='passed')
        synthetic['acceptance'] = controller.acceptance_flags(synthetic)
        synthetic_path = out/'binding/synthetic.json'
        synthetic_path.write_text(json.dumps(synthetic))
        objects = {'status': 'passed', 'repository': 'django/django',
                   'database_realpath': str(root/'synthetic-objects.git'),
                   'actual_tip': 'a05e16a7230fe33d3f43f4a3aaed103cebb42efc',
                   'django_first_parent': {'sha': 'synthetic-parent'}}
        object_path = out/'objects/django__django.json'
        object_path.write_text(json.dumps(objects))
        return config, synthetic, synthetic_path, objects, object_path

    def test_django_manifest_refusal_table(self):
        for case in ('valid', 'substituted-tip', 'parent-mismatch', 'noncanonical-objects',
                     'failed-objects', 'wrong-repository', 'failed-synthetic', 'wrong-synthetic-schema',
                     'false-derived-synthetic-acceptance'):
            with self.subTest(case=case), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                config, synthetic, synthetic_path, objects, object_path = self.fixture(root)
                requested_objects = object_path
                parent = objects['django_first_parent']['sha']
                if case == 'substituted-tip':
                    objects['actual_tip'] = 'b'*40
                elif case == 'parent-mismatch':
                    parent = 'different-parent'
                elif case == 'noncanonical-objects':
                    requested_objects = root/'substituted.json'
                    requested_objects.write_text(json.dumps(objects))
                elif case == 'failed-objects':
                    objects['status'] = 'refused'
                elif case == 'wrong-repository':
                    objects['repository'] = 'other/repository'
                elif case == 'failed-synthetic':
                    synthetic['status'] = 'refused'
                elif case == 'wrong-synthetic-schema':
                    synthetic['schema'] = 'scout.x1.binding-acceptance.v1'
                elif case == 'false-derived-synthetic-acceptance':
                    synthetic['runs']['bound']['head']['interlock']['model_attempts'] = 1
                object_path.write_text(json.dumps(objects))
                synthetic_path.write_text(json.dumps(synthetic))
                with patch.object(controller, 'git', return_value=(parent+'\n').encode()) as git:
                    if case == 'valid':
                        manifest, positive = controller.fixture_manifest(
                            config, {'git': {'path': '/synthetic/git'}}, 'django-pinned-pair',
                            requested_objects, root/'scratch')
                        self.assertEqual(manifest['synthetic_receipt'], pin(synthetic_path))
                        self.assertEqual(positive, synthetic['positive_control'])
                        git.assert_called_once_with('/synthetic/git', objects['database_realpath'],
                                                    'rev-parse', '--verify', objects['actual_tip']+'^1')
                    else:
                        with self.assertRaises(ValueError):
                            controller.fixture_manifest(config, {'git': {'path': '/synthetic/git'}},
                                                        'django-pinned-pair', requested_objects, root/'scratch')

    def test_django_evidence_pins_exact_synthetic_receipt_and_detects_change(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            config, synthetic, synthetic_path, objects, object_path = self.fixture(root)
            with patch.object(controller, 'git', return_value=b'synthetic-parent\n'), \
                 patch.object(controller, 'materialize', side_effect=materialize_empty), \
                 patch.object(controller, 'run_revision', side_effect=scanner_record):
                evidence = controller.execute_fixture(config, {'git': {'path': '/synthetic/git'}},
                                                      'django-pinned-pair', object_path, root/'scratch')
            self.assertEqual(evidence['synthetic_receipt'], pin(synthetic_path))
            self.assertEqual(evidence['synthetic_receipt'], evidence['fixture_manifest']['synthetic_receipt'])
            self.assertEqual(evidence['positive_control'], synthetic['positive_control'])
            prep.verify_tree_pins(evidence)
            synthetic['status'] = 'refused'
            synthetic_path.write_text(json.dumps(synthetic))
            with self.assertRaisesRegex(ValueError, 'pin mismatch'):
                prep.verify_tree_pins(evidence)

    def test_equal_synthetic_parent_head_issue_sets_refuse(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            config = {'paths': {'STUDY': str(root), 'OUT': str(root/'out')}}
            manifest = {'database': 'synthetic-db', 'parent': 'parent-sha', 'head': 'head-sha'}
            with patch.object(controller, 'generate', return_value=manifest), \
                 patch.object(controller, 'materialize', side_effect=materialize_empty), \
                 patch.object(controller, 'run_revision', side_effect=scanner_record):
                with self.assertRaisesRegex(InterlockFailure, 'synthetic parent/head issue sets must differ'):
                    controller.execute_fixture(config, {'git': {'path': '/synthetic/git'}},
                                               'synthetic', None, root/'scratch')


def materialize_empty(database, revision, root, git):
    # A scanner-route stand-in still provides the state directory carry_state needs.
    (root/'.scout').mkdir(parents=True)
    return []


def scanner_record(root, revision, route, run, runtime, frozen, closure,
                   resolve_only=False, expected_policy=None):
    evidence = helpers.DeterministicProofTests().evidence()
    selected_route = 'bound' if route == 'synthetic-unbound' else route
    record = copy.deepcopy(evidence['runs'][selected_route]['parent'])
    record['policy'].update(route=route, revision=revision)
    record['result'].update(route=route, revision=revision)
    # Keep result identities equal deliberately: only the synthetic fixture forbids it.
    record['effective_record'] = {'synthetic_effective_record': True}
    return record


if __name__ == '__main__':
    unittest.main()

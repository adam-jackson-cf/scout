"""Synthetic adoption-to-binding authority integration; no binaries are executed."""
import copy
import json
import os
from pathlib import Path
import shutil
import sys
import tempfile
import unittest
from unittest.mock import patch

EXEC = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(EXEC))
sys.path.insert(0, str(EXEC/'binding'))
import acceptance as controller
import preparation as prep
from interlock import InterlockFailure
from verify_inputs import exclusive_json, pin


class BindingAdoptionIntegrationTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.study = Path(self.temp.name).resolve()/'study'
        self.execution = self.study/'execution'
        self.out = self.study/'preparation-run'
        self.execution.mkdir(parents=True)
        self.addCleanup(patch.stopall)
        patch.object(controller, 'EXEC', self.execution).start()
        patch.dict(os.environ, {}, clear=True).start()

    def write_json(self, path, value):
        exclusive_json(path, value)
        return path

    def write_bytes(self, path, value):
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(value)
        return path

    def review_record(self, path):
        return self.study/'preparation-reviews'/(path.stem+'.council.txt')

    def review(self, path, scope, reviewed):
        record = self.review_record(path)
        self.write_bytes(record, ('Synthetic council fixture only.\n**Zero issues. Signed off.**\n'+
                                 'Signed-off scope: '+scope+'\n'+
                                 '\n'.join(row['sha256'] for row in reviewed)+'\n').encode())
        self.write_json(path, {'scope': scope, 'record': pin(record), 'reviewed': reviewed})
        return path

    def stage(self, omit=None):
        self.config_path = self.execution/'run-config.json'
        self.write_json(self.config_path, {
            'schema': 'scout.x1.run-config.v1',
            'execution_commit': 'synthetic-commit',
            'paths': {'STUDY': str(self.study), 'OUT': str(self.out),
                      'SCRATCH': str(self.study/'scratch')},
            'source_manifest': 'execution-source-manifest.json',
            'deferred_to_b_contract': {'gitlink': 'synthetic-gitlink'},
            'required_records': [
                {'path': 'binding/synthetic.json', 'kind': 'scout.x1.binding-acceptance.v1'},
                {'path': 'binding/django.json', 'kind': 'scout.x1.binding-acceptance.v1'}],
            'producer_contracts': {'binding': 'synthetic-original'}})
        # read_clock deliberately requires the immutable historical authority hashes.
        # Copy those public records locally; all mutable authority is synthetic.
        for name in ('round-9.md', 'rev9-digests.json'):
            destination = self.study/'reviews'/name
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(EXEC.parent/'reviews'/name, destination)
        verification = self.write_json(self.out/'input-verification.json', {'status': 'passed'})
        self.write_json(self.out/'clock.json', {
            'schema': 'scout.x1.preparation-clock.v1', 'study': 'X1',
            'execution_commit': 'synthetic-commit', 'first_action': 'a',
            'preparation_started_at': '2020-01-01T00:00:00Z',
            'gold_freeze_deadline': '2020-02-15T00:00:00Z',
            'authority': pin(self.study/'reviews/round-9.md'),
            'manifest': pin(self.study/'reviews/rev9-digests.json'),
            'execution_config': pin(self.config_path), 'input_verification': pin(verification),
            'deferred_submodule_gitlink': 'synthetic-gitlink'})
        self.recipe_path = self.write_json(
            self.study/'preparation/model-free-binding-check-v2.json', {'synthetic': 'recipe'})
        self.amendment_path = self.write_bytes(
            self.study/'preparation/model-free-binding-amendment.md',
            b'Synthetic amendment text only.\n')
        self.sources = [pin(self.write_bytes(self.execution/name, content)) for name, content in
                        [('binding/acceptance.py', b'synthetic controller source\n'),
                         ('binding/interlock.py', b'synthetic interlock source\n')]]
        self.proposal = self.execution/'run-config-amendments-proposed/001-binding-v2.json'
        self.adopted = self.execution/'run-config-amendments/001-binding-v2.json'
        self.review_path = self.out/'reviews/binding-v2-amendment.json'
        self.source_path = self.out/'execution-source-manifest.json'
        self.write_json(self.proposal, {
            'schema': 'scout.x1.run-config-amendment.v1',
            'amends': {'path': 'execution/run-config.json', 'sha256': pin(self.config_path)['sha256']},
            'reason': 'synthetic binding adoption regression',
            'authority': {'path': 'preparation-run/reviews/binding-v2-amendment.json',
                          'scope': 'model-free-binding-v2-amendment'},
            'changes': {'binding_record_kind': 'scout.x1.binding-acceptance.v2',
                        'binding_recipe': 'preparation/model-free-binding-check-v2.json',
                        'binding_producer_contract': 'synthetic-v2'}})
        reviewed = [pin(self.proposal), pin(self.recipe_path), pin(self.amendment_path)]
        if omit == 'proposal':
            # adopt checks the digest, whereas binding must also check its proposal path.
            # A byte-identical alternate path isolates that stricter binding boundary.
            alternate = self.write_bytes(self.execution/'reviewed-amendment-copy.json',
                                         self.proposal.read_bytes())
            reviewed[0] = pin(alternate)
        elif omit in ('recipe', 'amendment'):
            omitted_path = self.recipe_path if omit == 'recipe' else self.amendment_path
            reviewed = [row for row in reviewed if row['path'] != str(omitted_path)]
        elif omit == 'extra':
            reviewed.append(self.sources[-1])
        self.review(self.review_path, 'model-free-binding-v2-amendment', reviewed)
        self.preadoption_review = self.review_path.read_bytes()
        self.preadoption_record = self.review_record(self.review_path).read_bytes()
        self.assertFalse(self.adopted.exists())
        self.assertFalse(self.source_path.exists())
        reviewed_paths = {row['path'] for row in reviewed}
        self.assertNotIn(str(self.adopted), reviewed_paths)
        self.assertNotIn(str(self.source_path), reviewed_paths)

    def adopt_and_prepare_runtime(self):
        argv = [str(EXEC/'preparation.py'), 'adopt', '--config', str(self.config_path),
                '--amendment', str(self.proposal), '--review', str(self.review_path)]
        with patch.object(sys, 'argv', argv):
            adoption_pin = prep.adopt(self.config_path, self.proposal, self.review_path)
        config = prep.load_config(self.config_path)
        self.assertEqual(adoption_pin, pin(prep.adoption_paths(config, self.adopted)[0]))
        self.assertEqual(self.adopted.read_bytes(), self.proposal.read_bytes())
        self.assertEqual(config.amendments[0]['adoption_review'], pin(self.review_path))
        self.assertEqual(config['binding_recipe'], 'preparation/model-free-binding-check-v2.json')
        # These records intentionally do not exist until the real adoption has finished.
        # Source pins belong only to the post-adoption implementation review.
        runtime_files = {}
        for name in ('git', 'python', 'frontend', 'clawpatch', 'node'):
            executable = self.write_bytes(self.study/'dummy-bin'/name,
                                          ('never execute '+name+'\n').encode())
            executable.chmod(0o700)
            runtime_files[name] = pin(executable)
        goroot = self.study/'dummy-go'
        go = self.write_bytes(goroot/'bin/go', b'never execute go\n')
        go.chmod(0o700)
        sandbox = pin('/usr/bin/sandbox-exec')
        support = [runtime_files['git'], pin(go), sandbox]
        permitted = [runtime_files[name] for name in ('python', 'clawpatch', 'node')]
        self.runtime = dict(schema='scout.x1.binding-runtime.v2',
                            **{name: runtime_files[name] for name in ('git', 'python', 'frontend', 'clawpatch')},
                            executables=permitted+[runtime_files['frontend']]+support)
        self.write_json(self.out/'closure.json', {
            'status': 'passed', 'complete': True,
            'git': runtime_files['git'], 'go': {'goroot': str(goroot)},
            'os_executables': [runtime_files['git'], sandbox],
            'original_runtime_files': [runtime_files[name] for name in ('git', 'python', 'clawpatch')]
                                      + [pin(go), sandbox],
            'executable_resolutions': [dict(runtime_files[name], name=name, execution_allowed=True)
                                       for name in ('python', 'clawpatch', 'node')]})
        self.runtime_path = self.write_json(self.out/'binding/runtime.json', self.runtime)
        self.refresh_implementation_review()
        self.assert_review_unchanged()
        return config

    def refresh_implementation_review(self, omit_source=None, reviewed_paths=()):
        files = [pin(path) for path in sorted(self.execution.rglob('*'))
                 if path.is_file() and str(path) != omit_source]
        self.assertIn(pin(self.adopted), files)
        self.source_path.unlink(missing_ok=True)
        self.write_json(self.source_path, {'files': files})
        implementation_review = self.out/'reviews/implementation.json'
        implementation_review.unlink(missing_ok=True)
        self.review(implementation_review, 'execution-implementation',
                    [pin(self.source_path), pin(self.runtime_path),
                     *(pin(path) for path in reviewed_paths)])

    def assert_review_unchanged(self):
        self.assertEqual(self.review_path.read_bytes(), self.preadoption_review)
        self.assertEqual(self.review_record(self.review_path).read_bytes(),
                         self.preadoption_record)

    def test_support_pins_missing_wrong_or_duplicate_refuse_with_passing_null(self):
        self.stage()
        self.adopt_and_prepare_runtime()
        closure = json.loads((self.out/'closure.json').read_bytes())
        support = (self.runtime['git']['path'], str(Path(closure['go']['goroot'])/'bin/go'),
                   '/usr/bin/sandbox-exec')
        for path in support:
            for mutation in ('missing', 'wrong-sha256', 'duplicate', 'unattested-path'):
                with self.subTest(path=path, mutation=mutation):
                    controller.validate_runtime(self.runtime, closure)  # Planted null.
                    changed = copy.deepcopy(self.runtime)
                    changed['executables'] = [dict(row) for row in self.runtime['executables']]
                    row = next(row for row in changed['executables'] if row['path'] == path)
                    if mutation == 'missing':
                        changed['executables'].remove(row)
                    elif mutation == 'wrong-sha256':
                        row['sha256'] = '0'*64
                    elif mutation == 'duplicate':
                        changed['executables'].append(dict(row))
                    else:
                        row['path'] = str(self.study/'unattested'/Path(path).name)
                    with self.assertRaisesRegex(ValueError, 'support pin must appear exactly once'):
                        controller.validate_runtime(changed, closure)  # Planted effect.

    def test_unrelated_original_runtime_file_is_not_admitted(self):
        self.stage()
        self.adopt_and_prepare_runtime()
        closure = json.loads((self.out/'closure.json').read_bytes())
        controller.validate_runtime(self.runtime, closure)
        extra = self.write_bytes(self.study/'extra-bin/other-tool', b'never execute\n')
        extra.chmod(0o700)
        closure['original_runtime_files'].append(pin(extra))
        changed = copy.deepcopy(self.runtime)
        changed['executables'].append(pin(extra))
        with self.assertRaisesRegex(ValueError, 'executable outside attested closure'):
            controller.validate_runtime(changed, closure)

    def test_shadow_git_before_attested_directory_refuses_with_passing_null(self):
        self.stage()
        self.adopt_and_prepare_runtime()
        closure = json.loads((self.out/'closure.json').read_bytes())
        probe = self.write_bytes(self.study/'shadow-bin/probe', b'never execute\n')
        probe.chmod(0o700)
        closure['executable_resolutions'].append(dict(pin(probe), name='probe', execution_allowed=True))
        self.runtime['executables'].insert(0, pin(probe))
        controller.validate_runtime(self.runtime, closure)  # Directory alone is a null.
        shadow = self.write_bytes(probe.parent/'git', b'planted shadow, never execute\n')
        shadow.chmod(0o700)
        with self.assertRaisesRegex(ValueError, 'PATH resolution differs from attested pin: git'):
            controller.validate_runtime(self.runtime, closure)

    def test_planted_model_on_private_path_refuses_with_passing_null(self):
        self.stage()
        self.adopt_and_prepare_runtime()
        closure = json.loads((self.out/'closure.json').read_bytes())
        for name in ('claude', 'codex'):
            with self.subTest(name=name):
                controller.validate_runtime(self.runtime, closure)
                model = self.write_bytes(self.study/'dummy-bin'/name, b'planted, never execute\n')
                model.chmod(0o700)
                with self.assertRaisesRegex(ValueError, 'real model executable on private PATH: '+name):
                    controller.validate_runtime(self.runtime, closure)
                model.unlink()

    def test_real_adopt_load_and_entire_binding_authority_chain(self):
        self.stage()
        config = self.adopt_and_prepare_runtime()
        receipt = {}
        self.assertEqual(controller.reviewed_inputs(config, self.sources, receipt), self.runtime)
        self.assertEqual(receipt['binding_amendment_review'], pin(self.review_path))
        self.assertEqual(receipt['runtime'], pin(self.out/'binding/runtime.json'))
        self.assertEqual(receipt['runtime_closure'], pin(self.out/'closure.json'))
        self.assertEqual(len(receipt['configuration_amendment_reviews']), 1)
        self.assert_review_unchanged()
        # The authority remains loadable after the complete post-adoption review flow.
        self.assertEqual(prep.load_config(self.config_path).amendments, config.amendments)

    def test_proposal_digest_missing_refuses_real_adoption(self):
        self.stage()
        reviewed = [row for row in json.loads(self.preadoption_review)['reviewed']
                    if row['path'] != str(self.proposal)]
        # Rebuild only the pre-adoption synthetic authority, before it is frozen.
        self.review_path.unlink()
        self.review(self.review_path, 'model-free-binding-v2-amendment', reviewed)
        self.preadoption_review = self.review_path.read_bytes()
        self.preadoption_record = self.review_record(self.review_path).read_bytes()
        with self.assertRaisesRegex(ValueError, 'does not bind proposed amendment'):
            prep.adopt(self.config_path, self.proposal, self.review_path)
        self.assertFalse(self.adopted.exists())
        self.assert_review_unchanged()

    def test_proposal_path_missing_refuses_binding_even_with_reviewed_digest(self):
        self.stage(omit='proposal')
        config = self.adopt_and_prepare_runtime()
        with self.assertRaises(InterlockFailure):
            controller.reviewed_inputs(config, self.sources, {})
        self.assert_review_unchanged()

    def test_one_source_pin_missing_from_implementation_manifest_refuses_binding(self):
        self.stage()
        config = self.adopt_and_prepare_runtime()
        self.refresh_implementation_review(omit_source=self.sources[-1]['path'])
        with self.assertRaisesRegex(ValueError, 'binding source omitted from reviewed manifest'):
            controller.reviewed_inputs(config, self.sources, {})
        self.assert_review_unchanged()

    def test_extra_row_in_amendment_review_refuses_binding(self):
        self.stage(omit='extra')
        config = self.adopt_and_prepare_runtime()
        with self.assertRaises(InterlockFailure):
            controller.reviewed_inputs(config, self.sources, {})
        self.assert_review_unchanged()

    def test_duplicate_fourth_amendment_row_refuses_binding(self):
        self.stage()
        reviewed = json.loads(self.preadoption_review)['reviewed']
        self.review_path.unlink()  # Replace only pre-adoption synthetic authority.
        self.review(self.review_path, 'model-free-binding-v2-amendment',
                    reviewed + [reviewed[0]])
        self.preadoption_review = self.review_path.read_bytes()
        self.preadoption_record = self.review_record(self.review_path).read_bytes()
        config = self.adopt_and_prepare_runtime()
        with self.assertRaisesRegex(InterlockFailure, 'exactly proposal, recipe and amendment'):
            controller.reviewed_inputs(config, self.sources, {})
        self.assert_review_unchanged()

    def test_recipe_missing_from_amendment_review_refuses_binding(self):
        self.stage(omit='recipe')
        config = self.adopt_and_prepare_runtime()
        with self.assertRaises(InterlockFailure):
            controller.reviewed_inputs(config, self.sources, {})
        self.assert_review_unchanged()

    def test_amendment_missing_from_amendment_review_refuses_binding(self):
        self.stage(omit='amendment')
        config = self.adopt_and_prepare_runtime()
        with self.assertRaises(InterlockFailure):
            controller.reviewed_inputs(config, self.sources, {})
        self.assert_review_unchanged()

    def test_source_fix_after_real_adoption_accepts_fresh_implementation_review(self):
        self.stage()
        self.adopt_and_prepare_runtime()
        previous_source = self.sources[0]
        changed_source = self.write_bytes(Path(previous_source['path']),
                                          b'synthetic corrected controller source\n')
        self.sources[0] = pin(changed_source)
        self.assertNotEqual(self.sources[0]['sha256'], previous_source['sha256'])
        self.refresh_implementation_review()
        config = prep.load_config(self.config_path)
        receipt = {}
        self.assertEqual(controller.reviewed_inputs(config, self.sources, receipt), self.runtime)
        manifest = json.loads(self.source_path.read_bytes())['files']
        self.assertIn(self.sources[0], manifest)
        self.assertNotIn(previous_source, manifest)
        self.assertEqual(receipt['binding_amendment_review'], pin(self.review_path))
        self.assert_review_unchanged()

    def test_adopted_sha_mismatch_refuses_load_and_binding(self):
        self.stage()
        config = self.adopt_and_prepare_runtime()
        self.adopted.write_bytes(self.adopted.read_bytes()+b'\n')
        self.assertNotEqual(pin(self.adopted)['sha256'], pin(self.proposal)['sha256'])
        with self.assertRaises(ValueError):
            prep.load_config(self.config_path)
        with self.assertRaises((ValueError, InterlockFailure)):
            controller.reviewed_inputs(config, self.sources, {})
        self.assert_review_unchanged()

    def test_loaded_configuration_requires_exact_adopted_row(self):
        self.stage()
        config = self.adopt_and_prepare_runtime()
        adopted_row = config.amendments[0]
        cases = {
            'no adopted row': (),
            'proposal row only': (dict(adopted_row, **pin(self.proposal)),),
            'mismatched digest': (dict(adopted_row, sha256='0'*64),),
        }
        for name, amendments in cases.items():
            with self.subTest(name=name):
                config.amendments = amendments
                # Provenance validation may reject before the binding-specific gate.
                with self.assertRaises((ValueError, InterlockFailure)):
                    controller.reviewed_inputs(config, self.sources, {})
                self.assert_review_unchanged()

    def test_adoption_review_mismatched_pin_refuses_binding(self):
        self.stage()
        config = self.adopt_and_prepare_runtime()
        alternate = self.write_bytes(self.out/'reviews/alternate-amendment.json',
                                     self.preadoption_review)
        config.amendments = (dict(config.amendments[0], adoption_review=pin(alternate)),)
        with self.assertRaises(InterlockFailure):
            controller.reviewed_inputs(config, self.sources, {})
        self.assert_review_unchanged()


if __name__ == '__main__':
    unittest.main()

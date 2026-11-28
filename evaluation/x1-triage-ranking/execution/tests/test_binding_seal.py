"""Temporary-only v2 authority integration; substitute scanners, never binding gates."""
import copy
from datetime import datetime, timedelta, timezone
import hashlib
import json
from pathlib import Path
from types import SimpleNamespace
import unittest
from unittest.mock import patch
from contextlib import nullcontext

import test_binding_adoption as adoption
import test_binding as scanner_helpers
import acceptance as controller
import attest_closure
import fetch_objects
import preparation as prep
from interlock import evaluate
from verify_inputs import pin
from identifiers import cc2_delta


class BindingSealIntegrationTests(unittest.TestCase):
    # Compose only fixture operations, not the sibling TestCase's test methods.
    setUp = adoption.BindingAdoptionIntegrationTests.setUp
    write_json = adoption.BindingAdoptionIntegrationTests.write_json
    write_bytes = adoption.BindingAdoptionIntegrationTests.write_bytes
    review_record = adoption.BindingAdoptionIntegrationTests.review_record
    stage = adoption.BindingAdoptionIntegrationTests.stage
    adopt_and_prepare_runtime = adoption.BindingAdoptionIntegrationTests.adopt_and_prepare_runtime
    assert_review_unchanged = adoption.BindingAdoptionIntegrationTests.assert_review_unchanged

    # Core authority records live in the study's canonical review directories.
    # Reuse the exact scoped fixture format; never mock the authority validator.
    review = adoption.BindingAdoptionIntegrationTests.review

    def refresh_implementation_review(self, omit_source=None):
        reviewed_paths = [self.out/'closure.json']
        objects = self.out/'objects/django__django.json'
        if objects.exists():
            reviewed_paths.append(objects)
        adoption.BindingAdoptionIntegrationTests.refresh_implementation_review(
            self, omit_source=omit_source, reviewed_paths=reviewed_paths)

    def prepare(self):
        self.stage()
        self.clock_path = self.out/'clock.json'
        config = json.loads(self.config_path.read_bytes())
        config['paths'].update(ROOT=str(self.study), EXEC=str(self.execution))
        self.tip = 'a05e16a7230fe33d3f43f4a3aaed103cebb42efc'
        self.parent = 'synthetic-django-parent'
        config.update(required_files=[], historical_path_mapping={},
                      repositories=[{'repository': 'django/django', 'tip': self.tip, 'c0': self.parent}],
                      required_reviews=[{'path': 'reviews/implementation.json', 'scope': 'execution-implementation'},
                                        {'path': 'reviews/binding.json', 'scope': 'binding-acceptance'}])
        config['required_records'] += [
            {'path': 'closure.json', 'kind': 'scout.x1.installed-closure.v1'},
            {'path': 'objects/django__django.json', 'kind': 'scout.x1.objects.v1'}]
        self.config_path.write_text(json.dumps(config))
        proposal = json.loads(self.proposal.read_bytes())
        proposal['amends']['sha256'] = pin(self.config_path)['sha256']
        proposal['changes']['binding_acceptance_keys'] = list(controller.ACCEPTANCE)
        self.proposal.write_text(json.dumps(proposal))
        self.review_path.unlink()
        self.review(self.review_path, 'model-free-binding-v2-amendment',
                    [pin(self.proposal), pin(self.recipe_path), pin(self.amendment_path)])
        self.preadoption_review = self.review_path.read_bytes()
        self.preadoption_record = self.review_record(self.review_path).read_bytes()
        # Set the unit clock once before adoption. No production clock or time
        # validator is mocked; every later receipt must retain these same bytes.
        start = (datetime.now(timezone.utc)-timedelta(hours=1)).isoformat().replace('+00:00', 'Z')
        self.clock_path.write_text(json.dumps(prep.clock_value(config, self.config_path, 'a', start)))
        self.original_clock = self.clock_path.read_bytes()
        self.write_bytes(self.execution/'model_free/main.go', b'synthetic frontend source\n')
        self.stage_cc2_controls()
        patch.object(controller, 'HERE', self.execution/'binding').start()
        self.config = self.adopt_and_prepare_runtime()
        closure_path = self.out/'closure.json'
        closure = json.loads(closure_path.read_bytes())
        closure.update(schema='scout.x1.installed-closure.v1', clock=pin(self.clock_path),
                       retained_files=list(self.runtime['executables']))
        closure_path.write_text(json.dumps(closure))
        self.sources = self.current_sources()
        # Only these execution seams are replaced. Materialization verification,
        # policy/proof predicates, parity, authority, pins and journals stay real.
        patch.object(controller, 'generate', side_effect=self.generate).start()
        patch.object(controller, 'materialize', side_effect=self.materialize).start()
        patch.object(controller, 'run_revision', side_effect=self.run_revision).start()
        self.git = patch.object(controller, 'git', return_value=(self.parent+'\n').encode()).start()
        patch.object(prep, 'documentary_gate', return_value={'pins': [], 'deferred_to_b': []}).start()
        patch.object(attest_closure, 'verify_installed_closure').start()
        patch.object(fetch_objects, 'Git').start()
        patch.object(fetch_objects, 'inspect_store').start()
        patch.object(fetch_objects, 'first_parent_boundary',
                     return_value=(self.parent, 'synthetic', 1, 'tip-header', 'parent-header')).start()
        journals = {}
        for action in ('a', 'b', 'd'):
            identity = patch.object(prep.uuid, 'uuid4', return_value=SimpleNamespace(hex=cc2_delta.BASELINE_JOURNAL.removesuffix('.json'))) if action == 'd' else nullcontext()
            with identity:
                journal = prep.Journal(self.config, self.config_path, self.clock_path, action,
                                       environment={'HOME': str(self.study/'home')})
            if action == 'b':
                # Match attest_closure: bind the create-once closure to the
                # start of the operation that will finish successfully.
                closure['journal_start'] = pin(journal.start_path)
                closure_path.write_text(json.dumps(closure))
            journals[action] = journal.finish('passed', self.cc2_baseline_outputs if action == 'd' else ())
        self.objects_path = self.out/'objects/django__django.json'
        self.write_json(self.objects_path, {
            'schema': 'scout.x1.objects.v1', 'status': 'passed', 'clock': pin(self.clock_path),
            'repository': 'django/django', 'database_realpath': str(self.study/'objects.git'),
            'expected_tip': self.tip, 'actual_tip': self.tip, 'actual_c0': self.parent,
            'django_first_parent': {'sha': self.parent}, 'first_parent_membership': True,
            'no_shallow': True, 'no_promisor': True, 'no_alternates': True, 'no_replacements': True,
            'journal': journals['a'], 'git_identity': self.runtime['git'],
            'c0_derivation': 'synthetic', 'checked_edge_count': 1,
            'tip_header': 'tip-header', 'c0_header': 'parent-header'})
        self.refresh_implementation_review()

    def stage_cc2_controls(self):
        self.cc2_code = [pin(self.write_bytes(self.execution/name, ('synthetic '+name+'\n').encode())) for name in (
            'identifiers/cc2_delta.py', 'identifiers/cc2_readers.py', 'identifiers/cc2_schemas.py',
            'tests/test_cc2_delta.py', 'identifiers/projection.py', 'preparation.py', 'verify_inputs.py')]
        request = self.study/'preparation-reviews/identifiers-delta-cc2-REQUEST.md'
        self.cc2_code.append(pin(self.write_bytes(request, b'synthetic request\n')))
        self.cc2_source = pin(self.write_json(self.study/'preparation/exposure-sources.json',
                                            {'schema': 'scout.x1.exposure-source-manifest.v1'}))
        self.cc2_baseline_outputs = [pin(self.write_bytes(self.out/'identifiers'/name,
                                                        ('synthetic baseline '+name+'\n').encode()))
                                     for name in cc2_delta.BASELINE_NAMES]

    def publish_cc2(self):
        prior = [pin(self.out/'journal'/cc2_delta.BASELINE_JOURNAL), *self.cc2_baseline_outputs]
        self.cc2_implementation_review = self.study/'preparation-reviews/identifiers-delta-cc2-implementation.json'
        self.review(self.cc2_implementation_review, cc2_delta.REVIEW_SCOPE,
                    self.cc2_code + [self.cc2_source, pin(self.config_path)] + prior)
        common = {'status': 'passed', 'clock': pin(self.clock_path),
                  'implementation_review': pin(self.cc2_implementation_review),
                  'sources_manifest': self.cc2_source, 'producer_code': [], 'predecessors': prior,
                  'development_freshness_filter': False}
        journal = prep.Journal(self.config, self.config_path, self.clock_path, 'd',
                               source_pins=self.cc2_code + [self.cc2_source])
        self.cc2_outputs = cc2_delta.publish(self.out/'identifiers-delta-cc2', common, [], [], [])
        self.cc2_journal = journal.finish('passed', self.cc2_outputs)
        self.cc2_output_review = self.out/'reviews/identifier-delta-cc2-output.json'
        self.review(self.cc2_output_review, 'identifier-delta-cc2-output',
                    self.cc2_outputs + [self.cc2_journal])

    def current_sources(self):
        return [pin(path) for path in sorted((self.execution/'binding').rglob('*')) if path.is_file()] + [
            pin(self.execution/'model_free/main.go')]

    def generate(self, database, executable):
        database.mkdir(parents=True)
        return {'database': str(database), 'parent': 'synthetic-parent', 'head': 'synthetic-head'}

    def materialize(self, database, revision, root, executable):
        data = ('synthetic tracked policy '+revision+'\n').encode()
        path = self.write_bytes(root/'.scout/policy.yaml', data)
        path.chmod(0o644)
        return [{'root': str(root.resolve()), 'repo_path': '.scout/policy.yaml',
                 'revision': revision, 'mode': '100644', 'bytes': len(data),
                 'content_sha256': hashlib.sha256(data).hexdigest(), 'blob': 'synthetic-blob'}]

    def run_revision(self, root, revision, route, run, runtime, frozen, closure_path,
                     resolve_only=False, expected_policy=None):
        template = scanner_helpers.DeterministicProofTests().evidence()
        policy = copy.deepcopy(template['runs']['bound']['parent']['policy'])
        policy.update(route=route, revision=revision)
        policy['original_policy'] = {'model_cli': {'provider': 'claude' if revision.endswith('parent') else 'codex'}}
        run.mkdir(parents=True)
        effective = self.write_json(run/'effective.json', policy)
        if expected_policy is not None:
            controller.verify_effective_record(effective, expected_policy)
        result = None
        interlock = evaluate([])
        if not resolve_only:
            result = {**scanner_helpers.FixtureAndParityTests().result(),
                      **scanner_helpers.lane_result(route), 'route': route, 'revision': revision}
            occurrence = 'occurrence-'+revision
            result.update(issue_ids=[occurrence], multiplicity={occurrence: 1},
                          durable_links={occurrence: 'issue:quality:existing'})
            self.write_json(run/'result.json', result)
            if route == 'synthetic-unbound':
                provider = policy['original_policy']['model_cli']['provider']
                attempt = scanner_helpers.ModelAttemptTests().attempt(provider)
                self.write_bytes(run/'attempts.jsonl', (json.dumps(attempt)+'\n').encode())
                interlock = evaluate([attempt], positive=True, expected_provider=provider)
        return {'policy': policy, 'result': result, 'effective_record': pin(effective),
                'interlock': interlock, 'key_present': False,
                'pins': [pin(path) for path in sorted(run.iterdir()) if path.is_file()]}

    def rehearsal(self, name):
        report = self.study/'scratch'/name
        args = SimpleNamespace(fixture='synthetic', objects=None, rehearsal_report=None,
                               config=self.config_path, clock=self.clock_path, output=report)
        self.assertEqual(controller.rehearse(self.config, args, self.sources), 0)
        review = self.out/'reviews/binding-v2-rehearsal.json'
        review.unlink(missing_ok=True)
        self.review(review, 'model-free-binding-v2-rehearsal',
                    [pin(report), pin(self.runtime_path), *self.sources])
        authority = controller.verify_rehearsal(self.config, report, self.sources)
        self.assertEqual(set(authority), {'report', 'review', 'record'})
        self.assertEqual(authority['record'], pin(self.review_record(review)))
        return report, authority

    def canonical(self, report):
        self.publish_cc2()
        for fixture, name in [('synthetic', 'synthetic'), ('django-pinned-pair', 'django')]:
            argv = ['--config', str(self.config_path), '--clock', str(self.clock_path),
                    '--fixture', fixture, '--output', str(self.out/'binding'/(name+'.json'))]
            argv += (['--rehearsal-report', str(report)] if name == 'synthetic' else
                     ['--objects', str(self.objects_path)])
            self.assertEqual(controller.main(argv), 0)
        self.git.assert_called_with(self.runtime['git']['path'], str(self.study/'objects.git'),
                                    'rev-parse', '--verify', self.tip+'^1')
        self.review(self.out/'reviews/binding.json', 'binding-acceptance',
                    [pin(self.out/path) for path in ('binding/synthetic.json', 'binding/django.json',
                                                   'closure.json', 'objects/django__django.json')])

    def collect(self):
        return prep.collect_seal(prep.load_config(self.config_path), self.config_path, self.clock_path)

    def assert_sealed_bindings(self, seal, authority):
        self.assertEqual(seal['status'], 'sealed')
        outputs = {row['path']: row for row in seal['outputs']}
        for name in ('synthetic', 'django'):
            path = self.out/'binding'/(name+'.json')
            self.assertEqual(outputs[str(path)]['role'], 'scout.x1.binding-acceptance.v2')
            evidence = json.loads(path.read_bytes())
            self.assertEqual(evidence['acceptance'], dict.fromkeys(controller.ACCEPTANCE, True))
            self.assertEqual(evidence['runtime_closure'], pin(self.out/'closure.json'))
        synthetic = json.loads((self.out/'binding/synthetic.json').read_bytes())
        self.assertEqual(synthetic['reviewed_rehearsal'], authority)
        django = json.loads((self.out/'binding/django.json').read_bytes())
        self.assertEqual(django['revisions'], {'head': self.tip, 'parent': self.parent})
        self.assertEqual(django['synthetic_receipt'], pin(self.out/'binding/synthetic.json'))
        referenced = {row['path'] for row in seal['sources']+seal['outputs']}
        self.assertIn(authority['record']['path'], referenced)
        self.assertEqual(self.clock_path.read_bytes(), self.original_clock)
        self.assertLess(prep.parsed(seal['ended_at']),
                        prep.parsed(json.loads(self.original_clock)['gold_freeze_deadline']))
        self.assert_review_unchanged()

    def test_real_adopt_rehearse_canonical_django_and_v2_seal(self):
        self.prepare()
        report, authority = self.rehearsal('initial.json')
        self.canonical(report)
        self.assert_sealed_bindings(self.collect(), authority)
        # Real final review receipts must cover both v2 output roles, not only
        # claim a passing status. Removing either digest must make seal refuse.
        review_path = self.out/'reviews/binding.json'
        original_review = review_path.read_bytes()
        for name in ('synthetic', 'django'):
            with self.subTest(missing_reviewed=name):
                review = json.loads(original_review)
                review['reviewed'] = [row for row in review['reviewed']
                                      if row['path'] != str(self.out/'binding'/(name+'.json'))]
                review_path.write_text(json.dumps(review))
                with self.assertRaisesRegex(ValueError, 'scoped reviews do not bind all required evidence'):
                    self.collect()
                review_path.write_bytes(original_review)

    def test_post_rehearsal_source_fix_requires_fresh_review_before_reservation(self):
        self.prepare()
        old_report, old_authority = self.rehearsal('before-fix.json')
        source = self.execution/'binding/acceptance.py'
        source.write_bytes(b'synthetic corrected controller source\n')
        self.sources = self.current_sources()
        self.assertEqual(prep.load_config(self.config_path).amendments, self.config.amendments)
        with self.assertRaisesRegex(ValueError, 'pin mismatch'):
            controller.reviewed_inputs(self.config, self.sources, {})
        # This standalone synthetic record documents review order; code gates
        # on the council record pinned by implementation.json, regenerated by
        # refresh_implementation_review below, not on source-fix.json itself.
        self.review(self.study/'council/source-fix.json', 'binding-source-fix', self.sources)
        self.refresh_implementation_review()
        self.config = prep.load_config(self.config_path)
        self.assertEqual(controller.reviewed_inputs(self.config, self.sources, {}), self.runtime)
        with self.assertRaises((ValueError, controller.InterlockFailure)):
            controller.verify_rehearsal(self.config, old_report, self.sources)
        self.assertFalse((self.out/'binding/synthetic.json').exists())
        report, authority = self.rehearsal('after-fix.json')
        self.assertNotEqual(authority['report'], old_authority['report'])
        self.assertFalse((self.out/'binding/synthetic.json').exists())
        self.canonical(report)
        self.assert_sealed_bindings(self.collect(), authority)

    def test_rehearsal_record_mutation_invalidates_canonical_tree_and_seal(self):
        self.prepare()
        report, authority = self.rehearsal('reviewed.json')
        self.canonical(report)
        self.assert_sealed_bindings(self.collect(), authority)
        evidence = json.loads((self.out/'binding/synthetic.json').read_bytes())
        record = Path(authority['record']['path'])
        original = record.read_bytes()
        for mutation in ('changed', 'deleted'):
            with self.subTest(mutation=mutation):
                if mutation == 'changed':
                    record.write_bytes(original+b'Changed after canonical acceptance.\n')
                    expected = ValueError
                else:
                    record.unlink()
                    expected = FileNotFoundError
                # These checks inspect already-created canonical evidence; they
                # are deliberately not merely verify_rehearsal re-entry tests.
                with self.assertRaises(expected):
                    prep.verify_tree_pins(evidence)
                with self.assertRaises(expected):
                    self.collect()
                record.write_bytes(original)

    def test_rehearsal_review_omitting_current_runtime_or_source_refuses(self):
        self.prepare()
        report, _ = self.rehearsal('complete-current-report.json')
        review = self.out/'reviews/binding-v2-rehearsal.json'
        rows = [pin(report), pin(self.runtime_path), *self.sources]
        for missing in (pin(self.runtime_path), self.sources[0]):
            with self.subTest(missing=missing['path']):
                review.unlink()  # Reissue the temporary rehearsal review, not its report.
                self.review(review, 'model-free-binding-v2-rehearsal',
                            [row for row in rows if row != missing])
                with self.assertRaisesRegex(controller.InterlockFailure,
                                            'does not bind report, runtime and source pins'):
                    controller.verify_rehearsal(self.config, report, self.sources)


if __name__ == '__main__':
    unittest.main()

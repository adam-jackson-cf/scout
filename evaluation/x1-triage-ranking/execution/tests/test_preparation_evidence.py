"""Synthetic regressions for X1C-R1-01 through X1C-R1-07; no real actions."""
import copy
import json
from pathlib import Path
import sys
import tempfile
import unittest
from unittest.mock import patch
from datetime import datetime, timedelta, timezone

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
import preparation as p
import fetch_objects as fetch
from verify_inputs import pin, exclusive_json, SHARED
import test_core as core
import test_binding_seal as binding_fixture


class CoreReviewTests(unittest.TestCase):
    setUp = core.CoreTests.setUp
    amendment_fixture = core.CoreTests.amendment_fixture
    stage_amendment_review = core.CoreTests.stage_amendment_review
    approve_amendment = core.CoreTests.approve_amendment

    def authority(self, scope='wanted', text=None, record=None, rows=None):
        study = self.root/'study'
        receipt = study/'preparation-run/reviews/review.json'
        evidence = self.root/'evidence'
        if not evidence.exists():
            evidence.write_bytes(b'evidence')
        rows = rows or [pin(evidence)]
        record = record or study/'preparation-reviews/council.md'
        record.parent.mkdir(parents=True, exist_ok=True)
        if text is not None:
            record.write_text(text)
        elif not record.exists():
            record.write_text('**Zero issues. Signed off.**\nSigned-off scope: '+scope+'\n'+'\n'.join(row['sha256'] for row in rows)+'\n')
        exclusive_json(receipt, {'scope': scope, 'record': pin(record), 'reviewed': rows})
        return receipt, record, rows

    def test_r101_exact_approval_scope_and_digest(self):
        receipt, _, rows = self.authority()
        self.assertEqual(p.review_receipt(receipt, 'wanted')['reviewed'], rows)

    def test_r101_real_nonapproval_records_refused(self):
        study = Path(p.__file__).resolve().parent.parent
        proposal = study/'execution/run-config-amendments-proposed/001-binding-v2.json'
        for name in ('identifiers-impl-round-5.md', 'binding-v2-round-4.md'):
            with self.subTest(record=name):
                receipt = self.root/(name+'.json')
                exclusive_json(receipt, {'scope': 'model-free-binding-v2-amendment',
                                         'record': pin(study/'preparation-reviews'/name), 'reviewed': [pin(proposal)]})
                with self.assertRaisesRegex(ValueError, 'exact zero-issue verdict line|signed-off scope'):
                    p.review_receipt(receipt, 'model-free-binding-v2-amendment')

    def test_r101_negated_verdict_refused(self):
        receipt, _, _ = self.authority(text='Not zero issues. Not signed off.\nSigned-off scope: wanted\n')
        with self.assertRaisesRegex(ValueError, 'exact zero-issue verdict line'):
            p.review_receipt(receipt, 'wanted')

    def test_r101_other_scope_refused(self):
        receipt, _, _ = self.authority(text='**Zero issues. Signed off.**\nSigned-off scope: other\n')
        with self.assertRaisesRegex(ValueError, 'signed-off scope missing'):
            p.review_receipt(receipt, 'wanted')

    def test_r101_digest_in_other_scope_refused(self):
        evidence = self.root/'evidence'; evidence.write_bytes(b'evidence')
        receipt, _, _ = self.authority(text='**Zero issues. Signed off.**\nSigned-off scope: wanted\nnone\nSigned-off scope: other\n'+pin(evidence)['sha256']+'\n')
        with self.assertRaisesRegex(ValueError, 'scope block does not name reviewed digest'):
            p.review_receipt(receipt, 'wanted')

    def test_r101_authority_outside_study_review_dirs_refused(self):
        receipt, _, _ = self.authority(record=self.root/'outside.md')
        with self.assertRaisesRegex(ValueError, 'outside study review directories'):
            p.review_receipt(receipt, 'wanted')

    def test_r101_cross_scope_pooling_refused(self):
        out = self.root/'preparation-run'; out.mkdir()
        for name in ('manifest.json', 'closure.json', 'synthetic.json', 'django.json'):
            (out/name).write_text(name)
        (out/'binding').mkdir()
        for name in ('synthetic', 'django'):
            (out/'binding'/(name+'.json')).write_text(name)
        config = {'paths': {'OUT': str(out)}, 'source_manifest': 'manifest.json', 'repositories': [], 'required_records': [], 'required_files': []}
        all_hashes = {pin(path)['sha256'] for path in out.rglob('*') if path.is_file()}
        with self.assertRaisesRegex(ValueError, 'scoped reviews do not bind.*binding-acceptance'):
            p.verify_scoped_evidence(config, {'execution-implementation': all_hashes})

    def journal_fixture(self, actions='abcd'):
        study = self.root/'study'; out = study/'preparation-run'
        config = {'paths': {'ROOT': str(self.root), 'STUDY': str(study), 'OUT': str(out)}}
        config_path = study/'execution/run-config.json'
        exclusive_json(config_path, config)
        clock_path = out/'clock.json'
        clock = {'preparation_started_at': '2026-01-01T00:00:00Z'}
        exclusive_json(clock_path, clock)
        journals = []
        with patch.object(p, 'read_clock', return_value=clock):
            for action in actions:
                journal = p.Journal(config, config_path, clock_path, action)
                journal.finish('passed')
                journals.append(journal)
        return config, config_path, clock_path, clock, journals

    def collect(self, fixture):
        return p.collect_journals(*fixture[:4])

    def rewrite_terminal(self, journal, **changes):
        value = json.loads(journal.end_path.read_bytes()); value.update(changes)
        journal.end_path.write_text(json.dumps(value))
        return value

    def recovery(self, fixture, journal, rows=None):
        config, _, _, _, _ = fixture
        value = json.loads(journal.end_path.read_bytes())
        required = [pin(journal.start_path), pin(journal.end_path)] + p.verify_tree_pins(value['output_pins'])
        rows = required if rows is None else rows
        record = self.root/'study/preparation-reviews/recovery.md'
        record.parent.mkdir(parents=True, exist_ok=True)
        record.write_text('**Zero issues. Signed off.**\nSigned-off scope: operation-recovery\n'+'\n'.join(row['sha256'] for row in rows)+'\n')
        receipt = Path(config['paths']['OUT'])/'reviews'/('operation-recovery-'+journal.id+'.json')
        exclusive_json(receipt, {'scope': 'operation-recovery', 'record': pin(record), 'reviewed': rows})
        return receipt

    def test_r103_failed_unreviewed_operation_refused(self):
        fixture = self.journal_fixture()
        self.rewrite_terminal(fixture[4][-1], status='failed', exit_code=1)
        with self.assertRaisesRegex(ValueError, 'failed operation lacks operation-recovery review'):
            self.collect(fixture)

    def test_r103_reviewed_failure_retained_outside_operations(self):
        fixture = self.journal_fixture('abcdd')
        failed = fixture[4][-1]
        output = self.root/'study/preparation-run/failure.json'; output.write_text('retained failure')
        self.rewrite_terminal(failed, status='failed', exit_code=1, output_pins=[pin(output)])
        review = self.recovery(fixture, failed)
        operations, outputs, pins, deviations, passed = self.collect(fixture)
        self.assertNotIn(failed.id, {row['id'] for row in operations})
        self.assertIn(str(review), {row['path'] for row in deviations})
        self.assertIn(pin(output), pins)
        self.assertNotIn(str(failed.end_path), passed)

    def test_r103_recovery_must_cover_all_output_pins(self):
        fixture = self.journal_fixture('abcdd'); failed = fixture[4][-1]
        output = self.root/'study/preparation-run/failure.json'; output.write_text('retained failure')
        self.rewrite_terminal(failed, status='failed', exit_code=1, output_pins=[pin(output)])
        self.recovery(fixture, failed, [pin(failed.start_path), pin(failed.end_path)])
        with self.assertRaisesRegex(ValueError, 'operation-recovery review omits operation evidence'):
            self.collect(fixture)

    def test_r103_deleted_prior_pair_detected(self):
        fixture = self.journal_fixture()
        first = fixture[4][0]; first.start_path.unlink(); first.end_path.unlink()
        with self.assertRaisesRegex(FileNotFoundError, first.id):
            self.collect(fixture)

    def test_r103_duplicate_fetch_refuses_before_journal(self):
        study = self.root/'study'; out = study/'preparation-run'
        config = {'paths': {key: str(value) for key, value in {'OUT': out, 'CAND': self.root/'cand', 'SCRATCH': self.root/'scratch'}.items()}}
        receipt = out/'objects/example__repo.json'; exclusive_json(receipt, {})
        config_path = self.root/'config.json'; exclusive_json(config_path, config)
        repo = {'repository': 'example/repo', 'directory': 'repo.git'}
        with patch.object(fetch, 'read_clock'), patch.object(fetch, 'Journal') as journal:
            with self.assertRaisesRegex(ValueError, 'object receipt already exists'):
                fetch.fetch(config, config_path, out/'clock.json', repo)
            journal.assert_not_called()
        self.assertFalse((out/'journal').exists())

    def test_r104_shared_whole_file_observations_can_change(self):
        shared = self.root/'docs/STATUS.md'; shared.parent.mkdir(); shared.write_text('before')
        other = self.root/'immutable'; other.write_text('fixed')
        receipt = {'pins': [pin(shared), pin(other)]}
        shared.write_text('after sibling integration')
        p.verify_input_receipt(receipt, self.root)
        other.write_text('tampered')
        with self.assertRaisesRegex(ValueError, 'pin mismatch'):
            p.verify_input_receipt(receipt, self.root)

    def test_r105_seal_check_has_no_reservation_journal_or_writes(self):
        out = self.root/'out'; out.mkdir(); (out/'.operation.lock').write_bytes(b'')
        config = {'paths': {'OUT': str(out)}}
        before = {str(path): (path.read_bytes(), path.stat().st_mtime_ns) for path in out.iterdir()}
        with patch.object(p, 'collect_seal', return_value={'status': 'sealed'}) as collect, patch.object(p, 'Journal') as journal, patch.object(p, 'exclusive_json') as write:
            self.assertEqual(p.seal(config, self.root/'config', out/'clock.json', check=True), {'status': 'sealed'})
            journal.assert_not_called(); write.assert_not_called(); collect.assert_called_once()
        self.assertEqual(before, {str(path): (path.read_bytes(), path.stat().st_mtime_ns) for path in out.iterdir()})
        self.assertFalse((out/'preparation.seal.json').exists())

    def test_r106_new_journal_pins_entry_and_loaded_execution_modules(self):
        fixture = self.journal_fixture()
        first = json.loads(fixture[4][0].start_path.read_bytes())
        paths = {row['path'] for row in first['producer_pins']}
        self.assertIn(str(Path(p.__file__).resolve()), paths)
        self.assertIn(str(Path(fetch.__file__).resolve()), paths)
        self.assertTrue(first['producer_pins'])
        last = json.loads(fixture[4][-1].start_path.read_bytes())
        self.assertIn(str(fixture[4][0].end_path), {row['path'] for row in last['prior_journals']})

    def test_r106_legacy_producer_gap_disclosed(self):
        fixture = self.journal_fixture()
        last = fixture[4][-1]
        start = json.loads(last.start_path.read_bytes()); start.pop('producer_pins'); last.start_path.write_text(json.dumps(start))
        terminal = json.loads(last.end_path.read_bytes()); terminal.pop('producer_pins'); terminal['start_record'] = pin(last.start_path); last.end_path.write_text(json.dumps(terminal))
        deviations = self.collect(fixture)[3]
        self.assertIn(dict(pin(last.end_path), reason='producer bytes unrecorded'), deviations)

    def test_r107_unfinished_journal(self):
        fixture = self.journal_fixture(); fixture[4][-1].end_path.unlink()
        with self.assertRaisesRegex(ValueError, 'missing/nonterminal operation records'): self.collect(fixture)

    def test_r107_id_stem_mismatch(self):
        fixture = self.journal_fixture(); self.rewrite_terminal(fixture[4][-1], id='wrong')
        with self.assertRaisesRegex(ValueError, 'operation id/stem mismatch'): self.collect(fixture)

    def test_r107_stale_journal_clock(self):
        fixture = self.journal_fixture(); self.rewrite_terminal(fixture[4][-1], clock={})
        with self.assertRaisesRegex(ValueError, 'stale operation clock'): self.collect(fixture)

    def test_r107_journal_before_clock(self):
        fixture = self.journal_fixture(); self.rewrite_terminal(fixture[4][-1], started_at='2025-01-01T00:00:00Z')
        with self.assertRaisesRegex(ValueError, 'invalid operation timeline'): self.collect(fixture)

    def test_r107_terminal_before_start(self):
        fixture = self.journal_fixture(); self.rewrite_terminal(fixture[4][-1], ended_at='2025-01-01T00:00:00Z')
        with self.assertRaisesRegex(ValueError, 'invalid operation timeline'): self.collect(fixture)

    def test_r107_start_terminal_mismatch(self):
        fixture = self.journal_fixture(); self.rewrite_terminal(fixture[4][-1], argv=['different'])
        with self.assertRaisesRegex(ValueError, 'terminal/start operation record mismatch'): self.collect(fixture)

    def test_r107_missing_required_action(self):
        fixture = self.journal_fixture('abc')
        with self.assertRaisesRegex(ValueError, 'required action journals missing'): self.collect(fixture)

    def test_r107_noncanonical_start_record(self):
        fixture = self.journal_fixture(); self.rewrite_terminal(fixture[4][-1], start_record={'path': '/wrong'})
        with self.assertRaisesRegex(ValueError, 'noncanonical operation start path'): self.collect(fixture)

    def test_r107_adoption_existing_reservation(self):
        path, first, _ = self.amendment_fixture(); proposal, review = self.stage_amendment_review(path, first)
        receipt, _ = p.adoption_paths(json.loads(path.read_bytes()), first); exclusive_json(receipt, {})
        with patch.object(p, 'read_clock'), patch.object(sys, 'argv', [str(Path(p.__file__)), 'adopt']):
            with self.assertRaisesRegex(ValueError, 'adoption already reserved'): p.adopt(path, proposal, review)

    def test_r107_adoption_unfinished_operation(self):
        path, first, _ = self.amendment_fixture(); proposal, review = self.stage_amendment_review(path, first)
        exclusive_json(path.parents[1]/'preparation-run/journal/open.start.json', {'started_at': '2026-01-01T00:00:00Z'})
        with patch.object(p, 'read_clock'), patch.object(sys, 'argv', [str(Path(p.__file__)), 'adopt']):
            with self.assertRaisesRegex(ValueError, 'cannot adopt during a nonterminal operation'): p.adopt(path, proposal, review)

    def test_r107_adoption_twice(self):
        path, first, _ = self.amendment_fixture(); review = self.approve_amendment(path, first)
        proposal = path.parent/'run-config-amendments-proposed'/first.name
        with patch.object(p, 'read_clock'), patch.object(sys, 'argv', [str(Path(p.__file__)), 'adopt']):
            with self.assertRaisesRegex(ValueError, 'prior digest/path differs'): p.adopt(path, proposal, review)

    def test_r107_adoption_filename_order(self):
        path, first, original = self.amendment_fixture(); self.approve_amendment(path, first)
        proposal = path.parent/'run-config-amendments-proposed/000-earlier.json'
        amendment = dict(original, amends={'path': 'execution/run-config-amendments/'+first.name, 'sha256': pin(first)['sha256']}, authority={'path': 'preparation-run/reviews/earlier.json', 'scope': 'earlier'})
        exclusive_json(proposal, amendment)
        record = path.parents[1]/'preparation-reviews/earlier.md'; record.write_text('**Zero issues. Signed off.**\nSigned-off scope: earlier\n'+pin(proposal)['sha256'])
        review = path.parents[1]/'preparation-run/reviews/earlier.json'; exclusive_json(review, {'scope': 'earlier', 'record': pin(record), 'reviewed': [pin(proposal)]})
        with patch.object(p, 'read_clock'), patch.object(sys, 'argv', [str(Path(p.__file__)), 'adopt']):
            with self.assertRaisesRegex(ValueError, 'appended in filename order'): p.adopt(path, proposal, review)


    def test_r107_initial_seal_refusal_preserves_reservation_and_failure(self):
        config, config_path, clock_path, clock, journals = self.journal_fixture()
        with patch.object(p, 'read_clock', return_value=clock), patch.object(p, 'collect_seal', side_effect=ValueError('synthetic seal refusal')):
            with self.assertRaisesRegex(ValueError, 'synthetic seal refusal'):
                p.seal(config, config_path, clock_path)
        out = Path(config['paths']['OUT'])
        self.assertEqual((out/'preparation.seal.json').read_bytes(), b'')
        terminal = [json.loads(path.read_bytes()) for path in (out/'journal').glob('*.json') if not path.name.endswith('.start.json')]
        self.assertTrue(any(record['action'] == 'e' and record['status'] == 'refused' and record['failure_reason'] == 'synthetic seal refusal' for record in terminal))


class CoreSealReviewTests(unittest.TestCase):
    def ready(self):
        fixture = binding_fixture.BindingSealIntegrationTests('test_real_adopt_rehearse_canonical_django_and_v2_seal')
        fixture.setUp(); self.addCleanup(fixture.doCleanups)
        fixture.prepare(); report, authority = fixture.rehearsal('core-review.json'); fixture.canonical(report)
        return fixture

    def test_r102_predecessors_are_exactly_documentary_gate_pins(self):
        fixture = self.ready()
        signed = fixture.study/'signed'; signed.write_bytes(b'signed documentary bytes')
        gate = {'pins': [pin(signed)], 'deferred_to_b': []}
        with patch.object(p, 'documentary_gate', return_value=gate):
            result = fixture.collect()
        self.assertEqual(result['predecessors'], [dict(pin(signed), kind='signed-input')])
        self.assertEqual(gate['pins'], [pin(signed)])
        self.assertGreater(len(result['sources']), len(gate['pins']))

    def test_r107_integrated_unaccounted_output_refused(self):
        fixture = self.ready(); (fixture.out/'unaccounted').write_bytes(b'unaccounted')
        with self.assertRaisesRegex(ValueError, 'unaccounted run output'): fixture.collect()

    def test_r103_required_record_from_failed_operation_refused(self):
        fixture = self.ready()
        journal_paths = [path for path in (fixture.out/'journal').glob('*.json') if not path.name.endswith('.start.json')]
        # Choose the final operation: no later prior-journal pin can mask this check.
        terminal = max(journal_paths, key=lambda path: json.loads(path.read_bytes())['started_at'])
        value = json.loads(terminal.read_bytes()); value.update(status='failed', exit_code=1); terminal.write_text(json.dumps(value))
        start = Path(value['start_record']['path'])
        rows = [pin(start), pin(terminal)] + p.verify_tree_pins(value['output_pins'])
        fixture.review(fixture.out/'reviews'/('operation-recovery-'+value['id']+'.json'), 'operation-recovery', rows)
        with self.assertRaisesRegex(ValueError, 'required record lacks passed producer operation'): fixture.collect()

    def test_r107_clock_create_once_and_refusals(self):
        fixture = self.ready()
        with patch.object(p, 'documentary_gate', return_value={'deferred_to_b': []}), patch.object(p.subprocess, 'check_output', return_value=fixture.config['execution_commit']+'\n'):
            with self.assertRaisesRegex(FileExistsError, 'clock.json'): p.start(fixture.config, fixture.config_path, 'a')
        original = fixture.clock_path.read_bytes()
        mutations = [('schema', 'wrong', 'wrong clock schema'), ('study', 'wrong', 'wrong clock schema'), ('execution_commit', 'wrong', 'execution commit mismatch'), ('first_action', 'wrong', 'invalid first action'), ('gold_freeze_deadline', '2026-01-01T00:00:00Z', 'deadline mismatch'), ('deferred_submodule_gitlink', 'wrong', 'gitlink mismatch')]
        for key, value, message in mutations:
            with self.subTest(field=key):
                clock = json.loads(original); clock[key] = value; fixture.clock_path.write_text(json.dumps(clock))
                with self.assertRaisesRegex(ValueError, message): p.read_clock(fixture.config, fixture.config_path, fixture.clock_path)
        fixture.clock_path.write_bytes(original)
        with self.assertRaisesRegex(ValueError, 'noncanonical clock'): p.read_clock(fixture.config, fixture.config_path, fixture.out/'different.json')
        for field in ('authority', 'manifest', 'execution_config'):
            with self.subTest(pin=field):
                clock = json.loads(original); clock[field]['sha256'] = '0'*64
                fixture.clock_path.write_text(json.dumps(clock))
                with self.assertRaisesRegex(ValueError, 'clock '+field+' mismatch'):
                    p.read_clock(fixture.config, fixture.config_path, fixture.clock_path)
        clock = json.loads(original)
        future = datetime.now(timezone.utc)+timedelta(days=1)
        clock['preparation_started_at'] = future.isoformat()
        clock['gold_freeze_deadline'] = (future+timedelta(days=45)).isoformat()
        fixture.clock_path.write_text(json.dumps(clock))
        with self.assertRaisesRegex(ValueError, 'clock is in the future'):
            p.read_clock(fixture.config, fixture.config_path, fixture.clock_path)
        clock = json.loads(original); clock['input_verification']['path'] = str(fixture.out/'other.json')
        fixture.clock_path.write_text(json.dumps(clock))
        with self.assertRaisesRegex(ValueError, 'clock input receipt path mismatch'):
            p.read_clock(fixture.config, fixture.config_path, fixture.clock_path)
        clock = json.loads(original); clock['input_verification']['sha256'] = '0'*64
        fixture.clock_path.write_text(json.dumps(clock))
        with self.assertRaisesRegex(ValueError, 'pin mismatch'):
            p.read_clock(fixture.config, fixture.config_path, fixture.clock_path)
        fixture.clock_path.write_bytes(original)


if __name__ == '__main__':
    unittest.main()

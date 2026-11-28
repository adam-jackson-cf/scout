"""Synthetic tests only: local bare Git databases, no network or real actions."""
import importlib.util
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch
from concurrent.futures import ThreadPoolExecutor
from threading import Event

EXEC = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(EXEC))
import fetch_objects as fetch
import preparation as prep
import verify_inputs as inputs


class CoreTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.root = Path(self.temp.name).resolve()
        self.addCleanup(self.temp.cleanup)

    def test_exclusive_creation_preserves_original(self):
        path = self.root/'receipt.json'
        inputs.exclusive_json(path, {'first': 1})
        with self.assertRaises(FileExistsError):
            inputs.exclusive_json(path, {'second': 2})
        self.assertEqual(json.loads(path.read_bytes()), {'first': 1})

    def test_deadline_exact_seconds_and_leap_boundary(self):
        study = self.root/'study'
        (study/'reviews').mkdir(parents=True)
        (study/'reviews/round-9.md').write_bytes(b'authority')
        (study/'reviews/rev9-digests.json').write_bytes(b'manifest')
        (self.root/'input-verification.json').write_text('{"status":"passed"}')
        config = {'paths': {'STUDY': str(study), 'OUT': str(self.root)}, 'execution_commit': 'fixed', 'deferred_to_b_contract': {'gitlink': 'synthetic-gitlink'}}
        path = self.root/'config.json'
        path.write_text('{}')
        clock = prep.clock_value(config, path, 'a', '2028-02-01T12:34:56Z')
        self.assertEqual(clock['gold_freeze_deadline'], '2028-03-17T12:34:56Z')
        self.assertEqual((prep.parsed(clock['gold_freeze_deadline'])-prep.parsed(clock['preparation_started_at'])).total_seconds(), 45*86400)

    def test_exception_classification(self):
        original = b'| R-081 | owned |\n'
        sha = inputs.digest(original)
        owned = {'R-081': {'source': 'docs/research-ledger.md', 'selector': 'one row', 'sha256': sha}}
        self.assertEqual(inputs.classify('docs/research-ledger.md', original, sha, owned), 'exact')
        self.assertEqual(inputs.classify('docs/research-ledger.md', original+b'sibling\n', sha, owned), 'shared-file-exception')
        with self.assertRaises(ValueError):
            inputs.classify('docs/research-ledger.md', original.replace(b'owned', b'changed'), sha, owned)
        with self.assertRaises(ValueError):
            inputs.classify('signed.py', b'changed', sha, owned)

    def test_hd9_only_state_date_normalized(self):
        row = b'| HD9 | question | R-081 | proposed | X1 | result | 2026-11-08 |\n'
        self.assertEqual(inputs.normalized_hd9(row.replace(b'proposed', b'preregistered')), row)
        with self.assertRaises(ValueError):
            inputs.normalized_hd9(row.replace(b'proposed', b'tested'))
        self.assertNotEqual(inputs.normalized_hd9(row.replace(b'question', b'changed')), row)

    def test_recursive_pin_mismatch_and_missing(self):
        path = self.root/'record'
        path.write_bytes(b'first')
        value = {'nested': [inputs.pin(path)]}
        prep.verify_tree_pins(value)
        path.write_bytes(b'changed')
        with self.assertRaises(ValueError):
            prep.verify_tree_pins(value)
        path.unlink()
        with self.assertRaises(FileNotFoundError):
            prep.verify_tree_pins(value)

    def test_journal_pinned_output_is_accounted_and_sealed(self):
        out = self.root/'out'
        out.mkdir()
        canonical = out/'canonical.json'
        canonical.write_bytes(b'canonical')
        extra = out/'extra.json'
        extra.write_bytes(b'journal output')
        terminal = {'status': 'passed', 'output_pins': [inputs.pin(extra)]}
        sealed = prep.verify_output_accounting(out, [inputs.pin(canonical)], terminal['output_pins'])
        self.assertEqual(sealed, [dict(inputs.pin(extra), role='journal-output')])

    def test_canonical_output_role_preserved_when_journal_also_pins_it(self):
        out = self.root/'out'
        out.mkdir()
        receipt = out/'canonical.json'
        receipt.write_bytes(b'canonical receipt')
        canonical = inputs.pin(receipt, 'scout.x1.binding-acceptance.v2')
        outputs = [canonical]
        outputs.extend(prep.verify_output_accounting(out, outputs, [inputs.pin(receipt)]))
        self.assertEqual(outputs, [canonical])
        self.assertEqual(outputs[0]['role'], 'scout.x1.binding-acceptance.v2')

    def test_duplicate_journal_outputs_are_deduplicated_but_all_verified(self):
        out = self.root/'out'
        out.mkdir()
        receipt = out/'journal-only.json'
        receipt.write_bytes(b'journal output')
        valid = inputs.pin(receipt)
        outputs = prep.verify_output_accounting(out, [], [valid, valid])
        self.assertEqual(outputs, [dict(valid, role='journal-output')])
        with self.assertRaisesRegex(ValueError, 'pin mismatch'):
            prep.verify_output_accounting(out, [valid], [dict(valid, sha256='0'*64)])

    def test_unpinned_extra_output_is_refused(self):
        out = self.root/'out'
        out.mkdir()
        extra = out/'extra.json'
        extra.write_bytes(b'unpinned')
        with self.assertRaisesRegex(ValueError, 'unaccounted run output'):
            prep.verify_output_accounting(out, [], [])

    def test_tampered_journal_pinned_output_is_refused(self):
        out = self.root/'out'
        out.mkdir()
        extra = out/'extra.json'
        extra.write_bytes(b'original')
        terminal = {'status': 'passed', 'output_pins': [inputs.pin(extra)]}
        extra.write_bytes(b'tampered')
        with self.assertRaisesRegex(ValueError, 'pin mismatch'):
            prep.verify_output_accounting(out, [], terminal['output_pins'])

    def test_seal_missing_records_reservation_not_replaced(self):
        out = self.root/'out'
        config = {'paths': {'OUT': str(out)}}
        # A pre-existing crash reservation always refuses, even without valid inputs.
        out.mkdir()
        seal = out/'preparation.seal.json'
        seal.write_bytes(b'')
        with self.assertRaises(ValueError):
            prep.seal(config, self.root/'config', out/'clock.json')
        self.assertEqual(seal.read_bytes(), b'')

    def test_seal_missing_and_mismatched_canonical_record(self):
        out = self.root/'out'
        out.mkdir()
        inputs.exclusive_json(out/'input-verification.json', {'status': 'passed', 'pins': []})
        config = {'paths': {'OUT': str(out), 'ROOT': str(self.root)}, 'required_records': [{'path': 'closure.json', 'kind': 'expected-schema'}]}
        with patch.object(prep, 'read_clock', return_value={}), patch.object(prep, 'documentary_gate', return_value={'pins': []}):
            with self.assertRaisesRegex(FileNotFoundError, 'closure.json'):
                prep.collect_seal(config, self.root/'config', out/'clock.json')
            inputs.exclusive_json(out/'closure.json', {'schema': 'wrong-schema', 'status': 'passed'})
            with self.assertRaisesRegex(ValueError, 'missing/passing record schema'):
                prep.collect_seal(config, self.root/'config', out/'clock.json')

    def test_seal_deferred_rows_must_all_be_verified(self):
        row = {'path': 'module/source', 'sha256': inputs.digest(b'pinned'), 'bytes': 6}
        config = {'paths': {'ROOT': str(self.root)}}
        with self.assertRaises(ValueError):
            prep.verify_deferred_rows(config, {'deferred_to_b': [row]}, {})
        (self.root/'module').mkdir()
        (self.root/'module/source').write_bytes(b'pinned')
        observed = inputs.pin(self.root/'module/source')
        prep.verify_deferred_rows(config, {'deferred_to_b': [row]}, {'deferred_working_files_verified': [observed]})
        (self.root/'module/source').write_bytes(b'changed')
        with self.assertRaises(ValueError):
            prep.verify_deferred_rows(config, {'deferred_to_b': [row]}, {'deferred_working_files_verified': [observed]})

    def submodule_fixture(self, initialized=False):
        git, db, _ = self.git_fixture()
        commit = self.commit(git, db)
        subprocess.check_call([git.executable, '--git-dir='+str(db), 'update-index', '--add', '--cacheinfo', '160000,'+commit+',module'], env=git.env)
        tree = git.text(db, 'write-tree')
        env = dict(git.env, GIT_AUTHOR_NAME='Synthetic', GIT_AUTHOR_EMAIL='x@example.invalid', GIT_COMMITTER_NAME='Synthetic', GIT_COMMITTER_EMAIL='x@example.invalid')
        head = subprocess.check_output([git.executable, '--git-dir='+str(db), 'commit-tree', tree], env=env, input=b'submodule fixture\n').decode().strip()
        git.run(db, 'update-ref', 'HEAD', head)
        directory = self.root/'module'
        directory.mkdir()
        if initialized:
            (directory/'.git').write_text('gitdir: synthetic\n')
        row = {'path': 'module/source', 'sha256': inputs.digest(b'pinned'), 'bytes': 6}
        return db, directory, {'path': 'module', 'gitlink': commit, 'working_files': [row]}

    def test_start_binds_uninitialized_submodule_receipt_without_hashing_deferred_rows(self):
        db, _, spec = self.submodule_fixture()
        _, deferred = inputs.submodule_rows(db, spec)
        out = db/'out'
        out.mkdir()
        study = db/'study'
        (study/'reviews').mkdir(parents=True)
        (study/'reviews/round-9.md').write_bytes(b'authority')
        (study/'reviews/rev9-digests.json').write_bytes(b'manifest')
        receipt = {'status': 'passed', 'pins': [inputs.pin(study/'reviews/rev9-digests.json')], 'deferred_to_b': deferred}
        inputs.exclusive_json(out/'input-verification.json', receipt)
        head = subprocess.check_output([shutil.which('git'), '-C', str(db), 'rev-parse', 'HEAD'], text=True).strip()
        config = {'paths': {'ROOT': str(db), 'STUDY': str(study), 'OUT': str(out)}, 'execution_commit': head, 'deferred_to_b_contract': {'gitlink': spec['gitlink']}}
        config_path = db/'run-config.json'
        config_path.write_text(json.dumps(config))
        # Exercise start validation, but intercept its final create: no clock/action runs.
        with patch.object(prep, 'documentary_gate', return_value=receipt), patch.object(prep, 'exclusive_json') as create:
            prep.start(config, config_path, 'a')
        create.assert_called_once()
        destination, value = create.call_args.args
        self.assertEqual(destination, out/'clock.json')
        self.assertEqual(value['input_verification'], inputs.pin(out/'input-verification.json'))
        self.assertEqual(value['deferred_submodule_gitlink'], spec['gitlink'])
        self.assertFalse((db/'module/source').exists())
        self.assertFalse((out/'clock.json').exists())
        # Ordinary missing pins still refuse; only the explicit deferral list is skipped.
        with self.assertRaises(FileNotFoundError):
            prep.verify_input_receipt({'pins': deferred})

    def amendment_fixture(self):
        study = self.root/'study'
        execution = study/'execution'
        execution.mkdir(parents=True)
        path = execution/'run-config.json'
        config = {'schema': 'scout.x1.run-config.v1', 'paths': {'STUDY': str(study), 'OUT': str(study/'preparation-run')},
                  'required_records': [{'path': 'binding/synthetic.json', 'kind': 'scout.x1.binding-acceptance.v1'}, {'path': 'binding/django.json', 'kind': 'scout.x1.binding-acceptance.v1'}],
                  'producer_contracts': {'binding': 'original'}}
        inputs.exclusive_json(path, config)
        directory = execution/'run-config-amendments'
        directory.mkdir()
        amendment = {'schema': 'scout.x1.run-config-amendment.v1',
                     'amends': {'path': 'execution/run-config.json', 'sha256': inputs.pin(path)['sha256']},
                     'reason': 'synthetic', 'authority': {'path': 'preparation-run/reviews/amendment.json', 'scope': 'synthetic-binding-amendment'},
                     'changes': {'binding_record_kind': 'scout.x1.binding-acceptance.v2', 'binding_producer_contract': 'first'}}
        first = directory/'001-first.json'
        inputs.exclusive_json(first, amendment)
        return path, first, amendment

    def stage_amendment_review(self, config_path, amendment_path):
        proposal_directory = config_path.parent/'run-config-amendments-proposed'
        proposal_directory.mkdir(exist_ok=True)
        proposal = proposal_directory/amendment_path.name
        amendment_path.rename(proposal)
        amendment = json.loads(proposal.read_bytes())
        council = config_path.parents[1]/'preparation-reviews'/(proposal.stem+'.md')
        council.parent.mkdir(exist_ok=True)
        council.write_text('**Zero issues. Signed off.**\nSigned-off scope: '+amendment['authority']['scope']+'\n'+inputs.pin(proposal)['sha256']+'\n')
        receipt_path = config_path.parents[1]/amendment['authority']['path']
        inputs.exclusive_json(receipt_path, {'scope': amendment['authority']['scope'], 'record': inputs.pin(council),
                                            'reviewed': [inputs.pin(proposal)]})
        clock_path = config_path.parents[1]/'preparation-run/clock.json'
        if not clock_path.exists():
            inputs.exclusive_json(clock_path, {'execution_config': inputs.pin(config_path)})
        return proposal, receipt_path

    def approve_amendment(self, config_path, amendment_path, adopted_at='2026-01-02T00:00:00Z'):
        proposal, receipt_path = self.stage_amendment_review(config_path, amendment_path)
        with patch.object(prep, 'read_clock', return_value={}), patch.object(prep, 'now', return_value=adopted_at), patch.object(sys, 'argv', [str(EXEC/'preparation.py'), 'adopt']):
            prep.adopt(config_path, proposal, receipt_path)
        return receipt_path

    def test_original_config_is_unchanged(self):
        self.assertEqual(inputs.pin(EXEC/'run-config.json')['sha256'], 'c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54')

    def test_amendment_chain_applies_in_order_and_preserves_original(self):
        path, first, amendment = self.amendment_fixture()
        self.approve_amendment(path, first)
        original = path.read_bytes()
        second = dict(amendment, amends={'path': 'execution/run-config-amendments/001-first.json', 'sha256': inputs.pin(first)['sha256']},
                      authority={'path': 'preparation-run/reviews/amendment-2.json', 'scope': 'synthetic-binding-amendment'},
                      changes={'binding_producer_contract': 'second'})
        inputs.exclusive_json(first.parent/'002-second.json', second)
        self.approve_amendment(path, first.parent/'002-second.json', '2026-01-03T00:00:00Z')
        effective = prep.load_config(path)
        self.assertEqual(effective['producer_contracts']['binding'], 'second')
        self.assertTrue(all(row['kind'] == 'scout.x1.binding-acceptance.v2' for row in effective['required_records']))
        self.assertEqual(len(effective.amendments), 2)
        self.assertEqual(effective.amendments[0]['effective_sha256'], effective.amendments[1]['prior_effective_sha256'])
        self.assertEqual(prep.effective_configuration(effective)['sha256'], prep.config_digest(effective))
        self.assertEqual(path.read_bytes(), original)

    def test_wrong_prior_digest_refused(self):
        path, first, amendment = self.amendment_fixture()
        self.approve_amendment(path, first)
        second = dict(amendment, amends={'path': 'execution/run-config-amendments/001-first.json', 'sha256': '0'*64})
        inputs.exclusive_json(first.parent/'002-second.json', second)
        with self.assertRaisesRegex(ValueError, 'prior digest'):
            prep.load_config(path)

    def test_clock_pinned_original_config_cannot_drift(self):
        path, _, _ = self.amendment_fixture()
        out = path.parents[1]/'preparation-run'
        inputs.exclusive_json(out/'clock.json', {'execution_config': inputs.pin(path)})
        path.write_bytes(path.read_bytes()+b' ')
        with self.assertRaisesRegex(ValueError, 'pin mismatch'):
            prep.load_config(path)

    def test_unreviewed_amendment_refused_at_seal(self):
        path, first, _ = self.amendment_fixture()
        review = self.approve_amendment(path, first)
        effective = prep.load_config(path)
        review.unlink()
        with patch.object(prep, 'read_clock', return_value={}):
            with self.assertRaises(FileNotFoundError):
                prep.collect_seal(effective, path, path.parents[1]/'preparation-run/clock.json')

    def test_amendment_review_scope_and_exact_digest_required(self):
        path, first, amendment = self.amendment_fixture()
        self.approve_amendment(path, first)
        effective = prep.load_config(path)
        review_path = path.parents[1]/'preparation-run/reviews/amendment.json'
        value = json.loads(review_path.read_bytes())
        value['scope'] = 'wrong-scope'
        review_path.write_text(json.dumps(value))
        with self.assertRaisesRegex(ValueError, 'invalid scoped review receipt'):
            prep.review_receipt(review_path, amendment['authority']['scope'])

    def test_proposed_amendment_is_ignored(self):
        path, first, _ = self.amendment_fixture()
        proposed = path.parent/'run-config-amendments-proposed'
        proposed.mkdir()
        first.rename(proposed/first.name)
        effective = prep.load_config(path)
        self.assertEqual(effective.amendments, ())
        self.assertEqual(effective['producer_contracts']['binding'], 'original')
        self.assertEqual(prep.effective_configuration(effective)['adoption_state'],
                         {'condition': 'no-adopted-amendments', 'directory_contents': []})

    def test_adopted_amendment_without_review_refused_at_load(self):
        path, _, _ = self.amendment_fixture()
        with self.assertRaises(FileNotFoundError):
            prep.load_config(path)

    def test_seal_journal_validator_accepts_pre_and_post_adoption(self):
        path, first, _ = self.amendment_fixture()
        base = prep.RunConfig(json.loads(path.read_bytes()))
        pre_adoption = {'started_at': '2026-01-01T00:00:00Z', 'effective_configuration': prep.effective_configuration(base)}
        self.approve_amendment(path, first)
        effective = prep.load_config(path)
        post_adoption = {'started_at': '2026-01-03T00:00:00Z', 'effective_configuration': prep.effective_configuration(effective)}
        for record in (pre_adoption, post_adoption, {'started_at': '2026-01-01T00:00:00Z'}):
            prep.verify_journal_configuration(record, effective, path)
        with self.assertRaisesRegex(ValueError, 'adoption ordering'):
            prep.verify_journal_configuration(dict(pre_adoption, started_at='2026-01-03T00:00:00Z'), effective, path)
        with self.assertRaisesRegex(ValueError, 'post-adoption'):
            prep.verify_journal_configuration({'started_at': '2026-01-03T00:00:00Z'}, effective, path)
        changed = json.loads(json.dumps(post_adoption))
        changed['effective_configuration']['sha256'] = '0'*64
        with self.assertRaisesRegex(ValueError, 'digest differs'):
            prep.verify_journal_configuration(changed, effective, path)

    def test_edit_delete_or_rename_adopted_amendment_refused(self):
        path, first, _ = self.amendment_fixture()
        self.approve_amendment(path, first)
        effective = prep.load_config(path)
        first.write_bytes(first.read_bytes()+b' ')
        with self.assertRaises(ValueError):
            prep.verify_amendment_reviews(effective)
        first.unlink()
        with self.assertRaises(FileNotFoundError):
            prep.verify_amendment_reviews(effective)

    def test_hand_written_and_backdated_adoption_receipts_refused(self):
        path, first, _ = self.amendment_fixture()
        self.approve_amendment(path, first)
        config = prep.load_config(path)
        receipt_path, command_path = prep.adoption_paths(config, first)
        original = json.loads(receipt_path.read_bytes())
        changed = dict(original, adopted_at='2026-01-01T00:00:00Z')
        receipt_path.write_text(json.dumps(changed))
        with self.assertRaisesRegex(ValueError, 'timestamp differs'):
            prep.load_config(path)
        receipt_path.write_text(json.dumps(original))
        command_path.unlink()
        with self.assertRaises(FileNotFoundError):
            prep.load_config(path)

    def test_adoption_racing_journal_is_serialized(self):
        path, first, _ = self.amendment_fixture()
        proposal, review = self.stage_amendment_review(path, first)
        config = prep.load_config(path)
        clock = path.parents[1]/'preparation-run/clock.json'
        attempted, completed = Event(), Event()
        def adopting():
            attempted.set()
            result = prep.adopt(path, proposal, review)
            completed.set()
            return result
        with patch.object(prep, 'read_clock', return_value={'preparation_started_at': '2026-01-01T00:00:00Z'}), patch.object(sys, 'argv', [str(EXEC/'preparation.py'), 'adopt']):
            with ThreadPoolExecutor(max_workers=1) as pool:
                with prep.operation_lock(config):
                    journal = prep.Journal(config, path, clock, 'd')
                    future = pool.submit(adopting)
                    self.assertTrue(attempted.wait(5))
                    self.assertFalse(completed.wait(0.05))
                    journal.finish('passed')
                    self.assertFalse(completed.is_set())
                future.result(timeout=10)
        effective = prep.load_config(path)
        self.assertLess(prep.parsed(journal.started_at), prep.parsed(effective.amendments[0]['adopted_at']))
        prep.verify_journal_configuration(json.loads(journal.end_path.read_bytes()), effective, path)

    def test_load_and_seal_accept_legacy_ab_and_d_before_adoption(self):
        path, first, _ = self.amendment_fixture()
        proposal, review = self.stage_amendment_review(path, first)
        config = prep.load_config(path)
        out = path.parents[1]/'preparation-run'
        base = prep.effective_configuration(config)
        for operation, action in [('a-id', 'a'), ('b-id', 'b'), ('d-id', 'd')]:
            start = {'id': operation, 'action': action, 'started_at': '2026-01-01T00:00:00Z'}
            if action == 'd':
                start['effective_configuration'] = base
            inputs.exclusive_json(out/'journal'/(operation+'.start.json'), start)
            inputs.exclusive_json(out/'journal'/(operation+'.json'), dict(start, ended_at='2026-01-01T01:00:00Z'))
        with patch.object(prep, 'read_clock', return_value={}), patch.object(prep, 'now', return_value='2026-01-02T00:00:00Z'), patch.object(sys, 'argv', [str(EXEC/'preparation.py'), 'adopt']):
            prep.adopt(path, proposal, review)
        effective = prep.load_config(path)
        for record in (out/'journal').glob('*.json'):
            prep.verify_journal_configuration(json.loads(record.read_bytes()), effective, path)

    def test_uninitialized_submodule_deferred_not_passed(self):
        db, _, spec = self.submodule_fixture()
        checked, deferred = inputs.submodule_rows(db, spec)
        self.assertEqual(checked, [])
        self.assertEqual(len(deferred), 1)
        self.assertEqual(deferred[0]['classification'], 'deferred_to_b')

    def test_initialized_submodule_checks_working_bytes(self):
        db, _, spec = self.submodule_fixture()
        directory = db/'module'
        directory.mkdir()
        (directory/'.git').write_text('gitdir: synthetic\n')
        (directory/'source').write_bytes(b'pinned')
        checked, deferred = inputs.submodule_rows(db, spec)
        self.assertEqual(deferred, [])
        self.assertEqual(inputs.working_pin(db, checked[0])['sha256'], spec['working_files'][0]['sha256'])

    def test_initialized_submodule_mismatch_refused(self):
        db, _, spec = self.submodule_fixture()
        directory = db/'module'
        directory.mkdir()
        (directory/'.git').write_text('gitdir: synthetic\n')
        (directory/'source').write_bytes(b'changed')
        checked, _ = inputs.submodule_rows(db, spec)
        with self.assertRaises(ValueError):
            inputs.working_pin(db, checked[0])

    def git_fixture(self):
        home = self.root/'home'
        home.mkdir()
        git = fetch.Git(shutil.which('git'), home)
        db = self.root/'fixture.git'
        git.run(None, 'init', '--bare', '--template=', str(db))
        repo = {'url': 'https://github.com/example/fixture', 'tip': '0'*40, 'c0': None}
        git.run(db, 'remote', 'add', 'origin', repo['url'])
        return git, db, repo

    def commit(self, git, db, parents=()):
        # Git plumbing synthesises local commit objects without a checkout.
        tree = git.run(db, 'mktree', stdin=b'').stdout.decode().strip()
        env = dict(git.env, GIT_AUTHOR_NAME='Synthetic', GIT_AUTHOR_EMAIL='synthetic@example.invalid',
                   GIT_COMMITTER_NAME='Synthetic', GIT_COMMITTER_EMAIL='synthetic@example.invalid',
                   GIT_AUTHOR_DATE='2000-01-01T00:00:00Z', GIT_COMMITTER_DATE='2000-01-01T00:00:00Z')
        argv = [git.executable, '--git-dir='+str(db), 'commit-tree', tree]
        for parent in parents:
            argv += ['-p', parent]
        return subprocess.check_output(argv, env=env, input=b'synthetic\n').decode().strip()

    def test_merge_first_parent_and_offpath_refusal(self):
        git, db, repo = self.git_fixture()
        root = self.commit(git, db)
        left = self.commit(git, db, [root])
        # Distinguish the side branch with its own child.
        right = self.commit(git, db, [left])
        merge = self.commit(git, db, [left, right])
        actual, derivation, edges, tip_header, _ = fetch.first_parent_boundary(git, db, merge, root)
        self.assertEqual((actual, derivation, edges), (root, 'fixed', 2))
        self.assertEqual(tip_header['parents'], [left, right])
        with self.assertRaisesRegex(ValueError, 'first-parent'):
            fetch.first_parent_boundary(git, db, merge, right)
        actual, derivation, edges, _, _ = fetch.first_parent_boundary(git, db, merge)
        self.assertEqual((actual, derivation, edges), (root, 'root', 2))
        actual, derivation, edges, _, _ = fetch.first_parent_boundary(git, db, merge, dev_edges=1)
        self.assertEqual((actual, derivation, edges), (left, '2000-first-parent-edges', 1))

    def test_unsafe_existing_repository_refused(self):
        git, db, repo = self.git_fixture()
        self.assertTrue(fetch.inspect_store(git, db, repo)['no_shallow'])
        for name in ('shallow', 'objects/info/alternates', 'objects/info/http-alternates', 'info/grafts'):
            path = db/name
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(b'bad\n')
            with self.assertRaises(ValueError):
                fetch.inspect_store(git, db, repo)
            path.unlink()
        git.run(db, 'remote', 'set-url', 'origin', 'https://github.com/unexpected/repo')
        with self.assertRaises(ValueError):
            fetch.inspect_store(git, db, repo)

    def test_promisor_and_extra_ref_refused(self):
        git, db, repo = self.git_fixture()
        promisor = db/'objects/pack/test.promisor'
        promisor.write_bytes(b'')
        with self.assertRaises(ValueError):
            fetch.inspect_store(git, db, repo)
        promisor.unlink()
        root = self.commit(git, db)
        git.run(db, 'update-ref', 'refs/replace/'+root, root)
        with self.assertRaises(ValueError):
            fetch.inspect_store(git, db, repo)


if __name__ == '__main__':
    unittest.main()

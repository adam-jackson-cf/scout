"""Cross-process journal publication with a deliberately exposed partial file."""
from contextlib import contextmanager, nullcontext
import json
import multiprocessing
import os
from pathlib import Path
import sys
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
import preparation as prep
from verify_inputs import exclusive_json, pin


def writer(root, role, phase, exposed, ready, snapshot, release, done, result, unlocked, loaded_config, reload_outside):
    config_path = Path(root)/'config.json'
    config = loaded_config if loaded_config is not None else json.loads(config_path.read_bytes())
    clock_path = Path(config['paths']['OUT'])/'clock.json'
    clock = json.loads(clock_path.read_bytes())
    target = clock_path.parent/'journal'/('a'*32+('.start.json' if phase == 'start' else '.json'))
    original_lock, original_pin = prep.operation_lock, prep.pin
    original_publication, original_load = prep.journal_publication, prep.load_config
    reloads = []

    def load_config(path):
        reloads.append(path)
        return original_load(path)

    @contextmanager
    def publication(directory):
        if reload_outside and role == 'b':
            # Counterfactual: the real config re-check runs before acquiring
            # the unchanged publication guard, as in the reviewed code.
            prep.load_config(config_path)
        with original_publication(directory):
            yield

    @contextmanager
    def operation_lock(*args, **kwargs):
        with original_lock(*args, **kwargs):
            if role == 'b':
                ready.set()
            yield

    def observing_pin(path, *args, **kwargs):
        if role == 'b' and Path(path) == target:
            snapshot.set()
        return original_pin(path, *args, **kwargs)

    def publish(path, value):
        path = Path(path)
        if role != 'a' or path != target:
            return exclusive_json(path, value)
        path.parent.mkdir(parents=True, exist_ok=True)
        payload = json.dumps(value, sort_keys=True)+'\n'
        with path.open('x') as stream:
            stream.write(payload[:1])
            stream.flush()
            os.fsync(stream.fileno())
            exposed.set()
            if not release.wait(15):
                raise TimeoutError('publication was not released')
            stream.write(payload[1:])
            stream.flush()
            os.fsync(stream.fileno())

    try:
        guard = patch.object(prep, 'journal_publication', nullcontext if unlocked else publication)
        with guard, patch.object(prep, 'load_config', load_config), patch.object(sys, 'argv', [str(Path(__file__).resolve())]), patch.object(prep, 'read_clock', return_value=clock), patch.object(prep, 'operation_lock', operation_lock), patch.object(prep, 'exclusive_json', publish), patch.object(prep, 'pin', observing_pin), patch.object(prep.uuid, 'uuid4', return_value=SimpleNamespace(hex=role*32)):
            if loaded_config is not None:
                assert isinstance(config, prep.RunConfig)
            if loaded_config is not None and role == 'b':
                # Reproduce binding's empty canonical reservation before Journal.
                canonical = clock_path.parent/'binding/synthetic.json'
                canonical.parent.mkdir()
                with canonical.open('x') as stream:
                    journal = prep.Journal(config, config_path, clock_path, 'a')
                    assert reloads, 'RunConfig re-check did not call load_config'
                    json.dump({'status': 'passed'}, stream)
                    stream.flush()
                    os.fsync(stream.fileno())
                journal.finish('passed', [pin(canonical)])
            else:
                journal = prep.Journal(config, config_path, clock_path, 'a')
                journal.finish('passed')
            result.put((role, 'passed'))
    except BaseException as error:
        result.put((role, repr(error)))
    finally:
        done.set()


class JournalPublicationTests(unittest.TestCase):
    def overlap(self, phase, unlocked=False, real_config=False, reload_outside=False):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            out = root/'preparation-run'
            exclusive_json(root/'config.json', {'schema': 'scout.x1.run-config.v1',
                                               'paths': {'OUT': str(out), 'STUDY': str(root)}})
            exclusive_json(out/'clock.json', {'preparation_started_at': '2026-01-01T00:00:00Z',
                                             'execution_config': pin(root/'config.json')})
            loaded_config = prep.load_config(root/'config.json') if real_config else None
            context = multiprocessing.get_context('spawn')
            exposed, ready, snapshot, release, a_done, b_done = [context.Event() for _ in range(6)]
            result = context.Queue()
            self.addCleanup(result.close)
            common = (phase, exposed, ready, snapshot, release)
            a = context.Process(target=writer, args=(directory, 'a', *common, a_done, result, unlocked, loaded_config, reload_outside))
            b = context.Process(target=writer, args=(directory, 'b', *common, b_done, result, unlocked, loaded_config, reload_outside))
            a.start()
            processes = [a]
            try:
                if not exposed.wait(10):
                    self.fail('writer A did not expose its partial record: '+str(result.get(timeout=1) if a_done.is_set() else 'still running'))
                target = out/'journal'/('a'*32+('.start.json' if phase == 'start' else '.json'))
                self.assertEqual(target.read_bytes(), b'{')
                b.start(); processes.append(b)
                self.assertTrue(ready.wait(10), 'writer B did not acquire its shared operation lock')
                if reload_outside:
                    self.assertTrue(b_done.wait(10), 'unlocked reload did not refuse on the partial start')
                    self.assertEqual(target.read_bytes(), b'{')
                    self.assertFalse(snapshot.is_set(), 'failure must occur before predecessor pinning')
                else:
                    self.assertFalse(snapshot.wait(0.5), 'writer B pinned writer A during publication')
                    self.assertFalse(b_done.is_set(), 'writer B did not wait for publication')
            finally:
                release.set()
                for process in processes:
                    process.join(15)
                    if process.is_alive():
                        process.terminate(); process.join()
            self.assertEqual([process.exitcode for process in processes], [0, 0])
            outcomes = sorted(result.get(timeout=5) for _ in processes)
            if reload_outside:
                self.assertEqual(outcomes[0], ('a', 'passed'))
                self.assertEqual(outcomes[1][0], 'b')
                self.assertRegex(outcomes[1][1], r'^JSONDecodeError\(')
                self.assertFalse((out/'journal'/('b'*32+'.start.json')).exists())
                self.assertFalse((out/'journal'/('b'*32+'.json')).exists())
                self.assertEqual((out/'binding/synthetic.json').read_bytes(), b'')
                return
            self.assertEqual(outcomes, [('a', 'passed'), ('b', 'passed')])
            start = json.loads((out/'journal'/('b'*32+'.start.json')).read_bytes())
            predecessor = next(row for row in start['prior_journals'] if row['path'] == str(target))
            self.assertEqual(predecessor, pin(target))
            prep.verify_tree_pins(start)
            expected = {'.operation.lock', 'clock.json', 'journal'}
            if real_config:
                expected.add('binding')
                self.assertEqual(json.loads((out/'binding/synthetic.json').read_bytes()), {'status': 'passed'})
            self.assertEqual({path.name for path in out.iterdir()}, expected)
            result.close(); result.join_thread()

    def test_start_publication_blocks_predecessor_snapshot(self):
        self.overlap('start')

    def test_terminal_publication_blocks_predecessor_snapshot(self):
        self.overlap('terminal')

    def test_shared_operation_lock_alone_reproduces_both_publication_races(self):
        for phase in ('start', 'terminal'):
            with self.subTest(phase=phase), self.assertRaisesRegex(AssertionError, 'writer B pinned writer A during publication'):
                self.overlap(phase, unlocked=True)

    def test_run_config_reload_waits_for_start_publication(self):
        self.overlap('start', real_config=True)

    def test_unlocked_run_config_reload_reproduces_parse_failure_and_orphan(self):
        self.overlap('start', real_config=True, reload_outside=True)


if __name__ == '__main__':
    unittest.main()

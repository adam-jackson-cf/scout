"""Synthetic only: no installed tools, network, models, candidates or scanners."""
import base64
import csv
import hashlib
import json
from pathlib import Path
import struct
import sys
import tempfile
import unittest
from unittest.mock import patch

HERE = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(HERE))
import attest_closure as closure


class ClosureTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name).resolve()

    def file(self, relative, content=b'bytes'):
        path = self.root/relative
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(content)
        return path

    def test_link_chain(self):
        real = self.file('real')
        (self.root/'second').symlink_to('real')
        (self.root/'first').symlink_to('second')
        resolved, chain = closure.safe_resolve(self.root/'first')
        self.assertEqual(resolved, real)
        self.assertEqual(len(chain), 2)

    def test_codegraph_refused_even_via_link(self):
        (self.root/'alias').symlink_to('.codegraph')
        with self.assertRaisesRegex(ValueError, 'codegraph'):
            closure.safe_resolve(self.root/'alias/file')

    def test_cycle_refused(self):
        (self.root/'a').symlink_to('b')
        (self.root/'b').symlink_to('a')
        with self.assertRaisesRegex(ValueError, 'cycle'):
            closure.safe_resolve(self.root/'a')

    def test_inventory_preserves_links_and_files(self):
        source = self.file('tree/file')
        (source.parent/'link').symlink_to('file')
        inventory = closure.Inventory([source.parent])
        records = inventory.collect()[0]['records']
        self.assertEqual({r['relative_path'] for r in records}, {'.', 'file', 'link'})
        self.assertEqual(len(inventory.files), 1)
        self.assertEqual(next(r for r in records if r['type']=='symlink')['link_target'], 'file')

    def test_unknown_external_link_refused(self):
        source = self.file('outside/file')
        tree = self.root/'tree'
        tree.mkdir()
        (tree/'link').symlink_to(source)
        with self.assertRaisesRegex(ValueError, 'escaped'):
            closure.Inventory([tree]).collect()

    def test_explicit_external_root_counted(self):
        source = self.file('outside/file')
        tree = self.root/'tree'
        tree.mkdir()
        (tree/'link').symlink_to(source)
        inventory = closure.Inventory([tree, source.parent])
        inventory.collect()
        self.assertEqual(set(inventory.files), {str(source)})

    def test_launcher_bytes_never_normalized(self):
        source = self.file('launcher', b'#!/original/python\n')
        expected = closure.pin(source)
        source.write_bytes(b'#!/relocated/python\n')
        with self.assertRaisesRegex(ValueError, 'working-byte mismatch'):
            closure.checked(source, expected)

    def test_retention_deduplicates_and_preserves_original_mode(self):
        first, second = self.file('a'), self.file('b')
        mode = first.stat().st_mode
        rows = closure.retain([closure.pin(first), closure.pin(second)], self.root/'retained')
        self.assertEqual(len(rows), 1)
        self.assertEqual(first.stat().st_mode, mode)
        self.assertEqual(Path(rows[0]['path']).stat().st_mode & 0o222, 0)

    def test_incomplete_retained_object_refused(self):
        source = self.file('source')
        row = closure.pin(source)
        target = self.file('retained/'+row['sha256'], b'incomplete')
        target.chmod(0o444)
        with self.assertRaisesRegex(ValueError, 'retained object'):
            closure.retain([row], target.parent)

    def test_requirements_frozen_hashes(self):
        lock = self.file('requirements.lock', b'# comment\na_b==1.2 \\\n --hash=sha256:'+b'a'*64+b'\n')
        rows = closure.requirements(lock)
        self.assertEqual(rows['a-b']['version'], '1.2')
        self.assertEqual(rows['a-b']['archive_sha256'], ['a'*64])

    def test_requirements_unhashed_refused(self):
        lock = self.file('requirements.lock', b'pkg==1.0\n')
        with self.assertRaisesRegex(ValueError, 'syntax'):
            closure.requirements(lock)

    def python_fixture(self):
        site = self.root/'venv/lib/python3.11/site-packages'
        metadata = self.file(str((site/'demo-1.0.dist-info/METADATA').relative_to(self.root)), b'Name: demo\nVersion: 1.0\n')
        module = self.file(str((site/'demo.py').relative_to(self.root)), b'x = 1\n')
        record = metadata.parent/'RECORD'
        with record.open('w', newline='') as out:
            writer = csv.writer(out)
            for path in (metadata, module):
                value = base64.urlsafe_b64encode(hashlib.sha256(path.read_bytes()).digest()).decode().rstrip('=')
                writer.writerow([str(path.relative_to(site)), 'sha256='+value, path.stat().st_size])
            writer.writerow([str(record.relative_to(site)), '', ''])
        lock = self.file('requirements.lock', b'demo==1.0 --hash=sha256:'+b'a'*64+b'\n')
        return site, module, lock

    def test_python_record_verified(self):
        site, module, lock = self.python_fixture()
        rows = closure.python_packages(self.root/'venv', lock)
        self.assertEqual(rows['demo']['verified_record_rows'], 3)

    def test_python_tamper_refused(self):
        site, module, lock = self.python_fixture()
        module.write_text('tampered')
        with self.assertRaisesRegex(ValueError, 'RECORD mismatch'):
            closure.python_packages(self.root/'venv', lock)

    def test_npm_identity_not_just_version(self):
        self.file('node/package-lock.json', json.dumps({'lockfileVersion': 3, 'packages': {
            '': {}, 'node_modules/demo': {'version': '1', 'integrity': 'sha512-abc'}}}).encode())
        self.file('node/node_modules/demo/package.json', b'{"name":"substitution","version":"1"}')
        with self.assertRaisesRegex(ValueError, 'identity'):
            closure.npm_packages(self.root/'node')

    def test_missing_node_dependency_refused(self):
        package = self.file('node/package.json', b'{"name":"demo","dependencies":{"missing":"1"}}').parent
        with self.assertRaisesRegex(ValueError, 'missing required'):
            closure.node_edges(package, package, {})

    def test_model_cli_never_launched(self):
        with patch.object(closure.subprocess, 'run') as run:
            for model in ('claude', 'codex'):
                with self.assertRaisesRegex(ValueError, 'model executable'):
                    closure.Commands({}).run(['/forbidden/'+model, '--version'], self.root)
            run.assert_not_called()

    def test_json_stream_rejects_partial(self):
        self.assertEqual(closure.json_stream(' {"x":1}\n{"x":2} '), [{'x': 1}, {'x': 2}])
        with self.assertRaises(json.JSONDecodeError):
            closure.json_stream('{"x":')

    def test_native_missing_dependency_refused(self):
        image = self.file('image', struct.pack('<8I', 0xfeedfacf, 0x100000c, 0, 6, 0, 0, 0, 0))
        class FakeCommands:
            def run(self, argv, cwd):
                if '-L' in argv:
                    return str(image)+':\n\t/absent/test-library.dylib (compatibility version 1.0.0)\n'
                return ''
        with self.assertRaisesRegex(ValueError, 'missing native dependency'):
            closure.native_closure([image], FakeCommands(), self.root)

    def test_native_self_identity_not_dependency(self):
        image = self.file('image', struct.pack('<8I', 0xfeedfacf, 0x100000c, 0, 6, 0, 0, 0, 0))
        class FakeCommands:
            def run(self, argv, cwd):
                if '-L' in argv:
                    return str(image)+':\n\t@rpath/self.dylib (compatibility version 1.0.0)\n'
                return 'cmd LC_ID_DYLIB\ncmdsize 48\nname @rpath/self.dylib (offset 24)\n'
        result = closure.native_closure([image], FakeCommands(), self.root)
        self.assertEqual(result['images'][0]['dependencies'], [])


if __name__ == '__main__':
    unittest.main()

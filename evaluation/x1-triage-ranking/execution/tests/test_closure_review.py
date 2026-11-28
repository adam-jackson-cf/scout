"""Regression and planted-drift/null cases for closure implementation review R1."""
import base64
import hashlib
import json
from pathlib import Path
import struct
import sys
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
import attest_closure as c
from closure.native import image_filetype


class ClosureReviewTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name).resolve()

    def file(self, relative, content=b'bytes'):
        path = self.root/relative
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(content)
        return path

    def json(self, relative, value):
        return self.file(relative, json.dumps(value).encode())

    def test_r101_real_marker_lines(self):
        text = '\n'.join(line+' --hash=sha256:'+'a'*64 for line in (
            "deprecated==1.3.1 ; python_full_version < '3.12'",
            "deprecated==3.0.0 ; python_full_version >= '3.12'",
            "librt==0.16.0 ; platform_python_implementation != 'PyPy'"))
        with patch.object(c.platform, 'python_version_tuple', return_value=('3', '11', '11')), patch.object(c.platform, 'python_implementation', return_value='CPython'):
            rows = c.requirements(self.file('lock', text.encode()))
        self.assertEqual(set(rows), {'deprecated', 'librt'})
        self.assertEqual(rows['deprecated']['version'], '1.3.1')

    def test_r101_unknown_marker_variable_and_operator(self):
        for marker in ("os_name == 'posix'", "python_full_version ~= '3.11'", "platform_python_implementation < 'PyPy'"):
            with self.subTest(marker=marker), self.assertRaisesRegex(ValueError, 'unsupported'):
                c.marker_applies(marker)

    def test_r101_duplicate_applicable_row(self):
        text = ('demo==1 --hash=sha256:'+'a'*64+'\n')*2
        with self.assertRaisesRegex(ValueError, 'duplicate applicable'):
            c.requirements(self.file('lock', text.encode()))

    def macho(self, relative, filetype):
        return self.file(relative, struct.pack('<IIIIIIII', 0xfeedfacf, 0x100000c, 0, filetype, 0, 0, 0, 0))

    def test_r102_executable_context_carried_through_rpath(self):
        exe = self.macho('python/bin/python3.11', 2)
        lib = self.macho('python/lib/libpython3.11.dylib', 6)
        leaf = self.macho('python/lib/leaf.dylib', 6)
        class Fake:
            def run(_, argv, cwd):
                image = Path(argv[-1])
                if '-L' in argv:
                    dep = {exe: '@executable_path/../lib/libpython3.11.dylib', lib: '@rpath/leaf.dylib', leaf: None}[image]
                    return str(image)+':\n'+('\t'+dep+' (compatibility version 1.0.0)\n' if dep else '')
                return 'cmd LC_RPATH\ncmdsize 48\npath @executable_path/../lib (offset 12)\n' if image == lib else ''
        result = c.native_closure([exe, lib, leaf], Fake(), self.root)
        by_path = {row['image']['path']: row for row in result['images']}
        self.assertEqual(by_path[str(lib)]['executable_context'], str(exe))
        self.assertEqual(by_path[str(lib)]['dependencies'][0]['path'], str(leaf))

    def test_r102_missing_executable_context_refuses(self):
        lib = self.macho('lib.dylib', 6)
        class Fake:
            def run(_, argv, cwd):
                return str(lib)+':\n\t@executable_path/leaf.dylib (compatibility version 1.0.0)\n' if '-L' in argv else ''
        with self.assertRaisesRegex(ValueError, 'executable context'):
            c.native_closure([lib], Fake(), self.root)

    def pnpm_fixture(self):
        text = """lockfileVersion: '9.0'
packages:
  '@vitest/mocker@5.0.0':
    resolution: {integrity: sha512-66PGTMIiVJP3t4a5yxU9qPtf7MdTBs8jmToMvy+HVflB3Yy13WJZTtPePdvU+wjRV02SKK5doLbSA6o9pwOmiA==}
    peerDependencies:
      msw: ^2.4.9
      vite: ^8.2.2
    peerDependenciesMeta:
      msw:
        optional: true
      vite:
        optional: true
snapshots:
  '@vitest/mocker@5.0.0': {}
"""
        root = self.root/'claw'
        self.file('claw/pnpm-lock.yaml', text.encode())
        self.file('claw/node_modules/.pnpm/lock.yaml', text.encode())
        self.json('claw/package.json', {'name': 'claw', 'version': '0.7.3'})
        self.json('claw/node_modules/.pnpm/@vitest+mocker@5.0.0/node_modules/@vitest/mocker/package.json', {'name': '@vitest/mocker', 'version': '5.0.0'})
        return root

    def test_r103_nested_peer_keys_are_not_packages(self):
        value = c.pnpm_packages(self.pnpm_fixture())
        self.assertEqual(set(value['locked_packages']), {'@vitest/mocker@5.0.0'})

    def npm_fixture(self):
        root = self.root/'node'
        self.json('node/package-lock.json', {'lockfileVersion': 3, 'packages': {'': {}, 'node_modules/pnpm': {'version': '11.25.0', 'integrity': 'sha512-abc'}}})
        self.json('node/node_modules/pnpm/package.json', {'name': 'pnpm', 'version': '11.25.0'})
        return root

    def test_r104_internal_tarball_node_modules_pass(self):
        root = self.npm_fixture()
        self.json('node/node_modules/pnpm/dist/node_modules/yallist/package.json', {'name': 'yallist', 'version': '1'})
        self.json('node/node_modules/pnpm/dist/node_modules/@reflink/test/package.json', {'name': '@reflink/test', 'version': '1'})
        self.assertEqual(len(c.npm_packages(root)['packages']), 1)

    def test_r104_stray_top_level_package_refuses(self):
        root = self.npm_fixture()
        self.json('node/node_modules/stray/package.json', {'name': 'stray', 'version': '1'})
        with self.assertRaisesRegex(ValueError, 'unlocked npm'):
            c.npm_packages(root)

    def test_r104_stray_nested_dependency_refuses(self):
        root = self.npm_fixture()
        self.json('node/node_modules/pnpm/node_modules/stray/package.json', {'name': 'stray', 'version': '1'})
        with self.assertRaisesRegex(ValueError, 'unlocked npm'):
            c.npm_packages(root)

    def test_r105_consistently_substituted_installed_locks_refuse(self):
        root = self.pnpm_fixture()
        self.file('claw/pnpm-workspace.yaml', b'packages: []\n')
        rows = [dict(c.pin(root/name), path='third_party/clawpatch/'+name) for name in ('package.json', 'pnpm-lock.yaml', 'pnpm-workspace.yaml')]
        pins = {'submodule': {'working_files': rows}}
        c.verify_claw_assets(root, pins)
        for path in (root/'pnpm-lock.yaml', root/'node_modules/.pnpm/lock.yaml'):
            path.write_bytes(path.read_bytes()+b'# substituted\n')
        with self.assertRaisesRegex(ValueError, 'working-byte mismatch'):
            c.verify_claw_assets(root, pins)

    def git_fixture(self):
        exe = self.file('keg/bin/git')
        helper = self.file('keg/libexec/git-core/git-diff')
        identity = dict(c.pin(exe), version='git version 2.56.0')
        config = {'paths': {'OUT': str(self.root/'out')}, 'repositories': [{'repository': 'owner/repo'+str(n)} for n in range(5)]}
        for n in range(5):
            self.json('out/objects/owner__repo'+str(n)+'.json', {'schema': 'scout.x1.objects.v1', 'status': 'passed', 'git_identity': identity})
        return config, exe, helper

    def test_r106_git_agreement_and_helper_root(self):
        config, exe, helper = self.git_fixture()
        identity, root, receipts = c.object_git(config)
        self.assertEqual(identity['path'], str(exe))
        self.assertEqual(root, exe.parent.parent)
        inventory = c.Inventory([root]); inventory.collect()
        self.assertIn(str(helper), inventory.files)
        self.assertEqual(len(receipts), 5)

    def test_r106_git_disagreement_refuses(self):
        config, _, _ = self.git_fixture()
        path = self.root/'out/objects/owner__repo4.json'
        value = json.loads(path.read_bytes()); value['git_identity']['sha256'] = '0'*64
        path.write_text(json.dumps(value))
        with self.assertRaisesRegex(ValueError, 'disagree'):
            c.object_git(config)

    def test_r107_unpinned_selected_source_refuses(self):
        source = self.file('checkout/internal/untracked.go')
        with self.assertRaisesRegex(ValueError, 'unpinned selected'):
            c.selected_source(source, self.root/'checkout', {})
        self.assertEqual(c.selected_source(source, self.root/'checkout', {str(source): c.pin(source)}), c.pin(source))

    def test_r107_reviewed_frontend_bytes_must_stay_signed(self):
        root = self.root/'checkout'
        source = self.file('checkout/evaluation/x1-triage-ranking/execution/model_free/main.go', b'package main\n')
        row = c.pin(source)
        study = root/'evaluation/x1-triage-ranking'
        out = study/'preparation-run'
        review = self.file(str((study/'preparation-reviews/signoff.txt').relative_to(self.root)),
                           ('**Zero issues. Signed off.**\nSigned-off scope: '+c.GO_REVIEW_SCOPE+'\n'+row['sha256']+'\n').encode())
        receipt = self.json(str((out/c.GO_REVIEW_RELATIVE).relative_to(self.root)), {'scope': c.GO_REVIEW_SCOPE, 'record': c.pin(review), 'reviewed': [row]})
        selected, pinned = c.reviewed_go_sources(root, out)
        self.assertEqual(selected, {str(source): row})
        self.assertEqual(pinned, [c.pin(receipt), c.pin(review)])
        source.write_text('drift')
        with self.assertRaisesRegex(ValueError, 'pin mismatch'):
            c.reviewed_go_sources(root, out)

    def test_r108_own_closure_is_not_whole_execution_manifest(self):
        rows = c.own_sources()
        self.assertEqual({Path(row['path']).relative_to(c.HERE).as_posix() for row in rows},
                         {'attest_closure.py', 'closure/native.py', 'preparation.py', 'verify_inputs.py', 'run-config.json'})
        self.assertFalse(any('binding' in row['path'] or 'identifiers' in row['path'] for row in rows))

    def test_r109_relocated_runtime_refuses(self):
        runtime = {'numerical_interpreter': {'command': str(self.root/'original/.tools/venv/bin/python')}}
        with self.assertRaisesRegex(ValueError, 'runtime location'):
            c.runtime_python(self.root/'relocated', runtime)

    def test_r109_resolved_path_must_equal_signed_path(self):
        source = self.file('real')
        with self.assertRaisesRegex(ValueError, 'resolved path mismatch'):
            c.checked(source, dict(c.pin(source), resolved_path=str(self.root/'other')))

    def test_r109_interpreter_outside_inventory_refuses(self):
        source = self.file('outside/python')
        (self.root/'runtime').mkdir()
        with self.assertRaisesRegex(ValueError, 'outside enumerated'):
            c.verify_interpreter_roots([c.pin(source)], c.Inventory([self.root/'runtime']))

    def test_r109_wrong_running_interpreter_refuses(self):
        root = self.root/'runtime'
        python = self.file('runtime/.tools/venv/bin/python')
        runtime = {'numerical_interpreter': dict(c.pin(python), command=str(python))}
        with patch.object(c.sys, 'executable', str(self.root/'different')), self.assertRaisesRegex(ValueError, 'attestor requires pinned'):
            c.runtime_python(root, runtime)

    def test_r109_unisolated_interpreter_refuses(self):
        root = self.root/'runtime'
        python = self.file('runtime/.tools/venv/bin/python')
        runtime = {'numerical_interpreter': dict(c.pin(python), command=str(python))}
        with patch.object(c.sys, 'executable', str(python)), patch.object(c.sys, 'flags', SimpleNamespace(isolated=False)), patch.object(c.platform, 'python_version', return_value='3.11.11'), self.assertRaisesRegex(ValueError, 'attestor requires pinned'):
            c.runtime_python(root, runtime)

    def test_r110_nonexecutable_node_addon_is_inspected(self):
        addon = self.macho('oxlint.darwin-arm64.node', 6)
        addon.chmod(0o644)
        calls = []
        class Fake:
            def run(_, argv, cwd):
                calls.append(argv)
                return str(addon)+':\n' if '-L' in argv else ''
        result = c.native_closure([addon], Fake(), self.root)
        self.assertEqual(result['images'][0]['image']['path'], str(addon))
        self.assertEqual(len(calls), 2)

    def recheck_fixture(self):
        source = self.file('runtime/file')
        link = self.root/'runtime/link'; link.symlink_to('file')
        inventory = c.Inventory([source.parent]); enumeration = inventory.collect()
        retained = c.retain(list(inventory.files.values()), self.root/'retained')
        value = {'schema': 'scout.x1.installed-closure.v1', 'status': 'passed', 'complete': True,
                 'enumeration': enumeration, 'inputs': [], 'source_pins': [],
                 'original_runtime_files': list(inventory.files.values()), 'submodule_working_files': [],
                 'deferred_working_files_verified': [], 'retained_files': retained,
                 'executable_resolutions': [c.checked(link, c.pin(source))]}
        return self.json('closure.json', value), source, link, retained

    def test_r111_recheck_planted_null_pass(self):
        receipt, _, _, _ = self.recheck_fixture()
        c.verify_installed_closure(receipt)

    def test_r111_recheck_added_file_refuses(self):
        receipt, _, _, _ = self.recheck_fixture(); self.file('runtime/added')
        with self.assertRaisesRegex(ValueError, 'file/link/mode set drift'):
            c.verify_installed_closure(receipt)

    def test_r111_recheck_mode_change_refuses(self):
        receipt, source, _, _ = self.recheck_fixture(); source.chmod(0o555)
        with self.assertRaisesRegex(ValueError, 'file/link/mode set drift'):
            c.verify_installed_closure(receipt)

    def test_r111_recheck_retargeted_link_refuses(self):
        receipt, source, link, _ = self.recheck_fixture(); link.unlink(); link.symlink_to(source)
        with self.assertRaisesRegex(ValueError, 'file/link/mode set drift'):
            c.verify_installed_closure(receipt)

    def test_r111_recheck_writable_retained_object_refuses(self):
        receipt, _, _, retained = self.recheck_fixture(); Path(retained[0]['path']).chmod(0o644)
        with self.assertRaisesRegex(ValueError, 'became writable'):
            c.verify_installed_closure(receipt)

    def test_r111_recheck_entrypoint_resolution_drift_refuses(self):
        receipt, _, _, _ = self.recheck_fixture()
        external = self.file('external'); request = self.root/'launcher'; request.symlink_to('external')
        value = json.loads(receipt.read_bytes()); value['executable_resolutions'] = [c.checked(request, c.pin(external))]
        receipt.write_text(json.dumps(value)); request.unlink(); request.symlink_to('other')
        with self.assertRaisesRegex(ValueError, 'entrypoint resolution drift'):
            c.verify_installed_closure(receipt)

    def deferred_fixture(self):
        source = self.file('working')
        row = dict(c.pin(source), path='working', git_mode='100644', git_blob='f'*40)
        return row

    def test_r111_deferred_empty_still_verifies_working_files(self):
        row = self.deferred_fixture()
        working, deferred = c.verify_deferred(self.root, [row], [], 176)
        self.assertEqual(len(working), 1); self.assertEqual(deferred, [])
        (self.root/'working').write_text('drift')
        with self.assertRaisesRegex(ValueError, 'working-file bytes mismatch'):
            c.verify_deferred(self.root, [row], [], 176)

    def test_r111_deferred_count_refuses(self):
        row = self.deferred_fixture()
        with self.assertRaisesRegex(ValueError, 'incomplete/duplicated'):
            c.verify_deferred(self.root, [row], [row], 176)

    def test_r111_deferred_set_refuses(self):
        row = self.deferred_fixture()
        with self.assertRaisesRegex(ValueError, 'incomplete/duplicated'):
            c.verify_deferred(self.root, [row], [dict(row, path='other')], 1)

    def test_r111_deferred_field_refuses(self):
        row = self.deferred_fixture()
        for field in ('sha256', 'bytes', 'git_mode', 'git_blob'):
            with self.subTest(field=field), self.assertRaisesRegex(ValueError, 'differs from signed'):
                c.verify_deferred(self.root, [row], [dict(row, **{field: 'changed'})], 1)

    def test_r111_deferred_nonempty_null_pass(self):
        row = self.deferred_fixture()
        working, observed = c.verify_deferred(self.root, [row], [row], 1)
        self.assertEqual(working, observed)

    def go_fixture(self, replaced=False, mismatch=False):
        checkout = self.root/'checkout'; checkout.mkdir()
        gomod = self.file('modules/go.mod', b'module example.com/demo\n')
        source = self.file('modules/demo.go', b'package demo\n')
        modsum = 'h1:'+base64.b64encode(hashlib.sha256((c.digest(gomod)+'  go.mod\n').encode()).digest()).decode()
        self.file('checkout/go.sum', ('example.com/demo v1.0.0 h1:signed\nexample.com/demo v1.0.0/go.mod '+modsum+'\n').encode())
        self.file('cache/cache/download/example.com/demo/@v/v1.0.0.ziphash', b'h1:wrong' if mismatch else b'h1:signed')
        module = {'Path': 'example.com/demo', 'Version': 'v1.0.0', 'Dir': str(source.parent), 'GoMod': str(gomod)}
        if replaced:
            module['Replace'] = {'Path': '../substitute'}
        package = {'ImportPath': 'example.com/demo', 'Dir': str(source.parent), 'GoFiles': ['demo.go'], 'Module': module}
        root = self.root
        class Fake:
            environment = {'GOMODCACHE': str(root/'cache')}
            def run(_, argv, cwd):
                if 'env' in argv: return str(root/'goroot')
                if '-deps' in argv: return json.dumps(package)
                if '-m' in argv: return json.dumps(module)
                if 'verify' in argv: return 'all modules verified\n'
                if 'version' in argv: return 'go version go1.26.10 darwin/arm64\n'
                raise AssertionError(argv)
        return checkout, Fake()

    def test_r111_go_sum_mismatch_refuses(self):
        checkout, commands = self.go_fixture(mismatch=True)
        with self.assertRaisesRegex(ValueError, 'sum not bound'):
            c.go_metadata('/fake/go', checkout, commands)

    def test_r111_go_replace_refuses(self):
        checkout, commands = self.go_fixture(replaced=True)
        with self.assertRaisesRegex(ValueError, 'replaced Go module'):
            c.go_metadata('/fake/go', checkout, commands)

    def test_r111_go_metadata_planted_null_pass(self):
        checkout, commands = self.go_fixture()
        metadata, roots = c.go_metadata('/fake/go', checkout, commands)
        self.assertEqual(metadata['module_verification'], 'all modules verified')
        self.assertEqual(metadata['modules'][0]['verified_sum'], 'h1:signed')


if __name__ == '__main__':
    unittest.main()

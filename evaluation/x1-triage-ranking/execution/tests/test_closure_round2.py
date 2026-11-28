"""Planted-null/refusal and producer-wiring checks for closure review round 2."""
from contextlib import ExitStack, nullcontext
import json
from pathlib import Path
import struct
import sys
from types import SimpleNamespace
import unittest
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
import attest_closure as c
from closure.native import image_filetype
import test_closure_review as previous


class RoundTwoTests(unittest.TestCase):
    def setUp(self):
        self.fixture = previous.ClosureReviewTests()
        self.fixture.setUp()
        self.addCleanup(self.fixture.doCleanups)
        self.root = self.fixture.root
        self.file = self.fixture.file
        self.json = self.fixture.json

    def thin(self, kind=2, cpu=0x100000c):
        return struct.pack('<8I', 0xfeedfacf, cpu, 0, kind, 0, 0, 0, 0)

    def java(self):
        return (struct.pack('>IHHH', 0xcafebabe, 0, 52, 5)
                + b'\x01\x00\x04Test\x07\x00\x01\x01\x00\x10java/lang/Object\x07\x00\x03'
                + struct.pack('>7H', 0x21, 2, 4, 0, 0, 0, 0))

    def fat(self, wide=False):
        header = struct.pack('>II', 0xcafebabf if wide else 0xcafebabe, 2)
        entry = '>IIQQII' if wide else '>IIIII'
        def arch(cpu, offset):
            args = (cpu, 0, offset, 32, 8)
            return struct.pack(entry, *(args+(0,) if wide else args))
        data = header+arch(0x1000007, 256)+arch(0x100000c, 512)
        return data.ljust(256, b'\0')+self.thin(6, 0x1000007)+b'\0'*(512-288)+self.thin(2)

    def test_r201_java_is_non_native_without_suffix_filter(self):
        self.assertIsNone(image_filetype(self.file('opaque.data', self.java())))

    def test_r201_thin_arm64(self):
        self.assertEqual(image_filetype(self.file('thin', self.thin(6))), 6)

    def test_r201_fat32_uses_arm64_slice(self):
        self.assertEqual(image_filetype(self.file('fat32', self.fat())), 2)

    def test_r201_fat64_uses_arm64_slice(self):
        self.assertEqual(image_filetype(self.file('fat64', self.fat(True))), 2)

    def test_r201_truncated_fat_refuses(self):
        with self.assertRaisesRegex(ValueError, 'truncated fat'):
            image_filetype(self.file('bad', struct.pack('>II', 0xcafebabe, 2)+b'\0'*20))

    def test_r201_out_of_bounds_fat_refuses(self):
        data = struct.pack('>II5I', 0xcafebabe, 1, 0x100000c, 0, 256, 32, 8)
        with self.assertRaisesRegex(ValueError, 'bounds/alignment'):
            image_filetype(self.file('bad', data))

    def test_r201_misaligned_fat_refuses(self):
        data = struct.pack('>II5I', 0xcafebabe, 1, 0x100000c, 0, 257, 32, 8).ljust(257, b'\0')+self.thin()
        with self.assertRaisesRegex(ValueError, 'bounds/alignment'):
            image_filetype(self.file('bad', data))

    def test_r201_truncated_thin_refuses(self):
        with self.assertRaisesRegex(ValueError, 'truncated Mach-O'):
            image_filetype(self.file('bad', b'\xcf\xfa\xed\xfe'))

    def go_main_fixture(self, frontend=False, linked=False):
        checkout, delegate = self.fixture.go_fixture()
        relative = 'evaluation/x1-triage-ranking/execution/model_free/main.go' if frontend else 'internal/pkg/extra.go'
        logical = checkout/relative
        logical.parent.mkdir(parents=True, exist_ok=True)
        if linked:
            logical.symlink_to(self.root/'modules/demo.go')
        else:
            logical.write_text('package main\n' if frontend else 'package pkg\n')
        module = {'Main': True, 'Path': 'scout', 'Dir': str(checkout)}
        package = {'ImportPath': 'scout/pkg', 'Dir': str(logical.parent), 'GoFiles': [logical.name], 'Module': module}
        class Fake:
            environment = delegate.environment
            def run(_, argv, cwd):
                if '-deps' in argv: return json.dumps(package)
                if '-m' in argv: return json.dumps(module)
                return delegate.run(argv, cwd)
        return checkout, logical, Fake()

    def test_r202_go_metadata_logical_source_null(self):
        checkout, logical, commands = self.go_main_fixture()
        row = c.pin(logical)
        result, _ = c.go_metadata('/fake/go', checkout, commands, {str(logical): row})
        self.assertIn(row, result['selected_files'])

    def test_r202_go_metadata_unpinned_inside_root_refuses(self):
        checkout, _, commands = self.go_main_fixture()
        with self.assertRaisesRegex(ValueError, 'unpinned selected Go source'):
            c.go_metadata('/fake/go', checkout, commands)

    def test_r202_go_metadata_escaping_file_symlink_refuses(self):
        checkout, logical, commands = self.go_main_fixture(linked=True)
        # Even the target byte hash cannot authorise an unpinned logical chain.
        with self.assertRaisesRegex(ValueError, 'source link chain'):
            c.go_metadata('/fake/go', checkout, commands, {str(logical): c.pin(logical)})

    def authority(self):
        checkout = self.root/'checkout'
        source = self.file('checkout/evaluation/x1-triage-ranking/execution/model_free/main.go', b'package main\n')
        row = c.pin(source)
        study = checkout/'evaluation/x1-triage-ranking'
        record = self.file(str((study/'preparation-reviews/council.txt').relative_to(self.root)),
                           ('**Zero issues. Signed off.**\nSigned-off scope: '+c.GO_REVIEW_SCOPE+'\n'+row['sha256']+'\n').encode())
        out = study/'preparation-run'
        receipt = self.json(str((out/c.GO_REVIEW_RELATIVE).relative_to(self.root)), {'scope': c.GO_REVIEW_SCOPE, 'record': c.pin(record), 'reviewed': [row]})
        return checkout, out, source, record, receipt

    def test_r203_missing_dedicated_receipt_refuses(self):
        with self.assertRaises(FileNotFoundError):
            c.reviewed_go_sources(self.root/'checkout', self.root/'checkout/evaluation/x1-triage-ranking/preparation-run')

    def test_r203_valid_receipt_and_council_record_are_pinned(self):
        root, out, source, record, receipt = self.authority()
        reviewed, pins = c.reviewed_go_sources(root, out)
        self.assertEqual(reviewed, {str(source): c.pin(source)})
        self.assertEqual(pins, [c.pin(receipt), c.pin(record)])
        self.assertNotEqual(receipt.name, 'implementation.json')

    def test_r203_wrong_scope_refuses(self):
        root, out, _, _, receipt = self.authority()
        value = json.loads(receipt.read_bytes()); value['scope'] = 'execution-implementation'
        receipt.write_text(json.dumps(value))
        with self.assertRaisesRegex(ValueError, 'invalid scoped review'):
            c.reviewed_go_sources(root, out)

    def test_r203_drifted_receipt_refuses(self):
        root, out, _, _, receipt = self.authority()
        value = json.loads(receipt.read_bytes()); value['extra'] = 'drift'
        receipt.write_text(json.dumps(value))
        with self.assertRaisesRegex(ValueError, 'invalid scoped review'):
            c.reviewed_go_sources(root, out)

    def test_r203_drifted_council_record_refuses(self):
        root, out, _, record, _ = self.authority(); record.write_text('drift')
        with self.assertRaisesRegex(ValueError, 'pin mismatch'):
            c.reviewed_go_sources(root, out)

    def test_r203_unrelated_reviewed_byte_refuses(self):
        root, out, _, record, receipt = self.authority()
        unrelated = self.file('checkout/evaluation/x1-triage-ranking/execution/binding.py')
        value = json.loads(receipt.read_bytes()); value['reviewed'].append(c.pin(unrelated))
        record.write_text(record.read_text()+c.pin(unrelated)['sha256']+'\n')
        value['record'] = c.pin(record); receipt.write_text(json.dumps(value))
        with self.assertRaisesRegex(ValueError, 'only unique model_free'):
            c.reviewed_go_sources(root, out)

    def test_r203_exact_selected_frontend_review_set(self):
        root, logical, commands = self.go_main_fixture(frontend=True)
        allowed = {str(logical): c.pin(logical)}
        result, _ = c.go_metadata('/fake/go', root, commands, allowed, allowed)
        self.assertIn(c.pin(logical), result['selected_files'])
        for reviewed in ([], [str(logical), str(logical.parent/'not-selected.go')]):
            with self.subTest(reviewed=reviewed), self.assertRaisesRegex(ValueError, 'exactly the selected'):
                c.go_metadata('/fake/go', root, commands, allowed, reviewed)

    def test_r203_main_journal_binds_dedicated_receipt_and_record(self):
        root, out, _, record, receipt = self.authority()
        config_path, pins_path, clock = (self.file(name) for name in ('config.json', 'pins.json', 'clock.json'))
        scratch = self.root/'scratch'
        config = {'paths': {'ROOT': str(root), 'OUT': str(out), 'SCRATCH': str(scratch)}}
        captured = {}
        class Journal:
            id = 'synthetic'
            def __init__(_, *args, **kwargs): captured.update(kwargs)
            def finish(_, *args, **kwargs): pass
        argv = ['attest_closure.py', '--config', str(config_path), '--clock', str(clock), '--pins', str(pins_path),
                '--preflight', '--runtime-root', str(self.root/'runtime'), '--module-cache', str(scratch/'mod'), '--dyld-cache-root', str(self.root/'dyld')]
        with patch.object(c.sys, 'argv', argv), patch.object(c, 'load_config', return_value=config), patch.object(c, 'read_clock'), patch.object(c, 'operation_lock', return_value=nullcontext()), patch.object(c, 'Journal', Journal), patch.object(c, 'attest', return_value={'synthetic_test_only': True}):
            c.main()
        self.assertIn(c.pin(receipt), captured['source_pins'])
        self.assertIn(c.pin(record), captured['source_pins'])
        self.assertTrue((scratch/'closure-preflight/synthetic.json').is_file())
        self.assertFalse((out/'closure.json').exists())

    def change_git(self, config, identity=None, status=None):
        for repo in config['repositories']:
            path = Path(config['paths']['OUT'])/'objects'/(repo['repository'].replace('/', '__')+'.json')
            value = json.loads(path.read_bytes())
            if identity is not None: value['git_identity'] = identity
            if status is not None: value['status'] = status
            path.write_text(json.dumps(value))

    def test_r204_git_shim_refuses(self):
        config, exe, _ = self.fixture.git_fixture()
        shim = exe.parent/'shim'; shim.symlink_to(exe)
        self.change_git(config, dict(c.pin(shim), version='git version 2.56.0'))
        with self.assertRaisesRegex(ValueError, 'not a shim'):
            c.object_git(config)

    def test_r204_git_non_bin_refuses(self):
        config, _, _ = self.fixture.git_fixture()
        exe = self.file('not-bin/git')
        self.change_git(config, dict(c.pin(exe), version='git version 2.56.0'))
        with self.assertRaisesRegex(ValueError, 'not a shim'):
            c.object_git(config)

    def test_r204_missing_git_core_refuses(self):
        config, _, helper = self.fixture.git_fixture(); helper.unlink(); helper.parent.rmdir()
        with self.assertRaisesRegex(ValueError, 'lacks git-core'):
            c.object_git(config)

    def test_r204_nonpassing_object_receipt_refuses(self):
        config, _, _ = self.fixture.git_fixture(); self.change_git(config, status='refused')
        with self.assertRaisesRegex(ValueError, 'passing object receipt'):
            c.object_git(config)

    def git_commands(self, version='git version 2.56.0', head='a'*40):
        calls = []
        class Fake:
            def run(_, argv, cwd):
                calls.append(argv)
                return version if '--version' in argv else head
        return Fake(), calls

    def test_r204_keg_git_version_head_null_and_command_identity(self):
        commands, calls = self.git_commands()
        identity = {'path': '/synthetic/keg/bin/git', 'version': 'git version 2.56.0'}
        c.verify_git_runtime(identity, self.root, 'third_party/clawpatch', 'a'*40, commands)
        self.assertEqual([call[0] for call in calls], [identity['path']]*2)
        self.assertEqual(calls[1], [identity['path'], '-C', self.root/'third_party/clawpatch', 'rev-parse', 'HEAD'])

    def test_r204_git_version_drift_refuses(self):
        commands, _ = self.git_commands(version='drift')
        with self.assertRaisesRegex(ValueError, 'version drift'):
            c.verify_git_runtime({'path': '/keg/bin/git', 'version': 'git version 2.56.0'}, self.root, 'claw', 'a'*40, commands)

    def test_r204_git_head_mismatch_refuses(self):
        commands, _ = self.git_commands(head='b'*40)
        with self.assertRaisesRegex(ValueError, 'gitlink/HEAD mismatch'):
            c.verify_git_runtime({'path': '/keg/bin/git', 'version': 'git version 2.56.0'}, self.root, 'claw', 'a'*40, commands)

    def inventory_fixture(self):
        _, exe, helper = self.fixture.git_fixture()
        tools = self.file('runtime/tools/tool').parent
        base = self.file('runtime/python/python').parent
        return tools, base, exe.parent.parent, helper

    def test_r204_runtime_inventory_includes_keg_null(self):
        tools, base, keg, helper = self.inventory_fixture()
        with patch.object(c.sys, 'base_prefix', str(base)):
            inventory, _ = c.runtime_inventory(tools, [], keg, [], [])
        self.assertIn(str(helper), inventory.files)

    def test_r204_runtime_inventory_keg_escape_refuses(self):
        tools, base, keg, helper = self.inventory_fixture()
        outside = self.file('outside/helper')
        (helper.parent/'escape').symlink_to(outside)
        with patch.object(c.sys, 'base_prefix', str(base)), self.assertRaisesRegex(ValueError, 'escaped dependency'):
            c.runtime_inventory(tools, [], keg, [], [])

    def claw_fixture(self):
        root = self.fixture.pnpm_fixture()
        self.file('claw/pnpm-workspace.yaml', b'packages: []\n')
        rows = [dict(c.pin(root/name), path='third_party/clawpatch/'+name) for name in ('package.json', 'pnpm-lock.yaml', 'pnpm-workspace.yaml')]
        return root, {'submodule': {'working_files': rows}}

    def test_r204_claw_assets_precede_pnpm_null(self):
        root, pins = self.claw_fixture()
        original = c.pnpm_packages
        with patch.object(c, 'verify_claw_assets', wraps=c.verify_claw_assets) as verify:
            def interpret(path):
                self.assertEqual(verify.call_count, 1)
                return original(path)
            with patch.object(c, 'pnpm_packages', side_effect=interpret):
                assets, result = c.claw_runtime(root, pins)
        self.assertEqual(len(assets), 3)
        self.assertTrue(result['packages'])

    def test_r204_substituted_claw_refuses_before_pnpm(self):
        root, pins = self.claw_fixture()
        for path in (root/'pnpm-lock.yaml', root/'node_modules/.pnpm/lock.yaml'):
            path.write_bytes(path.read_bytes()+b'# consistent drift\n')
        with patch.object(c, 'pnpm_packages') as pnpm, self.assertRaisesRegex(ValueError, 'working-byte mismatch'):
            c.claw_runtime(root, pins)
        pnpm.assert_not_called()

    def native_inventory(self, missing=False):
        addon = self.file('native/nonexecutable.node', self.thin(6)); addon.chmod(0o644)
        self.file('native/java.class', self.java())
        inventory = c.Inventory([addon.parent]); inventory.collect()
        class Fake:
            def run(_, argv, cwd):
                if '-L' not in argv: return ''
                return str(argv[-1])+':\n'+('\t/absent/native.dylib (compatibility version 1.0.0)\n' if missing else '')
        return inventory, addon, Fake()

    def test_r204_native_inventory_selection_null(self):
        inventory, addon, commands = self.native_inventory()
        result = c.inventory_native_closure(inventory, [], [], commands, self.root, None)
        self.assertEqual([row['image']['path'] for row in result['images']], [str(addon)])

    def test_r204_native_inventory_selection_addon_failure_caught(self):
        inventory, _, commands = self.native_inventory(missing=True)
        with self.assertRaisesRegex(ValueError, 'missing native dependency'):
            c.inventory_native_closure(inventory, [], [], commands, self.root, None)

    def test_r204_attest_producer_wires_helpers_and_review_inputs(self):
        """Fake only external metadata; execute the real attestation orchestration."""
        root, out, _, record, receipt = self.authority()
        study = root/'evaluation/x1-triage-ranking'
        runtime_root = self.root/'runtime'; tools = runtime_root/'.tools'
        scratch = self.root/'scratch'
        config_path = self.file('config.json')
        clock = self.file('clock.json')
        self.json(str((out/'input-verification.json').relative_to(self.root)), {'deferred_to_b': []})
        patch_file = self.file('runtime/scripts/patch', b'patch')
        self.file('checkout/scripts/patch', b'patch')
        git = self.file('keg/bin/git'); self.file('keg/libexec/git-core/helper')
        git_identity = dict(c.pin(git), version='git version 2.56.0')
        go = self.file('go/bin/go')
        python = self.file('runtime/.tools/python/python')
        self.file('runtime/.tools/native.node', self.thin(6)).chmod(0o644)
        gitlink = 'a'*40
        claw = tools/'clawpatch'/gitlink
        for directory in ('dist', 'node_modules'):
            (claw/directory).mkdir(parents=True, exist_ok=True)
        self.json('runtime/.tools/clawpatch/'+gitlink+'/package.json', {'name': 'claw', 'version': '0.7.3'})
        self.file('runtime/.tools/clawpatch/'+gitlink+'/pnpm-lock.yaml', b'lock')
        self.file('runtime/.tools/clawpatch/'+gitlink+'/pnpm-workspace.yaml', b'workspace')
        self.file('runtime/.tools/clawpatch/'+gitlink+'/.scout-review-safety', ('scout-review-sanitize-v3 '+c.digest(patch_file)).encode())
        versions = {'GO_VERSION': '1.26.10', 'PYTHON_VERSION': '3.11.11', 'NODE_VERSION': '22.18.0', 'UV_VERSION': '0.13.5', 'GITLEAKS_VERSION': '8.30.1', 'CLAWPATCH_VERSION': '0.7.3',
                    'SEMGREP_VERSION': '1.100.0', 'RUFF_VERSION': '0.15.8', 'LIZARD_VERSION': '1.21.7', 'MYPY_VERSION': '2.4.1', 'PNPM_VERSION': '11.25.0', 'JSCPD_VERSION': '4.0.8', 'PRETTIER_VERSION': '3.6.2', 'GOVULNCHECK_VERSION': '1.1.4'}
        for package, version in [('pnpm', '11.25.0'), ('jscpd', '4.0.8'), ('prettier', '3.6.2')]:
            self.json('runtime/.tools/node-tools/node_modules/'+package+'/package.json', {'version': version})
        claw_rows = [dict(c.pin(claw/name), path='third_party/clawpatch/'+name) for name in ('package.json', 'pnpm-lock.yaml', 'pnpm-workspace.yaml')]
        source = dict(c.pin(patch_file), path='scripts/patch')
        pins_value = {'source': {'files': [source]}, 'submodule': {'path': 'third_party/clawpatch', 'gitlink': gitlink, 'working_files': claw_rows, 'patch': 'scripts/patch'},
                      'toolchain': {'installed_entrypoints': [], 'additional_assets': [], 'requested_versions': versions}}
        pins_path = self.json('checkout/evaluation/x1-triage-ranking/preparation/pins.json', pins_value)
        runtime_path = self.json('checkout/evaluation/x1-triage-ranking/preparation/runtime-attestation-rev6.json', {'go': dict(c.pin(go), binary=str(go), version='go version go1.26.10 darwin/arm64')})
        journal_path = self.file('journal.start.json')
        journal = SimpleNamespace(id='fake', started_at='2026-11-09T00:00:00Z', start_path=journal_path)
        config = {'paths': {key: str(value) for key, value in {'ROOT': root, 'OUT': out, 'STUDY': study, 'SCRATCH': scratch}.items()}, 'deferred_to_b_contract': {'working_file_count': 176}}
        args = SimpleNamespace(runtime_root=str(runtime_root), module_cache=str(scratch/'mod'), pins=str(pins_path), config=str(config_path), clock=str(clock), external_root=[], dyld_cache_root=None, preflight=False)
        original_checked = c.checked
        def checked(path, expected):
            return c.pin(path) if Path(path) in (pins_path, runtime_path) else original_checked(path, expected)
        git_commands, _ = self.git_commands(head=gitlink)
        def command(argv, cwd):
            text = [str(value) for value in argv]
            if text[0] == str(git): return git_commands.run(argv, cwd)
            if '-m' in text: return '\tmod\tgolang.org/x/vuln\tv1.1.4\n'
            if text[-1] == '--version': return 'v22.18.0' if text[0].endswith('node') else 'uv 0.13.5'
            if text[-1] == 'version': return '8.30.1'
            return 'synthetic macOS'
        metadata = {'version': 'go version go1.26.10 darwin/arm64', 'selected_files': []}
        py_packages = {name: {'version': versions[key]} for name, key in [('semgrep', 'SEMGREP_VERSION'), ('ruff', 'RUFF_VERSION'), ('lizard', 'LIZARD_VERSION'), ('mypy', 'MYPY_VERSION')]}
        with ExitStack() as stack:
            stack.enter_context(patch.object(c.platform, 'system', return_value='Darwin'))
            stack.enter_context(patch.object(c.platform, 'machine', return_value='arm64'))
            stack.enter_context(patch.object(c.platform, 'python_version', return_value='3.11.11'))
            stack.enter_context(patch.object(c.sys, 'base_prefix', str(python.parent)))
            stack.enter_context(patch.object(c, 'checked', side_effect=checked))
            stack.enter_context(patch.object(c, 'runtime_python', return_value=c.pin(python)))
            stack.enter_context(patch.object(c, 'verify_deferred', return_value=([], [])))
            stack.enter_context(patch.object(c, 'object_git', return_value=(git_identity, git.parent.parent, [])))
            stack.enter_context(patch.object(c, 'Commands', return_value=SimpleNamespace(run=command, records=[])))
            stack.enter_context(patch.object(c, 'go_metadata', return_value=(metadata, [])))
            stack.enter_context(patch.object(c, 'python_packages', return_value=py_packages))
            stack.enter_context(patch.object(c, 'npm_packages', return_value={}))
            stack.enter_context(patch.object(c, 'pnpm_packages', return_value={}))
            stack.enter_context(patch.object(c, 'native_closure', return_value={'images': [], 'cache_files': []}))
            stack.enter_context(patch.object(c, 'retain', return_value=[]))
            git_check = stack.enter_context(patch.object(c, 'verify_git_runtime', wraps=c.verify_git_runtime))
            claw_check = stack.enter_context(patch.object(c, 'claw_runtime', wraps=c.claw_runtime))
            inventory_check = stack.enter_context(patch.object(c, 'runtime_inventory', wraps=c.runtime_inventory))
            native_check = stack.enter_context(patch.object(c, 'inventory_native_closure', wraps=c.inventory_native_closure))
            result = c.attest(args, config, journal)
        git_check.assert_called_once(); claw_check.assert_called_once(); inventory_check.assert_called_once(); native_check.assert_called_once()
        self.assertEqual(result['git'], git_identity)
        self.assertIn(c.pin(receipt), result['inputs']); self.assertIn(c.pin(record), result['inputs'])
        self.assertIn(str(git.parent.parent), [row['root'] for row in result['enumeration']])
        self.assertIn(str(tools/'native.node'), native_check.call_args.args[0].files)


if __name__ == '__main__':
    unittest.main()

"""A-prime weak-link contracts: synthetic decisions plus frozen-host conformance."""
from contextlib import ExitStack, contextmanager, nullcontext
from collections import Counter
import copy
import ctypes
import importlib.util
import json
from pathlib import Path
import platform
import re
import struct
import subprocess
import sys
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import Mock, patch
import uuid

EXEC = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(EXEC))
sys.path.insert(0, str(Path(__file__).resolve().parent))
import attest_closure as base
from closure import native as original
from closure import native_v2 as native
_ORIGINAL_NATIVE, _ORIGINAL_SOURCES = base.native_closure, base.own_sources
import attest_closure_v2 as entry
import preparation
import test_closure_review as prior
import test_closure_round2 as round_two

REMOTE = '/System/Library/PrivateFrameworks/RemoteProcessingBlock.framework/Versions/A/RemoteProcessingBlock'
CORE_AUDIO = '/System/Library/Frameworks/CoreAudio.framework/Versions/A/CoreAudio'
DSP = '/System/Library/PrivateFrameworks/AudioDSPGraph.framework/Versions/A/AudioDSPGraph'
TCC = '/System/Library/PrivateFrameworks/TCC.framework/Versions/A/TCC'
WATCHDOG = '/System/Library/PrivateFrameworks/WatchdogClient.framework/Versions/A/WatchdogClient'
CACHE_ROOT = Path('/System/Volumes/Preboot/Cryptexes/OS/System/Library/dyld')
IMAGE_UUID = '00000000-0000-0000-0000-000000000001'
SYSTEM_ALTERNATES = tuple('/usr/lib/system/libsystem_'+name+'.dylib' for name in ('kernel', 'platform', 'pthread'))
REAL_RECORD_ROOT = Path('/Users/adamjackson/Projects/work/x1-preparation-scratch/closure-remedy-checks/r2-impl3')
NULL_SEED = Path('/Users/adamjackson/.local/share/mise/installs/go/1.26.10/src/runtime/race/race_darwin_arm64.syso')
NULL_SEED_SHA256 = '6ca6a32e8b650ac03b6a11b40c7a0cdd938d00046bd1c942f7c530e0fdaf23b2'


def linked(name, edges=()):
    return (name+' [arm64e]:\n    -uuid:\n        '+IMAGE_UUID
            +'\n    -linked_dylibs:\n        attributes     load path\n'
            + ''.join('        '+' '.join(attributes)+'    '+target+'\n'
                      for target, attributes in edges))


def without_v2(value):
    result = copy.deepcopy(value)
    result.pop('format', None)
    result.pop('absent_weak_links', None)
    for row in result['cached_images']:
        row.pop('linked_dylibs', None)
        row.pop('cache_files', None)
    for image in result['images']:
        for row in image['dependencies']:
            row.pop('load_commands', None)
    return result


class FakeCommands:
    def __init__(self, texts=None, loads=None, details=None):
        self.texts, self.loads, self.details = texts or {}, loads or {}, details or {}
        self.calls = []

    def run(self, argv, cwd):
        self.calls.append([str(item) for item in argv])
        name = str(argv[-1])
        if '-linked_dylibs' in argv:
            return self.texts.get(name, '')
        if '-L' in argv:
            return self.loads.get(name, name+':\n')
        return self.details.get(name, '')


class WeakLinkTests(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory()
        self.addCleanup(temporary.cleanup)
        self.root = Path(temporary.name).resolve()
        self.cache = self.root/'cache'
        self.cache.mkdir()
        (self.cache/'dyld_shared_cache_test').write_bytes(b'dyld_v1  arm64e\0'+b'\0'*72+b'\1'*16)
        self.target = '/System/Library/X1Absent.dylib'
        self.a, self.b = '/usr/lib/X1A.dylib', '/usr/lib/X1B.dylib'

    @contextmanager
    def boundaries(self, texts, entries=()):
        def fill(buffer):
            for index in range(16):
                buffer[index] = 1
            return True
        library = SimpleNamespace(_dyld_get_shared_cache_uuid=Mock(side_effect=fill))
        with ExitStack() as stack:
            stack.enter_context(patch.object(native.ctypes, 'CDLL', return_value=library))
            stack.enter_context(patch.object(native, 'shared_cache_contains', side_effect=lambda name: name in texts))
            stack.enter_context(patch.object(native, 'overlay_roots', return_value=[Path('/overlay')]))
            stack.enter_context(patch.object(native, 'filesystem_entries', return_value=list(entries)))
            yield

    def seeded(self, edges, texts=None, entries=(), cache=True):
        texts = texts or {}
        commands = FakeCommands(texts)
        with self.boundaries(texts, entries):
            value = native.collect_seed_edges(edges, commands, self.root, self.cache if cache else None,
                                              base.safe_resolve, base.pin)
        return value, commands

    def test_s1_absent_weak_evidence_and_probe_memoization(self):
        edges = [(self.a, self.target, True, {'attributes': ['weak-link']}),
                 (self.b, self.target, True, {'attributes': ['delay-init', 'weak-link']})]
        value, commands = self.seeded(edges)
        absent = value['absent_weak_links'][0]
        self.assertEqual(absent['install_name'], self.target)
        self.assertEqual(absent['linked_from'], [dict(source=source, **provenance) for source, _, _, provenance in edges])
        self.assertEqual(absent['evidence'], {'shared_cache_contains_path': False, 'filesystem_entries': [],
                         'probed_paths': [self.target, '/overlay'+self.target], 'dyld_info_stdout_empty': True})
        self.assertEqual(sum('-linked_dylibs' in argv for argv in commands.calls), 1)
        self.assertNotIn('path', absent)

    def test_s2_required_absence_refuses(self):
        with self.assertRaisesRegex(ValueError, 'required native dependency absent'):
            self.seeded([(self.a, self.target, False, {'attributes': []})])

    def test_s3_delay_init_is_not_weak(self):
        text = linked(self.a, [(self.target, ('delay-init',))])
        with self.assertRaisesRegex(ValueError, 'required native dependency absent'):
            self.seeded([('seed', self.a, False, {})], {self.a: text})

    def test_s4_reexport_weak_is_required(self):
        text = linked(self.a, [(self.target, ('re-export', 'weak-link'))])
        with self.assertRaisesRegex(ValueError, 'required native dependency absent'):
            self.seeded([('seed', self.a, False, {})], {self.a: text})

    def test_s5_present_cache_image_parse_errors_never_become_absent(self):
        for text in ('', 'garbage', linked(self.target)+'        '+IMAGE_UUID+'\n'):
            with self.subTest(text=text), self.assertRaisesRegex(ValueError, 'ambiguous cached native image identity'):
                self.seeded([(self.a, self.target, True, {})], {self.target: text})

    def test_s6_disk_or_overlay_entry_prevents_absence(self):
        for name in (self.target, '/overlay'+self.target):
            with self.subTest(name=name), self.assertRaisesRegex(ValueError, 'not provably absent'):
                self.seeded([(self.a, self.target, True, {})], entries=[name])

    def test_s7_nonempty_probe_stdout_prevents_absence(self):
        commands = FakeCommands({self.target: 'uninspectable'})
        with self.boundaries({}), self.assertRaisesRegex(ValueError, 'not provably absent'):
            native.collect_seed_edges([(self.a, self.target, True, {})], commands, self.root,
                                      self.cache, base.safe_resolve, base.pin)

    def test_s8_required_edge_refuses_in_both_visit_orders(self):
        weak = (self.a, self.target, True, {'attributes': ['weak-link']})
        required = (self.b, self.target, False, {'attributes': []})
        for edges in ((weak, required), (required, weak)):
            with self.subTest(edges=edges), self.assertRaisesRegex(ValueError, 'required native dependency absent'):
                self.seeded(edges)

    def test_s8_duplicate_target_strength_is_required(self):
        text = linked(self.a, [(self.target, ('weak-link',)), (self.target, ())])
        with self.assertRaisesRegex(ValueError, 'required native dependency absent'):
            self.seeded([('seed', self.a, False, {})], {self.a: text})

    def test_s9_parser_refuses_unknown_repeated_and_malformed_grammar(self):
        valid = linked(self.a, [(self.target, ('weak-link',))])
        variants = [valid.replace('weak-link', 'mystery'), valid.replace('weak-link', 'weak-link weak-link'),
                    valid+'    -linked_dylibs:\n', valid+'    -surprise:\n',
                    valid.replace('attributes     load path', 'attributes load paths'),
                    valid+linked(self.a), valid.replace(self.target, '@rpath/nope')]
        for text in variants:
            with self.subTest(text=text), self.assertRaises(ValueError):
                native.parse_linked_dylibs(self.a, text)

    def test_s9_rpaths_end_links_and_iossupport_path_stays_whole(self):
        target = '/System/iOSSupport/System/Library/Frameworks/Example.framework/Example'
        text = linked(self.a, [(target, ())])+'    -rpaths:\n        @loader_path/private\n'
        identity, edges = native.parse_linked_dylibs(self.a, text)
        self.assertEqual(identity, IMAGE_UUID)
        self.assertEqual(edges, [(target, ())])

    def disk_fixture(self, dependency, commands):
        source = self.root/'executable'
        source.write_bytes(struct.pack('<8I', 0xfeedfacf, 0x100000c, 0, 2, 0, 0, 0, 0))
        loads = str(source)+':\n\t'+dependency+' (compatibility version 1.0.0)\n'
        details = ''.join('cmd '+command+'\ncmdsize 48\nname '+dependency+' (offset 24)\n' for command in commands)
        return source, FakeCommands(loads={str(source): loads}, details={str(source): details})

    def test_s10_disk_weak_absence(self):
        source, commands = self.disk_fixture(self.target, ['LC_LOAD_WEAK_DYLIB'])
        with self.boundaries({}):
            value = native.collect([source], commands, self.root, self.cache, base.safe_resolve, base.pin)
        self.assertEqual(value['images'][0]['dependencies'][0]['resolution'], 'weak-absent')
        self.assertEqual(value['absent_weak_links'][0]['linked_from'],
                         [{'source': str(source), 'load_commands': ['LC_LOAD_WEAK_DYLIB']}])

    def test_s10_disk_required_and_mixed_commands_refuse(self):
        for commands in (['LC_LOAD_DYLIB'], ['LC_LOAD_DYLIB', 'LC_LOAD_WEAK_DYLIB'], []):
            source, fake = self.disk_fixture(self.target, commands)
            with self.subTest(commands=commands), self.boundaries({}), self.assertRaisesRegex(ValueError, 'required native dependency absent'):
                native.collect([source], fake, self.root, self.cache, base.safe_resolve, base.pin)

    def test_s10_unresolved_rpath_or_non_system_weak_refuses(self):
        for target in ('@rpath/X1.dylib', '/opt/X1Absent.dylib'):
            source, fake = self.disk_fixture(target, ['LC_LOAD_WEAK_DYLIB'])
            with self.subTest(target=target), self.boundaries({}), self.assertRaises(ValueError):
                native.collect([source], fake, self.root, self.cache, base.safe_resolve, base.pin)

    def test_s11_cached_chain_of_5000_is_iterative(self):
        names = ['/usr/lib/X1Chain'+str(index)+'.dylib' for index in range(5000)]
        texts = {name: linked(name, [(names[index+1], ())] if index+1 < len(names) else ())
                 for index, name in enumerate(names)}
        value, _ = self.seeded([('seed', names[0], False, {})], texts)
        self.assertEqual(len(value['cached_images']), 5000)
        self.assertEqual(value['absent_weak_links'], [])
        self.assertTrue(all('cache_files' not in row for row in value['cached_images']))

    def test_i7_present_system_cache_member_uses_active_cache_not_disk_alternate(self):
        is_file = Path.is_file
        texts = {self.a: linked(self.a, [(self.b, ())]), self.b: linked(self.b)}
        with patch.object(Path, 'is_file', lambda path: str(path) in texts or is_file(path)), \
                patch.object(native, 'image_filetype', side_effect=AssertionError('disk alternate classified')), \
                patch.object(native, 'loaded_cache_image_uuid', return_value=IMAGE_UUID):
            value, commands = self.seeded([('seed', self.a, False, {})], texts)
        self.assertEqual(value['images'], [])
        self.assertEqual([row['install_name'] for row in value['cached_images']], [self.a, self.b])
        self.assertEqual(commands.calls, [['/usr/bin/dyld_info', '-arch', 'arm64e', '-uuid', '-linked_dylibs', name]
                                         for name in (self.a, self.b)])
        for row in value['cached_images']:
            self.assertEqual(row['identity_check'], {'disk_alternate_architecture': 'arm64e',
                             'loaded_cache_image_uuid': IMAGE_UUID, 'loaded_from_active_cache_range': True})

    def test_i7_selected_architecture_identity_must_be_unique_and_present(self):
        is_file = Path.is_file
        for text in ('', linked(self.a)+linked(self.a)):
            commands = FakeCommands({self.a: text})
            with self.subTest(text=text), self.boundaries(commands.texts), \
                    patch.object(Path, 'is_file', lambda path: str(path) == self.a or is_file(path)), \
                    self.assertRaisesRegex(ValueError, 'ambiguous cached native image identity'):
                native.collect_seed_edges([('seed', self.a, False, {})], commands, self.root,
                                          self.cache, base.safe_resolve, base.pin)

    def test_i7_unidentifiable_cache_architecture_refuses_without_guessing(self):
        (self.cache/'dyld_shared_cache_test').write_bytes(b'dyld_v1'+b'\0'*81+b'\1'*16)
        is_file = Path.is_file
        with patch.object(Path, 'is_file', lambda path: str(path) == self.a or is_file(path)), \
                self.assertRaisesRegex(ValueError, 'cannot identify active dyld cache architecture'):
            self.seeded([('seed', self.a, False, {})], {self.a: linked(self.a)})
    def test_i7_loaded_uuid_mismatch_refuses(self):
        is_file = Path.is_file
        with patch.object(Path, 'is_file', lambda path: str(path) == self.a or is_file(path)), \
                patch.object(native, 'loaded_cache_image_uuid', return_value='FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF'), \
                self.assertRaisesRegex(ValueError, 'on-disk alternate differs from active-cache image'):
            self.seeded([('seed', self.a, False, {})], {self.a: linked(self.a)})

    def test_i7_loaded_uuid_helper_failure_propagates(self):
        is_file = Path.is_file
        failure = 'cannot establish active-cache identity: '+self.a
        with patch.object(Path, 'is_file', lambda path: str(path) == self.a or is_file(path)), \
                patch.object(native, 'loaded_cache_image_uuid', side_effect=ValueError(failure)), \
                self.assertRaisesRegex(ValueError, '^'+re.escape(failure)+'$'):
            self.seeded([('seed', self.a, False, {})], {self.a: linked(self.a)})

    def test_i7_uuid_identity_comparison_is_case_insensitive(self):
        identity = 'AbCd1234-5e6F-7890-aB12-cD34eF567890'
        is_file = Path.is_file
        with patch.object(Path, 'is_file', lambda path: str(path) == self.a or is_file(path)), \
                patch.object(native, 'loaded_cache_image_uuid', return_value=identity.lower()):
            value, _ = self.seeded([('seed', self.a, False, {})],
                                   {self.a: linked(self.a).replace(IMAGE_UUID, identity.upper())})
        self.assertEqual(value['cached_images'][0]['identity_check']['loaded_cache_image_uuid'], identity.lower())

    def test_i7_absent_disk_cache_member_is_queried_without_architecture(self):
        value, commands = self.seeded([('seed', self.a, False, {})], {self.a: linked(self.a)})
        self.assertNotIn('identity_check', value['cached_images'][0])
        self.assertEqual(commands.calls, [['/usr/bin/dyld_info', '-uuid', '-linked_dylibs', self.a]])

    def test_i7_present_nonmember_stays_on_disk(self):
        disk = self.root/'disk.dylib'
        disk.write_bytes(struct.pack('<8I', 0xfeedfacf, 0x100000c, 0, 6, 0, 0, 0, 0))
        texts = {self.a: linked(self.a, [(self.target, ())])}
        commands = FakeCommands(texts)
        is_file = Path.is_file
        def resolve(path):
            return (disk, []) if str(path) == self.target else base.safe_resolve(path)
        with self.boundaries(texts), \
                patch.object(Path, 'is_file', lambda path: str(path) == self.target or is_file(path)), \
                patch.object(native, 'image_filetype', wraps=native.image_filetype) as classify:
            value = native.collect_seed_edges([('seed', self.a, False, {})], commands, self.root,
                                              self.cache, resolve, base.pin)
        classify.assert_any_call(disk)
        self.assertEqual(value['images'][0]['image'], base.pin(disk))
        self.assertEqual([row['install_name'] for row in value['cached_images']], [self.a])
        self.assertFalse(any('-linked_dylibs' in argv and argv[-1] == self.target for argv in commands.calls))


    def test_s12_executable_context_null_matches_v1(self):
        fixture = prior.ClosureReviewTests()
        fixture.setUp()
        self.addCleanup(fixture.doCleanups)
        exe = fixture.macho('python/bin/python3.11', 2)
        lib = fixture.macho('python/lib/libpython3.11.dylib', 6)
        leaf = fixture.macho('python/lib/leaf.dylib', 6)
        class Commands:
            def run(_, argv, cwd):
                image = Path(argv[-1])
                if '-L' in argv:
                    dep = {exe: '@executable_path/../lib/libpython3.11.dylib', lib: '@rpath/leaf.dylib', leaf: None}[image]
                    return str(image)+':\n'+('\t'+dep+' (compatibility version 1.0.0)\n' if dep else '')
                return 'cmd LC_RPATH\ncmdsize 48\npath @executable_path/../lib (offset 12)\n' if image == lib else ''
        args = ([exe, lib, leaf], Commands(), fixture.root, None, base.safe_resolve, base.pin)
        self.assertEqual(without_v2(native.collect(*args)), without_v2(original.collect(*args)))

    def test_s12_cached_null_matches_v1_and_visit_order(self):
        texts = {self.a: linked(self.a, [(self.b, ())]), self.b: linked(self.b)}
        source, commands = self.disk_fixture(self.a, ['LC_LOAD_DYLIB'])
        commands.texts = texts
        with self.boundaries(texts):
            one = original.collect([source], commands, self.root, self.cache, base.safe_resolve, base.pin)
            commands.calls.clear()
            two = native.collect([source], commands, self.root, self.cache, base.safe_resolve, base.pin)
        self.assertEqual(without_v2(one), without_v2(two))
        self.assertEqual([argv[-1] for argv in commands.calls if '-linked_dylibs' in argv], [self.a, self.b])

    def test_s13_signed_consumers_accept_absent_rows(self):
        value, _ = self.seeded([(self.a, self.target, True, {'attributes': ['weak-link']})])
        pins = preparation.verify_tree_pins(value)
        self.assertEqual(pins, value['cache_files'])
        fixture = prior.ClosureReviewTests();fixture.setUp();self.addCleanup(fixture.doCleanups)
        path, _, _, _ = fixture.recheck_fixture()
        closure = json.loads(path.read_bytes());closure['native_dependencies'] = value
        path.write_text(json.dumps(closure))
        self.assertEqual(base.verify_installed_closure(path)['native_dependencies'], value)

    def test_s14_absence_requires_a_cache_family(self):
        with self.assertRaisesRegex(ValueError, 'unresolvable native runtime dependency'):
            self.seeded([(self.a, self.target, True, {})], cache=False)

    def test_membership_symbol_missing_refuses(self):
        with patch.object(native.ctypes, 'CDLL', return_value=SimpleNamespace()), self.assertRaisesRegex(ValueError, 'cannot query active'):
            native.shared_cache_contains(self.target)

    def test_filesystem_permission_failure_refuses(self):
        with patch.object(native, 'overlay_roots', return_value=[]), patch.object(native.os, 'lstat', side_effect=PermissionError()), self.assertRaisesRegex(ValueError, 'cannot establish native absence'):
            native.filesystem_entries(self.target)

    def test_overlay_permission_failure_refuses(self):
        with patch.object(native.os, 'scandir', side_effect=PermissionError()), self.assertRaisesRegex(ValueError, 'cannot establish native absence'):
            native.overlay_roots()


class LoadedCacheIdentityTests(unittest.TestCase):
    name = '/usr/lib/system/X1Identity.dylib'

    def query(self, commands=None, names=None, outside=False, below=False,
              magic=0xfeedfacf, ncmds=None, sizeofcmds=None, identity=IMAGE_UUID):
        if commands is None:
            commands = struct.pack('<2I', 0x1b, 24)+bytes.fromhex(identity.replace('-', ''))
        header = struct.pack('<8I', magic, 0x100000c, 0, 6,
                             1 if ncmds is None else ncmds,
                             len(commands) if sizeofcmds is None else sizeofcmds, 0, 0)
        memory = ctypes.create_string_buffer(header+commands+(b'\0'*64 if below else b''))
        start, size = ctypes.addressof(memory), ctypes.sizeof(memory)
        range_shift = len(header+commands) if below else 0
        loaded_names = [self.name.encode()] if names is None else names
        def cache_range(pointer):
            ctypes.cast(pointer, ctypes.POINTER(ctypes.c_size_t)).contents.value = size-range_shift
            return start+range_shift
        library = SimpleNamespace(
            _dyld_image_count=Mock(return_value=len(loaded_names)),
            _dyld_get_image_name=Mock(side_effect=lambda index: loaded_names[index]),
            _dyld_get_image_header=Mock(return_value=start+size if outside else start),
            _dyld_get_shared_cache_range=Mock(side_effect=cache_range))
        with patch.object(native.ctypes, 'CDLL', return_value=library):
            return native.loaded_cache_image_uuid(self.name)

    def test_loaded_uuid_matches_memory(self):
        self.assertEqual(self.query(), IMAGE_UUID)

    def test_image_not_loaded_refuses(self):
        with self.assertRaisesRegex(ValueError, 'cannot establish active-cache identity'):
            self.query(names=[])

    def test_duplicate_loaded_name_refuses(self):
        with self.assertRaisesRegex(ValueError, 'cannot establish active-cache identity'):
            self.query(names=[self.name.encode(), self.name.encode()])

    def test_header_outside_cache_refuses(self):
        with self.assertRaisesRegex(ValueError, 'cannot establish active-cache identity'):
            self.query(outside=True)

    def test_readable_header_below_cache_range_refuses(self):
        with self.assertRaisesRegex(ValueError, 'cannot establish active-cache identity'):
            self.query(below=True)

    def test_ncmds_bound_refuses_before_command_memory_reads(self):
        with patch.object(native.ctypes, 'string_at', wraps=ctypes.string_at) as read_memory:
            with self.assertRaisesRegex(ValueError, 'cannot establish active-cache identity'):
                self.query(ncmds=4)
        self.assertEqual(read_memory.call_count, 1)

    def test_trailing_command_slack_refuses(self):
        command = struct.pack('<2I', 0x1b, 24)+bytes.fromhex(IMAGE_UUID.replace('-', ''))+b'\0'*8
        with self.assertRaisesRegex(ValueError, 'cannot establish active-cache identity'):
            self.query(commands=command)

    def test_in_bounds_oversized_uuid_command_refuses(self):
        command = struct.pack('<2I', 0x1b, 32)+bytes.fromhex(IMAGE_UUID.replace('-', ''))+b'\0'*8
        with self.assertRaisesRegex(ValueError, 'cannot establish active-cache identity'):
            self.query(commands=command)

    def test_non_symmetric_mixed_case_uuid_preserves_byte_order(self):
        identity = 'AbCd1234-5e6F-7890-aB12-cD34eF567890'
        self.assertEqual(self.query(identity=identity), identity.upper())

    def test_invalid_header_and_command_bounds_refuse(self):
        variants = [{'magic': 0}, {'ncmds': 10}, {'sizeofcmds': 1000},
                    {'commands': struct.pack('<2I', 0x1b, 4)},
                    {'commands': struct.pack('<2I', 0x1b, 128)},
                    {'commands': struct.pack('<2I', 1, 8)},
                    {'commands': (struct.pack('<2I', 0x1b, 24)+b'\0'*16)*2, 'ncmds': 2}]
        for variant in variants:
            with self.subTest(variant=variant), self.assertRaisesRegex(ValueError, 'cannot establish active-cache identity'):
                self.query(**variant)


class WrapperTests(unittest.TestCase):
    def test_w1_drift_refuses_before_rebinding(self):
        for failed in entry.SIGNED_BASE:
            def digest(path):
                return 'drift' if Path(path) == EXEC/failed else entry.SIGNED_BASE[str(Path(path).relative_to(EXEC))]
            with self.subTest(failed=failed), patch.object(base, 'digest', side_effect=digest):
                with self.assertRaisesRegex(ValueError, 'signed closure base drift'):
                    entry.install()
                self.assertIs(base.native_closure, _ORIGINAL_NATIVE)
                self.assertIs(base.own_sources, _ORIGINAL_SOURCES)

    def test_w2_exact_rebinding_and_real_inventory_route(self):
        with patch.object(base, 'native_closure'), patch.object(base, 'own_sources'), patch.object(native, 'collect', return_value={'observed': True}) as collect:
            entry.install()
            self.assertIs(base.native_closure, entry.native_closure)
            self.assertIs(base.own_sources, entry.own_sources)
            inventory = SimpleNamespace(files={'disk': {}})
            result = base.inventory_native_closure(inventory, ['os'], [], 'commands', EXEC, 'cache')
            self.assertEqual(result, {'observed': True})
            collect.assert_called_once_with(['disk', 'os'], 'commands', EXEC, 'cache', base.safe_resolve, base.pin)

    def test_w3_exact_source_order(self):
        expected = ['attest_closure.py', 'closure/native.py', 'preparation.py', 'verify_inputs.py',
                    'run-config.json', 'attest_closure_v2.py', 'closure/native_v2.py']
        self.assertEqual([Path(row['path']).relative_to(EXEC).as_posix() for row in entry.own_sources()], expected)

    def test_w4_main_binds_v2_journal_sources_and_argv(self):
        fixture = round_two.RoundTwoTests();fixture.setUp();self.addCleanup(fixture.doCleanups)
        root, out, _, _, _ = fixture.authority()
        config_path, pins_path, clock = (fixture.file(name) for name in ('config.json', 'pins.json', 'clock.json'))
        scratch = fixture.root/'scratch'
        config = {'paths': {'ROOT': str(root), 'OUT': str(out), 'SCRATCH': str(scratch)}}
        captured = {}
        class Journal:
            id = 'weaklink-synthetic'
            def __init__(_, *args, **kwargs):captured.update(kwargs);captured['argv'] = list(sys.argv)
            def finish(_, *args, **kwargs):pass
        argv = [str(EXEC/'attest_closure_v2.py'), '--config', str(config_path), '--clock', str(clock), '--pins', str(pins_path),
                '--preflight', '--runtime-root', str(fixture.root/'runtime'), '--module-cache', str(scratch/'mod'), '--dyld-cache-root', str(scratch/'dyld')]
        with ExitStack() as stack:
            stack.enter_context(patch.object(base, 'native_closure'))
            stack.enter_context(patch.object(base, 'own_sources'))
            stack.enter_context(patch.object(sys, 'argv', argv))
            stack.enter_context(patch.object(base, 'load_config', return_value=config))
            stack.enter_context(patch.object(base, 'read_clock'))
            stack.enter_context(patch.object(base, 'operation_lock', return_value=nullcontext()))
            stack.enter_context(patch.object(base, 'Journal', Journal))
            stack.enter_context(patch.object(base, 'attest', return_value={'synthetic_test_only': True}))
            entry.main()
        self.assertEqual(captured['argv'][0], str(EXEC/'attest_closure_v2.py'))
        self.assertIn(base.pin(EXEC/'attest_closure_v2.py'), captured['source_pins'])
        self.assertIn(base.pin(EXEC/'closure/native_v2.py'), captured['source_pins'])

    def test_w5_import_has_no_mutation_and_signed_source_assertion_passes(self):
        spec = importlib.util.spec_from_file_location('weaklink_import_probe', EXEC/'attest_closure_v2.py')
        module = importlib.util.module_from_spec(spec)
        before = (base.native_closure, base.own_sources)
        spec.loader.exec_module(module)
        self.assertEqual((base.native_closure, base.own_sources), before)
        prior.ClosureReviewTests().test_r108_own_closure_is_not_whole_execution_manifest()


class RealByteTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        if platform.system() != 'Darwin' or platform.machine() != 'arm64':
            raise unittest.SkipTest('real-byte contracts require Darwin arm64')
        result = subprocess.run(['/usr/bin/sw_vers', '-buildVersion'], capture_output=True, text=True, check=True)
        if result.stdout.strip() != '25G72':
            raise unittest.SkipTest('real-byte contracts require frozen build 25G72; got '+result.stdout.strip())

    def setUp(self):
        self.root = REAL_RECORD_ROOT/self._testMethodName/uuid.uuid4().hex
        self.root.mkdir(parents=True)
        self.commands = self.recording_commands('primary')

    def recording_commands(self, label):
        return base.Commands({'PATH': '/usr/bin:/bin', 'LANG': 'C', 'LC_ALL': 'C'}, self.root/label)

    def test_r1_coreaudio_weak_absence_and_present_weak_targets(self):
        result = native.collect_seed_edges([('real-required-seed', CORE_AUDIO, False, {'attributes': []})],
                                          self.commands, EXEC, CACHE_ROOT, base.safe_resolve, base.pin)
        absent = {row['install_name']: row for row in result['absent_weak_links']}
        parents = {row['source']: row['attributes'] for row in absent[REMOTE]['linked_from']}
        self.assertEqual(parents[CORE_AUDIO], ['delay-init', 'weak-link'])
        self.assertEqual(parents[DSP], ['weak-link'])
        cached = {row['install_name']: row for row in result['cached_images']}
        for target in (TCC, WATCHDOG, DSP):
            self.assertRegex(cached[target]['image_uuid'], r'^[0-9A-Fa-f-]{36}$')
        self.assert_identity_checks(cached)
        for row in absent.values():
            self.assertFalse(row['evidence']['shared_cache_contains_path'])
            self.assertEqual(row['evidence']['filesystem_entries'], [])
            self.assertTrue(row['evidence']['dyld_info_stdout_empty'])

    def test_r2_required_remote_processing_block_refuses(self):
        with self.assertRaisesRegex(ValueError, 'required native dependency absent'):
            native.collect_seed_edges([('real-required-seed', REMOTE, False, {'attributes': []})],
                                      self.commands, EXEC, CACHE_ROOT, base.safe_resolve, base.pin)

    def assert_identity_checks(self, cached):
        for target in SYSTEM_ALTERNATES:
            row = cached[target]
            self.assertEqual(row['identity_check']['disk_alternate_architecture'], 'arm64e')
            self.assertTrue(row['identity_check']['loaded_from_active_cache_range'])
            self.assertEqual(row['identity_check']['loaded_cache_image_uuid'].lower(), row['image_uuid'].lower())

    def test_r3a_signed_v1_bin_sh_refusal_remains_visible(self):
        with self.assertRaises(ValueError) as refusal:
            original.collect(['/bin/sh'], self.commands, EXEC, CACHE_ROOT, base.safe_resolve, base.pin)
        self.assertEqual(str(refusal.exception), 'ambiguous cached native image identity: '+REMOTE)

    def test_r3b_pinned_real_object_null_matches_v1(self):
        self.assertEqual(base.pin(NULL_SEED)['sha256'], NULL_SEED_SHA256)
        first, second = self.recording_commands('v1'), self.recording_commands('v2')
        one = original.collect([NULL_SEED], first, EXEC, CACHE_ROOT, base.safe_resolve, base.pin)
        two = native.collect([NULL_SEED], second, EXEC, CACHE_ROOT, base.safe_resolve, base.pin)
        self.assertEqual(without_v2(one), without_v2(two))
        self.assertEqual(len(two['images']), 1)
        self.assertEqual(two['images'][0]['dependencies'], [])
        self.assertEqual(two['cached_images'], [])
        self.assertEqual(two['absent_weak_links'], [])
        self.assertEqual(one['cache_files'], two['cache_files'])
        self.assertEqual(one['active_cache_uuid'], two['active_cache_uuid'])
        self.assertEqual([row['argv'] for row in first.records], [row['argv'] for row in second.records])

    def test_r3c_real_cached_prefix_matches_signed_v1(self):
        first, second = self.recording_commands('v1'), self.recording_commands('v2')
        try:
            original.collect(['/bin/sh'], first, EXEC, CACHE_ROOT, base.safe_resolve, base.pin)
        except ValueError as error:
            self.assertEqual(str(error), 'ambiguous cached native image identity: '+REMOTE)
            traceback = error.__traceback__
            prefix = None
            while traceback is not None:
                if traceback.tb_frame.f_code is original.collect.__code__:
                    prefix = copy.deepcopy(traceback.tb_frame.f_locals['cached'])
                traceback = traceback.tb_next
        else:
            self.fail('signed v1 unexpectedly completed')
        self.assertIsNotNone(prefix)
        self.assertEqual(len(prefix), sum('-linked_dylibs' in row['argv'] for row in first.records)-1)
        two = native.collect(['/bin/sh'], second, EXEC, CACHE_ROOT, base.safe_resolve, base.pin)
        cached = {row['install_name']: row for row in two['cached_images']}
        for name, row in prefix.items():
            target = name if name in cached else '/System/iOSSupport'+name
            self.assertIn(target, cached)
            self.assertEqual(row['image_uuid'], cached[target]['image_uuid'])
            self.assertEqual(row['dependency_install_names'], cached[target]['dependency_install_names'])
        def records(commands):
            return Counter((tuple(row['argv']), row['exit_code'], row['stdout']) for row in commands.records)
        self.assertFalse(records(first)-records(second))
        absent = {row['install_name']: row for row in two['absent_weak_links']}
        self.assertIn(REMOTE, absent)
        for target, row in absent.items():
            self.assertTrue(target.startswith(native.SYSTEM_PREFIXES))
            self.assertFalse(native.shared_cache_contains(target))
            self.assertEqual(native.filesystem_entries(target), [])
            self.assertFalse(row['evidence']['shared_cache_contains_path'])
            self.assertEqual(row['evidence']['filesystem_entries'], [])
            self.assertTrue(row['evidence']['dyld_info_stdout_empty'])
            self.assertEqual(row['evidence']['probed_paths'], native.probe_paths(target))
            probe_argv = ['/usr/bin/dyld_info', '-uuid', '-linked_dylibs', target]
            self.assertTrue(any(record['argv'] == probe_argv and record['exit_code'] == 0
                                and record['stdout'] == '' for record in second.records))
        remote_records = [row for row in second.records if row['argv'][-1] == REMOTE]
        self.assertTrue(remote_records)
        self.assertTrue(all(row['exit_code'] == 0 and row['stdout'] == '' for row in remote_records))
        self.assert_identity_checks(cached)


if __name__ == '__main__':
    unittest.main()

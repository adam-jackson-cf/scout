"""Append-only native closure: iterative cached edges and proved weak absence."""
import ctypes
import os
from pathlib import Path
import re
import struct
import uuid

from closure.native import image_filetype, resolve_install_name

FORMAT = 'scout.x1.native-dependencies.v2'
SYSTEM_PREFIXES = ('/usr/lib/', '/System/Library/')
LINK_ATTRIBUTES = {'upward', 'delay-init', 'weak-link', 're-export', 'lazy-load'}
LOAD_COMMANDS = ('LC_LOAD_DYLIB', 'LC_LOAD_WEAK_DYLIB', 'LC_REEXPORT_DYLIB',
                 'LC_LOAD_UPWARD_DYLIB', 'LC_LAZY_LOAD_DYLIB')
OVERLAY_PARENTS = ('/System/Cryptexes', '/System/Volumes/Preboot/Cryptexes')


def shared_cache_contains(name):
    query = getattr(ctypes.CDLL(None), '_dyld_shared_cache_contains_path', None)
    if query is None:
        raise ValueError('cannot query active dyld shared-cache membership')
    query.argtypes, query.restype = [ctypes.c_char_p], ctypes.c_bool
    return bool(query(name.encode()))

def loaded_cache_image_uuid(name):
    """Read one loaded image's LC_UUID, bounded by the active cache range."""
    failure = 'cannot establish active-cache identity: '+name
    try:
        library = ctypes.CDLL(None)
        count = library._dyld_image_count
        count.argtypes, count.restype = [], ctypes.c_uint32
        image_name = library._dyld_get_image_name
        image_name.argtypes, image_name.restype = [ctypes.c_uint32], ctypes.c_char_p
        image_header = library._dyld_get_image_header
        image_header.argtypes, image_header.restype = [ctypes.c_uint32], ctypes.c_void_p
        cache_range = library._dyld_get_shared_cache_range
        cache_range.argtypes, cache_range.restype = [ctypes.POINTER(ctypes.c_size_t)], ctypes.c_void_p
        matches = [index for index in range(count()) if image_name(index) == name.encode()]
        if len(matches) != 1:
            raise ValueError(failure)
        size = ctypes.c_size_t()
        start = cache_range(ctypes.byref(size))
        address = image_header(matches[0])
        if not start or not address or not start <= address <= start+size.value-32:
            raise ValueError(failure)
        header = struct.unpack('<8I', ctypes.string_at(address, 32))
        ncmds, sizeofcmds = header[4:6]
        if header[0] != 0xfeedfacf or sizeofcmds > start+size.value-address-32 or ncmds > sizeofcmds//8:
            raise ValueError(failure)
        offset, identities = 0, []
        for _ in range(ncmds):
            if offset+8 > sizeofcmds:
                raise ValueError(failure)
            command, command_size = struct.unpack('<2I', ctypes.string_at(address+32+offset, 8))
            if command_size < 8 or offset+command_size > sizeofcmds:
                raise ValueError(failure)
            if command == 0x1b:
                if command_size != 24:
                    raise ValueError(failure)
                identities.append(str(uuid.UUID(bytes=ctypes.string_at(address+32+offset+8, 16))).upper())
            offset += command_size
        if offset != sizeofcmds or len(identities) != 1:
            raise ValueError(failure)
        return identities[0]
    except (AttributeError, OSError, TypeError, ValueError, struct.error) as error:
        raise ValueError(failure) from error



def overlay_roots():
    result = []
    for parent in OVERLAY_PARENTS:
        try:
            with os.scandir(parent) as children:
                result.extend(Path(child.path) for child in children if child.is_dir())
        except FileNotFoundError:
            continue
        except OSError as error:
            raise ValueError('cannot establish native absence: '+parent) from error
    return sorted(result)


def probe_paths(name):
    return [name, *[str(root)+name for root in overlay_roots()]]


def filesystem_entries(name):
    entries = []
    for path in probe_paths(name):
        try:
            os.lstat(path)
        except (FileNotFoundError, NotADirectoryError):
            continue
        except OSError as error:
            raise ValueError('cannot establish native absence: '+path) from error
        entries.append(path)
    return entries


def parse_linked_dylibs(name, text):
    identities = re.findall(r'^\s+([0-9A-Fa-f]{8}(?:-[0-9A-Fa-f]{4}){3}-[0-9A-Fa-f]{12})\s*$', text, re.M)
    headers = re.findall(r'^\S.* \[[^\]]+\]:$', text, re.M)
    if not text.startswith(name+' [') or len(identities) != 1 or len(headers) != 1:
        raise ValueError('ambiguous cached native image identity: '+name)
    sections, current, column, edges = set(), None, False, []
    for line in text.splitlines():
        if re.fullmatch(r'    -[A-Za-z_]+:', line):
            section = line.strip()
            if section not in ('-uuid:', '-linked_dylibs:', '-rpaths:') or section in sections:
                raise ValueError('unexpected dyld_info section: '+line)
            sections.add(section)
            current, column = section, section == '-linked_dylibs:'
            continue
        if current != '-linked_dylibs:' or not line.strip():
            continue
        if column:
            if line.split() != ['attributes', 'load', 'path']:
                raise ValueError('unexpected dyld_info link header: '+line)
            column = False
            continue
        match = re.fullmatch(r'\s+((?:[a-z][a-z-]* +)*)(/.*\S)', line)
        if not match:
            raise ValueError('unresolved cached-image load command: '+line)
        attributes = tuple(match[1].split())
        if set(attributes)-LINK_ATTRIBUTES or len(set(attributes)) != len(attributes):
            raise ValueError('unknown cached-image link attribute: '+line)
        edges.append((match[2], attributes))
    if '-linked_dylibs:' not in sections or column:
        raise ValueError('unexpected dyld_info section: missing linked-dylibs header')
    return identities[0], edges


def effective_weak(attributes):
    return 'weak-link' in attributes and 're-export' not in attributes


def collect(files, commands, cwd, cache_root, safe_resolve, pin):
    return _collect(files, commands, cwd, cache_root, safe_resolve, pin, ())


def collect_seed_edges(edges, commands, cwd, cache_root, safe_resolve, pin):
    """Exercise cached roots as (source, target, weak, provenance) seed edges."""
    return _collect((), commands, cwd, cache_root, safe_resolve, pin, edges)


def _collect(files, commands, cwd, cache_root, safe_resolve, pin, seed_edges):
    cache_files, cache_uuid, cached, result = [], None, {}, []
    if cache_root:
        root, _ = safe_resolve(cache_root)
        getter = getattr(ctypes.CDLL(None), '_dyld_get_shared_cache_uuid', None)
        if getter is None:
            raise ValueError('cannot identify active dyld shared cache')
        buffer = (ctypes.c_ubyte * 16)()
        getter.argtypes, getter.restype = [ctypes.c_void_p], ctypes.c_bool
        if not getter(buffer):
            raise ValueError('active dyld shared cache UUID unavailable')
        active, matches = bytes(buffer), []
        for path in sorted(root.iterdir()):
            if not path.name.startswith('dyld_shared_cache_') or not path.is_file():
                continue
            with path.open('rb') as stream:
                header = stream.read(104)
            if header.startswith(b'dyld_v') and header[88:104] == active:
                matches.append(path)
        if len(matches) != 1:
            raise ValueError('provided cache root does not uniquely match active cache UUID')
        main = matches[0]
        cache_uuid = str(uuid.UUID(bytes=active))
        cache_files = [pin(path) for path in sorted(root.iterdir())
                       if path.is_file() and (path.name == main.name or path.name.startswith(main.name+'.'))]
        if not cache_files:
            raise ValueError('empty active dyld shared-cache family')

    filetypes, absent = {}, {}
    cache_architecture = None

    def filetype_of(path):
        if path not in filetypes:
            filetypes[path] = image_filetype(path)
        return filetypes[path]

    pending, seen, contextual_sources = [(Path(path), None) for path in files], set(), set()
    pending.sort(key=lambda item: filetype_of(item[0]) == 2)

    def edge(source, target, weak, provenance):
        nonlocal cache_architecture
        on_disk = Path(target).is_file()
        cache_member = bool(cache_files) and (not on_disk or target.startswith(SYSTEM_PREFIXES)) and shared_cache_contains(target)
        if on_disk and not cache_member:
            pending.append((Path(target), None))
            return 'disk', None
        if not cache_files:
            raise ValueError('unresolvable native runtime dependency: '+target)
        if cache_member:
            if target in cached:
                return 'cached', None
            architecture_args = ()
            if on_disk:
                # Present system alternates are fat; select the active cache's architecture.
                if cache_architecture is None:
                    with main.open('rb') as stream:
                        magic = stream.read(16)
                    match = re.fullmatch(rb'dyld_v\d+\s+([A-Za-z0-9_]+)\x00*', magic)
                    if match is None:
                        raise ValueError('cannot identify active dyld cache architecture')
                    cache_architecture = match[1].decode('ascii')
                architecture_args = ('-arch', cache_architecture)
            text = commands.run(['/usr/bin/dyld_info', *architecture_args, '-uuid', '-linked_dylibs', target], cwd)
            image_uuid, edges = parse_linked_dylibs(target, text)
            identity_check = None
            if on_disk:
                loaded_uuid = loaded_cache_image_uuid(target)
                if loaded_uuid.lower() != image_uuid.lower():
                    raise ValueError('on-disk alternate differs from active-cache image: '+target)
                identity_check = {'disk_alternate_architecture': cache_architecture,
                                  'loaded_cache_image_uuid': loaded_uuid,
                                  'loaded_from_active_cache_range': True}
            requirements = {}
            for dependency, attributes in edges:
                requirements[dependency] = requirements.get(dependency, True) and effective_weak(attributes)
            record = {'install_name': target, 'image_uuid': image_uuid, 'active_cache_uuid': cache_uuid,
                      'dependency_install_names': [name for name, _ in edges],
                      'linked_dylibs': [{'install_name': name, 'attributes': list(attributes),
                                        'requirement': 'weak' if requirements[name] else 'required'}
                                       for name, attributes in edges]}
            if identity_check is not None:
                record['identity_check'] = identity_check
            cached[target] = record
            remaining = iter((dependency, requirements[dependency], {'attributes': list(attributes)})
                             for dependency, attributes in edges)
            return 'cached', (record, remaining)
        if weak and target.startswith(SYSTEM_PREFIXES):
            if target not in absent:
                paths = probe_paths(target)
                if filesystem_entries(target) or commands.run(
                        ['/usr/bin/dyld_info', '-uuid', '-linked_dylibs', target], cwd) != '':
                    raise ValueError('weak native dependency not provably absent: '+source+' -> '+target)
                absent[target] = {'install_name': target, 'active_cache_uuid': cache_uuid,
                                  'linked_from': [], 'evidence': {'shared_cache_contains_path': False,
                                  'filesystem_entries': [], 'probed_paths': paths,
                                  'dyld_info_stdout_empty': True}}
            parent = dict(source=source, **provenance)
            if parent not in absent[target]['linked_from']:
                absent[target]['linked_from'].append(parent)
            return 'weak-absent', None
        raise ValueError('required native dependency absent: '+source+' -> '+target)

    def walk(source, target, weak, provenance):
        resolution, frame = edge(source, target, weak, provenance)
        stack = [frame] if frame is not None else []
        while stack:
            record, remaining = stack[-1]
            try:
                dependency, child_weak, child_provenance = next(remaining)
            except StopIteration:
                stack.pop()
                continue
            _, child = edge(record['install_name'], dependency, child_weak, child_provenance)
            if child is not None:
                stack.append(child)
        return resolution

    for source, target, weak, provenance in seed_edges:
        walk(source, target, weak, provenance)
    while pending:
        requested, executable = pending.pop()
        source, _ = safe_resolve(requested)
        filetype = filetype_of(source)
        if filetype is None:
            continue
        if filetype == 2:
            executable = source
        if executable is None and source in contextual_sources:
            continue
        if executable is not None:
            contextual_sources.add(source)
        identity = (source, executable)
        if identity in seen:
            continue
        seen.add(identity)
        loads = commands.run(['/usr/bin/otool', '-L', source], cwd)
        details = commands.run(['/usr/bin/otool', '-l', source], cwd)
        rpaths = re.findall(r'cmd LC_RPATH\s+cmdsize \d+\s+path (.+?) \(offset', details)
        ids = set(re.findall(r'cmd LC_ID_DYLIB\s+cmdsize \d+\s+name (.+?) \(offset', details))
        load_commands = {}
        for command, name in re.findall(r'cmd (LC_[A-Z_]+)\s+cmdsize \d+\s+name (.+?) \(offset', details):
            if command in LOAD_COMMANDS:
                load_commands.setdefault(name, set()).add(command)
        deps = []
        for line in loads.splitlines()[1:]:
            if not line.startswith('\t'):
                continue
            dependency = line.strip().split(' (compatibility version', 1)[0]
            if not dependency or dependency in ids:
                continue
            named_by = sorted(load_commands.get(dependency, set()))
            candidates = resolve_install_name(dependency, source, executable, rpaths)
            existing = {safe_resolve(path)[0] for path in candidates if path.is_file()}
            if len(existing) == 1:
                target = next(iter(existing))
                deps.append(dict(pin(target), install_name=dependency, load_commands=named_by))
                pending.append((target, executable))
            elif not existing and dependency.startswith('/'):
                resolution = walk(str(source), dependency, named_by == ['LC_LOAD_WEAK_DYLIB'],
                                  {'load_commands': named_by})
                row = {'install_name': dependency, 'active_cache_uuid': cache_uuid, 'load_commands': named_by}
                if resolution == 'weak-absent':
                    row['resolution'] = resolution
                deps.append(row)
            else:
                raise ValueError('ambiguous/missing native dependency: '+str(source)+' -> '+dependency)
        result.append({'image': pin(source), 'executable_context': str(executable) if executable else None,
                       'dependencies': deps, 'rpaths': rpaths})
    for record in absent.values():
        record['linked_from'].sort(key=lambda row: row['source'])
    return {'format': FORMAT, 'images': result,
            'cached_images': sorted(cached.values(), key=lambda row: row['install_name']),
            'active_cache_uuid': cache_uuid, 'cache_files': cache_files,
            'absent_weak_links': sorted(absent.values(), key=lambda row: row['install_name'])}

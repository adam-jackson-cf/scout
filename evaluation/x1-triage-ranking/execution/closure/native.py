"""Mach-O load graph and active dyld split-cache byte closure."""
import ctypes
from pathlib import Path
import re
import struct
import uuid


MAGIC = {b'\xcf\xfa\xed\xfe', b'\xfe\xed\xfa\xcf', b'\xce\xfa\xed\xfe', b'\xfe\xed\xfa\xce',
         b'\xca\xfe\xba\xbe', b'\xbe\xba\xfe\xca', b'\xca\xfe\xba\xbf', b'\xbf\xba\xfe\xca'}


def image_filetype(path):
    """Classify real Mach-O structure, distinguishing the shared Java magic."""
    with Path(path).open('rb') as stream:
        length = stream.seek(0, 2)
        stream.seek(0)
        magic = stream.read(4)
        if magic not in MAGIC:
            return None

        def thin(offset, size):
            stream.seek(offset)
            data = stream.read(min(size, 32))
            if len(data) < 4 or data[:4] not in MAGIC - {
                    b'\xca\xfe\xba\xbe', b'\xbe\xba\xfe\xca', b'\xca\xfe\xba\xbf', b'\xbf\xba\xfe\xca'}:
                raise ValueError('invalid Mach-O slice magic')
            endian = '<' if data[:4] in (b'\xcf\xfa\xed\xfe', b'\xce\xfa\xed\xfe') else '>'
            header_size = 32 if data[:4] in (b'\xcf\xfa\xed\xfe', b'\xfe\xed\xfa\xcf') else 28
            if len(data) < header_size:
                raise ValueError('truncated Mach-O header')
            cpu, _, kind, count, command_bytes = struct.unpack_from(endian+'IIIII', data, 4)
            if not 1 <= kind <= 12 or header_size+command_bytes > size or count*8 > command_bytes:
                raise ValueError('invalid Mach-O header bounds/filetype')
            return cpu, kind

        fat = magic in (b'\xca\xfe\xba\xbe', b'\xbe\xba\xfe\xca', b'\xca\xfe\xba\xbf', b'\xbf\xba\xfe\xca')
        if not fat:
            return thin(0, length)[1]
        version = stream.read(4)
        if len(version) != 4:
            raise ValueError('truncated fat Mach-O header')
        # CAFEBABE is shared with Java. Only dyld's file bounds can rule out
        # the fat interpretation: our stricter validation must not hide code.
        _, major = struct.unpack('>HH', version)
        if magic == b'\xca\xfe\xba\xbe' and 45 <= major <= 100:
            java_count = struct.unpack('>I', version)[0]
            if 8+java_count*20 > length:
                return None
            for _ in range(java_count):
                entry = stream.read(20)
                offset, size = struct.unpack_from('>II', entry, 8)
                if offset > length or size > length-offset:
                    return None
            raise ValueError('ambiguous Java/fat Mach-O image')
        endian = '>' if magic[0] == 0xca else '<'
        wide = magic in (b'\xca\xfe\xba\xbf', b'\xbf\xba\xfe\xca')
        count = struct.unpack(endian+'I', version)[0]
        if not 1 <= count <= 32:
            raise ValueError('invalid fat Mach-O architecture count')
        entry_size = 32 if wide else 20
        table_end = 8+count*entry_size
        if table_end > length:
            raise ValueError('truncated fat Mach-O header')
        table = stream.read(count*entry_size)
        ranges, host = [], []
        for index in range(count):
            entry = table[index*entry_size:(index+1)*entry_size]
            cpu = struct.unpack_from(endian+'I', entry)[0]
            offset, size, align = struct.unpack_from(endian+('QQI' if wide else 'III'), entry, 8)
            if wide and struct.unpack_from(endian+'I', entry, 28)[0]:
                raise ValueError('invalid fat Mach-O reserved field')
            if align > 63 or offset < table_end or not size or offset+size > length or offset % (1 << align):
                raise ValueError('invalid fat Mach-O slice bounds/alignment')
            actual_cpu, kind = thin(offset, size)
            if actual_cpu != cpu:
                raise ValueError('fat Mach-O slice CPU disagrees with entry')
            ranges.append((offset, offset+size))
            if cpu == 0x100000c:
                host.append(kind)
        if any(left[1] > right[0] for left, right in zip(sorted(ranges), sorted(ranges)[1:])):
            raise ValueError('overlapping fat Mach-O slices')
        if len(host) != 1:
            raise ValueError('Mach-O has no unique arm64 slice')
        return host[0]


def resolve_install_name(name, source, executable, rpaths):
    def expand(value):
        value = value.replace('@loader_path', str(Path(source).parent))
        if '@executable_path' in value:
            if executable is None:
                raise ValueError('native dependency requires unresolved executable context: '+name)
            value = value.replace('@executable_path', str(Path(executable).parent))
        return Path(value)
    if name.startswith('@rpath/'):
        return [expand(path)/name[len('@rpath/'):] for path in rpaths]
    if name.startswith(('@loader_path/', '@executable_path/', '/')):
        return [expand(name)]
    raise ValueError('unresolved native install name: '+name)


def collect(files, commands, cwd, cache_root, safe_resolve, pin):
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

    filetypes = {}
    def filetype_of(path):
        if path not in filetypes:
            filetypes[path] = image_filetype(path)
        return filetypes[path]
    pending, seen, contextual_sources = [(Path(path), None) for path in files], set(), set()
    # Executables seed contexts before contextless initial library entries.
    pending.sort(key=lambda item: filetype_of(item[0]) == 2)

    def cached_image(name):
        if name in cached:
            return
        if not cache_files or not name.startswith(('/usr/lib/', '/System/Library/')):
            raise ValueError('unresolvable native runtime dependency: ' + name)
        text = commands.run(['/usr/bin/dyld_info', '-uuid', '-linked_dylibs', name], cwd)
        identities = re.findall(r'^\s+([0-9A-Fa-f]{8}(?:-[0-9A-Fa-f]{4}){3}-[0-9A-Fa-f]{12})\s*$', text, re.M)
        if len(identities) != 1 or not text.startswith(name+' ['):
            raise ValueError('ambiguous cached native image identity: ' + name)
        dependencies, in_links = [], False
        for line in text.splitlines():
            if line.strip() == '-linked_dylibs:':
                in_links = True
                continue
            if not in_links or not line.strip() or 'load path' in line:
                continue
            match = re.search(r'(/(?:usr/lib|System/Library)/.+)$', line)
            if not match:
                raise ValueError('unresolved cached-image load command: ' + line)
            dependencies.append(match[1])
        cached[name] = {'install_name': name, 'image_uuid': identities[0],
                        'active_cache_uuid': cache_uuid, 'cache_files': cache_files,
                        'dependency_install_names': dependencies}
        for dependency in dependencies:
            if Path(dependency).is_file():
                pending.append((Path(dependency), None))
            else:
                cached_image(dependency)

    while pending:
        requested, executable = pending.pop()
        source, _ = safe_resolve(requested)
        filetype = filetype_of(source)
        if filetype is None:
            continue
        if filetype == 2:  # MH_EXECUTE supplies its own executable context.
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
        deps = []
        # otool -L includes a dylib's own LC_ID_DYLIB: omit that identity,
        # using the LC_ID_DYLIB load command rather than guessing its spelling.
        ids = set(re.findall(r'cmd LC_ID_DYLIB\s+cmdsize \d+\s+name (.+?) \(offset', details))
        for line in loads.splitlines()[1:]:
            if not line.startswith('\t'):
                continue
            dependency = line.strip().split(' (compatibility version', 1)[0]
            if not dependency or dependency in ids:
                continue
            candidates = resolve_install_name(dependency, source, executable, rpaths)
            existing = {safe_resolve(p)[0] for p in candidates if p.is_file()}
            if not existing and dependency.startswith(('/usr/lib/', '/System/Library/')):
                cached_image(dependency)
                deps.append({'install_name': dependency, 'active_cache_uuid': cache_uuid})
                continue
            if len(existing) != 1:
                raise ValueError('ambiguous/missing native dependency: ' + str(source) + ' -> ' + dependency)
            target = next(iter(existing))
            deps.append(dict(pin(target), install_name=dependency))
            pending.append((target, executable))
        result.append({'image': pin(source), 'executable_context': str(executable) if executable else None,
                       'dependencies': deps, 'rpaths': rpaths})
    return {'images': result, 'cached_images': sorted(cached.values(), key=lambda r: r['install_name']),
            'active_cache_uuid': cache_uuid, 'cache_files': cache_files}

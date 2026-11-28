#!/usr/bin/env python3
"""X1 action (b): read-only runtime inspection, immutable byte retention, no bootstrap.

Never runs model CLIs or scanners. All subprocesses have a closed environment;
Go metadata is offline. Missing native libraries or an unidentified active dyld
cache refuse closure, rather than being inferred complete.
"""
import argparse
import base64
import csv
from email.parser import Parser
import hashlib
import json
import os
from pathlib import Path
import platform
import re
import shlex
import shutil
import stat
import subprocess
import sys

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))
from preparation import (Journal, load_config, operation_lock, paths, read_clock,
                         validate_pin, now)
from verify_inputs import exclusive_json, pin, working_pin

PINS_SHA = '6d556fb5156e8ced3c17ae7526a841e51688145faad4a3f8a6a2ce0c939100fc'
RUNTIME_SHA = '19463957becfc23a391eac03582b32eaa0ff392069bccd9aed9000f390ad2d62'
GO_REVIEW_RELATIVE = Path('reviews/closure-model-free.json')
GO_REVIEW_SCOPE = 'closure-model-free-go'
SOURCE_KEYS = ('GoFiles', 'CgoFiles', 'CFiles', 'CXXFiles', 'MFiles', 'HFiles',
               'FFiles', 'SFiles', 'SwigFiles', 'SwigCXXFiles', 'SysoFiles', 'EmbedFiles')


def digest(path):
    with Path(path).open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


def canonical(value):
    return json.dumps(value, sort_keys=True, separators=(',', ':')).encode()


def safe_resolve(path):
    """Resolve every component without ever entering .codegraph; retain link chain."""
    pending, current, chain = list(Path(path).absolute().parts[1:]), Path('/'), []
    while pending:
        part = pending.pop(0)
        if part in ('', '.'):
            continue
        if part == '..':
            current = current.parent
            continue
        if part == '.codegraph':
            raise ValueError('.codegraph access refused')
        current /= part
        if current.is_symlink():
            target = Path(os.readlink(current))
            chain.append({'link_path': str(current), 'link_target': str(target)})
            if len(chain) > 100:
                raise ValueError('symlink cycle: ' + str(path))
            current = Path('/') if target.is_absolute() else current.parent
            pending = list(target.parts[1:] if target.is_absolute() else target.parts) + pending
    return current, chain


def checked(path, expected):
    resolved, chain = safe_resolve(path)
    if expected.get('resolved_path') and resolved != Path(expected['resolved_path']):
        raise ValueError('signed resolved path mismatch: '+str(path))
    actual = pin(resolved)
    if actual['sha256'] != expected['sha256'] or ('bytes' in expected and actual['bytes'] != expected['bytes']):
        raise ValueError('signed working-byte mismatch: ' + str(path))
    with resolved.open('rb') as stream:
        first = stream.read(512).split(b'\n', 1)[0]
    return dict(actual, requested_path=str(path), symlink_chain=chain,
                expected_resolved_path=expected.get('resolved_path'),
                mode=stat.S_IMODE(resolved.stat().st_mode),
                shebang=first.decode('utf-8', 'strict') if first.startswith(b'#!') else None)


class Inventory:
    """Walk each real directory once; external links require explicit approved roots."""
    def __init__(self, roots):
        self.roots = sorted({safe_resolve(p)[0] for p in roots}, key=str)
        self.visited = set()
        self.records = []
        self.files = {}

    def permitted(self, path):
        return any(path.is_relative_to(root) for root in self.roots)

    def visit(self, path):
        path = Path(path)
        resolved, chain = safe_resolve(path)
        if not self.permitted(resolved):
            raise ValueError('unapproved escaped dependency root: ' + str(path) + ' -> ' + str(resolved))
        if path.is_symlink():
            self.records.append(dict(absolute_path=str(path), type='symlink',
                                     mode=stat.S_IMODE(path.lstat().st_mode), size=path.lstat().st_size,
                                     link_target=os.readlink(path), resolved_path=str(resolved), symlink_chain=chain))
        if resolved in self.visited:
            return
        self.visited.add(resolved)
        info = resolved.stat()
        base = dict(absolute_path=str(resolved), mode=stat.S_IMODE(info.st_mode), size=info.st_size)
        if stat.S_ISDIR(info.st_mode):
            self.records.append(dict(base, type='directory'))
            for child in sorted(resolved.iterdir()):
                self.visit(child)
        elif stat.S_ISREG(info.st_mode):
            record = pin(resolved)
            self.files[str(resolved)] = record
            self.records.append(dict(base, type='file', sha256=record['sha256']))
        else:
            raise ValueError('unsupported runtime entry: ' + str(resolved))

    def collect(self):
        for root in self.roots:
            self.visit(root)
        result = []
        for root in self.roots:
            records = [dict(row, relative_path=str(Path(row['absolute_path']).relative_to(root)))
                       for row in self.records if Path(row['absolute_path']).is_relative_to(root)]
            result.append({'root': str(root), 'records': sorted(records, key=lambda r: (r['relative_path'], r['type']))})
        return result


def retain(files, destination):
    """Content-addressed immutable backup, never chmod or write original runtimes."""
    destination = Path(destination)
    destination.mkdir(parents=True, exist_ok=True)
    result = {}
    for row in files:
        source = Path(row['path'])
        target = destination / row['sha256']
        if not target.exists():
            try:
                with target.open('xb') as out, source.open('rb') as inp:
                    while block := inp.read(1024 * 1024):
                        out.write(block)
                    out.flush()
                    os.fsync(out.fileno())
                target.chmod(0o444)
            except BaseException:
                # Preserve an incomplete reserved object for reviewed recovery.
                raise
        actual = pin(target)
        if actual['sha256'] != row['sha256'] or actual['bytes'] != row['bytes'] or target.stat().st_mode & 0o222:
            raise ValueError('incomplete/writable retained object: ' + str(target))
        if digest(source) != row['sha256']:
            raise ValueError('runtime changed during retention: ' + str(source))
        result[str(target)] = actual
    fd = os.open(destination, os.O_RDONLY)
    try:
        os.fsync(fd)
    finally:
        os.close(fd)
    return sorted(result.values(), key=lambda row: row['path'])


def marker_applies(marker):
    match = re.fullmatch(r"(python_full_version|platform_python_implementation)\s*(==|!=|<=|>=|<|>)\s*(['\"])([^'\"]+)\3", marker)
    if not match:
        raise ValueError('unsupported frozen requirement marker: '+marker)
    name, operator, _, expected = match.groups()
    if name == 'python_full_version':
        if not re.fullmatch(r'\d+(?:\.\d+){1,2}', expected):
            raise ValueError('unsupported Python version marker')
        observed = tuple(map(int, platform.python_version_tuple()))
        parts = tuple(map(int, expected.split('.')))
        expected = parts+(0,)*(3-len(parts))
    else:
        if operator not in ('==', '!='):
            raise ValueError('unsupported implementation marker operator')
        observed = platform.python_implementation()
    return {'==': observed == expected, '!=': observed != expected,
            '<': observed < expected, '>': observed > expected,
            '<=': observed <= expected, '>=': observed >= expected}[operator]


def requirements(path):
    text = re.sub(r'\\\n', ' ', Path(path).read_text())
    result = {}
    for line in text.splitlines():
        line = line.split('#', 1)[0].strip()
        if not line:
            continue
        declaration, separator, hashes = line.partition('--hash=')
        match = re.fullmatch(r'([A-Za-z0-9_.-]+)==([^\s;]+)(?:\s*;\s*(.+?))?\s*', declaration)
        hashes = separator+hashes
        if not match or not re.fullmatch(r'(?:--hash=sha256:[0-9a-f]{64}\s*)+', hashes):
            raise ValueError('unsupported frozen requirement syntax: ' + line[:80])
        if match[3] and not marker_applies(match[3]):
            continue
        name = re.sub('[-_.]+', '-', match[1]).lower()
        if name in result:
            raise ValueError('duplicate applicable frozen requirement: ' + name)
        result[name] = {'version': match[2], 'archive_sha256': re.findall(r'sha256:([0-9a-f]{64})', hashes)}
    if not result:
        raise ValueError('empty requirements lock')
    return result


def python_packages(venv, lock):
    wanted, observed = requirements(lock), {}
    sites = list((Path(venv)/'lib').glob('python*/site-packages'))
    if len(sites) != 1:
        raise ValueError('ambiguous Python site-packages')
    site = sites[0]
    for metadata in sorted(site.glob('*.dist-info/METADATA')):
        fields = Parser().parsestr(metadata.read_text())
        name = re.sub('[-_.]+', '-', fields['Name']).lower()
        version = fields['Version']
        if name in observed or name not in wanted or version != wanted[name]['version']:
            raise ValueError('unlocked/duplicate Python distribution: ' + name + '==' + version)
        record = metadata.parent/'RECORD'
        count = 0
        with record.open(newline='') as stream:
            for relative, encoded, size in csv.reader(stream):
                path, _ = safe_resolve(site/relative)
                if not path.is_relative_to(Path(venv).resolve()):
                    raise ValueError('Python RECORD escapes venv: ' + relative)
                if not path.is_file():
                    raise ValueError('missing Python RECORD file: ' + relative)
                if encoded:
                    algorithm, value = encoded.split('=', 1)
                    if algorithm not in ('sha256', 'sha384', 'sha512'):
                        raise ValueError('unsupported RECORD digest')
                    with path.open('rb') as inp:
                        actual = base64.urlsafe_b64encode(hashlib.file_digest(inp, algorithm).digest()).decode().rstrip('=')
                    if actual != value or (size and path.stat().st_size != int(size)):
                        raise ValueError('installed RECORD mismatch: ' + str(path))
                elif path != record and path.suffix != '.pyc':
                    raise ValueError('unhashed installed RECORD file: ' + str(path))
                count += 1
        observed[name] = dict(version=version, metadata=pin(metadata), record=pin(record),
                              requires_dist=fields.get_all('Requires-Dist', []),
                              verified_record_rows=count, requested_archive_sha256=wanted[name]['archive_sha256'],
                              archive_provenance='frozen hashed lock; installed bytes independently verified via RECORD')
    if set(observed) != set(wanted):
        raise ValueError('missing Python distributions: ' + ', '.join(sorted(set(wanted)-set(observed))))
    return observed


def npm_packages(root):
    root = Path(root)
    lock = json.loads((root/'package-lock.json').read_bytes())
    if lock.get('lockfileVersion') not in (2, 3) or not lock.get('packages'):
        raise ValueError('unsupported npm lock')
    records = []
    for relative, expected in sorted(lock['packages'].items()):
        if not relative:
            continue
        path = root/relative/'package.json'
        if not path.exists():
            oses, cpus = expected.get('os', []), expected.get('cpu', [])
            excluded = (oses and 'darwin' not in oses) or (cpus and 'arm64' not in cpus)
            if expected.get('optional') and excluded:
                records.append({'package': relative, 'not_installed_platform_optional': True})
                continue
            raise ValueError('missing npm package: ' + str(path))
        actual = json.loads(path.read_bytes())
        inferred_name = relative.rsplit('node_modules/', 1)[-1]
        if actual.get('name') != expected.get('name', inferred_name) or actual.get('version') != expected.get('version') or not expected.get('integrity'):
            raise ValueError('npm identity/integrity declaration mismatch: ' + relative)
        records.append(dict(package=relative, name=actual['name'], version=actual['version'],
                            integrity=expected['integrity'], resolved=expected.get('resolved'),
                            metadata=pin(path), dependencies=actual.get('dependencies', {})))
    # Only installation-layout node_modules below root/lock-listed packages count.
    parents = [root, *[root/name for name in lock['packages'] if name]]
    for parent in parents:
        modules = parent/'node_modules'
        if not modules.is_dir():
            continue
        for child in sorted(modules.iterdir()):
            if child.name == '.codegraph':
                raise ValueError('.codegraph access refused')
            if child.name.startswith('.'):
                continue
            packages = sorted(child.iterdir()) if child.name.startswith('@') else [child]
            for package in packages:
                if not package.is_dir():
                    continue
                relative = str(package.relative_to(root))
                if relative not in lock['packages']:
                    raise ValueError('unlocked npm package: '+relative)
    return {'lock': pin(root/'package-lock.json'), 'packages': records,
            'integrity_scope': 'frozen archive integrity declarations; installed trees separately SHA-256 enumerated'}


def node_edges(package, boundary, allowed):
    """Resolve required runtime dependencies with Node's node_modules ascent."""
    package, boundary = Path(package), Path(boundary)
    data = json.loads((package/'package.json').read_bytes())
    edges = []
    for name, requested in sorted(data.get('dependencies', {}).items()):
        if name in data.get('optionalDependencies', {}):
            continue
        if name.startswith('/') or '..' in Path(name).parts:
            raise ValueError('unsafe dependency name: ' + name)
        current, target = package, None
        while current.is_relative_to(boundary):
            candidate = current/'node_modules'/name/'package.json'
            if candidate.is_file():
                target = candidate
                break
            current = current.parent
        if target is None:
            raise ValueError('missing required Node dependency: '+data['name']+' -> '+name)
        actual = json.loads(target.read_bytes())
        identity = actual['name']+'@'+actual['version']
        if actual['name'] != name or identity not in allowed:
            raise ValueError('unlocked Node dependency edge: '+identity)
        edges.append({'name': name, 'requested': requested, 'observed': identity, 'metadata': pin(target)})
    return edges


def pnpm_packages(root):
    root = Path(root)
    frozen = root/'pnpm-lock.yaml'
    installed = root/'node_modules/.pnpm/lock.yaml'
    if frozen.read_bytes() != installed.read_bytes():
        raise ValueError('pnpm installed lock differs from frozen revision lock')
    text = frozen.read_text()
    # Frozen v9 package headers; reject another format instead of guessing YAML.
    if not re.search(r"^lockfileVersion: ['\"]?9\.0['\"]?$", text, re.M):
        raise ValueError('unsupported pnpm lock version')
    section, allowed = None, {}
    for line in text.splitlines():
        if line and not line.startswith(' '):
            section = line.rstrip(':')
        if section == 'packages':
            header = re.match(r"^  (?! )(['\"]?)(\S.*?)\1:$", line)
            if header:
                key = header[2]
                allowed[key] = []
            integrity = re.search(r'integrity: ([A-Za-z0-9+/=-]+)', line)
            if integrity and allowed:
                allowed[key].append(integrity[1])
    if not allowed or any(not value for value in allowed.values()):
        raise ValueError('missing pnpm package integrity declaration')
    records = []
    store = root/'node_modules/.pnpm'
    for directory in sorted(store.iterdir()):
        if not directory.is_dir() or directory.name == 'node_modules':
            continue
        modules = directory/'node_modules'
        if not modules.is_dir():
            raise ValueError('incomplete pnpm virtual store: ' + str(directory))
        for scope in sorted(modules.iterdir()):
            candidates = list(scope.iterdir()) if scope.name.startswith('@') and not scope.is_symlink() else [scope]
            for package in candidates:
                if package.is_symlink() or not (package/'package.json').is_file():
                    continue
                data = json.loads((package/'package.json').read_bytes())
                key = data['name'] + '@' + data['version']
                if key not in allowed:
                    raise ValueError('unlocked pnpm package: ' + key)
                records.append({'identity': key, 'metadata': pin(package/'package.json'),
                                'integrities': allowed[key], 'dependencies': data.get('dependencies', {})})
                records[-1]['resolved_dependencies'] = node_edges(package, root, allowed)
    if not records:
        raise ValueError('empty pnpm installed package graph')
    root_edges = node_edges(root, root, allowed)
    return {'frozen_lock': pin(frozen), 'installed_lock': pin(installed), 'packages': records,
            'root_dependencies': root_edges, 'locked_packages': allowed,
            'integrity_scope': 'archive declarations plus full installed-byte enumeration'}


class Commands:
    def __init__(self, environment, record_directory=None):
        self.environment, self.records = environment, []
        self.record_directory = record_directory

    def run(self, argv, cwd):
        # Deliberately no shell and no ambient environment. Never model executables.
        if Path(argv[0]).name in ('claude', 'codex'):
            raise ValueError('model executable forbidden')
        start = now()
        process = subprocess.run([str(x) for x in argv], cwd=cwd, env=self.environment,
                                 stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=False)
        stdout, stderr = process.stdout.decode(), process.stderr.decode()
        self.records.append({'argv': [str(x) for x in argv], 'cwd': str(cwd),
                             'started_at': start, 'ended_at': now(), 'exit_code': process.returncode,
                             'stdout': stdout, 'stderr': stderr})
        if self.record_directory is not None:
            exclusive_json(Path(self.record_directory)/f'{len(self.records):06d}.json',
                           dict(self.records[-1], environment=self.environment))
        if process.returncode:
            raise ValueError('closure metadata command failed: ' + repr(argv))
        return stdout


def json_stream(text):
    decoder, offset, result = json.JSONDecoder(), 0, []
    while offset < len(text):
        while offset < len(text) and text[offset].isspace():
            offset += 1
        if offset == len(text):
            break
        item, offset = decoder.raw_decode(text, offset)
        result.append(item)
    return result


def go_metadata(go, root, commands, allowed_sources=None, reviewed_frontend=()):
    go = Path(go)
    goroot = Path(commands.run([go, 'env', 'GOROOT'], root).strip())
    packages = json_stream(commands.run([go, 'list', '-pgo=off', '-deps', '-json',
                                        './evaluation/x1-triage-ranking/execution/model_free'], root))
    if not packages:
        raise ValueError('empty Go package closure')
    files, selected, frontend = {}, [], set()
    for package in packages:
        if package.get('Error') or package.get('DepsErrors') or package.get('Incomplete'):
            raise ValueError('incomplete Go package closure')
        selected.append({k: package[k] for k in ('ImportPath', 'Dir', 'Name', 'Standard', 'Module', 'Imports', 'ImportMap', *SOURCE_KEYS) if k in package})
        for key in SOURCE_KEYS:
            for name in package.get(key, []):
                logical = Path(os.path.abspath(Path(package['Dir'])/name))
                row = selected_source(logical, root, allowed_sources or {},
                                      main_module=package.get('Module', {}).get('Main', False))
                files[row['path']] = row
                if logical.parent == root/'evaluation/x1-triage-ranking/execution/model_free' and logical.suffix == '.go':
                    frontend.add(str(logical))
    if frontend != set(reviewed_frontend):
        raise ValueError('Go authority must cover exactly the selected model_free/*.go files')
    sums = {tuple(line.split()[:2]): line.split()[2] for line in (root/'go.sum').read_text().splitlines() if line.strip()}
    graph = json_stream(commands.run([go, 'list', '-m', '-json', 'all'], root))
    selected_modules = {package['Module']['Path'] for package in packages if 'Module' in package}
    modules = [module for module in graph if module['Path'] in selected_modules]
    cache = Path(commands.environment['GOMODCACHE'])
    for module in modules:
        if module.get('Main'):
            continue
        if module.get('Replace'):
            raise ValueError('unreviewed replaced Go module')
        escaped = lambda value: ''.join('!'+c.lower() if c.isupper() else c for c in value)
        download = cache/'cache/download'/escaped(module['Path'])/'@v'
        ziphash = download/(escaped(module['Version'])+'.ziphash')
        observed_sum = ziphash.read_text().strip()
        if sums.get((module['Path'], module['Version'])) != observed_sum:
            raise ValueError('Go cached module sum not bound by go.sum: '+module['Path'])
        gomod = Path(module['GoMod'])
        mod_hash = hashlib.sha256((digest(gomod)+'  go.mod\n').encode()).digest()
        observed_mod_sum = 'h1:'+base64.b64encode(mod_hash).decode()
        if sums.get((module['Path'], module['Version']+'/go.mod')) != observed_mod_sum:
            raise ValueError('Go module metadata not bound by go.sum: '+module['Path'])
        module['verified_sum'], module['verified_gomod_sum'] = observed_sum, observed_mod_sum
        files[str(ziphash)], files[str(gomod)] = pin(ziphash), pin(gomod)
    verification = commands.run([go, 'mod', 'verify'], root)
    if verification.strip() != 'all modules verified':
        raise ValueError('Go module directory verification incomplete')
    return {'version': commands.run([go, 'version'], root).strip(), 'goroot': str(goroot),
            'packages': sorted(selected, key=lambda p: p['ImportPath']), 'modules': modules, 'module_graph': graph,
            'selected_files': sorted(files.values(), key=lambda r: r['path']),
            'module_verification': verification.strip()}, [goroot, *[Path(m['Dir']) for m in modules if not m.get('Main')]]


def native_closure(files, commands, cwd, cache_root=None):
    from closure.native import collect
    return collect(files, commands, cwd, cache_root, safe_resolve, pin)


def own_sources():
    return [pin(HERE/name) for name in ('attest_closure.py', 'closure/native.py',
                                       'preparation.py', 'verify_inputs.py', 'run-config.json')]


def verify_installed_closure(path):
    """Recheck original byte/link/file sets before each authorised scanner run."""
    value = json.loads(Path(path).read_bytes())
    if value.get('schema') != 'scout.x1.installed-closure.v1' or value.get('status') != 'passed' or value.get('complete') is not True:
        raise ValueError('complete passing installed closure required')
    roots = [row['root'] for row in value['enumeration']]
    if Inventory(roots).collect() != value['enumeration']:
        raise ValueError('installed runtime file/link/mode set drift')
    for field in ('inputs', 'source_pins', 'original_runtime_files',
                  'submodule_working_files', 'deferred_working_files_verified', 'retained_files'):
        for row in value[field]:
            validate_pin(row)
    for row in value['retained_files']:
        if Path(row['path']).stat().st_mode & 0o222:
            raise ValueError('retained runtime object became writable')
    for row in value['executable_resolutions']:
        resolved, links = safe_resolve(row['requested_path'])
        if str(resolved) != row['path'] or links != row['symlink_chain']:
            raise ValueError('entrypoint resolution drift: '+row['requested_path'])
    return value


def verify_deferred(root, rows, deferred, count):
    expected = {row['path']: row for row in rows}
    if len(expected) != len(rows):
        raise ValueError('duplicate submodule working-file pin')
    if deferred and (len(deferred) != count or {row['path'] for row in deferred} != set(expected)):
        raise ValueError('input gate deferred working-file set is incomplete/duplicated')
    for row in deferred:
        if row['path'] not in expected or any(row[key] != expected[row['path']][key] for key in ('sha256', 'bytes', 'git_mode', 'git_blob')):
            raise ValueError('input gate deferral differs from signed working-file pin')
    working = {row['path']: working_pin(root, row) for row in rows}
    return list(working.values()), [working[row['path']] for row in deferred]


def verify_claw_assets(claw, pins):
    rows = {row['path']: row for row in pins['submodule']['working_files']}
    return [checked(Path(claw)/name, rows['third_party/clawpatch/'+name])
            for name in ('package.json', 'pnpm-lock.yaml', 'pnpm-workspace.yaml')]


def object_git(config):
    p = paths(config)
    receipts, identities = [], []
    for repo in config['repositories']:
        path = p['OUT']/'objects'/(repo['repository'].replace('/', '__')+'.json')
        value = json.loads(path.read_bytes())
        if value.get('schema') != 'scout.x1.objects.v1' or value.get('status') != 'passed':
            raise ValueError('passing object receipt required for Git identity')
        identities.append(value['git_identity'])
        receipts.append(pin(path))
    if len(identities) != 5 or any(row != identities[0] for row in identities[1:]):
        raise ValueError('object receipts disagree on Git identity')
    identity = identities[0]
    validate_pin(identity)
    executable, _ = safe_resolve(identity['path'])
    if executable != Path(identity['path']) or executable.parent.name != 'bin':
        raise ValueError('Git receipt must name resolved keg executable, not a shim')
    keg = executable.parent.parent
    if not (keg/'libexec/git-core').is_dir():
        raise ValueError('Git keg lacks git-core helper closure')
    return identity, keg, receipts


def runtime_python(runtime_root, runtime):
    command = Path(runtime_root)/'.tools/venv/bin/python'
    if command != Path(runtime['numerical_interpreter']['command']):
        raise ValueError('runtime location differs from signed interpreter command')
    python = checked(command, runtime['numerical_interpreter'])
    if safe_resolve(sys.executable)[0] != Path(python['path']) or platform.python_version() != '3.11.11' or not sys.flags.isolated or not sys.dont_write_bytecode:
        raise ValueError('attestor requires pinned runtime-root/.tools/venv/bin/python -I -B')
    return python


def verify_interpreter_roots(interpreters, inventory):
    for row in interpreters:
        if not inventory.permitted(safe_resolve(row['path'])[0]):
            raise ValueError('launcher interpreter outside enumerated roots: '+row['path'])


def reviewed_go_sources(root, out):
    from preparation import review_receipt
    review_path = Path(out)/GO_REVIEW_RELATIVE
    value = review_receipt(review_path, GO_REVIEW_SCOPE)
    directory = Path(root)/'evaluation/x1-triage-ranking/execution/model_free'
    rows = value['reviewed']
    selected = {row['path']: row for row in rows}
    if len(selected) != len(rows) or any(Path(name).parent != directory or Path(name).suffix != '.go' for name in selected):
        raise ValueError('dedicated closure review must contain only unique model_free/*.go files')
    return selected, [pin(review_path), validate_pin(value['record'])]


def selected_source(path, root, allowed, main_module=False):
    logical = Path(os.path.abspath(path))
    if main_module or logical.is_relative_to(root):
        expected = allowed.get(str(logical))
        if not logical.is_relative_to(root) or not expected:
            raise ValueError('unpinned selected Go source: '+str(logical))
        resolved, chain = safe_resolve(logical)
        # The current signed source/review rows pin bytes, not link chains.
        if chain or not resolved.is_relative_to(root):
            raise ValueError('unpinned selected Go source link chain: '+str(logical))
        if digest(logical) != expected['sha256']:
            raise ValueError('unpinned selected Go source: '+str(logical))
    resolved, _ = safe_resolve(logical)
    return pin(resolved)


def verify_git_runtime(identity, root, submodule, expected_head, commands):
    git = identity['path']
    if commands.run([git, '--version'], root).strip() != identity['version']:
        raise ValueError('object Git observed version drift')
    if commands.run([git, '-C', Path(root)/submodule, 'rev-parse', 'HEAD'], root).strip() != expected_head:
        raise ValueError('Clawpatch submodule gitlink/HEAD mismatch')


def claw_runtime(claw, pins):
    assets = verify_claw_assets(claw, pins)
    return assets, pnpm_packages(claw)


def runtime_inventory(tools, go_roots, git_root, external_roots, interpreters):
    inventory = Inventory([tools, Path(sys.base_prefix), git_root, *go_roots, *external_roots])
    verify_interpreter_roots(interpreters, inventory)
    return inventory, inventory.collect()


def inventory_native_closure(inventory, os_tools, interpreters, commands, root, cache_root):
    # Selection must inspect every regular file, not just executable/suffix rows.
    return native_closure([*inventory.files, *os_tools, *[row['path'] for row in interpreters]],
                          commands, root, cache_root)


def attest(args, config, journal):
    p = paths(config)
    if platform.system() != 'Darwin' or platform.machine() != 'arm64':
        raise ValueError('runtime recipe requires Darwin arm64')
    if not safe_resolve(args.module_cache)[0].is_relative_to(safe_resolve(p['SCRATCH'])[0]):
        raise ValueError('Go module cache must be a dedicated copy under SCRATCH')
    runtime_root = Path(args.runtime_root).absolute()
    pins_path = Path(args.pins).absolute()
    if pins_path != p['STUDY']/'preparation/pins.json':
        raise ValueError('noncanonical pins path')
    checked(pins_path, {'sha256': PINS_SHA})
    runtime_path = p['STUDY']/'preparation/runtime-attestation-rev6.json'
    checked(runtime_path, {'sha256': RUNTIME_SHA})
    pins = json.loads(pins_path.read_bytes())
    runtime = json.loads(runtime_path.read_bytes())
    source_records = own_sources()
    sources = [checked(runtime_root/row['path'], row) for row in pins['source']['files']]
    # Check mapped writer source too: Go's selected source must not drift from pins.
    sources.extend(checked(p['ROOT']/row['path'], row) for row in pins['source']['files'])
    deferred = json.loads((p['OUT']/'input-verification.json').read_bytes()).get('deferred_to_b', [])
    working, deferred_verified = verify_deferred(p['ROOT'], pins['submodule']['working_files'],
                                                deferred, config['deferred_to_b_contract']['working_file_count'])
    git_identity, git_root, git_receipts = object_git(config)
    reviewed_go, go_review_pins = reviewed_go_sources(p['ROOT'], p['OUT'])
    tool_rows = []
    for row in pins['toolchain']['installed_entrypoints'] + pins['toolchain']['additional_assets']:
        path = Path(row['path'])
        actual = checked(path if path.is_absolute() else runtime_root/path, row)
        tool_rows.append(dict(actual, name=row.get('name', row['path']), execution_allowed=row.get('name') not in ('claude', 'codex')))
    python = runtime_python(runtime_root, runtime)
    go = checked(runtime['go']['binary'], runtime['go'])
    private = p['SCRATCH']/'closure'/journal.id
    env = {'HOME': str(private/'home'), 'PATH': str(git_root/'bin')+':'+str(Path(go['path']).parent)+':'+str(runtime_root/'.tools/venv/bin')+':'+str(runtime_root/'.tools/bin')+':/usr/bin:/bin',
           'LANG': 'C', 'LC_ALL': 'C', 'GOCACHE': str(private/'gocache'), 'GOTMPDIR': str(private/'tmp'),
           'GOMODCACHE': str(Path(args.module_cache).absolute()), 'GOTOOLCHAIN': 'local', 'GOFLAGS': '-mod=readonly',
           'GOPROXY': 'off', 'GOSUMDB': 'off', 'CGO_ENABLED': '0', 'GOWORK': 'off', 'GOENV': 'off',
           'GOOS': 'darwin', 'GOARCH': 'arm64', 'GOEXPERIMENT': '', 'GOARM64': 'v8.0',
           'GIT_CONFIG_NOSYSTEM': '1', 'GIT_CONFIG_GLOBAL': '/dev/null', 'GIT_NO_REPLACE_OBJECTS': '1',
           'GIT_TERMINAL_PROMPT': '0', 'PYTHONDONTWRITEBYTECODE': '1'}
    for key in ('HOME', 'GOCACHE', 'GOTMPDIR'):
        Path(env[key]).mkdir(parents=True, exist_ok=False)
    commands = Commands(env, private/'commands')
    interpreter_files = []
    for row in tool_rows:
        if not row['execution_allowed'] or not row['shebang']:
            continue
        words = shlex.split(row['shebang'][2:])
        if not words:
            raise ValueError('empty launcher interpreter')
        launcher = words.pop(0)
        if launcher == '/usr/bin/env':
            if words and words[0] == '-S':
                words.pop(0)
            if not words or words[0].startswith('-') or '=' in words[0]:
                raise ValueError('unsupported env launcher interpreter')
            launcher = shutil.which(words.pop(0), path=env['PATH'])
        if not launcher or not Path(launcher).is_absolute():
            raise ValueError('unresolved launcher interpreter')
        resolved, chain = safe_resolve(launcher)
        interpreter = pin(resolved)
        row['interpreter_resolution'] = dict(interpreter, symlink_chain=chain, arguments=words)
        interpreter_files.append(interpreter)
    git = git_identity['path']
    verify_git_runtime(git_identity, p['ROOT'], pins['submodule']['path'], pins['submodule']['gitlink'], commands)
    allowed = {str(p['ROOT']/row['path']): row for row in pins['source']['files']}
    allowed.update(reviewed_go)
    go_info, go_roots = go_metadata(go['path'], p['ROOT'], commands, allowed, reviewed_go)
    if go_info['version'] != runtime['go']['version']:
        raise ValueError('Go observed version differs')
    tools = runtime_root/'.tools'
    claw = tools/'clawpatch'/pins['submodule']['gitlink']
    for name in ('dist', 'node_modules', 'package.json', 'pnpm-lock.yaml', '.scout-review-safety'):
        if not (claw/name).exists():
            raise ValueError('missing Clawpatch install component: ' + name)
    claw_assets, claw_info = claw_runtime(claw, pins)
    sources.extend(claw_assets)
    venv = tools/'venv'
    python_info = python_packages(venv, runtime_root/'toolchain/python-requirements.lock')
    node_info = npm_packages(tools/'node-tools')
    patch_hash = next(row['sha256'] for row in pins['source']['files'] if row['path'] == pins['submodule']['patch'])
    if (claw/'.scout-review-safety').read_text().strip() != 'scout-review-sanitize-v3 '+patch_hash:
        raise ValueError('Clawpatch safety-patch stamp mismatch')
    observed_versions = {
        'GO_VERSION': go_info['version'].split()[2].removeprefix('go'),
        'PYTHON_VERSION': platform.python_version(),
        'NODE_VERSION': commands.run([tools/'bin/node', '--version'], p['ROOT']).strip().removeprefix('v'),
        'UV_VERSION': commands.run([tools/'bin/uv', '--version'], p['ROOT']).split()[1],
        'GITLEAKS_VERSION': commands.run([tools/'bin/gitleaks', 'version'], p['ROOT']).strip(),
        'CLAWPATCH_VERSION': json.loads((claw/'package.json').read_bytes())['version'],
    }
    for key, distribution in (('SEMGREP_VERSION', 'semgrep'), ('RUFF_VERSION', 'ruff'),
                              ('LIZARD_VERSION', 'lizard'), ('MYPY_VERSION', 'mypy')):
        observed_versions[key] = python_info[distribution]['version']
    for key, package in (('PNPM_VERSION', 'pnpm'), ('JSCPD_VERSION', 'jscpd'), ('PRETTIER_VERSION', 'prettier')):
        observed_versions[key] = json.loads((tools/'node-tools/node_modules'/package/'package.json').read_bytes())['version']
    build_metadata = commands.run([go['path'], 'version', '-m', tools/'bin/govulncheck'], p['ROOT'])
    module_version = re.search(r'^\s*mod\s+golang.org/x/vuln\s+v([^\s]+)', build_metadata, re.M)
    if not module_version:
        raise ValueError('govulncheck module version metadata unavailable')
    observed_versions['GOVULNCHECK_VERSION'] = module_version[1]
    if observed_versions != pins['toolchain']['requested_versions']:
        raise ValueError('requested/observed tool versions differ: '+json.dumps(observed_versions, sort_keys=True))
    inventory, enumeration = runtime_inventory(tools, go_roots, git_root,
                                               [Path(x) for x in args.external_root], interpreter_files)
    roots = inventory.roots
    os_tools = ['/usr/bin/sandbox-exec', '/bin/sh', '/bin/bash', '/usr/bin/env', git,
                '/usr/bin/otool', '/usr/bin/dyld_info', '/usr/bin/sw_vers']
    os_files = [pin(safe_resolve(path)[0]) for path in os_tools]
    native = inventory_native_closure(inventory, os_tools, interpreter_files, commands, p['ROOT'], args.dyld_cache_root)
    for image in native['images']:
        for row in [image['image'], *image['dependencies']]:
            if 'path' in row:
                inventory.files[row['path']] = {k: row[k] for k in ('path', 'sha256', 'bytes')}
    for row in [*os_files, *native['cache_files']]:
        inventory.files[row['path']] = row
    for row in go_info['selected_files']:
        inventory.files[row['path']] = row
    retained_root = (p['SCRATCH']/'closure-preflight-retained'/journal.id if args.preflight
                     else p['SCRATCH']/'closure-retained/sha256')
    retained = retain(list(inventory.files.values()), retained_root)
    # Re-enumerate after all metadata subprocesses and copies; no concurrent mutation.
    recheck = Inventory(roots)
    if recheck.collect() != enumeration:
        raise ValueError('runtime tree mutated during attestation')
    return dict(schema='scout.x1.installed-closure.v1', status='passed', complete=True,
                clock=pin(args.clock), configuration=pin(args.config),
                inputs=[pin(pins_path), pin(runtime_path), pin(p['OUT']/'input-verification.json'), *git_receipts, *go_review_pins],
                source_pins=[*sources, *source_records],
                runtime_root=str(runtime_root), historical_path_mapping={str(runtime_root): str(runtime_root)},
                runtime_selection='byte-identical existing installation; read-only reuse, no bootstrap',
                deferred_working_files_verified=deferred_verified, submodule_working_files=working,
                requested_versions=pins['toolchain']['requested_versions'], executable_resolutions=tool_rows,
                observed_versions=observed_versions, govulncheck_build_metadata=build_metadata,
                numerical_interpreter=python, git=git_identity, go=go_info, python_packages=python_info, node_packages=node_info,
                clawpatch_packages=claw_info, enumeration=enumeration,
                original_runtime_files=sorted(inventory.files.values(), key=lambda r: r['path']), retained_files=retained,
                native_dependencies=native, os_executables=os_files,
                platform={'architecture': platform.machine(), 'system': platform.system(),
                          'release': platform.release(), 'version': commands.run(['/usr/bin/sw_vers'], p['ROOT'])},
                environment=env, commands=commands.records, argv=sys.argv, cwd=os.getcwd(),
                started_at=journal.started_at, ended_at=now(), key_present=False, failures=[],
                canonical_inventory_sha256=hashlib.sha256(canonical(enumeration)).hexdigest(),
                journal_start=pin(journal.start_path))


def dry_check():
    """Review-only real-byte checks; stdout only, no journal/output."""
    from closure.native import image_filetype, resolve_install_name
    parser = argparse.ArgumentParser(description=dry_check.__doc__)
    parser.add_argument('--dry-check', action='store_true', required=True)
    parser.add_argument('--runtime-root', required=True)
    parser.add_argument('--external-root', action='append', default=[],
                        help='read-only whole-root native-format classification, without loading images')
    args = parser.parse_args()
    if 'OPENROUTER_API_KEY' in os.environ:
        raise ValueError('key present; dry check refused')
    root = Path(args.runtime_root).absolute()
    pins_path = HERE.parent/'preparation/pins.json'
    checked(pins_path, {'sha256': PINS_SHA})
    pins = json.loads(pins_path.read_bytes())
    runtime_path = HERE.parent/'preparation/runtime-attestation-rev6.json'
    checked(runtime_path, {'sha256': RUNTIME_SHA})
    python = runtime_python(root, json.loads(runtime_path.read_bytes()))
    claw = root/'.tools/clawpatch'/pins['submodule']['gitlink']
    assets = verify_claw_assets(claw, pins)
    packages = python_packages(root/'.tools/venv', root/'toolchain/python-requirements.lock')
    pnpm = pnpm_packages(claw)
    npm = npm_packages(root/'.tools/node-tools')
    commands = Commands({'PATH': '/usr/bin:/bin', 'LANG': 'C', 'LC_ALL': 'C'})
    executable = Path(python['path'])
    if image_filetype(executable) != 2:
        raise ValueError('pinned Python is not MH_EXECUTE')
    loads = commands.run(['/usr/bin/otool', '-L', executable], root)
    name = '@executable_path/../lib/libpython3.11.dylib'
    if name not in loads:
        raise ValueError('expected signed Python executable-relative load absent')
    library, _ = safe_resolve(resolve_install_name(name, executable, executable, [])[0])
    classifications = []
    for external in args.external_root:
        inventory = Inventory([external])
        inventory.collect()
        images, classes = [], []
        for path in sorted(inventory.files):
            kind = image_filetype(path)
            if kind is not None:
                images.append({'image': inventory.files[path], 'filetype': kind})
            elif Path(path).suffix == '.class':
                classes.append(inventory.files[path])
        classifications.append({'root': str(safe_resolve(external)[0]), 'regular_files': len(inventory.files),
                                'mach_o_images': images, 'non_mach_o_classes': classes})
    result = {'schema': 'scout.x1.closure-review-dry-check.v1', 'status': 'passed',
              'canonical_writes': False, 'runtime_root': str(root),
              'R1-01': {'applicable_names': len(packages), 'installed_distributions': len(packages),
                        'deprecated': packages['deprecated']['version'], 'librt': packages['librt']['version']},
              'R1-02': {'python': python, 'resolved_libpython': pin(library), 'load_command': name},
              'R1-03': {'locked_packages': len(pnpm['locked_packages']), 'installed_packages': len(pnpm['packages']),
                        'signed_assets': assets},
              'R1-04': {'lock_listed_packages': len(npm['packages'])}, 'commands': commands.records}
    result['R2-01'] = {'external_root_classifications': classifications}
    print(json.dumps(result, indent=2, sort_keys=True))


def main():
    if '--dry-check' in sys.argv:
        return dry_check()
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--config', required=True)
    parser.add_argument('--clock', required=True)
    parser.add_argument('--pins', required=True)
    parser.add_argument('--output')
    parser.add_argument('--preflight', action='store_true', help='journal action b, write only a scratch receipt')
    parser.add_argument('--runtime-root', required=True, help='existing byte-identical source/install root; never modified')
    parser.add_argument('--module-cache', required=True, help='dedicated module cache copy under SCRATCH; no shared-cache writes')
    parser.add_argument('--external-root', action='append', default=[], help='explicit additional dependency root; whole tree retained')
    parser.add_argument('--dyld-cache-root', required=True, help='active split-cache directory; UUID must match running dyld')
    args = parser.parse_args()
    config = load_config(args.config)
    p = paths(config)
    if args.preflight:
        if args.output:
            raise ValueError('preflight output is derived under SCRATCH, not caller-selected')
    elif not args.output or Path(args.output).absolute() != p['OUT']/'closure.json':
        raise ValueError('noncanonical closure output')
    elif Path(args.output).exists():
        raise ValueError('closure already reserved; reviewed recovery required')
    if 'OPENROUTER_API_KEY' in os.environ:
        raise ValueError('key present; closure refused')
    read_clock(config, args.config, args.clock)
    _, go_review_pins = reviewed_go_sources(p['ROOT'], p['OUT'])
    with operation_lock(config):
        journal = Journal(config, args.config, args.clock, 'b',
                          source_pins=[*own_sources(), pin(args.pins), *go_review_pins],
                          environment={'OPENROUTER_API_KEY_present': False})
        if args.preflight:
            args.output = str(p['SCRATCH']/'closure-preflight'/(journal.id+'.json'))
        try:
            result = attest(args, config, journal)
            exclusive_json(args.output, result)
            journal.finish('passed', [pin(args.output)])
        except BaseException as error:
            if not Path(args.output).exists():
                exclusive_json(args.output, {'schema': 'scout.x1.installed-closure.v1', 'status': 'refused',
                                            'complete': False, 'clock': pin(args.clock),
                                            'configuration': pin(args.config), 'failures': [str(error)],
                                            'metadata_command_records': [pin(path) for path in sorted(
                                                (p['SCRATCH']/'closure'/journal.id/'commands').glob('*.json'))],
                                            'journal_start': pin(journal.start_path), 'ended_at': now()})
            journal.finish('refused', [pin(args.output)], reason=str(error), exit_code=1)
            raise


if __name__ == '__main__':
    main()

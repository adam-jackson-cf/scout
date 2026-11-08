#!/usr/bin/env python3
"""Read-only local tool/source metadata; never builds or scans repositories."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import platform
import stat
import subprocess
import sys
import sysconfig

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
SOURCE_KEYS = ('GoFiles', 'CgoFiles', 'CFiles', 'CXXFiles', 'MFiles', 'HFiles',
               'FFiles', 'SFiles', 'SwigFiles', 'SwigCXXFiles', 'SysoFiles', 'EmbedFiles')
PACKAGE_KEYS = ('ImportPath', 'Dir', 'Name', 'Standard', 'Goroot', 'Module',
                'Imports', 'ImportMap', 'EmbedPatterns', *SOURCE_KEYS)


def sha(path):
    with Path(path).open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


def safe_resolve(path):
    pending = list(Path(path).absolute().parts[1:])
    current = Path('/')
    links = 0
    while pending:
        part = pending.pop(0)
        if part in ('', '.'):
            continue
        if part == '..':
            current = current.parent
            continue
        if part == '.codegraph':
            raise ValueError('.codegraph access refused')
        current = current / part
        if current.is_symlink():
            links += 1
            if links > 100:
                raise ValueError('symlink cycle')
            target = Path(os.readlink(current))
            current = Path('/') if target.is_absolute() else current.parent
            pending = list(target.parts[1:] if target.is_absolute() else target.parts) + pending
    return current


def install_tree(root):
    root = safe_resolve(root)
    files, links, visited = {}, {}, set()
    def visit(path):
        if path.name == '.codegraph':
            raise ValueError('.codegraph access refused')
        resolved = safe_resolve(path)
        if not resolved.is_relative_to(root):
            raise ValueError('install symlink escapes root: ' + str(path))
        if path.is_symlink():
            links[str(path)] = {'target': os.readlink(path), 'resolved': str(resolved)}
        if resolved in visited:
            return
        visited.add(resolved)
        mode = resolved.stat().st_mode
        if stat.S_ISDIR(mode):
            for child in sorted(resolved.iterdir()):
                visit(child)
        elif stat.S_ISREG(mode):
            files[str(resolved)] = sha(resolved)
        else:
            raise ValueError('unsupported install entry: ' + str(path))
    visit(root)
    return {'root': str(root), 'files': dict(sorted(files.items())),
            'symlinks': dict(sorted(links.items()))}


def scanner_tree(root):
    root = safe_resolve(root)
    for name in ('dist', 'node_modules', 'package.json', 'pnpm-lock.yaml'):
        if not (root / name).exists():
            raise ValueError('incomplete revision install: ' + name)
    return {'schema': 'x11a-clawpatch-tree-v1', **install_tree(root)}


def python_runtime():
    python = safe_resolve(sys.executable)
    return {'schema': 'x11a-python-runtime-v2', 'path': str(python), 'sha256': sha(python),
            'version': sys.version, 'implementation': sys.implementation.name,
            'cache_tag': sys.implementation.cache_tag, 'executable_argv': [str(python), '-I', '-B'],
            'install': install_tree(sys.base_prefix),
            'stdlib': sysconfig.get_path('stdlib'), 'local_module_paths': [str(HERE)]}


def load_pin(pin):
    if sha(pin['path']) != pin['sha256']:
        raise ValueError('metadata digest mismatch: ' + pin['path'])
    return json.loads(Path(pin['path']).read_text())


def verify_python_runtime(prep, tools=None):
    prep = prep.get('export_preparation', prep)
    tools = tools or prep['tools']
    runtime = load_pin(prep['python_runtime'])
    if runtime['schema'] != 'x11a-python-runtime-v2':
        raise ValueError('unsupported Python runtime manifest')
    if safe_resolve(sys.executable) != Path(runtime['path']) or sha(sys.executable) != runtime['sha256']:
        raise ValueError('Python executable drift')
    if tools['python'] != {'path': runtime['path'], 'sha256': runtime['sha256']}:
        raise ValueError('Python tool/runtime pin disagreement')
    if not sys.flags.isolated or not sys.dont_write_bytecode:
        raise ValueError('Python requires isolated -I -B launch')
    if runtime['install'] != install_tree(runtime['install']['root']):
        raise ValueError('Python whole-install closure drift')
    if runtime['local_module_paths'] != [str(HERE)]:
        raise ValueError('Python local module paths mismatch')
    if any(HERE.rglob('__pycache__')) or any(HERE.glob('*.pyc')):
        raise ValueError('stale local bytecode cache; remove before isolated launch')
    return runtime


def verify_preparation_closure(config):
    """Check the reviewed immutable tool tree and OS closure before entrypoint code."""
    metadata_path = HERE / 'tool-pins-v6' / 'persistent-locations-v6b.json'
    entries = [entry for entry in config['recipe_sources']
               if Path(entry['path']).resolve() == metadata_path]
    if len(entries) != 1:
        raise ValueError('missing reviewed immutable tool/platform closure')
    metadata = load_pin(entries[0])
    tools, expected_platform = metadata['tools'], metadata['platform']
    if config['export_preparation']['tools'] != tools['executables']:
        raise ValueError('configured executable/immutable closure disagreement')
    actual_platform = {
        'architecture': platform.machine(),
        'macOS': subprocess.check_output(['/usr/bin/sw_vers', '-productVersion']).decode().strip(),
        'build': subprocess.check_output(['/usr/bin/sw_vers', '-buildVersion']).decode().strip(),
    }
    if any(actual_platform[key] != expected_platform[key] for key in actual_platform):
        raise ValueError('OS build drift; council micro-round required')
    sandbox = expected_platform['sandbox_exec']
    if sha(sandbox['path']) != sandbox['sha256']:
        raise ValueError('OS sandbox binary drift; council micro-round required')
    for path, expected in tools['OS_script_interpreters'].items():
        if sha(path) != expected['sha256']:
            raise ValueError('OS interpreter drift; council micro-round required')
    for label, installation in (('Git/Node/Clawpatch', tools),
                                ('Python', metadata['python_install'])):
        root = Path(installation['root'])
        actual = install_tree(root)
        expected_files = {str(root / path): value['sha256']
                          for path, value in installation['files'].items()}
        links = {str(Path(path).relative_to(root)): value['target']
                 for path, value in actual['symlinks'].items()}
        if actual['files'] != expected_files or links != installation['symlinks']:
            raise ValueError('immutable ' + label + ' installation closure drift')
        for path, expected in installation['files'].items():
            mode = stat.S_IMODE((root / path).stat().st_mode)
            if mode != (0o555 if expected['executable'] else 0o444):
                raise ValueError('immutable tool file mode drift')
        for path in (root, *root.rglob('*')):
            if not path.is_symlink() and path.is_dir() and stat.S_IMODE(path.stat().st_mode) != 0o555:
                raise ValueError('immutable tool directory mode drift')


def go_metadata(recipe, environment=None, runner=None):
    env = environment if environment is not None else recipe['environment']
    go = recipe['go']['path']
    if sha(go) != recipe['go']['sha256']:
        raise ValueError('Go executable drift')
    def command(*args):
        argv = [go, *args]
        result = (runner(argv, recipe['cwd'], env) if runner else
                  subprocess.check_output(argv, cwd=recipe['cwd'], env=env))
        return result.decode() if isinstance(result, bytes) else result
    target = recipe['argv'][-1]
    if '-pgo=off' not in recipe['argv']:
        raise ValueError('Go recipe must disable automatic PGO')
    argv = [go, 'list', '-pgo=off', '-deps', '-json', target]
    raw = command(*argv[1:])
    decoder, packages, offset = json.JSONDecoder(), [], 0
    while offset < len(raw):
        while offset < len(raw) and raw[offset].isspace():
            offset += 1
        if offset == len(raw):
            break
        package, offset = decoder.raw_decode(raw, offset)
        if package.get('Error') or package.get('DepsErrors') or package.get('Incomplete'):
            raise ValueError('incomplete Go package closure')
        # Stale/BuildID/Target depend on cache contents, not compiler input selection.
        packages.append({key: package[key] for key in PACKAGE_KEYS if key in package})
    packages.sort(key=lambda package: package['ImportPath'])
    files = {str(Path(recipe['cwd']) / name): sha(Path(recipe['cwd']) / name)
             for name in ('go.mod', 'go.sum')}
    for package in packages:
        for key in SOURCE_KEYS:
            for name in package.get(key, []):
                path = safe_resolve(Path(package['Dir']) / name)
                files[str(path)] = sha(path)
        module = package.get('Module', {})
        for source in (module, module.get('Replace', {})):
            if source.get('GoMod'):
                path = safe_resolve(source['GoMod'])
                files[str(path)] = sha(path)
    goroot = Path(command('env', 'GOROOT').strip())
    for directory in (goroot / 'pkg' / 'tool', goroot / 'pkg' / 'include'):
        for path in directory.rglob('*'):
            if '.codegraph' in path.parts:
                raise ValueError('unexpected .codegraph in toolchain')
            if path.is_file():
                files[str(safe_resolve(path))] = sha(path)
    for path in (Path(go), goroot / 'VERSION', goroot / 'go.env'):
        files[str(path)] = sha(path)
    return {'schema': 'x11a-go-build-closure-v2', 'go_version': command('version').strip(),
            'metadata_argv': argv, 'environment': recipe['environment'],
            'packages': packages, 'files': dict(sorted(files.items()))}


def verify_go_closure(recipe, expected_closure, *, environment=None, runner=None):
    actual = go_metadata(recipe, environment, runner)
    if actual != expected_closure:
        raise ValueError('Go selected package/file-set closure mismatch')
    return actual


def emit(path, data):
    with Path(path).open('x') as stream:
        json.dump(data, stream, sort_keys=True, indent=2)
        stream.write('\n')


def collect(go, scanner, output):
    output = Path(output).absolute()
    output.mkdir(parents=True, exist_ok=True)
    emit(output / 'clawpatch-tree-v6.json', scanner_tree(scanner))
    emit(output / 'python-runtime-v6.json', python_runtime())
    go = safe_resolve(go)
    private = HERE / 'artifacts' / 'preparation-runtime'
    env = {'HOME': str(private / 'home'), 'GOCACHE': str(private / 'gocache'),
           'GOTMPDIR': str(private / 'tmp'), 'PATH': str(go.parent), 'LANG': 'C',
           'GOTOOLCHAIN': 'local', 'GOFLAGS': '-mod=readonly', 'GOPROXY': 'off',
           'GOSUMDB': 'off', 'CGO_ENABLED': '0', 'GOWORK': 'off', 'GOENV': 'off',
           'GOOS': 'darwin', 'GOARCH': 'arm64', 'GOEXPERIMENT': '',
           'GOAMD64': 'v1', 'GOARM64': 'v8.0', 'GIT_NO_REPLACE_OBJECTS': '1'}
    # GOMODCACHE is an input store, never the private build-output cache.
    env['GOMODCACHE'] = str(Path.home() / 'go' / 'pkg' / 'mod')
    for key in ('HOME', 'GOCACHE', 'GOTMPDIR'):
        Path(env[key]).mkdir(parents=True, exist_ok=True)
    binary = str(HERE / 'artifacts' / 'frontend-v6' / 'frontend')
    recipe = {'schema': 'x11a-frontend-build-recipe-v1', 'cwd': str(ROOT),
              'environment': env, 'clear_environment': True,
              'go': {'path': str(go), 'sha256': sha(go)},
              'closure': {'path': str(output / 'frontend-go-closure-v6.json')},
              'argv': [str(go), 'build', '-pgo=off', '-trimpath', '-buildvcs=false', '-o', binary,
                       './evaluation/x11-merge-association/frontend'],
              'version_m_argv': [str(go), 'version', '-m', binary],
              'required_buildinfo_path': str(HERE / 'artifacts' / 'frontend-v6' / 'go-version-m.txt'),
              'execution_policy': 'Council-authorized preparation only; refresh exact package/file closure immediately before build in the same private environment; record buildinfo.'}
    emit(output / 'frontend-go-closure-v6.json', go_metadata(recipe))
    recipe['closure']['sha256'] = sha(recipe['closure']['path'])
    emit(output / 'frontend-build-recipe-v6.json', recipe)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--go', required=True)
    parser.add_argument('--scanner-root', required=True)
    parser.add_argument('--output', required=True)
    args = parser.parse_args()
    collect(args.go, args.scanner_root, args.output)


if __name__ == '__main__':
    main()

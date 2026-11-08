#!/usr/bin/env python3
"""Council-gated, offline X11 ownership and identifier export recipes."""
import argparse
import errno
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import re
import shutil
import subprocess
from datetime import datetime, timezone
import tempfile
import sys
import unicodedata
import platform
import signal
import time
import fcntl
from datetime import timedelta

SUBPROCESS_SECONDS = 3600
ATTEMPT_SECONDS = 14 * 86400
PREPARATION_SECONDS = 42 * 86400
MAX_REPOSITORY_CONCURRENCY = 5
ACTIVE_DEADLINE = None
FAILED_STEP = None


def canonical_budget(inputs):
    return Path(inputs).resolve().parent / 'artifacts' / 'preparation-budget-v7.json'


def canonical_paths(inputs, operation, repository=None):
    """Return the create-once (output, seal) paths for an authorised operation."""
    require(operation in {'ownership', 'build-frontend', 'exposures'}, 'unknown preparation operation')
    tag = operation
    if operation == 'ownership':
        require(repository in {r['repository'] for r in load(Path(inputs))['repositories']},
                'unknown preparation repository')
        require(re.fullmatch(r'[A-Za-z0-9-]+/[A-Za-z0-9_.-]+', repository) is not None,
                'invalid preparation repository')
        tag += '--' + repository.replace('/', '__')
    else:
        require(repository is None, 'repository only allowed for ownership')
    directory = canonical_budget(inputs).parent / 'preparation-v8'
    return directory / (tag + '.output.json'), directory / (tag + '.seal.json')


def canonical_seals(inputs):
    config = load(Path(inputs))
    return [canonical_paths(inputs, 'ownership', item['repository'])[1]
            for item in config['repositories']] + [
                canonical_paths(inputs, operation)[1] for operation in ('build-frontend', 'exposures')]


def require_no_global_stop(inputs):
    for path in canonical_seals(inputs):
        if not path.exists():
            continue
        # An empty reservation belongs to an operation still running.
        text = path.read_text()
        if not text:
            continue
        record = json.loads(text)
        require(not (record.get('status') == 'refused' and record.get('refusal', {}).get('code') in {
            'PreparationInterruptedOperational', 'PreparationBudgetExhausted'}),
            'X11a not testable (operational): study-wide interruption/deadline; no candidate exclusion or substitution')



def preparation_budget(path, inputs, authority):
    path = absolute(str(path))
    require(path == canonical_budget(inputs),
            'shared preparation budget must use canonical authorised-input path')
    value = load(path)
    require(set(value) == {'schema', 'inputs', 'authorization', 'started_utc', 'deadline_utc',
                          'subprocess_seconds', 'attempt_seconds', 'preparation_seconds',
                          'max_repository_concurrency'}, 'invalid shared preparation budget')
    require(value['schema'] == 'x11a-preparation-budget-v7'
            and value['inputs'] == file_pin(inputs) and value['authorization'] == file_pin(authority),
            'shared budget authority/input mismatch')
    require((value['subprocess_seconds'], value['attempt_seconds'], value['preparation_seconds'],
             value['max_repository_concurrency']) ==
            (SUBPROCESS_SECONDS, ATTEMPT_SECONDS, PREPARATION_SECONDS, MAX_REPOSITORY_CONCURRENCY),
            'shared preparation policy changed')
    start, end = (datetime.fromisoformat(value[k]) for k in ('started_utc', 'deadline_utc'))
    require(start.utcoffset().total_seconds() == end.utcoffset().total_seconds() == 0
            and (end - start).total_seconds() == PREPARATION_SECONDS
            and start.timestamp() <= time.time(), 'invalid persistent preparation deadline')
    return value


def start_budget(path, inputs, authority):
    require(absolute(str(path)) == canonical_budget(inputs),
            'budget output must use canonical authorised-input path')
    absolute(str(path)).parent.mkdir(parents=True, exist_ok=True)
    authorise(inputs, authority)
    started = datetime.now(timezone.utc)
    value = {'schema': 'x11a-preparation-budget-v7', 'inputs': file_pin(inputs),
             'authorization': file_pin(authority), 'started_utc': started.isoformat(),
             'deadline_utc': (started + timedelta(seconds=PREPARATION_SECONDS)).isoformat(),
             'subprocess_seconds': SUBPROCESS_SECONDS, 'attempt_seconds': ATTEMPT_SECONDS,
             'preparation_seconds': PREPARATION_SECONDS,
             'max_repository_concurrency': MAX_REPOSITORY_CONCURRENCY}
    with absolute(str(path)).open('x') as stream:
        json.dump(value, stream, indent=2, sort_keys=True)
        stream.write('\n')
    return value




def deadline_alarm(signum, frame):
    raise PreparationDeadline('preparation wall deadline reached')

HERE = Path(__file__).resolve().parent
SHA = re.compile(r'[0-9a-f]{40}')
DIGEST = re.compile(r'[0-9a-f]{64}')
TRACE = []
REVISION_COUNT = 0
ATTEMPT = 0
CURRENT_REVISION = None
NODE_HEAP_MB = 8192
SANDBOX_PROFILE = '(version 1)(allow default)(deny network*)(deny file-read* file-write* (regex #"/[.]codegraph(/|$)"))'
GIT_TEMPLATE = ['{git}', '-c', 'core.hooksPath=/dev/null', '-c', 'credential.helper=', '-C', '{repository_path}']
SCANNER_TEMPLATE = ['{node}', '{clawpatch}', '--root', '{materialized_tree}', '--state-dir', '{fresh_state}', '--json']
TEMPLATES = {
    'sandbox': ['{sandbox_exec}', '-p', SANDBOX_PROFILE, '{step_argv...}'],
    'tree': [*GIT_TEMPLATE, 'ls-tree', '-r', '-z', '{revision}'],
    'blob': [*GIT_TEMPLATE, 'cat-file', '--batch'],
    'chain': [*GIT_TEMPLATE, 'rev-list', '--first-parent', '--reverse', '{tip}'],
    'version': ['{node}', '{clawpatch}', '--version'],
    'init': [*SCANNER_TEMPLATE, 'init'],
    'map': [*SCANNER_TEMPLATE, 'map', '--source', 'heuristic'],
    'go_list': ['{go}', 'list', '-deps', '-json', '{reviewed_recipe.argv[-1]}'],
    'go_root': ['{go}', 'env', 'GOROOT'],
    'go_version': ['{go}', 'version'],
    'build': ['{reviewed_recipe.argv...}'],
    'buildinfo': ['{reviewed_recipe.version_m_argv...}'],
}


class ContentRefusal(ValueError):
    """A deterministic property of pinned repository content."""


class OperationalRefusal(ValueError):
    """Repairable runtime, tool, authority or location failure; never ineligible."""


class PreparationDeadline(OperationalRefusal):
    """A fixed wall-clock cap, never a content property."""

class CommandFailure(OperationalRefusal):
    def __init__(self, result, argv):
        super().__init__('offline pinned command failed (exit %d)' % result.returncode)
        self.signature = (tuple(map(str, argv)), result.returncode,
                          hashlib.sha256(result.stdout).hexdigest(),
                          hashlib.sha256(result.stderr).hexdigest())


def utc():
    return datetime.now(timezone.utc).isoformat()


def file_pin(path):
    return {'path': str(path.resolve()), 'sha256': digest(path)}


def content(condition, message):
    if not condition:
        raise ContentRefusal(message)


def require(condition, message):
    if not condition:
        raise ValueError(message)


def absolute(value):
    require(isinstance(value, str) and Path(value).is_absolute(), 'absolute path required')
    require('.codegraph' not in Path(value).parts, '.codegraph is excluded')
    return Path(value)


def digest(path):
    with path.open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


def load(path):
    return json.loads(path.read_text())


def pin(entry):
    require(set(entry) == {'path', 'sha256'}, 'pin requires exactly path and sha256')
    path = absolute(entry['path'])
    require(isinstance(entry['sha256'], str) and DIGEST.fullmatch(entry['sha256']), 'missing exact SHA-256 pin')
    require(path.is_file() and digest(path) == entry['sha256'], 'pinned file missing or changed: ' + str(path))
    return path.resolve()


def relative_path(value):
    require(isinstance(value, str) and value and '\\' not in value and '\x00' not in value, 'invalid identifier path')
    path = PurePosixPath(value)
    require(not path.is_absolute() and all(p not in {'', '.', '..', '.codegraph'} for p in value.split('/')), 'unsafe identifier path')
    return value


def authorise(inputs, authorization):
    from stages import verify_authorization
    auth = verify_authorization(authorization, inputs, scope='export-preparation')
    config = load(inputs)
    prep = config['export_preparation']
    require(prep['status'] == 'PINNED', 'export preparation source pins are not PINNED')
    require('PENDING' not in json.dumps(config).upper(), 'unresolved preparation pin')
    signed = {str(Path(entry['path']).resolve()): entry['sha256'] for entry in auth['files']}
    def signed_file(path):
        require(signed.get(str(path.resolve())) == digest(path), 'file not bound by export authorisation: ' + str(path))
    for name in ('prepare_exports.py', 'stages.py', 'inventory.py', 'pin_tool_closure.py'):
        signed_file(HERE / name)
    tools = {name: pin(prep['tools'][name]) for name in ('git', 'node', 'clawpatch', 'clawpatch_safety_marker', 'sandbox_exec', 'python')}
    require(Path(sys.executable).resolve() == tools['python'], 'use pinned Python interpreter')
    from pin_tool_closure import verify_python_runtime
    signed_file(pin(prep['python_runtime']))
    verify_python_runtime(prep)
    for name in ('clawpatch_tree', 'frontend_build_recipe'):
        signed_file(pin(prep[name]))
    tree = load(pin(prep['clawpatch_tree']))
    require(tree['root'] == str(tools['clawpatch'].parent.parent), 'scanner tree root mismatch')
    for entry in prep['tools'].values():
        signed_file(absolute(entry['path']))
    return config, prep, tools, signed_file


def environment(home, bin_dir):
    # Deliberate allowlist: no keys, proxy routes, NODE_OPTIONS, model CLI
    # configuration, Git overrides, credential helpers, or inherited PATH/HOME.
    git_root = (bin_dir / 'git').resolve().parent.parent
    return {'HOME': str(home), 'TMPDIR': str(home), 'PATH': str(bin_dir),
            'LANG': 'C.UTF-8', 'LC_ALL': 'C.UTF-8', 'CI': '1',
            'GIT_CONFIG_NOSYSTEM': '1', 'GIT_CONFIG_GLOBAL': '/dev/null',
            'GIT_TERMINAL_PROMPT': '0', 'GIT_NO_LAZY_FETCH': '1', 'GIT_NO_REPLACE_OBJECTS': '1',
            'GIT_ALLOW_PROTOCOL': '', 'GIT_OPTIONAL_LOCKS': '0',
            'GIT_CEILING_DIRECTORIES': '/', 'GIT_DISCOVERY_ACROSS_FILESYSTEM': '0',
            'GIT_EXEC_PATH': str(git_root / 'libexec/git-core'),
            'GIT_TEMPLATE_DIR': str(git_root / 'share/git-core/templates'),
            'NODE_OPTIONS': '--max-old-space-size=%d' % NODE_HEAP_MB}


def run(tools, env, argv, cwd, input_bytes=None, stdout_file=None, count=1, oid_list=None):
    command = [str(tools['sandbox_exec']), '-p', SANDBOX_PROFILE, *map(str, argv)]
    started = utc()
    remaining = SUBPROCESS_SECONDS if ACTIVE_DEADLINE is None else min(
        SUBPROCESS_SECONDS, ACTIVE_DEADLINE - time.time())
    if remaining <= 0:
        raise PreparationDeadline('preparation wall deadline reached before subprocess')
    process = subprocess.Popen(command, cwd=cwd, env=env, stdin=subprocess.PIPE if input_bytes is not None else None,
                               stdout=stdout_file if stdout_file is not None else subprocess.PIPE,
                               stderr=subprocess.PIPE, start_new_session=True)
    failure = None
    try:
        stdout, stderr = process.communicate(input_bytes, timeout=remaining)
    except BaseException as error:
        failure = error
        try:
            os.killpg(process.pid, signal.SIGKILL)
        except ProcessLookupError:
            pass
        stdout, stderr = process.communicate(timeout=10)
    result = subprocess.CompletedProcess(command, process.returncode, stdout, stderr)
    if stdout_file is None:
        stdout_digest = hashlib.sha256(result.stdout).hexdigest()
    else:
        stdout_file.flush(); stdout_file.seek(0)
        stdout_digest = hashlib.file_digest(stdout_file, 'sha256').hexdigest()
        stdout_file.seek(0)
    TRACE.append({'argv': command, 'cwd': str(cwd), 'returncode': result.returncode,
                  'stdout_sha256': stdout_digest,
                  'stderr_sha256': hashlib.sha256(result.stderr).hexdigest(),
                  'started_utc': started, 'ended_utc': utc(), 'attempt': ATTEMPT,
                  'revision': CURRENT_REVISION, 'count': count,
                  'oid_list_sha256': hashlib.sha256(oid_list or b'').hexdigest()})
    if failure is not None:
        if isinstance(failure, subprocess.TimeoutExpired):
            raise OperationalRefusal('subprocess timeout: whole process group killed') from failure
        raise failure
    if result.returncode:
        if result.stdout is None:
            result.stdout = b''
        raise CommandFailure(result, command)
    return result.stdout


def git(tools, env, repo, *args, input_bytes=None):
    return run(tools, env, [tools['git'], '-c', 'core.hooksPath=/dev/null', '-c', 'credential.helper=', '-C', repo, *args], repo, input_bytes)


def repository_prerequisites(tools, env, repo):
    common = Path(git(tools, env, repo, 'rev-parse', '--path-format=absolute', '--git-common-dir').decode().strip())
    require(not (common / 'shallow').exists() and not (common / 'info/grafts').exists(),
            'operational: shallow repository or grafts refused')
    require(not git(tools, env, repo, 'for-each-ref', '--format=%(refname)', 'refs/replace').strip(),
            'operational: replacement refs refused')
    with tempfile.TemporaryFile() as configuration:
        run(tools, env, [tools['git'], '-C', str(repo), 'config', '--null', '--list'],
            repo, stdout_file=configuration)
        settings = configuration.read().lower()
    require(not re.search(rb'(?:^|\x00)(?:extensions\.partialclone|remote\.[^\x00\n]+\.promisor)\n', settings)
            and not list((common / 'objects/pack').glob('*.promisor')),
            'operational: partial/promisor repository refused')


def materialize(tools, env, repo, commit, root):
    """One tree listing and one disk-backed batch per revision, exact blob bytes."""
    global FAILED_STEP
    FAILED_STEP = 'ls-tree'
    raw = git(tools, env, repo, 'ls-tree', '-r', '-z', commit)
    entries, excluded = [], 0
    for record in raw.split(b'\0'):
        if not record:
            continue
        header, name = record.split(b'\t', 1)
        mode, kind, oid = header.decode('ascii').split()
        try:
            name = name.decode('utf-8')
        except UnicodeDecodeError:
            raise ContentRefusal('non-UTF-8 committed path')
        if '.codegraph' in name.split('/'):
            excluded += 1
            continue
        content(mode == '160000' and kind == 'commit' or
                mode in {'100644', '100755', '120000'} and kind == 'blob', 'unsupported tree entry')
        content(SHA.fullmatch(oid), 'invalid blob identity')
        entries.append((name, mode, oid))
    FAILED_STEP = 'cat-file'
    oids = list(dict.fromkeys(oid for _, mode, oid in entries if mode != '160000'))
    oid_list = ''.join(oid + '\n' for oid in oids).encode()
    with tempfile.TemporaryFile() as batch:
        run(tools, env, [tools['git'], '-C', repo, 'cat-file', '--batch'], repo,
            oid_list, stdout_file=batch, count=len(oids), oid_list=oid_list)
        batch_stdout_sha256 = TRACE[-1]['stdout_sha256']
        offsets = {}
        for oid in oids:
            fields = batch.readline().decode('ascii').split()
            require(len(fields) == 3 and fields[:2] == [oid, 'blob'], 'batch blob identity mismatch')
            size = int(fields[2])
            offsets[oid] = (batch.tell(), size)
            batch.seek(size, 1)
            require(batch.read(1) == b'\n', 'batch blob framing mismatch')
        require(batch.read(1) == b'', 'unexpected batch output')
        def blob(oid):
            offset, size = offsets[oid]
            batch.seek(offset)
            data = batch.read(size)
            require(len(data) == size, 'truncated batch blob')
            return data
        materialize_entries(entries, root, blob)
    metadata = {'excluded_codegraph_entries': excluded,
                'gitlinks': sum(mode == '160000' for _, mode, _ in entries),
                'entry_count': len(entries), 'blob_count': len(oids),
                'tree_list_sha256': hashlib.sha256(raw).hexdigest(),
                'batch_stdout_sha256': batch_stdout_sha256,
                'oid_list_sha256': hashlib.sha256(oid_list).hexdigest()}
    TRACE.append({'step': 'materialized', 'revision': commit, 'attempt': ATTEMPT,
                  'count': len(entries), **metadata})
    return metadata


def materialize_entries(entries, root, blob):
    # Aliasing is a property of committed paths, never a scratch-FS probe.
    seen, leaves, data = {}, set(), {}
    for name, mode, oid in entries:
        try:
            relative_path(name)
        except ValueError as error:
            raise ContentRefusal(str(error)) from error
        content(mode in {'100644', '100755', '120000', '160000'}, 'unsupported tree mode')
        content(name not in leaves, 'duplicate tree path')
        leaves.add(name)
        for parent in [PurePosixPath(name), *PurePosixPath(name).parents]:
            if str(parent) == '.':
                continue
            key = unicodedata.normalize('NFD', str(parent)).casefold()
            content(key not in seen or seen[key] == str(parent),
                    'NFD/case path alias')
            seen[key] = str(parent)
        if mode == '120000':
            try:
                target = blob(oid).decode('utf-8')
            except UnicodeDecodeError:
                raise ContentRefusal('non-UTF-8 symlink')
            content(target and '\0' not in target and not PurePosixPath(target).is_absolute(),
                    'absolute or invalid symlink refused')
            resolved = (root / name).parent.joinpath(target).resolve()
            content(resolved.is_relative_to(root.resolve()) and '.codegraph' not in resolved.parts,
                    'escaping or excluded symlink refused')
            data[name] = target
    for name in leaves:
        content(not any(str(parent) in leaves for parent in PurePosixPath(name).parents),
                'file/symlink is ancestor of committed path')
    for name, mode, oid in entries:
        destination = root / name
        destination.parent.mkdir(parents=True, exist_ok=True)
        if mode == '160000':
            destination.mkdir()
        elif mode != '120000':
            with destination.open('xb') as stream:
                stream.write(blob(oid))
            destination.chmod(0o755 if mode == '100755' else 0o644)
    for name, mode, oid in entries:
        if mode == '120000':
            (root / name).symlink_to(data[name])
    for name in data:
        try:
            resolved = (root / name).resolve()
        except RuntimeError:
            raise ContentRefusal('cyclic symlink chain')
        except OSError as error:
            if error.errno == errno.ELOOP:
                raise ContentRefusal('cyclic symlink chain') from error
            raise
        content(resolved.is_relative_to(root.resolve()) and '.codegraph' not in resolved.parts,
                'unsafe symlink chain')
        # Python 3.14 resolve(strict=False) may leave a loop unresolved.
        # stat detects ELOOP without rejecting safe dangling relative links.
        try:
            (root / name).stat()
        except FileNotFoundError:
            pass
        except OSError as error:
            if error.errno == errno.ELOOP:
                raise ContentRefusal('cyclic symlink chain') from error
            raise
    verify_materialized(entries, root)


def verify_materialized(entries, root):
    """Verify selected entries; post-write drift is operational, not repository content."""
    expected = {}
    for name, mode, oid in entries:
        expected[name] = (mode, oid)
        for parent in PurePosixPath(name).parents:
            if str(parent) != '.':
                expected.setdefault(str(parent), ('040000', None))
    actual = set()
    for directory, dirs, files in os.walk(root, followlinks=False):
        for name in [*dirs, *files]:
            path = Path(directory) / name
            relative = path.relative_to(root).as_posix()
            require('.codegraph' not in path.relative_to(root).parts,
                    'unexpected excluded entry in materialized tree')
            actual.add(relative)
    require(actual == set(expected), 'materialized selected path set mismatch')
    for name, (mode, oid) in expected.items():
        path = root / name
        if mode in {'040000', '160000'}:
            require(path.is_dir() and not path.is_symlink(), 'materialized directory type mismatch')
            continue
        if mode == '120000':
            require(path.is_symlink(), 'materialized symlink type mismatch')
            data = os.readlink(path).encode('utf-8')
            actual_oid = hashlib.sha1(b'blob ' + str(len(data)).encode() + b'\0' + data).hexdigest()
        else:
            require(path.is_file() and not path.is_symlink(), 'materialized file type mismatch')
            require(path.stat().st_mode & 0o777 == (0o755 if mode == '100755' else 0o644),
                    'materialized file mode mismatch')
            identity = hashlib.sha1(b'blob ' + str(path.stat().st_size).encode() + b'\0')
            with path.open('rb') as stream:
                while block := stream.read(1024 * 1024):
                    identity.update(block)
            actual_oid = identity.hexdigest()
        require(actual_oid == oid, 'materialized blob bytes or symlink target mismatch')


def verify_scanner(prep, tools):
    from pin_tool_closure import scanner_tree
    expected = load(pin(prep['clawpatch_tree']))
    require(scanner_tree(tools['clawpatch'].parent.parent) == expected, 'whole scanner install changed')




def heuristic_paths(state):
    # Differential fixture compares these accepted canonical records with the
    # product LoadFeatures + !areas.IsAgent path, including absent owned paths.
    paths = set()
    features = state / 'features'
    require(features.is_dir(), 'mapper feature directory missing')
    for file in sorted(features.glob('*.json')):
        require(file.is_file() and not file.is_symlink(), 'invalid feature record')
        feature = load(file)
        require(feature.get('schemaVersion') == 1 and isinstance(feature.get('featureId'), str), 'invalid feature identity')
        if feature.get('status') == 'skipped' or feature.get('source') == 'agent-mapper':
            continue
        require(isinstance(feature.get('ownedFiles'), list), 'ownedFiles missing')
        for owned in feature['ownedFiles']:
            paths.add(relative_path(owned['path']))
    return sorted(paths)

def ownership(args, config, prep, tools):
    global REVISION_COUNT, CURRENT_REVISION
    matches = [r for r in config['repositories'] if r['repository'] == args.repository]
    require(len(matches) == 1, 'repository must match one pinned input')
    item = matches[0]
    repo = absolute(item['path'])
    for name in ('Cold', 'C0', 'tip'):
        require(isinstance(item[name], str) and SHA.fullmatch(item[name]), 'full commit pin required')
    revision = prep['clawpatch_revision']
    require(SHA.fullmatch(revision) and tools['clawpatch'].parent.name == 'dist', 'not the pinned safe Clawpatch entrypoint')
    require(tools['clawpatch_safety_marker'].read_text().strip() == 'scout-review-sanitize-v3 dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3', 'reviewed safety marker required')
    scratch = absolute(args.scratch)
    scratch.mkdir(parents=True, exist_ok=True)
    require(not scratch.resolve().is_relative_to(repo.resolve()), 'scratch must be outside source repository')
    with tempfile.TemporaryDirectory(prefix='x11-exports-', dir=scratch) as work:
        work = Path(work)
        home, bin_dir = work / 'home', work / 'bin'
        home.mkdir(); bin_dir.mkdir()
        (bin_dir / 'git').symlink_to(tools['git'])
        env = environment(home, bin_dir)
        env['GIT_CEILING_DIRECTORIES'] = str(work)
        repository_prerequisites(tools, env, repo)
        verify_scanner(prep, tools)
        version = run(tools, env, [tools['node'], tools['clawpatch'], '--version'], work).decode().strip()
        require(version == prep['clawpatch_version'], 'Clawpatch version mismatch')
        chain = git(tools, env, repo, 'rev-list', '--first-parent', '--reverse', item['tip']).decode().splitlines()
        require(item['C0'] in chain and item['Cold'] in chain, 'mapping anchors not on pinned first-parent chain')
        commits = list(dict.fromkeys([item['Cold'], item['C0'], *chain[chain.index(item['C0']) + 1:][:2000]]))
        REVISION_COUNT = len(commits)
        TRACE.append({'step': 'scope', 'attempt': ATTEMPT, 'revisions': commits,
                      'chain_sha256': TRACE[-1]['stdout_sha256']})
        # The output is a small digest index; each revision has a separate disk record.
        records_dir = Path(str(args.output) + '.records-attempt-%d' % ATTEMPT)
        records_dir.mkdir()
        revisions = {}
        for commit in commits:
            CURRENT_REVISION = commit
            require(SHA.fullmatch(commit), 'invalid commit identifier')
            root, state = work / 'root', work / 'state'
            root.mkdir(); state.mkdir()
            tree_metadata = materialize(tools, env, repo, commit, root)
            argv = [tools['node'], tools['clawpatch'], '--root', root, '--state-dir', state, '--json']
            run(tools, env, [*argv, 'init'], work)
            response = json.loads(run(tools, env, [*argv, 'map', '--source', 'heuristic'], work))
            require(isinstance(response, dict), 'invalid map response')
            paths = heuristic_paths(state)
            record_path = records_dir / (commit + '.json')
            with record_path.open('x') as stream:
                json.dump({'revision': commit, 'paths': paths, **tree_metadata}, stream, sort_keys=True)
                stream.write('\n')
            revisions[commit] = file_pin(record_path)
            shutil.rmtree(root); shutil.rmtree(state)
        verify_scanner(prep, tools)
        CURRENT_REVISION = None
    return {'schema': 'x11a-ownership-index-v6', 'source': 'heuristic', 'repository': item['repository'],
            'scanner_sha256': prep['tools']['clawpatch']['sha256'],
            'scanner_tree_sha256': prep['clawpatch_tree']['sha256'],
            'revision_order': commits, 'revisions': revisions}

def build_frontend(prep, tools, signed_file):
    recipe_path = pin(prep['frontend_build_recipe'])
    recipe = load(recipe_path)
    require(recipe['schema'] == 'x11a-frontend-build-recipe-v1' and recipe['clear_environment'] is True,
            'reviewed hermetic frontend build recipe required')
    go = pin(recipe['go'])
    require(recipe['go'] == prep['tools']['go'], 'build Go tool pin mismatch')
    closure_path = pin(recipe['closure'])
    signed_file(closure_path)
    closure = load(closure_path)
    require(closure['schema'] == 'x11a-go-build-closure-v2', 'wrong Go closure schema')
    for path, expected in closure['files'].items():
        pin({'path': path, 'sha256': expected})
    cwd = absolute(recipe['cwd'])
    binary = absolute(recipe['argv'][recipe['argv'].index('-o') + 1])
    buildinfo = absolute(recipe['required_buildinfo_path'])
    require(binary.parent == buildinfo.parent and binary.name != buildinfo.name,
            'atomic build outputs require one dedicated directory')
    require(not binary.parent.exists(), 'frontend output directory is create-once')
    require(recipe['argv'][0] == str(go) and recipe['version_m_argv'] == [str(go), 'version', '-m', str(binary)],
            'build command tool/output mismatch')
    binary.parent.parent.mkdir(parents=True, exist_ok=True)
    # Refresh package/file selection with precisely the build environment.
    from pin_tool_closure import verify_go_closure
    runtime_root = HERE / 'artifacts' / 'preparation-runtime'
    runtime_root.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix='x11-build-', dir=runtime_root) as private:
        private = Path(private)
        env = dict(recipe['environment'])
        for key in ('HOME', 'GOCACHE', 'GOTMPDIR'):
            location = private / key.lower()
            location.mkdir()
            env[key] = str(location)
        verify_go_closure(recipe, closure, environment=env,
                          runner=lambda argv, cwd, env: run(tools, env, argv, cwd))
        promotion = private / 'outputs'
        promotion.mkdir()
        staged_binary = promotion / binary.name
        staged_buildinfo = promotion / buildinfo.name
        argv = list(recipe['argv'])
        require('-pgo=off' in argv, 'automatic PGO is forbidden')
        argv[argv.index('-o') + 1] = str(staged_binary)
        run(tools, env, argv, cwd)
        record = run(tools, env, [str(go), 'version', '-m', str(staged_binary)], cwd)
        require(record.strip(), 'go version -m produced no buildinfo')
        staged_buildinfo.write_bytes(record)
        # Rename the directory, never a partially validated binary.
        os.rename(promotion, binary.parent)
    return {'frontend': {'path': str(binary), 'sha256': digest(binary)},
            'frontend_build': {'recipe': prep['frontend_build_recipe'],
                               'closure': recipe['closure'],
                               'buildinfo': {'path': str(buildinfo), 'sha256': digest(buildinfo)}}}


def exposures(config, prep, signed_file):
    # Owner-authorised identifier extraction already exists; its coverage is
    # evidence for council review, not an assertion of historical completeness.
    binding = prep['exposure_extraction']
    seal_path = pin(binding['seal'])
    signed_file(seal_path)
    seal = load(seal_path)
    require(seal['schema'] == 'x11a-exposure-extraction-seal-v2', 'wrong extraction seal')
    require(seal['scout_commit'] == binding['scout_commit'] and seal['inertia_commit'] == binding['inertia_commit'], 'source commit mismatch')
    require(seal['files'] == binding['files'], 'extraction file set mismatch')
    signed_file(pin(seal['recipe']))
    signed_file(pin(seal['contexts']))
    identifiers = None
    for entry in binding['files']:
        file = pin(entry)
        signed_file(file)
        if file.name == 'exposure-identifiers.json':
            identifiers = load(file)
    require(identifiers is not None, 'missing identifier projection')
    from inventory import validate_exposure
    validate_exposure(identifiers, config['exposure_studies'])
    files = {Path(entry['path']).name: entry for entry in binding['files']}
    TRACE.append({'step': 'exposure-bindings', 'attempt': ATTEMPT,
                  'files': binding['files'], 'extraction_seal': binding['seal']})
    return {'exposure': files['exposure-identifiers.json'],
            'unscoped_exposure': files['unattributed-identifiers.json'],
            'repository_exposure': files['additional-identifiers.json'],
            'exposure_studies': config['exposure_studies']}


def script_pins():
    paths = {Path(__file__).resolve()}
    for module in tuple(sys.modules.values()):
        path = getattr(module, '__file__', None)
        if path and Path(path).suffix == '.py' and Path(path).resolve().parent == HERE:
            paths.add(Path(path).resolve())
    return [file_pin(path) for path in sorted(paths)]


def verify_trace(record, result):
    attempts = record['attempts']
    require(isinstance(attempts, list) and 1 <= len(attempts) <= 3, 'missing bounded attempt history')
    for number, attempt in enumerate(attempts, 1):
        require(set(attempt) == {'attempt', 'status', 'code', 'failed_step'}
                and attempt['attempt'] == number and attempt['status'] in {'completed', 'refused'},
                'invalid attempt record')
        require(number == len(attempts) or attempt['status'] == 'refused', 'attempt after success')
    require(attempts[-1]['status'] == record['status'], 'attempt/seal status mismatch')
    require(record['commands'], 'empty self-asserted preparation trace')
    preflights = [entry for entry in record['commands'] if entry.get('step') == 'preflight']
    require([entry['attempt'] for entry in preflights] == list(range(1, len(attempts) + 1)),
            'missing per-attempt preflight trace')
    require(all(entry['inputs'] == record['inputs']
                and entry['authorization'] == record['authorization'] for entry in preflights),
            'preflight authority/input binding mismatch')
    for entry in record['commands']:
        require(type(entry['attempt']) is int and 1 <= entry['attempt'] <= len(attempts),
                'trace attempt mismatch')
    final = [entry for entry in record['commands'] if entry['attempt'] == len(attempts)]
    if record['status'] == 'refused':
        if record['refusal']['class'] == 'operational':
            require((len(attempts) == 3 and record['refusal']['code'] == 'PreparationInfeasibleOperational')
                    or record['refusal']['code'] in {'PreparationInterruptedOperational', 'PreparationBudgetExhausted'},
                    'operational refusal is not terminal')
        else:
            refusal = record['refusal']
            revision = refusal['revision']
            require(SHA.fullmatch(revision or '') and refusal['failed_step'] in {'ls-tree', 'cat-file'},
                    'unstructured content refusal')
            commands = [entry for entry in final if 'argv' in entry]
            tools = record['runtimes']
            prefix = [tools['sandbox_exec']['path'], '-p', SANDBOX_PROFILE]
            require(all(entry['argv'][:3] == prefix for entry in commands), 'unsandboxed refusal trace')
            chains = [entry for entry in commands if entry['argv'][-4:] ==
                      ['rev-list', '--first-parent', '--reverse', record['anchors']['tip']]]
            require(len(chains) == 1, 'content refusal lacks scoped revision chain')
            scopes = [entry for entry in final if entry.get('step') == 'scope']
            require(len(scopes) == 1 and revision in scopes[0]['revisions']
                    and 0 < len(scopes[0]['revisions']) <= 2002
                    and scopes[0]['revisions'][:len(set(record['anchors'][k] for k in ('Cold', 'C0')))]
                    == list(dict.fromkeys(record['anchors'][k] for k in ('Cold', 'C0')))
                    and scopes[0]['chain_sha256'] == chains[0]['stdout_sha256'],
                    'content refusal revision outside traced scope')
            steps = [entry for entry in commands if entry['revision'] == revision]
            require(any(entry['argv'][3] == tools['git']['path'] and entry['returncode'] == 0
                        and entry['argv'][-4:] == ['ls-tree', '-r', '-z', revision]
                        for entry in steps), 'content refusal lacks matching tree trace')
            if refusal['failed_step'] == 'cat-file':
                require(any(entry['argv'][3] == tools['git']['path'] and entry['returncode'] == 0
                            and entry['argv'][-2:] == ['cat-file', '--batch'] for entry in steps),
                        'content refusal lacks matching batch trace')
        return
    commands = [entry for entry in final if 'argv' in entry]
    require(all(entry['returncode'] == 0 for entry in commands), 'failed command in completed attempt')
    tools = record['runtimes']
    prefix = [tools['sandbox_exec']['path'], '-p', SANDBOX_PROFILE]
    require(all(entry['argv'][:3] == prefix for entry in commands), 'unsandboxed trace command')
    if record['operation'] == 'ownership':
        from inventory import ownership_record
        require(0 < record['revision_count'] <= 2002, 'unbounded ownership revisions')
        require(len(result['revision_order']) == record['revision_count'], 'revision count mismatch')
        traced_order = [entry['revision'] for entry in final if entry.get('step') == 'materialized']
        chains = [entry for entry in commands if entry['argv'][-4:] ==
                  ['rev-list', '--first-parent', '--reverse', record['anchors']['tip']]]
        require(len(chains) == 1, 'missing scoped revision chain trace')
        require(traced_order == result['revision_order'], 'materialization revision order mismatch')
        for revision in result['revision_order']:
            owned = ownership_record(result, revision)
            steps = [entry for entry in final if entry.get('revision') == revision]
            require(len(steps) == 5, 'expected tree, batch, materialization, init and map')
            tree, batch, materialized, init, mapped = steps
            require(tree['argv'][-4:] == ['ls-tree', '-r', '-z', revision], 'tree trace mismatch')
            require(tree['argv'][3] == tools['git']['path']
                    and batch['argv'][3] == tools['git']['path']
                    and batch['argv'][-2:] == ['cat-file', '--batch'], 'batch trace mismatch')
            require(materialized['step'] == 'materialized'
                    and materialized['count'] == owned['entry_count']
                    and materialized['oid_list_sha256'] == owned['oid_list_sha256']
                    and batch['oid_list_sha256'] == owned['oid_list_sha256']
                    and batch['count'] == owned['blob_count']
                    and tree['stdout_sha256'] == owned['tree_list_sha256']
                    and batch['stdout_sha256'] == owned['batch_stdout_sha256']
                    and materialized['excluded_codegraph_entries'] == owned['excluded_codegraph_entries']
                    and materialized['gitlinks'] == owned['gitlinks'], 'tree record/trace mismatch')
            for entry, suffix in ((init, ['init']), (mapped, ['map', '--source', 'heuristic'])):
                require(entry['argv'][3:5] == [tools['node']['path'], tools['clawpatch']['path']]
                        and entry['argv'][-len(suffix):] == suffix, 'scanner trace mismatch')
            require(init['argv'][5:-1] == mapped['argv'][5:-3], 'scanner recipe changed')
    elif record['operation'] == 'build-frontend':
        recipe = load(pin(result['frontend_build']['recipe']))
        argv = [entry['argv'][3:] for entry in commands]
        closure = load(pin(recipe['closure']))
        require(any(command == closure['metadata_argv'] for command in argv), 'missing exact Go list recipe')
        builds = [command for command in argv if command[1:2] == ['build']]
        infos = [entry for entry in commands if entry['argv'][4:6] == ['version', '-m']]
        require(len(builds) == len(infos) == 1, 'missing build/buildinfo trace')
        expected = list(recipe['argv'])
        expected[expected.index('-o') + 1] = builds[0][builds[0].index('-o') + 1]
        require(builds[0] == expected and infos[0]['argv'][-1] == expected[expected.index('-o') + 1],
                'build recipe trace mismatch')
        require(infos[0]['stdout_sha256'] == result['frontend_build']['buildinfo']['sha256'],
                'buildinfo trace digest mismatch')
    else:
        bindings = [entry for entry in final if entry.get('step') == 'exposure-bindings']
        require(len(bindings) == 1 and bindings[0]['extraction_seal']
                == load(pin(record['inputs']))['export_preparation']['exposure_extraction']['seal'],
                'missing exposure binding trace')


def verify_preparation_seal(path, config, inputs_path, authorization_path):
    """Verify provenance and all materialized outputs, not handwritten assertions."""
    seal_path = pin(path) if isinstance(path, dict) else absolute(str(path))
    record = load(seal_path)
    keys = {'schema', 'stage', 'operation', 'status', 'authorization', 'inputs',
            'scripts', 'argv', 'command_templates', 'commands', 'runtimes',
            'repository', 'anchors', 'revision_count', 'started_utc', 'ended_utc',
            'output', 'refusal', 'attempts', 'platform', 'budget'}
    require(set(record) == keys, 'invalid preparation seal fields')
    require(record['schema'] == 'x11a-export-preparation-seal-v7'
            and record['stage'] == 'export-preparation', 'wrong preparation seal')
    require(record['operation'] in {'ownership', 'build-frontend', 'exposures'},
            'wrong preparation operation')
    require(record['authorization'] == file_pin(Path(authorization_path))
            and record['inputs'] == file_pin(Path(inputs_path)), 'preparation authority/input mismatch')
    budget = preparation_budget(pin(record['budget']), Path(inputs_path), Path(authorization_path))
    require(budget['started_utc'] <= record['started_utc'], 'operation predates shared budget')
    from stages import verify_authorization
    authority = verify_authorization(authorization_path, inputs_path, scope='export-preparation')
    signed = {str(Path(p['path']).resolve()): p['sha256'] for p in authority['files']}
    require(record['scripts'] and any(p['path'] == str(Path(__file__).resolve()) for p in record['scripts']),
            'missing preparation implementation pin')
    core = {str(HERE / name) for name in ('prepare_exports.py', 'stages.py',
                                        'inventory.py', 'pin_tool_closure.py')}
    require(core <= {entry['path'] for entry in record['scripts']},
            'missing preparation import closure')
    for entry in record['scripts']:
        require(signed.get(str(pin(entry))) == entry['sha256'], 'unsigned preparation import')
    require(record['runtimes'] == config['export_preparation']['tools'], 'runtime pin mismatch')
    for entry in record['runtimes'].values():
        pin(entry)
    require(record['command_templates'] == TEMPLATES, 'command template mismatch')
    require(isinstance(record['argv'], list) and record['argv']
            and all(isinstance(v, str) for v in record['argv']), 'missing full argv')
    for key in ('started_utc', 'ended_utc'):
        require(datetime.fromisoformat(record[key]).utcoffset().total_seconds() == 0,
                'UTC timestamp required')
    require(record['started_utc'] <= record['ended_utc'], 'reversed preparation timestamps')
    require(set(record['platform']) == {'system', 'release', 'version', 'machine', 'node_heap_mb',
                                      'sandbox_exec_sha256', 'macos_build'}
            and record['platform']['node_heap_mb'] == NODE_HEAP_MB
            and record['platform']['sandbox_exec_sha256'] == record['runtimes']['sandbox_exec']['sha256'],
            'missing execution platform/heap binding')
    for command in record['commands']:
        if 'argv' not in command:
            require(command.get('step') in {'preflight', 'materialized', 'exposure-bindings', 'scope'}, 'unknown trace step')
            if command['step'] == 'preflight':
                require(set(command) == {'step', 'attempt', 'inputs', 'authorization'},
                        'invalid preflight trace')
            continue
        require(set(command) == {'argv', 'cwd', 'returncode', 'stdout_sha256',
                'stderr_sha256', 'started_utc', 'ended_utc', 'attempt', 'revision',
                'count', 'oid_list_sha256'}, 'invalid command trace')
        require(isinstance(command['argv'], list) and command['argv']
                and all(isinstance(v, str) for v in command['argv']), 'invalid command argv')
        absolute(command['cwd'])
        require(type(command['returncode']) is int and DIGEST.fullmatch(command['stdout_sha256'])
                and DIGEST.fullmatch(command['stderr_sha256']), 'invalid command result')
        for key in ('started_utc', 'ended_utc'):
            require(datetime.fromisoformat(command[key]).utcoffset().total_seconds() == 0,
                    'command UTC timestamp required')
        require(record['started_utc'] <= command['started_utc'] <= command['ended_utc'] <= record['ended_utc'],
                'command outside operation time window')
    require(type(record['revision_count']) is int and record['revision_count'] >= 0,
            'invalid revision count')
    if record['operation'] == 'ownership':
        items = [r for r in config['repositories'] if r['repository'] == record['repository']]
        require(len(items) == 1 and record['anchors'] == {k: items[0][k] for k in ('Cold', 'C0', 'tip')},
                'ownership anchors differ from signed inputs')
    else:
        require(record['repository'] is None and record['anchors'] is None
                and record['revision_count'] == 0, 'nonownership repository metadata')
    if record['status'] == 'completed':
        require(record['refusal'] is None and record['output'] is not None, 'invalid completion union')
        result = load(pin(record['output']))
        if record['operation'] == 'ownership':
            require(result['repository'] == record['repository'] and result['source'] == 'heuristic'
                    and len(result['revisions']) == record['revision_count'], 'ownership output mismatch')
            require({record['anchors'][key] for key in ('Cold', 'C0')} <= set(result['revisions'])
                    and all(SHA.fullmatch(revision) for revision in result['revisions']),
                    'ownership revision anchors missing')
            require(result['scanner_sha256'] == config['export_preparation']['tools']['clawpatch']['sha256']
                    and result['scanner_tree_sha256'] == config['export_preparation']['clawpatch_tree']['sha256'],
                    'ownership scanner mismatch')
            from inventory import validate_heuristic
            validate_heuristic(result, record['repository'],
                               config['export_preparation']['tools']['clawpatch']['sha256'],
                               config['export_preparation']['clawpatch_tree']['sha256'])
        elif record['operation'] == 'build-frontend':
            require(set(result) == {'frontend', 'frontend_build'}, 'invalid build result')
            expected_recipe = config['export_preparation']['frontend_build_recipe']
            recipe = load(pin(expected_recipe))
            require(set(result['frontend_build']) == {'recipe', 'closure', 'buildinfo'}
                    and result['frontend_build']['recipe'] == expected_recipe
                    and result['frontend_build']['closure'] == recipe['closure'],
                    'build recipe/closure differs from authority')
            require(result['frontend']['path'] == recipe['argv'][recipe['argv'].index('-o') + 1]
                    and result['frontend_build']['buildinfo']['path'] == recipe['required_buildinfo_path'],
                    'build output paths differ from recipe')
            for entry in (result['frontend'], *result['frontend_build'].values()):
                pin(entry)
        else:
            require(set(result) == {'exposure', 'unscoped_exposure', 'repository_exposure', 'exposure_studies'},
                    'invalid exposure result')
            files = {Path(entry['path']).name: entry
                     for entry in config['export_preparation']['exposure_extraction']['files']}
            expected = {'exposure': files['exposure-identifiers.json'],
                        'unscoped_exposure': files['unattributed-identifiers.json'],
                        'repository_exposure': files['additional-identifiers.json'],
                        'exposure_studies': config['exposure_studies']}
            require(result == expected, 'exposure outputs differ from signed extraction')
            for key in ('exposure', 'unscoped_exposure', 'repository_exposure'):
                pin(result[key])
    else:
        require(record['status'] == 'refused' and record['output'] is None, 'invalid refusal union')
        refusal = record['refusal']
        require(isinstance(refusal, dict) and set(refusal) == {'class', 'code', 'message', 'revision', 'failed_step'}
                and refusal['class'] in {'content', 'operational'}
                and all(isinstance(refusal[k], str) and refusal[k] for k in ('class', 'code', 'message')),
                'invalid refusal')
        require(refusal['class'] != 'content' or record['operation'] == 'ownership',
                'only ownership content can make a repository ineligible')
    verify_trace(record, load(pin(record['output'])) if record['status'] == 'completed' else None)
    return record


def _execute(args, budget, seal_stream):
    global REVISION_COUNT, ATTEMPT, CURRENT_REVISION, ACTIVE_DEADLINE, FAILED_STEP
    TRACE.clear()
    REVISION_COUNT = 0
    inputs, authority = absolute(args.inputs), absolute(args.authorization)
    output = absolute(args.output)
    # execute reserved the canonical seal before any commands.
    with seal_stream:
        started, config, tools, result, refusal = utc(), load(inputs), {}, None, None
        attempts = []
        for ATTEMPT in range(1, 4):
            CURRENT_REVISION = None
            REVISION_COUNT = 0
            FAILED_STEP = None
            ACTIVE_DEADLINE = min(time.time() + ATTEMPT_SECONDS,
                                  datetime.fromisoformat(budget['deadline_utc']).timestamp())
            try:
                TRACE.append({'step': 'preflight', 'attempt': ATTEMPT,
                              'inputs': file_pin(inputs), 'authorization': file_pin(authority)})
                if ACTIVE_DEADLINE <= time.time():
                    raise PreparationDeadline('persistent whole-preparation budget exhausted')
                signal.setitimer(signal.ITIMER_REAL, ACTIVE_DEADLINE - time.time())
                config, prep, tools, signed_file = authorise(inputs, authority)
                if args.operation == 'build-frontend':
                    result = build_frontend(prep, tools, signed_file)
                elif args.operation == 'ownership':
                    result = ownership(args, config, prep, tools)
                else:
                    result = exposures(config, prep, signed_file)
                for entry in script_pins():
                    signed_file(Path(entry['path']))
                with output.open('x') as stream:
                    json.dump(result, stream, indent=2, sort_keys=True)
                    stream.write('\n')
                attempts.append({'attempt': ATTEMPT, 'status': 'completed', 'code': None, 'failed_step': None})
                refusal = None
                break
            except SystemExit:
                raise
            except BaseException as error:
                content_failure = isinstance(error, ContentRefusal)
                last = TRACE[-1] if TRACE and TRACE[-1]['attempt'] == ATTEMPT else {}
                failed_step = last.get('argv', [last.get('step', 'prerequisites')])[-3:]
                attempts.append({'attempt': ATTEMPT, 'status': 'refused',
                                 'code': type(error).__name__, 'failed_step': failed_step})
                refusal = {'class': 'content' if content_failure else 'operational',
                           'code': type(error).__name__ if content_failure else
                           'PreparationInterruptedOperational' if isinstance(error, KeyboardInterrupt) else
                           'PreparationBudgetExhausted' if time.time() >= datetime.fromisoformat(budget['deadline_utc']).timestamp()
                           else 'PreparationInfeasibleOperational',
                           'revision': CURRENT_REVISION if content_failure else None,
                           'failed_step': FAILED_STEP if content_failure else failed_step,
                           'message': str(error) if content_failure else
                           'Operational deviation; no candidate exclusion. ' + type(error).__name__
                           + ('; interruption permitted only for documented host failure; report under Deviations.'
                              if isinstance(error, KeyboardInterrupt) else '')}
                if content_failure or isinstance(error, KeyboardInterrupt) or refusal['code'] == 'PreparationBudgetExhausted':
                    break
            finally:
                signal.setitimer(signal.ITIMER_REAL, 0)
                ACTIVE_DEADLINE = None
        repository = getattr(args, 'repository', None)
        item = next((r for r in config['repositories'] if r['repository'] == repository), None)
        record = {'schema': 'x11a-export-preparation-seal-v7', 'stage': 'export-preparation',
                  'operation': args.operation, 'status': 'refused' if refusal else 'completed',
                  'authorization': file_pin(authority), 'inputs': file_pin(inputs),
                  'scripts': script_pins(), 'argv': list(sys.orig_argv),
                  'command_templates': TEMPLATES, 'commands': list(TRACE),
                  'runtimes': config['export_preparation']['tools'],
                  'attempts': attempts,
                  'budget': file_pin(absolute(args.budget)),
                  'platform': {'system': platform.system(), 'release': platform.release(),
                               'version': platform.version(), 'machine': platform.machine(),
                               'macos_build': platform.mac_ver()[0],
                               'node_heap_mb': NODE_HEAP_MB,
                               'sandbox_exec_sha256': config['export_preparation']['tools']['sandbox_exec']['sha256']},
                  'repository': repository, 'anchors': {k: item[k] for k in ('Cold', 'C0', 'tip')} if item else None,
                  'revision_count': REVISION_COUNT, 'started_utc': started, 'ended_utc': utc(),
                  'output': None if refusal else file_pin(output), 'refusal': refusal}
        with absolute(args.budget).open() as budget_lock:
            fcntl.flock(budget_lock, fcntl.LOCK_EX)
            json.dump(record, seal_stream, indent=2, sort_keys=True)
            seal_stream.write('\n')
            seal_stream.flush()
    return record


def execute(args):
    inputs, authority = absolute(args.inputs), absolute(args.authorization)
    budget_path = absolute(args.budget)
    budget = preparation_budget(budget_path, inputs, authority)
    output, seal_path = canonical_paths(inputs, args.operation, getattr(args, 'repository', None))
    require(absolute(args.output) == output and absolute(args.seal) == seal_path,
            'preparation output and seal must use canonical paths')
    slots = []
    previous = signal.signal(signal.SIGALRM, deadline_alarm)
    try:
        for number in range(MAX_REPOSITORY_CONCURRENCY):
            slot = budget_path.with_name(budget_path.name + '.slot-' + str(number)).open('a')
            slots.append(slot)
            try:
                fcntl.flock(slot, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except BlockingIOError:
                continue
            with budget_path.open() as budget_lock:
                fcntl.flock(budget_lock, fcntl.LOCK_EX)
                require_no_global_stop(inputs)
                require(not output.exists(), 'output is create-once')
                seal_path.parent.mkdir(parents=True, exist_ok=True)
                seal_stream = seal_path.open('x')
            return _execute(args, budget, seal_stream)
        raise OperationalRefusal('five concurrent preparation operations already active')
    finally:
        signal.setitimer(signal.ITIMER_REAL, 0)
        signal.signal(signal.SIGALRM, previous)
        for slot in slots:
            slot.close()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--inputs', required=True)
    parser.add_argument('--authorization', required=True)
    parser.add_argument('--budget', help='Persistent shared create-once 42-day budget')
    sub = parser.add_subparsers(dest='operation', required=True)
    budget = sub.add_parser('start-budget')
    budget.add_argument('--output', required=True)
    own = sub.add_parser('ownership')
    own.add_argument('--repository', required=True)
    own.add_argument('--scratch', required=True)
    own.add_argument('--output', required=True)
    own.add_argument('--seal', required=True)
    exp = sub.add_parser('exposures')
    exp.add_argument('--output', required=True)
    exp.add_argument('--seal', required=True)
    build = sub.add_parser('build-frontend')
    build.add_argument('--output', required=True)
    build.add_argument('--seal', required=True)
    args = parser.parse_args()
    try:
        if args.operation == 'start-budget':
            start_budget(absolute(args.output), absolute(args.inputs), absolute(args.authorization))
            return
        require(args.budget is not None, 'shared --budget is required')
        record = execute(args)
    except (ValueError, KeyError, OSError, TypeError):
        parser.exit(2, 'OPERATIONAL: cannot create preparation record; repair prerequisites.\n')
    if record['status'] == 'refused':
        parser.exit(2, 'REFUSED: ' + record['refusal']['class'] + '; see sealed refusal.\n')


if __name__ == '__main__':
    main()

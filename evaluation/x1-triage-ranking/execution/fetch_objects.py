"""Fetch only X1 pinned objects; inspect commit headers, never candidate trees."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile

sys.path.insert(0, str(Path(__file__).resolve().parent))
from verify_inputs import exclusive_json, pin
from preparation import Journal, load_config, now, operation_lock, paths, read_clock

SHA = re.compile(r'^[0-9a-f]{40}$')


class Git:
    def __init__(self, executable, home):
        self.executable = str(Path(executable).resolve(strict=True))
        self.env = {'HOME': str(home), 'XDG_CONFIG_HOME': str(home), 'PATH': '/usr/bin:/bin',
                    'LANG': 'C', 'LC_ALL': 'C', 'GIT_CONFIG_NOSYSTEM': '1',
                    'GIT_CONFIG_GLOBAL': '/dev/null', 'GIT_NO_REPLACE_OBJECTS': '1',
                    'GIT_TERMINAL_PROMPT': '0', 'GIT_ATTR_NOSYSTEM': '1'}
        self.commands = []

    def run(self, db, *args, check=True, stdin=None):
        argv = [self.executable, '-c', 'core.hooksPath=/dev/null', '-c', 'protocol.file.allow=never',
                '-c', 'fetch.fsckObjects=true', '-c', 'transfer.fsckObjects=true']
        if db is not None:
            argv += ['--git-dir=' + str(db)]
        argv += list(args)
        start = now()
        result = subprocess.run(argv, env=self.env, input=stdin, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        self.commands.append({'argv': argv, 'started_at': start, 'ended_at': now(), 'exit_code': result.returncode,
                              'stdout_sha256': hashlib.sha256(result.stdout).hexdigest(),
                              'stderr_sha256': hashlib.sha256(result.stderr).hexdigest()})
        if check and result.returncode:
            raise ValueError('Git command refused: ' + repr(argv) + ': ' + result.stderr.decode(errors='replace'))
        return result

    def text(self, db, *args):
        return self.run(db, *args).stdout.decode().strip()


def inspect_store(git, db, repo):
    """Reject altered or ambient Git state, including on resumes."""
    db = Path(db)
    if db.is_symlink() or db.resolve() != db.absolute():
        raise ValueError('noncanonical database path')
    for name in ('config', 'HEAD', 'objects', 'objects/info', 'objects/pack', 'refs', 'packed-refs'):
        if (db/name).is_symlink():
            raise ValueError('symlinked Git metadata: ' + name)
    if any(path.is_symlink() for path in (db/'objects').rglob('*')):
        raise ValueError('symlinked object store')
    for name in ('shallow', 'objects/info/alternates', 'objects/info/http-alternates', 'info/grafts'):
        if (db/name).exists() or (db/name).is_symlink():
            raise ValueError('unsafe Git state: ' + name)
    if any(db.glob('objects/pack/*.promisor')):
        raise ValueError('promisor pack')
    raw = git.run(db, 'config', '--local', '--null', '--list').stdout
    allowed = {'core.repositoryformatversion': ['0'], 'core.filemode': ['true'], 'core.bare': ['true'],
               'remote.origin.url': [repo['url']], 'remote.origin.fetch': ['+refs/heads/*:refs/remotes/origin/*']}
    observed = {}
    for item in raw.split(b'\0'):
        if not item:
            continue
        key, sep, value = item.decode().partition('\n')
        if not sep:
            raise ValueError('unexpected Git config encoding')
        observed.setdefault(key, []).append(value)
    # Git init probes these filesystem properties on macOS. They do not enable
    # transport, object substitution, checkout, hooks or external commands.
    for key in ('core.ignorecase', 'core.precomposeunicode'):
        if key in observed and observed[key] in (['true'], ['false']):
            allowed[key] = observed[key]
    if observed != allowed:
        raise ValueError('unexpected local Git config/remotes: ' + repr(observed))
    refs = git.text(db, 'for-each-ref', '--format=%(refname) %(objectname)')
    permitted = {'refs/x1/pinned-tip': repo['tip']}
    if repo['c0']:
        permitted['refs/x1/pinned-c0'] = repo['c0']
    for line in refs.splitlines():
        ref, sha = line.split()
        if permitted.get(ref) != sha:
            raise ValueError('unexpected ref/object: ' + line)
    if git.text(db, 'rev-parse', '--is-shallow-repository') != 'false':
        raise ValueError('shallow repository')
    return {'no_shallow': True, 'no_promisor': True, 'no_alternates': True, 'no_replacements': True, 'local_config': observed}


def commit_header(git, db, sha):
    if not SHA.fullmatch(sha) or git.text(db, 'cat-file', '-t', sha) != 'commit':
        raise ValueError('literal commit SHA required: ' + sha)
    if git.text(db, 'rev-parse', '--verify', sha+'^{commit}') != sha:
        raise ValueError('commit identity mismatch')
    raw = git.run(db, 'cat-file', 'commit', sha).stdout
    header = raw.split(b'\n\n', 1)[0] + b'\n'
    parents = [line[7:].decode() for line in header.splitlines() if line.startswith(b'parent ')]
    if not all(SHA.fullmatch(x) for x in parents):
        raise ValueError('invalid commit parent header')
    return {'sha': sha, 'type': 'commit', 'header_sha256': hashlib.sha256(header).hexdigest(), 'parents': parents}


def first_parent_boundary(git, db, tip, c0=None, dev_edges=2000):
    current, edges = tip, 0
    tip_record = commit_header(git, db, tip)
    record = tip_record
    while (current != c0) if c0 is not None else (edges < dev_edges):
        if not record['parents']:
            if c0 is not None:
                raise ValueError('C0 is not on the first-parent path')
            return current, 'root', edges, tip_record, record
        current = record['parents'][0]
        edges += 1
        record = commit_header(git, db, current)
    return current, 'fixed' if c0 is not None else '2000-first-parent-edges', edges, tip_record, record


def fetch(config, config_path, clock_path, repo):
    p = paths(config)
    read_clock(config, config_path, clock_path)
    executable = shutil.which('git')
    if not executable:
        raise ValueError('Git executable unavailable')
    p['SCRATCH'].mkdir(parents=True, exist_ok=True)
    # Private state is retained, not silently removed after a failure.
    home = Path(tempfile.mkdtemp(prefix='fetch-'+repo['directory']+'-', dir=p['SCRATCH']))
    git = Git(executable, home)
    db = p['CAND']/repo['directory']
    name = repo['repository'].replace('/', '__')
    receipt_path = p['OUT']/'objects'/(name+'.json')
    with operation_lock(config):
        # A per-database nonblocking lock prevents concurrent writers for --repo.
        p['CAND'].mkdir(parents=True, exist_ok=True)
        import fcntl
        with (p['CAND']/(repo['directory']+'.lock')).open('a') as lock:
            fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
            if receipt_path.exists():
                raise ValueError('object receipt already exists')
            journal = Journal(config, config_path, clock_path, 'a', source_pins=[pin(config_path), pin(git.executable)], environment=git.env)
            try:
                version = git.text(None, '--version')
                reservation_path = p['OUT']/'objects'/(name+'.reservation.json')
                reservation = {'schema': 'scout.x1.object-reservation.v1', 'repository': repo, 'database': str(db.absolute()), 'clock': pin(clock_path), 'config': pin(config_path)}
                if reservation_path.exists():
                    if json.loads(reservation_path.read_bytes()) != reservation or not db.exists():
                        raise ValueError('incomplete/substituted database reservation; reviewed recovery required')
                else:
                    if db.exists() or db.is_symlink():
                        raise ValueError('unrecorded pre-existing database refused')
                    exclusive_json(reservation_path, reservation)
                if not db.exists():
                    git.run(None, 'init', '--bare', '--template=', str(db))
                    git.run(db, 'remote', 'add', 'origin', repo['url'])
                inspect_store(git, db, repo)
                # Exact literal object refspecs, no branch discovery, depth or filters.
                git.run(db, 'fetch', '--no-tags', '--no-recurse-submodules', 'origin', '+'+repo['tip']+':refs/x1/pinned-tip')
                if repo['c0']:
                    git.run(db, 'fetch', '--no-tags', '--no-recurse-submodules', 'origin', '+'+repo['c0']+':refs/x1/pinned-c0')
                state = inspect_store(git, db, repo)
                c0, derivation, edges, tip_header, c0_header = first_parent_boundary(git, db, repo['tip'], repo['c0'])
                if repo['c0'] is None:
                    fast = git.run(db, 'rev-parse', '--verify', repo['tip']+'~2000', check=False)
                    if derivation == '2000-first-parent-edges' and (fast.returncode or fast.stdout.decode().strip() != c0):
                        raise ValueError('development fast-path/header walk mismatch')
                    if derivation == 'root' and fast.returncode == 0:
                        raise ValueError('development root fallback disagrees with fast path')
                git.run(db, 'merge-base', '--is-ancestor', c0, repo['tip'])
                fsck = git.run(db, 'fsck', '--full', '--strict', '--no-reflogs', repo['tip'], c0)
                fsck_path = p['OUT']/'objects'/(name+'.fsck.json')
                exclusive_json(fsck_path, {'schema': 'scout.x1.fsck.v1', 'argv': git.commands[-1]['argv'], 'exit_code': fsck.returncode,
                                           'stdout': fsck.stdout.decode(), 'stderr': fsck.stderr.decode(), 'clock': pin(clock_path)})
                django_parent = None
                if repo['repository'] == 'django/django':
                    django_parent = git.text(db, 'rev-parse', '--verify', repo['tip']+'^1')
                    if not tip_header['parents'] or django_parent != tip_header['parents'][0]:
                        raise ValueError('Django first-parent identity mismatch')
                    django_parent = commit_header(git, db, django_parent)
                # Terminal journal binds evidence without circular receipt hashes.
                command_path = p['OUT']/'objects'/(name+'.commands.json')
                exclusive_json(command_path, {'schema': 'scout.x1.git-commands.v1', 'commands': git.commands, 'clock': pin(clock_path)})
                terminal = journal.finish('passed', outputs=[pin(fsck_path), pin(command_path)])
                receipt = dict(schema='scout.x1.objects.v1', status='passed', repository=repo['repository'], canonical_url=repo['url'],
                               database_realpath=str(db.resolve()), expected_tip=repo['tip'], actual_tip=tip_header['sha'],
                               expected_c0=repo['c0'], actual_c0=c0, c0_derivation=derivation, checked_edge_count=edges,
                               django_first_parent=django_parent, tip_header=tip_header, c0_header=c0_header,
                               first_parent_membership=True, fsck=pin(fsck_path), git_identity=dict(pin(git.executable), version=version),
                               git_commands=pin(command_path), database_reservation=pin(reservation_path), journal=terminal, clock=pin(clock_path), **state)
                exclusive_json(receipt_path, receipt)
            except BaseException as exc:
                # Preserve failed commands even if validation or fetch failed.
                failure = p['OUT']/'objects'/(name+'.failure-'+journal.id+'.json')
                exclusive_json(failure, {'schema': 'scout.x1.fetch-failure.v1', 'commands': git.commands, 'reason': str(exc), 'clock': pin(clock_path)})
                if not journal.end_path.exists():
                    journal.finish('refused', outputs=[pin(failure)], reason=str(exc), exit_code=1)
                raise
    return receipt_path


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--config', required=True, type=Path)
    parser.add_argument('--clock', required=True, type=Path)
    parser.add_argument('--repo', required=True)
    args = parser.parse_args()
    config = load_config(args.config)
    selected = [row for row in config['repositories'] if row['repository'] == args.repo]
    if len(selected) != 1:
        raise ValueError('--repo must be one canonical owner/repository from config')
    print(fetch(config, args.config, args.clock, selected[0]))


if __name__ == '__main__':
    try:
        main()
    except (OSError, ValueError, KeyError, subprocess.CalledProcessError) as exc:
        print('REFUSED: ' + str(exc), file=sys.stderr)
        sys.exit(1)

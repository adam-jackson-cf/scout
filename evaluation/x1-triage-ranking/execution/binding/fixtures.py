"""Deterministic synthetic Git fixture and filter-free object materialisation."""
import hashlib
from pathlib import Path, PurePosixPath
import subprocess

FIXTURES = Path(__file__).resolve().parent/'fixtures'
DATE = '2026-01-01T00:00:00+00:00'


def git(executable, directory, *args, data=None):
    env = {'PATH': str(Path(executable).parent), 'HOME': str(directory),
           'GIT_CONFIG_NOSYSTEM': '1', 'GIT_CONFIG_GLOBAL': '/dev/null',
           'GIT_NO_REPLACE_OBJECTS': '1', 'GIT_TERMINAL_PROMPT': '0',
           'GIT_AUTHOR_NAME': 'X1 synthetic fixture', 'GIT_AUTHOR_EMAIL': 'x1@example.invalid',
           'GIT_COMMITTER_NAME': 'X1 synthetic fixture', 'GIT_COMMITTER_EMAIL': 'x1@example.invalid',
           'GIT_AUTHOR_DATE': DATE, 'GIT_COMMITTER_DATE': DATE, 'LC_ALL': 'C'}
    return subprocess.check_output([str(executable), '-c', 'core.hooksPath=/dev/null',
                                   '-c', 'core.attributesFile=/dev/null',
                                   '-C', str(directory), *args], input=data, env=env)


def safe_path(value):
    p = PurePosixPath(value)
    if not value or str(p) != value or p.is_absolute() or any(x in ('..', '.git', '.codegraph') for x in p.parts):
        raise ValueError('unsafe or forbidden tracked path')
    return p


def generate(destination, executable):
    destination = Path(destination)
    destination.mkdir(parents=True, exist_ok=False)
    git(executable, destination, 'init', '--bare', '--template=', '--initial-branch=x1')
    commits, manifests = [], []
    for revision in ('parent', 'head'):
        files = {}
        for source in sorted((FIXTURES/'common').rglob('*')) + sorted((FIXTURES/revision).rglob('*')):
            if source.is_dir():
                continue
            base = FIXTURES/'common' if source.is_relative_to(FIXTURES/'common') else FIXTURES/revision
            name = str(source.relative_to(base)); safe_path(name)
            files[name] = source.read_bytes()
        entries = []
        for name, contents in sorted(files.items()):
            blob = git(executable, destination, 'hash-object', '-w', '--stdin', data=contents).decode().strip()
            entries.append((name, blob, contents))
        # mktree handles subdirectories explicitly; no working-tree checkout.
        def tree(prefix=''):
            children = {}
            for name, blob, contents in entries:
                if not name.startswith(prefix):
                    continue
                part, _, rest = name[len(prefix):].partition('/')
                children[part] = ('040000', 'tree', None) if rest else ('100644', 'blob', blob)
            rows = []
            for name, (mode, kind, oid) in sorted(children.items()):
                if kind == 'tree':
                    oid = tree(prefix+name+'/')
                rows.append(f'{mode} {kind} {oid}\t{name}\0'.encode())
            return git(executable, destination, 'mktree', '-z', data=b''.join(rows)).decode().strip()
        tree_id = tree()
        args = ['commit-tree', tree_id]
        if commits:
            args += ['-p', commits[-1]]
        commit = git(executable, destination, *args, data=('X1 synthetic '+revision+'\n').encode()).decode().strip()
        commits.append(commit)
        manifests.append({'revision': revision, 'commit': commit, 'tree': tree_id,
                          'root': str(destination.resolve()),
                          'files': [{'repo_path': n, 'blob': b, 'mode': '100644', 'bytes': len(c),
                                     'revision': commit, 'root': str(destination.resolve()),
                                     'content_sha256': hashlib.sha256(c).hexdigest()} for n,b,c in entries]})
    git(executable, destination, 'update-ref', 'refs/heads/x1', commits[-1])
    return {'database': str(destination.resolve()), 'parent': commits[0], 'head': commits[1],
            'generator_sha256': hashlib.sha256(Path(__file__).read_bytes()).hexdigest(), 'manifests': manifests}


def materialize(database, revision, destination, executable):
    """Only blobs, no checkout/filter/submodule/symlink execution; preserve bytes/modes."""
    destination = Path(destination)
    rows = git(executable, database, 'ls-tree', '-rz', '--full-tree', revision).split(b'\0')
    parsed = []
    for raw in rows:
        if not raw:
            continue
        metadata, name = raw.split(b'\t', 1)
        mode, kind, oid = metadata.decode().split()
        name = name.decode('utf-8', 'strict'); safe_path(name)
        if kind != 'blob' or mode not in ('100644', '100755'):
            raise ValueError('symlink/submodule/unknown mode requires scoped materialisation review: '+name)
        parsed.append((name, mode, oid))
    destination.mkdir(parents=True, exist_ok=False)
    # A private local object copy retains original commits without touching candidate refs.
    git(executable, destination, 'init', '--template=', '--initial-branch=x1')
    # Scout keys durable state to remote.origin.url, else to the root path
    # (inventory.repositoryID). Name the source database so parent and head
    # roots of one fixture share an identity; nothing ever fetches from it.
    git(executable, destination, 'config', 'remote.origin.url', str(Path(database).resolve()))
    # Pack only this revision's literal commit/tree/blob closure, not its
    # ancestry: materialising the Django pair must not collect the frame.
    all_entries = git(executable, database, 'ls-tree', '-rtz', '--full-tree', revision)
    objects = {revision, git(executable, database, 'rev-parse', revision+'^{tree}').decode().strip()}
    for entry in all_entries.split(b'\0'):
        if entry:
            objects.add(entry.split(b'\t',1)[0].decode().split()[2])
    pack = git(executable, database, 'pack-objects', '--stdout', data=('\n'.join(sorted(objects))+'\n').encode())
    git(executable, destination, 'unpack-objects', '-r', data=pack)
    git(executable, destination, 'update-ref', 'HEAD', revision)
    git(executable, destination, 'read-tree', revision)
    manifest = []
    for name, mode, oid in parsed:
        contents = git(executable, database, 'cat-file', 'blob', oid)
        target = destination/name
        target.parent.mkdir(parents=True, exist_ok=True)
        with target.open('xb') as stream:
            stream.write(contents)
        target.chmod(0o755 if mode == '100755' else 0o644)
        manifest.append({'repo_path': name, 'blob': oid, 'mode': mode, 'bytes': len(contents),
                         'revision': revision, 'root': str(destination.resolve()),
                         'content_sha256': hashlib.sha256(contents).hexdigest()})
    return manifest


def verify_materialization(root, manifest):
    for row in manifest:
        if row['root'] != str(Path(root).resolve()):
            raise ValueError('source manifest root mismatch')
        target = Path(root)/row['repo_path']
        if target.is_symlink() or not target.is_file():
            raise ValueError('tracked source replaced')
        mode = '100755' if target.stat().st_mode & 0o111 else '100644'
        if mode != row['mode'] or hashlib.sha256(target.read_bytes()).hexdigest() != row['content_sha256']:
            raise ValueError('tracked source bytes/mode changed: '+row['repo_path'])
    return True

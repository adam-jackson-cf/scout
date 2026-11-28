"""Documentary X1 signed-byte gate. No preparation actions or network access."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import platform
from datetime import datetime, timezone

MANIFEST = '2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da'
PREREG = 'abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb'
AUTHORITY = '6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc'
HISTORICAL = '/Users/adamjackson/Projects/work/scout-x1'
SHARED = {'docs/STATUS.md', 'docs/research-ledger.md', 'docs/hypothesis-ledger.md', 'docs/adrs/README.md', 'docs/scout-design.html'}
TRANSITION = '- 2026-11-08 — HD9: proposed → preregistered; X1 revision 9 signed off in X1 round 9 / ADR-014 round 7 (X1; R-064, R-081).\n'


def digest(data):
    return hashlib.sha256(data).hexdigest()


def pin(path, role=None):
    path = Path(path).absolute()
    data = path.read_bytes()
    result = {'path': str(path), 'sha256': digest(data), 'bytes': len(data)}
    if role:
        result['role'] = role
    return result


def exclusive_json(path, value):
    path = Path(path)
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open('x', encoding='utf-8') as stream:
        json.dump(value, stream, sort_keys=True, indent=2)
        stream.write('\n')
        stream.flush()
        os.fsync(stream.fileno())
    fd = os.open(path.parent, os.O_RDONLY)
    try:
        os.fsync(fd)
    finally:
        os.close(fd)


def normalized_hd9(row):
    cells = row.split(b'|')
    if len(cells) != 9:
        raise ValueError('unexpected HD9 row schema')
    if cells[4] == b' preregistered ':
        if cells[7] != b' 2026-11-08 ':
            raise ValueError('unexpected HD9 transition date')
        cells[4] = b' proposed '
        cells[7] = b' 2026-11-08 '
    elif cells[4] != b' proposed ':
        raise ValueError('unexpected HD9 state')
    return b'|'.join(cells)


def extract(data, name, selector):
    lines = data.splitlines(keepends=True)
    if name == 'ADR014Card':
        found = [block for block in re.findall(rb'  <article\b[^>]*>.*?</article>\n', data, re.S) if b'<span class="aid">ADR-014</span>' in block]
    elif name.startswith('R-') or name in ('HD9', 'ADR014Index'):
        key = 'ADR-014' if name == 'ADR014Index' else name
        found = [line for line in lines if line.startswith(('| ' + key + ' |').encode())]
    elif 'line beginning ' in selector:
        prefix = selector.split('line beginning ', 1)[1].encode()
        found = [line for line in lines if line.startswith(prefix)]
    else:
        raise ValueError('unknown extract selector: ' + selector)
    if len(found) != 1 or not found[0].endswith(b'\n'):
        raise ValueError('extract missing/ambiguous: ' + name)
    return normalized_hd9(found[0]) if name == 'HD9' else found[0]


def classify(path, data, expected, owned):
    if digest(data) == expected:
        return 'exact'
    if path not in SHARED:
        raise ValueError('unpermitted signed-byte difference: ' + path)
    applicable = [(k, v) for k, v in owned.items() if v['source'] == path]
    if not applicable:
        raise ValueError('shared file has no owned selectors')
    for name, row in applicable:
        if digest(extract(data, name, row['selector'])) != row['sha256']:
            raise ValueError('owned extract mismatch: ' + name)
    if path == 'docs/hypothesis-ledger.md':
        transitions = [x for x in data.decode().splitlines(keepends=True) if 'HD9: proposed → preregistered' in x]
        state = next(x for x in data.splitlines(keepends=True) if x.startswith(b'| HD9 |'))
        if b'| preregistered |' in state and transitions != [TRANSITION]:
            raise ValueError('HD9 transition history is missing or changed')
        # Earlier HD9 history is protected even when a sibling adds history elsewhere.
        earlier = [x for x in data.splitlines(keepends=True) if b'HD9' in x and x.startswith(b'- ') and x.decode() != TRANSITION]
        covered = [extract(data, k, v['selector']) for k, v in applicable if k.startswith('HD9Revision')]
        # Revisions 2–5 predate individual selectors: authenticate their bytes below
        # against the adoption snapshot whose reversed whole-file hash is signed.
        if not all(x in earlier for x in covered):
            raise ValueError('HD9 earlier history changed')
    return 'shared-file-exception'


def working_pin(root, row):
    actual = pin(Path(root)/row['path'])
    if actual['sha256'] != row['sha256'] or actual['bytes'] != row['bytes']:
        raise ValueError('working-file bytes mismatch: ' + row['path'])
    actual['original_path'] = HISTORICAL + '/' + row['path']
    return actual


def submodule_rows(root, submodule):
    """Defer only an uninitialised, empty gitlink directory to clocked bootstrap."""
    root = Path(root)
    directory = root/submodule['path']
    entry = subprocess.check_output(['git', '-C', str(root), 'ls-tree', 'HEAD', '--', submodule['path']], text=True).split()
    if entry[:3] != ['160000', 'commit', submodule['gitlink']]:
        raise ValueError('submodule gitlink differs')
    if not (directory/'.git').exists():
        if directory.is_symlink() or (directory.exists() and any(directory.iterdir())):
            raise ValueError('nonempty submodule directory without Git initialization')
        return [], [dict(row, classification='deferred_to_b', reason='uninitialised pinned submodule; action b must verify working bytes') for row in submodule['working_files']]
    return submodule['working_files'], []


def verify(root, manifest, review):
    root, manifest, review = Path(root), Path(manifest), Path(review)
    result = {'schema': 'scout.x1.input-verification.v1', 'status': 'refused', 'exact_matches': [], 'exceptions': [], 'failures': [], 'pins': [], 'path_mapping': {HISTORICAL: str(root)}, 'interpreter': pin(Path(sys.executable).resolve()), 'argv': sys.argv, 'cwd': os.getcwd()}
    result['verifier_source'] = pin(__file__)
    result['observed_at'] = datetime.now(timezone.utc).isoformat().replace('+00:00', 'Z')
    result['architecture'] = platform.machine()
    result['execution_root_realpath'] = str(root.resolve())
    result['git_observation'] = {key: subprocess.check_output(['git', '-C', str(root), 'rev-parse', *args], text=True).strip() for key, args in [('HEAD', ['HEAD']), ('origin_main', ['refs/remotes/origin/main']), ('common_directory', ['--git-common-dir'])]}
    def check(label, fn):
        try:
            return fn()
        except (OSError, ValueError, KeyError, subprocess.CalledProcessError) as exc:
            result['failures'].append({'input': label, 'reason': str(exc)})
    def require_hash(path, expected):
        actual = pin(path)
        result['pins'].append(actual)
        if actual['sha256'] != expected:
            raise ValueError('SHA-256 mismatch: ' + str(path))
    check('manifest', lambda: require_hash(manifest, MANIFEST))
    check('authority', lambda: require_hash(review, AUTHORITY))
    check('preregistration', lambda: require_hash(root/'evaluation/x1-triage-ranking/PREREGISTRATION.md', PREREG))
    def authority_copy():
        copy = root/'docs/adrs/2026-11-08-x1-delegated-decisions/round-7.md'
        if copy.read_bytes() != review.read_bytes():
            raise ValueError('ADR authority copy differs')
        result['pins'].append(pin(copy))
    check('authority copy', authority_copy)
    if result['failures']:
        return result
    submission = json.loads(manifest.read_bytes())
    if len(submission['files_sha256']) != 131:
        raise ValueError('expected 131 manifest rows')
    for path, expected in submission['files_sha256'].items():
        def row_check(path=path, expected=expected):
            data = (root/path).read_bytes()
            kind = classify(path, data, expected, submission['owned_extracts_sha256'])
            result['exact_matches' if kind == 'exact' else 'exceptions'].append({'path': path, 'expected_sha256': expected, 'actual_sha256': digest(data), 'classification': kind})
            result['pins'].append(pin(root/path))
        check(path, row_check)
    # Authenticate earlier HD9 history using adoption bytes after reversing only the
    # expressly authorised state/date and appended history transition.
    def history_check():
        baseline = subprocess.check_output(['git', '-C', str(root), 'show', '584da4a2b9e849f437c1fc6dfff57bbd2cf9ff48:docs/hypothesis-ledger.md'])
        baseline = baseline.replace(TRANSITION.encode(), b'')
        baseline = b''.join(normalized_hd9(x) if x.startswith(b'| HD9 |') else x for x in baseline.splitlines(keepends=True))
        if digest(baseline) != submission['files_sha256']['docs/hypothesis-ledger.md']:
            raise ValueError('cannot authenticate reviewed HD9 baseline')
        old = [x for x in baseline.splitlines() if x.startswith(b'- ') and b'HD9' in x]
        new = [x for x in (root/'docs/hypothesis-ledger.md').read_bytes().splitlines() if x.startswith(b'- ') and b'HD9' in x and x != TRANSITION.rstrip().encode()]
        if old != new:
            raise ValueError('earlier HD9 history changed')
    check('earlier HD9 history', history_check)
    pins = json.loads((root/'evaluation/x1-triage-ranking/preparation/pins.json').read_bytes())
    result['deferred_to_b'] = []
    working = check('submodule initialization', lambda: submodule_rows(root, pins['submodule']))
    working_rows, result['deferred_to_b'] = working if working is not None else ([], [])
    rows = pins['source']['files'] + working_rows
    for row in rows:
        def source_check(row=row):
            result['pins'].append(working_pin(root, row))
        check(row['path'], source_check)
    def gitlink_check():
        row = pins['submodule']
        entry = subprocess.check_output(['git', '-C', str(root), 'ls-tree', 'HEAD', '--', row['path']], text=True).split()
        if entry[:3] != ['160000', 'commit', row['gitlink']]:
            raise ValueError('submodule gitlink differs')
    check('submodule gitlink', gitlink_check)
    check('model-free policy', lambda: require_hash(root/pins['policy']['draft_source'], pins['policy']['sha256']))
    result['source_rows_checked'] = len(pins['source']['files'])
    result['working_file_rows_checked'] = len(working_rows)
    result['runtime_entrypoints_not_executed'] = True
    result['status'] = 'passed' if not result['failures'] else 'refused'
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ('root', 'manifest', 'review', 'output'):
        parser.add_argument('--' + name, required=True, type=Path)
    args = parser.parse_args()
    result = verify(args.root, args.manifest, args.review)
    exclusive_json(args.output, result)
    print(json.dumps({'status': result['status'], 'exact_matches': len(result['exact_matches']), 'exceptions': result['exceptions'], 'deferred_to_b_count': len(result.get('deferred_to_b', [])), 'failures': result['failures'], 'receipt': str(args.output)}, sort_keys=True))
    return 0 if result['status'] == 'passed' else 1


if __name__ == '__main__':
    sys.exit(main())

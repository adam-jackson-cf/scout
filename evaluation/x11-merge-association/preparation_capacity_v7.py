#!/usr/bin/env python3
"""Git-metadata-only sizes and synthetic additive preparation capacity projection."""
import argparse
from datetime import datetime, timezone
from fractions import Fraction
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import time

HERE = Path(__file__).resolve().parent


def pin(path):
    return {'path': str(path.resolve()), 'sha256': hashlib.sha256(path.read_bytes()).hexdigest()}


def fit(observations):
    """Exactly identify three coefficients; no regression uncertainty is claimed."""
    matrix = [[Fraction(row['files']), Fraction(row['bytes']), Fraction(1),
               Fraction(str(row['elapsed_seconds'])) / row['revisions']] for row in observations]
    if len(matrix) != 3:
        raise ValueError('three independent synthetic observations required')
    for column in range(3):
        pivot = next((i for i in range(column, 3) if matrix[i][column]), None)
        if pivot is None:
            raise ValueError('synthetic design is not identifiable')
        matrix[column], matrix[pivot] = matrix[pivot], matrix[column]
        divisor = matrix[column][column]
        matrix[column] = [x / divisor for x in matrix[column]]
        for i in range(3):
            if i != column:
                factor = matrix[i][column]
                matrix[i] = [x - factor*y for x, y in zip(matrix[i], matrix[column])]
    coefficients = tuple(matrix[i][3] for i in range(3))
    if any(x < 0 for x in coefficients):
        raise ValueError('negative additive coefficient; capacity unestablished, no clamping/refit')
    return coefficients


def projected(rows, coefficients, margin=Fraction(3, 2)):
    a, b, c = coefficients
    times = [margin * (a*row['files'] + b*row['bytes'] + c) for row in rows]
    return {'revision_count': len(rows), 'margin': str(margin),
            'per_attempt_seconds': float(sum(times)), 'three_attempt_seconds': float(3*sum(times)),
            'largest_revision_seconds': float(max(times, default=Fraction(0))),
            'attempt_fits': sum(times) <= 14*86400,
            'three_attempts_fit': 3*sum(times) <= 42*86400,
            'largest_revision_fits_subprocess_ceiling': max(times, default=Fraction(0)) <= 3600}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--inputs', type=Path, required=True)
    parser.add_argument('--inputs-sha256', required=True)
    args = parser.parse_args()
    if not args.inputs.is_absolute():
        parser.error('absolute inputs required')
    if pin(args.inputs)['sha256'] != args.inputs_sha256:
        raise ValueError('reviewed input drift')
    config = json.loads(args.inputs.read_text())
    git_pin = config['export_preparation']['tools']['git']
    git = Path(git_pin['path'])
    if pin(git) != git_pin:
        raise ValueError('pinned Git drift')
    observations = []
    preparation_digests = set()
    for name in ('prep-scale-construct-v7-attempt2.json', 'prep-scale-files-v7.json', 'prep-scale-bytes-v7.json'):
        path = HERE/name
        receipt = json.loads(path.read_text())
        if receipt['status'] != 'completed' or not receipt['synthetic_only']:
            raise ValueError('incomplete/non-synthetic timing receipt')
        preparation_digests.add(receipt['preparation_script']['sha256'])
        fixture = receipt['fixture']
        observations.append({'receipt': pin(path), 'files': fixture['files_per_revision'],
                             'bytes': fixture['bytes_per_revision_min'], 'revisions': fixture['revision_count'],
                             'elapsed_seconds': receipt['elapsed_seconds']})
    if preparation_digests != {pin(HERE/'prepare_exports.py')['sha256']}:
        raise ValueError('benchmarks must measure the final preparation implementation')
    coefficients = fit(observations)
    environment = {'PATH': str(git.parent), 'LANG': 'C', 'LC_ALL': 'C', 'HOME': '/var/empty',
        'GIT_CONFIG_NOSYSTEM': '1', 'GIT_CONFIG_GLOBAL': '/dev/null', 'GIT_CONFIG_COUNT': '0',
        'GIT_TERMINAL_PROMPT': '0', 'GIT_NO_REPLACE_OBJECTS': '1', 'GIT_NO_LAZY_FETCH': '1',
        'GIT_OPTIONAL_LOCKS': '0', 'GIT_EXEC_PATH': str(git.parent.parent/'libexec/git-core'),
        'GIT_TEMPLATE_DIR': str(git.parent.parent/'share/git-core/templates')}
    started = time.monotonic()
    commands = []
    def command(repo, *tail):
        argv = [str(git), '-c', 'core.hooksPath=/dev/null', '-c', 'credential.helper=', '-C', repo, *tail]
        begin = time.monotonic()
        result = subprocess.run(argv, env=environment, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                                timeout=3600, check=True)
        commands.append({'argv': argv, 'returncode': result.returncode,
                         'stdout_sha256': hashlib.sha256(result.stdout).hexdigest(),
                         'stderr_sha256': hashlib.sha256(result.stderr).hexdigest(),
                         'elapsed_seconds': time.monotonic()-begin})
        return result.stdout
    repositories = []
    for item in config['repositories']:
        repo = item['path']
        # Match production's chain-position scope; no source bytes, diff or scanner.
        chain = command(repo, 'rev-list', '--first-parent', '--reverse', item['tip']).decode('ascii').splitlines()
        chain = chain[chain.index(item['C0'])+1:][:2000]
        revisions = list(dict.fromkeys([item['Cold'], item['C0'], *chain]))
        rows = []
        for revision in revisions:
            raw = command(repo, 'ls-tree', '-r', '-l', '-z', revision)
            files = byte_count = excluded = 0
            for entry in raw.split(b'\0'):
                if not entry:
                    continue
                metadata, path = entry.split(b'\t', 1)
                if b'.codegraph' in path.split(b'/'):
                    excluded += 1
                    continue
                mode, kind, oid, size = metadata.split()
                # Count all materialised entries, including gitlinks, conservatively
                # in the file term; bytes are Git blob sizes, never content reads.
                files += 1
                if kind == b'blob':
                    byte_count += int(size)
            rows.append({'revision': revision, 'files': files, 'bytes': byte_count,
                         'excluded_codegraph_entries': excluded,
                         'ls_tree_stdout_sha256': hashlib.sha256(raw).hexdigest()})
        repositories.append({'repository': item['repository'], 'anchors': {k:item[k] for k in ('Cold','C0','tip')},
            'segment_length': len(chain), 'size_facts': rows, 'projection': projected(rows, coefficients),
            'projection_without_margin': projected(rows, coefficients, Fraction(1))})
    feasible = all(all(r['projection'][key] for key in ('attempt_fits', 'three_attempts_fit',
                    'largest_revision_fits_subprocess_ceiling')) for r in repositories)
    whole = max((r['projection']['three_attempt_seconds'] for r in repositories), default=0)
    report = {'schema': 'x11a-additive-preparation-capacity-v7',
        'scope': 'Owner-authorised Git-only size distributions; no blob contents, scans, inventory ranking, labels or associations.',
        'argv': sys.orig_argv, 'cwd': os.getcwd(), 'environment': environment,
        'inputs': pin(args.inputs), 'script': pin(Path(__file__)), 'git': git_pin,
        'observations': observations,
        'model': {'equation': 'seconds/revision = a*files + b*bytes + c',
                  'coefficients_exact': dict(zip(('a','b','c'), map(str, coefficients))),
                  'coefficients_seconds': dict(zip(('a','b','c'), map(float, coefficients))),
                  'identification': 'Three synthetic shapes, including preserved successful same-source construct. Total elapsed/revision includes setup/build/verification/frontend; this charges fixed costs again on every projected revision.',
                  'margin': '3/2; the only safety factor, not a confidence bound',
                  'limitations': 'Three-point exact fit with no residual degrees of freedom, error estimate or uncertainty interval. The 1.5x multiplier is the only safety factor. Linear scaling and fixture complexity need council review; no candidate runtime guarantee. Original max-rule failures remain evidence.'},
        'repositories': repositories, 'commands': commands,
        'budget': {'attempt_seconds': 14*86400, 'whole_preparation_seconds': 42*86400,
                   'subprocess_seconds': 3600, 'max_concurrent_repositories': 5,
                   'whole_three_attempt_seconds_at_five_concurrent': whole},
        'status': 'projected-feasible-not-authorised' if feasible else 'preparation not feasible',
        'gate': 'Every candidate attempt <=14d, three attempts <=42d and largest projected revision <=3600s, with 1.5x margin; otherwise preparation not feasible. No ceiling/refit/substitution changes.',
        'elapsed_seconds': time.monotonic()-started, 'recorded_utc': datetime.now(timezone.utc).isoformat()}
    print(json.dumps(report, sort_keys=True, indent=2))


if __name__ == '__main__':
    main()

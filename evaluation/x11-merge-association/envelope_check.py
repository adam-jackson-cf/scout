"""Read-only Git eligibility/object-count provenance; never scans source or computes accuracy."""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess
import sys


def digest(data):
    return hashlib.sha256(data).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--inputs', type=Path, required=True)
    parser.add_argument('--inputs-sha256', required=True)
    args = parser.parse_args()
    if not sys.flags.isolated or not sys.dont_write_bytecode:
        raise ValueError('use the pinned isolated -I -B interpreter')
    raw = args.inputs.read_bytes()
    if digest(raw) != args.inputs_sha256:
        raise ValueError('input digest mismatch')
    config = json.loads(raw)
    git = config['export_preparation']['tools']['git']
    if digest(Path(git['path']).read_bytes()) != git['sha256']:
        raise ValueError('pinned Git executable drift')
    root = Path(git['path']).parent.parent
    environment = {'PATH': str(root/'bin'), 'HOME': '/var/empty', 'LANG': 'C',
        'GIT_CONFIG_GLOBAL': '/dev/null', 'GIT_CONFIG_NOSYSTEM': '1',
        'GIT_NO_LAZY_FETCH': '1', 'GIT_NO_REPLACE_OBJECTS': '1',
        'GIT_OPTIONAL_LOCKS': '0', 'GIT_TERMINAL_PROMPT': '0',
        'GIT_EXEC_PATH': str(root/'libexec/git-core'),
        'GIT_TEMPLATE_DIR': str(root/'share/git-core/templates')}
    records = []
    for repository in config['repositories']:
        commands = []
        def command(*arguments):
            argv = [git['path'], '-C', repository['path'], *arguments]
            result = subprocess.run(argv, env=environment, capture_output=True,
                                    timeout=3600, check=True)
            commands.append({'argv': argv, 'exit_code': result.returncode,
                'stdout_sha256': digest(result.stdout),
                'stdout_lines': len(result.stdout.splitlines()),
                'stderr': result.stderr.decode('utf-8', errors='strict')})
            return result.stdout.decode('utf-8', errors='strict')
        chain = command('rev-list', '--first-parent', '--reverse', repository['tip']).splitlines()
        post = chain[chain.index(repository['C0']) + 1:]
        segment = post[:2000]
        segment_end = segment[-1] if segment else repository['C0']
        command('rev-list', '--objects', '--missing=error', segment_end, '--not', repository['Cold'])
        segment_object_lines = commands[-1]['stdout_lines']
        command('rev-list', '--objects', '--missing=error', repository['Cold'])
        cold_object_lines = commands[-1]['stdout_lines']
        head = command('rev-parse', 'HEAD').strip()
        records.append({'repository': repository['repository'], 'path': repository['path'],
            'anchors': {name: repository[name] for name in ('Cold', 'C0', 'tip')},
            'post_C0_envelope': len(post), 'segment_commits': len(segment),
            'segment_end': segment_end, 'eligible_length_floor_1500': len(segment) >= 1500,
            'segment_object_lines': segment_object_lines, 'Cold_object_lines': cold_object_lines,
            'observed_HEAD_unused': head, 'commands': commands})
    output = {'schema': 'x11a-readonly-envelope-v7', 'argv': sys.orig_argv,
        'inputs': {'path': str(args.inputs.resolve()), 'sha256': args.inputs_sha256},
        'git': git, 'environment': environment, 'repositories': records,
        'scope': 'Eligibility distributions and object counts observed; no source scan, diff, ownership export, inventory ranking or predictor-outcome analysis.',
        'HEAD_note': 'Fresh-clone default HEAD and the original resolved_remote_tip are separate observations; neither replaces the fixed Cold/C0/tip anchors.'}
    print(json.dumps(output, indent=2, sort_keys=True))


if __name__ == '__main__':
    main()

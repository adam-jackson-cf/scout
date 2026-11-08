#!/usr/bin/env python3
"""Authorised synthetic-only 80,000-file by 20-revision offline capacity evidence."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import platform
import resource
import shutil
import sys
import tempfile
import time

sys.path.insert(0, str(Path(__file__).resolve().parent))
import prepare_exports as prep


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--tool-root', type=Path, required=True)
    parser.add_argument('--scratch', type=Path, required=True)
    parser.add_argument('--scanner-tree', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    tools = {'git': args.tool_root/'git/bin/git', 'node': args.tool_root/'node/bin/node',
             'clawpatch': args.tool_root/'clawpatch/dist/cli.js', 'sandbox_exec': Path('/usr/bin/sandbox-exec')}
    pins = {key: prep.file_pin(value) for key, value in tools.items()}
    scanner_pin = prep.file_pin(args.scanner_tree)
    prep.verify_scanner({'clawpatch_tree': scanner_pin}, tools)
    evidence = {'schema': 'x11a-preparation-scale-v6', 'synthetic_only': True,
        'command': list(sys.orig_argv), 'tool_pins': pins, 'runner': prep.file_pin(Path(__file__)),
        'preparation_script': prep.file_pin(Path(prep.__file__)), 'files_per_revision': 80000,
        'revision_count': 20, 'node_heap_mb': prep.NODE_HEAP_MB,
        'scanner_tree': scanner_pin,
        'platform': {'system': platform.system(), 'version': platform.version(), 'machine': platform.machine()},
        'started_utc': prep.utc(), 'revisions': []}
    args.scratch.mkdir(parents=True, exist_ok=True)
    started = time.monotonic()
    with tempfile.TemporaryDirectory(prefix='x11-scale-v6-', dir=args.scratch) as directory:
        work = Path(directory)
        repo, home, binaries = work/'repo', work/'home', work/'bin'
        repo.mkdir(); home.mkdir(); binaries.mkdir()
        (binaries/'git').symlink_to(tools['git'])
        env = prep.environment(home, binaries)
        env.update(GIT_EXEC_PATH=str(args.tool_root/'git/libexec/git-core'),
                   GIT_TEMPLATE_DIR=str(args.tool_root/'git/share/git-core/templates'))
        def git(*argv, data=None):
            return prep.git(tools, env, repo, *argv, input_bytes=data)
        git('init', '-q')
        for group in range(80):
            folder = repo/('component%03d' % group); folder.mkdir()
            for item in range(1000):
                (folder/('file%04d.ts' % item)).write_text('export function item%d_%d() { return %d; }\n' % (group, item, item))
        git('add', '.')
        tree = git('write-tree').decode().strip()
        env.update(GIT_AUTHOR_NAME='Synthetic capacity', GIT_AUTHOR_EMAIL='synthetic@example.invalid',
                   GIT_COMMITTER_NAME='Synthetic capacity', GIT_COMMITTER_EMAIL='synthetic@example.invalid',
                   GIT_AUTHOR_DATE='2026-01-01T00:00:00Z', GIT_COMMITTER_DATE='2026-01-01T00:00:00Z')
        parent = None
        commits = []
        for number in range(20):
            # Distinct pinned revisions, identical bounded file content: scanner runs fresh each time.
            argv = ['commit-tree', tree]
            if parent:
                argv += ['-p', parent]
            parent = git(*argv, data=('synthetic revision %d\n' % number).encode()).decode().strip()
            commits.append(parent)
        evidence['fixture_setup_elapsed_seconds'] = time.monotonic()-started
        for number, revision in enumerate(commits):
            before = time.monotonic()
            root, state = work/'materialized', work/'state'
            root.mkdir(); state.mkdir()
            prep.CURRENT_REVISION = revision
            prep.ATTEMPT = 1
            metadata = prep.materialize(tools, env, repo, revision, root)
            materialized = time.monotonic()
            argv = [tools['node'], tools['clawpatch'], '--root', root, '--state-dir', state, '--json']
            prep.run(tools, env, [*argv, 'init'], work)
            prep.run(tools, env, [*argv, 'map', '--source', 'heuristic'], work)
            paths = prep.heuristic_paths(state)
            evidence['revisions'].append({'position': number, 'revision': revision,
                'materialization_seconds': materialized-before,
                'elapsed_seconds': time.monotonic()-before, 'owned_path_count': len(paths), **metadata})
            shutil.rmtree(root); shutil.rmtree(state)
        prep.verify_scanner({'clawpatch_tree': scanner_pin}, tools)
        evidence['status'] = 'completed'
    elapsed = time.monotonic()-started
    scan_elapsed = sum(row['elapsed_seconds'] for row in evidence['revisions'])
    evidence.update(elapsed_seconds=elapsed, revision_processing_seconds=scan_elapsed,
        peak_rss_bytes=max(resource.getrusage(resource.RUSAGE_CHILDREN).ru_maxrss,
                           resource.getrusage(resource.RUSAGE_SELF).ru_maxrss),
        rss_method='Darwin getrusage maximum child/self resident bytes; includes real sandboxed Node scanner',
        ended_utc=prep.utc(), trace_entry_count=len(prep.TRACE))
    evidence['per_repository_projection'] = {'revisions': 2002, 'assumed_files_per_revision': 80000,
        'assumption_not_candidate_measurement': True,
        'linear_processing_seconds': scan_elapsed/20*2002,
        'three_attempt_worst_case_seconds': scan_elapsed/20*2002*3,
        'limits': 'Identical small source blobs; extrapolation assumes <=80000 files and comparable per-file scanner work; not a candidate count or upper runtime guarantee.'}
    with args.output.open('x') as stream:
        json.dump(evidence, stream, indent=2, sort_keys=True); stream.write('\n')
    print(json.dumps({'output': str(args.output), 'elapsed_seconds': elapsed, 'peak_rss_bytes': evidence['peak_rss_bytes']}))


if __name__ == '__main__':
    main()

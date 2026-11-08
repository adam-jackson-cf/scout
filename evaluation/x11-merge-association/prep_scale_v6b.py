#!/usr/bin/env python3
"""Synthetic positive construct check and populated 80k x 20 offline scanner benchmark."""
import argparse
import hashlib
import json
from pathlib import Path
import platform
import resource
import shutil
import subprocess
import sys
import tempfile
import time

sys.path.insert(0, str(Path(__file__).resolve().parent))
import prepare_exports as prep


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--tool-root', type=Path, required=True)
    parser.add_argument('--scanner-tree', type=Path, required=True)
    parser.add_argument('--scratch', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--construct-only', action='store_true')
    args = parser.parse_args()
    groups, files_per_group, revision_count = (2, 3, 1) if args.construct_only else (80, 1000, 20)
    tools = {'git': args.tool_root/'git/bin/git', 'node': args.tool_root/'node/bin/node',
             'clawpatch': args.tool_root/'clawpatch/dist/cli.js', 'sandbox_exec': Path('/usr/bin/sandbox-exec')}
    scanner_pin = prep.file_pin(args.scanner_tree)
    prep.verify_scanner({'clawpatch_tree': scanner_pin}, tools)
    evidence = {'schema': 'x11a-preparation-scale-v6b', 'synthetic_only': True,
        'purpose': 'positive construct check' if args.construct_only else 'populated ownership capacity benchmark',
        'supersedes_for_capacity': str(prep.HERE/'prep-scale-v6.json'),
        'supersession_reason': 'The v6 source-only TypeScript layout produced zero real heuristic ownership and did not stress index/stream volume.',
        'command': list(sys.orig_argv), 'tool_pins': {key: prep.file_pin(value) for key, value in tools.items()},
        'runner': prep.file_pin(Path(__file__)), 'preparation_script': prep.file_pin(Path(prep.__file__)),
        'scanner_tree': scanner_pin, 'python_version': platform.python_version(),
        'platform': {'system': platform.system(), 'version': platform.version(), 'machine': platform.machine(),
            'macos_build': subprocess.check_output(['/usr/bin/sw_vers', '-buildVersion']).decode().strip()},
        'source_files_per_revision': groups*files_per_group, 'files_per_revision': groups*files_per_group+1,
        'revision_count': revision_count, 'node_heap_mb': prep.NODE_HEAP_MB,
        'fixture': {'language': 'Go', 'package_directories': groups, 'files_per_package': files_per_group,
                    'source_lines_per_file': 2, 'same_tree_each_revision': True,
                    'fresh_materialization_and_scanner_state_each_revision': True},
        'started_utc': prep.utc(), 'revisions': []}
    args.scratch.mkdir(parents=True, exist_ok=True)
    started = time.monotonic()
    with tempfile.TemporaryDirectory(prefix='x11-scale-v6b-', dir=args.scratch) as directory:
        work = Path(directory)
        repo, home, binaries, exports = work/'repo', work/'home', work/'bin', work/'exports'
        repo.mkdir(); home.mkdir(); binaries.mkdir(); exports.mkdir()
        (binaries/'git').symlink_to(tools['git'])
        env = prep.environment(home, binaries)
        def git(*argv, data=None):
            return prep.git(tools, env, repo, *argv, input_bytes=data)
        git('init', '-q')
        module = b'module example.invalid/scalesynthetic\n\ngo 1.24\n'
        (repo/'go.mod').write_bytes(module)
        source_bytes = 0
        expected_paths = set()
        for group in range(groups):
            folder = repo/'internal'/('component%03d' % group)
            folder.mkdir(parents=True)
            for item in range(files_per_group):
                data = ('package component%03d\nfunc Item%04d() int { return %d }\n' % (group,item,item)).encode()
                path = folder/('file%04d.go' % item)
                path.write_bytes(data)
                source_bytes += len(data)
                expected_paths.add(path.relative_to(repo).as_posix())
        evidence['fixture'].update(source_bytes_per_revision=source_bytes,
            total_bytes_per_revision=source_bytes+len(module), source_lines_per_revision=2*len(expected_paths))
        git('add', '.')
        tree = git('write-tree').decode().strip()
        env.update(GIT_AUTHOR_NAME='Synthetic capacity', GIT_AUTHOR_EMAIL='synthetic@example.invalid',
                   GIT_COMMITTER_NAME='Synthetic capacity', GIT_COMMITTER_EMAIL='synthetic@example.invalid',
                   GIT_AUTHOR_DATE='2026-01-01T00:00:00Z', GIT_COMMITTER_DATE='2026-01-01T00:00:00Z')
        parent = None
        commits = []
        for number in range(revision_count):
            argv = ['commit-tree', tree]
            if parent:
                argv += ['-p', parent]
            parent = git(*argv, data=('synthetic revision %d\n' % number).encode()).decode().strip()
            commits.append(parent)
        evidence['fixture_setup_elapsed_seconds'] = time.monotonic()-started
        index = {}
        stream_path = work/'ownership.ndjson'
        processing_started = time.monotonic()
        with stream_path.open('wb') as projected_stream:
            projected_stream.write((json.dumps({'repo':str(repo),'base':commits[0],'head':commits[-1]})+'\n').encode())
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
                prep.require(expected_paths <= set(paths), 'real scanner did not own every planted source path')
                feature_count = len(list((state/'features').glob('*.json')))
                prep.require(feature_count > 0, 'positive feature construct missing')
                record = {'revision':revision,'paths':paths,**metadata}
                record_path = exports/(revision+'.json')
                with record_path.open('x') as stream:
                    json.dump(record,stream,sort_keys=True); stream.write('\n')
                index[revision] = {'file':record_path.name,'sha256':prep.digest(record_path)}
                # Read the produced record, then emit the actual INV/FE NDJSON ownership schema.
                reloaded = json.loads(record_path.read_text())
                projected_stream.write((json.dumps({'revision':revision,'paths':reloaded['paths']})+'\n').encode())
                record_bytes = record_path.stat().st_size
                shutil.rmtree(root); shutil.rmtree(state)
                evidence['revisions'].append({'position':number,'revision':revision,
                    'materialization_seconds':materialized-before,'elapsed_including_cleanup_seconds':time.monotonic()-before,
                    'feature_count':feature_count,'owned_path_count':len(paths),
                    'all_planted_source_paths_owned':True,'ownership_record_bytes':record_bytes,**metadata})
        index_path = exports/'index.json'
        index_path.write_text(json.dumps(index,sort_keys=True)+'\n')
        streamed_records = streamed_paths = 0
        with stream_path.open() as stream:
            json.loads(next(stream))
            for line in stream:
                record = json.loads(line)
                prep.require(record['revision'] == commits[streamed_records], 'stream revision order changed')
                prep.require(expected_paths <= set(record['paths']), 'stream lost owned source paths')
                streamed_records += 1
                streamed_paths += len(record['paths'])
        prep.require(streamed_records == revision_count, 'stream record count changed')
        evidence.update(revision_processing_seconds=time.monotonic()-processing_started,
            ownership_index_sha256=prep.digest(index_path),ownership_index_bytes=index_path.stat().st_size,
            ownership_records_total_bytes=sum(row['ownership_record_bytes'] for row in evidence['revisions']),
            streamed_revision_records=streamed_records,streamed_owned_paths=streamed_paths,
            stream_bytes=stream_path.stat().st_size,stream_sha256=prep.digest(stream_path),
            stream_measurement='Actual indexed ownership records, NDJSON emission and one-record-at-a-time readback; not a frontend binary execution.')
        prep.verify_scanner({'clawpatch_tree':scanner_pin},tools)
        evidence['status'] = 'completed'
    evidence.update(elapsed_seconds=time.monotonic()-started,
        peak_rss_bytes=max(resource.getrusage(resource.RUSAGE_CHILDREN).ru_maxrss,resource.getrusage(resource.RUSAGE_SELF).ru_maxrss),
        rss_method='Darwin getrusage maximum child/self resident bytes, not concurrent aggregate memory',
        ended_utc=prep.utc(),trace_entry_count=len(prep.TRACE))
    if not args.construct_only:
        multiplier=(100000/evidence['files_per_revision'])*(2002/revision_count)
        evidence['per_repository_projection']=[{'repository':repository,'revisions':2002,
            'assumed_files_per_revision':100000,'not_candidate_measurement':True,
            'elapsed_multiplier':multiplier,'conservative_elapsed_seconds':evidence['elapsed_seconds']*multiplier,
            'three_attempt_seconds':evidence['elapsed_seconds']*multiplier*3,
            'assumption':'Linear files times revisions at comparable per-file cost; total elapsed includes setup and cleanup. Not a real-repository runtime guarantee.'}
            for repository in ('django/django','microsoft/TypeScript','gin-gonic/gin','spf13/cobra','prometheus/prometheus')]
    with args.output.open('x') as stream:
        json.dump(evidence,stream,indent=2,sort_keys=True); stream.write('\n')
    print(json.dumps({'output':str(args.output),'status':evidence['status'],'elapsed_seconds':evidence['elapsed_seconds'],
        'peak_rss_bytes':evidence['peak_rss_bytes'],'streamed_owned_paths':evidence['streamed_owned_paths']}))


if __name__ == '__main__':
    main()

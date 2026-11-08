#!/usr/bin/env python3
"""Extract identifiers only from pinned Scout evaluation git objects.

Owner authorised this extraction on 2026-11-08, not inventory/Clawpatch scans.
Never reports labels, scores, outcome values, model text or associations.
"""
import argparse
import csv
import gzip
import hashlib
import io
import json
import os
from pathlib import Path
import re
import subprocess
import sys

GIT = '/opt/homebrew/Cellar/git/2.56.0_2/bin/git'
GIT_SHA256 = '371fc6bfbfea4d9ace7d122feefe100503d24f827bc94f54ab3973583ffea11f'
PYTHON = '/Users/adamjackson/.local/share/uv/python/cpython-3.14.10-macos-aarch64-none/bin/python3.14'
PYTHON_SHA256 = '26b5194d2beb09e648b4676ddc18a1384f182a7ed84f3db73106096aba6febf4'
GIT_ENV = dict(os.environ, GIT_NO_LAZY_FETCH='1', GIT_TERMINAL_PROMPT='0',
               GIT_CONFIG_NOSYSTEM='1', GIT_CONFIG_GLOBAL='/dev/null')
GIT_ARGV = []
FULL = re.compile(r'[0-9a-fA-F]{40}')
SLUG = re.compile(r'(?:https?://github[.]com/)?([A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+?)(?:[.]git)?/?$')
REPO_KEYS = {'repo', 'repository', 'slug', 'repo_slug', 'repository_url', 'repo_url', 'repositoryurl', 'full_name', 'namewithowner'}
COMMIT_KEYS = {'commit', 'sha', 'merge_commit', 'merge_commit_sha', 'mergecommit', 'commit_sha', 'mergesha', 'merge_sha', 'revision', 'rev', 'tip', 'c0', 'cold', 't0', 't_end', 'head', 'base', 'parent', 'base_sha', 'head_sha', 'old_sha', 'new_sha', 'from_commit', 'to_commit'}
BLOB_KEYS = {'blob', 'blob_id', 'blob_sha', 'old_blob', 'new_blob', 'oid'}
PATH_KEYS = {'path', 'paths', 'file', 'files', 'filename', 'filenames', 'previous_filename', 'file_path', 'filepath', 'old_path', 'new_path', 'oldpath', 'newpath', 'members', 'sample_files', 'owned_files', 'ownedfiles', 'source_path'}
PR_KEYS = {'pr', 'pr_id', 'pr_number', 'pull_request', 'pull_request_number', 'merge_id'}
DATA_SUFFIXES = {'.json', '.jsonl', '.ndjson', '.csv', '.tsv', '.txt', '.log', '.stderr', '.sha256', '.runs'}
SOURCE_SUFFIXES = {'.py', '.go', '.js', '.ts', '.sh', '.md', '.diff', '.patch', '.gitkeep'}
ALIASES = {'fastapi':'fastapi/fastapi', 'axum':'tokio-rs/axum', 'django':'django/django', 'prometheus':'prometheus/prometheus', 'typescript':'microsoft/TypeScript', 'gin':'gin-gonic/gin', 'cobra':'spf13/cobra'}


def sha_bytes(data):
    return hashlib.sha256(data).hexdigest()


def git(repo, *args):
    argv = [GIT, '-C', str(repo), *args]
    GIT_ARGV.append(argv)
    return subprocess.check_output(argv, env=GIT_ENV, stderr=subprocess.PIPE)


def tree(repo, commit, prefix=None):
    args = ['ls-tree', '-r', '-z', commit]
    if prefix:
        args += ['--', prefix]
    rows = []
    for record in git(repo, *args).split(b'\0'):
        if not record:
            continue
        header, path = record.split(b'\t', 1)
        mode, kind, oid = header.decode().split()
        name = path.decode('utf-8', errors='strict')
        if '.codegraph' in name.split('/'):
            continue
        rows.append((name, kind, oid))
    return rows


class Batch:
    def __init__(self, repo):
        argv = [GIT, '-C', str(repo), 'cat-file', '--batch']
        GIT_ARGV.append(argv)
        self.process = subprocess.Popen(argv, stdin=subprocess.PIPE, stdout=subprocess.PIPE, env=GIT_ENV)
    def read(self, oid):
        self.process.stdin.write((oid+'\n').encode()); self.process.stdin.flush()
        header = self.process.stdout.readline().decode().split()
        if len(header) != 3 or header[1] != 'blob':
            raise ValueError('git object unavailable/not blob')
        data = self.process.stdout.read(int(header[2]))
        if self.process.stdout.read(1) != b'\n':
            raise ValueError('git batch framing failure')
        return data
    def close(self):
        self.process.stdin.close(); self.process.stdout.close()
        self.process.wait()


def repository(value):
    if isinstance(value, dict):
        full = value.get('full_name') or value.get('nameWithOwner')
        if full:
            return repository(full)
        owner = value.get('owner')
        if isinstance(owner, dict):
            owner = owner.get('login')
        if isinstance(owner, str) and isinstance(value.get('name'), str):
            return repository(owner+'/'+value['name'])
        return None
    if not isinstance(value, str):
        return None
    if value.lower() in ALIASES:
        return ALIASES[value.lower()]
    match = SLUG.fullmatch(value)
    return match.group(1).removesuffix('.git') if match else None


def path_identifier(value):
    return isinstance(value, str) and value and '\n' not in value and '\0' not in value and not value.startswith('/') and not any(p in ('.', '..', '.codegraph') for p in value.split('/'))


def filename_repository(path):
    parts = re.split(r'[/._-]', path.lower())
    matches = {v for k, v in ALIASES.items() if k in parts}
    if len(matches) == 1:
        return next(iter(matches))
    return None


def extract(value, inherited, emit):
    if isinstance(value, dict):
        own = {repository(v) for k, v in value.items() if k.lower() in REPO_KEYS}
        if isinstance(value.get('base'), dict):
            own.add(repository(value['base'].get('repo')))
        own.discard(None)
        context = next(iter(own)) if len(own) == 1 else inherited
        ids = []
        for key, item in value.items():
            lower = key.lower()
            if lower in REPO_KEYS:
                resolved = repository(item)
                if resolved:
                    emit(resolved, 'repository', resolved)
            if lower in COMMIT_KEYS and isinstance(item, str) and FULL.fullmatch(item):
                ids.append(('commit', item.lower()))
            if lower in BLOB_KEYS and isinstance(item, str) and FULL.fullmatch(item):
                ids.append(('blob', item.lower()))
            if lower in PATH_KEYS:
                entries = item if isinstance(item, list) else [item]
                for p in entries:
                    if path_identifier(p):
                        ids.append(('path', p))
                    elif isinstance(p, dict):
                        # ownedFiles/member objects: only identifier fields recurse.
                        extract({k:v for k,v in p.items() if k.lower() in PATH_KEYS | COMMIT_KEYS | BLOB_KEYS | REPO_KEYS}, context, emit)
            pr_context = any(k in value for k in ('merge_commit_sha', 'pull_request', 'merged_at')) or '/pull/' in str(value.get('html_url', ''))
            is_pr_identifier = pr_context and lower in ('number', 'id', 'node_id')
            if (lower in PR_KEYS or is_pr_identifier) and isinstance(item, (str, int)) and not isinstance(item, bool):
                if re.fullmatch(r'#?[A-Za-z0-9_:-]+', str(item)):
                    ids.append(('pr', str(item).lstrip('#')))
        for kind, identifier in ids:
            emit(context, kind, identifier)
        for key, item in value.items():
            if key.lower() in PATH_KEYS:
                continue
            if isinstance(item, (dict, list)):
                # A known repository alias supplies context; arbitrary paths do not.
                extract(item, ALIASES.get(key.lower()) or context, emit)
    elif isinstance(value, list):
        for item in value:
            extract(item, inherited, emit)


def parse_data(path, raw, context, emit):
    if path.endswith('.gz'):
        raw = gzip.decompress(raw)
        path = path[:-3]
    suffix = Path(path).suffix.lower()
    text = raw.decode('utf-8', errors='strict')
    if suffix == '.json':
        extract(json.loads(text), context, emit)
    elif suffix in ('.jsonl', '.ndjson'):
        for line in text.splitlines():
            if line.strip():
                extract(json.loads(line), context, emit)
    elif suffix in ('.csv', '.tsv'):
        for row in csv.DictReader(io.StringIO(text), delimiter='\t' if suffix == '.tsv' else ','):
            extract(row, context, emit)
    else:
        # Text/log/stdout content may contain model prose: extract only explicit
        # identifier-labelled lines, never arbitrary SHA-looking prose.
        for line in text.splitlines():
            m = re.fullmatch(r'\s*([A-Za-z_][A-Za-z_0-9]*)\s*[:=]\s*(\S+)\s*', line)
            if m:
                extract({m[1]: m[2]}, context, emit)


def write_once(path, data):
    with path.open('x') as stream:
        json.dump(data, stream, sort_keys=True, indent=2)
        stream.write('\n')


def data_file(path):
    suffix = Path(path[:-3] if path.endswith('.gz') else path).suffix.lower()
    return (suffix in DATA_SUFFIXES or path.endswith('.gz') or Path(path).name == 'SHA256SUMS'
            or any(part in {'data', 'raw', 'cache'} for part in path.split('/')[:-1]))


def run(args):
    scout, inertia, out = Path(args.scout), Path(args.inertia), Path(args.output)
    runtime = {}
    for name, executable, expected in (('git', GIT, GIT_SHA256), ('python', PYTHON, PYTHON_SHA256)):
        digest = sha_bytes(Path(executable).read_bytes())
        if digest != expected:
            raise ValueError(name+' binary digest mismatch')
        runtime[name] = {'path': executable, 'sha256': digest}
    if Path(sys.executable).resolve() != Path(PYTHON):
        raise ValueError('run with the exact pinned Python binary')
    runtime['python']['version'] = sys.version
    runtime['git']['version'] = subprocess.check_output([GIT, '--version'], env=GIT_ENV).decode().strip()
    GIT_ARGV.append([GIT, '--version'])
    mapping_path = Path(args.contexts)
    mappings = json.loads(mapping_path.read_text())
    if mappings['scout_commit'] != args.scout_commit:
        raise ValueError('context source commit mismatch')
    for repo, commit in ((scout,args.scout_commit),(inertia,args.inertia_commit)):
        if not FULL.fullmatch(commit) or git(repo, 'rev-parse', commit).decode().strip() != commit:
            raise ValueError('full pinned source SHA required')
    for mapping in mappings['mappings']:
        evidence = mapping['evidence']
        raw = git(scout, 'show', args.scout_commit+':'+evidence['path'])
        if sha_bytes(raw) != evidence['sha256']:
            raise ValueError('context evidence digest mismatch')
    if out.exists():
        raise ValueError('output directory create-once')
    out.mkdir(parents=True)
    listing = tree(scout, args.scout_commit, 'evaluation')
    ilisting = tree(inertia, args.inertia_commit)
    studies = sorted({p.split('/')[1] for p,_,_ in listing if len(p.split('/')) > 2})
    if any(len(p.split('/')) == 2 for p,_,_ in listing):
        studies.append('__evaluation_root__')
    inertia_studies = sorted({'inertia-bench/'+str(Path(p).parent) for p,_,_ in ilisting if data_file(p)})
    studies += inertia_studies
    exposure, unresolved, coverage = set(), set(), []
    for source_name, repo_path, source_commit, entries in (
            ('scout', scout, args.scout_commit, listing),
            ('inertia-bench', inertia, args.inertia_commit, ilisting)):
        batch = Batch(repo_path)
        try:
            for path, kind, oid in entries:
                study = (path.split('/')[1] if len(path.split('/')) > 2 else '__evaluation_root__') if source_name == 'scout' else 'inertia-bench/'+str(Path(path).parent)
                suffix = Path(path[:-3] if path.endswith('.gz') else path).suffix.lower()
                row = {'source_repository': source_name, 'source_commit': source_commit, 'path': path, 'blob': oid, 'study': study}
                if kind != 'blob':
                    row['status'] = 'unsupported_git_object'; coverage.append(row); continue
                if not data_file(path):
                    row['status'] = 'non_data_source_or_document' if suffix in SOURCE_SUFFIXES or Path(path).name == '.gitkeep' else 'unclassified_non_data'
                    coverage.append(row); continue
                matching = [m for m in mappings['mappings'] if source_name == 'scout' and path.startswith(m['prefix'])]
                mapped = max(matching, key=lambda m: len(m['prefix'])) if matching else None
                context = mapped['repository'] if mapped else filename_repository(path)
                row['inherited_repository'] = context
                row['context_mapping_prefix'] = mapped['prefix'] if mapped else None
                task = 'association' if source_name == 'scout' and study.startswith('x2d') else 'other'
                records_before, unresolved_before = len(exposure), len(unresolved)
                perfile = set()
                def emit(repo, key, identifier):
                    perfile.add((repo, key, identifier))
                    if key == 'repository':
                        exposure.add(json.dumps({'repository':repo,'study':study,'task':task,'repository_identity':identifier},sort_keys=True))
                    elif repo:
                        exposure.add(json.dumps({'repository':repo,'study':study,'task':task,key:identifier},sort_keys=True))
                    else:
                        unresolved.add(json.dumps({'source':path,'study':study,'task':task,'kind':key,'identifier':identifier,
                            'matchability':'repository_independent_sha' if key in {'commit','blob'} else 'unmatchable_without_repository'},sort_keys=True))
                try:
                    raw = batch.read(oid)
                    row['sha256'], row['bytes'] = sha_bytes(raw), len(raw)
                    parse_data(path, raw, context, emit)
                    row['status'] = 'parsed_identifier_projection'
                except (ValueError, UnicodeError, OSError, csv.Error, EOFError) as error:
                    row['status'] = 'could_not_parse'
                    row['error_class'] = type(error).__name__
                row['unique_identifiers_in_file'] = len(perfile)
                row['new_unique_identifiers'] = len(exposure)-records_before
                row['new_unattributed_identifiers'] = len(unresolved)-unresolved_before
                coverage.append(row)
        finally:
            batch.close()
    inertia_data = [r for r in coverage if r['source_repository'] == 'inertia-bench' and data_file(r['path'])]
    inertia_evidence = {'repository':str(inertia),'commit':args.inertia_commit,
        'recursive_listing':[p for p,_,_ in ilisting],
        'enumeration_result':{'data_files':len(inertia_data),'study_groups':inertia_studies,
            'screening_json_jsonl_csv_files':sum(r['path'].startswith('docs/design/screening/') and Path(r['path']).suffix in {'.json','.jsonl','.csv'} for r in inertia_data)},
        'scope':'All pinned paths enumerated with the same data-file predicate as Scout; identifiers only, task other.',
        'data_files':inertia_data}
    known = [json.loads(r) for r in sorted(exposure)]
    compatible = [{k:v for k,v in r.items() if k in {'repository','study','task','commit','path','blob'}} for r in known if {'commit','path','blob'} & set(r)]
    compatible = [json.loads(r) for r in sorted({json.dumps(r,sort_keys=True) for r in compatible})]
    write_once(out/'exposure-identifiers.json', {'studies':studies,'records':compatible})
    write_once(out/'additional-identifiers.json', {'records':[r for r in known if 'pr' in r or 'repository_identity' in r]})
    write_once(out/'unattributed-identifiers.json', {'records':[json.loads(r) for r in sorted(unresolved)]})
    counts = {}
    for row in coverage:
        counts[row['status']] = counts.get(row['status'],0)+1
    write_once(out/'coverage.json', {'schema':'x11a-identifier-extraction-coverage-v2','scout_commit':args.scout_commit,
        'inertia_commit':args.inertia_commit,'studies':studies,'counts':counts,'files':coverage,
        'limitations':['Only allowlisted identifier fields are projected; scalar prose is not searched.',
            'Unscoped commit/blob identifiers are matchable across repositories; unscoped paths and PRs are explicitly unmatchable.',
            'Unparseable and unclassified files are explicit, not positive completeness evidence.'],
        'candidate_repository_data_read':False,'label_or_outcome_values_emitted':False})
    write_once(out/'inertia-enumeration.json', inertia_evidence)
    source = Path(__file__).resolve()
    outputs = ['exposure-identifiers.json','additional-identifiers.json','unattributed-identifiers.json','coverage.json','inertia-enumeration.json']
    summary = {'studies':len(studies),'coverage_status_counts':counts,'compatible_records':len(compatible),
        'unattributed_records':len(unresolved),'fastapi_association_records':sum(r['repository']=='fastapi/fastapi' and r['task']=='association' for r in compatible),
        'inertia_data_files':len(inertia_data),'inertia_screening_identifiers':sum(r['unique_identifiers_in_file'] for r in inertia_data if r['path'].startswith('docs/design/screening/')),
        'inertia_scoped_records':sum(r['study'].startswith('inertia-bench/') for r in compatible),
        'unscoped_by_kind':{kind:sum(json.loads(r)['kind']==kind for r in unresolved) for kind in ('commit','blob','path','pr')}}
    write_once(out/'seal.json', {'schema':'x11a-exposure-extraction-seal-v2','authority':'Owner direction, 2026-11-08; create-once extraction-v4 only, not inventory/scan/council sign-off',
        'scout_commit':args.scout_commit,'inertia_commit':args.inertia_commit,'recipe':{'path':str(source),'sha256':sha_bytes(source.read_bytes())},
        'contexts':{'path':str(mapping_path),'sha256':sha_bytes(mapping_path.read_bytes())},
        'runtime_identity':runtime,'argv':[sys.executable,*sys.argv], 'git_argv':GIT_ARGV,
        'git_environment':{k:GIT_ENV[k] for k in ('GIT_NO_LAZY_FETCH','GIT_TERMINAL_PROMPT','GIT_CONFIG_NOSYSTEM','GIT_CONFIG_GLOBAL')},
        'summary':summary,'files':[{'path':str(out/name),'sha256':sha_bytes((out/name).read_bytes())} for name in outputs]})
    print(json.dumps(dict(summary,output=str(out))))


def main():
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--scout',required=True);p.add_argument('--scout-commit',required=True)
    p.add_argument('--inertia',required=True);p.add_argument('--inertia-commit',required=True)
    p.add_argument('--output',required=True)
    p.add_argument('--contexts',required=True)
    run(p.parse_args())


if __name__ == '__main__':
    main()

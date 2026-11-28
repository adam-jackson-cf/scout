"""X1 preparation-only, positively controlled deterministic model-free binding."""
import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys

HERE = Path(__file__).resolve().parent
EXEC = HERE.parent
sys.path.insert(0, str(EXEC))
from preparation import (Journal, load_config, operation_lock, paths, review_receipt, validate_pin,
                         verify_tree_pins, verify_amendment_reviews, read_clock)
from verify_inputs import pin
sys.path.insert(0, str(HERE))
from fixtures import generate, git, materialize, verify_materialization
from interlock import InterlockFailure, compare, evaluate, make_stubs

ACCEPTANCE = ('policy_bound_both_revisions', 'deterministic_model_free_proof', 'zero_model_attempts',
              'exact_nonempty_issue_id_parity', 'positive_control_expected_failure', 'scanner_health_complete')
# Equal-length roots preserve Spec.Batches' byte-budget decisions.
ROUTE_DIRS = {'bound':'boundroute', 'reference':'referencer', 'synthetic-unbound':'positiveXX'}


def private_env(run, path, stubdir, positive=False):
    env = {'PATH': str(stubdir)+':'+path, 'LC_ALL': 'C', 'GIT_CONFIG_NOSYSTEM': '1',
           'GIT_CONFIG_GLOBAL': '/dev/null', 'GIT_NO_REPLACE_OBJECTS': '1', 'GIT_TERMINAL_PROMPT': '0',
           'GOTOOLCHAIN': 'local', 'GOPROXY': 'off', 'GOSUMDB': 'off',
           'PYTHONDONTWRITEBYTECODE': '1',
           'SEMGREP_ENABLE_VERSION_CHECK': '0', 'SEMGREP_SEND_METRICS': 'off',
           'CLAWPATCH_CLAUDE_BIN': str(stubdir/'claude'), 'CLAWPATCH_CODEX_SANDBOX': 'read-only'}
    for key, name in [('HOME','home'),('XDG_CONFIG_HOME','config'),('XDG_CACHE_HOME','cache'),
                      ('XDG_STATE_HOME','state'),('TMPDIR','tmp')]:
        directory = run/name; directory.mkdir(mode=0o700)
        env[key] = str(directory)
    if positive:
        env['X1_SYNTHETIC_CONTROL'] = '1'
    return env


def private_path(runtime):
    return ':'.join(dict.fromkeys(str(Path(x['path']).parent) for x in runtime['executables']))


def validate_runtime(runtime, closure):
    required = {'schema', 'git', 'python', 'frontend', 'clawpatch', 'executables'}
    if set(runtime) != required or runtime['schema'] != 'scout.x1.binding-runtime.v2':
        raise ValueError('unknown binding runtime schema')
    originals = {(x['path'],x['sha256']) for x in closure['original_runtime_files']}
    permitted = {(x['path'],x['sha256']) for x in closure['executable_resolutions'] if x['execution_allowed']}
    for key in ('git','python','clawpatch'):
        validate_pin(runtime[key])
        if (runtime[key]['path'],runtime[key]['sha256']) not in originals:
            raise ValueError('runtime file absent from installed closure: '+key)
    if (runtime['git']['path'],runtime['git']['sha256']) != (closure['git']['path'],closure['git']['sha256']):
        raise ValueError('runtime git differs from closure git')
    support = {'git': closure['git']}
    for name, rows, expected in (
            ('go', closure['original_runtime_files'], str(Path(closure['go']['goroot'])/'bin'/'go')),
            ('sandbox-exec', closure['os_executables'], '/usr/bin/sandbox-exec')):
        selected = [row for row in rows if row['path'] == expected]
        if len(selected) != 1 or (selected[0]['path'],selected[0]['sha256']) not in originals:
            raise ValueError('unique closure-attested support pin required: '+name)
        support[name] = selected[0]
    identities = [(row['path'],row['sha256']) for row in runtime['executables']]
    for name, record in support.items():
        if identities.count((record['path'],record['sha256'])) != 1:
            raise ValueError('runtime support pin must appear exactly once: '+name)
    support_pins = {(row['path'],row['sha256']) for row in support.values()}
    validate_pin(runtime['frontend'])
    for record in runtime['executables']:
        validate_pin(record)
        if (record['path'],record['sha256']) not in permitted | support_pins and record != runtime['frontend']:
            raise ValueError('executable outside attested closure')
        if Path(record['path']).name in ('claude','codex'):
            raise ValueError('real model executable in runtime tool paths')
    node = [row for row in closure['executable_resolutions']
            if row.get('name') == 'node' and row['execution_allowed']]
    if len(node) != 1:
        raise ValueError('unique closure-attested node required')
    path = private_path(runtime)
    for name, record in (*support.items(), ('node', node[0])):
        resolved = shutil.which(name, path=path)
        if resolved is None or str(Path(resolved).resolve()) != record['path']:
            raise ValueError('private PATH resolution differs from attested pin: '+name)
        validate_pin(record)  # The resolved path is exactly this attested path.
    for name in ('claude','codex'):
        if shutil.which(name, path=path) is not None:
            raise ValueError('real model executable on private PATH: '+name)


def frontend_argv(runtime, root, revision, route, frozen, effective, output=None, resolve_only=False):
    argv = [runtime['frontend']['path'],'--root',str(root),'--revision',revision,'--route',route,
            '--policy',str(frozen),'--clawpatch',runtime['clawpatch']['path'],'--effective-output',str(effective)]
    if resolve_only:
        argv.append('--resolve-only')
    else:
        argv += ['--output',str(output)]
    return argv


def verify_effective_record(actual, expected):
    validate_pin(expected)
    if Path(actual).read_bytes() != Path(expected['path']).read_bytes():
        raise InterlockFailure('per-run effective policy differs from pre-scanner resolution')
    return True


def run_revision(root, revision, route, run, runtime, frozen, closure_path,
                 resolve_only=False, expected_policy=None):
    if 'OPENROUTER_API_KEY' in os.environ:
        raise InterlockFailure('key present at launch')
    from attest_closure import verify_installed_closure
    verify_installed_closure(closure_path)
    run.mkdir(parents=True,mode=0o700,exist_ok=False)
    stubdir=run/'stubs';attempt_file=make_stubs(stubdir,runtime['python']['path'])
    path=private_path(runtime)
    env=private_env(run,path,stubdir,route=='synthetic-unbound')
    effective_path,result_path=run/'effective.json',run/'result.json'
    argv=frontend_argv(runtime,root,revision,route,frozen,effective_path,result_path,resolve_only)
    # No outer sandbox: registered scanners apply their own existing sandbox.
    with (run/'stdout').open('xb') as stdout,(run/'stderr').open('xb') as stderr:
        code=subprocess.run(argv,cwd=root,env=env,stdin=subprocess.DEVNULL,
                            stdout=stdout,stderr=stderr,check=False).returncode
    verify_installed_closure(closure_path)
    attempts=[json.loads(x) for x in attempt_file.read_text().splitlines()]
    pre=json.loads(effective_path.read_bytes())
    provider=pre['original_policy']['model_cli']['provider'] if route=='synthetic-unbound' else None
    verdict=evaluate(attempts,positive=route=='synthetic-unbound' and not resolve_only,
                     expected_provider=provider)
    if (route!='synthetic-unbound' or resolve_only) and code:
        raise InterlockFailure('frontend/scanner exited nonzero')
    if expected_policy is not None:
        verify_effective_record(effective_path,expected_policy)
    return {'result':None if resolve_only else json.loads(result_path.read_bytes()),'policy':pre,
            'effective_record':pin(effective_path),'interlock':verdict,
            'argv':argv,'key_present':False,
            'pins':[pin(x) for x in sorted(run.rglob('*')) if x.is_file()]}


def carry_state(parent, head):
    target_root=head/'.scout';target_root.mkdir(mode=0o700,exist_ok=True)
    for source in sorted((parent/'.scout').iterdir()):
        if source.name=='policy.yaml':
            continue
        target=target_root/source.name
        if target.exists():
            raise ValueError('head has conflicting committed Scout state')
        if source.is_symlink():
            raise ValueError('state symlink refused')
        if source.is_dir():
            shutil.copytree(source,target,symlinks=True)
        else:
            shutil.copy2(source,target)


def verify_bound_policy(pre):
    if pre['key_present'] or not pre['association_nil'] or not pre['measures_nil'] or pre['clawpatch_egress']:
        raise InterlockFailure('effective API inputs violate scanner-only binding')
    policy=pre['effective_policy']
    if policy['profile']!='report' or policy['decisions']['enabled'] or policy['model_cli']['lanes']!=[] or policy['mapping']['association']['enabled'] or pre['model_lanes']!=[] or pre['contextual_egress']:
        raise InterlockFailure('effective frozen policy not bound at revision')
    return True

def verify_model_proof(record, route=None):
    fields={'effective_config_bound','model_lanes_disabled','jev_disabled',
            'association_disabled','pr_metadata_absent','heuristic_only'}
    proof=record.get('model_proof',{})
    if set(proof)!=fields or any(proof[key] is not True for key in fields) or record.get('model_call_impossible') is not True:
        raise InterlockFailure('deterministic model-call-impossible proof incomplete')
    if route is not None:
        if record.get('mapping',{}).get('Source')!='heuristic':
            raise InterlockFailure('actual mapping result is not heuristic-only')
        if record.get('jev_result')!='absent':
            raise InterlockFailure('Jev stage not absent')
        if route=='reference':
            if (record.get('model_stages')!='structurally_absent' or record.get('lane_report') is not None
                    or record.get('association_result')!='structurally_absent'):
                raise InterlockFailure('reference model stages not structurally absent')
        elif route=='bound':
            if record.get('model_stages')!='disabled' or record.get('association_result')!='disabled':
                raise InterlockFailure('bound model stages not disabled')
            lanes=record.get('lane_report')
            required={'contextual','clawpatch-review','association','association-fallback'}
            if not isinstance(lanes,list):
                raise InterlockFailure('actual bound lane report missing')
            if any(not isinstance(row,dict) for row in lanes):
                raise InterlockFailure('actual bound lane row malformed')
            if any(row.get('lane') not in required and row.get('requires') in ('model-cli','decision-egress')
                   for row in lanes):
                raise InterlockFailure('unexpected model-capable lane in actual bound report')
            selected=[row for row in lanes if row.get('lane') in required]
            if len(selected)!=len(required) or {row['lane'] for row in selected}!=required:
                raise InterlockFailure('actual bound model lane report incomplete')
            for row in selected:
                reasons={'disabled'} if row['lane'] in ('association','association-fallback') else {'skipped (not opted in)'}
                if row['lane'] in ('contextual','clawpatch-review') and row.get('units')==0:
                    reasons.add('no eligible inputs')
                if row.get('status')!='skipped' or row.get('reason') not in reasons:
                    raise InterlockFailure('actual bound model lane not disabled')
        else:
            raise InterlockFailure('unsupported deterministic proof route')
    return True



def fixture_manifest(config,runtime,fixture,objects,scratch):
    p=paths(config)
    if fixture=='synthetic':
        return generate(scratch/'objects.git',runtime['git']['path']),None
    synthetic=json.loads((p['OUT']/'binding/synthetic.json').read_bytes())
    synthetic_receipt=pin(p['OUT']/'binding/synthetic.json')
    if (synthetic.get('schema')!='scout.x1.binding-acceptance.v2' or synthetic['status']!='passed'
            or set(synthetic['acceptance'])!=set(ACCEPTANCE) or not all(acceptance_flags(synthetic).values())):
        raise ValueError('passing synthetic acceptance required before Django')
    verify_tree_pins(synthetic)
    object_path=p['OUT']/'objects/django__django.json'
    if objects is None or Path(objects).absolute()!=object_path:
        raise ValueError('canonical Django objects receipt required')
    record=json.loads(object_path.read_bytes());verify_tree_pins(record)
    if record['status']!='passed' or record['repository']!='django/django':
        raise ValueError('Django objects not verified')
    manifest={'database':record['database_realpath'],'head':record['actual_tip'],
              'parent':record['django_first_parent']['sha'],'synthetic_receipt':synthetic_receipt}
    if manifest['head']!='a05e16a7230fe33d3f43f4a3aaed103cebb42efc':
        raise ValueError('substituted Django tip')
    parent=git(runtime['git']['path'],manifest['database'],'rev-parse','--verify',manifest['head']+'^1').decode().strip()
    if parent!=manifest['parent']:
        raise ValueError('substituted Django first parent')
    return manifest,synthetic['positive_control']


def execute_fixture(config,runtime,fixture,objects,scratch):
    p=paths(config);manifest,positive=fixture_manifest(config,runtime,fixture,objects,scratch)
    routes=('bound','reference','synthetic-unbound') if fixture=='synthetic' else ('bound','reference')
    roots={};sources={};resolved={};records={}
    # Materialize and resolve every parent/head policy before the first scanner.
    for route in routes:
        for revision in ('parent','head'):
            key=(route,revision);root=scratch/ROUTE_DIRS[route]/revision
            roots[key]=root
            sources[key]=materialize(manifest['database'],manifest[revision],root,runtime['git']['path'])
            resolved[key]=run_revision(root,manifest[revision],route,scratch/'resolve'/route/revision,runtime,
                                      p['STUDY']/'preparation/model-free-policy.json',p['OUT']/'closure.json',resolve_only=True)
            if route in ('bound','reference'):verify_model_proof(resolved[key]['policy'])
            if route=='bound':verify_bound_policy(resolved[key]['policy'])
    if fixture=='synthetic':
        positive=[]
        for revision in ('parent','head'):
            key=('synthetic-unbound',revision)
            result=run_revision(roots[key],manifest[revision],key[0],scratch/'logs/positive'/revision,runtime,
                                p['STUDY']/'preparation/model-free-policy.json',p['OUT']/'closure.json',
                                expected_policy=resolved[key]['effective_record'])
            verify_materialization(roots[key],sources[key]);positive.append(result)
    for route in ('bound','reference'):
        for revision in ('parent','head'):
            key=(route,revision);root=roots[key]
            if revision=='head':carry_state(roots[route,'parent'],root)
            verify_materialization(root,sources[key])
            records[key]=run_revision(root,manifest[revision],route,scratch/'logs'/route/revision,runtime,
                                      p['STUDY']/'preparation/model-free-policy.json',p['OUT']/'closure.json',
                                      expected_policy=resolved[key]['effective_record'])
            if route=='bound':verify_bound_policy(records[key]['policy'])
            verify_model_proof(records[key]['policy'])
            verify_model_proof(records[key]['result'],route=route)
            verify_materialization(root,sources[key])
    parity={revision:compare(records['bound',revision]['result'],records['reference',revision]['result']) for revision in ('parent','head')}
    if fixture=='synthetic' and records['bound','parent']['result']['issue_ids']==records['bound','head']['result']['issue_ids']:
        raise InterlockFailure('synthetic parent/head issue sets must differ')
    evidence={'revisions':{'parent':manifest['parent'],'head':manifest['head']},'fixture_manifest':manifest,
              'source_identity':[{'route':r,'revision':v,'root':str(roots[r,v]),'files':sources[r,v]} for r,v in sources],
              'resolved_policies':{r:{v:resolved[r,v] for v in ('parent','head')} for r in routes},
              'parity':parity,'positive_control':positive,
              'runs':{r:{v:records[r,v] for v in ('parent','head')} for r in ('bound','reference')}}
    if fixture=='django-pinned-pair':
        evidence['synthetic_receipt']=manifest['synthetic_receipt']
    return evidence


def acceptance_flags(evidence):
    revisions={'parent','head'}
    if set(evidence['runs']) != {'bound','reference'} or set(evidence['parity']) != revisions:
        raise InterlockFailure('exact bound/reference parent/head layout required')
    if set(evidence['revisions']) != revisions or evidence['revisions']['parent'] == evidence['revisions']['head']:
        raise InterlockFailure('distinct parent/head revisions required')
    for route in ('bound','reference'):
        if set(evidence['runs'][route]) != revisions or set(evidence['resolved_policies'][route]) != revisions:
            raise InterlockFailure('missing or substituted revision evidence')
        for revision in ('parent','head'):
            expected_revision=evidence['revisions'][revision]
            for record in (evidence['runs'][route][revision]['policy'],
                           evidence['runs'][route][revision]['result'],
                           evidence['resolved_policies'][route][revision]['policy']):
                if record.get('route')!=route or record.get('revision')!=expected_revision:
                    raise InterlockFailure('route/revision identity mismatch')
    runs=[value for route in evidence['runs'].values() for value in route.values()]
    proof=all(verify_model_proof(evidence['resolved_policies'][r][v]['policy'])
              and verify_model_proof(evidence['runs'][r][v]['policy'])
              and verify_model_proof(evidence['runs'][r][v]['result'],route=r)
              for r in ('bound','reference') for v in ('parent','head'))
    controls=evidence['positive_control']
    parity=[compare(evidence['runs']['bound'][v]['result'],evidence['runs']['reference'][v]['result'])
            for v in ('parent','head')]
    return {'policy_bound_both_revisions':all(verify_bound_policy(evidence['resolved_policies']['bound'][v]['policy'])
                and verify_bound_policy(evidence['runs']['bound'][v]['policy']) for v in ('parent','head')),
            'deterministic_model_free_proof':proof,
            'zero_model_attempts':all(x['interlock']['model_attempts']==0 for x in runs),
            'exact_nonempty_issue_id_parity':all(x['equal'] and x['unique_count']>0 for x in parity),
            'positive_control_expected_failure':len(controls)==2 and
                [x['interlock'].get('provider') for x in controls]==['claude','codex'] and
                all(x['interlock']['expected_interlock_failure'] and x['interlock'].get('stub_exit_code')==97
                    and x['interlock']['model_attempts']>0 for x in controls),
            'scanner_health_complete':all(x['result']['scanner_health_complete'] for x in runs)}


def complete_operation(stream,canonical,journal,receipt,work):
    """Canonical path is already exclusively reserved; all ordinary errors terminate it."""
    try:
        receipt.update(work())
        receipt['acceptance']=acceptance_flags(receipt)
        if not all(receipt['acceptance'].values()):raise InterlockFailure('acceptance evidence incomplete')
        receipt['status']='passed';code=0
    except Exception as exc:
        receipt['status']='refused';receipt['failures'].append(type(exc).__name__+': '+str(exc));code=1
    json.dump(receipt,stream,sort_keys=True,indent=2);stream.write('\n');stream.flush();os.fsync(stream.fileno())
    journal.finish(receipt['status'],outputs=[pin(canonical)],reason='; '.join(receipt['failures']) or None,exit_code=code)
    return code


def scratch_output(config, output):
    scratch=paths(config)['SCRATCH'].resolve()
    target=Path(output).absolute()
    if '..' in target.parts or not target.resolve().is_relative_to(scratch) or target.resolve()==scratch:
        raise ValueError('rehearsal output must be strictly inside SCRATCH')
    if target.resolve().is_relative_to(paths(config)['OUT'].resolve()):
        raise ValueError('rehearsal output cannot be inside OUT')
    current=target
    while current!=current.parent:
        if current.is_symlink():
            raise ValueError('rehearsal output symlink refused')
        current=current.parent
    return target


def reviewed_inputs(config, source_pins, receipt):
    p=paths(config)
    receipt['configuration_amendment_reviews']=verify_amendment_reviews(config)
    if 'OPENROUTER_API_KEY' in os.environ:
        raise InterlockFailure('key present')
    if config.get('binding_recipe')!='preparation/model-free-binding-check-v2.json':
        raise InterlockFailure('explicit v2 binding recipe required')
    recipe=p['STUDY']/config['binding_recipe']
    amendment=p['STUDY']/'preparation/model-free-binding-amendment.md'
    adopted=pin(EXEC/'run-config-amendments/001-binding-v2.json')
    proposal=pin(EXEC/'run-config-amendments-proposed/001-binding-v2.json')
    amendment_review_path=p['OUT']/'reviews/binding-v2-amendment.json'
    amendment_review_pin=pin(amendment_review_path)
    if (adopted['sha256']!=proposal['sha256'] or
            not any(row.get('path')==adopted['path'] and row.get('sha256')==adopted['sha256']
                    and row.get('adoption_review')==amendment_review_pin
                    for row in getattr(config,'amendments',()))):
        raise InterlockFailure('v2 binding amendment not applied from exact reviewed proposal and adoption receipt')
    amendment_review=review_receipt(amendment_review_path,'model-free-binding-v2-amendment')
    amendment_reviewed={(x['path'],x['sha256']) for x in amendment_review['reviewed']}
    source_path=p['OUT']/config['source_manifest']
    # Adoption freezes only protocol; implementation can be re-reviewed after rehearsal.
    required={(x['path'],x['sha256']) for x in (proposal,pin(recipe),pin(amendment))}
    if len(amendment_review['reviewed'])!=3 or amendment_reviewed!=required:
        raise InterlockFailure('v2 amendment review must bind exactly proposal, recipe and amendment')
    receipt.update(binding_recipe=pin(recipe),binding_amendment=pin(amendment),
                   binding_amendment_review=pin(amendment_review_path))
    review=review_receipt(p['OUT']/'reviews/implementation.json','execution-implementation')
    reviewed={(x['path'],x['sha256']) for x in review['reviewed']}
    source=json.loads(source_path.read_bytes());verify_tree_pins(source)
    if (str(source_path),pin(source_path)['sha256']) not in reviewed:
        raise ValueError('implementation review does not bind source manifest')
    all_sources={(x['path'],x['sha256']) for x in source['files']}
    if not all((x['path'],x['sha256']) in all_sources for x in source_pins):
        raise ValueError('binding source omitted from reviewed manifest')
    closure_path=p['OUT']/'closure.json';closure=json.loads(closure_path.read_bytes())
    if closure['status']!='passed' or not closure['complete']:
        raise ValueError('complete installed closure required')
    verify_tree_pins(closure);receipt['runtime_closure']=pin(closure_path)
    runtime_path=p['OUT']/'binding/runtime.json'
    runtime=json.loads(runtime_path.read_bytes());validate_runtime(runtime,closure)
    if (str(runtime_path),pin(runtime_path)['sha256']) not in reviewed:
        raise ValueError('binding runtime not scoped-reviewed')
    receipt['runtime']=pin(runtime_path)
    return runtime


def verify_rehearsal(config, report_path, source_pins):
    p=paths(config);report_path=scratch_output(config,report_path)
    review_path=p['OUT']/'reviews/binding-v2-rehearsal.json'
    review=review_receipt(review_path,'model-free-binding-v2-rehearsal')
    reviewed={(x['path'],x['sha256']) for x in review['reviewed']}
    required=[pin(report_path),pin(p['OUT']/'binding/runtime.json'),*source_pins]
    if not all((x['path'],x['sha256']) in reviewed for x in required):
        raise InterlockFailure('rehearsal review does not bind report, runtime and source pins')
    report=json.loads(report_path.read_bytes());verify_tree_pins(report)
    if report.get('clock')!=pin(p['OUT']/'clock.json') or report.get('key_present') is not False:
        raise InterlockFailure('rehearsal immutable clock or key provenance mismatch')
    if (report.get('schema')!='scout.x1.binding-rehearsal.v2' or report.get('status')!='rehearsal'
            or report.get('fixture')!='synthetic' or report.get('failures')
            or set(report.get('acceptance_predicate',{}))!=set(ACCEPTANCE)
            or not all(report['acceptance_predicate'].values())
            or not all(acceptance_flags(report).values())):
        raise InterlockFailure('reviewed rehearsal predicates incomplete')
    if report['runs']['bound']['parent']['result']['issue_ids']==report['runs']['bound']['head']['result']['issue_ids']:
        raise InterlockFailure('rehearsal parent/head occurrence sets must differ')
    if report['runtime']!=pin(p['OUT']/'binding/runtime.json') or report['sources']!=source_pins:
        raise InterlockFailure('rehearsal runtime/source identity changed')
    return {'report':pin(report_path),'review':pin(review_path),'record':pin(review['record']['path'])}


def rehearse(config, args, source_pins):
    if args.fixture!='synthetic' or args.objects or args.rehearsal_report:
        raise ValueError('rehearsal is synthetic-only without object or prior report inputs')
    read_clock(config,args.config,args.clock)
    if 'OPENROUTER_API_KEY' in os.environ:
        raise InterlockFailure('key present')
    target=scratch_output(config,args.output)
    scratch=scratch_output(config,str(target)+'.work')
    if scratch.exists():
        raise ValueError('rehearsal work directory already exists')
    target.parent.mkdir(parents=True,mode=0o700,exist_ok=True)
    # Scratch reservation only: this path never opens a canonical receipt or journal.
    with target.open('x') as stream:
        report={'schema':'scout.x1.binding-rehearsal.v2','status':'rehearsal','fixture':'synthetic',
                'sources':source_pins,'clock':pin(args.clock),'key_present':False,
                'failures':[],'acceptance_predicate':dict.fromkeys(ACCEPTANCE,False)}
        try:
            runtime=reviewed_inputs(config,source_pins,report)
            scratch.mkdir(mode=0o700,exist_ok=False)
            report.update(execute_fixture(config,runtime,'synthetic',None,scratch))
            report['acceptance_predicate']=acceptance_flags(report)
        except Exception as exc:
            report['failures'].append(type(exc).__name__+': '+str(exc))
        # Preserve every available route result and audited-seam attempt log even
        # when a scanner or an interlock refused before producing full evidence.
        report['route_diagnostics']=[]
        if scratch.exists():
            for path in sorted(scratch.rglob('*')):
                if path.is_file() and path.name in ('result.json','attempts.jsonl','stderr','stdout'):
                    report['route_diagnostics'].append({'file':pin(path),'content':path.read_text(errors='replace')})
        json.dump(report,stream,sort_keys=True,indent=2);stream.write('\n');stream.flush();os.fsync(stream.fileno())
    return 0 if not report['failures'] and all(report['acceptance_predicate'].values()) else 1


def main(argv=None):
    parser=argparse.ArgumentParser()
    parser.add_argument('--config',required=True);parser.add_argument('--clock',required=True)
    parser.add_argument('--fixture',choices=('synthetic','django-pinned-pair'),required=True)
    parser.add_argument('--objects');parser.add_argument('--output',required=True)
    parser.add_argument('--rehearsal',action='store_true');parser.add_argument('--rehearsal-report')
    args=parser.parse_args(argv);config=load_config(args.config);p=paths(config)
    source_pins=[pin(x) for x in sorted(HERE.rglob('*')) if x.is_file() and '__pycache__' not in x.parts]+[pin(EXEC/'model_free/main.go')]
    if args.rehearsal:
        return rehearse(config,args,source_pins)
    canonical=p['OUT']/'binding'/('synthetic.json' if args.fixture=='synthetic' else 'django.json')
    if Path(args.output).absolute()!=canonical:raise ValueError('binding output noncanonical')
    rehearsal=None
    if args.fixture=='synthetic':
        if not args.rehearsal_report:raise ValueError('reviewed synthetic rehearsal report required')
        rehearsal=verify_rehearsal(config,args.rehearsal_report,source_pins)
    elif args.rehearsal_report:
        raise ValueError('Django uses the canonical synthetic prerequisite, not a rehearsal report')
    with operation_lock(config):
        canonical.parent.mkdir(parents=True,exist_ok=True)
        with canonical.open('x') as stream:
            stream.flush();os.fsync(stream.fileno())
            journal=Journal(config,args.config,args.clock,'c',source_pins=source_pins)
            receipt={'schema':'scout.x1.binding-acceptance.v2','status':'refused','clock':pin(args.clock),
                     'fixture':args.fixture,'key_present':'OPENROUTER_API_KEY' in os.environ,
                     'acceptance':dict.fromkeys(ACCEPTANCE,False),'failures':[],'sources':source_pins}
            if rehearsal is not None:receipt['reviewed_rehearsal']=rehearsal
            def work():
                runtime=reviewed_inputs(config,source_pins,receipt)
                scratch=p['SCRATCH']/('synthetic' if args.fixture=='synthetic' else 'django')
                scratch.mkdir(parents=True,mode=0o700,exist_ok=False)
                return execute_fixture(config,runtime,args.fixture,args.objects,scratch)
            return complete_operation(stream,canonical,journal,receipt,work)

if __name__=='__main__':raise SystemExit(main())

"""Synthetic unit evidence only; never scanner/acceptance or a real model run."""
import ast
import copy
import json
import os
import re
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import Mock, call, patch

EXEC=Path(__file__).resolve().parents[1]
BINDING=EXEC/'binding'
sys.path.insert(0,str(EXEC));sys.path.insert(0,str(BINDING))
from fixtures import FIXTURES, generate, materialize, safe_path, verify_materialization
from interlock import InterlockFailure, compare, evaluate, stub_source, make_stubs
from acceptance import (ROUTE_DIRS, acceptance_flags, carry_state, complete_operation, frontend_argv,
                        private_env, verify_bound_policy, verify_effective_record, verify_model_proof)
from preparation import verify_tree_pins
from verify_inputs import pin


def model_proof():
    return {'model_proof':dict.fromkeys(('effective_config_bound','model_lanes_disabled','jev_disabled',
            'association_disabled','pr_metadata_absent','heuristic_only'),True),'model_call_impossible':True}


def lane_result(route):
    record=model_proof()
    record.update(jev_result='absent',association_result='disabled' if route=='bound' else 'structurally_absent',
                  model_stages='disabled' if route=='bound' else 'structurally_absent',lane_report=None,
                  mapping={'Source':'heuristic'})
    if route=='bound':
        record['lane_report']=[{'lane':name,'status':'skipped','units':1,
                               'reason':'disabled' if name.startswith('association') else 'skipped (not opted in)'}
                              for name in ('contextual','clawpatch-review','association','association-fallback')]
    return record


class ModelAttemptTests(unittest.TestCase):
    def attempt(self,provider):
        return {'argv':['/private/stubs/'+provider]+(['-p','--output-format','json'] if provider=='claude'
                else ['exec','--sandbox','read-only']),'kind':'model-stub-attempt','pid':101,'ppid':100,'exit_code':97}

    def test_zero_and_positive_logs(self):
        self.assertEqual(evaluate([])['model_attempts'],0)
        for provider in ('claude','codex'):
            rows=[self.attempt(provider)]
            with self.assertRaises(InterlockFailure):evaluate(rows)
            verdict=evaluate(rows,positive=True,expected_provider=provider)
            self.assertTrue(verdict['expected_interlock_failure'])
            self.assertEqual(verdict['stub_exit_code'],97)

    def test_positive_boundaries(self):
        row=self.attempt('claude')
        cases=[([], 'claude'),([row],None),([row],'codex')]
        for key,value in (('exit_code',0),('kind','synthetic'),('pid',0),('ppid',0),
                          ('argv',['/private/stubs/claude','--not-production']),('argv','claude')):
            cases.append(([dict(row,**{key:value})],'claude'))
        for rows,provider in cases:
            with self.subTest(rows=rows,provider=provider):
                with self.assertRaises(InterlockFailure):evaluate(rows,positive=True,expected_provider=provider)

    def test_actual_failing_stub_evidence(self):
        for provider in ('claude','codex'):
            with self.subTest(provider=provider),tempfile.TemporaryDirectory() as directory:
                stubs=Path(directory)/'stubs';log=make_stubs(stubs,sys.executable)
                self.assertEqual(log.read_bytes(),b'')
                self.assertEqual(log.stat().st_mode&0o777,0o600)
                args=['-p','--output-format','json'] if provider=='claude' else ['exec','--sandbox','read-only']
                child=subprocess.run([str(stubs/provider),*args],check=False)
                self.assertEqual(child.returncode,97)
                rows=[json.loads(line) for line in log.read_text().splitlines()]
                self.assertEqual(len(rows),1)
                self.assertTrue(evaluate(rows,positive=True,expected_provider=provider)['expected_interlock_failure'])

    def test_stub_flags_planted_effect_and_null(self):
        expected='#!/synthetic/python -ISB'
        self.assertEqual(stub_source('/synthetic/python','log').splitlines()[0],expected)
        with self.assertRaises(AssertionError):
            self.assertEqual('#!/synthetic/python',expected)

    def test_stub_bytecode_planted_effect_and_null(self):
        for provider in ('claude','codex'):
            for repaired in (False,True):
                with self.subTest(provider=provider,repaired=repaired), tempfile.TemporaryDirectory() as directory:
                    root=Path(directory);cache=root/'pycache'
                    if repaired:
                        log=make_stubs(root/'stubs',sys.executable)
                    else:
                        (root/'stubs').mkdir();log=root/'attempts.jsonl';log.touch()
                        target=root/'stubs'/provider
                        target.write_text(stub_source(sys.executable,log).replace(' -ISB\n','\n',1))
                        target.chmod(0o700)
                    args=['-p','--output-format','json'] if provider=='claude' else ['exec','--sandbox','read-only']
                    result=subprocess.run([str(root/'stubs'/provider),*args],
                                          env={'PATH':'/usr/bin:/bin','PYTHONPYCACHEPREFIX':str(cache)},
                                          check=False,capture_output=True)
                    self.assertEqual(result.returncode,97,result.stderr)
                    rows=[json.loads(line) for line in log.read_text().splitlines()]
                    self.assertEqual(len(rows),1)
                    self.assertTrue(evaluate(rows,positive=True,expected_provider=provider)['expected_interlock_failure'])
                    if repaired:
                        self.assertFalse(cache.exists())
                    else:
                        self.assertTrue(list(cache.rglob('*.pyc')))


class EnvironmentAndSeamTests(unittest.TestCase):

    def test_private_env_stubs_all_provider_paths_and_no_outer_sandbox(self):
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory);stubs=root/'stubs'
            with patch.dict(os.environ,{'CLAWPATCH_CODEX_BIN':'/real/codex','CLAWPATCH_CODEX_SANDBOX':'bypass',
                                        'CLAWPATCH_CLAUDE_BIN':'/real/claude','OPENROUTER_API_KEY':'synthetic'},clear=True):
                env=private_env(root,'/pinned/tools',stubs,positive=True)
            self.assertEqual(env['PATH'],str(stubs)+':/pinned/tools')
            self.assertEqual(env['CLAWPATCH_CLAUDE_BIN'],str(stubs/'claude'))
            self.assertEqual(env['CLAWPATCH_CODEX_SANDBOX'],'read-only')
            self.assertEqual({key for key in env if key.startswith('CLAWPATCH_CODEX_')},{'CLAWPATCH_CODEX_SANDBOX'})
            self.assertNotIn('OPENROUTER_API_KEY',env)
            runtime={'frontend':{'path':'/pinned/frontend'},'clawpatch':{'path':'/pinned/clawpatch.js'}}
            argv=frontend_argv(runtime,root,'sha','bound',root/'policy',root/'effective',root/'result')
            self.assertEqual(argv[0],'/pinned/frontend');self.assertNotIn('sandbox-exec',' '.join(argv))
            resolve=frontend_argv(runtime,root,'sha','bound',root/'policy',root/'effective',resolve_only=True)
            self.assertIn('--resolve-only',resolve);self.assertNotIn('--output',resolve)

    def test_private_environment_bytecode_planted_effect_and_null(self):
        for protected in (True,False):
            with self.subTest(protected=protected), tempfile.TemporaryDirectory() as directory:
                root=Path(directory);cache=root/'pycache'
                env=private_env(root,'/usr/bin:/bin',root/'stubs')
                self.assertEqual(env['PYTHONDONTWRITEBYTECODE'],'1')
                if not protected:
                    del env['PYTHONDONTWRITEBYTECODE']
                env['PYTHONPYCACHEPREFIX']=str(cache)
                result=subprocess.run([sys.executable,'-c','import json'],env=env,
                                      check=False,capture_output=True)
                self.assertEqual(result.returncode,0,result.stderr)
                if protected:
                    self.assertFalse(cache.exists())
                else:
                    self.assertTrue(list(cache.rglob('*.pyc')))

    def test_stub_is_logging_only(self):
        text=stub_source('/synthetic/python','/synthetic/attempts.jsonl')
        self.assertIn('sys.exit(97)',text);self.assertNotIn('subprocess',text);self.assertNotIn('execv',text)
        for binary in ('/path with space/python','/bad\npython'):
            with self.assertRaises(InterlockFailure):stub_source(binary,'log')

    def test_actual_frontend_launch_is_direct_and_private(self):
        import acceptance as controller
        import attest_closure
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory)
            runtime={'python':{'path':sys.executable},
                     'executables':[{'path':str(root/'tools/git')},{'path':str(root/'tools/node')}],
                     'frontend':{'path':'/pinned/frontend'},'clawpatch':{'path':'/pinned/clawpatch'}}
            def frontend(argv,**kwargs):
                effective=Path(argv[argv.index('--effective-output')+1])
                effective.write_text(json.dumps(model_proof()))
                Path(argv[argv.index('--output')+1]).write_text(json.dumps(lane_result('bound')))
                self.assertEqual(argv[0],'/pinned/frontend')
                self.assertEqual(kwargs['stdin'],subprocess.DEVNULL)
                self.assertEqual(kwargs['cwd'],root)
                self.assertNotIn('OPENROUTER_API_KEY',kwargs['env'])
                self.assertTrue(kwargs['env']['HOME'].startswith(str(root/'run')))
                self.assertEqual(kwargs['env']['PATH'],
                                 str(root/'run/stubs')+':'+controller.private_path(runtime))
                self.assertEqual(controller.private_path(runtime),str(root/'tools'))
                return Mock(returncode=0)
            with patch.dict(os.environ,{},clear=True), \
                 patch.object(attest_closure,'verify_installed_closure') as closure, \
                 patch.object(controller.subprocess,'run',side_effect=frontend) as invoked:
                record=controller.run_revision(root,'sha','bound',root/'run',runtime,root/'policy',root/'closure')
            invoked.assert_called_once()
            self.assertEqual(closure.call_args_list,[call(root/'closure')]*2)
            self.assertEqual(record['interlock']['model_attempts'],0)
            log=root/'run/stubs/attempts.jsonl'
            self.assertEqual(log.read_bytes(),b'')
            self.assertIn(pin(log),record['pins'])

    def test_missing_attempt_log_refuses_instead_of_passing_zero(self):
        import acceptance as controller
        import attest_closure
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory)
            runtime={'python':{'path':sys.executable},'executables':[],
                     'frontend':{'path':'/pinned/frontend'},'clawpatch':{'path':'/pinned/clawpatch'}}
            def frontend(argv,**kwargs):
                (root/'run/stubs/attempts.jsonl').unlink()
                return Mock(returncode=0)
            with patch.dict(os.environ,{},clear=True), \
                 patch.object(attest_closure,'verify_installed_closure'), \
                 patch.object(controller.subprocess,'run',side_effect=frontend):
                with self.assertRaises(FileNotFoundError):
                    controller.run_revision(root,'sha','bound',root/'run',runtime,root/'policy',root/'closure')

    def test_static_provider_binary_seams(self):
        repo=EXEC.parents[2]
        provider=repo/'third_party/clawpatch/src/providers'
        claude=(provider/'claude.ts').read_text();codex=(provider/'codex.ts').read_text()
        self.assertIn('process.env["CLAWPATCH_CLAUDE_BIN"]',claude)
        self.assertIn('runCommandArgs("codex", ["--version"]',codex)
        self.assertIn('runCommandArgs("codex", args',codex)
        self.assertEqual(set(re.findall(r'CLAWPATCH_CODEX_[A-Z_]+',codex)),
                         {'CLAWPATCH_CODEX_TIMEOUT_MS','CLAWPATCH_CODEX_SANDBOX'})
        self.assertNotIn('CLAWPATCH_CODEX_BIN',codex)


class FixtureAndParityTests(unittest.TestCase):
    def result(self):
        return {'issue_ids':['occurrence-a'],'multiplicity':{'occurrence-a':1},'scanner_health_complete':True,
                'durable_issue_count':1,'durable_links':{'occurrence-a':'issue:quality:existing'},'carried_issue_ids':[],
                'binding_rows_complete':True,'accounted_coverage':['fixture.py\0quality\0complexity'],
                'scanner_inputs':{'quality':['fixture.py']},'omissions':[],'mapping_gaps':[]}

    def test_exact_sets_health_multiplicity_coverage_and_durable_link(self):
        result=self.result();reference=dict(result,durable_issue_count=0,durable_links={})
        self.assertTrue(compare(result,reference)['equal'])
        changes={'issue_ids':['other'],'multiplicity':{'occurrence-a':2},'scanner_health_complete':False,
                 'accounted_coverage':[],'scanner_inputs':{},'omissions':['missing'],'mapping_gaps':['gap']}
        for name,value in changes.items():
            with self.subTest(name=name):
                with self.assertRaises(InterlockFailure):compare(result,dict(reference,**{name:value}))
        for changed in (dict(result,durable_links={}),dict(result,binding_rows_complete=False),dict(result,issue_ids=[])):
            with self.assertRaises(InterlockFailure):compare(changed,reference)

    def test_equal_length_roots_and_missing_head_state_directory(self):
        self.assertEqual(len(ROUTE_DIRS['bound']),len(ROUTE_DIRS['reference']))
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory);parent=root/'parent';head=root/'head'
            (parent/'.scout').mkdir(parents=True);head.mkdir()
            (parent/'.scout/policy.yaml').write_text('must not copy')
            (parent/'.scout/state.sqlite').write_bytes(b'synthetic state')
            (parent/'.scout/z-dir').mkdir();(parent/'.scout/z-dir/file').write_text('state')
            carry_state(parent,head)
            self.assertEqual((head/'.scout/state.sqlite').read_bytes(),b'synthetic state')
            self.assertFalse((head/'.scout/policy.yaml').exists())
            self.assertEqual((head/'.scout').stat().st_mode&0o777,0o700)

    def test_source_manifest_not_mistaken_for_filesystem_pin(self):
        executable=shutil.which('git')
        self.assertIsNotNone(executable)
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory).resolve();fixture=generate(root/'objects.git',executable)
            rows=materialize(fixture['database'],fixture['parent'],root/'parent',executable)
            receipt={'fixture_manifest':fixture,'source_identity':rows,'source':pin(BINDING/'fixtures.py')}
            verify_tree_pins(receipt)
            self.assertTrue(verify_materialization(root/'parent',rows))
            self.assertTrue(all('repo_path' in row and 'content_sha256' in row and row['revision']==fixture['parent'] for row in rows))
            again=generate(root/'same.git',executable)
            self.assertEqual((fixture['parent'],fixture['head']),(again['parent'],again['head']))
            self.assertNotEqual(fixture['parent'],fixture['head'])

    def test_materialized_origin_planted_effect_and_null(self):
        import fixtures
        executable=shutil.which('git');self.assertIsNotNone(executable)
        real_git=fixtures.git
        def without_origin(executable,directory,*args,**kwargs):
            if args and args[0]=='config':
                return b''
            return real_git(executable,directory,*args,**kwargs)
        def origin(root):
            try:
                return real_git(executable,root,'config','--get','remote.origin.url').decode().strip()
            except subprocess.CalledProcessError as error:
                self.assertEqual(error.returncode,1)
                return None
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory).resolve();fixture=generate(root/'objects.git',executable)
            expected=str(Path(fixture['database']).resolve());self.assertTrue(expected)
            planted={}
            with patch.object(fixtures,'git',side_effect=without_origin):
                for revision,name in (('parent','p'),('head','h')):
                    planted[revision]=materialize(fixture['database'],fixture[revision],root/name,executable)
                    self.assertIsNone(origin(root/name))
                    self.assertTrue(verify_materialization(root/name,planted[revision]))
            self.assertNotEqual(str(root/'p'),str(root/'h'))
            self.assertFalse(origin(root/'p') is not None and origin(root/'p')==origin(root/'h'))
            for name in ('p','h'):
                shutil.rmtree(root/name)
            for revision,name in (('parent','p'),('head','h')):
                rows=materialize(fixture['database'],fixture[revision],root/name,executable)
                self.assertEqual(origin(root/name),expected)
                self.assertTrue(verify_materialization(root/name,rows))
                self.assertEqual(json.dumps(rows,sort_keys=True).encode(),
                                 json.dumps(planted[revision],sort_keys=True).encode())
            other=generate(root/'other.git',executable)
            materialize(other['database'],other['parent'],root/'q',executable)
            self.assertEqual(origin(root/'q'),str(Path(other['database']).resolve()))
            self.assertNotEqual(origin(root/'q'),origin(root/'p'))

    def test_fixture_trigger_constructs_and_revision_difference(self):
        unchanged=(FIXTURES/'common/unchanged.py').read_text()
        self.assertIn('yaml.load(data, Loader=yaml.Loader)',unchanged)
        parent=(FIXTURES/'parent/changed.py').read_text();head=(FIXTURES/'head/changed.py').read_text()
        self.assertEqual(parent.splitlines()[3],'        result.append(value)')
        self.assertEqual(head.splitlines()[4],'        result.append(value)')
        tree=ast.parse(head)
        complex_fn=next(x for x in tree.body if isinstance(x,ast.FunctionDef) and x.name=='complex_only_at_head')
        self.assertGreater(sum(isinstance(x,ast.If) for x in ast.walk(complex_fn))+1,10)
        self.assertNotIn('complex_only_at_head',parent)

    def test_path_escape_and_codegraph_refused(self):
        for value in ('.codegraph/db','src/.codegraph/file','../outside','/absolute','.git/config'):
            with self.assertRaises(ValueError):safe_path(value)


class DeterministicProofTests(unittest.TestCase):
    def evidence(self):
        policy={'profile':'report','decisions':{'enabled':False},'model_cli':{'lanes':[]},
                'mapping':{'association':{'enabled':False}}}
        pre={'key_present':False,'association_nil':True,'measures_nil':True,'clawpatch_egress':False,
             'contextual_egress':False,'model_lanes':[],'effective_policy':policy,**model_proof()}
        evidence={'revisions':{'parent':'parent-sha','head':'head-sha'},'runs':{},'resolved_policies':{},
                  'parity':dict.fromkeys(('parent','head'),{'equal':True,'unique_count':1}),
                  'positive_control':[{'interlock':{'expected_interlock_failure':True,'model_attempts':1,
                                      'stub_exit_code':97,'provider':p}} for p in ('claude','codex')]}
        for route in ('bound','reference'):
            evidence['runs'][route]={};evidence['resolved_policies'][route]={}
            for revision,sha in evidence['revisions'].items():
                effective={**copy.deepcopy(pre),'route':route,'revision':sha}
                result={**FixtureAndParityTests().result(),**lane_result(route),'route':route,'revision':sha}
                evidence['runs'][route][revision]={'policy':effective,'result':result,'interlock':evaluate([])}
                evidence['resolved_policies'][route][revision]={'policy':copy.deepcopy(effective)}
        return evidence

    def test_all_six_flags_derived_from_complete_evidence(self):
        evidence=self.evidence()
        self.assertTrue(all(acceptance_flags(evidence).values()))
        self.assertEqual(len(acceptance_flags(evidence)),6)
        for route in ('bound','reference'):
            for revision in ('parent','head'):
                changed=copy.deepcopy(evidence)
                changed['resolved_policies'][route][revision]['policy']['model_call_impossible']=False
                with self.assertRaises(InterlockFailure):acceptance_flags(changed)

    def test_layout_and_identity_cannot_pass_vacuously(self):
        base=self.evidence()
        for collection,key in (('runs','bound'),('parity','head'),('revisions','parent')):
            changed=copy.deepcopy(base);del changed[collection][key]
            with self.assertRaises(InterlockFailure):acceptance_flags(changed)
        for route in ('bound','reference'):
            changed=copy.deepcopy(base);changed['runs'][route]['head']=changed['runs'][route]['parent']
            with self.assertRaises(InterlockFailure):acceptance_flags(changed)

    def test_config_proof_and_actual_lanes_each_required(self):
        for name in model_proof()['model_proof']:
            changed=model_proof();changed['model_proof'][name]=False
            with self.assertRaises(InterlockFailure):verify_model_proof(changed)
        for route in ('bound','reference'):
            good=lane_result(route);self.assertTrue(verify_model_proof(good,route))
            for name,value in (('model_call_impossible',False),('association_result','enabled'),
                               ('jev_result','present'),('model_stages','enabled_control'),
                               ('mapping',{'Source':'agent'}),('mapping',{}),('mapping',{'source':'heuristic'})):
                with self.assertRaises(InterlockFailure):verify_model_proof(dict(good,**{name:value}),route)
        bound=lane_result('bound')
        for rows in ([],bound['lane_report'][:-1],bound['lane_report']+[bound['lane_report'][0]]):
            with self.assertRaises(InterlockFailure):verify_model_proof(dict(bound,lane_report=rows),'bound')
        for name,value in (('status','ran'),('reason','no model CLI')):
            changed=copy.deepcopy(bound);changed['lane_report'][0][name]=value
            with self.assertRaises(InterlockFailure):verify_model_proof(changed,'bound')
        for requirement in ('model-cli','decision-egress'):
            extra={'lane':'unexpected','requires':requirement,'status':'skipped','reason':'disabled'}
            with self.assertRaises(InterlockFailure):
                verify_model_proof(dict(bound,lane_report=bound['lane_report']+[extra]),'bound')


class RehearsalTests(unittest.TestCase):
    def test_rehearsal_main_never_opens_out_or_journal(self):
        import acceptance as controller
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory).resolve()
            config={'paths':{'SCRATCH':str(root/'scratch'),'OUT':str(root/'out'),'STUDY':str(root/'study')}}
            output=root/'scratch/report.json'
            clock=root/'clock.json';clock.write_text('{}')
            evidence=DeterministicProofTests().evidence()
            with patch.object(controller,'load_config',return_value=config), \
                 patch.object(controller,'read_clock'),patch.dict(os.environ,{},clear=True), \
                 patch.object(controller,'reviewed_inputs',return_value={}), \
                 patch.object(controller,'execute_fixture',return_value=evidence), \
                 patch.object(controller,'Journal') as journal,patch.object(controller,'operation_lock') as lock:
                code=controller.main(['--config','config','--clock',str(clock),'--fixture','synthetic',
                                      '--rehearsal','--output',str(output)])
            self.assertEqual(code,0);journal.assert_not_called();lock.assert_not_called()
            self.assertFalse((root/'out').exists())
            report=json.loads(output.read_bytes())
            self.assertEqual(report['status'],'rehearsal')
            self.assertNotIn('acceptance',report)
            self.assertTrue(all(report['acceptance_predicate'].values()))

    def test_rehearsal_rejects_canonical_escape_and_symlinks(self):
        import acceptance as controller
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory).resolve();scratch=root/'scratch';scratch.mkdir()
            config={'paths':{'SCRATCH':str(scratch),'OUT':str(root/'out')}}
            (scratch/'outside').symlink_to(root,target_is_directory=True)
            (scratch/'inside').mkdir();(scratch/'alias').symlink_to(scratch/'inside',target_is_directory=True)
            for target in (root/'out/binding/synthetic.json',scratch/'../escape.json',
                           scratch/'outside/report.json',scratch/'alias/report.json',scratch):
                with self.subTest(target=target):
                    with self.assertRaises(ValueError):controller.scratch_output(config,target)
            self.assertFalse((root/'out').exists())

    def test_canonical_gate_happens_before_lock_reservation_and_journal(self):
        import acceptance as controller
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory).resolve();config={'paths':{'OUT':str(root/'out'),'SCRATCH':str(root/'scratch')}}
            target=root/'out/binding/synthetic.json'
            with patch.object(controller,'load_config',return_value=config), \
                 patch.object(controller,'verify_rehearsal',side_effect=InterlockFailure('not reviewed')), \
                 patch.object(controller,'Journal') as journal,patch.object(controller,'operation_lock') as lock:
                with self.assertRaises(InterlockFailure):
                    controller.main(['--config','config','--clock','clock','--fixture','synthetic',
                                     '--output',str(target),'--rehearsal-report',str(root/'scratch/report.json')])
            journal.assert_not_called();lock.assert_not_called();self.assertFalse((root/'out').exists())

    def test_reviewed_rehearsal_requires_current_pins_and_every_predicate(self):
        import acceptance as controller
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory).resolve();out=root/'out';scratch=root/'scratch';scratch.mkdir()
            (out/'binding').mkdir(parents=True);(out/'reviews').mkdir()
            runtime=out/'binding/runtime.json';runtime.write_text('{}')
            clock=out/'clock.json';clock.write_text('{}')
            review=out/'reviews/binding-v2-rehearsal.json';review.write_text('{}')
            record=root/'rehearsal-council.txt';record.write_text('synthetic review record')
            source=pin(BINDING/'interlock.py');report_path=scratch/'report.json'
            config={'paths':{'OUT':str(out),'SCRATCH':str(scratch)}}
            report=DeterministicProofTests().evidence()
            report.update(schema='scout.x1.binding-rehearsal.v2',status='rehearsal',fixture='synthetic',
                          failures=[],runtime=pin(runtime),sources=[source],clock=pin(clock),key_present=False)
            for route in ('bound','reference'):
                result=report['runs'][route]['head']['result']
                result.update(issue_ids=['head-occurrence'],multiplicity={'head-occurrence':1},
                              durable_links={'head-occurrence':'head-issue'})
            report['acceptance_predicate']=acceptance_flags(report)
            def authorize():
                report_path.write_text(json.dumps(report))
                return patch.object(controller,'review_receipt',return_value={
                    'record':pin(record),'reviewed':[pin(report_path),pin(runtime),source]})
            with authorize():
                self.assertEqual(controller.verify_rehearsal(config,report_path,[source])['report'],pin(report_path))
            for flag in report['acceptance_predicate']:
                report['acceptance_predicate'][flag]=False
                with authorize(),self.assertRaises(InterlockFailure):
                    controller.verify_rehearsal(config,report_path,[source])
                report['acceptance_predicate'][flag]=True
            with authorize(),self.assertRaises(InterlockFailure):
                controller.verify_rehearsal(config,report_path,[])

    def test_wrong_or_missing_rehearsal_clock_refuses_before_scratch_writes(self):
        import acceptance as controller
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory).resolve()
            config={'paths':{'SCRATCH':str(root/'scratch'),'OUT':str(root/'out')}}
            for reason in ('missing clock','wrong immutable clock'):
                with patch.object(controller,'load_config',return_value=config), \
                     patch.object(controller,'read_clock',side_effect=ValueError(reason)), \
                     patch.object(controller,'Journal') as journal,patch.object(controller,'operation_lock') as lock:
                    with self.assertRaises(ValueError):
                        controller.main(['--config','config','--clock','bad-clock','--fixture','synthetic',
                                         '--rehearsal','--output',str(root/'scratch/report.json')])
                journal.assert_not_called();lock.assert_not_called()
                self.assertFalse((root/'out').exists());self.assertFalse((root/'scratch').exists())




class CompletionTests(unittest.TestCase):
    def test_both_revisions_resolved_before_any_scanner(self):
        import acceptance as controller
        policy={'profile':'report','decisions':{'enabled':False},'model_cli':{'lanes':[]},'mapping':{'association':{'enabled':False}}}
        pre={'key_present':False,'association_nil':True,'measures_nil':True,'clawpatch_egress':False,
             'contextual_egress':False,'model_lanes':[],'effective_policy':policy}
        pre.update(model_proof())
        calls=[]
        def fake_run(root,revision,route,run,runtime,frozen,closure,resolve_only=False,expected_policy=None):
            calls.append((route,revision,resolve_only))
            result=FixtureAndParityTests().result()
            key='occurrence-'+revision
            result.update(issue_ids=[key],multiplicity={key:1},durable_links={key:'issue:'+revision})
            result.update(lane_result('bound' if route=='synthetic-unbound' else route))
            return {'policy':copy.deepcopy(pre),'effective_record':{'path':'synthetic-pre-resolved'},
                    'result':result,'interlock':{'expected_interlock_failure':route=='synthetic-unbound'}}
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory)
            config={'paths':{'STUDY':str(root),'OUT':str(root)}}
            with patch.object(controller,'fixture_manifest',return_value=({'database':'synthetic-db','parent':'parent','head':'head'},None)), \
                 patch.object(controller,'materialize',return_value=[]),patch.object(controller,'carry_state'), \
                 patch.object(controller,'verify_materialization'),patch.object(controller,'run_revision',side_effect=fake_run):
                controller.execute_fixture(config,{'git':{'path':'synthetic-git'}},'synthetic',None,root)
        self.assertEqual(len(calls),12)
        self.assertEqual([x[2] for x in calls[:6]],[True]*6)
        self.assertEqual([x[2] for x in calls[6:]],[False]*6)
        self.assertEqual({(r,v) for r,v,_ in calls[:6]},
                         {(r,v) for r in ('bound','reference','synthetic-unbound') for v in ('parent','head')})

    def test_real_materialization_and_carry_share_origin_planted_effect_and_null(self):
        import acceptance as controller
        import fixtures
        executable=shutil.which('git');self.assertIsNotNone(executable)
        real_git=fixtures.git
        def without_origin(executable,directory,*args,**kwargs):
            if args and args[0]=='config':
                return b''
            return real_git(executable,directory,*args,**kwargs)
        for repaired in (False,True):
            with self.subTest(repaired=repaired),tempfile.TemporaryDirectory() as directory:
                root=Path(directory).resolve();calls=[]
                policy={'profile':'report','decisions':{'enabled':False},
                        'model_cli':{'lanes':[]},'mapping':{'association':{'enabled':False}}}
                pre={'key_present':False,'association_nil':True,'measures_nil':True,
                     'clawpatch_egress':False,'contextual_egress':False,'model_lanes':[],
                     'effective_policy':policy}
                pre.update(model_proof())
                def fake_run(checkout,revision,route,run,runtime,frozen,closure,
                             resolve_only=False,expected_policy=None):
                    try:
                        origin=real_git(executable,checkout,'config','--get','remote.origin.url').decode().strip()
                    except subprocess.CalledProcessError as error:
                        self.assertEqual(error.returncode,1);origin=None
                    marker=checkout/'.scout/state-marker'
                    calls.append({'route':route,'revision':revision,'resolve_only':resolve_only,
                                  'origin':origin,'marker':marker.exists()})
                    if route in ('bound','reference') and not resolve_only and checkout.name=='parent':
                        marker.parent.mkdir(exist_ok=True);marker.write_bytes(b'parent state')
                    result=FixtureAndParityTests().result();key='occurrence-'+revision
                    result.update(issue_ids=[key],multiplicity={key:1},durable_links={key:'issue:'+revision})
                    result.update(lane_result('bound' if route=='synthetic-unbound' else route))
                    return {'policy':copy.deepcopy(pre),'effective_record':{'path':'synthetic-pre-resolved'},
                            'result':result,'interlock':{'expected_interlock_failure':route=='synthetic-unbound'}}
                config={'paths':{'STUDY':str(root),'OUT':str(root)}}
                with patch.object(fixtures,'git',side_effect=real_git if repaired else without_origin), \
                     patch.object(controller,'run_revision',side_effect=fake_run):
                    evidence=controller.execute_fixture(config,{'git':{'path':executable}},
                                                        'synthetic',None,root)
                self.assertEqual(len(calls),12)
                execution=[row for row in calls if not row['resolve_only']]
                for route in ('bound','reference'):
                    pair=[row for row in execution if row['route']==route]
                    self.assertEqual(len(pair),2)
                    self.assertFalse(pair[0]['marker']);self.assertTrue(pair[1]['marker'])
                    shared=pair[0]['origin'] is not None and pair[0]['origin']==pair[1]['origin']
                    self.assertEqual(shared,repaired)
                    if not repaired:
                        self.assertEqual([row['origin'] for row in pair],[None,None])
                if repaired:
                    expected=str(Path(evidence['fixture_manifest']['database']).resolve())
                    self.assertTrue(expected)
                    self.assertTrue(all(row['origin']==expected for row in calls))

    def test_invalid_reference_head_resolution_prevents_every_scanner(self):
        import acceptance as controller
        pre=DeterministicProofTests().evidence()['runs']['bound']['parent']['policy']
        calls=[]
        def resolve(root,revision,route,run,runtime,frozen,closure,resolve_only=False,**kwargs):
            calls.append((route,revision,resolve_only))
            record=copy.deepcopy(pre)
            if route=='reference' and revision=='head':
                record['model_proof']['jev_disabled']=False
            return {'policy':record}
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory);config={'paths':{'STUDY':str(root),'OUT':str(root)}}
            with patch.object(controller,'fixture_manifest',return_value=({'database':'db','parent':'parent','head':'head'},None)), \
                 patch.object(controller,'materialize',return_value=[]), \
                 patch.object(controller,'run_revision',side_effect=resolve):
                with self.assertRaises(InterlockFailure):
                    controller.execute_fixture(config,{'git':{'path':'git'}},'synthetic',None,root)
        self.assertEqual(calls,[('bound','parent',True),('bound','head',True),
                                ('reference','parent',True),('reference','head',True)])

    def test_resolved_policy_must_match_later_bytes_and_pin(self):
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory);expected=root/'before.json';actual=root/'scanner.json'
            expected.write_bytes(b'{"model_lanes": []}\n');actual.write_bytes(expected.read_bytes())
            frozen=pin(expected)
            self.assertTrue(verify_effective_record(actual,frozen))
            actual.write_bytes(b'{"model_lanes":[]}\n')
            with self.assertRaises(InterlockFailure):verify_effective_record(actual,frozen)
            expected.write_bytes(actual.read_bytes())
            with self.assertRaises(ValueError):verify_effective_record(actual,frozen)

    def test_calledprocesserror_always_terminates_refused_receipt(self):
        with tempfile.TemporaryDirectory() as directory:
            target=Path(directory)/'synthetic.json';journal=Mock()
            receipt={'status':'refused','acceptance':{},'failures':[]}
            def fail():raise subprocess.CalledProcessError(7,['git','synthetic'])
            with target.open('x') as stream:
                self.assertEqual(complete_operation(stream,target,journal,receipt,fail),1)
            saved=json.loads(target.read_bytes())
            self.assertEqual(saved['status'],'refused');self.assertIn('CalledProcessError',saved['failures'][0])
            journal.finish.assert_called_once()
            self.assertEqual(journal.finish.call_args.args[0],'refused')
            with self.assertRaises(FileExistsError):target.open('x')

    def test_bound_policy_refuses_each_input_boundary(self):
        policy={'profile':'report','decisions':{'enabled':False},'model_cli':{'lanes':[]},'mapping':{'association':{'enabled':False}}}
        good={'key_present':False,'association_nil':True,'measures_nil':True,'clawpatch_egress':False,
              'contextual_egress':False,'model_lanes':[],'effective_policy':policy}
        self.assertTrue(verify_bound_policy(good))
        for key,value in (('key_present',True),('association_nil',False),('measures_nil',False),('clawpatch_egress',True),('contextual_egress',True),('model_lanes',['contextual'])):
            with self.subTest(key=key):
                with self.assertRaises(InterlockFailure):verify_bound_policy(dict(good,**{key:value}))

if __name__=='__main__':unittest.main()

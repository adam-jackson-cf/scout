package main

import (
 "context"
 "encoding/json"
 "errors"
 "flag"
 "os"
 "os/exec"
 "path/filepath"
 "reflect"
 "strings"
 "testing"

 "scout/internal/analyzers"
 "scout/internal/association"
 "scout/internal/contextual"
 "scout/internal/clawpatch"
 "scout/internal/domain"
 "scout/internal/engine"
 "scout/internal/inventory"
 "scout/internal/policy"
)

const canonicalKey = `{"family":"security","category":"unsafe","rule_version":"1","anchor":{"path":"fixture.py","symbol":"pkg.f","ruleKey":"scanner:rule","mechanism":"hello","fingerprint":"2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},"path":"fixture.py","start_line":2,"start_byte":4}`

func observation() domain.Observation {
 return domain.Observation{ObservationID:"obs-1",Family:domain.FamilySecurity,OriginalCategory:"unsafe",Source:"scanner",RuleID:"rule",RuleVersion:"1",Mechanism:" Hello ",Evidence:[]domain.Evidence{{Location:domain.SourceRange{Path:"fixture.py",Symbol:"pkg.f",StartLine:2,EndLine:2,StartByte:4,EndByte:8}}}}
}
func parsePolicy(t *testing.T, text string) policy.Policy {
 t.Helper();p,e:=policy.Parse([]byte(text));if e!=nil{t.Fatal(e)};return p
}

func TestFrozenDigestAndModelFreeBoundaries(t *testing.T) {
 valid:=[]byte("{\n  \"profile\": \"report\",\n  \"decisions\": {\n    \"enabled\": false\n  },\n  \"model_cli\": {\n    \"lanes\": []\n  },\n  \"mapping\": {\n    \"association\": {\n      \"enabled\": false\n    }\n  }\n}\n")
 for _,tc:=range []struct{name string;data []byte;digest string;wantError bool}{
  {"signed bytes",valid,"87de394937a225f87f1d53fed00c09bff88372a5de17b6669656478a32245f9f",false},
  {"wrong expected digest",valid,strings.Repeat("0",64),true},
  {"changed bytes",append(append([]byte{},valid...), ' '),"87de394937a225f87f1d53fed00c09bff88372a5de17b6669656478a32245f9f",true},
 }{t.Run(tc.name,func(t *testing.T){_,e:=frozenPolicy(tc.data,tc.digest);if (e!=nil)!=tc.wantError{t.Fatalf("error = %v",e)}})}
 for _,tc:=range []struct{name,policy string;wantError bool}{
  {"disabled",`{"profile":"report","decisions":{"enabled":false},"model_cli":{"lanes":[]},"mapping":{"association":{"enabled":false}}}`,false},
  {"decisions enabled",`{"profile":"report","decisions":{"enabled":true,"model":"typesafe/jev-1.14-20261014","egress_acknowledged":true},"model_cli":{"lanes":[]},"mapping":{"association":{"enabled":false}}}`,true},
  {"contextual enabled",`{"profile":"report","decisions":{"enabled":false},"model_cli":{"provider":"claude","lanes":["contextual"]},"mapping":{"association":{"enabled":false}}}`,true},
  {"review enabled",`{"profile":"report","decisions":{"enabled":false},"model_cli":{"provider":"claude","lanes":["clawpatch-review"]},"mapping":{"association":{"enabled":false}}}`,true},
  {"association enabled",`{"profile":"report","decisions":{"enabled":false},"model_cli":{"lanes":[]},"mapping":{"association":{"enabled":true}}}`,true},
  {"wrong profile",`{"profile":"guardian","decisions":{"enabled":false},"model_cli":{"lanes":[]},"mapping":{"association":{"enabled":false}}}`,true},
 }{t.Run(tc.name,func(t *testing.T){p:=parsePolicy(t,tc.policy);if e:=modelFree(p);(e!=nil)!=tc.wantError{t.Fatalf("error = %v",e)}})}
}

func TestKeyPresenceAndControlBoundaries(t *testing.T) {
 for _,tc:=range []struct{present,wantError bool}{{false,false},{true,true}}{if e:=requireNoKey(tc.present);(e!=nil)!=tc.wantError{t.Fatalf("presence %v: %v",tc.present,e)}}
 for _,tc:=range []struct{name,authorization,policy string;wantError bool}{
  {"authorized enabled","1",`{"decisions":{"enabled":true,"model":"typesafe/jev-1.14-20261014","egress_acknowledged":true},"model_cli":{"provider":"claude","lanes":["contextual"]}}`,false},
  {"absent authorization","",`{"decisions":{"enabled":true,"model":"typesafe/jev-1.14-20261014","egress_acknowledged":true},"model_cli":{"provider":"claude","lanes":["contextual"]}}`,true},
  {"nonexact authorization","true",`{"decisions":{"enabled":true,"model":"typesafe/jev-1.14-20261014","egress_acknowledged":true},"model_cli":{"provider":"claude","lanes":["contextual"]}}`,true},
  {"decisions disabled","1",`{"decisions":{"enabled":false},"model_cli":{"provider":"claude","lanes":["contextual"]}}`,true},
  {"lane absent","1",`{"decisions":{"enabled":true,"model":"typesafe/jev-1.14-20261014","egress_acknowledged":true},"model_cli":{"lanes":[]}}`,true},
  {"review alone","1",`{"decisions":{"enabled":true,"model":"typesafe/jev-1.14-20261014","egress_acknowledged":true},"model_cli":{"provider":"claude","lanes":["clawpatch-review"]}}`,true},
 }{t.Run(tc.name,func(t *testing.T){e:=controlPolicy(parsePolicy(t,tc.policy),tc.authorization);if (e!=nil)!=tc.wantError{t.Fatalf("error = %v",e)}})}
}

func TestPresentKeyRefusesBeforePolicyAccess(t *testing.T) {
 oldFlags,oldArgs:=flag.CommandLine,os.Args;t.Cleanup(func(){flag.CommandLine=oldFlags;os.Args=oldArgs})
 t.Setenv("OPENROUTER_API_KEY","")
 flag.CommandLine=flag.NewFlagSet("fixture",flag.ContinueOnError)
 os.Args=[]string{"frontend","--root","/absent/root","--revision","HEAD","--clawpatch","/absent/mapper","--effective-output",filepath.Join(t.TempDir(),"pre.json"),"--resolve-only"}
 if e:=run();e==nil||e.Error()!="key present"{t.Fatalf("expected key-present refusal, got %v",e)}
}

func TestInvalidRouteRefuses(t *testing.T) {
 if _,_,e:=resolve("unknown","",policy.Default(),engine.Config{},"");e==nil||e.Error()!="invalid route"{t.Fatalf("invalid route error %v",e)}
}

func TestOccurrenceKeyIndependentExpectedValues(t *testing.T) {
 for _,tc:=range []struct{name string;change func(*domain.Observation);expected string}{
  {"literal full tuple",func(o *domain.Observation){},canonicalKey},
  {"category",func(o *domain.Observation){o.OriginalCategory="other"},strings.Replace(canonicalKey,`"category":"unsafe"`,`"category":"other"`,1)},
  {"line",func(o *domain.Observation){o.Evidence[0].Location.StartLine=3},strings.Replace(canonicalKey,`"start_line":2`,`"start_line":3`,1)},
  {"byte",func(o *domain.Observation){o.Evidence[0].Location.StartByte=5},strings.Replace(canonicalKey,`"start_byte":4`,`"start_byte":5`,1)},
  {"rule version",func(o *domain.Observation){o.RuleVersion="2"},strings.Replace(canonicalKey,`"rule_version":"1"`,`"rule_version":"2"`,1)},
  {"family",func(o *domain.Observation){o.Family=domain.FamilyQuality},strings.Replace(canonicalKey,`"family":"security"`,`"family":"quality"`,1)},
  {"symbol",func(o *domain.Observation){o.Evidence[0].Location.Symbol="pkg.g"},strings.Replace(canonicalKey,`"symbol":"pkg.f"`,`"symbol":"pkg.g"`,1)},
  {"path",func(o *domain.Observation){o.Evidence[0].Location.Path="other.py"},strings.ReplaceAll(canonicalKey,`fixture.py`,`other.py`)},
  {"rule",func(o *domain.Observation){o.RuleID="rule2"},strings.Replace(canonicalKey,`scanner:rule`,`scanner:rule2`,1)},
  {"cosmetic text",func(o *domain.Observation){o.Mechanism="HELLO"},canonicalKey},
 }{t.Run(tc.name,func(t *testing.T){o:=observation();tc.change(&o);got,e:=occurrenceKey(o);if e!=nil||got!=tc.expected{t.Fatalf("key %s, error %v; want %s",got,e,tc.expected)}})}
 moved:=observation();moved.Evidence[0].Location.StartLine=3
 movedKey:=strings.Replace(canonicalKey,`"start_line":2`,`"start_line":3`,1)
 got,counts,e:=keys([]domain.Observation{observation(),moved,observation()})
 if e!=nil||!reflect.DeepEqual(got,[]string{canonicalKey,movedKey})||!reflect.DeepEqual(counts,map[string]int{canonicalKey:2,movedKey:1}){t.Fatalf("keys %v counts %v error %v",got,counts,e)}
 for _,tc:=range []struct{name string;o domain.Observation}{{"no evidence",domain.Observation{}},{"empty category",domain.Observation{Evidence:observation().Evidence}}}{t.Run(tc.name,func(t *testing.T){if _,_,e:=keys([]domain.Observation{tc.o});e==nil{t.Fatal("invalid key accepted")}})}
}

func TestDurableLinkBoundaries(t *testing.T) {
 o:=observation();moved:=observation();moved.ObservationID="obs-2";moved.Evidence[0].Location.StartLine=3
 movedKey:=strings.Replace(canonicalKey,`"start_line":2`,`"start_line":3`,1)
 for _,tc:=range []struct{name string;observations []domain.Observation;linked map[string][]domain.Observation;prior map[string]bool;want map[string]string;carried []string;wantError bool}{
  {"fresh bijection",[]domain.Observation{o},map[string][]domain.Observation{"issue-1":{o}},nil,map[string]string{canonicalKey:"issue-1"},[]string{},false},
  {"duplicate multiplicity same issue",[]domain.Observation{o,o},map[string][]domain.Observation{"issue-1":{o}},nil,map[string]string{canonicalKey:"issue-1"},[]string{},false},
  {"head carried ID",[]domain.Observation{moved},map[string][]domain.Observation{"issue-1":{o,moved}},map[string]bool{"issue-1":true},map[string]string{movedKey:"issue-1"},[]string{"issue-1"},false},
  {"unlinked observation",[]domain.Observation{o},map[string][]domain.Observation{},nil,nil,nil,true},
  {"multiply linked observation",[]domain.Observation{o},map[string][]domain.Observation{"a":{o},"b":{o}},nil,nil,nil,true},
  {"fresh extra durable issue",[]domain.Observation{o},map[string][]domain.Observation{"a":{o},"b":{}},nil,nil,nil,true},
  {"fresh two occurrences one issue",[]domain.Observation{o,moved},map[string][]domain.Observation{"a":{o,moved}},nil,nil,nil,true},
 }{t.Run(tc.name,func(t *testing.T){got,carried,e:=durableLinks(tc.observations,tc.linked,tc.prior);if (e!=nil)!=tc.wantError{t.Fatalf("error = %v",e)};if e==nil&&(!reflect.DeepEqual(got,tc.want)||!reflect.DeepEqual(carried,tc.carried)){t.Fatalf("links %v carried %v",got,carried)}})}
}

func healthReceipt() (receipt,[]domain.CoverageKey,[]analyzers.Spec) {
 key:=domain.CoverageKey{FilePath:"a.py",Family:domain.FamilySecurity,Category:"unsafe"}
 return receipt{Mapping:clawpatch.MapResult{Status:clawpatch.MapMapped},Adapters:[]domain.AdapterResult{{Adapter:"scanner",Status:domain.AdapterCompleted}},Coverage:[]domain.Coverage{{Key:key,Status:domain.CoverageInspected}}},[]domain.CoverageKey{key},[]analyzers.Spec{{ID:"scanner"}}
}
func TestSharedScannerHealthBoundaries(t *testing.T) {
 for _,tc:=range []struct{status domain.CoverageStatus;want bool}{
  {domain.CoverageInspected,true},{domain.CoverageReused,true},{domain.CoverageNotApplicable,true},{domain.CoverageCovered,true},{domain.CoverageFlagged,true},
  {domain.CoverageSkipped,false},{domain.CoverageBlocked,false},{domain.CoverageIncomplete,false},{domain.CoverageRequiresManualVerification,false},{domain.CoverageStatus("unknown"),false},
 }{t.Run(string(tc.status),func(t *testing.T){r,expected,specs:=healthReceipt();r.Coverage[0].Status=tc.status;got,accounted:=scannerHealth(r,expected,specs,nil,nil,nil);wantKeys:=[]string{};if tc.want{wantKeys=[]string{"a.py\x00\x00security\x00unsafe"}};if got!=tc.want||!reflect.DeepEqual(accounted,wantKeys){t.Fatalf("health %v keys %q",got,accounted)}})}
 for _,tc:=range []struct{name string;change func(*receipt);failures []domain.Failure;diagnostics []inventory.Diagnostic;runErr error;want bool}{
  {name:"missing adapter",change:func(r *receipt){r.Adapters=nil},want:false},
  {name:"adapter failure",change:func(r *receipt){r.Adapters[0].Failures=[]domain.Failure{{Code:"bad"}}},want:false},
  {name:"missing expected row",change:func(r *receipt){r.Coverage=nil},want:false},
  {name:"duplicate row",change:func(r *receipt){r.Coverage=append(r.Coverage,r.Coverage[0])},want:false},
  {name:"map blocked",change:func(r *receipt){r.Mapping.Status=clawpatch.MapBlocked},want:false},
  {name:"map absent",change:func(r *receipt){r.Mapping.Status=""},want:false},
  {name:"map reused",change:func(r *receipt){r.Mapping.Status=clawpatch.MapReused},want:true},
  {name:"partial map",change:func(r *receipt){r.Mapping.ExplorationIncomplete=true},want:false},
  {name:"omissions and gaps recorded",change:func(r *receipt){r.Omissions=[]domain.Omission{{Reason:"not mapped"}};r.MappingGaps=[]inventory.MappingGap{{Path:"a.py",Reason:"not mapped"}}},want:true},
  {name:"model rows do not change scanner health",change:func(r *receipt){r.Coverage=append(r.Coverage,domain.Coverage{Key:domain.CoverageKey{FeatureID:"f"},Status:domain.CoverageSkipped})},want:true},
  {name:"inventory failure",failures:[]domain.Failure{{Stage:domain.StageInventory}},want:false},
  {name:"source drift",failures:[]domain.Failure{{Stage:domain.StageInventory,Code:"source_drift"}},want:false},
  {name:"scanner failure",failures:[]domain.Failure{{Stage:domain.StageFileAnalysis}},want:false},
  {name:"mapping failure",failures:[]domain.Failure{{Stage:domain.StageMapping}},want:false},
  {name:"positive control contextual failure is not scanner failure",failures:[]domain.Failure{{Stage:domain.StageFeatureAnalysis}},want:true},
  {name:"inventory diagnostic",diagnostics:[]inventory.Diagnostic{{Code:"unreadable"}},want:false},
  {name:"execution error",runErr:errors.New("failed"),want:false},
 }{t.Run(tc.name,func(t *testing.T){for _,route:=range []string{"bound","reference"}{r,expected,specs:=healthReceipt();r.Route=route;if tc.change!=nil{tc.change(&r)};got,_:=scannerHealth(r,expected,specs,tc.failures,tc.diagnostics,tc.runErr);if got!=tc.want{t.Fatalf("%s health %v want %v",route,got,tc.want)}}})}
 for _,tc:=range []struct{status domain.AdapterStatus;want bool}{{domain.AdapterCompleted,true},{domain.AdapterNoApplicableInput,true},{domain.AdapterMissingTool,false},{domain.AdapterFailed,false},{domain.AdapterPartial,false},{domain.AdapterTimedOut,false},{domain.AdapterIneligible,true}}{t.Run(string(tc.status),func(t *testing.T){r,expected,specs:=healthReceipt();r.Adapters[0].Status=tc.status;if tc.status==domain.AdapterIneligible{r.Coverage[0].Status=domain.CoverageNotApplicable;r.Adapters[0].Coverage=append([]domain.Coverage{},r.Coverage...)};got,_:=scannerHealth(r,expected,specs,nil,nil,nil);if got!=tc.want{t.Fatalf("health %v",got)}})}
}

func TestBindingRowsBoundaries(t *testing.T) {
 key:=domain.CoverageKey{FeatureID:"feature",Family:domain.FamilySecurity,Category:"sensitive_data_flows"}
 for _,tc:=range []struct{name string;rows []domain.Coverage;jobs []domain.Job;want bool}{
  {"no features",nil,nil,true},
  {"skipped contextual",[]domain.Coverage{{Key:key,Status:domain.CoverageSkipped,Reason:"skipped (not opted in)"}},[]domain.Job{{Stage:domain.StageFeatureAnalysis,CoverageKey:&key}},true},
  {"skipped review",[]domain.Coverage{{Key:key,Status:domain.CoverageSkipped,Reason:"skipped (not opted in)"}},[]domain.Job{{Stage:domain.StageClawpatchReview,CoverageKey:&key}},true},
  {"missing planned row",nil,[]domain.Job{{Stage:domain.StageFeatureAnalysis,CoverageKey:&key}},false},
  {"missing review row",nil,[]domain.Job{{Stage:domain.StageClawpatchReview,CoverageKey:&key}},false},
  {"wrong reason",[]domain.Coverage{{Key:key,Status:domain.CoverageSkipped,Reason:"no CLI"}},nil,false},
  {"model ran",[]domain.Coverage{{Key:key,Status:domain.CoverageInspected,Reason:"skipped (not opted in)"}},nil,false},
  {"malformed job",nil,[]domain.Job{{Stage:domain.StageClawpatchReview}},false},
 }{t.Run(tc.name,func(t *testing.T){if got:=bindingRowsComplete(tc.rows,tc.jobs);got!=tc.want{t.Fatalf("binding rows %v",got)}})}
}

func TestPlannerInputsAndScannerManifest(t *testing.T) {
 root:=t.TempDir()
 for _,path:=range []string{"a.py","b.js","c.py"}{if e:=os.WriteFile(filepath.Join(root,path),[]byte("pass\n"),0600);e!=nil{t.Fatal(e)}}
 files:=[]domain.File{
  {Path:"c.py",Eligible:true,ContentDigest:"digest",Language:"python"},
  {Path:"excluded.py",Eligible:true,Excluded:true,ContentDigest:"digest",Language:"python"},
  {Path:"unreadable.py",Eligible:true,ContentDigest:"",Language:"python"},
  {Path:"ineligible.py",Eligible:false,ContentDigest:"digest",Language:"python"},
  {Path:"b.js",Eligible:true,ContentDigest:"digest",Language:"javascript"},
  {Path:"a.py",Eligible:true,ContentDigest:"digest",Language:"python"},
 }
 planned:=plannedFiles(files);paths:=[]string{};for _,file:=range planned{paths=append(paths,file.Path)}
 if !reflect.DeepEqual(paths,[]string{"a.py","b.js","c.py"}){t.Fatalf("planned %v",paths)}
 spec:=analyzers.Spec{ID:"python",Family:domain.FamilySecurity,Category:"unsafe",Eligibility:analyzers.Eligibility{Languages:[]string{"python"}},Inputs:analyzers.InputLimits{MaxFiles:2,MaxFileBytes:1024,MaxTotalBytes:4096,MaxManifestBytes:65536}}
 coverage:=[]domain.Coverage{}
 for _,row:=range []struct{path string;status domain.CoverageStatus}{{"a.py",domain.CoverageInspected},{"b.js",domain.CoverageNotApplicable},{"c.py",domain.CoverageFlagged}}{coverage=append(coverage,domain.Coverage{Key:domain.CoverageKey{FilePath:row.path,Family:domain.FamilySecurity,Category:"unsafe"},Status:row.status})}
 got:=scannerInputs(root,files,[]analyzers.Spec{spec},coverage)
 // b.js remains in the planner batch; filtering first would put a.py and c.py
 // together and contradict Assess's actual invocation boundaries.
 want:=map[string][][]string{"python":{{"a.py"},{"c.py"}}}
 if !reflect.DeepEqual(got,want){t.Fatalf("manifest %v want %v",got,want)}
 if e:=os.Remove(filepath.Join(root,"c.py"));e!=nil{t.Fatal(e)}
 want=map[string][][]string{"python":{{"a.py"},{}}}
 if got=scannerInputs(root,files,[]analyzers.Spec{spec},coverage);!reflect.DeepEqual(got,want){t.Fatalf("missing input manifest %v",got)}
 coverage[0].Status=domain.CoverageNotApplicable
 want=map[string][][]string{"python":{{},{}}}
 if got=scannerInputs(root,files,[]analyzers.Spec{spec},coverage);!reflect.DeepEqual(got,want){t.Fatalf("allowlisted input manifest %v",got)}
}

func TestResolveOnlyRequiredArguments(t *testing.T) {
 for _,tc:=range []struct{name,root,revision,mapper,effective,output string;resolve,wantError bool}{
  {"resolve without output","root","rev","mapper","effective","",true,false},
  {"scan needs output","root","rev","mapper","effective","",false,true},
  {"resolve needs effective","root","rev","mapper","","",true,true},
  {"resolve needs root","","rev","mapper","effective","",true,true},
  {"resolve needs revision","root","","mapper","effective","",true,true},
  {"resolve needs mapper","root","rev","","effective","",true,true},
  {"scan complete","root","rev","mapper","effective","output",false,false},
 }{t.Run(tc.name,func(t *testing.T){e:=requiredArguments(tc.root,tc.revision,tc.mapper,tc.effective,tc.output,tc.resolve);if (e!=nil)!=tc.wantError{t.Fatalf("error %v",e)}})}
}

func TestResolveOnlyWritesExclusiveMatchingRecordWithoutScanner(t *testing.T) {
 root:=t.TempDir()
 for _,args:=range [][]string{{"init","--quiet"},{"-c","user.name=Fixture","-c","user.email=fixture@example.invalid","commit","--quiet","--allow-empty","-m","fixture"}}{cmd:=exec.Command("git",args...);cmd.Dir=root;if out,e:=cmd.CombinedOutput();e!=nil{t.Fatalf("git fixture: %v: %s",e,out)}}
 frozen,e:=filepath.Abs("../../preparation/model-free-policy.json");if e!=nil{t.Fatal(e)}
 oldFlags,oldArgs:=flag.CommandLine,os.Args;t.Cleanup(func(){flag.CommandLine=oldFlags;os.Args=oldArgs})
 // An empty-but-present key is prohibited too. Never read or print any value.
 t.Setenv("OPENROUTER_API_KEY","")
 if e=os.Unsetenv("OPENROUTER_API_KEY");e!=nil{t.Fatal(e)}
 for _,route:=range []string{"bound","reference"}{t.Run(route,func(t *testing.T){
  effective:=filepath.Join(t.TempDir(),"effective.json")
  invoke:=func()error{flag.CommandLine=flag.NewFlagSet("fixture",flag.ContinueOnError);os.Args=[]string{"frontend","--root",root,"--revision","HEAD","--route",route,"--policy",frozen,"--clawpatch","/absent/mapper","--effective-output",effective,"--resolve-only"};return run()}
  if e:=invoke();e!=nil{t.Fatal(e)}
  got,e:=os.ReadFile(effective);if e!=nil{t.Fatal(e)}
  original,e:=policy.Load(root,"HEAD");if e!=nil{t.Fatal(e)}
  cfg,value,e:=resolve(route,frozen,original,engine.Config{Root:root,ClawpatchBinary:"/absent/mapper"},"");if e!=nil{t.Fatal(e)}
  cfg.Registry,e=analyzers.BuiltinRegistry();if e!=nil{t.Fatal(e)}
  scope:=domain.Scope{Kind:domain.ScopeRepository,SelectedFamilies:engine.DefaultFamilies()}
  later:=filepath.Join(t.TempDir(),"later.json")
  if e=writeJSON(later,effectiveRecord(route,"HEAD",original,value,cfg,engine.Options{Scope:scope,ForceFull:true,Trigger:domain.TriggerPush}));e!=nil{t.Fatal(e)}
  same,e:=os.ReadFile(later);if e!=nil{t.Fatal(e)}
  if string(got)!=string(same){t.Fatal("resolve-only bytes differ from scanner effective record")}
  if e=invoke();e==nil{t.Fatal("effective record overwritten")}
  var pre map[string]any;if e=json.Unmarshal(got,&pre);e!=nil{t.Fatal(e)}
  if pre["key_present"]!=false||pre["association_nil"]!=true||pre["ranking_stage"]!="absent"{t.Fatalf("unexpected pre-record keys %v",pre)}
 })}
 if _,e=os.Stat(filepath.Join(root,".scout"));!os.IsNotExist(e){t.Fatal("resolve-only accessed persistent assessment state")}
}

func TestFrozenPolicyWiresDisabledLanes(t *testing.T) {
 p:=parsePolicy(t,`{"profile":"report","decisions":{"enabled":false},"model_cli":{"lanes":[]},"mapping":{"association":{"enabled":false}}}`)
 cfg:=engine.ApplyModelCLIPolicy(engine.Config{ModelLanes:[]string{"contextual"}},p)
 if len(cfg.ModelLanes)!=0||p.Decisions().Enabled||p.Mapping().Association.Enabled{t.Fatal("binding did not disable model inputs")}
}

func TestSecretOccurrenceKeyIndependentExpectedValue(t *testing.T) {
 o:=observation();o.Mechanism="secret diagnostic reported";o.SourceFindingID="abc123"
 want:=`{"family":"security","category":"unsafe","rule_version":"1","anchor":{"path":"fixture.py","symbol":"pkg.f","ruleKey":"scanner:rule","mechanism":"secret diagnostic reported","fingerprint":"abc123","secretDigest":true},"path":"fixture.py","start_line":2,"start_byte":4}`
 if got,e:=occurrenceKey(o);e!=nil||got!=want{t.Fatalf("secret key %s error %v",got,e)}
}

func TestOneOccurrenceCannotLinkToTwoIssues(t *testing.T) {
 first,second:=observation(),observation();second.ObservationID="obs-2"
 if _,_,e:=durableLinks([]domain.Observation{first,second},map[string][]domain.Observation{"first":{first},"second":{second}},map[string]bool{"first":true});e==nil{t.Fatal("one occurrence mapped to two durable IDs")}
}

func TestDuplicateBindingRowsRefuse(t *testing.T) {
 row:=domain.Coverage{Key:domain.CoverageKey{FeatureID:"f",Family:domain.FamilyClawpatch,Category:"review"},Status:domain.CoverageSkipped,Reason:"skipped (not opted in)"}
 if bindingRowsComplete([]domain.Coverage{row,row},nil){t.Fatal("duplicate binding row accepted")}
}

func boundFixture() (engine.Config,engine.Options) {
 cfg:=engine.ApplyModelCLIPolicy(engine.Config{Root:"fixture",ClawpatchBinary:"mapper"},policy.Default())
 cfg.ModelLanes=nil
 return cfg,engine.Options{Scope:domain.Scope{Kind:domain.ScopeRepository,SelectedFamilies:engine.DefaultFamilies()},ForceFull:true,Trigger:domain.TriggerPush}
}
func TestBoundActualInputBoundaries(t *testing.T) {
 for _,tc:=range []struct{name string;change func(*engine.Config,*engine.Options)}{
  {"unknown lane",func(c *engine.Config,o *engine.Options){c.ModelLanes=[]string{"unknown"}}},
  {"contextual",func(c *engine.Config,o *engine.Options){c.ModelLanes=[]string{engine.LaneContextual}}},
  {"review",func(c *engine.Config,o *engine.Options){c.ModelLanes=[]string{engine.LaneClawpatchReview}}},
  {"fallback",func(c *engine.Config,o *engine.Options){c.ModelLanes=[]string{engine.LaneAssociationFallback}}},
  {"contextual egress",func(c *engine.Config,o *engine.Options){c.Provider.EgressAuthorized=true}},
  {"clawpatch egress",func(c *engine.Config,o *engine.Options){c.ClawpatchEgressAuthorized=true}},
  {"provider mismatch",func(c *engine.Config,o *engine.Options){c.ClawpatchProvider="claude"}},
  {"invalid provider",func(c *engine.Config,o *engine.Options){c.Provider.CLI="other";c.ClawpatchProvider="other"}},
  {"provider model",func(c *engine.Config,o *engine.Options){c.Provider.Model="unexpected"}},
  {"provider limits",func(c *engine.Config,o *engine.Options){c.Provider.MaxPromptBytes=1}},
  {"provider prompt",func(c *engine.Config,o *engine.Options){c.Provider.PromptVersion="unexpected"}},
  {"provider timeout",func(c *engine.Config,o *engine.Options){c.Provider.Timeout=1}},
  {"provider context limit",func(c *engine.Config,o *engine.Options){c.Provider.MaxContextBytes=1}},
  {"provider file limit",func(c *engine.Config,o *engine.Options){c.Provider.MaxFileBytes=1}},
  {"dependencies",func(c *engine.Config,o *engine.Options){c.Dependencies=[]inventory.DependencyEdge{{}}}},
  {"association",func(c *engine.Config,o *engine.Options){o.Association=&association.LaneOptions{}}},
  {"measures",func(c *engine.Config,o *engine.Options){o.Measures=&engine.MeasuresOptions{}}},
  {"inline history",func(c *engine.Config,o *engine.Options){o.Inline=[]engine.InlineLane{{Lane:engine.LaneHistory}}}},
  {"scope",func(c *engine.Config,o *engine.Options){o.Scope.SelectedFamilies=nil}},
  {"trigger",func(c *engine.Config,o *engine.Options){o.Trigger=domain.TriggerNightly}},
  {"force full",func(c *engine.Config,o *engine.Options){o.ForceFull=false}},
  {"exclusions",func(c *engine.Config,o *engine.Options){o.Exclusions=[]string{"*"}}},
  {"root",func(c *engine.Config,o *engine.Options){c.Root=""}},
  {"mapper",func(c *engine.Config,o *engine.Options){c.ClawpatchBinary=""}},
  {"guard",func(c *engine.Config,o *engine.Options){c.AssociationGuard.WindowMerges++}},
 }{t.Run(tc.name,func(t *testing.T){cfg,opts:=boundFixture();tc.change(&cfg,&opts);if boundInputs("bound",cfg,opts)==nil||completeProof(configProof("bound",cfg,opts)){t.Fatal("unsafe effective API inputs accepted")}})}
 cfg,opts:=boundFixture()
 if e:=boundInputs("bound",cfg,opts);e!=nil{t.Fatal(e)}
 if !completeProof(configProof("bound",cfg,opts)){t.Fatal("valid bound inputs not proved")}
 if completeProof(configProof("synthetic-unbound",cfg,opts)){t.Fatal("control claimed proof")}
 cfg.Provider=contextual.Config{CLI:"claude"};cfg.ClawpatchProvider="claude"
 if e:=boundInputs("bound",cfg,opts);e!=nil{t.Fatal(e)}
}

func disabledResult() engine.Result {
 return engine.Result{Lanes:[]domain.LaneHealth{
  {Lane:engine.LaneContextual,Status:domain.LaneSkipped,Reason:engine.ReasonNotOptedIn,Units:1},
  {Lane:engine.LaneClawpatchReview,Status:domain.LaneSkipped,Reason:engine.ReasonNotOptedIn,Units:1},
  {Lane:engine.LaneAssociation,Status:domain.LaneSkipped,Reason:"disabled"},
  {Lane:engine.LaneAssociationFallback,Status:domain.LaneSkipped,Reason:"disabled"},
 }}
}
func runtimeReceipt() receipt {
 return receipt{BindingRowsComplete:true,Mapping:clawpatch.MapResult{Status:clawpatch.MapMapped,Source:clawpatch.MapSourceHeuristic}}
}
func TestRuntimeProofRequiresActualDisabledHealth(t *testing.T) {
 cfg,opts:=boundFixture()
 for _,lane:=range []string{engine.LaneContextual,engine.LaneClawpatchReview,engine.LaneAssociation,engine.LaneAssociationFallback}{
  for _,status:=range []domain.LaneStatus{domain.LaneRan,domain.LaneFailed,domain.LaneBlocked,domain.LaneReused}{
   t.Run(lane+string(status),func(t *testing.T){result:=disabledResult();for i:=range result.Lanes{if result.Lanes[i].Lane==lane{result.Lanes[i].Status=status}}
    r:=runtimeReceipt();runtimeProof(&r,"bound",cfg,opts,&result);if r.ModelCallImpossible{t.Fatal("active model lane claimed impossible")}
   })
  }
 }
 result:=disabledResult();r:=runtimeReceipt();runtimeProof(&r,"bound",cfg,opts,&result)
 if !r.ModelCallImpossible||r.AssociationResult!="disabled"||r.JevResult!="absent"||!reflect.DeepEqual(r.LaneReport,result.Lanes)||&r.LaneReport[0]!=&result.Lanes[0]{t.Fatalf("disabled runtime proof %+v",r)}
 r=runtimeReceipt();r.BindingRowsComplete=false;runtimeProof(&r,"bound",cfg,opts,&result);if r.ModelCallImpossible{t.Fatal("missing skipped coverage accepted")}
 result.Lanes=nil;r=runtimeReceipt();runtimeProof(&r,"bound",cfg,opts,&result);if r.ModelCallImpossible{t.Fatal("missing association report accepted")}
 result=disabledResult();result.Lanes[0].Reason="no model CLI";r=runtimeReceipt();runtimeProof(&r,"bound",cfg,opts,&result);if r.ModelCallImpossible{t.Fatal("no CLI substituted for no opt-in")}
 result=disabledResult();r=runtimeReceipt();runtimeProof(&r,"synthetic-unbound",cfg,opts,&result);if r.ModelCallImpossible{t.Fatal("control claimed impossible")}
 cfg=engine.Config{Root:"fixture",ClawpatchBinary:"mapper"};r=runtimeReceipt();runtimeProof(&r,"reference",cfg,opts,nil)
 if !r.ModelCallImpossible||r.ModelStages!="structurally_absent"||r.AssociationResult!="structurally_absent"||r.LaneReport!=nil{t.Fatal("reference stages not structurally distinguished")}
}

func TestEffectiveRecordEnumeratesActualAPIFields(t *testing.T) {
 cfg,opts:=boundFixture();registry,e:=analyzers.BuiltinRegistry();if e!=nil{t.Fatal(e)};cfg.Registry=registry
 record:=effectiveRecord("bound","HEAD",policy.Default(),describe(policy.Default()),cfg,opts)
 config:=record["config"].(map[string]any)
 typ:=reflect.TypeOf(cfg);for i:=range typ.NumField(){if _,ok:=config[typ.Field(i).Name];!ok{t.Fatalf("Config field missing: %s",typ.Field(i).Name)}}
 encoded,e:=json.Marshal(record);if e!=nil{t.Fatal(e)}
 var decoded map[string]any;if e=json.Unmarshal(encoded,&decoded);e!=nil{t.Fatal(e)}
 options:=decoded["options"].(map[string]any);typ=reflect.TypeOf(opts)
 for i:=range typ.NumField(){if _,ok:=options[typ.Field(i).Name];!ok{t.Fatalf("Options field missing: %s",typ.Field(i).Name)}}
 identity:=config["Registry"].(map[string]any);specs:=identity["specs"].([]analyzers.Spec)
 if len(specs)==0||!reflect.DeepEqual(specs,registry.Specs())||len(identity["sha256"].(string))!=64{t.Fatal("registry identity absent")}
 if record["model_call_impossible"]!=true||record["history_stage"]!="absent"||record["ranking_stage"]!="absent"{t.Fatal("model-free API evidence absent")}
}

func TestRuntimeProofNoEligibleUnitsAndMissingResult(t *testing.T) {
 cfg,opts:=boundFixture();result:=disabledResult()
 for i:=range 2{result.Lanes[i].Units=0;result.Lanes[i].Reason="no eligible inputs"}
 r:=runtimeReceipt();runtimeProof(&r,"bound",cfg,opts,&result)
 if !r.ModelCallImpossible{t.Fatal("no eligible model units rejected")}
 r=runtimeReceipt();runtimeProof(&r,"bound",cfg,opts,nil)
 if r.ModelCallImpossible||completeProof(r.ModelProof){t.Fatal("missing assessment result claimed runtime proof")}
 if completeProof(map[string]bool{"effective_config_bound":true}){t.Fatal("incomplete proof accepted")}
 cfg=engine.Config{Root:"fixture",ClawpatchBinary:"mapper",Provider:contextual.Config{CLI:"claude"}}
 if boundInputs("reference",cfg,opts)==nil{t.Fatal("reference accepted a model provider")}
}

func TestRealIneligibleRunnerHealth(t *testing.T) {
 root:=t.TempDir()
 if e:=os.WriteFile(filepath.Join(root,"nonmatching.js"),[]byte("const local = 1;\n"),0600);e!=nil{t.Fatal(e)}
 registry,e:=analyzers.BuiltinRegistry();if e!=nil{t.Fatal(e)}
 const id="security.semgrep.python-unsafe-yaml-load"
 var spec analyzers.Spec
 for _,candidate:=range registry.Specs(){if candidate.ID==id{spec=candidate}}
 if spec.ID==""{t.Fatal("required synthetic no-applicable spec missing")}
 runner,e:=registry.Runner(id,"x1-ineligible-fixture");if e!=nil{t.Fatal(e)}
 // Real Runner returns before sandbox creation or scanner execution because
 // this JavaScript input cannot match this Python-only adapter.
 adapter,observations,e:=runner.Run(context.Background(),root,[]domain.File{{Path:"nonmatching.js",Language:"javascript",Eligible:true,ContentDigest:"fixture"}})
 if e!=nil||adapter.Status!=domain.AdapterIneligible||len(adapter.Failures)!=0||len(observations)!=0||len(adapter.Coverage)!=1||adapter.Coverage[0].Status!=domain.CoverageNotApplicable{t.Fatalf("unexpected no-applicable result: %+v, error %v",adapter,e)}
 for _,route:=range []string{"bound","reference"}{
  for _,tc:=range []struct{name string;status domain.CoverageStatus;failure bool;want bool}{
   {"nonapplicable",domain.CoverageNotApplicable,false,true},
   {"failure",domain.CoverageNotApplicable,true,false},
   {"incomplete",domain.CoverageIncomplete,false,false},
   {"footprint failure",domain.CoverageIncomplete,true,false},
   {"inspected",domain.CoverageInspected,false,false},
  }{t.Run(route+"/"+tc.name,func(t *testing.T){
   actual:=adapter;actual.Coverage=append([]domain.Coverage{},adapter.Coverage...);actual.Coverage[0].Status=tc.status
   if tc.failure{actual.Failures=[]domain.Failure{{Code:"project_footprint",Stage:domain.StageFileAnalysis}}}
   r:=receipt{Route:route,Mapping:clawpatch.MapResult{Status:clawpatch.MapMapped},Adapters:[]domain.AdapterResult{actual},Coverage:actual.Coverage}
   healthy,_:=scannerHealth(r,[]domain.CoverageKey{actual.Coverage[0].Key},[]analyzers.Spec{spec},nil,nil,nil)
   if healthy!=tc.want{t.Fatalf("healthy %v, want %v",healthy,tc.want)}
  })}
 }
}

func TestReferenceOffFileObservationFiltering(t *testing.T) {
 own:=observation();own.ObservationID="own"
 foreign:=observation();foreign.ObservationID="foreign";foreign.Evidence=append([]domain.Evidence{},foreign.Evidence...);foreign.Evidence[0].Location.Path="other.py"
 secondary:=foreign;secondary.ObservationID="secondary";secondary.Evidence=append(append([]domain.Evidence{},foreign.Evidence...),own.Evidence...)
 retained:=domain.Observation{ObservationID:"already retained"}
 got:=appendFileObservations([]domain.Observation{retained},[]domain.Observation{foreign,own,{ObservationID:"missing evidence"},secondary,own},"fixture.py")
 want:=[]domain.Observation{retained,own,own}
 if !reflect.DeepEqual(got,want){t.Fatalf("off-file split retained wrong observations: %+v",got)}
 r,expected,specs:=healthReceipt();r.Adapters[0].Status=domain.AdapterIneligible
 if healthy,_:=scannerHealth(r,expected,specs,nil,nil,nil);healthy{t.Fatal("ineligible adapter without coverage proof accepted")}
}

func TestActualLaneReportMissingDuplicateAndUnexpectedRowsRefuse(t *testing.T) {
 cfg,opts:=boundFixture()
 for index:=range 4{
  for _,duplicate:=range []bool{false,true}{
   result:=disabledResult()
   if duplicate{result.Lanes=append(result.Lanes,result.Lanes[index])}else{result.Lanes=append(result.Lanes[:index],result.Lanes[index+1:]...)}
   r:=runtimeReceipt();runtimeProof(&r,"bound",cfg,opts,&result)
   if r.ModelCallImpossible{t.Fatalf("accepted malformed actual lane %d, duplicate %v",index,duplicate)}
   if !reflect.DeepEqual(r.LaneReport,result.Lanes){t.Fatal("actual report reconstructed instead of retained")}
  }
 }
 for _,requirement:=range []domain.LaneRequires{domain.RequiresModelCLI,domain.RequiresDecisionEgress}{
  result:=disabledResult();result.Lanes=append(result.Lanes,domain.LaneHealth{Lane:"unexpected-model-stage",Requires:requirement,Status:domain.LaneSkipped,Reason:"disabled"})
  r:=runtimeReceipt();runtimeProof(&r,"bound",cfg,opts,&result)
  if r.ModelCallImpossible{t.Fatalf("unexpected model stage %s accepted",requirement)}
 }
}

func TestRuntimeProofRequiresActualHeuristicMapping(t *testing.T) {
 for _,route:=range []string{"bound","reference"}{
  for _,source:=range []clawpatch.MapSource{"",clawpatch.MapSourceAgent,clawpatch.MapSourceHeuristic}{
   cfg,opts:=boundFixture();result:=disabledResult()
   if route=="reference"{cfg=engine.Config{Root:"fixture",ClawpatchBinary:"mapper"}}
   r:=runtimeReceipt();r.Mapping.Source=source;runtimeProof(&r,route,cfg,opts,&result)
   want:=source==clawpatch.MapSourceHeuristic
   if r.ModelProof["heuristic_only"]!=want||r.ModelCallImpossible!=want{t.Fatalf("%s source %q proof %v",route,source,r.ModelProof)}
  }
 }
}

// modelFreeContractFixture exercises the emitted Go/Python boundary without
// assessment, scanners, subprocesses, environment reads, or a real checkout.
func modelFreeContractFixture() (map[string]any,error) {
 const root="/synthetic/x1-model-free-contract"
 const revision="contract-revision"
 original,e:=policy.Parse([]byte(`{"profile":"report","decisions":{"enabled":true,"model":"typesafe/jev-1.14-20261014","egress_acknowledged":true},"model_cli":{"provider":"claude","lanes":["contextual"]}}`));if e!=nil{return nil,e}
 registry,e:=analyzers.BuiltinRegistry();if e!=nil{return nil,e}
 effectiveRecords:=map[string]any{}
 receipts:=map[string]any{}
 for _,route:=range []string{"bound","reference","synthetic-unbound"}{
  cfg,opts:=boundFixture()
  cfg.Root=root;cfg.ClawpatchBinary=root+"/bin/clawpatch"
  if route!="bound"{cfg=engine.Config{Root:cfg.Root,ClawpatchBinary:cfg.ClawpatchBinary}}
  cfg,effective,e:=resolve(route,"../../preparation/model-free-policy.json",original,cfg,"1");if e!=nil{return nil,e}
  cfg.Registry=registry
  if e=boundInputs(route,cfg,opts);e!=nil{return nil,e}
  effectiveRecords[route]=effectiveRecord(route,revision,original,effective,cfg,opts)
  if route=="synthetic-unbound"{continue}
  r:=runtimeReceipt()
  r.Route=route;r.Revision=revision;r.OriginalPolicy=describe(original);r.Effective=effective
  var assessed *engine.Result
  if route=="bound"{result:=disabledResult();assessed=&result}
  runtimeProof(&r,route,cfg,opts,assessed)
  // Match run's final JSON normalization, retaining scanner health false:
  // this fixture deliberately has no scanner execution or scanner evidence.
  r.Keys,r.Multiplicity,e=keys(nil);if e!=nil{return nil,e}
  r.DurableLinks=map[string]string{};r.CarriedIDs=[]string{}
  r.Omissions=[]domain.Omission{};r.MappingGaps=[]inventory.MappingGap{}
  r.ScannerInputs=map[string][][]string{}
  receipts[route]=r
 }
 return map[string]any{"effective":effectiveRecords,"receipts":receipts},nil
}

func TestModelFreeCrossLanguageContractGolden(t *testing.T) {
 actual,e:=modelFreeContractFixture();if e!=nil{t.Fatal(e)}
 encoded,e:=json.Marshal(actual);if e!=nil{t.Fatal(e)}
 var normalized any
 if e=json.Unmarshal(encoded,&normalized);e!=nil{t.Fatal(e)}
 golden,e:=os.ReadFile("../tests/model_free_contract.json");if e!=nil{t.Fatal(e)}
 var expected any
 if e=json.Unmarshal(golden,&expected);e!=nil{t.Fatal(e)}
 if !reflect.DeepEqual(normalized,expected){t.Fatalf("model-free JSON contract differs from checked-in golden\nactual: %s",encoded)}
}

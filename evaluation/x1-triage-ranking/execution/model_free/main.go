// Experiment-local frontend. No policy is read from the Scout source checkout.
package main

import (
 "context"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "flag"
 "fmt"
 "os"
 "path/filepath"
 "reflect"
 "sort"

 "scout/internal/analyzers"
 "scout/internal/clawpatch"
 "scout/internal/contextual"
 "scout/internal/domain"
 "scout/internal/engine"
 "scout/internal/inventory"
 "scout/internal/issues"
 "scout/internal/policy"
 "scout/internal/state"
)

const frozenDigest = "87de394937a225f87f1d53fed00c09bff88372a5de17b6669656478a32245f9f"

// Include the complete correlation occurrence tuple, the full canonical anchor,
// and rule version. Multiplicity is retained separately; no ULIDs are remapped.
type findingKey struct {
 Family domain.Family `json:"family"`
 Category string `json:"category"`
 RuleVersion string `json:"rule_version"`
 Anchor domain.IssueAnchor `json:"anchor"`
 Path string `json:"path"`
 StartLine int `json:"start_line"`
 StartByte int `json:"start_byte"`
}

type receipt struct {
 Route string `json:"route"`
 Revision string `json:"revision"`
 OriginalPolicy policyRecord `json:"original_policy"`
 Effective any `json:"effective"`
 Keys []string `json:"issue_ids"`
 Multiplicity map[string]int `json:"multiplicity"`
 DurableCount int `json:"durable_issue_count"`
 DurableLinks map[string]string `json:"durable_links"`
 CarriedIDs []string `json:"carried_issue_ids"`
 Adapters []domain.AdapterResult `json:"adapters"`
 Coverage []domain.Coverage `json:"coverage"`
 AccountedCoverage []string `json:"accounted_coverage"`
 ScannerInputs map[string][][]string `json:"scanner_inputs"`
 Omissions []domain.Omission `json:"omissions"`
 MappingGaps []inventory.MappingGap `json:"mapping_gaps"`
 Mapping clawpatch.MapResult `json:"mapping"`
 Healthy bool `json:"scanner_health_complete"`
 BindingRowsComplete bool `json:"binding_rows_complete"`
 ModelProof map[string]bool `json:"model_proof"`
 LaneReport []domain.LaneHealth `json:"lane_report"`
 ModelCallImpossible bool `json:"model_call_impossible"`
 ModelStages string `json:"model_stages"`
 AssociationResult string `json:"association_result"`
 AssociationHealth []domain.LaneHealth `json:"association_health"`
 JevResult string `json:"jev_result"`
 Error string `json:"error,omitempty"`
}

type policyRecord struct {
 Profile string `json:"profile"`
 Decisions map[string]any `json:"decisions"`
 ModelCLI map[string]any `json:"model_cli"`
 Mapping map[string]any `json:"mapping"`
}
func describe(p policy.Policy) policyRecord {
 lanes:=p.ModelCLI().Lanes;if lanes==nil{lanes=[]string{}}
 return policyRecord{p.Profile(),
  map[string]any{"enabled":p.Decisions().Enabled,"resolved":p.Decisions()},
  map[string]any{"provider":p.ModelCLI().Provider,"lanes":lanes},
  map[string]any{"association":map[string]any{"enabled":p.Mapping().Association.Enabled},"resolved":p.Mapping()}}
}
func writeJSON(path string, v any) error {
 f,e:=os.OpenFile(path,os.O_WRONLY|os.O_CREATE|os.O_EXCL,0600); if e!=nil{return e}; defer f.Close()
 if e=json.NewEncoder(f).Encode(v);e!=nil{return e}; return f.Sync()
}

func occurrenceKey(o domain.Observation) (string,error) {
 a,ok:=domain.AnchorOf(o);if !ok{return "",errors.New("observation lacks canonical anchor")}
 category,e:=issues.SanitizeText(o.OriginalCategory);if e!=nil||category==""{return "",errors.New("observation lacks safe category")}
 location:=o.Evidence[0].Location
 b,e:=json.Marshal(findingKey{o.Family,category,o.RuleVersion,a,location.Path,location.StartLine,location.StartByte})
 return string(b),e
}
func keys(observations []domain.Observation) ([]string,map[string]int,error) {
 counts:=map[string]int{}
 for _,o:=range observations {k,e:=occurrenceKey(o);if e!=nil{return nil,nil,e};counts[k]++}
 out:=make([]string,0,len(counts));for k:=range counts {out=append(out,k)};sort.Strings(out);return out,counts,nil
}

// Link actual persisted provenance, not just issue counts or anchor similarity.
func durableLinks(observations []domain.Observation, linked map[string][]domain.Observation, prior map[string]bool) (map[string]string,[]string,error) {
 byObservation:=map[string]string{}
 carried:=[]string{}
 for id,rows:=range linked {
  if prior[id] {carried=append(carried,id)}
  for _,o:=range rows {
   if previous,ok:=byObservation[o.ObservationID];ok&&previous!=id{return nil,nil,errors.New("observation linked to multiple durable issues")}
   byObservation[o.ObservationID]=id
  }
 }
 links:=map[string]string{}
 for _,o:=range observations {
  k,e:=occurrenceKey(o);if e!=nil{return nil,nil,e}
  id:=byObservation[o.ObservationID];if id==""{return nil,nil,errors.New("current observation lacks durable issue link")}
  if previous,ok:=links[k];ok&&previous!=id{return nil,nil,errors.New("occurrence linked to multiple durable issues")}
  links[k]=id
 }
 if len(prior)==0 {
  unique:=map[string]bool{};for _,id:=range links{unique[id]=true}
  if len(unique)!=len(links)||len(unique)!=len(linked){return nil,nil,errors.New("fresh store durable issues are not one-to-one with occurrences")}
 }
 sort.Strings(carried)
 return links,carried,nil
}

// One deterministic-scanner health predicate for both routes. Model rows and
// mapping omissions are evidence, not failures of deterministic file scanning.
func scannerHealth(r receipt, expected []domain.CoverageKey, specs []analyzers.Spec, failures []domain.Failure, diagnostics []inventory.Diagnostic, runErr error) (bool,[]string) {
 healthy:=runErr==nil&&len(diagnostics)==0&&(r.Mapping.Status==clawpatch.MapMapped||r.Mapping.Status==clawpatch.MapReused)&&!r.Mapping.ExplorationIncomplete
 for _,failure:=range failures {
  if failure.Stage==domain.StageFileAnalysis||failure.Stage==domain.StageInventory||failure.Stage==domain.StageMapping{healthy=false}
 }
 seen:=map[string]bool{}
 for _,a:=range r.Adapters {
  seen[a.Adapter]=true
  acceptable:=a.Status==domain.AdapterCompleted||a.Status==domain.AdapterNoApplicableInput
  if a.Status==domain.AdapterIneligible{
   acceptable=len(a.Coverage)>0
   for _,row:=range a.Coverage{if row.Status!=domain.CoverageNotApplicable{acceptable=false}}
  }
  if len(a.Failures)>0||!acceptable{healthy=false}
 }
 for _,spec:=range specs{if !seen[spec.ID]{healthy=false}}
 rows:=map[string]bool{}
 for _,row:=range r.Coverage {
  if row.Key.FilePath==""{continue}
  k:=row.Key.String()
  if _,duplicate:=rows[k];duplicate{healthy=false}
  rows[k]=row.Status.Accounted()
  if !row.Status.Accounted(){healthy=false}
 }
 for _,key:=range expected{if key.FilePath!=""&&!rows[key.String()]{healthy=false}}
 accounted:=[]string{};for key,ok:=range rows{if ok{accounted=append(accounted,key)}};sort.Strings(accounted)
 return healthy,accounted
}

func bindingRowsComplete(coverage []domain.Coverage, jobs []domain.Job) bool {
 rows:=map[string]domain.Coverage{}
 for _,row:=range coverage {
  if row.Key.FeatureID==""{continue}
  if _,duplicate:=rows[row.Key.String()];duplicate{return false}
  if row.Status!=domain.CoverageSkipped||row.Reason!=engine.ReasonNotOptedIn{return false}
  rows[row.Key.String()]=row
 }
 for _,job:=range jobs {
  if job.Stage!=domain.StageFeatureAnalysis&&job.Stage!=domain.StageClawpatchReview{continue}
  if job.CoverageKey==nil{return false}
  if _,ok:=rows[job.CoverageKey.String()];!ok{return false}
 }
 return true
}

// Match inventory.Plan's eligible, readable units and CoverageKey job order.
// Do not prefilter Spec.Applies before Batches: Assess includes not-applicable
// units in those batches; Runner.selectInputs applies the spec inside each one.
func plannedFiles(files []domain.File) []domain.File {
 out:=[]domain.File{}
 for _,file:=range files{if file.Eligible&&!file.Excluded&&file.ContentDigest!=""{out=append(out,file)}}
 sort.Slice(out,func(i,j int)bool{return out[i].Path<out[j].Path})
 return out
}

// Record effective applicable/readable inputs within the planner's batches.
// Not-applicable rows also remove scanner-specific allowlisted paths. Failed
// or incomplete inputs cannot yield a healthy receipt.
func scannerInputs(root string, files []domain.File, specs []analyzers.Spec, coverage []domain.Coverage) map[string][][]string {
 rows:=map[string]domain.CoverageStatus{}
 for _,row:=range coverage{rows[row.Key.String()]=row.Status}
 inputs:=map[string][][]string{}
 planned:=plannedFiles(files)
 for _,spec:=range specs {
  inputs[spec.ID]=[][]string{}
  for _,batch:=range spec.Batches(root,planned) {
   paths:=[]string{}
   for _,file:=range batch {
    key:=domain.CoverageKey{FilePath:file.Path,Family:spec.Family,Category:spec.Category}
    status:=rows[key.String()]
    if !spec.Applies(file)||!status.Accounted()||status==domain.CoverageNotApplicable{continue}
    info,e:=os.Stat(filepath.Join(root,file.Path));if e!=nil||!info.Mode().IsRegular(){continue}
    paths=append(paths,file.Path)
   }
   inputs[spec.ID]=append(inputs[spec.ID],paths)
  }
 }
 return inputs
}

// This route never uses the override loader, ApplyModelCLIPolicy, Assess,
// correlation, ranking, association, or any model stage.
func reference(ctx context.Context, root, mapper string, scope domain.Scope, registry *analyzers.Registry) (receipt,[]domain.Observation,[]domain.CoverageKey,[]domain.Failure,[]inventory.Diagnostic,error) {
 r:=receipt{Route:"reference",Omissions:[]domain.Omission{},MappingGaps:[]inventory.MappingGap{}}
 fail:=func(e error)(receipt,[]domain.Observation,[]domain.CoverageKey,[]domain.Failure,[]inventory.Diagnostic,error){return r,nil,nil,nil,nil,e}
 client,e:=clawpatch.New(root,clawpatch.Config{BinaryPath:mapper,EgressAuthorized:false});if e!=nil{return fail(e)}
 mapped,e:=client.Map(ctx,true,clawpatch.MapSourceHeuristic);r.Mapping=mapped;if e!=nil{return fail(e)}
 scanned,e:=inventory.Scan(ctx,root,inventory.ScanOptions{Scope:scope,Features:mapped.Features});if e!=nil{return fail(e)}
 r.Omissions=append(r.Omissions,scanned.Omissions...);r.MappingGaps=append(r.MappingGaps,scanned.MappingGaps...)
 for _,gap:=range r.MappingGaps {
  key:=domain.CoverageKey{FilePath:gap.Path,FeatureID:gap.FeatureID,Family:domain.FamilyClawpatch,Category:"mapping"}
  if key.Validate()==nil{r.Omissions=append(r.Omissions,domain.Omission{CoverageKey:key,Reason:gap.Reason})}
 }
 var observations []domain.Observation
 expected:=[]domain.CoverageKey{}
 var scanErr error
 files:=plannedFiles(scanned.Files)
 for _,spec:=range registry.Specs() {
  runner,e:=registry.Runner(spec.ID,"x1-reference");if e!=nil{return fail(e)}
  for _,batch:=range spec.Batches(root,files) {
   adapter,found,runErr:=runner.Run(ctx,root,batch)
   if runErr!=nil{scanErr=errors.Join(scanErr,runErr)}
   r.Adapters=append(r.Adapters,adapter)
   // Match Assess.splitAdapter: off-file observations are not persisted.
   for _,file:=range batch {
    key:=domain.CoverageKey{FilePath:file.Path,Family:spec.Family,Category:spec.Category};expected=append(expected,key)
    for _,row:=range adapter.Coverage{if row.Key.FilePath==file.Path{r.Coverage=append(r.Coverage,row)}}
    observations=appendFileObservations(observations,found,file.Path)
   }
  }
 }
 final,e:=inventory.Scan(ctx,root,inventory.ScanOptions{Scope:scope,Features:mapped.Features})
 if e!=nil{scanErr=errors.Join(scanErr,e)}else if !reflect.DeepEqual(scanned.Snapshot,final.Snapshot){scanErr=errors.Join(scanErr,errors.New("source drift during reference scan"))}
 diagnostics:=append(scanned.Diagnostics,final.Diagnostics...)
 r.ScannerInputs=scannerInputs(root,scanned.Files,registry.Specs(),r.Coverage)
 return r,observations,expected,nil,diagnostics,scanErr
}

// Match the persisted splitAdapter boundary: only the primary evidence
// location determines which file owns an observation.
func appendFileObservations(out, found []domain.Observation, path string) []domain.Observation {
 for _,o:=range found{if len(o.Evidence)>0&&o.Evidence[0].Location.Path==path{out=append(out,o)}}
 return out
}

func modelFree(p policy.Policy) error {
 if p.Decisions().Enabled||len(p.ModelCLI().Lanes)!=0||p.Mapping().Association.Enabled||p.Profile()!=policy.ProfileReport{return errors.New("not model free")}
 return nil
}
func frozenPolicy(b []byte, expectedDigest string) (policy.Policy,error) {
 sum:=sha256.Sum256(b)
 if hex.EncodeToString(sum[:])!=expectedDigest{return policy.Policy{},errors.New("frozen policy digest mismatch")}
 p,e:=policy.Parse(b);if e!=nil{return p,e};return p,modelFree(p)
}
func controlPolicy(p policy.Policy, authorization string) error {
 if authorization!="1"{return errors.New("synthetic control authorization absent")}
 if !p.Decisions().Enabled||!p.LaneOptedIn(policy.LaneContextual){return errors.New("positive control policy must enable decisions and contextual lane")}
 return nil
}
func requireNoKey(present bool) error {if present{return errors.New("key present")};return nil}

func resolve(route, frozen string, original policy.Policy, cfg engine.Config, authorization string) (engine.Config,any,error) {
 switch route {
 case "bound":
  b,e:=os.ReadFile(frozen);if e!=nil{return cfg,nil,e}
  p,e:=frozenPolicy(b,frozenDigest);if e!=nil{return cfg,nil,e}
  return engine.ApplyModelCLIPolicy(cfg,p),describe(p),nil
 case "reference":
  return cfg,map[string]any{"scanner_only":true,"decisions_stage":"absent","association_stage":"absent","model_stage":"absent"},nil
 case "synthetic-unbound":
  if e:=controlPolicy(original,authorization);e!=nil{return cfg,nil,e}
  cfg=engine.ApplyModelCLIPolicy(cfg,original)
  // Only the isolated logging stub is on PATH; controller proves its bytes.
  cfg.Provider.EgressAuthorized=true
  return cfg,describe(original),nil
 default:return cfg,nil,errors.New("invalid route")
 }
}
// These are the exact API inputs, not a restatement of the policy. No PR
// metadata/history/ranking helper is called by this frontend; those are not
// fields of engine.Config or engine.Options.
func boundInputs(route string, cfg engine.Config, options engine.Options) error {
 if route=="synthetic-unbound"{return nil}
 if route!="bound"&&route!="reference"{return errors.New("invalid route")}
 expectedOptions:=engine.Options{Scope:domain.Scope{Kind:domain.ScopeRepository,SelectedFamilies:engine.DefaultFamilies()},ForceFull:true,Trigger:domain.TriggerPush}
 if !reflect.DeepEqual(options,expectedOptions){return errors.New("unexpected bound assessment options")}
 if len(cfg.ModelLanes)!=0||cfg.ClawpatchEgressAuthorized||cfg.Provider.EgressAuthorized{return errors.New("model lane or egress enabled")}
 expectedGuard:=policy.AssociationGuard{};if route=="bound"{expectedGuard=policy.Default().Mapping().Association.Guard}
 if cfg.Root==""||cfg.ClawpatchBinary==""||len(cfg.Dependencies)!=0||!reflect.DeepEqual(cfg.AssociationGuard,expectedGuard){return errors.New("unexpected bound assessment config")}
 if !reflect.DeepEqual(cfg.Provider,contextual.Config{CLI:cfg.Provider.CLI}){return errors.New("unexpected contextual provider config")}
 if route=="bound"&&(cfg.Provider.CLI!=""&&cfg.Provider.CLI!="claude"&&cfg.Provider.CLI!="codex"){return errors.New("invalid bound contextual provider")}
 if route=="bound"&&cfg.ClawpatchProvider!=cfg.Provider.CLI{return errors.New("inconsistent bound providers")}
 if route=="reference"&&(cfg.Provider.CLI!=""||cfg.ClawpatchProvider!=""){return errors.New("reference model provider present")}
 return nil
}
func configProof(route string, cfg engine.Config, options engine.Options) map[string]bool {
 valid:=route!="synthetic-unbound"&&boundInputs(route,cfg,options)==nil
 return map[string]bool{"effective_config_bound":valid,"model_lanes_disabled":valid&&len(cfg.ModelLanes)==0,"jev_disabled":valid&&options.Association==nil,"association_disabled":valid&&options.Association==nil,"pr_metadata_absent":valid,"heuristic_only":valid}
}
func completeProof(proof map[string]bool) bool {
 for _,key:=range []string{"effective_config_bound","model_lanes_disabled","jev_disabled","association_disabled","pr_metadata_absent","heuristic_only"}{if !proof[key]{return false}}
 return true
}
func runtimeProof(r *receipt, route string, cfg engine.Config, options engine.Options, result *engine.Result) {
 r.ModelProof=configProof(route,cfg,options)
 r.ModelProof["heuristic_only"]=r.ModelProof["heuristic_only"]&&r.Mapping.Source==clawpatch.MapSourceHeuristic
 r.AssociationResult="not_disabled";r.JevResult="not_proven_absent";r.ModelStages="enabled_control"
 if route=="reference"{
  r.ModelStages="structurally_absent";r.AssociationResult="structurally_absent";r.JevResult="absent"
  r.ModelCallImpossible=completeProof(r.ModelProof)
  return
 }
 if result==nil{r.ModelProof["model_lanes_disabled"]=false;r.ModelProof["association_disabled"]=false;return}
 // Assess already calls LaneHealthOf with its actual inline health. Retain
 // that report verbatim: reconstructing it could synthesize missing rows.
 r.LaneReport=result.Lanes
 var seen [4]int
 var disabled [4]bool
 reportValid:=true
 for _,h:=range r.LaneReport{
  index:=-1
  switch h.Lane{
  case engine.LaneContextual:index=0
  case engine.LaneClawpatchReview:index=1
  case engine.LaneAssociation:index=2
  case engine.LaneAssociationFallback:index=3
  }
  if index<0{
   if h.Requires==domain.RequiresModelCLI||h.Requires==domain.RequiresDecisionEgress{reportValid=false}
   continue
  }
  seen[index]++
  if index<2{
   disabled[index]=h.Status==domain.LaneSkipped&&(h.Reason==engine.ReasonNotOptedIn||(h.Units==0&&h.Reason=="no eligible inputs"))
  }else{
   r.AssociationHealth=append(r.AssociationHealth,h)
   disabled[index]=h.Status==domain.LaneSkipped&&h.Reason=="disabled"
  }
 }
 for _,count:=range seen{if count!=1{reportValid=false}}
 r.ModelProof["model_lanes_disabled"]=r.ModelProof["model_lanes_disabled"]&&reportValid&&disabled[0]&&disabled[1]&&r.BindingRowsComplete
 associationDisabled:=reportValid&&disabled[2]&&disabled[3]
 r.ModelProof["association_disabled"]=r.ModelProof["association_disabled"]&&associationDisabled
 r.ModelProof["jev_disabled"]=r.ModelProof["jev_disabled"]&&associationDisabled
 if associationDisabled{r.AssociationResult="disabled";r.JevResult="absent"}
 if route=="bound"{r.ModelStages="disabled"}
 r.ModelCallImpossible=completeProof(r.ModelProof)
}
func effectiveRecord(route, revision string, original policy.Policy, effective any, cfg engine.Config, options engine.Options) map[string]any {
 lanes:=cfg.ModelLanes;if lanes==nil{lanes=[]string{}}
 // All Config fields are represented; Registry's unexported implementation is
 // replaced by its complete serializable specs and a deterministic digest.
 var specs []analyzers.Spec;if cfg.Registry!=nil{specs=cfg.Registry.Specs()}
 encoded,_:=json.Marshal(specs);sum:=sha256.Sum256(encoded)
 config:=map[string]any{"Root":cfg.Root,"ClawpatchBinary":cfg.ClawpatchBinary,"ClawpatchEgressAuthorized":cfg.ClawpatchEgressAuthorized,"ClawpatchProvider":cfg.ClawpatchProvider,"Provider":cfg.Provider,"ModelLanes":lanes,"Registry":map[string]any{"specs":specs,"sha256":hex.EncodeToString(sum[:])},"Dependencies":cfg.Dependencies,"AssociationGuard":cfg.AssociationGuard}
 proof:=configProof(route,cfg,options)
 stages:="disabled";if route=="reference"{stages="structurally_absent"};if route=="synthetic-unbound"{stages="enabled_control"}
 return map[string]any{"route":route,"revision":revision,"original_policy":describe(original),"effective_policy":effective,"config":config,"options":options,"model_lanes":lanes,"association_nil":options.Association==nil,"measures_nil":options.Measures==nil,"ranking_stage":"absent","history_stage":"absent","pr_metadata_fetch":map[string]any{"present":false,"reason":"no PR metadata fetch API call in frontend"},"clawpatch_egress":cfg.ClawpatchEgressAuthorized,"contextual_egress":cfg.Provider.EgressAuthorized,"mapping_argv":[]string{"map","--source","heuristic"},"scope":options.Scope,"key_present":false,"model_proof":proof,"model_call_impossible":completeProof(proof),"model_stages":stages}
}
func requiredArguments(root,revision,mapper,effective,output string, resolveOnly bool) error {
 if root==""||revision==""||mapper==""||effective==""||(!resolveOnly&&output==""){return errors.New("all fixture/output arguments required")}
 return nil
}

func run() error {
 root:=flag.String("root","","fresh fixture revision root")
 revision:=flag.String("revision","","literal fixture revision")
 route:=flag.String("route","bound","bound, reference, or synthetic-unbound")
 frozen:=flag.String("policy","","signed model-free policy")
 mapper:=flag.String("clawpatch","","attested absolute Clawpatch entrypoint")
 effectivePath:=flag.String("effective-output","","exclusive pre-scanner policy/config record")
 output:=flag.String("output","","exclusive complete result")
 resolveOnly:=flag.Bool("resolve-only",false,"write effective record without Assess or scanners")
 flag.Parse()
 if e:=requiredArguments(*root,*revision,*mapper,*effectivePath,*output,*resolveOnly);e!=nil{return e}
 if _,present:=os.LookupEnv("OPENROUTER_API_KEY");present{return requireNoKey(true)}
 original,e:=policy.Load(*root,*revision);if e!=nil{return e}
 scope:=domain.Scope{Kind:domain.ScopeRepository,SelectedFamilies:engine.DefaultFamilies()}
 options:=engine.Options{Scope:scope,ForceFull:true,Trigger:domain.TriggerPush}
 cfg:=engine.Config{Root:*root,ClawpatchBinary:*mapper,Provider:contextual.Config{EgressAuthorized:false}}
 cfg,effective,e:=resolve(*route,*frozen,original,cfg,os.Getenv("X1_SYNTHETIC_CONTROL"));if e!=nil{return e}
 // Validate actual Config/Options before registry construction or scanners.
 if e=boundInputs(*route,cfg,options);e!=nil{return e}
 registry,e:=analyzers.BuiltinRegistry();if e!=nil{return e};cfg.Registry=registry
 if e=boundInputs(*route,cfg,options);e!=nil{return e}
 if e=writeJSON(*effectivePath,effectiveRecord(*route,*revision,original,effective,cfg,options));e!=nil{return e}
 // Registry construction is pure catalogue validation; no scanner or state access.
 if *resolveOnly{return nil}
 var r receipt;var observations []domain.Observation;var expected []domain.CoverageKey
 var failures []domain.Failure;var diagnostics []inventory.Diagnostic
 var assessed *engine.Result
 ctx:=context.Background()
 if *route=="reference" {
  r,observations,expected,failures,diagnostics,e=reference(ctx,*root,*mapper,scope,registry)
  r.BindingRowsComplete=true
 } else {
  prior:=map[string]bool{}
  store,se:=state.Open(*root);if se!=nil{return se}
  previous,se:=store.ListIssues(ctx);store.Close();if se!=nil{return se}
  for _,issue:=range previous{prior[issue.IssueID]=true}
  var result engine.Result
  result,e=engine.Assess(ctx,cfg,options)
  assessed=&result
  r=receipt{Adapters:result.Adapters,Coverage:result.Run.Coverage,Mapping:result.Map,MappingGaps:result.MappingGaps,Omissions:result.Run.Omissions}
  expected=result.Run.ExpectedCoverage;failures=result.Run.Failures
  r.BindingRowsComplete=bindingRowsComplete(r.Coverage,result.Plan.Jobs)
  if result.Run.RunID!="" {
   store,se=state.Open(*root);if se!=nil{return se}
   observations,se=store.LoadObservations(ctx,result.Run.RunID);if se!=nil{store.Close();return se}
   all,se:=store.ListIssues(ctx);if se!=nil{store.Close();return se}
   r.DurableCount=len(all)
   linked:=map[string][]domain.Observation{}
   for _,issue:=range all {
    rows,le:=store.ListIssueObservations(ctx,issue.IssueID);if le!=nil{store.Close();return le};linked[issue.IssueID]=rows
   }
   store.Close()
   var linkErr error
   r.DurableLinks,r.CarriedIDs,linkErr=durableLinks(observations,linked,prior)
   if linkErr!=nil{e=errors.Join(e,linkErr)}
  }
  current,se:=inventory.Scan(ctx,*root,inventory.ScanOptions{Scope:scope,Features:result.Map.Features})
  if se!=nil{e=errors.Join(e,se)}else{
   diagnostics=current.Diagnostics
   if result.Run.RunID!=""&&!reflect.DeepEqual(current.Snapshot,result.Run.Snapshot){e=errors.Join(e,errors.New("source drift after assessment"))}
   r.ScannerInputs=scannerInputs(*root,current.Files,registry.Specs(),r.Coverage)
  }
 }
 r.Route=*route;r.Revision=*revision;r.OriginalPolicy=describe(original);r.Effective=effective
 runtimeProof(&r,*route,cfg,options,assessed)
 if *route=="bound"&&!r.ModelCallImpossible{e=errors.Join(e,errors.New("deterministic model-free runtime proof incomplete"))}
 r.Healthy,r.AccountedCoverage=scannerHealth(r,expected,registry.Specs(),failures,diagnostics,e)
 if *route=="bound"&&!r.BindingRowsComplete{r.Healthy=false;e=errors.Join(e,errors.New("bound feature/review rows are not all skipped (not opted in)"))}
 if e!=nil{r.Error=e.Error()}
 r.Keys,r.Multiplicity,e=keys(observations);if e!=nil{return e}
 if r.DurableLinks==nil{r.DurableLinks=map[string]string{}};if r.CarriedIDs==nil{r.CarriedIDs=[]string{}}
 if r.Omissions==nil{r.Omissions=[]domain.Omission{}};if r.MappingGaps==nil{r.MappingGaps=[]inventory.MappingGap{}}
 sort.Slice(r.Omissions,func(i,j int)bool{left,_:=json.Marshal(r.Omissions[i]);right,_:=json.Marshal(r.Omissions[j]);return string(left)<string(right)})
 sort.Slice(r.MappingGaps,func(i,j int)bool{left,_:=json.Marshal(r.MappingGaps[i]);right,_:=json.Marshal(r.MappingGaps[j]);return string(left)<string(right)})
 if r.ScannerInputs==nil{r.ScannerInputs=map[string][][]string{}}
 if e=writeJSON(*output,r);e!=nil{return e};if !r.Healthy{return errors.New("scanner health incomplete; retained receipt")};return nil
}
func main(){if e:=run();e!=nil {fmt.Fprintln(os.Stderr,e);os.Exit(1)}}

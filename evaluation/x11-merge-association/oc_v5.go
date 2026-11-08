// Standalone synthetic X11a OC. No repository scanner, model, network or replay.
package main

import (
 "bytes"
 "crypto/sha256"
 "encoding/binary"
 "encoding/hex"
 "encoding/json"
 "flag"
 "fmt"
 "math"
 "math/rand/v2"
 "os"
 "runtime"
 "sort"
 "strconv"
 "strings"
 "time"
)
const scenarios=56160
var bootstrap=399
const budgetSeconds=384*3600
var looks=[]int{256,512,1024}
type Scenario struct { Index int; None [3]float64; Truth int; Rho,Cross,Age,Missing,DirectoryRho float64; Mechanism string; Uniform,SharedError float64; Temporal int }
type Counts struct { C,A,E,UA,U int }
type Case struct { Position int; Directory string; Existing bool; Assigned,Correct [2]bool; Priority float64; Missing,RawA,RawB,Agreed bool }
type Cell struct { Blocks [2000][2]Counts; Directories map[string][2]Counts; Quality bool; Cases int; Removed int }
type Result struct { Supported,Negative [2]bool; FalseSupport,Boundary bool; Studies,BootstrapDraws,GeneratedCases int }
type DirectoryCase struct { Position int `json:"position"`; DirectoryHash string `json:"directory_hash"` }
type Vector struct { SegmentLengths []int `json:"segment_lengths"`; Counts [][][]int `json:"counts"`; DirectoryCases [][][]DirectoryCase `json:"directory_cases"` }
func seed(index,rep int, stream string) uint64 { h:=sha256.Sum256([]byte(fmt.Sprintf("x11a-oc-v6\x00%d\x00%d\x00%s",index,rep,stream))); return binary.BigEndian.Uint64(h[:8]) }
func rng(index,rep int, stream string)*rand.Rand { s:=seed(index,rep,stream); return rand.New(rand.NewPCG(s,s^0x9e3779b97f4a7c15)) }
// Open interval even at the endpoints of a 53-bit draw; no inverse CDF sees 0/1.
func uniform(r *rand.Rand)float64 { return (float64(r.Uint64()>>12)+.5)/4503599627370496 }
func normal(r *rand.Rand)float64{return math.Sqrt(2)*math.Erfinv(2*uniform(r)-1)}
func phi(x float64)float64{return .5*math.Erfc(-x/math.Sqrt2)}
// Separable normal factor: sqrt(a*c)Z + sqrt(a*(1-c))Zarm
// + sqrt((1-a)*c)Zage + sqrt((1-a)*(1-c))Zagearm.
func copula(r *rand.Rand,a,c float64)[2][2]float64 { z:=normal(r); arm:=[2]float64{normal(r),normal(r)}; age:=[2]float64{normal(r),normal(r)}; var out [2][2]float64; for i:=0;i<2;i++ { for j:=0;j<2;j++ { e:=normal(r); out[i][j]=phi(math.Sqrt(a*c)*z+math.Sqrt(a*(1-c))*arm[j]+math.Sqrt((1-a)*c)*age[i]+math.Sqrt((1-a)*(1-c))*e) } }; return out }
func betaCF(a,b,x float64)float64 { const tiny=1e-300; qab:=a+b; qap:=a+1; qam:=a-1; c:=1.; d:=1-qab*x/qap; if math.Abs(d)<tiny {d=tiny}; d=1/d; h:=d; for m:=1;m<=300;m++ { fm:=float64(m); aa:=fm*(b-fm)*x/((qam+2*fm)*(a+2*fm)); d=1+aa*d;if math.Abs(d)<tiny{d=tiny};c=1+aa/c;if math.Abs(c)<tiny{c=tiny};d=1/d;h*=d*c; aa=-(a+fm)*(qab+fm)*x/((a+2*fm)*(qap+2*fm));d=1+aa*d;if math.Abs(d)<tiny{d=tiny};c=1+aa/c;if math.Abs(c)<tiny{c=tiny};d=1/d;delta:=d*c;h*=delta;if math.Abs(delta-1)<2e-14{return h} }; panic("beta continued fraction did not converge") }
func betaCDF(x,a,b float64)float64 { if x<=0{return 0};if x>=1{return 1}; la,_:=math.Lgamma(a);lb,_:=math.Lgamma(b);lab,_:=math.Lgamma(a+b);bt:=math.Exp(lab-la-lb+a*math.Log(x)+b*math.Log1p(-x));if x<(a+1)/(a+b+2){return bt*betaCF(a,b,x)/a};return 1-bt*betaCF(b,a,1-x)/b }
func betaInv(u,a,b float64)float64 { lo,hi:=0.,1.; for k:=0;k<48;k++ {m:=(lo+hi)/2;if betaCDF(m,a,b)<u{lo=m}else{hi=m}};return (lo+hi)/2 }
func betaMean(u,mean,rho float64)float64 { z:=0.;if u<mean{z=1};return (1-math.Sqrt(rho))*mean+math.Sqrt(rho)*z }
func scenario(index int)Scenario { original:=index;di:=index%2;index/=2;mi:=index%5;index/=5;ai:=index%2;index/=2;ci:=index%3;index/=3;ri:=index%2;index/=2;ti:=index%39;qi:=index/39;selected:=qi%3;qi/=3;second:=qi%2;first:=qi/2;s:=Scenario{Index:original,None:[3]float64{[]float64{.2,.4}[first],[]float64{.2,.4}[second],[]float64{.2,.4,.6}[selected]},Truth:ti,Rho:[]float64{0,.1}[ri],Cross:[]float64{0,.5,1}[ci],Age:[]float64{.5,1}[ai],DirectoryRho:[]float64{0,.1}[di],Mechanism:"uniform"};if mi>0{s.Missing=[]float64{.025,.05}[(mi-1)/2];if mi%2==0{s.Mechanism="correct"}};return s }
// Fixed one-factor sensitivity catalogue: 15 additional none vectors, then
// 12 binding none vectors crossed with seven one-factor departures.
func sensitivity(index int)Scenario {
 original:=index;di:=index%2;index/=2
 if index<15 {
  k:=0
  for a:=range 3 {for b:=range 3 {for c:=range 3 {
   if a<2&&b<2{continue}
   if k==index{return Scenario{Index:scenarios+original,None:[3]float64{[]float64{.2,.4,.6}[a],[]float64{.2,.4,.6}[b],[]float64{.2,.4,.6}[c]},Rho:.1,Cross:.5,Age:.5,DirectoryRho:[]float64{0,.1}[di],Missing:.025,Mechanism:"uniform"}}
   k++
  }}}
 }
 k:=index-15;q:=k/7;variant:=k%7
 s:=scenario(q*4680+80);s.Index=scenarios+original;s.Truth=0;s.Rho=.1;s.Cross=.5;s.Age=.5;s.DirectoryRho=[]float64{0,.1}[di];s.Missing=.025;s.Mechanism="uniform"
 switch variant {case 0:s.Uniform=.85;case 1:s.Uniform=.88;case 2:s.Mechanism="error";case 3:s.Rho=.3;case 4:s.Temporal=20;case 5:s.Temporal=40;case 6:s.SharedError=.02}
 return s
}
func truth(s Scenario,repo,age,arm int)(p,r float64){p,r=.9,.9;if s.Truth==1{return .8,.8};if s.Truth==2{v:=[]float64{.76,.84}[age];return v,v};if s.Truth>=3 { t:=s.Truth-3; both:=t>=24; if both{t-=24}; level:=t/12;if both{level=t/6};rest:=t%12;if both{rest=t%6};targetRepo:=rest/4;targetArm:=(rest/2)%2;metric:=rest%2;if both{targetRepo=rest/2};if repo==targetRepo&&(both||arm==targetArm){v:=[]float64{.75,.78}[level];if metric==0{p=v}else{r=v}} };q:=1-s.None[repo];if q*r*(1/p-1)/(1-q)>1{r=p*(1-q)/(q*(1-p))};return }
func margins(s Scenario,repo,age,arm int)(float64,float64){if s.Uniform>0{return s.Uniform,s.Uniform};return truth(s,repo,age,arm)}
func add(a,b Counts)Counts{return Counts{a.C+b.C,a.A+b.A,a.E+b.E,a.UA+b.UA,a.U+b.U}}
func ratios(c Counts)[4]float64{ d:=[4]int{c.A+c.UA,c.E+c.U,c.A+c.UA,c.E+c.UA};n:=[4]int{c.C,c.C,c.C+c.UA,c.C+c.UA};var out [4]float64;for k:=0;k<4;k++{if d[k]==0{out[k]=math.NaN()}else{out[k]=float64(n[k])/float64(d[k])}};return out}
func label(r *rand.Rand,existing bool,kind,wrongID int)(bool,int){covered:=existing;if uniform(r)<.01{covered=!covered};id:=0;if covered{if !existing||uniform(r)<.01{id=wrongID}};if !existing&&uniform(r)<.01{kind=1-kind};return covered,id*2+kind}
func generate(s Scenario,rep,repo int,v *Vector)[2]Cell {
 r:=rng(s.Index,rep,fmt.Sprintf("repo/%d",repo))
 var cells [2]Cell
 var cases [2][]Case
 var nLabels,coveredAgree,routed,existingA,existingB [2]int
 for age:=range 2 { n:=0;for _,m:=range v.Counts[repo][age]{n+=m};cases[age]=make([]Case,0,n) }
 keys:=directoryKeys(v,repo)
 directoryLatent:=make(map[string][2][2][2]float64,len(keys))
 dr:=rng(s.Index,rep,fmt.Sprintf("directory/%d",repo))
 for _,key:=range keys {directoryLatent[key]=[2][2][2]float64{copula(dr,s.Age,s.Cross),copula(dr,s.Age,s.Cross)}}
 var directoryAt [2][2000][]string
 for age:=range 2 {cells[age].Directories=make(map[string][2]Counts);for _,c:=range v.DirectoryCases[repo][age]{directoryAt[age][c.Position]=append(directoryAt[age][c.Position],c.DirectoryHash)}}
 var correct,fp [2][2]float64
 for pos:=range v.SegmentLengths[repo] {
  n:=max(v.Counts[repo][0][pos],v.Counts[repo][1][pos])
  if s.Temporal==0||pos%s.Temporal==0 {
   bu,fu:=copula(r,s.Age,s.Cross),copula(r,s.Age,s.Cross)
   for age:=range 2 {for arm:=range 2 {
    p,rec:=margins(s,repo,age,arm);q:=1-s.None[repo]
    correct[age][arm]=betaMean(bu[age][arm],rec,s.Rho)
    fp[age][arm]=betaMean(fu[age][arm],q*rec*(1/p-1)/(1-q),s.Rho)
   }}
  }
  for k:=range n {
   gold,trial:=copula(r,s.Age,1),copula(r,s.Age,s.Cross)
   for age:=range 2 {
    if k>=v.Counts[repo][age][pos]{continue}
    ex:=gold[age][0]<1-s.None[repo]
    key:=directoryAt[age][pos][k]
    c:=Case{Position:pos,Directory:key,Existing:ex}
    for arm:=range 2 {
     p,rec:=margins(s,repo,age,arm);q:=1-s.None[repo];w:=math.Sqrt(s.DirectoryRho);latent:=directoryLatent[key]
     dc,df:=0.,0.;if latent[0][age][arm]<rec{dc=1};if latent[1][age][arm]<q*rec*(1/p-1)/(1-q){df=1}
     if math.Sqrt(s.Rho)+w>1{panic("negative latent mean weight")}
     cp,fpCase:=correct[age][arm]+w*(dc-rec),fp[age][arm]+w*(df-q*rec*(1/p-1)/(1-q))
     if ex {c.Correct[arm]=trial[age][arm]<cp;c.Assigned[arm]=c.Correct[arm]} else {c.Assigned[arm]=trial[age][arm]<fpCase}
    }
    kind:=0;if !ex&&uniform(r)<.5{kind=1}
    a,aid:=label(r,ex,kind,1);b,bid:=label(r,ex,kind,2)
    if s.SharedError>0&&uniform(r)<s.SharedError {a,b=!ex,!ex;aid,bid=0,0;if a{aid,bid=2,2}}
    nLabels[age]++;if a{existingA[age]++};if b{existingB[age]++}
    if a==b{coveredAgree[age]++}
    agreed:=a==b&&((a&&aid/2==bid/2)||(!a&&aid%2==bid%2))
    if agreed {c.Existing=a;for arm:=range 2{if a!=ex{c.Correct[arm]=a&&c.Assigned[arm]}else if a&&aid/2!=0{c.Correct[arm]=false}}} else {routed[age]++}
    c.RawA,c.RawB,c.Agreed=a,b,agreed
    w:=1.
    if s.Mechanism=="correct"&&(c.Correct[0]||c.Correct[1]){w=4}
    if s.Mechanism=="error"&&((c.Assigned[0]&&!c.Correct[0])||(c.Assigned[1]&&!c.Correct[1])||(c.Existing&&(!c.Assigned[0]||!c.Assigned[1]))){w=4}
    c.Priority=-math.Log(uniform(r))/w
    cases[age]=append(cases[age],c)
   }
  }
 }
 for age:=range 2 {
  cs:=cases[age]
  var totals [2]Counts
  for _,c:=range cs {for arm:=range 2 {
   if c.Existing{totals[arm].E++};if c.Assigned[arm]{totals[arm].A++};if c.Correct[arm]{totals[arm].C++}
  }}
  order:=make([]int,len(cs));for k:=range order{order[k]=k}
  if s.Missing>0 {
   sort.Slice(order,func(i,j int)bool{a,b:=cs[order[i]],cs[order[j]];if a.Priority==b.Priority{return order[i]<order[j]};return a.Priority<b.Priority})
   for _,k:=range order {
    candidate:=totals
    for arm:=range 2 {candidate[arm].U++;if cs[k].Existing{candidate[arm].E--};if cs[k].Assigned[arm]{candidate[arm].A--;candidate[arm].UA++};if cs[k].Correct[arm]{candidate[arm].C--}}
    ok:=true
    for arm:=range 2 {t:=candidate[arm];if float64(t.U)/float64(t.E+t.U)>s.Missing||t.A+t.UA>0&&float64(t.UA)/float64(t.A+t.UA)>s.Missing{ok=false}}
    if !ok{break}
    cs[k].Missing=true;totals=candidate;cells[age].Removed++
   }
  }
  // Missing answers are excluded from raw A/B quality.
  nLabels[age],coveredAgree[age],routed[age],existingA[age],existingB[age]=0,0,0,0,0
  for _,c:=range cs {if c.Missing{continue};nLabels[age]++;if c.RawA{existingA[age]++};if c.RawB{existingB[age]++};if c.RawA==c.RawB{coveredAgree[age]++};if !c.Agreed{routed[age]++}}
  po:=float64(coveredAgree[age])/float64(nLabels[age])
  pa,pb:=float64(existingA[age])/float64(nLabels[age]),float64(existingB[age])/float64(nLabels[age])
  pe:=pa*pb+(1-pa)*(1-pb)
  cells[age].Quality=nLabels[age]>0&&pe<1&&(po-pe)/(1-pe)>=.60&&float64(routed[age])/float64(nLabels[age])<=.40
  cells[age].Cases=len(cs)
  for _,c:=range cs {for arm:=range 2 {
   d:=Counts{}
   if c.Missing {d.U=1;if c.Assigned[arm]{d.UA=1}} else {if c.Existing{d.E=1};if c.Assigned[arm]{d.A=1};if c.Correct[arm]{d.C=1}}
   cells[age].Blocks[c.Position][arm]=add(cells[age].Blocks[c.Position][arm],d)
   byDirectory:=cells[age].Directories[c.Directory];byDirectory[arm]=add(byDirectory[arm],d);cells[age].Directories[c.Directory]=byDirectory
  }}
 }
 return cells
}
func directoryKeys(v *Vector,repo int)[]string {
 set:=map[string]bool{}
 for age:=range 2 {for _,c:=range v.DirectoryCases[repo][age]{set[c.DirectoryHash]=true}}
 keys:=make([]string,0,len(set));for key:=range set{keys=append(keys,key)};sort.Strings(keys);return keys
}
func validateVector(v *Vector) {
 if len(v.SegmentLengths)!=3||len(v.Counts)!=3||len(v.DirectoryCases)!=3{panic("invalid repository dimensions")}
 for repo:=range 3 {
  length:=v.SegmentLengths[repo]
  if length<1500||length>2000||len(v.Counts[repo])!=2||len(v.DirectoryCases[repo])!=2{panic("invalid segment length/age dimensions")}
  for age:=range 2 {
  if len(v.Counts[repo][age])!=length||v.DirectoryCases[repo][age]==nil{panic("invalid count/case vector")}
  var counted [2000]int;last:=-1
  for _,c:=range v.DirectoryCases[repo][age] {
   if c.Position<last||c.Position<0||c.Position>=length{panic("invalid directory position/order")}
   raw,e:=hex.DecodeString(c.DirectoryHash);if e!=nil||len(raw)!=32||hex.EncodeToString(raw)!=c.DirectoryHash{panic("invalid directory hash")}
   counted[c.Position]++;last=c.Position
  }
  n:=0;for pos,m:=range v.Counts[repo][age]{if m<0||m!=counted[pos]{panic("counts/directory_cases mismatch")};n+=m};if n>2600{panic("projection exceeds label cap")}
 }}
}
func recordStop(status *string,reasons *[]string,reason string){for _,old:=range *reasons{if old==reason{return}};*reasons=append(*reasons,reason);if len(*reasons)==1{*status=reason}}
func commitSample(br *rand.Rand,length int,cells *[2]Cell,windows *[2][2000][2]Counts)([2][2]Counts,int){
 var total [2][2]Counts
 blocks:=(length+19)/20
 for block:=range blocks {index:=br.IntN(length);width:=min(20,length-block*20);for age:=range 2 {for arm:=range 2 {
  if width==20{total[age][arm]=add(total[age][arm],windows[age][index][arm])}else{for j:=range width{total[age][arm]=add(total[age][arm],cells[age].Blocks[(index+j)%length][arm])}}
 }}}
 return total,blocks
}
func combinedBounds(block,directory []float64)(float64,float64,bool) {
 minimum:=int(math.Ceil(.95*float64(bootstrap)))
 for _,values:=range [][]float64{block,directory}{for _,value:=range values{if math.IsNaN(value)||math.IsInf(value,0)||value<0||value>1{return 0,1,false}}}
 if len(block)<minimum||len(directory)<minimum{return 0,1,false}
 sort.Float64s(block);sort.Float64s(directory)
 lo:=math.Min(block[int(.025*float64(len(block)-1))],directory[int(.025*float64(len(directory)-1))])
 hi:=math.Max(block[int(.975*float64(len(block)-1))],directory[int(.975*float64(len(directory)-1))])
 return lo,hi,true
}
func study(s Scenario,rep int,v *Vector)Result {
 res:=Result{Studies:1,Supported:[2]bool{true,true}}
 defect,boundary:=[2]bool{},[2]bool{}
 for repo:=range 3 {
  cells:=generate(s,rep,repo,v)
  var total [2][2]Counts;var active [2]int
  length:=v.SegmentLengths[repo]
  for age:=range 2 {res.GeneratedCases+=cells[age].Cases;for pos:=range length {if v.Counts[repo][age][pos]>0{active[age]++};for arm:=range 2{total[age][arm]=add(total[age][arm],cells[age].Blocks[pos][arm])}}}
  var pooled [2][4]float64
  for arm:=range 2 {for age:=range 2 {r:=ratios(total[age][arm]);for metric:=range 4{pooled[arm][metric]+=.5*r[metric]}}
   p0,r0:=margins(s,repo,0,arm);p1,r1:=margins(s,repo,1,arm)
   for _,t:=range []float64{.5*(p0+p1),.5*(r0+r1)} {if t<.8-1e-14{defect[arm]=true};if math.Abs(t-.8)<1e-14{boundary[arm]=true}}
  }
  var windows [2][2000][2]Counts
  for age:=range 2 {for start:=range length {for j:=range 20 {for arm:=range 2 {windows[age][start][arm]=add(windows[age][start][arm],cells[age].Blocks[(start+j)%length][arm])}}}}
  keys:=directoryKeys(v,repo)
  var classes [][2][2]Counts
  classIndex:=make(map[[2][2]Counts]int)
  keyClass:=make([]int,len(keys))
  for i,key:=range keys {
   var d [2][2]Counts;for age:=range 2{d[age]=cells[age].Directories[key]}
   index,ok:=classIndex[d];if !ok{index=len(classes);classIndex[d]=index;classes=append(classes,d)}
   keyClass[i]=index
  }
  multiplicity:=make([]int,len(classes))
  var samples [2][2][4][]float64
  for method:=range 2 {
   stream:=fmt.Sprintf("bootstrap/%d",repo);if method==1{stream=fmt.Sprintf("directory-bootstrap/%d",repo)}
   br:=rng(s.Index,rep,stream)
   for range bootstrap {
    var bt [2][2]Counts
    if method==0 {
     var n int;bt,n=commitSample(br,length,&cells,&windows);res.BootstrapDraws+=n
    } else {
     clear(multiplicity)
     for range len(keys) {multiplicity[keyClass[br.IntN(len(keys))]]++;res.BootstrapDraws++}
     for i,n:=range multiplicity {if n==0{continue};for age:=range 2 {for arm:=range 2 {d:=classes[i][age][arm];bt[age][arm]=add(bt[age][arm],Counts{d.C*n,d.A*n,d.E*n,d.UA*n,d.U*n})}}}
    }
    for arm:=range 2 {r0,r1:=ratios(bt[0][arm]),ratios(bt[1][arm]);for metric:=range 4 {x:=.5*(r0[metric]+r1[metric]);if !math.IsNaN(x){samples[method][arm][metric]=append(samples[method][arm][metric],x)}}}
   }
  }
  for arm:=range 2 {
   supportOK:=true
   for age:=range 2 {t:=total[age][arm];if !cells[age].Quality||cells[age].Cases<1200||active[age]<100||t.A<800||t.E<800||float64(t.U)/float64(t.E+t.U)>.05||t.A+t.UA>0&&float64(t.UA)/float64(t.A+t.UA)>.05{supportOK=false}}
   for metric:=range 2 {
    lo,_,validLo:=combinedBounds(samples[0][arm][metric],samples[1][arm][metric])
    _,hi,validHi:=combinedBounds(samples[0][arm][metric+2],samples[1][arm][metric+2])
    if !validLo||pooled[arm][metric]<.8||lo<.8{supportOK=false}
    negativeAdequate:=cells[0].Quality&&cells[1].Quality
    for age:=range 2 {t:=total[age][arm];den:=t.A;if metric==1{den=t.E};if den<200||float64(t.U)/float64(t.E+t.U)>.05||t.A+t.UA>0&&float64(t.UA)/float64(t.A+t.UA)>.05{negativeAdequate=false}}
    if negativeAdequate&&validHi&&hi<.8{res.Negative[arm]=true}
   }
   res.Supported[arm]=res.Supported[arm]&&supportOK
  }
 }
 for arm:=range 2 {if res.Negative[arm]{res.Supported[arm]=false};res.FalseSupport=res.FalseSupport||res.Supported[arm]&&defect[arm];res.Boundary=res.Boundary||res.Supported[arm]&&boundary[arm]}
 return res
}
func interval(k,n int,alpha float64)(float64,float64){lo,hi:=0.,1.;if k>0{lo=betaInv(alpha/2,float64(k),float64(n-k+1))};if k<n{hi=betaInv(1-alpha/2,float64(k+1),float64(n-k))};return lo,hi}
func gateInterval(k,n int)(float64,float64){return interval(k,n,2*.05/3)}
func main() {
 mode:=flag.String("mode","timing","timing or oc (later council authority required)")
 out:=flag.String("out","","create-once output path")
 vectors:=flag.String("vectors","","frozen inventory projection JSON for authorised OC only")
 sourcePath:=flag.String("source","","absolute reviewed source path")
 timingReplicates:=flag.Int("timing-replicates",20,"synthetic replicates per timing scenario")
 profile:=flag.String("directory-profile","balanced","timing only: balanced, concentrated, singleton")
 lengthsFlag:=flag.String("segment-lengths","1500,1666,2000","timing only: three segment lengths1500..2000")
 flag.IntVar(&bootstrap,"bootstrap",399,"timing only; OC fixed at 399")
 flag.Parse()
 if *out==""||*sourcePath==""{panic("--out and --source required")}
 if bootstrap<1||*timingReplicates<1||(*mode!="timing"&&bootstrap!=399){panic("invalid work count")}
 if *mode!="timing" {flag.Visit(func(f *flag.Flag){if f.Name=="timing-replicates"||f.Name=="directory-profile"||f.Name=="segment-lengths"{panic("timing override forbidden")}})}
 source,e:=os.ReadFile(*sourcePath);if e!=nil{panic(e)};digest:=sha256.Sum256(source)
 started:=time.Now()
 var v Vector
 if *mode=="timing" {
  if *profile!="balanced"&&*profile!="concentrated"&&*profile!="singleton"{panic("unknown directory profile")}
  parts:=strings.Split(*lengthsFlag,",");if len(parts)!=3{panic("three segment lengths required")}
  v.SegmentLengths=make([]int,3);v.Counts=make([][][]int,3);v.DirectoryCases=make([][][]DirectoryCase,3)
  for repo,part:=range parts {n,err:=strconv.Atoi(part);if err!=nil||n<1500||n>2000{panic("invalid timing segment length")};v.SegmentLengths[repo]=n;v.Counts[repo]=make([][]int,2);v.DirectoryCases[repo]=make([][]DirectoryCase,2);for age:=range 2{v.Counts[repo][age]=make([]int,n);v.DirectoryCases[repo][age]=[]DirectoryCase{}}}
  for repo:=range 3 {for age:=range 2 {for k:=range 2600 {
   pos:=k%v.SegmentLengths[repo];v.Counts[repo][age][pos]++;dir:=k%100
   if *profile=="concentrated"{dir=0;if k>=1300{dir=1+k%99}};if *profile=="singleton"{dir=age*2600+k}
   h:=sha256.Sum256([]byte(fmt.Sprintf("x11a-directory-v6\x00synthetic/%d\x00dir/%d",repo,dir)))
   v.DirectoryCases[repo][age]=append(v.DirectoryCases[repo][age],DirectoryCase{Position:pos,DirectoryHash:hex.EncodeToString(h[:])})
  };sort.SliceStable(v.DirectoryCases[repo][age],func(i,j int)bool{return v.DirectoryCases[repo][age][i].Position<v.DirectoryCases[repo][age][j].Position})}}
 } else if *mode=="oc"||*mode=="sensitivity" {
  if *vectors==""{panic("--vectors required")}
  raw,e:=os.ReadFile(*vectors);if e!=nil{panic(e)};if bytes.Contains(raw,[]byte("null")){panic("null vector field")}
  d:=json.NewDecoder(bytes.NewReader(raw));d.DisallowUnknownFields();if e=d.Decode(&v);e!=nil{panic(e)}
  validateVector(&v)
 } else {panic("unknown mode")}
 validateVector(&v)
 rows:=[]map[string]any{}
 studies,cases,draws:=0,0,0
 status:="timing-only-not-an-OC-result"
 stopReasons:=[]string{}
 stop:=func(reason string){recordStop(&status,&stopReasons,reason)}
 indices:=[]int{0,191,1439,2879,40319,56159}
 if *mode=="oc"{indices=make([]int,scenarios);for k:=range indices{indices[k]=k};status="pass"}
 if *mode=="sensitivity"{indices=make([]int,198);for k:=range indices{indices[k]=k};status="descriptive-sensitivity-only"}
 timeBudget:=budgetSeconds;if *mode=="sensitivity"{timeBudget=8*3600}
 alpha:=.05/(float64(scenarios)*float64(len(looks))*6)
 for _,index:=range indices {
  s:=scenario(index);if *mode=="sensitivity"{s=sensitivity(index)}
  counts:=[6]int{}
  n:=0
  limit:=*timingReplicates;if *mode!="timing"{limit=looks[len(looks)-1]}
  decision:="unresolved"
  history:=[]map[string]any{}
  for rep:=range limit {
   if time.Since(started).Seconds()>float64(timeBudget){stop("not-testable-compute-budget");break}
   r:=study(s,rep,&v)
   studies++;cases+=r.GeneratedCases;draws+=r.BootstrapDraws;n++
   if r.Supported[0]{counts[0]++};if r.Negative[0]{counts[1]++}
   if r.Supported[1]{counts[2]++};if r.Negative[1]{counts[3]++}
   if r.FalseSupport{counts[4]++};if r.Boundary{counts[5]++}
   if *mode=="oc"&&(n==256||n==512||n==1024) {
    var bounds [6][2]float64
    var descriptive [6][2]float64
    for event:=range 6{bounds[event][0],bounds[event][1]=gateInterval(counts[event],n);descriptive[event][0],descriptive[event][1]=interval(counts[event],n,alpha)}
    history=append(history,map[string]any{"n":n,"counts":counts,"gate_one_sided_bounds":bounds,"descriptive_simultaneous_bounds":descriptive})
    if s.Truth==0 {
     if bounds[0][0]>=.8&&bounds[1][1]<=.05&&bounds[2][0]>=.8&&bounds[3][1]<=.05{decision="pass"}
     if bounds[0][1]<.8||bounds[1][0]>.05||bounds[2][1]<.8||bounds[3][0]>.05{decision="fail"}
    } else {
     if bounds[4][1]<=.05&&bounds[5][1]<=.05{decision="pass"}
     if bounds[4][0]>.05||bounds[5][0]>.05{decision="fail"}
    }
    if decision!="unresolved"{break}
   }
  }
  var finalBounds [6][2]float64
  for event:=range 6{finalBounds[event][0],finalBounds[event][1]=interval(counts[event],n,alpha)}
  rows=append(rows,map[string]any{"scenario":s,"replicates":n,"event_counts":counts,"bounds":finalBounds,"decision":decision,"looks":history})
  if *mode=="oc"&&decision!="pass"{stop("not-testable-OC-gate")}
  if time.Since(started).Seconds()>float64(timeBudget){stop("not-testable-compute-budget");break}
  if *mode=="oc"&&studies>0&&time.Since(started).Seconds()/float64(studies)*float64(scenarios*256)>float64(timeBudget){stop("not-testable-projected-compute-budget");break}
 }
 elapsed:=time.Since(started).Seconds()
 exe,_:=os.Executable()
 projectionScenarios:=scenarios;minimumReplicates:=256
 if *mode=="sensitivity"{projectionScenarios=len(indices);minimumReplicates=1024}
 directorySummary:=make([][]map[string]any,3)
 for repo:=range 3{directorySummary[repo]=make([]map[string]any,2);for age:=range 2{
  shares:=map[string]int{};for _,c:=range v.DirectoryCases[repo][age]{shares[c.DirectoryHash]++}
  n:=len(v.DirectoryCases[repo][age]);sumSquares:=0;largest:=0;for _,count:=range shares{sumSquares+=count*count;largest=max(largest,count)}
  effective,largestShare:=0.,0.;if n>0{effective=float64(n*n)/float64(sumSquares);largestShare=float64(largest)/float64(n)}
  directorySummary[repo][age]=map[string]any{"N":n,"K":len(shares),"effective_directory_count":effective,"largest_directory_share":largestShare}
 }}
 report:=map[string]any{
  "schema":"x11a-oc-v7-result-v1","mode":*mode,"status":status,"stop_reasons":stopReasons,"directory_profile":*profile,
  "segment_lengths":v.SegmentLengths,"commit_block_length":20,"commit_block_draws":[]int{(v.SegmentLengths[0]+19)/20,(v.SegmentLengths[1]+19)/20,(v.SegmentLengths[2]+19)/20},
  "final_block_positions":[]int{1+(v.SegmentLengths[0]-1)%20,1+(v.SegmentLengths[1]-1)%20,1+(v.SegmentLengths[2]-1)%20},
  "compute_projection_basis":"minimum 256-replicate projection over all binding scenarios",
  "directory_summary":directorySummary,
  "runtime":runtime.Version(),"goos":runtime.GOOS,"goarch":runtime.GOARCH,"executable":exe,
  "source_sha256":hex.EncodeToString(digest[:]),"elapsed_seconds":elapsed,
  "studies":studies,"generated_cases":cases,"adoption_arm_case_evaluations":cases*2,
  "three_arm_equivalent_case_count":cases*3,"directory_arm_is_not_simulated":true,
  "bootstrap_draws":draws,"bootstrap_per_study_repository":bootstrap,
  "scenario_count":len(indices),"binding_scenario_count":scenarios,
  "minimum_full_studies":projectionScenarios*minimumReplicates,"maximum_full_studies":projectionScenarios*1024,
  "projected_minimum_seconds":elapsed/float64(studies)*float64(projectionScenarios*minimumReplicates),
  "projected_maximum_seconds":elapsed/float64(studies)*float64(projectionScenarios*1024),
  "single_core_budget_seconds":timeBudget,"projection_is_not_a_runtime_guarantee":true,
  "descriptive_simultaneous_interval_alpha":alpha,"gate_one_sided_alpha":.05/3,"rows":rows,
 }
 f,e:=os.OpenFile(*out,os.O_CREATE|os.O_EXCL|os.O_WRONLY,0600);if e!=nil{panic(e)};defer f.Close()
 enc:=json.NewEncoder(f);enc.SetIndent("","  ");if e=enc.Encode(report);e!=nil{panic(e)}
}

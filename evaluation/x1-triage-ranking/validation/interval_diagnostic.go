// Synthetic-only X1 rev4 acceleration. Formulas and quantile indices match ranking.py.
package main
import("encoding/json";"os";"math";"math/rand/v2";"sort";"fmt")
type Triple [3]float64
type Group struct {Count int; Pi float64; J,B float64; Unknown bool}
type Run struct {Groups []Group; Capped bool}
type Request struct {Seed int64; Trials,B int; Method string; Population [][]Run; Repositories [][]Triple; Vertices [][][]Triple; Draws [][][]int}
type Result struct {Point float64; Interval [2]float64; Omitted int; BootstrapSD,Skew float64; Capped int}
// Match CPython 3.11: left-to-right sums, two-pass centered residuals,
// sum before repository division, identical after explicit float64 rounding.
func fit(v []Triple)(float64,float64,bool){n:=float64(len(v));t,x:=0.,0.;for _,a:=range v {t+=a[0];x+=a[1]-a[2]};if t<=0||n<2{return 0,0,false};d:=x/t;s:=0.;for _,a:=range v {e:=a[1]-a[2]-float64(d*a[0]);s+=e};center:=s/n;squares:=0.;for _,a:=range v {e:=a[1]-a[2]-float64(d*a[0]);z:=e-center;squares+=float64(z*z)};variance:=n*squares/((n-1)*(t*t));return d,variance,true}
func calculate(repos [][]Triple,draws [][][]int,method string) Result {
    r:=Result{};R:=float64(len(repos));variance:=0.
    for _,v:=range repos {d,s,ok:=fit(v);if !ok {r.Omitted=len(draws[0]);r.Interval=[2]float64{-1,1};return r};r.Point+=d;variance+=s};r.Point/=R;variance/=R*R
    dist,pivots:=[]float64{},[]float64{}
    buffers:=make([][]Triple,len(repos));for k:=range repos {buffers[k]=make([]Triple,len(draws[k][0]))}
    for b:=range draws[0] {
        d,s,valid:=0.,0.,true
        for k,v:=range repos {sample:=buffers[k];if len(sample)!=len(draws[k][b]) {sample=make([]Triple,len(draws[k][b]));buffers[k]=sample};for j,i:=range draws[k][b] {sample[j]=v[i]};x,y,ok:=fit(sample);if !ok {valid=false;break};d+=x;s+=y};d/=R;s/=R*R
        if !valid {continue};dist=append(dist,d)
        if s>0 {pivots=append(pivots,(d-r.Point)/math.Sqrt(s))} else if math.Abs(d-r.Point)<=1e-14 {pivots=append(pivots,0)} else {pivots=append(pivots,math.Copysign(math.Inf(1),d-r.Point))}
    }
    if len(dist)>1 {mean:=0.;for _,d:=range dist {mean+=d/float64(len(dist))};m2,m3:=0.,0.;for _,d:=range dist {e:=d-mean;m2+=float64(e*e);m3+=float64(float64(e*e)*e)};r.BootstrapSD=math.Sqrt(m2/float64(len(dist)-1));if m2>0 {r.Skew=(m3/float64(len(dist)))/math.Pow(m2/float64(len(dist)),1.5)}}
    values:=dist;if method=="studentized" {values=pivots};r.Omitted=len(draws[0])-len(values)
    if len(values)==0 {r.Interval=[2]float64{-1,1};return r}
    sort.Float64s(values);lo:=values[int(math.Floor(.025*float64(len(values))))];hi:=values[int(math.Ceil(.975*float64(len(values))))-1]
    if method=="studentized" {r.Interval=[2]float64{endpoint(r.Point,hi,variance),endpoint(r.Point,lo,variance)}} else {r.Interval=[2]float64{math.Max(-1,lo),math.Min(1,hi)}}
    return r
}
func endpoint(point,pivot,variance float64) float64 {if math.IsInf(pivot,1) {return -1};if math.IsInf(pivot,-1) {return 1};return math.Max(-1,math.Min(1,point-float64(pivot*math.Sqrt(variance))))}

type Vertex struct {Point,Variance float64; Values,Variances []float64; Valid []bool}
type Envelope struct {GainRange,Interval [2]float64; Omitted,Combinations int}
func envelope(q Request) Envelope {
    R:=len(q.Vertices);B:=len(q.Draws[0]);cache:=make([][]Vertex,R)
    out:=Envelope{Interval:[2]float64{1,-1}}
    for k,vertices:=range q.Vertices {
        cache[k]=make([]Vertex,len(vertices));low,high:=math.Inf(1),math.Inf(-1)
        for j,values:=range vertices {
            point,variance,ok:=fit(values);if !ok {panic("invalid original vertex denominator")}
            v:=Vertex{point,variance,make([]float64,B),make([]float64,B),make([]bool,B)}
            low=math.Min(low,point);high=math.Max(high,point)
            sample:=make([]Triple,len(q.Draws[k][0]))
            for b,indices:=range q.Draws[k] {if len(sample)!=len(indices) {sample=make([]Triple,len(indices))};for i,index:=range indices {sample[i]=values[index]};v.Values[b],v.Variances[b],v.Valid[b]=fit(sample)}
            cache[k][j]=v
        }
        out.GainRange[0]+=low;out.GainRange[1]+=high
    }
    out.GainRange[0]/=float64(R);out.GainRange[1]/=float64(R)
    chosen:=make([]Vertex,R);scratch:=make([]float64,0,B)
    var combine func(int)
    combine=func(k int) {
        if k<R {for _,v:=range cache[k] {chosen[k]=v;combine(k+1)};return}
        point,variance:=0.,0.;for _,v:=range chosen {point+=v.Point;variance+=v.Variance};point/=float64(R);variance/=float64(R*R)
        scratch=scratch[:0]
        for b:=range B {d,s,ok:=0.,0.,true;for _,v:=range chosen {if !v.Valid[b] {ok=false;break};d+=v.Values[b];s+=v.Variances[b]};if !ok {continue};d/=float64(R);s/=float64(R*R)
            if q.Method=="percentile" {scratch=append(scratch,d)} else if s>0 {scratch=append(scratch,(d-point)/math.Sqrt(s))} else if math.Abs(d-point)<=1e-14 {scratch=append(scratch,0)} else {scratch=append(scratch,math.Copysign(math.Inf(1),d-point))}
        }
        out.Omitted=max(out.Omitted,B-len(scratch));if len(scratch)==0 {return}
        sort.Float64s(scratch);n:=len(scratch);lo,hi:=scratch[int(math.Floor(.025*float64(n)))],scratch[int(math.Ceil(.975*float64(n)))-1]
        if q.Method=="studentized" {lo,hi=endpoint(point,hi,variance),endpoint(point,lo,variance)}
        out.Interval[0]=math.Min(out.Interval[0],math.Max(-1,lo));out.Interval[1]=math.Max(out.Interval[1],math.Min(1,hi));out.Combinations++
    }
    combine(0);if out.Combinations==0 {out.Interval=[2]float64{-1,1}};return out
}
func main(){
    var q Request;if err:=json.NewDecoder(os.Stdin).Decode(&q);err!=nil {panic(err)}
    if q.Method!="percentile"&&q.Method!="studentized" {panic("invalid method")}
    encoder:=json.NewEncoder(os.Stdout)
    if len(q.Vertices)>0 {if err:=encoder.Encode(envelope(q));err!=nil {panic(err)};return}
    if len(q.Repositories)>0 {if err:=encoder.Encode(calculate(q.Repositories,q.Draws,q.Method));err!=nil{panic(err)};return}
    if q.Trials<2000||q.B!=1999 {panic("diagnostic requires trials>=2000 B1999")}
    rng:=rand.New(rand.NewPCG(uint64(q.Seed),uint64(q.Seed)^0x9e3779b97f4a7c15))
    missing:=false;for _,pop:=range q.Population {for _,run:=range pop {for _,g:=range run.Groups {missing=missing||g.Unknown}}}
    for range q.Trials {
        repos:=make([][]Triple,len(q.Population));draws:=make([][][]int,len(repos))
        vertices:=make([][][]Triple,len(repos));capped:=0
        for k,pop:=range q.Population {
            chosen:=rng.Perm(len(pop))[:20];repos[k]=make([]Triple,20)
            if missing {vertices[k]=make([][]Triple,8);for v:=range 8 {vertices[k][v]=make([]Triple,20)}}
            for j,i:=range chosen {
                run:=pop[i];if run.Capped {capped++}
                for _,g:=range run.Groups {
                    count:=0;for range g.Count {if rng.Float64()<g.Pi {count++}}
                    w:=float64(count)/g.Pi
                    if !g.Unknown {repos[k][j][0]+=w;repos[k][j][1]+=float64(w*g.J);repos[k][j][2]+=float64(w*g.B)}
                    if missing {for v:=range 8 {
                        d:=int(g.J-g.B)+1
                        if !g.Unknown||(v>>uint(2-d))&1==1 {vertices[k][v][j][0]+=w;vertices[k][v][j][1]+=float64(w*g.J);vertices[k][v][j][2]+=float64(w*g.B)}
                    }}
                }
            }
            draws[k]=make([][]int,q.B);for b:=range draws[k] {draws[k][b]=make([]int,19);for j:=range draws[k][b] {draws[k][b][j]=rng.IntN(20)}}
        }
        if missing {if err:=encoder.Encode(envelope(Request{Method:q.Method,Vertices:vertices,Draws:draws}));err!=nil {panic(err)}} else {
            r:=calculate(repos,draws,q.Method);r.Capped=capped;if err:=encoder.Encode(r);err!=nil {fmt.Fprintln(os.Stderr,err);os.Exit(1)}
        }
    }
}

package decisions

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"scout/internal/domain"
)

const (
	testModel = "typesafe/jev-1.14-20261014"
	testKey   = "sk-or-v1-0123456789abcdefSECRET"
)

// fake is an httptest stand-in for OpenRouter's alpha/decisions route.
type fake struct {
	t        *testing.T
	server   *httptest.Server
	mu       sync.Mutex
	bodies   []string
	handler  func(n int, body map[string]any, w http.ResponseWriter)
	requests int
}

func newFake(t *testing.T, handler func(n int, body map[string]any, w http.ResponseWriter)) *fake {
	f := &fake{t: t, handler: handler}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if r.Header.Get("Authorization") != "Bearer "+testKey {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var body map[string]any
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Errorf("request is not JSON: %v", err)
		}
		f.mu.Lock()
		f.requests++
		n := f.requests
		f.bodies = append(f.bodies, string(raw))
		f.mu.Unlock()
		f.handler(n, body, w)
	}))
	t.Cleanup(f.server.Close)
	return f
}

func (f *fake) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.requests
}

func answer(w http.ResponseWriter, model string, p float64) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id":"gen-1","model":%q,"answers":{"q1":{"type":"noul","noul":%v}},"usage":{"input_tokens":10,"output_tokens":2,"cost":0.00005}}`, model, p)
}

// idOf reads the unit id the tests place at state.id.
func idOf(body map[string]any) string {
	state, _ := body["state"].(map[string]any)
	id, _ := state["id"].(string)
	return id
}

func client(f *fake) (*Client, *[]time.Duration) {
	var sleeps []time.Duration
	return &Client{Endpoint: f.server.URL, Key: testKey, Sleep: func(_ context.Context, d time.Duration) error { sleeps = append(sleeps, d); return nil }, Now: func() time.Time { return time.Unix(1_000, 0) }}, &sleeps
}

func unit(id string, baseline float64) Unit {
	return Unit{Family: FamilyTriage, Key: id, IssueID: id, Questions: []Question{TriageRealDefect}, State: json.RawMessage(fmt.Sprintf(`{"id":%q}`, id)), Baseline: baseline}
}

type memorySink struct {
	mu   sync.Mutex
	rows []domain.Annotation
}

func (m *memorySink) RecordAnnotation(_ context.Context, a domain.Annotation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rows = append(m.rows, a)
	return a.Validate()
}

func TestVersionKeyIsStableAndChangesOnEveryKeyedField(t *testing.T) {
	base := Question{
		ID: "q.one", Kind: KindChoice, Instructions: "pick", Criteria: map[string]string{"a": "first", "b": "second"},
		StateShape: []string{"state.x", "state.y"}, Order: OptionOrderRule{Rule: OrderPermutation},
	}
	want := base.Version(testModel)
	if !strings.HasPrefix(want, "q.one@") || len(want) != len("q.one@")+64 {
		t.Fatalf("version %q is not question_id@sha256", want)
	}
	if base.Version(testModel) != want {
		t.Fatal("version is not stable")
	}
	reordered := base
	reordered.StateShape = []string{"state.y", "state.x"}
	if reordered.Version(testModel) != want {
		t.Fatal("state shape is a set; order must not change the version")
	}
	notKeyed := base
	notKeyed.Positive = "a"
	if notKeyed.Version(testModel) != want {
		t.Fatal("Positive is a code-side reading and must not change the version")
	}
	if base.RealisedOrder("unit-1")[0] == "" || len(base.RealisedOrder("unit-1")) != 2 {
		t.Fatal("realised order must list every option")
	}
	mutations := map[string]func(Question) (Question, string){
		"instructions": func(q Question) (Question, string) { q.Instructions = "pick carefully"; return q, testModel },
		"criteria": func(q Question) (Question, string) {
			q.Criteria = map[string]string{"a": "first", "b": "other"}
			return q, testModel
		},
		"state shape": func(q Question) (Question, string) {
			q.StateShape = []string{"state.x", "state.y", "state.z"}
			return q, testModel
		},
		"option order": func(q Question) (Question, string) {
			q.Order = OptionOrderRule{Rule: OrderFixed, Options: []string{"a", "b"}}
			return q, testModel
		},
		"resolved model": func(q Question) (Question, string) { return q, "typesafe/jev-1.15-20270101" },
	}
	seen := map[string]string{want: "base"}
	for name, mutate := range mutations {
		q, model := mutate(base)
		got := q.Version(model)
		if prev, dup := seen[got]; dup {
			t.Fatalf("%s produced the same version as %s", name, prev)
		}
		seen[got] = name
	}
	other := base
	other.ID = "q.two"
	if !strings.HasPrefix(other.Version(testModel), "q.two@") || other.Version(testModel)[6:] != want[6:] {
		t.Fatal("the id prefixes the key but is not part of the hashed payload")
	}
}

func TestChoicePermutationIsSeededByUnitAndStoredNotKeyed(t *testing.T) {
	options := map[string]string{}
	for _, o := range []string{"a", "b", "c", "d", "e", "f"} {
		options[o] = "criterion " + o
	}
	q := Question{ID: "choice.q", Kind: KindChoice, Instructions: "pick", Criteria: options, StateShape: []string{"state.x"}, Order: OptionOrderRule{Rule: OrderPermutation}, Positive: "a"}
	if err := q.Validate(); err != nil {
		t.Fatal(err)
	}
	first := q.RealisedOrder("unit-1")
	if strings.Join(first, "") != strings.Join(q.RealisedOrder("unit-1"), "") {
		t.Fatal("same unit must get the same order")
	}
	different := false
	for i := range 20 {
		if strings.Join(q.RealisedOrder(fmt.Sprintf("unit-%d", i)), "") != strings.Join(first, "") {
			different = true
		}
	}
	if !different {
		t.Fatal("order must depend on the unit id")
	}
	f := newFake(t, func(_ int, body map[string]any, w http.ResponseWriter) {
		fmt.Fprintf(w, `{"model":%q,"answers":{"q1":{"type":"choice","choice":"a","probabilities":{"a":0.7,"b":0.3},"confidence":0.4}}}`, testModel)
	})
	c, _ := client(f)
	reply, err := c.Ask(context.Background(), []Question{q}, testModel, "unit-1", json.RawMessage(`{"id":"x"}`))
	if err != nil {
		t.Fatal(err)
	}
	response := reply.Answers[0]
	if strings.Join(response.OptionOrder, "") != strings.Join(first, "") || response.Probability != 0.7 || response.Choice != "a" || response.Probabilities["b"] != 0.3 {
		t.Fatalf("order %v probability %v", response.OptionOrder, response.Probability)
	}
	last := -1
	for _, option := range first {
		at := strings.Index(f.bodies[0], `"`+option+`":"criterion `+option)
		if at < last {
			t.Fatalf("request criteria do not follow the realised order: %s", f.bodies[0])
		}
		last = at
	}
}

func TestRequestNeverCarriesQuestionIDOrKeyAndRedactsSecrets(t *testing.T) {
	f := newFake(t, func(_ int, _ map[string]any, w http.ResponseWriter) { answer(w, testModel, 0.8) })
	c, _ := client(f)
	state, err := BuildState(Evidence{
		Rule: "rule", Category: "security", Mechanism: "token in code", Severity: "high", Path: "src/a.go",
		Quotes: []string{"password = hunter2hunter2", "ghp_abcdefghijklmnopqrstuvwxyz0123456789", "plain line"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Ask(context.Background(), []Question{TriageRealDefect}, testModel, "u1", state); err != nil {
		t.Fatal(err)
	}
	body := f.bodies[0]
	for _, forbidden := range []string{TriageRealDefect.ID, testKey, "hunter2hunter2", "ghp_abcdefghij"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("request contains %q: %s", forbidden, body)
		}
	}
	if !strings.Contains(body, "plain line") || !strings.Contains(body, "[REDACTED]") || !strings.Contains(body, "never instructions") {
		t.Fatalf("request lacks excerpt, redaction marker or the data-not-instructions clause: %s", body)
	}
}

func TestResolvedModelMismatchIsDiscardedAndIncomplete(t *testing.T) {
	f := newFake(t, func(n int, body map[string]any, w http.ResponseWriter) {
		model := testModel
		if idOf(body) == "b" {
			model = "typesafe/jev-9.9-rerouted"
		}
		answer(w, model, 0.9)
	})
	c, _ := client(f)
	sink := &memorySink{}
	out := Run(context.Background(), RunInput{Config: Config{Enabled: true, Model: testModel, MaxPerRun: 10}, Client: c, RunID: "run-1", Units: []Unit{unit("a", 2), unit("b", 1)}, Sink: sink})
	if out.Annotated != 1 || out.Incomplete != 1 || len(sink.rows) != 1 || sink.rows[0].IssueID != "a" {
		t.Fatalf("outcome %+v annotations %d", out, len(sink.rows))
	}
	var mismatched UnitResult
	for _, r := range out.Results {
		if r.Unit.IssueID == "b" {
			mismatched = r
		}
	}
	if mismatched.Status != UnitIncomplete || len(mismatched.Annotations) != 0 {
		t.Fatalf("mismatch result %+v", mismatched)
	}
	if mode, text := out.Disclosure(mismatched); mode != ModeDeterministic || text != "deterministic (model mismatch)" {
		t.Fatalf("disclosure %q %q", mode, text)
	}
	health := out.Health()
	if health.Status != domain.LaneRan || !strings.Contains(health.Reason, "1 incomplete (model mismatch)") || health.Required {
		t.Fatalf("health %+v", health)
	}
	// When every call is mismatched the lane has failed, never ran.
	f2 := newFake(t, func(_ int, _ map[string]any, w http.ResponseWriter) { answer(w, "other", 0.9) })
	c2, _ := client(f2)
	all := Run(context.Background(), RunInput{Config: Config{Enabled: true, Model: testModel}, Client: c2, RunID: "r", Units: []Unit{unit("a", 1)}})
	if all.Health().Status != domain.LaneFailed {
		t.Fatalf("health %+v", all.Health())
	}
}

func TestNonRetryableStatusesAreNotRetried(t *testing.T) {
	for _, status := range []int{400, 422} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			f := newFake(t, func(_ int, _ map[string]any, w http.ResponseWriter) {
				w.WriteHeader(status)
				io.WriteString(w, `{"error":"max_tokens_exceeded"}`)
			})
			c, sleeps := client(f)
			_, err := c.Ask(context.Background(), []Question{TriageRealDefect}, testModel, "u", json.RawMessage(`{}`))
			failure, ok := err.(*Failure)
			if !ok || failure.Kind != FailNonRetryable || failure.Status != status {
				t.Fatalf("error %v", err)
			}
			if f.count() != 1 || len(*sleeps) != 0 {
				t.Fatalf("requests %d sleeps %v: a non-retryable status must not be retried", f.count(), *sleeps)
			}
			if status == 422 && !strings.Contains(failure.Reason, "max_tokens_exceeded") {
				t.Fatalf("reason %q", failure.Reason)
			}
		})
	}
}

func TestRateLimitsRetryWithBackoffHonouringRetryAfter(t *testing.T) {
	f := newFake(t, func(n int, _ map[string]any, w http.ResponseWriter) {
		switch n {
		case 1:
			w.Header().Set("Retry-After", "7")
			w.WriteHeader(429)
		case 2:
			w.WriteHeader(529)
		default:
			answer(w, testModel, 0.5)
		}
	})
	c, sleeps := client(f)
	response, err := c.Ask(context.Background(), []Question{TriageRealDefect}, testModel, "u", json.RawMessage(`{}`))
	if err != nil || response.Attempts != 3 {
		t.Fatalf("response %+v err %v", response, err)
	}
	if got := fmt.Sprint(*sleeps); got != fmt.Sprint([]time.Duration{7 * time.Second, 2 * time.Second}) {
		t.Fatalf("fake-clock sleeps %v: want Retry-After 7s then exponential 2s", got)
	}

	// An HTTP-date Retry-After is relative to the (fake) clock, and every wait is bounded.
	g := newFake(t, func(n int, _ map[string]any, w http.ResponseWriter) {
		w.Header().Set("Retry-After", time.Unix(1_030, 0).UTC().Format(http.TimeFormat))
		w.WriteHeader(429)
	})
	cg, sleepsG := client(g)
	cg.MaxAttempts, cg.MaxWait = 3, 10*time.Second
	_, err = cg.Ask(context.Background(), []Question{TriageRealDefect}, testModel, "u", json.RawMessage(`{}`))
	failure, ok := err.(*Failure)
	if !ok || failure.Kind != FailRateLimited || g.count() != 3 {
		t.Fatalf("error %v after %d requests: backoff must stop at MaxAttempts", err, g.count())
	}
	if fmt.Sprint(*sleepsG) != fmt.Sprint([]time.Duration{10 * time.Second, 10 * time.Second}) {
		t.Fatalf("sleeps %v", *sleepsG)
	}
}

func TestOtherFailuresLeaveNoAnnotationAndNeverLeakTheKey(t *testing.T) {
	cases := map[string]func(http.ResponseWriter){
		"server error":   func(w http.ResponseWriter) { w.WriteHeader(500) },
		"not json":       func(w http.ResponseWriter) { io.WriteString(w, "<html>") },
		"missing answer": func(w http.ResponseWriter) { fmt.Fprintf(w, `{"model":%q,"answers":{}}`, testModel) },
		"out of range":   func(w http.ResponseWriter) { answer(w, testModel, 1.5) },
		"wrong type": func(w http.ResponseWriter) {
			fmt.Fprintf(w, `{"model":%q,"answers":{"q1":{"type":"choice","choice":"x"}}}`, testModel)
		},
	}
	for name, respond := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFake(t, func(_ int, _ map[string]any, w http.ResponseWriter) { respond(w) })
			c, _ := client(f)
			sink := &memorySink{}
			out := Run(context.Background(), RunInput{Config: Config{Model: testModel}, Client: c, RunID: "r", Units: []Unit{unit("a", 1)}, Sink: sink})
			if out.Failed != 1 || len(sink.rows) != 0 || len(out.Results[0].Annotations) != 0 {
				t.Fatalf("outcome %+v", out)
			}
			if mode, text := out.Disclosure(out.Results[0]); mode != ModeDeterministic || !strings.HasPrefix(text, "deterministic (") {
				t.Fatalf("disclosure %q", text)
			}
			if strings.Contains(out.Results[0].Reason, testKey) {
				t.Fatal("failure reason leaks the key")
			}
			if h := out.Health(); h.Status != domain.LaneFailed {
				t.Fatalf("health %+v", h)
			}
		})
	}
	// A dead endpoint is a network failure whose text never carries the key.
	dead := newFake(t, func(int, map[string]any, http.ResponseWriter) {})
	c, _ := client(dead)
	dead.server.Close()
	_, err := c.Ask(context.Background(), []Question{TriageRealDefect}, testModel, "u", json.RawMessage(`{}`))
	if failure, ok := err.(*Failure); !ok || failure.Kind != FailNetwork || strings.Contains(err.Error(), testKey) {
		t.Fatalf("error %v", err)
	}
}

func TestCapStopsAtMaxPerRunInDeterministicBaselineOrderAndCountsPartial(t *testing.T) {
	var asked []string
	var mu sync.Mutex
	f := newFake(t, func(_ int, body map[string]any, w http.ResponseWriter) {
		mu.Lock()
		asked = append(asked, idOf(body))
		mu.Unlock()
		answer(w, testModel, 0.5)
	})
	c, _ := client(f)
	units := []Unit{unit("e", 1), unit("c", 5), unit("b", 5), unit("a", 2), unit("d", 3)}
	out := Run(context.Background(), RunInput{Config: Config{Model: testModel, MaxPerRun: 3}, Client: c, RunID: "r", Units: units})
	if strings.Join(asked, ",") != "b,c,d" {
		t.Fatalf("asked %v: want baseline order with ties by issue id", asked)
	}
	if out.Total != 5 || out.Attempted != 3 || out.Annotated != 3 || out.Capped != 2 {
		t.Fatalf("outcome %+v", out)
	}
	health := out.Health()
	if health.Status != domain.LaneRan || health.Reason != "partial (3/5)" || health.Units != 5 || health.Required {
		t.Fatalf("health %+v", health)
	}
	var capped UnitResult
	for _, r := range out.Results {
		if r.Status == UnitCapped {
			capped = r
		}
	}
	if mode, text := out.Disclosure(capped); mode != ModePartial || text != "partial (3/5)" {
		t.Fatalf("disclosure %q %q", mode, text)
	}
}

func TestCapOrderAcrossFamiliesFollowsDesign(t *testing.T) {
	var asked []string
	var mu sync.Mutex
	f := newFake(t, func(_ int, body map[string]any, w http.ResponseWriter) {
		mu.Lock()
		asked = append(asked, idOf(body))
		mu.Unlock()
		answer(w, testModel, 0.5)
	})
	c, _ := client(f)
	mk := func(family Family, key string, baseline float64) Unit {
		u := unit(key, baseline)
		u.Family, u.Key = family, key
		encoded, _ := json.Marshal(map[string]string{"id": key})
		u.State = encoded
		return u
	}
	units := []Unit{
		mk(FamilyUnowned, "pkg/z.go", 0), mk(FamilyCommit, "bbb", 0), mk(FamilyCommit, "aaa", 0),
		mk(FamilyRubric, RubricKey("pkg/b.go", 10), 0), mk(FamilyRubric, RubricKey("pkg/a.go", 100), 0), mk(FamilyRubric, RubricKey("pkg/a.go", 9), 0),
		mk(FamilyTriage, "issue-low", 1), mk(FamilyTriage, "issue-high", 9),
	}
	Run(context.Background(), RunInput{Config: Config{Model: testModel, MaxPerRun: 100}, Client: c, RunID: "r", Units: units})
	want := []string{"issue-high", "issue-low", RubricKey("pkg/a.go", 9), RubricKey("pkg/a.go", 100), RubricKey("pkg/b.go", 10), "aaa", "bbb", "pkg/z.go"}
	if fmt.Sprintf("%q", asked) != fmt.Sprintf("%q", want) {
		t.Fatalf("asked %q\nwant  %q", asked, want)
	}
}

func TestEligibilityReasons(t *testing.T) {
	ok := Eligibility{Enabled: true, HasKey: true, EgressAllowed: true}
	if reason, allowed := ok.Skip(); !allowed || reason != "" {
		t.Fatal("fully consented run must be allowed")
	}
	cases := []struct {
		mutate func(*Eligibility)
		reason string
	}{
		{func(e *Eligibility) { e.Enabled = false }, "decisions not enabled"},
		{func(e *Eligibility) { e.Fork = true }, "fork"},
		{func(e *Eligibility) { e.Fork = true; e.Enabled = false }, "fork"},
		{func(e *Eligibility) { e.EgressAllowed = false }, "egress not allowed"},
		{func(e *Eligibility) { e.HasKey = false }, "no jev key"},
	}
	for _, tc := range cases {
		e := ok
		tc.mutate(&e)
		if reason, allowed := e.Skip(); allowed || reason != tc.reason {
			t.Fatalf("got %q allowed=%v, want %q", reason, allowed, tc.reason)
		}
	}
	if h := SkippedHealth("fork"); h.Status != domain.LaneSkipped || h.Reason != "fork" || h.Required {
		t.Fatalf("health %+v", h)
	}
	for _, alias := range []string{"", "jev-latest", "typesafe/jev-preview"} {
		if (Config{Model: alias}).Validate() == nil {
			t.Fatalf("model %q must be rejected", alias)
		}
	}
	if err := (Config{Model: testModel}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRankFormulaVersionAndOrdinalReranking(t *testing.T) {
	formula := LaunchFormula()
	version := formula.Version(testModel)
	if version != formula.Version(testModel) || version == formula.Version("typesafe/jev-1.15-20270101") {
		t.Fatal("rank-formula version must be stable and follow the constituent question versions")
	}
	heavier := LaunchFormula()
	heavier.Weights = map[string]float64{TriageRealDefect.ID: 2}
	if heavier.Version(testModel) == version {
		t.Fatal("changing a weight must change the formula version")
	}
	items := []Item{
		{IssueID: "a", Baseline: 4, Answers: map[string]float64{TriageRealDefect.ID: 0.2}},
		{IssueID: "b", Baseline: 1, Answers: map[string]float64{TriageRealDefect.ID: 0.9}},
		{IssueID: "c", Baseline: 9},
		{IssueID: "d", Baseline: 3},
	}
	ranked := Rank(formula, items)
	order := ""
	for _, r := range ranked {
		order += r.IssueID
	}
	if order != "bacd" || ranked[0].Position != 1 || ranked[3].Position != 4 {
		t.Fatalf("order %s: annotated by P, then unannotated by baseline", order)
	}
	// Re-ranking stored answers under a new weight needs no model call: Rank is pure.
	if again := Rank(heavier, items); again[0].Score != 1.8 {
		t.Fatalf("score %v", again[0].Score)
	}
}

type memoryAudit struct {
	rows map[string]domain.AuditDraw
}

func (m *memoryAudit) AuditDraw(_ context.Context, issue, formula, sampler string) (domain.AuditDraw, bool, error) {
	d, ok := m.rows[issue+formula+sampler]
	return d, ok, nil
}

func (m *memoryAudit) RecordAuditDraw(_ context.Context, d domain.AuditDraw) (domain.AuditDraw, error) {
	key := d.IssueID + d.FormulaVersion + d.SamplerVersion
	if existing, ok := m.rows[key]; ok {
		return existing, nil
	}
	m.rows[key] = d
	return d, nil
}

func TestAuditDrawStoresInclusionProbabilityAndSamplerVersion(t *testing.T) {
	var units []AuditUnit
	for i := 1; i <= 20; i++ {
		units = append(units, AuditUnit{IssueID: fmt.Sprintf("i%02d", i), Position: i, Annotated: i <= 16, Critical: i == 18 || i == 3})
	}
	assigned := map[string]Assignment{}
	for _, a := range Assign(units, ProvisionalFraction) {
		assigned[a.IssueID] = a
	}
	if assigned["i18"].Stratum != StratumCertainty || assigned["i18"].Probability != 1 {
		t.Fatalf("critical finding below the floor is certainty: %+v", assigned["i18"])
	}
	if assigned["i03"].Stratum != StratumTop {
		t.Fatalf("critical finding above the floor stays in its half: %+v", assigned["i03"])
	}
	if assigned["i12"].Stratum != StratumBottom || assigned["i19"].Stratum != StratumUnannotated || assigned["i19"].Probability <= 0 || assigned["i19"].Probability > 1 {
		t.Fatalf("strata %+v %+v", assigned["i12"], assigned["i19"])
	}
	// Expected audits are the fraction of the run, split equally over the strata.
	expected := 0.0
	for _, a := range assigned {
		if a.Stratum != StratumCertainty {
			expected += a.Probability
		}
	}
	if expected < 0.9 || expected > 1.1 {
		t.Fatalf("expected audits outside the certainty stratum = %v, want ~1 (5%% of 20)", expected)
	}

	store := &memoryAudit{rows: map[string]domain.AuditDraw{}}
	calls := 0
	rand := func() float64 { calls++; return 0.0 } // selects every unit with p > 0
	draws, err := DrawAudit(context.Background(), store, "run-1", "formula@1", units, ProvisionalFraction, rand, time.Unix(1, 0))
	if err != nil || len(draws) != 20 || calls != 20 {
		t.Fatalf("draws %d calls %d err %v", len(draws), calls, err)
	}
	for _, d := range draws {
		a := assigned[d.IssueID]
		if d.SamplerVersion != SamplerVersion || d.InclusionProbability != a.Probability || d.Stratum != a.Stratum || !d.Drawn || d.FormulaVersion != "formula@1" {
			t.Fatalf("stored draw %+v does not record the realised probability and sampler version", d)
		}
	}
	// A unit is drawn once, at its first eligible run; later runs keep the stored draw.
	again, err := DrawAudit(context.Background(), store, "run-2", "formula@1", units, 0.5, func() float64 { t.Fatal("redrawn"); return 1 }, time.Unix(2, 0))
	if err != nil || again[0].RunID != "run-1" || again[0].InclusionProbability != assigned[again[0].IssueID].Probability {
		t.Fatalf("redraw %+v err %v", again[0], err)
	}
	// A new formula version is a new sampling unit.
	third, err := DrawAudit(context.Background(), store, "run-3", "formula@2", units[:2], ProvisionalFraction, func() float64 { return 0.999 }, time.Unix(3, 0))
	if err != nil || third[0].RunID != "run-3" || third[0].Drawn && third[0].InclusionProbability < 1 {
		t.Fatalf("new formula draw %+v err %v", third[0], err)
	}
}

// issues must not import decisions (an annotation can only change a rank), and
// no Jev-quality metric is computed anywhere in the adapter.
func TestAdapterBoundariesAndNoQualityMetrics(t *testing.T) {
	fset := token.NewFileSet()
	check := func(dir string, forbidden string) {
		files, _ := filepath.Glob(filepath.Join(dir, "*.go"))
		for _, file := range files {
			parsed, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatal(err)
			}
			for _, imp := range parsed.Imports {
				if strings.Contains(imp.Path.Value, forbidden) {
					t.Fatalf("%s imports %s", file, imp.Path.Value)
				}
			}
		}
	}
	check("../issues", "scout/internal/decisions")
	check(".", "scout/internal/issues")

	metric := regexp.MustCompile(`(?i)precision|recall|accuracy|auroc|calibrat|brier|f1score|agreement|flip`)
	files, _ := filepath.Glob("*.go")
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(parsed, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && metric.MatchString(id.Name) {
				t.Errorf("%s declares or uses %q: no Jev-quality metric may be computed before X1", file, id.Name)
			}
			return true
		})
	}
}

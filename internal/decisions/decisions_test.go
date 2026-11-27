//go:build boundary

package decisions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"scout/internal/domain"
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
	writeRoute(w, `{"id":"gen-1","model":%q,"answers":{"q1":{"type":"noul","noul":%v}},"usage":{"input_tokens":10,"output_tokens":2,"cost":0.00005}}`, model, p)
}

func writeRoute(w io.Writer, format string, args ...any) {
	if _, err := fmt.Fprintf(w, format, args...); err != nil {
		panic(fmt.Sprintf("fake decisions route could not write its reply: %v", err))
	}
}

// idOf reads the unit id the tests place at state.id.
func idOf(body map[string]any) string {
	state, _ := body["state"].(map[string]any)
	id, _ := state["id"].(string)
	return id
}

func client(f *fake) *Client {
	return &Client{Endpoint: f.server.URL, Key: testKey, Now: func() time.Time { return time.Unix(1_000, 0) }, MaxWait: time.Millisecond}
}

func unit(id string, baseline float64) Unit {
	return Unit{Key: id, IssueID: id, Questions: []Question{TriageRealDefect}, State: json.RawMessage(fmt.Sprintf(`{"id":%q}`, id)), Baseline: baseline}
}

const (
	FamilyRubric Family = iota + 1
	FamilyCommit
	FamilyUnowned
)

// RubricKey is the stable key of a rubric function unit: changed-file path,
// then function start line.
func RubricKey(path string, startLine int) string {
	return fmt.Sprintf("%s\x00%010d", path, startLine)
}

type memorySink struct {
	mu   sync.Mutex
	rows []domain.Annotation
	fail error
}

func (m *memorySink) RecordAnnotations(_ context.Context, annotations []domain.Annotation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.fail != nil {
		return m.fail
	}
	for _, a := range annotations {
		if err := a.Validate(); err != nil {
			return err
		}
	}
	m.rows = append(m.rows, annotations...)
	return nil
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
	first := assertOrderSeededByUnit(t, q)
	f := newFake(t, func(_ int, body map[string]any, w http.ResponseWriter) {
		writeRoute(w, `{"model":%q,"answers":{"q1":{"type":"choice","choice":"a","probabilities":{"a":0.7,"b":0.3},"confidence":0.4}}}`, testModel)
	})
	c := client(f)
	reply, err := c.Ask(context.Background(), []Question{q}, testModel, "unit-1", json.RawMessage(`{"id":"x"}`))
	if err != nil {
		t.Fatal(err)
	}
	response := reply.Answers[0]
	if strings.Join(response.OptionOrder, "") != strings.Join(first, "") || response.Probability != 0.7 || response.Choice != "a" || response.Probabilities["b"] != 0.3 {
		t.Fatalf("order %v probability %v", response.OptionOrder, response.Probability)
	}
	assertCriteriaFollowOrder(t, f.bodies[0], first)
}

func assertOrderSeededByUnit(t *testing.T, q Question) []string {
	t.Helper()
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
	return first
}

func assertCriteriaFollowOrder(t *testing.T, body string, order []string) {
	t.Helper()
	last := -1
	for _, option := range order {
		at := strings.Index(body, `"`+option+`":"criterion `+option)
		if at < last {
			t.Fatalf("request criteria do not follow the realised order: %s", body)
		}
		last = at
	}
}

func TestRequestNeverCarriesQuestionIDOrKeyAndRedactsSecrets(t *testing.T) {
	f := newFake(t, func(_ int, _ map[string]any, w http.ResponseWriter) { answer(w, testModel, 0.8) })
	c := client(f)
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
	c := client(f)
	sink := &memorySink{}
	out := Run(context.Background(), RunInput{Config: testConfig(), Client: c, RunID: "run-1", Units: []Unit{unit("a", 2), unit("b", 1)}, Sink: sink})
	if out.Annotated != 1 || out.Incomplete != 1 || len(sink.rows) != 1 || sink.rows[0].IssueID != "a" {
		t.Fatalf("outcome %+v annotations %d", out, len(sink.rows))
	}
	assertMismatchDisclosed(t, out)
	// When every call is mismatched the lane has failed, never ran.
	f2 := newFake(t, func(_ int, _ map[string]any, w http.ResponseWriter) { answer(w, "other", 0.9) })
	c2 := client(f2)
	all := Run(context.Background(), RunInput{Config: testConfig(), Client: c2, RunID: "r", Units: []Unit{unit("a", 1)}})
	if all.Health().Status != domain.LaneFailed {
		t.Fatalf("health %+v", all.Health())
	}
}

func assertMismatchDisclosed(t *testing.T, out Outcome) {
	t.Helper()
	var mismatched UnitResult
	for _, r := range out.Results {
		if r.Unit.IssueID == "b" {
			mismatched = r
		}
	}
	if mismatched.Status != UnitIncomplete || len(mismatched.Annotations) != 0 {
		t.Fatalf("mismatch result %+v", mismatched)
	}
	if mode, text := out.Disclosure(mismatched); mode != domain.RankingModeDeterministic || text != "deterministic (model mismatch)" {
		t.Fatalf("disclosure %q %q", mode, text)
	}
	health := out.Health()
	if health.Status != domain.LaneRan || !strings.Contains(health.Reason, "1 incomplete (model mismatch)") || health.Required {
		t.Fatalf("health %+v", health)
	}
}

func TestNonRetryableStatusesAreNotRetried(t *testing.T) {
	for _, status := range []int{400, 422} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			f := newFake(t, func(_ int, _ map[string]any, w http.ResponseWriter) {
				w.WriteHeader(status)
				writeRoute(w, "%s", `{"error":"max_tokens_exceeded"}`)
			})
			c := client(f)
			_, err := c.Ask(context.Background(), []Question{TriageRealDefect}, testModel, "u", json.RawMessage(`{}`))
			failure, ok := err.(*Failure)
			if !ok || failure.Kind != FailNonRetryable || failure.Status != status {
				t.Fatalf("error %v", err)
			}
			if f.count() != 1 {
				t.Fatalf("requests %d: a non-retryable status must not be retried", f.count())
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
	c := client(f)
	response, err := c.Ask(context.Background(), []Question{TriageRealDefect}, testModel, "u", json.RawMessage(`{}`))
	if err != nil || response.Attempts != 3 {
		t.Fatalf("response %+v err %v", response, err)
	}
	waits := backoffWaits(&Client{Now: c.Now}, []string{"7", ""})
	if got := fmt.Sprint(waits); got != fmt.Sprint([]time.Duration{7 * time.Second, 2 * time.Second}) {
		t.Fatalf("fake-clock waits %v: want Retry-After 7s then exponential 2s", got)
	}

	// An HTTP-date Retry-After is relative to the (fake) clock, and every wait is bounded.
	g := newFake(t, func(n int, _ map[string]any, w http.ResponseWriter) {
		w.Header().Set("Retry-After", time.Unix(1_030, 0).UTC().Format(http.TimeFormat))
		w.WriteHeader(429)
	})
	cg := client(g)
	cg.MaxAttempts = 3
	_, err = cg.Ask(context.Background(), []Question{TriageRealDefect}, testModel, "u", json.RawMessage(`{}`))
	failure, ok := err.(*Failure)
	if !ok || failure.Kind != FailRateLimited || g.count() != 3 {
		t.Fatalf("error %v after %d requests: backoff must stop at MaxAttempts", err, g.count())
	}
	dated := time.Unix(1_030, 0).UTC().Format(http.TimeFormat)
	waitsG := backoffWaits(&Client{Now: cg.Now, MaxWait: 10 * time.Second}, []string{dated, dated})
	if fmt.Sprint(waitsG) != fmt.Sprint([]time.Duration{10 * time.Second, 10 * time.Second}) {
		t.Fatalf("waits %v", waitsG)
	}
}

func backoffWaits(c *Client, retryAfter []string) []time.Duration {
	policy := c.retryPolicy()
	waits := make([]time.Duration, 0, len(retryAfter))
	for index, value := range retryAfter {
		header := http.Header{}
		if value != "" {
			header.Set("Retry-After", value)
		}
		waits = append(waits, c.backoffWait(policy, index+1, header))
	}
	return waits
}

func TestOtherFailuresLeaveNoAnnotationAndNeverLeakTheKey(t *testing.T) {
	cases := map[string]func(http.ResponseWriter){
		"server error":   func(w http.ResponseWriter) { w.WriteHeader(500) },
		"not json":       func(w http.ResponseWriter) { writeRoute(w, "%s", "<html>") },
		"missing answer": func(w http.ResponseWriter) { writeRoute(w, `{"model":%q,"answers":{}}`, testModel) },
		"out of range":   func(w http.ResponseWriter) { answer(w, testModel, 1.5) },
		"wrong type": func(w http.ResponseWriter) {
			writeRoute(w, `{"model":%q,"answers":{"q1":{"type":"choice","choice":"x"}}}`, testModel)
		},
	}
	for name, respond := range cases {
		t.Run(name, func(t *testing.T) { assertFailedReplyLeavesNoAnnotation(t, respond) })
	}
	// A dead endpoint is a network failure whose text never carries the key.
	dead := newFake(t, func(int, map[string]any, http.ResponseWriter) {})
	c := client(dead)
	dead.server.Close()
	_, err := c.Ask(context.Background(), []Question{TriageRealDefect}, testModel, "u", json.RawMessage(`{}`))
	if failure, ok := err.(*Failure); !ok || failure.Kind != FailNetwork || strings.Contains(err.Error(), testKey) {
		t.Fatalf("error %v", err)
	}
}

func assertFailedReplyLeavesNoAnnotation(t *testing.T, respond func(http.ResponseWriter)) {
	t.Helper()
	f := newFake(t, func(_ int, _ map[string]any, w http.ResponseWriter) { respond(w) })
	c := client(f)
	sink := &memorySink{}
	out := Run(context.Background(), RunInput{Config: testConfig(), Client: c, RunID: "r", Units: []Unit{unit("a", 1)}, Sink: sink})
	if out.Failed != 1 || len(sink.rows) != 0 || len(out.Results[0].Annotations) != 0 {
		t.Fatalf("outcome %+v", out)
	}
	if mode, text := out.Disclosure(out.Results[0]); mode != domain.RankingModeDeterministic || !strings.HasPrefix(text, "deterministic (") {
		t.Fatalf("disclosure %q", text)
	}
	if strings.Contains(out.Results[0].Reason, testKey) {
		t.Fatal("failure reason leaks the key")
	}
	if h := out.Health(); h.Status != domain.LaneFailed {
		t.Fatalf("health %+v", h)
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
	c := client(f)
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
	Run(context.Background(), RunInput{Config: testConfig(), Client: c, RunID: "r", Units: units})
	want := []string{"issue-high", "issue-low", RubricKey("pkg/a.go", 9), RubricKey("pkg/a.go", 100), RubricKey("pkg/b.go", 10), "aaa", "bbb", "pkg/z.go"}
	if fmt.Sprintf("%q", asked) != fmt.Sprintf("%q", want) {
		t.Fatalf("asked %q\nwant  %q", asked, want)
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
	c := client(f)
	units := []Unit{unit("e", 1), unit("c", 5), unit("b", 5), unit("a", 2), unit("d", 3)}
	cfg := testConfig()
	cfg.MaxPerRun = 3
	out := Run(context.Background(), RunInput{Config: cfg, Client: c, RunID: "r", Units: units})
	if strings.Join(asked, ",") != "b,c,d" {
		t.Fatalf("asked %v: want baseline order with ties by issue id", asked)
	}
	if out.Total != 5 || out.Attempted != 3 || out.Annotated != 3 || out.Capped != 2 {
		t.Fatalf("outcome %+v", out)
	}
	assertPartialCapDisclosed(t, out)
}

func assertPartialCapDisclosed(t *testing.T, out Outcome) {
	t.Helper()
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
	if mode, text := out.Disclosure(capped); mode != domain.RankingModePartial || text != "partial (3/5)" {
		t.Fatalf("disclosure %q %q", mode, text)
	}
}

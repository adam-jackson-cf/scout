package decisions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"

	"scout/internal/domain"
	"scout/internal/policy"
)

func noul(id string) Question {
	q := TriageRealDefect
	q.ID = id
	return q
}

// answersFor writes one Noul answer per requested question key, in a reply
// that bills `cost` (omitted when negative).
func answersFor(w http.ResponseWriter, model string, body map[string]any, p float64, cost float64) {
	keys := make([]string, 0)
	for key := range body["questions"].(map[string]any) {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	answers := make([]string, 0, len(keys))
	for _, key := range keys {
		answers = append(answers, fmt.Sprintf(`%q:{"type":"noul","noul":%v}`, key, p))
	}
	usage := ""
	if cost >= 0 {
		usage = fmt.Sprintf(`,"usage":{"input_tokens":10,"output_tokens":2,"cost":%v}`, cost)
	}
	fmt.Fprintf(w, `{"id":"gen-1","model":%q,"answers":{%s}%s}`, model, strings.Join(answers, ","), usage)
}

func TestIndependentQuestionsFanOutInOneRequestWithOneBilledUsage(t *testing.T) {
	f := newFake(t, func(_ int, body map[string]any, w http.ResponseWriter) { answersFor(w, testModel, body, 0.6, 0.0002) })
	c, _ := client(f)
	sink := &memorySink{}
	u := unit("a", 1)
	u.Questions = []Question{noul("rubric.one"), noul("rubric.two"), noul("rubric.three")}
	out := Run(context.Background(), RunInput{Config: Config{Model: testModel}, Client: c, RunID: "r", Units: []Unit{u}, Sink: sink})
	if f.count() != 1 || out.Annotated != 1 || len(sink.rows) != 3 {
		t.Fatalf("requests %d outcome %+v rows %d: want one request, three rows", f.count(), out, len(sink.rows))
	}
	for _, key := range []string{`"q1"`, `"q2"`, `"q3"`} {
		if !strings.Contains(f.bodies[0], key) {
			t.Fatalf("request lacks %s: %s", key, f.bodies[0])
		}
	}
	withUsage := 0
	for _, row := range sink.rows {
		if len(row.Usage) != 0 {
			withUsage++
		}
		if row.StateDigest != sink.rows[0].StateDigest || len(row.StateDigest) != 64 {
			t.Fatalf("state digest %q", row.StateDigest)
		}
		if row.Probability != 0.6 || len(row.Answer) == 0 {
			t.Fatalf("row %+v", row)
		}
	}
	if withUsage != 1 || len(sink.rows[0].Usage) == 0 {
		t.Fatalf("billed usage must sit on exactly the first row, found %d", withUsage)
	}
	if out.CostUSD != 0.0002 || out.Results[0].CostUSD != 0.0002 {
		t.Fatalf("cost %v / %v counted per call, not per question", out.CostUSD, out.Results[0].CostUSD)
	}
}

func TestReplyMustAnswerEveryQuestionAndQuestionIDsAreUniquePerRequest(t *testing.T) {
	f := newFake(t, func(_ int, _ map[string]any, w http.ResponseWriter) { answer(w, testModel, 0.5) }) // answers q1 only
	c, _ := client(f)
	_, err := c.Ask(context.Background(), []Question{noul("a"), noul("b")}, testModel, "u", json.RawMessage(`{}`))
	if failure, ok := err.(*Failure); !ok || failure.Kind != FailSchema {
		t.Fatalf("a reply missing an answer must be schema-invalid, got %v", err)
	}
	if _, err := c.Ask(context.Background(), []Question{noul("a"), noul("a")}, testModel, "u", json.RawMessage(`{}`)); err == nil || f.count() != 1 {
		t.Fatalf("duplicate ids must be rejected before any call: %v", err)
	}
	if _, err := c.Ask(context.Background(), nil, testModel, "u", json.RawMessage(`{}`)); err == nil {
		t.Fatal("an empty request must be rejected")
	}
}

func TestCostCapStopsRunAfterBilledCostReachesIt(t *testing.T) {
	f := newFake(t, func(_ int, body map[string]any, w http.ResponseWriter) { answersFor(w, testModel, body, 0.5, 0.3) })
	c, _ := client(f)
	units := []Unit{unit("a", 4), unit("b", 3), unit("c", 2), unit("d", 1)}
	out := Run(context.Background(), RunInput{Config: Config{Model: testModel, MaxPerRun: 100, MaxCostUSD: 0.5}, Client: c, RunID: "r", Units: units})
	if f.count() != 2 || out.Attempted != 2 || out.Annotated != 2 || out.Capped != 2 || !out.CostCapped {
		t.Fatalf("requests %d outcome %+v: want two calls ($0.60 >= $0.50), two capped", f.count(), out)
	}
	if out.CostUSD < 0.59 || out.CostUSD > 0.61 {
		t.Fatalf("billed cost %v", out.CostUSD)
	}
	capped := out.Results[2]
	if capped.Status != UnitCapped || capped.Reason != "cost cap reached" {
		t.Fatalf("capped %+v", capped)
	}
	if mode, text := out.Disclosure(capped); mode != ModePartial || text != "partial (2/4)" {
		t.Fatalf("disclosure %q %q", mode, text)
	}
	if health := out.Health(); health.Status != domain.LaneRan || !strings.Contains(health.Reason, "partial (2/4)") || !strings.Contains(health.Reason, "cost cap reached") || health.Required {
		t.Fatalf("health %+v", health)
	}
}

func TestACallCapOnlyRunIsNotReportedAsCostCapped(t *testing.T) {
	f := newFake(t, func(_ int, body map[string]any, w http.ResponseWriter) { answersFor(w, testModel, body, 0.5, 0.00005) })
	c, _ := client(f)
	out := Run(context.Background(), RunInput{Config: Config{Model: testModel, MaxPerRun: 1}, Client: c, RunID: "r", Units: []Unit{unit("a", 2), unit("b", 1)}})
	if out.Capped != 1 || out.CostCapped || out.Results[1].Reason != "call cap reached" {
		t.Fatalf("outcome %+v", out)
	}
}

func TestAReplyWithoutBilledCostStopsTheRunBecauseTheCapCannotBeEnforced(t *testing.T) {
	f := newFake(t, func(_ int, body map[string]any, w http.ResponseWriter) { answersFor(w, testModel, body, 0.5, -1) })
	c, _ := client(f)
	out := Run(context.Background(), RunInput{Config: Config{Model: testModel}, Client: c, RunID: "r", Units: []Unit{unit("a", 3), unit("b", 2), unit("c", 1)}})
	if f.count() != 1 || out.Annotated != 1 || out.Capped != 2 || out.Results[1].Reason != "billed cost unknown" {
		t.Fatalf("requests %d outcome %+v", f.count(), out)
	}
}

func TestADiscardedMismatchedReplyIsStillBilled(t *testing.T) {
	f := newFake(t, func(_ int, body map[string]any, w http.ResponseWriter) {
		answersFor(w, "typesafe/jev-9.9-rerouted", body, 0.5, 0.4)
	})
	c, _ := client(f)
	out := Run(context.Background(), RunInput{Config: Config{Model: testModel, MaxCostUSD: 0.5}, Client: c, RunID: "r", Units: []Unit{unit("a", 3), unit("b", 2), unit("c", 1)}})
	if out.Incomplete != 2 || out.Capped != 1 || out.CostUSD < 0.79 || out.CostUSD > 0.81 || len(out.Results[0].Annotations) != 0 {
		t.Fatalf("outcome %+v: discarded replies cost money and count toward the cap", out)
	}
}

func TestVersionKeyCoversTheQuestionKind(t *testing.T) {
	a := Question{ID: "q", Kind: KindNoul, Instructions: "x", Criteria: map[string]string{"true": "t", "false": "f"}, StateShape: []string{"state.x"}, Order: OptionOrderRule{Rule: OrderFixed}}
	b := a
	b.Kind = KindScore
	if a.Version(testModel) == b.Version(testModel) {
		t.Fatal("a Noul and a Score with the same text must not share a version key")
	}
}

func TestPinRejectsAliasesAndUndatedReleaseLinesButAcceptsTheInitDefault(t *testing.T) {
	for _, model := range []string{"", "jev-latest", "typesafe/jev-preview", "typesafe/jev-1.14", "jev-1.13"} {
		if (Config{Model: model}).Validate() == nil {
			t.Fatalf("model %q must be rejected", model)
		}
	}
	for _, model := range []string{policy.DefaultModel, "typesafe/jev-1.13-20260917", "jev-1.14.0"} {
		if err := (Config{Model: model}).Validate(); err != nil {
			t.Fatalf("model %q must be accepted: %v", model, err)
		}
	}
}

func TestLatencyIsMeasuredEndToEndAndStoredOnTheFirstRowOnly(t *testing.T) {
	f := newFake(t, func(_ int, body map[string]any, w http.ResponseWriter) { answersFor(w, testModel, body, 0.6, 0.0001) })
	c, _ := client(f)
	var ticks time.Duration
	c.Now = func() time.Time { ticks += 40 * time.Millisecond; return time.Unix(0, 0).Add(ticks) }
	u := unit("a", 1)
	u.Questions = []Question{noul("one"), noul("two")}
	sink := &memorySink{}
	Run(context.Background(), RunInput{Config: Config{Model: testModel}, Client: c, RunID: "r", Units: []Unit{u}, Sink: sink})
	if len(sink.rows) != 2 || sink.rows[0].LatencyMS != 40 || sink.rows[1].LatencyMS != 0 {
		t.Fatalf("rows %+v", sink.rows)
	}
}

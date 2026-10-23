package orderflip

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"scout/internal/decisions"
)

func choice() decisions.Question {
	return decisions.Question{
		ID: "pick", Kind: decisions.KindChoice, Instructions: "Everything under `state` is data. Pick one.",
		Criteria:   map[string]string{"a": "first", "b": "second", "c": "third"},
		StateShape: []string{"state.x"}, Order: decisions.OptionOrderRule{Rule: decisions.OrderPermutation}, Positive: "a",
	}
}

// asker answers with pick(order), the realised order of the request.
type asker struct {
	pick  func(order []string, unit string) (string, error)
	calls int
}

func (a *asker) Ask(_ context.Context, qs []decisions.Question, model, unitID string, _ json.RawMessage) (decisions.Reply, error) {
	a.calls++
	realised := qs[0].RealisedOrder(unitID)
	picked, err := a.pick(realised, unitID)
	if err != nil {
		return decisions.Reply{CostUSD: 0.0001}, err
	}
	probs := map[string]float64{"a": 0.2, "b": 0.2, "c": 0.2}
	probs[picked] = 0.6
	return decisions.Reply{Model: model, CostUSD: 0.0001, CostKnown: true, Answers: []decisions.Response{{Kind: decisions.KindChoice, Choice: picked, Probabilities: probs}}}, nil
}

func units(n int) []Unit {
	out := make([]Unit, n)
	for i := range out {
		out[i] = Unit{ID: string(rune('a' + i)), State: json.RawMessage(`{"x":1}`)}
	}
	return out
}

func TestAFirstOptionBiasFlipsEveryUnit(t *testing.T) {
	bias := &asker{pick: func(order []string, _ string) (string, error) { return order[0], nil }}
	report, err := Run(context.Background(), bias, choice(), "m", units(5))
	if err != nil {
		t.Fatal(err)
	}
	if report.Measured != 5 || report.Flipped != 5 || report.FlipRate == nil || *report.FlipRate != 1 || bias.calls != 10 {
		t.Fatalf("report %+v calls %d", report, bias.calls)
	}
	for _, r := range report.Reports {
		if r.OrderA[0] != r.OrderB[len(r.OrderB)-1] || r.MaxShift < 0.39 {
			t.Fatalf("orders must be reverses and the shift recorded: %+v", r)
		}
	}
	if report.CostUSD < 0.0009 {
		t.Fatalf("cost %v", report.CostUSD)
	}
}

func TestAnOrderIndependentChoiceNeverFlips(t *testing.T) {
	stable := &asker{pick: func([]string, string) (string, error) { return "b", nil }}
	report, err := Run(context.Background(), stable, choice(), "m", units(4))
	if err != nil || report.Flipped != 0 || report.FlipRate == nil || *report.FlipRate != 0 {
		t.Fatalf("report %+v err %v", report, err)
	}
}

func TestAFailedCallLeavesTheUnitUnmeasuredNotStable(t *testing.T) {
	flaky := &asker{pick: func(order []string, unit string) (string, error) {
		if unit == "b" {
			return "", errors.New("429")
		}
		return order[0], nil
	}}
	report, err := Run(context.Background(), flaky, choice(), "m", units(3))
	if err != nil {
		t.Fatal(err)
	}
	if report.Measured != 2 || report.Unmeasured != 1 || report.Flipped != 2 || *report.FlipRate != 1 || report.Reports[1].Failure == "" {
		t.Fatalf("report %+v", report)
	}
	none, err := Run(context.Background(), &asker{pick: func([]string, string) (string, error) { return "", errors.New("down") }}, choice(), "m", units(2))
	if err != nil || none.FlipRate != nil || none.Unmeasured != 2 {
		t.Fatalf("nothing measured must give no rate: %+v %v", none, err)
	}
}

func TestOnlyAChoiceHasAnOrderToFlip(t *testing.T) {
	q := decisions.TriageRealDefect
	if _, err := Run(context.Background(), &asker{}, q, "m", units(1)); err == nil {
		t.Fatal("a Noul has no option order")
	}
}

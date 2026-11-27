package decisions

import (
	"testing"

	"scout/internal/policy"
)

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
		c := testConfig()
		c.Model = model
		if c.Validate() == nil {
			t.Fatalf("model %q must be rejected", model)
		}
	}
	for _, model := range []string{policy.DefaultModel, "typesafe/jev-1.13-20260917", "jev-1.14.0"} {
		c := testConfig()
		c.Model = model
		if err := c.Validate(); err != nil {
			t.Fatalf("model %q must be accepted: %v", model, err)
		}
	}
}

func TestFormulaScoreSumsInQuestionIDOrder(t *testing.T) {
	wa, wb, wc := 0.2, 0.3, 0.1
	pa, pb, pc := 0.7, 0.1, 0.9
	f := Formula{Weights: map[string]float64{"c": wc, "a": wa, "b": wb}}
	answers := map[string]float64{"a": pa, "b": pb, "c": pc}
	want := wa * pa
	want += wb * pb
	want += wc * pc
	for range 50 {
		if got, ok := f.Score(answers); !ok || got != want {
			t.Fatalf("score %v %v, want %v", got, ok, want)
		}
	}
	if _, ok := f.Score(map[string]float64{"a": 1}); ok {
		t.Fatal("a missing weighted answer must not score")
	}
}

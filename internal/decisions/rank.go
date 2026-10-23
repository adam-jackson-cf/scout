package decisions

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// Formula is the named rank formula composed in code over Jev outputs only,
// used ordinally. Changing a weight changes its version and re-ranks stored
// answers with no model call. A term mixing a rule weight with a Jev
// probability would need the reliability check of design §6 first, so none
// exists here.
type Formula struct {
	ID        string
	Weights   map[string]float64 // question id -> weight
	Questions []Question
}

// LaunchFormula orders by P(real_defect) from the triage question.
func LaunchFormula() Formula {
	return Formula{ID: "jev-ordinal-p/v1", Weights: map[string]float64{TriageRealDefect.ID: 1}, Questions: []Question{TriageRealDefect}}
}

type formulaPayload struct {
	ID        string             `json:"id"`
	Weights   map[string]float64 `json:"weights"`
	Questions []string           `json:"question_versions"`
}

// Version is the rank-formula version: formula id, weights and the
// constituent question versions (which carry the resolved model).
func (f Formula) Version(resolvedModelID string) string {
	versions := make([]string, 0, len(f.Questions))
	for _, q := range f.Questions {
		versions = append(versions, q.Version(resolvedModelID))
	}
	sort.Strings(versions)
	raw, err := json.Marshal(formulaPayload{f.ID, f.Weights, versions})
	if err != nil {
		panic("decisions: formula payload is not encodable: " + err.Error())
	}
	sum := sha256.Sum256(raw)
	return f.ID + "@" + hex.EncodeToString(sum[:])
}

// Score combines stored answers (question id -> probability); ok is false if
// a weighted question has no answer.
func (f Formula) Score(answers map[string]float64) (float64, bool) {
	score := 0.0
	for id, weight := range f.Weights {
		p, ok := answers[id]
		if !ok {
			return 0, false
		}
		score += weight * p
	}
	return score, true
}

// BaselineVersion identifies the deterministic ranking baseline (design §6):
// its formula id, the rule-weight table digest and the path-role classifier
// version. The engine owns the formula, so it fills this in.
type BaselineVersion struct {
	FormulaID        string
	RuleWeightDigest string
	PathRoleVersion  string
}

// String is the stored form.
func (b BaselineVersion) String() string {
	return b.FormulaID + "+weights:" + b.RuleWeightDigest + "+path-role:" + b.PathRoleVersion
}

// Mode of one ranked finding.
const (
	ModeJev           = "jev"
	ModeDeterministic = "deterministic"
	ModePartial       = "partial"
)

// Item is one finding to rank.
type Item struct {
	IssueID  string
	Baseline float64
	// Answers holds stored probabilities by question id; nil when the finding
	// has no annotation.
	Answers    map[string]float64
	Mode       string
	Disclosure string
}

// Ranked is an item in rank order.
type Ranked struct {
	Item
	Position  int
	Score     float64 // Jev formula score for annotated items; baseline otherwise
	Annotated bool
}

// Rank orders annotated findings by the formula (ties: baseline, then issue
// id), then unannotated findings by baseline (ties: issue id). The two tiers
// are not comparable numerically, and an annotation can only lower a finding
// below its baseline neighbours, never close or hide it.
func Rank(formula Formula, items []Item) []Ranked {
	annotated := make([]Ranked, 0, len(items))
	plain := make([]Ranked, 0, len(items))
	for _, item := range items {
		if score, ok := formula.Score(item.Answers); ok && item.Answers != nil {
			annotated = append(annotated, Ranked{Item: item, Score: score, Annotated: true})
		} else {
			plain = append(plain, Ranked{Item: item, Score: item.Baseline})
		}
	}
	sort.SliceStable(annotated, func(i, j int) bool {
		a, b := annotated[i], annotated[j]
		switch {
		case a.Score != b.Score:
			return a.Score > b.Score
		case a.Baseline != b.Baseline:
			return a.Baseline > b.Baseline
		}
		return a.IssueID < b.IssueID
	})
	sort.SliceStable(plain, func(i, j int) bool {
		a, b := plain[i], plain[j]
		if a.Baseline != b.Baseline {
			return a.Baseline > b.Baseline
		}
		return a.IssueID < b.IssueID
	})
	out := append(annotated, plain...)
	for index := range out {
		out[index].Position = index + 1
	}
	return out
}

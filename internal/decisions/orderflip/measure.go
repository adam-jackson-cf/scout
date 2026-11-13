// Package orderflip measures how often a Choice question's selected option
// changes when only the order of its options changes (jev-actions
// question-design: Jev leans toward the option listed first, so a Choice is
// asked in two orders and its flip rate reported).
//
// The harness needs no labels: it compares the question with itself. It lives
// outside the adapter package because the adapter computes no quality metric;
// a flip rate is a property of a question's wording on a state set, and the
// design runs it on a Choice's own labelled set before that Choice may rank.
package orderflip

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"scout/internal/decisions"
)

// Asker is the part of the adapter client the harness uses.
type Asker interface {
	Ask(ctx context.Context, questions []decisions.Question, model, unitID string, state json.RawMessage) (decisions.Reply, error)
}

// Unit is one state to ask the question over.
type Unit struct {
	ID    string
	State json.RawMessage
}

// UnitReport is one unit's two asks.
type UnitReport struct {
	ID     string   `json:"id"`
	OrderA []string `json:"order_a"`
	OrderB []string `json:"order_b"`
	// ChoiceA and ChoiceB are the options selected in each order; both are
	// empty for an unmeasured unit, whose Failure says why.
	ChoiceA string `json:"choice_a,omitempty"`
	ChoiceB string `json:"choice_b,omitempty"`
	Flipped bool   `json:"flipped"`
	// MaxShift is the largest change of any option's probability between the
	// two orders.
	MaxShift float64 `json:"max_shift"`
	Failure  string  `json:"failure,omitempty"`
	// CostUSD is the billed cost of the unit's two calls, as reported.
	CostUSD float64 `json:"cost_usd"`
}

// Report is the flip rate of one Choice on one state set.
type Report struct {
	QuestionID string `json:"question_id"`
	Version    string `json:"question_version"`
	Model      string `json:"model"`
	Units      int    `json:"units"`
	Measured   int    `json:"measured"`
	// Unmeasured units had a failed or mismatched call in either order; they
	// are never counted as stable.
	Unmeasured int `json:"unmeasured"`
	Flipped    int `json:"flipped"`
	// FlipRate is Flipped / Measured; nil when nothing was measured.
	FlipRate *float64     `json:"flip_rate"`
	CostUSD  float64      `json:"cost_usd"`
	Reports  []UnitReport `json:"reports"`
}

// Run asks q over every unit twice: once in the order the adapter realises for
// the unit (order A) and once in the reverse (order B). Only a Choice can be
// run. A failed call leaves its unit unmeasured.
func Run(ctx context.Context, asker Asker, q decisions.Question, model string, units []Unit) (Report, error) {
	if q.Kind != decisions.KindChoice {
		return Report{}, fmt.Errorf("question %q is a %s; only a Choice has an option order", q.ID, q.Kind)
	}
	if err := q.Validate(); err != nil {
		return Report{}, err
	}
	report := Report{QuestionID: q.ID, Version: q.Version(model), Model: model, Units: len(units), Reports: make([]UnitReport, 0, len(units))}
	for _, unit := range units {
		orderA := q.RealisedOrder(unit.ID)
		orderB := reverse(orderA)
		one := UnitReport{ID: unit.ID, OrderA: orderA, OrderB: orderB}
		a, errA := askIn(ctx, asker, q, model, unit, orderA)
		b, errB := askIn(ctx, asker, q, model, unit, orderB)
		one.CostUSD = a.CostUSD + b.CostUSD
		report.CostUSD += one.CostUSD
		if err := errors.Join(errA, errB); err != nil {
			one.Failure = err.Error()
			report.Unmeasured++
			report.Reports = append(report.Reports, one)
			if ctx.Err() != nil {
				return report, ctx.Err()
			}
			continue
		}
		answerA, answerB := a.Answers[0], b.Answers[0]
		one.ChoiceA, one.ChoiceB = answerA.Choice, answerB.Choice
		one.Flipped = answerA.Choice != answerB.Choice
		for option, pa := range answerA.Probabilities {
			one.MaxShift = math.Max(one.MaxShift, math.Abs(pa-answerB.Probabilities[option]))
		}
		report.Measured++
		if one.Flipped {
			report.Flipped++
		}
		report.Reports = append(report.Reports, one)
	}
	if report.Measured > 0 {
		rate := float64(report.Flipped) / float64(report.Measured)
		report.FlipRate = &rate
	}
	return report, nil
}

func askIn(ctx context.Context, asker Asker, q decisions.Question, model string, unit Unit, order []string) (decisions.Reply, error) {
	fixed := q
	fixed.Order = decisions.OptionOrderRule{Rule: decisions.OrderFixed, Options: order}
	reply, err := asker.Ask(ctx, []decisions.Question{fixed}, model, unit.ID, unit.State)
	if err != nil {
		return reply, err
	}
	if len(reply.Answers) != 1 {
		return reply, errors.New("reply has no single answer")
	}
	return reply, nil
}

func reverse(order []string) []string {
	out := make([]string, len(order))
	for i, option := range order {
		out[len(order)-1-i] = option
	}
	return out
}

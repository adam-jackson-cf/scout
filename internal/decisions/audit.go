package decisions

import (
	"context"
	"math"
	"time"

	"scout/internal/domain"
)

// SamplerVersion identifies the interim audit sampler. Its sampling unit
// (issue id, formula version), strata, fraction and floor are [ASSUMPTION]
// until X1's preregistration is signed off, which may discard interim draws.
const SamplerVersion = "interim-audit/v1"

// ProvisionalFraction is rev 2.1's 5 % total audit fraction.
const ProvisionalFraction = 0.05

// Strata of the interim sampler.
const (
	StratumTop         = "top"
	StratumBottom      = "bottom"
	StratumUnannotated = "unannotated"
	StratumCertainty   = "certainty"
)

// AuditUnit is one ranked finding offered to the sampler.
type AuditUnit struct {
	IssueID   string
	Position  int // 1-based rank position
	Annotated bool
	Critical  bool
}

// Assignment is a unit's stratum and inclusion probability.
type Assignment struct {
	AuditUnit
	Stratum     string
	Probability float64
}

// Assign places units in strata and sets each unit's inclusion probability.
// The audit floor is the rank position that halves the run: critical findings
// below it form the certainty stratum (probability 1) and are removed from
// the halves. The fraction is split equally across the non-empty strata of
// {top half, bottom half, unannotated}; within a stratum every unit has the
// same probability, so the expected sample is fraction × N / strata.
func Assign(units []AuditUnit, fraction float64) []Assignment {
	n := len(units)
	floor := (n + 1) / 2
	out := make([]Assignment, 0, n)
	size := map[string]int{}
	for _, unit := range units {
		stratum := StratumTop
		switch {
		case unit.Critical && unit.Position > floor:
			stratum = StratumCertainty
		case !unit.Annotated:
			stratum = StratumUnannotated
		case unit.Position > floor:
			stratum = StratumBottom
		}
		out = append(out, Assignment{AuditUnit: unit, Stratum: stratum})
		size[stratum]++
	}
	strata := 0
	for _, stratum := range []string{StratumTop, StratumBottom, StratumUnannotated} {
		if size[stratum] > 0 {
			strata++
		}
	}
	for index := range out {
		stratum := out[index].Stratum
		if stratum == StratumCertainty {
			out[index].Probability = 1
			continue
		}
		p := fraction * float64(n) / float64(strata) / float64(size[stratum])
		out[index].Probability = math.Min(1, p)
	}
	return out
}

// AuditStore holds draws. A draw is made once per (issue, formula version,
// sampler version), at the first eligible run, and never replaced.
type AuditStore interface {
	AuditDraw(ctx context.Context, issueID, formulaVersion, samplerVersion string) (domain.AuditDraw, bool, error)
	RecordAuditDraw(ctx context.Context, draw domain.AuditDraw) (domain.AuditDraw, error)
}

// DrawAudit makes the independent per-unit Bernoulli draw at each new unit's
// stored inclusion probability and returns the stored draws of every unit.
// Units already drawn keep their stored draw and probability. rand returns a
// uniform value in [0, 1).
func DrawAudit(ctx context.Context, store AuditStore, runID, formulaVersion string, units []AuditUnit, fraction float64, rand func() float64, now time.Time) ([]domain.AuditDraw, error) {
	assignments := Assign(units, fraction)
	draws := make([]domain.AuditDraw, 0, len(assignments))
	for _, assignment := range assignments {
		if existing, ok, err := store.AuditDraw(ctx, assignment.IssueID, formulaVersion, SamplerVersion); err != nil {
			return nil, err
		} else if ok {
			draws = append(draws, existing)
			continue
		}
		stored, err := store.RecordAuditDraw(ctx, domain.AuditDraw{
			RunID: runID, IssueID: assignment.IssueID, FormulaVersion: formulaVersion, SamplerVersion: SamplerVersion,
			Stratum: assignment.Stratum, InclusionProbability: assignment.Probability, Drawn: rand() < assignment.Probability,
			RankPosition: assignment.Position, CreatedAt: now,
		})
		if err != nil {
			return nil, err
		}
		draws = append(draws, stored)
	}
	return draws, nil
}

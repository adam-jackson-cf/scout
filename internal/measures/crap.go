package measures

import (
	"math"
	"sort"

	"scout/internal/domain"
)

// CrapThreshold is the CRAP value above which a function counts toward the
// scope's crap_load (design §8.2).
const CrapThreshold = 30.0

// Coverage is the revision-matched line coverage a CRAP computation reads.
type Coverage struct {
	// Lines maps a repository-relative path to line → hit count over the
	// file's instrumented lines.
	Lines map[string]map[int]int
	// Digest identifies the artifact set; it is part of the signal's tool
	// digest so measuring one snapshot with and without coverage never collides.
	Digest string
	// Unusable, when non-empty, is why no coverage artifact applies (absent,
	// unreadable, or its revision does not match the snapshot).
	Unusable string
}

// CRAP is Savoia and Evans' change-risk score: CC² × (1 − cov)³ + CC.
func CRAP(cc int, coverage float64) float64 {
	c := float64(cc)
	return c*c*math.Pow(1-coverage, 3) + c
}

// functionCoverage is the fraction of instrumented lines covered within the
// function's span: from its first line to the line before the next function
// of the file (the last function runs to the end of the file). ok is false
// when the artifact instruments no line in the span.
func functionCoverage(lines map[int]int, from, to int) (float64, bool) {
	instrumented, covered := 0, 0
	for line, hits := range lines {
		if line < from || line >= to {
			continue
		}
		instrumented++
		if hits > 0 {
			covered++
		}
	}
	if instrumented == 0 {
		return 0, false
	}
	return float64(covered) / float64(instrumented), true
}

// crapSignal computes scout.crap@1 for one scope population: Value =
// crap_load (Σ CRAP over functions with CRAP > 30), Numerator =
// Σ CC·(1−cov), Denominator = Σ CC over the functions that have coverage data
// (so Numerator ÷ Denominator = uncovered_cc_mass). Functions in files the
// artifact does not instrument are counted as `no_coverage_data` and are not
// treated as uncovered.
func crapSignal(s domain.Signal, population []domain.File, lizard map[string]FileMeasure, coverage *Coverage, detail func(map[string]any) map[string]any) domain.Signal {
	var load, uncovered, total float64
	var covered, withoutData, over int
	for _, file := range population {
		perLine, instrumented := coverage.Lines[file.Path]
		functions := append([]Function(nil), lizard[file.Path].Functions...)
		sort.SliceStable(functions, func(i, j int) bool { return functions[i].Line < functions[j].Line })
		for i, fn := range functions {
			end := math.MaxInt
			for _, next := range functions[i+1:] {
				if next.Line > fn.Line {
					end = next.Line
					break
				}
			}
			cov, ok := 0.0, false
			if instrumented {
				cov, ok = functionCoverage(perLine, fn.Line, end)
			}
			if !ok {
				withoutData++
				continue
			}
			covered++
			crap := CRAP(fn.CC, cov)
			total += float64(fn.CC)
			uncovered += float64(fn.CC) * (1 - cov)
			if crap > CrapThreshold {
				load += crap
				over++
			}
		}
	}
	extra := map[string]any{
		"artifact": coverage.Digest, "n_functions_with_coverage": covered, "n_functions_without_coverage_data": withoutData,
		"n_crap_over_threshold": over, "crap_threshold": CrapThreshold,
	}
	if covered == 0 {
		s.Missing, s.MissingReason = domain.SignalUnmeasured, "the coverage artifact has no line data for the scope's functions"
		s.Detail = detail(extra)
		return s
	}
	s.Numerator, s.Denominator, s.Value = domain.Float(uncovered), domain.Float(total), domain.Float(load)
	extra["crap_load"] = load
	extra["uncovered_cc_mass"] = uncovered / total
	s.Detail = detail(extra)
	return s
}

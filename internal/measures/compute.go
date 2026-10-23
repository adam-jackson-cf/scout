package measures

import (
	"fmt"
	"math"
	"sort"

	"scout/internal/domain"
)

// Metric IDs and the single version of each formula.
const (
	MetricCC            = "scout.cc"
	MetricComplexBurden = "scout.complex_burden"
	MetricBurdenDensity = "scout.burden_density"
	MetricDupPct        = "scout.dup_pct"
	MetricCrap          = "scout.crap"
	metricVersion       = 1
)

// Facts are the scanner outputs a computation needs. A non-empty LizardErr or
// JscpdErr marks the corresponding measures unmeasured with that reason.
type Facts struct {
	// Lizard maps a repository-relative path to its measurement; a candidate
	// absent from it was not measured by lizard.
	Lizard    map[string]FileMeasure
	LizardErr string

	// Clones and Lines describe the eligible population given to jscpd: Lines
	// maps each of its files to its number of source lines.
	Clones   []Clone
	Lines    map[string]int
	JscpdErr string
}

// Input is everything Compute needs; it performs no I/O.
type Input struct {
	Snapshot string
	// RunScope records what the run inventory covered (repository for a full
	// run, a narrower scope for a changed-files run); it is reported in Detail.
	RunScope string
	// Candidates are inventory files of supported production source: the
	// pre-NLOC population.
	Candidates []domain.File
	Features   []domain.Feature
	Facts      Facts
}

// Eligible returns the candidates that lizard measured at NLOC ≥ MinFileNLOC.
func Eligible(candidates []domain.File, lizard map[string]FileMeasure) []domain.File {
	var eligible []domain.File
	for _, file := range candidates {
		if measure, ok := lizard[file.Path]; ok && measure.NLOC >= MinFileNLOC {
			eligible = append(eligible, file)
		}
	}
	return eligible
}

// Candidates filters inventory files to production source of supported
// languages (design §8.2 path roles); only files Scout may read qualify.
func Candidates(files []domain.File) []domain.File {
	var out []domain.File
	for _, file := range files {
		if file.Excluded || !file.Eligible || !IsProductionSource(file.Path) {
			continue
		}
		out = append(out, file)
	}
	return out
}

// Compute derives the §8.2 signals for every scope: scout.cc, complex_burden,
// burden_density, dup_pct and an `unmeasured` scout.crap (no coverage artifact
// is ingested yet). Output order is deterministic: scope anchor, then metric.
func Compute(in Input) []domain.Signal {
	candidateByPath := make(map[string]domain.File, len(in.Candidates))
	for _, file := range in.Candidates {
		candidateByPath[file.Path] = file
	}
	scopes := BuildScopes(in.Candidates, in.Features)
	eligibleSet := map[string]bool{}
	for _, file := range Eligible(in.Candidates, in.Facts.Lizard) {
		eligibleSet[file.Path] = true
	}
	lizardOK := in.Facts.LizardErr == ""
	jscpdOK := lizardOK && in.Facts.JscpdErr == ""
	scopesOfFile := map[string]int{}
	for _, scope := range scopes {
		for _, path := range scope.Paths {
			scopesOfFile[path]++
		}
	}
	configDigest := ConfigDigest()
	lizardDigest, jscpdDigest := LizardDigest(), JscpdDigest()

	var signals []domain.Signal
	for _, scope := range scopes {
		if len(scope.Paths) == 0 {
			continue // no production source: nothing to measure, not a measurement gap
		}
		var population, members []domain.File
		unmeasuredFiles := 0
		for _, path := range scope.Paths {
			file := candidateByPath[path]
			members = append(members, file)
			if eligibleSet[path] {
				population = append(population, file)
			} else if _, measured := in.Facts.Lizard[path]; !measured && lizardOK {
				unmeasuredFiles++
			}
		}
		// Without lizard the population is undefined; the digest then covers
		// the candidates and Detail says so.
		digestFiles, populationKind := population, "eligible"
		if !lizardOK {
			digestFiles, populationKind = members, "candidates"
		}
		overlap := 0
		for _, file := range population {
			if scopesOfFile[file.Path] > 1 {
				overlap++
			}
		}
		base := domain.Signal{
			Version: metricVersion, ScopeKind: scope.Kind, ScopeAnchor: scope.Anchor,
			Snapshot: in.Snapshot, ConfigDigest: configDigest, Claim: domain.ClaimDescriptive,
			PopulationDigest: PopulationDigest(digestFiles),
		}
		detail := func(extra map[string]any) map[string]any {
			d := map[string]any{
				"population": populationKind, "n_files": len(population), "n_candidates": len(members),
				"unmeasured_files": unmeasuredFiles, "overlap_files": overlap, "run_scope": in.RunScope,
			}
			for key, value := range extra {
				d[key] = value
			}
			return d
		}
		lizardSignal := func(metric string) domain.Signal {
			s := base
			s.MetricID, s.ToolDigest = metric, lizardDigest
			return s
		}
		missing := func(s domain.Signal, kind domain.SignalMissing, reason string) domain.Signal {
			s.Missing, s.MissingReason, s.Detail = kind, reason, detail(nil)
			return s
		}

		cc, burden, density := lizardSignal(MetricCC), lizardSignal(MetricComplexBurden), lizardSignal(MetricBurdenDensity)
		switch {
		case !lizardOK:
			reason := "lizard did not run: " + in.Facts.LizardErr
			cc, burden, density = missing(cc, domain.SignalUnmeasured, reason), missing(burden, domain.SignalUnmeasured, reason), missing(density, domain.SignalUnmeasured, reason)
		case len(population) == 0:
			reason := fmt.Sprintf("scope has no eligible files (production source with NLOC >= %d)", MinFileNLOC)
			cc, burden, density = missing(cc, domain.SignalInsufficient, reason), missing(burden, domain.SignalInsufficient, reason), missing(density, domain.SignalInsufficient, reason)
		default:
			var ccs []int
			var nloc, hi, total float64
			for _, file := range population {
				measure := in.Facts.Lizard[file.Path]
				nloc += float64(measure.NLOC)
				for _, fn := range measure.Functions {
					ccs = append(ccs, fn.CC)
					weight := float64(fn.CC) * math.Sqrt(float64(fn.NLOC))
					total += weight
					if fn.CC > HighCC {
						hi += weight
					}
				}
			}
			sort.Ints(ccs)
			nHigh := 0
			for _, value := range ccs {
				if value > HighCC {
					nHigh++
				}
			}
			extra := map[string]any{"n_functions": len(ccs), "n_high_cc": nHigh}
			if len(ccs) > 0 {
				extra["cc_p90"] = ccs[int(math.Ceil(0.9*float64(len(ccs))))-1]
				extra["cc_max"] = ccs[len(ccs)-1]
			}
			cc.Numerator, cc.Denominator, cc.Value = domain.Float(float64(nHigh)), domain.Float(float64(len(ccs))), domain.Float(float64(nHigh))
			cc.Detail = detail(extra)

			burden.Numerator, burden.Denominator = domain.Float(hi), domain.Float(total)
			burden.Detail = detail(extra)
			if total == 0 {
				burden = missing(burden, domain.SignalInsufficient, "eligible files hold no functions")
				burden.Numerator, burden.Denominator = nil, nil
			} else {
				burden.Value = domain.Float(hi / total)
			}
			density.Numerator, density.Denominator, density.Value = domain.Float(hi), domain.Float(nloc), domain.Float(hi/nloc)
			density.Detail = detail(extra)
		}
		signals = append(signals, cc, burden, density)

		dup := base
		dup.MetricID, dup.ToolDigest = MetricDupPct, jscpdDigest
		switch {
		case !lizardOK:
			dup = missing(dup, domain.SignalUnmeasured, "the eligible population is undefined because lizard did not run: "+in.Facts.LizardErr)
		case !jscpdOK:
			dup = missing(dup, domain.SignalUnmeasured, "jscpd did not run: "+in.Facts.JscpdErr)
		case len(population) == 0:
			dup = missing(dup, domain.SignalInsufficient, fmt.Sprintf("scope has no eligible files (production source with NLOC >= %d)", MinFileNLOC))
		default:
			dupLines, lines, pairs, cross := duplicatedLines(population, in.Facts)
			dup.Numerator, dup.Denominator = domain.Float(float64(dupLines)), domain.Float(float64(lines))
			dup.Value = domain.Float(float64(dupLines) / float64(lines))
			dup.Detail = detail(map[string]any{"clone_pairs": pairs, "cross_scope_clone_pairs": cross})
		}
		crap := base
		crap.MetricID, crap.ToolDigest = MetricCrap, lizardDigest
		crap = missing(crap, domain.SignalUnmeasured, "no revision-matched coverage artifact has been ingested")
		signals = append(signals, dup, crap)
	}
	return signals
}

// duplicatedLines unions the clone spans over the scope's population:
// numerator = distinct duplicated line positions, denominator = line positions
// of the population as jscpd read them (so the ratio is within [0, 1] without a
// cap). A clone pair counts for the scope when either end lies in it and is
// cross-scope when the other end lies outside.
func duplicatedLines(population []domain.File, facts Facts) (duplicated, lines, pairs, cross int) {
	in := make(map[string]bool, len(population))
	for _, file := range population {
		in[file.Path] = true
		lines += facts.Lines[file.Path]
	}
	spans := map[string][][2]int{}
	for _, clone := range facts.Clones {
		a, b := in[clone.A.Path], in[clone.B.Path]
		if !a && !b {
			continue
		}
		pairs++
		if !(a && b) {
			cross++
		}
		for _, span := range []Span{clone.A, clone.B} {
			if in[span.Path] {
				end := span.End
				if limit := facts.Lines[span.Path]; end > limit {
					end = limit
				}
				if end >= span.Start {
					spans[span.Path] = append(spans[span.Path], [2]int{span.Start, end})
				}
			}
		}
	}
	for _, intervals := range spans {
		sort.Slice(intervals, func(i, j int) bool { return intervals[i][0] < intervals[j][0] })
		start, end := intervals[0][0], intervals[0][1]
		for _, interval := range intervals[1:] {
			if interval[0] > end+1 {
				duplicated += end - start + 1
				start, end = interval[0], interval[1]
			} else if interval[1] > end {
				end = interval[1]
			}
		}
		duplicated += end - start + 1
	}
	return duplicated, lines, pairs, cross
}

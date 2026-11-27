package association

import (
	"sort"
	"time"

	"scout/internal/domain"
	"scout/internal/policy"
)

// GuardRate preserves the population and missingness behind a guard measurement.
type GuardRate struct {
	Numerator   int     `json:"numerator"`
	Denominator int     `json:"denominator"`
	Value       float64 `json:"value"`
	Measured    bool    `json:"measured"`
	Reason      string  `json:"reason,omitempty"`
}

// GuardState is reproducible from exported decisions and the mapping generation.
// CatalogueStale is latched until a completed agent refresh changes Generation.
type GuardState struct {
	Generation     *time.Time `json:"generation"`
	WindowMerges   []string   `json:"windowMerges"`
	FallbackRate   GuardRate  `json:"fallbackRate"`
	PendingNewRate GuardRate  `json:"pendingNewRate"`
	CatalogueStale bool       `json:"catalogueStale"`
	Recommendation string     `json:"recommendation,omitempty"`
}

type guardMerge struct {
	commit                                string
	decided                               time.Time
	clusters, fallback, pending, answered int
	ineligible                            bool
}

// Guard replays decision-time windows, including historical trips outside the
// final window. It neither mutates decisions nor starts a mapping refresh.
func Guard(rows []domain.Association, lastRefresh *time.Time, cfg policy.AssociationGuard, prior GuardState) GuardState {
	out := GuardState{WindowMerges: []string{}}
	if lastRefresh != nil {
		generation := lastRefresh.UTC()
		out.Generation = &generation
	}
	out.CatalogueStale = sameGuardGeneration(prior.Generation, lastRefresh) && prior.CatalogueStale
	merges := guardMerges(rows, lastRefresh)
	start := replayGuardWindows(merges, cfg, &out)
	if len(merges) == 0 {
		out.FallbackRate = guardRate(0, 0, cfg.MinClusters, true)
		out.PendingNewRate = guardRate(0, 0, cfg.MinClusters, false)
	}
	for _, merge := range merges[start:] {
		out.WindowMerges = append(out.WindowMerges, merge.commit)
	}
	if out.CatalogueStale {
		out.Recommendation = "catalogue-stale: run scout map --refresh manually"
	}
	return out
}

func guardMerges(rows []domain.Association, lastRefresh *time.Time) []*guardMerge {
	// Skip is run-wide, even when the row reporting it is not in a rate.
	skippedRuns := make(map[string]bool)
	for _, row := range rows {
		if sameGuardGeneration(row.LastAgentRefreshAt, lastRefresh) && row.FallbackSkipped {
			skippedRuns[row.RunID] = true
		}
	}
	byMerge := make(map[string]*guardMerge)
	for _, row := range rows {
		if !guardCounts(row, lastRefresh) {
			continue
		}
		merge := byMerge[row.MergeCommit]
		if merge == nil {
			merge = &guardMerge{commit: row.MergeCommit}
			byMerge[row.MergeCommit] = merge
		}
		merge.add(row, skippedRuns[row.RunID])
	}
	merges := make([]*guardMerge, 0, len(byMerge))
	for _, merge := range byMerge {
		merges = append(merges, merge)
	}
	sort.Slice(merges, func(i, j int) bool {
		if !merges[i].decided.Equal(merges[j].decided) {
			return merges[i].decided.Before(merges[j].decided)
		}
		return merges[i].commit < merges[j].commit
	})
	return merges
}

func guardCounts(row domain.Association, lastRefresh *time.Time) bool {
	if !sameGuardGeneration(row.LastAgentRefreshAt, lastRefresh) || row.Outcome == domain.AssociationInherited {
		return false
	}
	switch row.Reason {
	case domain.AssociationReasonJevFailure, domain.AssociationReasonJevCap, domain.AssociationReasonPRCap:
		return false
	}
	return true
}

func (merge *guardMerge) add(row domain.Association, skippedRun bool) {
	if row.DecidedAt.After(merge.decided) {
		merge.decided = row.DecidedAt
	}
	merge.clusters++
	escalated := row.JevOutcome == "none" || row.JevOutcome == "unresolved"
	answered := escalated && fallbackAnswered(row)
	if escalated {
		merge.fallback++
	}
	if answered {
		merge.answered++
		if pendingAnswer(row) {
			merge.pending++
		}
	}
	if mergeDisqualified(row, skippedRun, escalated, answered) {
		merge.ineligible = true
	}
}

func mergeDisqualified(row domain.Association, skippedRun, escalated, answered bool) bool {
	return skippedRun || (escalated && !answered) || fallbackIneligible(row)
}

func pendingAnswer(row domain.Association) bool {
	return row.Outcome == domain.AssociationPendingNew || row.Reason == domain.AssociationReasonPendingCap
}

func fallbackAnswered(row domain.Association) bool {
	return row.CLIName == "codex" && (row.Outcome == domain.AssociationAssigned || row.Outcome == domain.AssociationPendingNew || (row.Outcome == domain.AssociationUnresolved && row.Reason == domain.AssociationReasonPendingCap))
}

func fallbackIneligible(row domain.Association) bool {
	return row.Reason == domain.AssociationReasonFallbackCap || row.Reason == domain.AssociationReasonTerraFailure
}

type guardWindow struct {
	clusters, fallback, eligible, pending, answered int
}

func (w *guardWindow) accumulate(merge *guardMerge, sign int) {
	w.clusters += sign * merge.clusters
	w.fallback += sign * merge.fallback
	if !merge.ineligible {
		w.eligible += sign * merge.clusters
		w.pending += sign * merge.pending
		w.answered += sign * merge.answered
	}
}

func replayGuardWindows(merges []*guardMerge, cfg policy.AssociationGuard, out *GuardState) int {
	var window guardWindow
	start := 0
	for i, merge := range merges {
		window.accumulate(merge, 1)
		if i-start+1 > cfg.WindowMerges {
			window.accumulate(merges[start], -1)
			start++
		}
		out.FallbackRate = guardRate(window.fallback, window.clusters, cfg.MinClusters, true)
		out.PendingNewRate = guardRate(window.pending, window.eligible, cfg.MinClusters, window.answered > 0)
		if guardTripped(*out, cfg) {
			out.CatalogueStale = true
		}
	}
	return start
}

func guardTripped(out GuardState, cfg policy.AssociationGuard) bool {
	return (out.FallbackRate.Measured && out.FallbackRate.Value > cfg.FallbackRateMax) || (out.PendingNewRate.Measured && out.PendingNewRate.Value > cfg.PendingNewRateMax)
}

func sameGuardGeneration(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Equal(*b)
}

func guardRate(numerator, denominator, minimum int, hasAnswer bool) GuardRate {
	rate := GuardRate{Numerator: numerator, Denominator: denominator}
	switch {
	case denominator < minimum || denominator == 0:
		rate.Reason = "insufficient_clusters"
	case !hasAnswer:
		rate.Reason = "no_fallback_answers"
	default:
		rate.Measured = true
		rate.Value = float64(numerator) / float64(denominator)
	}
	return rate
}

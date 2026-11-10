package shadow

import (
	"sort"
	"time"

	"scout/internal/domain"
)

type observation struct{ pr, tree, base string }

func observationOf(f domain.PRRunFact) observation {
	return observation{f.PRNumber, f.HeadTree, f.MergeBase}
}

func aggregate(l domain.ShadowLedger, v visibility, d *Dossier) {
	reps := make(map[observation]domain.PRRunFact)
	cohort := make(map[domain.AttemptID]domain.PRRunFact)
	window := make(map[domain.AttemptID]bool)
	for _, f := range l.Facts {
		if f.ID.RepositoryID != d.RepositoryID || f.CheckID != d.CheckID || f.RuleVersion != d.RuleVersion || f.Scope != domain.ShadowScopeCI || f.Status != "assessed" || !v.fact(f, d.AsOf) {
			continue
		}
		cohort[f.ID] = f
		if !f.StartedAt.After(d.Interval.After) {
			continue
		}
		window[f.ID] = true
		k := observationOf(f)
		old, ok := reps[k]
		if !ok || later(f, old) {
			reps[k] = f
		}
	}
	prs := make(map[string]bool)
	pins := make(map[domain.ShadowPins]int)
	mixtures := map[string]map[string]bool{"scout": {}, "tools": {}, "catalogue": {}, "policy": {}}
	for _, f := range reps {
		d.Stats.TotalRuns++
		prs[f.PRNumber] = true
		d.Representatives = append(d.Representatives, f.ID)
		pins[f.Pins]++
		mixtures["scout"][f.Pins.Scout] = true
		mixtures["tools"][f.Pins.Tools] = true
		mixtures["catalogue"][f.Pins.Catalogue] = true
		mixtures["policy"][f.Pins.Policy] = true
		if f.Comparison == "none" {
			d.Stats.UnknownAttribution++
		}
		switch f.Inspection {
		case "inspected":
			d.Inspected++
		case "not_applicable":
			d.NotApplicable++
		default:
			d.NotInspected++
		}
		if f.Comparison == "comparable" && f.Inspection == "inspected" {
			d.Stats.PRRuns++
			if d.Origin == nil || f.StartedAt.Before(*d.Origin) {
				x := f.StartedAt
				d.Origin = &x
			}
		}
	}
	d.Stats.CheckID = d.CheckID
	d.DistinctPRs = len(prs)
	d.PinMixtures = make(map[string][]string)
	for key, set := range mixtures {
		for x := range set {
			d.PinMixtures[key] = append(d.PinMixtures[key], x)
		}
		sort.Strings(d.PinMixtures[key])
	}
	for p, count := range pins {
		d.PinCounts = append(d.PinCounts, PinCount{Pins: p, Observations: count})
	}
	linked := make(map[string]bool)
	inWindow := make(map[string]bool)
	for _, link := range l.Links {
		if link.Scope != domain.ShadowScopeCI {
			continue
		}
		if _, ok := cohort[link.AttemptID]; ok {
			linked[link.EventID] = true
			if window[link.AttemptID] {
				inWindow[link.EventID] = true
			}
		}
	}
	seen := make(map[string]bool)
	for _, e := range l.Events {
		if seen[e.EventID] || !linked[e.EventID] || e.RepositoryID != d.RepositoryID || e.Scope != domain.ShadowScopeCI || e.CheckID != d.CheckID || e.RuleVersion != d.RuleVersion || !v.event(e, d.AsOf) {
			continue
		}
		seen[e.EventID] = true
		verdict, action, superseded := triageAt(l, e, d.AsOf)
		d.Triage = append(d.Triage, EventVerdict{e.EventID, verdict, action})
		d.SupersededNotWarranted += superseded
		if inWindow[e.EventID] {
			d.Stats.WouldBlock++
			switch verdict {
			case "warranted":
				d.Stats.Warranted++
			case "not_warranted":
				d.Stats.NotWarranted++
			default:
				d.Stats.Undecided++
			}
		} else {
			switch verdict {
			case "not_warranted":
				d.Stats.CohortNotWarranted++
				d.PriorNotWarranted++
			case "undecided":
				d.Stats.CohortUndecided++
				d.PriorUndecided++
			}
		}
	}
}

func triageAt(l domain.ShadowLedger, e domain.ProviderPREvent, t time.Time) (string, string, int) {
	var latest *domain.ShadowTriage
	warrantedHistory := 0
	for i := range l.Triage {
		row := &l.Triage[i]
		if row.RepositoryID != e.RepositoryID || row.EventID != e.EventID || row.Scope != domain.ShadowScopeCI || row.RecordedAt == nil || !Visible(*row.RecordedAt, false, row.RecordedAt, t) {
			continue
		}
		if row.Verdict == "not_warranted" {
			warrantedHistory++
		}
		if latest == nil || row.RecordedAt.After(*latest.RecordedAt) || row.RecordedAt.Equal(*latest.RecordedAt) && numeric(row.ActionID, latest.ActionID) > 0 {
			latest = row
		}
	}
	if latest == nil {
		return "undecided", "", 0
	}
	if latest.Verdict == "not_warranted" {
		warrantedHistory--
	}
	return latest.Verdict, latest.ActionID, warrantedHistory
}

package shadow

import (
	"fmt"
	"time"

	"scout/internal/domain"
	"scout/internal/policy"
)

// Evaluate never opens a store or consults installed policy or the network.
func Evaluate(l domain.ShadowLedger, q Query) (Report, error) {
	r := Report{Dossier: Dossier{EncodingVersion: q.EncodingVersion, RepositoryID: q.RepositoryID, CheckID: q.CheckID}, Live: LiveEnvelope{Certification: "current", ReaderTime: q.Now}}
	if r.Dossier.EncodingVersion == "" {
		r.Dossier.EncodingVersion = EncodingVersion
	}
	fail := func(code int, reason string) (Report, error) {
		r.ExitCode = code
		r.Dossier.Reason = reason
		return r, nil
	}
	if r.Dossier.EncodingVersion != EncodingVersion {
		return fail(3, "unsupported shadow encoding")
	}
	if err := policy.ValidateEligibleID(q.CheckID); err != nil {
		return fail(3, err.Error())
	}
	var c time.Time
	for _, f := range l.Frontiers {
		if f.RepositoryID == q.RepositoryID && f.CheckID == q.CheckID && f.CompleteThrough.After(c) {
			c = f.CompleteThrough
		}
	}
	if c.IsZero() {
		return fail(2, "unmeasured: no complete frontier")
	}
	t := c
	if q.AsOf != nil {
		t = *q.AsOf
		r.Live.Certification = "historical"
		if t.After(c) {
			if t.After(q.Now) {
				return fail(3, "cutoff exceeds frontier and reader time")
			}
			return fail(2, "cutoff exceeds complete frontier")
		}
	}
	r.Dossier.AsOf = t
	r.Live.FrontierAge = q.Now.Sub(c)
	v := newVisibility(l)
	var latest *domain.PRRunFact
	for i := range l.Facts {
		f := &l.Facts[i]
		if f.ID.RepositoryID != q.RepositoryID || f.CheckID != q.CheckID || f.Scope != domain.ShadowScopeCI || f.Status == "non_default_base" || f.RuleVersion == "" || !v.fact(*f, t) {
			continue
		}
		if !validID(f.ID) {
			return fail(2, "unmeasured: unavailable numeric attempt ordering")
		}
		if latest == nil || later(*f, *latest) {
			latest = f
		}
	}
	version := q.RuleVersion
	if version == "" {
		if latest == nil {
			return fail(2, "unmeasured: no visible attempt pin")
		}
		version = latest.RuleVersion
	}
	r.Dossier.RuleVersion = version
	known := false
	for _, f := range l.Facts {
		if f.ID.RepositoryID == q.RepositoryID && f.CheckID == q.CheckID && f.Scope == domain.ShadowScopeCI && f.RuleVersion == version && v.fact(f, t) {
			known = true
		}
	}
	if !known {
		return fail(3, "unknown exact rule version")
	}
	for _, row := range l.Triage {
		if row.RepositoryID != q.RepositoryID || row.Scope != domain.ShadowScopeCI || row.RecordedAt == nil || !Visible(*row.RecordedAt, false, row.RecordedAt, t) {
			continue
		}
		for _, event := range l.Events {
			if event.EventID == row.EventID && event.RepositoryID == q.RepositoryID && event.CheckID == q.CheckID && event.RuleVersion == version {
				if _, ok := decimal(row.ActionID); !ok {
					return fail(2, "unmeasured: unavailable numeric triage ordering")
				}
			}
		}
	}
	// A frontier seals the policy for every cutoff it covers. Select the earliest
	// covering record, not a later installed policy or a later reconciliation.
	var frontier *domain.ShadowFrontier
	for i := range l.Frontiers {
		f := &l.Frontiers[i]
		if f.RepositoryID == q.RepositoryID && f.CheckID == q.CheckID && f.RuleVersion == version && !f.CompleteThrough.Before(t) && (frontier == nil || f.CompleteThrough.Before(frontier.CompleteThrough)) {
			frontier = f
		}
	}
	if frontier == nil {
		return fail(2, "unmeasured: no cohort frontier")
	}
	r.Dossier.Frontier = *frontier
	var interval *domain.ShadowInterval
	for i := range l.Intervals {
		x := &l.Intervals[i]
		if x.RepositoryID == q.RepositoryID && x.CheckID == q.CheckID && x.RuleVersion == version && t.After(x.After) && !t.After(x.Through) && (interval == nil || x.After.After(interval.After)) {
			interval = x
		}
	}
	if interval == nil {
		return fail(2, "unmeasured: no complete containing interval")
	}
	r.Dossier.Interval = *interval
	// Through is clipped to the immutable selected frontier, never a mutable
	// extension of the stored interval beyond the historical query's frontier.
	if r.Dossier.Interval.Through.After(frontier.CompleteThrough) {
		r.Dossier.Interval.Through = frontier.CompleteThrough
	}
	for _, g := range l.GapPoints {
		if g.RepositoryID != q.RepositoryID || g.CheckID != q.CheckID || !v.gap(g, t) {
			continue
		}
		if g.At.Equal(t) {
			return fail(2, "unmeasured: cutoff is a gap point")
		}
		if g.At.After(r.Dossier.Interval.After) {
			r.Dossier.Interval.After = g.At
		}
		if v.late(g.At, func(p domain.ShadowPublication) bool {
			for _, x := range p.GapPoints {
				if x.ID == g.ID {
					return true
				}
			}
			return false
		}) {
			r.Dossier.LateGaps++
		}
	}
	aggregate(l, v, &r.Dossier)
	if r.Dossier.Origin == nil {
		return fail(2, "unmeasured: no inspected comparable representative")
	}
	r.Dossier.Stats.Days = int(t.Sub(*r.Dossier.Origin) / (24 * time.Hour))
	p, err := policy.Parse([]byte(fmt.Sprintf("blocking:\n  shadow: {min_days: %d, min_pr_runs: %d}\n", frontier.MinDays, frontier.MinObservations)))
	if err != nil {
		return fail(3, fmt.Sprintf("invalid frontier policy: %v", err))
	}
	r.Dossier.Ready, r.Dossier.Reason = p.ShadowReady(r.Dossier.Stats)
	if q.AsOf == nil && certificationRefused(l, v, q, c, version) {
		r.ExitCode = 2
		r.Live.Certification = "refused"
		r.Live.Reason = "pending late evidence or post-cutoff state-changing history"
	}
	return r, nil
}

func certificationRefused(l domain.ShadowLedger, v visibility, q Query, c time.Time, version string) bool {
	for _, f := range l.Facts {
		if f.ID.RepositoryID == q.RepositoryID && f.Scope == domain.ShadowScopeCI && (v.factLate(f) || f.RecordedAt == nil && !f.StartedAt.After(c)) && (f.RecordedAt == nil || f.RecordedAt.After(c)) {
			return true
		}
	}
	for _, e := range l.Events {
		if e.RepositoryID != q.RepositoryID || e.Scope != domain.ShadowScopeCI {
			continue
		}
		late := v.late(e.EventAt, func(p domain.ShadowPublication) bool {
			for _, x := range p.Events {
				if x.EventID == e.EventID {
					return true
				}
			}
			return false
		})
		if late && (e.RecordedAt == nil || e.RecordedAt.After(c)) {
			return true
		}
	}
	for _, g := range l.GapPoints {
		if g.RepositoryID == q.RepositoryID && !v.gap(g, c) && (g.RecordedAt == nil || g.RecordedAt.After(c)) && !g.At.After(c) {
			return true
		}
	}
	for _, row := range l.Triage {
		if row.RepositoryID != q.RepositoryID || row.Scope != domain.ShadowScopeCI || row.RecordedAt != nil && !row.RecordedAt.After(c) {
			continue
		}
		for _, e := range l.Events {
			if e.EventID == row.EventID && e.RepositoryID == q.RepositoryID && e.CheckID == q.CheckID && e.RuleVersion == version && e.Scope == domain.ShadowScopeCI {
				return true
			}
		}
	}
	return false
}

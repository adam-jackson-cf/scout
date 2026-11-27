package shadow

import (
	"fmt"
	"time"

	"scout/internal/domain"
	"scout/internal/policy"
)

type evaluation struct {
	l        domain.ShadowLedger
	q        Query
	r        Report
	v        visibility
	c        time.Time
	t        time.Time
	version  string
	frontier *domain.ShadowFrontier
}

type refusal struct {
	code   int
	reason string
}

// Evaluate never opens a store or consults installed policy or the network.
func Evaluate(l domain.ShadowLedger, q Query) (Report, error) {
	e := &evaluation{l: l, q: q, r: Report{Dossier: Dossier{EncodingVersion: q.EncodingVersion, RepositoryID: q.RepositoryID, CheckID: q.CheckID}, Live: LiveEnvelope{Certification: "current", ReaderTime: q.Now}}}
	if e.r.Dossier.EncodingVersion == "" {
		e.r.Dossier.EncodingVersion = EncodingVersion
	}
	for _, step := range []func() *refusal{e.checkRequest, e.selectCutoff, e.selectVersion, e.checkTriageOrdering, e.selectFrontier, e.selectInterval, e.applyGaps, e.aggregate, e.applyPolicy} {
		if stop := step(); stop != nil {
			e.r.ExitCode = stop.code
			e.r.Dossier.Reason = stop.reason
			return e.r, nil
		}
	}
	if q.AsOf == nil && certificationRefused(l, e.v, q, e.c, e.version) {
		e.r.ExitCode = 2
		e.r.Live.Certification = "refused"
		e.r.Live.Reason = "pending late evidence or post-cutoff state-changing history"
	}
	return e.r, nil
}

func (e *evaluation) checkRequest() *refusal {
	if e.r.Dossier.EncodingVersion != EncodingVersion {
		return &refusal{3, "unsupported shadow encoding"}
	}
	if err := policy.ValidateEligibleID(e.q.CheckID); err != nil {
		return &refusal{3, err.Error()}
	}
	return nil
}

func (e *evaluation) selectCutoff() *refusal {
	q := e.q
	var c time.Time
	for _, f := range e.l.Frontiers {
		if f.RepositoryID == q.RepositoryID && f.CheckID == q.CheckID && f.CompleteThrough.After(c) {
			c = f.CompleteThrough
		}
	}
	if c.IsZero() {
		return &refusal{2, "unmeasured: no complete frontier"}
	}
	t := c
	if q.AsOf != nil {
		t = *q.AsOf
		e.r.Live.Certification = "historical"
		if t.After(c) {
			if t.After(q.Now) {
				return &refusal{3, "cutoff exceeds frontier and reader time"}
			}
			return &refusal{2, "cutoff exceeds complete frontier"}
		}
	}
	e.c, e.t = c, t
	e.r.Dossier.AsOf = t
	e.r.Live.FrontierAge = q.Now.Sub(c)
	return nil
}

func (e *evaluation) queriedFact(f domain.PRRunFact) bool {
	return f.ID.RepositoryID == e.q.RepositoryID && f.CheckID == e.q.CheckID && f.Scope == domain.ShadowScopeCI
}

func (e *evaluation) latestVisibleFact() (*domain.PRRunFact, *refusal) {
	var latest *domain.PRRunFact
	for i := range e.l.Facts {
		f := &e.l.Facts[i]
		if !e.queriedFact(*f) || f.Status == "non_default_base" || f.RuleVersion == "" || !e.v.fact(*f, e.t) {
			continue
		}
		if !validID(f.ID) {
			return nil, &refusal{2, "unmeasured: unavailable numeric attempt ordering"}
		}
		if latest == nil || later(*f, *latest) {
			latest = f
		}
	}
	return latest, nil
}

func (e *evaluation) selectVersion() *refusal {
	e.v = newVisibility(e.l)
	latest, stop := e.latestVisibleFact()
	if stop != nil {
		return stop
	}
	version := e.q.RuleVersion
	if version == "" {
		if latest == nil {
			return &refusal{2, "unmeasured: no visible attempt pin"}
		}
		version = latest.RuleVersion
	}
	e.version = version
	e.r.Dossier.RuleVersion = version
	if !e.versionKnown() {
		return &refusal{3, "unknown exact rule version"}
	}
	return nil
}

func (e *evaluation) versionKnown() bool {
	known := false
	for _, f := range e.l.Facts {
		if e.queriedFact(f) && f.RuleVersion == e.version && e.v.fact(f, e.t) {
			known = true
		}
	}
	return known
}

func (e *evaluation) checkTriageOrdering() *refusal {
	q := e.q
	for _, row := range e.l.Triage {
		if row.RepositoryID != q.RepositoryID || row.Scope != domain.ShadowScopeCI || row.RecordedAt == nil || !Visible(*row.RecordedAt, false, row.RecordedAt, e.t) {
			continue
		}
		if e.triagesCohortEvent(row) {
			if _, ok := decimal(row.ActionID); !ok {
				return &refusal{2, "unmeasured: unavailable numeric triage ordering"}
			}
		}
	}
	return nil
}

func (e *evaluation) triagesCohortEvent(row domain.ShadowTriage) bool {
	for _, event := range e.l.Events {
		if event.EventID == row.EventID && event.RepositoryID == e.q.RepositoryID && event.CheckID == e.q.CheckID && event.RuleVersion == e.version {
			return true
		}
	}
	return false
}

// A frontier seals the policy for every cutoff it covers. Select the earliest
// covering record, not a later installed policy or a later reconciliation.
func (e *evaluation) selectFrontier() *refusal {
	var frontier *domain.ShadowFrontier
	for i := range e.l.Frontiers {
		f := &e.l.Frontiers[i]
		if e.coversCutoff(f) && (frontier == nil || f.CompleteThrough.Before(frontier.CompleteThrough)) {
			frontier = f
		}
	}
	if frontier == nil {
		return &refusal{2, "unmeasured: no cohort frontier"}
	}
	e.frontier = frontier
	e.r.Dossier.Frontier = *frontier
	return nil
}

func (e *evaluation) coversCutoff(f *domain.ShadowFrontier) bool {
	return f.RepositoryID == e.q.RepositoryID && f.CheckID == e.q.CheckID && f.RuleVersion == e.version && !f.CompleteThrough.Before(e.t)
}

func (e *evaluation) containsCutoff(x *domain.ShadowInterval) bool {
	return x.RepositoryID == e.q.RepositoryID && x.CheckID == e.q.CheckID && x.RuleVersion == e.version && e.t.After(x.After) && !e.t.After(x.Through)
}

func (e *evaluation) selectInterval() *refusal {
	var interval *domain.ShadowInterval
	for i := range e.l.Intervals {
		x := &e.l.Intervals[i]
		if e.containsCutoff(x) && (interval == nil || x.After.After(interval.After)) {
			interval = x
		}
	}
	if interval == nil {
		return &refusal{2, "unmeasured: no complete containing interval"}
	}
	e.r.Dossier.Interval = *interval
	// Through is clipped to the immutable selected frontier, never a mutable
	// extension of the stored interval beyond the historical query's frontier.
	if e.r.Dossier.Interval.Through.After(e.frontier.CompleteThrough) {
		e.r.Dossier.Interval.Through = e.frontier.CompleteThrough
	}
	return nil
}

func (e *evaluation) applyGaps() *refusal {
	for _, g := range e.l.GapPoints {
		if g.RepositoryID != e.q.RepositoryID || g.CheckID != e.q.CheckID || !e.v.gap(g, e.t) {
			continue
		}
		if g.At.Equal(e.t) {
			return &refusal{2, "unmeasured: cutoff is a gap point"}
		}
		if g.At.After(e.r.Dossier.Interval.After) {
			e.r.Dossier.Interval.After = g.At
		}
		if e.v.late(g.At, publishesGap(g.ID)) {
			e.r.Dossier.LateGaps++
		}
	}
	return nil
}

func publishesGap(id string) func(domain.ShadowPublication) bool {
	return func(p domain.ShadowPublication) bool {
		for _, x := range p.GapPoints {
			if x.ID == id {
				return true
			}
		}
		return false
	}
}

func publishesEvent(id string) func(domain.ShadowPublication) bool {
	return func(p domain.ShadowPublication) bool {
		for _, x := range p.Events {
			if x.EventID == id {
				return true
			}
		}
		return false
	}
}

func (e *evaluation) aggregate() *refusal {
	aggregate(e.l, e.v, &e.r.Dossier)
	if e.r.Dossier.Origin == nil {
		return &refusal{2, "unmeasured: no inspected comparable representative"}
	}
	e.r.Dossier.Stats.Days = int(e.t.Sub(*e.r.Dossier.Origin) / (24 * time.Hour))
	return nil
}

func (e *evaluation) applyPolicy() *refusal {
	p, err := policy.Parse([]byte(fmt.Sprintf("blocking:\n  shadow: {min_days: %d, min_pr_runs: %d}\n", e.frontier.MinDays, e.frontier.MinObservations)))
	if err != nil {
		return &refusal{3, fmt.Sprintf("invalid frontier policy: %v", err)}
	}
	e.r.Dossier.Ready, e.r.Dossier.Reason = p.ShadowReady(e.r.Dossier.Stats)
	return nil
}

func certificationRefused(l domain.ShadowLedger, v visibility, q Query, c time.Time, version string) bool {
	return factsPendingAfter(l, v, q, c) || eventsPendingAfter(l, v, q, c) || gapsPendingAfter(l, v, q, c) || triagePendingAfter(l, q, c, version)
}

func recordedAfter(stamp *time.Time, c time.Time) bool {
	return stamp == nil || stamp.After(c)
}

func factsPendingAfter(l domain.ShadowLedger, v visibility, q Query, c time.Time) bool {
	for _, f := range l.Facts {
		if f.ID.RepositoryID == q.RepositoryID && f.Scope == domain.ShadowScopeCI && (v.factLate(f) || f.RecordedAt == nil && !f.StartedAt.After(c)) && recordedAfter(f.RecordedAt, c) {
			return true
		}
	}
	return false
}

func eventsPendingAfter(l domain.ShadowLedger, v visibility, q Query, c time.Time) bool {
	for _, e := range l.Events {
		if e.RepositoryID != q.RepositoryID || e.Scope != domain.ShadowScopeCI {
			continue
		}
		if v.late(e.EventAt, publishesEvent(e.EventID)) && recordedAfter(e.RecordedAt, c) {
			return true
		}
	}
	return false
}

func gapsPendingAfter(l domain.ShadowLedger, v visibility, q Query, c time.Time) bool {
	for _, g := range l.GapPoints {
		if g.RepositoryID == q.RepositoryID && !v.gap(g, c) && recordedAfter(g.RecordedAt, c) && !g.At.After(c) {
			return true
		}
	}
	return false
}

func triagePendingAfter(l domain.ShadowLedger, q Query, c time.Time, version string) bool {
	for _, row := range l.Triage {
		if row.RepositoryID != q.RepositoryID || row.Scope != domain.ShadowScopeCI || row.RecordedAt != nil && !row.RecordedAt.After(c) {
			continue
		}
		if cohortEventExists(l, q, row.EventID, version) {
			return true
		}
	}
	return false
}

func cohortEventExists(l domain.ShadowLedger, q Query, eventID, version string) bool {
	for _, e := range l.Events {
		if e.EventID == eventID && e.RepositoryID == q.RepositoryID && e.CheckID == q.CheckID && e.RuleVersion == version && e.Scope == domain.ShadowScopeCI {
			return true
		}
	}
	return false
}

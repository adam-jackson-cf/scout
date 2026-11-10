// Package shadow computes read-only readiness from an immutable ledger snapshot.
package shadow

import (
	"strings"
	"time"

	"scout/internal/domain"
)

// Visible is the single coordinate/stamp rule for shadow evidence. Pending
// evidence is never visible; normal evidence is not gated by publication time.
func Visible(coordinate time.Time, late bool, stamp *time.Time, asOf time.Time) bool {
	return stamp != nil && !coordinate.After(asOf) && (!late || !stamp.After(asOf))
}

type visibility struct {
	parents map[string]time.Time
	ledger  domain.ShadowLedger
}

func newVisibility(l domain.ShadowLedger) visibility {
	v := visibility{parents: make(map[string]time.Time), ledger: l}
	pubs := make(map[string]domain.ShadowPublication, len(l.Publications))
	for _, p := range l.Publications {
		pubs[p.ID] = p
	}
	var frontier func(string) time.Time
	frontier = func(id string) time.Time {
		if c, ok := v.parents[id]; ok {
			return c
		}
		p, ok := pubs[id]
		if !ok {
			return time.Time{}
		}
		// Mark before descending so malformed cycles cannot recurse indefinitely.
		v.parents[id] = time.Time{}
		c := frontier(p.ParentID)
		for _, f := range p.Frontiers {
			if f.RepositoryID == p.RepositoryID && f.CompleteThrough.After(c) {
				c = f.CompleteThrough
			}
		}
		v.parents[id] = c
		return c
	}
	for id := range pubs {
		frontier(id)
	}
	return v
}
func (v visibility) late(coordinate time.Time, match func(domain.ShadowPublication) bool) bool {
	for _, p := range v.ledger.Publications {
		if match(p) {
			return !coordinate.After(v.parents[p.ParentID])
		}
	}
	return false
}
func (v visibility) fact(f domain.PRRunFact, t time.Time) bool {
	return Visible(f.StartedAt, v.factLate(f), f.RecordedAt, t)
}
func (v visibility) factLate(f domain.PRRunFact) bool {
	return v.late(f.StartedAt, func(p domain.ShadowPublication) bool {
		for _, x := range p.Facts {
			if x.ID == f.ID && x.CheckID == f.CheckID {
				return true
			}
		}
		return false
	})
}
func (v visibility) event(e domain.ProviderPREvent, t time.Time) bool {
	late := v.late(e.EventAt, func(p domain.ShadowPublication) bool {
		for _, x := range p.Events {
			if x.EventID == e.EventID && x.RepositoryID == e.RepositoryID {
				return true
			}
		}
		return false
	})
	return Visible(e.EventAt, late, e.RecordedAt, t)
}
func (v visibility) gap(g domain.ShadowGapPoint, t time.Time) bool {
	late := v.late(g.At, func(p domain.ShadowPublication) bool {
		for _, x := range p.GapPoints {
			if x.ID == g.ID && x.RepositoryID == g.RepositoryID {
				return true
			}
		}
		return false
	})
	return Visible(g.At, late, g.RecordedAt, t)
}
func decimal(s string) (string, bool) {
	if s == "" {
		return "", false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return "", false
		}
	}
	s = strings.TrimLeft(s, "0")
	if s == "" {
		s = "0"
	}
	return s, true
}
func numeric(a, b string) int {
	a, _ = decimal(a)
	b, _ = decimal(b)
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return strings.Compare(a, b)
}
func validID(id domain.AttemptID) bool {
	for _, s := range []string{id.RunID, id.Attempt, id.JobID} {
		if _, ok := decimal(s); !ok {
			return false
		}
	}
	return id.RepositoryID != ""
}
func compareID(a, b domain.AttemptID) int {
	if c := strings.Compare(a.RepositoryID, b.RepositoryID); c != 0 {
		return c
	}
	if c := numeric(a.RunID, b.RunID); c != 0 {
		return c
	}
	if c := numeric(a.Attempt, b.Attempt); c != 0 {
		return c
	}
	return numeric(a.JobID, b.JobID)
}
func later(a, b domain.PRRunFact) bool {
	if !a.StartedAt.Equal(b.StartedAt) {
		return a.StartedAt.After(b.StartedAt)
	}
	return compareID(a.ID, b.ID) > 0
}

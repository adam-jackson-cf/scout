package shadow

import (
	"encoding/json"
	"fmt"
	"scout/internal/domain"
	"scout/internal/state"
	"sort"
	"time"
)

// PreparePublication is called anew for each fetched CAS parent. It never uses
// a local clock, mutates sealed history or trusts raw scanner triage/frontiers.
func PreparePublication(parent domain.ShadowLedger, b domain.ShadowBatch, a PublicationAuthorization) (domain.ShadowPublication, error) {
	var p domain.ShadowPublication
	if !a.Authenticated || !a.SourceReachable || a.RepositoryID == "" || a.CurrentProtectedTip == "" || a.PublishingJobStartedAt.IsZero() || a.PublishingJobStartedAt.Nanosecond()%int(time.Millisecond) != 0 {
		return p, fmt.Errorf("publication requires authenticated reachable writer and provider job start")
	}
	if a.ChainProof.DestinationTip != "" && a.ChainProof.DestinationTip != a.ParentCommit {
		return p, fmt.Errorf("chain proof does not name fetched CAS parent")
	}
	var err error
	b, err = importMaterial(b, a.RepositoryID)
	if err != nil {
		return p, err
	}
	p.RepositoryID = a.RepositoryID
	p.RecordedAt = a.PublishingJobStartedAt.UTC()
	var c time.Time
	for _, f := range parent.Frontiers {
		if f.RepositoryID == a.RepositoryID && f.CompleteThrough.After(c) {
			c = f.CompleteThrough
		}
	}
	if t := c.Add(time.Millisecond); !c.IsZero() && t.After(p.RecordedAt) {
		p.RecordedAt = t
	}
	for _, old := range parent.Publications {
		if old.RepositoryID != a.RepositoryID {
			continue
		}
		if old.RecordedAt.Add(time.Millisecond).After(p.RecordedAt) {
			p.RecordedAt = old.RecordedAt.Add(time.Millisecond)
		}
		if p.ParentID == "" {
			p.ParentID = old.ID
		} else {
			for _, q := range parent.Publications {
				if q.ID == p.ParentID && old.RecordedAt.After(q.RecordedAt) {
					p.ParentID = old.ID
					break
				}
			}
		}
	}
	p.Inputs.Obligations = append(p.Inputs.Obligations, b.Inputs.Obligations...)
	p.Inputs.Accounting = append(p.Inputs.Accounting, b.Inputs.Accounting...)
	for _, f := range b.Ledger.Facts {
		if f.Scope != domain.ShadowScopeCI || f.ID.RepositoryID != a.RepositoryID {
			return p, fmt.Errorf("foreign fact")
		}
		found := false
		for _, old := range parent.Facts {
			if old.ID == f.ID && old.CheckID == f.CheckID {
				left, right := old, f
				left.RecordedAt = nil
				right.RecordedAt = nil
				if left != right {
					return p, fmt.Errorf("conflicting final attempt")
				}
				found = old.RecordedAt != nil
			}
		}
		if !found {
			f.RecordedAt = &p.RecordedAt
			p.Facts = append(p.Facts, f)
		}
	}
	for _, e := range b.Ledger.Events {
		if e.Scope != domain.ShadowScopeCI || e.RepositoryID != a.RepositoryID {
			return p, fmt.Errorf("foreign event")
		}
		found := false
		for _, old := range parent.Events {
			if old.EventID == e.EventID && old.RepositoryID == e.RepositoryID {
				left, right := old, e
				left.RecordedAt = nil
				right.RecordedAt = nil
				// Repeated observations link the first immutable event rather
				// than changing its provider coordinate or original H commit.
				right.EventAt = left.EventAt
				right.HeadCommit = left.HeadCommit
				if left != right {
					return p, fmt.Errorf("conflicting immutable event")
				}
				found = old.RecordedAt != nil
				e = old
			}
		}
		if !found {
			e.RecordedAt = &p.RecordedAt
			p.Events = append(p.Events, e)
		}
	}
	for _, l := range b.Ledger.Links {
		found := false
		for _, old := range parent.Links {
			if old == l {
				found = true
			}
		}
		if !found {
			p.Links = append(p.Links, l)
		}
	}
	for _, g := range b.Ledger.GapPoints {
		found := false
		for _, old := range parent.GapPoints {
			if old.ID == g.ID {
				found = true
			}
		}
		for _, old := range p.GapPoints {
			if old.ID == g.ID {
				found = true
			}
		}
		if !found {
			g.RecordedAt = &p.RecordedAt
			p.GapPoints = append(p.GapPoints, g)
		}
	}
	seen := map[string]bool{}
	events := map[string]bool{}
	for _, r := range parent.Triage {
		seen[r.ActionID] = true
	}
	for _, old := range parent.Publications {
		for _, r := range old.Inputs.ImportRefusals {
			seen[r.Original.ActionID] = true
		}
	}
	for _, r := range b.Inputs.ImportRefusals {
		seen[r.Original.ActionID] = true
	}
	for _, t := range b.AuthorizedTriage {
		r := t.Row
		if seen[r.ActionID] || events[r.EventID] {
			continue
		}
		if r.RepositoryID != a.RepositoryID || r.Scope != domain.ShadowScopeCI || r.PolicyRevision != a.CurrentProtectedTip {
			return p, fmt.Errorf("triage authorization is stale or foreign")
		}
		r.RecordedAt = &p.RecordedAt
		p.Triage = append(p.Triage, r)
		seen[r.ActionID] = true
		events[r.EventID] = true
	}
	// Import rows retain source order, not destination verdict order. One new
	// accepted/refused row per event per publication makes retries idempotent.
	var imported []struct {
		row     domain.ShadowTriage
		lineage string
	}
	for _, pending := range b.Inputs.PendingImports {
		if pending.RepositoryID != a.RepositoryID {
			return p, fmt.Errorf("foreign pending import")
		}
		rows := pending.Publication.Triage
		if pending.History != nil {
			rows = pending.History.Triage
		}
		for _, r := range rows {
			imported = append(imported, struct {
				row     domain.ShadowTriage
				lineage string
			}{r, pending.SourceLineage})
		}
	}
	sort.Slice(imported, func(i, j int) bool {
		l, r := imported[i].row, imported[j].row
		if l.RecordedAt != nil && r.RecordedAt != nil && !l.RecordedAt.Equal(*r.RecordedAt) {
			return l.RecordedAt.Before(*r.RecordedAt)
		}
		if len(l.ActionID) != len(r.ActionID) {
			return len(l.ActionID) < len(r.ActionID)
		}
		return l.ActionID < r.ActionID
	})
	for _, candidate := range imported {
		r := candidate.row
		if seen[r.ActionID] || events[r.EventID] {
			continue
		}
		seen[r.ActionID] = true
		events[r.EventID] = true
		if !a.OriginalAuthorizations[r.ActionID] || !a.Policy.Authenticated || a.Policy.Revision != a.CurrentProtectedTip || !authorizedActor(r.Actor, a.Policy) {
			p.Inputs.ImportRefusals = append(p.Inputs.ImportRefusals, domain.ShadowImportRefusal{ID: identity(a.RepositoryID, r.ActionID), RepositoryID: a.RepositoryID, Scope: domain.ShadowScopeCI, SourceLineage: candidate.lineage, Original: r, Reason: "original or current-tip authorization unavailable", PolicyRevision: a.CurrentProtectedTip, RecordedAt: p.RecordedAt})
			continue
		}
		r.RecordedAt = &p.RecordedAt
		p.Triage = append(p.Triage, r)
	}
	if b.Complete && !b.CompleteThrough.IsZero() && b.CompleteThrough.After(c) {
		cohorts := map[string]domain.ShadowFrontier{}
		for _, f := range parent.Frontiers {
			if f.RepositoryID == a.RepositoryID && (b.CheckID == "" || f.CheckID == b.CheckID) {
				cohorts[f.CheckID+"\x00"+f.RuleVersion] = f
			}
		}
		for _, f := range b.Ledger.Facts {
			cohorts[f.CheckID+"\x00"+f.RuleVersion] = domain.ShadowFrontier{CheckID: f.CheckID, RuleVersion: f.RuleVersion}
		}
		if b.CheckID != "" && b.RuleVersion != "" {
			cohorts[b.CheckID+"\x00"+b.RuleVersion] = domain.ShadowFrontier{CheckID: b.CheckID, RuleVersion: b.RuleVersion}
		}
		keys := make([]string, 0, len(cohorts))
		for k := range cohorts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			f := cohorts[k]
			f.RepositoryID = a.RepositoryID
			f.CompleteThrough = b.CompleteThrough
			f.PolicyRevision = a.CurrentProtectedTip
			f.MinDays = a.Policy.MinDays
			f.MinObservations = a.Policy.MinObservations
			if !a.Policy.Authenticated {
				f.MinDays = 14
				f.MinObservations = 5
			}
			p.Frontiers = append(p.Frontiers, f)
		}
	}
	raw, err := json.Marshal(p)
	if err != nil {
		return p, err
	}
	p.ID = identity(a.ParentCommit, string(raw))
	return p, nil
}

// Imported coordinates/links are retained, but never the source frontier or
// stamps. Their destination publication establishes lateness independently.
func importMaterial(b domain.ShadowBatch, repository string) (domain.ShadowBatch, error) {
	for _, pending := range b.Inputs.PendingImports {
		if pending.RepositoryID != repository || pending.Scope != domain.ShadowScopeCI {
			return b, fmt.Errorf("foreign pending import")
		}
		history := domain.ShadowLedger{
			Facts: pending.Publication.Facts, Events: pending.Publication.Events,
			Links: pending.Publication.Links, GapPoints: pending.Publication.GapPoints,
		}
		if pending.History != nil {
			if err := state.ValidateShadowLedger(*pending.History); err != nil {
				return b, err
			}
			history = *pending.History
		}
		b.Ledger.Facts = append(b.Ledger.Facts, history.Facts...)
		b.Ledger.Events = append(b.Ledger.Events, history.Events...)
		b.Ledger.Links = append(b.Ledger.Links, history.Links...)
		b.Ledger.GapPoints = append(b.Ledger.GapPoints, history.GapPoints...)
		if pending.Inputs != nil {
			b.Inputs.Obligations = append(b.Inputs.Obligations, pending.Inputs.Obligations...)
			b.Inputs.Accounting = append(b.Inputs.Accounting, pending.Inputs.Accounting...)
		}
	}
	return b, nil
}

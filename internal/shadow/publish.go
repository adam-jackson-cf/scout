package shadow

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"scout/internal/domain"
	"scout/internal/state"
)

// PreparePublication is called anew for each fetched CAS parent. It never uses
// a local clock, mutates sealed history or trusts raw scanner triage/frontiers.
func PreparePublication(parent domain.ShadowLedger, b domain.ShadowBatch, a PublicationAuthorization) (domain.ShadowPublication, error) {
	var p domain.ShadowPublication
	if err := checkPublicationAuthorization(a); err != nil {
		return p, err
	}
	var err error
	b, err = importMaterial(b, a.RepositoryID)
	if err != nil {
		return p, err
	}
	p.RepositoryID = a.RepositoryID
	c := parentCompleteThrough(parent, a.RepositoryID)
	p.RecordedAt = publicationStamp(parent, a, c)
	p.ParentID = parentPublication(parent, a.RepositoryID)
	p.Inputs.Obligations = append(p.Inputs.Obligations, b.Inputs.Obligations...)
	p.Inputs.Accounting = append(p.Inputs.Accounting, b.Inputs.Accounting...)
	for _, add := range []func(*domain.ShadowPublication, domain.ShadowLedger, domain.ShadowBatch, PublicationAuthorization) error{appendFacts, appendEvents, appendLinks, appendGapPoints, appendTriage} {
		if err := add(&p, parent, b, a); err != nil {
			return p, err
		}
	}
	if b.Complete && !b.CompleteThrough.IsZero() && b.CompleteThrough.After(c) {
		p.Frontiers = sealFrontiers(frontierCohorts(parent, b, a.RepositoryID), b, a)
	}
	raw, err := json.Marshal(p)
	if err != nil {
		return p, err
	}
	p.ID = identity(a.ParentCommit, string(raw))
	return p, nil
}

func checkPublicationAuthorization(a PublicationAuthorization) error {
	if !a.Authenticated || !a.SourceReachable || a.RepositoryID == "" || a.CurrentProtectedTip == "" || a.PublishingJobStartedAt.IsZero() || a.PublishingJobStartedAt.Nanosecond()%int(time.Millisecond) != 0 {
		return fmt.Errorf("publication requires authenticated reachable writer and provider job start")
	}
	if a.ChainProof.DestinationTip != "" && a.ChainProof.DestinationTip != a.ParentCommit {
		return fmt.Errorf("chain proof does not name fetched CAS parent")
	}
	return nil
}

func parentCompleteThrough(parent domain.ShadowLedger, repository string) time.Time {
	var c time.Time
	for _, f := range parent.Frontiers {
		if f.RepositoryID == repository && f.CompleteThrough.After(c) {
			c = f.CompleteThrough
		}
	}
	return c
}

func publicationStamp(parent domain.ShadowLedger, a PublicationAuthorization, c time.Time) time.Time {
	recorded := a.PublishingJobStartedAt.UTC()
	if t := c.Add(time.Millisecond); !c.IsZero() && t.After(recorded) {
		recorded = t
	}
	for _, old := range parent.Publications {
		if old.RepositoryID != a.RepositoryID {
			continue
		}
		if old.RecordedAt.Add(time.Millisecond).After(recorded) {
			recorded = old.RecordedAt.Add(time.Millisecond)
		}
	}
	return recorded
}

func parentPublication(parent domain.ShadowLedger, repository string) string {
	parentID := ""
	for _, old := range parent.Publications {
		if old.RepositoryID != repository {
			continue
		}
		if parentID == "" || recordedAfterPublication(parent.Publications, parentID, old) {
			parentID = old.ID
		}
	}
	return parentID
}

func recordedAfterPublication(publications []domain.ShadowPublication, id string, old domain.ShadowPublication) bool {
	for _, q := range publications {
		if q.ID == id && old.RecordedAt.After(q.RecordedAt) {
			return true
		}
	}
	return false
}

func appendFacts(p *domain.ShadowPublication, parent domain.ShadowLedger, b domain.ShadowBatch, a PublicationAuthorization) error {
	for _, f := range b.Ledger.Facts {
		if f.Scope != domain.ShadowScopeCI || f.ID.RepositoryID != a.RepositoryID {
			return fmt.Errorf("foreign fact")
		}
		found, err := factRecordedInParent(parent, f)
		if err != nil {
			return err
		}
		if !found {
			f.RecordedAt = &p.RecordedAt
			p.Facts = append(p.Facts, f)
		}
	}
	return nil
}

func factRecordedInParent(parent domain.ShadowLedger, f domain.PRRunFact) (bool, error) {
	found := false
	for _, old := range parent.Facts {
		if old.ID == f.ID && old.CheckID == f.CheckID {
			left, right := old, f
			left.RecordedAt = nil
			right.RecordedAt = nil
			if left != right {
				return false, fmt.Errorf("conflicting final attempt")
			}
			found = old.RecordedAt != nil
		}
	}
	return found, nil
}

func appendEvents(p *domain.ShadowPublication, parent domain.ShadowLedger, b domain.ShadowBatch, a PublicationAuthorization) error {
	for _, e := range b.Ledger.Events {
		if e.Scope != domain.ShadowScopeCI || e.RepositoryID != a.RepositoryID {
			return fmt.Errorf("foreign event")
		}
		e, found, err := eventRecordedInParent(parent, e)
		if err != nil {
			return err
		}
		if !found {
			e.RecordedAt = &p.RecordedAt
			p.Events = append(p.Events, e)
		}
	}
	return nil
}

func eventRecordedInParent(parent domain.ShadowLedger, e domain.ProviderPREvent) (domain.ProviderPREvent, bool, error) {
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
				return e, false, fmt.Errorf("conflicting immutable event")
			}
			found = old.RecordedAt != nil
			e = old
		}
	}
	return e, found, nil
}

func appendLinks(p *domain.ShadowPublication, parent domain.ShadowLedger, b domain.ShadowBatch, _ PublicationAuthorization) error {
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
	return nil
}

func appendGapPoints(p *domain.ShadowPublication, parent domain.ShadowLedger, b domain.ShadowBatch, _ PublicationAuthorization) error {
	for _, g := range b.Ledger.GapPoints {
		if !containsGapPoint(parent.GapPoints, g.ID) && !containsGapPoint(p.GapPoints, g.ID) {
			g.RecordedAt = &p.RecordedAt
			p.GapPoints = append(p.GapPoints, g)
		}
	}
	return nil
}

func containsGapPoint(points []domain.ShadowGapPoint, id string) bool {
	found := false
	for _, old := range points {
		if old.ID == id {
			found = true
		}
	}
	return found
}

type triageClaims struct {
	actions map[string]bool
	events  map[string]bool
}

func (c triageClaims) claimed(r domain.ShadowTriage) bool {
	return c.actions[r.ActionID] || c.events[r.EventID]
}

func (c triageClaims) claim(r domain.ShadowTriage) {
	c.actions[r.ActionID] = true
	c.events[r.EventID] = true
}

func priorTriageClaims(parent domain.ShadowLedger, b domain.ShadowBatch) triageClaims {
	claims := triageClaims{actions: map[string]bool{}, events: map[string]bool{}}
	for _, r := range parent.Triage {
		claims.actions[r.ActionID] = true
	}
	for _, old := range parent.Publications {
		for _, r := range old.Inputs.ImportRefusals {
			claims.actions[r.Original.ActionID] = true
		}
	}
	for _, r := range b.Inputs.ImportRefusals {
		claims.actions[r.Original.ActionID] = true
	}
	return claims
}

func appendTriage(p *domain.ShadowPublication, parent domain.ShadowLedger, b domain.ShadowBatch, a PublicationAuthorization) error {
	claims := priorTriageClaims(parent, b)
	for _, t := range b.AuthorizedTriage {
		r := t.Row
		if claims.claimed(r) {
			continue
		}
		if r.RepositoryID != a.RepositoryID || r.Scope != domain.ShadowScopeCI || r.PolicyRevision != a.CurrentProtectedTip {
			return fmt.Errorf("triage authorization is stale or foreign")
		}
		r.RecordedAt = &p.RecordedAt
		p.Triage = append(p.Triage, r)
		claims.claim(r)
	}
	return appendImportedTriage(p, b, a, claims)
}

type importedTriage struct {
	row     domain.ShadowTriage
	lineage string
}

// Import rows retain source order, not destination verdict order. One new
// accepted/refused row per event per publication makes retries idempotent.
func appendImportedTriage(p *domain.ShadowPublication, b domain.ShadowBatch, a PublicationAuthorization, claims triageClaims) error {
	imported, err := pendingImportedTriage(b, a.RepositoryID)
	if err != nil {
		return err
	}
	for _, candidate := range imported {
		r := candidate.row
		if claims.claimed(r) {
			continue
		}
		claims.claim(r)
		if !importAuthorized(a, r) {
			p.Inputs.ImportRefusals = append(p.Inputs.ImportRefusals, domain.ShadowImportRefusal{ID: identity(a.RepositoryID, r.ActionID), RepositoryID: a.RepositoryID, Scope: domain.ShadowScopeCI, SourceLineage: candidate.lineage, Original: r, Reason: "original or current-tip authorization unavailable", PolicyRevision: a.CurrentProtectedTip, RecordedAt: p.RecordedAt})
			continue
		}
		r.RecordedAt = &p.RecordedAt
		p.Triage = append(p.Triage, r)
	}
	return nil
}

func importAuthorized(a PublicationAuthorization, r domain.ShadowTriage) bool {
	return a.OriginalAuthorizations[r.ActionID] && a.Policy.Authenticated && a.Policy.Revision == a.CurrentProtectedTip && authorizedActor(r.Actor, a.Policy)
}

func pendingImportedTriage(b domain.ShadowBatch, repository string) ([]importedTriage, error) {
	var imported []importedTriage
	for _, pending := range b.Inputs.PendingImports {
		if pending.RepositoryID != repository {
			return nil, fmt.Errorf("foreign pending import")
		}
		rows := pending.Publication.Triage
		if pending.History != nil {
			rows = pending.History.Triage
		}
		for _, r := range rows {
			imported = append(imported, importedTriage{r, pending.SourceLineage})
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
	return imported, nil
}

func frontierCohorts(parent domain.ShadowLedger, b domain.ShadowBatch, repository string) map[string]domain.ShadowFrontier {
	cohorts := map[string]domain.ShadowFrontier{}
	for _, f := range parent.Frontiers {
		if f.RepositoryID == repository && (b.CheckID == "" || f.CheckID == b.CheckID) {
			cohorts[f.CheckID+"\x00"+f.RuleVersion] = f
		}
	}
	for _, f := range b.Ledger.Facts {
		cohorts[f.CheckID+"\x00"+f.RuleVersion] = domain.ShadowFrontier{CheckID: f.CheckID, RuleVersion: f.RuleVersion}
	}
	if b.CheckID != "" && b.RuleVersion != "" {
		cohorts[b.CheckID+"\x00"+b.RuleVersion] = domain.ShadowFrontier{CheckID: b.CheckID, RuleVersion: b.RuleVersion}
	}
	return cohorts
}

func sealFrontiers(cohorts map[string]domain.ShadowFrontier, b domain.ShadowBatch, a PublicationAuthorization) []domain.ShadowFrontier {
	keys := make([]string, 0, len(cohorts))
	for k := range cohorts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var frontiers []domain.ShadowFrontier
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
		frontiers = append(frontiers, f)
	}
	return frontiers
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

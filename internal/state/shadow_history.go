package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"scout/internal/domain"
)

func historyKey(parts ...string) string {
	raw, _ := json.Marshal(parts)
	return string(raw)
}
func factKey(f domain.PRRunFact) string {
	return historyKey(string(f.Scope), f.ID.RepositoryID, shadowAttemptKey(f.ID), f.CheckID)
}
func eventKey(scope domain.ShadowScope, repo, id string) string {
	return historyKey(string(scope), repo, id)
}
func frontierKey(f domain.ShadowFrontier) string {
	return historyKey(f.RepositoryID, f.CheckID, f.RuleVersion, f.CompleteThrough.Format(time.RFC3339Nano))
}

func indexHistory[T any](rows []T, key func(T) string) (map[string]T, error) {
	out := make(map[string]T, len(rows))
	for _, row := range rows {
		k := key(row)
		if old, ok := out[k]; ok && !sameHistory(old, row) {
			return nil, fmt.Errorf("conflicting immutable history %s: %w", k, ErrImmutable)
		}
		out[k] = row
	}
	return out, nil
}
func sameHistory[T any](a, b T) bool {
	left, _ := json.Marshal(a)
	right, _ := json.Marshal(b)
	return bytes.Equal(left, right)
}
func hasHistory[T any](index map[string]T, key string, row T) bool {
	old, ok := index[key]
	return ok && sameHistory(old, row)
}
func containsHistory[T any](dst, src []T, key func(T) string) bool {
	index, err := indexHistory(src, key)
	if err != nil {
		return false
	}
	for _, row := range dst {
		if !hasHistory(index, key(row), row) {
			return false
		}
	}
	return true
}
func linkKey(l domain.PRRunEventLink) string {
	return historyKey(string(l.Scope), l.AttemptID.RepositoryID, shadowAttemptKey(l.AttemptID), l.EventID)
}
func triageKey(t domain.ShadowTriage) string           { return eventKey(t.Scope, t.RepositoryID, t.ActionID) }
func pointKey(p domain.ShadowGapPoint) string          { return historyKey(p.RepositoryID, p.ID) }
func intervalKey(i domain.ShadowInterval) string       { return historyKey(i.RepositoryID, i.ID) }
func publicationKey(p domain.ShadowPublication) string { return historyKey(p.RepositoryID, p.ID) }

// ShadowHistoryContains compares every immutable row, not merely the frontier or
// pre-C evidence. The publication parent graph makes a contained chain a prefix.
// This checks bytes/structure only; the caller must separately authenticate Git
// ancestry, repository, ref and genesis before treating it as a fast-forward.
func ShadowHistoryContains(dst, src domain.ShadowLedger) bool {
	return containsHistory(dst.Facts, src.Facts, factKey) &&
		containsHistory(dst.Events, src.Events, func(e domain.ProviderPREvent) string { return eventKey(e.Scope, e.RepositoryID, e.EventID) }) &&
		containsHistory(dst.Links, src.Links, linkKey) && containsHistory(dst.Triage, src.Triage, triageKey) &&
		containsHistory(dst.Frontiers, src.Frontiers, frontierKey) && containsHistory(dst.GapPoints, src.GapPoints, pointKey) &&
		containsHistory(dst.Intervals, src.Intervals, intervalKey) && containsHistory(dst.Publications, src.Publications, publicationKey)
}

// ValidateShadowLedger rejects immutable collisions and incomplete graphs before
// any restore/adoption decision, including when the destination is empty.
func ValidateShadowLedger(l domain.ShadowLedger) error {
	ix, err := indexShadowLedger(l)
	if err != nil {
		return err
	}
	if err := ix.validateRows(); err != nil {
		return err
	}
	published := newShadowPublished()
	for _, p := range ix.publications {
		if err := ix.validatePublication(p, published); err != nil {
			return err
		}
	}
	return ix.checkPublished(published)
}

type shadowIndex struct {
	facts        map[string]domain.PRRunFact
	events       map[string]domain.ProviderPREvent
	links        map[string]domain.PRRunEventLink
	triage       map[string]domain.ShadowTriage
	frontiers    map[string]domain.ShadowFrontier
	points       map[string]domain.ShadowGapPoint
	intervals    map[string]domain.ShadowInterval
	publications map[string]domain.ShadowPublication
}

type shadowPublished struct {
	children  map[string]string
	facts     map[string]bool
	events    map[string]bool
	triage    map[string]bool
	frontiers map[string]bool
	points    map[string]bool
	intervals map[string]bool
}

func newShadowPublished() shadowPublished {
	return shadowPublished{children: map[string]string{}, facts: map[string]bool{}, events: map[string]bool{}, triage: map[string]bool{}, frontiers: map[string]bool{}, points: map[string]bool{}, intervals: map[string]bool{}}
}

func shadowEventKey(e domain.ProviderPREvent) string {
	return eventKey(e.Scope, e.RepositoryID, e.EventID)
}

func indexShadowLedger(l domain.ShadowLedger) (shadowIndex, error) {
	var ix shadowIndex
	var err error
	if ix.facts, err = indexHistory(l.Facts, factKey); err != nil {
		return ix, err
	}
	if ix.events, err = indexHistory(l.Events, shadowEventKey); err != nil {
		return ix, err
	}
	if ix.links, err = indexHistory(l.Links, linkKey); err != nil {
		return ix, err
	}
	if ix.triage, err = indexHistory(l.Triage, triageKey); err != nil {
		return ix, err
	}
	if ix.frontiers, err = indexHistory(l.Frontiers, frontierKey); err != nil {
		return ix, err
	}
	if ix.points, err = indexHistory(l.GapPoints, pointKey); err != nil {
		return ix, err
	}
	if ix.intervals, err = indexHistory(l.Intervals, intervalKey); err != nil {
		return ix, err
	}
	if ix.publications, err = indexHistory(l.Publications, publicationKey); err != nil {
		return ix, err
	}
	return ix, nil
}

func (ix shadowIndex) validateRows() error {
	if err := ix.validateFactsAndEvents(); err != nil {
		return err
	}
	if err := ix.validateLinks(); err != nil {
		return err
	}
	for _, t := range ix.triage {
		if err := ix.validateTriageRow(t); err != nil {
			return err
		}
	}
	return ix.validateCompletenessRows()
}

func (ix shadowIndex) validateFactsAndEvents() error {
	for _, f := range ix.facts {
		if err := validateShadowFact(f); err != nil {
			return err
		}
	}
	for _, e := range ix.events {
		if legacyEventPlaceholder(e) {
			continue
		}
		if err := validateShadowEvent(e); err != nil {
			return err
		}
	}
	return nil
}

func legacyEventPlaceholder(e domain.ProviderPREvent) bool {
	return e.Scope == domain.ShadowScopeLegacy && e.RepositoryID == "" && e.EventID != "" && e.RecordedAt == nil
}

func (ix shadowIndex) validateLinks() error {
	for _, link := range ix.links {
		e, ok := ix.events[eventKey(link.Scope, link.AttemptID.RepositoryID, link.EventID)]
		if !ok {
			return errors.New("dangling shadow event link")
		}
		f, ok := ix.facts[historyKey(string(link.Scope), link.AttemptID.RepositoryID, shadowAttemptKey(link.AttemptID), e.CheckID)]
		if !ok {
			return errors.New("dangling shadow attempt link")
		}
		if f.PRNumber != e.PRNumber || f.HeadTree != e.HeadTree || f.MergeBase != e.MergeBase || f.RuleVersion != e.RuleVersion {
			return errors.New("event link differs from attempt cohort/observation")
		}
	}
	return nil
}

func (ix shadowIndex) validateTriageRow(t domain.ShadowTriage) error {
	if _, ok := ix.events[eventKey(t.Scope, t.RepositoryID, t.EventID)]; !ok {
		return errors.New("dangling triage event")
	}
	if !validLedgerTriage(t) {
		return errors.New("invalid triage row")
	}
	if t.Scope == domain.ShadowScopeCI && !authorizedTriageProvenance(t) {
		return errors.New("invalid authorized triage provenance")
	}
	if t.Scope != domain.ShadowScopeCI && t.RecordedAt != nil {
		return errors.New("non-CI triage cannot have a trusted stamp")
	}
	return nil
}

func validLedgerTriage(t domain.ShadowTriage) bool {
	return t.ActionID != "" && domain.ValidTriage(t.Verdict) && shadowTime(t.TriagedAt) && validShadowScope(t.Scope)
}

func authorizedTriageProvenance(t domain.ShadowTriage) bool {
	return t.RecordedAt != nil && shadowTime(*t.RecordedAt) && t.Actor != "" && t.PolicyRevision != ""
}

func (ix shadowIndex) validateCompletenessRows() error {
	for _, f := range ix.frontiers {
		if !validLedgerFrontier(f) {
			return errors.New("invalid frontier policy")
		}
	}
	for _, p := range ix.points {
		if !validLedgerGapPoint(p) {
			return errors.New("invalid gap point")
		}
	}
	for _, i := range ix.intervals {
		if !validLedgerInterval(i) {
			return errors.New("invalid interval")
		}
	}
	return nil
}

func validLedgerFrontier(f domain.ShadowFrontier) bool {
	return f.RepositoryID != "" && f.CheckID != "" && f.RuleVersion != "" && f.PolicyRevision != "" && f.MinDays >= 0 && f.MinObservations >= 0 && shadowTime(f.CompleteThrough)
}

func validLedgerGapPoint(p domain.ShadowGapPoint) bool {
	return p.RepositoryID != "" && p.ID != "" && p.CheckID != "" && p.Reason != "" && shadowTime(p.At) && p.RecordedAt != nil && shadowTime(*p.RecordedAt)
}

func validLedgerInterval(i domain.ShadowInterval) bool {
	return i.RepositoryID != "" && i.ID != "" && i.CheckID != "" && i.RuleVersion != "" && shadowTime(i.After) && shadowTime(i.Through) && i.Through.After(i.After)
}

func (ix shadowIndex) validatePublication(p domain.ShadowPublication, published shadowPublished) error {
	if err := ix.checkPublicationChain(p, published.children); err != nil {
		return err
	}
	if err := ix.checkPublicationFacts(p, published); err != nil {
		return err
	}
	if err := ix.checkPublicationEvents(p, published); err != nil {
		return err
	}
	if err := ix.checkPublicationLinks(p); err != nil {
		return err
	}
	if err := ix.checkPublicationTriage(p, published); err != nil {
		return err
	}
	if err := ix.checkPublicationFrontiers(p, published); err != nil {
		return err
	}
	if err := ix.checkPublicationGapPoints(p, published); err != nil {
		return err
	}
	return ix.checkPublicationIntervals(p, published)
}

func (ix shadowIndex) checkPublicationChain(p domain.ShadowPublication, children map[string]string) error {
	if p.ID == "" || p.RepositoryID == "" || !shadowTime(p.RecordedAt) {
		return errors.New("invalid publication identity/stamp")
	}
	parentKey := historyKey(p.RepositoryID, p.ParentID)
	if prior, ok := children[parentKey]; ok && prior != p.ID {
		return errors.New("forked publication history")
	}
	children[parentKey] = p.ID
	if p.ParentID != "" {
		parent, ok := ix.publications[parentKey]
		if !ok || !p.RecordedAt.After(parent.RecordedAt) {
			return errors.New("missing publication parent or non-increasing stamp")
		}
	}
	return nil
}

func (ix shadowIndex) checkPublicationFacts(p domain.ShadowPublication, published shadowPublished) error {
	for _, f := range p.Facts {
		if f.ID.RepositoryID != p.RepositoryID || f.Scope != domain.ShadowScopeCI || !sameShadowStamp(f.RecordedAt, p.RecordedAt) || !hasHistory(ix.facts, factKey(f), f) {
			return errors.New("publication fact differs from immutable ledger")
		}
		published.facts[factKey(f)] = true
	}
	return nil
}

func (ix shadowIndex) checkPublicationEvents(p domain.ShadowPublication, published shadowPublished) error {
	for _, e := range p.Events {
		if e.RepositoryID != p.RepositoryID || e.Scope != domain.ShadowScopeCI || !sameShadowStamp(e.RecordedAt, p.RecordedAt) || !hasHistory(ix.events, eventKey(e.Scope, e.RepositoryID, e.EventID), e) {
			return errors.New("publication event differs from immutable ledger")
		}
		published.events[eventKey(e.Scope, e.RepositoryID, e.EventID)] = true
	}
	return nil
}

func (ix shadowIndex) checkPublicationLinks(p domain.ShadowPublication) error {
	for _, link := range p.Links {
		if link.Scope != domain.ShadowScopeCI || link.AttemptID.RepositoryID != p.RepositoryID || !hasHistory(ix.links, linkKey(link), link) {
			return errors.New("publication link differs from immutable ledger")
		}
	}
	return nil
}

func (ix shadowIndex) checkPublicationTriage(p domain.ShadowPublication, published shadowPublished) error {
	seen := map[string]bool{}
	for _, t := range p.Triage {
		if seen[t.EventID] || published.triage[triageKey(t)] || t.Scope != domain.ShadowScopeCI || t.RepositoryID != p.RepositoryID || !sameShadowStamp(t.RecordedAt, p.RecordedAt) || !hasHistory(ix.triage, triageKey(t), t) {
			return errors.New("publication triage differs from immutable ledger or repeats an action/event")
		}
		seen[t.EventID] = true
		published.triage[triageKey(t)] = true
	}
	return nil
}

func (ix shadowIndex) checkPublicationFrontiers(p domain.ShadowPublication, published shadowPublished) error {
	for _, f := range p.Frontiers {
		if f.RepositoryID != p.RepositoryID || !hasHistory(ix.frontiers, frontierKey(f), f) {
			return errors.New("publication frontier differs from immutable ledger")
		}
		published.frontiers[frontierKey(f)] = true
	}
	return nil
}

func (ix shadowIndex) checkPublicationGapPoints(p domain.ShadowPublication, published shadowPublished) error {
	for _, g := range p.GapPoints {
		if g.RepositoryID != p.RepositoryID || !sameShadowStamp(g.RecordedAt, p.RecordedAt) || !hasHistory(ix.points, pointKey(g), g) {
			return errors.New("publication point differs from immutable ledger")
		}
		published.points[pointKey(g)] = true
	}
	return nil
}

func (ix shadowIndex) checkPublicationIntervals(p domain.ShadowPublication, published shadowPublished) error {
	for _, i := range p.Intervals {
		if i.RepositoryID != p.RepositoryID || !hasHistory(ix.intervals, intervalKey(i), i) {
			return errors.New("publication interval differs from immutable ledger")
		}
		published.intervals[intervalKey(i)] = true
	}
	return nil
}

func (ix shadowIndex) checkPublished(published shadowPublished) error {
	if err := ix.checkStampedPublished(published); err != nil {
		return err
	}
	for k := range ix.frontiers {
		if !published.frontiers[k] {
			return errors.New("frontier has no publication")
		}
	}
	for k := range ix.points {
		if !published.points[k] {
			return errors.New("gap point has no publication")
		}
	}
	for k := range ix.intervals {
		if !published.intervals[k] {
			return errors.New("interval has no publication")
		}
	}
	return nil
}

func (ix shadowIndex) checkStampedPublished(published shadowPublished) error {
	for k, f := range ix.facts {
		if f.RecordedAt != nil && !published.facts[k] {
			return errors.New("stamped fact has no publication")
		}
	}
	for k, e := range ix.events {
		if e.RecordedAt != nil && !published.events[k] {
			return errors.New("stamped event has no publication")
		}
	}
	for k, t := range ix.triage {
		if t.Scope == domain.ShadowScopeCI && !published.triage[k] {
			return errors.New("authorized triage has no publication")
		}
	}
	return nil
}

func validateShadowInputs(l domain.ShadowLedger, in domain.ShadowInputs) error {
	obligations, err := indexHistory(in.Obligations, func(o domain.ProviderEventObligation) string { return eventKey(o.Scope, o.RepositoryID, o.ID) })
	if err != nil {
		return err
	}
	facts, _ := indexHistory(l.Facts, factKey)
	for _, o := range obligations {
		if !validShadowObligation(o) {
			return errors.New("invalid provider obligation")
		}
	}
	if err := validateShadowAccounting(in.Accounting, facts, obligations); err != nil {
		return err
	}
	if err := validateShadowRefusals(in.ImportRefusals); err != nil {
		return err
	}
	_, err = indexHistory(in.PendingImports, func(p domain.ShadowPendingImport) string { return eventKey(p.Scope, p.RepositoryID, p.ID) })
	if err != nil {
		return err
	}
	for _, p := range in.PendingImports {
		if err := validatePendingShadowImport(p); err != nil {
			return err
		}
	}
	return nil
}

func validateShadowAccounting(accounting []domain.PRRunObligationLink, facts map[string]domain.PRRunFact, obligations map[string]domain.ProviderEventObligation) error {
	_, err := indexHistory(accounting, func(a domain.PRRunObligationLink) string {
		return historyKey(string(a.Scope), a.AttemptID.RepositoryID, shadowAttemptKey(a.AttemptID), a.CheckID, a.ObligationID)
	})
	if err != nil {
		return err
	}
	for _, a := range accounting {
		f, ok := facts[historyKey(string(a.Scope), a.AttemptID.RepositoryID, shadowAttemptKey(a.AttemptID), a.CheckID)]
		o, found := obligations[eventKey(a.Scope, a.AttemptID.RepositoryID, a.ObligationID)]
		if !ok || !found || f.PRNumber != o.PRNumber || f.HeadCommit != o.HeadCommit || f.StartedAt.Before(o.EventAt) {
			return errors.New("dangling or mismatched provider event accounting")
		}
	}
	return nil
}

func validateShadowRefusals(refusals []domain.ShadowImportRefusal) error {
	_, err := indexHistory(refusals, func(r domain.ShadowImportRefusal) string { return eventKey(r.Scope, r.RepositoryID, r.ID) })
	if err != nil {
		return err
	}
	for _, r := range refusals {
		if !validImportRefusal(r) {
			return errors.New("invalid import refusal provenance")
		}
	}
	return nil
}

func validatePendingShadowImport(p domain.ShadowPendingImport) error {
	if p.ID == "" || p.RepositoryID == "" || p.Scope != domain.ShadowScopeCI || p.SourceLineage == "" {
		return errors.New("invalid pending import identity")
	}
	if p.History != nil {
		return validatePendingShadowHistory(p)
	}
	if p.Inputs != nil || p.Publication.ID == "" || p.Publication.RepositoryID != p.RepositoryID || !shadowTime(p.Publication.RecordedAt) {
		return errors.New("invalid pending import source")
	}
	return nil
}

func validatePendingShadowHistory(p domain.ShadowPendingImport) error {
	if p.Publication.ID != "" || p.Inputs == nil {
		return errors.New("pending whole history must not claim a publication")
	}
	if err := ValidateShadowLedger(*p.History); err != nil {
		return err
	}
	return validateShadowInputs(*p.History, *p.Inputs)
}

// ShadowInputsContains is the supplemental half of an exact replay check.
func ShadowInputsContains(dst, src domain.ShadowInputs) bool {
	return containsHistory(dst.Obligations, src.Obligations, func(o domain.ProviderEventObligation) string { return eventKey(o.Scope, o.RepositoryID, o.ID) }) &&
		containsHistory(dst.Accounting, src.Accounting, func(a domain.PRRunObligationLink) string {
			return historyKey(string(a.Scope), a.AttemptID.RepositoryID, shadowAttemptKey(a.AttemptID), a.CheckID, a.ObligationID)
		}) &&
		containsHistory(dst.ImportRefusals, src.ImportRefusals, func(r domain.ShadowImportRefusal) string { return eventKey(r.Scope, r.RepositoryID, r.ID) }) &&
		containsHistory(dst.PendingImports, src.PendingImports, func(p domain.ShadowPendingImport) string { return eventKey(p.Scope, p.RepositoryID, p.ID) })
}

// ValidateShadowOverlap checks final evidence shared by two independent chains;
// a different source stamp is provenance, not a different final assessment.
func ValidateShadowOverlap(dst, src domain.ShadowLedger) error {
	facts, err := indexHistory(dst.Facts, factKey)
	if err != nil {
		return err
	}
	for _, f := range src.Facts {
		if prior, ok := facts[factKey(f)]; ok {
			prior.RecordedAt, f.RecordedAt = nil, nil
			if !sameHistory(prior, f) {
				return ErrImmutable
			}
		}
	}
	events, err := indexHistory(dst.Events, func(e domain.ProviderPREvent) string { return eventKey(e.Scope, e.RepositoryID, e.EventID) })
	if err != nil {
		return err
	}
	for _, e := range src.Events {
		if prior, ok := events[eventKey(e.Scope, e.RepositoryID, e.EventID)]; ok {
			prior.RecordedAt, e.RecordedAt = nil, nil
			if !sameHistory(prior, e) {
				return ErrImmutable
			}
		}
	}
	return nil
}

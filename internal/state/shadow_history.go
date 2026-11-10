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
	facts, err := indexHistory(l.Facts, factKey)
	if err != nil {
		return err
	}
	events, err := indexHistory(l.Events, func(e domain.ProviderPREvent) string { return eventKey(e.Scope, e.RepositoryID, e.EventID) })
	if err != nil {
		return err
	}
	links, err := indexHistory(l.Links, linkKey)
	if err != nil {
		return err
	}
	triage, err := indexHistory(l.Triage, triageKey)
	if err != nil {
		return err
	}
	frontiers, err := indexHistory(l.Frontiers, frontierKey)
	if err != nil {
		return err
	}
	points, err := indexHistory(l.GapPoints, pointKey)
	if err != nil {
		return err
	}
	intervals, err := indexHistory(l.Intervals, intervalKey)
	if err != nil {
		return err
	}
	publications, err := indexHistory(l.Publications, publicationKey)
	if err != nil {
		return err
	}
	for _, f := range facts {
		if err := validateShadowFact(f); err != nil {
			return err
		}
	}
	for _, e := range events {
		if e.Scope == domain.ShadowScopeLegacy && e.RepositoryID == "" && e.EventID != "" && e.RecordedAt == nil {
			continue
		}
		if err := validateShadowEvent(e); err != nil {
			return err
		}
	}
	for _, link := range links {
		e, ok := events[eventKey(link.Scope, link.AttemptID.RepositoryID, link.EventID)]
		if !ok {
			return errors.New("dangling shadow event link")
		}
		f, ok := facts[historyKey(string(link.Scope), link.AttemptID.RepositoryID, shadowAttemptKey(link.AttemptID), e.CheckID)]
		if !ok {
			return errors.New("dangling shadow attempt link")
		}
		if f.PRNumber != e.PRNumber || f.HeadTree != e.HeadTree || f.MergeBase != e.MergeBase || f.RuleVersion != e.RuleVersion {
			return errors.New("event link differs from attempt cohort/observation")
		}
	}
	for _, t := range triage {
		if _, ok := events[eventKey(t.Scope, t.RepositoryID, t.EventID)]; !ok {
			return errors.New("dangling triage event")
		}
		if t.ActionID == "" || !domain.ValidTriage(t.Verdict) || !shadowTime(t.TriagedAt) || !validShadowScope(t.Scope) {
			return errors.New("invalid triage row")
		}
		if t.Scope == domain.ShadowScopeCI && (t.RecordedAt == nil || !shadowTime(*t.RecordedAt) || t.Actor == "" || t.PolicyRevision == "") {
			return errors.New("invalid authorized triage provenance")
		}
		if t.Scope != domain.ShadowScopeCI && t.RecordedAt != nil {
			return errors.New("non-CI triage cannot have a trusted stamp")
		}
	}
	for _, f := range frontiers {
		if f.RepositoryID == "" || f.CheckID == "" || f.RuleVersion == "" || f.PolicyRevision == "" || f.MinDays < 0 || f.MinObservations < 0 || !shadowTime(f.CompleteThrough) {
			return errors.New("invalid frontier policy")
		}
	}
	for _, p := range points {
		if p.RepositoryID == "" || p.ID == "" || p.CheckID == "" || p.Reason == "" || !shadowTime(p.At) || p.RecordedAt == nil || !shadowTime(*p.RecordedAt) {
			return errors.New("invalid gap point")
		}
	}
	for _, i := range intervals {
		if i.RepositoryID == "" || i.ID == "" || i.CheckID == "" || i.RuleVersion == "" || !shadowTime(i.After) || !shadowTime(i.Through) || !i.Through.After(i.After) {
			return errors.New("invalid interval")
		}
	}
	children := map[string]string{}
	publishedFacts, publishedEvents, publishedTriage := map[string]bool{}, map[string]bool{}, map[string]bool{}
	publishedFrontiers, publishedPoints, publishedIntervals := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, p := range publications {
		if p.ID == "" || p.RepositoryID == "" || !shadowTime(p.RecordedAt) {
			return errors.New("invalid publication identity/stamp")
		}
		parentKey := historyKey(p.RepositoryID, p.ParentID)
		if prior, ok := children[parentKey]; ok && prior != p.ID {
			return errors.New("forked publication history")
		}
		children[parentKey] = p.ID
		if p.ParentID != "" {
			parent, ok := publications[parentKey]
			if !ok || !p.RecordedAt.After(parent.RecordedAt) {
				return errors.New("missing publication parent or non-increasing stamp")
			}
		}
		seen := map[string]bool{}
		for _, f := range p.Facts {
			if f.ID.RepositoryID != p.RepositoryID || f.Scope != domain.ShadowScopeCI || !sameShadowStamp(f.RecordedAt, p.RecordedAt) || !hasHistory(facts, factKey(f), f) {
				return errors.New("publication fact differs from immutable ledger")
			}
			publishedFacts[factKey(f)] = true
		}
		for _, e := range p.Events {
			if e.RepositoryID != p.RepositoryID || e.Scope != domain.ShadowScopeCI || !sameShadowStamp(e.RecordedAt, p.RecordedAt) || !hasHistory(events, eventKey(e.Scope, e.RepositoryID, e.EventID), e) {
				return errors.New("publication event differs from immutable ledger")
			}
			publishedEvents[eventKey(e.Scope, e.RepositoryID, e.EventID)] = true
		}
		for _, link := range p.Links {
			if link.Scope != domain.ShadowScopeCI || link.AttemptID.RepositoryID != p.RepositoryID || !hasHistory(links, linkKey(link), link) {
				return errors.New("publication link differs from immutable ledger")
			}
		}
		for _, t := range p.Triage {
			if seen[t.EventID] || publishedTriage[triageKey(t)] || t.Scope != domain.ShadowScopeCI || t.RepositoryID != p.RepositoryID || !sameShadowStamp(t.RecordedAt, p.RecordedAt) || !hasHistory(triage, triageKey(t), t) {
				return errors.New("publication triage differs from immutable ledger or repeats an action/event")
			}
			seen[t.EventID] = true
			publishedTriage[triageKey(t)] = true
		}
		for _, f := range p.Frontiers {
			if f.RepositoryID != p.RepositoryID || !hasHistory(frontiers, frontierKey(f), f) {
				return errors.New("publication frontier differs from immutable ledger")
			}
			publishedFrontiers[frontierKey(f)] = true
		}
		for _, g := range p.GapPoints {
			if g.RepositoryID != p.RepositoryID || !sameShadowStamp(g.RecordedAt, p.RecordedAt) || !hasHistory(points, pointKey(g), g) {
				return errors.New("publication point differs from immutable ledger")
			}
			publishedPoints[pointKey(g)] = true
		}
		for _, i := range p.Intervals {
			if i.RepositoryID != p.RepositoryID || !hasHistory(intervals, intervalKey(i), i) {
				return errors.New("publication interval differs from immutable ledger")
			}
			publishedIntervals[intervalKey(i)] = true
		}
	}
	for k, f := range facts {
		if f.RecordedAt != nil && !publishedFacts[k] {
			return errors.New("stamped fact has no publication")
		}
	}
	for k, e := range events {
		if e.RecordedAt != nil && !publishedEvents[k] {
			return errors.New("stamped event has no publication")
		}
	}
	for k, t := range triage {
		if t.Scope == domain.ShadowScopeCI && !publishedTriage[k] {
			return errors.New("authorized triage has no publication")
		}
	}
	for k := range frontiers {
		if !publishedFrontiers[k] {
			return errors.New("frontier has no publication")
		}
	}
	for k := range points {
		if !publishedPoints[k] {
			return errors.New("gap point has no publication")
		}
	}
	for k := range intervals {
		if !publishedIntervals[k] {
			return errors.New("interval has no publication")
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
		if o.ID == "" || o.RepositoryID == "" || !validShadowScope(o.Scope) || !shadowDecimal(o.PRNumber) || o.HeadCommit == "" || !shadowTime(o.EventAt) || o.Selector == "" || o.ProviderProof == "" {
			return errors.New("invalid provider obligation")
		}
	}
	_, err = indexHistory(in.Accounting, func(a domain.PRRunObligationLink) string {
		return historyKey(string(a.Scope), a.AttemptID.RepositoryID, shadowAttemptKey(a.AttemptID), a.CheckID, a.ObligationID)
	})
	if err != nil {
		return err
	}
	for _, a := range in.Accounting {
		f, ok := facts[historyKey(string(a.Scope), a.AttemptID.RepositoryID, shadowAttemptKey(a.AttemptID), a.CheckID)]
		o, found := obligations[eventKey(a.Scope, a.AttemptID.RepositoryID, a.ObligationID)]
		if !ok || !found || f.PRNumber != o.PRNumber || f.HeadCommit != o.HeadCommit || f.StartedAt.Before(o.EventAt) {
			return errors.New("dangling or mismatched provider event accounting")
		}
	}
	_, err = indexHistory(in.ImportRefusals, func(r domain.ShadowImportRefusal) string { return eventKey(r.Scope, r.RepositoryID, r.ID) })
	if err != nil {
		return err
	}
	for _, r := range in.ImportRefusals {
		if r.ID == "" || r.RepositoryID == "" || r.Scope != domain.ShadowScopeCI || r.SourceLineage == "" || r.Original.ActionID == "" || r.Original.EventID == "" || r.Original.RepositoryID != r.RepositoryID || r.Original.Scope != domain.ShadowScopeCI || r.Reason == "" || r.PolicyRevision == "" || !shadowTime(r.RecordedAt) {
			return errors.New("invalid import refusal provenance")
		}
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

func validatePendingShadowImport(p domain.ShadowPendingImport) error {
	if p.ID == "" || p.RepositoryID == "" || p.Scope != domain.ShadowScopeCI || p.SourceLineage == "" {
		return errors.New("invalid pending import identity")
	}
	if p.History != nil {
		if p.Publication.ID != "" || p.Inputs == nil {
			return errors.New("pending whole history must not claim a publication")
		}
		if err := ValidateShadowLedger(*p.History); err != nil {
			return err
		}
		return validateShadowInputs(*p.History, *p.Inputs)
	}
	if p.Inputs != nil || p.Publication.ID == "" || p.Publication.RepositoryID != p.RepositoryID || !shadowTime(p.Publication.RecordedAt) {
		return errors.New("invalid pending import source")
	}
	return nil
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

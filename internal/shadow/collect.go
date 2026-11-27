package shadow

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"time"

	"scout/internal/domain"
)

func identity(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(strconv.Itoa(len(p)) + ":" + p))
	}
	return hex.EncodeToString(h.Sum(nil))
}

type collector struct {
	req       CollectionRequest
	snapshot  CollectionSnapshot
	batch     domain.ShadowBatch
	accounted map[string]bool
	emitted   map[string]bool
}

// Collect keeps eligible events independent of run/finding population. Missing
// proof can hold a bound or produce an explicit deadline gap, never an exclusion.
func Collect(ctx context.Context, p Provider, req CollectionRequest) (domain.ShadowBatch, error) {
	var b domain.ShadowBatch
	s, err := p.Enumerate(ctx, req)
	if err != nil {
		return b, err
	}
	if !populationMatches(req, s) {
		return b, fmt.Errorf("invalid collection population")
	}
	c := newCollector(req, s)
	sortAttempts(c.snapshot.Attempts)
	for _, a := range c.snapshot.Attempts {
		if err := c.attempt(a); err != nil {
			return c.batch, err
		}
	}
	for _, e := range c.snapshot.Events {
		if err := c.obligation(e); err != nil {
			return c.batch, err
		}
	}
	return c.batch, nil
}

func populationMatches(req CollectionRequest, s CollectionSnapshot) bool {
	return req.RepositoryID != "" && s.RepositoryID == req.RepositoryID && req.Selector != "" && s.Selector == req.Selector && req.ConsistencyLag >= 0
}

func sortAttempts(attempts []CollectionAttempt) {
	sort.Slice(attempts, func(i, j int) bool {
		left, right := attempts[i].Fact, attempts[j].Fact
		if !left.StartedAt.Equal(right.StartedAt) {
			return left.StartedAt.Before(right.StartedAt)
		}
		return compareID(left.ID, right.ID) < 0
	})
}

func newCollector(req CollectionRequest, s CollectionSnapshot) *collector {
	c := &collector{req: req, snapshot: s, accounted: map[string]bool{}, emitted: map[string]bool{}}
	b := &c.batch
	b.Complete = snapshotEnumerated(s) && snapshotTimed(s)
	b.CheckID = req.CheckID
	b.RuleVersion = req.RuleVersion
	if b.Complete {
		b.CompleteThrough = s.ProviderTime
		for _, at := range []time.Time{s.EventCompleteThrough, s.AttemptCompleteThrough} {
			if at.Before(b.CompleteThrough) {
				b.CompleteThrough = at
			}
		}
		b.CompleteThrough = b.CompleteThrough.Add(-req.ConsistencyLag)
	}
	return c
}

func snapshotEnumerated(s CollectionSnapshot) bool {
	return s.Complete && s.PagesComplete && !s.ListingCapped && s.EnumerationError == "" && s.EnumerationProof != ""
}

func snapshotTimed(s CollectionSnapshot) bool {
	return !s.ProviderTime.IsZero() && !s.EventCompleteThrough.IsZero() && !s.AttemptCompleteThrough.IsZero() && s.TimeProof != "" && s.LagJustification != ""
}

func (c *collector) hold(at time.Time) {
	b := &c.batch
	if b.Complete && !at.IsZero() && !b.CompleteThrough.Before(at) {
		b.CompleteThrough = at.Add(-time.Millisecond)
	}
}

func (c *collector) gap(at time.Time, key string) {
	if at.IsZero() {
		return
	}
	req := c.req
	c.batch.Ledger.GapPoints = append(c.batch.Ledger.GapPoints, domain.ShadowGapPoint{ID: identity(req.RepositoryID, req.CheckID, key, at.String()), RepositoryID: req.RepositoryID, CheckID: req.CheckID, At: at, Reason: "collection deadline or unaccounted obligation"})
}

func (c *collector) trustedAttempt(f domain.PRRunFact) bool {
	return f.ID.RepositoryID == c.req.RepositoryID && validID(f.ID) && f.Scope == domain.ShadowScopeCI && f.CheckID == c.req.CheckID && f.RuleVersion == c.req.RuleVersion && !f.StartedAt.IsZero()
}

func (c *collector) attempt(a CollectionAttempt) error {
	f := a.Fact
	if !c.trustedAttempt(f) {
		return fmt.Errorf("attempt lacks trusted population binding")
	}
	f.RecordedAt = nil
	f, final, excluded := attemptOutcome(a, f)
	if excluded {
		c.batch.Ledger.Facts = append(c.batch.Ledger.Facts, f)
	}
	bound := c.boundEvents(a, f)
	if final || excluded {
		f = c.account(f, bound)
	}
	switch {
	case final:
		c.recordFinal(a, f)
	case !excluded:
		c.recordUnfinished(a, f, bound)
	}
	return nil
}

func attemptOutcome(a CollectionAttempt, f domain.PRRunFact) (domain.PRRunFact, bool, bool) {
	excluded := excludedAttempt(a)
	final := a.Terminal && f.Status != "cancelled"
	if final && f.Status != "assessed" && f.Status != "not_assessed" {
		f.Status, f.Comparison, f.Inspection = "assessed", "none", "not_inspected"
	}
	if excluded {
		final = false
		f.Status, f.Comparison, f.Inspection = "not_assessed", "excluded", "not_inspected"
	}
	return f, final, excluded
}

func (c *collector) boundEvents(a CollectionAttempt, f domain.PRRunFact) []CollectionEvent {
	var bound []CollectionEvent
	for _, e := range c.snapshot.Events {
		if namesEvent(a.AccountedEventIDs, e.ID) && e.RepositoryID == f.ID.RepositoryID && e.PRNumber == f.PRNumber && e.HeadCommit == f.HeadCommit && !f.StartedAt.Before(e.EventAt) {
			bound = append(bound, e)
		}
	}
	return bound
}

func namesEvent(ids []string, eventID string) bool {
	named := false
	for _, id := range ids {
		if id == eventID {
			named = true
		}
	}
	return named
}

func (c *collector) account(f domain.PRRunFact, bound []CollectionEvent) domain.PRRunFact {
	for _, e := range bound {
		c.accounted[e.ID] = true
		c.batch.Inputs.Accounting = append(c.batch.Inputs.Accounting, domain.PRRunObligationLink{AttemptID: f.ID, Scope: domain.ShadowScopeCI, CheckID: f.CheckID, ObligationID: e.ID})
		if e.BaseBranch != c.snapshot.Trigger.BaseBranch || !e.BaseReachable {
			f.Status = "not_assessed"
			f.Comparison = "non_default_base"
			f.Inspection = "not_inspected"
		}
	}
	return f
}

func (c *collector) recordFinal(a CollectionAttempt, f domain.PRRunFact) {
	b := &c.batch
	b.Ledger.Facts = append(b.Ledger.Facts, f)
	for _, claim := range a.ScannerOutput.Events {
		if claim.IssueID == "" || claim.CheckID != f.CheckID || claim.RuleVersion != f.RuleVersion {
			continue
		}
		e := domain.ProviderPREvent{EventID: domain.ShadowEventID(f.MergeBase, f.HeadTree, f.CheckID, claim.IssueID, f.RuleVersion), RepositoryID: c.req.RepositoryID, Scope: domain.ShadowScopeCI, PRNumber: f.PRNumber, HeadCommit: f.HeadCommit, HeadTree: f.HeadTree, MergeBase: f.MergeBase, CheckID: f.CheckID, IssueID: claim.IssueID, RuleVersion: f.RuleVersion, EventAt: f.StartedAt}
		if !c.emitted[e.EventID] {
			b.Ledger.Events = append(b.Ledger.Events, e)
			c.emitted[e.EventID] = true
		}
		b.Ledger.Links = append(b.Ledger.Links, domain.PRRunEventLink{AttemptID: f.ID, Scope: domain.ShadowScopeCI, EventID: e.EventID})
	}
}

func (c *collector) recordUnfinished(a CollectionAttempt, f domain.PRRunFact, bound []CollectionEvent) {
	if a.Terminal {
		f.Comparison, f.Inspection = "none", "not_inspected"
		c.batch.Ledger.Facts = append(c.batch.Ledger.Facts, f)
	}
	if !c.pastDeadline(a.DeadlineAt) {
		c.hold(f.StartedAt)
		return
	}
	c.gap(f.StartedAt, fmt.Sprint(f.ID))
	for _, e := range bound {
		c.gap(e.EventAt, e.ID)
	}
}

func (c *collector) pastDeadline(deadline time.Time) bool {
	if deadline.IsZero() {
		deadline = c.req.Deadline
	}
	return !deadline.IsZero() && !c.snapshot.ProviderTime.Before(deadline)
}

func (c *collector) attestedObligation(e CollectionEvent) bool {
	return e.RepositoryID == c.req.RepositoryID && e.Scope == domain.ShadowScopeCI && e.Selector == c.req.Selector && e.ID != "" && !e.EventAt.IsZero() && e.ProviderProof != ""
}

func (c *collector) obligation(e CollectionEvent) error {
	if !c.attestedObligation(e) {
		return fmt.Errorf("unattested event obligation")
	}
	c.batch.Inputs.Obligations = append(c.batch.Inputs.Obligations, e.ProviderEventObligation)
	if c.accounted[e.ID] {
		return nil
	}
	if c.pastDeadline(e.DeadlineAt) || e.State == "closed" || e.State == "merged" {
		c.gap(e.EventAt, e.ID)
	} else {
		c.hold(e.EventAt)
	}
	return nil
}

func excludedAttempt(a CollectionAttempt) bool {
	switch a.CancellationCause {
	case "merge-conflict":
		return a.MergeConflictProven && !a.MergeResultAvailable
	case "system":
		return a.SystemCancellationProven
	case "configured-skip":
		return a.ConfiguredSkipProven && a.ConfigurationVerified && a.ExecutedConfigDigest == a.PinnedConfigDigest
	}
	return providerSupersession(a)
}

func providerSupersession(a CollectionAttempt) bool {
	cause := a.CancellationCause == "provider-concurrency" || (a.CancellationCause == "" && a.CancellationActor == "")
	return cause && a.Fact.Status == "cancelled" && originalConfigPinned(a) && successorConfigPinned(a) && successorSharesLane(a) && a.SuccessorCreatedAt.After(a.Fact.StartedAt) && a.SuccessorCreatedAt.Before(a.CancelledAt)
}

func originalConfigPinned(a CollectionAttempt) bool {
	return a.ConfigurationVerified && a.CancelInProgress && a.PinnedConfigDigest != "" && a.ExecutedConfigDigest == a.PinnedConfigDigest
}

func successorConfigPinned(a CollectionAttempt) bool {
	return a.SuccessorConfigurationVerified && a.SuccessorCancelInProgress && a.SuccessorExecutedConfigDigest == a.PinnedConfigDigest && a.SuccessorPinnedConfigDigest == a.PinnedConfigDigest
}

func successorSharesLane(a CollectionAttempt) bool {
	return a.SuccessorPRNumber == a.Fact.PRNumber && a.SuccessorHeadCommit != "" && a.SuccessorLane == a.Lane && a.Lane != "" && a.SuccessorConcurrencyGroup == a.ConcurrencyGroup && a.ConcurrencyGroup != ""
}

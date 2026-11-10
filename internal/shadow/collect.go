package shadow

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"scout/internal/domain"
	"sort"
	"time"
)

func identity(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		fmt.Fprintf(h, "%d:%s", len(p), p)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Collect keeps eligible events independent of run/finding population. Missing
// proof can hold a bound or produce an explicit deadline gap, never an exclusion.
func Collect(ctx context.Context, p Provider, req CollectionRequest) (domain.ShadowBatch, error) {
	var b domain.ShadowBatch
	s, err := p.Enumerate(ctx, req)
	if err != nil {
		return b, err
	}
	if req.RepositoryID == "" || s.RepositoryID != req.RepositoryID || req.Selector == "" || s.Selector != req.Selector || req.ConsistencyLag < 0 {
		return b, fmt.Errorf("invalid collection population")
	}
	b.Complete = s.Complete && s.PagesComplete && !s.ListingCapped && s.EnumerationError == "" && !s.ProviderTime.IsZero() && !s.EventCompleteThrough.IsZero() && !s.AttemptCompleteThrough.IsZero() && s.EnumerationProof != "" && s.TimeProof != "" && s.LagJustification != ""
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
	hold := func(at time.Time) {
		if b.Complete && !at.IsZero() && !b.CompleteThrough.Before(at) {
			b.CompleteThrough = at.Add(-time.Millisecond)
		}
	}
	gap := func(at time.Time, key string) {
		if at.IsZero() {
			return
		}
		b.Ledger.GapPoints = append(b.Ledger.GapPoints, domain.ShadowGapPoint{ID: identity(req.RepositoryID, req.CheckID, key, at.String()), RepositoryID: req.RepositoryID, CheckID: req.CheckID, At: at, Reason: "collection deadline or unaccounted obligation"})
	}
	accounted := map[string]bool{}
	sort.Slice(s.Attempts, func(i, j int) bool {
		left, right := s.Attempts[i].Fact, s.Attempts[j].Fact
		if !left.StartedAt.Equal(right.StartedAt) {
			return left.StartedAt.Before(right.StartedAt)
		}
		return compareID(left.ID, right.ID) < 0
	})
	emitted := map[string]bool{}
	for _, a := range s.Attempts {
		f := a.Fact
		if f.ID.RepositoryID != req.RepositoryID || !validID(f.ID) || f.Scope != domain.ShadowScopeCI || f.CheckID != req.CheckID || f.RuleVersion != req.RuleVersion || f.StartedAt.IsZero() {
			return b, fmt.Errorf("attempt lacks trusted population binding")
		}
		f.RecordedAt = nil
		excluded := excludedAttempt(a)
		final := a.Terminal && f.Status != "cancelled"
		if final && f.Status != "assessed" && f.Status != "not_assessed" {
			f.Status, f.Comparison, f.Inspection = "assessed", "none", "not_inspected"
		}
		if excluded {
			final = false
			f.Status, f.Comparison, f.Inspection = "not_assessed", "excluded", "not_inspected"
			b.Ledger.Facts = append(b.Ledger.Facts, f)
		}
		var bound []CollectionEvent
		for _, e := range s.Events {
			named := false
			for _, id := range a.AccountedEventIDs {
				if id == e.ID {
					named = true
				}
			}
			if named && e.RepositoryID == f.ID.RepositoryID && e.PRNumber == f.PRNumber && e.HeadCommit == f.HeadCommit && !f.StartedAt.Before(e.EventAt) {
				bound = append(bound, e)
			}
		}
		if final || excluded {
			for _, e := range bound {
				accounted[e.ID] = true
				b.Inputs.Accounting = append(b.Inputs.Accounting, domain.PRRunObligationLink{AttemptID: f.ID, Scope: domain.ShadowScopeCI, CheckID: f.CheckID, ObligationID: e.ID})
				if e.BaseBranch != s.Trigger.BaseBranch || !e.BaseReachable {
					f.Status = "not_assessed"
					f.Comparison = "non_default_base"
					f.Inspection = "not_inspected"
				}
			}
		}
		if final {
			b.Ledger.Facts = append(b.Ledger.Facts, f)
			for _, claim := range a.ScannerOutput.Events {
				if claim.IssueID == "" || claim.CheckID != f.CheckID || claim.RuleVersion != f.RuleVersion {
					continue
				}
				e := domain.ProviderPREvent{EventID: domain.ShadowEventID(f.MergeBase, f.HeadTree, f.CheckID, claim.IssueID, f.RuleVersion), RepositoryID: req.RepositoryID, Scope: domain.ShadowScopeCI, PRNumber: f.PRNumber, HeadCommit: f.HeadCommit, HeadTree: f.HeadTree, MergeBase: f.MergeBase, CheckID: f.CheckID, IssueID: claim.IssueID, RuleVersion: f.RuleVersion, EventAt: f.StartedAt}
				if !emitted[e.EventID] {
					b.Ledger.Events = append(b.Ledger.Events, e)
					emitted[e.EventID] = true
				}
				b.Ledger.Links = append(b.Ledger.Links, domain.PRRunEventLink{AttemptID: f.ID, Scope: domain.ShadowScopeCI, EventID: e.EventID})
			}
		} else if !excluded {
			if a.Terminal {
				f.Comparison, f.Inspection = "none", "not_inspected"
				b.Ledger.Facts = append(b.Ledger.Facts, f)
			}
			deadline := a.DeadlineAt
			if deadline.IsZero() {
				deadline = req.Deadline
			}
			if !deadline.IsZero() && !s.ProviderTime.Before(deadline) {
				gap(f.StartedAt, fmt.Sprint(f.ID))
				for _, e := range bound {
					gap(e.EventAt, e.ID)
				}
			} else {
				hold(f.StartedAt)
			}
		}
	}
	for _, e := range s.Events {
		if e.RepositoryID != req.RepositoryID || e.Scope != domain.ShadowScopeCI || e.Selector != req.Selector || e.ID == "" || e.EventAt.IsZero() || e.ProviderProof == "" {
			return b, fmt.Errorf("unattested event obligation")
		}
		b.Inputs.Obligations = append(b.Inputs.Obligations, e.ProviderEventObligation)
		if accounted[e.ID] {
			continue
		}
		deadline := e.DeadlineAt
		if deadline.IsZero() {
			deadline = req.Deadline
		}
		if (!deadline.IsZero() && !s.ProviderTime.Before(deadline)) || e.State == "closed" || e.State == "merged" {
			gap(e.EventAt, e.ID)
		} else {
			hold(e.EventAt)
		}
	}
	return b, nil
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
	cause := a.CancellationCause == "provider-concurrency" || (a.CancellationCause == "" && a.CancellationActor == "")
	return cause && a.Fact.Status == "cancelled" && a.ConfigurationVerified && a.CancelInProgress && a.PinnedConfigDigest != "" && a.ExecutedConfigDigest == a.PinnedConfigDigest && a.SuccessorConfigurationVerified && a.SuccessorCancelInProgress && a.SuccessorExecutedConfigDigest == a.PinnedConfigDigest && a.SuccessorPinnedConfigDigest == a.PinnedConfigDigest && a.SuccessorPRNumber == a.Fact.PRNumber && a.SuccessorHeadCommit != "" && a.SuccessorLane == a.Lane && a.Lane != "" && a.SuccessorConcurrencyGroup == a.ConcurrencyGroup && a.ConcurrencyGroup != "" && a.SuccessorCreatedAt.After(a.Fact.StartedAt) && a.SuccessorCreatedAt.Before(a.CancelledAt)
}

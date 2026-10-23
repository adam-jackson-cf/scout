package state

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"time"

	"scout/internal/domain"
)

// IssueMatch is one correlator decision that a run's observations belong to an
// existing durable issue (design §3.3). The issue ID never changes.
type IssueMatch struct {
	IssueID        string
	ObservationIDs []string
	// Anchor is the issue's anchor after the match: the rename-followed path and
	// the current fingerprint. Ranges are locators on observations, not here.
	Anchor domain.IssueAnchor
	// Files and FeatureIDs are secondary labels re-derived each run.
	Files      []string
	FeatureIDs []string
	// Drift records that the evidence fingerprint changed. The disposition is
	// kept and only an evidence_drift history entry is appended.
	Drift  bool
	Actor  string
	Reason string
}

// RecordIssueMatch links new observations to an issue, follows its anchor,
// records evidence drift without touching the human disposition, and moves a
// resolved issue to regressed. It is idempotent for an already-linked run.
func (s *Store) RecordIssueMatch(ctx context.Context, match IssueMatch) (domain.Issue, error) {
	if match.IssueID == "" || len(match.ObservationIDs) == 0 || match.Reason == "" {
		return domain.Issue{}, errors.New("issue ID, observations, and audit reason are required")
	}
	for _, observationID := range match.ObservationIDs {
		if observationID == "" {
			return domain.Issue{}, errors.New("observation ID must not be empty")
		}
	}
	var updated domain.Issue
	err := s.withTransaction(ctx, func(tx *sql.Tx) error {
		issue, err := loadIssueTx(ctx, tx, match.IssueID)
		if err != nil {
			return err
		}
		linked := make(map[string]struct{}, len(issue.ObservationIDs))
		for _, observationID := range issue.ObservationIDs {
			linked[observationID] = struct{}{}
		}
		newIDs := make([]string, 0, len(match.ObservationIDs))
		for _, observationID := range uniqueSorted(match.ObservationIDs) {
			if _, found := linked[observationID]; !found {
				newIDs = append(newIDs, observationID)
			}
		}
		if len(newIDs) == 0 {
			updated = issue
			return nil
		}
		issue.ObservationIDs = append(issue.ObservationIDs, newIDs...)
		issue.Anchor = match.Anchor
		if len(match.Files) != 0 {
			issue.Files = uniqueSorted(match.Files)
		}
		issue.FeatureIDs = uniqueSorted(match.FeatureIDs)
		now := time.Now().UTC()
		issue.UpdatedAt = now
		previous := issue.State
		current := issue.State
		if match.Drift {
			issue.History = append(issue.History, domain.IssueHistoryEntry{At: now, From: current, To: current, Event: domain.EventEvidenceDrift, Actor: match.Actor, Reason: "evidence fingerprint changed; disposition kept", ObservationIDs: newIDs})
		}
		if previous == domain.IssueResolved {
			current = domain.IssueRegressed
			issue.State = current
			issue.Verification = nil
			issue.History = append(issue.History, domain.IssueHistoryEntry{At: now, From: previous, To: current, Event: domain.EventRegressed, Actor: match.Actor, Reason: "failure mode reappeared after resolution", ObservationIDs: newIDs})
		} else if !match.Drift {
			issue.History = append(issue.History, domain.IssueHistoryEntry{At: now, From: current, To: current, Event: domain.EventMatched, Actor: match.Actor, Reason: match.Reason, ObservationIDs: newIDs})
		}
		if err := issue.Validate(); err != nil {
			return fmt.Errorf("updated issue: %w", err)
		}
		if err := saveIssueTx(ctx, tx, issue, false); err != nil {
			return err
		}
		updated = issue
		return nil
	})
	return updated, err
}

// MarkIssueAmbiguous records that issueID cannot be told apart from the other
// issues. The issues stay separate and open; they are never auto-merged.
func (s *Store) MarkIssueAmbiguous(ctx context.Context, issueID string, others []string, actor string) (domain.Issue, error) {
	if issueID == "" || len(others) == 0 {
		return domain.Issue{}, errors.New("issue ID and ambiguous peers are required")
	}
	var updated domain.Issue
	err := s.withTransaction(ctx, func(tx *sql.Tx) error {
		issue, err := loadIssueTx(ctx, tx, issueID)
		if err != nil {
			return err
		}
		merged := uniqueSorted(append(append([]string(nil), issue.AmbiguousWith...), others...))
		filtered := merged[:0]
		for _, id := range merged {
			if id != issueID {
				filtered = append(filtered, id)
			}
		}
		if reflect.DeepEqual([]string(filtered), issue.AmbiguousWith) {
			updated = issue
			return nil
		}
		sort.Strings(filtered)
		issue.AmbiguousWith = append([]string(nil), filtered...)
		issue.UpdatedAt = time.Now().UTC()
		issue.History = append(issue.History, domain.IssueHistoryEntry{At: issue.UpdatedAt, From: issue.State, To: issue.State, Event: domain.EventAmbiguous, Actor: actor, Reason: "cannot be told apart from other occurrences; not merged"})
		if err := issue.Validate(); err != nil {
			return fmt.Errorf("updated issue: %w", err)
		}
		if err := saveIssueTx(ctx, tx, issue, false); err != nil {
			return err
		}
		updated = issue
		return nil
	})
	return updated, err
}

// ResolveIssue resolves an issue after a relevant successful re-check: the
// verification run must hold inspected, covered or flagged coverage of the
// issue's own file (followed to followPath when the file was renamed, empty
// otherwise), family and category. Skipped, blocked, incomplete or reused
// coverage never resolves. Dispositions false_positive and wont_fix are kept.
func (s *Store) ResolveIssue(ctx context.Context, issueID string, verification domain.Verification, followPath, actor, reason string) (domain.Issue, error) {
	if issueID == "" || reason == "" {
		return domain.Issue{}, errors.New("issue ID and audit reason are required")
	}
	if err := verification.ValidateResolution(); err != nil {
		return domain.Issue{}, fmt.Errorf("resolution verification: %w", err)
	}
	var updated domain.Issue
	err := s.withTransaction(ctx, func(tx *sql.Tx) error {
		issue, err := loadIssueTx(ctx, tx, issueID)
		if err != nil {
			return err
		}
		if issue.State == domain.IssueResolved {
			updated = issue
			return nil
		}
		if issue.State == domain.IssueFalsePositive || issue.State == domain.IssueWontFix {
			return fmt.Errorf("issue %q keeps its %q disposition", issueID, issue.State)
		}
		if followPath != "" && followPath != issue.Anchor.Path {
			issue.Anchor.Path = followPath
			issue.Files = []string{followPath}
		}
		if verification.Reason == domain.ResolutionFileDeleted {
			if err := fileDeletedTx(ctx, tx, issue.Anchor.Path, verification.RunID); err != nil {
				return err
			}
		} else if err := verificationCoverageRelevantTx(ctx, tx, issue, verification); err != nil {
			return err
		}
		previous := issue.State
		issue.State = domain.IssueResolved
		issue.Verification = &verification
		issue.UpdatedAt = time.Now().UTC()
		issue.History = append(issue.History, domain.IssueHistoryEntry{At: issue.UpdatedAt, From: previous, To: domain.IssueResolved, Event: domain.EventResolved, Actor: actor, Reason: reason, ArtifactRefs: append([]domain.SafeArtifactRef(nil), verification.ArtifactRefs...)})
		if err := issue.Validate(); err != nil {
			return fmt.Errorf("updated issue: %w", err)
		}
		if err := saveIssueTx(ctx, tx, issue, false); err != nil {
			return err
		}
		updated = issue
		return nil
	})
	return updated, err
}

// fileDeletedTx proves from the run's own persisted inventory that the run was
// complete and no inventory entry (eligible or excluded) exists for the path.
func fileDeletedTx(ctx context.Context, tx *sql.Tx, path, runID string) error {
	var runState domain.RunState
	var payload []byte
	err := tx.QueryRowContext(ctx, `SELECT r.state, i.payload FROM runs r JOIN inventory_snapshots i ON i.run_id = r.run_id WHERE r.run_id = ?`, runID).Scan(&runState, &payload)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("run %q has no inventory snapshot: %w", runID, ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("load run inventory: %w", err)
	}
	if runState != domain.RunComplete {
		return errors.New("file deletion resolves only through a complete run")
	}
	var files []domain.File
	if err := decode(payload, &files); err != nil {
		return err
	}
	for _, file := range files {
		if file.Path == path {
			return fmt.Errorf("file %q is still present in the run inventory", path)
		}
	}
	return nil
}

package shadow

import (
	"context"
	"encoding/json"
	"scout/internal/domain"
	"time"
)

// Provider supplies authenticated control-plane evidence, never scanner claims.
type Provider interface {
	Enumerate(context.Context, CollectionRequest) (CollectionSnapshot, error)
	AuthenticateWriter(context.Context) (WriterIdentity, error)
	ProtectedTip(context.Context, string) (ProtectedPolicy, error)
	ProveChain(context.Context, ChainRequest) (domain.ChainProof, error)
}
type CollectionRequest struct {
	RepositoryID, CheckID, RuleVersion, Selector string
	Now, Deadline                                time.Time
	ConsistencyLag                               time.Duration
}
type TriggerPopulation struct {
	Events     []string
	BaseBranch string
	Paths      []string
}
type CollectionEvent struct {
	domain.ProviderEventObligation
	BaseBranch    string
	BaseReachable bool
	Author        string
	Fork          bool
	State         string
	DeadlineAt    time.Time
}
type ScannerMeasurement struct {
	Status string
	Events []ScannerEvent
}
type ScannerEvent struct{ IssueID, CheckID, RuleVersion string }
type CollectionAttempt struct {
	Fact                                                                                      domain.PRRunFact
	AccountedEventIDs                                                                         []string
	Terminal                                                                                  bool
	DeadlineAt                                                                                time.Time
	Lane, ConcurrencyGroup, PinnedConfigDigest, ExecutedConfigDigest                          string
	CancelInProgress, ConfigurationVerified                                                   bool
	CancelledAt                                                                               time.Time
	CancellationCause, CancellationActor, CommitMessage                                       string
	SuccessorPRNumber, SuccessorHeadCommit, SuccessorLane, SuccessorConcurrencyGroup          string
	SuccessorCreatedAt, SuccessorAttemptStartedAt                                             time.Time
	SuccessorAttempt, SuccessorExecutedConfigDigest, SuccessorPinnedConfigDigest              string
	SuccessorConfigurationVerified, SuccessorCancelInProgress                                 bool
	MergeConflictProven, MergeResultAvailable, SystemCancellationProven, ConfiguredSkipProven bool
	ScannerOutput                                                                             ScannerMeasurement
}
type CollectionSnapshot struct {
	RepositoryID                                                    string
	ProviderTime, EventCompleteThrough, AttemptCompleteThrough      time.Time
	Complete, PagesComplete, ListingCapped                          bool
	Selector                                                        string
	Trigger                                                         TriggerPopulation
	Events                                                          []CollectionEvent
	Attempts                                                        []CollectionAttempt
	EnumerationProof, TimeProof, LagJustification, EnumerationError string
}
type WriterIdentity struct {
	RepositoryID, OriginalActor, ExecutingActor, SourceRevision string
	DefaultBranch, SourceReachable, Authenticated               bool
}
type ProtectedPolicy struct {
	RepositoryID, Revision   string
	Maintainers              []string
	CODEOWNERS               string
	MinDays, MinObservations int
	Authenticated            bool
}
type ChainRequest struct{ RepositoryID, StateRef, GenesisCommit, SourceTip, DestinationTip string }
type TriageAction struct {
	RepositoryID, DispatchRunID, EventID, Verdict, Actor, Reason string
	OccurredAt                                                   time.Time
}
type PublicationAuthorization struct {
	RepositoryID, ParentCommit, CurrentProtectedTip string
	PublishingJobStartedAt                          time.Time
	Authenticated, SourceReachable                  bool
	Policy                                          ProtectedPolicy
	ChainProof                                      domain.ChainProof
	OriginalAuthorizations                          map[string]bool
}

// UnmarshalJSON preserves unavailable original run creation as unknown.
func (a *CollectionAttempt) UnmarshalJSON(raw []byte) error {
	type attempt CollectionAttempt
	wire := struct {
		*attempt
		SuccessorCreatedAt string
	}{attempt: (*attempt)(a)}
	if err := json.Unmarshal(raw, &wire); err != nil {
		return err
	}
	a.SuccessorCreatedAt = time.Time{}
	if wire.SuccessorCreatedAt != "" {
		at, err := time.Parse(time.RFC3339Nano, wire.SuccessorCreatedAt)
		if err != nil {
			return err
		}
		a.SuccessorCreatedAt = at
	}
	return nil
}

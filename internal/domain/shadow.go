package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

// ShadowScope prevents local assessments and migrated history from certifying CI.
type ShadowScope string

const (
	ShadowScopeCI     ShadowScope = "ci"
	ShadowScopeLocal  ShadowScope = "local"
	ShadowScopeLegacy ShadowScope = "legacy"
)

// AttemptID retains provider decimal IDs without floating-point conversion.
type AttemptID struct {
	RepositoryID string
	RunID        string
	JobID        string
	Attempt      string
}

// ShadowPins identifies the exact recipe and policy used by an assessment.
type ShadowPins struct {
	Scout     string
	Tools     string
	Catalogue string
	Policy    string
	Recipe    string
}

// PRRunFact is an immutable final attempt, including zero-event failures.
type PRRunFact struct {
	ID               AttemptID
	Scope            ShadowScope
	PRNumber         string
	HeadCommit       string
	HeadTree         string
	BaseCommit       string
	MergeBase        string
	AssessedRevision string
	CheckID          string
	RuleVersion      string
	CreatedAt        time.Time
	StartedAt        time.Time
	FinishedAt       time.Time
	Status           string
	Comparison       string
	Inspection       string
	Pins             ShadowPins
	RecordedAt       *time.Time
}

// ProviderPREvent is an immutable provider-bound shadow finding.
type ProviderPREvent struct {
	EventID      string
	RepositoryID string
	Scope        ShadowScope
	PRNumber     string
	HeadCommit   string
	HeadTree     string
	MergeBase    string
	CheckID      string
	IssueID      string
	RuleVersion  string
	EventAt      time.Time
	RecordedAt   *time.Time
}

// PRRunEventLink records every attempt that observed an immutable event.
type PRRunEventLink struct {
	AttemptID AttemptID
	Scope     ShadowScope
	EventID   string
}

// ShadowTriage preserves the append-only verdict and its authorization provenance.
type ShadowTriage struct {
	ActionID       string
	EventID        string
	RepositoryID   string
	Scope          ShadowScope
	Verdict        string
	Actor          string
	Reason         string
	PolicyRevision string
	TriagedAt      time.Time
	RecordedAt     *time.Time
}

// ShadowFrontier retains the effective policy at each completeness boundary.
type ShadowFrontier struct {
	RepositoryID    string
	CheckID         string
	RuleVersion     string
	CompleteThrough time.Time
	PolicyRevision  string
	MinDays         int
	MinObservations int
}

type ShadowGapPoint struct {
	ID           string
	RepositoryID string
	CheckID      string
	At           time.Time
	Reason       string
	RecordedAt   *time.Time
}

type ShadowInterval struct {
	ID           string
	RepositoryID string
	CheckID      string
	RuleVersion  string
	After        time.Time
	Through      time.Time
}

// ShadowPublication is assembled by the authenticated control plane, not scanners.
type ShadowPublication struct {
	ID           string
	ParentID     string
	RepositoryID string
	RecordedAt   time.Time
	Facts        []PRRunFact
	Events       []ProviderPREvent
	Links        []PRRunEventLink
	Triage       []ShadowTriage
	Frontiers    []ShadowFrontier
	GapPoints    []ShadowGapPoint
	Intervals    []ShadowInterval
	Inputs       ShadowInputs `json:",omitzero"`
}

// ShadowLedger is a consistent persistence snapshot. No frontier means unmeasured;
// neither local nor legacy facts/triage establish CI completeness or authority.
type ShadowLedger struct {
	Facts        []PRRunFact
	Events       []ProviderPREvent
	Links        []PRRunEventLink
	Triage       []ShadowTriage
	Frontiers    []ShadowFrontier
	GapPoints    []ShadowGapPoint
	Intervals    []ShadowInterval
	Publications []ShadowPublication
}

// ChainProof is supplied by a trusted provider adapter, never by imported state.
// An ancestry assertion without the repository, ref and genesis is insufficient.
type ChainProof struct {
	RepositoryID        string
	StateRef            string
	GenesisCommit       string
	SourceTip           string
	DestinationTip      string
	Authenticated       bool
	DestinationAncestor bool
}

// ShadowImportPlan separates verbatim restoration from material that still needs
// destination authorization and CAS publication. Pending retains source provenance.
type ShadowImportPlan struct {
	BytePreserving bool
	Adopt          ShadowLedger
	Pending        ShadowLedger
}

// AuthorizedTriage is produced by trusted current-policy authorization.
type AuthorizedTriage struct {
	Row ShadowTriage
}

// ShadowBatch is unsealed collection material awaiting a CAS publication.
type ShadowBatch struct {
	CheckID, RuleVersion string
	Ledger               ShadowLedger
	Inputs               ShadowInputs
	AuthorizedTriage     []AuthorizedTriage
	Complete             bool
	CompleteThrough      time.Time
}

// ShadowEventID binds a finding to its observation coordinates and exact rule.
// B and the assessed merge result are deliberately not identity coordinates.
func ShadowEventID(base, head, check, issue, version string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{base, head, check, issue, version}, "\x00")))
	return "wb-" + hex.EncodeToString(sum[:8])
}

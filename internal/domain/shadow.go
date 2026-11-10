package domain

import "time"

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

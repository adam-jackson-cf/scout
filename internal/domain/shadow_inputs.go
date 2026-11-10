package domain

import "time"

// ProviderEventObligation records the enumerated population independently of
// started runs or emitted findings. Selector changes establish new coverage;
// absence of a corresponding accounting link is not evidence of exclusion.
type ProviderEventObligation struct {
	ID            string
	RepositoryID  string
	Scope         ShadowScope
	PRNumber      string
	HeadCommit    string
	EventAt       time.Time
	Selector      string
	ProviderProof string
}

// PRRunObligationLink durably names each provider event accounted by an attempt.
// These are separate from PRRunEventLink, which links would-block findings.
type PRRunObligationLink struct {
	AttemptID    AttemptID
	Scope        ShadowScope
	CheckID      string
	ObligationID string
}

// ShadowImportRefusal is an immutable trusted decision, not a triage verdict.
// The original row and lineage remain available for later authorized handling.
type ShadowImportRefusal struct {
	ID             string
	RepositoryID   string
	Scope          ShadowScope
	SourceLineage  string
	Original       ShadowTriage
	Reason         string
	PolicyRevision string
	RecordedAt     time.Time
}

// ShadowPendingImport preserves source history awaiting trusted publication.
// Publishing never edits/deletes this record; published history proves progress.
type ShadowPendingImport struct {
	ID            string
	RepositoryID  string
	Scope         ShadowScope
	SourceLineage string
	Publication   ShadowPublication
	// History/Inputs retain a whole S2 import, including unpublished material and
	// refusals. Publication is used for individual S1 publication inputs.
	History *ShadowLedger `json:",omitempty"`
	Inputs  *ShadowInputs `json:",omitempty"`
}

// ShadowInputs contains supplemental immutable input records. Scope lives on
// each record; local or legacy material can never be upgraded by changing it.
type ShadowInputs struct {
	Obligations    []ProviderEventObligation
	Accounting     []PRRunObligationLink
	ImportRefusals []ShadowImportRefusal
	PendingImports []ShadowPendingImport
}

package domain

import (
	"errors"
	"fmt"
	"time"
)

// AssociationOutcome is how a merge-time cluster or file was decided (design §4).
type AssociationOutcome string

const (
	// AssociationAssigned: an `existing` answer (Jev or fallback) assigned the cluster to areas.
	AssociationAssigned AssociationOutcome = "assigned"
	// AssociationPendingNew: a fallback `new` answer recorded a pending area.
	AssociationPendingNew AssociationOutcome = "pending_new"
	// AssociationInherited: a file inherited its owners by the rename or variant rule; no model call.
	AssociationInherited AssociationOutcome = "inherited"
	// AssociationUnresolved: no decision; Reason says why.
	AssociationUnresolved AssociationOutcome = "unresolved"
)

// Reasons an association stays unresolved (Reason on an unresolved record).
const (
	AssociationReasonJevFailure    = "jev_failure"
	AssociationReasonJevCap        = "jev_cap"
	AssociationReasonPRCap         = "pr_cap"
	AssociationReasonFallbackCap   = "fallback_cap"
	AssociationReasonTerraFailure  = "terra_failure"
	AssociationReasonNoneUnsent    = "none_unsent"
	AssociationReasonPendingCap    = "pending_cap"
	AssociationReasonPermutationNo = "permutation_disagreement"
)

// Inherited rules (InheritedRule on an inherited record).
const (
	InheritedRename  = "rename"
	InheritedVariant = "variant"
)

// PRMetadata values: whether `--fetch-pr-metadata` data reached the decision.
const (
	PRMetadataAbsent  = "absent"
	PRMetadataPresent = "present"
)

// Association is one merge-time decision record: one per decided cluster per
// run and one per inherited file. Records are append-only apart from the
// Stale and Superseded flags. Jev request bodies in RawAnswer are redacted (§6).
type Association struct {
	ID                 string     `json:"id"`
	RunID              string     `json:"run_id"`
	DecidedAt          time.Time  `json:"decided_at"`
	LastAgentRefreshAt *time.Time `json:"last_agent_refresh_at"`

	MergeCommit string   `json:"merge_commit"`
	ClusterKey  string   `json:"cluster_key"`
	Files       []string `json:"files"`

	JevCovered       *float64 `json:"jev_covered,omitempty"`
	JevChoice        *float64 `json:"jev_choice,omitempty"`
	JevOutcome       string   `json:"jev_outcome,omitempty"`
	EscalationReason string   `json:"escalation_reason,omitempty"`

	CLIName         string `json:"cli_name,omitempty"`
	CLIVersion      string `json:"cli_version,omitempty"`
	ReportedModel   string `json:"reported_model,omitempty"`
	ReportedEffort  string `json:"reported_effort,omitempty"`
	PromptDigest    string `json:"prompt_digest,omitempty"`
	InputsDigest    string `json:"inputs_digest,omitempty"`
	CatalogueDigest string `json:"catalogue_digest,omitempty"`
	RawAnswer       string `json:"raw_answer,omitempty"`
	Retries         int    `json:"retries,omitempty"`
	LatencyMillis   int64  `json:"latency_ms,omitempty"`

	Outcome AssociationOutcome `json:"outcome"`
	Reason  string             `json:"reason,omitempty"`
	// AreaIDs are the areas an assigned record labels the files with.
	AreaIDs []string `json:"area_ids,omitempty"`
	// PendingAreaID is the pending area a pending_new record created.
	PendingAreaID string `json:"pending_area_id,omitempty"`

	InheritedRule    string   `json:"inherited_rule,omitempty"`
	InheritedFrom    string   `json:"inherited_from,omitempty"`
	InheritedAreaIDs []string `json:"inherited_area_ids,omitempty"`

	FallbackSkipped    bool   `json:"fallback_skipped"`
	FallbackSkipReason string `json:"fallback_skip_reason,omitempty"`

	Stale                 bool   `json:"stale"`
	Superseded            bool   `json:"superseded"`
	ThresholdsProvisional bool   `json:"thresholds_provisional"`
	PRMetadata            string `json:"pr_metadata"`
}

// Validate checks the invariants a stored record must hold.
func (a Association) Validate() error {
	switch {
	case a.ID == "" || a.RunID == "" || a.MergeCommit == "" || a.DecidedAt.IsZero():
		return errors.New("association requires id, run id, merge commit and decision time")
	case len(a.Files) == 0:
		return fmt.Errorf("association %q has no files", a.ID)
	case a.PRMetadata != PRMetadataAbsent && a.PRMetadata != PRMetadataPresent:
		return fmt.Errorf("association %q: pr_metadata %q is not absent or present", a.ID, a.PRMetadata)
	}
	switch a.Outcome {
	case AssociationAssigned:
		if len(a.AreaIDs) == 0 {
			return fmt.Errorf("association %q: assigned without areas", a.ID)
		}
	case AssociationPendingNew:
		if a.PendingAreaID == "" {
			return fmt.Errorf("association %q: pending_new without a pending area", a.ID)
		}
	case AssociationInherited:
		if a.InheritedRule == "" || len(a.InheritedAreaIDs) == 0 {
			return fmt.Errorf("association %q: inherited without a rule and areas", a.ID)
		}
	case AssociationUnresolved:
		if a.Reason == "" {
			return fmt.Errorf("association %q: unresolved without a reason", a.ID)
		}
	default:
		return fmt.Errorf("association %q: unknown outcome %q", a.ID, a.Outcome)
	}
	return nil
}

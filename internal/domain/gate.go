package domain

import (
	"errors"
	"time"
)

// LaneTrigger names when a lane runs (design §5.1).
type LaneTrigger string

const (
	TriggerPR      LaneTrigger = "pr"
	TriggerPush    LaneTrigger = "push"
	TriggerNightly LaneTrigger = "nightly"
)

// LaneMode is what a lane's findings may do.
type LaneMode string

const (
	ModeBlockingEligible LaneMode = "blocking-eligible"
	ModeReport           LaneMode = "report"
	ModeSignal           LaneMode = "signal"
)

// LaneRequires is the prerequisite a lane needs to run at all.
type LaneRequires string

const (
	RequiresNone           LaneRequires = "none"
	RequiresTool           LaneRequires = "tool"
	RequiresProviderEgress LaneRequires = "provider-egress"
	RequiresDecisionEgress LaneRequires = "decision-egress"
)

// LaneStatus is the health of one lane on one run.
type LaneStatus string

const (
	LaneRan     LaneStatus = "ran"
	LaneReused  LaneStatus = "reused"
	LaneSkipped LaneStatus = "skipped"
	LaneBlocked LaneStatus = "blocked"
	LaneFailed  LaneStatus = "failed"
)

// LaneHealth reports whether a lane produced trustworthy results on head.
// Required marks a blocking-eligible lane of the current trigger: only those
// can make a run incomplete.
type LaneHealth struct {
	Lane     string       `json:"lane"`
	Mode     LaneMode     `json:"mode"`
	Requires LaneRequires `json:"requires"`
	Required bool         `json:"required"`
	Status   LaneStatus   `json:"status"`
	Units    int          `json:"units"`
	Reason   string       `json:"reason,omitempty"`
}

// Attribution classifies a head issue against a same-policy base (design §7.3).
type Attribution string

const (
	AttributionIntroduced  Attribution = "introduced"
	AttributionRegressed   Attribution = "regressed"
	AttributionPreExisting Attribution = "pre-existing"
	AttributionUnknown     Attribution = "unknown"
)

// Comparison states how the base side was obtained.
type Comparison string

const (
	ComparisonBaseline   Comparison = "baseline"
	ComparisonRecomputed Comparison = "recomputed"
	ComparisonNone       Comparison = "none"
)

// AttributedIssue is one head issue with its gate-relevant facts.
type AttributedIssue struct {
	IssueID       string      `json:"issueId"`
	CheckID       string      `json:"checkId,omitempty"`
	RuleVersion   string      `json:"ruleVersion,omitempty"`
	Path          string      `json:"path,omitempty"`
	Title         string      `json:"title"`
	Severity      Severity    `json:"severity"`
	State         IssueState  `json:"state"`
	Attribution   Attribution `json:"attribution"`
	ChangedInPR   bool        `json:"changedInPr"`
	CheckStatus   string      `json:"checkStatus,omitempty"`
	Blocking      bool        `json:"blocking"`
	WouldBlock    bool        `json:"wouldBlock"`
	Overridden    bool        `json:"overridden,omitempty"`
	RankingScore  float64     `json:"rankingScore"`
	RankingReason string      `json:"rankingReason,omitempty"`
}

// Triage values for would_block rows.
const (
	TriageUndecided    = "undecided"
	TriageWarranted    = "warranted"
	TriageNotWarranted = "not_warranted"
)

// WouldBlock records that an issue in an eligible-but-not-active check would
// have gated the run had the check been active.
type WouldBlock struct {
	ID           string      `json:"id"`
	HeadRevision string      `json:"headRevision"`
	BaseRevision string      `json:"baseRevision"`
	IssueID      string      `json:"issueId"`
	CheckID      string      `json:"checkId"`
	RuleVersion  string      `json:"ruleVersion,omitempty"`
	Path         string      `json:"path,omitempty"`
	Attribution  Attribution `json:"attribution"`
	Profile      string      `json:"profile"`
	Triage       string      `json:"triage"`
	TriageActor  string      `json:"triageActor,omitempty"`
	TriageReason string      `json:"triageReason,omitempty"`
	RecordedAt   time.Time   `json:"recordedAt"`
	TriagedAt    *time.Time  `json:"triagedAt,omitempty"`
}

// ValidTriage reports whether v is a maintainer triage value.
func ValidTriage(v string) bool {
	return v == TriageUndecided || v == TriageWarranted || v == TriageNotWarranted
}

func (w WouldBlock) Validate() error {
	if w.ID == "" || w.IssueID == "" || w.CheckID == "" || w.HeadRevision == "" || w.BaseRevision == "" {
		return errors.New("would_block requires id, issue, check, and revisions")
	}
	if !ValidTriage(w.Triage) {
		return errors.New("would_block triage must be warranted, not_warranted or undecided")
	}
	return nil
}

// OverrideResult records one override request and its verdict.
type OverrideResult struct {
	CheckID  string   `json:"checkId"`
	Actor    string   `json:"actor"`
	Accepted bool     `json:"accepted"`
	Reason   string   `json:"reason"`
	IssueIDs []string `json:"issueIds,omitempty"`
}

// Gate is the differential gate verdict.
type Gate struct {
	Status     string           `json:"status"` // clean|gated|incomplete|unavailable
	ExitCode   int              `json:"exitCode"`
	Profile    string           `json:"profile"`
	Comparison Comparison       `json:"comparison"`
	Blocking   []string         `json:"blocking"` // issue IDs that gated
	Reasons    []string         `json:"reasons,omitempty"`
	Overrides  []OverrideResult `json:"overrides,omitempty"`
}

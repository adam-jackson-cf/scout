package domain

import (
	"encoding/json"
	"errors"
	"time"
)

// Annotation is one decision-model answer about one unit (design §6). It is
// advisory: it can change a ranking, never an issue state or the gate.
type Annotation struct {
	AnnotationID string `json:"annotationId"`
	RunID        string `json:"runId"`
	IssueID      string `json:"issueId"`
	UnitKey      string `json:"unitKey"`
	QuestionID   string `json:"questionId"`
	// QuestionVersion is question_id@sha256 over instructions, criteria,
	// state shape, option-order rule and the resolved model.
	QuestionVersion string `json:"questionVersion"`
	Kind            string `json:"kind"`
	ModelRequested  string `json:"modelRequested"`
	ModelResolved   string `json:"modelResolved"`
	// Probability is the stored full-precision output used for ordinal ranking
	// (P of the question's positive answer).
	Probability float64         `json:"probability"`
	Answer      json.RawMessage `json:"answer"`
	// StateDigest is the SHA-256 (hex) of the exact redacted state sent for
	// the unit, so rows asked over the same state can be grouped and a stored
	// state can be checked against its row.
	StateDigest string `json:"stateDigest,omitempty"`
	// LatencyMS is the end-to-end latency of the unit's call in milliseconds,
	// on the first row of a fanned-out unit only, like Usage.
	LatencyMS int64 `json:"latencyMs,omitempty"`
	// RequestBody is the exact redacted request that was sent, plus the
	// realised option order.
	RequestBody json.RawMessage `json:"requestBody"`
	OptionOrder []string        `json:"optionOrder,omitempty"`
	Usage       json.RawMessage `json:"usage,omitempty"`
	// Stale marks an annotation whose question version is no longer current.
	Stale     bool      `json:"stale"`
	CreatedAt time.Time `json:"createdAt"`
}

// Validate checks the identity fields a stored annotation must carry.
func (a Annotation) Validate() error {
	if a.AnnotationID == "" || a.RunID == "" || a.IssueID == "" || a.UnitKey == "" || a.QuestionVersion == "" || a.ModelResolved == "" {
		return errors.New("annotation requires identity, run, unit, question version and resolved model")
	}
	if len(a.RequestBody) == 0 || len(a.Answer) == 0 {
		return errors.New("annotation requires the stored request body and answer")
	}
	return nil
}

// Ranking modes of one finding.
const (
	RankingModeJev           = "jev"
	RankingModeDeterministic = "deterministic"
	RankingModePartial       = "partial"
)

// RankingRecord is the ranking disclosure stored for one finding on one run,
// with the baseline and rank-formula versions design §6 requires on every one.
type RankingRecord struct {
	RunID   string `json:"runId"`
	IssueID string `json:"issueId"`
	Mode    string `json:"mode"`
	// Disclosure is the per-finding text: `deterministic (reason)`,
	// `partial (n/N)` or `jev (unvalidated, X1 pending)`.
	Disclosure         string    `json:"disclosure"`
	BaselineScore      float64   `json:"baselineScore"`
	BaselineVersion    string    `json:"baselineVersion"`
	RankFormulaVersion string    `json:"rankFormulaVersion"`
	QuestionVersion    string    `json:"questionVersion,omitempty"`
	Probability        *float64  `json:"probability,omitempty"`
	RankPosition       int       `json:"rankPosition"`
	CreatedAt          time.Time `json:"createdAt"`
}

// AuditDraw is one interim audit-sample draw (design §6, "Audit"): the stored
// inclusion probability is the realised one, made under SamplerVersion.
type AuditDraw struct {
	RunID                string    `json:"runId"`
	IssueID              string    `json:"issueId"`
	FormulaVersion       string    `json:"formulaVersion"`
	SamplerVersion       string    `json:"samplerVersion"`
	Stratum              string    `json:"stratum"`
	InclusionProbability float64   `json:"inclusionProbability"`
	Drawn                bool      `json:"drawn"`
	RankPosition         int       `json:"rankPosition"`
	CreatedAt            time.Time `json:"createdAt"`
}

// Dispositions are recorded with the view they were made in.
const (
	ViewCLI   = "cli"
	ViewAudit = "audit"
)

// DispositionContext records, for one human disposition, the view it was made
// in, whether the annotation was visible, and the versions then current.
type DispositionContext struct {
	IssueID            string    `json:"issueId"`
	At                 time.Time `json:"at"`
	View               string    `json:"view"`
	AnnotationVisible  bool      `json:"annotationVisible"`
	RankFormulaVersion string    `json:"rankFormulaVersion,omitempty"`
	BaselineVersion    string    `json:"baselineVersion,omitempty"`
}

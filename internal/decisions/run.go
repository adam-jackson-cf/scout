package decisions

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"scout/internal/domain"
	"scout/internal/policy"
)

// LaneID is the decisions lane's id in lane health. The lane is optional
// (design §7.3): its health never changes the exit code.
const LaneID = "decisions"

// Config is the decisions consent and pin, as read from `.scout/policy.yaml`.
type Config struct {
	Enabled bool
	// Model is the pinned Jev model id; aliases are rejected.
	Model     string
	MaxPerRun int
	// MaxCostUSD caps the billed cost of one run; it is checked before each
	// call, so a run can exceed it by at most the call in flight.
	MaxCostUSD         float64
	EgressAcknowledged bool
	Endpoint           string
}

// Validate rejects a missing pin, an alias or a non-positive cap. The caps
// default in policy; a zero here is a wiring error, not a request for defaults.
func (c Config) Validate() error {
	if c.Model == "" {
		return errors.New("decisions model is required")
	}
	if policy.IsModelAlias(c.Model) {
		return fmt.Errorf("decisions model %q is an alias; pin an exact model id", c.Model)
	}
	if c.MaxPerRun <= 0 {
		return fmt.Errorf("decisions max per run must be positive, got %d", c.MaxPerRun)
	}
	if !(c.MaxCostUSD > 0) || math.IsInf(c.MaxCostUSD, 0) {
		return fmt.Errorf("decisions max cost must be a positive number, got %v", c.MaxCostUSD)
	}
	return nil
}

// Eligibility is what permits a Jev ranking on this run.
type Eligibility struct {
	Enabled       bool
	HasKey        bool
	EgressAllowed bool
	Fork          bool
}

// Skip returns why Jev must not run; ok is true when it may. A fork pull
// request is reported first: it can never run Jev, whatever the consent.
func (e Eligibility) Skip() (reason string, ok bool) {
	switch {
	case e.Fork:
		return "fork", false
	case !e.Enabled:
		return "decisions not enabled", false
	case !e.EgressAllowed:
		return "egress not allowed", false
	case !e.HasKey:
		return "no jev key", false
	}
	return "", true
}

// Families of units, in the order the cap consumes them (design §6): triage
// first, then rubric functions, commit units, unowned-file units.
type Family int

const (
	FamilyTriage Family = iota
	FamilyRubric
	FamilyCommit
	FamilyUnowned
)

// Unit is one set of independent questions about one thing, asked in one
// request (fan-out).
type Unit struct {
	Family    Family
	Key       string // stable unit key; for triage the issue id
	IssueID   string
	Questions []Question
	State     json.RawMessage
	// Baseline is the deterministic baseline score, used to order triage units.
	Baseline float64
}

// RubricKey is the stable key of a rubric function unit: changed-file path,
// then function start line.
func RubricKey(path string, startLine int) string {
	return fmt.Sprintf("%s\x00%010d", path, startLine)
}

// Order puts units in cap order: triage by baseline score descending (ties by
// issue id), then each later family in stable unit-key order.
func Order(units []Unit) []Unit {
	out := append([]Unit(nil), units...)
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Family != b.Family {
			return a.Family < b.Family
		}
		if a.Family == FamilyTriage {
			if a.Baseline != b.Baseline {
				return a.Baseline > b.Baseline
			}
			return a.IssueID < b.IssueID
		}
		return a.Key < b.Key
	})
	return out
}

// UnitStatus is the outcome of one unit.
type UnitStatus string

// Unit outcomes. Only `annotated` yields annotations.
const (
	UnitAnnotated  UnitStatus = "annotated"
	UnitFailed     UnitStatus = "failed"
	UnitIncomplete UnitStatus = "incomplete" // resolved model differed from the pin
	UnitCapped     UnitStatus = "capped"
)

// UnitResult is what happened to one unit.
type UnitResult struct {
	Unit   Unit
	Status UnitStatus
	Reason string
	// Annotations holds one row per question, in question order, for an
	// annotated unit. The call's billed usage is on the first row only, so
	// summing rows never double-counts it.
	Annotations []domain.Annotation
	// CostUSD is the billed cost of the unit's call, when the reply reported it.
	CostUSD float64
}

// Outcome is the whole run.
type Outcome struct {
	Results    []UnitResult
	Total      int
	Attempted  int
	Annotated  int
	Failed     int
	Incomplete int
	Capped     int
	// CostUSD is the billed cost summed over the run's replies; CostCapped
	// is set when the cost cap (or an unknowable cost), not the call cap,
	// stopped the run.
	CostUSD    float64
	CostCapped bool
	// ThresholdsStale is set by the caller when a stored threshold's version
	// key no longer matches the question's current one.
	ThresholdsStale bool
}

// Sink persists one unit's annotations atomically: all rows or none.
type Sink interface {
	RecordAnnotations(ctx context.Context, annotations []domain.Annotation) error
}

// RunInput drives one run of the adapter.
type RunInput struct {
	Config Config
	Client *Client
	RunID  string
	Units  []Unit
	Sink   Sink
	Now    func() time.Time
}

// Run asks every unit in cap order until MaxPerRun calls were made or the
// billed cost reached MaxCostUSD. A reply that reports no billed cost stops the
// run too, since the cost cap can no longer be enforced. Failed, mismatched
// and capped units get no annotation; none becomes a score.
func Run(ctx context.Context, in RunInput) Outcome {
	cap, costCap := in.Config.MaxPerRun, in.Config.MaxCostUSD
	now := in.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	units := Order(in.Units)
	out := Outcome{Total: len(units), Results: make([]UnitResult, 0, len(units))}
	blind := false
	for index, unit := range units {
		reason := ""
		switch {
		case index >= cap:
			reason = "call cap reached"
		case blind:
			reason = "billed cost unknown"
			out.CostCapped = true
		case out.CostUSD >= costCap:
			reason = "cost cap reached"
			out.CostCapped = true
		}
		if reason != "" {
			out.Capped++
			out.Results = append(out.Results, UnitResult{Unit: unit, Status: UnitCapped, Reason: reason})
			continue
		}
		out.Attempted++
		result, reply, replied := ask(ctx, in, unit, now())
		if replied {
			if reply.CostKnown {
				result.CostUSD = reply.CostUSD
				out.CostUSD += reply.CostUSD
			} else {
				blind = true
			}
		}
		switch result.Status {
		case UnitAnnotated:
			out.Annotated++
		case UnitIncomplete:
			out.Incomplete++
		default:
			out.Failed++
		}
		out.Results = append(out.Results, result)
	}
	return out
}

// ask makes the one request of a unit. replied is true when the route
// answered, whether or not the answer was usable, because that call was billed.
func ask(ctx context.Context, in RunInput, unit Unit, at time.Time) (result UnitResult, reply Reply, replied bool) {
	reply, err := in.Client.Ask(ctx, unit.Questions, in.Config.Model, unit.Key, unit.State)
	replied = reply.Model != "" || len(reply.Usage) != 0
	if err != nil {
		var failure *Failure
		if errors.As(err, &failure) {
			if failure.Kind == FailModelMismatch {
				return UnitResult{Unit: unit, Status: UnitIncomplete, Reason: "model mismatch"}, reply, replied
			}
			return UnitResult{Unit: unit, Status: UnitFailed, Reason: failure.Reason}, reply, replied
		}
		return UnitResult{Unit: unit, Status: UnitFailed, Reason: err.Error()}, reply, replied
	}
	stateDigest := sha256.Sum256(unit.State)
	annotations := make([]domain.Annotation, 0, len(unit.Questions))
	for index, q := range unit.Questions {
		response := reply.Answers[index]
		sum := sha256.Sum256([]byte(in.RunID + "\x00" + unit.Key + "\x00" + q.ID))
		annotation := domain.Annotation{
			AnnotationID: "annotation:" + hex.EncodeToString(sum[:12]), RunID: in.RunID, IssueID: unit.IssueID, UnitKey: unit.Key,
			QuestionID: q.ID, QuestionVersion: q.Version(reply.Model), Kind: string(response.Kind),
			ModelRequested: in.Config.Model, ModelResolved: reply.Model, Probability: response.Probability,
			Answer: response.Answer, StateDigest: hex.EncodeToString(stateDigest[:]), RequestBody: reply.RequestBody,
			OptionOrder: response.OptionOrder, CreatedAt: at,
		}
		if index == 0 {
			annotation.Usage = reply.Usage
			annotation.LatencyMS = reply.Latency.Milliseconds()
		}
		annotations = append(annotations, annotation)
	}
	if in.Sink != nil {
		if err := in.Sink.RecordAnnotations(ctx, annotations); err != nil {
			return UnitResult{Unit: unit, Status: UnitFailed, Reason: "annotations not persisted: " + err.Error()}, reply, replied
		}
	}
	return UnitResult{Unit: unit, Status: UnitAnnotated, Annotations: annotations}, reply, replied
}

// Ranking disclosures of an annotated finding (design §6). A stale threshold
// adds its note; nothing else changes the text.
const (
	JevDisclosure      = "jev (unvalidated, X1 pending)"
	JevStaleDisclosure = "jev (unvalidated, X1 pending; thresholds stale)"
)

// Disclosure is the per-finding ranking text of design §6.
func (o Outcome) Disclosure(result UnitResult) (mode domain.RankingMode, text string) {
	switch result.Status {
	case UnitAnnotated:
		if o.ThresholdsStale {
			return domain.RankingModeJev, JevStaleDisclosure
		}
		return domain.RankingModeJev, JevDisclosure
	case UnitCapped:
		return domain.RankingModePartial, fmt.Sprintf("partial (%d/%d)", o.Attempted, o.Total)
	}
	return domain.RankingModeDeterministic, "deterministic (" + result.Reason + ")"
}

// Health is the decisions lane's health for a run that was eligible.
func (o Outcome) Health() domain.LaneHealth {
	health := laneHealth()
	health.Units = o.Total
	switch {
	case o.Total == 0:
		health.Status, health.Reason = domain.LaneSkipped, "no units"
	case o.Annotated == 0 && o.Attempted > 0:
		health.Status, health.Reason = domain.LaneFailed, "every call failed: "+o.firstReason()
	default:
		health.Status = domain.LaneRan
		var parts []string
		if o.Capped > 0 {
			parts = append(parts, fmt.Sprintf("partial (%d/%d)", o.Attempted, o.Total))
		}
		if o.CostCapped {
			parts = append(parts, fmt.Sprintf("cost cap reached ($%.4f billed)", o.CostUSD))
		}
		if o.Failed > 0 {
			parts = append(parts, fmt.Sprintf("%d failed", o.Failed))
		}
		if o.Incomplete > 0 {
			parts = append(parts, fmt.Sprintf("%d incomplete (model mismatch)", o.Incomplete))
		}
		health.Reason = strings.Join(parts, "; ")
	}
	return health
}

func (o Outcome) firstReason() string {
	for _, result := range o.Results {
		if result.Status == UnitFailed || result.Status == UnitIncomplete {
			return result.Reason
		}
	}
	return "unknown"
}

// SkippedHealth is the lane health when consent, key, egress or fork status
// forbids the run.
func SkippedHealth(reason string) domain.LaneHealth {
	health := laneHealth()
	health.Status, health.Reason = domain.LaneSkipped, reason
	return health
}

func laneHealth() domain.LaneHealth {
	return domain.LaneHealth{Lane: LaneID, Mode: domain.ModeSignal, Requires: domain.RequiresDecisionEgress}
}

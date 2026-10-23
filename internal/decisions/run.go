package decisions

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"scout/internal/domain"
)

// LaneID is the decisions lane's id in lane health. The lane is optional
// (design §7.3): its health never changes the exit code.
const LaneID = "decisions"

// Config is the decisions consent and pin, as read from `.scout/policy.yaml`.
type Config struct {
	Enabled bool
	// Model is the pinned Jev model id; aliases are rejected.
	Model              string
	MaxPerRun          int
	EgressAcknowledged bool
	Endpoint           string
}

// DefaultMaxPerRun is the call cap of design §6 [ASSUMPTION].
const DefaultMaxPerRun = 2000

// Validate rejects a missing pin or an alias.
func (c Config) Validate() error {
	if c.Model == "" {
		return errors.New("decisions model is required")
	}
	slug := c.Model[strings.LastIndex(c.Model, "/")+1:]
	if slug == "jev-latest" || slug == "jev-preview" || strings.HasSuffix(slug, "-latest") || strings.HasSuffix(slug, "-preview") {
		return fmt.Errorf("decisions model %q is an alias; pin an exact model id", c.Model)
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

// Unit is one question about one thing.
type Unit struct {
	Family   Family
	Key      string // stable unit key; for triage the issue id
	IssueID  string
	Question Question
	State    json.RawMessage
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

// Unit outcomes. Only `annotated` yields an annotation.
const (
	UnitAnnotated  UnitStatus = "annotated"
	UnitFailed     UnitStatus = "failed"
	UnitIncomplete UnitStatus = "incomplete" // resolved model differed from the pin
	UnitCapped     UnitStatus = "capped"
)

// UnitResult is what happened to one unit.
type UnitResult struct {
	Unit       Unit
	Status     UnitStatus
	Reason     string
	Annotation *domain.Annotation
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
}

// Sink persists annotations as they are produced.
type Sink interface {
	RecordAnnotation(ctx context.Context, annotation domain.Annotation) error
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

// Run asks every unit in cap order until MaxPerRun calls were made. Failed,
// mismatched and capped units get no annotation; none becomes a score.
func Run(ctx context.Context, in RunInput) Outcome {
	cap := in.Config.MaxPerRun
	if cap <= 0 {
		cap = DefaultMaxPerRun
	}
	now := in.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	units := Order(in.Units)
	out := Outcome{Total: len(units), Results: make([]UnitResult, 0, len(units))}
	for index, unit := range units {
		if index >= cap {
			out.Capped++
			out.Results = append(out.Results, UnitResult{Unit: unit, Status: UnitCapped, Reason: "call cap reached"})
			continue
		}
		out.Attempted++
		result := ask(ctx, in, unit, now())
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

func ask(ctx context.Context, in RunInput, unit Unit, at time.Time) UnitResult {
	response, err := in.Client.Ask(ctx, unit.Question, in.Config.Model, unit.Key, unit.State)
	if err != nil {
		var failure *Failure
		if errors.As(err, &failure) {
			if failure.Kind == FailModelMismatch {
				return UnitResult{Unit: unit, Status: UnitIncomplete, Reason: "model mismatch"}
			}
			return UnitResult{Unit: unit, Status: UnitFailed, Reason: failure.Reason}
		}
		return UnitResult{Unit: unit, Status: UnitFailed, Reason: err.Error()}
	}
	sum := sha256.Sum256([]byte(in.RunID + "\x00" + unit.Key + "\x00" + unit.Question.ID))
	annotation := domain.Annotation{
		AnnotationID: "annotation:" + hex.EncodeToString(sum[:12]), RunID: in.RunID, IssueID: unit.IssueID, UnitKey: unit.Key,
		QuestionID: unit.Question.ID, QuestionVersion: unit.Question.Version(response.Model), Kind: string(response.Kind),
		ModelRequested: in.Config.Model, ModelResolved: response.Model, Probability: response.Probability,
		Answer: response.Answer, RequestBody: response.RequestBody, OptionOrder: response.OptionOrder, Usage: response.Usage, CreatedAt: at,
	}
	if in.Sink != nil {
		if err := in.Sink.RecordAnnotation(ctx, annotation); err != nil {
			return UnitResult{Unit: unit, Status: UnitFailed, Reason: "annotation not persisted: " + err.Error()}
		}
	}
	return UnitResult{Unit: unit, Status: UnitAnnotated, Annotation: &annotation}
}

// Disclosure is the per-finding ranking text of design §6.
func (o Outcome) Disclosure(result UnitResult) (mode, text string) {
	switch result.Status {
	case UnitAnnotated:
		return ModeJev, "jev (unvalidated, X1 pending)"
	case UnitCapped:
		return ModePartial, fmt.Sprintf("partial (%d/%d)", o.Attempted, o.Total)
	case UnitIncomplete:
		return ModeDeterministic, "deterministic (" + result.Reason + ")"
	}
	return ModeDeterministic, "deterministic (" + result.Reason + ")"
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

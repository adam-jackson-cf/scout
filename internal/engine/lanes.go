package engine

import (
	"context"
	"sort"

	"scout/internal/domain"
	"scout/internal/modelcli"
)

// Lane is one registered analysis lane (REQ-16).
type Lane struct {
	ID       string
	Family   domain.Family // coverage family the lane owns; empty for inline lanes
	Triggers []domain.LaneTrigger
	Mode     domain.LaneMode
	Requires domain.LaneRequires
}

var allTriggers = []domain.LaneTrigger{domain.TriggerPR, domain.TriggerPush, domain.TriggerNightly}

// Lane IDs.
const (
	LaneSecurity            = "security"
	LaneGuardrails          = "guardrails"
	LaneQuality             = "quality"
	LaneArchitecture        = "architecture"
	LanePerformance         = "performance"
	LaneContextual          = "contextual"
	LaneClawpatchReview     = "clawpatch-review"
	LaneMeasures            = "measures"
	LaneHistory             = "history"
	LaneDelta               = "delta"
	LaneAssociation         = "association"
	LaneAssociationFallback = "association-fallback"
)

// lanes is the registry in priority order. Contextual and clawpatch-review are
// optional: they require a logged-in model CLI and are report-only, so their
// health (including skipped (no model CLI)) never makes a required lane red.
var lanes = []Lane{
	{ID: LaneGuardrails, Triggers: []domain.LaneTrigger{domain.TriggerPR}, Mode: domain.ModeBlockingEligible, Requires: domain.RequiresNone},
	{ID: LaneDelta, Triggers: []domain.LaneTrigger{domain.TriggerPR}, Mode: domain.ModeReport, Requires: domain.RequiresTool},
	{ID: LaneSecurity, Family: domain.FamilySecurity, Triggers: []domain.LaneTrigger{domain.TriggerPR, domain.TriggerPush}, Mode: domain.ModeBlockingEligible, Requires: domain.RequiresTool},
	{ID: LaneQuality, Family: domain.FamilyQuality, Triggers: allTriggers, Mode: domain.ModeReport, Requires: domain.RequiresTool},
	{ID: LaneArchitecture, Family: domain.FamilyArchitecture, Triggers: allTriggers, Mode: domain.ModeReport, Requires: domain.RequiresTool},
	{ID: LanePerformance, Family: domain.FamilyPerformance, Triggers: []domain.LaneTrigger{domain.TriggerPush, domain.TriggerNightly}, Mode: domain.ModeReport, Requires: domain.RequiresTool},
	{ID: LaneContextual, Triggers: []domain.LaneTrigger{domain.TriggerNightly}, Mode: domain.ModeReport, Requires: domain.RequiresModelCLI},
	{ID: LaneClawpatchReview, Triggers: []domain.LaneTrigger{domain.TriggerNightly}, Mode: domain.ModeReport, Requires: domain.RequiresModelCLI},
	{ID: LaneMeasures, Triggers: allTriggers, Mode: domain.ModeSignal, Requires: domain.RequiresTool},
	{ID: LaneHistory, Triggers: []domain.LaneTrigger{domain.TriggerPush, domain.TriggerNightly}, Mode: domain.ModeSignal, Requires: domain.RequiresNone},
	{ID: LaneAssociation, Triggers: []domain.LaneTrigger{domain.TriggerPush}, Mode: domain.ModeSignal, Requires: domain.RequiresDecisionEgress},
	{ID: LaneAssociationFallback, Triggers: []domain.LaneTrigger{domain.TriggerPush}, Mode: domain.ModeSignal, Requires: domain.RequiresModelCLI},
}

// DefaultFamilies lists the scanner families assessed when none are selected.
func DefaultFamilies() []domain.Family {
	var families []domain.Family
	for _, lane := range lanes {
		if lane.Family != "" {
			families = append(families, lane.Family)
		}
	}
	sort.Slice(families, func(i, j int) bool { return familyOrder(families[i]) < familyOrder(families[j]) })
	return families
}

func familyOrder(f domain.Family) int {
	for i, want := range []domain.Family{domain.FamilyArchitecture, domain.FamilySecurity, domain.FamilyQuality, domain.FamilyPerformance} {
		if f == want {
			return i
		}
	}
	return 99
}

func (l Lane) runsOn(trigger domain.LaneTrigger) bool {
	for _, t := range l.Triggers {
		if t == trigger {
			return true
		}
	}
	return false
}

// Required reports whether a lane can make a run incomplete under trigger.
func (l Lane) Required(trigger domain.LaneTrigger) bool {
	return l.Mode == domain.ModeBlockingEligible && l.runsOn(trigger)
}

// InlineLane is an in-process lane whose observations are persisted with the run.
type InlineLane struct {
	Lane  string
	Units int
	Run   func(ctx context.Context, runID string) (InlineResult, error)
}

// InlineResult is what an inline lane produced. Unmeasured, when set, names
// the rules that could not run; the lane health reports it.
type InlineResult struct {
	Observations []domain.Observation
	Unmeasured   string
}

func inlineLaneHealth(lane InlineLane, result InlineResult, err error) domain.LaneHealth {
	health := domain.LaneHealth{Lane: lane.Lane, Units: lane.Units, Status: domain.LaneRan}
	if err != nil {
		health.Status, health.Reason = domain.LaneFailed, "lane failed: "+err.Error()
	} else if lane.Units == 0 {
		health.Status, health.Reason = domain.LaneSkipped, "no changed files"
	} else if result.Unmeasured != "" {
		health.Reason = result.Unmeasured
		if len(result.Observations) == 0 {
			health.Status = domain.LaneSkipped
		}
	}
	return health
}

// laneOfCoverage maps a coverage row to its lane; ok is false for rows that
// belong to no lane (canonical mapping gaps).
func laneOfCoverage(key domain.CoverageKey) (string, bool) {
	switch {
	case key.Family == domain.FamilyClawpatch && key.Category == LaneAssociation:
		return LaneAssociation, true
	case key.Family == domain.FamilyClawpatch && key.Category == "review":
		return LaneClawpatchReview, true
	case key.Family == domain.FamilyClawpatch:
		return "", false
	case key.FeatureID != "":
		return LaneContextual, true
	}
	for _, lane := range lanes {
		if lane.Family == key.Family {
			return lane.ID, true
		}
	}
	return "", false
}

// LaneHealthOf derives one health row per registered lane for a finished run.
// inline carries the health of lanes that ran in-process. A source-drift
// failure makes every required lane failed: its results describe a tree that
// no longer exists.
func LaneHealthOf(trigger domain.LaneTrigger, run domain.Run, inline []domain.LaneHealth) []domain.LaneHealth {
	if trigger == "" {
		trigger = domain.TriggerPush
	}
	inputs := laneHealthInputs{tallies: laneCoverageTallies(run.Coverage), drift: sourceDrifted(run.Failures), inline: map[string]domain.LaneHealth{}}
	for _, h := range inline {
		inputs.inline[h.Lane] = h
	}
	out := make([]domain.LaneHealth, 0, len(lanes))
	for _, lane := range lanes {
		out = append(out, inputs.row(lane, trigger))
	}
	return out
}

type laneCoverageTally struct {
	units, ran, reused, blocked, incomplete, notApplicable, noCLI int
	reason                                                        string
}

func laneCoverageTallies(coverage []domain.Coverage) map[string]*laneCoverageTally {
	tallies := map[string]*laneCoverageTally{}
	for _, row := range coverage {
		id, ok := laneOfCoverage(row.Key)
		if !ok {
			continue
		}
		t := tallies[id]
		if t == nil {
			t = &laneCoverageTally{}
			tallies[id] = t
		}
		t.count(row)
	}
	return tallies
}

func (t *laneCoverageTally) count(row domain.Coverage) {
	t.units++
	switch row.Status {
	case domain.CoverageReused:
		t.reused++
	case domain.CoverageNotApplicable:
		t.notApplicable++
	case domain.CoverageBlocked:
		t.blocked++
	case domain.CoverageSkipped:
		if row.Reason == modelcli.NoCLIReason || row.Reason == ReasonNotOptedIn {
			t.noCLI++
		} else {
			t.incomplete++
		}
	case domain.CoverageIncomplete, domain.CoverageRequiresManualVerification:
		t.incomplete++
	default:
		t.ran++
	}
	t.noteReason(row)
}

func (t *laneCoverageTally) noteReason(row domain.Coverage) {
	if t.reason == "" && row.Reason != "" && !row.Status.Accounted() {
		t.reason = row.Reason
	}
}

func (t *laneCoverageTally) status() domain.LaneStatus {
	switch {
	case t.blocked > 0:
		return domain.LaneBlocked
	case t.incomplete > 0:
		return domain.LaneFailed
	case t.ran > 0:
		return domain.LaneRan
	case t.reused > 0:
		return domain.LaneReused
	default:
		return domain.LaneSkipped
	}
}

func sourceDrifted(failures []domain.Failure) bool {
	drift := false
	for _, failure := range failures {
		if failure.Code == failureSourceDrift {
			drift = true
		}
	}
	return drift
}

type laneHealthInputs struct {
	inline  map[string]domain.LaneHealth
	tallies map[string]*laneCoverageTally
	drift   bool
}

func (in laneHealthInputs) row(lane Lane, trigger domain.LaneTrigger) domain.LaneHealth {
	health := domain.LaneHealth{Lane: lane.ID, Mode: lane.Mode, Requires: lane.Requires, Required: lane.Required(trigger)}
	if h, ok := in.inline[lane.ID]; ok {
		health.Status, health.Units, health.Reason = h.Status, h.Units, h.Reason
	} else if t := in.tallies[lane.ID]; t == nil {
		health.Status, health.Reason = domain.LaneSkipped, laneNotRunReason(lane)
	} else {
		health.Units = t.units
		health.Status = t.status()
		health.Reason = t.reason
	}
	if in.drift && health.Required && health.Status != domain.LaneSkipped {
		health.Status, health.Reason = domain.LaneFailed, "source changed during assessment"
	}
	return health
}

func laneNotRunReason(lane Lane) string {
	if lane.Family == "" && lane.ID != LaneContextual && lane.ID != LaneClawpatchReview {
		return "lane did not run"
	}
	return "no eligible inputs"
}

// Incomplete reports whether a required lane could not run on head. A required
// lane that is skipped for lack of inputs is not incomplete; optional lanes
// never count.
func Incomplete(health []domain.LaneHealth) (bool, []string) {
	var reasons []string
	for _, h := range health {
		if !h.Required {
			continue
		}
		if h.Status == domain.LaneBlocked || h.Status == domain.LaneFailed {
			reason := h.Lane + " " + string(h.Status)
			if h.Reason != "" {
				reason += ": " + h.Reason
			}
			reasons = append(reasons, reason)
		}
	}
	return len(reasons) > 0, reasons
}

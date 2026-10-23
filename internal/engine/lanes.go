package engine

import (
	"context"
	"sort"

	"scout/internal/domain"
	"scout/internal/modelcli"
)

// Lane is one registered analysis lane (design §5.1). The registry replaces
// the families and optional stages that used to be hard-coded in Assess.
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
	LaneSecurity        = "security"
	LaneGuardrails      = "guardrails"
	LaneQuality         = "quality"
	LaneArchitecture    = "architecture"
	LanePerformance     = "performance"
	LaneContextual      = "contextual"
	LaneClawpatchReview = "clawpatch-review"
	LaneMeasures        = "measures"
	LaneHistory         = "history"
)

// lanes is the registry in priority order. Contextual and clawpatch-review are
// optional: they require a logged-in model CLI and are report-only, so their
// health (including skipped (no model CLI)) never makes a required lane red.
var lanes = []Lane{
	{ID: LaneGuardrails, Triggers: []domain.LaneTrigger{domain.TriggerPR}, Mode: domain.ModeBlockingEligible, Requires: domain.RequiresNone},
	{ID: LaneSecurity, Family: domain.FamilySecurity, Triggers: []domain.LaneTrigger{domain.TriggerPR, domain.TriggerPush}, Mode: domain.ModeBlockingEligible, Requires: domain.RequiresTool},
	{ID: LaneQuality, Family: domain.FamilyQuality, Triggers: allTriggers, Mode: domain.ModeReport, Requires: domain.RequiresTool},
	{ID: LaneArchitecture, Family: domain.FamilyArchitecture, Triggers: allTriggers, Mode: domain.ModeReport, Requires: domain.RequiresTool},
	{ID: LanePerformance, Family: domain.FamilyPerformance, Triggers: []domain.LaneTrigger{domain.TriggerPush, domain.TriggerNightly}, Mode: domain.ModeReport, Requires: domain.RequiresTool},
	{ID: LaneContextual, Triggers: []domain.LaneTrigger{domain.TriggerNightly}, Mode: domain.ModeReport, Requires: domain.RequiresModelCLI},
	{ID: LaneClawpatchReview, Triggers: []domain.LaneTrigger{domain.TriggerNightly}, Mode: domain.ModeReport, Requires: domain.RequiresModelCLI},
	{ID: LaneMeasures, Triggers: allTriggers, Mode: domain.ModeSignal, Requires: domain.RequiresTool},
	{ID: LaneHistory, Triggers: []domain.LaneTrigger{domain.TriggerPush, domain.TriggerNightly}, Mode: domain.ModeSignal, Requires: domain.RequiresNone},
}

// Lanes returns a copy of the registry.
func Lanes() []Lane { return append([]Lane(nil), lanes...) }

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
	Run   func(ctx context.Context, runID string) ([]domain.Observation, error)
}

func inlineLaneHealth(lane InlineLane, found int, err error) domain.LaneHealth {
	health := domain.LaneHealth{Lane: lane.Lane, Units: lane.Units, Status: domain.LaneRan}
	if err != nil {
		health.Status, health.Reason = domain.LaneFailed, "lane failed: "+err.Error()
	} else if lane.Units == 0 {
		health.Status, health.Reason = domain.LaneSkipped, "no changed files"
	}
	_ = found
	return health
}

// laneOfCoverage maps a coverage row to its lane; ok is false for rows that
// belong to no lane (canonical mapping gaps).
func laneOfCoverage(key domain.CoverageKey) (string, bool) {
	switch {
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
	type tally struct {
		units, ran, reused, blocked, incomplete, notApplicable, noCLI int
		reason                                                        string
	}
	tallies := map[string]*tally{}
	for _, row := range run.Coverage {
		id, ok := laneOfCoverage(row.Key)
		if !ok {
			continue
		}
		t := tallies[id]
		if t == nil {
			t = &tally{}
			tallies[id] = t
		}
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
		if t.reason == "" && row.Reason != "" && !row.Status.Accounted() {
			t.reason = row.Reason
		}
	}
	drift := false
	for _, failure := range run.Failures {
		if failure.Code == "source_drift" {
			drift = true
		}
	}
	inlineBy := map[string]domain.LaneHealth{}
	for _, h := range inline {
		inlineBy[h.Lane] = h
	}
	out := make([]domain.LaneHealth, 0, len(lanes))
	for _, lane := range lanes {
		health := domain.LaneHealth{Lane: lane.ID, Mode: lane.Mode, Requires: lane.Requires, Required: lane.Required(trigger)}
		if h, ok := inlineBy[lane.ID]; ok {
			health.Status, health.Units, health.Reason = h.Status, h.Units, h.Reason
		} else if t := tallies[lane.ID]; t == nil {
			if lane.Family == "" && lane.ID != LaneContextual && lane.ID != LaneClawpatchReview {
				health.Status, health.Reason = domain.LaneSkipped, "lane did not run"
			} else {
				health.Status, health.Reason = domain.LaneSkipped, "no eligible inputs"
			}
		} else {
			health.Units = t.units
			switch {
			case t.blocked > 0:
				health.Status = domain.LaneBlocked
			case t.incomplete > 0:
				health.Status = domain.LaneFailed
			case t.ran > 0:
				health.Status = domain.LaneRan
			case t.reused > 0:
				health.Status = domain.LaneReused
			default:
				health.Status = domain.LaneSkipped
			}
			health.Reason = t.reason
		}
		if drift && health.Required && health.Status != domain.LaneSkipped {
			health.Status, health.Reason = domain.LaneFailed, "source changed during assessment"
		}
		out = append(out, health)
	}
	return out
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

package engine

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"scout/internal/domain"
	"scout/internal/state"
)

// GuardianAssessment runs the guardian's existing merge-result-versus-B path.
// fact supplies authenticated provider coordinates/pins and the selected check;
// it is never decoded from assessment output. Provider finish time is supplied
// after the assessment by the caller's final-attempt control-plane record.
func GuardianAssessment(ctx context.Context, cfg DiffConfig, fact domain.PRRunFact) (DiffReport, error) {
	if fact.Scope != domain.ShadowScopeCI || fact.AssessedRevision == "" || fact.BaseCommit == "" || fact.HeadCommit == "" || fact.HeadTree == "" || fact.MergeBase == "" || fact.FinishedAt.IsZero() {
		return DiffReport{}, fmt.Errorf("authenticated final guardian coordinates required")
	}
	// Explicit H commit/tree identity is distinct from the assessed merge result.
	tree, err := git(ctx, cfg.Config.Root, "rev-parse", "--verify", fact.HeadCommit+"^{tree}")
	if err != nil || strings.TrimSpace(string(tree)) != fact.HeadTree {
		return DiffReport{}, fmt.Errorf("guardian head tree binding unavailable")
	}
	cfg.BaseRef = fact.BaseCommit
	cfg.HeadRef = fact.AssessedRevision
	cfg.OverrideActor = ""
	cfg.OverrideText = ""
	report, assessmentErr := Diff(ctx, cfg)
	fact.RecordedAt = nil
	fact.Status = "assessed"
	fact.Comparison = "comparable"
	fact.Inspection = "not_inspected"
	if assessmentErr != nil || report.Comparison == domain.ComparisonNone {
		fact.Comparison = "none"
	}
	// Lane coverage is independent of finding count. A successful zero-unit lane
	// is not applicable; failed/limited/blocked lanes never certify inspection.
	lane := selectedLane(fact.CheckID)
	for _, health := range report.Health {
		if health.Lane != lane {
			continue
		}
		if health.Status == domain.LaneRan || health.Status == domain.LaneReused {
			if health.Units == 0 {
				fact.Inspection = "not_applicable"
			} else if fact.Comparison == "comparable" {
				fact.Inspection = "inspected"
			}
		}
	}
	if assessmentErr != nil {
		fact.Inspection = "not_inspected"
	}
	store, err := state.Open(cfg.Config.Root)
	if err != nil {
		return report, errors.Join(assessmentErr, err)
	}
	defer store.Close()
	ledger, err := store.ReadShadowLedger(ctx)
	if err != nil {
		return report, errors.Join(assessmentErr, err)
	}
	var events []domain.ProviderPREvent
	var links []domain.PRRunEventLink
	for _, finding := range report.WouldBlock {
		if finding.CheckID != fact.CheckID || finding.RuleVersion != fact.RuleVersion {
			continue
		}
		id := domain.ShadowEventID(fact.MergeBase, fact.HeadTree, fact.CheckID, finding.IssueID, fact.RuleVersion)
		events = append(events, domain.ProviderPREvent{EventID: id, RepositoryID: fact.ID.RepositoryID, Scope: domain.ShadowScopeCI, PRNumber: fact.PRNumber, HeadCommit: fact.HeadCommit, HeadTree: fact.HeadTree, MergeBase: fact.MergeBase, CheckID: fact.CheckID, IssueID: finding.IssueID, RuleVersion: fact.RuleVersion, EventAt: fact.StartedAt})
		for _, existing := range ledger.Events {
			if existing.EventID == id && existing.RepositoryID == fact.ID.RepositoryID {
				existing.RecordedAt = nil
				events[len(events)-1] = existing
				break
			}
		}
		links = append(links, domain.PRRunEventLink{AttemptID: fact.ID, Scope: domain.ShadowScopeCI, EventID: id})
	}
	return report, errors.Join(assessmentErr, store.RecordPRRun(ctx, fact, events, links))
}

func selectedLane(check string) string {
	switch {
	case strings.HasPrefix(check, "guardrails."):
		return LaneGuardrails
	case strings.HasPrefix(check, "security.secrets."):
		return LaneSecurity
	case check == "security.ci-mutable-action-ref":
		return LaneSecurity
	default:
		return check
	}
}

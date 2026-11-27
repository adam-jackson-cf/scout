//go:build boundary

package engine

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"scout/internal/domain"
	"scout/internal/state"
)

// GuardianAssessment runs the guardian's existing merge-result-versus-B path.
// fact supplies authenticated provider coordinates/pins and the selected check;
// it is never decoded from assessment output. Provider finish time is supplied
// after the assessment by the caller's final-attempt control-plane record.
func GuardianAssessment(ctx context.Context, cfg DiffConfig, fact domain.PRRunFact) (DiffReport, error) {
	if !guardianCoordinatesComplete(fact) {
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
	fact = guardianOutcome(fact, report, assessmentErr)
	return report, errors.Join(assessmentErr, recordGuardianRun(ctx, cfg.Config.Root, fact, report))
}

func guardianCoordinatesComplete(fact domain.PRRunFact) bool {
	return fact.Scope == domain.ShadowScopeCI && fact.AssessedRevision != "" && fact.BaseCommit != "" && fact.HeadCommit != "" && fact.HeadTree != "" && fact.MergeBase != "" && !fact.FinishedAt.IsZero()
}

func guardianOutcome(fact domain.PRRunFact, report DiffReport, assessmentErr error) domain.PRRunFact {
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
		if health.Lane == lane {
			fact.Inspection = laneInspection(health, fact.Comparison, fact.Inspection)
		}
	}
	if assessmentErr != nil {
		fact.Inspection = "not_inspected"
	}
	return fact
}

func laneInspection(health domain.LaneHealth, comparison, current string) string {
	if health.Status != domain.LaneRan && health.Status != domain.LaneReused {
		return current
	}
	if health.Units == 0 {
		return "not_applicable"
	}
	if comparison == "comparable" {
		return "inspected"
	}
	return current
}

func recordGuardianRun(ctx context.Context, root string, fact domain.PRRunFact, report DiffReport) (err error) {
	store, err := state.Open(root)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, store.Close()) }()
	ledger, err := store.ReadShadowLedger(ctx)
	if err != nil {
		return err
	}
	events, links := guardianEvents(fact, report, ledger.Events)
	return store.RecordPRRun(ctx, fact, events, links)
}

func guardianEvents(fact domain.PRRunFact, report DiffReport, recorded []domain.ProviderPREvent) ([]domain.ProviderPREvent, []domain.PRRunEventLink) {
	var events []domain.ProviderPREvent
	var links []domain.PRRunEventLink
	for _, finding := range report.WouldBlock {
		if finding.CheckID != fact.CheckID || finding.RuleVersion != fact.RuleVersion {
			continue
		}
		id := domain.ShadowEventID(fact.MergeBase, fact.HeadTree, fact.CheckID, finding.IssueID, fact.RuleVersion)
		event := domain.ProviderPREvent{EventID: id, RepositoryID: fact.ID.RepositoryID, Scope: domain.ShadowScopeCI, PRNumber: fact.PRNumber, HeadCommit: fact.HeadCommit, HeadTree: fact.HeadTree, MergeBase: fact.MergeBase, CheckID: fact.CheckID, IssueID: finding.IssueID, RuleVersion: fact.RuleVersion, EventAt: fact.StartedAt}
		events = append(events, recordedEvent(event, recorded))
		links = append(links, domain.PRRunEventLink{AttemptID: fact.ID, Scope: domain.ShadowScopeCI, EventID: id})
	}
	return events, links
}

func recordedEvent(event domain.ProviderPREvent, recorded []domain.ProviderPREvent) domain.ProviderPREvent {
	for _, existing := range recorded {
		if existing.EventID == event.EventID && existing.RepositoryID == event.RepositoryID {
			existing.RecordedAt = nil
			return existing
		}
	}
	return event
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

func TestGuardianPersistsZeroEventAndFailedAttempts(t *testing.T) {
	for _, failed := range []bool{false, true} {
		name := "zero_event"
		if failed {
			name = "authenticated_input_failure"
		}
		t.Run(name, func(t *testing.T) {
			root, base := deltaRepo(t, map[string]string{"a.py": "x = 1\n"})
			binary := fixtureClawpatch(t, root)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			f := guardianFixtureFact(ctx, t, root, base)
			if failed {
				f.BaseCommit = "absent-base"
			}
			report, err := GuardianAssessment(ctx, DiffConfig{Config: Config{Root: root, ClawpatchBinary: binary}, OverrideActor: "alice", OverrideText: "scout-override"}, f)
			if (err != nil) != failed {
				t.Fatalf("report=%+v error=%v", report, err)
			}
			assertGuardianLedger(ctx, t, root, f, failed)
		})
	}
}

func guardianFixtureFact(ctx context.Context, t *testing.T, root, base string) domain.PRRunFact {
	t.Helper()
	raw, err := git(ctx, root, "rev-parse", base+"^{tree}")
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	return domain.PRRunFact{ID: domain.AttemptID{RepositoryID: "repo", RunID: "1000", JobID: "2000", Attempt: "1"}, Scope: domain.ShadowScopeCI, PRNumber: "17", HeadCommit: base, HeadTree: strings.TrimSpace(string(raw)), BaseCommit: base, MergeBase: base, AssessedRevision: base, CheckID: "guardrails.config", RuleVersion: "v1", CreatedAt: at, StartedAt: at, FinishedAt: at.Add(time.Minute), Pins: domain.ShadowPins{Scout: "scout", Tools: "tools", Catalogue: "catalogue", Policy: "policy", Recipe: "guardian"}}
}

func assertGuardianLedger(ctx context.Context, t *testing.T, root string, f domain.PRRunFact, failed bool) {
	t.Helper()
	store, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	}()
	ledger, err := store.ReadShadowLedger(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(ledger.Facts) != 1 || len(ledger.Events) != 0 || ledger.Facts[0].RecordedAt != nil || ledger.Facts[0].Pins != f.Pins {
		t.Fatalf("attempt not durably staged: %+v", ledger)
	}
	if failed {
		assertNoFabricatedCoverage(t, ledger.Facts[0])
	}
}

func assertNoFabricatedCoverage(t *testing.T, fact domain.PRRunFact) {
	t.Helper()
	if fact.Comparison != "none" || fact.Inspection != "not_inspected" {
		t.Fatalf("failed attempt fabricated coverage: %+v", fact)
	}
}

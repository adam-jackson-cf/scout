package issues

import (
	"context"
	"testing"
	"time"

	"scout/internal/domain"
	"scout/internal/state"
)

func TestCorrelateLeavesAmbiguousCandidatesSeparate(t *testing.T) {
	ctx := context.Background()
	store, err := state.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	priorPlan, priorRun, priorObservation := correlationFixture("prior", "prior-observation")
	if err := store.SavePlan(ctx, priorPlan); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveRun(ctx, priorRun, []domain.Observation{priorObservation}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"issue:security:first", "issue:security:second"} {
		issue := domain.Issue{IssueID: id, State: domain.IssueOpen, Title: "security:test: unsafe path", Mechanism: "unsafe path", OriginalSeverity: domain.SeverityHigh, EffectiveSeverity: domain.SeverityHigh, ObservationIDs: []string{priorObservation.ObservationID}, Files: []string{"main.go"}, History: []domain.IssueHistoryEntry{{At: time.Now().UTC(), To: domain.IssueOpen, Actor: "test", Reason: "created"}}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
		if err := store.CreateIssue(ctx, issue); err != nil {
			t.Fatal(err)
		}
	}
	currentPlan, currentRun, currentObservation := correlationFixture("current", "current-observation")
	if err := store.SavePlan(ctx, currentPlan); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveRun(ctx, currentRun, []domain.Observation{currentObservation}); err != nil {
		t.Fatal(err)
	}
	created, err := Correlate(ctx, store, currentRun, []domain.Observation{currentObservation})
	if err != nil {
		t.Fatal(err)
	}
	if len(created) != 1 || created[0].IssueID == "issue:security:first" || created[0].IssueID == "issue:security:second" {
		t.Fatalf("ambiguous correlation merged a durable issue: %#v", created)
	}
	all, err := store.ListIssues(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("issues = %#v", all)
	}
}

func correlationFixture(suffix, observationID string) (domain.Plan, domain.Run, domain.Observation) {
	key := domain.CoverageKey{FilePath: "main.go", Family: domain.FamilySecurity, Category: "security:test"}
	snapshot := domain.Snapshot{RepositoryID: "repository", WorktreeDigest: "worktree-" + suffix, IgnoreRulesDigest: "ignore", ExclusionsDigest: "exclude"}
	plan := domain.Plan{PlanID: "plan-" + suffix, RunID: "run-" + suffix, Snapshot: snapshot, Scope: domain.Scope{Kind: domain.ScopeRepository, SelectedFamilies: []domain.Family{domain.FamilySecurity}}, CreatedAt: time.Now().UTC()}
	run := domain.Run{RunID: plan.RunID, PlanID: plan.PlanID, Snapshot: snapshot, Scope: plan.Scope, State: domain.RunPartial, StartedAt: time.Now().UTC(), Coverage: []domain.Coverage{{Key: key, Status: domain.CoverageFlagged, InspectedRanges: []domain.SourceRange{{Path: "main.go", StartLine: 1, EndLine: 1}}, ObservationIDs: []string{observationID}}}}
	observation := domain.Observation{ObservationID: observationID, RunID: run.RunID, Source: "scanner", SourceFindingID: "stable-finding", Family: domain.FamilySecurity, OriginalCategory: "security:test", RuleID: "rule", RuleVersion: "1", SourceSeverity: domain.SeverityHigh, EffectiveSeverity: domain.SeverityHigh, Confidence: domain.ConfidenceHigh, Mechanism: "unsafe path", Evidence: []domain.Evidence{{Location: domain.SourceRange{Path: "main.go", StartLine: 1, EndLine: 1}, RedactedQuote: "support"}}, Validated: true, ObservedAt: time.Now().UTC()}
	return plan, run, observation
}

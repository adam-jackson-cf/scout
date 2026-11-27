package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"scout/internal/association"
	"scout/internal/domain"
	"scout/internal/engine"
	"scout/internal/policy"
)

func guardCLIHistory() (*time.Time, []domain.Association) {
	generation := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	rows := make([]domain.Association, 25)
	for i := range rows {
		rows[i] = domain.Association{ID: fmt.Sprintf("history-%d", i), RunID: fmt.Sprintf("run-%d", i), MergeCommit: fmt.Sprintf("merge-%d", i), ClusterKey: fmt.Sprintf("cluster-%d", i), Files: []string{fmt.Sprintf("history/%d.go", i)}, DecidedAt: generation.Add(time.Duration(i+1) * time.Minute), LastAgentRefreshAt: &generation, Outcome: domain.AssociationAssigned, AreaIDs: []string{"area-a"}, JevOutcome: "none", CLIName: "codex", PRMetadata: domain.PRMetadataAbsent}
	}
	return &generation, rows
}

func TestAssociationGuardSummaryOutput(t *testing.T) {
	generation, rows := guardCLIHistory()
	guard := association.Guard(rows, generation, policy.DefaultMapping().Association.Guard, association.GuardState{})
	text := engine.Summary(engine.SummaryInput{Guard: guard})
	if !strings.Contains(text, "catalogue-stale") || !strings.Contains(text, "scout map --refresh") {
		t.Fatalf("summary lacks manual refresh diagnostic: %s", text)
	}
}

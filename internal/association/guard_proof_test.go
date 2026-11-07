package association

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"scout/internal/domain"
	"scout/internal/policy"
	"scout/internal/state"
)

func guardProofConfig() policy.AssociationGuard {
	return policy.AssociationGuard{WindowMerges: 20, MinClusters: 20, FallbackRateMax: .40, PendingNewRateMax: .20}
}
func guardProofGeneration() *time.Time { v := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC); return &v }
func guardProofRows(n int, mode string) []domain.Association {
	rows := make([]domain.Association, n)
	for i := range rows {
		r := domain.Association{ID: fmt.Sprintf("decision-%03d", i), RunID: fmt.Sprintf("run-%03d", i), DecidedAt: guardProofGeneration().Add(time.Duration(i+1) * time.Hour), LastAgentRefreshAt: guardProofGeneration(), MergeCommit: fmt.Sprintf("merge-%03d", n-i), ClusterKey: fmt.Sprintf("cluster-%03d", i), Files: []string{fmt.Sprintf("file-%03d.go", i)}, Outcome: domain.AssociationAssigned, AreaIDs: []string{"area-a"}, JevOutcome: "existing", PRMetadata: domain.PRMetadataAbsent}
		switch mode {
		case "healthy":
		case "new":
			r.JevOutcome = "none"
			r.CLIName = "codex"
			r.Outcome = domain.AssociationPendingNew
			r.PendingAreaID = "pending-a"
			r.AreaIDs = nil
		case "answered":
			r.JevOutcome = "unresolved"
			r.CLIName = "codex"
		case "pending_cap":
			r.JevOutcome = "none"
			r.CLIName = "codex"
			r.Outcome = domain.AssociationUnresolved
			r.Reason = domain.AssociationReasonPendingCap
			r.AreaIDs = nil
		default:
			r.Outcome = domain.AssociationUnresolved
			r.AreaIDs = nil
			r.JevOutcome = "none"
			r.Reason = mode
			if mode == "no model CLI" || mode == "no consent" {
				r.Reason = domain.AssociationReasonNoneUnsent
				r.FallbackSkipped = true
				r.FallbackSkipReason = mode
			}
		}
		rows[i] = r
	}
	return rows
}
func guardProofRate(t *testing.T, got GuardRate, num, den int, measured bool, reason string) {
	t.Helper()
	if got.Numerator != num || got.Denominator != den || got.Measured != measured || got.Reason != reason {
		t.Fatalf("rate=%+v want %d/%d measured=%v reason=%q", got, num, den, measured, reason)
	}
	if measured && got.Value != float64(num)/float64(den) {
		t.Fatalf("rate value=%v", got.Value)
	}
}
func TestGuardAvailabilityAndReasons(t *testing.T) {
	cases := []struct {
		name, mode              string
		fallbackDen, pendingDen int
		measured                bool
		reason                  string
	}{
		{"no CLI", "no model CLI", 20, 0, false, "insufficient_clusters"},
		{"no consent", "no consent", 20, 0, false, "insufficient_clusters"},
		{"all Terra failures", domain.AssociationReasonTerraFailure, 20, 0, false, "insufficient_clusters"},
		{"fallback cap", domain.AssociationReasonFallbackCap, 20, 0, false, "insufficient_clusters"},
		{"Jev failure", domain.AssociationReasonJevFailure, 0, 0, false, "insufficient_clusters"},
		{"Jev cap", domain.AssociationReasonJevCap, 0, 0, false, "insufficient_clusters"},
		{"PR cap", domain.AssociationReasonPRCap, 0, 0, false, "insufficient_clusters"},
		{"healthy without escalation", "healthy", 20, 20, false, "no_fallback_answers"},
		{"answered fallback", "answered", 20, 20, true, ""},
		{"valid new before pending cap", "pending_cap", 20, 20, true, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Guard(guardProofRows(20, tc.mode), guardProofGeneration(), guardProofConfig(), GuardState{})
			numerator := tc.fallbackDen
			if tc.mode == "healthy" {
				numerator = 0
			}
			fallbackReason := ""
			if tc.fallbackDen < 20 {
				fallbackReason = "insufficient_clusters"
			}
			guardProofRate(t, got.FallbackRate, numerator, tc.fallbackDen, tc.fallbackDen >= 20, fallbackReason)
			pending := 0
			if tc.mode == "pending_cap" {
				pending = 20
			}
			guardProofRate(t, got.PendingNewRate, pending, tc.pendingDen, tc.measured, tc.reason)
		})
	}
}
func TestGuardPartialAvailabilityAndRunEligibility(t *testing.T) {
	rows := guardProofRows(40, "answered")
	for i := range 20 {
		rows[i].FallbackSkipped = true
		rows[i].FallbackSkipReason = "no model CLI"
	}
	cfg := guardProofConfig()
	cfg.WindowMerges = 40
	got := Guard(rows, guardProofGeneration(), cfg, GuardState{})
	guardProofRate(t, got.FallbackRate, 40, 40, true, "")
	guardProofRate(t, got.PendingNewRate, 0, 20, true, "")
	// A skipped run invalidates even its separate, non-escalated merge.
	rows[20].RunID = rows[0].RunID
	rows[20].JevOutcome = "existing"
	rows[20].CLIName = ""
	got = Guard(rows, guardProofGeneration(), cfg, GuardState{})
	guardProofRate(t, got.PendingNewRate, 0, 19, false, "insufficient_clusters")
}
func TestGuardMinimumAndThresholdBoundaries(t *testing.T) {
	cfg := guardProofConfig()
	rows := guardProofRows(20, "healthy")
	for i := range 8 {
		rows[i].JevOutcome = "none"
		rows[i].CLIName = "codex"
	}
	for i := range 4 {
		rows[i].Outcome = domain.AssociationPendingNew
		rows[i].PendingAreaID = "pending"
		rows[i].AreaIDs = nil
	}
	got := Guard(rows, guardProofGeneration(), cfg, GuardState{})
	guardProofRate(t, got.FallbackRate, 8, 20, true, "")
	guardProofRate(t, got.PendingNewRate, 4, 20, true, "")
	if got.CatalogueStale {
		t.Fatal("equal limits are not exceeding (REQ-158)")
	}
	rows[8].JevOutcome = "none"
	rows[8].CLIName = "codex"
	if !Guard(rows, guardProofGeneration(), cfg, GuardState{}).CatalogueStale {
		t.Fatal("fallback strictly exceeds")
	}
	rows[8].JevOutcome = "existing"
	rows[8].CLIName = ""
	rows[4].Outcome = domain.AssociationPendingNew
	rows[4].PendingAreaID = "pending"
	rows[4].AreaIDs = nil
	if !Guard(rows, guardProofGeneration(), cfg, GuardState{}).CatalogueStale {
		t.Fatal("pending strictly exceeds")
	}
	got = Guard(rows[:19], guardProofGeneration(), cfg, GuardState{})
	if got.FallbackRate.Measured || got.PendingNewRate.Measured || got.CatalogueStale {
		t.Fatalf("19 clusters must be unmeasured: %+v", got)
	}
}
func TestGuardDecisionOrderAndInheritedExclusion(t *testing.T) {
	rows := guardProofRows(25, "healthy")
	for i := range 5 {
		rows[i].JevOutcome = "none"
		rows[i].FallbackSkipped = true
	}
	inherited := guardProofRows(1, "healthy")[0]
	inherited.ID = "inherited"
	inherited.MergeCommit = "inherited-merge"
	inherited.Outcome = domain.AssociationInherited
	inherited.InheritedRule = domain.InheritedRename
	inherited.InheritedAreaIDs = inherited.AreaIDs
	inherited.DecidedAt = rows[24].DecidedAt.Add(time.Hour)
	rows = append(rows, inherited)
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
	before := append([]domain.Association(nil), rows...)
	got := Guard(rows, guardProofGeneration(), guardProofConfig(), GuardState{})
	guardProofRate(t, got.FallbackRate, 0, 20, true, "")
	if len(got.WindowMerges) != 20 || got.WindowMerges[0] != "merge-020" || got.WindowMerges[19] != "merge-001" {
		t.Fatalf("decision ordered window=%v", got.WindowMerges)
	}
	if !reflect.DeepEqual(rows, before) {
		t.Fatal("guard mutated input")
	}
}
func TestGuardGenerationAndHistoricalLatch(t *testing.T) {
	rows := guardProofRows(45, "healthy")
	for i := 0; i < 20; i++ {
		rows[i].JevOutcome = "none"
		rows[i].FallbackSkipped = true
		rows[i].FallbackSkipReason = "no model CLI"
	}
	got := Guard(rows, guardProofGeneration(), guardProofConfig(), GuardState{})
	if !got.CatalogueStale || got.Recommendation == "" {
		t.Fatalf("historical trip lost: %+v", got)
	}
	guardProofRate(t, got.FallbackRate, 0, 20, true, "")
	for i := range rows {
		rows[i].CatalogueDigest = "changed-catalogue"
	}
	if !Guard(rows, guardProofGeneration(), guardProofConfig(), GuardState{}).CatalogueStale {
		t.Fatal("catalogue-only reset")
	}
	if !Guard(rows[25:], guardProofGeneration(), guardProofConfig(), got).CatalogueStale {
		t.Fatal("same generation prior latch lost")
	}
	next := guardProofGeneration().Add(100 * time.Hour)
	reset := Guard(rows, &next, guardProofConfig(), got)
	if reset.CatalogueStale || reset.Recommendation != "" || len(reset.WindowMerges) != 0 || reset.FallbackRate.Measured {
		t.Fatalf("old generation leaked: %+v", reset)
	}
}
func TestGuardRestoreReplaysHistoricalTrip(t *testing.T) {
	ctx := context.Background()
	source := proofStore(t)
	rows := guardProofRows(45, "healthy")
	for i := 0; i < 20; i++ {
		rows[i].JevOutcome = "none"
		rows[i].FallbackSkipped = true
		rows[i].FallbackSkipReason = "no model CLI"
	}
	if err := source.InsertAssociations(ctx, rows); err != nil {
		t.Fatal(err)
	}
	if err := source.SaveMappingRefresh(ctx, state.MappingRefresh{LastAgentRefreshAt: guardProofGeneration()}); err != nil {
		t.Fatal(err)
	}
	files, err := source.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	target := proofStore(t)
	if err := target.Import(ctx, files); err != nil {
		t.Fatal(err)
	}
	imported, err := target.ListAssociations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	refresh, err := target.LoadMappingRefresh(ctx)
	if err != nil {
		t.Fatal(err)
	}
	before := Guard(rows, guardProofGeneration(), guardProofConfig(), GuardState{})
	after := Guard(imported, refresh.LastAgentRefreshAt, guardProofConfig(), GuardState{})
	if !before.CatalogueStale || !reflect.DeepEqual(before, after) {
		t.Fatalf("REQ-51 guard changed: before=%+v after=%+v", before, after)
	}
}

func TestGuardWholeWindowAndFailedMergeEligibility(t *testing.T) {
	rows := guardProofRows(40, "healthy")
	for i := range rows {
		rows[i].MergeCommit = fmt.Sprintf("merge-%02d", i/2)
		rows[i].RunID = fmt.Sprintf("run-%02d", i/2)
	}
	for i := range 10 {
		rows[i].JevOutcome = "none"
		rows[i].CLIName = "codex"
		rows[i].Outcome = domain.AssociationPendingNew
		rows[i].PendingAreaID = "pending"
		rows[i].AreaIDs = nil
	}
	// One failed cluster disqualifies its whole merge, including the valid
	// new answer alongside it, but both clusters remain in fallback's window.
	rows[1].Outcome = domain.AssociationUnresolved
	rows[1].Reason = domain.AssociationReasonTerraFailure
	rows[1].PendingAreaID = ""
	got := Guard(rows, guardProofGeneration(), guardProofConfig(), GuardState{})
	guardProofRate(t, got.FallbackRate, 10, 40, true, "")
	guardProofRate(t, got.PendingNewRate, 8, 38, true, "")
	if len(got.WindowMerges) != 20 {
		t.Fatalf("clusters consumed merge window: %v", got.WindowMerges)
	}
}

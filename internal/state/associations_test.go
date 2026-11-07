package state

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"scout/internal/areas"
	"scout/internal/domain"
)

func testAssociations() []domain.Association {
	decided := time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)
	refreshed := time.Date(2026, 10, 30, 8, 0, 0, 0, time.UTC)
	covered, choice := 0.71, 0.6
	base := domain.Association{
		RunID: "run-1", DecidedAt: decided, LastAgentRefreshAt: &refreshed, MergeCommit: "abc123", PRMetadata: domain.PRMetadataAbsent,
		ThresholdsProvisional: true, FallbackSkipped: true, FallbackSkipReason: "no consent",
	}
	assigned := base
	assigned.ID, assigned.ClusterKey, assigned.Files = "assoc-b", "pkg/a|x", []string{"pkg/a/y.go", "pkg/a/x.go"}
	assigned.Outcome, assigned.AreaIDs = domain.AssociationAssigned, []string{"area-2", "area-1"}
	assigned.JevCovered, assigned.JevChoice, assigned.JevOutcome = &covered, &choice, "existing"
	assigned.RawAnswer, assigned.Retries, assigned.LatencyMillis = `{"a":"<b>"}`, 1, 420
	inherited := base
	inherited.ID, inherited.ClusterKey, inherited.Files = "assoc-a", "pkg/a|rename", []string{"pkg/a/z.go"}
	inherited.Outcome, inherited.InheritedRule, inherited.InheritedFrom, inherited.InheritedAreaIDs = domain.AssociationInherited, domain.InheritedRename, "pkg/a/old.go", []string{"area-1"}
	inherited.LastAgentRefreshAt = nil
	inherited.DecidedAt = decided.Add(time.Minute)
	unresolved := base
	unresolved.ID, unresolved.ClusterKey, unresolved.Files = "assoc-c", "pkg/b|x", []string{"pkg/b/x.go"}
	unresolved.Outcome, unresolved.Reason = domain.AssociationUnresolved, domain.AssociationReasonPRCap
	return []domain.Association{inherited, unresolved, assigned}
}

func TestAssociationCatalogueStaleIsOutputOnly(t *testing.T) {
	ctx := context.Background()
	store := seedLedger(t, ctx, t.TempDir())
	defer store.Close()
	rows := testAssociations()
	for i := range rows {
		rows[i].CatalogueStale = true
	}
	if err := store.InsertAssociations(ctx, rows); err != nil {
		t.Fatal(err)
	}
	stored, err := store.ListAssociations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range stored {
		if row.CatalogueStale {
			t.Fatal("output annotation persisted")
		}
	}
	if !rows[0].CatalogueStale {
		t.Fatal("normalization mutated caller's output")
	}
	files, err := store.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(files[ExportAssociations], []byte("catalogue_stale")) {
		t.Fatal("derived annotation entered state export")
	}
}

func TestAssociationsAndPendingAreasRoundTripByteIdentical(t *testing.T) {
	ctx := context.Background()
	source := seedLedger(t, ctx, t.TempDir())
	defer source.Close()
	before, err := source.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := before[ExportAssociations]; ok || !strings.Contains(string(before[ExportManifest]), ExportAssociations) {
		t.Fatal("associations.jsonl must be absent and listed under absent while the store holds none")
	}

	accepted := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	list := []areas.Area{
		{ScoutAreaID: "area-1", Title: "One", FileSet: []string{"a.go"}, Origin: areas.OriginAgent, FirstSeenSnapshot: "s", ClawpatchIDs: []string{"cp"}},
		{ScoutAreaID: "area-p", Title: "Pending <x>", FileSet: []string{"p.go"}, Origin: "association", FirstSeenSnapshot: "s", ClawpatchIDs: []string{}, Status: areas.StatusPending, ClusterID: "c1", DecisionID: "assoc-b"},
		{ScoutAreaID: "area-s", Title: "Accepted", FileSet: []string{"q.go"}, Origin: "association", FirstSeenSnapshot: "s", ClawpatchIDs: []string{}, Status: areas.StatusActive, AcceptedBy: "maintainer", AcceptedAt: &accepted},
	}
	if err := source.ReplaceAreas(ctx, list); err != nil {
		t.Fatal(err)
	}
	if err := source.InsertAssociations(ctx, testAssociations()); err != nil {
		t.Fatal(err)
	}
	if err := source.MarkAssociationsStale(ctx, []string{"assoc-c"}); err != nil {
		t.Fatal(err)
	}
	if err := source.MarkAssociationsSuperseded(ctx, []string{"assoc-b"}); err != nil {
		t.Fatal(err)
	}
	if err := source.InsertAssociations(ctx, testAssociations()[:1]); err == nil {
		t.Fatal("associations are append-only: a repeated ID must fail")
	}
	if err := source.MarkAssociationsStale(ctx, []string{"missing"}); err == nil {
		t.Fatal("marking an unknown association must fail")
	}

	got, err := source.ListAssociations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var order []string
	for _, record := range got {
		order = append(order, record.ID)
	}
	if want := []string{"assoc-b", "assoc-c", "assoc-a"}; !reflect.DeepEqual(order, want) {
		t.Fatalf("order = %v, want %v (decision time, then id)", order, want)
	}
	if !got[1].Stale || got[0].Stale || !got[0].Superseded {
		t.Fatalf("flags = %+v", got)
	}

	first, err := source.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(first[ExportAssociations], []byte(`\u003c`)) || len(SplitLines(first[ExportAssociations])) != 3 {
		t.Fatalf("bad associations.jsonl: %s", first[ExportAssociations])
	}
	target, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	if err := target.Import(ctx, first); err != nil {
		t.Fatal(err)
	}
	second, err := target.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != len(second) {
		t.Fatalf("file set changed: %d vs %d", len(first), len(second))
	}
	for path, data := range first {
		if !bytes.Equal(data, second[path]) {
			t.Fatalf("%s differs after export -> import -> export", path)
		}
	}
	imported, err := target.ListAreas(ctx)
	if err != nil || len(imported) != 3 || imported[1].Status != areas.StatusPending || imported[2].AcceptedBy != "maintainer" {
		t.Fatalf("imported areas = %+v, %v", imported, err)
	}
}

func TestPreAssociationExportImportsAsNoAssociations(t *testing.T) {
	ctx := context.Background()
	source := seedLedger(t, ctx, t.TempDir())
	defer source.Close()
	files, err := source.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	delete(files, ExportAssociations)
	delete(files, ExportManifest)
	files[ExportManifest] = RenderManifest(ManifestMeta{}, files)
	target, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	if err := target.Import(ctx, files); err != nil {
		t.Fatal(err)
	}
	if list, err := target.ListAssociations(ctx); err != nil || len(list) != 0 {
		t.Fatalf("associations = %v, %v", list, err)
	}
}

func TestAreaActiveByStatus(t *testing.T) {
	for status, want := range map[string]bool{"": true, areas.StatusActive: true, areas.StatusPending: false, areas.StatusSuperseded: false} {
		if got := (areas.Area{Status: status}).Active(); got != want {
			t.Errorf("status %q Active() = %v, want %v", status, got, want)
		}
	}
	if (areas.Area{Status: areas.StatusActive, RetiredSnapshot: "s"}).Active() {
		t.Error("a retired area is not active")
	}
}

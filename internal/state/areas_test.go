package state

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"scout/internal/areas"
	"scout/internal/domain"
)

func TestAreasExportImportExportIsByteIdenticalAndLeavesIdentityAlone(t *testing.T) {
	ctx := context.Background()
	source := seedLedger(t, ctx, t.TempDir())
	defer source.Close()

	before, err := source.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := before[ExportAreas]; ok {
		t.Fatal("areas.jsonl must be absent while the store holds no areas")
	}
	issuesBefore, err := source.ListIssues(ctx)
	if err != nil {
		t.Fatal(err)
	}
	signalsBefore, err := source.ListSignals(ctx, SignalFilter{})
	if err != nil {
		t.Fatal(err)
	}

	next, _ := areas.Reconcile(nil, []domain.Feature{
		{FeatureID: "cp-1", Title: "Auth <svc> & more", Source: areas.AgentSource, OwnedFiles: []domain.FeatureFileRef{{Path: "b.go"}, {Path: "a.go"}}},
		{FeatureID: "cp-2", Title: "Billing", Source: areas.AgentSource, OwnedFiles: []domain.FeatureFileRef{{Path: "c.go"}}},
	}, nil, areas.Config{Snapshot: "snap-1"})
	if err := source.ReplaceAreas(ctx, next); err != nil {
		t.Fatal(err)
	}

	issuesAfter, err := source.ListIssues(ctx)
	if err != nil {
		t.Fatal(err)
	}
	signalsAfter, err := source.ListSignals(ctx, SignalFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(issuesBefore, issuesAfter) || !reflect.DeepEqual(signalsBefore, signalsAfter) {
		t.Fatal("reconciling areas changed issue identity or signal series keys")
	}

	first, err := source.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(splitLines(first[ExportAreas])); got != 2 {
		t.Fatalf("areas.jsonl has %d rows, want 2", got)
	}
	if bytes.Contains(first[ExportAreas], []byte(`\u003c`)) {
		t.Fatal("areas export must not HTML-escape")
	}
	for _, path := range []string{ExportIssues, ExportDispositions, ExportWouldBlock, ExportCoverageSummary} {
		if !bytes.Equal(before[path], first[path]) {
			t.Fatalf("%s changed when only areas changed", path)
		}
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
		t.Fatalf("round trip changed the file set: %d vs %d files", len(first), len(second))
	}
	for path, data := range first {
		if !bytes.Equal(data, second[path]) {
			t.Fatalf("%s differs after export -> import -> export", path)
		}
	}
	imported, err := target.ListAreas(ctx)
	if err != nil || !reflect.DeepEqual(imported, next) {
		t.Fatalf("imported areas = %+v, %v; want %+v", imported, err, next)
	}
}

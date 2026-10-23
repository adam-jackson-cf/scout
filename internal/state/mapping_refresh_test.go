package state

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func TestMappingRefreshSurvivesExportImport(t *testing.T) {
	ctx := context.Background()
	source, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	if data, err := source.ExportMappingRefreshFile(ctx); err != nil || data != nil {
		t.Fatalf("empty store exports nothing: %q %v", data, err)
	}
	last := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	record := MappingRefresh{LastAgentRefreshAt: &last, Pending: &PendingRefresh{
		FirstTriggerAt: last.Add(time.Hour), LastTriggerAt: last.Add(2 * time.Hour),
		Reasons: []string{"manifest-added:go.mod", "unowned-added:x.go"}, Revisions: []string{"r1", "r2"},
	}}
	if err := source.SaveMappingRefresh(ctx, record); err != nil {
		t.Fatal(err)
	}
	files, err := source.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(files[ExportMappingRefresh]) == 0 {
		t.Fatal("export must carry mapping-refresh.json")
	}
	target, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	if err := target.Import(ctx, files); err != nil {
		t.Fatal(err)
	}
	got, err := target.LoadMappingRefresh(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.Pending == nil || !got.Pending.LastTriggerAt.Equal(record.Pending.LastTriggerAt) || len(got.Pending.Revisions) != 2 || got.LastAgentRefreshAt == nil || !got.LastAgentRefreshAt.Equal(last) {
		t.Fatalf("imported record = %+v", got)
	}
	again, err := target.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(again[ExportMappingRefresh], files[ExportMappingRefresh]) {
		t.Fatalf("round trip differs:\n%s\n%s", files[ExportMappingRefresh], again[ExportMappingRefresh])
	}
}

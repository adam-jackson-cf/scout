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
	record := MappingRefresh{LastAgentRefreshAt: &last}
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
	if got.LastAgentRefreshAt == nil || !got.LastAgentRefreshAt.Equal(last) {
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

// Baselines exported before revision 2.6 carry a `pending` object; they must
// still import and re-export without it.
func TestMappingRefreshImportsPre26PendingField(t *testing.T) {
	ctx := context.Background()
	last := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	source, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	if err := source.SaveMappingRefresh(ctx, MappingRefresh{LastAgentRefreshAt: &last}); err != nil {
		t.Fatal(err)
	}
	files, err := source.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	files[ExportMappingRefresh] = []byte(`{"last_agent_refresh_at":"2026-10-01T09:00:00Z","pending":{"first_trigger_at":"2026-10-01T10:00:00Z","last_trigger_at":"2026-10-01T11:00:00Z","reasons":["manifest-added:go.mod"],"revisions":["r1"]}}` + "\n")
	delete(files, ExportManifest)
	files[ExportManifest] = RenderManifest(ManifestMeta{}, files)
	target, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	if err := target.Import(ctx, files); err != nil {
		t.Fatalf("pre-2.6 baseline must import: %v", err)
	}
	got, err := target.LoadMappingRefresh(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.LastAgentRefreshAt == nil || !got.LastAgentRefreshAt.Equal(last) {
		t.Fatalf("imported record = %+v", got)
	}
	again, err := target.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(again[ExportMappingRefresh], []byte("pending")) {
		t.Fatalf("re-export must drop pending: %s", again[ExportMappingRefresh])
	}
}

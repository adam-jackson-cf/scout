package state

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"scout/internal/domain"
)

// withManifest returns a copy of files whose manifest is the original with
// mutate applied to its fields. The listed files and their digests are unchanged.
func withManifest(t *testing.T, files map[string][]byte, mutate func(map[string]any)) map[string][]byte {
	t.Helper()
	fields := map[string]any{}
	if err := unmarshalNumbers(files[ExportManifest], &fields); err != nil {
		t.Fatal(err)
	}
	mutate(fields)
	line, err := CanonicalLine(fields)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string][]byte{}
	for path, data := range files {
		out[path] = data
	}
	out[ExportManifest] = line
	return out
}

func withoutLaneHealth(t *testing.T, files map[string][]byte) map[string][]byte {
	t.Helper()
	return withManifest(t, files, func(fields map[string]any) { delete(fields, "uninspected") })
}

func exportMeta(t *testing.T, store *Store) (ManifestMeta, []byte) {
	t.Helper()
	files, err := store.Export(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	meta, err := ReadManifestMeta(files[ExportManifest])
	if err != nil {
		t.Fatal(err)
	}
	return meta, files[ExportManifest]
}

func requireNoIdentifiers(t *testing.T, meta ManifestMeta, manifest []byte) {
	t.Helper()
	if meta.Revision != "" || meta.ToolDigest != "" || meta.PolicyDigest != "" {
		t.Fatalf("identifiers re-emitted without lane health: %+v", meta)
	}
	if meta.Uninspected == nil || len(meta.Uninspected) != 0 || !bytes.Contains(manifest, []byte(`"uninspected":[]`)) {
		t.Fatalf("manifest must carry uninspected: [] beside empty identifiers: %s", manifest)
	}
}

func TestRenderManifestAlwaysWritesUninspected(t *testing.T) {
	manifest := RenderManifest(ManifestMeta{}, map[string][]byte{})
	if !bytes.Contains(manifest, []byte(`"uninspected":[]`)) {
		t.Fatalf("manifest omits uninspected: %s", manifest)
	}
}

// A manifest without `uninspected` (a v0.1.0/v0.1.1 publish) is imported, but
// none of its comparison identifiers are recorded: they only travel together
// with lane health, so a later publish without a run writes empty identifiers.
func TestRestoreOfManifestWithoutLaneHealthImportsRecordsButNoIdentifiers(t *testing.T) {
	ctx := context.Background()
	source := seedLedger(t, ctx, t.TempDir())
	defer source.Close()
	full, err := source.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if meta, _ := ReadManifestMeta(full[ExportManifest]); meta.Revision == "" || meta.Uninspected == nil {
		t.Fatalf("fixture export lacks identifiers or lane health: %+v", meta)
	}

	target, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	if err := target.Import(ctx, withoutLaneHealth(t, full)); err != nil {
		t.Fatalf("restore of a manifest without uninspected failed: %v", err)
	}
	if issues, err := target.ListIssues(ctx); err != nil || len(issues) != 6 {
		t.Fatalf("restored %d issues, %v; want 6", len(issues), err)
	}
	meta, manifest := exportMeta(t, target)
	requireNoIdentifiers(t, meta, manifest)
}

// A manifest that carries `uninspected` is re-emitted with its identifiers and
// the same list when the next publish has no finalized run.
func TestRestoreOfManifestWithLaneHealthReemitsIdentifiersAndList(t *testing.T) {
	ctx := context.Background()
	source := seedLedger(t, ctx, t.TempDir())
	defer source.Close()
	full, err := source.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	list := []UninspectedFamily{{Family: domain.FamilySecurity, Reason: "blocked: gitleaks missing"}}
	files := withManifest(t, full, func(fields map[string]any) { fields["uninspected"] = list })
	for _, name := range []string{"non-empty", "empty"} {
		t.Run(name, func(t *testing.T) {
			want := list
			input := files
			if name == "empty" {
				want = []UninspectedFamily{}
				input = withManifest(t, full, func(fields map[string]any) { fields["uninspected"] = []any{} })
			}
			target, err := Open(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			defer target.Close()
			if err := target.Import(ctx, input); err != nil {
				t.Fatal(err)
			}
			meta, _ := exportMeta(t, target)
			if meta.Revision != "abc123" || meta.ToolDigest != "tool-digest" || meta.PolicyDigest != "policy-digest" {
				t.Fatalf("imported identifiers not re-emitted: %+v", meta)
			}
			if !reflect.DeepEqual(meta.Uninspected, want) {
				t.Fatalf("uninspected = %#v, want %#v", meta.Uninspected, want)
			}
		})
	}
}

// A store at the current version that holds an export_base written by an
// earlier build from a manifest without the field ("uninspected": null, or the
// key absent) counts as holding no base.
func TestRecordedBaseWithoutLaneHealthIsTreatedAsNone(t *testing.T) {
	ctx := context.Background()
	for name, recorded := range map[string]string{
		"null":   `{"revision":"r1","toolDigest":"t1","policyDigest":"p1","uninspected":null}`,
		"absent": `{"revision":"r1","toolDigest":"t1","policyDigest":"p1"}`,
	} {
		t.Run(name, func(t *testing.T) {
			store, err := Open(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			if _, err := store.db.ExecContext(ctx, `INSERT INTO metadata(key, value) VALUES (?, ?)`, metadataExportBase, recorded); err != nil {
				t.Fatal(err)
			}
			meta, manifest := exportMeta(t, store)
			requireNoIdentifiers(t, meta, manifest)
		})
	}
}

// A second import of the same repository into a store that took a zero-issue
// first import must not re-emit the first manifest's identifiers beside the
// second manifest's records.
func TestSecondImportReplacesOrClearsTheRecordedBase(t *testing.T) {
	ctx := context.Background()
	empty, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Close()
	zeroIssues, err := empty.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	first := withManifest(t, zeroIssues, func(fields map[string]any) {
		fields["repositoryId"], fields["revision"], fields["toolDigest"], fields["policyDigest"] = "repository", "first-rev", "first-tool", "first-policy"
		fields["uninspected"] = []any{}
	})

	source := seedLedger(t, ctx, t.TempDir())
	defer source.Close()
	full, err := source.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}

	target, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	if err := target.Import(ctx, first); err != nil {
		t.Fatal(err)
	}
	if meta, _ := exportMeta(t, target); meta.Revision != "first-rev" {
		t.Fatalf("first import not recorded: %+v", meta)
	}
	if err := target.Import(ctx, withoutLaneHealth(t, full)); err != nil {
		t.Fatalf("second import: %v", err)
	}
	meta, manifest := exportMeta(t, target)
	requireNoIdentifiers(t, meta, manifest)
	if issues, err := target.ListIssues(ctx); err != nil || len(issues) != 6 {
		t.Fatalf("second manifest's records missing: %d, %v", len(issues), err)
	}
	var recorded int
	if err := target.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM metadata WHERE key = ?`, metadataExportBase).Scan(&recorded); err != nil || recorded != 0 {
		t.Fatalf("export_base left behind: %d, %v", recorded, err)
	}
}

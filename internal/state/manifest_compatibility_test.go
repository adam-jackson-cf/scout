package state

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"scout/internal/areas"
	"scout/internal/domain"
)

// Frozen v1 wire fixture: not a current export with its schema rewound.
func adr012S2V1(t *testing.T) map[string][]byte {
	t.Helper()
	files := map[string][]byte{
		ExportIssues: nil, ExportDispositions: nil, ExportWouldBlock: nil,
		ExportCoverageSummary: []byte("[]\n"),
	}
	entries := []map[string]any{}
	for _, path := range []string{ExportCoverageSummary, ExportDispositions, ExportIssues, ExportWouldBlock} {
		sum := sha256.Sum256(files[path])
		n := 0
		if strings.HasSuffix(path, ".json") {
			n = 1
		}
		entries = append(entries, map[string]any{"path": path, "records": n, "sha256": hex.EncodeToString(sum[:])})
	}
	files[ExportManifest], _ = CanonicalLine(map[string]any{
		"schema": "scout.state/v1", "files": entries, "wouldBlock": "present",
		"repositoryId": "repository", "revision": "", "toolDigest": "", "policyDigest": "", "uninspected": []any{},
	})
	return files
}

func adr012S2Export(t *testing.T, s *Store) map[string][]byte {
	t.Helper()
	files, err := s.Export(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func adr012S2Rehash(t *testing.T, files map[string][]byte) {
	t.Helper()
	meta, err := ReadManifestMeta(files[ExportManifest])
	if err != nil {
		t.Fatal(err)
	}
	files[ExportManifest] = RenderManifest(meta, files)
}

func adr012S2Copy(files map[string][]byte) map[string][]byte {
	out := make(map[string][]byte, len(files))
	for path, data := range files {
		out[path] = bytes.Clone(data)
	}
	return out
}

func adr012S2Counts(t *testing.T, files map[string][]byte) {
	t.Helper()
	var manifest struct {
		Schema string `json:"schema"`
		Files  []struct {
			Path    string `json:"path"`
			Records int    `json:"records"`
			SHA256  string `json:"sha256"`
		} `json:"files"`
	}
	if err := json.Unmarshal(files[ExportManifest], &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Schema != "scout.state/v2" {
		t.Errorf("missing manifest v2: got %q", manifest.Schema)
	}
	seen := map[string]bool{}
	for _, entry := range manifest.Files {
		if seen[entry.Path] {
			t.Errorf("duplicate manifest entry %s", entry.Path)
		}
		seen[entry.Path] = true
		data, ok := files[entry.Path]
		if !ok {
			t.Errorf("manifest lists absent %s", entry.Path)
		}
		sum := sha256.Sum256(data)
		if entry.SHA256 != hex.EncodeToString(sum[:]) {
			t.Errorf("missing digest for %s", entry.Path)
		}
		n := len(SplitLines(data))
		if strings.HasSuffix(entry.Path, ".json") {
			n = 1
		}
		if entry.Records != n {
			t.Errorf("missing record count for %s: %d != %d", entry.Path, entry.Records, n)
		}
	}
	for path := range files {
		if !strings.HasPrefix(path, "state/v1/") {
			t.Errorf("layout moved from state/v1: %s", path)
		}
		if path != ExportManifest && !seen[path] {
			t.Errorf("unchecked export file %s", path)
		}
	}
	for _, path := range []string{"state/v1/pr-runs.jsonl", "state/v1/would-block-triage.jsonl"} {
		if !seen[path] {
			t.Errorf("missing required v2 file %s", path)
		}
	}
}

func TestADR012ManifestV2Compatibility(t *testing.T) {
	ctx := context.Background()
	t.Run("genuine_v1_is_empty_unmeasured", func(t *testing.T) {
		s := adr012Open(t, t.TempDir())
		if err := s.Import(ctx, adr012S2V1(t)); err != nil {
			t.Fatalf("genuine v1 refused: %v", err)
		}
		if got := adr012Ledger(t, s); len(got.Facts)+len(got.Events)+len(got.Triage)+len(got.Frontiers)+len(got.Publications) != 0 {
			t.Fatalf("v1 invented CI evidence/completeness: %#v", got)
		}
		adr012S2Counts(t, adr012S2Export(t, s))
	})

	t.Run("fresh_restore_preserves_all_history_and_ordinary_state", func(t *testing.T) {
		s := seedLedger(t, ctx, t.TempDir())
		t.Cleanup(func() { s.Close() })
		list, _ := areas.Reconcile(nil, []domain.Feature{{FeatureID: "cp", Title: "Security", Source: areas.AgentSource, OwnedFiles: []domain.FeatureFileRef{{Path: "a.go"}}}}, nil, areas.Config{Snapshot: "snap"})
		if err := s.ReplaceAreas(ctx, list); err != nil {
			t.Fatal(err)
		}
		if err := s.InsertAssociations(ctx, testAssociations()); err != nil {
			t.Fatal(err)
		}
		refreshAt := adr012Time.Add(-time.Hour)
		if err := s.SaveMappingRefresh(ctx, MappingRefresh{LastAgentRefreshAt: &refreshAt}); err != nil {
			t.Fatal(err)
		}
		ordinary := adr012S2Export(t, s)

		f := adr012Fact(domain.ShadowScopeCI, "1000")
		e := adr012Event(f, "ci-event")
		stamp := adr012Time.Add(2 * time.Hour)
		f.RecordedAt, e.RecordedAt = &stamp, &stamp
		p := domain.ShadowPublication{ID: "pub-1", RepositoryID: "repository", RecordedAt: stamp,
			Facts: []domain.PRRunFact{f}, Events: []domain.ProviderPREvent{e}, Links: []domain.PRRunEventLink{adr012Link(f, e)},
			Triage:    []domain.ShadowTriage{{ActionID: "1000", EventID: e.EventID, RepositoryID: "repository", Scope: domain.ShadowScopeCI, Verdict: domain.TriageWarranted, Actor: "maintainer", Reason: "reviewed", PolicyRevision: "policy-tip", TriagedAt: adr012Time, RecordedAt: &stamp}},
			Frontiers: []domain.ShadowFrontier{{RepositoryID: "repository", CheckID: f.CheckID, RuleVersion: "1", CompleteThrough: adr012Time.Add(time.Hour), PolicyRevision: "policy-tip", MinDays: 14, MinObservations: 5}},
			GapPoints: []domain.ShadowGapPoint{{ID: "gap", RepositoryID: "repository", CheckID: f.CheckID, At: adr012Time.Add(-time.Hour), Reason: "deadline", RecordedAt: &stamp}},
			Intervals: []domain.ShadowInterval{{ID: "interval", RepositoryID: "repository", CheckID: f.CheckID, RuleVersion: "1", After: adr012Time.Add(-time.Hour), Through: adr012Time.Add(time.Hour)}},
		}
		if err := s.ApplyShadowPublication(ctx, p); err != nil {
			t.Fatal(err)
		}
		refusal := domain.ShadowImportRefusal{ID: "refusal", RepositoryID: "repository", Scope: domain.ShadowScopeCI, SourceLineage: "foreign", Original: p.Triage[0], Reason: "author revoked", PolicyRevision: "policy-tip", RecordedAt: stamp}
		if err := s.RecordShadowInputs(ctx, domain.ShadowInputs{ImportRefusals: []domain.ShadowImportRefusal{refusal}}); err != nil {
			t.Fatal(err)
		}
		withoutLocal := adr012S2Export(t, s)
		local := adr012Fact(domain.ShadowScopeLocal, "999")
		localEvent := adr012Event(local, "local-only-event")
		adr012Record(t, s, local, []domain.ProviderPREvent{localEvent}, []domain.PRRunEventLink{adr012Link(local, localEvent)})
		files := adr012S2Export(t, s)
		if !reflect.DeepEqual(files, withoutLocal) {
			t.Error("local-scope rows leaked into shared export")
		}
		adr012S2Counts(t, files)
		for _, path := range []string{ExportIssues, ExportDispositions, ExportWouldBlock, ExportCoverageSummary, ExportAreas, ExportAssociations, ExportMappingRefresh} {
			if !bytes.Equal(ordinary[path], files[path]) {
				t.Errorf("shadow export changed ordinary state/guard/health: %s", path)
			}
		}
		beforeMeta, _ := ReadManifestMeta(ordinary[ExportManifest])
		afterMeta, _ := ReadManifestMeta(files[ExportManifest])
		if !reflect.DeepEqual(beforeMeta, afterMeta) {
			t.Error("shadow export changed baseline identifiers/lane health")
		}
		dst := adr012Open(t, t.TempDir())
		if err := dst.Import(ctx, files); err != nil {
			t.Fatalf("fresh v2 restore: %v", err)
		}
		if got := adr012S2Export(t, dst); !reflect.DeepEqual(files, got) {
			t.Error("v2 fresh export/import/export is not byte-identical")
		}
		want := adr012Ledger(t, s)
		want.Facts = []domain.PRRunFact{f}
		want.Events = []domain.ProviderPREvent{e}
		want.Links = p.Links
		if got := adr012Ledger(t, dst); !reflect.DeepEqual(want, got) {
			t.Errorf("v2 lost scope/pins/links/frontiers/gaps/publications/triage: %#v", got)
		}
		inputs, err := dst.ReadShadowInputs(ctx)
		if err != nil || !reflect.DeepEqual(inputs.ImportRefusals, []domain.ShadowImportRefusal{refusal}) {
			t.Errorf("v2 lost immutable import refusal: %#v %v", inputs, err)
		}
	})

	t.Run("raw_legacy_bytes_not_reserialized", func(t *testing.T) {
		s := seedLedger(t, ctx, t.TempDir())
		t.Cleanup(func() { s.Close() })
		files := adr012S2Export(t, s)
		// Legal noncanonical whitespace and an opaque legacy field must survive.
		lines := SplitLines(files[ExportWouldBlock])
		for i, line := range lines {
			line = bytes.TrimSpace(line)
			lines[i] = append(append([]byte("  { \"legacyOpaque\": 1000000000000000000000001, "), line[1:len(line)-1]...), []byte(" }\n")...)
		}
		files[ExportWouldBlock] = bytes.Join(lines, nil)
		adr012S2Rehash(t, files)
		dst := adr012Open(t, t.TempDir())
		if err := dst.Import(ctx, files); err != nil {
			t.Fatal(err)
		}
		if got := adr012S2Export(t, dst); !bytes.Equal(files[ExportWouldBlock], got[ExportWouldBlock]) {
			t.Error("raw legacy event bytes were reserialized")
		}
	})

	t.Run("tampering_and_count_mismatch_rejected_atomically", func(t *testing.T) {
		files := adr012S2Export(t, adr012Open(t, t.TempDir()))
		for _, kind := range []string{"tamper", "count"} {
			t.Run(kind, func(t *testing.T) {
				bad := adr012S2Copy(files)
				if kind == "tamper" {
					bad[ExportIssues] = []byte("{}\n")
				} else {
					bad = withManifest(t, bad, func(m map[string]any) { m["files"].([]any)[0].(map[string]any)["records"] = 999 })
				}
				dst := adr012Open(t, t.TempDir())
				before := adr012S2Export(t, dst)
				if err := dst.Import(ctx, bad); err == nil {
					t.Errorf("missing %s rejection", kind)
				}
				if !reflect.DeepEqual(before, adr012S2Export(t, dst)) {
					t.Error("invalid import mutated destination")
				}
			})
		}
	})

	t.Run("dangling_link_and_conflicting_final_rejected", func(t *testing.T) {
		s := adr012Open(t, t.TempDir())
		f := adr012Fact(domain.ShadowScopeCI, "999")
		e := adr012Event(f, "event")
		adr012Record(t, s, f, []domain.ProviderPREvent{e}, []domain.PRRunEventLink{adr012Link(f, e)})
		files := adr012S2Export(t, s)
		// Mutate the actual exporter-selected record-kind representation rather
		// than defining a second wire convention in the test.
		for _, kind := range []string{"dangling", "conflicting_final"} {
			t.Run(kind, func(t *testing.T) {
				bad := adr012S2Copy(files)
				changed := false
				for path, data := range bad {
					if path == ExportManifest || path == ExportWouldBlock {
						continue
					}
					for _, line := range SplitLines(data) {
						if kind == "dangling" && adr012S2HasLink(line) {
							bad[path] = bytes.Replace(data, line, adr012S2BreakLinks(t, line), 1)
							changed = true
							break
						}
						if kind == "conflicting_final" && bytes.Contains(line, []byte(`"Inspection":"inspected"`)) {
							bad[path] = append(bytes.Clone(data), append(bytes.ReplaceAll(line, []byte(`"Inspection":"inspected"`), []byte(`"Inspection":"not_inspected"`)), '\n')...)
							changed = true
							break
						}
					}
					if changed {
						break
					}
				}
				if !changed {
					t.Fatal("missing exported immutable fact/link records")
				}
				adr012S2Rehash(t, bad)
				dst := adr012Open(t, t.TempDir())
				before := adr012Ledger(t, dst)
				if err := dst.Import(ctx, bad); err == nil {
					t.Errorf("missing %s import rejection", kind)
				}
				adr012Unchanged(t, dst, before)
			})
		}
	})
}

// Follow the exported record-kind container, but change only a link, never the
// immutable event it names (which may share a publication envelope).
func adr012S2HasLink(line []byte) bool {
	var value any
	if json.Unmarshal(line, &value) != nil {
		return false
	}
	return adr012S2WalkLink(value, false)
}

func adr012S2BreakLinks(t *testing.T, line []byte) []byte {
	t.Helper()
	var value any
	if err := json.Unmarshal(line, &value); err != nil {
		t.Fatal(err)
	}
	if !adr012S2WalkLink(value, true) {
		t.Fatal("fixture contains no attempt/event link")
	}
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func adr012S2WalkLink(value any, mutate bool) bool {
	found := false
	switch node := value.(type) {
	case map[string]any:
		attempt, event := "", ""
		for key := range node {
			if strings.EqualFold(key, "AttemptID") {
				attempt = key
			}
			if strings.EqualFold(key, "EventID") {
				event = key
			}
		}
		if attempt != "" && event != "" {
			found = true
			if mutate {
				node[event] = "absent"
			}
		}
		for _, child := range node {
			if adr012S2WalkLink(child, mutate) {
				found = true
			}
		}
	case []any:
		for _, child := range node {
			if adr012S2WalkLink(child, mutate) {
				found = true
			}
		}
	}
	return found
}

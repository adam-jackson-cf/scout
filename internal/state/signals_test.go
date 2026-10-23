package state

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"scout/internal/domain"
)

func TestSignalsAreAppendOnlyFilterableAndLatestPerScope(t *testing.T) {
	ctx := context.Background()
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	first := testSignals()
	if err := store.RecordSignals(ctx, first); err != nil {
		t.Fatal(err)
	}
	// Same identity and digests with a different value must not overwrite.
	changed := testSignals()
	changed[1].Value = domain.Float(99)
	if err := store.RecordSignals(ctx, changed); err != nil {
		t.Fatal(err)
	}
	all, err := store.ListSignals(ctx, SignalFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("recorded %d signals, want 3 (idempotent)", len(all))
	}
	for _, signal := range all {
		if signal.MetricID == "scout.cc" && *signal.Value != 1 {
			t.Fatalf("append-only violated: value %v", *signal.Value)
		}
	}

	// A later snapshot of the same scope and metric appends.
	later := testSignals()[1]
	later.Snapshot, later.Value = "def456", domain.Float(5)
	if err := store.RecordSignals(ctx, []domain.Signal{later}); err != nil {
		t.Fatal(err)
	}
	cc, err := store.ListSignals(ctx, SignalFilter{MetricID: "scout.cc", ScopeAnchor: "pkg/sub"})
	if err != nil || len(cc) != 2 {
		t.Fatalf("history of scout.cc = %d, %v", len(cc), err)
	}
	latest, err := store.ListSignals(ctx, SignalFilter{MetricID: "scout.cc@1", ScopeAnchor: "pkg/sub", Latest: true})
	if err != nil || len(latest) != 1 || latest[0].Snapshot != "def456" || *latest[0].Value != 5 {
		t.Fatalf("latest = %+v, %v", latest, err)
	}
	if other, _ := store.ListSignals(ctx, SignalFilter{MetricID: "scout.cc@2"}); len(other) != 0 {
		t.Fatal("metric version filter matched another version")
	}

	// Missingness survives storage: unmeasured carries no numbers.
	crap, err := store.ListSignals(ctx, SignalFilter{MetricID: "scout.crap"})
	if err != nil || len(crap) != 1 || crap[0].Missing != domain.SignalUnmeasured || crap[0].Value != nil || crap[0].Numerator != nil {
		t.Fatalf("crap = %+v, %v", crap, err)
	}
}

func TestSignalValidationRejectsMissingMasqueradingAsMeasured(t *testing.T) {
	ctx := context.Background()
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	good := testSignals()[1]
	for name, mutate := range map[string]func(*domain.Signal){
		"measured without value": func(s *domain.Signal) { s.Value = nil },
		"unmeasured with a zero": func(s *domain.Signal) { s.Missing, s.MissingReason = domain.SignalUnmeasured, "tool absent" },
		"unmeasured without reason": func(s *domain.Signal) {
			s.Value, s.Numerator, s.Denominator, s.Missing = nil, nil, nil, domain.SignalUnmeasured
		},
		"missing population digests": func(s *domain.Signal) { s.ToolDigest = "" },
		"no claim":                   func(s *domain.Signal) { s.Claim = "" },
	} {
		signal := good
		mutate(&signal)
		if err := store.RecordSignals(ctx, []domain.Signal{signal}); err == nil {
			t.Fatalf("%s: accepted", name)
		}
	}
	if all, _ := store.ListSignals(ctx, SignalFilter{}); len(all) != 0 {
		t.Fatalf("rejected signals were stored: %d", len(all))
	}
}

func TestExportCarriesSignalsPerMetricAndManifestListsThem(t *testing.T) {
	ctx := context.Background()
	store := seedLedger(t, ctx, t.TempDir())
	defer store.Close()
	files, err := store.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{ExportSignalsDir + "scout.cc@1.jsonl", ExportSignalsDir + "scout.crap@1.jsonl", ExportSignalsDir + "scout.changes@1.jsonl"} {
		if len(files[path]) == 0 {
			t.Fatalf("export lacks %s", path)
		}
	}
	if bytes.Contains(files[ExportSignalsDir+"scout.cc@1.jsonl"], []byte(`\u003c`)) {
		t.Fatal("signal export must not HTML-escape")
	}
	var manifest struct {
		Absent []string `json:"absent"`
		Files  []struct {
			Path string `json:"path"`
		} `json:"files"`
	}
	if err := json.Unmarshal(files[ExportManifest], &manifest); err != nil {
		t.Fatal(err)
	}
	listed := false
	for _, entry := range manifest.Files {
		listed = listed || entry.Path == ExportSignalsDir+"scout.cc@1.jsonl"
	}
	if !listed {
		t.Fatal("manifest does not list the signals file")
	}
	for _, path := range manifest.Absent {
		if path == ExportSignalsDir {
			t.Fatal("manifest reports signals absent while present")
		}
	}

	// Import restores the ledger including signals, and a store without
	// signals still reports them absent.
	target, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	if err := target.Import(ctx, files); err != nil {
		t.Fatal(err)
	}
	got, err := target.ListSignals(ctx, SignalFilter{})
	if err != nil || len(got) != 3 {
		t.Fatalf("imported %d signals, %v", len(got), err)
	}
	empty, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Close()
	bare, err := empty.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(bare[ExportManifest], []byte(`"state/v1/signals/"`)) {
		t.Fatal("empty export must list signals as absent")
	}
}

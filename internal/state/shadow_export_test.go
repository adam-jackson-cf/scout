package state

import (
	"context"
	"reflect"
	"testing"
	"time"

	"scout/internal/domain"
)

func TestShadowImportStagesUnprovenHistory(t *testing.T) {
	ctx := context.Background()
	source := adr012Open(t, t.TempDir())
	f := adr012Fact(domain.ShadowScopeCI, "1000")
	stamp := adr012Time.Add(time.Hour)
	f.RecordedAt = &stamp
	publication := domain.ShadowPublication{ID: "source-genesis", RepositoryID: f.ID.RepositoryID, RecordedAt: stamp, Facts: []domain.PRRunFact{f}}
	if err := source.ApplyShadowPublication(ctx, publication); err != nil {
		t.Fatal(err)
	}
	files := adr012S2Export(t, source)
	destination := adr012Open(t, t.TempDir())
	pending := adr012Fact(domain.ShadowScopeCI, "999")
	adr012Record(t, destination, pending, nil, nil)
	before := adr012Ledger(t, destination)
	for range 2 {
		if err := destination.Import(ctx, files); err != nil {
			t.Fatal(err)
		}
		if got := adr012Ledger(t, destination); !reflect.DeepEqual(before, got) {
			t.Fatal("unproven import altered destination evidence/stamps")
		}
		inputs, err := destination.ReadShadowInputs(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(inputs.PendingImports) != 1 || inputs.PendingImports[0].History == nil {
			t.Fatalf("whole source not staged exactly once: %#v", inputs)
		}
		if !reflect.DeepEqual(*inputs.PendingImports[0].History, adr012Ledger(t, source)) {
			t.Fatal("pending source history lost provenance")
		}
	}
	exported := adr012S2Export(t, destination)
	restored := adr012Open(t, t.TempDir())
	if err := restored.Import(ctx, exported); err != nil {
		t.Fatal(err)
	}
	if got := adr012S2Export(t, restored); !reflect.DeepEqual(got, exported) {
		t.Fatal("pending whole-history import did not round trip")
	}
}

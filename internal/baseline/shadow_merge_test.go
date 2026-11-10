package baseline

import (
	"reflect"
	"testing"

	"scout/internal/domain"
	"scout/internal/state"
)

func shadowTree(t *testing.T, l domain.ShadowLedger) Tree {
	t.Helper()
	tree := ledger(t, "2026-11-08T00:00:00Z")
	files, err := state.EncodeShadow(l, domain.ShadowInputs{})
	if err != nil {
		t.Fatal(err)
	}
	for path, data := range files {
		tree[path] = data
	}
	tree[state.ExportManifest] = state.RenderManifest(state.ManifestMeta{RepositoryID: "repository", WouldBlock: state.WouldBlockPresent}, tree)
	return tree
}

func TestMergeShadowStagesWithoutStamping(t *testing.T) {
	destination := adr012PlanLedger()
	source := adr012PlanClone(t, destination)
	source.Facts[0].ID.RunID = "1000"
	source.Links[0].AttemptID = source.Facts[0].ID
	source.Publications[0].ID = "foreign-genesis"
	source.Publications[0].Facts = source.Facts
	source.Publications[0].Links = source.Links
	current, local := shadowTree(t, destination), shadowTree(t, source)
	beforeCurrent, beforeLocal := adr012PlanClone(t, destination), adr012PlanClone(t, source)
	merged, err := MergeState(current, local)
	if err != nil {
		t.Fatal(err)
	}
	got, inputs, err := state.DecodeShadow(merged)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, destination) {
		t.Fatal("pure merge adopted foreign authority or changed a destination stamp")
	}
	if len(inputs.PendingImports) != 1 || inputs.PendingImports[0].History == nil || !reflect.DeepEqual(*inputs.PendingImports[0].History, source) {
		t.Fatal("pure merge lost pending source history")
	}
	replay, err := MergeState(merged, local)
	if err != nil || !reflect.DeepEqual(replay, merged) {
		t.Fatalf("merge replay changed history: %v", err)
	}
	if !reflect.DeepEqual(destination, beforeCurrent) || !reflect.DeepEqual(source, beforeLocal) {
		t.Fatal("merge mutated caller history")
	}
}

func TestMergeShadowRejectsConflictingFinalAcrossSources(t *testing.T) {
	destination := adr012PlanLedger()
	source := adr012PlanClone(t, destination)
	source.Facts[0].Inspection = "not_inspected"
	source.Publications[0].Facts = source.Facts
	if _, err := MergeState(shadowTree(t, destination), shadowTree(t, source)); err == nil {
		t.Fatal("conflicting immutable final accepted")
	}
}

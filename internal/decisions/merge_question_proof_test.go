package decisions

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"testing"

	"scout/internal/areas"
)

func mergeQuestionCatalogue() []areas.Area {
	return []areas.Area{
		{ScoutAreaID: "a", Title: "Core", Summary: "Request processing", FileSet: []string{"a.go", "b.go", "c.go", "d.go", "e.go", "f.go"}},
		{ScoutAreaID: "b", Title: "API", Summary: "Public endpoints", FileSet: []string{"api.go"}},
		{ScoutAreaID: "pending", Status: areas.StatusPending},
		{ScoutAreaID: "superseded", Status: areas.StatusSuperseded},
	}
}

// Frozen instruction digests are SHA256 of the decoded UTF-8 instructions
// field, not the JSON file bytes. No test reads evaluation/ at runtime.
func TestAssociationQuestionsFrozen(t *testing.T) {
	covered := MergeCovered
	choice := MergeWhichArea(mergeQuestionCatalogue())
	for _, tc := range []struct {
		name   string
		q      Question
		id     string
		kind   Kind
		digest string
	}{
		{name: "covered frozen question", q: covered, id: "merge.covered", kind: KindNoul, digest: "bc38de7e196d644c84d53fed60690aeb7926fa425db89c5f34eabb83bdb837ff"},
		{name: "choice frozen question", q: choice, id: "merge.which_area", kind: KindChoice, digest: "65150f590c54dd77980259578777c4e3362f1667e4f9ca289f00c3aa102d6273"},
	} {
		if tc.q.ID != tc.id || tc.q.Kind != tc.kind {
			t.Fatalf("question identity %+v", tc.q)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256([]byte(tc.q.Instructions))); got != tc.digest {
			t.Errorf("%s frozen text changed: %s", tc.id, got)
		}
		if err := tc.q.Validate(); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(covered.Criteria, map[string]string{"true": "At least one listed area's responsibility covers these files.", "false": "No listed area's responsibility covers them. Sharing a directory, file type, naming pattern or import with an area's files is not sufficient."}) {
		t.Fatalf("covered criteria %+v", covered.Criteria)
	}
	if covered.Order.Rule != OrderFixed || !reflect.DeepEqual(covered.Order.Options, []string{"true", "false"}) || covered.Positive != "true" {
		t.Fatalf("covered order/positive %+v", covered)
	}
	if choice.Order.Rule != OrderPermutation || len(choice.Criteria) != 2 || choice.Criteria["a"] != "Core: Request processing Sample files: a.go, b.go, c.go, d.go, e.go" || choice.Criteria["b"] != "API: Public endpoints Sample files: api.go" {
		t.Fatalf("whole active catalogue criteria %+v", choice.Criteria)
	}
	shape := []string{"state.diff_summary", "state.pr.body", "state.pr.changed_paths", "state.pr.title", "state.unit.change_kind", "state.unit.excerpt", "state.unit.paths", "state.unit.role"}
	if !reflect.DeepEqual(choice.StateShape, shape) {
		t.Fatalf("choice state shape %v", choice.StateShape)
	}
	if !reflect.DeepEqual(covered.StateShape, append([]string{"state.catalogue"}, shape...)) {
		t.Fatalf("covered state shape %v", covered.StateShape)
	}
}

package association

import (
	"reflect"
	"testing"
)

func TestUnportedRules(t *testing.T) {
	own := newFixtureOwnership(fxRound{Owners: map[string][]string{"axum/src/a.rs": {"A"}}})
	// Cross-scope rename (crate source -> examples): the pilot sent it to the residual; Scout inherits like any rename.
	r := Classify([]Change{{Kind: Renamed, OldPath: "axum/src/a.rs", Path: "examples/a.rs"}}, own, func(string) bool { return false })
	if len(r.Residual) != 0 || len(r.Inherited) != 1 || r.Inherited[0].Areas[0] != "A" {
		t.Fatalf("cross-scope rename: %+v", r)
	}
	// docs/<lang>/ non-markdown files are eligible (no translation rule).
	r = Classify([]Change{{Kind: Added, Path: "docs/hi/mkdocs.yml"}, {Kind: Added, Path: "docs/api/x.png"}}, own, func(string) bool { return false })
	if len(r.Residual) != 1 || r.Residual[0].Path != "docs/hi/mkdocs.yml" || r.Counts["excluded:media"] != 1 {
		t.Fatalf("translation path: %+v", r)
	}
}

func TestRules(t *testing.T) {
	own := newFixtureOwnership(fxRound{
		Owners:        map[string][]string{"docs_src/t/tutorial001.py": {"A"}, "own.go": {"B"}, "heur-owned.go": {"B"}},
		HeuristicOnly: []string{"heur.go", "heur-owned.go", "pooled-heur.go"},
		Pool:          []string{"pooled.go", "pooled-heur.go"},
	})
	tree := map[string]bool{"docs_src/t/tutorial001.py": true}
	in := func(p string) bool { return tree[p] }
	r := Classify([]Change{
		{Kind: Deleted, Path: "gone.go"},
		{Kind: Modified, Path: "own.go"},
		{Kind: Modified, Path: "heur.go"},       // rule (f): decided
		{Kind: Modified, Path: "heur-owned.go"}, // owned wins
		{Kind: Modified, Path: "nobody.go"},     // coverage gap
		{Kind: Renamed, OldPath: "own.go", Path: "own2.go"},
		{Kind: Renamed, OldPath: "pooled.go", Path: "pooled2.go"},                // pending move, not decided
		{Kind: Renamed, OldPath: "pooled-heur.go", Path: "heur.go"},              // pending move even though heuristic-owned
		{Kind: Renamed, OldPath: "unknown.go", Path: "heur-new.go", Empty: true}, // empty: residual (inert alone)
		{Kind: Added, Path: "docs_src/t/tutorial001_an_py39.py"},                 // variant of an owned file? ext py, stem tutorial001
		{Kind: Added, Path: "docs_src/t/tutorial001_an_py39.txt"},                // different extension: residual
	}, own, in)
	want := map[string]int{"deleted": 1, "modified_owned": 2, "heuristic_only": 1, "modified_unowned": 1, "modified_unowned:never_owned": 1,
		"renamed_inherit": 1, "renamed_pooled": 2, "renamed_empty_trap": 1, "variant_inherit": 1, "residual_added": 1}
	if !reflect.DeepEqual(r.Counts, want) {
		t.Fatalf("counts\n got %v\nwant %v", r.Counts, want)
	}
	if !reflect.DeepEqual(r.CoverageGaps, []string{"nobody.go"}) || !reflect.DeepEqual(r.Deleted, []string{"gone.go"}) {
		t.Fatalf("gaps %v deleted %v", r.CoverageGaps, r.Deleted)
	}
	wantIn := []Inherited{
		{Path: "own2.go", Rule: "rename", From: "own.go", Areas: []string{"B"}},
		{Path: "docs_src/t/tutorial001_an_py39.py", Rule: "variant", From: "docs_src/t/tutorial001.py", Areas: []string{"A"}},
	}
	if !reflect.DeepEqual(r.Inherited, wantIn) {
		t.Fatalf("inherited %+v", r.Inherited)
	}
	if len(r.PendingMoves) != 2 || r.PendingMoves[0].New != "pooled2.go" || r.PendingMoves[1].New != "heur.go" {
		t.Fatalf("pending moves %+v", r.PendingMoves)
	}
	live, inert := BuildClusters("c", r.Residual)
	if len(live) != 2 || len(inert) != 1 || inert[0].Files[0] != "heur-new.go" {
		t.Fatalf("live %+v inert %+v", live, inert)
	}
}

func TestVariantSiblingMustExistInParentTree(t *testing.T) {
	own := newFixtureOwnership(fxRound{Owners: map[string][]string{"s/tutorial001.py": {"A"}}})
	r := Classify([]Change{{Kind: Added, Path: "s/tutorial001_py39.py"}}, own, func(string) bool { return false })
	if len(r.Inherited) != 0 || len(r.Residual) != 1 {
		t.Fatalf("sibling absent from the parent tree must not inherit: %+v", r)
	}
}

func TestStems(t *testing.T) {
	for in, want := range map[string]string{"tutorial003_an_py39": "tutorial003", "test_tutorial003_py310": "test_tutorial003", "tutorial003_01": "tutorial003", "_01": "_01", "": ""} {
		if got := VariantStem(in); got != want {
			t.Errorf("VariantStem(%q) = %q, want %q", in, got, want)
		}
	}
	if got := NormDir("docs_src/app_testing/app_b_an"); got != "docs_src/app_testing/app_b" {
		t.Errorf("NormDir = %q", got)
	}
}

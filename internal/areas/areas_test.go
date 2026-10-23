package areas

import (
	"fmt"
	"reflect"
	"testing"

	"scout/internal/domain"
)

func feature(id, source string, files ...string) domain.Feature {
	f := domain.Feature{FeatureID: id, Title: "title " + id, Source: source}
	for _, path := range files {
		f.OwnedFiles = append(f.OwnedFiles, domain.FeatureFileRef{Path: path})
	}
	return f
}

func paths(prefix string, from, to int) []string {
	var out []string
	for i := from; i < to; i++ {
		out = append(out, fmt.Sprintf("%s/f%02d.go", prefix, i))
	}
	return out
}

func find(t *testing.T, list []Area, id string) Area {
	t.Helper()
	for _, a := range list {
		if a.ScoutAreaID == id {
			return a
		}
	}
	t.Fatalf("area %s not found in %+v", id, list)
	return Area{}
}

func TestReconcileMatchesAtThresholdKeepsIDAndUpdatesFileSet(t *testing.T) {
	first, _ := Reconcile(nil, []domain.Feature{feature("a1", AgentSource, paths("svc", 0, 4)...)}, nil, Config{Snapshot: "s1"})
	if len(first) != 1 || first[0].FirstSeenSnapshot != "s1" || first[0].Origin != OriginAgent {
		t.Fatalf("onboarding areas = %+v", first)
	}
	id := first[0].ScoutAreaID

	// {0,1,2,3} vs {2,3,4,5}: intersection 2, union 6 -> 0.33 (below 0.5).
	below, report := Reconcile(first, []domain.Feature{feature("a2", AgentSource, paths("svc", 2, 6)...)}, nil, Config{Snapshot: "s2"})
	if len(report.Matched) != 0 || len(report.Created) != 1 || len(below) != 2 {
		t.Fatalf("below threshold must create a new area: %+v %+v", report, below)
	}

	// {0,1,2,3} vs {0,1,2,3,4,5}... 4/6 = 0.67 matches; exactly 0.5: {0,1} vs {0,1,2,3}.
	exact, report := Reconcile(first, []domain.Feature{feature("a3", AgentSource, paths("svc", 0, 2)...)}, nil, Config{Snapshot: "s2"})
	if len(report.Matched) != 1 || report.Matched[0].Jaccard != 0.5 || report.Matched[0].ScoutAreaID != id {
		t.Fatalf("Jaccard 0.5 must match: %+v", report)
	}
	area := find(t, exact, id)
	if !reflect.DeepEqual(area.FileSet, paths("svc", 0, 2)) || !reflect.DeepEqual(area.ClawpatchIDs, []string{"a3"}) || area.FirstSeenSnapshot != "s1" || area.Title != "title a3" {
		t.Fatalf("matched area not updated: %+v", area)
	}
	if len(first[0].FileSet) != 4 {
		t.Fatal("Reconcile mutated its input")
	}

	// A stricter configured threshold rejects the same pair.
	_, report = Reconcile(first, []domain.Feature{feature("a3", AgentSource, paths("svc", 0, 2)...)}, nil, Config{Snapshot: "s2", MatchJaccard: 0.6})
	if len(report.Matched) != 0 || len(report.Created) != 1 {
		t.Fatalf("configured threshold ignored: %+v", report)
	}
}

func TestReconcileOneToOneBestMatchWins(t *testing.T) {
	prev, _ := Reconcile(nil, []domain.Feature{
		feature("x", AgentSource, paths("p", 0, 6)...),
		feature("y", AgentSource, paths("p", 6, 12)...),
	}, nil, Config{Snapshot: "s1"})
	next, report := Reconcile(prev, []domain.Feature{
		feature("n1", AgentSource, paths("p", 0, 6)...),
		feature("n2", AgentSource, paths("p", 0, 3)...), // also overlaps x at 0.5 but x is taken by n1
	}, nil, Config{Snapshot: "s2"})
	if len(report.Matched) != 1 || report.Matched[0].FeatureID != "n1" || len(report.Created) != 1 {
		t.Fatalf("best match must take the area and the loser become new: %+v", report)
	}
	if len(next) != 3 {
		t.Fatalf("want 3 areas, got %+v", next)
	}
}

func TestReconcileHeuristicFeaturesDoNotBecomeAreasAndOwnedAreasRetire(t *testing.T) {
	prev, _ := Reconcile(nil, []domain.Feature{
		feature("a", AgentSource, "owned/a.go", "owned/b.go"),
		feature("b", AgentSource, "owned/c.go", "free/d.go"),
	}, nil, Config{Snapshot: "s1"})
	heuristic := []domain.Feature{feature("h", "python-source-group", "owned/a.go", "owned/b.go", "owned/c.go")}
	owned := HeuristicOwner(heuristic)

	next, report := Reconcile(prev, heuristic, owned, Config{Snapshot: "s2"})
	if len(report.Created) != 0 || len(report.Matched) != 0 {
		t.Fatalf("heuristic features must not become areas: %+v", report)
	}
	if len(report.Retired) != 1 || len(report.Kept) != 1 {
		t.Fatalf("retire only the fully heuristic-owned area: %+v", report)
	}
	retired := find(t, next, report.Retired[0])
	if retired.Active() || retired.RetiredSnapshot != "s2" || retired.Title != "title a" {
		t.Fatalf("retired area = %+v", retired)
	}
	if !find(t, next, report.Kept[0]).Active() {
		t.Fatal("partially unowned area must stay active")
	}

	// A retired area is never matched again; the file set becomes a new area.
	again, report := Reconcile(next, []domain.Feature{feature("a9", AgentSource, "owned/a.go", "owned/b.go")}, owned, Config{Snapshot: "s3"})
	if len(report.Matched) != 0 || len(report.Created) != 1 || len(again) != 3 {
		t.Fatalf("retired area was revived: %+v", report)
	}
}

func TestReconcileKeepsAreaIDsAcrossAgentIDChurn(t *testing.T) {
	// A fixture of 6 areas; every refresh gets entirely fresh agent feature IDs
	// and drifts a file or two (CP07 §5: agent IDs churn, file sets mostly stay).
	build := func(prefix string, drift int) []domain.Feature {
		var out []domain.Feature
		for i := range 6 {
			files := paths(fmt.Sprintf("mod%d", i), 0, 10)
			if drift > 0 {
				files = append(files[drift:], fmt.Sprintf("mod%d/extra.go", i)) // drop `drift` files, add one
			}
			out = append(out, feature(fmt.Sprintf("%s-%d", prefix, i), AgentSource, files...))
		}
		return out
	}
	one, _ := Reconcile(nil, build("run1", 0), nil, Config{Snapshot: "s1"})
	two, report2 := Reconcile(one, build("run2", 1), nil, Config{Snapshot: "s2"})
	three, report3 := Reconcile(two, build("run3", 2), nil, Config{Snapshot: "s3"})

	ids := func(list []Area) []string {
		var out []string
		for _, a := range list {
			out = append(out, a.ScoutAreaID)
		}
		return out
	}
	if !reflect.DeepEqual(ids(one), ids(two)) || !reflect.DeepEqual(ids(two), ids(three)) || len(report2.Created)+len(report3.Created) != 0 {
		t.Fatalf("area IDs churned: %v / %v / %v", ids(one), ids(two), ids(three))
	}
	for i := range one {
		for _, pair := range [][2][]string{{one[i].FileSet, two[i].FileSet}, {two[i].FileSet, three[i].FileSet}} {
			if j := Jaccard(pair[0], pair[1]); j < 0.8 {
				t.Fatalf("file-set Jaccard across refreshes = %.2f, want >= 0.8", j)
			}
		}
		if one[i].FirstSeenSnapshot != "s1" || three[i].FirstSeenSnapshot != "s1" {
			t.Fatal("first-seen snapshot must survive reconciliation")
		}
	}
	if reflect.DeepEqual(one[0].ClawpatchIDs, three[0].ClawpatchIDs) {
		t.Fatal("provenance must follow the latest agent feature IDs")
	}
}

func TestReconcileMintedIDsAreDeterministicAndUnique(t *testing.T) {
	features := []domain.Feature{feature("a", AgentSource, "x.go"), feature("b", AgentSource, "y.go")}
	one, _ := Reconcile(nil, features, nil, Config{Snapshot: "s1"})
	two, _ := Reconcile(nil, features, nil, Config{Snapshot: "s1"})
	if !reflect.DeepEqual(one, two) || one[0].ScoutAreaID == one[1].ScoutAreaID {
		t.Fatalf("minted IDs not deterministic/unique: %+v %+v", one, two)
	}
}

func TestStaleOwnershipNeverRetiresAnArea(t *testing.T) {
	owned := HeuristicOwner([]domain.Feature{feature("fresh", "python-source-group", "new/a.go")})
	if owned("old/a.go") || !owned("new/a.go") {
		t.Fatal("ownership must read only the given (active) records")
	}
	// A stale record neither retires an area nor counts as ownership.
	prev, _ := Reconcile(nil, []domain.Feature{feature("g", AgentSource, "old/a.go")}, nil, Config{Snapshot: "s1"})
	next, report := Reconcile(prev, nil, owned, Config{Snapshot: "s2"})
	if len(report.Retired) != 0 || !next[0].Active() {
		t.Fatalf("stale heuristic record retired an area: %+v", report)
	}
}

func TestUnownedScopesAreDirectoryTruncatedAndStable(t *testing.T) {
	owned := HeuristicOwner([]domain.Feature{feature("h", "python-source-group", "pkg/a/b/owned.go")})
	files := []string{"pkg/a/b/owned.go", "pkg/a/b/c/d/x.go", "pkg/a/b/c/y.go", "src/web/ui/z/w.go", "main.go", "docs/readme.md"}
	scopes := UnownedScopes(files, owned)
	got := map[string][]string{}
	for _, scope := range scopes {
		got[scope.Directory] = scope.Files
	}
	want := map[string][]string{
		"pkg/a":      {"pkg/a/b/c/d/x.go", "pkg/a/b/c/y.go"},
		"src/web/ui": {"src/web/ui/z/w.go"},
		".":          {"main.go"},
		"docs":       {"docs/readme.md"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unowned scopes = %v, want %v", got, want)
	}
	// The scope key depends only on the path: adding or reconciling areas
	// cannot re-key it.
	prev, _ := Reconcile(nil, []domain.Feature{feature("a", AgentSource, files...)}, owned, Config{Snapshot: "s1"})
	_, _ = Reconcile(prev, []domain.Feature{feature("b", AgentSource, files[:3]...)}, owned, Config{Snapshot: "s2"})
	if !reflect.DeepEqual(UnownedScopes(files, owned), scopes) {
		t.Fatal("directory scopes changed after a reconcile")
	}
}

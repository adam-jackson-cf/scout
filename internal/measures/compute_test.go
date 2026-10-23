package measures

import (
	"math"
	"reflect"
	"testing"

	"scout/internal/domain"
)

func file(path, digest string, features ...string) domain.File {
	return domain.File{Path: path, ContentDigest: digest, Eligible: true, OwnedFeatureIDs: features}
}

// fixture: pkg/sub/a.py (NLOC 20; CC 12 over 16 NLOC and CC 2 over 4 NLOC),
// pkg/sub/c.py (NLOC 30; one function CC 3 over 9 NLOC) owned by feature F,
// pkg/sub/small.py (NLOC 9: below the eligibility floor).
func fixture() Input {
	return Input{
		Snapshot: "rev1", RunScope: "repository",
		Candidates: []domain.File{file("pkg/sub/a.py", "da"), file("pkg/sub/c.py", "dc", "F"), file("pkg/sub/small.py", "ds")},
		Features:   []domain.Feature{{FeatureID: "F"}},
		Facts: Facts{
			Lizard: map[string]FileMeasure{
				"pkg/sub/a.py":     {Path: "pkg/sub/a.py", NLOC: 20, Functions: []Function{{Name: "f1", Line: 1, CC: 12, NLOC: 16}, {Name: "f2", Line: 20, CC: 2, NLOC: 4}}},
				"pkg/sub/c.py":     {Path: "pkg/sub/c.py", NLOC: 30, Functions: []Function{{Name: "f3", Line: 1, CC: 3, NLOC: 9}}},
				"pkg/sub/small.py": {Path: "pkg/sub/small.py", NLOC: 9, Functions: []Function{{Name: "tiny", Line: 1, CC: 20, NLOC: 9}}},
			},
			Lines: map[string]int{"pkg/sub/a.py": 100, "pkg/sub/c.py": 50},
			Clones: []Clone{
				{A: Span{"pkg/sub/a.py", 10, 29}, B: Span{"pkg/sub/c.py", 1, 20}},
				{A: Span{"pkg/sub/a.py", 25, 34}, B: Span{"pkg/sub/c.py", 40, 49}},
			},
		},
	}
}

func find(t *testing.T, signals []domain.Signal, metric, anchor string) domain.Signal {
	t.Helper()
	for _, signal := range signals {
		if signal.MetricID == metric && signal.ScopeAnchor == anchor {
			return signal
		}
	}
	t.Fatalf("no %s signal for %s in %d signals", metric, anchor, len(signals))
	return domain.Signal{}
}

func near(t *testing.T, name string, got *float64, want float64) {
	t.Helper()
	if got == nil || math.Abs(*got-want) > 1e-12 {
		t.Fatalf("%s = %v, want %v", name, got, want)
	}
}

func TestPrincipalMeasureFormulasOnHandComputedFixture(t *testing.T) {
	signals := Compute(fixture())
	for _, signal := range signals {
		if err := signal.Validate(); err != nil {
			t.Fatalf("%s %s invalid: %v", signal.ID(), signal.ScopeAnchor, err)
		}
		if signal.Claim != domain.ClaimDescriptive {
			t.Fatalf("%s claim = %q", signal.ID(), signal.Claim)
		}
	}

	// Directory scope pkg/sub owns only a.py (c.py belongs to feature F).
	dir := "pkg/sub"
	cc := find(t, signals, MetricCC, dir)
	// n_high_cc = 1 of 2 functions; sorted CC [2 12]: p90 = rank ceil(1.8) = 2 → 12.
	near(t, "cc value", cc.Value, 1)
	near(t, "cc numerator", cc.Numerator, 1)
	near(t, "cc denominator", cc.Denominator, 2)
	if cc.Detail["cc_p90"] != 12 || cc.Detail["cc_max"] != 12 || cc.Detail["n_functions"] != 2 || cc.Detail["n_high_cc"] != 1 {
		t.Fatalf("cc detail = %v", cc.Detail)
	}
	// B = 12·√16 / (12·√16 + 2·√4) = 48/52.
	burden := find(t, signals, MetricComplexBurden, dir)
	near(t, "burden numerator", burden.Numerator, 48)
	near(t, "burden denominator", burden.Denominator, 52)
	near(t, "burden", burden.Value, 48.0/52.0)
	// Q = 48 / file NLOC 20 = 2.4; small.py (NLOC 9, CC 20) is not eligible.
	density := find(t, signals, MetricBurdenDensity, dir)
	near(t, "density numerator", density.Numerator, 48)
	near(t, "density denominator", density.Denominator, 20)
	near(t, "density", density.Value, 2.4)

	// Feature F: c.py alone: CC 3 is not high, so B = 0 / (3·√9) = 0.
	f := FeatureAnchor("F")
	fb := find(t, signals, MetricComplexBurden, f)
	near(t, "F burden denominator", fb.Denominator, 9)
	near(t, "F burden", fb.Value, 0)
	near(t, "F n_high_cc", find(t, signals, MetricCC, f).Value, 0)
	near(t, "F density", find(t, signals, MetricBurdenDensity, f).Value, 0)
}

func TestDuplicationUnionsSpansAndFlagsCrossScopeClones(t *testing.T) {
	signals := Compute(fixture())
	// a.py: [10,29] ∪ [25,34] = lines 10..34 = 25; c.py: 20 + 10 = 30.
	dir := find(t, signals, MetricDupPct, "pkg/sub")
	near(t, "dir dup numerator", dir.Numerator, 25)
	near(t, "dir dup denominator", dir.Denominator, 100)
	near(t, "dir dup_pct", dir.Value, 0.25)
	if dir.Detail["clone_pairs"] != 2 || dir.Detail["cross_scope_clone_pairs"] != 2 {
		t.Fatalf("dir clone detail = %v", dir.Detail)
	}
	feature := find(t, signals, MetricDupPct, FeatureAnchor("F"))
	near(t, "F dup numerator", feature.Numerator, 30)
	near(t, "F dup denominator", feature.Denominator, 50)
	near(t, "F dup_pct", feature.Value, 0.6)
	if feature.Detail["cross_scope_clone_pairs"] != 2 {
		t.Fatalf("F clone detail = %v", feature.Detail)
	}

	// Without cross-scope ownership both ends are in one scope: not cross, and
	// the union covers both files: (25+30)/150.
	merged := fixture()
	merged.Candidates[1].OwnedFeatureIDs = nil
	same := find(t, Compute(merged), MetricDupPct, "pkg/sub")
	near(t, "single-scope dup", same.Value, 55.0/150.0)
	if same.Detail["cross_scope_clone_pairs"] != 0 || same.Detail["clone_pairs"] != 2 {
		t.Fatalf("single-scope detail = %v", same.Detail)
	}
}

func TestPopulationDigestIgnoresNonEligibleFilesAndTracksEligibleOnes(t *testing.T) {
	base := find(t, Compute(fixture()), MetricCC, "pkg/sub").PopulationDigest

	nonEligible := fixture()
	nonEligible.Candidates[2].ContentDigest = "changed-small" // NLOC 9 file
	nonEligible.Candidates = append(nonEligible.Candidates, file("pkg/sub/extra.py", "x"))
	// extra.py has no lizard row (unmeasured): not in the population either.
	if got := find(t, Compute(nonEligible), MetricCC, "pkg/sub").PopulationDigest; got != base {
		t.Fatal("population digest changed for a non-eligible file")
	}
	// Order of candidates does not matter.
	reordered := fixture()
	reordered.Candidates[0], reordered.Candidates[1] = reordered.Candidates[1], reordered.Candidates[0]
	if got := find(t, Compute(reordered), MetricCC, "pkg/sub").PopulationDigest; got != base {
		t.Fatal("population digest depends on file order")
	}
	eligible := fixture()
	eligible.Candidates[0].ContentDigest = "new-a"
	if got := find(t, Compute(eligible), MetricCC, "pkg/sub").PopulationDigest; got == base {
		t.Fatal("population digest ignored an eligible file's content")
	}
	// The same digest on every metric of the scope.
	for _, signal := range Compute(fixture()) {
		if signal.ScopeAnchor == "pkg/sub" && signal.PopulationDigest != base {
			t.Fatalf("%s carries a different population digest", signal.ID())
		}
	}
}

func TestMissingnessIsNeverZero(t *testing.T) {
	// lizard did not run: everything unmeasured, population = candidates.
	down := fixture()
	down.Facts = Facts{LizardErr: "lizard is not installed"}
	for _, signal := range Compute(down) {
		if signal.Missing != domain.SignalUnmeasured || signal.Value != nil || signal.Numerator != nil || signal.Denominator != nil {
			t.Fatalf("%s %s = %+v, want unmeasured without numbers", signal.ID(), signal.ScopeAnchor, signal)
		}
		if err := signal.Validate(); err != nil {
			t.Fatal(err)
		}
	}

	// jscpd failed: duplication unmeasured, complexity still measured.
	noClones := fixture()
	noClones.Facts.JscpdErr = "jscpd timed out"
	signals := Compute(noClones)
	if dup := find(t, signals, MetricDupPct, "pkg/sub"); dup.Missing != domain.SignalUnmeasured || dup.Value != nil {
		t.Fatalf("dup_pct = %+v", dup)
	}
	if find(t, signals, MetricCC, "pkg/sub").Missing != "" {
		t.Fatal("complexity became unmeasured because jscpd failed")
	}

	// A scope whose production files are all below the NLOC floor is
	// insufficient: not zero and not good. A feature with no production source
	// at all has nothing to measure and records no signal.
	empty := fixture()
	empty.Candidates[2].OwnedFeatureIDs = []string{"S"}
	empty.Features = append(empty.Features, domain.Feature{FeatureID: "S"}, domain.Feature{FeatureID: "docs-only"})
	signals = Compute(empty)
	for _, metric := range []string{MetricCC, MetricComplexBurden, MetricBurdenDensity, MetricDupPct} {
		signal := find(t, signals, metric, FeatureAnchor("S"))
		if signal.Missing != domain.SignalInsufficient || signal.Value != nil {
			t.Fatalf("%s for a scope below the NLOC floor = %+v", metric, signal)
		}
	}
	for _, signal := range signals {
		if signal.ScopeAnchor == FeatureAnchor("docs-only") {
			t.Fatalf("recorded %s for a scope with no production source", signal.ID())
		}
	}

	// CRAP is unmeasured everywhere when the input carries no coverage artifact.
	for _, signal := range signals {
		if signal.MetricID == MetricCrap && (signal.Missing != domain.SignalUnmeasured || signal.Value != nil) {
			t.Fatalf("crap = %+v", signal)
		}
	}
}

func TestEligiblePopulationAndScopeRules(t *testing.T) {
	for path, want := range map[string]bool{
		"fastapi/applications.py": true, "pkg/x.go": true, "web/app.tsx": true, "a/b.mjs": true,
		"tests/test_x.py": false, "pkg/x_test.go": false, "docs_src/a.py": false, "vendor/x/y.go": false,
		"web/types.d.ts": false, "web/x.spec.ts": false, "gen/pb/x.pb.go": false, ".github/s.py": false,
		"node_modules/a/b.js": false, "README.md": false, "pkg/generated/x.go": false, "src/test/a.py": false,
	} {
		if got := IsProductionSource(path); got != want {
			t.Errorf("IsProductionSource(%q) = %v, want %v", path, got, want)
		}
	}
	for path, want := range map[string]string{
		"main.go":                       ".",
		"fastapi/applications.py":       "fastapi",
		"fastapi/dependencies/utils.py": "fastapi/dependencies",
		"fastapi/a/b/c/d.py":            "fastapi/a",
		"src/pkg/sub/deep/x.py":         "src/pkg/sub",
		"src/x.py":                      "src",
	} {
		if got := DirectoryScope(path); got != want {
			t.Errorf("DirectoryScope(%q) = %q, want %q", path, got, want)
		}
	}
	scopes := BuildScopes([]domain.File{file("a/x.py", "1", "F1", "F2"), file("b/y.py", "2"), file("c/z.py", "3", "F1")}, []domain.Feature{{FeatureID: "F3"}})
	got := map[string][]string{}
	for _, scope := range scopes {
		got[scope.Anchor] = scope.Paths
	}
	want := map[string][]string{"feature:F1": {"a/x.py", "c/z.py"}, "feature:F2": {"a/x.py"}, "feature:F3": nil, "b": {"b/y.py"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("scopes = %v, want %v", got, want)
	}
}

func TestComputeIsDeterministic(t *testing.T) {
	first, second := Compute(fixture()), Compute(fixture())
	if !reflect.DeepEqual(first, second) {
		t.Fatal("Compute is not deterministic")
	}
	for i := 1; i < len(first); i++ {
		if first[i-1].ScopeAnchor > first[i].ScopeAnchor {
			t.Fatal("signals are not ordered by scope")
		}
	}
}

func TestParsersReadScannerReports(t *testing.T) {
	lizard := `<?xml version="1.0" ?><cppncss>
	<measure type="Function"><labels><label>Nr.</label><label>NCSS</label><label>CCN</label></labels>
		<item name="Class::run(...) at pkg/a.py:7"><value>1</value><value>16</value><value>12</value></item>
		<item name="other(...) at pkg/b.py:3"><value>2</value><value>4</value><value>2</value></item>
	</measure>
	<measure type="File"><labels><label>Nr.</label><label>NCSS</label><label>CCN</label><label>Functions</label></labels>
		<item name="pkg/a.py"><value>1</value><value>20</value><value>12</value><value>1</value></item>
		<item name="pkg/b.py"><value>2</value><value>8</value><value>2</value><value>1</value></item>
	</measure></cppncss>`
	files, err := ParseLizardXML([]byte(lizard))
	if err != nil {
		t.Fatal(err)
	}
	a := files["pkg/a.py"]
	if a.NLOC != 20 || len(a.Functions) != 1 || a.Functions[0] != (Function{Name: "Class::run", Line: 7, CC: 12, NLOC: 16}) || a.FileCC() != 12 {
		t.Fatalf("pkg/a.py = %+v", a)
	}
	if files["pkg/b.py"].NLOC != 8 {
		t.Fatalf("pkg/b.py = %+v", files["pkg/b.py"])
	}

	// The second pair's secondFile.end (3) precedes its start, as jscpd 4.0.8
	// sometimes reports; the span comes from start and `lines`.
	report := `{"duplicates":[{"format":"python","lines":20,"firstFile":{"name":"/repo/pkg/a.py","start":10,"end":29},"secondFile":{"name":"/repo/pkg/c.py","start":1,"end":20}},
		{"format":"python","lines":5,"firstFile":{"name":"/repo/pkg/a.py","start":40,"end":44},"secondFile":{"name":"/repo/pkg/c.py","start":7,"end":3}}]}`
	clones, err := ParseJscpdReport([]byte(report), "/repo")
	if err != nil || len(clones) != 2 || clones[0].A != (Span{"pkg/a.py", 10, 29}) || clones[0].B != (Span{"pkg/c.py", 1, 20}) || clones[1].B != (Span{"pkg/c.py", 7, 11}) {
		t.Fatalf("clones = %+v, %v", clones, err)
	}
	if _, err := ParseJscpdReport([]byte(`{"duplicates":[{"firstFile":{"name":"/elsewhere/a.py","start":1,"end":2},"secondFile":{"name":"/repo/b.py","start":1,"end":2}}]}`), "/repo"); err == nil {
		t.Fatal("accepted a clone outside the repository")
	}
}

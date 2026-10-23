package measures

import (
	"testing"

	"scout/internal/domain"
)

// coverageFixture instruments the fixture() files (f1 spans lines 1-19, f2
// starts at 20; f3 is the only function of c.py):
//
//	f1: lines 1-4 all missed                → cov 0    → CRAP 12²·1 + 12      = 156
//	f2: lines 20-23, two hit                → cov 0.5  → CRAP 2²·0.125 + 2    = 2.5
//	f3: lines 1-4, three hit                → cov 0.75 → CRAP 3²·(1/64) + 3   = 3.140625
func coverageFixture() *Coverage {
	return &Coverage{
		Digest: "artifact-digest",
		Lines: map[string]map[int]int{
			"pkg/sub/a.py": {1: 0, 2: 0, 3: 0, 4: 0, 20: 1, 21: 3, 22: 0, 23: 0},
			"pkg/sub/c.py": {1: 1, 2: 1, 3: 4, 4: 0},
		},
	}
}

func TestCRAPFormula(t *testing.T) {
	for _, c := range []struct {
		cc   int
		cov  float64
		want float64
	}{
		{12, 0, 156}, {2, 0.5, 2.5}, {3, 0.75, 3.140625}, {6, 0.25, 21.1875}, {10, 1, 10}, {10, 0, 110},
	} {
		if got := CRAP(c.cc, c.cov); got != c.want {
			t.Errorf("CRAP(%d, %v) = %v, want %v", c.cc, c.cov, got, c.want)
		}
	}
}

func TestCrapSignalMatchesHandComputedFormula(t *testing.T) {
	input := fixture()
	input.Coverage = coverageFixture()
	signals := Compute(input)
	for _, signal := range signals {
		if err := signal.Validate(); err != nil {
			t.Fatalf("%s %s invalid: %v", signal.ID(), signal.ScopeAnchor, err)
		}
	}

	dir := find(t, signals, MetricCrap, "pkg/sub")
	if dir.Missing != domain.SignalMeasured || dir.Claim != domain.ClaimDescriptive {
		t.Fatalf("pkg/sub crap = %+v", dir)
	}
	// crap_load = Σ_{CRAP>30} = f1 only; uncovered_cc_mass = (12·1 + 2·0.5) ÷ 14.
	near(t, "crap_load", dir.Value, 156)
	near(t, "uncovered cc", dir.Numerator, 13)
	near(t, "total cc", dir.Denominator, 14)
	near(t, "uncovered_cc_mass", floatDetail(t, dir, "uncovered_cc_mass"), 13.0/14.0)
	if got := dir.Detail["n_crap_over_threshold"]; got != 1 {
		t.Fatalf("n_crap_over_threshold = %v", got)
	}

	// A scope whose functions are all under the threshold has crap_load 0: a
	// measured zero, because the artifact covers it.
	f := find(t, signals, MetricCrap, "feature:F")
	if f.Missing != domain.SignalMeasured {
		t.Fatalf("feature crap = %+v", f)
	}
	near(t, "F crap_load", f.Value, 0)
	near(t, "F uncovered", f.Numerator, 0.75)
	near(t, "F total", f.Denominator, 3)

	// The coverage artifact is part of the tool digest, so the same snapshot
	// measured without coverage is a distinct row.
	bare := find(t, Compute(fixture()), MetricCrap, "pkg/sub")
	if bare.ToolDigest == dir.ToolDigest {
		t.Fatal("coverage and no-coverage rows share a tool digest")
	}
	if dir.PopulationDigest != bare.PopulationDigest {
		t.Fatal("the population digest must not depend on coverage")
	}
}

func floatDetail(t *testing.T, s domain.Signal, key string) *float64 {
	t.Helper()
	value, ok := s.Detail[key].(float64)
	if !ok {
		t.Fatalf("detail %q = %#v", key, s.Detail[key])
	}
	return &value
}

func TestCrapIsUnmeasuredNeverZeroWithoutUsableCoverage(t *testing.T) {
	cases := map[string]*Coverage{
		"absent":     nil,
		"mismatched": {Unusable: "no coverage artifact matches the snapshot revision: ci.lcov: the artifact was produced at aaaa but the snapshot is bbbb"},
		// The artifact matched but instruments none of the scope's files.
		"no data for scope": {Digest: "d", Lines: map[string]map[int]int{"other/file.py": {1: 1}}},
	}
	for name, coverage := range cases {
		input := fixture()
		input.Coverage = coverage
		for _, signal := range Compute(input) {
			if signal.MetricID != MetricCrap {
				continue
			}
			if signal.Missing != domain.SignalUnmeasured || signal.Value != nil || signal.Numerator != nil || signal.MissingReason == "" {
				t.Fatalf("%s: %s crap = %+v", name, signal.ScopeAnchor, signal)
			}
			if err := signal.Validate(); err != nil {
				t.Fatalf("%s: %v", name, err)
			}
		}
	}
}

func TestCrapCountsFunctionsWithoutCoverageDataInsteadOfTreatingThemAsUncovered(t *testing.T) {
	input := fixture()
	input.Coverage = &Coverage{Digest: "d", Lines: map[string]map[int]int{"pkg/sub/a.py": {20: 1, 21: 1}}}
	dir := find(t, Compute(input), MetricCrap, "pkg/sub")
	// f1 (lines 1-19) has no instrumented line; only f2 (CC 2, fully covered) counts.
	near(t, "crap_load", dir.Value, 0)
	near(t, "uncovered", dir.Numerator, 0)
	near(t, "total", dir.Denominator, 2)
	if dir.Detail["n_functions_without_coverage_data"] != 1 {
		t.Fatalf("detail = %v", dir.Detail)
	}
}

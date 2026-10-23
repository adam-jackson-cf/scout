package importgraph

import (
	"testing"

	"scout/internal/domain"
)

func TestSignalsPerScope(t *testing.T) {
	g := build(t, map[string]string{
		"go.mod":    "module m\n",
		"x/x.go":    "package x\nimport _ \"m/y\"\nimport _ \"m/z\"\n",
		"y/y.go":    "package y\nimport _ \"m/x\"\nimport _ \"m/gone\"\n",
		"z/z.go":    "package z\n",
		"lone/l.go": "package lone\n",
		"docs/a.md": "# docs\n",
	})
	scopes := []ScopeInput{
		{Kind: domain.SignalScopeDirectory, Anchor: "x", Paths: []string{"x/x.go"}},
		{Kind: domain.SignalScopeDirectory, Anchor: "y", Paths: []string{"y/y.go", "z/z.go"}},
		{Kind: domain.SignalScopeDirectory, Anchor: "lone", Paths: []string{"lone/l.go"}, PopulationDigest: "pop-lone"},
		{Kind: domain.SignalScopeDirectory, Anchor: "docs", Paths: []string{"docs/a.md"}},
	}
	sigs := Signals(g, scopes, "snap")
	if len(sigs) != 4 {
		t.Fatalf("got %d signals", len(sigs))
	}
	for _, s := range sigs {
		if err := s.Validate(); err != nil {
			t.Fatalf("%s: %v", s.ScopeAnchor, err)
		}
		if s.MetricID != "scout.cycle_mass" || s.Version != 1 || s.Claim != domain.ClaimDescriptive {
			t.Fatalf("identity %+v", s)
		}
	}
	x, y, lone, docs := sigs[0], sigs[1], sigs[2], sigs[3]
	// x owns edges x->y (cycle) and x->z (not): 1 of 2
	if *x.Numerator != 1 || *x.Denominator != 2 || *x.Value != 0.5 {
		t.Fatalf("x: %v/%v=%v", *x.Numerator, *x.Denominator, *x.Value)
	}
	// y owns y->x (cycle): 1 of 1; one unresolved
	if *y.Numerator != 1 || *y.Denominator != 1 || y.Detail["unresolved_imports"] != 1 {
		t.Fatalf("y: %+v", y)
	}
	// the x<->y cycle spans two scopes
	if x.Detail["cross_boundary_cycles"] != 1 || y.Detail["cross_boundary_cycles"] != 1 {
		t.Fatalf("cross: x=%v y=%v", x.Detail, y.Detail)
	}
	if lone.Missing != domain.SignalInsufficient || lone.Value != nil || lone.PopulationDigest != "pop-lone" {
		t.Fatalf("lone must be insufficient, not 0: %+v", lone)
	}
	if docs.Missing != domain.SignalUnmeasured || docs.Value != nil {
		t.Fatalf("docs must be unmeasured: %+v", docs)
	}
}

func TestSignalsDeterministicDigest(t *testing.T) {
	g := build(t, map[string]string{"a.py": "import b\n", "b.py": "import a\n"})
	sc := []ScopeInput{{Kind: domain.SignalScopeRepository, Anchor: ""}}
	a, b := Signals(g, sc, "s"), Signals(g, sc, "s")
	if a[0].PopulationDigest == "" || a[0].PopulationDigest != b[0].PopulationDigest || *a[0].Value != 1 {
		t.Fatalf("%+v", a[0])
	}
}

package decisions

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"scout/internal/domain"
)

const (
	testModel = "typesafe/jev-1.14-20261014"
	testKey   = "sk-or-v1-0123456789abcdefSECRET"
)

func testConfig() Config {
	return Config{Model: testModel, MaxPerRun: 100, MaxCostUSD: 100}
}

func TestVersionKeyIsStableAndChangesOnEveryKeyedField(t *testing.T) {
	base := Question{
		ID: "q.one", Kind: KindChoice, Instructions: "pick", Criteria: map[string]string{"a": "first", "b": "second"},
		StateShape: []string{"state.x", "state.y"}, Order: OptionOrderRule{Rule: OrderPermutation},
	}
	want := base.Version(testModel)
	if !strings.HasPrefix(want, "q.one@") || len(want) != len("q.one@")+64 {
		t.Fatalf("version %q is not question_id@sha256", want)
	}
	assertVersionIgnoresUnkeyedFields(t, base, want)
	assertEveryKeyedFieldChangesVersion(t, base, want)
}

func assertVersionIgnoresUnkeyedFields(t *testing.T, base Question, want string) {
	t.Helper()
	if base.Version(testModel) != want {
		t.Fatal("version is not stable")
	}
	reordered := base
	reordered.StateShape = []string{"state.y", "state.x"}
	if reordered.Version(testModel) != want {
		t.Fatal("state shape is a set; order must not change the version")
	}
	notKeyed := base
	notKeyed.Positive = "a"
	if notKeyed.Version(testModel) != want {
		t.Fatal("Positive is a code-side reading and must not change the version")
	}
	if base.RealisedOrder("unit-1")[0] == "" || len(base.RealisedOrder("unit-1")) != 2 {
		t.Fatal("realised order must list every option")
	}
}

func assertEveryKeyedFieldChangesVersion(t *testing.T, base Question, want string) {
	t.Helper()
	mutations := map[string]func(Question) (Question, string){
		"instructions": func(q Question) (Question, string) { q.Instructions = "pick carefully"; return q, testModel },
		"criteria": func(q Question) (Question, string) {
			q.Criteria = map[string]string{"a": "first", "b": "other"}
			return q, testModel
		},
		"state shape": func(q Question) (Question, string) {
			q.StateShape = []string{"state.x", "state.y", "state.z"}
			return q, testModel
		},
		"option order": func(q Question) (Question, string) {
			q.Order = OptionOrderRule{Rule: OrderFixed, Options: []string{"a", "b"}}
			return q, testModel
		},
		"resolved model": func(q Question) (Question, string) { return q, "typesafe/jev-1.15-20270101" },
	}
	seen := map[string]string{want: "base"}
	for name, mutate := range mutations {
		q, model := mutate(base)
		got := q.Version(model)
		if prev, dup := seen[got]; dup {
			t.Fatalf("%s produced the same version as %s", name, prev)
		}
		seen[got] = name
	}
	other := base
	other.ID = "q.two"
	if !strings.HasPrefix(other.Version(testModel), "q.two@") || other.Version(testModel)[6:] != want[6:] {
		t.Fatal("the id prefixes the key but is not part of the hashed payload")
	}
}

func TestEligibilityReasons(t *testing.T) {
	ok := Eligibility{Enabled: true, HasKey: true, EgressAllowed: true}
	if reason, allowed := ok.Skip(); !allowed || reason != "" {
		t.Fatal("fully consented run must be allowed")
	}
	cases := []struct {
		mutate func(*Eligibility)
		reason string
	}{
		{func(e *Eligibility) { e.Enabled = false }, "decisions not enabled"},
		{func(e *Eligibility) { e.Fork = true }, "fork"},
		{func(e *Eligibility) { e.Fork = true; e.Enabled = false }, "fork"},
		{func(e *Eligibility) { e.EgressAllowed = false }, "egress not allowed"},
		{func(e *Eligibility) { e.HasKey = false }, "no jev key"},
	}
	for _, tc := range cases {
		e := ok
		tc.mutate(&e)
		if reason, allowed := e.Skip(); allowed || reason != tc.reason {
			t.Fatalf("got %q allowed=%v, want %q", reason, allowed, tc.reason)
		}
	}
	if h := SkippedHealth("fork"); h.Status != domain.LaneSkipped || h.Reason != "fork" || h.Required {
		t.Fatalf("health %+v", h)
	}
	assertConfigRejectsAliasesAndBadCaps(t)
}

func assertConfigRejectsAliasesAndBadCaps(t *testing.T) {
	t.Helper()
	for _, alias := range []string{"", "jev-latest", "typesafe/jev-preview"} {
		c := testConfig()
		c.Model = alias
		if c.Validate() == nil {
			t.Fatalf("model %q must be rejected", alias)
		}
	}
	for name, mutate := range map[string]func(*Config){
		"zero call cap":     func(c *Config) { c.MaxPerRun = 0 },
		"negative call cap": func(c *Config) { c.MaxPerRun = -1 },
		"zero cost cap":     func(c *Config) { c.MaxCostUSD = 0 },
		"negative cost cap": func(c *Config) { c.MaxCostUSD = -0.1 },
		"NaN cost cap":      func(c *Config) { c.MaxCostUSD = math.NaN() },
	} {
		c := testConfig()
		mutate(&c)
		if c.Validate() == nil {
			t.Fatalf("%s must be rejected", name)
		}
	}
	if err := (testConfig()).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRankFormulaVersionAndOrdinalReranking(t *testing.T) {
	formula := LaunchFormula()
	version := formula.Version(testModel)
	if version != formula.Version(testModel) || version == formula.Version("typesafe/jev-1.15-20270101") {
		t.Fatal("rank-formula version must be stable and follow the constituent question versions")
	}
	heavier := LaunchFormula()
	heavier.Weights = map[string]float64{TriageRealDefect.ID: 2}
	if heavier.Version(testModel) == version {
		t.Fatal("changing a weight must change the formula version")
	}
	items := []Item{
		{IssueID: "a", Baseline: 4, Answers: map[string]float64{TriageRealDefect.ID: 0.2}},
		{IssueID: "b", Baseline: 1, Answers: map[string]float64{TriageRealDefect.ID: 0.9}},
		{IssueID: "c", Baseline: 9},
		{IssueID: "d", Baseline: 3},
	}
	ranked := Rank(formula, items)
	order := ""
	for _, r := range ranked {
		order += r.IssueID
	}
	if order != "bacd" || ranked[0].Position != 1 || ranked[3].Position != 4 {
		t.Fatalf("order %s: annotated by P, then unannotated by baseline", order)
	}
	// Re-ranking stored answers under a new weight needs no model call: Rank is pure.
	if again := Rank(heavier, items); again[0].Score != 1.8 {
		t.Fatalf("score %v", again[0].Score)
	}
}

type memoryAudit struct {
	rows map[string]domain.AuditDraw
}

func (m *memoryAudit) AuditDraw(_ context.Context, issue, formula, sampler string) (domain.AuditDraw, bool, error) {
	d, ok := m.rows[issue+formula+sampler]
	return d, ok, nil
}

func (m *memoryAudit) RecordAuditDraw(_ context.Context, d domain.AuditDraw) (domain.AuditDraw, error) {
	key := d.IssueID + d.FormulaVersion + d.SamplerVersion
	if existing, ok := m.rows[key]; ok {
		return existing, nil
	}
	m.rows[key] = d
	return d, nil
}

func TestAuditDrawStoresInclusionProbabilityAndSamplerVersion(t *testing.T) {
	var units []AuditUnit
	for i := 1; i <= 20; i++ {
		units = append(units, AuditUnit{IssueID: fmt.Sprintf("i%02d", i), Position: i, Annotated: i <= 16, Critical: i == 18 || i == 3})
	}
	assigned := map[string]Assignment{}
	for _, a := range Assign(units, ProvisionalFraction) {
		assigned[a.IssueID] = a
	}
	assertAuditStrata(t, assigned)

	store := &memoryAudit{rows: map[string]domain.AuditDraw{}}
	assertFirstDrawsRecordProbability(t, store, units, assigned)
	assertStoredDrawsAreKept(t, store, units, assigned)
}

func assertAuditStrata(t *testing.T, assigned map[string]Assignment) {
	t.Helper()
	if assigned["i18"].Stratum != StratumCertainty || assigned["i18"].Probability != 1 {
		t.Fatalf("critical finding below the floor is certainty: %+v", assigned["i18"])
	}
	if assigned["i03"].Stratum != StratumTop {
		t.Fatalf("critical finding above the floor stays in its half: %+v", assigned["i03"])
	}
	if assigned["i12"].Stratum != StratumBottom || assigned["i19"].Stratum != StratumUnannotated || assigned["i19"].Probability <= 0 || assigned["i19"].Probability > 1 {
		t.Fatalf("strata %+v %+v", assigned["i12"], assigned["i19"])
	}
	assertExpectedAuditsMatchFraction(t, assigned)
}

func assertExpectedAuditsMatchFraction(t *testing.T, assigned map[string]Assignment) {
	t.Helper()
	// Expected audits are the fraction of the run, split equally over the strata.
	expected := 0.0
	for _, a := range assigned {
		if a.Stratum != StratumCertainty {
			expected += a.Probability
		}
	}
	if expected < 0.9 || expected > 1.1 {
		t.Fatalf("expected audits outside the certainty stratum = %v, want ~1 (5%% of 20)", expected)
	}
}

func assertFirstDrawsRecordProbability(t *testing.T, store AuditStore, units []AuditUnit, assigned map[string]Assignment) {
	t.Helper()
	calls := 0
	rand := func() float64 { calls++; return 0.0 } // selects every unit with p > 0
	draws, err := DrawAudit(context.Background(), store, AuditRequest{RunID: "run-1", FormulaVersion: "formula@1", Units: units, Fraction: ProvisionalFraction, Rand: rand, Now: time.Unix(1, 0)})
	if err != nil || len(draws) != 20 || calls != 20 {
		t.Fatalf("draws %d calls %d err %v", len(draws), calls, err)
	}
	for _, d := range draws {
		a := assigned[d.IssueID]
		if d.SamplerVersion != SamplerVersion || d.InclusionProbability != a.Probability || d.Stratum != a.Stratum || !d.Drawn || d.FormulaVersion != "formula@1" {
			t.Fatalf("stored draw %+v does not record the realised probability and sampler version", d)
		}
	}
}

func assertStoredDrawsAreKept(t *testing.T, store AuditStore, units []AuditUnit, assigned map[string]Assignment) {
	t.Helper()
	// A unit is drawn once, at its first eligible run; later runs keep the stored draw.
	again, err := DrawAudit(context.Background(), store, AuditRequest{RunID: "run-2", FormulaVersion: "formula@1", Units: units, Fraction: 0.5, Rand: func() float64 { t.Fatal("redrawn"); return 1 }, Now: time.Unix(2, 0)})
	if err != nil || again[0].RunID != "run-1" || again[0].InclusionProbability != assigned[again[0].IssueID].Probability {
		t.Fatalf("redraw %+v err %v", again[0], err)
	}
	// A new formula version is a new sampling unit.
	third, err := DrawAudit(context.Background(), store, AuditRequest{RunID: "run-3", FormulaVersion: "formula@2", Units: units[:2], Fraction: ProvisionalFraction, Rand: func() float64 { return 0.999 }, Now: time.Unix(3, 0)})
	if err != nil || third[0].RunID != "run-3" || third[0].Drawn && third[0].InclusionProbability < 1 {
		t.Fatalf("new formula draw %+v err %v", third[0], err)
	}
}

// issues must not import decisions (an annotation can only change a rank), and
// no Jev-quality metric is computed anywhere in the adapter.
func TestAdapterBoundariesAndNoQualityMetrics(t *testing.T) {
	fset := token.NewFileSet()
	check := func(dir string, forbidden string) {
		files, _ := filepath.Glob(filepath.Join(dir, "*.go"))
		for _, file := range files {
			parsed, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatal(err)
			}
			for _, imp := range parsed.Imports {
				if strings.Contains(imp.Path.Value, forbidden) {
					t.Fatalf("%s imports %s", file, imp.Path.Value)
				}
			}
		}
	}
	check("../issues", "scout/internal/decisions")
	check(".", "scout/internal/issues")

	metric := regexp.MustCompile(`(?i)precision|recall|accuracy|auroc|calibrat|brier|f1score|agreement|flip`)
	files, _ := filepath.Glob("*.go")
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(parsed, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && metric.MatchString(id.Name) {
				t.Errorf("%s declares or uses %q: no Jev-quality metric may be computed before X1", file, id.Name)
			}
			return true
		})
	}
}

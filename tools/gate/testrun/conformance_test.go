//go:build boundary

package testrun

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"scout/tools/gate/check"
	"scout/tools/gate/proof"
)

const runSettings = `schema = 1

[layout]
exclude = []
fixtures_dir = "testdata"

[layout.test_tiers]
boundary_tag = "boundary"
deferred_tag = "deferred"
package_timeout = "60s"
`

func fixtureRun(t *testing.T, fixture string) *check.Run {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "proofs.toml")
	if err := os.WriteFile(path, []byte(runSettings), 0o600); err != nil {
		t.Fatal(err)
	}
	pf, err := proof.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return &check.Run{Root: root, Mode: check.CI, Paths: []string{filepath.Join("tools", "gate", "testdata", "conformance", fixture)}, Explicit: true, Proofs: pf, TempDir: t.TempDir()}
}

func TestRunReportConformanceCasesProduceTheirVerdicts(t *testing.T) {
	for id, want := range map[string]string{
		"t1-network":               "TEST.t1-isolation",
		"t1-process":               "TEST.t1-isolation",
		"t1-teardown-process":      "TEST.t1-isolation",
		"t1-swallowed-violation":   "TEST.t1-isolation",
		"t1-clean":                 "",
		"user-data-write":          "TEST.user-data-guard",
		"user-data-temporary-home": "",
	} {
		t.Run(id, func(t *testing.T) {
			findings, err := runTests(fixtureRun(t, id))
			if err != nil {
				t.Fatal(err)
			}
			got := rules(findings)
			if want == "" && len(findings) > 0 || want != "" && !got[want] {
				t.Fatalf("%s: findings %v, want rule %q", id, findings, want)
			}
		})
	}
}

func TestRuntimeSkipKeepsItsOutcomeAndReasonInTheReport(t *testing.T) {
	r := fixtureRun(t, "runtime-skip-recorded")
	findings, err := runTests(r)
	if err != nil {
		t.Fatal(err)
	}
	if !rules(findings)["TEST.no-evasion"] {
		t.Fatalf("skip must be a finding: %v", findings)
	}
	rep, err := ReadReport(filepath.Join(r.TempDir, "report-t1.xml"))
	if err != nil {
		t.Fatal(err)
	}
	c := rep.Suites[0].Cases[0]
	if c.Outcome() != "skip" || c.Reason() != "waiting on vendor fix" {
		t.Fatalf("record = %s %q", c.Outcome(), c.Reason())
	}
}

func TestEmptyTierCollectionIsAFinding(t *testing.T) {
	findings, err := runTiers(fixtureRun(t, "empty-collection"))
	if err != nil {
		t.Fatal(err)
	}
	if !rules(findings)["TEST.failure-propagation"] {
		t.Fatalf("empty collection must fail: %v", findings)
	}
}

func TestTierListingSelectsExactlyItsTests(t *testing.T) {
	r := fixtureRun(t, "tier-selection")
	findings, err := runTiers(r)
	if err != nil || len(findings) > 0 {
		t.Fatalf("tiers: %v %v", findings, err)
	}
	s := sessionFor(r)
	got := map[string][]string{}
	for tier, keys := range s.list.byTier {
		for key := range keys {
			got[tier] = append(got[tier], s.list.byKey[key].ID())
		}
		sort.Strings(got[tier])
	}
	base := "tools/gate/testdata/conformance/tier-selection/engine/"
	if len(got[TierT1]) != 1 || got[TierT1][0] != base+"engine_test.go::TestFastTotal" {
		t.Fatalf("t1 = %v", got[TierT1])
	}
	if len(got[TierT2]) != 1 || got[TierT2][0] != base+"large_input_test.go::TestTotalOfLargeInputStaysExact" {
		t.Fatalf("t2 = %v", got[TierT2])
	}
}

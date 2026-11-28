package testsource

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"scout/tools/gate/check"
	"scout/tools/gate/proof"
)

const conformanceSettings = `schema = 1

[layout]
exclude = []
support_dir = "support"
fixtures_dir = "testdata"

[structure]
adapter_interfaces = ["canary/gateway.Gateway"]

[layout.test_tiers]
boundary_tag = "boundary"
deferred_tag = "deferred"
package_timeout = "60s"
`

const approvedSkip = `
[[exception]]
rule = "TEST.no-evasion"
path = "engine/engine_test.go"
reason = "conformance case approved-static-skip"
approved_by = "conformance fixture"
expires = "2999-12-31"
`

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func conformanceRun(t *testing.T, fixture, extra string) *check.Run {
	t.Helper()
	path := filepath.Join(t.TempDir(), "proofs.toml")
	if err := os.WriteFile(path, []byte(conformanceSettings+extra), 0o600); err != nil {
		t.Fatal(err)
	}
	pf, err := proof.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return &check.Run{Root: repoRoot(t), Mode: check.CI, Paths: []string{filepath.Join("tools", "gate", "testdata", "conformance", fixture)}, Explicit: true, Proofs: pf, TempDir: t.TempDir()}
}

type conformanceCase struct {
	id      string
	run     func(*check.Run) ([]check.Finding, error)
	rule    string
	finding bool
	extra   string
}

var sourceCases = []conformanceCase{
	{"self-mocking-dotted-target", runSelfMocking, "TEST.no-self-mocking", true, ""},
	{"self-mocking-imported-object", runSelfMocking, "TEST.no-self-mocking", true, ""},
	{"self-mocking-assignment", runSelfMocking, "TEST.no-self-mocking", true, ""},
	{"self-mocking-spec", runSelfMocking, "TEST.no-self-mocking", true, ""},
	{"boundary-fake-allowed", runSelfMocking, "TEST.no-self-mocking", false, ""},
	{"unapproved-static-skip", runEvasion, "TEST.no-evasion", true, ""},
	{"approved-static-skip", runEvasion, "TEST.no-evasion", false, approvedSkip},
	{"production-test-flag", runTestHooks, "TEST.no-test-hooks", true, ""},
	{"production-test-hook-parameter", runTestHooks, "TEST.no-test-hooks", true, ""},
	{"production-clean", runTestHooks, "TEST.no-test-hooks", false, ""},
	{"behaviour-name-present", runGrowth, "TEST.growth.justified", false, ""},
	{"behaviour-name-missing", runGrowth, "TEST.growth.justified", true, ""},
	{"behaviour-name-placeholder", runGrowth, "TEST.growth.justified", true, ""},
}

var runReportCases = []string{"t1-network", "t1-process", "t1-teardown-process", "t1-swallowed-violation", "t1-clean", "user-data-write", "user-data-temporary-home", "runtime-skip-recorded", "empty-collection", "tier-selection"}

var blockedCases = map[string]string{
	"retry-marker":              "Go testing has no rerun-until-green marker",
	"expected-failure-marker":   "Go testing has no expected-failure marker",
	"expected-failure-recorded": "go test -json has no xfail outcome",
	"property-without-seed":     "TEST.property-seed awaits the owner's not-applicable decision",
	"property-with-seed":        "TEST.property-seed awaits the owner's not-applicable decision",
}

var allConformanceCases = []string{
	"self-mocking-dotted-target", "self-mocking-imported-object", "self-mocking-assignment", "self-mocking-spec",
	"boundary-fake-allowed", "unapproved-static-skip", "approved-static-skip", "retry-marker", "expected-failure-marker",
	"property-without-seed", "property-with-seed", "production-test-flag", "production-test-hook-parameter",
	"production-clean", "t1-network", "t1-process", "t1-teardown-process", "t1-swallowed-violation", "t1-clean",
	"user-data-write", "user-data-temporary-home", "runtime-skip-recorded", "expected-failure-recorded",
	"behaviour-name-present", "behaviour-name-missing", "behaviour-name-placeholder", "empty-collection", "tier-selection",
}

func TestSourceConformanceCasesProduceTheirVerdicts(t *testing.T) {
	for _, tc := range sourceCases {
		t.Run(tc.id, func(t *testing.T) {
			findings, err := tc.run(conformanceRun(t, tc.id, tc.extra))
			if err != nil {
				t.Fatal(err)
			}
			got := false
			for _, f := range findings {
				got = got || f.Rule == tc.rule
			}
			if got != tc.finding {
				t.Fatalf("%s: finding = %v, want %v: %v", tc.id, got, tc.finding, findings)
			}
		})
	}
}

func TestEveryConformanceCaseIsCoveredOrNamedAsABlocker(t *testing.T) {
	covered := map[string]bool{}
	for _, tc := range sourceCases {
		covered[tc.id] = true
	}
	for _, id := range runReportCases {
		covered[id] = true
	}
	var missing []string
	for _, id := range allConformanceCases {
		if !covered[id] && blockedCases[id] == "" {
			missing = append(missing, id)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("conformance cases with neither a fixture nor a named blocker: %s", strings.Join(missing, ", "))
	}
}

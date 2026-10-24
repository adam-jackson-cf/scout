package analyzers

import (
	"testing"

	"scout/internal/domain"
)

func TestSemgrepJSONAccountsForScannedCleanAndFindingInputs(t *testing.T) {
	spec := specForTest(t, "security.semgrep.python-unsafe-yaml-load")
	raw := []byte(`{"version":"1.100.0","results":[{"check_id":"private.rules.scout.security.python.yaml-unsafe-load","path":"src/unsafe.py","start":{"line":4},"end":{"line":4},"extra":{"message":"unsafe YAML load","severity":"ERROR"}}],"errors":[],"paths":{"scanned":["src/unsafe.py","src/safe.py"]}}`)
	parsed, err := parseOutput(spec, raw, "/tmp/fixture")
	if err != nil {
		t.Fatal(err)
	}
	if !parsed.success || len(parsed.processed) != 2 || len(parsed.findings) != 1 {
		t.Fatalf("expected complete per-input accounting, got %+v", parsed)
	}
	if parsed.findings[0].ruleID != "scout.security.python.yaml-unsafe-load" || parsed.findings[0].severity != domain.SeverityHigh {
		t.Fatalf("unexpected normalized finding: %+v", parsed.findings[0])
	}

	incomplete := []byte(`{"version":"1.100.0","results":[],"errors":[{"type":"ParseError"}],"paths":{"scanned":["src/unsafe.py"]}}`)
	withError, err := parseOutput(spec, incomplete, "/tmp/fixture")
	if err != nil {
		t.Fatal(err)
	}
	if withError.success {
		t.Fatal("scanner-reported error must not establish complete coverage")
	}

	omitted := []byte(`{"version":"1.100.0","results":[],"errors":[],"paths":{"scanned":["src/unsafe.py"]}}`)
	partial, err := parseOutput(spec, omitted, "/tmp/fixture")
	if err != nil {
		t.Fatal(err)
	}
	inputs := map[string]sourceInput{"src/unsafe.py": {}, "src/safe.py": {}}
	status, _, _ := (&Runner{spec: spec}).resultStatus(nil, partial, inputs, nil, false)
	if status != domain.AdapterPartial {
		t.Fatalf("omitted declared input must be partial, got %s", status)
	}
}

func TestWorkflowRuleDoesNotClaimUnrelatedYAMLCoverage(t *testing.T) {
	spec := specForTest(t, "security.semgrep.actions-mutable-ref")
	if err := spec.Validate(); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]bool{
		".github/workflows/build.yml":  true,
		".github/workflows/build.yaml": true,
		".github/actions/build.yml":    false,
		"config/build.yml":             false,
		".github/workflows/build.json": false,
	} {
		file := domain.File{Path: path, Language: "yaml", Eligible: true}
		if got := spec.Applies(file); got != want {
			t.Errorf("workflow rule eligibility for %q: got %t, want %t", path, got, want)
		}
	}
}

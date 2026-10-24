package analyzers

import "testing"

func sarifFixture(guid string) []byte {
	return []byte(`{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"gitleaks"}},"invocations":[{"executionSuccessful":true}],"results":[{"ruleId":"generic-api-key","level":"warning","guid":` + guid + `,"message":{"text":"secret"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"a.py"},"region":{"startLine":3}}}]}]}]}`)
}

func TestSARIFResultIDsAreSafeOrTheOutputIsRejected(t *testing.T) {
	spec := specForTest(t, "security.gitleaks.default-credentials")
	parsed, err := parseSARIF(spec, sarifFixture(`"finding-1"`), "/repo")
	if err != nil || len(parsed.findings) != 1 || parsed.findings[0].id != "finding-1" {
		t.Fatalf("safe id: %+v, %v", parsed.findings, err)
	}
	if _, err := parseSARIF(spec, sarifFixture(`"../../etc/passwd"`), "/repo"); err == nil {
		t.Fatal("an unsafe SARIF result id was silently replaced instead of rejected")
	}
}

func TestSARIFOutputIsOnlyCompleteWhenTheSpecAttestsItsInputs(t *testing.T) {
	attested := specForTest(t, "security.gitleaks.default-credentials")
	parsed, err := parseSARIF(attested, []byte(`{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"gitleaks"}},"results":[]}]}`), "/repo")
	if err != nil || !parsed.success {
		t.Fatalf("attested spec: %+v, %v", parsed, err)
	}
	unattested := attested
	unattested.Attestation = AttestOutput
	parsed, err = parseSARIF(unattested, []byte(`{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"gitleaks"}},"results":[]}]}`), "/repo")
	if err != nil || parsed.success {
		t.Fatalf("unattested SARIF without processed inputs established coverage: %+v, %v", parsed, err)
	}
}

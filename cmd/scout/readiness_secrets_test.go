//go:build boundary

package main

import (
	"encoding/json"
	"strings"
	"testing"
)

const secretsFullVersion = "gitleaks-default-rules/8.30.1@sha256:e163e53b9e7e8a8511e77271e2b323ed057759542a6d988258afe3a1fa329caf"

func assertSecretsActivationDiagnostic(t *testing.T, version string) {
	t.Helper()
	const fullVersion = secretsFullVersion
	const checkID = "security.secrets.github-pat"
	root := newInitRepo(t)
	writeRepoFile(t, root, ".scout/policy.yaml", `profile: guardian
blocking:
  checks:
    - id: `+checkID+`
      version: "`+version+`"
      activated_at: 2026-11-08
      evidence: {would_block: 3, warranted: 3, window_days: 14, pr_runs: 9}
`)
	gitIn(t, root, "add", "-f", ".scout/policy.yaml")
	gitIn(t, root, "commit", "-qm", "policy")
	raw, _ := captureStdout(t, func() int {
		return readiness(boundedTestContext(t), []string{"--root", root, "--family", "architecture", "--json"})
	})
	var reply struct {
		Result struct {
			Capabilities []capability `json:"capabilities"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(raw), &reply); err != nil {
		t.Fatalf("decode readiness: %v: %s", err, raw)
	}
	if strings.Contains(raw, "private-fixture") {
		t.Fatal("readiness exposed a credential-shaped configured version")
	}
	found := false
	for _, check := range reply.Result.Capabilities {
		if check.Name != checkID {
			continue
		}
		found = true
		if !strings.Contains(check.Reason, fullVersion) || !strings.Contains(check.Reason, "shadow") {
			t.Errorf("missing full-version shadow diagnostic: %+v", check)
		}
		if check.Required || !check.Ready {
			t.Errorf("activation diagnostic must remain advisory: %+v", check)
		}
	}
	if want := version != fullVersion; found != want {
		t.Fatalf("diagnostic present = %t, want %t: %s", found, want, raw)
	}
}

func TestReadinessSecretsActivationDiagnostic(t *testing.T) {
	for _, version := range []string{"8.30.1", secretsFullVersion, "password: private-fixture"} {
		t.Run(version, func(t *testing.T) { assertSecretsActivationDiagnostic(t, version) })
	}
}

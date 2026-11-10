package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"scout/internal/domain"
)

func TestHasMatchingObservationIntersectsFamilyAndRun(t *testing.T) {
	observations := []domain.Observation{
		{RunID: "run-1", Family: domain.FamilySecurity},
		{RunID: "run-2", Family: domain.FamilyArchitecture},
	}

	if !hasMatchingObservation(observations, "", domain.FamilyArchitecture) {
		t.Fatal("family-only query did not match its issue observation")
	}
	if hasMatchingObservation(observations, "run-1", domain.FamilyArchitecture) {
		t.Fatal("family and run query matched separate observations")
	}
	if !hasMatchingObservation(observations, "run-2", domain.FamilyArchitecture) {
		t.Fatal("family and run query did not match the same observation")
	}
}
func TestRequireAuthorizationRejectsUnauditedMutation(t *testing.T) {
	if _, _, err := requireAuthorization(false, "reviewer", "triage"); err == nil {
		t.Fatal("unauthorized mutation was accepted")
	}
}

func TestArtifactInRunRequiresARecordedArtifact(t *testing.T) {
	run := domain.Run{ArtifactRefs: []domain.SafeArtifactRef{{ArtifactID: "artifact-1"}}}
	if _, err := artifactInRun(run, "artifact-1"); err != nil {
		t.Fatalf("recorded artifact was rejected: %v", err)
	}
	if _, err := artifactInRun(run, "artifact-2"); err == nil {
		t.Fatal("unrecorded artifact was accepted")
	}
}

func TestUnknownCommandIsUnavailable(t *testing.T) {
	if code := run(context.Background(), []string{"no-such-command"}); code != exitUnavailable {
		t.Fatalf("unknown command exit = %d, want %d", code, exitUnavailable)
	}
}

func TestJSONOutcomeExitAgreement(t *testing.T) {
	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create output pipe: %v", err)
	}
	os.Stdout = writer
	defer func() {
		os.Stdout = originalStdout
		_ = reader.Close()
	}()

	code := emit(&commonFlags{json: true}, "gated", exitGate, nil, nil)
	_ = writer.Close()
	encoded, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read JSON outcome: %v", err)
	}
	var report outcome
	if err := json.Unmarshal(encoded, &report); err != nil {
		t.Fatalf("decode JSON outcome: %v", err)
	}
	if code != 1 || report.ExitCode != code || report.Status != "gated" {
		t.Fatalf("gated outcome = %#v with code %d, want gated JSON and exit 1", report, code)
	}
}

func TestReadinessSecretsActivationDiagnostic(t *testing.T) {
	const fullVersion = "gitleaks-default-rules/8.30.1@sha256:e163e53b9e7e8a8511e77271e2b323ed057759542a6d988258afe3a1fa329caf"
	const checkID = "security.secrets.github-pat"
	for _, version := range []string{"8.30.1", fullVersion, "password: private-fixture"} {
		t.Run(version, func(t *testing.T) {
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
		})
	}
}

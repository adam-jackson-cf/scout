package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
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

func TestApprovalCommandsAreRemoved(t *testing.T) {
	for _, command := range []string{"approve", "verify"} {
		if code := run(context.Background(), []string{command}); code != exitUnavailable {
			t.Fatalf("scout %s exit = %d, want unknown-command exit %d", command, code, exitUnavailable)
		}
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

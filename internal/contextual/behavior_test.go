package contextual

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"scout/internal/domain"
)

func featureWith(paths ...string) domain.Feature {
	feature := testReviewRequest("").Feature
	feature.OwnedFiles = nil
	for _, path := range paths {
		feature.OwnedFiles = append(feature.OwnedFiles, domain.FeatureFileRef{Path: path, Reason: "service"})
	}
	return feature
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func omitted(manifest Manifest, path string) bool {
	for _, omission := range manifest.Omissions {
		if omission.Path == path {
			return true
		}
	}
	return false
}

func TestSourceThroughASymlinkOutsideTheRootIsNeverReadOrSent(t *testing.T) {
	outside := filepath.Join(t.TempDir(), "outside.txt")
	writeFile(t, outside, "outside-only-content\n")
	root := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "link.go")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := os.Symlink(filepath.Dir(outside), filepath.Join(root, "dir")); err != nil {
		t.Fatal(err)
	}
	manifest, sources := readFeatureSources(root, featureWith("link.go", "dir/outside.txt"), 1<<20, 1<<20)
	if len(sources) != 0 || len(manifest.Inspected) != 0 || !omitted(manifest, "link.go") || !omitted(manifest, "dir/outside.txt") {
		t.Fatalf("manifest %+v sources %+v", manifest, sources)
	}

	fake := installFakeCLI(t, "claude", "exit 1")
	request := testReviewRequest(root)
	request.Feature = featureWith("link.go")
	result, err := newTestClient(t, "claude").Review(context.Background(), request)
	if err == nil || fake.invoked() || result.Coverage.Status != domain.CoverageIncomplete {
		t.Fatalf("a feature with no readable in-root source must not reach the CLI: err=%v invoked=%t coverage=%+v", err, fake.invoked(), result.Coverage)
	}
}

func TestScoutStateIsExcludedFromSourceAndRecordedAsAnOmission(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".scout", "policy.yaml"), "state-only-content\n")
	writeFile(t, filepath.Join(root, "main.go"), "package main\n")
	manifest, sources := readFeatureSources(root, featureWith(".scout/policy.yaml", ".scout", "main.go"), 1<<20, 1<<20)
	if len(manifest.Inspected) != 1 || manifest.Inspected[0].Range.Path != "main.go" || !omitted(manifest, ".scout/policy.yaml") || !omitted(manifest, ".scout") {
		t.Fatalf("manifest %+v", manifest)
	}
	for _, source := range sources {
		if strings.Contains(source.Text, "state-only-content") {
			t.Fatal("Scout state reached the prompt sources")
		}
	}

	fake := installFakeCLI(t, "claude", "exit 1")
	request := testReviewRequest(root)
	request.Feature = featureWith(".scout/policy.yaml", "main.go")
	_, _ = newTestClient(t, "claude").Review(context.Background(), request)
	if stdin := fake.read(t, fake.stdin); strings.Contains(stdin, "state-only-content") {
		t.Fatal("Scout state reached the CLI")
	}
}

func TestOversizedFilesAndTheTotalContextLimitAreTruncatedAndDisclosed(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "big.go"), "line one\nline two\nline three\nline four\n")
	writeFile(t, filepath.Join(root, "second.go"), "second\n")

	manifest, sources := readFeatureSources(root, featureWith("big.go"), 1<<20, 20)
	if !manifest.Truncated || !omitted(manifest, "big.go") || len(sources) != 1 {
		t.Fatalf("per-file limit: manifest %+v", manifest)
	}
	if text := sources[0].Text; len(text) > 20 || !strings.HasSuffix(text, "\n") || strings.Contains(text, "three") {
		t.Fatalf("excerpt must be cut on a line boundary within the limit: %q", text)
	}
	digest := manifest.Inspected[0].SourceDigest
	if digest != sourceDigest([]byte("line one\nline two\nline three\nline four\n")) {
		t.Fatal("the manifest digest must cover the full file, not the excerpt")
	}

	manifest, sources = readFeatureSources(root, featureWith("second.go", "big.go"), 7, 1<<20)
	if !manifest.Truncated || len(sources) != 1 || sources[0].Path != "second.go" || !omitted(manifest, "big.go") {
		t.Fatalf("total limit: manifest %+v sources %d", manifest, len(sources))
	}
}

func TestEveryKindOfSecretIsRedactedBeforeEgress(t *testing.T) {
	secrets := []string{
		"-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAprivatekeymaterial\n-----END RSA PRIVATE KEY-----",
		"curl -H 'X-Auth: Bearer abcdefghijklmnop1234567890'",
		"password = hunter2hunter2",
		testSecret,
	}
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "main.go"), strings.Join(secrets, "\n")+"\n")
	fake := installFakeCLI(t, "claude", "exit 1")
	request := testReviewRequest(root)
	request.Feature.OwnedFiles[0].Reason = "owner token=" + testSecret
	_, _ = newTestClient(t, "claude").Review(context.Background(), request)
	stdin := fake.read(t, fake.stdin)
	for _, leaked := range []string{"privatekeymaterial", "abcdefghijklmnop" + "1234567890", "hunter2hunter2", testSecret} {
		if strings.Contains(stdin, leaked) {
			t.Fatalf("secret %q reached the model CLI", leaked)
		}
	}
}

func TestSeverityMatrixMapsEveryFamily(t *testing.T) {
	cases := []struct {
		family domain.Family
		raw    string
		want   domain.Severity
	}{
		{domain.FamilySecurity, `{"impact":1,"likelihood":1}`, domain.SeverityInformational},
		{domain.FamilySecurity, `{"impact":1,"likelihood":5}`, domain.SeverityLow},
		{domain.FamilySecurity, `{"impact":2,"likelihood":3}`, domain.SeverityMedium},
		{domain.FamilySecurity, `{"impact":3,"likelihood":4}`, domain.SeverityHigh},
		{domain.FamilySecurity, `{"impact":4,"likelihood":5}`, domain.SeverityCritical},
		{domain.FamilyArchitecture, `{"changeRisk":"blocking","boundary":"systemic_or_critical_boundary"}`, domain.SeverityCritical},
		{domain.FamilyArchitecture, `{"changeRisk":"major","boundary":"component"}`, domain.SeverityHigh},
		{domain.FamilyArchitecture, `{"changeRisk":"minor","boundary":"local"}`, domain.SeverityLow},
		{domain.FamilyArchitecture, `{"changeRisk":"blocking","boundary":"none"}`, domain.SeverityInformational},
		{domain.FamilyQuality, `{"changeRisk":"blocking","scope":"systemic"}`, domain.SeverityCritical},
		{domain.FamilyQuality, `{"changeRisk":"meaningful","scope":"component"}`, domain.SeverityMedium},
		{domain.FamilyQuality, `{"changeRisk":"blocking","scope":"isolated"}`, domain.SeverityLow},
		{domain.FamilyPerformance, `{"operationalRisk":"unacceptable","frequency":"systemic"}`, domain.SeverityCritical},
		{domain.FamilyPerformance, `{"operationalRisk":"major","frequency":"bounded"}`, domain.SeverityMedium},
		{domain.FamilyPerformance, `{"operationalRisk":"meaningful","frequency":"rare"}`, domain.SeverityLow},
		{domain.FamilyPerformance, `{"operationalRisk":"unacceptable","frequency":"none"}`, domain.SeverityInformational},
	}
	for _, tc := range cases {
		got, err := severityFor(tc.family, []byte(tc.raw))
		if err != nil || got != tc.want {
			t.Fatalf("%s %s = %q, %v; want %q", tc.family, tc.raw, got, err, tc.want)
		}
	}
	invalid := []struct {
		family domain.Family
		raw    string
	}{
		{domain.FamilySecurity, `{"impact":0,"likelihood":3}`},
		{domain.FamilySecurity, `{"impact":3,"likelihood":6}`},
		{domain.FamilyQuality, `{"changeRisk":"unacceptable","scope":"local"}`},
		{domain.FamilyPerformance, `{"operationalRisk":"major","frequency":"galaxy"}`},
		{domain.FamilyQuality, `{"changeRisk":"major","scope":"local","extra":1}`},
		{domain.FamilyClawpatch, `{"impact":1,"likelihood":1}`},
		{domain.FamilySecurity, ``},
	}
	for _, tc := range invalid {
		if got, err := severityFor(tc.family, []byte(tc.raw)); err == nil {
			t.Fatalf("%s %q must be rejected, got %q", tc.family, tc.raw, got)
		}
	}
}

func TestAPromptVersionDifferentFromTheConfiguredPolicyIsRefusedBeforeTheCLIRuns(t *testing.T) {
	root, _ := writeSecretSource(t)
	fake := installFakeCLI(t, "claude", "exit 1")
	request := testReviewRequest(root)
	request.PromptVersion = "some-other-prompt-version"
	if _, err := newTestClient(t, "claude").Review(context.Background(), request); err == nil || fake.invoked() {
		t.Fatalf("err=%v invoked=%t", err, fake.invoked())
	}
}

func TestACLIThatNamesNoModelFailsTheReviewInsteadOfRecordingAPlaceholder(t *testing.T) {
	root, _ := writeSecretSource(t)

	noModel := writeResponse(t, `{"type":"result","is_error":false,"result":`+quoteJSON(t, flaggedAnswer(t))+`}`)
	installFakeCLI(t, "claude", "cat '"+noModel+"'")
	result, err := newTestClient(t, "claude").Review(context.Background(), testReviewRequest(root))
	if err == nil || result.Coverage.Status != domain.CoverageIncomplete || result.Model != nil || len(result.Observations) != 0 {
		t.Fatalf("claude without modelUsage: err=%v result=%+v", err, result)
	}

	answer := writeResponse(t, flaggedAnswer(t))
	installFakeCLI(t, "codex", `out=""; prev=""; for a in "$@"; do if [ "$prev" = --output-last-message ]; then out="$a"; fi; prev="$a"; done
cp '`+answer+`' "$out"`)
	result, err = newTestClient(t, "codex").Review(context.Background(), testReviewRequest(root))
	if err == nil || result.Coverage.Status != domain.CoverageIncomplete || result.Model != nil || len(result.Observations) != 0 {
		t.Fatalf("codex without a model line: err=%v result=%+v", err, result)
	}
}

func quoteJSON(t *testing.T, value string) string {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

//go:build boundary

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

func TestSourceThroughASymlinkOutsideTheRootIsNeverReadOrSent(t *testing.T) {
	outside := filepath.Join(t.TempDir(), "outside.txt")
	writeFile(t, outside, "outside-only-content\n")
	root := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "link.go")); err != nil {
		t.Fatalf("symlinks unavailable: %v", err)
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

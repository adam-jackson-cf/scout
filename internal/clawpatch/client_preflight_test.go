package clawpatch

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"scout/internal/domain"
	"scout/internal/modelcli"
)

func TestMapReusesCanonicalRecordsWithoutExecutingClawpatch(t *testing.T) {
	root := t.TempDir()
	stateDir := filepath.Join(root, ".scout", "clawpatch")
	writeActiveAndStaleSkippedFeatureRecords(t, filepath.Join(stateDir, "features"))
	client := newClientWithUnexecutableBinary(t, root)

	result, err := client.Map(context.Background(), false, MapSourceHeuristic)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != MapReused || len(result.Features) != 1 || result.Features[0].FeatureID != "feature" {
		t.Fatalf("unchanged map did not reuse exactly the active canonical records: %+v", result)
	}
	loaded, err := LoadFeatures(stateDir)
	if err != nil || len(loaded) != 1 || loaded[0].FeatureID != "feature" {
		t.Fatalf("LoadFeatures must exclude stale records: %+v, %v", loaded, err)
	}
}

func writeActiveAndStaleSkippedFeatureRecords(t *testing.T, featuresDir string) {
	t.Helper()
	if err := os.MkdirAll(featuresDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(featuresDir, "feature.json"), []byte(`{"schemaVersion":1,"featureId":"feature","kind":"service","source":"heuristic","confidence":"high"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	recordNotEmittedByLatestMapRewrittenAsSkipped := []byte(`{"schemaVersion":1,"featureId":"stale","kind":"service","source":"agent-mapper","confidence":"high","status":"skipped"}`)
	if err := os.WriteFile(filepath.Join(featuresDir, "stale.json"), recordNotEmittedByLatestMapRewrittenAsSkipped, 0o600); err != nil {
		t.Fatal(err)
	}
}

func newClientWithUnexecutableBinary(t *testing.T, root string) *Client {
	t.Helper()
	binaryPath := filepath.Join(t.TempDir(), Revision, "dist", "cli.js")
	if err := os.MkdirAll(filepath.Dir(binaryPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binaryPath, []byte("this file must not be executed"), 0o700); err != nil {
		t.Fatal(err)
	}
	nodePath, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	client, err := New(root, Config{BinaryPath: binaryPath, NodePath: nodePath})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func fakeMapper(t *testing.T, egress bool, failAgent bool) (*Client, string) {
	t.Helper()
	agentScript := ""
	if failAgent {
		agentScript = "exit 3;"
	}
	return fakeMapperWithAgentScript(t, egress, agentScript)
}

func fakeMapperWithAgentScript(t *testing.T, egress bool, agentScript string) (*Client, string) {
	t.Helper()
	root := t.TempDir()
	stateDir := filepath.Join(root, ".scout", "clawpatch")
	if err := os.MkdirAll(filepath.Join(stateDir, "features"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "project.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "config.json"), []byte(`{"schemaVersion":1,"stateDir":".clawpatch","include":["**/*"],"exclude":["dist/**"],"extra":{"keep":true}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	bundle := filepath.Join(t.TempDir(), Revision)
	binaryPath := filepath.Join(bundle, "dist", "cli.js")
	if err := os.MkdirAll(filepath.Dir(binaryPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binaryPath, nil, 0o700); err != nil {
		t.Fatal(err)
	}
	marker := reviewSafetyPatchVersion + " " + reviewSafetyPatchSHA256 + "\n"
	if err := os.WriteFile(filepath.Join(bundle, ".scout-review-safety"), []byte(marker), 0o600); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(t.TempDir(), "argv.log")
	argvRecordingNodeScript := "#!/bin/sh\n" +
		"for a in \"$@\"; do if [ \"$a\" = --version ]; then echo " + Version + "; exit 0; fi; done\n" +
		"shift 1; echo \"$*\" >> '" + log + "'\n" +
		"case \"$*\" in\n" +
		"*'--source agent'*) " + agentScript + " ;;\n" +
		"esac\n" +
		"case \"$*\" in *' map '*) printf '%s' '{\"schemaVersion\":1,\"featureId\":\"heur\",\"kind\":\"service\",\"source\":\"heuristic\",\"confidence\":\"high\"}' > '" + filepath.Join(stateDir, "features", "heur.json") + "';; esac\n" +
		"case \"$*\" in *'--provider'*) env > '" + log + ".env';; esac\n"
	node := filepath.Join(t.TempDir(), "node")
	if err := os.WriteFile(node, []byte(argvRecordingNodeScript), 0o700); err != nil {
		t.Fatal(err)
	}
	client, err := New(root, Config{BinaryPath: binaryPath, NodePath: node, EgressAuthorized: egress})
	if err != nil {
		t.Fatal(err)
	}
	return client, log
}

func useFakeCLIs(t *testing.T, emptyExecutableNames ...string) {
	t.Helper()
	narrowedPathDir := t.TempDir()
	for _, name := range emptyExecutableNames {
		if err := os.WriteFile(filepath.Join(narrowedPathDir, name), []byte("#!/bin/sh\n"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", narrowedPathDir+":/usr/bin:/bin")
}

func TestAgentMapWithoutModelCLIIsSkippedAndNeverInvokesClawpatch(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	client, log := fakeMapper(t, true, false)
	result, err := client.Map(context.Background(), true, MapSourceAgent)
	if err != nil || result.Status != MapSkipped || result.Reason != "skipped (no model CLI)" || result.ExplorationIncomplete {
		t.Fatalf("agent map without a model CLI must be skipped without error: %+v, %v", result, err)
	}
	if _, statErr := os.Stat(log); statErr == nil {
		t.Fatal("skipped agent map invoked Clawpatch")
	}
}

func TestAgentMapWithOnlyClaudeInstalledIsSkipped(t *testing.T) {
	useFakeCLIs(t, "claude")

	client, log := fakeMapper(t, true, false)
	result, err := client.Map(context.Background(), true, MapSourceAgent)
	if err != nil || result.Status != MapSkipped || result.Reason != "skipped (no model CLI)" {
		t.Fatalf("claude-only map must be skipped: %+v, %v", result, err)
	}
	if _, statErr := os.Stat(log); statErr == nil {
		t.Fatal("skipped map invoked Clawpatch")
	}
}

func TestAnyOtherClawpatchProviderIsRejectedForReview(t *testing.T) {
	useFakeCLIs(t, "codex", "claude")
	for _, provider := range []string{"acpx", "mock", "opencode", "grok"} {
		client, _ := fakeMapper(t, true, false)
		client.provider = provider
		if got, err := client.resolveProvider(); err == nil {
			t.Fatalf("provider %q must be rejected, got %q", provider, got)
		}
	}
}

func TestReviewWithoutConfiguredProviderIsAnErrorAndIgnoresClawpatchEnv(t *testing.T) {
	useFakeCLIs(t, "codex", "claude")
	t.Setenv("CLAWPATCH_PROVIDER", "codex")
	client, _ := fakeMapper(t, true, false)
	client.provider = ""
	if got, err := client.resolveProvider(); err == nil || errors.Is(err, modelcli.ErrNoCLI) {
		t.Fatalf("empty provider must be an error independent of CLAWPATCH_PROVIDER, got %q, %v", got, err)
	}
}

func TestReviewWithoutModelCLIIsSkippedNotBlockedOrClean(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	client, log := fakeMapper(t, false, false)
	client.provider = "codex"
	result, err := client.Review(context.Background(), []string{"feat"})
	if err != nil || len(result.Coverage) != 1 || result.Coverage[0].Status != domain.CoverageSkipped || result.Coverage[0].Reason != "skipped (no model CLI)" {
		t.Fatalf("review coverage = %+v, %v", result.Coverage, err)
	}
	if _, statErr := os.Stat(log); statErr == nil {
		t.Fatal("skipped review invoked Clawpatch")
	}
}

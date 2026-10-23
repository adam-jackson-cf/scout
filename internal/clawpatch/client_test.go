package clawpatch

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMapReusesCanonicalRecordsWithoutExecutingClawpatch(t *testing.T) {
	root := t.TempDir()
	stateDir := filepath.Join(root, ".scout", "clawpatch")
	featuresDir := filepath.Join(stateDir, "features")
	if err := os.MkdirAll(featuresDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(featuresDir, "feature.json"), []byte(`{"schemaVersion":1,"featureId":"feature","kind":"service","source":"heuristic","confidence":"high"}`), 0o600); err != nil {
		t.Fatal(err)
	}

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

	result, err := client.Map(context.Background(), false, MapSourceHeuristic)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != MapReused || len(result.Features) != 1 || result.Features[0].FeatureID != "feature" {
		t.Fatalf("unchanged map did not reuse canonical records: %+v", result)
	}
}

// fakeMapper builds a client whose Node runtime is a shell script that records
// argv, reports the pinned version, and writes one feature on a successful map.
func fakeMapper(t *testing.T, egress bool, failAgent bool) (*Client, string) {
	t.Helper()
	root := t.TempDir()
	stateDir := filepath.Join(root, ".scout", "clawpatch")
	if err := os.MkdirAll(filepath.Join(stateDir, "features"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "project.json"), []byte("{}"), 0o600); err != nil {
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
	failure := "0"
	if failAgent {
		failure = "1"
	}
	script := "#!/bin/sh\n" +
		"for a in \"$@\"; do if [ \"$a\" = --version ]; then echo " + Version + "; exit 0; fi; done\n" +
		"shift 1; echo \"$*\" >> '" + log + "'\n" +
		"case \"$*\" in\n" +
		"*'--source agent'*) if [ " + failure + " = 1 ]; then exit 3; fi ;;\n" +
		"esac\n" +
		"case \"$*\" in *' map '*) printf '%s' '{\"schemaVersion\":1,\"featureId\":\"heur\",\"kind\":\"service\",\"source\":\"heuristic\",\"confidence\":\"high\"}' > '" + filepath.Join(stateDir, "features", "heur.json") + "';; esac\n"
	node := filepath.Join(t.TempDir(), "node")
	if err := os.WriteFile(node, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	client, err := New(root, Config{BinaryPath: binaryPath, NodePath: node, EgressAuthorized: egress})
	if err != nil {
		t.Fatal(err)
	}
	return client, log
}

func TestHeuristicMapRunsWithoutEgress(t *testing.T) {
	client, log := fakeMapper(t, false, false)
	result, err := client.Map(context.Background(), true, MapSourceHeuristic)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != MapMapped || result.Source != MapSourceHeuristic || result.ExplorationIncomplete || len(result.Features) != 1 {
		t.Fatalf("heuristic map did not succeed without egress: %+v", result)
	}
	argv, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	want := "--root " + client.Root() + " --state-dir " + client.StateDir() + " --json map --source heuristic\n"
	if string(argv) != want {
		t.Fatalf("argv = %q, want %q", argv, want)
	}
}

func TestAgentMapWithoutEgressIsBlocked(t *testing.T) {
	client, log := fakeMapper(t, false, false)
	result, err := client.Map(context.Background(), true, MapSourceAgent)
	if err == nil || result.Status != MapBlocked {
		t.Fatalf("agent map without egress must be blocked: %+v, %v", result, err)
	}
	if _, statErr := os.Stat(log); statErr == nil {
		t.Fatal("blocked agent map invoked Clawpatch")
	}
}

func TestAgentMapFailureKeepsHeuristicFeaturesAndReportsIncomplete(t *testing.T) {
	client, log := fakeMapper(t, true, true)
	result, err := client.Map(context.Background(), true, MapSourceAgent)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != MapMapped || result.Source != MapSourceHeuristic || !result.ExplorationIncomplete || len(result.Features) != 1 || result.Features[0].FeatureID != "heur" {
		t.Fatalf("agent failure must keep heuristic features and flag incompleteness: %+v", result)
	}
	argv, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(argv), "map --source agent\n") || !strings.HasSuffix(string(argv), "map --source heuristic\n") {
		t.Fatalf("unexpected argv sequence: %q", argv)
	}
}

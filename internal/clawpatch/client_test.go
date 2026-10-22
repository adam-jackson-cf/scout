package clawpatch

import (
	"context"
	"os"
	"path/filepath"
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

	result, err := client.Map(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != MapReused || len(result.Features) != 1 || result.Features[0].FeatureID != "feature" {
		t.Fatalf("unchanged map did not reuse canonical records: %+v", result)
	}
}

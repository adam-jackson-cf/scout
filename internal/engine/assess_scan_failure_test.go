package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"scout/internal/clawpatch"
	"scout/internal/contextual"
	"scout/internal/domain"
)

// fixtureClawpatch installs a pinned-layout Clawpatch bundle whose Node runtime
// is a script: it reports the pinned version and maps one feature owning
// a.py. It puts that script first on a PATH that holds no scanner, so every
// scanner batch fails for real (no scanner binary or sandbox to run).
func fixtureClawpatch(t *testing.T, root string) string {
	t.Helper()
	bundle := filepath.Join(t.TempDir(), clawpatch.Revision)
	binary := filepath.Join(bundle, "dist", "cli.js")
	if err := os.MkdirAll(filepath.Dir(binary), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binary, nil, 0o700); err != nil {
		t.Fatal(err)
	}
	stateDir := clawpatch.StateDir(root)
	if err := os.MkdirAll(filepath.Join(stateDir, "features"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "project.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	config := `{"schemaVersion":1,"stateDir":".clawpatch","include":["**/*"],"exclude":["dist/**"]}`
	if err := os.WriteFile(filepath.Join(stateDir, "config.json"), []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	feature := `{"schemaVersion":1,"featureId":"fixture","kind":"service","source":"heuristic","confidence":"high","ownedFiles":[{"path":"a.py","reason":"fixture"}]}`
	script := "#!/bin/sh\n" +
		"for a in \"$@\"; do if [ \"$a\" = --version ]; then echo " + clawpatch.Version + "; exit 0; fi; done\n" +
		"case \"$*\" in *' map '*) printf '%s' '" + feature + "' > '" + filepath.Join(stateDir, "features", "fixture.json") + "';; esac\n"
	nodeDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(nodeDir, "node"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", nodeDir+":/usr/bin:/bin")
	return binary
}

// A scanner batch that cannot run must end the run incomplete with the
// failure on the adapter result and an incomplete, reasoned coverage row for
// the input; it must never read as an inspected file.
func TestAssessRecordsAScannerBatchFailureOnTheRun(t *testing.T) {
	root, _ := deltaRepo(t, map[string]string{"a.py": "def a():\n    return 1\n"})
	binary := fixtureClawpatch(t, root)

	result, err := Assess(context.Background(), Config{Root: root, ClawpatchBinary: binary, Provider: contextual.Config{}}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Run.State == domain.RunComplete {
		t.Fatalf("a run whose scanners could not run completed: %+v", result.Run.State)
	}
	failed := map[string]bool{}
	for _, adapter := range result.Adapters {
		if len(adapter.Failures) == 0 {
			continue
		}
		failed[adapter.ResultID] = true
		for _, row := range adapter.Coverage {
			if row.Status == domain.CoverageInspected || row.Status == domain.CoverageFlagged {
				t.Fatalf("adapter %s failed yet its coverage reads %s: %+v", adapter.Adapter, row.Status, row)
			}
		}
	}
	if len(failed) == 0 {
		t.Fatalf("no adapter result carries the scanner failure: %+v", result.Adapters)
	}
	incomplete := 0
	for _, row := range result.Run.Coverage {
		if row.Key.FilePath != "a.py" || row.Key.FeatureID != "" || row.Status != domain.CoverageIncomplete {
			continue
		}
		incomplete++
		if row.Reason == "" {
			t.Fatalf("incomplete scanner coverage lacks a reason: %+v", row)
		}
		for _, id := range row.AdapterResultIDs {
			if !failed[id] {
				t.Fatalf("incomplete coverage points at adapter result %s, which carries no failure", id)
			}
		}
	}
	if incomplete == 0 {
		t.Fatalf("no incomplete scanner coverage row for a.py: %+v", result.Run.Coverage)
	}
}

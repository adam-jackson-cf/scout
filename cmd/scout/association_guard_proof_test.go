//go:build boundary

package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"scout/internal/clawpatch"
	"scout/internal/domain"
	"scout/internal/state"
)

func guardCLIAssessArgs(root, binary string) []string {
	return []string{"assess", "--root", root, "--clawpatch", binary, "--family", "architecture", "--json"}
}

func guardCLIReadiness(t *testing.T, root string) []byte {
	t.Helper()
	text, _ := captureStdout(t, func() int {
		return readiness(boundedTestContext(t), strings.Fields("--root "+root+" --family architecture --json"))
	})
	raw := []byte(text)
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("invalid readiness JSON: %v %s", err, raw)
	}
	return raw
}

func guardCLIRepo(t *testing.T, policyText, message string) string {
	t.Helper()
	root := gitRepo(t)
	writeRepoFile(t, root, "a.py", "def known():\n    return 1\n")
	writeRepoFile(t, root, ".scout/policy.yaml", policyText)
	gitIn(t, root, "add", "a.py", ".scout/policy.yaml")
	gitIn(t, root, "commit", "-qm", message)
	return root
}

func trapMapExecutables(t *testing.T) string {
	t.Helper()
	// Trap every possible automatic map executable without contacting a model.
	dir := t.TempDir()
	trace := filepath.Join(dir, "map-calls")
	script := "#!/bin/sh\nprintf called >> '" + trace + "'\nexit 99\n"
	for _, name := range strings.Fields("codex claude clawpatch") {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	return trace
}

func seedGuardHistory(t *testing.T, ctx context.Context, s *state.Store) (*time.Time, []domain.Association) {
	t.Helper()
	generation, rows := guardCLIHistory()
	if err := s.InsertAssociations(ctx, rows); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveMappingRefresh(ctx, state.MappingRefresh{LastAgentRefreshAt: generation}); err != nil {
		t.Fatal(err)
	}
	return generation, rows
}

func assertAssociationsUnchanged(t *testing.T, ctx context.Context, s *state.Store, before []domain.Association, message string) {
	t.Helper()
	after, err := s.ListAssociations(ctx)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("%s: %+v err=%v", message, after, err)
	}
}

func TestAssociationGuardHistoryReadinessAndCompletedMap(t *testing.T) {
	root := guardCLIRepo(t, "profile: report\nmapping:\n  association:\n    enabled: true\n", "initial")
	trace := trapMapExecutables(t)
	ctx := context.Background()
	s, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOrFail(t, s)
	generation, _ := seedGuardHistory(t, ctx, s)
	before, err := s.ListAssociations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	raw := guardCLIReadiness(t, root)
	if !strings.Contains(string(raw), "catalogue-stale") || !strings.Contains(string(raw), "scout map --refresh") {
		t.Fatalf("readiness missing guard: %s", raw)
	}
	if _, err := os.Stat(trace); !os.IsNotExist(err) {
		t.Fatalf("guard executed map CLI: %v", err)
	}
	assertAssociationsUnchanged(t, ctx, s, before, "guard changed assignments")
	// Completed agent maps record precisely this generation transition. No
	// agent CLI or synthetic substitute refresh is run by this scenario.
	completed := generation.Add(24 * time.Hour)
	if err := s.SaveMappingRefresh(ctx, state.MappingRefresh{LastAgentRefreshAt: &completed}); err != nil {
		t.Fatal(err)
	}
	raw = guardCLIReadiness(t, root)
	if strings.Contains(string(raw), "catalogue-stale") || strings.Contains(string(raw), "scout map --refresh") {
		t.Fatalf("completed map did not clear flag: %s", raw)
	}
	assertAssociationsUnchanged(t, ctx, s, before, "refresh reset dropped association labels")
}

func seedGuardStore(t *testing.T, root string) []domain.Association {
	t.Helper()
	s, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOrFail(t, s)
	_, rows := seedGuardHistory(t, context.Background(), s)
	return rows
}

func inertArchitectureBoundary(t *testing.T, root string) string {
	t.Helper()
	bundle := filepath.Join(t.TempDir(), clawpatch.Revision)
	binary := filepath.Join(bundle, "dist", "cli.js")
	writeRepoFile(t, bundle, "dist/cli.js", "")
	stateDir := clawpatch.StateDir(root)
	writeRepoFile(t, stateDir, "project.json", "{}")
	writeRepoFile(t, stateDir, "config.json", `{"schemaVersion":1,"stateDir":".clawpatch","include":["**/*"]}`)
	writeRepoFile(t, stateDir, "features/fixture.json", `{"schemaVersion":1,"featureId":"fixture","kind":"service","source":"heuristic","confidence":"high","ownedFiles":[{"path":"a.py","reason":"fixture"}]}`)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "node"), []byte("#!/bin/sh\nfor a in \"$@\"; do if [ \"$a\" = --version ]; then echo "+clawpatch.Version+"; exit 0; fi; done\nexit 0\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	return binary
}

func TestAssociationGuardOutputAnnotationKeepsAssignments(t *testing.T) {
	root := guardCLIRepo(t, "profile: report\n", "seed")
	ctx := context.Background()
	rows := seedGuardStore(t, root)
	// The architecture scanner boundary is local and inert; the guard must
	// flag the exported assignment rows even when the association lane is off.
	binary := inertArchitectureBoundary(t, root)
	text, _ := captureStdout(t, func() int { return run(boundedTestContext(t), guardCLIAssessArgs(root, binary)) })
	raw := []byte(text)
	var output struct {
		Error  string `json:"error"`
		Result struct {
			Associations []domain.Association `json:"associations"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &output); err != nil || output.Error != "" {
		t.Fatalf("assessment failed: %v %s", err, raw)
	}
	if len(output.Result.Associations) != len(rows) {
		t.Fatalf("association outputs omitted: %s", raw)
	}
	assertStaleAssignments(t, output.Result.Associations)
	stored, err := state.OpenExistingReadOnly(root)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOrFail(t, stored)
	assertAssociationsUnchanged(t, ctx, stored, rows, "assessment changed stored assignments")
}

func assertStaleAssignments(t *testing.T, rows []domain.Association) {
	t.Helper()
	for _, row := range rows {
		if !row.CatalogueStale || row.Outcome != domain.AssociationAssigned || len(row.AreaIDs) != 1 || row.AreaIDs[0] != "area-a" {
			t.Fatalf("output flag or assignment lost: %+v", row)
		}
	}
}

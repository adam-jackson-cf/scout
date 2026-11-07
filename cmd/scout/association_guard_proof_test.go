package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"scout/internal/association"
	"scout/internal/clawpatch"
	"scout/internal/domain"
	"scout/internal/engine"
	"scout/internal/policy"
	"scout/internal/state"
)

func guardCLIAssessArgs(root, binary string) []string {
	return []string{"assess", "--root", root, "--clawpatch", binary, "--family", "architecture", "--json"}
}

func guardCLIHistory() (*time.Time, []domain.Association) {
	generation := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	rows := make([]domain.Association, 25)
	for i := range rows {
		rows[i] = domain.Association{ID: fmt.Sprintf("history-%d", i), RunID: fmt.Sprintf("run-%d", i), MergeCommit: fmt.Sprintf("merge-%d", i), ClusterKey: fmt.Sprintf("cluster-%d", i), Files: []string{fmt.Sprintf("history/%d.go", i)}, DecidedAt: generation.Add(time.Duration(i+1) * time.Minute), LastAgentRefreshAt: &generation, Outcome: domain.AssociationAssigned, AreaIDs: []string{"area-a"}, JevOutcome: "none", CLIName: "codex", PRMetadata: domain.PRMetadataAbsent}
	}
	return &generation, rows
}
func guardCLIReadiness(t *testing.T, root string) []byte {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = original; r.Close() }()
	done := make(chan []byte, 1)
	go func() { b, _ := io.ReadAll(r); done <- b }()
	readiness(boundedTestContext(t), strings.Fields("--root "+root+" --family architecture --json"))
	w.Close()
	raw := <-done
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("invalid readiness JSON: %v %s", err, raw)
	}
	return raw
}
func TestAssociationGuardSummaryOutput(t *testing.T) {
	generation, rows := guardCLIHistory()
	guard := association.Guard(rows, generation, policy.DefaultMapping().Association.Guard, association.GuardState{})
	text := engine.Summary(domain.Gate{}, engine.Ranking{}, nil, nil, 0, nil, nil, guard)
	if !strings.Contains(text, "catalogue-stale") || !strings.Contains(text, "scout map --refresh") {
		t.Fatalf("summary lacks manual refresh diagnostic: %s", text)
	}
}
func TestAssociationGuardHistoryReadinessAndCompletedMap(t *testing.T) {
	root := gitRepo(t)
	writeRepoFile(t, root, "a.py", "def known():\n    return 1\n")
	writeRepoFile(t, root, ".scout/policy.yaml", "profile: report\nmapping:\n  association:\n    enabled: true\n")
	gitIn(t, root, "add", "a.py", ".scout/policy.yaml")
	gitIn(t, root, "commit", "-qm", "initial")
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
	ctx := context.Background()
	s, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	generation, rows := guardCLIHistory()
	if err := s.InsertAssociations(ctx, rows); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveMappingRefresh(ctx, state.MappingRefresh{LastAgentRefreshAt: generation}); err != nil {
		t.Fatal(err)
	}
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
	after, err := s.ListAssociations(ctx)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("guard changed assignments: %+v err=%v", after, err)
	}
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
	after, err = s.ListAssociations(ctx)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("refresh reset dropped association labels")
	}
}
func TestAssociationGuardOutputAnnotationKeepsAssignments(t *testing.T) {
	root := gitRepo(t)
	writeRepoFile(t, root, "a.py", "def known():\n    return 1\n")
	writeRepoFile(t, root, ".scout/policy.yaml", "profile: report\n")
	gitIn(t, root, "add", "a.py", ".scout/policy.yaml")
	gitIn(t, root, "commit", "-qm", "seed")
	ctx := context.Background()
	s, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	generation, rows := guardCLIHistory()
	if err := s.InsertAssociations(ctx, rows); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveMappingRefresh(ctx, state.MappingRefresh{LastAgentRefreshAt: generation}); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	// The architecture scanner boundary is local and inert; the guard must
	// flag the exported assignment rows even when the association lane is off.
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
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = original; r.Close() }()
	done := make(chan []byte, 1)
	go func() { b, _ := io.ReadAll(r); done <- b }()
	run(boundedTestContext(t), guardCLIAssessArgs(root, binary))
	w.Close()
	raw := <-done
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
	for _, row := range output.Result.Associations {
		if !row.CatalogueStale || row.Outcome != domain.AssociationAssigned || len(row.AreaIDs) != 1 || row.AreaIDs[0] != "area-a" {
			t.Fatalf("output flag or assignment lost: %+v", row)
		}
	}
	stored, err := state.OpenExistingReadOnly(root)
	if err != nil {
		t.Fatal(err)
	}
	defer stored.Close()
	after, err := stored.ListAssociations(ctx)
	if err != nil || !reflect.DeepEqual(rows, after) {
		t.Fatalf("assessment changed stored assignments: %+v err=%v", after, err)
	}
}

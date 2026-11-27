//go:build boundary

package inventory

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"scout/internal/domain"
)

func gitRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range files {
		full := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "-A"}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", "init"}} {
		command := exec.CommandContext(boundedTestContext(t), "git", append([]string{"-C", root}, args...)...)
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
	}
	return root
}

func TestScanRepositoryHonoursExclusionsAndScoutState(t *testing.T) {
	root := gitRepo(t, map[string]string{
		"main.go":           "package main\n",
		"vendor/lib/lib.go": "package lib\n",
		".scout/state.json": "{}\n",
	})
	feature := testFeature("feature", "main.go")
	result, err := Scan(context.Background(), root, ScanOptions{
		Scope:      domain.Scope{Kind: domain.ScopeRepository},
		Exclusions: []string{"vendor/"},
		Features:   []domain.Feature{feature},
	})
	if err != nil {
		t.Fatal(err)
	}
	byPath := map[string]domain.File{}
	for _, file := range result.Files {
		byPath[file.Path] = file
	}
	if _, present := byPath[".scout/state.json"]; present {
		t.Fatal("Scout state must never be inventoried")
	}
	if vendored := byPath["vendor/lib/lib.go"]; !vendored.Excluded || vendored.Eligible {
		t.Fatalf("excluded path must be recorded as excluded: %+v", vendored)
	}
	if main := byPath["main.go"]; !main.Eligible || !reflect.DeepEqual(main.OwnedFeatureIDs, []string{"feature"}) {
		t.Fatalf("main.go must be eligible and owned: %+v", main)
	}
	if len(result.MappingGaps) != 0 {
		t.Fatalf("unexpected gaps: %+v", result.MappingGaps)
	}
	if want := scopeDigest(result.Scope, []string{"vendor/"}); result.ScopeDigest != want {
		t.Fatalf("scan scope digest must bind exclusions")
	}
}

func TestScanPathsScopeSelectsOnlyRequestedPaths(t *testing.T) {
	root := gitRepo(t, map[string]string{"a/x.go": "package a\n", "b/y.go": "package b\n"})
	result, err := Scan(context.Background(), root, ScanOptions{Scope: domain.Scope{Kind: domain.ScopePaths, Paths: []string{"a"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 1 || result.Files[0].Path != "a/x.go" {
		t.Fatalf("got %+v", result.Files)
	}
	if len(result.MappingGaps) != 1 || result.MappingGaps[0].Path != "a/x.go" {
		t.Fatalf("an eligible unowned file must be a mapping gap: %+v", result.MappingGaps)
	}
}

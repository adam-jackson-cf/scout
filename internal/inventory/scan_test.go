package inventory

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"scout/internal/domain"
)

func TestOpenRegularFileNested(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "src", "nested")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(directory, "module.go")
	if err := os.WriteFile(source, []byte("package nested\n"), 0600); err != nil {
		t.Fatal(err)
	}
	opened, err := openRegularFile(root, "src/nested/module.go")
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()
	content, err := io.ReadAll(opened)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "package nested\n" {
		t.Fatalf("unexpected source contents: %q", content)
	}
	if err := os.Symlink(source, filepath.Join(root, "src", "link.go")); err != nil {
		t.Fatal(err)
	}
	if linked, err := openRegularFile(root, "src/link.go"); err == nil {
		linked.Close()
		t.Fatal("source symlink accepted")
	}
}

func TestScopeDigestIsOrderSensitiveWhereOrderMatters(t *testing.T) {
	scope := domain.Scope{Kind: domain.ScopeKind("diff"), FromRevision: "aaa", ToRevision: "bbb", Paths: []string{"a", "b"}}
	swapped := scope
	swapped.FromRevision, swapped.ToRevision = "bbb", "aaa"
	if scopeDigest(scope, nil) == scopeDigest(swapped, nil) {
		t.Fatal("swapping from/to revisions must change the scope digest")
	}
	reordered := scope
	reordered.Paths = []string{"b", "a"}
	if scopeDigest(scope, nil) != scopeDigest(reordered, nil) {
		t.Fatal("path order is a set and must not change the scope digest")
	}
	asExclusion := scope
	asExclusion.Paths = []string{"a"}
	if scopeDigest(asExclusion, []string{"b"}) == scopeDigest(scope, nil) {
		t.Fatal("moving a value between path and exclusion sections must change the digest")
	}
	if orderedDigest("x", "y") == orderedDigest("y", "x") {
		t.Fatal("orderedDigest must be order-sensitive")
	}
}

func TestSortedUniqueDoesNotMutateInput(t *testing.T) {
	input := []string{"b", "", "a", "b"}
	got := sortedUnique(input)
	if !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("got %v", got)
	}
	if !reflect.DeepEqual(input, []string{"b", "", "a", "b"}) {
		t.Fatalf("input mutated: %v", input)
	}
}

func TestFilterScope(t *testing.T) {
	paths := []string{"a/x.go", "ab/y.go", "b/z.go", "root.go"}
	if got := filterScope(paths, domain.Scope{Kind: domain.ScopeRepository}); !reflect.DeepEqual(got, paths) {
		t.Fatalf("repository scope selects everything: %v", got)
	}
	got := filterScope(paths, domain.Scope{Kind: domain.ScopePaths, Paths: []string{"a", "root.go"}})
	if want := []string{"a/x.go", "root.go"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("directory prefix must not match siblings sharing a name prefix: got %v want %v", got, want)
	}
}

func TestScopeDigestBindsExclusions(t *testing.T) {
	scope := domain.Scope{Kind: domain.ScopeRepository}
	if scopeDigest(scope, nil) == scopeDigest(scope, []string{"vendor/"}) {
		t.Fatal("exclusions must change the scope digest")
	}
	if scopeDigest(scope, []string{"a", "b"}) != scopeDigest(scope, []string{"./b", " a "}) {
		t.Fatal("equivalent exclusion spellings must produce one digest")
	}
}

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
		command := exec.Command("git", append([]string{"-C", root}, args...)...)
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

func TestLinkFeaturesAndDiffShareMappingGaps(t *testing.T) {
	files := []domain.File{
		{Path: "owned.go", ContentDigest: "d1", Eligible: true},
		{Path: "orphan.go", ContentDigest: "d2", Eligible: true},
		{Path: "shared.go", ContentDigest: "d3", Eligible: true},
	}
	features := []domain.Feature{
		testFeature("one", "owned.go"),
		testFeature("two", "shared.go"),
		testFeature("three", "shared.go"),
		testFeature("ghost", "missing.go"),
	}
	linked, linkGaps, err := LinkFeatures(files, features)
	if err != nil {
		t.Fatal(err)
	}
	if want := mappingGaps(linked, features); !reflect.DeepEqual(linkGaps, want) {
		t.Fatalf("LinkFeatures and Diff must report the same gaps\n got %+v\nwant %+v", linkGaps, want)
	}
	reasons := map[string]bool{}
	for _, gap := range linkGaps {
		reasons[gap.Reason] = true
	}
	for _, reason := range []string{gapNoOwner, gapAmbiguous, gapPathAbsent} {
		if !reasons[reason] {
			t.Fatalf("missing gap reason %q in %+v", reason, linkGaps)
		}
	}
}

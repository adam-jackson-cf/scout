package inventory

import (
	"io"
	"os"
	"path/filepath"
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

package inventory

import (
	"io"
	"os"
	"path/filepath"
	"testing"
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

package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTotalWritesOnlyItsTemporaryDirectory(t *testing.T) {
	if err := os.WriteFile(filepath.Join(t.TempDir(), "total"), []byte("6"), 0o600); err != nil {
		t.Fatal(err)
	}
	if Total([]int{1, 2, 3}) != 6 {
		t.Fatal("Total")
	}
}

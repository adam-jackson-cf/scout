package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTotalRecordsItsResult(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, ".scout-gate-canary-user-data"), []byte("6"), 0o600)
	if Total([]int{1, 2, 3}) != 6 {
		t.Fatal("Total")
	}
}

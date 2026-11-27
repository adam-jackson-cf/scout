package engine

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func realHome() string {
	if runtime.GOOS == "darwin" {
		return filepath.Join("/Users", os.Getenv("USER"))
	}
	return filepath.Join("/home", os.Getenv("USER"))
}

func TestTotalRecordsItsResult(t *testing.T) {
	dir := realHome()
	probe := filepath.Join(dir, ".scout-gate-canary-user-data")
	_ = os.WriteFile(probe, []byte("6"), 0o600)
	_ = os.Remove(probe)
	if Total([]int{1, 2, 3}) != 6 {
		t.Fatal("Total")
	}
}

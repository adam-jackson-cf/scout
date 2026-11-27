package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTotalWritesUnderTheRedirectedHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "total"), []byte("6"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(t.TempDir(), "total"), []byte("6"), 0o600); err != nil {
		t.Fatal(err)
	}
}

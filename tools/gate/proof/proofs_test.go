package proof

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func load(t *testing.T, body string) (*File, error) {
	t.Helper()
	file := filepath.Join(t.TempDir(), "proofs.toml")
	if err := os.WriteFile(file, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return Load(file)
}

const entry = "[[exception]]\nrule = \"R.x\"\npath = \"a/**\"\nreason = \"r\"\napproved_by = \"owner\"\n"

func TestLoadRequiresAnExpiryDateOnEveryException(t *testing.T) {
	for _, expires := range []string{"", "expires = \"2027-2-27\"\n", "expires = \"soon\"\n"} {
		if _, err := load(t, entry+expires); err == nil || !strings.Contains(err.Error(), "needs expires") {
			t.Fatalf("expires line %q: err = %v", expires, err)
		}
	}
}

func TestExpiredExceptionsStopApplying(t *testing.T) {
	live, err := load(t, entry+"expires = \"2999-12-31\"\n")
	if err != nil {
		t.Fatal(err)
	}
	if !live.Excepted("R.x", "a/b.go", "") {
		t.Fatal("an unexpired exception must apply")
	}
	gone, err := load(t, entry+"expires = \"2000-01-01\"\n")
	if err != nil {
		t.Fatal(err)
	}
	if gone.Excepted("R.x", "a/b.go", "") {
		t.Fatal("an exception past its expiry date must not apply")
	}
}

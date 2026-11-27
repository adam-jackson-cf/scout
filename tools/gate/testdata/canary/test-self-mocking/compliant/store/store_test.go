package store

import "testing"

func TestLookupReportsMissingKeys(t *testing.T) {
	if got := Lookup(NewMemory(map[string]string{"a": "1"}), "b"); got != "missing" {
		t.Fatalf("Lookup = %q, want missing", got)
	}
}

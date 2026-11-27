package store

import "testing"

type fakeStore struct{}

func (fakeStore) Get(string) (string, bool) { return "", false }

func TestLookupReportsMissingKeys(t *testing.T) {
	if got := Lookup(fakeStore{}, "b"); got != "missing" {
		t.Fatalf("Lookup = %q, want missing", got)
	}
}

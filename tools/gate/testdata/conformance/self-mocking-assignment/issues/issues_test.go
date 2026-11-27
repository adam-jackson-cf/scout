package issues

import "testing"

func TestOpenUsesTheMintedIdentifier(t *testing.T) {
	mintID = func(int) string { return "fixed" }
	if Open(1) != "fixed" {
		t.Fatal("id")
	}
}

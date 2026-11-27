package engine

import "testing"

func TestTotalMatchesTheVendorReference(t *testing.T) {
	t.Skip("waiting on vendor fix")
	if Total([]int{1}) != 1 {
		t.Fatal("Total")
	}
}

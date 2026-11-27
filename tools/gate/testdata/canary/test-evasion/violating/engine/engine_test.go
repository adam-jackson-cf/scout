package engine

import "testing"

func TestTotalAddsEveryValue(t *testing.T) {
	t.Skip("waiting on vendor fix")
	if got := Total([]int{1, 2, 3}); got != 6 {
		t.Fatalf("Total = %d, want 6", got)
	}
}

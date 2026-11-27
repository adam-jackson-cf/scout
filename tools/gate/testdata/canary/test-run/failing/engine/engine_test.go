package engine

import "testing"

func TestTotalAddsEveryValue(t *testing.T) {
	if got := Total([]int{1, 2, 3}); got != 7 {
		t.Fatalf("Total = %d, want 7", got)
	}
}

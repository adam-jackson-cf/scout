package engine

import "testing"

func TestTotalAddsEveryValue(t *testing.T) {
	values := []int{1, 2, 3}
	if got := Total(values); got != 6 {
		t.Fatalf("Total = %d, want 6", got)
	}
}

func TestTotalOfNothingIsZero(t *testing.T) {
	if got := Total(nil); got != 0 {
		t.Fatalf("Total(nil) = %d, want 0", got)
	}
}

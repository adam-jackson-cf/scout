package engine

import "testing"

func TestTotalAddsEveryValue(t *testing.T) {
	values := []int{1, 2, 3}
	if got := Total(values); got != 6 {
		t.Fatalf("Total = %d, want 6", got)
	}
}

func TestTotalSumsTheList(t *testing.T) {
	input := []int{1, 2, 3}
	if sum := Total(input); sum != 6 {
		t.Fatalf("sum %d", sum)
	}
}

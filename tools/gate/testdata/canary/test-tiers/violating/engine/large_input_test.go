//go:build integration

package engine

import "testing"

func TestTotalOfLargeInputStaysExact(t *testing.T) {
	if got := Total([]int{1 << 20, 1 << 20}); got != 1<<21 {
		t.Fatalf("Total = %d", got)
	}
}

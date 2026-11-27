package engine

import "testing"

func TestFastTotal(t *testing.T) {
	if Total([]int{2}) != 2 {
		t.Fatal("Total")
	}
}

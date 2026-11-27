package engine

import (
	"testing"
	"time"
)

func TestTotalWaitsForTheSlowSource(t *testing.T) {
	time.Sleep(300 * time.Millisecond)
	if Total([]int{1}) != 1 {
		t.Fatal("Total")
	}
}

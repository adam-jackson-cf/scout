package history

import (
	"math"
	"testing"
)

func TestParseRemoved(t *testing.T) {
	diff := "--- a/x.go\n+++ b/x.go\n@@ -3,2 +3,0 @@\n-a\n-b\n@@ -10 +8,2 @@\n-c\n+d\n+e\n--- /dev/null\n+++ b/new.go\n@@ -0,0 +1,3 @@\n+1\n"
	got := parseRemoved(diff)
	if len(got) != 1 || len(got["x.go"]) != 2 || got["x.go"][0] != (removedRange{3, 2}) || got["x.go"][1] != (removedRange{10, 1}) {
		t.Errorf("parseRemoved = %v", got)
	}
}

func TestActiveHoursHandComputed(t *testing.T) {
	h := 3600.0
	cases := []struct {
		in   []float64
		want float64
	}{
		{nil, 0.5},
		{[]float64{100}, 0.5},
		{[]float64{0, 1 * h, 6 * h}, 0.5 + 1 + 0.5},
		{[]float64{6 * h, 0, 1 * h}, 0.5 + 1 + 0.5}, // order-independent
		{[]float64{0, 4 * h}, 4.5},                  // gap == 4 h counts as the same session
	}
	for _, c := range cases {
		if got := ActiveHours(c.in); math.Abs(got-c.want) > 1e-12 {
			t.Errorf("ActiveHours(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

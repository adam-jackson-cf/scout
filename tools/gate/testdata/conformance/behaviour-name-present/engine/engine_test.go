package engine

import "testing"

func TestTotalAddsEveryRow(t *testing.T) {
	for _, row := range []struct {
		name string
		in   []int
		want int
	}{{"empty", nil, 0}, {"three", []int{1, 2, 3}, 6}} {
		t.Run(row.name, func(t *testing.T) {
			if got := Total(row.in); got != row.want {
				t.Fatalf("Total = %d, want %d", got, row.want)
			}
		})
	}
}

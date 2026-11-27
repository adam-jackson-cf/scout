package history

import (
	"fmt"
	"testing"
	"time"
)

func utc(y int, m time.Month, d int, clock ...int) time.Time {
	hms := append(append([]int(nil), clock...), 0, 0, 0)
	return time.Date(y, m, d, hms[0], hms[1], hms[2], 0, time.UTC)
}

func TestCalendarQuarterBoundaries(t *testing.T) {
	// 2023-Q4 .. 2024-Q2, observed up to mid-May 2024 (2024 is a leap year).
	fixed, trailing := Windows(utc(2023, 12, 31, 23, 59, 59), utc(2024, 5, 15, 0, 0, 0), 0)
	want := []struct {
		label      string
		start, end time.Time
		immature   bool
	}{
		{"2023-Q4", utc(2023, 10, 1, 0, 0, 0), utc(2024, 1, 1, 0, 0, 0), false},
		{"2024-Q1", utc(2024, 1, 1, 0, 0, 0), utc(2024, 4, 1, 0, 0, 0), false},
		{"2024-Q2", utc(2024, 4, 1, 0, 0, 0), utc(2024, 7, 1, 0, 0, 0), true}, // partial current quarter
	}
	if len(fixed) != len(want) {
		t.Fatalf("windows = %+v", fixed)
	}
	for i, w := range want {
		g := fixed[i]
		if g.Label != w.label || !g.Start.Equal(w.start) || !g.End.Equal(w.end) || g.Immature != w.immature {
			t.Errorf("window %d = %+v, want %+v", i, g, w)
		}
	}
	if d := fixed[1].End.Sub(fixed[1].Start).Hours() / 24; d != 91 { // leap year
		t.Errorf("2024-Q1 spans %v days, want 91", d)
	}
	if trailing.Label != "trailing-365d" || trailing.Immature {
		t.Errorf("trailing = %+v", trailing)
	}
}

func TestCustomWindowAnchoredAtJanuaryFirst(t *testing.T) {
	fixed, _ := Windows(utc(2023, 12, 20, 0, 0, 0), utc(2024, 2, 20, 0, 0, 0), 30)
	got := ""
	for _, w := range fixed {
		got += fmt.Sprintf("%s[%s,%s) ", w.Label, w.Start.Format("01-02"), w.End.Format("01-02"))
	}
	want := "2023-30d-12[11-27,12-27) 2023-30d-13[12-27,01-01) 2024-30d-1[01-01,01-31) 2024-30d-2[01-31,03-01) "
	if got != want {
		t.Errorf("windows = %s\nwant      %s", got, want)
	}
	if !fixed[3].Immature || fixed[2].Immature {
		t.Errorf("immature flags wrong: %+v", fixed)
	}
}

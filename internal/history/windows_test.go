package history

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func utc(y int, m time.Month, d, h, min, s int) time.Time {
	return time.Date(y, m, d, h, min, s, 0, time.UTC)
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

func TestQuarterBucketingAcrossLeapDayAndImmaturity(t *testing.T) {
	f := newFixture(t)
	times := []time.Time{
		utc(2024, 2, 29, 23, 59, 59), // leap day: Q1
		utc(2024, 3, 31, 23, 59, 59), // last second of Q1
		utc(2024, 4, 1, 0, 0, 0),     // first second of Q2
		utc(2024, 5, 10, 12, 0, 0),
	}
	for i, ts := range times {
		f.commit(ts, "dev", "dev@example.com", map[string]string{"a/f.go": lines(i + 1)})
	}
	commits, _ := Load(context.Background(), f.dir)
	sigs := Compute(commits, Options{Scopes: scopesAB()[:1], End: utc(2024, 5, 15, 0, 0, 0), Snapshot: "s"})
	q1 := find(t, sigs, "scout.changes@1", "a", "2024-Q1")
	q2 := find(t, sigs, "scout.changes@1", "a", "2024-Q2")
	near(t, "Q1", q1.Value, 2)
	near(t, "Q2", q2.Value, 2)
	if q1.Detail["immature"] != false || q2.Detail["immature"] != true {
		t.Errorf("immature flags: Q1 %v, Q2 %v", q1.Detail["immature"], q2.Detail["immature"])
	}
	if q1.Detail["window_start"] != "2024-01-01T00:00:00Z" || q1.Detail["window_end"] != "2024-04-01T00:00:00Z" {
		t.Errorf("Q1 bounds %v..%v", q1.Detail["window_start"], q1.Detail["window_end"])
	}
	for _, s := range sigs {
		if err := s.Validate(); err != nil {
			t.Fatal(err)
		}
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

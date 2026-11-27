package stamp

import (
	"testing"
	"time"
)

func TestYearReadsTheInjectedClock(t *testing.T) {
	fixed := func() time.Time { return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC) }
	if Year(Options{Now: fixed}) != 2020 {
		t.Fatal("year")
	}
}

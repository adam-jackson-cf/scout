package report

import (
	"testing"
	"time"

	"canary/clock"
)

func TestStampUsesTheCurrentYear(t *testing.T) {
	clock.Now = func() time.Time { return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC) }
	if Stamp() != "2020" {
		t.Fatal("year")
	}
}

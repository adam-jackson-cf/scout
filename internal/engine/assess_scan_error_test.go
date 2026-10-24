package engine

import (
	"errors"
	"testing"

	"scout/internal/domain"
)

func TestScanBatchErrorWithoutFailureIsRecorded(t *testing.T) {
	adapter := domain.AdapterResult{}
	recordScanError(&adapter, errors.New("boom"))
	if len(adapter.Failures) != 1 || adapter.Failures[0].Code != failureScannerFailed {
		t.Fatalf("a scanner error must become a recorded failure: %+v", adapter.Failures)
	}
	own := domain.AdapterResult{Failures: []domain.Failure{{Code: "timed_out"}}}
	recordScanError(&own, errors.New("scanner timed out"))
	if len(own.Failures) != 1 || own.Failures[0].Code != "timed_out" {
		t.Fatalf("the scanner's own failure must be kept: %+v", own.Failures)
	}
	clean := domain.AdapterResult{}
	recordScanError(&clean, nil)
	if len(clean.Failures) != 0 {
		t.Fatalf("no error, no failure: %+v", clean.Failures)
	}
}

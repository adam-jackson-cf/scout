package main

import (
	"testing"
)

func TestSanitizedNeverReturnsEmptyForRejectedText(t *testing.T) {
	if got := sanitized("bad\x00text"); got == "" {
		t.Fatal("rejected text produced an empty message")
	}
	if got := sanitized("fine"); got != "fine" {
		t.Fatalf("sanitized(fine) = %q", got)
	}
}

package engine

import (
	"context"
	"os/exec"
	"testing"
	"time"
)

func TestTotalAfterAChildProcessRuns(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx, "true").Run(); err != nil {
		t.Fatal(err)
	}
	if Total([]int{1}) != 1 {
		t.Fatal("Total")
	}
}

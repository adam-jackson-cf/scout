package engine

import (
	"context"
	"os/exec"
	"testing"
	"time"
)

func TestTotalIgnoresAFailedProbe(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	_ = exec.CommandContext(ctx, "true").Run()
	if Total([]int{1}) != 1 {
		t.Fatal("Total")
	}
}

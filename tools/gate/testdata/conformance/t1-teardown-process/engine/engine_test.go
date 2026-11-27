package engine

import (
	"context"
	"os/exec"
	"testing"
	"time"
)

func TestTotalCleansUpWithAProcess(t *testing.T) {
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = exec.CommandContext(ctx, "true").Run()
	})
	if Total([]int{1}) != 1 {
		t.Fatal("Total")
	}
}

//go:build boundary

package runner

import (
	"context"
	"os/exec"
	"testing"
	"time"
)

func TestGitReportsItsVersion(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx, "git", "--version").Run(); err != nil {
		t.Fatal(err)
	}
}

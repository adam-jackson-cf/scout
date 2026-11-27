package compliant

import (
	"context"
	"os/exec"
	"time"
)

func Version() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, "git", "--version").Output()
}

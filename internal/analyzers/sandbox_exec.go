package analyzers

import (
	"context"
	"os/exec"
)

// SandboxAvailable reports whether the read-only scanner sandbox exists.
func SandboxAvailable() error {
	_, err := sandboxBinary()
	return err
}

// SandboxedCommand wraps binary in the same network-denied, write-confined
// sandbox that scanner runners use. Writes are permitted only beneath state.
// Callers set Env, Stdout and Stderr.
func SandboxedCommand(ctx context.Context, workDir, state, binary string, args []string) (*exec.Cmd, error) {
	command, err := sandboxedCommand(ctx, workDir, state, "", binary, args)
	if err != nil {
		return nil, err
	}
	command.Dir = workDir
	return command, nil
}

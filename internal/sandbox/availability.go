// Package sandbox runs external scanners in a network-denied, write-confined
// subprocess with a scrubbed environment and bounded output. It is the single
// owner of that security boundary for the analyzers and measures packages.
package sandbox

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"strconv"
)

// ErrUnavailable means the platform sandbox executable is not installed.
var ErrUnavailable = errors.New("read-only scanner isolation is unavailable")

func executable() (string, error) {
	var name string
	switch runtime.GOOS {
	case "darwin":
		name = "sandbox-exec"
	case "linux":
		name = "bwrap"
	default:
		return "", ErrUnavailable
	}
	path, err := exec.LookPath(name)
	if err != nil {
		return "", ErrUnavailable
	}
	return path, nil
}

// Available reports whether the read-only scanner sandbox exists.
func Available() error {
	_, err := executable()
	return err
}

// command wraps binary in the sandbox. Writes are permitted only beneath
// state; rulesRoot, when set, is mounted read-only. The caller sets Env and
// the output writers.
func command(ctx context.Context, workDir, state, rulesRoot, binary string, args []string) (*exec.Cmd, error) {
	sandbox, err := executable()
	if err != nil {
		return nil, err
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.CommandContext(ctx, sandbox, append([]string{"-p", darwinProfile(state, rulesRoot), binary}, args...)...)
	} else {
		commandArgs := []string{
			"--die-with-parent",
			"--unshare-net",
			"--ro-bind", "/", "/",
			"--bind", state, state,
		}
		if rulesRoot != "" {
			commandArgs = append(commandArgs, "--ro-bind", rulesRoot, rulesRoot)
		}
		commandArgs = append(commandArgs,
			"--unshare-pid",
			"--proc", "/proc",
			"--dev", "/dev",
			"--chdir", workDir,
			"--",
			binary,
		)
		cmd = exec.CommandContext(ctx, sandbox, append(commandArgs, args...)...)
	}
	cmd.Dir = workDir
	return cmd, nil
}

func darwinProfile(state, rulesRoot string) string {
	profile := "(version 1)\n" +
		"(allow default)\n" +
		"(deny network*)\n" +
		"(deny file-write* (require-not (subpath " + strconv.Quote(state) + ")))\n"
	if rulesRoot != "" {
		profile += "(deny file-write* (subpath " + strconv.Quote(rulesRoot) + "))\n"
	}
	return profile
}

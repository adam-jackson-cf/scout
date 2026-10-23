package analyzers

import (
	"context"
	"errors"
	"os"
	"os/exec"
)

// CheckPrerequisites verifies the declared executable, isolation runtime and
// exact scanner version without invoking analysis or writing assessed source.
func (r *Runner) CheckPrerequisites(ctx context.Context, root string) error {
	canonical, err := canonicalRoot(root)
	if err != nil {
		return errors.New("assessment root unavailable")
	}
	if _, err := sandboxBinary(); err != nil {
		return errors.New("read-only scanner sandbox unavailable")
	}
	binary, err := exec.LookPath(r.spec.Binary.Path)
	if err != nil {
		return errors.New("required scanner executable unavailable")
	}
	if err := externalScannerBinary(canonical, binary); err != nil {
		return err
	}
	state, err := newScannerState()
	if err != nil {
		return errors.New("scanner private state unavailable")
	}
	defer os.RemoveAll(state)
	if err := r.verifyVersion(ctx, canonical, state, "", binary); err != nil {
		return errors.New("required scanner version unavailable")
	}
	return nil
}

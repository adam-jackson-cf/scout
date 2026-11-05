package history

import (
	"context"
	"testing"
	"time"
)

// boundedTestContext keeps child processes within the per-test budget.
func boundedTestContext(t testing.TB) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

package main

import (
	"context"

	"scout/internal/engine"
)

// readinessMatrix prints REQ-18 for this repository: the Markdown table in
// text mode, the structured matrix under --json. It never fails a repository
// for a missing optional capability.
func readinessMatrix(ctx context.Context, flags *commonFlags, cfg engine.Config) int {
	matrix, err := engine.CapabilityMatrix(ctx, cfg, flags.root)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	if flags.json {
		return emit(flags, "complete", exitComplete, matrix, nil)
	}
	return printStdout(matrix.Render(), exitComplete)
}

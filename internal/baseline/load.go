package baseline

import (
	"context"
	"errors"

	"scout/internal/domain"
	"scout/internal/state"
)

// Base is a published baseline loaded read-only for a differential run. Issues
// include resolved ones so a reappearance can be recognised as regressed.
type Base struct {
	Issues []domain.Issue
	// Tip is the scout/state commit the baseline was read from.
	Tip string
	// Revision, ToolDigest and PolicyDigest come from the default-branch run
	// that produced the baseline; empty when the exporter had no finalized run.
	Revision     string
	ToolDigest   string
	PolicyDigest string
	// Uninspected lists the scanner families the baseline run did not fully
	// inspect; nil means the manifest carries no lane health (legacy export).
	Uninspected []state.UninspectedFamily
}

// LoadBase fetches the scout/state tip read-only. ok is false (with a nil
// error) when the remote has no scout/state branch, so nothing is introduced yet.
func LoadBase(ctx context.Context, opts Options) (Base, bool, error) {
	tree, tip, err := Fetch(ctx, opts)
	if errors.Is(err, ErrNoState) {
		return Base{}, false, nil
	}
	if err != nil {
		return Base{}, false, err
	}
	issues, err := state.DecodeIssues(tree)
	if err != nil {
		return Base{}, false, err
	}
	meta, err := state.ReadManifestMeta(tree[state.ExportManifest])
	if err != nil {
		return Base{}, false, err
	}
	return Base{Issues: issues, Tip: tip, Revision: meta.Revision, ToolDigest: meta.ToolDigest, PolicyDigest: meta.PolicyDigest, Uninspected: meta.Uninspected}, true, nil
}

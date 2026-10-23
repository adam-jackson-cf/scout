package engine

import (
	"context"
	"fmt"

	"scout/internal/history"
	"scout/internal/prmeta"
	"scout/internal/state"
)

// loadPRMetadata resolves the PR-metadata source. With no source it returns
// nothing (PR-level signals become `unmeasured`). A source that fails to load
// returns a note for lane health and no PRs: the signals stay `unmeasured`,
// never zero. Provenance of a loaded source is recorded in the store.
func loadPRMetadata(ctx context.Context, store *state.Store, o *PRMetadataOptions, snapshot string) ([]history.PR, *history.PRSource, string) {
	if o == nil || (o.File == "" && !o.Fetch.Enabled) {
		return nil, nil, ""
	}
	var (
		res prmeta.Result
		err error
	)
	if o.File != "" {
		res, err = prmeta.LoadFile(o.File)
	} else {
		res, _, err = prmeta.Fetch(ctx, o.Fetch)
	}
	if err != nil {
		return nil, nil, "PR metadata: " + err.Error()
	}
	if err := store.RecordPRMetadataSource(ctx, state.PRMetadataSource{
		Snapshot: snapshot, Kind: res.Source.Kind, Origin: res.Source.Origin, Digest: res.Source.Digest,
		FetchedAt: res.Source.FetchedAt, PRs: res.Source.Count,
	}); err != nil {
		return nil, nil, fmt.Sprintf("PR metadata provenance: %v", err)
	}
	return res.PRs, &res.Source, ""
}

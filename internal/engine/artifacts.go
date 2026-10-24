package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"scout/internal/artifacts"
	"scout/internal/domain"
	"scout/internal/measures"
	"scout/internal/state"
)

// ingestArtifacts loads the run's coverage and test-result artifacts, records
// the provenance of each for the snapshot and returns the coverage CRAP may
// use: non-nil when specs were given (with Unusable set when no artifact
// matches the snapshot revision) and nil when none were. The note describes
// ingestion problems for the measures lane health; a failure never fails the
// run and leaves scout.crap `unmeasured`.
func ingestArtifacts(ctx context.Context, store *state.Store, root string, snapshot domain.Snapshot, specs []artifacts.Spec) (*measures.Coverage, string) {
	if len(specs) == 0 {
		return nil, ""
	}
	result := artifacts.Load(root, specs, snapshot)
	var notes []string
	key := measures.SnapshotKey(snapshot)
	for _, artifact := range result.Artifacts {
		var payload []byte
		if artifact.Tests != nil {
			var err error
			if payload, err = json.Marshal(artifact.Tests); err != nil {
				notes = append(notes, fmt.Sprintf("artifacts: encode %s: %v", artifact.Source, err))
				continue
			}
		}
		if err := store.RecordIngestedArtifact(ctx, state.IngestedArtifact{
			Snapshot: key, Digest: artifact.Digest, Kind: artifact.Kind, Revision: artifact.Revision,
			Producer: artifact.Producer, Source: artifact.Source, RevisionMatch: artifact.RevisionMatch,
			Reason: artifact.Reason, Payload: payload,
		}); err != nil {
			notes = append(notes, fmt.Sprintf("artifacts: record %s: %v", artifact.Source, err))
		}
	}
	if len(notes) == 0 && len(result.Failures) > 0 {
		notes = append(notes, fmt.Sprintf("artifacts: %s: %v", result.Failures[0].Source, result.Failures[0].Err))
	}
	note := strings.Join(notes, "; ")
	merged := result.MatchedCoverage()
	return &measures.Coverage{Lines: merged.Lines, Digest: merged.Digest, Unusable: merged.Reason}, note
}

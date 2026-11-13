package shadow

import (
	"encoding/json"
	"scout/internal/domain"
	"testing"
)

// Shadow half of P05: pending/cross-lineage imports cannot adopt authority.
// Baseline's same-named proof owns actual CAS parent and whole-history prefix.
func TestADR012ImportWholeHistoryPrefix(t *testing.T) {
	for _, raw := range []string{
		`{"RepositoryID":"repo","ParentCommit":"parent","PublishingJobStartedAt":"2026-11-01T12:00:00Z","CurrentProtectedTip":"tip","Authenticated":false,"SourceReachable":true}`,
		`{"RepositoryID":"repo","ParentCommit":"parent","PublishingJobStartedAt":"2026-11-01T12:00:00Z","CurrentProtectedTip":"tip","Authenticated":true,"SourceReachable":false}`,
		`{"RepositoryID":"repo","ParentCommit":"claimed-parent","PublishingJobStartedAt":"2026-11-01T12:00:00Z","CurrentProtectedTip":"tip","Authenticated":true,"SourceReachable":true,"ChainProof":{"RepositoryID":"repo","DestinationTip":"actual-parent","Authenticated":true,"DestinationAncestor":true}}`,
	} {
		var auth PublicationAuthorization
		if err := json.Unmarshal([]byte(raw), &auth); err != nil {
			t.Fatal(err)
		}
		batch := domain.ShadowBatch{Inputs: domain.ShadowInputs{PendingImports: []domain.ShadowPendingImport{{ID: "pending", RepositoryID: "repo", Scope: domain.ShadowScopeCI, SourceLineage: "foreign-genesis"}}}}
		if _, err := PreparePublication(domain.ShadowLedger{}, batch, auth); err == nil {
			t.Fatalf("cross-lineage authority accepted: %s", raw)
		}
	}
}

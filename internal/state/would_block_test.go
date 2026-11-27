package state

import (
	"context"
	"testing"

	"scout/internal/domain"
)

func TestLegacyWouldBlockIsIdempotentAndPreservesExportTriage(t *testing.T) {
	ctx := context.Background()
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer closeOrFail(t, store)
	row := domain.WouldBlock{ID: "wb-1", HeadRevision: "h", BaseRevision: "b", IssueID: "issue:security:1", CheckID: "security.secrets.github-pat", Attribution: domain.AttributionIntroduced, Profile: "report"}
	for range 2 {
		if err := store.RecordWouldBlock(ctx, row); err != nil {
			t.Fatal(err)
		}
	}
	// Simulate a historical triage value. Production has no mutable triage API.
	if _, err := store.db.ExecContext(ctx, `UPDATE would_block SET triage=?,triage_actor='alice',triage_reason='real token',triaged_at=1 WHERE id='wb-1'`, domain.TriageWarranted); err != nil {
		t.Fatal(err)
	}
	if err := store.RecordWouldBlock(ctx, row); err != nil {
		t.Fatal(err)
	}
	checkLegacyTriagePreserved(t, store)
}

func checkLegacyTriagePreserved(t *testing.T, store *Store) {
	t.Helper()
	rows, err := store.ListLegacyWouldBlock(context.Background())
	if err != nil || len(rows) != 1 {
		t.Fatalf("rows = %d, %v; want one idempotent row", len(rows), err)
	}
	if rows[0].Triage != domain.TriageWarranted || rows[0].TriageActor != "alice" || rows[0].TriagedAt == nil {
		t.Fatalf("re-recording clobbered triage: %+v", rows[0])
	}
}

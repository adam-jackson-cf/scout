package state

import (
	"context"
	"testing"

	"scout/internal/domain"
)

func TestWouldBlockIsIdempotentAndKeepsMaintainerTriage(t *testing.T) {
	ctx := context.Background()
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	row := domain.WouldBlock{ID: "wb-1", HeadRevision: "h", BaseRevision: "b", IssueID: "issue:security:1", CheckID: "security.secrets.github-pat", Attribution: domain.AttributionIntroduced, Profile: "report"}
	for range 2 {
		if err := store.RecordWouldBlock(ctx, row); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.TriageWouldBlock(ctx, "wb-1", domain.TriageWarranted, "alice", "real token"); err != nil {
		t.Fatal(err)
	}
	if err := store.RecordWouldBlock(ctx, row); err != nil {
		t.Fatal(err)
	}
	rows, err := store.ListWouldBlock(ctx)
	if err != nil || len(rows) != 1 {
		t.Fatalf("rows = %d, %v; want one idempotent row", len(rows), err)
	}
	if rows[0].Triage != domain.TriageWarranted || rows[0].TriageActor != "alice" || rows[0].TriagedAt == nil {
		t.Fatalf("re-recording clobbered triage: %+v", rows[0])
	}
	if err := store.TriageWouldBlock(ctx, "missing", domain.TriageWarranted, "alice", ""); err == nil {
		t.Fatal("triage of an unknown row must fail")
	}
	if err := store.TriageWouldBlock(ctx, "wb-1", "maybe", "alice", ""); err == nil {
		t.Fatal("invalid triage value accepted")
	}
}

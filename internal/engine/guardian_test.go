package engine

import (
	"context"
	"strings"
	"testing"
	"time"

	"scout/internal/analyzers"
	"scout/internal/domain"
	"scout/internal/state"
)

func TestGuardianPersistsZeroEventAndFailedAttempts(t *testing.T) {
	for _, failed := range []bool{false, true} {
		name := "zero_event"
		if failed {
			name = "authenticated_input_failure"
		}
		t.Run(name, func(t *testing.T) {
			root, base := deltaRepo(t, map[string]string{"a.py": "x = 1\n"})
			binary := fixtureClawpatch(t, root)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			raw, err := git(ctx, root, "rev-parse", base+"^{tree}")
			if err != nil {
				t.Fatal(err)
			}
			at := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
			f := domain.PRRunFact{ID: domain.AttemptID{RepositoryID: "repo", RunID: "1000", JobID: "2000", Attempt: "1"}, Scope: domain.ShadowScopeCI, PRNumber: "17", HeadCommit: base, HeadTree: strings.TrimSpace(string(raw)), BaseCommit: base, MergeBase: base, AssessedRevision: base, CheckID: "guardrails.config", RuleVersion: "v1", CreatedAt: at, StartedAt: at, FinishedAt: at.Add(time.Minute), Pins: domain.ShadowPins{Scout: "scout", Tools: "tools", Catalogue: "catalogue", Policy: "policy", Recipe: "guardian"}}
			if failed {
				f.BaseCommit = "absent-base"
			}
			report, err := GuardianAssessment(ctx, DiffConfig{Config: Config{Root: root, ClawpatchBinary: binary, Registry: analyzers.NewRegistry()}, OverrideActor: "alice", OverrideText: "scout-override"}, f)
			if (err != nil) != failed {
				t.Fatalf("report=%+v error=%v", report, err)
			}
			store, err := state.Open(root)
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			ledger, err := store.ReadShadowLedger(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if len(ledger.Facts) != 1 || len(ledger.Events) != 0 || ledger.Facts[0].RecordedAt != nil || ledger.Facts[0].Pins != f.Pins {
				t.Fatalf("attempt not durably staged: %+v", ledger)
			}
			if failed && (ledger.Facts[0].Comparison != "none" || ledger.Facts[0].Inspection != "not_inspected") {
				t.Fatalf("failed attempt fabricated coverage: %+v", ledger.Facts[0])
			}
		})
	}
}

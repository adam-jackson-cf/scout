package baseline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"scout/internal/domain"
	"scout/internal/state"
)

// ShadowImportPlan separates verbatim restoration from material that still needs
// destination authorization and CAS publication. Pending retains source provenance.
type ShadowImportPlan struct {
	BytePreserving bool
	Adopt          domain.ShadowLedger
	Pending        domain.ShadowLedger
}

// A missing pure planner, not a provider/CAS scaffold, is the intended red.
var adr012S2Planner = PlanShadowImport

// PlanShadowImport is pure. A fresh restore or an authenticated whole-history
// fast-forward retains the original source exactly. Every other source remains
// pending; this function cannot authorize its authors or mint destination stamps.
func PlanShadowImport(dst, src domain.ShadowLedger, proof domain.ChainProof) (ShadowImportPlan, error) {
	if err := validatePlanHistories(dst, src); err != nil {
		return ShadowImportPlan{}, err
	}
	if emptyShadowLedger(dst) || (chainProven(dst, src, proof) && state.ShadowHistoryContains(dst, src)) {
		return ShadowImportPlan{BytePreserving: true, Adopt: src}, nil
	}
	return ShadowImportPlan{Pending: src}, nil
}

func validatePlanHistories(dst, src domain.ShadowLedger) error {
	// A store may have final material without a publication chain. Validate its
	// graph as staged material, but never use its claimed stamps as lineage proof.
	destination := dst
	if len(dst.Publications) == 0 {
		destination = withoutClaimedStamps(dst)
	}
	if err := state.ValidateShadowLedger(destination); err != nil {
		return fmt.Errorf("destination history: %w", err)
	}
	if err := state.ValidateShadowLedger(src); err != nil {
		return fmt.Errorf("source history: %w", err)
	}
	return state.ValidateShadowOverlap(dst, src)
}

func withoutClaimedStamps(dst domain.ShadowLedger) domain.ShadowLedger {
	destination := dst
	destination.Facts = append([]domain.PRRunFact(nil), dst.Facts...)
	destination.Events = append([]domain.ProviderPREvent(nil), dst.Events...)
	for i := range destination.Facts {
		destination.Facts[i].RecordedAt = nil
	}
	for i := range destination.Events {
		destination.Events[i].RecordedAt = nil
	}
	return destination
}

func chainProven(dst, src domain.ShadowLedger, proof domain.ChainProof) bool {
	return chainProofComplete(proof) && len(dst.Publications) != 0 &&
		ledgerInRepository(dst, proof.RepositoryID) && ledgerInRepository(src, proof.RepositoryID)
}

func chainProofComplete(proof domain.ChainProof) bool {
	return proof.Authenticated && proof.DestinationAncestor && proof.RepositoryID != "" && proof.StateRef != "" && proof.GenesisCommit != "" && proof.SourceTip != "" && proof.DestinationTip != ""
}

func ledgerInRepository(l domain.ShadowLedger, repositoryID string) bool {
	for _, p := range l.Publications {
		if p.RepositoryID != repositoryID {
			return false
		}
	}
	for _, f := range l.Facts {
		if f.ID.RepositoryID != repositoryID || f.RecordedAt == nil {
			return false
		}
	}
	for _, e := range l.Events {
		if e.Scope == domain.ShadowScopeCI && e.RepositoryID != repositoryID {
			return false
		}
	}
	return true
}

func adr012PlanLedger() domain.ShadowLedger {
	at := time.Date(2026, 11, 8, 12, 0, 0, 123000000, time.UTC)
	stamp := at.Add(time.Hour)
	fact := domain.PRRunFact{ID: domain.AttemptID{RepositoryID: "repository", RunID: "999", JobID: "1000", Attempt: "2"}, Scope: domain.ShadowScopeCI,
		PRNumber: "17", HeadCommit: "head", HeadTree: "tree", BaseCommit: "base", MergeBase: "merge-base", AssessedRevision: "merge-result", CheckID: "security:test", RuleVersion: "1",
		CreatedAt: at.Add(-time.Minute), StartedAt: at, FinishedAt: at.Add(time.Minute), Status: "assessed", Comparison: "comparable", Inspection: "inspected",
		Pins: domain.ShadowPins{Scout: "scout", Tools: "tools", Catalogue: "catalogue", Policy: "policy", Recipe: "guardian"}, RecordedAt: &stamp}
	event := domain.ProviderPREvent{EventID: "event", RepositoryID: "repository", Scope: domain.ShadowScopeCI, PRNumber: "17", HeadCommit: "head", HeadTree: "tree", MergeBase: "merge-base", CheckID: fact.CheckID, RuleVersion: "1", IssueID: "issue", EventAt: at, RecordedAt: &stamp}
	link := domain.PRRunEventLink{AttemptID: fact.ID, Scope: domain.ShadowScopeCI, EventID: event.EventID}
	triage := domain.ShadowTriage{ActionID: "999", EventID: event.EventID, RepositoryID: "repository", Scope: domain.ShadowScopeCI, Verdict: domain.TriageWarranted, Actor: "maintainer", PolicyRevision: "policy-tip", TriagedAt: at, RecordedAt: &stamp}
	frontier := domain.ShadowFrontier{RepositoryID: "repository", CheckID: fact.CheckID, RuleVersion: "1", CompleteThrough: at.Add(30 * time.Minute), PolicyRevision: "policy-tip", MinDays: 14, MinObservations: 5}
	p := domain.ShadowPublication{ID: "genesis", RepositoryID: "repository", RecordedAt: stamp, Facts: []domain.PRRunFact{fact}, Events: []domain.ProviderPREvent{event}, Links: []domain.PRRunEventLink{link}, Triage: []domain.ShadowTriage{triage}, Frontiers: []domain.ShadowFrontier{frontier}}
	return domain.ShadowLedger{Facts: p.Facts, Events: p.Events, Links: p.Links, Triage: p.Triage, Frontiers: p.Frontiers, Publications: []domain.ShadowPublication{p}}
}

func adr012PlanClone(t *testing.T, ledger domain.ShadowLedger) domain.ShadowLedger {
	t.Helper()
	raw, err := json.Marshal(ledger)
	if err != nil {
		t.Fatal(err)
	}
	var clone domain.ShadowLedger
	if err := json.Unmarshal(raw, &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}

type importPlanCase struct {
	name     string
	dst, src domain.ShadowLedger
	proof    domain.ChainProof
	adopt    bool
}

type importPlanFixture struct {
	proof domain.ChainProof
	dst   domain.ShadowLedger
	src   domain.ShadowLedger
	post  domain.PRRunFact
	stamp time.Time
}

func newImportPlanFixture(t *testing.T) importPlanFixture {
	t.Helper()
	proof := domain.ChainProof{RepositoryID: "repository", StateRef: "refs/heads/scout/state", GenesisCommit: "genesis-commit", SourceTip: "source-tip", DestinationTip: "destination-tip", Authenticated: true, DestinationAncestor: true}
	dst := adr012PlanLedger()
	src := adr012PlanClone(t, dst)
	// Source advances with post-C evidence: these rows are part of the prefix
	// comparison even though they do not alter the sealed frontier.
	post := src.Facts[0]
	post.ID.RunID = "1000"
	stamp := src.Publications[0].RecordedAt.Add(time.Millisecond)
	post.StartedAt = stamp.Add(-time.Second)
	post.FinishedAt = stamp
	post.RecordedAt = &stamp
	p := domain.ShadowPublication{ID: "next", ParentID: "genesis", RepositoryID: "repository", RecordedAt: stamp, Facts: []domain.PRRunFact{post}}
	src.Facts = append(src.Facts, post)
	src.Publications = append(src.Publications, p)
	return importPlanFixture{proof: proof, dst: dst, src: src, post: post, stamp: stamp}
}

func (f importPlanFixture) provenanceCases() []importPlanCase {
	dst, src, proof := f.dst, f.src, f.proof
	cases := []importPlanCase{
		{"fresh_restore_without_ancestry_preserves_bytes", domain.ShadowLedger{}, src, domain.ChainProof{}, true},
		{"authenticated_whole_history_fast_forward", dst, src, proof, true},
	}
	for _, name := range []string{"missing_authentication", "missing_ancestry", "missing_genesis", "fork_restored_chain"} {
		unproven := proof
		switch name {
		case "missing_authentication":
			unproven.Authenticated = false
		case "missing_ancestry":
			unproven.DestinationAncestor = false
		case "missing_genesis":
			unproven.GenesisCommit = ""
		case "fork_restored_chain":
			unproven.RepositoryID = "fork-repository"
		}
		cases = append(cases, importPlanCase{name, dst, src, unproven, false})
	}
	return cases
}

func (f importPlanFixture) divergedCases(t *testing.T) []importPlanCase {
	t.Helper()
	dst, src, proof, stamp := f.dst, f.src, f.proof, f.stamp
	diverged := adr012PlanClone(t, dst)
	correction := f.post
	correction.ID.RunID = "1001"
	correction.Inspection = "not_inspected"
	diverged.Facts = append(diverged.Facts, correction)
	diverged.Publications = append(diverged.Publications, domain.ShadowPublication{ID: "destination-only", ParentID: "genesis", RepositoryID: "repository", RecordedAt: stamp, Facts: []domain.PRRunFact{correction}})
	cases := []importPlanCase{
		{"destination_only_post_C_correction_is_not_prefix", diverged, src, proof, false},
		{"destination_only_post_C_triage_is_not_prefix", triageDivergedLedger(t, dst, stamp), src, proof, false},
		{"destination_only_frontier_is_not_prefix", frontierDivergedLedger(t, dst, stamp), src, proof, false},
	}
	noC := adr012PlanClone(t, diverged)
	noC.Frontiers = nil
	for i := range noC.Publications {
		noC.Publications[i].Frontiers = nil
	}
	cases = append(cases, importPlanCase{"nonempty_without_C_is_not_vacuous_prefix", noC, src, proof, false})
	// A destination-only pending final has no publication or C at all.
	pending := correction
	pending.RecordedAt = nil
	return append(cases, importPlanCase{"pending_no_C_history_is_not_fresh", domain.ShadowLedger{Facts: []domain.PRRunFact{pending}}, src, proof, false})
}

func triageDivergedLedger(t *testing.T, dst domain.ShadowLedger, stamp time.Time) domain.ShadowLedger {
	t.Helper()
	triageDiverged := adr012PlanClone(t, dst)
	correctionRow := triageDiverged.Triage[0]
	correctionRow.ActionID = "1000"
	correctionRow.Verdict = domain.TriageNotWarranted
	correctionRow.RecordedAt = &stamp
	triageDiverged.Triage = append(triageDiverged.Triage, correctionRow)
	triageDiverged.Publications = append(triageDiverged.Publications, domain.ShadowPublication{ID: "destination-triage", ParentID: "genesis", RepositoryID: "repository", RecordedAt: stamp, Triage: []domain.ShadowTriage{correctionRow}})
	return triageDiverged
}

func frontierDivergedLedger(t *testing.T, dst domain.ShadowLedger, stamp time.Time) domain.ShadowLedger {
	t.Helper()
	frontierDiverged := adr012PlanClone(t, dst)
	nextFrontier := frontierDiverged.Frontiers[0]
	nextFrontier.CompleteThrough = stamp
	frontierDiverged.Frontiers = append(frontierDiverged.Frontiers, nextFrontier)
	frontierDiverged.Publications = append(frontierDiverged.Publications, domain.ShadowPublication{ID: "destination-frontier", ParentID: "genesis", RepositoryID: "repository", RecordedAt: stamp, Frontiers: []domain.ShadowFrontier{nextFrontier}})
	return frontierDiverged
}

func TestADR012S2ImportPlan(t *testing.T) {
	fixture := newImportPlanFixture(t)
	cases := append(fixture.provenanceCases(), fixture.divergedCases(t)...)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { checkImportPlanCase(t, tc) })
	}
	t.Run("immutable_collision_and_dangling_link_fail_before_adoption", func(t *testing.T) {
		checkPlannerRejectsInvalidSource(t, fixture.src, fixture.proof)
	})
	t.Run("empty_destination_never_skips_manifest_validation", checkEmptyDestinationValidatesManifest)
}

func checkImportPlanCase(t *testing.T, tc importPlanCase) {
	t.Helper()
	beforeDst, _ := json.Marshal(tc.dst)
	beforeSrc, _ := json.Marshal(tc.src)
	var plan ShadowImportPlan
	var err error
	plan, err = adr012S2Planner(tc.dst, tc.src, tc.proof)
	if err != nil {
		t.Fatalf("valid history must be classified, not rejected: %v", err)
	}
	afterDst, _ := json.Marshal(tc.dst)
	afterSrc, _ := json.Marshal(tc.src)
	if !bytes.Equal(beforeDst, afterDst) || !bytes.Equal(beforeSrc, afterSrc) {
		t.Error("pure import planner mutated its inputs")
	}
	assertImportPlanPartition(t, tc, plan)
	replay, err := adr012S2Planner(tc.dst, tc.src, tc.proof)
	if err != nil || !reflect.DeepEqual(plan, replay) {
		t.Error("same history/proof produced nondeterministic import plan")
	}
}

func assertImportPlanPartition(t *testing.T, tc importPlanCase, plan ShadowImportPlan) {
	t.Helper()
	if plan.BytePreserving != tc.adopt {
		t.Errorf("byte-preserving adoption=%v, want %v", plan.BytePreserving, tc.adopt)
	}
	if tc.adopt {
		if !reflect.DeepEqual(plan.Adopt, tc.src) {
			t.Error("fresh/proven fast-forward did not adopt complete source history verbatim")
		}
		if !reflect.DeepEqual(plan.Pending, domain.ShadowLedger{}) {
			t.Error("byte-preserving adoption incorrectly staged cross-lineage history")
		}
	} else {
		if !reflect.DeepEqual(plan.Adopt, domain.ShadowLedger{}) {
			t.Error("cross-lineage plan adopted source authority/stamps")
		}
		if !reflect.DeepEqual(plan.Pending, tc.src) {
			t.Error("cross-lineage plan lost original immutable material/order/stamps pending S3 authorization")
		}
	}
}

func checkPlannerRejectsInvalidSource(t *testing.T, src domain.ShadowLedger, proof domain.ChainProof) {
	t.Helper()
	for _, kind := range []string{"conflicting_final", "dangling_link"} {
		t.Run(kind, func(t *testing.T) {
			bad := adr012PlanClone(t, src)
			if kind == "conflicting_final" {
				conflict := bad.Facts[0]
				conflict.Inspection = "not_inspected"
				bad.Facts = append(bad.Facts, conflict)
			} else {
				bad.Links[0].EventID = "absent"
			}
			if _, err := adr012S2Planner(domain.ShadowLedger{}, bad, proof); err == nil {
				t.Errorf("fresh destination skipped %s validation", kind)
			}
		})
	}
}

func checkEmptyDestinationValidatesManifest(t *testing.T) {
	invalid := Tree{state.ExportIssues: []byte("{}\n")}
	if _, err := MergeState(nil, invalid); err == nil {
		t.Error("empty destination accepted unvalidated local state")
	}
	for _, schema := range []string{"scout.state/v1", "scout.state/v3"} {
		files := Tree{state.ExportIssues: nil, state.ExportDispositions: nil, "state/v1/pr-runs.jsonl": nil, "state/v1/would-block-triage.jsonl": nil}
		files[state.ExportManifest] = state.RenderManifest(state.ManifestMeta{}, files)
		var fields map[string]any
		if err := json.Unmarshal(files[state.ExportManifest], &fields); err != nil {
			t.Fatal(err)
		}
		fields["schema"] = schema
		files[state.ExportManifest], _ = state.CanonicalLine(fields)
		if _, err := MergeState(nil, files); err == nil {
			t.Errorf("empty destination bypassed %s compatibility gate", schema)
		}
		if _, err := MergeState(files, nil); err == nil {
			t.Errorf("empty local tree bypassed current %s compatibility gate", schema)
		}
	}
}

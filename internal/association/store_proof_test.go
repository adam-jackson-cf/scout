package association

import (
	"context"
	"reflect"
	"testing"
	"time"

	"scout/internal/areas"
	"scout/internal/domain"
	"scout/internal/state"
)

func proofStore(t *testing.T) *state.Store {
	t.Helper()
	s, err := state.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}
func proofRecord(id, path string) domain.Association {
	return domain.Association{ID: id, RunID: "prior", DecidedAt: time.Unix(100, 0).UTC(), MergeCommit: "merge", ClusterKey: path, Files: []string{path}, Outcome: domain.AssociationAssigned, AreaIDs: []string{"a"}, PRMetadata: domain.PRMetadataAbsent, ThresholdsProvisional: true}
}
func proofFingerprint() Fingerprint {
	return Fingerprint{CoveredScore: .60, TChoice: .535, CoveredQuestion: "covered-v1", ChoiceQuestion: "choice-v1", JevModel: "typesafe/jev-1.14-20261014"}
}

func areaFixtures(values ...areas.Area) []areas.Area {
	return values
}

func ownershipFeatures() []domain.Feature {
	return []domain.Feature{{Source: "heuristic", OwnedFiles: []domain.FeatureFileRef{{Path: "heuristic.go"}}}, {Source: areas.AgentSource, OwnedFiles: []domain.FeatureFileRef{{Path: "agent-only.go"}}}}
}
func staleProof(t *testing.T, change func(*Fingerprint), want bool) {
	t.Helper()
	ctx := context.Background()
	s := proofStore(t)
	old := proofFingerprint()
	if err := s.ReplaceAreas(ctx, []areas.Area{{ScoutAreaID: "a", Origin: areas.OriginAgent, FileSet: []string{"mapped.go"}}}); err != nil {
		t.Fatal(err)
	}
	a := proofRecord("prior", "associated.go")
	a.DecisionDigest = old.Digest()
	if err := s.InsertAssociations(ctx, []domain.Association{a}); err != nil {
		t.Fatal(err)
	}
	next := old
	change(&next)
	if err := MarkStale(ctx, s, next); err != nil {
		t.Fatal(err)
	}
	rows, err := s.ListAssociations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	expected := a
	expected.Stale = want
	if len(rows) != 1 || !reflect.DeepEqual(rows[0], expected) {
		t.Fatalf("staleness must preserve row/label: got %+v want %+v", rows, expected)
	}
	own, err := NewStoreOwnership(ctx, s, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := own.AreaOwners("associated.go"); !reflect.DeepEqual(got, []string{"a"}) {
		t.Fatalf("stale labels lost: %v", got)
	}
	// Applying the same fingerprint again never clears a stale label or adds rows.
	if err := MarkStale(ctx, s, next); err != nil {
		t.Fatal(err)
	}
	rows, err = s.ListAssociations(ctx)
	if err != nil || len(rows) != 1 || rows[0].Stale != want {
		t.Fatalf("repeat mutation rows=%+v err=%v", rows, err)
	}
}
func TestAssociationStaleCoveredScore(t *testing.T) {
	staleProof(t, func(f *Fingerprint) { f.CoveredScore = .61 }, true)
}
func TestAssociationStaleChoiceThreshold(t *testing.T) {
	staleProof(t, func(f *Fingerprint) { f.TChoice = .536 }, true)
}
func TestAssociationStaleQuestionText(t *testing.T) {
	t.Run("covered", func(t *testing.T) {
		staleProof(t, func(f *Fingerprint) { f.CoveredQuestion = "changed-covered-text" }, true)
	})
	t.Run("choice", func(t *testing.T) {
		staleProof(t, func(f *Fingerprint) { f.ChoiceQuestion = "changed-choice-text" }, true)
	})
}
func TestAssociationStaleJevModel(t *testing.T) {
	staleProof(t, func(f *Fingerprint) { f.JevModel = "typesafe/jev-new-pin" }, true)
}
func TestAssociationCatalogueDoesNotStale(t *testing.T) {
	// Catalogue content lives outside Fingerprint and changes independently.
	ctx := context.Background()
	s := proofStore(t)
	f := proofFingerprint()
	a := proofRecord("prior", "associated.go")
	a.DecisionDigest = f.Digest()
	a.CatalogueDigest = "old-catalogue"
	if err := s.InsertAssociations(ctx, []domain.Association{a}); err != nil {
		t.Fatal(err)
	}
	if err := s.ReplaceAreas(ctx, areaFixtures(areas.Area{ScoutAreaID: "a", Title: "Changed title", Origin: areas.OriginAgent, FileSet: []string{"new-sample.go"}}, areas.Area{ScoutAreaID: "b", Title: "New option", Origin: areas.OriginAgent})); err != nil {
		t.Fatal(err)
	}
	if err := MarkStale(ctx, s, f); err != nil {
		t.Fatal(err)
	}
	rows, err := s.ListAssociations(ctx)
	if err != nil || len(rows) != 1 || !reflect.DeepEqual(rows[0], a) {
		t.Fatalf("catalogue-only change mutated record: %+v %v", rows, err)
	}
}

func TestAssociationStoreOwnership(t *testing.T) {
	ctx := context.Background()
	s := proofStore(t)
	aa := areaFixtures(
		areas.Area{ScoutAreaID: "a", Origin: areas.OriginAgent, FileSet: []string{"mapped.go"}},
		areas.Area{ScoutAreaID: "b", Origin: areas.OriginAgent, FileSet: []string{"other.go"}},
		areas.Area{ScoutAreaID: "pending", Origin: "association", Status: areas.StatusPending, FileSet: []string{"pool.go"}},
		areas.Area{ScoutAreaID: "superseded", Origin: areas.OriginAgent, Status: areas.StatusSuperseded, FileSet: []string{"superseded.go"}},
		areas.Area{ScoutAreaID: "retired", Origin: areas.OriginAgent, RetiredSnapshot: "old", FileSet: []string{"retired.go"}},
	)
	if err := s.ReplaceAreas(ctx, aa); err != nil {
		t.Fatal(err)
	}
	before, err := s.ListAreas(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assigned := proofRecord("assigned", "svc/item.py")
	assigned.Stale = true
	inherited := proofRecord("inherited", "svc/item_alt.py")
	inherited.Outcome = domain.AssociationInherited
	inherited.AreaIDs = nil
	inherited.InheritedAreaIDs = []string{"b"}
	inherited.InheritedRule = domain.InheritedVariant
	inherited.InheritedFrom = "other.go"
	dead := proofRecord("dead", "dead.go")
	dead.Superseded = true
	invalid := proofRecord("inactive-target", "inactive.go")
	invalid.AreaIDs = []string{"pending", "retired", "superseded"}
	if err := s.InsertAssociations(ctx, []domain.Association{assigned, inherited, dead, invalid}); err != nil {
		t.Fatal(err)
	}
	features := ownershipFeatures()
	own, err := NewStoreOwnership(ctx, s, features)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, path string
		want       []string
	}{
		{name: "mapped owner", path: "mapped.go", want: []string{"a"}},
		{name: "stale assigned owner", path: "svc/item.py", want: []string{"a"}},
		{name: "inherited owner", path: "svc/item_alt.py", want: []string{"b"}},
		{name: "pending is not owner", path: "pool.go"},
		{name: "superseded area is not owner", path: "superseded.go"},
		{name: "retired area is not owner", path: "retired.go"},
		{name: "superseded assignment is not owner", path: "dead.go"},
		{name: "inactive assignment targets are not owners", path: "inactive.go"},
	} {
		if got := own.AreaOwners(tc.path); len(got) != len(tc.want) || (len(got) > 0 && !reflect.DeepEqual(got, tc.want)) {
			t.Errorf("AreaOwners(%s)=%v want %v", tc.path, got, tc.want)
		}
	}
	if got := own.PendingAreas("pool.go"); !reflect.DeepEqual(got, []string{"pending"}) {
		t.Fatalf("pending pool %v", got)
	}
	if len(own.PendingAreas("svc/item.py")) != 0 || !own.HeuristicOnly("heuristic.go") || own.HeuristicOnly("agent-only.go") {
		t.Fatal("pool/heuristic distinctions lost")
	}
	got := own.SiblingFiles(SiblingKeyOf("svc/item.py"))
	found := false
	for _, p := range got {
		if p == "svc/item.py" {
			found = true
		}
	}
	if !found {
		t.Fatalf("association-owned sibling omitted: %v", got)
	}
	// Neither labels nor derived ownership are written into agent-map FileSets.
	after, err := s.ListAreas(ctx)
	if err != nil || !reflect.DeepEqual(after, before) {
		t.Fatalf("ownership mutated catalogue: %+v %v", after, err)
	}
}

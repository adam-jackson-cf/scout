package association

import (
	"context"
	"testing"

	"scout/internal/domain"
)

func TestTerraAssignmentVisibleToNextMerge(t *testing.T) {
	root, s, o, _ := terraLaneFixture(t, 1, "none")
	write(t, root, "service0/handler_an.py", "def serve_variant(request):\n return request\n")
	gitIn(t, root, "add", "service0/handler_an.py")
	gitIn(t, root, "commit", "-qm", "variant after assignment")
	o.Head = gitIn(t, root, "rev-parse", "HEAD")
	dir := fakeTerra(t, terraExisting(), terraDiagnostic(), "")
	enableTerraGit(t, dir)
	_, rows := executeTerraLane(t, root, s, o)
	assigned, inherited := 0, 0
	for _, r := range rows {
		if r.Outcome == domain.AssociationAssigned {
			assigned++
		}
		if r.Outcome == domain.AssociationInherited && r.InheritedRule == domain.InheritedVariant {
			inherited++
		}
	}
	if terraCallCount(t, dir) != 1 || assigned != 1 || inherited != 1 {
		t.Fatalf("calls=%d assigned=%d inherited=%d rows=%+v", terraCallCount(t, dir), assigned, inherited, rows)
	}
}
func TestTerraPendingVisibleToNextMerge(t *testing.T) {
	root, s, o, _ := terraLaneFixture(t, 1, "none")
	gitIn(t, root, "mv", "service0/handler.py", "service0/renamed.py")
	gitIn(t, root, "commit", "-qm", "rename after pending")
	o.Head = gitIn(t, root, "rev-parse", "HEAD")
	dir := fakeTerra(t, terraNew(), terraDiagnostic(), "")
	enableTerraGit(t, dir)
	_, rows := executeTerraLane(t, root, s, o)
	if terraCallCount(t, dir) != 1 || len(rows) != 1 || rows[0].Outcome != domain.AssociationPendingNew {
		t.Fatalf("calls=%d rows=%+v", terraCallCount(t, dir), rows)
	}
	aa, err := s.ListAreas(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, a := range aa {
		if a.Status != "pending" {
			continue
		}
		for _, p := range a.FileSet {
			if p == "service0/handler.py" {
				t.Fatal("pending rename retained old path")
			}
			if p == "service0/renamed.py" {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("pending rename not propagated: %+v", aa)
	}
}

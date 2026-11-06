package state

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"scout/internal/areas"
)

func pendingActionFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	s, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	list := make([]areas.Area, 0)
	list = append(list, areas.Area{ScoutAreaID: "pending", Status: areas.StatusPending, Origin: "association", FileSet: append([]string(nil), "p.go"), FirstSeenSnapshot: "s1"})
	list = append(list, areas.Area{ScoutAreaID: "target", Origin: areas.OriginAgent, FileSet: append([]string(nil), "t.go"), FirstSeenSnapshot: "s1"})
	if err := s.ReplaceAreas(context.Background(), list); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	return root
}
func TestPendingAreaReadOnlyRefuses(t *testing.T) {
	root := pendingActionFixture(t)
	s, err := OpenExistingReadOnly(root)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	before, err := s.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.AcceptArea(ctx, "pending", "maintainer", "reviewed"); err == nil || !strings.Contains(err.Error(), "read-only") {
		t.Fatalf("read-only accept: %v", err)
	}
	if _, err := s.MergeArea(ctx, "pending", "target", "maintainer", "same concern"); err == nil || !strings.Contains(err.Error(), "read-only") {
		t.Fatalf("read-only merge: %v", err)
	}
	after, err := s.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("read-only actions mutated rows")
	}
}

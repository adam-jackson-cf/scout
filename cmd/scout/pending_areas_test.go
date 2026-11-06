package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"scout/internal/areas"
	"scout/internal/domain"
	"scout/internal/engine"
	"scout/internal/state"
)

func pendingPaths(paths ...string) []string { return paths }
func pendingArgs(args ...string) []string   { return args }
func pendingFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	store, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	list := make([]areas.Area, 0)
	list = append(list, areas.Area{ScoutAreaID: "proposal", Title: "Proposed concern", Status: areas.StatusPending, Origin: "association", FileSet: pendingPaths("pkg/new.go", "pkg/other.go"), FirstSeenSnapshot: "s1", ClusterID: "cluster", DecisionID: "decision"})
	list = append(list, areas.Area{ScoutAreaID: "remaining", Title: "Another concern", Status: areas.StatusPending, Origin: "association", FileSet: pendingPaths("pkg/remaining.go"), FirstSeenSnapshot: "s1"})
	list = append(list, areas.Area{ScoutAreaID: "target", Title: "Existing concern", Origin: areas.OriginAgent, FileSet: pendingPaths("pkg/target.go"), FirstSeenSnapshot: "s1"})
	list = append(list, areas.Area{ScoutAreaID: "superseded", Status: areas.StatusSuperseded, Origin: "association", FileSet: pendingPaths("pkg/resolved.go"), FirstSeenSnapshot: "s1"})
	list = append(list, areas.Area{ScoutAreaID: "retired", RetiredSnapshot: "s2", Origin: areas.OriginAgent, FileSet: pendingPaths("pkg/retired.go"), FirstSeenSnapshot: "s1"})
	if err := store.ReplaceAreas(context.Background(), list); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	return root
}
func pendingStore(t *testing.T, root string) *state.Store {
	t.Helper()
	s, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}
func pendingFind(t *testing.T, list []areas.Area, id string) areas.Area {
	t.Helper()
	for _, a := range list {
		if a.ScoutAreaID == id {
			return a
		}
	}
	t.Fatalf("missing area %q: %+v", id, list)
	return areas.Area{}
}
func pendingRun(t *testing.T, args ...string) (outcome, string) {
	t.Helper()
	text, code := captureStdout(t, func() int { return run(context.Background(), args) })
	var reply outcome
	if err := json.Unmarshal([]byte(text), &reply); err != nil {
		t.Fatalf("command %v did not emit envelope (exit %d): %q: %v", args, code, text, err)
	}
	if reply.ExitCode != code {
		t.Fatalf("envelope exit %d != command exit %d", reply.ExitCode, code)
	}
	return reply, text
}
func pendingRoundTrip(t *testing.T, root string) {
	t.Helper()
	ctx := context.Background()
	dir := t.TempDir()
	if code := run(ctx, pendingArgs("export", "--root", root, "--out", dir)); code != 0 {
		t.Fatalf("export exit %d", code)
	}
	target := t.TempDir()
	if code := run(ctx, pendingArgs("import", "--root", target, "--from", dir)); code != 0 {
		t.Fatalf("import exit %d", code)
	}
	again := t.TempDir()
	if code := run(ctx, pendingArgs("export", "--root", target, "--out", again)); code != 0 {
		t.Fatalf("re-export exit %d", code)
	}
	first := pendingStore(t, root)
	second := pendingStore(t, target)
	a, err := first.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	b, err := second.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatal("maintainer action was not export/import byte-stable")
	}
	firstAreas, err := first.ListAreas(ctx)
	if err != nil {
		t.Fatal(err)
	}
	secondAreas, err := second.ListAreas(ctx)
	if err != nil || !reflect.DeepEqual(firstAreas, secondAreas) {
		t.Fatalf("restored areas differ: %+v %v", secondAreas, err)
	}
	firstActions, err := first.ListAreaActions(ctx)
	if err != nil {
		t.Fatal(err)
	}
	secondActions, err := second.ListAreaActions(ctx)
	if err != nil || !reflect.DeepEqual(firstActions, secondActions) {
		t.Fatalf("restored audit differs: %+v %v", secondActions, err)
	}
	firstAssociations, err := first.ListAssociations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	secondAssociations, err := second.ListAssociations(ctx)
	if err != nil || !reflect.DeepEqual(firstAssociations, secondAssociations) {
		t.Fatalf("restored labels differ: %+v %v", secondAssociations, err)
	}
	entries, err := os.ReadDir(filepath.Join(dir, "state", "v1"))
	if err != nil || len(entries) == 0 {
		t.Fatalf("export files: %v %v", entries, err)
	}
}
func TestPendingAreasList(t *testing.T) {
	root := pendingFixture(t)
	reply, text := pendingRun(t, "areas", "list", "--pending", "--root", root, "--json")
	if reply.ExitCode != 0 || reply.Status != "complete" {
		t.Fatalf("list: %s", text)
	}
	var decoded struct {
		Result []areas.Area `json:"result"`
	}
	if err := json.Unmarshal([]byte(text), &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Result) != 2 || decoded.Result[0].ScoutAreaID != "proposal" || !reflect.DeepEqual(decoded.Result[0].FileSet, pendingPaths("pkg/new.go", "pkg/other.go")) || decoded.Result[1].ScoutAreaID != "remaining" {
		t.Fatalf("pending list: %s", text)
	}
	plain, code := captureStdout(t, func() int {
		return run(context.Background(), pendingArgs("areas", "list", "--pending", "--root", root))
	})
	if code != 0 || !strings.Contains(plain, "proposal") || !strings.Contains(plain, "pkg/new.go") || !strings.Contains(plain, "pkg/other.go") || strings.Contains(plain, "pkg/target.go") {
		t.Fatalf("text pending list: exit %d %s", code, plain)
	}
}
func TestPendingAreaAccept(t *testing.T) {
	root := pendingFixture(t)
	before := time.Now().UTC()
	reply, text := pendingRun(t, "areas", "accept", "proposal", "--root", root, "--actor", "maintainer", "--reason", "reviewed concern", "--authorize", "--json")
	if reply.ExitCode != 0 {
		t.Fatalf("accept: %s", text)
	}
	store := pendingStore(t, root)
	list, err := store.ListAreas(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	got := pendingFind(t, list, "proposal")
	if !got.Active() || got.Status != areas.StatusActive || got.Origin != "association" || got.AcceptedBy != "maintainer" || got.AcceptedAt == nil || got.AcceptedAt.Before(before) || got.AcceptedAt.After(time.Now().UTC()) {
		t.Fatalf("acceptance: %+v", got)
	}
	actions, err := store.ListAreaActions(context.Background())
	if err != nil || len(actions) != 1 || actions[0].Action != "accept" || actions[0].AreaID != "proposal" || actions[0].Actor != "maintainer" || actions[0].Reason != "reviewed concern" || actions[0].At.IsZero() {
		t.Fatalf("accept audit: %+v %v", actions, err)
	}
	store.Close()
	_, pending := pendingRun(t, "areas", "list", "--pending", "--root", root, "--json")
	if strings.Contains(pending, "proposal") || !strings.Contains(pending, "remaining") {
		t.Fatalf("accepted area still pending: %s", pending)
	}
	pendingRoundTrip(t, root)
}
func TestPendingAreaMerge(t *testing.T) {
	root := pendingFixture(t)
	reply, text := pendingRun(t, "areas", "merge", "proposal", "target", "--root", root, "--actor", "maintainer", "--reason", "same concern", "--authorize", "--json")
	if reply.ExitCode != 0 {
		t.Fatalf("merge: %s", text)
	}
	store := pendingStore(t, root)
	ctx := context.Background()
	list, err := store.ListAreas(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got := pendingFind(t, list, "proposal"); got.Status != areas.StatusSuperseded || got.Active() || got.AcceptedAt != nil {
		t.Fatalf("merged pending: %+v", got)
	}
	if got := pendingFind(t, list, "target"); !reflect.DeepEqual(got.FileSet, pendingPaths("pkg/target.go")) {
		t.Fatalf("merge polluted map FileSet: %+v", got)
	}
	records, err := store.ListAssociations(ctx)
	if err != nil || len(records) != 1 || records[0].Outcome != domain.AssociationAssigned || !reflect.DeepEqual(records[0].AreaIDs, pendingPaths("target")) || !reflect.DeepEqual(records[0].Files, pendingPaths("pkg/new.go", "pkg/other.go")) || records[0].Reason != "maintainer_merge" || records[0].Superseded {
		t.Fatalf("merge labels: %+v %v", records, err)
	}
	actions, err := store.ListAreaActions(ctx)
	if err != nil || len(actions) != 1 || actions[0].Action != "merge" || actions[0].AreaID != "proposal" || actions[0].TargetAreaID != "target" || actions[0].Actor != "maintainer" || actions[0].Reason != "same concern" || actions[0].At.IsZero() {
		t.Fatalf("merge audit: %+v %v", actions, err)
	}
	store.Close()
	_, pending := pendingRun(t, "areas", "list", "--pending", "--root", root, "--json")
	if strings.Contains(pending, "proposal") {
		t.Fatalf("merged area listed pending: %s", pending)
	}
	pendingRoundTrip(t, root)
}
func TestPendingAreaCommandRejections(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"missing_authorize", pendingArgs("accept", "proposal", "--actor", "m", "--reason", "r")},
		{"missing_actor", pendingArgs("accept", "proposal", "--authorize", "--reason", "r")},
		{"missing_reason", pendingArgs("accept", "proposal", "--authorize", "--actor", "m")},
		{"unknown_pending", pendingArgs("accept", "unknown", "--authorize", "--actor", "m", "--reason", "r")},
		{"active_not_pending", pendingArgs("accept", "target", "--authorize", "--actor", "m", "--reason", "r")},
		{"superseded_not_pending", pendingArgs("accept", "superseded", "--authorize", "--actor", "m", "--reason", "r")},
		{"merge_missing_authorize", pendingArgs("merge", "proposal", "target", "--actor", "m", "--reason", "r")},
		{"merge_missing_actor", pendingArgs("merge", "proposal", "target", "--authorize", "--reason", "r")},
		{"merge_missing_reason", pendingArgs("merge", "proposal", "target", "--authorize", "--actor", "m")},
		{"merge_unknown_pending", pendingArgs("merge", "unknown", "target", "--authorize", "--actor", "m", "--reason", "r")},
		{"merge_active_source", pendingArgs("merge", "target", "proposal", "--authorize", "--actor", "m", "--reason", "r")},
		{"merge_unknown_target", pendingArgs("merge", "proposal", "unknown", "--authorize", "--actor", "m", "--reason", "r")},
		{"merge_pending_target", pendingArgs("merge", "proposal", "remaining", "--authorize", "--actor", "m", "--reason", "r")},
		{"merge_retired_target", pendingArgs("merge", "proposal", "retired", "--authorize", "--actor", "m", "--reason", "r")},
		{"merge_superseded_target", pendingArgs("merge", "proposal", "superseded", "--authorize", "--actor", "m", "--reason", "r")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := pendingFixture(t)
			s := pendingStore(t, root)
			before, err := s.Export(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			s.Close()
			args := append(pendingArgs("areas"), tc.args...)
			args = append(args, "--root", root, "--json")
			reply, text := pendingRun(t, args...)
			if reply.ExitCode != 3 || reply.Status != "unavailable" || reply.Error == "" || reply.Result != nil {
				t.Fatalf("refusal: %s", text)
			}
			s = pendingStore(t, root)
			after, err := s.Export(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, after) {
				t.Fatal("rejected command mutated exported rows")
			}
		})
	}
}
func TestPendingSummary(t *testing.T) {
	a := areas.Area{ScoutAreaID: "proposal", Title: "Proposed concern", Status: areas.StatusPending, FileSet: pendingPaths("pkg/new.go", "pkg/other.go")}
	pending := append([]areas.Area(nil), a)
	report := engine.DiffReport{PendingAreas: pending}
	text := engine.Summary(domain.Gate{Status: "clean"}, engine.Ranking{}, nil, nil, 0, nil, pending)
	if !strings.Contains(text, "proposal") || !strings.Contains(text, "pkg/new.go") || !strings.Contains(text, "pkg/other.go") {
		t.Fatalf("summary omits pending files: %s", text)
	}
	raw, code := captureStdout(t, func() int { return emit(&commonFlags{json: true}, "complete", 0, report, nil) })
	var decoded struct {
		Result struct {
			PendingAreas []areas.Area `json:"pendingAreas"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil || code != 0 || !reflect.DeepEqual(decoded.Result.PendingAreas, pending) {
		t.Fatalf("emitted pendingAreas: %s (exit %d), %v", raw, code, err)
	}
}
func TestAcceptedPendingSurvivesMapRefresh(t *testing.T) {
	root := pendingFixture(t)
	reply, text := pendingRun(t, "areas", "accept", "proposal", "--root", root, "--actor", "m", "--reason", "reviewed", "--authorize", "--json")
	if reply.ExitCode != 0 {
		t.Fatal(text)
	}
	s := pendingStore(t, root)
	ctx := context.Background()
	before, err := s.ListAreas(ctx)
	if err != nil {
		t.Fatal(err)
	}
	next, report := areas.Reconcile(before, nil, func(string) bool { return true }, areas.Config{Snapshot: "refresh"})
	accepted := pendingFind(t, next, "proposal")
	if !accepted.Active() || accepted.AcceptedBy != "m" || accepted.AcceptedAt == nil || accepted.RetiredSnapshot != "" {
		t.Fatalf("accepted area lost on unmatched refresh: %+v %+v", accepted, report)
	}
	if err := s.ReplaceAreas(ctx, next); err != nil {
		t.Fatal(err)
	}
	var pending []areas.Area
	for _, a := range next {
		if a.Status == areas.StatusPending {
			pending = append(pending, a)
		}
	}
	summary := engine.Summary(domain.Gate{Status: "clean"}, engine.Ranking{}, nil, nil, 0, nil, pending)
	if len(pending) != 1 || !strings.Contains(summary, "remaining") || !strings.Contains(summary, "pkg/remaining.go") || strings.Contains(summary, "proposal") {
		t.Fatalf("refresh pending summary: %s", summary)
	}
}

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
	t.Cleanup(func() { closeOrFail(t, s) })
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
func runExitZero(t *testing.T, label string, args ...string) {
	t.Helper()
	if code := run(context.Background(), args); code != 0 {
		t.Fatalf("%s exit %d", label, code)
	}
}

func assertSameRows[T any](t *testing.T, message string, first, second func(context.Context) (T, error)) {
	t.Helper()
	ctx := context.Background()
	a, err := first(ctx)
	if err != nil {
		t.Fatal(err)
	}
	b, err := second(ctx)
	if err != nil || !reflect.DeepEqual(a, b) {
		t.Fatalf("%s: %+v %v", message, b, err)
	}
}

func pendingRoundTrip(t *testing.T, root string) {
	t.Helper()
	dir := t.TempDir()
	runExitZero(t, "export", pendingArgs("export", "--root", root, "--out", dir)...)
	target := t.TempDir()
	runExitZero(t, "import", pendingArgs("import", "--root", target, "--from", dir)...)
	again := t.TempDir()
	runExitZero(t, "re-export", pendingArgs("export", "--root", target, "--out", again)...)
	first := pendingStore(t, root)
	second := pendingStore(t, target)
	assertSameRows(t, "maintainer action was not export/import byte-stable", first.Export, second.Export)
	assertSameRows(t, "restored areas differ", first.ListAreas, second.ListAreas)
	assertSameRows(t, "restored labels differ", first.ListAssociations, second.ListAssociations)
	entries, err := os.ReadDir(filepath.Join(dir, "state", "v1"))
	if err != nil || len(entries) == 0 {
		t.Fatalf("export files: %v %v", entries, err)
	}
}

func decodeAreaResult(t *testing.T, text string) []areas.Area {
	t.Helper()
	var decoded struct {
		Result []areas.Area `json:"result"`
	}
	if err := json.Unmarshal([]byte(text), &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded.Result
}

func assertPlainPendingList(t *testing.T, root string) {
	t.Helper()
	plain, code := captureStdout(t, func() int {
		return run(context.Background(), pendingArgs("areas", "list", "--pending", "--root", root))
	})
	if code != 0 || !strings.Contains(plain, "proposal") || !strings.Contains(plain, "pkg/new.go") || !strings.Contains(plain, "pkg/other.go") || strings.Contains(plain, "pkg/target.go") {
		t.Fatalf("text pending list: exit %d %s", code, plain)
	}
}

func TestPendingAreasList(t *testing.T) {
	root := pendingFixture(t)
	reply, text := pendingRun(t, "areas", "list", "--pending", "--root", root, "--json")
	if reply.ExitCode != 0 || reply.Status != "complete" {
		t.Fatalf("list: %s", text)
	}
	result := decodeAreaResult(t, text)
	if len(result) != 2 || result[0].ScoutAreaID != "proposal" || !reflect.DeepEqual(result[0].FileSet, pendingPaths("pkg/new.go", "pkg/other.go")) || result[1].ScoutAreaID != "remaining" {
		t.Fatalf("pending list: %s", text)
	}
	assertPlainPendingList(t, root)
}

func assertAcceptedProposal(t *testing.T, got areas.Area, before time.Time) {
	t.Helper()
	if !got.Active() || got.Status != areas.StatusActive || got.Origin != "association" || got.AcceptedBy != "maintainer" || got.AcceptedAt == nil || got.AcceptedAt.Before(before) || got.AcceptedAt.After(time.Now().UTC()) {
		t.Fatalf("acceptance: %+v", got)
	}
}

func exportedAreaActions(store *state.Store) ([]state.AreaAction, error) {
	files, err := store.Export(context.Background())
	if err != nil {
		return nil, err
	}
	var actions []state.AreaAction
	for _, line := range state.SplitLines(files[state.ExportAreaActions]) {
		var action state.AreaAction
		if err := json.Unmarshal(line, &action); err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}
	return actions, nil
}

func assertOneAreaAction(t *testing.T, store *state.Store, want state.AreaAction) {
	t.Helper()
	actions, err := exportedAreaActions(store)
	if err != nil || len(actions) != 1 {
		t.Fatalf("%s audit: %+v %v", want.Action, actions, err)
	}
	got := actions[0]
	if got.Action != want.Action || got.AreaID != want.AreaID || got.TargetAreaID != want.TargetAreaID || got.Actor != want.Actor || got.Reason != want.Reason || got.At.IsZero() {
		t.Fatalf("%s audit: %+v %v", want.Action, actions, err)
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
	assertAcceptedProposal(t, pendingFind(t, list, "proposal"), before)
	assertOneAreaAction(t, store, state.AreaAction{Action: "accept", AreaID: "proposal", Actor: "maintainer", Reason: "reviewed concern"})
	closeOrFail(t, store)
	_, pending := pendingRun(t, "areas", "list", "--pending", "--root", root, "--json")
	if strings.Contains(pending, "proposal") || !strings.Contains(pending, "remaining") {
		t.Fatalf("accepted area still pending: %s", pending)
	}
	pendingRoundTrip(t, root)
}

func assertMergedAreas(t *testing.T, list []areas.Area) {
	t.Helper()
	if got := pendingFind(t, list, "proposal"); got.Status != areas.StatusSuperseded || got.Active() || got.AcceptedAt != nil {
		t.Fatalf("merged pending: %+v", got)
	}
	if got := pendingFind(t, list, "target"); !reflect.DeepEqual(got.FileSet, pendingPaths("pkg/target.go")) {
		t.Fatalf("merge polluted map FileSet: %+v", got)
	}
}

func assertMergeLabel(t *testing.T, store *state.Store) {
	t.Helper()
	records, err := store.ListAssociations(context.Background())
	if err != nil || len(records) != 1 {
		t.Fatalf("merge labels: %+v %v", records, err)
	}
	got := records[0]
	if got.Outcome != domain.AssociationAssigned || !reflect.DeepEqual(got.AreaIDs, pendingPaths("target")) || !reflect.DeepEqual(got.Files, pendingPaths("pkg/new.go", "pkg/other.go")) || got.Reason != "maintainer_merge" || got.Superseded {
		t.Fatalf("merge labels: %+v %v", records, err)
	}
}

func TestPendingAreaMerge(t *testing.T) {
	root := pendingFixture(t)
	reply, text := pendingRun(t, "areas", "merge", "proposal", "target", "--root", root, "--actor", "maintainer", "--reason", "same concern", "--authorize", "--json")
	if reply.ExitCode != 0 {
		t.Fatalf("merge: %s", text)
	}
	store := pendingStore(t, root)
	list, err := store.ListAreas(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assertMergedAreas(t, list)
	assertMergeLabel(t, store)
	assertOneAreaAction(t, store, state.AreaAction{Action: "merge", AreaID: "proposal", TargetAreaID: "target", Actor: "maintainer", Reason: "same concern"})
	closeOrFail(t, store)
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
			closeOrFail(t, s)
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
	text := engine.Summary(engine.SummaryInput{Gate: domain.Gate{Status: "clean"}, Pending: pending})
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
func assertAcceptedSurvivesRefresh(t *testing.T, accepted areas.Area, report areas.Report) {
	t.Helper()
	if !accepted.Active() || accepted.AcceptedBy != "m" || accepted.AcceptedAt == nil || accepted.RetiredSnapshot != "" {
		t.Fatalf("accepted area lost on unmatched refresh: %+v %+v", accepted, report)
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
	assertAcceptedSurvivesRefresh(t, pendingFind(t, next, "proposal"), report)
	if err := s.ReplaceAreas(ctx, next); err != nil {
		t.Fatal(err)
	}
	pending := onlyPendingAreas(next)
	summary := engine.Summary(engine.SummaryInput{Gate: domain.Gate{Status: "clean"}, Pending: pending})
	if len(pending) != 1 || !strings.Contains(summary, "remaining") || !strings.Contains(summary, "pkg/remaining.go") || strings.Contains(summary, "proposal") {
		t.Fatalf("refresh pending summary: %s", summary)
	}
}

package areas

import (
	"reflect"
	"testing"
	"time"

	"scout/internal/domain"
)

func lifecycleArea(status string, accepted bool) Area {
	a := Area{ScoutAreaID: "pending", Title: "Proposed", FileSet: lifecycleFiles("a.go", "b.go"), Origin: "association", FirstSeenSnapshot: "s1", Status: status, ClusterID: "cluster", DecisionID: "decision"}
	if accepted {
		at := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
		a.AcceptedBy = "maintainer"
		a.AcceptedAt = &at
	}
	return a
}
func lifecycleFiles(paths ...string) []string                { return paths }
func lifecycleAreas(a ...Area) []Area                        { return a }
func lifecycleFeatures(f ...domain.Feature) []domain.Feature { return f }

func TestPendingReconcileLifecycle(t *testing.T) {
	cases := []struct {
		name              string
		status            string
		accepted, matched bool
	}{
		{"pending_unmatched", StatusPending, false, false},
		{"pending_overlap", StatusPending, false, true},
		{"accepted_matched", StatusActive, true, true},
		{"accepted_unmatched_heuristic", StatusActive, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := lifecycleArea(tc.status, tc.accepted)
			var features []domain.Feature
			if tc.matched {
				features = lifecycleFeatures(feature("fresh", AgentSource, "a.go", "b.go", "c.go"))
			}
			next, report := Reconcile(lifecycleAreas(before), features, func(string) bool { return true }, Config{Snapshot: "s2"})
			got := find(t, next, before.ScoutAreaID)
			if len(report.Retired) != 0 || got.RetiredSnapshot != "" {
				t.Fatalf("protected area retired: %+v %+v", got, report)
			}
			if tc.accepted && tc.matched {
				if len(report.Matched) != 1 || got.Title != "title fresh" || got.Origin != OriginAgent || !reflect.DeepEqual(got.FileSet, lifecycleFiles("a.go", "b.go", "c.go")) || !reflect.DeepEqual(got.ClawpatchIDs, lifecycleFiles("fresh")) {
					t.Fatalf("accepted match not refreshed: %+v %+v", got, report)
				}
			} else {
				if len(report.Matched) != 0 || !reflect.DeepEqual(got, before) {
					t.Fatalf("unmatched/pending area changed: %+v %+v", got, report)
				}
			}
			if got.AcceptedBy != before.AcceptedBy || !reflect.DeepEqual(got.AcceptedAt, before.AcceptedAt) {
				t.Fatalf("acceptance lost: %+v", got)
			}
			if !reflect.DeepEqual(before, lifecycleArea(tc.status, tc.accepted)) {
				t.Fatal("input mutated")
			}
		})
	}
}

func TestPendingMapSupersession(t *testing.T) {
	cases := []struct {
		name              string
		covered, existing []string
		superseded        bool
	}{
		{"all_files_active", lifecycleFiles("a.go", "b.go"), lifecycleFiles("a.go", "b.go"), true},
		{"partial_active", lifecycleFiles("a.go"), lifecycleFiles("a.go", "b.go"), false},
		{"deleted_path", lifecycleFiles("a.go"), lifecycleFiles("a.go"), true},
		{"renamed_old_path", lifecycleFiles("new.go"), lifecycleFiles("new.go"), true},
		{"renamed_uncovered_live_path", lifecycleFiles("new.go"), lifecycleFiles("b.go", "new.go"), false},
		{"all_deleted", nil, nil, true},
		{"heuristic_only", nil, lifecycleFiles("a.go", "b.go"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pending := lifecycleArea(StatusPending, false)
			active := Area{ScoutAreaID: "active", FileSet: tc.covered, Origin: OriginAgent}
			record := domain.Association{ID: "decision", Files: lifecycleFiles("a.go", "b.go"), Outcome: domain.AssociationAssigned, AreaIDs: lifecycleFiles("active")}
			records := append([]domain.Association(nil), record)
			exists := func(path string) bool {
				for _, p := range tc.existing {
					if p == path {
						return true
					}
				}
				return false
			}
			next, decisions := Supersede(lifecycleAreas(pending, active), records, exists)
			got := find(t, next, "pending")
			want := StatusPending
			if tc.superseded {
				want = StatusSuperseded
			}
			if got.Status != want || len(next) != 2 || got.AcceptedAt != nil || got.AcceptedBy != "" || !reflect.DeepEqual(got.FileSet, pending.FileSet) {
				t.Fatalf("supersession = %+v, want status %s without acceptance/deletion", next, want)
			}
			if len(decisions) != 1 || decisions[0].Superseded != tc.superseded {
				t.Fatalf("association supersession = %+v", decisions)
			}
			if records[0].Superseded || pending.Status != StatusPending {
				t.Fatal("supersession mutated input")
			}
			// Pending and retired FileSets must not count as active coverage.
			active.Status = StatusPending
			excluded, _ := Supersede(lifecycleAreas(pending, active), nil, func(string) bool { return true })
			if find(t, excluded, "pending").Status != StatusPending {
				t.Fatal("pending FileSet counted as active")
			}
			active.Status = StatusActive
			active.RetiredSnapshot = "old"
			excluded, _ = Supersede(lifecycleAreas(pending, active), nil, func(string) bool { return true })
			if find(t, excluded, "pending").Status != StatusPending {
				t.Fatal("retired FileSet counted as active")
			}
		})
	}
}

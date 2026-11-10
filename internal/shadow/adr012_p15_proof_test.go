package shadow

import (
	"bytes"
	"fmt"
	"scout/internal/domain"
	"strings"
	"testing"
	"time"
)

// Each proof owns its fixtures so it can be overlaid and compiled alone.
func p15Day(n int) time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(n) * 24 * time.Hour)
}
func p15Stamp(n int) *time.Time { x := p15Day(n); return &x }
func p15Fact(run string, day int) domain.PRRunFact {
	return domain.PRRunFact{ID: domain.AttemptID{RepositoryID: "repo", RunID: run, JobID: "1", Attempt: "1"}, Scope: domain.ShadowScopeCI,
		PRNumber: run, HeadCommit: "head-" + run, HeadTree: "tree-" + run, BaseCommit: "base", MergeBase: "mb", AssessedRevision: "merge",
		CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", CreatedAt: p15Day(day), StartedAt: p15Day(day), FinishedAt: p15Day(day).Add(time.Minute),
		Status: "assessed", Comparison: "comparable", Inspection: "inspected", Pins: domain.ShadowPins{Scout: "scout1", Tools: "tools1", Catalogue: "cat1", Policy: "policy1", Recipe: "recipe1"}, RecordedAt: p15Stamp(day + 1)}
}
func p15Ledger() domain.ShadowLedger {
	l := domain.ShadowLedger{Frontiers: []domain.ShadowFrontier{{RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", CompleteThrough: p15Day(30), PolicyRevision: "protected1", MinDays: 14, MinObservations: 5}},
		Intervals: []domain.ShadowInterval{{ID: "interval1", RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", After: p15Day(-1), Through: p15Day(30)}}}
	for i := 1; i <= 5; i++ {
		f := p15Fact(fmt.Sprint(i), i)
		f.RecordedAt = p15Stamp(30)
		l.Facts = append(l.Facts, f)
	}
	initial := l.Frontiers[0]
	initial.CompleteThrough = p15Day(0)
	l.Frontiers = append([]domain.ShadowFrontier{initial}, l.Frontiers...)
	l.Publications = []domain.ShadowPublication{{ID: "genesis", RepositoryID: "repo", RecordedAt: p15Day(0), Frontiers: l.Frontiers[:1]}, {ID: "sealed", ParentID: "genesis", RepositoryID: "repo", RecordedAt: p15Day(30), Facts: l.Facts, Frontiers: l.Frontiers[1:], Intervals: l.Intervals}}
	return l
}
func p15Query() Query {
	return Query{RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", Now: p15Day(60)}
}
func p15Eval(t *testing.T, l domain.ShadowLedger, q Query, code int) Report {
	t.Helper()
	r, err := Evaluate(l, q)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if r.ExitCode != code {
		t.Fatalf("exit=%d want %d; reason=%s", r.ExitCode, code, r.Dossier.Reason)
	}
	return r
}
func p15Canonical(t *testing.T, r Report) ([]byte, string) {
	t.Helper()
	b, d, e := CanonicalDossier(r.Dossier)
	if e != nil {
		t.Fatal(e)
	}
	if len(b) == 0 || d == "" {
		t.Fatal("empty canonical dossier or digest")
	}
	return b, d
}
func p15Same(t *testing.T, a, b Report) {
	t.Helper()
	ab, ad := p15Canonical(t, a)
	bb, bd := p15Canonical(t, b)
	if !bytes.Equal(ab, bb) || ad != bd {
		t.Fatalf("historical dossier changed: %s / %s", ad, bd)
	}
}
func p15Event(l *domain.ShadowLedger, f domain.PRRunFact, id, verdict string, day int) {
	e := domain.ProviderPREvent{EventID: id, RepositoryID: f.ID.RepositoryID, Scope: domain.ShadowScopeCI, PRNumber: f.PRNumber, HeadCommit: f.HeadCommit, HeadTree: f.HeadTree, MergeBase: f.MergeBase, CheckID: f.CheckID, RuleVersion: f.RuleVersion, IssueID: id, EventAt: f.StartedAt, RecordedAt: f.RecordedAt}
	l.Events = append(l.Events, e)
	l.Links = append(l.Links, domain.PRRunEventLink{AttemptID: f.ID, Scope: domain.ShadowScopeCI, EventID: id})
	if verdict != "" {
		l.Triage = append(l.Triage, domain.ShadowTriage{ActionID: fmt.Sprint(day + 100), EventID: id, RepositoryID: "repo", Scope: domain.ShadowScopeCI, Verdict: verdict, Actor: "maintainer", PolicyRevision: "protected1", TriagedAt: p15Day(day), RecordedAt: p15Stamp(day)})
	}
}
func p15Advance(l *domain.ShadowLedger, day int) {
	f := l.Frontiers[len(l.Frontiers)-1]
	f.CompleteThrough = p15Day(day)
	l.Frontiers = append(l.Frontiers, f)
	for i := range l.Intervals {
		l.Intervals[i].Through = p15Day(day)
	}
}
func p15Veto(t *testing.T, r Report, word string) {
	t.Helper()
	if r.Dossier.Ready || !strings.Contains(strings.ToLower(r.Dossier.Reason), word) {
		t.Fatalf("ready=%v reason=%q; want %s veto", r.Dossier.Ready, r.Dossier.Reason, word)
	}
}

// S4 checks given ledger visibility. Export/merge/import byte parity belongs to S2;
// pending imported-history adoption and trusted publication belong to S2/S3.
func TestADR012HistoricalLateVisibility(t *testing.T) {

	l := p15Ledger()
	q := p15Query()
	q.AsOf = p15Stamp(30)
	before := p15Eval(t, l, q, 0)
	for _, version := range []string{"1", "2"} {
		t.Run("late-version-"+version, func(t *testing.T) {
			next := p15Ledger()
			f := p15Fact("100", 20)
			f.PRNumber = next.Facts[0].PRNumber
			f.HeadTree = next.Facts[0].HeadTree
			f.RuleVersion = version
			f.Inspection = "not_inspected"
			f.RecordedAt = nil
			next.Facts = append(next.Facts, f)
			p15Same(t, before, p15Eval(t, next, q, 0))
			f.RecordedAt = p15Stamp(32)
			next.Facts[len(next.Facts)-1] = f
			next.Publications = append(next.Publications, domain.ShadowPublication{ID: "late", ParentID: "sealed", RepositoryID: "repo", RecordedAt: p15Day(32), Facts: []domain.PRRunFact{f}})
			p15Same(t, before, p15Eval(t, next, q, 0))
			if version == "1" {
				p15Event(&next, f, "late-event", "", 32)
				p15Advance(&next, 40)
				current := p15Query()
				p15Veto(t, p15Eval(t, next, current, 0), "undecided")
				if !p15Eval(t, p15Ledger(), current, 0).Dossier.Ready {
					t.Fatal("ready control failed")
				}
			}
		})
	}
	for _, tc := range []struct {
		late  bool
		stamp *time.Time
		want  bool
	}{{false, p15Stamp(32), true}, {true, p15Stamp(32), false}, {true, p15Stamp(30), true}, {false, nil, false}} {
		if got := Visible(p15Day(20), tc.late, tc.stamp, p15Day(30)); got != tc.want {
			t.Fatalf("Visible late=%v stamp=%v: %v", tc.late, tc.stamp, got)
		}
	}

}

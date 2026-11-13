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
func p16Day(n int) time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(n) * 24 * time.Hour)
}
func p16Stamp(n int) *time.Time { x := p16Day(n); return &x }
func p16Fact(run string, day int) domain.PRRunFact {
	return domain.PRRunFact{ID: domain.AttemptID{RepositoryID: "repo", RunID: run, JobID: "1", Attempt: "1"}, Scope: domain.ShadowScopeCI,
		PRNumber: run, HeadCommit: "head-" + run, HeadTree: "tree-" + run, BaseCommit: "base", MergeBase: "mb", AssessedRevision: "merge",
		CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", CreatedAt: p16Day(day), StartedAt: p16Day(day), FinishedAt: p16Day(day).Add(time.Minute),
		Status: "assessed", Comparison: "comparable", Inspection: "inspected", Pins: domain.ShadowPins{Scout: "scout1", Tools: "tools1", Catalogue: "cat1", Policy: "policy1", Recipe: "recipe1"}, RecordedAt: p16Stamp(day + 1)}
}
func p16Ledger() domain.ShadowLedger {
	l := domain.ShadowLedger{Frontiers: []domain.ShadowFrontier{{RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", CompleteThrough: p16Day(30), PolicyRevision: "protected1", MinDays: 14, MinObservations: 5}},
		Intervals: []domain.ShadowInterval{{ID: "interval1", RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", After: p16Day(-1), Through: p16Day(30)}}}
	for i := 1; i <= 5; i++ {
		f := p16Fact(fmt.Sprint(i), i)
		f.RecordedAt = p16Stamp(30)
		l.Facts = append(l.Facts, f)
	}
	initial := l.Frontiers[0]
	initial.CompleteThrough = p16Day(0)
	l.Frontiers = append([]domain.ShadowFrontier{initial}, l.Frontiers...)
	l.Publications = []domain.ShadowPublication{{ID: "genesis", RepositoryID: "repo", RecordedAt: p16Day(0), Frontiers: l.Frontiers[:1]}, {ID: "sealed", ParentID: "genesis", RepositoryID: "repo", RecordedAt: p16Day(30), Facts: l.Facts, Frontiers: l.Frontiers[1:], Intervals: l.Intervals}}
	return l
}
func p16Query() Query {
	return Query{RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", Now: p16Day(60)}
}
func p16Eval(t *testing.T, l domain.ShadowLedger, q Query, code int) Report {
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
func p16Canonical(t *testing.T, r Report) ([]byte, string) {
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
func p16Same(t *testing.T, a, b Report) {
	t.Helper()
	ab, ad := p16Canonical(t, a)
	bb, bd := p16Canonical(t, b)
	if !bytes.Equal(ab, bb) || ad != bd {
		t.Fatalf("historical dossier changed: %s / %s", ad, bd)
	}
}
func p16Event(l *domain.ShadowLedger, f domain.PRRunFact, id, verdict string, day int) {
	e := domain.ProviderPREvent{EventID: id, RepositoryID: f.ID.RepositoryID, Scope: domain.ShadowScopeCI, PRNumber: f.PRNumber, HeadCommit: f.HeadCommit, HeadTree: f.HeadTree, MergeBase: f.MergeBase, CheckID: f.CheckID, RuleVersion: f.RuleVersion, IssueID: id, EventAt: f.StartedAt, RecordedAt: f.RecordedAt}
	l.Events = append(l.Events, e)
	l.Links = append(l.Links, domain.PRRunEventLink{AttemptID: f.ID, Scope: domain.ShadowScopeCI, EventID: id})
	if verdict != "" {
		l.Triage = append(l.Triage, domain.ShadowTriage{ActionID: fmt.Sprint(day + 100), EventID: id, RepositoryID: "repo", Scope: domain.ShadowScopeCI, Verdict: verdict, Actor: "maintainer", PolicyRevision: "protected1", TriagedAt: p16Day(day), RecordedAt: p16Stamp(day)})
	}
}
func p16Advance(l *domain.ShadowLedger, day int) {
	f := l.Frontiers[len(l.Frontiers)-1]
	f.CompleteThrough = p16Day(day)
	l.Frontiers = append(l.Frontiers, f)
	for i := range l.Intervals {
		l.Intervals[i].Through = p16Day(day)
	}
}
func p16Veto(t *testing.T, r Report, word string) {
	t.Helper()
	if r.Dossier.Ready || !strings.Contains(strings.ToLower(r.Dossier.Reason), word) {
		t.Fatalf("ready=%v reason=%q; want %s veto", r.Dossier.Ready, r.Dossier.Reason, word)
	}
}

func TestADR012GapPointIntervals(t *testing.T) {

	for _, at := range []int{20, 30} {
		t.Run(fmt.Sprint(at), func(t *testing.T) {
			l := p16Ledger()
			point := domain.ShadowGapPoint{ID: "gap", RepositoryID: "repo", CheckID: p16Query().CheckID, At: p16Day(at), RecordedAt: p16Stamp(32)}
			// The publication's parent has C=10: even a stamp after new C must not hide a normal point.
			l.Publications[1].Frontiers[0].CompleteThrough = p16Day(10)
			l.GapPoints = []domain.ShadowGapPoint{point}
			l.Publications = append(l.Publications, domain.ShadowPublication{ID: "normal-point", ParentID: "sealed", RepositoryID: "repo", RecordedAt: p16Day(32), GapPoints: l.GapPoints, Frontiers: l.Frontiers})
			l.Facts = append(l.Facts, p16Fact("20", at), p16Fact("21", at+1))
			p16Advance(&l, 35)
			q := p16Query()
			q.AsOf = p16Stamp(at)
			p16Eval(t, l, q, 2)
			r := p16Eval(t, l, p16Query(), 0)
			if r.Dossier.Origin == nil || !r.Dossier.Origin.After(p16Day(at)) {
				t.Fatalf("origin=%v must be strictly after point", r.Dossier.Origin)
			}
		})
	}
	l := p16Ledger()
	q := p16Query()
	q.AsOf = p16Stamp(30)
	before := p16Eval(t, l, q, 0)
	// Independent points in one late publication: equality at parent C is late too.
	points := []domain.ShadowGapPoint{{ID: "event", RepositoryID: "repo", CheckID: q.CheckID, At: p16Day(20), RecordedAt: p16Stamp(32)}, {ID: "start", RepositoryID: "repo", CheckID: q.CheckID, At: p16Day(30), RecordedAt: p16Stamp(32)}, {ID: "normal", RepositoryID: "repo", CheckID: q.CheckID, At: p16Day(31), RecordedAt: p16Stamp(32)}}
	l.GapPoints = points
	l.Publications = append(l.Publications, domain.ShadowPublication{ID: "points", ParentID: "sealed", RepositoryID: "repo", RecordedAt: p16Day(32), GapPoints: points})
	p16Same(t, before, p16Eval(t, l, q, 0))
	p16Advance(&l, 40)
	l.Facts = append(l.Facts, p16Fact("40", 33))
	r := p16Eval(t, l, p16Query(), 0)
	if r.Dossier.Origin == nil || !r.Dossier.Origin.After(p16Day(31)) {
		t.Fatal("late points did not split recovered interval")
	}

}

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
func p12Day(n int) time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(n) * 24 * time.Hour)
}
func p12Stamp(n int) *time.Time { x := p12Day(n); return &x }
func p12Fact(run string, day int) domain.PRRunFact {
	return domain.PRRunFact{ID: domain.AttemptID{RepositoryID: "repo", RunID: run, JobID: "1", Attempt: "1"}, Scope: domain.ShadowScopeCI,
		PRNumber: run, HeadCommit: "head-" + run, HeadTree: "tree-" + run, BaseCommit: "base", MergeBase: "mb", AssessedRevision: "merge",
		CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", CreatedAt: p12Day(day), StartedAt: p12Day(day), FinishedAt: p12Day(day).Add(time.Minute),
		Status: "assessed", Comparison: "comparable", Inspection: "inspected", Pins: domain.ShadowPins{Scout: "scout1", Tools: "tools1", Catalogue: "cat1", Policy: "policy1", Recipe: "recipe1"}, RecordedAt: p12Stamp(day + 1)}
}
func p12Ledger() domain.ShadowLedger {
	l := domain.ShadowLedger{Frontiers: []domain.ShadowFrontier{{RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", CompleteThrough: p12Day(30), PolicyRevision: "protected1", MinDays: 14, MinObservations: 5}},
		Intervals: []domain.ShadowInterval{{ID: "interval1", RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", After: p12Day(-1), Through: p12Day(30)}}}
	for i := 1; i <= 5; i++ {
		f := p12Fact(fmt.Sprint(i), i)
		f.RecordedAt = p12Stamp(30)
		l.Facts = append(l.Facts, f)
	}
	initial := l.Frontiers[0]
	initial.CompleteThrough = p12Day(0)
	l.Frontiers = append([]domain.ShadowFrontier{initial}, l.Frontiers...)
	l.Publications = []domain.ShadowPublication{{ID: "genesis", RepositoryID: "repo", RecordedAt: p12Day(0), Frontiers: l.Frontiers[:1]}, {ID: "sealed", ParentID: "genesis", RepositoryID: "repo", RecordedAt: p12Day(30), Facts: l.Facts, Frontiers: l.Frontiers[1:], Intervals: l.Intervals}}
	return l
}
func p12Query() Query {
	return Query{RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", Now: p12Day(60)}
}
func p12Eval(t *testing.T, l domain.ShadowLedger, q Query, code int) Report {
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
func p12Canonical(t *testing.T, r Report) ([]byte, string) {
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
func p12Same(t *testing.T, a, b Report) {
	t.Helper()
	ab, ad := p12Canonical(t, a)
	bb, bd := p12Canonical(t, b)
	if !bytes.Equal(ab, bb) || ad != bd {
		t.Fatalf("historical dossier changed: %s / %s", ad, bd)
	}
}
func p12Event(l *domain.ShadowLedger, f domain.PRRunFact, id, verdict string, day int) {
	e := domain.ProviderPREvent{EventID: id, RepositoryID: f.ID.RepositoryID, Scope: domain.ShadowScopeCI, PRNumber: f.PRNumber, HeadCommit: f.HeadCommit, HeadTree: f.HeadTree, MergeBase: f.MergeBase, CheckID: f.CheckID, RuleVersion: f.RuleVersion, IssueID: id, EventAt: f.StartedAt, RecordedAt: f.RecordedAt}
	l.Events = append(l.Events, e)
	l.Links = append(l.Links, domain.PRRunEventLink{AttemptID: f.ID, Scope: domain.ShadowScopeCI, EventID: id})
	if verdict != "" {
		l.Triage = append(l.Triage, domain.ShadowTriage{ActionID: fmt.Sprint(day + 100), EventID: id, RepositoryID: "repo", Scope: domain.ShadowScopeCI, Verdict: verdict, Actor: "maintainer", PolicyRevision: "protected1", TriagedAt: p12Day(day), RecordedAt: p12Stamp(day)})
	}
}
func p12Advance(l *domain.ShadowLedger, day int) {
	f := l.Frontiers[len(l.Frontiers)-1]
	f.CompleteThrough = p12Day(day)
	l.Frontiers = append(l.Frontiers, f)
	for i := range l.Intervals {
		l.Intervals[i].Through = p12Day(day)
	}
}
func p12Veto(t *testing.T, r Report, word string) {
	t.Helper()
	if r.Dossier.Ready || !strings.Contains(strings.ToLower(r.Dossier.Reason), word) {
		t.Fatalf("ready=%v reason=%q; want %s veto", r.Dossier.Ready, r.Dossier.Reason, word)
	}
}

// S4 consumes the given deadline gaps; deadline selection, obligation validation,
// discharge and trusted late stamping themselves belong to S3.
func TestADR012DeadlineLateLifecycle(t *testing.T) {

	l := p12Ledger()
	q := p12Query()
	q.AsOf = p12Stamp(30)
	before := p12Eval(t, l, q, 0)
	// Given a deadline-discharge publication, query must classify each obligation independently.
	points := []domain.ShadowGapPoint{{ID: "started", RepositoryID: "repo", CheckID: q.CheckID, At: p12Day(20), RecordedAt: p12Stamp(31)}, {ID: "event-a", RepositoryID: "repo", CheckID: q.CheckID, At: p12Day(18), RecordedAt: p12Stamp(31)}, {ID: "event-b", RepositoryID: "repo", CheckID: q.CheckID, At: p12Day(19), RecordedAt: p12Stamp(31)}}
	l.GapPoints = points
	l.Publications = append(l.Publications, domain.ShadowPublication{ID: "deadline", ParentID: "sealed", RepositoryID: "repo", RecordedAt: p12Day(31), GapPoints: points})
	p12Same(t, before, p12Eval(t, l, q, 0))
	p12Advance(&l, 32)
	terminal := p12Fact("100", 20)
	terminal.RecordedAt = p12Stamp(34)
	l.Facts = append(l.Facts, terminal)
	p12Event(&l, terminal, "late-terminal-event", "", 34)
	l.Publications = append(l.Publications, domain.ShadowPublication{ID: "terminal", ParentID: "deadline", RepositoryID: "repo", RecordedAt: p12Day(34), Facts: []domain.PRRunFact{terminal}, Events: l.Events, Links: l.Links})
	// Reproduction before stamps is unchanged; current certification refuses until recovery.
	p12Same(t, before, p12Eval(t, l, q, 0))
	p12Eval(t, l, p12Query(), 2)
	for i := 0; i < 5; i++ {
		l.Facts = append(l.Facts, p12Fact(fmt.Sprint(200+i), 35+i))
	}
	p12Advance(&l, 55)
	r := p12Eval(t, l, p12Query(), 0)
	p12Veto(t, r, "undecided")
	control := l
	control.Events = nil
	control.Links = nil
	control.Publications = append([]domain.ShadowPublication(nil), l.Publications...)
	control.Publications[len(control.Publications)-1].Events = nil
	control.Publications[len(control.Publications)-1].Links = nil
	if !p12Eval(t, control, p12Query(), 0).Dossier.Ready {
		t.Fatal("otherwise-identical recovered control must be ready")
	}
	for _, point := range points {
		if !Visible(point.At, true, point.RecordedAt, p12Day(31)) || Visible(point.At, true, point.RecordedAt, p12Day(30)) {
			t.Fatalf("independent obligation visibility %s", point.ID)
		}
	}

}

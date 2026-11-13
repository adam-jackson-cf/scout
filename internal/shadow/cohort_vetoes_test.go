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
func p17Day(n int) time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(n) * 24 * time.Hour)
}
func p17Stamp(n int) *time.Time { x := p17Day(n); return &x }
func p17Fact(run string, day int) domain.PRRunFact {
	return domain.PRRunFact{ID: domain.AttemptID{RepositoryID: "repo", RunID: run, JobID: "1", Attempt: "1"}, Scope: domain.ShadowScopeCI,
		PRNumber: run, HeadCommit: "head-" + run, HeadTree: "tree-" + run, BaseCommit: "base", MergeBase: "mb", AssessedRevision: "merge",
		CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", CreatedAt: p17Day(day), StartedAt: p17Day(day), FinishedAt: p17Day(day).Add(time.Minute),
		Status: "assessed", Comparison: "comparable", Inspection: "inspected", Pins: domain.ShadowPins{Scout: "scout1", Tools: "tools1", Catalogue: "cat1", Policy: "policy1", Recipe: "recipe1"}, RecordedAt: p17Stamp(day + 1)}
}
func p17Ledger() domain.ShadowLedger {
	l := domain.ShadowLedger{Frontiers: []domain.ShadowFrontier{{RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", CompleteThrough: p17Day(30), PolicyRevision: "protected1", MinDays: 14, MinObservations: 5}},
		Intervals: []domain.ShadowInterval{{ID: "interval1", RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", After: p17Day(-1), Through: p17Day(30)}}}
	for i := 1; i <= 5; i++ {
		f := p17Fact(fmt.Sprint(i), i)
		f.RecordedAt = p17Stamp(30)
		l.Facts = append(l.Facts, f)
	}
	initial := l.Frontiers[0]
	initial.CompleteThrough = p17Day(0)
	l.Frontiers = append([]domain.ShadowFrontier{initial}, l.Frontiers...)
	l.Publications = []domain.ShadowPublication{{ID: "genesis", RepositoryID: "repo", RecordedAt: p17Day(0), Frontiers: l.Frontiers[:1]}, {ID: "sealed", ParentID: "genesis", RepositoryID: "repo", RecordedAt: p17Day(30), Facts: l.Facts, Frontiers: l.Frontiers[1:], Intervals: l.Intervals}}
	return l
}
func p17Query() Query {
	return Query{RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", Now: p17Day(60)}
}
func p17Eval(t *testing.T, l domain.ShadowLedger, q Query, code int) Report {
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
func p17Canonical(t *testing.T, r Report) ([]byte, string) {
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
func p17Same(t *testing.T, a, b Report) {
	t.Helper()
	ab, ad := p17Canonical(t, a)
	bb, bd := p17Canonical(t, b)
	if !bytes.Equal(ab, bb) || ad != bd {
		t.Fatalf("historical dossier changed: %s / %s", ad, bd)
	}
}
func p17Event(l *domain.ShadowLedger, f domain.PRRunFact, id, verdict string, day int) {
	e := domain.ProviderPREvent{EventID: id, RepositoryID: f.ID.RepositoryID, Scope: domain.ShadowScopeCI, PRNumber: f.PRNumber, HeadCommit: f.HeadCommit, HeadTree: f.HeadTree, MergeBase: f.MergeBase, CheckID: f.CheckID, RuleVersion: f.RuleVersion, IssueID: id, EventAt: f.StartedAt, RecordedAt: f.RecordedAt}
	l.Events = append(l.Events, e)
	l.Links = append(l.Links, domain.PRRunEventLink{AttemptID: f.ID, Scope: domain.ShadowScopeCI, EventID: id})
	if verdict != "" {
		l.Triage = append(l.Triage, domain.ShadowTriage{ActionID: fmt.Sprint(day + 100), EventID: id, RepositoryID: "repo", Scope: domain.ShadowScopeCI, Verdict: verdict, Actor: "maintainer", PolicyRevision: "protected1", TriagedAt: p17Day(day), RecordedAt: p17Stamp(day)})
	}
}
func p17Advance(l *domain.ShadowLedger, day int) {
	f := l.Frontiers[len(l.Frontiers)-1]
	f.CompleteThrough = p17Day(day)
	l.Frontiers = append(l.Frontiers, f)
	for i := range l.Intervals {
		l.Intervals[i].Through = p17Day(day)
	}
}
func p17Veto(t *testing.T, r Report, word string) {
	t.Helper()
	if r.Dossier.Ready || !strings.Contains(strings.ToLower(r.Dossier.Reason), word) {
		t.Fatalf("ready=%v reason=%q; want %s veto", r.Dossier.Ready, r.Dossier.Reason, word)
	}
}

func TestADR012CohortWideVetoes(t *testing.T) {

	for _, verdict := range []string{"not_warranted", ""} {
		for _, position := range []string{"prior-interval", "at-gap", "nonrepresentative"} {
			t.Run(verdict+"/"+position, func(t *testing.T) {
				l := p17Ledger()
				old := p17Fact("100", 5)
				if position == "at-gap" {
					old.StartedAt = p17Day(10)
				}
				if position == "nonrepresentative" {
					old.PRNumber = l.Facts[0].PRNumber
					old.HeadTree = l.Facts[0].HeadTree
					newer := old
					newer.ID.RunID = "101"
					newer.StartedAt = p17Day(15)
					l.Facts = append(l.Facts, newer)
				}
				l.Facts = append(l.Facts, old)
				p17Event(&l, old, "old-event", verdict, 16)
				l.GapPoints = []domain.ShadowGapPoint{{ID: "gap", RepositoryID: "repo", CheckID: old.CheckID, At: p17Day(10), RecordedAt: p17Stamp(11)}}
				for i := 0; i < 5; i++ {
					l.Facts[i].StartedAt = p17Day(11 + i)
				}
				word := verdict
				if word == "" {
					word = "undecided"
				}
				r := p17Eval(t, l, p17Query(), 0)
				p17Veto(t, r, word)
			})
		}
	}
	for _, outside := range []string{"repository", "check", "version", "local", "legacy"} {
		t.Run("exclude-"+outside, func(t *testing.T) {
			l := p17Ledger()
			f := p17Fact("100", 2)
			switch outside {
			case "repository":
				f.ID.RepositoryID = "other"
			case "check":
				f.CheckID = "other"
			case "version":
				f.RuleVersion = "other"
			case "local":
				f.Scope = domain.ShadowScopeLocal
			case "legacy":
				f.Scope = domain.ShadowScopeLegacy
			}
			l.Facts = append(l.Facts, f)
			p17Event(&l, f, "foreign", "", 3)
			q := p17Query()
			q.RuleVersion = "1"
			if !p17Eval(t, l, q, 0).Dossier.Ready {
				t.Fatal("foreign evidence vetoed cohort")
			}
		})
	}

}

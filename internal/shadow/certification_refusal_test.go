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
func p20Day(n int) time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(n) * 24 * time.Hour)
}
func p20Stamp(n int) *time.Time { x := p20Day(n); return &x }
func p20Fact(run string, day int) domain.PRRunFact {
	return domain.PRRunFact{ID: domain.AttemptID{RepositoryID: "repo", RunID: run, JobID: "1", Attempt: "1"}, Scope: domain.ShadowScopeCI,
		PRNumber: run, HeadCommit: "head-" + run, HeadTree: "tree-" + run, BaseCommit: "base", MergeBase: "mb", AssessedRevision: "merge",
		CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", CreatedAt: p20Day(day), StartedAt: p20Day(day), FinishedAt: p20Day(day).Add(time.Minute),
		Status: "assessed", Comparison: "comparable", Inspection: "inspected", Pins: domain.ShadowPins{Scout: "scout1", Tools: "tools1", Catalogue: "cat1", Policy: "policy1", Recipe: "recipe1"}, RecordedAt: p20Stamp(day + 1)}
}
func p20Ledger() domain.ShadowLedger {
	l := domain.ShadowLedger{Frontiers: []domain.ShadowFrontier{{RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", CompleteThrough: p20Day(30), PolicyRevision: "protected1", MinDays: 14, MinObservations: 5}},
		Intervals: []domain.ShadowInterval{{ID: "interval1", RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", After: p20Day(-1), Through: p20Day(30)}}}
	for i := 1; i <= 5; i++ {
		f := p20Fact(fmt.Sprint(i), i)
		f.RecordedAt = p20Stamp(30)
		l.Facts = append(l.Facts, f)
	}
	initial := l.Frontiers[0]
	initial.CompleteThrough = p20Day(0)
	l.Frontiers = append([]domain.ShadowFrontier{initial}, l.Frontiers...)
	l.Publications = []domain.ShadowPublication{{ID: "genesis", RepositoryID: "repo", RecordedAt: p20Day(0), Frontiers: l.Frontiers[:1]}, {ID: "sealed", ParentID: "genesis", RepositoryID: "repo", RecordedAt: p20Day(30), Facts: l.Facts, Frontiers: l.Frontiers[1:], Intervals: l.Intervals}}
	return l
}
func p20Query() Query {
	return Query{RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", Now: p20Day(60)}
}
func p20Eval(t *testing.T, l domain.ShadowLedger, q Query, code int) Report {
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
func p20Canonical(t *testing.T, r Report) ([]byte, string) {
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
func p20Same(t *testing.T, a, b Report) {
	t.Helper()
	ab, ad := p20Canonical(t, a)
	bb, bd := p20Canonical(t, b)
	if !bytes.Equal(ab, bb) || ad != bd {
		t.Fatalf("historical dossier changed: %s / %s", ad, bd)
	}
}
func p20Event(l *domain.ShadowLedger, f domain.PRRunFact, id, verdict string, day int) {
	e := domain.ProviderPREvent{EventID: id, RepositoryID: f.ID.RepositoryID, Scope: domain.ShadowScopeCI, PRNumber: f.PRNumber, HeadCommit: f.HeadCommit, HeadTree: f.HeadTree, MergeBase: f.MergeBase, CheckID: f.CheckID, RuleVersion: f.RuleVersion, IssueID: id, EventAt: f.StartedAt, RecordedAt: f.RecordedAt}
	l.Events = append(l.Events, e)
	l.Links = append(l.Links, domain.PRRunEventLink{AttemptID: f.ID, Scope: domain.ShadowScopeCI, EventID: id})
	if verdict != "" {
		l.Triage = append(l.Triage, domain.ShadowTriage{ActionID: fmt.Sprint(day + 100), EventID: id, RepositoryID: "repo", Scope: domain.ShadowScopeCI, Verdict: verdict, Actor: "maintainer", PolicyRevision: "protected1", TriagedAt: p20Day(day), RecordedAt: p20Stamp(day)})
	}
}
func p20Advance(l *domain.ShadowLedger, day int) {
	f := l.Frontiers[len(l.Frontiers)-1]
	f.CompleteThrough = p20Day(day)
	l.Frontiers = append(l.Frontiers, f)
	for i := range l.Intervals {
		l.Intervals[i].Through = p20Day(day)
	}
}
func p20Veto(t *testing.T, r Report, word string) {
	t.Helper()
	if r.Dossier.Ready || !strings.Contains(strings.ToLower(r.Dossier.Reason), word) {
		t.Fatalf("ready=%v reason=%q; want %s veto", r.Dossier.Ready, r.Dossier.Reason, word)
	}
}

// Excludes S3/S4 integration: independently pending imported history and
// state-changing import-refusal records, absent from the S1 ledger contract.
func TestADR012CertificationRefusalScope(t *testing.T) {

	t.Run("last-clearing-triage-and-quiet-reconcile", func(t *testing.T) {
		l := p20Ledger()
		p20Event(&l, l.Facts[0], "e", "", 2)
		q := p20Query()
		q.AsOf = p20Stamp(30)
		historical := p20Eval(t, l, q, 0)
		p20Veto(t, historical, "undecided")
		l.Triage = append(l.Triage, domain.ShadowTriage{ActionID: "1000", RepositoryID: "repo", Scope: domain.ShadowScopeCI, EventID: "e", Verdict: "warranted", Actor: "maintainer", PolicyRevision: "protected1", TriagedAt: p20Day(31), RecordedAt: p20Stamp(32)})
		p20Eval(t, l, p20Query(), 2)
		p20Same(t, historical, p20Eval(t, l, q, 0))
		p20Advance(&l, 35)
		if !p20Eval(t, l, p20Query(), 0).Dossier.Ready {
			t.Fatal("clearing triage did not recover after quiet reconcile")
		}
	})
	for _, kind := range []string{"normal-post-C", "other-check-triage", "frontier-only", "publication-only", "other-check-late", "pending-local-late"} {
		t.Run(kind, func(t *testing.T) {
			l := p20Ledger()
			want := 0
			switch kind {
			case "normal-post-C":
				f := p20Fact("100", 31)
				l.Facts = append(l.Facts, f)
			case "other-check-triage":
				f := p20Fact("100", 1)
				f.CheckID = "other"
				l.Facts = append(l.Facts, f)
				p20Event(&l, f, "other-event", "warranted", 32)
			case "frontier-only":
				p20Advance(&l, 35)
			case "publication-only":
				l.Publications = append(l.Publications, domain.ShadowPublication{ID: "empty", ParentID: "sealed", RepositoryID: "repo", RecordedAt: p20Day(32)})
			case "other-check-late", "pending-local-late":
				f := p20Fact("100", 20)
				f.CheckID = "other"
				f.RecordedAt = p20Stamp(32)
				if kind == "pending-local-late" {
					f.RecordedAt = nil
				}
				l.Facts = append(l.Facts, f)
				l.Publications = append(l.Publications, domain.ShadowPublication{ID: "late", ParentID: "sealed", RepositoryID: "repo", RecordedAt: p20Day(32), Facts: []domain.PRRunFact{f}})
				want = 2
			}
			q := p20Query()
			q.AsOf = p20Stamp(30)
			old := p20Eval(t, p20Ledger(), q, 0)
			p20Same(t, old, p20Eval(t, l, q, 0))
			p20Eval(t, l, p20Query(), want)
			if kind == "other-check-late" {
				p20Advance(&l, 35)
				p20Eval(t, l, p20Query(), 0)
			}
			if kind == "pending-local-late" {
				p20Advance(&l, 35)
				p20Eval(t, l, p20Query(), 2)
				l.Facts[len(l.Facts)-1].RecordedAt = p20Stamp(36)
				l.Publications[len(l.Publications)-1].Facts[0].RecordedAt = p20Stamp(36)
				l.Publications[len(l.Publications)-1].RecordedAt = p20Day(36)
				p20Advance(&l, 40)
				p20Eval(t, l, p20Query(), 0)
			}
		})
	}

}

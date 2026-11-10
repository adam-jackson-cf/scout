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
func p21Day(n int) time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(n) * 24 * time.Hour)
}
func p21Stamp(n int) *time.Time { x := p21Day(n); return &x }
func p21Fact(run string, day int) domain.PRRunFact {
	return domain.PRRunFact{ID: domain.AttemptID{RepositoryID: "repo", RunID: run, JobID: "1", Attempt: "1"}, Scope: domain.ShadowScopeCI,
		PRNumber: run, HeadCommit: "head-" + run, HeadTree: "tree-" + run, BaseCommit: "base", MergeBase: "mb", AssessedRevision: "merge",
		CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", CreatedAt: p21Day(day), StartedAt: p21Day(day), FinishedAt: p21Day(day).Add(time.Minute),
		Status: "assessed", Comparison: "comparable", Inspection: "inspected", Pins: domain.ShadowPins{Scout: "scout1", Tools: "tools1", Catalogue: "cat1", Policy: "policy1", Recipe: "recipe1"}, RecordedAt: p21Stamp(day + 1)}
}
func p21Ledger() domain.ShadowLedger {
	l := domain.ShadowLedger{Frontiers: []domain.ShadowFrontier{{RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", CompleteThrough: p21Day(30), PolicyRevision: "protected1", MinDays: 14, MinObservations: 5}},
		Intervals: []domain.ShadowInterval{{ID: "interval1", RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", After: p21Day(-1), Through: p21Day(30)}}}
	for i := 1; i <= 5; i++ {
		f := p21Fact(fmt.Sprint(i), i)
		f.RecordedAt = p21Stamp(30)
		l.Facts = append(l.Facts, f)
	}
	initial := l.Frontiers[0]
	initial.CompleteThrough = p21Day(0)
	l.Frontiers = append([]domain.ShadowFrontier{initial}, l.Frontiers...)
	l.Publications = []domain.ShadowPublication{{ID: "genesis", RepositoryID: "repo", RecordedAt: p21Day(0), Frontiers: l.Frontiers[:1]}, {ID: "sealed", ParentID: "genesis", RepositoryID: "repo", RecordedAt: p21Day(30), Facts: l.Facts, Frontiers: l.Frontiers[1:], Intervals: l.Intervals}}
	return l
}
func p21Query() Query {
	return Query{RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", Now: p21Day(60)}
}
func p21Eval(t *testing.T, l domain.ShadowLedger, q Query, code int) Report {
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
func p21Canonical(t *testing.T, r Report) ([]byte, string) {
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
func p21Same(t *testing.T, a, b Report) {
	t.Helper()
	ab, ad := p21Canonical(t, a)
	bb, bd := p21Canonical(t, b)
	if !bytes.Equal(ab, bb) || ad != bd {
		t.Fatalf("historical dossier changed: %s / %s", ad, bd)
	}
}
func p21Event(l *domain.ShadowLedger, f domain.PRRunFact, id, verdict string, day int) {
	e := domain.ProviderPREvent{EventID: id, RepositoryID: f.ID.RepositoryID, Scope: domain.ShadowScopeCI, PRNumber: f.PRNumber, HeadCommit: f.HeadCommit, HeadTree: f.HeadTree, MergeBase: f.MergeBase, CheckID: f.CheckID, RuleVersion: f.RuleVersion, IssueID: id, EventAt: f.StartedAt, RecordedAt: f.RecordedAt}
	l.Events = append(l.Events, e)
	l.Links = append(l.Links, domain.PRRunEventLink{AttemptID: f.ID, Scope: domain.ShadowScopeCI, EventID: id})
	if verdict != "" {
		l.Triage = append(l.Triage, domain.ShadowTriage{ActionID: fmt.Sprint(day + 100), EventID: id, RepositoryID: "repo", Scope: domain.ShadowScopeCI, Verdict: verdict, Actor: "maintainer", PolicyRevision: "protected1", TriagedAt: p21Day(day), RecordedAt: p21Stamp(day)})
	}
}
func p21Advance(l *domain.ShadowLedger, day int) {
	f := l.Frontiers[len(l.Frontiers)-1]
	f.CompleteThrough = p21Day(day)
	l.Frontiers = append(l.Frontiers, f)
	for i := range l.Intervals {
		l.Intervals[i].Through = p21Day(day)
	}
}
func p21Veto(t *testing.T, r Report, word string) {
	t.Helper()
	if r.Dossier.Ready || !strings.Contains(strings.ToLower(r.Dossier.Reason), word) {
		t.Fatalf("ready=%v reason=%q; want %s veto", r.Dossier.Ready, r.Dossier.Reason, word)
	}
}

func TestADR012DossierHistoricalPolicyAndDigest(t *testing.T) {
	t.Run("canonical-decision-fields", func(t *testing.T) {
		original := p21Eval(t, p21Ledger(), p21Query(), 0)
		_, digest := p21Canonical(t, original)
		for _, field := range []string{"reason", "ready", "counts", "version"} {
			changed := original
			switch field {
			case "reason":
				changed.Dossier.Reason = "different immutable reason"
			case "ready":
				changed.Dossier.Ready = !changed.Dossier.Ready
			case "counts":
				changed.Dossier.Stats.PRRuns++
			case "version":
				changed.Dossier.RuleVersion = "2"
			}
			_, other := p21Canonical(t, changed)
			if other == digest {
				t.Fatalf("digest omitted %s", field)
			}
		}
	})

	l := p21Ledger()
	q := p21Query()
	q.AsOf = p21Stamp(30)
	before := p21Eval(t, l, q, 0)
	if !before.Dossier.Ready {
		t.Fatal("historical policy control not ready")
	}
	// A later protected-policy snapshot cannot rewrite the minima sealed at T.
	later := l.Frontiers[0]
	later.CompleteThrough = p21Day(40)
	later.PolicyRevision = "protected2"
	later.MinDays = 100
	later.MinObservations = 100
	l.Frontiers = append(l.Frontiers, later)
	l.Intervals[0].Through = p21Day(40)
	l.Publications = append(l.Publications, domain.ShadowPublication{ID: "policy-change", ParentID: "sealed", RepositoryID: "repo", RecordedAt: p21Day(41), Frontiers: []domain.ShadowFrontier{later}})
	p21Same(t, before, p21Eval(t, l, q, 0))
	current := p21Eval(t, l, p21Query(), 0)
	if current.Dossier.Ready {
		t.Fatal("new frontier minima ignored")
	}
	q.Now = p21Day(200)
	p21Same(t, before, p21Eval(t, l, q, 0))
	readerBehind := p21Query()
	readerBehind.Now = p21Day(1)
	r := p21Eval(t, l, readerBehind, 0)
	if !r.Dossier.AsOf.Equal(p21Day(40)) {
		t.Fatal("reader clock clipped provider frontier")
	}
	// Same counts and reason but a distinct cutoff is a distinct dossier.
	a := p21Query()
	ta := p21Day(29).Add(time.Hour)
	a.AsOf = &ta
	b := p21Query()
	tb := p21Day(29).Add(2 * time.Hour)
	b.AsOf = &tb
	ra := p21Eval(t, p21Ledger(), a, 0)
	rb := p21Eval(t, p21Ledger(), b, 0)
	if ra.Dossier.Stats != rb.Dossier.Stats {
		t.Fatal("fixture must have identical counts")
	}
	_, da := p21Canonical(t, ra)
	_, db := p21Canonical(t, rb)
	if da == db {
		t.Fatal("digest omitted resolved cutoff")
	}
	// Canonicalisation of the same immutable dossier is stable despite live publication metadata.
	extra := domain.ShadowPublication{ID: "diagnostics-only", ParentID: "policy-change", RepositoryID: "repo", RecordedAt: p21Day(50)}
	l.Publications = append(l.Publications, extra)
	p21Same(t, before, p21Eval(t, l, q, 0))

}

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
func p18Day(n int) time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(n) * 24 * time.Hour)
}
func p18Stamp(n int) *time.Time { x := p18Day(n); return &x }
func p18Fact(run string, day int) domain.PRRunFact {
	return domain.PRRunFact{ID: domain.AttemptID{RepositoryID: "repo", RunID: run, JobID: "1", Attempt: "1"}, Scope: domain.ShadowScopeCI,
		PRNumber: run, HeadCommit: "head-" + run, HeadTree: "tree-" + run, BaseCommit: "base", MergeBase: "mb", AssessedRevision: "merge",
		CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", CreatedAt: p18Day(day), StartedAt: p18Day(day), FinishedAt: p18Day(day).Add(time.Minute),
		Status: "assessed", Comparison: "comparable", Inspection: "inspected", Pins: domain.ShadowPins{Scout: "scout1", Tools: "tools1", Catalogue: "cat1", Policy: "policy1", Recipe: "recipe1"}, RecordedAt: p18Stamp(day + 1)}
}
func p18Ledger() domain.ShadowLedger {
	l := domain.ShadowLedger{Frontiers: []domain.ShadowFrontier{{RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", CompleteThrough: p18Day(30), PolicyRevision: "protected1", MinDays: 14, MinObservations: 5}},
		Intervals: []domain.ShadowInterval{{ID: "interval1", RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", After: p18Day(-1), Through: p18Day(30)}}}
	for i := 1; i <= 5; i++ {
		f := p18Fact(fmt.Sprint(i), i)
		f.RecordedAt = p18Stamp(30)
		l.Facts = append(l.Facts, f)
	}
	initial := l.Frontiers[0]
	initial.CompleteThrough = p18Day(0)
	l.Frontiers = append([]domain.ShadowFrontier{initial}, l.Frontiers...)
	l.Publications = []domain.ShadowPublication{{ID: "genesis", RepositoryID: "repo", RecordedAt: p18Day(0), Frontiers: l.Frontiers[:1]}, {ID: "sealed", ParentID: "genesis", RepositoryID: "repo", RecordedAt: p18Day(30), Facts: l.Facts, Frontiers: l.Frontiers[1:], Intervals: l.Intervals}}
	return l
}
func p18Query() Query {
	return Query{RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", Now: p18Day(60)}
}
func p18Eval(t *testing.T, l domain.ShadowLedger, q Query, code int) Report {
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
func p18Canonical(t *testing.T, r Report) ([]byte, string) {
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
func p18Same(t *testing.T, a, b Report) {
	t.Helper()
	ab, ad := p18Canonical(t, a)
	bb, bd := p18Canonical(t, b)
	if !bytes.Equal(ab, bb) || ad != bd {
		t.Fatalf("historical dossier changed: %s / %s", ad, bd)
	}
}
func p18Event(l *domain.ShadowLedger, f domain.PRRunFact, id, verdict string, day int) {
	e := domain.ProviderPREvent{EventID: id, RepositoryID: f.ID.RepositoryID, Scope: domain.ShadowScopeCI, PRNumber: f.PRNumber, HeadCommit: f.HeadCommit, HeadTree: f.HeadTree, MergeBase: f.MergeBase, CheckID: f.CheckID, RuleVersion: f.RuleVersion, IssueID: id, EventAt: f.StartedAt, RecordedAt: f.RecordedAt}
	l.Events = append(l.Events, e)
	l.Links = append(l.Links, domain.PRRunEventLink{AttemptID: f.ID, Scope: domain.ShadowScopeCI, EventID: id})
	if verdict != "" {
		l.Triage = append(l.Triage, domain.ShadowTriage{ActionID: fmt.Sprint(day + 100), EventID: id, RepositoryID: "repo", Scope: domain.ShadowScopeCI, Verdict: verdict, Actor: "maintainer", PolicyRevision: "protected1", TriagedAt: p18Day(day), RecordedAt: p18Stamp(day)})
	}
}
func p18Advance(l *domain.ShadowLedger, day int) {
	f := l.Frontiers[len(l.Frontiers)-1]
	f.CompleteThrough = p18Day(day)
	l.Frontiers = append(l.Frontiers, f)
	for i := range l.Intervals {
		l.Intervals[i].Through = p18Day(day)
	}
}
func p18Veto(t *testing.T, r Report, word string) {
	t.Helper()
	if r.Dossier.Ready || !strings.Contains(strings.ToLower(r.Dossier.Reason), word) {
		t.Fatalf("ready=%v reason=%q; want %s veto", r.Dossier.Ready, r.Dossier.Reason, word)
	}
}

// Excludes S3/S4 integration: base-family-gap attribution (including OR/double-count),
// and derivation of inspection from per-unit/base/inline-guardrail assessment coverage.
func TestADR012RepresentativesCountsAndPins(t *testing.T) {

	t.Run("representative-and-origin", func(t *testing.T) {
		l := p18Ledger()
		l.Facts = nil
		a := p18Fact("1", 0)
		a.Comparison = "none"
		a.Inspection = "not_inspected"
		b := a
		b.ID.RunID = "2"
		b.StartedAt = p18Day(10)
		b.FinishedAt = p18Day(10).Add(time.Minute)
		b.Comparison = "comparable"
		b.Inspection = "inspected"
		b.RecordedAt = p18Stamp(11)
		l.Facts = []domain.PRRunFact{a, b}
		r := p18Eval(t, l, p18Query(), 0)
		if r.Dossier.Stats.TotalRuns != 1 || r.Dossier.Stats.PRRuns != 1 || r.Dossier.Stats.Days != 20 || r.Dossier.Origin == nil || !r.Dossier.Origin.Equal(p18Day(10)) {
			t.Fatalf("representative/origin: %+v origin=%v", r.Dossier.Stats, r.Dossier.Origin)
		}
		p18Event(&l, a, "older-event", "not_warranted", 12)
		p18Veto(t, p18Eval(t, l, p18Query(), 0), "not_warranted")
		for _, inspection := range []string{"not_inspected", "not_applicable"} {
			later := b
			later.ID.RunID = "3"
			later.StartedAt = p18Day(15)
			later.Inspection = inspection
			l.Facts = append([]domain.PRRunFact{a, b}, later)
			p18Eval(t, l, p18Query(), 2)
		}
	})
	t.Run("dedup-trees-and-distinct-prs", func(t *testing.T) {
		l := p18Ledger()
		f := l.Facts[0]
		f.ID.RunID = "100"
		f.StartedAt = p18Day(10)
		l.Facts = append(l.Facts, f)
		r := p18Eval(t, l, p18Query(), 0)
		if r.Dossier.Stats.TotalRuns != 5 || r.Dossier.DistinctPRs != 5 {
			t.Fatalf("rerun inflated: %+v", r.Dossier)
		}
		f.ID.RunID = "101"
		f.HeadTree = "distinct-tree"
		l.Facts = append(l.Facts, f)
		r = p18Eval(t, l, p18Query(), 0)
		if r.Dossier.Stats.TotalRuns != 6 || r.Dossier.DistinctPRs != 5 {
			t.Fatal("distinct tree/PR accounting incorrect")
		}
	})
	t.Run("unknown-boundary-and-floor", func(t *testing.T) {
		l := p18Ledger()
		for i := 6; i <= 10; i++ {
			l.Facts = append(l.Facts, p18Fact(fmt.Sprint(i), i))
		}
		l.Facts[9].Comparison = "none"
		r := p18Eval(t, l, p18Query(), 0)
		if r.Dossier.Stats.UnknownAttribution != 1 || r.Dossier.Stats.TotalRuns != 10 || !r.Dossier.Ready {
			t.Fatalf("10%% equality must pass: %+v", r.Dossier)
		}
		q := p18Query()
		cutoff := p18Day(29).Add(23 * time.Hour)
		q.AsOf = &cutoff
		r = p18Eval(t, l, q, 0)
		if r.Dossier.Stats.Days != 28 {
			t.Fatalf("Days must floor at as_of: %d", r.Dossier.Stats.Days)
		}
	})
	t.Run("inspection-exclusions", func(t *testing.T) {
		for _, status := range []string{"not_inspected", "not_applicable"} {
			l := p18Ledger()
			early := p18Fact("99", 0)
			early.Inspection = status
			l.Facts = append(l.Facts, early)
			r := p18Eval(t, l, p18Query(), 0)
			if r.Dossier.Stats.PRRuns != 5 || !r.Dossier.Origin.Equal(p18Day(1)) {
				t.Fatalf("%s inflated exposure", status)
			}
		}
	})
	t.Run("mixed-pins-and-rollback", func(t *testing.T) {
		l := p18Ledger()
		for i, key := range []string{"scout", "tools", "catalogue", "policy"} {
			f := p18Fact(fmt.Sprint(100+i), 10+i)
			switch key {
			case "scout":
				f.Pins.Scout = "scout2"
			case "tools":
				f.Pins.Tools = "tools2"
			case "catalogue":
				f.Pins.Catalogue = "cat2"
			case "policy":
				f.Pins.Policy = "policy2"
			}
			l.Facts = append(l.Facts, f)
		}
		r := p18Eval(t, l, p18Query(), 0)
		for _, key := range []string{"scout", "tools", "catalogue", "policy"} {
			if len(r.Dossier.PinMixtures[key]) != 2 {
				t.Fatalf("%s mixtures not disclosed: %v", key, r.Dossier.PinMixtures)
			}
		}
		f := p18Fact("200", 20)
		f.RuleVersion = "2"
		l.Facts = append(l.Facts, f)
		f.ID.RunID = "201"
		f.StartedAt = p18Day(21)
		f.RuleVersion = "1"
		l.Facts = append(l.Facts, f)
		r = p18Eval(t, l, p18Query(), 0)
		if r.Dossier.RuleVersion != "1" {
			t.Fatal("rollback must select latest pin, not maximum version")
		}
	})

}

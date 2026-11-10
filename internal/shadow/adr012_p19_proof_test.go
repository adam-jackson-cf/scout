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
func p19Day(n int) time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(n) * 24 * time.Hour)
}
func p19Stamp(n int) *time.Time { x := p19Day(n); return &x }
func p19Fact(run string, day int) domain.PRRunFact {
	return domain.PRRunFact{ID: domain.AttemptID{RepositoryID: "repo", RunID: run, JobID: "1", Attempt: "1"}, Scope: domain.ShadowScopeCI,
		PRNumber: run, HeadCommit: "head-" + run, HeadTree: "tree-" + run, BaseCommit: "base", MergeBase: "mb", AssessedRevision: "merge",
		CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", CreatedAt: p19Day(day), StartedAt: p19Day(day), FinishedAt: p19Day(day).Add(time.Minute),
		Status: "assessed", Comparison: "comparable", Inspection: "inspected", Pins: domain.ShadowPins{Scout: "scout1", Tools: "tools1", Catalogue: "cat1", Policy: "policy1", Recipe: "recipe1"}, RecordedAt: p19Stamp(day + 1)}
}
func p19Ledger() domain.ShadowLedger {
	l := domain.ShadowLedger{Frontiers: []domain.ShadowFrontier{{RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", CompleteThrough: p19Day(30), PolicyRevision: "protected1", MinDays: 14, MinObservations: 5}},
		Intervals: []domain.ShadowInterval{{ID: "interval1", RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", RuleVersion: "1", After: p19Day(-1), Through: p19Day(30)}}}
	for i := 1; i <= 5; i++ {
		f := p19Fact(fmt.Sprint(i), i)
		f.RecordedAt = p19Stamp(30)
		l.Facts = append(l.Facts, f)
	}
	initial := l.Frontiers[0]
	initial.CompleteThrough = p19Day(0)
	l.Frontiers = append([]domain.ShadowFrontier{initial}, l.Frontiers...)
	l.Publications = []domain.ShadowPublication{{ID: "genesis", RepositoryID: "repo", RecordedAt: p19Day(0), Frontiers: l.Frontiers[:1]}, {ID: "sealed", ParentID: "genesis", RepositoryID: "repo", RecordedAt: p19Day(30), Facts: l.Facts, Frontiers: l.Frontiers[1:], Intervals: l.Intervals}}
	return l
}
func p19Query() Query {
	return Query{RepositoryID: "repo", CheckID: "security.ci-mutable-action-ref", Now: p19Day(60)}
}
func p19Eval(t *testing.T, l domain.ShadowLedger, q Query, code int) Report {
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
func p19Canonical(t *testing.T, r Report) ([]byte, string) {
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
func p19Same(t *testing.T, a, b Report) {
	t.Helper()
	ab, ad := p19Canonical(t, a)
	bb, bd := p19Canonical(t, b)
	if !bytes.Equal(ab, bb) || ad != bd {
		t.Fatalf("historical dossier changed: %s / %s", ad, bd)
	}
}
func p19Event(l *domain.ShadowLedger, f domain.PRRunFact, id, verdict string, day int) {
	e := domain.ProviderPREvent{EventID: id, RepositoryID: f.ID.RepositoryID, Scope: domain.ShadowScopeCI, PRNumber: f.PRNumber, HeadCommit: f.HeadCommit, HeadTree: f.HeadTree, MergeBase: f.MergeBase, CheckID: f.CheckID, RuleVersion: f.RuleVersion, IssueID: id, EventAt: f.StartedAt, RecordedAt: f.RecordedAt}
	l.Events = append(l.Events, e)
	l.Links = append(l.Links, domain.PRRunEventLink{AttemptID: f.ID, Scope: domain.ShadowScopeCI, EventID: id})
	if verdict != "" {
		l.Triage = append(l.Triage, domain.ShadowTriage{ActionID: fmt.Sprint(day + 100), EventID: id, RepositoryID: "repo", Scope: domain.ShadowScopeCI, Verdict: verdict, Actor: "maintainer", PolicyRevision: "protected1", TriagedAt: p19Day(day), RecordedAt: p19Stamp(day)})
	}
}
func p19Advance(l *domain.ShadowLedger, day int) {
	f := l.Frontiers[len(l.Frontiers)-1]
	f.CompleteThrough = p19Day(day)
	l.Frontiers = append(l.Frontiers, f)
	for i := range l.Intervals {
		l.Intervals[i].Through = p19Day(day)
	}
}
func p19Veto(t *testing.T, r Report, word string) {
	t.Helper()
	if r.Dossier.Ready || !strings.Contains(strings.ToLower(r.Dossier.Reason), word) {
		t.Fatalf("ready=%v reason=%q; want %s veto", r.Dossier.Ready, r.Dossier.Reason, word)
	}
}

// Export/import round-trip ordering is the S2 integration part of P19.
func TestADR012NumericOrderingAndEncoding(t *testing.T) {

	for _, field := range []string{"run", "attempt", "job"} {
		t.Run(field, func(t *testing.T) {
			var stable []byte
			var stableDigest string
			for _, reverse := range []bool{false, true} {
				l := p19Ledger()
				a := p19Fact("1", 20)
				b := a
				a.RuleVersion = "old"
				b.RuleVersion = "1"
				switch field {
				case "run":
					a.ID.RunID = "999"
					b.ID.RunID = "1000"
				case "attempt":
					a.ID.Attempt = "999"
					b.ID.Attempt = "1000"
				case "job":
					a.ID.JobID = "999"
					b.ID.JobID = "1000"
				}
				if reverse {
					l.Facts = append(l.Facts, b, a)
				} else {
					l.Facts = append(l.Facts, a, b)
				}
				r := p19Eval(t, l, p19Query(), 0)
				if r.Dossier.RuleVersion != "1" {
					t.Fatal("numeric 1000 must follow 999 for default cohort")
				}
				cb, cd := p19Canonical(t, r)
				if stable == nil {
					stable = cb
					stableDigest = cd
				} else if !bytes.Equal(stable, cb) || stableDigest != cd {
					t.Fatal("input order changed canonical cohort")
				}
				a.RuleVersion = "1"
				a.Inspection = "not_inspected"
				l.Facts = append(p19Ledger().Facts, a, b)
				r = p19Eval(t, l, p19Query(), 0)
				found := false
				for _, id := range r.Dossier.Representatives {
					if id == b.ID {
						found = true
					}
					if id == a.ID {
						t.Fatal("numeric 999 incorrectly representative")
					}
				}
				if !found {
					t.Fatal("numeric 1000 representative missing")
				}
				b1, d1 := p19Canonical(t, r)
				b2, d2 := p19Canonical(t, r)
				if !bytes.Equal(b1, b2) || d1 != d2 || r.Dossier.EncodingVersion == "" {
					t.Fatal("canonical encoding must be retained and deterministic")
				}
				q := p19Query()
				q.EncodingVersion = r.Dossier.EncodingVersion
				p19Same(t, r, p19Eval(t, l, q, 0))
			}
		})
	}
	for _, bad := range []string{"", "unavailable", "1.0", "9007199254740993x"} {
		l := p19Ledger()
		l.Facts[0].ID.RunID = bad
		p19Eval(t, l, p19Query(), 2)
	}
	l := p19Ledger()
	l.Facts[0].ID.RunID = "9007199254740993"
	l.Facts[1].ID.RunID = "9007199254740994"
	p19Eval(t, l, p19Query(), 0)
	q := p19Query()
	q.EncodingVersion = "unsupported-future-format"
	r, _ := Evaluate(p19Ledger(), q)
	if r.ExitCode != 3 {
		t.Fatalf("unknown encoding exit=%d want 3", r.ExitCode)
	}
	d := p19Eval(t, p19Ledger(), p19Query(), 0).Dossier
	d.EncodingVersion = "unsupported-future-format"
	if _, _, err := CanonicalDossier(d); err == nil {
		t.Fatal("canonical encoder silently replaced unknown immutable format version")
	}

}

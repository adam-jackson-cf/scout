package issues

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"scout/internal/domain"
	"scout/internal/state"
	"scout/internal/state/statetest"
)

const testCategory = "security:test"

type finding struct {
	path, symbol, quote string
	line                int
	mechanism           string
	digest              string
}

// harness persists runs and correlates them the way an assessment does.
type harness struct {
	t     *testing.T
	ctx   context.Context
	store *state.Store
	runs  int
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	store, err := state.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	return &harness{t: t, ctx: context.Background(), store: store}
}

// run saves a partial run whose coverage rows are given per path (status
// flagged where a finding exists, otherwise the supplied status) and correlates it.
func (h *harness) run(findings []finding, coverage map[string]domain.CoverageStatus, opts ...Option) []domain.Issue {
	h.t.Helper()
	h.runs++
	suffix := fmt.Sprint(h.runs)
	snapshot := domain.Snapshot{RepositoryID: "repository", WorktreeDigest: "worktree-" + suffix, IgnoreRulesDigest: "ignore", ExclusionsDigest: "exclude"}
	scope := domain.Scope{Kind: domain.ScopeRepository, SelectedFamilies: []domain.Family{domain.FamilySecurity}}
	plan := domain.Plan{PlanID: "plan-" + suffix, RunID: "run-" + suffix, Snapshot: snapshot, Scope: scope, CreatedAt: time.Now().UTC()}
	observationsByPath := map[string][]string{}
	observations := make([]domain.Observation, 0, len(findings))
	for index, f := range findings {
		mechanism := f.mechanism
		if mechanism == "" {
			mechanism = "unsafe path"
		}
		id := fmt.Sprintf("observation-%s-%d", suffix, index)
		observations = append(observations, domain.Observation{
			ObservationID: id, RunID: plan.RunID, Source: "scanner", SourceFindingID: f.digest, Family: domain.FamilySecurity,
			OriginalCategory: testCategory, RuleID: "rule", RuleVersion: "1", SourceSeverity: domain.SeverityHigh,
			EffectiveSeverity: domain.SeverityHigh, Confidence: domain.ConfidenceHigh, Mechanism: mechanism,
			Evidence:  []domain.Evidence{{Location: domain.SourceRange{Path: f.path, StartLine: f.line, EndLine: f.line, Symbol: f.symbol}, RedactedQuote: f.quote}},
			Validated: true, ObservedAt: time.Now().UTC(),
		})
		observationsByPath[f.path] = append(observationsByPath[f.path], id)
	}
	run := domain.Run{RunID: plan.RunID, PlanID: plan.PlanID, Snapshot: snapshot, Scope: scope, State: domain.RunPartial, StartedAt: time.Now().UTC()}
	for path, status := range coverage {
		row := domain.Coverage{Key: domain.CoverageKey{FilePath: path, Family: domain.FamilySecurity, Category: testCategory}, Status: status, Reason: "test"}
		if ids := observationsByPath[path]; len(ids) != 0 {
			row.Status, row.ObservationIDs = domain.CoverageFlagged, ids
		}
		if status == domain.CoverageSkipped || status == domain.CoverageIncomplete || status == domain.CoverageBlocked {
			row.Reason = "not inspected"
		} else {
			row.InspectedRanges = []domain.SourceRange{{Path: path, StartLine: 1, EndLine: 1}}
		}
		if status == domain.CoverageReused {
			row.ReusedFromRunID = "run-0"
		}
		run.Coverage = append(run.Coverage, row)
	}
	statetest.SeedRun(h.t, h.ctx, h.store, plan, run, observations, nil)
	issues, err := Correlate(h.ctx, h.store, run, observations, opts...)
	if err != nil {
		h.t.Fatal(err)
	}
	return issues
}

func (h *harness) all() []domain.Issue {
	h.t.Helper()
	issues, err := h.store.ListIssues(h.ctx)
	if err != nil {
		h.t.Fatal(err)
	}
	return issues
}

func (h *harness) load(id string) domain.Issue {
	h.t.Helper()
	issue, err := h.store.LoadIssue(h.ctx, id)
	if err != nil {
		h.t.Fatal(err)
	}
	return issue
}

func hasEvent(issue domain.Issue, event string) bool {
	for _, entry := range issue.History {
		if entry.Event == event {
			return true
		}
	}
	return false
}

func covered(paths ...string) map[string]domain.CoverageStatus {
	rows := map[string]domain.CoverageStatus{}
	for _, path := range paths {
		rows[path] = domain.CoverageInspected
	}
	return rows
}

func TestIssueIDIsOpaqueULID(t *testing.T) {
	h := newHarness(t)
	issues := h.run([]finding{{path: "a.go", quote: "eval(x)", line: 3}}, covered("a.go"))
	id := issues[0].IssueID
	if !strings.HasPrefix(id, "issue:security:") || len(strings.TrimPrefix(id, "issue:security:")) != 26 {
		t.Fatalf("issue ID is not issue:<family>:<ulid>: %q", id)
	}
}

func TestIdentityStableAcrossInsertionAboveAndRename(t *testing.T) {
	for _, symbol := range []string{"", "pkg.Handle"} {
		t.Run("symbol="+symbol, func(t *testing.T) {
			h := newHarness(t)
			first := h.run([]finding{{path: "a.go", symbol: symbol, quote: "eval(x)", line: 10}}, covered("a.go"))
			// 20 lines inserted above the flagged function: only the locator moves.
			second := h.run([]finding{{path: "a.go", symbol: symbol, quote: "eval(x)", line: 30}}, covered("a.go"))
			// The file is renamed (content-stable rename from the inventory diff).
			third := h.run([]finding{{path: "b.go", symbol: symbol, quote: "eval(x)", line: 31}}, covered("b.go"), WithRenames(map[string]string{"a.go": "b.go"}))
			if first[0].IssueID != second[0].IssueID || first[0].IssueID != third[0].IssueID {
				t.Fatalf("identity changed: %q %q %q", first[0].IssueID, second[0].IssueID, third[0].IssueID)
			}
			if all := h.all(); len(all) != 1 {
				t.Fatalf("issues = %d, want 1: %#v", len(all), all)
			}
			issue := h.load(first[0].IssueID)
			if issue.Anchor.Path != "b.go" || issue.State != domain.IssueOpen || hasEvent(issue, domain.EventEvidenceDrift) || len(issue.ObservationIDs) != 3 {
				t.Fatalf("anchor did not follow the rename cleanly: %#v", issue)
			}
		})
	}
}

func TestEvidenceDriftKeepsDisposition(t *testing.T) {
	for _, symbol := range []string{"", "pkg.Handle"} {
		t.Run("symbol="+symbol, func(t *testing.T) {
			h := newHarness(t)
			created := h.run([]finding{{path: "a.go", symbol: symbol, quote: "eval(x)", line: 10}}, covered("a.go"))[0]
			if _, err := h.store.TriageIssue(h.ctx, created.IssueID, domain.IssueAccepted, "owner", "accepted risk", "human", "reviewed"); err != nil {
				t.Fatal(err)
			}
			drifted := h.run([]finding{{path: "a.go", symbol: symbol, quote: "eval(x, y)", line: 12}}, covered("a.go"))[0]
			if drifted.IssueID != created.IssueID {
				t.Fatalf("drift created a new issue: %q != %q", drifted.IssueID, created.IssueID)
			}
			issue := h.load(created.IssueID)
			if issue.State != domain.IssueAccepted || issue.DispositionReason != "accepted risk" || issue.Owner != "owner" {
				t.Fatalf("disposition not intact: %#v", issue)
			}
			if !hasEvent(issue, domain.EventEvidenceDrift) {
				t.Fatalf("no evidence_drift history entry: %#v", issue.History)
			}
			if again := h.run([]finding{{path: "a.go", symbol: symbol, quote: "eval(x, y)", line: 12}}, covered("a.go"))[0]; again.IssueID != created.IssueID {
				t.Fatal("drifted anchor was not adopted")
			}
			if len(h.all()) != 1 {
				t.Fatal("drift split the issue")
			}
		})
	}
}

func TestSameRuleOccurrencesInOneFileAreDistinctIssues(t *testing.T) {
	for name, quotes := range map[string][2]string{"different evidence": {"eval(a)", "eval(b)"}, "identical evidence": {"eval(a)", "eval(a)"}} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t)
			findings := []finding{{path: "a.go", quote: quotes[0], line: 5}, {path: "a.go", quote: quotes[1], line: 20}}
			first := h.run(findings, covered("a.go"))
			if len(first) != 2 || first[0].IssueID == first[1].IssueID {
				t.Fatalf("two occurrences share one issue: %#v", first)
			}
			second := h.run([]finding{{path: "a.go", quote: quotes[0], line: 25}, {path: "a.go", quote: quotes[1], line: 40}}, covered("a.go"))
			if len(h.all()) != 2 || second[0].IssueID != first[0].IssueID || second[1].IssueID != first[1].IssueID {
				t.Fatalf("occurrence identities not stable: %#v vs %#v", first, second)
			}
		})
	}
}

func TestResolveOnlyAfterRelevantCheckThenRegress(t *testing.T) {
	h := newHarness(t)
	id := h.run([]finding{{path: "a.go", quote: "eval(x)", line: 10}}, covered("a.go"))[0].IssueID

	// Absence from skipped, incomplete, reused or unrelated coverage never resolves.
	for _, status := range []domain.CoverageStatus{domain.CoverageSkipped, domain.CoverageIncomplete, domain.CoverageReused} {
		h.run(nil, map[string]domain.CoverageStatus{"a.go": status})
		if state := h.load(id).State; state != domain.IssueOpen {
			t.Fatalf("%s coverage changed state to %q", status, state)
		}
	}
	h.run(nil, covered("other.go"))
	if state := h.load(id).State; state != domain.IssueOpen {
		t.Fatalf("unrelated coverage changed state to %q", state)
	}

	// The code is deleted and the file is successfully re-inspected.
	h.run(nil, covered("a.go"))
	resolved := h.load(id)
	if resolved.State != domain.IssueResolved || resolved.Verification == nil || !hasEvent(resolved, domain.EventResolved) {
		t.Fatalf("not resolved after relevant successful check: %#v", resolved)
	}

	// The code is re-introduced: the same durable ID regresses.
	back := h.run([]finding{{path: "a.go", quote: "eval(x)", line: 7}}, covered("a.go"))[0]
	regressed := h.load(id)
	if back.IssueID != id || regressed.State != domain.IssueRegressed || regressed.Verification != nil || !hasEvent(regressed, domain.EventRegressed) {
		t.Fatalf("re-introduced finding did not regress the resolved issue: %#v", regressed)
	}
	if len(h.all()) != 1 {
		t.Fatal("regression created a second issue")
	}
}

func TestAmbiguousCandidatesAreNeverMerged(t *testing.T) {
	h := newHarness(t)
	first := h.run([]finding{{path: "a.go", quote: "eval(a)", line: 5}, {path: "a.go", quote: "eval(b)", line: 20}}, covered("a.go"))
	originals := []string{first[0].IssueID, first[1].IssueID}

	// One occurrence with new evidence while both originals are unmatched: the
	// ladder cannot tell which one it is.
	created := h.run([]finding{{path: "a.go", quote: "eval(c)", line: 8}}, covered("a.go"))
	if len(created) != 1 || created[0].IssueID == originals[0] || created[0].IssueID == originals[1] {
		t.Fatalf("ambiguous occurrence was merged into a durable issue: %#v", created)
	}
	if len(h.all()) != 3 {
		t.Fatalf("issues = %d, want 3", len(h.all()))
	}
	newcomer := h.load(created[0].IssueID)
	if len(newcomer.AmbiguousWith) != 2 || !hasEvent(newcomer, domain.EventAmbiguous) || newcomer.State != domain.IssueOpen {
		t.Fatalf("new issue lacks ambiguous_with: %#v", newcomer)
	}
	for _, id := range originals {
		original := h.load(id)
		if original.State != domain.IssueOpen || len(original.AmbiguousWith) != 1 || original.AmbiguousWith[0] != newcomer.IssueID {
			t.Fatalf("original issue %s must stay open and reference the ambiguity: %#v", id, original)
		}
	}
}

func TestSecretMatchesByDigestAcrossMoves(t *testing.T) {
	h := newHarness(t)
	secret := func(path string, line int) finding {
		return finding{path: path, line: line, quote: "[REDACTED]", mechanism: domain.SecretMechanism, digest: "sha256-of-secret"}
	}
	first := h.run([]finding{secret("config/a.env", 3)}, covered("config/a.env"))[0]
	// Moved to another file with no rename information: still the same issue.
	moved := h.run([]finding{secret("deploy/b.env", 9)}, covered("deploy/b.env"))[0]
	if first.IssueID != moved.IssueID || len(h.all()) != 1 {
		t.Fatalf("moved secret became a new issue: %q vs %q", first.IssueID, moved.IssueID)
	}
	if h.load(first.IssueID).Anchor.Path != "deploy/b.env" {
		t.Fatal("secret anchor did not follow the move")
	}
}

func TestCorrelationIsIdempotentForOneRun(t *testing.T) {
	h := newHarness(t)
	findings := []finding{{path: "a.go", quote: "eval(a)", line: 5}, {path: "a.go", quote: "eval(b)", line: 20}}
	issues := h.run(findings, covered("a.go"))
	before := h.load(issues[0].IssueID)
	// Correlating the persisted run again must not change any issue.
	runs, err := h.store.LoadRun(h.ctx, "run-1")
	if err != nil {
		t.Fatal(err)
	}
	observations, err := h.store.LoadObservations(h.ctx, "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Correlate(h.ctx, h.store, runs, observations); err != nil {
		t.Fatal(err)
	}
	after := h.load(issues[0].IssueID)
	if len(h.all()) != 2 || len(after.History) != len(before.History) {
		t.Fatalf("second correlation changed state: %d issues, history %d -> %d", len(h.all()), len(before.History), len(after.History))
	}
}

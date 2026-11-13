package baseline

import (
	"encoding/json"
	"os/exec"
	"reflect"
	"scout/internal/domain"
	"scout/internal/state"
	"strings"
	"testing"
	"time"
)

func p11Git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	c := exec.CommandContext(boundedTestContext(t), "git", args...)
	c.Dir = dir
	b, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, b)
	}
	return strings.TrimSpace(string(b))
}
func p11Remote(t *testing.T) (string, []Options) {
	t.Helper()
	bare := t.TempDir()
	p11Git(t, bare, "init", "--bare", "--quiet")
	opts := []Options{}
	for range 2 {
		dir := t.TempDir()
		p11Git(t, dir, "init", "--quiet")
		p11Git(t, dir, "remote", "add", "origin", bare)
		opts = append(opts, Options{Repo: dir, Remote: bare, Trusted: true, Now: func() time.Time { return time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC) }})
	}
	return bare, opts
}
func p11Tree(t *testing.T, l domain.ShadowLedger) Tree {
	t.Helper()
	files, err := state.EncodeShadow(l, domain.ShadowInputs{})
	if err != nil {
		t.Fatal(err)
	}
	files[state.ExportIssues] = nil
	files[state.ExportDispositions] = nil
	files[state.ExportWouldBlock] = nil
	files[state.ExportCoverageSummary] = []byte("{}\n")
	files[state.ExportManifest] = state.RenderManifest(state.ManifestMeta{RepositoryID: "repo", WouldBlock: state.WouldBlockPresent}, files)
	return Tree(files)
}
func p11Decode(t *testing.T, raw string, dst any) {
	t.Helper()
	if err := json.Unmarshal([]byte(raw), dst); err != nil {
		t.Fatal(err)
	}
}
func p11Ledger() domain.ShadowLedger {
	at := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	stamp := at.Add(time.Hour)
	f := domain.PRRunFact{ID: domain.AttemptID{RepositoryID: "repo", RunID: "999", JobID: "1000", Attempt: "1"}, Scope: domain.ShadowScopeCI, PRNumber: "17", HeadCommit: "H", HeadTree: "tree", BaseCommit: "B", MergeBase: "MB", AssessedRevision: "merge", CheckID: "security:test", RuleVersion: "v1", CreatedAt: at.Add(-time.Minute), StartedAt: at, FinishedAt: at.Add(time.Minute), Status: "assessed", Comparison: "comparable", Inspection: "inspected", Pins: domain.ShadowPins{Scout: "scout", Tools: "tools", Catalogue: "catalogue", Policy: "policy", Recipe: "guardian"}, RecordedAt: &stamp}
	e := domain.ProviderPREvent{EventID: "event", RepositoryID: "repo", Scope: domain.ShadowScopeCI, PRNumber: "17", HeadCommit: "H", HeadTree: "tree", MergeBase: "MB", CheckID: "security:test", IssueID: "issue", RuleVersion: "v1", EventAt: at, RecordedAt: &stamp}
	link := domain.PRRunEventLink{AttemptID: f.ID, Scope: domain.ShadowScopeCI, EventID: e.EventID}
	c := domain.ShadowFrontier{RepositoryID: "repo", CheckID: "security:test", RuleVersion: "v1", CompleteThrough: at.Add(-time.Hour), PolicyRevision: "policy", MinDays: 14, MinObservations: 5}
	pub := domain.ShadowPublication{ID: "genesis", RepositoryID: "repo", RecordedAt: stamp, Facts: []domain.PRRunFact{f}, Events: []domain.ProviderPREvent{e}, Links: []domain.PRRunEventLink{link}, Frontiers: []domain.ShadowFrontier{c}}
	return domain.ShadowLedger{Facts: pub.Facts, Events: pub.Events, Links: pub.Links, Frontiers: pub.Frontiers, Publications: []domain.ShadowPublication{pub}}
}

// Baseline half of P11: the retry supplies actual fetched commit AND tree.
// The same-named shadow proof owns stamps, current-tip policy and frontier
// decisions, avoiding a forbidden baseline->shadow dependency.
func TestADR012CASParentStampAndFrontier(t *testing.T) {
	_, opts := p11Remote(t)
	ctx := boundedTestContext(t)
	initial := p11Ledger()
	seed, err := Publish(ctx, opts[0], func(Parent) (Tree, error) { return p11Tree(t, initial), nil })
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	var winner Result
	var seen []string
	_, err = Publish(ctx, opts[0], func(parent Parent) (Tree, error) {
		calls++
		seen = append(seen, parent.Commit)
		current, _, err := state.DecodeShadow(parent.Tree)
		if err != nil {
			return nil, err
		}
		if calls == 1 {
			competing := p11Ledger()
			stamp := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
			c := competing.Frontiers[0]
			c.CompleteThrough = stamp.Add(time.Hour)
			c.PolicyRevision = "tip-2"
			pub := domain.ShadowPublication{ID: "winner", ParentID: "genesis", RepositoryID: "repo", RecordedAt: stamp, Frontiers: []domain.ShadowFrontier{c}}
			competing.Frontiers = append(competing.Frontiers, c)
			competing.Publications = append(competing.Publications, pub)
			winner, err = Publish(ctx, opts[1], func(Parent) (Tree, error) { return p11Tree(t, competing), nil })
			if err != nil {
				return nil, err
			}
		} else {
			if current.Frontiers[len(current.Frontiers)-1].PolicyRevision != "tip-2" || !current.Publications[len(current.Publications)-1].RecordedAt.Equal(time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)) {
				t.Fatalf("retry tree stale: %+v", current)
			}
		}
		// A real change forces CAS even though another writer changes the parent
		// while this callback is executing; no goroutine/timing-dependent race.
		next := p11Tree(t, current)
		next[state.ExportCoverageSummary] = []byte("{\"changed\":true}\n")
		next[state.ExportManifest] = state.RenderManifest(state.ManifestMeta{RepositoryID: "repo", WouldBlock: state.WouldBlockPresent}, next)
		return next, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || !reflect.DeepEqual(seen, []string{seed.Commit, winner.Commit}) {
		t.Fatalf("actual fetched parent not supplied on retry: %v", seen)
	}
}

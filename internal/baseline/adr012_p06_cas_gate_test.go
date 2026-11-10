package baseline

import (
	"context"
	"encoding/json"
	"os/exec"
	"reflect"
	"scout/internal/domain"
	"scout/internal/state"
	"strings"
	"testing"
	"time"
)

func p06Git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	c := exec.CommandContext(boundedTestContext(t), "git", args...)
	c.Dir = dir
	b, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, b)
	}
	return strings.TrimSpace(string(b))
}
func p06Remote(t *testing.T) (string, []Options) {
	t.Helper()
	bare := t.TempDir()
	p06Git(t, bare, "init", "--bare", "--quiet")
	opts := []Options{}
	for range 2 {
		dir := t.TempDir()
		p06Git(t, dir, "init", "--quiet")
		p06Git(t, dir, "remote", "add", "origin", bare)
		opts = append(opts, Options{Repo: dir, Remote: bare, Trusted: true, Now: func() time.Time { return time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC) }})
	}
	return bare, opts
}
func p06Tree(t *testing.T, l domain.ShadowLedger) Tree {
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
func p06Decode(t *testing.T, raw string, dst any) {
	t.Helper()
	if err := json.Unmarshal([]byte(raw), dst); err != nil {
		t.Fatal(err)
	}
}
func p06Ledger() domain.ShadowLedger {
	at := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	stamp := at.Add(time.Hour)
	f := domain.PRRunFact{ID: domain.AttemptID{RepositoryID: "repo", RunID: "999", JobID: "1000", Attempt: "1"}, Scope: domain.ShadowScopeCI, PRNumber: "17", HeadCommit: "H", HeadTree: "tree", BaseCommit: "B", MergeBase: "MB", AssessedRevision: "merge", CheckID: "security:test", RuleVersion: "v1", CreatedAt: at.Add(-time.Minute), StartedAt: at, FinishedAt: at.Add(time.Minute), Status: "assessed", Comparison: "comparable", Inspection: "inspected", Pins: domain.ShadowPins{Scout: "scout", Tools: "tools", Catalogue: "catalogue", Policy: "policy", Recipe: "guardian"}, RecordedAt: &stamp}
	e := domain.ProviderPREvent{EventID: "event", RepositoryID: "repo", Scope: domain.ShadowScopeCI, PRNumber: "17", HeadCommit: "H", HeadTree: "tree", MergeBase: "MB", CheckID: "security:test", IssueID: "issue", RuleVersion: "v1", EventAt: at, RecordedAt: &stamp}
	link := domain.PRRunEventLink{AttemptID: f.ID, Scope: domain.ShadowScopeCI, EventID: e.EventID}
	c := domain.ShadowFrontier{RepositoryID: "repo", CheckID: "security:test", RuleVersion: "v1", CompleteThrough: at.Add(-time.Hour), PolicyRevision: "policy", MinDays: 14, MinObservations: 5}
	pub := domain.ShadowPublication{ID: "genesis", RepositoryID: "repo", RecordedAt: stamp, Facts: []domain.PRRunFact{f}, Events: []domain.ProviderPREvent{e}, Links: []domain.PRRunEventLink{link}, Frontiers: []domain.ShadowFrontier{c}}
	return domain.ShadowLedger{Facts: pub.Facts, Events: pub.Events, Links: pub.Links, Frontiers: pub.Frontiers, Publications: []domain.ShadowPublication{pub}}
}

// Call through the real exported signature, accepting either the pre-cutover
// Tree argument or the frozen Parent argument. This keeps the compatibility
// defect behavioural-red today without a local stand-in for baseline.Parent.
func p06Publish(t *testing.T, ctx context.Context, o Options, apply func(Tree) (Tree, error)) (Result, error) {
	t.Helper()
	fn := reflect.ValueOf(Publish)
	ct := fn.Type().In(2)
	cb := reflect.MakeFunc(ct, func(args []reflect.Value) []reflect.Value {
		v := args[0]
		if v.Kind() == reflect.Struct {
			v = v.FieldByName("Tree")
		}
		next, err := apply(v.Interface().(Tree))
		ev := reflect.Zero(ct.Out(1))
		if err != nil {
			ev = reflect.ValueOf(err)
		}
		return []reflect.Value{reflect.ValueOf(next), ev}
	})
	out := fn.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(o), cb})
	r := out[0].Interface().(Result)
	if out[1].IsNil() {
		return r, nil
	}
	return r, out[1].Interface().(error)
}
func p06PushUnchecked(t *testing.T, ctx context.Context, o Options, tree Tree, parent string) string {
	t.Helper()
	if _, _, err := fetchTip(ctx, git{dir: o.Repo}, o); err != nil {
		t.Fatal(err)
	}
	t.Helper()
	g := git{dir: o.Repo}
	oid, err := writeTree(ctx, g, tree)
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"-c", "commit.gpgsign=false", "commit-tree", oid, "-m", "incompatible competitor"}
	if parent != "" {
		args = append(args, "-p", parent)
	}
	env := []string{"GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid", "GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid", "GIT_AUTHOR_DATE=1793534400 +0000", "GIT_COMMITTER_DATE=1793534400 +0000"}
	commit, err := g.output(ctx, env, args...)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.output(ctx, nil, "push", "--quiet", o.Remote, commit+":refs/heads/"+DefaultBranch); err != nil {
		t.Fatal(err)
	}
	return commit
}
func TestADR012CASCompatibilityGate(t *testing.T) {
	for _, schema := range []string{"scout.state/v1", "scout.state/v99"} {
		t.Run(schema, func(t *testing.T) {
			_, opts := p06Remote(t)
			ctx := boundedTestContext(t)
			good := p06Tree(t, p06Ledger())
			seed, err := p06Publish(t, ctx, opts[0], func(Tree) (Tree, error) { return good, nil })
			if err != nil {
				t.Fatal(err)
			}
			broken := Tree{}
			for k, v := range good {
				broken[k] = append([]byte(nil), v...)
			}
			broken[state.ExportManifest] = []byte(strings.Replace(string(good[state.ExportManifest]), "scout.state/v2", schema, 1))
			calls := 0
			competitor := ""
			_, err = p06Publish(t, ctx, opts[0], func(current Tree) (Tree, error) {
				calls++
				if calls == 1 {
					competitor = p06PushUnchecked(t, ctx, opts[1], broken, seed.Commit)
				}
				// A caller may replace rather than merge state: the Publish boundary owns
				// the parent compatibility gate, not just validation of this valid output.
				next := p06Tree(t, p06Ledger())
				next[state.ExportCoverageSummary] = []byte("{\"changed\":true}\n")
				next[state.ExportManifest] = state.RenderManifest(state.ManifestMeta{RepositoryID: "repo", WouldBlock: state.WouldBlockPresent}, next)
				return next, nil
			})
			if err == nil {
				t.Fatalf("CAS retry overwrote incompatible %s parent; callbacks=%d", schema, calls)
			}
			tip, tipErr := remoteTip(ctx, git{dir: opts[0].Repo}, opts[0])
			if tipErr != nil {
				t.Fatal(tipErr)
			}
			if tip != competitor {
				t.Fatalf("incompatible parent overwritten: %s want %s", tip, competitor)
			}
			if schema == "scout.state/v1" {
				if !strings.Contains(err.Error(), "last intact v2") || !strings.Contains(err.Error(), "reapply") {
					t.Fatalf("missing recover-and-reapply remedy: %v", err)
				}
			} else if !strings.Contains(err.Error(), "upgrade") {
				t.Fatalf("missing writer upgrade remedy: %v", err)
			}
		})
	}
	t.Run("empty_destination_still_validates_local", func(t *testing.T) {
		bad := p06Tree(t, domain.ShadowLedger{})
		bad[state.ExportPRRuns] = []byte("not json\n")
		if _, err := MergeState(Tree{}, bad); err == nil {
			t.Fatal("empty destination bypassed manifest validation")
		}
	})
}

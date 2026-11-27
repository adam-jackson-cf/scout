//go:build boundary

package engine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"scout/internal/artifacts"
	"scout/internal/domain"
	"scout/internal/inventory"
	"scout/internal/state"
	"scout/internal/state/statetest"
)

func writeCrapFixtureFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func crapFixtureRepo(t *testing.T) (root, source string) {
	t.Helper()
	root = t.TempDir()
	source = filepath.Join(root, "pkg", "core", "busy.py")
	if err := os.MkdirAll(filepath.Dir(source), 0o755); err != nil {
		t.Fatal(err)
	}
	writeCrapFixtureFile(t, source, busyPythonModule())
	gitRun(t, root, "init", "--quiet")
	gitRun(t, root, "add", ".")
	gitRun(t, root, "commit", "--quiet", "-m", "one")
	return root, source
}

func runCrapMeasures(t *testing.T, ctx context.Context, store *state.Store, root string, specs ...artifacts.Spec) string {
	t.Helper()
	scan, err := inventory.Scan(ctx, root, inventory.ScanOptions{Scope: domain.Scope{Kind: domain.ScopeRepository}})
	if err != nil {
		t.Fatal(err)
	}
	options := Options{Scope: domain.Scope{Kind: domain.ScopeRepository}, Measures: &MeasuresOptions{Artifacts: specs}}
	runMeasuresLanes(ctx, measuresLaneInput{store: store, root: root, current: scan, options: options})
	return scan.Snapshot.Revision
}

func singleCrapSignal(t *testing.T, ctx context.Context, store *state.Store, revision string) domain.Signal {
	t.Helper()
	signals, err := store.ListSignals(ctx, state.SignalFilter{MetricID: "scout.crap", ScopeAnchor: "pkg/core", Snapshot: revision})
	if err != nil || len(signals) != 1 {
		t.Fatalf("crap signals for %s = %+v, %v", revision, signals, err)
	}
	return signals[0]
}

func assertCrapWithMatchedCoverage(t *testing.T, ctx context.Context, store *state.Store, revision string) {
	t.Helper()
	signals, err := store.ListSignals(ctx, state.SignalFilter{MetricID: "scout.crap", ScopeAnchor: "pkg/core", Snapshot: revision})
	if err != nil || len(signals) != 2 {
		t.Fatalf("with and without coverage must be distinct rows: %+v, %v", signals, err)
	}
	var present domain.Signal
	for _, s := range signals {
		if s.Missing == domain.SignalMeasured {
			present = s
		}
	}
	// One function, CC 12, no covered line: CRAP = 12²·1 + 12 = 156 > 30.
	if present.Value == nil || *present.Value != 156 || *present.Numerator != 12 || *present.Denominator != 12 {
		t.Fatalf("crap with matched artifact = %+v", present)
	}
}

func assertMatchedArtifactProvenance(t *testing.T, ctx context.Context, store *state.Store, revision string) {
	t.Helper()
	provenance, err := statetest.IngestedArtifacts(ctx, store, revision)
	if err != nil || len(provenance) != 2 {
		t.Fatalf("provenance = %+v, %v", provenance, err)
	}
	for _, p := range provenance {
		if p.RevisionMatch != artifacts.MatchMatch || len(p.Digest) != 64 || p.Revision != revision {
			t.Fatalf("provenance row = %+v", p)
		}
	}
	assertJUnitProvenancePayload(t, provenance[0])
}

func assertJUnitProvenancePayload(t *testing.T, p state.IngestedArtifact) {
	t.Helper()
	if p.Kind != artifacts.KindJUnit || !strings.Contains(string(p.Payload), `"tests":2`) || !strings.Contains(string(p.Payload), `"failures":1`) {
		t.Fatalf("junit provenance = %+v", p)
	}
}

func assertMismatchedArtifactIsUnmeasured(t *testing.T, ctx context.Context, store *state.Store, revision string) {
	t.Helper()
	signals, err := store.ListSignals(ctx, state.SignalFilter{MetricID: "scout.crap", ScopeAnchor: "pkg/core", Snapshot: revision})
	if err != nil || len(signals) != 1 || signals[0].Missing != domain.SignalUnmeasured || signals[0].Value != nil {
		t.Fatalf("mismatched artifact: crap = %+v, %v", signals, err)
	}
	// The row already recorded for this snapshot and tool digest is kept
	// (append-only); the mismatch reason lives in the artifact provenance.
	mismatch, err := statetest.IngestedArtifacts(ctx, store, revision)
	if err != nil || len(mismatch) != 1 || mismatch[0].RevisionMatch != artifacts.MatchMismatch || mismatch[0].Reason == "" {
		t.Fatalf("mismatch provenance = %+v, %v", mismatch, err)
	}
}

func TestCrapIsPresentOnlyForARevisionMatchedArtifactAndUnmeasuredOtherwise(t *testing.T) {
	requireLizardOnPath(t)
	ctx := context.Background()
	root, source := crapFixtureRepo(t)

	store := openEngineTestStore(t, root)

	// CI artifacts live outside the checkout: the tree stays clean.
	ci := t.TempDir()
	lcov := filepath.Join(ci, "coverage.lcov")
	writeCrapFixtureFile(t, lcov, "TN:\nSF:pkg/core/busy.py\nDA:1,0\nDA:2,0\nDA:3,0\nDA:4,0\nend_of_record\n")
	junit := filepath.Join(ci, "junit.xml")
	writeCrapFixtureFile(t, junit, `<testsuite name="s" tests="2" failures="1"><testcase name="a"/><testcase name="b"><failure/></testcase></testsuite>`)

	// Run 1: artifact produced at this commit.
	first := runCrapMeasures(t, ctx, store, root)
	if s := singleCrapSignal(t, ctx, store, first); s.Missing != domain.SignalUnmeasured || s.Value != nil {
		t.Fatalf("no artifact: crap = %+v", s)
	}
	writeCrapFixtureFile(t, lcov+artifacts.SidecarSuffix, `{"revision":"`+first+`","producer":"pytest-cov"}`)
	if again := runCrapMeasures(t, ctx, store, root, artifacts.Spec{Path: lcov}, artifacts.Spec{Path: junit, Revision: first}); again != first {
		t.Fatalf("snapshot changed: %s vs %s", again, first)
	}
	assertCrapWithMatchedCoverage(t, ctx, store, first)
	assertMatchedArtifactProvenance(t, ctx, store, first)

	// Run 2: the next commit, artifact not attached: unmeasured.
	writeCrapFixtureFile(t, source, busyPythonModule()+"\n# changed\n")
	gitRun(t, root, "commit", "--quiet", "-am", "two")
	second := runCrapMeasures(t, ctx, store, root)
	if second == first {
		t.Fatal("expected a new snapshot")
	}
	if s := singleCrapSignal(t, ctx, store, second); s.Missing != domain.SignalUnmeasured || s.Value != nil || s.Numerator != nil {
		t.Fatalf("next run without artifact: crap = %+v", s)
	}

	// Run 3: the stale artifact (produced at the first commit) is attached to the second.
	runCrapMeasures(t, ctx, store, root, artifacts.Spec{Path: lcov})
	assertMismatchedArtifactIsUnmeasured(t, ctx, store, second)
}

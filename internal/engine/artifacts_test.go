package engine

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"scout/internal/artifacts"
	"scout/internal/domain"
	"scout/internal/inventory"
	"scout/internal/state"
)

func TestCrapIsPresentOnlyForARevisionMatchedArtifactAndUnmeasuredOtherwise(t *testing.T) {
	if _, err := exec.LookPath("lizard"); err != nil {
		t.Skip("lizard is not on PATH (source .tools/activate)")
	}
	ctx := context.Background()
	root := t.TempDir()
	var module strings.Builder
	module.WriteString("def busy(x):\n    total = 0\n")
	for i := range 11 {
		module.WriteString("    if x == " + string(rune('a'+i)) + ":\n        total += 1\n")
	}
	module.WriteString("    return total\n")
	source := filepath.Join(root, "pkg", "core", "busy.py")
	if err := os.MkdirAll(filepath.Dir(source), 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(path, content string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(source, module.String())
	gitRun(t, root, "init", "--quiet")
	gitRun(t, root, "add", ".")
	gitRun(t, root, "commit", "--quiet", "-m", "one")

	store, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	run := func(specs ...artifacts.Spec) (string, []domain.LaneHealth) {
		t.Helper()
		scan, err := inventory.Scan(ctx, root, inventory.ScanOptions{Scope: domain.Scope{Kind: domain.ScopeRepository}})
		if err != nil {
			t.Fatal(err)
		}
		options := Options{Scope: domain.Scope{Kind: domain.ScopeRepository}, Measures: &MeasuresOptions{Artifacts: specs}}
		health := runMeasuresLanes(ctx, store, root, scan, nil, options)
		return scan.Snapshot.Revision, health
	}
	crap := func(revision string) domain.Signal {
		t.Helper()
		signals, err := store.ListSignals(ctx, state.SignalFilter{MetricID: "scout.crap", ScopeAnchor: "pkg/core", Snapshot: revision})
		if err != nil || len(signals) != 1 {
			t.Fatalf("crap signals for %s = %+v, %v", revision, signals, err)
		}
		return signals[0]
	}

	// CI artifacts live outside the checkout: the tree stays clean.
	ci := t.TempDir()
	lcov := filepath.Join(ci, "coverage.lcov")
	write(lcov, "TN:\nSF:pkg/core/busy.py\nDA:1,0\nDA:2,0\nDA:3,0\nDA:4,0\nend_of_record\n")
	junit := filepath.Join(ci, "junit.xml")
	write(junit, `<testsuite name="s" tests="2" failures="1"><testcase name="a"/><testcase name="b"><failure/></testcase></testsuite>`)

	// Run 1: artifact produced at this commit.
	first, _ := run()
	if s := crap(first); s.Missing != domain.SignalUnmeasured || s.Value != nil {
		t.Fatalf("no artifact: crap = %+v", s)
	}
	write(lcov+artifacts.SidecarSuffix, `{"revision":"`+first+`","producer":"pytest-cov"}`)
	if again, _ := run(artifacts.Spec{Path: lcov}, artifacts.Spec{Path: junit, Revision: first}); again != first {
		t.Fatalf("snapshot changed: %s vs %s", again, first)
	}
	signals, err := store.ListSignals(ctx, state.SignalFilter{MetricID: "scout.crap", ScopeAnchor: "pkg/core", Snapshot: first})
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

	provenance, err := store.ListIngestedArtifacts(ctx, first)
	if err != nil || len(provenance) != 2 {
		t.Fatalf("provenance = %+v, %v", provenance, err)
	}
	for _, p := range provenance {
		if p.RevisionMatch != artifacts.MatchMatch || len(p.Digest) != 64 || p.Revision != first {
			t.Fatalf("provenance row = %+v", p)
		}
	}
	if p := provenance[0]; p.Kind != artifacts.KindJUnit || !strings.Contains(string(p.Payload), `"tests":2`) || !strings.Contains(string(p.Payload), `"failures":1`) {
		t.Fatalf("junit provenance = %+v", p)
	}

	// Run 2: the next commit, artifact not attached: unmeasured.
	write(source, module.String()+"\n# changed\n")
	gitRun(t, root, "commit", "--quiet", "-am", "two")
	second, _ := run()
	if second == first {
		t.Fatal("expected a new snapshot")
	}
	if s := crap(second); s.Missing != domain.SignalUnmeasured || s.Value != nil || s.Numerator != nil {
		t.Fatalf("next run without artifact: crap = %+v", s)
	}

	// Run 3: the stale artifact (produced at the first commit) is attached to the second.
	run(artifacts.Spec{Path: lcov})
	signals, err = store.ListSignals(ctx, state.SignalFilter{MetricID: "scout.crap", ScopeAnchor: "pkg/core", Snapshot: second})
	if err != nil || len(signals) != 1 || signals[0].Missing != domain.SignalUnmeasured || signals[0].Value != nil {
		t.Fatalf("mismatched artifact: crap = %+v, %v", signals, err)
	}
	// The row already recorded for this snapshot and tool digest is kept
	// (append-only); the mismatch reason lives in the artifact provenance.
	mismatch, err := store.ListIngestedArtifacts(ctx, second)
	if err != nil || len(mismatch) != 1 || mismatch[0].RevisionMatch != artifacts.MatchMismatch || mismatch[0].Reason == "" {
		t.Fatalf("mismatch provenance = %+v, %v", mismatch, err)
	}
}

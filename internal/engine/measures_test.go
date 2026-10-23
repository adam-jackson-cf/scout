package engine

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"scout/internal/domain"
	"scout/internal/inventory"
	"scout/internal/measures"
	"scout/internal/state"
)

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-c", "user.name=Dev", "-c", "user.email=dev@example.com"}, args...)...)
	command.Dir = dir
	if out, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func TestMeasuresLanesRecordSignalsPerScopeAndReportHealth(t *testing.T) {
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
	if err := os.MkdirAll(filepath.Join(root, "pkg", "core"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "pkg", "core", "busy.py"), []byte(module.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, root, "init", "--quiet")
	gitRun(t, root, "add", ".")
	gitRun(t, root, "commit", "--quiet", "-m", "init")

	store, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	scan, err := inventory.Scan(ctx, root, inventory.ScanOptions{Scope: domain.Scope{Kind: domain.ScopeRepository, SelectedFamilies: []domain.Family{domain.FamilyQuality}}})
	if err != nil {
		t.Fatal(err)
	}
	options := Options{Scope: domain.Scope{Kind: domain.ScopeRepository}, Measures: &MeasuresOptions{History: true}}
	health := runMeasuresLanes(ctx, store, root, scan, nil, options)
	byLane := map[string]domain.LaneHealth{}
	for _, h := range health {
		byLane[h.Lane] = h
	}
	if byLane[LaneMeasures].Status != domain.LaneRan || byLane[LaneMeasures].Units != 1 {
		t.Fatalf("measures lane = %+v", byLane[LaneMeasures])
	}
	if h := byLane[LaneHistory]; h.Status != domain.LaneRan || h.Units != 1 {
		t.Fatalf("history lane = %+v", h)
	}
	signals, err := store.ListSignals(ctx, state.SignalFilter{ScopeAnchor: "pkg/core"})
	if err != nil {
		t.Fatal(err)
	}
	metrics := map[string]domain.Signal{}
	for _, signal := range signals {
		if signal.Window == "" {
			metrics[signal.MetricID] = signal
		}
	}
	cc := metrics[measures.MetricCC]
	if cc.Value == nil || *cc.Value != 1 || cc.PopulationDigest == "" || cc.Snapshot == "" {
		t.Fatalf("scout.cc = %+v", cc)
	}
	for _, metric := range []string{measures.MetricCC, measures.MetricComplexBurden, measures.MetricBurdenDensity, measures.MetricDupPct, measures.MetricCrap, "scout.cycle_mass"} {
		signal, ok := metrics[metric]
		if !ok || signal.PopulationDigest != cc.PopulationDigest {
			t.Fatalf("%s missing or carries another population digest: %+v", metric, signal)
		}
	}
	if metrics[measures.MetricCrap].Missing != domain.SignalUnmeasured {
		t.Fatal("scout.crap must be unmeasured")
	}
	history := 0
	for _, signal := range signals {
		if signal.Window != "" {
			history++
		}
	}
	if history == 0 {
		t.Fatal("history lane recorded no windowed signals for the scope")
	}

	// A PR run restricted to changed files measures only those files.
	options.Measures = &MeasuresOptions{Only: []string{"does/not/exist.py"}}
	if h := runMeasuresLanes(ctx, store, root, scan, nil, options); len(h) != 1 || h[0].Status != domain.LaneSkipped {
		t.Fatalf("changed-files run with nothing to measure = %+v", h)
	}
}

// Every signal of the measures, import-graph and history lanes carries bare
// 64-hex population, tool and config digests, measured or not.
func TestAllLaneSignalsCarryBareHexDigests(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	body := "def f(x):\n" + strings.Repeat("    x += 1\n", 12) + "    return x\n"
	if err := os.WriteFile(filepath.Join(root, "a.py"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, root, "init", "--quiet")
	gitRun(t, root, "add", ".")
	gitRun(t, root, "commit", "--quiet", "-m", "init")
	store, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	scan, err := inventory.Scan(ctx, root, inventory.ScanOptions{Scope: domain.Scope{Kind: domain.ScopeRepository, SelectedFamilies: []domain.Family{domain.FamilyQuality}}})
	if err != nil {
		t.Fatal(err)
	}
	// Missing scanners: measures signals are unmeasured but must still carry digests.
	options := Options{Scope: domain.Scope{Kind: domain.ScopeRepository}, Measures: &MeasuresOptions{History: true, Tools: measures.Tools{Lizard: "scout-no-lizard", Jscpd: "scout-no-jscpd"}}}
	runMeasuresLanes(ctx, store, root, scan, nil, options)
	signals, err := store.ListSignals(ctx, state.SignalFilter{})
	if err != nil {
		t.Fatal(err)
	}
	prefixes := map[string]bool{}
	for _, signal := range signals {
		prefixes[strings.SplitN(signal.MetricID, ".", 2)[1]] = true
		for name, digest := range map[string]string{"population": signal.PopulationDigest, "tool": signal.ToolDigest, "config": signal.ConfigDigest} {
			if !domain.IsDigest(digest) {
				t.Fatalf("%s %s digest %q is not 64 lowercase hex", signal.ID(), name, digest)
			}
		}
	}
	for _, want := range []string{"cc", "cycle_mass", "changes"} {
		if !prefixes[want] {
			t.Fatalf("no %s signal recorded (have %v)", want, prefixes)
		}
	}
}

//go:build boundary

package measures

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"scout/internal/domain"
	"scout/internal/sandbox"
)

func requireScanners(t *testing.T) {
	t.Helper()
	for _, binary := range []string{"lizard", "jscpd", "node"} {
		if _, err := exec.LookPath(binary); err != nil {
			t.Fatalf("%s is not on PATH (source .tools/activate)", binary)
		}
	}
	if err := sandbox.Available(); err != nil {
		t.Fatal("scanner sandbox is not available")
	}
}

func TestRunMeasuresRealScannersOnTemporaryRepository(t *testing.T) {
	requireScanners(t)
	root := t.TempDir()
	// Two identical production modules (a long clone), one module below the
	// 10-NLOC floor, and a test file that must be excluded by path role.
	writeRepoFile(t, root, "pkg/one/a.py", singleFunctionPythonModule("alpha", 11))
	writeRepoFile(t, root, "pkg/one/b.py", singleFunctionPythonModule("alpha", 11))
	writeRepoFile(t, root, "pkg/one/tiny.py", "def tiny():\n    return 1\n")
	writeRepoFile(t, root, "tests/test_a.py", singleFunctionPythonModule("alpha", 11))
	files := []domain.File{
		{Path: "pkg/one/a.py", ContentDigest: "da", Eligible: true},
		{Path: "pkg/one/b.py", ContentDigest: "db", Eligible: true},
		{Path: "pkg/one/tiny.py", ContentDigest: "dt", Eligible: true},
		{Path: "tests/test_a.py", ContentDigest: "dx", Eligible: true},
	}
	result := Run(t.Context(), Config{Root: root, Snapshot: domain.Snapshot{Revision: "rev", WorktreeDigest: "w"}, RunScope: "repository", Files: files})
	if result.LizardErr != "" || result.JscpdErr != "" {
		t.Fatalf("scanner errors: lizard=%q jscpd=%q", result.LizardErr, result.JscpdErr)
	}
	if len(result.Measured) != 2 {
		t.Fatalf("measured = %v, want a.py and b.py", result.Measured)
	}
	assertRealScannerSignals(t, result.Signals)
}

func writeRepoFile(t *testing.T, root, path, content string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertRealScannerSignals(t *testing.T, signals []domain.Signal) {
	t.Helper()
	cc := find(t, signals, MetricCC, "pkg/one")
	near(t, "n_high_cc", cc.Value, 2)
	near(t, "functions", cc.Denominator, 2)
	if cc.Detail["cc_max"] != 12 {
		t.Fatalf("cc_max = %v, want 12", cc.Detail["cc_max"])
	}
	burden := find(t, signals, MetricComplexBurden, "pkg/one")
	near(t, "burden", burden.Value, 1) // every function is CC 12 > 10
	dup := find(t, signals, MetricDupPct, "pkg/one")
	if dup.Value == nil || *dup.Value < 0.5 || *dup.Value > 1 {
		t.Fatalf("dup_pct of two identical modules = %v", dup.Value)
	}
	if got := *dup.Denominator; got != float64(2*(2*11+3)) {
		t.Fatalf("dup denominator = %v, want the two modules' 50 lines", got)
	}
	for _, signal := range signals {
		if err := signal.Validate(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRunRecordsCycleMassPerScope(t *testing.T) {
	requireScanners(t)
	root := t.TempDir()
	var files []domain.File
	for name, partner := range map[string]string{"a": "b", "b": "a"} {
		dir := filepath.Join(root, "pkg", "core")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		content := "from . import " + partner + "\n\n" + singleFunctionPythonModule("fn_"+name, 6)
		if err := os.WriteFile(filepath.Join(dir, name+".py"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		files = append(files, domain.File{Path: "pkg/core/" + name + ".py", ContentDigest: "d" + name, Eligible: true})
	}
	if err := os.WriteFile(filepath.Join(root, "pkg", "core", "__init__.py"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	result := Run(t.Context(), Config{Root: root, Snapshot: domain.Snapshot{Revision: "rev", WorktreeDigest: "w"}, RunScope: "repository", Files: files})
	cycle := find(t, result.Signals, MetricCycleMass, "pkg/core")
	if err := cycle.Validate(); err != nil {
		t.Fatal(err)
	}
	near(t, "cycle edges", cycle.Numerator, 2)
	near(t, "internal edges", cycle.Denominator, 2)
	near(t, "cycle_mass", cycle.Value, 1)
	if want := find(t, result.Signals, MetricCC, "pkg/core").PopulationDigest; cycle.PopulationDigest != want {
		t.Fatal("cycle_mass population digest differs from the scope's other measures")
	}
}

package engine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"scout/internal/analyzers"
	"scout/internal/domain"
	"scout/internal/measures"
)

func requireDeltaTools(t *testing.T) {
	t.Helper()
	for _, binary := range []string{"lizard", "jscpd", "node", "git"} {
		if _, err := exec.LookPath(binary); err != nil {
			t.Skipf("%s is not on PATH (source .tools/activate)", binary)
		}
	}
	if err := analyzers.SandboxAvailable(); err != nil {
		t.Skip("scanner sandbox is not available")
	}
}

func deltaGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root, "-c", "user.name=t", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false"}, args...)...)
	out, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func deltaWrite(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// branchy returns a Python function with exactly `branches` decision points
// (CC = branches+1) padded past the 10-NLOC eligibility floor.
func branchy(name string, branches int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "def %s(x):\n    total = 0\n", name)
	for i := range branches {
		fmt.Fprintf(&b, "    if x == \"%s-%d\":\n        total_%s_%d = x.upper() + \"%s\"\n        total += len(total_%s_%d) * %d\n", name, i, name, i, name, name, i, i+3)
	}
	b.WriteString("    return total\n")
	return b.String()
}

func deltaRepo(t *testing.T, files map[string]string) (root, base string) {
	t.Helper()
	root = t.TempDir()
	deltaGit(t, root, "init", "-q", "-b", "main")
	for name, content := range files {
		deltaWrite(t, root, name, content)
	}
	deltaGit(t, root, "add", "-A")
	deltaGit(t, root, "commit", "-qm", "base")
	return root, deltaGit(t, root, "rev-parse", "HEAD")
}

func commitChange(t *testing.T, root string, files map[string]string) []string {
	t.Helper()
	var changed []string
	for name, content := range files {
		deltaWrite(t, root, name, content)
		changed = append(changed, name)
	}
	deltaGit(t, root, "add", "-A")
	deltaGit(t, root, "commit", "-qm", "pr")
	return changed
}

func TestDeltaCCCrossingIsOneDescriptiveReportOnlyIssueOnChangedFilesOnly(t *testing.T) {
	requireDeltaTools(t)
	// stable.py also sits at CC 14 untouched; wide.py goes 9 → 14.
	root, base := deltaRepo(t, map[string]string{
		"wide.py":   branchy("wide", 8) + "\n\n" + branchy("steady", 3),
		"stable.py": branchy("big", 13),
	})
	changed := commitChange(t, root, map[string]string{"wide.py": branchy("wide", 13) + "\n\n" + branchy("steady", 3)})
	report := ComputeDeltas(context.Background(), root, base, changed, measures.Tools{})
	if len(report.Unmeasured) != 0 {
		t.Fatalf("unmeasured: %v", report.Unmeasured)
	}
	if len(report.Issues) != 1 {
		t.Fatalf("want exactly one delta, got %+v", report.Issues)
	}
	issue := report.Issues[0]
	if issue.Claim != domain.ClaimDescriptive || issue.Mode != "report-only" || issue.Rule != "delta.cc-threshold@1" {
		t.Fatalf("identity: %+v", issue)
	}
	if issue.Path != "wide.py" || issue.Function == "" || issue.Base != 9 || issue.Head != 14 || issue.Threshold != 10 || issue.BaseRev != base {
		t.Fatalf("evidence: %+v", issue)
	}
	obs := issue.observation("run-1", time.Unix(0, 0))
	if err := obs.Validate(); err != nil {
		t.Fatal(err)
	}
	anchor, _ := domain.AnchorOf(obs)
	if obs.Claim != domain.ClaimDescriptive || obs.RuleID != "delta.cc-threshold@1" || CheckIDOf(obs.Source, obs.RuleID) != "delta.cc-threshold@1" ||
		anchor.Path != "wide.py" || anchor.Symbol != issue.Function || !strings.Contains(obs.Evidence[0].RedactedQuote, "base=9 head=14 threshold=10 baseRevision="+base) {
		t.Fatalf("observation: %+v anchor %+v", obs, anchor)
	}
}

func TestDeltaIgnoresNonCrossingsAndPopulationChange(t *testing.T) {
	requireDeltaTools(t)
	root, base := deltaRepo(t, map[string]string{
		"a.py": branchy("a", 4) + "\n\n" + branchy("gone", 12),
		"b.py": branchy("b", 12),
	})
	changed := commitChange(t, root, map[string]string{
		"a.py":   branchy("a", 7) + "\n\n" + branchy("fresh", 20), // 5 → 8 stays under; a new function is population change
		"b.py":   branchy("b", 20),                                // 13 → 21: already above the line, not a crossing
		"new.py": branchy("n", 20),                                // a new file is population change
	})
	report := ComputeDeltas(context.Background(), root, base, changed, measures.Tools{})
	if len(report.Unmeasured) != 0 || len(report.Issues) != 0 {
		t.Fatalf("want no delta, got %+v", report)
	}
	// Nothing changed: nothing to report.
	if r := ComputeDeltas(context.Background(), root, base, nil, measures.Tools{}); len(r.Issues) != 0 {
		t.Fatalf("no changed files: %+v", r)
	}
}

func TestDeltaCloneAddedIsReportedAndUntouchedCloneIsNot(t *testing.T) {
	requireDeltaTools(t)
	block := branchy("shared", 12)
	root, base := deltaRepo(t, map[string]string{
		"orig.py":  block,
		"twin1.py": branchy("twin", 14), "twin2.py": branchy("twin", 14), // pre-existing clone, left alone
		"other.py": branchy("other", 3),
	})
	changed := commitChange(t, root, map[string]string{
		"copy.py":  strings.ReplaceAll(block, "def shared", "def shared_copy"),
		"other.py": branchy("other", 3) + "\n\n" + branchy("more", 2),
	})
	report := ComputeDeltas(context.Background(), root, base, changed, measures.Tools{})
	if len(report.Unmeasured) != 0 {
		t.Fatalf("unmeasured: %v", report.Unmeasured)
	}
	var dups []DeltaIssue
	for _, issue := range report.Issues {
		if issue.Rule == "delta.dup-added@1" {
			dups = append(dups, issue)
		}
	}
	if len(dups) != 1 || dups[0].Path != "copy.py" || dups[0].Partner != "orig.py" || dups[0].Claim != domain.ClaimDescriptive || dups[0].Mode != "report-only" {
		t.Fatalf("dup deltas: %+v (all %+v)", dups, report.Issues)
	}
}

func TestDeltaWithoutScannersIsUnmeasuredNeverEmpty(t *testing.T) {
	root, base := deltaRepo(t, map[string]string{"a.py": branchy("a", 3)})
	changed := commitChange(t, root, map[string]string{"a.py": branchy("a", 12)})
	report := ComputeDeltas(context.Background(), root, base, changed, measures.Tools{Lizard: "scout-no-lizard", Jscpd: "scout-no-jscpd"})
	if len(report.Issues) != 0 || len(report.Unmeasured) == 0 {
		t.Fatalf("missing scanners must be stated, not silent: %+v", report)
	}
}

func TestCapabilityMatrixReportsPerRepoAvailability(t *testing.T) {
	root := t.TempDir()
	deltaWrite(t, root, "main.go", "package main\n")
	deltaWrite(t, root, "app.py", "x = 1\n")
	t.Setenv("PATH", t.TempDir()) // no model CLI on PATH
	matrix, err := CapabilityMatrix(context.Background(), Config{}, root)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(matrix.InRepo, ",") != "Go,Python" || len(matrix.Rows) != 4 {
		t.Fatalf("matrix: %+v", matrix)
	}
	rows := map[string]MatrixRow{}
	for _, row := range matrix.Rows {
		rows[row.Capability] = row
	}
	review := rows["Semantic review (contextual)"]
	for _, language := range MatrixLanguages {
		if cell := review.Cells[language]; cell.State != CellSkipped || renderCell(cell) != "skipped (no model CLI)" {
			t.Fatalf("%s review = %+v", language, cell)
		}
	}
	graph := rows["Import graph / cycles; guardrails; rubric"]
	if graph.Cells["Go"].State != CellAvailable || graph.Cells["Rust"].State != CellPlanned || renderCell(graph.Cells["C#"]) != "–" {
		t.Fatalf("import graph row: %+v", graph.Cells)
	}
	crap := rows["CRAP (needs coverage artifact)"]
	if crap.Cells["Go"].State != CellWhen {
		t.Fatalf("crap without artifact: %+v", crap.Cells["Go"])
	}
	deltaWrite(t, root, "coverage/lcov.info", "TN:\n")
	again, _ := CapabilityMatrix(context.Background(), Config{}, root)
	for _, row := range again.Rows {
		if strings.HasPrefix(row.Capability, "CRAP") && row.Cells["Go"].State != CellAvailable {
			t.Fatalf("crap with artifact: %+v", row.Cells["Go"])
		}
	}
	text := matrix.Render()
	for _, want := range []string{"| Capability | Go * | Python * | TS/JS | Rust | Java | C# |", "CRAP (needs coverage artifact)", "present in this repository: Go, Python"} {
		if !strings.Contains(text, want) {
			t.Fatalf("render lacks %q:\n%s", want, text)
		}
	}
}

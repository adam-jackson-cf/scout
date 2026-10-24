package issues

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"scout/internal/analyzers"
	"scout/internal/domain"
	"scout/internal/state/statetest"
)

// Real analyzer → correlator path: Gitleaks reports the secret redacted; the
// runner derives a digest of the raw value in memory, so moving the secret to
// another file, line and variable keeps one issue and the raw value is never
// persisted.
func TestGitleaksSecretKeepsIssueAcrossMoveEndToEnd(t *testing.T) {
	registry, err := analyzers.BuiltinRegistry()
	if err != nil {
		t.Fatal(err)
	}
	secret := "AKIA" + "IOSFODNN7EXAMPLQ"
	h := newHarness(t)
	scan := func(path, content string) []domain.Issue {
		t.Helper()
		root := t.TempDir()
		full := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		h.runs++
		suffix := fmt.Sprint(h.runs)
		runner, err := registry.Runner("security.gitleaks.default-credentials", "run-"+suffix)
		if err != nil {
			t.Fatal(err)
		}
		result, observations, _ := runner.Run(h.ctx, root, []domain.File{{Path: path, Language: "python", Eligible: true}})
		if result.Status == domain.AdapterMissingTool {
			t.Skip("gitleaks or sandbox is not installed")
		}
		if len(observations) != 1 {
			t.Fatalf("adapter status %s produced %d observations", result.Status, len(observations))
		}
		encoded, _ := json.Marshal(observations)
		if strings.Contains(string(encoded), secret) {
			t.Fatal("raw secret reached an observation")
		}
		snapshot := domain.Snapshot{RepositoryID: "repository", WorktreeDigest: "worktree-" + suffix, IgnoreRulesDigest: "ignore", ExclusionsDigest: "exclude"}
		scope := domain.Scope{Kind: domain.ScopeRepository, SelectedFamilies: []domain.Family{domain.FamilySecurity}}
		plan := domain.Plan{PlanID: "plan-" + suffix, RunID: "run-" + suffix, Snapshot: snapshot, Scope: scope, CreatedAt: time.Now().UTC()}
		row := domain.Coverage{Key: domain.CoverageKey{FilePath: path, Family: domain.FamilySecurity, Category: observations[0].OriginalCategory}, Status: domain.CoverageFlagged, InspectedRanges: []domain.SourceRange{{Path: path, StartLine: 1, EndLine: 1}}, ObservationIDs: []string{observations[0].ObservationID}}
		run := domain.Run{RunID: plan.RunID, PlanID: plan.PlanID, Snapshot: snapshot, Scope: scope, State: domain.RunPartial, StartedAt: time.Now().UTC(), Coverage: []domain.Coverage{row}}
		statetest.SeedRun(t, h.ctx, h.store, plan, run, observations, nil)
		issues, err := Correlate(h.ctx, h.store, run, observations)
		if err != nil {
			t.Fatal(err)
		}
		return issues
	}
	first := scan("a.py", "x = 1\nAWS_KEY = \""+secret+"\"\n")
	moved := scan("config/b.py", "# moved\n\n\n\nTOKEN_ID = \""+secret+"\"\n")
	if first[0].IssueID != moved[0].IssueID || len(h.all()) != 1 {
		t.Fatalf("moved secret created a new issue: %q vs %q", first[0].IssueID, moved[0].IssueID)
	}
}

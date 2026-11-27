//go:build boundary

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func bareOrigin(t *testing.T, work string) string {
	t.Helper()
	origin := filepath.Join(work, "origin.git")
	if out, err := exec.CommandContext(boundedTestContext(t), "git", "init", "-q", "--bare", "-b", "main", origin).CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	return origin
}

func cloneInto(t *testing.T, work, origin, name string) string {
	t.Helper()
	dir := filepath.Join(work, name)
	if out, err := exec.CommandContext(boundedTestContext(t), "git", "clone", "-q", origin, dir).CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	return dir
}

func publishMainBaseline(t *testing.T, main, origin string) {
	t.Helper()
	commitFiles(t, main, "base", map[string]string{"pkg/legacy.py": complexFn("legacy", 12), "README.md": "demo\n"})
	gitIn(t, main, "push", "-q", "origin", "main")
	ctx := context.Background()
	if code := run(ctx, []string{"assess", "--root", main, "--json"}); code == exitUnavailable {
		t.Fatalf("assess exit %d", code)
	}
	if code := run(ctx, []string{"state", "publish", "--root", main, "--remote", origin, "--trusted"}); code != 0 {
		t.Fatalf("publish exit %d", code)
	}
}

func readManifest(t *testing.T, manifestPath string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	return manifest
}

func pushManifest(t *testing.T, state, message string, manifest map[string]any) {
	t.Helper()
	raw, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	commitFiles(t, state, message, map[string]string{"state/v1/manifest.json": string(raw)})
	gitIn(t, state, "push", "-q", "origin", "scout/state")
}

// A trusted main run publishes scout/state; a fresh clone's PR diff must use
// that baseline when nothing changed and fall back, with a reason, when the
// baseline's tool digest no longer matches (a rule/tool version bump).
func TestDiffUsesPublishedBaselineUnlessDigestsDiffer(t *testing.T) {
	requireTools(t, "ruff", "gitleaks", "semgrep")
	t.Setenv("SCOUT_CLAWPATCH_BINARY", pinnedClawpatch(t))
	work := t.TempDir()
	origin := bareOrigin(t, work)
	publishMainBaseline(t, cloneInto(t, work, origin, "main"), origin)

	pr := cloneInto(t, work, origin, "pr")
	gitIn(t, pr, "checkout", "-q", "-b", "feature")
	commitFiles(t, pr, "pr", map[string]string{"pkg/new.py": "import os  # noqa\n\n\ndef helper():\n    return os.getcwd()\n"})

	_, report := runDiffJSON(t, pr, "--base", "origin/main", "--baseline-remote", origin)
	if report["comparison"] != "baseline" || report["comparison_reason"] != nil {
		t.Fatalf("unchanged baseline: comparison %v reason %v", report["comparison"], report["comparison_reason"])
	}

	// Simulate a tool/rule version bump: the published manifest's tool digest changes.
	state := cloneInto(t, work, origin, "state")
	gitIn(t, state, "checkout", "-q", "scout/state")
	manifest := readManifest(t, filepath.Join(state, "state", "v1", "manifest.json"))
	manifest["toolDigest"] = "bumped-rule-pack"
	pushManifest(t, state, "bump", manifest)

	_, report = runDiffJSON(t, pr, "--base", "origin/main", "--baseline-remote", origin)
	reason, _ := report["comparison_reason"].(string)
	if report["comparison"] != "recomputed" || !strings.Contains(reason, "tool digest") {
		t.Fatalf("bumped baseline: comparison %v reason %q", report["comparison"], reason)
	}

	// A manifest without `uninspected` (a v0.1.0/v0.1.1 publish) is not a
	// baseline: the differential run ends unavailable (exit 3), naming the
	// cause and remedy, instead of recomputing around it.
	delete(manifest, "uninspected")
	pushManifest(t, state, "stale", manifest)
	code, report := runDiffJSON(t, pr, "--base", "origin/main", "--baseline-remote", origin)
	encoded, _ := json.Marshal(report)
	if code != exitUnavailable || !strings.Contains(string(encoded), "scout/state manifest carries no lane health; re-baseline") {
		t.Fatalf("manifest without uninspected: exit %d report %s", code, encoded)
	}
}

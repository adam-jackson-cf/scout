package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"scout/internal/domain"
	"scout/internal/engine"
)

// fakeModelCLIs puts executable fakes first on PATH. Each appends its name to a
// shared log when invoked and then fails, so no real model is ever called.
func fakeModelCLIs(t *testing.T, names ...string) (log string) {
	t.Helper()
	dir := t.TempDir()
	log = filepath.Join(dir, "invocations.log")
	for _, name := range names {
		script := "#!/bin/sh\necho " + name + " >> '" + log + "'\ncat >/dev/null 2>&1\nexit 1\n"
		if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("SCOUT_MODEL_CLI", "")
	t.Setenv("CLAWPATCH_PROVIDER", "")
	return log
}

// isolatedPath makes PATH contain only git, node and sh, so a real claude or
// codex on the developer machine cannot be detected.
func isolatedPath(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	for _, tool := range []string{"git", "node", "sh", "cat", "env"} {
		if path, err := exec.LookPath(tool); err == nil {
			if err := os.Symlink(path, filepath.Join(dir, tool)); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Setenv("PATH", dir)
	t.Setenv("SCOUT_MODEL_CLI", "")
	t.Setenv("CLAWPATCH_PROVIDER", "")
}

func invoked(t *testing.T, log string) string {
	t.Helper()
	data, err := os.ReadFile(log)
	if os.IsNotExist(err) {
		return ""
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func consentRepo(t *testing.T, initFlags ...string) string {
	t.Helper()
	root := newInitRepo(t)
	writeRepoFile(t, root, "pkg/a.py", complexFn("a", 3))
	writeRepoFile(t, root, "pyproject.toml", "[project]\nname = \"demo\"\nrequires-python = \">=3.11\"\n\n[project.scripts]\ndemo = \"pkg.a:a\"\n")
	if len(initFlags) > 0 {
		if code := runInit(initArgs(root, initFlags...), strings.NewReader(""), os.Stderr); code != exitComplete {
			t.Fatalf("init exit %d", code)
		}
	}
	gitIn(t, root, "add", "-A")
	gitIn(t, root, "commit", "-qm", "r1")
	return root
}

func assessWithConsent(t *testing.T, root string, authorized bool) engine.Result {
	t.Helper()
	cfg, err := engineConfig(&commonFlags{root: root, binary: pinnedClawpatch(t)}, authorized)
	if err != nil {
		t.Fatal(err)
	}
	scope, err := scopeOptions("repository", nil, "", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := engine.Assess(context.Background(), cfg, engine.Options{Scope: scope, Trigger: domain.TriggerNightly})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func modelLanes(result engine.Result) map[string]domain.LaneHealth {
	out := map[string]domain.LaneHealth{}
	for _, h := range result.Lanes {
		if h.Lane == engine.LaneContextual || h.Lane == engine.LaneClawpatchReview {
			out[h.Lane] = h
		}
	}
	return out
}

func TestModelCLILanesNotOptedInSkipWithoutInvokingACLI(t *testing.T) {
	log := fakeModelCLIs(t, "claude", "codex")
	root := consentRepo(t)
	result := assessWithConsent(t, root, true)
	if got := invoked(t, log); got != "" {
		t.Fatalf("a model CLI ran without opt-in: %q", got)
	}
	for id, h := range modelLanes(result) {
		if h.Status == domain.LaneFailed || h.Required {
			t.Fatalf("%s must be optional and not failed: %+v", id, h)
		}
	}
	if h := modelLanes(result)[engine.LaneContextual]; h.Status != domain.LaneSkipped || h.Reason != "skipped (not opted in)" {
		t.Fatalf("contextual = %+v", h)
	}
}

func TestModelCLILaneOptedInUsesPolicyProviderNotFirstOnPath(t *testing.T) {
	log := fakeModelCLIs(t, "codex", "claude") // auto-detection would prefer codex
	root := consentRepo(t, "--model-cli", "claude", "--lane", "contextual", "--lane", "clawpatch-review")
	t.Setenv("SCOUT_MODEL_CLI", "codex") // a local override must lose to policy
	assessWithConsent(t, root, true)
	got := invoked(t, log)
	if !strings.Contains(got, "claude") || strings.Contains(got, "codex") {
		t.Fatalf("invocations = %q, want claude only", got)
	}
}

func TestModelCLILaneOptedInWithoutCLIIsSkippedNoModelCLI(t *testing.T) {
	root := consentRepo(t, "--model-cli", "claude", "--lane", "contextual")
	isolatedPath(t)
	result := assessWithConsent(t, root, true)
	if h := modelLanes(result)[engine.LaneContextual]; h.Status != domain.LaneSkipped || h.Reason != "skipped (no model CLI)" || h.Required {
		t.Fatalf("contextual = %+v", h)
	}
}

func TestSCOUTModelCLIAppliesOnlyWhenPolicyProviderUnset(t *testing.T) {
	root := consentRepo(t)
	t.Setenv("SCOUT_MODEL_CLI", "codex")
	cfg, err := engineConfig(&commonFlags{root: root}, true)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Provider.CLI != "codex" || cfg.ClawpatchProvider != "codex" {
		t.Fatalf("env override ignored with policy provider unset: %+v", cfg)
	}
}

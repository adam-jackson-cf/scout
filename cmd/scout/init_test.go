package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"scout/internal/policy"
)

func initArgs(root string, extra ...string) []string {
	return append([]string{"--root", root, "--yes"}, extra...)
}

func loadCommitted(t *testing.T, root string) policy.Policy {
	t.Helper()
	gitIn(t, root, "add", "-A")
	gitIn(t, root, "commit", "-q", "-m", "policy")
	p, err := policy.Load(root, "HEAD")
	if err != nil {
		t.Fatalf("policy.Load: %v", err)
	}
	return p
}

func newInitRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	gitIn(t, root, "init", "-q")
	writeRepoFile(t, root, "README.md", "x\n")
	return root
}

func TestInitYesWritesConsentThatPolicyLoadRoundTrips(t *testing.T) {
	root := newInitRepo(t)
	code := runInit(initArgs(root, "--decisions", "--decisions-model", "jev-2026-10-01", "--acknowledge-egress", "--model-cli", "claude", "--lane", "contextual"), strings.NewReader(""), os.Stderr)
	if code != exitComplete {
		t.Fatalf("exit %d", code)
	}
	p := loadCommitted(t, root)
	d := p.Decisions()
	if p.Profile() != policy.ProfileReport || !d.Enabled || d.Model != "jev-2026-10-01" || !d.EgressAcknowledged || d.MaxPerRun != policy.DefaultMaxPerRun {
		t.Fatalf("profile/decisions not recorded: %s %+v", p.Profile(), d)
	}
	if m := p.ModelCLI(); m.Provider != "claude" || !p.LaneOptedIn("contextual") || p.LaneOptedIn("clawpatch-review") || len(m.Lanes) != 1 {
		t.Fatalf("model-cli opt-in not recorded: %+v", m)
	}
}

func TestInitWithoutFlagsRecordsReportAndNoConsent(t *testing.T) {
	root := newInitRepo(t)
	if code := runInit(initArgs(root), strings.NewReader(""), os.Stderr); code != exitComplete {
		t.Fatalf("exit %d", code)
	}
	p := loadCommitted(t, root)
	if p.Profile() != policy.ProfileReport || p.Decisions().Enabled || p.ModelCLI().Provider != "" {
		t.Fatalf("unexpected consent: %+v %+v", p.Decisions(), p.ModelCLI())
	}
}

func TestInitIsIdempotentAndNeverTouchesActivatedBlocking(t *testing.T) {
	root := newInitRepo(t)
	existing := `# maintainers own this file
profile: guardian
blocking:
  shadow: {min_days: 14, min_pr_runs: 5}
  checks:
    - id: guardrails.SUPPRESS.added.unscoped
      version: "1"
      activated_at: 2026-01-01
      evidence: {would_block: 2, warranted: 2, window_days: 15, pr_runs: 6, dossier_digest: "sha256:abc"}
maintainers: [alice]
`
	writeRepoFile(t, root, policy.Path, existing)
	args := initArgs(root, "--decisions", "--decisions-model", "jev-2026-10-01", "--acknowledge-egress")
	if code := runInit(args, strings.NewReader(""), os.Stderr); code != exitComplete {
		t.Fatalf("exit %d", code)
	}
	path := filepath.Join(root, filepath.FromSlash(policy.Path))
	first, _ := os.ReadFile(path)
	if code := runInit(args, strings.NewReader(""), os.Stderr); code != exitComplete {
		t.Fatalf("second exit %d", code)
	}
	second, _ := os.ReadFile(path)
	if string(first) != string(second) {
		t.Fatalf("re-running init changed the file:\n%s\n---\n%s", first, second)
	}
	p := loadCommitted(t, root)
	if p.Profile() != policy.ProfileGuardian {
		t.Fatalf("existing guardian profile was lowered to %s", p.Profile())
	}
	acts := p.Activations()
	if len(acts) != 1 || acts[0].ID != "guardrails.SUPPRESS.added.unscoped" || acts[0].Version != "1" || acts[0].Evidence == nil || acts[0].Evidence.Warranted != 2 {
		t.Fatalf("activated blocking entry altered: %+v", acts)
	}
	if !strings.Contains(string(second), "maintainers own this file") || !p.Decisions().Enabled {
		t.Fatalf("comment lost or consent missing:\n%s", second)
	}
}

func TestInitRejectsInconsistentConsent(t *testing.T) {
	for name, extra := range map[string][]string{
		"alias model":             {"--decisions", "--decisions-model", "jev-latest", "--acknowledge-egress"},
		"preview alias":           {"--decisions", "--decisions-model", "typesafe/jev-preview", "--acknowledge-egress"},
		"undated release line":    {"--decisions", "--decisions-model", "typesafe/jev-1.14", "--acknowledge-egress"},
		"enabled without egress":  {"--decisions", "--decisions-model", "jev-2026-10-01"},
		"default model no egress": {"--decisions"},
		"unknown model cli":       {"--model-cli", "gemini"},
		"unknown lane":            {"--model-cli", "codex", "--lane", "nope"},
	} {
		t.Run(name, func(t *testing.T) {
			root := newInitRepo(t)
			if code := runInit(initArgs(root, extra...), strings.NewReader(""), os.Stderr); code != exitUnavailable {
				t.Fatalf("exit %d, want %d", code, exitUnavailable)
			}
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(policy.Path))); err == nil {
				t.Fatal("a rejected init wrote a policy file")
			}
		})
	}
}

func TestInitInteractiveRecordsAnswers(t *testing.T) {
	root := newInitRepo(t)
	answers := "y\njev-2026-10-01\ny\ny\ncodex\ny\nn\n"
	if code := runInit([]string{"--root", root}, strings.NewReader(answers), &strings.Builder{}); code != exitComplete {
		t.Fatalf("exit %d", code)
	}
	p := loadCommitted(t, root)
	if !p.Decisions().Enabled || p.ModelCLI().Provider != "codex" || !p.LaneOptedIn("contextual") || p.LaneOptedIn("clawpatch-review") {
		t.Fatalf("answers not recorded: %+v %+v", p.Decisions(), p.ModelCLI())
	}
}

// The pin defaults to Jev 1.14 through OpenRouter (user decision 2026-10-23):
// the dated slug OpenRouter lists as the endpoint of `typesafe/jev-1.14`.
func TestInitDecisionsPinsJev114ByDefaultAndKeepsAnExistingPin(t *testing.T) {
	if policy.DefaultModel != "typesafe/jev-1.14-20261014" {
		t.Fatalf("default pin %q", policy.DefaultModel)
	}
	root := newInitRepo(t)
	if code := runInit(initArgs(root, "--decisions", "--acknowledge-egress"), strings.NewReader(""), os.Stderr); code != exitComplete {
		t.Fatalf("exit %d", code)
	}
	d := loadCommitted(t, root).Decisions()
	if !d.Enabled || d.Model != "typesafe/jev-1.14-20261014" || d.MaxCostUSD != policy.DefaultMaxCostUSD {
		t.Fatalf("decisions %+v", d)
	}
	// Re-running without a model must not replace the pin, whatever it is.
	writeRepoFile(t, root, policy.Path, "profile: report\ndecisions:\n  enabled: true\n  provider: jev\n  model: jev-2026-10-01\n  egress_acknowledged: true\n")
	if code := runInit(initArgs(root, "--decisions", "--acknowledge-egress"), strings.NewReader(""), os.Stderr); code != exitComplete {
		t.Fatalf("exit %d", code)
	}
	if got := loadCommitted(t, root).Decisions().Model; got != "jev-2026-10-01" {
		t.Fatalf("existing pin replaced by %q", got)
	}
	// Interactive: an empty answer takes the default for a new pin.
	fresh := newInitRepo(t)
	if code := runInit([]string{"--root", fresh}, strings.NewReader("y\n\ny\nn\n"), &strings.Builder{}); code != exitComplete {
		t.Fatalf("exit %d", code)
	}
	if got := loadCommitted(t, fresh).Decisions().Model; got != policy.DefaultModel {
		t.Fatalf("interactive default %q", got)
	}
}

func TestInitRecordsABilledCostCap(t *testing.T) {
	root := newInitRepo(t)
	if code := runInit(initArgs(root, "--decisions", "--acknowledge-egress", "--max-cost-usd", "0.25"), strings.NewReader(""), os.Stderr); code != exitComplete {
		t.Fatalf("exit %d", code)
	}
	if got := loadCommitted(t, root).Decisions().MaxCostUSD; got != 0.25 {
		t.Fatalf("cost cap %v", got)
	}
}

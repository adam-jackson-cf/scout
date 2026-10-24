package policy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const guardianPolicy = `
profile: guardian
blocking:
  shadow: {min_days: 14, min_pr_runs: 5}
  checks:
    - id: guardrails.SUPPRESS.added.unscoped
      version: "1"
      activated_at: 2026-01-01
      evidence: {would_block: 2, warranted: 2, window_days: 15, pr_runs: 6, dossier_digest: "sha256:abc"}
    - id: security.secrets.github-pat
      version: "2"
      activated_at: 2026-01-02T00:00:00Z
      provisional: true
maintainers: [alice, "@Bob"]
history: {bots: ["dependabot[bot]"]}
`

func mustParse(t *testing.T, s string) Policy {
	t.Helper()
	p, err := Parse([]byte(s))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestAbsentIsReportNothingActive(t *testing.T) {
	p := mustParse(t, "")
	if p.Profile() != "report" {
		t.Fatalf("profile %s", p.Profile())
	}
	if got := p.CheckStatus("guardrails.SUPPRESS.added.unscoped", "1"); got != Shadow {
		t.Fatalf("got %s", got)
	}
	if len(p.Activations()) != 0 {
		t.Fatal("activations present")
	}
}

func TestParseErrors(t *testing.T) {
	entry := func(id string) string {
		return "profile: guardian\nblocking:\n  - id: " + id + "\n    version: \"1\"\n    activated_at: 2026-01-01\n    evidence: {would_block: 1}\n"
	}
	cases := map[string]string{
		"wildcard":                         entry("guardrails.CONFIG.*"),
		"secrets wild":                     entry("security.secrets.*"),
		"generic key":                      entry("security.secrets.generic-api-key"),
		"unknown id":                       entry("performance.ruff-perf"),
		"unknown key":                      "profile: guardian\nbogus: 1\n",
		"bad profile":                      "profile: strict\n",
		"missing version":                  "blocking:\n  - id: guardrails.SUPPRESS.added.unscoped\n    activated_at: 2026-01-01\n    provisional: true\n",
		"bad date":                         "blocking:\n  - id: guardrails.SUPPRESS.added.unscoped\n    version: \"1\"\n    activated_at: yesterday\n    provisional: true\n",
		"no evidence":                      "blocking:\n  - id: guardrails.SUPPRESS.added.unscoped\n    version: \"1\"\n    activated_at: 2026-01-01\n",
		"duplicate":                        entry("guardrails.SUPPRESS.added.unscoped") + strings.SplitN(entry("guardrails.SUPPRESS.added.unscoped"), "blocking:\n", 2)[1],
		"unknown entry key":                "blocking:\n  - id: guardrails.SUPPRESS.added.unscoped\n    version: \"1\"\n    activated_at: 2026-01-01\n    nope: 1\n    provisional: true\n",
		"removed mapping.refresh.debounce": "mapping:\n  refresh:\n    debounce: 1h\n",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(src)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
	if _, err := Parse([]byte("blocking:\n  - id: security.secrets.generic-api-key\n")); err == nil || !strings.Contains(err.Error(), "generic-api-key") {
		t.Fatalf("error should name the id: %v", err)
	}
}

func TestMappingPolicy(t *testing.T) {
	if got := mustParse(t, "").Mapping(); got != DefaultMapping() || got.AreasMatchJaccard != 0.5 {
		t.Fatalf("default mapping = %+v", got)
	}
	if got := mustParse(t, "mapping:\n  areas:\n    match_jaccard: 0.4\n").Mapping(); got.AreasMatchJaccard != 0.4 {
		t.Fatalf("match_jaccard = %+v", got)
	}
	if _, err := Parse([]byte("mapping:\n  refresh:\n    debounce: 1h\n")); err == nil || !strings.Contains(err.Error(), "invalid policy") {
		t.Fatalf("mapping.refresh was removed in revision 2.6 and must fail to parse: %v", err)
	}
}

func TestEligibleIDs(t *testing.T) {
	for _, id := range []string{
		"security.ci-mutable-action-ref", "security.secrets.aws-access-token", "security.secrets.private-key",
		"guardrails.CONFIG.ignores-added.golangci", "guardrails.CONFIG.strictness-lowered.mypy",
		"guardrails.SUPPRESS.added.unscoped", "guardrails.TESTQ.focused-test-added",
	} {
		if err := ValidateEligibleID(id); err != nil {
			t.Errorf("%s: %v", id, err)
		}
	}
	for _, id := range []string{"", "guardrails.CONFIG.ignores-added.", "guardrails.CONFIG.ignores-added", "guardrails.GATE.coverage-pragma-added", "security.secrets.slack-*", "delta.cc-threshold", "delta.cc-threshold@1", "delta.dup-added@1"} {
		if ValidateEligibleID(id) == nil {
			t.Errorf("%q should be ineligible", id)
		}
	}
	if _, err := Parse([]byte("blocking:\n  - id: delta.cc-threshold@1\n")); err == nil || !strings.Contains(err.Error(), "delta") {
		t.Fatalf("a delta rule must be rejected as a blocking activation: %v", err)
	}
}

func TestCheckStatus(t *testing.T) {
	p := mustParse(t, guardianPolicy)
	cases := []struct {
		id, ver string
		want    Status
	}{
		{"guardrails.SUPPRESS.added.unscoped", "1", Active},
		{"guardrails.SUPPRESS.added.unscoped", "2", Shadow}, // version bump returns to shadow
		{"security.secrets.github-pat", "2", ActiveProvisional},
		{"security.ci-mutable-action-ref", "1", Shadow}, // eligible, not activated
		{"security.secrets.generic-api-key", "1", NotEligible},
		{"guardrails.CONFIG.*", "1", NotEligible},
		{"inertia.delta", "1", NotEligible},
	}
	for _, c := range cases {
		if got := p.CheckStatus(c.id, c.ver); got != c.want {
			t.Errorf("%s@%s = %s, want %s", c.id, c.ver, got, c.want)
		}
	}
	if p.HistoryBots[0] != "dependabot[bot]" {
		t.Fatal("history.bots lost")
	}
	// report profile never activates even with entries.
	r := mustParse(t, strings.Replace(guardianPolicy, "profile: guardian", "profile: report", 1))
	if got := r.CheckStatus("guardrails.SUPPRESS.added.unscoped", "1"); got != Shadow {
		t.Fatalf("report profile gave %s", got)
	}
}

func TestListFormAndShadowOverride(t *testing.T) {
	p := mustParse(t, "profile: guardian\nblocking:\n  - id: guardrails.TESTQ.focused-test-added\n    version: \"1\"\n    activated_at: 2026-02-01\n    provisional: true\n")
	if p.CheckStatus("guardrails.TESTQ.focused-test-added", "1") != ActiveProvisional {
		t.Fatal("list form not honoured")
	}
	if p.Shadow() != (ShadowConfig{14, 5}) {
		t.Fatalf("defaults %+v", p.Shadow())
	}
	q := mustParse(t, "blocking:\n  shadow: {min_days: 7, min_pr_runs: 2}\n")
	if q.Shadow() != (ShadowConfig{7, 2}) {
		t.Fatalf("override %+v", q.Shadow())
	}
}

func TestShadowReady(t *testing.T) {
	p := Default()
	const id = "guardrails.SUPPRESS.added.unscoped"
	base := ShadowStats{CheckID: id, Days: 14, PRRuns: 5, WouldBlock: 3, Warranted: 3, TotalRuns: 20}
	with := func(f func(*ShadowStats)) ShadowStats { s := base; f(&s); return s }
	cases := []struct {
		name  string
		s     ShadowStats
		ready bool
		why   string
	}{
		{"boundary ok", base, true, "satisfied"},
		{"13 days", with(func(s *ShadowStats) { s.Days = 13 }), false, "min_days"},
		{"4 runs", with(func(s *ShadowStats) { s.PRRuns = 4 }), false, "min_pr_runs"},
		{"one not_warranted", with(func(s *ShadowStats) { s.NotWarranted = 1; s.Warranted = 2 }), false, "not_warranted"},
		{"11% unknown", with(func(s *ShadowStats) { s.TotalRuns = 100; s.UnknownAttribution = 11 }), false, "unknown"},
		{"10% unknown", with(func(s *ShadowStats) { s.TotalRuns = 100; s.UnknownAttribution = 10 }), true, ""},
		{"undecided pending", with(func(s *ShadowStats) { s.Undecided = 1 }), false, "undecided"},
		{"rare security zero events", ShadowStats{CheckID: "security.secrets.github-pat", Days: 14, TotalRuns: 3}, true, "provisional"},
		{"rare security short window", ShadowStats{CheckID: "security.secrets.github-pat", Days: 13}, false, "min_days"},
		{"rare security with events needs runs", ShadowStats{CheckID: "security.ci-mutable-action-ref", Days: 14, PRRuns: 2, WouldBlock: 1, Warranted: 1}, false, "min_pr_runs"},
		{"non-rare zero events not provisional", ShadowStats{CheckID: id, Days: 30, PRRuns: 2}, false, "min_pr_runs"},
		{"ineligible id", ShadowStats{CheckID: "security.secrets.generic-api-key", Days: 99, PRRuns: 99}, false, "excluded"},
	}
	for _, c := range cases {
		ok, why := p.ShadowReady(c.s)
		if ok != c.ready || !strings.Contains(why, c.why) {
			t.Errorf("%s: got (%v, %q)", c.name, ok, why)
		}
	}
	custom := mustParse(t, "blocking:\n  shadow: {min_days: 3, min_pr_runs: 1}\n")
	if ok, _ := custom.ShadowReady(ShadowStats{CheckID: id, Days: 3, PRRuns: 1, WouldBlock: 1, Warranted: 1}); !ok {
		t.Fatal("configured window ignored")
	}
}

func TestAcceptOverride(t *testing.T) {
	p := mustParse(t, guardianPolicy)
	co := []byte("# owners\n* @org/core carol\n/docs/ dave@example.com # trailing\n")
	cases := []struct {
		actor string
		want  bool
	}{
		{"alice", true}, {"Bob", true}, {"@alice", true},
		{"carol", true}, {"dave@example.com", true}, {"org/core", true},
		{"mallory", false}, {"", false}, {"owners", false},
	}
	for _, c := range cases {
		if got, why := p.AcceptOverride(c.actor, co); got != c.want {
			t.Errorf("%q = %v (%s)", c.actor, got, why)
		}
	}
	if ok, _ := Default().AcceptOverride("alice", nil); ok {
		t.Fatal("default policy has no maintainers")
	}
}

func run(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestLoadReadsBaseRefOnly(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "init", "-q", "-b", "main")
	write := func(content string) {
		p := filepath.Join(dir, Path)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Base without policy.
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "add", ".")
	run(t, dir, "commit", "-qm", "base")
	run(t, dir, "tag", "base-nopolicy")

	// Base with a report policy.
	write("profile: report\n")
	run(t, dir, "add", ".")
	run(t, dir, "commit", "-qm", "policy")
	run(t, dir, "branch", "base")

	// PR head rewrites the policy to activate a check, committed and in the working tree.
	run(t, dir, "checkout", "-q", "-b", "pr")
	write(guardianPolicy)
	run(t, dir, "add", ".")
	run(t, dir, "commit", "-qm", "pr activates")

	p, err := Load(dir, "base")
	if err != nil {
		t.Fatal(err)
	}
	if p.Profile() != "report" || p.CheckStatus("guardrails.SUPPRESS.added.unscoped", "1") != Shadow {
		t.Fatal("PR-head policy leaked into base load")
	}
	head, err := Load(dir, "pr")
	if err != nil || head.CheckStatus("guardrails.SUPPRESS.added.unscoped", "1") != Active {
		t.Fatalf("control: head policy should activate: %v", err)
	}

	abs, err := Load(dir, "base-nopolicy")
	if err != nil || abs.Profile() != "report" || len(abs.Activations()) != 0 {
		t.Fatalf("absent file: %v %+v", err, abs)
	}
	if _, err := Load(dir, "no-such-ref"); err == nil {
		t.Fatal("unresolvable ref must error")
	}

	// Invalid policy at base surfaces an error naming the ref.
	run(t, dir, "checkout", "-q", "-b", "bad", "main")
	write("profile: guardian\nblocking:\n  - id: guardrails.CONFIG.*\n")
	run(t, dir, "add", ".")
	run(t, dir, "commit", "-qm", "bad")
	if _, err := Load(dir, "bad"); err == nil || !strings.Contains(err.Error(), "wildcard") {
		t.Fatalf("expected wildcard error, got %v", err)
	}
}

package clawpatch

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"scout/internal/domain"
)

func TestMapReusesCanonicalRecordsWithoutExecutingClawpatch(t *testing.T) {
	root := t.TempDir()
	stateDir := filepath.Join(root, ".scout", "clawpatch")
	featuresDir := filepath.Join(stateDir, "features")
	if err := os.MkdirAll(featuresDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(featuresDir, "feature.json"), []byte(`{"schemaVersion":1,"featureId":"feature","kind":"service","source":"heuristic","confidence":"high"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	// Clawpatch never deletes records; a record the latest map did not emit is
	// rewritten with status "skipped" and must not reach ownership.
	if err := os.WriteFile(filepath.Join(featuresDir, "stale.json"), []byte(`{"schemaVersion":1,"featureId":"stale","kind":"service","source":"agent-mapper","confidence":"high","status":"skipped"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	binaryPath := filepath.Join(t.TempDir(), Revision, "dist", "cli.js")
	if err := os.MkdirAll(filepath.Dir(binaryPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binaryPath, []byte("this file must not be executed"), 0o700); err != nil {
		t.Fatal(err)
	}
	nodePath, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	client, err := New(root, Config{BinaryPath: binaryPath, NodePath: nodePath})
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.Map(context.Background(), false, MapSourceHeuristic)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != MapReused || len(result.Features) != 1 || result.Features[0].FeatureID != "feature" {
		t.Fatalf("unchanged map did not reuse exactly the active canonical records: %+v", result)
	}
	loaded, err := LoadFeatures(stateDir)
	if err != nil || len(loaded) != 1 || loaded[0].FeatureID != "feature" {
		t.Fatalf("LoadFeatures must exclude stale records: %+v, %v", loaded, err)
	}
}

// fakeMapper builds a client whose Node runtime is a shell script that records
// argv, reports the pinned version, and writes one feature on a successful map.
func fakeMapper(t *testing.T, egress bool, failAgent bool) (*Client, string) {
	t.Helper()
	root := t.TempDir()
	stateDir := filepath.Join(root, ".scout", "clawpatch")
	if err := os.MkdirAll(filepath.Join(stateDir, "features"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "project.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "config.json"), []byte(`{"schemaVersion":1,"stateDir":".clawpatch","include":["**/*"],"exclude":["dist/**"],"extra":{"keep":true}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	bundle := filepath.Join(t.TempDir(), Revision)
	binaryPath := filepath.Join(bundle, "dist", "cli.js")
	if err := os.MkdirAll(filepath.Dir(binaryPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binaryPath, nil, 0o700); err != nil {
		t.Fatal(err)
	}
	marker := reviewSafetyPatchVersion + " " + reviewSafetyPatchSHA256 + "\n"
	if err := os.WriteFile(filepath.Join(bundle, ".scout-review-safety"), []byte(marker), 0o600); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(t.TempDir(), "argv.log")
	failure := "0"
	if failAgent {
		failure = "1"
	}
	script := "#!/bin/sh\n" +
		"for a in \"$@\"; do if [ \"$a\" = --version ]; then echo " + Version + "; exit 0; fi; done\n" +
		"shift 1; echo \"$*\" >> '" + log + "'\n" +
		"case \"$*\" in\n" +
		"*'--source agent'*) if [ " + failure + " = 1 ]; then exit 3; fi ;;\n" +
		"esac\n" +
		"case \"$*\" in *' map '*) printf '%s' '{\"schemaVersion\":1,\"featureId\":\"heur\",\"kind\":\"service\",\"source\":\"heuristic\",\"confidence\":\"high\"}' > '" + filepath.Join(stateDir, "features", "heur.json") + "';; esac\n" +
		"case \"$*\" in *'--provider'*) env > '" + log + ".env';; esac\n"
	node := filepath.Join(t.TempDir(), "node")
	if err := os.WriteFile(node, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	client, err := New(root, Config{BinaryPath: binaryPath, NodePath: node, EgressAuthorized: egress})
	if err != nil {
		t.Fatal(err)
	}
	return client, log
}

func TestHeuristicMapRunsWithoutEgress(t *testing.T) {
	client, log := fakeMapper(t, false, false)
	result, err := client.Map(context.Background(), true, MapSourceHeuristic)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != MapMapped || result.Source != MapSourceHeuristic || result.ExplorationIncomplete || len(result.Features) != 1 {
		t.Fatalf("heuristic map did not succeed without egress: %+v", result)
	}
	argv, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	want := "--root " + client.Root() + " --state-dir " + client.StateDir() + " --json map --source heuristic\n"
	if string(argv) != want {
		t.Fatalf("argv = %q, want %q", argv, want)
	}
}

// useFakeCLIs puts empty executables named names on a narrowed PATH.
func useFakeCLIs(t *testing.T, names ...string) {
	t.Helper()
	dir := t.TempDir()
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir+":/usr/bin:/bin")
}

func setProviderCredentials(t *testing.T) {
	t.Helper()
	for _, name := range []string{"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "CLAWPATCH_OPENAI_API_KEY", "CLAWPATCH_ANTHROPIC_API_KEY", "CLAWPATCH_CODEX_API_KEY", "OPENROUTER_API_KEY"} {
		t.Setenv(name, "must-not-reach-clawpatch")
	}
	t.Setenv("CLAWPATCH_PROVIDER", "")
}

func assertCredentialFreeHostEnv(t *testing.T, envLog string) {
	t.Helper()
	raw, err := os.ReadFile(envLog)
	if err != nil {
		t.Fatalf("Clawpatch environment was not recorded: %v", err)
	}
	if !strings.Contains(string(raw), "CLAWPATCH_CLAUDE_AUTH_CONTEXT=host\n") {
		t.Fatalf("host auth context missing from environment")
	}
	if strings.Contains(string(raw), "must-not-reach-clawpatch") {
		t.Fatal("an API key reached Clawpatch")
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if name, _, _ := strings.Cut(line, "="); strings.Contains(strings.ToUpper(name), "API_KEY") {
			t.Fatalf("%s reached Clawpatch", name)
		}
	}
}

func TestAgentMapWithoutEgressIsBlocked(t *testing.T) {
	useFakeCLIs(t, "codex")
	client, log := fakeMapper(t, false, false)
	result, err := client.Map(context.Background(), true, MapSourceAgent)
	if err == nil || result.Status != MapBlocked {
		t.Fatalf("agent map without egress must be blocked: %+v, %v", result, err)
	}
	if _, statErr := os.Stat(log); statErr == nil {
		t.Fatal("blocked agent map invoked Clawpatch")
	}
}

func TestAgentMapWithoutModelCLIIsSkippedAndNeverInvokesClawpatch(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("CLAWPATCH_PROVIDER", "")
	client, log := fakeMapper(t, true, false)
	result, err := client.Map(context.Background(), true, MapSourceAgent)
	if err != nil || result.Status != MapSkipped || result.Reason != "skipped (no model CLI)" || result.ExplorationIncomplete {
		t.Fatalf("agent map without a model CLI must be skipped without error: %+v, %v", result, err)
	}
	if _, statErr := os.Stat(log); statErr == nil {
		t.Fatal("skipped agent map invoked Clawpatch")
	}
}

func TestAgentMapUsesCodexOrClaudeInHostAuthContextWithoutAPIKeys(t *testing.T) {
	for name, tc := range map[string]struct {
		installed []string
		configure string
		want      string
	}{
		"codex is the default": {installed: []string{"codex", "claude"}, want: "codex"},
		"falls back to claude": {installed: []string{"claude"}, want: "claude"},
		"explicit claude":      {installed: []string{"codex", "claude"}, configure: "claude", want: "claude"},
	} {
		t.Run(name, func(t *testing.T) {
			useFakeCLIs(t, tc.installed...)
			setProviderCredentials(t)
			client, log := fakeMapper(t, true, false)
			client.provider = tc.configure
			result, err := client.Map(context.Background(), true, MapSourceAgent)
			if err != nil || result.Status != MapMapped || result.Source != MapSourceAgent {
				t.Fatalf("agent map = %+v, %v", result, err)
			}
			argv, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(argv), "--json map --source agent --provider "+tc.want+"\n") {
				t.Fatalf("agent map argv = %q", argv)
			}
			assertCredentialFreeHostEnv(t, log+".env")
			env, err := os.ReadFile(log + ".env")
			if err != nil {
				t.Fatal(err)
			}
			// Clawpatch's claude default is 180 s, too short for an agent map.
			for _, want := range []string{"CLAWPATCH_CODEX_TIMEOUT_MS=900000", "CLAWPATCH_CLAUDE_TIMEOUT_MS=900000"} {
				if !strings.Contains(string(env), want) {
					t.Fatalf("agent map env lacks %s", want)
				}
			}
		})
	}
}

func TestAnyOtherClawpatchProviderIsRejected(t *testing.T) {
	useFakeCLIs(t, "codex", "claude")
	for _, provider := range []string{"acpx", "mock", "opencode", "grok"} {
		for _, viaEnv := range []bool{false, true} {
			client, log := fakeMapper(t, true, false)
			t.Setenv("CLAWPATCH_PROVIDER", "")
			if viaEnv {
				t.Setenv("CLAWPATCH_PROVIDER", provider)
			} else {
				client.provider = provider
			}
			result, err := client.Map(context.Background(), true, MapSourceAgent)
			if err == nil || result.Status != MapBlocked {
				t.Fatalf("provider %q (env=%t) must be rejected: %+v, %v", provider, viaEnv, result, err)
			}
			if _, statErr := os.Stat(log); statErr == nil {
				t.Fatalf("rejected provider %q invoked Clawpatch", provider)
			}
		}
	}
}

func TestReviewInvokesProviderInHostAuthContextWithoutAPIKeys(t *testing.T) {
	useFakeCLIs(t, "claude")
	setProviderCredentials(t)
	client, log := fakeMapper(t, true, false)
	if err := os.WriteFile(filepath.Join(client.StateDir(), "features", "feat.json"), []byte(`{"schemaVersion":1,"featureId":"feat","kind":"service","source":"heuristic","confidence":"high"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(client.StateDir(), "config.json"), []byte(`{"schemaVersion":1,"review":{"maxFindingsPerFeature":3}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Review(context.Background(), []string{"feat"}); err != nil {
		t.Fatal(err)
	}
	argv, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(argv), "--json review --feature feat --provider claude\n") {
		t.Fatalf("review argv = %q", argv)
	}
	assertCredentialFreeHostEnv(t, log+".env")
}

func TestReviewWithoutModelCLIIsSkippedNotBlockedOrClean(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("CLAWPATCH_PROVIDER", "")
	client, log := fakeMapper(t, false, false)
	result, err := client.Review(context.Background(), []string{"feat"})
	if err != nil || len(result.Coverage) != 1 || result.Coverage[0].Status != domain.CoverageSkipped || result.Coverage[0].Reason != "skipped (no model CLI)" {
		t.Fatalf("review coverage = %+v, %v", result.Coverage, err)
	}
	if _, statErr := os.Stat(log); statErr == nil {
		t.Fatal("skipped review invoked Clawpatch")
	}
}

func TestAgentMapFailureKeepsHeuristicFeaturesAndReportsIncomplete(t *testing.T) {
	useFakeCLIs(t, "codex")
	client, log := fakeMapper(t, true, true)
	result, err := client.Map(context.Background(), true, MapSourceAgent)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != MapMapped || result.Source != MapSourceHeuristic || !result.ExplorationIncomplete || len(result.Features) != 1 || result.Features[0].FeatureID != "heur" {
		t.Fatalf("agent failure must keep heuristic features and flag incompleteness: %+v", result)
	}
	argv, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(argv), "map --source agent --provider codex\n") || !strings.HasSuffix(string(argv), "map --source heuristic\n") {
		t.Fatalf("unexpected argv sequence: %q", argv)
	}
}

func TestMapExcludesScoutStateFromClawpatchInventory(t *testing.T) {
	client, _ := fakeMapper(t, false, false)
	if _, err := client.Map(context.Background(), true, MapSourceHeuristic); err != nil {
		t.Fatal(err)
	}
	if err := client.ensureScoutExcluded(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(client.StateDir(), "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	var config struct {
		Exclude []string        `json:"exclude"`
		Extra   map[string]bool `json:"extra"`
	}
	if err := json.Unmarshal(raw, &config); err != nil {
		t.Fatal(err)
	}
	if len(config.Exclude) != 2 || config.Exclude[0] != "dist/**" || config.Exclude[1] != ScoutExcludePattern || !config.Extra["keep"] {
		t.Fatalf("config must gain .scout/** once and keep other fields: %s", raw)
	}
}

func TestMapFailsClosedWithoutStateConfig(t *testing.T) {
	client, _ := fakeMapper(t, false, false)
	if err := os.Remove(filepath.Join(client.StateDir(), "config.json")); err != nil {
		t.Fatal(err)
	}
	if result, err := client.Map(context.Background(), true, MapSourceHeuristic); err == nil || result.Status != MapBlocked {
		t.Fatalf("missing state config must block the map: %+v, %v", result, err)
	}
}

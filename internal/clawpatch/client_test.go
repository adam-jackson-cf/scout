//go:build boundary

package clawpatch

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func setProviderCredentials(t *testing.T) {
	t.Helper()
	for _, name := range []string{"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "CLAWPATCH_OPENAI_API_KEY", "CLAWPATCH_ANTHROPIC_API_KEY", "CLAWPATCH_CODEX_API_KEY", "OPENROUTER_API_KEY"} {
		t.Setenv(name, "must-not-reach-clawpatch")
	}

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

func TestAgentMapUsesCodexLunaXHighInHostAuthContextWithoutAPIKeys(t *testing.T) {
	useFakeCLIs(t, "codex", "claude")
	setProviderCredentials(t)
	client, log := fakeMapper(t, true, false)
	client.provider = "codex"
	result, err := client.Map(context.Background(), true, MapSourceAgent)
	if err != nil || result.Status != MapMapped || result.Source != MapSourceAgent {
		t.Fatalf("agent map = %+v, %v", result, err)
	}
	argv, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(argv), "--json map --source agent --provider codex --model gpt-6-luna --reasoning-effort xhigh\n") {
		t.Fatalf("agent map argv = %q", argv)
	}
	assertCredentialFreeHostEnv(t, log+".env")
	env, err := os.ReadFile(log + ".env")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"CLAWPATCH_CODEX_TIMEOUT_MS=900000", "CLAWPATCH_CLAUDE_TIMEOUT_MS=900000"} {
		if !strings.Contains(string(env), want) {
			t.Fatalf("agent map env lacks %s", want)
		}
	}
}

func TestAgentMapIgnoresProviderConfig(t *testing.T) {
	useFakeCLIs(t, "codex", "claude")
	client, log := fakeMapper(t, true, false)
	client.provider = "claude"
	result, err := client.Map(context.Background(), true, MapSourceAgent)
	if err != nil || result.Status != MapMapped || result.Source != MapSourceAgent {
		t.Fatalf("map = %+v, %v", result, err)
	}
	argv, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(argv), "--json map --source agent --provider codex --model gpt-6-luna --reasoning-effort xhigh\n") {
		t.Fatalf("map argv = %q", argv)
	}
}

func TestReviewInvokesProviderInHostAuthContextWithoutAPIKeys(t *testing.T) {
	useFakeCLIs(t, "claude")
	setProviderCredentials(t)
	client, log := fakeMapper(t, true, false)
	client.provider = "claude"
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
	if !strings.Contains(string(argv), "map --source agent --provider codex --model gpt-6-luna --reasoning-effort xhigh\n") || !strings.HasSuffix(string(argv), "map --source heuristic\n") {
		t.Fatalf("unexpected argv sequence: %q", argv)
	}
}

func TestAgentMapZeroFeaturesKeepsHeuristicFeaturesAndReportsIncomplete(t *testing.T) {
	useFakeCLIs(t, "codex")
	// A successful agent invocation emits no records; the same harness emits
	// the "heur" feature when heuristic mapping is requested.
	client, _ := fakeMapperWithAgentScript(t, true, "exit 0;")
	result, err := client.Map(context.Background(), true, MapSourceAgent)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != MapMapped || result.Source != MapSourceHeuristic || !result.ExplorationIncomplete || len(result.Features) != 1 || result.Features[0].FeatureID != "heur" {
		t.Fatalf("zero-feature agent map must keep heuristic features and flag incompleteness: %+v", result)
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

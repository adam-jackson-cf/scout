//go:build boundary

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"scout/internal/domain"
	"scout/internal/state"
)

func TestInitAssociationTerraConsent(t *testing.T) {
	root := newInitRepo(t)
	if code := runInit(initArgs(root, "--model-cli", "codex", "--lane", "association-fallback"), strings.NewReader(""), io.Discard); code != exitComplete {
		t.Fatalf("init exit=%d", code)
	}
	p := loadCommitted(t, root)
	if !p.LaneOptedIn("association-fallback") || p.ModelCLI().Provider != "codex" {
		t.Fatalf("consent %+v", p.ModelCLI())
	}
}
func TestInitAssociationTerraDisclosure(t *testing.T) {
	root := newInitRepo(t)
	var prompts bytes.Buffer
	// Decline Jev, enable model CLI, select codex, decline the other lanes,
	// then explicitly consent to association-fallback.
	code := runInit([]string{"--root", root}, strings.NewReader("n\ny\ncodex\nn\nn\ny\n"), &prompts)
	if code != exitComplete {
		t.Fatalf("init exit=%d", code)
	}
	p := loadCommitted(t, root)
	if !p.LaneOptedIn("association-fallback") {
		t.Fatal("interactive fallback consent not captured")
	}
	disclosure := strings.ToLower(prompts.String())
	for _, term := range strings.Fields("codex openai excerpts pr title body catalogue") {
		if !strings.Contains(disclosure, term) {
			t.Fatalf("missing disclosure %q: %s", term, disclosure)
		}
	}
}
func terraScenarioRepo(t *testing.T) (string, string, *atomic.Int64) {
	t.Helper()
	t.Setenv("OPENROUTER_API_KEY", "cli-local-fixture-key")
	pol := cliAssociationPolicy(true, true, .60, .8) + "model_cli:\n  provider: claude\n  lanes: [association-fallback]\n"
	root, base, _, binary := cliAssociationFixture(t, pol)
	writeRepoFile(t, root, "second/worker.py", "def build_artifact(source):\n    return source.encode()\n")
	gitIn(t, root, "add", "second/worker.py")
	gitIn(t, root, "commit", "-qm", "add artifact worker")
	revision, err := exec.CommandContext(boundedTestContext(t), "git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	head := strings.TrimSpace(string(revision))
	cliAssociationEvent(t, "push", "refs/heads/main", base, head)
	return root, binary, cliAssociationJev(t)
}

func installTerraCLIs(t *testing.T, available bool) string {
	t.Helper()
	dir := t.TempDir()
	claudeTrace := filepath.Join(dir, "claude-calls")
	if err := os.WriteFile(filepath.Join(dir, "claude"), []byte("#!/bin/sh\necho called >> '"+claudeTrace+"'\nexit 99\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if available {
		script := "#!/bin/sh\nprompt=$(/bin/cat)\nout=''\nwhile [ $# -gt 0 ]; do if [ \"$1\" = '-o' ]; then shift; out=$1; fi; shift; done\ncase \"$prompt\" in *'second/worker.py'*) answer='{\"decision\":\"new\",\"new_title\":\"Packaging\",\"new_summary\":\"Builds distribution artifacts.\"}';; *) answer='{\"decision\":\"existing\",\"area_id\":\"area-0\"}';; esac\nprintf '%s' \"$answer\" > \"$out\"\nprintf 'model: gpt-6.1-terra\\nreasoning effort: high\\n' >&2\n"
		if err := os.WriteFile(filepath.Join(dir, "codex"), []byte(script), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	return claudeTrace
}

func runTerraAssess(t *testing.T, root, binary string) []domain.LaneHealth {
	t.Helper()
	// Architecture has no external file scanner; keep this association
	// scenario independent of scanner provisioning in both CLI-availability arms.
	text, code := captureStdout(t, func() int {
		return run(boundedTestContext(t), []string{"assess", "--root", root, "--clawpatch", binary, "--family", "architecture", "--json"})
	})
	raw := []byte(text)
	var output struct {
		ExitCode int    `json:"exitCode"`
		Error    string `json:"error"`
		Result   struct {
			Health []domain.LaneHealth `json:"health"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &output); err != nil {
		t.Fatalf("%v: %s", err, raw)
	}
	if output.Error != "" || code != exitComplete || output.ExitCode != exitComplete {
		t.Fatalf("optional fallback altered exit: code=%d %s", code, raw)
	}
	return output.Result.Health
}

func assertTerraHealth(t *testing.T, health []domain.LaneHealth, available bool) {
	t.Helper()
	found := false
	for _, h := range health {
		if h.Lane != "association-fallback" {
			continue
		}
		found = true
		if available {
			if h.Status != domain.LaneRan || h.Units != 2 {
				t.Fatalf("health %+v", h)
			}
		} else if h.Status != domain.LaneSkipped || h.Reason != "no model CLI" {
			t.Fatalf("health %+v", h)
		}
	}
	if !found {
		t.Fatalf("missing fallback health: %+v", health)
	}
}

func terraOutcomes(rows []domain.Association) (assigned, pending, unowned int) {
	for _, r := range rows {
		switch r.Outcome {
		case domain.AssociationAssigned:
			assigned++
		case domain.AssociationPendingNew:
			pending++
		default:
			unowned++
		}
	}
	return assigned, pending, unowned
}

func assertTerraPendingCatalogue(t *testing.T, s *state.Store) {
	t.Helper()
	aa, err := s.ListAreas(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, a := range aa {
		if a.Status == "pending" && a.Title == "Packaging" && a.Origin == "association" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("pending catalogue %+v", aa)
	}
}

func assertTerraRows(t *testing.T, root string, calls *atomic.Int64, available bool) {
	t.Helper()
	s, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOrFail(t, s)
	rows, err := s.ListAssociations(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() == 0 || len(rows) != 2 {
		t.Fatalf("Jev=%d rows=%+v", calls.Load(), rows)
	}
	assertTerraCounts(t, rows, available)
	if available {
		assertTerraPendingCatalogue(t, s)
	}
}

func assertTerraCounts(t *testing.T, rows []domain.Association, available bool) {
	t.Helper()
	assigned, pending, unowned := terraOutcomes(rows)
	if !available {
		if unowned != 2 || assigned != 0 || pending != 0 {
			t.Fatalf("unowned rows %+v", rows)
		}
		return
	}
	if assigned != 1 || pending != 1 || unowned != 0 {
		t.Fatalf("rows=%+v", rows)
	}
}

func cliTerraScenario(t *testing.T, available bool) {
	t.Helper()
	root, binary, calls := terraScenarioRepo(t)
	claudeTrace := installTerraCLIs(t, available)
	assertTerraHealth(t, runTerraAssess(t, root, binary), available)
	assertTerraRows(t, root, calls, available)
	if _, err := os.Stat(claudeTrace); !os.IsNotExist(err) {
		t.Fatalf("claude called: %v", err)
	}
}
func TestAssociationTerraPushScenario(t *testing.T)   { cliTerraScenario(t, true) }
func TestAssociationTerraAbsentScenario(t *testing.T) { cliTerraScenario(t, false) }

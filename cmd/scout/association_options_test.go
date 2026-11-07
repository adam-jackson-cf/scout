package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"scout/internal/association"
	"scout/internal/clawpatch"
	"scout/internal/decisions"
	"scout/internal/domain"
	"scout/internal/state"
)

const cliAssociationModel = "typesafe/jev-1.14-20261014"

type associationRoundTripper func(*http.Request) (*http.Response, error)

func (f associationRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// Redirect only the production Jev route to a real local HTTP boundary. Every
// other destination is rejected, so these CLI tests cannot contact a service.
func cliAssociationJev(t *testing.T) *atomic.Int64 {
	t.Helper()
	calls := new(atomic.Int64)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var request struct {
			Questions map[string]struct {
				Type string `json:"type"`
			} `json:"questions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			http.Error(w, "invalid fixture request", 400)
			return
		}
		answers := map[string]any{}
		for key, q := range request.Questions {
			switch q.Type {
			case "noul":
				answers[key] = map[string]any{"type": "noul", "noul": .5}
			case "choice":
				answers[key] = map[string]any{"type": "choice", "choice": "area-0", "probabilities": map[string]float64{"area-0": .7, "area-1": .3}, "confidence": .7}
			default:
				t.Errorf("unexpected Jev question type %q", q.Type)
				http.Error(w, "unexpected question", 400)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{"id": "cli-fixture", "model": cliAssociationModel, "answers": answers, "usage": map[string]any{"cost": .00001, "input_tokens": 10, "output_tokens": 2}}); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)
	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	production, err := url.Parse(decisions.DefaultEndpoint)
	if err != nil {
		t.Fatal(err)
	}
	original := http.DefaultTransport
	http.DefaultTransport = associationRoundTripper(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != production.Host || r.URL.Path != production.Path {
			return nil, fmt.Errorf("fixture refuses non-Jev destination")
		}
		local := r.Clone(r.Context())
		copyURL := *r.URL
		local.URL = &copyURL
		local.URL.Scheme = target.Scheme
		local.URL.Host = target.Host
		local.Host = target.Host
		return original.RoundTrip(local)
	})
	t.Cleanup(func() { http.DefaultTransport = original })
	return calls
}

func cliAssociationPolicy(enabled, ack bool, covered, choice float64) string {
	// A policy cannot enable decision consent without acknowledgement; the
	// missing-ack fixture therefore leaves that independent consent disabled.
	return fmt.Sprintf("profile: report\nmapping:\n  association:\n    enabled: %t\n    covered_score: %g\n    t_choice: %g\ndecisions:\n  enabled: %t\n  model: %s\n  max_per_run: 100\n  max_cost_usd: 1\n  egress_acknowledged: %t\n", enabled, covered, choice, ack, cliAssociationModel, ack)
}

func cliAssociationFixture(t *testing.T, pol string) (root, base, head, binary string) {
	t.Helper()
	root = gitRepo(t)
	writeRepoFile(t, root, "a.py", "def known():\n    return 1\n")
	writeRepoFile(t, root, ".scout/policy.yaml", pol)
	commitAll(t, root)
	gitIn(t, root, "branch", "-M", "main")
	revision := func() string {
		t.Helper()
		out, err := exec.CommandContext(boundedTestContext(t), "git", "-C", root, "rev-parse", "HEAD").Output()
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(string(out))
	}
	base = revision()
	gitIn(t, root, "checkout", "-q", "-b", "change")
	writeRepoFile(t, root, "new/service.py", "def serve(request):\n    return request\n")
	commitAll(t, root)
	gitIn(t, root, "checkout", "-q", "main")
	gitIn(t, root, "merge", "-q", "--no-ff", "-m", "merge service", "change")
	head = revision()
	s, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	catalogue, err := s.ListAreas(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`[{"scoutAreaId":"area-0","title":"Requests","origin":"agent","fileSet":["known/0.py"]},{"scoutAreaId":"area-1","title":"Artifacts","origin":"agent","fileSet":["known/1.py"]}]`), &catalogue); err != nil {
		t.Fatal(err)
	}
	if err := s.ReplaceAreas(context.Background(), catalogue); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	bundle := filepath.Join(t.TempDir(), clawpatch.Revision)
	binary = filepath.Join(bundle, "dist", "cli.js")
	writeRepoFile(t, bundle, "dist/cli.js", "")
	stateDir := clawpatch.StateDir(root)
	writeRepoFile(t, stateDir, "project.json", "{}")
	writeRepoFile(t, stateDir, "config.json", `{"schemaVersion":1,"stateDir":".clawpatch","include":["**/*"],"exclude":["dist/**"]}`)
	if err := os.MkdirAll(filepath.Join(stateDir, "features"), 0700); err != nil {
		t.Fatal(err)
	}
	feature := `{"schemaVersion":1,"featureId":"fixture","kind":"service","source":"heuristic","confidence":"high","ownedFiles":[{"path":"a.py","reason":"fixture"}]}`
	script := "#!/bin/sh\nfor a in \"$@\"; do if [ \"$a\" = --version ]; then echo " + clawpatch.Version + "; exit 0; fi; done\ncase \"$*\" in *' map '*) printf '%s' '" + feature + "' > '" + filepath.Join(stateDir, "features", "fixture.json") + "';; esac\n"
	nodeDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(nodeDir, "node"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", nodeDir+":/usr/bin:/bin")
	return
}

func cliAssociationEvent(t *testing.T, name, ref, base, head string) {
	t.Helper()
	t.Setenv("GITHUB_EVENT_NAME", name)
	if name == "" {
		t.Setenv("GITHUB_EVENT_PATH", "")
		return
	}
	payload := map[string]any{"ref": ref, "before": base, "after": head, "repository": map[string]string{"default_branch": "main"}}
	if name == "pull_request" {
		payload["pull_request"] = map[string]any{"head": map[string]any{"repo": map[string]bool{"fork": false}}}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "event.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_EVENT_PATH", path)
}

func cliAssociationAssess(t *testing.T, root, binary string, extra ...string) (domain.LaneHealth, []domain.Association) {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = writer
	defer func() { os.Stdout = original; reader.Close() }()
	captured := make(chan []byte, 1)
	go func() { b, _ := io.ReadAll(reader); captured <- b }()
	args := append([]string{"assess", "--root", root, "--clawpatch", binary, "--json"}, extra...)
	code := run(boundedTestContext(t), args)
	writer.Close()
	raw := <-captured
	var output struct {
		ExitCode int    `json:"exitCode"`
		Error    string `json:"error"`
		Result   struct {
			Health []domain.LaneHealth `json:"health"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &output); err != nil {
		t.Fatalf("decode assess output: %v\n%s", err, raw)
	}
	if output.Error != "" || output.ExitCode != code {
		t.Fatalf("assess failed: code=%d output=%s", code, raw)
	}
	s, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	rows, err := s.ListAssociations(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range output.Result.Health {
		if h.Lane == "association" {
			return h, rows
		}
	}
	t.Fatalf("assess output omitted association health: %s", raw)
	return domain.LaneHealth{}, nil
}

func TestAssociationCLIOptionsPolicyAndPushRange(t *testing.T) {
	for _, tc := range []struct {
		name            string
		covered, choice float64
		outcome, reason string
	}{
		{name: "covered threshold terminal", covered: .45, choice: .6, outcome: "none", reason: domain.AssociationReasonNoneUnsent},
		{name: "choice threshold unresolved", covered: .65, choice: .8, outcome: "unresolved", reason: domain.AssociationReasonBelowTChoice},
		{name: "choice threshold assigned", covered: .65, choice: .6, outcome: "assigned"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("OPENROUTER_API_KEY", "cli-local-fixture-key")
			root, base, head, binary := cliAssociationFixture(t, cliAssociationPolicy(true, true, tc.covered, tc.choice))
			cliAssociationEvent(t, "push", "refs/heads/main", base, head)
			calls := cliAssociationJev(t)
			h, rows := cliAssociationAssess(t, root, binary)
			if h.Status != domain.LaneRan || h.Units != 1 || len(rows) != 1 || calls.Load() == 0 {
				t.Fatalf("enabled default push: health=%+v rows=%+v calls=%d", h, rows, calls.Load())
			}
			a := rows[0]
			expected := association.Fingerprint{CoveredScore: tc.covered, TChoice: tc.choice, CoveredQuestion: decisions.MergeCovered.Version(cliAssociationModel), ChoiceQuestion: decisions.MergeWhichAreaTextDigest(), JevModel: cliAssociationModel, PromptDigest: association.TerraPromptDigest, FallbackModel: "gpt-6.1-terra"}
			if a.DecisionDigest != expected.Digest() || a.MergeCommit != head || len(a.Files) != 1 || a.Files[0] != "new/service.py" || a.JevOutcome != tc.outcome || a.Reason != tc.reason {
				t.Fatalf("policy/range decision: %+v", a)
			}
			if tc.outcome == "assigned" && (a.Outcome != domain.AssociationAssigned || len(a.AreaIDs) != 1 || a.AreaIDs[0] != "area-0") {
				t.Fatalf("assignment not persisted: %+v", a)
			}
			if tc.outcome == "none" && calls.Load() != 1 {
				t.Fatalf("covered_score should stop before Choice: %d calls", calls.Load())
			}
		})
	}
}

func TestAssociationCLIOptionsSkipReasons(t *testing.T) {
	for _, tc := range []struct {
		name, reason, event, ref, flag       string
		disabled, noKey, noAck, missingRange bool
	}{
		{name: "disabled", reason: "disabled", disabled: true, event: "push", ref: "refs/heads/main"},
		{name: "fork flag", reason: "fork", flag: "--fork", event: "push", ref: "refs/heads/main"},
		{name: "egress flag", reason: "egress not allowed", flag: "--no-egress", event: "push", ref: "refs/heads/main"},
		{name: "missing key", reason: "no jev key", noKey: true, event: "push", ref: "refs/heads/main"},
		{name: "missing acknowledgement", reason: "no acknowledgement", noAck: true, event: "push", ref: "refs/heads/main"},
		{name: "PR event", reason: "trigger", event: "pull_request", ref: "refs/pull/1/merge"},
		{name: "feature push", reason: "trigger", event: "push", ref: "refs/heads/feature"},
		{name: "absent range", reason: "push range unavailable", missingRange: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			key := "cli-local-fixture-key"
			if tc.noKey {
				key = ""
			}
			t.Setenv("OPENROUTER_API_KEY", key)
			root, base, head, binary := cliAssociationFixture(t, cliAssociationPolicy(!tc.disabled, !tc.noAck, .65, .6))
			if tc.missingRange {
				base, head = "", ""
			}
			cliAssociationEvent(t, tc.event, tc.ref, base, head)
			calls := cliAssociationJev(t)
			var flags []string
			if tc.flag != "" {
				flags = append(flags, tc.flag)
			}
			h, rows := cliAssociationAssess(t, root, binary, flags...)
			if h.Status != domain.LaneSkipped || h.Reason != tc.reason || h.Units != 0 || len(rows) != 0 || calls.Load() != 0 {
				t.Fatalf("CLI skip: health=%+v rows=%+v calls=%d; want skipped (%s)", h, rows, calls.Load(), tc.reason)
			}
		})
	}
}

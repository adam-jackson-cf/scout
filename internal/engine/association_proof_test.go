package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"scout/internal/association"
	"scout/internal/decisions"
	"scout/internal/domain"
	"scout/internal/policy"
	"scout/internal/state"
)

const associationProofModel = "typesafe/jev-1.14-20261014"

const associationAssignedSource = "def serve(request):\n    return {\"status\": \"ok\", \"request\": request}\n"
const associationUnresolvedSource = "def inspect_artifact(artifact):\n    return tuple(sorted(artifact.items()))\n"

func associationRepo(t *testing.T) (root, base, head string) {
	t.Helper()
	root = t.TempDir()
	put := func(p string, b []byte) {
		t.Helper()
		p = filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, b, 0644); err != nil {
			t.Fatal(err)
		}
	}
	revision := func() string {
		t.Helper()
		c := exec.CommandContext(boundedTestContext(t), "git", "-C", root, "rev-parse", "HEAD")
		b, err := c.Output()
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(string(b))
	}
	gitRun(t, root, "init", "-q", "-b", "main")
	put("a.py", []byte("def known():\n    return 1\n"))
	gitRun(t, root, "add", ".")
	gitRun(t, root, "commit", "-q", "-m", "base")
	base = revision()
	for i, name := range []string{"assigned", "unresolved"} {
		branch := fmt.Sprintf("feature-%d", i)
		gitRun(t, root, "checkout", "-q", "-b", branch)
		source := associationAssignedSource
		if name == "unresolved" {
			source = associationUnresolvedSource
		}
		put("new/"+name+".py", []byte(source))
		gitRun(t, root, "add", ".")
		gitRun(t, root, "commit", "-q", "-m", name)
		gitRun(t, root, "checkout", "-q", "main")
		gitRun(t, root, "merge", "-q", "--no-ff", "-m", "merge "+name, branch)
	}
	return root, base, revision()
}

func associationStore(t *testing.T, root string, n int) *state.Store {
	t.Helper()
	s, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	// Decode the public area payload through the store's inferred catalogue
	// type, keeping this engine boundary fixture independent of area internals.
	catalogue, err := s.ListAreas(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	payload := make([]map[string]any, n)
	for i := range payload {
		payload[i] = map[string]any{"scoutAreaId": fmt.Sprintf("area-%03d", i), "title": fmt.Sprintf("Responsibility %d", i), "origin": "agent", "fileSet": []string{fmt.Sprintf("known/%d.py", i)}}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &catalogue); err != nil {
		t.Fatal(err)
	}
	if err := s.ReplaceAreas(context.Background(), catalogue); err != nil {
		t.Fatal(err)
	}
	return s
}

// The server consumes actual decisions requests and returns valid typed answers;
// it never calls a model. Mode partial fails only the second merge's requests.
func associationServer(t *testing.T, mode string) (*decisions.Client, *atomic.Int64) {
	t.Helper()
	count := new(atomic.Int64)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		var req struct {
			State     json.RawMessage `json:"state"`
			Questions map[string]struct {
				Type string `json:"type"`
			} `json:"questions"`
		}
		if err := json.Unmarshal(raw, &req); err != nil {
			t.Error(err)
			http.Error(w, "invalid", 400)
			return
		}
		if mode == "failure" || (mode == "partial" && strings.Contains(string(req.State), "unresolved.py")) {
			http.Error(w, "fixture failure", 400)
			return
		}
		answers := map[string]any{}
		for key, q := range req.Questions {
			switch q.Type {
			case "noul":
				answers[key] = map[string]any{"type": "noul", "noul": .9}
			case "choice":
				p := .9
				if strings.Contains(string(req.State), "unresolved.py") {
					p = .51
				}
				answers[key] = map[string]any{"type": "choice", "choice": "area-000", "probabilities": map[string]float64{"area-000": p, "area-001": 1 - p}, "confidence": p}
			default:
				t.Errorf("unexpected question type %q", q.Type)
				http.Error(w, "unexpected", 400)
				return
			}
		}
		model := associationProofModel
		if mode == "mismatch" {
			model = "typesafe/jev-wrong-model"
		}
		cost := .00001
		if mode == "cost" {
			cost = 2
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{"id": "fixture", "model": model, "answers": answers, "usage": map[string]any{"input_tokens": 10, "output_tokens": 2, "cost": cost}}); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)
	return &decisions.Client{Endpoint: server.URL, Key: "fixture-local-key", MaxAttempts: 1, Sleep: func(context.Context, time.Duration) error { return nil }}, count
}
func associationOptions(base, head string, c *decisions.Client) association.LaneOptions {
	return association.LaneOptions{Base: base, Head: head, RunID: "proof-run", Mapping: policy.Association{Enabled: true, CoveredScore: .60, TChoice: .535}, Config: decisions.Config{Enabled: true, Model: associationProofModel, MaxPerRun: 100, MaxCostUSD: 100, EgressAcknowledged: true, Endpoint: c.Endpoint}, Eligibility: decisions.Eligibility{Enabled: true, HasKey: true, EgressAllowed: true}, Client: c, Fingerprint: association.Fingerprint{CoveredScore: .60, TChoice: .535, CoveredQuestion: decisions.MergeCovered.Version(associationProofModel), ChoiceQuestion: "frozen-choice-text-v1", JevModel: associationProofModel}}
}
func assertAssociationHealth(t *testing.T, h domain.LaneHealth, status domain.LaneStatus, reason string, units int) {
	t.Helper()
	if h.Lane != LaneAssociation || h.Status != status || h.Reason != reason || h.Units != units || h.Mode != domain.ModeSignal || string(h.Requires) != "decision-egress" || h.Required {
		t.Fatalf("association health %+v; want status %s reason %q units %d", h, status, reason, units)
	}
	text := string(status)
	if status == domain.LaneRan {
		text = fmt.Sprintf("ran (%d)", units)
	} else if status == domain.LaneSkipped {
		text = "skipped (" + reason + ")"
	}
	summary := Summary(domain.Gate{}, Ranking{}, nil, nil, 0, []domain.LaneHealth{h}, nil)
	if !strings.Contains(summary, "| "+text+" |") {
		t.Fatalf("summary lacks health %q: %s", text, summary)
	}
}

func TestAssociationLaneZeroCalls(t *testing.T) {
	for _, tc := range []struct {
		name, reason string
		trigger      domain.LaneTrigger
		size         int
		change       func(*association.LaneOptions)
	}{
		{name: "PR", reason: "trigger", trigger: domain.TriggerPR, size: 2},
		{name: "fork", reason: "fork", trigger: domain.TriggerPush, size: 2, change: func(o *association.LaneOptions) { o.Eligibility.Fork = true }},
		{name: "disabled", reason: "disabled", trigger: domain.TriggerPush, size: 2, change: func(o *association.LaneOptions) { o.Mapping.Enabled = false }},
		{name: "no key", reason: "no jev key", trigger: domain.TriggerPush, size: 2, change: func(o *association.LaneOptions) { o.Eligibility.HasKey = false; o.Client.Key = "" }},
		{name: "no acknowledgement", reason: "no acknowledgement", trigger: domain.TriggerPush, size: 2, change: func(o *association.LaneOptions) { o.Config.EgressAcknowledged = false }},
		{name: "egress denied", reason: "egress not allowed", trigger: domain.TriggerPush, size: 2, change: func(o *association.LaneOptions) { o.Eligibility.EgressAllowed = false }},
		{name: "catalogue empty", reason: "catalogue size", trigger: domain.TriggerPush, size: 0},
		{name: "catalogue one", reason: "catalogue size", trigger: domain.TriggerPush, size: 1},
		{name: "catalogue 256", reason: "catalogue size", trigger: domain.TriggerPush, size: 256},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, base, head := associationRepo(t)
			s := associationStore(t, root, tc.size)
			c, count := associationServer(t, "success")
			o := associationOptions(base, head, c)
			if tc.change != nil {
				tc.change(&o)
			}
			healths, err := RunAssociationLane(context.Background(), s, root, tc.trigger, o)
			if err != nil {
				t.Fatal(err)
			}
			h := healths[0]
			assertAssociationHealth(t, h, domain.LaneSkipped, tc.reason, 0)
			if count.Load() != 0 {
				t.Fatalf("skip called Jev %d times", count.Load())
			}
			rows, err := s.ListAssociations(context.Background())
			if err != nil || len(rows) != 0 {
				t.Fatalf("skip wrote rows %+v err %v", rows, err)
			}
		})
	}
}

func TestAssociationLaneFailureHealth(t *testing.T) {
	for _, mode := range []string{"failure", "partial", "mismatch"} {
		t.Run(mode, func(t *testing.T) {
			root, base, head := associationRepo(t)
			s := associationStore(t, root, 2)
			c, count := associationServer(t, mode)
			o := associationOptions(base, head, c)
			healths, err := RunAssociationLane(context.Background(), s, root, domain.TriggerPush, o)
			if err != nil {
				t.Fatal(err)
			}
			h := healths[0]
			want := domain.LaneFailed
			if mode == "partial" {
				want = domain.LaneRan
			}
			assertAssociationHealth(t, h, want, "", 2)
			if count.Load() == 0 {
				t.Fatal("no attempted requests")
			}
			rows, err := s.ListAssociations(context.Background())
			if err != nil || len(rows) != 2 {
				t.Fatalf("rows %+v %v", rows, err)
			}
			failures := 0
			for _, a := range rows {
				if a.Outcome == domain.AssociationUnresolved && a.Reason == domain.AssociationReasonJevFailure {
					failures++
				}
			}
			wantFailures := 2
			if mode == "partial" {
				wantFailures = 1
			}
			if failures != wantFailures {
				t.Fatalf("failure rows %d want %d: %+v", failures, wantFailures, rows)
			}
			disabled := o
			disabled.Mapping.Enabled = false
			disabled.RunID = "disabled"
			off, err := RunAssociationLane(context.Background(), s, root, domain.TriggerPush, disabled)
			if err != nil {
				t.Fatal(err)
			}
			onGate := Evaluate(GateInput{Policy: policy.Default(), Lanes: []domain.LaneHealth{h}}).Gate
			offGate := Evaluate(GateInput{Policy: policy.Default(), Lanes: off}).Gate
			if onGate.ExitCode != offGate.ExitCode {
				t.Fatalf("optional lane changed exit code %d -> %d", offGate.ExitCode, onGate.ExitCode)
			}
		})
	}
}

func TestAssociationLaneCaps(t *testing.T) {
	for _, mode := range []string{"call", "cost"} {
		t.Run(mode, func(t *testing.T) {
			root, base, head := associationRepo(t)
			s := associationStore(t, root, 2)
			c, count := associationServer(t, mode)
			o := associationOptions(base, head, c)
			if mode == "call" {
				o.Config.MaxPerRun = 1
			} else {
				o.Config.MaxCostUSD = .5
			}
			healths, err := RunAssociationLane(context.Background(), s, root, domain.TriggerPush, o)
			if err != nil {
				t.Fatal(err)
			}
			h := healths[0]
			if h.Status != domain.LaneRan {
				t.Fatalf("successful covered call then cap: %+v", h)
			}
			if count.Load() != 1 {
				t.Fatalf("cap allowed %d requests", count.Load())
			}
			rows, err := s.ListAssociations(context.Background())
			if err != nil || len(rows) != 2 {
				t.Fatalf("rows %+v %v", rows, err)
			}
			for _, a := range rows {
				if a.Outcome != domain.AssociationUnresolved || a.Reason != domain.AssociationReasonJevCap {
					t.Fatalf("capped cluster %+v", a)
				}
			}
		})
	}
}

func TestAssociationAssessScenario(t *testing.T) {
	var exits [2]int
	for i, enabled := range []bool{false, true} {
		t.Run(fmt.Sprintf("enabled=%v", enabled), func(t *testing.T) {
			root, base, head := associationRepo(t)
			s := associationStore(t, root, 2)
			s.Close()
			c, count := associationServer(t, "success")
			o := associationOptions(base, head, c)
			o.Mapping.Enabled = enabled
			binary := fixtureClawpatch(t, root)
			result, err := Assess(context.Background(), Config{Root: root, ClawpatchBinary: binary}, Options{Trigger: domain.TriggerPush, Association: &o})
			if err != nil {
				t.Fatal(err)
			}
			var h domain.LaneHealth
			found := false
			for _, lane := range result.Lanes {
				if lane.Lane == LaneAssociation {
					h = lane
					found = true
				}
			}
			if !found {
				t.Fatal("association health omitted from assess")
			}
			s, err = state.Open(root)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			rows, err := s.ListAssociations(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if !enabled {
				assertAssociationHealth(t, h, domain.LaneSkipped, "disabled", 0)
				if len(rows) != 0 || count.Load() != 0 {
					t.Fatalf("disabled rows=%d calls=%d", len(rows), count.Load())
				}
			} else {
				assertAssociationHealth(t, h, domain.LaneRan, "", 2)
				if len(rows) != 2 || count.Load() < 4 {
					t.Fatalf("enabled rows=%d calls=%d", len(rows), count.Load())
				}
				assigned, unresolved := 0, 0
				for _, a := range rows {
					if a.RunID != result.Run.RunID || a.DecidedAt.IsZero() || a.MergeCommit == "" || a.ClusterKey == "" || len(a.Files) != 1 || !a.ThresholdsProvisional || a.PRMetadata != domain.PRMetadataAbsent || a.InputsDigest == "" || a.CatalogueDigest == "" || a.DecisionDigest != o.Fingerprint.Digest() || a.JevCovered == nil || a.JevChoice == nil {
						t.Fatalf("incomplete persisted decision %+v", a)
					}
					if strings.Contains(a.RawAnswer, "fixture-local-key") {
						t.Fatal("transport credential persisted")
					}
					switch a.Outcome {
					case domain.AssociationAssigned:
						assigned++
						if len(a.AreaIDs) != 1 || a.AreaIDs[0] != "area-000" || a.JevOutcome != "assigned" {
							t.Fatalf("assignment mapping %+v", a)
						}
					case domain.AssociationUnresolved:
						unresolved++
						if a.Reason != domain.AssociationReasonBelowTChoice || a.JevOutcome != "unresolved" {
							t.Fatalf("abstention mapping %+v", a)
						}
					default:
						t.Fatalf("unexpected outcome %+v", a)
					}
				}
				if assigned != 1 || unresolved != 1 {
					t.Fatalf("assigned=%d unresolved=%d", assigned, unresolved)
				}
			}
			gate := Evaluate(GateInput{Policy: policy.Default(), Lanes: result.Lanes}).Gate
			exits[i] = gate.ExitCode
		})
	}
	if exits[0] != exits[1] {
		t.Fatalf("enabled lane changes exit %v", exits)
	}
}

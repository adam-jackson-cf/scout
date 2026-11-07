package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"scout/internal/decisions"
	"scout/internal/decisions/orderflip"
)

func flipRepo(t *testing.T, acknowledged bool) string {
	t.Helper()
	root := newInitRepo(t)
	policy := fmt.Sprintf("profile: report\ndecisions:\n  enabled: %t\n  provider: jev\n  model: typesafe/jev-1.14-20261014\n  max_per_run: 100\n  max_cost_usd: 1\n  egress_acknowledged: %t\n", acknowledged, acknowledged)
	writeRepoFile(t, root, ".scout/policy.yaml", policy)
	gitIn(t, root, "add", "-f", ".scout/policy.yaml")
	gitIn(t, root, "commit", "-qm", "consent")
	return root
}

func flipUnits(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "units.jsonl")
	var text strings.Builder
	for _, id := range strings.Fields("flip stable failed") {
		fmt.Fprintf(&text, `{"id":%q,"state":{"unit":{"role":%q},"catalogue":[{"id":"a","title":"A","summary":"First","sample_files":["a.py"]},{"id":"b","title":"B","summary":"Second","sample_files":["b.py"]}]}}`+"\n", id, id)
	}
	if err := os.WriteFile(path, []byte(text.String()), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func flipServer(t *testing.T, fail bool) *atomic.Int64 {
	t.Helper()
	calls := new(atomic.Int64)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var req struct {
			State struct {
				Unit struct {
					Role string `json:"role"`
				} `json:"unit"`
			} `json:"state"`
			Questions map[string]struct {
				Criteria json.RawMessage `json:"criteria"`
			} `json:"questions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			w.WriteHeader(400)
			return
		}
		if fail && req.State.Unit.Role == "failed" {
			w.WriteHeader(400)
			return
		}
		dec := json.NewDecoder(strings.NewReader(string(req.Questions["q1"].Criteria)))
		if _, err := dec.Token(); err != nil {
			t.Error(err)
			return
		}
		first, err := dec.Token()
		if err != nil {
			t.Error(err)
			return
		}
		chosen := "a"
		if req.State.Unit.Role == "flip" {
			chosen = first.(string)
		}
		pa, pb := .75, .25
		if chosen == "b" {
			pa, pb = .25, .75
		}
		fmt.Fprintf(w, `{"model":"typesafe/jev-1.14-20261014","answers":{"q1":{"type":"choice","choice":%q,"confidence":0.75,"probabilities":{"a":%v,"b":%v}}},"usage":{"cost":0.125}}`, chosen, pa, pb)
	}))
	t.Cleanup(server.Close)
	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	original := http.DefaultTransport
	http.DefaultTransport = associationRoundTripper(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != decisions.DefaultEndpoint {
			return nil, fmt.Errorf("unexpected destination")
		}
		local := r.Clone(r.Context())
		copied := *r.URL
		local.URL = &copied
		local.URL.Scheme = target.Scheme
		local.URL.Host = target.Host
		return original.RoundTrip(local)
	})
	t.Cleanup(func() { http.DefaultTransport = original })
	return calls
}

func flipCommand(t *testing.T, root, path, question string, asJSON bool) (string, int) {
	t.Helper()
	args := strings.Fields("decisions orderflip --root " + root + " --units " + path + " --question " + question)
	if asJSON {
		args = append(args, "--json")
	}
	return captureStdout(t, func() int { return run(t.Context(), args) })
}

func flipResult(t *testing.T, text string) (orderflip.Report, float64) {
	t.Helper()
	var reply struct {
		Result struct {
			orderflip.Report
			MaxShift float64 `json:"max_shift"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(text), &reply); err != nil {
		t.Fatal(err)
	}
	return reply.Result.Report, reply.Result.MaxShift
}

func TestOrderflipCLIExactMeasurement(t *testing.T) {
	calls := flipServer(t, false)
	t.Setenv("OPENROUTER_API_KEY", "fixture")
	root, path := flipRepo(t, true), flipUnits(t)
	text, code := flipCommand(t, root, path, "merge.which_area", true)
	report, shift := flipResult(t, text)
	if code != 0 || calls.Load() != 6 || report.Measured != 3 || report.Unmeasured != 0 || report.Flipped != 1 || report.FlipRate == nil || *report.FlipRate != 1.0/3 || report.CostUSD != .75 || shift != .5 || len(report.Reports) != 3 {
		t.Fatalf("code %d calls %d report %+v shift %v output %s", code, calls.Load(), report, shift, text)
	}
	for i, row := range report.Reports {
		if row.ID != strings.Fields("flip stable failed")[i] || row.Flipped != (i == 0) || row.CostUSD != .25 || row.Failure != "" || len(row.OrderA) != 2 || row.OrderA[0] != row.OrderB[1] || row.OrderA[1] != row.OrderB[0] {
			t.Fatalf("row %+v", row)
		}
		if i == 0 {
			if row.ChoiceA != row.OrderA[0] || row.ChoiceB != row.OrderB[0] || row.MaxShift != .5 {
				t.Fatalf("flipped row %+v", row)
			}
		} else if row.ChoiceA != "a" || row.ChoiceB != "a" || row.MaxShift != 0 {
			t.Fatalf("stable row %+v", row)
		}
	}
	summary, code := flipCommand(t, root, path, "merge.which_area", false)
	if code != 0 || !strings.Contains(summary, "flip rate: 0.333333") || !strings.Contains(summary, "measured: 3") || !strings.Contains(summary, "max probability shift: 0.5") {
		t.Fatalf("summary %q code %d", summary, code)
	}
}

func TestOrderflipCLIFailureExcluded(t *testing.T) {
	calls := flipServer(t, true)
	t.Setenv("OPENROUTER_API_KEY", "fixture")
	text, code := flipCommand(t, flipRepo(t, true), flipUnits(t), "merge.which_area", true)
	report, _ := flipResult(t, text)
	if code != 0 || calls.Load() != 6 || report.Measured != 2 || report.Unmeasured != 1 || report.FlipRate == nil || *report.FlipRate != .5 || report.Reports[2].Failure == "" || report.Reports[2].ChoiceA != "" || math.Abs(report.CostUSD-.5) > 1e-12 {
		t.Fatalf("code %d calls %d report %+v", code, calls.Load(), report)
	}
}

func TestOrderflipCLINoKey(t *testing.T) {
	calls := flipServer(t, false)
	t.Setenv("OPENROUTER_API_KEY", "")
	text, code := flipCommand(t, flipRepo(t, true), flipUnits(t), "merge.which_area", true)
	if code != exitUnavailable || calls.Load() != 0 || !strings.Contains(text, "no jev key") {
		t.Fatalf("code %d calls %d output %s", code, calls.Load(), text)
	}
}

func TestOrderflipCLINoAcknowledgement(t *testing.T) {
	calls := flipServer(t, false)
	t.Setenv("OPENROUTER_API_KEY", "fixture")
	text, code := flipCommand(t, flipRepo(t, false), flipUnits(t), "merge.which_area", true)
	if code != exitUnavailable || calls.Load() != 0 || !strings.Contains(text, "egress_acknowledged") {
		t.Fatalf("code %d calls %d output %s", code, calls.Load(), text)
	}
}

func TestOrderflipCLIRefusesNonChoice(t *testing.T) {
	calls := flipServer(t, false)
	t.Setenv("OPENROUTER_API_KEY", "fixture")
	text, code := flipCommand(t, flipRepo(t, true), flipUnits(t), "merge.covered", true)
	if code != exitUnavailable || calls.Load() != 0 || !strings.Contains(text, "only a Choice") {
		t.Fatalf("code %d calls %d output %s", code, calls.Load(), text)
	}
}

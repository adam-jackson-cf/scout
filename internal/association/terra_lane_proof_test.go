package association

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"scout/internal/decisions"
	"scout/internal/domain"
	"scout/internal/policy"
	"scout/internal/state"
)

func terraLaneFixture(t *testing.T, n int, mode string) (string, *state.Store, LaneOptions, *atomic.Int64) {
	t.Helper()
	root := t.TempDir()
	gitIn(t, root, "init", "-q", "-b", "main")
	write(t, root, "known.py", "def known():\n return 1\n")
	gitIn(t, root, "add", ".")
	gitIn(t, root, "commit", "-qm", "base")
	base := gitIn(t, root, "rev-parse", "HEAD")
	gitIn(t, root, "checkout", "-qb", "change")
	for i := range n {
		write(t, root, fmt.Sprintf("service%d/handler.py", i), fmt.Sprintf("def serve%d(request):\n return request\n", i))
	}
	gitIn(t, root, "add", ".")
	gitIn(t, root, "commit", "-qm", "change")
	gitIn(t, root, "checkout", "-q", "main")
	gitIn(t, root, "merge", "-q", "--no-ff", "-m", "merge change", "change")
	head := gitIn(t, root, "rev-parse", "HEAD")
	s, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	if err := s.ReplaceAreas(context.Background(), terraCatalogue()); err != nil {
		t.Fatal(err)
	}
	calls := new(atomic.Int64)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if mode == "failure" {
			http.Error(w, "fixture", 400)
			return
		}
		var req struct {
			State     json.RawMessage `json:"state"`
			Questions map[string]struct {
				Type string `json:"type"`
			} `json:"questions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		if err := os.WriteFile(filepath.Join(root, "jev-state"), req.State, 0600); err != nil {
			t.Error(err)
			return
		}
		answers := map[string]any{}
		for key, q := range req.Questions {
			if q.Type == "noul" {
				p := .1
				if mode == "unresolved" {
					p = .9
				}
				answers[key] = map[string]any{"type": "noul", "noul": p}
			} else {
				answers[key] = map[string]any{"type": "choice", "choice": "a", "confidence": .51, "probabilities": map[string]float64{"a": .51, "b": .49}}
			}
		}
		json.NewEncoder(w).Encode(map[string]any{"model": policy.DefaultModel, "answers": answers, "usage": map[string]any{"cost": .00001, "input_tokens": 10, "output_tokens": 2}})
	}))
	t.Cleanup(server.Close)
	cfg := policy.DefaultMapping().Association
	cfg.Enabled = true
	cfg.Fallback = terraConfig()
	cfg.MaxPendingPerRun = 10
	opts := LaneOptions{Base: base, Head: head, RunID: "current", Mapping: cfg, FallbackConsent: true, Config: decisions.Config{Enabled: true, Model: policy.DefaultModel, MaxPerRun: 100, MaxCostUSD: 1, EgressAcknowledged: true}, Eligibility: decisions.Eligibility{Enabled: true, HasKey: true, EgressAllowed: true}, Client: &decisions.Client{Endpoint: server.URL, Key: "fixture-local", MaxAttempts: 1}, Fingerprint: proofFingerprint()}
	return root, s, opts, calls
}
func executeTerraLane(t *testing.T, root string, s *state.Store, o LaneOptions) ([]domain.LaneHealth, []domain.Association) {
	t.Helper()
	hs, err := Execute(context.Background(), s, root, domain.TriggerPush, o)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := s.ListAssociations(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return hs, rows
}
func fallbackHealth(t *testing.T, hs []domain.LaneHealth, status domain.LaneStatus, reason string, n int) {
	t.Helper()
	for _, h := range hs {
		if h.Lane == "association-fallback" {
			if h.Status != status || h.Reason != reason || h.Units != n || h.Required || h.Mode != domain.ModeSignal || string(h.Requires) != "model-cli" {
				t.Fatalf("fallback health %+v", h)
			}
			return
		}
	}
	t.Fatal("missing fallback health")
}
func enableTerraGit(t *testing.T, dir string) { t.Helper(); t.Setenv("PATH", dir+":/usr/bin:/bin") }
func TestTerraSelectsCurrentOrdinaryClusters(t *testing.T) {
	root, s, o, _ := terraLaneFixture(t, 1, "unresolved")
	prior := proofRecord("earlier-none", "earlier.py")
	prior.JevOutcome = "none"
	prior.Outcome = domain.AssociationUnresolved
	prior.AreaIDs = nil
	prior.Reason = domain.AssociationReasonNoneUnsent
	if err := s.InsertAssociations(context.Background(), []domain.Association{prior}); err != nil {
		t.Fatal(err)
	}
	dir := fakeTerra(t, terraExisting(), terraDiagnostic(), "")
	enableTerraGit(t, dir)
	hs, rows := executeTerraLane(t, root, s, o)
	fallbackHealth(t, hs, domain.LaneRan, "", 1)
	if terraCallCount(t, dir) != 1 || len(rows) != 2 {
		t.Fatalf("calls=%d rows=%+v", terraCallCount(t, dir), rows)
	}
	for _, row := range rows {
		if row.RunID == "prior" && row.Outcome != domain.AssociationUnresolved {
			t.Fatal("earlier run revisited")
		}
		if row.RunID == "current" && row.Outcome != domain.AssociationAssigned {
			t.Fatalf("ordinary unresolved not assigned: %+v", row)
		}
	}
}
func TestTerraExcludesJevFailure(t *testing.T) {
	root, s, o, _ := terraLaneFixture(t, 1, "failure")
	dir := fakeTerra(t, terraExisting(), terraDiagnostic(), "")
	enableTerraGit(t, dir)
	_, rows := executeTerraLane(t, root, s, o)
	if terraCallCount(t, dir) != 0 || len(rows) != 1 || rows[0].Reason != domain.AssociationReasonJevFailure {
		t.Fatalf("calls=%d rows=%+v", terraCallCount(t, dir), rows)
	}
}
func TestTerraExcludesJevCap(t *testing.T) {
	root, s, o, _ := terraLaneFixture(t, 2, "none")
	o.Config.MaxPerRun = 1
	dir := fakeTerra(t, terraExisting(), terraDiagnostic(), "")
	enableTerraGit(t, dir)
	_, rows := executeTerraLane(t, root, s, o)
	capped := 0
	for _, r := range rows {
		if r.Reason == domain.AssociationReasonJevCap {
			capped++
		}
	}
	if terraCallCount(t, dir) != 1 || capped != 1 {
		t.Fatalf("calls=%d capped=%d rows=%+v", terraCallCount(t, dir), capped, rows)
	}
}
func TestTerraExcludesPRCap(t *testing.T) {
	root, s, o, _ := terraLaneFixture(t, 12, "none")
	dir := fakeTerra(t, terraExisting(), terraDiagnostic(), "")
	enableTerraGit(t, dir)
	_, rows := executeTerraLane(t, root, s, o)
	capped := 0
	for _, r := range rows {
		if r.Reason == domain.AssociationReasonPRCap {
			capped++
		}
	}
	if capped == 0 || terraCallCount(t, dir) != len(rows)-capped {
		t.Fatalf("calls=%d prcap=%d rows=%d", terraCallCount(t, dir), capped, len(rows))
	}
}
func TestTerraAttemptCap(t *testing.T) {
	root, s, o, _ := terraLaneFixture(t, 3, "none")
	o.Mapping.Fallback.MaxPerRun = 1
	dir := fakeTerra(t, terraExisting(), terraDiagnostic(), "")
	enableTerraGit(t, dir)
	hs, rows := executeTerraLane(t, root, s, o)
	fallbackHealth(t, hs, domain.LaneRan, "", 1)
	capped := 0
	for _, r := range rows {
		if r.Reason == domain.AssociationReasonFallbackCap && r.Outcome == domain.AssociationUnresolved {
			capped++
		}
	}
	if terraCallCount(t, dir) != 1 || capped != 2 {
		t.Fatalf("calls=%d cap=%d rows=%+v", terraCallCount(t, dir), capped, rows)
	}
}
func TestTerraPendingCap(t *testing.T) {
	root, s, o, _ := terraLaneFixture(t, 3, "none")
	o.Mapping.MaxPendingPerRun = 1
	dir := fakeTerra(t, terraNew(), terraDiagnostic(), "")
	enableTerraGit(t, dir)
	_, rows := executeTerraLane(t, root, s, o)
	pending, capped := 0, 0
	for _, r := range rows {
		if r.Outcome == domain.AssociationPendingNew {
			pending++
			if r.PendingAreaID == "" {
				t.Fatal("missing pending id")
			}
		}
		if r.Reason == domain.AssociationReasonPendingCap && r.Outcome == domain.AssociationUnresolved {
			capped++
		}
	}
	aa, err := s.ListAreas(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	stored := 0
	for _, a := range aa {
		if a.Status == "pending" {
			stored++
			if a.Origin != "association" || a.Title != "Artifacts" || a.Summary == "" || a.ClusterID == "" || a.DecisionID == "" || a.Active() {
				t.Fatalf("pending provenance %+v", a)
			}
		}
	}
	if pending != 1 || stored != 1 || capped != 2 {
		t.Fatalf("pending=%d stored=%d cap=%d rows=%+v", pending, stored, capped, rows)
	}
}
func TestTerraFailureHealth(t *testing.T) {
	root, s, o, _ := terraLaneFixture(t, 1, "none")
	dir := fakeTerra(t, "invalid", terraDiagnostic(), "")
	enableTerraGit(t, dir)
	hs, rows := executeTerraLane(t, root, s, o)
	fallbackHealth(t, hs, domain.LaneFailed, "", 1)
	if len(rows) != 1 || rows[0].Reason != domain.AssociationReasonTerraFailure || terraCallCount(t, dir) != 2 {
		t.Fatalf("rows=%+v", rows)
	}
}

func TestTerraPartialFailureHealth(t *testing.T) {
	root, s, o, _ := terraLaneFixture(t, 2, "none")
	markers := t.TempDir()
	before := "if [ ! -f '" + markers + "/first' ]; then : > '" + markers + "/first'; exit 17; fi\nif [ ! -f '" + markers + "/retry' ]; then : > '" + markers + "/retry'; exit 17; fi"
	dir := fakeTerra(t, terraExisting(), terraDiagnostic(), before)
	enableTerraGit(t, dir)
	hs, rows := executeTerraLane(t, root, s, o)
	fallbackHealth(t, hs, domain.LaneRan, "", 2)
	failed, assigned := 0, 0
	for _, row := range rows {
		if row.Reason == domain.AssociationReasonTerraFailure {
			failed++
		}
		if row.Outcome == domain.AssociationAssigned {
			assigned++
		}
	}
	if failed != 1 || assigned != 1 || terraCallCount(t, dir) != 3 {
		t.Fatalf("failed=%d assigned=%d calls=%d rows=%+v", failed, assigned, terraCallCount(t, dir), rows)
	}
}
func TestTerraSkipNoCodexKeepsJev(t *testing.T) {
	root, s, o, calls := terraLaneFixture(t, 1, "none")
	dir := fakeTerra(t, terraExisting(), terraDiagnostic(), "")
	if err := os.Remove(filepath.Join(dir, "codex")); err != nil {
		t.Fatal(err)
	}
	enableTerraGit(t, dir)
	hs, rows := executeTerraLane(t, root, s, o)
	fallbackHealth(t, hs, domain.LaneSkipped, "no model CLI", 0)
	if calls.Load() == 0 || len(rows) != 1 || rows[0].JevOutcome != "none" || rows[0].Outcome != domain.AssociationUnresolved {
		t.Fatalf("Jev calls=%d rows=%+v", calls.Load(), rows)
	}
	assertNoClaude(t, dir)
}
func TestTerraSkipNoConsentKeepsJev(t *testing.T) {
	root, s, o, calls := terraLaneFixture(t, 1, "none")
	o.FallbackConsent = false
	dir := fakeTerra(t, terraExisting(), terraDiagnostic(), "")
	enableTerraGit(t, dir)
	hs, rows := executeTerraLane(t, root, s, o)
	fallbackHealth(t, hs, domain.LaneSkipped, "no consent", 0)
	if calls.Load() == 0 || len(rows) != 1 || terraCallCount(t, dir) != 0 {
		t.Fatalf("Jev calls=%d rows=%+v", calls.Load(), rows)
	}
	assertNoClaude(t, dir)
}
func TestTerraUsesSameJevState(t *testing.T) {
	root, s, o, _ := terraLaneFixture(t, 1, "none")
	dir := fakeTerra(t, terraExisting(), terraDiagnostic(), "")
	enableTerraGit(t, dir)
	_, rows := executeTerraLane(t, root, s, o)
	b, err := os.ReadFile(filepath.Join(dir, "prompt"))
	if err != nil {
		t.Fatal(err)
	}
	jev, err := os.ReadFile(filepath.Join(root, "jev-state"))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || !strings.Contains(string(b), string(jev)) || rows[0].InputsDigest != laneDigest(string(jev)) {
		t.Fatal("fallback must reuse the identical bounded Jev state and digest")
	}
	if len(rows) != 1 || !strings.Contains(string(b), "def serve0") || !strings.Contains(string(b), "Requests") || !strings.Contains(string(b), "catalogue") {
		t.Fatalf("bounded Jev input not forwarded: rows=%+v", rows)
	}
}

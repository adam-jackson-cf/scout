package baseline

import (
	"encoding/json"
	"os/exec"
	"reflect"
	"scout/internal/domain"
	"scout/internal/state"
	"strings"
	"testing"
	"time"
)

func p05Git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	c := exec.CommandContext(boundedTestContext(t), "git", args...)
	c.Dir = dir
	b, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, b)
	}
	return strings.TrimSpace(string(b))
}
func p05Remote(t *testing.T) (string, []Options) {
	t.Helper()
	bare := t.TempDir()
	p05Git(t, bare, "init", "--bare", "--quiet")
	opts := []Options{}
	for range 2 {
		dir := t.TempDir()
		p05Git(t, dir, "init", "--quiet")
		p05Git(t, dir, "remote", "add", "origin", bare)
		opts = append(opts, Options{Repo: dir, Remote: bare, Trusted: true, Now: func() time.Time { return time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC) }})
	}
	return bare, opts
}
func p05Tree(t *testing.T, l domain.ShadowLedger) Tree {
	t.Helper()
	files, err := state.EncodeShadow(l, domain.ShadowInputs{})
	if err != nil {
		t.Fatal(err)
	}
	files[state.ExportIssues] = nil
	files[state.ExportDispositions] = nil
	files[state.ExportWouldBlock] = nil
	files[state.ExportCoverageSummary] = []byte("{}\n")
	files[state.ExportManifest] = state.RenderManifest(state.ManifestMeta{RepositoryID: "repo", WouldBlock: state.WouldBlockPresent}, files)
	return Tree(files)
}
func p05Decode(t *testing.T, raw string, dst any) {
	t.Helper()
	if err := json.Unmarshal([]byte(raw), dst); err != nil {
		t.Fatal(err)
	}
}
func p05Ledger() domain.ShadowLedger {
	at := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	stamp := at.Add(time.Hour)
	f := domain.PRRunFact{ID: domain.AttemptID{RepositoryID: "repo", RunID: "999", JobID: "1000", Attempt: "1"}, Scope: domain.ShadowScopeCI, PRNumber: "17", HeadCommit: "H", HeadTree: "tree", BaseCommit: "B", MergeBase: "MB", AssessedRevision: "merge", CheckID: "security:test", RuleVersion: "v1", CreatedAt: at.Add(-time.Minute), StartedAt: at, FinishedAt: at.Add(time.Minute), Status: "assessed", Comparison: "comparable", Inspection: "inspected", Pins: domain.ShadowPins{Scout: "scout", Tools: "tools", Catalogue: "catalogue", Policy: "policy", Recipe: "guardian"}, RecordedAt: &stamp}
	e := domain.ProviderPREvent{EventID: "event", RepositoryID: "repo", Scope: domain.ShadowScopeCI, PRNumber: "17", HeadCommit: "H", HeadTree: "tree", MergeBase: "MB", CheckID: "security:test", IssueID: "issue", RuleVersion: "v1", EventAt: at, RecordedAt: &stamp}
	link := domain.PRRunEventLink{AttemptID: f.ID, Scope: domain.ShadowScopeCI, EventID: e.EventID}
	c := domain.ShadowFrontier{RepositoryID: "repo", CheckID: "security:test", RuleVersion: "v1", CompleteThrough: at.Add(-time.Hour), PolicyRevision: "policy", MinDays: 14, MinObservations: 5}
	pub := domain.ShadowPublication{ID: "genesis", RepositoryID: "repo", RecordedAt: stamp, Facts: []domain.PRRunFact{f}, Events: []domain.ProviderPREvent{e}, Links: []domain.PRRunEventLink{link}, Frontiers: []domain.ShadowFrontier{c}}
	return domain.ShadowLedger{Facts: pub.Facts, Events: pub.Events, Links: pub.Links, Frontiers: pub.Frontiers, Publications: []domain.ShadowPublication{pub}}
}

// Baseline half of P05: actual fetched parent/chain and whole-history prefix.
// P05 in internal/shadow owns destination authorization/PreparePublication;
// sealed-local versus fresh-restore dossier equality is S4.
func TestADR012ImportWholeHistoryPrefix(t *testing.T) {
	source := p05Ledger()
	dst := source
	proof := domain.ChainProof{RepositoryID: "repo", StateRef: "refs/heads/scout/state", GenesisCommit: "genesis", SourceTip: "source-tip", DestinationTip: "destination-tip", Authenticated: true, DestinationAncestor: true}
	for _, tc := range []struct {
		name     string
		change   func(*domain.ShadowLedger, *domain.ChainProof)
		preserve bool
	}{
		{"authenticated_exact_history", func(*domain.ShadowLedger, *domain.ChainProof) {}, true},
		{"missing_ancestry", func(_ *domain.ShadowLedger, p *domain.ChainProof) { p.DestinationAncestor = false }, false},
		{"missing_genesis", func(_ *domain.ShadowLedger, p *domain.ChainProof) { p.GenesisCommit = "" }, false},
		{"fork_restored_claimed_chain", func(_ *domain.ShadowLedger, p *domain.ChainProof) { p.Authenticated = false }, false},
		{"nonempty_without_C", func(d *domain.ShadowLedger, _ *domain.ChainProof) { d.Frontiers = nil; d.Publications = nil }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var d domain.ShadowLedger
			raw, _ := json.Marshal(dst)
			p05Decode(t, string(raw), &d)
			p := proof
			tc.change(&d, &p)
			plan, err := PlanShadowImport(d, source, p)
			if err != nil {
				t.Fatal(err)
			}
			if plan.BytePreserving != tc.preserve {
				t.Fatalf("whole-history adoption=%v want %v", plan.BytePreserving, tc.preserve)
			}
			if tc.preserve {
				if !reflect.DeepEqual(plan.Adopt, source) {
					t.Fatal("same-lineage adoption rewrote post-C history")
				}
			} else if len(plan.Pending.Facts) != 1 {
				t.Fatal("cross-lineage path dropped post-C fact")
			}
		})
	}
	t.Run("destination_only_post_C_correction", func(t *testing.T) {
		d := p05Ledger()
		at := d.Publications[0].RecordedAt.Add(time.Millisecond)
		row := domain.ShadowTriage{ActionID: "repo:1000", EventID: "event", RepositoryID: "repo", Scope: domain.ShadowScopeCI, Verdict: domain.TriageNotWarranted, Actor: "alice", Reason: "correction", PolicyRevision: "current-policy", TriagedAt: at, RecordedAt: &at}
		d.Triage = []domain.ShadowTriage{row}
		d.Publications = append(d.Publications, domain.ShadowPublication{ID: "correction", ParentID: "genesis", RepositoryID: "repo", RecordedAt: at, Triage: d.Triage})
		plan, err := PlanShadowImport(d, source, proof)
		if err != nil {
			t.Fatal(err)
		}
		if plan.BytePreserving {
			t.Fatal("destination post-C correction ignored")
		}
	})
	t.Run("fast_forward_all_rows", func(t *testing.T) {
		src := p05Ledger()
		at := src.Publications[0].RecordedAt.Add(time.Millisecond)
		row := domain.ShadowTriage{ActionID: "repo:1000", EventID: "event", RepositoryID: "repo", Scope: domain.ShadowScopeCI, Verdict: domain.TriageWarranted, Actor: "alice", PolicyRevision: "policy", TriagedAt: at, RecordedAt: &at}
		src.Triage = []domain.ShadowTriage{row}
		src.Publications = append(src.Publications, domain.ShadowPublication{ID: "second", ParentID: "genesis", RepositoryID: "repo", RecordedAt: at, Triage: src.Triage})
		plan, err := PlanShadowImport(dst, src, proof)
		if err != nil {
			t.Fatal(err)
		}
		if !plan.BytePreserving || !reflect.DeepEqual(plan.Adopt, src) {
			t.Fatal("post-C rows lost")
		}
	})
	t.Run("chain_proof_uses_actual_fetched_parent", func(t *testing.T) {
		_, opts := p05Remote(t)
		ctx := boundedTestContext(t)
		seed, err := Publish(ctx, opts[0], func(parent Parent) (Tree, error) {
			if parent.Commit != "" {
				t.Fatal("unexpected genesis parent")
			}
			return p05Tree(t, dst), nil
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = Publish(ctx, opts[0], func(parent Parent) (Tree, error) {
			if parent.Commit != seed.Commit {
				t.Fatalf("claimed parent instead of fetched parent: %s", parent.Commit)
			}
			current, _, err := state.DecodeShadow(parent.Tree)
			if err != nil {
				return nil, err
			}
			p := proof
			p.DestinationTip = parent.Commit
			plan, err := PlanShadowImport(current, source, p)
			if err != nil {
				return nil, err
			}
			if !plan.BytePreserving {
				t.Fatal("proven identical parent refused")
			}
			return parent.Tree, nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}

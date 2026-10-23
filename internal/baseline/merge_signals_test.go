package baseline

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"scout/internal/domain"
	"scout/internal/state"
)

func signalFile(t *testing.T, signals ...domain.Signal) []byte {
	t.Helper()
	var out []byte
	for _, signal := range signals {
		raw, err := json.Marshal(signal)
		if err != nil {
			t.Fatal(err)
		}
		fields := map[string]any{}
		if err := json.Unmarshal(raw, &fields); err != nil {
			t.Fatal(err)
		}
		line, err := state.CanonicalLine(fields)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, line...)
	}
	return out
}

func TestMergeUnionsAppendOnlySignalsAndRegeneratesManifest(t *testing.T) {
	signal := func(snapshot string, value float64) domain.Signal {
		return domain.Signal{MetricID: "scout.cc", Version: 1, ScopeKind: domain.SignalScopeDirectory, ScopeAnchor: "pkg", Snapshot: snapshot,
			Numerator: domain.Float(value), Denominator: domain.Float(10), Value: domain.Float(value), PopulationDigest: "p", ToolDigest: "t", ConfigDigest: "c", Claim: domain.ClaimDescriptive}
	}
	path := state.ExportSignalsDir + "scout.cc@1.jsonl"
	current := ledger(t, "2026-10-01T00:00:00Z", "issue:a")
	current[path] = signalFile(t, signal("s1", 1), signal("s2", 2))
	current[state.ExportManifest] = state.RenderManifest(state.ManifestMeta{RepositoryID: "repo", WouldBlock: state.WouldBlockPresent}, current)
	local := ledger(t, "2026-10-01T00:00:00Z", "issue:a")
	local[path] = signalFile(t, signal("s2", 2), signal("s3", 3))
	local[state.ExportManifest] = state.RenderManifest(state.ManifestMeta{RepositoryID: "repo", WouldBlock: state.WouldBlockPresent}, local)

	merged, err := MergeState(current, local)
	if err != nil {
		t.Fatal(err)
	}
	var snapshots []string
	for _, line := range state.SplitLines(merged[path]) {
		var got domain.Signal
		if err := json.Unmarshal(line, &got); err != nil {
			t.Fatal(err)
		}
		snapshots = append(snapshots, got.Snapshot)
	}
	if got := strings.Join(snapshots, ","); got != "s1,s2,s3" {
		t.Fatalf("merged signal snapshots = %q, want the union s1,s2,s3", got)
	}
	sum := sha256.Sum256(merged[path])
	if manifest := string(merged[state.ExportManifest]); !strings.Contains(manifest, hex.EncodeToString(sum[:])) || !strings.Contains(manifest, path) {
		t.Fatalf("manifest does not describe the merged signals file: %s", manifest)
	}
}

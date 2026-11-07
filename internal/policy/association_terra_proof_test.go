package policy

import (
	"strings"
	"testing"
	"time"
)

func TestAssociationTerraPolicyDefaults(t *testing.T) {
	p, err := Parse([]byte("profile: report\n"))
	if err != nil {
		t.Fatal(err)
	}
	a := p.Mapping().Association
	if a.MaxPendingPerRun != 10 || a.Fallback.MaxPerRun != 40 || a.Fallback.Model != "gpt-6.1-terra" || a.Fallback.Effort != "high" || a.Fallback.Timeout != 900*time.Second {
		t.Fatalf("defaults %+v", a)
	}
	if p.LaneOptedIn("association-fallback") {
		t.Fatal("fallback silently consented")
	}
}
func TestAssociationTerraPolicyOverrides(t *testing.T) {
	p, err := Parse([]byte("profile: report\nmapping:\n  association:\n    max_pending_per_run: 3\n    fallback:\n      model: gpt-6.1-terra\n      effort: high\n      timeout: 2s\n      max_per_run: 7\nmodel_cli:\n  provider: claude\n  lanes: [association-fallback]\n"))
	if err != nil {
		t.Fatal(err)
	}
	a := p.Mapping().Association
	if a.MaxPendingPerRun != 3 || a.Fallback.MaxPerRun != 7 || a.Fallback.Timeout != 2*time.Second || !p.LaneOptedIn("association-fallback") || p.ModelCLI().Provider != "claude" {
		t.Fatalf("policy %+v %+v", a, p.ModelCLI())
	}
}
func rejectTerraPolicy(t *testing.T, entry string) {
	t.Helper()
	_, err := Parse([]byte("profile: report\nmapping:\n  association:\n" + entry))
	if err == nil {
		t.Fatalf("accepted invalid policy %s", entry)
	}
}
func TestAssociationTerraNegativeAttemptCap(t *testing.T) {
	rejectTerraPolicy(t, "    fallback: {max_per_run: -1}\n")
}
func TestAssociationTerraNegativePendingCap(t *testing.T) {
	rejectTerraPolicy(t, "    max_pending_per_run: -1\n")
}
func TestAssociationTerraNonintegerCap(t *testing.T) {
	rejectTerraPolicy(t, "    fallback: {max_per_run: 1.5}\n")
}
func TestAssociationTerraNegativeTimeout(t *testing.T) {
	rejectTerraPolicy(t, "    fallback: {timeout: -1s}\n")
}
func TestAssociationTerraZeroTimeout(t *testing.T) {
	rejectTerraPolicy(t, "    fallback: {timeout: 0s}\n")
}
func TestAssociationTerraMalformedTimeout(t *testing.T) {
	rejectTerraPolicy(t, "    fallback: {timeout: soon}\n")
}
func TestAssociationTerraAliasModel(t *testing.T) {
	rejectTerraPolicy(t, "    fallback: {model: latest}\n")
}
func TestAssociationTerraUnknownEffort(t *testing.T) {
	rejectTerraPolicy(t, "    fallback: {effort: impossible}\n")
}
func TestAssociationTerraConsentRoundTrip(t *testing.T) {
	provider := "codex"
	lanes := strings.Fields("association-fallback")
	b, changed, err := ApplyConsent(nil, ConsentUpdate{ModelCLIProvider: &provider, ModelCLILanes: &lanes})
	if err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	p, err := Parse(b)
	if err != nil || !p.LaneOptedIn("association-fallback") {
		t.Fatalf("consent not persisted: %s %v", b, err)
	}
}

package policy

import "testing"

func TestAssociationGuardDefaults(t *testing.T) {
	p, err := Parse([]byte("profile: report\n"))
	if err != nil {
		t.Fatal(err)
	}
	g := p.Mapping().Association.Guard
	if g.WindowMerges != 20 || g.MinClusters != 20 || g.FallbackRateMax != .40 || g.PendingNewRateMax != .20 {
		t.Fatalf("guard defaults=%+v", g)
	}
}
func TestAssociationGuardOverrides(t *testing.T) {
	p, err := Parse([]byte("mapping:\n  association:\n    guard: {window_merges: 7, min_clusters: 13, fallback_rate_max: 0, pending_new_rate_max: 1}\n"))
	if err != nil {
		t.Fatal(err)
	}
	g := p.Mapping().Association.Guard
	if g.WindowMerges != 7 || g.MinClusters != 13 || g.FallbackRateMax != 0 || g.PendingNewRateMax != 1 {
		t.Fatalf("guard overrides=%+v", g)
	}
}
func TestAssociationGuardBounds(t *testing.T) {
	for _, tc := range []struct{ name, value string }{
		{"zero merge window", "window_merges: 0"},
		{"negative merge window", "window_merges: -1"},
		{"fractional merge window", "window_merges: 1.5"},
		{"zero cluster minimum", "min_clusters: 0"},
		{"negative cluster minimum", "min_clusters: -1"},
		{"fractional cluster minimum", "min_clusters: 2.5"},
		{"negative fallback rate", "fallback_rate_max: -0.01"},
		{"fallback rate above one", "fallback_rate_max: 1.01"},
		{"nonfinite fallback rate", "fallback_rate_max: .nan"},
		{"negative pending rate", "pending_new_rate_max: -0.01"},
		{"pending rate above one", "pending_new_rate_max: 1.01"},
		{"nonfinite pending rate", "pending_new_rate_max: .inf"},
		{"string fallback rate", "fallback_rate_max: '0.4'"},
		{"unknown guard key", "windows: 20"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse([]byte("mapping:\n  association:\n    guard: {" + tc.value + "}\n"))
			if err == nil {
				t.Fatalf("accepted %s", tc.value)
			}
		})
	}
}

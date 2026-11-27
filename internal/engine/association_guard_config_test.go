package engine

import (
	"testing"

	"scout/internal/policy"
)

func TestAssociationGuardConfigDefaultsPreserveZeroLimits(t *testing.T) {
	if got, want := (Config{}).associationGuard(), policy.DefaultMapping().Association.Guard; got != want {
		t.Fatalf("default guard %+v; want %+v", got, want)
	}
	configured := policy.AssociationGuard{WindowMerges: 4, MinClusters: 2, FallbackRateMax: 0, PendingNewRateMax: 0}
	if got := (Config{AssociationGuard: configured}).associationGuard(); got != configured {
		t.Fatalf("configured zero limits changed: %+v", got)
	}
}

package engine

import (
	"testing"
)

func TestDeterministicRankingForAnUnavailableRunStillCarriesVersions(t *testing.T) {
	ranking := DeterministicRanking(nil, "run unavailable")
	if ranking.Method != "deterministic (run unavailable)" || ranking.BaselineVersion == "" || ranking.RankFormulaVersion == "" || ranking.Items == nil {
		t.Fatalf("%+v", ranking)
	}
}

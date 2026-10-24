package engine

import (
	"context"
	"errors"
	"strings"
	"testing"

	"scout/internal/decisions"
	"scout/internal/decisions/thresholds"
	"scout/internal/domain"
)

func TestJevRankingDisclosesProvisionalThresholdsAndStoresTheStateDigest(t *testing.T) {
	store, issues := seed(t, threeIssues)
	fake := newRankFake(t, map[string]float64{"src/a.go": 0.2, "src/b.go": 0.9, "src/c.go": 0.6})
	ranking, _, err := RankingOf(context.Background(), RankingInput{Issues: issues, RunID: "run-1", Trigger: domain.TriggerPR, Policy: consent(), Key: rankKey, Store: store, Client: fake.client()})
	if err != nil {
		t.Fatal(err)
	}
	if ranking.Thresholds != thresholds.StatusProvisional || ranking.Method != decisions.JevDisclosure {
		t.Fatalf("ranking %+v", ranking)
	}
	rows, err := store.ListAnnotations(context.Background(), "issue:security:a")
	if err != nil || len(rows) != 1 || len(rows[0].StateDigest) != 64 {
		t.Fatalf("annotations %+v err %v", rows, err)
	}
}

func TestThresholdsForAnotherVersionKeyAreStaleAndDisclosed(t *testing.T) {
	body := `{"schema":1,"decisions":[{"question_id":"triage.real_defect","version_key":"triage.real_defect@tuned-for-another-model","gating":false,"threshold":0.4,
	"min_precision":0.8,"min_recall":0.8,"minimums_provenance":"agreed","metric":"f1","label_set":{"summary":"40/40"},"score_space":"raw","status":"ready","cost_rule":"c","assumptions":[]}]}`
	store, issues := seed(t, threeIssues)
	fake := newRankFake(t, map[string]float64{"src/a.go": 0.2, "src/b.go": 0.9, "src/c.go": 0.6})
	ranking, _, err := RankingOf(context.Background(), RankingInput{Issues: issues, RunID: "run-1", Trigger: domain.TriggerPR, Policy: consent(), Key: rankKey, Store: store, Client: fake.client(),
		ReadThresholds: func() ([]byte, error) { return []byte(body), nil }})
	if err != nil {
		t.Fatal(err)
	}
	if order(ranking) != "bca" || ranking.Method != decisions.JevStaleDisclosure || !strings.HasPrefix(ranking.Thresholds, "provisional (stale") {
		t.Fatalf("a version-key change must keep ranking but mark thresholds stale: %+v", ranking)
	}
	for _, item := range ranking.Items {
		if item.Ranking != decisions.JevStaleDisclosure {
			t.Fatalf("item %+v", item)
		}
	}
}

func TestInvalidOrUnreadableThresholdsFallBackToTheDeterministicRankingWithoutACall(t *testing.T) {
	for name, read := range map[string]func() ([]byte, error){
		"invalid": func() ([]byte, error) {
			return []byte(`{"schema":1,"decisions":[{"question_id":"triage.real_defect"}]}`), nil
		},
		"unreadable": func() ([]byte, error) { return nil, errors.New("git failed") },
	} {
		t.Run(name, func(t *testing.T) {
			store, issues := seed(t, threeIssues)
			fake := newRankFake(t, nil)
			ranking, health, err := RankingOf(context.Background(), RankingInput{Issues: issues, RunID: "run-1", Trigger: domain.TriggerPR, Policy: consent(), Key: rankKey, Store: store, Client: fake.client(), ReadThresholds: read})
			if err != nil {
				t.Fatal(err)
			}
			if fake.count() != 0 || ranking.Mode != domain.RankingModeDeterministic || !strings.HasPrefix(ranking.Method, "deterministic (thresholds invalid") || health.Status != domain.LaneSkipped || health.Required {
				t.Fatalf("ranking %+v health %+v calls %d", ranking, health, fake.count())
			}
		})
	}
}

func TestPolicyCostCapPartiallyRanksAndIsDisclosed(t *testing.T) {
	store, issues := seed(t, threeIssues)
	fake := newRankFake(t, map[string]float64{"src/a.go": 0.2, "src/b.go": 0.9, "src/c.go": 0.6})
	pol := consent()
	pol.MaxCostUSD = 0.00007 // two calls at $0.00005 reach it
	ranking, health, err := RankingOf(context.Background(), RankingInput{Issues: issues, RunID: "run-1", Trigger: domain.TriggerPR, Policy: pol, Key: rankKey, Store: store, Client: fake.client()})
	if err != nil {
		t.Fatal(err)
	}
	if fake.count() != 2 || !strings.Contains(ranking.Method, "partial (2/3)") || health.Status != domain.LaneRan || !strings.Contains(health.Reason, "cost cap reached") {
		t.Fatalf("ranking %+v health %+v calls %d", ranking, health, fake.count())
	}
}

// Probabilities order findings only: no renderer shows one as a likelihood
// until a calibration check on Scout data passes (calibration-check: show
// probabilities as confidence only after the reliability criterion).
func TestSummaryNeverShowsAJevProbability(t *testing.T) {
	store, issues := seed(t, threeIssues)
	fake := newRankFake(t, map[string]float64{"src/a.go": 0.2271, "src/b.go": 0.9137, "src/c.go": 0.6389})
	ranking, _, err := RankingOf(context.Background(), RankingInput{Issues: issues, RunID: "run-1", Trigger: domain.TriggerPR, Policy: consent(), Key: rankKey, Store: store, Client: fake.client()})
	if err != nil {
		t.Fatal(err)
	}
	summary := Summary(domain.Gate{Status: "passed"}, ranking, issues, nil, 0, nil)
	for _, shown := range []string{"0.2271", "0.9137", "0.6389", "22.7", "91.3", "63.9", "probab", "likelihood", "confidence"} {
		if strings.Contains(strings.ToLower(summary), shown) {
			t.Fatalf("summary shows %q:\n%s", shown, summary)
		}
	}
}

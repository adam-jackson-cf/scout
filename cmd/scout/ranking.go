package main

import (
	"context"
	"os"

	"scout/internal/domain"
	"scout/internal/engine"
	"scout/internal/policy"
)

// jevKey reads the Jev key (OpenRouter) from the environment. It is passed on
// only to the decisions client; nothing prints it.
func jevKey() string { return os.Getenv("OPENROUTER_API_KEY") }

// headPolicy loads the policy committed at HEAD; policy.Load returns the
// default (decisions and model-CLI lanes off) only when HEAD has no policy
// file, and every other failure is returned.
func headPolicy(root string) (policy.Policy, error) {
	return policy.Load(root, "HEAD")
}

// rankAssessment ranks a plain assessment's findings and returns the ranking
// and the health with the decisions lane applied. Consent comes from pol, the
// policy committed at HEAD.
func rankAssessment(ctx context.Context, flags *commonFlags, pol policy.Policy, result engine.Result, verdict engine.GateOutput) (engine.Ranking, []domain.LaneHealth, error) {
	ranking, health, err := engine.RankingOf(ctx, engine.RankingInput{
		Issues: verdict.Issues, Root: flags.root, RunID: result.Run.RunID, Trigger: domain.TriggerPush,
		Policy: pol.Decisions(), Key: jevKey(), ReadThresholds: engine.CommittedThresholds(flags.root, "HEAD"),
	})
	if err != nil {
		return engine.Ranking{}, nil, err
	}
	return ranking, engine.ApplyDecisionsHealth(result.Lanes, health), nil
}

package main

import (
	"context"
	"os"
	"path/filepath"

	"scout/internal/domain"
	"scout/internal/engine"
	"scout/internal/policy"
)

// jevKey reads the Jev key (OpenRouter) from the environment. It is passed on
// only to the decisions client; nothing prints it.
func jevKey() string { return os.Getenv("OPENROUTER_API_KEY") }

// headPolicy loads the policy committed at HEAD; a repository without one
// gets the default (decisions and model-CLI lanes off).
func headPolicy(root string) (policy.Policy, error) {
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(policy.Path))); err != nil {
		return policy.Default(), nil
	}
	return policy.Load(root, "HEAD")
}

// rankAssessment ranks a plain assessment's findings and returns the ranking
// and the health with the decisions lane applied. Consent comes from the
// policy committed at HEAD; a repository without one has decisions off.
func rankAssessment(ctx context.Context, flags *commonFlags, result engine.Result, verdict engine.GateOutput) (engine.Ranking, []domain.LaneHealth, error) {
	pol, err := headPolicy(flags.root)
	if err != nil {
		return engine.Ranking{}, nil, err
	}
	ranking, health, err := engine.RankingOf(ctx, engine.RankingInput{
		Issues: verdict.Issues, Root: flags.root, RunID: result.Run.RunID, Trigger: domain.TriggerPush,
		Policy: pol.Decisions(), Key: jevKey(), ReadThresholds: engine.CommittedThresholds(flags.root, "HEAD"),
	})
	if err != nil {
		return engine.Ranking{}, nil, err
	}
	return ranking, engine.ApplyDecisionsHealth(result.Lanes, health), nil
}

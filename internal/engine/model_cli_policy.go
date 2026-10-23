package engine

import (
	"scout/internal/policy"
)

// ReasonNotOptedIn is the coverage and lane-health reason when a model-CLI
// lane has not been opted in through `scout init` (policy `model_cli.lanes`).
const ReasonNotOptedIn = "skipped (not opted in)"

// ApplyModelCLIPolicy binds the committed policy's model-CLI consent to cfg.
// Lanes run only when opted in. CLI choice precedence: policy
// `model_cli.provider`; else the SCOUT_MODEL_CLI value already in
// cfg.Provider.CLI (a local override honoured only when policy leaves the
// provider unset); else auto-detection. The same choice is used for the
// Clawpatch review provider.
func ApplyModelCLIPolicy(cfg Config, pol policy.Policy) Config {
	consent := pol.ModelCLI()
	cfg.ModelLanes = consent.Lanes
	if consent.Provider != "" {
		cfg.Provider.CLI = consent.Provider
	}
	cfg.ClawpatchProvider = cfg.Provider.CLI
	return cfg
}

func (c Config) laneOptedIn(lane string) bool {
	for _, l := range c.ModelLanes {
		if l == lane {
			return true
		}
	}
	return false
}

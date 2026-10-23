package clawpatch

import (
	"fmt"
	"os"
	"strings"

	"scout/internal/domain"
	"scout/internal/modelcli"
)

// Provider names Scout passes to Clawpatch. Clawpatch also knows acpx, grok,
// opencode, pi, cursor and mock; Scout rejects them (design rev 2.2 §4):
// reasoning-LLM use goes only through `codex` or `claude` in the CLI's own
// subscription login, with no API-key environment.
const (
	ProviderCodex  = string(modelcli.Codex)
	ProviderClaude = string(modelcli.Claude)

	// claudeHostAuthEnv switches Clawpatch's claude provider from --bare
	// (API key only) to --safe-mode with the host's `claude` login
	// (Claude Code >= 2.1.169; third_party/clawpatch/docs/providers.md).
	claudeHostAuthEnv = "CLAWPATCH_CLAUDE_AUTH_CONTEXT=host"

	// SkippedNoModelCLI is the reason when neither CLI is installed.
	SkippedNoModelCLI = modelcli.NoCLIReason
)

// resolveProvider picks the Clawpatch provider for a model-using invocation.
// The configured value, else CLAWPATCH_PROVIDER, must be codex or claude;
// anything else is rejected. With none configured, codex (Clawpatch's own
// default) then claude is used. The CLI must be on PATH; otherwise the error
// is modelcli.ErrNoCLI and the caller reports `skipped (no model CLI)`.
func (c *Client) resolveProvider() (string, error) {
	configured := strings.TrimSpace(c.provider)
	if configured == "" {
		configured = strings.TrimSpace(os.Getenv("CLAWPATCH_PROVIDER"))
	}
	switch configured {
	case ProviderCodex, ProviderClaude:
		if _, _, err := modelcli.Detect(configured); err != nil {
			return "", err
		}
		return configured, nil
	case "":
		for _, name := range []string{ProviderCodex, ProviderClaude} {
			if _, _, err := modelcli.Detect(name); err == nil {
				return name, nil
			}
		}
		return "", modelcli.ErrNoCLI
	default:
		return "", fmt.Errorf("Clawpatch provider %q is not allowed: only codex and claude are supported", configured)
	}
}

// resolveMapProvider picks the provider for the agent map: always codex.
// The configured provider and CLAWPATCH_PROVIDER govern only the
// clawpatch-review lane and are ignored here. Without codex on PATH the error
// is modelcli.ErrNoCLI.
func (c *Client) resolveMapProvider() (string, error) {
	if _, _, err := modelcli.Detect(ProviderCodex); err != nil {
		return "", err
	}
	return ProviderCodex, nil
}

// modelInvocation returns the provider arguments and environment for a
// model-using Clawpatch call: host auth context, no API-key variables.
func modelInvocation(provider string, extraEnv ...string) ([]string, []string) {
	env := append([]string{claudeHostAuthEnv}, extraEnv...)
	return []string{"--provider", provider}, env
}

func skippedReviewResult(ids []string) ReviewResult {
	coverage := make([]domain.Coverage, 0, len(ids))
	for _, id := range ids {
		coverage = append(coverage, domain.Coverage{Key: reviewCoverageKey(id), Status: domain.CoverageSkipped, Reason: SkippedNoModelCLI})
	}
	return ReviewResult{Status: domain.AdapterPartial, Coverage: coverage, Reason: SkippedNoModelCLI}
}

package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"scout/internal/policy"
)

// initCmd records onboarding consent in .scout/policy.yaml (design §6 "Launch
// default and fallback", §7.4 first adoption). Interactive unless --yes. It
// touches only profile, decisions and model_cli, so re-running is idempotent
// and never rewrites activated `blocking` entries.
func initCmd(_ context.Context, args []string) int {
	return runInit(args, os.Stdin, os.Stderr)
}

func runInit(args []string, in io.Reader, prompts io.Writer) int {
	set := flag.NewFlagSet("init", flag.ContinueOnError)
	flags := common(set)
	var yes, decisions, ack bool
	var profile, model, cli string
	var maxPerRun int
	var maxCost float64
	var lanes stringsFlag
	set.BoolVar(&yes, "yes", false, "non-interactive: use flags only, never prompt")
	set.StringVar(&profile, "profile", "", "report|guardian (new files default to report; an existing profile is kept)")
	set.BoolVar(&decisions, "decisions", false, "enable Jev (OpenRouter) ranking; needs --decisions-model and --acknowledge-egress")
	set.StringVar(&model, "decisions-model", "", "pinned Jev model id (a new pin defaults to "+policy.DefaultModel+"; aliases and undated lines such as jev-latest or typesafe/jev-1.14 are rejected)")
	set.IntVar(&maxPerRun, "max-per-run", 0, "decisions.max_per_run call cap (default 2000)")
	set.Float64Var(&maxCost, "max-cost-usd", 0, "decisions.max_cost_usd billed-cost cap per run in USD (default 0.5)")
	set.BoolVar(&ack, "acknowledge-egress", false, "acknowledge that bounded, redacted code excerpts are sent to the provider")
	set.StringVar(&cli, "model-cli", "", "claude|codex: CLI used by opted-in model-CLI lanes")
	set.Var(&lanes, "lane", "model-CLI lane to opt in: contextual|clawpatch-review (repeatable)")
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	given := map[string]bool{}
	set.Visit(func(f *flag.Flag) { given[f.Name] = true })

	if info, err := os.Stat(flags.root); err != nil || !info.IsDir() {
		return emit(flags, "unavailable", exitUnavailable, nil, fmt.Errorf("--root %q is not a directory", flags.root))
	}
	path := filepath.Join(flags.root, filepath.FromSlash(policy.Path))
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}

	if !yes {
		ask := &asker{r: bufio.NewReader(in), w: prompts}
		if !given["decisions"] {
			decisions, given["decisions"] = ask.yesNo("Enable Jev (OpenRouter) ranking of findings? Advisory only; it never gates."), true
		}
		if decisions {
			if !given["decisions-model"] {
				model, given["decisions-model"] = ask.line("Pinned Jev model id, no aliases (empty = "+policy.DefaultModel+")"), true
			}
			if !given["acknowledge-egress"] {
				ack, given["acknowledge-egress"] = ask.yesNo("Acknowledge that bounded, redacted code excerpts are sent to OpenRouter on trusted runs?"), true
			}
		}
		if !given["model-cli"] && !given["lane"] && ask.yesNo("Opt in to model-CLI lanes (claude/codex, only where the CLI is logged in)?") {
			cli, given["model-cli"] = ask.line("Model CLI (claude|codex)"), true
			if ask.yesNo("  contextual lane?") {
				lanes, given["lane"] = append(lanes, policy.LaneContextual), true
			}
			if ask.yesNo("  clawpatch-review lane?") {
				lanes, given["lane"] = append(lanes, policy.LaneClawpatchRev), true
			}
		}
	}

	if decisions && given["decisions"] && model == "" {
		// A new pin defaults to Jev 1.14; an existing pin is never replaced or
		// cleared, so re-running init stays idempotent.
		existingModel := ""
		if len(strings.TrimSpace(string(existing))) != 0 {
			if p, err := policy.Parse(existing); err == nil {
				existingModel = p.Decisions().Model
			}
		}
		given["decisions-model"] = existingModel == ""
		if existingModel == "" {
			model = policy.DefaultModel
		}
	}

	var u policy.ConsentUpdate
	if given["profile"] {
		u.Profile = &profile
	}
	if given["decisions"] {
		u.DecisionsEnabled = &decisions
	}
	if given["decisions-model"] {
		u.Model = &model
	}
	if given["max-per-run"] {
		u.MaxPerRun = &maxPerRun
	}
	if given["max-cost-usd"] {
		u.MaxCostUSD = &maxCost
	}
	if given["acknowledge-egress"] || given["decisions"] {
		u.EgressAcknowledged = &ack
	}
	if given["model-cli"] {
		u.ModelCLIProvider = &cli
	}
	if given["lane"] {
		l := []string(lanes)
		u.ModelCLILanes = &l
	}
	out, changed, err := policy.ApplyConsent(existing, u)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	if changed {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return emit(flags, "unavailable", exitUnavailable, nil, err)
		}
		if err := os.WriteFile(path, out, 0o644); err != nil {
			return emit(flags, "unavailable", exitUnavailable, nil, err)
		}
	}
	p, err := policy.Parse(out)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	d := p.Decisions()
	return emit(flags, "complete", exitComplete, map[string]any{
		"path": policy.Path, "changed": changed, "profile": p.Profile(),
		"decisions": map[string]any{"enabled": d.Enabled, "model": d.Model, "maxPerRun": d.MaxPerRun, "maxCostUsd": d.MaxCostUSD, "egressAcknowledged": d.EgressAcknowledged},
		"modelCli":  map[string]any{"provider": p.ModelCLI().Provider, "lanes": p.ModelCLI().Lanes},
		"next":      "commit .scout/policy.yaml, add the scout-guardian workflow, then run a full default-branch run to publish the baseline",
	}, nil)
}

type asker struct {
	r *bufio.Reader
	w io.Writer
}

func (a *asker) line(q string) string {
	fmt.Fprintf(a.w, "%s: ", q)
	s, _ := a.r.ReadString('\n')
	return strings.TrimSpace(s)
}

func (a *asker) yesNo(q string) bool {
	switch strings.ToLower(a.line(q + " [y/N]")) {
	case "y", "yes":
		return true
	}
	return false
}

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

// initCmd records onboarding consent in .scout/policy.yaml (REQ-97, REQ-27).
// Interactive unless --yes. It
// touches only profile, decisions and model_cli, so re-running is idempotent
// and never rewrites activated `blocking` entries.
func initCmd(_ context.Context, args []string) int {
	return runInit(args, os.Stdin, os.Stderr)
}

type initFlags struct {
	yes, decisions, ack bool
	profile, model, cli string
	maxPerRun           int
	maxCost             float64
	lanes               stringsFlag
	given               map[string]bool
}

func registerInitFlags(set *flag.FlagSet) *initFlags {
	opts := &initFlags{given: map[string]bool{}}
	set.BoolVar(&opts.yes, "yes", false, "non-interactive: use flags only, never prompt")
	set.StringVar(&opts.profile, "profile", "", "report|guardian (new files default to report; an existing profile is kept)")
	set.BoolVar(&opts.decisions, "decisions", false, "enable Jev (OpenRouter) ranking; needs --decisions-model and --acknowledge-egress")
	set.StringVar(&opts.model, "decisions-model", "", "pinned Jev model id (a new pin defaults to "+policy.DefaultModel+"; aliases and undated lines such as jev-latest or typesafe/jev-1.14 are rejected)")
	set.IntVar(&opts.maxPerRun, "max-per-run", 0, "decisions.max_per_run call cap (default 2000)")
	set.Float64Var(&opts.maxCost, "max-cost-usd", 0, "decisions.max_cost_usd billed-cost cap per run in USD (default 0.5)")
	set.BoolVar(&opts.ack, "acknowledge-egress", false, "acknowledge that bounded, redacted code excerpts are sent to the provider")
	set.StringVar(&opts.cli, "model-cli", "", "claude|codex: CLI used by opted-in lanes (association-fallback always uses codex)")
	set.Var(&opts.lanes, "lane", "model-CLI lane to opt in: contextual|clawpatch-review|association-fallback (repeatable)")
	return opts
}

func runInit(args []string, in io.Reader, prompts io.Writer) int {
	set := flag.NewFlagSet("init", flag.ContinueOnError)
	flags := common(set)
	opts := registerInitFlags(set)
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	set.Visit(func(f *flag.Flag) { opts.given[f.Name] = true })

	path, existing, err := readExistingPolicy(flags.root)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	if err := opts.prompt(in, prompts); err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	opts.defaultModel(existing)

	out, changed, err := policy.ApplyConsent(existing, opts.update())
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	if err := writePolicyIfChanged(path, out, changed); err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	p, err := policy.Parse(out)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	return emit(flags, "complete", exitComplete, initSummary(p, changed), nil)
}

func readExistingPolicy(root string) (string, []byte, error) {
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		return "", nil, fmt.Errorf("--root %q is not a directory", root)
	}
	path := filepath.Join(root, filepath.FromSlash(policy.Path))
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", nil, err
	}
	return path, existing, nil
}

func (o *initFlags) prompt(in io.Reader, prompts io.Writer) error {
	if o.yes {
		return nil
	}
	ask := &asker{r: bufio.NewReader(in), w: prompts}
	o.askDecisions(ask)
	o.askModelCLI(ask)
	return ask.err
}

func (o *initFlags) askDecisions(ask *asker) {
	if !o.given["decisions"] {
		o.decisions, o.given["decisions"] = ask.yesNo("Enable Jev (OpenRouter) ranking of findings? Advisory only; it never gates."), true
	}
	if o.decisions {
		if !o.given["decisions-model"] {
			o.model, o.given["decisions-model"] = ask.line("Pinned Jev model id, no aliases (empty = "+policy.DefaultModel+")"), true
		}
		if !o.given["acknowledge-egress"] {
			o.ack, o.given["acknowledge-egress"] = ask.yesNo("Acknowledge that bounded, redacted code excerpts are sent to OpenRouter on trusted runs?"), true
		}
	}
}

func (o *initFlags) askModelCLI(ask *asker) {
	if !o.given["model-cli"] && !o.given["lane"] && ask.yesNo("Opt in to model-CLI lanes (claude/codex, only where the CLI is logged in)?") {
		o.cli, o.given["model-cli"] = ask.line("Model CLI (claude|codex)"), true
		if ask.yesNo("  contextual lane?") {
			o.lanes, o.given["lane"] = append(o.lanes, policy.LaneContextual), true
		}
		if ask.yesNo("  clawpatch-review lane?") {
			o.lanes, o.given["lane"] = append(o.lanes, policy.LaneClawpatchRev), true
		}
		if ask.yesNo("  association-fallback lane? Sends bounded file excerpts, PR title/body and catalogue content through codex to OpenAI, regardless of the selected model CLI provider.") {
			o.lanes, o.given["lane"] = append(o.lanes, policy.LaneAssociationFallback), true
		}
	}
}

func (o *initFlags) defaultModel(existing []byte) {
	if !o.decisions || !o.given["decisions"] || o.model != "" {
		return
	}
	// A new pin defaults to Jev 1.14; an existing pin is never replaced or
	// cleared, so re-running init stays idempotent.
	existingModel := committedDecisionsModel(existing)
	o.given["decisions-model"] = existingModel == ""
	if existingModel == "" {
		o.model = policy.DefaultModel
	}
}

func committedDecisionsModel(existing []byte) string {
	if len(strings.TrimSpace(string(existing))) == 0 {
		return ""
	}
	p, err := policy.Parse(existing)
	if err != nil {
		return ""
	}
	return p.Decisions().Model
}

func (o *initFlags) update() policy.ConsentUpdate {
	var u policy.ConsentUpdate
	if o.given["profile"] {
		u.Profile = &o.profile
	}
	if o.given["decisions"] {
		u.DecisionsEnabled = &o.decisions
	}
	if o.given["decisions-model"] {
		u.Model = &o.model
	}
	if o.given["max-per-run"] {
		u.MaxPerRun = &o.maxPerRun
	}
	if o.given["max-cost-usd"] {
		u.MaxCostUSD = &o.maxCost
	}
	if o.given["acknowledge-egress"] || o.given["decisions"] {
		u.EgressAcknowledged = &o.ack
	}
	if o.given["model-cli"] {
		u.ModelCLIProvider = &o.cli
	}
	if o.given["lane"] {
		l := []string(o.lanes)
		u.ModelCLILanes = &l
	}
	return u
}

func writePolicyIfChanged(path string, out []byte, changed bool) error {
	if !changed {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

func initSummary(p policy.Policy, changed bool) map[string]any {
	d := p.Decisions()
	return map[string]any{
		"path": policy.Path, "changed": changed, "profile": p.Profile(),
		"decisions": map[string]any{"enabled": d.Enabled, "model": d.Model, "maxPerRun": d.MaxPerRun, "maxCostUsd": d.MaxCostUSD, "egressAcknowledged": d.EgressAcknowledged},
		"modelCli":  map[string]any{"provider": p.ModelCLI().Provider, "lanes": p.ModelCLI().Lanes},
		"next":      "commit .scout/policy.yaml, add the scout-guardian workflow, then run a full default-branch run to publish the baseline",
	}
}

type asker struct {
	r   *bufio.Reader
	w   io.Writer
	err error
}

func (a *asker) line(q string) string {
	if _, err := fmt.Fprintf(a.w, "%s: ", q); err != nil && a.err == nil {
		a.err = err
	}
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

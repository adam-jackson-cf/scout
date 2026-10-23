package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"scout/internal/baseline"
	"scout/internal/domain"
	"scout/internal/engine"
	"scout/internal/policy"
)

// diff gates a pull request: head issues are attributed against the merge
// base, policy read from the base ref decides which introduced issues block.
func diff(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("diff", flag.ContinueOnError)
	flags := common(set)
	var base, head, summary, actor, text, remote, branch string
	var fork, noEgress bool
	set.StringVar(&base, "base", "", "base ref (required)")
	set.StringVar(&head, "head", "HEAD", "head ref; must be the checked-out HEAD")
	set.StringVar(&summary, "summary", "", "write the Markdown summary (<= 60 lines) to this file")
	set.StringVar(&actor, "override-actor", "", "actor who requested the override")
	set.StringVar(&text, "override-text", "", "PR body or label text carrying `scout:allow <id> reason=... ref=...` or `scout-override`")
	set.StringVar(&remote, "baseline-remote", "", "git remote (name, path or URL) holding the scout/state baseline")
	set.StringVar(&branch, "baseline-branch", "", "baseline branch (default scout/state)")
	set.BoolVar(&fork, "fork", false, "fork pull request: Jev never runs, ranking is the deterministic baseline")
	set.BoolVar(&noEgress, "no-egress", false, "forbid the Jev call even when consented")
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	cfg := engine.DiffConfig{
		Config:  engine.Config{Root: flags.root, ClawpatchBinary: flags.binary, Provider: providerConfig(false)},
		BaseRef: base, HeadRef: head, OverrideActor: actor, OverrideText: text, Log: os.Stderr, JevKey: jevKey(), Fork: fork, NoEgress: noEgress,
	}
	if remote != "" {
		cfg.Baseline = func(ctx context.Context) (*engine.BaseState, error) {
			loaded, ok, err := baseline.LoadBase(ctx, baseline.Options{Repo: flags.root, Remote: remote, Branch: branch})
			if err != nil || !ok {
				return nil, err
			}
			return &engine.BaseState{Issues: loaded.Issues, Revision: loaded.Revision, ToolDigest: loaded.ToolDigest, PolicyDigest: loaded.PolicyDigest}, nil
		}
	}
	report, err := engine.Diff(ctx, cfg)
	if err != nil {
		return emit(flags, report.Gate.Status, report.Gate.ExitCode, report, err)
	}
	text = engine.Summary(report.Gate, report.Ranking, report.Issues, report.WouldBlock, report.PreExisting, report.Health)
	if summary != "" {
		if writeErr := os.WriteFile(summary, []byte(text), 0o644); writeErr != nil {
			report.Gate.Reasons = append(report.Gate.Reasons, "summary not written: "+writeErr.Error())
			report.Gate.ExitCode = engine.ExitCode(true, false, false)
			report.Gate.Status = engine.StatusOf(report.Gate.ExitCode)
			return emit(flags, report.Gate.Status, report.Gate.ExitCode, report, writeErr)
		}
	}
	if !flags.json {
		fmt.Fprint(os.Stdout, text)
		return report.Gate.ExitCode
	}
	return emit(flags, report.Gate.Status, report.Gate.ExitCode, report, nil)
}

// assessGate derives the gate block for a plain assessment: with no base there
// is no differential, so nothing can be introduced or gated; only an
// incomplete required lane changes the exit code.
func assessGate(result engine.Result) engine.GateOutput {
	return engine.Evaluate(engine.GateInput{
		Policy: policy.Default(), Head: result.Issues, Comparison: domain.ComparisonNone, Lanes: result.Lanes,
	})
}

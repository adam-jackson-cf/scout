package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"scout/internal/areas"
	"scout/internal/clawpatch"
	"scout/internal/domain"
	"scout/internal/engine"
	"scout/internal/state"
)

// mapCmd is `scout map`: the Clawpatch agent map runs only at onboarding (the
// `--event` check, attempted until one map has completed) and on manual
// `--refresh` (design §4).
func mapCmd(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("map", flag.ContinueOnError)
	flags := common(set)
	var force, defaultBranch, authorized bool
	var event string
	set.BoolVar(&force, "refresh", false, "force one agent map now")
	set.BoolVar(&authorized, "allow-egress", false, "authorize source egress for the agent mapper")
	set.StringVar(&event, "event", "", "run kind for the onboarding check: pr|push|nightly")
	set.BoolVar(&defaultBranch, "default-branch", false, "the push targets the default branch")
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	if !force && event == "" {
		return emit(flags, "unavailable", exitUnavailable, nil, errors.New("--event or --refresh is required"))
	}
	root, err := filepath.Abs(flags.root)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	flags.root = root
	cfg, err := engineConfig(flags, authorized)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	pol, err := headPolicy(root)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	store, err := openStore(flags)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	defer store.Close()
	client, err := clawpatch.New(root, clawpatch.Config{BinaryPath: cfg.ClawpatchBinary, EgressAuthorized: cfg.ClawpatchEgressAuthorized, Provider: cfg.ClawpatchProvider})
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	run := engine.RefreshRun{Event: domain.LaneTrigger(event), DefaultBranch: defaultBranch, Force: force}
	head, err := exec.CommandContext(ctx, "git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, fmt.Errorf("resolve HEAD for the area snapshot: %w", err))
	}
	revision := strings.TrimSpace(string(head))
	mapping := pol.Mapping()
	var reconciled *areas.Report
	refresher := &engine.Refresher{
		Store: store, Mapper: client, Now: time.Now,
		Reconcile: func(ctx context.Context, mapped clawpatch.MapResult) error {
			report, err := reconcileAreas(ctx, store, mapped, mapping.AreasMatchJaccard, revision)
			if err != nil {
				return err
			}
			reconciled = &report
			return nil
		},
	}
	outcome, err := refresher.Run(ctx, run)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	code := exitComplete
	if outcome.Status == engine.RefreshBlocked {
		code = exitPartial
	}
	result := struct {
		engine.RefreshOutcome
		Areas *areas.Report `json:"areas,omitempty"`
	}{outcome, reconciled}
	return emit(flags, string(outcome.Status), code, result, nil)
}

// reconcileAreas folds a successful agent refresh into the stored areas and
// returns what changed. Only the latest merged output decides heuristic
// ownership.
func reconcileAreas(ctx context.Context, store *state.Store, mapped clawpatch.MapResult, matchJaccard float64, revision string) (areas.Report, error) {
	prev, err := store.ListAreas(ctx)
	if err != nil {
		return areas.Report{}, err
	}
	active := mapped.Features // clawpatch returns only the latest map's (non-stale) records
	next, report := areas.Reconcile(prev, active, areas.HeuristicOwner(active), areas.Config{MatchJaccard: matchJaccard, Snapshot: revision})
	return report, store.ReplaceAreas(ctx, next)
}

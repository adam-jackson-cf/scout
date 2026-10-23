package main

import (
	"context"
	"errors"
	"flag"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"scout/internal/areas"
	"scout/internal/clawpatch"
	"scout/internal/engine"
	"scout/internal/state"
)

// mapCmd is `scout map`: it evaluates the agent-map refresh triggers for a run
// and performs at most one debounced refresh (design §4). `--refresh` forces one.
func mapCmd(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("map", flag.ContinueOnError)
	flags := common(set)
	var force, defaultBranch, authorized bool
	var event, base, head string
	set.BoolVar(&force, "refresh", false, "force one agent map refresh now")
	set.BoolVar(&authorized, "allow-egress", false, "authorize source egress for the agent mapper")
	set.StringVar(&event, "event", "", "run kind for trigger evaluation: pr|push|nightly")
	set.BoolVar(&defaultBranch, "default-branch", false, "the push targets the default branch")
	set.StringVar(&base, "base", "", "commit before the push (diff base)")
	set.StringVar(&head, "head", "HEAD", "pushed commit")
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
	features, _ := clawpatch.LoadFeatures(client.StateDir()) // absent before the first map
	run := engine.RefreshRun{Event: engine.RefreshEvent(event), DefaultBranch: defaultBranch, Force: force, Features: features}
	if out, gitErr := exec.CommandContext(ctx, "git", "-C", root, "rev-parse", head).Output(); gitErr == nil {
		run.Revision = strings.TrimSpace(string(out))
	}
	if event == string(engine.RefreshPush) && defaultBranch && !force {
		if base == "" {
			return emit(flags, "unavailable", exitUnavailable, nil, errors.New("push evaluation requires --base"))
		}
		if run.Changes, err = engine.GitChanges(ctx, root, base, head); err != nil {
			return emit(flags, "unavailable", exitUnavailable, nil, err)
		}
	}
	mapping := pol.Mapping()
	refresher := &engine.Refresher{
		Store: store, Mapper: client, Now: time.Now, Mapping: mapping,
		Reconcile: func(ctx context.Context, mapped clawpatch.MapResult) error {
			return reconcileAreas(ctx, store, mapped, mapping.AreasMatchJaccard, run.Revision)
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
	return emit(flags, string(outcome.Status), code, outcome, nil)
}

// reconcileAreas folds a successful agent refresh into the stored areas. Only
// the latest merged output decides heuristic ownership.
func reconcileAreas(ctx context.Context, store *state.Store, mapped clawpatch.MapResult, matchJaccard float64, revision string) error {
	prev, err := store.ListAreas(ctx)
	if err != nil {
		return err
	}
	active := mapped.Features // clawpatch returns only the latest map's (non-stale) records
	snapshot := revision
	if snapshot == "" {
		snapshot = "refresh-" + time.Now().UTC().Format(time.RFC3339)
	}
	next, _ := areas.Reconcile(prev, active, areas.HeuristicOwner(active), areas.Config{MatchJaccard: matchJaccard, Snapshot: snapshot})
	return store.ReplaceAreas(ctx, next)
}

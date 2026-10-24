package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"scout/internal/baseline"
	"scout/internal/state"
)

// exportCmd writes the deterministic state/v1 export of the local ledger.
func exportCmd(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("export", flag.ContinueOnError)
	flags := common(set)
	var out string
	set.StringVar(&out, "out", "", "directory that receives state/v1/*")
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	if out == "" {
		return emit(flags, "unavailable", exitUnavailable, nil, errors.New("--out is required"))
	}
	store, err := state.OpenExistingReadOnly(flags.root)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	defer store.Close()
	files, err := store.Export(ctx)
	if err == nil {
		err = state.WriteExport(out, files)
	}
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	return emit(flags, "complete", exitComplete, map[string]any{"out": out, "files": len(files)}, nil)
}

// importCmd rebuilds the ledger from a state/v1 export into a fresh store.
func importCmd(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("import", flag.ContinueOnError)
	flags := common(set)
	var from string
	set.StringVar(&from, "from", "", "directory containing state/v1/*")
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	if from == "" {
		return emit(flags, "unavailable", exitUnavailable, nil, errors.New("--from is required"))
	}
	files, err := state.ReadExport(from)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	return importFiles(ctx, flags, files)
}

func importFiles(ctx context.Context, flags *commonFlags, files map[string][]byte) int {
	store, err := state.Open(flags.root)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	defer store.Close()
	if err := store.Import(ctx, files); err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	return emit(flags, "complete", exitComplete, map[string]any{"files": len(files)}, nil)
}

// stateCmd dispatches `scout state publish|restore`.
func stateCmd(ctx context.Context, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: scout state <publish|restore> --remote <name> [flags]")
		return exitUnavailable
	}
	set := flag.NewFlagSet("state "+args[0], flag.ContinueOnError)
	flags := common(set)
	var remote, branch string
	var trusted bool
	set.StringVar(&remote, "remote", "", "git remote name, path or URL")
	set.StringVar(&branch, "branch", baseline.DefaultBranch, "state branch")
	switch args[0] {
	case "publish":
		set.BoolVar(&trusted, "trusted", false, "assert a trusted default-branch run; PR and fork runs are read-only")
	case "restore":
	default:
		fmt.Fprintf(os.Stderr, "unknown state command %q\n", args[0])
		return exitUnavailable
	}
	if !parse(set, args[1:], flags) {
		return exitUnavailable
	}
	if remote == "" {
		return emit(flags, "unavailable", exitUnavailable, nil, errors.New("--remote is required"))
	}
	opts := baseline.Options{Repo: flags.root, Remote: remote, Branch: branch, Trusted: trusted}
	if args[0] == "restore" {
		return restoreState(ctx, flags, opts)
	}
	return publishState(ctx, flags, opts)
}

func publishState(ctx context.Context, flags *commonFlags, opts baseline.Options) int {
	if !opts.Trusted {
		return emit(flags, "unavailable", exitUnavailable, nil, baseline.ErrUntrusted)
	}
	store, err := state.OpenExistingReadOnly(flags.root)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	local, err := store.Export(ctx)
	_ = store.Close()
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	result, err := baseline.Publish(ctx, opts, func(current baseline.Tree) (baseline.Tree, error) {
		return baseline.MergeState(current, baseline.Tree(local))
	})
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	return emit(flags, "complete", exitComplete, result, nil)
}

func restoreState(ctx context.Context, flags *commonFlags, opts baseline.Options) int {
	tree, tip, err := baseline.Fetch(ctx, opts)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	code := importFiles(ctx, flags, tree)
	if code == exitComplete && !flags.json {
		fmt.Fprintln(os.Stderr, "scout: restored state from", tip)
	}
	return code
}

package main

import (
	"context"
	"errors"
	"flag"

	"scout/internal/areas"
	"scout/internal/state"
)

func leadingAreaIDs(rest []string) ([]string, []string) {
	var ids []string
	for len(rest) > 0 && len(rest[0]) > 0 && rest[0][0] != '-' {
		ids = append(ids, rest[0])
		rest = rest[1:]
	}
	return ids, rest
}

func validAreasCommand(command string, ids []string, pending bool) error {
	count, known := map[string]int{"list": 0, "accept": 1, "merge": 2}[command]
	if !known || len(ids) != count || (command == "list" && !pending) {
		return errors.New("use areas list --pending, accept ID, or merge PENDING_ID TARGET_ID")
	}
	return nil
}

func onlyPendingAreas(all []areas.Area) []areas.Area {
	pending := make([]areas.Area, 0)
	for _, area := range all {
		if area.Status == areas.StatusPending {
			pending = append(pending, area)
		}
	}
	return pending
}

type areasRequest struct {
	command       string
	ids           []string
	actor, reason string
}

func runAreasCommand(ctx context.Context, store *state.Store, req areasRequest) (any, error) {
	command, ids, actor, reason := req.command, req.ids, req.actor, req.reason
	switch command {
	case "list":
		all, err := store.ListAreas(ctx)
		if err != nil {
			return nil, err
		}
		return onlyPendingAreas(all), nil
	case "accept":
		return store.AcceptArea(ctx, ids[0], actor, reason)
	default:
		return store.MergeArea(ctx, ids[0], ids[1], actor, reason)
	}
}

func areasCmd(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("areas", flag.ContinueOnError)
	flags := common(set)
	var pending, authorized bool
	var actor, reason string
	set.BoolVar(&pending, "pending", false, "list pending area proposals")
	set.BoolVar(&authorized, "authorize", false, "explicitly authorize area mutation")
	set.StringVar(&actor, "actor", "", "maintainer actor")
	set.StringVar(&reason, "reason", "", "audit reason")
	if len(args) == 0 {
		return emit(flags, "unavailable", exitUnavailable, nil, errors.New("areas requires list, accept or merge"))
	}
	command := args[0]
	ids, rest := leadingAreaIDs(args[1:])
	if !parse(set, rest, flags) {
		return exitUnavailable
	}
	if len(set.Args()) != 0 {
		return emit(flags, "unavailable", exitUnavailable, nil, errors.New("area IDs must precede flags"))
	}
	if err := validAreasCommand(command, ids, pending); err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	if command != "list" {
		var err error
		actor, reason, err = requireAuthorization(authorized, actor, reason)
		if err != nil {
			return emit(flags, "unavailable", exitUnavailable, nil, err)
		}
	}
	store, err := openStore(flags)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	defer closeStore(store)
	result, err := runAreasCommand(ctx, store, areasRequest{command: command, ids: ids, actor: actor, reason: reason})
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	return emit(flags, "complete", exitComplete, result, nil)
}

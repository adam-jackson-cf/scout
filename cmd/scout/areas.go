package main

import (
	"context"
	"errors"
	"flag"

	"scout/internal/areas"
)

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
	rest := args[1:]
	var ids []string
	for len(rest) > 0 && len(rest[0]) > 0 && rest[0][0] != '-' {
		ids = append(ids, rest[0])
		rest = rest[1:]
	}
	if !parse(set, rest, flags) {
		return exitUnavailable
	}
	if len(set.Args()) != 0 {
		return emit(flags, "unavailable", exitUnavailable, nil, errors.New("area IDs must precede flags"))
	}
	if (command == "list" && (len(ids) != 0 || !pending)) || (command == "accept" && len(ids) != 1) || (command == "merge" && len(ids) != 2) || (command != "list" && command != "accept" && command != "merge") {
		return emit(flags, "unavailable", exitUnavailable, nil, errors.New("use areas list --pending, accept ID, or merge PENDING_ID TARGET_ID"))
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
	defer store.Close()
	if command == "list" {
		all, err := store.ListAreas(ctx)
		if err != nil {
			return emit(flags, "unavailable", exitUnavailable, nil, err)
		}
		result := make([]areas.Area, 0)
		for _, a := range all {
			if a.Status == areas.StatusPending {
				result = append(result, a)
			}
		}
		return emit(flags, "complete", exitComplete, result, nil)
	}
	var result areas.Area
	if command == "accept" {
		result, err = store.AcceptArea(ctx, ids[0], actor, reason)
	} else {
		result, err = store.MergeArea(ctx, ids[0], ids[1], actor, reason)
	}
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	return emit(flags, "complete", exitComplete, result, nil)
}

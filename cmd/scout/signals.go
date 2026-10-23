package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"scout/internal/domain"
	"scout/internal/state"
)

// signalsCmd lists recorded descriptive signals (design §3.2, §8). By default
// it shows the most recent snapshot of each metric per scope.
func signalsCmd(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("signals", flag.ContinueOnError)
	flags := common(set)
	var filter state.SignalFilter
	var history bool
	set.StringVar(&filter.ScopeAnchor, "scope", "", "scope anchor: feature:<id> or a directory such as fastapi/dependencies")
	set.StringVar(&filter.MetricID, "metric", "", "metric, e.g. scout.cc or scout.cc@1")
	set.StringVar(&filter.Snapshot, "snapshot", "", "restrict to one snapshot key")
	set.BoolVar(&history, "history", false, "show every recorded snapshot, not only the latest")
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	filter.Latest = !history && filter.Snapshot == ""
	store, err := state.OpenExistingReadOnly(flags.root)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	defer store.Close()
	signals, err := store.ListSignals(ctx, filter)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	if signals == nil {
		signals = []domain.Signal{}
	}
	if flags.json {
		return emit(flags, "complete", exitComplete, map[string]any{"signals": signals}, nil)
	}
	for _, signal := range signals {
		fmt.Fprintln(os.Stdout, signalLine(signal))
	}
	return exitComplete
}

func signalLine(s domain.Signal) string {
	parts := []string{s.ID(), s.ScopeKind + ":" + s.ScopeAnchor}
	if s.Window != "" {
		parts = append(parts, "window="+s.Window)
	}
	if s.Missing != "" {
		parts = append(parts, string(s.Missing)+" ("+s.MissingReason+")")
	} else {
		parts = append(parts, fmt.Sprintf("value=%.6g", *s.Value), fmt.Sprintf("n=%.6g", *s.Numerator), fmt.Sprintf("d=%.6g", *s.Denominator))
	}
	parts = append(parts, "population="+shortDigest(s.PopulationDigest))
	return strings.Join(parts, "  ")
}

func shortDigest(digest string) string {
	if len(digest) > 12 {
		return digest[:12]
	}
	return digest
}

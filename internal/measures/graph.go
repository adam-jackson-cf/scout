package measures

import (
	"path"
	"sort"
	"strings"

	"scout/internal/domain"
	"scout/internal/importgraph"
)

const metricCycleMass = "scout.cycle_mass"

// graphConfigFiles are the files import resolution reads in addition to the
// eligible source population.
func graphConfigFile(p string) bool {
	base := path.Base(p)
	return base == "go.mod" || base == "go.work" || base == "package.json" || base == "jsconfig.json" ||
		(strings.HasPrefix(base, "tsconfig") && strings.HasSuffix(base, ".json"))
}

// cycleMass computes scout.cycle_mass@1 for every scope over the eligible
// population, or an `unmeasured` row per scope when the graph cannot be built.
func cycleMass(root string, in Input, eligible []domain.File, allFiles []domain.File, buildFailure string) []domain.Signal {
	eligibleSet := make(map[string]bool, len(eligible))
	for _, file := range eligible {
		eligibleSet[file.Path] = true
	}
	candidateByPath := make(map[string]domain.File, len(in.Candidates))
	for _, file := range in.Candidates {
		candidateByPath[file.Path] = file
	}
	var inputs []importgraph.ScopeInput
	for _, scope := range BuildScopes(in.Candidates, in.Features) {
		if len(scope.Paths) == 0 {
			continue
		}
		var population []domain.File
		for _, p := range scope.Paths {
			if eligibleSet[p] {
				population = append(population, candidateByPath[p])
			}
		}
		paths := make([]string, len(population))
		for i, file := range population {
			paths[i] = file.Path
		}
		inputs = append(inputs, importgraph.ScopeInput{Kind: scope.Kind, Anchor: scope.Anchor, Paths: paths, PopulationDigest: PopulationDigest(population)})
	}

	var signals []domain.Signal
	if buildFailure == "" {
		files := []string{}
		for _, file := range eligible {
			files = append(files, file.Path)
		}
		for _, file := range allFiles {
			if !file.Excluded && graphConfigFile(file.Path) {
				files = append(files, file.Path)
			}
		}
		sort.Strings(files)
		graph, err := importgraph.Build(importgraph.Options{Root: root, Files: files})
		if err == nil {
			signals = importgraph.Signals(graph, inputs, in.Snapshot)
		} else {
			buildFailure = "import graph extraction failed: " + err.Error()
		}
	}
	if signals == nil {
		for _, input := range inputs {
			signals = append(signals, domain.Signal{
				MetricID: metricCycleMass, Version: metricVersion, ScopeKind: input.Kind, ScopeAnchor: input.Anchor,
				Snapshot: in.Snapshot, PopulationDigest: input.PopulationDigest,
				ToolDigest: importgraph.ToolDigest, ConfigDigest: importgraph.ConfigDigest,
				Missing: domain.SignalUnmeasured, MissingReason: buildFailure, Claim: domain.ClaimDescriptive,
			})
		}
	}
	for i := range signals {
		if signals[i].Detail == nil {
			signals[i].Detail = map[string]any{}
		}
		signals[i].Detail["run_scope"] = in.RunScope
	}
	return signals
}

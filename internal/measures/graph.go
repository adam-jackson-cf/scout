package measures

import (
	"path"
	"sort"
	"strings"

	"scout/internal/domain"
	"scout/internal/importgraph"
)

// MetricCycleMass is the import-graph cycle metric id (formula version metricVersion).
const MetricCycleMass = "scout.cycle_mass"

// graphConfigFile reports files import resolution reads in addition to the
// eligible source population.
func graphConfigFile(p string) bool {
	base := path.Base(p)
	return base == "go.mod" || base == "go.work" || base == "package.json" || base == "jsconfig.json" ||
		(strings.HasPrefix(base, "tsconfig") && strings.HasSuffix(base, ".json"))
}

// cycleMass computes scout.cycle_mass@1 for every scope over the eligible
// population, or an `unmeasured` row per scope when the graph cannot be built.
func cycleMass(root string, in Input, eligible []domain.File, allFiles []domain.File, buildFailure string) []domain.Signal {
	inputs := cycleMassScopes(in, eligible)
	var signals []domain.Signal
	if buildFailure == "" {
		graph, err := importgraph.Build(importgraph.Options{Root: root, Files: graphFiles(eligible, allFiles)})
		if err == nil {
			signals = importgraph.Signals(graph, inputs, in.Snapshot)
		} else {
			buildFailure = "import graph extraction failed: " + err.Error()
		}
	}
	if signals == nil {
		signals = unmeasuredCycleMass(in, inputs, buildFailure)
	}
	setRunScope(signals, in.RunScope)
	return signals
}

func cycleMassScopes(in Input, eligible []domain.File) []importgraph.ScopeInput {
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
		inputs = append(inputs, importgraph.ScopeInput{Kind: scope.Kind, Anchor: scope.Anchor, Paths: filePaths(population), PopulationDigest: PopulationDigest(population)})
	}
	return inputs
}

func filePaths(files []domain.File) []string {
	paths := make([]string, len(files))
	for i, file := range files {
		paths[i] = file.Path
	}
	return paths
}

func graphFiles(eligible []domain.File, allFiles []domain.File) []string {
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
	return files
}

func unmeasuredCycleMass(in Input, inputs []importgraph.ScopeInput, reason string) []domain.Signal {
	var signals []domain.Signal
	for _, input := range inputs {
		signals = append(signals, domain.Signal{
			MetricID: MetricCycleMass, Version: metricVersion, ScopeKind: input.Kind, ScopeAnchor: input.Anchor,
			Snapshot: in.Snapshot, PopulationDigest: input.PopulationDigest,
			ToolDigest: importgraph.ToolDigest, ConfigDigest: importgraph.ConfigDigest,
			Missing: domain.SignalUnmeasured, MissingReason: reason, Claim: domain.ClaimDescriptive,
		})
	}
	return signals
}

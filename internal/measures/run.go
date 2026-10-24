package measures

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sort"

	"scout/internal/domain"
)

// Config selects what Run measures.
type Config struct {
	// Root is the repository root (the tree the Files were scanned from).
	Root     string
	Snapshot domain.Snapshot
	// RunScope names the inventory scope of the run (domain.ScopeKind), reported
	// in every signal's detail; a narrower scope yields a changed-files population.
	RunScope string
	Files    []domain.File
	Features []domain.Feature
	Tools    Tools
	// Coverage is the revision-matched coverage artifact, or nil when none was provided.
	Coverage *Coverage
}

// Result is the outcome of one measurement run.
type Result struct {
	Signals []domain.Signal
	// Measured is the measured population (production source, NLOC ≥ 10).
	Measured []domain.File
	// Lizard holds every candidate lizard measured, for consumers (hotspot
	// needs per-file CC and NLOC).
	Lizard map[string]FileMeasure
	// Candidates counts production-source files handed to the scanners.
	Candidates int
	// LizardErr and JscpdErr are non-empty when a scanner did not run; the
	// affected signals are `unmeasured`.
	LizardErr, JscpdErr string
}

// FileCC is the sum of a file's function CC, the file-level complexity used by
// hotspot (design §8.3).
func (m FileMeasure) FileCC() int {
	total := 0
	for _, fn := range m.Functions {
		total += fn.CC
	}
	return total
}

// Run measures the scopes of the given files with lizard and jscpd. A scanner
// failure never aborts the run: the affected signals are recorded
// `unmeasured` with the reason.
func Run(ctx context.Context, cfg Config) Result {
	var result Result
	candidates := Candidates(cfg.Files)
	result.Candidates = len(candidates)
	facts := Facts{}
	// The sandbox profile and path checks need the resolved root; an
	// unresolvable root leaves every measure unmeasured with that reason.
	root, rootErr := filepath.EvalSymlinks(cfg.Root)
	switch {
	case rootErr != nil:
		facts.LizardErr = "repository root is unavailable: " + rootErr.Error()
	case len(candidates) > 0:
		paths := make([]string, len(candidates))
		for i, file := range candidates {
			paths[i] = file.Path
		}
		sort.Strings(paths)
		lizard, err := cfg.Tools.RunLizard(ctx, root, paths)
		if err != nil {
			facts.LizardErr = err.Error()
		} else {
			facts.Lizard = lizard
			result.Lizard = lizard
		}
	default:
		facts.Lizard = map[string]FileMeasure{}
	}
	if facts.LizardErr == "" {
		result.Measured = MeasuredPopulation(candidates, facts.Lizard)
		if len(result.Measured) > 0 {
			paths := make([]string, len(result.Measured))
			for i, file := range result.Measured {
				paths[i] = file.Path
			}
			sort.Strings(paths)
			lines, err := lineCounts(root, paths)
			if err == nil {
				facts.Lines = lines
				facts.Clones, err = cfg.Tools.RunJscpd(ctx, root, paths)
			}
			if err != nil {
				facts.JscpdErr = err.Error()
			}
		}
	}
	result.LizardErr, result.JscpdErr = facts.LizardErr, facts.JscpdErr
	input := Input{
		Snapshot: SnapshotKey(cfg.Snapshot), RunScope: cfg.RunScope,
		Candidates: candidates, Features: cfg.Features, Facts: facts, Coverage: cfg.Coverage,
	}
	result.Signals = Compute(input)
	graphFailure := ""
	if facts.LizardErr != "" {
		graphFailure = "the eligible population is undefined because lizard did not run: " + facts.LizardErr
	}
	result.Signals = append(result.Signals, cycleMass(root, input, result.Measured, cfg.Files, graphFailure)...)
	return result
}

// lineCounts counts line positions as jscpd reads them: newline-separated
// lines, the last one counted when it has content.
func lineCounts(root string, paths []string) (map[string]int, error) {
	counts := make(map[string]int, len(paths))
	for _, path := range paths {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			return nil, err
		}
		lines := bytes.Count(data, []byte{'\n'})
		if len(data) > 0 && data[len(data)-1] != '\n' {
			lines++
		}
		counts[path] = lines
	}
	return counts, nil
}

package analyzers

import (
	"time"

	"scout/internal/domain"
)

// performanceSpecs covers Python micro-optimizations and explicit CSS height
// transitions. The CSS rule reports the source-level layout animation
// mechanism, never measured page timing, and deliberately excludes
// compositor-only and ambiguous `transition: all` declarations. Go, Rust, C#,
// JavaScript, TypeScript, and SQL remain uncovered until an adapter can supply
// a bounded invocation and parseable evidence for those inputs.

func performanceSpecs() []Spec {
	return []Spec{
		{
			ID:       "performance:ruff",
			Family:   domain.FamilyPerformance,
			Category: "micro-optimization",
			Binary: Binary{
				Path:        "ruff",
				VersionArgs: []string{"--version"},
				Version:     "0.15.8",
				Args: []string{
					"check",
					"--isolated",
					"--no-cache",
					"--select", "PERF",
					"--output-format", "sarif",
					"{input_paths}",
				},
			},
			Rules: Rules{Version: "ruff/PERF@0.15.8"},
			Eligibility: Eligibility{
				Languages:  []string{"Python"},
				Extensions: []string{".py", ".pyi"},
			},
			Inputs: InputLimits{
				MaxFiles:         10_000,
				MaxFileBytes:     2 << 20,
				MaxTotalBytes:    256 << 20,
				MaxManifestBytes: 4 << 20,
			},
			// Ruff's isolated --show-files listing must match all declared
			// inputs before its successful SARIF result can complete coverage.
			Attestation: AttestRuffFileListing,
			Output: Output{
				Format:        OutputSARIF,
				SchemaVersion: "2.1.0",
			},
			ExitCodes: map[int]ExitSemantics{
				0: ExitSuccess,
				1: ExitFindings,
				2: ExitFailure,
			},
			Timeout:   2 * time.Minute,
			MaxOutput: 32 << 20,
			Permissions: Permissions{
				ReadOnlySource: true,
				Network:        false,
			},
		},
		{
			ID:       "performance.semgrep.css-layout-height-transition",
			Family:   domain.FamilyPerformance,
			Category: "frontend-layout-animation",
			Binary: Binary{
				Path:        "semgrep",
				VersionArgs: []string{"--version"},
				Version:     "1.100.0",
				Args: []string{
					"scan",
					"--config", "{bundled_rule_files}",
					"--disable-version-check",
					"--metrics=off",
					"--no-git-ignore",
					"--json",
					"--error",
					"{input_paths}",
				},
			},
			Rules: Rules{
				Version:      "scout-semgrep-css-height-transition/v3@sha256:1181a0a5393c91fc63bbe72a8417e00093069179ed93a60eacd4aab5685fb5b8",
				BundledFiles: []string{semgrepCSSHeightTransitionRule},
				RuleIDs:      []string{"scout.performance.css.layout-height-transition"},
			},
			Eligibility: Eligibility{
				AllLanguages: true,
				Extensions:   []string{".css"},
			},
			Inputs: InputLimits{
				MaxFiles:         10_000,
				MaxFileBytes:     2 << 20,
				MaxTotalBytes:    256 << 20,
				MaxManifestBytes: 4 << 20,
			},
			Output: Output{
				Format:        OutputSemgrepJSON,
				SchemaVersion: "1.100.0",
			},
			ExitCodes: map[int]ExitSemantics{
				0:  ExitSuccess,
				1:  ExitFindings,
				2:  ExitFailure,
				3:  ExitFailure,
				4:  ExitFailure,
				5:  ExitFailure,
				7:  ExitFailure,
				8:  ExitFailure,
				13: ExitFailure,
				99: ExitFailure,
			},
			Timeout:   2 * time.Minute,
			MaxOutput: 32 << 20,
			Permissions: Permissions{
				ReadOnlySource: true,
				Network:        false,
			},
		},
	}
}

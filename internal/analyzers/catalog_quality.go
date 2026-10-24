package analyzers

import (
	"time"

	"scout/internal/domain"
)

// Ruff SARIF omits clean inputs; Runner accepts coverage only when a second
// isolated --show-files invocation lists exactly the declared scan inputs.
func qualitySpecs() []Spec {
	return []Spec{{
		ID:       "quality.ruff.complexity",
		Family:   domain.FamilyQuality,
		Category: "complexity",
		Binary: Binary{
			Path: "ruff", VersionArgs: []string{"--version"}, Version: "0.15.8",
			Args: []string{"check", "--isolated", "--no-cache", "--select", "C90", "--output-format", "sarif", "{input_paths}"},
		},
		Rules:       Rules{Version: "ruff/C90@0.15.8"},
		Eligibility: Eligibility{Languages: []string{"Python"}, Extensions: []string{".py", ".pyi"}},
		Inputs:      InputLimits{MaxFiles: 10_000, MaxFileBytes: 2 << 20, MaxTotalBytes: 256 << 20, MaxManifestBytes: 4 << 20},
		Output:      Output{Format: OutputSARIF, SchemaVersion: "2.1.0"},
		Attestation: AttestRuffFileListing,
		ExitCodes:   map[int]ExitSemantics{0: ExitSuccess, 1: ExitFindings, 2: ExitFailure},
		Timeout:     2 * time.Minute, MaxOutput: 32 << 20,
		Permissions: Permissions{ReadOnlySource: true},
	}}
}

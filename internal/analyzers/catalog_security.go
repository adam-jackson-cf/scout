package analyzers

import (
	"time"

	"scout/internal/domain"
)

// securitySpecs declares offline, evidence-checked rules for credential
// leakage, unsafe PyYAML deserialization, vulnerable direct multipart
// requirements, DOM-clobberable webpack public-path flows, and mutable GitHub
// Actions refs. The multipart rule reports only bounds that exclude its known
// fix; the JavaScript rule requires the unsafe auto-public-path runtime
// template and excludes the documented SCRIPT element guard. Bundled rules are
// materialized in private scanner state; target rule config and scripts are
// never invoked.
// Semgrep attests scanned paths. Gitleaks scans a private staged copy of each
// batch and attests per-file trace events plus the scanned byte total, with
// single-file fallback for unproven files. Other security mechanisms remain
// explicit contextual or
// deterministic coverage gaps.
func securitySpecs() []Spec {
	return []Spec{
		{
			ID:       "security.semgrep.python-unsafe-yaml-load",
			Family:   domain.FamilySecurity,
			Category: "dangerous-deserialization",
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
				Version:      "scout-semgrep-security/v1@sha256:5395708e0a5f4f593f58d62eb55b8f4a261aeafbdff42979b60bd6646d3df6bb",
				BundledFiles: []string{semgrepPythonUnsafeYAMLLoadRule},
				RuleIDs:      []string{"scout.security.python.yaml-unsafe-load"},
			},
			Eligibility: Eligibility{
				Languages:  []string{"python"},
				Extensions: []string{".py"},
			},
			Inputs: InputLimits{
				MaxFiles:         10_000,
				MaxFileBytes:     2 << 20,
				MaxTotalBytes:    256 << 20,
				MaxManifestBytes: 4 << 20,
			},
			// Semgrep JSON explicitly identifies scanned paths, including
			// clean files, and reports processing errors separately.
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
		{
			ID:       "security.semgrep.python-multipart-cve-2024-24762",
			Family:   domain.FamilySecurity,
			Category: "dependency-vulnerability",
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
				Version:      "scout-semgrep-python-multipart/v1@sha256:9c0a8465c628cda3fb102339cb85ced19343e6cabd567d2f9bc3ba154e5d745a",
				BundledFiles: []string{semgrepPythonMultipartRule},
				RuleIDs:      []string{"scout.security.python-multipart-cve-2024-24762"},
			},
			Eligibility: Eligibility{
				Languages:  []string{"toml"},
				Extensions: []string{".toml"},
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
		{
			ID:       "security.semgrep.javascript-current-script-dom-clobbering",
			Family:   domain.FamilySecurity,
			Category: "dom-clobbering",
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
				Version:      "scout-semgrep-current-script/v1@sha256:47d4c1818fa6b3a22c23d6fc1fc8ad5922dfe255a9a32780773e17851a43a457",
				BundledFiles: []string{semgrepJavaScriptCurrentScriptRule},
				RuleIDs:      []string{"scout.security.javascript.current-script-dom-clobbering"},
			},
			Eligibility: Eligibility{
				Languages:  []string{"javascript"},
				Extensions: []string{".js", ".mjs", ".cjs"},
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
		{
			ID:       "security.semgrep.actions-mutable-ref",
			Family:   domain.FamilySecurity,
			Category: "workflow-supply-chain",
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
				Version:      "scout-semgrep-actions/v1@sha256:25790a46d68968667f58f0be5fe9d9379858ba2b99017e89f8683723fa0d60fc",
				BundledFiles: []string{semgrepActionsMutableRefRule},
				RuleIDs:      []string{"scout.security.actions.mutable-ref"},
			},
			Eligibility: Eligibility{
				Languages:    []string{"yaml"},
				Extensions:   []string{".yaml", ".yml"},
				PathPrefixes: []string{".github/workflows/"},
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
		{
			ID:       "security.gitleaks.default-credentials",
			Family:   domain.FamilySecurity,
			Category: "credential-leakage",
			Binary: Binary{
				Path:           "gitleaks",
				VersionArgs:    []string{"version"},
				Version:        "8.30.1",
				PrivateWorkDir: true,
				Args: []string{
					"dir",
					"--no-banner",
					"--no-color",
					"--log-level", "trace",
					"--config", "{bundled_rule_files}",
					"--ignore-gitleaks-allow",
					"--gitleaks-ignore-path", "/dev/null",
					"--redact=100",
					"--max-target-megabytes", "8",
					"--report-format", "sarif",
					"--report-path", "-",
					"--exit-code", "1",
					"{staged_input_dir}",
				},
			},
			Rules: Rules{
				Version:      "gitleaks-default-rules/8.30.1@sha256:e163e53b9e7e8a8511e77271e2b323ed057759542a6d988258afe3a1fa329caf",
				BundledFiles: []string{gitleaksDefaultRule},
				RuleIDs:      gitleaksDefaultRuleIDs,
			},
			Eligibility: Eligibility{
				Languages: []string{
					"csharp", "go", "java", "javascript", "python", "rust", "shell", "typescript", "yaml", "json", "toml", "xml",
				},
				Extensions: []string{
					".cs", ".go", ".java", ".js", ".jsx", ".mjs", ".cjs", ".ts", ".tsx", ".mts", ".cts", ".py", ".pyi", ".rs", ".sh", ".yaml", ".yml", ".json", ".toml", ".xml",
				},
			},
			Inputs: InputLimits{
				MaxFiles:         200,
				MaxFileBytes:     8 * 1024 * 1024,
				MaxTotalBytes:    32 * 1024 * 1024,
				MaxManifestBytes: 256 * 1024,
			},
			Attestation: AttestGitleaksScanLog,
			Output: Output{
				Format:            OutputSARIF,
				SchemaVersion:     "2.1.0",
				SecretDiagnostics: true,
			},
			ExitCodes: map[int]ExitSemantics{
				0: ExitSuccess,
				1: ExitFindings,
			},
			Timeout:   2 * time.Minute,
			MaxOutput: 32 * 1024 * 1024,
			Permissions: Permissions{
				ReadOnlySource: true,
				Network:        false,
			},
		},
	}
}

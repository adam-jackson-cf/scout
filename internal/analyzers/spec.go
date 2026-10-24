// Package analyzers runs declared, read-only external analysis adapters.
package analyzers

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"scout/internal/domain"
)

const manifestVersion = "scout.input-manifest/v1"

type OutputFormat string

const (
	OutputJSON        OutputFormat = "json"
	OutputSARIF       OutputFormat = "sarif"
	OutputSemgrepJSON OutputFormat = "semgrep_json"
)

type ExitSemantics string

const (
	ExitSuccess  ExitSemantics = "success"
	ExitFindings ExitSemantics = "findings"
	ExitPartial  ExitSemantics = "partial"
	ExitFailure  ExitSemantics = "failure"
)

// Attestation selects the strategy that proves a scanner inspected an input.
// The runner dispatches on it; it never branches on a scanner's name.
type Attestation string

const (
	// AttestOutput: the structured output itself carries the proof (adapter
	// JSON processed inputs, Semgrep paths.scanned, a SARIF property).
	AttestOutput Attestation = ""
	// AttestGitleaksScanLog: the staged batch is proven per file from the
	// trace log's scanned-bytes total.
	AttestGitleaksScanLog Attestation = "gitleaks-scan-log"
	// AttestRuffFileListing: an exact `--show-files` listing of the inputs.
	AttestRuffFileListing Attestation = "ruff-file-listing"
)

func (s Spec) gitleaks() bool { return s.Attestation == AttestGitleaksScanLog }

// Binary declares the exact executable and invocation used by an adapter.
type Binary struct {
	Path           string   `json:"path"`
	VersionArgs    []string `json:"versionArgs"`
	Version        string   `json:"version"`
	Runtime        string   `json:"runtime,omitempty"`
	Args           []string `json:"args"`
	PrivateWorkDir bool     `json:"privateWorkDir,omitempty"`
}

// Rules identifies the immutable rule corpus and rule IDs an adapter may emit.
// BundledFiles are compiled into Scout and materialized outside the assessed
// repository for a scanner invocation.
type Rules struct {
	Version      string   `json:"version"`
	Files        []string `json:"files,omitempty"`
	ConfigFiles  []string `json:"configFiles,omitempty"`
	BundledFiles []string `json:"bundledFiles,omitempty"`
	RuleIDs      []string `json:"ruleIds,omitempty"`
}

// Eligibility makes an adapter's language and file applicability explicit.
type Eligibility struct {
	Languages    []string `json:"languages,omitempty"`
	Extensions   []string `json:"extensions,omitempty"`
	AllLanguages bool     `json:"allLanguages,omitempty"`
	PathPrefixes []string `json:"pathPrefixes,omitempty"`
}

// InputLimits bound scanner input and declare its project-level read footprint.
type InputLimits struct {
	ProjectFootprint []string `json:"projectFootprint,omitempty"`
	MaxFiles         int      `json:"maxFiles"`
	MaxFileBytes     int64    `json:"maxFileBytes"`
	MaxTotalBytes    int64    `json:"maxTotalBytes"`
	MaxManifestBytes int      `json:"maxManifestBytes"`
}

// Output declares a structured scanner output contract. SARIF without a
// processed-inputs property retains findings but cannot complete coverage;
// Semgrep JSON supplies its own scanned-path attestation.
type Output struct {
	Format                  OutputFormat `json:"format"`
	SchemaVersion           string       `json:"schemaVersion"`
	ProcessedInputsProperty string       `json:"processedInputsProperty,omitempty"`
	SecretDiagnostics       bool         `json:"secretDiagnostics,omitempty"`
}

// Permissions are deliberately restrictive: scanners may read declared source
// inputs but may not modify source/configuration or use the network.
type Permissions struct {
	ReadOnlySource bool `json:"readOnlySource"`
	Network        bool `json:"network"`
}

// Spec is the complete contract for one subprocess adapter. It never describes
// a project script: Binary must name the scanner executable itself.
type Spec struct {
	ID          string                `json:"id"`
	Family      domain.Family         `json:"family"`
	Category    string                `json:"category"`
	Binary      Binary                `json:"binary"`
	Rules       Rules                 `json:"rules"`
	Eligibility Eligibility           `json:"eligibility"`
	Inputs      InputLimits           `json:"inputs"`
	Output      Output                `json:"output"`
	ExitCodes   map[int]ExitSemantics `json:"exitCodes"`
	Timeout     time.Duration         `json:"timeout"`
	MaxOutput   int64                 `json:"maxOutput"`
	Permissions Permissions           `json:"permissions"`
	// Attestation says how the adapter proves it inspected every declared
	// input when its structured output cannot.
	Attestation Attestation `json:"attestation,omitempty"`
}

type InputManifest struct {
	Version          string          `json:"version"`
	Inputs           []ManifestInput `json:"inputs"`
	ProjectFootprint []string        `json:"projectFootprint,omitempty"`
	RuleFiles        []string        `json:"ruleFiles,omitempty"`
	ConfigFiles      []string        `json:"configFiles,omitempty"`
	BundledRules     []RuleDigest    `json:"bundledRules,omitempty"`
}

// RuleDigest binds a materialized bundled rule's logical path to its immutable
// compiled content.
type RuleDigest struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

type ManifestInput struct {
	Path     string `json:"path"`
	Language string `json:"language,omitempty"`
	Digest   string `json:"digest"`
	Bytes    int64  `json:"bytes"`
	Lines    int    `json:"lines"`
}

func (s Spec) Validate() error {
	if !safeName(s.ID) || !s.Family.Valid() || strings.TrimSpace(s.Category) == "" {
		return errors.New("adapter spec requires a safe ID, family, and category")
	}
	if strings.TrimSpace(s.Binary.Path) == "" || strings.TrimSpace(s.Binary.Version) == "" || len(s.Binary.VersionArgs) == 0 || len(s.Binary.Args) == 0 {
		return errors.New("adapter spec requires binary path, exact version, version arguments, and arguments")
	}
	if filepath.Base(s.Binary.Path) != s.Binary.Path || !safeName(s.Binary.Path) {
		return errors.New("adapter binary must be a bare scanner executable, not a project script")
	}
	for _, arg := range s.Binary.VersionArgs {
		if strings.ContainsRune(arg, '\x00') || strings.ContainsAny(arg, "{}") {
			return errors.New("adapter version arguments cannot contain placeholders or NUL")
		}
	}
	if strings.TrimSpace(s.Binary.Runtime) != "" && (strings.ContainsRune(s.Binary.Runtime, '\x00') || strings.ContainsAny(s.Binary.Runtime, "\r\n")) {
		return errors.New("adapter runtime declaration is unsafe")
	}
	manifest, paths, absolutePaths, stagedDir, bundledRules := false, false, false, false, false
	for _, arg := range append(append([]string{}, s.Binary.VersionArgs...), s.Binary.Args...) {
		if strings.ContainsRune(arg, '\x00') {
			return errors.New("adapter arguments cannot contain NUL")
		}
		manifest = manifest || arg == "{input_manifest}"
		paths = paths || arg == "{input_paths}"
		absolutePaths = absolutePaths || arg == "{absolute_input_paths}"
		stagedDir = stagedDir || arg == "{staged_input_dir}"
		bundledRules = bundledRules || arg == "{bundled_rule_files}"
	}
	if !manifest && !paths && !absolutePaths && !stagedDir {
		return errors.New("adapter arguments must declare input_manifest or input_paths")
	}
	switch s.Attestation {
	case AttestOutput:
	case AttestGitleaksScanLog, AttestRuffFileListing:
		if s.Output.Format != OutputSARIF {
			return errors.New("externally attested adapters must emit SARIF")
		}
	default:
		return errors.New("unknown adapter attestation")
	}
	if stagedDir != (s.Attestation == AttestGitleaksScanLog) {
		return errors.New("staged input directory and the Gitleaks attestation must agree")
	}
	if s.Binary.PrivateWorkDir && paths {
		return errors.New("private scanner workdir requires absolute input paths")
	}
	if strings.TrimSpace(s.Rules.Version) == "" {
		return errors.New("adapter spec requires an exact rule version")
	}
	if bundledRules != (len(s.Rules.BundledFiles) != 0) {
		return errors.New("adapter bundled rule arguments and files must agree")
	}
	ruleIDs := make(map[string]struct{}, len(s.Rules.RuleIDs))
	for _, ruleID := range s.Rules.RuleIDs {
		if !safeName(ruleID) {
			return errors.New("adapter rule ID must be safe")
		}
		if _, duplicate := ruleIDs[ruleID]; duplicate {
			return errors.New("adapter rule IDs must be unique")
		}
		ruleIDs[ruleID] = struct{}{}
	}
	bundledPaths := make(map[string]struct{}, len(s.Rules.BundledFiles))
	for _, path := range s.Rules.BundledFiles {
		if err := validRelativePath(path); err != nil {
			return fmt.Errorf("adapter bundled rule path: %w", err)
		}
		if _, duplicate := bundledPaths[path]; duplicate {
			return errors.New("adapter bundled rule paths must be unique")
		}
		digest, err := bundledRuleDigest(path)
		if err != nil || !strings.Contains(s.Rules.Version, "sha256:"+digest) {
			return errors.New("adapter rule version does not fingerprint bundled rule content")
		}
		bundledPaths[path] = struct{}{}
	}
	if len(s.Rules.BundledFiles) != 0 && len(s.Rules.RuleIDs) == 0 {
		return errors.New("adapter bundled rules require declared rule IDs")
	}
	if s.Eligibility.AllLanguages == (len(s.Eligibility.Languages) != 0) {
		return errors.New("eligibility must declare all languages or explicit languages")
	}
	for _, language := range s.Eligibility.Languages {
		if strings.TrimSpace(language) == "" {
			return errors.New("eligibility language cannot be empty")
		}
	}
	for _, extension := range s.Eligibility.Extensions {
		if !strings.HasPrefix(extension, ".") || strings.ContainsAny(extension, "/\\") {
			return fmt.Errorf("invalid eligibility extension %q", extension)
		}
	}
	for _, prefix := range s.Eligibility.PathPrefixes {
		if !strings.HasSuffix(prefix, "/") || validRelativePath(strings.TrimSuffix(prefix, "/")) != nil {
			return fmt.Errorf("invalid eligibility path prefix %q", prefix)
		}
	}
	if s.Inputs.MaxFiles <= 0 || s.Inputs.MaxFileBytes <= 0 || s.Inputs.MaxTotalBytes <= 0 || s.Inputs.MaxManifestBytes <= 0 || s.Timeout <= 0 || s.MaxOutput <= 0 {
		return errors.New("adapter spec requires positive input, timeout, and output bounds")
	}
	if s.Permissions.Network || !s.Permissions.ReadOnlySource {
		return errors.New("adapter specs must deny network and require read-only source")
	}
	for _, path := range append(append(append([]string{}, s.Rules.Files...), s.Rules.ConfigFiles...), s.Inputs.ProjectFootprint...) {
		if err := validRelativePath(path); err != nil {
			return fmt.Errorf("adapter path: %w", err)
		}
	}
	if s.Output.Format == OutputJSON && s.Output.SchemaVersion != "scout.adapter/v1" {
		return errors.New("JSON adapters require schema scout.adapter/v1")
	}
	if s.Output.Format == OutputSARIF && s.Output.SchemaVersion != "2.1.0" {
		return errors.New("SARIF adapters require schema 2.1.0")
	}
	if s.Output.Format == OutputSemgrepJSON && s.Output.SchemaVersion != s.Binary.Version {
		return errors.New("Semgrep JSON schema must match the pinned scanner version")
	}
	if s.Output.Format != OutputJSON && s.Output.Format != OutputSARIF && s.Output.Format != OutputSemgrepJSON {
		return errors.New("unsupported adapter output format")
	}
	if len(s.ExitCodes) == 0 {
		return errors.New("adapter spec requires versioned exit code semantics")
	}
	if _, ok := s.ExitCodes[0]; !ok {
		return errors.New("adapter exit semantics must declare exit code 0")
	}
	for code, meaning := range s.ExitCodes {
		if code < 0 || (meaning != ExitSuccess && meaning != ExitFindings && meaning != ExitPartial && meaning != ExitFailure) {
			return errors.New("adapter exit semantics are invalid")
		}
	}
	return nil
}

// Applies reports whether file is in this adapter contract's population.
func (s Spec) Applies(file domain.File) bool {
	if !file.Eligible || file.Excluded {
		return false
	}
	if len(s.Eligibility.PathPrefixes) != 0 {
		matched := false
		for _, prefix := range s.Eligibility.PathPrefixes {
			if strings.HasPrefix(file.Path, prefix) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if !s.Eligibility.AllLanguages {
		found := false
		for _, language := range s.Eligibility.Languages {
			if strings.EqualFold(language, file.Language) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(s.Eligibility.Extensions) == 0 {
		return true
	}
	for _, extension := range s.Eligibility.Extensions {
		if strings.EqualFold(extension, filepath.Ext(file.Path)) {
			return true
		}
	}
	return false
}

func (s Spec) allowsRuleID(ruleID string) bool {
	if len(s.Rules.RuleIDs) == 0 {
		return true
	}
	for _, declared := range s.Rules.RuleIDs {
		if ruleID == declared {
			return true
		}
	}
	return false
}

// Registry is safe for orchestration-time registration.
type Registry struct {
	mu         sync.RWMutex
	specs      map[string]Spec
	categories map[string]string
}

func NewRegistry() *Registry {
	return &Registry{specs: make(map[string]Spec), categories: make(map[string]string)}
}

// BuiltinRegistry contains only scanner contracts with verified offline
// invocations. Unsupported categories remain explicit gaps for contextual work.
func BuiltinRegistry() (*Registry, error) {
	registry := NewRegistry()
	for _, family := range [][]Spec{securitySpecs(), qualitySpecs(), performanceSpecs()} {
		for _, spec := range family {
			if err := registry.Register(spec); err != nil {
				return nil, err
			}
		}
	}
	return registry, nil
}

func (r *Registry) Register(spec Spec) error {
	if err := spec.Validate(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.specs[spec.ID]; exists {
		return fmt.Errorf("adapter %q is already registered", spec.ID)
	}
	categoryKey := string(spec.Family) + "\x00" + strings.ToLower(strings.TrimSpace(spec.Category))
	if existing, exists := r.categories[categoryKey]; exists {
		return fmt.Errorf("adapter %q duplicates family/category already registered by %q", spec.ID, existing)
	}
	r.specs[spec.ID] = spec
	r.categories[categoryKey] = spec.ID
	return nil
}

func (r *Registry) Runner(id, runID string) (*Runner, error) {
	r.mu.RLock()
	spec, found := r.specs[id]
	r.mu.RUnlock()
	if !found {
		return nil, fmt.Errorf("adapter %q is not registered", id)
	}
	return NewRunner(spec, runID)
}

func (r *Registry) Specs() []Spec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Spec, 0, len(r.specs))
	for _, spec := range r.specs {
		result = append(result, spec)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

var safeNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`)

func safeName(value string) bool { return safeNamePattern.MatchString(value) }

func validRelativePath(value string) error {
	if value == "" || strings.ContainsRune(value, '\x00') || strings.Contains(value, "\\") || filepath.IsAbs(value) || filepath.Clean(value) != value || value == "." || value == ".." || strings.HasPrefix(value, "../") {
		return errors.New("path must be normalized and relative")
	}
	return nil
}

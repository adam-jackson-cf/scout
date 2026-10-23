// Package domain contains the durable, transport-safe contracts shared by Scout.
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// Severity is retained even when a presentation threshold filters it.
type Severity string

const (
	SeverityInformational Severity = "informational"
	SeverityLow           Severity = "low"
	SeverityMedium        Severity = "medium"
	SeverityHigh          Severity = "high"
	SeverityCritical      Severity = "critical"
)

func (s Severity) Valid() bool {
	switch s {
	case SeverityInformational, SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical:
		return true
	default:
		return false
	}
}

// Rank orders severities from informational through critical.
func (s Severity) Rank() int {
	switch s {
	case SeverityInformational:
		return 0
	case SeverityLow:
		return 1
	case SeverityMedium:
		return 2
	case SeverityHigh:
		return 3
	case SeverityCritical:
		return 4
	default:
		return -1
	}
}

// Confidence expresses confidence in the finding, not coverage completeness.
type Confidence string

const (
	ConfidenceLow    Confidence = "low"
	ConfidenceMedium Confidence = "medium"
	ConfidenceHigh   Confidence = "high"
)

func (c Confidence) Valid() bool {
	return c == ConfidenceLow || c == ConfidenceMedium || c == ConfidenceHigh
}

type Family string

const (
	FamilyArchitecture Family = "architecture"
	FamilySecurity     Family = "security"
	FamilyQuality      Family = "quality"
	FamilyPerformance  Family = "performance"
	FamilyClawpatch    Family = "clawpatch"
)

func (f Family) Valid() bool {
	switch f {
	case FamilyArchitecture, FamilySecurity, FamilyQuality, FamilyPerformance, FamilyClawpatch:
		return true
	default:
		return false
	}
}

// Stage identifies a durable phase of an assessment job or failure.
type Stage string

const (
	StageInventory        Stage = "inventory"
	StageMapping          Stage = "mapping"
	StagePlanning         Stage = "planning"
	StageFileAnalysis     Stage = "file_analysis"
	StageFeatureAnalysis  Stage = "feature_analysis"
	StageClawpatchReview  Stage = "clawpatch_review"
	StageIssueCorrelation Stage = "issue_correlation"
	StageVerification     Stage = "verification"
)

func (s Stage) Valid() bool {
	switch s {
	case StageInventory, StageMapping, StagePlanning, StageFileAnalysis, StageFeatureAnalysis, StageClawpatchReview, StageIssueCorrelation, StageVerification:
		return true
	default:
		return false
	}
}

// CoverageStatus records the outcome of a declared unit and category.
type CoverageStatus string

const (
	CoverageInspected                  CoverageStatus = "inspected"
	CoverageReused                     CoverageStatus = "reused"
	CoverageNotApplicable              CoverageStatus = "not_applicable"
	CoverageSkipped                    CoverageStatus = "skipped"
	CoverageBlocked                    CoverageStatus = "blocked"
	CoverageIncomplete                 CoverageStatus = "incomplete"
	CoverageCovered                    CoverageStatus = "covered"
	CoverageFlagged                    CoverageStatus = "flagged"
	CoverageRequiresManualVerification CoverageStatus = "requires_manual_verification"
)

func (s CoverageStatus) Valid() bool {
	switch s {
	case CoverageInspected, CoverageReused, CoverageNotApplicable, CoverageSkipped, CoverageBlocked, CoverageIncomplete, CoverageCovered, CoverageFlagged, CoverageRequiresManualVerification:
		return true
	default:
		return false
	}
}

// Accounted reports whether this status can satisfy a completed run's ledger.
func (s CoverageStatus) Accounted() bool {
	switch s {
	case CoverageInspected, CoverageReused, CoverageNotApplicable, CoverageCovered, CoverageFlagged:
		return true
	default:
		return false
	}
}

// AdapterStatus preserves the actual status returned by a selected adapter.
type AdapterStatus string

const (
	AdapterCompleted         AdapterStatus = "completed"
	AdapterNoApplicableInput AdapterStatus = "no_applicable_input"
	AdapterMissingTool       AdapterStatus = "missing_tool"
	AdapterIneligible        AdapterStatus = "ineligible"
	AdapterFailed            AdapterStatus = "failed"
	AdapterTimedOut          AdapterStatus = "timed_out"
	AdapterPartial           AdapterStatus = "partial"
)

func (s AdapterStatus) Valid() bool {
	switch s {
	case AdapterCompleted, AdapterNoApplicableInput, AdapterMissingTool, AdapterIneligible, AdapterFailed, AdapterTimedOut, AdapterPartial:
		return true
	default:
		return false
	}
}

type RunState string

const (
	RunPlanned   RunState = "planned"
	RunRunning   RunState = "running"
	RunComplete  RunState = "complete"
	RunPartial   RunState = "partial"
	RunFailed    RunState = "failed"
	RunCancelled RunState = "cancelled"
)

func (s RunState) Valid() bool {
	switch s {
	case RunPlanned, RunRunning, RunComplete, RunPartial, RunFailed, RunCancelled:
		return true
	default:
		return false
	}
}

type JobState string

const (
	JobPending   JobState = "pending"
	JobRunning   JobState = "running"
	JobCompleted JobState = "completed"
	JobReused    JobState = "reused"
	JobSkipped   JobState = "skipped"
	JobBlocked   JobState = "blocked"
	JobFailed    JobState = "failed"
	JobCancelled JobState = "cancelled"
)

func (s JobState) Valid() bool {
	switch s {
	case JobPending, JobRunning, JobCompleted, JobReused, JobSkipped, JobBlocked, JobFailed, JobCancelled:
		return true
	default:
		return false
	}
}

type IssueState string

const (
	IssueOpen          IssueState = "open"
	IssueInvestigating IssueState = "investigating"
	IssueAccepted      IssueState = "accepted"
	IssueFalsePositive IssueState = "false_positive"
	IssueWontFix       IssueState = "wont_fix"
	IssueResolved      IssueState = "resolved"
	IssueRegressed     IssueState = "regressed"
)

func (s IssueState) Valid() bool {
	switch s {
	case IssueOpen, IssueInvestigating, IssueAccepted, IssueFalsePositive, IssueWontFix, IssueResolved, IssueRegressed:
		return true
	default:
		return false
	}
}

// File is the scoped index record; its feature links are Clawpatch IDs, never Scout-created IDs.
type File struct {
	Path              string   `json:"path"`
	Language          string   `json:"language,omitempty"`
	ContentDigest     string   `json:"contentDigest,omitempty"`
	PreviousPath      string   `json:"previousPath,omitempty"`
	Eligible          bool     `json:"eligible"`
	Excluded          bool     `json:"excluded"`
	ExclusionReason   string   `json:"exclusionReason,omitempty"`
	OwnedFeatureIDs   []string `json:"ownedFeatureIds,omitempty"`
	ContextFeatureIDs []string `json:"contextFeatureIds,omitempty"`
}

func (f File) Validate() error {
	if err := validateScopedPath(f.Path); err != nil {
		return fmt.Errorf("file path: %w", err)
	}
	if f.PreviousPath != "" {
		if err := validateScopedPath(f.PreviousPath); err != nil {
			return fmt.Errorf("file previous path: %w", err)
		}
	}
	if f.Excluded && f.ExclusionReason == "" {
		return errors.New("excluded file requires an exclusion reason")
	}
	if f.Eligible && f.Excluded {
		return errors.New("file cannot be both eligible and excluded")
	}
	return validateIDs(append(append([]string{}, f.OwnedFeatureIDs...), f.ContextFeatureIDs...), "feature ID")
}

type FeatureFileRef struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

func (r FeatureFileRef) Validate() error {
	if err := validateScopedPath(r.Path); err != nil {
		return fmt.Errorf("feature file path: %w", err)
	}
	if r.Reason == "" {
		return errors.New("feature file reference requires a reason")
	}
	return nil
}

type FeatureEntrypoint struct {
	Path    string  `json:"path"`
	Symbol  *string `json:"symbol"`
	Route   *string `json:"route"`
	Command *string `json:"command"`
}

func (e FeatureEntrypoint) Validate() error {
	return validateScopedPath(e.Path)
}

type FeatureTestRef struct {
	Path    string  `json:"path"`
	Command *string `json:"command"`
}

func (r FeatureTestRef) Validate() error {
	return validateScopedPath(r.Path)
}

type FeatureLock struct {
	LockedByRunID string    `json:"lockedByRunId"`
	LockedAt      time.Time `json:"lockedAt"`
	Hostname      string    `json:"hostname"`
	PID           int       `json:"pid"`
}

type AnalysisHistoryEntry struct {
	RunID     string    `json:"runId"`
	Kind      string    `json:"kind"`
	Summary   string    `json:"summary"`
	Provider  *string   `json:"provider"`
	Model     *string   `json:"model"`
	CreatedAt time.Time `json:"createdAt"`
}

// Feature preserves the canonical Clawpatch record and its unmodified source JSON.
type Feature struct {
	SchemaVersion   int                    `json:"schemaVersion"`
	FeatureID       string                 `json:"featureId"`
	Title           string                 `json:"title"`
	Summary         string                 `json:"summary"`
	Kind            string                 `json:"kind"`
	Source          string                 `json:"source"`
	Confidence      Confidence             `json:"confidence"`
	Entrypoints     []FeatureEntrypoint    `json:"entrypoints"`
	OwnedFiles      []FeatureFileRef       `json:"ownedFiles"`
	ContextFiles    []FeatureFileRef       `json:"contextFiles"`
	Tests           []FeatureTestRef       `json:"tests"`
	Tags            []string               `json:"tags"`
	TrustBoundaries []string               `json:"trustBoundaries"`
	Status          string                 `json:"status"`
	Lock            *FeatureLock           `json:"lock"`
	FindingIDs      []string               `json:"findingIds"`
	PatchAttemptIDs []string               `json:"patchAttemptIds"`
	AnalysisHistory []AnalysisHistoryEntry `json:"analysisHistory"`
	CreatedAt       time.Time              `json:"createdAt"`
	UpdatedAt       time.Time              `json:"updatedAt"`
	RawOriginal     json.RawMessage        `json:"rawOriginal"`
}

func (f Feature) Validate() error {
	if f.SchemaVersion != 1 {
		return fmt.Errorf("unsupported Clawpatch feature schema version %d", f.SchemaVersion)
	}
	if f.FeatureID == "" || f.Kind == "" || f.Source == "" || !f.Confidence.Valid() {
		return errors.New("feature requires canonical ID, kind, source, and confidence")
	}
	if len(f.RawOriginal) == 0 || !json.Valid(f.RawOriginal) {
		return errors.New("feature requires valid raw canonical Clawpatch JSON")
	}
	var original struct {
		SchemaVersion int    `json:"schemaVersion"`
		FeatureID     string `json:"featureId"`
		Kind          string `json:"kind"`
		Source        string `json:"source"`
	}
	if err := json.Unmarshal(f.RawOriginal, &original); err != nil {
		return fmt.Errorf("feature raw canonical JSON: %w", err)
	}
	if original.SchemaVersion != f.SchemaVersion || original.FeatureID != f.FeatureID || original.Kind != f.Kind || original.Source != f.Source {
		return errors.New("feature fields must retain the raw canonical Clawpatch identity")
	}
	for _, entrypoint := range f.Entrypoints {
		if err := entrypoint.Validate(); err != nil {
			return fmt.Errorf("feature entrypoint: %w", err)
		}
	}
	for _, ref := range append(append([]FeatureFileRef{}, f.OwnedFiles...), f.ContextFiles...) {
		if err := ref.Validate(); err != nil {
			return err
		}
	}
	for _, test := range f.Tests {
		if err := test.Validate(); err != nil {
			return fmt.Errorf("feature test: %w", err)
		}
	}
	return nil
}

// SourceRange identifies inspected source without storing source file contents.
type SourceRange struct {
	Path      string `json:"path"`
	StartLine int    `json:"startLine,omitempty"`
	EndLine   int    `json:"endLine,omitempty"`
	StartByte int    `json:"startByte,omitempty"`
	EndByte   int    `json:"endByte,omitempty"`
	Symbol    string `json:"symbol,omitempty"`
}

func (r SourceRange) Validate() error {
	if err := validateScopedPath(r.Path); err != nil {
		return fmt.Errorf("source range path: %w", err)
	}
	if (r.StartLine == 0) != (r.EndLine == 0) || r.StartLine < 0 || r.EndLine < 0 || r.StartLine > r.EndLine {
		return errors.New("source range has invalid line bounds")
	}
	if (r.StartByte == 0) != (r.EndByte == 0) || r.StartByte < 0 || r.EndByte < 0 || r.StartByte > r.EndByte {
		return errors.New("source range has invalid byte bounds")
	}
	if r.StartLine == 0 && r.StartByte == 0 && r.Symbol == "" {
		return errors.New("source range requires a line range, byte range, or symbol")
	}
	return nil
}

// SafeArtifactRef must name a sanitized artifact, never raw tool output.
type SafeArtifactRef struct {
	ArtifactID  string `json:"artifactId"`
	Digest      string `json:"digest"`
	ContentType string `json:"contentType,omitempty"`
}

func (r SafeArtifactRef) Validate() error {
	if r.ArtifactID == "" || r.Digest == "" {
		return errors.New("safe artifact reference requires an ID and digest")
	}
	return nil
}

// Evidence is sanitized provenance; RedactedQuote must never contain raw credential material.
type Evidence struct {
	Location      SourceRange      `json:"location"`
	RedactedQuote string           `json:"redactedQuote,omitempty"`
	Artifact      *SafeArtifactRef `json:"artifact,omitempty"`
}

func (e Evidence) Validate() error {
	if err := e.Location.Validate(); err != nil {
		return err
	}
	if e.Artifact != nil {
		return e.Artifact.Validate()
	}
	return nil
}

// Observation is immutable, sanitized evidence from one analyzer or review run.
type Observation struct {
	ObservationID    string `json:"observationId"`
	RunID            string `json:"runId"`
	Source           string `json:"source"`
	SourceFindingID  string `json:"sourceFindingId,omitempty"`
	Family           Family `json:"family"`
	OriginalCategory string `json:"originalCategory"`
	RuleID           string `json:"ruleId,omitempty"`
	RuleVersion      string `json:"ruleVersion,omitempty"`
	// Claim is set only on observations produced by a named promotion rule
	// (design §3.2); an ordinary scanner finding leaves it empty.
	Claim             SignalClaim       `json:"claim,omitempty"`
	ModelVersion      string            `json:"modelVersion,omitempty"`
	SourceSeverity    Severity          `json:"sourceSeverity"`
	EffectiveSeverity Severity          `json:"effectiveSeverity"`
	Confidence        Confidence        `json:"confidence"`
	Mechanism         string            `json:"mechanism"`
	Uncertainty       string            `json:"uncertainty,omitempty"`
	Evidence          []Evidence        `json:"evidence"`
	ArtifactRefs      []SafeArtifactRef `json:"artifactRefs,omitempty"`
	Validated         bool              `json:"validated"`
	ObservedAt        time.Time         `json:"observedAt"`
}

func (o Observation) Validate() error {
	if o.ObservationID == "" || o.RunID == "" || o.Source == "" || o.OriginalCategory == "" || o.Mechanism == "" {
		return errors.New("observation requires identity, source, category, and mechanism")
	}
	if !o.Family.Valid() || !o.SourceSeverity.Valid() || !o.EffectiveSeverity.Valid() || !o.Confidence.Valid() || !o.Claim.ValidIssueClaim() {
		return errors.New("observation has invalid family, severity, or confidence")
	}
	if o.EffectiveSeverity.Rank() > o.SourceSeverity.Rank() {
		return errors.New("effective severity cannot exceed immutable source severity")
	}
	if len(o.Evidence) == 0 {
		return errors.New("observation requires safe evidence")
	}
	for _, evidence := range o.Evidence {
		if err := evidence.Validate(); err != nil {
			return fmt.Errorf("observation evidence: %w", err)
		}
	}
	for _, artifact := range o.ArtifactRefs {
		if err := artifact.Validate(); err != nil {
			return fmt.Errorf("observation artifact: %w", err)
		}
	}
	if !o.Validated {
		return errors.New("observation evidence must be validated before persistence")
	}
	return nil
}

// CoverageKey is the unique file or canonical-feature category ledger key.
type CoverageKey struct {
	FilePath  string `json:"filePath,omitempty"`
	FeatureID string `json:"featureId,omitempty"`
	Family    Family `json:"family"`
	Category  string `json:"category"`
}

func (k CoverageKey) Validate() error {
	if (k.FilePath == "") == (k.FeatureID == "") {
		return errors.New("coverage key requires exactly one file path or feature ID")
	}
	if k.FilePath != "" {
		if err := validateScopedPath(k.FilePath); err != nil {
			return fmt.Errorf("coverage key: %w", err)
		}
	}
	if !k.Family.Valid() || k.Category == "" {
		return errors.New("coverage key requires family and category")
	}
	return nil
}

func (k CoverageKey) String() string {
	return strings.Join([]string{k.FilePath, k.FeatureID, string(k.Family), k.Category}, "\x00")
}

// Coverage makes execution, reuse, omissions, and uninspectable inputs explicit.
type Coverage struct {
	Key              CoverageKey    `json:"key"`
	Status           CoverageStatus `json:"status"`
	Reason           string         `json:"reason,omitempty"`
	InspectedRanges  []SourceRange  `json:"inspectedRanges,omitempty"`
	ReusedFromRunID  string         `json:"reusedFromRunId,omitempty"`
	AdapterResultIDs []string       `json:"adapterResultIds,omitempty"`
	ObservationIDs   []string       `json:"observationIds,omitempty"`
	OmittedPaths     []string       `json:"omittedPaths,omitempty"`
}

func (c Coverage) Validate() error {
	if err := c.Key.Validate(); err != nil {
		return err
	}
	if !c.Status.Valid() {
		return errors.New("coverage has invalid status")
	}
	if (c.Status == CoverageReused) != (c.ReusedFromRunID != "") {
		return errors.New("reused coverage requires, and only reuse may have, a baseline run ID")
	}
	if (c.Status == CoverageInspected || c.Status == CoverageCovered || c.Status == CoverageFlagged) && len(c.InspectedRanges) == 0 {
		return errors.New("inspected coverage requires inspected ranges")
	}
	if (c.Status == CoverageSkipped || c.Status == CoverageBlocked || c.Status == CoverageIncomplete || c.Status == CoverageRequiresManualVerification) && c.Reason == "" {
		return errors.New("incomplete coverage requires a reason")
	}
	for _, sourceRange := range c.InspectedRanges {
		if err := sourceRange.Validate(); err != nil {
			return fmt.Errorf("coverage inspected range: %w", err)
		}
	}
	for _, path := range c.OmittedPaths {
		if err := validateScopedPath(path); err != nil {
			return fmt.Errorf("coverage omitted path: %w", err)
		}
	}
	return nil
}

type Snapshot struct {
	RepositoryID      string `json:"repositoryId"`
	Revision          string `json:"revision,omitempty"`
	WorktreeDigest    string `json:"worktreeDigest"`
	Dirty             bool   `json:"dirty"`
	IgnoreRulesDigest string `json:"ignoreRulesDigest"`
	ExclusionsDigest  string `json:"exclusionsDigest"`
}

func (s Snapshot) Validate() error {
	if s.RepositoryID == "" || s.WorktreeDigest == "" || s.IgnoreRulesDigest == "" || s.ExclusionsDigest == "" {
		return errors.New("snapshot requires repository, worktree, ignore, and exclusion identities")
	}
	return nil
}

type ScopeKind string

const (
	ScopeRepository  ScopeKind = "repository"
	ScopeSubpath     ScopeKind = "subpath"
	ScopeStaged      ScopeKind = "staged"
	ScopeCommitRange ScopeKind = "commit_range"
	ScopePaths       ScopeKind = "paths"
)

func (k ScopeKind) Valid() bool {
	switch k {
	case ScopeRepository, ScopeSubpath, ScopeStaged, ScopeCommitRange, ScopePaths:
		return true
	default:
		return false
	}
}

type Scope struct {
	Kind             ScopeKind `json:"kind"`
	Paths            []string  `json:"paths,omitempty"`
	FromRevision     string    `json:"fromRevision,omitempty"`
	ToRevision       string    `json:"toRevision,omitempty"`
	SelectedFamilies []Family  `json:"selectedFamilies"`
}

func (s Scope) Validate() error {
	if !s.Kind.Valid() {
		return errors.New("scope has invalid kind")
	}
	if (s.Kind == ScopeSubpath || s.Kind == ScopePaths) && len(s.Paths) == 0 {
		return errors.New("path scope requires paths")
	}
	if s.Kind == ScopeCommitRange && (s.FromRevision == "" || s.ToRevision == "") {
		return errors.New("commit range scope requires both revisions")
	}
	for _, path := range s.Paths {
		if err := validateScopedPath(path); err != nil {
			return fmt.Errorf("scope path: %w", err)
		}
	}
	for _, family := range s.SelectedFamilies {
		if !family.Valid() || family == FamilyClawpatch {
			return errors.New("scope has invalid selected family")
		}
	}
	return nil
}

type Job struct {
	JobID              string       `json:"jobId"`
	Stage              Stage        `json:"stage"`
	CoverageKey        *CoverageKey `json:"coverageKey,omitempty"`
	InputFingerprint   string       `json:"inputFingerprint"`
	State              JobState     `json:"state"`
	InvalidationReason string       `json:"invalidationReason,omitempty"`
	CacheKey           string       `json:"cacheKey,omitempty"`
}

func (j Job) Validate() error {
	if j.JobID == "" || !j.Stage.Valid() || j.InputFingerprint == "" || !j.State.Valid() {
		return errors.New("job requires identity, stage, input fingerprint, and state")
	}
	if j.CoverageKey != nil {
		if err := j.CoverageKey.Validate(); err != nil {
			return fmt.Errorf("job coverage key: %w", err)
		}
	}
	return nil
}

// Plan is persisted before jobs execute so a crash can resume unfinished work.
type Plan struct {
	PlanID       string    `json:"planId"`
	RunID        string    `json:"runId"`
	Snapshot     Snapshot  `json:"snapshot"`
	Scope        Scope     `json:"scope"`
	Jobs         []Job     `json:"jobs"`
	PolicyDigest string    `json:"policyDigest,omitempty"`
	ToolDigest   string    `json:"toolDigest,omitempty"`
	ModelDigest  string    `json:"modelDigest,omitempty"`
	MapperDigest string    `json:"mapperDigest,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (p Plan) Validate() error {
	if p.PlanID == "" || p.RunID == "" {
		return errors.New("plan requires ID and run ID")
	}
	if err := p.Snapshot.Validate(); err != nil {
		return fmt.Errorf("plan snapshot: %w", err)
	}
	if err := p.Scope.Validate(); err != nil {
		return fmt.Errorf("plan scope: %w", err)
	}
	ids := make(map[string]struct{}, len(p.Jobs))
	for _, job := range p.Jobs {
		if err := job.Validate(); err != nil {
			return fmt.Errorf("plan job: %w", err)
		}
		if _, exists := ids[job.JobID]; exists {
			return fmt.Errorf("plan repeats job ID %q", job.JobID)
		}
		ids[job.JobID] = struct{}{}
	}
	return nil
}

type Failure struct {
	Stage     Stage  `json:"stage"`
	Code      string `json:"code,omitempty"`
	Message   string `json:"message"`
	JobID     string `json:"jobId,omitempty"`
	Retryable bool   `json:"retryable"`
}

func (f Failure) Validate() error {
	if !f.Stage.Valid() || f.Message == "" {
		return errors.New("failure requires stage and message")
	}
	return nil
}

type Omission struct {
	CoverageKey CoverageKey `json:"coverageKey"`
	Reason      string      `json:"reason"`
}

func (o Omission) Validate() error {
	if err := o.CoverageKey.Validate(); err != nil {
		return err
	}
	if o.Reason == "" {
		return errors.New("omission requires a reason")
	}
	return nil
}

// Run is an immutable assessment snapshot and its complete-or-partial coverage result.
type Run struct {
	RunID            string            `json:"runId"`
	PlanID           string            `json:"planId"`
	Snapshot         Snapshot          `json:"snapshot"`
	Scope            Scope             `json:"scope"`
	State            RunState          `json:"state"`
	StartedAt        time.Time         `json:"startedAt"`
	FinishedAt       *time.Time        `json:"finishedAt,omitempty"`
	ExpectedCoverage []CoverageKey     `json:"expectedCoverage"`
	Coverage         []Coverage        `json:"coverage"`
	Failures         []Failure         `json:"failures,omitempty"`
	Omissions        []Omission        `json:"omissions,omitempty"`
	ArtifactRefs     []SafeArtifactRef `json:"artifactRefs,omitempty"`
}

func (r Run) Validate() error {
	if r.RunID == "" || r.PlanID == "" || !r.State.Valid() {
		return errors.New("run requires IDs and state")
	}
	if err := r.Snapshot.Validate(); err != nil {
		return fmt.Errorf("run snapshot: %w", err)
	}
	if err := r.Scope.Validate(); err != nil {
		return fmt.Errorf("run scope: %w", err)
	}
	expected := make(map[string]struct{}, len(r.ExpectedCoverage))
	for _, key := range r.ExpectedCoverage {
		if err := key.Validate(); err != nil {
			return fmt.Errorf("expected coverage: %w", err)
		}
		encoded := key.String()
		if _, exists := expected[encoded]; exists {
			return fmt.Errorf("run repeats expected coverage %q", encoded)
		}
		expected[encoded] = struct{}{}
	}
	coverage := make(map[string]Coverage, len(r.Coverage))
	for _, row := range r.Coverage {
		if err := row.Validate(); err != nil {
			return fmt.Errorf("run coverage: %w", err)
		}
		encoded := row.Key.String()
		if _, exists := coverage[encoded]; exists {
			return fmt.Errorf("run repeats coverage %q", encoded)
		}
		coverage[encoded] = row
	}
	for _, failure := range r.Failures {
		if err := failure.Validate(); err != nil {
			return fmt.Errorf("run failure: %w", err)
		}
	}
	for _, omission := range r.Omissions {
		if err := omission.Validate(); err != nil {
			return fmt.Errorf("run omission: %w", err)
		}
	}
	for _, artifact := range r.ArtifactRefs {
		if err := artifact.Validate(); err != nil {
			return fmt.Errorf("run artifact: %w", err)
		}
	}
	if r.State == RunComplete {
		if r.FinishedAt == nil {
			return errors.New("complete run requires a finish time")
		}
		for encoded := range expected {
			row, found := coverage[encoded]
			if !found || !row.Status.Accounted() {
				return fmt.Errorf("complete run has unaccounted coverage %q", encoded)
			}
		}
		if len(r.Failures) != 0 || len(r.Omissions) != 0 {
			return errors.New("complete run cannot retain failures or omissions")
		}
	}
	return nil
}

// AdapterResult contains structured, sanitized adapter output and per-input coverage.
type AdapterResult struct {
	ResultID       string            `json:"resultId"`
	RunID          string            `json:"runId"`
	Adapter        string            `json:"adapter"`
	AdapterVersion string            `json:"adapterVersion"`
	RuleVersion    string            `json:"ruleVersion,omitempty"`
	Status         AdapterStatus     `json:"status"`
	Command        string            `json:"command,omitempty"`
	StartedAt      time.Time         `json:"startedAt"`
	FinishedAt     *time.Time        `json:"finishedAt,omitempty"`
	Coverage       []Coverage        `json:"coverage"`
	Observations   []string          `json:"observationIds,omitempty"`
	Failures       []Failure         `json:"failures,omitempty"`
	ArtifactRefs   []SafeArtifactRef `json:"artifactRefs,omitempty"`
}

func (r AdapterResult) Validate() error {
	if r.ResultID == "" || r.RunID == "" || r.Adapter == "" || r.AdapterVersion == "" || !r.Status.Valid() {
		return errors.New("adapter result requires identity, adapter version, and status")
	}
	if r.Status == AdapterCompleted && len(r.Failures) != 0 {
		return errors.New("completed adapter result cannot retain failures")
	}
	for _, coverage := range r.Coverage {
		if err := coverage.Validate(); err != nil {
			return fmt.Errorf("adapter coverage: %w", err)
		}
	}
	for _, failure := range r.Failures {
		if err := failure.Validate(); err != nil {
			return fmt.Errorf("adapter failure: %w", err)
		}
	}
	for _, artifact := range r.ArtifactRefs {
		if err := artifact.Validate(); err != nil {
			return fmt.Errorf("adapter artifact: %w", err)
		}
	}
	return nil
}

// Issue history events recorded in IssueHistoryEntry.Event.
const (
	EventCreated       = "created"
	EventMatched       = "matched"
	EventEvidenceDrift = "evidence_drift"
	EventAmbiguous     = "ambiguous_with"
	EventRegressed     = "regressed"
	EventResolved      = "resolved"
	EventMigrated      = "migrated"
	EventTriaged       = "triaged"
)

type IssueHistoryEntry struct {
	At             time.Time         `json:"at"`
	From           IssueState        `json:"from,omitempty"`
	To             IssueState        `json:"to"`
	Event          string            `json:"event,omitempty"`
	Actor          string            `json:"actor,omitempty"`
	Reason         string            `json:"reason"`
	ObservationIDs []string          `json:"observationIds,omitempty"`
	ArtifactRefs   []SafeArtifactRef `json:"artifactRefs,omitempty"`
}

func (h IssueHistoryEntry) Validate() error {
	if !h.To.Valid() || h.Reason == "" {
		return errors.New("issue history requires destination state and reason")
	}
	if h.From != "" && !h.From.Valid() {
		return errors.New("issue history has invalid prior state")
	}
	for _, artifact := range h.ArtifactRefs {
		if err := artifact.Validate(); err != nil {
			return fmt.Errorf("issue history artifact: %w", err)
		}
	}
	return nil
}

// ResolutionFileDeleted is the Verification.Reason for an issue whose file the
// inventory diff of a complete run shows deleted (not renamed).
const ResolutionFileDeleted = "file_deleted"

// Verification records the relevant successful re-check that resolved an issue.
type Verification struct {
	RunID             string            `json:"runId"`
	Reason            string            `json:"reason,omitempty"`
	Relevant          bool              `json:"relevant"`
	InspectionStatus  CoverageStatus    `json:"inspectionStatus,omitempty"`
	FailureModeAbsent bool              `json:"failureModeAbsent"`
	ArtifactRefs      []SafeArtifactRef `json:"artifactRefs,omitempty"`
	VerifiedAt        time.Time         `json:"verifiedAt"`
}

func (v Verification) ValidateResolution() error {
	if v.RunID == "" || !v.Relevant || !v.FailureModeAbsent {
		return errors.New("resolution requires a relevant recheck proving the failure mode absent")
	}
	if v.Reason == ResolutionFileDeleted {
		return nil
	}
	if v.InspectionStatus != CoverageInspected && v.InspectionStatus != CoverageCovered && v.InspectionStatus != CoverageFlagged {
		return errors.New("resolution requires successful relevant inspection, not missing or reused coverage")
	}
	for _, artifact := range v.ArtifactRefs {
		if err := artifact.Validate(); err != nil {
			return fmt.Errorf("verification artifact: %w", err)
		}
	}
	return nil
}

// IssueAnchor is the identity-bearing evidence of an issue (design §3.3).
// Line and byte ranges are locators on observations and never appear here.
type IssueAnchor struct {
	// Path is the rename-followed repository path. Empty for a digest-matched secret.
	Path string `json:"path,omitempty"`
	// Symbol is the qualified enclosing symbol when the analyzer supplied one.
	Symbol string `json:"symbol,omitempty"`
	// RuleKey is "<source>\x00<ruleId>".
	RuleKey string `json:"ruleKey"`
	// Mechanism is the normalized mechanism text.
	Mechanism string `json:"mechanism"`
	// Fingerprint is the normalized evidence fingerprint, or the secret digest.
	Fingerprint string `json:"fingerprint,omitempty"`
	// Secret marks Fingerprint as a secret digest that matches across paths.
	Secret bool `json:"secretDigest,omitempty"`
}

// Issue is a durable, mutable correlation of immutable observations.
// IssueID is opaque: issue:<family>:<ulid>, minted once.
type Issue struct {
	IssueID           string              `json:"issueId"`
	LegacyID          string              `json:"legacyId,omitempty"`
	State             IssueState          `json:"state"`
	Family            Family              `json:"family,omitempty"`
	Category          string              `json:"category,omitempty"`
	Title             string              `json:"title"`
	Mechanism         string              `json:"mechanism"`
	Anchor            IssueAnchor         `json:"anchor"`
	OriginalSeverity  Severity            `json:"originalSeverity"`
	EffectiveSeverity Severity            `json:"effectiveSeverity"`
	Owner             string              `json:"owner,omitempty"`
	DispositionReason string              `json:"dispositionReason,omitempty"`
	Uncertainty       string              `json:"uncertainty,omitempty"`
	AmbiguousWith     []string            `json:"ambiguousWith,omitempty"`
	ObservationIDs    []string            `json:"observationIds"`
	FeatureIDs        []string            `json:"featureIds,omitempty"`
	Files             []string            `json:"files,omitempty"`
	History           []IssueHistoryEntry `json:"history"`
	// Claim and Rule are set only on an issue promoted by a named, versioned
	// rule (design §3.2): Rule is "<rule id>@<version>". Empty for ordinary findings.
	Claim        SignalClaim   `json:"claim,omitempty"`
	Rule         string        `json:"rule,omitempty"`
	Verification *Verification `json:"verification,omitempty"`
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt"`
}

func (i Issue) Validate() error {
	if i.IssueID == "" || i.Title == "" || i.Mechanism == "" || !i.State.Valid() {
		return errors.New("issue requires identity, title, mechanism, and state")
	}
	if !i.Claim.ValidIssueClaim() || (i.Claim == "") != (i.Rule == "") {
		return errors.New("issue claim and promoting rule must be set together")
	}
	if !i.OriginalSeverity.Valid() || !i.EffectiveSeverity.Valid() {
		return errors.New("issue has invalid severity")
	}
	if i.EffectiveSeverity.Rank() > i.OriginalSeverity.Rank() {
		return errors.New("effective severity cannot exceed immutable source severity")
	}
	if len(i.ObservationIDs) == 0 {
		return errors.New("issue requires at least one observation")
	}
	if (i.State == IssueFalsePositive || i.State == IssueWontFix) && i.DispositionReason == "" {
		return errors.New("false-positive and won't-fix issues require an explicit disposition")
	}
	for _, path := range i.Files {
		if err := validateScopedPath(path); err != nil {
			return fmt.Errorf("issue file: %w", err)
		}
	}
	if err := validateIDs(i.FeatureIDs, "feature ID"); err != nil {
		return err
	}
	for _, entry := range i.History {
		if err := entry.Validate(); err != nil {
			return fmt.Errorf("issue history: %w", err)
		}
	}
	if i.State == IssueResolved {
		if i.Verification == nil {
			return errors.New("resolved issue requires verification")
		}
		if err := i.Verification.ValidateResolution(); err != nil {
			return err
		}
	}
	return nil
}

func validateScopedPath(value string) error {
	if value == "" || strings.ContainsRune(value, '\x00') || strings.Contains(value, "\\") {
		return errors.New("path must be a non-empty slash-separated relative path")
	}
	if filepath.IsAbs(value) || value == "." || value == ".." || strings.HasPrefix(value, "../") || filepath.Clean(value) != value {
		return errors.New("path escapes or is not normalized to the assessment scope")
	}
	return nil
}

func validateIDs(values []string, label string) error {
	for _, value := range values {
		if value == "" {
			return fmt.Errorf("%s must not be empty", label)
		}
	}
	return nil
}

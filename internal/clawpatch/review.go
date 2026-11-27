package clawpatch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"scout/internal/domain"
	"scout/internal/modelcli"
)

const (
	reviewSafetyPatchVersion = "scout-review-sanitize-v3"
	reviewSafetyPatchSHA256  = "dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3"
)

// SafetyFingerprint invalidates mapping/review caches when the pinned source
// or reviewed sanitizer patch changes.
func SafetyFingerprint() string {
	return Revision + ":" + reviewSafetyPatchSHA256
}

var (
	sensitiveValue = regexp.MustCompile(`(?i)\b(api[_-]?key|access[_-]?token|auth(?:entication|orization)?|client[_-]?secret|password|passwd|private[_-]?key|secret|session[_-]?token)\b\s*[:=]\s*([^\s,;)}\]]+)`)
	bearerValue    = regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=-]+\b`)
	jwtValue       = regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\b`)
	githubToken    = regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9_]{20,}\b`)
	openAIToken    = regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{16,}\b`)
	awsAccessKey   = regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)
	pemBlock       = regexp.MustCompile(`-----BEGIN [^-]+-----[\s\S]*?-----END [^-]+-----`)
)

// ReviewResult preserves completed and incomplete per-feature work rather than
// turning a failed provider call, dropped finding, or truncated prompt into a
// clean result.
type ReviewResult struct {
	Status       domain.AdapterStatus
	Coverage     []domain.Coverage
	Observations []domain.Observation
	Manifests    []ReviewManifest
	Reason       string
}

// ReviewManifest attributes the exact prompt surface Clawpatch reported for
// one explicit feature review. Consumers must treat an incomplete manifest as
// incomplete review coverage rather than a clean result.
type ReviewManifest struct {
	FeatureID         string
	MaxOwnedFiles     int
	MaxContextFiles   int
	PromptBytes       int
	ApproximateTokens int
	FindingCount      int
	FindingCap        int
	FindingCapReached bool
	IncludedFiles     []ReviewInputFile
	OmittedFiles      []ReviewOmittedFile
}

// ReviewInputFile describes one source excerpt included in a Clawpatch review.
type ReviewInputFile struct {
	Path          string
	Role          string
	Bytes         int
	IncludedBytes int
	Ranges        []domain.SourceRange
	Truncated     bool
	Readable      bool
	SkippedReason string
}

// ReviewOmittedFile describes source Clawpatch did not include in its prompt.
type ReviewOmittedFile struct {
	Path   string
	Role   string
	Reason string
}

type reviewResponse struct {
	Run             string                  `json:"run"`
	Findings        int                     `json:"findings"`
	Errors          []reviewError           `json:"errors"`
	ReviewManifests []featureReviewManifest `json:"reviewManifests"`
}

type reviewError struct {
	Message string  `json:"message"`
	Code    *string `json:"code"`
}

type featureReviewManifest struct {
	FeatureID string               `json:"featureId"`
	Manifest  reviewPromptManifest `json:"manifest"`
}

type reviewPromptManifest struct {
	MaxOwnedFiles     *int                `json:"maxOwnedFiles"`
	MaxContextFiles   *int                `json:"maxContextFiles"`
	PromptBytes       *int                `json:"promptBytes"`
	ApproximateTokens *int                `json:"approximateTokens"`
	IncludedFiles     []reviewPromptFile  `json:"includedFiles"`
	OmittedFiles      []reviewOmittedFile `json:"omittedFiles"`
}

type reviewPromptFile struct {
	Path               string            `json:"path"`
	Role               string            `json:"role"`
	Bytes              *int              `json:"bytes"`
	IncludedBytes      *int              `json:"includedBytes"`
	IncludedStartLine  *int              `json:"includedStartLine"`
	IncludedEndLine    *int              `json:"includedEndLine"`
	IncludedLineRanges []reviewLineRange `json:"includedLineRanges"`
	Truncated          bool              `json:"truncated"`
	Readable           bool              `json:"readable"`
	SkippedReason      *string           `json:"skippedReason"`
}

type reviewOmittedFile struct {
	Path   string `json:"path"`
	Role   string `json:"role"`
	Reason string `json:"reason"`
}

type reviewLineRange struct {
	StartLine int `json:"startLine"`
	EndLine   int `json:"endLine"`
}

type rawFinding struct {
	SchemaVersion  int               `json:"schemaVersion"`
	FindingID      string            `json:"findingId"`
	FeatureID      string            `json:"featureId"`
	CreatedByRunID string            `json:"createdByRunId"`
	Title          string            `json:"title"`
	Category       string            `json:"category"`
	Severity       domain.Severity   `json:"severity"`
	Confidence     domain.Confidence `json:"confidence"`
	Evidence       []rawEvidence     `json:"evidence"`
	Reasoning      string            `json:"reasoning"`
}

type rawEvidence struct {
	Path      string  `json:"path"`
	StartLine *int    `json:"startLine"`
	EndLine   *int    `json:"endLine"`
	Symbol    *string `json:"symbol"`
	Quote     *string `json:"quote"`
}

// Review invokes the security-reviewed patched CLI once for every explicit
// canonical feature ID. It never invokes fix. The marker ties execution to the
// reviewed source patch; an absent or altered marker blocks provider egress.
func (c *Client) Review(ctx context.Context, featureIDs []string) (ReviewResult, error) {
	ids, err := uniqueFeatureIDs(featureIDs)
	if err != nil {
		return ReviewResult{Status: domain.AdapterFailed}, err
	}
	if len(ids) == 0 {
		return ReviewResult{Status: domain.AdapterNoApplicableInput}, nil
	}
	invocation, stop, err := c.prepareReview(ctx, ids)
	if stop != nil {
		return *stop, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.reviewFeaturesLocked(ctx, ids, invocation)
}

type reviewInvocation struct {
	providerArgs []string
	env          []string
}

type featureReview struct {
	coverage     domain.Coverage
	manifest     ReviewManifest
	observations []domain.Observation
	partial      bool
}

func (c *Client) prepareReview(ctx context.Context, ids []string) (reviewInvocation, *ReviewResult, error) {
	provider, err := c.resolveProvider()
	if errors.Is(err, modelcli.ErrNoCLI) {
		skipped := skippedReviewResult(ids)
		return reviewInvocation{}, &skipped, nil
	}
	if err != nil {
		blocked := blockedReviewResult(ids, err.Error())
		return reviewInvocation{}, &blocked, nil
	}
	if !c.egressAuthorized {
		blocked := blockedReviewResult(ids, "Clawpatch review provider egress is not authorized")
		return reviewInvocation{}, &blocked, nil
	}
	providerArgs, reviewEnv := modelInvocation(provider)
	if err := c.Init(ctx); err != nil {
		blocked := blockedReviewResult(ids, err.Error())
		return reviewInvocation{}, &blocked, err
	}
	if err := c.verifyReviewSafetyPatch(); err != nil {
		blocked := blockedReviewResult(ids, err.Error())
		return reviewInvocation{}, &blocked, nil
	}
	return reviewInvocation{providerArgs: providerArgs, env: reviewEnv}, nil, nil
}

func (c *Client) reviewFeaturesLocked(ctx context.Context, ids []string, invocation reviewInvocation) (ReviewResult, error) {
	features, err := loadFeatures(c.stateDir)
	if err != nil {
		return blockedReviewResult(ids, "canonical Clawpatch feature records are unavailable"), err
	}
	findingCap, err := reviewFindingCap(c.stateDir)
	if err != nil {
		return blockedReviewResult(ids, "Clawpatch review finding cap is unavailable"), err
	}
	known := featuresByID(features)

	result := ReviewResult{Status: domain.AdapterCompleted}
	for _, id := range ids {
		feature, ok := known[id]
		if !ok {
			result.Coverage = append(result.Coverage, blockedCoverage(id, "selected canonical Clawpatch feature record is unavailable"))
			result.Status = domain.AdapterPartial
			continue
		}
		review := c.reviewFeatureLocked(ctx, feature, findingCap, invocation)
		result.Coverage = append(result.Coverage, review.coverage)
		if review.manifest.FeatureID != "" {
			result.Manifests = append(result.Manifests, review.manifest)
		}
		result.Observations = append(result.Observations, review.observations...)
		if review.partial {
			result.Status = domain.AdapterPartial
		}
	}
	if result.Status == domain.AdapterPartial {
		result.Reason = "one or more selected Clawpatch reviews were blocked or incomplete"
	}
	return result, nil
}

func featuresByID(features []domain.Feature) map[string]domain.Feature {
	known := make(map[string]domain.Feature, len(features))
	for _, feature := range features {
		known[feature.FeatureID] = feature
	}
	return known
}

func (c *Client) reviewFeatureLocked(ctx context.Context, feature domain.Feature, findingCap int, invocation reviewInvocation) featureReview {
	id := feature.FeatureID
	output, err := c.invokeEnvLocked(ctx, invocation.env, append([]string{"review", "--feature", id}, invocation.providerArgs...)...)
	if err != nil {
		return featureReview{coverage: incompleteCoverage(id, "pinned Clawpatch review did not complete"), partial: true}
	}
	featureResult, err := decodeReviewResponse(output)
	if err != nil {
		return featureReview{coverage: incompleteCoverage(id, "pinned Clawpatch review returned invalid JSON"), partial: true}
	}
	coverage, manifest, incomplete := coverageFromReview(feature, featureResult, findingCap)
	review := featureReview{coverage: coverage, manifest: manifest, partial: incomplete}
	observations, err := loadReviewObservations(c.root, c.stateDir, id, featureResult.Run)
	if err != nil {
		review.coverage = markIncompleteReview(review.coverage, "one or more sanitized Clawpatch findings could not be validated against the snapshot")
		review.partial = true
		return review
	}
	if len(observations) != featureResult.Findings {
		review.coverage = markIncompleteReview(review.coverage, "review findings could not be attributed to the reported Clawpatch run")
		review.partial = true
	}
	review.observations = observations
	if len(observations) != 0 && review.coverage.Status == domain.CoverageInspected {
		review.coverage = flagObservedCoverage(review.coverage, observations)
	}
	return review
}

func flagObservedCoverage(coverage domain.Coverage, observations []domain.Observation) domain.Coverage {
	observationIDs := make([]string, 0, len(observations))
	for _, observation := range observations {
		observationIDs = append(observationIDs, observation.ObservationID)
	}
	coverage.Status = domain.CoverageFlagged
	coverage.ObservationIDs = observationIDs
	return coverage
}

func (c *Client) verifyReviewSafetyPatch() error {
	markerPath := filepath.Join(filepath.Dir(filepath.Dir(c.binaryPath)), ".scout-review-safety")
	marker, err := os.ReadFile(markerPath)
	if err != nil {
		return errors.New("review safety patch marker is unavailable; provider review is blocked")
	}
	if strings.TrimSpace(string(marker)) != reviewSafetyPatchVersion+" "+reviewSafetyPatchSHA256 {
		return errors.New("review safety patch marker does not match the reviewed build; provider review is blocked")
	}
	return nil
}

func decodeReviewResponse(raw []byte) (reviewResponse, error) {
	var response reviewResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return reviewResponse{}, err
	}
	if response.Run == "" {
		return reviewResponse{}, errors.New("review response has no run ID")
	}
	return response, nil
}

func coverageFromReview(feature domain.Feature, response reviewResponse, findingCap int) (domain.Coverage, ReviewManifest, bool) {
	source, missingReason := featurePromptManifest(feature.FeatureID, response.ReviewManifests)
	if missingReason != "" {
		return incompleteCoverage(feature.FeatureID, missingReason), ReviewManifest{}, true
	}
	review := newManifestReview(feature.FeatureID, response, findingCap, source)
	for _, omittedFile := range source.OmittedFiles {
		review.addOmittedFile(omittedFile)
	}
	for _, file := range source.IncludedFiles {
		review.addIncludedFile(file)
	}
	review.accountFeaturePaths(feature)
	return review.coverage(feature.FeatureID, len(response.Errors) != 0)
}

func featurePromptManifest(featureID string, manifests []featureReviewManifest) (*reviewPromptManifest, string) {
	var source *reviewPromptManifest
	for _, candidate := range manifests {
		if candidate.FeatureID != featureID {
			continue
		}
		if source != nil {
			return nil, "review returned multiple manifests for one feature"
		}
		manifest := candidate.Manifest
		source = &manifest
	}
	if source == nil {
		return nil, "review did not return the inspected-file manifest"
	}
	return source, ""
}

type manifestReview struct {
	manifest       ReviewManifest
	incomplete     bool
	ranges         []domain.SourceRange
	omitted        []string
	accountedPaths map[string]struct{}
}

func newManifestReview(featureID string, response reviewResponse, findingCap int, source *reviewPromptManifest) *manifestReview {
	review := &manifestReview{
		manifest:       ReviewManifest{FeatureID: featureID, FindingCount: response.Findings, FindingCap: findingCap},
		ranges:         make([]domain.SourceRange, 0, len(source.IncludedFiles)),
		omitted:        make([]string, 0, len(source.OmittedFiles)),
		accountedPaths: make(map[string]struct{}, len(source.IncludedFiles)+len(source.OmittedFiles)),
	}
	limitsValid := applyPromptLimits(&review.manifest, source)
	review.incomplete = len(response.Errors) != 0 || !limitsValid
	review.manifest.FindingCapReached = findingCap < 1 || response.Findings >= findingCap
	review.incomplete = review.incomplete || review.manifest.FindingCapReached
	return review
}

func applyPromptLimits(manifest *ReviewManifest, source *reviewPromptManifest) bool {
	if source.MaxOwnedFiles == nil || source.MaxContextFiles == nil || source.PromptBytes == nil || source.ApproximateTokens == nil {
		return false
	}
	manifest.MaxOwnedFiles = *source.MaxOwnedFiles
	manifest.MaxContextFiles = *source.MaxContextFiles
	manifest.PromptBytes = *source.PromptBytes
	manifest.ApproximateTokens = *source.ApproximateTokens
	return manifest.MaxOwnedFiles >= 1 && manifest.MaxContextFiles >= 1 && manifest.PromptBytes >= 1 && manifest.ApproximateTokens >= 1
}

func (r *manifestReview) addOmittedFile(omittedFile reviewOmittedFile) {
	r.manifest.OmittedFiles = append(r.manifest.OmittedFiles, ReviewOmittedFile(omittedFile))
	if !safeScopedPath(omittedFile.Path) {
		r.incomplete = true
		return
	}
	r.accountedPaths[omittedFile.Path] = struct{}{}
	r.omitted = append(r.omitted, omittedFile.Path)
	r.incomplete = true
}

func (r *manifestReview) addIncludedFile(file reviewPromptFile) {
	input := reviewInputFile(file)
	r.manifest.IncludedFiles = append(r.manifest.IncludedFiles, input)
	if !safeScopedPath(file.Path) {
		r.incomplete = true
		return
	}
	r.accountedPaths[file.Path] = struct{}{}
	if !includedFileComplete(file) {
		r.incomplete = true
		r.omitted = append(r.omitted, file.Path)
		return
	}
	for _, lineRange := range file.IncludedLineRanges {
		if lineRange.StartLine < 1 || lineRange.EndLine < lineRange.StartLine {
			r.incomplete = true
			continue
		}
		location := domain.SourceRange{Path: file.Path, StartLine: lineRange.StartLine, EndLine: lineRange.EndLine}
		input.Ranges = append(input.Ranges, location)
		r.ranges = append(r.ranges, location)
	}
	r.manifest.IncludedFiles[len(r.manifest.IncludedFiles)-1] = input
}

func reviewInputFile(file reviewPromptFile) ReviewInputFile {
	input := ReviewInputFile{Path: file.Path, Role: file.Role, Truncated: file.Truncated, Readable: file.Readable}
	if file.Bytes != nil {
		input.Bytes = *file.Bytes
	}
	if file.IncludedBytes != nil {
		input.IncludedBytes = *file.IncludedBytes
	}
	if file.SkippedReason != nil {
		input.SkippedReason = *file.SkippedReason
	}
	return input
}

func includedFileComplete(file reviewPromptFile) bool {
	return file.Bytes != nil && file.IncludedBytes != nil && file.Readable && !file.Truncated && file.SkippedReason == nil && *file.Bytes >= 0 && *file.IncludedBytes >= 0 && *file.IncludedBytes <= *file.Bytes && len(file.IncludedLineRanges) != 0
}

func (r *manifestReview) accountFeaturePaths(feature domain.Feature) {
	for _, path := range featureReviewPaths(feature) {
		if _, found := r.accountedPaths[path]; !found {
			r.incomplete = true
			r.omitted = append(r.omitted, path)
		}
	}
}

func (r *manifestReview) coverage(featureID string, reportedErrors bool) (domain.Coverage, ReviewManifest, bool) {
	if r.incomplete || len(r.ranges) == 0 {
		reason := "review manifest omitted, truncated, unreadable, or unattributed work"
		if reportedErrors {
			reason = "review reported dropped or invalid findings"
		} else if r.manifest.FindingCapReached {
			reason = "review reached the configured finding cap"
		}
		return domain.Coverage{
			Key:             reviewCoverageKey(featureID),
			Status:          domain.CoverageIncomplete,
			Reason:          reason,
			InspectedRanges: r.ranges,
			OmittedPaths:    uniquePaths(r.omitted),
		}, r.manifest, true
	}
	return domain.Coverage{
		Key:             reviewCoverageKey(featureID),
		Status:          domain.CoverageInspected,
		InspectedRanges: r.ranges,
	}, r.manifest, false
}

func featureReviewPaths(feature domain.Feature) []string {
	paths := make([]string, 0, len(feature.OwnedFiles)+len(feature.ContextFiles)+len(feature.Tests))
	for _, ref := range feature.OwnedFiles {
		paths = append(paths, ref.Path)
	}
	for _, ref := range feature.ContextFiles {
		paths = append(paths, ref.Path)
	}
	for _, ref := range feature.Tests {
		paths = append(paths, ref.Path)
	}
	return uniquePaths(paths)
}

func markIncompleteReview(coverage domain.Coverage, reason string) domain.Coverage {
	coverage.Status = domain.CoverageIncomplete
	coverage.Reason = reason
	coverage.ObservationIDs = nil
	return coverage
}

func reviewFindingCap(stateDir string) (int, error) {
	raw, err := os.ReadFile(filepath.Join(stateDir, "config.json"))
	if err != nil {
		return 0, fmt.Errorf("read Clawpatch review configuration: %w", err)
	}
	var config struct {
		SchemaVersion int `json:"schemaVersion"`
		Review        struct {
			MaxFindingsPerFeature int `json:"maxFindingsPerFeature"`
		} `json:"review"`
	}
	if err := json.Unmarshal(raw, &config); err != nil {
		return 0, fmt.Errorf("decode Clawpatch review configuration: %w", err)
	}
	if config.SchemaVersion != 1 || config.Review.MaxFindingsPerFeature < 1 {
		return 0, errors.New("invalid Clawpatch review finding cap")
	}
	return config.Review.MaxFindingsPerFeature, nil
}

func loadReviewObservations(root, stateDir, featureID, runID string) ([]domain.Observation, error) {
	// Review rewrites the feature record with its new finding IDs, so the
	// record is read again after the review rather than reused from before it.
	features, err := loadFeatures(stateDir)
	if err != nil {
		return nil, err
	}
	for _, feature := range features {
		if feature.FeatureID != featureID {
			continue
		}
		observations := make([]domain.Observation, 0, len(feature.FindingIDs))
		for _, findingID := range feature.FindingIDs {
			finding, err := loadRawFinding(stateDir, findingID)
			if err != nil {
				return observations, err
			}
			if finding.FeatureID != featureID {
				return observations, errors.New("finding does not belong to reviewed feature")
			}
			if finding.CreatedByRunID != runID {
				continue
			}
			observation, err := observationFromFinding(root, runID, finding)
			if err != nil {
				return observations, err
			}
			observations = append(observations, observation)
		}
		return observations, nil
	}
	return nil, errors.New("reviewed feature record disappeared")
}

func loadRawFinding(stateDir, findingID string) (rawFinding, error) {
	if findingID == "" || filepath.Base(findingID) != findingID || strings.ContainsRune(findingID, filepath.Separator) {
		return rawFinding{}, errors.New("unsafe Clawpatch finding ID")
	}
	raw, err := os.ReadFile(filepath.Join(stateDir, "findings", findingID+".json"))
	if err != nil {
		return rawFinding{}, err
	}
	var finding rawFinding
	if err := json.Unmarshal(raw, &finding); err != nil {
		return rawFinding{}, err
	}
	if !validRawFinding(finding, findingID) {
		return rawFinding{}, errors.New("invalid Clawpatch finding record")
	}
	return finding, nil
}

func validRawFinding(finding rawFinding, findingID string) bool {
	return finding.SchemaVersion == 1 && finding.FindingID == findingID && rawFindingHasContent(finding) && finding.Severity.Valid() && finding.Confidence.Valid()
}

func rawFindingHasContent(finding rawFinding) bool {
	return finding.FeatureID != "" && finding.CreatedByRunID != "" && finding.Title != "" && finding.Category != "" && finding.Reasoning != "" && len(finding.Evidence) != 0
}

func observationFromFinding(root, runID string, finding rawFinding) (domain.Observation, error) {
	evidence := make([]domain.Evidence, 0, len(finding.Evidence))
	for _, raw := range finding.Evidence {
		location, err := validateRawEvidence(root, raw)
		if err != nil {
			return domain.Observation{}, err
		}
		evidence = append(evidence, domain.Evidence{Location: location})
	}
	observation := domain.Observation{
		ObservationID:     "clawpatch:" + runID + ":" + finding.FindingID,
		RunID:             runID,
		Source:            "clawpatch@" + Version + "+" + reviewSafetyPatchVersion,
		SourceFindingID:   finding.FindingID,
		Family:            domain.FamilyClawpatch,
		OriginalCategory:  finding.Category,
		SourceSeverity:    finding.Severity,
		EffectiveSeverity: finding.Severity,
		Confidence:        finding.Confidence,
		Mechanism:         sanitizeFindingText(finding.Title + ": " + finding.Reasoning),
		Evidence:          evidence,
		Validated:         true,
		ObservedAt:        time.Now().UTC(),
	}
	if err := observation.Validate(); err != nil {
		return domain.Observation{}, err
	}
	return observation, nil
}

func validateRawEvidence(root string, evidence rawEvidence) (domain.SourceRange, error) {
	relativePath, realPath, err := resolveEvidencePath(root, evidence.Path)
	if err != nil {
		return domain.SourceRange{}, err
	}
	location := domain.SourceRange{Path: filepath.ToSlash(relativePath)}
	if evidence.StartLine != nil || evidence.EndLine != nil {
		if err := checkEvidenceLines(realPath, evidence); err != nil {
			return domain.SourceRange{}, err
		}
		location.StartLine = *evidence.StartLine
		location.EndLine = *evidence.EndLine
	}
	if evidence.Symbol != nil {
		location.Symbol = *evidence.Symbol
	}
	if err := location.Validate(); err != nil {
		return domain.SourceRange{}, err
	}
	return location, nil
}

func resolveEvidencePath(root, path string) (string, string, error) {
	if path == "" || filepath.IsAbs(path) {
		return "", "", errors.New("finding has unsafe evidence path")
	}
	fullPath := filepath.Join(root, path)
	relativePath, err := filepath.Rel(root, fullPath)
	if err != nil || escapesBase(relativePath) {
		return "", "", errors.New("finding evidence escapes assessment root")
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", "", err
	}
	realPath, err := filepath.EvalSymlinks(fullPath)
	if err != nil {
		return "", "", err
	}
	realRelative, err := filepath.Rel(realRoot, realPath)
	if err != nil || escapesBase(realRelative) {
		return "", "", errors.New("finding evidence resolves outside assessment root")
	}
	return relativePath, realPath, nil
}

func escapesBase(relativePath string) bool {
	return relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator))
}

func checkEvidenceLines(realPath string, evidence rawEvidence) error {
	if evidence.StartLine == nil || evidence.EndLine == nil || *evidence.StartLine < 1 || *evidence.EndLine < *evidence.StartLine {
		return errors.New("finding has invalid evidence line range")
	}
	contents, err := os.ReadFile(realPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(contents), "\n")
	if *evidence.EndLine > len(lines) {
		return errors.New("finding evidence line range exceeds snapshot")
	}
	if !evidenceQuoteMatches(evidence, lines) {
		return errors.New("finding evidence quote does not match snapshot")
	}
	return nil
}

func evidenceQuoteMatches(evidence rawEvidence, lines []string) bool {
	return evidence.Quote == nil || strings.Contains(*evidence.Quote, "[SCOUT_REDACTED") || strings.Contains(compactWhitespace(strings.Join(lines[*evidence.StartLine-1:*evidence.EndLine], "\n")), compactWhitespace(*evidence.Quote))
}

func blockedReviewResult(ids []string, reason string) ReviewResult {
	coverage := make([]domain.Coverage, 0, len(ids))
	for _, id := range ids {
		coverage = append(coverage, blockedCoverage(id, reason))
	}
	return ReviewResult{Status: domain.AdapterPartial, Coverage: coverage, Reason: reason}
}

func blockedCoverage(featureID, reason string) domain.Coverage {
	return domain.Coverage{Key: reviewCoverageKey(featureID), Status: domain.CoverageBlocked, Reason: reason}
}

func incompleteCoverage(featureID, reason string) domain.Coverage {
	return domain.Coverage{Key: reviewCoverageKey(featureID), Status: domain.CoverageIncomplete, Reason: reason}
}

func reviewCoverageKey(featureID string) domain.CoverageKey {
	return domain.CoverageKey{FeatureID: featureID, Family: domain.FamilyClawpatch, Category: "review"}
}

func uniqueFeatureIDs(featureIDs []string) ([]string, error) {
	seen := make(map[string]struct{}, len(featureIDs))
	ids := make([]string, 0, len(featureIDs))
	for _, id := range featureIDs {
		if id == "" {
			return nil, errors.New("clawpatch review requires a canonical feature ID")
		}
		if _, duplicate := seen[id]; duplicate {
			return nil, fmt.Errorf("duplicate selected Clawpatch feature ID %q", id)
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids, nil
}

func uniquePaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		result = append(result, path)
	}
	sort.Strings(result)
	return result
}

func safeScopedPath(path string) bool {
	if path == "" || filepath.IsAbs(path) || strings.ContainsRune(path, '\x00') {
		return false
	}
	clean := filepath.Clean(path)
	return clean != "." && clean != ".." && !strings.HasPrefix(clean, ".."+string(filepath.Separator))
}

func compactWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func sanitizeFindingText(value string) string {
	value = pemBlock.ReplaceAllStringFunc(value, func(match string) string {
		return "[SCOUT_REDACTED_SECRET]" + strings.Map(func(r rune) rune {
			if r == '\n' || r == '\r' {
				return r
			}
			return -1
		}, match)
	})
	value = sensitiveValue.ReplaceAllString(value, "$1=[SCOUT_REDACTED_SECRET]")
	value = bearerValue.ReplaceAllString(value, "Bearer [SCOUT_REDACTED_SECRET]")
	value = jwtValue.ReplaceAllString(value, "[SCOUT_REDACTED_JWT]")
	value = githubToken.ReplaceAllString(value, "[SCOUT_REDACTED_TOKEN]")
	value = openAIToken.ReplaceAllString(value, "[SCOUT_REDACTED_TOKEN]")
	return awsAccessKey.ReplaceAllString(value, "[SCOUT_REDACTED_AWS_KEY]")
}

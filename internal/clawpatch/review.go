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
)

const (
	reviewSafetyPatchVersion     = "scout-review-sanitize-v2"
	reviewSafetyPatchSHA256      = "f39d73dad42b5d6701b1fee51f7784d3eb2b1548808d1ee47d8ef146ca97a865"
	defaultMaxFindingsPerFeature = 10
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
	if !c.egressAuthorized {
		return blockedReviewResult(ids, "Clawpatch review provider egress is not authorized"), nil
	}
	if err := c.Init(ctx); err != nil {
		return blockedReviewResult(ids, err.Error()), err
	}
	if err := c.verifyReviewSafetyPatch(); err != nil {
		return blockedReviewResult(ids, err.Error()), nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	features, err := loadFeatures(c.stateDir)
	if err != nil {
		return blockedReviewResult(ids, "canonical Clawpatch feature records are unavailable"), err
	}
	findingCap, err := reviewFindingCap(c.stateDir)
	if err != nil {
		return blockedReviewResult(ids, "Clawpatch review finding cap is unavailable"), err
	}
	known := make(map[string]domain.Feature, len(features))
	for _, feature := range features {
		known[feature.FeatureID] = feature
	}

	result := ReviewResult{Status: domain.AdapterCompleted}
	for _, id := range ids {
		feature, ok := known[id]
		if !ok {
			result.Coverage = append(result.Coverage, blockedCoverage(id, "selected canonical Clawpatch feature record is unavailable"))
			result.Status = domain.AdapterPartial
			continue
		}
		output, err := c.invokeLocked(ctx, "review", "--feature", id)
		if err != nil {
			result.Coverage = append(result.Coverage, incompleteCoverage(id, "pinned Clawpatch review did not complete"))
			result.Status = domain.AdapterPartial
			continue
		}
		featureResult, err := decodeReviewResponse(output)
		if err != nil {
			result.Coverage = append(result.Coverage, incompleteCoverage(id, "pinned Clawpatch review returned invalid JSON"))
			result.Status = domain.AdapterPartial
			continue
		}
		coverage, manifest, incomplete := coverageFromReview(feature, featureResult, findingCap)
		result.Coverage = append(result.Coverage, coverage)
		if manifest.FeatureID != "" {
			result.Manifests = append(result.Manifests, manifest)
		}
		if incomplete {
			result.Status = domain.AdapterPartial
		}
		observations, err := loadReviewObservations(c.root, c.stateDir, id, featureResult.Run)
		if err != nil {
			result.Coverage[len(result.Coverage)-1] = markIncompleteReview(result.Coverage[len(result.Coverage)-1], "one or more sanitized Clawpatch findings could not be validated against the snapshot")
			result.Status = domain.AdapterPartial
			continue
		}
		if len(observations) != featureResult.Findings {
			result.Coverage[len(result.Coverage)-1] = markIncompleteReview(result.Coverage[len(result.Coverage)-1], "review findings could not be attributed to the reported Clawpatch run")
			result.Status = domain.AdapterPartial
		}
		result.Observations = append(result.Observations, observations...)
		if len(observations) != 0 && result.Coverage[len(result.Coverage)-1].Status == domain.CoverageInspected {
			observationIDs := make([]string, 0, len(observations))
			for _, observation := range observations {
				observationIDs = append(observationIDs, observation.ObservationID)
			}
			result.Coverage[len(result.Coverage)-1].Status = domain.CoverageFlagged
			result.Coverage[len(result.Coverage)-1].ObservationIDs = observationIDs
		}
	}
	if result.Status == domain.AdapterPartial {
		result.Reason = "one or more selected Clawpatch reviews were blocked or incomplete"
	}
	return result, nil
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
	var source *reviewPromptManifest
	for _, candidate := range response.ReviewManifests {
		if candidate.FeatureID != feature.FeatureID {
			continue
		}
		if source != nil {
			return incompleteCoverage(feature.FeatureID, "review returned multiple manifests for one feature"), ReviewManifest{}, true
		}
		manifest := candidate.Manifest
		source = &manifest
	}
	if source == nil {
		return incompleteCoverage(feature.FeatureID, "review did not return the inspected-file manifest"), ReviewManifest{}, true
	}

	manifest := ReviewManifest{FeatureID: feature.FeatureID, FindingCount: response.Findings, FindingCap: findingCap}
	incomplete := len(response.Errors) != 0
	if source.MaxOwnedFiles == nil || source.MaxContextFiles == nil || source.PromptBytes == nil || source.ApproximateTokens == nil {
		incomplete = true
	} else {
		manifest.MaxOwnedFiles = *source.MaxOwnedFiles
		manifest.MaxContextFiles = *source.MaxContextFiles
		manifest.PromptBytes = *source.PromptBytes
		manifest.ApproximateTokens = *source.ApproximateTokens
		if manifest.MaxOwnedFiles < 1 || manifest.MaxContextFiles < 1 || manifest.PromptBytes < 1 || manifest.ApproximateTokens < 1 {
			incomplete = true
		}
	}
	manifest.FindingCapReached = findingCap < 1 || response.Findings >= findingCap
	incomplete = incomplete || manifest.FindingCapReached

	ranges := make([]domain.SourceRange, 0, len(source.IncludedFiles))
	omitted := make([]string, 0, len(source.OmittedFiles))
	accountedPaths := make(map[string]struct{}, len(source.IncludedFiles)+len(source.OmittedFiles))
	for _, omittedFile := range source.OmittedFiles {
		manifest.OmittedFiles = append(manifest.OmittedFiles, ReviewOmittedFile{Path: omittedFile.Path, Role: omittedFile.Role, Reason: omittedFile.Reason})
		if !safeScopedPath(omittedFile.Path) {
			incomplete = true
			continue
		}
		accountedPaths[omittedFile.Path] = struct{}{}
		omitted = append(omitted, omittedFile.Path)
		incomplete = true
	}
	for _, file := range source.IncludedFiles {
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
		manifest.IncludedFiles = append(manifest.IncludedFiles, input)
		if !safeScopedPath(file.Path) {
			incomplete = true
			continue
		}
		accountedPaths[file.Path] = struct{}{}
		if file.Bytes == nil || file.IncludedBytes == nil || !file.Readable || file.Truncated || file.SkippedReason != nil || *file.Bytes < 0 || *file.IncludedBytes < 0 || *file.IncludedBytes > *file.Bytes || len(file.IncludedLineRanges) == 0 {
			incomplete = true
			omitted = append(omitted, file.Path)
			continue
		}
		for _, lineRange := range file.IncludedLineRanges {
			if lineRange.StartLine < 1 || lineRange.EndLine < lineRange.StartLine {
				incomplete = true
				continue
			}
			location := domain.SourceRange{Path: file.Path, StartLine: lineRange.StartLine, EndLine: lineRange.EndLine}
			input.Ranges = append(input.Ranges, location)
			ranges = append(ranges, location)
		}
		manifest.IncludedFiles[len(manifest.IncludedFiles)-1] = input
	}
	for _, path := range featureReviewPaths(feature) {
		if _, found := accountedPaths[path]; !found {
			incomplete = true
			omitted = append(omitted, path)
		}
	}
	if incomplete || len(ranges) == 0 {
		reason := "review manifest omitted, truncated, unreadable, or unattributed work"
		if len(response.Errors) != 0 {
			reason = "review reported dropped or invalid findings"
		} else if manifest.FindingCapReached {
			reason = "review reached the configured finding cap"
		}
		return domain.Coverage{
			Key:             reviewCoverageKey(feature.FeatureID),
			Status:          domain.CoverageIncomplete,
			Reason:          reason,
			InspectedRanges: ranges,
			OmittedPaths:    uniquePaths(omitted),
		}, manifest, true
	}
	return domain.Coverage{
		Key:             reviewCoverageKey(feature.FeatureID),
		Status:          domain.CoverageInspected,
		InspectedRanges: ranges,
	}, manifest, false
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
	if finding.SchemaVersion != 1 || finding.FindingID != findingID || finding.FeatureID == "" || finding.CreatedByRunID == "" || finding.Title == "" || finding.Category == "" || finding.Reasoning == "" || len(finding.Evidence) == 0 || !finding.Severity.Valid() || !finding.Confidence.Valid() {
		return rawFinding{}, errors.New("invalid Clawpatch finding record")
	}
	return finding, nil
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
		Source:            "clawpatch@0.7.3+scout-review-sanitize-v2",
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
	if evidence.Path == "" || filepath.IsAbs(evidence.Path) {
		return domain.SourceRange{}, errors.New("finding has unsafe evidence path")
	}
	fullPath := filepath.Join(root, evidence.Path)
	relativePath, err := filepath.Rel(root, fullPath)
	if err != nil || relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		return domain.SourceRange{}, errors.New("finding evidence escapes assessment root")
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return domain.SourceRange{}, err
	}
	realPath, err := filepath.EvalSymlinks(fullPath)
	if err != nil {
		return domain.SourceRange{}, err
	}
	realRelative, err := filepath.Rel(realRoot, realPath)
	if err != nil || realRelative == ".." || strings.HasPrefix(realRelative, ".."+string(filepath.Separator)) {
		return domain.SourceRange{}, errors.New("finding evidence resolves outside assessment root")
	}
	location := domain.SourceRange{Path: filepath.ToSlash(relativePath)}
	if evidence.StartLine != nil || evidence.EndLine != nil {
		if evidence.StartLine == nil || evidence.EndLine == nil || *evidence.StartLine < 1 || *evidence.EndLine < *evidence.StartLine {
			return domain.SourceRange{}, errors.New("finding has invalid evidence line range")
		}
		contents, err := os.ReadFile(realPath)
		if err != nil {
			return domain.SourceRange{}, err
		}
		lines := strings.Split(string(contents), "\n")
		if *evidence.EndLine > len(lines) {
			return domain.SourceRange{}, errors.New("finding evidence line range exceeds snapshot")
		}
		if evidence.Quote != nil && !strings.Contains(*evidence.Quote, "[SCOUT_REDACTED") && !strings.Contains(compactWhitespace(strings.Join(lines[*evidence.StartLine-1:*evidence.EndLine], "\n")), compactWhitespace(*evidence.Quote)) {
			return domain.SourceRange{}, errors.New("finding evidence quote does not match snapshot")
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
			return nil, errors.New("Clawpatch review requires a canonical feature ID")
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

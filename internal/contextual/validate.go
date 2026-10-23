package contextual

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"scout/internal/domain"
)

type providerResponse struct {
	Status                 string             `json:"status"`
	Reason                 string             `json:"reason"`
	Inspected              []providerCitation `json:"inspected"`
	DeterministicBlindSpot string             `json:"deterministicBlindSpot"`
	Findings               []providerFinding  `json:"findings"`
	Omissions              []string           `json:"omissions"`
}

type providerCitation struct {
	Path      string `json:"path"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
	Quote     string `json:"quote"`
}

type providerFinding struct {
	Mechanism    string           `json:"mechanism"`
	Evidence     providerCitation `json:"evidence"`
	Confidence   string           `json:"confidence"`
	Uncertainty  string           `json:"uncertainty"`
	Reproduction string           `json:"reproduction"`
	Severity     json.RawMessage  `json:"severity"`
}

func validateProviderResponse(request ReviewRequest, raw []byte, manifest Manifest, source []promptSource, model, promptVersion string) (ReviewResult, error) {
	response, err := decodeProviderResponse(raw)
	if err != nil {
		return emptyResult(request), err
	}
	status, err := responseCoverageStatus(response.Status)
	if err != nil {
		return emptyResult(request), err
	}
	if (status == domain.CoverageNotApplicable || status == domain.CoverageRequiresManualVerification) && strings.TrimSpace(response.Reason) == "" {
		return emptyResult(request), errors.New("provider response requires a reason for non-covered status")
	}
	if len(response.Inspected) == 0 {
		return emptyResult(request), errors.New("provider response has no inspected source rows")
	}
	if strings.TrimSpace(response.DeterministicBlindSpot) == "" {
		return emptyResult(request), errors.New("provider response has no deterministic blind-spot judgment")
	}
	known := sourceByPath(source)
	inspected := make([]domain.SourceRange, 0, len(response.Inspected))
	for _, citation := range response.Inspected {
		resolved, err := validateCitation(citation, manifest, known)
		if err != nil {
			return emptyResult(request), fmt.Errorf("provider inspected row: %w", err)
		}
		inspected = appendUniqueRange(inspected, resolved)
	}
	if len(inspected) == 0 {
		return emptyResult(request), errors.New("provider response has no valid inspected source rows")
	}
	if status == domain.CoverageFlagged && len(response.Findings) == 0 {
		return emptyResult(request), errors.New("flagged provider response has no findings")
	}
	if status == domain.CoverageCovered && len(response.Findings) != 0 {
		return emptyResult(request), errors.New("covered provider response cannot contain findings")
	}
	if status == domain.CoverageNotApplicable && len(response.Findings) != 0 {
		return emptyResult(request), errors.New("not-applicable provider response cannot contain findings")
	}

	redactor := secretRedactor{}
	result := ReviewResult{
		Coverage: domain.Coverage{
			Key:              domain.CoverageKey{FeatureID: request.Feature.FeatureID, Family: request.Family, Category: request.Category},
			Status:           status,
			Reason:           redactor.redact(strings.TrimSpace(response.Reason)),
			InspectedRanges:  inspected,
			AdapterResultIDs: adapterResultIDs(request.ToolResults),
		},
		Metadata: &ReviewMetadata{DeterministicBlindSpot: redactor.redact(strings.TrimSpace(response.DeterministicBlindSpot))},
	}
	result.Coverage.OmittedPaths = validOmittedPaths(response.Omissions, manifest)
	var failures []error
	for _, finding := range response.Findings {
		observation, err := validatedObservation(request, finding, manifest, known, model, promptVersion, redactor)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		result.Observations = append(result.Observations, observation)
		result.Coverage.ObservationIDs = append(result.Coverage.ObservationIDs, observation.ObservationID)
		if result.Metadata.ReproductionByObservationID == nil {
			result.Metadata.ReproductionByObservationID = make(map[string]string)
		}
		result.Metadata.ReproductionByObservationID[observation.ObservationID] = redactor.redact(strings.TrimSpace(finding.Reproduction))
	}
	if len(failures) != 0 {
		return result, errors.Join(failures...)
	}
	if err := result.Coverage.Validate(); err != nil {
		return result, fmt.Errorf("validated contextual coverage: %w", err)
	}
	return result, nil
}

func emptyResult(request ReviewRequest) ReviewResult {
	return ReviewResult{Coverage: domain.Coverage{Key: domain.CoverageKey{FeatureID: request.Feature.FeatureID, Family: request.Family, Category: request.Category}}}
}

// decodeProviderResponse parses the model's answer strictly: one JSON object
// with the declared fields only. No fence stripping, envelope or repair.
func decodeProviderResponse(raw []byte) (providerResponse, error) {
	var response providerResponse
	if err := decodeStrictJSON(raw, &response); err != nil {
		return providerResponse{}, fmt.Errorf("decode model answer JSON: %w", err)
	}
	if response.Status == "" {
		return providerResponse{}, errors.New("model answer has no status")
	}
	return response, nil
}

func responseCoverageStatus(value string) (domain.CoverageStatus, error) {
	switch value {
	case string(domain.CoverageCovered):
		return domain.CoverageCovered, nil
	case string(domain.CoverageFlagged):
		return domain.CoverageFlagged, nil
	case string(domain.CoverageNotApplicable):
		return domain.CoverageNotApplicable, nil
	case string(domain.CoverageRequiresManualVerification):
		return domain.CoverageRequiresManualVerification, nil
	default:
		return "", fmt.Errorf("invalid provider coverage status %q", value)
	}
}

func sourceByPath(sources []promptSource) map[string]promptSource {
	byPath := make(map[string]promptSource, len(sources))
	for _, source := range sources {
		byPath[source.Path] = source
	}
	return byPath
}

func validateCitation(citation providerCitation, manifest Manifest, sources map[string]promptSource) (domain.SourceRange, error) {
	if citation.Path == "" || citation.StartLine < 1 || citation.EndLine < citation.StartLine || strings.TrimSpace(citation.Quote) == "" {
		return domain.SourceRange{}, errors.New("citation requires path, line range, and quote")
	}
	source, exists := sources[citation.Path]
	if !exists {
		return domain.SourceRange{}, errors.New("citation path was not included in provider context")
	}
	if citation.StartLine < source.StartLine || citation.EndLine > source.EndLine {
		return domain.SourceRange{}, errors.New("citation range was not included in provider context")
	}
	if !containsExactLineQuote(source.Text, citation.StartLine-source.StartLine+1, citation.EndLine-source.StartLine+1, citation.Quote) {
		return domain.SourceRange{}, errors.New("citation quote does not match the source snapshot range")
	}
	for _, entry := range manifest.Inspected {
		if entry.Range.Path == citation.Path && citation.StartLine >= entry.Range.StartLine && citation.EndLine <= entry.Range.EndLine {
			return domain.SourceRange{Path: citation.Path, StartLine: citation.StartLine, EndLine: citation.EndLine}, nil
		}
	}
	return domain.SourceRange{}, errors.New("citation range is absent from inspected manifest")
}

func containsExactLineQuote(content string, start, end int, quote string) bool {
	lines := strings.Split(content, "\n")
	if start < 1 || end > len(lines) || start > end {
		return false
	}
	return strings.Contains(strings.Join(lines[start-1:end], "\n"), quote)
}

func appendUniqueRange(ranges []domain.SourceRange, rangeValue domain.SourceRange) []domain.SourceRange {
	for _, existing := range ranges {
		if existing == rangeValue {
			return ranges
		}
	}
	return append(ranges, rangeValue)
}

func validOmittedPaths(values []string, manifest Manifest) []string {
	known := make(map[string]bool, len(manifest.Omissions))
	for _, omission := range manifest.Omissions {
		if omission.Path != "" {
			known[omission.Path] = true
		}
	}
	output := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if !known[value] || seen[value] {
			continue
		}
		seen[value] = true
		output = append(output, value)
	}
	return output
}

func adapterResultIDs(results []domain.AdapterResult) []string {
	ids := make([]string, 0, len(results))
	seen := make(map[string]bool, len(results))
	for _, result := range results {
		if result.ResultID != "" && !seen[result.ResultID] {
			ids, seen[result.ResultID] = append(ids, result.ResultID), true
		}
	}
	return ids
}

func validatedObservation(request ReviewRequest, finding providerFinding, manifest Manifest, sources map[string]promptSource, model, promptVersion string, redactor secretRedactor) (domain.Observation, error) {
	if strings.TrimSpace(finding.Mechanism) == "" {
		return domain.Observation{}, errors.New("finding has no mechanism")
	}
	if strings.TrimSpace(finding.Uncertainty) == "" {
		return domain.Observation{}, errors.New("finding has no uncertainty")
	}
	if strings.TrimSpace(finding.Reproduction) == "" {
		return domain.Observation{}, errors.New("finding has no reproduction route")
	}
	confidence := domain.Confidence(finding.Confidence)
	if !confidence.Valid() {
		return domain.Observation{}, errors.New("finding has invalid confidence")
	}
	evidenceRange, err := validateCitation(finding.Evidence, manifest, sources)
	if err != nil {
		return domain.Observation{}, fmt.Errorf("finding evidence: %w", err)
	}
	severity, err := severityFor(request.Family, finding.Severity)
	if err != nil {
		return domain.Observation{}, err
	}
	quote := redactor.redact(finding.Evidence.Quote)
	mechanism := redactor.redact(strings.TrimSpace(finding.Mechanism))
	uncertainty := redactor.redact(strings.TrimSpace(finding.Uncertainty))
	identity := sha256.Sum256([]byte(strings.Join([]string{request.RunID, request.Feature.FeatureID, string(request.Family), request.Category, mechanism, evidenceRange.Path, fmt.Sprint(evidenceRange.StartLine), fmt.Sprint(evidenceRange.EndLine)}, "\x00")))
	observation := domain.Observation{
		ObservationID: hex.EncodeToString(identity[:]), RunID: request.RunID, Source: "contextual-provider", SourceFindingID: hex.EncodeToString(identity[:]),
		Family: request.Family, OriginalCategory: request.Category, RuleVersion: promptVersion, ModelVersion: model,
		SourceSeverity: severity, EffectiveSeverity: severity, Confidence: confidence, Mechanism: mechanism, Uncertainty: uncertainty,
		Evidence: []domain.Evidence{{Location: evidenceRange, RedactedQuote: quote}}, Validated: true, ObservedAt: time.Now().UTC(),
	}
	if err := observation.Validate(); err != nil {
		return domain.Observation{}, fmt.Errorf("validated contextual observation: %w", err)
	}
	return observation, nil
}

func severityFor(family domain.Family, raw json.RawMessage) (domain.Severity, error) {
	if len(raw) == 0 {
		return "", errors.New("finding has no severity dimensions")
	}
	switch family {
	case domain.FamilySecurity:
		var value struct {
			Impact     int `json:"impact"`
			Likelihood int `json:"likelihood"`
		}
		if err := decodeStrictJSON(raw, &value); err != nil || value.Impact < 1 || value.Impact > 5 || value.Likelihood < 1 || value.Likelihood > 5 {
			return "", errors.New("security severity requires impact and likelihood from 1 through 5")
		}
		score := value.Impact * value.Likelihood
		switch {
		case score == 1:
			return domain.SeverityInformational, nil
		case score <= 5:
			return domain.SeverityLow, nil
		case score <= 11:
			return domain.SeverityMedium, nil
		case score <= 19:
			return domain.SeverityHigh, nil
		default:
			return domain.SeverityCritical, nil
		}
	case domain.FamilyArchitecture:
		var value struct {
			ChangeRisk string `json:"changeRisk"`
			Boundary   string `json:"boundary"`
		}
		if err := decodeStrictJSON(raw, &value); err != nil {
			return "", errors.New("architecture severity is invalid")
		}
		return matrixSeverity(value.ChangeRisk, value.Boundary, architectureMatrix, architectureColumns)
	case domain.FamilyQuality:
		var value struct {
			ChangeRisk string `json:"changeRisk"`
			Scope      string `json:"scope"`
		}
		if err := decodeStrictJSON(raw, &value); err != nil {
			return "", errors.New("quality severity is invalid")
		}
		return matrixSeverity(value.ChangeRisk, value.Scope, qualityMatrix, qualityColumns)
	case domain.FamilyPerformance:
		var value struct {
			OperationalRisk string `json:"operationalRisk"`
			Frequency       string `json:"frequency"`
		}
		if err := decodeStrictJSON(raw, &value); err != nil {
			return "", errors.New("performance severity is invalid")
		}
		return matrixSeverity(value.OperationalRisk, value.Frequency, performanceMatrix, performanceColumns)
	default:
		return "", errors.New("contextual finding family is unsupported")
	}
}

func decodeStrictJSON(raw json.RawMessage, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("trailing JSON")
	}
	return nil
}

var architectureMatrix = map[string][]domain.Severity{
	"blocking":      {domain.SeverityCritical, domain.SeverityCritical, domain.SeverityHigh, domain.SeverityMedium, domain.SeverityInformational},
	"major":         {domain.SeverityHigh, domain.SeverityHigh, domain.SeverityHigh, domain.SeverityMedium, domain.SeverityInformational},
	"meaningful":    {domain.SeverityMedium, domain.SeverityMedium, domain.SeverityMedium, domain.SeverityMedium, domain.SeverityInformational},
	"minor":         {domain.SeverityLow, domain.SeverityLow, domain.SeverityLow, domain.SeverityLow, domain.SeverityInformational},
	"informational": {domain.SeverityInformational, domain.SeverityInformational, domain.SeverityInformational, domain.SeverityInformational, domain.SeverityInformational},
}
var qualityMatrix = map[string][]domain.Severity{
	"blocking":      {domain.SeverityCritical, domain.SeverityCritical, domain.SeverityHigh, domain.SeverityMedium, domain.SeverityLow},
	"major":         {domain.SeverityHigh, domain.SeverityHigh, domain.SeverityHigh, domain.SeverityMedium, domain.SeverityLow},
	"meaningful":    {domain.SeverityMedium, domain.SeverityMedium, domain.SeverityMedium, domain.SeverityMedium, domain.SeverityLow},
	"minor":         {domain.SeverityLow, domain.SeverityLow, domain.SeverityLow, domain.SeverityLow, domain.SeverityLow},
	"informational": {domain.SeverityInformational, domain.SeverityInformational, domain.SeverityInformational, domain.SeverityInformational, domain.SeverityInformational},
}
var performanceMatrix = map[string][]domain.Severity{
	"unacceptable":  {domain.SeverityCritical, domain.SeverityCritical, domain.SeverityHigh, domain.SeverityMedium, domain.SeverityInformational},
	"major":         {domain.SeverityHigh, domain.SeverityHigh, domain.SeverityMedium, domain.SeverityMedium, domain.SeverityInformational},
	"meaningful":    {domain.SeverityMedium, domain.SeverityMedium, domain.SeverityMedium, domain.SeverityLow, domain.SeverityInformational},
	"minor":         {domain.SeverityLow, domain.SeverityLow, domain.SeverityLow, domain.SeverityLow, domain.SeverityInformational},
	"informational": {domain.SeverityInformational, domain.SeverityInformational, domain.SeverityInformational, domain.SeverityInformational, domain.SeverityInformational},
}

var architectureColumns = map[string]int{
	"systemic_or_critical_boundary":    0,
	"multi_area_or_important_boundary": 1,
	"component":                        2,
	"local":                            3,
	"none":                             4,
}

var qualityColumns = map[string]int{
	"systemic":     0,
	"multi_module": 1,
	"component":    2,
	"local":        3,
	"isolated":     4,
}

var performanceColumns = map[string]int{
	"systemic": 0,
	"common":   1,
	"bounded":  2,
	"rare":     3,
	"none":     4,
}

func matrixSeverity(row, column string, matrix map[string][]domain.Severity, columns map[string]int) (domain.Severity, error) {
	values, exists := matrix[row]
	index, known := columns[column]
	if !exists || !known {
		return "", errors.New("finding has invalid closed severity dimensions")
	}
	return values[index], nil
}

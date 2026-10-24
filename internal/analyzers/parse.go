package analyzers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"scout/internal/domain"
)

type parsedOutput struct {
	success          bool
	processed        map[string]struct{}
	processingErrors map[string]struct{}
	findings         []parsedFinding
}

type parsedFinding struct {
	id          string
	ruleID      string
	severity    domain.Severity
	confidence  domain.Confidence
	mechanism   string
	uncertainty string
	path        string
	startLine   int
	endLine     int
	startByte   int
	endByte     int
	quote       string
	startColumn int
	endColumn   int
}

type adapterJSON struct {
	SchemaVersion    string                `json:"schemaVersion"`
	Success          *bool                 `json:"success"`
	ProcessedInputs  []string              `json:"processedInputs"`
	ProcessingErrors []adapterProcessError `json:"processingErrors"`
	Findings         []adapterFinding      `json:"findings"`
}

type adapterProcessError struct {
	Path string `json:"path"`
	Code string `json:"code"`
}

type adapterFinding struct {
	ID          string            `json:"id"`
	RuleID      string            `json:"ruleId"`
	Severity    domain.Severity   `json:"severity"`
	Confidence  domain.Confidence `json:"confidence"`
	Mechanism   string            `json:"mechanism"`
	Uncertainty string            `json:"uncertainty,omitempty"`
	Path        string            `json:"path"`
	StartLine   int               `json:"startLine"`
	EndLine     int               `json:"endLine"`
	StartByte   int               `json:"startByte,omitempty"`
	EndByte     int               `json:"endByte,omitempty"`
	Quote       string            `json:"quote,omitempty"`
}

func parseOutput(spec Spec, raw []byte, root string) (parsedOutput, error) {
	switch spec.Output.Format {
	case OutputJSON:
		return parseAdapterJSON(spec, raw)
	case OutputSARIF:
		return parseSARIF(spec, raw, root)
	case OutputSemgrepJSON:
		return parseSemgrepJSON(spec, raw)
	default:
		return parsedOutput{}, errors.New("unsupported output format")
	}
}

func parseAdapterJSON(spec Spec, raw []byte) (parsedOutput, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var value adapterJSON
	if err := decoder.Decode(&value); err != nil {
		return parsedOutput{}, errors.New("malformed adapter JSON output")
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return parsedOutput{}, errors.New("malformed adapter JSON output")
	}
	if value.SchemaVersion != spec.Output.SchemaVersion || value.Success == nil {
		return parsedOutput{}, errors.New("unknown adapter JSON schema")
	}
	result := parsedOutput{success: *value.Success, processed: make(map[string]struct{}), processingErrors: make(map[string]struct{})}
	for _, path := range value.ProcessedInputs {
		if err := validRelativePath(path); err != nil {
			return parsedOutput{}, errors.New("malformed processed input path")
		}
		result.processed[path] = struct{}{}
	}
	for _, processError := range value.ProcessingErrors {
		if err := validRelativePath(processError.Path); err != nil || !safeName(processError.Code) {
			return parsedOutput{}, errors.New("malformed processing error")
		}
		result.processingErrors[processError.Path] = struct{}{}
	}
	for _, finding := range value.Findings {
		parsed, err := parseAdapterFinding(finding)
		if err != nil {
			return parsedOutput{}, err
		}
		if !spec.allowsRuleID(parsed.ruleID) {
			return parsedOutput{}, errors.New("adapter emitted an undeclared rule ID")
		}
		result.findings = append(result.findings, parsed)
	}
	return result, nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("unexpected trailing JSON")
	}
	return nil
}

func parseAdapterFinding(value adapterFinding) (parsedFinding, error) {
	if !safeName(value.ID) || !value.Severity.Valid() || value.Severity == domain.SeverityInformational || !value.Confidence.Valid() || value.Mechanism == "" || validRelativePath(value.Path) != nil || value.StartLine <= 0 || value.EndLine < value.StartLine || (value.StartByte == 0) != (value.EndByte == 0) || value.StartByte < 0 || value.EndByte < value.StartByte {
		return parsedFinding{}, errors.New("malformed adapter finding")
	}
	ruleID, err := sanitizeIdentifier(value.RuleID)
	if err != nil {
		return parsedFinding{}, errors.New("malformed adapter finding rule")
	}
	return parsedFinding{id: value.ID, ruleID: ruleID, severity: value.Severity, confidence: value.Confidence, mechanism: value.Mechanism, uncertainty: value.Uncertainty, path: value.Path, startLine: value.StartLine, endLine: value.EndLine, startByte: value.StartByte, endByte: value.EndByte, quote: value.Quote}, nil
}

type sarifLog struct {
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool        sarifTool                  `json:"tool"`
	Results     []sarifResult              `json:"results"`
	Invocations []sarifInvocation          `json:"invocations"`
	Properties  map[string]json.RawMessage `json:"properties"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}
type sarifDriver struct {
	Name string `json:"name"`
}
type sarifInvocation struct {
	ExecutionSuccessful        *bool             `json:"executionSuccessful"`
	ToolExecutionNotifications []json.RawMessage `json:"toolExecutionNotifications"`
}
type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifMessage    `json:"message"`
	Locations []sarifLocation `json:"locations"`
	GUID      string          `json:"guid"`
}
type sarifMessage struct {
	Text string `json:"text"`
}
type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}
type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           sarifRegion           `json:"region"`
}
type sarifArtifactLocation struct {
	URI       string `json:"uri"`
	URIBaseID string `json:"uriBaseId"`
}
type sarifRegion struct {
	StartLine   int `json:"startLine"`
	EndLine     int `json:"endLine"`
	CharOffset  int `json:"charOffset"`
	StartColumn int `json:"startColumn"`
	EndColumn   int `json:"endColumn"`
	CharLength  int `json:"charLength"`
}

func parseSARIF(spec Spec, raw []byte, root string) (parsedOutput, error) {
	var log sarifLog
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(&log); err != nil || ensureJSONEOF(decoder) != nil || log.Version != spec.Output.SchemaVersion || len(log.Runs) == 0 {
		return parsedOutput{}, errors.New("malformed or unknown SARIF schema")
	}
	// SARIF alone cannot attest inputs. Runner additionally requires the
	// spec's attestation strategy to prove them before completion.
	externallyAttested := spec.Attestation != AttestOutput
	result := parsedOutput{success: spec.Output.ProcessedInputsProperty != "" || externallyAttested, processed: make(map[string]struct{}), processingErrors: make(map[string]struct{})}
	hasInvocation := false
	for _, run := range log.Runs {
		if run.Tool.Driver.Name == "" {
			return parsedOutput{}, errors.New("malformed SARIF tool record")
		}
		for _, invocation := range run.Invocations {
			hasInvocation = true
			if invocation.ExecutionSuccessful == nil || !*invocation.ExecutionSuccessful || len(invocation.ToolExecutionNotifications) != 0 {
				result.success = false
			}
		}
		if spec.Output.ProcessedInputsProperty != "" {
			rawPaths, found := run.Properties[spec.Output.ProcessedInputsProperty]
			if !found {
				result.success = false
			} else {
				var paths []string
				if err := json.Unmarshal(rawPaths, &paths); err != nil {
					return parsedOutput{}, errors.New("malformed SARIF processed inputs")
				}
				for _, path := range paths {
					if err := validRelativePath(path); err != nil {
						return parsedOutput{}, errors.New("malformed SARIF processed input path")
					}
					result.processed[path] = struct{}{}
				}
			}
		}
		for _, item := range run.Results {
			finding, err := parseSARIFFinding(spec, item, root)
			if err != nil {
				return parsedOutput{}, err
			}
			result.findings = append(result.findings, finding)
		}
	}
	if !hasInvocation && !externallyAttested {
		result.success = false
	}
	return result, nil
}

func parseSARIFFinding(spec Spec, value sarifResult, root string) (parsedFinding, error) {
	if len(value.Locations) != 1 || value.Message.Text == "" {
		return parsedFinding{}, errors.New("malformed SARIF result")
	}
	ruleID, err := sanitizeIdentifier(value.RuleID)
	if err != nil {
		return parsedFinding{}, errors.New("malformed SARIF rule ID")
	}
	if !spec.allowsRuleID(ruleID) {
		return parsedFinding{}, errors.New("SARIF emitted an undeclared rule ID")
	}
	path, err := sarifPath(value.Locations[0].PhysicalLocation.ArtifactLocation, root)
	if err != nil {
		return parsedFinding{}, errors.New("malformed SARIF location")
	}
	region := value.Locations[0].PhysicalLocation.Region
	if err := validRelativePath(path); err != nil || region.StartLine <= 0 {
		return parsedFinding{}, errors.New("malformed SARIF location")
	}
	endLine := region.EndLine
	if endLine == 0 {
		endLine = region.StartLine
	}
	if endLine < region.StartLine || region.CharOffset < 0 || region.CharLength < 0 {
		return parsedFinding{}, errors.New("malformed SARIF range")
	}
	severity := domain.SeverityMedium
	switch value.Level {
	case "note", "none":
		severity = domain.SeverityLow
	case "warning", "":
	case "error":
		severity = domain.SeverityHigh
	default:
		return parsedFinding{}, errors.New("unknown SARIF result level")
	}
	id := value.GUID
	if id == "" {
		id = digestID(ruleID, path, fmt.Sprint(region.StartLine), value.Message.Text)
	}
	if !safeName(id) {
		return parsedFinding{}, errors.New("unsafe SARIF result ID")
	}
	return parsedFinding{id: id, ruleID: ruleID, severity: severity, confidence: domain.ConfidenceMedium, mechanism: value.Message.Text, path: path, startLine: region.StartLine, endLine: endLine, startColumn: region.StartColumn, endColumn: region.EndColumn}, nil
}

// SARIF producers use repository-relative paths, Semgrep's %SRCROOT% base ID,
// absolute file URIs, or Gitleaks' absolute URI paths without a scheme.
// Resolve only paths contained in this scan root; verify source afterward.
func sarifPath(location sarifArtifactLocation, root string) (string, error) {
	if location.URIBaseID != "" && location.URIBaseID != "%SRCROOT%" {
		return "", errors.New("unsupported SARIF URI base")
	}
	uri, err := url.Parse(location.URI)
	if err != nil || uri.RawQuery != "" || uri.Fragment != "" || uri.User != nil {
		return "", errors.New("malformed SARIF URI")
	}
	var path string
	switch uri.Scheme {
	case "":
		if uri.Host != "" || uri.Opaque != "" {
			return "", errors.New("unsupported SARIF URI")
		}
		path = strings.TrimPrefix(uri.Path, "./")
		if filepath.IsAbs(path) {
			if location.URIBaseID != "" {
				return "", errors.New("absolute SARIF path has a base ID")
			}
			path, err = filepath.Rel(root, path)
			if err != nil {
				return "", err
			}
		}
	case "file":
		if location.URIBaseID != "" || (uri.Host != "" && uri.Host != "localhost") || uri.Opaque != "" || !filepath.IsAbs(uri.Path) {
			return "", errors.New("unsupported absolute SARIF URI")
		}
		path, err = filepath.Rel(root, uri.Path)
		if err != nil {
			return "", err
		}
	default:
		return "", errors.New("unsupported SARIF URI scheme")
	}
	if err := validRelativePath(path); err != nil {
		return "", err
	}
	return path, nil
}

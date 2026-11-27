package analyzers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"scout/internal/domain"
)

type semgrepOutput struct {
	Version string            `json:"version"`
	Results []semgrepResult   `json:"results"`
	Errors  []json.RawMessage `json:"errors"`
	Paths   struct {
		Scanned []string `json:"scanned"`
	} `json:"paths"`
}

type semgrepResult struct {
	CheckID string `json:"check_id"`
	Path    string `json:"path"`
	Start   struct {
		Line int `json:"line"`
	} `json:"start"`
	End struct {
		Line int `json:"line"`
	} `json:"end"`
	Extra struct {
		Message  string `json:"message"`
		Severity string `json:"severity"`
	} `json:"extra"`
}

// parseSemgrepJSON accepts only the pinned scan schema. Semgrep's paths.scanned
// is completion evidence for clean files as well as files with findings.
func parseSemgrepJSON(spec Spec, raw []byte) (parsedOutput, error) {
	output, err := decodeSemgrepJSON(spec, raw)
	if err != nil {
		return parsedOutput{}, err
	}
	parsed := parsedOutput{success: len(output.Errors) == 0, processed: make(map[string]struct{}, len(output.Paths.Scanned)), processingErrors: make(map[string]struct{})}
	for _, path := range output.Paths.Scanned {
		if err := validRelativePath(path); err != nil {
			return parsedOutput{}, errors.New("invalid Semgrep scanned path")
		}
		parsed.processed[path] = struct{}{}
	}
	for _, item := range output.Results {
		finding, err := parseSemgrepFinding(spec, item)
		if err != nil {
			return parsedOutput{}, err
		}
		if _, scanned := parsed.processed[finding.path]; !scanned {
			parsed.success = false
		}
		parsed.findings = append(parsed.findings, finding)
	}
	return parsed, nil
}

func decodeSemgrepJSON(spec Spec, raw []byte) (semgrepOutput, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	var output semgrepOutput
	if err := decoder.Decode(&output); err != nil || ensureJSONEOF(decoder) != nil {
		return semgrepOutput{}, errors.New("malformed or unknown Semgrep JSON schema")
	}
	if output.Version != spec.Output.SchemaVersion || output.Results == nil || output.Errors == nil || output.Paths.Scanned == nil {
		return semgrepOutput{}, errors.New("malformed or unknown Semgrep JSON schema")
	}
	return output, nil
}

func parseSemgrepFinding(spec Spec, item semgrepResult) (parsedFinding, error) {
	path := item.Path
	ruleID, err := sanitizeIdentifier(normalizeSemgrepRuleID(spec, item.CheckID))
	if err != nil || !spec.allowsRuleID(ruleID) || validRelativePath(path) != nil || item.Start.Line <= 0 || item.End.Line < item.Start.Line || item.Extra.Message == "" {
		return parsedFinding{}, errors.New("invalid Semgrep finding")
	}
	severity, err := semgrepSeverity(item.Extra.Severity)
	if err != nil {
		return parsedFinding{}, err
	}
	return parsedFinding{id: digestID(ruleID, path, fmt.Sprint(item.Start.Line), item.Extra.Message), ruleID: ruleID, severity: severity, confidence: domain.ConfidenceMedium, mechanism: item.Extra.Message, path: path, startLine: item.Start.Line, endLine: item.End.Line}, nil
}

func semgrepSeverity(level string) (domain.Severity, error) {
	switch level {
	case "ERROR":
		return domain.SeverityHigh, nil
	case "WARNING":
		return domain.SeverityMedium, nil
	case "INFO":
		return domain.SeverityLow, nil
	default:
		return "", errors.New("unknown Semgrep severity")
	}
}

func normalizeSemgrepRuleID(spec Spec, raw string) string {
	for _, declared := range spec.Rules.RuleIDs {
		if raw == declared || strings.HasSuffix(raw, "."+declared) {
			return declared
		}
	}
	return raw
}

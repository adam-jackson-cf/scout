package analyzers

import (
	"errors"
	"regexp"
	"strings"
)

const gitleaksAllowlistReason = "excluded by the pinned Gitleaks rule pack path allowlist"

// gitleaksGlobalPathAllowlist extracts the `paths` regexes of the pack's
// top-level [allowlist] table. Gitleaks never reads files matching them, so no
// scan can prove their inspection. Rule-level [[rules.allowlists]] tables only
// suppress findings of one rule and are deliberately ignored. The pinned global
// allowlist sets no condition or regexTarget, so a path match alone excludes a
// file; a pack that adds either is rejected rather than misread.
func gitleaksGlobalPathAllowlist(pack []byte) ([]*regexp.Regexp, error) {
	var section []string
	inSection := false
	for _, line := range strings.Split(string(pack), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			inSection = trimmed == "[allowlist]"
			continue
		}
		if inSection {
			section = append(section, trimmed)
		}
	}
	var patterns []*regexp.Regexp
	inPaths := false
	for _, line := range section {
		switch {
		case strings.HasPrefix(line, "condition") || strings.HasPrefix(line, "regexTarget"):
			return nil, errors.New("global allowlist uses unsupported condition or regexTarget")
		case strings.HasPrefix(line, "paths") && strings.HasSuffix(line, "["):
			inPaths = true
		case inPaths && line == "]":
			inPaths = false
		case inPaths && strings.HasPrefix(line, "'''"):
			end := strings.LastIndex(line, "'''")
			if end < 3 {
				return nil, errors.New("malformed global allowlist path")
			}
			compiled, err := regexp.Compile(line[3:end])
			if err != nil {
				return nil, err
			}
			patterns = append(patterns, compiled)
		}
	}
	if len(patterns) == 0 {
		return nil, errors.New("global path allowlist is empty")
	}
	return patterns, nil
}

func pathAllowlisted(patterns []*regexp.Regexp, path string) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(path) {
			return true
		}
	}
	return false
}

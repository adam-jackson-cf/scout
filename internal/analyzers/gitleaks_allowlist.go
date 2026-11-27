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
	patterns, err := allowlistPathPatterns(globalAllowlistSection(pack))
	if err != nil {
		return nil, err
	}
	if len(patterns) == 0 {
		return nil, errors.New("global path allowlist is empty")
	}
	return patterns, nil
}

func globalAllowlistSection(pack []byte) []string {
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
	return section
}

func allowlistPathPatterns(section []string) ([]*regexp.Regexp, error) {
	var patterns []*regexp.Regexp
	inPaths := false
	for _, line := range section {
		switch {
		case unsupportedAllowlistKey(line):
			return nil, errors.New("global allowlist uses unsupported condition or regexTarget")
		case opensPathsArray(line):
			inPaths = true
		case !inPaths:
		case line == "]":
			inPaths = false
		case strings.HasPrefix(line, "'''"):
			compiled, err := allowlistPathPattern(line)
			if err != nil {
				return nil, err
			}
			patterns = append(patterns, compiled)
		}
	}
	return patterns, nil
}

func unsupportedAllowlistKey(line string) bool {
	return strings.HasPrefix(line, "condition") || strings.HasPrefix(line, "regexTarget")
}

func opensPathsArray(line string) bool {
	return strings.HasPrefix(line, "paths") && strings.HasSuffix(line, "[")
}

func allowlistPathPattern(line string) (*regexp.Regexp, error) {
	end := strings.LastIndex(line, "'''")
	if end < 3 {
		return nil, errors.New("malformed global allowlist path")
	}
	return regexp.Compile(line[3:end])
}

func pathAllowlisted(patterns []*regexp.Regexp, path string) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(path) {
			return true
		}
	}
	return false
}

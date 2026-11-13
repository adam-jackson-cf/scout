// Package redact owns credential detection and redaction. Every package that
// persists, prints or emits text that may carry a credential uses it, so the
// rules change in one place.
package redact

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	privateKey    = regexp.MustCompile(`(?s)-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----.+?-----END [A-Z0-9 ]*PRIVATE KEY-----`)
	assignment    = regexp.MustCompile(`(?i)((?:api[_-]?key|access[_-]?key|secret|token|password|passwd|credential|client[_-]?secret)["']?\s*[:=]\s*["']?)[^\s"']+`)
	urlCredential = regexp.MustCompile(`://[^\s/@:]+:[^\s/@]+@`)
	jwt           = regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\b`)
	knownToken    = regexp.MustCompile(`\b(?:AKIA[0-9A-Z]{16}|gh[pousr]_[A-Za-z0-9_]{20,}|xox[baprs]-[A-Za-z0-9-]{20,})\b`)
	longToken     = regexp.MustCompile(`\b[A-Za-z0-9+/_=-]{32,}\b`)
)

// ErrUnsafe is returned for text that is not valid UTF-8 or contains NUL.
var ErrUnsafe = errors.New("unsafe text")

// Contains reports whether value carries a credential-shaped secret or NUL.
// It is a refusal test for persistence, so it ignores the long-token heuristic
// that Text applies when masking free text.
func Contains(value string) bool {
	return strings.ContainsRune(value, '\x00') || privateKey.MatchString(value) || assignment.MatchString(value) ||
		urlCredential.MatchString(value) || jwt.MatchString(value) || knownToken.MatchString(value)
}

// Text masks credential-shaped values and strips control characters other than
// newline and tab. It never falls back to returning the original string.
func Text(value string) (string, error) {
	if !utf8.ValidString(value) || strings.ContainsRune(value, '\x00') {
		return "", ErrUnsafe
	}
	value = privateKey.ReplaceAllString(value, "[REDACTED PRIVATE KEY]")
	value = assignment.ReplaceAllString(value, "$1[REDACTED]")
	value = urlCredential.ReplaceAllString(value, "://[REDACTED]@")
	value = jwt.ReplaceAllString(value, "[REDACTED JWT]")
	value = knownToken.ReplaceAllString(value, "[REDACTED TOKEN]")
	value = longToken.ReplaceAllString(value, "[REDACTED]")
	return strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, value), nil
}

// Truncate cuts value to at most maximum bytes on a UTF-8 boundary and marks
// the cut with an ellipsis.
func Truncate(value string, maximum int) string {
	if len(value) <= maximum {
		return value
	}
	value = value[:maximum]
	for !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value + "…"
}

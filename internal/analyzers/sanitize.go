package analyzers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	maxEvidenceText = 4096
	maxIdentifier   = 256
)

var (
	privateKeyPattern    = regexp.MustCompile(`(?s)-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----.+?-----END [A-Z0-9 ]*PRIVATE KEY-----`)
	assignmentPattern    = regexp.MustCompile(`(?i)(\b(?:api[_-]?key|access[_-]?key|secret|token|password|passwd|credential|client[_-]?secret)\b\s*[:=]\s*["']?)[^\s"']+`)
	urlCredentialPattern = regexp.MustCompile(`://[^\s/@:]+:[^\s/@]+@`)
	jwtPattern           = regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\b`)
	knownTokenPattern    = regexp.MustCompile(`\b(?:AKIA[0-9A-Z]{16}|gh[pousr]_[A-Za-z0-9_]{20,}|xox[baprs]-[A-Za-z0-9-]{20,})\b`)
	longTokenPattern     = regexp.MustCompile(`\b[A-Za-z0-9+/_=-]{32,}\b`)
	identifierPattern    = regexp.MustCompile(`^[A-Za-z0-9._:/-]{1,256}$`)
)

// sanitizeText removes credential-bearing material before data can leave the
// subprocess boundary. It never falls back to returning the original string.
func sanitizeText(value string, maximum int) (string, error) {
	if maximum <= 0 || !utf8.ValidString(value) || strings.ContainsRune(value, '\x00') {
		return "", errors.New("unsafe diagnostic text")
	}
	value = privateKeyPattern.ReplaceAllString(value, "[REDACTED PRIVATE KEY]")
	value = assignmentPattern.ReplaceAllString(value, "$1[REDACTED]")
	value = urlCredentialPattern.ReplaceAllString(value, "://[REDACTED]@")
	value = jwtPattern.ReplaceAllString(value, "[REDACTED JWT]")
	value = knownTokenPattern.ReplaceAllString(value, "[REDACTED TOKEN]")
	value = longTokenPattern.ReplaceAllString(value, "[REDACTED]")
	value = strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, value)
	if len(value) > maximum {
		value = value[:maximum]
		for !utf8.ValidString(value) {
			value = value[:len(value)-1]
		}
		value += "…"
	}
	return value, nil
}

func sanitizeIdentifier(value string) (string, error) {
	if !identifierPattern.MatchString(value) {
		return "", errors.New("unsafe diagnostic identifier")
	}
	return value, nil
}

func digestID(parts ...string) string {
	hash := sha256.New()
	for _, part := range parts {
		hash.Write([]byte(part))
		hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil))
}

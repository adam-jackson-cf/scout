package analyzers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"

	"scout/internal/redact"
)

const (
	maxEvidenceText = 4096
	maxIdentifier   = 256
)

var identifierPattern = regexp.MustCompile(`^[A-Za-z0-9._:/-]{1,256}$`)

// sanitizeText removes credential-bearing material before data can leave the
// subprocess boundary. It never falls back to returning the original string.
func sanitizeText(value string, maximum int) (string, error) {
	if maximum <= 0 {
		return "", errors.New("unsafe diagnostic text")
	}
	value, err := redact.Text(value)
	if err != nil {
		return "", errors.New("unsafe diagnostic text")
	}
	return redact.Truncate(value, maximum), nil
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

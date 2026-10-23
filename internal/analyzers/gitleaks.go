package analyzers

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Version-pinned Gitleaks 8.30.1 reports an exact decimal byte count in this
// progress line. One accepted input, valid SARIF, and an exact count are all
// required; exit status or a finding alone never establishes coverage.
var gitleaksScannedBytes = regexp.MustCompile(`scanned ~([0-9]+) bytes \(`)

func attestGitleaksInput(parsed *parsedOutput, inputs map[string]sourceInput, stderr []byte) {
	if len(inputs) != 1 {
		parsed.success = false
		return
	}
	matches := gitleaksScannedBytes.FindAllSubmatch(stderr, 2)
	if len(matches) != 1 {
		parsed.success = false
		return
	}
	scanned, err := strconv.ParseInt(string(matches[0][1]), 10, 64)
	if err != nil {
		parsed.success = false
		return
	}
	for path, input := range inputs {
		if scanned != input.manifest.Bytes {
			parsed.success = false
			return
		}
		parsed.processed[path] = struct{}{}
	}
}

// secretDigest returns the hex SHA-256 of the raw secret text a Gitleaks
// finding points at. Gitleaks runs with --redact=100, so the value is re-read
// from the validated source file, hashed in memory and dropped: it is never
// stored, logged or returned. Gitleaks 8.30.1 reports 1-based columns shifted
// by one (start = offset+2, end = last offset+2); the slice below matches that
// and only needs to be deterministic. An unusable range hashes the whole line.
// It returns "" when the line cannot be read, so the quote fingerprint applies.
func secretDigest(root string, finding parsedFinding) string {
	if validRelativePath(finding.path) != nil || finding.startLine <= 0 || finding.endLine != finding.startLine {
		return ""
	}
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(finding.path)))
	if err != nil {
		return ""
	}
	lines := strings.Split(string(content), "\n")
	if finding.startLine > len(lines) {
		return ""
	}
	line := []rune(strings.TrimSuffix(lines[finding.startLine-1], "\r"))
	value := line
	if start, end := finding.startColumn-2, finding.endColumn-1; start >= 0 && end > start && end <= len(line) {
		value = line[start:end]
	}
	sum := sha256.Sum256([]byte("scout-secret-digest-v1\x00" + string(value)))
	return hex.EncodeToString(sum[:])
}

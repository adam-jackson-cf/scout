package analyzers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Version-pinned Gitleaks 8.30.1 `dir` takes exactly one path (extra arguments
// are ignored and the working directory is scanned instead), so a batch is a
// private staged directory of byte-verified copies. At --log-level trace it
// logs one line per file it reads, "scanning path path=<abs path>", and
// "skipping empty file path=<abs path>" for empty files, then one total
// "scanned ~N bytes (N bytes)". A file is proven only when it has exactly one
// matching event and the scanned total equals the summed size of the proven
// files; exit status, SARIF, or a finding alone never establishes coverage.
// Files that are not proven are rescanned one per process by the runner.
var (
	gitleaksScannedBytes = regexp.MustCompile(`scanned ~([0-9]+) bytes \(`)
	gitleaksScanEvent    = regexp.MustCompile(`^[0-9]{1,2}:[0-9]{2}(?:AM|PM) (?:TRC scanning path|DBG skipping empty file) path=(.*)$`)
)

// gitleaksEvents counts per-file events keyed by path relative to stageDir.
// unambiguous is false when any event names a path outside stageDir or that
// cannot be decoded, which voids the whole batch proof.
func gitleaksEvents(stderr []byte, stageDir string) (scanned, empty map[string]int, unambiguous bool) {
	scanned, empty, unambiguous = make(map[string]int), make(map[string]int), true
	for _, line := range bytes.Split(stderr, []byte{'\n'}) {
		line = bytes.TrimSuffix(line, []byte{'\r'})
		match := gitleaksScanEvent.FindSubmatch(line)
		if match == nil {
			continue
		}
		path := string(match[1])
		if strings.HasPrefix(path, `"`) {
			unquoted, err := strconv.Unquote(path)
			if err != nil {
				unambiguous = false
				continue
			}
			path = unquoted
		}
		relative, err := filepath.Rel(stageDir, path)
		if err != nil || validRelativePath(filepath.ToSlash(relative)) != nil {
			unambiguous = false
			continue
		}
		relative = filepath.ToSlash(relative)
		if bytes.Contains(line, []byte("skipping empty file")) {
			empty[relative]++
		} else {
			scanned[relative]++
		}
	}
	return scanned, empty, unambiguous
}

// attestGitleaksInput records in parsed.processed the staged inputs whose exact
// bytes Gitleaks provably scanned. Missing or repeated totals, unknown paths or
// a byte mismatch prove nothing; the caller then rescans those inputs singly.
func attestGitleaksInput(parsed *parsedOutput, inputs map[string]sourceInput, stderr []byte, stageDir string) {
	matches := gitleaksScannedBytes.FindAllSubmatch(stderr, 2)
	if len(matches) != 1 {
		parsed.success = false
		return
	}
	total, err := strconv.ParseInt(string(matches[0][1]), 10, 64)
	if err != nil {
		parsed.success = false
		return
	}
	scanned, empty, unambiguous := gitleaksEvents(stderr, stageDir)
	if !unambiguous {
		return
	}
	for path := range scanned {
		if _, known := inputs[path]; !known {
			return
		}
	}
	for path := range empty {
		if _, known := inputs[path]; !known {
			return
		}
	}
	proven := make([]string, 0, len(inputs))
	var sum int64
	for path, input := range inputs {
		size := input.manifest.Bytes
		if (size > 0 && scanned[path] == 1 && empty[path] == 0) || (size == 0 && empty[path] == 1 && scanned[path] == 0) {
			proven = append(proven, path)
			sum += size
		}
	}
	if sum != total {
		return
	}
	for _, path := range proven {
		parsed.processed[path] = struct{}{}
	}
}

// stageGitleaksInputs copies each accepted input into a fresh private
// directory under state at its repository-relative path and returns the staged
// subset. Each copy is hashed while written; an input whose current bytes no
// longer match the manifest digest is not staged and stays unproven.
func stageGitleaksInputs(root, state string, inputs map[string]sourceInput) (string, map[string]sourceInput, error) {
	stageDir, err := os.MkdirTemp(state, "stage-")
	if err != nil {
		return "", nil, err
	}
	staged := make(map[string]sourceInput, len(inputs))
	for path, input := range inputs {
		if err := stageInput(root, stageDir, input); err == nil {
			staged[path] = input
		}
	}
	return stageDir, staged, nil
}

func stageInput(root, stageDir string, input sourceInput) error {
	source, err := os.Open(filepath.Join(root, filepath.FromSlash(input.manifest.Path)))
	if err != nil {
		return err
	}
	defer source.Close()
	destination := filepath.Join(stageDir, filepath.FromSlash(input.manifest.Path))
	if err := os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
		return err
	}
	target, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	hash := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(target, hash), io.LimitReader(source, input.manifest.Bytes+1))
	if closeErr := target.Close(); copyErr == nil {
		copyErr = closeErr
	}
	if copyErr != nil || written != input.manifest.Bytes || hex.EncodeToString(hash.Sum(nil)) != input.manifest.Digest {
		os.Remove(destination)
		return errors.New("staged input does not match its manifest")
	}
	return nil
}

// secretDigest returns the hex SHA-256 of the raw secret text a Gitleaks
// finding points at. Gitleaks runs with --redact=100, so the value is re-read
// from the validated source file, hashed in memory and dropped: it is never
// stored, logged or returned. Gitleaks 8.30.1 reports 1-based columns shifted
// by one (start = offset+2, end = last offset+2); the slice below matches that
// and only needs to be deterministic.
//
// A finding that spans several lines has no single-line value to digest and
// returns "": its identity is the normalized evidence fingerprint
// (domain.AnchorOf). Every other failure — unreadable file, line or column
// range outside the source — is an error, and the caller rejects the finding.
func secretDigest(root string, finding parsedFinding) (string, error) {
	if validRelativePath(finding.path) != nil || finding.startLine <= 0 {
		return "", errors.New("secret finding has an invalid location")
	}
	if finding.endLine != finding.startLine {
		return "", nil
	}
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(finding.path)))
	if err != nil {
		return "", errors.New("secret finding source is unreadable")
	}
	lines := strings.Split(string(content), "\n")
	if finding.startLine > len(lines) {
		return "", errors.New("secret finding line is outside the source")
	}
	line := []rune(strings.TrimSuffix(lines[finding.startLine-1], "\r"))
	start, end := finding.startColumn-2, finding.endColumn-1
	if start < 0 || end <= start || end > len(line) {
		return "", errors.New("secret finding column range is outside the line")
	}
	sum := sha256.Sum256([]byte("scout-secret-digest-v1\x00" + string(line[start:end])))
	return hex.EncodeToString(sum[:]), nil
}

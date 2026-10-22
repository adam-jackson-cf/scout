package analyzers

import (
	"regexp"
	"strconv"
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

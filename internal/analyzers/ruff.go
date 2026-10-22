package analyzers

import (
	"context"
	"path/filepath"
	"strings"
)

// Ruff's SARIF omits clean inputs. The pinned CLI's --show-files mode uses the
// same isolated rule selection as the scan and lists every input it accepts.
// Require an exact match and a successful SARIF scan; neither alone suffices.
func (r *Runner) attestRuffInputs(ctx context.Context, root, state, rulesRoot, binary string, args []string, inputs map[string]sourceInput, parsed *parsedOutput) {
	listingArgs := make([]string, len(args)+1)
	copy(listingArgs, args)
	listingArgs[len(args)] = "--show-files"
	stdout, stderr, err := r.runCommand(ctx, root, state, rulesRoot, binary, listingArgs)
	if err != nil || stdout.exceeded || stderr.exceeded || !ruffInputsMatch(root, inputs, stdout.Bytes()) {
		parsed.success = false
		return
	}
	for path := range inputs {
		parsed.processed[path] = struct{}{}
	}
}

func ruffInputsMatch(root string, inputs map[string]sourceInput, raw []byte) bool {
	if len(inputs) == 0 || len(raw) == 0 || raw[len(raw)-1] != '\n' {
		return false
	}
	expected := make(map[string]struct{}, len(inputs))
	for path := range inputs {
		resolved, err := filepath.EvalSymlinks(filepath.Join(root, path))
		if err != nil {
			return false
		}
		expected[resolved] = struct{}{}
	}
	lines := strings.Split(string(raw[:len(raw)-1]), "\n")
	if len(lines) != len(expected) {
		return false
	}
	for _, line := range lines {
		if !filepath.IsAbs(line) {
			return false
		}
		resolved, err := filepath.EvalSymlinks(line)
		if err != nil {
			return false
		}
		if _, present := expected[resolved]; !present {
			return false
		}
		delete(expected, resolved)
	}
	return len(expected) == 0
}

package analyzers

import (
	"context"
	"path/filepath"
	"strings"
)

// Ruff's SARIF omits clean inputs. The pinned CLI's --show-files mode uses the
// same isolated rule selection as the scan and lists every input it accepts.
// Require an exact match and a successful SARIF scan; neither alone suffices.
func (b *batch) attestRuffInputs(ctx context.Context, parsed *parsedOutput) {
	args, inputs := b.plan.args, b.inputs.byPath
	listingArgs := make([]string, len(args)+1)
	copy(listingArgs, args)
	listingArgs[len(args)] = "--show-files"
	stdout, stderr, err := b.runner.runCommand(ctx, b.root, &b.plan, listingArgs)
	if err != nil || stdout.Exceeded() || stderr.Exceeded() || !ruffInputsMatch(b.root, inputs, stdout.Bytes()) {
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
	expected, ok := resolvedInputs(root, inputs)
	if !ok {
		return false
	}
	lines := strings.Split(string(raw[:len(raw)-1]), "\n")
	if len(lines) != len(expected) {
		return false
	}
	return listingConsumesExactly(lines, expected)
}

func resolvedInputs(root string, inputs map[string]sourceInput) (map[string]struct{}, bool) {
	expected := make(map[string]struct{}, len(inputs))
	for path := range inputs {
		resolved, err := filepath.EvalSymlinks(filepath.Join(root, path))
		if err != nil {
			return nil, false
		}
		expected[resolved] = struct{}{}
	}
	return expected, true
}

func listingConsumesExactly(lines []string, expected map[string]struct{}) bool {
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

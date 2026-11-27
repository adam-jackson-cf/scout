package tooling

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"scout/internal/toolchain"
	"scout/tools/gate/check"
)

type pinnedTool struct {
	bin     string
	key     string
	args    []string
	pattern *regexp.Regexp
}

var (
	golangciTool   = pinnedTool{"golangci-lint", "GOLANGCI_LINT_VERSION", []string{"version", "--short"}, regexp.MustCompile(`^v?(\S+)`)}
	shellcheckTool = pinnedTool{"shellcheck", "SHELLCHECK_VERSION", []string{"--version"}, regexp.MustCompile(`(?m)^version: (\S+)`)}
	shfmtTool      = pinnedTool{"shfmt", "SHFMT_VERSION", []string{"--version"}, regexp.MustCompile(`^v?(\S+)`)}
	prettierTool   = pinnedTool{"prettier", "PRETTIER_VERSION", []string{"--version"}, regexp.MustCompile(`^(\S+)`)}
	ruffTool       = pinnedTool{"ruff", "RUFF_VERSION", []string{"--version"}, regexp.MustCompile(`^ruff (\S+)`)}
	mypyTool       = pinnedTool{"mypy", "MYPY_VERSION", []string{"--version"}, regexp.MustCompile(`^mypy (\S+)`)}
	osvTool        = pinnedTool{"osv-scanner", "OSV_SCANNER_VERSION", []string{"--version"}, regexp.MustCompile(`(?m)^osv-scanner version: (\S+)`)}
)

func resolveTool(r *check.Run, t pinnedTool) (string, error) {
	bin, err := toolPath(r.Root, t.bin)
	if err != nil {
		return "", err
	}
	pins, err := toolchain.Manifest(r.Root)
	if err != nil {
		return "", err
	}
	want := pins[t.key]
	if want == "" {
		return "", fmt.Errorf("%s does not pin %s", toolchain.ManifestPath, t.key)
	}
	res, err := runTool(invocation{dir: r.Root, timeout: 30 * time.Second}, bin, t.args...)
	if err != nil {
		return "", err
	}
	m := t.pattern.FindStringSubmatch(strings.TrimSpace(res.stdout + res.stderr))
	if m == nil || m[1] != want {
		got := strings.TrimSpace(res.stdout)
		return "", fmt.Errorf("%s reports %q, %s pins %s=%s; rerun scripts/bootstrap-tools.sh", bin, got, toolchain.ManifestPath, t.key, want)
	}
	return bin, nil
}

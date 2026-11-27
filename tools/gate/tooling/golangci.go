package tooling

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"

	"scout/tools/gate/check"
)

var golangciFixes = map[string]string{
	"errcheck":    "handle or return the error",
	"govet":       "fix the reported misuse",
	"ineffassign": "remove the ineffective assignment",
	"staticcheck": "apply the change staticcheck describes",
	"unused":      "delete the unused code",
	"gocyclo":     "split the function until gocyclo stays within the .golangci.yml min-complexity",
	"funlen":      "split the function until it stays within the .golangci.yml funlen statements limit",
	"revive":      "group related parameters into a struct",
	"forbidigo":   "use exec.CommandContext with a context deadline",
	"nolintlint":  "name the linter in //nolint:<linter> and explain why after //",
	"gofmt":       "run golangci-lint fmt on the file",
	"goimports":   "run golangci-lint fmt on the file",
}

type golangciReport struct {
	Issues []struct {
		FromLinter string
		Text       string
		Pos        struct {
			Filename string
			Line     int
		}
	}
}

func goPackageDirs(r *check.Run) ([]string, error) {
	files, err := selectFiles(r, byExt(".go"))
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var dirs []string
	for _, f := range files {
		d := "./" + path.Dir(f)
		if path.Dir(f) == "." {
			d = "."
		}
		if !seen[d] {
			seen[d] = true
			dirs = append(dirs, d)
		}
	}
	sort.Strings(dirs)
	return dirs, nil
}

func runGolangci(r *check.Run) ([]check.Finding, error) {
	dirs, err := goPackageDirs(r)
	if err != nil || len(dirs) == 0 {
		return nil, err
	}
	bin, err := resolveTool(r, golangciTool)
	if err != nil {
		return nil, err
	}
	report := filepath.Join(r.TempDir, "golangci.json")
	args := append([]string{"run", "--config", filepath.Join(r.Root, ".golangci.yml"), "--output.json.path=" + report, "--output.text.path=" + filepath.Join(r.TempDir, "golangci.txt"), "--show-stats=false"}, dirs...)
	res, err := runTool(invocation{dir: r.Root, timeout: 10 * time.Minute}, bin, args...)
	if err != nil {
		return nil, err
	}
	return readGolangciReport(res, report)
}

func readGolangciReport(res result, report string) ([]check.Finding, error) {
	if res.exitCode != 0 && res.exitCode != 1 {
		return nil, fmt.Errorf("golangci-lint exited %d: %s", res.exitCode, res.stderr)
	}
	data, err := os.ReadFile(report)
	if err != nil {
		return nil, fmt.Errorf("golangci-lint wrote no report: %w: %s", err, res.stderr)
	}
	findings, err := parseGolangci(data)
	if err != nil {
		return nil, err
	}
	if res.exitCode == 1 && len(findings) == 0 {
		return nil, fmt.Errorf("golangci-lint failed without issues: %s", res.stderr)
	}
	return findings, nil
}

func parseGolangci(data []byte) ([]check.Finding, error) {
	var rep golangciReport
	if err := json.Unmarshal(data, &rep); err != nil {
		return nil, fmt.Errorf("golangci-lint report: %w", err)
	}
	out := make([]check.Finding, 0, len(rep.Issues))
	for _, i := range rep.Issues {
		fix := golangciFixes[i.FromLinter]
		if fix == "" {
			fix = "fix the " + i.FromLinter + " finding"
		}
		out = append(out, check.Finding{File: filepath.ToSlash(i.Pos.Filename), Line: i.Pos.Line, Rule: "GOLANGCI." + i.FromLinter, Problem: i.Text, Fix: fix})
	}
	return out, nil
}

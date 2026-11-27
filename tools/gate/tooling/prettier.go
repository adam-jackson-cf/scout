package tooling

import (
	"fmt"
	"strings"
	"time"

	"scout/tools/gate/check"
)

func runPrettier(r *check.Run) ([]check.Finding, error) {
	files, err := selectFiles(r, byExt(".md", ".yml", ".yaml"))
	if err != nil || len(files) == 0 {
		return nil, err
	}
	bin, err := resolveTool(r, prettierTool)
	if err != nil {
		return nil, err
	}
	res, err := runTool(invocation{dir: r.Root, timeout: 5 * time.Minute}, bin, append([]string{"--list-different", "--no-cache", "--"}, files...)...)
	if err != nil {
		return nil, err
	}
	if res.exitCode != 0 && res.exitCode != 1 {
		return nil, fmt.Errorf("prettier exited %d: %s", res.exitCode, res.stderr)
	}
	var findings []check.Finding
	for _, f := range strings.Fields(res.stdout) {
		findings = append(findings, check.Finding{File: f, Rule: "PRETTIER.format", Problem: "not formatted as .prettierrc.json requires", Fix: "run prettier --write " + f})
	}
	if res.exitCode == 1 && len(findings) == 0 {
		return nil, fmt.Errorf("prettier failed: %s", res.stderr)
	}
	return findings, nil
}

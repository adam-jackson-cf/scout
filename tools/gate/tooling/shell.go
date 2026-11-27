package tooling

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"scout/tools/gate/check"
)

var shellcheckLine = regexp.MustCompile(`^(.+?):(\d+):\d+: (\w+): (.*) \[(SC\d+)\]$`)

func runShellcheck(r *check.Run) ([]check.Finding, error) {
	files, err := selectFiles(r, shellScript)
	if err != nil || len(files) == 0 {
		return nil, err
	}
	bin, err := resolveTool(r, shellcheckTool)
	if err != nil {
		return nil, err
	}
	res, err := runTool(invocation{dir: r.Root, timeout: 5 * time.Minute}, bin, append([]string{"--format=gcc", "--"}, files...)...)
	if err != nil {
		return nil, err
	}
	findings := parseShellcheck(res.stdout)
	if res.exitCode != 0 && len(findings) == 0 {
		return nil, fmt.Errorf("shellcheck exited %d: %s%s", res.exitCode, res.stdout, res.stderr)
	}
	return findings, nil
}

func parseShellcheck(out string) []check.Finding {
	var findings []check.Finding
	for _, line := range strings.Split(out, "\n") {
		m := shellcheckLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		n, _ := strconv.Atoi(m[2])
		findings = append(findings, check.Finding{
			File: m[1], Line: n, Rule: "SHELLCHECK." + m[5],
			Problem: m[3] + ": " + m[4],
			Fix:     "change the code as https://www.shellcheck.net/wiki/" + m[5] + " describes",
		})
	}
	return findings
}

func runShfmt(r *check.Run) ([]check.Finding, error) {
	files, err := selectFiles(r, shellScript)
	if err != nil || len(files) == 0 {
		return nil, err
	}
	bin, err := resolveTool(r, shfmtTool)
	if err != nil {
		return nil, err
	}
	res, err := runTool(invocation{dir: r.Root, timeout: 2 * time.Minute}, bin, append([]string{"--list"}, files...)...)
	if err != nil {
		return nil, err
	}
	if res.exitCode != 0 && strings.TrimSpace(res.stdout) == "" {
		return nil, fmt.Errorf("shfmt exited %d: %s", res.exitCode, res.stderr)
	}
	var findings []check.Finding
	for _, f := range strings.Fields(res.stdout) {
		findings = append(findings, check.Finding{File: f, Rule: "SHFMT.format", Problem: "not formatted as .editorconfig [*.{sh,bash}] requires", Fix: "run shfmt --write " + f})
	}
	return findings, nil
}

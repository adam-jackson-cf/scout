package tooling

import "scout/tools/gate/check"

func Checks() []check.Check {
	return []check.Check{
		{Name: "golangci", Rules: []string{"GOLANGCI.*"}, Run: runGolangci},
		{Name: "shellcheck", Rules: []string{"SHELLCHECK.*"}, Run: runShellcheck},
		{Name: "shfmt", Rules: []string{"SHFMT.*"}, Run: runShfmt},
		{Name: "prettier", Rules: []string{"PRETTIER.version"}, Run: runPrettier},
		{Name: "ruff", Rules: []string{"RUFF.*"}, Run: runRuff},
		{Name: "mypy", Rules: []string{"MYPY.*"}, Run: runMypy},
		{Name: "parity", Rules: []string{"PARITY.*"}, Run: runParity},
		{Name: "dependency-audit", Rules: []string{"SECURITY.dependency-audit"}, Run: runAudit},
		{Name: "coverage-ratchet", Rules: []string{"TEST.coverage-ratchet"}, Run: runRatchet},
	}
}

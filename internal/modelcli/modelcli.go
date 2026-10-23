// Package modelcli finds the logged-in model CLIs (`claude`, `codex`) that
// Scout's reasoning-LLM lanes run through, and builds the credential-free
// environment those subprocesses receive (design rev 2.2 §4, §6; REQ-31).
package modelcli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Name identifies a supported model CLI.
type Name string

// Supported model CLIs. Anthropic and OpenAI models only.
const (
	Claude Name = "claude"
	Codex  Name = "codex"
)

// NoCLIReason is the lane-health and coverage reason when no model CLI exists.
const NoCLIReason = "skipped (no model CLI)"

// ErrNoCLI means no supported model CLI is installed. It is not a failure:
// optional model-CLI lanes report it as skipped and the exit code is unchanged.
var ErrNoCLI = errors.New(NoCLIReason)

// Parse validates a configured CLI name. The empty string means "first found".
func Parse(value string) (Name, error) {
	switch Name(value) {
	case "", Claude, Codex:
		return Name(value), nil
	}
	return "", fmt.Errorf("unsupported model CLI %q: only claude and codex are allowed", value)
}

// Detect resolves the CLI to use and its absolute path. With an empty
// preference it tries claude then codex. A CLI counts as available when its
// executable is on PATH; a missing login surfaces as a failed run, which the
// lane reports as incomplete rather than clean. No CLI on PATH returns ErrNoCLI.
func Detect(preferred string) (Name, string, error) {
	want, err := Parse(preferred)
	if err != nil {
		return "", "", err
	}
	candidates := []Name{Claude, Codex}
	if want != "" {
		candidates = []Name{want}
	}
	for _, name := range candidates {
		if path, lookErr := exec.LookPath(string(name)); lookErr == nil {
			return name, path, nil
		}
	}
	return "", "", ErrNoCLI
}

// credentialNames are removed (with any OPENROUTER_*, *API_KEY*, *SECRET*
// name) from every model-CLI subprocess environment: a CLI must use its own
// subscription login, never an API key (REQ-31, CI carries no Anthropic/OpenAI keys).
var credentialNames = map[string]bool{
	"ANTHROPIC_AUTH_TOKEN": true,
	"GITHUB_TOKEN":         true,
	"GH_TOKEN":             true,
	"INFISICAL_TOKEN":      true,
	"ANTHROPIC_BASE_URL":   true,
	"OPENAI_BASE_URL":      true,
	"OPENAI_ORG_ID":        true,
}

// IsCredential reports whether an environment variable name carries a
// credential or a provider re-route that model-CLI subprocesses must not see.
func IsCredential(name string) bool {
	upper := strings.ToUpper(name)
	return credentialNames[upper] || strings.HasPrefix(upper, "OPENROUTER_") ||
		strings.Contains(upper, "API_KEY") || strings.Contains(upper, "SECRET")
}

// Env returns the current environment without credentials, followed by extra
// KEY=VALUE entries (which win over inherited entries of the same name).
func Env(extra ...string) []string {
	return ScrubEnv(os.Environ(), extra...)
}

// ScrubEnv filters credentials from base and appends extra.
func ScrubEnv(base []string, extra ...string) []string {
	override := map[string]bool{}
	for _, entry := range extra {
		if name, _, ok := strings.Cut(entry, "="); ok {
			override[name] = true
		}
	}
	out := make([]string, 0, len(base)+len(extra))
	for _, entry := range base {
		name, _, ok := strings.Cut(entry, "=")
		if !ok || IsCredential(name) || override[name] {
			continue
		}
		out = append(out, entry)
	}
	return append(out, extra...)
}

package policy

import (
	"fmt"
	"strings"
	"testing"
)

// ADR-011 implementation contract: add
//
//	func (p Policy) ActivationDiagnostic(checkID, ruleVersion string) string
//
// Return an empty string for a matching activation; for an eligible secrets
// activation with a nonmatching version, name the required full ruleVersion and
// explain that the check stays in shadow. The existing `scout readiness`
// capabilities[].reason output (cmd/scout/main.go, readiness) is the selected
// user-facing diagnostic surface; no new command is needed. Wire this diagnostic
// there as an advisory, not a failed policy parse or an activation migration.
// The interface assertion keeps the base compiling so the proof fails at runtime
// for missing behavior rather than preventing the other package tests from running.
type adr011ActivationDiagnoser interface {
	ActivationDiagnostic(checkID, ruleVersion string) string
}

func TestADR011ShortVersionDiagnostic(t *testing.T) {
	const fullVersion = "gitleaks-default-rules/8.30.1@sha256:e163e53b9e7e8a8511e77271e2b323ed057759542a6d988258afe3a1fa329caf"
	const checkID = "security.secrets.github-pat"
	if err := ValidateEligibleID(checkID); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name    string
		version string
		want    Status
	}{
		{name: "short binary version", version: "8.30.1", want: Shadow},
		{name: "pack label without fingerprint", version: strings.SplitN(fullVersion, "@", 2)[0], want: Shadow},
		{name: "other nonfull version", version: "not-the-full-pack", want: Shadow},
		{name: "full pack string", version: fullVersion, want: Active},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := fmt.Sprintf(`profile: guardian
blocking:
  checks:
    - id: %s
      version: %q
      activated_at: 2026-11-08
      evidence: {would_block: 3, warranted: 3, window_days: 14, pr_runs: 9}
`, checkID, tc.version)
			p, err := Parse([]byte(data))
			if err != nil {
				t.Fatalf("version mismatch must remain a usable shadow policy, not a parse failure: %v", err)
			}
			if got := p.CheckStatus(checkID, fullVersion); got != tc.want {
				t.Errorf("CheckStatus = %q, want %q", got, tc.want)
			}
			diagnoser, ok := any(p).(adr011ActivationDiagnoser)
			if !ok {
				t.Fatal("missing ADR-011 activation diagnostic: policy.Policy must implement ActivationDiagnostic(checkID, ruleVersion string) string for readiness output")
			}
			diagnostic := diagnoser.ActivationDiagnostic(checkID, fullVersion)
			if tc.want == Active {
				if diagnostic != "" {
					t.Fatalf("matching full pack activation has diagnostic %q, want none", diagnostic)
				}
				return
			}
			if !strings.Contains(diagnostic, fullVersion) || !strings.Contains(strings.ToLower(diagnostic), "shadow") {
				t.Fatalf("diagnostic %q must name required full string %q and explain shadow", diagnostic, fullVersion)
			}
		})
	}
}

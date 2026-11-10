package policy

import (
	"strings"
	"testing"
)

func TestADR012ProvisionalUndecidedVeto(t *testing.T) {
	p := Default()
	for _, check := range []string{"security.ci-mutable-action-ref", "security.secrets.github-pat"} {
		t.Run(check, func(t *testing.T) {
			s := ShadowStats{CheckID: check, Days: 30, PRRuns: 7, TotalRuns: 7, WouldBlock: 0, Undecided: 1}
			ready, reason := p.ShadowReady(s)
			if ready || !strings.Contains(strings.ToLower(reason), "undecided") {
				t.Fatalf("cohort Undecided must veto provisional: ready=%v reason=%q", ready, reason)
			}
		})
	}
	t.Run("provisional-reason-discloses-inspected-count", func(t *testing.T) {
		s := ShadowStats{CheckID: "security.ci-mutable-action-ref", Days: 30, PRRuns: 7, TotalRuns: 9}
		ready, reason := p.ShadowReady(s)
		if !ready || !strings.HasPrefix(reason, "provisional:") || !strings.Contains(reason, "7") || !strings.Contains(strings.ToLower(reason), "inspect") {
			t.Fatalf("provisional reason must disclose 7 inspected observations: ready=%v reason=%q", ready, reason)
		}
	})
}

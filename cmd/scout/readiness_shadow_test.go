package main

import (
	"testing"
)

func TestShadowReadinessRejectsEmptyExplicitSelectors(t *testing.T) {
	for _, name := range []string{"as-of", "version"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			p22Hostile(t)
			code, stdout, stderr := p22Run(t, root, "--shadow", "--check", p22Check, "--json", "--"+name+"=")
			if code != exitUnavailable {
				t.Fatalf("explicit empty --%s exit=%d, want 3: %s %s", name, code, stdout, stderr)
			}
			reply := p22JSON(t, stdout, code)
			if reply.Status != "unavailable" || reply.Error == "" {
				t.Fatalf("empty selector must report invalid input, not default to a current query: %+v", reply)
			}
		})
	}
}

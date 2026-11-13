package analyzers

import (
	"strings"
	"testing"
)

func TestADR011GitleaksPackVersionMatchesBinary(t *testing.T) {
	pinned := specForTest(t, "security.gitleaks.default-credentials")
	prefix, fingerprint, ok := strings.Cut(pinned.Rules.Version, "@")
	if !ok || prefix != "gitleaks-default-rules/"+pinned.Binary.Version {
		t.Fatalf("unexpected pinned Gitleaks pack: %q for binary %q", pinned.Rules.Version, pinned.Binary.Version)
	}
	cases := []struct {
		name    string
		mutate  func(*Spec)
		wantErr bool
	}{
		{name: "current pinned spec", mutate: func(*Spec) {}},
		{name: "bumped binary with stale pack label", mutate: func(s *Spec) {
			s.Binary.Version = "8.30.2"
		}, wantErr: true},
		{name: "different pack version with unchanged content", mutate: func(s *Spec) {
			s.Rules.Version = "gitleaks-default-rules/8.30.0@" + fingerprint
		}, wantErr: true},
		{name: "missing version segment", mutate: func(s *Spec) {
			s.Rules.Version = "gitleaks-default-rules/@" + fingerprint
		}, wantErr: true},
		{name: "missing version separator and segment", mutate: func(s *Spec) {
			s.Rules.Version = "gitleaks-default-rules@" + fingerprint
		}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec := pinned
			tc.mutate(&spec)
			err := spec.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("Validate accepted Gitleaks pack %q with binary version %q; pack version segment must equal the binary pin", spec.Rules.Version, spec.Binary.Version)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("Validate rejected current pinned Gitleaks spec: %v", err)
			}
		})
	}
}

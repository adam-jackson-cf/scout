package analyzers

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"scout/internal/domain"
)

func specForTest(t *testing.T, id string) Spec {
	t.Helper()
	for _, spec := range append(securitySpecs(), performanceSpecs()...) {
		if spec.ID == id {
			return spec
		}
	}
	t.Fatalf("adapter %q was not registered", id)
	return Spec{}
}

func TestReviewedIssueRuleContractsAreNarrowAndContentPinned(t *testing.T) {
	tests := []struct {
		name      string
		adapterID string
		positive  domain.File
		negative  domain.File
	}{
		{
			name:      "multipart vulnerable declaration",
			adapterID: "security.semgrep.python-multipart-cve-2024-24762",
			positive:  domain.File{Path: "pyproject.toml", Language: "toml", Eligible: true},
			negative:  domain.File{Path: "requirements.txt", Language: "", Eligible: true},
		},
		{
			name:      "current script public path",
			adapterID: "security.semgrep.javascript-current-script-dom-clobbering",
			positive:  domain.File{Path: "runtime/public-path.js", Language: "javascript", Eligible: true},
			negative:  domain.File{Path: "runtime/public-path.ts", Language: "typescript", Eligible: true},
		},
		{
			name:      "height transition",
			adapterID: "performance.semgrep.css-layout-height-transition",
			positive:  domain.File{Path: "wwwroot/site.css", Language: "", Eligible: true},
			negative:  domain.File{Path: "wwwroot/site.scss", Language: "", Eligible: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec := specForTest(t, test.adapterID)
			if err := spec.Validate(); err != nil {
				t.Fatal(err)
			}
			if !spec.Applicable(test.positive) {
				t.Fatalf("positive input %+v is ineligible", test.positive)
			}
			if spec.Applicable(test.negative) {
				t.Fatalf("negative-boundary input %+v is eligible", test.negative)
			}
			if len(spec.Rules.BundledFiles) != 1 {
				t.Fatalf("expected one bundled rule, got %d", len(spec.Rules.BundledFiles))
			}
			content, err := bundledRule(spec.Rules.BundledFiles[0])
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(content)
			if !strings.Contains(spec.Rules.Version, "sha256:"+hex.EncodeToString(sum[:])) {
				t.Fatal("rule version does not pin the embedded rule content")
			}
		})
	}
}

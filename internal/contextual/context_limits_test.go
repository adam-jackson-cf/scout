package contextual

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"scout/internal/domain"
)

func featureWith(paths ...string) domain.Feature {
	feature := testReviewRequest("").Feature
	feature.OwnedFiles = nil
	for _, path := range paths {
		feature.OwnedFiles = append(feature.OwnedFiles, domain.FeatureFileRef{Path: path, Reason: "service"})
	}
	return feature
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func omitted(manifest Manifest, path string) bool {
	for _, omission := range manifest.Omissions {
		if omission.Path == path {
			return true
		}
	}
	return false
}

func TestOversizedFilesAndTheTotalContextLimitAreTruncatedAndDisclosed(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "big.go"), "line one\nline two\nline three\nline four\n")
	writeFile(t, filepath.Join(root, "second.go"), "second\n")

	assertPerFileLimitTruncates(t, root)

	manifest, sources := readFeatureSources(root, featureWith("second.go", "big.go"), 7, 1<<20)
	if !manifest.Truncated || len(sources) != 1 || sources[0].Path != "second.go" || !omitted(manifest, "big.go") {
		t.Fatalf("total limit: manifest %+v sources %d", manifest, len(sources))
	}
}

func assertPerFileLimitTruncates(t *testing.T, root string) {
	t.Helper()
	manifest, sources := readFeatureSources(root, featureWith("big.go"), 1<<20, 20)
	if !manifest.Truncated || !omitted(manifest, "big.go") || len(sources) != 1 {
		t.Fatalf("per-file limit: manifest %+v", manifest)
	}
	if text := sources[0].Text; len(text) > 20 || !strings.HasSuffix(text, "\n") || strings.Contains(text, "three") {
		t.Fatalf("excerpt must be cut on a line boundary within the limit: %q", text)
	}
	digest := manifest.Inspected[0].SourceDigest
	if digest != sourceDigest([]byte("line one\nline two\nline three\nline four\n")) {
		t.Fatal("the manifest digest must cover the full file, not the excerpt")
	}
}

func TestSeverityMatrixMapsEveryFamily(t *testing.T) {
	cases := []struct {
		family domain.Family
		raw    string
		want   domain.Severity
	}{
		{domain.FamilySecurity, `{"impact":1,"likelihood":1}`, domain.SeverityInformational},
		{domain.FamilySecurity, `{"impact":1,"likelihood":5}`, domain.SeverityLow},
		{domain.FamilySecurity, `{"impact":2,"likelihood":3}`, domain.SeverityMedium},
		{domain.FamilySecurity, `{"impact":3,"likelihood":4}`, domain.SeverityHigh},
		{domain.FamilySecurity, `{"impact":4,"likelihood":5}`, domain.SeverityCritical},
		{domain.FamilyArchitecture, `{"changeRisk":"blocking","boundary":"systemic_or_critical_boundary"}`, domain.SeverityCritical},
		{domain.FamilyArchitecture, `{"changeRisk":"major","boundary":"component"}`, domain.SeverityHigh},
		{domain.FamilyArchitecture, `{"changeRisk":"minor","boundary":"local"}`, domain.SeverityLow},
		{domain.FamilyArchitecture, `{"changeRisk":"blocking","boundary":"none"}`, domain.SeverityInformational},
		{domain.FamilyQuality, `{"changeRisk":"blocking","scope":"systemic"}`, domain.SeverityCritical},
		{domain.FamilyQuality, `{"changeRisk":"meaningful","scope":"component"}`, domain.SeverityMedium},
		{domain.FamilyQuality, `{"changeRisk":"blocking","scope":"isolated"}`, domain.SeverityLow},
		{domain.FamilyPerformance, `{"operationalRisk":"unacceptable","frequency":"systemic"}`, domain.SeverityCritical},
		{domain.FamilyPerformance, `{"operationalRisk":"major","frequency":"bounded"}`, domain.SeverityMedium},
		{domain.FamilyPerformance, `{"operationalRisk":"meaningful","frequency":"rare"}`, domain.SeverityLow},
		{domain.FamilyPerformance, `{"operationalRisk":"unacceptable","frequency":"none"}`, domain.SeverityInformational},
	}
	for _, tc := range cases {
		got, err := severityFor(tc.family, []byte(tc.raw))
		if err != nil || got != tc.want {
			t.Fatalf("%s %s = %q, %v; want %q", tc.family, tc.raw, got, err, tc.want)
		}
	}
	invalid := []struct {
		family domain.Family
		raw    string
	}{
		{domain.FamilySecurity, `{"impact":0,"likelihood":3}`},
		{domain.FamilySecurity, `{"impact":3,"likelihood":6}`},
		{domain.FamilyQuality, `{"changeRisk":"unacceptable","scope":"local"}`},
		{domain.FamilyPerformance, `{"operationalRisk":"major","frequency":"galaxy"}`},
		{domain.FamilyQuality, `{"changeRisk":"major","scope":"local","extra":1}`},
		{domain.FamilyClawpatch, `{"impact":1,"likelihood":1}`},
		{domain.FamilySecurity, ``},
	}
	for _, tc := range invalid {
		if got, err := severityFor(tc.family, []byte(tc.raw)); err == nil {
			t.Fatalf("%s %q must be rejected, got %q", tc.family, tc.raw, got)
		}
	}
}

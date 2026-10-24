package artifacts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"scout/internal/domain"
)

const head = "0123456789abcdef0123456789abcdef01234567"

func write(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

const lcov = "TN:\nSF:{ROOT}/pkg/a.py\nFN:1,f\nDA:1,0\nDA:2,5\nend_of_record\nSF:pkg/b.py\nDA:1,1\nDA:3,0\nend_of_record\n"

const cobertura = `<?xml version="1.0"?>
<coverage line-rate="0.5"><sources><source>{ROOT}/pkg</source></sources>
<packages><package name="p"><classes>
<class filename="a.py"><methods><method name="f"><lines><line number="1" hits="9"/></lines></method></methods>
<lines><line number="1" hits="0"/><line number="2" hits="5"/></lines></class>
</classes></package></packages></coverage>`

const junit = `<testsuites>
<testsuite name="a" tests="3" failures="1" errors="0" skipped="1" time="0.5">
<testcase name="ok" time="0.1"/><testcase name="bad" time="0.2"><failure message="x"/></testcase><testcase name="skip"><skipped/></testcase></testsuite>
<testsuite name="b" tests="2" failures="0" errors="1" skipped="0" time="1.5"/></testsuites>`

func TestParsersNormalizeToRepositoryPaths(t *testing.T) {
	root := t.TempDir()
	for name, parse := range map[string]func() (Coverage, error){
		"lcov": func() (Coverage, error) { return ParseLCOV([]byte(strings.ReplaceAll(lcov, "{ROOT}", root)), root) },
		"cobertura": func() (Coverage, error) {
			return ParseCobertura([]byte(strings.ReplaceAll(cobertura, "{ROOT}", root)), root)
		},
	} {
		got, err := parse()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if a := got["pkg/a.py"]; len(a) != 2 || a[1] != 0 || a[2] != 5 {
			t.Fatalf("%s: pkg/a.py = %v", name, a)
		}
	}
	got, _ := ParseLCOV([]byte(strings.ReplaceAll(lcov, "{ROOT}", root)), root)
	if b := got["pkg/b.py"]; len(b) != 2 || b[1] != 1 || b[3] != 0 {
		t.Fatalf("relative SF path: %v", b)
	}
	if _, err := ParseLCOV([]byte("DA:1,1\n"), root); err == nil {
		t.Fatal("DA outside a record was accepted")
	}
}

func TestJUnitTotals(t *testing.T) {
	got, err := ParseJUnit([]byte(junit))
	if err != nil {
		t.Fatal(err)
	}
	want := TestResults{Suites: 2, Tests: 5, Failures: 1, Errors: 1, Skipped: 1, Seconds: 0.1 + 0.2 + 1.5}
	if got != want {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

func TestMatchRevision(t *testing.T) {
	clean := domain.Snapshot{Revision: head}
	for _, c := range []struct {
		name     string
		revision string
		snapshot domain.Snapshot
		want     string
	}{
		{"full", head, clean, MatchMatch},
		{"abbreviated", head[:8], clean, MatchMatch},
		{"too short", head[:4], clean, MatchMismatch},
		{"other commit", "ffffffffffff", clean, MatchMismatch},
		{"unstated", "", clean, MatchUnknown},
		{"dirty tree", head, domain.Snapshot{Revision: head, Dirty: true}, MatchMismatch},
		{"no commit", head, domain.Snapshot{}, MatchMismatch},
	} {
		got, reason := MatchRevision(c.revision, c.snapshot)
		if got != c.want || (got != MatchMatch && reason == "") {
			t.Errorf("%s: got %q (%q), want %q with a reason otherwise", c.name, got, reason, c.want)
		}
	}
}

func TestLoadAttachesProvenanceAndOnlyMatchedCoverageIsUsable(t *testing.T) {
	root := t.TempDir()
	write(t, root, "ci/coverage.lcov", strings.ReplaceAll(lcov, "{ROOT}", root))
	write(t, root, "ci/coverage.lcov.scout.json", `{"revision":"`+head+`","producer":"pytest-cov 5"}`)
	write(t, root, "ci/cobertura.xml", strings.ReplaceAll(cobertura, "{ROOT}", root))
	write(t, root, "ci/junit.xml", junit)
	snapshot := domain.Snapshot{Revision: head}

	result := Load(root, []Spec{{Path: "ci/*.lcov"}, {Path: "ci/cobertura.xml", Revision: "ffffffffffff", Producer: "cobertura-tool"}, {Path: "ci/junit.xml", Revision: head[:10]}}, snapshot)
	if len(result.Failures) != 0 || len(result.Artifacts) != 3 {
		t.Fatalf("result = %+v", result)
	}
	byKind := map[string]Artifact{}
	for _, a := range result.Artifacts {
		byKind[a.Kind] = a
	}
	l := byKind[KindLCOV]
	if l.Revision != head || l.Producer != "pytest-cov 5" || l.RevisionMatch != MatchMatch || len(l.Digest) != 64 || l.Source != "ci/coverage.lcov" {
		t.Fatalf("lcov provenance = %+v", l.Provenance)
	}
	if c := byKind[KindCobertura]; c.RevisionMatch != MatchMismatch || c.Producer != "cobertura-tool" {
		t.Fatalf("cobertura provenance = %+v", c.Provenance)
	}
	j := byKind[KindJUnit]
	if j.RevisionMatch != MatchMatch || j.Tests == nil || j.Tests.Tests != 5 || j.Coverage != nil {
		t.Fatalf("junit = %+v", j)
	}

	merged := result.MatchedCoverage()
	if merged.Digest == "" || len(merged.Sources) != 1 || merged.Sources[0] != "ci/coverage.lcov" || merged.Reason != "" {
		t.Fatalf("merged = %+v", merged)
	}
	if merged.Lines["pkg/a.py"][2] != 5 {
		t.Fatalf("merged lines = %v", merged.Lines)
	}

	// Same artifacts against another commit: nothing matches, with a reason.
	other := Load(root, []Spec{{Path: "ci/*.lcov"}}, domain.Snapshot{Revision: strings.Repeat("a", 40)}).MatchedCoverage()
	if other.Lines != nil || other.Digest != "" || !strings.Contains(other.Reason, "no coverage artifact matches") {
		t.Fatalf("mismatched merged = %+v", other)
	}
}

func TestLoadReportsUnreadableAndUnrecognizedArtifacts(t *testing.T) {
	root := t.TempDir()
	write(t, root, "notes.txt", "hello\n")
	write(t, root, "bad.xml", "<coverage><oops")
	result := Load(root, []Spec{{Path: "missing.lcov"}, {Path: "notes.txt"}, {Path: "bad.xml"}, {Path: "nomatch/*.xml"}}, domain.Snapshot{Revision: head})
	if len(result.Artifacts) != 0 || len(result.Failures) != 4 {
		t.Fatalf("result = %+v", result)
	}
	if reason := result.MatchedCoverage().Reason; !strings.Contains(reason, "could not be read") {
		t.Fatalf("reason = %q", reason)
	}
}

func TestUnreadableSidecarFailsInsteadOfFallingBackToTheFlagRevision(t *testing.T) {
	root := t.TempDir()
	write(t, root, "cov.lcov", strings.ReplaceAll(lcov, "{ROOT}", root))
	if err := os.Mkdir(filepath.Join(root, "cov.lcov"+SidecarSuffix), 0o755); err != nil {
		t.Fatal(err)
	}
	result := Load(root, []Spec{{Path: "cov.lcov", Revision: head}}, domain.Snapshot{Revision: head})
	if len(result.Artifacts) != 0 || len(result.Failures) != 1 {
		t.Fatalf("unreadable sidecar produced %+v", result)
	}
	if merged := result.MatchedCoverage(); merged.Digest != "" {
		t.Fatalf("coverage was used despite the unreadable sidecar: %+v", merged)
	}
}

func TestMatchedCoverageReportsMismatchesAndFailuresTogether(t *testing.T) {
	root := t.TempDir()
	write(t, root, "stale.lcov", strings.ReplaceAll(lcov, "{ROOT}", root))
	write(t, root, "notes.txt", "hello\n")
	result := Load(root, []Spec{{Path: "stale.lcov", Revision: "ffffffffffff"}, {Path: "notes.txt"}}, domain.Snapshot{Revision: head})
	reason := result.MatchedCoverage().Reason
	if !strings.Contains(reason, "no coverage artifact matches") || !strings.Contains(reason, "could not be read") {
		t.Fatalf("reason dropped a cause: %q", reason)
	}
}

func TestPathsStartingWithDotsStayInsideTheRoot(t *testing.T) {
	root := t.TempDir()
	if got := displayPath(root, filepath.Join(root, "..cache", "cov.lcov")); got != "..cache/cov.lcov" {
		t.Fatalf("displayPath = %q", got)
	}
	if got := displayPath(root, filepath.Join(filepath.Dir(root), "outside.lcov")); got != filepath.ToSlash(filepath.Join(filepath.Dir(root), "outside.lcov")) {
		t.Fatalf("outside path was made relative: %q", got)
	}
}

func TestEmptySpecsGiveNoCoverageReason(t *testing.T) {
	merged := Result{}.MatchedCoverage()
	if merged.Lines != nil || merged.Reason == "" {
		t.Fatalf("merged = %+v", merged)
	}
}

package tooling

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, text := range files {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestClassifyFindsLocksManifestsAndUnknownFormats(t *testing.T) {
	files := map[string]string{
		"go.mod":                       "module x\n",
		"web/package.json":             `{"dependencies":{"a":"1.0.0"}}`,
		"web/package-lock.json":        "{}",
		"bare/package.json":            `{"dependencies":{"a":"1.0.0"}}`,
		"empty/package.json":           `{"name":"x"}`,
		"py/requirements.in":           "a==1\n",
		"py/requirements.lock":         "a==1 \\\n    --hash=sha256:00\n",
		"study/requirements.txt":       "numpy==2.5.4\n",
		"odd/deps.lock":                "?",
		"orphan/tools-requirements.in": "b==2\n",
	}
	root := writeFiles(t, files)
	var names []string
	for f := range files {
		names = append(names, f)
	}
	c := classify(root, sortedCopy(names))
	got := map[string]string{}
	for _, l := range c.locks {
		got[l.path] = "lock:" + l.parseAs
	}
	for _, f := range c.failures {
		got[f.path] = f.kind
	}
	want := map[string]string{
		"go.mod": "lock:go.mod", "web/package-lock.json": "lock:package-lock.json", "py/requirements.lock": "lock:requirements.txt",
		"bare/package.json": "manifest-without-lock", "study/requirements.txt": "manifest-without-lock",
		"odd/deps.lock": "unknown-lock-format", "orphan/tools-requirements.in": "manifest-without-lock",
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestOSVReportPrintsTheCatalogueGrammar(t *testing.T) {
	report := `{"results":[{"source":{"path":"/r/web/package-lock.json"},"packages":[{"package":{"name":"lodash","version":"4.17.15","ecosystem":"npm"},
		"vulnerabilities":[{"id":"GHSA-p6mc-m468-83gw","aliases":["CVE-2020-8203"],"affected":[{"package":{"name":"lodash","ecosystem":"npm"},"ranges":[{"events":[{"introduced":"0"},{"fixed":"4.17.19"}]}]}]}]}]}]}`
	got, err := parseOSV("/r", []byte(report))
	if err != nil {
		t.Fatal(err)
	}
	want := "SECURITY.dependency-audit lock=web/package-lock.json ecosystem=npm package=lodash version=4.17.15 advisory=GHSA-p6mc-m468-83gw aliases=CVE-2020-8203 fix=4.17.19"
	if len(got) != 1 || got[0].line() != want {
		t.Fatalf("got %+v", got)
	}
	if _, err := parseOSV("/r", []byte("garbage")); err == nil {
		t.Fatal("untrusted output must fail")
	}
}

func TestMaliciousAdvisoryAsksForCredentialRotation(t *testing.T) {
	f := advisory{lock: "l", ecosystem: "npm", pkg: "evil", version: "1.0.0", id: "MAL-2026-1"}.finding()
	if !strings.Contains(f.Fix, "rotate every credential") {
		t.Fatalf("got %q", f.Fix)
	}
}

func entry(fields map[string]string) map[string]any {
	out := map[string]any{"advisory": "GHSA-1", "package": "otel", "version": "1.37.0", "lock": "req.lock", "reason": "r", "exposure": "dev only", "expires": "2026-12-31"}
	for k, v := range fields {
		if v == "" {
			delete(out, k)
			continue
		}
		out[k] = v
	}
	return out
}

func TestExceptionEntryShapeIsValidated(t *testing.T) {
	for name, tc := range map[string]map[string]string{
		"missing field": {"exposure": ""},
		"unknown field": {"note": "x"},
		"bad date":      {"expires": "31/12/2026"},
		"http source":   {"blocked_by": "semgrep@1.185.0", "constraint": "c", "constraint_source": "http://pypi.org/x"},
	} {
		if _, reason := validateEntry(entry(tc)); reason == "" {
			t.Errorf("%s: want a validation failure", name)
		}
	}
	if _, reason := validateEntry(entry(nil)); reason != "" {
		t.Fatalf("valid entry rejected: %s", reason)
	}
}

func TestExceptionsSuppressOnlyJustifiedAdvisories(t *testing.T) {
	root := writeFiles(t, map[string]string{"req.lock": "semgrep==1.185.0 \\\notel==1.37.0 \\\n"})
	now := time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC)
	locks := []lockFile{{path: "req.lock", parseAs: "requirements.txt"}}
	noFix := advisory{lock: "req.lock", ecosystem: "PyPI", pkg: "otel", version: "1.37.0", id: "GHSA-1"}
	withFix := noFix
	withFix.fixes = []string{"1.38.0"}
	blocked := map[string]string{"blocked_by": "semgrep@1.185.0", "constraint": "semgrep==1.185.0 requires otel~=1.37.0", "constraint_source": "https://pypi.org/project/semgrep/1.185.0/"}
	cases := []struct {
		name      string
		adv       advisory
		fields    map[string]string
		suppress  bool
		failWords string
	}{
		{"no-fix entry", noFix, nil, true, ""},
		{"expired entry", noFix, map[string]string{"expires": "2026-11-26"}, false, "expired"},
		{"unblocked fix", withFix, nil, false, "names no blocking dependent"},
		{"blocked fix", withFix, blocked, true, ""},
		{"compatible fix", advisory{lock: "req.lock", ecosystem: "PyPI", pkg: "otel", version: "1.37.0", id: "GHSA-1", fixes: []string{"1.37.5"}}, blocked, false, "satisfies"},
		{"missing blocker", withFix, map[string]string{"blocked_by": "semgrep@1.184.0", "constraint": "semgrep==1.184.0 requires otel~=1.37.0", "constraint_source": "https://pypi.org/x"}, false, "is not in"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e, reason := validateEntry(entry(tc.fields))
			if reason != "" {
				t.Fatal(reason)
			}
			remaining, failures := applyExceptions([]advisory{tc.adv}, []auditException{e}, exceptionContext{locks: locks, root: root, now: now})
			if tc.suppress != (len(remaining) == 0) {
				t.Fatalf("remaining %v failures %v", remaining, failures)
			}
			if tc.failWords != "" && (len(failures) != 1 || !strings.Contains(failures[0].reason, tc.failWords)) {
				t.Fatalf("failures %+v, want %q", failures, tc.failWords)
			}
		})
	}
	stale, _ := validateEntry(entry(map[string]string{"advisory": "GHSA-9"}))
	if _, failures := applyExceptions(nil, []auditException{stale}, exceptionContext{locks: locks, root: root, now: now}); len(failures) != 1 || !strings.Contains(failures[0].reason, "matches no reported advisory") {
		t.Fatalf("stale entry: %+v", failures)
	}
	if _, failures := applyExceptions(nil, []auditException{stale}, exceptionContext{locks: locks, root: root, now: now, explicit: true}); len(failures) != 0 {
		t.Fatalf("an explicit --path run audits a subset, so unmatched entries are not stale: %+v", failures)
	}
}

func TestVendoredExceptionHoldsOnlyForLocksInsideASubmodule(t *testing.T) {
	now := time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC)
	inside := advisory{lock: "third_party/up/pnpm-lock.yaml", ecosystem: "npm", pkg: "vite", version: "8.2.2", id: "GHSA-1", fixes: []string{"8.2.4"}}
	outside := inside
	outside.lock = "web/pnpm-lock.yaml"
	noFix := inside
	noFix.fixes = nil
	ctx := exceptionContext{now: now, submodules: map[string]bool{"third_party/up": true}}
	cases := []struct {
		name      string
		adv       advisory
		vendored  string
		failWords string
	}{
		{"lock inside the submodule", inside, "third_party/up", ""},
		{"directory is not a submodule", inside, "third_party", "is not a git submodule"},
		{"lock outside the submodule", outside, "third_party/up", "is not inside"},
		{"advisory without a fix", noFix, "third_party/up", "no fixed version"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e, reason := validateEntry(entry(map[string]string{"lock": tc.adv.lock, "package": "vite", "version": "8.2.2", "vendored_in": tc.vendored}))
			if reason != "" {
				t.Fatal(reason)
			}
			remaining, failures := applyExceptions([]advisory{tc.adv}, []auditException{e}, ctx)
			if tc.failWords == "" && (len(remaining) != 0 || len(failures) != 0) {
				t.Fatalf("remaining %v failures %v", remaining, failures)
			}
			if tc.failWords != "" && (len(remaining) != 1 || len(failures) != 1 || !strings.Contains(failures[0].reason, tc.failWords)) {
				t.Fatalf("remaining %v failures %+v, want %q", remaining, failures, tc.failWords)
			}
		})
	}
}

func pathEntry(t *testing.T, p, expires string) auditException {
	t.Helper()
	e, reason := validateEntry(map[string]any{"path": p, "reason": "frozen study record", "exposure": "never installed", "expires": expires})
	if reason != "" {
		t.Fatal(reason)
	}
	return e
}

func TestPathExclusionDropsOnlyUnexpiredListedFiles(t *testing.T) {
	files := []string{"study/requirements.txt", "go.mod"}
	live := pathEntry(t, "study/requirements.txt", "2027-02-27")
	excluded, failures := fileExclusions([]auditException{live}, files, "2026-11-27", false)
	if !excluded["study/requirements.txt"] || len(failures) != 0 {
		t.Fatalf("excluded %v failures %+v", excluded, failures)
	}
	excluded, failures = fileExclusions([]auditException{live}, files, "2027-02-28", false)
	if excluded["study/requirements.txt"] || len(failures) != 1 || !strings.Contains(failures[0].reason, "expired on 2027-02-27") {
		t.Fatalf("expired: excluded %v failures %+v", excluded, failures)
	}
	gone := pathEntry(t, "old/requirements.txt", "2027-02-27")
	if _, failures := fileExclusions([]auditException{gone}, files, "2026-11-27", false); len(failures) != 1 || !strings.Contains(failures[0].reason, "matches no audited file") {
		t.Fatalf("stale path: %+v", failures)
	}
	if _, failures := fileExclusions([]auditException{gone}, files, "2026-11-27", true); len(failures) != 0 {
		t.Fatalf("explicit run: %+v", failures)
	}
}

func TestPathEntryMustNameOneExactFileWithItsExposure(t *testing.T) {
	for _, bad := range []string{"study/*.txt", "study/"} {
		if _, reason := validateEntry(map[string]any{"path": bad, "reason": "r", "exposure": "e", "expires": "2027-02-27"}); reason == "" {
			t.Errorf("path %q: want a validation failure", bad)
		}
	}
	if _, reason := validateEntry(map[string]any{"path": "a.txt", "reason": "r", "expires": "2027-02-27"}); reason == "" {
		t.Error("path entry without exposure: want a validation failure")
	}
}

func TestVersionSpecifiers(t *testing.T) {
	for _, tc := range []struct {
		spec, version string
		want          bool
	}{
		{"~=1.37.0", "1.37.9", true}, {"~=1.37.0", "1.38.0", false}, {">=1,<2", "1.5", true}, {">=1,<2", "2.0", false},
		{"^1.2.3", "1.9.0", true}, {"^1.2.3", "2.0.0", false}, {"~1.2.3", "1.2.9", true}, {"~1.2.3", "1.3.0", false}, {"==4.17.15", "4.17.15", true},
	} {
		got, err := satisfies(tc.spec, tc.version)
		if err != nil || got != tc.want {
			t.Errorf("satisfies(%q, %q) = %v, %v; want %v", tc.spec, tc.version, got, err, tc.want)
		}
	}
}

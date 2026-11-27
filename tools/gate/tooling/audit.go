package tooling

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"scout/tools/gate/check"
)

const (
	auditRule      = "SECURITY.dependency-audit"
	exceptionsFile = "toolchain/dependency-audit-exceptions.json"
)

var lockParsers = map[string]string{
	"go.mod": "go.mod", "package-lock.json": "package-lock.json", "npm-shrinkwrap.json": "package-lock.json",
	"pnpm-lock.yaml": "pnpm-lock.yaml", "yarn.lock": "yarn.lock", "bun.lock": "bun.lock", "Cargo.lock": "Cargo.lock",
	"Gemfile.lock": "Gemfile.lock", "composer.lock": "composer.lock", "poetry.lock": "poetry.lock",
	"Pipfile.lock": "Pipfile.lock", "pdm.lock": "pdm.lock", "uv.lock": "uv.lock", "pubspec.lock": "pubspec.lock",
	"mix.lock": "mix.lock", "gradle.lockfile": "gradle.lockfile", "packages.lock.json": "packages.lock.json",
	"renv.lock": "renv.lock", "conan.lock": "conan.lock",
}

var manifestLocks = map[string][]string{
	"package.json":   {"package-lock.json", "npm-shrinkwrap.json", "pnpm-lock.yaml", "yarn.lock", "bun.lock"},
	"pyproject.toml": {"uv.lock", "poetry.lock", "pdm.lock"},
	"Cargo.toml":     {"Cargo.lock"},
	"Gemfile":        {"Gemfile.lock"},
	"composer.json":  {"composer.lock"},
	"Pipfile":        {"Pipfile.lock"},
	"pubspec.yaml":   {"pubspec.lock"},
}

var (
	requirementsName = regexp.MustCompile(`(^|[-_.])requirements([-_.][^/]*)?\.(txt|lock|in)$`)
	lockLikeName     = regexp.MustCompile(`(\.lock|[-.]lock\.(json|ya?ml)|\.lockfile)$`)
	hashedPin        = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._\[\],-]*==\S+`)
)

type lockFile struct {
	path, parseAs string
}

type auditFailure struct {
	kind, path, reason string
}

func (f auditFailure) finding() check.Finding {
	return check.Finding{File: f.path, Rule: auditRule, Problem: fmt.Sprintf("%s failure=%s path=%s reason=%s", auditRule, f.kind, f.path, f.reason), Fix: failureFix[f.kind]}
}

var failureFix = map[string]string{
	"unknown-lock-format":   "use a lock format osv-scanner parses, or rename a file that is not a lock",
	"manifest-without-lock": "commit the lock file the manifest's package manager generates",
	"exception":             "correct or remove the entry; only the owner edits " + exceptionsFile + " and proofs.toml exceptions",
	"auditor":               "rerun with network access to api.osv.dev and the pinned osv-scanner from scripts/bootstrap-tools.sh",
}

type classified struct {
	locks    []lockFile
	failures []auditFailure
}

func classify(root string, files []string) classified {
	present := map[string]bool{}
	for _, f := range files {
		present[f] = true
	}
	var c classified
	for _, f := range files {
		base := path.Base(f)
		dir := path.Dir(f)
		switch {
		case lockParsers[base] != "":
			c.locks = append(c.locks, lockFile{path: f, parseAs: lockParsers[base]})
		case requirementsName.MatchString(base):
			c.classifyRequirements(root, f, present)
		case manifestLocks[base] != nil:
			if hasDependencies(root, f) && !anyPresent(present, dir, manifestLocks[base]) {
				c.failures = append(c.failures, auditFailure{"manifest-without-lock", f, "no " + strings.Join(manifestLocks[base], " or ") + " beside it"})
			}
		case lockLikeName.MatchString(base):
			c.failures = append(c.failures, auditFailure{"unknown-lock-format", f, "osv-scanner has no parser for " + base})
		}
	}
	return c
}

func (c *classified) classifyRequirements(root, f string, present map[string]bool) {
	if strings.HasSuffix(f, ".in") {
		stem := strings.TrimSuffix(f, ".in")
		if !present[stem+".lock"] && !present[stem+".txt"] {
			c.failures = append(c.failures, auditFailure{"manifest-without-lock", f, "no " + path.Base(stem) + ".lock or .txt compiled from it"})
		}
		return
	}
	if isHashedLock(filepath.Join(root, f)) {
		c.locks = append(c.locks, lockFile{path: f, parseAs: "requirements.txt"})
		return
	}
	c.failures = append(c.failures, auditFailure{"manifest-without-lock", f, "requirements without exact == pins and --hash for every package"})
}

func isHashedLock(file string) bool {
	data, err := os.ReadFile(file)
	if err != nil {
		return false
	}
	text := strings.ReplaceAll(string(data), "\\\n", " ")
	seen := false
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "--") {
			continue
		}
		if !hashedPin.MatchString(line) || !strings.Contains(line, "--hash=sha256:") {
			return false
		}
		seen = true
	}
	return seen
}

func hasDependencies(root, f string) bool {
	data, err := os.ReadFile(filepath.Join(root, f))
	if err != nil {
		return true
	}
	text := string(data)
	if path.Base(f) == "package.json" {
		var pkg map[string]json.RawMessage
		if json.Unmarshal(data, &pkg) != nil {
			return true
		}
		for _, k := range []string{"dependencies", "devDependencies", "optionalDependencies", "peerDependencies"} {
			if v, ok := pkg[k]; ok && strings.TrimSpace(string(v)) != "{}" {
				return true
			}
		}
		return false
	}
	return strings.Contains(text, "dependencies") || path.Base(f) != "pyproject.toml"
}

func anyPresent(present map[string]bool, dir string, names []string) bool {
	for _, n := range names {
		if present[path.Join(dir, n)] {
			return true
		}
	}
	return false
}

func auditCandidates(r *check.Run) ([]string, []auditFailure, error) {
	if r.Explicit {
		files, err := explicitFiles(r.Root, r.Paths, func(string, string) bool { return true })
		return files, nil, err
	}
	raw, err := r.Git("ls-files", "-z", "--cached", "--others", "--exclude-standard")
	if err != nil {
		return nil, nil, err
	}
	var files []string
	for _, f := range strings.Split(raw, "\x00") {
		if f != "" {
			files = append(files, f)
		}
	}
	sub, failures, err := submoduleFiles(r)
	if err != nil {
		return nil, nil, err
	}
	return append(files, sub...), failures, nil
}

func gitlinks(r *check.Run) ([]string, error) {
	raw, err := r.Git("ls-files", "-z", "--stage")
	if err != nil {
		return nil, err
	}
	var out []string
	for _, entry := range strings.Split(raw, "\x00") {
		meta, rel, ok := strings.Cut(entry, "\t")
		if ok && strings.HasPrefix(meta, "160000 ") {
			out = append(out, rel)
		}
	}
	return out, nil
}

func submoduleFiles(r *check.Run) ([]string, []auditFailure, error) {
	links, err := gitlinks(r)
	if err != nil {
		return nil, nil, err
	}
	var files []string
	var failures []auditFailure
	for _, rel := range links {
		if _, err := os.Stat(filepath.Join(r.Root, rel, ".git")); err != nil {
			failures = append(failures, auditFailure{"auditor", rel, "submodule is not checked out, so its lock files cannot be audited; run git submodule update --init " + rel})
			continue
		}
		listed, err := r.Git("-C", rel, "ls-files", "-z", "--cached")
		if err != nil {
			return nil, nil, err
		}
		for _, f := range strings.Split(listed, "\x00") {
			if f != "" {
				files = append(files, path.Join(rel, f))
			}
		}
	}
	return files, failures, nil
}

func pathExclusions(r *check.Run) (map[string]bool, []auditFailure) {
	out := map[string]bool{}
	var failures []auditFailure
	if r.Proofs == nil {
		return out, nil
	}
	for _, e := range r.Proofs.Exceptions {
		if e.Rule != auditRule {
			continue
		}
		if strings.ContainsAny(e.Path, "*?[") || strings.HasSuffix(e.Path, "/") {
			failures = append(failures, auditFailure{"exception", "proofs.toml", "exclusion " + e.Path + " is a glob or directory; name each fixture, vendored or frozen file exactly"})
			continue
		}
		if strings.TrimSpace(e.Reason) == "" || strings.TrimSpace(e.ApprovedBy) == "" {
			failures = append(failures, auditFailure{"exception", "proofs.toml", "exclusion " + e.Path + " lacks reason or approved_by origin evidence"})
			continue
		}
		out[e.Path] = true
	}
	return out, failures
}

func runAudit(r *check.Run) ([]check.Finding, error) {
	files, failures, err := auditCandidates(r)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	entries, entryFailures := loadExceptions(r.Root)
	failures = append(failures, entryFailures...)
	pathEntries, advisoryEntries := splitExceptions(entries)
	kept, exFailures := keptFiles(r, files, pathEntries, now)
	failures = append(failures, exFailures...)
	c := classify(r.Root, kept)
	failures = append(failures, c.failures...)
	advisories, auditorFailures := scanLocks(r, c.locks)
	failures = append(failures, auditorFailures...)
	links, err := gitlinks(r)
	if err != nil {
		return nil, err
	}
	ctx := exceptionContext{locks: c.locks, root: r.Root, now: now, submodules: setOf(links), explicit: r.Explicit}
	remaining, matchFailures := applyExceptions(advisories, advisoryEntries, ctx)
	failures = append(failures, matchFailures...)
	return auditFindings(remaining, failures), nil
}

func keptFiles(r *check.Run, files []string, pathEntries []auditException, now time.Time) ([]string, []auditFailure) {
	excluded, failures := pathExclusions(r)
	listed, listFailures := fileExclusions(pathEntries, files, now.Format("2006-01-02"), r.Explicit)
	failures = append(failures, listFailures...)
	if r.Explicit {
		return files, failures
	}
	var kept []string
	for _, f := range files {
		if !excluded[f] && !listed[f] {
			kept = append(kept, f)
		}
	}
	return kept, failures
}

func setOf(items []string) map[string]bool {
	out := make(map[string]bool, len(items))
	for _, s := range items {
		out[s] = true
	}
	return out
}

func auditFindings(remaining []advisory, failures []auditFailure) []check.Finding {
	var out []check.Finding
	for _, a := range remaining {
		out = append(out, a.finding())
	}
	for _, f := range failures {
		out = append(out, f.finding())
	}
	return out
}

type advisory struct {
	lock, ecosystem, pkg, version, id string
	aliases, fixes                    []string
}

func (a advisory) key() string { return a.lock + ":" + a.pkg + "@" + a.version + ":" + a.id }

func (a advisory) line() string {
	return fmt.Sprintf("%s lock=%s ecosystem=%s package=%s version=%s advisory=%s aliases=%s fix=%s",
		auditRule, a.lock, a.ecosystem, a.pkg, a.version, a.id, strings.Join(a.aliases, ","), strings.Join(a.fixes, ","))
}

func (a advisory) finding() check.Finding {
	fix := "upgrade " + a.pkg + " to a fixed version and relock"
	if len(a.fixes) == 0 {
		fix = "no fixed release: replace or mitigate " + a.pkg + ", or the owner records a dated entry in " + exceptionsFile
	}
	if isMalicious(a) {
		fix = "malicious package: remove or replace " + a.pkg + " and rotate every credential exposed where it was installed"
	}
	return check.Finding{File: a.lock, Rule: auditRule, Problem: a.line(), Fix: fix}
}

func isMalicious(a advisory) bool {
	for _, id := range append([]string{a.id}, a.aliases...) {
		if strings.HasPrefix(id, "MAL-") {
			return true
		}
	}
	return false
}

type osvReport struct {
	Results []struct {
		Source struct {
			Path string `json:"path"`
		} `json:"source"`
		Packages []struct {
			Package struct {
				Name      string `json:"name"`
				Version   string `json:"version"`
				Ecosystem string `json:"ecosystem"`
			} `json:"package"`
			Vulnerabilities []osvVuln `json:"vulnerabilities"`
		} `json:"packages"`
	} `json:"results"`
}

type osvVuln struct {
	ID       string   `json:"id"`
	Aliases  []string `json:"aliases"`
	Affected []struct {
		Package struct {
			Name      string `json:"name"`
			Ecosystem string `json:"ecosystem"`
		} `json:"package"`
		Ranges []struct {
			Events []struct {
				Fixed string `json:"fixed"`
			} `json:"events"`
		} `json:"ranges"`
	} `json:"affected"`
}

func scanLocks(r *check.Run, locks []lockFile) ([]advisory, []auditFailure) {
	if len(locks) == 0 {
		return nil, nil
	}
	fail := func(reason string) []auditFailure { return []auditFailure{{"auditor", locks[0].path, oneLine(reason)}} }
	data, err := runOSV(r, locks)
	if err != nil {
		return nil, fail(err.Error())
	}
	advisories, err := parseOSV(r.Root, data.report)
	if err != nil {
		return nil, fail(err.Error())
	}
	if data.exitCode == 1 && len(advisories) == 0 {
		return nil, fail("osv-scanner exited 1 without advisories: " + data.stderr)
	}
	return advisories, nil
}

type osvRun struct {
	report   []byte
	exitCode int
	stderr   string
}

func runOSV(r *check.Run, locks []lockFile) (osvRun, error) {
	bin, err := resolveTool(r, osvTool)
	if err != nil {
		return osvRun{}, err
	}
	report := filepath.Join(r.TempDir, "osv.json")
	args := []string{"scan", "source", "--format=json", "--output=" + report}
	for _, l := range locks {
		args = append(args, "--lockfile="+l.parseAs+":"+filepath.Join(r.Root, l.path))
	}
	res, err := runTool(invocation{dir: r.TempDir, timeout: 5 * time.Minute}, bin, args...)
	if err != nil {
		return osvRun{}, err
	}
	if res.exitCode != 0 && res.exitCode != 1 {
		return osvRun{}, fmt.Errorf("osv-scanner exited %d: %s", res.exitCode, res.stderr)
	}
	data, err := os.ReadFile(report)
	if err != nil {
		return osvRun{}, fmt.Errorf("osv-scanner wrote no report: %w", err)
	}
	return osvRun{report: data, exitCode: res.exitCode, stderr: res.stderr}, nil
}

var noSpace = regexp.MustCompile(`^\S+$`)

func parseOSV(root string, data []byte) ([]advisory, error) {
	var rep osvReport
	if err := json.Unmarshal(data, &rep); err != nil {
		return nil, fmt.Errorf("untrusted osv-scanner output: %w", err)
	}
	var out []advisory
	for _, res := range rep.Results {
		lock := relTo(root, res.Source.Path)
		for _, p := range res.Packages {
			for _, v := range p.Vulnerabilities {
				a := advisory{lock: lock, ecosystem: p.Package.Ecosystem, pkg: p.Package.Name, version: p.Package.Version, id: v.ID, aliases: sortedUnique(v.Aliases), fixes: fixedVersions(v, p.Package.Name, p.Package.Ecosystem)}
				for _, field := range []string{a.lock, a.ecosystem, a.pkg, a.version, a.id} {
					if !noSpace.MatchString(field) {
						return nil, fmt.Errorf("untrusted osv-scanner output: empty or spaced field in %s", a.key())
					}
				}
				out = append(out, a)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].key() < out[j].key() })
	return out, nil
}

func fixedVersions(v osvVuln, name, ecosystem string) []string {
	var fixes []string
	for _, aff := range v.Affected {
		if aff.Package.Name != name || aff.Package.Ecosystem != ecosystem {
			continue
		}
		for _, rg := range aff.Ranges {
			for _, e := range rg.Events {
				if e.Fixed != "" {
					fixes = append(fixes, e.Fixed)
				}
			}
		}
	}
	return sortedUnique(fixes)
}

func sortedUnique(v []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range v {
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

func oneLine(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 300 {
		s = s[:300]
	}
	return s
}

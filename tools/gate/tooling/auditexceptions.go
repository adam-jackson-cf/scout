package tooling

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	noFixFields   = []string{"advisory", "package", "version", "lock", "reason", "exposure", "expires"}
	blockedFields = []string{"blocked_by", "constraint", "constraint_source"}
	vendoredField = "vendored_in"
	pathFields    = []string{"path", "reason", "exposure", "expires"}
	constraintRe  = regexp.MustCompile(`^(\S+?)(?:==|@)(\S+) requires ([A-Za-z0-9@/._-]+?)((?:[=!<>~^]=?|===)\S+)$`)
	specPart      = regexp.MustCompile(`^(===|==|!=|~=|>=|<=|>|<|\^|~|=)?\s*(\S+)$`)
)

type auditException struct {
	fields map[string]string
	index  int
}

func (e auditException) key() string {
	return e.fields["lock"] + ":" + e.fields["package"] + "@" + e.fields["version"] + ":" + e.fields["advisory"]
}

func (e auditException) blocked() bool { return e.fields["blocked_by"] != "" }

func (e auditException) vendored() bool { return e.fields[vendoredField] != "" }

func (e auditException) excludesPath() bool { return e.fields["path"] != "" }

func (e auditException) expired(today string) bool { return e.fields["expires"] < today }

func loadExceptions(root string) ([]auditException, []auditFailure) {
	data, err := os.ReadFile(filepath.Join(root, exceptionsFile))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, []auditFailure{{"exception", exceptionsFile, oneLine(err.Error())}}
	}
	var raw []map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, []auditFailure{{"exception", exceptionsFile, "not a JSON array of entries: " + oneLine(err.Error())}}
	}
	var out []auditException
	var failures []auditFailure
	for i, entry := range raw {
		e, reason := validateEntry(entry)
		if reason != "" {
			failures = append(failures, auditFailure{"exception", exceptionsFile, fmt.Sprintf("entry %d: %s", i, reason)})
			continue
		}
		e.index = i
		out = append(out, e)
	}
	return out, failures
}

func validateEntry(entry map[string]any) (auditException, string) {
	fields, reason := stringFields(entry)
	if reason != "" {
		return auditException{}, reason
	}
	if got, exp := sortedKeys(fields), sortedCopy(entryShape(fields)); strings.Join(got, ",") != strings.Join(exp, ",") {
		return auditException{}, fmt.Sprintf("fields %v, want exactly %v", got, exp)
	}
	if _, err := time.Parse("2006-01-02", fields["expires"]); err != nil {
		return auditException{}, "expires " + fields["expires"] + " is not a YYYY-MM-DD date"
	}
	if src, ok := fields["constraint_source"]; ok && !httpsURL(src) {
		return auditException{}, "constraint_source " + src + " is not an absolute HTTPS URL"
	}
	if p, ok := fields["path"]; ok && (strings.ContainsAny(p, "*?[") || strings.HasSuffix(p, "/")) {
		return auditException{}, "path " + p + " is a glob or directory; name each fixture, vendored or frozen file exactly"
	}
	return auditException{fields: fields}, ""
}

func entryShape(fields map[string]string) []string {
	if _, ok := fields["path"]; ok {
		return pathFields
	}
	want := append([]string(nil), noFixFields...)
	if _, ok := fields["blocked_by"]; ok {
		return append(want, blockedFields...)
	}
	if _, ok := fields[vendoredField]; ok {
		return append(want, vendoredField)
	}
	return want
}

func splitExceptions(entries []auditException) (paths, advisories []auditException) {
	for _, e := range entries {
		if e.excludesPath() {
			paths = append(paths, e)
			continue
		}
		advisories = append(advisories, e)
	}
	return paths, advisories
}

func fileExclusions(entries []auditException, files []string, today string, explicit bool) (map[string]bool, []auditFailure) {
	present := map[string]bool{}
	for _, f := range files {
		present[f] = true
	}
	out := map[string]bool{}
	var failures []auditFailure
	for _, e := range entries {
		p := e.fields["path"]
		switch {
		case e.expired(today):
			failures = append(failures, auditFailure{"exception", exceptionsFile, fmt.Sprintf("entry %d (path %s): expired on %s", e.index, p, e.fields["expires"])})
		case !present[p] && !explicit:
			failures = append(failures, auditFailure{"exception", exceptionsFile, fmt.Sprintf("entry %d (path %s) matches no audited file; remove it", e.index, p)})
		default:
			out[p] = true
		}
	}
	return out, failures
}

func stringFields(entry map[string]any) (map[string]string, string) {
	fields := map[string]string{}
	for k, v := range entry {
		s, ok := v.(string)
		if !ok || strings.TrimSpace(s) == "" {
			return nil, "field " + k + " is not a nonblank string"
		}
		fields[k] = s
	}
	return fields, ""
}

func httpsURL(src string) bool {
	u, err := url.Parse(src)
	return err == nil && u.Scheme == "https" && u.Host != ""
}

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

type exceptionContext struct {
	locks      []lockFile
	root       string
	now        time.Time
	submodules map[string]bool
	explicit   bool
}

func applyExceptions(advisories []advisory, entries []auditException, ctx exceptionContext) ([]advisory, []auditFailure) {
	byKey := map[string]auditException{}
	for _, e := range entries {
		byKey[e.key()] = e
	}
	parseAs := map[string]string{}
	for _, l := range ctx.locks {
		parseAs[l.path] = l.parseAs
	}
	matched := map[string]bool{}
	var remaining []advisory
	var failures []auditFailure
	for _, a := range advisories {
		e, ok := byKey[a.key()]
		if !ok {
			remaining = append(remaining, a)
			continue
		}
		matched[a.key()] = true
		if reason := ctx.judge(e, a, parseAs[a.lock]); reason != "" {
			failures = append(failures, auditFailure{"exception", exceptionsFile, fmt.Sprintf("entry %d (%s): %s", e.index, a.key(), reason)})
			remaining = append(remaining, a)
		}
	}
	if ctx.explicit {
		return remaining, failures
	}
	for _, e := range entries {
		if !matched[e.key()] {
			failures = append(failures, auditFailure{"exception", exceptionsFile, fmt.Sprintf("entry %d (%s) matches no reported advisory; remove it", e.index, e.key())})
		}
	}
	return remaining, failures
}

func (ctx exceptionContext) judge(e auditException, a advisory, parseAs string) string {
	if e.expired(ctx.now.Format("2006-01-02")) {
		return "expired on " + e.fields["expires"]
	}
	switch {
	case e.blocked():
		return judgeBlocked(e, a, ctx.root, parseAs)
	case e.vendored():
		return judgeVendored(e, a, ctx.submodules)
	case len(a.fixes) > 0:
		return "the advisory has fixed versions " + strings.Join(a.fixes, ",") + " and the entry names no blocking dependent; upgrade instead"
	}
	return ""
}

func judgeVendored(e auditException, a advisory, submodules map[string]bool) string {
	dir := e.fields[vendoredField]
	if !submodules[dir] {
		return vendoredField + " " + dir + " is not a git submodule of this repository"
	}
	if !strings.HasPrefix(a.lock, dir+"/") {
		return "lock " + a.lock + " is not inside " + vendoredField + " " + dir
	}
	if len(a.fixes) == 0 {
		return "the advisory has no fixed version; use a no-fix entry without " + vendoredField
	}
	return ""
}

func judgeBlocked(e auditException, a advisory, root, parseAs string) string {
	at := strings.LastIndex(e.fields["blocked_by"], "@")
	if at <= 0 {
		return "blocked_by is not <package>@<exact-version>"
	}
	blocker, blockerVersion := e.fields["blocked_by"][:at], e.fields["blocked_by"][at+1:]
	if !lockHas(root, a.lock, parseAs, blocker, blockerVersion) {
		return "blocker " + e.fields["blocked_by"] + " is not in " + a.lock + " at that version"
	}
	m := constraintRe.FindStringSubmatch(e.fields["constraint"])
	if m == nil {
		return "constraint does not parse as '<blocker>==<version> requires <package><specifier>'"
	}
	if !strings.EqualFold(m[1], blocker) || m[2] != blockerVersion || !strings.EqualFold(m[3], a.pkg) {
		return "constraint names " + m[1] + "@" + m[2] + " requiring " + m[3] + ", not " + e.fields["blocked_by"] + " requiring " + a.pkg
	}
	if len(a.fixes) == 0 {
		return "the advisory has no fixed version; use a no-fix entry without blocking fields"
	}
	return compatibleFix(m[4], a.fixes)
}

func compatibleFix(spec string, fixes []string) string {
	for _, fix := range fixes {
		ok, err := satisfies(spec, fix)
		if err != nil {
			return "constraint specifier " + spec + ": " + err.Error()
		}
		if ok {
			return "fixed version " + fix + " satisfies " + spec + ", so the fix is not blocked; upgrade"
		}
	}
	return ""
}

func lockHas(root, lock, parseAs, pkg, version string) bool {
	data, err := os.ReadFile(filepath.Join(root, lock))
	if err != nil {
		return false
	}
	text := string(data)
	switch parseAs {
	case "requirements.txt":
		return regexp.MustCompile(`(?mi)^` + regexp.QuoteMeta(pkg) + `==` + regexp.QuoteMeta(version) + `(\s|\\|$)`).MatchString(text)
	case "package-lock.json":
		var l struct {
			Packages map[string]struct {
				Version string `json:"version"`
			} `json:"packages"`
		}
		return json.Unmarshal(data, &l) == nil && l.Packages["node_modules/"+pkg].Version == version
	case "go.mod":
		return regexp.MustCompile(`(?m)^\s*(require\s+)?` + regexp.QuoteMeta(pkg) + `\s+v?` + regexp.QuoteMeta(strings.TrimPrefix(version, "v")) + `\b`).MatchString(text)
	}
	return strings.Contains(text, pkg+"@"+version) || strings.Contains(text, pkg+"@npm:"+version)
}

func satisfies(spec, version string) (bool, error) {
	for _, part := range strings.Split(spec, ",") {
		m := specPart.FindStringSubmatch(strings.TrimSpace(part))
		if m == nil {
			return false, fmt.Errorf("cannot parse %q", part)
		}
		ok, err := satisfiesOne(m[1], m[2], version)
		if err != nil || !ok {
			return false, err
		}
	}
	return true, nil
}

func satisfiesOne(op, bound, version string) (bool, error) {
	c := compareVersions(version, bound)
	switch op {
	case "", "=", "==", "===":
		return c == 0, nil
	case "!=":
		return c != 0, nil
	case ">=":
		return c >= 0, nil
	case "<=":
		return c <= 0, nil
	case ">":
		return c > 0, nil
	case "<":
		return c < 0, nil
	case "~=", "~", "^":
		prefix := compatiblePrefix(op, strings.Split(bound, "."))
		return c >= 0 && (version == prefix || strings.HasPrefix(version, prefix+".")), nil
	}
	return false, fmt.Errorf("unsupported operator %q", op)
}

func compatiblePrefix(op string, parts []string) string {
	keep := len(parts) - 1
	switch op {
	case "^":
		keep = 1
		for keep < len(parts) && parts[keep-1] == "0" {
			keep++
		}
	case "~":
		keep = min(2, len(parts))
	}
	return strings.Join(parts[:max(keep, 1)], ".")
}

func compareVersions(a, b string) int {
	as, bs := versionParts(a), versionParts(b)
	for i := 0; i < len(as) || i < len(bs); i++ {
		var x, y string
		if i < len(as) {
			x = as[i]
		}
		if i < len(bs) {
			y = bs[i]
		}
		if x == y {
			continue
		}
		xi, xerr := strconv.Atoi(nonBlank(x))
		yi, yerr := strconv.Atoi(nonBlank(y))
		if xerr == nil && yerr == nil {
			if xi < yi {
				return -1
			}
			return 1
		}
		if x < y {
			return -1
		}
		return 1
	}
	return 0
}

func nonBlank(s string) string {
	if s == "" {
		return "0"
	}
	return s
}

func versionParts(v string) []string {
	return strings.FieldsFunc(strings.TrimPrefix(v, "v"), func(r rune) bool { return r == '.' || r == '-' || r == '+' })
}

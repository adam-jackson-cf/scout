package tooling

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"scout/internal/toolchain"
	"scout/tools/gate/check"
)

type ruffDiagnostic struct {
	Code     *string `json:"code"`
	Message  string  `json:"message"`
	Filename string  `json:"filename"`
	Location struct {
		Row int `json:"row"`
	} `json:"location"`
	URL *string `json:"url"`
}

func runRuff(r *check.Run) ([]check.Finding, error) {
	files, err := selectFiles(r, byExt(".py", ".pyi"))
	if err != nil {
		return nil, err
	}
	var findings []check.Finding
	if !r.Explicit {
		if findings, err = ruffPolicy(r); err != nil {
			return nil, err
		}
	}
	if len(files) == 0 {
		return findings, nil
	}
	bin, err := resolveTool(r, ruffTool)
	if err != nil {
		return nil, err
	}
	inv := invocation{dir: r.Root, env: toolEnv(r), timeout: 5 * time.Minute}
	config := filepath.Join(r.Root, "ruff.toml")
	lint, err := ruffLint(inv, bin, config, r.Root, files)
	if err != nil {
		return nil, err
	}
	format, err := ruffFormat(inv, bin, config, files)
	if err != nil {
		return nil, err
	}
	return append(append(findings, lint...), format...), nil
}

func ruffLint(inv invocation, bin, config, root string, files []string) ([]check.Finding, error) {
	res, err := runTool(inv, bin, append([]string{"check", "--config", config, "--no-cache", "--force-exclude", "--ignore-noqa", "--no-fix", "--output-format=json", "--"}, files...)...)
	if err != nil {
		return nil, err
	}
	lint, err := parseRuff(root, res.stdout)
	if err != nil || (res.exitCode != 0 && len(lint) == 0) {
		return nil, fmt.Errorf("ruff check exited %d: %v %s", res.exitCode, err, res.stderr)
	}
	return lint, nil
}

func ruffFormat(inv invocation, bin, config string, files []string) ([]check.Finding, error) {
	res, err := runTool(inv, bin, append([]string{"format", "--config", config, "--no-cache", "--force-exclude", "--check", "--output-format=concise", "--"}, files...)...)
	if err != nil {
		return nil, err
	}
	format := parseRuffFormat(res.stdout)
	if res.exitCode != 0 && len(format) == 0 {
		return nil, fmt.Errorf("ruff format exited %d: %s%s", res.exitCode, res.stdout, res.stderr)
	}
	return format, nil
}

func parseRuff(root, out string) ([]check.Finding, error) {
	var diags []ruffDiagnostic
	if err := json.Unmarshal([]byte(out), &diags); err != nil {
		return nil, fmt.Errorf("ruff report: %w", err)
	}
	findings := make([]check.Finding, 0, len(diags))
	for _, d := range diags {
		code := "syntax"
		if d.Code != nil {
			code = *d.Code
		}
		fix := "fix the code; inline noqa is ignored, so a justified exception goes in ruff.toml per-file-ignores with a rationale comment"
		if d.URL != nil {
			fix = "change the code as " + *d.URL + " describes; inline noqa is ignored"
		}
		findings = append(findings, check.Finding{File: relTo(root, d.Filename), Line: d.Location.Row, Rule: "RUFF." + code, Problem: d.Message, Fix: fix})
	}
	return findings, nil
}

var ruffUnformatted = regexp.MustCompile(`^(.+?):\d+:\d+: unformatted: `)

func parseRuffFormat(out string) []check.Finding {
	var findings []check.Finding
	for _, line := range strings.Split(out, "\n") {
		m := ruffUnformatted.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		findings = append(findings, check.Finding{File: m[1], Rule: "RUFF.FORMAT", Problem: "not formatted as ruff.toml [format] requires", Fix: "run ruff format " + m[1]})
	}
	return findings
}

func relTo(root, file string) string {
	if rel, err := filepath.Rel(root, file); err == nil && !strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(file)
}

func pythonMinor(root string) (string, error) {
	pins, err := toolchain.Manifest(root)
	if err != nil {
		return "", err
	}
	parts := strings.Split(pins["PYTHON_VERSION"], ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("%s PYTHON_VERSION %q is not major.minor[.patch]", toolchain.ManifestPath, pins["PYTHON_VERSION"])
	}
	return parts[0] + "." + parts[1], nil
}

var mypyLine = regexp.MustCompile(`^(.+?):(\d+): error: (.*?)(?:  \[([a-z-]+)\])?$`)

func runMypy(r *check.Run) ([]check.Finding, error) {
	var findings []check.Finding
	var files []string
	if r.Explicit {
		var err error
		if files, err = selectFiles(r, byExt(".py", ".pyi")); err != nil || len(files) == 0 {
			return nil, err
		}
	} else {
		policy, err := mypyPolicy(r)
		if err != nil {
			return nil, err
		}
		findings = append(findings, policy...)
	}
	bin, err := resolveTool(r, mypyTool)
	if err != nil {
		return nil, err
	}
	args := append([]string{"--config-file", filepath.Join(r.Root, "mypy.ini"), "--cache-dir", filepath.Join(r.TempDir, "mypy"), "--no-error-summary", "--no-pretty"}, files...)
	res, err := runTool(invocation{dir: r.Root, env: toolEnv(r), timeout: 5 * time.Minute}, bin, args...)
	if err != nil {
		return nil, err
	}
	typed := parseMypy(res.stdout)
	if res.exitCode != 0 && len(typed) == 0 {
		return nil, fmt.Errorf("mypy exited %d: %s%s", res.exitCode, res.stdout, res.stderr)
	}
	return append(findings, typed...), nil
}

func parseMypy(out string) []check.Finding {
	var findings []check.Finding
	for _, line := range strings.Split(out, "\n") {
		m := mypyLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		n, _ := strconv.Atoi(m[2])
		code := m[4]
		if code == "" {
			code = "error"
		}
		findings = append(findings, check.Finding{File: m[1], Line: n, Rule: "MYPY." + code, Problem: m[3], Fix: "fix the type error; a [mypy-<module>] override needs evidence in proofs.toml [python_quality.mypy_overrides]"})
	}
	return findings
}

type iniFile map[string]map[string]iniValue

type iniValue struct {
	value string
	line  int
}

func parseINI(text string) iniFile {
	p := iniParser{out: iniFile{}}
	for n, raw := range strings.Split(text, "\n") {
		p.line(n+1, raw)
	}
	return p.out
}

type iniParser struct {
	out     iniFile
	section string
	lastKey string
}

func (p *iniParser) line(n int, raw string) {
	line := strings.TrimSpace(raw)
	switch {
	case line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";"):
	case strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]"):
		p.open(strings.TrimSpace(line[1 : len(line)-1]))
	case p.section == "":
	case raw != strings.TrimLeft(raw, " \t") && p.lastKey != "":
		v := p.out[p.section][p.lastKey]
		v.value += "\n" + line
		p.out[p.section][p.lastKey] = v
	default:
		p.assign(n, line)
	}
}

func (p *iniParser) open(section string) {
	p.section, p.lastKey = section, ""
	if p.out[section] == nil {
		p.out[section] = map[string]iniValue{}
	}
}

func (p *iniParser) assign(n int, line string) {
	if key, value, ok := strings.Cut(line, "="); ok {
		p.lastKey = strings.TrimSpace(key)
		p.out[p.section][p.lastKey] = iniValue{value: strings.TrimSpace(value), line: n}
	}
}

func splitList(v string) []string {
	var out []string
	for _, f := range strings.FieldsFunc(v, func(r rune) bool { return r == ',' || r == ':' || r == '\n' }) {
		if f = strings.TrimSuffix(strings.TrimSpace(f), "/"); f != "" {
			out = append(out, f)
		}
	}
	return out
}

type pythonQuality struct {
	MypyOverrides map[string]string `toml:"mypy_overrides"`
}

func mypyPolicy(r *check.Run) ([]check.Finding, error) {
	data, err := os.ReadFile(filepath.Join(r.Root, mypyConfig))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", mypyConfig, err)
	}
	ini := parseINI(string(data))
	files, err := selectFiles(r, byExt(".py", ".pyi"))
	if err != nil {
		return nil, err
	}
	minor, err := pythonMinor(r.Root)
	if err != nil {
		return nil, err
	}
	var pq pythonQuality
	if r.Proofs != nil {
		if _, err := r.Proofs.Section("python_quality", &pq); err != nil {
			return nil, fmt.Errorf("proofs.toml [python_quality]: %w", err)
		}
	}
	return judgeMypyConfig(ini, files, minor, pq.MypyOverrides, func(p string) bool {
		_, err := os.Stat(filepath.Join(r.Root, p))
		return err == nil
	}), nil
}

func judgeMypyConfig(ini iniFile, files []string, minor string, evidence map[string]string, exists func(string) bool) []check.Finding {
	main, ok := ini["mypy"]
	if !ok {
		return []check.Finding{{File: mypyConfig, Rule: "MYPY.package-coverage", Problem: "no [mypy] section", Fix: "add [mypy] with files naming every maintained Python root"}}
	}
	out := mainSectionFindings(main, minor)
	roots := splitList(main["files"].value)
	out = append(out, rootFindings(main, roots, exists)...)
	out = append(out, fileFindings(files, roots)...)
	return append(out, overrideFindings(ini, evidence)...)
}

const mypyConfig = "mypy.ini"

func mypyFinding(line int, rule, problem, fix string) check.Finding {
	return check.Finding{File: mypyConfig, Line: line, Rule: rule, Problem: problem, Fix: fix}
}

func mainSectionFindings(main map[string]iniValue, minor string) []check.Finding {
	var out []check.Finding
	if v := main["python_version"]; v.value != minor {
		out = append(out, mypyFinding(v.line, "MYPY.SETTING.python_version", fmt.Sprintf("python_version %q differs from toolchain/versions.env PYTHON_VERSION %s", v.value, minor), "set python_version = "+minor))
	}
	if v := main["follow_imports"]; v.value != "normal" {
		out = append(out, mypyFinding(v.line, "MYPY.follow_imports", fmt.Sprintf("follow_imports %q; the global value must be normal", v.value), "set follow_imports = normal"))
	}
	if v, ok := main["exclude"]; ok {
		out = append(out, mypyFinding(v.line, "MYPY.SETTING.exclude", "exclude has no effect on the explicit files targets and hides drift", "remove exclude; scope comes from files"))
	}
	if v, ok := main["ignore_missing_imports"]; ok {
		out = append(out, mypyFinding(v.line, "MYPY.module-overrides", "global ignore_missing_imports", "remove it; scope a published-stub-less package in a named [mypy-<module>] section"))
	}
	return out
}

func rootFindings(main map[string]iniValue, roots []string, exists func(string) bool) []check.Finding {
	var out []check.Finding
	if len(roots) == 0 {
		out = append(out, mypyFinding(main["files"].line, "MYPY.package-coverage", "files names no maintained Python root", "list every maintained Python root in files"))
	}
	var dirRoots []string
	for _, root := range roots {
		if !exists(root) {
			out = append(out, mypyFinding(main["files"].line, "MYPY.package-coverage", "files names "+root+", which does not exist", "remove it or restore the root"))
		}
		if !strings.HasSuffix(root, ".py") && !strings.HasSuffix(root, ".pyi") {
			dirRoots = append(dirRoots, root)
		}
	}
	if paths := splitList(main["mypy_path"].value); strings.Join(sortedCopy(paths), ",") != strings.Join(sortedCopy(dirRoots), ",") {
		out = append(out, mypyFinding(main["mypy_path"].line, "MYPY.SETTING.mypy_path", fmt.Sprintf("mypy_path %v differs from the files directories %v", paths, dirRoots), "make mypy_path list exactly the directory roots in files"))
	}
	return out
}

func fileFindings(files, roots []string) []check.Finding {
	var out []check.Finding
	for _, f := range files {
		if strings.HasSuffix(f, ".pyi") {
			out = append(out, check.Finding{File: f, Rule: "MYPY.follow_imports", Problem: "in-repo type stub", Fix: "delete it; use the package's annotations or a published stub package"})
		}
		if !underAny(f, roots) {
			out = append(out, check.Finding{File: f, Rule: "MYPY.package-coverage", Problem: "maintained Python file outside every mypy.ini files root", Fix: "add its root to mypy.ini files and mypy_path"})
		}
	}
	return out
}

func overrideFindings(ini iniFile, evidence map[string]string) []check.Finding {
	var out []check.Finding
	for _, name := range sortedCopy(sectionNames(ini)) {
		module, ok := strings.CutPrefix(name, "mypy-")
		if !ok {
			continue
		}
		section := ini[name]
		line := firstLine(section)
		if v, ok := section["follow_imports"]; ok && v.value == "skip" {
			out = append(out, mypyFinding(v.line, "MYPY.follow_imports", "["+name+"] follow_imports = skip", "remove it; silence a vendored copy with follow_imports = silent"))
		}
		if _, ok := section["ignore_missing_imports"]; ok && strings.Contains(module, "*") {
			out = append(out, mypyFinding(line, "MYPY.module-overrides", "["+name+"] ignore_missing_imports on a wildcard module", "name the exact module"))
		}
		if strings.TrimSpace(evidence[module]) == "" {
			out = append(out, mypyFinding(line, "MYPY.module-overrides", "["+name+"] has no evidence in proofs.toml [python_quality.mypy_overrides]", "record why "+module+" needs it (no annotations and no published stubs) or remove the section"))
		}
	}
	return out
}

func sectionNames(ini iniFile) []string {
	out := make([]string, 0, len(ini))
	for name := range ini {
		out = append(out, name)
	}
	return out
}

func firstLine(section map[string]iniValue) int {
	line := 0
	for _, v := range section {
		if line == 0 || v.line < line {
			line = v.line
		}
	}
	return line
}

func underAny(f string, roots []string) bool {
	for _, r := range roots {
		if f == r || strings.HasPrefix(f, r+"/") {
			return true
		}
	}
	return false
}

func sortedCopy(v []string) []string {
	out := append([]string(nil), v...)
	sort.Strings(out)
	return out
}

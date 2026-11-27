package toolchain

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	ManifestPath  = "toolchain/versions.env"
	BootstrapPath = "scripts/bootstrap-tools.sh"
)

var guidanceDocs = []string{"README.md", "CONTRIBUTING.md", "AGENTS.md"}

type Problem struct {
	File    string
	Line    int
	Rule    string
	Problem string
	Fix     string
}

type Pin struct {
	Key   string
	Value string
	Line  int
}

func ReadManifest(root string) ([]Pin, error) {
	data, err := os.ReadFile(filepath.Join(root, ManifestPath))
	if err != nil {
		return nil, err
	}
	var pins []Pin
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for n := 1; scanner.Scan(); n++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || value == "" {
			return nil, fmt.Errorf("%s:%d: malformed line %q", ManifestPath, n, line)
		}
		pins = append(pins, Pin{Key: key, Value: value, Line: n})
	}
	return pins, scanner.Err()
}

func Manifest(root string) (map[string]string, error) {
	pins, err := ReadManifest(root)
	if err != nil {
		return nil, err
	}
	return pinValues(pins), nil
}

func ToolName(key string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSuffix(key, "_VERSION")), "_", "-")
}

func pinKey(name string) string {
	return strings.ToUpper(strings.ReplaceAll(name, "-", "_")) + "_VERSION"
}

type Tree struct {
	Root     string
	Files    []string
	ToolsDir string
}

func (t Tree) has(rel string) bool {
	for _, f := range t.Files {
		if f == rel {
			return true
		}
	}
	return false
}

func (t Tree) read(rel string) (string, error) {
	data, err := os.ReadFile(filepath.Join(t.Root, rel))
	return string(data), err
}

func (t Tree) withSuffix(suffix string) []string {
	var out []string
	for _, f := range t.Files {
		if strings.HasSuffix(f, suffix) {
			out = append(out, f)
		}
	}
	return out
}

func Check(t Tree) ([]Problem, error) {
	pins, err := ReadManifest(t.Root)
	if err != nil {
		return nil, err
	}
	var out []Problem
	for _, fn := range []func(Tree, []Pin) ([]Problem, error){lockVersions, dependencyInventory, installGuidance, packageManager} {
		p, err := fn(t, pins)
		if err != nil {
			return nil, err
		}
		out = append(out, p...)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].File != out[j].File {
			return out[i].File < out[j].File
		}
		return out[i].Line < out[j].Line
	})
	return out, nil
}

var requirementPin = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9._-]*)==(\S+)`)

type declared struct {
	file, name, version string
	line                int
}

func pythonPins(t Tree, rel string) ([]declared, error) {
	text, err := t.read(rel)
	if err != nil {
		return nil, err
	}
	var out []declared
	for n, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}
		m := requirementPin.FindStringSubmatch(line)
		if m == nil {
			out = append(out, declared{file: rel, name: line, line: n + 1})
			continue
		}
		out = append(out, declared{file: rel, name: strings.ToLower(m[1]), version: strings.TrimSuffix(m[2], "\\"), line: n + 1})
	}
	return out, nil
}

type packageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	PackageManager  string            `json:"packageManager"`
}

type packageLock struct {
	Packages map[string]struct {
		Version         string            `json:"version"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	} `json:"packages"`
}

func readJSON(t Tree, rel string, v any) error {
	text, err := t.read(rel)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(text), v); err != nil {
		return fmt.Errorf("%s: %w", rel, err)
	}
	return nil
}

func nodePins(t Tree, rel string) ([]declared, packageJSON, error) {
	var pkg packageJSON
	if err := readJSON(t, rel, &pkg); err != nil {
		return nil, pkg, err
	}
	var out []declared
	for _, deps := range []map[string]string{pkg.Dependencies, pkg.DevDependencies} {
		for name, version := range deps {
			out = append(out, declared{file: rel, name: name, version: version})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
	return out, pkg, nil
}

type problemSet struct {
	rule string
	list []Problem
}

func (p *problemSet) add(file string, line int, problem, fix string) {
	p.list = append(p.list, Problem{File: file, Line: line, Rule: p.rule, Problem: problem, Fix: fix})
}

func pinValues(pins []Pin) map[string]string {
	values := map[string]string{}
	for _, p := range pins {
		values[p.Key] = p.Value
	}
	return values
}

func lockVersions(t Tree, pins []Pin) ([]Problem, error) {
	values := pinValues(pins)
	out := &problemSet{rule: "PARITY.lock-version"}
	for _, in := range t.withSuffix(".in") {
		if err := pythonLockVersions(t, in, values, out); err != nil {
			return nil, err
		}
	}
	for _, pj := range t.withSuffix("package.json") {
		if path.Base(pj) != "package.json" || !t.has(path.Join(path.Dir(pj), "package-lock.json")) {
			continue
		}
		if err := nodeLockVersions(t, pj, values, out); err != nil {
			return nil, err
		}
	}
	return out.list, nil
}

func pythonLockVersions(t Tree, in string, values map[string]string, out *problemSet) error {
	lockRel := strings.TrimSuffix(in, ".in") + ".lock"
	decl, err := pythonPins(t, in)
	if err != nil {
		return err
	}
	lock, err := t.read(lockRel)
	if err != nil {
		out.add(in, 0, "no lock file "+lockRel+" beside the manifest", "generate it with scripts/bootstrap-tools.sh --relock")
		return nil
	}
	for _, d := range decl {
		if d.version == "" {
			out.add(in, d.line, fmt.Sprintf("requirement %q is not an exact == pin", d.name), "pin it as name==version")
			continue
		}
		if !regexp.MustCompile(`(?mi)^` + regexp.QuoteMeta(d.name) + `==` + regexp.QuoteMeta(d.version) + `(\s|\\|$)`).MatchString(lock) {
			out.add(lockRel, 0, fmt.Sprintf("does not lock %s==%s declared at %s:%d", d.name, d.version, in, d.line), "run scripts/bootstrap-tools.sh --relock")
		}
		manifestAgrees(in, d, values, out)
	}
	return nil
}

func manifestAgrees(file string, d declared, values map[string]string, out *problemSet) {
	if v, ok := values[pinKey(d.name)]; ok && v != d.version {
		out.add(file, d.line, fmt.Sprintf("%s %s differs from %s %s=%s", d.name, d.version, ManifestPath, pinKey(d.name), v), "make "+file+" and "+ManifestPath+" agree, then relock")
	}
}

var exactVersion = regexp.MustCompile(`^\d+(\.\d+)*([.-][0-9A-Za-z.-]+)?$`)

func nodeLockVersions(t Tree, pj string, values map[string]string, out *problemSet) error {
	decl, _, err := nodePins(t, pj)
	if err != nil {
		return err
	}
	lockRel := path.Join(path.Dir(pj), "package-lock.json")
	var lock packageLock
	if err := readJSON(t, lockRel, &lock); err != nil {
		return err
	}
	rootEntry := lock.Packages[""]
	for _, d := range decl {
		if !exactVersion.MatchString(d.version) {
			out.add(pj, 0, fmt.Sprintf("dependency %s %q is not an exact version", d.name, d.version), "pin the exact version")
		}
		locked := rootEntry.Dependencies[d.name] + rootEntry.DevDependencies[d.name]
		if locked != d.version {
			out.add(lockRel, 0, fmt.Sprintf("root entry declares %s %q, package.json %q", d.name, locked, d.version), "run scripts/bootstrap-tools.sh --relock")
		}
		if got := lock.Packages["node_modules/"+d.name].Version; got != d.version {
			out.add(lockRel, 0, fmt.Sprintf("locks %s %q, package.json %q", d.name, got, d.version), "run scripts/bootstrap-tools.sh --relock")
		}
		manifestAgrees(pj, d, values, out)
	}
	return nil
}

var requiredLoop = regexp.MustCompile(`(?m)^for key in ([A-Z0-9_ ]+); do$`)

func dependencyInventory(t Tree, pins []Pin) ([]Problem, error) {
	out := &problemSet{rule: "PARITY.dependency-inventory"}
	boot, err := t.read(BootstrapPath)
	if err != nil {
		return nil, err
	}
	loop := requiredLoop.FindStringSubmatch(boot)
	if loop == nil {
		out.add(BootstrapPath, 0, "no 'for key in ...; do' list of required toolchain/versions.env keys", "restore the required-key loop that rejects a missing pin")
		return out.list, nil
	}
	requiredKeys(strings.Fields(loop[1]), pins, out)
	installed, err := installedPackages(t)
	if err != nil {
		return nil, err
	}
	usage := strings.Replace(boot, loop[0], "", 1) + otherScripts(t)
	for _, p := range pins {
		used := regexp.MustCompile(`\b` + p.Key + `\b`).MatchString(usage)
		if _, ok := installed[ToolName(p.Key)]; !ok && !used {
			out.add(ManifestPath, p.Line, p.Key+" is pinned but nothing installs it", "install it in "+BootstrapPath+" or remove the pin")
		}
	}
	declaredKeys := map[string]bool{}
	for _, p := range pins {
		declaredKeys[p.Key] = true
	}
	for _, n := range sortedNames(installed) {
		if !declaredKeys[pinKey(n)] {
			out.add(installed[n], 0, fmt.Sprintf("%s is installed but %s has no %s", n, ManifestPath, pinKey(n)), "pin "+pinKey(n)+" in "+ManifestPath+" and add it to the required-key list")
		}
	}
	return out.list, nil
}

func requiredKeys(loop []string, pins []Pin, out *problemSet) {
	required := map[string]bool{}
	for _, k := range loop {
		required[k] = true
	}
	declared := map[string]bool{}
	for _, p := range pins {
		declared[p.Key] = true
		if !required[p.Key] {
			out.add(ManifestPath, p.Line, p.Key+" is not in the required-key list of "+BootstrapPath, "add "+p.Key+" to the 'for key in' list")
		}
	}
	for _, k := range loop {
		if !declared[k] {
			out.add(BootstrapPath, 0, "requires "+k+", which "+ManifestPath+" does not pin", "pin "+k+" in "+ManifestPath+" or drop it from the list")
		}
	}
}

func otherScripts(t Tree) string {
	var b strings.Builder
	for _, f := range t.Files {
		if strings.HasPrefix(f, "scripts/") && strings.HasSuffix(f, ".sh") && f != BootstrapPath {
			if text, err := t.read(f); err == nil {
				b.WriteString("\n" + text)
			}
		}
	}
	return b.String()
}

func installedPackages(t Tree) (map[string]string, error) {
	installed := map[string]string{}
	for _, in := range t.withSuffix(".in") {
		decl, err := pythonPins(t, in)
		if err != nil {
			return nil, err
		}
		for _, d := range decl {
			installed[d.name] = in
		}
	}
	for _, pj := range t.withSuffix("package.json") {
		decl, _, err := nodePins(t, pj)
		if err != nil {
			return nil, err
		}
		for _, d := range decl {
			installed[d.name] = pj
		}
	}
	return installed, nil
}

func sortedNames(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for n := range m {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

var (
	goVersionMention = regexp.MustCompile(`go@(\d+\.\d+(?:\.\d+)?)|badge/go-(\d+\.\d+(?:\.\d+)?)-`)
	scriptFlag       = regexp.MustCompile(`(scripts/[A-Za-z0-9_.-]+\.sh)((?:\s+--?[A-Za-z][A-Za-z0-9-]*)+)`)
	installSentence  = "scripts/bootstrap-tools.sh` installs"
)

func installGuidance(t Tree, pins []Pin) ([]Problem, error) {
	values := pinValues(pins)
	out := &problemSet{rule: "PARITY.install-guidance"}
	sentenceSeen := false
	for _, doc := range guidanceDocs {
		if !t.has(doc) {
			continue
		}
		text, err := t.read(doc)
		if err != nil {
			return nil, err
		}
		for n, line := range strings.Split(text, "\n") {
			goVersions(doc, n+1, line, values["GO_VERSION"], out)
			documentedFlags(t, doc, n+1, line, out)
			if strings.Contains(line, installSentence) {
				sentenceSeen = true
				installList(doc, n+1, line, pins, out)
			}
		}
	}
	if !sentenceSeen && t.has("README.md") {
		out.add("README.md", 0, "no sentence starting \"`scripts/bootstrap-tools.sh` installs\" lists the installed tools", "add one naming every tool pinned in "+ManifestPath)
	}
	return out.list, nil
}

func goVersions(doc string, line int, text, want string, out *problemSet) {
	for _, m := range goVersionMention.FindAllStringSubmatch(text, -1) {
		if v := m[1] + m[2]; v != want {
			out.add(doc, line, fmt.Sprintf("documents Go %s, toolchain/versions.env pins GO_VERSION=%s", v, want), "write "+want)
		}
	}
}

func documentedFlags(t Tree, doc string, line int, text string, out *problemSet) {
	for _, m := range scriptFlag.FindAllStringSubmatch(text, -1) {
		script, err := t.read(m[1])
		if err != nil {
			out.add(doc, line, "documents "+m[1]+", which does not exist", "name an existing script")
			continue
		}
		for _, flag := range strings.Fields(m[2]) {
			if !strings.Contains(script, flag+")") {
				out.add(doc, line, fmt.Sprintf("documents %s %s, which the script does not accept", m[1], flag), "use the flags its usage line lists")
			}
		}
	}
}

func installList(doc string, line int, text string, pins []Pin, out *problemSet) {
	lower := strings.ToLower(text)
	for _, p := range pins {
		name := ToolName(p.Key)
		if !regexp.MustCompile(`(^|[^a-z0-9-])` + regexp.QuoteMeta(name) + `([^a-z0-9-]|$)`).MatchString(lower) {
			out.add(doc, line, fmt.Sprintf("the install list omits %s (%s)", name, p.Key), "name "+name+" in the sentence that lists what scripts/bootstrap-tools.sh installs")
		}
	}
}

var managerLock = map[string]string{"npm": "package-lock.json", "pnpm": "pnpm-lock.yaml", "yarn": "yarn.lock", "bun": "bun.lock"}

var exactManager = regexp.MustCompile(`^(npm|pnpm|yarn|bun)@(\d+\.\d+\.\d+)$`)

func packageManager(t Tree, _ []Pin) ([]Problem, error) {
	out := &problemSet{rule: "PARITY.package-manager"}
	expected := map[string]bool{}
	for _, pj := range t.withSuffix("package.json") {
		if path.Base(pj) != "package.json" {
			continue
		}
		lock, err := declaredManager(t, pj, out)
		if err != nil {
			return nil, err
		}
		if lock != "" {
			expected[lock] = true
		}
	}
	for _, f := range t.Files {
		for _, lockName := range managerLock {
			if path.Base(f) == lockName && !expected[f] {
				out.add(f, 0, "lock file of a package manager no package.json declares", "delete it or declare that manager in packageManager")
			}
		}
	}
	return out.list, nil
}

func declaredManager(t Tree, pj string, out *problemSet) (string, error) {
	_, pkg, err := nodePins(t, pj)
	if err != nil {
		return "", err
	}
	m := exactManager.FindStringSubmatch(pkg.PackageManager)
	if m == nil {
		out.add(pj, 0, fmt.Sprintf("packageManager %q is not one exact <manager>@<x.y.z> pin", pkg.PackageManager), "declare \"packageManager\": \"npm@<version bundled with NODE_VERSION>\"")
		return "", nil
	}
	lock := path.Join(path.Dir(pj), managerLock[m[1]])
	if !t.has(lock) {
		out.add(pj, 0, "packageManager "+m[1]+" has no "+managerLock[m[1]]+" beside it", "generate it with scripts/bootstrap-tools.sh --relock")
	}
	if m[1] != "npm" || t.ToolsDir == "" {
		return lock, nil
	}
	bundled, err := bundledNpm(t.ToolsDir)
	if err != nil {
		return "", err
	}
	if bundled != m[2] {
		out.add(pj, 0, fmt.Sprintf("packageManager npm@%s, the pinned Node bundles npm %s", m[2], bundled), "write npm@"+bundled)
	}
	return lock, nil
}

func bundledNpm(tools string) (string, error) {
	var npm struct {
		Version string `json:"version"`
	}
	data, err := os.ReadFile(filepath.Join(tools, "node", "lib", "node_modules", "npm", "package.json"))
	if err != nil {
		return "", fmt.Errorf("read the npm bundled with the pinned Node: %w", err)
	}
	if err := json.Unmarshal(data, &npm); err != nil {
		return "", err
	}
	return npm.Version, nil
}

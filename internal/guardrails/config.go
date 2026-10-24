package guardrails

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// Tree is a read-only view of one revision of the repository.
type Tree interface {
	// Read returns the file content at a repo-relative slash path.
	Read(p string) ([]byte, bool)
	// Paths lists repo-relative slash paths (used to discover config files).
	Paths() []string
}

// MapTree is an in-memory Tree.
type MapTree map[string][]byte

// Read implements Tree.
func (m MapTree) Read(p string) ([]byte, bool) { b, ok := m[p]; return b, ok }

// Paths implements Tree.
func (m MapTree) Paths() []string {
	out := make([]string, 0, len(m))
	for p := range m {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

const (
	// IgnoresPrefix and StrictnessPrefix are the CONFIG check-id families; the
	// tool slug is appended.
	IgnoresPrefix    = "guardrails.CONFIG.ignores-added."
	StrictnessPrefix = "guardrails.CONFIG.strictness-lowered."
	configVersion    = "@1"
)

// facts is the effective, tool-level view of one config file: what is
// ignored, what is enforced, and how strictly.
type facts struct {
	// ignores holds "set|value" elements; growth is ignores-added.
	ignores map[string]bool
	// flags are effective booleans; true→false is strictness-lowered.
	flags map[string]bool
	// nums are thresholds; a decrease is strictness-lowered.
	nums map[string]float64
	// ranks are enum levels (higher is stricter); a decrease is lowered.
	ranks map[string]int
	// sev are rule severities (0 off, 1 warn, 2 error); →0 is ignores-added.
	sev map[string]int
	// enabled are sets whose shrinkage is strictness-lowered.
	enabled map[string]bool
	// unknown is set when the effective config could not be fully resolved
	// (unresolvable extends, opaque expressions).
	unknown bool
	// warnLowers makes error→warn a strictness-lowered finding.
	warnLowers bool
}

func newFacts() *facts {
	return &facts{
		ignores: map[string]bool{}, flags: map[string]bool{}, nums: map[string]float64{},
		ranks: map[string]int{}, sev: map[string]int{}, enabled: map[string]bool{},
	}
}

func (f *facts) addIgnores(set string, values []string) {
	for _, v := range values {
		f.ignores[set+"|"+v] = true
	}
}

// extractor reads the tool facts one file contributes at one revision.
type extractor func(t Tree, p string) map[string]*facts

type configKind struct {
	match   func(base, p string) bool
	extract extractor
}

var configKinds []configKind

func nameIn(names ...string) func(string, string) bool {
	return func(base, _ string) bool {
		for _, n := range names {
			if base == n {
				return true
			}
		}
		return false
	}
}

func init() {
	configKinds = []configKind{
		{nameIn(".golangci.yml", ".golangci.yaml"), extractGolangci},
		{nameIn("staticcheck.conf"), extractStaticcheck},
		{nameIn("pyproject.toml"), extractPyproject},
		{nameIn("ruff.toml", ".ruff.toml"), extractRuffToml},
		{nameIn(".flake8", "setup.cfg", "tox.ini"), extractINIFlake8},
		{nameIn("mypy.ini", ".mypy.ini", "setup.cfg", "tox.ini"), extractINIMypy},
		{nameIn(".coveragerc", "setup.cfg", "tox.ini"), extractINICoverage},
		{nameIn("pyrightconfig.json"), extractPyrightJSON},
		{func(b, _ string) bool { return strings.HasPrefix(b, ".eslintrc") || b == "package.json" }, extractESLintLegacy},
		{func(b, _ string) bool { return strings.HasPrefix(b, "eslint.config.") }, extractESLintFlat},
		{nameIn(".eslintignore"), extractESLintIgnore},
		{nameIn("biome.json", "biome.jsonc"), extractBiome},
		{func(b, _ string) bool { return strings.HasPrefix(b, "tsconfig") && strings.HasSuffix(b, ".json") }, extractTSConfig},
		{func(b, _ string) bool { return strings.HasPrefix(b, "jest.config.") || b == "package.json" }, extractJest},
	}
}

func skipConfigPath(p string) bool {
	for _, seg := range strings.Split(p, "/") {
		switch seg {
		case "node_modules", "vendor", ".git", "testdata", ".venv", "venv", "site-packages":
			return true
		}
	}
	return false
}

// ConfigChanges compares guardrail configuration between base and head and
// reports effective erosion: ignore sets that grew, strictness that dropped.
// Only tools configured in base are examined; a config new in head has no
// baseline. Findings carry base and head evidence.
func ConfigChanges(base, head Tree) []Finding {
	seen := map[string]bool{}
	var paths []string
	for _, t := range []Tree{base, head} {
		for _, p := range t.Paths() {
			if !seen[p] && !skipConfigPath(p) {
				seen[p] = true
				paths = append(paths, p)
			}
		}
	}
	sort.Strings(paths)
	// A tool whose facts vanish from every head config file (file deleted,
	// unparseable, section removed) has been disabled, not merely edited.
	headTools := map[string]bool{}
	for _, p := range paths {
		for _, kind := range configKinds {
			if kind.match(path.Base(p), p) {
				for tool := range kind.extract(head, p) {
					headTools[tool] = true
				}
			}
		}
	}
	var out []Finding
	for _, p := range paths {
		name := path.Base(p)
		for _, kind := range configKinds {
			if !kind.match(name, p) {
				continue
			}
			bf := kind.extract(base, p)
			if len(bf) == 0 {
				continue
			}
			hf := kind.extract(head, p)
			hc, _ := head.Read(p)
			tools := make([]string, 0, len(bf))
			for tool := range bf {
				tools = append(tools, tool)
			}
			sort.Strings(tools)
			for _, tool := range tools {
				if h := hf[tool]; h != nil {
					out = append(out, diffFacts(tool, p, bf[tool], h, string(hc))...)
				} else if !headTools[tool] {
					out = append(out, Finding{
						Path: p, Line: 1, CheckID: StrictnessPrefix + tool, Version: StrictnessPrefix + tool + configVersion,
						Mechanism: "config-removed:" + mechSlug(tool),
						Evidence:  Evidence{Base: tool + " configuration present", Head: tool + " configuration absent or unparseable"},
					})
				}
			}
		}
	}
	sortFindings(out)
	return out
}

func sortFindings(out []Finding) {
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		if a.CheckID != b.CheckID {
			return a.CheckID < b.CheckID
		}
		return a.Mechanism < b.Mechanism
	})
}

func lineOf(content, needle string) int {
	if needle == "" {
		return 1
	}
	for i, l := range strings.Split(content, "\n") {
		if strings.Contains(l, needle) {
			return i + 1
		}
	}
	return 1
}

func mechSlug(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', strings.ContainsRune("._-/:@", r):
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	out := b.String()
	if len(out) > 80 {
		out = out[:80]
	}
	return out
}

func sevName(n int) string {
	switch n {
	case 0:
		return "off"
	case 1:
		return "warn"
	}
	return "error"
}

func diffFacts(tool, p string, b, h *facts, headContent string) []Finding {
	var out []Finding
	add := func(check, mech, needle, baseEv, headEv string) {
		out = append(out, Finding{
			Path: p, Line: lineOf(headContent, needle), CheckID: check + tool, Version: check + tool + configVersion,
			Mechanism: mech + ":" + mechSlug(tool+":"+needle), Evidence: Evidence{Head: headEv, Base: baseEv},
		})
	}
	// ignores grew
	setCount := map[string]int{}
	for k := range b.ignores {
		set, _, _ := strings.Cut(k, "|")
		setCount[set]++
	}
	for _, k := range sortedKeys(h.ignores) {
		if b.ignores[k] {
			continue
		}
		set, val, _ := strings.Cut(k, "|")
		add(IgnoresPrefix, "ignores-added", val,
			fmt.Sprintf("%s: %q absent (%d entries)", set, val, setCount[set]),
			fmt.Sprintf("%s: +%s", set, val))
	}
	// rule severity
	for _, k := range sortedKeys(b.sev) {
		bs := b.sev[k]
		hs, ok := h.sev[k]
		if !ok {
			if h.unknown || b.unknown {
				continue
			}
			hs = 0
		}
		switch {
		case bs > 0 && hs == 0:
			add(IgnoresPrefix, "rule-off", k, fmt.Sprintf("%s: %s", k, sevName(bs)), fmt.Sprintf("%s: %s", k, sevName(hs)))
		case bs == 2 && hs == 1 && h.warnLowers:
			add(StrictnessPrefix, "severity", k, fmt.Sprintf("%s: %s", k, sevName(bs)), fmt.Sprintf("%s: %s", k, sevName(hs)))
		}
	}
	// boolean strictness
	for _, k := range sortedKeys(b.flags) {
		if b.flags[k] {
			if hv, ok := h.flags[k]; ok && !hv {
				add(StrictnessPrefix, "flag", k, k+" = true", k+" = false")
			}
		}
	}
	for _, k := range sortedKeys(b.nums) {
		if hv, ok := h.nums[k]; ok && hv < b.nums[k] {
			add(StrictnessPrefix, "threshold", k, fmt.Sprintf("%s = %g", k, b.nums[k]), fmt.Sprintf("%s = %g", k, hv))
		}
	}
	for _, k := range sortedKeys(b.ranks) {
		if hv, ok := h.ranks[k]; ok && hv < b.ranks[k] {
			add(StrictnessPrefix, "level", k, fmt.Sprintf("%s = %s", k, rankName(b.ranks[k])), fmt.Sprintf("%s = %s", k, rankName(hv)))
		}
	}
	if !h.unknown {
		for _, k := range sortedKeys(b.enabled) {
			if !h.enabled[k] {
				set, val, _ := strings.Cut(k, "|")
				add(StrictnessPrefix, "enabled-removed", val, fmt.Sprintf("%s: %s enabled", set, val), fmt.Sprintf("%s: %s removed", set, val))
			}
		}
	}
	return out
}

func rankName(n int) string {
	switch n {
	case 0:
		return "none/off"
	case 1:
		return "basic"
	case 2:
		return "standard"
	}
	return "strict/all"
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ChangesAll runs every guardrail mechanism between base and head: scoped and
// unscoped suppressions and focused tests over changed paths, and config
// erosion over the whole trees. Results are sorted.
func ChangesAll(base, head Tree, changed []string) []Finding {
	baseFiles, headFiles := map[string][]byte{}, map[string][]byte{}
	for _, p := range changed {
		if b, ok := head.Read(p); ok {
			headFiles[p] = b
		}
		if b, ok := base.Read(p); ok {
			baseFiles[p] = b
		}
	}
	var out []Finding
	if len(changed) > 0 {
		out = append(out, SuppressAdded(baseFiles, headFiles, changed)...)
		out = append(out, FocusedTestsAdded(baseFiles, headFiles, changed)...)
	}
	out = append(out, ConfigChanges(base, head)...)
	sortFindings(out)
	return out
}

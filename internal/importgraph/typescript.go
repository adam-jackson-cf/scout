package importgraph

import (
	"encoding/json"
	"path"
	"regexp"
	"sort"
	"strings"
)

var tsExts = []string{".ts", ".tsx", ".mts", ".cts", ".js", ".jsx", ".mjs", ".cjs"}

func isTSFile(f string) bool {
	if strings.HasSuffix(f, ".d.ts") {
		return false
	}
	for _, e := range tsExts {
		if strings.HasSuffix(f, e) {
			return true
		}
	}
	return false
}

type tsImport struct {
	spec     string
	typeOnly bool
}

// kw anchors a keyword at a token start (not after an identifier char, dot or quote).
const kw = "(?:^|[^\\w$.'\"])"

var (
	reFrom    = regexp.MustCompile(kw + `import\s+(type\s+)?([\w$\s,{}*]*?)\s*from\s*['"]([^'"]+)['"]`)
	reExport  = regexp.MustCompile(kw + `export\s+(type\s+)?(\*(?:\s+as\s+[\w$]+)?|\{[^}]*\})\s*from\s*['"]([^'"]+)['"]`)
	reBare    = regexp.MustCompile(kw + `import\s*['"]([^'"]+)['"]`)
	reRequire = regexp.MustCompile(kw + `require\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	reDynamic = regexp.MustCompile(kw + `import\s*\(\s*['"]([^'"]+)['"]\s*\)`)
)

// allInlineType reports whether a named-import clause `{ type A, type B }`
// has every specifier marked `type` (the import then emits no runtime edge).
func allInlineType(clause string) bool {
	i, j := strings.Index(clause, "{"), strings.LastIndex(clause, "}")
	if i < 0 || j < i {
		return false
	}
	if strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(clause[:i]), ",")) != "" {
		return false
	}
	n := 0
	for _, s := range strings.Split(clause[i+1:j], ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		n++
		if !strings.HasPrefix(s, "type ") {
			return false
		}
	}
	return n > 0
}

// stripComments blanks // and /* */ comments while keeping string contents.
func stripComments(src string) string {
	b := []byte(src)
	for i := 0; i < len(b); {
		c := b[i]
		switch {
		case c == '"' || c == '\'' || c == '`':
			q := c
			i++
			for i < len(b) && b[i] != q {
				if b[i] == '\\' {
					i++
				}
				if q != '`' && i < len(b) && b[i] == '\n' {
					break
				}
				i++
			}
			i++
		case c == '/' && i+1 < len(b) && b[i+1] == '/':
			for i < len(b) && b[i] != '\n' {
				b[i] = ' '
				i++
			}
		case c == '/' && i+1 < len(b) && b[i+1] == '*':
			for i < len(b) && !(b[i] == '*' && i+1 < len(b) && b[i+1] == '/') {
				if b[i] != '\n' {
					b[i] = ' '
				}
				i++
			}
			if i+1 < len(b) {
				b[i], b[i+1] = ' ', ' '
			}
			i += 2
		default:
			i++
		}
	}
	return string(b)
}

func parseTS(src string) []tsImport {
	s := stripComments(src)
	var out []tsImport
	for _, m := range reFrom.FindAllStringSubmatch(s, -1) {
		out = append(out, tsImport{m[3], m[1] != "" || allInlineType(m[2])})
	}
	for _, m := range reExport.FindAllStringSubmatch(s, -1) {
		out = append(out, tsImport{m[3], m[1] != ""})
	}
	for _, m := range reBare.FindAllStringSubmatch(s, -1) {
		out = append(out, tsImport{m[1], false})
	}
	for _, m := range reRequire.FindAllStringSubmatch(s, -1) {
		out = append(out, tsImport{m[1], false})
	}
	for _, m := range reDynamic.FindAllStringSubmatch(s, -1) {
		out = append(out, tsImport{m[1], false})
	}
	return out
}

type tsConfig struct {
	dir     string // repo-relative dir containing tsconfig.json
	baseURL string // repo-relative, "" if unset
	hasBase bool
	paths   map[string][]string
	pathsAt string // repo-relative dir that `paths` targets resolve against
}

var reTrailingComma = regexp.MustCompile(`,\s*([}\]])`)

func loadTSConfig(root, file string, seen map[string]bool) *tsConfig {
	dir := path.Dir(file)
	if dir == "." {
		dir = ""
	}
	if seen[file] {
		return nil
	}
	seen[file] = true
	b, err := readFile(root, file)
	if err != nil {
		return nil
	}
	var raw struct {
		Extends         json.RawMessage `json:"extends"` // string, or an array (TS 5)
		CompilerOptions struct {
			BaseURL *string             `json:"baseUrl"`
			Paths   map[string][]string `json:"paths"`
		} `json:"compilerOptions"`
	}
	clean := reTrailingComma.ReplaceAllString(stripComments(string(b)), "$1")
	if json.Unmarshal([]byte(clean), &raw) != nil {
		return nil
	}
	cfg := &tsConfig{dir: dir}
	var extends []string
	if len(raw.Extends) > 0 {
		var one string
		if json.Unmarshal(raw.Extends, &one) == nil {
			extends = []string{one}
		} else {
			_ = json.Unmarshal(raw.Extends, &extends)
		}
	}
	for _, e := range extends { // later entries override earlier ones
		if !strings.HasPrefix(e, ".") {
			continue
		}
		ext := path.Join(dir, e)
		if !strings.HasSuffix(ext, ".json") {
			ext += ".json"
		}
		if parent := loadTSConfig(root, ext, seen); parent != nil {
			if parent.hasBase {
				cfg.baseURL, cfg.hasBase = parent.baseURL, true
			}
			if parent.paths != nil {
				cfg.paths, cfg.pathsAt = parent.paths, parent.pathsAt
			}
		}
	}
	if raw.CompilerOptions.BaseURL != nil {
		cfg.baseURL, cfg.hasBase = path.Join(dir, *raw.CompilerOptions.BaseURL), true
	}
	if raw.CompilerOptions.Paths != nil {
		cfg.paths = raw.CompilerOptions.Paths
		cfg.pathsAt = dir
	}
	// `paths` targets resolve against baseUrl when one is in effect (even an
	// inherited `paths` after a child overrides baseUrl), else the defining dir.
	if cfg.paths != nil && cfg.hasBase {
		cfg.pathsAt = cfg.baseURL
	}
	return cfg
}

func buildTS(root string, files []string, es *edgeSet) ([]string, error) {
	fileSet := map[string]bool{}
	var ts []string
	for _, f := range files {
		fileSet[f] = true
		if isTSFile(f) {
			ts = append(ts, f)
		}
	}
	if len(ts) == 0 {
		return nil, nil
	}
	cfgCache := map[string]*tsConfig{}
	nearest := func(f string) *tsConfig {
		for d := path.Dir(f); ; d = path.Dir(d) {
			dd := d
			if dd == "." {
				dd = ""
			}
			if c, ok := cfgCache[dd]; ok {
				if c != nil {
					return c
				}
			} else {
				c := (*tsConfig)(nil)
				if fileSet[path.Join(dd, "tsconfig.json")] {
					c = loadTSConfig(root, path.Join(dd, "tsconfig.json"), map[string]bool{})
				}
				cfgCache[dd] = c
				if c != nil {
					return c
				}
			}
			if d == "." || d == "/" {
				return nil
			}
		}
	}
	// resolveFile tries p as file or directory index.
	resolveFile := func(p string) (string, bool) {
		if isTSFile(p) && fileSet[p] {
			return p, true
		}
		cands := []string{p}
		for _, pair := range [][2]string{{".js", ".ts"}, {".js", ".tsx"}, {".jsx", ".tsx"}, {".mjs", ".mts"}, {".cjs", ".cts"}} {
			if strings.HasSuffix(p, pair[0]) {
				cands = append(cands, strings.TrimSuffix(p, pair[0])+pair[1])
			}
		}
		for _, c := range cands {
			if isTSFile(c) && fileSet[c] {
				return c, true
			}
		}
		for _, e := range tsExts {
			if fileSet[p+e] && isTSFile(p+e) {
				return p + e, true
			}
		}
		for _, e := range tsExts {
			ix := path.Join(p, "index"+e)
			if fileSet[ix] {
				return ix, true
			}
		}
		return "", false
	}
	exists := func(p string) bool { // any repo file/dir at p (assets such as .css/.json)
		if fileSet[p] {
			return true
		}
		pre := p + "/"
		i := sort.SearchStrings(files, pre)
		return i < len(files) && strings.HasPrefix(files[i], pre)
	}
	for _, f := range ts {
		src, err := readFile(root, f)
		if err != nil {
			return nil, err
		}
		cfg := (*tsConfig)(nil)
		cfgLoaded := false
		for _, imp := range parseTS(string(src)) {
			spec := imp.spec
			if strings.HasPrefix(spec, ".") {
				p := path.Join(path.Dir(f), spec)
				if t, ok := resolveFile(p); ok {
					es.add(f, t, imp.typeOnly)
				} else if !exists(p) {
					es.unres[Unresolved{f, spec}] = true
				}
				continue
			}
			if !cfgLoaded {
				cfg, cfgLoaded = nearest(f), true
			}
			if cfg != nil {
				if t, matched, ok := resolveAlias(cfg, spec, resolveFile); matched {
					if ok {
						es.add(f, t, imp.typeOnly)
					} else {
						es.unres[Unresolved{f, spec}] = true
					}
					continue
				}
				if cfg.hasBase {
					if t, ok := resolveFile(path.Join(cfg.baseURL, spec)); ok {
						es.add(f, t, imp.typeOnly)
						continue
					}
				}
			}
			es.ext++
		}
	}
	return ts, nil
}

// resolveAlias matches spec against tsconfig paths (longest literal prefix wins).
func resolveAlias(cfg *tsConfig, spec string, resolve func(string) (string, bool)) (target string, matched, ok bool) {
	keys := make([]string, 0, len(cfg.paths))
	for k := range cfg.paths {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if len(keys[i]) != len(keys[j]) {
			return len(keys[i]) > len(keys[j])
		}
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		var capture string
		if star := strings.Index(k, "*"); star >= 0 {
			pre, post := k[:star], k[star+1:]
			if len(spec) < len(pre)+len(post) || !strings.HasPrefix(spec, pre) || !strings.HasSuffix(spec, post) {
				continue
			}
			capture = spec[len(pre) : len(spec)-len(post)]
		} else if k != spec {
			continue
		}
		for _, tpl := range cfg.paths[k] {
			p := strings.Replace(tpl, "*", capture, 1)
			if t, found := resolve(path.Join(cfg.pathsAt, p)); found {
				return t, true, true
			}
		}
		if k == "*" { // catch-all also covers third-party names: fall back to node resolution
			return "", false, false
		}
		return "", true, false
	}
	return "", false, false
}

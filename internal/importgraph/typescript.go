package importgraph

import (
	"encoding/json"
	"errors"
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
			i = skipStringLiteral(b, i)
		case startsWith(b, i, '/', '/'):
			i = blankLineComment(b, i)
		case startsWith(b, i, '/', '*'):
			i = blankBlockComment(b, i)
		default:
			i++
		}
	}
	return string(b)
}

func startsWith(b []byte, i int, first, second byte) bool {
	return b[i] == first && i+1 < len(b) && b[i+1] == second
}

func skipStringLiteral(b []byte, i int) int {
	q := b[i]
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
	return i + 1
}

func blankLineComment(b []byte, i int) int {
	for i < len(b) && b[i] != '\n' {
		b[i] = ' '
		i++
	}
	return i
}

func blankBlockComment(b []byte, i int) int {
	for i < len(b) && !startsWith(b, i, '*', '/') {
		if b[i] != '\n' {
			b[i] = ' '
		}
		i++
	}
	if i+1 < len(b) {
		b[i], b[i+1] = ' ', ' '
	}
	return i + 2
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

// configError is a tsconfig that was read but could not be used; the importing
// files are recorded as skipped. Read failures are returned as plain errors.
type configError struct{ msg string }

func (e *configError) Error() string { return e.msg }

func loadTSConfig(root, file string, seen map[string]bool) (*tsConfig, error) {
	dir := path.Dir(file)
	if dir == "." {
		dir = ""
	}
	if seen[file] {
		return nil, &configError{"tsconfig extends cycle at " + file}
	}
	seen[file] = true
	defer delete(seen, file) // seen is the current extends chain: a diamond is not a cycle
	raw, err := readTSConfigFile(root, file)
	if err != nil {
		return nil, err
	}
	extends, err := parseTSExtends(raw.Extends, file)
	if err != nil {
		return nil, err
	}
	cfg := &tsConfig{dir: dir}
	for _, e := range extends { // later entries override earlier ones
		if err := cfg.inherit(root, e, seen); err != nil {
			return nil, err
		}
	}
	cfg.apply(raw.CompilerOptions)
	return cfg, nil
}

type tsConfigFile struct {
	Extends         json.RawMessage   `json:"extends"`
	CompilerOptions tsCompilerOptions `json:"compilerOptions"`
}

type tsCompilerOptions struct {
	BaseURL *string             `json:"baseUrl"`
	Paths   map[string][]string `json:"paths"`
}

func readTSConfigFile(root, file string) (tsConfigFile, error) {
	var raw tsConfigFile
	b, err := readFile(root, file)
	if err != nil {
		return raw, err
	}
	clean := reTrailingComma.ReplaceAllString(stripComments(string(b)), "$1")
	if err := json.Unmarshal([]byte(clean), &raw); err != nil {
		return raw, &configError{"tsconfig " + file + " is not valid JSON: " + err.Error()}
	}
	return raw, nil
}

func parseTSExtends(raw json.RawMessage, file string) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var one string
	if json.Unmarshal(raw, &one) == nil {
		return []string{one}, nil
	}
	var extends []string
	if err := json.Unmarshal(raw, &extends); err != nil {
		return nil, &configError{"tsconfig " + file + " has an unsupported extends value: " + err.Error()}
	}
	return extends, nil
}

func (cfg *tsConfig) inherit(root, e string, seen map[string]bool) error {
	if !strings.HasPrefix(e, ".") {
		return nil
	}
	ext := path.Join(cfg.dir, e)
	if !strings.HasSuffix(ext, ".json") {
		ext += ".json"
	}
	parent, err := loadTSConfig(root, ext, seen)
	if err != nil {
		return err
	}
	if parent.hasBase {
		cfg.baseURL, cfg.hasBase = parent.baseURL, true
	}
	if parent.paths != nil {
		cfg.paths, cfg.pathsAt = parent.paths, parent.pathsAt
	}
	return nil
}

func (cfg *tsConfig) apply(opts tsCompilerOptions) {
	if opts.BaseURL != nil {
		cfg.baseURL, cfg.hasBase = path.Join(cfg.dir, *opts.BaseURL), true
	}
	if opts.Paths != nil {
		cfg.paths = opts.Paths
		cfg.pathsAt = cfg.dir
	}
	// `paths` targets resolve against baseUrl when one is in effect (even an
	// inherited `paths` after a child overrides baseUrl), else the defining dir.
	if cfg.paths != nil && cfg.hasBase {
		cfg.pathsAt = cfg.baseURL
	}
}

func buildTS(root string, files []string, es *edgeSet) ([]string, error) {
	r := &tsResolver{root: root, files: files, fileSet: map[string]bool{}, cfgCache: map[string]*loadedTSConfig{}}
	var ts []string
	for _, f := range files {
		r.fileSet[f] = true
		if isTSFile(f) {
			ts = append(ts, f)
		}
	}
	if len(ts) == 0 {
		return nil, nil
	}
	for _, f := range ts {
		if err := r.addFileEdges(es, f); err != nil {
			return nil, err
		}
	}
	return ts, nil
}

type loadedTSConfig struct {
	cfg *tsConfig
	bad *configError
}

type tsResolver struct {
	root     string
	files    []string
	fileSet  map[string]bool
	cfgCache map[string]*loadedTSConfig
	loadErr  error
}

func (r *tsResolver) addFileEdges(es *edgeSet, f string) error {
	src, err := readFile(r.root, f)
	if err != nil {
		return err
	}
	cfg, bad := r.nearest(f)
	if r.loadErr != nil {
		return r.loadErr
	}
	if bad != nil {
		es.skip(f, bad.msg)
	}
	for _, imp := range parseTS(string(src)) {
		r.addImport(es, f, cfg, imp)
	}
	return nil
}

// nearest returns the closest tsconfig above f. A config that exists but is
// unusable stops the search and is returned as bad.
func (r *tsResolver) nearest(f string) (*tsConfig, *configError) {
	for d := path.Dir(f); ; d = path.Dir(d) {
		dd := d
		if dd == "." {
			dd = ""
		}
		l := r.configIn(dd)
		if l.cfg != nil || l.bad != nil {
			return l.cfg, l.bad
		}
		if d == "." || d == "/" {
			return nil, nil
		}
	}
}

func (r *tsResolver) configIn(dd string) *loadedTSConfig {
	if l, ok := r.cfgCache[dd]; ok {
		return l
	}
	l := &loadedTSConfig{}
	if file := path.Join(dd, "tsconfig.json"); r.fileSet[file] {
		cfg, err := loadTSConfig(r.root, file, map[string]bool{})
		var ce *configError
		switch {
		case errors.As(err, &ce):
			l.bad = ce
		case err != nil:
			r.loadErr = err
		default:
			l.cfg = cfg
		}
	}
	r.cfgCache[dd] = l
	return l
}

// resolveFile tries p as file or directory index.
func (r *tsResolver) resolveFile(p string) (string, bool) {
	if c, ok := r.firstTSFile(tsRewriteCandidates(p)); ok {
		return c, true
	}
	if c, ok := r.firstTSFile(withSuffixes(p, tsExts)); ok {
		return c, true
	}
	for _, e := range tsExts {
		ix := path.Join(p, "index"+e)
		if r.fileSet[ix] {
			return ix, true
		}
	}
	return "", false
}

func tsRewriteCandidates(p string) []string {
	cands := []string{p}
	for _, pair := range [][2]string{{".js", ".ts"}, {".js", ".tsx"}, {".jsx", ".tsx"}, {".mjs", ".mts"}, {".cjs", ".cts"}} {
		if strings.HasSuffix(p, pair[0]) {
			cands = append(cands, strings.TrimSuffix(p, pair[0])+pair[1])
		}
	}
	return cands
}

func withSuffixes(p string, suffixes []string) []string {
	out := make([]string, len(suffixes))
	for i, s := range suffixes {
		out[i] = p + s
	}
	return out
}

func (r *tsResolver) firstTSFile(cands []string) (string, bool) {
	for _, c := range cands {
		if isTSFile(c) && r.fileSet[c] {
			return c, true
		}
	}
	return "", false
}

func (r *tsResolver) anyRepoPathAt(p string) bool {
	if r.fileSet[p] {
		return true
	}
	pre := p + "/"
	i := sort.SearchStrings(r.files, pre)
	return i < len(r.files) && strings.HasPrefix(r.files[i], pre)
}

func (r *tsResolver) addImport(es *edgeSet, f string, cfg *tsConfig, imp tsImport) {
	if strings.HasPrefix(imp.spec, ".") {
		p := path.Join(path.Dir(f), imp.spec)
		if t, ok := r.resolveFile(p); ok {
			es.add(f, t, imp.typeOnly)
		} else if !r.anyRepoPathAt(p) {
			es.unres[Unresolved{f, imp.spec}] = true
		}
		return
	}
	if cfg != nil && r.addConfigured(es, f, cfg, imp) {
		return
	}
	es.ext++
}

func (r *tsResolver) addConfigured(es *edgeSet, f string, cfg *tsConfig, imp tsImport) bool {
	if t, matched, ok := resolveAlias(cfg, imp.spec, r.resolveFile); matched {
		if ok {
			es.add(f, t, imp.typeOnly)
		} else {
			es.unres[Unresolved{f, imp.spec}] = true
		}
		return true
	}
	if !cfg.hasBase {
		return false
	}
	t, ok := r.resolveFile(path.Join(cfg.baseURL, imp.spec))
	if ok {
		es.add(f, t, imp.typeOnly)
	}
	return ok
}

// resolveAlias matches spec against tsconfig paths (longest literal prefix wins).
func resolveAlias(cfg *tsConfig, spec string, resolve func(string) (string, bool)) (target string, matched, ok bool) {
	for _, k := range aliasKeys(cfg.paths) {
		capture, hit := matchAlias(k, spec)
		if !hit {
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

func aliasKeys(paths map[string][]string) []string {
	keys := make([]string, 0, len(paths))
	for k := range paths {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if len(keys[i]) != len(keys[j]) {
			return len(keys[i]) > len(keys[j])
		}
		return keys[i] < keys[j]
	})
	return keys
}

func matchAlias(k, spec string) (string, bool) {
	star := strings.Index(k, "*")
	if star < 0 {
		return "", k == spec
	}
	pre, post := k[:star], k[star+1:]
	if len(spec) < len(pre)+len(post) || !strings.HasPrefix(spec, pre) || !strings.HasSuffix(spec, post) {
		return "", false
	}
	return spec[len(pre) : len(spec)-len(post)], true
}

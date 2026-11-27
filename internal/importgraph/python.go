package importgraph

import (
	"path"
	"sort"
	"strings"
)

type pyImport struct {
	level    int      // leading dots
	module   string   // dotted module (may be empty for "from . import x")
	names    []string // names for from-imports; nil for plain import
	typeOnly bool
}

// parsePython extracts import statements line-wise (no execution, no AST).
// Imports under `if TYPE_CHECKING:` are type-only.
func parsePython(src string) []pyImport {
	sc := &pyLineScanner{lines: strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n"), tcIndent: -1}
	var out []pyImport
	for i := 0; i < len(sc.lines); i++ {
		trim, ok := sc.importLine(sc.lines[i])
		if !ok {
			continue
		}
		var stmt string
		stmt, i = joinPyContinuation(sc.lines, i, stripPyComment(trim))
		out = append(out, parsePyStatements(stmt, sc.tcIndent >= 0)...)
	}
	return out
}

type pyLineScanner struct {
	lines    []string
	tcIndent int
	inTriple string
}

func (sc *pyLineScanner) importLine(raw string) (string, bool) {
	if sc.inTripleString(raw) {
		return "", false
	}
	trim := strings.TrimLeft(raw, " \t")
	if trim == "" || trim[0] == '#' {
		return "", false
	}
	indent := len(raw) - len(trim)
	if sc.opensTripleString(trim) || sc.entersTypeChecking(trim, indent) {
		return "", false
	}
	return trim, strings.HasPrefix(trim, "import ") || strings.HasPrefix(trim, "from ")
}

func (sc *pyLineScanner) inTripleString(raw string) bool {
	if sc.inTriple == "" {
		return false
	}
	if strings.Contains(raw, sc.inTriple) {
		sc.inTriple = ""
	}
	return true
}

func (sc *pyLineScanner) opensTripleString(trim string) bool {
	for _, q := range []string{`"""`, `'''`} {
		if c := strings.Count(trim, q); c%2 == 1 {
			sc.inTriple = q
		}
	}
	return sc.inTriple != ""
}

func (sc *pyLineScanner) entersTypeChecking(trim string, indent int) bool {
	if sc.tcIndent >= 0 && indent <= sc.tcIndent {
		sc.tcIndent = -1
	}
	if strings.HasPrefix(trim, "if TYPE_CHECKING:") || strings.HasPrefix(trim, "if typing.TYPE_CHECKING:") {
		sc.tcIndent = indent
		return true
	}
	return false
}

func joinPyContinuation(lines []string, i int, stmt string) (string, int) {
	// join continuation lines: backslash or open parenthesis
	for (strings.HasSuffix(stmt, "\\") || strings.Count(stmt, "(") > strings.Count(stmt, ")")) && i+1 < len(lines) {
		i++
		stmt = strings.TrimSuffix(stmt, "\\") + " " + stripPyComment(strings.TrimSpace(lines[i]))
	}
	return stmt, i
}

func parsePyStatements(stmt string, typeOnly bool) []pyImport {
	var out []pyImport
	// "a; b" — only first statement is an import we care about; split on ';'
	for _, part := range strings.Split(stmt, ";") {
		part = strings.TrimSpace(part)
		imps, ok := parsePyStmt(part)
		if !ok {
			continue
		}
		for _, imp := range imps {
			imp.typeOnly = typeOnly
			out = append(out, imp)
		}
	}
	return out
}

func stripPyComment(s string) string {
	if i := strings.Index(s, "#"); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}

func parsePyStmt(s string) ([]pyImport, bool) {
	if rest, ok := strings.CutPrefix(s, "import "); ok {
		return parsePlainPyImport(rest)
	}
	rest, ok := strings.CutPrefix(s, "from ")
	if !ok {
		return nil, false
	}
	return parseFromPyImport(rest)
}

func parsePlainPyImport(rest string) ([]pyImport, bool) {
	var out []pyImport
	for _, p := range strings.Split(rest, ",") {
		f := strings.Fields(p)
		if len(f) > 0 {
			out = append(out, pyImport{module: f[0]})
		}
	}
	return out, len(out) > 0
}

func parseFromPyImport(rest string) ([]pyImport, bool) {
	k := strings.Index(rest, " import ")
	if k < 0 {
		return nil, false
	}
	mod := strings.TrimSpace(rest[:k])
	names := strings.NewReplacer("(", " ", ")", " ").Replace(rest[k+len(" import "):])
	level := 0
	for level < len(mod) && mod[level] == '.' {
		level++
	}
	imp := pyImport{level: level, module: mod[level:]}
	for _, n := range strings.Split(names, ",") {
		if f := strings.Fields(n); len(f) > 0 {
			imp.names = append(imp.names, f[0])
		}
	}
	if len(imp.names) == 0 {
		return nil, false
	}
	return []pyImport{imp}, true
}

// detectPythonRoots returns "" plus parents of top-level packages (dirs with
// __init__.py whose parent has none), e.g. "src".
func detectPythonRoots(files []string) []string {
	hasInit := map[string]bool{}
	for _, f := range files {
		if path.Base(f) == "__init__.py" {
			hasInit[path.Dir(f)] = true
		}
	}
	roots := map[string]bool{"": true}
	for d := range hasInit {
		p := path.Dir(d)
		if p == "." {
			continue
		}
		if !hasInit[p] {
			roots[p] = true
		}
	}
	out := make([]string, 0, len(roots))
	for r := range roots {
		out = append(out, r)
	}
	sort.Strings(out)
	return out
}

func buildPython(opts Options, files []string, es *edgeSet) ([]string, error) {
	var py []string
	for _, f := range files {
		if strings.HasSuffix(f, ".py") {
			py = append(py, f)
		}
	}
	if len(py) == 0 {
		return nil, nil
	}
	roots := opts.PythonRoots
	if len(roots) == 0 {
		roots = detectPythonRoots(py)
	}
	r := newPyResolver(py, roots)
	for _, f := range py {
		src, err := readFile(opts.Root, f)
		if err != nil {
			return nil, err
		}
		for _, imp := range parsePython(string(src)) {
			r.addImport(es, f, imp)
		}
	}
	return py, nil
}

type pyResolver struct {
	roots    []string
	fileSet  map[string]bool
	topLevel map[string]bool
}

func newPyResolver(py, roots []string) *pyResolver {
	r := &pyResolver{roots: roots, fileSet: map[string]bool{}, topLevel: map[string]bool{}}
	for _, f := range py {
		r.fileSet[f] = true
	}
	for _, f := range py {
		for _, root := range roots {
			r.addTopLevel(root, f)
		}
	}
	return r
}

func (r *pyResolver) addTopLevel(root, f string) {
	rel, ok := relTo(root, f)
	if !ok {
		return
	}
	parts := strings.Split(rel, "/")
	if len(parts) == 1 {
		r.topLevel[strings.TrimSuffix(parts[0], ".py")] = true
	} else if r.fileSet[path.Join(root, parts[0], "__init__.py")] {
		r.topLevel[parts[0]] = true
	}
}

func (r *pyResolver) resolveMod(base, dotted string) (string, bool) {
	rel := strings.ReplaceAll(dotted, ".", "/")
	p := path.Join(base, rel)
	if dotted == "" {
		p = base
	}
	if dotted != "" && r.fileSet[p+".py"] {
		return p + ".py", true
	}
	if r.fileSet[path.Join(p, "__init__.py")] {
		return path.Join(p, "__init__.py"), true
	}
	return "", false
}

func (r *pyResolver) resolveAbs(dotted string) (string, bool) {
	parts := strings.Split(dotted, ".")
	for n := len(parts); n >= 1; n-- {
		for _, root := range r.roots {
			if f, ok := r.resolveMod(root, strings.Join(parts[:n], ".")); ok {
				return f, true
			}
		}
	}
	return "", false
}

func (r *pyResolver) addImport(es *edgeSet, f string, imp pyImport) {
	switch {
	case imp.level > 0:
		r.addRelative(es, f, imp)
	case imp.names == nil:
		r.addPlain(es, f, imp)
	default:
		r.addFrom(es, f, imp)
	}
}

func (r *pyResolver) addRelative(es *edgeSet, f string, imp pyImport) {
	base := path.Dir(f)
	for i := 1; i < imp.level; i++ {
		base = path.Dir(base)
	}
	if base == "." {
		base = ""
	}
	target, ok := r.resolveMod(base, imp.module)
	anyName := false
	for _, n := range imp.names {
		sub := n
		if imp.module != "" {
			sub = imp.module + "." + n
		}
		if t, ok2 := r.resolveMod(base, sub); ok2 {
			es.add(f, t, imp.typeOnly)
			anyName = true
		}
	}
	if anyName {
		return
	}
	if ok {
		es.add(f, target, imp.typeOnly)
	} else {
		es.unres[Unresolved{f, strings.Repeat(".", imp.level) + imp.module}] = true
	}
}

func (r *pyResolver) addPlain(es *edgeSet, f string, imp pyImport) {
	if strings.HasPrefix(imp.module, ".") {
		es.unres[Unresolved{f, imp.module}] = true
		return
	}
	if t, ok := r.resolveAbs(imp.module); ok {
		es.add(f, t, imp.typeOnly)
		return
	}
	r.unresolvedOrExternal(es, f, imp.module)
}

func (r *pyResolver) addFrom(es *edgeSet, f string, imp pyImport) {
	base, ok := r.resolveAbs(imp.module)
	if !ok {
		r.unresolvedOrExternal(es, f, imp.module)
		return
	}
	anyName := false
	for _, n := range imp.names {
		if t, ok2 := r.resolveAbs(imp.module + "." + n); ok2 && t != base {
			es.add(f, t, imp.typeOnly)
			anyName = true
		}
	}
	if !anyName {
		es.add(f, base, imp.typeOnly)
	}
}

func (r *pyResolver) unresolvedOrExternal(es *edgeSet, f, module string) {
	if r.topLevel[strings.Split(module, ".")[0]] {
		es.unres[Unresolved{f, module}] = true
	} else {
		es.ext++
	}
}

func relTo(root, f string) (string, bool) {
	if root == "" {
		return f, true
	}
	if strings.HasPrefix(f, root+"/") {
		return f[len(root)+1:], true
	}
	return "", false
}

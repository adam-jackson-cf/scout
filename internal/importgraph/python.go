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
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	var out []pyImport
	tcIndent := -1
	inTriple := ""
	for i := 0; i < len(lines); i++ {
		raw := lines[i]
		if inTriple != "" {
			if strings.Contains(raw, inTriple) {
				inTriple = ""
			}
			continue
		}
		trim := strings.TrimLeft(raw, " \t")
		if trim == "" || trim[0] == '#' {
			continue
		}
		indent := len(raw) - len(trim)
		for _, q := range []string{`"""`, `'''`} {
			if c := strings.Count(trim, q); c%2 == 1 {
				inTriple = q
			}
		}
		if inTriple != "" {
			continue
		}
		if tcIndent >= 0 && indent <= tcIndent {
			tcIndent = -1
		}
		if strings.HasPrefix(trim, "if TYPE_CHECKING:") || strings.HasPrefix(trim, "if typing.TYPE_CHECKING:") {
			tcIndent = indent
			continue
		}
		isImport := strings.HasPrefix(trim, "import ") || strings.HasPrefix(trim, "from ")
		if !isImport {
			continue
		}
		stmt := stripPyComment(trim)
		// join continuation lines: backslash or open parenthesis
		for (strings.HasSuffix(stmt, "\\") || strings.Count(stmt, "(") > strings.Count(stmt, ")")) && i+1 < len(lines) {
			i++
			stmt = strings.TrimSuffix(stmt, "\\") + " " + stripPyComment(strings.TrimSpace(lines[i]))
		}
		// "a; b" — only first statement is an import we care about; split on ';'
		for _, part := range strings.Split(stmt, ";") {
			part = strings.TrimSpace(part)
			if imps, ok := parsePyStmt(part); ok {
				for _, imp := range imps {
					imp.typeOnly = tcIndent >= 0
					out = append(out, imp)
				}
			}
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
		var out []pyImport
		for _, p := range strings.Split(rest, ",") {
			f := strings.Fields(p)
			if len(f) > 0 {
				out = append(out, pyImport{module: f[0]})
			}
		}
		return out, len(out) > 0
	}
	rest, ok := strings.CutPrefix(s, "from ")
	if !ok {
		return nil, false
	}
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
	fileSet := map[string]bool{}
	for _, f := range py {
		fileSet[f] = true
	}
	// resolveMod: dotted module relative to dir prefix -> file
	resolveMod := func(base, dotted string) (string, bool) {
		rel := strings.ReplaceAll(dotted, ".", "/")
		p := path.Join(base, rel)
		if dotted == "" {
			p = base
		}
		if dotted != "" && fileSet[p+".py"] {
			return p + ".py", true
		}
		if fileSet[path.Join(p, "__init__.py")] {
			return path.Join(p, "__init__.py"), true
		}
		return "", false
	}
	// top-level names known in each root (to tell external from unresolved)
	topLevel := map[string]bool{}
	for _, f := range py {
		for _, r := range roots {
			rel, ok := relTo(r, f)
			if !ok {
				continue
			}
			parts := strings.Split(rel, "/")
			if len(parts) == 1 {
				topLevel[strings.TrimSuffix(parts[0], ".py")] = true
			} else if fileSet[path.Join(r, parts[0], "__init__.py")] {
				topLevel[parts[0]] = true
			}
		}
	}
	resolveAbs := func(dotted string) (string, bool) {
		parts := strings.Split(dotted, ".")
		for n := len(parts); n >= 1; n-- {
			for _, r := range roots {
				if f, ok := resolveMod(r, strings.Join(parts[:n], ".")); ok {
					return f, true
				}
			}
		}
		return "", false
	}
	for _, f := range py {
		src, err := readFile(opts.Root, f)
		if err != nil {
			return nil, err
		}
		pkgDirOf := path.Dir(f) // package containing f (for __init__.py its own dir)
		for _, imp := range parsePython(string(src)) {
			if imp.level > 0 {
				base := pkgDirOf
				for i := 1; i < imp.level; i++ {
					base = path.Dir(base)
				}
				if base == "." {
					base = ""
				}
				target := ""
				ok := false
				if imp.module != "" {
					target, ok = resolveMod(base, imp.module)
				} else {
					target, ok = resolveMod(base, "")
				}
				if imp.names == nil {
					continue
				}
				anyName := false
				for _, n := range imp.names {
					sub := n
					if imp.module != "" {
						sub = imp.module + "." + n
					}
					if t, ok2 := resolveMod(base, sub); ok2 {
						es.add(f, t, imp.typeOnly)
						anyName = true
					}
				}
				if !anyName {
					if ok {
						es.add(f, target, imp.typeOnly)
					} else {
						es.unres[Unresolved{f, strings.Repeat(".", imp.level) + imp.module}] = true
					}
				}
				continue
			}
			if imp.names == nil { // import a.b.c
				if t, ok := resolveAbs(imp.module); ok {
					es.add(f, t, imp.typeOnly)
				} else if topLevel[strings.Split(imp.module, ".")[0]] {
					es.unres[Unresolved{f, imp.module}] = true
				} else {
					es.ext++
				}
				continue
			}
			// from a.b import x, y
			base, ok := resolveAbs(imp.module)
			if !ok {
				if topLevel[strings.Split(imp.module, ".")[0]] {
					es.unres[Unresolved{f, imp.module}] = true
				} else {
					es.ext++
				}
				continue
			}
			anyName := false
			for _, n := range imp.names {
				if t, ok2 := resolveAbs(imp.module + "." + n); ok2 && t != base {
					es.add(f, t, imp.typeOnly)
					anyName = true
				}
			}
			if !anyName {
				es.add(f, base, imp.typeOnly)
			}
		}
	}
	return py, nil
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

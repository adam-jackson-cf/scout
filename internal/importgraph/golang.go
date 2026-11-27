package importgraph

import (
	"go/parser"
	"go/token"
	"path"
	"strconv"
	"strings"
)

type goModule struct {
	dir  string // repo-relative dir of go.mod ("" for root)
	path string // module path
}

func parseModulePath(src []byte) string {
	for _, line := range strings.Split(string(src), "\n") {
		line = strings.TrimSpace(line)
		if rest, ok := strings.CutPrefix(line, "module"); ok && (rest == "" || rest[0] == ' ' || rest[0] == '\t') {
			rest = strings.TrimSpace(rest)
			if i := strings.Index(rest, "//"); i >= 0 {
				rest = strings.TrimSpace(rest[:i])
			}
			return strings.Trim(rest, "\"`")
		}
	}
	return ""
}

// buildGo adds package-level edges. Nodes are package directories ("." root).
func buildGo(root string, files []string, es *edgeSet) ([]string, error) {
	mods, err := goModules(root, files)
	if err != nil || len(mods) == 0 {
		return nil, err
	}
	goFiles, pkgs := goSources(files)
	r := goResolver{mods: mods, byImport: goImportIndex(mods, pkgs)}
	fset := token.NewFileSet()
	for _, f := range goFiles {
		if err := r.addFileEdges(fset, root, f, es); err != nil {
			return nil, err
		}
	}
	return sortedKeys(pkgs), nil
}

func goModules(root string, files []string) ([]goModule, error) {
	var mods []goModule
	for _, f := range files {
		if path.Base(f) != "go.mod" {
			continue
		}
		b, err := readFile(root, f)
		if err != nil {
			return nil, err
		}
		if mp := parseModulePath(b); mp != "" {
			d := path.Dir(f)
			if d == "." {
				d = ""
			}
			mods = append(mods, goModule{d, mp})
		}
	}
	return mods, nil
}

func goSources(files []string) ([]string, map[string]bool) {
	pkgs := map[string]bool{}
	var goFiles []string
	for _, f := range files {
		if strings.HasSuffix(f, ".go") && !strings.HasSuffix(f, "_test.go") {
			goFiles = append(goFiles, f)
			pkgs[pkgDir(f)] = true
		}
	}
	return goFiles, pkgs
}

func goImportIndex(mods []goModule, pkgs map[string]bool) map[string]string {
	byImport := map[string]string{}
	for p := range pkgs {
		if m, ok := owner(mods, p); ok {
			byImport[m.importPath(p)] = p
		}
	}
	return byImport
}

type goResolver struct {
	mods     []goModule
	byImport map[string]string
}

func (r goResolver) addFileEdges(fset *token.FileSet, root, f string, es *edgeSet) error {
	src, err := readFile(root, f)
	if err != nil {
		return err
	}
	from := pkgDir(f)
	af, err := parser.ParseFile(fset, f, src, parser.ImportsOnly)
	if err != nil {
		es.skip(from, "go parse error in "+f+": "+err.Error())
		if af == nil {
			return nil
		}
	}
	for _, imp := range af.Imports {
		r.addImport(from, f, imp.Path.Value, es)
	}
	return nil
}

func (r goResolver) addImport(from, f, literal string, es *edgeSet) {
	ip, err := strconv.Unquote(literal)
	if err != nil {
		es.skip(from, "go import path in "+f+" is not a valid string literal: "+literal)
		return
	}
	if to, ok := r.byImport[ip]; ok {
		es.add(from, to, false)
		return
	}
	if r.inModule(ip) {
		es.unres[Unresolved{from, ip}] = true
	} else {
		es.ext++
	}
}

func (r goResolver) inModule(ip string) bool {
	for _, m := range r.mods {
		if ip == m.path || strings.HasPrefix(ip, m.path+"/") {
			return true
		}
	}
	return false
}

func pkgDir(f string) string { return path.Dir(f) }

// owner returns the module with the longest dir prefix containing dir.
func owner(mods []goModule, dir string) (goModule, bool) {
	best, ok := goModule{}, false
	for _, m := range mods {
		in := m.dir == "" || dir == m.dir || strings.HasPrefix(dir, m.dir+"/")
		if in && (!ok || len(m.dir) > len(best.dir)) {
			best, ok = m, true
		}
	}
	return best, ok
}

func (m goModule) importPath(dir string) string {
	rel := dir
	if m.dir != "" {
		rel = strings.TrimPrefix(strings.TrimPrefix(dir, m.dir), "/")
	}
	if rel == "." || rel == "" {
		return m.path
	}
	return m.path + "/" + rel
}

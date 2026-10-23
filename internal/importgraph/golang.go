package importgraph

import (
	"go/parser"
	"go/token"
	"path"
	"sort"
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
	var mods []goModule
	for _, f := range files {
		if path.Base(f) == "go.mod" {
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
	}
	if len(mods) == 0 {
		return nil, nil
	}
	pkgs := map[string]bool{}
	var goFiles []string
	for _, f := range files {
		if strings.HasSuffix(f, ".go") && !strings.HasSuffix(f, "_test.go") {
			goFiles = append(goFiles, f)
			pkgs[pkgDir(f)] = true
		}
	}
	// import path -> package dir
	byImport := map[string]string{}
	for p := range pkgs {
		if m, ok := owner(mods, p); ok {
			byImport[m.importPath(p)] = p
		}
	}
	fset := token.NewFileSet()
	for _, f := range goFiles {
		src, err := readFile(root, f)
		if err != nil {
			return nil, err
		}
		af, err := parser.ParseFile(fset, f, src, parser.ImportsOnly)
		if err != nil && af == nil {
			continue
		}
		from := pkgDir(f)
		for _, imp := range af.Imports {
			ip, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			if to, ok := byImport[ip]; ok {
				es.add(from, to, false)
				continue
			}
			internal := false
			for _, m := range mods {
				if ip == m.path || strings.HasPrefix(ip, m.path+"/") {
					internal = true
				}
			}
			if internal {
				es.unres[Unresolved{from, ip}] = true
			} else {
				es.ext++
			}
		}
	}
	out := make([]string, 0, len(pkgs))
	for p := range pkgs {
		out = append(out, p)
	}
	sort.Strings(out)
	return out, nil
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

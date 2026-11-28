package structure

import (
	"go/parser"
	"go/token"
	"path"
	"sort"
	"strings"

	"scout/tools/gate/pathscope"
)

type edge struct {
	Target string
	Line   int
}

func imports(file string, src []byte, module string) ([]edge, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, src, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	var out []edge
	for _, imp := range f.Imports {
		p := strings.Trim(imp.Path.Value, "\"")
		if !strings.HasPrefix(p, module+"/") {
			continue
		}
		out = append(out, edge{Target: strings.TrimPrefix(p, module+"/"), Line: fset.Position(imp.Pos()).Line})
	}
	return out, nil
}

func goSource(p string, fixture bool) bool {
	if !strings.HasSuffix(p, ".go") {
		return false
	}
	if fixture {
		return true
	}
	for _, s := range strings.Split(p, "/") {
		if s == "testdata" || s == "vendor" {
			return false
		}
	}
	return true
}

type graph struct {
	readers  map[string]map[string]bool
	files    map[string]string
	excluded map[string]bool
}

func buildGraph(s *pathscope.Scope, module string) (*graph, error) {
	g := &graph{readers: map[string]map[string]bool{}, files: map[string]string{}, excluded: map[string]bool{}}
	files, err := s.HeadFiles()
	if err != nil {
		return nil, err
	}
	wanted := graphSources(s, files)
	srcs, err := s.ReadHeadMany(wanted)
	if err != nil {
		return nil, err
	}
	for _, f := range wanted {
		if s.Fixture == "" && s.Excluded(f) {
			g.excluded[path.Dir(f)] = true
		}
		if err := g.add(f, srcs[f], module); err != nil {
			return nil, err
		}
	}
	return g, nil
}

func graphSources(s *pathscope.Scope, files []string) []string {
	var wanted []string
	for _, f := range files {
		if goSource(f, s.Fixture != "") && !strings.HasSuffix(f, "_test.go") {
			wanted = append(wanted, f)
		}
	}
	return wanted
}

func (g *graph) add(f string, src []byte, module string) error {
	dir := path.Dir(f)
	if _, ok := g.files[dir]; !ok {
		g.files[dir] = f
	}
	es, err := imports(f, src, module)
	if err != nil {
		return err
	}
	for _, e := range es {
		if e.Target == dir {
			continue
		}
		if g.readers[e.Target] == nil {
			g.readers[e.Target] = map[string]bool{}
		}
		g.readers[e.Target][dir] = true
	}
	return nil
}

func (g *graph) tierReaders(c Config, tier int) []string {
	set := map[string]bool{}
	for target, rs := range g.readers {
		if c.tierOf(target) != tier {
			continue
		}
		for r := range rs {
			if c.tierOf(r) != tier {
				set[r] = true
			}
		}
	}
	return sortedKeys(set)
}

func (g *graph) populated(c Config, tier int) []string {
	var out []string
	for dir := range g.files {
		if c.tierOf(dir) == tier {
			out = append(out, dir)
		}
	}
	sort.Strings(out)
	return out
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

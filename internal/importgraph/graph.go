// Package importgraph extracts internal import edges (Go, Python, TypeScript/JS)
// and computes scout.cycle_mass@1 (design §8.2). Nodes are repo-relative slash
// paths: source files for Python/TS, package directories for Go. Output is
// deterministic (sorted). Type-only edges are kept apart from runtime edges.
package importgraph

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Edge is a directed internal import From -> To between two nodes.
type Edge struct {
	From string
	To   string
}

// Unresolved is an import that looks internal (relative, aliased, or rooted in
// a package that exists in the repo) but whose target could not be found.
type Unresolved struct {
	From string
	Spec string
}

// Graph is the extracted internal import graph.
type Graph struct {
	Nodes         []string     // sorted node paths that are source files/packages
	Edges         []Edge       // runtime internal edges, deduped, sorted, no self edges
	TypeOnlyEdges []Edge       // type-only edges not also present as runtime edges; excluded from cycles
	Unresolved    []Unresolved // sorted
	External      int          // imports of third-party/stdlib packages (not internal)
	Languages     []string     // languages with at least one source node, sorted
}

// Options controls extraction.
type Options struct {
	Root string
	// Files optionally restricts the population to these repo-relative slash
	// paths (the eligible-source population). When nil, Root is walked with
	// default exclusions (vendored/generated dirs, tests).
	Files []string
	// PythonRoots are repo-relative source roots; when empty they are detected.
	PythonRoots []string
}

var skipDirs = map[string]bool{
	"node_modules": true, "vendor": true, ".git": true, "dist": true, "build": true,
	"testdata": true, "__pycache__": true, "venv": true, ".venv": true, "site-packages": true,
	".tox": true, ".mypy_cache": true, ".codegraph": true, "coverage": true,
}

func isTestPath(p string) bool {
	base := filepath.Base(p)
	if strings.HasSuffix(base, "_test.go") || strings.HasPrefix(base, "test_") && strings.HasSuffix(base, ".py") ||
		strings.HasSuffix(base, "_test.py") || base == "conftest.py" {
		return true
	}
	for _, m := range []string{".test.", ".spec."} {
		if strings.Contains(base, m) {
			return true
		}
	}
	for _, seg := range strings.Split(filepath.ToSlash(filepath.Dir(p)), "/") {
		if seg == "tests" || seg == "test" || seg == "__tests__" {
			return true
		}
	}
	return false
}

func walkFiles(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if p != root && (skipDirs[name] || strings.HasPrefix(name, ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !isTestPath(rel) {
			out = append(out, rel)
		}
		return nil
	})
	sort.Strings(out)
	return out, err
}

// edgeSet accumulates deduped runtime and type-only edges.
type edgeSet struct {
	runtime map[Edge]bool
	typed   map[Edge]bool
	unres   map[Unresolved]bool
	ext     int
}

func newEdgeSet() *edgeSet {
	return &edgeSet{runtime: map[Edge]bool{}, typed: map[Edge]bool{}, unres: map[Unresolved]bool{}}
}

func (s *edgeSet) add(from, to string, typeOnly bool) {
	if from == to {
		return
	}
	if typeOnly {
		s.typed[Edge{from, to}] = true
	} else {
		s.runtime[Edge{from, to}] = true
	}
}

func sortEdges(m map[Edge]bool, skip map[Edge]bool) []Edge {
	out := make([]Edge, 0, len(m))
	for e := range m {
		if !skip[e] {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		return out[i].To < out[j].To
	})
	return out
}

// Build extracts the import graph for all supported languages under opts.Root.
func Build(opts Options) (*Graph, error) {
	files := opts.Files
	if files == nil {
		var err error
		if files, err = walkFiles(opts.Root); err != nil {
			return nil, err
		}
	} else {
		files = append([]string(nil), files...)
		sort.Strings(files)
	}
	es := newEdgeSet()
	nodes := map[string]bool{}
	langs := map[string]bool{}

	if n, err := buildGo(opts.Root, files, es); err != nil {
		return nil, err
	} else if len(n) > 0 {
		langs["go"] = true
		for _, x := range n {
			nodes[x] = true
		}
	}
	if n, err := buildPython(opts, files, es); err != nil {
		return nil, err
	} else if len(n) > 0 {
		langs["python"] = true
		for _, x := range n {
			nodes[x] = true
		}
	}
	if n, err := buildTS(opts.Root, files, es); err != nil {
		return nil, err
	} else if len(n) > 0 {
		langs["typescript"] = true
		for _, x := range n {
			nodes[x] = true
		}
	}

	g := &Graph{External: es.ext}
	for n := range nodes {
		g.Nodes = append(g.Nodes, n)
	}
	sort.Strings(g.Nodes)
	for l := range langs {
		g.Languages = append(g.Languages, l)
	}
	sort.Strings(g.Languages)
	g.Edges = sortEdges(es.runtime, nil)
	g.TypeOnlyEdges = sortEdges(es.typed, es.runtime)
	for u := range es.unres {
		g.Unresolved = append(g.Unresolved, u)
	}
	sort.Slice(g.Unresolved, func(i, j int) bool {
		if g.Unresolved[i].From != g.Unresolved[j].From {
			return g.Unresolved[i].From < g.Unresolved[j].From
		}
		return g.Unresolved[i].Spec < g.Unresolved[j].Spec
	})
	return g, nil
}

// Result is scout.cycle_mass@1 for a graph (or a scope's share of it).
type Result struct {
	CycleEdges    int // internal runtime edges on a directed cycle
	InternalEdges int // internal runtime edges
	CrossBoundary int // non-trivial cycle components spanning more than one scope
	Unresolved    int
	TypeOnly      int
	cycleEdge     map[Edge]bool
}

// CycleMass computes cycle edges (both endpoints in one strongly connected
// component of size > 1), internal edges, cross-boundary cycle components
// (components whose nodes lie in more than one scope per scopeOf; nil scopeOf
// means a single scope) and the unresolved count.
func CycleMass(g *Graph, scopeOf func(node string) string) Result {
	idx := map[string]int{}
	for i, n := range g.Nodes {
		idx[n] = i
	}
	// edges may reference nodes absent from Nodes only if caller built Graph by hand
	adj := make([][]int, len(g.Nodes))
	for _, e := range g.Edges {
		f, ok1 := idx[e.From]
		t, ok2 := idx[e.To]
		if ok1 && ok2 {
			adj[f] = append(adj[f], t)
		}
	}
	comp := tarjan(adj)
	size := map[int]int{}
	for _, c := range comp {
		size[c]++
	}
	res := Result{InternalEdges: len(g.Edges), Unresolved: len(g.Unresolved), TypeOnly: len(g.TypeOnlyEdges), cycleEdge: map[Edge]bool{}}
	for _, e := range g.Edges {
		f, ok1 := idx[e.From]
		t, ok2 := idx[e.To]
		if ok1 && ok2 && comp[f] == comp[t] && size[comp[f]] > 1 {
			res.CycleEdges++
			res.cycleEdge[e] = true
		}
	}
	scopes := map[int]map[string]bool{}
	for i, n := range g.Nodes {
		c := comp[i]
		if size[c] < 2 {
			continue
		}
		if scopes[c] == nil {
			scopes[c] = map[string]bool{}
		}
		s := ""
		if scopeOf != nil {
			s = scopeOf(n)
		}
		scopes[c][s] = true
	}
	for _, s := range scopes {
		if len(s) > 1 {
			res.CrossBoundary++
		}
	}
	return res
}

// tarjan returns the SCC id of every vertex (iterative, deterministic).
func tarjan(adj [][]int) []int {
	n := len(adj)
	index := make([]int, n)
	low := make([]int, n)
	on := make([]bool, n)
	comp := make([]int, n)
	for i := range index {
		index[i] = -1
	}
	var stack []int
	next, ncomp := 0, 0
	type frame struct{ v, i int }
	for s := range n {
		if index[s] != -1 {
			continue
		}
		work := []frame{{s, 0}}
		index[s], low[s] = next, next
		next++
		stack = append(stack, s)
		on[s] = true
		for len(work) > 0 {
			f := &work[len(work)-1]
			v := f.v
			if f.i < len(adj[v]) {
				w := adj[v][f.i]
				f.i++
				if index[w] == -1 {
					index[w], low[w] = next, next
					next++
					stack = append(stack, w)
					on[w] = true
					work = append(work, frame{w, 0})
				} else if on[w] && index[w] < low[v] {
					low[v] = index[w]
				}
				continue
			}
			if low[v] == index[v] {
				for {
					w := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					on[w] = false
					comp[w] = ncomp
					if w == v {
						break
					}
				}
				ncomp++
			}
			work = work[:len(work)-1]
			if len(work) > 0 {
				p := work[len(work)-1].v
				if low[v] < low[p] {
					low[p] = low[v]
				}
			}
		}
	}
	return comp
}

func readFile(root, rel string) ([]byte, error) {
	return os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
}

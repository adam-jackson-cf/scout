// Package importgraph extracts internal import edges (Go, Python, TypeScript/JS)
// and computes scout.cycle_mass@1 (REQ-106). Nodes are repo-relative slash
// paths: source files for Python/TS, package directories for Go. Output is
// deterministic (sorted). Type-only edges are kept apart from runtime edges.
package importgraph

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
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

// Skipped is an input the extractors could not use. Its edges are absent from
// the graph, so a nonzero count means the graph undercounts.
type Skipped struct {
	From   string // importing node the skip affects
	Reason string
}

// Graph is the extracted internal import graph.
type Graph struct {
	Nodes         []string     // sorted node paths that are source files/packages
	Edges         []Edge       // runtime internal edges, deduped, sorted, no self edges
	TypeOnlyEdges []Edge       // type-only edges not also present as runtime edges; excluded from cycles
	Unresolved    []Unresolved // sorted
	Skipped       []Skipped    // inputs that could not be parsed or loaded; their edges are missing; sorted
	External      int          // imports of third-party/stdlib packages (not internal)
	Languages     []string     // languages with at least one source node, sorted
}

// Options controls extraction.
type Options struct {
	Root string
	// Files is the population: repo-relative slash paths of the eligible
	// sources plus the config files the extractors read (go.mod, tsconfig).
	// Required; the caller owns the exclusion rules.
	Files []string
	// PythonRoots are repo-relative source roots; when empty they are detected.
	PythonRoots []string
}

// edgeSet accumulates deduped runtime and type-only edges.
type edgeSet struct {
	runtime map[Edge]bool
	typed   map[Edge]bool
	unres   map[Unresolved]bool
	skips   map[Skipped]bool
	ext     int
}

func newEdgeSet() *edgeSet {
	return &edgeSet{runtime: map[Edge]bool{}, typed: map[Edge]bool{}, unres: map[Unresolved]bool{}, skips: map[Skipped]bool{}}
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

func (s *edgeSet) skip(from, reason string) {
	s.skips[Skipped{from, reason}] = true
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

// Build extracts the import graph for all supported languages from opts.Files
// read under opts.Root.
func Build(opts Options) (*Graph, error) {
	if opts.Files == nil {
		return nil, errors.New("importgraph: Options.Files is required")
	}
	files := append([]string(nil), opts.Files...)
	sort.Strings(files)
	es := newEdgeSet()
	nodes := map[string]bool{}
	langs := map[string]bool{}
	extractors := []struct {
		lang  string
		build func() ([]string, error)
	}{
		{"go", func() ([]string, error) { return buildGo(opts.Root, files, es) }},
		{"python", func() ([]string, error) { return buildPython(opts, files, es) }},
		{"typescript", func() ([]string, error) { return buildTS(opts.Root, files, es) }},
	}
	for _, x := range extractors {
		n, err := x.build()
		if err != nil {
			return nil, err
		}
		if len(n) > 0 {
			langs[x.lang] = true
			for _, node := range n {
				nodes[node] = true
			}
		}
	}
	return es.graph(nodes, langs), nil
}

func (s *edgeSet) graph(nodes, langs map[string]bool) *Graph {
	return &Graph{
		Nodes:         sortedKeys(nodes),
		Edges:         sortEdges(s.runtime, nil),
		TypeOnlyEdges: sortEdges(s.typed, s.runtime),
		Unresolved:    sortedUnresolved(s.unres),
		Skipped:       sortedSkipped(s.skips),
		External:      s.ext,
		Languages:     sortedKeys(langs),
	}
}

func sortedKeys(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedUnresolved(m map[Unresolved]bool) []Unresolved {
	var out []Unresolved
	for u := range m {
		out = append(out, u)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		return out[i].Spec < out[j].Spec
	})
	return out
}

func sortedSkipped(m map[Skipped]bool) []Skipped {
	var out []Skipped
	for k := range m {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		return out[i].Reason < out[j].Reason
	})
	return out
}

// tarjan returns the SCC id of every vertex (iterative, deterministic).
func tarjan(adj [][]int) []int {
	n := len(adj)
	st := &tarjanState{adj: adj, index: make([]int, n), low: make([]int, n), on: make([]bool, n), comp: make([]int, n)}
	for i := range st.index {
		st.index[i] = -1
	}
	for s := range n {
		if st.index[s] == -1 {
			st.strongConnect(s)
		}
	}
	return st.comp
}

type tarjanState struct {
	adj              [][]int
	index, low, comp []int
	on               []bool
	stack            []int
	next, ncomp      int
}

type tarjanFrame struct{ v, i int }

func (st *tarjanState) visit(v int) {
	st.index[v], st.low[v] = st.next, st.next
	st.next++
	st.stack = append(st.stack, v)
	st.on[v] = true
}

func (st *tarjanState) strongConnect(s int) {
	st.visit(s)
	work := []tarjanFrame{{s, 0}}
	for len(work) > 0 {
		f := &work[len(work)-1]
		v := f.v
		if f.i < len(st.adj[v]) {
			w := st.adj[v][f.i]
			f.i++
			if st.index[w] == -1 {
				st.visit(w)
				work = append(work, tarjanFrame{w, 0})
			} else if st.on[w] && st.index[w] < st.low[v] {
				st.low[v] = st.index[w]
			}
			continue
		}
		st.closeComponent(v)
		work = work[:len(work)-1]
		if len(work) > 0 {
			p := work[len(work)-1].v
			if st.low[v] < st.low[p] {
				st.low[p] = st.low[v]
			}
		}
	}
}

func (st *tarjanState) closeComponent(v int) {
	if st.low[v] != st.index[v] {
		return
	}
	for {
		w := st.stack[len(st.stack)-1]
		st.stack = st.stack[:len(st.stack)-1]
		st.on[w] = false
		st.comp[w] = st.ncomp
		if w == v {
			break
		}
	}
	st.ncomp++
}

func readFile(root, rel string) ([]byte, error) {
	return os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
}

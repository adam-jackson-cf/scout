package importgraph

import (
	"path"
	"sort"
	"strings"

	"scout/internal/domain"
)

var (
	// ToolDigest identifies the extraction implementation behind
	// scout.cycle_mass@1 (Go: go/parser imports; Python: line parser with
	// source-root resolution; TS/JS: comment-stripped regex with tsconfig paths).
	ToolDigest = domain.Digest("scout-importgraph@1:go-parser+py-lines+ts-regex")
	// ConfigDigest identifies the fixed rules: tests excluded, type-only edges
	// excluded from cycles, self edges dropped, unique edges per node pair.
	ConfigDigest = domain.Digest("importgraph-rules@1:runtime-edges,type-only-separate,no-tests")
)

// ScopeInput is one measurement scope: its anchor and the repo-relative
// eligible files it owns. PopulationDigest is stored on the signal verbatim; if
// empty it is derived from Paths. A repository-kind scope with no Paths owns
// every node.
type ScopeInput struct {
	Kind             string // domain.SignalScope*
	Anchor           string
	Paths            []string
	PopulationDigest string
}

func digestPaths(paths []string) string {
	p := append([]string(nil), paths...)
	sort.Strings(p)
	return domain.Digest(strings.Join(p, "\n"))
}

// nodeScopes assigns every graph node to one scope index (-1 for none). File
// nodes follow their file; a Go package-directory node follows the scope owning
// most of its files (ties: lowest scope index).
func nodeScopes(g *Graph, scopes []ScopeInput) map[string]int {
	fileScope := map[string]int{}
	for i, s := range scopes {
		for _, p := range s.Paths {
			if _, taken := fileScope[p]; !taken {
				fileScope[p] = i
			}
		}
	}
	dirVotes := map[string]map[int]int{}
	for p, i := range fileScope {
		d := path.Dir(p)
		if dirVotes[d] == nil {
			dirVotes[d] = map[int]int{}
		}
		dirVotes[d][i]++
	}
	out := make(map[string]int, len(g.Nodes))
	for _, n := range g.Nodes {
		if i, ok := fileScope[n]; ok {
			out[n] = i
			continue
		}
		best, bestN := -1, 0
		for i, c := range dirVotes[n] {
			if c > bestN || c == bestN && i < best {
				best, bestN = i, c
			}
		}
		out[n] = best
	}
	return out
}

// Signals returns scout.cycle_mass@1 per scope at snapshot. Edges are
// attributed to the scope of their importing node. Scopes without any
// supported-language node are `unmeasured`; scopes with nodes but no internal
// edge are `insufficient` (ratio undefined), never 0.
func Signals(g *Graph, scopes []ScopeInput, snapshot string) []domain.Signal {
	owner := nodeScopes(g, scopes)
	crossComp := map[int]bool{}
	{
		seen := map[int]map[int]bool{}
		for n, c := range components(g) {
			if seen[c] == nil {
				seen[c] = map[int]bool{}
			}
			seen[c][owner[n]] = true
			crossComp[c] = len(seen[c]) > 1
		}
	}
	comp := components(g)
	supported := strings.Join(g.Languages, ",")

	out := make([]domain.Signal, 0, len(scopes))
	for si, sc := range scopes {
		all := sc.Kind == domain.SignalScopeRepository && len(sc.Paths) == 0
		in := func(n string) bool { return all || owner[n] == si }
		pop := sc.PopulationDigest
		if pop == "" {
			if all {
				pop = digestPaths(g.Nodes)
			} else {
				pop = digestPaths(sc.Paths)
			}
		}
		sig := domain.Signal{
			MetricID: "scout.cycle_mass", Version: 1,
			ScopeKind: sc.Kind, ScopeAnchor: sc.Anchor, Snapshot: snapshot,
			PopulationDigest: pop, ToolDigest: ToolDigest, ConfigDigest: ConfigDigest,
			Claim: domain.ClaimDescriptive,
		}
		nodes := 0
		for _, n := range g.Nodes {
			if in(n) {
				nodes++
			}
		}
		if nodes == 0 {
			sig.Missing = domain.SignalUnmeasured
			sig.MissingReason = "no Go, Python or TypeScript/JS source in scope (import graph supports: go, python, typescript)"
			out = append(out, sig)
			continue
		}
		var internal, cyc, typeOnly, unres int
		for _, e := range g.Edges {
			if in(e.From) {
				internal++
				if c, ok := comp[e.From]; ok {
					if c2, ok2 := comp[e.To]; ok2 && c == c2 {
						cyc++
					}
				}
			}
		}
		for _, e := range g.TypeOnlyEdges {
			if in(e.From) {
				typeOnly++
			}
		}
		for _, u := range g.Unresolved {
			if in(u.From) {
				unres++
			}
		}
		touched := map[int]bool{}
		for n, c := range comp {
			if crossComp[c] && in(n) {
				touched[c] = true
			}
		}
		cross := len(touched)
		sig.Detail = map[string]any{
			"nodes":                 nodes,
			"cross_boundary_cycles": cross,
			"unresolved_imports":    unres,
			"type_only_edges":       typeOnly,
			"languages_supported":   supported,
		}
		if internal == 0 {
			sig.Missing = domain.SignalInsufficient
			sig.MissingReason = "no internal import edges in scope; ratio undefined"
			out = append(out, sig)
			continue
		}
		sig.Numerator = domain.Float(float64(cyc))
		sig.Denominator = domain.Float(float64(internal))
		sig.Value = domain.Float(float64(cyc) / float64(internal))
		out = append(out, sig)
	}
	return out
}

// components returns the strongly-connected-component id of every node that is
// in a cycle (component size > 1).
func components(g *Graph) map[string]int {
	idx := make(map[string]int, len(g.Nodes))
	for i, n := range g.Nodes {
		idx[n] = i
	}
	adj := make([][]int, len(g.Nodes))
	for _, e := range g.Edges {
		f, ok1 := idx[e.From]
		t, ok2 := idx[e.To]
		if ok1 && ok2 {
			adj[f] = append(adj[f], t)
		}
	}
	c := tarjan(adj)
	size := map[int]int{}
	for _, x := range c {
		size[x]++
	}
	out := map[string]int{}
	for i, n := range g.Nodes {
		if size[c[i]] > 1 {
			out[n] = c[i]
		}
	}
	return out
}

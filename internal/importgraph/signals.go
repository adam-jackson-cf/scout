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
	fileScope := fileScopes(scopes)
	dirVotes := scopeVotesByDir(fileScope)
	out := make(map[string]int, len(g.Nodes))
	for _, n := range g.Nodes {
		if i, ok := fileScope[n]; ok {
			out[n] = i
			continue
		}
		out[n] = majorityScope(dirVotes[n])
	}
	return out
}

func fileScopes(scopes []ScopeInput) map[string]int {
	fileScope := map[string]int{}
	for i, s := range scopes {
		for _, p := range s.Paths {
			if _, taken := fileScope[p]; !taken {
				fileScope[p] = i
			}
		}
	}
	return fileScope
}

func scopeVotesByDir(fileScope map[string]int) map[string]map[int]int {
	dirVotes := map[string]map[int]int{}
	for p, i := range fileScope {
		d := path.Dir(p)
		if dirVotes[d] == nil {
			dirVotes[d] = map[int]int{}
		}
		dirVotes[d][i]++
	}
	return dirVotes
}

func majorityScope(votes map[int]int) int {
	best, bestN := -1, 0
	for i, c := range votes {
		if c > bestN || c == bestN && i < best {
			best, bestN = i, c
		}
	}
	return best
}

// Signals returns scout.cycle_mass@1 per scope at snapshot. Edges are
// attributed to the scope of their importing node. Scopes without any
// supported-language node are `unmeasured`; scopes with nodes but no internal
// edge are `insufficient` (ratio undefined), never 0.
func Signals(g *Graph, scopes []ScopeInput, snapshot string) []domain.Signal {
	owner := nodeScopes(g, scopes)
	comp := components(g)
	c := signalContext{
		g: g, owner: owner, comp: comp, crossComp: crossScopeComponents(comp, owner),
		supported: strings.Join(g.Languages, ","), snapshot: snapshot,
	}
	out := make([]domain.Signal, 0, len(scopes))
	for si, sc := range scopes {
		out = append(out, c.scopeSignal(si, sc))
	}
	return out
}

func crossScopeComponents(comp map[string]int, owner map[string]int) map[int]bool {
	crossComp := map[int]bool{}
	seen := map[int]map[int]bool{}
	for n, c := range comp {
		if seen[c] == nil {
			seen[c] = map[int]bool{}
		}
		seen[c][owner[n]] = true
		crossComp[c] = len(seen[c]) > 1
	}
	return crossComp
}

type signalContext struct {
	g         *Graph
	owner     map[string]int
	comp      map[string]int
	crossComp map[int]bool
	supported string
	snapshot  string
}

func (c signalContext) scopeSignal(si int, sc ScopeInput) domain.Signal {
	all := sc.Kind == domain.SignalScopeRepository && len(sc.Paths) == 0
	in := func(n string) bool { return all || c.owner[n] == si }
	sig := domain.Signal{
		MetricID: "scout.cycle_mass", Version: 1,
		ScopeKind: sc.Kind, ScopeAnchor: sc.Anchor, Snapshot: c.snapshot,
		PopulationDigest: populationDigest(sc, all, c.g.Nodes), ToolDigest: ToolDigest, ConfigDigest: ConfigDigest,
		Claim: domain.ClaimDescriptive,
	}
	nodes := countWhere(c.g.Nodes, in)
	if nodes == 0 {
		sig.Missing = domain.SignalUnmeasured
		sig.MissingReason = "no Go, Python or TypeScript/JS source in scope (import graph supports: go, python, typescript)"
		return sig
	}
	internal, cyc := c.cycleEdges(in)
	sig.Detail = map[string]any{
		"nodes":                 nodes,
		"cross_boundary_cycles": c.crossBoundaryCycles(in),
		"unresolved_imports":    countWhere(c.g.Unresolved, func(u Unresolved) bool { return in(u.From) }),
		"skipped_inputs":        countWhere(c.g.Skipped, func(k Skipped) bool { return in(k.From) }),
		"type_only_edges":       countWhere(c.g.TypeOnlyEdges, func(e Edge) bool { return in(e.From) }),
		"languages_supported":   c.supported,
	}
	if internal == 0 {
		sig.Missing = domain.SignalInsufficient
		sig.MissingReason = "no internal import edges in scope; ratio undefined"
		return sig
	}
	sig.Numerator = domain.Float(float64(cyc))
	sig.Denominator = domain.Float(float64(internal))
	sig.Value = domain.Float(float64(cyc) / float64(internal))
	return sig
}

func populationDigest(sc ScopeInput, all bool, nodes []string) string {
	if sc.PopulationDigest != "" {
		return sc.PopulationDigest
	}
	if all {
		return digestPaths(nodes)
	}
	return digestPaths(sc.Paths)
}

func countWhere[T any](items []T, keep func(T) bool) int {
	n := 0
	for _, item := range items {
		if keep(item) {
			n++
		}
	}
	return n
}

func (c signalContext) cycleEdges(in func(string) bool) (internal, cyc int) {
	for _, e := range c.g.Edges {
		if !in(e.From) {
			continue
		}
		internal++
		if c.inOneCycle(e) {
			cyc++
		}
	}
	return internal, cyc
}

func (c signalContext) inOneCycle(e Edge) bool {
	from, ok := c.comp[e.From]
	if !ok {
		return false
	}
	to, ok := c.comp[e.To]
	return ok && from == to
}

func (c signalContext) crossBoundaryCycles(in func(string) bool) int {
	touched := map[int]bool{}
	for n, comp := range c.comp {
		if c.crossComp[comp] && in(n) {
			touched[comp] = true
		}
	}
	return len(touched)
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

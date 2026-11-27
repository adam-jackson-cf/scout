package testsource

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"

	"scout/tools/gate/check"
)

type symbol struct {
	file   *goFile
	pos    token.Pos
	name   string
	kind   string
	method bool
}

type refCounts struct {
	local    map[string]int
	qualeds  map[string]int
	selector map[string]int
	keyed    map[string]int
}

func newRefCounts() refCounts {
	return refCounts{local: map[string]int{}, qualeds: map[string]int{}, selector: map[string]int{}, keyed: map[string]int{}}
}

func testOnlySymbols(c *corpus) []check.Finding {
	declared, declPos := productionSymbols(c)
	interfaceMethodNames := ownedMethodNames(c)
	prod, tests := references(c, declPos)
	var out []check.Finding
	for _, s := range declared {
		if s.method && (implicitMethods[s.name] || interfaceMethodNames[s.name]) {
			continue
		}
		if usedOnlyByTests(c, s, prod, tests) && !hookExcepted(c, s) {
			out = append(out, check.Finding{File: c.display(s.file.Path), Line: c.line(s.pos), Rule: "TEST.no-test-hooks",
				Problem: fmt.Sprintf("%s %s is used only by tests", s.kind, s.name),
				Fix:     "delete it and test through the production entry point, or give it a production caller"})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].String() < out[j].String() })
	return out
}

func hookExcepted(c *corpus, s symbol) bool {
	if c.run == nil || c.run.Proofs == nil {
		return false
	}
	return c.run.Proofs.Excepted("TEST.no-test-hooks", c.display(s.file.Path), "test.hooks:"+s.name)
}

func ownedMethodNames(c *corpus) map[string]bool {
	out := map[string]bool{}
	for _, list := range ownedInterfaces(c) {
		for _, oi := range list {
			for _, m := range oi.methods {
				out[m.name] = true
			}
		}
	}
	return out
}

func references(c *corpus, declPos map[token.Pos]bool) (refCounts, refCounts) {
	prod, tests := newRefCounts(), newRefCounts()
	for _, f := range c.files {
		if f.AST == nil {
			continue
		}
		target := prod
		if !c.production(f) {
			target = tests
		}
		countRefs(c, f, target, declPos)
	}
	for _, f := range c.callers {
		countRefs(c, f, prod, declPos)
	}
	return prod, tests
}

func usedOnlyByTests(c *corpus, s symbol, prod, tests refCounts) bool {
	key := c.importPath(s.file.Dir) + "." + s.name
	if s.method {
		return prod.selector[s.name] == 0 && tests.selector[s.name] > 0
	}
	if s.kind == "field" {
		return prod.keyed[s.name] == 0 && tests.keyed[s.name] > 0
	}
	local := s.file.Dir + "." + s.name
	return prod.local[local]+prod.qualeds[key] == 0 && tests.local[local]+tests.qualeds[key] > 0
}

func productionSymbols(c *corpus) ([]symbol, map[token.Pos]bool) {
	var out []symbol
	pos := map[token.Pos]bool{}
	add := func(f *goFile, id *ast.Ident, kind string, method bool) {
		if id.Name == "_" || id.Name == "init" || id.Name == "main" {
			return
		}
		out = append(out, symbol{file: f, pos: id.Pos(), name: id.Name, kind: kind, method: method})
		pos[id.Pos()] = true
	}
	for _, f := range c.files {
		if !c.production(f) {
			continue
		}
		for _, d := range f.AST.Decls {
			declSymbols(f, d, add)
		}
	}
	return out, pos
}

func declSymbols(f *goFile, d ast.Decl, add func(*goFile, *ast.Ident, string, bool)) {
	switch d := d.(type) {
	case *ast.FuncDecl:
		if d.Recv != nil {
			add(f, d.Name, "method", true)
		} else {
			add(f, d.Name, "function", false)
		}
	case *ast.GenDecl:
		for _, s := range d.Specs {
			switch s := s.(type) {
			case *ast.TypeSpec:
				add(f, s.Name, "type", false)
				hookFields(f, s, add)
			case *ast.ValueSpec:
				for _, n := range s.Names {
					add(f, n, d.Tok.String(), false)
				}
			}
		}
	}
}

func hookFields(f *goFile, s *ast.TypeSpec, add func(*goFile, *ast.Ident, string, bool)) {
	st, ok := s.Type.(*ast.StructType)
	if !ok {
		return
	}
	for _, fl := range st.Fields.List {
		if fl.Tag != nil || !hookType(fl.Type) {
			continue
		}
		for _, n := range fl.Names {
			add(f, n, "field", false)
		}
	}
}

func hookType(e ast.Expr) bool {
	switch e.(type) {
	case *ast.FuncType, *ast.InterfaceType, *ast.StarExpr:
		return true
	}
	return false
}

func countRefs(c *corpus, f *goFile, rc refCounts, decl map[token.Pos]bool) {
	imports := importNames(f.AST)
	ast.Inspect(f.AST, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.SelectorExpr:
			rc.selector[n.Sel.Name]++
			if id, ok := n.X.(*ast.Ident); ok && imports[id.Name] != "" {
				rc.qualeds[imports[id.Name]+"."+n.Sel.Name]++
			}
		case *ast.KeyValueExpr:
			if id, ok := n.Key.(*ast.Ident); ok {
				rc.keyed[id.Name]++
			}
		case *ast.AssignStmt:
			countAssigned(n, rc)
		case *ast.Ident:
			if !decl[n.Pos()] {
				rc.local[f.Dir+"."+n.Name]++
			}
		}
		return true
	})
}

func countAssigned(a *ast.AssignStmt, rc refCounts) {
	for _, l := range a.Lhs {
		if sel, ok := l.(*ast.SelectorExpr); ok {
			rc.keyed[sel.Sel.Name]++
		}
	}
}

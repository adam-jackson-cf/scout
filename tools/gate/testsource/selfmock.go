package testsource

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
	"strings"

	"scout/tools/gate/check"
)

type method struct {
	name    string
	params  int
	results int
}

type ownedInterface struct {
	pkg     string
	name    string
	methods []method
}

func (o ownedInterface) id() string { return o.pkg + "." + o.name }

func ownedInterfaces(c *corpus) map[string][]ownedInterface {
	out := map[string][]ownedInterface{}
	for _, f := range c.files {
		if !c.production(f) {
			continue
		}
		for _, d := range f.AST.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, s := range gd.Specs {
				ts := s.(*ast.TypeSpec)
				it, ok := ts.Type.(*ast.InterfaceType)
				if ok && len(interfaceMethods(it)) > 0 {
					pkg := c.importPath(f.Dir)
					out[pkg] = append(out[pkg], ownedInterface{pkg: pkg, name: ts.Name.Name, methods: interfaceMethods(it)})
				}
			}
		}
	}
	return out
}

func interfaceMethods(it *ast.InterfaceType) []method {
	var out []method
	for _, m := range it.Methods.List {
		ft, ok := m.Type.(*ast.FuncType)
		if !ok {
			continue
		}
		for _, n := range m.Names {
			out = append(out, method{name: n.Name, params: fieldCount(ft.Params), results: fieldCount(ft.Results)})
		}
	}
	return out
}

func fieldCount(fl *ast.FieldList) int {
	if fl == nil {
		return 0
	}
	n := 0
	for _, f := range fl.List {
		if len(f.Names) == 0 {
			n++
		}
		n += len(f.Names)
	}
	return n
}

type testType struct {
	file    *goFile
	spec    *ast.TypeSpec
	methods map[string]method
}

func runSelfMocking(r *check.Run) ([]check.Finding, error) {
	c, err := loadCorpus(r)
	if err != nil {
		return nil, err
	}
	ifaces := ownedInterfaces(c)
	allowed := map[string]bool{}
	for _, a := range c.set.Structure.AdapterInterfaces {
		allowed[a] = true
	}
	vars := packageVars(c)
	var out []check.Finding
	for dir, types := range testTypes(c) {
		visible := visibleInterfaces(c, dir, ifaces)
		for _, tt := range types {
			out = append(out, fakeFindings(c, tt, visible, allowed)...)
		}
	}
	for _, f := range c.files {
		if f.Test && f.AST != nil {
			out = append(out, rebindFindings(c, f, vars)...)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].String() < out[j].String() })
	return out, nil
}

func testTypes(c *corpus) map[string][]*testType {
	byName := map[string]*testType{}
	out := map[string][]*testType{}
	for _, f := range c.files {
		if f.Test && f.AST != nil {
			for _, tt := range fileTypes(f) {
				byName[f.Dir+"."+tt.spec.Name.Name] = tt
				out[f.Dir] = append(out[f.Dir], tt)
			}
		}
	}
	for _, f := range c.files {
		if f.Test && f.AST != nil {
			addMethods(f, byName)
		}
	}
	return out
}

func fileTypes(f *goFile) []*testType {
	var out []*testType
	for _, d := range f.AST.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, s := range gd.Specs {
			out = append(out, &testType{file: f, spec: s.(*ast.TypeSpec), methods: map[string]method{}})
		}
	}
	return out
}

func addMethods(f *goFile, byName map[string]*testType) {
	for _, d := range f.AST.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}
		if tt := byName[f.Dir+"."+recvType(fn.Recv.List[0].Type)]; tt != nil {
			tt.methods[fn.Name.Name] = method{name: fn.Name.Name, params: fieldCount(fn.Type.Params), results: fieldCount(fn.Type.Results)}
		}
	}
}

func recvType(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.StarExpr:
		return recvType(t.X)
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr:
		return recvType(t.X)
	}
	return ""
}

func visibleInterfaces(c *corpus, dir string, ifaces map[string][]ownedInterface) map[string]ownedInterface {
	out := map[string]ownedInterface{}
	pkgs := map[string]bool{c.importPath(dir): true}
	for _, f := range c.files {
		if f.Dir != dir || f.AST == nil {
			continue
		}
		for _, p := range importNames(f.AST) {
			pkgs[p] = true
		}
	}
	for p := range pkgs {
		for _, oi := range ifaces[p] {
			out[oi.id()] = oi
		}
	}
	return out
}

func fakeFindings(c *corpus, tt *testType, visible map[string]ownedInterface, allowed map[string]bool) []check.Finding {
	var out []check.Finding
	if st, ok := tt.spec.Type.(*ast.StructType); ok {
		for _, fl := range st.Fields.List {
			if len(fl.Names) == 0 {
				if oi, ok := embeddedInterface(c, tt.file, fl.Type, visible); ok && !allowed[oi.id()] {
					out = append(out, selfMockFinding(c, tt, oi, "embeds"))
				}
			}
		}
	}
	for _, oi := range visible {
		if !allowed[oi.id()] && implements(tt.methods, oi.methods) {
			out = append(out, selfMockFinding(c, tt, oi, "implements"))
		}
	}
	return out
}

func selfMockFinding(c *corpus, tt *testType, oi ownedInterface, how string) check.Finding {
	return check.Finding{File: c.display(tt.file.Path), Line: c.line(tt.spec.Pos()), Rule: "TEST.no-self-mocking",
		Problem: fmt.Sprintf("test double %s %s owned interface %s", tt.spec.Name.Name, how, oi.id()),
		Fix:     "drive the real implementation through its public entry point, or substitute only an owned adapter interface listed in proofs.toml [structure] adapter_interfaces"}
}

func embeddedInterface(c *corpus, f *goFile, e ast.Expr, visible map[string]ownedInterface) (ownedInterface, bool) {
	switch t := e.(type) {
	case *ast.Ident:
		oi, ok := visible[c.importPath(f.Dir)+"."+t.Name]
		return oi, ok
	case *ast.SelectorExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			oi, ok := visible[importNames(f.AST)[id.Name]+"."+t.Sel.Name]
			return oi, ok
		}
	}
	return ownedInterface{}, false
}

func implements(have map[string]method, want []method) bool {
	for _, m := range want {
		got, ok := have[m.name]
		if !ok || got.params != m.params || got.results != m.results {
			return false
		}
	}
	return len(want) > 0
}

func packageVars(c *corpus) map[string]bool {
	out := map[string]bool{}
	for _, f := range c.files {
		if !c.production(f) {
			continue
		}
		for _, d := range f.AST.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				continue
			}
			for _, s := range gd.Specs {
				for _, n := range s.(*ast.ValueSpec).Names {
					out[c.importPath(f.Dir)+"."+n.Name] = true
				}
			}
		}
	}
	return out
}

func rebindFindings(c *corpus, f *goFile, vars map[string]bool) []check.Finding {
	imports := importNames(f.AST)
	pkg := c.importPath(f.Dir)
	var out []check.Finding
	ast.Inspect(f.AST, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok || as.Tok != token.ASSIGN {
			return true
		}
		for _, lhs := range as.Lhs {
			if target := ownedTarget(lhs, imports, pkg, c.module, vars); target != "" {
				out = append(out, check.Finding{File: c.display(f.Path), Line: c.line(as.Pos()), Rule: "TEST.no-self-mocking",
					Problem: fmt.Sprintf("test rebinds owned %s", target),
					Fix:     "pass the dependency through the public entry point instead of replacing owned package state"})
			}
		}
		return true
	})
	return out
}

func ownedTarget(e ast.Expr, imports map[string]string, pkg, module string, vars map[string]bool) string {
	root, chain := e, []string{}
	for {
		sel, ok := root.(*ast.SelectorExpr)
		if !ok {
			break
		}
		chain = append([]string{sel.Sel.Name}, chain...)
		root = sel.X
	}
	id, ok := root.(*ast.Ident)
	if !ok {
		return ""
	}
	if p := imports[id.Name]; p != "" && len(chain) > 0 && (p == module || strings.HasPrefix(p, module+"/")) {
		return p + "." + strings.Join(chain, ".")
	}
	if vars[pkg+"."+id.Name] && id.Obj == nil {
		return pkg + "." + strings.Join(append([]string{id.Name}, chain...), ".")
	}
	return ""
}

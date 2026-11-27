package testsource

import (
	"fmt"
	"go/ast"
	"go/token"

	"scout/tools/gate/check"
)

const timeoutFix = "pass a context from context.WithTimeout or context.WithDeadline with a positive duration (for example boundedTestContext(t))"

func runSubprocessTimeouts(r *check.Run) ([]check.Finding, error) {
	c, err := loadCorpus(r)
	if err != nil {
		return nil, err
	}
	bounded := boundedHelpers(c)
	var out []check.Finding
	for _, f := range c.files {
		if f.AST == nil || c.production(f) {
			continue
		}
		imports := importNames(f.AST)
		for _, d := range f.AST.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if ok && fn.Body != nil {
				out = append(out, timeoutFindings(c, f, fn, imports, bounded[f.Dir])...)
			}
		}
	}
	return out, nil
}

func boundedHelpers(c *corpus) map[string]map[string]bool {
	out := map[string]map[string]bool{}
	for _, f := range c.files {
		if f.AST == nil {
			continue
		}
		imports := importNames(f.AST)
		for _, d := range f.AST.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if ok && fn.Body != nil && fn.Recv == nil && returnsBounded(fn, imports) {
				if out[f.Dir] == nil {
					out[f.Dir] = map[string]bool{}
				}
				out[f.Dir][fn.Name.Name] = true
			}
		}
	}
	return out
}

func returnsBounded(fn *ast.FuncDecl, imports map[string]string) bool {
	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok && isPkgCall(call, imports, "context", "WithTimeout", "WithDeadline") {
			found = true
		}
		return !found
	})
	return found
}

func isPkgCall(call *ast.CallExpr, imports map[string]string, pkg string, names ...string) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	if !ok || imports[id.Name] != pkg {
		return false
	}
	for _, n := range names {
		if sel.Sel.Name == n {
			return true
		}
	}
	return false
}

type ctxScope struct {
	bounded map[string]bool
	params  map[string]bool
	helpers map[string]bool
	imports map[string]string
}

func timeoutFindings(c *corpus, f *goFile, fn *ast.FuncDecl, imports map[string]string, helpers map[string]bool) []check.Finding {
	sc := ctxScope{bounded: map[string]bool{}, params: map[string]bool{}, helpers: helpers, imports: imports}
	for _, fl := range fn.Type.Params.List {
		for _, id := range fl.Names {
			sc.params[id.Name] = true
		}
	}
	var out []check.Finding
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.AssignStmt:
			sc.recordAssign(n)
		case *ast.CallExpr:
			if p := sc.problem(n); p != "" {
				out = append(out, check.Finding{File: c.display(f.Path), Line: c.line(n.Pos()), Rule: "TEST.subprocess-timeouts", Problem: p, Fix: timeoutFix})
			}
		}
		return true
	})
	return out
}

func (sc ctxScope) recordAssign(a *ast.AssignStmt) {
	if len(a.Rhs) != 1 || len(a.Lhs) == 0 {
		return
	}
	call, ok := a.Rhs[0].(*ast.CallExpr)
	if !ok {
		return
	}
	id, ok := a.Lhs[0].(*ast.Ident)
	if !ok {
		return
	}
	if sc.boundedCall(call) {
		sc.bounded[id.Name] = true
	}
}

func (sc ctxScope) boundedCall(call *ast.CallExpr) bool {
	if isPkgCall(call, sc.imports, "context", "WithTimeout", "WithDeadline") {
		return positiveDuration(call)
	}
	if id, ok := call.Fun.(*ast.Ident); ok {
		return sc.helpers[id.Name]
	}
	return false
}

func (sc ctxScope) problem(call *ast.CallExpr) string {
	switch {
	case isPkgCall(call, sc.imports, "os/exec", "Command"):
		return "exec.Command starts a process with no timeout"
	case isPkgCall(call, sc.imports, "os", "StartProcess"), isPkgCall(call, sc.imports, "syscall", "ForkExec", "Exec"):
		return "a process is started with no timeout"
	case isPkgCall(call, sc.imports, "context", "WithTimeout", "WithDeadline") && !positiveDuration(call):
		return "context timeout is not a positive duration"
	case isPkgCall(call, sc.imports, "os/exec", "CommandContext") && len(call.Args) > 0 && !sc.boundedArg(call.Args[0]):
		return fmt.Sprintf("exec.CommandContext context %s has no deadline", exprText(call.Args[0]))
	}
	return ""
}

func (sc ctxScope) boundedArg(e ast.Expr) bool {
	switch e := e.(type) {
	case *ast.Ident:
		return sc.bounded[e.Name] || sc.params[e.Name]
	case *ast.CallExpr:
		return sc.boundedCall(e)
	}
	return false
}

func positiveDuration(call *ast.CallExpr) bool {
	if len(call.Args) < 2 {
		return false
	}
	if lit, ok := call.Args[1].(*ast.BasicLit); ok && lit.Kind == token.INT {
		return lit.Value != "0"
	}
	if u, ok := call.Args[1].(*ast.UnaryExpr); ok && u.Op == token.SUB {
		return false
	}
	return true
}

func exprText(e ast.Expr) string {
	switch e := e.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.CallExpr:
		return exprText(e.Fun) + "()"
	case *ast.SelectorExpr:
		return exprText(e.X) + "." + e.Sel.Name
	}
	return "expression"
}

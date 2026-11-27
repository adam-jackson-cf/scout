package testsource

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	"scout/tools/gate/check"
)

const evasionCheck = "test.evasion"

var skipCalls = map[string]bool{"Skip": true, "Skipf": true, "SkipNow": true}

func runEvasion(r *check.Run) ([]check.Finding, error) {
	c, err := loadCorpus(r)
	if err != nil {
		return nil, err
	}
	var out []check.Finding
	for _, f := range c.files {
		if f.Test && f.AST != nil && !r.Proofs.Excepted("TEST.no-evasion", f.Path, evasionCheck) {
			out = append(out, skipFindings(c, f)...)
		}
	}
	increases, err := raisedTimeouts(c)
	return append(out, increases...), err
}

func raisedTimeouts(c *corpus) ([]check.Finding, error) {
	chs, err := c.changes()
	if err != nil {
		return nil, err
	}
	var out []check.Finding
	for _, ch := range chs {
		if ch.Base == nil {
			continue
		}
		if f := c.byPath[ch.Path]; f != nil && f.Test && f.AST != nil {
			out = append(out, timeoutIncreases(c, f, ch.Base)...)
		}
		if ch.Path == "proofs.toml" {
			out = append(out, packageTimeoutIncrease(c, ch.Base)...)
		}
	}
	return out, nil
}

func skipFindings(c *corpus, f *goFile) []check.Finding {
	imports := importNames(f.AST)
	tvars := testingVars(f.AST, imports)
	var out []check.Finding
	ast.Inspect(f.AST, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		x, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		switch {
		case imports[x.Name] == "testing" && sel.Sel.Name == "Short":
			out = append(out, check.Finding{File: c.display(f.Path), Line: c.line(call.Pos()), Rule: "TEST.no-evasion", Problem: "testing.Short() changes what the test checks depending on a run flag", Fix: "remove the branch so every run checks the same behaviour, or record an owner-approved [[exception]] in proofs.toml"})
		case tvars[x.Name] && skipCalls[sel.Sel.Name]:
			out = append(out, check.Finding{File: c.display(f.Path), Line: c.line(call.Pos()), Rule: "TEST.no-evasion", Problem: fmt.Sprintf("%s.%s skips the test without an approved exception", x.Name, sel.Sel.Name), Fix: "make the missing prerequisite a failure (t.Fatal) or record an owner-approved [[exception]] for this path in proofs.toml"})
		}
		return true
	})
	return out
}

func timeoutIncreases(c *corpus, f *goFile, base []byte) []check.Finding {
	baseAST, err := parser.ParseFile(token.NewFileSet(), f.Path, base, 0)
	if err != nil {
		return nil
	}
	before := maxTimeouts(baseAST)
	var out []check.Finding
	for name, now := range maxTimeouts(f.AST) {
		if was, ok := before[name]; ok && now.d > was.d {
			out = append(out, check.Finding{File: c.display(f.Path), Line: c.line(now.pos), Rule: "TEST.no-evasion", Problem: fmt.Sprintf("%s raises its timeout from %s to %s", name, was.d, now.d), Fix: "keep the earlier timeout and make the test faster, or record an owner-approved [[exception]] in proofs.toml"})
		}
	}
	return out
}

type timeoutAt struct {
	d   time.Duration
	pos token.Pos
}

func maxTimeouts(af *ast.File) map[string]timeoutAt {
	imports := importNames(af)
	out := map[string]timeoutAt{}
	for _, d := range af.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if dur, ok := callTimeout(call, imports); ok && dur > out[fn.Name.Name].d {
				out[fn.Name.Name] = timeoutAt{d: dur, pos: call.Pos()}
			}
			return true
		})
	}
	return out
}

func callTimeout(call *ast.CallExpr, imports map[string]string) (time.Duration, bool) {
	switch {
	case isPkgCall(call, imports, "context", "WithTimeout") && len(call.Args) == 2:
		return durationOf(call.Args[1], imports)
	case isPkgCall(call, imports, "time", "After", "NewTimer", "AfterFunc") && len(call.Args) >= 1:
		return durationOf(call.Args[0], imports)
	}
	return 0, false
}

var units = map[string]time.Duration{"Nanosecond": time.Nanosecond, "Microsecond": time.Microsecond, "Millisecond": time.Millisecond, "Second": time.Second, "Minute": time.Minute, "Hour": time.Hour}

func durationOf(e ast.Expr, imports map[string]string) (time.Duration, bool) {
	switch e := e.(type) {
	case *ast.ParenExpr:
		return durationOf(e.X, imports)
	case *ast.BasicLit:
		v, err := strconv.ParseFloat(e.Value, 64)
		return time.Duration(v), err == nil
	case *ast.SelectorExpr:
		id, ok := e.X.(*ast.Ident)
		u, known := units[e.Sel.Name]
		return u, ok && known && imports[id.Name] == "time"
	case *ast.BinaryExpr:
		if e.Op != token.MUL {
			return 0, false
		}
		a, okA := durationOf(e.X, imports)
		b, okB := durationOf(e.Y, imports)
		return a * b, okA && okB
	}
	return 0, false
}

func packageTimeoutIncrease(c *corpus, base []byte) []check.Finding {
	var old struct {
		Layout layout `toml:"layout"`
	}
	if _, err := toml.Decode(string(base), &old); err != nil {
		return nil
	}
	was, errWas := time.ParseDuration(old.Layout.Tiers.PackageTimeout)
	now, errNow := time.ParseDuration(c.set.Tiers.PackageTimeout)
	if errWas != nil || errNow != nil || now <= was {
		return nil
	}
	if c.run.Proofs.Excepted("TEST.no-evasion", "proofs.toml", evasionCheck) {
		return nil
	}
	return []check.Finding{{File: "proofs.toml", Rule: "TEST.no-evasion", Problem: fmt.Sprintf("[layout.test_tiers] package_timeout rises from %s to %s", was, now), Fix: "keep the earlier hang timeout and speed up the slow package, or record an owner-approved [[exception]] for proofs.toml"}}
}

func testingVars(af *ast.File, imports map[string]string) map[string]bool {
	out := map[string]bool{}
	ast.Inspect(af, func(n ast.Node) bool {
		fl, ok := n.(*ast.Field)
		if !ok || !isTestingType(fl.Type, imports) {
			return true
		}
		for _, id := range fl.Names {
			out[id.Name] = true
		}
		return true
	})
	return out
}

func isTestingType(e ast.Expr, imports map[string]string) bool {
	if st, ok := e.(*ast.StarExpr); ok {
		e = st.X
	}
	sel, ok := e.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && imports[id.Name] == "testing" && strings.Contains("T B F TB", sel.Sel.Name)
}

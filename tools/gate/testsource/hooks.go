package testsource

import (
	"fmt"
	"go/ast"
	"go/token"
	"regexp"
	"sort"
	"strconv"

	"scout/tools/gate/check"
)

var (
	testEnvName     = regexp.MustCompile(`(?i)(^|[_.-])tests?($|[_.-])|(^|[_.-])testing($|[_.-])`)
	seamComment     = regexp.MustCompile(`(?i)\b(test seams?|test hooks?|(replaceable|used|set|exported) (only )?(in|by|for) tests|for tests only|tests only|for testing only|only (in|for) tests)\b`)
	implicitMethods = map[string]bool{"String": true, "Error": true, "GoString": true, "Format": true, "MarshalJSON": true, "UnmarshalJSON": true, "MarshalText": true, "UnmarshalText": true, "MarshalYAML": true, "UnmarshalYAML": true, "MarshalTOML": true, "UnmarshalTOML": true, "Len": true, "Less": true, "Swap": true, "Push": true, "Pop": true, "ServeHTTP": true, "Read": true, "Write": true, "Close": true, "Unwrap": true, "Is": true, "As": true, "Scan": true, "Value": true, "Seek": true, "WriteTo": true, "ReadFrom": true}
)

const hookFix = "remove the test-only path and give tests the production input (configuration, flag or adapter) real callers use"

func runTestHooks(r *check.Run) ([]check.Finding, error) {
	c, err := loadCorpus(r)
	if err != nil {
		return nil, err
	}
	var out []check.Finding
	for _, f := range c.files {
		if c.production(f) {
			out = append(out, markerFindings(c, f)...)
		}
	}
	out = append(out, testOnlySymbols(c)...)
	sort.Slice(out, func(i, j int) bool { return out[i].String() < out[j].String() })
	return out, nil
}

func markerFindings(c *corpus, f *goFile) []check.Finding {
	var out []check.Finding
	add := func(pos token.Pos, problem string) {
		out = append(out, check.Finding{File: c.display(f.Path), Line: c.line(pos), Rule: "TEST.no-test-hooks", Problem: problem, Fix: hookFix})
	}
	for _, im := range f.AST.Imports {
		if im.Path.Value == "\"testing\"" {
			add(im.Pos(), "production code imports testing")
		}
	}
	for _, cg := range f.AST.Comments {
		if m := seamComment.FindString(cg.Text()); m != "" {
			add(cg.Pos(), fmt.Sprintf("production code declares a test-only path (%q)", m))
		}
	}
	imports := importNames(f.AST)
	ast.Inspect(f.AST, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if p := markerCall(call, imports); p != "" {
				add(call.Pos(), p)
			}
		}
		return true
	})
	return out
}

func markerCall(call *ast.CallExpr, imports map[string]string) string {
	switch {
	case isPkgCall(call, imports, "testing", "Testing"):
		return "production code branches on testing.Testing()"
	case isPkgCall(call, imports, "os", "Getenv", "LookupEnv") && len(call.Args) == 1:
		if name := stringLit(call.Args[0]); name != "" && testEnvName.MatchString(name) {
			return fmt.Sprintf("production code reads test-marker environment variable %s", name)
		}
	case isPkgCall(call, imports, "flag", "Lookup") && len(call.Args) == 1:
		if name := stringLit(call.Args[0]); len(name) > 5 && name[:5] == "test." {
			return fmt.Sprintf("production code reads go test flag %s", name)
		}
	}
	return ""
}

func stringLit(e ast.Expr) string {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return ""
	}
	s, err := strconv.Unquote(lit.Value)
	if err != nil {
		return ""
	}
	return s
}

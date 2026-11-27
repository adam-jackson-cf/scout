package testsource

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/printer"
	"go/scanner"
	"go/token"
	"sort"
	"strconv"
	"strings"

	"scout/tools/gate/check"
)

type testCase struct {
	file *goFile
	fn   *ast.FuncDecl
	key  string
}

var failureCalls = map[string]bool{"Error": true, "Errorf": true, "Fatal": true, "Fatalf": true, "Log": true, "Logf": true, "Skip": true, "Skipf": true}

func Fingerprint(fset *token.FileSet, importPath string, af *ast.File, fn *ast.FuncDecl) string {
	if fn.Body == nil || len(fn.Body.List) == 0 {
		return ""
	}
	locals := localNames(fn)
	imports := importNames(af)
	var buf bytes.Buffer
	cfg := printer.Config{Mode: printer.RawFormat}
	if err := cfg.Fprint(&buf, fset, dropMessages(fn.Body)); err != nil {
		return ""
	}
	var s scanner.Scanner
	file := token.NewFileSet().AddFile("", -1, buf.Len())
	s.Init(file, buf.Bytes(), nil, 0)
	var toks []string
	nz := normaliser{locals: locals, imports: imports, pkg: importPath, renamed: map[string]string{}}
	prevPeriod := false
	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		toks = append(toks, nz.token(tok, lit, prevPeriod))
		prevPeriod = tok == token.PERIOD
	}
	sum := sha256.Sum256([]byte(strings.Join(toks, " ")))
	return fmt.Sprintf("%x", sum[:12])
}

type normaliser struct {
	locals  map[string]bool
	imports map[string]string
	pkg     string
	renamed map[string]string
}

func (nz normaliser) token(tok token.Token, lit string, afterPeriod bool) string {
	locals, imports, pkg, renamed := nz.locals, nz.imports, nz.pkg, nz.renamed
	if tok == token.COMMENT || tok == token.SEMICOLON && lit == "\n" {
		return ""
	}
	if tok != token.IDENT {
		if lit != "" {
			return lit
		}
		return tok.String()
	}
	switch {
	case afterPeriod:
		return lit
	case locals[lit]:
		if _, ok := renamed[lit]; !ok {
			renamed[lit] = "v" + strconv.Itoa(len(renamed))
		}
		return renamed[lit]
	case imports[lit] != "":
		return imports[lit]
	}
	return pkg + "." + lit
}

func localNames(fn *ast.FuncDecl) map[string]bool {
	out := map[string]bool{}
	ast.Inspect(fn, func(n ast.Node) bool {
		for _, id := range declaredIdents(n) {
			if ident, ok := id.(*ast.Ident); ok {
				out[ident.Name] = true
			}
		}
		return true
	})
	return out
}

func declaredIdents(n ast.Node) []ast.Expr {
	var out []ast.Expr
	switch n := n.(type) {
	case *ast.Field:
		for _, id := range n.Names {
			out = append(out, id)
		}
	case *ast.ValueSpec:
		for _, id := range n.Names {
			out = append(out, id)
		}
	case *ast.AssignStmt:
		if n.Tok == token.DEFINE {
			out = n.Lhs
		}
	case *ast.RangeStmt:
		out = []ast.Expr{n.Key, n.Value}
	}
	return out
}

func importNames(af *ast.File) map[string]string {
	out := map[string]string{}
	for _, im := range af.Imports {
		p, err := strconv.Unquote(im.Path.Value)
		if err != nil {
			continue
		}
		name := p[strings.LastIndex(p, "/")+1:]
		if im.Name != nil {
			name = im.Name.Name
		}
		out[name] = p
	}
	return out
}

func dropMessages(body *ast.BlockStmt) ast.Node {
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok && failureCalls[sel.Sel.Name] {
			call.Args = nil
		}
		return true
	})
	return body
}

func runDuplicates(r *check.Run) ([]check.Finding, error) {
	c, err := loadCorpus(r)
	if err != nil {
		return nil, err
	}
	chs, err := c.changes()
	if err != nil {
		return nil, err
	}
	fresh := map[string]map[string]bool{}
	for _, ch := range chs {
		if f := c.byPath[ch.Path]; f != nil && f.Test && f.AST != nil {
			fresh[f.Path] = newTestNames(c, f, ch.Base)
		}
	}
	return duplicateFindings(c, fingerprints(c), fresh), nil
}

func fingerprints(c *corpus) map[string][]testCase {
	byKey := map[string][]testCase{}
	for _, f := range c.files {
		if !f.Test || f.AST == nil {
			continue
		}
		for _, fn := range topTests(f.AST) {
			if key := Fingerprint(c.fset, c.importPath(f.Dir), f.AST, fn); key != "" {
				byKey[key] = append(byKey[key], testCase{file: f, fn: fn, key: key})
			}
		}
	}
	return byKey
}

func newTestNames(c *corpus, f *goFile, base []byte) map[string]bool {
	out := map[string]bool{}
	old := map[string]bool{}
	if base != nil {
		for _, n := range growthOld(c, f, base) {
			old[n] = true
		}
	}
	for _, fn := range topTests(f.AST) {
		if !old[fn.Name.Name] {
			out[fn.Name.Name] = true
		}
	}
	return out
}

func duplicateFindings(c *corpus, byKey map[string][]testCase, fresh map[string]map[string]bool) []check.Finding {
	var out []check.Finding
	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		cases := byKey[k]
		if len(cases) < 2 {
			continue
		}
		for i, tc := range cases {
			if !fresh[tc.file.Path][tc.fn.Name.Name] {
				continue
			}
			other := cases[0]
			if i == 0 {
				other = cases[1]
			}
			out = append(out, check.Finding{File: c.display(tc.file.Path), Line: c.line(tc.fn.Pos()), Rule: "TEST.duplicate-path",
				Problem: fmt.Sprintf("new test %s repeats the call path and assertions of %s (%s:%d)", tc.fn.Name.Name, other.fn.Name.Name, c.display(other.file.Path), c.line(other.fn.Pos())),
				Fix:     "merge the two into one table-driven test or remove the new one"})
		}
	}
	return out
}

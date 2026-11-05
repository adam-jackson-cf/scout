// Command inspect emits parser-backed guardrail inputs; it never changes source.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
)

type Case struct {
	Name        string
	Fingerprint string
}
type Test struct {
	Name       string
	Body       string
	Calls      []string
	Assertions []string
	Rows       int
	Cases      []Case
}
type File struct {
	Path          string
	Package       string
	Imports       []string
	Symbols       []string
	Tests         []Test
	Lines         int
	Unbounded     []string
	BoundedHelper bool
	UsesHelper    bool
}

func text(fs *token.FileSet, n ast.Node) string {
	var b bytes.Buffer
	format.Node(&b, fs, n)
	return b.String()
}
func digest(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func callName(c *ast.CallExpr) string {
	if s, ok := c.Fun.(*ast.SelectorExpr); ok {
		if x, ok := s.X.(*ast.Ident); ok {
			return x.Name + "." + s.Sel.Name
		}
	}
	return ""
}
func main() {
	out := []File{}
	for _, path := range os.Args[1:] {
		fs := token.NewFileSet()
		f, err := parser.ParseFile(fs, path, nil, parser.AllErrors)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: NAME.conventions: repair Go syntax: %v\n", path, err)
			os.Exit(1)
		}
		row := File{Path: path, Package: f.Name.Name, Lines: fs.Position(f.End()).Line}
		execAlias, contextAlias := "exec", "context"
		for _, im := range f.Imports {
			s, _ := strconv.Unquote(im.Path.Value)
			row.Imports = append(row.Imports, s)
			if im.Name != nil {
				if s == "os/exec" {
					execAlias = im.Name.Name
				}
				if s == "context" {
					contextAlias = im.Name.Name
				}
			}
		}
		for _, d := range f.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Body != nil {
				row.Symbols = append(row.Symbols, fn.Name.Name)
				bounded := map[string]bool{}
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					if a, ok := n.(*ast.AssignStmt); ok && len(a.Rhs) == 1 {
						if c, ok := a.Rhs[0].(*ast.CallExpr); ok && (callName(c) == contextAlias+".WithTimeout" || callName(c) == contextAlias+".WithDeadline") && len(a.Lhs) > 0 {
							if id, ok := a.Lhs[0].(*ast.Ident); ok {
								bounded[id.Name] = true
							}
						}
					}
					return true
				})
				if fn.Name.Name == "boundedTestContext" {
					cleanup, returns := false, false
					ast.Inspect(fn.Body, func(n ast.Node) bool {
						if c, ok := n.(*ast.CallExpr); ok && callName(c) == "t.Cleanup" && len(c.Args) == 1 && text(fs, c.Args[0]) == "cancel" {
							cleanup = true
						}
						if r, ok := n.(*ast.ReturnStmt); ok && len(r.Results) == 1 && text(fs, r.Results[0]) == "ctx" {
							returns = true
						}
						return true
					})
					row.BoundedHelper = bounded["ctx"] && cleanup && returns
				}
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					if c, ok := n.(*ast.CallExpr); ok {
						name := callName(c)
						if name == execAlias+".Command" {
							row.Unbounded = append(row.Unbounded, fmt.Sprintf("%s:%d", fn.Name.Name, fs.Position(c.Pos()).Line))
						}
						if name == execAlias+".CommandContext" {
							safe := false
							if len(c.Args) > 0 {
								if id, ok := c.Args[0].(*ast.Ident); ok {
									safe = bounded[id.Name]
								}
								if h, ok := c.Args[0].(*ast.CallExpr); ok {
									if id, ok := h.Fun.(*ast.Ident); ok && id.Name == "boundedTestContext" {
										safe = true
										row.UsesHelper = true
									}
								}
							}
							if !safe {
								row.Unbounded = append(row.Unbounded, fmt.Sprintf("%s:%d", fn.Name.Name, fs.Position(c.Pos()).Line))
							}
						}
					}
					return true
				})
				if !strings.HasPrefix(fn.Name.Name, "Test") {
					continue
				}
				t := Test{Name: fn.Name.Name, Body: text(fs, fn.Body)}
				// Parser object bindings distinguish local variables from package/function names.
				locals := map[*ast.Object]string{}
				next := 0
				ast.Inspect(fn, func(n ast.Node) bool {
					if id, ok := n.(*ast.Ident); ok && id.Obj != nil && id.Obj.Kind == ast.Var {
						if _, ok := locals[id.Obj]; !ok {
							locals[id.Obj] = fmt.Sprintf("v%d", next)
							next++
						}
					}
					return true
				})
				ast.Inspect(fn, func(n ast.Node) bool {
					if id, ok := n.(*ast.Ident); ok {
						if name, ok := locals[id.Obj]; ok {
							id.Name = name
						}
					}
					return true
				})
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					if c, ok := n.(*ast.CallExpr); ok {
						name := text(fs, c)
						if s, ok := c.Fun.(*ast.SelectorExpr); ok && s.Sel.Name == "Run" {
							// A subtest's label is identity, not behavior.
							name = text(fs, c.Fun)
						}
						t.Calls = append(t.Calls, name)
						if s, ok := c.Fun.(*ast.SelectorExpr); ok {
							switch s.Sel.Name {
							case "Fatal", "Fatalf", "Error", "Errorf", "Fail", "FailNow", "Equal", "NotEqual", "True", "False", "NoError", "ErrorIs", "Contains", "Len", "Empty", "NotEmpty", "Nil", "NotNil":
								t.Assertions = append(t.Assertions, text(fs, c))
							}
						}
					}
					return true
				})
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					if branch, ok := n.(*ast.IfStmt); ok {
						t.Assertions = append(t.Assertions, text(fs, branch.Cond))
					}
					return true
				})
				fp := digest(strings.Join(t.Calls, "\n") + "\n" + strings.Join(t.Assertions, "\n"))
				t.Cases = append(t.Cases, Case{Name: t.Name, Fingerprint: fp})
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					if c, ok := n.(*ast.CallExpr); ok {
						if s, ok := c.Fun.(*ast.SelectorExpr); ok && s.Sel.Name == "Run" && len(c.Args) > 1 {
							if lit, ok := c.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
								name, _ := strconv.Unquote(lit.Value)
								t.Cases = append(t.Cases, Case{Name: t.Name + "/" + name, Fingerprint: digest(text(fs, c.Args[1]))})
							}
						}
					}
					if c, ok := n.(*ast.CompositeLit); ok {
						switch c.Type.(type) {
						case *ast.ArrayType, *ast.MapType:
							for _, elt := range c.Elts {
								name := ""
								if kv, ok := elt.(*ast.KeyValueExpr); ok {
									if lit, ok := kv.Key.(*ast.BasicLit); ok && lit.Kind == token.STRING {
										name, _ = strconv.Unquote(lit.Value)
									}
									elt = kv.Value
								}
								if r, ok := elt.(*ast.CompositeLit); ok {
									values := []string{}
									for i, e := range r.Elts {
										v := e
										key := ""
										if kv, ok := e.(*ast.KeyValueExpr); ok {
											v = kv.Value
											key = text(fs, kv.Key)
										}
										if lit, ok := v.(*ast.BasicLit); ok && lit.Kind == token.STRING && (key == "name" || key == "Name" || i == 0 && key == "" && name == "") {
											name, _ = strconv.Unquote(lit.Value)
											continue
										}
										values = append(values, text(fs, v))
									}
									if name == "" {
										name = "row-" + digest(strings.Join(values, "|"))[:16]
									}
									t.Rows++
									t.Cases = append(t.Cases, Case{Name: t.Name + "/table/" + name, Fingerprint: digest(fp + strings.Join(values, "|"))})
								}
							}
						}
					}
					return true
				})
				row.Tests = append(row.Tests, t)
			}
		}
		out = append(out, row)
	}
	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		panic(err)
	}
}

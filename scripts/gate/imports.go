// Command imports prints each Go file's import paths as JSON; it never changes source.
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
)

type File struct {
	Path    string
	Package string
	Symbols []string
	Imports []string
}

func main() {
	out := []File{}
	for _, path := range os.Args[1:] {
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: go-syntax: repair Go syntax: %v\n", path, err)
			os.Exit(1)
		}
		row := File{Path: path, Package: f.Name.Name}
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				row.Symbols = append(row.Symbols, d.Name.Name)
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						row.Symbols = append(row.Symbols, s.Name.Name)
					case *ast.ValueSpec:
						for _, name := range s.Names {
							row.Symbols = append(row.Symbols, name.Name)
						}
					}
				}
			}
		}
		for _, im := range f.Imports {
			s, _ := strconv.Unquote(im.Path.Value)
			row.Imports = append(row.Imports, s)
		}
		out = append(out, row)
	}
	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

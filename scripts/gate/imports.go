// Command imports prints each Go file's import paths as JSON; it never changes source.
package main

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"strconv"
)

type File struct {
	Path    string
	Imports []string
}

func main() {
	out := []File{}
	for _, path := range os.Args[1:] {
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: go-syntax: repair Go syntax: %v\n", path, err)
			os.Exit(1)
		}
		row := File{Path: path}
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

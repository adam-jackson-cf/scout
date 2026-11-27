package codeintegrity

import (
	"fmt"
	"go/scanner"
	"go/token"
	"strings"
)

func goComments(src []byte) ([]comment, error) {
	fset := token.NewFileSet()
	file := fset.AddFile("source.go", -1, len(src))
	var errs scanner.ErrorList
	var s scanner.Scanner
	s.Init(file, src, func(pos token.Position, msg string) { errs.Add(pos, msg) }, scanner.ScanComments)
	var out []comment
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok != token.COMMENT {
			continue
		}
		line := fset.Position(pos).Line
		body := strings.TrimPrefix(lit, "//")
		if strings.HasPrefix(lit, "/*") {
			body = strings.TrimSuffix(strings.TrimPrefix(lit, "/*"), "*/")
		}
		out = append(out, comment{
			Line: line, EndLine: line + strings.Count(lit, "\n"),
			Raw: lit, Body: strings.TrimSpace(body),
		})
	}
	if errs.Len() > 0 {
		errs.Sort()
		return out, fmt.Errorf("line %d: %s", errs[0].Pos.Line, errs[0].Msg)
	}
	return out, nil
}

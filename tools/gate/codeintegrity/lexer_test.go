package codeintegrity

import (
	"reflect"
	"testing"
	"time"
)

func bodies(src string, sx syntax) []string {
	var out []string
	for _, c := range lexComments([]byte(src), sx) {
		out = append(out, c.Body)
	}
	return out
}

func TestLexerFindsCommentsOutsideQuotes(t *testing.T) {
	cases := []struct {
		name string
		src  string
		sx   syntax
		want []string
	}{
		{"shell hash after space", "echo hi # note\n", hashSyntax, []string{"note"}},
		{"shell hash inside quotes", "echo \"# no\" '# no' x#y ${#v}\n", hashSyntax, nil},
		{"shell multiline string", "echo \"a\n# still string\"\n# real\n", hashSyntax, []string{"real"}},
		{"slash line and block", "a := 1 // one\n/* two\nlines */\n", slashSyntax, []string{"one", "two\nlines"}},
		{"slash url in string", "u = \"http://x\" // tail\n", slashSyntax, []string{"tail"}},
		{"template literal spans lines", "s = `a\n// not` // yes\n", slashSyntax, []string{"yes"}},
		{"css block only", "a { b: url(http://x); } /* c */\n", cssSyntax, []string{"c"}},
		{"sql dash and quote", "select '--x' -- real\n", sqlSyntax, []string{"real"}},
		{"lua block before line", "--[[ block ]]\n-- line\n", luaSyntax, []string{"block", "line"}},
		{"haskell block", "{- note -}\nx = 1 -- tail\n", haskSyntax, []string{"note", "tail"}},
		{"html comment", "<p>x</p><!-- note -->\n", markupSyntax, []string{"note"}},
		{"tex percent escaped", "50\\% off % real\n", texSyntax, []string{"real"}},
		{"lisp semicolon", "(a \";no\") ; yes\n", lispSyntax, []string{"yes"}},
		{"ini start only", "a = b ; not\n; yes\n", iniSyntax, []string{"yes"}},
		{"properties bang", "! yes\nkey=value # not\n", propsSyntax, []string{"yes"}},
		{"yaml apostrophe in prose", "d: don't # yes\n", yamlSyntax, []string{"yes"}},
		{"yaml quoted hash", "n: \"a # b\" # yes\n", yamlSyntax, []string{"yes"}},
		{"yaml anchor url", "u: https://x/#a\n", yamlSyntax, nil},
		{"toml hash anywhere", "a=\"x#y\"#z\n", tomlSyntax, []string{"z"}},
		{"markdown fence and inline code", "`<!-- a -->`\n```\n<!-- b -->\n```\n<!-- c -->\n", mdSyntax, []string{"c"}},
		{"markdown blank lines inside fence", "```\n\n\n<!-- b -->\n\n```\n\n<!-- c -->", mdSyntax, []string{"c"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := bodies(tc.src, tc.sx); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLexerTerminatesOnEveryEdgeInput(t *testing.T) {
	inputs := []string{"", "\n", "\n\n\n", "\\", "\\\n", "\"", "'", "`", "/*", "<!--", "--[[", "{-", "#|", "```", "```\n\n", "~~~\n\n\n", "`a", "a\\\n#b", "'\n'"}
	for ext, sx := range byExtension {
		for _, in := range inputs {
			done := make(chan struct{})
			go func() {
				lexComments([]byte(in), sx)
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatalf("%s lexer did not terminate on %q", ext, in)
			}
		}
	}
}

func TestLexerReportsBlockCommentLineSpan(t *testing.T) {
	got := lexComments([]byte("x\n/* a\nb\nc */ y\n// z\n"), slashSyntax)
	if len(got) != 2 || got[0].Line != 2 || got[0].EndLine != 4 || got[1].Line != 5 {
		t.Fatalf("spans = %+v", got)
	}
}

func TestGoScannerSeparatesDirectivesFromComments(t *testing.T) {
	src := "//go:build linux\n\npackage p\n\n// doc\nvar s = \"// not\" /* block\nend */\n"
	got, err := goComments([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0].Raw != "//go:build linux" || got[1].Body != "doc" || got[2].Line != 6 || got[2].EndLine != 7 {
		t.Fatalf("comments = %+v", got)
	}
	if !isDirective(got[0], goSource) || isDirective(got[1], goSource) {
		t.Fatal("directive classification wrong")
	}
}

func TestGoScannerReportsSyntaxErrors(t *testing.T) {
	if _, err := goComments([]byte("package p\nvar s = \"open\n")); err == nil {
		t.Fatal("unterminated string scanned cleanly")
	}
}

func TestDirectivesKeepToolCommentsOnly(t *testing.T) {
	keep := []comment{
		{Line: 1, Raw: "#!/bin/sh"},
		{Line: 2, Raw: "# -*- coding: utf-8 -*-", Body: "-*- coding: utf-8 -*-"},
		{Line: 1, Raw: "# coding=latin-1", Body: "coding=latin-1"},
		{Line: 9, Body: "noqa: E501"},
		{Line: 9, Body: "NOQA"},
		{Line: 9, Body: "type: ignore[assignment]"},
		{Line: 9, Body: "pyright: ignore[reportGeneralTypeIssues]"},
		{Line: 9, Body: "fmt: off"},
		{Line: 9, Body: "ruff: noqa: F401"},
		{Line: 9, Body: "isort: skip_file"},
		{Line: 9, Body: "pylint: disable=too-many-locals"},
		{Line: 9, Body: "eslint-disable-next-line no-console"},
		{Line: 9, Body: "@ts-expect-error legacy typing"},
		{Line: 9, Body: "@ts-ignore"},
		{Line: 9, Body: "@ts-nocheck"},
		{Line: 9, Body: "prettier-ignore"},
		{Line: 9, Body: "biome-ignore lint/style: reason"},
		{Line: 9, Body: "nolint:errcheck // reason"},
		{Line: 9, Body: "SPDX-License-Identifier: MIT"},
		{Line: 9, Body: "markdownlint-disable MD013"},
		{Line: 9, Body: "vim: set ts=4 sw=4:"},
		{Line: 9, Body: "vi: noai:ts=4:sw=4"},
		{Line: 9, Body: "-*- mode: python -*-"},
	}
	for _, c := range keep {
		if !isDirective(c, lexical) {
			t.Errorf("%+v not treated as a directive", c)
		}
	}
	block := []comment{
		{Line: 3, Raw: "#!/bin/sh"},
		{Line: 5, Raw: "# coding: utf-8", Body: "coding: utf-8"},
		{Line: 9, Body: "explain why noqa is here"},
		{Line: 9, Body: "ex: see the note below"},
		{Line: 9, Body: "TODO fix later"},
		{Line: 9, Raw: "// go:build linux", Body: "go:build linux"},
	}
	for _, c := range block {
		if isDirective(c, goSource) {
			t.Errorf("%+v treated as a directive", c)
		}
	}
}

func TestClassifyChoosesSyntaxByNameExtensionAndShebang(t *testing.T) {
	cases := []struct {
		name string
		data string
		want language
	}{
		{"main.go", "", goSource},
		{"a/b.py", "", pythonSource},
		{"stubs.pyi", "", pythonSource},
		{"Dockerfile", "", lexical},
		{".github/workflows/gate.yml", "", lexical},
		{"go.mod", "", lexical},
		{".githooks/pre-commit", "#!/bin/sh\n", lexical},
		{"bin/tool", "#!/usr/bin/env python3\n", pythonSource},
		{"bin/data", "plain\n", unsupported},
		{"data.json", "", unsupported},
	}
	for _, tc := range cases {
		if got, _ := classify(tc.name, []byte(tc.data)); got != tc.want {
			t.Errorf("classify(%s) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

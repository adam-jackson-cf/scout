package importgraph

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for p, c := range files {
		full := filepath.Join(root, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func build(t *testing.T, files map[string]string) *Graph {
	t.Helper()
	names := make([]string, 0, len(files))
	for p := range files {
		names = append(names, p)
	}
	g, err := Build(Options{Root: writeTree(t, files), Files: names})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func edgeStrings(es []Edge) []string {
	var out []string
	for _, e := range es {
		out = append(out, e.From+"->"+e.To)
	}
	return out
}

func TestGoThreePackageCycle(t *testing.T) {
	g := build(t, map[string]string{
		"go.mod":      "module example.com/m\n\ngo 1.22\n",
		"a/a.go":      "package a\nimport _ \"example.com/m/b\"\nimport _ \"fmt\"\n",
		"b/b.go":      "package b\nimport _ \"example.com/m/c\"\n",
		"c/c.go":      "package c\nimport _ \"example.com/m/a\"\n",
		"d/d.go":      "package d\nimport _ \"example.com/m/a\"\nimport _ \"example.com/m/missing\"\nimport _ \"github.com/x/y\"\n",
		"a/a_test.go": "package a\nimport _ \"example.com/m/d\"\n",
	})
	res := CycleMass(g, nil)
	if res.CycleEdges != 3 || res.InternalEdges != 4 {
		t.Fatalf("cycle=%d internal=%d, want 3 of 4 (edges %v)", res.CycleEdges, res.InternalEdges, edgeStrings(g.Edges))
	}
	if res.Unresolved != 1 || g.External != 2 {
		t.Fatalf("unresolved=%d external=%d, want 1 and 2 (fmt, github.com/x/y)", res.Unresolved, g.External)
	}
	if res.CrossBoundary != 0 {
		t.Fatalf("single scope cross=%d", res.CrossBoundary)
	}
}

func TestCrossBoundaryCycleCounted(t *testing.T) {
	g := build(t, map[string]string{
		"go.mod": "module m\n",
		"x/x.go": "package x\nimport _ \"m/y\"\n",
		"y/y.go": "package y\nimport _ \"m/x\"\n",
	})
	same := CycleMass(g, func(string) string { return "s" })
	if same.CycleEdges != 2 || same.CrossBoundary != 0 {
		t.Fatalf("same-scope: %+v", same)
	}
	cross := CycleMass(g, func(n string) string { return n })
	if cross.CrossBoundary != 1 {
		t.Fatalf("cross=%d want 1", cross.CrossBoundary)
	}
}

func TestPythonCycleRelativeAndExternal(t *testing.T) {
	g := build(t, map[string]string{
		"pkg/__init__.py": "",
		"pkg/a.py":        "from . import b\nimport requests\nimport os, sys\n",
		"pkg/b.py":        "from .c import thing\n",
		"pkg/c.py":        "from pkg.a import x  # absolute closes the cycle\n",
		"pkg/d.py":        "from .a import q\nfrom .nope import z\nfrom pkg import a\n",
		"pkg/e.py":        "from typing import TYPE_CHECKING\nif TYPE_CHECKING:\n    from .a import A\n",
		"pkg/doc.py":      "\"\"\"\nimport pkg.a\n\"\"\"\n",
	})
	res := CycleMass(g, nil)
	got := edgeStrings(g.Edges)
	want := []string{"pkg/a.py->pkg/b.py", "pkg/b.py->pkg/c.py", "pkg/c.py->pkg/a.py", "pkg/d.py->pkg/a.py"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("edges %v want %v", got, want)
	}
	if res.CycleEdges != 3 || res.InternalEdges != 4 {
		t.Fatalf("cycle=%d of %d", res.CycleEdges, res.InternalEdges)
	}
	if res.Unresolved != 1 || g.External != 4 {
		t.Fatalf("unresolved=%d external=%d, want 1 and 4 (requests, os, sys, typing)", res.Unresolved, g.External)
	}
	if res.TypeOnly != 1 {
		t.Fatalf("type-only=%d want 1", res.TypeOnly)
	}
}

func TestPythonSrcRoot(t *testing.T) {
	g := build(t, map[string]string{
		"src/lib/__init__.py": "",
		"src/lib/a.py":        "import lib.b\n",
		"src/lib/b.py":        "from lib import a\n",
	})
	if r := CycleMass(g, nil); r.CycleEdges != 2 || r.InternalEdges != 2 {
		t.Fatalf("%+v %v", r, edgeStrings(g.Edges))
	}
}

func TestTypeScriptCycleAliasAndTypeOnly(t *testing.T) {
	g := build(t, map[string]string{
		"tsconfig.json": `{
  // comment
  "compilerOptions": {"baseUrl": ".", "paths": {"@app/*": ["src/*"],},},
}`,
		"src/a.ts":  "import { b } from './b';\nimport React from 'react';\n",
		"src/b.ts":  "import { c } from '@app/c';\n",
		"src/c.ts":  "export { a } from '@app/a.js';\nimport type { T } from './t';\nimport { type U } from './t';\n",
		"src/t.ts":  "import type { X } from './a';\nimport './missing';\n",
		"src/x.tsx": "const y = require('./a'); // import z from './zzz'\n",
	})
	res := CycleMass(g, nil)
	if res.CycleEdges != 3 {
		t.Fatalf("cycle edges %d, edges=%v", res.CycleEdges, edgeStrings(g.Edges))
	}
	// runtime edges: a->b, b->c, c->a, x->a, c->t excluded (type only), t->a excluded (type only)
	if res.InternalEdges != 4 {
		t.Fatalf("internal %d: %v", res.InternalEdges, edgeStrings(g.Edges))
	}
	if res.TypeOnly != 2 {
		t.Fatalf("type-only %d want 2 (c->t, t->a): %v", res.TypeOnly, edgeStrings(g.TypeOnlyEdges))
	}
	if res.Unresolved != 1 || g.External != 1 {
		t.Fatalf("unresolved=%d external=%d", res.Unresolved, g.External)
	}
}

func TestTypeOnlyCycleIsNotACycle(t *testing.T) {
	g := build(t, map[string]string{
		"a.ts": "import type { B } from './b';\n",
		"b.ts": "import { A } from './a';\n",
	})
	if r := CycleMass(g, nil); r.CycleEdges != 0 || r.InternalEdges != 1 || r.TypeOnly != 1 {
		t.Fatalf("%+v", r)
	}
}

func TestDeterministicOutput(t *testing.T) {
	files := map[string]string{
		"go.mod": "module m\n", "p/p.go": "package p\nimport _ \"m/q\"\n", "q/q.go": "package q\nimport _ \"m/p\"\n",
		"a.py": "import b\n", "b.py": "import a\n",
		"x.ts": "import './y'\n", "y.ts": "import './x'\n",
	}
	a := build(t, files)
	b := build(t, files)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("non-deterministic graph")
	}
	if got := strings.Join(a.Languages, ","); got != "go,python,typescript" {
		t.Fatal(got)
	}
	if r := CycleMass(a, nil); r.CycleEdges != 6 || r.InternalEdges != 6 {
		t.Fatalf("%+v", r)
	}
}

func TestBuildRequiresFiles(t *testing.T) {
	if _, err := Build(Options{Root: t.TempDir()}); err == nil {
		t.Fatal("expected error when Files is nil")
	}
	if g, err := Build(Options{Root: t.TempDir(), Files: []string{}}); err != nil || len(g.Nodes) != 0 {
		t.Fatalf("empty population is valid: %v %+v", err, g)
	}
}

func TestSkippedInputsAreRecordedAndSurfaced(t *testing.T) {
	g := build(t, map[string]string{
		"go.mod":            "module m\n",
		"a/a.go":            "package a\nimport _ \"m/b\"\n",
		"b/b.go":            "package b\nimport (\n",
		"web/tsconfig.json": "{not json",
		"web/src/x.ts":      "import { y } from './y';\n",
		"web/src/y.ts":      "export const y = 1;\n",
		"ok/tsconfig.json":  `{"extends": 5}`,
		"ok/z.ts":           "export const z = 1;\n",
		"p.py":              "import .rel\n",
	})
	reasons := map[string]string{}
	for _, k := range g.Skipped {
		reasons[k.From] = k.Reason
	}
	for _, from := range []string{"b", "web/src/x.ts", "web/src/y.ts", "ok/z.ts"} {
		if reasons[from] == "" {
			t.Fatalf("missing skip for %s: %+v", from, g.Skipped)
		}
	}
	if len(g.Unresolved) != 1 || g.Unresolved[0] != (Unresolved{"p.py", ".rel"}) {
		t.Fatalf("unresolved: %+v", g.Unresolved)
	}
	sigs := Signals(g, []ScopeInput{{Kind: "repository"}}, "s")
	if sigs[0].Detail["skipped_inputs"] != len(g.Skipped) || len(g.Skipped) < 4 {
		t.Fatalf("detail %+v", sigs[0].Detail)
	}
	if CycleMass(g, nil).Skipped != len(g.Skipped) {
		t.Fatal("CycleMass must carry the skipped count")
	}
}

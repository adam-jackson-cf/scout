package importgraph

import (
	"reflect"
	"testing"
)

// A workspace whose app tsconfig inherits `paths` from a base config (TS 5
// array `extends`), with an exact alias, a wildcard alias, a catch-all, and a
// baseUrl-only bare specifier. Only runtime edges may form cycles; `import
// type` edges are counted separately.
func TestTypeScriptPathsBaseUrlAndTypeOnlySeparated(t *testing.T) {
	g := build(t, map[string]string{
		"libs/ui/tsconfig.json": `{"extends": "../../tsconfig.base.json"}`,
		"tsconfig.base.json":    `{"compilerOptions": {"baseUrl": ".", "paths": {"@core": ["libs/core/index.ts"], "@ui/*": ["libs/ui/src/*"], "*": ["types/*"]}}}`,
		"apps/web/tsconfig.json": `{
  "extends": ["../../tsconfig.base.json"],
  "compilerOptions": {"strict": true}
}`,
		// runtime cycle: main -> @ui/button -> bare 'libs/core' ... -> main
		"apps/web/src/main.ts":  "import { Button } from '@ui/button';\nimport { core } from '@core';\nimport React from 'react';\n",
		"libs/ui/src/button.ts": "import { core } from 'libs/core/index';\nimport type { Theme } from '@ui/theme';\n",
		"libs/ui/src/theme.ts":  "import { Button } from '@ui/button';\nexport type Theme = {};\n",
		"libs/core/index.ts":    "import { run } from '../../apps/web/src/main';\nexport const core = 1;\n",
	})
	got := edgeStrings(g.Edges)
	want := []string{
		"apps/web/src/main.ts->libs/core/index.ts",
		"apps/web/src/main.ts->libs/ui/src/button.ts",
		"libs/core/index.ts->apps/web/src/main.ts",
		"libs/ui/src/button.ts->libs/core/index.ts",
		"libs/ui/src/theme.ts->libs/ui/src/button.ts",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("runtime edges\n got %v\nwant %v", got, want)
	}
	if got, want := edgeStrings(g.TypeOnlyEdges), []string{"libs/ui/src/button.ts->libs/ui/src/theme.ts"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("type-only edges\n got %v\nwant %v", got, want)
	}
	// 'react' misses the catch-all `*` -> types/* and is external, not unresolved.
	if len(g.Unresolved) != 0 || g.External != 1 {
		t.Fatalf("unresolved=%v external=%d", g.Unresolved, g.External)
	}
	res := CycleMass(g, nil)
	// The SCC {main, button, core} holds 4 runtime edges (main<->core, main->button->core);
	// theme -> button is outside it, and the type-only button -> theme edge is not counted.
	if res.CycleEdges != 4 || res.InternalEdges != 5 || res.TypeOnly != 1 {
		t.Fatalf("%+v", res)
	}
}

// A child config that sets only baseUrl re-anchors inherited `paths`.
func TestTypeScriptChildBaseUrlReanchorsInheritedPaths(t *testing.T) {
	g := build(t, map[string]string{
		"base.json":     `{"compilerOptions": {"paths": {"@x/*": ["*"]}}}`,
		"tsconfig.json": `{"extends": "./base", "compilerOptions": {"baseUrl": "src"}}`,
		"src/a.ts":      "import { b } from '@x/b';\n",
		"src/b.ts":      "export const b = 1;\n",
	})
	if got, want := edgeStrings(g.Edges), []string{"src/a.ts->src/b.ts"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v (unresolved %v)", got, want, g.Unresolved)
	}
}

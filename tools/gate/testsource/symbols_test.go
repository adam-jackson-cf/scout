package testsource

import (
	"go/parser"
	"go/token"
	"testing"

	"scout/tools/gate/check"
	"scout/tools/gate/proof"
)

func parsedFile(t *testing.T, fset *token.FileSet, p, src string) *goFile {
	t.Helper()
	f := &goFile{Path: p, Dir: p[:len(p)-len("/x.go")], Test: len(p) > 8 && p[len(p)-8:] == "_test.go", Src: []byte(src)}
	var err error
	if f.AST, err = parser.ParseFile(fset, p, src, 0); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestExcludedPackageCallsCountAsProductionUse(t *testing.T) {
	fset := token.NewFileSet()
	lib := parsedFile(t, fset, "lib/x.go", "package lib\n\nfunc Load() int { return 1 }\n")
	test := parsedFile(t, fset, "lib/x_test.go", "package lib\n\nimport \"testing\"\n\nfunc TestLoad(t *testing.T) { _ = Load() }\n")
	test.Dir = "lib"
	c := &corpus{fset: fset, module: "m", files: []*goFile{lib, test}}
	if got := testOnlySymbols(c); len(got) != 1 {
		t.Fatalf("without the excluded caller want 1 finding, got %v", got)
	}
	caller := parsedFile(t, fset, "tool/x.go", "package main\n\nimport \"m/lib\"\n\nfunc main() { _ = lib.Load() }\n")
	c.callers = []*goFile{caller}
	if got := testOnlySymbols(c); len(got) != 0 {
		t.Fatalf("an excluded production caller must count as use, got %v", got)
	}
}

func TestExcludedCallerSkipsTestsAndFixtures(t *testing.T) {
	ex := []string{"evaluation/**"}
	for p, want := range map[string]bool{
		"evaluation/replay.go":             true,
		"evaluation/replay_test.go":        false,
		"evaluation/x/testdata/fixture.go": false,
		"internal/eval/corpus.go":          false,
		"evaluation/notes.md":              false,
	} {
		if got := excludedCaller(ex, p); got != want {
			t.Errorf("excludedCaller(%q) = %v, want %v", p, got, want)
		}
	}
}

func TestApprovedExceptionNamesOneSymbol(t *testing.T) {
	fset := token.NewFileSet()
	lib := parsedFile(t, fset, "lib/x.go", "package lib\n\nfunc Load() int { return 1 }\n\nfunc Save() int { return 2 }\n")
	test := parsedFile(t, fset, "lib/x_test.go", "package lib\n\nimport \"testing\"\n\nfunc TestLoad(t *testing.T) { _ = Load() + Save() }\n")
	test.Dir = "lib"
	pf := &proof.File{Exceptions: []proof.Exception{{Rule: "TEST.no-test-hooks", Path: "lib/x.go", Check: "test.hooks:Load", Reason: "r", ApprovedBy: "owner"}}}
	c := &corpus{fset: fset, module: "m", files: []*goFile{lib, test}, run: &check.Run{Proofs: pf}}
	got := testOnlySymbols(c)
	if len(got) != 1 || got[0].Line != 5 {
		t.Fatalf("want only Save reported, got %v", got)
	}
}

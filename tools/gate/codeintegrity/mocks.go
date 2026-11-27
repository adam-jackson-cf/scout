package codeintegrity

import (
	"bytes"
	"fmt"
	"path"
	"regexp"
	"strings"

	"scout/tools/gate/check"
	"scout/tools/gate/codeintegrity/repofile"
)

var (
	pythonMockLibraries = []string{"unittest.mock", "mock", "pytest_mock"}
	scriptMockLibraries = []string{"sinon", "jest-mock", "ts-mockito", "testdouble", "mock-require", "proxyquire", "mock"}
	scriptExtensions    = map[string]bool{".js": true, ".jsx": true, ".mjs": true, ".cjs": true, ".ts": true, ".tsx": true, ".mts": true, ".cts": true}
	scriptImports       = []*regexp.Regexp{
		regexp.MustCompile(`\bimport\s+(?:type\s+)?[^'";]*?\bfrom\s*['"]([^'"\n]+)['"]`),
		regexp.MustCompile(`\bimport\s*['"]([^'"\n]+)['"]`),
		regexp.MustCompile(`\bexport\s+[^'";]*?\bfrom\s*['"]([^'"\n]+)['"]`),
		regexp.MustCompile(`\brequire\s*\(\s*['"]([^'"\n]+)['"]\s*\)`),
		regexp.MustCompile(`\bimport\s*\(\s*['"]([^'"\n]+)['"]\s*\)`),
	}
	scriptMockCall    = regexp.MustCompile(`\b(jest\s*\.\s*(?:mock|doMock|unstable_mockModule)|vi\s*\.\s*(?:mock|doMock))\s*\(`)
	pythonTestFile    = regexp.MustCompile(`^(?:test_.*\.py|.*_test\.py|conftest\.py)$`)
	scriptTestFile    = regexp.MustCompile(`\.(?:test|spec)\.(?:ts|tsx|mts|cts|js|jsx|mjs|cjs)$`)
	goTestFile        = regexp.MustCompile(`_test\.go$`)
	scriptTestFolders = "__tests__"
)

type layout struct {
	SupportDir       string   `toml:"support_dir"`
	FixturesDir      string   `toml:"fixtures_dir"`
	TestFilePatterns []string `toml:"test_file_patterns"`
}

type mockUse struct {
	line    int
	endLine int
	name    string
	call    bool
}

func loadLayout(r *check.Run) (layout, error) {
	var l layout
	if r.Proofs == nil {
		return l, nil
	}
	if _, err := r.Proofs.Section("layout", &l); err != nil {
		return l, fmt.Errorf("proofs.toml [layout]: %w", err)
	}
	return l, nil
}

func (l layout) testPath(name string) bool {
	return testFileName(path.Base(name)) || inTestFolder(name) || l.inDeclaredDir(name) || l.matchesTestPattern(name)
}

func testFileName(base string) bool {
	return pythonTestFile.MatchString(base) || scriptTestFile.MatchString(base) || goTestFile.MatchString(base)
}

func inTestFolder(name string) bool {
	for _, seg := range strings.Split(path.Dir(name), "/") {
		if seg == scriptTestFolders {
			return true
		}
	}
	return false
}

func (l layout) inDeclaredDir(name string) bool {
	for _, dir := range []string{l.SupportDir, l.FixturesDir} {
		dir = strings.TrimSuffix(dir, "/")
		if dir != "" && strings.HasPrefix(name, dir+"/") {
			return true
		}
	}
	return false
}

func (l layout) matchesTestPattern(name string) bool {
	for _, p := range l.TestFilePatterns {
		if repofile.Match(p, name) {
			return true
		}
	}
	return false
}

func libraryMatch(module string, libraries []string, sep string) (string, bool) {
	for _, lib := range libraries {
		if module == lib || strings.HasPrefix(module, lib+sep) {
			return lib, true
		}
	}
	return "", false
}

func scriptMocks(src []byte) []mockUse {
	lx := scan(src, slashSyntax)
	code := append([]byte(nil), src...)
	for _, c := range lx.out {
		blank(code, c.Start, c.Stop)
	}
	calls := append([]byte(nil), code...)
	for _, q := range lx.quoted {
		blank(calls, q[0]+1, q[1]-1)
	}
	lineAt := func(off int) int { return bytes.Count(code[:off], []byte("\n")) + 1 }
	var out []mockUse
	for _, re := range scriptImports {
		for _, m := range re.FindAllSubmatchIndex(code, -1) {
			lib, ok := libraryMatch(string(code[m[2]:m[3]]), scriptMockLibraries, "/")
			if ok {
				out = append(out, mockUse{line: lineAt(m[0]), endLine: lineAt(m[1]), name: lib})
			}
		}
	}
	for _, m := range scriptMockCall.FindAllSubmatchIndex(calls, -1) {
		name := strings.Join(strings.Fields(strings.ReplaceAll(string(calls[m[2]:m[3]]), ".", " . ")), "")
		out = append(out, mockUse{line: lineAt(m[0]), endLine: lineAt(m[1]), name: name, call: true})
	}
	return out
}

func blank(b []byte, from, to int) {
	for i := from; i < to && i < len(b); i++ {
		if b[i] != '\n' {
			b[i] = ' '
		}
	}
}

func pythonMocks(rep pyReport) []mockUse {
	var out []mockUse
	type key struct {
		line int
		lib  string
	}
	seen := map[key]bool{}
	for _, imp := range rep.Imports {
		lib, ok := libraryMatch(imp.Module, pythonMockLibraries, ".")
		k := key{imp.Line, lib}
		if ok && !seen[k] {
			seen[k] = true
			out = append(out, mockUse{line: imp.Line, endLine: imp.EndLine, name: lib})
		}
	}
	return out
}

func runMocks(r *check.Run) ([]check.Finding, error) {
	a, err := analyse(r)
	if err != nil {
		return nil, err
	}
	l, err := loadLayout(r)
	if err != nil {
		return nil, err
	}
	var out []check.Finding
	for _, f := range a.files {
		if l.testPath(f.Path) {
			continue
		}
		uses, failure := fileMocks(f, a.python[f.Path])
		if failure != nil {
			out = append(out, *failure)
			continue
		}
		out = append(out, mockFindings(f, uses)...)
	}
	return filter(r, "code-mocks", out), nil
}

func fileMocks(f repofile.File, rep pyReport) ([]mockUse, *check.Finding) {
	if lang, _ := classify(f.Path, f.Data); lang == pythonSource {
		if rep.Error != "" {
			failure := parseFailure(f, rep, ruleMocks)
			return nil, &failure
		}
		return pythonMocks(rep), nil
	}
	if scriptExtensions[strings.ToLower(path.Ext(f.Path))] {
		return scriptMocks(f.Data), nil
	}
	return nil, nil
}

func mockFindings(f repofile.File, uses []mockUse) []check.Finding {
	var out []check.Finding
	for _, u := range uses {
		line, ok := f.AnyAdded(u.line, u.endLine)
		if !ok {
			continue
		}
		problem := fmt.Sprintf("mock library %q imported in product code", u.name)
		if u.call {
			problem = fmt.Sprintf("mock call %s() in product code", u.name)
		}
		out = append(out, check.Finding{
			File: f.Path, Line: line, Rule: ruleMocks, Problem: problem,
			Fix: "move the mock into a declared test, support or fixture path, or give product code an owned adapter that tests substitute",
		})
	}
	return out
}

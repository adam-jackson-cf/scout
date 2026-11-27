package naming

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"

	"scout/tools/gate/check"
	"scout/tools/gate/pathscope"
)

const (
	ruleConventions = "NAME.conventions"
	ruleTestFiles   = "NAME.test-files"
	checkRepeat     = "child-repeats-parent"
)

type layout struct {
	TestFilePatterns []string    `toml:"test_file_patterns"`
	Skills           skillLayout `toml:"skills"`
}

type skillLayout struct {
	Root string `toml:"root"`
}

type judge struct {
	seen     map[string]bool
	lay      layout
	excepted func(rule, file, check string) bool
	isDir    func(string) bool
}

type issue struct {
	problem, fix string
}

func (j judge) pathFindings(e pathscope.Entry) []check.Finding {
	var out []check.Finding
	segs := strings.Split(e.Path, "/")
	for i := e.FirstNew; i < len(segs); i++ {
		out = append(out, j.segmentFindings(e.Path, segs, i)...)
	}
	return out
}

func (j judge) segmentFindings(p string, segs []string, i int) []check.Finding {
	seg := segs[i]
	isFile := i == len(segs)-1
	if fixedSegment(seg, isFile) {
		return nil
	}
	at := p
	if !isFile {
		at = strings.Join(segs[:i+1], "/")
	}
	var out []check.Finding
	for _, is := range j.segmentIssues(p, segs, i) {
		if j.seen[at+is.problem] {
			continue
		}
		j.seen[at+is.problem] = true
		out = append(out, check.Finding{File: at, Rule: ruleConventions, Problem: is.problem, Fix: is.fix})
	}
	return out
}

func fixedSegment(seg string, isFile bool) bool {
	return strings.HasPrefix(seg, ".") || pathscope.FixedFolders[seg] || (isFile && pathscope.FixedFile(seg))
}

func segmentName(seg string, isFile bool) string {
	if isFile {
		return pathscope.Stem(seg)
	}
	return seg
}

func (j judge) segmentIssues(p string, segs []string, i int) []issue {
	seg := segs[i]
	isFile := i == len(segs)-1
	name := segmentName(seg, isFile)
	var out []issue
	if isFile && pathscope.BannedWords[strings.ToLower(name)] {
		out = append(out, issue{fmt.Sprintf("file stem %q is a grab-bag word (house-style banned_words)", name), "name the file for what it holds or does"})
	}
	if pathscope.VersionToken(name) {
		out = append(out, issue{fmt.Sprintf("%q carries a v<N> version token", seg), "drop the version token; versions live only as <name>/v<N> folders or files"})
	}
	if i > 0 && j.repeats(segs, i) && !j.excepted(ruleConventions, p, checkRepeat) {
		out = append(out, issue{fmt.Sprintf("%q repeats its parent %q", seg, segs[i-1]), fmt.Sprintf("rename it without the %q tokens", segs[i-1])})
	}
	if !isFile && pathscope.Plural(seg) {
		out = append(out, issue{fmt.Sprintf("folder %q is plural but not a catalogue (house-style catalogue_plurals)", seg), fmt.Sprintf("rename it to %q", pathscope.Singular(seg))})
	}
	if fix, bad := j.casing(segs, i); bad {
		out = append(out, issue{fmt.Sprintf("%q breaks the casing convention", seg), fix})
	}
	return out
}

func (j judge) repeats(segs []string, i int) bool {
	seg, parent := segs[i], segs[i-1]
	if i == len(segs)-1 && strings.HasSuffix(seg, ".go") {
		stem := strings.TrimSuffix(strings.TrimSuffix(seg, ".go"), "_test")
		if stem == parent {
			return false
		}
	}
	child := pathscope.Tokens(seg)
	if i == len(segs)-1 {
		child = pathscope.Tokens(pathscope.Stem(seg))
	}
	return pathscope.ContainsSeq(child, pathscope.Tokens(parent))
}

func (j judge) casing(segs []string, i int) (string, bool) {
	seg, last := segs[i], len(segs)-1
	file := segs[last]
	parent := strings.Join(segs[:i], "/")
	switch {
	case j.skillFolder(parent, i, last):
		return skillCasing(seg)
	case i == last && parent == ".github/workflows":
		return workflowCasing(seg)
	case i < last && strings.HasSuffix(file, ".go"):
		return goPackageCasing(seg)
	case strings.HasSuffix(file, ".py") || strings.HasSuffix(file, ".pyi"):
		return j.pythonCasing(segs, i)
	}
	return "", false
}

func (j judge) skillFolder(parent string, i, last int) bool {
	return j.lay.Skills.Root != "" && parent == strings.TrimSuffix(j.lay.Skills.Root, "/") && i < last
}

func skillCasing(seg string) (string, bool) {
	if !pathscope.KebabCase.MatchString(seg) {
		return fmt.Sprintf("rename the skill folder to kebab-case %q", kebab(seg)), true
	}
	return "", false
}

func workflowCasing(seg string) (string, bool) {
	if !pathscope.KebabCase.MatchString(pathscope.Stem(seg)) {
		return fmt.Sprintf("rename the workflow to kebab-case %q", kebab(pathscope.Stem(seg))+path.Ext(seg)), true
	}
	return "", false
}

func goPackageCasing(seg string) (string, bool) {
	if !pathscope.Lowercase.MatchString(seg) {
		return fmt.Sprintf("rename the Go package folder to lowercase %q", strings.ToLower(strings.ReplaceAll(seg, "-", ""))), true
	}
	return "", false
}

func (j judge) pythonCasing(segs []string, i int) (string, bool) {
	isFile := i == len(segs)-1
	pkgDir := strings.Join(segs[:i+1], "/")
	if isFile {
		pkgDir = strings.Join(segs[:i], "/")
	}
	if !j.isDir(path.Join(pkgDir, "__init__.py")) {
		return "", false
	}
	name := segmentName(segs[i], isFile)
	if !pathscope.SnakeCase.MatchString(name) {
		return fmt.Sprintf("rename it to snake_case %q", snake(name)), true
	}
	return "", false
}

func kebab(s string) string { return strings.Join(pathscope.Tokens(s), "-") }
func snake(s string) string { return strings.Join(pathscope.Tokens(s), "_") }

func (j judge) testFile(e pathscope.Entry) []check.Finding {
	segs := strings.Split(e.Path, "/")
	name := segs[len(segs)-1]
	if !j.matchesTest(e.Path, name) {
		return misnamedTest(e.Path, name)
	}
	return tokenFindings(e.Path, name)
}

func misnamedTest(p, name string) []check.Finding {
	switch {
	case strings.HasSuffix(name, ".go") && strings.HasPrefix(name, "test_"):
		return []check.Finding{{File: p, Rule: ruleTestFiles,
			Problem: "looks like a test file but go test only reads *_test.go",
			Fix:     fmt.Sprintf("rename it to %s_test.go", strings.TrimSuffix(strings.TrimPrefix(name, "test_"), ".go"))}}
	case strings.HasSuffix(name, ".py") && strings.HasPrefix(name, "test"):
		return []check.Finding{{File: p, Rule: ruleTestFiles,
			Problem: "looks like a test file but matches no test pattern (test_*.py, *_test.py)",
			Fix:     fmt.Sprintf("rename it to test_%s", strings.TrimPrefix(strings.TrimPrefix(name, "test"), "_"))}}
	}
	return nil
}

func tokenFindings(p, name string) []check.Finding {
	var out []check.Finding
	for _, t := range pathscope.Tokens(pathscope.Stem(name)) {
		if pathscope.TestNameIgnored[t] || pathscope.NumberedAllowlist[t] {
			continue
		}
		if kind := tokenKind(t); kind != "" {
			out = append(out, check.Finding{File: p, Rule: ruleTestFiles,
				Problem: fmt.Sprintf("test file name carries the %s token %q", kind, t),
				Fix:     "name the file for the behaviour it checks, without ticket, version, date or number tokens"})
		}
	}
	return out
}

func tokenKind(t string) string {
	switch {
	case pathscope.DateTok.MatchString(t):
		return "date"
	case pathscope.TicketTok.MatchString(t):
		return "ticket"
	case pathscope.VersionTk2.MatchString(t):
		return "version"
	case pathscope.TrailNum.MatchString(t):
		return "numbered"
	}
	return ""
}

func (j judge) matchesTest(p, name string) bool {
	for _, pat := range []string{"*_test.go", "test_*.py", "*_test.py"} {
		if ok, _ := path.Match(pat, name); ok {
			return true
		}
	}
	for _, pat := range j.lay.TestFilePatterns {
		if pathscope.Match(pat, p) {
			return true
		}
	}
	return false
}

func goSymbols(file string, src []byte) ([]check.Finding, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, src, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}
	test := strings.HasSuffix(file, "_test.go")
	var out []check.Finding
	for _, id := range declIdents(f.Decls) {
		n := id.Name
		if !underscored(n, test) {
			continue
		}
		out = append(out, check.Finding{File: file, Line: fset.Position(id.Pos()).Line, Rule: ruleConventions,
			Problem: fmt.Sprintf("Go identifier %q uses underscores", n),
			Fix:     fmt.Sprintf("rename it to MixedCaps %q", mixed(n))})
	}
	return out, nil
}

func underscored(n string, test bool) bool {
	if n == "_" || !strings.Contains(n, "_") {
		return false
	}
	return !test || !testEntryPoint(n)
}

func testEntryPoint(n string) bool {
	for _, p := range []string{"Test", "Benchmark", "Example", "Fuzz"} {
		if strings.HasPrefix(n, p) {
			return true
		}
	}
	return false
}

func declIdents(decls []ast.Decl) []*ast.Ident {
	var ids []*ast.Ident
	for _, d := range decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			ids = append(ids, d.Name)
		case *ast.GenDecl:
			ids = append(ids, specIdents(d.Specs)...)
		}
	}
	return ids
}

func specIdents(specs []ast.Spec) []*ast.Ident {
	var ids []*ast.Ident
	for _, sp := range specs {
		switch sp := sp.(type) {
		case *ast.TypeSpec:
			ids = append(ids, sp.Name)
		case *ast.ValueSpec:
			ids = append(ids, sp.Names...)
		}
	}
	return ids
}

func mixed(n string) string {
	parts := strings.Split(n, "_")
	out := parts[0]
	for _, p := range parts[1:] {
		if p != "" {
			out += strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return out
}

func exists(root string) func(string) bool {
	return func(p string) bool {
		_, err := os.Stat(filepath.Join(root, filepath.FromSlash(p)))
		return err == nil
	}
}

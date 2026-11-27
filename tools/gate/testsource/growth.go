package testsource

import (
	"fmt"
	"go/parser"
	"strings"
	"unicode"

	"scout/tools/gate/check"
)

var genericWords = map[string]bool{"test": true, "case": true, "todo": true, "tbd": true, "placeholder": true, "example": true}

func BehaviourWords(name string) []string {
	name = strings.TrimPrefix(name, "Test")
	var words []string
	for _, part := range strings.Split(name, "_") {
		words = append(words, splitCase(part)...)
	}
	var out []string
	for _, w := range words {
		w = strings.ToLower(w)
		if w != "" && !genericWords[w] {
			out = append(out, w)
		}
	}
	return out
}

func splitCase(s string) []string {
	rs := []rune(s)
	var out []string
	start := 0
	for i := 1; i < len(rs); i++ {
		prev, cur := rs[i-1], rs[i]
		nextLower := i+1 < len(rs) && unicode.IsLower(rs[i+1])
		if unicode.IsUpper(cur) && (unicode.IsLower(prev) || unicode.IsDigit(prev) || (unicode.IsUpper(prev) && nextLower)) {
			out = append(out, string(rs[start:i]))
			start = i
		}
	}
	return append(out, string(rs[start:]))
}

func runGrowth(r *check.Run) ([]check.Finding, error) {
	c, err := loadCorpus(r)
	if err != nil {
		return nil, err
	}
	chs, err := c.changes()
	if err != nil {
		return nil, err
	}
	var out []check.Finding
	known := map[string]map[string]bool{}
	for _, ch := range chs {
		if !strings.HasSuffix(ch.Path, "_test.go") {
			continue
		}
		f := c.byPath[ch.Path]
		if f == nil {
			continue
		}
		if f.Err != nil {
			out = append(out, check.Finding{File: c.display(f.Path), Rule: "TEST.growth.justified", Problem: fmt.Sprintf("changed test file cannot be collected: %v", f.Err), Fix: "fix the syntax so its test functions can be collected"})
			continue
		}
		if known[f.Dir] == nil {
			if known[f.Dir], err = c.baseTests(f.Dir); err != nil {
				return nil, err
			}
		}
		out = append(out, growthFindings(c, f, known[f.Dir], ch.Base)...)
	}
	return out, nil
}

func growthOld(c *corpus, f *goFile, baseSrc []byte) []string {
	af, err := parser.ParseFile(c.fset, f.Path, baseSrc, parser.SkipObjectResolution)
	if err != nil {
		return nil
	}
	var out []string
	for _, fn := range topTests(af) {
		out = append(out, fn.Name.Name)
	}
	return out
}

func growthFindings(c *corpus, f *goFile, base map[string]bool, baseSrc []byte) []check.Finding {
	old := map[string]bool{}
	for _, n := range growthOld(c, f, baseSrc) {
		old[n] = true
	}
	var out []check.Finding
	for _, fn := range topTests(f.AST) {
		n := fn.Name.Name
		if base[n] || old[n] {
			continue
		}
		if words := BehaviourWords(n); len(words) < 2 {
			out = append(out, check.Finding{File: c.display(f.Path), Line: c.line(fn.Pos()), Rule: "TEST.growth.justified", Problem: fmt.Sprintf("new test %s does not name the behaviour it proves (needs two words besides test, case, todo, tbd, placeholder or example)", n), Fix: "rename it to state the condition and outcome it checks, or remove it"})
		}
	}
	return out
}

package naming

import (
	"strings"

	"scout/tools/gate/check"
	"scout/tools/gate/pathscope"
)

func Checks() []check.Check {
	return []check.Check{{Name: "naming", Rules: []string{ruleConventions, ruleTestFiles}, Run: run}}
}

func underTestdata(p string) bool {
	for _, s := range strings.Split(p, "/") {
		if s == "testdata" {
			return true
		}
	}
	return false
}

func run(r *check.Run) ([]check.Finding, error) {
	var lay layout
	if r.Proofs != nil {
		if _, err := r.Proofs.Section("layout", &lay); err != nil {
			return nil, err
		}
	}
	s, err := pathscope.New(r)
	if err != nil {
		return nil, err
	}
	j := judge{seen: map[string]bool{}, lay: lay, excepted: r.Proofs.Excepted, isDir: exists(scanRoot(r, s))}
	var out []check.Finding
	for _, e := range s.Entries {
		out = append(out, j.pathFindings(e)...)
		out = append(out, j.testFile(e)...)
		fs, err := sourceFindings(r.Explicit, s, e)
		if err != nil {
			return nil, err
		}
		out = append(out, fs...)
	}
	return out, nil
}

func scanRoot(r *check.Run, s *pathscope.Scope) string {
	if s.Fixture != "" {
		return s.Fixture
	}
	return r.Root
}

func sourceFindings(explicit bool, s *pathscope.Scope, e pathscope.Entry) ([]check.Finding, error) {
	if !strings.HasSuffix(e.Path, ".go") || (!explicit && underTestdata(e.Path)) {
		return nil, nil
	}
	src, err := s.Read(e)
	if err != nil {
		return nil, err
	}
	return goSymbols(e.Path, src)
}

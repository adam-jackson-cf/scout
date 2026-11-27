package structure

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"scout/tools/gate/check"
	"scout/tools/gate/pathscope"
)

const rulePrefix = "PREFIX.candidates"

func concept(name string) string {
	stem := pathscope.Stem(name)
	stem = strings.TrimSuffix(stem, "_test")
	return strings.TrimPrefix(stem, "test_")
}

func firstToken(name string) string {
	toks := pathscope.Tokens(concept(name))
	if len(toks) == 0 {
		return ""
	}
	return toks[0]
}

func children(files []string) map[string]map[string]bool {
	out := map[string]map[string]bool{}
	for _, f := range files {
		segs := strings.Split(f, "/")
		for i := range segs {
			parent := strings.Join(segs[:i], "/")
			if out[parent] == nil {
				out[parent] = map[string]bool{}
			}
			out[parent][segs[i]] = true
		}
	}
	return out
}

type member struct {
	Path, Parent, Name string
}

func clusters(min int, base, head map[string]map[string]bool, added []member, excepted func(string) bool) []check.Finding {
	var out []check.Finding
	for _, m := range added {
		if f, ok := cluster(min, base, head, m, excepted); ok {
			out = append(out, f)
		}
	}
	return out
}

func cluster(min int, base, head map[string]map[string]bool, m member, excepted func(string) bool) (check.Finding, bool) {
	if pathscope.FixedFile(m.Name) || strings.HasPrefix(m.Name, ".") {
		return check.Finding{}, false
	}
	if hasConcept(base[m.Parent], concept(m.Name)) {
		return check.Finding{}, false
	}
	tok := firstToken(m.Name)
	if tok == "" {
		return check.Finding{}, false
	}
	names := tokenConcepts(head[m.Parent], tok)
	if len(names) < min || excepted(m.Path) {
		return check.Finding{}, false
	}
	pattern := path.Join(m.Parent, tok+"*")
	return check.Finding{File: m.Path, Rule: rulePrefix,
		Problem: fmt.Sprintf("joins the sibling prefix cluster %s (%d members: %s)", pattern, len(names), strings.Join(names, ", ")),
		Fix:     fmt.Sprintf("extract the concept into %s/ or record a proofs.toml [[exception]] with rule %q and path %q", path.Join(m.Parent, tok), rulePrefix, pattern)}, true
}

func hasConcept(siblings map[string]bool, c string) bool {
	for n := range siblings {
		if concept(n) == c {
			return true
		}
	}
	return false
}

func tokenConcepts(siblings map[string]bool, tok string) []string {
	set := map[string]bool{}
	for n := range siblings {
		if firstToken(n) == tok && !pathscope.FixedFile(n) {
			set[concept(n)] = true
		}
	}
	return sortedKeys(set)
}

func runPrefix(r *check.Run) ([]check.Finding, error) {
	c, _, err := load(r)
	if err != nil {
		return nil, err
	}
	if c.PrefixClusterMin < 2 {
		return nil, errors.New("proofs.toml [structure].prefix_cluster_min must be at least 2")
	}
	if r.Explicit {
		return nil, errors.New("prefix needs base and head: run it with --base, or through the scenario canary")
	}
	s, err := pathscope.New(r)
	if err != nil || s.Base == "" {
		return nil, err
	}
	baseFiles, err := s.TreeFiles(s.Base)
	if err != nil {
		return nil, err
	}
	headFiles, err := s.HeadFiles()
	if err != nil {
		return nil, err
	}
	var added []member
	for _, e := range s.Entries {
		segs := strings.Split(e.Path, "/")
		added = append(added, member{Path: e.Path, Parent: strings.Join(segs[:e.FirstNew], "/"), Name: segs[e.FirstNew]})
	}
	excepted := func(p string) bool { return r.Proofs.Excepted(rulePrefix, p, "") }
	return clusters(c.PrefixClusterMin, children(baseFiles), children(headFiles), added, excepted), nil
}

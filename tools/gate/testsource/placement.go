package testsource

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"scout/tools/gate/check"
)

var tierFix = "use no build constraint (T1), //go:build boundary (T2), //go:build deferred or //go:build boundary && deferred"

func ValidTier(expr string, t Tiers) bool {
	switch expr {
	case "", t.Boundary, t.Deferred, t.Boundary + " && " + t.Deferred, t.Deferred + " && " + t.Boundary:
		return true
	}
	return false
}

func runPlacement(r *check.Run) ([]check.Finding, error) {
	c, err := loadCorpus(r)
	if err != nil {
		return nil, err
	}
	chs, err := c.changes()
	if err != nil {
		return nil, err
	}
	owners := packageDirs(c)
	var out []check.Finding
	for _, ch := range chs {
		if ch.Status != "A" && ch.Status != "R" && ch.Status != "C" {
			continue
		}
		out = append(out, placementFindings(c, ch.Path, owners)...)
	}
	return out, nil
}

type pkgDirs struct {
	code  map[string]bool
	tests map[string]bool
	names map[string][]string
}

func packageDirs(c *corpus) pkgDirs {
	p := pkgDirs{code: map[string]bool{}, tests: map[string]bool{}, names: map[string][]string{}}
	for _, f := range c.files {
		if f.AST == nil || inFixtures(f.Path, c.set.Layout.FixturesDir) {
			continue
		}
		if f.Test {
			p.tests[f.Dir] = true
			continue
		}
		if !p.code[f.Dir] {
			p.names[f.AST.Name.Name] = append(p.names[f.AST.Name.Name], f.Dir)
		}
		p.code[f.Dir] = true
	}
	return p
}

func placementFindings(c *corpus, p string, owners pkgDirs) []check.Finding {
	f := c.byPath[p]
	if inFixtures(p, c.set.Layout.FixturesDir) {
		return fixtureFinding(c, p, owners)
	}
	if f == nil || f.AST == nil {
		return nil
	}
	if f.Test {
		return testFilePlacement(c, f, owners)
	}
	if c.production(f) && importsTesting(f) {
		return []check.Finding{{File: c.display(p), Rule: "TEST.placement", Problem: "test support code sits outside the declared support layout", Fix: supportFix(c)}}
	}
	return nil
}

func testFilePlacement(c *corpus, f *goFile, owners pkgDirs) []check.Finding {
	var out []check.Finding
	if !ValidTier(f.Constraint, c.set.Tiers) {
		out = append(out, check.Finding{File: c.display(f.Path), Rule: "TEST.placement", Problem: fmt.Sprintf("test file constraint %q selects no single test tier", f.Constraint), Fix: tierFix})
	}
	if !owners.code[f.Dir] {
		out = append(out, check.Finding{File: c.display(f.Path), Rule: "TEST.placement", Problem: "test file sits in a directory with no package code to own it", Fix: "move it to " + destination(owners, strings.TrimSuffix(f.AST.Name.Name, "_test"))})
	}
	return out
}

func supportFix(c *corpus) string {
	if c.set.Layout.SupportDir == "" {
		return "move it into the _test.go file of the package that uses it"
	}
	return "move it to " + c.set.Layout.SupportDir + "/ or into the _test.go file of the package that uses it"
}

func destination(owners pkgDirs, pkg string) string {
	dirs := append([]string(nil), owners.names[pkg]...)
	sort.Strings(dirs)
	if len(dirs) == 1 {
		return dirs[0] + "/ beside package " + pkg
	}
	return "the directory of the package it tests"
}

func fixtureFinding(c *corpus, p string, owners pkgDirs) []check.Finding {
	segs := strings.Split(p, "/")
	owner := "."
	for i, s := range segs {
		if s == c.set.Layout.FixturesDir {
			owner = path.Join(segs[:i]...)
			break
		}
	}
	if owner == "" {
		owner = "."
	}
	if owners.tests[owner] {
		return nil
	}
	return []check.Finding{{File: c.display(p), Rule: "TEST.placement", Problem: fmt.Sprintf("fixture sits under %s/%s, but no test in %s loads fixtures there", owner, c.set.Layout.FixturesDir, owner), Fix: fmt.Sprintf("move it to <package>/%s/ beside the tests that load it", c.set.Layout.FixturesDir)}}
}

func importsTesting(f *goFile) bool {
	for _, im := range f.AST.Imports {
		if im.Path.Value == "\"testing\"" {
			return true
		}
	}
	return false
}

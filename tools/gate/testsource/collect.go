package testsource

import (
	"fmt"

	"scout/tools/gate/check"
)

type StaticTest struct {
	ImportPath string
	Dir        string
	File       string
	Display    string
	Name       string
	Constraint string
}

type Collection struct {
	Root    string
	Module  string
	Fixture bool
	Tiers   Tiers
	Exclude []string
	Tests   []StaticTest
	Files   []StaticFile
}

type StaticFile struct {
	Path       string
	Display    string
	Dir        string
	Constraint string
	Err        error
}

func Collect(r *check.Run) (*Collection, error) {
	c, err := loadFrom(r, true)
	if err != nil {
		return nil, err
	}
	col := &Collection{Root: c.root, Module: c.module, Fixture: c.fixture, Tiers: c.set.Tiers, Exclude: c.set.Layout.Exclude}
	for _, f := range c.files {
		if !f.Test || inFixtures(f.Path, c.set.Layout.FixturesDir) {
			continue
		}
		col.Files = append(col.Files, StaticFile{Path: f.Path, Display: c.display(f.Path), Dir: f.Dir, Constraint: f.Constraint, Err: f.Err})
		if f.AST == nil {
			continue
		}
		for _, fn := range topTests(f.AST) {
			col.Tests = append(col.Tests, StaticTest{ImportPath: c.importPath(f.Dir), Dir: f.Dir, File: f.Path, Display: c.display(f.Path), Name: fn.Name.Name, Constraint: f.Constraint})
		}
	}
	return col, nil
}

func (s StaticTest) ID() string { return fmt.Sprintf("%s::%s", s.Display, s.Name) }

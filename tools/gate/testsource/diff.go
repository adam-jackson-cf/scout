package testsource

import (
	"go/ast"
	"go/parser"
	"path"
	"strings"

	"scout/tools/gate/check"
)

type changed struct {
	check.Change
	Base []byte
}

func (c *corpus) changes() ([]changed, error) {
	if c.run.Explicit {
		var out []changed
		for _, p := range c.explicit {
			out = append(out, changed{Change: check.Change{Status: "A", Path: p}})
		}
		return out, nil
	}
	raw, err := c.run.Diff()
	if err != nil {
		return nil, err
	}
	base, err := c.baseRev()
	if err != nil || base == "" {
		return nil, err
	}
	var out []changed
	for _, ch := range raw {
		if !Excluded(c.set.Layout.Exclude, ch.Path) && ch.Status != "D" {
			out = append(out, c.withBase(ch, base))
		}
	}
	return out, nil
}

func (c *corpus) withBase(ch check.Change, base string) changed {
	item := changed{Change: ch}
	from := ch.Path
	if ch.From != "" {
		from = ch.From
	}
	if ch.Status != "A" {
		if src, err := c.run.Git("show", base+":"+from); err == nil {
			item.Base = []byte(src)
		}
	}
	return item
}

func (c *corpus) baseRev() (string, error) {
	if c.run.Mode == check.Commit {
		if _, err := c.run.Git("rev-parse", "--verify", "--quiet", "HEAD^{commit}"); err != nil {
			return "", nil
		}
		return "HEAD", nil
	}
	if c.run.Base == "" {
		return "", nil
	}
	mb, err := c.run.Git("merge-base", c.run.Base, "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(mb), nil
}

func (c *corpus) baseTests(dir string) (map[string]bool, error) {
	names := map[string]bool{}
	if c.run.Explicit {
		return names, nil
	}
	base, err := c.baseRev()
	if err != nil || base == "" {
		return names, err
	}
	raw, err := c.run.Git("ls-tree", "--name-only", "-z", base, dir+"/")
	if err != nil {
		return names, nil
	}
	for _, p := range strings.Split(raw, "\x00") {
		if strings.HasSuffix(p, "_test.go") && path.Dir(p) == dir {
			if err := c.addBaseNames(base, p, names); err != nil {
				return nil, err
			}
		}
	}
	return names, nil
}

func (c *corpus) addBaseNames(base, p string, names map[string]bool) error {
	src, err := c.run.Git("show", base+":"+p)
	if err != nil {
		return err
	}
	af, err := parser.ParseFile(c.fset, p, src, parser.SkipObjectResolution)
	if err != nil {
		return nil
	}
	for _, fn := range topTests(af) {
		names[fn.Name.Name] = true
	}
	return nil
}

func topTests(af *ast.File) []*ast.FuncDecl {
	var out []*ast.FuncDecl
	for _, d := range af.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if ok && fn.Recv == nil && isTestName(fn.Name.Name) {
			out = append(out, fn)
		}
	}
	return out
}

func isTestName(n string) bool {
	if !strings.HasPrefix(n, "Test") || n == "TestMain" {
		return false
	}
	rest := n[len("Test"):]
	return rest == "" || rest[0] < 'a' || rest[0] > 'z'
}

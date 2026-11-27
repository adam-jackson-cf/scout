package testsource

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/build/constraint"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"scout/tools/gate/check"
)

type goFile struct {
	Path       string
	Dir        string
	Test       bool
	Src        []byte
	AST        *ast.File
	Err        error
	Constraint string
}

type layout struct {
	Exclude     []string `toml:"exclude"`
	SupportDir  string   `toml:"support_dir"`
	FixturesDir string   `toml:"fixtures_dir"`
	Tiers       Tiers    `toml:"test_tiers"`
}

type structure struct {
	AdapterInterfaces []string `toml:"adapter_interfaces"`
}

type settings struct {
	Layout    layout
	Structure structure
	Tiers     Tiers
}

type Tiers struct {
	Boundary       string `toml:"boundary_tag"`
	Deferred       string `toml:"deferred_tag"`
	PackageTimeout string `toml:"package_timeout"`
}

type corpus struct {
	run      *check.Run
	root     string
	module   string
	fset     *token.FileSet
	files    []*goFile
	byPath   map[string]*goFile
	set      settings
	fixture  bool
	unstaged map[string]bool
	explicit []string
	worktree bool
	callers  []*goFile
}

func LoadTiers(r *check.Run) (Tiers, error) {
	var l layout
	if _, err := r.Proofs.Section("layout", &l); err != nil {
		return l.Tiers, fmt.Errorf("proofs.toml [layout]: %w", err)
	}
	t := l.Tiers
	if t.Boundary == "" || t.Deferred == "" || t.PackageTimeout == "" {
		return t, errors.New("proofs.toml [layout.test_tiers] must set boundary_tag, deferred_tag and package_timeout")
	}
	return t, nil
}

func loadSettings(r *check.Run) (settings, error) {
	var s settings
	if _, err := r.Proofs.Section("layout", &s.Layout); err != nil {
		return s, fmt.Errorf("proofs.toml [layout]: %w", err)
	}
	if _, err := r.Proofs.Section("structure", &s.Structure); err != nil {
		return s, fmt.Errorf("proofs.toml [structure]: %w", err)
	}
	t, err := LoadTiers(r)
	if err != nil {
		return s, err
	}
	s.Tiers = t
	if s.Layout.FixturesDir == "" {
		return s, errors.New("proofs.toml [layout] must set fixtures_dir")
	}
	return s, nil
}

func Excluded(patterns []string, p string) bool {
	for _, pat := range patterns {
		if strings.HasSuffix(pat, "/**") {
			dir := strings.TrimSuffix(pat, "/**")
			if strings.HasPrefix(p, dir+"/") {
				return true
			}
			continue
		}
		if ok, _ := path.Match(pat, p); ok {
			return true
		}
	}
	return false
}

func loadCorpus(r *check.Run) (*corpus, error) {
	return loadFrom(r, false)
}

func loadFrom(r *check.Run, worktree bool) (*corpus, error) {
	set, err := loadSettings(r)
	if err != nil {
		return nil, err
	}
	c := &corpus{run: r, root: r.Root, fset: token.NewFileSet(), byPath: map[string]*goFile{}, set: set, worktree: worktree}
	paths, err := c.listPaths()
	if err != nil {
		return nil, err
	}
	if c.module, err = modulePath(c.root); err != nil {
		return nil, err
	}
	for _, p := range paths {
		if err := c.add(p); err != nil {
			return nil, err
		}
	}
	if err := c.addExcludedCallers(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *corpus) addExcludedCallers() error {
	if c.run.Explicit {
		return nil
	}
	tracked, err := c.run.Tracked()
	if err != nil {
		return err
	}
	for _, p := range tracked {
		if !excludedCaller(c.set.Layout.Exclude, p) {
			continue
		}
		src, err := os.ReadFile(filepath.Join(c.root, filepath.FromSlash(p)))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		f := &goFile{Path: p, Dir: path.Dir(p), Src: src}
		if f.AST, f.Err = parser.ParseFile(c.fset, p, src, 0); f.Err == nil {
			c.callers = append(c.callers, f)
		}
	}
	return nil
}

func excludedCaller(patterns []string, p string) bool {
	if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") || !Excluded(patterns, p) {
		return false
	}
	return !inFixtures(p, "testdata")
}

func (c *corpus) listPaths() ([]string, error) {
	if c.run.Explicit {
		return c.explicitPaths()
	}
	tracked, err := c.run.Tracked()
	if err != nil {
		return nil, err
	}
	unstaged, err := c.run.Git("diff", "--name-only", "-z")
	if err != nil {
		return nil, err
	}
	c.unstaged = map[string]bool{}
	for _, p := range strings.Split(unstaged, "\x00") {
		c.unstaged[p] = true
	}
	var out []string
	for _, p := range tracked {
		if strings.HasSuffix(p, ".go") && !Excluded(c.set.Layout.Exclude, p) {
			out = append(out, p)
		}
	}
	return out, nil
}

func (c *corpus) explicitPaths() ([]string, error) {
	var out []string
	for _, p := range c.run.Paths {
		abs := p
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(c.run.Root, p)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			rel, err := filepath.Rel(c.run.Root, abs)
			if err != nil {
				return nil, err
			}
			out = append(out, filepath.ToSlash(rel))
			c.explicit = append(c.explicit, filepath.ToSlash(rel))
			continue
		}
		c.root, c.fixture = abs, true
		err = filepath.WalkDir(abs, func(fp string, d fs.DirEntry, werr error) error {
			if werr != nil || d.IsDir() {
				return werr
			}
			rel, rerr := filepath.Rel(abs, fp)
			c.explicit = append(c.explicit, filepath.ToSlash(rel))
			if strings.HasSuffix(fp, ".go") {
				out = append(out, filepath.ToSlash(rel))
			}
			return rerr
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(out)
	return out, nil
}

func modulePath(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("owned module: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if f := strings.Fields(line); len(f) == 2 && f[0] == "module" {
			return f[1], nil
		}
	}
	return "", errors.New("go.mod declares no module path")
}

func (c *corpus) read(p string) ([]byte, error) {
	if c.run.Mode == check.Commit && !c.run.Explicit && !c.worktree && c.unstaged[p] {
		out, err := c.run.Git("show", ":"+p)
		return []byte(out), err
	}
	return os.ReadFile(filepath.Join(c.root, filepath.FromSlash(p)))
}

func (c *corpus) add(p string) error {
	src, err := c.read(p)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	f := &goFile{Path: p, Dir: path.Dir(p), Test: strings.HasSuffix(p, "_test.go"), Src: src}
	f.AST, f.Err = parser.ParseFile(c.fset, p, src, parser.ParseComments)
	f.Constraint = buildConstraint(src)
	c.files = append(c.files, f)
	c.byPath[p] = f
	return nil
}

func buildConstraint(src []byte) string {
	for _, line := range bytes.Split(src, []byte("\n")) {
		s := strings.TrimSpace(string(line))
		if constraint.IsGoBuild(s) {
			if e, err := constraint.Parse(s); err == nil {
				return e.String()
			}
			return s
		}
		if strings.HasPrefix(s, "package ") {
			return ""
		}
	}
	return ""
}

func (c *corpus) production(f *goFile) bool {
	if f.Test || f.AST == nil {
		return false
	}
	if c.set.Layout.SupportDir != "" && (f.Dir == c.set.Layout.SupportDir || strings.HasPrefix(f.Dir, c.set.Layout.SupportDir+"/")) {
		return false
	}
	return !inFixtures(f.Path, c.set.Layout.FixturesDir)
}

func inFixtures(p, dir string) bool {
	for _, seg := range strings.Split(path.Dir(p), "/") {
		if seg == dir {
			return true
		}
	}
	return false
}

func (c *corpus) importPath(dir string) string {
	if dir == "." {
		return c.module
	}
	return c.module + "/" + dir
}

func (c *corpus) line(pos token.Pos) int { return c.fset.Position(pos).Line }

func (c *corpus) display(p string) string {
	if c.fixture {
		rel, err := filepath.Rel(c.run.Root, filepath.Join(c.root, filepath.FromSlash(p)))
		if err == nil && !strings.HasPrefix(rel, "..") {
			return filepath.ToSlash(rel)
		}
	}
	return p
}

package tooling

import (
	"io/fs"
	"path"
	"path/filepath"
	"sort"

	"scout/internal/toolchain"
	"scout/tools/gate/check"
)

func runParity(r *check.Run) ([]check.Finding, error) {
	var trees []toolchain.Tree
	if r.Explicit {
		for _, p := range r.Paths {
			t, err := fixtureTree(r.Root, p)
			if err != nil {
				return nil, err
			}
			trees = append(trees, t)
		}
	} else {
		files, err := selectFiles(r, func(string, string) bool { return true })
		if err != nil {
			return nil, err
		}
		trees = append(trees, toolchain.Tree{Root: r.Root, Files: files, ToolsDir: toolsDir(r.Root)})
	}
	var out []check.Finding
	for i, t := range trees {
		problems, err := toolchain.Check(t)
		if err != nil {
			return nil, err
		}
		prefix := ""
		if r.Explicit {
			prefix = r.Paths[i]
		}
		for _, p := range problems {
			out = append(out, check.Finding{File: path.Join(prefix, p.File), Line: p.Line, Rule: p.Rule, Problem: p.Problem, Fix: p.Fix})
		}
	}
	return out, nil
}

func fixtureTree(root, rel string) (toolchain.Tree, error) {
	dir := filepath.Join(root, rel)
	var files []string
	err := filepath.WalkDir(dir, func(file string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		f, err := filepath.Rel(dir, file)
		if err == nil {
			files = append(files, filepath.ToSlash(f))
		}
		return err
	})
	sort.Strings(files)
	return toolchain.Tree{Root: dir, Files: files}, err
}

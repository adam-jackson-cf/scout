package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"scout/tools/gate/basediff"
	"scout/tools/gate/check"
	"scout/tools/gate/codeintegrity"
	"scout/tools/gate/featuremap"
	"scout/tools/gate/naming"
	"scout/tools/gate/proof"
	"scout/tools/gate/pyquality"
	"scout/tools/gate/skill"
	"scout/tools/gate/structure"
	"scout/tools/gate/testrun"
	"scout/tools/gate/testsource"
	"scout/tools/gate/tooling"
)

func registry() []check.Check {
	var all []check.Check
	for _, f := range []func() []check.Check{
		tooling.Checks, codeintegrity.Checks, pyquality.Checks,
		structure.Checks, naming.Checks, skill.Checks, featuremap.Checks, basediff.Checks,
		testsource.Checks, testrun.Checks,
	} {
		all = append(all, f()...)
	}
	return all
}

type pathList []string

func (p *pathList) String() string     { return strings.Join(*p, ",") }
func (p *pathList) Set(v string) error { *p = append(*p, v); return nil }

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	index, err := check.IsolateGitEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, "gate:", err)
		return 2
	}
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: go run ./tools/gate commit|ci|list [--base REF] [--only NAME,...] [--path P ...]")
		return 2
	}
	mode := args[0]
	fs := flag.NewFlagSet("gate", flag.ContinueOnError)
	base := fs.String("base", "", "base ref for ci diffs")
	only := fs.String("only", "", "comma-separated check names")
	var paths pathList
	fs.Var(&paths, "path", "explicit path to judge (repeatable)")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	checks := registry()
	if mode == "list" {
		for _, c := range checks {
			fmt.Printf("%s\t%s\n", c.Name, strings.Join(c.Rules, ","))
		}
		return 0
	}
	if mode != string(check.Commit) && mode != string(check.CI) {
		fmt.Fprintf(os.Stderr, "gate: unknown mode %q\n", mode)
		return 2
	}
	r, err := newRun(check.Mode(mode), *base, paths, index)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gate:", err)
		return 2
	}
	defer removeTemp(r.TempDir)
	return execute(r, selectChecks(checks, r.Mode, *only))
}

func newRun(mode check.Mode, base string, paths []string, index string) (*check.Run, error) {
	r := &check.Run{Mode: mode, Base: base, Paths: paths, Explicit: len(paths) > 0, IndexFile: index}
	top, err := r.Git("rev-parse", "--show-toplevel")
	if err != nil {
		wd, werr := os.Getwd()
		if werr != nil {
			return nil, errors.Join(err, werr)
		}
		r.Root = wd
	} else {
		r.Root = strings.TrimSpace(top)
	}
	for i, p := range r.Paths {
		if rel, ok := repoRelative(r.Root, p); ok {
			r.Paths[i] = rel
		}
	}
	if r.Proofs, err = proof.Load(filepath.Join(r.Root, "proofs.toml")); err != nil {
		return nil, err
	}
	if r.TempDir, err = os.MkdirTemp("", "scout-gate-"); err != nil {
		return nil, err
	}
	return r, nil
}

func removeTemp(dir string) {
	if err := os.RemoveAll(dir); err != nil {
		fmt.Fprintln(os.Stderr, "gate:", err)
	}
}

func selectChecks(checks []check.Check, mode check.Mode, only string) []check.Check {
	want := map[string]bool{}
	for _, n := range strings.Split(only, ",") {
		if n != "" {
			want[n] = true
		}
	}
	var out []check.Check
	for _, c := range checks {
		if (len(want) == 0 || want[c.Name]) && c.Runs(mode) {
			out = append(out, c)
		}
	}
	return out
}

func execute(r *check.Run, checks []check.Check) int {
	status := 0
	for _, c := range checks {
		findings, err := c.Run(r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gate: %s: %v\n", c.Name, err)
			status = 1
			continue
		}
		for _, f := range findings {
			fmt.Println(f.String())
			status = 1
		}
	}
	if status == 0 {
		fmt.Fprintf(os.Stderr, "gate: %s: %d checks passed\n", r.Mode, len(checks))
	}
	return status
}

func repoRelative(root, p string) (string, bool) {
	if !filepath.IsAbs(p) {
		return "", false
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		resolvedRoot = root
	}
	resolved, err := filepath.EvalSymlinks(p)
	if err != nil {
		resolved = p
	}
	rel, err := filepath.Rel(resolvedRoot, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, "../") {
		return "", false
	}
	return filepath.ToSlash(rel), true
}

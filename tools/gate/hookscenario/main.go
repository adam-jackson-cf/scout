package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"scout/tools/gate/check"
)

const branch = "hookscenario"

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	if _, err := check.IsolateGitEnv(); err != nil {
		fmt.Fprintln(os.Stderr, "hookscenario:", err)
		return 2
	}
	fl := flag.NewFlagSet("hookscenario", flag.ContinueOnError)
	plant := fl.String("plant", "", "patch applied to the linked worktree's files before the commit")
	if err := fl.Parse(args); err != nil || fl.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: go run ./tools/gate/hookscenario [--plant patch]")
		return 2
	}
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "hookscenario:", err)
		return 2
	}
	problems, err := scenario(root, *plant)
	if err != nil {
		fmt.Fprintln(os.Stderr, "hookscenario:", err)
		return 2
	}
	for _, p := range problems {
		fmt.Println("hookscenario: FAIL:", p)
	}
	if len(problems) > 0 {
		return 1
	}
	fmt.Println("hookscenario: refs, branches, git config and submodule indexes unchanged after a commit through the installed hook from a linked worktree")
	return 0
}

type repos struct {
	clone, wt string
	subs      []string
}

func scenario(root, plant string) (problems []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Minute)
	defer cancel()
	tmp, err := os.MkdirTemp("", "scout-hookscenario-")
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(tmp)) }()
	if tmp, err = filepath.EvalSymlinks(tmp); err != nil {
		return nil, err
	}
	r, err := prepare(ctx, root, tmp)
	if err != nil {
		return nil, err
	}
	if err := applyPlant(ctx, r.wt, plant); err != nil {
		return nil, err
	}
	if err := stageRootChange(ctx, r.wt); err != nil {
		return nil, err
	}
	before, err := snapshot(ctx, r)
	if err != nil {
		return nil, err
	}
	problems, committed := commitThroughHook(ctx, r.wt)
	after, err := snapshot(ctx, r)
	if err != nil {
		return nil, err
	}
	return append(problems, compare(ctx, r, before, after, committed)...), nil
}

func applyPlant(ctx context.Context, wt, plant string) error {
	if plant == "" {
		return nil
	}
	abs, err := filepath.Abs(plant)
	if err != nil {
		return err
	}
	_, err = git(ctx, wt, "apply", abs)
	return err
}

func commitThroughHook(ctx context.Context, wt string) ([]string, bool) {
	out, err := git(ctx, wt, "commit", "-q", "-m", "hookscenario: commit through the installed hook")
	if err != nil {
		return []string{"the commit through the hook failed: " + tail(err.Error(), 4000)}, false
	}
	fmt.Print(tail(out, 2000))
	if !strings.Contains(out, "gate: commit:") {
		return []string{"the installed hook did not run the gate; output: " + tail(out, 2000)}, true
	}
	return nil, true
}

func prepare(ctx context.Context, root, tmp string) (repos, error) {
	r := repos{clone: filepath.Join(tmp, "clone"), wt: filepath.Join(tmp, "wt")}
	head, err := git(ctx, root, "rev-parse", "HEAD")
	if err != nil {
		return r, err
	}
	steps := [][]string{
		{"clone", "--quiet", "--shared", "--no-checkout", root, r.clone},
		{"-C", r.clone, "checkout", "-q", "-b", "scenario-base", strings.TrimSpace(head)},
		{"-C", r.clone, "config", "user.name", "hookscenario"},
		{"-C", r.clone, "config", "user.email", "hookscenario@invalid"},
	}
	for _, s := range steps {
		if _, err := git(ctx, tmp, s...); err != nil {
			return r, err
		}
	}
	if r.subs, err = submodules(ctx, root, r.clone); err != nil {
		return r, err
	}
	if err := checkoutSubmodules(ctx, r.clone); err != nil {
		return r, err
	}
	for _, s := range [][]string{{"config", "core.hooksPath", ".githooks"}, {"worktree", "add", "-q", "-b", branch, r.wt}} {
		if _, err := git(ctx, r.clone, s...); err != nil {
			return r, err
		}
	}
	return r, checkoutSubmodules(ctx, r.wt)
}

func stageRootChange(ctx context.Context, wt string) error {
	f, err := os.OpenFile(filepath.Join(wt, ".gitignore"), os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	_, werr := f.WriteString("/.hookscenario\n")
	if err := errors.Join(werr, f.Close()); err != nil {
		return err
	}
	_, err = git(ctx, wt, "add", ".gitignore")
	return err
}

func submodules(ctx context.Context, root, clone string) ([]string, error) {
	raw, err := git(ctx, clone, "config", "-f", ".gitmodules", "--get-regexp", "^submodule\\..*\\.path$")
	if err != nil {
		return nil, nil
	}
	var paths []string
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		key, p, ok := strings.Cut(line, " ")
		if !ok {
			continue
		}
		name := strings.TrimSuffix(strings.TrimPrefix(key, "submodule."), ".path")
		if _, err := git(ctx, clone, "config", "submodule."+name+".url", filepath.Join(root, p)); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return paths, nil
}

func checkoutSubmodules(ctx context.Context, dir string) error {
	_, err := git(ctx, dir, "-c", "protocol.file.allow=always", "submodule", "update", "--init", "--quiet")
	return err
}

func snapshot(ctx context.Context, r repos) (map[string]string, error) {
	s := map[string]string{}
	queries := map[string][]string{
		"refs":      {"for-each-ref", "--format=%(refname) %(objectname)"},
		"worktrees": {"worktree", "list", "--porcelain"},
	}
	for key, args := range queries {
		out, err := git(ctx, r.clone, args...)
		if err != nil {
			return nil, err
		}
		s[key] = out
	}
	dirs := map[string]string{"clone": r.clone, "wt": r.wt}
	for _, p := range r.subs {
		dirs["clone/"+p], dirs["wt/"+p] = filepath.Join(r.clone, p), filepath.Join(r.wt, p)
	}
	for name, dir := range dirs {
		if err := describe(ctx, s, name, dir); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func describe(ctx context.Context, s map[string]string, name, dir string) error {
	for _, q := range [][]string{{"HEAD", "rev-parse", "HEAD"}, {"branch", "symbolic-ref", "-q", "HEAD"}, {"index", "ls-files", "--stage"}, {"local-refs", "for-each-ref", "--format=%(refname) %(objectname)"}} {
		out, _ := git(ctx, dir, q[1:]...)
		s[name+" "+q[0]] = out
	}
	for _, f := range []string{"config", "config.worktree"} {
		p, err := git(ctx, dir, "rev-parse", "--path-format=absolute", "--git-path", f)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(strings.TrimSpace(p))
		if errors.Is(err, os.ErrNotExist) {
			s[name+" "+f] = "absent"
			continue
		}
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		s[name+" "+f] = hex.EncodeToString(sum[:])
	}
	return nil
}

func compare(ctx context.Context, r repos, before, after map[string]string, committed bool) []string {
	ref := "refs/heads/" + branch
	old, now := refValue(before["refs"], ref), refValue(after["refs"], ref)
	problems := refMove(ctx, r.clone, move{ref: ref, old: old, now: now, committed: committed})
	if !committed {
		now = old
	}
	expect := expected(before, r.wt, ref, old, now)
	keys := make([]string, 0, len(expect))
	for k := range expect {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if expect[k] != after[k] {
			problems = append(problems, fmt.Sprintf("%s changed:\n--- expected\n%s\n--- found\n%s", k, tail(expect[k], 1500), tail(after[k], 1500)))
		}
	}
	return problems
}

type move struct {
	ref, old, now string
	committed     bool
}

func refMove(ctx context.Context, clone string, m move) []string {
	switch {
	case !m.committed && m.now != m.old:
		return []string{m.ref + " moved although the commit failed"}
	case !m.committed:
		return nil
	case m.old == m.now:
		return []string{m.ref + " did not move; no commit was made"}
	}
	if parent, err := git(ctx, clone, "rev-parse", m.now+"^"); err != nil || strings.TrimSpace(parent) != m.old {
		return []string{m.ref + " moved to " + m.now + ", which is not a child of " + m.old}
	}
	return nil
}

func expected(before map[string]string, wt, ref, old, now string) map[string]string {
	expect := map[string]string{}
	for k, v := range before {
		expect[k] = v
	}
	expect["refs"] = strings.Replace(before["refs"], ref+" "+old, ref+" "+now, 1)
	expect["worktrees"] = moveWorktreeHead(before["worktrees"], wt, old, now)
	expect["wt HEAD"] = strings.Replace(before["wt HEAD"], old, now, 1)
	expect["clone local-refs"] = expect["refs"]
	expect["wt local-refs"] = expect["refs"]
	return expect
}

func moveWorktreeHead(porcelain, wt, old, now string) string {
	blocks := strings.Split(porcelain, "\n\n")
	for i, b := range blocks {
		if strings.HasPrefix(b, "worktree "+wt+"\n") || strings.Contains(b, "worktree "+wt+"\n") {
			blocks[i] = strings.Replace(b, "HEAD "+old, "HEAD "+now, 1)
		}
	}
	return strings.Join(blocks, "\n\n")
}

func refValue(refs, name string) string {
	for _, line := range strings.Split(refs, "\n") {
		if n, v, ok := strings.Cut(line, " "); ok && n == name {
			return v
		}
	}
	return ""
}

func git(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, tail(string(out), 4000))
	}
	return string(out), nil
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return "..." + s[len(s)-n:]
}

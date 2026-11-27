package check

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"scout/tools/gate/proof"
)

type Mode string

const (
	Commit Mode = "commit"
	CI     Mode = "ci"
)

type Finding struct {
	File    string
	Line    int
	Rule    string
	Problem string
	Fix     string
}

func (f Finding) String() string {
	loc := f.File
	if f.Line > 0 {
		loc = fmt.Sprintf("%s:%d", f.File, f.Line)
	}
	return fmt.Sprintf("%s: %s: %s; %s", loc, f.Rule, f.Problem, f.Fix)
}

type Change struct {
	Status string
	Path   string
	From   string
}

type Run struct {
	Root      string
	Mode      Mode
	Base      string
	Paths     []string
	Explicit  bool
	IndexFile string
	Proofs    *proof.File
	TempDir   string
}

type Check struct {
	Name  string
	Rules []string
	Modes []Mode
	Run   func(*Run) ([]Finding, error)
}

func (c Check) Runs(m Mode) bool {
	if len(c.Modes) == 0 {
		return true
	}
	for _, x := range c.Modes {
		if x == m {
			return true
		}
	}
	return false
}

func (r *Run) Git(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.Root
	cmd.Env = r.GitEnv()
	if len(args) > 0 && args[0] == "-C" {
		cmd.Env = os.Environ()
	}
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(errb.String()))
	}
	return out.String(), nil
}

func (r *Run) Diff() ([]Change, error) {
	if r.Explicit {
		out := make([]Change, 0, len(r.Paths))
		for _, p := range r.Paths {
			out = append(out, Change{Status: "A", Path: p})
		}
		return out, nil
	}
	args := []string{"diff", "--name-status", "-z", "--find-renames", "--no-ext-diff"}
	switch r.Mode {
	case Commit:
		args = append(args, "--cached")
	default:
		if r.Base == "" {
			return nil, nil
		}
		mb, err := r.Git("merge-base", r.Base, "HEAD")
		if err != nil {
			return nil, err
		}
		args = append(args, strings.TrimSpace(mb), "HEAD")
	}
	raw, err := r.Git(args...)
	if err != nil {
		return nil, err
	}
	return parseNameStatus(raw), nil
}

func parseNameStatus(raw string) []Change {
	parts := strings.Split(strings.TrimRight(raw, "\x00"), "\x00")
	var out []Change
	for i := 0; i < len(parts); i++ {
		st := parts[i]
		if st == "" {
			continue
		}
		if st[0] == 'R' || st[0] == 'C' {
			if i+2 >= len(parts) {
				break
			}
			out = append(out, Change{Status: st[:1], From: parts[i+1], Path: parts[i+2]})
			i += 2
			continue
		}
		if i+1 >= len(parts) {
			break
		}
		out = append(out, Change{Status: st[:1], Path: parts[i+1]})
		i++
	}
	return out
}

func (r *Run) Added() ([]string, error) {
	changes, err := r.Diff()
	if err != nil {
		return nil, err
	}
	var out []string
	for _, c := range changes {
		if c.Status == "A" || c.Status == "R" || c.Status == "C" {
			out = append(out, c.Path)
		}
	}
	sort.Strings(out)
	return out, nil
}

func (r *Run) Tracked() ([]string, error) {
	if r.Explicit {
		return append([]string(nil), r.Paths...), nil
	}
	raw, err := r.Git("ls-files", "-z", "--cached")
	if err != nil {
		return nil, err
	}
	var out []string
	for _, p := range strings.Split(raw, "\x00") {
		if p != "" {
			out = append(out, p)
		}
	}
	return out, nil
}

func IsolateGitEnv() (string, error) {
	var index string
	if idx := os.Getenv("GIT_INDEX_FILE"); idx != "" {
		abs, err := filepath.Abs(idx)
		if err != nil {
			return "", fmt.Errorf("GIT_INDEX_FILE %q: %w", idx, err)
		}
		index = abs
	}
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(k, "GIT_") && k != "GIT_CEILING_DIRECTORIES" {
			if err := os.Unsetenv(k); err != nil {
				return "", err
			}
		}
	}
	return index, nil
}

func (r *Run) GitEnv() []string {
	env := os.Environ()
	if r.IndexFile != "" {
		env = append(env, "GIT_INDEX_FILE="+r.IndexFile)
	}
	return env
}

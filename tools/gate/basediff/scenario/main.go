package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"scout/tools/gate/check"
)

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	if _, err := check.IsolateGitEnv(); err != nil {
		fmt.Fprintln(os.Stderr, "scenario:", err)
		return 2
	}
	fl := flag.NewFlagSet("scenario", flag.ContinueOnError)
	only := fl.String("only", "", "gate checks to run")
	if err := fl.Parse(args); err != nil || fl.NArg() != 1 || *only == "" {
		fmt.Fprintln(os.Stderr, "usage: go run ./tools/gate/basediff/scenario --only NAME,... <fixture with base/ and head/>")
		return 2
	}
	code, err := scenario(fl.Arg(0), *only)
	if err != nil {
		fmt.Fprintln(os.Stderr, "scenario:", err)
		return 2
	}
	return code
}

func scenario(fixture, only string) (code int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 110*time.Second)
	defer cancel()
	root, err := os.Getwd()
	if err != nil {
		return 0, err
	}
	tmp, err := os.MkdirTemp("", "scout-gate-scenario-")
	if err != nil {
		return 0, err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(tmp)) }()
	env, hooks, err := sandboxEnv(tmp)
	if err != nil {
		return 0, err
	}
	bin := filepath.Join(tmp, "gate")
	if _, err := cmd(ctx, root, env, "go", "build", "-o", bin, "./tools/gate"); err != nil {
		return 0, err
	}
	repo, err := initRepo(ctx, tmp, env, hooks)
	if err != nil {
		return 0, err
	}
	proofs, err := os.ReadFile(filepath.Join(root, "proofs.toml"))
	if err != nil {
		return 0, err
	}
	sb := sandbox{repo: repo, env: env, proofs: proofs}
	base, err := commitSides(ctx, sb, fixture)
	if err != nil {
		return 0, err
	}
	return runGate(ctx, bin, sb, base, only)
}

type sandbox struct {
	repo   string
	env    []string
	proofs []byte
}

func sandboxEnv(tmp string) ([]string, string, error) {
	gitconfig := filepath.Join(tmp, "gitconfig")
	hooks := filepath.Join(tmp, "hooks")
	if err := os.WriteFile(gitconfig, nil, 0o600); err != nil {
		return nil, "", err
	}
	if err := os.Mkdir(hooks, 0o700); err != nil {
		return nil, "", err
	}
	env := append(os.Environ(), "GIT_CONFIG_GLOBAL="+gitconfig, "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=scenario", "GIT_AUTHOR_EMAIL=scenario@invalid", "GIT_COMMITTER_NAME=scenario", "GIT_COMMITTER_EMAIL=scenario@invalid")
	return env, hooks, nil
}

func initRepo(ctx context.Context, tmp string, env []string, hooks string) (string, error) {
	repo := filepath.Join(tmp, "repo")
	if _, err := cmd(ctx, tmp, env, "git", "init", "-q", repo); err != nil {
		return "", err
	}
	if _, err := cmd(ctx, repo, env, "git", "config", "core.hooksPath", hooks); err != nil {
		return "", err
	}
	return repo, nil
}

func commitSides(ctx context.Context, sb sandbox, fixture string) (string, error) {
	var base string
	for _, side := range []string{"base", "head"} {
		if err := commitSide(ctx, sb, fixture, side); err != nil {
			return "", err
		}
		if side == "base" {
			out, err := cmd(ctx, sb.repo, sb.env, "git", "rev-parse", "HEAD")
			if err != nil {
				return "", err
			}
			base = strings.TrimSpace(out)
		}
	}
	return base, nil
}

func commitSide(ctx context.Context, sb sandbox, fixture, side string) error {
	if err := reset(sb.repo); err != nil {
		return err
	}
	if err := copyTree(filepath.Join(fixture, side), sb.repo); err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(sb.repo, "proofs.toml")); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(filepath.Join(sb.repo, "proofs.toml"), sb.proofs, 0o644); err != nil {
			return err
		}
	}
	if _, err := cmd(ctx, sb.repo, sb.env, "git", "add", "-A"); err != nil {
		return err
	}
	_, err := cmd(ctx, sb.repo, sb.env, "git", "commit", "-q", "-m", side)
	return err
}

func runGate(ctx context.Context, bin string, sb sandbox, base, only string) (int, error) {
	gate := exec.CommandContext(ctx, bin, "ci", "--base", base, "--only", only)
	gate.Dir, gate.Env, gate.Stdout, gate.Stderr = sb.repo, sb.env, os.Stdout, os.Stderr
	err := gate.Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return exit.ExitCode(), nil
	}
	return 0, err
}

func cmd(ctx context.Context, dir string, env []string, name string, args ...string) (string, error) {
	c := exec.CommandContext(ctx, name, args...)
	var out, errb bytes.Buffer
	c.Dir, c.Env, c.Stdout, c.Stderr = dir, env, &out, &errb
	if err := c.Run(); err != nil {
		return "", fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(errb.String()))
	}
	return out.String(), nil
}

func reset(repo string) error {
	entries, err := os.ReadDir(repo)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.Name() == ".git" {
			continue
		}
		if err := os.RemoveAll(filepath.Join(repo, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

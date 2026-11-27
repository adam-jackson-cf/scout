//go:build boundary

package codeintegrity

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"scout/tools/gate/check"
	"scout/tools/gate/proof"
)

type repo struct {
	t    *testing.T
	root string
}

func newRepo(t *testing.T) *repo {
	t.Helper()
	r := &repo{t: t, root: t.TempDir()}
	r.git("init", "-q", "-b", "main")
	r.git("config", "user.email", "gate@example.com")
	r.git("config", "user.name", "gate")
	r.git("config", "commit.gpgsign", "false")
	return r
}

func (r *repo) git(args ...string) string {
	r.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.root
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return string(out)
}

func (r *repo) write(rel, body string) {
	r.t.Helper()
	p := filepath.Join(r.root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		r.t.Fatal(err)
	}
}

func (r *repo) commit(msg string) {
	r.git("add", "-A")
	r.git("commit", "-q", "-m", msg)
}

func (r *repo) run(mode check.Mode, base string) map[string][]check.Finding {
	r.t.Helper()
	pf, err := proof.Load(filepath.Join(r.root, "proofs.toml"))
	if err != nil {
		r.t.Fatal(err)
	}
	run := &check.Run{Root: r.root, Mode: mode, Base: base, Proofs: pf, TempDir: r.t.TempDir()}
	out := map[string][]check.Finding{}
	for _, c := range Checks() {
		found, err := c.Run(run)
		if err != nil {
			r.t.Fatalf("%s: %v", c.Name, err)
		}
		out[c.Name] = found
	}
	return out
}

func lines(found []check.Finding) string {
	var parts []string
	for _, f := range found {
		parts = append(parts, f.File+":"+strconv.Itoa(f.Line))
	}
	return strings.Join(parts, " ")
}

const scope = "schema = 1\n[scope]\nmode = \"whole\"\n[layout]\nexclude = [\"frozen/**\"]\n"

func TestCommitJudgesOnlyStagedAddedLines(t *testing.T) {
	r := newRepo(t)
	r.write("proofs.toml", scope)
	r.write("a.go", "package a\n\n// old comment stays\nfunc A() int { return 1 }\n")
	r.write("tool.py", "def f():\n    pass\n")
	r.write("frozen/x.go", "package x\n")
	r.commit("base")
	r.write("a.go", "package a\n\n// old comment stays\nfunc A() int { return 2 }\n\n// new comment\nfunc B() int { return 3 }\n")
	r.write("tool.py", "def f():\n    pass\n\n\ndef g():\n    \"\"\"Doc.\"\"\"\n    raise NotImplementedError\n")
	r.write("frozen/x.go", "package x\n\n// excluded\n")
	r.write("app.py", "from unittest import mock\n")
	r.write("app_test.py", "from unittest import mock\n")
	r.write("unstaged.go", "package a\n// not staged\n")
	r.git("add", "a.go", "tool.py", "frozen/x.go", "app.py", "app_test.py")
	got := r.run(check.Commit, "")
	if s := lines(got["code-comments"]); s != "a.go:6 tool.py:6" {
		t.Errorf("comments = %s", s)
	}
	if s := lines(got["code-placeholders"]); s != "tool.py:7" {
		t.Errorf("placeholders = %s", s)
	}
	if s := lines(got["code-mocks"]); s != "app.py:1" {
		t.Errorf("mocks = %s", s)
	}
}

func TestCIJudgesMergeBaseDiffAndIgnoresPureRenames(t *testing.T) {
	r := newRepo(t)
	r.write("proofs.toml", scope)
	r.write("docs/a.md", "# A\n\n<!-- kept -->\ntext\n")
	r.write("run.sh", "#!/bin/sh\n# kept\necho hi\n")
	r.commit("base")
	r.git("checkout", "-q", "-b", "topic")
	r.git("mv", "docs/a.md", "docs/b.md")
	r.write("run.sh", "#!/bin/sh\n# kept\necho hi # added\n")
	r.commit("topic")
	got := r.run(check.CI, "main")
	if s := lines(got["code-comments"]); s != "run.sh:3" {
		t.Errorf("comments = %s", s)
	}
}

func TestTouchedCommentLineCountsAsAdded(t *testing.T) {
	r := newRepo(t)
	r.write("proofs.toml", scope)
	r.write("a.go", "package a\n\n/* one\ntwo */\nvar X = 1\n")
	r.commit("base")
	r.write("a.go", "package a\n\n/* one\nthree */\nvar X = 1\n")
	r.git("add", "a.go")
	if s := lines(r.run(check.Commit, "")["code-comments"]); s != "a.go:4" {
		t.Errorf("comments = %s", s)
	}
}

func TestPythonParseFailureFailsClosed(t *testing.T) {
	r := newRepo(t)
	r.write("proofs.toml", scope)
	r.commit("base")
	r.write("bad.py", "def (:\n")
	r.git("add", "bad.py")
	got := r.run(check.Commit, "")
	for _, name := range []string{"code-comments", "code-placeholders", "code-mocks"} {
		if len(got[name]) != 1 || !strings.Contains(got[name][0].Problem, "cannot judge Python source") {
			t.Errorf("%s = %+v", name, got[name])
		}
	}
}

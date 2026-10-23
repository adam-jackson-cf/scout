package engine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// gitTree is a guardrails.Tree over one git revision.
type gitTree struct {
	ctx   context.Context
	root  string
	rev   string
	paths []string
}

func newGitTree(ctx context.Context, root, rev string) (*gitTree, error) {
	out, err := git(ctx, root, "ls-tree", "-r", "--name-only", "-z", rev)
	if err != nil {
		return nil, err
	}
	return &gitTree{ctx: ctx, root: root, rev: rev, paths: splitNUL(out)}, nil
}

func (t *gitTree) Paths() []string { return t.paths }

func (t *gitTree) Read(p string) ([]byte, bool) {
	out, err := git(t.ctx, t.root, "show", t.rev+":"+p)
	if err != nil || len(out) > maxGuardrailBytes {
		return nil, false
	}
	return out, true
}

// workTree is a guardrails.Tree over the working tree (tracked plus untracked,
// non-ignored files).
type workTree struct {
	root  string
	paths []string
}

func newWorkTree(ctx context.Context, root string) (*workTree, error) {
	out, err := git(ctx, root, "ls-files", "-co", "--exclude-standard", "-z")
	if err != nil {
		return nil, err
	}
	return &workTree{root: root, paths: splitNUL(out)}, nil
}

func (t *workTree) Paths() []string { return t.paths }

func (t *workTree) Read(p string) ([]byte, bool) {
	b, err := os.ReadFile(filepath.Join(t.root, filepath.FromSlash(p)))
	if err != nil || len(b) > maxGuardrailBytes {
		return nil, false
	}
	return b, true
}

func splitNUL(b []byte) []string {
	var out []string
	for _, s := range strings.Split(string(b), "\x00") {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

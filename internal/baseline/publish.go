// Package baseline publishes and restores Scout's exported ledger on the
// protected scout/state git branch (REQ-26) using git plumbing only, so no
// working tree is touched and no .sqlite/WAL file can ever be committed.
package baseline

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"scout/internal/state"
)

// DefaultBranch is the branch that carries the exported ledger.
const DefaultBranch = "scout/state"

// DefaultAttempts bounds the compare-and-swap retry loop.
const DefaultAttempts = 8

const stateDir = state.ExportDir + "/"

var (
	// ErrUntrusted means the caller did not assert a trusted run. PR and fork runs
	// are read-only and must never write the branch.
	ErrUntrusted = errors.New("publishing scout/state requires a trusted run (--trusted)")
	// ErrNoState means the remote has no scout/state branch yet.
	ErrNoState = errors.New("remote has no scout/state branch")
	// ErrContention means the branch kept moving until the retry budget ran out.
	ErrContention = errors.New("scout/state tip kept changing; retry budget exhausted")
)

// Tree maps repository-relative slash paths to file content.
type Tree map[string][]byte

// Parent is the actual fetched CAS tip and its validated state tree.
type Parent struct {
	Commit string
	Tree   Tree
}

// Options selects the repository, remote and branch to operate on.
type Options struct {
	// Repo is the git repository whose object store and remote configuration are used.
	Repo string
	// Remote is a remote name, path or URL.
	Remote string
	// Branch defaults to DefaultBranch.
	Branch string
	// Trusted must be true to publish.
	Trusted bool
	// MaxAttempts defaults to DefaultAttempts.
	MaxAttempts int
	// Now supplies the commit time; defaults to time.Now.
	Now func() time.Time
}

// Result describes a publish.
type Result struct {
	// Commit is the branch tip after the call (the new commit, or the unchanged tip).
	Commit string
	// Parent is the tip the commit was built on; empty for the first publish.
	Parent string
	// Attempts counts compare-and-swap attempts, including the successful one.
	Attempts int
	// Changed is false when the merged tree equalled the tip's tree and nothing was pushed.
	Changed bool
}

func (o Options) branch() string {
	if o.Branch == "" {
		return DefaultBranch
	}
	return o.Branch
}

func (o Options) ref() string { return "refs/heads/" + o.branch() }

// Publish applies apply to the current tip's tree and pushes the result as one
// child commit using an expected-parent lease. When another publisher moves the
// branch first, Publish re-reads the new tip, re-applies apply to it and tries
// again, up to MaxAttempts times. apply receives the actual fetched parent
// (empty when absent), and must revalidate any provider authorization on retry.
func Publish(ctx context.Context, opts Options, apply func(Parent) (Tree, error)) (Result, error) {
	attempts, now, err := publishLimits(opts)
	if err != nil {
		return Result{}, err
	}
	p := publisher{g: git{dir: opts.Repo}, opts: opts, apply: apply, now: now}
	for attempt := 1; attempt <= attempts; attempt++ {
		result, retry, err := p.attempt(ctx, attempt)
		if err != nil {
			return Result{}, err
		}
		if !retry {
			return result, nil
		}
	}
	return Result{}, ErrContention
}

func publishLimits(opts Options) (int, func() time.Time, error) {
	if !opts.Trusted {
		return 0, nil, ErrUntrusted
	}
	if opts.Repo == "" || opts.Remote == "" {
		return 0, nil, errors.New("repository and remote are required")
	}
	attempts := opts.MaxAttempts
	if attempts <= 0 {
		attempts = DefaultAttempts
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	return attempts, now, nil
}

type publisher struct {
	g     git
	opts  Options
	apply func(Parent) (Tree, error)
	now   func() time.Time
}

func (p publisher) attempt(ctx context.Context, attempt int) (Result, bool, error) {
	tip, next, err := p.nextTree(ctx)
	if err != nil {
		return Result{}, false, err
	}
	treeSHA, err := writeTree(ctx, p.g, next)
	if err != nil {
		return Result{}, false, err
	}
	unchanged, err := p.tipHasTree(ctx, tip, treeSHA)
	if err != nil {
		return Result{}, false, err
	}
	if unchanged {
		return Result{Commit: tip, Parent: tip, Attempts: attempt}, false, nil
	}
	commit, err := p.commit(ctx, tip, treeSHA)
	if err != nil {
		return Result{}, false, err
	}
	retry, err := p.push(ctx, tip, commit, attempt)
	if err != nil || retry {
		return Result{}, retry, err
	}
	return Result{Commit: commit, Parent: tip, Attempts: attempt, Changed: true}, false, nil
}

func (p publisher) nextTree(ctx context.Context) (string, Tree, error) {
	tip, current, err := fetchTip(ctx, p.g, p.opts)
	if err != nil && !errors.Is(err, ErrNoState) {
		return "", nil, err
	}
	if tip != "" {
		if _, err := state.ValidateState(current); err != nil {
			return "", nil, err
		}
	}
	next, err := p.apply(Parent{Commit: tip, Tree: current})
	if err != nil {
		return "", nil, err
	}
	if err := validatePublishTree(next); err != nil {
		return "", nil, err
	}
	return tip, next, nil
}

func (p publisher) tipHasTree(ctx context.Context, tip, treeSHA string) (bool, error) {
	if tip == "" {
		return false, nil
	}
	tipTree, err := p.g.output(ctx, nil, "rev-parse", tip+"^{tree}")
	if err != nil {
		return false, err
	}
	return tipTree == treeSHA, nil
}

func (p publisher) commit(ctx context.Context, tip, treeSHA string) (string, error) {
	stamp := strconv.FormatInt(p.now().Unix(), 10) + " +0000"
	env := []string{
		"GIT_AUTHOR_NAME=scout", "GIT_AUTHOR_EMAIL=scout@users.noreply.github.com", "GIT_AUTHOR_DATE=" + stamp,
		"GIT_COMMITTER_NAME=scout", "GIT_COMMITTER_EMAIL=scout@users.noreply.github.com", "GIT_COMMITTER_DATE=" + stamp,
	}
	commitArgs := []string{"-c", "commit.gpgsign=false", "commit-tree", treeSHA, "-m", "scout: publish state"}
	if tip != "" {
		commitArgs = append(commitArgs, "-p", tip)
	}
	return p.g.output(ctx, env, commitArgs...)
}

func (p publisher) push(ctx context.Context, tip, commit string, attempt int) (bool, error) {
	lease := "--force-with-lease=" + p.opts.ref() + ":" + tip
	if _, err := p.g.output(ctx, nil, "push", "--quiet", lease, p.opts.Remote, commit+":"+p.opts.ref()); err != nil {
		// A rejected lease means someone else moved the tip; anything else is real.
		actual, lookupErr := remoteTip(ctx, p.g, p.opts)
		if lookupErr == nil && actual != tip {
			if err := backoff(ctx, attempt); err != nil {
				return false, err
			}
			return true, nil
		}
		return false, fmt.Errorf("push %s: %w", p.opts.ref(), err)
	}
	return false, nil
}

// Fetch reads the remote branch tip and returns its state/v1 files without
// writing anything to the remote or the working tree. It returns ErrNoState
// when the branch does not exist.
func Fetch(ctx context.Context, opts Options) (Tree, string, error) {
	if opts.Repo == "" || opts.Remote == "" {
		return nil, "", errors.New("repository and remote are required")
	}
	tip, tree, err := fetchTip(ctx, git{dir: opts.Repo}, opts)
	return tree, tip, err
}

func backoff(ctx context.Context, attempt int) error {
	var jitter [1]byte
	_, _ = rand.Read(jitter[:])
	wait := time.Duration(attempt*10+int(jitter[0])%40) * time.Millisecond
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func validatePublishTree(tree Tree) error {
	for path := range tree {
		lower := strings.ToLower(path)
		if !strings.HasPrefix(path, stateDir) || strings.Contains(path, "..") || strings.Contains(path, "//") {
			return fmt.Errorf("refusing to publish %q: only %s* is allowed", path, stateDir)
		}
		for _, suffix := range []string{".sqlite", ".sqlite3", ".db", "-wal", "-shm", "-journal", ".lock"} {
			if strings.HasSuffix(lower, suffix) {
				return fmt.Errorf("refusing to publish database file %q", path)
			}
		}
	}
	return nil
}

var fetchCounter atomic.Uint64

func remoteTip(ctx context.Context, g git, opts Options) (string, error) {
	out, err := g.output(ctx, nil, "ls-remote", opts.Remote, opts.ref())
	if err != nil {
		return "", fmt.Errorf("list remote: %w", err)
	}
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == opts.ref() {
			return fields[0], nil
		}
	}
	return "", nil
}

// fetchTip returns the remote tip and its state/v1 files (empty tree and
// ErrNoState when the branch is absent).
func fetchTip(ctx context.Context, g git, opts Options) (string, Tree, error) {
	tip, err := remoteTip(ctx, g, opts)
	if err != nil {
		return "", Tree{}, err
	}
	if tip == "" {
		return "", Tree{}, ErrNoState
	}
	var salt [4]byte
	_, _ = rand.Read(salt[:])
	local := fmt.Sprintf("refs/scout/fetch/%d-%d-%s", os.Getpid(), fetchCounter.Add(1), hex.EncodeToString(salt[:]))
	if _, err := g.output(ctx, nil, "fetch", "--quiet", "--no-tags", "--no-write-fetch-head", opts.Remote, "+"+opts.ref()+":"+local); err != nil {
		return "", Tree{}, fmt.Errorf("fetch %s: %w", opts.ref(), err)
	}
	defer func() { _, _ = g.output(context.Background(), nil, "update-ref", "-d", local) }()
	fetched, err := g.output(ctx, nil, "rev-parse", local)
	if err != nil {
		return "", Tree{}, err
	}
	tree, err := readTree(ctx, g, fetched)
	if err != nil {
		return "", Tree{}, err
	}
	return fetched, tree, nil
}

func readTree(ctx context.Context, g git, commit string) (Tree, error) {
	listing, err := g.raw(ctx, nil, nil, "ls-tree", "-r", "-z", commit, "--", state.ExportDir)
	if err != nil {
		return nil, err
	}
	tree := Tree{}
	for _, entry := range bytes.Split(listing, []byte{0}) {
		if len(entry) == 0 {
			continue
		}
		meta, path, ok := strings.Cut(string(entry), "\t")
		fields := strings.Fields(meta)
		if !ok || len(fields) != 3 || fields[1] != "blob" {
			continue
		}
		data, err := g.raw(ctx, nil, nil, "cat-file", "blob", fields[2])
		if err != nil {
			return nil, err
		}
		tree[path] = data
	}
	return tree, nil
}

type node struct {
	blobs map[string]string
	dirs  map[string]*node
}

func newNode() *node { return &node{blobs: map[string]string{}, dirs: map[string]*node{}} }

func writeTree(ctx context.Context, g git, files Tree) (string, error) {
	root := newNode()
	for path, data := range files {
		sha, err := g.rawString(ctx, nil, data, "hash-object", "-w", "--stdin")
		if err != nil {
			return "", err
		}
		parts := strings.Split(path, "/")
		current := root
		for _, part := range parts[:len(parts)-1] {
			child := current.dirs[part]
			if child == nil {
				child = newNode()
				current.dirs[part] = child
			}
			current = child
		}
		current.blobs[parts[len(parts)-1]] = sha
	}
	return mktree(ctx, g, root)
}

func mktree(ctx context.Context, g git, n *node) (string, error) {
	var lines []string
	for name, sha := range n.blobs {
		lines = append(lines, "100644 blob "+sha+"\t"+name)
	}
	for name, child := range n.dirs {
		sha, err := mktree(ctx, g, child)
		if err != nil {
			return "", err
		}
		lines = append(lines, "040000 tree "+sha+"\t"+name)
	}
	sort.Strings(lines)
	input := strings.Join(lines, "\n")
	if input != "" {
		input += "\n"
	}
	return g.rawString(ctx, nil, []byte(input), "mktree")
}

type git struct{ dir string }

func (g git) output(ctx context.Context, env []string, args ...string) (string, error) {
	return g.rawString(ctx, env, nil, args...)
}

func (g git) rawString(ctx context.Context, env []string, stdin []byte, args ...string) (string, error) {
	out, err := g.raw(ctx, env, stdin, args...)
	return strings.TrimSpace(string(out)), err
}

func (g git) raw(ctx context.Context, env []string, stdin []byte, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, "git", args...)
	command.Dir = g.dir
	command.Env = append(os.Environ(), append([]string{"GIT_TERMINAL_PROMPT=0", "GIT_OPTIONAL_LOCKS=0"}, env...)...)
	command.Stdin = bytes.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

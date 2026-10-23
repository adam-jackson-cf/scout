// Package history computes the §8.3 exposure and interest measures from git
// data. Everything here is a pure function of the commit list (and optional PR
// metadata / per-file complexity) except Load and the rework blame step, which
// shell out to git with a fixed argument vector.
package history

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

// FileChange is one numstat row of a commit (renames are not followed).
type FileChange struct {
	Path    string
	Added   int
	Removed int
	Binary  bool
}

// Commit is one commit with author identity and numstat.
type Commit struct {
	Hash        string
	AuthorName  string
	AuthorEmail string
	Time        time.Time // author timestamp, UTC
	Parents     int
	Files       []FileChange
}

const (
	recSep  = "\x1e"
	fieldSp = "\x1f"
)

func runGit(ctx context.Context, repo string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", repo, "-c", "core.quotepath=off"}, args...)...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(errb.String()))
	}
	return out.Bytes(), nil
}

// GitVersion returns `git --version` output, used in the tool digest.
func GitVersion(ctx context.Context) string {
	out, err := exec.CommandContext(ctx, "git", "--version").Output()
	if err != nil {
		return "git-unknown"
	}
	return strings.TrimSpace(string(out))
}

// Head returns the full hash of HEAD in repo.
func Head(ctx context.Context, repo string) (string, error) {
	out, err := runGit(ctx, repo, "rev-parse", "HEAD")
	return strings.TrimSpace(string(out)), err
}

// Load reads every commit reachable from HEAD with numstat, oldest first
// (author time, then hash). Rename detection is off so a path is a path.
func Load(ctx context.Context, repo string) ([]Commit, error) {
	out, err := runGit(ctx, repo, "log", "--no-renames", "--numstat",
		"--format="+recSep+"%H"+fieldSp+"%an"+fieldSp+"%ae"+fieldSp+"%at"+fieldSp+"%P")
	if err != nil {
		return nil, err
	}
	return parseLog(string(out))
}

func parseLog(s string) ([]Commit, error) {
	var commits []Commit
	for _, rec := range strings.Split(s, recSep) {
		if strings.TrimSpace(rec) == "" {
			continue
		}
		nl := strings.IndexByte(rec, '\n')
		head, body := rec, ""
		if nl >= 0 {
			head, body = rec[:nl], rec[nl+1:]
		}
		f := strings.Split(head, fieldSp)
		if len(f) != 5 {
			return nil, fmt.Errorf("history: malformed log header %q", head)
		}
		at, err := strconv.ParseInt(f[3], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("history: author time %q: %w", f[3], err)
		}
		c := Commit{Hash: f[0], AuthorName: f[1], AuthorEmail: f[2], Time: time.Unix(at, 0).UTC(), Parents: len(strings.Fields(f[4]))}
		for _, line := range strings.Split(body, "\n") {
			if line == "" {
				continue
			}
			p := strings.SplitN(line, "\t", 3)
			if len(p) != 3 {
				return nil, fmt.Errorf("history: malformed numstat %q", line)
			}
			fc := FileChange{Path: p[2]}
			if p[0] == "-" || p[1] == "-" {
				fc.Binary = true
			} else {
				fc.Added, _ = strconv.Atoi(p[0])
				fc.Removed, _ = strconv.Atoi(p[1])
			}
			c.Files = append(c.Files, fc)
		}
		commits = append(commits, c)
	}
	// git log is newest-first; present oldest-first, deterministic.
	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}
	sortCommits(commits)
	return commits, nil
}

func sortCommits(cs []Commit) {
	sort.SliceStable(cs, func(i, j int) bool { return lessCommit(cs[i], cs[j]) })
}

func lessCommit(a, b Commit) bool {
	if !a.Time.Equal(b.Time) {
		return a.Time.Before(b.Time)
	}
	return a.Hash < b.Hash
}

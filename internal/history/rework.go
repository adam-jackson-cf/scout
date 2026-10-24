package history

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"scout/internal/domain"
)

// removedRange is a run of lines removed from the parent's version of a file.
type removedRange struct{ start, count int }

var hunkRe = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+`)

// parseRemoved extracts removed-line ranges per path from `diff -U0` output.
func parseRemoved(diff string) map[string][]removedRange {
	out := map[string][]removedRange{}
	cur := ""
	for _, l := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(l, "--- "):
			if p, ok := strings.CutPrefix(l, "--- a/"); ok {
				cur = p
			} else {
				cur = "" // /dev/null: new file, nothing removed
			}
		case strings.HasPrefix(l, "@@") && cur != "":
			m := hunkRe.FindStringSubmatch(l)
			if m == nil {
				continue
			}
			start, _ := strconv.Atoi(m[1])
			n := 1
			if m[2] != "" {
				n, _ = strconv.Atoi(m[2])
			}
			if n > 0 {
				out[cur] = append(out[cur], removedRange{start, n})
			}
		}
	}
	return out
}

// blameOrigins returns, for a file at `parent`, the 39-char origin prefix of
// each line in the given ranges (`git blame -l -s -L …`).
func blameOrigins(ctx context.Context, repo, parent, path string, rs []removedRange) ([]string, error) {
	args := []string{"blame", "-l", "-s"}
	for _, r := range rs {
		args = append(args, "-L", fmt.Sprintf("%d,%d", r.start, r.start+r.count-1))
	}
	args = append(args, parent, "--", path)
	out, err := runGit(ctx, repo, args...)
	if err != nil {
		return nil, err
	}
	var origins []string
	for _, l := range strings.Split(string(out), "\n") {
		if l == "" {
			continue
		}
		h := strings.TrimPrefix(strings.Fields(l)[0], "^")
		if len(h) > 39 {
			h = h[:39]
		}
		origins = append(origins, h)
	}
	return origins, nil
}

type removedLine struct {
	path   string
	origin string
}

// Rework computes scout.rework@1 per scope and fixed window: removed lines
// (in eligible scope files of the window's changes) whose blame origin, at the
// change's first parent, is itself a change in the same window, ÷ all removed
// lines. Windows with no removed lines are `unmeasured`, not 0. Trailing view
// included. Blame failures are returned as errors, never silently dropped.
func Rework(ctx context.Context, repo string, commits []Commit, o Options) ([]domain.Signal, error) {
	changes, _ := buildChanges(commits, o)
	if len(commits) == 0 {
		return nil, nil
	}
	end := o.End
	if end.IsZero() {
		end = commits[len(commits)-1].Time
	}
	fixed, trailing := Windows(commits[0].Time, end, o.window())
	wins := append(append([]Window(nil), fixed...), trailing)

	cache := map[string][]removedLine{}
	removed := func(hash string) ([]removedLine, error) {
		if r, ok := cache[hash]; ok {
			return r, nil
		}
		diff, err := runGit(ctx, repo, "diff", "-U0", "--no-renames", hash+"^", hash)
		if err != nil {
			return nil, err
		}
		var res []removedLine
		byPath := parseRemoved(string(diff))
		paths := make([]string, 0, len(byPath))
		for p := range byPath {
			paths = append(paths, p)
		}
		sort.Strings(paths)
		for _, p := range paths {
			if !o.eligible(p) {
				continue
			}
			origins, err := blameOrigins(ctx, repo, hash+"^", p, byPath[p])
			if err != nil {
				return nil, err
			}
			for _, og := range origins {
				res = append(res, removedLine{p, og})
			}
		}
		cache[hash] = res
		return res, nil
	}

	parents := map[string]int{}
	for _, c := range commits {
		parents[c.Hash] = c.Parents
	}
	var sigs []domain.Signal
	for _, sc := range o.Scopes {
		for _, w := range wins {
			inWindow := map[string]bool{} // 39-char prefixes of repo-wide changes in the window
			for _, c := range changes {
				if w.has(c.t) {
					inWindow[c.hash[:39]] = true
				}
			}
			sw := newScopeWindow(sc, w, changes, o, end)
			sig := sw.base("scout.rework@1")
			num, den := 0, 0
			for _, it := range sw.items {
				if parents[it.c.hash] == 0 {
					continue // root commit has no parent to blame
				}
				rl, err := removed(it.c.hash)
				if err != nil {
					return nil, err
				}
				for _, l := range rl {
					if !sc.Contains(l.path) {
						continue
					}
					den++
					if inWindow[l.origin] {
						num++
					}
				}
			}
			if den == 0 {
				sigs = append(sigs, markMissing(sig, "unmeasured", "no removed lines in scope in window"))
				continue
			}
			sig = setValue(sig, float64(num)/float64(den), float64(num), float64(den))
			if len(sw.items) < MinChanges {
				sig = markMissing(sig, "insufficient", fmt.Sprintf("%d changes < %d", len(sw.items), MinChanges))
			}
			sigs = append(sigs, sig)
		}
	}
	return sigs, nil
}

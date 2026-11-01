package association

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Repo reads the first-parent line of a git repository.
type Repo struct{ Dir string }

// Merge is one first-parent commit of a push (a merge commit or a squash or
// rebase merge) with its changes against its first parent.
type Merge struct {
	SHA, Parent, Subject string
	Changes              []Change
}

// renameMinSimilarity is the pilot's -M50% rename detection.
const renameMinSimilarity = 50

func (r Repo) git(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", r.Dir, "-c", "core.quotepath=off"}, args...)...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(errb.String()))
	}
	return out.Bytes(), nil
}

// Merges lists the first-parent commits of base..head, oldest first, with their
// changes. Root commits are skipped (no parent to diff against). Rename
// detection is git -M50%.
func (r Repo) Merges(ctx context.Context, base, head string) ([]Merge, error) {
	out, err := r.git(ctx, "rev-list", "--first-parent", "--reverse", "--format=%H%x1f%P%x1f%s", "--no-commit-header", base+".."+head)
	if err != nil {
		return nil, err
	}
	empty, err := r.emptyBlob(ctx)
	if err != nil {
		return nil, err
	}
	var merges []Merge
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		f := strings.SplitN(line, "\x1f", 3)
		if len(f) != 3 || f[1] == "" {
			continue
		}
		m := Merge{SHA: f[0], Parent: strings.Fields(f[1])[0], Subject: f[2]}
		if m.Changes, err = r.changes(ctx, m.Parent, m.SHA, empty); err != nil {
			return nil, err
		}
		merges = append(merges, m)
	}
	return merges, nil
}

func (r Repo) emptyBlob(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", r.Dir, "hash-object", "-t", "blob", "--stdin")
	cmd.Stdin = strings.NewReader("")
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

// changes parses `git diff-tree -r -z --raw -M50%`. Type changes count as
// modified, copies as added.
func (r Repo) changes(ctx context.Context, parent, sha, emptyBlob string) ([]Change, error) {
	raw, err := r.git(ctx, "diff-tree", "-r", "-z", "--raw", "--no-abbrev", fmt.Sprintf("-M%d%%", renameMinSimilarity), parent, sha)
	if err != nil {
		return nil, err
	}
	tok := strings.Split(string(raw), "\x00")
	var out []Change
	for i := 0; i < len(tok); i++ {
		if !strings.HasPrefix(tok[i], ":") {
			continue
		}
		meta := strings.Fields(tok[i][1:]) // oldmode newmode oldblob newblob status
		if len(meta) < 5 {
			return nil, fmt.Errorf("git diff-tree: malformed record %q", tok[i])
		}
		newBlob, status := meta[3], meta[4]
		switch status[0] {
		case 'R', 'C':
			if i+2 >= len(tok) {
				return nil, fmt.Errorf("git diff-tree: truncated record %q", tok[i])
			}
			c := Change{Kind: Renamed, OldPath: tok[i+1], Path: tok[i+2], Empty: newBlob == emptyBlob}
			if status[0] == 'C' {
				c = Change{Kind: Added, Path: tok[i+2], Empty: newBlob == emptyBlob}
			}
			out = append(out, c)
			i += 2
		case 'D':
			out = append(out, Change{Kind: Deleted, Path: tok[i+1]})
			i++
		case 'A':
			out = append(out, Change{Kind: Added, Path: tok[i+1], Empty: newBlob == emptyBlob})
			i++
		default: // M, T
			out = append(out, Change{Kind: Modified, Path: tok[i+1]})
			i++
		}
	}
	return out, nil
}

// TreePaths returns every file path of the tree at rev.
func (r Repo) TreePaths(ctx context.Context, rev string) (map[string]struct{}, error) {
	out, err := r.git(ctx, "ls-tree", "-r", "--name-only", "-z", rev)
	if err != nil {
		return nil, err
	}
	set := map[string]struct{}{}
	for _, p := range strings.Split(string(out), "\x00") {
		if p != "" {
			set[p] = struct{}{}
		}
	}
	return set, nil
}

// File returns the content of path at rev (the merge revision).
func (r Repo) File(ctx context.Context, rev, path string) ([]byte, error) {
	return r.git(ctx, "show", rev+":"+path)
}

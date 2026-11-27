package repofile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"scout/tools/gate/check"
)

type File struct {
	Path  string
	Data  []byte
	Added map[int]bool
}

func (f File) IsAdded(line int) bool {
	return f.Added == nil || f.Added[line]
}

func (f File) AnyAdded(from, to int) (int, bool) {
	for l := from; l <= to; l++ {
		if f.IsAdded(l) {
			return l, true
		}
	}
	return 0, false
}

type layoutTable struct {
	Exclude []string `toml:"exclude"`
}

func Exclusions(r *check.Run) ([]string, error) {
	if r.Proofs == nil {
		return nil, nil
	}
	var l layoutTable
	if _, err := r.Proofs.Section("layout", &l); err != nil {
		return nil, fmt.Errorf("proofs.toml [layout]: %w", err)
	}
	for _, p := range l.Exclude {
		if strings.TrimSpace(p) == "" {
			return nil, errors.New("proofs.toml [layout].exclude holds an empty pattern")
		}
	}
	return l.Exclude, nil
}

func Match(pattern, name string) bool {
	return matchSegments(strings.Split(pattern, "/"), strings.Split(name, "/"))
}

func matchSegments(pat, name []string) bool {
	for len(pat) > 0 {
		if pat[0] == "**" {
			rest := pat[1:]
			for i := 0; i <= len(name); i++ {
				if matchSegments(rest, name[i:]) {
					return true
				}
			}
			return false
		}
		if len(name) == 0 {
			return false
		}
		if ok, err := path.Match(pat[0], name[0]); err != nil || !ok {
			return false
		}
		pat, name = pat[1:], name[1:]
	}
	return len(name) == 0
}

func excluded(patterns []string, name string) bool {
	for _, p := range patterns {
		if Match(p, name) {
			return true
		}
	}
	return false
}

func Selected(r *check.Run, want func(string) bool) ([]File, error) {
	if r.Explicit {
		return explicitFiles(r, want)
	}
	names, err := r.Tracked()
	if err != nil {
		return nil, err
	}
	names, err = inScope(r, names)
	if err != nil {
		return nil, err
	}
	return repoFiles(r, names, want, nil)
}

func Changed(r *check.Run, want func(string) bool) ([]File, error) {
	if r.Explicit {
		return explicitFiles(r, want)
	}
	changes, err := r.Diff()
	if err != nil {
		return nil, err
	}
	var names []string
	for _, c := range changes {
		if c.Status != "D" {
			names = append(names, c.Path)
		}
	}
	if len(names) == 0 {
		return nil, nil
	}
	added, err := addedLines(r)
	if err != nil {
		return nil, err
	}
	return repoFiles(r, names, want, added)
}

func inScope(r *check.Run, names []string) ([]string, error) {
	if r.Proofs == nil {
		return names, nil
	}
	switch r.Proofs.Scope.Mode {
	case "", "whole":
		return names, nil
	case "paths":
		return underPaths(names, r.Proofs.Scope.Paths), nil
	case "changed":
		return changedNames(r, names)
	}
	return nil, fmt.Errorf("proofs.toml [scope].mode %q is not whole, changed or paths", r.Proofs.Scope.Mode)
}

func underPaths(names, paths []string) []string {
	var out []string
	for _, n := range names {
		for _, p := range paths {
			p = strings.TrimSuffix(p, "/")
			if n == p || strings.HasPrefix(n, p+"/") {
				out = append(out, n)
				break
			}
		}
	}
	return out
}

func changedNames(r *check.Run, names []string) ([]string, error) {
	changes, err := r.Diff()
	if err != nil {
		return nil, err
	}
	keep := map[string]bool{}
	for _, c := range changes {
		keep[c.Path] = c.Status != "D"
	}
	var out []string
	for _, n := range names {
		if keep[n] {
			out = append(out, n)
		}
	}
	return out, nil
}

func explicitFiles(r *check.Run, want func(string) bool) ([]File, error) {
	var out []File
	seen := map[string]bool{}
	for _, p := range r.Paths {
		rel, err := relative(r.Root, p)
		if err != nil {
			return nil, err
		}
		err = filepath.WalkDir(filepath.Join(r.Root, rel), explicitVisitor(r.Root, want, seen, &out))
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

func explicitVisitor(root string, want func(string) bool, seen map[string]bool, out *[]File) fs.WalkDirFunc {
	return func(abs string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		if d.IsDir() {
			return nil
		}
		name, err := relative(root, abs)
		if err != nil {
			return err
		}
		if seen[name] || !want(name) {
			return nil
		}
		seen[name] = true
		if d.Type()&fs.ModeSymlink != 0 {
			return symlinkInside(root, abs, name)
		}
		data, err := os.ReadFile(abs)
		if err != nil {
			return err
		}
		*out = append(*out, File{Path: name, Data: data})
		return nil
	}
}

func relative(root, p string) (string, error) {
	abs := p
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(root, p)
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, "../") {
		return "", fmt.Errorf("path %s is outside the repository %s", p, root)
	}
	return filepath.ToSlash(rel), nil
}

func symlinkInside(root, abs, name string) error {
	target, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return fmt.Errorf("%s: unresolvable symlink: %w", name, err)
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	if _, err := relative(realRoot, target); err != nil {
		return fmt.Errorf("%s: symlink resolves outside the repository; replace it with the owned file", name)
	}
	return nil
}

func repoFiles(r *check.Run, names []string, want func(string) bool, added map[string]map[int]bool) ([]File, error) {
	keep, err := keptNames(r, names, want)
	if err != nil {
		return nil, err
	}
	data, err := blobs(r, keep)
	if err != nil {
		return nil, err
	}
	out := make([]File, 0, len(keep))
	for _, n := range keep {
		f := File{Path: n, Data: data[n]}
		if added != nil {
			f.Added = added[n]
			if f.Added == nil {
				f.Added = map[int]bool{}
			}
		}
		out = append(out, f)
	}
	return out, nil
}

func keptNames(r *check.Run, names []string, want func(string) bool) ([]string, error) {
	ex, err := Exclusions(r)
	if err != nil {
		return nil, err
	}
	modes, err := entryModes(r)
	if err != nil {
		return nil, err
	}
	var keep []string
	for _, n := range names {
		if excluded(ex, n) || !want(n) {
			continue
		}
		switch modes[n] {
		case "160000":
			continue
		case "120000":
			if err := linkInside(r, n); err != nil {
				return nil, err
			}
			continue
		}
		keep = append(keep, n)
	}
	sort.Strings(keep)
	return keep, nil
}

func revision(r *check.Run) string {
	if r.Mode == check.Commit {
		return ""
	}
	return "HEAD"
}

func entryModes(r *check.Run) (map[string]string, error) {
	var raw string
	var err error
	if r.Mode == check.Commit {
		raw, err = r.Git("ls-files", "-s", "-z")
	} else {
		raw, err = r.Git("ls-tree", "-r", "-z", "HEAD")
	}
	if err != nil {
		return nil, err
	}
	modes := map[string]string{}
	for _, rec := range strings.Split(raw, "\x00") {
		meta, name, ok := strings.Cut(rec, "\t")
		if !ok {
			continue
		}
		mode, _, _ := strings.Cut(meta, " ")
		modes[name] = mode
	}
	return modes, nil
}

func linkInside(r *check.Run, name string) error {
	data, err := blobs(r, []string{name})
	if err != nil {
		return err
	}
	target := path.Join(path.Dir(name), string(data[name]))
	if path.IsAbs(string(data[name])) || target == ".." || strings.HasPrefix(target, "../") {
		return fmt.Errorf("%s: tracked symlink resolves outside the repository; replace it with the owned file", name)
	}
	return nil
}

func blobs(r *check.Run, names []string) (map[string][]byte, error) {
	out := map[string][]byte{}
	if len(names) == 0 {
		return out, nil
	}
	rev := revision(r)
	buf, err := catFiles(r, rev, names)
	if err != nil {
		return nil, err
	}
	return parseBatch(buf, names, rev, out)
}

func catFiles(r *check.Run, rev string, names []string) ([]byte, error) {
	var in bytes.Buffer
	for _, n := range names {
		fmt.Fprintf(&in, "%s:%s\n", rev, n)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "cat-file", "--batch")
	cmd.Dir = r.Root
	cmd.Env = r.GitEnv()
	cmd.Stdin = &in
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git cat-file --batch: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

func parseBatch(buf []byte, names []string, rev string, out map[string][]byte) (map[string][]byte, error) {
	for _, n := range names {
		nl := bytes.IndexByte(buf, '\n')
		if nl < 0 {
			return nil, errors.New("git cat-file --batch: truncated output")
		}
		header := strings.Fields(string(buf[:nl]))
		buf = buf[nl+1:]
		if len(header) != 3 {
			return nil, fmt.Errorf("%s: not readable from %s; stage or commit the file", n, describe(rev))
		}
		size, err := strconv.Atoi(header[2])
		if err != nil || size > len(buf) {
			return nil, fmt.Errorf("git cat-file --batch: bad header for %s", n)
		}
		out[n] = append([]byte(nil), buf[:size]...)
		buf = buf[size:]
		if len(buf) > 0 && buf[0] == '\n' {
			buf = buf[1:]
		}
	}
	return out, nil
}

func describe(rev string) string {
	if rev == "" {
		return "the index"
	}
	return rev
}

func addedLines(r *check.Run) (map[string]map[int]bool, error) {
	args := []string{"-c", "core.quotePath=false", "diff", "-U0", "--no-color", "--no-ext-diff", "--find-renames", "--src-prefix=a/", "--dst-prefix=b/"}
	if r.Mode == check.Commit {
		args = append(args, "--cached")
	} else {
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
	return ParsePatch(raw), nil
}

func ParsePatch(raw string) map[string]map[int]bool {
	out := map[string]map[int]bool{}
	text := map[string]map[int]string{}
	removed := map[string]int{}
	var current map[int]bool
	var currentText map[int]string
	next := 0
	for _, line := range strings.Split(raw, "\n") {
		switch {
		case strings.HasPrefix(line, "+++ "):
			name := strings.TrimPrefix(line, "+++ ")
			if unq, err := strconv.Unquote(name); err == nil {
				name = unq
			}
			current, currentText = nil, nil
			if strings.HasPrefix(name, "b/") {
				name = strings.TrimPrefix(name, "b/")
				current, currentText = map[int]bool{}, map[int]string{}
				out[name], text[name] = current, currentText
			}
		case strings.HasPrefix(line, "--- "):
		case strings.HasPrefix(line, "@@ "):
			next, _ = hunkRange(line)
		case strings.HasPrefix(line, "+") && current != nil:
			current[next] = true
			currentText[next] = strings.TrimSpace(line[1:])
			next++
		case strings.HasPrefix(line, "-"):
			removed[strings.TrimSpace(line[1:])]++
		}
	}
	dropMoved(out, text, removed)
	return out
}

func dropMoved(out map[string]map[int]bool, text map[string]map[int]string, removed map[string]int) {
	names := make([]string, 0, len(out))
	for n := range out {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		lines := make([]int, 0, len(out[n]))
		for l := range out[n] {
			lines = append(lines, l)
		}
		sort.Ints(lines)
		for _, l := range lines {
			t := text[n][l]
			if t != "" && removed[t] > 0 {
				removed[t]--
				delete(out[n], l)
			}
		}
	}
}

func hunkRange(header string) (int, int) {
	fields := strings.Fields(header)
	if len(fields) < 3 || !strings.HasPrefix(fields[2], "+") {
		return 0, 0
	}
	spec := strings.TrimPrefix(fields[2], "+")
	startText, countText, hasCount := strings.Cut(spec, ",")
	start, err := strconv.Atoi(startText)
	if err != nil {
		return 0, 0
	}
	count := 1
	if hasCount {
		if count, err = strconv.Atoi(countText); err != nil {
			return 0, 0
		}
	}
	return start, count
}

func unpack(r *check.Run, scripts fs.FS) (string, error) {
	dir, err := os.MkdirTemp(r.TempDir, "python-")
	if err != nil {
		return "", err
	}
	names, err := fs.Glob(scripts, "*.py")
	if err != nil {
		return "", err
	}
	for _, name := range names {
		data, err := fs.ReadFile(scripts, name)
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
			return "", err
		}
	}
	return dir, nil
}

func RunPython(r *check.Run, scripts fs.FS, script string, request, response any) error {
	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}
	dir, err := unpack(r, scripts)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "python3", "-B", "-E", "-s", filepath.Join(dir, script))
	cmd.Dir = r.TempDir
	cmd.Env = append(os.Environ(), "PYTHONDONTWRITEBYTECODE=1", "PYTHONPYCACHEPREFIX="+filepath.Join(r.TempDir, "pycache"))
	cmd.Stdin = bytes.NewReader(payload)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("python3 %s: %w: %s", script, err, strings.TrimSpace(stderr.String()))
	}
	if err := json.Unmarshal(stdout.Bytes(), response); err != nil {
		return fmt.Errorf("python3 %s: unreadable response: %w", script, err)
	}
	return nil
}

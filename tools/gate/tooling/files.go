package tooling

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"scout/tools/gate/check"
	"scout/tools/gate/pathscope"
)

type layout struct {
	Exclude []string `toml:"exclude"`
}

func readLayout(r *check.Run) (layout, error) {
	var l layout
	if r.Proofs == nil {
		return l, nil
	}
	if _, err := r.Proofs.Section("layout", &l); err != nil {
		return l, fmt.Errorf("proofs.toml [layout]: %w", err)
	}
	return l, nil
}

func (l layout) routineExcluded(p string) bool {
	for _, pattern := range l.Exclude {
		if pathscope.Match(pattern, p) {
			return true
		}
	}
	return false
}

type selector func(root, rel string) bool

func byExt(exts ...string) selector {
	return func(_, rel string) bool {
		ext := path.Ext(rel)
		for _, e := range exts {
			if ext == e {
				return true
			}
		}
		return false
	}
}

func shellScript(root, rel string) bool {
	switch path.Ext(rel) {
	case ".sh", ".bash":
		return true
	case "":
		return hasShellShebang(filepath.Join(root, rel))
	}
	return false
}

func hasShellShebang(file string) bool {
	data, err := os.ReadFile(file)
	if err != nil {
		return false
	}
	line, _, _ := strings.Cut(string(data), "\n")
	if !strings.HasPrefix(line, "#!") {
		return false
	}
	fields := strings.Fields(strings.TrimPrefix(line, "#!"))
	if len(fields) == 0 {
		return false
	}
	interp := path.Base(fields[0])
	if interp == "env" && len(fields) > 1 {
		interp = fields[len(fields)-1]
	}
	switch interp {
	case "sh", "bash", "dash", "ksh":
		return true
	}
	return false
}

func selectFiles(r *check.Run, want selector) ([]string, error) {
	if r.Explicit {
		return explicitFiles(r.Root, r.Paths, want)
	}
	l, err := readLayout(r)
	if err != nil {
		return nil, err
	}
	raw, err := r.Git("ls-files", "-z", "--cached", "--stage")
	if err != nil {
		return nil, err
	}
	var out []string
	for _, entry := range strings.Split(raw, "\x00") {
		if rel, ok := routineFile(r.Root, entry, l, want); ok {
			out = append(out, rel)
		}
	}
	sort.Strings(out)
	return out, nil
}

func routineFile(root, entry string, l layout, want selector) (string, bool) {
	meta, rel, ok := strings.Cut(entry, "\t")
	if !ok || strings.HasPrefix(meta, "120000 ") || strings.HasPrefix(meta, "160000 ") {
		return "", false
	}
	if l.routineExcluded(rel) || !want(root, rel) {
		return "", false
	}
	_, err := os.Lstat(filepath.Join(root, rel))
	return rel, err == nil
}

func explicitFiles(root string, paths []string, want selector) ([]string, error) {
	var out []string
	for _, p := range paths {
		found, err := explicitPath(root, filepath.ToSlash(p), want)
		if err != nil {
			return nil, err
		}
		out = append(out, found...)
	}
	sort.Strings(out)
	return out, nil
}

func explicitPath(root, p string, want selector) ([]string, error) {
	abs := filepath.Join(root, p)
	info, err := os.Lstat(abs)
	if err != nil {
		return nil, fmt.Errorf("--path %s: %w", p, err)
	}
	if !info.IsDir() {
		if want(root, p) {
			return []string{p}, nil
		}
		return nil, nil
	}
	var out []string
	err = filepath.WalkDir(abs, func(file string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Type()&fs.ModeSymlink != 0 {
			return err
		}
		rel, err := filepath.Rel(root, file)
		if err == nil && want(root, filepath.ToSlash(rel)) {
			out = append(out, filepath.ToSlash(rel))
		}
		return err
	})
	return out, err
}

func toolsDir(root string) string {
	if d := os.Getenv("SCOUT_TOOLS_DIR"); d != "" {
		return d
	}
	return filepath.Join(root, ".tools")
}

func toolPath(root, name string) (string, error) {
	p := filepath.Join(toolsDir(root), "bin", name)
	if _, err := os.Stat(p); err != nil {
		return "", fmt.Errorf("%s is not installed at %s; run scripts/bootstrap-tools.sh and source .tools/activate", name, p)
	}
	return p, nil
}

type result struct {
	stdout   string
	stderr   string
	exitCode int
}

type invocation struct {
	dir     string
	env     []string
	timeout time.Duration
	stdin   string
}

func runTool(inv invocation, bin string, args ...string) (result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), inv.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = inv.dir
	cmd.Env = append(os.Environ(), inv.env...)
	if inv.stdin != "" {
		cmd.Stdin = strings.NewReader(inv.stdin)
	}
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	res := result{stdout: out.String(), stderr: errb.String()}
	if ctx.Err() != nil {
		return res, fmt.Errorf("%s timed out after %s", path.Base(bin), inv.timeout)
	}
	if exit, ok := err.(*exec.ExitError); ok {
		res.exitCode = exit.ExitCode()
		return res, nil
	}
	return res, err
}

func toolEnv(r *check.Run) []string {
	return []string{
		"PYTHONDONTWRITEBYTECODE=1",
		"XDG_CACHE_HOME=" + filepath.Join(r.TempDir, "cache"),
	}
}

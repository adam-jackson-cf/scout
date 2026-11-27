package testrun

import (
	"path"
	"strings"
	"time"

	"scout/tools/gate/check"
	"scout/tools/gate/testsource"
)

func (s *session) affected(r *check.Run) (map[string]bool, error) {
	changes, err := r.Diff()
	if err != nil {
		return nil, err
	}
	dirs, err := s.packageDirs(r)
	if err != nil {
		return nil, err
	}
	changed := map[string]bool{}
	for _, ch := range changes {
		for _, p := range []string{ch.Path, ch.From} {
			if p == "" {
				continue
			}
			dir, ok := owningDir(p, dirs)
			if !ok {
				return nil, nil
			}
			changed[importPathOf(s.col, dir)] = true
		}
	}
	return s.dependents(changed)
}

func (s *session) packageDirs(r *check.Run) (map[string]bool, error) {
	files, err := r.Tracked()
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, f := range files {
		if strings.HasSuffix(f, ".go") && !testsource.Excluded(s.col.Exclude, f) && !strings.Contains("/"+f, "/testdata/") {
			out[path.Dir(f)] = true
		}
	}
	return out, nil
}

func owningDir(p string, dirs map[string]bool) (string, bool) {
	if i := strings.Index("/"+p, "/testdata/"); i >= 0 {
		owner := strings.TrimSuffix(("/" + p)[:i], "/")
		owner = strings.TrimPrefix(owner, "/")
		if owner == "" {
			owner = "."
		}
		return owner, dirs[owner]
	}
	d := path.Dir(p)
	return d, dirs[d]
}

func (s *session) dependents(changed map[string]bool) (map[string]bool, error) {
	var pkgs []string
	for key := range s.list.byTier[TierT2] {
		pkgs = append(pkgs, s.list.byKey[key].ImportPath)
	}
	if len(pkgs) == 0 {
		return map[string]bool{}, nil
	}
	args := append([]string{"list", "-test", "-tags", s.col.Tiers.Boundary, "-f", "{{.ImportPath}}|{{join .Deps \",\"}}"}, uniq(pkgs)...)
	res, err := goCommand(s.col.Root, s.list.env, 5*time.Minute, args...)
	if err != nil || res.exit != 0 {
		return nil, err
	}
	out := map[string]bool{}
	for _, line := range strings.Split(string(res.out), "\n") {
		name, deps, ok := strings.Cut(line, "|")
		if !ok {
			continue
		}
		base := strings.Fields(name)[0]
		hit := changed[base]
		for _, d := range strings.Split(deps, ",") {
			hit = hit || changed[strings.Fields(d + " ")[0]]
		}
		if hit {
			out[base] = true
		}
	}
	return out, nil
}

func uniq(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

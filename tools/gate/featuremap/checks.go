package featuremap

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"scout/tools/gate/check"
	"scout/tools/gate/pathscope"
)

const rule = "SURFACE.feature-map"

type surface struct {
	FeatureMapPath string   `toml:"feature_map_path"`
	Index          string   `toml:"feature_map_index"`
	Lever          []string `toml:"lever_command"`
}

func Checks() []check.Check {
	return []check.Check{{Name: "featuremap", Rules: []string{rule}, Run: run}}
}

func Sections(lever []string) []string {
	return []string{"Sub-features", "How to get to it (user POV)", "Driving it with `" + strings.Join(lever, " ") + "`", "Gotchas"}
}

type Input struct {
	Dir      string
	Index    string
	Sections []string
	Files    map[string]bool
	Tests    map[string]bool
	Read     func(string) ([]byte, error)
}

func run(r *check.Run) ([]check.Finding, error) {
	sf, err := loadSurface(r)
	if err != nil {
		return nil, err
	}
	s, err := pathscope.New(r)
	if err != nil {
		return nil, err
	}
	in := Input{Dir: strings.TrimSuffix(sf.FeatureMapPath, "/"), Index: sf.Index, Sections: Sections(sf.Lever), Files: map[string]bool{}, Tests: map[string]bool{}}
	prefix := fixturePrefix(s, in.Dir)
	files, err := gatherFiles(s, prefix)
	if err != nil {
		return nil, err
	}
	in.Read = reader(s, prefix)
	for _, f := range files {
		in.Files[f] = true
	}
	if err := collectTests(r.Root, s, files, prefix, in.Tests); err != nil {
		return nil, err
	}
	return Judge(in)
}

func loadSurface(r *check.Run) (surface, error) {
	var sf surface
	ok, err := r.Proofs.Section("surface", &sf)
	if err != nil {
		return sf, err
	}
	if !ok || sf.FeatureMapPath == "" || sf.Index == "" || len(sf.Lever) == 0 {
		return sf, errors.New("proofs.toml [surface] needs feature_map_path, feature_map_index and lever_command")
	}
	return sf, nil
}

func fixturePrefix(s *pathscope.Scope, dir string) string {
	if s.Fixture != "" {
		return dir + "/"
	}
	return ""
}

func gatherFiles(s *pathscope.Scope, prefix string) ([]string, error) {
	var files []string
	if prefix != "" {
		for _, e := range s.Entries {
			files = append(files, prefix+e.Path)
		}
	}
	head, err := headFiles(s)
	if err != nil {
		return nil, err
	}
	for _, f := range head {
		if prefix != "" && strings.HasPrefix(f, prefix) {
			continue
		}
		files = append(files, f)
	}
	return files, nil
}

func reader(s *pathscope.Scope, prefix string) func(string) ([]byte, error) {
	return func(p string) ([]byte, error) {
		if prefix != "" && strings.HasPrefix(p, prefix) {
			return s.ReadHead(strings.TrimPrefix(p, prefix))
		}
		return readRepo(s, p)
	}
}

func collectTests(root string, s *pathscope.Scope, files []string, prefix string, tests map[string]bool) error {
	for _, f := range files {
		if !strings.HasSuffix(f, "_test.go") || (prefix == "" && s.Excluded(f)) {
			continue
		}
		if err := addTests(testDiskPath(root, s, prefix, f), tests); err != nil {
			return err
		}
	}
	return nil
}

func testDiskPath(root string, s *pathscope.Scope, prefix, f string) string {
	if prefix != "" && strings.HasPrefix(f, prefix) {
		return filepath.Join(s.Fixture, filepath.FromSlash(strings.TrimPrefix(f, prefix)))
	}
	return filepath.Join(root, filepath.FromSlash(f))
}

func addTests(disk string, tests map[string]bool) error {
	src, err := os.ReadFile(disk)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, m := range testFunc.FindAllStringSubmatch(string(src), -1) {
		tests[m[1]] = true
	}
	return nil
}

func headFiles(s *pathscope.Scope) ([]string, error) {
	if s.Fixture != "" || s.Run.Explicit {
		saved := s.Fixture
		s.Fixture = ""
		defer func() { s.Fixture = saved }()
	}
	return s.HeadFiles()
}

func readRepo(s *pathscope.Scope, p string) ([]byte, error) {
	saved := s.Fixture
	s.Fixture = ""
	defer func() { s.Fixture = saved }()
	if s.Run.Explicit {
		saved := s.Full
		s.Full = true
		defer func() { s.Full = saved }()
	}
	return s.ReadHead(p)
}

var (
	testFunc = regexp.MustCompile(`(?m)^func (Test[A-Za-z0-9_]*)\(`)
	testName = regexp.MustCompile(`\bTest[A-Z][A-Za-z0-9_]*`)
	linkRE   = regexp.MustCompile(`\]\(([^)\s]+)\)`)
)

func Judge(in Input) ([]check.Finding, error) {
	index := in.Dir + "/" + in.Index
	var docs []string
	for f := range in.Files {
		if path.Dir(f) == in.Dir && strings.HasSuffix(f, ".md") && f != index {
			docs = append(docs, f)
		}
	}
	sort.Strings(docs)
	var out []check.Finding
	if !in.Files[index] {
		return []check.Finding{{File: index, Rule: rule, Problem: "feature map has no index", Fix: "add " + index + " linking one document per feature"}}, nil
	}
	isrc, err := in.Read(index)
	if err != nil {
		return nil, err
	}
	linked := map[string]bool{}
	out = append(out, checkLinks(in, index, isrc, linked)...)
	for _, d := range docs {
		if !linked[d] {
			out = append(out, check.Finding{File: d, Rule: rule, Problem: "feature document is not listed in the index", Fix: "add a row linking it to " + index})
		}
		src, err := in.Read(d)
		if err != nil {
			return nil, err
		}
		out = append(out, sections(in, d, src)...)
		out = append(out, checkLinks(in, d, src, nil)...)
		out = append(out, checkTests(in, d, src)...)
	}
	return out, nil
}

func sections(in Input, file string, src []byte) []check.Finding {
	var got []string
	for _, line := range strings.Split(string(src), "\n") {
		if strings.HasPrefix(line, "## ") {
			got = append(got, strings.TrimSpace(strings.TrimPrefix(line, "## ")))
		}
	}
	if strings.Join(got, "\x00") == strings.Join(in.Sections, "\x00") {
		return nil
	}
	return []check.Finding{{File: file, Rule: rule,
		Problem: fmt.Sprintf("H2 sections are [%s], not exactly [%s]", strings.Join(got, " | "), strings.Join(in.Sections, " | ")),
		Fix:     "use exactly the four H2 sections in that order and move other content under them"}}
}

func checkLinks(in Input, file string, src []byte, linked map[string]bool) []check.Finding {
	var out []check.Finding
	for i, line := range strings.Split(string(src), "\n") {
		for _, m := range linkRE.FindAllStringSubmatch(line, -1) {
			target := m[1]
			if strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") || strings.HasPrefix(target, "#") {
				continue
			}
			target = strings.SplitN(target, "#", 2)[0]
			p := path.Clean(path.Join(path.Dir(file), target))
			if linked != nil {
				linked[p] = true
			}
			if !in.Files[p] && !dirExists(in.Files, p) {
				out = append(out, check.Finding{File: file, Line: i + 1, Rule: rule, Problem: fmt.Sprintf("stale link %q", m[1]), Fix: "point the link at an existing file or remove it"})
			}
		}
	}
	return out
}

func checkTests(in Input, file string, src []byte) []check.Finding {
	var out []check.Finding
	seen := map[string]bool{}
	for i, line := range strings.Split(string(src), "\n") {
		for _, n := range testName.FindAllString(line, -1) {
			if seen[n] || resolves(in.Tests, n) {
				continue
			}
			seen[n] = true
			out = append(out, check.Finding{File: file, Line: i + 1, Rule: rule, Problem: fmt.Sprintf("names test %s, which matches no test function in a tracked _test.go file", n), Fix: "name the test that proves the feature, or restore it"})
		}
	}
	return out
}

func resolves(tests map[string]bool, name string) bool {
	for t := range tests {
		if strings.HasPrefix(t, name) {
			return true
		}
	}
	return false
}

func dirExists(files map[string]bool, p string) bool {
	for f := range files {
		if strings.HasPrefix(f, p+"/") {
			return true
		}
	}
	return false
}

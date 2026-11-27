package skill

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"scout/tools/gate/check"
	"scout/tools/gate/pathscope"
)

const rule = "SKILL.metadata"

type Layout struct {
	Root     string   `toml:"root"`
	Required []string `toml:"frontmatter_required"`
	Allowed  []string `toml:"frontmatter_allowed"`
}

type layout struct {
	Skills Layout `toml:"skills"`
}

func Checks() []check.Check {
	return []check.Check{{Name: "skill", Rules: []string{rule}, Run: run}}
}

func run(r *check.Run) ([]check.Finding, error) {
	l, err := loadLayout(r)
	if err != nil {
		return nil, err
	}
	s, err := pathscope.New(r)
	if err != nil {
		return nil, err
	}
	root := strings.TrimSuffix(l.Skills.Root, "/") + "/"
	files, err := skillFiles(s, root)
	if err != nil {
		return nil, err
	}
	return Judge(l.Skills, files, skillReader(s, root))
}

func loadLayout(r *check.Run) (layout, error) {
	var l layout
	if _, err := r.Proofs.Section("layout", &l); err != nil {
		return l, err
	}
	if l.Skills.Root == "" || len(l.Skills.Required) == 0 || len(l.Skills.Allowed) == 0 {
		return l, errors.New("proofs.toml [layout.skills] needs root, frontmatter_required and frontmatter_allowed")
	}
	return l, nil
}

func skillFiles(s *pathscope.Scope, root string) ([]string, error) {
	var files []string
	if s.Fixture != "" {
		for _, e := range s.Entries {
			files = append(files, root+e.Path)
		}
		return files, nil
	}
	head, err := s.HeadFiles()
	if err != nil {
		return nil, err
	}
	for _, f := range head {
		if strings.HasPrefix(f, root) {
			files = append(files, f)
		}
	}
	return files, nil
}

func skillReader(s *pathscope.Scope, root string) func(string) ([]byte, error) {
	return func(p string) ([]byte, error) {
		if s.Fixture != "" {
			return s.ReadHead(strings.TrimPrefix(p, root))
		}
		return s.ReadHead(p)
	}
}

func Judge(l Layout, files []string, read func(string) ([]byte, error)) ([]check.Finding, error) {
	root := strings.TrimSuffix(l.Root, "/") + "/"
	have, names := skillIndex(files, root)
	set := skillSet{files: files, have: have, read: read}
	var out []check.Finding
	for _, name := range names {
		fs, err := skillFindings(l, set, root+name, name)
		if err != nil {
			return nil, err
		}
		out = append(out, fs...)
	}
	return out, nil
}

func skillIndex(files []string, root string) (map[string]bool, []string) {
	have := map[string]bool{}
	skills := map[string]bool{}
	for _, f := range files {
		have[f] = true
		rest := strings.TrimPrefix(f, root)
		if i := strings.Index(rest, "/"); i > 0 {
			skills[rest[:i]] = true
		}
	}
	names := make([]string, 0, len(skills))
	for n := range skills {
		names = append(names, n)
	}
	sort.Strings(names)
	return have, names
}

type skillSet struct {
	files []string
	have  map[string]bool
	read  func(string) ([]byte, error)
}

func skillFindings(l Layout, set skillSet, dir, name string) ([]check.Finding, error) {
	md := dir + "/SKILL.md"
	if !set.have[md] {
		return []check.Finding{{File: dir, Rule: rule, Problem: "skill folder has no SKILL.md", Fix: "add " + md + " with quoted name and description frontmatter"}}, nil
	}
	src, err := set.read(md)
	if err != nil {
		return nil, err
	}
	out := frontmatter(l, name, md, src)
	for _, f := range set.files {
		if f != md && !reference(dir, f) {
			continue
		}
		body := src
		if f != md {
			if body, err = set.read(f); err != nil {
				return nil, err
			}
		}
		out = append(out, links(dir, f, body, set.have)...)
	}
	return out, nil
}

func reference(dir, f string) bool {
	return strings.HasPrefix(f, dir+"/references/") && strings.HasSuffix(f, ".md")
}

func lineOne(file, problem, fix string) check.Finding {
	return check.Finding{File: file, Line: 1, Rule: rule, Problem: problem, Fix: fix}
}

func frontmatter(l Layout, name, file string, src []byte) []check.Finding {
	if !bytes.HasPrefix(src, []byte("---\n")) {
		return []check.Finding{lineOne(file, "SKILL.md does not start with YAML frontmatter", "start the file with ---, quoted name and description, then ---")}
	}
	end := bytes.Index(src[4:], []byte("\n---"))
	if end < 0 {
		return []check.Finding{lineOne(file, "frontmatter is not closed", "close the frontmatter with a --- line")}
	}
	var meta map[string]any
	if err := yaml.Unmarshal(src[4:4+end], &meta); err != nil {
		return []check.Finding{lineOne(file, "frontmatter is not valid YAML: "+err.Error(), "fix the YAML and quote every string value")}
	}
	out := keyFindings(l, file, meta)
	return append(out, identityFindings(name, file, meta)...)
}

func keyFindings(l Layout, file string, meta map[string]any) []check.Finding {
	var out []check.Finding
	allowed := map[string]bool{}
	for _, k := range l.Allowed {
		allowed[k] = true
	}
	for _, k := range l.Required {
		if _, ok := meta[k]; !ok {
			out = append(out, lineOne(file, fmt.Sprintf("frontmatter lacks required key %q", k), fmt.Sprintf("add %s: \"...\" to the frontmatter", k)))
		}
	}
	keys := make([]string, 0, len(meta))
	for k := range meta {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !allowed[k] {
			out = append(out, lineOne(file, fmt.Sprintf("frontmatter key %q is outside the skill schema (%s)", k, strings.Join(l.Allowed, ", ")), "remove the key or add it to proofs.toml [layout.skills].frontmatter_allowed"))
		}
	}
	return out
}

func identityFindings(name, file string, meta map[string]any) []check.Finding {
	var out []check.Finding
	if n, _ := meta["name"].(string); n != name {
		out = append(out, lineOne(file, fmt.Sprintf("name %q does not match folder %q", n, name), fmt.Sprintf("set name: %q", name)))
	}
	if d, _ := meta["description"].(string); !strings.HasPrefix(strings.ToLower(strings.TrimSpace(d)), "use when") {
		out = append(out, lineOne(file, "description does not start with \"use when\"", "rewrite the description as one routing sentence starting with \"USE WHEN\""))
	}
	return out
}

var linkRE = regexp.MustCompile(`\]\(([^)\s]+)\)`)

func links(dir, file string, src []byte, have map[string]bool) []check.Finding {
	var out []check.Finding
	for i, line := range strings.Split(string(src), "\n") {
		for _, m := range linkRE.FindAllStringSubmatch(line, -1) {
			target := m[1]
			if strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") || strings.HasPrefix(target, "#") {
				continue
			}
			target = strings.SplitN(target, "#", 2)[0]
			p := path.Clean(path.Join(path.Dir(file), target))
			if !strings.HasPrefix(p, dir+"/") {
				out = append(out, check.Finding{File: file, Line: i + 1, Rule: rule, Problem: fmt.Sprintf("link %q leaves the skill folder", m[1]), Fix: "ship the linked file inside the skill folder and link it relatively"})
				continue
			}
			if !have[p] {
				out = append(out, check.Finding{File: file, Line: i + 1, Rule: rule, Problem: fmt.Sprintf("link %q points at a missing file", m[1]), Fix: "add " + p + " or fix the link"})
			}
		}
	}
	return out
}

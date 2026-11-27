package pathscope

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"scout/tools/gate/check"
)

type Entry struct {
	Path     string
	From     string
	Disk     string
	FirstNew int
	Fixture  bool
}

type Layout struct {
	Exclude []string `toml:"exclude"`
}

type Scope struct {
	Run     *check.Run
	Entries []Entry
	Base    string
	Full    bool
	Fixture string
	exclude []string
}

func New(r *check.Run) (*Scope, error) {
	var lay Layout
	if r.Proofs != nil {
		if _, err := r.Proofs.Section("layout", &lay); err != nil {
			return nil, err
		}
	}
	s := &Scope{Run: r, exclude: lay.Exclude}
	switch {
	case r.Explicit:
		return s, s.explicit()
	case r.Mode == check.CI && r.Base == "":
		s.Full = true
		return s, s.full()
	default:
		return s, s.diff()
	}
}

func (s *Scope) Excluded(p string) bool {
	for _, pat := range s.exclude {
		if Match(pat, p) {
			return true
		}
	}
	return false
}

func (s *Scope) explicit() error {
	for _, p := range s.Run.Paths {
		abs := p
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(s.Run.Root, p)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			s.addExplicitFile(abs)
			continue
		}
		if err := s.addFixture(abs); err != nil {
			return err
		}
	}
	sortEntries(s.Entries)
	return nil
}

func (s *Scope) addExplicitFile(abs string) {
	rel := filepath.Base(abs)
	if r, err := filepath.Rel(s.Run.Root, abs); err == nil && !strings.HasPrefix(r, "..") {
		rel = filepath.ToSlash(r)
	}
	s.Entries = append(s.Entries, Entry{Path: rel, Disk: abs, FirstNew: len(strings.Split(rel, "/")) - 1})
}

func (s *Scope) addFixture(abs string) error {
	s.Fixture = abs
	return filepath.WalkDir(abs, func(fp string, d fs.DirEntry, werr error) error {
		if werr != nil || d.IsDir() {
			return werr
		}
		rel, rerr := filepath.Rel(abs, fp)
		if rerr != nil {
			return rerr
		}
		s.Entries = append(s.Entries, Entry{Path: filepath.ToSlash(rel), Disk: fp, Fixture: true})
		return nil
	})
}

func (s *Scope) full() error {
	files, err := s.Run.Tracked()
	if err != nil {
		return err
	}
	for _, f := range files {
		if s.Excluded(f) {
			continue
		}
		s.Entries = append(s.Entries, Entry{Path: f, Disk: filepath.Join(s.Run.Root, filepath.FromSlash(f))})
	}
	sortEntries(s.Entries)
	return nil
}

func (s *Scope) diff() error {
	base, err := BaseRev(s.Run)
	if err != nil {
		return err
	}
	s.Base = base
	changes, err := s.Run.Diff()
	if err != nil {
		return err
	}
	dirs, err := s.baseDirs(base)
	if err != nil {
		return err
	}
	for _, c := range changes {
		if !addedStatus(c.Status) || s.Excluded(c.Path) {
			continue
		}
		s.Entries = append(s.Entries, Entry{Path: c.Path, From: c.From, Disk: filepath.Join(s.Run.Root, filepath.FromSlash(c.Path)), FirstNew: firstNewSegment(c.Path, dirs)})
	}
	sortEntries(s.Entries)
	return nil
}

func (s *Scope) baseDirs(base string) (map[string]bool, error) {
	dirs := map[string]bool{}
	if base == "" {
		return dirs, nil
	}
	files, err := s.TreeFiles(base)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		for d := path.Dir(f); d != "."; d = path.Dir(d) {
			dirs[d] = true
		}
	}
	return dirs, nil
}

func addedStatus(status string) bool {
	return status == "A" || status == "R" || status == "C"
}

func firstNewSegment(p string, dirs map[string]bool) int {
	segs := strings.Split(p, "/")
	for i := 0; i < len(segs)-1; i++ {
		if !dirs[strings.Join(segs[:i+1], "/")] {
			return i
		}
	}
	return len(segs) - 1
}

func BaseRev(r *check.Run) (string, error) {
	if r.Explicit {
		return "", nil
	}
	if r.Mode == check.Commit {
		if _, err := r.Git("rev-parse", "--verify", "--quiet", "HEAD^{commit}"); err != nil {
			return "", nil
		}
		return "HEAD", nil
	}
	if r.Base == "" {
		return "", nil
	}
	mb, err := r.Git("merge-base", r.Base, "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(mb), nil
}

func (s *Scope) HeadRev() string {
	if s.Run.Mode == check.Commit {
		return ""
	}
	return "HEAD"
}

func (s *Scope) TreeFiles(rev string) ([]string, error) {
	raw, err := s.Run.Git("ls-tree", "-r", "-z", "--name-only", "--full-tree", rev)
	if err != nil {
		return nil, err
	}
	return splitZ(raw), nil
}

func (s *Scope) HeadFiles() ([]string, error) {
	if s.Fixture != "" {
		out := make([]string, 0, len(s.Entries))
		for _, e := range s.Entries {
			out = append(out, e.Path)
		}
		return out, nil
	}
	if s.Run.Mode == check.Commit {
		raw, err := s.Run.Git("ls-files", "-z", "--cached")
		if err != nil {
			return nil, err
		}
		return splitZ(raw), nil
	}
	return s.TreeFiles("HEAD")
}

func (s *Scope) Read(e Entry) ([]byte, error) {
	if s.Run.Explicit || s.Full {
		return os.ReadFile(e.Disk)
	}
	return s.Show(e.Path)
}

func (s *Scope) Show(p string) ([]byte, error) {
	spec := s.HeadRev() + ":" + p
	out, err := s.Run.Git("show", spec)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

func (s *Scope) Disk(p string) string {
	if s.Fixture != "" {
		return filepath.Join(s.Fixture, filepath.FromSlash(p))
	}
	return filepath.Join(s.Run.Root, filepath.FromSlash(p))
}

func (s *Scope) ReadHead(p string) ([]byte, error) {
	if s.Fixture != "" || s.Full {
		return os.ReadFile(s.Disk(p))
	}
	data, err := s.Show(p)
	if err != nil {
		if _, statErr := os.Stat(s.Disk(p)); errors.Is(statErr, os.ErrNotExist) {
			return nil, os.ErrNotExist
		}
	}
	return data, err
}

func splitZ(raw string) []string {
	var out []string
	for _, p := range strings.Split(raw, "\x00") {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func sortEntries(es []Entry) {
	sort.Slice(es, func(i, j int) bool { return es[i].Path < es[j].Path })
}

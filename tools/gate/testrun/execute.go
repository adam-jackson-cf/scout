package testrun

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"scout/tools/gate/check"
	"scout/tools/gate/testsource"
)

type session struct {
	once     sync.Once
	err      error
	col      *testsource.Collection
	list     *listing
	reports  map[string]string
	guard    sandbox
	realHome string
	started  time.Time
	timeout  time.Duration
	goVer    string
}

var (
	mu       sync.Mutex
	sessions = map[*check.Run]*session{}
)

func sessionFor(r *check.Run) *session {
	mu.Lock()
	defer mu.Unlock()
	s := sessions[r]
	if s == nil {
		s = &session{reports: map[string]string{}}
		sessions[r] = s
	}
	return s
}

func (s *session) prepare(r *check.Run) error {
	s.once.Do(func() { s.err = s.load(r) })
	return s.err
}

func (s *session) load(r *check.Run) error {
	s.started = time.Now()
	col, err := testsource.Collect(r)
	if err != nil {
		return err
	}
	s.col = col
	if s.timeout, err = time.ParseDuration(col.Tiers.PackageTimeout); err != nil || s.timeout <= 0 {
		return fmt.Errorf("proofs.toml [layout.test_tiers] package_timeout %q is not a positive duration", col.Tiers.PackageTimeout)
	}
	env, err := s.environment(r)
	if err != nil {
		return err
	}
	s.list, err = buildListing(r, col, env)
	return err
}

func (s *session) environment(r *check.Run) ([]string, error) {
	vals, err := goEnv(s.col.Root, "GOCACHE", "GOMODCACHE", "GOPATH", "GOENV", "GOVERSION")
	if err != nil {
		return nil, err
	}
	s.goVer = vals["GOVERSION"]
	if s.realHome, err = accountHome(); err != nil {
		return nil, err
	}
	if s.realHome, err = realPath(s.realHome); err != nil {
		return nil, err
	}
	home := filepath.Join(r.TempDir, "home")
	goTmp := filepath.Join(r.TempDir, "gotmp")
	for _, d := range []string{home, goTmp, filepath.Join(home, ".config"), filepath.Join(home, ".cache"), filepath.Join(home, ".local", "share"), filepath.Join(home, ".local", "state")} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			return nil, err
		}
	}
	if goTmp, err = realPath(goTmp); err != nil {
		return nil, err
	}
	skip := map[string]bool{"HOME": true, "XDG_CONFIG_HOME": true, "XDG_CACHE_HOME": true, "XDG_DATA_HOME": true, "XDG_STATE_HOME": true, "GOTMPDIR": true, "GOFLAGS": true, "GOCACHE": true, "GOMODCACHE": true, "GOPATH": true, "GOENV": true}
	env := append(inheritedEnv(skip), "HOME="+home, "XDG_CONFIG_HOME="+filepath.Join(home, ".config"), "XDG_CACHE_HOME="+filepath.Join(home, ".cache"),
		"XDG_DATA_HOME="+filepath.Join(home, ".local", "share"), "XDG_STATE_HOME="+filepath.Join(home, ".local", "state"),
		"GOTMPDIR="+goTmp, "GOCACHE="+vals["GOCACHE"], "GOMODCACHE="+vals["GOMODCACHE"], "GOPATH="+vals["GOPATH"], "GOENV="+vals["GOENV"], "GOWORK=off")
	tmp, err := realPath(r.TempDir)
	if err != nil {
		return nil, err
	}
	s.guard, err = t1Guard(tmp, goTmp, s.realHome)
	return env, err
}

func inheritedEnv(skip map[string]bool) []string {
	var env []string
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		if skip[k] || strings.HasPrefix(k, "GIT_") && k != "GIT_CEILING_DIRECTORIES" {
			continue
		}
		env = append(env, kv)
	}
	return env
}

func (s *session) tiersFor(r *check.Run) ([]string, error) {
	if r.Mode == check.CI || r.Explicit {
		return []string{TierT1, TierT1Deferred, TierT2}, nil
	}
	return []string{TierT1, TierT2}, nil
}

func (s *session) run(r *check.Run, tier string) (string, error) {
	if p, ok := s.reports[tier]; ok {
		return p, nil
	}
	want, err := s.selection(r, tier)
	if err != nil {
		return "", err
	}
	path := filepath.Join(r.TempDir, "report-"+tier+".xml")
	s.reports[tier] = path
	if len(want) == 0 {
		return path, WriteReport(path, Report{Tier: tier, Timestamp: time.Now().UTC().Format(time.RFC3339Nano)})
	}
	args, guard := s.args(tier, want)
	res, err := goCommand(s.col.Root, s.list.env, 4*s.timeout+10*time.Minute, args...)
	if err != nil {
		return "", err
	}
	return path, WriteReport(path, Normalise(tier, guard, res, parseEvents(res.out)))
}

func (s *session) args(tier string, want map[string]bool) ([]string, string) {
	t := s.col.Tiers
	args := []string{"test", "-json", "-count=1", "-timeout", s.timeout.String()}
	guard := ""
	switch tier {
	case TierT1:
		args = append(args, "-exec", s.guard.exec)
		guard = s.guard.summary
	case TierT1Deferred:
		args = append(args, "-exec", s.guard.exec, "-tags", t.Deferred)
		guard = s.guard.summary
	case TierT2:
		args = append(args, "-tags", t.Boundary)
	}
	pkgs, names := map[string]bool{}, map[string]bool{}
	for key := range want {
		st := s.list.byKey[key]
		pkgs[st.ImportPath], names[st.Name] = true, true
	}
	if tier != TierT1 {
		args = append(args, "-run", "^("+strings.Join(sortedKeys(names, regexp.QuoteMeta), "|")+")$")
	}
	return append(args, sortedKeys(pkgs, nil)...), guard
}

func sortedKeys(m map[string]bool, quote func(string) string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		if quote != nil {
			k = quote(k)
		}
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func (s *session) selection(r *check.Run, tier string) (map[string]bool, error) {
	all := s.list.byTier[tier]
	if tier != TierT2 || r.Mode == check.CI || r.Explicit {
		return all, nil
	}
	affected, err := s.affected(r)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for key := range all {
		if affected == nil || affected[s.list.byKey[key].ImportPath] {
			out[key] = true
		}
	}
	return out, nil
}

func accountHome() (string, error) {
	if u, err := user.Current(); err == nil && u.HomeDir != "" {
		return u.HomeDir, nil
	}
	return os.UserHomeDir()
}

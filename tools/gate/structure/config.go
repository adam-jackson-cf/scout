package structure

import (
	"errors"
	"path"
	"strings"

	"scout/tools/gate/check"
	"scout/tools/gate/pathscope"
)

type Config struct {
	ImportPackage    string            `toml:"import_package"`
	SharedTiers      int               `toml:"shared_tiers"`
	Tiers            [][]string        `toml:"tiers"`
	Owners           map[string]string `toml:"owners"`
	PrefixClusterMin int               `toml:"prefix_cluster_min"`
}

type SkillLayout struct {
	Root    string   `toml:"root"`
	Allowed []string `toml:"allowed"`
}

type layout struct {
	Skills SkillLayout `toml:"skills"`
}

func load(r *check.Run) (Config, SkillLayout, error) {
	var c Config
	var l layout
	if r.Proofs == nil {
		return c, l.Skills, errors.New("proofs.toml is missing")
	}
	ok, err := r.Proofs.Section("structure", &c)
	if err != nil {
		return c, l.Skills, err
	}
	if !ok || c.ImportPackage == "" || len(c.Tiers) == 0 || len(c.Owners) == 0 {
		return c, l.Skills, errors.New("proofs.toml [structure] needs import_package, tiers and owners")
	}
	if _, err := r.Proofs.Section("layout", &l); err != nil {
		return c, l.Skills, err
	}
	return c, l.Skills, nil
}

func (c Config) tierOf(dir string) int {
	for i, tier := range c.Tiers {
		for _, p := range tier {
			if p == dir {
				return i
			}
		}
	}
	for i, tier := range c.Tiers {
		for _, p := range tier {
			if strings.ContainsAny(p, "*?[") && pathscope.Match(p, dir) {
				return i
			}
		}
	}
	return -1
}

func (c Config) owner(p string) (string, bool) {
	best, name := "", ""
	for k, v := range c.Owners {
		if (k == p || (strings.HasSuffix(k, "/") && strings.HasPrefix(p, k))) && len(k) > len(best) {
			best, name = k, v
		}
	}
	return name, best != ""
}

func (c Config) nearestOwner(p string) string {
	best, score := "", -1
	segs := strings.Split(path.Dir(p), "/")
	for k := range c.Owners {
		if !strings.HasSuffix(k, "/") {
			continue
		}
		ks := strings.Split(strings.TrimSuffix(k, "/"), "/")
		n := 0
		for n < len(ks) && n < len(segs) && ks[n] == segs[n] {
			n++
		}
		if n > 0 && (n > score || (n == score && k < best)) {
			best, score = k, n
		}
	}
	return best
}

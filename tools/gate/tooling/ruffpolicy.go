package tooling

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"

	"scout/tools/gate/check"
)

const ruffConfig = "ruff.toml"

type ruffIgnores struct {
	TargetVersion string   `toml:"target-version"`
	Ignore        []string `toml:"ignore"`
	ExtendIgnore  []string `toml:"extend-ignore"`
	Lint          struct {
		Ignore         []string            `toml:"ignore"`
		ExtendIgnore   []string            `toml:"extend-ignore"`
		PerFileIgnores map[string][]string `toml:"per-file-ignores"`
		ExtendPFI      map[string][]string `toml:"extend-per-file-ignores"`
	} `toml:"lint"`
}

func (c ruffIgnores) entries() map[string]bool {
	out := map[string]bool{}
	for _, list := range [][]string{c.Ignore, c.ExtendIgnore, c.Lint.Ignore, c.Lint.ExtendIgnore} {
		for _, code := range list {
			out["*\x00"+code] = true
		}
	}
	for _, m := range []map[string][]string{c.Lint.PerFileIgnores, c.Lint.ExtendPFI} {
		for glob, codes := range m {
			for _, code := range codes {
				out[glob+"\x00"+code] = true
			}
		}
	}
	return out
}

func ruffPolicy(r *check.Run) ([]check.Finding, error) {
	data, err := os.ReadFile(filepath.Join(r.Root, ruffConfig))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ruffConfig, err)
	}
	var head ruffIgnores
	if _, err := toml.Decode(string(data), &head); err != nil {
		return nil, fmt.Errorf("%s: %w", ruffConfig, err)
	}
	minor, err := pythonMinor(r.Root)
	if err != nil {
		return nil, err
	}
	var out []check.Finding
	if want := "py" + strings.ReplaceAll(minor, ".", ""); head.TargetVersion != want {
		out = append(out, check.Finding{File: ruffConfig, Rule: "RUFF.SETTING.target-version", Problem: fmt.Sprintf("target-version %q differs from toolchain/versions.env PYTHON_VERSION %s", head.TargetVersion, minor), Fix: "set target-version = \"" + want + "\""})
	}
	revs, ok, err := revisions(r)
	if err != nil || !ok {
		return out, err
	}
	baseText, found, err := showAt(r, revs.base, ruffConfig)
	if err != nil {
		return nil, err
	}
	var base ruffIgnores
	if found {
		if _, err := toml.Decode(baseText, &base); err != nil {
			return nil, fmt.Errorf("%s at %s: %w", ruffConfig, revs.base, err)
		}
	}
	return append(out, addedIgnores(base, head, string(data))...), nil
}

func addedIgnores(base, head ruffIgnores, headText string) []check.Finding {
	before := base.entries()
	var added []string
	for k := range head.entries() {
		if !before[k] {
			added = append(added, k)
		}
	}
	sort.Strings(added)
	lines := strings.Split(headText, "\n")
	var out []check.Finding
	for _, k := range added {
		glob, code, _ := strings.Cut(k, "\x00")
		line, justified := rationaleLine(lines, code)
		if justified {
			continue
		}
		scope := "all files"
		if glob != "*" {
			scope = glob
		}
		out = append(out, check.Finding{File: ruffConfig, Line: line, Rule: "RUFF.IGNORES", Problem: fmt.Sprintf("adds an ignore of %s for %s without a rationale", code, scope), Fix: "fix the code instead, or add a same-line # comment giving the owner-approved reason"})
	}
	return out
}

func rationaleLine(lines []string, code string) (int, bool) {
	quoted := regexp.MustCompile(`["']` + regexp.QuoteMeta(code) + `["']`)
	first := 0
	for n, line := range lines {
		loc := quoted.FindStringIndex(line)
		if loc == nil {
			continue
		}
		if first == 0 {
			first = n + 1
		}
		if c := strings.Index(line[loc[1]:], "#"); c >= 0 && strings.TrimSpace(line[loc[1]+c+1:]) != "" {
			return n + 1, true
		}
	}
	return first, false
}

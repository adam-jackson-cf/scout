package structure

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"scout/tools/gate/check"
	"scout/tools/gate/pathscope"
)

const ruleRename = "RENAME.residue"

type retired struct {
	Old, New, From, To string
}

func retiredNames(from, to string) []retired {
	fs, ts := strings.Split(from, "/"), strings.Split(to, "/")
	toSet := segmentNames(ts)
	var out []retired
	for i, s := range fs {
		name := segmentName(fs, i)
		if toSet[s] || toSet[name] || name == "" {
			continue
		}
		repl := replacement(fs, ts, i)
		if sameTokens(name, repl) {
			continue
		}
		out = append(out, retired{Old: name, New: repl, From: from, To: to})
	}
	return out
}

func segmentNames(segs []string) map[string]bool {
	set := map[string]bool{}
	for i, s := range segs {
		set[s] = true
		if i == len(segs)-1 {
			set[pathscope.Stem(s)] = true
		}
	}
	return set
}

func segmentName(segs []string, i int) string {
	if i == len(segs)-1 {
		return pathscope.Stem(segs[i])
	}
	return segs[i]
}

func replacement(fs, ts []string, i int) string {
	if len(fs) == len(ts) {
		return segmentName(ts, i)
	}
	return pathscope.Stem(ts[len(ts)-1])
}

func sameTokens(a, b string) bool {
	return strings.Join(pathscope.Tokens(a), "/") == strings.Join(pathscope.Tokens(b), "/")
}

func variants(name string) []string {
	toks := pathscope.Tokens(name)
	out := map[string]bool{name: true}
	if len(toks) > 1 {
		out[strings.Join(toks, "_")] = true
		out[strings.Join(toks, "-")] = true
		pascal := ""
		for _, t := range toks {
			pascal += strings.ToUpper(t[:1]) + t[1:]
		}
		out[pascal] = true
		out[strings.ToLower(pascal[:1])+pascal[1:]] = true
	}
	list := make([]string, 0, len(out))
	for v := range out {
		list = append(list, v)
	}
	sort.Strings(list)
	return list
}

func preserved(name string, headFiles []string) bool {
	for _, f := range headFiles {
		segs := strings.Split(f, "/")
		for i, s := range segs {
			if s == name || (i == len(segs)-1 && pathscope.Stem(s) == name) {
				return true
			}
		}
	}
	return false
}

func runRename(r *check.Run) ([]check.Finding, error) {
	if r.Explicit {
		return nil, errors.New("rename needs base and head: run it with --base, or through the scenario canary")
	}
	s, err := pathscope.New(r)
	if err != nil || s.Base == "" {
		return nil, err
	}
	headFiles, err := s.HeadFiles()
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []check.Finding
	for _, e := range s.Entries {
		if e.From == "" {
			continue
		}
		fs, err := entryResidue(s, e, headFiles, seen)
		if err != nil {
			return nil, err
		}
		out = append(out, fs...)
	}
	return out, nil
}

func entryResidue(s *pathscope.Scope, e pathscope.Entry, headFiles []string, seen map[string]bool) ([]check.Finding, error) {
	var out []check.Finding
	for _, rt := range retiredNames(e.From, e.Path) {
		if seen[rt.Old] || preserved(rt.Old, headFiles) {
			continue
		}
		seen[rt.Old] = true
		fs, err := residue(s, rt)
		if err != nil {
			return nil, err
		}
		out = append(out, fs...)
	}
	return out, nil
}

func residue(s *pathscope.Scope, rt retired) ([]check.Finding, error) {
	hits, err := s.Grep(variants(rt.Old))
	if err != nil {
		return nil, err
	}
	var out []check.Finding
	for _, h := range hits {
		out = append(out, check.Finding{File: h.File, Line: h.Line, Rule: ruleRename,
			Problem: fmt.Sprintf("still uses retired name %q (renamed %s -> %s)", rt.Old, rt.From, rt.To),
			Fix:     fmt.Sprintf("replace it with %q in this change, with no alias", rt.New)})
	}
	return out, nil
}

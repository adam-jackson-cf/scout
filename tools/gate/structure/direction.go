package structure

import (
	"fmt"
	"path"
	"strings"

	"scout/tools/gate/check"
	"scout/tools/gate/pathscope"
)

const (
	ruleDirection   = "STRUCT.dependency-direction"
	ruleSibling     = "STRUCT.sibling-isolation"
	ruleReadership  = "STRUCT.tier-readership"
	ruleSingleOwner = "STRUCT.single-owner-placement"
	rulePlacement   = "STRUCT.folder-placement"
)

func direction(c Config, file string, src []byte) ([]check.Finding, error) {
	dir := path.Dir(file)
	edges, err := imports(file, src, c.ImportPackage)
	if err != nil {
		return nil, err
	}
	st := c.tierOf(dir)
	if st < 0 {
		return []check.Finding{{File: file, Line: 1, Rule: ruleDirection,
			Problem: fmt.Sprintf("package %s has no tier in proofs.toml [structure].tiers", dir),
			Fix:     fmt.Sprintf("add %q to the tier its imports allow, or move the file into a tiered package", dir)}}, nil
	}
	var out []check.Finding
	for _, e := range edges {
		if e.Target == dir {
			continue
		}
		tt := c.tierOf(e.Target)
		switch {
		case tt < 0:
			out = append(out, check.Finding{File: file, Line: e.Line, Rule: ruleDirection,
				Problem: fmt.Sprintf("edge %s -> %s targets a package with no tier", dir, e.Target),
				Fix:     fmt.Sprintf("add %q to a tier below %d in proofs.toml [structure].tiers, or drop the import", e.Target, st)})
		case tt > st:
			out = append(out, check.Finding{File: file, Line: e.Line, Rule: ruleDirection,
				Problem: fmt.Sprintf("edge %s (tier %d) -> %s (tier %d) points up the tier order", dir, st, e.Target, tt),
				Fix:     upFix(dir, st)})
		case tt == st:
			out = append(out, check.Finding{File: file, Line: e.Line, Rule: ruleSibling,
				Problem: fmt.Sprintf("edge %s -> %s joins two parts on tier %d", dir, e.Target, st),
				Fix:     siblingFix(st)})
		}
	}
	return out, nil
}

func upFix(dir string, st int) string {
	if st == 0 {
		return fmt.Sprintf("invert the dependency: %s is on the lowest tier and imports no tiered package", dir)
	}
	return fmt.Sprintf("move the code %s needs into a package on tier %d or lower, or invert the dependency", dir, st-1)
}

func siblingFix(st int) string {
	if st == 0 {
		return "merge the two packages, or move the shared part into a new lowest tier both import"
	}
	return fmt.Sprintf("move the part both need into a package on tier %d or lower and import it from there", st-1)
}

func list(xs []string) string {
	if len(xs) == 0 {
		return "none"
	}
	return strings.Join(xs, ", ")
}

func readership(c Config, g *graph, dirs []string) []check.Finding {
	var out []check.Finding
	seenTier := map[int]bool{}
	for _, dir := range dirs {
		t := c.tierOf(dir)
		if t < 0 || t >= c.SharedTiers {
			continue
		}
		file := g.files[dir]
		if file == "" {
			continue
		}
		if !seenTier[t] {
			seenTier[t] = true
			if rs := g.tierReaders(c, t); len(rs) < 2 {
				out = append(out, check.Finding{File: file, Rule: ruleReadership,
					Problem: fmt.Sprintf("shared tier %d (%s) is read by %d part(s): %s", t, list(g.populated(c, t)), len(rs), list(rs)),
					Fix:     "keep these packages inside their one reader until a second part reads the tier, or add the second reader in this change"})
			}
		}
		if rs := sortedKeys(g.readers[dir]); len(rs) == 1 {
			out = append(out, check.Finding{File: file, Rule: ruleSingleOwner,
				Problem: fmt.Sprintf("shared-tier package %s is read only by %s", dir, rs[0]),
				Fix:     fmt.Sprintf("move it under its owner as %s/%s", rs[0], path.Base(dir))})
		}
	}
	return out
}

func runDirection(r *check.Run) ([]check.Finding, error) {
	c, skills, err := load(r)
	if err != nil {
		return nil, err
	}
	s, err := pathscope.New(r)
	if err != nil {
		return nil, err
	}
	out, dirSet, err := scanEntries(c, skills, s, r)
	if err != nil {
		return nil, err
	}
	if len(dirSet) == 0 {
		return out, nil
	}
	g, err := buildGraph(s, c.ImportPackage)
	if err != nil {
		return nil, err
	}
	return append(out, readership(c, g, sortedKeys(dirSet))...), nil
}

func scanEntries(c Config, skills SkillLayout, s *pathscope.Scope, r *check.Run) ([]check.Finding, map[string]bool, error) {
	var out []check.Finding
	dirSet := map[string]bool{}
	for _, e := range s.Entries {
		out = append(out, placement(c, skills, e)...)
		if !goSource(e.Path, s.Fixture != "" || r.Explicit) {
			continue
		}
		src, err := s.Read(e)
		if err != nil {
			return nil, nil, err
		}
		fs, err := direction(c, e.Path, src)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, fs...)
		if !strings.HasSuffix(e.Path, "_test.go") {
			dirSet[path.Dir(e.Path)] = true
		}
	}
	return out, dirSet, nil
}

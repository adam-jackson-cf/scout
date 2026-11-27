package structure

import (
	"fmt"
	"strings"

	"scout/tools/gate/check"
	"scout/tools/gate/pathscope"
)

func placement(c Config, skills SkillLayout, e pathscope.Entry) []check.Finding {
	var out []check.Finding
	if _, ok := c.owner(e.Path); !ok {
		fix := "move it under the owner folder it serves, or add an owner entry for it to [structure].owners"
		if near := c.nearestOwner(e.Path); near != "" {
			fix = fmt.Sprintf("move it under its owner (nearest: %s), or add an owner entry for it to [structure].owners", near)
		}
		out = append(out, check.Finding{File: e.Path, Rule: rulePlacement,
			Problem: "lies outside every owner in proofs.toml [structure].owners", Fix: fix})
	}
	segs := strings.Split(e.Path, "/")
	for i := e.FirstNew; i < len(segs)-1; i++ {
		if pathscope.BannedWords[strings.ToLower(segs[i])] {
			out = append(out, check.Finding{File: e.Path, Rule: rulePlacement,
				Problem: fmt.Sprintf("folder %q is a grab-bag name (house-style banned_words)", segs[i]),
				Fix:     "move the file under the owner it serves, or name the folder for its content"})
		}
	}
	if f, ok := skillPlacement(skills, e.Path); ok {
		out = append(out, f)
	}
	return out
}

func skillPlacement(l SkillLayout, p string) (check.Finding, bool) {
	root := strings.TrimSuffix(l.Root, "/") + "/"
	if l.Root == "" || !strings.HasPrefix(p, root) {
		return check.Finding{}, false
	}
	rest := strings.TrimPrefix(p, root)
	i := strings.Index(rest, "/")
	if i < 0 {
		return check.Finding{File: p, Rule: rulePlacement,
			Problem: "sits directly in the skills root, outside any skill folder",
			Fix:     fmt.Sprintf("move it into %s<skill-id>/ with the skill it belongs to", root)}, true
	}
	inner := rest[i+1:]
	for _, a := range l.Allowed {
		if pathscope.Match(a, inner) {
			return check.Finding{}, false
		}
	}
	return check.Finding{File: p, Rule: rulePlacement,
		Problem: fmt.Sprintf("%q is not part of the skill layout (%s)", inner, strings.Join(l.Allowed, ", ")),
		Fix:     fmt.Sprintf("move it to %s%s/references/<name>.md, assets/ or scripts/", root, rest[:i])}, true
}

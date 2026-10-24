package guardrails

import (
	"path"
	"strings"
)

func eslintSev(v any) (int, bool) {
	if l, ok := v.([]any); ok && len(l) > 0 {
		v = l[0]
	}
	switch tv := v.(type) {
	case string:
		switch tv {
		case "off":
			return 0, true
		case "warn":
			return 1, true
		case "error":
			return 2, true
		}
	case float64:
		if tv >= 0 && tv <= 2 {
			return int(tv), true
		}
	}
	return 0, false
}

// eslintRules folds a rules object into f.sev under prefix.
func eslintRules(f *facts, prefix string, rules map[string]any) {
	if hasOpaque(rules) {
		f.unknown = true
	}
	for name, v := range rules {
		if n, ok := eslintSev(v); ok {
			f.sev[prefix+name] = n
		}
	}
}

func mergeFacts(dst, src *facts) {
	for k := range src.ignores {
		dst.ignores[k] = true
	}
	for k, v := range src.sev {
		dst.sev[k] = v
	}
	for k, v := range src.flags {
		dst.flags[k] = v
	}
	for k, v := range src.nums {
		dst.nums[k] = v
	}
	dst.unknown = dst.unknown || src.unknown
}

func eslintLegacy(t Tree, p string, doc map[string]any, depth int) *facts {
	f := newFacts()
	f.warnLowers = true
	if hasOpaque(doc) {
		f.unknown = true
	}
	for _, ref := range strList(doc["extends"]) {
		var parent *facts
		if pp, ok := resolveRel(p, ref); ok && depth < maxExtendsDepth {
			if pdoc, ok := loadDoc(t, pp); ok {
				if m := asMap(pdoc); m != nil {
					parent = eslintLegacy(t, pp, m, depth+1)
				}
			}
		}
		if parent == nil {
			f.unknown = true // shareable/plugin configs are not resolvable statically
			continue
		}
		mergeFacts(f, parent)
	}
	if _, ok := doc["extends"].([]any); ok {
		for _, e := range asList(doc["extends"]) {
			if _, op := e.(opaque); op {
				f.unknown = true
			}
		}
	}
	if rules := asMap(doc["rules"]); rules != nil {
		eslintRules(f, "", rules)
	}
	f.addIgnores("ignorePatterns", strList(doc["ignorePatterns"]))
	for _, o := range asList(doc["overrides"]) {
		om := asMap(o)
		if om == nil {
			continue
		}
		files := strings.Join(strList(om["files"]), ",")
		if rules := asMap(om["rules"]); rules != nil {
			eslintRules(f, "["+files+"] ", rules)
		}
	}
	return f
}

func extractESLintLegacy(t Tree, p string) map[string]*facts {
	base := path.Base(p)
	if base != "package.json" && !strings.HasPrefix(base, ".eslintrc") {
		return nil
	}
	doc, ok := loadDoc(t, p)
	m := asMap(doc)
	if !ok || m == nil {
		return nil
	}
	if base == "package.json" {
		if m = asMap(m["eslintConfig"]); m == nil {
			return nil
		}
	}
	return map[string]*facts{"eslint": eslintLegacy(t, p, m, 0)}
}

func extractESLintFlat(t Tree, p string) map[string]*facts {
	doc, ok := loadDoc(t, p)
	if !ok {
		return nil
	}
	entries := asList(doc)
	if m := asMap(doc); m != nil {
		entries = []any{m}
	}
	f := newFacts()
	f.warnLowers = true
	for _, e := range entries {
		m := asMap(e)
		if m == nil {
			f.unknown = true // spread / call (shared config): not statically known
			continue
		}
		prefix := ""
		if files := strList(m["files"]); len(files) > 0 {
			prefix = "[" + strings.Join(files, ",") + "] "
		}
		if rules := asMap(m["rules"]); rules != nil {
			eslintRules(f, prefix, rules)
		}
		f.addIgnores("ignores", strList(m["ignores"]))
	}
	return map[string]*facts{"eslint": f}
}

func extractESLintIgnore(t Tree, p string) map[string]*facts {
	b, ok := t.Read(p)
	if !ok {
		return nil
	}
	f := newFacts()
	for _, l := range splitLines(b) {
		if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "#") {
			f.ignores[".eslintignore|"+l] = true
		}
	}
	return map[string]*facts{"eslint": f}
}

// ---- biome ----

func biomeSev(v any) (int, bool) {
	if m := asMap(v); m != nil {
		v = m["level"]
	}
	if s, ok := v.(string); ok {
		switch s {
		case "off":
			return 0, true
		case "info", "warn":
			return 1, true
		case "error", "on":
			return 2, true
		}
	}
	return 0, false
}

func biomeFacts(t Tree, p string, doc map[string]any, depth int) *facts {
	f := newFacts()
	for _, ref := range strList(doc["extends"]) {
		var parent *facts
		if pp, ok := resolveRel(p, ref); ok && depth < maxExtendsDepth {
			if pdoc, ok := loadDoc(t, pp); ok && asMap(pdoc) != nil {
				parent = biomeFacts(t, pp, asMap(pdoc), depth+1)
			}
		}
		if parent == nil {
			f.unknown = true
			continue
		}
		mergeFacts(f, parent)
	}
	linter := asMap(doc["linter"])
	if v, ok := toBool(linter["enabled"]); ok && !v {
		f.ignores["linter|enabled=false"] = true
	}
	for group, rules := range asMap(linter["rules"]) {
		gm := asMap(rules)
		for rule, v := range gm {
			if n, ok := biomeSev(v); ok {
				f.sev[group+"/"+rule] = n
			}
		}
	}
	files := asMap(doc["files"])
	f.addIgnores("files.ignore", strList(files["ignore"]))
	f.addIgnores("linter.ignore", strList(linter["ignore"]))
	for _, set := range []struct {
		name string
		v    any
	}{{"files.includes", files["includes"]}, {"linter.includes", linter["includes"]}} {
		for _, e := range strList(set.v) {
			if strings.HasPrefix(e, "!") {
				f.ignores[set.name+"|"+e] = true
			}
		}
	}
	return f
}

func extractBiome(t Tree, p string) map[string]*facts {
	doc, ok := loadDoc(t, p)
	m := asMap(doc)
	if !ok || m == nil {
		return nil
	}
	return map[string]*facts{"biome": biomeFacts(t, p, m, 0)}
}

// ---- tsconfig ----

type tsEffective struct {
	opts    map[string]any
	exclude []string
	hasExcl bool
	unknown bool
}

func tsResolve(t Tree, p string, doc map[string]any, depth int) tsEffective {
	eff := tsEffective{opts: map[string]any{}}
	exts := strList(doc["extends"])
	for _, ref := range exts {
		var parent tsEffective
		found := false
		if pp, ok := resolveRel(p, ref); ok && depth < maxExtendsDepth {
			for _, cand := range []string{pp, pp + ".json"} {
				if pdoc, ok := loadDoc(t, cand); ok && asMap(pdoc) != nil {
					parent = tsResolve(t, cand, asMap(pdoc), depth+1)
					found = true
					break
				}
			}
		}
		if !found {
			eff.unknown = true // package extends (@tsconfig/...) are not resolvable
			continue
		}
		for k, v := range parent.opts {
			eff.opts[k] = v
		}
		if parent.hasExcl {
			eff.exclude, eff.hasExcl = parent.exclude, true
		}
		eff.unknown = eff.unknown || parent.unknown
	}
	for k, v := range asMap(doc["compilerOptions"]) {
		eff.opts[k] = v
	}
	if v, ok := doc["exclude"]; ok {
		eff.exclude, eff.hasExcl = strList(v), true
	}
	return eff
}

func extractTSConfig(t Tree, p string) map[string]*facts {
	doc, ok := loadDoc(t, p)
	m := asMap(doc)
	if !ok || m == nil {
		return nil
	}
	eff := tsResolve(t, p, m, 0)
	f := newFacts()
	f.unknown = eff.unknown
	f.addIgnores("exclude", eff.exclude)
	strict, strictSet := toBool(eff.opts["strict"])
	if strictSet {
		f.flags["strict"] = strict
	} else if !eff.unknown {
		f.flags["strict"] = false
	}
	for _, k := range []string{"strictNullChecks", "noImplicitAny"} {
		switch v, set := toBool(eff.opts[k]); {
		case set:
			f.flags[k] = v
		case strictSet:
			f.flags[k] = strict
		case !eff.unknown:
			f.flags[k] = false
		}
	}
	if v, set := toBool(eff.opts["noUncheckedIndexedAccess"]); set {
		f.flags["noUncheckedIndexedAccess"] = v
	} else if !eff.unknown {
		f.flags["noUncheckedIndexedAccess"] = false
	}
	return map[string]*facts{"tsconfig": f}
}

// ---- jest ----

func extractJest(t Tree, p string) map[string]*facts {
	base := path.Base(p)
	if base != "package.json" && !strings.HasPrefix(base, "jest.config.") {
		return nil
	}
	doc, ok := loadDoc(t, p)
	m := asMap(doc)
	if !ok || m == nil {
		return nil
	}
	if base == "package.json" {
		if m = asMap(m["jest"]); m == nil {
			return nil
		}
	}
	f := newFacts()
	for scope, v := range asMap(m["coverageThreshold"]) {
		if n, ok := toNum(v); ok {
			if n >= 0 { // negative thresholds are "max uncovered" counts
				f.nums[scope] = n
			}
			continue
		}
		for metric, mv := range asMap(v) {
			if n, ok := toNum(mv); ok && n >= 0 {
				f.nums[scope+"."+metric] = n
			}
		}
	}
	return map[string]*facts{"jest": f}
}

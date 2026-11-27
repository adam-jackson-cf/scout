package guardrails

import (
	"path"
	"strconv"
	"strings"
)

// Python: pyproject.toml, ruff.toml, flake8/mypy/coverage INI, pyright.

func extractPyproject(t Tree, p string) map[string]*facts {
	root := loadTOML(t, p)
	tool := asMap(root["tool"])
	if tool == nil {
		return nil
	}
	out := map[string]*facts{}
	if r := asMap(tool["ruff"]); r != nil {
		out["ruff"] = ruffFacts(t, p, r)
	}
	if m := asMap(tool["mypy"]); m != nil {
		var overrides []map[string]any
		for _, o := range asList(m["overrides"]) {
			if om := asMap(o); om != nil {
				overrides = append(overrides, om)
			}
		}
		out["mypy"] = mypyFacts(m, overrides)
	}
	if pr := asMap(tool["pyright"]); pr != nil {
		out["pyright"] = pyrightFacts(t, p, pr, 0)
	}
	if cov := asMap(lookup(tool, "coverage", "report")); cov != nil {
		out["coverage"] = coverageFacts(cov["fail_under"])
	}
	return out
}

func extractRuffToml(t Tree, p string) map[string]*facts {
	root := loadTOML(t, p)
	if root == nil {
		return nil
	}
	return map[string]*facts{"ruff": ruffFacts(t, p, root)}
}

func coverageFacts(v any) *facts {
	f := newFacts()
	if n, ok := toNum(v); ok {
		f.nums["fail_under"] = n
	}
	return f
}

func ruffFlatten(v any) []string {
	switch tv := v.(type) {
	case map[string]any:
		var out []string
		for _, k := range sortedKeys(tv) {
			for _, code := range strList(tv[k]) {
				out = append(out, k+"=>"+code)
			}
		}
		return out
	default:
		return strList(v)
	}
}

// ruffEffective resolves `extend` chains: replace-style keys (ignore, exclude,
// per-file-ignores) are overridden by the child; extend-* keys accumulate.
func ruffEffective(t Tree, p string, tbl map[string]any, depth int, unknown *bool) map[string][]string {
	res := map[string][]string{}
	if ref, ok := tbl["extend"].(string); ok {
		parent := path.Join(path.Dir(p), ref)
		ptbl := ruffParentTable(t, parent, ref, depth)
		if ptbl == nil {
			*unknown = true
		} else {
			for k, v := range ruffEffective(t, parent, ptbl, depth+1, unknown) {
				res[k] = append([]string(nil), v...)
			}
		}
	}
	lint := asMap(tbl["lint"])
	for _, g := range ruffGroups {
		g.collect(res, []map[string]any{tbl, lint})
	}
	return res
}

func ruffParentTable(t Tree, parent, ref string, depth int) map[string]any {
	resolved := !strings.HasPrefix(parent, "../") && !path.IsAbs(ref)
	if !resolved || depth >= maxExtendsDepth {
		return nil
	}
	root := loadTOML(t, parent)
	if root == nil {
		return nil
	}
	if path.Base(parent) == "pyproject.toml" {
		return asMap(lookup(root, "tool", "ruff"))
	}
	return root
}

type ruffGroup struct {
	name            string
	replace, extend []string
}

var ruffGroups = []ruffGroup{
	{"ignore", []string{"ignore"}, []string{"extend-ignore"}},
	{"exclude", []string{"exclude"}, []string{"extend-exclude"}},
	{"per-file-ignores", []string{"per-file-ignores"}, []string{"extend-per-file-ignores"}},
}

func (g ruffGroup) collect(res map[string][]string, tables []map[string]any) {
	var repl, ext []string
	has := false
	for _, c := range tables {
		for _, k := range g.replace {
			if v, ok := c[k]; ok {
				has = true
				repl = append(repl, ruffFlatten(v)...)
			}
		}
		for _, k := range g.extend {
			if v, ok := c[k]; ok {
				ext = append(ext, ruffFlatten(v)...)
			}
		}
	}
	if has {
		res[g.name] = repl
	}
	res[g.name] = append(res[g.name], ext...)
}

func ruffFacts(t Tree, p string, tbl map[string]any) *facts {
	f := newFacts()
	eff := ruffEffective(t, p, tbl, 0, &f.unknown)
	for group, vals := range eff {
		f.addIgnores(group, vals)
	}
	return f
}

func mypyFacts(top map[string]any, overrides []map[string]any) *facts {
	f := newFacts()
	mypyTopLevel(f, top)
	for _, o := range overrides {
		mypyOverride(f, o)
	}
	return f
}

func mypyTopLevel(f *facts, top map[string]any) {
	strict, strictKnown := toBool(top["strict"])
	if strictKnown {
		f.flags["strict"] = strict
	}
	for _, k := range []string{"disallow_untyped_defs", "check_untyped_defs"} {
		if v, ok := toBool(top[k]); ok {
			f.flags[k] = v
		} else if strictKnown {
			f.flags[k] = strict
		}
	}
	if v, ok := toBool(top["ignore_errors"]); ok && v {
		f.ignores["mypy|ignore_errors"] = true
	}
}

func mypyOverride(f *facts, o map[string]any) {
	mod := strings.Join(strList(o["module"]), ",")
	if v, ok := toBool(o["ignore_errors"]); ok && v {
		f.ignores["overrides|"+mod+" ignore_errors"] = true
	}
	if codes := strList(o["disable_error_code"]); len(codes) > 0 {
		f.ignores["overrides|"+mod+" disable_error_code="+strings.Join(codes, ",")] = true
	}
	for _, k := range []string{"strict", "disallow_untyped_defs", "check_untyped_defs"} {
		if v, ok := toBool(o[k]); ok && !v {
			f.ignores["overrides|"+mod+" "+k+"=false"] = true
		}
	}
}

var pyrightRank = map[string]int{"off": 0, "basic": 1, "standard": 2, "strict": 3}

func pyrightSev(v any) (int, bool) {
	switch tv := v.(type) {
	case bool:
		if tv {
			return 2, true
		}
		return 0, true
	case string:
		switch tv {
		case "error":
			return 2, true
		case "warning", "information":
			return 1, true
		case "none":
			return 0, true
		}
	}
	return 0, false
}

func pyrightFacts(t Tree, p string, cfg map[string]any, depth int) *facts {
	f := newFacts()
	f.warnLowers = true
	if ref, ok := cfg["extends"].(string); ok {
		parent, pcfg := pyrightParent(t, p, ref, depth)
		if pcfg == nil {
			f.unknown = true
		} else {
			pf := pyrightFacts(t, parent, pcfg, depth+1)
			*f = *pf
		}
	}
	pyrightMode(f, cfg)
	pyrightIgnores(f, cfg)
	pyrightReports(f, cfg)
	return f
}

func pyrightParent(t Tree, p, ref string, depth int) (string, map[string]any) {
	parent, resolved := resolveRel(p, ref)
	if !resolved && !strings.HasPrefix(ref, ".") {
		parent, resolved = resolveRel(p, "./"+ref)
	}
	if !resolved || depth >= maxExtendsDepth {
		return parent, nil
	}
	doc, ok := loadDoc(t, parent)
	if !ok {
		return parent, nil
	}
	return parent, asMap(doc)
}

func pyrightMode(f *facts, cfg map[string]any) {
	if s, ok := cfg["typeCheckingMode"].(string); ok {
		if r, ok := pyrightRank[s]; ok {
			f.ranks["typeCheckingMode"] = r
		}
	} else if _, set := f.ranks["typeCheckingMode"]; !set && !f.unknown {
		f.ranks["typeCheckingMode"] = pyrightRank["standard"]
	}
}

func pyrightIgnores(f *facts, cfg map[string]any) {
	for _, key := range []string{"ignore", "exclude"} {
		if v, ok := cfg[key]; ok {
			for k := range f.ignores {
				if strings.HasPrefix(k, key+"|") {
					delete(f.ignores, k)
				}
			}
			f.addIgnores(key, strList(v))
		}
	}
}

func pyrightReports(f *facts, cfg map[string]any) {
	for k, v := range cfg {
		if strings.HasPrefix(k, "report") {
			if n, ok := pyrightSev(v); ok {
				f.sev[k] = n
			}
		}
	}
}

func extractPyrightJSON(t Tree, p string) map[string]*facts {
	doc, ok := loadDoc(t, p)
	m := asMap(doc)
	if !ok || m == nil {
		return nil
	}
	return map[string]*facts{"pyright": pyrightFacts(t, p, m, 0)}
}

// ---- INI: flake8, mypy, coverage ----

func readINI(t Tree, p string) map[string]map[string]string {
	b, ok := t.Read(p)
	if !ok {
		return nil
	}
	return parseINI(b)
}

func extractINIFlake8(t Tree, p string) map[string]*facts {
	sect := readINI(t, p)["flake8"]
	if sect == nil {
		return nil
	}
	f := newFacts()
	for _, key := range []string{"ignore", "extend-ignore", "per-file-ignores", "exclude", "extend-exclude"} {
		if v, ok := sect[key]; ok {
			group := strings.TrimPrefix(key, "extend-")
			f.addIgnores(group, iniList(v))
		}
	}
	return map[string]*facts{"flake8": f}
}

func extractINIMypy(t Tree, p string) map[string]*facts {
	ini := readINI(t, p)
	sect := ini["mypy"]
	if sect == nil {
		return nil
	}
	toMap := func(s map[string]string) map[string]any {
		m := map[string]any{}
		for k, v := range s {
			if b, ok := iniBool(v); ok {
				m[k] = b
			} else {
				m[k] = v
			}
		}
		return m
	}
	var overrides []map[string]any
	for _, name := range sortedKeys(ini) {
		if mod, ok := strings.CutPrefix(name, "mypy-"); ok {
			o := toMap(ini[name])
			o["module"] = mod
			overrides = append(overrides, o)
		}
	}
	return map[string]*facts{"mypy": mypyFacts(toMap(sect), overrides)}
}

func extractINICoverage(t Tree, p string) map[string]*facts {
	ini := readINI(t, p)
	for _, name := range []string{"coverage:report", "report"} {
		if sect, ok := ini[name]; ok {
			if v, ok := sect["fail_under"]; ok {
				f := newFacts()
				if n, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
					f.nums["fail_under"] = n
				}
				return map[string]*facts{"coverage": f}
			}
		}
	}
	return nil
}

package guardrails

import (
	"encoding/json"
	"path"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const maxExtendsDepth = 8

// loadDoc reads a config file as a generic document, choosing the parser by
// extension.
func loadDoc(t Tree, p string) (any, bool) {
	b, ok := t.Read(p)
	if !ok {
		return nil, false
	}
	switch strings.ToLower(path.Ext(p)) {
	case ".yaml", ".yml":
		var v any
		if err := yaml.Unmarshal(b, &v); err != nil || v == nil {
			return nil, false
		}
		return v, true
	case ".js", ".cjs", ".mjs", ".ts", ".mts", ".cts":
		return parseJSConfig(b)
	case ".toml":
		m, err := parseTOML(b)
		return m, err == nil
	default:
		v, err := parseLenient(b)
		return v, err == nil
	}
}

func loadTOML(t Tree, p string) map[string]any {
	b, ok := t.Read(p)
	if !ok {
		return nil
	}
	m, err := parseTOML(b)
	if err != nil {
		return nil
	}
	return m
}

// resolveRel resolves a relative extends reference against the directory of
// the extending file. Package references (bare names) are not resolvable.
func resolveRel(from, ref string) (string, bool) {
	if !strings.HasPrefix(ref, "./") && !strings.HasPrefix(ref, "../") {
		return "", false
	}
	out := path.Join(path.Dir(from), ref)
	if strings.HasPrefix(out, "../") || out == ".." {
		return "", false
	}
	return out, true
}

func canon(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "?"
	}
	return string(b)
}

func hasOpaque(m map[string]any) bool {
	_, ok := m["..."]
	return ok
}

func toBool(v any) (bool, bool) { b, ok := v.(bool); return b, ok }

func toNum(v any) (float64, bool) { f, ok := v.(float64); return f, ok }

func canonList(v any) []string {
	var out []string
	for _, e := range asList(v) {
		if _, op := e.(opaque); op {
			continue
		}
		if s, ok := e.(string); ok {
			out = append(out, s)
		} else {
			out = append(out, canon(e))
		}
	}
	return out
}

// ---- golangci-lint ----

func extractGolangci(t Tree, p string) map[string]*facts {
	doc, ok := loadDoc(t, p)
	m := asMap(doc)
	if !ok || m == nil {
		return nil
	}
	f := newFacts()
	linters := asMap(m["linters"])
	issues := asMap(m["issues"])
	run := asMap(m["run"])
	f.addIgnores("linters.disable", strList(linters["disable"]))
	for _, name := range strList(linters["enable"]) {
		f.enabled["linters.enable|"+name] = true
	}
	if v, ok := toBool(linters["enable-all"]); ok {
		f.flags["linters.enable-all"] = v
	}
	f.addIgnores("issues.exclude-rules", canonList(issues["exclude-rules"]))
	f.addIgnores("issues.exclude-dirs", strList(issues["exclude-dirs"]))
	f.addIgnores("issues.exclude-files", strList(issues["exclude-files"]))
	f.addIgnores("issues.exclude", strList(issues["exclude"]))
	f.addIgnores("run.skip-dirs", strList(run["skip-dirs"]))
	f.addIgnores("run.skip-files", strList(run["skip-files"]))
	excl := asMap(linters["exclusions"])
	f.addIgnores("linters.exclusions.rules", canonList(excl["rules"]))
	f.addIgnores("linters.exclusions.paths", strList(excl["paths"]))
	f.addIgnores("linters.exclusions.presets", strList(excl["presets"]))
	ranks := map[string]int{"all": 3, "standard": 2, "fast": 1, "none": 0}
	if s, ok := linters["default"].(string); ok {
		if r, ok := ranks[s]; ok {
			f.ranks["linters.default"] = r
		}
	} else if v, _ := m["version"].(string); v == "2" || m["version"] == float64(2) {
		f.ranks["linters.default"] = 2
	}
	return map[string]*facts{"golangci": f}
}

// ---- staticcheck ----

func extractStaticcheck(t Tree, p string) map[string]*facts {
	m := loadTOML(t, p)
	if m == nil {
		return nil
	}
	f := newFacts()
	for _, c := range strList(m["checks"]) {
		if strings.HasPrefix(c, "-") {
			f.ignores["checks|"+c] = true
		}
	}
	return map[string]*facts{"staticcheck": f}
}

// ---- Python: pyproject.toml, ruff.toml ----

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
		resolved := !strings.HasPrefix(parent, "../") && !path.IsAbs(ref)
		var ptbl map[string]any
		if resolved && depth < maxExtendsDepth {
			if root := loadTOML(t, parent); root != nil {
				if path.Base(parent) == "pyproject.toml" {
					ptbl = asMap(lookup(root, "tool", "ruff"))
				} else {
					ptbl = root
				}
			}
		}
		if ptbl == nil {
			*unknown = true
		} else {
			for k, v := range ruffEffective(t, parent, ptbl, depth+1, unknown) {
				res[k] = append([]string(nil), v...)
			}
		}
	}
	lint := asMap(tbl["lint"])
	collect := func(group string, replace, extend []string) {
		var repl, ext []string
		has := false
		for _, c := range []map[string]any{tbl, lint} {
			for _, k := range replace {
				if v, ok := c[k]; ok {
					has = true
					repl = append(repl, ruffFlatten(v)...)
				}
			}
			for _, k := range extend {
				if v, ok := c[k]; ok {
					ext = append(ext, ruffFlatten(v)...)
				}
			}
		}
		if has {
			res[group] = repl
		}
		res[group] = append(res[group], ext...)
	}
	collect("ignore", []string{"ignore"}, []string{"extend-ignore"})
	collect("exclude", []string{"exclude"}, []string{"extend-exclude"})
	collect("per-file-ignores", []string{"per-file-ignores"}, []string{"extend-per-file-ignores"})
	return res
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
	for _, o := range overrides {
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
	return f
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
		parent, resolved := resolveRel(p, ref)
		if !resolved && !strings.HasPrefix(ref, ".") {
			parent, resolved = resolveRel(p, "./"+ref)
		}
		var pcfg map[string]any
		if resolved && depth < maxExtendsDepth {
			if doc, ok := loadDoc(t, parent); ok {
				pcfg = asMap(doc)
			}
		}
		if pcfg == nil {
			f.unknown = true
		} else {
			pf := pyrightFacts(t, parent, pcfg, depth+1)
			*f = *pf
		}
	}
	if s, ok := cfg["typeCheckingMode"].(string); ok {
		if r, ok := pyrightRank[s]; ok {
			f.ranks["typeCheckingMode"] = r
		}
	} else if _, set := f.ranks["typeCheckingMode"]; !set && !f.unknown {
		f.ranks["typeCheckingMode"] = pyrightRank["standard"]
	}
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
	for k, v := range cfg {
		if strings.HasPrefix(k, "report") {
			if n, ok := pyrightSev(v); ok {
				f.sev[k] = n
			}
		}
	}
	return f
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

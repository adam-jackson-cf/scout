package guardrails

import "strings"

// golangci-lint and staticcheck.

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

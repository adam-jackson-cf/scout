package guardrails

import (
	"path"
	"slices"
	"sort"
	"strings"
)

type lang int

const (
	langNone lang = iota
	langPython
	langGo
	langTS
)

func langOf(p string) lang {
	switch strings.ToLower(path.Ext(p)) {
	case ".py":
		return langPython
	case ".go":
		return langGo
	case ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", ".mts", ".cts":
		return langTS
	}
	return langNone
}

func isTestPath(p string, l lang) bool {
	base := path.Base(p)
	lower := strings.ToLower(p)
	for _, seg := range strings.Split(path.Dir(lower), "/") {
		if seg == "tests" || seg == "test" || seg == "__tests__" {
			return true
		}
	}
	switch l {
	case langGo:
		return strings.HasSuffix(base, "_test.go")
	case langPython:
		return strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.py") || base == "conftest.py"
	case langTS:
		lb := strings.ToLower(base)
		return strings.Contains(lb, ".test.") || strings.Contains(lb, ".spec.")
	}
	return false
}

func isGenerated(content []byte) bool {
	lines := strings.SplitN(string(content), "\n", 21)
	if len(lines) > 20 {
		lines = lines[:20]
	}
	for _, l := range lines {
		if strings.Contains(l, "@generated") || strings.Contains(l, "DO NOT EDIT") ||
			strings.Contains(l, "Code generated") {
			return true
		}
	}
	return false
}

func splitLines(b []byte) []string {
	if len(b) == 0 {
		return nil
	}
	s := strings.TrimSuffix(string(b), "\n")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSuffix(l, "\r")
	}
	return lines
}

func contains(xs []string, x string) bool { return slices.Contains(xs, x) }

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// scanHeadFiles runs scan over each distinct path in changed (every head path
// when changed is empty) that exists in head, and returns the sorted findings.
func scanHeadFiles(head map[string][]byte, changed []string, scan func(p string, content []byte) []Finding) []Finding {
	paths := changed
	if len(paths) == 0 {
		for p := range head {
			paths = append(paths, p)
		}
	}
	seen := map[string]bool{}
	var out []Finding
	for _, p := range paths {
		if seen[p] {
			continue
		}
		seen[p] = true
		hc, ok := head[p]
		if !ok {
			continue // deleted in head
		}
		out = append(out, scan(p, hc)...)
	}
	sortFindings(out)
	return out
}

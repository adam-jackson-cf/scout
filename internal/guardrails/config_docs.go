package guardrails

import (
	"encoding/json"
	"path"
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

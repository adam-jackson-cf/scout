package pathscope

import (
	"path"
	"strings"
	"unicode"
)

func Match(pattern, p string) bool {
	if pattern == p {
		return true
	}
	if strings.HasSuffix(pattern, "/") {
		return strings.HasPrefix(p, pattern)
	}
	return matchSegs(strings.Split(pattern, "/"), strings.Split(p, "/"))
}

func matchSegs(pat, segs []string) bool {
	for len(pat) > 0 {
		if pat[0] == "**" {
			if len(pat) == 1 {
				return len(segs) > 0
			}
			for i := 0; i <= len(segs); i++ {
				if matchSegs(pat[1:], segs[i:]) {
					return true
				}
			}
			return false
		}
		if len(segs) == 0 {
			return false
		}
		if ok, _ := path.Match(pat[0], segs[0]); !ok {
			return false
		}
		pat, segs = pat[1:], segs[1:]
	}
	return len(segs) == 0
}

func Stem(name string) string {
	if strings.HasPrefix(name, ".") {
		return name
	}
	if i := strings.Index(name, "."); i >= 0 {
		return name[:i]
	}
	return name
}

func Tokens(name string) []string {
	var out []string
	for _, part := range strings.FieldsFunc(name, func(r rune) bool { return r == '_' || r == '-' || r == '.' }) {
		out = append(out, splitCase(part)...)
	}
	for i := range out {
		out[i] = strings.ToLower(out[i])
	}
	return out
}

func splitCase(s string) []string {
	rs := []rune(s)
	var out []string
	start := 0
	for i := 1; i < len(rs); i++ {
		prev, cur := rs[i-1], rs[i]
		lowerToUpper := (unicode.IsLower(prev) || unicode.IsDigit(prev)) && unicode.IsUpper(cur)
		acronym := unicode.IsUpper(prev) && unicode.IsUpper(cur) && i+1 < len(rs) && unicode.IsLower(rs[i+1])
		if lowerToUpper || acronym {
			out = append(out, string(rs[start:i]))
			start = i
		}
	}
	if start < len(rs) {
		out = append(out, string(rs[start:]))
	}
	return out
}

func ContainsSeq(child, parent []string) bool {
	if len(parent) == 0 || len(parent) > len(child) {
		return false
	}
	for i := 0; i+len(parent) <= len(child); i++ {
		ok := true
		for j := range parent {
			if child[i+j] != parent[j] {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

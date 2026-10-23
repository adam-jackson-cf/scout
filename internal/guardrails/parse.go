package guardrails

import (
	"errors"
	"strconv"
	"strings"
)

// opaque marks a value the lenient parser could not evaluate statically
// (identifier, call, spread, ...). Consumers treat it as "unknown".
type opaque struct{}

// ---- lenient JSON / JSONC / JS-object-literal / TOML-value parser ----

type lenient struct {
	s   string
	i   int
	err error
}

var errSyntax = errors.New("guardrails: unparsable config")

func (p *lenient) fail() { p.err = errSyntax }

func (p *lenient) skipSpace() {
	for p.i < len(p.s) {
		c := p.s[p.i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			p.i++
		case strings.HasPrefix(p.s[p.i:], "//"):
			for p.i < len(p.s) && p.s[p.i] != '\n' {
				p.i++
			}
		case strings.HasPrefix(p.s[p.i:], "/*"):
			end := strings.Index(p.s[p.i+2:], "*/")
			if end < 0 {
				p.i = len(p.s)
			} else {
				p.i += end + 4
			}
		default:
			return
		}
	}
}

func isIdent(c byte) bool {
	return c == '_' || c == '$' || c == '-' || c == '+' || c == '.' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func (p *lenient) str() string {
	q := p.s[p.i]
	p.i++
	var b strings.Builder
	for p.i < len(p.s) {
		c := p.s[p.i]
		switch {
		case c == '\\' && p.i+1 < len(p.s):
			n := p.s[p.i+1]
			switch n {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'u':
				if p.i+6 <= len(p.s) {
					if v, err := strconv.ParseUint(p.s[p.i+2:p.i+6], 16, 32); err == nil {
						b.WriteRune(rune(v))
						p.i += 6
						continue
					}
				}
				b.WriteByte(n)
			default:
				b.WriteByte(n)
			}
			p.i += 2
		case c == q:
			p.i++
			return b.String()
		default:
			b.WriteByte(c)
			p.i++
		}
	}
	p.fail()
	return b.String()
}

// skipOpaque advances past a non-literal expression up to a top-level `,` or
// closing bracket.
func (p *lenient) skipOpaque() {
	depth := 0
	for p.i < len(p.s) {
		c := p.s[p.i]
		switch {
		case c == '"' || c == '\'' || c == '`':
			p.str()
			continue
		case strings.HasPrefix(p.s[p.i:], "//") || strings.HasPrefix(p.s[p.i:], "/*"):
			p.skipSpace()
			continue
		case c == '(' || c == '[' || c == '{':
			depth++
		case c == ')' || c == ']' || c == '}':
			if depth == 0 {
				return
			}
			depth--
		case c == ',' && depth == 0:
			return
		}
		p.i++
	}
}

func (p *lenient) value() any {
	p.skipSpace()
	if p.i >= len(p.s) {
		p.fail()
		return nil
	}
	switch c := p.s[p.i]; {
	case c == '{':
		return p.object()
	case c == '[':
		return p.array()
	case c == '"' || c == '\'' || c == '`':
		return p.str()
	case isIdent(c):
		start := p.i
		for p.i < len(p.s) && isIdent(p.s[p.i]) {
			p.i++
		}
		tok := p.s[start:p.i]
		p.skipSpace()
		if p.i < len(p.s) && !strings.ContainsRune(",]}", rune(p.s[p.i])) && p.s[p.i] != '\n' {
			p.skipOpaque()
			return opaque{}
		}
		switch tok {
		case "true":
			return true
		case "false":
			return false
		case "null":
			return nil
		}
		if f, err := strconv.ParseFloat(strings.ReplaceAll(tok, "_", ""), 64); err == nil {
			return f
		}
		return opaque{}
	default:
		p.skipOpaque()
		return opaque{}
	}
}

func (p *lenient) array() []any {
	p.i++ // [
	out := []any{}
	for {
		p.skipSpace()
		if p.i >= len(p.s) {
			p.fail()
			return out
		}
		switch p.s[p.i] {
		case ']':
			p.i++
			return out
		case ',':
			p.i++
			continue
		}
		if strings.HasPrefix(p.s[p.i:], "...") {
			p.skipOpaque()
			out = append(out, opaque{})
			continue
		}
		out = append(out, p.value())
		if p.err != nil {
			return out
		}
	}
}

func (p *lenient) object() map[string]any {
	p.i++ // {
	out := map[string]any{}
	for {
		p.skipSpace()
		if p.i >= len(p.s) {
			p.fail()
			return out
		}
		switch p.s[p.i] {
		case '}':
			p.i++
			return out
		case ',':
			p.i++
			continue
		}
		if strings.HasPrefix(p.s[p.i:], "...") {
			p.skipOpaque()
			out["..."] = opaque{}
			continue
		}
		var key string
		switch c := p.s[p.i]; {
		case c == '"' || c == '\'' || c == '`':
			key = p.str()
		case isIdent(c):
			start := p.i
			for p.i < len(p.s) && isIdent(p.s[p.i]) {
				p.i++
			}
			key = p.s[start:p.i]
		default:
			p.fail()
			return out
		}
		p.skipSpace()
		if p.i >= len(p.s) || (p.s[p.i] != ':' && p.s[p.i] != '=') {
			// shorthand property or method: opaque
			p.skipOpaque()
			out[key] = opaque{}
			continue
		}
		p.i++
		out[key] = p.value()
		if p.err != nil {
			return out
		}
	}
}

// parseLenient parses the first JSON/JS value in src (comments, trailing
// commas, single quotes and unquoted keys tolerated).
func parseLenient(src []byte) (any, error) {
	p := &lenient{s: strings.TrimPrefix(string(src), "\ufeff")}
	v := p.value()
	if p.err != nil {
		return nil, p.err
	}
	if _, ok := v.(opaque); ok {
		return nil, errSyntax
	}
	return v, nil
}

// parseJSConfig finds the exported object/array literal of a JS/TS config
// (`module.exports =`, `export default`, or the first `= {`/`= [`).
func parseJSConfig(src []byte) (any, bool) {
	s := string(src)
	for _, marker := range []string{"module.exports", "export default", "= {", "= ["} {
		for from := 0; ; {
			idx := strings.Index(s[from:], marker)
			if idx < 0 {
				break
			}
			at := from + idx + len(marker)
			if marker == "= {" || marker == "= [" {
				at--
			}
			open := strings.IndexAny(s[at:], "{[")
			from = from + idx + len(marker)
			if open < 0 {
				continue
			}
			p := &lenient{s: s, i: at + open}
			v := p.value()
			if p.err == nil {
				switch v.(type) {
				case map[string]any, []any:
					return v, true
				}
			}
		}
	}
	return nil, false
}

// ---- TOML subset ----

func splitTopLevel(s string, sep byte) []string {
	var parts []string
	var quote byte
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == sep:
			parts = append(parts, strings.TrimSpace(s[start:i]))
			start = i + 1
		}
	}
	return append(parts, strings.TrimSpace(s[start:]))
}

func unquoteKey(k string) string {
	k = strings.TrimSpace(k)
	if len(k) >= 2 && (k[0] == '"' || k[0] == '\'') && k[len(k)-1] == k[0] {
		return k[1 : len(k)-1]
	}
	return k
}

func tomlKeyPath(k string) []string {
	parts := splitTopLevel(k, '.')
	for i := range parts {
		parts[i] = unquoteKey(parts[i])
	}
	return parts
}

// stripTOMLComment removes a trailing `#` comment outside strings.
func stripTOMLComment(line string) string {
	var quote byte
	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '#':
			return line[:i]
		}
	}
	return line
}

// bracketDepth reports the net open brackets of value text outside strings.
func bracketDepth(s string) int {
	var quote byte
	d := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '[' || c == '{':
			d++
		case c == ']' || c == '}':
			d--
		}
	}
	return d
}

func tomlTable(root map[string]any, path []string, appendArray bool) map[string]any {
	cur := root
	for i, k := range path {
		last := i == len(path)-1
		switch next := cur[k].(type) {
		case map[string]any:
			cur = next
		case []any:
			if last && appendArray {
				m := map[string]any{}
				cur[k] = append(next, m)
				return m
			}
			if len(next) > 0 {
				if m, ok := next[len(next)-1].(map[string]any); ok {
					cur = m
					continue
				}
			}
			m := map[string]any{}
			cur[k] = append(next, m)
			cur = m
		default:
			m := map[string]any{}
			if last && appendArray {
				cur[k] = []any{m}
			} else {
				cur[k] = m
			}
			cur = m
		}
	}
	return cur
}

// topLevelIndex returns the index of the first sep outside quotes, or -1.
func topLevelIndex(s string, sep byte) int {
	var quote byte
	for i := range len(s) {
		c := s[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == sep:
			return i
		}
	}
	return -1
}

// parseTOML parses the TOML subset used by tool configs: tables, arrays of
// tables, dotted keys, strings, numbers, booleans, arrays and inline tables.
func parseTOML(src []byte) (map[string]any, error) {
	root := map[string]any{}
	cur := root
	lines := strings.Split(strings.ReplaceAll(string(src), "\r\n", "\n"), "\n")
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(stripTOMLComment(lines[i]))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[[") && strings.HasSuffix(line, "]]") {
			cur = tomlTable(root, tomlKeyPath(line[2:len(line)-2]), true)
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			cur = tomlTable(root, tomlKeyPath(line[1:len(line)-1]), false)
			continue
		}
		eq := topLevelIndex(line, '=')
		if eq < 0 {
			return nil, errSyntax
		}
		key := line[:eq]
		val := strings.TrimSpace(line[eq+1:])
		if strings.HasPrefix(val, `"""`) || strings.HasPrefix(val, `'''`) {
			delim := val[:3]
			text := val[3:]
			for !strings.Contains(text, delim) && i+1 < len(lines) {
				i++
				text += "\n" + lines[i]
			}
			if end := strings.Index(text, delim); end >= 0 {
				text = text[:end]
			}
			setTOML(cur, tomlKeyPath(key), text)
			continue
		}
		for bracketDepth(val) > 0 && i+1 < len(lines) {
			i++
			val += "\n" + stripTOMLComment(lines[i])
		}
		p := &lenient{s: val}
		v := p.value()
		if p.err != nil {
			return nil, p.err
		}
		setTOML(cur, tomlKeyPath(key), v)
	}
	return root, nil
}

func setTOML(tbl map[string]any, path []string, v any) {
	for _, k := range path[:len(path)-1] {
		next, ok := tbl[k].(map[string]any)
		if !ok {
			next = map[string]any{}
			tbl[k] = next
		}
		tbl = next
	}
	tbl[path[len(path)-1]] = v
}

// ---- INI ----

// parseINI returns section -> key -> raw value. Continuation lines (indented)
// join with "\n". Comments start a line with # or ;.
func parseINI(src []byte) map[string]map[string]string {
	out := map[string]map[string]string{}
	var sect map[string]string
	lastKey := ""
	for _, raw := range strings.Split(strings.ReplaceAll(string(src), "\r\n", "\n"), "\n") {
		trim := strings.TrimSpace(raw)
		if trim == "" || strings.HasPrefix(trim, "#") || strings.HasPrefix(trim, ";") {
			continue
		}
		if strings.HasPrefix(trim, "[") && strings.HasSuffix(trim, "]") {
			name := strings.TrimSpace(trim[1 : len(trim)-1])
			sect = out[name]
			if sect == nil {
				sect = map[string]string{}
				out[name] = sect
			}
			lastKey = ""
			continue
		}
		if sect == nil {
			continue
		}
		if (raw[0] == ' ' || raw[0] == '\t') && lastKey != "" {
			sect[lastKey] += "\n" + trim
			continue
		}
		idx := strings.IndexAny(trim, "=:")
		if idx < 0 {
			continue
		}
		lastKey = strings.TrimSpace(trim[:idx])
		sect[lastKey] = strings.TrimSpace(trim[idx+1:])
	}
	return out
}

// iniList splits an INI list value on commas, whitespace and newlines.
func iniList(v string) []string {
	return strings.FieldsFunc(v, func(r rune) bool { return r == ',' || r == '\n' || r == ' ' || r == '\t' })
}

func iniBool(v string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "true", "1", "yes", "on":
		return true, true
	case "false", "0", "no", "off":
		return false, true
	}
	return false, false
}

// ---- value helpers ----

func asMap(v any) map[string]any { m, _ := v.(map[string]any); return m }
func asList(v any) []any         { l, _ := v.([]any); return l }

// strList returns the strings of a string or list-of-strings value.
func strList(v any) []string {
	switch t := v.(type) {
	case string:
		return []string{t}
	case []any:
		var out []string
		for _, e := range t {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func lookup(m map[string]any, path ...string) any {
	var cur any = m
	for _, k := range path {
		mm := asMap(cur)
		if mm == nil {
			return nil
		}
		cur = mm[k]
	}
	return cur
}

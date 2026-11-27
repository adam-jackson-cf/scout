package codeintegrity

import (
	"bytes"
	"strings"
)

type comment struct {
	Line    int
	EndLine int
	Raw     string
	Body    string
	Start   int
	Stop    int
}

type syntax struct {
	line        []string
	block       [][2]string
	quotes      string
	multiline   string
	escape      bool
	rawSingle   bool
	lineStart   bool
	spaceBefore bool
	yaml        bool
	markdown    bool
}

type lexer struct {
	src       []byte
	sx        syntax
	i         int
	line      int
	lineStart int
	out       []comment
	quoted    [][2]int
}

func lexComments(src []byte, sx syntax) []comment {
	return scan(src, sx).out
}

func scan(src []byte, sx syntax) *lexer {
	lx := &lexer{src: src, sx: sx, line: 1}
	fence := ""
	for lx.i < len(src) {
		if sx.markdown && lx.i == lx.lineStart {
			if marker, ok := fenceAt(src, lx.i); ok || fence != "" {
				fence = nextFence(fence, marker, ok)
				lx.skipLine()
				if lx.i < len(src) {
					lx.newline(lx.i)
					lx.i++
				}
				continue
			}
		}
		lx.step()
	}
	return lx
}

func (lx *lexer) step() {
	c := lx.src[lx.i]
	switch {
	case c == '\n':
		lx.newline(lx.i)
		lx.i++
	case lx.sx.escape && c == '\\':
		lx.i = lx.skipEscape(lx.i)
	case lx.tryBlock():
	case lx.tryLine():
	case lx.sx.markdown && c == '`':
		lx.skipInlineCode()
	case lx.quoteAt():
		lx.skipQuote()
	default:
		lx.i++
	}
}

func (lx *lexer) skipEscape(at int) int {
	if at+1 < len(lx.src) && lx.src[at+1] == '\n' {
		lx.newline(at + 1)
	}
	return at + 2
}

func (lx *lexer) escapesIn(q byte) bool {
	return lx.sx.escape && (!lx.sx.rawSingle || q != '\'')
}

func (lx *lexer) newline(at int) {
	lx.line++
	lx.lineStart = at + 1
}

func (lx *lexer) skipLine() {
	end := bytes.IndexByte(lx.src[lx.i:], '\n')
	if end < 0 {
		lx.i = len(lx.src)
		return
	}
	lx.i += end
}

func (lx *lexer) allowed() bool {
	before := lx.src[lx.lineStart:lx.i]
	if lx.sx.lineStart && len(bytes.TrimLeft(before, " \t")) > 0 {
		return false
	}
	if lx.sx.spaceBefore && len(before) > 0 {
		prev := before[len(before)-1]
		return prev == ' ' || prev == '\t'
	}
	return true
}

func (lx *lexer) tryBlock() bool {
	for _, b := range lx.sx.block {
		if !bytes.HasPrefix(lx.src[lx.i:], []byte(b[0])) || !lx.allowed() {
			continue
		}
		start, line := lx.i, lx.line
		bodyStart := lx.i + len(b[0])
		stop, bodyEnd := len(lx.src), len(lx.src)
		if end := bytes.Index(lx.src[bodyStart:], []byte(b[1])); end >= 0 {
			bodyEnd = bodyStart + end
			stop = bodyEnd + len(b[1])
		}
		for j := start; j < stop; j++ {
			if lx.src[j] == '\n' {
				lx.newline(j)
			}
		}
		lx.out = append(lx.out, comment{
			Line: line, EndLine: lx.line,
			Raw:   string(lx.src[start:stop]),
			Body:  strings.TrimSpace(string(lx.src[bodyStart:bodyEnd])),
			Start: start, Stop: stop,
		})
		lx.i = stop
		return true
	}
	return false
}

func (lx *lexer) tryLine() bool {
	for _, m := range lx.sx.line {
		if !bytes.HasPrefix(lx.src[lx.i:], []byte(m)) || !lx.allowed() {
			continue
		}
		stop := len(lx.src)
		if end := bytes.IndexByte(lx.src[lx.i:], '\n'); end >= 0 {
			stop = lx.i + end
		}
		raw := strings.TrimRight(string(lx.src[lx.i:stop]), "\r")
		lx.out = append(lx.out, comment{
			Line: lx.line, EndLine: lx.line, Raw: raw,
			Body:  strings.TrimSpace(strings.TrimPrefix(raw, m)),
			Start: lx.i, Stop: stop,
		})
		lx.i = stop
		return true
	}
	return false
}

func (lx *lexer) skipInlineCode() {
	end := bytes.IndexAny(lx.src[lx.i+1:], "`\n")
	if end < 0 || lx.src[lx.i+1+end] == '\n' {
		lx.i++
		return
	}
	lx.i += end + 2
}

func (lx *lexer) quoteAt() bool {
	c := lx.src[lx.i]
	if !strings.ContainsRune(lx.sx.quotes, rune(c)) {
		return false
	}
	if !lx.sx.yaml {
		return true
	}
	if c == '\'' && lx.i > lx.lineStart && lx.i+1 < len(lx.src) && isWord(lx.src[lx.i-1]) && isWord(lx.src[lx.i+1]) {
		return false
	}
	prev := bytes.TrimRight(lx.src[lx.lineStart:lx.i], " \t")
	return len(prev) == 0 || strings.ContainsRune(":-[{,?", rune(prev[len(prev)-1]))
}

func isWord(b byte) bool {
	return b == '_' || b >= '0' && b <= '9' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}

func (lx *lexer) skipQuote() {
	start := lx.i
	defer func() { lx.quoted = append(lx.quoted, [2]int{start, lx.i}) }()
	q := lx.src[lx.i]
	j := lx.i + 1
	for j < len(lx.src) {
		ch := lx.src[j]
		switch {
		case ch == '\\' && lx.escapesIn(q):
			j = lx.skipEscape(j)
			continue
		case ch == q:
			lx.i = j + 1
			return
		case ch == '\n':
			if !strings.ContainsRune(lx.sx.multiline, rune(q)) {
				lx.i = j
				return
			}
			lx.newline(j)
		}
		j++
	}
	lx.i = len(lx.src)
}

func fenceAt(src []byte, i int) (string, bool) {
	rest := bytes.TrimLeft(src[i:], " ")
	if len(src[i:])-len(rest) > 3 {
		return "", false
	}
	for _, f := range []string{"```", "~~~"} {
		if bytes.HasPrefix(rest, []byte(f)) {
			n := len(f)
			for n < len(rest) && rest[n] == f[0] {
				n++
			}
			return string(rest[:n]), true
		}
	}
	return "", false
}

func nextFence(open, marker string, isFence bool) string {
	switch {
	case !isFence:
		return open
	case open == "":
		return marker
	case marker[0] == open[0] && len(marker) >= len(open):
		return ""
	}
	return open
}

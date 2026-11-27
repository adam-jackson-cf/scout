package codeintegrity

import (
	"regexp"
	"strings"
)

var directivePrefixes = []string{
	"type: ignore", "type:ignore", "pyright:", "fmt:", "ruff:", "isort:", "pylint:", "eslint-",
	"@ts-expect-error", "@ts-ignore", "@ts-nocheck", "prettier-ignore", "biome-ignore", "nolint",
	"SPDX-License-Identifier:", "markdownlint-",
}

var (
	vimModeline    = regexp.MustCompile(`^(?:vi|vim|Vim|ex):\s*(?:(?:se|set)\s+\S.*|\S*=\S*(?:\s+\S+)*)$`)
	encodingCookie = regexp.MustCompile(`^[ \t\f]*#.*?coding[:=][ \t]*[-\w.]+`)
)

func isDirective(c comment, lang language) bool {
	return isHeaderDirective(c, lang) || isBodyDirective(c.Body)
}

func isHeaderDirective(c comment, lang language) bool {
	if c.Line == 1 && strings.HasPrefix(c.Raw, "#!") {
		return true
	}
	if lang == goSource && (strings.HasPrefix(c.Raw, "//go:") || strings.HasPrefix(c.Raw, "//nolint")) {
		return true
	}
	return c.Line <= 2 && encodingCookie.MatchString(c.Raw)
}

func hasDirectivePrefix(body string) bool {
	for _, p := range directivePrefixes {
		if strings.HasPrefix(body, p) {
			return true
		}
	}
	return false
}

func isBodyDirective(body string) bool {
	if strings.HasPrefix(strings.ToLower(body), "noqa") || hasDirectivePrefix(body) {
		return true
	}
	if strings.HasPrefix(body, "-*-") && strings.Contains(body[3:], "-*-") {
		return true
	}
	return vimModeline.MatchString(body)
}

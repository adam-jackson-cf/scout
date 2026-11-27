package pathscope

import (
	"path"
	"regexp"
	"strings"
)

var BannedWords = map[string]bool{
	"common": true, "compat": true, "fallback": true, "helpers": true, "legacy": true, "manager": true,
	"misc": true, "new": true, "old": true, "stuff": true, "temp": true, "utils": true,
}

var CataloguePlurals = map[string]bool{
	"agents": true, "assets": true, "commands": true, "docs": true, "experiments": true, "extensions": true,
	"fixtures": true, "problems": true, "references": true, "rules": true, "scripts": true, "skills": true,
	"stages": true, "templates": true, "tests": true, "tools": true, "workflows": true,
}

var FixedFolders = map[string]bool{
	"internal": true, "cmd": true, "testdata": true, "vendor": true, "__pycache__": true,
	".devcontainer": true, ".github": true, ".git": true, "__tests__": true, "node_modules": true,
}

var fixedFiles = []string{
	"doc.go", "main.go", "__init__.py", "__main__.py", "conftest.py",
	"Dockerfile", "Dockerfile.*", "Containerfile", "Makefile", "GNUmakefile", "makefile",
	"package.json", "package-lock.json", "npm-shrinkwrap.json", "tsconfig.json", "tsconfig.*.json", "jsconfig.json",
	"pyproject.toml", "go.mod", "go.sum", "AGENTS.md", "CLAUDE.md", "README", "README.md", "LICENSE", "LICENSE.*",
	"COPYING", "devcontainer.json", ".devcontainer.json", ".gitignore", ".gitattributes", ".editorconfig", ".npmrc",
	".nvmrc", ".python-version", ".tool-versions", "uv.lock", "bun.lock", "pnpm-lock.yaml", "yarn.lock", "go.work",
	"go.work.sum", "requirements.txt", "ruff.toml", ".ruff.toml", "mypy.ini", ".mypy.ini", ".importlinter",
	".golangci.yml", ".golangci.yaml", ".golangci.toml", ".golangci.json", ".shellcheckrc", "shellcheckrc",
	"biome.json", "biome.jsonc", "eslint.config.*", ".oxlintrc.json", "oxlint.config.ts", "vitest.config.*",
	".node-version", "index.ts", "index.tsx", "index.js", "index.jsx", "index.mts", "index.cts", "index.mjs",
	"index.cjs", "*.d.ts", "*.d.mts", "*.d.cts",
}

func FixedFile(name string) bool {
	for _, p := range fixedFiles {
		if ok, _ := path.Match(p, name); ok {
			return true
		}
	}
	return false
}

var (
	Lowercase  = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	SnakeCase  = regexp.MustCompile(`^_*[a-z][a-z0-9_]*$`)
	KebabCase  = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)
	VersionTok = regexp.MustCompile(`^v\d+$`)
	attached   = regexp.MustCompile(`.[-_]?v\d+$`)
	TicketTok  = regexp.MustCompile(`^(?:\d+|(?:issue|bug|ticket|pr|gh|jira)\d+)$`)
	VersionTk2 = regexp.MustCompile(`^(?:v\d+|[a-z]+v\d+)$`)
	DateTok    = regexp.MustCompile(`^(?:(?:19|20)\d{2}(?:\d{4})?)$`)
	TrailNum   = regexp.MustCompile(`^[a-z]+\d+$`)
)

var NumberedAllowlist = map[string]bool{
	"base32": true, "base64": true, "float32": true, "float64": true, "http2": true, "http3": true, "int16": true,
	"int32": true, "int64": true, "int8": true, "ipv4": true, "ipv6": true, "md5": true, "oauth2": true, "s3": true,
	"sha1": true, "sha256": true, "sha512": true, "uint8": true, "utf16": true, "utf8": true, "x509": true,
}

var TestNameIgnored = map[string]bool{"test": true, "tests": true, "spec": true}

func VersionToken(name string) bool {
	if VersionTok.MatchString(name) {
		return false
	}
	for _, t := range Tokens(name) {
		if VersionTok.MatchString(t) {
			return true
		}
	}
	return attached.MatchString(strings.ToLower(name))
}

func Plural(seg string) bool {
	if CataloguePlurals[strings.ToLower(seg)] {
		return false
	}
	toks := Tokens(seg)
	if len(toks) == 0 {
		return false
	}
	last := toks[len(toks)-1]
	if len(last) < 3 || !strings.HasSuffix(last, "s") {
		return false
	}
	for _, end := range []string{"ss", "us", "is"} {
		if strings.HasSuffix(last, end) {
			return false
		}
	}
	return true
}

func Singular(seg string) string {
	switch {
	case strings.HasSuffix(seg, "ies"):
		return strings.TrimSuffix(seg, "ies") + "y"
	default:
		return strings.TrimSuffix(seg, "s")
	}
}

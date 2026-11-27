package codeintegrity

import (
	"bytes"
	"path"
	"strings"
)

type language int

const (
	unsupported language = iota
	goSource
	pythonSource
	lexical
)

var (
	hashSyntax   = syntax{line: []string{"#"}, quotes: "\"'", multiline: "\"'", escape: true, rawSingle: true, spaceBefore: true}
	tomlSyntax   = syntax{line: []string{"#"}, quotes: "\"'", multiline: "\"'", escape: true, rawSingle: true}
	yamlSyntax   = syntax{line: []string{"#"}, quotes: "\"'", escape: true, rawSingle: true, spaceBefore: true, yaml: true}
	slashSyntax  = syntax{line: []string{"//"}, block: [][2]string{{"/*", "*/"}}, quotes: "\"'`", multiline: "`", escape: true}
	rustSyntax   = syntax{line: []string{"//"}, block: [][2]string{{"/*", "*/"}}, quotes: "\"", multiline: "\"", escape: true}
	cssSyntax    = syntax{block: [][2]string{{"/*", "*/"}}, quotes: "\"'", escape: true}
	phpSyntax    = syntax{line: []string{"//", "#"}, block: [][2]string{{"/*", "*/"}}, quotes: "\"'`", multiline: "\"'", escape: true}
	sqlSyntax    = syntax{line: []string{"--"}, block: [][2]string{{"/*", "*/"}}, quotes: "'\"", multiline: "'\"", escape: true}
	luaSyntax    = syntax{line: []string{"--"}, block: [][2]string{{"--[[", "]]"}}, quotes: "\"'", escape: true}
	haskSyntax   = syntax{line: []string{"--"}, block: [][2]string{{"{-", "-}"}}, quotes: "\"", escape: true}
	markupSyntax = syntax{block: [][2]string{{"<!--", "-->"}}}
	mdSyntax     = syntax{block: [][2]string{{"<!--", "-->"}}, markdown: true}
	texSyntax    = syntax{line: []string{"%"}, escape: true}
	erlSyntax    = syntax{line: []string{"%"}, quotes: "\"", escape: true}
	lispSyntax   = syntax{line: []string{";"}, block: [][2]string{{"#|", "|#"}}, quotes: "\"", multiline: "\"", escape: true}
	iniSyntax    = syntax{line: []string{";", "#"}, lineStart: true}
	propsSyntax  = syntax{line: []string{"#", "!"}, lineStart: true}
	goModSyntax  = syntax{line: []string{"//"}, quotes: "\"`", multiline: "`", escape: true}
)

var byExtension = map[string]syntax{
	".sh": hashSyntax, ".bash": hashSyntax, ".zsh": hashSyntax, ".ksh": hashSyntax, ".fish": hashSyntax,
	".rb": hashSyntax, ".pl": hashSyntax, ".pm": hashSyntax, ".r": hashSyntax, ".ps1": hashSyntax,
	".tf": phpSyntax, ".tfvars": hashSyntax, ".cmake": hashSyntax, ".mk": hashSyntax, ".nix": hashSyntax,
	".conf": hashSyntax, ".env": hashSyntax, ".dockerfile": hashSyntax, ".gitignore": hashSyntax,
	".toml": tomlSyntax, ".yaml": yamlSyntax, ".yml": yamlSyntax,
	".js": slashSyntax, ".jsx": slashSyntax, ".mjs": slashSyntax, ".cjs": slashSyntax,
	".ts": slashSyntax, ".tsx": slashSyntax, ".mts": slashSyntax, ".cts": slashSyntax,
	".c": slashSyntax, ".h": slashSyntax, ".cc": slashSyntax, ".cpp": slashSyntax, ".cxx": slashSyntax,
	".hpp": slashSyntax, ".hh": slashSyntax, ".m": slashSyntax, ".mm": slashSyntax, ".java": slashSyntax,
	".kt": slashSyntax, ".kts": slashSyntax, ".swift": slashSyntax, ".cs": slashSyntax, ".scala": slashSyntax,
	".dart": slashSyntax, ".groovy": slashSyntax, ".gradle": slashSyntax, ".proto": slashSyntax,
	".jsonc": slashSyntax, ".json5": slashSyntax, ".zig": slashSyntax, ".sol": slashSyntax,
	".rs": rustSyntax, ".css": cssSyntax, ".scss": slashSyntax, ".less": slashSyntax, ".php": phpSyntax,
	".sql": sqlSyntax, ".lua": luaSyntax, ".hs": haskSyntax, ".elm": haskSyntax,
	".html": markupSyntax, ".htm": markupSyntax, ".xml": markupSyntax, ".svg": markupSyntax,
	".xsd": markupSyntax, ".xsl": markupSyntax, ".plist": markupSyntax, ".vue": markupSyntax,
	".svelte": markupSyntax, ".md": mdSyntax, ".markdown": mdSyntax, ".mdx": mdSyntax,
	".tex": texSyntax, ".sty": texSyntax, ".cls": texSyntax, ".erl": erlSyntax, ".hrl": erlSyntax,
	".lisp": lispSyntax, ".lsp": lispSyntax, ".el": lispSyntax, ".clj": lispSyntax, ".cljs": lispSyntax,
	".cljc": lispSyntax, ".edn": lispSyntax, ".scm": lispSyntax, ".ss": lispSyntax, ".rkt": lispSyntax,
	".ini": iniSyntax, ".cfg": iniSyntax, ".properties": propsSyntax,
}

var byName = map[string]syntax{
	"Dockerfile": hashSyntax, "Containerfile": hashSyntax, "Makefile": hashSyntax, "GNUmakefile": hashSyntax,
	"makefile": hashSyntax, "CMakeLists.txt": hashSyntax, "Gemfile": hashSyntax, "Rakefile": hashSyntax,
	".gitignore": hashSyntax, ".gitattributes": hashSyntax, ".dockerignore": hashSyntax,
	".prettierignore": hashSyntax, ".npmrc": iniSyntax, ".shellcheckrc": hashSyntax, ".gitleaksignore": hashSyntax,
	".editorconfig": iniSyntax, ".gitconfig": iniSyntax, ".gitmodules": iniSyntax, ".env": hashSyntax,
	".python-version": hashSyntax, ".tool-versions": hashSyntax, "go.mod": goModSyntax, "go.work": goModSyntax,
}

func classify(name string, data []byte) (language, syntax) {
	base := path.Base(name)
	ext := strings.ToLower(path.Ext(base))
	switch ext {
	case ".go":
		return goSource, syntax{}
	case ".py", ".pyi":
		return pythonSource, syntax{}
	}
	if sx, ok := byName[base]; ok {
		return lexical, sx
	}
	if strings.HasPrefix(base, "Dockerfile.") || strings.HasPrefix(base, ".env.") {
		return lexical, hashSyntax
	}
	if sx, ok := byExtension[ext]; ok {
		return lexical, sx
	}
	if ext == "" {
		return shebangLanguage(data)
	}
	return unsupported, syntax{}
}

func shebangLanguage(data []byte) (language, syntax) {
	if !bytes.HasPrefix(data, []byte("#!")) {
		return unsupported, syntax{}
	}
	first, _, _ := bytes.Cut(data, []byte("\n"))
	switch line := string(first); {
	case strings.Contains(line, "python"):
		return pythonSource, syntax{}
	case strings.Contains(line, "node") || strings.Contains(line, "deno") || strings.Contains(line, "bun"):
		return lexical, slashSyntax
	}
	return lexical, hashSyntax
}

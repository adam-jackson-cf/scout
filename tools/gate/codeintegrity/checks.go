package codeintegrity

import (
	"bytes"
	"embed"
	"fmt"
	"path"
	"sort"
	"strings"
	"sync"

	"scout/tools/gate/check"
	"scout/tools/gate/codeintegrity/repofile"
)

const (
	ruleComments     = "CODE.no-comments"
	rulePlaceholders = "CODE.no-placeholder-bodies"
	ruleMocks        = "CODE.no-product-mocks"
	inspector        = "inspect-python-source.py"
)

//go:embed inspect-python-source.py
var scripts embed.FS

func Checks() []check.Check {
	return []check.Check{
		{Name: "code-comments", Rules: []string{ruleComments}, Run: runComments},
		{Name: "code-placeholders", Rules: []string{rulePlaceholders}, Run: runPlaceholders},
		{Name: "code-mocks", Rules: []string{ruleMocks}, Run: runMocks},
	}
}

type pySpan struct {
	Line    int    `json:"line"`
	EndLine int    `json:"end_line"`
	Text    string `json:"text"`
	Name    string `json:"name"`
	Module  string `json:"module"`
}

type pyReport struct {
	Error        string   `json:"error"`
	ErrorLine    int      `json:"error_line"`
	Comments     []pySpan `json:"comments"`
	Docstrings   []pySpan `json:"docstrings"`
	Placeholders []pySpan `json:"placeholders"`
	Imports      []pySpan `json:"imports"`
}

type pyFile struct {
	Path   string `json:"path"`
	Source []byte `json:"source"`
}

type analysis struct {
	files  []repofile.File
	python map[string]pyReport
}

var (
	cacheMu sync.Mutex
	cache   = map[*check.Run]*analysis{}
)

func judged(name string) bool {
	lang, _ := classify(name, nil)
	return lang != unsupported || path.Ext(name) == ""
}

func analyse(r *check.Run) (*analysis, error) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if a, ok := cache[r]; ok {
		return a, nil
	}
	files, err := repofile.Changed(r, judged)
	if err != nil {
		return nil, err
	}
	a := &analysis{python: map[string]pyReport{}}
	var py []pyFile
	for _, f := range files {
		if bytes.IndexByte(f.Data, 0) >= 0 {
			continue
		}
		a.files = append(a.files, f)
		if lang, _ := classify(f.Path, f.Data); lang == pythonSource {
			py = append(py, pyFile{Path: f.Path, Source: f.Data})
		}
	}
	if len(py) > 0 {
		var resp struct {
			Files map[string]pyReport `json:"files"`
		}
		if err := repofile.RunPython(r, scripts, inspector, map[string]any{"files": py}, &resp); err != nil {
			return nil, err
		}
		a.python = resp.Files
	}
	cache[r] = a
	return a, nil
}

func parseFailure(f repofile.File, rep pyReport, rule string) check.Finding {
	line := rep.ErrorLine
	if line < 1 {
		line = 1
	}
	return check.Finding{
		File: f.Path, Line: line, Rule: rule,
		Problem: "cannot judge Python source: " + rep.Error,
		Fix:     "fix the syntax or encoding error so the file parses",
	}
}

func snippet(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 60 {
		s = s[:57] + "..."
	}
	return s
}

func runComments(r *check.Run) ([]check.Finding, error) {
	a, err := analyse(r)
	if err != nil {
		return nil, err
	}
	docsRequired, err := docstringsRequired(r.Root)
	if err != nil {
		return nil, err
	}
	var out []check.Finding
	for _, f := range a.files {
		found, err := fileComments(f, a.python[f.Path], docsRequired)
		if err != nil {
			return nil, err
		}
		out = append(out, found...)
	}
	return filter(r, "code-comments", out), nil
}

func fileComments(f repofile.File, rep pyReport, docsRequired bool) ([]check.Finding, error) {
	lang, sx := classify(f.Path, f.Data)
	if lang == unsupported {
		return nil, nil
	}
	comments, failure := collectComments(f, rep, lang, sx)
	if failure != nil {
		return []check.Finding{*failure}, nil
	}
	out := commentFindings(f, comments, lang)
	if lang == pythonSource && !docsRequired {
		out = append(out, docstringFindings(f, rep)...)
	}
	return out, nil
}

func collectComments(f repofile.File, rep pyReport, lang language, sx syntax) ([]comment, *check.Finding) {
	switch lang {
	case goSource:
		found, err := goComments(f.Data)
		if err != nil {
			return nil, &check.Finding{
				File: f.Path, Line: 1, Rule: ruleComments,
				Problem: "cannot scan Go source: " + err.Error(),
				Fix:     "fix the Go syntax error so the file scans",
			}
		}
		return found, nil
	case pythonSource:
		if rep.Error != "" {
			failure := parseFailure(f, rep, ruleComments)
			return nil, &failure
		}
		return pythonComments(rep), nil
	}
	return lexComments(f.Data, sx), nil
}

func commentFindings(f repofile.File, comments []comment, lang language) []check.Finding {
	var out []check.Finding
	for _, c := range comments {
		if isDirective(c, lang) {
			continue
		}
		if line, ok := f.AnyAdded(c.Line, c.EndLine); ok {
			out = append(out, check.Finding{
				File: f.Path, Line: line, Rule: ruleComments,
				Problem: fmt.Sprintf("added comment %q", snippet(c.Raw)),
				Fix:     "delete the comment and carry the intent in names, tests or docs; only tool directives may stay",
			})
		}
	}
	return out
}

func pythonComments(rep pyReport) []comment {
	out := make([]comment, 0, len(rep.Comments))
	for _, c := range rep.Comments {
		out = append(out, comment{
			Line: c.Line, EndLine: c.EndLine, Raw: c.Text,
			Body: strings.TrimSpace(strings.TrimPrefix(c.Text, "#")),
		})
	}
	return out
}

func docstringFindings(f repofile.File, rep pyReport) []check.Finding {
	var out []check.Finding
	for _, d := range rep.Docstrings {
		if line, ok := f.AnyAdded(d.Line, d.EndLine); ok {
			out = append(out, check.Finding{
				File: f.Path, Line: line, Rule: ruleComments,
				Problem: "added comment as a docstring",
				Fix:     "delete the docstring; no repository lint rule requires it",
			})
		}
	}
	return out
}

func runPlaceholders(r *check.Run) ([]check.Finding, error) {
	a, err := analyse(r)
	if err != nil {
		return nil, err
	}
	var out []check.Finding
	for _, f := range a.files {
		if lang, _ := classify(f.Path, f.Data); lang != pythonSource || strings.HasSuffix(f.Path, ".pyi") {
			continue
		}
		rep := a.python[f.Path]
		if rep.Error != "" {
			out = append(out, parseFailure(f, rep, rulePlaceholders))
			continue
		}
		for _, p := range rep.Placeholders {
			if line, ok := f.AnyAdded(p.Line, p.EndLine); ok {
				out = append(out, check.Finding{
					File: f.Path, Line: line, Rule: rulePlaceholders,
					Problem: fmt.Sprintf("placeholder body in function %q", p.Name),
					Fix:     "implement the function, delete it, or declare it as a Protocol, abstract or @overload method",
				})
			}
		}
	}
	return filter(r, "code-placeholders", out), nil
}

func filter(r *check.Run, name string, in []check.Finding) []check.Finding {
	var out []check.Finding
	for _, f := range in {
		if r.Proofs != nil && r.Proofs.Excepted(f.Rule, f.File, name) {
			continue
		}
		out = append(out, f)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].File != out[j].File {
			return out[i].File < out[j].File
		}
		return out[i].Line < out[j].Line
	})
	return out
}

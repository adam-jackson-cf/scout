package pyquality

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"scout/tools/gate/check"
	"scout/tools/gate/codeintegrity/repofile"
)

type config struct {
	MaintainedRoots            []string            `toml:"maintained_roots" json:"maintained_roots"`
	ScriptRoots                []string            `toml:"script_roots" json:"script_roots"`
	ScriptNames                []string            `toml:"script_names" json:"script_names"`
	TestRoots                  []string            `toml:"test_roots" json:"test_roots"`
	TestGroups                 []string            `toml:"test_groups" json:"test_groups"`
	GenericNameAllowlist       []string            `toml:"generic_name_allowlist" json:"generic_name_allowlist"`
	ClassMaxLines              int                 `toml:"class_max_lines" json:"class_max_lines"`
	ClassMaxInitArgs           int                 `toml:"class_max_init_args" json:"class_max_init_args"`
	ClassMaxInstanceAttributes int                 `toml:"class_max_instance_attributes" json:"class_max_instance_attributes"`
	SubprocessWrappers         map[string][]string `toml:"subprocess_wrappers" json:"subprocess_wrappers"`
	DynamicImportAdapters      []string            `toml:"dynamic_import_adapters" json:"dynamic_import_adapters"`
	ContractFiles              []string            `toml:"contract_files" json:"contract_files"`
	MaxLoopDepth               int                 `toml:"max_loop_depth" json:"max_loop_depth"`
	FanoutMax                  int                 `toml:"fanout_max" json:"fanout_max"`
	LargeModuleNloc            int                 `toml:"large_module_nloc" json:"large_module_nloc"`
	LargeModuleFanoutMax       int                 `toml:"large_module_fanout_max" json:"large_module_fanout_max"`
}

//go:embed *.py
var scripts embed.FS

type family struct {
	name     string
	script   string
	rules    []string
	required func(config) map[string]int
}

var families = []family{
	{
		name:   "python-quality",
		script: "check-python-quality.py",
		rules: []string{
			"PYQ100", "PYQ101", "PYQ102", "PYQ103", "PYQ104", "PYQ107", "PYQ200", "PYQ202",
			"PYQ203", "PYQ220", "PYQ221", "PYQ230", "PYQ231", "PYQ299",
		},
		required: func(c config) map[string]int {
			return map[string]int{
				"class_max_lines": c.ClassMaxLines, "class_max_init_args": c.ClassMaxInitArgs,
				"class_max_instance_attributes": c.ClassMaxInstanceAttributes,
			}
		},
	},
	{
		name:     "python-scalability",
		script:   "check-python-scalability.py",
		rules:    []string{"PYS000", "PYS250"},
		required: func(c config) map[string]int { return map[string]int{"max_loop_depth": c.MaxLoopDepth} },
	},
	{
		name:   "python-fanout",
		script: "check-python-fanout.py",
		rules:  []string{"FANOUT.general", "FANOUT.large-module"},
		required: func(c config) map[string]int {
			return map[string]int{
				"fanout_max": c.FanoutMax, "large_module_nloc": c.LargeModuleNloc,
				"large_module_fanout_max": c.LargeModuleFanoutMax,
			}
		},
	},
}

func Checks() []check.Check {
	out := make([]check.Check, 0, len(families))
	for _, f := range families {
		out = append(out, check.Check{Name: f.name, Rules: f.rules, Run: f.run})
	}
	return out
}

func isPython(name string) bool {
	return strings.HasSuffix(name, ".py") || strings.HasSuffix(name, ".pyi")
}

func loadConfig(r *check.Run, f family) (config, error) {
	var c config
	found := false
	if r.Proofs != nil {
		var err error
		if found, err = r.Proofs.Section("python_quality", &c); err != nil {
			return c, fmt.Errorf("proofs.toml [python_quality]: %w", err)
		}
	}
	if !found {
		return c, fmt.Errorf("proofs.toml has no [python_quality] table; %s needs its maintained roots and limits", f.name)
	}
	if len(c.MaintainedRoots) == 0 {
		return c, fmt.Errorf("proofs.toml [python_quality].maintained_roots is empty; name every maintained Python root")
	}
	var missing []string
	for key, v := range f.required(c) {
		if v < 1 {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return c, fmt.Errorf("proofs.toml [python_quality] needs positive %s for %s", strings.Join(missing, ", "), f.name)
	}
	return c, nil
}

type source struct {
	Path   string `json:"path"`
	Source []byte `json:"source"`
}

type result struct {
	Path    string `json:"path"`
	Line    int    `json:"line"`
	Rule    string `json:"rule"`
	Problem string `json:"problem"`
	Fix     string `json:"fix"`
}

func (f family) run(r *check.Run) ([]check.Finding, error) {
	files, err := repofile.Selected(r, isPython)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}
	c, err := loadConfig(r, f)
	if err != nil {
		return nil, err
	}
	sources := make([]source, 0, len(files))
	modules := make([]string, 0, len(files))
	for _, file := range files {
		sources = append(sources, source{Path: file.Path, Source: file.Data})
		modules = append(modules, file.Path)
	}
	var results []result
	request := map[string]any{"config": c, "files": sources, "modules": modules}
	if err := repofile.RunPython(r, scripts, f.script, request, &results); err != nil {
		return nil, err
	}
	return f.findings(r, results), nil
}

func (f family) findings(r *check.Run, results []result) []check.Finding {
	var out []check.Finding
	for _, res := range results {
		if r.Proofs != nil && r.Proofs.Excepted(res.Rule, res.Path, f.name) {
			continue
		}
		out = append(out, check.Finding{File: res.Path, Line: res.Line, Rule: res.Rule, Problem: res.Problem, Fix: res.Fix})
	}
	return out
}

package guardrails

import (
	"sort"
	"strings"
	"testing"
)

// changesBetween runs every mechanism on a base→head fixture; changed is the
// set of paths whose content differs.
func changesBetween(base, head map[string]string) []Finding {
	toTree := func(m map[string]string) MapTree {
		t := MapTree{}
		for k, v := range m {
			t[k] = []byte(v)
		}
		return t
	}
	var changed []string
	for p, v := range head {
		if b, ok := base[p]; !ok || b != v {
			changed = append(changed, p)
		}
	}
	sort.Strings(changed)
	return ChangesAll(toTree(base), toTree(head), changed)
}

func checkIDs(fs []Finding) []string {
	seen := map[string]bool{}
	for _, f := range fs {
		seen[f.CheckID] = true
	}
	return sortedKeys(seen)
}

type guardCase struct {
	name string
	base map[string]string
	head map[string]string
	// want lists exact check ids expected (as a set); empty means no
	// unscoped/blocking-eligible finding at all.
	want []string
	// path, when set, is the file every finding must be reported on.
	path string
}

func runGuardCases(t *testing.T, cases []guardCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var fs []Finding
			for _, f := range changesBetween(c.base, c.head) {
				if f.CheckID == ScopedCheckID {
					continue // report-only; asserted separately
				}
				fs = append(fs, f)
			}
			got := checkIDs(fs)
			want := append([]string(nil), c.want...)
			sort.Strings(want)
			if strings.Join(got, ",") != strings.Join(want, ",") {
				t.Fatalf("check ids = %v, want %v\nfindings: %+v", got, want, fs)
			}
			for _, f := range fs {
				if c.path != "" && f.Path != c.path {
					t.Fatalf("finding on %q, want %q: %+v", f.Path, c.path, f)
				}
				if f.Evidence.Head == "" {
					t.Fatalf("missing head evidence: %+v", f)
				}
				if strings.HasPrefix(f.CheckID, "guardrails.CONFIG.") && f.Evidence.Base == "" {
					t.Fatalf("config finding without base evidence: %+v", f)
				}
				if !strings.HasSuffix(f.Version, "@1") || !strings.HasPrefix(f.Version, f.CheckID) {
					t.Fatalf("version %q does not extend check id %q", f.Version, f.CheckID)
				}
			}
		})
	}
}

func ign(tool string) string    { return IgnoresPrefix + tool }
func strict(tool string) string { return StrictnessPrefix + tool }

func TestGoGuardrails(t *testing.T) {
	runGuardCases(t, []guardCase{
		{
			name: "golangci linters.disable grows",
			base: map[string]string{".golangci.yml": "linters:\n  disable:\n    - lll\n"},
			head: map[string]string{".golangci.yml": "linters:\n  disable:\n    - lll\n    - errcheck\n"},
			want: []string{ign("golangci")}, path: ".golangci.yml",
		},
		{
			name: "golangci exclude-rules and exclude-dirs grow",
			base: map[string]string{".golangci.yaml": "issues:\n  exclude-dirs: [gen]\n"},
			head: map[string]string{".golangci.yaml": "issues:\n  exclude-dirs: [gen, internal]\n  exclude-rules:\n    - path: x.go\n      linters: [gosec]\n"},
			want: []string{ign("golangci")},
		},
		{
			name: "golangci enabling a linter is not erosion",
			base: map[string]string{".golangci.yml": "linters:\n  enable:\n    - gosec\n  disable:\n    - lll\n"},
			head: map[string]string{".golangci.yml": "linters:\n  enable:\n    - gosec\n    - errcheck\n  disable:\n    - lll\n"},
		},
		{
			name: "golangci linters.enable shrinks",
			base: map[string]string{".golangci.yml": "linters:\n  enable:\n    - gosec\n    - errcheck\n"},
			head: map[string]string{".golangci.yml": "linters:\n  enable:\n    - errcheck\n"},
			want: []string{strict("golangci")},
		},
		{
			name: "golangci enable-all true to false",
			base: map[string]string{".golangci.yml": "linters:\n  enable-all: true\n"},
			head: map[string]string{".golangci.yml": "linters:\n  enable-all: false\n"},
			want: []string{strict("golangci")},
		},
		{
			name: "golangci v2 default all to standard",
			base: map[string]string{".golangci.yml": "version: \"2\"\nlinters:\n  default: all\n"},
			head: map[string]string{".golangci.yml": "version: \"2\"\nlinters:\n  default: standard\n"},
			want: []string{strict("golangci")},
		},
		{
			name: "golangci new config has no baseline",
			base: map[string]string{"main.go": "package main\n"},
			head: map[string]string{"main.go": "package main\n", ".golangci.yml": "linters:\n  disable:\n    - lll\n"},
		},
		{
			name: "staticcheck check disabled",
			base: map[string]string{"staticcheck.conf": "checks = [\"all\"]\n"},
			head: map[string]string{"staticcheck.conf": "checks = [\"all\", \"-SA1019\"]\n"},
			want: []string{ign("staticcheck")},
		},
		{
			name: "staticcheck unrelated change",
			base: map[string]string{"staticcheck.conf": "checks = [\"all\"]\n"},
			head: map[string]string{"staticcheck.conf": "checks = [\"all\", \"ST1000\"]\n"},
		},
		{
			name: "bare nolint added",
			base: map[string]string{"a.go": "package a\n\nfunc f() error { return nil }\n"},
			head: map[string]string{"a.go": "package a\n\nfunc f() error { return nil } //nolint\n"},
			want: []string{UnscopedCheckID},
		},
		{
			name: "nolint:all added",
			base: map[string]string{"a.go": "package a\n\nfunc f() {}\n"},
			head: map[string]string{"a.go": "package a\n\nfunc f() {} //nolint:all\n"},
			want: []string{UnscopedCheckID},
		},
		{
			name: "scoped nolint is not unscoped",
			base: map[string]string{"a.go": "package a\n\nfunc f() {}\n"},
			head: map[string]string{"a.go": "package a\n\nfunc f() {} //nolint:errcheck // best effort close\n"},
		},
		{
			name: "pre-existing nolint moved",
			base: map[string]string{"a.go": "package a\n\nfunc f() {} //nolint\n"},
			head: map[string]string{"a.go": "package a\n\n// moved\nfunc f() {} //nolint\n"},
		},
		{
			name: "Go has no focused-test construct",
			base: map[string]string{"a_test.go": "package a\n"},
			head: map[string]string{"a_test.go": "package a\n\nfunc TestX(t *testing.T) { t.Skip() }\n"},
		},
	})
}

func TestPythonGuardrails(t *testing.T) {
	runGuardCases(t, []guardCase{
		{
			name: "ruff pyproject ignore grows",
			base: map[string]string{"pyproject.toml": "[tool.ruff.lint]\nignore = [\"E501\"]\n"},
			head: map[string]string{"pyproject.toml": "[tool.ruff.lint]\nignore = [\"E501\", \"F401\"]\n"},
			want: []string{ign("ruff")}, path: "pyproject.toml",
		},
		{
			name: "ruff per-file-ignores and exclude grow",
			base: map[string]string{"ruff.toml": "exclude = [\"build\"]\n"},
			head: map[string]string{"ruff.toml": "exclude = [\"build\", \"src\"]\n\n[lint.per-file-ignores]\n\"api/*.py\" = [\"S101\"]\n"},
			want: []string{ign("ruff")},
		},
		{
			name: "ruff selecting more rules is not erosion",
			base: map[string]string{"ruff.toml": "[lint]\nselect = [\"E\"]\nignore = [\"E501\"]\n"},
			head: map[string]string{"ruff.toml": "[lint]\nselect = [\"E\", \"F\"]\nignore = [\"E501\"]\n"},
		},
		{
			name: "ruff extend chain: parent adds ignore, child file unchanged",
			base: map[string]string{
				"ruff.toml":      "extend = \"ruff.base.toml\"\n",
				"ruff.base.toml": "[lint]\nignore = [\"E501\"]\n",
				"svc/ruff.toml":  "extend = \"../ruff.toml\"\n",
			},
			head: map[string]string{
				"ruff.toml":      "extend = \"ruff.base.toml\"\n",
				"ruff.base.toml": "[lint]\nignore = [\"E501\"]\nextend-ignore = [\"D100\"]\n",
				"svc/ruff.toml":  "extend = \"../ruff.toml\"\n",
			},
			want: []string{ign("ruff")},
		},
		{
			name: "ruff extend chain: parent change that only tightens",
			base: map[string]string{"ruff.toml": "extend = \"ruff.base.toml\"\n", "ruff.base.toml": "[lint]\nignore = [\"E501\", \"D100\"]\n"},
			head: map[string]string{"ruff.toml": "extend = \"ruff.base.toml\"\n", "ruff.base.toml": "[lint]\nignore = [\"E501\"]\n"},
		},
		{
			name: "flake8 ignore grows",
			base: map[string]string{".flake8": "[flake8]\nignore = E501\n"},
			head: map[string]string{".flake8": "[flake8]\nignore = E501,W503\n"},
			want: []string{ign("flake8")},
		},
		{
			name: "mypy strict true to false",
			base: map[string]string{"pyproject.toml": "[tool.mypy]\nstrict = true\n"},
			head: map[string]string{"pyproject.toml": "[tool.mypy]\nstrict = false\n"},
			want: []string{strict("mypy")},
		},
		{
			name: "mypy strict kept, one flag lowered",
			base: map[string]string{"mypy.ini": "[mypy]\nstrict = True\n"},
			head: map[string]string{"mypy.ini": "[mypy]\nstrict = True\ndisallow_untyped_defs = False\n"},
			want: []string{strict("mypy")},
		},
		{
			name: "mypy ignore_errors override added",
			base: map[string]string{"pyproject.toml": "[tool.mypy]\nstrict = true\n"},
			head: map[string]string{"pyproject.toml": "[tool.mypy]\nstrict = true\n\n[[tool.mypy.overrides]]\nmodule = \"legacy.*\"\nignore_errors = true\n"},
			want: []string{ign("mypy")},
		},
		{
			name: "mypy ignore_missing_imports override is benign",
			base: map[string]string{"pyproject.toml": "[tool.mypy]\nstrict = true\n"},
			head: map[string]string{"pyproject.toml": "[tool.mypy]\nstrict = true\n\n[[tool.mypy.overrides]]\nmodule = \"requests.*\"\nignore_missing_imports = true\n"},
		},
		{
			name: "pyright strict to basic",
			base: map[string]string{"pyproject.toml": "[tool.pyright]\ntypeCheckingMode = \"strict\"\n"},
			head: map[string]string{"pyproject.toml": "[tool.pyright]\ntypeCheckingMode = \"basic\"\n"},
			want: []string{strict("pyright")},
		},
		{
			name: "pyright extends chain lowers mode",
			base: map[string]string{"pyrightconfig.json": "{\"extends\": \"./pr.base.json\"}", "pr.base.json": "{\"typeCheckingMode\": \"strict\"}"},
			head: map[string]string{"pyrightconfig.json": "{\"extends\": \"./pr.base.json\"}", "pr.base.json": "{\"typeCheckingMode\": \"standard\"}"},
			want: []string{strict("pyright")},
		},
		{
			name: "pyright unchanged mode",
			base: map[string]string{"pyproject.toml": "[tool.pyright]\ntypeCheckingMode = \"strict\"\n"},
			head: map[string]string{"pyproject.toml": "[tool.pyright]\ntypeCheckingMode = \"strict\"\ninclude = [\"src\"]\n"},
		},
		{
			name: "coverage fail_under lowered",
			base: map[string]string{"pyproject.toml": "[tool.coverage.report]\nfail_under = 90\n"},
			head: map[string]string{"pyproject.toml": "[tool.coverage.report]\nfail_under = 80\n"},
			want: []string{strict("coverage")},
		},
		{
			name: "coverage fail_under raised",
			base: map[string]string{"pyproject.toml": "[tool.coverage.report]\nfail_under = 90\n"},
			head: map[string]string{"pyproject.toml": "[tool.coverage.report]\nfail_under = 95\n"},
		},
		{
			name: "bare noqa added",
			base: map[string]string{"a.py": "x = 1\n"},
			head: map[string]string{"a.py": "x = 1\ny = 2  # noqa\n"},
			want: []string{UnscopedCheckID},
		},
		{
			name: "bare type: ignore added",
			base: map[string]string{"a.py": "x: int = 1\n"},
			head: map[string]string{"a.py": "x: int = 'a'  # type: ignore\n"},
			want: []string{UnscopedCheckID},
		},
		{
			name: "scoped noqa and type ignore",
			base: map[string]string{"a.py": "x = 1\n"},
			head: map[string]string{"a.py": "x = 1\ny = 2  # noqa: E501  long url\nz: int = 'a'  # type: ignore[assignment]\n"},
		},
		{
			name: "Python has no focused-test construct",
			base: map[string]string{"test_a.py": "def test_a():\n    pass\n"},
			head: map[string]string{"test_a.py": "def test_a():\n    pass\n\n\ndef test_b():\n    pass  # only\n"},
		},
	})
}

func TestTypeScriptGuardrails(t *testing.T) {
	runGuardCases(t, []guardCase{
		{
			name: "eslintrc rule turned off",
			base: map[string]string{".eslintrc.json": `{"rules": {"eqeqeq": "error", "no-console": "warn"}}`},
			head: map[string]string{".eslintrc.json": `{"rules": {"eqeqeq": "off", "no-console": "warn"}}`},
			want: []string{ign("eslint")}, path: ".eslintrc.json",
		},
		{
			name: "eslintrc rule set to 0",
			base: map[string]string{".eslintrc.json": `{"rules": {"eqeqeq": [2, "always"]}}`},
			head: map[string]string{".eslintrc.json": `{"rules": {"eqeqeq": 0}}`},
			want: []string{ign("eslint")},
		},
		{
			name: "eslintrc error to warn",
			base: map[string]string{".eslintrc.json": `{"rules": {"eqeqeq": "error"}}`},
			head: map[string]string{".eslintrc.json": `{"rules": {"eqeqeq": "warn"}}`},
			want: []string{strict("eslint")},
		},
		{
			name: "eslintrc stays error with new options",
			base: map[string]string{".eslintrc.json": `{"rules": {"eqeqeq": "error"}}`},
			head: map[string]string{".eslintrc.json": `{"rules": {"eqeqeq": ["error", "always"]}}`},
		},
		{
			name: "eslintrc ignorePatterns grows",
			base: map[string]string{".eslintrc.json": `{"ignorePatterns": ["dist"]}`},
			head: map[string]string{".eslintrc.json": `{"ignorePatterns": ["dist", "src/legacy"]}`},
			want: []string{ign("eslint")},
		},
		{
			name: "eslintrc extends chain: child turns parent rule off",
			base: map[string]string{".eslintrc.json": `{"extends": "./eslint.base.json"}`, "eslint.base.json": `{"rules": {"eqeqeq": "error"}}`},
			head: map[string]string{".eslintrc.json": `{"extends": "./eslint.base.json", "rules": {"eqeqeq": "off"}}`, "eslint.base.json": `{"rules": {"eqeqeq": "error"}}`},
			want: []string{ign("eslint")},
		},
		{
			name: "eslintrc extends chain: parent weakened, child untouched",
			base: map[string]string{".eslintrc.json": `{"extends": ["./eslint.base.json"]}`, "eslint.base.json": `{"rules": {"eqeqeq": "error"}}`},
			head: map[string]string{".eslintrc.json": `{"extends": ["./eslint.base.json"]}`, "eslint.base.json": `{"rules": {"eqeqeq": "warn"}}`},
			want: []string{strict("eslint")}, path: ".eslintrc.json",
		},
		{
			name: "eslintrc extends chain: unrelated parent change",
			base: map[string]string{".eslintrc.json": `{"extends": "./eslint.base.json"}`, "eslint.base.json": `{"rules": {"eqeqeq": "error"}}`},
			head: map[string]string{".eslintrc.json": `{"extends": "./eslint.base.json"}`, "eslint.base.json": `{"rules": {"eqeqeq": "error", "semi": "off"}}`},
		},
		{
			name: "eslintrc.cjs rule off",
			base: map[string]string{".eslintrc.cjs": "module.exports = {\n  extends: ['eslint:recommended'],\n  rules: { 'no-undef': 'error' },\n};\n"},
			head: map[string]string{".eslintrc.cjs": "module.exports = {\n  extends: ['eslint:recommended'],\n  rules: { 'no-undef': 'off' },\n};\n"},
			want: []string{ign("eslint")},
		},
		{
			name: "eslint flat config rule off",
			base: map[string]string{"eslint.config.js": "export default [\n  { rules: { \"no-console\": \"error\" } },\n];\n"},
			head: map[string]string{"eslint.config.js": "export default [\n  { rules: { \"no-console\": \"off\" } },\n];\n"},
			want: []string{ign("eslint")},
		},
		{
			name: "eslint flat config ignores grow",
			base: map[string]string{"eslint.config.mjs": "export default [{ ignores: ['dist'] }];\n"},
			head: map[string]string{"eslint.config.mjs": "export default [{ ignores: ['dist', 'src/gen'] }];\n"},
			want: []string{ign("eslint")},
		},
		{
			name: ".eslintignore grows",
			base: map[string]string{".eslintignore": "dist\n", ".eslintrc.json": `{}`},
			head: map[string]string{".eslintignore": "dist\nsrc/legacy\n", ".eslintrc.json": `{}`},
			want: []string{ign("eslint")},
		},
		{
			name: "package.json eslintConfig rule off",
			base: map[string]string{"package.json": `{"eslintConfig": {"rules": {"eqeqeq": "error"}}}`},
			head: map[string]string{"package.json": `{"eslintConfig": {"rules": {"eqeqeq": "off"}}}`},
			want: []string{ign("eslint")},
		},
		{
			name: "biome rule off",
			base: map[string]string{"biome.json": `{"linter": {"rules": {"suspicious": {"noExplicitAny": "error"}}}}`},
			head: map[string]string{"biome.json": `{"linter": {"rules": {"suspicious": {"noExplicitAny": "off"}}}}`},
			want: []string{ign("biome")},
		},
		{
			name: "biome rule stays on",
			base: map[string]string{"biome.json": `{"linter": {"rules": {"suspicious": {"noExplicitAny": "error"}}}}`},
			head: map[string]string{"biome.json": `{"linter": {"rules": {"suspicious": {"noExplicitAny": {"level": "error"}}}}}`},
		},
		{
			name: "tsconfig strict true to false",
			base: map[string]string{"tsconfig.json": "{\n  // comment\n  \"compilerOptions\": {\"strict\": true,},\n}"},
			head: map[string]string{"tsconfig.json": "{\n  \"compilerOptions\": {\"strict\": false}\n}"},
			want: []string{strict("tsconfig")}, path: "tsconfig.json",
		},
		{
			name: "tsconfig strict kept, noImplicitAny lowered",
			base: map[string]string{"tsconfig.json": `{"compilerOptions": {"strict": true}}`},
			head: map[string]string{"tsconfig.json": `{"compilerOptions": {"strict": true, "noImplicitAny": false}}`},
			want: []string{strict("tsconfig")},
		},
		{
			name: "tsconfig extends chain: base strict dropped, child untouched",
			base: map[string]string{"tsconfig.json": `{"extends": "./base.tsconfig.json"}`, "base.tsconfig.json": `{"compilerOptions": {"strict": true}}`},
			head: map[string]string{"tsconfig.json": `{"extends": "./base.tsconfig.json"}`, "base.tsconfig.json": `{"compilerOptions": {"strict": false}}`},
			want: []string{strict("tsconfig")}, path: "tsconfig.json",
		},
		{
			name: "tsconfig extends chain: child overrides parent, still strict",
			base: map[string]string{"tsconfig.json": `{"extends": "./base.tsconfig.json"}`, "base.tsconfig.json": `{"compilerOptions": {"strict": true}}`},
			head: map[string]string{"tsconfig.json": `{"extends": "./base.tsconfig.json", "compilerOptions": {"strict": true}}`, "base.tsconfig.json": `{"compilerOptions": {"strict": false}}`},
		},
		{
			name: "tsconfig unresolvable package extends is unknown, not lowered",
			base: map[string]string{"tsconfig.json": `{"extends": "@tsconfig/strictest/tsconfig.json"}`},
			head: map[string]string{"tsconfig.json": `{"extends": "@tsconfig/strictest/tsconfig.json", "include": ["src"]}`},
		},
		{
			name: "tsconfig exclude grows",
			base: map[string]string{"tsconfig.json": `{"exclude": ["dist"]}`},
			head: map[string]string{"tsconfig.json": `{"exclude": ["dist", "src/legacy"]}`},
			want: []string{ign("tsconfig")},
		},
		{
			name: "tsconfig include grows",
			base: map[string]string{"tsconfig.json": `{"include": ["src"], "exclude": ["dist"]}`},
			head: map[string]string{"tsconfig.json": `{"include": ["src", "scripts"], "exclude": ["dist"]}`},
		},
		{
			name: "jest coverageThreshold lowered",
			base: map[string]string{"jest.config.js": "module.exports = { coverageThreshold: { global: { branches: 80, lines: 90 } } };\n"},
			head: map[string]string{"jest.config.js": "module.exports = { coverageThreshold: { global: { branches: 70, lines: 90 } } };\n"},
			want: []string{strict("jest")},
		},
		{
			name: "jest coverageThreshold raised (package.json)",
			base: map[string]string{"package.json": `{"jest": {"coverageThreshold": {"global": {"lines": 80}}}}`},
			head: map[string]string{"package.json": `{"jest": {"coverageThreshold": {"global": {"lines": 85}}}}`},
		},
		{
			name: "jest coverageThreshold lowered (package.json)",
			base: map[string]string{"package.json": `{"jest": {"coverageThreshold": {"global": {"lines": 80}}}}`},
			head: map[string]string{"package.json": `{"jest": {"coverageThreshold": {"global": {"lines": 50}}}}`},
			want: []string{strict("jest")},
		},
		{
			name: "ts-nocheck added",
			base: map[string]string{"a.ts": "export const a = 1;\n"},
			head: map[string]string{"a.ts": "// @ts-nocheck\nexport const a = 1;\n"},
			want: []string{UnscopedCheckID},
		},
		{
			name: "bare ts-ignore and eslint-disable added",
			base: map[string]string{"a.ts": "export const a = 1;\n"},
			head: map[string]string{"a.ts": "/* eslint-disable */\n// @ts-ignore\nexport const a: number = 'x';\n"},
			want: []string{UnscopedCheckID},
		},
		{
			name: "scoped eslint-disable-next-line with reason",
			base: map[string]string{"a.ts": "export const a = 1;\n"},
			head: map[string]string{"a.ts": "// eslint-disable-next-line no-console -- debug hook\nconsole.log(a);\nexport const a = 1;\n"},
		},
		{
			name: "it.only added in spec",
			base: map[string]string{"a.test.ts": "describe('a', () => {\n  it('works', () => {});\n});\n"},
			head: map[string]string{"a.test.ts": "describe('a', () => {\n  it.only('works', () => {});\n});\n"},
			want: []string{FocusedCheckID},
		},
		{
			name: "fit and fdescribe and test.describe.only",
			base: map[string]string{"a.spec.js": "x();\n"},
			head: map[string]string{"a.spec.js": "x();\nfit('a', () => {});\nfdescribe('b', () => {});\ntest.describe.only('c', () => {});\n"},
			want: []string{FocusedCheckID},
		},
		{
			name: "focus markers that are not focusing",
			base: map[string]string{"a.test.ts": "x();\n"},
			head: map[string]string{"a.test.ts": "x();\n// it.only('later', () => {});\nconst s = \"describe.only(\";\nit.skip('a', () => {});\nbenefit('b');\nobj.fit('c');\n"},
		},
		{
			name: "pre-existing it.only is not added",
			base: map[string]string{"a.test.ts": "it.only('a', () => {});\n"},
			head: map[string]string{"a.test.ts": "// header\nit.only('a', () => {});\n"},
		},
		{
			name: "it.only in a non-test file is not a focused test",
			base: map[string]string{"helper.ts": "export {};\n"},
			head: map[string]string{"helper.ts": "export {};\nit.only('a', () => {});\n"},
		},
	})
}

func TestFocusedFindingCarriesBaseAndHeadEvidence(t *testing.T) {
	fs := changesBetween(
		map[string]string{"a.test.ts": "it('works', () => {});\n"},
		map[string]string{"a.test.ts": "it.only('works', () => {});\n"},
	)
	if len(fs) != 1 || fs[0].Evidence.Base != "it('works', () => {});" || fs[0].Evidence.Head != "it.only('works', () => {});" || fs[0].Line != 1 {
		t.Fatalf("findings: %+v", fs)
	}
}

func TestConfigFindingEvidenceAndLine(t *testing.T) {
	fs := changesBetween(
		map[string]string{"tsconfig.json": "{\n  \"compilerOptions\": {\n    \"strict\": true\n  }\n}\n"},
		map[string]string{"tsconfig.json": "{\n  \"compilerOptions\": {\n    \"strict\": false\n  }\n}\n"},
	)
	var found bool
	for _, f := range fs {
		if f.Mechanism == "flag:tsconfig:strict" {
			found = true
			if f.Evidence.Base != "strict = true" || f.Evidence.Head != "strict = false" || f.Line != 3 {
				t.Fatalf("finding %+v", f)
			}
		}
	}
	if !found {
		t.Fatalf("strict transition missing: %+v", fs)
	}
}

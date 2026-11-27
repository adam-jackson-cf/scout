package tooling

import (
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

const goodMypy = `[mypy]
python_version = 3.11
strict = True
follow_imports = normal
files = src, tools/run.py
mypy_path = src
`

func mypyFindings(t *testing.T, ini string, files []string, evidence map[string]string) []string {
	t.Helper()
	got := judgeMypyConfig(parseINI(ini), files, "3.11", evidence, func(string) bool { return true })
	var out []string
	for _, f := range got {
		out = append(out, f.Rule+" "+f.File+" "+f.Problem)
	}
	return out
}

func TestMypyConfigCoveringEveryRootPasses(t *testing.T) {
	if got := mypyFindings(t, goodMypy, []string{"src/a.py", "tools/run.py"}, nil); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestMypyConfigRejectsEachPolicyBreak(t *testing.T) {
	cases := []struct {
		name, ini string
		files     []string
		want      string
	}{
		{"uncovered file", goodMypy, []string{"other/b.py"}, "MYPY.package-coverage other/b.py"},
		{"in-repo stub", goodMypy, []string{"src/lib.pyi"}, "MYPY.follow_imports src/lib.pyi"},
		{"python version drift", strings.Replace(goodMypy, "3.11", "3.12", 1), nil, "MYPY.SETTING.python_version"},
		{"global follow_imports silent", strings.Replace(goodMypy, "normal", "silent", 1), nil, "MYPY.follow_imports mypy.ini follow_imports"},
		{"mypy_path drift", strings.Replace(goodMypy, "mypy_path = src", "mypy_path = lib", 1), nil, "MYPY.SETTING.mypy_path"},
		{"inert exclude", goodMypy + "exclude = ^evaluation/\n", nil, "MYPY.SETTING.exclude"},
		{"global ignore_missing_imports", goodMypy + "ignore_missing_imports = True\n", nil, "MYPY.module-overrides mypy.ini global"},
		{"unevidenced override", goodMypy + "\n[mypy-lizard]\nfollow_untyped_imports = True\n", nil, "MYPY.module-overrides mypy.ini [mypy-lizard] has no evidence"},
		{"skip override", goodMypy + "\n[mypy-x]\nfollow_imports = skip\n", nil, "MYPY.follow_imports mypy.ini [mypy-x] follow_imports = skip"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, g := range mypyFindings(t, tc.ini, tc.files, nil) {
				if strings.HasPrefix(g, tc.want) {
					return
				}
			}
			t.Fatalf("want %q in %v", tc.want, mypyFindings(t, tc.ini, tc.files, nil))
		})
	}
}

func TestEvidencedMypyOverridePasses(t *testing.T) {
	ini := goodMypy + "\n[mypy-lizard]\nfollow_untyped_imports = True\n"
	if got := mypyFindings(t, ini, nil, map[string]string{"lizard": "no py.typed and no published stubs"}); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}

func decodeRuff(t *testing.T, text string) ruffIgnores {
	t.Helper()
	var c ruffIgnores
	if _, err := toml.Decode(text, &c); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestRuffIgnoreAddedWithoutRationaleIsReported(t *testing.T) {
	base := decodeRuff(t, "[lint]\nignore = [\"E501\"]\n")
	head := "[lint]\nignore = [\"E501\", \"F401\"]\n[lint.per-file-ignores]\n\"x.py\" = [\"B006\"]\n"
	got := addedIgnores(base, decodeRuff(t, head), head)
	if len(got) != 2 || got[0].Rule != "RUFF.IGNORES" || !strings.Contains(got[0].Problem+got[1].Problem, "F401 for all files") || !strings.Contains(got[0].Problem+got[1].Problem, "B006 for x.py") {
		t.Fatalf("got %+v", got)
	}
}

func TestRuffIgnoreAddedWithSameLineRationalePasses(t *testing.T) {
	base := decodeRuff(t, "[lint]\nignore = []\n")
	head := "[lint]\nignore = [\n  \"F401\", # re-exported API, approved by owner 2026-11-27\n]\n"
	if got := addedIgnores(base, decodeRuff(t, head), head); len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestCoveragePragmaMatchesOnlyComments(t *testing.T) {
	src := "package p\n\nfunc F() { // coverage-ignore\n}\n\nconst p = \"// coverage-ignore\"\n\n/* coverage-ignore */\n"
	got := pragmaLines([]byte(src))
	if len(got) != 1 || got[0] != 3 {
		t.Fatalf("want only the line-3 comment, got %v", got)
	}
}

func TestAddedLinesFollowHunkNumbers(t *testing.T) {
	diff := "+++ b/a.go\n@@ -1,0 +3,2 @@\n+x\n+y\n+++ b/evaluation/z.go\n@@ -0,0 +1 @@\n+z\n"
	got := addedLines(diff, layout{Exclude: []string{"evaluation/**"}})
	if len(got) != 1 || !got["a.go"][3] || !got["a.go"][4] || got["a.go"][5] {
		t.Fatalf("got %v", got)
	}
}

func TestCoverageFloorChangesAreDetected(t *testing.T) {
	base, err := parseFloor("threshold:\n  total: 80\noverride:\n  - threshold: 85\n    path: ^a$\n")
	if err != nil {
		t.Fatal(err)
	}
	same, _ := parseFloor("threshold:\n  total: 81\noverride:\n  - threshold: 90\n    path: ^a$\n")
	if got := floorChanges(base, same); len(got) != 0 {
		t.Fatalf("raising the floor reported %v", got)
	}
	lower, _ := parseFloor("threshold:\n  total: 70\nexclude:\n  paths: [x]\n")
	got := floorChanges(base, lower)
	if len(got) != 3 {
		t.Fatalf("want lowered total, dropped override and changed exclusions, got %v", got)
	}
}

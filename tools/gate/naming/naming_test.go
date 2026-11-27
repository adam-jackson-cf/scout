package naming

import (
	"strings"
	"testing"

	"scout/tools/gate/check"
	"scout/tools/gate/pathscope"
)

func newJudge() judge {
	return judge{seen: map[string]bool{}, lay: layout{Skills: skillLayout{Root: ".agents/skills"}},
		excepted: func(string, string, string) bool { return false }, isDir: func(string) bool { return false }}
}

func problems(fs []check.Finding) string {
	var out []string
	for _, f := range fs {
		out = append(out, f.Rule+" "+f.Problem)
	}
	return strings.Join(out, "\n")
}

func TestChildRepeatingParentIsRejected(t *testing.T) {
	fs := newJudge().pathFindings(pathscope.Entry{Path: "internal/audit/audit_log.go", FirstNew: 1})
	if !strings.Contains(problems(fs), "repeats its parent") {
		t.Fatalf("got %s", problems(fs))
	}
}

func TestGoPackageFilesMayRepeatParent(t *testing.T) {
	for _, p := range []string{"internal/areas/areas.go", "internal/areas/areas_test.go"} {
		if fs := newJudge().pathFindings(pathscope.Entry{Path: p, FirstNew: 2}); strings.Contains(problems(fs), "repeats") {
			t.Fatalf("%s: %s", p, problems(fs))
		}
	}
}

func TestOnlyNewSegmentsAreJudged(t *testing.T) {
	fs := newJudge().pathFindings(pathscope.Entry{Path: "internal/analyzers/batch.go", FirstNew: 2})
	if len(fs) != 0 {
		t.Fatalf("legacy plural parent rejudged: %s", problems(fs))
	}
	fs = newJudge().pathFindings(pathscope.Entry{Path: "internal/widgets/batch.go", FirstNew: 1})
	if !strings.Contains(problems(fs), "plural") {
		t.Fatalf("new plural folder accepted: %s", problems(fs))
	}
}

func TestDirectoryFindingsAreReportedOnce(t *testing.T) {
	j := newJudge()
	a := j.pathFindings(pathscope.Entry{Path: "internal/widgets/a.go", FirstNew: 1})
	b := j.pathFindings(pathscope.Entry{Path: "internal/widgets/b.go", FirstNew: 1})
	if len(a) != 1 || len(b) != 0 || a[0].File != "internal/widgets" {
		t.Fatalf("got %v then %v", a, b)
	}
}

func TestVersionBannedAndCasingRules(t *testing.T) {
	j := newJudge()
	got := problems(j.pathFindings(pathscope.Entry{Path: "internal/parser/utils.go", FirstNew: 2}))
	if !strings.Contains(got, "grab-bag") {
		t.Errorf("banned stem accepted: %s", got)
	}
	got = problems(j.pathFindings(pathscope.Entry{Path: "internal/parser/parse_v2.go", FirstNew: 2}))
	if !strings.Contains(got, "version token") {
		t.Errorf("version token accepted: %s", got)
	}
	got = problems(j.pathFindings(pathscope.Entry{Path: "internal/BadPkg/a.go", FirstNew: 1}))
	if !strings.Contains(got, "casing") {
		t.Errorf("uppercase Go folder accepted: %s", got)
	}
	got = problems(j.pathFindings(pathscope.Entry{Path: ".agents/skills/Bad_Skill/SKILL.md", FirstNew: 2}))
	if !strings.Contains(got, "casing") {
		t.Errorf("non-kebab skill accepted: %s", got)
	}
	got = problems(j.pathFindings(pathscope.Entry{Path: "README.md"}))
	if got != "" {
		t.Errorf("fixed file judged: %s", got)
	}
}

func TestTestFileTokensAndNearMisses(t *testing.T) {
	j := newJudge()
	for p, want := range map[string]string{
		"internal/a/entry_issue42_test.go": "ticket",
		"internal/a/entry_v2_test.go":      "version",
		"internal/a/entry_2026_test.go":    "date",
		"internal/a/test_entry.go":         "go test only reads",
		"tools/testentry.py":               "matches no test pattern",
	} {
		if got := problems(j.testFile(pathscope.Entry{Path: p})); !strings.Contains(got, want) {
			t.Errorf("%s: got %q, want %q", p, got, want)
		}
	}
	for _, p := range []string{"internal/a/entry_test.go", "internal/a/sha256_test.go", "tools/test_entry_order.py"} {
		if got := problems(j.testFile(pathscope.Entry{Path: p})); got != "" {
			t.Errorf("%s: %s", p, got)
		}
	}
}

func TestGoIdentifiersUseMixedCaps(t *testing.T) {
	fs, err := goSymbols("a.go", []byte("package a\nfunc do_thing() {}\nconst Max_Size = 1\nvar _ = 1\n"))
	if err != nil || len(fs) != 2 || fs[0].Fix != "rename it to MixedCaps \"doThing\"" {
		t.Fatalf("got %v %v", fs, err)
	}
	fs, _ = goSymbols("a_test.go", []byte("package a\nfunc Test_order(t *T) {}\nfunc Example_run() {}\n"))
	if len(fs) != 0 {
		t.Fatalf("test entry points flagged: %v", fs)
	}
}

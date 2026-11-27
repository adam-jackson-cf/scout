package structure

import (
	"strings"
	"testing"

	"scout/tools/gate/check"
	"scout/tools/gate/pathscope"
)

func config() Config {
	return Config{
		ImportPackage: "scout", SharedTiers: 1, PrefixClusterMin: 3,
		Tiers:  [][]string{{"internal/domain", "internal/redact"}, {"internal/engine", "tools/gate/*"}, {"cmd/scout"}},
		Owners: map[string]string{"internal/domain/": "domain", "cmd/scout/": "cli", "README.md": "repo", ".agents/skills/": "skills"},
	}
}

func rules(fs []check.Finding) string {
	var out []string
	for _, f := range fs {
		out = append(out, f.Rule)
	}
	return strings.Join(out, ",")
}

func TestDirectionRejectsUpwardEdge(t *testing.T) {
	fs, err := direction(config(), "internal/domain/a.go", []byte("package domain\nimport _ \"scout/cmd/scout\"\n"))
	if err != nil || rules(fs) != ruleDirection || fs[0].Line != 2 {
		t.Fatalf("got %v %v", fs, err)
	}
}

func TestDirectionRejectsSameTierSibling(t *testing.T) {
	fs, _ := direction(config(), "internal/domain/a.go", []byte("package domain\nimport _ \"scout/internal/redact\"\n"))
	if rules(fs) != ruleSibling {
		t.Fatalf("got %v", fs)
	}
}

func TestDirectionAcceptsDownwardEdgeAndExternalImports(t *testing.T) {
	fs, _ := direction(config(), "cmd/scout/main.go", []byte("package main\nimport (\n\"fmt\"\n\"scout/internal/domain\"\n)\n"))
	if len(fs) != 0 {
		t.Fatalf("got %v", fs)
	}
}

func TestDirectionReportsUntieredPackage(t *testing.T) {
	fs, _ := direction(config(), "internal/misc/a.go", []byte("package misc\n"))
	if rules(fs) != ruleDirection || !strings.Contains(fs[0].Problem, "no tier") {
		t.Fatalf("got %v", fs)
	}
}

func TestTierGlobMatchesOnlyDirectChildren(t *testing.T) {
	c := config()
	if c.tierOf("tools/gate/naming") != 1 || c.tierOf("tools/gate/basediff/scenario") != -1 {
		t.Fatal("glob tier lookup is wrong")
	}
}

func TestReadershipNeedsTwoReaders(t *testing.T) {
	g := &graph{
		readers: map[string]map[string]bool{"internal/domain": {"cmd/scout": true}},
		files:   map[string]string{"internal/domain": "internal/domain/a.go", "cmd/scout": "cmd/scout/main.go"},
	}
	fs := readership(config(), g, []string{"internal/domain"})
	if rules(fs) != ruleReadership+","+ruleSingleOwner {
		t.Fatalf("got %v", fs)
	}
	g.readers["internal/domain"]["internal/engine"] = true
	if fs := readership(config(), g, []string{"internal/domain"}); len(fs) != 0 {
		t.Fatalf("two readers still reported %v", fs)
	}
}

func TestPlacementRejectsUnownedAndGrabBagPaths(t *testing.T) {
	fs := placement(config(), SkillLayout{}, pathscope.Entry{Path: "misc/notes.md"})
	if rules(fs) != rulePlacement+","+rulePlacement {
		t.Fatalf("got %v", fs)
	}
	if fs := placement(config(), SkillLayout{}, pathscope.Entry{Path: "README.md"}); len(fs) != 0 {
		t.Fatalf("owned root file rejected: %v", fs)
	}
}

func TestPlacementJudgesOnlyNewFolders(t *testing.T) {
	c := config()
	c.Owners["internal/"] = "internal"
	e := pathscope.Entry{Path: "internal/common/a.go", FirstNew: 2}
	if fs := placement(c, SkillLayout{}, e); len(fs) != 0 {
		t.Fatalf("legacy grab-bag parent rejudged: %v", fs)
	}
}

func TestSkillLayoutAcceptsCreateSkillShape(t *testing.T) {
	l := SkillLayout{Root: ".agents/skills", Allowed: []string{"SKILL.md", "README.md", "references/*.md", "assets/**", "scripts/**"}}
	for _, p := range []string{"SKILL.md", "README.md", "references/a.md", "assets/templates/x.md", "scripts/size-check.sh"} {
		if _, bad := skillPlacement(l, ".agents/skills/demo/"+p); bad {
			t.Errorf("rejected %s", p)
		}
	}
	for _, p := range []string{".agents/skills/demo/notes.txt", ".agents/skills/demo/references/deep/x.md", ".agents/skills/stray.md"} {
		if _, bad := skillPlacement(l, p); !bad {
			t.Errorf("accepted %s", p)
		}
	}
}

func TestClustersBlockNewPrefixMember(t *testing.T) {
	base := children([]string{"s/audit_log.go", "s/audit_trail.go", "s/audit_log_test.go"})
	head := children([]string{"s/audit_log.go", "s/audit_trail.go", "s/audit_log_test.go", "s/audit_report.go"})
	none := func(string) bool { return false }
	fs := clusters(3, base, head, []member{{Path: "s/audit_report.go", Parent: "s", Name: "audit_report.go"}}, none)
	if rules(fs) != rulePrefix {
		t.Fatalf("got %v", fs)
	}
	if fs := clusters(3, base, head, []member{{Path: "s/audit_report.go", Parent: "s", Name: "audit_report.go"}}, func(string) bool { return true }); len(fs) != 0 {
		t.Fatal("exception ignored")
	}
}

func TestClustersIgnoreTestOfExistingMember(t *testing.T) {
	base := children([]string{"s/audit_log.go", "s/audit_trail.go", "s/audit_report.go"})
	head := children([]string{"s/audit_log.go", "s/audit_trail.go", "s/audit_report.go", "s/audit_log_test.go"})
	fs := clusters(3, base, head, []member{{Path: "s/audit_log_test.go", Parent: "s", Name: "audit_log_test.go"}}, func(string) bool { return false })
	if len(fs) != 0 {
		t.Fatalf("test for an existing member blocked: %v", fs)
	}
}

func TestRetiredNamesTakeWholeRenamedSegments(t *testing.T) {
	got := retiredNames("internal/gate_effort/run.go", "internal/gate_budget/run.go")
	if len(got) != 1 || got[0].Old != "gate_effort" || got[0].New != "gate_budget" {
		t.Fatalf("got %v", got)
	}
	if got := retiredNames("a/old_name.go", "a/new_name.go"); len(got) != 1 || got[0].Old != "old_name" {
		t.Fatalf("file stem not retired: %v", got)
	}
}

func TestSeparatorOnlyRenameRetiresNoName(t *testing.T) {
	if got := retiredNames("gen/make_fixtures.py", "gen/make-fixtures.py"); len(got) != 0 {
		t.Fatalf("separator-only rename retired %v", got)
	}
}

func TestVariantsCoverIdentifierForms(t *testing.T) {
	got := strings.Join(variants("audit_log"), ",")
	for _, v := range []string{"AuditLog", "auditLog", "audit-log", "audit_log"} {
		if !strings.Contains(got, v) {
			t.Errorf("missing %s in %s", v, got)
		}
	}
	if len(variants("audit")) != 1 {
		t.Error("single token name widened")
	}
}

func TestPreservedWhenHeadStillOwnsTheName(t *testing.T) {
	if !preserved("gate", []string{"tools/gate/main.go"}) || preserved("gate_effort", []string{"tools/gate/main.go"}) {
		t.Fatal("preservation check is wrong")
	}
}

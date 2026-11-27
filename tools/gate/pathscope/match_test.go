package pathscope

import (
	"reflect"
	"testing"
)

func TestTokensSplitCaseAndSeparators(t *testing.T) {
	cases := map[string][]string{
		"audit_log":   {"audit", "log"},
		"HTTPServer":  {"http", "server"},
		"parserV2":    {"parser", "v2"},
		"gate-effort": {"gate", "effort"},
		"a..b":        {"a", "b"},
	}
	for in, want := range cases {
		if got := Tokens(in); !reflect.DeepEqual(got, want) {
			t.Errorf("Tokens(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestMatchHandlesDoubleStarAndPrefix(t *testing.T) {
	yes := [][2]string{{"evaluation/**", "evaluation/a/b.go"}, {"docs/", "docs/x.md"}, {"tools/gate/*", "tools/gate/naming"}, {"**/x.go", "a/b/x.go"}}
	no := [][2]string{{"evaluation/**", "evaluation"}, {"tools/gate/*", "tools/gate/basediff/scenario"}, {"docs/", "doc/x.md"}}
	for _, c := range yes {
		if !Match(c[0], c[1]) {
			t.Errorf("Match(%q, %q) = false", c[0], c[1])
		}
	}
	for _, c := range no {
		if Match(c[0], c[1]) {
			t.Errorf("Match(%q, %q) = true", c[0], c[1])
		}
	}
}

func TestContainsSeqNeedsWholeOrderedTokens(t *testing.T) {
	if !ContainsSeq(Tokens("log_audit"), Tokens("audit")) {
		t.Error("single-token parent inside child not found")
	}
	if ContainsSeq(Tokens("railways"), Tokens("rail")) {
		t.Error("substring counted as a repeat")
	}
	if ContainsSeq(Tokens("log-audit"), Tokens("audit-log")) {
		t.Error("reordered tokens counted as a repeat")
	}
}

func TestVersionTokenAllowsStandaloneFolders(t *testing.T) {
	for _, bad := range []string{"parser-v2", "parserV2", "e1b-v2", "dev2"} {
		if !VersionToken(bad) {
			t.Errorf("VersionToken(%q) = false", bad)
		}
	}
	for _, ok := range []string{"v2", "level", "vector"} {
		if VersionToken(ok) {
			t.Errorf("VersionToken(%q) = true", ok)
		}
	}
}

func TestPluralSkipsCataloguesAndSibilants(t *testing.T) {
	for _, p := range []string{"analyzers", "decisions"} {
		if !Plural(p) {
			t.Errorf("Plural(%q) = false", p)
		}
	}
	for _, s := range []string{"skills", "tests", "status", "analysis", "class", "gate"} {
		if Plural(s) {
			t.Errorf("Plural(%q) = true", s)
		}
	}
	if Singular("entries") != "entry" || Singular("areas") != "area" {
		t.Error("Singular gave the wrong stem")
	}
}

func TestStemKeepsDotfiles(t *testing.T) {
	if Stem(".gitignore") != ".gitignore" || Stem("a.tar.gz") != "a" {
		t.Error("Stem split the wrong dot")
	}
}

func TestParseGrepReadsHeadPrefixedHits(t *testing.T) {
	hits := parseGrep("HEAD:README.md\x0012\x00uses auditlog\nx.go\x003\x00y\n")
	want := []Hit{{File: "README.md", Line: 12, Text: "uses auditlog"}, {File: "x.go", Line: 3, Text: "y"}}
	if !reflect.DeepEqual(hits, want) {
		t.Errorf("parseGrep = %v", hits)
	}
}

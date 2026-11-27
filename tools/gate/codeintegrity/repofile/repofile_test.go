package repofile

import (
	"reflect"
	"testing"
)

func TestMatchSupportsDoubleStarSegments(t *testing.T) {
	cases := []struct {
		pattern, name string
		want          bool
	}{
		{"evaluation/**", "evaluation/x/y.py", true},
		{"evaluation/**", "evaluationx/y.py", false},
		{"tests/**/test_*.py", "tests/test_a.py", true},
		{"tests/**/test_*.py", "tests/a/b/test_a.py", true},
		{"tests/**/test_*.py", "src/tests/test_a.py", false},
		{"*.md", "a/b.md", false},
		{"**/*.md", "a/b.md", true},
	}
	for _, tc := range cases {
		if got := Match(tc.pattern, tc.name); got != tc.want {
			t.Errorf("Match(%q, %q) = %v", tc.pattern, tc.name, got)
		}
	}
}

func TestParsePatchCollectsAddedLinesPerFile(t *testing.T) {
	patch := "diff --git a/x.go b/x.go\n--- a/x.go\n+++ b/x.go\n@@ -3,0 +4,2 @@\n+a\n+b\n@@ -10 +12 @@\n-c\n+d\n" +
		"diff --git a/gone.go b/gone.go\n--- a/gone.go\n+++ /dev/null\n@@ -1 +0,0 @@\n-x\n" +
		"diff --git a/sp ace.md b/sp ace.md\n--- a/sp ace.md\n+++ \"b/sp ace.md\"\n@@ -1,2 +1,0 @@\n-x\n-y\n"
	got := ParsePatch(patch)
	want := map[string]map[int]bool{"x.go": {4: true, 5: true, 12: true}, "sp ace.md": {}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParsePatchTreatsMovedLinesAsUnchanged(t *testing.T) {
	patch := "diff --git a/old_test.go b/old_test.go\n--- a/old_test.go\n+++ b/old_test.go\n@@ -5,2 +4,0 @@\n-\t// keeps the order stable\n-func helper() {}\n" +
		"diff --git a/new_test.go b/new_test.go\n--- /dev/null\n+++ b/new_test.go\n@@ -0,0 +1,3 @@\n+// keeps the order stable\n+func helper() {}\n+// a new note\n"
	got := ParsePatch(patch)
	want := map[string]map[int]bool{"old_test.go": {}, "new_test.go": {3: true}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestFileAddedLines(t *testing.T) {
	all := File{}
	some := File{Added: map[int]bool{7: true}}
	if l, ok := all.AnyAdded(3, 4); !ok || l != 3 {
		t.Fatal("explicit file should count every line as added")
	}
	if _, ok := some.AnyAdded(1, 6); ok {
		t.Fatal("unchanged span reported as added")
	}
	if l, ok := some.AnyAdded(5, 9); !ok || l != 7 {
		t.Fatal("touched span not reported at its added line")
	}
}

package featuremap

import (
	"errors"
	"strings"
	"testing"
)

const good = "# W\n\n## Sub-features\n\n- `TestWidgetWorks`\n\n## How to get to it (user POV)\n\n## Driving it with `go run ./cmd/scout`\n\n## Gotchas\n"

func judge(t *testing.T, files map[string]string, tests ...string) string {
	t.Helper()
	in := Input{Dir: "docs/features", Index: "README.md", Sections: Sections([]string{"go", "run", "./cmd/scout"}), Files: map[string]bool{}, Tests: map[string]bool{}}
	for p := range files {
		in.Files[p] = true
	}
	for _, n := range tests {
		in.Tests[n] = true
	}
	in.Read = func(p string) ([]byte, error) {
		if s, ok := files[p]; ok {
			return []byte(s), nil
		}
		return nil, errors.New("missing")
	}
	fs, err := Judge(in)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, f := range fs {
		out = append(out, f.Problem)
	}
	return strings.Join(out, "\n")
}

func TestCompliantFeatureMapPasses(t *testing.T) {
	got := judge(t, map[string]string{"docs/features/README.md": "[W](w.md) [t](../../cmd/a_test.go)", "docs/features/w.md": good, "cmd/a_test.go": ""}, "TestWidgetWorksTwice")
	if got != "" {
		t.Fatalf("got %s", got)
	}
}

func TestFeatureMapViolationsAreNamed(t *testing.T) {
	got := judge(t, map[string]string{
		"docs/features/README.md": "[W](w.md) [gone](gone.md)",
		"docs/features/w.md":      "## Sub-features\n\n- `TestWidgetWorks`\n",
		"docs/features/x.md":      good,
	})
	for _, want := range []string{"stale link \"gone.md\"", "H2 sections are", "TestWidgetWorks", "not listed in the index"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %s", want, got)
		}
	}
}

func TestMissingIndexIsRejected(t *testing.T) {
	if got := judge(t, map[string]string{"docs/features/w.md": good}); !strings.Contains(got, "no index") {
		t.Fatalf("got %s", got)
	}
}

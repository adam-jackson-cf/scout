package skill

import (
	"errors"
	"strings"
	"testing"
)

func layoutFor() Layout {
	return Layout{Root: ".agents/skills", Required: []string{"name", "description"}, Allowed: []string{"name", "description", "hide", "disable-model-invocation"}}
}

func judge(t *testing.T, files map[string]string) string {
	t.Helper()
	var paths []string
	for p := range files {
		paths = append(paths, p)
	}
	fs, err := Judge(layoutFor(), paths, func(p string) ([]byte, error) {
		if s, ok := files[p]; ok {
			return []byte(s), nil
		}
		return nil, errors.New("missing " + p)
	})
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, f := range fs {
		out = append(out, f.Problem)
	}
	return strings.Join(out, "\n")
}

func TestCreateSkillShapePasses(t *testing.T) {
	got := judge(t, map[string]string{
		".agents/skills/demo/SKILL.md":            "---\nname: \"demo\"\ndescription: \"USE WHEN demoing.\"\nhide: true\n---\n\nSee [guide](references/guide.md).\n",
		".agents/skills/demo/references/guide.md": "Back to [the skill](../SKILL.md).\n",
	})
	if got != "" {
		t.Fatalf("got %s", got)
	}
}

func TestMetadataViolationsAreNamed(t *testing.T) {
	got := judge(t, map[string]string{
		".agents/skills/demo/SKILL.md": "---\nname: \"other\"\ndescription: \"Writes.\"\nowner: \"me\"\n---\n[x](references/none.md) [y](../../README.md)\n",
	})
	for _, want := range []string{"does not match folder", "does not start with \"use when\"", "outside the skill schema", "missing file", "leaves the skill folder"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %s", want, got)
		}
	}
}

func TestSkillWithoutManifestIsRejected(t *testing.T) {
	if got := judge(t, map[string]string{".agents/skills/demo/README.md": "x"}); !strings.Contains(got, "no SKILL.md") {
		t.Fatalf("got %s", got)
	}
}

package basediff

import (
	"testing"

	"scout/tools/gate/check"
)

func TestGoldenChangeNeedsRecordEntry(t *testing.T) {
	golden := []string{"internal/a/testdata/*.jsonl"}
	cs := []check.Change{{Status: "M", Path: "internal/a/testdata/gold.jsonl"}, {Status: "M", Path: "internal/a/a.go"}}
	if fs := Golden(golden, "CHANGELOG.md", cs, nil); len(fs) != 1 || fs[0].File != "internal/a/testdata/gold.jsonl" {
		t.Fatalf("got %v", fs)
	}
	if fs := Golden(golden, "CHANGELOG.md", cs, []string{"- Regenerate internal/a/testdata/gold.jsonl."}); len(fs) != 0 {
		t.Fatalf("named entry rejected: %v", fs)
	}
}

func TestGoldenRenameChecksBothSides(t *testing.T) {
	cs := []check.Change{{Status: "R", From: "g/old.json", Path: "x/new.json"}}
	if fs := Golden([]string{"g/*.json"}, "CHANGELOG.md", cs, nil); len(fs) != 1 || fs[0].File != "g/old.json" {
		t.Fatalf("got %v", fs)
	}
}

func TestAddedLinesSkipsHeaders(t *testing.T) {
	got := added("+++ b/CHANGELOG.md\n@@ -1 +1 @@\n-old\n+new line\n")
	if len(got) != 1 || got[0] != "new line" {
		t.Fatalf("got %v", got)
	}
}

func TestIdentityArtefactBlocksTestOnlyChange(t *testing.T) {
	art := map[string]string{"scripts/p.patch": "hashed"}
	testOnly := []check.Change{{Status: "M", Path: "scripts/p.patch"}, {Status: "M", Path: "internal/a/a_test.go"}, {Status: "A", Path: "internal/a/testdata/x.json"}}
	if fs := Identity(art, nil, testOnly); len(fs) != 1 {
		t.Fatalf("got %v", fs)
	}
	withProduction := append(testOnly, check.Change{Status: "M", Path: "internal/a/a.go"})
	if fs := Identity(art, nil, withProduction); len(fs) != 0 {
		t.Fatalf("production change still blocked: %v", fs)
	}
	if fs := Identity(art, nil, []check.Change{{Status: "M", Path: "internal/a/a_test.go"}}); len(fs) != 0 {
		t.Fatalf("untouched artefact reported: %v", fs)
	}
}

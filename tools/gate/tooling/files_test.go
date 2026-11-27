package tooling

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShellSelectionUsesExtensionOrShebang(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"hook":      "#!/usr/bin/env bash\nset -eu\n",
		"posix":     "#!/bin/sh\n:\n",
		"script.py": "#!/usr/bin/env python3\n",
		"notes":     "plain text\n",
		"lib.sh":    "f() { :; }\n",
	}
	for name, text := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := explicitFiles(root, []string{"."}, shellScript)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"hook", "lib.sh", "posix"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestLayoutExcludeGlobsDropRoutinePaths(t *testing.T) {
	l := layout{Exclude: []string{"evaluation/**", "tools/gate/testdata/**"}}
	for p, want := range map[string]bool{
		"evaluation/x/a.go":                        true,
		"tools/gate/testdata/canary/ruff/a.py":     true,
		"tools/gate/tooling/files.go":              false,
		"internal/association/testdata/gen/gen.py": false,
	} {
		if got := l.routineExcluded(p); got != want {
			t.Errorf("routineExcluded(%q) = %v, want %v", p, got, want)
		}
	}
}

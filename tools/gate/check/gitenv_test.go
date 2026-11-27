package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func gitVars(env []string) []string {
	var out []string
	for _, kv := range env {
		if strings.HasPrefix(kv, "GIT_") {
			out = append(out, kv)
		}
	}
	return out
}

func TestIsolateGitEnvDropsHookVariablesAndReturnsTheIndex(t *testing.T) {
	t.Setenv("GIT_DIR", "/tmp/elsewhere/.git")
	t.Setenv("GIT_WORK_TREE", "/tmp/elsewhere")
	t.Setenv("GIT_INDEX_FILE", "index.lock")
	t.Setenv("GIT_CEILING_DIRECTORIES", "/tmp")
	index, err := IsolateGitEnv()
	if err != nil {
		t.Fatal(err)
	}
	if got := gitVars(os.Environ()); len(got) != 1 || got[0] != "GIT_CEILING_DIRECTORIES=/tmp" {
		t.Fatalf("child processes keep only the discovery ceiling, got %v", got)
	}
	want, err := filepath.Abs("index.lock")
	if err != nil {
		t.Fatal(err)
	}
	if index != want {
		t.Fatalf("index = %q, want %q", index, want)
	}
	r := &Run{IndexFile: index}
	if got := gitVars(r.GitEnv()); len(got) != 2 || got[1] != "GIT_INDEX_FILE="+want {
		t.Fatalf("gate git calls get %v, want the ceiling and GIT_INDEX_FILE=%s", got, want)
	}
}

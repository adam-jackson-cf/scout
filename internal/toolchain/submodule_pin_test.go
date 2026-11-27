//go:build boundary

package toolchain

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"scout/internal/clawpatch"
)

func boundedTestContext(t testing.TB) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestClawpatchPinMatchesManifestAndSubmodule(t *testing.T) {
	if clawpatch.Version != manifest(t)["CLAWPATCH_VERSION"] {
		t.Errorf("clawpatch.Version %s != CLAWPATCH_VERSION", clawpatch.Version)
	}
	output, err := exec.CommandContext(boundedTestContext(t), "git", "-C", repoRoot, "ls-files", "-s", "third_party/clawpatch").Output()
	if err != nil || len(output) == 0 {
		t.Fatalf("the third_party/clawpatch submodule gitlink cannot be read from a git checkout (the pin check cannot be skipped): %v", err)
	}
	fields := strings.Fields(string(output))
	if len(fields) < 2 || fields[0] != "160000" {
		t.Fatalf("third_party/clawpatch is not a submodule gitlink: %q", output)
	}
	if fields[1] != clawpatch.Revision {
		t.Errorf("submodule gitlink %s != clawpatch.Revision %s (an upgrade needs a compatibility review)", fields[1], clawpatch.Revision)
	}
}

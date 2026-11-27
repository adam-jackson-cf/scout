package measures

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"scout/internal/domain"
)

func singleFunctionPythonModule(name string, ifBranches int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "def %s(x):\n    total = 0\n", name)
	for i := range ifBranches {
		fmt.Fprintf(&b, "    if x == %d:\n        total += %d\n", i, i*7+3)
	}
	b.WriteString("    return total\n")
	return b.String()
}

func TestRunWithoutScannersIsUnmeasuredNotZero(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.py"), []byte(singleFunctionPythonModule("a", 11)), 0o644); err != nil {
		t.Fatal(err)
	}
	result := Run(t.Context(), Config{
		Root: root, Snapshot: domain.Snapshot{WorktreeDigest: "w"}, RunScope: "repository",
		Files: []domain.File{{Path: "a.py", ContentDigest: "d", Eligible: true}},
		Tools: Tools{Lizard: "scout-no-such-lizard", Jscpd: "scout-no-such-jscpd"},
	})
	if result.LizardErr == "" {
		t.Fatal("missing lizard not reported")
	}
	for _, signal := range result.Signals {
		if signal.Missing != domain.SignalUnmeasured || signal.Value != nil {
			t.Fatalf("%s = %+v, want unmeasured", signal.ID(), signal)
		}
	}
	if len(result.Signals) == 0 {
		t.Fatal("no signals recorded for the scope")
	}
}

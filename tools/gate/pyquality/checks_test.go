package pyquality

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"scout/tools/gate/check"
	"scout/tools/gate/proof"
)

func runWith(t *testing.T, body string) *check.Run {
	t.Helper()
	dir := t.TempDir()
	file := filepath.Join(dir, "proofs.toml")
	if err := os.WriteFile(file, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	pf, err := proof.Load(file)
	if err != nil {
		t.Fatal(err)
	}
	return &check.Run{Root: dir, Proofs: pf}
}

func TestConfigRequiresRootsAndEveryLimit(t *testing.T) {
	cases := []struct {
		body    string
		family  int
		message string
	}{
		{"schema = 1\n", 0, "no [python_quality] table"},
		{"[python_quality]\nclass_max_lines = 700\n", 0, "maintained_roots is empty"},
		{"[python_quality]\nmaintained_roots = [\"src\"]\nclass_max_lines = 700\n", 0, "class_max_init_args, class_max_instance_attributes"},
		{"[python_quality]\nmaintained_roots = [\"src\"]\n", 1, "max_loop_depth"},
		{"[python_quality]\nmaintained_roots = [\"src\"]\nfanout_max = 10\n", 2, "large_module_fanout_max, large_module_nloc"},
	}
	for _, tc := range cases {
		_, err := loadConfig(runWith(t, tc.body), families[tc.family])
		if err == nil || !strings.Contains(err.Error(), tc.message) {
			t.Errorf("%q: err = %v, want %q", tc.body, err, tc.message)
		}
	}
	body := "[python_quality]\nmaintained_roots = [\"src\"]\nmax_loop_depth = 2\n"
	c, err := loadConfig(runWith(t, body), families[1])
	if err != nil || c.MaxLoopDepth != 2 {
		t.Fatalf("valid config rejected: %v", err)
	}
}

func TestEveryFamilyNamesItsRules(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range Checks() {
		if len(c.Rules) == 0 || c.Run == nil {
			t.Fatalf("%s has no rules or runner", c.Name)
		}
		for _, r := range c.Rules {
			if seen[r] {
				t.Fatalf("rule %s owned twice", r)
			}
			seen[r] = true
		}
	}
	if len(seen) != 18 {
		t.Fatalf("rules = %d, want 18", len(seen))
	}
}

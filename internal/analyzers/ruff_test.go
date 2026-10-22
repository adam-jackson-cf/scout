package analyzers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRuffInputListMustMatchDeclaredSources(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{"one.py", "two.py"} {
		if err := os.WriteFile(filepath.Join(root, path), []byte("pass\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	inputs := map[string]sourceInput{"one.py": {}, "two.py": {}}
	one := filepath.Join(root, "one.py")
	two := filepath.Join(root, "two.py")
	if !ruffInputsMatch(root, inputs, []byte(one+"\n"+two+"\n")) {
		t.Fatal("complete selected input listing was rejected")
	}
	for _, listing := range []string{one + "\n", one + "\n" + one + "\n", one + "\n" + filepath.Join(root, "outside.py") + "\n", "one.py\n" + two + "\n"} {
		if ruffInputsMatch(root, inputs, []byte(listing)) {
			t.Fatalf("incomplete or forged input listing was accepted")
		}
	}
}

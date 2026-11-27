package analyzers

import (
	"os"
	"path/filepath"
	"testing"

	"scout/internal/domain"
)

func batchFixture(t *testing.T) (string, []domain.File) {
	t.Helper()
	root := t.TempDir()
	sources := map[string]string{
		"a_unsafe.py": "import yaml\n\ndef load(raw):\n    return yaml.unsafe_load(raw)\n",
		"b_clean.py":  "def add(a, b):\n    return a + b\n",
		"c_unsafe.py": "import yaml\n\ndef other(raw):\n    return yaml.unsafe_load(raw)\n",
	}
	var files []domain.File
	for _, name := range []string{"a_unsafe.py", "b_clean.py", "c_unsafe.py"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(sources[name]), 0o600); err != nil {
			t.Fatal(err)
		}
		files = append(files, domain.File{Path: name, Language: "python", Eligible: true})
	}
	return root, files
}

func TestBatchesRespectDeclaredLimits(t *testing.T) {
	root, files := batchFixture(t)
	spec := securitySpecs()[0]
	spec.Inputs.MaxFiles = 2
	batches := spec.Batches(root, files)
	if len(batches) != 2 || len(batches[0]) != 2 || len(batches[1]) != 1 || batches[0][0].Path != "a_unsafe.py" || batches[1][0].Path != "c_unsafe.py" {
		t.Fatalf("batches = %v", batches)
	}
}

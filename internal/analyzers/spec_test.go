package analyzers

import (
	"reflect"
	"testing"
)

func validSpecForTest(t *testing.T) Spec {
	t.Helper()
	spec := specForTest(t, "security.semgrep.python-unsafe-yaml-load")
	if err := spec.Validate(); err != nil {
		t.Fatal(err)
	}
	return spec
}

func TestSpecValidateRejectsUnsafeContracts(t *testing.T) {
	tests := map[string]func(*Spec){
		"project script binary":         func(s *Spec) { s.Binary.Path = "./scan.sh" },
		"nested binary path":            func(s *Spec) { s.Binary.Path = "bin/semgrep" },
		"network permission":            func(s *Spec) { s.Permissions.Network = true },
		"writable source":               func(s *Spec) { s.Permissions.ReadOnlySource = false },
		"unbounded output":              func(s *Spec) { s.MaxOutput = 0 },
		"unknown attestation":           func(s *Spec) { s.Attestation = "trust-me" },
		"attestation on non-SARIF":      func(s *Spec) { s.Attestation = AttestRuffFileListing },
		"unknown invocation format":     func(s *Spec) { s.Output.Format = "xml" },
		"exit semantics without zero":   func(s *Spec) { s.ExitCodes = map[int]ExitSemantics{1: ExitFindings} },
		"input declaration is required": func(s *Spec) { s.Binary.Args = []string{"scan"} },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			spec := validSpecForTest(t)
			mutate(&spec)
			if err := spec.Validate(); err == nil {
				t.Fatal("unsafe adapter contract was accepted")
			}
		})
	}
}

func TestStagedInputsRequireTheGitleaksAttestation(t *testing.T) {
	spec := specForTest(t, "security.gitleaks.default-credentials")
	if err := spec.Validate(); err != nil {
		t.Fatal(err)
	}
	spec.Attestation = AttestOutput
	if err := spec.Validate(); err == nil {
		t.Fatal("staged input directory without its attestation was accepted")
	}
}

func TestExpandArgsSubstitutesDeclaredPlaceholdersOnly(t *testing.T) {
	inputs := []ManifestInput{{Path: "a.py"}, {Path: "dir/b.py"}}
	got, err := expandArgs([]string{"scan", "{input_paths}", "--rules", "{bundled_rule_files}", "--m", "{input_manifest}"}, "/repo", "/state/m.json", inputs, []string{"/rules/r.yaml"}, "")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"scan", "a.py", "dir/b.py", "--rules", "/rules/r.yaml", "--m", "/state/m.json"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %v, want %v", got, want)
	}
	absolute, err := expandArgs([]string{"{absolute_input_paths}"}, "/repo", "", inputs, nil, "")
	if err != nil || !reflect.DeepEqual(absolute, []string{"/repo/a.py", "/repo/dir/b.py"}) {
		t.Fatalf("absolute args = %v, %v", absolute, err)
	}
	if _, err := expandArgs([]string{"{shell_command}"}, "/repo", "", inputs, nil, ""); err == nil {
		t.Fatal("unknown placeholder was expanded")
	}
}

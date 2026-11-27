package testsource

import (
	"go/parser"
	"go/token"
	"reflect"
	"testing"
)

func TestBehaviourWordsSplitAtCaseChangesAndUnderscores(t *testing.T) {
	for name, want := range map[string][]string{
		"TestRules":                         {"rules"},
		"TestTodoPlaceholderCase":           nil,
		"TestParseHTTPHeader_RejectsEmpty":  {"parse", "http", "header", "rejects", "empty"},
		"TestADR012CASParentStampAndFronti": {"adr012", "cas", "parent", "stamp", "and", "fronti"},
	} {
		if got := BehaviourWords(name); !reflect.DeepEqual(got, want) {
			t.Errorf("BehaviourWords(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestFingerprintIgnoresLocalNamesAndFailureMessages(t *testing.T) {
	src := `package p

import "testing"

func TestA(t *testing.T) { v := []int{1}; if got := Sum(v); got != 1 { t.Fatalf("a %d", got) } }
func TestB(x *testing.T) { in := []int{1}; if s := Sum(in); s != 1 { x.Fatalf("b") } }
func TestC(t *testing.T) { v := []int{2}; if got := Sum(v); got != 2 { t.Fatalf("a %d", got) } }
`
	fset := token.NewFileSet()
	af, err := parser.ParseFile(fset, "p_test.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	tests := topTests(af)
	a, b, c := Fingerprint(fset, "p", af, tests[0]), Fingerprint(fset, "p", af, tests[1]), Fingerprint(fset, "p", af, tests[2])
	if a == "" || a != b {
		t.Fatalf("renamed copy should share a fingerprint: %q %q", a, b)
	}
	if a == c {
		t.Fatal("different inputs must not share a fingerprint")
	}
}

func TestValidTierAcceptsOnlyTheDeclaredTierExpressions(t *testing.T) {
	tiers := Tiers{Boundary: "boundary", Deferred: "deferred"}
	for expr, want := range map[string]bool{"": true, "boundary": true, "deferred": true, "boundary && deferred": true, "integration": false, "!boundary": false, "linux": false} {
		if got := ValidTier(expr, tiers); got != want {
			t.Errorf("ValidTier(%q) = %v, want %v", expr, got, want)
		}
	}
}

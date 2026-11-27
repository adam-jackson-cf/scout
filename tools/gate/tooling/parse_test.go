package tooling

import (
	"strings"
	"testing"
)

func TestGolangciReportBecomesFindingsNamingLinter(t *testing.T) {
	got, err := parseGolangci([]byte(`{"Issues":[{"FromLinter":"forbidigo","Text":"use of exec.Command forbidden","Pos":{"Filename":"a/b.go","Line":6}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].String() != "a/b.go:6: GOLANGCI.forbidigo: use of exec.Command forbidden; use exec.CommandContext with a context deadline" {
		t.Fatalf("got %+v", got)
	}
}

func TestGolangciMalformedReportIsAnError(t *testing.T) {
	if _, err := parseGolangci([]byte("not json")); err == nil {
		t.Fatal("want an error")
	}
}

func TestShellcheckGccLinesBecomeFindings(t *testing.T) {
	got := parseShellcheck("scripts/x.sh:3:7: warning: Double quote to prevent globbing. [SC2086]\nnoise\n")
	if len(got) != 1 || got[0].Rule != "SHELLCHECK.SC2086" || got[0].Line != 3 || !strings.Contains(got[0].Fix, "SC2086") {
		t.Fatalf("got %+v", got)
	}
}

func TestRuffJSONAndFormatOutputBecomeFindings(t *testing.T) {
	lint, err := parseRuff("/r", `[{"code":"F401","message":"unused","filename":"/r/x.py","location":{"row":2},"url":"https://docs.astral.sh/ruff/rules/unused-import"}]`)
	if err != nil {
		t.Fatal(err)
	}
	if len(lint) != 1 || lint[0].File != "x.py" || lint[0].Rule != "RUFF.F401" || lint[0].Line != 2 {
		t.Fatalf("got %+v", lint)
	}
	format := parseRuffFormat("x.py:3:1: unformatted: File would be reformatted\n1 file would be reformatted\n")
	if len(format) != 1 || format[0].File != "x.py" || format[0].Rule != "RUFF.FORMAT" {
		t.Fatalf("got %+v", format)
	}
}

func TestMypyErrorLinesKeepTheirErrorCode(t *testing.T) {
	got := parseMypy("x.py:4: error: Incompatible types in assignment  [assignment]\nx.py:5: note: hint\n")
	if len(got) != 1 || got[0].Rule != "MYPY.assignment" || got[0].Line != 4 {
		t.Fatalf("got %+v", got)
	}
}

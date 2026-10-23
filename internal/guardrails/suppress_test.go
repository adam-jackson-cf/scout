package guardrails

import (
	"reflect"
	"testing"
)

type want struct {
	line      int
	check     string
	mechanism string
}

func run(path, base, head string) []Finding {
	b := map[string][]byte{}
	if base != "" {
		b[path] = []byte(base)
	}
	return SuppressAdded(b, map[string][]byte{path: []byte(head)}, []string{path})
}

func summarize(fs []Finding) []want {
	var out []want
	for _, f := range fs {
		out = append(out, want{f.Line, f.CheckID, f.Mechanism})
	}
	return out
}

func TestSuppressAddedForms(t *testing.T) {
	U, S := UnscopedCheckID, ScopedCheckID
	cases := []struct {
		name, path, head string
		want             []want
	}{
		{"py bare noqa", "a.py", "x = 1  # noqa\n", []want{{1, U, "noqa"}}},
		{"py bare noqa trailing colon", "a.py", "x = 1  # noqa:\n", []want{{1, U, "noqa"}}},
		{"py scoped noqa", "a.py", "x = 1  # noqa: E501\n", []want{{1, S, "noqa"}}},
		{"py bare type ignore", "a.py", "x = f()  # type: ignore\n", []want{{1, U, "type-ignore"}}},
		{"py scoped type ignore", "a.py", "x = f()  # type: ignore[attr-defined]\n", []want{{1, S, "type-ignore"}}},
		{"py pylint all", "a.py", "x = 1  # pylint: disable=all\n", []want{{1, U, "pylint-disable"}}},
		{"py pylint scoped", "a.py", "x = 1  # pylint: disable=unused-import\n", []want{{1, S, "pylint-disable"}}},
		{"py mypy ignore-errors", "a.py", "# mypy: ignore-errors\n", []want{{1, U, "mypy-ignore-errors"}}},
		{"py no_type_check", "a.py", "@no_type_check\ndef f(): pass\n", []want{{1, U, "no_type_check"}}},
		{"py none", "a.py", "x = 1  # fine\n", nil},
		{"go bare nolint", "a.go", "package a\n\nfunc f() {} //nolint\n", []want{{3, U, "nolint"}}},
		{"go nolint all", "a.go", "package a\n\nfunc f() {} //nolint:all\n", []want{{3, U, "nolint"}}},
		{"go scoped nolint", "a.go", "package a\n\nfunc f() {} //nolint:errcheck // reason\n", []want{{3, S, "nolint"}}},
		{"go none", "a.go", "package a\n\n// nolintish words\nfunc f() {}\n", nil},
		{"ts eslint-disable bare", "a.ts", "/* eslint-disable */\nlet a = 1\n", []want{{1, U, "eslint-disable"}}},
		{"ts eslint-disable-next-line bare", "a.ts", "// eslint-disable-next-line\nlet a = 1\n", []want{{1, U, "eslint-disable-next-line"}}},
		{"ts eslint-disable-line bare", "a.js", "let a = 1 // eslint-disable-line\n", []want{{1, U, "eslint-disable-line"}}},
		{"ts eslint scoped", "a.ts", "// eslint-disable-next-line no-console\nlet a = 1\n", []want{{1, S, "eslint-disable-next-line"}}},
		{"ts eslint bare with description", "a.ts", "// eslint-disable-next-line -- because\n", []want{{1, U, "eslint-disable-next-line"}}},
		{"ts ts-ignore", "a.tsx", "// @ts-ignore\nlet a: number = 'x'\n", []want{{1, U, "ts-ignore"}}},
		{"ts ts-nocheck", "a.ts", "// @ts-nocheck\n", []want{{1, U, "ts-nocheck"}}},
		{"ts block ts-ignore", "a.ts", "/* @ts-ignore */ let a = 1\n", []want{{1, U, "ts-ignore"}}},
		{"ts expect-error not flagged", "a.ts", "// @ts-expect-error\n", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := summarize(run(c.path, "", c.head))
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v want %+v", got, c.want)
			}
		})
	}
}

func TestSuppressAddedIDsVersionsExact(t *testing.T) {
	fs := run("a.py", "", "x = 1  # noqa\ny = 2  # noqa: E501\n")
	if len(fs) != 2 {
		t.Fatalf("got %d findings", len(fs))
	}
	if fs[0].CheckID != "guardrails.SUPPRESS.added.unscoped" || fs[0].Version != "guardrails.SUPPRESS.added.unscoped@1" || !fs[0].BlockingEligible() {
		t.Fatalf("unscoped: %+v", fs[0])
	}
	if fs[1].CheckID != "guardrails.SUPPRESS.added.scoped" || fs[1].Version != "guardrails.SUPPRESS.added.scoped@1" || fs[1].BlockingEligible() {
		t.Fatalf("scoped: %+v", fs[1])
	}
	if fs[0].Path != "a.py" || fs[0].Line != 1 {
		t.Fatalf("location: %+v", fs[0])
	}
}

func TestSuppressAddedOnlyAddedLines(t *testing.T) {
	base := "a = 1  # noqa\nb = 2\nc = 3  # noqa\n"
	t.Run("unchanged does not fire", func(t *testing.T) {
		if fs := run("a.py", base, base); len(fs) != 0 {
			t.Fatalf("unexpected %+v", fs)
		}
	})
	t.Run("lines inserted around existing do not fire", func(t *testing.T) {
		head := "import os\n" + base + "d = 4\n"
		if fs := run("a.py", base, head); len(fs) != 0 {
			t.Fatalf("unexpected %+v", fs)
		}
	})
	t.Run("moved suppression does not fire", func(t *testing.T) {
		head := "b = 2\nc = 3  # noqa\nx = 9\na = 1  # noqa\n"
		if fs := run("a.py", base, head); len(fs) != 0 {
			t.Fatalf("unexpected %+v", fs)
		}
	})
	t.Run("reindented suppression does not fire", func(t *testing.T) {
		head := "if True:\n    a = 1  # noqa\nb = 2\nc = 3  # noqa\n"
		if fs := run("a.py", base, head); len(fs) != 0 {
			t.Fatalf("unexpected %+v", fs)
		}
	})
	t.Run("edited line keeping its suppression does not fire", func(t *testing.T) {
		head := "a = 10  # noqa\nb = 2\nc = 3  # noqa\n"
		if fs := run("a.py", base, head); len(fs) != 0 {
			t.Fatalf("unexpected %+v", fs)
		}
	})
	t.Run("new suppression among old fires once with evidence", func(t *testing.T) {
		head := "a = 1  # noqa\nb = 2  # noqa\nc = 3  # noqa\n"
		fs := run("a.py", base, head)
		if len(fs) != 1 || fs[0].Line != 2 {
			t.Fatalf("got %+v", fs)
		}
		if fs[0].Evidence.Head != "b = 2  # noqa" || fs[0].Evidence.Base != "b = 2" {
			t.Fatalf("evidence %+v", fs[0].Evidence)
		}
	})
	t.Run("scoped upgraded to bare fires", func(t *testing.T) {
		b := "x = 1  # noqa: E501\n"
		h := "x = 1  # noqa\n"
		fs := run("a.py", b, h)
		if len(fs) != 1 || fs[0].CheckID != UnscopedCheckID {
			t.Fatalf("got %+v", fs)
		}
		if fs[0].Evidence.Base != "x = 1  # noqa: E501" || fs[0].Evidence.Head != "x = 1  # noqa" {
			t.Fatalf("evidence %+v", fs[0].Evidence)
		}
	})
	t.Run("pure addition has empty base evidence", func(t *testing.T) {
		fs := run("a.py", "a = 1\n", "a = 1\nz = 0  # noqa\n")
		if len(fs) != 1 || fs[0].Evidence.Base != "" || fs[0].Evidence.Head != "z = 0  # noqa" {
			t.Fatalf("got %+v", fs)
		}
	})
}

func TestSuppressAddedNotInStringLiterals(t *testing.T) {
	cases := []struct{ name, path, head string }{
		{"py single-line string", "a.py", "s = \"# noqa\"\nt = '# type: ignore'\n"},
		{"py triple-quoted multiline", "a.py", "doc = \"\"\"\nexplains # noqa\n# type: ignore\n\"\"\"\nx = 1\n"},
		{"py decorator text in string", "a.py", "s = \"@no_type_check\"\n"},
		{"go interpreted string", "a.go", "package a\n\nvar s = \"//nolint\"\n"},
		{"go raw string multiline", "a.go", "package a\n\nvar s = `\n//nolint\n`\n"},
		{"go rune then string", "a.go", "package a\n\nvar r, s = '\"', \"//nolint\"\n"},
		{"ts double string", "a.ts", "const s = \"// eslint-disable-next-line\"\n"},
		{"ts single string", "a.ts", "const s = '// @ts-ignore'\n"},
		{"ts template multiline", "a.ts", "const s = `\n// @ts-nocheck\n/* eslint-disable */\n`\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if fs := run(c.path, "", c.head); len(fs) != 0 {
				t.Fatalf("fired in string: %+v", fs)
			}
		})
	}
	t.Run("comment after a string still fires", func(t *testing.T) {
		fs := run("a.py", "", "s = \"x\"  # noqa\n")
		if len(fs) != 1 {
			t.Fatalf("got %+v", fs)
		}
	})
	t.Run("code after closed template fires", func(t *testing.T) {
		fs := run("a.ts", "", "const s = `a`; // @ts-ignore\n")
		if len(fs) != 1 {
			t.Fatalf("got %+v", fs)
		}
	})
}

func TestSuppressAddedScopeOfFiles(t *testing.T) {
	src := "x = 1  # noqa\n"
	files := map[string][]byte{
		"pkg/tests/a.py":           []byte(src),
		"pkg/test_a.py":            []byte(src),
		"pkg/a_test.py":            []byte(src),
		"pkg/gen.py":               []byte("# @generated\n" + src),
		"pkg/notes.txt":            []byte(src),
		"pkg/a.py":                 []byte(src),
		"x/b_test.go":              []byte("package b\nfunc f() {} //nolint\n"),
		"web/a.test.ts":            []byte("// @ts-ignore\n"),
		"gen/z.go":                 []byte("// Code generated by x. DO NOT EDIT.\npackage z\nfunc f() {} //nolint\n"),
		"web/untouched.ts":         []byte("// @ts-ignore\n"),
		"web/__tests__/b.ts":       []byte("// @ts-ignore\n"),
		"pkg/changed_elsewhere.py": []byte(src),
	}
	changed := []string{"pkg/tests/a.py", "pkg/test_a.py", "pkg/a_test.py", "pkg/gen.py", "pkg/notes.txt",
		"pkg/a.py", "x/b_test.go", "web/a.test.ts", "gen/z.go", "web/__tests__/b.ts", "deleted.py"}
	fs := SuppressAdded(nil, files, changed)
	if len(fs) != 1 || fs[0].Path != "pkg/a.py" {
		t.Fatalf("got %+v", fs)
	}
}

func TestSuppressAddedNewFileAndDeterministicOrder(t *testing.T) {
	head := map[string][]byte{
		"b.py": []byte("x = 1  # noqa\n"),
		"a.go": []byte("package a\nfunc f() {} //nolint\n"),
	}
	fs := SuppressAdded(nil, head, nil)
	if len(fs) != 2 || fs[0].Path != "a.go" || fs[1].Path != "b.py" {
		t.Fatalf("got %+v", fs)
	}
}

package codeintegrity

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScriptMocksFindStaticImportsAndCalls(t *testing.T) {
	src := "import sinon from 'sinon';\nimport {\n  instance,\n} from \"ts-mockito\";\nconst pq = require('proxyquire');\n// import x from 'testdouble'\nconst s = \"jest.mock('a')\";\njest.mock('./api');\nvi.doMock('./b');\nimport './side';\nimport t from 'sinon-extra';\n"
	got := scriptMocks([]byte(src))
	want := map[string]int{"sinon": 1, "ts-mockito": 2, "proxyquire": 5, "jest.mock": 8, "vi.doMock": 9}
	if len(got) != len(want) {
		t.Fatalf("uses = %+v", got)
	}
	for _, u := range got {
		if line, ok := want[u.name]; !ok || line != u.line {
			t.Errorf("unexpected use %+v", u)
		}
	}
}

func TestPythonMocksMatchLibrariesAndSubmodules(t *testing.T) {
	rep := pyReport{Imports: []pySpan{
		{Line: 1, Module: "unittest"}, {Line: 1, Module: "unittest.mock"},
		{Line: 2, Module: "mock"}, {Line: 2, Module: "mock.patch"},
		{Line: 3, Module: "pytest_mock.plugin"}, {Line: 4, Module: "mockito"},
	}}
	got := pythonMocks(rep)
	if len(got) != 3 || got[0].name != "unittest.mock" || got[1].name != "mock" || got[2].name != "pytest_mock" {
		t.Fatalf("uses = %+v", got)
	}
}

func TestLayoutSeparatesTestSupportAndFixturePaths(t *testing.T) {
	l := layout{SupportDir: "tests/support", FixturesDir: "tests/fixtures", TestFilePatterns: []string{"qa/**/check_*.py"}}
	for _, p := range []string{"pkg/test_x.py", "pkg/x_test.py", "conftest.py", "web/a.test.ts", "web/__tests__/a.ts", "tests/support/fake.py", "tests/fixtures/data.py", "qa/deep/check_x.py", "pkg/a_test.go"} {
		if !l.testPath(p) {
			t.Errorf("%s should be a test path", p)
		}
	}
	for _, p := range []string{"pkg/x.py", "web/a.ts", "tests/supported.py", "qa/check_x.ts", "testdata/x.py"} {
		if l.testPath(p) {
			t.Errorf("%s should be product code", p)
		}
	}
}

func TestDocstringsRequiredOnlyWhenRuffSelectsPydocstyle(t *testing.T) {
	cases := map[string]bool{
		"[lint]\nselect = [\"E\", \"F\"]\n":    false,
		"[lint]\nselect = [\"E\", \"D\"]\n":    true,
		"[lint]\nextend-select = [\"D103\"]\n": true,
		"select = [\"ALL\"]\n":                 true,
		"[lint]\nselect = [\"DTZ\", \"D2\"]\n": false,
	}
	for body, want := range cases {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "ruff.toml"), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		got, err := docstringsRequired(dir)
		if err != nil || got != want {
			t.Errorf("%q: got %v %v, want %v", body, got, err, want)
		}
	}
	got, err := docstringsRequired(t.TempDir())
	if err != nil || got {
		t.Fatalf("no config: got %v %v", got, err)
	}
}

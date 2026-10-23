package artifacts

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// ParseLCOV reads an LCOV tracefile: SF:<path> opens a record, DA:<line>,<hits>
// adds an instrumented line, end_of_record closes it. Other records are ignored.
func ParseLCOV(raw []byte, roots ...string) (Coverage, error) {
	coverage := Coverage{}
	var current map[int]int
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 64*1024), 16<<20)
	for number := 1; scanner.Scan(); number++ {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "SF:"):
			path := normalizePath(roots, line[3:])
			if coverage[path] == nil {
				coverage[path] = map[int]int{}
			}
			current = coverage[path]
		case line == "end_of_record":
			current = nil
		case strings.HasPrefix(line, "DA:"):
			if current == nil {
				return nil, fmt.Errorf("lcov line %d: DA outside an SF record", number)
			}
			fields := strings.Split(line[3:], ",")
			if len(fields) < 2 {
				return nil, fmt.Errorf("lcov line %d: malformed DA record", number)
			}
			at, errLine := strconv.Atoi(fields[0])
			hits, errHits := strconv.ParseFloat(fields[1], 64)
			if errLine != nil || errHits != nil || at < 1 || hits < 0 {
				return nil, fmt.Errorf("lcov line %d: malformed DA record", number)
			}
			current[at] += int(hits)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("lcov: %w", err)
	}
	if len(coverage) == 0 {
		return nil, fmt.Errorf("lcov: no SF record")
	}
	return coverage, nil
}

// ParseCobertura reads Cobertura XML: class `filename` joined to the
// `<source>` directories, with its class-level `<line number hits>` entries
// (method-level lines repeat them and are not counted).
func ParseCobertura(raw []byte, roots ...string) (Coverage, error) {
	var document struct {
		Sources []string `xml:"sources>source"`
		Classes []struct {
			Filename string `xml:"filename,attr"`
			Lines    []struct {
				Number int `xml:"number,attr"`
				Hits   int `xml:"hits,attr"`
			} `xml:"lines>line"`
		} `xml:"packages>package>classes>class"`
	}
	if err := xml.Unmarshal(raw, &document); err != nil {
		return nil, fmt.Errorf("cobertura: %w", err)
	}
	coverage := Coverage{}
	for _, class := range document.Classes {
		if class.Filename == "" {
			continue
		}
		path := resolveCoberturaPath(roots, document.Sources, class.Filename)
		target := coverage[path]
		if target == nil {
			target = map[int]int{}
			coverage[path] = target
		}
		for _, line := range class.Lines {
			if line.Number < 1 || line.Hits < 0 {
				return nil, fmt.Errorf("cobertura: malformed line in %s", class.Filename)
			}
			// A source file spread over several classes lists each line once per class.
			if line.Hits > target[line.Number] || !hasLine(target, line.Number) {
				target[line.Number] = line.Hits
			}
		}
	}
	if len(coverage) == 0 {
		return nil, fmt.Errorf("cobertura: no class entries")
	}
	return coverage, nil
}

func hasLine(m map[int]int, line int) bool { _, ok := m[line]; return ok }

// resolveCoberturaPath prefers the source directory under which the file is
// inside the repository root; else the filename as given.
func resolveCoberturaPath(roots []string, sources []string, filename string) string {
	if filepath.IsAbs(filename) {
		return normalizePath(roots, filename)
	}
	for _, source := range sources {
		source = strings.TrimSpace(source)
		if filepath.IsAbs(source) {
			joined := filepath.Join(source, filename)
			for _, root := range roots {
				if rel, err := filepath.Rel(root, joined); err == nil && !strings.HasPrefix(rel, "..") {
					return filepath.ToSlash(rel)
				}
			}
		}
	}
	return normalizePath(roots, filename)
}

type junitSuite struct {
	Tests    int          `xml:"tests,attr"`
	Failures int          `xml:"failures,attr"`
	Errors   int          `xml:"errors,attr"`
	Skipped  int          `xml:"skipped,attr"`
	Time     float64      `xml:"time,attr"`
	Cases    []junitCase  `xml:"testcase"`
	Suites   []junitSuite `xml:"testsuite"`
}

type junitCase struct {
	Time    float64   `xml:"time,attr"`
	Failure *struct{} `xml:"failure"`
	Error   *struct{} `xml:"error"`
	Skipped *struct{} `xml:"skipped"`
}

// ParseJUnit sums a JUnit report. Totals come from the test cases when the
// suite lists them, else from the suite's attributes.
func ParseJUnit(raw []byte) (TestResults, error) {
	var document struct {
		XMLName xml.Name
		junitSuite
	}
	if err := xml.Unmarshal(raw, &document); err != nil {
		return TestResults{}, fmt.Errorf("junit: %w", err)
	}
	var total TestResults
	switch document.XMLName.Local {
	case "testsuites":
		for _, suite := range document.Suites {
			addSuite(&total, suite)
		}
	case "testsuite":
		addSuite(&total, document.junitSuite)
	default:
		return TestResults{}, fmt.Errorf("junit: unexpected root element %q", document.XMLName.Local)
	}
	if total.Suites == 0 {
		return TestResults{}, fmt.Errorf("junit: no test suite")
	}
	return total, nil
}

func addSuite(total *TestResults, suite junitSuite) {
	if len(suite.Suites) > 0 && len(suite.Cases) == 0 {
		for _, child := range suite.Suites {
			addSuite(total, child)
		}
		return
	}
	total.Suites++
	if len(suite.Cases) == 0 {
		total.Tests += suite.Tests
		total.Failures += suite.Failures
		total.Errors += suite.Errors
		total.Skipped += suite.Skipped
		total.Seconds += suite.Time
		return
	}
	for _, c := range suite.Cases {
		total.Tests++
		total.Seconds += c.Time
		switch {
		case c.Failure != nil:
			total.Failures++
		case c.Error != nil:
			total.Errors++
		case c.Skipped != nil:
			total.Skipped++
		}
	}
	for _, child := range suite.Suites {
		addSuite(total, child)
	}
}

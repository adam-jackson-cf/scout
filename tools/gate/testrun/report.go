package testrun

import (
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"
)

type Record struct {
	Package string
	Name    string
	Outcome string
	Seconds float64
	Reason  string
	Output  string
}

type Report struct {
	XMLName   xml.Name `xml:"testsuites"`
	Tier      string   `xml:"tier,attr"`
	Guard     string   `xml:"guard,attr"`
	Timestamp string   `xml:"timestamp,attr"`
	Wall      float64  `xml:"time,attr"`
	Exit      int      `xml:"exit,attr"`
	Suites    []Suite  `xml:"testsuite"`
}

type Suite struct {
	Name    string  `xml:"name,attr"`
	Outcome string  `xml:"outcome,attr"`
	Reason  string  `xml:"reason,attr,omitempty"`
	Cases   []Case  `xml:"testcase"`
	Time    float64 `xml:"time,attr"`
}

type Case struct {
	Name    string   `xml:"name,attr"`
	Class   string   `xml:"classname,attr"`
	Time    float64  `xml:"time,attr"`
	Failure *Message `xml:"failure,omitempty"`
	Skipped *Message `xml:"skipped,omitempty"`
}

type Message struct {
	Message string `xml:"message,attr"`
	Body    string `xml:",chardata"`
}

func (c Case) Outcome() string {
	switch {
	case c.Failure != nil:
		return "fail"
	case c.Skipped != nil:
		return "skip"
	}
	return "pass"
}

func (c Case) Reason() string {
	if c.Failure != nil {
		return c.Failure.Message
	}
	if c.Skipped != nil {
		return c.Skipped.Message
	}
	return ""
}

type pkgState struct {
	outcome string
	elapsed float64
	output  strings.Builder
	running string
	cases   map[string]*Record
	order   []string
}

func Normalise(tier, guard string, res result, events []event) Report {
	pkgs := map[string]*pkgState{}
	var names []string
	for _, ev := range events {
		if ev.Package == "" {
			continue
		}
		ps := pkgs[ev.Package]
		if ps == nil {
			ps = &pkgState{cases: map[string]*Record{}}
			pkgs[ev.Package] = ps
			names = append(names, ev.Package)
		}
		ps.apply(ev)
	}
	sort.Strings(names)
	rep := Report{Tier: tier, Guard: guard, Timestamp: res.started.UTC().Format(time.RFC3339Nano), Wall: res.elapsed.Seconds(), Exit: res.exit}
	for _, n := range names {
		rep.Suites = append(rep.Suites, pkgs[n].suite(n))
	}
	return rep
}

func (ps *pkgState) apply(ev event) {
	if ev.Test == "" {
		switch ev.Action {
		case "pass", "fail", "skip":
			ps.outcome, ps.elapsed = ev.Action, ev.Elapsed
		case "output":
			ps.output.WriteString(ev.Output)
		}
		return
	}
	rec := ps.cases[ev.Test]
	if rec == nil {
		rec = &Record{Package: ev.Package, Name: ev.Test}
		ps.cases[ev.Test] = rec
		ps.order = append(ps.order, ev.Test)
	}
	switch ev.Action {
	case "run":
		ps.running = ev.Test
	case "output":
		rec.Output += ev.Output
	case "pass", "fail", "skip":
		rec.Outcome, rec.Seconds = ev.Action, ev.Elapsed
		if ps.running == ev.Test {
			ps.running = ""
		}
	}
}

func (ps *pkgState) suite(name string) Suite {
	s := Suite{Name: name, Outcome: ps.outcome, Time: ps.elapsed}
	out := ps.output.String()
	if ps.outcome == "fail" {
		s.Reason = packageReason(out)
	}
	for _, t := range ps.order {
		rec := ps.cases[t]
		if rec.Outcome == "" {
			rec.Outcome = "fail"
			rec.Reason = "did not finish: " + s.Reason
		}
		s.Cases = append(s.Cases, caseOf(rec))
	}
	return s
}

func packageReason(out string) string {
	switch {
	case strings.Contains(out, "signal: killed"):
		return "killed by the T1 guard"
	case strings.Contains(out, "panic: test timed out"):
		return "package exceeded the hang timeout"
	case strings.Contains(out, "build failed") || strings.Contains(out, "[setup failed]"):
		return "build failed"
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	return lines[len(lines)-1]
}

func caseOf(rec *Record) Case {
	c := Case{Name: rec.Name, Class: rec.Package, Time: rec.Seconds}
	msg := lastLines(rec.Output, 6)
	switch rec.Outcome {
	case "fail":
		reason := rec.Reason
		if reason == "" {
			reason = "test failed"
		}
		c.Failure = &Message{Message: reason, Body: msg}
	case "skip":
		c.Skipped = &Message{Message: skipMessage(rec.Output), Body: msg}
	}
	return c
}

func skipMessage(out string) string {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "=== ") || strings.HasPrefix(line, "--- ") {
			continue
		}
		if _, msg, ok := strings.Cut(line, ": "); ok && strings.Contains(line, ".go:") {
			return msg
		}
		return line
	}
	return "skipped"
}

func lastLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

func WriteReport(path string, rep Report) error {
	data, err := xml.MarshalIndent(rep, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append([]byte(xml.Header), data...), 0o600)
}

func ReadReport(path string) (Report, error) {
	var rep Report
	data, err := os.ReadFile(path)
	if err != nil {
		return rep, fmt.Errorf("run report %s is missing: %w", path, err)
	}
	if err := xml.Unmarshal(data, &rep); err != nil {
		return rep, fmt.Errorf("run report %s is invalid: %w", path, err)
	}
	if _, err := time.Parse(time.RFC3339Nano, rep.Timestamp); err != nil {
		return rep, fmt.Errorf("run report %s has no valid timestamp", path)
	}
	if invalid(rep.Wall) {
		return rep, fmt.Errorf("run report %s has an invalid wall time %v", path, rep.Wall)
	}
	for _, s := range rep.Suites {
		for _, c := range s.Cases {
			if invalid(c.Time) {
				return rep, fmt.Errorf("run report %s has an invalid duration for %s", path, c.Name)
			}
		}
	}
	return rep, nil
}

func invalid(v float64) bool { return v < 0 || math.IsNaN(v) || math.IsInf(v, 0) }

package testrun

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"scout/tools/gate/check"
	"scout/tools/gate/testsource"
)

const (
	TierT1         = "t1"
	TierT1Deferred = "t1-deferred"
	TierT2         = "t2"
	TierT2Deferred = "t2-deferred"
)

type event struct {
	Time    time.Time
	Action  string
	Package string
	Test    string
	Output  string
	Elapsed float64
}

type listing struct {
	col      *testsource.Collection
	packages []string
	byTier   map[string]map[string]bool
	byKey    map[string]testsource.StaticTest
	failures []check.Finding
	env      []string
}

func pkgKey(pkg, name string) string { return pkg + "." + name }

func (l *listing) tagSets() map[string]string {
	used := map[string]bool{}
	for _, f := range l.col.Files {
		used[f.Constraint] = true
	}
	t := l.col.Tiers
	sets := map[string]string{"": "", t.Boundary: t.Boundary}
	if used[t.Deferred] || used[t.Boundary+" && "+t.Deferred] || used[t.Deferred+" && "+t.Boundary] {
		sets[t.Deferred] = t.Deferred
		sets[t.Boundary+","+t.Deferred] = t.Boundary + "," + t.Deferred
	}
	return sets
}

func buildListing(r *check.Run, col *testsource.Collection, env []string) (*listing, error) {
	l := &listing{col: col, byTier: map[string]map[string]bool{}, byKey: map[string]testsource.StaticTest{}, env: env}
	pkgs := map[string]bool{}
	for _, t := range col.Tests {
		pkgs[t.ImportPath] = true
		l.byKey[pkgKey(t.ImportPath, t.Name)] = t
	}
	for _, f := range col.Files {
		pkgs[importPathOf(col, f.Dir)] = true
	}
	for p := range pkgs {
		l.packages = append(l.packages, p)
	}
	sort.Strings(l.packages)
	if len(l.packages) == 0 {
		return l, nil
	}
	raw, err := l.listAll(r)
	if err != nil {
		return nil, err
	}
	l.assign(raw)
	return l, nil
}

func importPathOf(col *testsource.Collection, dir string) string {
	if dir == "." {
		return col.Module
	}
	return col.Module + "/" + dir
}

func (l *listing) listAll(r *check.Run) (map[string]map[string]bool, error) {
	sets := l.tagSets()
	out := map[string]map[string]bool{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	errs := make(chan error, len(sets))
	for key, tags := range sets {
		wg.Add(1)
		go func(key, tags string) {
			defer wg.Done()
			ids, fails, err := l.list(tags)
			mu.Lock()
			defer mu.Unlock()
			out[key] = ids
			l.failures = append(l.failures, fails...)
			if err != nil {
				errs <- err
			}
		}(key, tags)
	}
	wg.Wait()
	close(errs)
	if err, ok := <-errs; ok {
		return nil, err
	}
	return out, nil
}

func (l *listing) list(tags string) (map[string]bool, []check.Finding, error) {
	args := []string{"test", "-json", "-list", "^(Test|Benchmark|Example|Fuzz)"}
	if tags != "" {
		args = append(args, "-tags", tags)
	}
	args = append(args, l.packages...)
	res, err := goCommand(l.col.Root, l.env, 20*time.Minute, args...)
	if err != nil {
		return nil, nil, err
	}
	ids, status := listedIDs(parseEvents(res.out))
	var fails []check.Finding
	for _, p := range l.packages {
		if status[p] == "fail" || (res.exit != 0 && status[p] == "") {
			fails = append(fails, check.Finding{File: p, Rule: "TEST.failure-propagation", Problem: fmt.Sprintf("tests of %s cannot be collected with tags %q", p, tags), Fix: "fix the build error go test -list reports for this package"})
		}
	}
	return ids, fails, nil
}

func listedIDs(events []event) (map[string]bool, map[string]string) {
	ids := map[string]bool{}
	status := map[string]string{}
	for _, ev := range events {
		if ev.Test != "" {
			continue
		}
		switch ev.Action {
		case "fail", "pass", "skip":
			status[ev.Package] = ev.Action
		case "output":
			if name := strings.TrimSpace(ev.Output); isTopTest(name) {
				ids[pkgKey(ev.Package, name)] = true
			}
		}
	}
	return ids, status
}

func isTopTest(s string) bool {
	return strings.HasPrefix(s, "Test") && !strings.ContainsAny(s, " \t/")
}

func parseEvents(raw []byte) []event {
	var out []event
	sc := bufio.NewScanner(bytes.NewReader(raw))
	sc.Buffer(make([]byte, 0, 1024*1024), 64*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 || line[0] != '{' {
			continue
		}
		var ev event
		if json.Unmarshal(line, &ev) == nil {
			out = append(out, ev)
		}
	}
	return out
}

func (l *listing) assign(raw map[string]map[string]bool) {
	t := l.col.Tiers
	base, boundary := raw[""], raw[t.Boundary]
	deferred, both := raw[t.Deferred], raw[t.Boundary+","+t.Deferred]
	l.byTier[TierT1] = base
	l.byTier[TierT2] = minus(boundary, base)
	l.byTier[TierT1Deferred] = minus(deferred, base)
	l.byTier[TierT2Deferred] = minus(minus(both, boundary), deferred)
}

func minus(a, b map[string]bool) map[string]bool {
	out := map[string]bool{}
	for k := range a {
		if !b[k] {
			out[k] = true
		}
	}
	return out
}

func expectedTier(constraint string, t testsource.Tiers) string {
	switch constraint {
	case "":
		return TierT1
	case t.Boundary:
		return TierT2
	case t.Deferred:
		return TierT1Deferred
	case t.Boundary + " && " + t.Deferred, t.Deferred + " && " + t.Boundary:
		return TierT2Deferred
	}
	return ""
}

func (l *listing) tierOf(key string) []string {
	var out []string
	for _, tier := range []string{TierT1, TierT1Deferred, TierT2, TierT2Deferred} {
		if l.byTier[tier][key] {
			out = append(out, tier)
		}
	}
	return out
}

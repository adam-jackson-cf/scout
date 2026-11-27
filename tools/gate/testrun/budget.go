package testrun

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"scout/tools/gate/check"
	"scout/tools/gate/proof"
)

type machine struct {
	ID       string  `toml:"id"`
	Wall     float64 `toml:"wall_limit_s"`
	PerTest  float64 `toml:"per_test_limit_s"`
	Approval string  `toml:"approval"`
}

type budget struct {
	Machines []machine `toml:"machine"`
}

const fixtureMachine = "fixture"

func budgetSource(r *check.Run, s *session) (*proof.File, bool, error) {
	if !s.col.Fixture {
		return r.Proofs, false, nil
	}
	loaded, err := proof.Load(filepath.Join(s.col.Root, "proofs.toml"))
	return loaded, true, err
}

func limitsFor(r *check.Run, s *session, id string) (machine, bool, error) {
	pf, allowFixture, err := budgetSource(r, s)
	if err != nil {
		return machine{}, false, err
	}
	var b budget
	if _, err := pf.Section("test_budget", &b); err != nil {
		return machine{}, false, fmt.Errorf("proofs.toml [test_budget]: %w", err)
	}
	for _, m := range b.Machines {
		if m.ID == id || allowFixture && m.ID == fixtureMachine {
			return m, true, validMachine(m)
		}
	}
	return machine{}, false, nil
}

func validMachine(m machine) error {
	if m.Wall <= 0 || m.PerTest <= 0 || m.Approval == "" {
		return fmt.Errorf("proofs.toml [[test_budget.machine]] %s needs positive wall_limit_s and per_test_limit_s and an approval", m.ID)
	}
	return nil
}

func runBudget(r *check.Run) ([]check.Finding, error) {
	s := sessionFor(r)
	if err := s.prepare(r); err != nil {
		return nil, err
	}
	path, err := s.run(r, TierT1)
	if err != nil {
		return nil, err
	}
	rep, err := ReadReport(path)
	if err != nil {
		return []check.Finding{{File: path, Rule: "TEST.budget.t1-time", Problem: err.Error(), Fix: "rerun the commit; the budget needs this run's T1 report"}}, nil
	}
	if f := staleReport(path, rep, s.started); f != nil {
		return []check.Finding{*f}, nil
	}
	id := MachineID(s.goVer)
	limits, known, err := limitsFor(r, s, id)
	if err != nil {
		return nil, err
	}
	if !known {
		reportOnly(id, rep)
		return nil, nil
	}
	return overBudget(s, rep, limits), nil
}

func staleReport(path string, rep Report, started time.Time) *check.Finding {
	ts, _ := time.Parse(time.RFC3339Nano, rep.Timestamp)
	if ts.Before(started) {
		return &check.Finding{File: path, Rule: "TEST.budget.t1-time", Problem: fmt.Sprintf("T1 report from %s predates this run (%s)", rep.Timestamp, started.UTC().Format(time.RFC3339)), Fix: "rerun the commit so the budget reads a fresh T1 report"}
	}
	return nil
}

func reportOnly(id string, rep Report) {
	fmt.Fprintf(os.Stderr, "gate: test.budget: machine %s has no approved limits; T1 wall %.2fs. Record at least 5 baseline commits on this machine and ask the owner to approve [[test_budget.machine]] id = %q in proofs.toml\n", id, rep.Wall, id)
	slow := slowest(rep, 5)
	for _, c := range slow {
		fmt.Fprintf(os.Stderr, "gate: test.budget:   %.2fs %s.%s\n", c.Time, c.Class, c.Name)
	}
}

func slowest(rep Report, n int) []Case {
	var all []Case
	for _, s := range rep.Suites {
		for _, c := range s.Cases {
			if !isSubtest(c.Name) {
				all = append(all, c)
			}
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Time > all[j].Time })
	if len(all) > n {
		all = all[:n]
	}
	return all
}

func isSubtest(name string) bool {
	for _, r := range name {
		if r == '/' {
			return true
		}
	}
	return false
}

func overBudget(s *session, rep Report, m machine) []check.Finding {
	var out []check.Finding
	if rep.Wall > m.Wall {
		out = append(out, check.Finding{File: "T1", Rule: "TEST.budget.t1-time", Problem: fmt.Sprintf("T1 took %.2fs, over the %.2fs limit approved for %s", rep.Wall, m.Wall, m.ID), Fix: "speed up the slowest T1 tests; never move a test to another tier for speed"})
	}
	for _, suite := range rep.Suites {
		for _, c := range suite.Cases {
			if isSubtest(c.Name) || c.Time <= m.PerTest {
				continue
			}
			file := s.list.byKey[pkgKey(c.Class, c.Name)].Display
			out = append(out, check.Finding{File: file, Rule: "TEST.budget.per-test-time", Problem: fmt.Sprintf("%s took %.2fs, over the %.2fs per-test limit approved for %s", c.Name, c.Time, m.PerTest, m.ID), Fix: "speed the test up; never move it to another tier for speed, and defer it only with an owner-approved exception"})
		}
	}
	return out
}
